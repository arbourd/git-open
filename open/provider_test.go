package open

import (
	"testing"
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
		branch      string
		path        string
		expectedURL string
	}{
		"github": {
			p:           DefaultProviders[0],
			branch:      "main",
			path:        "LICENSE",
			expectedURL: "https://github.com/arbourd/git-open/tree/main/LICENSE",
		},
		"gitlab": {
			p:           DefaultProviders[1],
			branch:      "main",
			path:        "LICENSE",
			expectedURL: "https://gitlab.com/arbourd/git-open/-/tree/main/LICENSE",
		},
		"bitbucket": {
			p:           DefaultProviders[2],
			branch:      "main",
			path:        "LICENSE",
			expectedURL: "https://bitbucket.org/arbourd/git-open/src/main/LICENSE",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			url := c.p.PathURL(repo, c.branch, c.path)
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
