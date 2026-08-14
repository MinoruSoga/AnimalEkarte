# staging ← main preflight（read-only · agent 2026-08-14）

| 項目 | 結果 |
|------|------|
| remote fetch | origin/main · origin/staging 取得済 |
| tip main | `1386e1db0` |
| tip staging | `e6faaa01d` |
| `staging...main` left-right | **4 / 1346**（staging-only 4 · main-ahead 1346） |
| merge 方式（方針） | merge-commit PR のみ · squash/直接 merge/STG reset **禁止** |
| agent 実施範囲 | **read-only まで** · PR 作成・merge・migrate **未実施** |

## staging-only 4 commits
```
e6faaa01d Merge pull request #267 from MinoruSoga/main
b8be3ad69 Merge pull request #265 from MinoruSoga/main
6c81594fb fix(smoke): CRUDスモークをデータセット非依存にする（フルデモ対応） (#264)
b8318a506 fix(ci): CFデプロイのCWDバグ修正 — 実Worker(animalekarte-stg-api)へ正しく配布する (#263)
```

## main tip（staging に未取り込み側の先頭）
```
1386e1db0 docs: drain bug.md — remove FIXED bugs; open code bugs none
2f76cf8b3 docs: close bug.md after r4 land (033/035/036/037/038 FIXED)
48a17ae68 Merge branch 'fix/bug-038-clinic-master-list' into main (bug.md r4 land)
97ea499d8 Merge branch 'fix/bug-037-hosp-required-cage' into main (bug.md r4 land)
5b22180d5 Merge branch 'fix/bug-036-shift-required-times' into main (bug.md r4 land)
0cc6eb739 Merge branch 'fix/bug-035-mr-finalized-lock' into main (bug.md r4 land)
ca9e706ce Merge branch 'fix/bug-033-exam-completed-lock' into main (bug.md r4 land)
56538fbf0 fix(BUG-037): require cage on hospitalization create/update
```

## migration / seeds diff（stat 要約）
```
 .../seeds/003_demo/trimming_course_types.csv       |   12 +-
 .../migrations/seeds/003_demo/trimming_courses.csv |   23 +-
 .../migrations/seeds/003_demo/trimming_options.csv |   19 +-
 backend/migrations/seeds/003_demo/vaccinations.csv |   16 +-
 backend/migrations/seeds/003_demo/vaccines.csv     |   36 +-
 .../migrations/seeds/003_demo/vital_records.csv    |   32 +-
 .../004_staging/appointment_trimming_details.csv   |    4 +-
 92 files changed, 2110 insertions(+), 5573 deletions(-)
```

## 未完（人間 · §E-7 残り）
- staging-only 4 件の維持/移管/競合 disposition
- migration key/checksum · PlanetScale ownership
- backup/restore · rollback owner
- remote CI 期待定義
- 全 green 後のみ merge-commit PR

```
staging_preflight=PARTIAL agent_readonly=PASS count_4_1346=CONFIRMED merge=NOT_DONE opaque_ref=reports/todo-walk-2026-08-14/staging-preflight-readonly.md
```
