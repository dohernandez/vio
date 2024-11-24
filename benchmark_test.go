package vio_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/bool64/ctxd"
	"github.com/bool64/httptestbench"
	"github.com/dohernandez/goservicing"
	"github.com/dohernandez/servers"
	"github.com/dohernandez/vio/internal/domain/model"
	"github.com/dohernandez/vio/internal/platform/app"
	"github.com/dohernandez/vio/internal/platform/config"
	"github.com/dohernandez/vio/internal/platform/helpers"
	"github.com/dohernandez/vio/internal/platform/storage"
	"github.com/dohernandez/vio/pkg/must"
	"github.com/nhatthm/clockdog"
	"github.com/valyala/fasthttp"
)

//nolint:gochecknoinits // Initializing resource for multiple benchmarks.
func init() {
	ctx := context.Background()

	// load configurations
	err := config.WithEnvFiles(".env.integration-test")
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load env from .env.integration-test"))
	cfg, err = config.GetConfig()
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	cfg.Environment = "test"
	cfg.Log.Output = io.Discard

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

// For performance.
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
					func(_ int, req *fasthttp.Request) {
						req.SetRequestURI(requestURI)
					},
					func(_ int, resp *fasthttp.Response) bool {
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

// getBenchmark benchmark for {/v1/}.
func getBenchmark() []benchmarkItem {
	ctx := context.Background()

	cleanDatabase(ctx)

	// Load sample data
	// 200.106.141.15,SI,Nepal,DuBuquemouth,-84.87503094689836,7.206435933364332,7823011346
	// 160.103.7.140,CZ,Nicaragua,New Neva,-68.31023296602508,-37.62435199624531,7301823115
	// 70.95.73.73,TL,Saudi Arabia,Gradymouth,-49.16675918861615,-86.05920084416894,2559997162
	data, err := helpers.LoadSampleData(3, 0)
	if err != nil {
		panic(err)
	}

	geos := make([]any, 0, len(data))

	for _, d := range data {
		geolocation, err := model.DecodeGeolocation(d)
		if err != nil {
			panic(err)
		}

		geos = append(geos, geolocation)
	}

	loadDatabase(ctx, geos)

	return []benchmarkItem{
		{
			name: "GeolocationByIPExposer",
			uri:  "/v1/geolocations/200.106.141.15",
		},
	}
}

func cleanDatabase(ctx context.Context) {
	// Deleting from table
	_, err := deps.Storage.Exec(
		ctx,
		deps.Storage.DeleteStmt(storage.GeolocationTable),
	)
	must.NotFail(ctxd.WrapError(ctx, err, "failed cleaning table"))
}

func loadDatabase(ctx context.Context, data []any) {
	for _, d := range data {
		_, err := deps.Storage.Exec(
			ctx,
			deps.Storage.InsertStmt(storage.GeolocationTable, d),
		)
		must.NotFail(ctxd.WrapError(ctx, err, "failed loading", "data", d))
	}
}
