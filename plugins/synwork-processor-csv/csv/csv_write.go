package csv

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

type csvCol struct {
	Name   string
	Path   []string
	Format string
}

func csv_write(ctx context.Context, rd *schema.MethodData, provider interface{}) error {
	fileName := rd.GetConfig("file_name").(string)
	columnsDefinitions := rd.GetConfig("column").([]interface{})
	colField := func(idx int, field string) string {
		cd := columnsDefinitions[idx].(map[string]interface{})
		if v, ok := cd[field]; ok {
			return v.(string)
		} else {
			return ""
		}
	}
	columnsFact := make([]csvCol, len(columnsDefinitions))
	for i := range columnsDefinitions {
		columnsFact[i].Name = colField(i, "name")
		columnsFact[i].Format = colField(i, "format")
		columnsFact[i].Path = strings.Split(strings.Trim(colField(i, "path"), "/"), "/")
	}
	b := new(bytes.Buffer)
	csvOut := csv.NewWriter(b)
	data := rd.GetConfig("data").([]interface{})
	for _, d := range data {
		row := make([]string, len(columnsFact))
		for ci, cf := range columnsFact {
			if v, ok := schema.GetValueMap(d, cf.Path); ok {
				row[ci] = fmt.Sprintf(cf.Format, v)
			}
		}
		csvOut.Write(row)
	}
	csvOut.Flush()
	os.WriteFile(fileName, b.Bytes(), 0644)
	return nil
}
