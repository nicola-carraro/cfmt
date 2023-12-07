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
	SingleLineComment
	MultilineComment
)

type Token struct {
	Type       TokenType
	Content    string
	Whitespace Whitespace
}

type StructUnionEnum struct {
	Indent int
}

type Parser struct {
	PreviousToken         Token
	Token                 Token
	NextToken             Token
	Indent                int
	InputLine             int
	InputColumn           int
	OutputLine            int
	OutputColumn          int
	Input                 string
	Output                strings.Builder
	SavedInput            string
	IsParenthesis         bool
	IsDirective           bool
	IsIncludeDirective    bool
	IsEndOfDirective      bool
	RightSideOfAssignment bool
}

type Whitespace struct {
	NewLines          int
	HasUnescapedLines bool
}

const indentation = "    "

const allowWrap = 90

const maxInlineFunctionArgs = 5

func isIdentifierStart(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r == '_')
}

func isIdentifierChar(r rune) bool {
	return isIdentifierStart(r) || isDigit(r)
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r' || r == '\n' || r == '\v' || r == '\f'
}

func (parser *Parser) consumeSpace(Whitespace *Whitespace) bool {

	newLineInDirective := []string{"\\\r\n", "\\\n"}

	if parser.IsDirective {
		for _, nl := range newLineInDirective {
			if parser.IsDirective && strings.HasPrefix(parser.Input, nl) {
				parser.Input = parser.Input[len(nl):]
				Whitespace.NewLines++
				return true
			}
		}

	}

	r, size := peakRune(parser.Input)

	if r == '\n' {
		parser.Input = parser.Input[size:]
		Whitespace.NewLines++
		Whitespace.HasUnescapedLines = true

		if parser.IsDirective {
			parser.IsDirective = false
			parser.IsIncludeDirective = false
		}
		return true
	}

	otherSpaces := []rune{' ', '\t', '\r', '\v', '\f'}

	if slices.Contains(otherSpaces, r) {
		parser.Input = parser.Input[size:]
		return true

	}

	return false
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
	case SingleLineComment:
		return "SingleLineComment"
	case MultilineComment:
		return "MultilineComment"
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
		"#define", "#elif", "#else", "#endif",
		"#error", "#ifndef", "#ifdef", "#if",
		"#import", "#include", "#line", "#pragma",
		"#undef", "#using"}

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

func parseMultilineComment(text string) Token {
	tokenSize := 0
	next := text

	_, size := peakRune(next)

	for !strings.HasPrefix(next, "*/") {
		if len(next) == 0 {
			log.Fatalln("Unclosed comment")
		}
		tokenSize += size
		next = next[size:]
		_, size = peakRune(next)
	}

	tokenSize += 2

	token := Token{Type: MultilineComment, Content: text[:tokenSize]}
	return token
}

