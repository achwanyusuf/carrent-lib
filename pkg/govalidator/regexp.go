package govalidator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

const (
	alphaNumericSpaceRegexString = "^[a-zA-Z0-9 ]+$"
	rfc3339RegexString           = "^(\\d+)-(0[1-9]|1[012])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):([0-5]\\d):([0-5]\\d|60)(\\.\\d+)?(([Zz])|([\\+|\\-]([01]\\d|2[0-3]):([0-5]\\d|59)))$"
)

var (
	alphaNumericSpaceRegex = regexp.MustCompile(alphaNumericSpaceRegexString)
	rfc3339Regex           = regexp.MustCompile(rfc3339RegexString)
)

func isAlphaNumericSpace(fl validator.FieldLevel) bool {
	return alphaNumericSpaceRegex.MatchString(fl.Field().String())
}

func isRFC3339(fl validator.FieldLevel) bool {
	return rfc3339Regex.MatchString(fl.Field().String())
}

var validators = map[string]func(fl validator.FieldLevel) bool{
	"alphanumspace": isAlphaNumericSpace,
	"rfc3339":       isRFC3339,
}
