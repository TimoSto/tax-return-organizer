package httpadapter_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter"
)

const specPath = "../../../../api/openapi.yaml"

// TestOpenAPI_Routes fails if an endpoint exists only in the code or only in api/openapi.yaml.
// It compares method and path, not schemas. Update the spec with the sync-openapi skill.
func TestOpenAPI_Routes(t *testing.T) {
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}

	var spec struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		t.Fatal(err)
	}

	var inSpec []string

	for path, ops := range spec.Paths {
		for method := range ops {
			if slices.Contains([]string{"get", "put", "post", "delete", "patch", "head", "options"}, method) {
				inSpec = append(inSpec, strings.ToUpper(method)+" "+path)
			}
		}
	}

	var inCode []string
	for _, r := range httpadapter.Routes(httpadapter.Deps{}) {
		inCode = append(inCode, r.Pattern)
	}

	for _, r := range inCode {
		if !slices.Contains(inSpec, r) {
			t.Errorf("route %q is not in openapi.yaml; run the sync-openapi skill", r)
		}
	}

	for _, r := range inSpec {
		if !slices.Contains(inCode, r) {
			t.Errorf("openapi.yaml documents %q but no handler serves it; run the sync-openapi skill", r)
		}
	}
}
