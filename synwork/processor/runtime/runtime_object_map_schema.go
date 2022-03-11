package runtime

import (
	"fmt"

	"sbl.systems/go/synwork/plugin-sdk/schema"
	"sbl.systems/go/synwork/synwork/ast"
)

func mapSchemaAndNode(path string, sma map[string]*schema.Schema, n ast.BlockContentNode) (map[string]interface{}, []*Reference, error) {
	value := map[string]interface{}{}
	references := []*Reference{}
	for _, i := range n.Assignments {
		if s, ok := sma[i.Identifier]; ok {
			switch v := i.Value.(type) {
			case *ast.StringValue:
				if s.Type == schema.TypeString {
					value[i.Identifier] = v.Value
				}
			case *ast.IntValue:
				if s.Type == schema.TypeInt {
					value[i.Identifier] = v.Value
				}

			case *ast.FloatValue:
				if s.Type == schema.TypeFloat {
					value[i.Identifier] = v.Value
				}
			case *ast.ComplexValue:
				switch s.Type {
				case schema.TypeList:
					if _, ok := value[i.Identifier]; !ok {
						value[i.Identifier] = []interface{}{}
					}
					list := value[i.Identifier].([]interface{})
					obj, ref2, err := mapSchemaAndNode(fmt.Sprintf("%s/%d/%s", path, len(list), i.Identifier), s.Elem, v.BlockContentNode)
					if err != nil {
						return nil, nil, err
					}
					references = append(references, ref2...)
					list = append(list, obj)
					value[i.Identifier] = list
				case schema.TypeMap:
					obj, ref2, err := mapSchemaAndNode(fmt.Sprintf("%s/%s", path, i.Identifier), s.Elem, v.BlockContentNode)
					if err != nil {
						return nil, nil, err
					}
					references = append(references, ref2...)
					if _, ok := value[i.Identifier]; !ok {
						value[i.Identifier] = obj
					}
				}
			case *ast.ReferenceValue:
				ref, err := NewReference(v.RefParts[0], v.RefParts[1:], fmt.Sprintf("%s/%s", path, i.Identifier), *s)
				if err != nil {
					return nil, nil, err
				}
				value[i.Identifier] = ref
				references = append(references, ref)
			}
		} else {
			return nil, nil, fmt.Errorf("[%s - %s] invalid key %s ", n.Begin, n.End, i.Identifier)
		}
	}
	for _, i := range n.Blocks {
		key := i.Type
		if s, ok := sma[key]; ok {
			switch s.Type {
			case schema.TypeList:
				if _, ok := value[key]; !ok {
					value[key] = []interface{}{}
				}
				list := value[key].([]interface{})
				obj, ref, err := mapSchemaAndNode(fmt.Sprintf("%s/%s/%d", path, key, len(list)), s.Elem, *i.Content)
				if err != nil {
					return nil, nil, err
				}
				references = append(references, ref...)
				list = append(list, obj)
				value[key] = list
			case schema.TypeMap:
				obj, ref, err := mapSchemaAndNode(fmt.Sprintf("%s/%s", path, key), s.Elem, *i.Content)
				if err != nil {
					return nil, nil, err
				}
				references = append(references, ref...)
				if _, ok := value[key]; !ok {
					value[key] = obj
				}
			}
		} else {
			return nil, nil, fmt.Errorf("[%s - %s] invalid key %s ", n.Begin, n.End, i.Type)
		}
	}
	for k, s := range sma {
		if _, ok := value[k]; !ok {
			switch s.Type {
			case schema.TypeString, schema.TypeInt:
				if s.DefaultValue != nil {
					value[k] = s.DefaultValue
				} else if s.DefaultFunc != nil {
					value[k] = s.DefaultFunc
				} else if s.Required {
					return nil, nil, fmt.Errorf("missing default value for required field %s", k)
				}
			}
		}
	}

	return value, references, nil

}

func MapSchemaAndNode(sma map[string]*schema.Schema, n ast.BlockContentNode) (*RuntimeObject, error) {
	inMap, references, err := mapSchemaAndNode("", sma, n)
	if err != nil {
		return nil, err
	}
	runObj := NewRuntimeObject(sma, inMap, references)
	return runObj, nil
}
