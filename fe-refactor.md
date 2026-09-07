# frontend コード規約監査 — 完了履歴

2026-09-04 の第3期監査と実装の記録。FE-RC-201〜228 は FIXED、FE-RC-W3-LEAVE は CLOSED。当時の残 FINDING は 0。
これは過去の検証結果であり、現在の全ファイルの再監査・UAT・release 判定を示さない。未完了作業の入口は [`todo.md`](todo.md)、実行状態の正本は Linear（[台帳ルール](docs/work/README.md)）。

## 証跡

- 対象: 監査時点の frontend tracked 2017 パス。監査時分類は PASS 1786 / EXCLUDE 125 / FINDING 106。実装後に対象 FINDING を解消。
- FE-RC-W3-LEAVE: Docker scoped Vitest 4 files / 27 tests passed、E2E TypeScript check exit 0、filename ratchet 0。
- 詳細な指摘、13 レーンの分類、修正内容、検証コマンドは次の固定履歴を参照する。

```bash
git show d9a757cc09386a42fb3288cd16fd8ae5966feaf3:fe-refactor.md
```

## 維持する判断

- 第1期 FE-RC-001〜089 / 第2期 FE-RC-101〜128 の FIXED を本履歴から再オープンしない。
- FE-RC-117 `use-reservation-type-color-map` の `@/hooks` 正本を維持する。
- `useInventoryList` の `useGet` 改名は FE-RC-055 派生 facade 例外として見送る。
- 公開型のない feature に空の `types/` / `hooks/` を新設しない。
- 却下済みの manual chunk、死亡行グレーアウト、一覧行アクション制限を再提案しない。再開条件は [`frontend/CLAUDE.md`](frontend/CLAUDE.md) を参照する。
