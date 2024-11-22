package metrics

import (
	"context"
	"net"
)

// NewMetricsService creates an instance of metrics service.
func NewMetricsService(
	_ context.Context,
	listener net.Listener,
) *Server {
	opts := []Option{
		WithListener(listener, true),
	}

	return NewServer(opts...)
}
