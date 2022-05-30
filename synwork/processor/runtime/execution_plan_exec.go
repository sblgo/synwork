package runtime

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

type ExecContext struct {
	Context      context.Context
	RuntimeNodes map[string]*ExecRuntimeNode
	Log          log.Logger
	Error        error
	Parameters   map[string]string
}

type ExecRuntimeNode struct {
	PlanNode *ExecPlanNode
	Data     map[string]interface{}
	Result   map[string]interface{}
	Required []*ExecRuntimeNode
	once     sync.Once
}

func (ep *ExecPlan) Exec(ec *ExecContext) error {
	ec.RuntimeNodes = map[string]*ExecRuntimeNode{}
	startNodes := []*ExecRuntimeNode{}
	for _, epn := range ep.TargetMethods {
		if ern := ep.createRuntimeNodes(ec, epn); ern != nil {
			startNodes = append(startNodes, ern)
		}
	}

	for _, ern := range startNodes {
		if err := ep.executeRuntimeNode(ec, ern); err != nil {

		}
	}

	return nil
}

func (ep *ExecPlan) createRuntimeNodes(ec *ExecContext, epn *ExecPlanNode) *ExecRuntimeNode {
	if _, ok := ec.RuntimeNodes[epn.Id]; ok {
		return nil
	}
	ern := &ExecRuntimeNode{
		PlanNode: epn,
		Required: make([]*ExecRuntimeNode, 0),
	}
	ec.RuntimeNodes[ern.PlanNode.Id] = ern
	for _, epn2 := range epn.Required {
		if ern2 := ep.createRuntimeNodes(ec, epn2); ern2 != nil {
			ern.Required = append(ern.Required, ern2)
		}
	}
	return ern
}

func (ep *ExecPlan) executeRuntimeNode(ec *ExecContext, ern *ExecRuntimeNode) error {
	for _, ern2 := range ern.Required {
		if err := ep.executeRuntimeNode(ec, ern2); err != nil {
			return err
		}
	}
	return ern.executeNode(ec)
}

func (ern *ExecRuntimeNode) executeNode(ec *ExecContext) (err error) {
	if ern.PlanNode.Method != nil {
		ern.once.Do(func() { err = ern.executeMethod(ec) })
	} else if ern.PlanNode.Processor != nil {
		ern.once.Do(func() { err = ern.executeProcessor(ec) })
	} else if ern.PlanNode.Variable != nil {
		ern.once.Do(func() { err = ern.executeVariable(ec) })
	}
	return
}

func (ern *ExecRuntimeNode) executeMethod(ec *ExecContext) error {
	runtimeObj := ern.PlanNode.Method.Data
	objData, err := ern.createObjectData(ec, runtimeObj)
	if err != nil {
		ec.Log.Printf("%s [%s->%s] prepare error %s", ern.PlanNode.Id, ern.PlanNode.Method.Processor.PluginName, ern.PlanNode.Method.Name, err.Error())
		return err
	}
	result, err := ern.PlanNode.Method.Plugin.Call(ern.PlanNode.Method.Instance, ern.PlanNode.Method.Name, objData)
	ern.Result = result
	if err != nil {
		ec.Log.Printf("%s [%s->%s] call error %s", ern.PlanNode.Id, ern.PlanNode.Method.Plugin.name, ern.PlanNode.Method.Name, err.Error())
	} else {
		ec.Log.Printf("%s [%s->%s] call done", ern.PlanNode.Id, ern.PlanNode.Method.Plugin.name, ern.PlanNode.Method.Name)
	}
	return err

}

func (ern *ExecRuntimeNode) executeProcessor(ec *ExecContext) error {
	runtimeObj := ern.PlanNode.Processor.data
	objData, err := ern.createObjectData(ec, runtimeObj)
	if err != nil {
		ec.Log.Printf("%s [%s->init] prepare error %s", ern.PlanNode.Id, ern.PlanNode.Processor.PluginName, err.Error())
		return err
	}
	err = ern.PlanNode.Processor.plugin.Init(ern.PlanNode.Processor.Id, objData)
	if err != nil {
		ec.Log.Printf("%s [%s->init] error %s", ern.PlanNode.Id, ern.PlanNode.Processor.PluginName, err.Error())
	} else {
		ec.Log.Printf("%s [%s->init] done", ern.PlanNode.Id, ern.PlanNode.Processor.PluginName)
	}
	return err
}

func (ern *ExecRuntimeNode) executeVariable(ec *ExecContext) error {
	runtimeObj := ern.PlanNode.Variable.objectData
	objData, err := ern.createObjectData(ec, runtimeObj)
	if err != nil {
		ec.Log.Printf("%s [%s->get] prepare error %s", ern.PlanNode.Id, ern.PlanNode.Variable.Id, err.Error())
		return err
	}
	value, err := ern.PlanNode.Variable.Eval(ec, objData)
	if err != nil {
		ec.Log.Printf("%s [%s->get] error %s", ern.PlanNode.Id, ern.PlanNode.Variable.Id, err.Error())
		return err
	} else {
		ec.Log.Printf("%s [%s->get] done", ern.PlanNode.Id, ern.PlanNode.Variable.Id)
	}
	if value == nil {
		err = fmt.Errorf("missing value for variable")
		ec.Log.Printf("%s [%s->get] error %s", ern.PlanNode.Id, ern.PlanNode.Variable.Id, err.Error())
		return err
	}
	ern.Result = map[string]interface{}{
		"var": value,
	}
	return err
}

func (ern *ExecRuntimeNode) createObjectData(ec *ExecContext, runObj *RuntimeObject) (map[string]interface{}, error) {
	newValue := schema.MapCopy(runObj.Value)
	objData := newValue
	for _, ref := range runObj.ReferenceMap {
		switch ref.Type {
		case ReferenceTypeProcessor, ReferenceTypeMethod:
			id := ref.Type.Id(ref.Key)
			if rn, ok := ec.RuntimeNodes[id]; ok {
				if val, ok := schema.GetValueMap(rn.Result, ref.Path); ok {
					if newVal, err := MapSchemaAndInterface(&ref.Schema, val); err != nil {
						return nil, err
					} else {
						tgtPath := strings.Split(strings.Trim(ref.Location, "/"), "/")
						schema.SetValueMap(newValue, tgtPath, newVal)
					}
				}
			}
		case ReferenceTypeVariable:
			id := ref.Type.Id(ref.Key)
			if rn, ok := ec.RuntimeNodes[id]; ok {
				if newVal, ok := rn.Result["var"]; ok {
					tgtPath := strings.Split(strings.Trim(ref.Location, "/"), "/")
					schema.SetValueMap(newValue, tgtPath, newVal)
				}
			}
		case ReferenceTypeEnvironment:
		}

	}
	return objData, nil
}
