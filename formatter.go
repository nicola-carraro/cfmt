package main

import (
	"fmt"
	"slices"
	"strings"
)

type Formatter struct {
	PreviousToken         Token
	Token                 Token
	NextToken             Token
	Indent                int
	TokenIndex            int
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
	IsPragmaDirective     bool
	IsEndOfInclude        bool
	IsEndOfPragma         bool
	RightSideOfAssignment bool
	AcceptStructOrUnion   bool
	AcceptEnum            bool
	IsForLoop             bool
	ForOpenParenthesis    int
	Nodes                 []Node
	LastNodeId            int
	PreviousNode          Node
	WrappingNode          int
	OpenBraces            int
}

type NodeType int

const (
	NodeTypeNone NodeType = iota
	NodeTypeMacroDef
	NodeTypeOtherDirective
	NodeTypeFunctionDef
	NodeTypeInvokation
	NodeTypeBlock
	NodeTypeInitialiserList
	NodeTypeStructOrUnion
	NodeTypeEnum
	NodeTypeOtherParenthesis
)

type Node struct {
	Type               NodeType
	Id                 int
	FirstToken         int
	LastToken          int
	InitialIndent      int
	InitialParenthesis int
	InitialBraces      int
}

type StructUnionEnum struct {
	Indent int
}

func (f *Formatter) wrapping() bool {
	return f.WrappingNode != 0
}

func format(input string) string {

	f := newFormatter(input)

	for f.parseToken() {

		//fmt.Println(f.PreviousToken, " ", f.Token, " ", f.NextToken)

		//	fmt.Println("wrapping ", f.wrapping())

		// fmt.Println(f.Node())
		// fmt.Println("comment or directive ", f.Token.isComment() || f.Token.isDirective())

		f.formatToken()

		if f.alwaysOneLine() {
			//fmt.Println(f.Token)
			f.writeNewLines(1)
		} else if f.isEndOfDirective() || f.alwaysDefaultLines() {
			f.writeDefaultLines()
		} else if f.wrapping() && f.Token.hasNewLines() {
			f.wrap()
		} else if !f.neverSpace() &&
			!f.NextToken.isRightBrace() &&
			!f.Token.isLeftBrace() {
			f.writeString(" ")
		}

		if f.Token.isSemicolon() && !f.IsParenthesis() && f.WrappingNode == f.Node().Id {
			f.WrappingNode = 0
			fmt.Println("RESET")
			continue
		}

	}

	return f.Output.String()
}

func newFormatter(input string) *Formatter {
	formatter := Formatter{Input: input, Output: strings.Builder{}}

	(&formatter).pushNode(NodeTypeNone)

	return &formatter
}

// func (f *Formatter) formatBlockBody() {
// 	wrapping := false
// 	f.Indent++

// 	if f.NextToken.isRightBrace() {

// 		f.Indent--
// 	}

// 	f.writeNewLines(1)

// 	saved := *f

// 	for f.parseToken() {

// 		f.formatToken()

// 		if f.startsFunctionArguments() {
// 			f.formatFunctionCall()
// 		}

// 		if f.NextToken.isRightBrace() {
// 			f.Indent--
// 		}

// 		if f.Token.isRightBrace() {
// 			return
// 		}

// 		if f.Token.isDefineDirective() {
// 			f.formatMacro()
// 		} else if f.Token.isLeftBrace() {

// 			if f.PreviousToken.isAssignment() {
// 				f.formatInitializerList()
// 				continue
// 			} else if f.AcceptStructOrUnion {
// 				f.formatStructOrUnion()
// 			} else if f.AcceptEnum {
// 				f.formatEnum()
// 			} else {
// 				isDoWhileLoop := f.PreviousToken.isDo()
// 				wrapping = false
// 				f.formatBlockBody()
// 				if isDoWhileLoop {
// 					f.writeString(" ")
// 				} else {
// 					f.oneOrTwoLines()
// 				}

// 				saved = *f
// 				continue
// 			}
// 		}

// 		if f.OutputColumn > 80 && !wrapping {
// 			*f = saved
// 			wrapping = true
// 			continue
// 		}

// 		if f.alwaysOneLine() || f.NextToken.isRightBrace() {
// 			f.writeNewLines(1)
// 		} else if f.alwaysDefaultLines() {
// 			f.writeDefaultLines()
// 		} else if wrapping && f.Token.hasNewLines() {
// 			f.wrap()
// 		} else if !f.neverSpace() {
// 			f.writeString(" ")
// 		}

