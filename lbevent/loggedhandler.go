package lbevent

import (
	"context"
	"log/slog"
)

// LoggedHandler is a LeafBridge event handler that sends events to
// a structured log handler.
type LoggedHandler struct {
	Handler slog.Handler
}

// Handle processes the given event record.
func (lh LoggedHandler) Handle(r Record) {
	h := lh.Handler
	if lh.Handler == nil {
		h = slog.Default().Handler()
	}
	h.Handle(context.Background(), r.ToLog())
}
