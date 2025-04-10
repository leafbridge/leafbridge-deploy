package lbengine

import (
	"context"
	"fmt"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
	"github.com/leafbridge/leafbridge-deploy/lbevent"
	"github.com/leafbridge/leafbridge-deploy/stagingfs"
	"github.com/leafbridge/leafbridge-deploy/tempfs"
)

// packageData holds the ID and definition for a package.
type packageData struct {
	ID         lbdeploy.PackageID
	Definition lbdeploy.Package
}

// packageEngine manages package-related actions.
type packageEngine struct {
	deployment lbdeploy.Deployment
	flow       flowData
	action     actionData
	pkg        packageData
	events     lbevent.Recorder
	state      *engineState
}

// preparePackage performs a package preparation action.
func (engine *packageEngine) PreparePackage(ctx context.Context) error {
	// Open the package file, or create it if it doesn't exist.
	file, err := engine.openPackageFile()
	if err != nil {
		return err
	}
	defer file.Close()

	// Prepare a download engine.
	de := downloadEngine{
		deployment: engine.deployment,
		flow:       engine.flow,
		action:     engine.action,
		events:     engine.events,
		state:      engine.state,
	}

	// Download and verify the package data.
	//
	// If the file already contains the expected data, the download will be
	// skipped.
	//
	// If the file was partially downloaded, the download will be resumed.
	return de.DownloadAndVerifyPackage(ctx, engine.pkg, file)
}

// InvokeCommand performs a package command invocation action.
func (engine *packageEngine) InvokeCommand(ctx context.Context, command lbdeploy.PackageCommandID) error {
	// Find the command within the package.
	commandDefinition, exists := engine.pkg.Definition.Commands[command]
	if !exists {
		return fmt.Errorf("the command \"%s\" does not exist within the \"%s\" package", command, engine.pkg.ID)
	}

	// Check the state to see whether we've already downloaded, verified and
	// extracted the files in this package.
	extractedFiles, alreadyExtracted := engine.state.extractedPackages[engine.pkg.ID]

	// Download, verify and extract the package if we haven't done so already.
	if !alreadyExtracted {
		// Open the package file, or create it if it doesn't exist.
		packageFile, err := engine.openPackageFile()
		if err != nil {
			return fmt.Errorf("failed to prepare package file: %w", err)
		}
		defer packageFile.Close()

		// Prepare a download engine.
		de := downloadEngine{
			deployment: engine.deployment,
			flow:       engine.flow,
			action:     engine.action,
			events:     engine.events,
			state:      engine.state,
		}

		// Download and verify the package data.
		//
		// If the file already contains the expected data, the download will be
		// skipped.
		//
		// If the file was partially downloaded, the download will be resumed.
		if err := de.DownloadAndVerifyPackage(ctx, engine.pkg, packageFile); err != nil {
			return err
		}

		// Create a temporary directory to hold the extracted files.
		extractedFiles, err = tempfs.OpenExtractionDirForPackage(lbdeploy.PackageContent{
			ID:          engine.pkg.ID,
			PrimaryHash: engine.pkg.Definition.Attributes.Hashes.Primary(),
		}, tempfs.Options{
			DeleteOnClose: true,
		})
		if err != nil {
			return fmt.Errorf("failed to prepare a directory for file extraction: %w", err)
		}

		// Prepare an extraction engine.
		ee := extractionEngine{
			deployment: engine.deployment,
			flow:       engine.flow,
			action:     engine.action,
			events:     engine.events,
			state:      engine.state,
		}

		// Extract the files.
		if err := ee.ExtractPackage(ctx, packageFile, extractedFiles); err != nil {
			extractedFiles.Close()
			return fmt.Errorf("extraction failed: %w", err)
		}

		// Add the extracted files to the engine's state, so that they'll be
		// available for other flows.
		//
		// This will also cause the deployment engine to close the extracted
		// files after the orginating flow has finished.
		engine.state.extractedPackages[engine.pkg.ID] = extractedFiles
	}

	// Prepare a command engine.
	ce := commandEngine{
		deployment: engine.deployment,
		flow:       engine.flow,
		action:     engine.action,
		pkg:        engine.pkg,
		command: commandData{
			ID:         command,
			Definition: commandDefinition,
		},
		events: engine.events,
		state:  engine.state,
	}

	// Invoke the command.
	return ce.Invoke(ctx, extractedFiles)
}

func (engine *packageEngine) openPackageFile() (stagingfs.PackageFile, error) {
	// Open the deployment's staging directory.
	deployDir, err := stagingfs.OpenDeployment(engine.deployment.ID)
	if err != nil {
		return stagingfs.PackageFile{}, err
	}
	defer deployDir.Close()

	// Open the package's staging directory.
	packageDir, err := deployDir.OpenPackage(lbdeploy.PackageContent{
		ID:          engine.pkg.ID,
		PrimaryHash: engine.pkg.Definition.Attributes.Hashes.Primary(),
	})
	if err != nil {
		return stagingfs.PackageFile{}, err
	}
	defer packageDir.Close()

	// Open the package file, or create it if it doesn't exist.
	return packageDir.OpenFile(engine.pkg.Definition)
}
