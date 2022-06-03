package based

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type ResponseErrors []error

func (re ResponseErrors) Error() (s string) {
	str := []string{}
	for _, e := range re {
		str = append(str, e.Error())
	}
	return strings.Join(str, ", ")
}

func SanitizeResponse(response *map[string]string, form *Form) {
	for k := range *response {
		var isValid bool
		for _, f := range form.Fields {
			if k == f {
				isValid = true
			}
		}

		if !isValid {
			delete(*response, k)
		}
	}
}

func ValidateResponse(responses map[string]string, form *Form) ResponseErrors {
	errs := ResponseErrors{}
	for _, f := range form.Fields {
		_, ok := responses[f]
		if !ok {
			errs = append(errs, errors.New(f+" has no matching response field"))
			continue
		}
	}

	for _, l := range form.Layout {
		if _, ok := responses[l.ID]; ok {
			for _, v := range form.Fields {
				if l.ID == v {
					for i, r := range l.Rules {
						ruleErrs := ValidateRule(r, i, l.ID, responses)
						errs = append(errs, ruleErrs)
					}
					break
				}
			}

		}
	}
	return errs
}

func ValidateRule(r Rule, i int, ID string, responses map[string]string) (errs ResponseErrors) {
	var matchesRule bool
	switch r.Action {
	case OR:
		matchesRule = false
		ruleErrs := ResponseErrors{}
		for i, rule := range r.Param.([]Rule) {
			ruleErrs = append(ruleErrs, ValidateRule(rule, i, ID, responses))
			if len(ruleErrs) == 0 {
				matchesRule = true
				break
			}
		}
		errs = append(errs, ruleErrs)
	case AND:
		matchesRule = true
		for i, rule := range r.Param.([]Rule) {
			ruleErrs := ValidateRule(rule, i, ID, responses)
			if len(ruleErrs) > 0 {
				errs = append(errs, ruleErrs)
				matchesRule = false
				break
			}
		}
	case EQ:
		matchesRule = responses[r.Param.(string)] == responses[ID]
	case INEQ:
		matchesRule = responses[r.Param.(string)] != responses[ID]
	case REGEX:
		reg := regexp.MustCompile(r.Param.(string))
		matchesRule = reg.Match([]byte(responses[ID]))
	}
	if !matchesRule {
		errs = append(errs, errors.New(ID+" does not match rule "+fmt.Sprint(i)))
	}
	return
}
