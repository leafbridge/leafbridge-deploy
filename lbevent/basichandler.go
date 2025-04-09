package lbevent

import (
	"fmt"
	"io"
)

const timestampFormat = "2006-01-02 15:04:05"

// BasicHandler is a LeafBridge event handler that prints timestamped event
// messages to an io.Writer.
type BasicHandler struct {
	w io.Writer
}

// NewBasicHandler returns a BasicHandler that will write to w.
func NewBasicHandler(w io.Writer) BasicHandler {
	return BasicHandler{w: w}
}

// Handle processes the given event record.
func (h BasicHandler) Handle(r Record) {
	fmt.Fprintf(h.w, "%s: %-6s %s\n", r.Time().Local().Format(timestampFormat), r.Level().String()+":", r.Message())
}
