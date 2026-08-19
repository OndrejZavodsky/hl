package graph

import (
	"hl/ui"
	"strings"
	"testing"
)

func TestGenerateDAG_And_Validation(t *testing.T) {
	tests := []struct {
		name        string
		pipeline    ui.Pipeline
		expectErr   bool
		errContains string
		checkGraph  func(t *testing.T, g *Graph)
	}{
		{
			name: "Valid linear pipeline",
			pipeline: ui.Pipeline{
				Steps: []ui.Step{
					{Name: "build", Tool: "go"},
					{Name: "test", Tool: "go", DependsOn: []string{"build"}},
					{Name: "deploy", Tool: "docker", DependsOn: []string{"test"}},
				},
			},
			expectErr: false,
			checkGraph: func(t *testing.T, g *Graph) {
				if len(g.Nodes) != 3 {
					t.Fatalf("expected 3 nodes, got %d", len(g.Nodes))
				}
				if g.Nodes["build"].InDegree != 0 {
					t.Errorf("expected build InDegree to be 0, got %d", g.Nodes["build"].InDegree)
				}
				if len(g.Nodes["build"].Children) != 1 || g.Nodes["build"].Children[0].Step.Name != "test" {
					t.Errorf("expected build child to be test")
				}
				if g.Nodes["test"].InDegree != 1 {
					t.Errorf("expected test InDegree to be 1, got %d", g.Nodes["test"].InDegree)
				}
				if len(g.Nodes["test"].Parents) != 1 || g.Nodes["test"].Parents[0].Step.Name != "build" {
					t.Errorf("expected test parent to be build")
				}
			},
		},
		{
			name: "Valid diamond pipeline (parallel paths)",
			pipeline: ui.Pipeline{
				Steps: []ui.Step{
					{Name: "setup", Tool: "bash"},
					{Name: "test-unit", Tool: "go", DependsOn: []string{"setup"}},
					{Name: "test-int", Tool: "go", DependsOn: []string{"setup"}},
					{Name: "report", Tool: "bash", DependsOn: []string{"test-unit", "test-int"}},
				},
			},
			expectErr: false,
			checkGraph: func(t *testing.T, g *Graph) {
				if g.Nodes["report"].InDegree != 2 {
					t.Errorf("expected report InDegree to be 2, got %d", g.Nodes["report"].InDegree)
				}
				if len(g.Nodes["setup"].Children) != 2 {
					t.Errorf("expected setup to have 2 children, got %d", len(g.Nodes["setup"].Children))
				}
			},
		},
		{
			name: "Error on duplicate step name",
			pipeline: ui.Pipeline{
				Steps: []ui.Step{
					{Name: "build", Tool: "go"},
					{Name: "build", Tool: "docker"},
				},
			},
			expectErr:   true,
			errContains: "duplicate step name detected",
		},
		{
			name: "Error on missing dependency",
			pipeline: ui.Pipeline{
				Steps: []ui.Step{
					{Name: "test", Tool: "go", DependsOn: []string{"non-existent-step"}},
				},
			},
			expectErr:   true,
			errContains: "depends on non-existent step",
		},
		{
			name: "Error on self-dependency",
			pipeline: ui.Pipeline{
				Steps: []ui.Step{
					{Name: "build", Tool: "go", DependsOn: []string{"build"}},
				},
			},
			expectErr:   true,
			errContains: "cannot depend on itself",
		},
		{
			name: "Error on direct circular dependency (A <-> B)",
			pipeline: ui.Pipeline{
				Steps: []ui.Step{
					{Name: "stepA", Tool: "bash", DependsOn: []string{"stepB"}},
					{Name: "stepB", Tool: "bash", DependsOn: []string{"stepA"}},
				},
			},
			expectErr:   true,
			errContains: "circular dependency detected",
		},
		{
			name: "Error on indirect circular dependency (A -> B -> C -> A)",
			pipeline: ui.Pipeline{
				Steps: []ui.Step{
					{Name: "stepA", Tool: "bash", DependsOn: []string{"stepC"}},
					{Name: "stepB", Tool: "bash", DependsOn: []string{"stepA"}},
					{Name: "stepC", Tool: "bash", DependsOn: []string{"stepB"}},
				},
			},
			expectErr:   true,
			errContains: "circular dependency detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := GenerateDAG(tt.pipeline)

			if err == nil {
				err = g.Validate()
			}

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errContains)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkGraph != nil {
				tt.checkGraph(t, g)
			}
		})
	}
}
