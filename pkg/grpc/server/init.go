package server

import (
	"context"
	"net"

	grpcZapLogger "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// InitGRPCServiceConfig GRPC server init configuration.
type InitGRPCServiceConfig struct {
	Listener       net.Listener
	Service        ServiceServer
	Logger         *zap.Logger
	UInterceptor   []grpc.UnaryServerInterceptor
	WithReflective bool
	Options        []Option
}

// InitGRPCService initialize an instance of grpc service, with all the instrumentation.
func InitGRPCService(
	_ context.Context,
	cfg InitGRPCServiceConfig,
) *Server {
	grpcZapLogger.ReplaceGrpcLoggerV2(cfg.Logger)

	opts := make([]Option, 0)

	opts = append(opts, cfg.Options...)
	opts = append(opts,
		WithListener(cfg.Listener, true),
		// registering point service using the point service registerer
		WithService(cfg.Service),
		ChainUnaryInterceptor(cfg.UInterceptor...),
	)

	// Enabling reflection in dev and testing env.
	if cfg.WithReflective {
		opts = append(opts, WithReflective())
	}

	return NewServer(opts...)
}
