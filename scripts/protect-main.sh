#!/usr/bin/env bash

set -euo pipefail

OWNER="${1:-carlisle0615}"
REPO="${2:-OpenAgentCleaner}"
BRANCH="${3:-main}"

gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  "repos/${OWNER}/${REPO}/branches/${BRANCH}/protection" \
  --input - <<JSON
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["build-and-test"]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": null,
  "restrictions": null,
  "required_linear_history": false,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": true,
  "lock_branch": false,
  "allow_fork_syncing": true
}
JSON

echo "Protected ${OWNER}/${REPO}:${BRANCH} with required check build-and-test"
