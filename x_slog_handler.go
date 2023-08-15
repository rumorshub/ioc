package ioc

import (
	"context"
	"log/slog"

	xSlog "golang.org/x/exp/slog"
)

type xSlogHandler struct {
	inner slog.Handler
}

func (h xSlogHandler) Enabled(ctx context.Context, lvl xSlog.Level) bool {
	return h.inner.Enabled(ctx, slog.Level(lvl))
}

func (h xSlogHandler) Handle(ctx context.Context, xRecord xSlog.Record) error {
	attrs := make([]slog.Attr, 0, xRecord.NumAttrs())
	xRecord.Attrs(func(xAttr xSlog.Attr) bool {
		attrs = append(attrs, xAttrToAttr(xAttr))
		return true
	})

	record := slog.NewRecord(xRecord.Time, slog.Level(xRecord.Level), xRecord.Message, xRecord.PC)
	record.AddAttrs(attrs...)

	return h.inner.Handle(ctx, record)
}

func (h xSlogHandler) WithAttrs(xAttrs []xSlog.Attr) xSlog.Handler {
	attrs := make([]slog.Attr, len(xAttrs))
	for i, xAttr := range xAttrs {
		attrs[i] = xAttrToAttr(xAttr)
	}
	return xSlogHandler{inner: h.inner.WithAttrs(attrs)}
}

func (h xSlogHandler) WithGroup(name string) xSlog.Handler {
	return xSlogHandler{inner: h.inner.WithGroup(name)}
}

func xAttrToAttr(a xSlog.Attr) slog.Attr {
	return slog.Attr{Key: a.Key, Value: xValueToValue(a.Value)}
}

func xValueToValue(v xSlog.Value) slog.Value {
	switch v.Kind() {
	case xSlog.KindAny:
		return slog.AnyValue(v.Any())
	case xSlog.KindBool:
		return slog.BoolValue(v.Bool())
	case xSlog.KindDuration:
		return slog.DurationValue(v.Duration())
	case xSlog.KindFloat64:
		return slog.Float64Value(v.Float64())
	case xSlog.KindInt64:
		return slog.Int64Value(v.Int64())
	case xSlog.KindString:
		return slog.StringValue(v.String())
	case xSlog.KindTime:
		return slog.TimeValue(v.Time())
	case xSlog.KindUint64:
		return slog.Uint64Value(v.Uint64())
	case xSlog.KindGroup:
		group := make([]slog.Attr, len(v.Group()))
		for i, a := range v.Group() {
			group[i] = xAttrToAttr(a)
		}
		return slog.GroupValue(group...)
	case xSlog.KindLogValuer:
		return xValueToValue(v.LogValuer().LogValue())
	default:
		return slog.AnyValue(v.Any())
	}
}
