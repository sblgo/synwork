package runtime

import (
	"context"
	"strings"
	"sync"

	"sbl.systems/go/synwork/plugin-sdk/schema"
)

type ExecContext struct {
	Context      context.Context
	RuntimeNodes map[string]*ExecRuntimeNode
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
		ep.executeRuntimeNode(ec, ern)
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

func (ep *ExecPlan) executeRuntimeNode(ec *ExecContext, ern *ExecRuntimeNode) {
	for _, ern2 := range ern.Required {
		ep.executeRuntimeNode(ec, ern2)
	}
	ern.executeNode(ec)
}

func (ern *ExecRuntimeNode) executeNode(ec *ExecContext) {
	if ern.PlanNode.Method != nil {
		ern.once.Do(func() { ern.executeMethod(ec) })
	} else if ern.PlanNode.Processor != nil {
		ern.once.Do(func() { ern.executeProcessor(ec) })
	}
}

func (ern *ExecRuntimeNode) executeMethod(ec *ExecContext) error {
	runtimeObj := ern.PlanNode.Method.Data
	objData, err := ern.createObjectData(ec, runtimeObj)
	if err != nil {
		return err
	}
	result, err := ern.PlanNode.Method.Plugin.Call(ern.PlanNode.Method.Instance, ern.PlanNode.Method.Name, objData)
	ern.Result = result
	return err

}

func (ern *ExecRuntimeNode) executeProcessor(ec *ExecContext) error {
	runtimeObj := ern.PlanNode.Processor.data
	objData, err := ern.createObjectData(ec, runtimeObj)
	if err != nil {
		return err
	}
	err = ern.PlanNode.Processor.plugin.Init(ern.PlanNode.Processor.Id, objData)
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
					tgtPath := strings.Split(strings.Trim(ref.Location, "/"), "/")
					schema.SetValueMap(newValue, tgtPath, val)
				}
			}
		case ReferenceTypeVariable:
		case ReferenceTypeEnvironment:
		}

	}
	return objData, nil
}
