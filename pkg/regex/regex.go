package regex

import "regexp"

const (
	alphaNumericSpaceRegexString = "^[a-zA-Z0-9 ]+$"
	rfc3339RegexString           = "^(\\d+)-(0[1-9]|1[012])-(0[1-9]|[12]\\d|3[01])T([01]\\d|2[0-3]):([0-5]\\d):([0-5]\\d|60)(\\.\\d+)?(([Zz])|([\\+|\\-]([01]\\d|2[0-3]):([0-5]\\d|59)))$"
	urlRegexString               = "^(https:\\/\\/www\\.|http:\\/\\/www\\.|https:\\/\\/|http:\\/\\/)[a-zA-Z]{2,}(\\.[a-zA-Z]{2,})(\\.[a-zA-Z]{2,})?\\/[a-zA-Z0-9]{2,}|((https:\\/\\/www\\.|http:\\/\\/www\\.|https:\\/\\/|http:\\/\\/)[a-zA-Z]{2,}(\\.[a-zA-Z]{2,})(\\.[a-zA-Z]{2,})?)|(https:\\/\\/www\\.|http:\\/\\/www\\.|https:\\/\\/|http:\\/\\/)[a-zA-Z0-9]{2,}\\.[a-zA-Z0-9]{2,}\\.[a-zA-Z0-9]{2,}(\\.[a-zA-Z0-9]{2,})?$"
)

var (
	alphaNumericSpaceRegex = regexp.MustCompile(alphaNumericSpaceRegexString)
	rfc3339Regex           = regexp.MustCompile(rfc3339RegexString)
	urlRegex               = regexp.MustCompile(urlRegexString)
)

func IsAlphaNumericSpace(v string) bool {
	return alphaNumericSpaceRegex.MatchString(v)
}

func IsRFC3339(v string) bool {
	return rfc3339Regex.MatchString(v)
}

func IsURL(v string) bool {
	return urlRegex.MatchString(v)
}
