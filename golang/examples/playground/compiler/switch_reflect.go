package main

import (
	"fmt"
	"reflect"
)

func ByReflect(value any) (string, error) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Bool:
		// convert boolean to "1" or "" for true and false
		if v.Bool() {
			return "1", nil
		} else {
			return "", nil
		}
	case reflect.Int, reflect.Int64:
		return fmt.Sprintf("%d", v.Int()), nil
	case reflect.Array, reflect.Slice:
		return "Array", nil
	}

	return "", nil
}
