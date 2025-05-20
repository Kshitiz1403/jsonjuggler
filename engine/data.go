package engine

// WorkflowData holds the data state during workflow execution
type WorkflowData struct {
	// Initial holds the initial data being passed to the workflow
	Initial interface{} `json:"initial"`
	// Current holds the current data being processed
	Current interface{} `json:"current"`
	// States holds the output data from each state
	States map[string]interface{} `json:"states"`
	// Globals holds workflow-level variables that persist throughout execution
	Globals map[string]interface{} `json:"globals"`
}

// NewWorkflowData creates a new WorkflowData instance
func NewWorkflowData(input interface{}, globals map[string]interface{}) *WorkflowData {
	if globals == nil {
		globals = make(map[string]interface{})
	}
	return &WorkflowData{
		Initial: input,
		Current: input,
		States:  make(map[string]interface{}),
		Globals: globals,
	}
}

// ToMap converts WorkflowData to a map for JQ evaluation
func (w *WorkflowData) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"initial": w.Initial,
		"current": w.Current,
		"states":  w.States,
		"globals": w.Globals,
	}
}
