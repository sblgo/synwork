package parser

import (
	"fmt"
	"go/scanner"
	"go/token"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"sbl.systems/go/synwork/synwork/ast"
)

type Parser struct {
	dirname string
	scan    scanner.Scanner
	fset    *token.FileSet
	files   []fs.FileInfo
	idx     int
	errors  []error
	Blocks  []*ast.BlockNode
}

type parseError struct {
	pos string
	msg string
}

func (p parseError) Error() string {
	return fmt.Sprintf("ParserError at [%s]: %s", p.pos, p.msg)
}

func NewParser(dirname string) (*Parser, error) {
	p := &Parser{
		dirname: dirname,
		fset:    token.NewFileSet(),
		files:   []fs.FileInfo{},
		errors:  []error{},
	}

	var err error
	files, err := ioutil.ReadDir(p.dirname)
	if err != nil {
		return nil, err
	}
	for _, r := range files {
		if strings.HasSuffix(r.Name(), ".snw") {
			p.files = append(p.files, r)
		}
	}
	if len(p.files) == 0 {
		return nil, fmt.Errorf("missing snw file in %s", p.dirname)
	}
	return p, nil
}

func (p *Parser) Parse() error {
	p.Blocks = []*ast.BlockNode{}
	for ; p.idx < len(p.files); p.idx++ {
		name := filepath.Join(p.dirname, p.files[p.idx].Name())
		src, err := ioutil.ReadFile(name)
		if err != nil {
			return fmt.Errorf("error opening %s. details %s", name, err.Error())
		}
		file := p.fset.AddFile(name, p.fset.Base(), len(src)) // register input "file"
		p.scan.Init(file, src, nil /* no error handler */, 0)
		if err = p.parseFile(); err != nil {
			return err
		}
	}
	if len(p.errors) > 0 {
		errstr := "Errors during parsing:"
		for _, e := range p.errors {
			errstr += "\n\t" + e.Error()
		}
		return fmt.Errorf(errstr)
	}
	return nil
}

func (p *Parser) parseFile() error {
	for {
		if block, ok := p.parseBlock(); ok {
			p.Blocks = append(p.Blocks, block)
		} else {
			return nil
		}
	}
}

func (p *Parser) parseBlock() (*ast.BlockNode, bool) {
	node := &ast.BlockNode{
		Identifiers: []string{},
	}
	for {
		pos, tok, lit := p.scan.Scan()
		switch tok {
		case token.IDENT:
			node.Begin = p.posString(pos)
			node.Type = lit
			goto params
		case token.SEMICOLON:
		case token.EOF:
			return nil, false
		default:
			p.error(pos, fmt.Sprintf("expected IDENT but found <%s> with %s", tok.String(), lit))
		}
	}
params:
	for {
		pos, tok, lit := p.scan.Scan()
		switch tok {
		case token.STRING:
			node.Identifiers = append(node.Identifiers, cleanStr(lit))
			goto params
		case token.LBRACE:
			if sub, end, ok := p.parseBlockContent(pos); ok {
				node.Content = sub
				node.End = end
				return node, true
			} else {
				return nil, false
			}
		case token.SEMICOLON:
		case token.EOF:
			p.error(pos, "unexpected EOF in block definition")
			return nil, false
		default:
			p.error(pos, fmt.Sprintf("expected String or { but found <%s> with %s", tok.String(), lit))
		}

	}
}

func (p *Parser) error(pos token.Pos, msg string) {
	p.errors = append(p.errors, parseError{
		msg: msg,
		pos: p.fset.Position(pos).String(),
	})
}

func (p *Parser) posString(pos token.Pos) string {
	return p.fset.Position(pos).String()
}

func (p *Parser) parseBlockContent(pos token.Pos) (*ast.BlockContentNode, string, bool) {
	node := &ast.BlockContentNode{
		Begin:       p.posString(pos),
		Assignments: []*ast.AssignmentNode{},
		Blocks:      []*ast.BlockNode{},
	}
	ident := ""
identifier:
	for {
		pos, tok, lit := p.scan.Scan()
		switch tok {
		case token.IDENT:
			ident = lit
			goto assignment
		case token.RBRACE:
			node.End = p.posString(pos)
			return node, node.End, true
		case token.EOF:
			p.error(pos, "EOF in BLOCK before }")
			return nil, "", false
		case token.SEMICOLON:
		default:
			p.error(pos, fmt.Sprintf("expected IDENT or } but found <%s> with %s", tok.String(), lit))
		}
	}
assignment:
	for {
		pos, tok, lit := p.scan.Scan()
		switch tok {
		case token.ASSIGN:
			if asg, ok := p.parseAssignment(pos, ident); ok {
				node.Assignments = append(node.Assignments, asg)
			}
			goto identifier
		case token.LBRACE:
			if sub, end, ok := p.parseBlockContent(pos); ok {
				node.Blocks = append(node.Blocks, &ast.BlockNode{
					Begin:   p.posString(pos),
					End:     end,
					Type:    ident,
					Content: sub,
				})
			}
			goto identifier
		case token.SEMICOLON:
		case token.EOF:
			p.error(pos, "EOF in BLOCK before }")
			return nil, "", false
		default:
			p.error(pos, fmt.Sprintf("expected = or { but found <%s> with %s", tok.String(), lit))
		}
	}
}

func (p *Parser) parseAssignment(pos token.Pos, lit string) (*ast.AssignmentNode, bool) {
	node := &ast.AssignmentNode{
		Begin:      p.posString(pos),
		Identifier: lit,
	}
	pos, tok, lit := p.scan.Scan()
	switch tok {
	case token.STRING:
		node.Value = &ast.StringValue{
			Begin: p.posString(pos),
			Value: cleanStr(lit),
		}
		node.End = p.posString(pos)
	case token.INT:
		val, _ := strconv.Atoi(lit)
		node.Value = &ast.IntValue{
			Begin: p.posString(pos),
			Value: val,
		}
	case token.FLOAT:
		val, _ := strconv.ParseFloat(lit, 64)
		node.Value = &ast.FloatValue{
			Begin: p.posString(pos),
			Value: val,
		}
	case token.ILLEGAL:
		if lit == "$" {
			refParts, ok := p.parseReference()
			if !ok {
				return nil, false
			}
			node.Value = &ast.ReferenceValue{
				Begin:    p.posString(pos),
				RefParts: refParts,
			}

		} else {
			p.error(pos, fmt.Sprintf("expected once of String, Int, Float, $ or { but found <%s> with %s", tok.String(), lit))
			return nil, false
		}
	case token.LBRACE:
		asglist, end, ok := p.parseBlockContent(pos)
		if ok {
			node.Value = &ast.ComplexValue{
				BlockContentNode: *asglist,
			}
			node.End = end
		} else {
			return nil, false
		}
	default:
		p.error(pos, fmt.Sprintf("expected once of String, Int, Float, $ or { but found <%s> with %s", tok.String(), lit))
		return nil, false
	}
	return node, true
}

func (p *Parser) parseReference() ([]string, bool) {
	refParts := []string{}
next_ref_part:
	_, tok, lit := p.scan.Scan()
	switch tok {
	case token.IDENT:
		refParts = append(refParts, lit)
		_, tok, _ = p.scan.Scan()
		switch tok {
		case token.PERIOD:
			goto next_ref_part
		case token.SEMICOLON:
			return refParts, true
		}
	default:
		return nil, false
	}

	return nil, false
}

func cleanStr(str string) string {
	return strings.Trim(str, "\"")
}
