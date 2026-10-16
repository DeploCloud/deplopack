package buildkit

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

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
	encoded, _ := json.Marshal(env.Variables)
	return fmt.Sprintf("%x", sha256.Sum256(encoded))
}
