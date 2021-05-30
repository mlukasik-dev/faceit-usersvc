// +build unit

package store

import (
	"testing"

	"github.com/mlukasik-dev/faceit-usersvc/pkg/deref"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Validate(t *testing.T) {
	t.Run("all valid", func(t *testing.T) {
		t.Run("CreateValidationKind", func(t *testing.T) {
			user := &User{FirstName: "John", LastName: "Doe", Nickname: deref.StringAddr("johndoe1961"), Email: "john.doe@gmail.com", Country: "UK"}
			errs := user.Validate(CreateValidationKind)
			assert.Nil(t, errs)
		})
		t.Run("UpdateValidationKind", func(t *testing.T) {
			user := &User{Email: "new.john.doe@gmail.com"}
			errs := user.Validate(UpdateValidationKind)
			assert.Nil(t, errs)
		})
		t.Run("FilterValidationKind", func(t *testing.T) {
			user := &User{LastName: "Doe", Country: "UK"}
			errs := user.Validate(FilterValidationKind)
			assert.Nil(t, errs)
		})
	})

	t.Run("missed fields", func(t *testing.T) {
		t.Run("CreateValidationKind", func(t *testing.T) {
			user := &User{FirstName: "John", LastName: "Doe", Nickname: deref.StringAddr("johndoe1961")}
			errs := user.Validate(CreateValidationKind)
			require.NotNil(t, errs)
			assert.EqualValues(t, &ValidationErrors{
				map[string]string{"email": "The field 'email' is required.", "country": "The field 'country' is required."},
			}, errs)
		})
	})

	t.Run("invalid fields", func(t *testing.T) {
		t.Run("CreateValidationKind", func(t *testing.T) {
			user := &User{FirstName: "John", LastName: "Doe", Email: "john.doe#gmail.com", Country: "UK"}
			errs := user.Validate(CreateValidationKind)
			require.NotNil(t, errs)
			assert.EqualValues(t, &ValidationErrors{
				map[string]string{"email": "The field 'email' is not a valid email."},
			}, errs)
		})
		t.Run("UpdateValidationKind", func(t *testing.T) {
			user := &User{FirstName: "John123", LastName: "Doe123"}
			errs := user.Validate(UpdateValidationKind)
			require.NotNil(t, errs)
			assert.EqualValues(t, &ValidationErrors{
				map[string]string{
					"first_name": "The field 'first_name' should contain only apha characters.",
					"last_name":  "The field 'last_name' should contain only apha characters.",
				},
			}, errs)
		})
		t.Run("FilterValidationKind", func(t *testing.T) {
			user := &User{Nickname: deref.StringAddr("-.-")}
			errs := user.Validate(FilterValidationKind)
			require.NotNil(t, errs)
			assert.EqualValues(t, &ValidationErrors{
				map[string]string{"nickname": "The field 'nickname' should contain only apha-numeric characters."},
			}, errs)
		})

	})

	t.Run("empty update and filter validation kinds", func(t *testing.T) {
		user := &User{}
		errs := user.Validate(UpdateValidationKind)
		assert.Nil(t, errs)
		errs = user.Validate(FilterValidationKind)
		assert.Nil(t, errs)
	})
}
