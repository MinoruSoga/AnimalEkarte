# LINE R-06 / R-07 — delivery-monitor nav honesty + tags RBAC align

**Date:** 2026-07-31  
**Unit:** TODO-MD-OPEN-REMAINING-ORCH-WAVE-20260731-V2  
**Agent:** W-LINE-FIX  
**Branch:** `agent/w-line-fix`  
**Claim:** `claim/LINE-R-FIX` (user-owned release)

## Product philosophy gate

This change is **nav honesty / RBAC completeness**, not a new product domain:

- R-06: The delivery-monitor screen already exists (route + page + RBAC guard). The incompleteness was that the sidebar never exposed it.
- R-07: Tags route already requires `ResourceLstepAnalytics`. The sidebar still advertised `ResourceHospitalSettings`, so menu visibility disagreed with the route guard.

No new screens, no new permissions, no webhook/credential policy work (R-01 / R-05 out of scope).

## Changes

### R-06 — delivery-monitor nav honesty

1. `frontend/src/config/paths.ts`  
   - Added `paths.lstep.deliveryMonitor`  
   - `path` / `getHref` → `/lstep/delivery-monitor`

2. `frontend/src/components/shared/Layout/sidebar-menu.tsx`  
   - Under **Lステップ連携** subItems, added **配信監視**  
   - path: `paths.lstep.deliveryMonitor`  
   - resource: `ResourceLstepAnalytics` (mirrors analytics / screen34 route guard)

### R-07 — tags RBAC align

1. Same `sidebar-menu.tsx`  
   - **タグ管理** subItem resource: `ResourceHospitalSettings` → `ResourceLstepAnalytics`  
   - Aligns with `settings-routes.tsx` RequirePermission and `settings-routes.lstep-tags.test.tsx`

## Tests

New: `frontend/src/components/shared/Layout/sidebar-menu.lstep-nav.test.tsx`

Asserts:

- `paths.lstep.deliveryMonitor` exposes `/lstep/delivery-monitor`
- 配信監視 subItem present with `ResourceLstepAnalytics`
- タグ管理 resource === `ResourceLstepAnalytics`
- 分析レポート resource === `ResourceLstepAnalytics` (mirror sanity)

Related existing route guards (not modified):

- `src/app/routes/settings-routes.lstep-tags.test.tsx`
- `src/app/routes/operations-routes.lstep-delivery-monitor.test.tsx`

## Verification

### rg evidence (worktree)

```text
frontend/src/config/paths.ts
  deliveryMonitor: {
    path: "/lstep/delivery-monitor",
    getHref: () => "/lstep/delivery-monitor",
  },

frontend/src/components/shared/Layout/sidebar-menu.tsx
  タグ管理 … resource: ResourceLstepAnalytics
  配信監視 … path: paths.lstep.deliveryMonitor … resource: ResourceLstepAnalytics
```

### Scoped Docker test command

```bash
# Prefer from worktree (compose mounts ./frontend relative to CWD):
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-w-line
docker compose exec frontend pnpm test:run -- \
  src/components/shared/Layout/sidebar-menu.lstep-nav.test.tsx \
  src/app/routes/settings-routes.lstep-tags.test.tsx \
  src/app/routes/operations-routes.lstep-delivery-monitor.test.tsx
```

**Agent session note:** This implementer session had no shell tool; Docker/git were not executed here. Code + unit test are in place for land. If the active compose project is the main tree (`AnimalEkarte`) and not this worktree, run the above after landing (main mounts `./frontend` of the main repo only).

### Static evidence (read/rg equivalent)

| Check | Result |
|-------|--------|
| `paths.lstep.deliveryMonitor` | `/lstep/delivery-monitor` present |
| sidebar 配信監視 | path + `ResourceLstepAnalytics` |
| sidebar タグ管理 | `ResourceLstepAnalytics` (was HospitalSettings) |
| route guards (existing) | tags + delivery-monitor already require LstepAnalytics |

## Allowlist touched (≤12)

1. `frontend/src/config/paths.ts`
2. `frontend/src/components/shared/Layout/sidebar-menu.tsx`
3. `frontend/src/components/shared/Layout/sidebar-menu.lstep-nav.test.tsx` (new)
4. `reports/2026-07-31-line-r06-r07-nav-honesty.md`

## Out of scope (explicit non-goals)

- R-01 webhook policy
- R-05 credential source of truth
- Backend / todo.md / non-allowlisted paths
