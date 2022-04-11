package model

type Action struct {
	Name        string            `yaml:"name"`
	Author      string            `yaml:"author"`
	Description string            `yaml:"description"`
	Inputs      map[string]Input  `yaml:"inputs"`
	Outputs     map[string]Output `yaml:"outputs"`
	Runs        ActionRuns        `yaml:"runs"`
	//Branding    struct {
	//	Color string `yaml:"color"`
	//	Icon  string `yaml:"icon"`
	//} `yaml:"branding"`
}

type ActionRuns struct {
	Using      string            `yaml:"using"`
	Env        map[string]string `yaml:"env"`
	Main       string            `yaml:"main"`
	Image      string            `yaml:"image"`
	Entrypoint string            `yaml:"entrypoint"`
	Args       []string          `yaml:"args"`
	Steps      []Step            `yaml:"steps"`
}

// Input parameters allow you to specify data that the action expects to use during runtime. GitHub stores input parameters as environment variables. Input ids with uppercase letters are converted to lowercase during runtime. We recommended using lowercase input ids.
type Input struct {
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
	Default     string `yaml:"default"`
}

// Output parameters allow you to declare data that an action sets. Actions that run later in a workflow can use the output data set in previously run actions. For example, if you had an action that performed the addition of two inputs (x + y = z), the action could output the sum (z) for other actions to use as an input.
type Output struct {
	Description string `yaml:"description"`
	Value       string `yaml:"value"`
}
