package main

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
)

type StructUnionEnum struct {
	Indent int
}

type Formatter struct {
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
	OpenParenthesis       int
	IsDirective           bool
	IsIncludeDirective    bool
	IsEndOfDirective      bool
	RightSideOfAssignment bool
	wrapping              bool
}

func (formatter *Formatter) consumeSpace(Whitespace *Whitespace) bool {

	newLineInDirective := []string{"\\\r\n", "\\\n"}

	if formatter.IsDirective {
		for _, nl := range newLineInDirective {
			if formatter.IsDirective && strings.HasPrefix(formatter.Input, nl) {
				formatter.Input = formatter.Input[len(nl):]
				Whitespace.NewLines++
				return true
			}
		}

	}

	r, size := peakRune(formatter.Input)

	if r == '\n' {
		formatter.Input = formatter.Input[size:]
		Whitespace.NewLines++
		Whitespace.HasUnescapedLines = true

		if formatter.IsDirective {
			formatter.IsDirective = false
			formatter.IsIncludeDirective = false
		}
		return true
	}

	otherSpaces := []rune{' ', '\t', '\r', '\v', '\f'}

	if slices.Contains(otherSpaces, r) {
		formatter.Input = formatter.Input[size:]
		return true

	}

	return false
}

func (formatter *Formatter) parseToken() bool {

	if formatter.Token.isAbsent() {
		_ = skipSpaceAndCountNewLines(formatter)

		formatter.Token = parseToken(formatter.Input)
		formatter.Input = formatter.Input[len(formatter.Token.Content):]
	} else {
		formatter.PreviousToken = formatter.Token
		formatter.Token = formatter.NextToken
	}

	formatter.IsEndOfDirective = false

	if formatter.Token.isDirective() {
		formatter.IsDirective = true
	}

	wasDirective := formatter.IsDirective

	if formatter.Token.isIncludeDirective() {
		formatter.IsIncludeDirective = true
	}

	if formatter.Token.isAssignment() {
		formatter.RightSideOfAssignment = true
	}

	if formatter.Token.isSemicolon() {
		formatter.RightSideOfAssignment = false
	}

	formatter.Token.Whitespace = skipSpaceAndCountNewLines(formatter)

	formatter.NextToken = parseToken(formatter.Input)
	formatter.Input = formatter.Input[len(formatter.NextToken.Content):]

	if formatter.Token.Whitespace.HasUnescapedLines || formatter.NextToken.isAbsent() {
		formatter.IsDirective = false
		formatter.IsIncludeDirective = false
	}

	if wasDirective && !formatter.IsDirective {
		formatter.IsEndOfDirective = true
	}

	if formatter.Token.isLeftParenthesis() {
		formatter.OpenParenthesis++
	}

	if formatter.Token.isRightParenthesis() {
		formatter.OpenParenthesis--

	}

	return !formatter.Token.isAbsent()
}

func (formatter *Formatter) IsParenthesis() bool {
	return formatter.OpenParenthesis > 0
}

func skipSpaceAndCountNewLines(formatter *Formatter) Whitespace {

	result := Whitespace{}

	for formatter.consumeSpace(&result) {
		result.HasSpace = true
	}

	return result

}

func newParser(input string) *Formatter {
	formatter := Formatter{Input: input, Output: strings.Builder{}}

	return &formatter
}

func (formatter *Formatter) writeString(str string) {
	formatter.Output.WriteString(str)
	formatter.OutputColumn += len(str)
}

func (formatter *Formatter) writeNewLines(lines int) {
	const newLine = "\n"

	for line := 0; line < lines; line++ {

		if formatter.IsDirective {
			formatter.writeString("\\")
		}
		formatter.writeString(newLine)
		formatter.OutputColumn = 0
		formatter.OutputLine++
	}
	const indentation = "    "
	if !formatter.NextToken.isDirective() {
		for indentLevel := 0; indentLevel < formatter.Indent; indentLevel++ {
			formatter.writeString(indentation)
		}
	}

}

func tryFormatInlineInitialiserList(formatter *Formatter) bool {
	openBraces := 1

	for formatter.parseToken() {
		formatter.formatToken()

		if formatter.Token.isRightBrace() {
			openBraces--
		}

		if formatter.Token.isComma() && formatter.Token.hasNewLines() {
			return false
		}

		if formatter.Token.isComment() {
			return false
		}

		if formatter.Token.isLeftBrace() {

			if !tryFormatInlineInitialiserList(formatter) {
				return false
			}
			continue
		}

		if formatter.alwaysOneLine() || formatter.alwaysDefaultLines() {
			formatter.writeNewLines(1)
		} else if !neverWhitespace(formatter) &&
			!formatter.NextToken.isRightBrace() &&
			!formatter.Token.isRightBrace() {
			formatter.writeString(" ")
		}

		if openBraces == 0 {
			return true
		}
	}

	log.Fatal("Unclosed initialiser list")
	panic("unreachable")
}

