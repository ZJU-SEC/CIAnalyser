package model

import (
	"fmt"
	"github.com/robertkrimen/otto"
	"gopkg.in/yaml.v3"
	"reflect"
	"regexp"
	"strings"
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
	FailFastString    string    `yaml:"fail-fast"`
	MaxParallelString string    `yaml:"max-parallel"`
	RawMatrix         yaml.Node `yaml:"matrix"`
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

func commonKeysMatch(a map[string]interface{}, b map[string]interface{}) bool {
	for aKey, aVal := range a {
		if bVal, ok := b[aKey]; ok && !reflect.DeepEqual(aVal, bVal) {
			return false
		}
	}
	return true
}

// ContainerSpec is the specification of the container to usecases for the job
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

// Environments returns string-based key=value map for a step
func (s *Step) Environment() map[string]string {
	return environment(s.Env)
}

// GetEnv gets the env for a step
func (s *Step) GetEnv() map[string]string {
	env := s.Environment()

	if env == nil {
		env = make(map[string]string)
	}

	for k, v := range s.With {
		envKey := regexp.MustCompile("[^A-Z0-9-]").ReplaceAllString(strings.ToUpper(k), "_")
		envKey = fmt.Sprintf("INPUT_%s", strings.ToUpper(envKey))
		env[envKey] = v
	}
	return env
}

// RunsOn list for Job. Note that RunsOn will interpolate matrix automatically
func (j *Job) RunsOn() []string {
	switch j.RawRunsOn.Kind {
	case yaml.ScalarNode:
		var val string
		err := j.RawRunsOn.Decode(&val)
		if err != nil {
			return nil
		}

		if !strings.Contains(val, "${{") || !strings.Contains(val, "}}") {
			return []string{val}
		} else {
			var runners []string

			for _, matrix := range j.GetMatrixes() {
				vm := otto.New()
				vm.Set("matrix", matrix)
				runners = append(runners, interpolate(val, vm))
			}
			return runners
		}
	case yaml.SequenceNode:
		var val []string
		err := j.RawRunsOn.Decode(&val)
		if err != nil {
			return nil
		}
		return val
	}
	return nil
}

func interpolate(in string, vm *otto.Otto) string {
	pattern := regexp.MustCompile(`\${{\s*(.+?)\s*}}`)
	out := in
	for {
		out = pattern.ReplaceAllStringFunc(in, func(match string) string {
			// Extract and trim the actual expression inside ${{...}} delimiters
			expression := pattern.ReplaceAllString(match, "$1")

			// Evaluate the expression and retrieve errors if any
			rawExpr, _ := vm.Run(expression)
			expr, _ := rawExpr.ToString()

			return expr
		})

		if out == in {
			// No replacement occurred, we're done!
			break
		}
		in = out
	}

	return out
}

func environment(yml yaml.Node) map[string]string {
	env := make(map[string]string)
	if yml.Kind == yaml.MappingNode {
		if err := yml.Decode(&env); err != nil {
			return nil
		}
	}
	return env
}

// Environment returns string-based key=value map for a job
func (j *Job) Environment() map[string]string {
	return environment(j.Env)
}

// Matrix decodes RawMatrix YAML node
func (j *Job) Matrix() map[string][]interface{} {
	if j.Strategy == nil {
		return nil
	}

	if j.Strategy.RawMatrix.Kind == yaml.MappingNode {
		var val map[string][]interface{}
		if err := j.Strategy.RawMatrix.Decode(&val); err != nil {
			return nil
		}
		return val
	}
	return nil
}

// GetMatrixes returns the matrix cross product
// It skips includes and hard fails excludes for non-existing keys
// nolint:gocyclo
func (j *Job) GetMatrixes() []map[string]interface{} {
	matrixes := make([]map[string]interface{}, 0)
	if j.Strategy != nil {
		if m := j.Matrix(); m != nil {
			includes := make([]map[string]interface{}, 0)
			for _, v := range m["include"] {
				switch t := v.(type) {
				case []interface{}:
					for _, i := range t {
						i := i.(map[string]interface{})
						for k := range i {
							if _, ok := m[k]; ok {
								includes = append(includes, i)
								break
							}
						}
					}
				case interface{}:
					v := v.(map[string]interface{})
					includes = append(includes, v)
				}
			}
			delete(m, "include")

			excludes := make([]map[string]interface{}, 0)
			for _, e := range m["exclude"] {
				e := e.(map[string]interface{})
				for k := range e {
					if _, ok := m[k]; ok {
						excludes = append(excludes, e)
					}
				}
			}
			delete(m, "exclude")

			matrixProduct := cartesianProduct(m)
		MATRIX:
			for _, matrix := range matrixProduct {
				for _, exclude := range excludes {
					if commonKeysMatch(matrix, exclude) {
						continue MATRIX
					}
				}
				matrixes = append(matrixes, matrix)
			}
			for _, include := range includes {
				matrixes = append(matrixes, include)
			}
		} else {
			matrixes = append(matrixes, make(map[string]interface{}))
		}
	} else {
		matrixes = append(matrixes, make(map[string]interface{}))
	}
	return matrixes
}

// cartesianProduct takes map of lists and returns list of unique tuples
func cartesianProduct(mapOfLists map[string][]interface{}) []map[string]interface{} {
	listNames := make([]string, 0)
	lists := make([][]interface{}, 0)
	for k, v := range mapOfLists {
		listNames = append(listNames, k)
		lists = append(lists, v)
	}

	listCart := cartN(lists...)

	rtn := make([]map[string]interface{}, 0)
	for _, list := range listCart {
		vMap := make(map[string]interface{})
		for i, v := range list {
			vMap[listNames[i]] = v
		}
		rtn = append(rtn, vMap)
	}
	return rtn
}

func cartN(a ...[]interface{}) [][]interface{} {
	c := 1
	for _, a := range a {
		c *= len(a)
	}
	if c == 0 || len(a) == 0 {
		return nil
	}
	p := make([][]interface{}, c)
	b := make([]interface{}, c*len(a))
	n := make([]int, len(a))
	s := 0
	for i := range p {
		e := s + len(a)
		pi := b[s:e]
		p[i] = pi
		s = e
		for j, n := range n {
			pi[j] = a[j][n]
		}
		for j := len(n) - 1; j >= 0; j-- {
			n[j]++
			if n[j] < len(a[j]) {
				break
			}
			n[j] = 0
		}
	}
	return p
}
