# Security Policy

## Purpose and intended use

`go-metasploit-framework` is an **educational re-implementation** of the architecture of a
security testing framework. It is intended for **authorized** use only: your own lab
environments, penetration testing under contract, CTF competitions, and defensive research.

Offensive-capability modules (exploit/payload/post/encoder/nop/evasion) are provided as
**interfaces and scaffolding only**. Do not use this project against systems you do not own
or do not have explicit, written authorization to test. Misuse may be illegal.

## Supported versions

The project is pre-1.0 and under active development. Only the `main` branch is supported.

## Reporting a vulnerability

Please report security issues **privately**. Do not open a public issue for vulnerabilities.

- Preferred: open a **GitHub Security Advisory** (repository → Security → Advisories → Report a vulnerability).
- Provide: affected version/commit, reproduction steps, impact, and any suggested remediation.

You will receive an acknowledgement as soon as possible. Please allow reasonable time for a
fix before any public disclosure.

## Scope

In scope: the framework code in this repository (core, module engine, database/workspaces,
sessions/handlers management, recon modules, interfaces).

Out of scope: use of the tool against third parties, and requests to add weaponized
offensive logic — these are intentionally left as scaffolding.

## Hardening in this repository

- Protected `main` branch: no direct pushes, required reviews from code owners, required status checks.
- Secret scanning and push protection, dependency review, and CodeQL on pull requests.
- Secret detection (gitleaks) and `govulncheck` in CI.
- Signed commits and linear history required.
