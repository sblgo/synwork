package parser

import (
	"fmt"
	"go/scanner"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"

	"sbl.systems/go/synwork/synwork/ast"
)

type (
	Tokenizer struct {
		actualToken *Token
		parseErrors []ErrorEntry

		scanner         scanner.Scanner
		Blocks          []*ast.BlockNode
		files           []string
		file            *token.File
		fileSet         *token.FileSet
		contentProvider func(string) ([]byte, error)
		fileIndex       int
		dirname         string
	}

	Token struct {
		Value     int
		baseValue token.Token
		Position  string
		RawString string
	}

	ErrorEntry struct {
		Message string
		Token   *Token
	}

	ParseError []ErrorEntry
)

func (p ParseError) Error() string {
	var result string
	for _, e := range p {
		result += fmt.Sprintf("%s: %s\n", e.Token.Position, e.Message)
	}
	return result
}

func NewTokenizerForTest(fileName string, fileContent string) (*Tokenizer, error) {
	p := &Tokenizer{
		dirname:         ".",
		fileSet:         token.NewFileSet(),
		files:           []string{fileName},
		parseErrors:     []ErrorEntry{},
		Blocks:          []*ast.BlockNode{},
		contentProvider: func(s string) ([]byte, error) { return []byte(fileContent), nil },
	}
	return p, nil
}

func NewTokenizer(dirname string) (*Tokenizer, error) {
	p := &Tokenizer{
		dirname:         dirname,
		fileSet:         token.NewFileSet(),
		files:           []string{},
		parseErrors:     []ErrorEntry{},
		Blocks:          []*ast.BlockNode{},
		contentProvider: ioutil.ReadFile,
	}

	var err error
	files, err := ioutil.ReadDir(p.dirname)
	if err != nil {
		return nil, err
	}
	for _, r := range files {
		if strings.HasSuffix(r.Name(), ".snw") {
			p.files = append(p.files, r.Name())
		}
	}
	if len(p.files) == 0 {
		return nil, fmt.Errorf("missing snw file in %s", p.dirname)
	}

	return p, nil
}

func (t *Tokenizer) Lex(yy *yySymType) int {
	t.actualToken = t.next()
	(*yy).raw = t.actualToken
	return t.actualToken.Value
}

func (t *Tokenizer) Error(s string) {
	t.parseErrors = append(t.parseErrors, ErrorEntry{
		Message: s,
		Token:   t.actualToken,
	})
}

func (t *Tokenizer) consumeBlock(n *ast.BlockNode) {
	t.Blocks = append(t.Blocks, n)
}

func (t *Tokenizer) next() *Token {

	if t.file == nil || t.actualToken.baseValue == token.EOF {
		if t.fileIndex < len(t.files) {
			name := filepath.Join(t.dirname, t.files[t.fileIndex])
			t.fileIndex++
			src, err := t.contentProvider(name)
			if err != nil {
				t.Error(fmt.Sprintf("error opening %s. details %s", name, err.Error()))
				return t.next()
			}
			t.file = t.fileSet.AddFile(name, t.fileSet.Base(), len(src)) // register input "file"
			t.scanner.Init(t.file, src, nil /* no error handler */, 0)

		} else {
			return &Token{
				baseValue: token.EOF,
				Value:     -1,
			}
		}

	}

	pos, tok, val := t.scanner.Scan()
	position := t.fileSet.Position(pos)
	t.actualToken = &Token{
		baseValue: tok,
		Position:  position.String(),
		RawString: val,
	}
	switch t.actualToken.baseValue {
	case token.IDENT:
		t.actualToken.Value = IDENT
	case token.LBRACE:
		t.actualToken.Value = LBRACE
	case token.RBRACE:
		t.actualToken.Value = RBRACE

	case token.STRING:
		t.actualToken.Value = STRING
		t.actualToken.RawString = cleanStr(t.actualToken.RawString)
	case token.ASSIGN:
		t.actualToken.Value = ASSIGN
	case token.INT:
		t.actualToken.Value = INTEGER
	case token.FLOAT:
		t.actualToken.Value = FLOAT
	case token.ADD:
		t.actualToken.Value = PLUS
	case token.SUB:
		t.actualToken.Value = MINUS
	case token.PERIOD:
		t.actualToken.Value = DOT
	case token.EOF:
		if t.fileIndex < len(t.files) {
			return t.next()
		}
	case token.SEMICOLON:
		return t.next()
	}
	switch t.actualToken.baseValue {
	case token.IDENT, token.ILLEGAL:
		switch t.actualToken.RawString {
		case "true", "TRUE":
			t.actualToken.Value = TRUE
		case "false", "FALSE":
			t.actualToken.Value = FALSE
		case "$":
			t.actualToken.Value = DOLLAR

		}
	}

	return t.actualToken
}

func cleanStr(str string) string {
	return strings.Trim(str, "\"")
}
