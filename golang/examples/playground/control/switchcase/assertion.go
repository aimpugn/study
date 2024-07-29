package switchcase

import "fmt"

func ByAssertion() {
	var val interface{} = complex128(100)

	switch val.(type) {
	case string:
		fmt.Println("string type")
	case int:
		fmt.Println("int type")
	case int16:
		fmt.Println("int16 type")
	case int32:
		fmt.Println("int32 type")
	case int64:
		fmt.Println("int64 type")
	case uint:
		fmt.Println("uint type")
	case uint8:
		fmt.Println("uint8 type")
	case uint16:
		fmt.Println("uint16 type")
	case uint32:
		fmt.Println("uint32 type")
	case uint64:
		fmt.Println("uint64 type")
	case float32:
		fmt.Println("float32 type")
	case float64:
		fmt.Println("float64 type")
	case complex64:
		fmt.Println("complex64 type")
	case complex128:
		fmt.Println("complex128 type")
	case bool:
		fmt.Println("bool type")
	default:
		fmt.Println("unknown type")
	}
}
