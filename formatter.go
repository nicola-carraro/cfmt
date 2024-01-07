package main

import (
	"fmt"
	"slices"
	"strings"
)

type Formatter struct {
	PreviousToken       Token
	Token               Token
	NextToken           Token
	Indent              int
	TokenIndex          int
	InputLine           int
	InputColumn         int
	OutputLine          int
	OutputColumn        int
	Input               *string
	Output              []byte
	SavedInput          string
	OpenParenthesis     int
	IsDirective         bool
	IsIncludeDirective  bool
	IsEndOfInclude      bool
	AcceptStructOrUnion bool
	AcceptEnum          bool
	IsForLoop           bool
	ForOpenParenthesis  int
	Nodes               []Node
	LastNodeId          int
	LastPop             Node
	WrappingNode        int
	OpenBraces          int
	Wrapping            bool
	Tokens              *[]Token
	WrappingStrategy    WrappingStrategy
	OpenNodeCount       [15]int
}

type NodeType int

const (
	NodeTypeNone NodeType = iota
	NodeTypeTopLevel
	NodeTypeMacroDef
	NodeTypeOtherDirective
	NodeTypeFuncOrMacro
	NodeTypeBlock
	NodeTypeInitialiserList
	NodeTypeStructOrUnion
	NodeTypeEnum
)

type WrappingStrategy int

const (
	WrappingStrategyNone WrappingStrategy = iota
	WrappingStrategyComma
	WrappingStrategyLineBreak
	WrappingStrategyLineBreakAfterComma
)

type BlockType int

const (
	BlockTypeNone BlockType = iota
	BlockTypeDoWhile
)

type DirectiveType int

const (
	DirectiveTypeNone DirectiveType = iota
	DirectiveTypeDefine
	DirectiveTypeError
	DirectiveTypeIf
	DirectiveTypeElif
	DirectiveTypeElse
	DirectiveTypeEndif
	DirectiveTypeIfdef
	DirectiveTypeIfndef
	DirectiveTypeIfDef
	DirectiveTypeUndef
	DirectiveTypeInclude
	DirectiveTypeLine
	DirectiveTypePragma
)

type Node struct {
	Type                  NodeType
	Id                    int
	FirstToken            int
	LastToken             int
	InitialIndent         int
	InitialParenthesis    int
	InitialBraces         int
	BlockType             BlockType
	DirectiveType         DirectiveType
	RightSideOfAssignment bool
}

type StructUnionEnum struct {
	Indent int
}

func (f *Formatter) wrapping() bool {
	return f.WrappingNode != 0
}

func (f *Formatter) shouldWrap() bool {
	return f.OutputColumn > 80 ||
		((f.Node().isInitialiserList() || f.Node().isFuncOrMacro()) &&
			(f.NextToken.isComment() || f.NextToken.isDirective())) ||
		(f.isInsideFuncOrMacro() && f.Node().isBlock())

}

func format(input string) string {

	f := Formatter{Input: &input, Tokens: new([]Token)}

	(&f).pushNode(NodeTypeTopLevel)

	saved := f

	savedNodes := slices.Clone(f.Nodes)

	for f.parseToken() {
		f.formatToken()

		if !f.Wrapping && f.shouldWrap() {
			f = saved
			f.Wrapping = true
			f.Nodes = slices.Clone(savedNodes)
			f.WrappingStrategy = WrappingStrategyNone
			continue
		}

		if f.alwaysOneLine() {
			f.writeNewLines(1)
		} else if f.isEndOfDirective() || f.alwaysDefaultLines() {
			f.writeDefaultLines()
		} else if f.indentedWrapping() {
			f.Indent++
			f.writeNewLines(1)
			f.Indent--
		} else if !f.neverSpace() &&
			!f.NextToken.isRightBrace() &&
			!f.Token.isLeftBrace() {
			f.writeString(" ")
		}

		if !f.isInsideFuncOrMacro() {
			if (f.isBlockStart()) || ((!f.Node().isStructOrUnion() && !f.Node().isDirective()) && f.Token.isSemicolon()) {
				f.Wrapping = false
				f.WrappingNode = 0
				f.WrappingStrategy = WrappingStrategyNone
				saved = f
				savedNodes = slices.Clone(f.Nodes)
			}
		}

	}

	return string(f.Output)
}

