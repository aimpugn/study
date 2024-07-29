package switchcase

import (
	"fmt"
	"reflect"
)

func ByReflection() {
	var val interface{} = complex128(100)

	switch reflect.ValueOf(val).Kind() {
	case reflect.String:
		fmt.Println("string type")
	case reflect.Int:
		fmt.Println("int type")
	case reflect.Int16:
		fmt.Println("int16 type")
	case reflect.Int32:
		fmt.Println("int32 type")
	case reflect.Int64:
		fmt.Println("int64 type")
	case reflect.Uint:
		fmt.Println("uint type")
	case reflect.Uint8:
		fmt.Println("uint8 type")
	case reflect.Uint16:
		fmt.Println("uint16 type")
	case reflect.Uint32:
		fmt.Println("uint32 type")
	case reflect.Uint64:
		fmt.Println("uint64 type")
	case reflect.Float32:
		fmt.Println("float32 type")
	case reflect.Float64:
		fmt.Println("float64 type")
	case reflect.Complex64:
		fmt.Println("complex64 type")
	case reflect.Complex128:
		fmt.Println("complex128 type")
	case reflect.Bool:
		fmt.Println("bool type")
	default:
		fmt.Println("unknown type")
	}
}
