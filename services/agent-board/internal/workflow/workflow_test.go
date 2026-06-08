package workflow_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// workflowPath returns the absolute path to .github/workflows/e2e.yml,
// resolved relative to this test file's location (services/agent-board/internal/workflow/).
func workflowPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	// services/agent-board/internal/workflow → up 4 dirs = repo root
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
	return filepath.Join(repoRoot, ".github", "workflows", "e2e.yml")
}

// workflowDoc is a minimal representation of the e2e.yml structure.
type workflowDoc struct {
	On   onBlock            `yaml:"on"`
	Jobs map[string]jobSpec `yaml:"jobs"`
}

type onBlock struct {
	PullRequest prBlock `yaml:"pull_request"`
}

type prBlock struct {
	Branches []string `yaml:"branches"`
}

type jobSpec struct {
	RunsOn string     `yaml:"runs-on"`
	Steps  []stepSpec `yaml:"steps"`
}

type stepSpec struct {
	Name  string `yaml:"name"`
	Uses  string `yaml:"uses"`
	Run   string `yaml:"run"`
	If    string `yaml:"if"`
	With  map[string]interface{} `yaml:"with"`
}

func parseWorkflow(t *testing.T) workflowDoc {
	t.Helper()
	path := workflowPath(t)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "workflow file must exist at %s", path)

	var doc workflowDoc
	require.NoError(t, yaml.Unmarshal(data, &doc), "workflow YAML must be valid")
	return doc
}

// UT-US003-001: workflow file exists and triggers only on pull_request to main.
func TestWorkflow_TriggerPullRequestToMain(t *testing.T) {
	doc := parseWorkflow(t)

	branches := doc.On.PullRequest.Branches
	require.Len(t, branches, 1, "pull_request trigger must specify exactly one branch")
	assert.Equal(t, "main", branches[0], "pull_request trigger branch must be 'main'")
}

// UT-US003-002: workflow has no push trigger.
func TestWorkflow_NoPushTrigger(t *testing.T) {
	path := workflowPath(t)
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	// Parse as a generic map to detect the presence of 'push' under 'on'.
	var raw map[string]interface{}
	require.NoError(t, yaml.Unmarshal(data, &raw))

	onSection, ok := raw["on"]
	require.True(t, ok, "'on' key must exist in workflow")

	onMap, ok := onSection.(map[string]interface{})
	require.True(t, ok, "'on' must be a mapping")

	_, hasPush := onMap["push"]
	assert.False(t, hasPush, "workflow must NOT have a 'push' trigger")
}

// UT-US003-003: workflow uses make e2e-up, make e2e-seed, make e2e-run steps.
func TestWorkflow_MakeTargetsPresent(t *testing.T) {
	doc := parseWorkflow(t)

	var runs []string
	for _, job := range doc.Jobs {
		for _, step := range job.Steps {
			if step.Run != "" {
				runs = append(runs, step.Run)
			}
		}
	}

	found := func(target string) bool {
		for _, r := range runs {
			if contains(r, target) {
				return true
			}
		}
		return false
	}

	assert.True(t, found("make e2e-up"), "workflow must call 'make e2e-up'")
	assert.True(t, found("make e2e-seed"), "workflow must call 'make e2e-seed'")
	assert.True(t, found("make e2e-run"), "workflow must call 'make e2e-run'")
}

// UT-US003-004: artifact upload and teardown steps both use if: always().
func TestWorkflow_AlwaysConditionsPresent(t *testing.T) {
	doc := parseWorkflow(t)

	var artifactAlways, teardownAlways bool
	for _, job := range doc.Jobs {
		for _, step := range job.Steps {
			if step.If == "always()" {
				if step.Uses != "" && contains(step.Uses, "upload-artifact") {
					artifactAlways = true
				}
				if step.Run != "" && contains(step.Run, "make e2e-down") {
					teardownAlways = true
				}
			}
		}
	}

	assert.True(t, artifactAlways, "an upload-artifact step must have 'if: always()'")
	assert.True(t, teardownAlways, "the 'make e2e-down' step must have 'if: always()'")
}

// UT-US003-005: runner is ubuntu-latest.
func TestWorkflow_RunsOnUbuntuLatest(t *testing.T) {
	doc := parseWorkflow(t)

	for name, job := range doc.Jobs {
		assert.Equal(t, "ubuntu-latest", job.RunsOn,
			"job %q must run on ubuntu-latest", name)
	}
	require.NotEmpty(t, doc.Jobs, "workflow must have at least one job")
}

// contains is a helper for substring checks.
func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
