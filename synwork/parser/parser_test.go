package parser

import (
	"testing"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

func TestParseDir01(t *testing.T) {
	p, err := NewParser("dir01")
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse()
	if err != nil {
		t.Fatal(err)
	} else {
		for _, b := range p.Blocks {
			t.Logf("%#v", *b)
		}
	}

}

func TestObjectDataMapSchemaNode(t *testing.T) {
	p, err := NewParser("t_od")
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	schema := map[string]*schema.Schema{
		"required_processor": {
			Type: schema.TypeList,
			Elem: map[string]*schema.Schema{
				"source": {
					Type: schema.TypeString,
				},
				"version": {
					Type:         schema.TypeString,
					DefaultValue: "",
				},
			},
		},
	}
	t.Logf("%#v\n", schema)
}
