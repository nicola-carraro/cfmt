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
	PreprocessorDirective
)

type Token struct {
	Type       TokenType
	Content    string
	HasNewLine bool
	NewLines   int
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
		return fmt.Sprintf("Token{Type: %s, len: %d, NewLines: %d}", t.Type, len(t.Content), t.NewLines)
	} else if t.Type == NoTokenType {
		return fmt.Sprintf("Token{Type: %s}", t.Type)
	} else {
		return fmt.Sprintf("Token{Type: %s, Content: \"%s\"}", t.Type, t.Content)
	}
}


func tryParsePreprocessorDirective(s string) (Token, bool) {
	directives := [...]string{
		"#define",
		"#elif",
		"#else",
		"#endif",
		"#error",
		"#if",
		"#ifdef",
		"#ifndef",
		"#import",
		"#include",
		"#line"}

	for _, directive := range directives {
		if strings.HasPrefix(s, directive) {
			return Token{Type: PreprocessorDirective, Content: directive}, true
		}
	}

	return Token{}, false
}

func peakRune(text string) (rune, int) {
	r, size := utf8.DecodeRuneInString(text)

	if r == utf8.RuneError && size == 1 {
		log.Fatal("Invalid character")
	}

	return r, size
}

func parseSpace(text string) Token {

	tokenSize := 0

	next := text

	r, size := peakRune(next)

	newLines := 0

	for isSpace(r) {
		if r == '\n' {
			newLines++
		}

		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}

	token := Token{Type: Space, Content: text[:tokenSize], NewLines: newLines}
	return token
}

func parseIdentifier(text string) Token {

	tokenSize := 0
	next := text

	r, size := peakRune(next)

	for isIdentifierChar(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}

	token := Token{Type: Identifier, Content: text[:tokenSize]}
	return token
}

