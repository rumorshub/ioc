package logwise

import (
	"context"
	"os"
	"strings"

	"github.com/roadrunner-server/endure/v2/dep"
)

const PluginName = "log"

type Plugin struct {
	Log Logger
}

func (p *Plugin) Init() error {
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop(context.Context) error {
	if log, ok := p.Log.(*Log); ok {
		if err := log.Sync(); err != nil {
			e := err.Error()
			if !strings.Contains(e, os.Stderr.Name()) && !strings.Contains(e, os.Stdout.Name()) {
				return err
			}
		}
	}
	return nil
}

func (p *Plugin) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*Logger)(nil), p.Logger),
	}
}

func (p *Plugin) Logger() Logger {
	return p.Log
}

func (p *Plugin) Name() string {
	return PluginName
}
