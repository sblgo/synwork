package schema

type SchemaType uint

type SchemaDefaultFunc interface{}

const (
	TypeUndefined SchemaType = 0
	TypeString    SchemaType = 1
	TypeInt       SchemaType = 2
	TypeFloat     SchemaType = 3
	TypeBool      SchemaType = 4
	TypeMap       SchemaType = 5
	TypeList      SchemaType = 6
	TypeGeneric   SchemaType = 7
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

var schemaTypeString = map[SchemaType]string{
	TypeUndefined: "TypeUndefined",
	TypeString:    "TypeString",
	TypeInt:       "TypeInt",
	TypeFloat:     "TypeFloat",
	TypeBool:      "TypeBool",
	TypeMap:       "TypeMap",
	TypeList:      "TypeList",
	TypeGeneric:   "TypeGeneric",
}

func (st SchemaType) String() string {
	if v, ok := schemaTypeString[st]; ok {
		return v
	} else {
		return "Undefined"
	}
}
