package logwise

import (
	"log/slog"
	"strings"
	"sync"

	"go.uber.org/zap"
	xslog "golang.org/x/exp/slog"

	"github.com/rumorshub/ioc/configwise"
	"github.com/rumorshub/ioc/errs"
)

var _ Logger = (*Log)(nil)

type Logger interface {
	NamedLogger(name string) *slog.Logger
	NamedXLogger(name string) *xslog.Logger
	NamedZapLogger(name string) *zap.Logger
}

type Log struct {
	mu sync.RWMutex

	attrs    []slog.Attr
	base     *slog.Logger
	channels ChannelConfig

	syncs []HandlerSyncer
}

func NewLogger(cfg Config, channels ChannelConfig, attrs ...slog.Attr) *Log {
	base := errs.Must(cfg.Logger(attrs...))

	slog.SetDefault(base)

	return &Log{
		attrs:    attrs,
		channels: channels,
		base:     base,
		syncs:    []HandlerSyncer{base.Handler().(HandlerSyncer)},
	}
}

func (l *Log) NamedLogger(name string) *slog.Logger {
	if cfg, ok := l.channels.Channels[name]; ok {
		log := errs.Must(cfg.Logger(l.attrs...))

		l.mu.Lock()
		defer l.mu.Unlock()

		l.syncs = append(l.syncs, log.Handler().(HandlerSyncer))

		return log.WithGroup(name)
	}

	return l.base.WithGroup(name)
}

func (l *Log) NamedXLogger(name string) *xslog.Logger {
	return xslog.New(NewXHandlerWrapper(l.NamedLogger(name).Handler()))
}

func (l *Log) NamedZapLogger(name string) *zap.Logger {
	return NewZap(l.NamedLogger(name))
}

func (l *Log) Sync() (err error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, hs := range l.syncs {
		err = errs.Append(err, hs.Sync())
	}
	if syncer, ok := l.base.Handler().(HandlerSyncer); ok {
		err = errs.Append(err, syncer.Sync())
	}
	return
}

func ToLeveler(level string) slog.Leveler {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func ToAttrs(data map[string]any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(data))
	for key, value := range data {
		attrs = append(attrs, slog.Any(key, value))
	}
	return attrs
}

func (cfg *Config) Logger(attrs ...slog.Attr) (*slog.Logger, error) {
	syncer, err := cfg.OpenSinks()
	if err != nil {
		return nil, err
	}

	handler := cfg.Opts().NewHandler(syncer, cfg.Encoding)
	handler = handler.WithAttrs(append(ToAttrs(cfg.Attrs), attrs...))

	return slog.New(NewHandlerSyncer(syncer, handler)), nil
}

func NewChannelConfig(cfg configwise.Configurer, key string) (c ChannelConfig, err error) {
	if cfg.Has(key) {
		err = cfg.UnmarshalKey(key, &c)
	} else {
		c.Channels = make(map[string]Config)
	}
	return
}

func NewConfig(cfg configwise.Configurer, key string) (c Config, err error) {
	if cfg.Has(key) {
		err = cfg.UnmarshalKey(key, &c)
	}
	return
}

func Load(cfg configwise.Configurer) (*Log, error) {
	c, err := NewConfig(cfg, PluginName)
	if err != nil {
		return nil, err
	}

	channels, err := NewChannelConfig(cfg, PluginName)
	if err != nil {
		return nil, err
	}

	return NewLogger(c, channels, slog.String("version", cfg.Version())), nil
}
