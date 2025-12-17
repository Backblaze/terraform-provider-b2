//####################################################################
//
// File: b2/convert.go
//
// Copyright 2025 Backblaze Inc. All Rights Reserved.
//
// License https://www.backblaze.com/using_b2_code.html
//
//####################################################################

package b2

import (
	"reflect"
	"strings"
	"unicode"
)

func convertStructToMap(input interface{}) map[string]interface{} {
	return convertReflectValue(reflect.ValueOf(input), false).(map[string]interface{})
}

func convertReflectValue(v reflect.Value, wrapPointers bool) interface{} {
	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return []interface{}{}
		}
		elem := v.Elem()
		if elem.Kind() == reflect.Struct && wrapPointers {
			// Wrap pointer to struct in array for Terraform (only when inside a struct field)
			return []interface{}{convertReflectValue(elem, true)}
		}
		v = elem
	}

	switch v.Kind() {
	case reflect.Struct:
		m := make(map[string]interface{})
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")

			// Handle anonymous embedded structs by flattening their fields
			if field.Anonymous && jsonTag == "" {
				embeddedMap := convertReflectValue(v.Field(i), true)
				if em, ok := embeddedMap.(map[string]interface{}); ok {
					// Merge embedded struct fields into parent map
					for k, v := range em {
						m[k] = v
					}
				}
				continue
			}

			if jsonTag != "" && jsonTag != "-" {
				// Extract field name from JSON tag (before comma for options like omitempty)
				fieldName := jsonTag
				if commaIdx := strings.Index(jsonTag, ","); commaIdx != -1 {
					fieldName = jsonTag[:commaIdx]
				}
				// Convert camelCase to snake_case
				snakeKey := convertCamelToSnake(fieldName)
				// When processing struct fields, wrap pointers in arrays
				m[snakeKey] = convertReflectValue(v.Field(i), true)
			}
		}
		return m
	case reflect.Slice:
		s := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			s[i] = convertReflectValue(v.Index(i), true)
		}
		return s
	default:
		return v.Interface()
	}
}

func convertCamelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}
