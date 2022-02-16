package parser

import "testing"

func TestTokenizer01(t *testing.T) {
	tonz, err := NewTokenizer("tok01")
	if err != nil {
		t.Fatal(err)
	}
	for {
		yy := &yySymType{}
		if tonz.Lex(yy) != -1 {
			t.Log(yy.raw.Position, yy.raw.baseValue, yy.raw.Value, yy.raw.RawString)
		} else {
			return
		}
	}
}
