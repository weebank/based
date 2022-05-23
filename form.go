package based

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
	Param  string                 `json:"param"`
	Props  map[string]interface{} `json:"props,omitempty"`
}

type Item struct {
	ID    string                 `json:"id"`
	Item  string                 `json:"item"`
	Props map[string]interface{} `json:"props,omitempty"`
	Rules []Rule                 `json:"rules,omitempty"`
}

type Form struct {
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
	Fields  []string `json:"fields"`
	Layout  []Item   `json:"layout"`
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

	ver, ok := raw["_version"]
	if !ok {
		errs = append(errs, errors.New("form \""+formName+"\" has no \"_version\" field"))
		return
	}
	if _, ok := ver.(int); !ok {
		errs = append(errs, errors.New("form \""+formName+"\" has a \"_version\" field that is not a number"))
		return
	}

	items, ok := raw["_items"]
	if !ok {
		errs = append(errs, errors.New("form \""+formName+"\" has no \"_items\" field"))
		return
	}
	if ty := reflect.ValueOf(items); ty.Kind() != reflect.Slice {
		errs = append(errs, errors.New("form \""+formName+"\" has an \"_items\" field that is not a list"))
		return
	}

	form = &Form{Name: formName, Actions: []string{}, Fields: []string{}, Layout: []Item{}}

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
		item := Item{ID: key, Props: map[string]interface{}{}, Rules: []Rule{}}

		it, ok := v["_item"]
		if !ok {
			errs = append(errs, errors.New("item \""+key+"\" has no \"_item\" field"))
			continue
		}
		if _, ok := it.(string); !ok {
			errs = append(errs, errors.New("item \""+key+"\" has an \"_item\" field that is not a string"))
		}
		item.Item = it.(string)

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

		if rules, ok := v["_rules"]; ok {
			if ty := reflect.ValueOf(rules); ty.Kind() != reflect.Slice {
				errs = append(errs, errors.New("item \""+key+"\" has a \"_rules\" field that is not a list"))
				continue
			}
			for i, rule := range rules.([]interface{}) {
				rule, ok := rule.(map[interface{}]interface{})
				if !ok {
					errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") that is not an object"))
					continue
				}

				isInvalid := false

				action, ok := rule["_action"]
				if !ok {
					errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") that has no \"_action\" field"))
					isInvalid = true
				} else if _, ok := action.(string); !ok {
					errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") whose \"_action\" field is not a string"))
					isInvalid = true
				}

				param, ok := rule["_param"]
				if !ok {
					errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") that has no \"_param\" field"))
					isInvalid = true
				} else if _, ok := param.(string); !ok {
					errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") whose \"_param\" field is not a string"))
					isInvalid = true
				}

				if isInvalid {
					continue
				}

				switch action {
				case "==", "!=":
				case "regex":
					_, err := regexp.Compile(param.(string))
					if err != nil {
						fmt.Println(err)
						errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") whose \"_action\" is \"regex\" but its \"_param\" is not a valid regex"))
						continue
					}
				default:
					errs = append(errs, errors.New("item \""+key+"\" has a rule ("+fmt.Sprint(i)+") has an unknown \"_action\": "+action.(string)))
					continue
				}

				r := Rule{Action: action.(string), Param: param.(string), Props: map[string]interface{}{}}
				for k, v := range rule {
					key, ok := k.(string)
					if !ok {
						continue
					}
					if !strings.HasPrefix(key, "_") {
						r.Props[key] = v
					}
				}

				item.Rules = append(item.Rules, r)
			}
		}

		for k, p := range v {
			if !strings.HasPrefix(k.(string), "_") {
				item.Props[k.(string)] = p
			}
		}

		form.Layout = append(form.Layout, item)
	}

	return
}
