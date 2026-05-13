# Git Convention

This document outlines the standard Git practices and workflow for our development team.

# 1. Branching Model

We typically use three main types of branches:

main / master: No direct commits allowed.

dev: Integration branch where all features are merged for testing.

member_name/feature/feature_name: Feature branches created for individual tasks or bug fixes.

# 2. Workflow

## Step 1: Pull the Latest Code

Always start by ensuring your local dev branch is up to date with the remote repository.
```
    git checkout dev
    git pull origin dev
```

## Step 2: Create a New Branch

Create a descriptive feature branch from the dev branch.

### Example: git checkout -b trkien/feature/login-api
git checkout -b member_name/feature/feature-name
- feat: New features or significant additions.
- fix: Bug fixes.
- refactor: Code changes that neither fix a bug nor add a feature (e.g., cleaning up code).

## Step 3: Develop and Commit

Keep your commits clear and concise. We follow standard prefix conventions.

```
    git add .
    git commit -m "feat: implement login API"
```


## Common Commit Prefixes:
- feat: New features or significant additions.
- fix: Bug fixes.
- refactor: Code changes that neither fix a bug nor add a feature (e.g., cleaning up code).
- config: Changes to configuration files, CI/CD pipelines, or tools.
- docs: Documentation changes only (e.g., updating README).

## Step 4: Push to Remote
- Upload your local branch to the remote server.
- git push origin member_name/feature/feature-name


## Step 5: Create a Pull Request (PR)
- Initiate a PR from feature/* → dev.
- Every PR requires a review from at least 1–2 team members.

## Step 6: Merge
- To Dev: Once approved and tests pass, merge the feature branch into dev.
- To Main: When the dev branch is stable and ready for release, it is merged into the main branch.