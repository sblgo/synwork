package utils

import "testing"
import "fmt"

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
	fmt.Printf("decode %#v\n", d)
}
