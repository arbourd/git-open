---
name: git-open
description: Get the web URL for, or open in the browser, the current Git repo, a file, folder, commit, or line with `git open`. Use when the user asks for a link to code, or to open or share the repo, a file, a line, or a commit, e.g. "link me to line 42 of main.go" or "give me a permalink". Not for PRs, issues, CI, blame, or reading local files.
---

# git-open

`git open [options] [path|commit]` prints or opens the web page for the current repo, built from the remote URL (the default remote, or `-R <name>`).

## Rules

1. **Match the flag to the request.** For a browser request, run without `-n`. For a link, URL, permalink, or sharing request, pass `-n` (`--no-browser`) and return the printed URL. If unclear, print with `-n` and offer to open it.
2. **Return the printed URL unchanged.** Provider path prefixes and line anchors differ (GitLab `/-/tree/main/a.go#L3-5`, Bitbucket `/src/main/a.go#lines-3:5`), so a hand-built URL may be wrong.
3. **Run in the repo.** For another directory, use `git -C <dir> open` with the same arguments.

## Recipes

```sh
git open -n main.go:42   # "link me to line 42 of main.go": print the URL
# https://github.com/owner/repo/tree/feature-x/main.go#L42

git open main.go         # "open main.go in the browser": launch it
```

With no argument it targets the repo root, and a path may be a file or a folder. Other forms, shown without a flag (add `-n` to print the URL instead of opening the browser):

```sh
git open main.go:42         # line; `main.go:42-50` for a range
git open 7605d91            # commit
git open -r main README.md  # link to main, for an unpushed branch; any branch or tag works
git open -R fork main.go    # build the URL from remote `fork`
git open -- -odd-name       # `--` makes later args paths, even with a leading `-`
```

**Stable link (permalink):** a plain path uses the current branch, so the link changes as the branch moves. Pin it to a commit with `-r`:

```sh
git open -n -r "$(git rev-parse HEAD)" main.go:42
# https://github.com/owner/repo/tree/<full-sha>/main.go#L42
```

## Gotchas

- **Ref:** the current branch, or the full SHA when HEAD is detached. `-r` replaces it and is used literally, so `origin/main` may 404; use the branch name on the host, such as `main`.
- **Commits:** a commit argument is a lowercase 7–64 character hex SHA, unless a file by that name exists, and can't be combined with `-r`. `HEAD`, branch names, tags, and uppercase SHAs are treated as paths; use `-r <ref>` for a branch or tag, and `git rev-parse` to resolve `HEAD`.
- **Paths and lines:** paths are relative to the current directory or absolute. Missing paths are kept because they may exist at the ref. A path outside the repo, or `.` at the repo root, gives the ref's tree URL. A `:line` suffix is ignored for directories and treated literally if that full path exists.
- **Remote state:** URLs use local Git state and are not checked against the host. An unpushed ref or path absent at that ref can 404; line numbers refer to the pushed file and local edits may shift them.

## Self-hosted providers

An unknown host errors with `unable to find provider`. Add one with git config; `commitprefix` and `pathprefix` are required, `lineformat` is optional (`%l` is a line number). Suggest the commands to the user rather than changing their config:

```sh
git config --global open.https://git.example.com.commitprefix commit
git config --global open.https://git.example.com.pathprefix tree
git config --global open.https://git.example.com.lineformat "L%l-L%l"
```

## More

`git open -h` prints usage; `git open --help` opens the man page.
