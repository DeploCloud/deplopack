package buildkit

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/railwayapp/railpack/core/app"
	"github.com/railwayapp/railpack/core/plan"
)

// Checks that every secret referenced by the plan is available before building.
func ValidateSecrets(plan *plan.BuildPlan, env *app.Environment) error {
	for _, secret := range plan.Secrets {
		if _, ok := env.Variables[secret]; !ok {
			return fmt.Errorf("missing environment variable: %s", secret)
		}
	}
	return nil
}

// Preserves upstream cache invalidation when build secret values change.
func GetSecretsHash(env *app.Environment) string {
	var secretsValue strings.Builder
	for _, v := range env.Variables {
		secretsValue.WriteString(v)
	}
	hasher := sha256.New()
	hasher.Write([]byte(secretsValue.String()))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}
