# Security Scan Workflow

This repository uses `security-scan.yml` with AgentShield to run security checks.

- **Triggers**
  - `push` on `main`, `staging`
  - `pull_request` targeting `main`, `staging`, `production`
  - `workflow_dispatch` (manual run)

- **Failure behavior**
  - Fails CI only for:
    - PRs targeting `main`
  - Push events and non-`main` PR targets are reported but do not fail the pipeline.

- **Manual run**
  - `workflow_dispatch` input `min_severity` controls minimum severity to fail/report.
  - Options: `doc`, `low`, `medium`, `high`, `critical` (default: `medium`).
