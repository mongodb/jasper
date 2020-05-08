package buildsystem

type Target struct {
	Name                string
	Tags                []string
	RequiredEnvironment map[string]string
	Commands            []Command
	ExecutionOptions    ExecutionOptions
	Dependencies        []Dependency
	Output              Output
}

type Command struct {
	Script         string
	Args           []string
	CommandOptions CommandOptions
}

type ExecutionOptions struct {
	ContinueOnError bool
}

type Dependency struct {
	Target []string
	Files  []string
}

type CommandOptions struct {
	Args          []string
	RunBackground bool
}

type Output struct {
	Name   string
	Path   string
	Format string
}
