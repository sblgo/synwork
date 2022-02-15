package runtime

import "sbl.systems/go/synwork/plugin-sdk/schema"

var PluginSchema map[string]*schema.Schema = map[string]*schema.Schema{
	"required_processor": {
		Type: schema.TypeList,
		Elem: map[string]*schema.Schema{
			"source": {
				Type:     schema.TypeString,
				Required: true,
			},
			"version": {
				Type:         schema.TypeString,
				DefaultValue: "",
				Optional:     true,
			},
			"name": {
				Type:         schema.TypeString,
				DefaultValue: "",
				Optional:     true,
			},
		},
	},
}
