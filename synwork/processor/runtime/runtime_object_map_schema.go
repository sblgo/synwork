package runtime

import (
	"fmt"

	"sbl.systems/go/synwork/plugin-sdk/schema"
	"sbl.systems/go/synwork/plugin-sdk/utils"
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
						return nil, nil, utils.CompError(err, "invalid value for %s (at %s - %s)", i.Identifier, i.Begin, i.End)
					}
					references = append(references, ref2...)
					list = append(list, obj)
					value[i.Identifier] = list
				case schema.TypeMap:
					obj, ref2, err := mapSchemaAndNode(fmt.Sprintf("%s/%s", path, i.Identifier), s.Elem, v.BlockContentNode)
					if err != nil {
						return nil, nil, utils.CompError(err, "invalid value for %s (at %s - %s)", i.Identifier, i.Begin, i.End)
					}
					references = append(references, ref2...)
					if _, ok := value[i.Identifier]; !ok {
						value[i.Identifier] = obj
					}
				}
			case *ast.ReferenceValue:
				ref, err := NewReference(v.RefParts[0], v.RefParts[1:], fmt.Sprintf("%s/%s", path, i.Identifier), *s)
				if err != nil {
					return nil, nil, utils.CompError(err, "invalid reference for %s (at %s - %s)", i.Identifier, i.Begin, i.End)
				}
				value[i.Identifier] = ref
				references = append(references, ref)
			}
		} else {
			return nil, nil, fmt.Errorf("[%s - %s] invalid key %s (mapSchemaAndNode2.1)", n.Begin, n.End, i.Identifier)
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
					return nil, nil, utils.CompError(err, "invalid value (at %s - %s)", i.Begin, i.End)
				}
				references = append(references, ref...)
				list = append(list, obj)
				value[key] = list
			case schema.TypeMap:
				obj, ref, err := mapSchemaAndNode(fmt.Sprintf("%s/%s", path, key), s.Elem, *i.Content)
				if err != nil {
					return nil, nil, utils.CompError(err, "invalid value (at %s - %s)", i.Begin, i.End)
				}
				references = append(references, ref...)
				if _, ok := value[key]; !ok {
					value[key] = obj
				}
			}
		} else {
			return nil, nil, fmt.Errorf("[%s - %s] invalid key %s (mapSchemaAndNode2.2)", n.Begin, n.End, i.Type)
		}
	}
	for k, s := range sma {
		if _, ok := value[k]; !ok {
			switch s.Type {
			case schema.TypeString, schema.TypeInt, schema.TypeBool, schema.TypeFloat:
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

type mapSchemaIn struct {
	parent      *mapSchemaIn
	inValue     interface{}
	inMapValue  map[string]interface{}
	inArrValue  []interface{}
	outValue    interface{}
	outMapValue map[string]interface{}
	outArrValue []interface{}
	sma         *schema.Schema
}

func (m *mapSchemaIn) mapValues() error {
	switch m.sma.Type {
	case schema.TypeGeneric:
		m.outValue = m.inValue
	case schema.TypeList:
		return m.mapListValue()
	case schema.TypeMap:
		return m.mapMapValue()
	default:
		return m.mapSkalarValue()
	}
	return nil
}

func (m *mapSchemaIn) mapMapValue() error {
	if l, ok := m.inValue.(map[string]interface{}); ok {
		m.inMapValue = l
	} else {
		return fmt.Errorf("expected map but get it isn't a map")
	}
	newValue := map[string]interface{}{}
	for subName, subField := range m.sma.Elem {
		cm := mapSchemaIn{
			parent:  m,
			sma:     subField,
			inValue: m.inMapValue[subName],
		}
		if err := cm.mapValues(); err != nil {
			return utils.CompError(err, "field '%s'", subName)
		}
		newValue[subName] = cm.outValue
	}
	m.outMapValue = newValue
	m.outValue = m.outMapValue
	return nil
}

func (m *mapSchemaIn) mapListValue() error {
	if l, ok := m.inValue.([]interface{}); ok {
		m.inArrValue = l
	} else if m.inValue == nil {
		return nil
	} else {
		return fmt.Errorf("expected list but get it isn't a list")
	}
	m.outArrValue = make([]interface{}, 0)
	m.outValue = m.outArrValue
	for _, v := range m.inArrValue {
		if m.sma.ElemType == schema.TypeMap || m.sma.ElemType == schema.TypeUndefined {
			inMapValue := v.(map[string]interface{})
			newValue := map[string]interface{}{}
			for subName, subField := range m.sma.Elem {
				cm := mapSchemaIn{
					parent:  m,
					sma:     subField,
					inValue: inMapValue[subName],
				}
				if err := cm.mapValues(); err != nil {
					return utils.CompError(err, "field '%s'", subName)
				}
				newValue[subName] = cm.outValue
			}
			m.outArrValue = append(m.outArrValue, newValue)
		} else {
			cm := mapSchemaIn{
				parent: m,
				sma: &schema.Schema{
					Type: m.sma.ElemType,
				},
				inValue: v,
			}
			if err := cm.mapValues(); err != nil {
				return err
			}
			m.outArrValue = append(m.outArrValue, cm.outValue)
		}
	}
	m.outValue = m.outArrValue
	return nil
}

func (m *mapSchemaIn) mapSkalarValue() error {
	if m.inValue == nil {
		if m.sma.Required {
			return fmt.Errorf("missing required value")
		} else if m.sma.Optional {
			if m.sma.DefaultFunc != nil {
				m.outValue = m.sma.DefaultFunc
			} else {
				m.outValue = m.sma.DefaultValue
			}
			// if m.outValue == nil {
			// 	return fmt.Errorf("missing required default value")
			// }
		}
		return nil
	}
	switch t := m.inValue.(type) {
	case string:
		if m.sma.Type == schema.TypeString {
			m.outValue = t
		} else {
			return fmt.Errorf("invalid type expected %s but get string", m.sma.Type)
		}
	case int:
		if m.sma.Type == schema.TypeInt {
			m.outValue = int(t)
		} else {
			return fmt.Errorf("invalid type expected %s but get int", m.sma.Type)
		}
	case int64:
		if m.sma.Type == schema.TypeInt {
			m.outValue = int(t)
		} else {
			return fmt.Errorf("invalid type expected %s but get int", m.sma.Type)
		}
	case int32:
		if m.sma.Type == schema.TypeInt {
			m.outValue = int(t)
		} else {
			return fmt.Errorf("invalid type expected %s but get int", m.sma.Type)
		}
	case bool:
		if m.sma.Type == schema.TypeBool {
			m.outValue = bool(t)
		} else {
			return fmt.Errorf("invalid type expected %s but get bool", m.sma.Type)
		}
	case float32:
		if m.sma.Type == schema.TypeFloat {
			m.outValue = float64(t)
		} else {
			return fmt.Errorf("invalid type expected %s but get float32", m.sma.Type)
		}
	case float64:
		if m.sma.Type == schema.TypeFloat {
			m.outValue = float64(t)
		} else {
			return fmt.Errorf("invalid type expected %s but get float64", m.sma.Type)
		}
	default:
		return fmt.Errorf("invalid type")
	}
	return nil
}

func MapSchemaAndInterface(sma *schema.Schema, v interface{}) (interface{}, error) {
	mapper := &mapSchemaIn{
		inValue: v,
		sma:     sma,
	}
	if err := mapper.mapValues(); err != nil {
		return nil, err
	}
	return mapper.outValue, nil
}
