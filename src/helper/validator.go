package helper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// ValidateRequestBody validates and unmarshals a request body into a struct
func ValidateRequestBody(r *http.Request, dst interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// Unmarshal JSON data
	if err := json.Unmarshal(body, dst); err != nil {
		return err
	}

	// Validate struct binding tags
	return validateStruct(dst)
}

// validateStruct recursively validates binding tags on struct fields
func validateStruct(obj interface{}) error {
	val := reflect.ValueOf(obj)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()

	// Iterate over struct fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Get binding tag
		bindingTag := field.Tag.Get("binding")
		if bindingTag == "" {
			// If this field is a struct or slice of structs, validate recursively
			switch fieldVal.Kind() {
			case reflect.Struct:
				if err := validateStruct(fieldVal.Interface()); err != nil {
					return err
				}
			case reflect.Slice:
				if fieldVal.Type().Elem().Kind() == reflect.Struct {
					for j := 0; j < fieldVal.Len(); j++ {
						if err := validateStruct(fieldVal.Index(j).Interface()); err != nil {
							return err
						}
					}
				}
			}
			continue
		}

		// Parse binding tags
		validations := strings.Split(bindingTag, ",")

		for _, validation := range validations {
			switch {
			case validation == "required":
				if isEmptyValue(fieldVal) {
					return fmt.Errorf("field %s is required", field.Name)
				}
			case strings.HasPrefix(validation, "min="):
				min := strings.TrimPrefix(validation, "min=")
				if fieldVal.Kind() == reflect.Slice && fieldVal.Len() < parseInt(min) {
					return fmt.Errorf("field %s must have at least %s items", field.Name, min)
				}
			}
		}

		// If this field is a struct or slice of structs, validate recursively
		switch fieldVal.Kind() {
		case reflect.Struct:
			if err := validateStruct(fieldVal.Interface()); err != nil {
				return err
			}
		case reflect.Slice:
			if fieldVal.Type().Elem().Kind() == reflect.Struct {
				for j := 0; j < fieldVal.Len(); j++ {
					if err := validateStruct(fieldVal.Index(j).Interface()); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// isEmptyValue checks if a value is empty (zero value)
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// parseInt converts string to int
func parseInt(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}
