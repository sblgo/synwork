package initcmd

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, err := runtime.NewRuntime(ctx, cf)
	if err != nil {
		panic(err)
	}
	defer r.Shutdown()
	if err = r.StartUp(runtime.RuntimeOptionsInit); err != nil {
		panic(err)
	}
}

func (c *cmd) parseArgs(args []string) {
	fs := flag.NewFlagSet("initialize", flag.PanicOnError)
	fs.StringVar(&c.DirName, "f", ".", "directory containing configuration files")
	fs.Parse(args)
}
