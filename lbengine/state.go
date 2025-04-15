package lbengine

import (
	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
	"github.com/leafbridge/leafbridge-deploy/tempfs"
)

// engineState keeps track of the overall state of an flow.
type engineState struct {
	activeFlows       flowSet
	extractedPackages map[lbdeploy.PackageID]tempfs.ExtractionDir
	locks             *lockManager
}

func newEngineState() *engineState {
	return &engineState{
		activeFlows:       make(flowSet),
		extractedPackages: make(map[lbdeploy.PackageID]tempfs.ExtractionDir),
		locks:             newLockManager(),
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
