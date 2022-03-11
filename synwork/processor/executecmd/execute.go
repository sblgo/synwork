package executecmd

import (
	"context"
	"flag"
	"log"
	"os"

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
	log := log.New(os.Stderr, "[SYNWORK-EVAL]", log.Ltime|log.Lmicroseconds)
	c.parseArgs(args)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, err := runtime.NewRuntime(ctx, cf)
	if err != nil {
		log.Println(err)
		return
	}
	defer r.Shutdown()
	if err = r.StartUp(runtime.RuntimeOptionsExec); err != nil {
		log.Println(err)
		return
	}
	if err = r.Dump(ctx); err != nil {
		log.Println(err)
		return
	}
	if err = r.Exec(ctx); err != nil {
		log.Println(err)
		return
	}
}

func (c *cmd) parseArgs(args []string) {
	fs := flag.NewFlagSet("execute", flag.PanicOnError)
	fs.StringVar(&c.DirName, "f", ".", "directory containing configuration files")
	fs.Parse(args)
}
