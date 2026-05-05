package ports

import "github.com/go-playground/validator/v10"

type HttpError interface {
	AppError
	GetHttpStatus() int
	GetMetaData(c *HttpContext) interface{}
}

type FieldError struct {
	FieldName    string      `json:"field_name"`
	CurrentValue interface{} `json:"current_value"`
	Errors       string      `json:"errors"`
}

var validate = validator.New()

func ValidateStruct(str interface{}) []FieldError {
	var errors []FieldError
	err := validate.Struct(str)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element FieldError
			element.FieldName = err.StructField()
			element.CurrentValue = err.Value()
			element.Errors = err.Error()
			errors = append(errors, element)
		}
	}
	return errors
}
