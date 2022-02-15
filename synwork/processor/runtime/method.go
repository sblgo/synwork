package runtime

import "sbl.systems/go/synwork/plugin-sdk/schema"

type Method struct {
	Id        string
	Name      string
	Instance  string
	Data      *RuntimeObject
	Processor *Processor
	Plugin    *Plugin
	Schema    schema.Method
}
