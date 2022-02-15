package csv

import (
	"bytes"
	"context"
	"encoding/csv"
	"os"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

func csv_read(ctx context.Context, rd *schema.MethodData, provider interface{}) error {
	fileName := rd.GetConfig("file_name").(string)
	columnsDefinitions := rd.GetConfig("column").([]interface{})
	delimiter := rd.GetConfig("delimiter").(string)
	additional := struct {
		from int
		to   int
		name string
	}{
		from: rd.GetConfig("additional/from_column").(int),
		to:   rd.GetConfig("additional/to_column").(int),
		name: rd.GetConfig("additional/name").(string),
	}
	type colDef struct {
		name   string
		offset int
	}
	columnsDef := make([]colDef, len(columnsDefinitions))
	for i, x := range columnsDefinitions {
		colDef := x.(map[string]interface{})
		columnsDef[i].name = colDef["name"].(string)
		columnsDef[i].offset = colDef["column"].(int)
	}
	by, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	reader := csv.NewReader(bytes.NewReader(by))
	reader.Comma = rune(delimiter[0])
	lines, err := reader.ReadAll()
	if err != nil {
		return err
	}
	result := []interface{}{}
	for _, line := range lines {
		size := len(line)
		item := map[string]interface{}{}
		for _, col := range columnsDef {
			if 0 < col.offset && col.offset <= size {
				item[col.name] = line[col.offset-1]
			}
		}
		if additional.to <= 0 && additional.from <= size && additional.from > 0 {
			item[additional.name] = line[additional.from-1:]
		} else if additional.from > 0 && additional.to > additional.from && additional.to <= size {
			item[additional.name] = line[additional.from-1 : additional.to-1]
		} else {
			item[additional.name] = []interface{}{}
		}
		result = append(result, item)
	}
	rd.SetResult("data", result)
	return nil
}
