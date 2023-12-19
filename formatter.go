package main

import (
	"log"
	"slices"
	"strings"
)

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

type StructUnionEnum struct {
	Indent int
}

func format(input string) string {

	f := newFormatter(input)

	saved := *f

	structUnionOrEnum := false

	for f.parseToken() {

		if f.OutputColumn > 80 && !f.wrapping {
			*f = saved
			f.wrapping = true
			continue
		}

		f.formatToken()

		if f.startsFunctionArguments() {
			if f.IsDirective {
				f.formatFunctionCallOrMacro()
			} else {
				f.formatFunctionDecl()
			}
		}

		if f.Token.isLeftBrace() {
			if f.PreviousToken.isAssignment() {
				f.formatInitialiserList()
				continue
			} else if structUnionOrEnum {
				f.wrapping = false
				f.formatDeclarationBody()
				structUnionOrEnum = false
			} else {
				f.wrapping = false
				formatBlockBody(f)
				f.twoLinesOrEof()
				continue
			}
		}

		if f.Token.isStructUnionEnumKeyword() {
			structUnionOrEnum = true
		}

		if f.alwaysOneLine() {
			f.writeNewLines(1)
		} else if f.IsEndOfDirective || f.alwaysDefaultLines() {
			f.twoLinesOrEof()
		} else if f.wrapping && f.Token.hasNewLines() {
			f.wrap()
		} else if !f.neverWhitespace() &&
			!f.NextToken.isRightBrace() &&
			!f.Token.isLeftBrace() {
			f.writeString(" ")
		}

		if f.Token.isSemicolon() && !f.IsParenthesis() {
			saved = *f
			f.wrapping = false
			continue
		}
	}

	return f.Output.String()
}

func newFormatter(input string) *Formatter {
	formatter := Formatter{Input: input, Output: strings.Builder{}}

	return &formatter
}

func formatBlockBody(f *Formatter) {
	f.Indent++

	if f.NextToken.isRightBrace() {

		f.Indent--
	}

	f.writeNewLines(1)

	structUnionOrEnum := false

	saved := *f

	for f.parseToken() {

		if f.NextToken.isRightBrace() {
			f.Indent--
		}

		if f.Token.isStructUnionEnumKeyword() {
			structUnionOrEnum = true
		}

		f.formatToken()

		if f.startsFunctionArguments() {
			f.formatFunctionCallOrMacro()
		}

		if f.Token.isRightBrace() {
			return
		}

		if f.Token.isLeftBrace() {

			if f.PreviousToken.isAssignment() {
				f.formatInitialiserList()
				continue
			} else if structUnionOrEnum {
				f.formatDeclarationBody()
				structUnionOrEnum = false
			} else {
				isDoWhileLoop := f.PreviousToken.isDo()
				f.wrapping = false
				formatBlockBody(f)
				if isDoWhileLoop {
					f.writeString(" ")
				} else {
					f.oneOrTwoLines()
				}

				saved = *f
				continue
			}
		}

		if f.OutputColumn > 80 && !f.wrapping {
			*f = saved
			f.wrapping = true
			continue
		}

		if f.alwaysOneLine() || f.NextToken.isRightBrace() {
			f.writeNewLines(1)
		} else if f.alwaysDefaultLines() {
			f.oneOrTwoLines()
		} else if f.wrapping && f.Token.hasNewLines() {
			f.wrap()
		} else if !f.neverWhitespace() {
			f.writeString(" ")
		}

		if f.Token.isSemicolon() && !f.IsParenthesis() {
			f.wrapping = false
			saved = *f
		}
	}

	log.Fatal("Unclosed block")
}

func (f *Formatter) formatFunctionCallOrMacro() {
	saved := *f
	succeess := f.tryFormatFunctionArguments(true, false)

	if !succeess {
		*f = saved
		_ = f.tryFormatFunctionArguments(false, false)
	}
}

func (f *Formatter) formatFunctionDecl() {
	saved := *f
	if !f.tryFormatFunctionArguments(true, true) {
		*f = saved
		f.tryFormatFunctionArguments(false, true)
	}
}

func (f *Formatter) formatInitialiserList() {

	initialState := *f

	if !f.tryFormatInlineInitialiserList() {
		*f = initialState
		f.formatMultilineInitialiserList()
	}

}

