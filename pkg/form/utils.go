package form

import (
	"reflect"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func CollectErrorsToForm(trans ut.Translator, form interface{}, errs [](validator.FieldError)) {
	formStruct := reflect.ValueOf(form)
	if formStruct.Kind() == reflect.Ptr {
		formStruct = formStruct.Elem()
	}

	for _, err := range errs {
		fieldName := err.StructField()
		field := formStruct.FieldByName(fieldName)

		if !field.IsValid() ||
			field.Kind() != reflect.Ptr ||
			field.Type().Elem().Name() != "FormItemContext" ||
			field.IsNil() {
			continue
		}

		formCtx := field.Elem()
		setFormField(formCtx, "Invalid", true)
		setFormField(formCtx, "ErrorMessage", err.Translate(trans))
	}
}

func CopyFormValueFromPayload(form interface{}, payload interface{}) {
	formStruct := reflect.ValueOf(form)
	if formStruct.Kind() == reflect.Ptr {
		formStruct = formStruct.Elem()
	}

	payloadStruct := reflect.ValueOf(payload)
	if payloadStruct.Kind() == reflect.Ptr {
		payloadStruct = payloadStruct.Elem()
	}

	for i := 0; i < formStruct.NumField(); i++ {
		formField := formStruct.Field(i)
		fieldName := formStruct.Type().Field(i).Name

		if !formField.IsValid() ||
			formField.Kind() != reflect.Ptr ||
			formField.Type().Elem().Name() != "FormItemContext" ||
			formField.IsNil() {
			continue
		}

		if payload := payloadStruct.FieldByName(fieldName); payload.IsValid() {
			setFormField(formField.Elem(), "Value", payload.Interface())
		}
	}
}

func setFormField(form reflect.Value, fieldName string, value interface{}) {
	field := form.FieldByName(fieldName)
	if field.IsValid() && field.CanSet() {
		field.Set(reflect.ValueOf(value))
	}
}
