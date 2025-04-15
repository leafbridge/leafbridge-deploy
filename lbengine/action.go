package lbengine

import (
	"context"
	"fmt"
	"time"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
	"github.com/leafbridge/leafbridge-deploy/lbdeployevent"
	"github.com/leafbridge/leafbridge-deploy/lbevent"
)

// actionData holds the index and definition for an action.
type actionData struct {
	Index      int
	Definition lbdeploy.Action
}

// actionEngine manages execution of an action within a flow.
type actionEngine struct {
	deployment lbdeploy.Deployment
	flow       flowData
	action     actionData
	events     lbevent.Recorder
	force      bool
	state      *engineState
}

func (engine *actionEngine) Invoke(ctx context.Context) error {
	// Record the start of the action.
	engine.events.Record(lbdeployevent.ActionStarted{
		Deployment:  engine.deployment.ID,
		Flow:        engine.flow.ID,
		ActionIndex: engine.action.Index,
		ActionType:  engine.action.Definition.Type,
	})

	// Record the time that the action started.
	started := time.Now()

	// Execute the action.
	err := func() error {
		switch engine.action.Definition.Type {
		case "start-flow":
			if err := engine.startFlow(ctx); err != nil {
				return err
			}
		case "prepare-package":
			if err := engine.preparePackage(ctx); err != nil {
				return err
			}
		case "invoke-package":
			if err := engine.invokePackage(ctx); err != nil {
				return err
			}
		case "copy-file":
			if err := engine.copyFile(ctx); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unrecognized deployment action type \"%s\"", engine.action.Definition.Type)
		}
		return nil
	}()

	// Record the time that the action stopped.
	stopped := time.Now()

	// Record the end of the action.
	engine.events.Record(lbdeployevent.ActionStopped{
		Deployment:  engine.deployment.ID,
		Flow:        engine.flow.ID,
		ActionIndex: engine.action.Index,
		ActionType:  engine.action.Definition.Type,
		Started:     started,
		Stopped:     stopped,
		Err:         err,
	})

	return err
}

// startFlow starts another flow within the LeafBridge deployment.
func (engine *actionEngine) startFlow(ctx context.Context) error {
	flow := engine.action.Definition.Flow

	// Find the requested flow within the deployment.
	definition, found := engine.deployment.Flows[flow]
	if !found {
		return fmt.Errorf("the flow \"%s\" does not exist within the \"%s\" deployment", flow, engine.deployment.ID)
	}

	// Prepare the flow engine.
	fe := flowEngine{
		deployment: engine.deployment,
		flow: flowData{
			ID:         flow,
			Definition: definition,
		},
		events: engine.events,
		force:  engine.force,
		state:  engine.state,
	}

	// Invoke the requested flow.
	return fe.Invoke(ctx)
}

// preparePackage performs a package preparation action as part of a
// LeafBridge deployment.
func (engine *actionEngine) preparePackage(ctx context.Context) error {
	// Look up the package by its ID.
	pkg, found := engine.deployment.Resources.Packages[engine.action.Definition.Package]
	if !found {
		return fmt.Errorf("the package \"%s\" does not exist within the \"%s\" deployment", engine.action.Definition.Package, engine.deployment.ID)
	}

	// Prepare a package engine.
	pe := packageEngine{
		deployment: engine.deployment,
		flow:       engine.flow,
		action:     engine.action,
		pkg: packageData{
			ID:         engine.action.Definition.Package,
			Definition: pkg,
		},
		events: engine.events,
		force:  engine.force,
		state:  engine.state,
	}

	// Execute the prepare-package action via the package engine.
	return pe.PreparePackage(ctx)
}

// invokePackage invokes a package command action.
func (engine *actionEngine) invokePackage(ctx context.Context) error {
	// Look up the package by its ID.
	pkg, found := engine.deployment.Resources.Packages[engine.action.Definition.Package]
	if !found {
		return fmt.Errorf("the package \"%s\" does not exist within the \"%s\" deployment", engine.action.Definition.Package, engine.deployment.ID)
	}

	// Prepare a package engine.
	pe := packageEngine{
		deployment: engine.deployment,
		flow:       engine.flow,
		action:     engine.action,
		pkg: packageData{
			ID:         engine.action.Definition.Package,
			Definition: pkg,
		},
		events: engine.events,
		force:  engine.force,
		state:  engine.state,
	}

	// Execute the install-package action via the package engine.
	return pe.InvokeCommand(ctx, engine.action.Definition.Command)
}

// copyFile performs a file copy operation.
func (engine *actionEngine) copyFile(ctx context.Context) error {
	// Prepare a file engine.
	fe := fileEngine{
		deployment: engine.deployment,
		flow:       engine.flow,
		action:     engine.action,
		events:     engine.events,
		state:      engine.state,
	}

	// Execute the copy-file action via the file engine.
	return fe.CopyFile(ctx)
}
