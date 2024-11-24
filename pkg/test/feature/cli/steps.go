package cli

import (
	"bytes"
	"strings"

	"github.com/cucumber/godog"
	"github.com/urfave/cli/v2"
)

type (
	// Command manages the run of cobra command.
	Command struct {
		app    func() *cli.App
		Output bytes.Buffer
		Err    error
	}

	// App manages the execution of cli.App command.
	App struct {
		Commands map[string]*Command
	}
)

// RegisterContext adds command context to test suite.
func RegisterContext(s *godog.ScenarioContext, cliApp *App) {
	s.Step(`^I run the command "([^"]*)" with the arguments "([^"]*)"$`, cliApp.run)

	s.Step(`^the command "([^"]*)" finishes successfully$`, cliApp.shouldNotFailed)
	s.Step(`^the command "([^"]*)" failed$`, cliApp.shouldFailed)

	s.Step(`the command "([^"]*)" should output "([^"]*)"$`, cliApp.outputShouldEqual)
}

// Add adds a command to the app.
func (c *App) Add(command string, app func() *cli.App) {
	if c.Commands == nil {
		c.Commands = make(map[string]*Command)
	}

	c.Commands[command] = &Command{
		app: app,
	}
}

func (c *App) run(command, args string) error {
	cmd := c.Commands[command]

	cmd.Output = bytes.Buffer{}
	cmd.Err = nil

	app := cmd.app()
	app.Writer = &cmd.Output

	ctx := cli.NewContext(app, nil, nil)

	cliArgs := []string{command}
	cliArgs = append(cliArgs, strings.Split(args, " ")...)

	err := app.Command(command).Run(ctx, cliArgs...)

	cmd.Err = err

	return nil
}

func (c *App) shouldNotFailed(command string) error {
	cmd := c.Commands[command]

	if cmd.Err != nil {
		return cmd.Err
	}

	return nil
}

func (c *App) shouldFailed(command string) error {
	cmd := c.Commands[command]

	if cmd.Err == nil {
		return cmd.Err
	}

	return nil
}

func (c *App) outputShouldEqual(command, output string) error {
	cmd := c.Commands[command]

	if cmd.Output.String() != output {
		return cmd.Err
	}

	return nil
}
