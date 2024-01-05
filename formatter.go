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
	Input                 *string
	Output                []byte
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
	Wrapping              bool
	Tokens                *[]Token
	WrappingStrategy      WrappingStrategy
}

type NodeType int

const (
	NodeTypeNone NodeType = iota
	NodeTypeTopLevel
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

type WrappingStrategy int

const (
	WrappingStrategyNone WrappingStrategy = iota
	WrappingStrategyComma
	WrappingStrategyLineBreak
	WrappingStrategyLineBreakAfterComma
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

	f := Formatter{Input: &input, Tokens: new([]Token)}

	(&f).pushNode(NodeTypeTopLevel)

	saved := f

	for f.parseToken() {
		f.formatToken()

		if !f.Wrapping && f.OutputColumn > 80 {
			f = saved
			f.Wrapping = true
			continue
		}

		if f.isBlockStart() || (!f.Node().isStructOrUnion() && f.Token.isSemicolon()) {
			f.Wrapping = false
			f.WrappingNode = 0
			saved = f
		}

		if f.alwaysOneLine() {
			f.writeNewLines(1)
		} else if f.isEndOfDirective() || f.alwaysDefaultLines() {
			f.writeDefaultLines()
		} else if !f.neverSpace() &&
			!f.NextToken.isRightBrace() &&
			!f.Token.isLeftBrace() {
			f.writeString(" ")
		}

	}

	return string(f.Output)
}

func (f *Formatter) getToken(index int) Token {
	if len(*f.Tokens) <= index {
		token := parseToken(*f.Input)
		*f.Input = (*f.Input)[len(token.Content):]
		token.Whitespace = f.skipSpaceAndCountNewLines()
		(*f.Tokens) = append(*f.Tokens, token)
	}

	return (*f.Tokens)[index]
}

func (f *Formatter) parseToken() bool {

	if f.Token.isAbsent() {
		_ = f.skipSpaceAndCountNewLines()

		f.Token = f.getToken(f.TokenIndex)

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

	f.NextToken = f.getToken(f.TokenIndex)

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
			if f.PreviousToken.isAssignment() || f.Node().isInitialiserList() {
				f.pushNode(NodeTypeInitialiserList)
			} else if f.AcceptStructOrUnion || f.Node().isStructOrUnion() {
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

	if f.Wrapping && f.WrappingStrategy == WrappingStrategyNone {
		if f.isFunctionStart() {
			f.WrappingStrategy = WrappingStrategyComma
			f.WrappingNode = f.Node().Id
		}

		if f.isInitialiserListStart() {
			f.WrappingStrategy = WrappingStrategyLineBreakAfterComma
			f.WrappingNode = f.Node().Id
		}
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

	if f.shouldIncreaseIndent() {
		f.Indent++

	}

	if f.shouldDecreaseIndent() {
		f.Indent--
	}

	if f.Node().isDirective() &&
		f.Token.hasUnescapedLines() {
		f.popNode()
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

func (f *Formatter) shouldIncreaseIndent() bool {
	return ((f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && f.isNodeStart()) ||
		(f.isWrappingNode() && (f.isInitialiserListStart() || f.isInvokationStart() || f.isFunctionDefStart()))
}

func (f *Formatter) shouldDecreaseIndent() bool {
	return ((f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && f.NextToken.isRightBrace()) ||
		(f.isWrappingNode() && f.Node().isInitialiserList() && f.NextToken.isRightBrace())
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
			if formatter.IsDirective && strings.HasPrefix((*formatter.Input), nl) {
				*formatter.Input = (*formatter.Input)[len(nl):]
				Whitespace.NewLines++
				Whitespace.HasEscapedLines = true
				return true
			}
		}

	}

	r, size := peakRune(*formatter.Input)

	if r == '\n' {
		*formatter.Input = (*formatter.Input)[size:]
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
		*formatter.Input = (*formatter.Input)[size:]
		return true

	}

	return false
}

func (formatter *Formatter) writeString(str string) {
	formatter.Output = append(formatter.Output, []byte(str)...)
	formatter.OutputColumn += len(str)
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
	case NodeTypeTopLevel:
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

func (f *Formatter) isBlockStart() bool {
	return f.Node().isBlock() && f.isNodeStart()
}

func (f *Formatter) isInitialiserListStart() bool {
	return f.Node().isInitialiserList() && f.isNodeStart()
}

func (f *Formatter) isInvokationStart() bool {
	return f.Node().isInvokation() && f.isNodeStart()
}

func (f *Formatter) isFunctionDefStart() bool {
	return f.Node().isFunctionDef() && f.isNodeStart()
}

func (f *Formatter) isFunctionName() bool {
	return f.Token.Type == TokenTypeIdentifier && f.NextToken.isLeftParenthesis() && (!f.IsDirective || !f.Token.Whitespace.HasSpace)
}

func (f *Formatter) isFunctionStart() bool {
	return f.PreviousToken.Type == TokenTypeIdentifier && f.Token.isLeftParenthesis() && (!f.IsDirective || !f.Token.Whitespace.HasSpace)
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
		(f.Node().isStructOrUnion() && f.Token.isSemicolon()) ||
		((f.Node().isEnum()) && f.Token.isComma()) ||
		((f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && (f.isNodeStart() || f.NextToken.isRightBrace())) ||
		(f.Wrapping && f.isWrappingNode() && f.WrappingStrategy == WrappingStrategyComma && f.Token.isComma()) ||
		(f.Wrapping && f.isWrappingNode() && f.isInitialiserListStart()) ||
		f.isBlockStart() ||
		(f.Wrapping && f.isWrappingNode() && f.Node().isInitialiserList() && f.NextToken.isRightBrace()) ||
		(f.WrappingStrategy == WrappingStrategyLineBreakAfterComma && f.Token.isComma() && f.Token.hasNewLines())

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
}

func (f *Formatter) popNode() {

	f.PreviousNode = f.Node()
	f.PreviousNode.LastToken = f.TokenIndex
	if f.WrappingNode == f.Node().Id {
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

func (f *Formatter) isWrappingNode() bool {
	return f.WrappingNode == f.Node().Id
}

func (f *Formatter) isTopLevelInNode() bool {
	return f.OpenBraces == f.Node().InitialBraces && f.OpenParenthesis == f.Node().InitialParenthesis
}

func (t NodeType) String() string {
	switch t {
	case NodeTypeNone:
		return "NodeTypeNone"
	case NodeTypeTopLevel:
		return "NodeTypeTopLevel"
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

func (n Node) isTopLevel() bool {
	return n.Type == NodeTypeTopLevel
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

func (n Node) isPresent() bool {
	return n.Type != NodeTypeNone
}

func (n Node) isInitialiserList() bool {
	return n.Type == NodeTypeInitialiserList
}

func (n Node) isInvokation() bool {
	return n.Type == NodeTypeInvokation
}

func (n Node) isFunctionDef() bool {
	return n.Type == NodeTypeFunctionDef
}
