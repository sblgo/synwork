package runtime

import "sbl.systems/go/synwork/plugin-sdk/schema"

var VariableSchema map[string]*schema.Schema = map[string]*schema.Schema{
	"type": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "describes the type of the variable accepted values are string,int,float",
	},
	"default": {
		Type:        schema.TypeGeneric,
		Required:    true,
		Description: "set the value, if no other kind works",
	},
	"param-name": {
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "the name of the program parameter, which should used here, if exists",
		DefaultValue: "",
	},
	"env-name": {
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "the name of an environment variable, which would used here, if exists and no program param is set",
		DefaultValue: "",
	},
}
