package api

import (
	"context"

	"github.com/smart-echo/micro/api/handler"
	"github.com/smart-echo/micro/api/handler/rpc"
	"github.com/smart-echo/micro/api/router/registry"
	"github.com/smart-echo/micro/api/server"
	"github.com/smart-echo/micro/api/server/http"
)

type api struct {
	options Options

	server server.Server
}

func newApi(opts ...Option) Api {
	options := NewOptions(opts...)

	rtr := options.Router

	if rtr == nil {
		// TODO: make configurable
		rtr = registry.NewRouter()
	}

	// TODO: make configurable
	hdlr := rpc.NewHandler(
		handler.WithRouter(rtr),
	)

	// TODO: make configurable
	// create a new server
	srv := http.NewServer(options.Address)

	// TODO: allow multiple handlers
	// define the handler
	srv.Handle("/", hdlr)

	return &api{
		options: options,
		server:  srv,
	}
}

func (a *api) Init(opts ...Option) error {
	for _, o := range opts {
		o(&a.options)
	}
	return nil
}

// Get the options.
func (a *api) Options() Options {
	return a.options
}

// Register a http handler.
func (a *api) Register(*Endpoint) error {
	return nil
}

// Register a route.
func (a *api) Deregister(*Endpoint) error {
	return nil
}

func (a *api) Run(ctx context.Context) error {
	if err := a.server.Start(); err != nil {
		return err
	}

	// wait to finish
	<-ctx.Done()

	if err := a.server.Stop(); err != nil {
		return err
	}

	return nil
}

func (a *api) String() string {
	return "http"
}