# Plan: Fix Local Path Remote Bug

This plan addresses the issue where `git-open` generates invalid URLs when the Git remote is a local filesystem path (common in local or bare clones).

## 1. Problem Analysis
When a Git repository's remote is a local path (e.g., `/Users/dylan/repo`), the `parseRepository` function returns an empty `host`. 

The current matching logic in `GetURL` is:
```go
if strings.Contains(provider.BaseURL, host) { ... }
```
Since `strings.Contains(any, "")` is always true, the first provider (GitHub) is selected by default, and the local absolute path is used as the repository name, resulting in invalid URLs like `https://github.com/Users/dylan/repo`.

## 2. Objectives
- **Prevent Invalid Matches**: Stop matching providers if the host is empty.
- **Improved Provider Matching**: Use robust host matching instead of simple substring containment.
- **Informative Errors**: Provide a clear error message when a local remote is detected, as these cannot be opened in a browser via a provider.

## 3. Changes

### `open/open.go`
- **Update `GetURL`**:
    - Check if `host` is empty after calling `parseRepository`. If empty, return a descriptive error (e.g., "local remotes are not supported").
    - Refactor the provider matching loop to parse the `provider.BaseURL` and perform a strict comparison of the hostname to ensure accuracy.

## 4. Verification Plan

### Automated Tests
- Update `TestParseRepository` in `open/open_test.go` to include a local path case.
- Update `TestGetURL` to verify that an empty host results in an error.

### Manual Verification
1. **Local Remote**:
   - `git clone . ../local-test`
   - `cd ../local-test && git open`
   - Verify it returns an error instead of a broken URL.
2. **Standard Remote**:
   - Verify GitHub/GitLab/Bitbucket repos continue to work as expected.
