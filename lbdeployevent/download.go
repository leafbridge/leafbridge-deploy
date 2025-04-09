package lbdeployevent

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
)

// DownloadStarted is an event that occurs when a file download has started.
type DownloadStarted struct {
	Deployment lbdeploy.DeploymentID
	Flow       lbdeploy.FlowID
	Action     lbdeploy.ActionType
	Source     lbdeploy.PackageSource
	FileName   string
	Path       string
	Offset     int64
}

// Component identifies the component that generated the event.
func (e DownloadStarted) Component() string {
	return "download"
}

// Level returns the level of the event.
func (e DownloadStarted) Level() slog.Level {
	return slog.LevelInfo
}

// Message returns a description of the event.
func (e DownloadStarted) Message() string {
	if e.Offset > 0 {
		return fmt.Sprintf("Resuming download of \"%s\" from \"%s\" at offset %d.", e.FileName, e.Source.URL, e.Offset)
	}
	return fmt.Sprintf("Starting download of \"%s\" from \"%s\".", e.FileName, e.Source.URL)
}

// Attrs returns a set of structured log attributes for the event.
func (e DownloadStarted) Attrs() []slog.Attr {
	return []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
		slog.String("action", string(e.Action)),
		slog.Group("source", "type", string(e.Source.Type), "url", e.Source.URL),
		slog.String("path", string(e.Path)),
		slog.Int64("offset", e.Offset),
	}
}

// DownloadStopped is an event that occurs when a file download has stopped.
type DownloadStopped struct {
	Deployment lbdeploy.DeploymentID
	Flow       lbdeploy.FlowID
	Action     lbdeploy.ActionType
	Source     lbdeploy.PackageSource
	FileName   string
	Path       string
	Downloaded int64
	FileSize   int64
	Started    time.Time
	Stopped    time.Time
	Err        error
}

// Component identifies the component that generated the event.
func (e DownloadStopped) Component() string {
	return "download"
}

// Level returns the level of the event.
func (e DownloadStopped) Level() slog.Level {
	if e.Err != nil {
		return slog.LevelError
	}
	return slog.LevelInfo
}

// Message returns a description of the event.
func (e DownloadStopped) Message() string {
	duration := e.Duration().Round(time.Millisecond * 10)
	if e.Err != nil {
		if e.Downloaded > 0 {
			return fmt.Sprintf("The download of \"%s\" from \"%s\" failed after receiving %d %s over %s (%s mbps) due to an error: %s.",
				e.FileName,
				e.Source.URL,
				e.Downloaded,
				plural(e.Downloaded, "byte", "bytes"),
				duration,
				e.BitrateInMbps(),
				e.Err)
		}
		return fmt.Sprintf("The download of \"%s\" from \"%s\" failed due to an error: %s.", e.FileName, e.Source.URL, e.Err)
	}
	return fmt.Sprintf("The download of \"%s\" from \"%s\" was completed in %s (%s mbps).", e.FileName, e.Source.URL, duration, e.BitrateInMbps())
}

// Attrs returns a set of structured log attributes for the event.
func (e DownloadStopped) Attrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
		slog.String("action", string(e.Action)),
		slog.Group("source", "type", string(e.Source.Type), "url", e.Source.URL),
		slog.String("path", string(e.Path)),
		slog.Int64("downloaded", e.Downloaded),
		slog.Int64("file-size", e.FileSize),
		slog.Int64("bitrate", e.FileSize),
		slog.Time("started", e.Started),
		slog.Time("stopped", e.Stopped),
	}
	if e.Err != nil {
		attrs = append(attrs, slog.String("error", e.Err.Error()))
	}
	return attrs
}

// Duration returns the duration of the download.
func (e DownloadStopped) Duration() time.Duration {
	return e.Stopped.Sub(e.Started)
}

// BitrateInMbps returns the bitrate of the download in mebibits per second.
func (e DownloadStopped) BitrateInMbps() string {
	return bitrate(e.Downloaded, e.Duration())
}

// DownloadResetReason identifies the reason that a download was reset.
type DownloadResetReason string

// Possible reasons for a download being reset.
const (
	ExistingFileTooLarge             DownloadResetReason = "existing-file-too-large"
	ExistingFileVerificationFailed   DownloadResetReason = "existing-file-verification-failed"
	HTTPServerDoesNotSupportResume   DownloadResetReason = "http-server-does-not-support-resume"
	DownloadedFileVerificationFailed DownloadResetReason = "downloaded-file-verification-failed"
)

// Description returns a string describing the reason that the download was
// reset.
func (reason DownloadResetReason) Description() string {
	switch reason {
	case ExistingFileTooLarge:
		return "the existing file is larger than expected"
	case ExistingFileVerificationFailed:
		return "the existing file did not pass verification"
	case HTTPServerDoesNotSupportResume:
		return "the HTTP server does not support resuming downloads"
	case DownloadedFileVerificationFailed:
		return "the downloaded file did not pass verification"
	default:
		return string(reason)
	}
}

// DownloadReset is an event that occurs when previously downloaded
// content is discarded, forcing the download to start from the beginning
// again.
type DownloadReset struct {
	Deployment lbdeploy.DeploymentID
	Flow       lbdeploy.FlowID
	Action     lbdeploy.ActionType
	Source     lbdeploy.PackageSource
	FileName   string
	Path       string
	Reason     DownloadResetReason
}

// Component identifies the component that generated the event.
func (e DownloadReset) Component() string {
	return "download"
}

// Level returns the level of the event.
func (e DownloadReset) Level() slog.Level {
	if e.Reason == HTTPServerDoesNotSupportResume {
		return slog.LevelWarn
	}
	return slog.LevelError
}

// Message returns a description of the event.
func (e DownloadReset) Message() string {
	if e.Source.URL != "" {
		return fmt.Sprintf("The downloaded content of \"%s\" from \"%s\" was discarded because %s. The file will be redownloaded.",
			e.FileName,
			e.Source.URL,
			e.Reason.Description())
	}

	return fmt.Sprintf("The downloaded content of \"%s\" was discarded because %s. The file will be redownloaded.",
		e.FileName,
		e.Reason.Description())
}

// Attrs returns a set of structured log attributes for the event.
func (e DownloadReset) Attrs() []slog.Attr {
	attrs := []slog.Attr{
		slog.String("deployment", string(e.Deployment)),
		slog.String("flow", string(e.Flow)),
		slog.String("action", string(e.Action)),
		slog.String("path", e.Path),
		slog.String("reason", string(e.Reason)),
	}
	if e.Source.URL != "" {
		attrs = append(attrs, slog.Group("source", "type", string(e.Source.Type), "url", e.Source.URL))
	}
	return attrs
}
