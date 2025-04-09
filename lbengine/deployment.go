package lbengine

import (
	"context"
	"fmt"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
	"github.com/leafbridge/leafbridge-deploy/lbevent"
)

// DeploymentEngine is a LeafBridge engine that is responsible for invocation
// of deployments.
type DeploymentEngine struct {
	deployment lbdeploy.Deployment
	events     lbevent.Recorder
}

// NewDeploymentEngine returns a new LeafBridge deployment engine for the
// given deployment and options.
func NewDeploymentEngine(deployment lbdeploy.Deployment, opts Options) DeploymentEngine {
	return DeploymentEngine{
		deployment: deployment,
		events:     opts.Events,
	}
}

// Invoke executes a flow within a LeafBridge deployment.
func (engine DeploymentEngine) Invoke(ctx context.Context, flow lbdeploy.FlowID) error {
	// TODO: Generate some sort of random UUID for the deployment invocation
	// that can be used for log analysis?

	// Ensure that the deployment is valid.
	if err := engine.deployment.Validate(); err != nil {
		return err
	}

	// Find the requested flow within the deployment.
	definition, found := engine.deployment.Flows[flow]
	if !found {
		return fmt.Errorf("the flow \"%s\" does not exist within the \"%s\" deployment", flow, engine.deployment.ID)
	}

	// Invoke the requested flow.
	fe := flowEngine{
		deployment: engine.deployment,
		flow: flowData{
			ID:         flow,
			Definition: definition,
		},
		events: engine.events,
	}

	return fe.Invoke(ctx)
}
