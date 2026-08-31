# Production Cloudflare setup contract

> Human-operated checklist for the planned Cloudflare Workers + Containers / PlanetScale production environment. **Do not execute from this document until the currently approved go-live date and Linear delivery item are confirmed.** No external runtime/account/billing state was verified during this update.

Checked-in `backend/wrangler.production.jsonc` and `infra/cloudflare/production/` are drafts. Their presence does not prove that GitHub Environment protection, billing, PlanetScale, Cloudflare, DNS, R2, or Vercel production configuration exists.

## 1. Fail-closed preconditions

A human owner must record dated evidence for all items before deployment:

- target account/zone/project and billing are approved;
- production database, role, backup/restore owner, R2, DNS, and certificate are verified;
- a GitHub Environment with required reviewers exists and its exact case-sensitive name matches the workflow;
- production secrets are environment-scoped and staging values are not reused unless explicitly approved;
- frontend production settings are verified. Checked-in `frontend/vercel.json` and `.env.production` contain STG/AWS-era assumptions, so “Vercel auto-deployed” is not proof that `VITE_API_URL` targets production;
- workflow/config tests and `actionlint` are green on the reviewed change;
- current Linear delivery state and go-live date are confirmed externally.

Stop if any item is unknown. Do not infer runtime state from this draft.

## 2. Required production secret **names**

The inventory source is `backend/wrangler.production.jsonc` `secrets.required`. Never put values in docs, shell history, logs, Terraform state, or tickets.

| Category | Required names |
|---|---|
| Database | `DB_HOST`, `DB_USER`, `DB_PASSWORD` |
| R2/S3 binding | `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` |
| Migration | `MIGRATE_RUN_SECRET` |
| Scheduler | `SCHEDULER_OPS_SECRET`, `SCHEDULER_INTERNAL_TOKEN`, `SCHEDULER_ACCESS_TEAM_DOMAIN`, `SCHEDULER_ACCESS_AUDIENCE`, `SCHEDULER_ALERT_ALLOWED_HOST`, `SCHEDULER_ALERT_WEBHOOK_URL`, `SCHEDULER_ALERT_WEBHOOK_SECRET` |
| Application | `JWT_SECRET`, `INTEGRATION_ENCRYPTION_KEY` |
| SMTP | `SMTP_HOST`, `SMTP_USER`, `SMTP_PASS` |

Secret names must be checked against the config again at execution time. For each name, use the protected operator channel from `backend/`:

```bash
npx wrangler secret put <NAME> -c wrangler.production.jsonc
```

Do not use `pnpm exec wrangler`; the effective working directory previously caused a wrong-Worker deployment. For scheduler rollout, create and validate `SCHEDULER_INTERNAL_TOKEN` and its access/alert dependencies **before** the first deploy. Deploy last. Missing scheduler secrets must fail closed.

## 3. R2 credentials

The old generic-token API and “SHA-256 of `result.value`” recipe is removed. Before credential creation, verify the current official Cloudflare R2 S3 credential documentation and use its supported UI/API flow. Confirm the repository is allowed to expose any account identifier; otherwise use `<CLOUDFLARE_ACCOUNT_ID>` from the protected operator store. Do not run copied `curl` commands blindly.

## 4. Seed contract

At HEAD `70dc7405`, only `backend/migrations/seeds/002_master` exists. `seedbundle.BundleOrderForEnv` returns master-only for every environment. `003_demo`, `004_staging`, “seed +3”, and environment-specific demo cleanup are historical behavior and must not be used.

`APP_ENV` is not passed through current Wrangler/Container source and is not a production prerequisite. The master-only behavior does not depend on it. Production UAT/synthetic users require a separate approved import, owner, expiry, and cleanup path.

## 5. IaC and resource verification

Terraform plan must be reviewed by a human before apply. It currently covers only the checked-in resources; Hyperdrive/notification tombstones are not active resources. Apply/destroy, database creation, credential changes, shared environment writes, billing, and DNS changes are human-only external operations.

## 6. Deployment workflow acceptance criteria

Do not copy an embedded patch from docs. Modify and review the workflow against current HEAD. Production activation is acceptable only when:

1. production trigger is enabled **after** Environment protection exists;
2. production job binds the exact protected GitHub Environment;
3. staging uses `npx wrangler deploy`;
4. production uses `npx wrangler deploy -c wrangler.production.jsonc`;
5. environment-specific Worker URL and secret scope are selected explicitly;
6. workflow tests and `actionlint` pass;
7. execution order is **deploy → migrate → `/health` → optional smoke**.

At HEAD `70dc7405`, `backend-deploy.yml` lacks the production trigger/Environment gate. Therefore a production workflow invocation is **not runnable** until the implementation is merged and verified.

## 7. Go-live verification

- verify workflow run `headSha`, environment approval, Worker name, route, and config file;
- verify `/health`, then DB-backed behavior separately because `/health` does not query the DB;
- verify frontend production API target from deployed output, not from branch name;
- use only explicitly provisioned synthetic credentials with lifecycle ownership for CRUD smoke;
- record non-secret evidence and rollback owner.

All checkboxes are requirements, not claims of current external state.
