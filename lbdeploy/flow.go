package lbdeploy

// FlowMap holds a set of deployment flows mapped by their identifiers.
type FlowMap map[FlowID]Flow

// FlowID is a unique identifier for a flow within a deployment.
type FlowID string

// Flow is a flow of actions within a deployment.
type Flow struct {
	Preconditions ConditionList `json:"preconditions,omitzero"`
	Locks         []LockID      `json:"locks,omitzero"`
	Actions       []Action      `json:"actions,omitzero"`
}
