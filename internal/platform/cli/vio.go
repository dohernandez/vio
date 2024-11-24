package cli

import (
	"github.com/bool64/ctxd"
	"github.com/dohernandez/vio/internal/domain/usecase"
	"github.com/dohernandez/vio/internal/platform/app"
	"github.com/dohernandez/vio/internal/platform/config"
	readplatform "github.com/dohernandez/vio/internal/platform/reader"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap/zapcore"
)

var parseFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "parallel",
		Usage:       "Number of parallel processes to parse the geolocation data.",
		Required:    false,
		DefaultText: "1",
		Value:       "1",
		EnvVars:     []string{"PARALLEL"},
		Aliases:     []string{"p"},
	},
	&cli.BoolFlag{
		Name:        "verbose",
		Required:    false,
		Usage:       "enable verbose output",
		DefaultText: "false",
		Aliases:     []string{"v"},
	},
}

var parseFilesystemFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "file",
		Usage:       "File to read the geolocation data.",
		Required:    true,
		DefaultText: "transactions.csv",
		Value:       "transactions.csv",
		EnvVars:     []string{"FILE", "DATA_FILE"},
		Aliases:     []string{"f"},
	},
}

// NewCliApp creates a new cli app.
func NewCliApp() *cli.App {
	return &cli.App{
		Name:  "vio",
		Usage: "Vio is command line tool for geolocation.",
		Commands: []*cli.Command{
			{
				Name:  "parse",
				Usage: "Load, parse and store geolocation data.",
				Subcommands: []*cli.Command{
					{
						Name:  "filesystem",
						Usage: "Parse geolocation data from a file from a filesystem.",
						Flags: append(parseFlags, parseFilesystemFlags...),
						Action: func(c *cli.Context) error {
							cfg, err := config.GetConfig()
							if err != nil {
								return ctxd.WrapError(c.Context, err, "failed to load configurations")
							}

							// set log level
							if c.Bool("verbose") {
								cfg.Log.Level = zapcore.DebugLevel
								// set output to command line writer
								cfg.Log.Output = c.App.Writer
							}

							// initialize locator
							deps, err := app.NewServiceLocator(cfg, app.WithNoService())
							if err != nil {
								return ctxd.WrapError(c.Context, err, "failed to initialize service locator")
							}

							// initialize reader
							reader := readplatform.NewFileSystem(c.String("file"), deps.CtxdLogger())

							// parse data
							parser := usecase.NewParseGeolocationData(deps.GeoStorage(), deps.CtxdLogger())

							return parser.Process(c.Context, reader, c.Uint("parallel"))
						},
					},
				},
			},
		},
	}
}
