package gitw

import (
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/ldez/go-git-cmd-wrapper/v2/types"
)

// runGit runs git in dir with a hermetic environment: inherited repo-selection
// variables are dropped and commit signing is disabled.
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	cmd.Env = slices.DeleteFunc(os.Environ(), func(kv string) bool {
		key, _, _ := strings.Cut(kv, "=")
		return key == "GIT_DIR" || key == "GIT_WORK_TREE" || key == "GIT_INDEX_FILE"
	})
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestConfigGetRegexp(t *testing.T) {
	dir := t.TempDir()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	runGit(t, dir, "config", "open.https://git.example.dev.commitprefix", "commit")
	runGit(t, dir, "config", "open.https://git.example.dev.pathprefix", "tree")

	t.Chdir(dir)

	out := ConfigGetRegexp(`^open\..*prefix$`)
	if !strings.Contains(out, "open.https://git.example.dev.commitprefix commit") {
		t.Fatalf("unexpected output:\n\t(GOT): %q", out)
	}
	if !strings.Contains(out, "open.https://git.example.dev.pathprefix tree") {
		t.Fatalf("unexpected output:\n\t(GOT): %q", out)
	}
}

func TestRemoteURL(t *testing.T) {
	dir := t.TempDir()

	runGit(t, dir, "init")
	runGit(t, dir, "remote", "add", "origin", "https://example.com/a/origin.git")
	runGit(t, dir, "remote", "add", "fork", "https://example.com/b/fork.git")

	cases := map[string]struct {
		remote  string
		want    string
		wantErr bool
	}{
		"default":        {remote: "", want: "https://example.com/a/origin.git"},
		"named":          {remote: "fork", want: "https://example.com/b/fork.git"},
		"unknown":        {remote: "nope", wantErr: true},
		"url not a name": {remote: "https://example.com/c/d.git", wantErr: true},
		"flag-like":      {remote: "--exit-code", wantErr: true},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := RemoteURL(dir, c.remote)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil || got != c.want {
				t.Fatalf("unexpected result:\n\t(GOT): %q, %v\n\t(WNT): %q", got, err, c.want)
			}
		})
	}
}

func TestCwd(t *testing.T) {
	cases := map[string]struct {
		path         string
		expectedOpts []string
	}{
		"non-empty path": {
			path:         "/some/path",
			expectedOpts: []string{"-C", "/some/path"},
		},
		"empty path": {
			path:         "",
			expectedOpts: nil,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			g := &types.Cmd{}
			cwd(c.path)(g)

			if !slices.Equal(g.BaseOptions, c.expectedOpts) {
				t.Fatalf("unexpected BaseOptions:\n\t(GOT): %#v\n\t(WNT): %#v", g.BaseOptions, c.expectedOpts)
			}
		})
	}
}

func TestCurrentRef(t *testing.T) {
	dir := t.TempDir()

	runGit(t, dir, "init", "-b", "test-branch")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	runGit(t, dir, "commit", "--allow-empty", "-m", "init")

	t.Run("branch", func(t *testing.T) {
		ref, err := CurrentRef(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref != "test-branch" {
			t.Fatalf("unexpected ref:\n\t(GOT): %#v\n\t(WNT): %#v", ref, "test-branch")
		}
	})

	t.Run("detached HEAD", func(t *testing.T) {
		sha := runGit(t, dir, "rev-parse", "HEAD")
		runGit(t, dir, "checkout", "--detach")

		ref, err := CurrentRef(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref != sha {
			t.Fatalf("unexpected ref:\n\t(GOT): %#v\n\t(WNT): %#v", ref, sha)
		}
	})
}
