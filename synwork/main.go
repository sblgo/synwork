package main

import (
	"encoding/gob"
	"fmt"

	"sbl.systems/go/synwork/synwork/processor/cfg"
	"sbl.systems/go/synwork/synwork/processor/executecmd"
	"sbl.systems/go/synwork/synwork/processor/helpcmd"
	"sbl.systems/go/synwork/synwork/processor/initcmd"
	_ "sbl.systems/go/synwork/synwork/providers/awsprovider"
)

var cmds = []string{
	"exec",
	"execute",
	"init",
	"help",
}

func main() {
	gobRegister()
	args, cfg, err := cfg.ParseArgs(cmds)
	if err != nil {
		panic(err)
	}
	switch args[0] {
	case "exec", "execute":
		cmd := executecmd.NewCmd()
		cmd.Eval(cfg, args[1:])
	case "init":
		cmd := initcmd.NewCmd()
		cmd.Eval(cfg, args[1:])
	case "help":
		cmd := helpcmd.NewCmd()
		cmd.Eval(cfg, args[1:])
	default:
		panic(fmt.Errorf("unknown cmd %s", args[0]))
	}
}

func gobRegister() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}
