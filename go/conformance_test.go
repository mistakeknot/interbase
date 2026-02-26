//go:build conformance

package interbase

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mistakeknot/interbase/toolerror"

	"gopkg.in/yaml.v3"
)

type conformanceSuite struct {
	Domain    string            `yaml:"domain"`
	Languages []string          `yaml:"languages,omitempty"`
	Tests     []conformanceTest `yaml:"tests"`
}

type conformanceTest struct {
	Name           string            `yaml:"name"`
	Setup          map[string]string `yaml:"setup,omitempty"`
	Call           string            `yaml:"call"`
	Args           []string          `yaml:"args,omitempty"`
	Expect         any               `yaml:"expect,omitempty"`
	ExpectError    *bool             `yaml:"expect_error,omitempty"`
	ExpectContains string            `yaml:"expect_contains,omitempty"`
	ExpectNoError  *bool             `yaml:"expect_no_error,omitempty"`
	ExpectJSON     map[string]any    `yaml:"expect_json,omitempty"`
	Data           map[string]any    `yaml:"data,omitempty"`
}

func TestConformance(t *testing.T) {
	confDir := filepath.Join("..", "tests", "conformance")
	files, err := filepath.Glob(filepath.Join(confDir, "*.yaml"))
	if err != nil {
		t.Fatalf("glob conformance dir: %v", err)
	}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		var suite conformanceSuite
		if err := yaml.Unmarshal(data, &suite); err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}

		// Skip if language-restricted and Go is not listed
		if len(suite.Languages) > 0 {
			found := false
			for _, l := range suite.Languages {
				if l == "go" {
					found = true
					break
				}
			}
			if !found {
				t.Logf("SKIP %s (not for go)", f)
				continue
			}
		}

		for _, tc := range suite.Tests {
			t.Run(tc.Name, func(t *testing.T) {
				// Apply setup
				for k, v := range tc.Setup {
					t.Setenv(k, v)
				}
				runConformanceCall(t, tc)
			})
		}
	}
}

func runConformanceCall(t *testing.T, tc conformanceTest) {
	t.Helper()

	switch tc.Call {
	case "has_ic":
		got := HasIC()
		assertBool(t, got, tc.Expect)
	case "has_bd":
		got := HasBD()
		assertBool(t, got, tc.Expect)
	case "has_companion":
		arg := ""
		if len(tc.Args) > 0 {
			arg = tc.Args[0]
		}
		got := HasCompanion(arg)
		assertBool(t, got, tc.Expect)
	case "get_bead":
		got := GetBead()
		assertString(t, got, tc.Expect)
	case "in_ecosystem":
		got := InEcosystem()
		assertBool(t, got, tc.Expect)
	case "in_sprint":
		got := InSprint()
		assertBool(t, got, tc.Expect)
	case "phase_set":
		// PhaseSet returns nothing — just verify it doesn't panic
		PhaseSet("bead-123", "planned")
	case "emit_event":
		// EmitEvent returns nothing — just verify it doesn't panic
		EmitEvent("run-123", "test-event")
	case "session_status":
		got := SessionStatus()
		if tc.ExpectContains != "" && !strings.Contains(got, tc.ExpectContains) {
			t.Errorf("got %q, want contains %q", got, tc.ExpectContains)
		}
	case "plugin_cache_path":
		arg := ""
		if len(tc.Args) > 0 {
			arg = tc.Args[0]
		}
		got := PluginCachePath(arg)
		assertString(t, got, tc.Expect)
	case "ecosystem_root":
		got := EcosystemRoot()
		if tc.Expect != nil {
			assertString(t, got, tc.Expect)
		}

	// --- toolerror calls ---
	case "toolerror_new":
		if len(tc.Args) < 2 {
			t.Fatal("toolerror_new needs 2 args")
		}
		te := toolerror.New(tc.Args[0], "%s", tc.Args[1])
		assertToolErrorJSON(t, te, tc.ExpectJSON)
	case "toolerror_new_with_data":
		if len(tc.Args) < 2 {
			t.Fatal("toolerror_new_with_data needs 2 args")
		}
		data := make(map[string]any)
		for k, v := range tc.Data {
			data[k] = v
		}
		te := toolerror.New(tc.Args[0], "%s", tc.Args[1]).WithData(data)
		assertToolErrorJSON(t, te, tc.ExpectJSON)
	case "toolerror_str":
		if len(tc.Args) < 2 {
			t.Fatal("toolerror_str needs 2 args")
		}
		te := toolerror.New(tc.Args[0], "%s", tc.Args[1])
		got := te.Error()
		assertString(t, got, tc.Expect)
	case "toolerror_wrap_tool_error":
		if len(tc.Args) < 2 {
			t.Fatal("toolerror_wrap_tool_error needs 2 args")
		}
		original := toolerror.New(tc.Args[0], "%s", tc.Args[1])
		wrapped := toolerror.Wrap(original)
		assertToolErrorJSON(t, wrapped, tc.ExpectJSON)
	case "toolerror_wrap_generic":
		if len(tc.Args) < 1 {
			t.Fatal("toolerror_wrap_generic needs 1 arg")
		}
		wrapped := toolerror.Wrap(java_error(tc.Args[0]))
		assertToolErrorJSON(t, wrapped, tc.ExpectJSON)
	default:
		t.Skipf("unknown function: %s", tc.Call)
	}
}

// java_error is a simple error type for testing Wrap with generic errors.
type java_error string

func (e java_error) Error() string { return string(e) }

func assertBool(t *testing.T, got bool, expect any) {
	t.Helper()
	if expect == nil {
		return
	}
	var want bool
	switch v := expect.(type) {
	case bool:
		want = v
	case string:
		want = v == "true"
	}
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func assertString(t *testing.T, got string, expect any) {
	t.Helper()
	if expect == nil {
		return
	}
	want := ""
	switch v := expect.(type) {
	case string:
		want = v
	default:
		b, _ := json.Marshal(v)
		want = string(b)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertToolErrorJSON(t *testing.T, te *toolerror.ToolError, expectJSON map[string]any) {
	t.Helper()
	if expectJSON == nil {
		return
	}

	// Parse the JSON output
	jsonStr := te.JSON()
	var got map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &got); err != nil {
		t.Fatalf("failed to parse toolerror JSON %q: %v", jsonStr, err)
	}

	for k, v := range expectJSON {
		if k == "expect_message_contains" {
			msg, _ := got["message"].(string)
			vs, _ := v.(string)
			if !strings.Contains(msg, vs) {
				t.Errorf("message %q does not contain %q", msg, vs)
			}
			continue
		}
		gotVal := got[k]
		if !jsonEqual(gotVal, v) {
			t.Errorf("key %q: got %v (%T), want %v (%T)", k, gotVal, gotVal, v, v)
		}
	}
}

// jsonEqual compares two values from JSON unmarshaling (handles float64 vs int, nested maps).
func jsonEqual(a, b any) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return string(aj) == string(bj)
}
