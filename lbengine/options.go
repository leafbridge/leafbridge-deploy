package lbengine

import "github.com/leafbridge/leafbridge-deploy/lbevent"

// Options hold configuration options for a LeafBridge deployment engine.
type Options struct {
	Events lbevent.Recorder
}
