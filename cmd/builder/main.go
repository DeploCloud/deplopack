package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/railwayapp/railpack/buildkit"
	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/config"
	"github.com/railwayapp/railpack/core/mise"
	"github.com/railwayapp/railpack/core/plan"
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
	core.PrettyPrintBuildResult(result, core.PrintOptions{Version: version})
	if !result.Success {
		return fmt.Errorf("build planning failed")
	}
	serialized, err := json.MarshalIndent(struct {
		Schema string `json:"$schema"`
		*plan.BuildPlan
	}{Schema: config.SchemaUrl, BuildPlan: result.Plan}, "", "  ")
	if err != nil {
		return err
	}
	core.PrettyPrintSectionHeader(os.Stdout, "Generated railpack-plan.json")
	core.PrettyPrintJSON(os.Stdout, serialized)
	if err := buildkit.ValidateSecrets(result.Plan, env); err != nil {
		return err
	}
	return buildkit.BuildWithBuildkitClient(source.Source, result.Plan, buildkit.BuildWithBuildkitClientOptions{
		ProgressMode: "plain",
		SecretsHash:  buildkit.GetSecretsHash(env),
		Secrets:      env.Variables,
		GitHubToken:  os.Getenv("GITHUB_TOKEN"),
	})
}
