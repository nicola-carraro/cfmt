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
	Type    TokenType
	Content string
}

type StructUnionEnum struct {
	Indent int
}

type Parser struct {
	PreviousToken Token
	Token         Token
	NextToken     Token
	Indent        int
	InputLine     int
	InputColumn   int
	OutputLine    int
	OutputColumn  int
	Input         string
	Output        strings.Builder
	NewLinesAfter int
	IsParenthesis bool
}

const indentation = "    "

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
		return fmt.Sprintf("Token{Type: %s, len: %d}", t.Type, len(t.Content))
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

func (parser *Parser) parseToken() bool {

	skipSpaceAndCountNewLines(parser)

	if isAbsent(parser.Token) {
		parser.Token = parseToken(parser.Input)
		parser.Input = parser.Input[len(parser.Token.Content):]
		skipSpaceAndCountNewLines(parser)
	} else {
		parser.PreviousToken = parser.Token
		parser.Token = parser.NextToken
	}

	parser.NextToken = parseToken(parser.Input)
	parser.Input = parser.Input[len(parser.NextToken.Content):]

	//fmt.Printf("updateParser, PreviousToken:%s, Token:%s, NextToken:%s\n", parser.PreviousToken, parser.Token, parser.NextToken)

	if isLeftParenthesis(parser.Token) {
		parser.IsParenthesis = true
	}

	if isRightParenthesis(parser.Token) {
		parser.IsParenthesis = false
	}

	return !isAbsent(parser.Token)
}

func parseToken(input string) Token {

	//fmt.Printf("parseToken, current token %s\n", parser.Token)

	if len(input) == 0 {
		return Token{}
	}

	r, _ := peakRune(input)

	token, isFloat := tryParseFloat(input)
	if isFloat {
		return token
	}

	token, directive := tryParseDirective(input)
	if directive {
		return token
	}

	token, keyword := tryParseKeyword(input)
	if keyword {
		return token
	}

	if isIdentifierStart(r) {
		return parseIdentifier(input)
	}

	if isDoubleQuote(r) {
		return parseString(input)
	}

	if isSingleQuote(r) {
		return parseChar(input)
	}

	if isFourCharsPunctuation(input) {
		token.Type = Punctuation
		token.Content = input[:4]
		return token
	}

	if isThreeCharsPunctuation(input) {
		token.Type = Punctuation
		token.Content = input[:3]
	} else if isTwoCharsPunctuation(input) {
		token.Type = Punctuation
		token.Content = input[:2]

		return token
	}

	if isOneCharPunctuation(input) {
		token.Type = Punctuation
		token.Content = input[:1]

		return token
	}

	if isDigit(r) {
		//TODO: handle octal and hex
		return parseDecimal(input)
	}

	max := 10
	start := input
	if len(start) > max {
		start = start[:max]
	}
	log.Fatalf("Unrecognised token, starts with %s", start)

	panic("Unreachable")
}

func skipSpaceAndCountNewLines(parser *Parser) {

	parser.NewLinesAfter = 0

	for r, size := peakRune(parser.Input); isSpace(r); r, size = peakRune(parser.Input) {
		if r == '\n' {
			parser.NewLinesAfter++
		}
		parser.Input = parser.Input[size:]
	}

}

func isHash(t Token) bool {
	return t.Type == Punctuation && t.Content == "#"
}

