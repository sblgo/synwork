package runtime

import "sbl.systems/go/synwork/synwork/processor/cfg"

type Command interface {
	Eval(c *cfg.Config, args []string)
}
