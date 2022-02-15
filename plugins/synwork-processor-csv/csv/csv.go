package csv

import (
	"sbl.systems/go/synwork/plugin-sdk/plugin"
	"sbl.systems/go/synwork/plugin-sdk/schema"
)

var Opts = plugin.PluginOptions{
	Provider: func() schema.Processor {
		return schema.Processor{
			InitFunc: csv_init,
			Schema:   map[string]*schema.Schema{},
			MethodMap: map[string]*schema.Method{
				"write": {
					Schema: map[string]*schema.Schema{
						"file_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "file name",
						},
						"data": {
							Type:     schema.TypeList,
							Required: true,
							ElemType: schema.TypeGeneric,
						},
						"column": {
							Type:     schema.TypeList,
							Required: true,
							Elem: map[string]*schema.Schema{
								"path": {
									Type:     schema.TypeString,
									Required: true,
								},
								"name": {
									Type:         schema.TypeString,
									Optional:     true,
									DefaultValue: "",
								},
								"format": {
									Type:         schema.TypeString,
									Optional:     true,
									DefaultValue: "%v",
								},
							},
						},
					},
					Result:   map[string]*schema.Schema{},
					ExecFunc: csv_write,
				},
				"read": {
					Schema: map[string]*schema.Schema{
						"file_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "file name",
						},
						"delimiter": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "file name",
							DefaultValue: ";",
						},
						"column": {
							Type:     schema.TypeList,
							Required: true,
							Elem: map[string]*schema.Schema{
								"name": {
									Type:         schema.TypeString,
									Optional:     true,
									DefaultValue: "",
								},
								"column": {
									Type:     schema.TypeInt,
									Required: true,
								},
							},
						},
						"additional": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: map[string]*schema.Schema{
								"name": {
									Type:         schema.TypeString,
									Optional:     true,
									DefaultValue: "",
								},
								"from_column": {
									Type:     schema.TypeInt,
									Required: true,
								},
								"to_column": {
									Type:         schema.TypeInt,
									Optional:     true,
									DefaultValue: -1,
								},
							},
						},
					},
					Result: map[string]*schema.Schema{
						"data": {
							Type:     schema.TypeList,
							Required: true,
							ElemType: schema.TypeGeneric,
						},
					},
					ExecFunc: csv_read,
				},
			},
		}
	},
}
