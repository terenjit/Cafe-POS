package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{validate: validator.New()}
}

func (v *Validator) Validate(i interface{}) map[string]string {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	errs := make(map[string]string)
	for _, fe := range err.(validator.ValidationErrors) {
		key := jsonKey(t, fe.StructField())
		errs[key] = errorMessage(fe)
	}
	return errs
}

func jsonKey(t reflect.Type, structField string) string {
	f, ok := t.FieldByName(structField)
	if !ok {
		return strings.ToLower(structField)
	}
	tag := f.Tag.Get("json")
	name := strings.Split(tag, ",")[0]
	if name == "" || name == "-" {
		return strings.ToLower(structField)
	}
	return name
}

func errorMessage(fe validator.FieldError) string {
	param := fe.Param()
	switch fe.Tag() {
	case "required":
		return "is required"
	case "min":
		if isString(fe.Kind()) {
			return fmt.Sprintf("minimum %s characters", param)
		}
		return fmt.Sprintf("minimum %s", param)
	case "max":
		if isString(fe.Kind()) {
			return fmt.Sprintf("maximum %s characters", param)
		}
		return fmt.Sprintf("maximum %s", param)
	case "email":
		return "invalid email format"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", param)
	case "uuid4":
		return "invalid UUID format"
	case "len":
		return fmt.Sprintf("must be exactly %s characters", param)
	case "numeric":
		return "must be a number"
	default:
		return "invalid"
	}
}

func isString(k reflect.Kind) bool {
	return k == reflect.String
}
