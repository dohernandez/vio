package service

import (
	"context"
	"github.com/bool64/ctxd"
	"github.com/dohernandez/servers"
	api "github.com/dohernandez/vio/pkg/proto"
	"google.golang.org/grpc"
)

type VioServiceConfig struct {
	servers.Config

	logger ctxd.Logger
}

type VioService struct {
	*servers.GRPC

	// UnimplementedVioServiceServer must be embedded to have forward compatible implementations.
	api.UnimplementedVioServiceServer

	logger ctxd.Logger
}

func NewVioService(cfg VioServiceConfig, opts ...servers.Option) *VioService {
	srv := &VioService{
		logger: cfg.logger,
	}

	opts = append(opts, servers.WithRegisterService(srv))

	srv.GRPC = servers.NewGRPC(cfg.Config, opts...)

	return srv
}

func (v *VioService) RegisterService(s grpc.ServiceRegistrar) {
	api.RegisterVioServiceServer(s, v)
}

func (v *VioService) SayHello(ctx context.Context, req *api.HelloRequest) (*api.HelloReply, error) {
	return &api.HelloReply{Message: "Hello " + req.GetName()}, nil
}
