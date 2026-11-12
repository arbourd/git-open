# git-open

`git open` opens the Git repository in the web browser.

## Installation

Install with `brew`.

```console
$ brew install arbourd/tap/git-open
```

Install with `go install`.

```console
$ go install github.com/arbourd/git-open@latest
```

## Usage

```console
$ git open [options] [path|commit]
```

```console
# Open the root of the repository
$ git open

# Open a file or folder
$ git open main.go

# Open a line, or a range of lines, of a file
$ git open main.go:42
$ git open main.go:42-50

# Open a commit
$ git open 7605d91

# Open a repository other than the current directory
$ git -C ~/src/my-repo open
```

### Options

| Option                | Description                                  |
| --------------------- | -------------------------------------------- |
| `-n`, `--no-browser`  | Print the URL instead of opening the browser |
| `-h`, `--help`        | Show usage                                   |
| `-v`, `--version`     | Show the version                             |
| `--`                  | Treat every later argument as a path         |

`--no-browser` can go before or after the path and prints only the URL, which is
useful in scripts.

```console
$ git open main.go -n
https://github.com/arbourd/git-open/tree/main/main.go
$ git open --no-browser 7605d91
https://github.com/arbourd/git-open/commit/7605d91
```

Use `--` to open a path that starts with a dash.

```console
$ git open -- -odd-file-name
```

### Providers

By default, four providers are supported: [github.com](https://github.com), [gitlab.com](https://gitlab.com), [bitbucket.org](https://bitbucket.org) and [codeberg.org](https://codeberg.org).

| Provider  | URL                     | Commit prefix | Path prefix | Line format   |
| --------- | ----------------------- | ------------- | ----------- | ------------- |
| GitHub    | `https://github.com`    | `commit`      | `tree`      | `L%l-L%l`     |
| GitLab    | `https://gitlab.com`    | `-/commit`    | `-/tree`    | `L%l-%l`      |
| Bitbucket | `https://bitbucket.org` | `commits`     | `src`       | `lines-%l:%l` |
| Codeberg  | `https://codeberg.org`  | `commit`      | `tree`      | `L%l-L%l`     |

To add custom Git providers and their URLs, set their values within the global `git config`.

```ini
[open "https://git.mydomain.dev"]
    commitprefix = commit
    pathprefix = tree
    lineformat = L%l-L%l
```

This can also be set using the `git` CLI.

```console
$ git config --global open.https://git.mydomain.dev.commitprefix commit
$ git config --global open.https://git.mydomain.dev.pathprefix tree
$ git config --global open.https://git.mydomain.dev.lineformat "L%l-L%l"
```

`commitprefix` and `pathprefix` are used to template the URI for your provider.

```go
fmt.Println(host + "/" + repository + "/" + commitprefix )
// https://git.mydomain.dev/<repository>/commit

fmt.Println(host + "/" + repository + "/" + pathprefix )
// https://git.mydomain.dev/<repository>/tree
```

`lineformat` templates the line anchor where `%l` denotes the line number. Up to two
`%l` may be provided, denoting the start and end lines. 3 or more `%l` or unsupported
Go-style verbs like `%v` will disable the line anchor templating and issue a warning.

```go
fmt.Sprintf(lineformat, startLine)          // one %l
fmt.Sprintf(lineformat, startLine, endLine) // two %l
// L42 or L42-L50
```
