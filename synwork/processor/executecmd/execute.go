package executecmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"sbl.systems/go/synwork/synwork/processor/cfg"
	"sbl.systems/go/synwork/synwork/processor/runtime"
)

type (
	Parameter struct {
		Name, Value string
	}
	Parameters []Parameter

	cmd struct {
		DirName   string
		config    *cfg.Config
		Paramters Parameters
	}

	cmdProvider struct {
	}
)

func (p *Parameters) Set(val string) error {
	parts := strings.SplitN(val, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid parameter value '%s'", val)
	}
	*p = append(*p, Parameter{parts[0], parts[1]})
	return nil
}

func (p *Parameters) String() string {
	j := make([]string, len(*p))
	for i, p := range *p {
		j[i] = fmt.Sprintf("%s=%s", p.Name, p.Value)
	}
	return strings.Join(j, ", ")
}

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
	params := map[string]string{}
	for _, p := range c.Paramters {
		params[p.Name] = p.Value
	}
	if err = r.Exec(ctx, params); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (c *cmd) parseArgs(args []string) error {
	fs := flag.NewFlagSet("execute", flag.PanicOnError)
	fs.StringVar(&c.DirName, "f", ".", "directory containing configuration files")
	fs.Var(&c.Paramters, "p", "define one or multiple parameters name=value")
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
	c := &cmd{
		Paramters: make(Parameters, 0),
	}
	if err := c.parseArgs(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (*cmdProvider) Help() string {
	return `	execute a configuration

	parameters:
	-f directory containing configuration files (default .)
	-p define one or more parameters each entry with name=value
	`

}
