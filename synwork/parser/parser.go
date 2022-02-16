package parser

import (
	"sbl.systems/go/synwork/synwork/ast"
)

type Parser struct {
	dirname   string
	tokenizer *Tokenizer
	Blocks    []*ast.BlockNode
}

func NewParser(dirname string) (*Parser, error) {
	tk, err := NewTokenizer(dirname)
	if err != nil {
		return nil, err
	}
	p := &Parser{
		dirname:   dirname,
		tokenizer: tk,
	}
	return p, nil
}

func (p *Parser) Parse() error {
	yyParse(p.tokenizer)
	if len(p.tokenizer.parseErrors) > 0 {
		return ParseError(p.tokenizer.parseErrors)
	}
	p.Blocks = p.tokenizer.Blocks
	return nil
}
