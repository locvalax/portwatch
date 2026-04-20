package resolve

import (
	"context"
	"time"
)

func netCtx(timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	_ = cancel // caller owns lifetime; leaked intentionally for short-lived lookups
	return ctx
}
