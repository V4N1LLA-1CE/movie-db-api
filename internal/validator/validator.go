package validator

import (
	"regexp"
	"slices"
)

var (
	EmailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

// helper to create new validator with empty errors map
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// return true if there are no errors
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// add an error message to map if no entry exists for the key
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// add an error message to map if validation check is not 'ok'
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// generic func that returns true if value is in list of permitted values
// comparable types include: numbers, strings, booleans, pointers, channels, interfaces, arrays, and structs of comparable types
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// returns true if string value matches specific regex pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// generic func returns true if all vals in slice are unique
// make and insert keyvals into new map (will overwrite if same value so no duplicates)
// compare length of new map with length of old map and return result
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, val := range values {
		uniqueValues[val] = true
	}

	return len(values) == len(uniqueValues)
}
