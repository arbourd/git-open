package open

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/arbourd/git-open/gitw"
)

func TestGetURLWindowsAbsolutePath(t *testing.T) {
	ref, err := gitw.CurrentRef(".")
	if err != nil {
		t.Fatalf("unable to get local ref for test: %v", err)
	}

	abs, err := filepath.Abs("open_test.go")
	if err != nil {
		t.Fatalf("unable to get absolute path: %v", err)
	}

	url, err := GetURL(abs+":3-10", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := fmt.Sprintf("https://github.com/arbourd/git-open/tree/%s/open/open_test.go#L3-L10", ref)
	if url != expectedURL {
		t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, expectedURL)
	}
}

// TestParsePathWindowsThroughFile checks that a path through a file, which fails with
// ERROR_PATH_NOT_FOUND on Windows (ENOTDIR on POSIX), is kept like any other missing path.
func TestParsePathWindowsThroughFile(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolving temp dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "f.txt"), nil, 0o644); err != nil {
		t.Fatalf("writing f.txt: %v", err)
	}
	through := filepath.Join(root, "f.txt", "sub", "x.go")

	cases := map[string]struct {
		path          string
		expectedPath  string
		expectedStart int
	}{
		"absolute drive path through a file": {
			path:         through,
			expectedPath: "f.txt/sub/x.go",
		},
		"absolute drive path through a file with line": {
			path:          through + ":7",
			expectedPath:  "f.txt/sub/x.go",
			expectedStart: 7,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			path, start, _ := parsePath(c.path, root)
			if path != c.expectedPath || start != c.expectedStart {
				t.Fatalf("unexpected result:\n\t(GOT): path=%#v start=%#v\n\t(WNT): path=%#v start=%#v",
					path, start, c.expectedPath, c.expectedStart)
			}
		})
	}
}
