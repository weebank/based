package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
		if _, ok := form.Fields[k]; !ok {
			delete(*response, k)
		}
	}
}

func ValidateResponse(response map[string]string, form *Form) ResponseErrors {
	errs := ResponseErrors{}
	for k, v := range form.Fields {
		r, ok := response[k]
		if !ok {
			errs = append(errs, errors.New(k+" has no matching response field"))
			continue
		}

		for i, v := range v.Rules {
			var matchesRule bool
			switch v.Action {
			case "==":
				matchesRule = response[v.Param] == r
			case "!=":
				matchesRule = response[v.Param] != r
			case "regex":
				reg := regexp.MustCompile(v.Param)
				matchesRule = reg.Match([]byte(r))
			}
			if !matchesRule {
				errs = append(errs, errors.New(k+" does not match rule "+fmt.Sprint(i)))
			}
		}
	}
	return errs
}

func main() {
	var arg string
	if len(os.Args) > 1 {
		arg = os.Args[1]
	}

	form, errs := LoadForm("sign-up")
	m := map[string]string{}
	json.Unmarshal(
		[]byte(arg), &m,
	)
	if len(errs) > 0 {
		fmt.Println(errs.Error())
	}

	SanitizeResponse(&m, form)
	errs1 := ValidateResponse(m, form)
	if len(errs1) > 0 {
		fmt.Println(errs1.Error())
	}
}
