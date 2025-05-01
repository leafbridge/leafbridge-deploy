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

// Name returns a name for the handler.
func (h WindowsHandler) Name() string {
	return "windows-application-log"
}

// Handle processes the given event record.
func (h WindowsHandler) Handle(r Record) (err error) {
	// Log the event according to the event level.
	switch level := r.Level(); {
	case level >= slog.LevelError:
		err = h.elog.Error(300, eventMessageWithDetails(r))
	case level >= slog.LevelWarn:
		err = h.elog.Warning(200, eventMessageWithDetails(r))
	case level >= slog.LevelInfo:
		err = h.elog.Info(100, eventMessageWithDetails(r))
	default:
		return nil // Drop debug messages.
	}

	// If we failed to log the event, try again without the message details.
	if err != nil {
		switch level := r.Level(); {
		case level >= slog.LevelError:
			h.elog.Error(300, r.Message())
		case level >= slog.LevelWarn:
			h.elog.Warning(200, r.Message())
		case level >= slog.LevelInfo:
			h.elog.Info(100, r.Message())
		}
	}

	return err
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

func eventMessageWithDetails(r Record) string {
	message := r.Message()
	if details := r.Details(); details != "" {
		return fmt.Sprintf("%s\n\n%s", message, details)
	}
	return message
}
