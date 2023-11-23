package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"unicode/utf8"
)

type TokenType uint32

const (
	NoTokenType TokenType = iota
	Space
	Keyword
	Identifier
	Integer
	Float
	Char
	String
	Punctuation
	Directive
)

type Token struct {
	Type       TokenType
	Content    string
	HasNewLine bool
	NewLines   int
}

type StructUnionEnum struct {
	indent int
}

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
	case Identifier:
		return "Identifier"
	case Keyword:
		return "Keyword"
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
	case Directive:
		return "Directive"
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

func tryParseKeyword(s string) (Token, bool) {
	keywords := [...]string{
		"auto",
		"break",
		"case",
		"char",
		"const",
		"continue",
		"default",
		"double",
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

	for _, keyword := range keywords {
		if strings.HasPrefix(s, keyword) {
			return Token{Type: Keyword, Content: keyword}, true
		}
	}

	return Token{}, false
}

func tryParseDirective(s string) (Token, bool) {
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
			return Token{Type: Directive, Content: directive}, true
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

	hasDigit := false

	r, size := peakRune(next)

	if !isDigit(r) && r != '.' {
		return Token{}, false
	}

	for isDigit(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
		hasDigit = true
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
		hasDigit = true
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

	if !hasDigit {
		return Token{}, false
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

	if len(text) == 0 {
		return Token{}
	}

	r, _ := peakRune(text)

	token, isFloat := tryParseFloat(text)
	if isFloat {
		return token
	}

	token, directive := tryParseDirective(text)
	if directive {
		return token
	}

	token, keyword := tryParseKeyword(text)
	if keyword {
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

func skipSpaceAndCountNewLines(text string) (string, int) {
	newLines := 0

	for r, size := peakRune(text); isSpace(r); r, size = peakRune(text) {
		if r == '\n' {
			newLines++
		}
		text = text[size:]
	}

	return text, newLines
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

func isDirective(token Token) bool {
	return token.Type == Directive
}

func isIdentifier(token Token) bool {
	return token.Type == Identifier
}

func canBeLeftOperand(token Token) bool {
	return token.Type == Identifier ||
		token.Type == String ||
		token.Type == Integer ||
		token.Type == Float ||
		token.Type == Char ||
		(token.Type == Punctuation && token.Content == ")")
}

func canBePointerOperator(token Token) bool {
	return token.Type == Punctuation && (token.Content == "&" || token.Content == "*")
}

func isIncrDecrOperator(token Token) bool {
	return token.Type == Punctuation && (token.Content == "++" || token.Content == "--")
}

func isDotOperator(token Token) bool {
	return token.Type == Punctuation && (token.Content == ".")
}

func isArrowOperator(token Token) bool {
	return token.Type == Punctuation && (token.Content == "->")
}

func isKeyword(token Token) bool {
	return token.Type == Keyword
}

func isStructUnionEnumKeyword(token Token) bool {
	return token.Type == Keyword && (token.Content == "struct" || token.Content == "union" || token.Content == "enum")

}

func isAssignement(token Token) bool {
	assignmentOps := []string{"=", "*=", "/=", "%=", "+=", "-=", "<<=", ">>=", "&=", "^=", "|="}

	return token.Type == Punctuation && slices.Contains(assignmentOps, token.Content)
}

func format(text string) string {

	newLinesAfter := 0
	indent := 0

	isParenthesis := false

	b := strings.Builder{}

	directive := false

	structUnionEnums := make([]StructUnionEnum, 0)

	text, _ = skipSpaceAndCountNewLines(text)

	prevT := Token{}
	t := parseToken(text)
	text = text[len(t.Content):]

	for t.Type != NoTokenType {

		//fmt.Println(t)

		if isStructUnionEnumKeyword(t) {
			structUnionEnums = append(structUnionEnums, StructUnionEnum{indent})
		}

		text, newLinesAfter = skipSpaceAndCountNewLines(text)

		nextT := parseToken(text)

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

		endOfStructUnionEnumBody := false

		if isRightBrace(t) && len(structUnionEnums) > 0 && (structUnionEnums[len(structUnionEnums)-1]).indent == indent {
			structUnionEnums = structUnionEnums[:len(structUnionEnums)-1]
			endOfStructUnionEnumBody = true
		}

		if isDirective(t) {
			directive = true
		}

		isEndOfStatement := isSemicolon(t) && !isParenthesis

		const newLine = "\r\n"

		const maxNewLines = 2

		endOfDirective := directive && newLinesAfter > 0

		isFunctionName := t.Type == Identifier && isLeftParenthesis(nextT)

		isPointerOperator := canBePointerOperator(t) && !canBeLeftOperand(prevT)

		hasPostfixIncrDecr := isIncrDecrOperator(nextT) && (isIdentifier(t))

		isBlockStart := isLeftBrace(t) && !isAssignement(prevT)

		indentation := "    "

		if isBlockStart || (isSemicolon(t) && isRightBrace(nextT)) {
			b.WriteString(newLine)
			for indentLevel := 0; indentLevel < indent; indentLevel++ {
				b.WriteString(indentation)
			}

		} else if (isRightBrace(t) && (!endOfStructUnionEnumBody && !isSemicolon(nextT))) ||
			endOfDirective ||
			isDirective(nextT) ||
			isEndOfStatement ||
			(isSemicolon(t) && !isParenthesis) {
			b.WriteString(newLine)

			if newLinesAfter > 1 {
				b.WriteString(newLine)

			}
			for indentLevel := 0; indentLevel < indent; indentLevel++ {
				b.WriteString(indentation)
			}

		} else if !isSemicolon(nextT) &&
			!isRightBrace(nextT) &&
			!isLeftParenthesis(t) &&
			!isRightParenthesis(nextT) &&
			!isPointerOperator &&
			!isFunctionName &&
			!hasPostfixIncrDecr &&
			!isIncrDecrOperator(t) &&
			!isDotOperator(t) &&
			!isDotOperator(nextT) &&
			!isArrowOperator(t) &&
			!isArrowOperator(nextT) &&
			!isLeftBrace(t) {
			b.WriteString(" ")
		}

		if newLinesAfter > 0 {
			directive = false
		}

		text = text[len(nextT.Content):]

		prevT = t

		t = nextT
	}

	return b.String()
}

func main() {
	const path = "test.c"
	//const path = "scratch.c"

	data, error := os.ReadFile(path)

	if error != nil {
		log.Fatal("Error reading ", path)
	}

	text := fmt.Sprintf("%s", data)

	formattedText := format(text)

	fmt.Println(formattedText)
}
