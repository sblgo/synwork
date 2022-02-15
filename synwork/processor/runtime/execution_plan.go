package runtime

import "fmt"

type ExecPlan struct {
	Processor     map[string]*ExecPlanNode
	TargetMethods map[string]*ExecPlanNode
}

type ExecPlanNode struct {
	Id        string
	Processor *Processor
	Method    *Method
	Required  []*ExecPlanNode
	Dependend []*ExecPlanNode
}

func (ep *ExecPlan) AddProcessor(p *Processor) {
	epn := &ExecPlanNode{
		Id:        ReferenceTypeProcessor.Id(p.Id),
		Processor: p,
		Dependend: make([]*ExecPlanNode, 0),
	}
	ep.Processor[epn.Id] = epn

}

func (ep *ExecPlan) AddMethod(m *Method) {
	epn := &ExecPlanNode{
		Id:        ReferenceTypeMethod.Id(m.Id),
		Dependend: make([]*ExecPlanNode, 0),
		Required:  make([]*ExecPlanNode, 0),
		Method:    m,
	}
	ep.TargetMethods[epn.Id] = epn
}

func (ep *ExecPlan) Build() error {
	keys := []string{}
	for _, epn := range ep.TargetMethods {
		if epn2, ok := ep.Processor[ReferenceTypeProcessor.Id(epn.Method.Instance)]; ok {
			epn.AddRequired(epn2)
			epn2.AddDependend(epn)
		} else {
			return fmt.Errorf("processor instance %s not found", epn.Method.Instance)
		}

		for _, rfs := range epn.Method.Data.ReferenceMap {
			switch rfs.Type {
			case ReferenceTypeMethod:
				id := ReferenceTypeMethod.Id(rfs.Key)
				if epn2, ok := ep.TargetMethods[id]; ok {
					epn.AddRequired(epn2)
					epn2.AddDependend(epn)
					keys = append(keys, string(id))
				} else {
					return fmt.Errorf("method instance %s not found", id)
				}
			case ReferenceTypeProcessor:
				return fmt.Errorf("processor can't be part in reference")
			}

		}
	}
	for _, key := range keys {
		delete(ep.TargetMethods, key)
	}
	if len(ep.TargetMethods) == 0 {
		return fmt.Errorf("no method call left")
	}
	return nil
}

func (epn *ExecPlanNode) AddRequired(epn2 *ExecPlanNode) {
	epn.Required = append(epn.Required, epn2)
}

func (epn *ExecPlanNode) AddDependend(epn2 *ExecPlanNode) {
	epn.Dependend = append(epn.Dependend, epn2)
}
