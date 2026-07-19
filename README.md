# git-open

`git open` opens the Git repository in the web browser.

## Usage

Open the root of a repository.

```console
$ git open
```

Open a specific file or folder of a repository.

```console
$ git open main.go
```

Open a specific commit of a repository.

```console
$ git open 7605d91
```

Open a specific line, or range of lines, of a file.

```console
$ git open main.go:42
$ git open main.go:42-50
```

Open a different repository than `cwd`.

```console
$ git -C ~/src/my-repo open
```

### Providers

By default, three providers [github.com](https://github.com), [gitlab.com](https://gitlab.com) and [bitbucket.org](https://bitbucket.org) are supported.

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

## Installation

Install with `brew`.

```console
$ brew install arbourd/tap/git-open
```

Install with `go install`.

```console
$ go install github.com/arbourd/git-open@latest
```
