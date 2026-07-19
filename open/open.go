package open

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/arbourd/git-get/get"
	"github.com/arbourd/git-open/gitw"
)

// Type represents the type of Git URL to open
type Type int

const (
	// Commit is a specific commit in the repository
	Commit Type = iota

	// Path is a file, folder or path in the repository
	Path

	// Root is the root of the repository
	Root
)

// InBrowser opens a URL in the default browser
func InBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "--", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", "--", url)
	}

	return cmd.Start()
}

// GetURL returns the URL to open based on the arg provided
func GetURL(arg string) (string, error) {
	gitroot, err := gitw.Toplevel(".")
	if err != nil {
		// If toplevel fails, we might be in a bare repo
		gitroot, err = gitw.AbsoluteGitDir(".")
		if err != nil {
			return "", fmt.Errorf("not a git repository")
		}
	}

	t := parseType(arg)
	if t == Commit {
		if _, err := os.Stat(arg); err == nil {
			t = Path
		}
	}
	var lstart, lend int
	if t == Path {
		// Ignore parsePath errors: invalid or out-of-repo paths fall back to the root URL
		arg, lstart, lend, _ = parsePath(arg, gitroot)
	}

	remote, ref, err := getRemoteRef(gitroot)
	if err != nil {
		return "", err
	}

	host, repo, err := parseRepository(remote)
	if err != nil {
		return "", err
	}
	if host == "" {
		return "", fmt.Errorf("local remotes are not supported")
	}

	providers := Providers()

	// Find the provider by exact host comparison.
	var p Provider
	for _, provider := range providers {
		u, err := url.Parse(provider.BaseURL())
		if err != nil {
			continue
		}
		if u.Host == host {
			p = provider
			break
		}
	}

	if len(p.BaseURL()) == 0 {
		return "", fmt.Errorf("unable to find provider for: \"%s\"", host)
	}

	var openURL string
	switch t {
	case Commit:
		openURL = p.CommitURL(repo, arg)
	case Path:
		openURL = p.PathURL(repo, ref, arg, lstart, lend)
	case Root:
		openURL = p.RootURL(repo)
	}

	return openURL, nil
}

// parsePath returns the cleaned path, relative to the gitroot, and the parsed start and end line numbers
func parsePath(path, gitroot string) (string, int, int, error) {
	if path == "" {
		return "", 0, 0, nil
	}

	var lstart, lend int
	if stripped, start, end := stripLine(path); stripped != path {
		// Prefer the literal, colon-suffixed argument when it names a real
		// file or directory; otherwise treat the suffix as a line spec.
		if _, statErr := os.Stat(path); statErr != nil {
			path, lstart, lend = stripped, start, end
		}
	}
	path = filepath.Clean(path)

	info, err := os.Stat(path)
	if err != nil && (os.IsNotExist(err) || ancestorIsFile(path)) {
		return "", 0, 0, err
	}
	if err == nil && info.IsDir() {
		lstart, lend = 0, 0
	}

	path, _ = filepath.Abs(path)
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}

	// Check if path is within Git root
	rel, err := filepath.Rel(gitroot, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", 0, 0, fmt.Errorf("path does not contain gitroot: %s; %s", path, gitroot)
	}

	if rel == "." {
		return "", 0, 0, nil
	}

	// Convert all path separators to `/` and trim trailing `/`
	return filepath.ToSlash(rel), lstart, lend, nil
}

// ancestorIsFile reports whether path is invalid directory because an ancestor
// is a file -- POSIX's ENOTDIR or Windows' ERROR_PATH_NOT_FOUND
func ancestorIsFile(path string) bool {
	dir := filepath.Dir(path)
	for {
		parent, err := os.Stat(dir)
		if err == nil {
			return !parent.IsDir()
		}
		next := filepath.Dir(dir)
		if next == dir {
			return false
		}
		dir = next
	}
}

// getRemoteRef returns the Git remote and reference (branch, tag, commit), for a provided Git repository
func getRemoteRef(gitroot string) (remote string, ref string, err error) {
	remote, err = gitw.RemoteURL(gitroot)
	if err != nil {
		return "", "", err
	}

	ref, err = gitw.CurrentRef(gitroot)
	return remote, ref, err
}

// parseRepository parses the host and repository (username or organization and repository name) from a remote string
func parseRepository(remote string) (host string, repo string, err error) {
	url, err := get.ParseURL(remote)
	if err != nil {
		return "", "", fmt.Errorf("unable to parse remote url: %w", err)
	}

	repo = strings.TrimPrefix(strings.TrimSuffix(path.Clean(url.Path), ".git"), "/")
	return url.Host, repo, nil
}

// lineSuffixRegex matches trailing line numbers
var lineSuffixRegex = regexp.MustCompile(`^[0-9-]+$`)

// stripLine splits a trailing line suffix from arg, returning the path and line numbers
func stripLine(arg string) (path string, start int, end int) {
	i := strings.LastIndex(arg, ":")
	if i <= 0 {
		return arg, 0, 0
	}
	path, suffix := arg[:i], arg[i+1:]
	if !lineSuffixRegex.MatchString(suffix) {
		return arg, 0, 0
	}

	before, after, isRange := strings.Cut(suffix, "-")

	start, err := strconv.Atoi(before)
	if err != nil || start < 1 {
		return path, 0, 0
	}
	if !isRange {
		return path, start, 0
	}

	end, err = strconv.Atoi(after)
	if err != nil || end < 1 {
		return path, start, 0
	}
	return path, start, end
}

var commitSHARegex = regexp.MustCompile(`^[0-9a-f]{7,64}$`)

// parseType parses and returns the Type of argument
func parseType(arg string) Type {
	// Return root type if no arg provided
	if len(arg) == 0 {
		return Root
	}

	// Check if arg is commit sha
	if commitSHARegex.MatchString(arg) {
		return Commit
	}

	// Assume all other arg are paths
	return Path
}
