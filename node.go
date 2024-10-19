package main

import "fmt"

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

type NodeType int

const (
	NodeTypeNone NodeType = iota
	NodeTypeTopLevel
	NodeTypeDirective
	NodeTypeFuncOrMacroCall
	NodeTypeFuncOrMacroDef
	NodeTypeBlock
	NodeTypeInitializerList
	NodeTypeStructOrUnion
	NodeTypeEnum
	NodeTypeForLoopParenthesis
	NodeTypeCount
)

type BlockType int

const (
	BlockTypeNone BlockType = iota
	BlockTypeDoWhile
)

func (t NodeType) String() string {
	switch t {
	case NodeTypeNone:
		return "NodeTypeNone"
	case NodeTypeTopLevel:
		return "NodeTypeTopLevel"
	case NodeTypeDirective:
		return "NodeTypeDirective"
	case
		NodeTypeFuncOrMacroCall:
		return "NodeTypeFuncOrMacroCall"
	case
		NodeTypeFuncOrMacroDef:
		return "NodeTypeFuncOrDef"
	case NodeTypeBlock:
		return "NodeTypeBlock"
	case NodeTypeInitializerList:
		return "NodeTypeInitializerList"
	case NodeTypeStructOrUnion:
		return "NodeTypeStructOrUnion"
	case NodeTypeEnum:
		return "NodeTypeEnum"
	case NodeTypeForLoopParenthesis:
		return "NodeTypeForLoopParenthesis"
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
	return n.Type == NodeTypeDirective
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

func (n Node) isInitializerList() bool {
	return n.Type == NodeTypeInitializerList
}

func (n Node) isFuncOrMacro() bool {
	return n.Type == NodeTypeFuncOrMacroCall || n.Type == NodeTypeFuncOrMacroDef
}

func (n Node) isFuncOrMacroDef() bool {
	return n.Type == NodeTypeFuncOrMacroDef
}

func (n Node) isForLoopParenthesis() bool {
	return n.Type == NodeTypeForLoopParenthesis
}

func (n Node) isIncludeDirective() bool {
	return n.Type == NodeTypeDirective && n.DirectiveType == DirectiveTypeInclude
}

func (n Node) isPragmaDirective() bool {
	return n.Type == NodeTypeDirective && n.DirectiveType == DirectiveTypePragma
}
