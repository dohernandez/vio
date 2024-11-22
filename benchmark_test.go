//go:build bench
// +build bench

package kit_template_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/bool64/ctxd"
	"github.com/bool64/httptestbench"
	"github.com/dohernandez/kit-template/internal/platform/app"
	"github.com/dohernandez/kit-template/internal/platform/config"
	grpcRest "github.com/dohernandez/kit-template/pkg/grpc/rest"
	grpcServer "github.com/dohernandez/kit-template/pkg/grpc/server"
	"github.com/dohernandez/kit-template/pkg/must"
	"github.com/dohernandez/kit-template/pkg/servicing"
	"github.com/nhatthm/clockdog"
	"github.com/valyala/fasthttp"
)

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

// getBenchmark benchmark for {/v1/}
func getBenchmark() []benchmarkItem {
	ctx := context.Background()

	cleanDatabase(ctx)

	data := nil

	loadDatabase(ctx, data)

	return []benchmarkItem{}
}

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

	deps, err = app.NewServiceLocator(cfg, func(l *app.Locator) {
		l.ClockProvider = clock
	})
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init service locator"))
}

func cleanDatabase(ctx context.Context) {
	// Deleting from table
	_, err := deps.Storage.Exec(
		ctx,
		deps.Storage.DeleteStmt("table"),
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed cleaning table"))
}

func loadDatabase(ctx context.Context, data interface{}) {
	for _, d := range data {
		_, err := deps.Storage.Exec(
			context.Background(),
			deps.Storage.InsertStmt("table", d),
		)
		must.NotFail(ctxd.WrapError(ctx, err, "failed loading", "data", d))
	}
}

func intServices(ctx context.Context) (string, *servicing.ServiceGroup) {
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppGRPCPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init GRPC service listener"))

	srvGRPC := grpcServer.InitGRPCService(
		ctx,
		grpcServer.InitGRPCServiceConfig{
			Listener:       grpcListener,
			Service:        deps.BidTrackerService,
			Logger:         deps.ZapLogger(),
			UInterceptor:   deps.GRPCUnitaryInterceptors,
			WithReflective: cfg.IsDev(),
			Options: []grpcServer.Option{
				grpcServer.WithAddrAssigned(),
			},
		},
	)

	restTListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppRESTPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service listener"))

	srvREST, err := grpcRest.InitRESTService(
		ctx,
		grpcRest.InitRESTServiceConfig{
			Listener:         restTListener,
			Service:          deps.BidTrackerRESTService,
			UInterceptor:     deps.GRPCUnitaryInterceptors,
			Handlers:         deps.Handlers,
			ResponseModifier: deps.ResponseModifier,
			Options: []grpcRest.Option{
				grpcRest.WithAddrAssigned(),
			},
		},
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service"))

	services := servicing.WithGracefulSutDown(
		func(ctx context.Context) {
			app.GracefulDBShutdown(ctx, deps)
		},
	)

	go func() {
		err = services.Start(ctx,
			func(ctx context.Context, msg string) {
				deps.CtxdLogger().Important(ctx, msg)
			},
			srvGRPC,
			srvREST,
		)
		must.NotFail(ctxd.WrapError(ctx, err, "failed to start the services"))
	}()

	baseRESTURL := <-srvREST.AddrAssigned

	return baseRESTURL, services
}
