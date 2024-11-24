package service

import (
	"context"
	"errors"
	"net"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/servers"
	"github.com/dohernandez/vio/internal/domain/model"
	"github.com/dohernandez/vio/internal/domain/usecase"
	api "github.com/dohernandez/vio/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// VioServiceConfig holds the configuration for the VioService.
type VioServiceConfig struct {
	servers.Config

	logger ctxd.Logger
}

// VioService is the gRPC service.
type VioService struct {
	*servers.GRPC

	// UnimplementedVioServiceServer must be embedded to have forward compatible implementations.
	api.UnimplementedVioServiceServer

	geoIPFinder *usecase.GeolocationByIPExposer

	logger ctxd.Logger
}

// NewVioService creates a new VioService.
func NewVioService(cfg VioServiceConfig, geoIPFinder *usecase.GeolocationByIPExposer, opts ...servers.Option) *VioService {
	srv := &VioService{
		geoIPFinder: geoIPFinder,
		logger:      cfg.logger,
	}

	if cfg.logger == nil {
		cfg.logger = ctxd.NoOpLogger{}
	}

	opts = append(opts, servers.WithRegisterService(srv))

	srv.GRPC = servers.NewGRPC(cfg.Config, opts...)

	return srv
}

// RegisterService registers the service with api.RegisterVioServiceServer registrar.
func (v *VioService) RegisterService(s grpc.ServiceRegistrar) {
	api.RegisterVioServiceServer(s, v)
}

// GeolocationByIPExposer expose the geolocation data by IP.
//
// Receives a request with the ip. Responses with the geolocation data otherwise not.
func (v *VioService) GeolocationByIPExposer(ctx context.Context, req *api.GeolocationByIPExposerRequest) (*api.GeolocationByIPExposerResponse, error) {
	if req.GetIp() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "ip address is required")
	}

	if net.ParseIP(req.GetIp()) == nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid ip address")
	}

	geo, err := v.geoIPFinder.ExposeGeolocationByIP(ctx, req.GetIp())
	if err == nil {
		return &api.GeolocationByIPExposerResponse{
			IpAddress:    geo.IPAddress,
			Country:      geo.Country,
			CountryCode:  geo.CountryCode,
			City:         geo.City,
			Latitude:     geo.Latitude,
			Longitude:    geo.Longitude,
			MysteryValue: geo.MysteryValue,
		}, nil
	}

	if errors.Is(err, model.ErrGeolocationNotFound) {
		return nil, status.Errorf(codes.NotFound, "geolocation not found")
	}

	return nil, status.Errorf(codes.Internal, "%s", err.Error())
}
