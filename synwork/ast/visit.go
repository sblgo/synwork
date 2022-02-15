package ast

type Visitor interface {
	Visit(n Node) Visitor
}

func Walk(v Visitor, n Node) {
	var s Visitor
	if s = v.Visit(n); s == nil {
		return
	}
	switch i := n.(type) {
	case *BlockNode:
		Walk(s, i.Content)
	case *BlockContentNode:
		for _, c := range i.Assignments {
			Walk(s, c)
		}
		for _, c := range i.Blocks {
			Walk(s, c)
		}
	case *ComplexValue:
		for _, c := range i.Assignments {
			Walk(s, c)
		}
		for _, c := range i.Blocks {
			Walk(s, c)
		}
	case *AssignmentNode:
		Walk(s, i.Value)
	case *StringValue:

	}
	v.Visit(nil)
}
