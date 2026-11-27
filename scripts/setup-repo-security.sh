#!/usr/bin/env bash
#
# Lock down the repository so that no code can be added without a reviewed PR.
# Run once, after the repository exists on GitHub and `gh` is authenticated.
#
#   Requirements: GitHub CLI (`gh auth login`) with admin rights on the repo.
#   Usage:        ./scripts/setup-repo-security.sh [owner/repo]
#
set -euo pipefail

REPO="${1:-$(gh repo view --json nameWithOwner -q .nameWithOwner)}"
echo ">> Target repository: $REPO"

echo ">> Applying branch protection to 'main'..."
gh api -X PUT "repos/$REPO/branches/main/protection" \
  --input - <<'JSON'
{
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "Build, vet, test",
      "Static analysis (staticcheck)",
      "govulncheck",
      "Secret scan (gitleaks)",
      "CodeQL",
      "Conventional Commits + attribution policy"
    ]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": true,
    "required_approving_review_count": 1,
    "require_last_push_approval": true
  },
  "restrictions": null,
  "required_linear_history": true,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": true
}
JSON

# Optional hardening. A failure is reported as a warning but does not abort the
# script, so one unavailable feature does not block the others. (set -e still
# guards the mandatory branch-protection call above.)
enable() { # <description> <gh api args...>
  local desc="$1"; shift
  if gh api "$@" >/dev/null; then
    echo "   ${desc}: enabled"
  else
    echo "   WARNING: could not enable ${desc} (may need a different plan or admin rights)" >&2
  fi
}

echo ">> Requiring signed commits on 'main'..."
# Signature enforcement has its own endpoint; it is not a field of the protection payload.
enable "signed commits on main" -X POST "repos/$REPO/branches/main/protection/required_signatures"

echo ">> Enabling secret scanning and push protection..."
# gh api -F sends flat fields; the nested object is sent via --input JSON.
enable "secret scanning + push protection" -X PATCH "repos/$REPO" --input - <<'JSON'
{
  "security_and_analysis": {
    "secret_scanning": { "status": "enabled" },
    "secret_scanning_push_protection": { "status": "enabled" }
  }
}
JSON

echo ">> Enabling Dependabot alerts and private vulnerability reporting..."
enable "private vulnerability reporting" -X PUT "repos/$REPO/private-vulnerability-reporting"
enable "Dependabot vulnerability alerts" -X PUT "repos/$REPO/vulnerability-alerts"
enable "automated security fixes" -X PUT "repos/$REPO/automated-security-fixes"

echo ">> Branch protection applied: 'main' requires reviewed PRs, code-owner approval,"
echo "   linear history, passing checks, and blocks direct/force pushes."
echo ">> Review any WARNING lines above and verify optional features in repository settings."
echo ">> Note: required status check names must match the CI job names after the first run."
