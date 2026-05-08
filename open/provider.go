package open

import (
	"net/url"
	"strings"

	"github.com/ldez/go-git-cmd-wrapper/v2/config"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
)

// DefaultProviders are a list of supported Providers
var DefaultProviders = []Provider{
	{
		BaseURL:      "https://github.com",
		CommitPrefix: "commit",
		PathPrefix:   "tree",
	},
	{
		BaseURL:      "https://gitlab.com",
		CommitPrefix: "-/commit",
		PathPrefix:   "-/tree",
	},
	{
		BaseURL:      "https://bitbucket.org",
		CommitPrefix: "commits",
		PathPrefix:   "src",
	},
}

// Provider represents the Git platforms domain and URL pathing.
type Provider struct {
	BaseURL      string
	CommitPrefix string
	PathPrefix   string
}

// CommitURL returns URL of a commit as a string
func (p Provider) CommitURL(repo, commitSHA string) string {
	return escapePath(strings.Join([]string{p.BaseURL, repo, p.CommitPrefix, commitSHA}, "/"))
}

// PathURL returns URL of a file as a string
func (p Provider) PathURL(repo, ref, path string) string {
	return escapePath(strings.Join([]string{p.BaseURL, repo, p.PathPrefix, ref, path}, "/"))
}

// RootURL returns URL of the root repository as a string
func (p Provider) RootURL(repo string) string {
	return escapePath(strings.Join([]string{p.BaseURL, repo}, "/"))
}

// escapePath escapes the URL with url.PathEscape, but unescapes `/` and trims trailing `/`
func escapePath(u string) string {
	return strings.TrimSuffix(strings.ReplaceAll(url.PathEscape(u), "%2F", "/"), "/")
}

const getRegex = `^open\..*prefix$`

// LoadProviders returns a slice of [Provider] from the global Git config.
//
// The Git config structure includes a base URL as an argument, and commit prefix and path prefix keys and values.
//
//	[open "https://git.mydomain.dev"]
//	  commitprefix = commit
//	  pathprefix = tree
func LoadProviders() []Provider {
	p := []Provider{}
	out, _ := git.Config(config.Global, config.GetRegexp(getRegex, ""))
	out = strings.TrimSpace(out)
	if len(out) == 0 {
		return p
	}

	urls := make(map[string]*struct {
		commitPrefix string
		pathPrefix   string
	})
	for line := range strings.SplitSeq(out, "\n") {
		s := strings.SplitN(line, " ", 2)
		if len(s) != 2 {
			continue
		}
		fullKey, value := s[0], s[1]

		rest, ok := strings.CutPrefix(fullKey, "open.")
		if !ok {
			continue
		}

		i := strings.LastIndex(rest, ".")
		if i == -1 {
			continue
		}

		rawURL, key := rest[:i], rest[i+1:]

		entry := urls[rawURL]
		if entry == nil {
			entry = &struct {
				commitPrefix string
				pathPrefix   string
			}{}
			urls[rawURL] = entry
		}

		switch key {
		case "commitprefix":
			entry.commitPrefix = value
		case "pathprefix":
			entry.pathPrefix = value
		}
	}

	for k, v := range urls {
		p = append(p, Provider{
			BaseURL:      k,
			CommitPrefix: v.commitPrefix,
			PathPrefix:   v.pathPrefix,
		})
	}

	return p
}
