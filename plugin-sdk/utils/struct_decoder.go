package utils

import (
	"fmt"
	"reflect"
	"strings"
)

type (
	Decoder struct {
		tagName        string
		fieldNameToPos map[string]int
		camel          UnderscoreEnc
	}

	Unmarshaller interface {
		UnmarshallStruct(v interface{}) error
	}

	decoder struct {
		Decoder        *Decoder
		parent         *decoder
		fieldNameToPos map[string]int
		name           string
	}
)

var unmarshallerType = reflect.TypeOf((*Unmarshaller)(nil)).Elem()

func NewDecoder() *Decoder {
	return &Decoder{
		tagName:        "snw",
		fieldNameToPos: map[string]int{},
		camel:          NewUnderscoreEnc(),
	}
}

func (d *Decoder) Decode(t interface{}, source interface{}) error {
	hd := &decoder{
		Decoder: d,
	}
	reflectVal := reflect.ValueOf(t)
	return hd.convertReflectIn(reflectVal, source)
}

func (d *decoder) convertReflectIn(reflectValIn reflect.Value, source interface{}) error {
	reflectVal := reflectValIn
	if reflectVal.Kind() == reflect.Pointer {
		reflectVal = reflectVal.Elem()
	}
	if reflectVal.Kind() == reflect.Pointer {
		reflectVal = reflectVal.Elem()
	}
	switch reflectVal.Kind() {
	case reflect.Struct:
		if sourceMap, ok := source.(map[string]interface{}); ok {
			if reflectVal.Addr().Type().Implements(unmarshallerType) {
				if err := reflectVal.Addr().Interface().(Unmarshaller).UnmarshallStruct(sourceMap); err != nil {
					return err
				}
			} else {
				return d.sub(reflectVal).convertReflectStruct(reflectVal, sourceMap)
			}
		}
	case reflect.Array, reflect.Slice:
		requiredElement := true
		elemType := reflectVal.Type().Elem()
		if elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
			requiredElement = false
		}
		if sourceArr, ok := source.([]interface{}); ok {
			for _, sourceArrItem := range sourceArr {
				subReflectVal := reflect.New(elemType)
				if err := d.sub(subReflectVal).convertReflectIn(subReflectVal, sourceArrItem); err != nil {
					return err
				}
				if requiredElement {
					reflectVal = reflect.Append(reflectVal, subReflectVal.Elem())
				} else {
					reflectVal = reflect.Append(reflectVal, subReflectVal)
				}
			}
		}
		reflectValIn.Elem().Set(reflectVal)
	}

	return nil
}

func (d *decoder) sub(modelReflect reflect.Value) *decoder {
	return &decoder{
		parent:         d,
		Decoder:        d.Decoder,
		fieldNameToPos: make(map[string]int),
		name:           modelReflect.String(),
	}
}

func (d *decoder) convertReflectStruct(modelReflect reflect.Value, sourceMap map[string]interface{}) error {
	if modelReflect.Kind() == reflect.Ptr {
		modelReflect = modelReflect.Elem()
	}
	modelRefType := modelReflect.Type()
	fieldsCount := modelReflect.NumField()

	for i := 0; i < fieldsCount; i++ {
		fieldName := d.Decoder.camel.ConvertFieldName(modelRefType.Field(i), d.Decoder.tagName)
		d.fieldNameToPos[fieldName] = i
		fieldName = strings.ReplaceAll(fieldName, "_", "-")
		d.fieldNameToPos[fieldName] = i
	}
	for key, value := range sourceMap {
		if fieldIdx, ok := d.fieldNameToPos[key]; !ok {
			return fmt.Errorf("missing field %s in structure %s", key, modelReflect.Type().Name())
		} else {
			if err := d.setReflectValue(modelReflect.Field(fieldIdx), value); err != nil {
				return CompError(err, " for key %s", key)
			}
		}
	}
	return nil
}

func (d *decoder) setReflectValue(modelReflectIn reflect.Value, value interface{}) error {
	modelReflect := modelReflectIn
	if modelReflect.Kind() == reflect.Pointer && modelReflect.IsNil() {
		modelReflect = reflect.New(modelReflect.Type().Elem())
		modelReflectIn.Set(modelReflect)
		modelReflect = modelReflectIn.Elem()
	} else if modelReflect.Kind() == reflect.Pointer {
		modelReflect = modelReflect.Elem()
	}
	switch modelReflect.Kind() {
	case reflect.Struct:
		if valueMap, ok := value.(map[string]interface{}); ok {
			if modelReflect.Addr().Type().Implements(unmarshallerType) {
				if err := modelReflect.Addr().Interface().(Unmarshaller).UnmarshallStruct(valueMap); err != nil {
					return err
				}
			} else {
				return d.sub(modelReflect).convertReflectStruct(modelReflect, valueMap)
			}
		}
	case reflect.Array, reflect.Slice:
		if valueArr, ok := value.([]interface{}); ok {
			reflectArr := reflect.MakeSlice(modelReflect.Type(), 0, len(valueArr))
			for _, itemValue := range valueArr {
				elemType, takeElem := modelReflect.Type().Elem(), true

				if elemType.Kind() == reflect.Ptr {
					elemType = elemType.Elem()
					takeElem = false

				}
				itemReflect := reflect.New(elemType)
				if takeElem {
					itemReflect = itemReflect.Elem()
				}
				if err := d.setReflectValue(itemReflect, itemValue); err != nil {
					return err
				}
				reflectArr = reflect.Append(reflectArr, itemReflect)
				//				fmt.Printf("item %#v array %#v\n", itemReflect.Interface(), reflectArr.Interface())
			}
			modelReflect.Set(reflectArr)
		}
	default:
		_, ok1 := value.(map[string]interface{})
		_, ok2 := value.([]interface{})
		if !ok1 && !ok2 {
			switch modelReflect.Kind() {
			case reflect.Float32:
				if v, ok := toFloat64(value); ok {
					value = float32(v)
				}
			case reflect.Float64:
				if v, ok := toFloat64(value); ok {
					value = v
				}
			case reflect.Int:
				if v, ok := toInt64(value); ok {
					value = int(v)
				}
			case reflect.Int8:
				if v, ok := toInt64(value); ok {
					value = int8(v)
				}
			case reflect.Int16:
				if v, ok := toInt64(value); ok {
					value = int16(v)
				}
			case reflect.Int32:
				if v, ok := toInt64(value); ok {
					value = int16(v)
				}
			case reflect.Int64:
				if v, ok := toInt64(value); ok {
					value = int16(v)
				}
			case reflect.Uint:
				if v, ok := toInt64(value); ok {
					value = uint(v)
				}
			case reflect.Uint8:
				if v, ok := toInt64(value); ok {
					value = uint8(v)
				}
			case reflect.Uint16:
				if v, ok := toInt64(value); ok {
					value = uint16(v)
				}
			case reflect.Uint32:
				if v, ok := toInt64(value); ok {
					value = uint16(v)
				}
			case reflect.Uint64:
				if v, ok := toInt64(value); ok {
					value = uint16(v)
				}

			}
			newValue := reflect.ValueOf(value)
			modelReflect.Set(newValue)
		}
	}
	return nil
}

func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

func toInt64(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return int64(v), true
	default:
		return 0, false
	}

}

func DecodeSchemaData(target interface{}, source interface{}) error {
	return NewDecoder().Decode(target, source)
}
