package controller

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/snivilised/pants"
)

type shellResult struct {
	output []byte
	err    error
}

// jayShellSession is a non-interactive persistent shell session that uses
// zsh +m (no -i flag) to avoid terminal corruption while keeping custom
// shell functions available via a one-time startup source of ~/.zshrc.
type jayShellSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	reader *bufio.Reader
	marker string
}

// profilePathFor returns the path to the user's shell profile/rc file
// for a given shell name, or empty string if unknown.
func profilePathFor(shell, home string) string {
	switch shell {
	case "zsh":
		return home + "/.zshrc"
	case "bash":
		return home + "/.bashrc"
	case "sh", "dash", "ksh":
		return home + "/.profile"
	default:
		return ""
	}
}

// newJayShellSession creates a new non-interactive persistent shell session.
// It sources the user's profile once at startup so custom functions are
// available, without needing the -i flag (which causes terminal corruption).
func newJayShellSession(shellPath string) (*jayShellSession, error) {
	cmd := exec.Command(shellPath, "+m")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil { //nolint:govet // shadow ok
		return nil, err
	}

	marker := "__JAYWALK_CMD_DONE__"
	reader := bufio.NewReader(stdout)

	// Source profile once at startup so shell functions are available.
	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		rcPath := profilePathFor(shellPath, home)
		if rcPath != "" {
			startupMarker := "__JAYWALK_STARTUP__"

			_, _ = io.WriteString(stdin,
				fmt.Sprintf(". %s >/dev/null 2>&1; echo %s\n", rcPath, startupMarker),
			)

			// Wait for startup completion with a generous timeout to guard
			// against the rare case where the profile does something blocking.
			done := make(chan struct{})
			go func() {
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						return
					}
					if strings.Contains(line, startupMarker) {
						close(done)
						return
					}
				}
			}()

			select {
			case <-done:
			case <-time.After(15 * time.Second):
			}
		}
	}

	return &jayShellSession{
		cmd:    cmd,
		stdin:  stdin,
		reader: reader,
		marker: marker,
	}, nil
}

// Execute sends a command to the shell and waits for it to complete.
func (s *jayShellSession) Execute(_ context.Context, command string) (string, error) {
	fullCmd := fmt.Sprintf("%s; echo %s\n", command, s.marker)
	if _, err := io.WriteString(s.stdin, fullCmd); err != nil {
		return "", err
	}

	var output strings.Builder
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return output.String(), err
		}
		if strings.Contains(line, s.marker) {
			break
		}
		cleaned := stripBasicEscape(line)
		output.WriteString(cleaned)
	}

	return strings.TrimSpace(output.String()), nil
}

// Close terminates the shell session.
func (s *jayShellSession) Close() error {
	_ = s.stdin.Close()
	return s.cmd.Wait()
}

// newJayShellPool creates a ManifoldStatePool that manages persistent
// non-interactive shell sessions with proper concurrent execution.
func newJayShellPool(
	ctx context.Context,
	shell string,
	wg pants.WaitGroup,
	size, inputSize, outputSize uint,
	interval, timeout time.Duration,
) (*pants.ManifoldStatePool[string, string, pants.ShellSession], error) {
	mf := func(command string, session pants.ShellSession) (string, error) {
		return session.Execute(ctx, command)
	}

	initialiser := func(id pants.RoutineID) any {
		sess, err := newJayShellSession(shell)
		if err != nil {
			return nil
		}
		return sess
	}

	finaliser := func(state any) {
		if session, ok := state.(pants.ShellSession); ok {
			_ = session.Close()
		}
	}

	return pants.NewManifoldStatePool(ctx, mf, wg,
		pants.WithSize(size),
		pants.WithInput(inputSize),
		pants.WithOutput(outputSize, interval, timeout),
		pants.WithStateInitialiser(initialiser),
		pants.WithStateFinaliser(finaliser),
	)
}

// shellPoolExecutor wraps a ManifoldStatePool and provides a synchronous
// Execute method that submits commands to the pool and waits for results.
type shellPoolExecutor struct {
	pool    *pants.ManifoldStatePool[string, string, pants.ShellSession]
	counter uint64
	once    sync.Once
	done    chan struct{}
	mux     sync.Mutex
	pending map[string]chan shellResult
}

func newShellPoolExecutor(pool *pants.ManifoldStatePool[string, string, pants.ShellSession]) *shellPoolExecutor {
	return &shellPoolExecutor{
		pool:    pool,
		done:    make(chan struct{}),
		pending: make(map[string]chan shellResult),
	}
}

func (e *shellPoolExecutor) Execute(
	ctx context.Context,
	command string,
) ([]byte, error) {
	id := strconv.FormatUint(atomic.AddUint64(&e.counter, 1), 36)
	marker := fmt.Sprintf("__JAYWALK_SHELL_STATUS_%s__:", id)
	resultCh := make(chan shellResult, 1)

	e.once.Do(func() {
		go e.observe()
	})

	e.mux.Lock()
	e.pending[marker] = resultCh
	e.mux.Unlock()

	if err := e.pool.Post(ctx, wrapShellCommand(command, marker)); err != nil {
		e.mux.Lock()
		delete(e.pending, marker)
		e.mux.Unlock()

		return nil, err
	}

	select {
	case result := <-resultCh:
		return result.output, result.err

	case <-ctx.Done():
		e.mux.Lock()
		delete(e.pending, marker)
		e.mux.Unlock()

		return nil, ctx.Err()

	case <-e.done:
		return nil, nil
	}
}

func (e *shellPoolExecutor) observe() {
	defer close(e.done)

	for output := range e.pool.Observe() {
		payload := output.Payload

		e.mux.Lock()
		for marker, resultCh := range e.pending {
			if result, ok := parseShellResult(payload, marker, output.Error); ok {
				delete(e.pending, marker)
				resultCh <- result
				close(resultCh)
				break
			}
		}
		e.mux.Unlock()
	}
}

func (e *shellPoolExecutor) closeAll() {
	e.pool.Conclude(context.Background())
}

func wrapShellCommand(command, marker string) string {
	return fmt.Sprintf("{\n%s\n} 2>&1\n__jaywalk_status=$?\nprintf '\\n%s%%d\\n' \"$__jaywalk_status\"",
		command,
		marker,
	)
}

// stripBasicEscape removes OSC terminal title sequences from a line.
// These are emitted by zsh+powerlevel10k even in non-interactive mode.
func stripBasicEscape(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b {
			i++
			if i >= len(s) {
				break
			}
			switch s[i] {
			case ']': // OSC: skip until BEL or ESC\
				for i < len(s) {
					if s[i] == 0x07 {
						break
					}
					if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '\\' {
						i++
						break
					}
					i++
				}
			default:
				// skip single-char escape sequences
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}

func parseShellResult(
	payload, marker string,
	err error,
) (shellResult, bool) {
	index := strings.LastIndex(payload, marker)
	if index < 0 {
		return shellResult{}, false
	}

	body := strings.TrimSuffix(payload[:index], "\n")
	statusText := strings.TrimSpace(payload[index+len(marker):])
	if fields := strings.Fields(statusText); len(fields) > 0 {
		statusText = fields[0]
	}

	if status, convErr := strconv.Atoi(statusText); convErr == nil && status != 0 && err == nil {
		err = fmt.Errorf("shell command exited with status %d", status)
	}

	return shellResult{
		output: []byte(body),
		err:    err,
	}, true
}
