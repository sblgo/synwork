package executecmd

import (
	"context"
	"flag"
	"log"
	"os"

	"sbl.systems/go/synwork/synwork/processor/cfg"
	"sbl.systems/go/synwork/synwork/processor/runtime"
)

type (
	cmd struct {
		DirName string
		config  *cfg.Config
	}

	cmdProvider struct {
	}
)

func (c *cmd) Exec() error {
	log := log.New(os.Stderr, "[SYNWORK-EVAL]", log.Ltime|log.Lmicroseconds)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, err := runtime.NewRuntime(ctx, c.config)
	if err != nil {
		log.Println(err)
		return err
	}
	defer r.Shutdown()
	if err = r.StartUp(runtime.RuntimeOptionsExec); err != nil {
		log.Println(err)
		return err
	}
	if err = r.Dump(ctx); err != nil {
		log.Println(err)
		return err
	}
	if err = r.Exec(ctx); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (c *cmd) parseArgs(args []string) error {
	fs := flag.NewFlagSet("execute", flag.PanicOnError)
	fs.StringVar(&c.DirName, "f", ".", "directory containing configuration files")
	return fs.Parse(args)
}

func init() {
	cfg.RegisterCmd("exec", &cmdProvider{})
}

func (c *cmd) Init(config *cfg.Config) error {
	c.config = config
	return nil
}

func (*cmdProvider) Parse(args []string) (cfg.Cmd, error) {
	c := &cmd{}
	if err := c.parseArgs(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (*cmdProvider) Help() string {
	return `	execute a configuration

	parameters:
	-f directory containing configuration files (default .)
	`

}
