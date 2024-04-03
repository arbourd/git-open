package main

import (
	"fmt"
	"os"

	"github.com/arbourd/git-open/open"
)

func main() {
	arg, err := processArgs(os.Args)
	if err != nil {
		fmt.Printf("error: \"%s\"\n", err)
		os.Exit(1)
	}

	url, err := open.GetURL(arg)
	if err != nil {
		fmt.Printf("error: \"%s\"\n", err)
		os.Exit(1)
	}

	fmt.Printf("Opening %s in your browser.\n", url)
	err = open.InBrowser(url)
	if err != nil {
		fmt.Printf("error: unable to open in browser: \"%s\"\n", err)
		os.Exit(1)
	}
}

// processArgs returns a single argument or returns an error if more than 1 argument is provided
func processArgs(args []string) (string, error) {
	switch len(args) {
	case 1:
		return "", nil
	case 2:
		return args[1], nil
	default:
		return "", fmt.Errorf("recieved %d args, accepts 1", len(args)-1)
	}
}
