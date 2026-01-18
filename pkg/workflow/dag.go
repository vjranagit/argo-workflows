package workflow

import "fmt"

// DAGBuilder provides a fluent API for constructing DAG templates.
// Unlike Hera's >> operator for dependencies, this uses explicit methods.
type DAGBuilder struct {
	name  string
	tasks []DAGTask
}

// NewDAG creates a new DAG builder.
func NewDAG(name string) *DAGBuilder {
	return &DAGBuilder{
		name:  name,
		tasks: make([]DAGTask, 0),
	}
}

// AddTask adds a task to the DAG.
func (d *DAGBuilder) AddTask(task DAGTask) *DAGBuilder {
	d.tasks = append(d.tasks, task)
	return d
}

// Task creates a DAG task with the given name and template.
// This is more explicit than Hera's approach where tasks are created
// by calling template functions directly.
func (d *DAGBuilder) Task(name, template string, options ...TaskOption) *DAGBuilder {
	task := DAGTask{
		Name:     name,
		Template: template,
	}

	for _, opt := range options {
		opt(&task)
	}

	d.tasks = append(d.tasks, task)
	return d
}

// Build creates a Template with the DAG configuration.
func (d *DAGBuilder) Build() Template {
	return Template{
		Name: d.name,
		DAG: &DAG{
			Tasks: d.tasks,
		},
	}
}

// TaskOption is a functional option for configuring DAG tasks.
type TaskOption func(*DAGTask)

// WithDependencies sets task dependencies.
// Unlike Hera's A >> [B, C] syntax, this is explicit.
func WithDependencies(deps ...string) TaskOption {
	return func(t *DAGTask) {
		t.Dependencies = deps
	}
}

// WithArguments sets task arguments.
func WithArguments(args *Arguments) TaskOption {
	return func(t *DAGTask) {
		t.Arguments = args
	}
}

// WithCondition sets a when condition for the task.
func WithCondition(condition string) TaskOption {
	return func(t *DAGTask) {
		t.When = condition
	}
}

// DependencyGraph helps visualize and validate DAG dependencies.
// This is a helper that Hera doesn't provide - useful for debugging.
type DependencyGraph struct {
	tasks map[string]*DAGTask
}

// NewDependencyGraph creates a new dependency graph from DAG tasks.
func NewDependencyGraph(tasks []DAGTask) *DependencyGraph {
	graph := &DependencyGraph{
		tasks: make(map[string]*DAGTask),
	}

	for i := range tasks {
		graph.tasks[tasks[i].Name] = &tasks[i]
	}

	return graph
}

// Validate checks for cycles and missing dependencies.
func (g *DependencyGraph) Validate() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for name := range g.tasks {
		if err := g.hasCycle(name, visited, recStack); err != nil {
			return err
		}
	}

	// Check for missing dependencies
	for name, task := range g.tasks {
		for _, dep := range task.Dependencies {
			if _, ok := g.tasks[dep]; !ok {
				return fmt.Errorf("task %q depends on non-existent task %q", name, dep)
			}
		}
	}

	return nil
}

// hasCycle performs DFS to detect cycles.
func (g *DependencyGraph) hasCycle(task string, visited, recStack map[string]bool) error {
	visited[task] = true
	recStack[task] = true

	if t, ok := g.tasks[task]; ok {
		for _, dep := range t.Dependencies {
			if !visited[dep] {
				if err := g.hasCycle(dep, visited, recStack); err != nil {
					return err
				}
			} else if recStack[dep] {
				return fmt.Errorf("cycle detected involving tasks %q and %q", task, dep)
			}
		}
	}

	recStack[task] = false
	return nil
}

// TopologicalSort returns tasks in execution order.
// This can help with visualization and understanding workflow execution.
func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}

	visited := make(map[string]bool)
	stack := make([]string, 0)

	for name := range g.tasks {
		if !visited[name] {
			g.topologicalSortUtil(name, visited, &stack)
		}
	}

	// Stack is already in correct topological order (reversed during DFS)
	// No need to reverse again
	return stack, nil
}

func (g *DependencyGraph) topologicalSortUtil(task string, visited map[string]bool, stack *[]string) {
	visited[task] = true

	if t, ok := g.tasks[task]; ok {
		for _, dep := range t.Dependencies {
			if !visited[dep] {
				g.topologicalSortUtil(dep, visited, stack)
			}
		}
	}

	*stack = append(*stack, task)
}
