%{
package parser

import "sbl.systems/go/synwork/synwork/ast"
import "log"
//import "strconv"

func print(s string) {
  log.Println(s)
}

%}

%union {
    raw     *Token
    node    ast.Node
    array   []interface{}
    strar   []string
}


%token <raw> IDENT
%token <raw> LBRACE
%token <raw> RBRACE
%token <raw> STRING
%token <raw> ASSIGN
%token <raw> INTEGER
%token <raw> FLOAT
%token <raw> MINUS
%token <raw> PLUS
%token <raw> DOLLAR
%token <raw> DOT
%token <raw> TRUE
%token <raw> FALSE

%type <node> blocks
%type <node> named_block
%type <node> block_body
%type <node> block_entry
%type <node> unnamed_block
%type <node> value
%type <node> numeric_value
%type <node> reference
%type <strar> reference_parts
%type <node> boolean_value
%type <strar> string_list

%%

blocks:
    { }
    |
    blocks named_block
    { yylex.(*Tokenizer).consumeBlock($2.(*ast.BlockNode)) }


named_block:
    IDENT string_list unnamed_block
    { 
        bcn := $3.(*ast.ComplexValue)    
        $$ = &ast.BlockNode{Begin: $1.Position, Type: $1.RawString, Identifiers: $2, Content: &bcn.BlockContentNode}
    }
    
unnamed_block:
    LBRACE block_body RBRACE
    { $$ = &ast.ComplexValue{*($2.(*ast.BlockContentNode))} }

block_body:
    block_body block_entry 
    { 
        bcn := $1.(*ast.BlockContentNode)
        switch n := $2.(type) {
            case *ast.AssignmentNode:
            bcn.Assignments = append(bcn.Assignments,n)
            case *ast.BlockNode:
            bcn.Blocks = append(bcn.Blocks,n)
        }
        $$ = bcn
    }
    |
    { $$ = &ast.BlockContentNode{ Assignments: []*ast.AssignmentNode{},Blocks: []*ast.BlockNode{} } }

block_entry:
    IDENT unnamed_block
    { 
        bcn := $2.(*ast.ComplexValue)    
        $$ = &ast.BlockNode{Begin: $1.Position, Type: $1.RawString, Content: &bcn.BlockContentNode } 
    }
    |
    IDENT ASSIGN value
    { $$ = &ast.AssignmentNode{ Begin: $1.Position, Identifier: $1.RawString, Value: $3.(ast.ValueNode) } }

value:
    unnamed_block
    { $$ = $1 }
    |
    STRING
    { $$ = &ast.StringValue{Begin: $1.Position, Value: $1.RawString}}
    |
    numeric_value
    { $$ = $1 }
    |
    boolean_value
    { $$ = $1 }
    |
    reference
    { $$ = $1 }

numeric_value:
    INTEGER
    { $$ = &ast.IntValue{Begin:$1.Position, Value: convertIntValue(1,$1.RawString)} }
    |
    PLUS INTEGER
    { $$ = &ast.IntValue{Begin:$1.Position, Value:  convertIntValue(1,$2.RawString)} }
    |
    MINUS INTEGER
    { $$ = &ast.IntValue{Begin:$1.Position, Value:  convertIntValue(-1,$2.RawString)} }
    |
    FLOAT
    { $$ = &ast.FloatValue{Begin:$1.Position, Value:  convertFloatValue(1,$1.RawString)} }
    |
    PLUS FLOAT
    { $$ = &ast.FloatValue{Begin:$1.Position, Value:  convertFloatValue(1,$2.RawString)} }
    |
    MINUS FLOAT
    { $$ = &ast.FloatValue{Begin:$1.Position, Value:  convertFloatValue(-1,$2.RawString)} }

reference:
    DOLLAR reference_parts
    { $$ = &ast.ReferenceValue{Begin:$1.Position,RefParts:$2}}

reference_parts:
    reference_parts DOT IDENT
    { $$ = append($1,$3.RawString) }
    |
    IDENT
    { $$ = []string{$1.RawString}}


boolean_value:
    TRUE
    { $$ = &ast.BoolValue{Begin:$1.Position,Value:true} }
    |
    FALSE
    { $$ = &ast.BoolValue{Begin:$1.Position,Value:false} }


string_list:
    string_list STRING
    { $$ = append($1,$2.RawString) }
    |
    { $$ = []string{} }
