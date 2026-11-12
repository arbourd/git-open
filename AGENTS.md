# AGENTS.md

`git-open` opens the current Git repository (or a specific file, folder, or commit) in the browser. Installed as `git-open` on PATH; invoked as `git open [arg]`.

## Commands

```sh
go build .            # produces ./git-open binary
go vet ./...
go test -v -race ./...
golangci-lint run ./...
govulncheck ./...
actionlint            # local only; not run in CI
```

## Layout

```
main.go               CLI entry point — processArgs (flags, usage), Version (set by goreleaser ldflags)
main_test.go          unit test for processArgs (help/version, `-n`/`--no-browser`, unknown option, too many arguments, `--`)
man/git-open.1        Man page (must be updated whenever flags or usage change); shipped in release archives
.goreleaser.yml       Release archives (ship LICENSE*, README*, man/*); ldflags `-s -w -X main.Version`
.golangci.yml         golangci-lint config (errcheck excludes `(*os.File).Close`)
.github/workflows/ci.yml  test, lint, vulncheck; snapshot on branches; release on `v*` tags
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
main() → processArgs → GetURL(arg) → parseType → parsePath (→ stripLine) / getRemoteRef → parseRepository → Provider.{Commit,Path,Root}URL + Provider.lineAnchor
```

0. `processArgs`: handles flags (see CLI flags below); `--` makes later args positional.
1. Find repo root: `gitw.Toplevel` → falls back to `gitw.AbsoluteGitDir` (bare repos, worktrees).
2. Classify `arg`: empty → `Root`; 7–64 hex chars → `Commit`; anything else → `Path`. If `Commit` exists on disk as a file, reclassify to `Path`.
3. Resolve `Path` to a repo-relative slash path via `parsePath`, which first splits a trailing `:line`/`:start-end` with `stripLine` (only if the literal arg doesn't exist on disk; dropped for directories). Missing, outside-repo, or file-ancestor paths (`ancestorIsFile`, unifying POSIX `ENOTDIR` and Windows `ERROR_PATH_NOT_FOUND`) silently fall back to the root URL.
4. Get remote URL + current ref via `getRemoteRef` (delegates to `gitw`).
5. Parse host + repo path with `parseRepository` → delegates to `git-get`'s `get.ParseURL`.
6. Match provider by exact host comparison: `defaultProviders` first, then `fromConfig()` results (both via `Providers()`). URL patterns: `{BaseURL}/{repo}[/{CommitPrefix|PathPrefix}/{ref/sha}[/{path}]]`; all URLs run through `escapePath` (`url.PathEscape`, unescape `/`, trim trailing `/`).
7. For `Path` with lines, append `Provider.lineAnchor(start, end)` using `lineFormatRange` (`%d` verbs, converted from raw `%l` by `parseRawLineFormat`); uses single-verb `lineFormat` (dropping `end`) when `end` is 0 or `lineFormatRange` is absent.
8. Build and return the URL. With `-n`/`--no-browser`, `main` prints it bare to stdout instead of calling `InBrowser`.

## Hard constraints

**Git operations:** use `gitw` (or `ldez/go-git-cmd-wrapper/v2` directly). Never use `exec.Command("git", ...)` raw.

**CLI flags:** `-n`/`--no-browser` prints only the URL to stdout and never opens the browser (a failed write exits 1). It can appear before or after the path. `-h`/`--help` and `-v`/`--version` exit 0 and win over `-n` and other args before `--`. Any new flag needs cases in `TestProcessArgs`. Unknown options and extra args are `error: <message>` + usage on stderr, exit 1 (`unknown option %q`, `too many arguments: %q`). Git intercepts `git open --help` (man page), so `-h` is the in-binary path. Update the `usage` const and `man/git-open.1` together.

**Core/CLI separation:** keep URL generation logic in `open/` decoupled from `main.go`.

**Error handling:** `fmt.Errorf("context: %w", err)`. Explicit errors: `"local remotes are not supported"`, `"unable to find provider for: <host>"`.

**Providers:** defaults are `github.com`, `gitlab.com`, `bitbucket.org`, `codeberg.org`, each with a `rawLineFormat` (fake `%l` verb) converted by `parseRawLineFormat`. Custom: `git config open.<https-url>.{commitprefix,pathprefix,lineformat}` (local wins over global). `commitprefix`/`pathprefix` are required and `BaseURL` must be `http`/`https` with a host, or the provider is skipped. `lineformat` is optional with zero to two `%l` verbs; a missing/invalid one drops only line anchors (stderr warning). `TestDefaultProviders` and `TestDefaultProvidersLineFormat` enforce this for defaults.

**Tests:** `TestGetURL` and `TestParsePath` (`open/open_test.go`) run against the real checked-out repo — don't move `open/` without updating them. `TestGetURLErrors`, `TestGetURLBareRepo`, `TestGetURLWorktree` use hermetic temp repos. `TestGetURLWindowsAbsolutePath` is Windows-only (drive-letter paths). `TestFromConfig`/`TestFromConfigLocal` isolate git config via `GIT_CONFIG_GLOBAL`. `gitw` wraps `go-git-cmd-wrapper`; empty `path` → cwd.

## Do not

- Use `exec.Command("git", ...)` directly — use `gitw` or `go-git-cmd-wrapper`.
- Add logic to `main.go` beyond arg parsing, printing the URL (`-n`/`--no-browser`), opening the browser, and error printing.
- Add a default provider without cases in `TestCommitURL`, `TestPathURL`, `TestLineAnchor`, `TestRootURL`, `TestDefaultProvidersLineFormat` — for user-specific instances prefer `git config` over a hardcoded addition.
- Hardcode path separators — use `filepath.ToSlash` when building URLs from filesystem paths.
