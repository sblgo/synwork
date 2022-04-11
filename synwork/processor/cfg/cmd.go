package cfg

import (
	"fmt"
	"os"
	"strings"
)

type (
	ProgParams struct {
	}

	Cmd interface {
		Init(*Config) error
		Exec() error
	}

	CmdProvider interface {
		Help() string
		Parse(args []string) (Cmd, error)
	}
)

var (
	cmds = map[string]CmdProvider{
		"help": &helpCmdProvider{},
	}
)

func RegisterCmd(cmdName string, provider CmdProvider) {
	cmds[cmdName] = provider
}

func ParseArgs() (Cmd, error) {
	args := os.Args
	for i := 0; i < len(args); i++ {
		if provider, ok := cmds[args[i]]; ok {
			if progParams, err := parseProgParams(args[:i]); err != nil {
				return nil, err
			} else if cmd, err := provider.Parse(args[i+1:]); err != nil {
				return nil, err
			} else if err = cmd.Init(progParams); err != nil {
				return nil, err
			} else {
				return cmd, nil
			}

		}
	}
	cmdNames := []string{}
	for k := range cmds {
		cmdNames = append(cmdNames, k)
	}

	return nil, fmt.Errorf("use one of the commands [%s]", strings.Join(cmdNames, ", "))
}

func parseProgParams(args []string) (*Config, error) {
	return parseConfig(args)
}

type helpCmdProvider struct {
}

func (hc *helpCmdProvider) Help() string {
	return `
	use: synwork help <cmd>
	`
}

func (hc *helpCmdProvider) Parse(args []string) (Cmd, error) {
	if len(args) > 0 {
		return helpCmd(args[0]), nil
	}
	return nil, fmt.Errorf("missing command name\n%s", hc.Help())
}

type helpCmd string

func (hc helpCmd) Init(*Config) error {
	return nil
}
func (hc helpCmd) Exec() error {
	if prov, ok := cmds[string(hc)]; ok {
		fmt.Printf("synwork %s\n", hc)
		fmt.Println(prov.Help())
	} else {
		return fmt.Errorf("unknown command %s", hc)
	}
	return nil
}
