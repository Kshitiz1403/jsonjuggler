package engine

import sw "github.com/serverlessworkflow/sdk-go/v2/model"

type ServerlessWorkflow = sw.Workflow

// StateResult represents the result of a state execution
type StateResult struct {
	Data      interface{}
	NextState string
	Error     error
}
