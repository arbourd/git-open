# AGENTS.md

`git-open` opens the current Git repository (or a specific file, folder, or commit) in the browser. Installed as `git-open` on PATH; invoked as `git open [arg]`.

## Commands

```sh
go test -race ./...                                         # required project check
golangci-lint run ./...
govulncheck ./...
```

Run the project check before reporting completion. Run `actionlint` when changing `.github/workflows`, `mandoc -T lint -W warning man/git-open.1` when changing the man page, and `bunx --package renovate renovate-config-validator --strict` when changing `.github/renovate.json`. These local checks are not run in CI.

Always add `-n` when running `git open`, `git-open` or `./git-open`; without it, it launches a real browser. Only run it without `-n` when deliberately testing the browser launch, and ask first.

## CLI flags

- Before `--`, help and version requests exit 0 and take precedence over other arguments, subject to any tool-specific flag checked first.
- `--` ends option parsing; later arguments are positional input.
- Invalid input prints `error: <message>` plus usage to stderr and exits 1.
- Git intercepts `git open --help` to open the man page; use `-h` for in-binary help.
- `-r`/`--ref <ref>` and `-R`/`--remote <name>` (also `--ref=<ref>` and `--remote=<name>`) take a value; an empty value or one starting with `-` is a missing-value error, and an unknown remote is a `no such remote` error.
- A non-empty `--ref` replaces the current ref and skips `CurrentRef`; `Root` becomes `Path` with an empty path (the tree URL), and a `Commit` is an error.
- Value flags share `optionValue` and are listed in the help/version pre-scan, which skips their value; add each new value flag to both, or `-x -v` prints the version. `-R -v` and `-R -- -h` are missing-value errors, not help or version.
- `-n`/`--no-browser` before or after the path never opens the browser: it prints only the URL to stdout, and a failed write exits 1.
- `-h`/`--help` and `-v`/`--version` exit 0 and win over `-n` and other arguments before `--`.
- Add cases to `TestProcessArgs` for every new flag.
- When flags, output, or behavior change, update `usage` in `main.go`, `man/git-open.1`, `README.md`, and `skills/git-open/SKILL.md` together.
## Hard constraints

- **Git operations:** use `gitw` (or `ldez/go-git-cmd-wrapper/v2` directly); never `exec.Command("git", ...)` in non-test code. The repo root is `gitw.Toplevel`, falling back to `gitw.AbsoluteGitDir` (bare repos, worktrees); keep the fallback.
- **Error handling:** never put the remote in the parse error; it can carry `user:pw@`. `no such remote` is unwrapped on purpose because git's own error is just `exit status 2`.
- **Providers:** match by exact host: `defaultProviders` first, then git config.
  - Custom providers use `git config open.<https-url>.{commitprefix,pathprefix,lineformat}` (local wins over global). `commitprefix`/`pathprefix` are required, and `BaseURL` must be `http`/`https` with a host, or the provider is skipped.
  - `lineformat` is optional; a missing or invalid value drops only line anchors, with a stderr warning.
  - A new default provider needs URL, line-anchor, and line-format test cases; user-specific instances go in git config.
- **Argument parsing:** `arg` is `Root` if empty, `Commit` if 7–64 hex chars (unless a file by that name exists), else `Path`. A trailing `:line`/`:start-end` is split only if the literal arg does not exist on disk, and the line is dropped for directories. Paths missing from disk (including under a file) are kept as `tree/<ref>/<path>`; only paths outside the repo (and `.` at the root) silently give the ref's tree URL, with or without `--ref`.
- **Remote parsing:** SCP `user@` is optional. Without a user, the path must contain `/`, not start with `//`, and not contain `@`, so `https://…`, `mailto:x` and `user:pw@host:…` fall through to `url.Parse`; opaque URLs error.

## Testing

- Tests that use the real checkout (`TestGetURL`, `TestParsePath`) must be updated if `open/` moves; the others use hermetic temp repos.
- Keep Windows-only cases behind their guard.
- A new remote format needs a `TestParseRepository` case.
- Tests that read git config isolate it via `GIT_CONFIG_GLOBAL`.

## Conventions

- Keep `main.go` to arg parsing, printing the URL, opening the browser, and error printing; URL generation lives in `open/`.
- Build URLs from filesystem paths with `filepath.ToSlash`.
