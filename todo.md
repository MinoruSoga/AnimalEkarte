# タスク台帳 — Linear が正本

統合日: 2026-09-08（下記の作業状態・外部観測は各記録時点のまま。今回の統合では再判定していない）

| 項目 | 値 |
|------|-----|
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **セキュリティ修正** | **[BRT-226](https://linear.app/baritechllc/issue/BRT-226)**（Review · `origin/main` 済み · Done は人間） |
| **本ファイルの範囲** | repo と強く結び付く **未完了作業の入口**、確認済み製品 FAIL、PO 入口、分離した完了履歴・維持制約 |

状態・Done は Linear を正本とする。行値・秘密は書かない。完了項目は実行キューから除く。以下の履歴は当時の証跡への入口であり、現在の受入・release 判定ではない。

`bug.md`・`todo-now.md`・`todo-po.md`・`todo-refactor.md` は本ファイルへ統合して削除した。別台帳として再作成しない。`todo-fix-auth.md` 等、今回指定外の文書は統合・削除していない。

入口: [実行キュー](#対応順実行キュー) · [製品 FAIL](#product-bugs) · [PO / 人間レーン](#human-lane) · [Astra 完了履歴](#astra-history) · [FE 完了履歴・維持制約](#refactor-history)

エージェントは PlanetScale、共有 STG apply、`DROP SCHEMA`、本番 cutover、`make reset`、八王子 CSV の producer 出力を実行しない。push / dispatch / Linear Done / 秘密変更は明示承認が必要。

claim は ID ごとに初回編集前に取得する。エージェントは claim を削除しない。

この統合の claim: `claim/LEDGER-TODO-CONSOLIDATE`（USER のみ解除）。旧台帳更新時の claim 記録: `claim/LEDGER-TODO-PRUNE`。旧 claim の現在状態は本記録では断定しない。

---

## 対応順（実行キュー）

上から 1 件だけ着手する。USER / old_db に当たったら止めて提示する。deferred はキューに入れない。

| 順 | ID | 実行者 | なぜこの順 | 状態 |
|----|----|--------|------------|------|
| 1 | **META-LINEAR-APPLY** | USER | repo の対応案は [linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md)。書き込みと Done は USER | **BLOCKED**（Linear MCP / `LINEAR_API_KEY` なし。公開ページはログイン壁。エージェントは書かない） |
| 2 | **H0-2 / HAC-CSV-1** | old_db / USER | STG 八王子の先頭。これより前の STG 行は進めない | **BLOCKED**（HAC-INPUT-2。完全 KNJO 未受領。同一 BAK 再実行と producer は禁止） |
| 3 | **H0-3b → Lane3 HAC → H3-9 → H3-11 → Lane 4** | USER | 2 の依存どおり | 待ち |
| 4 | **P1 → P2 → P3 → P4 → P5 → P6 → P7 → P8** | USER | go-live 依存。E1 / E2 は P4 の一部 | 待ち |

次は医院/ベンダーからの完全 KNJO 再取得、または城東主経路（JOU-G2-2 の Azure 承認）。H0-3b には入らない。Linear 書き込みと Done は USER。

---

## 1. 受入残（PASS にしていない）

helper / 再実行スライスは済。UAT / E2E を PASS にしない。正本は [UAT-DOMAIN-STATUS.md](docs/ops/testing/UAT-DOMAIN-STATUS.md)。

| ID | 残 | 状態 |
|----|----|------|
| **QA-UAT-S09-FIXTURE** | HTTP/CLI とブラウザ再実行 | S09 は BLOCKED |
| **QA-UAT-V04-RETEST** | live HTTP は 403。clinic 1/2 の権限昇格なし | V04 は UNKNOWN |
| **QA-FULL-CLINICAL-E2E** | `--clinical` 未実行。e2e.yml job は未 | E2E は未証明 |

設計: [S09-FIXTURE-DESIGN.md](docs/ops/testing/S09-FIXTURE-DESIGN.md) · [CLINICAL-E2E-DESIGN.md](docs/ops/testing/CLINICAL-E2E-DESIGN.md)。

---

## 2. USER ゲート（秘密・本番・外部環境）

外部状態は実行直前に再確認する。エージェントは秘密値の作成・表示・投入、共有 STG/PROD apply、production 構築、go-live を自動実行しない。

| 順 | ID | 実行者 | 状態 | 完了条件 |
|----|----|--------|------|----------|
| P1 | **SEC-SECRETS-5 / #89 / #97** | USER | 4系統 rotation receipt 未記入 | 新発行→投入→再 deploy→health→旧値 revoke→旧値拒否。値は記録しない |
| P2 | **#253 / U12 PROD-SETUP** | USER / 開発 | Production 未構築 | Cloudflare 本番、Required reviewers、workflow、rollback、backup rehearsal、URL/CI receipt |
| P3 | **#250 PROD-DATA-MIGRATION** | USER / 開発 | 事前準備待ち | rehearsal、最終 import、入力停止、backup/rollback、件数・clinic_id・金額突合 |
| P4 | **#254 AUTHENTICATED-UAT** | USER / agent | full UAT 未証明 | 全業務 scenario の受入結果を確定。PARTIAL / BLOCKED / UNKNOWN を PASS にしない |
| P5 | **#255 STAFF-PROVISION** | USER | 入力未記入 | roster、email 方針、clinic、role、actor、環境承認。PII-free receipt |
| P6 | **#258 / U1〜U12 DELIVERY** | USER | 最終承認待ち | P1・P2 と契約責任者の非機密事実を `DELIVERY_PACKAGE.md` へ反映 |
| P7 | **#256 / U13 TRAINING** | USER | 操作説明会未完 | 日程・形式・範囲・結果・opaque receipt |
| P8 | **#257 GOLIVE** | USER | HOLD | P1〜P7 の受入と第2段階条件。重大 FAIL や当日 import 未達なら No-Go |
| E1 | **QA-UAT-LSTEP-REAL** | USER | 外部環境待ち | write 有効な LSTEP で S01 同期と V05-17 remove |
| E2 | **QA-UAT-LINE-IDTOKEN** | USER | mock 外・未証明 | 実 LINE idToken で link / 409 / 期限切れ 400 |

P4 の延期例外: 臨床安全、会計金額、clinic / owner / pet / staff 分離、認証・権限、データ消失の未解消 FAIL は go-live 前に解消する。それ以外は Linear に受容条件を残し、USER の明示受容がある場合だけ延期できる。

---

## 3. STG 実データ（USER / old_db）

対象は八王子 `clinic_id=1` と城東 `clinic_id=2`。証跡が無いだけなら **UNKNOWN**。再 apply の前に USER が現行状態を確認する。

| 優先 | ID | 実行者 | 状態 | blocked-by |
|------|----|--------|------|------------|
| 1 | **H0-2 / HAC-CSV-1** | old_db / USER | HAC-CSV-1 は C1 `HAC-INPUT-2` 待ち。AE `hachioji/` 空。export なし | CHECKDB clean な新規 BAK、または全32列完全 KNJO。同一/既知破損 BAK の再復元は禁止。城東 live を上書きする load も禁止 |
| 2 | **H0-3b / H1-2** | USER | 待ち | H0-2 |
| 3 | **AE-STG-UAT-LANE3-HAC** | USER | UNKNOWN・投入判断待ち | 現行状態、H0-2、H0-3b |
| 4 | **H3-9 staff attach apply** | USER | 入力あり・apply 実施有無 UNKNOWN | 現行 attach と STG 実行ゲート |
| 5 | **H3-11 画面確認** | USER | UNKNOWN・証跡未取得 | H3-9 と自医院ログイン |
| 6 | **Lane 4** | 医院スタッフ / USER | 完了未証明 | 両院 Lane 3 verify、H3-11 |

索引から外したもの: AE-OLD-DB-MR-UNIQ、Lane3 城東 21表、H3-7 敷島 / Hako。

STG 実行ゲート: 対象環境、data owner、operator、maintenance window、backup / restore、rollback、承認。正本は [STG 手順の停止ゲート](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md#2-pre-deploy-stop-gates)。

観測（2026-09-06）: 城東・敷島・箱は AE handoff に bundle あり。八王子ディレクトリは空。old_db export も八王子 run なし。producer は未実行。Lane 4 は 5営業日証跡なし。

---

## 4. 触ったときだけ / deferred

横断キャンペーンにしない。

| ID | 再開条件 |
|----|----------|
| **TASK-444** | generated/models の公開契約・codegen・consumer 移行計画が揃ってから |
| **BE-RC-005** | 新規・変更 service から 5xx 二重ログを解消 |
| **BE-RC-009** | 新規 consumer または対象機能変更時に利用側最小 port へ分割 |
| **BE-RC-014** | typed error が使えるようになったら `errors.As` へ |
| **BE-RC-015** | 新規・変更面から package.Type stutter を避ける |
| **BE-RC-017** | 対象 repository 変更時に unexported update + typed command |
| **BE-RC-019** | lab / hospitalization 等の境界が成立する変更時だけ |

---

<a id="product-bugs"></a>

## 5. 確認済み製品 FAIL（旧 bug.md）

記録対象は確認済み製品 FAIL のみ（[TEST_ARCHITECTURE.md](docs/ops/testing/TEST_ARCHITECTURE.md) §6）。環境・seed・権限・fixture 不足による BLOCKED / PARTIAL や受入未実施を混ぜない。証跡に credential・token・cookie・idToken・個人情報（PHI）を含めない。起票後は Linear で追跡し、見出し ID は本節内で重複させない。新規項目は `### BUG-XXX` で本節に追加する。

| ID | status | area | severity | scenario | 層 |
|:---|:---|:---|:---|:---|:---|
| （2026-09-06 の旧台帳では未対応なし） | — | — | — | — | — |

これは旧台帳の記録であり、現在の製品全体に不具合がないという判定ではない。認証レビューの修正計画は [todo-fix-auth.md](todo-fix-auth.md) を参照する（本統合の対象外）。対応済み項目は本節の未対応一覧から除き、履歴は Git と `reports/uat-YYYY-MM-DD/` を参照する。

旧台帳の解消記録（2026-09-06）:

- S06 UAT の auto-draft FAIL は、general reservation type id=5 による V01 再確認 PASS 後に除外。証跡識別子: `reports/uat-2026-09-06/v01/`。
- console-full-sweep の `BUG-20260906-001..004` は実装対応済みとの記録。詳細は Git。当時の Linear 照会は未実施。

<a id="human-lane"></a>

## 6. PO / 人間レーン（旧 todo-po.md）

実行 SoT は Linear hub [BRT-4](https://linear.app/baritechllc/issue/BRT-4) · Project ノア動物病院電子カルテ。人間ゲート（UAT・PO 確認）も Linear を正本とし、別の Open 行台帳を再構築しない。本ファイルの受入残・USER ゲートを入口にする。

会社側索引: CorpVault `50_Projects/ノア動物病院電子カルテ/05_Linearマップ.md`。旧詳細本文は Git 履歴。

<a id="astra-history"></a>

## 7. Astra 品質監査 — 完了履歴（旧 todo-now.md）

以下は旧台帳の観測記録。Astra 監査 F1〜F6 の実装は当時 `origin/main` へ統合済み。未完了の後続は本ファイルの実行キューを入口とし、実行状態の正本は Linear（[作業台帳ルール](docs/work/README.md)）。

- 監査対象: `main` / `c41ba8b1c1aef7f8150457dc310fe7f61fca1a75`（2026-09-05）。
- 最終観測: 2026-09-06 / `b987729fd57b05b5f94f0a7e0d5860d515401421`。
- 旧更新時は `claim/TODO-NOW` が不在だったため、`todo.md` の `LEDGER-TODO-NOW-POINTER` として入口ポインタ化した。現在の claim 状態を示すものではない。
- F1〜F6 の実装と F3 manual E2E auth smoke（run `33972458396`）。
- F4〜F6 の push 後通常 CI。
- F3 k6 aggregate の原因は `metric.values.*` のみ参照。修正の追跡 ID は `CI-K6-SUMMARY-SCHEMA`（当時の `todo.md`、詳細は Git）。
- 「診断カテゴリ」「診断病名」label 接続の追跡 ID は `FE-CLINICAL-PLAN-SELECT-LABELS`。

Linear の F1〜F6 対応案は [linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md)。旧台帳時点の Linear 本体は UNKNOWN、Done は USER。今回の統合では外部状態を再照会していない。

<a id="refactor-history"></a>

## 8. FE リファクタ — 完了履歴・維持制約（旧 todo-refactor.md）

2026-09-02 の FE 規約ゲート完了記録。対象は `frontend/src`・`frontend/liff/src`・`frontend/line-reserve/src` の production TS/TSX。現在の未完了作業は本ファイル、実行状態の正本は Linear（[台帳ルール](docs/work/README.md)）。

### 当時の証跡

- 当時の完了結果: cross-feature deep import、queryKey 配列リテラル、生産の `window.confirm` を対象範囲で解消。150 行超の生産関数は 0、800 行超ファイルは意図的残置の `design-tokens.ts` のみ。
- Docker scoped Vitest: 第1束 16 files / 400 passed、第2束 11 files / 85 passed。
- 2026-09-02 にユーザーが `make lint-front` / `make test-front` の成功を報告。`pnpm build` は未実行。
- 上記は当時の観測であり、現在の runtime・UAT・release 判定ではない。詳細なカテゴリ別変更、検証、claim の解放記録は固定履歴を参照する。ここでは claim の現在状態を断定しない。

```bash
git show 12f15fe4fc85f3e9900f70575277b5ab5bc645e4:todo-refactor.md
# BE フェーズの完了時本文
git show ad63bdf28:todo-refactor.md
```

### 維持する対象外・制約

- `design-tokens.ts` / `query-keys.ts` / `paths.ts` の表分割、50 行までの機械分割、200–399 行ファイルの薄型化だけを目的とした切断は行わない。
- `utils/` を再作成しない。generated/models の一括移行は TASK-444 の別トラック。当時の allowlist 267 件は分割追従のみ。
- `app/pages` の合成と owners `loaders.ts` の例外を維持する。
- 権限 ref、死亡 sentinel、`useActionState`、queryKey タプルの契約を維持する。
- FE12 却下（manual chunk、死亡行グレーアウト、owners 行アクションをペット生死で止める）は維持する。
- 当時のトリミングフォームの権限・死亡ガード欠落は対象外だった。本履歴から現在の未修正・修正済みを判断しない。

---

## 参照

| 文書 | 役割 |
|------|------|
| [Astra 完了履歴](#astra-history) | Astra F1〜F6 の完了履歴 |
| [docs/work/linear-f1-f6-mapping.md](docs/work/linear-f1-f6-mapping.md) | F1〜F6 対応案。Linear は UNKNOWN |
| [docs/ops/testing/S09-FIXTURE-DESIGN.md](docs/ops/testing/S09-FIXTURE-DESIGN.md) | S09 helper 設計 |
| [docs/ops/testing/CLINICAL-E2E-DESIGN.md](docs/ops/testing/CLINICAL-E2E-DESIGN.md) | clinical E2E 設計 |
| [docs/ops/testing/UAT-DOMAIN-STATUS.md](docs/ops/testing/UAT-DOMAIN-STATUS.md) | UAT 集計の正本 |
| [製品 FAIL](#product-bugs) | 確認済み製品 FAIL |
| [docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md) | ローカル handoff |
| [docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) | STG 破壊境界 |
