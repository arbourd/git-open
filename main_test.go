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
		"ref short": {
			args: []string{"-r", "v1.0.0", "LICENSE"},
			want: options{arg: "LICENSE", ref: "v1.0.0"},
		},
		"ref long after path": {
			args: []string{"LICENSE", "--ref", "main"},
			want: options{arg: "LICENSE", ref: "main"},
		},
		"ref equals": {
			args: []string{"--ref=feature/x"},
			want: options{ref: "feature/x"},
		},
		"ref with no-browser": {
			args: []string{"-n", "--ref", "main"},
			want: options{ref: "main", noBrowser: true},
		},
		"ref missing value": {
			args:    []string{"--ref"},
			wantErr: `option "--ref" requires a value`,
		},
		"ref empty value": {
			args:    []string{"--ref="},
			wantErr: `option "--ref" requires a value`,
		},
		"ref empty argument": {
			args:    []string{"-r", ""},
			wantErr: `option "-r" requires a value`,
		},
		"ref value that is a flag": {
			args:    []string{"-r", "-n", "LICENSE"},
			wantErr: `option "-r" requires a value`,
		},
		"ref value that is double dash": {
			args:    []string{"-r", "--", "LICENSE"},
			wantErr: `option "-r" requires a value`,
		},
		"ref value is not positional": {
			args:    []string{"-r", "main", "a", "b"},
			wantErr: "too many arguments",
		},
		"ref after double dash is a path": {
			args: []string{"--", "-r"},
			want: options{arg: "-r"},
		},
		"short remote": {
			args: []string{"-R", "fork", "LICENSE"},
			want: options{arg: "LICENSE", remote: "fork"},
		},
		"long remote after path": {
			args: []string{"LICENSE", "--remote", "fork"},
			want: options{arg: "LICENSE", remote: "fork"},
		},
		"long remote with equals": {
			args: []string{"--remote=fork"},
			want: options{remote: "fork"},
		},
		"remote with ref and no browser": {
			args: []string{"-n", "-r", "main", "-R", "fork"},
			want: options{ref: "main", remote: "fork", noBrowser: true},
		},
		"remote without value": {
			args:    []string{"-R"},
			wantErr: `option "-R" requires a value`,
		},
		"remote empty value": {
			args:    []string{"--remote="},
			wantErr: `option "--remote" requires a value`,
		},
		"remote value is a flag": {
			args:    []string{"-R", "-n"},
			wantErr: `option "-R" requires a value`,
		},
		"remote value is double dash": {
			args:    []string{"-R", "--", "x"},
			wantErr: `option "-R" requires a value`,
		},
		"remote value is not help or version": {
			args:    []string{"-R", "-v"},
			wantErr: `option "-R" requires a value`,
		},
		"remote value is double dash before help": {
			args:    []string{"-R", "--", "-h"},
			wantErr: `option "-R" requires a value`,
		},
		"ref value is not help or version": {
			args:    []string{"-r", "-h"},
			wantErr: `option "-r" requires a value`,
		},
		"repeated remote last wins": {
			args: []string{"-R", "a", "-R", "b"},
			want: options{remote: "b"},
		},
		"remote after double dash is a path": {
			args: []string{"--", "-R"},
			want: options{arg: "-R"},
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
