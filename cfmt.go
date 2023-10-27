package main

import (
	"fmt"
	"log"
	"os"
	"unicode/utf8"
)

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

func consumeRune(data *[]byte) rune {
	r, size := utf8.DecodeRune(*data)

	if r == utf8.RuneError {
		log.Fatal("Invalid character")
	}

	*data = (*data)[size:]

	return r
}

func main() {

	const path = "test.c"

	data, error := os.ReadFile(path)

	if error != nil {
		log.Fatal("Error reading ", path)
	}

	//tokens := make([]Token, 100)

	index := 0
	for len(data) > 0 {
		r := consumeRune(&data)

		fmt.Print(index, " ")

		if isSpace(r) {
			fmt.Println("Space")
		} else {
			fmt.Println("Not space")
		}

		index++
	}

}