func (f *Formatter) formatDeclarationBody() {
	f.Indent++

	f.writeNewLines(1)

	for f.parseToken() {
		f.formatToken()

		if f.NextToken.isRightBrace() {
			f.Indent--
		}

		if f.Token.isRightBrace() {
			return
		}

		if f.Token.isLeftBrace() {
			f.formatDeclarationBody()
		}

		if f.alwaysOneLine() || f.alwaysDefaultLines() {
			f.writeNewLines(1)
		} else if !f.neverWhitespace() &&
			!f.NextToken.isSemicolon() {
			f.writeString(" ")
		}
	}

	log.Fatal("Unclosed declaration braces")
}

func (f *Formatter) tryFormatInlineInitialiserList() bool {
	openBraces := 1

	for f.parseToken() {
		f.formatToken()

		if f.Token.isRightBrace() {
			openBraces--
		}

		if f.Token.isComma() && f.Token.hasNewLines() {
			return false
		}

		if f.Token.isComment() {
			return false
		}

		if f.Token.isLeftBrace() {

			if !f.tryFormatInlineInitialiserList() {
				return false
			}
			continue
		}

		if f.alwaysOneLine() || f.alwaysDefaultLines() {
			f.writeNewLines(1)
		} else if !f.neverWhitespace() &&
			!f.NextToken.isRightBrace() &&
			!f.Token.isRightBrace() {
			f.writeString(" ")
		}

		if openBraces == 0 {
			return true
		}
	}

	log.Fatal("Unclosed initialiser list")
	panic("unreachable")
}

func (f *Formatter) formatMultilineInitialiserList() {
	openBraces := 1

	f.Indent++

	f.writeNewLines(1)

	for f.parseToken() {

		f.formatToken()
		if f.NextToken.isRightBrace() {
			f.Indent--
		}

		if f.Token.isRightBrace() {
			openBraces--
		}

		if f.Token.isLeftBrace() {
			f.formatMultilineInitialiserList()
			continue
		}

		if f.alwaysOneLine() || f.NextToken.isRightBrace() ||
			(f.Token.isComma() && f.Token.hasNewLines()) || f.alwaysDefaultLines() {
			f.writeNewLines(1)
		} else if !f.neverWhitespace() &&
			!f.NextToken.isRightBrace() &&
			!f.Token.isRightBrace() {
			f.writeString(" ")
		}

		if openBraces == 0 {
			return
		}
	}

	log.Fatal("Unclosed initialiser list")
}

func (f *Formatter) tryFormatFunctionArguments(inline bool, isFunctionDecl bool) bool {
	commas := 0
	openParenthesis := 1
	if !inline {
		f.Indent++
		f.writeNewLines(1)
	}

	for f.parseToken() {

		topLevelComma := false

		if f.Token.isComma() && openParenthesis == 1 {
			topLevelComma = true
			commas++
		}

		if f.OutputColumn > 80 && commas > 0 {
			return false
		}

		f.formatToken()

		if f.Token.isRightParenthesis() {
			openParenthesis--
		}

		if f.Token.isLeftParenthesis() {
			openParenthesis++
		}

		if openParenthesis == 0 {
			return true
		}

		beforeLastParenthesis := openParenthesis == 1 && f.NextToken.isRightParenthesis() && !inline

		if beforeLastParenthesis {
			f.Indent--
		}

		if f.alwaysOneLine() || f.alwaysDefaultLines() || (topLevelComma && !inline) || beforeLastParenthesis {
			f.writeNewLines(1)
		} else if !f.neverWhitespace() &&
			!f.NextToken.isRightBrace() &&
			!f.Token.isRightBrace() {
			f.writeString(" ")
		}
	}

	log.Fatal("Unclosed function arguments")

	panic("unreachable")
}

func (f *Formatter) parseToken() bool {

	if f.Token.isAbsent() {
		_ = f.skipSpaceAndCountNewLines()

		f.Token = parseToken(f.Input)
		f.Input = f.Input[len(f.Token.Content):]
	} else {
		f.PreviousToken = f.Token
		f.Token = f.NextToken
	}

	f.IsEndOfDirective = false

	if f.Token.isDirective() {
		f.IsDirective = true
	}

	wasDirective := f.IsDirective

	if f.Token.isIncludeDirective() {
		f.IsIncludeDirective = true
	}

	if f.Token.isAssignment() {
		f.RightSideOfAssignment = true
	}

	if f.Token.isSemicolon() {
		f.RightSideOfAssignment = false
	}

	f.Token.Whitespace = f.skipSpaceAndCountNewLines()

	f.NextToken = parseToken(f.Input)
	f.Input = f.Input[len(f.NextToken.Content):]

	if f.Token.Whitespace.HasUnescapedLines || f.NextToken.isAbsent() {
		f.IsDirective = false
		f.IsIncludeDirective = false
	}

	if wasDirective && !f.IsDirective {
		f.IsEndOfDirective = true
	}

	if f.Token.isLeftParenthesis() {
		f.OpenParenthesis++
	}

	if f.Token.isRightParenthesis() {
		f.OpenParenthesis--

	}

	return !f.Token.isAbsent()
}

