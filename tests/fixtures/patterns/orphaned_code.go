package main

import "fmt"

// This is valid at top level
var globalVar = "valid"

// This is orphaned code (statement outside function)
fmt.Println("This is orphaned - should be inside a function")

func main() {
	fmt.Println("Hello")
}

