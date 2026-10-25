package httpadapter_test

import (
	"os"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter"
	"github.com/TimoSto/tax-return-organizer/backend/internal/adapters/inbound/httpadapter/httpx"
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

// TestOpenAPI_ProblemCodes fails if the Problem.code enum in api/openapi.yaml differs from the codes httpx can
// answer with (see httpx/problemcodes.go).
func TestOpenAPI_ProblemCodes(t *testing.T) {
	raw, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}

	var spec struct {
		Components struct {
			Schemas struct {
				Problem struct {
					Properties struct {
						Code struct {
							Enum []string `yaml:"enum"`
						} `yaml:"code"`
					} `yaml:"properties"`
				} `yaml:"Problem"`
			} `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		t.Fatal(err)
	}

	inSpec := slices.Clone(spec.Components.Schemas.Problem.Properties.Code.Enum)
	slices.Sort(inSpec)

	inCode := httpx.ProblemCodes()

	for _, c := range inCode {
		if !slices.Contains(inSpec, c) {
			t.Errorf("code %q is not in the Problem.code enum of openapi.yaml; run the sync-openapi skill", c)
		}
	}

	for _, c := range inSpec {
		if !slices.Contains(inCode, c) {
			t.Errorf("openapi.yaml lists code %q but no error produces it", c)
		}
	}
}
