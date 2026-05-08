package open

import (
	"fmt"
	"os"
	"os/exec"
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

	f, err := os.Create("abcdef1")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove("abcdef1") })

	cases := map[string]struct {
		gitdir      string
		arg         string
		expectedURL string
		wantErr     bool
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
		"hex-named file": {
			arg:         "abcdef1",
			expectedURL: "https://github.com/arbourd/git-open/tree/%s/open/abcdef1",
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
				t.Fatalf("Unable to get local ref for test: %v", err)
			}

			expectedURL := c.expectedURL
			if strings.Count(c.expectedURL, "%s") == 1 {
				expectedURL = fmt.Sprintf(c.expectedURL, ref)
			}

			url, err := GetURL(c.arg)
			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil\n")
			} else if url != expectedURL {
				t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, expectedURL)
			}
		})
	}
}

func TestGetURLWorktree(t *testing.T) {
	mainDir := t.TempDir()
	worktreeDir := filepath.Join(t.TempDir(), "wt")

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = mainDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	run("remote", "add", "origin", "https://github.com/example/repo.git")
	run("commit", "--allow-empty", "-m", "init")
	run("worktree", "add", "--detach", worktreeDir)

	if err := os.WriteFile(filepath.Join(worktreeDir, "file.txt"), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	t.Chdir(worktreeDir)

	ref, err := gitw.CurrentRef(worktreeDir)
	if err != nil {
		t.Fatalf("unable to get ref: %v", err)
	}

	cases := map[string]struct {
		arg         string
		expectedURL string
	}{
		"root": {
			arg:         "",
			expectedURL: "https://github.com/example/repo",
		},
		"path": {
			arg:         "file.txt",
			expectedURL: fmt.Sprintf("https://github.com/example/repo/tree/%s/file.txt", ref),
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			url, err := GetURL(c.arg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if url != c.expectedURL {
				t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, c.expectedURL)
			}
		})
	}
}

func TestParseRepository(t *testing.T) {
	cases := map[string]struct {
		remote       string
		expectedHost string
		expectedPath string
		wantErr      bool
	}{
		"https protocol": {
			remote:       "https://github.com/arbourd/git-open",
			expectedHost: "github.com",
			expectedPath: "arbourd/git-open",
		},
		"https with .git suffix": {
			remote:       "https://github.com/arbourd/git-open.git",
			expectedHost: "github.com",
			expectedPath: "arbourd/git-open",
		},
		"https with extra slashes": {
			remote:       "https://github.com////arbourd/git-open.git",
			expectedHost: "github.com",
			expectedPath: "arbourd/git-open",
		},
		"ssh protocol": {
			remote:       "git@github.com:arbourd/git-open.git",
			expectedHost: "github.com",
			expectedPath: "arbourd/git-open",
		},
		"git protocol": {
			remote:       "git://github.com/arbourd/git-open.git",
			expectedHost: "github.com",
			expectedPath: "arbourd/git-open",
		},
		"invalid url": {
			remote:  "github/arbourd/git-open.git%x",
			wantErr: true,
		},
		"local absolute path": {
			remote:       "/Users/dylan/repo",
			expectedHost: "",
			expectedPath: "Users/dylan/repo",
		},
		"local relative path": {
			remote:       "../repo",
			expectedHost: "",
			expectedPath: "../repo",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			host, path, err := parseRepository(c.remote)
			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %s\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil")
			} else if host != c.expectedHost {
				t.Fatalf("unexpected host:\n\t(GOT): %#v\n\t(WNT): %#v", host, c.expectedHost)
			} else if path != c.expectedPath {
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
	gitroot, err := gitw.Toplevel(".")
	if err != nil {
		panic("not a git repository")
	}

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
