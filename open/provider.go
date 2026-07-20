package open

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/arbourd/git-open/gitw"
)

// defaultProviders is a list of built-in Providers
var defaultProviders = []Provider{
	{
		baseURL:      "https://github.com",
		commitPrefix: "commit",
		pathPrefix:   "tree",

		rawLineFormat:   "L%l-L%l",
		lineFormat:      "#L%d",
		lineFormatRange: "#L%d-L%d",
	},
	{
		baseURL:      "https://gitlab.com",
		commitPrefix: "-/commit",
		pathPrefix:   "-/tree",

		rawLineFormat:   "L%l-%l",
		lineFormat:      "#L%d",
		lineFormatRange: "#L%d-%d",
	},
	{
		baseURL:      "https://bitbucket.org",
		commitPrefix: "commits",
		pathPrefix:   "src",

		rawLineFormat:   "lines-%l:%l",
		lineFormat:      "#lines-%d",
		lineFormatRange: "#lines-%d:%d",
	},
	{
		baseURL:      "https://codeberg.org",
		commitPrefix: "commit",
		pathPrefix:   "tree",

		rawLineFormat:   "L%l-L%l",
		lineFormat:      "#L%d",
		lineFormatRange: "#L%d-L%d",
	},
}

func Providers() []Provider {
	return slices.Concat(defaultProviders, fromConfig())
}

// Provider represents the Git platforms domain and URL pathing.
type Provider struct {
	baseURL      string
	commitPrefix string
	pathPrefix   string

	rawLineFormat   string
	lineFormat      string
	lineFormatRange string
}

// BaseURL returns the provider's base URL as a string
func (p Provider) BaseURL() string {
	return p.baseURL
}

// CommitURL returns URL of a commit as a string
func (p Provider) CommitURL(repo, commitSHA string) string {
	return escapePath(strings.Join([]string{p.baseURL, repo, p.commitPrefix, commitSHA}, "/"))
}

// PathURL returns URL of a file with line anchors as a string
func (p Provider) PathURL(repo, ref, path string, lstart, lend int) string {
	u := strings.Join([]string{p.baseURL, repo, p.pathPrefix, ref, path}, "/")
	return escapePath(u) + p.lineAnchor(lstart, lend)
}

// RootURL returns URL of the root repository as a string
func (p Provider) RootURL(repo string) string {
	return escapePath(strings.Join([]string{p.baseURL, repo}, "/"))
}

// lineAnchor returns a URL anchor highlighting a line or range of lines like `#L3` or `#L3-L10`
func (p Provider) lineAnchor(start, end int) string {
	if start == 0 || p.lineFormat == "" {
		return ""
	}
	if end == 0 || p.lineFormatRange == "" {
		return fmt.Sprintf(p.lineFormat, start)
	}
	return fmt.Sprintf(p.lineFormatRange, start, end)
}

const getRegex = `^open\..*(prefix|format)$`

// fromConfig returns a slice of [Provider] from the global Git config.
//
// The Git config structure includes a base URL as an argument, commit prefix, path prefix and line format string.
//
//	[open "https://git.mydomain.dev"]
//	  commitprefix = commit
//	  pathprefix = tree
//	  lineformat = L%l-L%l
func fromConfig() []Provider {
	providers := []Provider{}
	out := gitw.ConfigGetRegexp(getRegex)
	if len(out) == 0 {
		return providers
	}

	var order []string
	urls := make(map[string]*Provider)
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
			entry = &Provider{}
			urls[rawURL] = entry
			order = append(order, rawURL)
		}

		switch key {
		case "commitprefix":
			entry.commitPrefix = value
		case "pathprefix":
			entry.pathPrefix = value
		case "lineformat":
			entry.rawLineFormat = value
		}
	}

	for _, k := range order {
		var skip bool

		u, err := url.Parse(k)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
			fmt.Fprintf(os.Stderr, "warning: invalid provider URL in git config: %q, skipping provider\n", k)
			skip = true
		}

		v := urls[k]
		if v.commitPrefix == "" {
			fmt.Fprintf(os.Stderr, "warning: provider %q is missing commitprefix in git config, skipping provider\n", k)
			skip = true
		}
		if v.pathPrefix == "" {
			fmt.Fprintf(os.Stderr, "warning: provider %q is missing pathprefix in git config, skipping provider\n", k)
			skip = true
		}
		if skip {
			continue
		}

		lineFormat, lineFormatRange, err := parseRawLineFormat(v.rawLineFormat)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: invalid lineformat for %q in git config: %v\n", k, err)
		} else if lineFormat == "" {
			fmt.Fprintf(os.Stderr, "warning: provider %q is missing lineformat in git config\n", k)
		}

		v.baseURL = k
		v.lineFormat = lineFormat
		v.lineFormatRange = lineFormatRange
		providers = append(providers, *v)
	}

	return providers
}

// parseRawLineFormat parses the single argument line format and range from the raw line format
// with an html anchor or returns a validation error
func parseRawLineFormat(rawLineFormat string) (lineFormat, lineFormatRange string, err error) {
	count, err := validateLineFormat(rawLineFormat)
	if err != nil {
		return "", "", err
	}
	if count == 0 {
		return "", "", nil
	}
	if !strings.HasPrefix(rawLineFormat, "#") {
		rawLineFormat = "#" + rawLineFormat
	}

	rawLineFormat = ltod(rawLineFormat)
	if strings.Count(rawLineFormat, "%d") != count {
		return "", "", errors.New("ambiguous line format: escaped `%%` collides with `%d`")
	}

	if count == 1 {
		return rawLineFormat, "", nil
	}

	lineFormat = strings.SplitAfterN(rawLineFormat, "%d", 2)[0]
	return lineFormat, rawLineFormat, nil
}

// validateLineFormat returns the number of `%l` verbs in a lineformat string, or an error if
// there are more than two verbs or any verb other than the fake `%l` is used
func validateLineFormat(format string) (count int, err error) {
	stripped := strings.ReplaceAll(format, "%%", "")
	count = strings.Count(stripped, "%l")
	if strings.Count(stripped, "%") != count {
		return 0, fmt.Errorf("unsupported verb in line format: %q", format)
	}
	if count > 2 {
		return 0, fmt.Errorf("too many %%l verbs in line format (max 2): %q", format)
	}

	return count, nil
}

// ltod convers the custom format `%l` verb to `%d`
func ltod(s string) string {
	return strings.ReplaceAll(s, "%l", "%d")
}

// escapePath escapes the URL with url.PathEscape, but unescapes `/` and trims trailing `/`
func escapePath(u string) string {
	return strings.TrimSuffix(strings.ReplaceAll(url.PathEscape(u), "%2F", "/"), "/")
}
