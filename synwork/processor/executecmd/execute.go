package executecmd

import (
	"context"
	"flag"

	"sbl.systems/go/synwork/synwork/processor/cfg"
	"sbl.systems/go/synwork/synwork/processor/runtime"
)

type cmd struct {
	DirName string
}

func NewCmd() runtime.Command {
	return &cmd{}
}

func (c *cmd) Eval(cf *cfg.Config, args []string) {
	c.parseArgs(args)
	r, err := runtime.NewRuntime(cf)
	if err != nil {
		panic(err)
	}
	defer r.Shutdown()
	ctx := context.Background()
	r.Exec(ctx)
}

func (c *cmd) parseArgs(args []string) {
	fs := flag.NewFlagSet("execute", flag.PanicOnError)
	fs.StringVar(&c.DirName, "f", ".", "directory containing configuration files")
	fs.Parse(args)
}
