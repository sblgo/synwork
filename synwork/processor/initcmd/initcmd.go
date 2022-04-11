package initcmd

import (
	"context"
	"flag"

	"sbl.systems/go/synwork/synwork/processor/cfg"
	"sbl.systems/go/synwork/synwork/processor/runtime"
)

type (
	cmd struct {
		config  *cfg.Config
		DirName string
	}
	cmdProvider struct {
	}
)

func (c *cmd) parseArgs(args []string) {
	fs := flag.NewFlagSet("initialize", flag.PanicOnError)
	fs.StringVar(&c.DirName, "f", ".", "directory containing configuration files")
	fs.Parse(args)
}

func init() {
	cfg.RegisterCmd("init", &cmdProvider{})
}

func (c *cmd) Init(config *cfg.Config) error {
	c.config = config
	return nil
}

func (c *cmd) Exec() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, err := runtime.NewRuntime(ctx, c.config)
	if err != nil {
		return err
	}
	defer r.Shutdown()
	if err = r.StartUp(runtime.RuntimeOptionsInit); err != nil {
		return err
	}
	return nil
}

func (*cmdProvider) Parse(args []string) (cfg.Cmd, error) {
	c := &cmd{}
	fs := flag.NewFlagSet("initialize", flag.PanicOnError)
	fs.StringVar(&c.DirName, "f", ".", "directory containing configuration files")
	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (*cmdProvider) Help() string {
	return `initialize a configuration. 
	synwork check, if all required processors local exists. It looks for new processor versions or download an initial version.

	
	`
}
