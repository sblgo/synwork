package runtime

import (
	"sync"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

type RuntimeObject struct {
	Schema       map[string]*schema.Schema
	Value        map[string]interface{}
	ReferenceMap map[string]*Reference
	once         sync.Once
}

func NewRuntimeObject(
	schema map[string]*schema.Schema,
	value map[string]interface{},
	references []*Reference,
) *RuntimeObject {
	referenceMap := map[string]*Reference{}
	for _, ref := range references {
		referenceMap[ref.Location] = ref
	}
	return &RuntimeObject{
		Schema:       schema,
		Value:        value,
		ReferenceMap: referenceMap,
	}
}
