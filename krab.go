package main

import (
	"fmt"

	HclParser "github.com/hashicorp/hcl/hcl/parser"
)

const Version = "0.0.1"

func main() {
	fmt.Println("Krab v{}", Version)

	ast, err := HclParser.Parse([]byte(""))
	if err != nil {
		panic("Failed to parse")
	}
	fmt.Println("test {}", ast)
}
