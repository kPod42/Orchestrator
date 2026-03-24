package executor

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"Orch/internal/agent/model"
)

type Executor struct{}

func New() *Executor {
	return &Executor{}
}

func (e *Executor) RunAction(ctx context.Context, action string, args map[string]string) (<-chan model.Event, error) {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "echo":
		ch := make(chan model.Event, 2)
		go func() {
			defer close(ch)

			text := args["text"]
			ch <- model.Event{
				Output: &model.Output{
					Stream: "stdout",
					Chunk:  text,
				},
			}
			ch <- model.Event{
				Result: &model.Result{
					Success:  true,
					ExitCode: 0,
					Message:  "action completed",
				},
			}
		}()
		return ch, nil
	case "hostname":
		if runtime.GOOS == "windows" {
			return e.runCommand(ctx, "cmd", "hostname", 0)
		}
		return e.runCommand(ctx, "sh", "hostname", 0)
	case "list_processes":
		if runtime.GOOS == "windows" {
			return e.runCommand(ctx, "cmd", "tasklist", 0)
		}
		return e.runCommand(ctx, "sh", "ps -eo pid,comm,%cpu,%mem", 0)
	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (e *Executor) RunExec(ctx context.Context, shell, command string, timeout time.Duration) (<-chan model.Event, error) {
	return e.runCommand(ctx, shell, command, timeout)
}

func (e *Executor) runCommand(ctx context.Context, shell, command string, timeout time.Duration) (<-chan model.Event, error) {
	if strings.TrimSpace(command) == "" {
		return nil, fmt.Errorf("empty command")
	}
	cmd, execCtx, cancel, err := buildCommand(ctx, shell, command, timeout)
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	ch := make(chan model.Event, 64)

	go func() {
		defer close(ch)
		defer cancel()
		if err := cmd.Start(); err != nil {
			ch <- model.Event{
				Result: &model.Result{
					Success:  false,
					ExitCode: 1,
					Message:  fmt.Sprintf("start failed: %v", err),
				},
			}
			return
		}
		var wg sync.WaitGroup
		wg.Add(2)
		go readPipe(&wg, stdout, "stdout", ch)
		go readPipe(&wg, stderr, "stderr", ch)

		waitErr := cmd.Wait()
		wg.Wait()

		if execCtx.Err() != nil {
			if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
				ch <- model.Event{
					Result: &model.Result{
						Success:  false,
						ExitCode: 1,
						Message:  fmt.Sprintf("command timed out"),
					},
				}
				return
			}
			ch <- model.Event{
				Result: &model.Result{
					Success:  false,
					ExitCode: 125,
					Message:  fmt.Sprintf("command canceled"),
				},
			}
			return
		}

		if waitErr != nil {
			exitCode := int32(1)
			var exitErr *exec.ExitError
			if errors.As(waitErr, &exitErr) {
				exitCode = int32(exitErr.ExitCode())
			}
			ch <- model.Event{
				Result: &model.Result{
					Success:  false,
					ExitCode: exitCode,
					Message:  waitErr.Error(),
				},
			}
			return
		}
		ch <- model.Event{
			Result: &model.Result{
				Success:  true,
				ExitCode: 0,
				Message:  "command completed",
			},
		}
	}()
	return ch, nil
}

func buildCommand(parent context.Context, shell, command string, timeout time.Duration) (*exec.Cmd, context.Context, context.CancelFunc, error) {
	var execCtx context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		execCtx, cancel = context.WithTimeout(parent, timeout)
	} else {
		execCtx, cancel = context.WithCancel(parent)
	}

	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "powershell":
		return exec.CommandContext(execCtx, "powershell", "-NoProfile", "-Command", command), execCtx, cancel, nil
	case "cmd":
		if runtime.GOOS != "windows" {
			cancel()
			return nil, nil, nil, fmt.Errorf("cmd shell is only available on windows")
		}
		return exec.CommandContext(execCtx, "cmd", "/C", command), execCtx, cancel, nil
	case "bash":
		return exec.CommandContext(execCtx, "bash", "-lc", command), execCtx, cancel, nil
	case "sh":
		return exec.CommandContext(execCtx, "sh", "-c", command), execCtx, cancel, nil
	default:
		cancel()
		return nil, nil, nil, fmt.Errorf("unsupported shell: %s", shell)
	}
}

func readPipe(wg *sync.WaitGroup, r io.Reader, stream string, out chan<- model.Event) {
	defer wg.Done()

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		out <- model.Event{
			Output: &model.Output{
				Stream: stream,
				Chunk:  scanner.Text(),
			},
		}
	}

	if err := scanner.Err(); err != nil {
		out <- model.Event{
			Output: &model.Output{
				Stream: "stderr",
				Chunk:  fmt.Sprintf("scanner error: %v", err),
			},
		}
	}
}
