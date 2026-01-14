package main

import "fmt"

// This file contains duplicate function definitions (vibe coding issue)
func hello() {
	fmt.Println("Hello")
}

func hello() {
	fmt.Println("Hello again")
}

func main() {
	hello()
}












