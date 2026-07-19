package open

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
		"file with line": {
			arg:         "open_test.go:3",
			expectedURL: "https://github.com/arbourd/git-open/tree/%s/open/open_test.go#L3",
		},
		"file with line range": {
			arg:         "open_test.go:3-10",
			expectedURL: "https://github.com/arbourd/git-open/tree/%s/open/open_test.go#L3-L10",
		},
		"file with line zero drops the line, keeps the file": {
			arg:         "open_test.go:0",
			expectedURL: "https://github.com/arbourd/git-open/tree/%s/open/open_test.go",
		},
		"file with reversed line range": {
			arg:         "open_test.go:10-3",
			expectedURL: "https://github.com/arbourd/git-open/tree/%s/open/open_test.go#L10-L3",
		},
		"non-numeric suffix is not a line spec, no literal match falls back to root": {
			arg:         "open_test.go:abc",
			expectedURL: "https://github.com/arbourd/git-open/tree/%s",
		},
		"literal colon filename with no matching file falls back to root": {
			arg:         "notes:42",
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

func TestGetURLErrors(t *testing.T) {
	git := func(t *testing.T, dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	cases := map[string]struct {
		setup   func(t *testing.T, dir string)
		wantErr string
	}{
		"not a git repository": {
			setup:   func(t *testing.T, dir string) {},
			wantErr: "not a git repository",
		},
		"local remote": {
			setup: func(t *testing.T, dir string) {
				git(t, dir, "init")
				git(t, dir, "config", "user.email", "test@example.com")
				git(t, dir, "config", "user.name", "Test")
				git(t, dir, "remote", "add", "origin", filepath.Join(t.TempDir(), "local", "repo"))
				git(t, dir, "commit", "--allow-empty", "-m", "init")
			},
			wantErr: "local remotes are not supported",
		},
		"unsupported provider": {
			setup: func(t *testing.T, dir string) {
				git(t, dir, "init")
				git(t, dir, "config", "user.email", "test@example.com")
				git(t, dir, "config", "user.name", "Test")
				git(t, dir, "remote", "add", "origin", "https://unknown.example.com/user/repo.git")
				git(t, dir, "commit", "--allow-empty", "-m", "init")
			},
			wantErr: `unable to find provider for: "unknown.example.com"`,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			c.setup(t, dir)
			t.Chdir(dir)

			_, err := GetURL("")
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Fatalf("unexpected error:\n\t(GOT): %q\n\t(WNT): contains %q", err.Error(), c.wantErr)
			}
		})
	}
}

