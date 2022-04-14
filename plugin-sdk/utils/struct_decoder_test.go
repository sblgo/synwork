package utils

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestDecoderDecode01(t *testing.T) {
	type (
		DSub01 struct {
			Name  string
			Large float64
			Mix   []int
		}
		DSub02 struct {
			Key   string
			Value int
		}
		DTest01 struct {
			FileName   string
			FieldIndex int
			Detail     DSub01
			Properties []*DSub02
		}
	)
	var d DTest01
	source := map[string]interface{}{
		"file_name":   "filetest.txt",
		"field_index": 10,
		"detail": map[string]interface{}{
			"name":  "width",
			"large": 2.34,
			"mix":   []interface{}{2, 3, 4},
		},
		"properties": []interface{}{
			map[string]interface{}{
				"key":   "length",
				"value": 3,
			},
		},
	}
	dec := NewDecoder()
	dec.Decode(&d, source)
	fmt.Printf("decode01 %#v\n", d)
	fmt.Printf("decode01 %#v\n", *d.Properties[0])
}

func TestDecoderDecode02(t *testing.T) {
	type (
		Data struct {
			Name string
		}
	)
	datas := make([]Data, 0)
	source := []interface{}{
		map[string]interface{}{
			"name": "a1",
		},
		map[string]interface{}{
			"name": "b1",
		},
	}
	err := NewDecoder().Decode(&datas, source)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("decode %#v\n", datas)
}

type (
	Data01Test03 struct {
		When []Data02Test03
	}
	Data02Test03 struct {
		Name string
	}
)

func (d *Data02Test03) UnmarshallStruct(v interface{}) error {

	return nil
}

func TestDecodeDecode03(t *testing.T) {
	_json := `{
		"when": [
			{
				"name1":"name1"
			},
			{
				"name2":"name2"
			}
		]
	}`
	_jsonRaw := map[string]interface{}{}
	err := json.Unmarshal([]byte(_json), &_jsonRaw)
	if err != nil {
		t.Fatal(err)
	}
	data := Data01Test03{}
	err = NewDecoder().Decode(&data, _jsonRaw)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("decode03 %#v\n", data)
}
