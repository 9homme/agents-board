// Package workflow_test validates the .github/workflows/e2e.yml GitHub Actions
// workflow file against the US003 acceptance criteria (D-003, D-004).
package workflow_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	// services/agent-board/internal/workflow → 4 levels up = repo root
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
	return filepath.Clean(filepath.Join(repoRoot, ".github", "workflows", "e2e.yml"))
}

// workflowDoc is a minimal representation of the e2e.yml structure for assertions.
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
	Name string                 `yaml:"name"`
	Uses string                 `yaml:"uses"`
	Run  string                 `yaml:"run"`
	If   string                 `yaml:"if"`
	With map[string]interface{} `yaml:"with"`
}

// parseWorkflow reads and unmarshals the workflow YAML; fails the test on any error.
func parseWorkflow(t *testing.T) workflowDoc {
	t.Helper()
	data, err := os.ReadFile(workflowPath(t))
	require.NoError(t, err, "workflow file must exist at %s", workflowPath(t))

	var doc workflowDoc
	require.NoError(t, yaml.Unmarshal(data, &doc), "workflow YAML must be valid")
	return doc
}

// UT-US003-001: workflow file exists and triggers only on pull_request to main (D-003).
func TestWorkflow_TriggerPullRequestToMain(t *testing.T) {
	doc := parseWorkflow(t)

	branches := doc.On.PullRequest.Branches
	require.Len(t, branches, 1, "pull_request trigger must specify exactly one branch")
	assert.Equal(t, "main", branches[0], "pull_request trigger branch must be 'main'")
}

// UT-US003-002: workflow has no push trigger (D-003: PR-to-main only).
func TestWorkflow_NoPushTrigger(t *testing.T) {
	data, err := os.ReadFile(workflowPath(t))
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

// UT-US003-003: workflow uses make e2e-up, make e2e-seed, make e2e-run (D-004).
func TestWorkflow_MakeTargetsPresent(t *testing.T) {
	doc := parseWorkflow(t)

	var runCommands []string
	for _, job := range doc.Jobs {
		for _, step := range job.Steps {
			if step.Run != "" {
				runCommands = append(runCommands, step.Run)
			}
		}
	}

	runsContain := func(target string) bool {
		for _, r := range runCommands {
			if strings.Contains(r, target) {
				return true
			}
		}
		return false
	}

	assert.True(t, runsContain("make e2e-up"), "workflow must call 'make e2e-up'")
	assert.True(t, runsContain("make e2e-seed"), "workflow must call 'make e2e-seed'")
	assert.True(t, runsContain("make e2e-run"), "workflow must call 'make e2e-run'")
}

// UT-US003-004: artifact upload and teardown steps both use if: always() (D-004).
func TestWorkflow_AlwaysConditionsPresent(t *testing.T) {
	doc := parseWorkflow(t)

	var artifactAlways, teardownAlways bool
	for _, job := range doc.Jobs {
		for _, step := range job.Steps {
			if step.If != "always()" {
				continue
			}
			if strings.Contains(step.Uses, "upload-artifact") {
				artifactAlways = true
			}
			if strings.Contains(step.Run, "make e2e-down") {
				teardownAlways = true
			}
		}
	}

	assert.True(t, artifactAlways, "an upload-artifact step must have 'if: always()'")
	assert.True(t, teardownAlways, "the 'make e2e-down' step must have 'if: always()'")
}

// UT-US003-005: runner is ubuntu-latest.
func TestWorkflow_RunsOnUbuntuLatest(t *testing.T) {
	doc := parseWorkflow(t)

	require.NotEmpty(t, doc.Jobs, "workflow must have at least one job")
	for name, job := range doc.Jobs {
		assert.Equal(t, "ubuntu-latest", job.RunsOn,
			"job %q must run on ubuntu-latest", name)
	}
}
