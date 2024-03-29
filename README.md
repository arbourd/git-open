# git-open

`git-open` opens your browser at points it to the path you selected in the VCS.

## Usage

Open the root of a repository.

```console
$ git open
```

Open a specific file or folder of a repository.

```console
$ git open LICENSE
```

Open a specific commit of a repository.

```console
$ git open 7605d91
```

Open a different repository than `cwd`.

```console
$ git -C ~/src/my-repo open
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
