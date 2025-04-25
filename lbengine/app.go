package lbengine

import (
	"fmt"

	"github.com/gentlemanautomaton/winapp/unpackaged/appregistry"
	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
)

// AppEngine is responsible for evaluating the status of applications on the
// local system.
type AppEngine struct {
	deployment lbdeploy.Deployment
}

// NewAppEngine prepares an app engine for the given deployment.
func NewAppEngine(dep lbdeploy.Deployment) AppEngine {
	return AppEngine{
		deployment: dep,
	}
}

// IsInstalled returns true if the application is installed on the local
// system.
//
// If it is unable to make a determination, it returns an error.
func (engine AppEngine) IsInstalled(app lbdeploy.AppID) (bool, error) {
	// Find the app within the deployment.
	definition, found := engine.deployment.Apps[app]
	if !found {
		return false, fmt.Errorf("the \"%s\" app does not exist within the \"%s\" deployment", app, engine.deployment.ID)
	}

	// If a condition has been supplied, use that to determine the
	// application's status.
	if definition.Condition != "" {
		ce := NewConditionEngine(engine.deployment)
		return ce.Evaluate(definition.Condition)
	}

	// Use the application registry that matches the application's
	// architecture (x64 or x86) and scope (machine or user).
	view, err := appregistry.ViewFor(definition.Architecture, definition.Scope)
	if err != nil {
		return false, err
	}

	// Look for the application in the registry.
	return view.Contains(definition.ProductCode)
}

// InstalledApps returns any of the apps in the list that are installed on the
// local system.
func (engine AppEngine) InstalledApps(list lbdeploy.AppList) (installed lbdeploy.AppList, err error) {
	for _, appID := range list {
		appIsInstalled, err := engine.IsInstalled(appID)
		if err != nil {
			return nil, fmt.Errorf("unable to determine the installation state of application \"%s\": %w", appID, err)
		}
		if appIsInstalled {
			installed = append(installed, appID)
		}
	}
	return
}

// MissingApps returns any of the apps in the list that are not installed on the
// local system.
func (engine AppEngine) MissingApps(list lbdeploy.AppList) (missing lbdeploy.AppList, err error) {
	for _, appID := range list {
		appIsInstalled, err := engine.IsInstalled(appID)
		if err != nil {
			return nil, fmt.Errorf("unable to determine the installation state of application \"%s\": %w", appID, err)
		}
		if !appIsInstalled {
			missing = append(missing, appID)
		}
	}
	return
}

// EvaluateAppChanges evaluates the changes needed to effect the given set of
// application installs and uninstalls.
func (engine AppEngine) EvaluateAppChanges(installs, uninstalls lbdeploy.AppList) (changes lbdeploy.AppEvaluation, err error) {
	alreadyInstalled, err := engine.InstalledApps(installs)
	if err != nil {
		return changes, err
	}
	toInstall := installs.Difference(alreadyInstalled)

	alreadyUninstalled, err := engine.MissingApps(uninstalls)
	if err != nil {
		return changes, err
	}
	toUninstall := uninstalls.Difference(alreadyUninstalled)

	return lbdeploy.AppEvaluation{
		AlreadyInstalled:   alreadyInstalled,
		AlreadyUninstalled: alreadyUninstalled,
		ToInstall:          toInstall,
		ToUninstall:        toUninstall,
	}, nil
}

// SummarizeAppChanges summarizes the effectiveness of application installs
// and uninstalls anticipated by a previous evaluation.
func (engine AppEngine) SummarizeAppChanges(evaluation lbdeploy.AppEvaluation) (changes lbdeploy.AppSummary, err error) {
	stillNotInstalled, err := engine.MissingApps(evaluation.ToInstall)
	if err != nil {
		return changes, err
	}
	installed := evaluation.ToInstall.Difference(stillNotInstalled)

	stillNotUninstalled, err := engine.InstalledApps(evaluation.ToUninstall)
	if err != nil {
		return changes, err
	}
	uninstalled := evaluation.ToUninstall.Difference(stillNotUninstalled)

	return lbdeploy.AppSummary{
		Installed:           installed,
		Uninstalled:         uninstalled,
		StillNotInstalled:   stillNotInstalled,
		StillNotUninstalled: stillNotUninstalled,
	}, nil
}
