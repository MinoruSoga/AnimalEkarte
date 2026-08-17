# staging preflight status（2026-08-15 · #299 MERGED）

## 状態サマリ

| 項目 | 結果 |
|------|------|
| PR #299 | **MERGED** https://github.com/MinoruSoga/AnimalEkarte/pull/299 |
| merge commit | `85bb651ee744acf98b1c4239a3497b08737e32ba` |
| mergedAt | 2026-08-15T07:54:24Z |
| method | **merge-commit**（squash なし） |
| main tip at merge | `1836378d0` |
| CI at merge | green |
| USER 残セル | 推奨案採用 · `PR299-MERGE-GATE-SIGNED-2026-08-15.md` |
| backup pre-merge | **N**（NOT_TAKEN · rollback=CF last-known-good · owner=USER） |
| STG reset | **禁止継続** |

## 残（merge 後）

- [ ] Deploy / stg-smoke 確認
- [ ] `schema_migrations` read-only 検証（RUNBOOK）
- [ ] Linear BRT-55 に post-verify 結果
- [ ] BRT-68 H1/H2 は STG health 後

```
staging_preflight=MERGED pr=299 merge=85bb651ee tip=1836378d0 post_verify=PENDING
```
