package array_slice

import "fmt"

func ReplaceByAppend() {
	// original is an array with 3 elements.
	original := [3]string{"🍔", "🌭", "🥦"}

	// fastfood is a slice into the first two elements of a.
	fastfood := original[0:2]

	// broccoli is a slice into the last element of a.
	broccoli := original[2:3]

	fmt.Println("before", broccoli)      // [🥦]
	fmt.Println("1. original", original) // [🍔 🌭 🥦]

	fastfood = append(fastfood, "🍕")

	fmt.Println("after", broccoli)       // [🍕]
	fmt.Println("2. original", original) // [🍔 🌭 🍕]
}
