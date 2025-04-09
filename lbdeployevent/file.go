package lbdeployevent

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
)

// FileVerification is an event that records the result of verifying
// a downloaded file.
type FileVerification struct {
	Deployment lbdeploy.DeploymentID
	Flow       lbdeploy.FlowID
	Action     lbdeploy.ActionType
	Source     lbdeploy.PackageSource
	FileName   string
	Path       string
	Expected   lbdeploy.FileAttributes
	Actual     lbdeploy.FileAttributes
}

// Component identifies the component that generated the event.
func (e FileVerification) Component() string {
	return "verification"
}

// Level returns the level of the event.
func (e FileVerification) Level() slog.Level {
	if len(e.Expected.Features()) == 0 {
		return slog.LevelWarn
	}
	if !lbdeploy.EqualFileAttributes(e.Expected, e.Actual) {
		return slog.LevelError
	}
	if len(e.Expected.Hashes) == 0 {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}

// Message returns a description of the event.
func (e FileVerification) Message() string {
	if len(e.Expected.Features()) == 0 {
		return fmt.Sprintf("The \"%s\" file could not be verified because file verification data was not provided.", e.FileName)
	}
	if !lbdeploy.EqualFileAttributes(e.Expected, e.Actual) {
		return fmt.Sprintf("The \"%s\" file does not have the expected file attributes and has failed verification.", e.FileName)
	}
	if len(e.Expected.Hashes) == 0 {
		return fmt.Sprintf("The \"%s\" file has the expected file size, but no file hashes were provided for verification.", e.FileName)
	}
	return fmt.Sprintf("The \"%s\" file was verified with the following features: %s.", e.FileName, strings.Join(e.Actual.Features(), ", "))
}

// Attrs returns a set of structured log attributes for the event.
func (e FileVerification) Attrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
		slog.String("action", string(e.Action)),
	}
	if e.Source.URL != "" {
		attrs = append(attrs, slog.Group("source", "type", string(e.Source.Type), "url", e.Source.URL))
	}
	if e.Path != "" {
		attrs = append(attrs, slog.String("path", string(e.Path)))
	}
	attrs = append(attrs, slog.Group("expected", "size", e.Expected.Size, "hashes", e.Expected.Hashes))
	attrs = append(attrs, slog.Group("actual", "size", e.Actual.Size, "hashes", e.Actual.Hashes))
	return attrs
}
