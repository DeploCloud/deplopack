package detection

import (
	"cmp"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/railwayapp/railpack/core"
	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/logger"
	"github.com/railwayapp/railpack/core/mise"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers"
	"github.com/railwayapp/railpack/core/providers/procfile"
	"github.com/railwayapp/railpack/core/providers/staticfile"
	"github.com/railwayapp/railpack/core/resolver"
)

type Detection struct {
	Type     string            `json:"type"`
	Path     string            `json:"path,omitempty"`
	RootDir  string            `json:"rootDir,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Inputs   []Input           `json:"inputs,omitempty"`
}

type Result struct {
	Success    bool         `json:"success"`
	Detections []Detection  `json:"detections"`
	Logs       []logger.Msg `json:"logs,omitempty"`
}

// Reads tool declarations with Mise without resolving releases or generating a build plan.
func Analyze(source *app.App, env *app.Environment, options *core.GenerateBuildPlanOptions) (*Result, error) {
	result := &Result{Detections: []Detection{}}
	logs := logger.NewLogger()
	defer func() { result.Logs = logs.Logs }()
	if env == nil {
		env = app.NewEnvironment(nil)
	}
	if options == nil {
		options = &core.GenerateBuildPlanOptions{}
	}
	config, err := core.GetConfig(source, env, options, logs)
	if err != nil {
		return result, err
	}
	ignore, err := plan.NewDockerignoreContext(source)
	if err != nil {
		return result, fmt.Errorf("parse .dockerignore: %w", err)
	}
	if err := source.SetExcludePatterns(slices.Concat(ignore.Excludes, config.Exclude)); err != nil {
		return result, err
	}
	// Detection itself only needs repository files; Mise is initialized for matched candidates.
	ctx := &generate.GenerateContext{App: source, Env: env, Config: config, Logger: logs}
	for _, provider := range providers.GetLanguageProviders() {
		ctx.Config = config
		matched, err := provider.Detect(ctx)
		if err != nil {
			logs.LogError("Failed to detect provider %s: %s", provider.Name(), err)
			continue
		}
		if matched {
			detection := Detection{Type: provider.Name()}
			selectedOptions := *options
			selectedOptions.Provider = provider.Name()
			// Initial config loading already reported file diagnostics.
			ctx.Config, err = core.GetConfig(source, env, &selectedOptions, logger.NewLogger())
			if err != nil {
				return result, err
			}
			ctx.Resolver, err = resolver.NewResolver(mise.InstallDir)
			if err != nil {
				return result, err
			}
			ctx.Deploy = generate.NewDeployBuilder()
			if ctx.MiseStepBuilder != nil {
				ctx.MiseStepBuilder.Resolver = ctx.Resolver
			}
			versions, err := ctx.GetMiseStepBuilder().GetMisePackageVersions(ctx)
			if err != nil {
				return result, fmt.Errorf("read Mise declarations: %w", err)
			}
			ctx.GetMiseStepBuilder().MisePackages = nil
			// Each candidate has independent metadata and package requests.
			ctx.Metadata = generate.NewMetadata()
			if inspector, ok := provider.(interface {
				Inspect(*generate.GenerateContext) error
			}); ok {
				if err := inspector.Inspect(ctx); err != nil {
					logs.LogError("Failed to inspect provider %s: %s", provider.Name(), err)
					continue
				}
			}
			if staticProvider, ok := provider.(*staticfile.StaticfileProvider); ok {
				rootDir, err := staticProvider.RootDir(ctx)
				if err != nil {
					logs.LogError("Failed to resolve staticfile root: %s", err)
					continue
				}
				detection.RootDir = rootDir
			}
			procfileProvider := &procfile.ProcfileProvider{}
			if _, err := procfileProvider.Plan(ctx); err != nil {
				return result, fmt.Errorf("read Procfile: %w", err)
			}
			if ctx.Deploy.StartCmd != "" {
				ctx.Metadata.Set("startCommand", ctx.Deploy.StartCmd)
			}
			if ctx.Config.Deploy.StartCmd != "" {
				ctx.Metadata.Set("startCommand", ctx.Config.Deploy.StartCmd)
			}
			detection.Metadata = ctx.Metadata.Properties
			detection.Inputs = inputsFor(strings.ToLower(provider.Name()), ctx, localDefaults(ctx, detection.RootDir, versions))
			result.Detections = append(result.Detections, detection)
		}
	}
	// File detections stay outside the upstream language registry.
	if source.HasFile("Dockerfile") {
		result.Detections = append(result.Detections, Detection{Type: "dockerfile", Path: "Dockerfile"})
	}
	// Compose files may be excluded from image contexts but still define deployments.
	composeFiles, err := source.FindAllFiles("**/{compose,docker-compose}{,.*}.{yaml,yml}")
	if err != nil {
		return result, fmt.Errorf("find Compose files: %w", err)
	}
	standardNames := []string{"compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"}
	priority := func(path string) int {
		index := slices.Index(standardNames, filepath.Base(path))
		if index < 0 {
			return len(standardNames)
		}
		return index
	}
	slices.SortFunc(composeFiles, func(a, b string) int {
		if order := cmp.Compare(priority(a), priority(b)); order != 0 {
			return order
		}
		if order := cmp.Compare(strings.Count(a, "/"), strings.Count(b, "/")); order != 0 {
			return order
		}
		return strings.Compare(a, b)
	})
	for _, file := range composeFiles {
		result.Detections = append(result.Detections, Detection{Type: "compose", Path: file})
	}
	result.Success = true
	for _, msg := range logs.Logs {
		if msg.Level == logger.Error {
			result.Success = false
		}
	}
	return result, nil
}
