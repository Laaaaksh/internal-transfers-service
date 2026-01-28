# Branch Protection Rules

This document describes the branch protection rules configured for this repository.

## Overview

Branch protection rules ensure code quality and prevent accidental changes to critical branches. The `master` branch is protected with the following rules.

## Protected Branches

### `master` Branch

The main branch is protected with comprehensive rules to ensure stability.

#### Required Status Checks

All pull requests must pass these CI checks before merging:

| Check | Description |
|-------|-------------|
| **Lint** | Code must pass golangci-lint with no errors |
| **Test** | All tests must pass with race detector enabled |
| **Build** | Binary must compile successfully |

Additionally:
- **Require branches to be up to date**: PRs must be rebased on the latest master before merging

#### Pull Request Reviews

| Rule | Setting |
|------|---------|
| Required approving reviews | 1 |
| Dismiss stale reviews | Yes |
| Require Code Owner review | Yes |
| Require approval of most recent push | Yes |

**What this means:**
- Every PR needs at least 1 approval
- If you push new commits after approval, you need re-approval
- Code Owners (defined in `.github/CODEOWNERS`) must review changes in their areas

#### Branch Restrictions

| Rule | Setting |
|------|---------|
| Require linear history | Yes |
| Allow force pushes | No |
| Allow deletions | No |
| Require conversation resolution | Yes |

**What this means:**
- No merge commits - use rebase or squash merging
- Cannot force push to master (even admins)
- Cannot delete the master branch
- All PR comments must be resolved before merging

## Code Owners

The `.github/CODEOWNERS` file defines who must review changes to specific parts of the codebase:

```
# Default owners for everything
* @Laaaaksh

# Core application code
/cmd/ @Laaaaksh
/internal/ @Laaaaksh
/pkg/ @Laaaaksh

# Database migrations - require extra review
/internal/database/migrations/ @Laaaaksh

# CI/CD and deployment
/.github/ @Laaaaksh
/deployment/ @Laaaaksh
```

## Setting Up Branch Protection

### Using the Setup Script

Run the provided script to configure branch protection:

```bash
# Make the script executable
chmod +x scripts/setup-branch-protection.sh

# Run the script (auto-detects repository)
./scripts/setup-branch-protection.sh

# Or specify repository explicitly
./scripts/setup-branch-protection.sh Laaaaksh/internal-transfers-service
```

### Prerequisites

1. **GitHub CLI**: Install from https://cli.github.com/
2. **Authentication**: Run `gh auth login` and authenticate
3. **Admin Access**: You need admin permissions on the repository

### Manual Setup (GitHub UI)

1. Go to **Settings** → **Branches**
2. Click **Add branch protection rule**
3. Enter `master` as the branch name pattern
4. Configure the following:

   **Protect matching branches:**
   - [x] Require a pull request before merging
     - [x] Require approvals: 1
     - [x] Dismiss stale pull request approvals when new commits are pushed
     - [x] Require review from Code Owners
     - [x] Require approval of the most recent reviewable push
   - [x] Require status checks to pass before merging
     - [x] Require branches to be up to date before merging
     - Search and add: `Lint`, `Test`, `Build`
   - [x] Require conversation resolution before merging
   - [x] Require linear history
   - [x] Do not allow bypassing the above settings

   **Rules applied to everyone including administrators:**
   - [ ] Allow force pushes (keep unchecked)
   - [ ] Allow deletions (keep unchecked)

5. Click **Create** or **Save changes**

## Workflow for Contributors

### Creating a Pull Request

1. Create a feature branch from `master`:
   ```bash
   git checkout master
   git pull origin master
   git checkout -b feature/my-feature
   ```

2. Make your changes and commit:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

3. Push and create PR:
   ```bash
   git push -u origin feature/my-feature
   gh pr create --title "feat: add new feature" --body "Description of changes"
   ```

4. Wait for CI checks to pass and request review

5. Address review feedback (new commits will require re-approval)

6. Once approved and all checks pass, merge using **Squash and merge** or **Rebase and merge**

### Keeping Your Branch Up to Date

If master has new commits:

```bash
git fetch origin
git rebase origin/master
git push --force-with-lease
```

Note: Force pushing to your feature branch is allowed; force pushing to `master` is blocked.

## Troubleshooting

### "Required status check is expected"

The CI pipeline hasn't run yet. Push a commit or re-run the workflow.

### "Review required from Code Owners"

A Code Owner hasn't approved yet. Check `.github/CODEOWNERS` to see who needs to review.

### "Merge blocked: linear history required"

Your branch has merge commits. Rebase instead:

```bash
git fetch origin
git rebase origin/master
git push --force-with-lease
```

### "Conversations must be resolved"

There are unresolved review comments. Resolve them in the GitHub UI before merging.

## Related Documentation

- [Development Guide](development.md)
- [Getting Started](getting-started.md)
- [CI/CD Workflow](.github/workflows/ci.yml)
