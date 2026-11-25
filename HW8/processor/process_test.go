package processor

import (
	"path/filepath"
	"testing"
)

func TestProcessFilesCases(t *testing.T) {
	path := filepath.Join("..", "testdata", "orders-a.txt")
	missing := filepath.Join(t.TempDir(), "missing.txt")
	for _, tt := range []struct {
		name    string
		paths   []string
		workers int
	}{
		{"empty", nil, 3},
		{"zero workers", []string{path}, 0},
		{"negative workers", []string{path}, -2},
		{"duplicates", []string{path, path}, 3},
		{"missing file", []string{missing, path}, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			remaining := make(map[string]int)
			for _, p := range tt.paths {
				remaining[p]++
			}
			for got := range ProcessFiles(tt.paths, tt.workers) {
				if remaining[got.Path] == 0 {
					t.Fatalf("unexpected result: %+v", got)
				}
				remaining[got.Path]--
				if got.Path == missing {
					if got.Err == nil {
						t.Fatal("expected an error for the missing file")
					}
				} else if got.Err != nil || got.Lines != 3 || got.Words != 5 {
					t.Fatalf("unexpected file statistics: %+v", got)
				}
			}
			for p, count := range remaining {
				if count != 0 {
					t.Errorf("missing %d results for %s", count, p)
				}
			}
		})
	}
}
