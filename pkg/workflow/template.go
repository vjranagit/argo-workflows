package workflow

// TemplateBuilder provides helper functions for creating common template types.
// Unlike Hera's decorators, these are explicit constructor functions.

// TemplateOption is a functional option for template-level configuration.
type TemplateOption func(*Template)

// ContainerTemplate creates a container template.
func ContainerTemplate(name string, opts ...interface{}) Template {
	tmpl := Template{
		Name:      name,
		Container: &Container{},
	}

	for _, opt := range opts {
		switch o := opt.(type) {
		case ContainerOption:
			o(tmpl.Container)
		case TemplateOption:
			o(&tmpl)
		}
	}

	return tmpl
}

// ContainerOption is a functional option for container configuration.
type ContainerOption func(*Container)

// WithImage sets the container image.
func WithImage(image string) ContainerOption {
	return func(c *Container) {
		c.Image = image
	}
}

// WithCommand sets the container command.
func WithCommand(command ...string) ContainerOption {
	return func(c *Container) {
		c.Command = command
	}
}

// WithArgs sets the container args.
func WithArgs(args ...string) ContainerOption {
	return func(c *Container) {
		c.Args = args
	}
}

// WithEnv adds environment variables.
func WithEnv(env ...EnvVar) ContainerOption {
	return func(c *Container) {
		c.Env = append(c.Env, env...)
	}
}

// WithResources sets resource requirements.
func WithResources(resources *Resources) ContainerOption {
	return func(c *Container) {
		c.Resources = resources
	}
}

// ScriptTemplate creates a script template.
// Different from Hera's @script decorator - this is explicit.
func ScriptTemplate(name string, opts ...interface{}) Template {
	tmpl := Template{
		Name:   name,
		Script: &Script{},
	}

	for _, opt := range opts {
		switch o := opt.(type) {
		case ScriptOption:
			o(tmpl.Script)
		case TemplateOption:
			o(&tmpl)
		}
	}

	return tmpl
}

// ScriptOption is a functional option for script configuration.
type ScriptOption func(*Script)

// WithScriptImage sets the script container image.
func WithScriptImage(image string) ScriptOption {
	return func(s *Script) {
		s.Image = image
	}
}

// WithScriptCommand sets the script command.
func WithScriptCommand(command ...string) ScriptOption {
	return func(s *Script) {
		s.Command = command
	}
}

// WithSource sets the script source code.
// Unlike Hera which extracts Python function source via AST,
// we require explicit source code as a string.
func WithSource(source string) ScriptOption {
	return func(s *Script) {
		s.Source = source
	}
}

// WithScriptEnv adds environment variables to script.
func WithScriptEnv(env ...EnvVar) ScriptOption {
	return func(s *Script) {
		s.Env = append(s.Env, env...)
	}
}

// WithScriptResources sets script resource requirements.
func WithScriptResources(resources *Resources) ScriptOption {
	return func(s *Script) {
		s.Resources = resources
	}
}

// WithInputs adds inputs to a template.
func WithInputs(inputs *Inputs) TemplateOption {
	return func(t *Template) {
		t.Inputs = inputs
	}
}

// WithOutputs adds outputs to a template.
func WithOutputs(outputs *Outputs) TemplateOption {
	return func(t *Template) {
		t.Outputs = outputs
	}
}

// NewInputs creates a new Inputs object.
func NewInputs() *Inputs {
	return &Inputs{
		Parameters: make([]Parameter, 0),
		Artifacts:  make([]Artifact, 0),
	}
}

// AddParameter adds a parameter to inputs.
func (i *Inputs) AddParameter(param Parameter) *Inputs {
	i.Parameters = append(i.Parameters, param)
	return i
}

// AddArtifact adds an artifact to inputs.
func (i *Inputs) AddArtifact(artifact Artifact) *Inputs {
	i.Artifacts = append(i.Artifacts, artifact)
	return i
}

// NewOutputs creates a new Outputs object.
func NewOutputs() *Outputs {
	return &Outputs{
		Parameters: make([]Parameter, 0),
		Artifacts:  make([]Artifact, 0),
	}
}

// AddParameter adds a parameter to outputs.
func (o *Outputs) AddParameter(param Parameter) *Outputs {
	o.Parameters = append(o.Parameters, param)
	return o
}

// AddArtifact adds an artifact to outputs.
func (o *Outputs) AddArtifact(artifact Artifact) *Outputs {
	o.Artifacts = append(o.Artifacts, artifact)
	return o
}

// WithResult sets the result output.
func (o *Outputs) WithResult(result string) *Outputs {
	o.Result = result
	return o
}

// NewArguments creates a new Arguments object.
func NewArguments() *Arguments {
	return &Arguments{
		Parameters: make([]Parameter, 0),
		Artifacts:  make([]Artifact, 0),
	}
}

// AddParameter adds a parameter to arguments.
func (a *Arguments) AddParameter(param Parameter) *Arguments {
	a.Parameters = append(a.Parameters, param)
	return a
}

// AddArtifact adds an artifact to arguments.
func (a *Arguments) AddArtifact(artifact Artifact) *Arguments {
	a.Artifacts = append(a.Artifacts, artifact)
	return a
}
