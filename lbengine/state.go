package lbengine

import (
	"github.com/leafbridge/leafbridge-deploy/idset"
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
type flowSet = idset.SetOf[lbdeploy.FlowID]
