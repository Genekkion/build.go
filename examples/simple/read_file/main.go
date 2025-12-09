package main

import (
	_ "embed"
	"fmt"
)

//go:embed file.txt
var fileText string

func main() {
	fmt.Println("Printing file contents")
	fmt.Println(fileText)
}