func (f *Formatter) getToken(index int) Token {

	if index < 0 {
		return Token{}
	}

	for len(*f.Tokens) <= index {
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

	if f.Token.isDirective() {
		f.IsDirective = true
	}
	wasInclude := f.IsIncludeDirective

	if f.Token.isIncludeDirective() {
		f.IsIncludeDirective = true
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

	if f.PreviousToken.isAssignment() && f.isTopLevelInNode() {
		f.Node().RightSideOfAssignment = true
	}

	if f.Token.isSemicolon() && f.isTopLevelInNode() {
		f.Node().RightSideOfAssignment = false
	}

	f.NextToken = f.getToken(f.TokenIndex)

	if f.Token.isLeftParenthesis() {
		f.OpenParenthesis++
	}

	if f.Token.isLeftBrace() {
		f.OpenBraces++
	}

	if !f.Node().isDirective() {

		if f.Token.isDefineDirective() {
			f.pushNode(NodeTypeMacroDef)
		} else if f.Token.isDirective() {
			f.pushNode(NodeTypeOtherDirective)

		} else if f.startsFunctionArguments() {
			f.pushNode(NodeTypeFuncOrMacro)

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

	if f.Node().Type == NodeTypeFuncOrMacro &&
		f.Token.isRightParenthesis() && f.OpenParenthesis == (f.Node().InitialParenthesis) {
		f.popNode()
	}
	if f.Node().isDirective() &&
		(f.Token.hasUnescapedLines() || f.NextToken.isAbsent()) {
		f.popNode()
	}

	if f.Wrapping && f.WrappingStrategy == WrappingStrategyNone {
		if f.isFunctionStart() && (!f.isRightSideOfAssignment() || f.functionIsEntireRightSide()) {
			f.WrappingStrategy = WrappingStrategyComma
			f.WrappingNode = f.Node().Id
		} else if f.isInitialiserListStart() {
			f.WrappingStrategy = WrappingStrategyLineBreakAfterComma
			f.WrappingNode = f.Node().Id
		} else if (f.Node().isTopLevel() || f.Node().isBlock()) && f.Node().RightSideOfAssignment && !f.isFunctionName() {
			f.WrappingStrategy = WrappingStrategyLineBreak
			f.WrappingNode = f.Node().Id
		}
	}

	if f.Token.isRightParenthesis() {
		f.OpenParenthesis--

	}

	if f.Token.isRightBrace() {
		f.OpenBraces--

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

	if f.Token.Whitespace.HasUnescapedLines || f.NextToken.isAbsent() {
		f.IsDirective = false
		f.IsIncludeDirective = false
	}

	if wasInclude && !f.IsIncludeDirective {
		f.IsEndOfInclude = true
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
		(f.Wrapping && f.isWrappingNode() && (f.isInitialiserListStart() || f.isFuncOrMacroStart()))
}

func (f *Formatter) shouldDecreaseIndent() bool {
	return ((f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && f.NextToken.isRightBrace()) ||
		(f.Wrapping && f.isWrappingNode() && f.Node().isInitialiserList() && f.NextToken.isRightBrace()) ||
		(f.Wrapping && f.isWrappingNode() && f.beforeEndOfFuncOrMacro())
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
	return f.LastPop.isDirective() && f.LastPop.LastToken == f.TokenIndex
}

func (formatter *Formatter) writeNewLines(lines int) {
	const newLine = "\n"

	for line := 0; line < lines; line++ {
		if formatter.Node().isDirective() {
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
	case NodeTypeMacroDef, NodeTypeFuncOrMacro, NodeTypeInitialiserList, NodeTypeStructOrUnion, NodeTypeEnum:
		f.writeNewLines(1)
	case NodeTypeBlock:
		f.oneOrTwoLines()
	default:
		panic("unreacheable")
	}
}

func (formatter *Formatter) IsParenthesis() bool {
	return formatter.OpenParenthesis > 0
}

func (f *Formatter) isPointerOperator() bool {
	return f.Token.canBePointerOperator() && (!f.PreviousToken.canBeLeftOperand() || !f.Node().RightSideOfAssignment)
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

func (f *Formatter) isFuncOrMacroStart() bool {
	return f.Node().isFuncOrMacro() && f.isNodeStart()
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
		(f.Node().DirectiveType == DirectiveTypeInclude && ((f.NextToken.isGreaterThanSign()) || f.Token.isLessThanSign()))
}

func (f *Formatter) alwaysOneLine() bool {

	return f.NextToken.isAbsent() ||
		(f.Token.isComment() && (f.PreviousToken.hasNewLines() || f.PreviousToken.isAbsent())) ||
		(f.afterInclude() && f.NextToken.isIncludeDirective()) ||
		(f.afterPragma() && f.NextToken.isPragmaDirective()) ||
		(f.afterPragma() && f.NextToken.isPragmaDirective()) ||
		(f.Node().isStructOrUnion() && f.Token.isSemicolon()) ||
		((f.Node().isEnum()) && f.Token.isComma()) ||
		((f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && (f.isNodeStart() || f.NextToken.isRightBrace())) ||
		(f.Wrapping && f.isWrappingNode() && f.WrappingStrategy == WrappingStrategyComma && f.Token.isComma()) ||
		(f.Wrapping && f.isWrappingNode() && f.isInitialiserListStart()) ||
		(f.Wrapping && f.isWrappingNode() && f.isFuncOrMacroStart()) ||
		(f.Wrapping && f.isWrappingNode() && f.beforeEndOfFuncOrMacro()) ||
		f.isBlockStart() ||
		(f.Wrapping && f.isWrappingNode() && f.Node().isInitialiserList() && f.NextToken.isRightBrace()) ||
		(f.Wrapping && f.isWrappingNode() && f.WrappingStrategy == WrappingStrategyLineBreakAfterComma && f.Token.isComma() && f.Token.hasNewLines())

}

func (f *Formatter) indentedWrapping() bool {
	return (f.Wrapping && f.isWrappingNode() && (f.Node().isBlock() || f.Node().isTopLevel()) && f.WrappingStrategy == WrappingStrategyLineBreak && f.Token.hasNewLines())
}

func (f *Formatter) alwaysDefaultLines() bool {
	return (f.NextToken.isDirective() && !f.PreviousToken.isAbsent()) ||
		f.isEndOfDirective() ||
		(f.Token.isComment() && !f.PreviousToken.hasNewLines() && !f.PreviousToken.isAbsent()) ||
		f.NextToken.isMultilineComment() ||
		(f.Token.isSemicolon() && !f.IsForLoop && !f.hasTrailingComment()) ||
		(f.Node().isDirective() && f.Token.hasEscapedLines()) ||
		(f.afterEndOfBlock() && !(f.LastPop.BlockType == BlockTypeDoWhile))
}

func (f *Formatter) Node() *Node {

	if len(f.Nodes) == 0 {
		return &Node{}
	}
	return &f.Nodes[len(f.Nodes)-1]
}

func (f *Formatter) ParentNode() *Node {
	if len(f.Nodes) < 2 {
		return new(Node)
	}

	return &f.Nodes[len(f.Nodes)-2]
}

func (f *Formatter) pushNode(t NodeType) {
	f.LastNodeId++

	blockType := BlockTypeNone

	if f.PreviousToken.isDo() {
		blockType = BlockTypeDoWhile
	}

	node := Node{
		Type:               t,
		Id:                 f.LastNodeId,
		FirstToken:         f.TokenIndex,
		InitialIndent:      f.Indent,
		InitialParenthesis: f.OpenParenthesis,
		InitialBraces:      f.OpenBraces,
		BlockType:          blockType,
	}

	if t == NodeTypeMacroDef || t == NodeTypeOtherDirective {
		node.DirectiveType = f.Token.DirectiveType
		f.Indent = 0
	}

	f.Nodes = append(f.Nodes, node)

	f.OpenNodeCount[t]++
}

func (f *Formatter) popNode() {

	f.LastPop = *f.Node()
	f.LastPop.LastToken = f.TokenIndex
	if f.WrappingNode == f.Node().Id {
		f.WrappingNode = 0
	}
	if f.Node().isDirective() {
		f.Indent = f.Node().InitialIndent
	}

	f.OpenNodeCount[f.Node().Type]--

	f.Nodes = f.Nodes[:len(f.Nodes)-1]

}

func (f *Formatter) isInsideNode(nodeType NodeType) bool {
	return f.OpenNodeCount[nodeType] > 0
}

func (f *Formatter) isInsideFuncOrMacro() bool {
	return f.isInsideNode(NodeTypeFuncOrMacro)
}

func (f *Formatter) isNodeStart() bool {
	return f.TokenIndex == f.Node().FirstToken
}

func (f *Formatter) afterEndOfBlock() bool {
	return f.LastPop.isBlock() && f.LastPop.LastToken == f.TokenIndex
}

func (f *Formatter) isWrappingNode() bool {
	return f.WrappingNode == f.Node().Id
}

func (f *Formatter) isTopLevelInNode() bool {
	return f.OpenBraces == f.Node().InitialBraces && f.OpenParenthesis == f.Node().InitialParenthesis
}

func (f *Formatter) afterInclude() bool {
	return f.LastPop.isIncludeDirective() && f.LastPop.LastToken == f.TokenIndex
}

func (f *Formatter) afterPragma() bool {
	return f.LastPop.isPragmaDirective() && f.LastPop.LastToken == f.TokenIndex
}

func (f *Formatter) beforeEndOfFuncOrMacro() bool {
	return f.Node().isFuncOrMacro() && f.NextToken.isRightParenthesis() && f.OpenParenthesis == f.Node().InitialParenthesis
}

func (f *Formatter) isRightSideOfAssignment() bool {
	for _, node := range f.Nodes {
		if node.RightSideOfAssignment {
			return true
		}
	}

	return false
}

func (f *Formatter) functionIsEntireRightSide() bool {
	if !f.getToken(f.TokenIndex - 3).isAssignment() {
		return false
	}

	i := f.TokenIndex

	openParenthesis := 1

	for token := f.getToken(i); !token.isAbsent(); token = f.getToken(i) {
		if token.isLeftParenthesis() {
			openParenthesis++
		}
		if token.isRightParenthesis() {
			openParenthesis--
		}

		if openParenthesis == 0 {
			break
		}
		i++
	}

	return f.getToken(i + 1).isSemicolon()
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
		NodeTypeFuncOrMacro:
		return "NodeTypeFuncOrMacro"
	case NodeTypeBlock:
		return "NodeTypeBlock"
	case NodeTypeInitialiserList:
		return "NodeTypeInitialiserList"
	case NodeTypeStructOrUnion:
		return "NodeTypeStructOrUnion"
	case NodeTypeEnum:
		return "NodeTypeEnum"
	default:
		panic(fmt.Sprintf("Unexpected node type %d", t))
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

func (n Node) isFuncOrMacro() bool {
	return n.Type == NodeTypeFuncOrMacro
}

func (n Node) isIncludeDirective() bool {
	return n.Type == NodeTypeOtherDirective && n.DirectiveType == DirectiveTypeInclude
}

func (n Node) isPragmaDirective() bool {
	return n.Type == NodeTypeOtherDirective && n.DirectiveType == DirectiveTypePragma
}