// TODO: handle wide strings
func parseString(text string) Token {
	tokenSize := 1
	next := text[1:]

	for true {
		r, size := peakRune(next)
		tokenSize += size
		next = next[size:]
		if r == '"' {
			token := Token{Type: String, Content: text[:tokenSize]}
			return token
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
func parseChar(text string) Token {
	tokenSize := 1
	next := text[1:]

	for true {
		r, size := peakRune(next)
		tokenSize += size
		next = next[size:]
		if r == '\'' {
			token := Token{Type: Char, Content: text[:tokenSize]}
			return token
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

func parseDecimal(text string) Token {
	tokenSize := 0

	next := text

	r, size := peakRune(text)

	for isDigit(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}
	//TODO: handle suffixes

	token := Token{Type: Integer, Content: text[:tokenSize]}
	return token

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

func tryParseFloat(text string) (Token, bool) {

	tokenSize := 0
	next := text

	r, size := peakRune(next)

	if !isDigit(r) && r != '.' {
		return Token{}, false
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
		return Token{}, false
	}

	if isFloatSuffix(r) {
		tokenSize += size
	}

	token := Token{Type: Float, Content: text[:tokenSize]}

	return token, true

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

func parseToken(text string) Token {
	r, _ := peakRune(text)

	token, isFloat := tryParseFloat(text)
	if isFloat {
		return token
	}

	token, isDirective := tryParsePreprocessorDirective(text)
	if isDirective {
		return token
	}
	if isSpace(r) {
		return parseSpace(text)
	}

	if isIdentifierStart(r) {
		return parseIdentifier(text)
	}

	if isDoubleQuote(r) {
		return parseString(text)
	}

	if isSingleQuote(r) {
		return parseChar(text)
	}

	if isFourCharsPunctuation(text) {
		token.Type = Punctuation
		token.Content = text[:4]
		return token
	}

	if isThreeCharsPunctuation(text) {
		token.Type = Punctuation
		token.Content = text[:3]
	} else if isTwoCharsPunctuation(text) {
		token.Type = Punctuation
		token.Content = text[:2]

		return token
	}

	if isOneCharPunctuation(text) {
		token.Type = Punctuation
		token.Content = text[:1]

		return token
	}

	if isDigit(r) {
		//TODO: handle octal and hex
		return parseDecimal(text)
	}

	max := 10
	start := text
	if len(start) > max {
		start = start[:max]
	}
	log.Fatalf("Unrecognised token, starts with %s", start)

	panic("Unreachable")
}

func tokenize(text string) []Token {

	tokens := make([]Token, 0, 100)

	for len(text) > 0 {
		token := parseToken(text)
		text = text[len(token.Content):]
		tokens = append(tokens, token)
	}

	return tokens
}

func skipSpaceAndCountNewLines(tokens []Token) (Token, int) {

	newLines := 0
	for _, t := range tokens {
		if t.Type != Space {
			return t, newLines
		} else {
			newLines = t.NewLines
		}
	}

	return Token{}, newLines
}

func isHash(t Token) bool {
	return t.Type == Punctuation && t.Content == "#"
}

func containsNewLine(t Token) bool {
	return t.Type == Punctuation && t.Content == "#"
}

func hasNewline(t Token) bool {
	return t.Type == Space && t.NewLines > 0
}

func isLeftParenthesis(token Token) bool {
	return token.Type == Punctuation && token.Content == "("
}

func isRightParenthesis(token Token) bool {
	return token.Type == Punctuation && token.Content == ")"
}

func isLeftBrace(token Token) bool {
	return token.Type == Punctuation && token.Content == "{"
}

func isRightBrace(token Token) bool {
	return token.Type == Punctuation && token.Content == "}"
}

func isSemicolon(token Token) bool {
	return token.Type == Punctuation && token.Content == ";"
}

func isPreprocessorDirective(token Token) bool {
	return token.Type == PreprocessorDirective
}

func format(tokens []Token) string {
	newLinesBefore := 0

	newLinesAfter := 0

	prevT := Token{}
	nextT := Token{}

	indent := 0

	isParenthesis := false

	b := strings.Builder{}

	isDirective := false
	for i, t := range tokens {

		if i < len(tokens) {
			nextT, newLinesAfter = skipSpaceAndCountNewLines(tokens[i+1:])
			_ = newLinesAfter
		}

		if t.Type == Space {
			newLinesBefore = t.NewLines
			_ = newLinesBefore
		} else {
			b.WriteString(t.Content)

			if isLeftBrace(t) {
				indent++
			}

			if isLeftParenthesis(t) {
				isParenthesis = true
			}

			if isRightParenthesis(t) {
				isParenthesis = false
			}

			if isRightBrace(nextT) {
				indent--
			}

			if isPreprocessorDirective(t) {
				isDirective = true
			}

			isEndOfStatement := isSemicolon(t) && !isParenthesis

			const newLine = "\r\n"

			const maxNewLines = 2

			endOfDirective := isDirective && newLinesAfter > 0

			if isLeftBrace(t) || isRightBrace(nextT) {
				b.WriteString(newLine)
				for indentLevel := 0; indentLevel < indent; indentLevel++ {
					b.WriteString("  ")
				}

			} else if isRightBrace(t) ||
				endOfDirective ||
				isPreprocessorDirective(nextT) ||
				isEndOfStatement {
				b.WriteString(newLine)

				if newLinesAfter > 1 {
					b.WriteString(newLine)

				}
				for indentLevel := 0; indentLevel < indent; indentLevel++ {
					b.WriteString("  ")
				}

			} else if !isSemicolon(nextT) && !isLeftParenthesis(t) && !isRightParenthesis(nextT) {
				b.WriteString(" ")
			}

			prevT = t

			_ = prevT

			if newLinesAfter > 0 {
				isDirective = false
			}

		}

	}

	return b.String()
}

func main() {
	const path = "test.c"

	data, error := os.ReadFile(path)

	if error != nil {
		log.Fatal("Error reading ", path)
	}

	text := fmt.Sprintf("%s", data)

	tokens := tokenize(text)

	formattedText := format(tokens)

	fmt.Println(formattedText)
}
