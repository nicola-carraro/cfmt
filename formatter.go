package main

import (
	"slices"
	"strings"
)

type Formatter struct {
	Indent              int
	TokenIndex          int
	InputLine           int
	InputColumn         int
	OutputLine          int
	OutputColumn        int
	Input               *string
	Output              []byte
	OpenParenthesis     int
	AcceptStructOrUnion bool
	AcceptEnum          bool
	Nodes               []Node
	LastNodeId          int
	LastPop             Node
	WrappingNode        int
	OpenBraces          int
	Wrapping            bool
	Tokens              *[]Token
	OpenNodeCount       [15]int
}

func (f *Formatter) token() Token {
	return f.tokenAt(f.TokenIndex)
}

func (f *Formatter) previousToken() Token {
	return f.tokenAt(f.TokenIndex - 1)
}

func (f *Formatter) nextToken() Token {
	return f.tokenAt(f.TokenIndex + 1)
}

func (f *Formatter) shouldWrap() bool {
	return f.OutputColumn > 80 ||
		((f.Node().isInitializerList() || f.Node().isFuncOrMacro()) &&
			(f.nextToken().isComment() || f.nextToken().isDirective())) ||
		(f.isInsideFuncOrMacro() && f.Node().isBlock())

}

func format(input string) string {

	f := Formatter{Input: &input, Tokens: new([]Token)}

	(&f).pushNode(NodeTypeTopLevel)

	saved := Formatter{}
	savedNodes := slices.Clone(f.Nodes)

	_ = f.skipSpaceAndCountNewLines()
	for f.update() {
		f.formatToken()

		if !f.Wrapping && f.shouldWrap() {
			f = saved
			f.Wrapping = true
			f.Nodes = slices.Clone(savedNodes)
		} else {
			if f.alwaysOneLine() {
				f.writeNewLines(1)
			} else if f.isEndOfDirective() || f.alwaysDefaultLines() {
				f.writeDefaultLines()
			} else if f.indentedWrapping() {
				f.Indent++
				f.writeNewLines(1)
				f.Indent--
			} else if !f.neverSpace() &&
				!f.nextToken().isRightBrace() &&
				!f.token().isLeftBrace() {
				f.writeString(" ")
			}

			if f.TokenIndex == 0 {
				saved = f
			}

			if !f.isInsideFuncOrMacro() {
				if (f.isBlockStart()) || ((!f.Node().isStructOrUnion() && !f.Node().isDirective()) && f.token().isSemicolon()) {
					f.Wrapping = false
					f.WrappingNode = 0
					saved = f
					savedNodes = slices.Clone(f.Nodes)
				}
			}
		}

		f.TokenIndex++

	}

	return string(f.Output)
}

