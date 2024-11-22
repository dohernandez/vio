package main

import (
	"context"
	"fmt"
	"net"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/kit-template/internal/platform/app"
	"github.com/dohernandez/kit-template/internal/platform/config"
	grpcMetrics "github.com/dohernandez/kit-template/pkg/grpc/metrics"
	grpcRest "github.com/dohernandez/kit-template/pkg/grpc/rest"
	grpcServer "github.com/dohernandez/kit-template/pkg/grpc/server"
	"github.com/dohernandez/kit-template/pkg/must"
	"github.com/dohernandez/kit-template/pkg/servicing"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// load configurations
	cfg, err := config.GetConfig()
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	metricsListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppMetricsPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init Metrics service listener"))

	srvMetrics := grpcMetrics.NewMetricsService(ctx, metricsListener)

	// initialize locator
	deps, err := app.NewServiceLocator(cfg, func(l *app.Locator) {
		l.GRPCUnitaryInterceptors = append(l.GRPCUnitaryInterceptors,
			// adding metrics
			srvMetrics.ServerMetrics().UnaryServerInterceptor(),
		)
	})
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init locator"))

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppGRPCPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init GRPC service listener"))

	srvGRPC := grpcServer.InitGRPCService(
		ctx,
		grpcServer.InitGRPCServiceConfig{
			Listener:       grpcListener,
			Service:        deps.KitTemplateService,
			Logger:         deps.ZapLogger(),
			UInterceptor:   deps.GRPCUnitaryInterceptors,
			WithReflective: cfg.IsDev(),
			Options: []grpcServer.Option{
				grpcServer.WithMetrics(srvMetrics.ServerMetrics()),
			},
		},
	)

	restTListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppRESTPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service listener"))

	srvREST, err := grpcRest.InitRESTService(
		ctx,
		grpcRest.InitRESTServiceConfig{
			Listener:         restTListener,
			Service:          deps.KitTemplateRESTService,
			UInterceptor:     deps.GRPCUnitaryInterceptors,
			Handlers:         deps.Handlers,
			ResponseModifier: deps.ResponseModifier,
		},
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service"))

	services := servicing.WithGracefulSutDown(
		func(ctx context.Context) {
			app.GracefulDBShutdown(ctx, deps)
		},
	)

	err = services.Start(
		ctx,
		func(ctx context.Context, msg string) {
			deps.CtxdLogger().Important(ctx, msg)
		},
		srvMetrics,
		srvGRPC,
		srvREST,
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to start the services"))
}
