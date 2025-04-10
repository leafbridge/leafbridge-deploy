package lbdeployevent

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
)

// CommandStarted is an event that occurs when a command has started.
type CommandStarted struct {
	Deployment  lbdeploy.DeploymentID
	Flow        lbdeploy.FlowID
	ActionIndex int
	ActionType  lbdeploy.ActionType
	Package     lbdeploy.PackageID
	Command     lbdeploy.PackageCommandID
	CommandLine string
}

// Component identifies the component that generated the event.
func (e CommandStarted) Component() string {
	return "command"
}

// Level returns the level of the event.
func (e CommandStarted) Level() slog.Level {
	return slog.LevelInfo
}

// Message returns a description of the event.
func (e CommandStarted) Message() string {
	return fmt.Sprintf("%s: %s: %d: Invoke %s.%s: Starting command: %s",
		e.Deployment,
		e.Flow,
		e.ActionIndex+1,
		e.Package,
		e.Command,
		e.CommandLine)
}

// Attrs returns a set of structured log attributes for the event.
func (e CommandStarted) Attrs() []slog.Attr {
	return []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
		slog.Group("action", "index", e.ActionIndex, "type", e.ActionType),
		slog.String("package", string(e.Package)),
		slog.Group("command", "id", e.Command, "invocation", e.CommandLine),
	}
}

// CommandStopped is an event that occurs when a command has stopped.
type CommandStopped struct {
	Deployment  lbdeploy.DeploymentID
	Flow        lbdeploy.FlowID
	ActionIndex int
	ActionType  lbdeploy.ActionType
	Package     lbdeploy.PackageID
	Command     lbdeploy.PackageCommandID
	CommandLine string
	Started     time.Time
	Stopped     time.Time
	Err         error
}

// Component identifies the component that generated the event.
func (e CommandStopped) Component() string {
	return "command"
}

// Level returns the level of the event.
func (e CommandStopped) Level() slog.Level {
	if e.Err != nil {
		return slog.LevelError
	}
	return slog.LevelInfo
}

// Message returns a description of the event.
func (e CommandStopped) Message() string {
	duration := e.Duration().Round(time.Millisecond * 10)
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %d: Invoke %s.%s: Stopped command due to an error: %s. (%s)",
			e.Deployment,
			e.Flow,
			e.ActionIndex+1,
			e.Package,
			e.Command,
			e.Err,
			duration)
	}
	return fmt.Sprintf("%s: %s: %d: Invoke %s.%s: Completed command. (%s)",
		e.Deployment,
		e.Flow,
		e.ActionIndex+1,
		e.Package,
		e.Command,
		duration)
}

// Attrs returns a set of structured log attributes for the event.
func (e CommandStopped) Attrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
		slog.Group("action", "index", e.ActionIndex, "type", e.ActionType),
		slog.String("package", string(e.Package)),
		slog.Group("command", "id", e.Command, "invocation", e.CommandLine),
		slog.Time("started", e.Started),
		slog.Time("stopped", e.Stopped),
	}
	if e.Err != nil {
		attrs = append(attrs, slog.String("error", e.Err.Error()))
	}
	return attrs
}

// Duration returns the duration of the action.
func (e CommandStopped) Duration() time.Duration {
	return e.Stopped.Sub(e.Started)
}
