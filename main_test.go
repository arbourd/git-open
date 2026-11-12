package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestProcessArgs(t *testing.T) {
	cases := map[string]struct {
		args        []string
		expectedArg string
		expectedOut string
		noBrowser   bool
		wantDone    bool
		wantErr     string
	}{
		"no argument": {
			args: []string{},
		},
		"one argument": {
			args:        []string{"LICENSE"},
			expectedArg: "LICENSE",
		},
		"two arguments": {
			args:    []string{"LICENSE", "README.md"},
			wantErr: "too many arguments",
		},
		"no-browser short": {
			args:      []string{"-n"},
			noBrowser: true,
		},
		"no-browser long after path": {
			args:        []string{"LICENSE", "--no-browser"},
			expectedArg: "LICENSE",
			noBrowser:   true,
		},
		"no-browser before path": {
			args:        []string{"-n", "LICENSE"},
			expectedArg: "LICENSE",
			noBrowser:   true,
		},
		"no-browser repeated": {
			args:      []string{"-n", "--no-browser"},
			noBrowser: true,
		},
		"double dash makes no-browser a path": {
			args:        []string{"--", "-n"},
			expectedArg: "-n",
		},
		"no-browser then double dash then path": {
			args:        []string{"-n", "--", "--no-browser"},
			expectedArg: "--no-browser",
			noBrowser:   true,
		},
		"no-browser after double dash is an extra argument": {
			args:    []string{"--", "LICENSE", "-n"},
			wantErr: "too many arguments",
		},
		"no-browser with help": {
			args:        []string{"-n", "--help"},
			expectedOut: "Usage: git-open",
			wantDone:    true,
		},
		"no-browser with version": {
			args:        []string{"--no-browser", "-v"},
			expectedOut: "git-open dev\n",
			wantDone:    true,
		},
		"no-browser with unknown option": {
			args:    []string{"-n", "--foo"},
			wantErr: `unknown option "--foo"`,
		},
		"no-browser with two paths": {
			args:    []string{"-n", "a", "b"},
			wantErr: "too many arguments",
		},
		"help": {
			args:        []string{"--help"},
			expectedOut: "Usage: git-open",
			wantDone:    true,
		},
		"short help after path": {
			args:        []string{"LICENSE", "-h"},
			expectedOut: "Usage: git-open",
			wantDone:    true,
		},
		"version": {
			args:        []string{"--version"},
			expectedOut: "git-open dev\n",
			wantDone:    true,
		},
		"short version": {
			args:        []string{"-v"},
			expectedOut: "git-open dev\n",
			wantDone:    true,
		},
		"unknown option": {
			args:    []string{"--foo"},
			wantErr: `unknown option "--foo"`,
		},
		"unknown short option": {
			args:    []string{"-x"},
			wantErr: `unknown option "-x"`,
		},
		"double dash makes path positional": {
			args:        []string{"--", "-v"},
			expectedArg: "-v",
		},
		"help wins over unknown option": {
			args:        []string{"--foo", "--help"},
			expectedOut: "Usage: git-open",
			wantDone:    true,
		},
		"version wins over too many arguments": {
			args:        []string{"a", "b", "-v"},
			expectedOut: "git-open dev\n",
			wantDone:    true,
		},
		"help after double dash is a path": {
			args:        []string{"--", "--help"},
			expectedArg: "--help",
		},
		"double dash alone": {
			args: []string{"--"},
		},
		"double dash with two paths": {
			args:    []string{"--", "a", "b"},
			wantErr: "too many arguments",
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			arg, noBrowser, done, err := processArgs(c.args, &out)

			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("unexpected error:\n\t(GOT): %v\n\t(WNT): %q", err, c.wantErr)
				}
				if !strings.Contains(err.Error(), usage) {
					t.Fatalf("error does not include usage:\n\t(GOT): %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error:\n\t(GOT): %#v\n\t(WNT): nil", err)
			}
			if arg != c.expectedArg {
				t.Fatalf("unexpected arg:\n\t(GOT): %#v\n\t(WNT): %#v", arg, c.expectedArg)
			}
			if noBrowser != c.noBrowser {
				t.Fatalf("unexpected noBrowser:\n\t(GOT): %v\n\t(WNT): %v", noBrowser, c.noBrowser)
			}
			if done != c.wantDone {
				t.Fatalf("unexpected done:\n\t(GOT): %v\n\t(WNT): %v", done, c.wantDone)
			}
			if !strings.Contains(out.String(), c.expectedOut) {
				t.Fatalf("unexpected output:\n\t(GOT): %q\n\t(WNT): contains %q", out.String(), c.expectedOut)
			}
		})
	}
}