func formatMultilineInitialiserList(formatter *Formatter) {
	openBraces := 1

	formatter.Indent++

	formatter.writeNewLines(1)

	for formatter.parseToken() {

		formatter.formatToken()
		if formatter.NextToken.isRightBrace() {
			formatter.Indent--
		}

		if formatter.Token.isRightBrace() {
			openBraces--
		}

		if formatter.Token.isLeftBrace() {

			formatMultilineInitialiserList(formatter)

			continue
		}

		if formatter.alwaysOneLine() || formatter.NextToken.isRightBrace() ||
			(formatter.Token.isComma() && formatter.Token.hasNewLines()) || formatter.alwaysDefaultLines() {
			formatter.writeNewLines(1)
		} else if !neverWhitespace(formatter) &&
			!formatter.NextToken.isRightBrace() &&
			!formatter.Token.isRightBrace() {
			formatter.writeString(" ")
		}

		if openBraces == 0 {
			return
		}
	}

	log.Fatal("Unclosed initialiser list")
}

func formatFunctionDecl(formatter *Formatter) {
	saved := *formatter
	if !tryFormatFunctionArguments(formatter, true, true) {
		*formatter = saved
		tryFormatFunctionArguments(formatter, false, true)
	}
}

func formatInitialiserList(formatter *Formatter) {

	initialState := *formatter

	if !tryFormatInlineInitialiserList(formatter) {
		*formatter = initialState
		formatMultilineInitialiserList(formatter)
	}

}

func (formatter *Formatter) oneOrTwoLines() {
	if formatter.Token.Whitespace.NewLines <= 1 || formatter.NextToken.isRightBrace() {
		formatter.writeNewLines(1)

	} else {
		formatter.writeNewLines(2)
	}
}

func (formatter *Formatter) twoLinesOrEof() {
	if formatter.NextToken.isAbsent() {
		formatter.writeNewLines(1)
	} else {
		formatter.writeNewLines(2)
	}
}

func isPointerOperator(formatter *Formatter) bool {
	return formatter.Token.canBePointerOperator() && (!formatter.PreviousToken.canBeLeftOperand() || !formatter.RightSideOfAssignment)
}

func hasPostfixIncrDecr(formatter *Formatter) bool {
	return formatter.NextToken.isIncrDecrOperator() && (formatter.Token.isIdentifier() || formatter.Token.isRightParenthesis())
}

func isPrefixIncrDecr(formatter *Formatter) bool {
	return formatter.Token.isIncrDecrOperator() && (formatter.NextToken.isIdentifier() || formatter.NextToken.isLeftParenthesis())
}

func (formatter *Formatter) hasTrailingComment() bool {
	return formatter.NextToken.isSingleLineComment() && formatter.Token.Whitespace.NewLines == 0
}

func (formatter *Formatter) wrap() {
	formatter.Indent++
	formatter.writeNewLines(1)
	formatter.Indent--
}

func tryFormatFunctionArguments(formatter *Formatter, inline bool, isFunctionDecl bool) bool {
	commas := 0
	openParenthesis := 1
	if !inline {
		formatter.Indent++
		formatter.writeNewLines(1)
	}

	for formatter.parseToken() {

		topLevelComma := false

		if formatter.Token.isComma() && openParenthesis == 1 {
			topLevelComma = true
			commas++
		}

		if formatter.OutputColumn > 80 && commas > 0 {
			return false
		}

		formatter.formatToken()

		if formatter.Token.isRightParenthesis() {
			openParenthesis--
		}

		if formatter.Token.isLeftParenthesis() {
			openParenthesis++
		}

		if openParenthesis == 0 {
			return true
		}

		beforeLastParenthesis := openParenthesis == 1 && formatter.NextToken.isRightParenthesis() && !inline

		if beforeLastParenthesis {
			formatter.Indent--
		}

		if formatter.alwaysOneLine() || formatter.alwaysDefaultLines() || (topLevelComma && !inline) || beforeLastParenthesis {
			formatter.writeNewLines(1)
		} else if !neverWhitespace(formatter) &&
			!formatter.NextToken.isRightBrace() &&
			!formatter.Token.isRightBrace() {
			formatter.writeString(" ")
		}
	}

	log.Fatal("Unclosed function arguments")

	panic("unreachable")
}

