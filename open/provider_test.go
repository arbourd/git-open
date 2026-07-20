package open

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ldez/go-git-cmd-wrapper/v2/config"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
)

const repo = "arbourd/git-open"

func TestDefaultProviders(t *testing.T) {
	for _, p := range defaultProviders {
		u, err := url.Parse(p.BaseURL())
		if err != nil {
			t.Errorf("provider %q: invalid baseURL: %v", p.baseURL, err)
			continue
		}
		if u.Host == "" {
			t.Errorf("provider %q: baseURL has no host", p.baseURL)
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			t.Errorf("provider %q: baseURL scheme must be http or https, got %q", p.baseURL, u.Scheme)
		}
	}
}

func TestDefaultProvidersLineFormat(t *testing.T) {
	for _, p := range defaultProviders {
		t.Run(p.baseURL, func(t *testing.T) {
			lineFormat, lineFormatRange, err := parseRawLineFormat(p.rawLineFormat)
			if err != nil {
				t.Fatalf("unexpected error parsing rawLineFormat %q: %v", p.rawLineFormat, err)
			}
			if lineFormat != p.lineFormat {
				t.Errorf("unexpected lineFormat:\n\t(GOT): %#v\n\t(WNT): %#v", lineFormat, p.lineFormat)
			}
			if lineFormatRange != p.lineFormatRange {
				t.Errorf("unexpected lineFormatRange:\n\t(GOT): %#v\n\t(WNT): %#v", lineFormatRange, p.lineFormatRange)
			}
		})
	}
}

func TestCommitURL(t *testing.T) {
	cases := map[string]struct {
		p           Provider
		commit      string
		expectedURL string
	}{
		"github": {
			p:           defaultProviders[0],
			commit:      "7605d91",
			expectedURL: "https://github.com/arbourd/git-open/commit/7605d91",
		},
		"gitlab": {
			p:           defaultProviders[1],
			commit:      "7605d91",
			expectedURL: "https://gitlab.com/arbourd/git-open/-/commit/7605d91",
		},
		"bitbucket": {
			p:           defaultProviders[2],
			commit:      "7605d91",
			expectedURL: "https://bitbucket.org/arbourd/git-open/commits/7605d91",
		},
		"codeberg": {
			p:           defaultProviders[3],
			commit:      "7605d91",
			expectedURL: "https://codeberg.org/arbourd/git-open/commit/7605d91",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			url := c.p.CommitURL(repo, c.commit)
			if url != c.expectedURL {
				t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, c.expectedURL)
			}
		})
	}
}

func TestPathURL(t *testing.T) {
	cases := map[string]struct {
		p           Provider
		ref         string
		path        string
		lstart      int
		lend        int
		expectedURL string
	}{
		"github": {
			p:           defaultProviders[0],
			ref:         "main",
			path:        "main.go",
			expectedURL: "https://github.com/arbourd/git-open/tree/main/main.go",
		},
		"github with line anchor": {
			p:           defaultProviders[0],
			ref:         "main",
			path:        "main.go",
			lstart:      3,
			expectedURL: "https://github.com/arbourd/git-open/tree/main/main.go#L3",
		},
		"github with line anchor range": {
			p:           defaultProviders[0],
			ref:         "main",
			path:        "main.go",
			lstart:      3,
			lend:        5,
			expectedURL: "https://github.com/arbourd/git-open/tree/main/main.go#L3-L5",
		},
		"gitlab": {
			p:           defaultProviders[1],
			ref:         "main",
			path:        "main.go",
			expectedURL: "https://gitlab.com/arbourd/git-open/-/tree/main/main.go",
		},
		"bitbucket": {
			p:           defaultProviders[2],
			ref:         "main",
			path:        "main.go",
			expectedURL: "https://bitbucket.org/arbourd/git-open/src/main/main.go",
		},
		"codeberg": {
			p:           defaultProviders[3],
			ref:         "main",
			path:        "main.go",
			expectedURL: "https://codeberg.org/arbourd/git-open/tree/main/main.go",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			url := c.p.PathURL(repo, c.ref, c.path, c.lstart, c.lend)
			if url != c.expectedURL {
				t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, c.expectedURL)
			}
		})
	}
}

