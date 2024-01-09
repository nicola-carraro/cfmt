package main

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

type Token struct {
	Type          TokenType
	Content       string
	Whitespace    Whitespace
	DirectiveType DirectiveType
	KeywordType   KeywordType
	Line          int
	Column        int
}

type TokenType uint32

const (
	TokenTypeNone TokenType = iota
	TokenTypeKeyword
	TokenTypeIdentifier
	TokenTypeConstant
	TokenTypePunctuation
	TokenTypeDirective
	TokenTypeSingleLineComment
	TokenTypeMultilineComment
	TokenTypeInvalid
)

type DirectiveName struct {
	Name          string
	DirectiveType DirectiveType
}

type KeywordType int

const (
	KeywordTypeNone KeywordType = iota
	KeywordTypeAuto
	KeywordTypeBreak
	KeywordTypeCase
	KeywordTypeChar
	KeywordTypeConst
	KeywordTypeContinue
	KeywordTypeDefault
	KeywordTypeDo
	KeywordTypeDouble
	KeywordTypeElse
	KeywordTypeEnum
	KeywordTypeExtern
	KeywordTypeFloat
	KeywordTypeFor
	KeywordTypeGoto
	KeywordTypeIf
	KeywordTypeInline
	KeywordTypeInt
	KeywordTypeLong
	KeywordTypeRegister
	KeywordTypeRestrict
	KeywordTypeReturn
	KeywordTypeShort
	KeywordTypeSigned
	KeywordTypeSizeof
	KeywordTypeStatic
	KeywordTypeStruct
	KeywordTypeSwitch
	KeywordTypeTypedef
	KeywordTypeUnion
	KeywordTypeUnsigned
	KeywordTypeVoid
	KeywordTypeVolatile
	KeywordTypeWhile
	KeywordTypeAlignas
	KeywordTypeAlignof
	KeywordTypeAtomic
	KeywordTypeBool
	KeywordTypeComplex
	KeywordTypeGeneric
	KeywordTypeImaginary
	KeywordTypeNoreturn
	KeywordTypeStaticAssert
	KeywordTypeThreadLocal
	KeywordTypeAsm
	KeywordTypeBased
	KeywordTypeCdecl
	KeywordTypeDeclspec
	KeywordTypeExcept
	KeywordTypeFastcall
	KeywordTypeFinally
	KeywordTypeInt16
	KeywordTypeInt32
	KeywordTypeInt64
	KeywordTypeInt8
	KeywordTypeLeave
	KeywordTypeStdcall
	KeywordTypeTry
)

type KeywordName struct {
	Name        string
	KeywordType KeywordType
}

type Whitespace struct {
	HasSpace          bool
	NewLines          int
	HasUnescapedLines bool
	HasEscapedLines   bool
}

type IsDigitFunction func(r rune) bool

