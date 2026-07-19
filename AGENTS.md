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
main_test.go          unit test for processArgs
open/
  open.go             GetURL, InBrowser, path/type/repo parsing
  provider.go         Provider struct, defaultProviders, Providers/fromConfig, line-format parsing (parseRawLineFormat, validateLineFormat)
  open_test.go        integration tests (run against the real checked-out repo)
  provider_test.go    unit + integration tests for provider loading
gitw/
  gitw.go             thin git wrappers: Toplevel, AbsoluteGitDir, RemoteURL, CurrentRef, ConfigGetRegexp
  gitw_test.go        unit tests for gitw
```

## Data flow

```
main() → GetURL(arg) → parseType → parsePath (→ stripLine) / getRemoteRef → parseRepository → Provider.{Commit,Path,Root}URL + Provider.lineAnchor
```

1. Find repo root: `gitw.Toplevel` → falls back to `gitw.AbsoluteGitDir` (bare repos, worktrees).
2. Classify `arg`: empty → `Root`; 7–64 hex chars → `Commit`; anything else → `Path`. If `Commit` exists on disk as a file, reclassify to `Path`.
3. Resolve `Path` arg to a repo-relative slash path via `parsePath`, which runs `stripLine` first to split a trailing `:line`/`:start-end` suffix (only treated as a line spec if the literal arg doesn't exist on disk; dropped for directories). Failures — missing, outside repo, or an ancestor being a file (`ancestorIsFile`, unifying POSIX `ENOTDIR` and Windows `ERROR_PATH_NOT_FOUND`) — silently fall back to root URL.
4. Get remote URL + current ref via `getRemoteRef` (delegates to `gitw`).
5. Parse host + repo path with `parseRepository` → delegates to `git-get`'s `get.ParseURL`.
6. Match provider by exact host comparison: `defaultProviders` first, then `fromConfig()` results (both via `Providers()`). URL patterns: `{BaseURL}/{repo}[/{CommitPrefix|PathPrefix}/{ref/sha}[/{path}]]`; all URLs run through `escapePath` (`url.PathEscape`, unescape `/`, trim trailing `/`).
7. For `Path` with line numbers, append `Provider.lineAnchor(start, end)`, formatting `lineFormat`/`lineFormatRange`'s `%d` verb(s) (converted from raw `%l` by `parseRawLineFormat`); falls back to the single-verb `lineFormat` (dropping `end`) when `end` is 0 or `lineFormatRange` is absent.
8. Build and return the URL.

## Hard constraints

**Git operations:** use `gitw` (or `ldez/go-git-cmd-wrapper/v2` directly). Never use `exec.Command("git", ...)` raw.

**Core/CLI separation:** keep URL generation logic in `open/` decoupled from `main.go`.

**Error handling:** `fmt.Errorf("context: %w", err)`. Explicit errors: `"local remotes are not supported"`, `"unable to find provider for: <host>"`.

**Providers:** default providers are `github.com`, `gitlab.com`, `bitbucket.org`, each with a `rawLineFormat` (fake `%l` verb) converted to `lineFormat`/`lineFormatRange` by `parseRawLineFormat`. Custom providers: `git config open.<https-url>.{commitprefix,pathprefix,lineformat}` (local repo config wins over global). `commitprefix`/`pathprefix` are required and `BaseURL` must be `http`/`https` with a non-empty host — any violation skips the provider entirely. `lineformat` is optional: zero, one, or two `%l` verbs are valid (more, or any other verb, is not) — a missing/invalid `lineformat` drops only that provider's line-anchor support (stderr warning), not the provider itself. `TestDefaultProviders` and `TestDefaultProvidersLineFormat` enforce this for built-in defaults.

**Tests:** `TestGetURL` and `TestParsePath` in `open/open_test.go` run against the real checked-out repo — do not move `open/` without updating them. `TestGetURLErrors`, `TestGetURLBareRepo`, and `TestGetURLWorktree` spin up hermetic temp repos. `TestGetURLWindowsAbsolutePath` is Windows-only (`t.Skip` elsewhere) and covers drive-letter path parsing. `TestFromConfig` and `TestFromConfigLocal` isolate global/local git config via `GIT_CONFIG_GLOBAL`. `gitw` functions are thin wrappers over `go-git-cmd-wrapper`: `Toplevel`, `AbsoluteGitDir`, `RemoteURL`, `CurrentRef`, `ConfigGetRegexp`; empty `path` arg → cwd.

## Do not

- Use `exec.Command("git", ...)` directly — use `gitw` or `go-git-cmd-wrapper`.
- Add logic to `main.go` beyond arg parsing and error printing.
- Add a default provider without cases in `TestCommitURL`, `TestPathURL`, `TestLineAnchor`, `TestRootURL`, `TestDefaultProvidersLineFormat` — for user-specific instances prefer `git config` over a hardcoded addition.
- Hardcode path separators — use `filepath.ToSlash` when building URLs from filesystem paths.
