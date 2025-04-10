package lbdeployevent

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
)

// ActionStarted is an event that occurs when a deployment action has started.
type ActionStarted struct {
	Deployment  lbdeploy.DeploymentID
	Flow        lbdeploy.FlowID
	ActionIndex int
	ActionType  lbdeploy.ActionType
}

// Component identifies the component that generated the event.
func (e ActionStarted) Component() string {
	return "action"
}

// Level returns the level of the event.
func (e ActionStarted) Level() slog.Level {
	return slog.LevelInfo
}

// Message returns a description of the event.
func (e ActionStarted) Message() string {
	return fmt.Sprintf("%s: %s: %d: %s: Starting action.", e.Deployment, e.Flow, e.ActionIndex+1, e.ActionType)
}

// Attrs returns a set of structured log attributes for the event.
func (e ActionStarted) Attrs() []slog.Attr {
	return []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
		slog.Group("action", "index", e.ActionIndex, "type", e.ActionType),
	}
}

// ActionStopped is an event that occurs when a deployment action has stopped.
type ActionStopped struct {
	Deployment  lbdeploy.DeploymentID
	Flow        lbdeploy.FlowID
	ActionIndex int
	ActionType  lbdeploy.ActionType
	Started     time.Time
	Stopped     time.Time
	Err         error
}

// Component identifies the component that generated the event.
func (e ActionStopped) Component() string {
	return "action"
}

// Level returns the level of the event.
func (e ActionStopped) Level() slog.Level {
	if e.Err != nil {
		return slog.LevelError
	}
	return slog.LevelInfo
}

// Message returns a description of the event.
func (e ActionStopped) Message() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %d: %s: Stopped action due to an error: %s.", e.Deployment, e.Flow, e.ActionIndex+1, e.ActionType, e.Err)
	}
	return fmt.Sprintf("%s: %s: %d: %s: Completed action.", e.Deployment, e.Flow, e.ActionIndex+1, e.ActionType)
}

// Attrs returns a set of structured log attributes for the event.
func (e ActionStopped) Attrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
		slog.Group("action", "index", e.ActionIndex, "type", e.ActionType),
		slog.Time("started", e.Started),
		slog.Time("stopped", e.Stopped),
	}
	if e.Err != nil {
		attrs = append(attrs, slog.String("error", e.Err.Error()))
	}
	return attrs
}

// Duration returns the duration of the action.
func (e ActionStopped) Duration() time.Duration {
	return e.Stopped.Sub(e.Started)
}
