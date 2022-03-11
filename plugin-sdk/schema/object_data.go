package schema

import (
	"strconv"
	"strings"
)

type ObjectData struct {
	schema map[string]*Schema
	value  map[string]interface{}
	//	once   sync.Once
}

func NewObjectData(
	schema map[string]*Schema,
	value map[string]interface{},
) *ObjectData {
	return &ObjectData{
		schema: schema,
		value:  value,
	}
}

func (o *ObjectData) Set(key string, value interface{}) error {
	setMap(o.schema, o.value, strings.Split(strings.Trim(key, "/"), "/"), value)
	return nil
}

func setMap(sma map[string]*Schema, value map[string]interface{}, path []string, val interface{}) {
	switch len(path) {
	case 0:
	case 1:
		if smc, ok := sma[path[0]]; ok {
			if _, ok := valueTypeCheck(smc, val); ok {
				value[path[0]] = val
			}
		}
	default:
		if smc, ok := sma[path[0]]; ok {
			switch smc.Type {
			case TypeFloat, TypeInt, TypeString:
			case TypeMap:
				if vc, ok := value[path[0]]; ok {
					if vm, ok := vc.(map[string]interface{}); ok {
						setMap(smc.Elem, vm, path[1:], val)
					}
				}
			case TypeList:
				if vc, ok := value[path[0]]; ok {
					if vl, ok := vc.([]interface{}); ok {
						idxStr := path[1]
						idx, err := strconv.Atoi(idxStr)
						if err != nil {
							panic(err)
						}
						if 0 <= idx && idx < len(vl) {
							if vi, ok := vl[idx].(map[string]interface{}); ok {
								setMap(smc.Elem, vi, path[2:], val)
							}
						} else if idx >= len(vl) {
							for x := len(vl); x <= idx; x++ {
								vl = append(vl, map[string]interface{}{})
							}
							setMap(smc.Elem, vl[idx].(map[string]interface{}), path[2:], val)
						}
					}
				}
			case TypeGeneric:
			}
		}
	}
}

func (o *ObjectData) Get(key string) interface{} {
	path := strings.Split(strings.Trim(key, "/"), "/")
	return getMap(o.schema, o.value, path)
}

func getMap(sma map[string]*Schema, value map[string]interface{}, path []string) interface{} {
	if len(path) > 0 {
		k := path[0]
		if detailSchema, ok := sma[k]; ok {
			detailValue := value[k]
			return getValue(detailSchema, detailValue, path[1:])

		} else {
			return nil
		}
	}
	return nil
}

func getValue(sma *Schema, value interface{}, path []string) interface{} {
	if len(path) == 0 {
		switch sma.Type {
		case TypeMap:
			if vm, ok := value.(map[string]interface{}); ok {
				return vm
			}
		case TypeList:
			if vl, ok := value.([]interface{}); ok {
				return vl
			}
		case TypeString, TypeFloat, TypeInt:
			if vi, ok, _ := defaultValue(sma, value); ok {
				if vs, ok := valueTypeCheck(sma, vi); ok {
					return vs
				}
			}
		case TypeGeneric:
			return value
		}
	} else {
		switch sma.Type {
		case TypeMap:
			if vm, ok := value.(map[string]interface{}); ok {
				return getMap(sma.Elem, vm, path)
			}
		case TypeList:
			if vl, ok := value.([]interface{}); ok {
				idxStr := path[0]
				if idx, err := strconv.Atoi(idxStr); err != nil {
					panic(err)
				} else if idx < 0 {
					panic("index < 0")
				} else if idx > len(vl) {
					return nil
				} else if vc, ok := vl[idx].(map[string]interface{}); ok {
					return getMap(sma.Elem, vc, path[1:])
				} else {
					panic("element type not map[string]interface{}")
				}
			}
		case TypeGeneric:
			vi, _ := GetValueMap(value, path)
			return vi
		}
	}
	return nil
}

func defaultValue(sma *Schema, val interface{}) (interface{}, bool, error) {
	if val == nil && sma.Optional && sma.DefaultValue != nil {
		return sma.DefaultValue, true, nil
	} else if val == nil && sma.Optional && sma.DefaultFunc != nil {
		return sma.DefaultFunc, true, nil
	} else if val == nil && sma.Required {
		return nil, false, nil
	} else if val != nil {
		return val, true, nil
	}
	return nil, false, nil
}

func valueTypeCheck(sma *Schema, val interface{}) (interface{}, bool) {
	switch sma.Type {
	case TypeList:
		vf, ok := val.([]interface{})
		return vf, ok
	case TypeMap:
		vf, ok := val.(map[string]interface{})
		return vf, ok
	case TypeFloat:
		vf, ok := val.(float64)
		return vf, ok
	case TypeInt:
		vi, ok := val.(int)
		return vi, ok
	case TypeString:
		vs, ok := val.(string)
		return vs, ok
	case TypeGeneric:
		return val, true
	default:
		return nil, false
	}
}
