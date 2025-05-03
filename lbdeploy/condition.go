package lbdeploy

import (
	"fmt"
	"strings"

	"github.com/gentlemanautomaton/structformat"
	"github.com/leafbridge/leafbridge-deploy/lbvalue"
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

// ConditionCache holds a cache of evaluated conditions.
type ConditionCache map[ConditionID]bool

// ConditionID is a unique identifier for a condition.
type ConditionID string

// ConditionType identifies a type of condition.
type ConditionType string

// Supported condition types.
const (
	ConditionSubcondition            ConditionType = "condition"
	ConditionProcessIsRunning        ConditionType = "resource.process:running"
	ConditionMutexExists             ConditionType = "resource.mutex:exists"
	ConditionRegistryKeyExists       ConditionType = "resource.registry.key:exists"
	ConditionRegistryValueExists     ConditionType = "resource.registry.value:exists"
	ConditionRegistryValueComparison ConditionType = "resource.registry.value:comparison"
	ConditionDirectoryExists         ConditionType = "resource.file-system.directory:exists"
	ConditionFileExists              ConditionType = "resource.file-system.file:exists"
)

// Condition describes a condition that can be evaluated.
type Condition struct {
	Label      string             `json:"label,omitempty"`
	Type       ConditionType      `json:"type,omitempty"`
	Subject    string             `json:"subject,omitempty"`
	Comparison lbvalue.Comparison `json:"comparison,omitzero"`
	Value      lbvalue.Value      `json:"value,omitzero"`
	Negated    bool               `json:"negated,omitempty"`
	Any        []Condition        `json:"any,omitzero"`
	All        []Condition        `json:"all,omitzero"`
	Violation  string             `json:"violation,omitempty"`
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
	ID           ConditionID
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
	var builder structformat.Builder
	switch {
	case e.ID != "" && e.Label != "":
		builder.WritePrimary(fmt.Sprintf("%s (%s)", e.ID, e.Label))
	case e.ID != "":
		builder.WritePrimary(string(e.ID))
	case e.Label != "":
		builder.WritePrimary(string(e.Label))
	}

	switch e.Element {
	case ConditionElementAny:
		builder.WritePrimary(fmt.Sprintf("Any [%d]", e.SubCondition))
	case ConditionElementAll:
		builder.WritePrimary(fmt.Sprintf("All [%d]", e.SubCondition))
	default:
		if e.Type != "" {
			builder.WritePrimary(string(e.Type))
		}
	}

	builder.WriteStandard(e.Err.Error())

	return builder.String()
}

func conditionSelfError(c Condition, err error) error {
	return ConditionError{
		Label:   c.Label,
		Type:    c.Type,
		Element: ConditionElementSelf,
		Err:     err,
	}
}
