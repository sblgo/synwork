package main

import (
	"encoding/gob"
	"fmt"
	"os"

	"sbl.systems/go/synwork/synwork/processor/cfg"
	_ "sbl.systems/go/synwork/synwork/processor/executecmd"
	_ "sbl.systems/go/synwork/synwork/processor/helpcmd"
	_ "sbl.systems/go/synwork/synwork/processor/initcmd"
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
	if cmd, err := cfg.ParseArgs(); err != nil {
		fmt.Println(err)
		os.Exit(2)
	} else if err = cmd.Exec(); err != nil {
		fmt.Println(err)
		os.Exit(2)
	} else {
		return
	}
}

func gobRegister() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}
