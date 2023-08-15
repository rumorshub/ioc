package configwise

import "github.com/roadrunner-server/endure/v2/dep"

const PluginName string = "config"

type Plugin struct {
	Cfg Configurer
}

func (p *Plugin) Init() error {
	return nil
}

func (p *Plugin) Provides() []*dep.Out {
	return []*dep.Out{
		dep.Bind((*Configurer)(nil), p.Configurer),
	}
}

func (p *Plugin) Configurer() Configurer {
	return p.Cfg
}

func (p *Plugin) Name() string {
	return PluginName
}
