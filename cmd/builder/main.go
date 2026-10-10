package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/railwayapp/railpack/buildkit"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/mise"
	"github.com/railwayapp/railpack/internal/deplopack"
)

var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(version)
		return
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if mise.IsTemporary(err) {
			os.Exit(75)
		}
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 1 {
		return fmt.Errorf("builder accepts configuration through environment variables only")
	}
	provider := strings.TrimSpace(os.Getenv("DEPLOPACK_PROVIDER"))
	if provider == "" {
		return fmt.Errorf("DEPLOPACK_PROVIDER is required")
	}

	source, err := app.NewApp(".")
	if err != nil {
		return err
	}
	env := deplopack.FromEnviron(os.Environ())
	// Call the upstream planner and BuildKit directly without registering CLI commands.
	result, err := core.GenerateBuildPlan(source, env, &core.GenerateBuildPlanOptions{
		RailpackVersion: version,
		Provider:        provider,
	})
	if err != nil {
		return err
	}
	core.PrettyPrintBuildResult(result, core.PrintOptions{Version: version, Name: "DeploPack"})
	if !result.Success {
		return fmt.Errorf("build planning failed")
	}
	if err := buildkit.ValidateSecrets(result.Plan, env); err != nil {
		return err
	}
	output := &progressWriter{output: os.Stdout}
	buildErr := buildkit.BuildWithBuildkitClient(source.Source, result.Plan, buildkit.BuildWithBuildkitClientOptions{
		ProgressMode:   "plain",
		ProgressWriter: output,
		SecretsHash:    buildkit.GetSecretsHash(env),
		Secrets:        env.Variables,
		GitHubToken:    os.Getenv("GITHUB_TOKEN"),
	})
	flushErr := output.Flush()
	if buildErr != nil {
		return buildErr
	}
	return flushErr
}
