package lbengine

import (
	"fmt"
	"os"

	"github.com/gentlemanautomaton/winobj/winmutex"
	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
	"github.com/leafbridge/leafbridge-deploy/localfs"
	"github.com/leafbridge/leafbridge-deploy/localregistry"
)

// ConditionEngine is responsible for evaluating conditions on the local
// system.
type ConditionEngine struct {
	deployment lbdeploy.Deployment
}

// NewConditionEngine prepare a condition engine for the given deployment.
func NewConditionEngine(dep lbdeploy.Deployment) ConditionEngine {
	return ConditionEngine{
		deployment: dep,
	}
}

// Evaluate returns true if the given condition is currently true.
//
// TODO: Consider returning some sort of evaluation struct that describes
// the condition (or subconditions) that failed.
func (engine ConditionEngine) Evaluate(condition lbdeploy.ConditionID) (bool, error) {
	definition, found := engine.deployment.Conditions[condition]
	if !found {
		return false, fmt.Errorf("the condition \"%s\" does not exist within the \"%s\" deployment", condition, engine.deployment.ID)
	}

	return engine.evaluate(definition)
}

func (engine ConditionEngine) evaluate(condition lbdeploy.Condition) (bool, error) {
	// Evaluate the condition.
	result, err := func() (bool, error) {
		// Evaluate "any" conditions.
		if len(condition.Any) > 0 {
			for i, candidate := range condition.Any {
				result, err := engine.evaluate(candidate)
				if err != nil {
					return false, lbdeploy.ConditionError{
						Label:        condition.Label,
						Type:         condition.Type,
						Element:      lbdeploy.ConditionElementAny,
						SubCondition: i,
						Err:          err,
					}
				}
				if result {
					return true, nil
				}
			}
			return false, nil
		}

		// Evaluate "all" conditions.
		if len(condition.All) > 0 {
			for i, candidate := range condition.All {
				result, err := engine.evaluate(candidate)
				if err != nil {
					return false, lbdeploy.ConditionError{
						Label:        condition.Label,
						Type:         condition.Type,
						Element:      lbdeploy.ConditionElementAll,
						SubCondition: i,
						Err:          err,
					}
				}
				if !result {
					return false, nil
				}
			}
			return true, nil
		}

		// Evaluate individual conditions.
		switch condition.Type {
		case lbdeploy.ConditionProcessIsRunning:
			process, found := engine.deployment.Resources.Processes[lbdeploy.ProcessResourceID(condition.Value)]
			if !found {
				return false, conditionSelfError(condition, fmt.Errorf("the \"%s\" process is not defined in the deployment", condition.Value))
			}
			running, err := NumberOfRunningProcesses(process.Match)
			if err != nil {
				return false, conditionSelfError(condition, err)
			}
			return running > 0, nil
		case lbdeploy.ConditionMutexExists:
			mutex, found := engine.deployment.Resources.Mutexes[lbdeploy.MutexID(condition.Value)]
			if !found {
				return false, conditionSelfError(condition, fmt.Errorf("the \"%s\" mutex is not defined in the deployment", condition.Value))
			}
			name, err := mutex.ObjectName()
			if err != nil {
				return false, conditionSelfError(condition, err)
			}
			exists, err := winmutex.Exists(name)
			if err != nil {
				return false, conditionSelfError(condition, err)
			}
			return exists, nil
		case lbdeploy.ConditionRegistryKeyExists:
			ref, err := engine.deployment.Resources.Registry.ResolveKey(lbdeploy.RegistryKeyResourceID(condition.Value))
			if err != nil {
				return false, conditionSelfError(condition, err)
			}
			key, err := localregistry.OpenKey(ref)
			if err != nil {
				if os.IsNotExist(err) {
					return false, nil
				}
				return false, conditionSelfError(condition, err)
			}
			defer key.Close()
			return true, nil
		case lbdeploy.ConditionRegistryValueExists:
			ref, err := engine.deployment.Resources.Registry.ResolveValue(lbdeploy.RegistryValueResourceID(condition.Value))
			if err != nil {
				return false, conditionSelfError(condition, err)
			}
			key, err := localregistry.OpenKey(ref.Key())
			if err != nil {
				if os.IsNotExist(err) {
					return false, nil
				}
				return false, conditionSelfError(condition, err)
			}
			defer key.Close()
			return key.HasValue(ref.ValueName)
		case lbdeploy.ConditionDirectoryExists:
			ref, err := engine.deployment.Resources.FileSystem.ResolveDirectory(lbdeploy.DirectoryResourceID(condition.Value))
			if err != nil {
				return false, conditionSelfError(condition, err)
			}
			dir, err := localfs.OpenDir(ref)
			if err != nil {
				if os.IsNotExist(err) {
					return false, nil
				}
				return false, conditionSelfError(condition, err)
			}
			defer dir.Close()
			return true, nil
		case lbdeploy.ConditionFileExists:
			ref, err := engine.deployment.Resources.FileSystem.ResolveFile(lbdeploy.FileResourceID(condition.Value))
			if err != nil {
				return false, conditionSelfError(condition, err)
			}
			dir, err := localfs.OpenDir(ref.Dir())
			if err != nil {
				if os.IsNotExist(err) {
					return false, nil
				}
				return false, conditionSelfError(condition, err)
			}
			defer dir.Close()
			fi, err := dir.System().Stat(ref.FilePath)
			if err != nil {
				return false, conditionSelfError(condition, err)
			}
			if fi.Mode().IsRegular() {
				return true, nil
			}
			path, err := ref.Path()
			if err != nil {
				return false, conditionSelfError(condition, fmt.Errorf("file \"%s\": the path exists but it is not a regular file", condition.Value))
			}
			return false, conditionSelfError(condition, fmt.Errorf("file \"%s\": the \"%s\" path exists but it is not a regular file", condition.Value, path))
		default:
			return false, conditionSelfError(condition, fmt.Errorf("unrecognized condition type: %s", condition.Type))
		}
	}()

	// Negate the result if requested.
	if condition.Negated {
		result = !result
	}

	return result, err
}

func conditionSelfError(c lbdeploy.Condition, err error) error {
	return lbdeploy.ConditionError{
		Label:   c.Label,
		Type:    c.Type,
		Element: lbdeploy.ConditionElementSelf,
		Err:     err,
	}
}
