package service

import (
	"context"

	"github.com/dohernandez/servers"
	api "github.com/dohernandez/vio/pkg/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// VioRESTServiceConfig holds the configuration for rest service.
type VioRESTServiceConfig struct {
	servers.Config

	GRPCAddr string
}

// VioRESTService is the rest service.
type VioRESTService struct {
	*servers.GRPCRest

	grpcAddr string
}

// NewVioRESTService creates an instance of rest Service wrapping grpc service.
func NewVioRESTService(cfg VioRESTServiceConfig, opts ...servers.Option) (*VioRESTService, error) {
	srv := &VioRESTService{
		grpcAddr: cfg.GRPCAddr,
	}

	opts = append(opts, servers.WithRegisterServiceHandler(srv))

	grpcRest, err := servers.NewGRPCRest(cfg.Config, opts...)
	if err != nil {
		return nil, err
	}

	srv.GRPCRest = grpcRest

	return srv, nil
}

// RegisterServiceHandler registers the service implementation to mux.
func (s *VioRESTService) RegisterServiceHandler(mux *runtime.ServeMux) error {
	// register rest service
	return api.RegisterVioServiceHandlerFromEndpoint(context.Background(), mux, s.grpcAddr, []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())})
}
