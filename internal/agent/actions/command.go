package actions

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

func runCommand(ctx context.Context, shell, command string, timeout time.Duration) (<-chan model.Event, error) {
	if strings.TrimSpace(command) == "" {
		return nil, fmt.Errorf("empty command")
	}

	cmd, execCtx, cancel, err := buildCommand(ctx, shell, command, timeout)
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
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
			ch <- resultEvent(false, 1, fmt.Sprintf("start failed: %v", err))
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
				ch <- resultEvent(false, 1, "command timed out")
				return
			}

			ch <- resultEvent(false, 125, "command canceled")
			return
		}

		if waitErr != nil {
			exitCode := int32(1)

			var exitErr *exec.ExitError
			if errors.As(waitErr, &exitErr) {
				exitCode = int32(exitErr.ExitCode())
			}

			ch <- resultEvent(false, exitCode, waitErr.Error())
			return
		}

		ch <- resultEvent(true, 0, "command completed")
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

	name, args, err := shellCommand(shell, command)
	if err != nil {
		cancel()
		return nil, nil, nil, err
	}

	return exec.CommandContext(execCtx, name, args...), execCtx, cancel, nil
}

func shellCommand(shell, command string) (string, []string, error) {
	switch strings.ToLower(strings.TrimSpace(shell)) {
	case "powershell":
		return "powershell", []string{"-NoProfile", "-Command", command}, nil

	case "cmd":
		if runtime.GOOS != "windows" {
			return "", nil, fmt.Errorf("cmd shell is only available on windows")
		}
		return "cmd", []string{"/C", command}, nil

	case "bash":
		return "bash", []string{"-lc", command}, nil

	case "sh":
		return "sh", []string{"-c", command}, nil

	default:
		return "", nil, fmt.Errorf("unsupported shell: %s", shell)
	}
}

func readPipe(wg *sync.WaitGroup, r io.Reader, stream string, out chan<- model.Event) {
	defer wg.Done()

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		out <- outputEvent(stream, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		out <- outputEvent("stderr", fmt.Sprintf("scanner error: %v", err))
	}
}
