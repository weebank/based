package form

import (
	"errors"
	"regexp"
	"strconv"
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

func SanitizeResponse(form *Form, step string, responses *ResponseCollection) {
	for field := range *responses {
		if _, ok := form.Steps[step][field]; !ok {
			delete(*responses, field)
		}
	}
}

func ValidateResponse(form *Form, step string, responses ResponseCollection) (err ResponseErrors) {
	for name, field := range form.Steps[step] {
		response, ok := responses[name]
		if !ok {
			err = append(err, errors.New(name+" has no matching response field"))
			continue
		}

		ruleErr := ValidateRule(name, field.Rule, response)
		if ruleErr != nil {
			err = append(err, ruleErr)
		}
	}

	if len(err) > 0 {
		return err
	} else {
		return nil
	}
}

func ValidateRule(field string, rule Rule, response string) (err ResponseErrors) {
	var matchesRule bool
	switch rule.Op {
	case OR:
		matchesRule = false
		for _, r := range rule.Param.([]Rule) {
			ruleErr := ValidateRule(field, r, response)
			if ruleErr == nil {
				matchesRule = true
				break
			} else {
				err = append(err, ruleErr)
			}
		}
	case AND:
		matchesRule = true
		for _, r := range rule.Param.([]Rule) {
			ruleErr := ValidateRule(field, r, response)
			if ruleErr != nil {
				err = append(err, ruleErr)
				matchesRule = false
				break
			}
		}
	case EQ:
		number, _ := strconv.ParseFloat(response, 64)
		matchesRule = number == rule.Param.(float64)
	case NEQ:
		number, _ := strconv.ParseFloat(response, 64)
		matchesRule = number != rule.Param.(float64)
	case GT:
		number, _ := strconv.ParseFloat(response, 64)
		matchesRule = number > rule.Param.(float64)
	case GTE:
		number, _ := strconv.ParseFloat(response, 64)
		matchesRule = number >= rule.Param.(float64)
	case LT:
		number, _ := strconv.ParseFloat(response, 64)
		matchesRule = number < rule.Param.(float64)
	case LTE:
		number, _ := strconv.ParseFloat(response, 64)
		matchesRule = number <= rule.Param.(float64)
	case REGEX:
		reg := regexp.MustCompile(rule.Param.(string))
		matchesRule = reg.Match([]byte(response))
	}
	if !matchesRule {
		err = append(err, errors.New(field+" does not match the assigned rule"))
	}

	if len(err) > 0 {
		return err
	} else {
		return nil
	}
}