func parseSingleLineComment(text string) Token {
	tokenSize := 0
	next := text

	_, size := peakRune(next)

	for len(next) > 0 && !startsWithNewLine(next) {
		tokenSize += size
		next = next[size:]
		_, size = peakRune(next)
	}

	token := Token{Type: SingleLineComment, Content: text[:tokenSize]}
	//fmt.Println(token)
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
	var tokenSize int
	if text[0] == 'L' {
		tokenSize = 2
	} else {
		tokenSize = 1
	}

	next := text[tokenSize:]

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

func longSuffixLength(text string) int {
	suffixes := []string{"i64", "ll", "l"}

	for _, s := range suffixes {
		if strings.HasPrefix(text, s) || strings.HasPrefix(text, strings.ToUpper(s)) {
			return len(s)
		}
	}

	return 0
}

func suffixLength(text string) int {
	next := text
	result := 0
	if isUnsignedSuffix(text) {
		result++
		next = next[1:]
		result += longSuffixLength(next)
	} else {
		lSuffixLength := longSuffixLength(next)
		next = next[lSuffixLength:]
		result += lSuffixLength
		if isUnsignedSuffix(next) {
			result++
		}
	}

	return result
}

func isUnsignedSuffix(text string) bool {
	return strings.HasPrefix(text, "u") || strings.HasPrefix(text, "U")
}

func isOctalDigit(r rune) bool {
	return (r >= '0' && r <= '7')
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func parseHex(text string) Token {
	tokenSize := 2

	next := text[2:]

	r, size := peakRune(next)

	for isHexDigit(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}

	tokenSize += suffixLength(next)

	token := Token{Type: Integer, Content: text[:tokenSize]}
	return token
}

func parseOctal(text string) Token {
	tokenSize := 0

	next := text

	r, size := peakRune(next)

	for isOctalDigit(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}

	tokenSize += suffixLength(next)

	token := Token{Type: Integer, Content: text[:tokenSize]}
	return token
}

func parseDecimal(text string) Token {
	tokenSize := 0

	next := text

	r, size := peakRune(next)

	for isDigit(r) {
		tokenSize += size
		next = next[size:]
		r, size = peakRune(next)
	}

	tokenSize += suffixLength(next)

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

func isIncludeDirective(token Token) bool {
	return isDirective(token) && token.Content == "#include"
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

	if isAbsent(parser.Token) {
		_ = skipSpaceAndCountNewLines(parser)

		parser.Token = parseToken(parser.Input)
		parser.Input = parser.Input[len(parser.Token.Content):]
	} else {
		parser.PreviousToken = parser.Token
		parser.Token = parser.NextToken
	}

	parser.IsEndOfDirective = false

	if isDirective(parser.Token) {
		parser.IsDirective = true
	}

	wasDirective := parser.IsDirective

	if isIncludeDirective(parser.Token) {
		parser.IsIncludeDirective = true
	}

	if isAssignment(parser.Token) {
		parser.RightSideOfAssignment = true
	}

	if isSemicolon(parser.Token) {
		parser.RightSideOfAssignment = false
	}

	parser.Token.Whitespace = skipSpaceAndCountNewLines(parser)

	parser.NextToken = parseToken(parser.Input)
	parser.Input = parser.Input[len(parser.NextToken.Content):]

	if parser.Token.Whitespace.HasUnescapedLines {
		parser.IsDirective = false
		parser.IsIncludeDirective = false
	}

	if wasDirective && !parser.IsDirective {
		parser.IsEndOfDirective = true
		//fmt.Printf("END OF DIRECTIVE %s\n", parser.Token)
	}

	//fmt.Printf("updateParser, PreviousToken:%s, Token:%s, NextToken:%s\n", parser.PreviousToken, parser.Token, parser.NextToken)

	if isLeftParenthesis(parser.Token) {
		parser.IsParenthesis = true
	}

	if isRightParenthesis(parser.Token) {
		parser.IsParenthesis = false
	}

	return !isAbsent(parser.Token)
}

func isOneToNine(r rune) bool {
	return r >= '1' && r <= '9'
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

	if isDoubleQuote(r) || strings.HasPrefix(input, "L\"") {
		return parseString(input)
	}

	if isIdentifierStart(r) {
		return parseIdentifier(input)
	}

	if strings.HasPrefix(input, "//") {
		return parseSingleLineComment(input)
	}

	if strings.HasPrefix(input, "/*") {
		return parseMultilineComment(input)
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

	if strings.HasPrefix(input, "0x") || strings.HasPrefix(input, "0X") {
		return parseHex(input)

	}

	if isOneToNine(r) {
		//TODO: handle octal and hex
		return parseDecimal(input)
	}

	if r == '0' {
		return parseOctal(input)
	}

	max := 20
	start := input
	if len(start) > max {
		start = start[:max]
	}
	log.Fatalf("Unrecognised token, starts with %s", start)

	panic("Unreachable")
}

func startsWithNewLine(input string) bool {
	return strings.HasPrefix(input, "\r\n")
}

func skipSpaceAndCountNewLines(parser *Parser) Whitespace {

	result := Whitespace{}

	for parser.consumeSpace(&result) {
	}

	return result

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

func isSingleLineComment(token Token) bool {
	return token.Type == SingleLineComment
}

func isMultilineComment(token Token) bool {
	return token.Type == MultilineComment
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

func isAssignment(token Token) bool {
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

func (parser *Parser) writeString(str string) {
	parser.Output.WriteString(str)
	parser.OutputColumn += len(str)
}

func (parser *Parser) writeNewLines(lines int) {
	const newLine = "\r\n"

	for line := 0; line < lines; line++ {

		if parser.IsDirective {
			parser.writeString("\\")
		}
		parser.writeString(newLine)
		parser.OutputColumn = 0
		parser.OutputLine++
	}

	for indentLevel := 0; indentLevel < parser.Indent; indentLevel++ {
		parser.writeString(indentation)
	}
}

func hasNewLines(token Token) bool {
	return token.Whitespace.NewLines > 0
}

func isComment(token Token) bool {
	return isSingleLineComment(token) || isMultilineComment(token)
}

func tryFormatInlineInitialiserList(parser *Parser) bool {
	openBraces := 1

	//fmt.Println(parser.Input)

	for parser.parseToken() {

		parser.formatToken()

		if isRightBrace(parser.Token) {
			openBraces--
		}

		if isComma(parser.Token) && hasNewLines(parser.Token) {
			return false
		}

		if isComment(parser.Token) {
			return false
		}

		if isLeftBrace(parser.Token) {

			if !tryFormatInlineInitialiserList(parser) {
				return false
			}
			continue
		}

		if isComment(parser.Token) {
			parser.writeNewLines(1)
		} else if !neverWhitespace(parser) &&
			!isRightBrace(parser.NextToken) &&
			!isRightBrace(parser.Token) {
			parser.writeString(" ")
		}

		if openBraces == 0 {
			return true
		}
	}

	log.Fatal("Unclosed initialiser list")
	panic("unreachable")
}

func formatMultilineInitialiserList(parser *Parser) {
	openBraces := 1

	parser.Indent++

	parser.writeNewLines(1)

	for parser.parseToken() {

		parser.formatToken()
		if isRightBrace(parser.NextToken) {
			parser.Indent--
		}

		if isRightBrace(parser.Token) {
			openBraces--
		}

		if isLeftBrace(parser.Token) {

			formatMultilineInitialiserList(parser)

			continue
		}

		if isRightBrace(parser.NextToken) {
			parser.writeNewLines(1)
		} else if isComma(parser.Token) && hasNewLines(parser.Token) {
			parser.writeNewLines(1)
		} else if isComment(parser.Token) || isMultilineComment(parser.Token) {
			parser.writeNewLines(1)
		} else if !neverWhitespace(parser) &&
			!isRightBrace(parser.NextToken) &&
			!isRightBrace(parser.Token) {
			parser.writeString(" ")
		}

		if openBraces == 0 {
			return
		}
	}

	log.Fatal("Unclosed initialiser list")
}

func tryFormatInlineFunctionArguments(parser *Parser) bool {
	commas := 0
	for parser.parseToken() {

		if(isComma(parser.Token)){
			commas++
		}

		if parser.OutputColumn > allowWrap  {
			return false
		}

		if(commas >= maxInlineFunctionArgs){
			return false
		}

		parser.formatToken()

		if isRightParenthesis(parser.Token) {
			parser.writeString(" ")
			return true
		}

		if isSingleLineComment(parser.Token) {
			parser.writeNewLines(1)
		} else if !neverWhitespace(parser) &&
			!isRightBrace(parser.NextToken) &&
			!isRightBrace(parser.Token) {
			parser.writeString(" ")
		}
	}

	log.Fatal("Unclosed function arguments")

	panic("unreachable")
}

func formatMultilineFunctionArguments(parser *Parser) {
	parser.Indent++

	parser.writeNewLines(1)

	for parser.parseToken() {
		parser.formatToken()

		if isRightParenthesis(parser.NextToken) {
			parser.Indent--
		}

		if isRightParenthesis(parser.Token) {
			if isAbsent(parser.NextToken) {
				parser.writeNewLines(1)

			} else {
				parser.writeString(" ")
			}
			return
		}

		if isComma(parser.Token) || isRightParenthesis(parser.NextToken) {
			parser.writeNewLines(1)
		} else if isSingleLineComment(parser.Token) {
			parser.writeNewLines(1)
		} else if !neverWhitespace(parser) &&
			!isRightBrace(parser.NextToken) &&
			!isRightBrace(parser.Token) {
			parser.writeString(" ")
		}
	}

	log.Fatal("Unclosed function arguments")
}

func formatFunctionArguments(parser *Parser) {
	saved := *parser
	if !tryFormatInlineFunctionArguments(parser) {
		*parser = saved
		formatMultilineFunctionArguments(parser)
	}
}

func formatInitialiserList(parser *Parser) {

	initialState := *parser

	if !tryFormatInlineInitialiserList(parser) {
		*parser = initialState
		formatMultilineInitialiserList(parser)
	}

}

func (parser *Parser) oneOrTwoLines() {
	if parser.Token.Whitespace.NewLines <= 1 || isRightBrace(parser.NextToken) {
		parser.writeNewLines(1)

	} else {
		parser.writeNewLines(2)
	}
}

func (parser *Parser) threeLinesOrEof() {
	if isAbsent(parser.NextToken) {
		parser.writeNewLines(1)
	} else {
		parser.writeNewLines(3)
	}
}

func isPointerOperator(parser *Parser) bool {
	return canBePointerOperator(parser.Token) && (!canBeLeftOperand(parser.PreviousToken) || !parser.RightSideOfAssignment)
}

func hasPostfixIncrDecr(parser *Parser) bool {
	return isIncrDecrOperator(parser.NextToken) && (isIdentifier(parser.Token) || isRightParenthesis(parser.Token))
}

func isPrefixIncrDecr(parser *Parser) bool {
	return isIncrDecrOperator(parser.Token) && (isIdentifier(parser.NextToken) || isLeftParenthesis(parser.NextToken))
}

func (parser *Parser) hasTrailingComment() bool {
	return isSingleLineComment(parser.NextToken) && parser.Token.Whitespace.NewLines == 0
}

func formatBlockBody(parser *Parser) {

	parser.Indent++

	if isRightBrace(parser.NextToken) {

		parser.Indent--
	}

	parser.writeNewLines(1)

	structUnionOrEnum := false

	//fmt.Printf("start %s\n", parser.NextToken)

	for parser.parseToken() {

		if isRightBrace(parser.NextToken) {
			parser.Indent--
		}

		if isStructUnionEnumKeyword(parser.Token) {
			structUnionOrEnum = true
		}

		parser.formatToken()

		if isRightBrace(parser.Token) {
			return
		}

		if isLeftBrace(parser.Token) {

			if isAssignment(parser.PreviousToken) {
				formatInitialiserList(parser)
			} else if structUnionOrEnum {
				formatDeclarationBody(parser)
				structUnionOrEnum = false
			} else {
				formatBlockBody(parser)
				parser.oneOrTwoLines()
			}
		} else if isComment(parser.Token) || isMultilineComment(parser.Token) {
			parser.writeNewLines(1)
		} else if (isSemicolon(parser.Token) && !parser.IsParenthesis && !parser.hasTrailingComment()) || parser.IsEndOfDirective {
			parser.oneOrTwoLines()
		} else if !neverWhitespace(parser) &&
			!isDotOperator(parser.NextToken) {
			parser.writeString(" ")
		}
	}

	log.Fatal("Unclosed block")
}

func (parser *Parser) formatMultilineComment() {
	text := strings.TrimSpace(parser.Token.Content[2 : len(parser.Token.Content)-2])

	lines := strings.Split(text, "\n")
	parser.writeString("/*")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		parser.writeNewLines(1)
		if len(trimmed) > 0 {
			parser.writeString("   ")
			parser.writeString(trimmed)
		}
	}
	parser.writeNewLines(1)
	parser.writeString("*/")
}

func (parser *Parser) formatSingleLineComment() {
	text := strings.TrimSpace(parser.Token.Content[2:])
	parser.writeString("// ")
	parser.writeString(text)
}

func (parser *Parser) formatToken() {

	if isMultilineComment(parser.Token) {
		parser.formatMultilineComment()
	} else if isSingleLineComment(parser.Token) {
		parser.formatSingleLineComment()
	} else {
		parser.writeString(parser.Token.Content)
	}
}

func formatDeclarationBody(parser *Parser) {

	//fmt.Printf("DECLARATION: %s\n", parser.Input)

	initialIndent := parser.Indent
	parser.Indent++

	parser.writeNewLines(1)

	for parser.parseToken() {

		parser.formatToken()

		if isRightBrace(parser.NextToken) {
			parser.Indent--

		}

		if isLeftBrace(parser.Token) {
			formatDeclarationBody(parser)
		} else if isComment(parser.Token) || isComment(parser.NextToken) {
			parser.writeNewLines(1)
		} else if isSemicolon(parser.Token) && !parser.hasTrailingComment() {
			parser.writeNewLines(1)
		} else if !neverWhitespace(parser) &&
			!isDotOperator(parser.NextToken) &&
			!isSemicolon(parser.NextToken) {
			parser.writeString(" ")
		}

		if parser.Indent == initialIndent {
			return
		}
	}

	log.Fatal("Unclosed declaration braces")
}

func startsFunctionArguments(parser *Parser) bool {
	return parser.PreviousToken.Type == Identifier && isLeftParenthesis(parser.Token)
}

func isFunctionName(parser *Parser) bool {
	return parser.Token.Type == Identifier && isLeftParenthesis(parser.NextToken)
}

func isInclude(token Token) bool {
	return token.Type == Directive && token.Content == "#include"
}

func isLeftBracket(token Token) bool {
	return token.Type == Punctuation && token.Content == "["
}

func isRightBracket(token Token) bool {
	return token.Type == Punctuation && token.Content == "]"
}

func isNegation(token Token) bool {
	return token.Type == Punctuation && token.Content == "!"
}

func neverWhitespace(parser *Parser) bool {

	return isSemicolon(parser.NextToken) ||
		isLeftParenthesis(parser.Token) ||
		isRightParenthesis(parser.NextToken) ||
		isLeftBracket(parser.Token) ||
		isLeftBracket(parser.NextToken) ||
		isRightBracket(parser.NextToken) ||
		isPointerOperator(parser) ||
		isFunctionName(parser) ||
		hasPostfixIncrDecr(parser) ||
		isPrefixIncrDecr(parser) ||
		isDotOperator(parser.Token) ||
		isArrowOperator(parser.Token) ||
		isArrowOperator(parser.NextToken) ||
		isComma(parser.NextToken) ||
		isNegation(parser.Token) ||
		(parser.IsIncludeDirective && (isGreaterThanSign(parser.NextToken) || isLessThanSign(parser.Token)))
}

func isGreaterThanSign(token Token) bool {
	return token.Type == Punctuation && token.Content == ">"
}

func isLessThanSign(token Token) bool {
	return token.Type == Punctuation && token.Content == "<"
}

func format(input string) string {

	parser := newParser(input)

	structUnionOrEnum := false

	for parser.parseToken() {

		parser.formatToken()

		if startsFunctionArguments(parser) {
			formatFunctionArguments(parser)
			continue
		}

		if isLeftBrace(parser.Token) {
			if isAssignment(parser.PreviousToken) {
				formatInitialiserList(parser)
			} else if structUnionOrEnum {
				formatDeclarationBody(parser)
				structUnionOrEnum = false
			} else {
				formatBlockBody(parser)
				parser.threeLinesOrEof()
			}
			continue
		}

		if isStructUnionEnumKeyword(parser.Token) {
			structUnionOrEnum = true
		}

		const maxNewLines = 2

		isBlockStart := isLeftBrace(parser.Token) && !isAssignment(parser.PreviousToken)

		if isBlockStart || isComment(parser.Token) {
			parser.writeNewLines(1)
		} else if parser.IsEndOfDirective ||
			isDirective(parser.NextToken) ||
			(isSemicolon(parser.Token) && !parser.IsParenthesis && !parser.hasTrailingComment()) {
			parser.threeLinesOrEof()
		} else if !neverWhitespace(parser) &&
			!isRightBrace(parser.NextToken) &&
			!isDotOperator(parser.NextToken) &&
			!isLeftBrace(parser.Token) {
			parser.writeString(" ")
		}
	}

	return parser.Output.String()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path>\n", os.Args[0])
	}

	path := os.Args[1]
	data, err := os.ReadFile(path)

	if err != nil {
		log.Fatalf("Error reading %s: %s", path, err)
	}

	text := fmt.Sprintf("%s", data)

	formattedText := format(text)

	fmt.Println(formattedText)
}