func TestLineAnchor(t *testing.T) {
	cases := map[string]struct {
		p           Provider
		start       int
		end         int
		expectedURL string
	}{
		"github no line": {
			p:           defaultProviders[0],
			start:       0,
			end:         0,
			expectedURL: "",
		},
		"github single line": {
			p:           defaultProviders[0],
			start:       3,
			end:         0,
			expectedURL: "#L3",
		},
		"github range": {
			p:           defaultProviders[0],
			start:       3,
			end:         10,
			expectedURL: "#L3-L10",
		},
		"gitlab single line": {
			p:           defaultProviders[1],
			start:       3,
			end:         0,
			expectedURL: "#L3",
		},
		"gitlab range": {
			p:           defaultProviders[1],
			start:       3,
			end:         10,
			expectedURL: "#L3-10",
		},
		"bitbucket single line": {
			p:           defaultProviders[2],
			start:       3,
			end:         0,
			expectedURL: "#lines-3",
		},
		"bitbucket range": {
			p:           defaultProviders[2],
			start:       3,
			end:         10,
			expectedURL: "#lines-3:10",
		},
		"codeberg single line": {
			p:           defaultProviders[3],
			start:       3,
			end:         0,
			expectedURL: "#L3",
		},
		"codeberg range": {
			p:           defaultProviders[3],
			start:       3,
			end:         10,
			expectedURL: "#L3-L10",
		},
		"single-verb format single line": {
			p:           Provider{lineFormat: "#line-%d"},
			start:       3,
			end:         0,
			expectedURL: "#line-3",
		},
		"single-verb format ignores range end": {
			p:           Provider{lineFormat: "#line-%d"},
			start:       3,
			end:         10,
			expectedURL: "#line-3",
		},
		"empty line format": {
			p:           Provider{},
			start:       3,
			end:         0,
			expectedURL: "",
		},
		"github reversed range": {
			p:           defaultProviders[0],
			start:       10,
			end:         3,
			expectedURL: "#L10-L3",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			url := c.p.lineAnchor(c.start, c.end)
			if url != c.expectedURL {
				t.Fatalf("unexpected anchor:\n\t(GOT): %#v\n\t(WNT): %#v", url, c.expectedURL)
			}
		})
	}
}

func TestRootURL(t *testing.T) {
	cases := map[string]struct {
		p           Provider
		expectedURL string
	}{
		"github": {
			p:           defaultProviders[0],
			expectedURL: "https://github.com/arbourd/git-open",
		},
		"gitlab": {
			p:           defaultProviders[1],
			expectedURL: "https://gitlab.com/arbourd/git-open",
		},
		"bitbucket": {
			p:           defaultProviders[2],
			expectedURL: "https://bitbucket.org/arbourd/git-open",
		},
		"codeberg": {
			p:           defaultProviders[3],
			expectedURL: "https://codeberg.org/arbourd/git-open",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			url := c.p.RootURL(repo)
			if url != c.expectedURL {
				t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, c.expectedURL)
			}
		})
	}
}

