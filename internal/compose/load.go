// Package compose loads Compose files through compose-go and rejects keys
// that Stevedore cannot yet convert.
package compose

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

// supportedTopLevel and supportedServiceKeys are allowlists, not denylists, so
// that a Compose key we have not thought about fails instead of being dropped.
var supportedTopLevel = map[string]bool{
	"name":     true,
	"services": true,
}

var supportedServiceKeys = map[string]bool{
	"image": true,
}

// Load reads the Compose file at path, resolves it exactly as Docker Compose
// would, and returns an error for any key Stevedore cannot convert.
func Load(ctx context.Context, path string) (*types.Project, error) {
	opts, err := cli.NewProjectOptions(
		[]string{path},
		cli.WithOsEnv,
		cli.WithDotEnv,
	)
	if err != nil {
		return nil, fmt.Errorf("compose options for %q: %w", path, err)
	}

	// The raw model is checked before normalisation, because normalisation
	// adds keys (such as the default network) that the user never wrote.
	if err := cli.WithNormalization(false)(opts); err != nil {
		return nil, fmt.Errorf("compose options for %q: %w", path, err)
	}
	raw, err := opts.LoadModel(ctx)
	if err != nil {
		return nil, fmt.Errorf("load compose model %q: %w", path, err)
	}
	if err := validate(raw); err != nil {
		return nil, fmt.Errorf("validate %q: %w", path, err)
	}

	if err := cli.WithNormalization(true)(opts); err != nil {
		return nil, fmt.Errorf("compose options for %q: %w", path, err)
	}
	// Docker lets COMPOSE_PROJECT_NAME override name: from the file. The name
	// feeds the instance label (ADR 0002), so pin it to the file's value to
	// keep output independent of the environment.
	if err := cli.WithName(raw["name"].(string))(opts); err != nil {
		return nil, fmt.Errorf("project name in %q: %w", path, err)
	}
	project, err := opts.LoadProject(ctx)
	if err != nil {
		return nil, fmt.Errorf("load compose project %q: %w", path, err)
	}
	return project, nil
}

func validate(raw map[string]any) error {
	var errs []error

	for _, key := range sortedKeys(raw) {
		if isExtension(key) || supportedTopLevel[key] {
			continue
		}
		errs = append(errs, fmt.Errorf("top-level key %q is not supported yet; remove it (supported: %s)",
			key, list(supportedTopLevel)))
	}

	// Without name:, compose-go derives the project name from the directory,
	// so the same file would render differently depending on where it lives.
	if name, _ := raw["name"].(string); name == "" {
		errs = append(errs, errors.New("no top-level name; add name: <project> so output does not depend on the directory or environment"))
	}

	services, _ := raw["services"].(map[string]any)
	if len(services) == 0 {
		errs = append(errs, errors.New("no services defined; add at least one service with an image"))
	}
	for _, name := range sortedKeys(services) {
		svc, _ := services[name].(map[string]any)
		errs = append(errs, validateService(name, svc)...)
	}

	return errors.Join(errs...)
}

func validateService(name string, svc map[string]any) []error {
	var errs []error

	_, hasImage := svc["image"]
	if _, hasBuild := svc["build"]; hasBuild && !hasImage {
		errs = append(errs, fmt.Errorf("service %q has build but no image; Stevedore does not build images, so build and push it in CI, then set image", name))
	} else if !hasImage {
		errs = append(errs, fmt.Errorf("service %q has no image; add the image reference the cluster should pull", name))
	}

	for _, key := range sortedKeys(svc) {
		if isExtension(key) || supportedServiceKeys[key] {
			continue
		}
		if key == "build" && !hasImage {
			continue // already reported above with a more specific fix
		}
		errs = append(errs, fmt.Errorf("service %q: key %q is not supported yet; remove it (supported: %s)",
			name, key, list(supportedServiceKeys)))
	}
	return errs
}

func isExtension(key string) bool { return strings.HasPrefix(key, "x-") }

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func list(m map[string]bool) string {
	return strings.Join(sortedKeys(m), ", ")
}
