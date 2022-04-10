package model

type Action struct {
	Using      string            `yaml:"using"`
	Env        map[string]string `yaml:"env"`
	Main       string            `yaml:"main"`
	Image      string            `yaml:"image"`
	Entrypoint string            `yaml:"entrypoint"`
	Args       []string          `yaml:"args"`
	Steps      []Step            `yaml:"steps"`
}
