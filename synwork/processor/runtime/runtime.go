package runtime

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-version"
	"sbl.systems/go/synwork/synwork/ast"
	"sbl.systems/go/synwork/synwork/parser"
	"sbl.systems/go/synwork/synwork/processor/cfg"
)

type Runtime struct {
	dirName          string
	portFrom, portTo int
	Blocks           []*ast.BlockNode
	pluginSources    []*PluginSource
	plugins          map[string]*Plugin
	processors       map[string]*Processor
	variables        map[string]*Variable
	methods          map[string]*Method
	execPlan         *ExecPlan
	config           *cfg.Config
	context          context.Context
}

type RuntimeOption func(rt *Runtime) error

var (
	useLocalPlugins = func(ps *PluginSource) error {
		if err := ps.verifyPlugin(); err != nil {
			return err
		}
		if err := ps.selectPluginProgram(); err != nil {
			return err
		}
		return nil
	}

	updateLocalPlugins = func(ps *PluginSource) error {
		if err := ps.verifyAndLoadPlugin(); err != nil {
			return err
		}
		if err := ps.selectPluginProgram(); err != nil {
			return err
		}
		return nil
	}

	RuntimeOptionsInit = []RuntimeOption{
		parse,
		initPluginSource(updateLocalPlugins),
	}
	RuntimeOptionsHelp = []RuntimeOption{
		parse,
		initPluginSource(useLocalPlugins),
		initVariable,
		initPlugins,
	}

	RuntimeOptionsExec = []RuntimeOption{
		parse,
		initPluginSource(useLocalPlugins),
		initVariable,
		initPlugins,
		initProcessors,
		initMethods,
		buildExecPlan,
	}
)

func NewRuntime(ctx context.Context, c *cfg.Config) (*Runtime, error) {
	r := &Runtime{
		dirName:       c.WorkDir,
		pluginSources: []*PluginSource{},
		plugins:       map[string]*Plugin{},
		processors:    map[string]*Processor{},
		variables:     map[string]*Variable{},
		methods:       map[string]*Method{},
		portFrom:      c.PortFrom,
		portTo:        c.PortTo,
		config:        c,
		context:       ctx,
	}
	return r, nil
}

func (r *Runtime) StartUp(options []RuntimeOption) error {
	for _, f := range options {
		if err := f(r); err != nil {
			return err
		}
	}
	return nil
}

func parse(r *Runtime) error {
	p, err := parser.NewParser(r.dirName)
	if err != nil {
		return err
	}
	err = p.Parse()
	if err != nil {
		return err
	}
	r.Blocks = p.Blocks
	return nil
}

func initVariable(r *Runtime) error {
	for _, b := range r.Blocks {
		if b.Type == "variable" {
			var name string
			if len(b.Identifiers) != 1 {
				return fmt.Errorf("name of variable isn't exact defined %s", b.Pos())
			}
			if obj, err := MapSchemaAndNode(VariableSchema, *b.Content); err != nil {
				return err
			} else {
				if _, ok := r.variables[name]; ok {
					return fmt.Errorf("variable %s always defined before %s", name, b.Pos())
				}
				variable := NewVariable(name, obj)
				r.variables[name] = variable
			}
		}
	}
	return nil
}

func initPluginSource(factory func(ps *PluginSource) error) func(r *Runtime) error {
	return func(r *Runtime) error {
		for _, b := range r.Blocks {
			if b.Type == "synwork" {
				if obj, err := MapSchemaAndNode(PluginSchema, *b.Content); err != nil {
					return err
				} else {
					if list, ok := obj.Value["required_processor"]; ok {
						for _, plgl := range list.([]interface{}) {
							plg := plgl.(map[string]interface{})
							source, versionSelector := plg["source"].(string), plg["version"].(string)
							if versionSelector == "" {
								versionSelector = ">=0.0.1"
							}
							plgSource, err := NewPluginSourceFromSource(r.config, source)
							if err != nil {
								return err
							}
							if cnstr, err := version.NewConstraint(versionSelector); err != nil {
								return err
							} else {
								plgSource.VersionConstraint = cnstr
							}
							if err := factory(plgSource); err != nil {
								return err
							} else {
								r.pluginSources = append(r.pluginSources, plgSource)
							}
						}
					}
				}
			}
		}
		return nil
	}
}

