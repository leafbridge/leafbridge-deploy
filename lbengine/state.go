package lbengine

import "github.com/leafbridge/leafbridge-deploy/lbdeploy"

// engineState keeps track of the overall state of an flow.
type engineState struct {
	activeFlows flowSet
}

func newEngineState() *engineState {
	return &engineState{
		activeFlows: make(flowSet),
	}
}

// flowSet keeps track of a set of flows.
type flowSet map[lbdeploy.FlowID]struct{}

func (fs flowSet) Contains(flow lbdeploy.FlowID) bool {
	_, present := fs[flow]
	return present
}

func (fs flowSet) Add(flow lbdeploy.FlowID) {
	fs[flow] = struct{}{}
}

func (fs flowSet) Remove(flow lbdeploy.FlowID) {
	delete(fs, flow)
}
