package form

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Field struct {
	Hidden bool `json:"hidden,omitempty"`
	Rule   Rule `json:"rule"`
}

type Rule struct {
	Op    string                 `json:"op"`
	Param interface{}            `json:"param"`
	Props map[string]interface{} `json:"props,omitempty"`
}

const (
	AND   = "&&"
	OR    = "||"
	EQ    = "=="
	NEQ   = "!="
	LT    = "<"
	GT    = ">"
	LTE   = "<="
	GTE   = ">="
	REGEX = "regex"
)

type Form struct {
	Name  string                      `json:"name"`
	Steps map[string]map[string]Field `json:"steps"`
}

type FormErrors []error

func (fe FormErrors) Error() (s string) {
	str := []string{}
	for _, e := range fe {
		str = append(str, e.Error())
	}
	return strings.Join(str, ", ")
}

// Load form from file
func loadForm(path string) (raw map[string]interface{}, err error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(file, &raw)
	if err != nil {
		return nil, err
	}

	return
}

// Load form from file and compile it
func CompileForm(path string) (form *Form, errs FormErrors) {
	raw, err := loadForm(path)
	if err != nil {
		errs = append(errs, errors.New("failed to load form: "+err.Error()))
		return
	}

	fileName := strings.Split(filepath.Base(path), ".")
	name := strings.Join(fileName[:len(fileName)-1], "")

	form = &Form{Name: name, Steps: make(map[string]map[string]Field)}

	// Iterate through steps
	for step, fields := range raw {
		// Need to cast to general map, not map[string]interface{}
		obj, ok := fields.(map[interface{}]interface{})
		if !ok {
			errs = append(errs, errors.New("step \""+fmt.Sprint(step)+"\" is not a valid field map"))
			continue
		}

		form.Steps[step] = make(map[string]Field)

		for name, field := range obj {
			// It's safe to assume that "name" is a string, otherwise the JSON would be invalid
			str := name.(string)

			// Check if field is a map
			f, ok := field.(map[interface{}]interface{})
			if !ok {
				errs = append(errs, errors.New("field \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\" is not a valid field object"))
				continue
			}

			// Field struct
			var newField Field

			// Check if field has a "hidden" property
			if hidden, ok := f["hidden"]; ok {
				isHidden, ok := hidden.(bool)
				if !ok {
					errs = append(errs, errors.New("field \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\"has a \"hidden\" field that is not a boolean"))
					continue
				}
				newField.Hidden = isHidden
			}

			// Check if field has a "rule" property
			if rule, ok := f["rule"]; ok {
				compiledRule, err := compileRule(rule, step, str)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				newField.Rule = compiledRule
			}

			// Add field to form
			form.Steps[step][str] = newField
		}
	}

	return
}

// Compile and validate a rule
func compileRule(obj interface{}, step, name string) (rule Rule, errs FormErrors) {
	// Check if rule is a map
	r, ok := obj.(map[interface{}]interface{})
	if !ok {
		errs = append(errs, errors.New("rule \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\" is not a valid rule"))
		return
	}

	// Check if rule has an "op" property
	op, ok := r["op"]
	if !ok {
		errs = append(errs, errors.New("rule \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\" has no \"op\" field"))
		return
	} else if _, ok := op.(string); !ok {
		errs = append(errs, errors.New("rule \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\"has an \"op\" field that is not a string"))
		return
	}

	// Check if rule has a "param" property
	param, ok := r["param"]
	if !ok {
		errs = append(errs, errors.New("rule \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\" has no \"param\" field"))
		return
	}

	// Build rule
	rule = Rule{Op: op.(string), Props: map[string]interface{}{}}
	switch op {
	case AND, OR:
		if ty := reflect.ValueOf(param); ty.Kind() != reflect.Slice {
			errs = append(errs, errors.New("rule \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\" has an \"op\" field that demands a list parameter, but \"param\" field is not a list"))
			return
		}
		rules := []Rule{}
		for _, rule := range param.([]interface{}) {
			r, ruleErr := compileRule(rule, step, name)
			if ruleErr != nil {
				errs = append(errs, ruleErr)
			}
			rules = append(rules, r)
		}
		rule.Param = rules
	case EQ, NEQ, GT, GTE, LT, LTE:
		if _, ok := param.(float64); !ok {
			errs = append(errs, errors.New("rule \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\" has an \"op\" field that demands a number parameter, but \"param\" field is not a number"))
			return
		}

		rule.Param = param
	case REGEX:
		if _, ok := param.(string); !ok {
			errs = append(errs, errors.New("rule \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\" has an \"op\" field that demands a string parameter, but \"param\" field is not a string"))
			return
		}

		rule.Param = param

		_, err := regexp.Compile(param.(string))
		if err != nil {
			log.Println(err)
			errs = append(errs, errors.New("rule \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\" has an \"op\" field that demands a regex parameter, but \"param\" field is not a valid regex"))
			return
		}
	default:
		errs = append(errs, errors.New("rule \""+fmt.Sprint(name)+"\" from step \""+fmt.Sprint(step)+"\" has an unknown \"op\": "+op.(string)))
		return
	}

	// Add custom props
	for k, v := range r {
		prop, ok := k.(string)
		if !ok {
			continue
		}
		switch prop {
		case "op", "param":
		default:
			rule.Props[prop] = v
		}
	}
	return
}
