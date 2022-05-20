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

func ValidateResponse(response map[string]string, form *Form) ResponseErrors {
	errs := ResponseErrors{}
	for _, f := range form.Fields {
		_, ok := response[f]
		if !ok {
			errs = append(errs, errors.New(f+" has no matching response field"))
			continue
		}
	}

	for _, l := range form.Layout {
		if res, ok := response[l.ID]; ok {
			for i, r := range l.Rules {
				var matchesRule bool
				switch r.Action {
				case "==":
					matchesRule = response[r.Param] == res
				case "!=":
					matchesRule = response[r.Param] != res
				case "regex":
					reg := regexp.MustCompile(r.Param)
					matchesRule = reg.Match([]byte(res))
				}
				if !matchesRule {
					errs = append(errs, errors.New(l.ID+" does not match rule "+fmt.Sprint(i)))
				}
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

	form, errs := CompileForm("forms/sign-up.yml")
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

	b, _ := json.MarshalIndent(form, "", "    ")
	fmt.Println(string(b))
}