func TestGetURLBareRepo(t *testing.T) {
	mainDir := t.TempDir()
	bareDir := t.TempDir()

	run := func(dir string, args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run(mainDir, "init")
	run(mainDir, "config", "user.email", "test@example.com")
	run(mainDir, "config", "user.name", "Test")
	run(mainDir, "remote", "add", "origin", "https://github.com/example/repo.git")
	run(mainDir, "commit", "--allow-empty", "-m", "init")
	run(mainDir, "clone", "--bare", mainDir, bareDir)
	run(bareDir, "remote", "set-url", "origin", "https://github.com/example/repo.git")

	t.Chdir(bareDir)

	url, err := GetURL("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://github.com/example/repo" {
		t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, "https://github.com/example/repo")
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

func TestGetURLWindowsAbsolutePath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("windows drive-letter path parsing only applies on windows")
	}

	ref, err := gitw.CurrentRef(".")
	if err != nil {
		t.Fatalf("unable to get local ref for test: %v", err)
	}

	abs, err := filepath.Abs("open_test.go")
	if err != nil {
		t.Fatalf("unable to get absolute path: %v", err)
	}

	url, err := GetURL(abs + ":3-10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := fmt.Sprintf("https://github.com/arbourd/git-open/tree/%s/open/open_test.go#L3-L10", ref)
	if url != expectedURL {
		t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, expectedURL)
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

func TestStripLine(t *testing.T) {
	cases := map[string]struct {
		arg           string
		expectedPath  string
		expectedStart int
		expectedEnd   int
	}{
		"no line suffix": {
			arg:          "main.go",
			expectedPath: "main.go",
		},
		"single line": {
			arg:           "main.go:3",
			expectedPath:  "main.go",
			expectedStart: 3,
		},
		"line range": {
			arg:           "main.go:3-10",
			expectedPath:  "main.go",
			expectedStart: 3,
			expectedEnd:   10,
		},
		"nested path with line range": {
			arg:           "a/b/main.go:3-10",
			expectedPath:  "a/b/main.go",
			expectedStart: 3,
			expectedEnd:   10,
		},
		"multiple colons splits on the last one": {
			arg:           "a:b:3",
			expectedPath:  "a:b",
			expectedStart: 3,
		},
		"trailing colon falls back to full arg": {
			arg:          "main.go:",
			expectedPath: "main.go:",
		},
		"no path before colon falls back to full arg": {
			arg:          ":3",
			expectedPath: ":3",
		},
		"non-numeric suffix falls back to full arg": {
			arg:          "main.go:abc",
			expectedPath: "main.go:abc",
		},
		"non-numeric range end falls back to full arg": {
			arg:          "main.go:5-abc",
			expectedPath: "main.go:5-abc",
		},
		"windows absolute path with no line suffix falls back to full arg": {
			arg:          `C:\Example\file.txt`,
			expectedPath: `C:\Example\file.txt`,
		},
		"windows absolute path with line range": {
			arg:           `C:\Example\file.txt:5-10`,
			expectedPath:  `C:\Example\file.txt`,
			expectedStart: 5,
			expectedEnd:   10,
		},
		"windows absolute path with reversed range": {
			arg:           `C:\Example\file.txt:10-3`,
			expectedPath:  `C:\Example\file.txt`,
			expectedStart: 10,
			expectedEnd:   3,
		},
		"trailing hyphen falls back to start": {
			arg:           "main.go:3-",
			expectedPath:  "main.go",
			expectedStart: 3,
		},
		"line zero drops to path": {
			arg:          "main.go:0",
			expectedPath: "main.go",
		},
		"range end zero falls back to start": {
			arg:           "main.go:5-0",
			expectedPath:  "main.go",
			expectedStart: 5,
		},
		"range start zero drops to path": {
			arg:          "main.go:0-10",
			expectedPath: "main.go",
		},
		"reversed range is allowed": {
			arg:           "main.go:10-3",
			expectedPath:  "main.go",
			expectedStart: 10,
			expectedEnd:   3,
		},
		"reversed range close together is allowed": {
			arg:           "main.go:5-1",
			expectedPath:  "main.go",
			expectedStart: 5,
			expectedEnd:   1,
		},
		"negative start drops to path": {
			arg:          "main.go:-1-5",
			expectedPath: "main.go",
		},
		"equal range": {
			arg:           "main.go:5-5",
			expectedPath:  "main.go",
			expectedStart: 5,
			expectedEnd:   5,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			path, start, end := stripLine(c.arg)
			if path != c.expectedPath || start != c.expectedStart || end != c.expectedEnd {
				t.Fatalf("unexpected result:\n\t(GOT): path=%#v start=%#v end=%#v\n\t(WNT): path=%#v start=%#v end=%#v",
					path, start, end, c.expectedPath, c.expectedStart, c.expectedEnd)
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
		path          string
		expectedPath  string
		expectedStart int
		expectedEnd   int
		wantErr       bool
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
		"local file with line": {
			path:          "open_test.go:3",
			expectedPath:  "open/open_test.go",
			expectedStart: 3,
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
		"path walks through a file": {
			path:         filepath.FromSlash("open_test.go/foo"),
			expectedPath: "",
			wantErr:      true,
		},
		"git root with line is dropped": {
			path:         filepath.FromSlash("../:5"),
			expectedPath: "",
		},
		"directory with line is dropped": {
			path:         filepath.FromSlash("../open:5"),
			expectedPath: "open",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			path, start, end, err := parsePath(c.path, gitroot)
			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %s\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil")
			} else if path != c.expectedPath || start != c.expectedStart || end != c.expectedEnd {
				t.Fatalf("unexpected result:\n\t(GOT): path=%#v start=%#v end=%#v\n\t(WNT): path=%#v start=%#v end=%#v",
					path, start, end, c.expectedPath, c.expectedStart, c.expectedEnd)
			}
		})
	}
}
