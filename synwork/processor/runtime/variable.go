package runtime

type Variable struct {
	Name         string
	objectData   *RuntimeObject
	defaultValue interface{}
	valueFunc    func(RuntimeContext) interface{}
}

func NewVariable(name string, data *RuntimeObject) *Variable {

	variable := &Variable{
		Name:       name,
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
