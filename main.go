package main

import (
	"fmt"
	"github.com/qbart/krab/krab"
	"github.com/qbart/krab/krab/parser"
)

func main() {
	fmt.Println("Krab v{}", krab.Version)

	parsed, err := parser.ParseFromFile("test/fixtures/migrations/create_table.hcl")
	if err != nil {
		fmt.Print(err)
	}

	fmt.Println(parsed.Ast.Node)
}
