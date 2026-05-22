package research

import "context"

type progressKey struct{}

// WithProgress returns a context that delivers progress messages to fn.
func WithProgress(ctx context.Context, fn func(string)) context.Context {
	return context.WithValue(ctx, progressKey{}, fn)
}

func reportProgress(ctx context.Context, msg string) {
	if fn, ok := ctx.Value(progressKey{}).(func(string)); ok && fn != nil {
		fn(msg)
	}
}
