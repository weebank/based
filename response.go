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
			for _, v := range form.Fields {
				if l.ID == v {
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
					break
				}
			}

		}
	}
	return errs
}