func initPlugins(r *Runtime) error {
	for _, b := range r.pluginSources {
		plgRuntime := &PluginRuntime{
			PluginKey:     b.PluginKey,
			Config:        r.config,
			pluginProgram: b.PluginProgram,
		}
		if plgRun, err := plgRuntime.Start(r.context, r.portFrom, r.portTo); err != nil {
			r.Shutdown()
			return err
		} else {
			r.portFrom = plgRun.port + 1
			if err = r.initPlugin(plgRun); err != nil {
				r.Shutdown()
				return err
			}
		}

	}

	return nil
}

func initProcessors(r *Runtime) error {
	for _, b := range r.Blocks {
		if b.Type == "processor" {
			if len(b.Identifiers) != 2 {
				return fmt.Errorf("invalid definition for processor missing identifiers at %s", b.Pos())
			}
			processor := &Processor{
				Id:         b.Identifiers[1],
				PluginName: b.Identifiers[0],
			}
			if plugin, ok := r.plugins[processor.PluginName]; ok {
				processor.plugin = plugin
			} else {
				return fmt.Errorf("plugin %s not defined see %s", processor.PluginName, b.Pos())
			}
			processor.schema = processor.plugin.schema
			if obj, err := MapSchemaAndNode(processor.schema, *b.Content); err != nil {
				return err
			} else {
				processor.data = obj
			}
			r.processors[processor.Id] = processor
		}
	}
	return nil
}

func initMethods(r *Runtime) error {
	for _, b := range r.Blocks {
		if b.Type == "method" {
			if len(b.Identifiers) != 3 {
				return fmt.Errorf("invalid definition for processor missing identifiers at %s", b.Pos())
			}
			methodName, instanceName, methodId := b.Identifiers[0], b.Identifiers[1], b.Identifiers[2]
			if instanceObj, ok := r.processors[instanceName]; !ok {
				return fmt.Errorf("processor %s used at %s doesn't exist", instanceName, b.Pos())
			} else {
				if methodDef, ok := instanceObj.plugin.methodMap[methodName]; !ok {
					return fmt.Errorf("method %s for processor %s of type %s isn't defined. (%s)", methodName, instanceName, instanceObj.PluginName, b.Pos())
				} else {
					runtimeObj, err := MapSchemaAndNode(methodDef.Schema, *b.Content)
					if err != nil {
						return err
					}
					method := &Method{
						Id:        methodId,
						Name:      methodName,
						Instance:  instanceName,
						Data:      runtimeObj,
						Processor: instanceObj,
						Plugin:    instanceObj.plugin,
						Schema:    *methodDef,
					}
					r.methods[methodId] = method
				}
			}

		}
	}
	return nil
}

func (r *Runtime) Shutdown() {
	for _, plg := range r.plugins {
		plg.Shutdown()
	}
}

func (r *Runtime) initPlugin(plgRun *Plugin) error {
	err := plgRun.Schema()
	if err != nil {
		return err
	}
	err = r.addPlugin(plgRun)
	if err != nil {
		return err
	}

	return err
}

func (r *Runtime) addPlugin(plgRun *Plugin) error {
	if plg2, ok := r.plugins[plgRun.name]; ok {
		return fmt.Errorf("plugin %s always defined %s(%s)", plgRun.name, plg2.source, plg2.name)
	} else {
		r.plugins[plgRun.name] = plgRun
	}

	return nil
}

func buildExecPlan(r *Runtime) error {
	ep := &ExecPlan{
		Processor:     map[string]*ExecPlanNode{},
		TargetMethods: map[string]*ExecPlanNode{},
	}
	for _, p := range r.processors {
		ep.AddProcessor(p)

	}
	for _, m := range r.methods {
		ep.AddMethod(m)
	}
	err := ep.Build()
	if err != nil {
		return err
	}
	r.execPlan = ep

	return nil
}

func GetPlugins(set func(map[string]*Plugin)) func(r *Runtime) error {

	return func(r *Runtime) error {
		set(r.plugins)
		return nil
	}
}

func (r *Runtime) Dump(ctx context.Context) error {
	ctx2 := &ExecContext{
		Context:      ctx,
		RuntimeNodes: map[string]*ExecRuntimeNode{},
		Log:          *log.New(os.Stderr, "[SYNWORK-PLAN]", 0),
	}
	err := r.execPlan.Dump(ctx2)
	return err
}

func (r *Runtime) Exec(ctx context.Context) error {

	ctx2 := &ExecContext{
		Context:      ctx,
		RuntimeNodes: map[string]*ExecRuntimeNode{},
		Log:          *log.New(os.Stderr, "[SYNWORK-EXEC]", log.Ltime),
	}
	err := r.execPlan.Exec(ctx2)
	return err
}
