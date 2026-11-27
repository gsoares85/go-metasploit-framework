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

echo ">> Requiring signed commits on 'main'..."
# Signature enforcement has its own endpoint; it is not a field of the protection payload.
gh api -X POST "repos/$REPO/branches/main/protection/required_signatures" >/dev/null
echo "   signed commits required"

echo ">> Enabling secret scanning and push protection..."
# gh api -F sends flat fields; the nested object is sent via --input JSON.
gh api -X PATCH "repos/$REPO" --input - >/dev/null <<'JSON' || \
  echo "   (secret scanning may require GitHub Advanced Security on private repos)"
{
  "security_and_analysis": {
    "secret_scanning": { "status": "enabled" },
    "secret_scanning_push_protection": { "status": "enabled" }
  }
}
JSON

echo ">> Enabling private vulnerability reporting..."
gh api -X PUT "repos/$REPO/private-vulnerability-reporting" >/dev/null
echo "   private vulnerability reporting enabled"

gh api -X PUT "repos/$REPO/vulnerability-alerts" >/dev/null
echo "   vulnerability alerts enabled"
gh api -X PUT "repos/$REPO/automated-security-fixes" >/dev/null
echo "   automated security fixes enabled"

echo ">> Done. 'main' now requires reviewed PRs, code-owner approval, signed commits,"
echo "   linear history, passing checks, and blocks direct/force pushes."
echo ">> Note: required status check names must match the CI job names after the first run."
