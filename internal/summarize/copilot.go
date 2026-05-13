package summarize

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type commandRunner interface {
	SetStdin(io.Reader)
	Run(stdout, stderr io.Writer) error
}

type execCmd struct {
	cmd *exec.Cmd
}

func (c execCmd) SetStdin(r io.Reader) {
	c.cmd.Stdin = r
}

func (c execCmd) Run(stdout, stderr io.Writer) error {
	c.cmd.Stdout = stdout
	c.cmd.Stderr = stderr
	return c.cmd.Run()
}

var execCommandContext = func(ctx context.Context, name string, args ...string) commandRunner {
	return execCmd{cmd: exec.CommandContext(ctx, name, args...)}
}

func summarizeWithCopilot(ctx context.Context, content, prompt, model string) (string, error) {
	fullPrompt := systemPrompt + "\n\n" + buildUserPrompt(content, prompt)
	args := []string{"-s"} // -s = --silent, documented for scripting/non-interactive usage.
	if strings.TrimSpace(model) != "" {
		args = append(args, "--model", model)
	}

	cmd := execCommandContext(ctx, "copilot", args...)
	cmd.SetStdin(strings.NewReader(fullPrompt))
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if err := cmd.Run(&stdout, &stderr); err != nil {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fallbackContent(content), fmt.Errorf("copilot command failed: %s", msg)
		}
		return fallbackContent(content), fmt.Errorf("copilot command failed: %w", err)
	}

	result := strings.TrimSpace(stdout.String())
	if result == "" {
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			_, _ = fmt.Fprintf(os.Stderr, "warning: copilot returned no stdout; stderr: %s\n", msg)
		}
		if msg := strings.TrimSpace(stderr.String()); msg != "" {
			return fallbackContent(content), fmt.Errorf("copilot returned no output: %s", msg)
		}
		return fallbackContent(content), fmt.Errorf("copilot returned no output")
	}

	return result, nil
}
