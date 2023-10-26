package main

import "fmt"
import "os"
import "unicode/utf8"

type TokenType uint32

const (
	None TokenType = iota
	Space
	Identifier
	IntConst
	CharConst
	FloatConst
)

type Token struct {
	Type    TokenType
	Content string
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '\v' || r == '\f'
}

func main() {

	const path = "test.c"

	data, error := os.ReadFile(path)

	if error != nil {
		fmt.Println("Error reading ", path)
		os.Exit(1)
	}

	//tokens := make([]Token, 100)

	firstRune, _ := utf8.DecodeRune(data)

	if isSpace(firstRune) {
		fmt.Println("Space")
	} else {
		fmt.Println("Not space")
	}

}
