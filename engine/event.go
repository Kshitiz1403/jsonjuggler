package engine

import (
	"context"
	"fmt"

	sw "github.com/serverlessworkflow/sdk-go/v2/model"
)

func (e *Engine) executeEventBasedSwitch(ctx context.Context, state *sw.SwitchState, data *WorkflowData, stateExec *StateExecution) (*StateResult, error) {
	// TODO: Implement event-based switching when needed
	return nil, fmt.Errorf("event-based switch not implemented")
}
