package gitw

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/ldez/go-git-cmd-wrapper/v2/types"
)

func TestConfigGetRegexp(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	run("config", "open.https://git.example.dev.commitprefix", "commit")
	run("config", "open.https://git.example.dev.pathprefix", "tree")

	t.Chdir(dir)

	out := ConfigGetRegexp(`^open\..*prefix$`)
	if !strings.Contains(out, "open.https://git.example.dev.commitprefix commit") {
		t.Fatalf("unexpected output:\n\t(GOT): %q", out)
	}
	if !strings.Contains(out, "open.https://git.example.dev.pathprefix tree") {
		t.Fatalf("unexpected output:\n\t(GOT): %q", out)
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

			if len(g.BaseOptions) != len(c.expectedOpts) {
				t.Fatalf("unexpected BaseOptions:\n\t(GOT): %#v\n\t(WNT): %#v", g.BaseOptions, c.expectedOpts)
			}
			for i := range c.expectedOpts {
				if g.BaseOptions[i] != c.expectedOpts[i] {
					t.Fatalf("unexpected BaseOptions:\n\t(GOT): %#v\n\t(WNT): %#v", g.BaseOptions, c.expectedOpts)
				}
			}
		})
	}
}

func TestCurrentRef(t *testing.T) {
	dir := t.TempDir()

	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}

	run("init", "-b", "test-branch")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	run("commit", "--allow-empty", "-m", "init")

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
		sha := run("rev-parse", "HEAD")
		run("checkout", "--detach")

		ref, err := CurrentRef(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ref != sha {
			t.Fatalf("unexpected ref:\n\t(GOT): %#v\n\t(WNT): %#v", ref, sha)
		}
	})
}
