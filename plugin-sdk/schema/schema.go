package schema

type SchemaType uint

type SchemaDefaultFunc func() (interface{}, error)

const (
	TypeUndefined SchemaType = 0
	TypeString    SchemaType = 1
	TypeInt       SchemaType = 2
	TypeFloat     SchemaType = 3
	TypeMap       SchemaType = 4
	TypeList      SchemaType = 5
	TypeGeneric   SchemaType = 6
)

type Schema struct {
	Type         SchemaType
	Optional     bool
	Required     bool
	DefaultValue interface{}
	DefaultFunc  SchemaDefaultFunc
	Description  string
	Elem         map[string]*Schema
	ElemType     SchemaType
}
