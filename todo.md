# タスク台帳 — 入口

最終照合: 2026-09-15（JST）。ローカル `main` の実装、9月13日のバグ記録、9月15日の医院フィードバック、GitHub / Linear の読取結果を照合。**未完了の作業だけを掲載する。** 実装済みの詳細・完了項目は Git 履歴と元の UAT 記録を参照する。

## 着手プランの確認

2026-09-15、未完了タスクの計画を全件点検した。大枠の方針は既存文書にあり、不足していたコード/手順の入口、前提、実施順、成果物・完了条件を [開発・調査](todo-issue.md)、[検証・受入](todo-verification.md)、[運用・納品](todo-operations.md) に補完した。各表の ID から計画へ進める。`todo-performance.md` は技術記録、バグ/医院報告は元の根拠として参照する。計画があることと、実装・外部実行・受入の完了は区別する。

## 残作業の入口

| 区分 | 次に行うこと | 正本 |
|---|---|---|
| 着手可能な再現・設計 | 治療数量の Enter 確定、治療検索一覧の高さを検討。保存仕様・表示条件を固定してから実装 | [開発・調査キュー](todo-issue.md#open) |
| 医院・PO の入力待ち | マスタ登録経路、排他制御の目的、カルテ表示端末、処置移行の範囲 | [未完了 Issue](todo-issue.md#open) |
| 移行・データ調査 | 性別コード、ワクチン種、未納、死亡データの訂正方針 | [未完了 Issue](todo-issue.md#open) |
| 実装後の確認 | 検索・保険割合・過去カルテ導線の STG ブラウザ確認、全ページ UAT の未確認操作 | [検証 TODO](todo-verification.md#uat-followup) |
| 既存の検証残件 | OWNER 実DB、S09 / V04 / clinical E2E、性能実測、認証 D1 | [検証 TODO](todo-verification.md) |
| 環境・本番・納品 | handoff 再生成、承認後のデータ訂正、Lane 3–4、本番構築・移行・研修 | [運用 TODO](todo-operations.md) |
| deferred | `TASK-444-ADDENDUM-CODEGEN` | [未完了 Issue](todo-issue.md#task-444-addendum-codegen) |

## 管理先と完了の扱い

- 新規の実装・調査・PO 課題は [todo-issue.md](todo-issue.md)。検証・外部作業は各専用 TODO に置き、同じ残件の詳細を複製しない。
- 既存チケットの実行状態は Linear Team **Baritech** / Project **ノア動物病院電子カルテ** / [BRT-4](https://linear.app/baritechllc/issue/BRT-4)。新規 Issue を作らない既存方針を維持する。コメント・状態変更は明示承認後。
- 実装完了と、STG ブラウザ受入・production 配備・go-live を分ける。実装済みの項目は開発キューから削除し、未完了の受入・運用だけを残す。
- 確認元: [医院フィードバック](docs/work/stg-uat-clinic-feedback-q1-q4.md)、[予約・スタッフの UAT 記録](bug.md)、[全ページ UAT 記録](bug-2.md)。元記録の古い計画・当時の状態は現在のキューに読み替えない。

<a id="development-tasks"></a>

## 開発タスク

医院フィードバックのうち **まだ直すもの**。原文と回答は [stg-uat-clinic-feedback-q1-q4.md](docs/work/stg-uat-clinic-feedback-q1-q4.md)。詳細・受入は [todo-issue.md](todo-issue.md#open)。実装済み（検索 AND・保険 50/70・問診抜粋リンク）と手順案内（明日以降の予約編集）は載せない。

| ID | 内容 | 状態 |
|---|---|---|
| [UAT-R2-TREATMENT-COMMIT](todo-issue.md#uat-r2-treatment-commit) | 治療タブ: Enter 1回で確定しない。2回目で数量 PATCH。反映の遅さを減らす | READY（再現・設計） |
| [UAT-R2-MASTER-LIST-HEIGHT](todo-issue.md#uat-r2-master-list-height) | 治療検索の選択肢をスクロールなしで見える高さにする | READY（再現・設計） |
| [UAT-R2-MASTER-PATH](todo-issue.md#uat-r2-master-path) | マスタ入力後に金額が空／会計画面に出ない。登録画面を特定してから直す | BLOCKED（医院入力） |
| [UAT-R2-CHART-FIT](todo-issue.md#uat-r2-chart-fit) | サイドバー展開時にカルテ編集を1画面に収める | BLOCKED（端末条件） |
| [UAT-R2-EXCLUSIVE-LOCK](todo-issue.md#uat-r2-exclusive-lock) | 旧システム相当の「他PC入力禁止」は未実装。全面ロックは製品判断後 | PO 判断待ち |
| [UAT-Q3-GENDER-MAP](todo-issue.md#uat-q3-gender-map) | 避妊去勢済みが性別不明。コード 3/4 を雄/雌へ。STG 訂正は運用 | 未完了（old_db・運用） |
| [UAT-Q2-VACCINE-SPECIES](todo-issue.md#uat-q2-vaccine-species) | 猫に犬用ワクチン（Proheart・6種等）。件数調査のあと種を付ける | 調査待ち |
| [UAT-Q4-UNPAID-TRIAGE](todo-issue.md#uat-q4-unpaid-triage) | 未納はデモではない。実未納と突合漏れを集計で切る。一括完了しない | 調査待ち |
| [UAT-Q2-TREATMENTS-IMPORT](todo-issue.md#uat-q2-treatments-import) | 旧カルテの処置・処方明細が未移行 | DEFERRED（PO 未決） |

先行するのは `UAT-R2-TREATMENT-COMMIT` と `UAT-R2-MASTER-LIST-HEIGHT`。データ運用や入力待ちを、無条件の実装 READY にしない。

<a id="product-bugs"></a>

## 確認済み製品 FAIL

現在の未解消報告は [todo-issue.md](todo-issue.md) の症状別調査・データ課題に整理した。`bug.md` / `bug-2.md` の FIXED / SPEC-OK を未修正の製品 FAIL として再登録しない。環境・fixture 不足は運用、未確認操作は検証に置く。新たに製品欠陥が確定したら、既存 ID のまま再現・原因・担当範囲を確定する。

<a id="human-lane"></a>

## PO / 人間レーン

医院入力待ち・製品判断は [todo-issue.md](todo-issue.md#open)、実環境の承認と操作は [todo-operations.md](todo-operations.md)、受入・go-live は [todo-verification.md](todo-verification.md)。既存の人間ゲートは BRT-4 配下で追跡する。

<a id="refactor-constraints"></a>

## FE 維持制約

- `design-tokens.ts` / `query-keys.ts` / `paths.ts` の表分割や行数だけを目的とした機械的分割をしない。
- `utils/` 再作成・generated/models 一括移行をしない。必要性は [裁定記録](docs/work/development-task-decisions.md#task-444) に従って判断する。
- `app/pages` の合成と owners `loaders.ts` の例外、権限 ref、死亡 sentinel、`useActionState`、queryKey タプルを維持する。
- FE12 却下（manual chunk、死亡行グレーアウト、owners 行アクションをペット生死で止める）を維持する。

着手時は [AGENTS.md](AGENTS.md) の claim・worktree 規則に従う。秘密・患者情報は台帳へ書かない。
