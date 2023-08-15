package ioc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/roadrunner-server/endure/v2"
	rrErrs "github.com/roadrunner-server/errors"
	xSlog "golang.org/x/exp/slog"

	"github.com/rumorshub/ioc/configwise"
	"github.com/rumorshub/ioc/errs"
	"github.com/rumorshub/ioc/logwise"
)

const endureKey = "endure"

var ErrContainerNotFound = errors.New("container not found in context, call WithContainer")

type Container struct {
	inner   *endure.Endure
	cfg     configwise.Configurer
	log     *slog.Logger
	plugins []interface{}
}

type containerKey struct{}

func init() {
	rrErrs.Separator = " |> "
}

func NewContainer(cfg configwise.Configurer, log logwise.Logger) *Container {
	return &Container{
		cfg: cfg,
		log: log.NamedLogger(endureKey),
		plugins: []interface{}{
			&configwise.Plugin{Cfg: cfg},
			&logwise.Plugin{Log: log},
		},
	}
}

func (c *Container) RegisterAll(plugins ...interface{}) {
	c.plugins = append(c.plugins, plugins...)
}

func (c *Container) Init() error {
	const op = rrErrs.Op("container_init")

	cfg, err := NewConfig(c.cfg, endureKey)
	if err != nil {
		return rrErrs.E(op, err)
	}

	c.cfg.SetGracefulTimeout(cfg.GracePeriod)

	opts := []endure.Options{
		endure.GracefulShutdownTimeout(cfg.GracePeriod),
		endure.LogHandler(xSlogHandler{inner: c.log.Handler()}),
	}

	if cfg.PrintGraph {
		opts = append(opts, endure.Visualize())
	}

	c.inner = endure.New(xSlog.LevelError, opts...)

	if err = c.inner.RegisterAll(c.plugins...); err != nil {
		return rrErrs.E(op, err)
	}

	if err = c.inner.Init(); err != nil {
		if errs.IsSuccess(err) {
			return err
		}
		return rrErrs.E(op, err)
	}
	return nil
}

func (c *Container) Serve() (<-chan *endure.Result, error) {
	const op = rrErrs.Op("container_run")

	errCh, err := c.inner.Serve()
	if err != nil {
		if errs.IsSuccess(err) {
			return nil, err
		}
		return nil, rrErrs.E(op, err)
	}
	return errCh, nil
}

func (c *Container) Stop() error {
	const op = rrErrs.Op("container_stop")

	if err := c.inner.Stop(); err != nil {
		return rrErrs.E(op, err)
	}
	return nil
}

func (c *Container) Run() error {
	if err := c.Init(); err != nil {
		return errs.Go(err)
	}

	errCh, err := c.Serve()
	if err != nil {
		return errs.Go(err)
	}

	oss, stop := make(chan os.Signal, 5), make(chan struct{}, 1) //nolint:gomnd
	signal.Notify(oss, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-oss

		stop <- struct{}{}

		<-oss
		c.log.Info("exit forced")
		os.Exit(1)
	}()

	for {
		select {
		case e := <-errCh:
			if errs.IsSuccess(e.Error) {
				return c.Stop()
			}

			err1 := fmt.Errorf("plugin: %s. %w", e.VertexID, e.Error)
			err2 := c.Stop()

			return errs.Append(err1, err2)
		case <-stop:
			c.log.Info(fmt.Sprintf("stop signal received, grace timeout is: %0.f seconds", c.cfg.GracefulTimeout().Seconds()))

			return c.Stop()
		}
	}
}

func WithContainer(ctx context.Context, container *Container) context.Context {
	return context.WithValue(ctx, containerKey{}, container)
}

func Run(ctx context.Context, plugins ...interface{}) error {
	if container, ok := ctx.Value(containerKey{}).(*Container); ok {
		container.RegisterAll(plugins...)

		return container.Run()
	}
	return ErrContainerNotFound
}
