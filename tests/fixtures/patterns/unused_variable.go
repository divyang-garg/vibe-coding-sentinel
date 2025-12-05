package main

import "fmt"

func main() {
	var unusedVar string = "not used"
	var usedVar string = "used"
	
	fmt.Println(usedVar)
	// unusedVar is declared but never used
}

