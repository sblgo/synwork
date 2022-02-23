package runtime

import (
	"fmt"
	"net/rpc"
	"time"

	"sbl.systems/go/synwork/plugin-sdk/comstrs"
	"sbl.systems/go/synwork/plugin-sdk/schema"
)

type Plugin struct {
	port      int
	name      string
	source    string
	client    *rpc.Client
	schema    map[string]*schema.Schema
	methodMap map[string]*schema.Method
	Provider  schema.Processor
}

func NewPlugin(port int, name string, source string) (*Plugin, error) {
	serverAddress := "127.0.0.1"
	for cnt := 5; true; cnt-- {

		client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", serverAddress, port))
		if err != nil {
			if cnt == 0 {
				return nil, err
			}
		} else {
			return &Plugin{
				port:   port,
				name:   name,
				source: source,
				client: client,
			}, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil, fmt.Errorf("")
}

func (p *Plugin) Shutdown() error {
	x := 1
	var reply int
	err := p.client.Call("Plugin.Shutdown", &x, &reply)

	return err
}

func (p *Plugin) Schema() error {
	dataIn := &comstrs.PluginSchemaIn{}
	dataOut := &comstrs.PluginSchemaOut{}
	err := p.client.Call("Plugin.Schema", dataIn, dataOut)
	p.schema = dataOut.Provider.Schema
	p.methodMap = dataOut.Provider.MethodMap
	p.Provider = dataOut.Provider
	return err
}

func (p *Plugin) Init(instName string, data map[string]interface{}) error {
	dataIn := &comstrs.PluginProcessorIn{
		InstanceName: instName,
		Data:         data,
	}
	dataOut := &comstrs.PluginProcessorOut{}
	err := p.client.Call("Plugin.Init", dataIn, dataOut)

	return err
}

func (p *Plugin) Call(instName string, methName string, data map[string]interface{}) (map[string]interface{}, error) {
	dataIn := &comstrs.PluginMethodIn{
		InstanceName: instName,
		MethodName:   methName,
		Data:         data,
	}
	dataOut := &comstrs.PluginMethodOut{
		Result: map[string]interface{}{},
	}
	methDetail, ok := p.methodMap[methName]
	if !ok {
		return nil, fmt.Errorf("plugin %s has no method %s", p.name, methName)
	}
	for k, _ := range methDetail.Result {
		dataOut.Result[k] = map[string]interface{}{}
	}
	err := p.client.Call("Plugin.Call", dataIn, dataOut)

	return dataOut.Result, err
}