func containsNewLine(t Token) bool {
	return t.Type == Punctuation && t.Content == "#"
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

func isComma(token Token) bool {

	return token.Type == Punctuation && token.Content == ","
}

func newParser(input string) *Parser {
	parser := Parser{Input: input, Output: strings.Builder{}}

	return &parser
}

func isAbsent(token Token) bool {
	return token.Type == NoTokenType
}

func (parser *Parser) writeNewLines(lines int) {
	const newLine = "\r\n"

	for line := 0; line < lines; line++ {
		parser.Output.WriteString(newLine)
	}

	for indentLevel := 0; indentLevel < parser.Indent; indentLevel++ {
		parser.Output.WriteString(indentation)
	}
}

func indent(parser *Parser) {

}

func formatInitialiserList(parser *Parser) {
	openBraces := 1

	//fmt.Println(parser.Input)

	for parser.parseToken() {

		parser.Output.WriteString(parser.Token.Content)

		if isRightBrace(parser.Token) {
			openBraces--
		}

		if isLeftBrace(parser.Token) {

			formatInitialiserList(parser)
			continue
		}

		if !neverWhiteSpace(parser) &&
			!isRightBrace(parser.NextToken) &&
			!isRightBrace(parser.Token) {
			parser.Output.WriteString(" ")
		}

		if openBraces == 0 {
			return
		}
	}

	log.Fatal("Unclosed initialiser list")
}

func isPointerOperator(parser *Parser) bool {
	return canBePointerOperator(parser.Token) && !canBeLeftOperand(parser.PreviousToken)
}

func hasPostfixIncrDecr(parser *Parser) bool {
	return isIncrDecrOperator(parser.NextToken) && (isIdentifier(parser.Token))
}

func formatBlockBody(parser *Parser) {

	parser.Indent++

	parser.writeNewLines(1)

	structUnionOrEnum := false

	//fmt.Printf("start %s\n", parser.NextToken)

	for parser.parseToken() {

		if isStructUnionEnumKeyword(parser.Token) {
			structUnionOrEnum = true
		}

		parser.Output.WriteString(parser.Token.Content)

		if isRightBrace(parser.NextToken) {
			parser.Indent--
		}

		if isLeftBrace(parser.Token) {

			if isAssignement(parser.PreviousToken) {
				formatInitialiserList(parser)
			} else if structUnionOrEnum {
				formatDeclarationBody(parser)
				structUnionOrEnum = false
			} else {
				formatBlockBody(parser)
			}
		} else if isSemicolon(parser.Token) && !parser.IsParenthesis {

			if parser.NewLinesAfter > 1 {
				parser.writeNewLines(2)

			} else {
				parser.writeNewLines(1)
			}
		} else if isRightBrace(parser.Token) {

			//fmt.Printf("end %s\n", parser.PreviousToken)
			if parser.NewLinesAfter > 1 {
				parser.writeNewLines(2)

			} else {
				parser.writeNewLines(1)
			}
			return
		} else if !neverWhiteSpace(parser) &&
			!isDotOperator(parser.NextToken) {
			parser.Output.WriteString(" ")
		}
	}

	log.Fatal("Unclosed block")
}

func formatDeclarationBody(parser *Parser) {

	//fmt.Printf("DECLARATION: %s\n", parser.Input)
	parser.Indent++

	parser.writeNewLines(1)

	for parser.parseToken() {

		parser.Output.WriteString(parser.Token.Content)

		if isRightBrace(parser.NextToken) {
			parser.Indent--
		}

		if isLeftBrace(parser.Token) {
			formatDeclarationBody(parser)
		} else if isSemicolon(parser.Token) {
			parser.writeNewLines(1)
		} else if !neverWhiteSpace(parser) &&
			!isDotOperator(parser.NextToken) &&
			!isSemicolon(parser.NextToken) {
			parser.Output.WriteString(" ")
		}

		if isRightBrace(parser.Token) {
			return
		}
	}

	log.Fatal("Unclosed declaration braces")
}

func isFunctionName(parser *Parser) bool {
	return parser.Token.Type == Identifier && isLeftParenthesis(parser.NextToken)
}

func neverWhiteSpace(parser *Parser) bool {

	return isSemicolon(parser.NextToken) ||
		isLeftParenthesis(parser.Token) ||
		isRightParenthesis(parser.NextToken) ||
		isPointerOperator(parser) ||
		isFunctionName(parser) ||
		hasPostfixIncrDecr(parser) ||
		isIncrDecrOperator(parser.Token) ||
		isDotOperator(parser.Token) ||
		isArrowOperator(parser.Token) ||
		isArrowOperator(parser.NextToken) ||
		isComma(parser.NextToken)
}

func format(input string) string {

	parser := newParser(input)

	directive := false

	structUnionOrEnum := false

	for parser.parseToken() {

		//fmt.Println(t)

		parser.Output.WriteString(parser.Token.Content)

		if isLeftBrace(parser.Token) {
			if isAssignement(parser.PreviousToken) {
				formatInitialiserList(parser)
				continue
			} else if structUnionOrEnum {
				formatDeclarationBody(parser)
				structUnionOrEnum = false
				continue
			} else {
				formatBlockBody(parser)
				continue
			}

		}

		if isStructUnionEnumKeyword(parser.Token) {
			structUnionOrEnum = true
		}

		if isDirective(parser.Token) {
			directive = true
		}

		const maxNewLines = 2

		endOfDirective := directive && parser.NewLinesAfter > 0

		isBlockStart := isLeftBrace(parser.Token) && !isAssignement(parser.PreviousToken)

		if isBlockStart {
			parser.writeNewLines(1)
		} else if endOfDirective ||
			isDirective(parser.NextToken) ||
			(isSemicolon(parser.Token) && !parser.IsParenthesis) {

			if parser.NewLinesAfter > 1 {
				parser.writeNewLines(2)

			} else {
				parser.writeNewLines(1)
			}

		} else if !neverWhiteSpace(parser) &&
			!isRightBrace(parser.NextToken) &&
			!isDotOperator(parser.NextToken) &&
			!isLeftBrace(parser.Token) {
			parser.Output.WriteString(" ")
		}

		if parser.NewLinesAfter > 0 {
			directive = false
		}

	}

	return parser.Output.String()
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
