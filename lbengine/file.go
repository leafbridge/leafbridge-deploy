package lbengine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/leafbridge/leafbridge-deploy/filetime"
	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
	"github.com/leafbridge/leafbridge-deploy/lbdeployevent"
	"github.com/leafbridge/leafbridge-deploy/lbevent"
	"github.com/leafbridge/leafbridge-deploy/localfs"
)

// fileEngine handles file system operations within a deployment.
type fileEngine struct {
	deployment lbdeploy.Deployment
	flow       flowData
	action     actionData
	events     lbevent.Recorder
	state      *engineState
}

// CopyFile performs a file copy operation.
func (engine *fileEngine) CopyFile(ctx context.Context) error {
	// Find the relevant source file within the deployment.
	sourceFileID := engine.action.Definition.SourceFile
	sourceFileRef, err := engine.deployment.Resources.FileSystem.ResolveFile(sourceFileID)
	if err != nil {
		return fmt.Errorf("source file: %w", err)
	}

	// Find the relevant destination file within the deployment.
	destFileID := engine.action.Definition.DestinationFile
	destFileRef, err := engine.deployment.Resources.FileSystem.ResolveFile(destFileID)
	if err != nil {
		return fmt.Errorf("destination file: %w", err)
	}

	// Record the time that the file copy started.
	started := time.Now()

	var (
		sourceFilePath string
		destFilePath   string
		fileSize       int64
	)
	err = func() error {
		// Open the root above the destination file.
		destDir, err := localfs.OpenDir(destFileRef.Dir())
		if err != nil {
			return fmt.Errorf("unable to open the destination directory: %w", err)
		}
		defer destDir.Close()

		// Record the destination path for event logging.
		{
			localized, err := filepath.Localize(destFileRef.FilePath)
			if err == nil {
				destFilePath = filepath.Join(destDir.Path(), localized)
			}
		}

		// If there is an existing file, stop.
		fi, err := destDir.System().Stat(destFileRef.FilePath)
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("unable to evaluate the destination file: %w", err)
			}
		} else if fi.Mode().IsRegular() {
			// The file already exists.
			//
			// TODO: Support replacing existing files, optionally via
			// configuration.
			return nil
		} else {
			return errors.New("the destination file path already exists but is not a regular file")
		}

		// Open the source file.
		sourceFile, err := localfs.OpenFile(sourceFileRef)
		if err != nil {
			return fmt.Errorf("unable to open the source file: %w", err)
		}
		defer sourceFile.Close()

		// Record the source path and file size for event logging.
		sourceFilePath = sourceFile.Path()
		if fi, err := sourceFile.System().Stat(); err == nil {
			fileSize = fi.Size()
		}

		// Open the destination file.
		destFile, err := destDir.System().Create(destFileRef.FilePath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		// Record the destination path for event logging.
		sourceFilePath = sourceFile.Path()

		// Copy file data.
		if _, err := io.Copy(destFile, sourceFile.System()); err != nil {
			return err
		}

		// Copy the file modification date.
		sourceFileInfo, err := sourceFile.System().Stat()
		if err != nil {
			return err
		}
		if modTime := sourceFileInfo.ModTime(); !modTime.IsZero() {
			if err := filetime.SetFileModificationTime(destFile, modTime); err != nil {
				return fmt.Errorf("failed to set file modification time: %w", err)
			}
		}
		return nil
	}()

	// Record the time that the file copy stopped.
	stopped := time.Now()

	// Record the file copy.
	engine.events.Record(lbdeployevent.FileCopy{
		Deployment:      engine.deployment.ID,
		Flow:            engine.flow.ID,
		Action:          engine.action.Definition.Type,
		SourceID:        sourceFileID,
		SourcePath:      sourceFilePath,
		DestinationID:   destFileID,
		DestinationPath: destFilePath,
		FileSize:        fileSize,
		Started:         started,
		Stopped:         stopped,
		Err:             err,
	})

	return nil
}
