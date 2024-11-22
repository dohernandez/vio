package feature

import (
	"bytes"
	"flag"
	"fmt"
	"math/rand"
	"testing"

	"github.com/bool64/godogx"
	"github.com/bool64/godogx/allure"
	"github.com/cucumber/godog"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

// Used by init().
//
//nolint:gochecknoglobals
var (
	runWithTags string
	runFeature  string
	runAllure   bool
	out         = bytes.Buffer{}
	opt         = godog.Options{
		Tags:   runWithTags,
		Strict: true,
		Output: &out,
	}
)

// This has to run on init to define -feature flag.
//
//nolint:gochecknoinits
func init() {
	flagSet := pflag.CommandLine

	flagSet.StringVar(&runFeature, "feature", "",
		"Optional feature to run. Filename without the extension .feature")

	flag.BoolVar(&runAllure, "allure", false,
		"Enable allure formatter")

	godog.BindCommandLineFlags("", &opt)
}

// RunFeatures run feature tests.
func RunFeatures(t *testing.T, path string, featureContext func(t *testing.T, s *godog.ScenarioContext)) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	flag.Parse()

	if runFeature != "" {
		path = fmt.Sprintf("%s/%s.feature", path, runFeature)
	}

	godogx.RegisterPrettyFailedFormatter()

	if opt.Randomize == 0 {
		opt.Randomize = rand.Int63() // nolint: gosec
	}

	if opt.Format == "" {
		opt.Format = "pretty-failed"
	}

	opt.Paths = []string{path}
	opt.TestingT = t
	opt.StopOnFailure = true

	suite := godog.TestSuite{
		Name: "Integration test",
		ScenarioInitializer: func(s *godog.ScenarioContext) {
			featureContext(t, s)
		},
		Options: &opt,
	}

	if runAllure {
		allure.RegisterFormatter()

		suite.Options.Format += ",allure"
	}

	assert.Equal(t, 0, suite.Run(), "non-zero status returned, failed to run feature tests")
}
