package gitw

import (
	"strings"

	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
)

// GitDir checks if the directory is a Git repository,
// with the Git directory specified by path
//
// git -C path rev-parse --git-dir
func GitDir(path string) (string, error) {
	out, err := git.Raw("-C", func(g *types.Cmd) {
		g.AddOptions(path)
		g.AddOptions("rev-parse")
		g.AddOptions("--git-dir")
	})
	return strings.TrimSpace(out), err
}

// RemoteURL returns the URL of the remote,
// with the Git directory specified by path
//
// git -C path ls-remote --get-url
func RemoteURL(path string) (string, error) {
	out, err := git.Raw("-C", func(g *types.Cmd) {
		g.AddOptions(path)
		g.AddOptions("ls-remote")
		g.AddOptions("--get-url")
	})
	return strings.TrimSpace(out), err
}

// CurrentRef returns the current reference or branch name,
// with the Git directory specified by path
//
// git -C path rev-parse --abbrev-ref HEAD
func CurrentRef(path string) (string, error) {
	out, err := git.Raw("-C", func(g *types.Cmd) {
		g.AddOptions(path)
		g.AddOptions("rev-parse")
		g.AddOptions("--abbrev-ref")
		g.AddOptions("HEAD")
	})
	return strings.TrimSpace(out), err
}
