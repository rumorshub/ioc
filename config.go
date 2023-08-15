package ioc

import (
	"time"

	"github.com/rumorshub/ioc/configwise"
)

type Config struct {
	GracePeriod time.Duration
	PrintGraph  bool
}

// NewConfig creates endure container configuration.
func NewConfig(cfg configwise.Configurer, key string) (*Config, error) {
	if !cfg.Has(key) {
		return &Config{
			GracePeriod: configwise.DefaultGracefulTimeout,
			PrintGraph:  false,
		}, nil
	}

	cfgEndure := struct {
		GracePeriod time.Duration `mapstructure:"grace_period"`
		PrintGraph  bool          `mapstructure:"print_graph"`
	}{}

	if err := cfg.UnmarshalKey(key, &cfgEndure); err != nil {
		return nil, err
	}

	if cfgEndure.GracePeriod == 0 {
		cfgEndure.GracePeriod = configwise.DefaultGracefulTimeout
	}

	return &Config{
		GracePeriod: cfgEndure.GracePeriod,
		PrintGraph:  cfgEndure.PrintGraph,
	}, nil
}
