package runtime

import "os"

type Variable struct {
	Id           string
	objectData   *RuntimeObject
	defaultValue interface{}
	valueFunc    func(RuntimeContext) interface{}
}

func NewVariable(name string, pos string, data *RuntimeObject) *Variable {

	variable := &Variable{
		Id:         name,
		objectData: data,
	}

	return variable
}

func (v *Variable) Get(rc RuntimeContext) interface{} {
	if v.valueFunc != nil {
		return v.valueFunc(rc)
	}
	return v.defaultValue
}

func (v *Variable) Eval(rc *ExecContext, data map[string]interface{}) (interface{}, error) {
	var val interface{}
	val = nil
	destType := data["type"].(string)
	paramName := data["param-name"].(string)
	if tmpVal, ok := rc.Parameters[paramName]; ok && paramName != "" {
		if tmpVal, err := v.castType(destType, tmpVal); err == nil {
			val = tmpVal
		}
	}
	if val == nil {
		envName := data["env-name"].(string)
		strVal := os.Getenv(envName)
		if strVal != "" {
			if tmpVal, err := v.castType(destType, strVal); err == nil {
				val = tmpVal
			}
		}
	}
	if val == nil {
		val = data["default"]
	}

	return val, nil
}

func (v *Variable) castType(typeName string, value interface{}) (interface{}, error) {

	return value, nil
}
