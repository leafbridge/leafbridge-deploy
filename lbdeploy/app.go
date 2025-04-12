package lbdeploy

import (
	"fmt"
	"strings"

	"github.com/gentlemanautomaton/winapp/appcode"
	"github.com/gentlemanautomaton/winapp/unpackaged"
	"github.com/gentlemanautomaton/winapp/unpackaged/appscope"
)

// AppMap holds a set of applications mapped by their identifiers.
//
// It is used to identify relevant applications for a deployment.
type AppMap map[AppID]Application

// AppList is a list of relevant applications for a deployment.
type AppList []AppID

// Difference returns all members of list that are not members of other.
func (list AppList) Difference(other AppList) AppList {
	lookup := make(map[AppID]struct{}, len(other))
	for _, app := range other {
		lookup[app] = struct{}{}
	}
	var diff AppList
	for _, app := range list {
		if _, excepted := lookup[app]; !excepted {
			diff = append(diff, app)
		}
	}
	return diff
}

// String returns a string represenation of the list.
func (list AppList) String() string {
	var out strings.Builder
	for i, item := range list {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(string(item))
	}
	return out.String()
}

// AppID is a unique identifier for an application within LeafBridge.
type AppID string

// Application hold identifying information for an application.
type Application struct {
	Name         string               `json:"name"`
	Architecture appcode.Architecture `json:"architecture"`
	Scope        appscope.Scope       `json:"scope"`
	ID           unpackaged.AppID     `json:"id"`
}

// AppEvaluation is an evaluation of potential changes to the set of installed
// applications.
type AppEvaluation struct {
	AlreadyInstalled   AppList
	AlreadyUninstalled AppList
	ToInstall          AppList
	ToUninstall        AppList
}

// IsZero returns true if the app evaluation is empty.
func (e AppEvaluation) IsZero() bool {
	if len(e.AlreadyInstalled) > 0 {
		return false
	}
	if len(e.AlreadyUninstalled) > 0 {
		return false
	}
	if len(e.ToInstall) > 0 {
		return false
	}
	if len(e.ToUninstall) > 0 {
		return false
	}
	return true
}

// ActionsNeeded returns true if any apps need to be installed or uninstalled.
func (e AppEvaluation) ActionsNeeded() bool {
	if len(e.ToInstall) > 0 {
		return true
	}
	if len(e.ToUninstall) > 0 {
		return true
	}
	return false
}

// AppSummary is a summary of changes to the set of installed applications.
type AppSummary struct {
	Installed           AppList
	Uninstalled         AppList
	StillNotInstalled   AppList
	StillNotUninstalled AppList
}

// IsZero returns true if the app summary is empty.
func (s AppSummary) IsZero() bool {
	if len(s.Installed) > 0 {
		return false
	}
	if len(s.Uninstalled) > 0 {
		return false
	}
	if len(s.StillNotInstalled) > 0 {
		return false
	}
	if len(s.StillNotUninstalled) > 0 {
		return false
	}

	return true
}

// Err returns a non-nil error if any of the expected application changes did
// not take effect.
func (s AppSummary) Err() error {
	switch {
	case len(s.StillNotInstalled) > 0 && len(s.StillNotUninstalled) > 0:
		return fmt.Errorf("some applications were not installed (%s) and some applications were not uninstalled (%s)", s.StillNotInstalled, s.StillNotUninstalled)
	case len(s.StillNotInstalled) > 0:
		return fmt.Errorf("the following applications were not installed properly: %s", s.StillNotInstalled)
	case len(s.StillNotUninstalled) > 0:
		return fmt.Errorf("the following applications were not uninstalled properly: %s", s.StillNotUninstalled)
	default:
		return nil
	}
}
