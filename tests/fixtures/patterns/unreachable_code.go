package main

import "fmt"

func example() {
	fmt.Println("Before return")
	return
	fmt.Println("This is unreachable code")
	fmt.Println("This too")
}

func main() {
	example()
}












