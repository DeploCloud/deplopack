package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/railwayapp/railpack/cli"
	urfave "github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if exit, ok := err.(urfave.ExitCoder); ok {
			os.Exit(exit.ExitCode())
		}
		os.Exit(cli.ExitCodeFailure)
	}
}

func run() error {
	if len(os.Args) != 1 {
		return fmt.Errorf("builder accepts configuration through environment variables only")
	}
	if strings.TrimSpace(os.Getenv("DEPLOPACK_PROVIDER")) == "" {
		return fmt.Errorf("DEPLOPACK_PROVIDER is required")
	}

	cli.Version = version
	// Reuse upstream planning, secret validation, BuildKit execution and exit codes.
	command := *cli.BuildCommand
	command.Name = "deplopack-builder"
	command.DisableSliceFlagSeparator = true
	command.Flags = append(command.Flags, &urfave.StringFlag{Name: "provider"})
	args := []string{command.Name, "--progress", "plain"}
	for _, option := range []struct{ env, flag string }{
		{"DEPLOPACK_PROVIDER", "provider"},
		{"DEPLOPACK_BUILD_COMMAND", "build-cmd"},
		{"DEPLOPACK_START_COMMAND", "start-cmd"},
		{"DEPLOPACK_CONFIG_FILE", "config-file"},
		{"DEPLOPACK_IMAGE_NAME", "name"},
		{"DEPLOPACK_PLATFORM", "platform"},
		{"DEPLOPACK_OUTPUT_DIR", "output"},
	} {
		if value := os.Getenv(option.env); value != "" {
			args = append(args, "--"+option.flag, value)
		}
	}
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if strings.HasPrefix(name, "DEPLOPACK_") || name == "BUILDKIT_HOST" || name == "DOCKER_HOST" || name == "DOCKER_CONFIG" {
			continue
		}
		// Passing names preserves empty values without copying secrets into arguments.
		args = append(args, "--env", name)
	}
	if value, ok := os.LookupEnv("DEPLOPACK_SHOW_PLAN"); ok {
		args = append(args, "--show-plan="+value)
	}
	return command.Run(context.Background(), append(args, "."))
}
