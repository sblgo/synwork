package randlist

import (
	"context"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

func random_list(ctx context.Context, rd *schema.MethodData, provider interface{}) error {
	r := provider.(*randomList)
	minId := rd.GetConfig("min_id").(int)
	maxId := rd.GetConfig("max_id").(int)

	list := []interface{}{}
	for val := minId; val <= maxId; val++ {
		list = append(list, map[string]interface{}{
			"id":    val,
			"value": int(r.random.Int31()),
		})
	}
	rd.SetResult("result", list)
	return nil
}
