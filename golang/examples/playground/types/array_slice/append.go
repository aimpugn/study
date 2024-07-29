package array_slice

import "fmt"

func Append() {
	fruits := []string{"🍋", "🍎", "🍒"}

	food := append(fruits, "🍔", "🌭", "🍕") // `food` is new slice

	fmt.Println("fruits remains untouched", fruits)
	fmt.Println("food", food)
}
