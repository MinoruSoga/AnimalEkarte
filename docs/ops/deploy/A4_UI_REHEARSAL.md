# A4 UI rehearsal — isolated local stack

This runbook creates the disposable AnimalEkarte environment used by the
`old_db` A4 screen rehearsal. It never uses the normal `animalekarte` Compose
project or its database.

## Stop conditions

Before starting, record the named execution owner in the private work log.
Stop without creating evidence if the formal 21-table bundle, complete KNJO
recovery, target release commit, empty target band, or backup/restore preflight
is missing. Never copy credentials, row values, identifiers, or screenshots
into Git, terminal transcripts, chat, or the aggregate evidence JSON.

The target AnimalEkarte checkout must be committed, clean, and remain
quiescent until evidence import completes. Use a unique project matching
`animalekarte-a4-*`; never reuse an existing project or volume.

## Configuration

Create an owner-only minimal env file under ignored `sensitive-local/`.
Allowed keys are `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`,
`JWT_SECRET`, `APP_ENV`, `APP_PORT`, `TRUSTED_PROXY_CIDR`, and
`STORAGE_TYPE`; unrelated cloud, LINE, and production credentials are rejected.
Choose host ports that do not overlap the normal stack.

```bash
install -d -m 700 sensitive-local
# Create sensitive-local/a4-rehearsal.env with only the allowed local values,
# then chmod 600 it.
export A4_COMPOSE_PROJECT=animalekarte-a4-<clinic-run-slug>
export A4_RUN_ID=<migration-run-id>
export A4_TARGET_RELEASE_COMMIT="$(git rev-parse HEAD)"
export A4_ENV_FILE="$PWD/sensitive-local/a4-rehearsal.env"
export A4_DB_PORT=15434
export A4_BACKEND_PORT=18080
export A4_FRONTEND_PORT=13003

make a4-rehearsal-contract-test
make a4-rehearsal-config-check
```

The rendered configuration must pass before any container starts. It enforces
one localhost-only port per service, one project-specific internal network,
one project-specific DB volume, disposable/run labels, and OCI revision labels.

## Operator sequence

The following commands change only the named disposable project. Execute them
manually after the stop conditions are satisfied.

```bash
make a4-rehearsal-up
make a4-rehearsal-ps
```

Run the explicit `a4-csv-import-preflight`, `a4-csv-import`, and
`a4-csv-import-verify` targets with the formal bundle variables documented in
[CLINIC_CSV_IMPORT.md](CLINIC_CSV_IMPORT.md). The ordinary `csv-import-*`
targets belong to the normal development stack and are forbidden for A4.

After apply and verify, attest the two operator-owned preflights and generate
the runtime report. `A4_APPLY_REPORT` must point to an owner-only regular file
under `sensitive-local/`.

```bash
export A4_APPLY_REPORT=sensitive-local/csv-import-reports/<apply-report>.json
export A4_EMPTY_BAND_PREFLIGHT=PASS
export A4_BACKUP_RESTORE_PREFLIGHT=PASS
make a4-rehearsal-runtime-report
```

The report is written with mode `0600`, atomically and without overwrite, to
`sensitive-local/a4-rehearsal-reports/`. It contains aggregate runtime identity
only. The `old_db` evidence importer independently re-inspects the live
containers, images, network, volume, target DB identity, and clean Git HEAD,
then runs canonical `csv-import-verify` again.

Copy the report as a regular owner-only file into `old_db/sensitive-local/`;
symlinks are forbidden. Compare both SHA-256 values before setting
`UI_REHEARSAL_RUNTIME_REPORT`.

```bash
install -m 600 \
  sensitive-local/a4-rehearsal-reports/<runtime-report>.json \
  ../old_db/sensitive-local/ui-rehearsal/runtime-report.json
shasum -a 256 \
  sensitive-local/a4-rehearsal-reports/<runtime-report>.json \
  ../old_db/sensitive-local/ui-rehearsal/runtime-report.json
```

Perform the seven UI journeys from the `old_db` A4 contract only inside this
environment. Any failed health check, contract check, import verification, or
mapping check is a stop condition; do not emit PASS evidence.

## Explicit cleanup

Cleanup is destructive only to the exact validated A4 project and its volumes.
Run it after evidence import or after an aborted rehearsal:

```bash
make a4-rehearsal-down
```

Do not run generic `make down`, `make reset`, or manual volume deletion as part
of A4. Confirm the normal development stack container IDs and health are
unchanged before and after the rehearsal.
