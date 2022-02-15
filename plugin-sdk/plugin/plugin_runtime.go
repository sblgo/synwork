package plugin

import (
	"context"
	"fmt"

	"sbl.systems/go/synwork/plugin-sdk/comstrs"
	"sbl.systems/go/synwork/plugin-sdk/schema"
)

type Plugin struct {
	shutdown        chan struct{}
	provider        func() schema.Processor
	providerDetails schema.Processor
	instances       map[string]*runtimeObject
}

type runtimeObject struct {
	name     string
	instance interface{}
}

func (p *Plugin) Shutdown(level *int, status *int) error {
	defer func() {
		close(p.shutdown)
	}()
	return fmt.Errorf("shutdown %d", *level)
}

func (p *Plugin) Env(env *map[string]interface{}, res *map[string]interface{}) error {
	fmt.Printf("env %#v", *env)
	(*res) = (*env)
	return nil
}

func (p *Plugin) Schema(dataIn *comstrs.PluginSchemaIn, dataOut *comstrs.PluginSchemaOut) error {
	provider := p.provider()
	p.providerDetails = provider
	dataOut.Provider = provider
	return nil
}

func (p *Plugin) Init(dataIn *comstrs.PluginProcessorIn, dataOut *comstrs.PluginProcessorOut) error {
	if p.instances == nil {
		p.instances = map[string]*runtimeObject{}
	}
	ctx := context.Background()
	instance, ok := p.instances[dataIn.InstanceName]
	if !ok {
		instance = &runtimeObject{}
	}
	objData := schema.NewObjectData(p.providerDetails.Schema, dataIn.Data)
	if newInst, err := p.providerDetails.InitFunc(ctx, objData, instance.instance); err != nil {
		return err
	} else {
		p.instances[dataIn.InstanceName] = &runtimeObject{
			instance: newInst,
			name:     dataIn.InstanceName,
		}
		return nil
	}
}

func (p *Plugin) Call(dataIn *comstrs.PluginMethodIn, dataOut *comstrs.PluginMethodOut) error {
	if p.instances == nil {
		return fmt.Errorf("missing initialisation")
	}
	instance, ok := p.instances[dataIn.InstanceName]
	if !ok {
		return fmt.Errorf("unknown instance %s", dataIn.InstanceName)
	}
	method, ok := p.providerDetails.MethodMap[dataIn.MethodName]
	if !ok {
		return fmt.Errorf("plugin %s has no method %s", dataIn.InstanceName, dataIn.MethodName)
	}
	ctx := context.Background()
	results := map[string]interface{}{}
	for k := range method.Result {
		results[k] = map[string]interface{}{}
	}
	data := schema.NewMethodData(*schema.NewObjectData(method.Schema, dataIn.Data), *schema.NewObjectData(method.Result, results))
	err := method.ExecFunc(ctx, data, instance.instance)
	if err != nil {
		return err
	}

	dataOut.Result = results

	return nil
}
