package main

import (
	"context"
	"fmt"
	"github.com/dohernandez/servers"
	"net"
	"time"

	"github.com/bool64/ctxd"
	"github.com/dohernandez/goservicing"
	"github.com/dohernandez/vio/internal/platform/app"
	"github.com/dohernandez/vio/internal/platform/config"
	"github.com/dohernandez/vio/pkg/must"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// load configurations
	cfg, err := config.GetConfig()
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	metricsListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppMetricsPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init Metrics service listener"))

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppGRPCPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init GRPC service listener"))

	optReflection := func(srv any) {}

	if cfg.IsDev() {
		optReflection = servers.WithReflection()
	}

	restTListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppRESTPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service listener"))

	// initialize locator
	deps, err := app.NewServiceLocator(cfg,
		app.WithGRPCOptions(
			optReflection,
			servers.WithListener(grpcListener, true),
		),
		app.WithGRPCRestOptions(
			servers.WithListener(restTListener, true),
		),
		app.WithMetrics(
			servers.WithListener(metricsListener, true),
		),
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init locator"))

	services := goservicing.WithGracefulShutDown(
		func(ctx context.Context) {
			app.GracefulDBShutdown(ctx, deps)
		},
	)

	err = services.Start(
		ctx,
		time.Second*5,
		func(ctx context.Context, msg string) {
			deps.CtxdLogger().Important(ctx, msg)
		},
		deps.VioMetricsService,
		deps.VioService,
		deps.VioRESTService,
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to start the services"))
}
