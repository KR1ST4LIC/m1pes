package logging

import (
	"context"
	"errors"
	"log/slog"
)

type SlogWrapper struct {
	next slog.Handler
}

func NewSlogWrapper(next slog.Handler) slog.Handler {
	return &SlogWrapper{next: next}
}

type LogCtx struct {
	userId  int64
	orderId string
	coinTag string
}

var LogCtxKey = "LogCtxKey"

func (s SlogWrapper) Enabled(ctx context.Context, level slog.Level) bool {
	return s.next.Enabled(ctx, level)
}

func (s SlogWrapper) Handle(ctx context.Context, record slog.Record) error {
	if c, ok := ctx.Value(LogCtxKey).(LogCtx); ok {
		if c.userId != 0 {
			record.Add("userId", c.userId)
		}
		if c.orderId != "" {
			record.Add("orderId", c.orderId)
		}
		if c.coinTag != "" {
			record.Add("coinTag", c.coinTag)
		}
	}
	return s.next.Handle(ctx, record)
}

func (s SlogWrapper) WithAttrs(attrs []slog.Attr) slog.Handler {
	return s.next.WithAttrs(attrs)
}

func (s SlogWrapper) WithGroup(name string) slog.Handler {
	return s.next.WithGroup(name)
}

func WithUserId(ctx context.Context, userId int64) context.Context {
	if c, ok := ctx.Value(LogCtxKey).(LogCtx); ok {
		c.userId = userId
		return context.WithValue(ctx, LogCtxKey, c)
	}
	return context.WithValue(ctx, LogCtxKey, LogCtx{userId: userId})
}

func WithOrderId(ctx context.Context, orderId string) context.Context {
	if c, ok := ctx.Value(LogCtxKey).(LogCtx); ok {
		c.orderId = orderId
		return context.WithValue(ctx, LogCtxKey, c)
	}
	return context.WithValue(ctx, LogCtxKey, LogCtx{orderId: orderId})
}

func WithCoinTag(ctx context.Context, coinTag string) context.Context {
	if c, ok := ctx.Value(LogCtxKey).(LogCtx); ok {
		c.coinTag = coinTag
		return context.WithValue(ctx, LogCtxKey, c)
	}
	return context.WithValue(ctx, LogCtxKey, LogCtx{coinTag: coinTag})
}

type ErrorWithCtx struct {
	next error
	ctx  LogCtx
}

func (e *ErrorWithCtx) Error() string {
	return e.next.Error()
}

func WrapError(ctx context.Context, err error) error {
	c := LogCtx{}
	if x, ok := ctx.Value(LogCtxKey).(LogCtx); ok {
		c = x
	}
	return &ErrorWithCtx{next: err, ctx: c}
}

func ErrorCtx(ctx context.Context, err error) context.Context {
	var errCtx *ErrorWithCtx
	if errors.As(err, &errCtx) {
		return context.WithValue(ctx, LogCtxKey, errCtx.ctx)
	}
	return ctx
}
