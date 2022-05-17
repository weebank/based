package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type InputType uint8

type Rule struct {
	Action string                 `json:"action"`
	Param  string                 `json:"param"`
	Props  map[string]interface{} `json:"props,omitempty"`
}

type Field struct {
	Rules []Rule `json:"rules,omitempty"`
}

type Item struct {
	ID    string                 `json:"id"`
	Item  string                 `json:"item"`
	Props map[string]interface{} `json:"props,omitempty"`
}

type Form struct {
	Name    string           `json:"name"`
	Actions []string         `json:"actions"`
	Fields  map[string]Field `json:"fields"`
	Layout  []Item           `json:"layout"`
}

type FormErrors []error

func (fe FormErrors) Error() (s string) {
	str := []string{}
	for _, e := range fe {
		str = append(str, e.Error())
	}
	return strings.Join(str, ", ")
}

func LoadForm(name string) (form *Form, errs FormErrors) {
	file, err := ioutil.ReadFile("forms/" + name + ".yml")
	if err != nil {
		errs = append(errs, err)
		return
	}

	raw := map[string]map[string]interface{}{}
	err = yaml.Unmarshal(file, &raw)
	if err != nil {
		errs = append(errs, err)
		return
	}

	form = &Form{Name: name, Actions: []string{}, Fields: map[string]Field{}, Layout: []Item{}}

	for k, v := range raw {
		item := Item{ID: k, Props: map[string]interface{}{}}

		it, ok := v["_item"]
		if !ok {
			errs = append(errs, errors.New("item \""+k+"\" has no \"_item\" field"))
			continue
		}
		if _, ok := it.(string); !ok {
			errs = append(errs, errors.New("item \""+k+"\" has an \"_item\" field that is not a string"))
		}
		item.Item = it.(string)

		ty, ok := v["_type"]
		if !ok {
			continue
		}
		if _, ok := ty.(string); !ok {
			errs = append(errs, errors.New("item \""+k+"\" has a \"_type\" field that is not a string"))
			continue
		}

		switch ty {
		case "none":
		case "action":
			form.Actions = append(form.Actions, k)
		case "field":
			if rules, ok := v["_rules"]; ok {
				if tyRules := reflect.ValueOf(rules); tyRules.Kind() != reflect.Slice {
					errs = append(errs, errors.New("item \""+k+"\" has a \"_rules\" field that is not a list"))
					continue
				}

				form.Fields[k] = Field{
					Rules: []Rule{},
				}

				for i, rule := range rules.([]interface{}) {
					rule, ok := rule.(map[interface{}]interface{})
					if !ok {
						errs = append(errs, errors.New("item \""+k+"\" has a rule ("+fmt.Sprint(i)+") that is not an object"))
						continue
					}

					isInvalid := false

					action, ok := rule["_action"]
					if !ok {
						errs = append(errs, errors.New("item \""+k+"\" has a rule ("+fmt.Sprint(i)+") that has no \"_action\" field"))
						isInvalid = true
					} else if _, ok := action.(string); !ok {
						errs = append(errs, errors.New("item \""+k+"\" has a rule ("+fmt.Sprint(i)+") whose \"_action\" field is not a string"))
						isInvalid = true
					}

					param, ok := rule["_param"]
					if !ok {
						errs = append(errs, errors.New("item \""+k+"\" has a rule ("+fmt.Sprint(i)+") that has no \"_param\" field"))
						isInvalid = true
					} else if _, ok := param.(string); !ok {
						errs = append(errs, errors.New("item \""+k+"\" has a rule ("+fmt.Sprint(i)+") whose \"_param\" field is not a string"))
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
							errs = append(errs, errors.New("item \""+k+"\" has a rule ("+fmt.Sprint(i)+") whose \"_action\" is \"regex\" but its \"_param\" is not a valid regex"))
							continue
						}
					default:
						errs = append(errs, errors.New("item \""+k+"\" has a rule ("+fmt.Sprint(i)+") has an unknown \"_action\": "+action.(string)))
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

					field := form.Fields[k]
					field.Rules = append(form.Fields[k].Rules, r)
					form.Fields[k] = field
				}
			}
		default:
			errs = append(errs, errors.New("item \""+k+"\" has an unknown \"_type\": "+ty.(string)))
		}

		for k, p := range v {
			if !strings.HasPrefix(k, "_") {
				item.Props[k] = p
			}
		}

		form.Layout = append(form.Layout, item)
	}

	return
}
