package ast

import "sbl.systems/go/synwork/plugin-sdk/schema"

type Node interface {
	Pos() string
}

type ValueNode interface {
	Node
	Type() schema.SchemaType
}

type AssignmentNode struct {
	Begin, End string
	Identifier string
	Value      ValueNode
}
type BlockNode struct {
	Begin, End  string
	Type        string
	Identifiers []string
	Content     *BlockContentNode
}

type BlockContentNode struct {
	Begin, End  string
	Assignments []*AssignmentNode
	Blocks      []*BlockNode
}

type StringValue struct {
	Begin string
	Value string
}

type IntValue struct {
	Begin string
	Value int
}

type FloatValue struct {
	Begin string
	Value float64
}

type BoolValue struct {
	Begin string
	Value bool
}

type ReferenceValue struct {
	Begin    string
	RefParts []string
}

type ComplexValue struct {
	BlockContentNode
}

func (p *AssignmentNode) Pos() string   { return p.Begin }
func (p *BlockNode) Pos() string        { return p.Begin }
func (p *BlockContentNode) Pos() string { return p.Begin }
func (p *StringValue) Pos() string      { return p.Begin }
func (p *ComplexValue) Pos() string     { return p.Begin }
func (p *ReferenceValue) Pos() string   { return p.Begin }
func (p *IntValue) Pos() string         { return p.Begin }
func (p *FloatValue) Pos() string       { return p.Begin }
func (p *BoolValue) Pos() string        { return p.Begin }

func (p *StringValue) Type() schema.SchemaType    { return schema.TypeString }
func (p *ReferenceValue) Type() schema.SchemaType { return schema.TypeString }
func (p *ComplexValue) Type() schema.SchemaType   { return schema.TypeMap }
func (p *IntValue) Type() schema.SchemaType       { return schema.TypeInt }
func (p *FloatValue) Type() schema.SchemaType     { return schema.TypeFloat }
func (p *BoolValue) Type() schema.SchemaType      { return schema.TypeBool }