// 		if f.Token.isSemicolon() && !f.IsParenthesis() {
// 			wrapping = false
// 			saved = *f
// 		}
// 	}

// 	log.Fatal("Unclosed block")
// }

// func (f *Formatter) formatMacro() {

// 	oldIndent := f.Indent

// 	f.Indent = 0

// 	f.writeString(" ")

// 	for f.parseToken() {
// 		if f.Token.hasEscapedLines() {
// 			if f.Token.isLeftBrace() || f.Token.isLeftParenthesis() {
// 				f.Indent++
// 			}

// 			if f.NextToken.isRightBrace() || f.NextToken.isRightParenthesis() {
// 				f.Indent--
// 			}
// 		}

// 		f.formatToken()

// 		if f.Token.hasUnescapedLines() {
// 			f.Indent = oldIndent
// 			f.IsDirective = false

// 			return
// 		}

// 		if f.alwaysOneLine() || f.Token.hasEscapedLines() {
// 			f.writeNewLines(1)
// 		} else if f.alwaysDefaultLines() {
// 			f.writeDefaultLines()
// 		} else if !f.neverSpace() {
// 			f.writeString(" ")
// 		}

// 	}
// }

// func (f *Formatter) formatFunctionCall() {
// 	saved := *f
// 	success := f.tryFormatFunctionArguments(true, false)

// 	if !success {
// 		*f = saved
// 		_ = f.tryFormatFunctionArguments(false, false)
// 	}
// }

// func (f *Formatter) formatFunctionDecl() {
// 	saved := *f
// 	if !f.tryFormatFunctionArguments(true, true) {
// 		*f = saved
// 		f.tryFormatFunctionArguments(false, true)
// 	}
// }

// func (f *Formatter) formatInitializerList() {

// 	initialState := *f

// 	if !f.tryFormatInlineInitializerList() {
// 		*f = initialState
// 		f.formatMultilineInitializerList()
// 	}

// }

// func (f *Formatter) formatStructOrUnion() {
// 	f.AcceptStructOrUnion = false
// 	f.Indent++

// 	f.writeNewLines(1)

// 	for f.parseToken() {
// 		f.formatToken()

// 		if f.NextToken.isRightBrace() {
// 			f.Indent--
// 		}

// 		if f.Token.isRightBrace() {
// 			return
// 		}

// 		if f.Token.isLeftBrace() {
// 			f.formatStructOrUnion()
// 		} else if f.Token.isDefineDirective() {
// 			f.formatMacro()
// 		}

// 		if f.alwaysOneLine() || f.alwaysDefaultLines() {
// 			f.writeDefaultLines()

// 		} else if !f.neverSpace() &&
// 			!f.NextToken.isSemicolon() {
// 			f.writeString(" ")
// 		}
// 	}

// 	log.Fatal("Unclosed declaration braces")
// }

// func (f *Formatter) formatEnum() {
// 	f.AcceptEnum = false
// 	f.Indent++

// 	f.writeNewLines(1)

// 	for f.parseToken() {
// 		f.formatToken()

// 		if f.NextToken.isRightBrace() {
// 			f.Indent--
// 		}

// 		if f.Token.isRightBrace() {
// 			return
// 		}

// 		if f.Token.isDefineDirective() {
// 			f.formatMacro()
// 		}

// 		if f.alwaysOneLine() || f.alwaysDefaultLines() || f.Token.isComma() || f.NextToken.isRightBrace() {
// 			f.writeDefaultLines()

// 		} else if !f.neverSpace() &&
// 			!f.NextToken.isSemicolon() {
// 			f.writeString(" ")
// 		}
// 	}

// 	log.Fatal("Unclosed declaration braces")
// }

// func (f *Formatter) tryFormatInlineInitializerList() bool {
// 	openBraces := 1

// 	for f.parseToken() {
// 		f.formatToken()

// 		if f.Token.isRightBrace() {
// 			openBraces--
// 		}

// 		if f.Token.isComma() && f.Token.hasNewLines() {
// 			return false
// 		}

// 		if f.Token.isComment() {
// 			return false
// 		}

// 		if f.Token.isLeftBrace() {

// 			if !f.tryFormatInlineInitializerList() {
// 				return false
// 			}
// 			continue
// 		} else if f.Token.isDefineDirective() {
// 			f.formatMacro()
// 		}

