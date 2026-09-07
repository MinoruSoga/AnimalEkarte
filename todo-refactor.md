# FE リファクタ — 完了履歴

2026-09-02 の FE 規約ゲート完了記録。対象は `frontend/src`・`frontend/liff/src`・`frontend/line-reserve/src` の production TS/TSX。現在の未完了作業は [`todo.md`](todo.md)、実行状態の正本は Linear（[台帳ルール](docs/work/README.md)）。

## 証跡

- 当時の完了結果: cross-feature deep import、queryKey 配列リテラル、生産の `window.confirm` を対象範囲で解消。150 行超の生産関数は 0、800 行超ファイルは意図的残置の `design-tokens.ts` のみ。
- Docker scoped Vitest: 第1束 16 files / 400 passed、第2束 11 files / 85 passed。
- 2026-09-02 にユーザーが `make lint-front` / `make test-front` の成功を報告。`pnpm build` は未実行。
- 上記は当時の観測であり、現在の runtime・UAT・release 判定ではない。詳細なカテゴリ別変更、検証、claim の解放記録は固定履歴を参照する。ここでは claim の現在状態を断定しない。

```bash
git show 12f15fe4fc85f3e9900f70575277b5ab5bc645e4:todo-refactor.md
# BE フェーズの完了時本文
git show ad63bdf28:todo-refactor.md
```

## 維持する対象外・制約

- `design-tokens.ts` / `query-keys.ts` / `paths.ts` の表分割、50 行までの機械分割、200–399 行ファイルの薄型化だけを目的とした切断は行わない。
- `utils/` を再作成しない。generated/models の一括移行は TASK-444 の別トラック。当時の allowlist 267 件は分割追従のみ。
- `app/pages` の合成と owners `loaders.ts` の例外を維持する。
- 権限 ref、死亡 sentinel、`useActionState`、queryKey タプルの契約を維持する。
- FE12 却下（manual chunk、死亡行グレーアウト、owners 行アクションをペット生死で止める）は維持する。
- 当時のトリミングフォームの権限・死亡ガード欠落は対象外だった。本履歴から現在の未修正・修正済みを判断しない。
