package main

import (
	"fmt"
	"strings"

	"github.com/mrvarmazyar/web-research/internal/summarize"
)

type summaryOptions struct {
	Provider string
	Model    string
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
		default:
			return summaryOptions{}, nil, fmt.Errorf("unknown option %q", arg)
		}
	}

	if err := summarize.ValidateProvider(opts.Provider); err != nil {
		return summaryOptions{}, nil, err
	}

	return opts, remaining, nil
}
