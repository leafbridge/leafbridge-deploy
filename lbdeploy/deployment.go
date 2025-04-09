package lbdeploy

import (
	"errors"
)

// DeploymentID is a unique identifier for a deployment.
type DeploymentID string

// Validate returns a non-nil error if the deployment ID is invalid.
func (id DeploymentID) Validate() error {
	if id == "" {
		return errors.New("a deployment ID is missing")
	}
	return nil
}

// Deployment defines a deployment package.
type Deployment struct {
	ID        DeploymentID `json:"id,omitempty"`
	Name      string       `json:"name,omitempty"`
	Apps      AppMap       `json:"apps,omitempty"`
	Resources Resources    `json:"resources,omitempty"`
	Flows     FlowMap      `json:"flows,omitempty"`
}

// Validate returns a non-nil error if the deployment contains invalid
// configuration.
func (dep Deployment) Validate() error {
	if err := dep.ID.Validate(); err != nil {
		return err
	}

	return nil
}
