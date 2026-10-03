package detection

import (
	"slices"
	"strings"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
	"github.com/railwayapp/railpack/core/providers/deno"
	"github.com/railwayapp/railpack/core/providers/dotnet"
	"github.com/railwayapp/railpack/core/providers/golang"
	"github.com/railwayapp/railpack/core/providers/java"
	"github.com/railwayapp/railpack/core/providers/node"
	"github.com/railwayapp/railpack/core/providers/python"
	"github.com/railwayapp/railpack/core/providers/ruby"
)

type Input struct {
	Type         string   `json:"type"`
	Label        string   `json:"label"`
	Env          string   `json:"env"`
	Description  string   `json:"description"`
	DefaultValue any      `json:"defaultValue,omitempty"`
	Placeholder  string   `json:"placeholder,omitempty"`
	Options      []string `json:"options,omitempty"`
}

type inputDefinition struct {
	providers   string
	name        string
	label       string
	description string
	boolean     bool
}

// Only configuration read by the builder belongs here; runtime application secrets do not.
var inputDefinitions = []inputDefinition{
	{"", "CONFIG_FILE", "Config file", "Path to the Railpack JSON configuration file.", false},
	{"", "INSTALL_CMD", "Install command", "Override the dependency installation command.", false},
	{"", "BUILD_CMD", "Build command", "Override the build command.", false},
	{"", "START_CMD", "Start command", "Override the container start command.", false},
	{"", "PACKAGES", "Additional packages", "Space-separated Mise packages, optionally with @version.", false},
	{"", "BUILD_APT_PACKAGES", "Build apt packages", "Space-separated apt packages needed during build.", false},
	{"", "DEPLOY_APT_PACKAGES", "Runtime apt packages", "Space-separated apt packages to include in the final image.", false},
	{"", "DISABLE_CACHES", "Disabled caches", "Space-separated cache names, or * to disable all caches.", false},
	{"node php ruby elixir", "NODE_VERSION", "Node version", "Node version or version constraint; automatic resolution may also use Mise.", false},
	{"node php ruby elixir", "BUN_VERSION", "Bun version", "Bun version when Bun is used by the project.", false},
	{"node php ruby elixir", "NODE_NPM_INSTALL", "npm install command", "Custom install command when the selected package manager is npm.", false},
	{"node php ruby elixir", "NODE_PRUNE_CMD", "Dependency prune command", "Override the command used to prune Node dependencies.", false},
	{"node php ruby elixir", "NODE_INSTALL_PATTERNS", "Install file patterns", "Space-separated additional file patterns copied into the install step.", false},
	{"node", "NODE_PLAYWRIGHT_INSTALL", "Install Playwright browsers", "Install Playwright browsers and their required system packages.", true},
	{"node php ruby elixir", "PRUNE_DEPS", "Prune dependencies", "Remove development dependencies from the deployed application.", true},
	{"node", "NO_SPA", "Disable static deployment", "Disable automatic SPA deployment with Caddy.", true},
	{"node", "SPA_OUTPUT_DIR", "Static output directory", "Directory containing generated static files; setting it forces SPA deployment unless disabled.", false},
	{"node", "NX_APP", "Nx application", "Name or package name of the Nx Next.js application to deploy.", false},
	{"node", "ANGULAR_PROJECT", "Angular project", "Angular project to use when reading its output directory.", false},
	{"python", "PYTHON_VERSION", "Python version", "Python version or version constraint.", false},
	{"python", "PYTHON_PLAYWRIGHT_INSTALL", "Install Playwright browsers", "Install Playwright browsers and their required system packages.", true},
	{"python", "DJANGO_APP_NAME", "Django application", "Django module containing the WSGI application.", false},
	{"golang", "GO_VERSION", "Go version", "Go version or version constraint.", false},
	{"golang", "GO_BIN", "Go binary", "Name of the Go binary to build.", false},
	{"golang", "GO_WORKSPACE_MODULE", "Go workspace module", "Module path to build in a Go workspace.", false},
	{"rust", "RUST_VERSION", "Rust version", "Rust version or version constraint; project toolchain files can also influence selection.", false},
	{"rust", "RUST_BIN", "Rust binary", "Cargo binary target to build.", false},
	{"rust", "CARGO_WORKSPACE", "Cargo workspace member", "Cargo workspace member to build.", false},
	{"ruby", "RUBY_VERSION", "Ruby version", "Ruby version or version constraint.", false},
	{"php", "PHP_EXTENSIONS", "PHP extensions", "Space-separated PHP extensions added to those requested by Composer.", false},
	{"php", "PHP_ROOT_DIR", "PHP document root", "Document root served by the PHP container.", false},
	{"java", "JDK_VERSION", "JDK version", "Java version or version constraint.", false},
	{"java", "GRADLE_VERSION", "Gradle version", "Gradle version or version constraint.", false},
	{"dotnet", "DOTNET_VERSION", ".NET version", ".NET SDK version or version constraint.", false},
	{"deno", "DENO_VERSION", "Deno version", "Deno version or version constraint.", false},
	{"elixir", "ELIXIR_VERSION", "Elixir version", "Elixir version or version constraint.", false},
	{"elixir", "ERLANG_VERSION", "Erlang version", "Erlang version or version constraint.", false},
	{"gleam", "GLEAM_INCLUDE_SOURCE", "Include Gleam source", "Include source files in the deployed application.", true},
	{"staticfile", "STATIC_FILE_ROOT", "Static root directory", "Directory containing the files to serve.", false},
	{"shell", "SHELL_SCRIPT", "Shell script", "Path to the startup script, normally start.sh.", false},
}

