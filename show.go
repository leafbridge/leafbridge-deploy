package main

import (
	"context"
	"encoding/json"
	"fmt"
)

// ShowCmd shows information that is relevant to a LeafBridge deployment.
type ShowCmd struct {
	Config ShowConfigCmd `kong:"cmd,help='Shows configuration loaded from a deployment configuration file.'"`
}

// ShowConfigCmd shows the configuration of a LeafBridge deployment.
type ShowConfigCmd struct {
	ConfigFile string `kong:"required,name='config-file',help='Path to a deployment file describing the deployment.'"`
}

// Run executes the LeafBridge show config command.
func (cmd ShowConfigCmd) Run(ctx context.Context) error {
	// Read the deployment file.
	dep, err := loadDeployment(cmd.ConfigFile)
	if err != nil {
		return err
	}

	// Print the loaded configuration.
	out, err := json.MarshalIndent(dep, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}
