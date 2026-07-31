# TASK-022 Identity Link Closeout — 2026-07-31

## Summary

Fixed weak authorization fallback in `CreatePetGroup`: actors missing the parent owner-group anchor clinic (or any active owner-member clinic) can no longer create pet groups based on partial clinic visibility. Mutations require the full parent-owner clinic set plus all pet clinics. Zero writes on reject. Phase 2 (auto-link, merge, extra surfaces, new DDL, record moves) **not started**.

## Code change

- **Package**: `backend/internal/identitylink/`
- **Bug**: `CreatePetGroup` allowed mutation when actor lacked `ownerGroup.CreatedClinicID` if any owner member clinic was in actor scope (`anyOwnerMemberInActorScope`). Also did not require non-anchor owner-member clinics when anchor was present.
- **Fix**: Unconditional `assertActorCoversOwnerGroupClinics(actor, ownerGroup, ownerMembers)` — requires anchor + every active owner member clinic. Removed `anyOwnerMemberInActorScope`.
- **Other mutation paths scanned**:
  - `AddPetMembers` / `UnlinkPetMember`: use `assertCanManagePetGroup` (pet group anchor + owner-group anchor + all pet member clinics). No any-member fallback. Left as-is to preserve deterministic lock order (pet group lock before pets; locking parent owner group here would invert CreatePetGroup order).
  - `CreateOwnerGroup` / `AddOwnerMembers` / `UnlinkOwnerMember`: already use `assertCanManageOwnerGroup` / ref scope checks.
  - Read paths: continue clinic-filtered list; view vs edit separation unchanged.

## Tests

### RED (before fix)

```
docker compose -f .../AnimalEkarte/docker-compose.yml run --rm -T --no-deps \
  -v .../AnimalEkarte-TASK-022/backend:/app -w /app --entrypoint go \
  backend test -p 1 ./internal/identitylink -count=1 \
  -run 'TestCreatePetGroup_RejectsMissingParentOwner|TestCreatePetGroup_AllowsWhenActorCoversAllParentOwner'
```

Result:

- `TestCreatePetGroup_RejectsMissingParentOwnerAnchorClinic_NoPartialWrite` — FAIL (got nil error; weak fallback allowed write)
- `TestCreatePetGroup_RejectsMissingParentOwnerMemberClinic_NoPartialWrite` — FAIL (got nil error)
- Happy path would still pass under old code when full clinic set present

### GREEN (after fix)

Same filter: **ok**. Full scoped gate:

```
go test -p 1 ./internal/identitylink ./internal/apicontract -count=1
```

Result: **ok** both packages (identitylink ~0.95s, apicontract ~0.26s). Existing mixed/hidden/audit atomicity tests remain green.

### New regression tests

| Test | Intent |
|:--|:--|
| `TestCreatePetGroup_RejectsMissingParentOwnerAnchorClinic_NoPartialWrite` | Missing anchor, has member clinics → Forbidden; zero CreatePetGroup/CreatePetMembers/audit |
| `TestCreatePetGroup_RejectsMissingParentOwnerMemberClinic_NoPartialWrite` | Has anchor, missing member clinic → Forbidden; zero writes |
| `TestCreatePetGroup_AllowsWhenActorCoversAllParentOwnerAndPetClinics` | Full clinic set → success + audit |

## DOCKER-MOUNT-PROOF

Worktree backend bind-mounted over compose service `/app` so tests see this worktree, not shared main.

| File | Host SHA-256 | Container SHA-256 |
|:--|:--|:--|
| `internal/identitylink/service.go` | `ccb2e79df7b020ebea034cc4132a269be3baf0e0aaea4e1d790bd932455354d4` | `ccb2e79df7b020ebea034cc4132a269be3baf0e0aaea4e1d790bd932455354d4` |
| `internal/identitylink/service_test.go` | `06af2f1b36a7162dc5bdcc4dcfe344e58ba2f83fb7ee9c0e186a66fa9f2d703d` | `06af2f1b36a7162dc5bdcc4dcfe344e58ba2f83fb7ee9c0e186a66fa9f2d703d` |

Match confirmed before trusting GREEN results. (Container uses `sha256sum`; host uses `shasum -a 256`.)

## Docs

- Updated `docs/spec/screens/40-identity-links.md` — §2.5 parent-owner full clinic set; AC-3b.
- New `docs/ops/testing/scenarios/S13-identity-links-manual-correction.md` — 2-clinic link→history→unlink→relink; human sign-off slots PENDING.
- Docs symbol drift: baseline failures only (table count 115 vs 116 class); **no new FAIL lines** introduced by this task (see session drift diff).

## Phase 2

**Not started.** No auto-link, merge, candidate UI, extra surfaces, new DDL, or record moves.

## PENDING (human only)

| Item | Owner field | Status |
|:--|:--|:--|
| S13 manual correction run (2-clinic link→history→unlink→relink) | S13 実施者 / 承認者 | PENDING |
| Named signer on S13 sign-off table | 承認者（記名） | PENDING |
| RLS runtime proof under DB role (if required for release) | Ops / DBA | PENDING — intent proven at source/test (`assertActorCoversOwnerGroupClinics` + regression tests); runtime role exercise not automated here |

## Out of scope / non-goals

- No seeds, no migrate apply, no codegen, no push.
- No PHI or credentials in this report.
- Claim branch `claim/TASK-022` not deleted (coordinator-owned).
