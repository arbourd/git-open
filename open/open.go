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
	"strings"

	"github.com/arbourd/git-open/gitw"
)

// Type represents the type of Git URL to open
type Type int

const (
	// Folder is a specific commit in the repository
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
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

// GetURL returns the URL to open based on the arg provided
func GetURL(arg string) (string, error) {
	gitdir, err := gitw.GitDir(".")
	if err != nil {
		return "", fmt.Errorf("not a git repository")
	}
	gitdir, _ = filepath.Abs(gitdir)
	gitroot := strings.TrimSuffix(gitdir, ".git")

	t := parseType(arg)
	if t == Path {
		arg, _ = parsePath(arg, gitroot)
	}

	remote, ref, err := getRemoteRef(gitdir)
	if err != nil {
		return "", err
	}
	host, repo := parseRemote(remote)

	// Find the provider by comparing hosts
	var p Provider
	for _, provider := range DefaultProviders {
		if strings.Contains(provider.BaseURL, host) {
			p = provider
			break
		}
	}

	// Error if provider not found
	if len(p.BaseURL) == 0 {
		return "", fmt.Errorf("unable to find provider for: \"%s\"", host)
	}

	var url string
	switch t {
	case Commit:
		url = p.CommitURL(repo, arg)
	case Path:
		url = p.PathURL(repo, ref, arg)
	case Root:
		url = p.RootURL(repo)
	}

	return url, nil
}

// parsePath returns a cleaned path if the file and folder exist and belong to the root Git repository
func parsePath(path, gitroot string) (string, error) {
	if path == "" {
		return "", nil
	}
	path = filepath.Clean(path)

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", err
	}
	path, _ = filepath.Abs(path)

	// Check if path shares Git root
	if !strings.HasPrefix(path, gitroot) {
		return "", fmt.Errorf("path does not contain gitroot: %s; %s", path, gitroot)
	}

	// Remove gitroot from absolute path, making a relative path from the Git root
	path = strings.Replace(path, gitroot, "", 1)

	// Convert all path seperators to `/` and trim trailing `/`
	path = strings.TrimPrefix(filepath.ToSlash(path), "/")
	return path, nil
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

// parseRemote parses the host and repository (username or organization and repository name) from a remote string
func parseRemote(remote string) (host, repo string) {
	u, _ := url.Parse(remote)
	repo = strings.TrimPrefix(strings.TrimSuffix(path.Clean(u.Path), ".git"), "/")
	return u.Host, repo
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
