package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dohernandez/servers"
	"github.com/dohernandez/vio/resources/swagger"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"net/http"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/bool64/zapctxd"
	"github.com/dohernandez/vio/internal/platform/config"
	"github.com/dohernandez/vio/internal/platform/service"
	grpcLogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	_ "github.com/jackc/pgx/v5/stdlib" //nolint:gci // Postgres driver
	"github.com/jmoiron/sqlx"
	"github.com/nhatthm/go-clock"
	clockSrv "github.com/nhatthm/go-clock/service"
	"github.com/opencensus-integrations/ocsql"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const driver = "pgx"

type locatorOptions struct {
	grpcOpts     []servers.Option
	grpcRestOpts []servers.Option

	enableMetrics bool
	metricsOpts   []servers.Option
}

// Option sets up service locator.
type Option func(l *Locator)

// WithGRPCOptions sets up gRPC server options.
func WithGRPCOptions(opts ...servers.Option) Option {
	return func(l *Locator) {
		l.opts.grpcOpts = append(l.opts.grpcOpts, opts...)
	}
}

// WithGRPCRestOptions sets up gRPC REST server options.
func WithGRPCRestOptions(opts ...servers.Option) Option {
	return func(l *Locator) {
		l.opts.grpcRestOpts = append(l.opts.grpcRestOpts, opts...)
	}
}

// WithMetrics sets up metrics server.
func WithMetrics(opts ...servers.Option) Option {
	return func(l *Locator) {
		l.opts.enableMetrics = true
		l.opts.metricsOpts = append(l.opts.metricsOpts, opts...)
	}
}

// Locator defines application resources.
type Locator struct {
	Config *config.Config

	opts locatorOptions

	DBx     *sqlx.DB
	Storage *sqluct.Storage

	logger *zapctxd.Logger
	ctxd.LoggerProvider

	clockSrv.ClockProvider

	restHandlers            map[string]http.Handler
	grpcUnitaryInterceptors []grpc.UnaryServerInterceptor

	VioService        *service.VioService
	VioRESTService    *service.VioRESTService
	VioMetricsService *servers.Metrics

	// use cases
}

// NewServiceLocator creates application locator.
func NewServiceLocator(cfg *config.Config, opts ...Option) (*Locator, error) {
	l := Locator{
		Config:        cfg,
		ClockProvider: clock.New(),
	}

	for _, o := range opts {
		o(&l)
	}

	l.appendStandardHandlers(cfg.ServiceName)
	//handler.SetResponseModifier(&l.Provider)

	var err error

	// logger stuff
	l.setLogger()

	// Database stuff.
	//l.Config.PostgresDB.DriverName = driver
	//
	//l.DBx, err = makeDBx(cfg.PostgresDB)
	//if err != nil {
	//	return nil, err
	//}
	//
	//l.Storage = makeStorage(l.DBx, l.CtxdLogger())

	l.setGRPCUnitaryInterceptors()

	// setting up use cases dependencies
	l.setupUsecaseDependencies()

	// setting up services
	err = l.setupServices()
	if err != nil {
		return nil, err
	}

	return &l, nil
}

func (l *Locator) appendStandardHandlers(serverName string) {
	l.restHandlers = map[string]http.Handler{
		"/":        servers.NewRestRootHandler(serverName),
		"/version": servers.NewRestVersionHandler(),
	}

	for path, handler := range servers.NewRestAPIDocsHandlers(
		serverName,
		"/docs/",
		"/docs/service.swagger.json",
		swagger.SwgJSON,
	) {
		l.restHandlers[path] = handler
	}
}

func (l *Locator) setLogger() {
	if l.LoggerProvider == nil {
		l.logger = zapctxd.New(zapctxd.Config{
			Level:   l.Config.Log.Level,
			DevMode: l.Config.IsDev(),
			FieldNames: ctxd.FieldNames{
				Timestamp: "timestamp",
				Message:   "message",
			},
			StripTime: l.Config.Log.LockTime,
			Output:    l.Config.Log.Output,
		})

		l.LoggerProvider = l.logger
	}
}

