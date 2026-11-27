# Contributing

Thanks for your interest. This repository is **tightly governed**: the `main` branch is
protected and **all** changes land through reviewed pull requests. Direct pushes to `main`
are not accepted from anyone, including maintainers.

## Ground rules

- **No direct commits to `main`.** Work on a branch and open a pull request.
- **All PRs require review and approval from a code owner** (see `CODEOWNERS`) and must pass
  all required status checks.
- **Signed commits** are required. Configure commit signing before contributing.
- **Code and comments must be in English.**
- **Conventional Commits** for commit messages and PR titles (`feat:`, `fix:`, `docs:`,
  `refactor:`, `test:`, `chore:`).
- **Self-contained PR descriptions.** Do not reference local or unversioned paths. The
  reviewer only has the diff and your description.

## Development workflow

1. Create a feature branch from `main`.
2. Make your change with tests.
3. Run locally and make sure these pass:
   ```sh
   gofmt -l .        # must print nothing
   go vet ./...
   go test ./... -race
   go build ./...
   ```
4. Open a PR using the template. Fill every section.

## Code quality expectations

- **Clean Architecture:** dependencies point inward; the domain does not depend on UI/infra
  details. Program to interfaces.
- **Clean Code:** small functions, intention-revealing names, explicit error handling with
  wrapping (`fmt.Errorf("...: %w", err)`), no duplication, tests for business rules.
- **Security:** validate all untrusted input, use parameterized queries, never commit
  secrets or sensitive engagement data.

## Scope

This project implements framework **infrastructure**. Pull requests that add weaponized
offensive logic (working exploits, remote-control agents, detection-evasion techniques) are
out of scope and will not be merged. Offensive module types remain interfaces with `// TODO`.

## Security issues

Do not open public issues for vulnerabilities. See `SECURITY.md`.
