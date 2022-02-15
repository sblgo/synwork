package randlist

import (
	"sbl.systems/go/synwork/plugin-sdk/plugin"
	"sbl.systems/go/synwork/plugin-sdk/schema"
)

var Opts = plugin.PluginOptions{
	Provider: func() schema.Processor {
		return schema.Processor{

			Schema: map[string]*schema.Schema{
				"random_type": {
					Type:         schema.TypeString,
					DefaultValue: "Int31",
					Optional:     true,
				},

				"seed": {
					Type:         schema.TypeInt,
					DefaultValue: 100,
					Optional:     true,
				},
			},
			MethodMap: map[string]*schema.Method{
				"random_list": {
					Schema: map[string]*schema.Schema{
						"min_id": {
							Type:         schema.TypeInt,
							DefaultValue: 1,
							Optional:     true,
						},
						"max_id": {
							Type:         schema.TypeInt,
							DefaultValue: 10,
							Optional:     true,
						},
					},
					Description: "creates a list with random data",
					Result: map[string]*schema.Schema{
						"result": {
							Type: schema.TypeList,
							Elem: map[string]*schema.Schema{
								"id": {
									Type: schema.TypeInt,
								},
								"value": {
									Type: schema.TypeInt,
								},
							},
						},
					},
					ExecFunc: random_list,
				},
			},
			InitFunc: random_init,
		}
	},
}
