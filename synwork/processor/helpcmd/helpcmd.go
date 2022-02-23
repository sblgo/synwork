package helpcmd

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"

	"sbl.systems/go/synwork/plugin-sdk/schema"
	"sbl.systems/go/synwork/synwork/processor/cfg"
	"sbl.systems/go/synwork/synwork/processor/runtime"
)

type cmd struct {
	Plugin string
	Method string
	Config bool
	Result bool
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
	var pluginsMap map[string]*runtime.Plugin
	set := func(m map[string]*runtime.Plugin) {
		pluginsMap = m
	}
	if err = r.StartUp(append(runtime.RuntimeOptionsHelp, runtime.GetPlugins(set))); err != nil {
		panic(err)
	}
	if c.Plugin == "" {
		for k, pl := range pluginsMap {
			fmt.Println(k)
			fmt.Println("Description:")
			fmt.Println(pl.Provider.Description)
			fmt.Println()
		}
	} else {
		if item, ok := pluginsMap[c.Plugin]; ok {
			if c.Method == "" && !c.Config {
				fmt.Println(c.Plugin)
				fmt.Println("Methods:")
				for k, m := range item.Provider.MethodMap {
					fmt.Println(k)
					fmt.Println("Descritpion:")
					fmt.Println(m.Description)
					fmt.Println()
				}
			} else if c.Method == "" && c.Config {
				fmt.Println(c.Plugin)
				fmt.Println("Configuration")
				fmt.Println(formatSchema(item.Provider.Schema))
			} else if c.Method != "" {
				if method, ok := item.Provider.MethodMap[c.Method]; ok {
					if c.Config {
						fmt.Println(c.Plugin, "-", c.Method)
						fmt.Println("Configuration")
						fmt.Println(formatSchema(method.Schema))
					}
					if c.Result {
						fmt.Println(c.Plugin, "-", c.Method)
						fmt.Println("Result")
						fmt.Println(formatSchema(method.Result))
					}
					if !c.Result && !c.Config {
						fmt.Println(c.Plugin, "-", c.Method)
						fmt.Println("Description")
						fmt.Println(method.Description)
					}
				}
			}
		} else {
			fmt.Println(c.Plugin, "not found")
		}
	}
}

func (c *cmd) parseArgs(args []string) {
	fs := flag.NewFlagSet("initialize", flag.PanicOnError)
	fs.StringVar(&c.Method, "m", "", "method of the plugin")
	fs.StringVar(&c.Plugin, "p", "", "processor")
	fs.BoolVar(&c.Config, "c", false, "display the details of configuration")
	fs.BoolVar(&c.Result, "r", false, "display the result of a method")
	fs.Parse(args)
}

func formatSchema(s map[string]*schema.Schema) string {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.SetIndent("...", "   ")
	err := encoder.Encode(schemaMapToMap(s))
	if err != nil {
		fmt.Println(err)
	}
	return b.String()
}

func schemaMapToMap(m map[string]*schema.Schema) map[string]map[string]interface{} {
	r := map[string]map[string]interface{}{}
	for k, v := range m {
		r[k] = schemaToMap(v)
	}
	return r
}
func schemaToMap(s *schema.Schema) map[string]interface{} {

	r := map[string]interface{}{
		"01-Type":         s.Type.String(),
		"02-Optional":     s.Optional,
		"03-Required":     s.Required,
		"04-DefaultValue": s.DefaultValue,
		"05-ElemType":     s.ElemType.String(),
	}
	r["06-Elem"] = schemaMapToMap(s.Elem)
	return r
}
