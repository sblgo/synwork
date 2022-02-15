package randlist

import (
	"context"
	"math/rand"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

type randomList struct {
	random rand.Rand
}

func random_init(ctx context.Context, data *schema.ObjectData, obj interface{}) (interface{}, error) {
	var r *randomList
	if obj == nil {
		seed := data.Get("seed").(int)
		r = &randomList{
			random: *rand.New(rand.NewSource(int64(seed))),
		}
	} else {
		r = obj.(*randomList)
	}

	return r, nil
}
