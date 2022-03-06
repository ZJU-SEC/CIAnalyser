package analyzer

import (
	"CIHunter/src/utils"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"reflect"
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

// RunsOn list for Job. Note that RunsOn will interpolate matrix automatically
func (j *Job) RunsOn() []string {
	switch j.RawRunsOn.Kind {
	case yaml.ScalarNode:
		var val string
		err := j.RawRunsOn.Decode(&val)
		if err != nil {
			log.Fatal(err)
		}

		if !strings.Contains(val, "${{") || !strings.Contains(val, "}}") {
			return []string{val}
		} else {
			matrixes := j.GetMatrixes()
			var osList []string
			for _, ele := range matrixes {
				for k, v := range ele {
					if fmt.Sprint(k) == "os" || fmt.Sprint(k) == "platform" {
						osList = append(osList, fmt.Sprint(v))
					}
				}
			}
			return osList
		}
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
					for k := range v {
						if _, ok := m[k]; ok {
							includes = append(includes, v)
							break
						}
					}
				}
			}
			delete(m, "include")

			excludes := make([]map[string]interface{}, 0)
			for _, e := range m["exclude"] {
				e := e.(map[string]interface{})
				for k := range e {
					if _, ok := m[k]; ok {
						excludes = append(excludes, e)
					} else {
						// We fail completely here because that's what GitHub does for non-existing matrix keys, fail on exclude, silent skip on include
						log.Fatalf("The workflow is not valid. Matrix exclude key '%s' does not match any key within the matrix", k)
					}
				}
			}
			delete(m, "exclude")

			matrixProduct := utils.CartesianProduct(m)
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