func inputsFor(provider string, ctx *generate.GenerateContext, defaults map[string]string) []Input {
	inputs := []Input{}
	for _, definition := range inputDefinitions {
		if definition.providers != "" && !slices.Contains(strings.Fields(definition.providers), provider) {
			continue
		}
		input := Input{
			Type: "text", Label: definition.label, Env: "DEPLOPACK_" + definition.name,
			Description: definition.description, Placeholder: "Automatic",
		}
		defaultValue := defaults[definition.name]
		if definition.boolean {
			input.Type = "select"
			input.Options = []string{"true", "false"}
			if defaultValue == "" {
				defaultValue = "false"
			}
		}
		if value, _ := ctx.Env.GetConfigVariable(definition.name); value != "" && (!strings.HasSuffix(definition.name, "_VERSION") || defaultValue == "") {
			defaultValue = value
			if definition.boolean {
				defaultValue = "false"
				if ctx.Env.IsConfigVariableTruthy(definition.name) {
					defaultValue = "true"
				}
			}
		}
		switch definition.name {
		case "INSTALL_CMD", "BUILD_CMD", "START_CMD", "NODE_NPM_INSTALL", "NODE_PRUNE_CMD":
			input.Type = "text-list"
			if commands := commandDefaults(definition.name, defaultValue, ctx); commands != nil {
				input.DefaultValue = commands
			}
		default:
			if defaultValue != "" {
				input.DefaultValue = defaultValue
			}
		}
		if input.DefaultValue != nil {
			input.Placeholder = ""
		} else {
			switch definition.name {
			case "PACKAGES":
				input.Placeholder = "ffmpeg@latest"
			case "BUILD_APT_PACKAGES", "DEPLOY_APT_PACKAGES":
				input.Placeholder = "curl libpq-dev"
			case "DISABLE_CACHES":
				input.Placeholder = "cache-name or *"
			case "NODE_INSTALL_PATTERNS":
				input.Placeholder = "patches/** scripts/**"
			}
		}
		inputs = append(inputs, input)
	}
	return inputs
}

func commandDefaults(name, fallback string, ctx *generate.GenerateContext) []string {
	if value, _ := ctx.Env.GetConfigVariable(name); value != "" {
		return []string{value}
	}
	if name == "INSTALL_CMD" || name == "BUILD_CMD" {
		stepName := strings.ToLower(strings.TrimSuffix(name, "_CMD"))
		if step := ctx.Config.Steps[stepName]; step != nil && step.Commands != nil {
			commands := make([]string, 0, len(step.Commands))
			for _, command := range step.Commands {
				exec, ok := command.(plan.ExecCommand)
				if !ok || command.IsSpread() {
					ctx.Logger.LogInfo("%s step includes structured commands; its command input remains automatic", stepName)
					return nil
				}
				value := exec.Cmd
				// String config commands retain their original shell text as the custom name.
				if exec.CustomName != "" && exec.Cmd == plan.ShellCommandString(exec.CustomName) {
					value = exec.CustomName
				}
				commands = append(commands, value)
			}
			return commands
		}
	}
	if fallback != "" {
		return []string{fallback}
	}
	return nil
}

