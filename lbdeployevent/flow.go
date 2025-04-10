package lbdeployevent

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
)

// TODO: Add some sort of random UUID for the deployment instance?

// FlowStarted is an event that occurs when a deployment flow has started.
type FlowStarted struct {
	Deployment lbdeploy.DeploymentID
	Flow       lbdeploy.FlowID
}

// Component identifies the component that generated the event.
func (e FlowStarted) Component() string {
	return "flow"
}

// Level returns the level of the event.
func (e FlowStarted) Level() slog.Level {
	return slog.LevelInfo
}

// Message returns a description of the event.
func (e FlowStarted) Message() string {
	return fmt.Sprintf("%s: %s: Starting invocation.", e.Deployment, e.Flow)
}

// Attrs returns a set of structured log attributes for the event.
func (e FlowStarted) Attrs() []slog.Attr {
	return []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
	}
}

// FlowStopped is an event that occurs when a deployment flow has stopped.
type FlowStopped struct {
	Deployment lbdeploy.DeploymentID
	Flow       lbdeploy.FlowID
	Started    time.Time
	Stopped    time.Time
	Err        error
}

// Component identifies the component that generated the event.
func (e FlowStopped) Component() string {
	return "flow"
}

// Level returns the level of the event.
func (e FlowStopped) Level() slog.Level {
	if e.Err != nil {
		return slog.LevelError
	}
	return slog.LevelInfo
}

// Message returns a description of the event.
func (e FlowStopped) Message() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: Stopped invocation due to an error: %s.", e.Deployment, e.Flow, e.Err)
	}
	return fmt.Sprintf("%s: %s: Completed invocation.", e.Deployment, e.Flow)
}

// Attrs returns a set of structured log attributes for the event.
func (e FlowStopped) Attrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
		slog.Time("started", e.Started),
		slog.Time("stopped", e.Stopped),
	}
	if e.Err != nil {
		attrs = append(attrs, slog.String("error", e.Err.Error()))
	}
	return attrs
}

// Duration returns the duration of the flow.
func (e FlowStopped) Duration() time.Duration {
	return e.Stopped.Sub(e.Started)
}

// FlowAlreadyRunning is an event that occurs when a deployment flow cannot
// be started because the flow is already running. This might indicate a cycle
// in the flow logic.
type FlowAlreadyRunning struct {
	Deployment lbdeploy.DeploymentID
	Flow       lbdeploy.FlowID
}

// Component identifies the component that generated the event.
func (e FlowAlreadyRunning) Component() string {
	return "flow"
}

// Level returns the level of the event.
func (e FlowAlreadyRunning) Level() slog.Level {
	return slog.LevelError
}

// Message returns a description of the event.
func (e FlowAlreadyRunning) Message() string {
	return fmt.Sprintf("%s: %s: Unable to start the flow. Another instance is already running. Is there a cycle in the flow logic?", e.Deployment, e.Flow)
}

// Attrs returns a set of structured log attributes for the event.
func (e FlowAlreadyRunning) Attrs() []slog.Attr {
	return []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
	}
}