// makeDBx initializes database.
func makeDBx(cfg config.DBConfig) (*sqlx.DB, error) {
	db, err := makeDB(cfg)
	if err != nil {
		return nil, err
	}

	return sqlx.NewDb(db, cfg.DriverName), nil
}

// makeDB initializes database.
func makeDB(cfg config.DBConfig) (*sql.DB, error) {
	driverName, err := ocsql.Register(cfg.DriverName,
		ocsql.WithQuery(true),
		ocsql.WithRowsClose(true),
		ocsql.WithRowsAffected(true),
		ocsql.WithAllowRoot(true),
	)
	if err != nil {
		return nil, err
	}

	ocsql.RegisterAllViews()

	db, err := sql.Open(driverName, cfg.DSN)
	if err != nil {
		return nil, err
	}

	ocsql.RecordStats(db, time.Second)

	db.SetConnMaxLifetime(cfg.MaxLifetime)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	return db, nil
}

// makeStorage initializes database storage.
func makeStorage(
	db *sqlx.DB,
	logger ctxd.Logger,
) *sqluct.Storage {
	st := sqluct.NewStorage(db)

	st.Format = squirrel.Dollar
	st.OnError = func(ctx context.Context, err error) {
		logger.Error(ctx, "storage failure", "error", err)
	}

	return st
}

func (l *Locator) setGRPCUnitaryInterceptors() {
	l.grpcUnitaryInterceptors = append(l.grpcUnitaryInterceptors, []grpc.UnaryServerInterceptor{
		// recovering from panic
		grpcRecovery.UnaryServerInterceptor(),
		grpcLogging.UnaryServerInterceptor(grpcInterceptorLogger(l.logger)),
	}...)
}

func (l *Locator) setupUsecaseDependencies() {
}

func (l *Locator) setupServices() error {
	grpcOpts := append(
		l.opts.grpcOpts,
		servers.WithChainUnaryInterceptor(l.grpcUnitaryInterceptors...),
	)

	// If metrics service is enabled, add metrics observer.
	if l.opts.enableMetrics {
		grpcOpts = append(
			grpcOpts,
			servers.WithChainUnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
			servers.WithChainStreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		)

		grpc_prometheus.EnableHandlingTimeHistogram()
	}

	l.VioService = service.NewVioService(
		service.VioServiceConfig{
			Config: servers.Config{
				Name: "grpc " + l.Config.ServiceName,
			},
		},
		grpcOpts...,
	)

	var err error

	grpcRestOpts := append(
		l.opts.grpcRestOpts,
		servers.WithHandlers(l.restHandlers),
	)

	l.VioRESTService, err = service.NewVioRESTService(
		service.VioRESTServiceConfig{
			Config: servers.Config{
				Name: "grpc rest " + l.Config.ServiceName,
			},
			GRPCAddr: l.VioService.Addr(),
		},
		grpcRestOpts...,
	)
	if err != nil {
		return err
	}

	// Check if metrics service is enabled.
	if !l.opts.enableMetrics {
		return nil
	}

	metricsOpts := append(
		l.opts.metricsOpts,
		servers.WithGRPCServer(l.VioService.GRPC),
	)

	l.VioMetricsService = servers.NewMetrics(
		servers.Config{
			Name: "metrics " + l.Config.ServiceName,
		},
		metricsOpts...,
	)

	return nil
}

// grpcInterceptorLogger adapts zapctxd logger to interceptor logger.
func grpcInterceptorLogger(l *zapctxd.Logger) grpcLogging.Logger {
	return grpcLogging.LoggerFunc(func(ctx context.Context, lvl grpcLogging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		switch lvl {
		case grpcLogging.LevelDebug:
			l.Debug(ctx, msg)
		case grpcLogging.LevelInfo:
			l.Info(ctx, msg)
		case grpcLogging.LevelWarn:
			l.Warn(ctx, msg)
		case grpcLogging.LevelError:
			l.Error(ctx, msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