func parseToken(input string) Token {

	if len(input) == 0 {
		return Token{}
	}

	r, _ := utf8.DecodeRuneInString(input)

	token, isFloat := tryParseFloat(input)
	if isFloat {
		return token
	}

	token, directive := tryParseDirective(input)
	if directive {
		return token
	}

	if isDoubleQuote(r) || strings.HasPrefix(input, "L\"") {
		return parseString(input)
	}

	if isIdentifierStart(r) {
		return parseIdentifierOrKeyword(input)
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
		token.Type = TokenTypePunctuation
		token.Content = input[:4]
		return token
	}

	if isThreeCharsPunctuation(input) {
		token.Type = TokenTypePunctuation
		token.Content = input[:3]

		return token
	}

	if isTwoCharsPunctuation(input) {
		token.Type = TokenTypePunctuation
		token.Content = input[:2]

		return token
	}

	if isOneCharPunctuation(input) {
		token.Type = TokenTypePunctuation
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

	return Token{Type: TokenTypeInvalid}
}

func parseIdentifierOrKeyword(text string) Token {

	tokenSize := 0
	next := text

	r, size := utf8.DecodeRuneInString(next)

	for isIdentifierChar(r) {
		tokenSize += size
		next = next[size:]
		r, size = utf8.DecodeRuneInString(next)
	}

	content := text[:tokenSize]

	keywords := [...]KeywordName{
		{"auto", KeywordTypeAuto},
		{"break", KeywordTypeBreak},
		{"case", KeywordTypeCase},
		{"char", KeywordTypeChar},
		{"const", KeywordTypeConst},
		{"continue", KeywordTypeContinue},
		{"default", KeywordTypeDefault},
		{"double", KeywordTypeDouble},
		{"do", KeywordTypeDo},
		{"else", KeywordTypeElse},
		{"enum", KeywordTypeEnum},
		{"extern", KeywordTypeExtern},
		{"float", KeywordTypeFloat},
		{"for", KeywordTypeFor},
		{"goto", KeywordTypeGoto},
		{"if", KeywordTypeIf},
		{"inline", KeywordTypeInline},
		{"int", KeywordTypeInt},
		{"long", KeywordTypeLong},
		{"register", KeywordTypeRegister},
		{"return", KeywordTypeReturn},
		{"short", KeywordTypeShort},
		{"signed", KeywordTypeSigned},
		{"sizeof", KeywordTypeSizeof},
		{"static_assert", KeywordTypeStaticAssert},
		{"static", KeywordTypeStatic},
		{"struct", KeywordTypeStruct},
		{"switch", KeywordTypeSwitch},
		{"typedef", KeywordTypeTypedef},
		{"union", KeywordTypeUnion},
		{"unsigned", KeywordTypeUnsigned},
		{"void", KeywordTypeVoid},
		{"volatile", KeywordTypeVolatile},
		{"while", KeywordTypeWhile},
		{"_Alignof", KeywordTypeAlignof},
		{"_Atomic", KeywordTypeAtomic},
		{"_Bool", KeywordTypeBool},
		{"_Complex", KeywordTypeComplex},
		{"_Generic", KeywordTypeGeneric},
		{"_Imaginary", KeywordTypeImaginary},
		{"_Noreturn", KeywordTypeNoreturn},
		{"_Static_assert", KeywordTypeStaticAssert},
		{"_Thread_local", KeywordTypeThreadLocal},
		{"__asm", KeywordTypeAsm},
		{"__based", KeywordTypeBased},
		{"__cdecl", KeywordTypeCdecl},
		{"__declspec", KeywordTypeDeclspec},
		{"__except", KeywordTypeExcept},
		{"__fastcall", KeywordTypeFastcall},
		{"__finally", KeywordTypeFinally},
		{"__inline", KeywordTypeInline},
		{"__int16", KeywordTypeInt16},
		{"__int32", KeywordTypeInt32},
		{"__int64", KeywordTypeInt64},
		{"__int8", KeywordTypeInt8},
		{"__leave", KeywordTypeLeave},
		{"__restrict", KeywordTypeRestrict},
		{"__stdcall", KeywordTypeStdcall},
		{"__try", KeywordTypeTry},
	}

	for _, keyword := range keywords {
		if content == keyword.Name {
			return Token{Type: TokenTypeKeyword, Content: content, KeywordType: keyword.KeywordType}
		}
	}

	return Token{Type: TokenTypeIdentifier, Content: content}
}

func parseString(text string) Token {
	var tokenSize int
	if text[0] == 'L' {
		tokenSize = 2
	} else {
		tokenSize = 1
	}

	next := text[tokenSize:]

	for {
		r, size := utf8.DecodeRuneInString(next)
		tokenSize += size
		next = next[size:]
		if r == '"' {
			token := Token{Type: TokenTypeConstant, Content: text[:tokenSize]}
			return token
		} else if r == '\\' {
			_, size := utf8.DecodeRuneInString(next)
			tokenSize += size
			next = next[size:]
		} else if r == utf8.RuneError && size == 0 {
			return Token{Type: TokenTypeInvalid}
		}
	}
}

func parseChar(text string) Token {
	tokenSize := 1
	next := text[1:]

	for {
		r, size := utf8.DecodeRuneInString(next)
		tokenSize += size
		next = next[size:]
		if r == '\'' {
			token := Token{Type: TokenTypeConstant, Content: text[:tokenSize]}
			return token
		} else if r == '\\' {
			_, size := utf8.DecodeRuneInString(next)
			next = next[size:]
			tokenSize += size

		} else if r == utf8.RuneError && size == 0 {
			return Token{Type: TokenTypeInvalid}
		}
	}
}

func tryParseFloat(text string) (Token, bool) {

	tokenSize := 0
	next := text

	hasDigit := false

	r, size := utf8.DecodeRuneInString(next)

	if !isDecimal(r) && r != '.' {
		return Token{}, false
	}

	for isDecimal(r) {
		tokenSize += size
		next = next[size:]
		r, size = utf8.DecodeRuneInString(next)
		hasDigit = true
	}

	hasDot := false

	if r == '.' {
		tokenSize += size
		next = next[size:]
		r, size = utf8.DecodeRuneInString(next)
		hasDot = true
	}

	for isDecimal(r) {

		tokenSize += size
		next = next[size:]
		r, size = utf8.DecodeRuneInString(next)
		hasDigit = true
	}

	hasExponent := false

	if isExponentStart(r) {
		tokenSize += size
		next = next[size:]
		r, size = utf8.DecodeRuneInString(next)

		if isSignStart(r) {
			tokenSize += size
			next = next[size:]
			r, size = utf8.DecodeRuneInString(next)
		}

		for isDecimal(r) {
			hasExponent = true
			tokenSize += size
			next = next[size:]
			r, size = utf8.DecodeRuneInString(next)
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

	token := Token{Type: TokenTypeConstant, Content: text[:tokenSize]}

	return token, true

}

func parseDecimal(text string) Token {
	return parseInt(text, 0, isDecimal)
}

func parseHex(text string) Token {
	return parseInt(text, 2, isHexDigit)
}

func parseOctal(text string) Token {
	return parseInt(text, 0, isOctalDigit)
}

func parseInt(text string, prefixLen int, isDigit IsDigitFunction) Token {

	tokenSize := prefixLen

	next := text[prefixLen:]
	r, size := utf8.DecodeRuneInString(next)

	for isDigit(r) {
		tokenSize += size
		next = next[size:]
		r, size = utf8.DecodeRuneInString(next)
	}

	tokenSize += suffixLength(next)

	return Token{Type: TokenTypeConstant, Content: text[:tokenSize]}
}

func parseMultilineComment(text string) Token {
	tokenSize := 2
	next := text[2:]

	_, size := utf8.DecodeRuneInString(next)

	for !strings.HasPrefix(next, "*/") {
		if len(next) == 0 {
			return Token{Type: TokenTypeInvalid}
		}
		tokenSize += size
		next = next[size:]
		_, size = utf8.DecodeRuneInString(next)
	}

	tokenSize += 2

	token := Token{Type: TokenTypeMultilineComment, Content: text[:tokenSize]}
	return token
}

func tryParseDirective(s string) (Token, bool) {

	directives := [...]DirectiveName{
		{"#define", DirectiveTypeDefine},
		{"#elif", DirectiveTypeElif}, {"#else", DirectiveTypeElse},
		{"#endif", DirectiveTypeEndif},
		{"#error", DirectiveTypeError},
		{"#ifndef", DirectiveTypeIfndef},
		{"#ifdef", DirectiveTypeIfdef},
		{"#if", DirectiveTypeIf},
		{"#include", DirectiveTypeInclude},
		{"#include", DirectiveTypeInclude},
		{"#pragma", DirectiveTypePragma},
		{"#undef", DirectiveTypeUndef},
		{"#undef", DirectiveTypeUndef},
	}

	for _, directive := range directives {
		if strings.HasPrefix(s, directive.Name) {
			return Token{Type: TokenTypeDirective, Content: directive.Name, DirectiveType: directive.DirectiveType}, true
		}
	}

	return Token{}, false
}

func parseSingleLineComment(text string) Token {
	tokenSize := 0
	next := text

	_, size := utf8.DecodeRuneInString(next)

	for len(next) > 0 && !startsWithNewLine(next) {
		tokenSize += size
		next = next[size:]
		_, size = utf8.DecodeRuneInString(next)
	}

	token := Token{Type: TokenTypeSingleLineComment, Content: text[:tokenSize]}
	return token
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

func isOneToNine(r rune) bool {
	return r >= '1' && r <= '9'
}

func startsWithNewLine(input string) bool {
	return strings.HasPrefix(input, "\r\n") || strings.HasPrefix(input, "\n")
}

func isDoubleQuote(r rune) bool {
	return r == '"'
}

func isSingleQuote(r rune) bool {
	return r == '\''
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

func isExponentStart(r rune) bool {
	return r == 'e' || r == 'E'
}

func isFloatSuffix(r rune) bool {
	return r == 'f' || r == 'l' || r == 'F' || r == 'L'
}

func isSignStart(r rune) bool {
	return r == '+' || r == '-'
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

func isIdentifierStart(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r == '_')
}

func isIdentifierChar(r rune) bool {
	return isIdentifierStart(r) || isDecimal(r)
}

func isDecimal(r rune) bool {
	return r >= '0' && r <= '9'
}

func isTwoCharsPunctuation(text string) bool {
	twoChars := [...]string{"++", "--", "<<", ">>", "<=", ">=", "==",
		"!=", "&&", "||", "->", "*=", "/=", "%=", "+=", "-=", "&=", "^=", "|=", "##",
		"<:", ":>", "<%", "%>", "%:", "#@"}

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

func (t Token) String() string {
	if t.Type == TokenTypeNone {
		return fmt.Sprintf("Token{Type: %s}", t.Type)
	} else if t.Type == TokenTypeKeyword {
		return fmt.Sprintf("Token{Type: %s, Content: \"%s\", KeywordType: %s}", t.Type, t.Content, t.KeywordType)

	} else {
		return fmt.Sprintf("Token{Type: %s, Content: \"%s\"}", t.Type, t.Content)
	}
}

func (t Token) isDirective() bool {
	return t.Type == TokenTypeDirective
}

func (t Token) isIdentifier() bool {
	return t.Type == TokenTypeIdentifier
}

func (t Token) isLeftBrace() bool {
	return t.Type == TokenTypePunctuation && t.Content == "{"
}

func (t Token) isSingleLineComment() bool {
	return t.Type == TokenTypeSingleLineComment
}

func (t Token) isMultilineComment() bool {
	return t.Type == TokenTypeMultilineComment
}

func (t Token) isAbsent() bool {
	return t.Type == TokenTypeNone
}

func (t Token) isIncludeDirective() bool {
	return t.isDirective() && t.Content == "#include"
}

func (t Token) isPragmaDirective() bool {
	return t.isDirective() && t.Content == "#pragma"
}

func (t Token) isLeftParenthesis() bool {
	return t.Type == TokenTypePunctuation && t.Content == "("
}

func (t Token) isRightParenthesis() bool {
	return t.Type == TokenTypePunctuation && t.Content == ")"
}

func (t Token) isRightBrace() bool {
	return t.Type == TokenTypePunctuation && t.Content == "}"
}

func (t Token) isSemicolon() bool {
	return t.Type == TokenTypePunctuation && t.Content == ";"
}

func (t Token) canBeLeftOperand() bool {
	return t.Type == TokenTypeIdentifier ||
		t.Type == TokenTypeConstant ||
		(t.Type == TokenTypePunctuation && t.Content == ")")
}

func (t Token) canBePointerOperator() bool {
	return t.Type == TokenTypePunctuation && (t.Content == "&" || t.Content == "*")
}

func (t Token) isIncrDecrOperator() bool {
	return t.Type == TokenTypePunctuation && (t.Content == "++" || t.Content == "--")
}

func (t Token) isDotOperator() bool {
	return t.Type == TokenTypePunctuation && (t.Content == ".")
}

func (t Token) isArrowOperator() bool {
	return t.Type == TokenTypePunctuation && (t.Content == "->")
}

func (t Token) isStructOrUnion() bool {
	return t.Type == TokenTypeKeyword && (t.KeywordType == KeywordTypeStruct || t.KeywordType == KeywordTypeUnion)
}

func (t Token) isEnum() bool {
	return t.Type == TokenTypeKeyword && t.KeywordType == KeywordTypeEnum
}

func (t Token) isAssignment() bool {
	assignmentOps := []string{"=", "*=", "/=", "%=", "+=", "-=", "<<=", ">>=", "&=", "^=", "|="}

	return t.Type == TokenTypePunctuation && slices.Contains(assignmentOps, t.Content)
}

func (t Token) isComma() bool {

	return t.Type == TokenTypePunctuation && t.Content == ","
}

func (t Token) hasNewLines() bool {
	return t.Whitespace.NewLines > 0
}

func (t Token) isComment() bool {
	return t.isSingleLineComment() || t.isMultilineComment()
}

func (t Token) isStringizingOp() bool {
	return t.Type == TokenTypePunctuation && t.Content == "#"
}

func (t Token) isCharizingOp() bool {
	return t.Type == TokenTypePunctuation && t.Content == "#@"
}

func (t Token) isTokenPastingOp() bool {
	return t.Type == TokenTypePunctuation && t.Content == "##"
}

func (t Token) isLeftBracket() bool {
	return t.Type == TokenTypePunctuation && t.Content == "["
}

func (t Token) isRightBracket() bool {
	return t.Type == TokenTypePunctuation && t.Content == "]"
}

func (t Token) isNegation() bool {
	return t.Type == TokenTypePunctuation && (t.Content == "!" || t.Content == "~")
}

func (t Token) isSizeOf() bool {
	return t.Type == TokenTypeKeyword && t.KeywordType == KeywordTypeSizeof
}

func (t Token) isGreaterThanSign() bool {
	return t.Type == TokenTypePunctuation && t.Content == ">"
}

func (t Token) isLessThanSign() bool {
	return t.Type == TokenTypePunctuation && t.Content == "<"
}

func (t Token) isDo() bool {
	return t.Type == TokenTypeKeyword && t.KeywordType == KeywordTypeDo
}

func (t Token) isFor() bool {
	return t.Type == TokenTypeKeyword && t.KeywordType == KeywordTypeFor
}

func (t Token) hasEscapedLines() bool {
	return t.Whitespace.HasEscapedLines
}
func (t Token) hasUnescapedLines() bool {
	return t.Whitespace.HasUnescapedLines
}

func (t Token) isInvalid() bool {
	return t.Type == TokenTypeInvalid
}

func (t KeywordType) String() string {
	switch t {
	case KeywordTypeNone:
		return "KeywordTypeNone"
	case KeywordTypeAuto:
		return "KeywordTypeAuto"
	case KeywordTypeBreak:
		return "KeywordTypeBreak"
	case KeywordTypeCase:
		return "KeywordTypeCase"
	case KeywordTypeChar:
		return "KeywordTypeChar"
	case KeywordTypeConst:
		return "KeywordTypeConst"
	case KeywordTypeContinue:
		return "KeywordTypeContinue"
	case KeywordTypeDefault:
		return "KeywordTypeDefault"
	case KeywordTypeDo:
		return "KeywordTypeDo"
	case KeywordTypeDouble:
		return "KeywordTypeDouble"
	case KeywordTypeElse:
		return "KeywordTypeElse"
	case KeywordTypeEnum:
		return "KeywordTypeEnum"
	case KeywordTypeExtern:
		return "KeywordTypeExtern"
	case KeywordTypeFloat:
		return "KeywordTypeFloat"
	case KeywordTypeFor:
		return "KeywordTypeFor"
	case KeywordTypeGoto:
		return "KeywordTypeGoto"
	case KeywordTypeIf:
		return "KeywordTypeIf"
	case KeywordTypeInline:
		return "KeywordTypeInline"
	case KeywordTypeInt:
		return "KeywordTypeInt"
	case KeywordTypeLong:
		return "KeywordTypeLong"
	case KeywordTypeRegister:
		return "KeywordTypeRegister"
	case KeywordTypeRestrict:
		return "KeywordTypeRestrict"
	case KeywordTypeReturn:
		return "KeywordTypeReturn"
	case KeywordTypeShort:
		return "KeywordTypeShort"
	case KeywordTypeSigned:
		return "KeywordTypeSigned"
	case KeywordTypeSizeof:
		return "KeywordTypeSizeof"
	case KeywordTypeStatic:
		return "KeywordTypeStatic"
	case KeywordTypeStruct:
		return "KeywordTypeStruct"
	case KeywordTypeSwitch:
		return "KeywordTypeSwitch"
	case KeywordTypeTypedef:
		return "KeywordTypeTypedef"
	case KeywordTypeUnion:
		return "KeywordTypeUnion"
	case KeywordTypeUnsigned:
		return "KeywordTypeUnsigned"
	case KeywordTypeVoid:
		return "KeywordTypeVoid"
	case KeywordTypeVolatile:
		return "KeywordTypeVolatile"
	case KeywordTypeWhile:
		return "KeywordTypeWhile"
	case KeywordTypeAlignas:
		return "KeywordTypeAlignas"
	case KeywordTypeAlignof:
		return "KeywordTypeAlignof"
	case KeywordTypeAtomic:
		return "KeywordTypeAtomic"
	case KeywordTypeBool:
		return "KeywordTypeBool"
	case KeywordTypeComplex:
		return "KeywordTypeComplex"
	case KeywordTypeGeneric:
		return "KeywordTypeGeneric"
	case KeywordTypeImaginary:
		return "KeywordTypeImaginary"
	case KeywordTypeNoreturn:
		return "KeywordTypeNoreturn"
	case KeywordTypeStaticAssert:
		return "KeywordTypeStaticAssert"
	case KeywordTypeThreadLocal:
		return "KeywordTypeThreadLocal"
	case KeywordTypeAsm:
		return "KeywordTypeAsm"
	case KeywordTypeBased:
		return "KeywordTypeBased"
	case KeywordTypeCdecl:
		return "KeywordTypeCdecl"
	case KeywordTypeDeclspec:
		return "KeywordTypeDeclspec"
	case KeywordTypeExcept:
		return "KeywordTypeExcept"
	case KeywordTypeFastcall:
		return "KeywordTypeFastcall"
	case KeywordTypeFinally:
		return "KeywordTypeFinally"
	case KeywordTypeInt16:
		return "KeywordTypeInt16"
	case KeywordTypeInt32:
		return "KeywordTypeInt32"
	case KeywordTypeInt64:
		return "KeywordTypeInt64"
	case KeywordTypeInt8:
		return "KeywordTypeInt8"
	case KeywordTypeLeave:
		return "KeywordTypeLeave"
	case KeywordTypeStdcall:
		return "KeywordTypeStdcall"
	case KeywordTypeTry:
		return "KeywordTypeTry"
	default:
		panic(fmt.Sprintf("unknown keyword type %d", t))
	}

}

func (t TokenType) String() string {
	switch t {
	case TokenTypeNone:
		return "TokenTypeNone"
	case TokenTypeIdentifier:
		return "TokenTypeIdentifier"
	case TokenTypeKeyword:
		return "TokenTypeKeyword"
	case TokenTypeConstant:
		return "TokenTypeConstant"
	case TokenTypePunctuation:
		return "TokenTypePunctuation"
	case TokenTypeDirective:
		return "TokenTypeDirective"
	case TokenTypeSingleLineComment:
		return "TokenTypeSingleLineComment"
	case TokenTypeMultilineComment:
		return "TokenTypeMultilineComment"
	case TokenTypeInvalid:
		return "TokenTypeInvalid"
	default:
		panic("Invalid TokenType")
	}
}
