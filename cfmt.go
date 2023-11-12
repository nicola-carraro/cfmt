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

type NodeType uint32

const (
	NoNodeType NodeType = iota
	Braces
	Parenthesis
	Statement
	Preprocessor
	FuncSpecifier
)

type Node struct {
	Type     NodeType
	Tokens   []Token
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
		return fmt.Sprintf("Token{Type: %s, len: %d, NewLines: %d}", t.Type, len(t.Content), t.NewLines)
	} else if t.Type == NoTokenType {
		return fmt.Sprintf("Token{Type: %s}", t.Type)
	} else {
		return fmt.Sprintf("Token{Type: %s, Content: \"%s\"}", t.Type, t.Content)
	}
}

func (n NodeType) String() string {

	switch n {
	case NoNodeType:
		return "None"
	case Braces:
		return "Braces"
	case Parenthesis:
		return "Parenthesis"
	case Statement:
		return "Statement"
	case Preprocessor:
		return "Preprocessor"
	case FuncSpecifier:
		return "FuncSpecifier"
	default:
		panic("Unknown NodeType")
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

func (n Node) String() string {

	b := strings.Builder{}
	for _, t := range n.Tokens {
		var s string
		if t.Type == Punctuation {
			s = fmt.Sprintf("%s, ", t.Content)
		} else {
			s = fmt.Sprintf("%s, ", t.Type)

		}
		_, _ = b.WriteString(s)
	}
	tokens := b.String()

	b.Reset()
	for _, c := range n.Children {
		s := fmt.Sprintf("%s, ", c.Type)
		_, _ = b.WriteString(s)
	}
	children := b.String()

	return fmt.Sprintf("Node{Type: %s, Tokens: [%s], Children: [%s]}", n.Type, tokens, children)
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

func parseParenthesis(tokens []Token) Node {

	len := 0

	for _, t := range tokens {
		len++
		if isRightParenthesis(t) {
			break
		}

	}

	node := Node{Type: Parenthesis, Tokens: tokens[:len]}

	return node
}

func parseBraces(tokens []Token) Node {

	length := 1

	node := Node{Type: Braces, Tokens: tokens, Children: make([]Node, 0)}

	//log.Printf("tokens %s\n", tokens)

	content := tokens[1:]

	for len(content) > 0 {
		t, _ := skipSpaceAndCountNewLines(content)
		//log.Printf("content %s\n", content)

		if isRightBrace(t) {
			length++
			break
		}

		child := parseNode(content)

		log.Println("child ", child)
		node.Children = append(node.Children, child)
		content = content[len(child.Tokens):]
		length += len(node.Tokens)

	}

	node.Tokens = node.Tokens[:length]

	return node
}

func parsePreprocessor(tokens []Token) Node {

	len := 0

	for _, t := range tokens {
		len++
		if hasNewline(t) {
			break
		}

	}

	node := Node{Type: Preprocessor, Tokens: tokens[:len]}

	return node
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

func parseStatementOrFuncSpecifier(tokens []Token) Node {
	length := 0
	var node Node

	for _, t := range tokens {

		if isLeftBrace(t) {
			node = Node{Type: FuncSpecifier, Tokens: tokens[:length]}
			return node
		}

		length++

		if isSemicolon(t) {
			node = Node{Type: Statement, Tokens: tokens[:length]}
			return node
		}

	}

	log.Fatalf("Unexpected end of file: expected } or ;, found %s\n", tokens[len(tokens)-1])

	panic("Unreacheable")
}

func parseNode(tokens []Token) Node {
	token, _ := skipSpaceAndCountNewLines(tokens)
	var node Node
	if isHash(token) {
		//fmt.Println(token)
		node = parsePreprocessor(tokens)
	} else if isLeftParenthesis(token) {
		node = parseParenthesis(tokens)
	} else if isLeftBrace(token) {
		node = parseBraces(tokens)
	} else {
		node = parseStatementOrFuncSpecifier(tokens)

	}

	return node
}

func parse(tokens []Token) []Node {
	nodes := make([]Node, 0)

	for len(tokens) > 0 {
		node := parseNode(tokens)
		nodes = append(nodes, node)
		tokens = tokens[len(node.Tokens):]
	}

	return nodes
}

func isPreprocessorDirective(token Token) bool {
	return token.Type == PreprocessorDirective
}

func main() {
	const path = "test.c"

	data, error := os.ReadFile(path)

	if error != nil {
		log.Fatal("Error reading ", path)
	}

	text := fmt.Sprintf("%s", data)

	tokens := tokenize(text)

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

			endOfDirective := isDirective && newLinesAfter > 0

			if isLeftBrace(t) || isRightBrace(t) || isRightBrace(nextT) || endOfDirective || isPreprocessorDirective(nextT) || isEndOfStatement {
				b.WriteString(newLine)

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

	fmt.Println(b.String())

	//nodes := parse(tokens)
	// for _, n := range nodes {
	// 	fmt.Println(n)
	// }

}
