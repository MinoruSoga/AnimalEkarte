# F8 G4 synthetic failure rehearsal

This runbook produces the failure-side evidence consumed by the `old_db` F8
contract. It uses a dedicated PostgreSQL volume and internal Docker network.
It cannot accept production CSVs, arbitrary SQL, a remote host, or a shared
AnimalEkarte database.

The producer rejects Docker environment overrides and requires the active
Docker context to resolve to a local Unix socket. It strips all `GIT_*`
variables before attestation, verifies the repository top-level, archives the
exact requested backend tree, and builds a no-cache runner image from that
immutable archive. The pinned Go/PostgreSQL base manifests, backend archive
digest/tree, local Docker daemon identity, database image, and runner image are
bound into the target database identity digest.

## Stop conditions

Do not run unless the target release checkout is committed and clean. The
runner rejects a dirty or mismatched HEAD, an existing Compose project, a DB
name outside `animalekarte_f8_g4_*`, clinic ordinal other than `1`, missing
seed bindings, non-localhost port publication, a non-internal network, or
resources without the exact disposable/run labels.

The fixed fixture copies one synthetic owner inside a serializable transaction,
then inserts one synthetic pet referencing a guaranteed-missing in-band owner.
Only PostgreSQL SQLSTATE `23503` followed by an explicit successful rollback,
identical zero-row counts across all 21 bands, and a passing seed/empty-band
preflight can produce evidence. The normal `csv-import apply` command has no
fault-injection option.

Each host phase and the database runner have finite time limits. A timeout
fails closed and emits no evidence; already-created disposable resources are
left in place for inspection and must be removed with the cleanup command.

## Configuration

Create an owner-only file such as
`sensitive-local/f8-g4-rehearsal.env` containing only:

```dotenv
DB_USER=<disposable-user>
DB_PASSWORD=<disposable-password>
DB_NAME=animalekarte_f8_g4_<run_slug>
```

Then set the non-secret bindings:

```bash
export F8_G4_COMPOSE_PROJECT=animalekarte-f8-g4-<run-slug>
export F8_G4_RUN_ID=<failure-run-id>
export F8_G4_TARGET_RELEASE_COMMIT="$(git rev-parse HEAD)"
export F8_G4_ENV_FILE="$PWD/sensitive-local/f8-g4-rehearsal.env"
export F8_G4_DB_PORT=15435
export F8_G4_CLINIC_CODE=<same-clinic-code-as-normal-rehearsal>
export F8_G4_CLINIC_ORDINAL=1
export TARGET_CLINIC_ID=<id>
export FALLBACK_ANIMAL_SPECIES_ID=<id>
export FALLBACK_EXAM_TYPE_ID=<id>
export TRIMMING_RESERVATION_TYPE_ID=<id>
export PAYMENT_METHOD_CASH_ID=<id>
export PAYMENT_METHOD_CREDIT_CARD_ID=<id>
```

Run the static and rendered configuration checks before execution:

```bash
make f8-g4-rehearsal-contract-test
make f8-g4-rehearsal-config-check
```

## Execute and transfer

```bash
make f8-g4-rehearsal-run
```

The producer leaves the isolated DB running for inspection and writes an
owner-only, no-clobber directory under
`sensitive-local/f8-g4-rehearsal/<run-id>/`. It contains the exact fixture,
runtime, apply, preflight, and failure evidence plus canonical before/after
count and transaction sidecars. Transfer regular files to the `old_db`
owner-only F8 input directory; never use symlinks.

The command prints the three sidecar SHA-256 values required by:

- `F8_FAILURE_TRANSACTION_EVIDENCE_EXPECTED_SHA256`
- `F8_FAILURE_BEFORE_COUNTS_EXPECTED_SHA256`
- `F8_FAILURE_AFTER_COUNTS_EXPECTED_SHA256`

Record those digests through the operator-controlled evidence channel. The
producer output alone does not complete F8, authorize production, or replace
the independent `old_db` evidence import.

## Cleanup

After inspection and evidence transfer:

```bash
make f8-g4-rehearsal-down
```

Cleanup re-attests the same local Docker daemon immediately before removal,
then verifies the exact Compose project, run ID, and disposable labels before
removing only that project and its dedicated volumes.
