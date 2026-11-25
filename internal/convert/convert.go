// Package convert wires Compose loading, the model and rendering together so
// the CLI and the golden tests run exactly the same pipeline.
package convert

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/leeclarkuk/stevedore/internal/compose"
	"github.com/leeclarkuk/stevedore/internal/model"
	"github.com/leeclarkuk/stevedore/internal/render"
)

// Convert loads the Compose file at composePath and returns the rendered
// Kustomize base.
func Convert(ctx context.Context, composePath string) ([]render.File, error) {
	project, err := compose.Load(ctx, composePath)
	if err != nil {
		return nil, err
	}
	files, err := render.Render(model.FromProject(project))
	if err != nil {
		return nil, fmt.Errorf("render %q: %w", composePath, err)
	}
	return files, nil
}

// Write writes files into dir, creating it if needed. Existing files with
// other names are left alone.
func Write(dir string, files []render.File) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", dir, err)
	}
	for _, f := range files {
		path := filepath.Join(dir, f.Path)
		if err := os.WriteFile(path, f.Content, 0o644); err != nil {
			return fmt.Errorf("write %q: %w", path, err)
		}
	}
	return nil
}
