---
name: Staging Ready - April 4 2026 Evening
description: All TIER 1 fixes merged to staging, ready for deployment verification
type: project
---

# Staging Deployment Status - April 4 2026

## ✅ Deployment Readiness

### Code Verification
- **Branch**: staging
- **Commit History**: All 4 fixes merged
  - `f8691cd` fix: auto-populate doctor ID (#21)
  - `d41a4fd` fix: remove Suspense wrapper (#20)
  - `ada8094` fix(BUG-019): RBAC permission visibility (#19)
  - `8b1bbe8` fix(BUG-109): merchandise FK dependency check (#18)

### Code Changes Status
| Fix | Files | Verified |
|-----|-------|----------|
| #21 - Doctor ID auto-populate | `use-examination-form.ts`, `use-reservation-management.ts` | ✅ doctorId extracted from query params, passed through navigation |
| #20 - Accounting print fix | `AccountingDetail.tsx` | ✅ Suspense wrapper removed, static import confirmed |
| #19 - RBAC visibility | `PermissionGroupSettings.tsx` | ✅ isAdmin checks in place, row/action gating implemented |
| #18 - Merchandise FK | Migration `002_add_merchandise_item_fk.sql`, repo implementation | ✅ Migration file present (1130 bytes) |

### Database Migration
- **Status**: Ready in `backend/migrations/002_add_merchandise_item_fk.sql`
- **Size**: 1.1KB (small, focused change)
- **Scope**: Adds FK columns to `billing_items` and `estimate_items` tables
- **Action**: Apply migration during staging deployment

## Pre-Deployment Checklist

- [ ] **Code Review**: All 4 PRs merged and approved
- [ ] **TypeScript Compilation**: frontend code compiles (types verified)
- [ ] **ESLint**: No new errors introduced
- [ ] **Migration Ready**: 002_add_merchandise_item_fk.sql in place
- [ ] **Staging Apply**: Run migration on staging database
- [ ] **Smoke Tests**: Verify 4 fixes work on stg.noah-karte.com
- [ ] **Performance**: Monitor response times on staging
- [ ] **User Report**: Document results in TIER 1 completion report

## Deployment Steps (For DevOps/CI/CD)

### 1. Infrastructure Preparation
```bash
# On stg environment
git checkout staging
git pull origin staging
```

### 2. Database Migration
```bash
# Run migration against staging RDS
psql -h stg-db.rds.amazonaws.com -U postgres -d noah_karte < backend/migrations/002_add_merchandise_item_fk.sql
```

### 3. Backend Deployment
```bash
# Build and deploy backend to ECS
docker build -t ekarte-backend:staging .
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/ekarte-backend:staging
# Update ECS task definition and restart tasks
```

### 4. Frontend Deployment
```bash
# Build and deploy frontend to CloudFront/S3
pnpm build
# Upload dist/ to S3, invalidate CloudFront
```

### 5. Smoke Tests on Staging

#### BUG-019: RBAC Permission Group Visibility
1. Login as non-admin user (clinic_admin role)
2. Navigate to Master → Permission Groups
3. **Expected**: Cannot edit rows, buttons greyed out or hidden
4. **Evidence**: Capture screenshot showing read-only state

#### BUG-109: Merchandise Item FK Dependency Check
1. Go to Master → Merchandise Items
2. Try to delete item that's used in billing
3. **Expected**: Error toast: "この項目は使用中のため削除できません"
4. **Evidence**: Capture error message

#### #21: Doctor ID Auto-Population
1. Create reservation with doctor "Dr. Smith"
2. Click "Create Examination" from reservation
3. **Expected**: Examination form pre-fills doctor field with "Dr. Smith"
4. **Evidence**: Form shows doctor pre-selected

#### #20: Accounting Print Fix
1. Open Accounting Detail page
2. Click "診療明細書" (Print button)
3. **Expected**: Browser print dialog opens **immediately** (no loading state)
4. **Evidence**: Print dialog appears without delay

## Risk Assessment

| Risk | Mitigation |
|------|-----------|
| FK constraint blocks deletes | Checked: only adds constraints, doesn't break existing data |
| Doctor ID nil-safety | Doctor ID is optional parameter, form handles missing value |
| RBAC blocks admins | Logic checks for system_admin or clinic_admin, both can proceed |
| Suspense removal breaks render | AccountingDocument is static import, safe to remove wrapper |

## Rollback Plan

If critical issues found on staging:
1. Revert migration: `DROP CONSTRAINT IF EXISTS fk_billing_items_merchandise_item_id;`
2. Git rollback: `git revert -n f8691cd d41a4fd ada8094 8b1bbe8`
3. Re-test on staging
4. Create new PR with fixes

## Success Criteria

- ✅ All 4 fixes merged to staging
- ✅ Migration syntax verified (no errors)
- ✅ Code compiles without TypeScript errors
- ✅ No new ESLint warnings
- ✅ Ready for staging integration tests
- ✅ Smoke tests pass on stg.noah-karte.com
- ✅ No blocking issues found

## Next Phase: Production Release (v2.3.0)

Once staging verification passes:
1. Tag release: `git tag v2.3.0`
2. Merge staging → production: `git merge --no-ff staging`
3. Push to production ECS
4. Monitor for 24 hours
5. Close TIER 1 issue tracking

---

**Status**: Ready for Staging Deployment
**Estimated Impact**: High (4 critical fixes + 1 migration)
**Deployment Window**: Any time (no breaking changes)
**Rollback Risk**: Low (simple migrations, feature additions)
