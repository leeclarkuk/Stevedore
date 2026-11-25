package compose

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// named prefixes a Compose body with name:, which every valid file needs.
func named(body string) string { return "name: demo\n" + body }

func writeCompose(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "compose.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		compose string
		wantErr []string // substrings that must all appear; empty means success
	}{
		{
			name:    "image only",
			compose: named("services:\n  web:\n    image: nginx:1.27\n"),
		},
		{
			name:    "extension fields allowed",
			compose: named("x-stevedore: {}\nservices:\n  web:\n    image: nginx:1.27\n    x-stevedore: {}\n"),
		},
		{
			name:    "missing name",
			compose: "services:\n  web:\n    image: nginx:1.27\n",
			wantErr: []string{"no top-level name", "add name: <project>"},
		},
		{
			name:    "empty name",
			compose: "name: \"\"\nservices:\n  web:\n    image: nginx:1.27\n",
			wantErr: []string{"no top-level name"},
		},
		{
			name:    "build without image",
			compose: named("services:\n  api:\n    build: .\n"),
			wantErr: []string{`service "api" has build but no image`, "build and push it in CI"},
		},
		{
			name:    "build with image is still unsupported",
			compose: named("services:\n  api:\n    build: .\n    image: example/api:1\n"),
			wantErr: []string{`service "api": key "build" is not supported yet`},
		},
		{
			name:    "missing image",
			compose: named("services:\n  web:\n    command: [\"true\"]\n"),
			wantErr: []string{`service "web" has no image`, `key "command" is not supported yet`},
		},
		{
			name:    "unsupported service key",
			compose: named("services:\n  web:\n    image: nginx:1.27\n    ports: [\"8080:80\"]\n"),
			wantErr: []string{`service "web": key "ports" is not supported yet`, "supported: image"},
		},
		{
			name:    "unsupported top-level key",
			compose: named("services:\n  web:\n    image: nginx:1.27\nvolumes:\n  data: {}\n"),
			wantErr: []string{`top-level key "volumes" is not supported yet`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(context.Background(), writeCompose(t, tt.compose))
			if len(tt.wantErr) == 0 {
				if err != nil {
					t.Fatalf("Load() unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Load() succeeded, want error containing %q", tt.wantErr)
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("Load() error = %q, want it to contain %q", err, want)
				}
			}
		})
	}
}

// Docker Compose lets COMPOSE_PROJECT_NAME override name:, which would make
// output depend on the environment. Stevedore must use the file's name.
func TestLoadIgnoresProjectNameEnv(t *testing.T) {
	t.Setenv("COMPOSE_PROJECT_NAME", "fromenv")

	project, err := Load(context.Background(), writeCompose(t, named("services:\n  web:\n    image: nginx:1.27\n")))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if project.Name != "demo" {
		t.Errorf("project name = %q, want %q from the file", project.Name, "demo")
	}
}
