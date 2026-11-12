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
  -n, --no-browser  Print the URL instead of opening it
  -h, --help        Show this help message
  -v, --version     Show version`

func main() {
	arg, noBrowser, done, err := processArgs(os.Args[1:], os.Stdout)
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

	if noBrowser {
		if _, err := fmt.Println(url); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("Opening %s in your browser.\n", url)
	err = open.InBrowser(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to open in browser: %s\n", err)
		os.Exit(1)
	}
}

// processArgs returns the single path or commit argument and whether to skip the browser, or done after printing help or version.
// A "--" makes every later argument positional.
func processArgs(args []string, stdout io.Writer) (arg string, noBrowser, done bool, err error) {
	// Help and version win wherever they appear before "--".
	for _, a := range args {
		if a == "--" {
			break
		}
		switch a {
		case "-h", "--help":
			_, err = fmt.Fprintln(stdout, usage)
			return "", false, true, err
		case "-v", "--version":
			_, err = fmt.Fprintf(stdout, "git-open %s\n", Version)
			return "", false, true, err
		}
	}

	positional := false
	count := 0
	for _, a := range args {
		switch {
		case positional || !strings.HasPrefix(a, "-"):
			if count++; count > 1 {
				return "", false, false, fmt.Errorf("too many arguments: %q\n\n%s", a, usage)
			}
			arg = a
		case a == "--":
			positional = true
		case a == "-n" || a == "--no-browser":
			noBrowser = true
		default:
			return "", false, false, fmt.Errorf("unknown option %q\n\n%s", a, usage)
		}
	}
	return arg, noBrowser, false, nil
}
