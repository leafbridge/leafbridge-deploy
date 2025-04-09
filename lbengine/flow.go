package lbengine

import (
	"context"
	"time"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
	"github.com/leafbridge/leafbridge-deploy/lbdeployevent"
	"github.com/leafbridge/leafbridge-deploy/lbevent"
)

// flowData holds the ID and definition for a flow.
type flowData struct {
	ID         lbdeploy.FlowID
	Definition lbdeploy.Flow
}

// flowEngine manages execution of a flow within a deployment.
type flowEngine struct {
	deployment lbdeploy.Deployment
	flow       flowData
	events     lbevent.Recorder
}

func (engine flowEngine) Invoke(ctx context.Context) error {
	// Record the start of the flow.
	engine.events.Record(lbdeployevent.FlowStarted{
		Deployment: engine.deployment.ID,
		Flow:       engine.flow.ID,
	})

	// Record the time that the flow started.
	started := time.Now()

	// Execute each action in the flow.
	err := func() error {
		for i, action := range engine.flow.Definition.Actions {
			ae := actionEngine{
				deployment: engine.deployment,
				flow:       engine.flow,
				action: actionData{
					Index:      i,
					Definition: action,
				},
				events: engine.events,
			}
			if err := ae.Invoke(ctx); err != nil {
				return err
			}
		}
		return nil
	}()

	// Record the time that the flow stopped.
	stopped := time.Now()

	// Record the end of the flow.
	engine.events.Record(lbdeployevent.FlowStopped{
		Deployment: engine.deployment.ID,
		Flow:       engine.flow.ID,
		Started:    started,
		Stopped:    stopped,
		Err:        err,
	})

	return err
}
