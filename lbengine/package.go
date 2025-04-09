package lbengine

import (
	"context"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
	"github.com/leafbridge/leafbridge-deploy/lbevent"
	"github.com/leafbridge/leafbridge-deploy/stagingfs"
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

// preparePackage performs a package preparation action as part of a
// LeafBridge deployment.
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