func TestFromConfig(t *testing.T) {
	root := t.TempDir()
	gitconfig := filepath.Join(root, ".gitconfig")

	if runtime.GOOS == "darwin" {
		err := os.Setenv("XDG_CONFIG_HOME", root)
		if err != nil {
			t.Fatalf("unable to set XDG_CONFIG_HOME: %s", err)
		}
	}

	// Skip fixture on Windows in CI
	if !(os.Getenv("CI") == "true" && runtime.GOOS == "windows") {
		_, err := os.Create(gitconfig)
		if err != nil {
			t.Fatalf("unable create .gitconfig: %s", err)
		}

		err = os.Setenv("GIT_CONFIG_GLOBAL", gitconfig)
		if err != nil {
			t.Fatalf("unable to set GIT_CONFIG_GLOBAL: %s", err)
		}
	}

	cases := map[string]struct {
		config            []string
		expectedProviders []Provider
	}{
		"empty git config": {
			expectedProviders: []Provider{},
		},
		"single provider": {
			config: []string{
				"open.https://my.domain.dev.commitprefix -/commit",
				"open.https://my.domain.dev.pathprefix -/tree",
				"open.https://my.domain.dev.lineformat L%l-L%l",
			},
			expectedProviders: []Provider{
				{baseURL: "https://my.domain.dev", commitPrefix: "-/commit", pathPrefix: "-/tree", rawLineFormat: "L%l-L%l", lineFormat: "#L%d", lineFormatRange: "#L%d-L%d"},
			},
		},
		"single provider with # prefix": {
			config: []string{
				"open.https://my.domain.dev.commitprefix -/commit",
				"open.https://my.domain.dev.pathprefix -/tree",
				"open.https://my.domain.dev.lineformat #L%l-L%l",
			},
			expectedProviders: []Provider{
				{baseURL: "https://my.domain.dev", commitPrefix: "-/commit", pathPrefix: "-/tree", rawLineFormat: "#L%l-L%l", lineFormat: "#L%d", lineFormatRange: "#L%d-L%d"},
			},
		},
		"multiple providers": {
			config: []string{
				"open.https://git.example1.dev.commitprefix -/commit",
				"open.https://git.example1.dev.pathprefix -/tree",
				"open.https://git.example1.dev.lineformat L%l-L%l",
				"open.https://git.example2.dev.commitprefix commit",
				"open.https://git.example2.dev.pathprefix tree",
				"open.https://git.example2.dev.lineformat L%l-L%l",
			},
			expectedProviders: []Provider{
				{baseURL: "https://git.example1.dev", commitPrefix: "-/commit", pathPrefix: "-/tree", rawLineFormat: "L%l-L%l", lineFormat: "#L%d", lineFormatRange: "#L%d-L%d"},
				{baseURL: "https://git.example2.dev", commitPrefix: "commit", pathPrefix: "tree", rawLineFormat: "L%l-L%l", lineFormat: "#L%d", lineFormatRange: "#L%d-L%d"},
			},
		},
		"multiple subdomains": {
			config: []string{
				"open.https://git.internal.corp.com.commitprefix commit",
				"open.https://git.internal.corp.com.pathprefix tree",
				"open.https://git.internal.corp.com.lineformat L%l-L%l",
			},
			expectedProviders: []Provider{
				{baseURL: "https://git.internal.corp.com", commitPrefix: "commit", pathPrefix: "tree", rawLineFormat: "L%l-L%l", lineFormat: "#L%d", lineFormatRange: "#L%d-L%d"},
			},
		},
		"custom single-verb line format": {
			config: []string{
				"open.https://git.example3.dev.commitprefix commit",
				"open.https://git.example3.dev.pathprefix tree",
				"open.https://git.example3.dev.lineformat line-%l",
			},
			expectedProviders: []Provider{
				{baseURL: "https://git.example3.dev", commitPrefix: "commit", pathPrefix: "tree", rawLineFormat: "line-%l", lineFormat: "#line-%d", lineFormatRange: ""},
			},
		},
		"custom two-verb line format doubles as range format": {
			config: []string{
				"open.https://git.example4.dev.commitprefix commit",
				"open.https://git.example4.dev.pathprefix tree",
				"open.https://git.example4.dev.lineformat #line-%l:%l",
			},
			expectedProviders: []Provider{
				{baseURL: "https://git.example4.dev", commitPrefix: "commit", pathPrefix: "tree", rawLineFormat: "#line-%l:%l", lineFormat: "#line-%d", lineFormatRange: "#line-%d:%d"},
			},
		},
		"invalid url": {
			config: []string{
				"open.not-a-url.commitprefix commit",
				"open.not-a-url.pathprefix tree",
				"open.not-a-url.lineformat L%l-L%l",
			},
			expectedProviders: []Provider{},
		},
		"non-web scheme": {
			config: []string{
				"open.ssh://git.example.dev.commitprefix commit",
				"open.ssh://git.example.dev.pathprefix tree",
				"open.ssh://git.example.dev.lineformat L%l-L%l",
			},
			expectedProviders: []Provider{},
		},
		"empty line format is ignored but not dropped": {
			config: []string{
				"open.https://git.example6.dev.commitprefix commit",
				"open.https://git.example6.dev.pathprefix tree",
			},
			expectedProviders: []Provider{
				{baseURL: "https://git.example6.dev", commitPrefix: "commit", pathPrefix: "tree", rawLineFormat: "", lineFormat: "", lineFormatRange: ""},
			},
		},
		"invalid line format is ignored but not dropped": {
			config: []string{
				"open.https://git.example6.dev.commitprefix commit",
				"open.https://git.example6.dev.pathprefix tree",
				"open.https://git.example6.dev.lineformat #L%s",
			},
			expectedProviders: []Provider{
				{baseURL: "https://git.example6.dev", commitPrefix: "commit", pathPrefix: "tree", rawLineFormat: "#L%s", lineFormat: "", lineFormatRange: ""},
			},
		},
		"line format with too many verbs is ignored but not dropped": {
			config: []string{
				"open.https://git.example7.dev.commitprefix commit",
				"open.https://git.example7.dev.pathprefix tree",
				"open.https://git.example7.dev.lineformat #L%d-%d-%d",
			},
			expectedProviders: []Provider{
				{baseURL: "https://git.example7.dev", commitPrefix: "commit", pathPrefix: "tree", rawLineFormat: "#L%d-%d-%d", lineFormat: "", lineFormatRange: ""},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			// removes all `open.https://` Git config entries
			out, _ := git.Config(config.Global, config.GetRegexp(getRegex, ""))
			for v := range strings.SplitSeq(strings.TrimSpace(out), "\n") {
				key, _, _ := strings.Cut(strings.TrimSpace(v), " ")
				git.Config(config.Global, config.Unset(key, ""))
			}

			for _, v := range c.config {
				key, val, _ := strings.Cut(v, " ")
				git.Config(config.Global, config.Entry(key, val))
			}

			p := fromConfig()
			if len(p) != len(c.expectedProviders) {
				t.Logf("unexpected number of providers\n\t(GOT): %#v\n\t(WNT): %#v", len(p), len(c.expectedProviders))
			}
			sortOpt := cmpopts.SortSlices(func(a, b Provider) bool { return a.baseURL < b.baseURL })
			if !cmp.Equal(p, c.expectedProviders, sortOpt, cmp.AllowUnexported(Provider{})) {
				t.Fatalf("unexpected providers:\n\t(GOT): %#v\n\t(WNT): %#v", p, c.expectedProviders)
			}
		})
	}
}

