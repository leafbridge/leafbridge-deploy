package lbengine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/leafbridge/leafbridge-deploy/internal/mergereader"
	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
	"github.com/leafbridge/leafbridge-deploy/lbdeployevent"
	"github.com/leafbridge/leafbridge-deploy/lbevent"
	"github.com/leafbridge/leafbridge-deploy/tempfs"
)

// commandData holds the ID and definition for a package.
type commandData struct {
	ID         lbdeploy.PackageCommandID
	Definition lbdeploy.PackageCommand
}

// commandEngine manages invocation of a command for a package.
type commandEngine struct {
	deployment lbdeploy.Deployment
	flow       flowData
	action     actionData
	pkg        packageData
	command    commandData
	apps       lbdeploy.AppEvaluation
	events     lbevent.Recorder
	force      bool
	state      *engineState
}

// Invoke runs the command.
func (engine *commandEngine) Invoke(ctx context.Context, files tempfs.ExtractionDir) error {
	// Get information about the executable file from the package.
	fileID := engine.command.Definition.Executable
	fileData, exists := engine.pkg.Definition.Files[fileID]
	if !exists {
		return fmt.Errorf("the command \"%s\" refers to an executable file \"%s\" that is not defined in the \"%s\" package", engine.command.ID, fileID, engine.pkg.ID)
	}

	// Verify that the executable file exists within the extracted file set.
	fi, err := files.Stat(fileData.Path)
	if err != nil {
		return fmt.Errorf("verification of the command executable failed: %w", err)
	}
	if !fi.Mode().IsRegular() {
		return errors.New("verification of the command executable failed: the executable file path is not a regular file")
	}

	// Prepare an absolute path for the command and its working directory.
	execPath, err := files.FilePath(fileData.Path)
	if err != nil {
		return fmt.Errorf("an executable file path could not be prepared for the \"%s\" command in the \"%s\" package: %w", engine.command.ID, engine.pkg.ID, err)
	}

	execDir := filepath.Dir(execPath)
	if execDir == "" {
		return fmt.Errorf("a working directory could not be determined for the \"%s\" command in the \"%s\" package", engine.command.ID, engine.pkg.ID)
	}

	// Check for cancellation before starting the command.
	if err := ctx.Err(); err != nil {
		return err
	}

	// Prepare a command that will be terminated when ctx is cancelled.
	cmd := exec.CommandContext(ctx, execPath, engine.command.Definition.Args...)

	// Set the command's working directory.
	cmd.Dir = execDir

	// Configure the command to wait up to one minute for the command to close
	// out gracefully when its context is cancelled.
	//
	// TODO: Make this configurable.
	cmd.WaitDelay = time.Minute

	// Prepare two sets of output pipes for the command.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Record the start of the command.
	engine.events.Record(lbdeployevent.CommandStarted{
		Deployment:  engine.deployment.ID,
		Flow:        engine.flow.ID,
		ActionIndex: engine.action.Index,
		ActionType:  engine.action.Definition.Type,
		Package:     engine.pkg.ID,
		Command:     engine.command.ID,
		CommandLine: cmd.String(),
		Apps:        engine.apps,
	})

	// Prepare a buffer to hold the combined command output.
	var output bytes.Buffer

	// Record the time that the command started.
	started := time.Now()

	// Start the command.
	err = cmd.Start()

	// If the command started successfully, send its output to stdout and
	// stderr as well as the output buffer, then wait for it to finish.
	if err == nil {
		// Tee stdout and stderr to the console.
		r1 := io.TeeReader(stdout, os.Stdout)
		r2 := io.TeeReader(stderr, os.Stderr)

		// Combine the output of both stdout and stderr.
		merged := mergereader.New(r1, r2)

		// Read the combined output from the command.
		io.Copy(&output, merged)

		// Wait for the command to be completed.
		err = cmd.Wait()
	}

	// Record the time that the command stopped.
	stopped := time.Now()

	// Evaluate the effectiveness of any expected application changes.
	appSummary, appSummaryErr := SummarizeAppChanges(engine.deployment.Apps, engine.apps)
	if appSummaryErr != nil {
		appSummaryErr = fmt.Errorf("failed to determine the state of installed applications after the command was invoked: %w", appSummaryErr)
		if err == nil {
			err = appSummaryErr
		}
	}

	// Record the end of the command.
	engine.events.Record(lbdeployevent.CommandStopped{
		Deployment:  engine.deployment.ID,
		Flow:        engine.flow.ID,
		ActionIndex: engine.action.Index,
		ActionType:  engine.action.Definition.Type,
		Package:     engine.pkg.ID,
		Command:     engine.command.ID,
		CommandLine: cmd.String(),
		Output:      output.String(),
		AppsBefore:  engine.apps,
		AppsAfter:   appSummary,
		Started:     started,
		Stopped:     stopped,
		Err:         err,
	})

	// Wait 5 seconds to let the file system and file locks quiesce before
	// continuing on. This is especially important if this is the last action
	// and LeafBridge attempts to delete extracted files immediately after
	// this command has run.
	//
	// TODO: Make this delay configurable.
	timer := time.NewTimer(time.Second * 5)
	select {
	case <-ctx.Done():
		timer.Stop()
	case <-timer.C:
	}

	// If the command returned an error, return that.
	if err != nil {
		return err
	}

	// If the application summary indicates that an expected change to the
	// installed set of applications didn't take effect, return the error.
	return appSummary.Err()
}
