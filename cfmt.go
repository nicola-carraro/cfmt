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
	Integer
	Float
	Char
	String
)

var keywords = [...]string{
	"auto",
	"break",
	"case",
	"char",
	"const",
	"continue",
	"default",
	"do",
	"enum",
	"extern",
	"float",
	"for",
	"goto",
	"if",
	"int",
	"long",
	"register",
	"return",
	"short",
	"signed",
	"sizeof",
	"static",
	"struct",
	"switch",
	"typedef",
	"union",
	"unsigned",
	"void",
	"volatile",
	"while"}

func isIdentifierStart(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r == '_')
}

func isIdentifierChar(r rune) bool {
	return isIdentifierStart(r) || isDigit(r)
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isNonzeroDigit(r rune) bool {
	return r >= '1' && r <= '9'
}

type Token struct {
	Type    TokenType
	Content []byte
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '\v' || r == '\f'
}

func isDoubleQuote(r rune) bool {
	return r == '"'
}

func isSingleQuote(r rune) bool {
	return r == '\''
}

func peakRune(data []byte) (rune, int) {
	r, size := utf8.DecodeRune(data)

	if r == utf8.RuneError && size == 1 {
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

func consumeIdentifier(data *[]byte) int {

	result := 0

	r, size := peakRune(*data)

	for isIdentifierChar(r) {
		result += size
		consumeBytes(data, size)
		r, size = peakRune(*data)
	}

	return result
}

// TODO: handle wide strings
func consumeString(data *[]byte) int {
	result := 1
	consumeBytes(data, 1)

	for true {
		r, size := peakRune(*data)
		result += size
		consumeBytes(data, size)
		if r == '"' {
			return result
		} else if r == '\\' {
			_, size := peakRune(*data)
			consumeBytes(data, size)
		} else if r == utf8.RuneError && size == 0 {
			log.Fatal("Unclosed string literal")
		}
	}

	panic("unreachable")
}

// TODO: handle wide chars
func consumeChar(data *[]byte) int {
	result := 1
	consumeBytes(data, 1)

	for true {
		r, size := peakRune(*data)
		result += size
		consumeBytes(data, size)
		if r == '\'' {
			return result
		} else if r == '\\' {
			_, size := peakRune(*data)
			consumeBytes(data, size)
		} else if r == utf8.RuneError && size == 0 {
			log.Fatal("Unclosed character literal")
		}
	}

	panic("unreachable")
}

func consumeDecimalInteger(data *[]byte) int {
	result := 0

	r, size := peakRune(*data)

	for isDigit(r) {
		result += size
		consumeBytes(data, size)
		r, size = peakRune(*data)
	}
	//TODO: handle suffixes

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
			token.Content = start[:tSize]

			tokens = append(tokens, token)
		} else if isIdentifierStart(r) {
			token.Type = Identifier
			tSize := consumeIdentifier(&data)
			token.Content = start[:tSize]
			tokens = append(tokens, token)
		} else if isDoubleQuote(r) {
			token.Type = String
			tSize := consumeString(&data)
			token.Content = start[:tSize]
			tokens = append(tokens, token)
		} else if isSingleQuote(r) {
			token.Type = Char
			tSize := consumeChar(&data)
			token.Content = start[:tSize]
			tokens = append(tokens, token)
		} else if isNonzeroDigit(r) {
			//TODO: handle octal and hex
			token.Type = Integer
			tSize := consumeDecimalInteger(&data)
			token.Content = start[:tSize]
			tokens = append(tokens, token)
		} else {
			consumeBytes(&data, size)
		}

	}

	for i, token := range tokens {
		if token.Type == Space {
			fmt.Println(i, " ", "Space", " ", len(token.Content))
		} else if token.Type == Identifier {
			fmt.Println(i, " ", "Identifier", " ")
			fmt.Printf("%s\n", token.Content)
		} else if token.Type == String {
			fmt.Print(i, " ", "String", " ")
			fmt.Printf("%s\n", token.Content)
		} else if token.Type == Char {
			fmt.Print(i, " ", "Char", " ")
			fmt.Printf("%s\n", token.Content)
		} else if token.Type == Integer {
			fmt.Print(i, " ", "Integer", " ")
			fmt.Printf("%s\n", token.Content)
		}
	}

}
