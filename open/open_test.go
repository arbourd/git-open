package open

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arbourd/git-open/gitw"
)

func TestGetURL(t *testing.T) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic("cannot get home directory")
	}

	cases := map[string]struct {
		gitdir      string
		arg         string
		expectedURL string
		err         string
	}{
		"no argument": {
			arg:         "",
			expectedURL: "https://github.com/arbourd/git-open",
		},
		"current directory": {
			arg:         ".",
			expectedURL: "https://github.com/arbourd/git-open/tree/%s/open",
		},
		"up one directory": {
			arg:         filepath.FromSlash("../"),
			expectedURL: "https://github.com/arbourd/git-open/tree/%s",
		},
		"file": {
			arg:         "open_test.go",
			expectedURL: "https://github.com/arbourd/git-open/tree/%s/open/open_test.go",
		},
		"commit sha": {
			arg:         "7605d91",
			expectedURL: "https://github.com/arbourd/git-open/commit/7605d91",
		},
		"commit sha with extension": {
			arg:         "7605d91.txt",
			expectedURL: "https://github.com/arbourd/git-open/tree/%s",
		},
		"commit sha as a folder": {
			arg:         "7605d91/example.txt",
			expectedURL: "https://github.com/arbourd/git-open/tree/%s",
		},
		"out of git dir relative path": {
			arg:         filepath.FromSlash("../../.."),
			expectedURL: "https://github.com/arbourd/git-open/tree/%s",
		},
		"out of git dir absolute path": {
			arg:         filepath.FromSlash(homedir),
			expectedURL: "https://github.com/arbourd/git-open/tree/%s",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ref, err := gitw.CurrentRef(c.gitdir)
			if err != nil {
				panic("Unable to get local ref for test")
			}

			expectedURL := c.expectedURL
			if strings.Count(c.expectedURL, "%s") == 1 {
				expectedURL = fmt.Sprintf(c.expectedURL, ref)
			}

			url, err := GetURL(c.arg)
			if err != nil && c.err == "" {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && len(c.err) > 0 {
				t.Fatalf("expected error:\n\t(GOT): nil\n\t(WNT): %s", c.err)
			} else if url != expectedURL {
				t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, expectedURL)
			}
		})
	}
}

func TestParseRemote(t *testing.T) {
	cases := map[string]struct {
		remote       string
		expectedHost string
		expectedPath string
	}{
		"simple": {
			remote:       "https://github.com/arbourd/git-open",
			expectedHost: "github.com",
			expectedPath: "arbourd/git-open",
		},
		".git suffix": {
			remote:       "https://github.com/arbourd/git-open.git",
			expectedHost: "github.com",
			expectedPath: "arbourd/git-open",
		},
		"extra slashes": {
			remote:       "https://github.com////arbourd/git-open.git",
			expectedHost: "github.com",
			expectedPath: "arbourd/git-open",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			host, path := parseRemote(c.remote)
			if host != c.expectedHost {
				t.Fatalf("unexpected host:\n\t(GOT): %#v\n\t(WNT): %#v", host, c.expectedHost)
			}
			if path != c.expectedPath {
				t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, c.expectedPath)
			}
		})
	}
}

func TestParseType(t *testing.T) {
	cases := map[string]struct {
		arg          string
		expectedType Type
	}{
		"empty argument": {
			arg:          "",
			expectedType: Root,
		},
		"long commit sha": {
			arg:          "7605d912812a5cdc58dc9415026750b43c33928b",
			expectedType: Commit,
		},
		"short commit sha": {
			arg:          "7605d91",
			expectedType: Commit,
		},
		"too short commit sha": {
			arg:          "76",
			expectedType: Path,
		},
		"invalid commit sha": {
			arg:          "7605d91xyz",
			expectedType: Path,
		},
		"commit sha with extension": {
			arg:          "7605d91.txt",
			expectedType: Path,
		},
		"path": {
			arg:          "open_test.go",
			expectedType: Path,
		},
		"relative path": {
			arg:          filepath.FromSlash("../"),
			expectedType: Path,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			rtype := parseType(c.arg)
			if rtype != c.expectedType {
				t.Fatalf("unexpected type:\n\t(GOT): %#v\n\t(WNT): %#v", rtype, c.expectedType)
			}
		})
	}
}

func TestParsePath(t *testing.T) {
	gitdir, err := gitw.GitDir(".")
	if err != nil {
		panic("not a git repository")
	}
	gitdir, _ = filepath.Abs(gitdir)
	gitroot := strings.TrimSuffix(gitdir, ".git")

	cases := map[string]struct {
		path         string
		expectedPath string
		wantErr      bool
	}{
		"empty argument": {
			path:         "",
			expectedPath: "",
		},
		"current directory": {
			path:         ".",
			expectedPath: "open",
		},
		"local file": {
			path:         "open_test.go",
			expectedPath: "open/open_test.go",
		},
		"relative file": {
			path:         filepath.FromSlash("../LICENSE"),
			expectedPath: "LICENSE",
		},
		"relative file with unclean path": {
			path:         filepath.FromSlash(".././/LICENSE"),
			expectedPath: "LICENSE",
		},
		"out of git root": {
			path:         filepath.FromSlash("../../../.."),
			expectedPath: "",
			wantErr:      true,
		},
		"file does not exist": {
			path:         "README.txt",
			expectedPath: "",
			wantErr:      true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			path, err := parsePath(c.path, gitroot)
			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %s\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil")
			} else if path != c.expectedPath {
				t.Fatalf("unexpected path:\n\t(GOT): %#v\n\t(WNT): %#v", path, c.expectedPath)
			}
		})
	}
}
