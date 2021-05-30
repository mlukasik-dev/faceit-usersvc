package store

import "github.com/gookit/validate"

var fieldAliases = validate.MS{
	"User.FirstName": "first_name",
	"User.LastName":  "last_name",
	"User.Nickname":  "nickname",
	"User.Email":     "email",
	"User.Country":   "country",
}

type ValidationErrors struct {
	Errors map[string]string
}

func (e *ValidationErrors) Error() string {
	return "Invalid input data"
}

func (e *ValidationErrors) One() string {
	for _, one := range e.Errors {
		return one
	}
	return ""
}

func fromValidateErrors(errs validate.Errors) *ValidationErrors {
	m := make(map[string]string)
	for field, messages := range errs.All() {
		for _, message := range messages {
			// Asign a first message and get out of the loop.
			m[fieldAliases[field]] = message
			break
		}
	}
	return &ValidationErrors{m}
}

type userValidation struct {
	*User
	ValidationKind string
}

func (v userValidation) Messages() map[string]string {
	return validate.MS{
		"alpha":       "The field '{field}' should contain only apha characters.",
		"alphaNum":    "The field '{field}' should contain only apha-numeric characters.",
		"email":       "The field '{field}' is not a valid email.",
		"required_if": "The field '{field}' is required.",
	}
}

func (v userValidation) Translates() map[string]string {
	return fieldAliases
}

const (
	CreateValidationKind = "create"
	UpdateValidationKind = "update"
	FilterValidationKind = "filter"
)

// Validate validates user according to validation kind.
func (u *User) Validate(kind string) *ValidationErrors {
	v := validate.Struct(userValidation{u, kind})
	v.StopOnError = false
	if !v.Validate() {
		return fromValidateErrors(v.Errors)
	}
	return nil
}