func (f *Formatter) tokenAt(index int) Token {

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

func (f *Formatter) update() bool {

	if f.token().isStructOrUnion() {
		f.AcceptStructOrUnion = true
	}

	if f.token().isEnum() {
		f.AcceptEnum = true
	}

	if f.startsFunctionArguments() {
		f.AcceptStructOrUnion = false
		f.AcceptEnum = false
	}

	if f.previousToken().isAssignment() && f.isTopLevelInNode() {
		f.Node().RightSideOfAssignment = true
	}

	if f.token().isSemicolon() && f.isTopLevelInNode() {
		f.Node().RightSideOfAssignment = false
	}

	if f.token().isLeftParenthesis() {
		f.OpenParenthesis++
	}

	if f.token().isLeftBrace() {
		f.OpenBraces++
	}

	if !f.Node().isDirective() {

		if f.token().isDirective() {
			f.pushNode(NodeTypeDirective)
		} else if f.startsFunctionArguments() {
			f.pushNode(NodeTypeFuncOrMacro)

		} else if f.token().isLeftParenthesis() && f.previousToken().isFor() {
			f.pushNode(NodeTypeForLoopParenthesis)

		} else if f.token().isLeftBrace() {
			if f.previousToken().isAssignment() || f.Node().isInitializerList() {
				f.pushNode(NodeTypeInitializerList)
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
		f.Node().Type == NodeTypeInitializerList ||
		f.Node().Type == NodeTypeEnum ||
		f.Node().Type == NodeTypeStructOrUnion) &&
		f.token().isRightBrace() {
		f.popNode()
	}

	if (f.Node().isFuncOrMacro() || f.Node().isForLoopParenthesis()) &&
		f.token().isRightParenthesis() && f.OpenParenthesis == (f.Node().InitialParenthesis) {
		f.popNode()
	}
	if f.Node().isDirective() &&
		(f.token().hasUnescapedLines() || f.nextToken().isAbsent()) {
		f.popNode()
	}

	if f.Wrapping && f.WrappingNode == 0 {
		if f.isFunctionStart() && (!f.isRightSideOfAssignment() || f.functionIsEntireRightSide()) {
			f.WrappingNode = f.Node().Id
		} else if f.isInitializerListStart() {
			f.WrappingNode = f.Node().Id
		} else if (f.Node().isTopLevel() || f.Node().isBlock()) && f.Node().RightSideOfAssignment && !f.isFunctionName() {
			f.WrappingNode = f.Node().Id
		}
	}

	if f.token().isRightParenthesis() {
		f.OpenParenthesis--

	}

	if f.token().isRightBrace() {
		f.OpenBraces--

	}

	if f.Node().isDirective() {
		if f.token().hasEscapedLines() {
			if f.token().isLeftBrace() || f.token().isLeftParenthesis() {
				f.Indent++
			}

			if f.nextToken().isRightBrace() || f.nextToken().isRightParenthesis() {
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

	return !f.token().isAbsent()
}

func (f *Formatter) shouldIncreaseIndent() bool {
	return ((f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && f.isNodeStart()) ||
		(f.Wrapping && f.isWrappingNode() && (f.isInitializerListStart() || f.isFuncOrMacroStart()))
}

func (f *Formatter) shouldDecreaseIndent() bool {
	return ((f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && f.nextToken().isRightBrace()) ||
		(f.Wrapping && f.isWrappingNode() && f.Node().isInitializerList() && f.nextToken().isRightBrace()) ||
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

	if formatter.Node().isDirective() {
		for _, nl := range newLineInDirective {
			if formatter.Node().isDirective() && strings.HasPrefix((*formatter.Input), nl) {
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
	if formatter.token().Whitespace.NewLines <= 1 || formatter.nextToken().isRightBrace() {
		formatter.writeNewLines(1)

	} else {
		formatter.writeNewLines(2)
	}
}

func (formatter *Formatter) twoLinesOrEof() {
	if formatter.nextToken().isAbsent() {
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
	if !formatter.nextToken().isDirective() {
		for indentLevel := 0; indentLevel < formatter.Indent; indentLevel++ {
			formatter.writeString(indentation)
		}
	}

}

func (f *Formatter) writeDefaultLines() {

	switch f.Node().Type {
	case NodeTypeTopLevel:
		f.twoLinesOrEof()
	case NodeTypeDirective, NodeTypeFuncOrMacro, NodeTypeInitializerList, NodeTypeStructOrUnion, NodeTypeEnum, NodeTypeForLoopParenthesis:
		f.writeNewLines(1)
	case NodeTypeBlock:
		f.oneOrTwoLines()
	default:
		panic("unreachable")
	}
}

func (formatter *Formatter) IsParenthesis() bool {
	return formatter.OpenParenthesis > 0
}

func (f *Formatter) isPointerOperator() bool {
	return f.token().canBePointerOperator() && (!f.previousToken().canBeLeftOperand() || !f.Node().RightSideOfAssignment)
}

func (f *Formatter) hasPostfixIncrDecr() bool {
	return f.nextToken().isIncrDecrOperator() && (f.token().isIdentifier() || f.token().isRightParenthesis())
}

func (f *Formatter) isPrefixIncrDecr() bool {
	return f.token().isIncrDecrOperator() && (f.nextToken().isIdentifier() || f.nextToken().isLeftParenthesis())
}

func (f *Formatter) hasTrailingComment() bool {
	return f.nextToken().isSingleLineComment() && f.token().Whitespace.NewLines == 0
}

func (f *Formatter) formatMultilineComment() {
	text := strings.TrimSpace(f.token().Content[2 : len(f.token().Content)-2])

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
	text := strings.TrimSpace(f.token().Content[2:])
	f.writeString("// ")
	f.writeString(text)
}

func (f *Formatter) formatToken() {

	if f.token().isMultilineComment() {
		f.formatMultilineComment()
	} else if f.token().isSingleLineComment() {
		f.formatSingleLineComment()
	} else {
		f.writeString(f.token().Content)
	}
}

func (f *Formatter) startsFunctionArguments() bool {
	if !f.Node().isDirective() {
		return f.previousToken().Type == TokenTypeIdentifier && f.token().isLeftParenthesis()

	} else {
		return f.previousToken().Type == TokenTypeIdentifier && f.token().isLeftParenthesis() && !f.previousToken().Whitespace.HasSpace
	}
}

func (f *Formatter) isBlockStart() bool {
	return f.Node().isBlock() && f.isNodeStart()
}

func (f *Formatter) isInitializerListStart() bool {
	return f.Node().isInitializerList() && f.isNodeStart()
}

func (f *Formatter) isFuncOrMacroStart() bool {
	return f.Node().isFuncOrMacro() && f.isNodeStart()
}

func (f *Formatter) isFunctionName() bool {
	return f.token().Type == TokenTypeIdentifier && f.nextToken().isLeftParenthesis() && (!f.Node().isDirective() || !f.token().Whitespace.HasSpace)
}

func (f *Formatter) isFunctionStart() bool {
	return f.previousToken().Type == TokenTypeIdentifier && f.token().isLeftParenthesis() && (!f.Node().isDirective() || !f.token().Whitespace.HasSpace)
}

func (f *Formatter) neverSpace() bool {

	return f.nextToken().isSemicolon() ||
		f.token().isLeftParenthesis() ||
		f.nextToken().isRightParenthesis() ||
		f.token().isLeftBrace() ||
		f.nextToken().isRightBrace() ||
		f.token().isLeftBracket() ||
		f.nextToken().isLeftBracket() ||
		f.nextToken().isRightBracket() ||
		f.isPointerOperator() ||
		f.isFunctionName() ||
		f.hasPostfixIncrDecr() ||
		f.isPrefixIncrDecr() ||
		(f.token().isIdentifier() && f.nextToken().isDotOperator()) ||
		f.token().isDotOperator() ||
		f.token().isArrowOperator() ||
		f.nextToken().isArrowOperator() ||
		f.nextToken().isComma() ||
		f.token().isNegation() ||
		f.token().isSizeOf() ||
		f.token().isStringizingOp() ||
		f.token().isCharizingOp() ||
		f.token().isTokenPastingOp() ||
		f.nextToken().isTokenPastingOp() ||
		(f.Node().DirectiveType == DirectiveTypeInclude && ((f.nextToken().isGreaterThanSign()) || f.token().isLessThanSign()))
}

func (f *Formatter) wrappingStrategyComma() bool {
	return f.Node().isFuncOrMacro()
}

func (f *Formatter) wrappingStrategyLineBreakAfterComma() bool {
	return f.Node().isInitializerList()
}

func (f *Formatter) alwaysOneLine() bool {

	return f.nextToken().isAbsent() ||
		(f.token().isComment() && (f.previousToken().hasNewLines() || f.previousToken().isAbsent())) ||
		(f.afterInclude() && f.nextToken().isIncludeDirective()) ||
		(f.afterPragma() && f.nextToken().isPragmaDirective()) ||
		(f.afterPragma() && f.nextToken().isPragmaDirective()) ||
		(f.Node().isStructOrUnion() && f.token().isSemicolon()) ||
		((f.Node().isEnum()) && f.token().isComma()) ||
		((f.Node().isStructOrUnion() || f.Node().isBlock() || f.Node().isEnum()) && (f.isNodeStart() || f.nextToken().isRightBrace())) ||
		(f.Wrapping && f.isWrappingNode() && f.wrappingStrategyComma() && f.token().isComma()) ||
		(f.Wrapping && f.isWrappingNode() && f.isInitializerListStart()) ||
		(f.Wrapping && f.isWrappingNode() && f.isFuncOrMacroStart()) ||
		(f.Wrapping && f.isWrappingNode() && f.beforeEndOfFuncOrMacro()) ||
		f.isBlockStart() ||
		(f.Wrapping && f.isWrappingNode() && f.Node().isInitializerList() && f.nextToken().isRightBrace()) ||
		(f.Wrapping && f.isWrappingNode() && f.wrappingStrategyLineBreakAfterComma() && f.token().isComma() && f.token().hasNewLines())
}

func (f *Formatter) indentedWrapping() bool {
	return (f.Wrapping && f.isWrappingNode() && (f.Node().isBlock() || f.Node().isTopLevel()) && f.token().hasNewLines())
}

func (f *Formatter) alwaysDefaultLines() bool {
	return (f.nextToken().isDirective() && !f.previousToken().isAbsent()) ||
		f.isEndOfDirective() ||
		(f.token().isComment() && !f.previousToken().hasNewLines() && !f.previousToken().isAbsent()) ||
		f.nextToken().isMultilineComment() ||
		(f.token().isSemicolon() && !f.Node().isForLoopParenthesis() && !f.hasTrailingComment()) ||
		(f.Node().isDirective() && f.token().hasEscapedLines()) ||
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

	if f.previousToken().isDo() {
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

	if t == NodeTypeDirective {
		node.DirectiveType = f.token().DirectiveType
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
	return f.Node().isFuncOrMacro() && f.nextToken().isRightParenthesis() && f.OpenParenthesis == f.Node().InitialParenthesis
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
	if !f.tokenAt(f.TokenIndex - 2).isAssignment() {
		return false
	}

	i := f.TokenIndex + 1

	openParenthesis := 1

	for token := f.tokenAt(i); !token.isAbsent(); token = f.tokenAt(i) {
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

	return f.tokenAt(i + 1).isSemicolon()
}
