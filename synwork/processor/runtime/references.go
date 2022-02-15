package runtime

import (
	"fmt"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

type ReferenceType string

const (
	ReferenceTypeMethod      ReferenceType = ReferenceType("method")
	ReferenceTypeVariable    ReferenceType = ReferenceType("variable")
	ReferenceTypeProcessor   ReferenceType = ReferenceType("processor")
	ReferenceTypePlugin      ReferenceType = ReferenceType("plugin")
	ReferenceTypeEnvironment ReferenceType = ReferenceType("environment")
)

var referenceTypes map[string]ReferenceType = map[string]ReferenceType{
	"method":    ReferenceTypeMethod,
	"variable":  ReferenceTypeVariable,
	"processor": ReferenceTypeProcessor,
	"plugin":    ReferenceTypePlugin,
	"env":       ReferenceTypeEnvironment,
}

func (r ReferenceType) Id(k string) string {
	return fmt.Sprintf("%s/%s", r, k)
}

type Reference struct {
	Type     ReferenceType
	Key      string
	Path     []string
	Location string
	Schema   schema.Schema
}

func NewReference(t string, path []string, loc string, sma schema.Schema) (*Reference, error) {
	if rt, ok := referenceTypes[t]; !ok {
		return nil, fmt.Errorf("unknown reference type %s", t)
	} else {
		switch len(path) {
		case 0:
			return nil, fmt.Errorf("reference has no identifier")
		default:
			return &Reference{
				Type:     rt,
				Key:      path[0],
				Path:     path[1:],
				Location: loc,
				Schema:   sma,
			}, nil
		}
	}
}
