package main

import (
	"fmt"
	"log"
	"os"
	"strings"
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
	Punctuation
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
	Content string
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

func peakRune(text string) (rune, int) {
	r, size := utf8.DecodeRuneInString(text)

	if r == utf8.RuneError && size == 1 {
		log.Fatal("Invalid character")
	}

	return r, size
}

func _consumeBytes(data *[]byte, size int) {

	*data = (*data)[size:]
}

func parseSpace(text string) string {

	tokenSize := 0

	next := text

	r, size := peakRune(next)

	for isSpace(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}

	return text[:tokenSize]
}

func parseIdentifier(text string) string {

	tokenSize := 0
	next := text

	r, size := peakRune(next)

	for isIdentifierChar(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}

	return text[:tokenSize]
}

// TODO: handle wide strings
func parseString(text string) string {
	tokenSize := 1
	next := text[1:]

	for true {
		r, size := peakRune(next)
		tokenSize += size
		next = next[size:]
		if r == '"' {
			return text[:tokenSize]
		} else if r == '\\' {
			_, size := peakRune(next)
			next = next[size:]
		} else if r == utf8.RuneError && size == 0 {
			log.Fatal("Unclosed string literal")
		}
	}

	panic("unreachable")
}

// TODO: handle wide chars
func parseChar(text string) string {
	tokenSize := 1
	next := text[1:]

	for true {
		r, size := peakRune(next)
		tokenSize += size
		next = next[size:]
		if r == '\'' {
			return text[:tokenSize]
		} else if r == '\\' {
			_, size := peakRune(next)
			next = next[size:]
			tokenSize += size

		} else if r == utf8.RuneError && size == 0 {
			log.Fatal("Unclosed character literal")
		}
	}

	panic("unreachable")
}

func parseDecimal(text string) string {
	tokenSize := 0

	next := text

	r, size := peakRune(text)

	for isDigit(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}
	//TODO: handle suffixes

	return text[:tokenSize]

}

func isExponentStart(r rune) bool {
	return r == 'e' || r == 'E'
}

func isFloatSuffix(r rune) bool {
	return r == 'f' || r == 'l' || r == 'F' || r == 'L'
}

// Todo: handle floats of the form 123e+1
func tryParseFloat(text string) (string, bool) {

	tokenSize := 0
	next := text

	r, size := peakRune(next)

	if !isDigit(r) && r != '.' {
		return "", false
	}

	for isDigit(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}

	if r != '.' {
		return "", false
	}

	tokenSize += size
	next = next[size:]
	r, size = peakRune(next)

	for isDigit(r) {

		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}

	if isFloatSuffix(r) {
		tokenSize += size
	}

	return text[:tokenSize], true

}

func isFourCharsPunctuation(text string) bool {
	return strings.HasPrefix(text, "%:%:")
}

func isThreeCharsPunctuation(text string) bool {
	threeChars := [...]string{"<<=", ">>=", "..."}

	for _, o := range threeChars {
		if strings.HasPrefix(text, o) {
			return true

		}
	}

	return false
}

func isTwoCharsPunctuation(text string) bool {
	twoChars := [...]string{"++", "--", "<<", ">>", "<=", ">=", "==",
		"!=", "&&", "||", "->", "*=", "/=", "%=", "+=", "-=", "&=", "^=", "|=", "##",
		"<:", ":>", "<%", "%>", "%:"}

	for _, o := range twoChars {
		if strings.HasPrefix(text, o) {
			return true

		}
	}

	return false
}

func isOneCharPunctuation(text string) bool {
	oneChar := [...]string{"~", "*", "/", "%", "+", "-", "<", ">",
		"&", "|", "^", ",", "=", "[", "]", "(", ")", "{",
		"}", ".", "!", "?", ":", ";", "#"}

	for _, o := range oneChar {
		if strings.HasPrefix(text, o) {
			return true

		}
	}

	return false
}

func main() {

	const path = "test.c"

	data, error := os.ReadFile(path)

	if error != nil {
		log.Fatal("Error reading ", path)
	}

	text := fmt.Sprintf("%s", data)

	tokens := make([]Token, 0, 100)

	for len(text) > 0 {
		r, size := peakRune(text)
		token := Token{}
		if isSpace(r) {
			token.Type = Space
			token.Content = parseSpace(text)
			text, _ = strings.CutPrefix(text, token.Content)
			tokens = append(tokens, token)
		} else if isIdentifierStart(r) {
			token.Type = Identifier
			token.Content = parseIdentifier(text)
			text, _ = strings.CutPrefix(text, token.Content)
			tokens = append(tokens, token)
		} else if isDoubleQuote(r) {
			token.Type = String
			token.Content = parseString(text)
			text, _ = strings.CutPrefix(text, token.Content)
			tokens = append(tokens, token)
		} else if isSingleQuote(r) {
			token.Type = Char
			token.Content = parseChar(text)
			text, _ = strings.CutPrefix(text, token.Content)
			tokens = append(tokens, token)
		} else if isFourCharsPunctuation(text) {
			token.Type = Punctuation
			token.Content = text[:4]
			text = text[4:]
			tokens = append(tokens, token)
		} else if isThreeCharsPunctuation(text) {
			token.Type = Punctuation
			token.Content = text[:3]
			text = text[3:]
			tokens = append(tokens, token)
		} else if isTwoCharsPunctuation(text) {
			token.Type = Punctuation
			token.Content = text[:2]
			text = text[2:]
			tokens = append(tokens, token)
		} else if isOneCharPunctuation(text) {
			token.Type = Punctuation
			token.Content = text[:1]
			text = text[1:]
			tokens = append(tokens, token)
		} else {
			content, isFloat := tryParseFloat(text)

			if isFloat {
				token.Type = Float
				token.Content = content
				text, _ = strings.CutPrefix(text, token.Content)
				tokens = append(tokens, token)
			} else if isNonzeroDigit(r) {
				//TODO: handle octal and hex
				token.Type = Integer
				//TODO: handle 0
				token.Content = parseDecimal(text)
				text, _ = strings.CutPrefix(text, token.Content)
				tokens = append(tokens, token)
			} else {
				text = text[size:]
			}
		}

	}

	for i, token := range tokens {
		if token.Type == Space {
			fmt.Println(i, " ", "Space", " ", len(token.Content))
		} else if token.Type == Identifier {
			fmt.Println(i, " ", "Identifier", " ", token.Content)
		} else if token.Type == String {
			fmt.Println(i, " ", "String", " ", token.Content)
		} else if token.Type == Char {
			fmt.Println(i, " ", "Char", " ", token.Content)
		} else if token.Type == Integer {
			fmt.Println(i, " ", "Integer", " ", token.Content)
		} else if token.Type == Float {
			fmt.Println(i, " ", "Float", " ", token.Content)
		} else if token.Type == Punctuation {
			fmt.Println(i, " ", "Punctuation", " ", token.Content)
		}
	}

}
