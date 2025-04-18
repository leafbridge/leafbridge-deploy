package lbdeployevent

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gentlemanautomaton/structformat"
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
	var builder structformat.Builder

	builder.WritePrimary(string(e.Deployment))
	builder.WritePrimary(string(e.Flow))
	builder.WriteStandard(fmt.Sprintf("Starting invocation"))

	return builder.String()
}

// Details returns additional details about the event. It might include
// multiple lines of text. An empty string is returned when no details
// are available.
func (e FlowStarted) Details() string {
	return ""
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
	var builder structformat.Builder

	builder.WritePrimary(string(e.Deployment))
	builder.WritePrimary(string(e.Flow))
	if e.Err != nil {
		builder.WriteStandard(fmt.Sprintf("Stopped invocation due to an error: %s.", e.Err))
	} else {
		builder.WriteStandard("Completed invocation.")
	}
	builder.WriteNote(e.Duration().Round(time.Millisecond * 10).String())

	return builder.String()
}

// Details returns additional details about the event. It might include
// multiple lines of text. An empty string is returned when no details
// are available.
func (e FlowStopped) Details() string {
	return ""
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

// FlowCondition is an event that occurs when a deployment flow evalutes
// its preconditions.
type FlowCondition struct {
	Deployment lbdeploy.DeploymentID
	Flow       lbdeploy.FlowID
	Passed     lbdeploy.ConditionList
	Failed     lbdeploy.ConditionList
	Err        error
}

// Component identifies the component that generated the event.
func (e FlowCondition) Component() string {
	return "flow"
}

// Level returns the level of the event.
func (e FlowCondition) Level() slog.Level {
	if e.Err != nil || len(e.Failed) > 0 {
		return slog.LevelError
	}
	return slog.LevelDebug
}

// Message returns a description of the event.
func (e FlowCondition) Message() string {
	var builder structformat.Builder

	builder.WritePrimary(string(e.Deployment))
	builder.WritePrimary(string(e.Flow))
	if e.Err != nil {
		builder.WriteStandard(fmt.Sprintf("Unable to evaluate preconditions: %s", e.Err))
	} else if len(e.Failed) > 0 {
		builder.WriteStandard(fmt.Sprintf("One or more preconditions failed: %s.", e.Failed))
	} else {
		builder.WriteStandard(fmt.Sprintf("All preconditions passed: %s.", e.Passed))
	}

	return builder.String()
}

// Details returns additional details about the event. It might include
// multiple lines of text. An empty string is returned when no details
// are available.
func (e FlowCondition) Details() string {
	return ""
}

// Attrs returns a set of structured log attributes for the event.
func (e FlowCondition) Attrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
		slog.Group("conditions", "passed", e.Passed, "failed", e.Failed),
	}
	if e.Err != nil {
		attrs = append(attrs, slog.String("error", e.Err.Error()))
	}
	return attrs
}

// FlowLockNotAcquired is an event that occurs when a deployment flow cannot
// be started because one of its locks could not be acquired.
type FlowLockNotAcquired struct {
	Deployment lbdeploy.DeploymentID
	Flow       lbdeploy.FlowID
	Lock       lbdeploy.LockID
	Err        error
}

// Component identifies the component that generated the event.
func (e FlowLockNotAcquired) Component() string {
	return "flow"
}

// Level returns the level of the event.
func (e FlowLockNotAcquired) Level() slog.Level {
	return slog.LevelError
}

// Message returns a description of the event.
func (e FlowLockNotAcquired) Message() string {
	var builder structformat.Builder

	builder.WritePrimary(string(e.Deployment))
	builder.WritePrimary(string(e.Flow))
	if e.Err != nil {
		builder.WriteStandard(fmt.Sprintf("Unable to start the flow: %s", e.Err))
	} else {
		builder.WriteStandard(fmt.Sprintf("Unable to start the flow: The %s lock could not be acquired.", e.Lock))
	}

	return builder.String()
}

// Details returns additional details about the event. It might include
// multiple lines of text. An empty string is returned when no details
// are available.
func (e FlowLockNotAcquired) Details() string {
	return ""
}

// Attrs returns a set of structured log attributes for the event.
func (e FlowLockNotAcquired) Attrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
	}
	if e.Lock != "" {
		attrs = append(attrs, slog.String("lock", string(e.Lock)))
	}
	if e.Err != nil {
		attrs = append(attrs, slog.String("error", e.Err.Error()))
	}
	return attrs
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
	var builder structformat.Builder

	builder.WritePrimary(string(e.Deployment))
	builder.WritePrimary(string(e.Flow))
	builder.WriteStandard("Unable to start the flow. Another instance is already running. Is there a cycle in the flow logic?")

	return builder.String()
}

// Details returns additional details about the event. It might include
// multiple lines of text. An empty string is returned when no details
// are available.
func (e FlowAlreadyRunning) Details() string {
	return ""
}

// Attrs returns a set of structured log attributes for the event.
func (e FlowAlreadyRunning) Attrs() []slog.Attr {
	return []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
	}
}
