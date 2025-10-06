package open

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ldez/go-git-cmd-wrapper/v2/config"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
)

const repo = "arbourd/git-open"

func TestCommitURL(t *testing.T) {
	cases := map[string]struct {
		p           Provider
		commit      string
		expectedURL string
	}{
		"github": {
			p:           DefaultProviders[0],
			commit:      "7605d91",
			expectedURL: "https://github.com/arbourd/git-open/commit/7605d91",
		},
		"gitlab": {
			p:           DefaultProviders[1],
			commit:      "7605d91",
			expectedURL: "https://gitlab.com/arbourd/git-open/-/commit/7605d91",
		},
		"bitbucket": {
			p:           DefaultProviders[2],
			commit:      "7605d91",
			expectedURL: "https://bitbucket.org/arbourd/git-open/commits/7605d91",
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
		expectedURL string
	}{
		"github": {
			p:           DefaultProviders[0],
			ref:         "main",
			path:        "LICENSE",
			expectedURL: "https://github.com/arbourd/git-open/tree/main/LICENSE",
		},
		"gitlab": {
			p:           DefaultProviders[1],
			ref:         "main",
			path:        "LICENSE",
			expectedURL: "https://gitlab.com/arbourd/git-open/-/tree/main/LICENSE",
		},
		"bitbucket": {
			p:           DefaultProviders[2],
			ref:         "main",
			path:        "LICENSE",
			expectedURL: "https://bitbucket.org/arbourd/git-open/src/main/LICENSE",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			url := c.p.PathURL(repo, c.ref, c.path)
			if url != c.expectedURL {
				t.Fatalf("unexpected url:\n\t(GOT): %#v\n\t(WNT): %#v", url, c.expectedURL)
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
			p:           DefaultProviders[0],
			expectedURL: "https://github.com/arbourd/git-open",
		},
		"gitlab": {
			p:           DefaultProviders[1],
			expectedURL: "https://gitlab.com/arbourd/git-open",
		},
		"bitbucket": {
			p:           DefaultProviders[2],
			expectedURL: "https://bitbucket.org/arbourd/git-open",
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

func TestLoadProviders(t *testing.T) {
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
			},
			expectedProviders: []Provider{
				{BaseURL: "https://my.domain.dev", CommitPrefix: "-/commit", PathPrefix: "-/tree"},
			},
		},
		"multiple providers": {
			config: []string{
				"open.https://git.example1.dev.commitprefix -/commit",
				"open.https://git.example1.dev.pathprefix -/tree",
				"open.https://git.example2.dev.commitprefix commit",
				"open.https://git.example2.dev.pathprefix tree",
			},
			expectedProviders: []Provider{
				{BaseURL: "https://git.example1.dev", CommitPrefix: "-/commit", PathPrefix: "-/tree"},
				{BaseURL: "https://git.example2.dev", CommitPrefix: "commit", PathPrefix: "tree"},
			},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			// removes all `open.https://` Git config entries
			out, _ := git.Config(config.Global, config.GetRegexp(getRegex, ""))
			for _, v := range strings.Split(strings.TrimSpace(out), "\n") {
				key := strings.Split(strings.TrimSpace(v), " ")[0]

				git.Config(config.Global, config.Unset(key, ""))
			}

			for _, v := range c.config {
				s := strings.Split(v, " ")
				git.Config(config.Global, config.Entry(s[0], s[1]))
			}

			p := LoadProviders()
			if len(p) != len(c.expectedProviders) {
				t.Logf("unexpected number of providers\n\t(GOT): %#v\n\t(WNT): %#v", len(p), len(c.expectedProviders))
			}
			if !cmp.Equal(p, c.expectedProviders) {
				t.Fatalf("unexpected providers:\n\t(GOT): %#v\n\t(WNT): %#v", p, c.expectedProviders)
			}
		})
	}
}
