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
  -r, --ref <ref>      Open at a branch, tag, or commit instead of the current branch
  -R, --remote <name>  Use the named remote instead of the default
  -n, --no-browser     Print the URL instead of opening it
  -h, --help           Show this help message
  -v, --version        Show version`

func main() {
	opts, err := processArgs(os.Args[1:], os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
	if opts.done {
		return
	}

	url, err := open.GetURL(opts.arg, opts.ref, opts.remote)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if opts.noBrowser {
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

type options struct {
	arg       string
	ref       string
	remote    string
	noBrowser bool
	done      bool // help or version was requested
}

// processArgs parses args, printing help or the version to stdout if asked.
// Help and version win wherever they appear before "--", which makes later arguments positional.
func processArgs(args []string, stdout io.Writer) (options, error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			break
		}
		switch a {
		case "-r", "--ref", "-R", "--remote":
			// The value is not a flag, unless it is "--": that is a missing value, reported below.
			if i+1 < len(args) && args[i+1] != "--" {
				i++
			}
		case "-h", "--help":
			_, err := fmt.Fprintln(stdout, usage)
			return options{done: true}, err
		case "-v", "--version":
			_, err := fmt.Fprintf(stdout, "git-open %s\n", Version)
			return options{done: true}, err
		}
	}

	var opts options
	var err error
	positional := false
	count := 0
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case positional || !strings.HasPrefix(a, "-"):
			if count++; count > 1 {
				return options{}, fmt.Errorf("too many arguments: %q\n\n%s", a, usage)
			}
			opts.arg = a
		case a == "--":
			positional = true
		case a == "-r" || a == "--ref" || strings.HasPrefix(a, "--ref="):
			if opts.ref, err = optionValue(args, &i); err != nil {
				return options{}, err
			}
		case a == "-R" || a == "--remote" || strings.HasPrefix(a, "--remote="):
			if opts.remote, err = optionValue(args, &i); err != nil {
				return options{}, err
			}
		case a == "-n" || a == "--no-browser":
			opts.noBrowser = true
		default:
			return options{}, fmt.Errorf("unknown option %q\n\n%s", a, usage)
		}
	}
	return opts, nil
}

// optionValue returns the value of the option at args[*i], from "--name=value" or the next argument (advancing *i).
// An empty value, or one starting with "-", means the value is missing: it is almost certainly the next flag.
// This also rules out the rare remote name that starts with "-".
func optionValue(args []string, i *int) (string, error) {
	name, value, hasValue := strings.Cut(args[*i], "=")
	if !hasValue && *i+1 < len(args) {
		*i++
		value = args[*i]
	}
	if value == "" || strings.HasPrefix(value, "-") {
		return "", fmt.Errorf("option %q requires a value\n\n%s", name, usage)
	}
	return value, nil
}
