package lbdeploy

// AppMap holds a set of applications mapped by their identifiers.
//
// It is used to identify relevant applications for a deployment.
type AppMap map[AppID]Application

// AppID is a unique identifier for an application.
type AppID string

// Application hold identifying information for an application.
type Application struct {
	Name         string `json:"name"`
	Architecture string `json:"architecture"`
	Scope        string `json:"scope"`
	ID           string `json:"id"`
}
