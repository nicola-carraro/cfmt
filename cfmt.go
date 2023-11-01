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
	NoTokenType TokenType = iota
	Space
	Identifier
	Integer
	Float
	Char
	String
	Punctuation
)

type Token struct {
	Type       TokenType
	Content    string
	HasNewLine bool
}

type NodeType uint32

const (
	NoNodeType NodeType = iota
	Root
	Braces
	Parenthesis
	Statement
	Preprocessor
)

type Node struct {
	Type     NodeType
	Content  []Token
	Parent   *Node
	Children []Node
}

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

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '\v' || r == '\f'
}

func isDoubleQuote(r rune) bool {
	return r == '"'
}

func isSingleQuote(r rune) bool {
	return r == '\''
}

func (t TokenType) String() string {
	switch t {
	case NoTokenType:
		return "None"
	case Space:
		return "Space"
	case Identifier:
		return "Identifier"
	case Integer:
		return "Integer"
	case Float:
		return "Float"
	case Char:
		return "Char"
	case String:
		return "String"
	case Punctuation:
		return "Punctuation"
	default:
		panic("Invalid TokenType")
	}
}

func (t Token) String() string {
	if t.Type == Space {
		return fmt.Sprintf("Token{Type: %s, len: %d}", t.Type, len(t.Content))
	} else if t.Type == NoTokenType {
		return fmt.Sprintf("Token{Type: %s}", t.Type)
	} else {
		return fmt.Sprintf("Token{Type: %s, Content: \"%s\"}", t.Type, t.Content)
	}
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
			tokenSize += size
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

func isSignStart(r rune) bool {
	return r == '+' || r == '-'
}

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

	hasDot := false

	if r == '.' {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
		hasDot = true
	}

	for isDigit(r) {

		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}

	hasExponent := false

	if isExponentStart(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)

		if isSignStart(r) {
			tokenSize += size
			next = next[size:]
			r, size = peakRune(next)
		}

		for isDigit(r) {
			hasExponent = true
			tokenSize += size
			next = next[size:]
			r, size = peakRune(next)
		}
	}

	if !hasExponent && !hasDot {
		return "", false
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

func tokenize(text string) []Token {

	tokens := make([]Token, 0, 100)

	for len(text) > 0 {
		r, _ := peakRune(text)
		token := Token{}

		content, isFloat := tryParseFloat(text)
		if isFloat {
			token.Type = Float
			token.Content = content
			text, _ = strings.CutPrefix(text, token.Content)
		} else if isSpace(r) {
			token.Type = Space
			token.Content = parseSpace(text)
			text, _ = strings.CutPrefix(text, token.Content)
		} else if isIdentifierStart(r) {
			token.Type = Identifier
			token.Content = parseIdentifier(text)
			text, _ = strings.CutPrefix(text, token.Content)
		} else if isDoubleQuote(r) {
			token.Type = String
			token.Content = parseString(text)
			text, _ = strings.CutPrefix(text, token.Content)
		} else if isSingleQuote(r) {
			token.Type = Char
			token.Content = parseChar(text)
			text, _ = strings.CutPrefix(text, token.Content)
		} else if isFourCharsPunctuation(text) {
			token.Type = Punctuation
			token.Content = text[:4]
			text = text[4:]
		} else if isThreeCharsPunctuation(text) {
			token.Type = Punctuation
			token.Content = text[:3]
			text = text[3:]
		} else if isTwoCharsPunctuation(text) {
			token.Type = Punctuation
			token.Content = text[:2]
			text = text[2:]
		} else if isOneCharPunctuation(text) {
			token.Type = Punctuation
			token.Content = text[:1]
			text = text[1:]
		} else if isDigit(r) {
			//TODO: handle octal and hex
			token.Type = Integer
			token.Content = parseDecimal(text)
			text, _ = strings.CutPrefix(text, token.Content)
		} else {
			max := 10
			start := text
			if len(start) > max {
				start = start[:max]
			}
			log.Fatalf("Unrecognised token, starts with %s", start)
		}

		tokens = append(tokens, token)
	}

	return tokens
}

func skipSpace(tokens []Token, index int) int {

	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type != Space {
			return i
		}
	}

	return len(tokens)
}

func isHashtag(t Token) bool {
	return t.Type == Punctuation && t.Content == "#"
}

func containsNewLine(t Token) bool {
	return t.Type == Punctuation && t.Content == "#"
}

func parse(tokens []Token) Node {
	root := Node{Type: Root, Content: tokens}
	cur := skipSpace(tokens, 0)

	token := tokens[cur]

	if isHashtag(token) {

	}

	return root
}

func main() {
	const path = "test.c"

	data, error := os.ReadFile(path)

	if error != nil {
		log.Fatal("Error reading ", path)
	}

	text := fmt.Sprintf("%s", data)

	tokens := tokenize(text)

	for _, token := range tokens {
		fmt.Println(token)
	}
}
