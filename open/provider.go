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
	for _, v := range strings.Split(out, "\n") {
		s := strings.Split(strings.TrimPrefix(v, "open."), " ")

		var value string
		if len(s) == 2 {
			value = s[1]
		}

		s = strings.Split(s[0], ".")
		key := s[len(s)-1]
		url := strings.Join(s[0:len(s)-1], ".")

		if urls[url] == nil {
			urls[url] = &struct {
				commitPrefix string
				pathPrefix   string
			}{}
		}

		switch key {
		case "commitprefix":
			urls[url].commitPrefix = value
		case "pathprefix":
			urls[url].pathPrefix = value
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
