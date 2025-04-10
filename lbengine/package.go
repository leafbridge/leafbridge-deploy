package lbengine

import (
	"context"

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
	}

	// Download and verify the package data.
	//
	// If the file already contains the expected data, the download will be
	// skipped.
	//
	// If the file was partially downloaded, the download will be resumed.
	if err := de.DownloadAndVerifyPackage(ctx, engine.pkg, file); err != nil {
		return err
	}

	// Extract the package.
	return engine.extractPackage(ctx, file)
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

func (engine *packageEngine) extractPackage(ctx context.Context, source stagingfs.PackageFile) error {
	// Create a temporary directory to hold the extracted files.
	destination, err := tempfs.OpenExtractionDirForPackage(lbdeploy.PackageContent{
		ID:          engine.pkg.ID,
		PrimaryHash: engine.pkg.Definition.Attributes.Hashes.Primary(),
	}, tempfs.Options{
		DeleteOnClose: true,
	})
	if err != nil {
		return err
	}

	// Delete all of the extracted files when we are finished.
	defer destination.Close()

	// Extract the files.
	ee := extractionEngine{
		deployment: engine.deployment,
		flow:       engine.flow,
		action:     engine.action,
		events:     engine.events,
	}

	return ee.extractPackage(ctx, source, destination)
}
