package open

import (
	"net/url"
	"strings"
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
