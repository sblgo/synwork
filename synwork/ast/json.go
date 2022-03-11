package ast

import (
	"fmt"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

func MapSchemaAndNode2(sma map[string]*schema.Schema, n BlockContentNode) (map[string]interface{}, []*ReferenceValue, error) {
	value := map[string]interface{}{}
	references := []*ReferenceValue{}
	for _, i := range n.Assignments {
		if s, ok := sma[i.Identifier]; ok {
			switch v := i.Value.(type) {
			case *StringValue:
				if s.Type == schema.TypeString {
					value[i.Identifier] = v.Value
				}
			case *ComplexValue:
				obj, ref2, err := MapSchemaAndNode2(s.Elem, v.BlockContentNode)
				if err != nil {
					return nil, nil, err
				}
				references = append(references, ref2...)
				switch s.Type {
				case schema.TypeList:
					if _, ok := value[i.Identifier]; !ok {
						value[i.Identifier] = []interface{}{}
					}
					list := value[i.Identifier].([]interface{})
					list = append(list, obj)
					value[i.Identifier] = list
				case schema.TypeMap:
					if _, ok := value[i.Identifier]; !ok {
						value[i.Identifier] = obj
					}
				}
			case *ReferenceValue:
			}
		} else {
			return nil, nil, fmt.Errorf("[%s - %s] invalid key %s ", n.Begin, n.End, i.Identifier)
		}
	}
	for _, i := range n.Blocks {
		key := i.Type
		if s, ok := sma[key]; ok {
			obj, ref, err := MapSchemaAndNode2(s.Elem, *i.Content)
			if err != nil {
				return nil, nil, err
			}
			references = append(references, ref...)
			switch s.Type {
			case schema.TypeList:
				if _, ok := value[key]; !ok {
					value[key] = []interface{}{}
				}
				list := value[key].([]interface{})
				list = append(list, obj)
				value[key] = list
			case schema.TypeMap:
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
			case schema.TypeString:
				if s.DefaultValue != nil {
					value[k] = s.DefaultValue.(string)
				} else if s.DefaultFunc != nil {
					value[k] = s.DefaultFunc.(string)
				}
			}
		}
	}

	return value, references, nil
}
