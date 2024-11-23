package config_test

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/dohernandez/vio/internal/platform/config"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

var cfg = &config.Config{
	ServiceName:    "test-server",
	AppGRPCPort:    0,
	AppRESTPort:    0,
	AppMetricsPort: 0,
	Environment:    "test",
	PostgresDB: config.DBConfig{
		DSN:          "postgres://test:test@database:5432/test?sslmode=disable",
		MaxLifetime:  4 * time.Hour,
		MaxIdleConns: 20,
		MaxOpenConns: 20,
	},
	Log: config.LoggerConfig{
		Level:      zapcore.DebugLevel,
		FieldNames: "true",
	},
}

func TestGetConfig_EnvSuccessfully(t *testing.T) {
	t.Parallel()

	require.NoError(t, os.Setenv("SERVICE_NAME", "test-server"))
	require.NoError(t, os.Setenv("APP_GRPC_PORT", "0"))
	require.NoError(t, os.Setenv("APP_REST_PORT", "0"))
	require.NoError(t, os.Setenv("APP_METRICS_PORT", "0"))
	require.NoError(t, os.Setenv("ENVIRONMENT", "test"))
	require.NoError(t, os.Setenv("DATABASE_DSN", "postgres://test:test@database:5432/test?sslmode=disable"))
	require.NoError(t, os.Setenv("LOG_LEVEL", "DEBUG"))

	got, err := config.GetConfig()
	require.NoError(t, err, "GetConfig() error = %v", err)

	if !reflect.DeepEqual(got, cfg) {
		t.Errorf("GetConfig() got = %v, want %v", got, cfg)
	}
}

func TestGetConfig_FileSuccessfully(t *testing.T) {
	t.Parallel()

	require.NoError(t, config.WithEnvFiles("./__testdata/.env.template"))

	got, err := config.GetConfig()
	require.NoError(t, err, "GetConfig() error = %v", err)

	if !reflect.DeepEqual(got, cfg) {
		t.Errorf("GetConfig() got = %v, want %v", got, cfg)
	}
}
