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
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers"
	"github.com/railwayapp/railpack/core/providers/staticfile"
)

type Detection struct {
	Type    string `json:"type"`
	Path    string `json:"path,omitempty"`
	RootDir string `json:"rootDir,omitempty"`
}

type Result struct {
	Success    bool         `json:"success"`
	Detections []Detection  `json:"detections"`
	Logs       []logger.Msg `json:"logs,omitempty"`
}

// Reuses upstream detection without initializing providers or resolving tool versions.
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
	// Detect currently uses only file access, environment, configuration and logging.
	ctx := &generate.GenerateContext{App: source, Env: env, Config: config, Logger: logs}
	for _, provider := range providers.GetLanguageProviders() {
		matched, err := provider.Detect(ctx)
		if err != nil {
			logs.LogError("Failed to detect provider %s: %s", provider.Name(), err)
			continue
		}
		if matched {
			detection := Detection{Type: provider.Name()}
			if staticProvider, ok := provider.(*staticfile.StaticfileProvider); ok {
				rootDir, err := staticProvider.RootDir(ctx)
				if err != nil {
					logs.LogError("Failed to resolve staticfile root: %s", err)
					continue
				}
				detection.RootDir = rootDir
			}
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
