package vio_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/bool64/ctxd"
	"github.com/bool64/dbdog"
	"github.com/bool64/httpdog"
	"github.com/bool64/sqluct"
	"github.com/cucumber/godog"
	"github.com/dohernandez/goservicing"
	"github.com/dohernandez/servers"
	"github.com/dohernandez/vio/internal/domain/model"
	"github.com/dohernandez/vio/internal/platform/app"
	"github.com/dohernandez/vio/internal/platform/cli"
	"github.com/dohernandez/vio/internal/platform/config"
	stplatform "github.com/dohernandez/vio/internal/platform/storage"
	"github.com/dohernandez/vio/pkg/must"
	"github.com/dohernandez/vio/pkg/test/feature"
	fcli "github.com/dohernandez/vio/pkg/test/feature/cli"
	dbdogcleaner "github.com/dohernandez/vio/pkg/test/feature/database"
	"github.com/nhatthm/clockdog"
)

func TestIntegration(t *testing.T) {
	ctx := context.Background()

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// load configurations
	err := config.WithEnvFiles(".env.integration-test")
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load env from .env.integration-test"))
	cfg, err := config.GetConfig()
	must.NotFail(ctxd.WrapError(ctx, err, "failed to load configurations"))

	cfg.Environment = "test"
	cfg.Log.Output = io.Discard

	clock := clockdog.New()

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppGRPCPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init GRPC service listener"))

	restTListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AppRESTPort))
	must.NotFail(ctxd.WrapError(ctx, err, "failed to init REST service listener"))

	deps, err := app.NewServiceLocator(
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

	dbm := initDBManager(deps.Storage)
	dbmCleaner := initDBMCleaner(dbm)

	services := goservicing.WithGracefulShutDown(
		func(ctx context.Context) {
			app.GracefulDBShutdown(ctx, deps)
		},
	)

	go func() {
		err = services.Start(
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
	local := httpdog.NewLocal(baseRESTURL)

	fcliApp := fcli.App{}
	fcliApp.Add("parse", cli.NewCliApp)

	feature.RunFeatures(t, "features", func(_ *testing.T, s *godog.ScenarioContext) {
		local.RegisterSteps(s)

		dbm.RegisterSteps(s)
		dbmCleaner.RegisterSteps(s)

		clock.RegisterContext(s)

		fcli.RegisterContext(s, &fcliApp)
	})

	must.NotFail(services.Close())
}

func initDBManager(storage *sqluct.Storage) *dbdog.Manager {
	tableMapper := dbdog.NewTableMapper()

	dbm := dbdog.Manager{
		TableMapper: tableMapper,
	}

	dbm.Instances = map[string]dbdog.Instance{
		"postgres": {
			Storage: storage,
			Tables: map[string]interface{}{
				stplatform.GeolocationTable: new(model.Geolocation),
			},
		},
	}

	return &dbm
}

func initDBMCleaner(dbm *dbdog.Manager) *dbdogcleaner.ManagerCleaner {
	return &dbdogcleaner.ManagerCleaner{
		Manager: dbm,
	}
}
