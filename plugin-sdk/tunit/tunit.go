package tunit

import (
	"context"
	"strings"
	"testing"

	"sbl.systems/go/synwork/plugin-sdk/schema"
	"sbl.systems/go/synwork/synwork/parser"
	"sbl.systems/go/synwork/synwork/processor/runtime"
)

type MethodMock struct {
	ProcessorDef func() schema.Processor
	InstanceMock interface{}
	ExecFunc     schema.MethodFunc
	References   map[string]interface{}
}

func CallMockMethod(t *testing.T, mm MethodMock, defs string) map[string]interface{} {
	p, err := parser.NewParserForTest("test", defs)
	if err != nil {
		t.Fatal(err)
	}
	err = p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Blocks) != 1 {
		t.Fatal("invalid block")
	}
	b := p.Blocks[0]
	if b.Type == "method" {
		if len(b.Identifiers) != 3 {
			t.Fatalf("invalid definition for processor missing identifiers at %s", b.Pos())
		}
		methodName, instanceName, _ := b.Identifiers[0], b.Identifiers[1], b.Identifiers[2]
		procDef := mm.ProcessorDef()

		if methodDef, ok := procDef.MethodMap[methodName]; !ok {
			t.Fatalf("method %s for processor %s of type %s isn't defined. (%s)", methodName, instanceName, "<testplugin>", b.Pos())
		} else {
			runtimeObj, err := runtime.MapSchemaAndNode(methodDef.Schema, *b.Content)
			if err != nil {
				t.Fatalf("method %s for processor %s of type %s has wrong definition. (%s)", methodName, instanceName, "<testplugin>", err.Error())
			}
			dataMap, err := createObjectData(mm, runtimeObj)
			results := map[string]interface{}{}

			data := schema.NewMethodData(*schema.NewObjectData(methodDef.Schema, dataMap), *schema.NewObjectData(methodDef.Result, results))

			err = methodDef.ExecFunc(context.Background(), data, mm.InstanceMock)
			if err != nil {
				t.Fatalf("method %s for processor %s of type %s call failed. (%s)", methodName, instanceName, "<testplugin>", err.Error())
			}
			return results
		}

	}
	return map[string]interface{}{}
}

func createObjectData(mm MethodMock, runObj *runtime.RuntimeObject) (map[string]interface{}, error) {
	newValue := schema.MapCopy(runObj.Value)
	objData := newValue
	for _, ref := range runObj.ReferenceMap {
		switch ref.Type {
		case runtime.ReferenceTypeProcessor, runtime.ReferenceTypeMethod:
			path := append([]string{"method", ref.Key}, ref.Path...)
			if val, ok := schema.GetValueMap(mm.References, path); ok {
				tgtPath := strings.Split(strings.Trim(ref.Location, "/"), "/")
				schema.SetValueMap(newValue, tgtPath, val)
			}
		case runtime.ReferenceTypeVariable:
		case runtime.ReferenceTypeEnvironment:
		}

	}
	return objData, nil
}
