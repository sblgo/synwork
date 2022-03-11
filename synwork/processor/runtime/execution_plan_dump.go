package runtime

import "strings"

func (ep *ExecPlan) Dump(ec *ExecContext) error {
	ec.RuntimeNodes = map[string]*ExecRuntimeNode{}
	for _, epn := range ep.TargetMethods {
		ep.dumpItem(ec, epn, " - ")
		ec.RuntimeNodes[epn.Id] = &ExecRuntimeNode{}
	}

	return nil
}

func (ep *ExecPlan) dumpItem(ec *ExecContext, epn *ExecPlanNode, indent string) error {
	dublicateFlag := " "
	newIndent := indent
	var isDublicate bool
	if _, ok := ec.RuntimeNodes[epn.Id]; ok {
		dublicateFlag = "*"
		isDublicate = true
	} else {
		newIndent += ".. "
		ec.RuntimeNodes[epn.Id] = &ExecRuntimeNode{}
	}
	for _, c := range epn.Required {
		ep.dumpItem(ec, c, newIndent)
	}
	if isDublicate {
		return nil
	}
	if epn.Method != nil {
		ec.Log.Printf("%scall%s METHOD %s->%s (%s) ", indent, dublicateFlag, epn.Method.Instance, epn.Method.Name, strings.Replace(epn.Method.Id, "method/", "", 0))
	} else if epn.Processor != nil {
		ec.Log.Printf("%sinit%s PROCESSOR %s (%s) ", indent, dublicateFlag, epn.Processor.PluginName, strings.Replace(epn.Processor.Id, "processor/", "", 0))
	}
	return nil
}
