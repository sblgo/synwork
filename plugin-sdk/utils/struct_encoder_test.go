package utils

import (
	"fmt"
	"testing"
)

func TestEncoderEncode01(t *testing.T) {
	type (
		TCol01 struct {
			ColumnName         string
			Value              string
			FormatPrintFString string
		}
		TStruct01 struct {
			FileName01 string
			Column     []TCol01 `snw:"column"`
		}
	)
	f := TStruct01{
		FileName01: "fiel01.txt",
		Column: []TCol01{
			{
				ColumnName:         "Id01",
				Value:              "01",
				FormatPrintFString: "%v",
			},
			{
				ColumnName:         "Id01",
				Value:              "01",
				FormatPrintFString: "%v",
			},
		},
	}
	e := NewEncoder()
	res, _ := e.Encode(f)
	mapStr, ok := res.(map[string]interface{})
	if !ok {
		t.Fail()
	}

	colRow, ok := mapStr["column"]
	if !ok {
		t.Fail()
	}
	colArr, ok := colRow.([]interface{})
	if !ok {
		t.Fail()
	}
	if len(colArr) != len(f.Column) {
		t.Fail()
	}
	fmt.Printf("%#v\n", mapStr)
}
