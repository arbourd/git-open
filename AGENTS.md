# AGENTS.md

`git-open` opens the current Git repository (or a specific file, folder, or commit) in the browser. Installed as `git-open` on PATH; invoked as `git open [arg]`.

## Commands

```sh
go build .            # produces ./git-open binary
go vet ./...
go test -v -race ./...
```

## Layout

```
main.go               CLI entry point — arg parsing only
open/
  open.go             GetURL, InBrowser, path/type/repo parsing
  provider.go         Provider struct, DefaultProviders, LoadProviders
  open_test.go        integration tests (run against the real checked-out repo)
  provider_test.go    unit + integration tests for provider loading
gitw/
  gitw.go             thin git wrappers: Toplevel, AbsoluteGitDir, RemoteURL, CurrentRef
  gitw_test.go        unit tests for gitw
```

## Data flow

```
main() → GetURL(arg) → parseType → parsePath / getRemoteRef → parseRepository → Provider.{Commit,Path,Root}URL
```

1. Find repo root: `gitw.Toplevel` → falls back to `gitw.AbsoluteGitDir` (bare repos, worktrees).
2. Classify `arg`: empty → `Root`; 7–64 hex chars → `Commit`; anything else → `Path`. If `Commit` exists on disk as a file, reclassify to `Path`.
3. Resolve `Path` arg to a repo-relative slash path via `parsePath`; if it fails (non-existent, outside repo), silently falls back to root URL.
4. Get remote URL + current ref via `getRemoteRef` (delegates to `gitw`).
5. Parse host + repo path with `parseRepository` → delegates to `git-get`'s `get.ParseURL`.
6. Match provider by exact host comparison: `DefaultProviders` first, then `LoadProviders()` results. URL patterns: `{BaseURL}/{repo}[/{CommitPrefix|PathPrefix}/{ref/sha}[/{path}]]`; all URLs run through `escapePath` (`url.PathEscape`, unescape `/`, trim trailing `/`).
7. Build and return the URL.

## Hard constraints

**Git operations:** use `gitw` (or `ldez/go-git-cmd-wrapper/v2` directly). Never use `exec.Command("git", ...)` raw.

**Core/CLI separation:** keep URL generation logic in `open/` decoupled from `main.go`.

**Error handling:** `fmt.Errorf("context: %w", err)`. Explicit errors: `"local remotes are not supported"`, `"unable to find provider for: <host>"`.

**Providers:** default providers are `github.com`, `gitlab.com`, `bitbucket.org`. Custom providers can be added via `git config --global open.<https-url>.{commitprefix,pathprefix}`; non-http/https or missing host → stderr warning + skip. `BaseURL` must be an `http`/`https` URL with a non-empty host — `TestDefaultProviders` enforces this for built-in defaults.

**Tests:** `TestGetURL` and `TestParsePath` in `open/open_test.go` run against the real checked-out repo — do not move `open/` without updating them. `TestLoadProviders` isolates global git config via `GIT_CONFIG_GLOBAL`. `gitw` functions are thin wrappers over `go-git-cmd-wrapper`: `Toplevel`, `AbsoluteGitDir`, `RemoteURL`, `CurrentRef`; empty `path` arg → cwd.

## Do not

- Use `exec.Command("git", ...)` directly — use `gitw` or `go-git-cmd-wrapper`.
- Add logic to `main.go` beyond arg parsing and error printing.
- Add a default provider without cases in `TestCommitURL`, `TestPathURL`, `TestRootURL` — for user-specific instances prefer `git config` over a hardcoded addition.
- Hardcode path separators — use `filepath.ToSlash` when building URLs from filesystem paths.
