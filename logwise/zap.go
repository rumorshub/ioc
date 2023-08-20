package logwise

import (
	"context"
	"log/slog"
	"math"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapCore struct {
	log *slog.Logger
}

func NewZap(log *slog.Logger) *zap.Logger {
	return zap.New(NewZapCore(log))
}

func NewZapCore(log *slog.Logger) zapcore.Core {
	return &ZapCore{log: log}
}

func (c *ZapCore) Enabled(lvl zapcore.Level) bool {
	return c.log.Enabled(context.TODO(), toSlogLevel(lvl))
}

func (c *ZapCore) With(fields []zapcore.Field) zapcore.Core {
	return &ZapCore{log: c.log.With(toSlogFields(fields))}
}

func (c *ZapCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *ZapCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	log := c.log

	if ent.LoggerName != "" {
		for _, name := range strings.Split(ent.LoggerName, ".") {
			log = log.WithGroup(name)
		}
	}

	log.LogAttrs(context.TODO(), toSlogLevel(ent.Level), ent.Message, toSlogFields(fields)...)

	return nil
}

func (c *ZapCore) Sync() error {
	if s, ok := c.log.Handler().(HandlerSyncer); ok {
		return s.Sync()
	}
	return nil
}

func toSlogLevel(lvl zapcore.Level) slog.Level {
	switch lvl {
	case zapcore.DebugLevel:
		return slog.LevelDebug
	case zapcore.InfoLevel:
		return slog.LevelInfo
	case zapcore.WarnLevel:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}

func toSlogFields(fields []zapcore.Field) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(fields))
	for _, f := range fields {
		var v slog.Value

		switch f.Type {
		case zapcore.BoolType:
			v = slog.BoolValue(f.Integer == 1)
		case zapcore.DurationType:
			v = slog.DurationValue(time.Duration(f.Integer))
		case zapcore.Float64Type:
		case zapcore.Float32Type:
			v = slog.Float64Value(math.Float64frombits(uint64(f.Integer)))
		case zapcore.Int64Type:
		case zapcore.Int32Type:
		case zapcore.Int16Type:
		case zapcore.Int8Type:
			v = slog.Int64Value(f.Integer)
		case zapcore.StringType:
			v = slog.StringValue(f.String)
		case zapcore.TimeType:
			if f.Interface != nil {
				v = slog.TimeValue(time.Unix(0, f.Integer).In(f.Interface.(*time.Location)))
			} else {
				// Fall back to UTC if location is nil.
				v = slog.TimeValue(time.Unix(0, f.Integer))
			}
		case zapcore.TimeFullType:
			v = slog.TimeValue(f.Interface.(time.Time))
		case zapcore.Uint64Type:
		case zapcore.Uint32Type:
		case zapcore.Uint16Type:
		case zapcore.Uint8Type:
		case zapcore.UintptrType:
			v = slog.Uint64Value(uint64(f.Integer))
		default:
			v = slog.AnyValue(f.Interface)
		}

		attrs = append(attrs, slog.Attr{
			Key:   f.Key,
			Value: v,
		})
	}
	return attrs
}
