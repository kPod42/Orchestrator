package console

import (
	"fmt"
	"io"
	"strings"

	"Orch/internal/host/model"
)

type Printer struct {
	out io.Writer
	err io.Writer
}

func NewPrinter(out io.Writer, err io.Writer) *Printer {
	return &Printer{
		out: out,
		err: err,
	}
}

func (p *Printer) Banner() {
	fmt.Fprintln(p.out, "Orchestrator Host Console")
	fmt.Fprintln(p.out, "Введите help для списка команд.")
	fmt.Fprintln(p.out)
}

func (p *Printer) Prompt() {
	fmt.Fprint(p.out, "host> ")
}

func (p *Printer) Help() {
	fmt.Fprintln(p.out, "Доступные команды:")
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, "  nodes")
	fmt.Fprintln(p.out, "      показать активные узлы")
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, "  action <name> on <target> [key=value...]")
	fmt.Fprintln(p.out, "      выполнить Action через Coordinator")
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, "      примеры:")
	fmt.Fprintln(p.out, "        action hostname on agent-node1")
	fmt.Fprintln(p.out, "        action hostname on agent-node1,agent-node2")
	fmt.Fprintln(p.out, "        action hostname on all")
	fmt.Fprintln(p.out, "        action echo on agent-node1 text=\"hello from host\"")
	fmt.Fprintln(p.out, "        action list_processes on all")
	fmt.Fprintln(p.out)
	fmt.Fprintln(p.out, "  exit")
	fmt.Fprintln(p.out, "      завершить работу Host")
	fmt.Fprintln(p.out)
}

func (p *Printer) Error(err error) {
	fmt.Fprintf(p.err, "Ошибка: %v\n", err)
}

func (p *Printer) Message(message string) {
	fmt.Fprintln(p.out, message)
}

func (p *Printer) Nodes(nodes []model.Node) {
	if len(nodes) == 0 {
		fmt.Fprintln(p.out, "Активные узлы не найдены.")
		return
	}

	for _, node := range nodes {
		fmt.Fprintf(p.out, "Узел: %s\n", node.ID)
		fmt.Fprintf(p.out, "  Занят: %v\n", node.Busy)

		if len(node.Capabilities) == 0 {
			fmt.Fprintln(p.out, "  Возможности: отсутствуют")
		} else {
			fmt.Fprintf(p.out, "  Возможности: %s\n", strings.Join(node.Capabilities, ", "))
		}

		if len(node.Endpoints) == 0 {
			fmt.Fprintln(p.out, "  Endpoints: отсутствуют")
		} else {
			fmt.Fprintln(p.out, "  Endpoints:")
			for _, endpoint := range node.Endpoints {
				fmt.Fprintf(
					p.out,
					"    - kind=%s address=%s scope=%s priority=%d\n",
					endpoint.Kind,
					endpoint.Address,
					endpoint.Scope,
					endpoint.Priority,
				)
			}
		}

		fmt.Fprintln(p.out)
	}
}

func (p *Printer) ExecutionResults(results []model.NodeExecutionResult) {
	if len(results) == 0 {
		fmt.Fprintln(p.out, "Результаты отсутствуют.")
		return
	}

	for _, result := range results {
		status := translateStatus(result.Status)

		fmt.Fprintf(p.out, "Узел %s — статус: %s\n", result.NodeID, status)

		if result.Status != "completed" {
			fmt.Fprintf(p.out, "  Код: %d\n", result.ExitCode)
			fmt.Fprintf(p.out, "  Сообщение: %s\n", result.Message)
		}
	}
}

func translateStatus(status string) string {
	switch status {
	case "completed":
		return "выполнено"
	case "error":
		return "ошибка"
	default:
		return status
	}
}
