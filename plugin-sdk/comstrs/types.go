package comstrs

import "sbl.systems/go/synwork/plugin-sdk/schema"

type PluginSchemaIn struct {
	Env map[string]interface{}
}

type PluginSchemaOut struct {
	Provider schema.Processor
}

type PluginProcessorIn struct {
	InstanceName string
	Data         map[string]interface{}
}

type PluginProcessorOut struct {
	Data map[string]interface{}
}

type PluginMethodIn struct {
	InstanceName string
	MethodName   string
	Data         map[string]interface{}
}

type PluginMethodOut struct {
	Result map[string]interface{}
}
