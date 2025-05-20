package parser

import (
	"fmt"
	"os"

	"github.com/kshitiz1403/jsonjuggler/activities"
	sw "github.com/serverlessworkflow/sdk-go/v2/model"
	"github.com/serverlessworkflow/sdk-go/v2/parser"
)

// Parser handles workflow DSL parsing and validation
type Parser struct {
	registry *activities.Registry
}

// NewParser creates a new workflow parser with an activity registry
func NewParser(registry *activities.Registry) *Parser {
	return &Parser{
		registry: registry,
	}
}

// ParseFromFile parses a workflow from a JSON file
func (p *Parser) ParseFromFile(filePath string) (*sw.Workflow, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	return p.ParseFromBytes(data)
}

// ParseFromBytes parses a workflow from JSON bytes
func (p *Parser) ParseFromBytes(data []byte) (*sw.Workflow, error) {
	workflow, err := parser.FromJSONSource(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}

	// Additional custom validation if needed
	if err := p.validateCustomRules(workflow); err != nil {
		return nil, err
	}

	return workflow, nil
}

// validateCustomRules performs any additional validation beyond what the SDK provides
func (p *Parser) validateCustomRules(workflow *sw.Workflow) error {
	// Track all unique activities referenced in the workflow
	referencedActivities := make(map[string]bool)

	// Collect all activity references from states
	for _, state := range workflow.States {
		switch state.GetType() {
		case sw.StateTypeOperation:
			opState := state.(*sw.OperationState)
			for _, action := range opState.Actions {
				if action.FunctionRef == nil || action.FunctionRef.RefName == "" {
					return fmt.Errorf("state '%s' has an action with missing function reference", state.GetName())
				}
				referencedActivities[action.FunctionRef.RefName] = true
			}
		}
	}

	// Validate all referenced activities are registered
	for activityName := range referencedActivities {
		if _, exists := p.registry.Get(activityName); !exists {
			return fmt.Errorf("activity '%s' is referenced in workflow but not registered", activityName)
		}
	}

	return nil
}
