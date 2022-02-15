package runtime

import "sbl.systems/go/synwork/plugin-sdk/schema"

type Processor struct {
	Id         string
	PluginName string
	plugin     *Plugin
	schema     map[string]*schema.Schema
	data       *RuntimeObject
}