func TestFromConfigLocal(t *testing.T) {
	// Redirect global config to an empty file so only local config is visible.
	globalConfig := filepath.Join(t.TempDir(), ".gitconfig")
	f, err := os.Create(globalConfig)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Setenv("GIT_CONFIG_GLOBAL", globalConfig)

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
	run("config", "open.https://local.example.dev.commitprefix", "commit")
	run("config", "open.https://local.example.dev.pathprefix", "tree")
	run("config", "open.https://local.example.dev.lineformat", "L%l-L%l")

	t.Chdir(dir)

	providers := fromConfig()
	expected := []Provider{{
		baseURL:         "https://local.example.dev",
		commitPrefix:    "commit",
		pathPrefix:      "tree",
		rawLineFormat:   "L%l-L%l",
		lineFormat:      "#L%d",
		lineFormatRange: "#L%d-L%d",
	}}

	sortOpt := cmpopts.SortSlices(func(a, b Provider) bool { return a.baseURL < b.baseURL })
	if !cmp.Equal(providers, expected, sortOpt, cmp.AllowUnexported(Provider{})) {
		t.Fatalf("unexpected providers:\n\t(GOT): %#v\n\t(WNT): %#v", providers, expected)
	}
}

func TestEscapePath(t *testing.T) {
	cases := map[string]struct {
		url         string
		expectedURL string
	}{
		"simple": {
			url:         "https://github.com/arbourd/git-open",
			expectedURL: "https://github.com/arbourd/git-open",
		},
		"with spaces": {
			url:         "https://github.com/arbourd/git-open/tree/main/file with a space.txt",
			expectedURL: "https://github.com/arbourd/git-open/tree/main/file%20with%20a%20space.txt",
		},
		"trailing slash": {
			url:         "https://github.com/arbourd/git-open/",
			expectedURL: "https://github.com/arbourd/git-open",
		},
		"hash in branch": {
			url:         "https://github.com/arbourd/git-open/tree/fix/#123",
			expectedURL: "https://github.com/arbourd/git-open/tree/fix/%23123",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			url := escapePath(c.url)
			if url != c.expectedURL {
				t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, c.expectedURL)
			}
		})
	}
}

