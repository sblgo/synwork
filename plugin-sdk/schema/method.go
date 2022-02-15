package schema

import "context"

type MethodData struct {
	config ObjectData
	result ObjectData
}

func NewMethodData(config ObjectData, result ObjectData) *MethodData {
	return &MethodData{
		config: config,
		result: result,
	}
}

type MethodFunc func(ctx context.Context, data *MethodData, client interface{}) error

type Method struct {
	// Schema contains references to previous Processor data
	Schema      map[string]*Schema
	Result      map[string]*Schema
	Description string
	ExecFunc    MethodFunc
}

func (m *MethodData) GetConfig(path string) interface{} {
	return m.config.Get(path)
}

func (m *MethodData) SetResult(path string, val interface{}) {
	m.result.Set(path, val)
}
