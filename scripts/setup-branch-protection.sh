#!/bin/bash
# Branch Protection Rules Setup Script
# This script configures branch protection rules for the repository using GitHub CLI
#
# Prerequisites:
#   - GitHub CLI installed: https://cli.github.com/
#   - Authenticated: gh auth login
#
# Usage:
#   ./scripts/setup-branch-protection.sh [OWNER/REPO]
#
# Example:
#   ./scripts/setup-branch-protection.sh Laaaaksh/internal-transfers-service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Get repository from argument or detect from git remote
if [ -n "$1" ]; then
    REPO="$1"
else
    REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null || echo "")
    if [ -z "$REPO" ]; then
        echo -e "${RED}Error: Could not detect repository. Please provide OWNER/REPO as argument.${NC}"
        echo "Usage: $0 OWNER/REPO"
        exit 1
    fi
fi

echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  Branch Protection Rules Setup${NC}"
echo -e "${GREEN}  Repository: ${REPO}${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) is not installed.${NC}"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo -e "${RED}Error: Not authenticated with GitHub CLI.${NC}"
    echo "Run: gh auth login"
    exit 1
fi

echo -e "${YELLOW}Setting up branch protection for 'master' branch...${NC}"
echo ""

# Create branch protection rule for master branch using JSON input
gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "/repos/${REPO}/branches/master/protection" \
  --input - <<EOF
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["Lint", "Test", "Build"]
  },
  "enforce_admins": false,
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
  "required_conversation_resolution": true,
  "lock_branch": false,
  "allow_fork_syncing": false
}
EOF

echo ""
echo -e "${GREEN}============================================${NC}"
echo -e "${GREEN}  Branch Protection Rules Applied!${NC}"
echo -e "${GREEN}============================================${NC}"
echo ""
echo -e "${GREEN}Rules configured for 'master' branch:${NC}"
echo ""
echo "  ${YELLOW}Required Status Checks:${NC}"
echo "    - Lint (must pass)"
echo "    - Test (must pass)"
echo "    - Build (must pass)"
echo "    - Require branches to be up to date before merging"
echo ""
echo "  ${YELLOW}Pull Request Reviews:${NC}"
echo "    - Require 1 approving review"
echo "    - Dismiss stale reviews on new commits"
echo "    - Require review from Code Owners"
echo "    - Require approval of most recent push"
echo ""
echo "  ${YELLOW}Additional Protections:${NC}"
echo "    - Require linear history (no merge commits)"
echo "    - Require conversation resolution before merging"
echo "    - Block force pushes"
echo "    - Block branch deletion"
echo ""
echo -e "${GREEN}View settings: https://github.com/${REPO}/settings/branches${NC}"
echo ""
