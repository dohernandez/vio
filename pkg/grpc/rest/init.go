package rest

import (
	"context"
	"net"
	"net/http"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	mux "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// InitRESTServiceConfig REST server init configuration.
type InitRESTServiceConfig struct {
	Listener         net.Listener
	Service          ServiceServer
	UInterceptor     []grpc.UnaryServerInterceptor
	Handlers         []HandlerPathOption
	Options          []Option
	ResponseModifier func(context.Context, http.ResponseWriter, proto.Message) error
}

// InitRESTService initialize an instance of REST service based on the GRPC service.
func InitRESTService(
	ctx context.Context,
	cfg InitRESTServiceConfig,
) (*Server, error) {
	cfg.Service.WithUnaryServerInterceptor(
		grpcMiddleware.ChainUnaryServer(cfg.UInterceptor...),
	)

	opts := make([]Option, 0)

	opts = append(opts, cfg.Options...)
	opts = append(opts,
		WithListener(cfg.Listener, true),
		// use to registering point service using the point service registerer
		WithService(cfg.Service),
	)

	for _, handler := range cfg.Handlers {
		h := handler

		opts = append(opts,
			WithHandlerPathOption(func(mux *mux.ServeMux) error {
				return mux.HandlePath(h.Method, h.PathPattern, h.Handler)
			}),
		)
	}

	if cfg.ResponseModifier != nil {
		opts = append(opts,
			WithServerMuxOption(
				mux.WithForwardResponseOption(cfg.ResponseModifier),
			),
		)
	}

	return NewServer(ctx, opts...)
}