func localDefaults(ctx *generate.GenerateContext, rootDir string, versions map[string]*generate.MisePackageInfo) map[string]string {
	defaults := map[string]string{"STATIC_FILE_ROOT": rootDir}
	for name, key := range map[string]string{
		"NODE_VERSION": "nodeVersionConstraint", "BUN_VERSION": "bunVersionConstraint",
		"PYTHON_VERSION": "pythonVersionConstraint", "RUBY_VERSION": "rubyVersionConstraint",
		"GO_VERSION": "goVersionConstraint", "DOTNET_VERSION": "dotnetVersionConstraint",
		"RUST_VERSION": "rustVersionConstraint", "DJANGO_APP_NAME": "djangoAppName",
		"ELIXIR_VERSION": "elixirVersionConstraint", "ERLANG_VERSION": "erlangVersionConstraint",
		"SHELL_SCRIPT": "shellScript",
		"PHP_ROOT_DIR": "phpRootDirectory", "SPA_OUTPUT_DIR": "outputDirectory",
		"BUILD_CMD": "buildCommand", "START_CMD": "startCommand",
		"INSTALL_CMD": "installCommand", "NODE_NPM_INSTALL": "npmInstallCommand",
	} {
		defaults[name] = ctx.Metadata.Get(key)
	}
	for name, fallback := range map[string]string{
		"NODE_VERSION": node.DEFAULT_NODE_VERSION, "BUN_VERSION": node.DEFAULT_BUN_VERSION,
		"PYTHON_VERSION": python.DEFAULT_PYTHON_VERSION, "RUBY_VERSION": ruby.DEFAULT_RUBY_VERSION,
		"GO_VERSION": golang.DEFAULT_GO_VERSION, "DENO_VERSION": deno.DEFAULT_DENO_VERSION,
		"DOTNET_VERSION": dotnet.DEFAULT_DOTNET_VERSION,
	} {
		if defaults[name] == "" {
			defaults[name] = fallback
		}
	}
	// Apply declarations in the same order as each provider's package setup.
	versionTools := map[string]string{
		"NODE_VERSION": "node", "BUN_VERSION": "bun", "DOTNET_VERSION": "dotnet",
		"PYTHON_VERSION": "python", "RUBY_VERSION": "ruby", "GO_VERSION": "go",
		"DENO_VERSION": "deno", "RUST_VERSION": "rust",
		"ELIXIR_VERSION": "elixir", "ERLANG_VERSION": "erlang",
	}
	for name, tool := range versionTools {
		if value, _ := ctx.Env.GetConfigVariable(name); value != "" {
			defaults[name] = value
		}
		if pkg := versions[tool]; pkg != nil {
			defaults[name] = pkg.Version
		}
		// These providers apply environment overrides after reading Mise declarations.
		if name == "PYTHON_VERSION" || name == "RUBY_VERSION" || name == "DENO_VERSION" {
			if value, _ := ctx.Env.GetConfigVariable(name); value != "" {
				defaults[name] = value
			}
		}
		// Node and .NET inspection use the builder's complete declaration precedence.
		if tool == "node" || tool == "bun" || tool == "dotnet" {
			if pkg := ctx.Resolver.Get(tool); pkg != nil {
				defaults[name] = pkg.Version
			}
		}
	}
	// Java does not apply Mise declarations when choosing its JDK.
	if ctx.Metadata.Get("javaPackageManager") == "maven" {
		defaults["JDK_VERSION"] = java.DEFAULT_JDK_VERSION
	}
	for _, name := range []string{"JDK_VERSION", "GRADLE_VERSION"} {
		if value, _ := ctx.Env.GetConfigVariable(name); value != "" {
			defaults[name] = value
		}
	}
	// The builder applies configured packages after provider and Mise declarations.
	versionTools["JDK_VERSION"] = "java"
	versionTools["GRADLE_VERSION"] = "gradle"
	for name, tool := range versionTools {
		if version, ok := ctx.Config.Packages[tool]; ok {
			version = strings.TrimSpace(version)
			if version == "" {
				version = "latest"
			}
			defaults[name] = version
			key := tool + "VersionConstraint"
			if ctx.Metadata.Get(key) != "" {
				ctx.Metadata.Set(key, version)
			}
		}
	}
	if ctx.App.HasFile("railpack.json") {
		defaults["CONFIG_FILE"] = "railpack.json"
	}
	if ctx.Config.Deploy.StartCmd != "" {
		defaults["START_CMD"] = ctx.Config.Deploy.StartCmd
	}
	defaults["BUILD_APT_PACKAGES"] = strings.Join(ctx.Config.BuildAptPackages, " ")
	defaults["DEPLOY_APT_PACKAGES"] = strings.Join(ctx.Config.Deploy.AptPackages, " ")
	packages := make([]string, 0, len(ctx.Config.Packages))
	for name, version := range ctx.Config.Packages {
		if version != "" {
			name += "@" + version
		}
		packages = append(packages, name)
	}
	slices.Sort(packages)
	defaults["PACKAGES"] = strings.Join(packages, " ")
	return defaults
}
