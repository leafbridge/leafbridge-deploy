package lbdeploy

// MutexMap holds a set of mutex resources mapped by their identifiers.
type MutexMap map[MutexID]Mutex

// MutexID is a unique identifier for a system-wide mutex resource.
type MutexID string

// MutexName is the name of a system-wide mutex on the machine a
// deployment is running on.
type MutexName string

// MutexNamespace is the namespace within a mutex exists.
// It can be "leafbridge", "global" or "system".
type MutexNamespace string

// Mutex is a system-wide mutex that can be evaluated by conditions or used
// by locks.
type Mutex struct {
	Description string         `json:"description,omitempty"`
	Name        MutexName      `json:"name"`
	Namespace   MutexNamespace `json:"namespace"`
}
