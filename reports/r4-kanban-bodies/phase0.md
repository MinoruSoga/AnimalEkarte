# Phase0 r4 (2026-08-13)

Repo: /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
main tip at classify: 5cf775706
Runtime: frontend/backend Up ~35h, bind-mount main tree.

## Classification

| BUG | Class | Evidence |
|-----|-------|----------|
| 033 | STILL_OPEN | Unit lock helpers PASS on main; UAT 2026-08-13 still editable completed exam 1014563 (results/delete/add). Investigate seal (revision_version null-only) + FE wiring; not pure ENV_STALE after bind-mount recreate history. |
| 035 | STILL_OPEN | Unit/permissions tests PASS; UAT finalized MR 1425559 still editable (chief/treatment disabled=false, save remains) despite fieldset lock. Find residual unlock path (portals/custom controls/status map). |
| 036 | STILL_OPEN | ShiftFormDialog allows empty start/end; DB row id=2460 NULL times. |
| 037 | STILL_OPEN | Hospitalization create without cage; DB id=8 cage_id=NULL. |
| 038 | STILL_OPEN | /settings/clinic shows 0 while DB clinics=4; check scope=all + hospital-settings.can_view + error swallow. |
| 027 | SPEC | no code |
| S04 dates disabled | DEFER/env | not code sprint |

User authorized: after verify PASS, serial merge + normal push origin main. No force-push. No auto migrate. No Linear Done.
