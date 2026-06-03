package console

import (
	"bufio"
	"context"
	"io"
	"time"

	"Orch/internal/host/core"
	"Orch/internal/host/parser"
)

type UI struct {
	input          io.Reader
	printer        *Printer
	parser         *parser.Parser
	core           *core.Service
	requestTimeout time.Duration
}

func New(
	input io.Reader,
	output io.Writer,
	errorOutput io.Writer,
	parser *parser.Parser,
	core *core.Service,
	requestTimeout time.Duration,
) *UI {
	return &UI{
		input:          input,
		printer:        NewPrinter(output, errorOutput),
		parser:         parser,
		core:           core,
		requestTimeout: requestTimeout,
	}
}

func (u *UI) Run(ctx context.Context) error {
	u.printer.Banner()

	scanner := bufio.NewScanner(u.input)

	for {
		if ctx.Err() != nil {
			u.printer.Message("Host остановлен.")
			return nil
		}

		u.printer.Prompt()

		if !scanner.Scan() {
			return scanner.Err()
		}

		if ctx.Err() != nil {
			u.printer.Message("Host остановлен.")
			return nil
		}

		line := scanner.Text()

		parsed, err := u.parser.ParseCommand(line)
		if err != nil {
			u.printer.Error(err)
			continue
		}

		switch parsed.Kind {
		case parser.KindEmpty:
			continue

		case parser.KindExit:
			u.printer.Message("Host остановлен.")
			return nil

		case parser.KindHelp:
			u.printer.Help()

		case parser.KindNodes:
			opCtx, cancel := context.WithTimeout(ctx, u.requestTimeout)

			nodes, err := u.core.GetNodes(opCtx)

			cancel()

			if err != nil {
				u.printer.Error(err)
				continue
			}

			u.printer.Nodes(nodes)

		case parser.KindAction:
			opCtx, cancel := context.WithTimeout(ctx, u.requestTimeout)

			response, err := u.core.ExecuteAction(
				opCtx,
				parsed.ActionName,
				parsed.Targets,
				parsed.Args,
			)

			cancel()

			if err != nil {
				u.printer.Error(err)
				continue
			}

			u.printer.ExecutionResults(response.Results)

		default:
			u.printer.Message("Неизвестная команда. Введите help.")
		}
	}
}