func TestParseRawLineFormat(t *testing.T) {
	cases := map[string]struct {
		rawLineFormat       string
		expectedLineFormat  string
		expectedRangeFormat string
		expectErr           bool
	}{
		"single verb": {
			rawLineFormat:      "L%l",
			expectedLineFormat: "#L%d",
		},
		"single verb already has #": {
			rawLineFormat:      "#L%l",
			expectedLineFormat: "#L%d",
		},
		"two verbs": {
			rawLineFormat:       "L%l-L%l",
			expectedLineFormat:  "#L%d",
			expectedRangeFormat: "#L%d-L%d",
		},
		"two verbs already has #": {
			rawLineFormat:       "#L%l-L%l",
			expectedLineFormat:  "#L%d",
			expectedRangeFormat: "#L%d-L%d",
		},
		"no verbs": {
			rawLineFormat: "L",
		},
		"too many verbs": {
			rawLineFormat: "L%l-%l-%l",
			expectErr:     true,
		},
		"disallowed verb": {
			rawLineFormat: "L%s",
			expectErr:     true,
		},
		"escaped percent adjacent to verb": {
			rawLineFormat: "%%d%l-%l",
			expectErr:     true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			lineFormat, lineFormatRange, err := parseRawLineFormat(c.rawLineFormat)
			if c.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if lineFormat != c.expectedLineFormat {
				t.Errorf("unexpected lineFormat:\n\t(GOT): %#v\n\t(WNT): %#v", lineFormat, c.expectedLineFormat)
			}
			if lineFormatRange != c.expectedRangeFormat {
				t.Errorf("unexpected lineFormatRange:\n\t(GOT): %#v\n\t(WNT): %#v", lineFormatRange, c.expectedRangeFormat)
			}
		})
	}
}

func TestValidateLineFormat(t *testing.T) {
	cases := map[string]struct {
		format        string
		expectedCount int
		expectErr     bool
	}{
		"no verbs": {
			format:        "L",
			expectedCount: 0,
		},
		"one verb": {
			format:        "L%l",
			expectedCount: 1,
		},
		"two verbs": {
			format:        "L%l-%l",
			expectedCount: 2,
		},
		"three verbs": {
			format:    "L%l-%l-%l",
			expectErr: true,
		},
		"escaped percent is not a verb": {
			format:        "L%%%l",
			expectedCount: 1,
		},
		"trailing percent": {
			format:    "L%",
			expectErr: true,
		},
		"disallowed verb %s": {
			format:    "L%s",
			expectErr: true,
		},
		"disallowed verb %q": {
			format:    "L%q",
			expectErr: true,
		},
		"disallowed verb with flags %#v": {
			format:    "L%#v",
			expectErr: true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			count, err := validateLineFormat(c.format)
			if c.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if count != c.expectedCount {
				t.Errorf("unexpected count:\n\t(GOT): %#v\n\t(WNT): %#v", count, c.expectedCount)
			}
		})
	}
}

func TestLtod(t *testing.T) {
	cases := map[string]struct {
		s        string
		expected string
	}{
		"single verb": {
			s:        "#L%l",
			expected: "#L%d",
		},
		"two verbs": {
			s:        "#L%l-L%l",
			expected: "#L%d-L%d",
		},
		"no verb": {
			s:        "#L%d",
			expected: "#L%d",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			got := ltod(c.s)
			if got != c.expected {
				t.Errorf("unexpected result:\n\t(GOT): %#v\n\t(WNT): %#v", got, c.expected)
			}
		})
	}
}