func formatFunctionCallOrMacro(formatter *Formatter) {
	saved := *formatter
	succeess := tryFormatFunctionArguments(formatter, true, false)

	if !succeess {
		*formatter = saved
		_ = tryFormatFunctionArguments(formatter, false, false)
	}
}

func isDo(token Token) bool {
	return token.Type == TokenTypeKeyword && token.Content == "do"
}

func formatBlockBody(formatter *Formatter) {
	formatter.Indent++

	if formatter.NextToken.isRightBrace() {

		formatter.Indent--
	}

	formatter.writeNewLines(1)

	structUnionOrEnum := false

	saved := *formatter

	for formatter.parseToken() {

		if formatter.NextToken.isRightBrace() {
			formatter.Indent--
		}

		if formatter.Token.isStructUnionEnumKeyword() {
			structUnionOrEnum = true
		}

		formatter.formatToken()

		if startsFunctionArguments(formatter) {
			formatFunctionCallOrMacro(formatter)
		}

		if formatter.Token.isRightBrace() {
			return
		}

		if formatter.Token.isLeftBrace() {

			if formatter.PreviousToken.isAssignment() {
				formatInitialiserList(formatter)
				continue
			} else if structUnionOrEnum {
				formatDeclarationBody(formatter)
				structUnionOrEnum = false
			} else {
				isDoWhileLoop := isDo(formatter.PreviousToken)
				formatter.wrapping = false
				formatBlockBody(formatter)
				if isDoWhileLoop {
					formatter.writeString(" ")
				} else {
					formatter.oneOrTwoLines()
				}

				saved = *formatter
				continue
			}
		}

		if formatter.OutputColumn > 80 && !formatter.wrapping {
			*formatter = saved
			formatter.wrapping = true
			continue
		}

		if formatter.alwaysOneLine() || formatter.NextToken.isRightBrace() {
			formatter.writeNewLines(1)
		} else if formatter.alwaysDefaultLines() {
			formatter.oneOrTwoLines()
		} else if formatter.wrapping && formatter.Token.hasNewLines() {
			formatter.wrap()
		} else if !neverWhitespace(formatter) {
			formatter.writeString(" ")
		}

		if formatter.Token.isSemicolon() && !formatter.IsParenthesis() {
			formatter.wrapping = false
			saved = *formatter
		}
	}

	log.Fatal("Unclosed block")
}

func (formatter *Formatter) formatMultilineComment() {
	text := strings.TrimSpace(formatter.Token.Content[2 : len(formatter.Token.Content)-2])

	lines := strings.Split(text, "\n")
	formatter.writeString("/*")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		formatter.writeNewLines(1)
		if len(trimmed) > 0 {
			formatter.writeString("   ")
			formatter.writeString(trimmed)
		}
	}

	formatter.writeNewLines(1)
	formatter.writeString("*/")
}

func (formatter *Formatter) formatSingleLineComment() {
	text := strings.TrimSpace(formatter.Token.Content[2:])
	formatter.writeString("// ")
	formatter.writeString(text)
}

func (formatter *Formatter) formatToken() {

	if formatter.Token.isMultilineComment() {
		formatter.formatMultilineComment()
	} else if formatter.Token.isSingleLineComment() {
		formatter.formatSingleLineComment()
	} else {
		formatter.writeString(formatter.Token.Content)
	}
}

func formatDeclarationBody(formatter *Formatter) {
	formatter.Indent++

	formatter.writeNewLines(1)

	for formatter.parseToken() {
		formatter.formatToken()

		if formatter.NextToken.isRightBrace() {
			formatter.Indent--
		}

		if formatter.Token.isRightBrace() {
			return
		}

		if formatter.Token.isLeftBrace() {
			formatDeclarationBody(formatter)
		}

		if formatter.alwaysOneLine() || formatter.alwaysDefaultLines() {
			formatter.writeNewLines(1)
		} else if !neverWhitespace(formatter) &&
			!formatter.NextToken.isSemicolon() {
			formatter.writeString(" ")
		}
	}

	log.Fatal("Unclosed declaration braces")
}

