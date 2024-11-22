package config

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap/zapcore"
)

// Config represents config with variables needed for an app.
type Config struct {
	ServiceName    string   `envconfig:"SERVICE_NAME"`
	AppGRPCPort    int      `envconfig:"APP_GRPC_PORT" default:"8000"`
	AppRESTPort    int      `envconfig:"APP_REST_PORT" default:"8080"`
	AppMetricsPort int      `envconfig:"APP_METRICS_PORT" default:"8080"`
	Environment    string   `envconfig:"ENVIRONMENT" default:"dev"`
	PostgresDB     DBConfig `split_words:"true"`
	Log            LoggerConfig
}

// DBConfig represents the DB configuration fields and values.
type DBConfig struct {
	DSN          string        `envconfig:"DATABASE_DSN" required:"true"`
	MaxLifetime  time.Duration `envconfig:"MAX_LIFETIME" default:"4h"`
	MaxIdleConns int           `envconfig:"MAX_IDLE_CONNECTIONS" default:"20"`
	MaxOpenConns int           `envconfig:"MAX_OPEN_CONNECTIONS" default:"20"`
	DriverName   string
}

// LoggerConfig is log configuration.
type LoggerConfig struct {
	Level      zapcore.Level `envconfig:"LOG_LEVEL" default:"error"`
	FieldNames string        `envconfig:"LOG_FILENAMES" default:"true"`
	Output     io.Writer
	// LockTime disables time variance in logger.
	LockTime bool

	// CallerSkip configures how deeply func calls should be skipped, default 1.
	CallerSkip int
}

// GetConfig returns service config, filled from environment variables.
func GetConfig() (*Config, error) {
	var c Config

	if err := envconfig.Process("", &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// IsDev returns true for development environment.
func (c *Config) IsDev() bool {
	return len(c.Environment) > 2 && strings.ToLower(c.Environment[0:3]) == "dev"
}

// IsTest returns true for testing environment.
func (c *Config) IsTest() bool {
	return len(c.Environment) > 2 && strings.ToLower(c.Environment[0:4]) == "test"
}

// WithEnvFiles populates env vars from provided files.
//
// It returns an error if file does not exist.
func WithEnvFiles(files ...string) error {
	var found []string

	for _, f := range files {
		if fileExists(f) {
			found = append(found, f)
		}
	}

	if len(found) == 0 {
		return nil
	}

	return godotenv.Load(files...)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}