// 		if f.alwaysOneLine() || f.alwaysDefaultLines() {
// 			f.writeDefaultLines()

// 		} else if !f.neverSpace() &&
// 			!f.NextToken.isRightBrace() &&
// 			!f.Token.isRightBrace() {
// 			f.writeString(" ")
// 		}

// 		if openBraces == 0 {
// 			return true
// 		}
// 	}

// 	log.Fatal("Unclosed initializer list")
// 	panic("unreachable")
// }

// func (f *Formatter) formatMultilineInitializerList() {
// 	openBraces := 1

// 	f.Indent++

// 	f.writeNewLines(1)

// 	for f.parseToken() {

// 		f.formatToken()
// 		if f.NextToken.isRightBrace() {
// 			f.Indent--
// 		}

// 		if f.Token.isRightBrace() {
// 			openBraces--
// 		}

// 		if f.Token.isLeftBrace() {
// 			f.formatInitializerList()
// 			continue
// 		} else if f.Token.isDefineDirective() {
// 			f.formatMacro()
// 		}

// 		if f.alwaysOneLine() || f.NextToken.isRightBrace() ||
// 			(f.Token.isComma() && f.Token.hasNewLines()) || f.alwaysDefaultLines() {
// 			f.writeDefaultLines()

// 		} else if !f.neverSpace() &&
// 			!f.NextToken.isRightBrace() &&
// 			!f.Token.isRightBrace() {
// 			f.writeString(" ")
// 		}

// 		if openBraces == 0 {
// 			return
// 		}
// 	}

// 	log.Fatal("Unclosed initializer list")
// }

// func (f *Formatter) tryFormatFunctionArguments(inline bool, isFunctionDecl bool) bool {
// 	commas := 0
// 	openParenthesis := 1
// 	if !inline {
// 		f.Indent++
// 		f.writeNewLines(1)
// 	}

// 	newLines := f.Token.Whitespace.NewLines

// 	for f.parseToken() {
// 		newLines += f.Token.Whitespace.NewLines
// 		topLevelComma := false

// 		if f.Token.isComma() && openParenthesis == 1 {
// 			topLevelComma = true
// 			commas++
// 		}

// 		if inline &&
// 			((f.OutputColumn > 80 && commas > 0 && (isFunctionDecl || newLines > 0)) ||
// 				(f.Token.isComment() || f.Token.isDirective())) {
// 			return false
// 		}

// 		f.formatToken()

// 		if f.Token.isRightParenthesis() {
// 			openParenthesis--
// 		}

// 		if f.Token.isLeftParenthesis() {
// 			openParenthesis++
// 		}

// 		if f.Token.isDefineDirective() {
// 			f.formatMacro()
// 		} else if f.Token.isLeftBrace() {
// 			if inline {
// 				return false
// 			}
// 			f.formatBlockBody()

// 		}

// 		if openParenthesis == 0 {
// 			return true
// 		}

// 		beforeLastParenthesis := openParenthesis == 1 && f.NextToken.isRightParenthesis() && !inline

// 		if beforeLastParenthesis {
// 			f.Indent--
// 		}

// 		if f.alwaysOneLine() || f.alwaysDefaultLines() || (topLevelComma && !inline) || beforeLastParenthesis {
// 			f.writeDefaultLines()

// 		} else if !f.neverSpace() &&
// 			!f.NextToken.isRightBrace() &&
// 			!f.Token.isRightBrace() {
// 			f.writeString(" ")
// 		}
// 	}

// 	log.Fatal("Unclosed function arguments")

// 	panic("unreachable")
// }

