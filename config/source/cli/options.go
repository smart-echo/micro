package cli

import (
	"context"

	"github.com/smart-echo/micro/config/source"
	"github.com/urfave/cli/v2"
)

type contextKey struct{}

// Context sets the cli context.
func Context(c *cli.Context) source.Option {
	return func(o *source.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, contextKey{}, c)
	}
}
