package array_slice

import "fmt"

func Index() {
	request := map[string]interface{}{"mid": 1, "name": "test"}
	sprinted := fmt.Sprintf("%10s", fmt.Sprintf("%v", request["none"]))
	fmt.Println(len(sprinted), sprinted)
}
