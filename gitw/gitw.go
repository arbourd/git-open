package gitw

import (
	"strings"

	"github.com/ldez/go-git-cmd-wrapper/v2/config"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/revparse"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
)

// cwd returns an option that runs the git command from the given path.
// When path is empty the option is a no-op, leaving git to use the current directory.
func cwd(path string) types.Option {
	if path == "" {
		return func(*types.Cmd) {}
	}

	return func(g *types.Cmd) {
		g.AddBaseOptions("-C")
		g.AddBaseOptions(path)
	}
}

// AbsoluteGitDir returns the absolute path to the Git directory,
// with the Git directory specified by path
//
// git -C path rev-parse --absolute-git-dir
func AbsoluteGitDir(path string) (string, error) {
	out, err := git.RevParse(cwd(path), revparse.AbsoluteGitDir)
	return strings.TrimSpace(out), err
}

// Toplevel returns the root of the working tree,
// with the Git directory specified by path
//
// git -C path rev-parse --show-toplevel
func Toplevel(path string) (string, error) {
	out, err := git.RevParse(cwd(path), revparse.ShowToplevel)
	return strings.TrimSpace(out), err
}

// RemoteURL returns the URL of the remote,
// with the Git directory specified by path
//
// git -C path ls-remote --get-url
func RemoteURL(path string) (string, error) {
	out, err := git.Raw("ls-remote", cwd(path), func(g *types.Cmd) {
		g.AddOptions("--get-url")
	})
	return strings.TrimSpace(out), err
}

// CurrentRef returns the current reference or branch name,
// with the Git directory specified by path.
// Falls back to the full commit SHA when in detached HEAD state.
//
// git -C path rev-parse --abbrev-ref HEAD
// git -C path rev-parse HEAD
func CurrentRef(path string) (string, error) {
	out, err := git.RevParse(cwd(path), revparse.AbbrevRef(""), revparse.Args("HEAD"))
	if err != nil {
		return "", err
	}

	ref := strings.TrimSpace(out)
	if ref == "HEAD" {
		out, err = git.RevParse(cwd(path), revparse.Args("HEAD"))
		return strings.TrimSpace(out), err
	}

	return ref, nil
}

// ConfigGetRegexp returns trimmed git config output for all keys matching pattern.
// Reads all config scopes (system, global, local) with local taking precedence over global for the same key.
func ConfigGetRegexp(pattern string) string {
	out, _ := git.Config(config.GetRegexp(pattern, ""))
	return strings.TrimSpace(out)
}
