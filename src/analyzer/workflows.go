package analyzer

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
)

// Workflow is the structure of the files in .github/workflows
type Workflow struct {
	File  string
	Name  string            `yaml:"name"`
	RawOn yaml.Node         `yaml:"on"`
	Env   map[string]string `yaml:"env"`
	Jobs  map[string]*Job   `yaml:"jobs"`
	//Defaults Defaults          `yaml:"defaults"`
}

// CompositeRestrictions is the structure to control what is allowed in composite actions
//type CompositeRestrictions struct {
//	AllowCompositeUses            bool
//	AllowCompositeIf              bool
//	AllowCompositeContinueOnError bool
//}

// Job is the structure of one job in a workflow
type Job struct {
	Name string `yaml:"name"`
	//RawNeeds       yaml.Node                 `yaml:"needs"`
	RawRunsOn yaml.Node `yaml:"runs-on"`
	Env       yaml.Node `yaml:"env"`
	//If             yaml.Node                 `yaml:"if"`
	Steps []*Step `yaml:"steps"`
	//TimeoutMinutes int64                     `yaml:"timeout-minutes"`
	Services     map[string]*ContainerSpec `yaml:"services"`
	Strategy     *Strategy                 `yaml:"strategy"`
	RawContainer yaml.Node                 `yaml:"container"`
	//Defaults     Defaults                  `yaml:"defaults"`
	//Outputs map[string]string `yaml:"outputs"`
	//Result  string
}

// Strategy for the job
type Strategy struct {
	//FailFast          bool
	//MaxParallel       int
	//FailFastString    string    `yaml:"fail-fast"`
	//MaxParallelString string    `yaml:"max-parallel"`
	RawMatrix yaml.Node `yaml:"matrix"`
}

// Default settings that will apply to all steps in the job or workflow
//type Defaults struct {
//	Run RunDefaults `yaml:"run"`
//}

// Defaults for all run steps in the job or workflow
//type RunDefaults struct {
//	Shell            string `yaml:"shell"`
//	WorkingDirectory string `yaml:"working-directory"`
//}

// ContainerSpec is the specification of the container to use for the job
type ContainerSpec struct {
	Image       string            `yaml:"image"`
	Env         map[string]string `yaml:"env"`
	Ports       []string          `yaml:"ports"`
	Volumes     []string          `yaml:"volumes"`
	Options     string            `yaml:"options"`
	Credentials map[string]string `yaml:"credentials"`
	//Entrypoint  string
	//Args        string
	//Name        string
	//Reuse       bool
}

// Step is the structure of one step in a job
type Step struct {
	//ID               string            `yaml:"id"`
	//If               yaml.Node         `yaml:"if"`
	//Name             string            `yaml:"name"`
	Uses string `yaml:"uses"`
	Run  string `yaml:"run"`
	//WorkingDirectory string            `yaml:"working-directory"`
	//Shell            string            `yaml:"shell"`
	Env  yaml.Node         `yaml:"env"`
	With map[string]string `yaml:"with"`
	//ContinueOnError  bool              `yaml:"continue-on-error"`
	//TimeoutMinutes   int64             `yaml:"timeout-minutes"`
}

// RunsOn list for Job. Note that RunsOn will interpolate matrix automatically
func (j *Job) RunsOn() []string {
	switch j.RawRunsOn.Kind {
	case yaml.ScalarNode:
		var val string
		err := j.RawRunsOn.Decode(&val)
		if err != nil {
			log.Fatal(err)
		}
		return []string{val}

		//if !strings.Contains(val, "${{") || !strings.Contains(val, "}}") {
		//} else {
		//	// TODO interpolate matrix
		//}
	case yaml.SequenceNode:
		var val []string
		err := j.RawRunsOn.Decode(&val)
		if err != nil {
			log.Fatal(err)
		}
		return val
	}
	return nil
}

// Matrix decodes RawMatrix YAML node
func (j *Job) Matrix() map[string][]interface{} {
	if j.Strategy == nil {
		return nil
	}

	if j.Strategy.RawMatrix.Kind == yaml.MappingNode {
		var val map[string][]interface{}
		if err := j.Strategy.RawMatrix.Decode(&val); err != nil {
			log.Fatal(err)
		}
		fmt.Println(val)
		return val
	}
	return nil
}
