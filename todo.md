# タスク台帳 — 入口

> Task migrated to Plane `EMR-199`. This file remains supporting acceptance/evidence material; use Plane for current status.

最終照合: 2026-09-23。未完了タスクと状態はPlaneが正本。以下の旧HEAD・コード/UAT照合内容は当時の根拠であり、現在の配備・受入状態を示さない。移行対応は [Plane移行記録](docs/work/plane-md-migration-20260923-receipt.md)。

## Plane task tracking

Current unfinished implementation, verification, data, performance, delivery, and human-gated work is tracked in Plane. The local task text from this file and the canonical ledgers was migrated on 2026-09-23; the source-to-Plane crosswalk is [the migration receipt](docs/work/plane-md-migration-20260923-receipt.md). Historical UAT evidence and operator procedures remain in their source documents. Migration preserves each recorded status and does not imply implementation, runtime acceptance, or closure.

<a id="development-tasks"></a>
<a id="product-bugs"></a>

## PO / 人間レーン

Plane がタスク状態の正本です。人手入力待ち、旧Chrome受入、実環境操作、go-live の個別ゲートは移行先チケットに保持しました。詳細な移行対応は [移行記録](docs/work/plane-md-migration-20260923-receipt.md) を参照してください。

<a id="refactor-constraints"></a>

## FE 維持制約

- `design-tokens.ts` / `query-keys.ts` / `paths.ts` の表分割や行数だけを目的とした機械的分割をしない。
- `utils/` 再作成・generated/models 一括移行をしない。必要性は [裁定記録](docs/work/development-task-decisions.md#task-444) に従って判断する。
- `app/pages` の合成と owners `loaders.ts` の例外、権限 ref、死亡 sentinel、`useActionState`、queryKey タプルを維持する。
- FE12 却下（manual chunk、死亡行グレーアウト、owners 行アクションをペット生死で止める）を維持する。

着手時は [AGENTS.md](AGENTS.md) の claim・worktree 規則に従う。秘密・患者情報は台帳へ書かない。
