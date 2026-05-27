package main

import "fmt"

// Greet returns a greeting for the given name.
func Greet(name string) string {
	if name == "" {
		return "Hello, World!"
	}
	return fmt.Sprintf("Hello, %s!", name)
}

// IsValid checks if an age is within valid range.
func IsValid(age int) bool {
	return age >= 0 && age <= 150
}
