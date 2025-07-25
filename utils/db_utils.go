package utils

import (
	"fmt"
	"reflect"
)

func FilterDBUpdates(obj any, userUpdates *map[string]any, ignoreColumns ...string) error {
	// Get the type and value of the struct using reflection
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	// Ensure the input is a struct
	if t.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct, got %s", t.Kind())
	}

	// Iterate through the fields of the struct
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i)   // Get information about the field (e.g., name, type)
		value := v.Field(i) // Get the value of the field

		for _, keyName := range append(ignoreColumns, "id", "created_at", "updated_at") {
			if key.Name == keyName {
				continue // Skip this field if it's in the ignore list
			}
		}

		if value.IsValid() && !value.IsZero() {
			(*userUpdates)[key.Name] = value.Interface()
		}
	}

	return nil
}

func clearValuesForObject(obj any, columns ...string) error {
	v := reflect.ValueOf(obj).Elem()
	for _, fieldName := range columns {
		f := v.FieldByName(fieldName)
		if f.IsValid() && f.CanSet() {
			switch f.Kind() {
			case reflect.String:
				f.SetString("")
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				f.SetInt(0)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				f.SetUint(0)
			case reflect.Float32, reflect.Float64:
				f.SetFloat(0.0)
			case reflect.Bool:
				f.SetBool(false)
			default:
				return fmt.Errorf("unsupported field type: %s", f.Kind())
			}
		}
	}
	return nil
}

func ClearValues(obj any, columns ...string) error {
	if len(columns) == 0 {
		return nil
	}

	if reflect.TypeOf(obj).Kind() != reflect.Slice {
		return clearValuesForObject(&obj, columns...)
	}

	for i := 0; i < reflect.ValueOf(obj).Len(); i++ {
		if err := clearValuesForObject(reflect.ValueOf(obj).Index(i).Addr().Interface(), columns...); err != nil {
			return err
		}
	}

	return nil
}
