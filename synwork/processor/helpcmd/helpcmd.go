package helpcmd

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"sbl.systems/go/synwork/plugin-sdk/schema"
	"sbl.systems/go/synwork/synwork/processor/cfg"
	"sbl.systems/go/synwork/synwork/processor/runtime"
)

type (
	cmd struct {
		Plugin string
		Method string
		Config bool
		Result bool
		All    bool
		config *cfg.Config
	}
	cmdProvider struct {
	}
)

func (c *cmd) Exec() error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r, err := runtime.NewRuntime(ctx, c.config)
	if err != nil {
		return err
	}
	defer r.Shutdown()
	var pluginsMap map[string]*runtime.Plugin
	set := func(m map[string]*runtime.Plugin) {
		pluginsMap = m
	}
	if err = r.StartUp(append(runtime.RuntimeOptionsHelp, runtime.GetPlugins(set))); err != nil {
		return err
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
			if c.All {
				fmt.Printf("processor: %s\n", c.Plugin)
				fmt.Println("Description:")
				fmt.Println(item.Provider.Description)
				for k, v := range item.Provider.MethodMap {
					c.printConfiguration(k, v, true, true)
				}
			} else if c.Method == "" && !c.Config {
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
					c.printConfiguration(c.Method, method, c.Config, c.Result)
				}
			}
		} else {
			fmt.Println(c.Plugin, "not found")
		}
	}
	return nil
}

func (c *cmd) printConfiguration(name string, method *schema.Method, config bool, result bool) {
	if c.Config {
		fmt.Printf("%s->%s\n", c.Plugin, name)
		fmt.Println("Configuration")
		fmt.Println(formatSchema(method.Schema))
	}
	if config {
		fmt.Printf("%s->%s\n", c.Plugin, name)
		fmt.Println("Description")
		fmt.Println(method.Description)
	}
	if result {
		fmt.Printf("%s->%s\n", c.Plugin, name)
		fmt.Println("Result")
		fmt.Println(formatSchema(method.Result))
	}

}

func (c *cmd) parseArgs(args []string) error {
	fs := flag.NewFlagSet("initialize", flag.PanicOnError)
	fs.StringVar(&c.Method, "m", "", "method of the plugin")
	fs.StringVar(&c.Plugin, "p", "", "processor")
	fs.BoolVar(&c.Config, "c", false, "display the details of configuration")
	fs.BoolVar(&c.Result, "r", false, "display the result of a method")
	fs.BoolVar(&c.All, "a", false, "display all details for all methods and configuration")
	return fs.Parse(args)
}

func formatSchema(s map[string]*schema.Schema) string {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.SetIndent("...", "   ")
	//	err := encoder.Encode(schemaMapToMap(s))
	err := encoder.Encode(s)
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

func init() {
	cfg.RegisterCmd("details", &cmdProvider{})
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
	return `	shows details to a processor
	
	parameters:
	m	method of the processor (string)
	p	processor name (string)
	c	display the details of configuration (bool)
	r	display the result of a method (bool)
	a	display all details for all methods and configuration (bool)
	`
}
