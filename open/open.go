package open

import (
	"errors"
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

// GetURL returns the URL to open based on the arg provided.
// A non-empty ref (branch, tag, commit) replaces the current ref for paths and the root.
// A non-empty remote names the Git remote to build the URL from, instead of the default remote.
func GetURL(arg, ref, remote string) (string, error) {
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
	override := ref != ""
	if override {
		if t == Commit {
			return "", fmt.Errorf("a ref cannot be combined with a commit")
		}
		// The root at a ref is the path URL with an empty path
		t = Path
	}
	var lstart, lend int
	if t == Path {
		// Ignore parsePath errors: invalid or out-of-repo paths fall back to the root URL
		arg, lstart, lend, _ = parsePath(arg, gitroot, !override)
	}

	remoteURL, ref, err := getRemoteRef(gitroot, ref, remote)
	if err != nil {
		return "", err
	}

	host, repo, err := parseRepository(remoteURL)
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

// parsePath returns the cleaned path, relative to the gitroot, and the parsed start and end line numbers.
// A path missing from disk is an error if mustExist, otherwise it is kept, as it may exist at another ref.
func parsePath(path, gitroot string, mustExist bool) (string, int, int, error) {
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
	if err != nil && mustExist && (os.IsNotExist(err) || ancestorIsFile(path)) {
		return "", 0, 0, err
	}
	if err == nil && info.IsDir() {
		lstart, lend = 0, 0
	}

	path, _ = filepath.Abs(path)
	path = resolveSymlinks(path)

	// Check if path is within Git root
	rel, err := filepath.Rel(gitroot, path)
	if err != nil || !filepath.IsLocal(rel) {
		return "", 0, 0, fmt.Errorf("path does not contain gitroot: %s; %s", path, gitroot)
	}

	if rel == "." {
		return "", 0, 0, nil
	}

	// Convert all path separators to `/` and trim trailing `/`
	return filepath.ToSlash(rel), lstart, lend, nil
}

// resolveSymlinks resolves symlinks in path, or in its deepest existing ancestor if path is missing
func resolveSymlinks(path string) string {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	dir := filepath.Dir(path)
	if dir == path {
		return path
	}
	return filepath.Join(resolveSymlinks(dir), filepath.Base(path))
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

// getRemoteRef returns the URL of the named remote (default when empty) and reference (branch, tag, commit), for a provided Git repository.
// A non-empty ref is returned as is, otherwise the current ref is used.
func getRemoteRef(gitroot, ref, remote string) (string, string, error) {
	remoteURL, err := gitw.RemoteURL(gitroot, remote)
	if err != nil || ref != "" {
		return remoteURL, ref, err
	}

	ref, err = gitw.CurrentRef(gitroot)
	return remoteURL, ref, err
}

// scpRegex matches SCP-like remotes, e.g. git@github.com:owner/repo.git (user optional)
var scpRegex = regexp.MustCompile(`^(?:([^@/:\s]+)@)?([\w.-]+):(.*)$`)

// parseRepository parses the host and repository (username or organization and repository name) from a remote string
func parseRepository(remote string) (host string, repo string, err error) {
	// Absolute paths (including Windows drive paths, which url.Parse reads as a scheme) are local.
	if filepath.IsAbs(remote) {
		return "", cleanRepo(filepath.ToSlash(remote)), nil
	}

	// SCP-like syntax: git@host:owner/repo. Without a user, the path must look like
	// owner/repo, so URLs ("://"), "user:pw@host:..." and "mailto:x" fall through.
	if m := scpRegex.FindStringSubmatch(remote); m != nil &&
		(m[1] != "" || (strings.Contains(m[3], "/") && !strings.HasPrefix(m[3], "//") && !strings.Contains(m[3], "@"))) {
		return m[2], cleanRepo(m[3]), nil
	}

	u, err := url.Parse(remote)
	if err != nil {
		return "", "", fmt.Errorf("unable to parse remote url: %w", err)
	}
	if u.Opaque != "" {
		return "", "", errors.New("unable to parse remote url: unsupported format")
	}
	return u.Host, cleanRepo(u.Path), nil
}

// cleanRepo normalizes a remote path to owner/repo form
func cleanRepo(p string) string {
	return strings.TrimPrefix(strings.TrimSuffix(path.Clean(p), ".git"), "/")
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
