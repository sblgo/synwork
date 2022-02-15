package runtime

import "sbl.systems/go/synwork/plugin-sdk/schema"

var VariableSchema map[string]*schema.Schema = map[string]*schema.Schema{
	"type": {
		Type: schema.TypeString,
	},
	"default": {
		Type: schema.TypeGeneric,
	},
}
