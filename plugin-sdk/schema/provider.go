package schema

import "context"

type ProcessorFunc func(context.Context, *ObjectData, interface{}) (interface{}, error)

type Processor struct {
	Schema    map[string]*Schema
	MethodMap map[string]*Method
	InitFunc  ProcessorFunc
}
