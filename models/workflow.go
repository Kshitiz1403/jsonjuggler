package models

// Workflow represents a serverless workflow definition
type Workflow struct {
	Version     string  `json:"version"`
	SpecVersion string  `json:"specVersion"`
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Start       string  `json:"start"`
	Retries     []Retry `json:"retries,omitempty"`
	States      []State `json:"states"`
}

// State represents a workflow state
type State struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Actions    []Action `json:"actions"`
	Transition string   `json:"transition,omitempty"`
	End        bool     `json:"end,omitempty"`
}

// Action represents a state action
type Action struct {
	RetryRef    string      `json:"retryRef,omitempty"`
	FunctionRef FunctionRef `json:"functionRef"`
}

// FunctionRef represents a function reference
type FunctionRef struct {
	RefName   string         `json:"refName"`
	Arguments map[string]any `json:"arguments"`
}

// Retry represents a retry strategy
type Retry struct {
	Name        string `json:"name"`
	Delay       string `json:"delay"`
	Multiplier  int    `json:"multiplier"`
	MaxAttempts int    `json:"maxAttempts"`
}
