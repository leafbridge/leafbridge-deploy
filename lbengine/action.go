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
		case "prepare-package":
			if err := engine.preparePackage(ctx); err != nil {
				return err
			}
		//case "invoke-package":
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
	}

	// Execute the prepare-package action via the package engine.
	return pe.PreparePackage(ctx)
}

/*
// preparePackage performs a package preparation action as part of a
// LeafBridge deployment.
func (engine *actionEngine) preparePackage(ctx context.Context) error {
	// Look up the package by its ID.
	pkg, found := engine.deployment.Resources.Packages[engine.action.Definition.Package]
	if !found {
		return fmt.Errorf("the package \"%s\" does not exist within the \"%s\" deployment", engine.action.Definition.Package, engine.deployment.ID)
	}

	// Open the deployment's staging directory.
	deployDir, err := stagingfs.OpenDeployment(engine.deployment.ID)
	if err != nil {
		return err
	}
	defer deployDir.Close()

	// Open the package's staging directory.
	packageDir, err := deployDir.OpenPackage(lbdeploy.PackageContent{
		ID:          engine.action.Definition.Package,
		PrimaryHash: pkg.Attributes.Hashes.Primary(),
	})
	if err != nil {
		return err
	}
	defer packageDir.Close()

	// Open the package file, or create it if it doesn't exist.
	file, err := packageDir.OpenFile(pkg)
	if err != nil {
		return err
	}
	defer file.Close()

	// Download and verify the package data.
	//
	// If the file already contains the expected data, the download will be
	// skipped.
	//
	// If the file was partially downloaded, the download will be resumed.
	return engine.downloadAndVerify(ctx, flow, id, pkg, file)
}
*/
