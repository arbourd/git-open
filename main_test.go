package main

import (
	"testing"
)

func TestProcessArgs(t *testing.T) {
	cases := map[string]struct {
		args        []string
		expectedArg string
		err         string
	}{
		"no argument": {
			args:        []string{"git-open"},
			expectedArg: "",
		},
		"one argument": {
			args:        []string{"git-open", "LICENSE"},
			expectedArg: "LICENSE",
		},
		"two arguments": {
			args:        []string{"git-open", "LICENSE", "README.md"},
			expectedArg: "",
			err:         "recieved 2 args, accepts 1",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			arg, err := processArgs(c.args)

			if err != nil && c.err == "" {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && len(c.err) > 0 {
				t.Fatalf("expected error:\n\t(GOT): nil\n\t(WNT): %s", c.err)
			} else if arg != c.expectedArg {
				t.Fatalf("unexpected arg:\n\t(GOT): %#v\n\t(WNT): %#v", arg, c.expectedArg)
			}
		})
	}
}
