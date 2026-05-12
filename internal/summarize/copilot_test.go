package summarize

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
)

type fakeCmd struct {
	run func(stdout, stderr io.Writer) error
}

func (c fakeCmd) Run(stdout, stderr io.Writer) error {
	if c.run == nil {
		return nil
	}
	return c.run(stdout, stderr)
}

func TestSummarizeWithCopilotBuildsExpectedCommand(t *testing.T) {
	original := execCommandContext
	defer func() { execCommandContext = original }()

	var gotName string
	var gotArgs []string
	execCommandContext = func(ctx context.Context, name string, args ...string) commandRunner {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return fakeCmd{run: func(stdout, stderr io.Writer) error {
			_, _ = io.WriteString(stdout, "summary")
			return nil
		}}
	}

	got, err := summarizeWithCopilot(context.Background(), "content", "prompt", "gpt-5-mini")
	if err != nil {
		t.Fatalf("summarizeWithCopilot() error = %v", err)
	}
	if got != "summary" {
		t.Fatalf("summary = %q, want %q", got, "summary")
	}
	if gotName != "copilot" {
		t.Fatalf("command = %q, want copilot", gotName)
	}
	wantArgs := []string{"-p", systemPrompt + "\n\n" + buildUserPrompt("content", "prompt"), "-s", "--model", "gpt-5-mini"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
}

func TestSummarizeWithCopilotReturnsStderrOnFailure(t *testing.T) {
	original := execCommandContext
	defer func() { execCommandContext = original }()

	execCommandContext = func(ctx context.Context, name string, args ...string) commandRunner {
		return fakeCmd{run: func(stdout, stderr io.Writer) error {
			_, _ = io.WriteString(stderr, "No authentication information found")
			return errors.New("exit status 1")
		}}
	}

	got, err := summarizeWithCopilot(context.Background(), "content", "prompt", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if got != "content" {
		t.Fatalf("fallback = %q, want original content", got)
	}
	if !bytes.Contains([]byte(err.Error()), []byte("No authentication information found")) {
		t.Fatalf("error = %q, want stderr text", err.Error())
	}
}

func TestSummarizeWithCopilotEmptyOutput(t *testing.T) {
	original := execCommandContext
	defer func() { execCommandContext = original }()

	execCommandContext = func(ctx context.Context, name string, args ...string) commandRunner {
		return fakeCmd{run: func(stdout, stderr io.Writer) error { return nil }}
	}

	got, err := summarizeWithCopilot(context.Background(), "content", "prompt", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if got != "content" {
		t.Fatalf("fallback = %q, want original content", got)
	}
}
