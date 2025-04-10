package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
	"github.com/leafbridge/leafbridge-deploy/lbengine"
	"github.com/leafbridge/leafbridge-deploy/lbevent"
)

// DeployCmd deploys software according to a LeafBridge deployment
// configuration.
type DeployCmd struct {
	ConfigFile string          `kong:"required,name='config-file',help='Path to a deployment file describing the deployment.'"`
	Flow       lbdeploy.FlowID `kong:"required,name='flow',help='The flow to invoke within the deployment.'"`
	ShowConfig bool            `kong:"optional,name='show-config',help='Show the loaded configuration.'"`
	Verbose    bool            `kong:"optional,name='verbose',short='v',help='Show debug messages on the command line.'"`
}

// Run executes the LeafBridge deploy command.
func (cmd DeployCmd) Run(ctx context.Context) error {
	// Read the deployment file.
	dep, err := loadDeployment(cmd.ConfigFile)
	if err != nil {
		return err
	}

	// Print the loaded configuration if requested.
	if cmd.ShowConfig {
		out, err := json.MarshalIndent(dep, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(out))
		return nil
	}

	// Select an event recorder.
	/*
		recorder := lbevent.Recorder{Handler: lbevent.LoggedHandler{}}
		recorder := lbevent.Recorder{Handler: lbevent.LoggedHandler{
			Handler: slog.NewJSONHandler(os.Stdout, nil),
		}}
	*/

	// Attempt to use a Windows event handler, but carry on regardless if it
	// doens't work out. The most likely reason it won't work is if the
	// running process isn't elevated.
	var handler lbevent.Handler
	{
		min := slog.LevelInfo
		if cmd.Verbose {
			min = slog.LevelDebug
		}
		basicHandler := lbevent.NewBasicHandler(os.Stdout, min)
		windowsHandler, err := lbevent.NewWindowsHandler()
		if err != nil {
			handler = basicHandler
		} else {
			handler = lbevent.MultiHandler{basicHandler, windowsHandler}
		}
	}
	recorder := lbevent.Recorder{Handler: handler}

	// Prepare a new deployment engine for the deployment.
	engine := lbengine.NewDeploymentEngine(dep, lbengine.Options{
		Events: recorder,
	})

	// Invoke the requested flow within the deployment.
	return engine.Invoke(ctx, cmd.Flow)
}
