package app

import (
	"context"
)

// GracefulDBShutdown close all db opened connection.
func GracefulDBShutdown(ctx context.Context, l *Locator) {
	if err := l.DBx.Close(); err != nil {
		l.LoggerProvider.CtxdLogger().Error(
			ctx,
			"Failed to close connection to Postgres",
			"error",
			err,
		)
	}
}
