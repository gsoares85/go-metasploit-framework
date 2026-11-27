# go-metasploit-framework

An educational re-implementation, in Go, of the **architecture** of a modular security
testing framework (inspired by the Metasploit Framework). The focus is the engineering of
the framework itself — interactive console, module system, workspaces, database, session
management and interfaces — not the production of weaponized offensive capabilities.

> ⚠️ **Intended use.** For learning and **authorized** security testing only: your own lab
> environments, penetration testing under contract, CTF competitions, and defensive
> research. Offensive module types (exploit/payload/post/encoder/nop/evasion) are provided
> as **interfaces and scaffolding only**. Do not use this software against systems you do not
> own or lack explicit written authorization to test. See [SECURITY.md](SECURITY.md).

## Status

Early development (pre-1.0). The foundational framework is being built sprint by sprint:
core, module engine, database & workspaces, reconnaissance modules, session/handler
management, and interfaces.

## Design principles

- **Security first** — untrusted input is validated, database access is parameterized, no
  secrets or sensitive engagement data are committed, concurrency is race-free.
- **Clean Architecture** — dependencies point inward; the domain does not depend on UI or
  infrastructure details. Everything is programmed against interfaces.
- **Clean Code** — small functions, intention-revealing names, explicit error handling,
  tests for business rules.

## Scope (roadmap)

Nothing in the table below is implemented yet — the repository is at the foundation
stage (a console entry point plus placeholder packages). The marks describe the
**planned scope** of each area, not its current status:

- ✅ **will be built** — neutral framework infrastructure.
- ⚠️ **scaffolding only** — interfaces and `// TODO`; no weaponized logic.

| Area | Planned scope |
|------|:-------------:|
| Core (framework object, datastore, events, jobs) | ✅ |
| Module engine (contract, loader, registry, search) | ✅ |
| Workspaces & database (hosts, services, loot, creds, notes) | ✅ |
| Reconnaissance modules (port/service/host discovery) | ✅ |
| Session / handler / job management | ✅ |
| Interfaces (console, RPC, db utility, resource scripts) | ✅ |
| Exploit / payload / post / encoder / nop / evasion bodies | ⚠️ |
| Denial-of-service | not planned |

## Getting started

Requires Go (see `go.mod` for the version).

```sh
# build (the output name must differ from the ./cmd directory)
go build -o msf-go ./cmd

# run the console
go run ./cmd

# full local quality gate (see Makefile)
make check        # gofmt check, vet, build, tests with -race
```

## Repository governance

This repository is tightly governed so that **no code is merged without review**:

- `main` is protected: no direct pushes, required pull requests, required approval from
  code owners (`CODEOWNERS`), required status checks, signed commits and linear history.
- CI runs build, `go vet`, `staticcheck`, race-enabled tests and formatting checks.
- Security CI runs `govulncheck`, secret scanning (gitleaks), dependency review and CodeQL.
- A commit-policy check enforces Conventional Commits and attribution rules.

To apply the branch protection and security settings after pushing to GitHub:

```sh
./scripts/setup-repo-security.sh
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution workflow.

## Security

Please report vulnerabilities privately — see [SECURITY.md](SECURITY.md). Do not open public
issues for security problems.

## License

No license is currently granted: all rights are reserved by the author. This is intentional
while the project is locked down. A formal license may be added later.

## Author

Guilherme Soares — guilhermeluzsoares@gmail.com
