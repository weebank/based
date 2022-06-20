package form

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type ResponseErrors []error

type ResponseCollection map[string]string

func (re ResponseErrors) Error() (s string) {
	str := []string{}
	for _, e := range re {
		str = append(str, e.Error())
	}
	return strings.Join(str, ", ")
}

func SanitizeResponse(responses *ResponseCollection, form *Form) {
	for k := range *responses {
		var isValid bool
		for _, f := range form.Fields {
			if k == f {
				isValid = true
			}
		}

		if !isValid {
			delete(*responses, k)
		}
	}
}

func ValidateResponse(responses ResponseCollection, form *Form) ResponseErrors {
	errs := ResponseErrors{}
	for _, f := range form.Fields {
		_, ok := responses[f]
		if !ok {
			errs = append(errs, errors.New(f+" has no matching response field"))
			continue
		}
	}

	for _, l := range form.Layout {
		for _, i := range l {
			if _, ok := responses[i.ID]; ok {
				for _, v := range form.Fields {
					if i.ID == v {
						ruleErrs := ValidateRule(*i.Rule, 0, i.ID, responses)
						errs = append(errs, ruleErrs)
						break
					}
				}

			}
		}
	}
	return errs
}

func ValidateRule(r Rule, i int, ID string, response ResponseCollection) (errs ResponseErrors) {
	var matchesRule bool
	switch r.Action {
	case OR:
		matchesRule = false
		ruleErrs := ResponseErrors{}
		for i, rule := range r.Param.([]Rule) {
			ruleErrs = append(ruleErrs, ValidateRule(rule, i, ID, response))
			if len(ruleErrs) == 0 {
				matchesRule = true
				break
			}
		}
		errs = append(errs, ruleErrs)
	case AND:
		matchesRule = true
		for i, rule := range r.Param.([]Rule) {
			ruleErrs := ValidateRule(rule, i, ID, response)
			if len(ruleErrs) > 0 {
				errs = append(errs, ruleErrs)
				matchesRule = false
				break
			}
		}
	case EQ:
		matchesRule = response[r.Param.(string)] == response[ID]
	case INEQ:
		matchesRule = response[r.Param.(string)] != response[ID]
	case REGEX:
		reg := regexp.MustCompile(r.Param.(string))
		matchesRule = reg.Match([]byte(response[ID]))
	}
	if !matchesRule {
		errs = append(errs, errors.New(ID+" does not match rule "+fmt.Sprint(i)))
	}
	return
}
