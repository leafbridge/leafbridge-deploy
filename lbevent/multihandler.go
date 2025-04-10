package lbevent

import "errors"

// MultiHandler is a LeafBridge event handler that sends events to multiple
// underlying handlers.
type MultiHandler []Handler

// Handle processes the given event record.
func (h MultiHandler) Handle(r Record) error {
	var errs []error
	for _, handler := range h {
		if err := handler.Handle(r); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
