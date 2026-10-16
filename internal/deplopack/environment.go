package deplopack

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/railwayapp/railpack/core/app"
)

// Converts Deplopack configuration after collecting upstream values so aliases always win.
func FromEnviron(variables []string) *app.Environment {
	env := app.NewEnvironment(nil)
	overrides := make(map[string]string)
	for _, variable := range variables {
		name, value, _ := strings.Cut(variable, "=")
		switch name {
		case "DEPLOPACK_OCI_IMAGES", "DEPLOPACK_CACHE_IMPORTS", "DEPLOPACK_PROVIDER", "DEPLOPACK_ENV_NAMES", "BUILDKIT_HOST", "DOCKER_HOST", "DOCKER_CONFIG", "PATH", "HOME":
			continue
		}
		if suffix, ok := strings.CutPrefix(name, "DEPLOPACK_"); ok {
			overrides["RAILPACK_"+suffix] = value
			continue
		}
		env.SetVariable(name, value)
	}
	for name, value := range overrides {
		env.SetVariable(name, value)
	}
	return env
}

// Only explicitly forwarded variables participate in builds from newer agents.
func BuilderEnvironment(variables []string) (*app.Environment, error) {
	values := make(map[string]string)
	for _, variable := range variables {
		name, value, _ := strings.Cut(variable, "=")
		values[name] = value
	}
	names, provided := values["DEPLOPACK_ENV_NAMES"]
	if !provided {
		return FromEnviron(variables), nil
	}
	var allowed []string
	if err := json.Unmarshal([]byte(names), &allowed); err != nil {
		return nil, fmt.Errorf("invalid build environment names: %w", err)
	}
	forwarded := make([]string, 0, len(allowed))
	for _, name := range allowed {
		if name == "" || strings.ContainsAny(name, "=\x00") {
			return nil, fmt.Errorf("invalid build environment name")
		}
		value, ok := values[name]
		if !ok {
			return nil, fmt.Errorf("missing build environment variable: %s", name)
		}
		forwarded = append(forwarded, name+"="+value)
	}
	return FromEnviron(forwarded), nil
}
