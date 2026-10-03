package deplopack

import (
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
		case "DEPLOPACK_PROVIDER", "BUILDKIT_HOST", "DOCKER_HOST", "DOCKER_CONFIG", "PATH", "HOME":
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