func startsFunctionArguments(formatter *Formatter) bool {
	if !formatter.IsDirective {
		return formatter.PreviousToken.Type == TokenTypeIdentifier && formatter.Token.isLeftParenthesis()

	} else {
		return formatter.PreviousToken.Type == TokenTypeIdentifier && formatter.Token.isLeftParenthesis() && !formatter.PreviousToken.Whitespace.HasSpace
	}
}

func isFunctionName(formatter *Formatter) bool {
	return formatter.Token.Type == TokenTypeIdentifier && formatter.NextToken.isLeftParenthesis() && (!formatter.IsDirective || !formatter.Token.Whitespace.HasSpace)
}

func neverWhitespace(formatter *Formatter) bool {

	return formatter.NextToken.isSemicolon() ||
		formatter.Token.isLeftParenthesis() ||
		formatter.NextToken.isRightParenthesis() ||
		formatter.Token.isLeftBracket() ||
		formatter.NextToken.isLeftBracket() ||
		formatter.NextToken.isRightBracket() ||
		isPointerOperator(formatter) ||
		isFunctionName(formatter) ||
		hasPostfixIncrDecr(formatter) ||
		isPrefixIncrDecr(formatter) ||
		(formatter.Token.isIdentifier() && formatter.NextToken.isDotOperator()) ||
		formatter.Token.isDotOperator() ||
		formatter.Token.isArrowOperator() ||
		formatter.NextToken.isArrowOperator() ||
		formatter.NextToken.isComma() ||
		formatter.Token.isNegation() ||
		formatter.Token.isSizeOf() ||
		formatter.Token.isStringizingOp() ||
		formatter.Token.isCharizingOp() ||
		formatter.Token.isTokenPastingOp() ||
		formatter.NextToken.isTokenPastingOp() ||
		(formatter.IsIncludeDirective && ((formatter.NextToken.isGreaterThanSign()) || formatter.Token.isLessThanSign()))
}

func (formatter *Formatter) alwaysOneLine() bool {
	return formatter.NextToken.isAbsent() || formatter.Token.isComment()
}

func (formatter *Formatter) alwaysDefaultLines() bool {
	return (formatter.NextToken.isDirective() && !formatter.PreviousToken.isAbsent()) ||
		formatter.IsEndOfDirective ||
		formatter.NextToken.isMultilineComment() ||
		(formatter.Token.isSemicolon() && !formatter.IsParenthesis() && !formatter.hasTrailingComment())
}

func format(input string) string {

	formatter := newParser(input)

	saved := *formatter

	structUnionOrEnum := false

	for formatter.parseToken() {

		if formatter.OutputColumn > 80 && !formatter.wrapping {
			*formatter = saved
			formatter.wrapping = true
			continue
		}

		formatter.formatToken()

		if startsFunctionArguments(formatter) {
			if formatter.IsDirective {
				formatFunctionCallOrMacro(formatter)
			} else {
				formatFunctionDecl(formatter)

			}
		}

		if formatter.Token.isLeftBrace() {
			if formatter.PreviousToken.isAssignment() {
				formatInitialiserList(formatter)
				continue
			} else if structUnionOrEnum {
				formatter.wrapping = false
				formatDeclarationBody(formatter)
				structUnionOrEnum = false

			} else {
				formatter.wrapping = false

				formatBlockBody(formatter)
				formatter.twoLinesOrEof()
				continue
			}
		}

		if formatter.Token.isStructUnionEnumKeyword() {
			structUnionOrEnum = true
		}

		if formatter.alwaysOneLine() {
			formatter.writeNewLines(1)
		} else if formatter.IsEndOfDirective || formatter.alwaysDefaultLines() {
			formatter.twoLinesOrEof()
		} else if formatter.wrapping && formatter.Token.hasNewLines() {
			formatter.wrap()
		} else if !neverWhitespace(formatter) &&
			!formatter.NextToken.isRightBrace() &&
			!formatter.Token.isLeftBrace() {
			formatter.writeString(" ")
		}

		if formatter.Token.isSemicolon() && !formatter.IsParenthesis() {
			saved = *formatter
			formatter.wrapping = false
			continue
		}
	}

	return formatter.Output.String()
}

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <path>\n", os.Args[0])
	}

	for i := 1; i < len(os.Args); i++ {
		path := os.Args[i]

		data, err := os.ReadFile(path)

		if err != nil {
			log.Fatalf("Error reading %s: %s", path, err)
		}

		text := string(data)

		formattedText := format(text)

		fmt.Print(formattedText)

		// os.WriteFile(path, []byte(formattedText), 0600)

		// if err != nil {
		// 	log.Fatalf("Error writing %s: %s", path, err)
		// }

	}

}
