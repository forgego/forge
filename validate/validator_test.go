package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Slug      string  `validate:"slug"`
	Phone     string  `validate:"phone"`
	Status    string  `validate:"choices=draft published archived"`
	Code      string  `validate:"unique"`
	Price     float64 `validate:"decimal_places=2,decimal_max_digits=10"`
	BigNumber float64 `validate:"decimal_places=2,decimal_max_digits=12"`
}

func TestValidator_CustomValidators(t *testing.T) {
	v := NewValidator()

	// Valid struct
	valid := testStruct{
		Slug:      "hello-world_123",
		Phone:     "+1 (555) 123-4567",
		Status:    "published",
		Code:      "PROD-001",
		Price:     99.99,
		BigNumber: 1234567.89, // large number that would be scientific notation in %g
	}

	err := v.ValidateStruct(&valid)
	assert.NoError(t, err)

	// Invalid slug
	invalidSlug := valid
	invalidSlug.Slug = "invalid slug with spaces!"
	err = v.ValidateStruct(&invalidSlug)
	assert.Error(t, err)

	// Invalid status choice
	invalidStatus := valid
	invalidStatus.Status = "unknown_status"
	err = v.ValidateStruct(&invalidStatus)
	assert.Error(t, err)

	// Invalid decimal places
	invalidPrice := valid
	invalidPrice.Price = 99.999
	err = v.ValidateStruct(&invalidPrice)
	assert.Error(t, err)
}

func TestValidateField_Choices(t *testing.T) {
	v := NewValidator()

	assert.NoError(t, v.ValidateField("active", "choices=active inactive pending"))
	assert.NoError(t, v.ValidateField("pending", "choices=active inactive pending"))
	assert.Error(t, v.ValidateField("deleted", "choices=active inactive pending"))
}

func TestValidateField_UniqueTag(t *testing.T) {
	v := NewValidator()

	// Struct validation must not fail with "unregistered tag" error on unique
	type modelWithUnique struct {
		Email string `validate:"required,unique"`
	}

	m := modelWithUnique{Email: "test@example.com"}
	assert.NoError(t, v.ValidateStruct(&m))
}

func TestValidator_UsesJSONFieldNameInErrors(t *testing.T) {
	type requestModel struct {
		CustomerID int64 `json:"customer_id" validate:"required"`
	}

	err := NewValidator().ValidateStruct(&requestModel{})
	require.Error(t, err)
	var validationErrors *ValidationErrors
	require.ErrorAs(t, err, &validationErrors)
	require.Len(t, validationErrors.Errors, 1)
	assert.Equal(t, "customer_id", validationErrors.Errors[0].Field)
}
