package main

import (
	"testing"
)

func TestProcessArgs(t *testing.T) {
	cases := map[string]struct {
		args        []string
		expectedArg string
		wantErr     bool
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
			wantErr:     true,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			arg, err := processArgs(c.args)

			if err != nil && !c.wantErr {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			} else if err == nil && c.wantErr {
				t.Fatalf("expected error:\n\t(GOT): nil\n")
			} else if arg != c.expectedArg {
				t.Fatalf("unexpected arg:\n\t(GOT): %#v\n\t(WNT): %#v", arg, c.expectedArg)
			}
		})
	}
}