func (f *Formatter) skipSpaceAndCountNewLines() Whitespace {

	result := Whitespace{}

	for f.consumeSpace(&result) {
		result.HasSpace = true
	}

	return result

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

func (formatter *Formatter) writeString(str string) {
	formatter.Output.WriteString(str)
	formatter.OutputColumn += len(str)
}

func (formatter *Formatter) wrap() {
	formatter.Indent++
	formatter.writeNewLines(1)
	formatter.Indent--
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

func (formatter *Formatter) IsParenthesis() bool {
	return formatter.OpenParenthesis > 0
}

func (f *Formatter) isPointerOperator() bool {
	return f.Token.canBePointerOperator() && (!f.PreviousToken.canBeLeftOperand() || !f.RightSideOfAssignment)
}

func (f *Formatter) hasPostfixIncrDecr() bool {
	return f.NextToken.isIncrDecrOperator() && (f.Token.isIdentifier() || f.Token.isRightParenthesis())
}

func (f *Formatter) isPrefixIncrDecr() bool {
	return f.Token.isIncrDecrOperator() && (f.NextToken.isIdentifier() || f.NextToken.isLeftParenthesis())
}

func (f *Formatter) hasTrailingComment() bool {
	return f.NextToken.isSingleLineComment() && f.Token.Whitespace.NewLines == 0
}

func (f *Formatter) formatMultilineComment() {
	text := strings.TrimSpace(f.Token.Content[2 : len(f.Token.Content)-2])

	lines := strings.Split(text, "\n")
	f.writeString("/*")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		f.writeNewLines(1)
		if len(trimmed) > 0 {
			f.writeString("   ")
			f.writeString(trimmed)
		}
	}

	f.writeNewLines(1)
	f.writeString("*/")
}

func (f *Formatter) formatSingleLineComment() {
	text := strings.TrimSpace(f.Token.Content[2:])
	f.writeString("// ")
	f.writeString(text)
}

func (f *Formatter) formatToken() {

	if f.Token.isMultilineComment() {
		f.formatMultilineComment()
	} else if f.Token.isSingleLineComment() {
		f.formatSingleLineComment()
	} else {
		f.writeString(f.Token.Content)
	}
}

func (f *Formatter) startsFunctionArguments() bool {
	if !f.IsDirective {
		return f.PreviousToken.Type == TokenTypeIdentifier && f.Token.isLeftParenthesis()

	} else {
		return f.PreviousToken.Type == TokenTypeIdentifier && f.Token.isLeftParenthesis() && !f.PreviousToken.Whitespace.HasSpace
	}
}

func (f *Formatter) isFunctionName() bool {
	return f.Token.Type == TokenTypeIdentifier && f.NextToken.isLeftParenthesis() && (!f.IsDirective || !f.Token.Whitespace.HasSpace)
}

func (f *Formatter) neverWhitespace() bool {

	return f.NextToken.isSemicolon() ||
		f.Token.isLeftParenthesis() ||
		f.NextToken.isRightParenthesis() ||
		f.Token.isLeftBracket() ||
		f.NextToken.isLeftBracket() ||
		f.NextToken.isRightBracket() ||
		f.isPointerOperator() ||
		f.isFunctionName() ||
		f.hasPostfixIncrDecr() ||
		f.isPrefixIncrDecr() ||
		(f.Token.isIdentifier() && f.NextToken.isDotOperator()) ||
		f.Token.isDotOperator() ||
		f.Token.isArrowOperator() ||
		f.NextToken.isArrowOperator() ||
		f.NextToken.isComma() ||
		f.Token.isNegation() ||
		f.Token.isSizeOf() ||
		f.Token.isStringizingOp() ||
		f.Token.isCharizingOp() ||
		f.Token.isTokenPastingOp() ||
		f.NextToken.isTokenPastingOp() ||
		(f.IsIncludeDirective && ((f.NextToken.isGreaterThanSign()) || f.Token.isLessThanSign()))
}

func (f *Formatter) alwaysOneLine() bool {
	return f.NextToken.isAbsent() || f.Token.isComment()
}

func (f *Formatter) alwaysDefaultLines() bool {
	return (f.NextToken.isDirective() && !f.PreviousToken.isAbsent()) ||
		f.IsEndOfDirective ||
		f.NextToken.isMultilineComment() ||
		(f.Token.isSemicolon() && !f.IsParenthesis() && !f.hasTrailingComment())
}