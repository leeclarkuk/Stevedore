// Package model is Stevedore's intermediate representation, sitting between
// compose-go's types and Kubernetes objects so that renderers never depend on
// Compose directly.
package model

import (
	"sort"

	"github.com/compose-spec/compose-go/v2/types"
)

// App is one Compose project.
type App struct {
	Name     string
	Services []Service // sorted by Name, so rendering order is stable
}

// Service is one Compose service.
type Service struct {
	Name  string
	Image string
}

// FromProject builds an App from a loaded and validated Compose project.
func FromProject(p *types.Project) App {
	app := App{Name: p.Name}
	for name, svc := range p.Services {
		app.Services = append(app.Services, Service{Name: name, Image: svc.Image})
	}
	sort.Slice(app.Services, func(i, j int) bool {
		return app.Services[i].Name < app.Services[j].Name
	})
	return app
}
