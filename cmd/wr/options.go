package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mrvarmazyar/web-research/internal/summarize"
)

type summaryOptions struct {
	Provider string
	Model    string
	Mode     string
	TopK     int
}

func parseSummaryOptions(args []string) (summaryOptions, []string, error) {
	var opts summaryOptions
	remaining := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			remaining = append(remaining, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "--") {
			remaining = append(remaining, args[i:]...)
			break
		}
		switch {
		case arg == "--provider":
			i++
			if i >= len(args) {
				return summaryOptions{}, nil, fmt.Errorf("missing value for --provider")
			}
			opts.Provider = args[i]
		case strings.HasPrefix(arg, "--provider="):
			opts.Provider = strings.TrimPrefix(arg, "--provider=")
		case arg == "--model":
			i++
			if i >= len(args) {
				return summaryOptions{}, nil, fmt.Errorf("missing value for --model")
			}
			opts.Model = args[i]
		case strings.HasPrefix(arg, "--model="):
			opts.Model = strings.TrimPrefix(arg, "--model=")
		case arg == "--mode":
			i++
			if i >= len(args) {
				return summaryOptions{}, nil, fmt.Errorf("missing value for --mode")
			}
			opts.Mode = args[i]
		case strings.HasPrefix(arg, "--mode="):
			opts.Mode = strings.TrimPrefix(arg, "--mode=")
		case arg == "--top-k":
			i++
			if i >= len(args) {
				return summaryOptions{}, nil, fmt.Errorf("missing value for --top-k")
			}
			n, err := strconv.Atoi(args[i])
			if err != nil {
				return summaryOptions{}, nil, fmt.Errorf("--top-k must be an integer, got %q", args[i])
			}
			opts.TopK = n
		case strings.HasPrefix(arg, "--top-k="):
			val := strings.TrimPrefix(arg, "--top-k=")
			n, err := strconv.Atoi(val)
			if err != nil {
				return summaryOptions{}, nil, fmt.Errorf("--top-k must be an integer, got %q", val)
			}
			opts.TopK = n
		default:
			return summaryOptions{}, nil, fmt.Errorf("unknown option %q", arg)
		}
	}

	if err := summarize.ValidateProvider(opts.Provider); err != nil {
		return summaryOptions{}, nil, err
	}
	if err := validateModeFlag(opts.Mode); err != nil {
		return summaryOptions{}, nil, err
	}

	return opts, remaining, nil
}

func validateModeFlag(mode string) error {
	switch mode {
	case "", "summarize", "lossless", "chunks":
		return nil
	default:
		return fmt.Errorf("mode must be one of: summarize, lossless, chunks")
	}
}
