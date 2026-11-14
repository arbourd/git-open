package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestProcessArgs(t *testing.T) {
	cases := map[string]struct {
		args        []string
		want        options
		expectedOut string
		wantErr     string
	}{
		"no argument": {
			args: []string{},
		},
		"one argument": {
			args: []string{"LICENSE"},
			want: options{arg: "LICENSE"},
		},
		"two arguments": {
			args:    []string{"LICENSE", "README.md"},
			wantErr: "too many arguments",
		},
		"no-browser short": {
			args: []string{"-n"},
			want: options{noBrowser: true},
		},
		"no-browser long after path": {
			args: []string{"LICENSE", "--no-browser"},
			want: options{arg: "LICENSE", noBrowser: true},
		},
		"no-browser before path": {
			args: []string{"-n", "LICENSE"},
			want: options{arg: "LICENSE", noBrowser: true},
		},
		"no-browser repeated": {
			args: []string{"-n", "--no-browser"},
			want: options{noBrowser: true},
		},
		"double dash makes no-browser a path": {
			args: []string{"--", "-n"},
			want: options{arg: "-n"},
		},
		"no-browser then double dash then path": {
			args: []string{"-n", "--", "--no-browser"},
			want: options{arg: "--no-browser", noBrowser: true},
		},
		"no-browser after double dash is an extra argument": {
			args:    []string{"--", "LICENSE", "-n"},
			wantErr: "too many arguments",
		},
		"no-browser with help": {
			args:        []string{"-n", "--help"},
			expectedOut: "Usage: git-open",
			want:        options{done: true},
		},
		"no-browser with version": {
			args:        []string{"--no-browser", "-v"},
			expectedOut: "git-open dev\n",
			want:        options{done: true},
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
			want:        options{done: true},
		},
		"short help after path": {
			args:        []string{"LICENSE", "-h"},
			expectedOut: "Usage: git-open",
			want:        options{done: true},
		},
		"version": {
			args:        []string{"--version"},
			expectedOut: "git-open dev\n",
			want:        options{done: true},
		},
		"short version": {
			args:        []string{"-v"},
			expectedOut: "git-open dev\n",
			want:        options{done: true},
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
			args: []string{"--", "-v"},
			want: options{arg: "-v"},
		},
		"help wins over unknown option": {
			args:        []string{"--foo", "--help"},
			expectedOut: "Usage: git-open",
			want:        options{done: true},
		},
		"version wins over too many arguments": {
			args:        []string{"a", "b", "-v"},
			expectedOut: "git-open dev\n",
			want:        options{done: true},
		},
		"help after double dash is a path": {
			args: []string{"--", "--help"},
			want: options{arg: "--help"},
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
			opts, err := processArgs(c.args, &out)

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
			if opts != c.want {
				t.Fatalf("unexpected options:\n\t(GOT): %#v\n\t(WNT): %#v", opts, c.want)
			}
			if c.expectedOut == "" && out.Len() != 0 {
				t.Fatalf("unexpected output:\n\t(GOT): %q\n\t(WNT): none", out.String())
			}
			if !strings.Contains(out.String(), c.expectedOut) {
				t.Fatalf("unexpected output:\n\t(GOT): %q\n\t(WNT): contains %q", out.String(), c.expectedOut)
			}
		})
	}
}
