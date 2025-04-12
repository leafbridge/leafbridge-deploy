package lbengine

import (
	"fmt"

	"github.com/gentlemanautomaton/winapp/unpackaged/appregistry"
	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
)

// IsAppInstalled returns true if the application is installed on the local
// system.
//
// If it is unable to make a determination, it returns a non-nil error.
func IsAppInstalled(appID lbdeploy.AppID, app lbdeploy.Application) (bool, error) {
	// Use the application registry that matches the application's
	// architecture (x64 or x86) and scope (machine or user).
	view, err := appregistry.ViewFor(app.Architecture, app.Scope)
	if err != nil {
		return false, err
	}

	// Look for the application in the registry.
	return view.Contains(app.ID)
}

// InstalledApps returns any of the apps in the list that are installed on the
// local system.
func InstalledApps(apps lbdeploy.AppMap, list lbdeploy.AppList) (installed lbdeploy.AppList, err error) {
	// If the command installs one or more applications, check whether
	// they are already installed.
	for _, appID := range list {
		if len(apps) == 0 {
			return nil, fmt.Errorf("the application \"%s\" is not defined in the deployment", appID)
		}
		appData, found := apps[appID]
		if !found {
			return nil, fmt.Errorf("the application \"%s\" is not defined in the deployment", appID)
		}
		appIsInstalled, err := IsAppInstalled(appID, appData)
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
func MissingApps(apps lbdeploy.AppMap, list lbdeploy.AppList) (missing lbdeploy.AppList, err error) {
	for _, appID := range list {
		if len(apps) == 0 {
			return nil, fmt.Errorf("the application \"%s\" is not defined in the deployment", appID)
		}
		appData, found := apps[appID]
		if !found {
			return nil, fmt.Errorf("the application \"%s\" is not defined in the deployment", appID)
		}
		appIsInstalled, err := IsAppInstalled(appID, appData)
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
func EvaluateAppChanges(apps lbdeploy.AppMap, installs, uninstalls lbdeploy.AppList) (changes lbdeploy.AppEvaluation, err error) {
	alreadyInstalled, err := InstalledApps(apps, installs)
	if err != nil {
		return changes, err
	}
	toInstall := installs.Difference(alreadyInstalled)

	alreadyUninstalled, err := MissingApps(apps, uninstalls)
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
func SummarizeAppChanges(apps lbdeploy.AppMap, evaluation lbdeploy.AppEvaluation) (changes lbdeploy.AppSummary, err error) {
	stillNotInstalled, err := MissingApps(apps, evaluation.ToInstall)
	if err != nil {
		return changes, err
	}
	installed := evaluation.ToInstall.Difference(stillNotInstalled)

	stillNotUninstalled, err := InstalledApps(apps, evaluation.ToUninstall)
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
