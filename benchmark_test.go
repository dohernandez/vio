//go:build bench
// +build bench

package vio_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bool64/ctxd"
	"github.com/bool64/httptestbench"
	"github.com/dohernandez/goservicing"
	"github.com/dohernandez/servers"
	"github.com/dohernandez/vio/internal/platform/app"
	"github.com/dohernandez/vio/internal/platform/config"
	"github.com/dohernandez/vio/pkg/must"
	"github.com/nhatthm/clockdog"
	"github.com/valyala/fasthttp"
)

// nolint:gochecknoinits // Initializing resource for multiple benchmarks.
func init() {
	ctx := context.Background()

	// load configurations
	err := config.WithEnvFiles(".env.integration-test")
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load env from .env.integration-test"))
	cfg, err = config.GetConfig()
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	cfg.Environment = "test"
	cfg.Log.Output = ioutil.Discard

	clock := clockdog.New()

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppGRPCPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init GRPC service listener"))

	restTListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppRESTPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service listener"))

	deps, err = app.NewServiceLocator(
		cfg,
		func(l *app.Locator) {
			l.ClockProvider = clock
		},
		app.WithGRPCOptions(
			servers.WithAddrAssigned(),
			servers.WithReflection(),
			servers.WithListener(grpcListener, true),
		),
		app.WithGRPCRestOptions(
			servers.WithAddrAssigned(),
			servers.WithListener(restTListener, true),
		),
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init service locator"))
}

var (
	deps *app.Locator
	cfg  *config.Config
)

type benchmarkItem struct {
	name string
	uri  string
}

// nolint:dupl // For performance.
func BenchmarkIntegration(b *testing.B) {
	baseURL, services := intServices(context.Background())
	defer func() {
		must.NotFail(services.Close())
	}()

	baseURL = strings.Replace(baseURL, "[::]", "127.0.0.1", 1)

	tests := []struct {
		benchmark func() []benchmarkItem
	}{
		{
			benchmark: getBenchmark,
		},
	}

	for _, tt := range tests {
		bis := tt.benchmark()

		for _, bi := range bis {
			requestURI := "http://" + baseURL + bi.uri

			b.Run(bi.name, func(b *testing.B) {
				httptestbench.RoundTrip(b, 50,
					func(i int, req *fasthttp.Request) {
						req.SetRequestURI(requestURI)
					},
					func(i int, resp *fasthttp.Response) bool {
						return resp.StatusCode() == http.StatusOK
					},
				)
			})
		}
	}
}

func intServices(ctx context.Context) (string, *goservicing.ServiceGroup) {
	services := goservicing.WithGracefulShutDown(
		func(ctx context.Context) {
			app.GracefulDBShutdown(ctx, deps)
		},
	)

	go func() {
		err := services.Start(
			ctx,
			time.Second*5,
			func(ctx context.Context, msg string) {
				deps.CtxdLogger().Important(ctx, msg)
			},
			deps.VioService,
			deps.VioRESTService,
		)
		must.NotFail(ctxd.WrapError(ctx, err, "failed to start the services"))
	}()

	baseRESTURL := <-deps.VioRESTService.AddrAssigned

	return baseRESTURL, services
}

// getBenchmark benchmark for {/v1/}
func getBenchmark() []benchmarkItem {
	//ctx := context.Background()

	//cleanDatabase(ctx)
	//
	//data := nil
	//
	//loadDatabase(ctx, data)

	return []benchmarkItem{
		{
			name: "SayHello",
			uri:  "/",
		},
	}
}

//
//func cleanDatabase(ctx context.Context) {
//	// Deleting from table
//	_, err := deps.Storage.Exec(
//		ctx,
//		deps.Storage.DeleteStmt("table"),
//	)
//	must.NotFail(ctxd.WrapError(ctx, err, "failed cleaning table"))
//}
//
//func loadDatabase(ctx context.Context, data interface{}) {
//	for _, d := range data {
//		_, err := deps.Storage.Exec(
//			context.Background(),
//			deps.Storage.InsertStmt("table", d),
//		)
//		must.NotFail(ctxd.WrapError(ctx, err, "failed loading", "data", d))
//	}
//}
