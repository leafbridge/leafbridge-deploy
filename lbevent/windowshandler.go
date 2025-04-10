package lbevent

import (
	"fmt"
	"log/slog"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc/eventlog"
)

const lbEventSource = "LeafBridge"

// WindowsHandler is a LeafBridge event handler that sends events to the
// Windows event log.
type WindowsHandler struct {
	elog *eventlog.Log
}

// NewWindowsHandler returns a WindowsHandler that sends events to the
// Windows event log.
func NewWindowsHandler() (WindowsHandler, error) {
	// Register the event source if it isn't already registered.
	alreadyRegisterd, err := IsWindowsEventSourceRegistered(lbEventSource)
	if err != nil {
		return WindowsHandler{}, err
	}

	if !alreadyRegisterd {
		const eventTypes = eventlog.Error | eventlog.Warning | eventlog.Info
		if err := eventlog.InstallAsEventCreate(lbEventSource, eventTypes); err != nil {
			// Report the error but press on regardless
			return WindowsHandler{}, fmt.Errorf("failed to register event log source for \"%s\": %w", lbEventSource, err)
		}
	}

	// Open the event source.
	elog, err := eventlog.Open(lbEventSource)
	if err != nil {
		return WindowsHandler{}, fmt.Errorf("failed to open event log source for \"%s\": %w", lbEventSource, err)
	}
	return WindowsHandler{elog: elog}, nil
}

// Handle processes the given event record.
func (h WindowsHandler) Handle(r Record) error {
	switch level := r.Level(); {
	case level >= slog.LevelError:
		return h.elog.Error(300, r.Message())
	case level >= slog.LevelWarn:
		return h.elog.Warning(200, r.Message())
	case level >= slog.LevelInfo:
		return h.elog.Info(100, r.Message())
	default:
		return nil // Drop debug messages.
	}
}

// Close releases any resources consumed by the Windows event handler.
func (handler WindowsHandler) Close() error {
	return handler.elog.Close()

	// Remove the event source if needed
	// eventlog.Remove(eventSource)
}

// IsWindowsEventSourceRegistered checks to see whether an event log with the
// given source name has been registered.
func IsWindowsEventSourceRegistered(eventSource string) (bool, error) {
	const addKeyName = `SYSTEM\CurrentControlSet\Services\EventLog\Application`

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, addKeyName+`\`+eventSource, registry.QUERY_VALUE)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	defer key.Close()

	return true, nil
}
