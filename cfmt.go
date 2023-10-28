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
	Content []byte
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '\v' || r == '\f'
}

func peakRune(data []byte) (rune, int) {
	r, size := utf8.DecodeRune(data)

	if r == utf8.RuneError && size == 0 {
		log.Fatal("Invalid character")
	}

	return r, size
}

func consumeBytes(data *[]byte, size int) {

	*data = (*data)[size:]
}

func consumeSpace(data *[]byte) int {

	result := 0

	r, size := peakRune(*data)

	for isSpace(r) {
		result += size
		consumeBytes(data, size)
		r, size = peakRune(*data)
	}

	return result
}

func main() {

	const path = "test.c"

	data, error := os.ReadFile(path)

	if error != nil {
		log.Fatal("Error reading ", path)
	}

	tokens := make([]Token, 0, 100)

	for len(data) > 0 {
		r, size := peakRune(data)
		token := Token{}
		start := data
		if isSpace(r) {
			token.Type = Space
			tSize := consumeSpace(&data)

			fmt.Println("Space", " ", tSize)

			token.Content = start[:tSize]

			tokens = append(tokens, token)
		}
		if !isSpace(r) {
			consumeBytes(&data, size)
		}

	}

	for i, token := range tokens {
		fmt.Println(i, " ", len(token.Content))
	}

}
