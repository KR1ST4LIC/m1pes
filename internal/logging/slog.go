package logging

import (
	"context"
	"log/slog"
)

type SlogWrapper struct {
	next slog.Handler
}

func NewSlogWrapper(next slog.Handler) slog.Handler {
	return &SlogWrapper{next: next}
}

type LogCtx struct {
	userId int64
}

func (s SlogWrapper) Enabled(ctx context.Context, level slog.Level) bool {
	return s.next.Enabled(ctx, level)
}

func (s SlogWrapper) Handle(ctx context.Context, record slog.Record) error {
	if c, ok := ctx.Value(LogCtx{}).(LogCtx); ok {
		if c.userId != 0 {
			record.Add("userId", c.userId)
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
	return context.WithValue(ctx, LogCtx{}, LogCtx{userId: userId})
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
	if x, ok := ctx.Value(LogCtx{}).(LogCtx); ok {
		c = x
	}
	return &ErrorWithCtx{next: err, ctx: c}
}

func ErrorCtx(ctx context.Context, err error) context.Context {
	if e, ok := err.(*ErrorWithCtx); ok {
		return context.WithValue(ctx, LogCtx{}, e.ctx)
	}
	return ctx
}
