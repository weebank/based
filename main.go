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
	Action string
	Param  string
	Fields map[interface{}]interface{}
}

type Field struct {
	Type  string
	Rules []Rule
}

type Form struct {
	Name    string
	Items   map[string]map[string]interface{}
	Actions []string
	Fields  map[string]Field
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

	form = &Form{Name: name, Items: map[string]map[string]interface{}{}, Actions: []string{}, Fields: map[string]Field{}}
	err = yaml.Unmarshal(file, &(form.Items))
	if err != nil {
		errs = append(errs, err)
		return
	}

	for k, v := range form.Items {
		if ty, ok := v["_item"]; ok {
			if _, ok := ty.(string); !ok {
				errs = append(errs, errors.New("item \""+k+"\" has a \"_item\" field that is not a string"))
				continue
			}
		} else {
			errs = append(errs, errors.New("item \""+k+"\" has a \"_item\" field that is not a string"))
			continue
		}

		ty, ok := v["_type"]
		if !ok {
			continue
		}

		if _, ok := ty.(string); !ok {
			errs = append(errs, errors.New("item \""+k+"\" has an \"_type\" field that is not a string"))
			continue
		}
		if ty == "action" {
			form.Actions = append(form.Actions, ty.(string))
		} else {
			switch ty {
			case "none":
			case "bool", "number", "string", "object", "array":
				if rules, ok := v["_rules"]; ok {
					if tyRules := reflect.ValueOf(rules); tyRules.Kind() != reflect.Slice {
						errs = append(errs, errors.New("item \""+k+"\" has a \"_rules\" field that is not a list"))
						continue
					}

					form.Fields[k] = Field{
						Type:  ty.(string),
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
						} else {
							delete(rule, "_action")
						}

						param, ok := rule["_param"]
						if !ok {
							errs = append(errs, errors.New("item \""+k+"\" has a rule ("+fmt.Sprint(i)+") that has no \"_param\" field"))
							isInvalid = true
						} else if _, ok := param.(string); !ok {
							errs = append(errs, errors.New("item \""+k+"\" has a rule ("+fmt.Sprint(i)+") whose \"_param\" field is not a string"))
							isInvalid = true
						} else {
							delete(rule, "_param")
						}

						if isInvalid {
							continue
						}

						switch action {
						case "==", "!=", "<", "<=", ">", ">=":
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

						field := form.Fields[k]
						field.Rules = append(form.Fields[k].Rules, Rule{Action: action.(string), Param: param.(string), Fields: rule})
						form.Fields[k] = field
					}
				}
			default:
				errs = append(errs, errors.New("item \""+k+"\" has an unknown \"_type\": "+ty.(string)))
			}
		}

	}

	return
}

type ResponseErrors []error

func (re ResponseErrors) Error() (s string) {
	str := []string{}
	for _, e := range re {
		str = append(str, e.Error())
	}
	return strings.Join(str, ", ")
}

func SanitizeResponse(response *map[string]interface{}, form *Form) {
	for k := range *response {
		if _, ok := form.Items[k]; !ok {
			delete(*response, k)
		}
	}
}

func ValidateResponse(response map[string]interface{}, form *Form) ResponseErrors {
	errs := ResponseErrors{}
	for k, v := range form.Fields {
		if r, ok := response[k]; !ok {
			errs = append(errs, errors.New(k+" has no matching response field"))
		} else {
			hasWrongType := false
			t := reflect.ValueOf(r)
			switch t.Kind() {
			case reflect.Bool:
				hasWrongType = v.Type != "bool"
			case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
				hasWrongType = v.Type != "number"
			case reflect.String:
				hasWrongType = v.Type != "string"
			case reflect.Map:
				_, ok := r.(map[string]interface{})
				hasWrongType = !ok || v.Type != "object"
			case reflect.Slice:
				hasWrongType = v.Type != "array"
			default:
				hasWrongType = true
			}
			if hasWrongType {
				errs = append(errs, errors.New(k+" has wrong type"))
			}
		}
	}
	return errs
}

func main() {
	form, errs := LoadForm("sign-up")
	m := map[string]interface{}{
		"name":      "1",
		"test":      map[string]interface{}{"2": "a"},
		"mama":      "eu",
		"email":     "eaemen",
		"password1": "asd",
	}
	if len(errs) != 0 {
		fmt.Println(errs.Error())
	}
	SanitizeResponse(&m, form)
	errs1 := ValidateResponse(m, form)
	if len(errs1) != 0 {
		fmt.Println(errs1.Error())
	}
}
