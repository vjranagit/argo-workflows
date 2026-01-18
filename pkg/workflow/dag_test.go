package workflow

import (
	"testing"
)

func TestDAGBuilder(t *testing.T) {
	dag := NewDAG("test-dag").
		Task("A", "template-a").
		Task("B", "template-b", WithDependencies("A")).
		Task("C", "template-c", WithDependencies("A")).
		Task("D", "template-d", WithDependencies("B", "C")).
		Build()

	if dag.Name != "test-dag" {
		t.Errorf("Expected name 'test-dag', got '%s'", dag.Name)
	}

	if dag.DAG == nil {
		t.Fatal("DAG should not be nil")
	}

	if len(dag.DAG.Tasks) != 4 {
		t.Errorf("Expected 4 tasks, got %d", len(dag.DAG.Tasks))
	}
}

func TestDependencyGraphValidate(t *testing.T) {
	tasks := []DAGTask{
		{Name: "A", Template: "t1"},
		{Name: "B", Template: "t2", Dependencies: []string{"A"}},
		{Name: "C", Template: "t3", Dependencies: []string{"A"}},
		{Name: "D", Template: "t4", Dependencies: []string{"B", "C"}},
	}

	graph := NewDependencyGraph(tasks)
	if err := graph.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestDependencyGraphCycle(t *testing.T) {
	tasks := []DAGTask{
		{Name: "A", Template: "t1", Dependencies: []string{"B"}},
		{Name: "B", Template: "t2", Dependencies: []string{"A"}},
	}

	graph := NewDependencyGraph(tasks)
	if err := graph.Validate(); err == nil {
		t.Error("Expected cycle detection error")
	}
}

func TestDependencyGraphMissingDep(t *testing.T) {
	tasks := []DAGTask{
		{Name: "A", Template: "t1"},
		{Name: "B", Template: "t2", Dependencies: []string{"nonexistent"}},
	}

	graph := NewDependencyGraph(tasks)
	if err := graph.Validate(); err == nil {
		t.Error("Expected missing dependency error")
	}
}

func TestTopologicalSort(t *testing.T) {
	tasks := []DAGTask{
		{Name: "A", Template: "t1"},
		{Name: "B", Template: "t2", Dependencies: []string{"A"}},
		{Name: "C", Template: "t3", Dependencies: []string{"A"}},
		{Name: "D", Template: "t4", Dependencies: []string{"B", "C"}},
	}

	graph := NewDependencyGraph(tasks)
	order, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	if len(order) != 4 {
		t.Errorf("Expected 4 tasks in order, got %d", len(order))
	}

	// A should be first or among first tasks (no dependencies)
	// D should be last (depends on B and C)
	firstIdx := -1
	lastIdx := -1
	for i, name := range order {
		if name == "A" {
			firstIdx = i
		}
		if name == "D" {
			lastIdx = i
		}
	}

	if firstIdx == -1 || lastIdx == -1 {
		t.Error("Expected both A and D in order")
	}

	// A should come before D
	if firstIdx > lastIdx {
		t.Errorf("A should come before D in topological order")
	}
}
