package convert

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

var update = flag.Bool("update", false, "regenerate testdata/golden from examples/ (use make golden)")

const (
	examplesDir = "../../examples"
	goldenDir   = "../../testdata/golden"
)

// TestGolden converts every app in examples/ and compares the result with
// testdata/golden/<app>/. Run make golden to accept intended changes.
func TestGolden(t *testing.T) {
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		t.Fatalf("read examples: %v", err)
	}

	found := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		found++
		name := e.Name()
		t.Run(name, func(t *testing.T) {
			files, err := Convert(context.Background(), filepath.Join(examplesDir, name, "compose.yaml"))
			if err != nil {
				t.Fatalf("Convert: %v", err)
			}
			want := filepath.Join(goldenDir, name)

			if *update {
				if err := os.RemoveAll(want); err != nil {
					t.Fatal(err)
				}
				if err := Write(want, files); err != nil {
					t.Fatal(err)
				}
				return
			}

			// Compare the file set too, so a resource that stops being
			// rendered fails the test instead of leaving a stale golden file.
			var gotNames []string
			for _, f := range files {
				gotNames = append(gotNames, f.Path)
				golden, err := os.ReadFile(filepath.Join(want, f.Path))
				if err != nil {
					t.Errorf("%s: missing golden file (run make golden): %v", f.Path, err)
					continue
				}
				if !bytes.Equal(f.Content, golden) {
					t.Errorf("%s differs from golden file (run make golden and review the diff)\n--- got\n%s\n--- want\n%s",
						f.Path, f.Content, golden)
				}
			}

			goldenEntries, err := os.ReadDir(want)
			if err != nil {
				t.Fatalf("read golden dir: %v", err)
			}
			var wantNames []string
			for _, g := range goldenEntries {
				wantNames = append(wantNames, g.Name())
			}
			sort.Strings(gotNames)
			sort.Strings(wantNames)
			if !equal(gotNames, wantNames) {
				t.Errorf("rendered files %v, golden files %v (run make golden)", gotNames, wantNames)
			}
		})
	}
	if found == 0 {
		t.Fatal("no examples found")
	}
}

// TestDeterministic guards the "same input, same bytes" principle directly,
// since a golden test alone can pass by luck when map order happens to agree.
func TestDeterministic(t *testing.T) {
	path := filepath.Join(examplesDir, "web-redis", "compose.yaml")
	first, err := Convert(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	for i := range 20 {
		again, err := Convert(context.Background(), path)
		if err != nil {
			t.Fatal(err)
		}
		if len(again) != len(first) {
			t.Fatalf("run %d: %d files, want %d", i, len(again), len(first))
		}
		for j := range first {
			if first[j].Path != again[j].Path || !bytes.Equal(first[j].Content, again[j].Content) {
				t.Fatalf("run %d: output differs at %s", i, first[j].Path)
			}
		}
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
