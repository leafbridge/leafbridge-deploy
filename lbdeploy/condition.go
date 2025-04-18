package lbdeploy

import (
	"fmt"
	"strings"
)

// ConditionMap holds a set of conditions mapped by their identifiers.
type ConditionMap map[ConditionID]Condition

// ConditionList is a list of condition IDs.
type ConditionList []ConditionID

// String returns a string representation of the list.
func (list ConditionList) String() string {
	var out strings.Builder
	for i, item := range list {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(string(item))
	}
	return out.String()
}

// ConditionID is a unique identifier for a condition.
type ConditionID string

// ConditionType identifies a type of condition.
type ConditionType string

// Supported condition types.
const (
	ConditionProcessIsRunning ConditionType = "resource.process:running"
	ConditionMutexExists      ConditionType = "resource.mutex:exists"
	ConditionDirectoryExists  ConditionType = "resource.file-system.directory:exists"
	ConditionFileExists       ConditionType = "resource.file-system.file:exists"
)

// Condition describes a condition that can be evaluated.
type Condition struct {
	Label     string        `json:"label,omitempty"`
	Type      ConditionType `json:"type,omitempty"`
	Value     string        `json:"value,omitempty"`
	Negated   bool          `json:"negated,omitempty"`
	Any       []Condition   `json:"any,omitzero"`
	All       []Condition   `json:"all,omitzero"`
	Violation string        `json:"violation,omitempty"`
}

// ConditionElement identifies an element of a condition.
type ConditionElement int

// Elements of a condition that can lead to an error.
const (
	ConditionElementSelf ConditionElement = iota
	ConditionElementAny
	ConditionElementAll
)

// ConditionError is returned when a condition fails due to an error.
type ConditionError struct {
	Label        string
	Type         ConditionType
	Element      ConditionElement
	SubCondition int
	Err          error
}

// Unwrap returns the underlying error for the condition.
func (e ConditionError) Unwrap() error {
	return e.Err
}

// Error returns the error as a string.
func (e ConditionError) Error() string {
	switch e.Element {
	case ConditionElementAny:
		if e.Label != "" {
			return fmt.Sprintf("%s: Any [%d]: %s", e.Label, e.SubCondition, e.Err)
		}
		return fmt.Sprintf("Any [%d]: %s", e.SubCondition, e.Err)
	case ConditionElementAll:
		if e.Label != "" {
			return fmt.Sprintf("%s: All [%d]: %s", e.Label, e.SubCondition, e.Err)
		}
		return fmt.Sprintf("All [%d]: %s", e.SubCondition, e.Err)
	default:
		if e.Type != "" {
			if e.Label != "" {
				return fmt.Sprintf("%s: %s: %s", e.Label, e.Type, e.Err)
			}
			return fmt.Sprintf("%s: %s", e.Type, e.Err)
		}
		if e.Label != "" {
			return fmt.Sprintf("%s: %s", e.Label, e.Err)
		}
		return e.Err.Error()
	}
}

func conditionSelfError(c Condition, err error) error {
	return ConditionError{
		Label:   c.Label,
		Type:    c.Type,
		Element: ConditionElementSelf,
		Err:     err,
	}
}
