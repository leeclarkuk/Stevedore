package compose

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		compose string
		wantErr []string // substrings that must all appear; empty means success
	}{
		{
			name:    "image only",
			compose: "services:\n  web:\n    image: nginx:1.27\n",
		},
		{
			name:    "extension fields allowed",
			compose: "x-stevedore: {}\nservices:\n  web:\n    image: nginx:1.27\n    x-stevedore: {}\n",
		},
		{
			name:    "build without image",
			compose: "services:\n  api:\n    build: .\n",
			wantErr: []string{`service "api" has build but no image`, "build and push it in CI"},
		},
		{
			name:    "build with image is still unsupported",
			compose: "services:\n  api:\n    build: .\n    image: example/api:1\n",
			wantErr: []string{`service "api": key "build" is not supported yet`},
		},
		{
			name:    "missing image",
			compose: "services:\n  web:\n    command: [\"true\"]\n",
			wantErr: []string{`service "web" has no image`, `key "command" is not supported yet`},
		},
		{
			name:    "unsupported service key",
			compose: "services:\n  web:\n    image: nginx:1.27\n    ports: [\"8080:80\"]\n",
			wantErr: []string{`service "web": key "ports" is not supported yet`, "supported: image"},
		},
		{
			name:    "unsupported top-level key",
			compose: "services:\n  web:\n    image: nginx:1.27\nvolumes:\n  data: {}\n",
			wantErr: []string{`top-level key "volumes" is not supported yet`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "compose.yaml")
			if err := os.WriteFile(path, []byte(tt.compose), 0o600); err != nil {
				t.Fatal(err)
			}

			_, err := Load(context.Background(), path)
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
