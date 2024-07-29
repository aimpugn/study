package main

import (
	"fmt"
)

func ByAssert(value any) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int, int64:
		return fmt.Sprintf("%v", v), nil
	case bool:
		// return "1" for true and an empty string("") for false
		if v {
			return "1", nil
		} else {
			return "", nil
		}
	}

	return "", nil
}
