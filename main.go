package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arbourd/git-open/open"
)

// Version is set via -ldflags at build time.
var Version = "dev"

const usage = `Usage: git-open [options] [path|commit]

Open the repository, a file, folder, or commit in the browser.

Arguments:
  path|commit  A file, folder, or commit hash (default: repository root)
               Files accept a :line or :start-end suffix

Options:
  -h, --help     Show this help message
  -v, --version  Show version`

func main() {
	arg, done, err := processArgs(os.Args[1:], os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	if done {
		return
	}

	url, err := open.GetURL(arg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Opening %s in your browser.\n", url)
	err = open.InBrowser(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to open in browser: %s\n", err)
		os.Exit(1)
	}
}

// processArgs returns the single path or commit argument, or done after printing help or version.
// A "--" makes every later argument positional.
func processArgs(args []string, stdout io.Writer) (arg string, done bool, err error) {
	// Help and version win wherever they appear before "--".
	for _, a := range args {
		if a == "--" {
			break
		}
		switch a {
		case "-h", "--help":
			_, err = fmt.Fprintln(stdout, usage)
			return "", true, err
		case "-v", "--version":
			_, err = fmt.Fprintf(stdout, "git-open %s\n", Version)
			return "", true, err
		}
	}

	positional := false
	count := 0
	for _, a := range args {
		switch {
		case positional || !strings.HasPrefix(a, "-"):
			if count++; count > 1 {
				return "", false, fmt.Errorf("too many arguments: %q\n\n%s", a, usage)
			}
			arg = a
		case a == "--":
			positional = true
		default:
			return "", false, fmt.Errorf("unknown option %q\n\n%s", a, usage)
		}
	}
	return arg, false, nil
}
