package form

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Rule struct {
	Action string                 `json:"action"`
	Param  interface{}            `json:"param"`
	Props  map[string]interface{} `json:"props,omitempty"`
}

const (
	AND   = "and"
	OR    = "or"
	EQ    = "=="
	INEQ  = "!="
	REGEX = "regex"
)

type Item struct {
	ID    string                 `json:"id"`
	Props map[string]interface{} `json:"props,omitempty"`
	Rule  *Rule                  `json:"rule,omitempty"`
}

type Group struct {
	Class string `json:"class"`
	Items []Item `json:"items"`
}

type Form struct {
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
	Fields  []string `json:"fields"`
	Layout  []Group  `json:"layout"`
}

type FormErrors []error

func (fe FormErrors) Error() (s string) {
	str := []string{}
	for _, e := range fe {
		str = append(str, e.Error())
	}
	return strings.Join(str, ", ")
}

func LoadForm(path string) (raw map[string]interface{}, err error) {
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

func CompileForm(path string) (form *Form, errs FormErrors) {
	raw, err := LoadForm(path)
	if err != nil {
		errs = append(errs, errors.New("failed to load form: "+err.Error()))
		return
	}

	name := strings.Split(filepath.Base(path), ".")
	formName := strings.Join(name[:len(name)-1], "")

	items, ok := raw["_items"]
	if !ok {
		errs = append(errs, errors.New("form \""+formName+"\" has no \"_items\" field"))
		return
	}
	if ty := reflect.ValueOf(items); ty.Kind() != reflect.Slice {
		errs = append(errs, errors.New("form \""+formName+"\" has an \"_items\" field that is not a list"))
		return
	}

	form = &Form{Name: formName, Actions: make([]string, 0), Fields: make([]string, 0), Layout: make([]Group, 0)}
	group := Group{Items: make([]Item, 0)}
	for k, i := range items.([]interface{}) {
		v := i.(map[interface{}]interface{})
		id, ok := v["_id"]
		if !ok {
			errs = append(errs, errors.New("item of index \""+fmt.Sprint(k)+"\" has no \"_id\" field"))
			continue
		}
		if _, ok := id.(string); !ok {
			errs = append(errs, errors.New("item of index \""+fmt.Sprint(k)+"\" has an \"_id\" field that is not a string"))
		}

		key := id.(string)
		item := Item{ID: key, Props: make(map[string]interface{})}

		cl, ok := v["_class"]
		if !ok {
			errs = append(errs, errors.New("item \""+key+"\" has no \"_class\" field"))
			continue
		}
		if _, ok := cl.(string); !ok {
			errs = append(errs, errors.New("item \""+key+"\" has an \"_class\" field that is not a string"))
		}
		class := cl.(string)

		if ty, ok := v["_type"]; ok {
			if _, ok := ty.(string); ok {
				switch ty {
				case "none":
				case "action":
					form.Actions = append(form.Actions, key)
				case "field":
					form.Fields = append(form.Fields, key)
				default:
					errs = append(errs, errors.New("item \""+key+"\" has an unknown \"_type\": "+ty.(string)))
				}
			} else {
				errs = append(errs, errors.New("item \""+key+"\" has a \"_type\" field that is not a string"))
				continue
			}
		}

		if rule, ok := v["_rule"]; ok {
			r := new(Rule)
			rule, ruleErrs := compileRule(rule, key, 0)
			*r = rule

			if len(ruleErrs) > 0 {
				errs = append(errs, ruleErrs)
			}

			item.Rule = r
		}

		for k, p := range v {
			if !strings.HasPrefix(k.(string), "_") {
				item.Props[k.(string)] = p
			}
		}

		if group.Class == "" {
			group.Class = class
		}
		if group.Class == class {
			group.Items = append(group.Items, item)
		} else {
			form.Layout = append(form.Layout, group)
			group = Group{Class: class, Items: []Item{item}}

		}
	}

	if group.Class != "" {
		form.Layout = append(form.Layout, group)
	}

	return
}

func compileRule(ruleObj interface{}, key string, i int) (r Rule, errs FormErrors) {
	rule, ok := ruleObj.(map[interface{}]interface{})
	if !ok {
		errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") that is not an object"))
		return
	}

	action, ok := rule["_action"]
	if !ok {
		errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") that has no \"_action\" field"))
		return
	} else if _, ok := action.(string); !ok {
		errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") whose \"_action\" field is not a string"))
		return
	}

	param, ok := rule["_param"]
	if !ok {
		errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") that has no \"_param\" field"))
		return
	}

	r = Rule{Action: action.(string), Props: map[string]interface{}{}}
	switch action {
	case AND, OR:
		if ty := reflect.ValueOf(param); ty.Kind() != reflect.Slice {
			errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") whose \"_action\" field demands a list parameter, but \"_param\" field is not a list"))
			return
		}
		rules := []Rule{}
		for i, rule := range param.([]interface{}) {
			r, ruleErrs := compileRule(rule, key, i)
			if len(ruleErrs) > 0 {
				errs = append(errs, ruleErrs)
			}
			rules = append(rules, r)
		}
		r.Param = rules
	case EQ, INEQ, REGEX:
		if _, ok := param.(string); !ok {
			errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") whose \"_action\" field demands a string parameter, but \"_param\" field is not a string"))
			return
		}
		r.Param = param
		if action != REGEX {
			break
		}

		_, err := regexp.Compile(param.(string))
		if err != nil {
			fmt.Println(err)
			errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") whose \"_action\" is \"regex\" but its \"_param\" is not a valid regex"))
			return
		}
	default:
		errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") has an unknown \"_action\": "+action.(string)))
		return
	}

	for k, v := range rule {
		key, ok := k.(string)
		if !ok {
			continue
		}
		if !strings.HasPrefix(key, "_") {
			r.Props[key] = v
		}
	}
	return
}
