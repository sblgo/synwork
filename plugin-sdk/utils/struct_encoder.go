package utils

import (
	"reflect"
	"regexp"
	"strings"
)

type Encoder struct {
	tagName string
	camel   UnderscoreEnc
}

func NewEncoder() *Encoder {
	return &Encoder{
		tagName: "snw",
		camel:   NewUnderscoreEnc(),
	}
}

func (e *Encoder) Encode(obj interface{}) (interface{}, error) {
	ret := e.convertToIn(obj)
	return ret, nil
}

func (e *Encoder) convertToIn(model interface{}) interface{} {
	modelReflect := reflect.ValueOf(model)

	if modelReflect.Kind() == reflect.Ptr {
		modelReflect = modelReflect.Elem()
	}
	return e.convertReflectValue(modelReflect)
}

func (e *Encoder) convertReflectValue(modelReflect reflect.Value) interface{} {
	if modelReflect.Kind() == reflect.Ptr {
		modelReflect = modelReflect.Elem()
	}
	switch modelReflect.Kind() {
	case reflect.Ptr:
		return e.convertReflectValue(modelReflect.Elem())
	case reflect.Array, reflect.Slice:
		return e.convertReflectArray(modelReflect)
	case reflect.Struct:
		return e.convertReflectStruct(modelReflect)
	default:
		return modelReflect.Interface()
	}

}

func (e *Encoder) convertReflectArray(modelReflect reflect.Value) []interface{} {
	ret := []interface{}{}
	for idx := 0; idx < modelReflect.Len(); idx++ {
		value := e.convertReflectValue(modelReflect.Index(idx))
		ret = append(ret, value)

	}
	return ret
}

func (e *Encoder) convertReflectStruct(modelReflect reflect.Value) map[string]interface{} {
	ret := map[string]interface{}{}
	if modelReflect.Kind() == reflect.Ptr {
		modelReflect = modelReflect.Elem()
	}

	modelRefType := modelReflect.Type()
	fieldsCount := modelReflect.NumField()

	var fieldData interface{}

	for i := 0; i < fieldsCount; i++ {
		field := modelReflect.Field(i)

		switch field.Kind() {
		case reflect.Struct:
			fieldData = e.convertReflectStruct(field)
		case reflect.Ptr:
			fieldData = e.convertReflectValue(field)
		case reflect.Array, reflect.Slice:
			fieldData = e.convertReflectArray(field)
		default:
			fieldData = field.Interface()
		}
		fieldName := e.camel.ConvertFieldName(modelRefType.Field(i), e.tagName)
		ret[fieldName] = fieldData
	}

	return ret
}

func (u UnderscoreEnc) ConvertFieldName(field reflect.StructField, tagName string) string {
	name := field.Tag.Get(tagName)
	if name != "" {
		return name
	}
	name = u.underscore(field.Name)

	return name
}

type UnderscoreEnc struct {
	camel *regexp.Regexp
}

func NewUnderscoreEnc() UnderscoreEnc {
	return UnderscoreEnc{
		camel: regexp.MustCompile("(^[^A-Z0-9]*|[A-Z0-9]*)([A-Z0-9][^A-Z]+|$)"),
	}
}

func (u UnderscoreEnc) underscore(s string) string {
	var a []string
	for _, sub := range u.camel.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}
	return strings.ToLower(strings.Join(a, "_"))
}
