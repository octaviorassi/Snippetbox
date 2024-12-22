package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	FieldErrors map[string]string
}

var EmailRx = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

/*	Valid returns True if there are no errors registered	*/
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}


func (v *Validator) AddFieldError(key, msg string) {

	// Initialize the map if it has not been done yet
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	// Add the msg to the key only if it did not exist
	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = msg
	}

}

/*	CheckField() adds an error message to the FieldErrors map only if a
	validation check is not 'ok'. */
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

/*	NotBlank() returns true if a value is not an empty string.	*/
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

/*	MaxChars() returns true if a value contains no more than n characters.	*/
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

/*	MinChars() returns true if a value contains at least n characters.	*/
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

/*	Matches() returns true if the given value matches the provided regexp	*/
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}





/*	PermittedValue() returns true if a value is in a list of specific permitted
	values.	*/
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}