func (f *Formatter) parseToken() bool {

	if f.Token.isAbsent() {
		_ = f.skipSpaceAndCountNewLines()

		f.Token = parseToken(f.Input)
		f.Input = f.Input[len(f.Token.Content):]
		f.TokenIndex++
	} else {
		f.PreviousToken = f.Token
		f.Token = f.NextToken
		f.TokenIndex++

	}

	f.IsEndOfInclude = false
	f.IsEndOfPragma = false

	if f.Token.isDirective() {
		f.IsDirective = true
	}
	wasInclude := f.IsIncludeDirective
	wasPragma := f.IsPragmaDirective

	if f.Token.isIncludeDirective() {
		f.IsIncludeDirective = true
	}

	if f.Token.isPragmaDirective() {
		f.IsPragmaDirective = true
	}

	if f.Token.isStructOrUnion() {
		f.AcceptStructOrUnion = true
	}

	if f.Token.isEnum() {
		f.AcceptEnum = true
	}

	if f.startsFunctionArguments() {
		f.AcceptStructOrUnion = false
		f.AcceptEnum = false
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

	if !f.Node().isDirective() {

		if f.Token.isDefineDirective() {
			f.pushNode(NodeTypeMacroDef)
		} else if f.Token.isDirective() {
			f.pushNode(NodeTypeOtherDirective)

		} else if f.startsFunctionArguments() {
			if f.IsDirective {
				f.pushNode(NodeTypeInvokation)

			} else {
				f.pushNode(NodeTypeFunctionDef)
			}
		} else if f.Token.isLeftParenthesis() {
			f.pushNode(NodeTypeOtherParenthesis)
		} else if f.Token.isLeftBrace() {
			if f.PreviousToken.isAssignment() {
				f.pushNode(NodeTypeInitialiserList)
			} else if f.AcceptStructOrUnion {
				f.pushNode(NodeTypeStructOrUnion)

			} else if f.AcceptEnum {
				f.pushNode(NodeTypeEnum)

			} else {
				f.pushNode(NodeTypeBlock)

			}
		}
	}

	if (f.Node().Type == NodeTypeBlock ||
		f.Node().Type == NodeTypeInitialiserList ||
		f.Node().Type == NodeTypeEnum ||
		f.Node().Type == NodeTypeStructOrUnion) &&
		f.Token.isRightBrace() {
		f.popNode()
	}

	if (f.Node().Type == NodeTypeInvokation ||
		f.Node().Type == NodeTypeFunctionDef ||
		f.Node().Type == NodeTypeOtherParenthesis) &&
		f.Token.isRightParenthesis() {
		f.popNode()
	}

	if f.Node().isDirective() {
		if f.Token.hasEscapedLines() {
			if f.Token.isLeftBrace() || f.Token.isLeftParenthesis() {
				f.Indent++
			}

			if f.NextToken.isRightBrace() || f.NextToken.isRightParenthesis() {
				f.Indent--
			}
		}
	}

	if (f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && f.isNodeStart() {
		f.Indent++

	}

	if (f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && f.NextToken.isRightBrace() {
		f.Indent--
	}

	if f.Node().isDirective() &&
		f.Token.hasUnescapedLines() {
		f.popNode()
	}

	if (f.Node().Type == NodeTypeFunctionDef || f.Node().Type == NodeTypeInvokation) && f.NextToken.isRightParenthesis() {
		f.Indent--
		//fmt.Println(f.Token)
	}

	if f.Token.Whitespace.HasUnescapedLines || f.NextToken.isAbsent() {
		f.IsDirective = false
		f.IsIncludeDirective = false
		f.IsPragmaDirective = false
	}

	if wasInclude && !f.IsIncludeDirective {
		f.IsEndOfInclude = true
	}

	if wasPragma && !f.IsPragmaDirective {
		f.IsEndOfPragma = true
	}

	if f.Token.isLeftParenthesis() {
		f.OpenParenthesis++
	}

	if f.Token.isRightParenthesis() {
		f.OpenParenthesis--

	}

	if f.Token.isLeftBrace() {
		f.OpenBraces++
	}

	if f.Token.isRightBrace() {
		f.OpenBraces--

	}

	if f.Token.isFor() && f.NextToken.isLeftParenthesis() {
		f.IsForLoop = true
		f.ForOpenParenthesis = f.OpenParenthesis
	} else if f.ForOpenParenthesis == f.OpenParenthesis {
		f.IsForLoop = false
		f.ForOpenParenthesis = 0
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
				Whitespace.HasEscapedLines = true
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
			formatter.IsPragmaDirective = false
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

func (f *Formatter) isEndOfDirective() bool {
	return f.PreviousNode.isDirective() && f.PreviousNode.LastToken == f.TokenIndex
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

func (f *Formatter) writeDefaultLines() {
	switch f.Node().Type {
	case NodeTypeNone:
		f.twoLinesOrEof()
	case NodeTypeMacroDef, NodeTypeInvokation, NodeTypeInitialiserList, NodeTypeStructOrUnion, NodeTypeEnum:
		f.writeNewLines(1)
	case NodeTypeFunctionDef, NodeTypeBlock:
		f.oneOrTwoLines()
	default:
		panic("unreacheable")

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

func (f *Formatter) neverSpace() bool {

	return f.NextToken.isSemicolon() ||
		f.Token.isLeftParenthesis() ||
		f.NextToken.isRightParenthesis() ||
		f.Token.isLeftBrace() ||
		f.NextToken.isRightBrace() ||
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

	return f.NextToken.isAbsent() ||
		(f.Token.isComment() && (f.PreviousToken.hasNewLines() || f.PreviousToken.isAbsent())) ||
		(f.IsEndOfInclude && f.NextToken.isIncludeDirective()) ||
		(f.IsEndOfPragma && f.NextToken.isPragmaDirective()) ||
		(f.Node().Type == NodeTypeInvokation || f.Node().Type == NodeTypeFunctionDef) && f.WrappingNode == f.Node().Id && (f.Token.isLeftParenthesis() || f.Token.isComma() || f.NextToken.isRightParenthesis()) ||
		((f.Node().isStructOrUnion() || f.Node().isBlock()) && f.Token.isSemicolon()) ||
		((f.Node().isEnum()) && f.Token.isComma()) ||
		((f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && f.isNodeStart() || f.NextToken.isRightBrace())

}

func (f *Formatter) alwaysDefaultLines() bool {
	return (f.NextToken.isDirective() && !f.PreviousToken.isAbsent()) ||
		f.isEndOfDirective() ||
		(f.Token.isComment() && !f.PreviousToken.hasNewLines() && !f.PreviousToken.isAbsent()) ||
		f.NextToken.isMultilineComment() ||
		(f.Token.isSemicolon() && !f.IsForLoop && !f.hasTrailingComment()) ||
		(f.Node().isDirective() && f.Token.hasEscapedLines())
}

func (f *Formatter) Node() Node {
	return f.Nodes[len(f.Nodes)-1]
}

func (f *Formatter) pushNode(t NodeType) {
	f.LastNodeId++
	node := Node{
		Type:               t,
		Id:                 f.LastNodeId,
		FirstToken:         f.TokenIndex,
		InitialIndent:      f.Indent,
		InitialParenthesis: f.OpenParenthesis,
		InitialBraces:      f.OpenBraces,
	}
	f.Nodes = append(f.Nodes, node)
	if node.Type == NodeTypeFunctionDef || node.Type == NodeTypeInvokation {
		f.Indent++
	}
	//fmt.Printf("Push %s, %s\n", node, f.Token)
}

func (f *Formatter) popNode() {
	//fmt.Printf("Pop %s, %s \n", f.Node(), f.Token)

	f.PreviousNode = f.Node()
	f.PreviousNode.LastToken = f.TokenIndex
	if f.WrappingNode == f.Node().Id {
		//fmt.Println("POP NODE")
		f.WrappingNode = 0
	}
	if f.Node().isDirective() {
		f.Indent = f.Node().InitialIndent
	}
	f.Nodes = f.Nodes[:len(f.Nodes)-1]

}

func (f *Formatter) isNodeStart() bool {
	return f.TokenIndex == f.Node().FirstToken

}

func (t NodeType) String() string {
	switch t {
	case NodeTypeNone:
		return "NodeTypeNone"
	case NodeTypeMacroDef:
		return "NodeTypeMacroDef"
	case NodeTypeOtherDirective:
		return "NodeTypeOtherDirective"
	case
		NodeTypeFunctionDef:
		return "NodeTypeFunctionDef"
	case
		NodeTypeInvokation:
		return "NodeTypeInvokation"
	case NodeTypeBlock:
		return "NodeTypeBlock"
	case NodeTypeInitialiserList:
		return "NodeTypeInitialiserList"
	case NodeTypeStructOrUnion:
		return "NodeTypeStructOrUnion"
	case NodeTypeEnum:
		return "NodeTypeEnum"
	default:
		panic("Unexpected node type")
	}
}

func (n Node) String() string {
	return fmt.Sprintf("Node{Type: %s, Id: %d}", n.Type, n.Id)
}

func (n Node) isDirective() bool {
	return n.Type == NodeTypeMacroDef || n.Type == NodeTypeOtherDirective
}

func (n Node) isStructOrUnion() bool {
	return n.Type == NodeTypeStructOrUnion
}

func (n Node) isBlock() bool {
	return n.Type == NodeTypeBlock
}

func (n Node) isEnum() bool {
	return n.Type == NodeTypeEnum
}
