package models

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Patterns are anchored at the start only, so a value such as "updated_at"
// or "gh/{*}" still passes when its prefix matches.
var (
	keywordPattern = regexp.MustCompile(`^([a-zA-Z0-9\-\/]+)`)
	lettersPattern = regexp.MustCompile(`^([a-zA-Z]+)`)
	validate       = newValidator()
)

func newValidator() *validator.Validate {
	v := validator.New()
	if err := v.RegisterValidation("keywordfmt", matchPattern(keywordPattern)); err != nil {
		panic(err)
	}
	if err := v.RegisterValidation("lettersfmt", matchPattern(lettersPattern)); err != nil {
		panic(err)
	}
	return v
}

func matchPattern(re *regexp.Regexp) validator.Func {
	return func(fl validator.FieldLevel) bool {
		return re.MatchString(fl.Field().String())
	}
}
