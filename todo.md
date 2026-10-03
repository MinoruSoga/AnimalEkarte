# タスク台帳 — 入口

> Task migrated to Plane `EMR-199`. This file remains supporting acceptance/evidence material; use Plane for current status.

最終照合: 2026-09-27（Plane 状態とコード・git履歴を照合し、Done/Cancelled/Duplicate かつ対応コミットまたは受入記録を確認できた項目を削除: EMR-90/92/94/101/102/103、EMR-135/137/138/192）。未完了タスクと状態はPlaneが正本。以下の旧HEAD・コード/UAT照合内容は当時の根拠であり、現在の配備・受入状態を示さない。移行対応は [Plane移行記録](docs/work/plane-md-migration-20260923-receipt.md)。 2026-09-27 追記: Plane 移行済みの項目行・`移行済み` スタブ・対応済みの状態表を本文から削除し、旧 ID アンカーは外部リンク維持のため「旧タスク ID」節へ集約した。同日再照合（2回目）: 全参照 ID の Plane 現行状態を確認し、Done 到達の EMR-84/85/87/120/121/122/123/124/190 をトピック対応表へ反映、EMR-136/204 の Cancelled 終了と EMR-142/143/203 の Ready 遷移を記録した。

本ファイルは旧 `todo-issue.md`・`todo-verification.md`・`todo-operations.md`・`todo-performance.md` を 2026-09-26 に統合した。節: [Issue 出典・仕様履歴](#issue-ledger) · [検証・受入](#verification-ledger) · [運用・外部実行](#operations-ledger) · [性能](#performance-ledger)。統合前の行番号参照（`L175` 等）は Git の旧ファイルを指す。

## Plane task tracking

Current unfinished implementation, verification, data, performance, delivery, and human-gated work is tracked in Plane. The local task text from this file and the canonical ledgers was migrated on 2026-09-23; the source-to-Plane crosswalk is [the migration receipt](docs/work/plane-md-migration-20260923-receipt.md). Historical UAT evidence and operator procedures remain in their source documents. Migration preserves each recorded status and does not imply implementation, runtime acceptance, or closure.

<a id="development-tasks"></a>
<a id="product-bugs"></a>

## 確認済み製品 FAIL

移行前の一覧は Plane `EMR-199` へ移行済み（[移行記録](docs/work/plane-md-migration-20260923-receipt.md)）。本節には移行後に新規確定した UAT 製品 FAIL だけを記録し、Linear/Plane intake は別途行う。

### UAT 2026-09-25 確定分（証拠: `reports/uat-2026-09-25/tickets/EMR-87.md`）

| ID | severity | 領域 | 症状 | シナリオ | Plane / 修正PR |
|:---|:---|:---|:---|:---|:---|
| BUG-INQUIRY-DEFAULT-TEXT-NOT-SENT | Low | medical-records / 問診 | 保存済み主訴本文を定型文(DEFAULT_CHIEF_COMPLAINT)と完全一致へ戻すと `chief_complaint` が未送信(undefined)になり旧値が残る。`notes`/治療方針も同型（use-medical-record-save-action.ts 問診分岐） | S27 再検証中に観測 | EMR-215 / PR #498（修正マージ済み・UAT中） |

### UAT 2026-10-01 確定分（証拠: `reports/uat-2026-10-01/s15-results.md` / `s15-evidence.jsonl`）

| ID | severity | 領域 | 症状 | シナリオ | Plane / 修正 |
|:---|:---|:---|:---|:---|:---|
| BUG-ACCOUNTING-INSURANCE-FORM-RESET | High | accounting / 確定済み会計修正 | 確定済み会計で「修正を保存する → 修正する」を実行すると、React 19 の post-action `form.reset()` が Radix Switch を mount 時値(false)へ巻き戻し、UI は保険 ON 表示のまま PATCH が `has_insurance:false / billing_amount:1100` を送信して BE 400。受理されれば保険をサイレント消去する不具合だった | S15 手順6 | 修正済み commit `df8bc796e`（InsuranceCard Switch/Select に `key={現在値}` で remount）。系統的リスク棚卸しは EMR-252 |
| BUG-ACCOUNTING-WAITING-FINALIZE-DEADEND | Medium | accounting / waiting billing 詳細 | `status=waiting` の billing 詳細画面に「会計を確定する」が表示され PATCH を送信するが、BE は waiting→completed PATCH を 400 で必ず拒否（正規導線は `/accounting/new` → `POST /accountings/complete`）。到達可能な行き止まり | S15 fixture 調査中に観測 | EMR-253 |

### UAT 2026-10-01 確定分 S16（証拠: `reports/uat-2026-10-01/s16-results.md` / `s16-evidence.jsonl` / `s16-demo-rows-fail-excerpt.txt`）

| ID | severity | 領域 | 症状 | シナリオ | Plane / 修正 |
|:---|:---|:---|:---|:---|:---|
| BUG-INTERVIEW-HISTORY-DEMO-ROWS | High | medical-records / 問診タブ治療履歴 | 履歴 API が 0 件のペットでも `DEFAULT_HISTORY_ITEMS`（ハードコードのデモ3行）が治療履歴に表示され、各行が `/medical-records/{1,2,3}` への実リンクとして機能し無関係なカルテへ誤誘導する | S16 履歴0件観測 | EMR-254 / 修正済み（`MedicalRecordInterview` のフォールバック除去。回帰ガード: `s16-interview-history-navigation.spec.ts` 履歴0件テスト） |

### codex-security スキャン累積 findings トリアージ（2026-10-03）

累積 56 findings（CLI 表示 51 件・dedupe 後、frontend scan `yxoHGs` + backend scan `Vy74Z7`、スキャン対象 `01fdcef`）を HEAD `fd65d1b2b` で照合。**対応済み 48 / 本日修正実施 6 / 製品判断レーン 4 / 部分対応 2**。全件証跡は [トリアージ報告](docs/work/security-triage-2026-10-03.md)。Plane intake は別途行う。

| ID | severity | 領域 | 症状 | 区分 | 対応先 |
|:---|:---|:---|:---|:---|:---|
| SEC-O1 | Medium | inventory / 処置数量 | `DecreaseStock` が `int(quantity)` 切捨て。数量0.5の処置で在庫減算が0になり減算を回避できる | 対応済み(未コミット) | `DecreaseStock` 非整数拒否 + Create 時 InventoryID 連動は整数必須化。テスト追加済み |
| SEC-O2 | Medium | LIFF / body limit | liff グループに body limit なし。`application/octet-stream` ラベルの JSON で生JSON上限を回避 | 対応済み(未コミット) | `liffBodyLimit` 注入 + `liff.Use` で全 LIFF route に 1MiB cap |
| SEC-O3 | Medium | CI / actions pin | credential 到達 job を含む全 workflow が v-tag 参照(`checkout@v7.0.1`, `setup-node@v7`, `chromaui/action@v18.9.5` 等) | 対応済み(未コミット) | `peaceiris/actions-gh-pages`・`chromaui/action` を SHA pin |
| SEC-O4 | Medium | CI / k6 install | `performance-tests.yml` が `curl \| apt-key add` + 無版 `apt-get install k6` | 対応済み(未コミット) | k6 v2.3.0 pin + SHA-256 検証インストール |
| SEC-O5 | Low | lstep-migrate / CSV | `cmd/lstep-migrate` が raw csv.Writer で owner_name/error_message 出力。数式注入のまま | 対応済み | `reporter.go` に `sanitizeCSVCell` 適用済み(既存コード・初回見落とし訂正) |
| SEC-O6 | Low | medical-records / upload | 宣言 MIME allowlisted なら拡張子不検査 + 非S3時 `/uploads` 無認証 StaticFS（非release環境限定の実害） | 対応済み(未コミット) | 拡張子↔MIME一致必須化 + content-sniff 照合 + 保存拡張子を検証済みmimeから導出 + `/uploads` nosniff |
| SEC-O7 | Medium | CI / secret-sync | `wrangler@4.107.0` 版 pin 済みだが integrity 検証なし（部分残存） | 対応済み(未コミット) | `pnpm install --frozen-lockfile` + `pnpm exec wrangler`(lockfile integrity) |
| SEC-O8 | High | seedlogin / STG exec | STG/local に repository-public `SharedPassword="password"` の全医院 executive を provision。production/unknown は拒否済み | 対応済み(WIP) | staging は `SEEDLOGIN_DEMO_PASSWORD` シークレットのみ受付・未設定 fail-closed。repo-public は local 限定(未コミット WIP) |
| SEC-P1 | High | support / cross-clinic board | bug-reports + chat-exchanges が全医院共有（3 occurrence 統合）。mutation(status/Plane起票/削除)は WIP で報告元clinic絞り済み。read 共有は owner 承認済みの製品決定（SECURITY.md:116） | 受容済み | 情報区分・外部送信条件の未決残件は EMR-263 |
| SEC-P2 | High | seedlogin / staging credential | O8 と同根。repo-public password 問題は WIP で解消。STG デモ運用は SECURITY.md に実測記載済み | 受容済み | 対応完了 |
| SEC-P3 | High | RBAC / 拠点横断 #86 | GET fallback opt-in 化・write selected-clinic 必須は修正済み。`ResolveListClinicIDs` の membership 拡張は #86 配送済み設計 | 受容(意図仕様) | 対応完了 |
| SEC-P4 | Medium | support / chat unredacted | P1 の一部。会話本文の全院共有 | 受容済み | P1 と同じ（EMR-263 で情報区分追跡） |

<a id="human-lane"></a>

## PO / 人間レーン

Plane がタスク状態の正本です。人手入力待ち、旧Chrome受入、実環境操作、go-live の個別ゲートは移行先チケットに保持しました。詳細な移行対応は [移行記録](docs/work/plane-md-migration-20260923-receipt.md) を参照してください。

<a id="refactor-constraints"></a>

## FE 維持制約

- `design-tokens.ts` / `query-keys.ts` / `paths.ts` の表分割や行数だけを目的とした機械的分割をしない。
- `utils/` 再作成・generated/models 一括移行をしない。必要性は [裁定記録](docs/work/development-task-decisions.md#task-444) に従って判断する。
- `app/pages` の合成と owners `loaders.ts` の例外、権限 ref、死亡 sentinel、`useActionState`、queryKey タプルを維持する。
- FE12 却下（manual chunk、死亡行グレーアウト、owners 行アクションをペット生死で止める）を維持する。

着手時は [AGENTS.md](AGENTS.md) の claim・worktree 規則に従う。秘密・患者情報は台帳へ書かない。

---

<a id="issue-ledger"></a>

<a id="未完了-issue-台帳repo-正本"></a>

> Task migrated to Plane `EMR-198`. This file remains supporting acceptance/evidence material; use Plane for current status.

## Issue出典・仕様履歴（未完了タスク状態はPlane）

最終タスク移行: 2026-09-23。未完了の実装・調査・PO課題の状態はPlaneを正本とし、このファイルはSlack出典と確定仕様の根拠記録として保持する。完了した実装・仕様どおりの報告・回答済みの操作案内は 未対応エリアに戻さず、[Slack 出典対応表](#slack-source-map) に処理区分を残す。検証は [todo-verification.md](#verification-ledger)、外部実行は [todo-operations.md](#operations-ledger) を正本とする。PO/evidence のローカル判断材料は [remaining-campaign-20260920](docs/work/remaining-campaign-20260920/) に作成済み。PO裁定・実機/STGの新しい結果は今回未照合（UNKNOWN）。

Linear/BRT-4に関する以下の文言は過去の運用方針記録。2026-09-23のタスク移行後、現在の未完了タスク状態はPlaneを参照する。

このファイルはSlack出典・仕様確定の履歴を保持する。未完了タスクと現在状態はPlaneにあり、対応は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照する。

### 旧タスク ID（全件 Plane 移行済み）

この台帳の未完了タスク本文は Plane へ移行しました。現在状態と全 ID の対応（crosswalk・source SHA）は[移行記録](docs/work/plane-md-migration-20260923-receipt.md)を参照してください。docs/work 等からの `todo.md#<id>` リンク維持のため、旧 ID のアンカーのみ残置します（本文なし）。

<a id="open"></a><a id="slack-tasks"></a><a id="slack-全件からの着手プラン2026-09-19"></a><a id="area-ready"></a><a id="area-po"></a><a id="area-evidence"></a><a id="area-deferred"></a><a id="area-resolved"></a>
<a id="uat-q2-treatments-import"></a><a id="slack-manual-urine"></a><a id="slack-cross-clinic"></a><a id="slack-vaccine-print"></a><a id="slack-deceased"></a><a id="slack-latency"></a><a id="slack-vaccine-multi"></a><a id="uat-q3-gender-map"></a><a id="uat-q2-vaccine-species"></a><a id="uat-q4-unpaid-triage"></a><a id="po-pet-deceased-data-backfill"></a><a id="slack-billing-uat"></a><a id="slack-clinical-uat"></a><a id="slack-access"></a><a id="slack-uat-schedule"></a><a id="slack-hac-import"></a><a id="slack-lab"></a><a id="slack-exam-history"></a><a id="slack-reservation-reference"></a><a id="slack-backlog61"></a><a id="slack-staff-select"></a><a id="slack-intake"></a><a id="slack-ocr"></a><a id="slack-smaregi"></a>
<a id="todo-v-v04-qa-uat-v04-retest"></a><a id="todo-v-v04"></a><a id="qa-uat-v04-retest"></a><a id="todo-v-clinical-e2e-qa-full-clinical-e2e"></a><a id="todo-v-clinical-e2e"></a><a id="qa-full-clinical-e2e"></a><a id="todo-v-stg-data"></a><a id="todo-v-release"></a><a id="p4-254-authenticated-uat"></a><a id="p4"></a><a id="254-authenticated-uat"></a><a id="e1-qa-uat-lstep-real"></a><a id="e1"></a><a id="qa-uat-lstep-real"></a><a id="e2-qa-uat-line-idtoken"></a><a id="e2"></a><a id="qa-uat-line-idtoken"></a><a id="p8-257-golive"></a><a id="p8"></a><a id="257-golive"></a><a id="auth-v-d1-preflight"></a><a id="auth-v-d1-apply"></a><a id="auth-v-d1-mail"></a><a id="perf-v-mitigation"></a><a id="perf-v-bundle"></a>
<a id="h0-2-hac-csv-1"></a><a id="h0-2"></a><a id="hac-csv-1"></a><a id="h0-3b-h1-2"></a><a id="h0-3b"></a><a id="h1-2"></a><a id="ae-stg-uat-lane3-hac"></a><a id="h3-9"></a><a id="h3-11"></a><a id="lane-4"></a><a id="p1-sec-secrets-5-89-97"></a><a id="p1"></a><a id="sec-secrets-5"></a><a id="89"></a><a id="97"></a><a id="p2-253-prod-setup"></a><a id="p2"></a><a id="253"></a><a id="prod-setup"></a><a id="p3-250-prod-data-migration"></a><a id="p3"></a><a id="250-prod-data-migration"></a><a id="p5-255-staff-provision"></a><a id="p5"></a><a id="255-staff-provision"></a><a id="p6-258-u1-u12-delivery"></a><a id="p6"></a><a id="258"></a><a id="u1-u12-delivery"></a><a id="p7-256-training"></a><a id="p7"></a><a id="256"></a><a id="training"></a><a id="uat-r2-chart-fit"></a><a id="slack-camera"></a><a id="slack-microchip"></a><a id="slack-owner-height"></a><a id="task-444-addendum-codegen"></a><a id="note2-sweep-coverage"></a>
<a id="ver-todo-v-stg-data"></a><a id="ver-todo-v-release"></a><a id="ver-auth-v-d1-preflight"></a><a id="ver-auth-v-d1-apply"></a><a id="ver-auth-v-d1-mail"></a><a id="ops-ae-stg-uat-lane3-hac"></a><a id="ops-h3-9"></a><a id="ops-h3-11"></a><a id="ops-lane-4"></a><a id="p5--staff-provision"></a>

### 共通の前提・確定事項

要件出所は2026-09-19の本会話の依頼者回答（[確定票](#decision-inputs)）。今回は文書更新のみで、実装・検証・移行は未実施。個人名を伴う要件責任者・受入者の参照は実行担当が製品仕様変更前に記録する。これを、範囲の再質問やローカル調査の停止理由にはしない。共有STG/PROD操作、外部送信、実データ移行の承認は別に必要。

<a id="decision-inputs"></a>

#### 医院・PO 入力の確定票

回答先は同じ ID。担当者はこの表を依頼文の下書きに使い、既知の回答を再質問しない。個人名や原票を公開台帳へ転記せず、必要なら権限制限された要件記録の参照を残す。

| ID | 確定済み（2026-09-19依頼者回答） | 開発/QAが調べること・後段の入力 |
|---|---|---|
| UAT-R2-MASTER-PATH | 「可能性のあるページをすべて検証」。発生ページの特定を開始条件から外す | 金額を持つ登録/編集フォームと下流経路をコードから全列挙。合成額で保存/再読込/利用を追跡し、再現した問題を修正 |
| UAT-R2-EXCLUSIVE-LOCK | 「両方」＝同じカルテの上書きと同じ会計の二重確定を防ぐ | 既存version/状態検証/行ロック/冪等性を対象操作ごとに照合。別端末が異なる再送キーを使うケースも含める |
| UAT-R2-CHART-FIT | Windows 8、Chrome、15.6インチ、1366（横）×625（縦）でも見切れないこと。既報の125%は維持 | 1366×625を再現基準にし全9タブ/ダイアログを検証。物理寸法とCSS viewportを同一視せず、実機値・Chrome版は検証担当が採取 |
| UAT-Q2-TREATMENTS-IMPORT | 「全部」＝処置名/料金のマスタと患者ごとの全期間の処置履歴。期間や処置種類で意図的に絞らない | 対象医院ごとに提供原本の全期間を対応づける。関連注射・薬剤/処方行も棚卸しして抜けを報告。無関係な業務データの全DB移行や過去分の再請求を許可した意味にはしない |

業務目的は、料金の再入力/転記の解消、見切れによる操作不能の解消、競合による診療記録消失/二重請求の防止、過去診療の参照欠落の解消。回答済みの4点は再質問しない。実機の採取、旧列の調査、テストで決められる事項は担当者の作業にする。全面ロック、全マスタの会計選択への混在、日付捏造を採用しない。

<a id="data-investigation-contracts"></a>

#### 集計設計の確定事項（9月18日の source 照合）

**ワクチン種:** 現行 [接種フォーム](frontend/src/features/vaccinations/hooks/use-vaccination-form.ts) は有効マスタだけを選び、[取得 hook](frontend/src/hooks/use-treatment-master.ts) は species を送らない。[repository](backend/internal/medicalrecord/vaccine_repository.go) の species 指定は完全一致で、`cat` に `both` が自動で含まれる契約ではない。単に `species=cat` を追加する修正計画では既存の受入条件を満たさない。old_db `030_stage.sql` の種 NULL と、アプリ選択条件の問題を分ける。

- 集計単位は対象医院内の接種。pet/vaccine の参照先も同一医院かを検証し、不一致・削除済み参照を通常集計へ混ぜない。出力は医院の匿名ラベル、ペット種、マスタ種（欠損を独立）、承認されたマスタ識別子、件数、参照不一致件数。個体名・履歴本文・実個体 ID は共有しない。
- A=マスタ種欠損、B=旧記録との照合で参照違いを確認、C=旧記録と一致、未照合=UNKNOWN。A と C は同時に成立し得るので「種品質」と「履歴照合結果」を別列にする。名称だけから B/C を決めない。
- 合成期待表は猫/犬 × `cat`/`dog`/`both`/未設定、無効マスタ、他医院参照、既存履歴を含める。調査後の選択 UI は猫に `cat`/`both`、犬に `dog`/`both` を残す。未設定・犬猫以外の扱いは医院判断を記録し、既存履歴を新規候補の条件で消さない。

**未納:** [現行計算](backend/internal/billing/unpaid_amount.go) に合わせ、payment なしの waiting は `billings.total_amount`、payment ありの waiting/completed は `max(0, total_amount - insurance_amount - discount_amount - billing_amount)`（payment の各額）で算出する。他 status は0。completed の訂正残も対象であり、waiting 件数だけで未納全体を説明しない。

- 対象医院・支払予定日の範囲を [未納一覧仕様](docs/spec/screens/30-unpaid-list.md) と合わせる。請求1件1行を母集団にし、同一医院の未削除 payment との対応件数を検査する。複数対応・医院不一致は別の異常件数に出して停止し、payment_splits の直接 JOIN で請求額を増幅しない。
- 出力は status、payment 有無、未納額帯（0円 / 1–9,999円 / 10,000円以上）、請求件数、請求額合計、未納額合計、旧記録との照合済み/未照合。金額帯は調査用区分で、回収判断や製品仕様の変更ではない。
- 合成期待表: waiting/paymentなし/請求1,000円→未納1,000円、completed/paymentあり/総額1,000円・保険0・割引0・支払額600円→400円、同支払額1,000円→0円。waiting/paymentありにも同じ差額式を使う。分類前後の件数と未納額を照合し、payment の有無だけで「実未納/突合失敗」と断定しない。

以上はローカル集計設計の入力。実データの取得・更新は [運用 TODO](#uat-data-operations) の条件を満たしてから行う。

<a id="slack-answered"></a>

### 出典とトピックの対応

#### トピック別の処理先

45キーを以下へ対応づける。30件はSlack個別プラン、8キーは既存9課題（EXISTING-OTHER は2件）、5キーは実装済み部分の受入先、2キーは回答済み操作案内。課題本文は上の5エリアへ配置している。本文の保存やローカル検証を、報告されたソフトウェア不具合の解消とは扱わない。

| Topic key | 処理区分・担当課題 |
|---|---|
| ACCESS | 運用準備 → [SLACK-ACCESS](#slack-access) |
| UAT-SCHEDULE | 受入準備 → [SLACK-UAT-SCHEDULE](#slack-uat-schedule) |
| HAC-IMPORT | 既存運用へ接続 → [SLACK-HAC-IMPORT](#slack-hac-import) |
| LAB | 調査・実機受入待ち → [SLACK-LAB](#slack-lab) |
| MANUAL-URINE | 合成回帰追加済み・M4/臨床受入は残る → [SLACK-MANUAL-URINE](#slack-manual-urine) |
| OCR | 納品後 DEFERRED → [SLACK-OCR](#slack-ocr) |
| EXAM-HISTORY | 移行元の区別・受入 → [SLACK-EXAM-HISTORY](#slack-exam-history) |
| SMAREGI | 納品後 DEFERRED → [SLACK-SMAREGI](#slack-smaregi) |
| SHIFT | 回答済み → [操作案内の保持](#slack-answered) |
| RESERVATION-REFERENCE | 対応報告あり・再確認待ち → [SLACK-RESERVATION-REFERENCE](#slack-reservation-reference) |
| STAFF-SELECT | 候補状態表示を部分対応・実機再確認待ち → [SLACK-STAFF-SELECT](#slack-staff-select) |
| CROSS-CLINIC | 権限境界の設計・PO → [SLACK-CROSS-CLINIC](#slack-cross-clinic) |
| OWNER-HEIGHT | 検索modalコード対応済み・実機受入待ち → [SLACK-OWNER-HEIGHT](#slack-owner-height) |
| SEARCH-AND | コード対応・実機受入完了（Plane Done・EMR-122） → [UAT-Q1-SEARCH-AND](docs/work/plane-md-migration-20260923-receipt.md) |
| HISTORY | 導線受入完了（Plane Done・EMR-124） → [UAT-Q2-HISTORY-NAV](docs/work/plane-md-migration-20260923-receipt.md) |
| SPECIES | 既存調査 → [UAT-Q2-VACCINE-SPECIES](#uat-q2-vaccine-species) |
| GENDER | old_db 統合済み・bundle/実データ受入待ち → [UAT-Q3-GENDER-MAP](#uat-q3-gender-map) |
| UNPAID | 原因未確定・既存調査 → [UAT-Q4-UNPAID-TRIAGE](#uat-q4-unpaid-triage) |
| INSURANCE | 受入完了（Plane Done・EMR-123） → [UAT-Q4-INSURANCE-RATES](docs/work/plane-md-migration-20260923-receipt.md) |
| MASTER | 全経路受入完了（Plane Done・EMR-84） → [UAT-R2-MASTER-PATH](docs/work/plane-md-migration-20260923-receipt.md) |
| CONCURRENCY | 受入完了（Plane Done・EMR-85） → [UAT-R2-EXCLUSIVE-LOCK](docs/work/plane-md-migration-20260923-receipt.md) |
| LATENCY | 比較条件設計済み・実環境/実測待ち → [SLACK-LATENCY](#slack-latency) |
| ENTER | 実機IME受入完了（Plane Done・EMR-120） → [UAT-R2-TREATMENT-COMMIT](docs/work/plane-md-migration-20260923-receipt.md) |
| MASTER-HEIGHT | 実画面受入完了（Plane Done・EMR-121） → [UAT-R2-MASTER-LIST-HEIGHT](docs/work/plane-md-migration-20260923-receipt.md) |
| CHART-FIT | scroll/高さコード対応済み・全9タブ/実機受入待ち → [UAT-R2-CHART-FIT](#uat-r2-chart-fit) |
| RESERVATION-EDIT | 回答・謝辞あり → [操作案内の保持](#slack-answered) |
| COMPLAINT | 修正・受入完了（Plane Done・EMR-87） → [SLACK-COMPLAINT](docs/work/plane-md-migration-20260923-receipt.md) |
| BACKGROUND | 対象欄の特定・PO → [SLACK-BACKGROUND](docs/work/plane-md-migration-20260923-receipt.md)（Plane で Duplicate 終了・統合先は Plane 参照） |
| VITALS | ヘッダー表示・臨床受入完了（Plane Done・EMR-190） → [SLACK-VITALS](docs/work/plane-md-migration-20260923-receipt.md) |
| MICROCHIP | ヘッダー表示コード対応済み・受入待ち → [SLACK-MICROCHIP](#slack-microchip) |
| DANGER | 既存高危険表示の保持・PO → [SLACK-DANGER](docs/work/plane-md-migration-20260923-receipt.md)（Plane で Duplicate 終了・統合先は Plane 参照） |
| DETAILS | 導線比較済み・PO裁定待ち → [SLACK-DETAILS](docs/work/plane-md-migration-20260923-receipt.md)（Plane で Duplicate 終了・統合先は Plane 参照） |
| STORY | 記録目的/保存先のPO → [SLACK-STORY](docs/work/plane-md-migration-20260923-receipt.md)（Plane で Duplicate 終了・統合先は Plane 参照） |
| VACCINE-PRINT | 専用証明書の仕様確認 → [SLACK-VACCINE-PRINT](#slack-vaccine-print) |
| VACCINE-MULTI | 単件/順次POST回帰追加済み・元症状/実機待ち → [SLACK-VACCINE-MULTI](#slack-vaccine-multi) |
| DECEASED | 死亡後連絡記録のPO → [SLACK-DECEASED](#slack-deceased) |
| PLAN-MANUAL | 検索優先回帰追加済み・案内/導線採否のPO待ち → [SLACK-PLAN-MANUAL](docs/work/plane-md-migration-20260923-receipt.md)（Plane で Duplicate 終了・統合先は Plane 参照） |
| COPY | 安全条件整理済み・PO裁定待ち → [SLACK-COPY](docs/work/plane-md-migration-20260923-receipt.md)（Plane で Duplicate 終了・統合先は Plane 参照） |
| CAMERA | 撮影入口コード対応済み・実機受入待ち → [SLACK-CAMERA](#slack-camera) |
| BILLING-UAT | 最優先の実機受入 → [SLACK-BILLING-UAT](#slack-billing-uat) |
| CLINICAL-UAT | 各院の診療/看護受入 → [SLACK-CLINICAL-UAT](#slack-clinical-uat) |
| INTAKE | repo内分類済み・外部照合条件待ち → [SLACK-INTAKE](#slack-intake) |
| BACKLOG61 | 原票・PDF内容待ち → [SLACK-BACKLOG61](#slack-backlog61) |
| TREATMENTS-IMPORT | 列写像調査済み・全期間履歴契約案へ → [UAT-Q2-TREATMENTS-IMPORT](#uat-q2-treatments-import) |
| EXISTING-OTHER | 保存 → [死亡日限定訂正](#po-pet-deceased-data-backfill)、[追記型生成](#task-444-addendum-codegen) |

<a id="slack-source-map"></a>

#### 全50親投稿・返信の出典対応表

行範囲は返信を含む。応答・謝辞・日程調整は各行の履歴へ含め、返信中の別要件をトピックへ分割した。親として再掲された2件は同一イベントであり、重複の実装課題を作らない。50親ブロックの内訳は通常対応44、重複再掲2、対象外4（参加通知3、表紙PDFの提出・謝辞1）。対象外は製品課題を追加しない判断であり、助成金受理の確認済みを意味しない。

| 出典行 | 親 timestamp | 処理先・理由 |
|---|---|---|
| 32–41 | `1787984405.983319` | ACCESS / UAT-SCHEDULE |
| 42–57 | `1787988264.830289` | UAT-SCHEDULE。候補日は履歴 |
| 58–76 | `1787992807.018379` | ACCESS / UAT-SCHEDULE |
| 77–100 | `1788261491.711849` | ACCESS / HAC-IMPORT。伏字の秘密は入力に使わない |
| 101–126 | `1788406963.496419` | SMAREGI / BILLING-UAT。購入内容を顧客と確認する要件を保持 |
| 127–159 | `1788422744.682529` | ACCESS / UAT-SCHEDULE / HAC-IMPORT。旧DB共通という返信は新システムの医院分離解除の根拠にしない |
| 160–173 | `1788427309.960679` | ACCESS / UAT-SCHEDULE |
| 174–178 | `1788529344.397139` | 対象外: 参加通知 |
| 179–183 | `1788662604.083949` | 対象外: 参加通知 |
| 184–211 | `1788671384.944389` | 対象外: 表紙付き提案PDFの送付・謝辞あり。助成金受理は未確認 |
| 212–294 | `1788686430.832129` | LAB / MANUAL-URINE。猫・八王子の写真、敷島コアグの追加返信も含む |
| 295–301 | `1788764745.291709` | HAC-IMPORT / UAT-SCHEDULE |
| 302–402 | `1788832046.595209` | ACCESS / HAC-IMPORT。9月11日のCSV送付は取込完了の証拠ではない |
| 403–423 | `1788853449.630089` | HAC-IMPORT。過去の「2日後」は新期限へ転用しない |
| 424–460 | `1788866843.926759` | OCR。納品後へ延期、当面の画像/PDF閲覧を保持 |
| 461–482 | `1788931012.789169` | 重複再掲: 302–402の返信。ACCESS / HAC-IMPORT に統合 |
| 483–505 | `1788934802.744769` | ACCESS / SMAREGI / UAT-SCHEDULE。7月転送原文は未収録 |
| 506–529 | `1788940302.976299` | EXAM-HISTORY。DrOne単体の移行除外は回答済み、旧カルテ検査履歴の受入は残る |
| 530–559 | `1789109051.075399` | SMAREGI。取り置き・顧客・端数・trial、本件は後日の延期に従う |
| 560–599 | `1789176984.729069` | SHIFT。操作回答・謝辞あり、予約成功の実証とは分離 |
| 600–642 | `1789189178.939309` | SMAREGI / UAT-SCHEDULE。会議日程調整は回答済み、新しい会議タスクなし |
| 643–697 | `1789191550.389859` | RESERVATION-REFERENCE / STAFF-SELECT / CROSS-CLINIC。Chromeでも選択不能の返信を保持 |
| 698–703 | `1789259845.886529` | UAT-SCHEDULE。テスト日の通知 |
| 704–724 | `1789276317.113699` | UAT-SCHEDULE / CLINICAL-UAT。敷島の交代勤務者の網羅 |
| 725–735 | `1789277858.989679` | 重複再掲: 643–697の返信。STAFF-SELECT / CROSS-CLINIC に統合 |
| 736–763 | `1789351591.207249` | OWNER-HEIGHT。最大化済み、Win8/Chrome・1366×625・15.6インチ |
| 764–783 | `1789388624.382679` | SEARCH-AND。実装報告と謝辞、UAT は既存検証先 |
| 784–794 | `1789389491.915289` | HISTORY。導線/移行済みデータの受入 |
| 795–805 | `1789389696.355619` | SPECIES。猫履歴に犬用、原因未確定 |
| 806–815 | `1789390572.658379` | GENDER。旧3/4コード修正後の実データ受入 |
| 816–831 | `1789391240.242759` | UNPAID / INSURANCE。移行不備との当時の推測は原因確定にしない |
| 832–847 | `1789437234.502929` | CROSS-CLINIC。代表アカウントという追加条件 |
| 848–882 | `1789453043.269699` | MASTER / CONCURRENCY。9月16日の価格・再診察・追加明細・照会混入も別ケース化 |
| 883–896 | `1789459005.162369` | LATENCY / ENTER / MASTER-HEIGHT。3要件を分離 |
| 897–906 | `1789459375.695359` | CHART-FIT |
| 907–923 | `1789463634.587649` | RESERVATION-EDIT。案内・謝辞あり |
| 924–928 | `1789538822.397429` | 対象外: 参加通知 |
| 929–937 | `1789539164.115569` | COMPLAINT / BACKGROUND / VITALS。3要件を分離 |
| 938–943 | `1789539363.704399` | MICROCHIP |
| 944–948 | `1789539576.256969` | DANGER |
| 949–955 | `1789539700.846829` | DETAILS / STORY。詳細導線と記録欄を分離 |
| 956–960 | `1789539775.666449` | VACCINE-PRINT |
| 961–967 | `1789540156.256659` | MASTER。マスタなし/0円/手入力案内 |
| 968–975 | `1789540169.467719` | VACCINE-MULTI。登録失敗と同日2–3件入力を分離 |
| 976–981 | `1789540531.163329` | LAB。受信可否と対応時期は実機条件確認後 |
| 982–986 | `1789540849.748249` | DECEASED。死亡日訂正とは異なる連絡記録 |
| 987–994 | `1789541345.488879` | PLAN-MANUAL / COPY。画面の違い・手入力と前回複写を分離 |
| 995–999 | `1789541513.692419` | CAMERA |
| 1000–1063 | `1789550614.370909` | BILLING-UAT / CLINICAL-UAT / INTAKE / SMAREGI。後者延期を優先、黒塗り重複行の削除回答は外部削除実施証拠ではない |
| 1064–1070 | `1789633031.590089` | BACKLOG61。PDF内容は UNKNOWN |

### 更新規則

ID ごとに状態・次の一手・完了条件・根拠を持つ。主な開始条件が変わったら本文を該当エリアへ移し、件数・トピック対応を更新する。同じIDの本文を複数エリアへ複製しない。完了項目は未対応エリアから外し、エリア5とSlack対応表に回答済み/完了/重複/対象外の根拠と移管先を残す。未完了の検証・運用は専用 TODO へ参照を残す。既存 claim は [AGENTS.md](AGENTS.md) に従って確認し、別担当の作業を取り込まない。

---

<a id="verification-ledger"></a>

## 検証・受入の履歴（現在のタスク状態はPlane）

> Current task state is in Plane. Local task details were migrated 2026-09-23; see [crosswalk](docs/work/plane-md-migration-20260923-receipt.md).

最終タスク移行: 2026-09-23。未完了の検証・受入タスクと状態はPlaneを正本とする。実行記録・検証原則・証拠はローカルで保持する。ローカルケース票の一部は [linmig-campaign-20260919](docs/work/linmig-campaign-20260919/) と [remaining-campaign-20260920](docs/work/remaining-campaign-20260920/) に作成済み。今回 runtime は再実行していない。過去の結果は当時の revision に限定し、追加実装の受入は現在の receipt 未照合として UNKNOWN を維持する。

着手プラン確認: 2026-09-22。前回の [追加実装のキュー](#code-followup-20260921) を維持し、今回8件の追加対応から [残検証・受入](#ready8-followup-20260922) を更新した。既存ケース票を再利用し、追加済みunit/mockと、ローカル未カバー・実DB・実機・医院/POの未確認を分ける。調査票のGREENは当時の記録で、今回の再実行結果ではない。

受入タスクの現行状態・未実施条件はPlaneの移行済み項目を参照する。過去runの証拠と判定原則はこのファイルに残し、他タスクの起動中コンテナや共有DBを検証先に流用しない。

### 判定原則

- source・静的チェック・unit・CI・配備成功と、ブラウザ UAT・production・go-live を分ける。
- 過去の未実施記録だけから、現在も未実施と断定しない。新しい証拠を照合していないものは UNKNOWN、既知の前提不足は BLOCKED とする。
- 実行前に対象 revision、環境、操作者、承認、fixture、証拠保存先を固定する。秘密・接続文字列・患者情報は台帳や共有ログに含めない。
- データ書込み・migration・配備・外部送信は承認された単位のみ。今回の文書更新ではテスト、DB 操作、ブラウザ UAT を実行していない。

<a id="readiness-preparation"></a>

### 今着手する検証準備

QA/開発が作るケース票の共通列は `ID / case / revision / 環境・fixture参照 / 操作 / 期待値 / 実際値 / 証拠参照 / 後処理 / 判定`。準備時の実際値は「未実行」とし、同じ ID の既存 run に十分な証拠があれば再実行せず対応づける。準備完了は「未実行の行と不足入力が特定できた」状態。ブラウザが使えることだけで共有環境のデータ操作まで許可されたとは扱わない。

`PERF-V-CLIENT-TRACE` / `PERF-V-CF-EVENTS` / `PERF-V-DECIDE-OBSERVATION` / `PERF-V-MITIGATION` / `PERF-V-BUNDLE` / `PERF-V-STG-ACCEPTANCE` は [性能の6単位](#性能タスクの着手順と成果物) がケース票の正本。因果区間は E4–E8 で局在済み。MITIGATION/BUNDLE（`EMR-142`/`EMR-143`）は Plane Ready。依頼・承認が必要な行も、ケース票作成はローカルで先行できる。

<a id="uat-followup"></a>

### 直近 UAT の残作業

各項目の受入状態・現在状態は Plane を正本とする（[移行記録](docs/work/plane-md-migration-20260923-receipt.md) 参照）。

<a id="code-followup-20260921"></a>

#### 追加実装に伴う受入（2026-09-21照合）

以下は同じ Issue ID の検証範囲であり、別の開発チケットではない。コード対応済み5件と、部分対応の STAFF / EXCLUSIVE を分ける。各行の実行前に対象 build・端末/環境・合成 fixture・操作者・操作範囲/承認・後処理・証拠保存先を固定する。新たに確認したのはコードとテスト定義の存在までで、テスト実行・配備・実機成功は今回確認していない。

各項目の受入状態・現在状態は Plane を正本とする（[移行記録](docs/work/plane-md-migration-20260923-receipt.md) 参照）。

再現した不一致は同じ Issue ID に戻す。元報告の画面・端末や臨床上の期待値が一致しない場合、既存修正の成功から補外せず、未確認ケースと必要な判断を残す。

<a id="ready8-followup-20260922"></a>

#### 追加対応8件の残検証・受入（2026-09-22照合）

基準は `cd2feaa14`。以下は既存IDの残条件で、新規課題や追加済み回帰の再作成ではない。実行前の対象build・専用環境/fixture・操作者・操作範囲/承認・後処理・証拠保存先は上の共通条件に従う。unit/mockは実DB永続化・実端末操作・医院の臨床期待の代替にしない。

各項目の受入状態・現在状態は Plane を正本とする（[移行記録](docs/work/plane-md-migration-20260923-receipt.md) 参照）。

Q1 / Q4保険 / Q2履歴の実装は再開しない。根拠は [医院フィードバック](docs/work/stg-uat-clinic-feedback-q1-q4.md) と、9月15日に読取確認した [PR #411](https://github.com/MinoruSoga/AnimalEkarte/pull/411)（merged）、[Backend Deploy](https://github.com/MinoruSoga/AnimalEkarte/actions/runs/34923018516) / [Frontend Deploy](https://github.com/MinoruSoga/AnimalEkarte/actions/runs/34923018544)（ともに success、`d337f016`）。この配備記録はブラウザ確認や本番反映の代替ではない。

`NOTE2-SWEEP-COVERAGE` の受入条件はPlane `EMR-200` へ移行済み。route×operation カバレッジ表と API/契約レベルの証拠は [docs/work/emr-200-note2-coverage/NOTE2-SWEEP-COVERAGE.md](docs/work/emr-200-note2-coverage/NOTE2-SWEEP-COVERAGE.md) を参照（2026-09-23、revision `923bb99ba`。ブラウザ UAT は未実施 UNKNOWN/BLOCKED）。

### STG データレーン

入力受領と操作は [運用 TODO](#stg-data-lanes)。受入では source manifest・対象医院・件数・金額・監査・画面を照合する。Lane 4 は両院の Lane 3 verify、H3-11、所定の運用日数の証拠が揃うまで完了にしない。9月13日の一部 CRUD 成功をこの受入全体に拡張しない。

<a id="linear-reconciliation"></a>

### Linear 照合の過去記録（現在の状態はPlane）

9月15日の読取結果（当時）: [BRT-4](https://linear.app/baritechllc/issue/BRT-4) は Backlog、[BRT-45](https://linear.app/baritechllc/issue/BRT-45) / [BRT-68](https://linear.app/baritechllc/issue/BRT-68) は Needs Human。これは当時の読取記録。9月18日の Linear MCP 再照会も未接続（`USER_NOT_LOGGED_IN`）で失敗し、現在の状態・対応先は UNKNOWN。完了済みチケットは残件表から除く。

Linear 系の残件は全て終了を確認（2026-09-27 照合・Linear MCP 未接続のまま Plane で終端）: `TODO-V-LINEAR / META-LINEAR-APPLY` → `EMR-135`、`PERF-V-LINEAR` → `EMR-136`、`AUTH-V-LINEAR-READ` → `EMR-137`、`AUTH-V-LINEAR-WRITE` → `EMR-138` はいずれも Cancelled。残件表は削除した。

新規 Issue を作らない方針は維持するが、過去の free issue limit を read-only 照会や既存 Issue 更新の技術的ブロッカーにしない。類似語だけでチケットを割り当てず、不明なら UNKNOWN とする。今回、外部投稿・状態変更は未実施。

### PERF-STG-LOGIN

技術記録は [todo-performance.md](#performance-ledger)。Worker 観測は現在 tracked code に存在し、未導入 WIP の扱いを終了した。`PERF-V-IMPLEMENT-OBSERVATION` の実 proxy 4 tests・worker typecheck（`index.test.ts` include済み）は `72807128` の [既存検証](.planning/agent-fast-campaign/four-candidate-integration-20260916/evidence/rev7-reverify-72807128-codex/controller/RECONCILIATION.md) で完了し、開いたキューから外した。残るのは遅延の因果測定、常時観測の必要性・出力範囲の再判定、STG 受入である。（2026-09-23 E5 追記: 因果は `containerFetch` 区間まで局在、常時観測は KEEP 判定、STG 受入証拠は4ケース取得済。各項目の現在状態は Plane を参照。）

2026-09-23 E5 追記（証拠は `reports/perf-e5-residual-20260923/`）: CLIENT-TRACE は `/login` cold/warm 2遷移を記録し OPTIONS 0・FCP 692/88ms を取得（[client-trace](reports/perf-e5-residual-20260923/client-trace/README.md)）。CF-EVENTS は cf-ray 相関で `container_fetch` 866–3848ms 支配・edge+worker 約60–115ms・稼働 instance `maa01`・`scheduling_policy` deployed=`default` vs config=`regional` の乖離を取得（[cf-events](reports/perf-e5-residual-20260923/cf-events/README.md)）。DECIDE-OBSERVATION は KEEP と判定（[obs-decision](reports/perf-e5-residual-20260923/obs-decision/README.md)）。STG-ACCEPTANCE は匿名/既存session/復旧/ログイン後の4ケースを記録し login POST 3882ms・`/v1/me` 1475ms ゲート・OPTIONS 0（[stg-acceptance](reports/perf-e5-residual-20260923/stg-acceptance/README.md)）。測定値は現行 STG 配信版のもので、E5 のコード変更は未配備。PERF-V-LINEAR は対応先未確定のまま。

2026-09-23 E6 追記（post-deploy、証拠は `reports/perf-e5-postdeploy-verify-20260923/`、配信版 worker `15a85636`・コンテナ v70→v71）: DEPLOY-VERIFY は配信版が perf コミット `453be4ecc` を後置することと edge OPTIONS 204（Go ヘッダ無し）を確認し SERVES-NEW-BUILD（[deploy-verify](reports/perf-e5-postdeploy-verify-20260923/deploy-verify/README.md)）。OPTIONS-EDGE は 6 probe で edge 契約を検証（allowlist echo・非 allowlist は ACAO 無し・`_internal` 404・OPTIONS 中央値 ~60ms、[options-edge](reports/perf-e5-postdeploy-verify-20260923/options-edge/README.md)）。WARM-MEASURE は warm 中央値が E5 と ±0.01s 内で同一（[warm-measure](reports/perf-e5-postdeploy-verify-20260923/warm-measure/README.md)）。LOGIN-MEASURE は中央値 1.471s vs E5 ~3.33s（−56%）で bcrypt-skip を方向的に確認（warmth 混在のため単独寄与は UNKNOWN、[login-measure](reports/perf-e5-postdeploy-verify-20260923/login-measure/README.md)）。PLACEMENT は `maa01` 継続・`scheduling_policy` 乖離は未解消・再抽選を推奨（[placement](reports/perf-e5-postdeploy-verify-20260923/placement/README.md)）。CF-EVENTS は `container_fetch` 1001–1988ms・edge+worker ~57–63ms・観測窓内で `sin14` へ再作成（[cf-events](reports/perf-e5-postdeploy-verify-20260923/cf-events/README.md)）。BROWSER-PAGES は実ブラウザで N+1 解消を確認し `accountings?owner_id=` 1.9–2.7s を新規候補として記録（[browser-pages](reports/perf-e5-postdeploy-verify-20260923/browser-pages/README.md)）。STG-ACCEPTANCE は login POST 1451ms（E5 比 −62.6%）・`/v1/me` ゲート 1100.7ms・全ケース OPTIONS 0（[stg-acceptance](reports/perf-e5-postdeploy-verify-20260923/stg-acceptance/README.md)）。未解決の正直な記録: AXIOS-RETRY は配備済みだが観測窓に 503 が無くフィールド未検証（unverified-in-field、失敗ではない）、INSTANCE-TYPE/MITIGATION/BUNDLE はトリガー付き DEFERRED のまま、SLACK-LATENCY はユーザーレーン（`EMR-104`）、PERF-V-LINEAR は Linear MCP 未接続（`USER_NOT_LOGGED_IN`）で BLOCKED のまま。

2026-09-24 E7 追記（perf-lane-all-20260924、証拠は `reports/perf-lane-all-20260924/`、配信版 worker `a52c2795`・コンテナ v72）: PERF-ACCT-VERIFY（EMR-201）は配信版が `e818195ec` を含み migration 007 適用済みを確認し、`accountings?owner_id=` warm 中央値 1.091s（n=5）vs E6 baseline 1.9–2.7s で **improved**（[accountings-verify](reports/perf-lane-all-20260924/accountings-verify/README.md)）。PERF-PLACEMENT-CHECK（EMR-202）は v72 で `scheduling_policy` deployed=`default` vs config=`regional` の乖離継続・singleton `maa01`・migrate-runner `bom09` を再観測し、文書記載 8/8 MATCH・再抽選推奨 YES（[placement](reports/perf-lane-all-20260924/placement/README.md)）。PERF-COST-MEMO（EMR-204）は instance_type 引上げのコスト判断メモを作成し open decision（std-1 ≈ +$2.8–20.4/mo・std-2 ≈ +$4.7–34.3/mo の ESTIMATE / reject / defer、[instance-type-cost](reports/perf-lane-all-20260924/instance-type-cost/README.md)）。未解決の正直な記録: EMR-203（AXIOS-RETRY）は 503 無しで unverified-in-field のまま、EMR-142/EMR-143 はトリガー付き DEFERRED のまま、EMR-104 はユーザーレーン、EMR-199 は終端（prior campaign）、EMR-136 は cancelled（Linear MCP 未接続）。

- `PERF-V-CLIENT-TRACE`: 承認済み対象・時間枠・停止担当・証拠保存先を固定する。相対時刻、method、status、protocol、initiator、OPTIONS/GET 対応、FCP を保存し、Cookie・Authorization・本文・個人情報は含めない。HAR 等は保存前に機密除去。単発値や未使用時間だけで p95/p99・cold start・改善完了と判定しない。
- `PERF-V-CF-EVENTS`: provider 時刻と browser 時刻を対応づけ、Container 起動証拠がない場合は Worker 所要時間だけで起動待ちと断定しない。観測で設定・配備を変更しない。
- 将来の変更時だけ、既存依存を利用し対象候補を mount した隔離 Docker の scoped Vitest と worker typecheck を別々に確認する。依存インストールや今回の再実行は不要。
- `PERF-V-BUNDLE`: cache 条件を分ける。過去の HTML load 約0.31秒だけで約23秒の待ちを bundle 起因としない。公開 entrypoint を壊す deep import や一括 chunk 再編を先行させない。

[STG パフォーマンス測定チェックシート](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md) に従い、証拠は既存の非公開 run 保存先へ置く。STG traffic・Cloudflare 読取・設定変更・配備・Linear 更新は各対象の承認範囲で行う。

#### 性能タスクの着手順と成果物

各 run の保存先は `reports/uat-YYYY-MM-DD/performance/<RUN-ID>/`。チェックシートを複製し、対象 revision・承認範囲・測定条件を先に記入する。Git ignore と機密除去を確認し、測定していない p95/p99 や未合意の SLO を達成済みにしない。

各単位の状態・次の一手は Plane（`EMR-142` / `EMR-143` ほか）と上記 E5–E8 の記録を参照。

<a id="認証認可の外部境界"></a>

### 認証・認可の外部境界

認証 D1 の対象環境での付与・メール確認をここで一元管理する。契約は [認証設計](docs/architecture/auth.md)、実行手順は [初回管理者手順](docs/ops/deploy/FIRST_SYSTEM_ADMIN.md)、合成検証の参照は [D1 SQL fixture](backend/internal/auth/testdata/first_system_admin.sql)。ローカル実装・合成 fixture の完了を、対象環境への付与やメール経路の成功に読み替えない。

過去のLinear照合記録は [こちら](#linear-reconciliation)。現在のタスク状態は各Plane項目を参照。共有 `ekarte_db` / `old-db-postgres` / STG / PROD をテスト DB に使わず、本番付与・メール・migration を自動実行しない。receipt は対象環境、日時、担当、承認、結果、復旧可否のみを既存 runbook へ保存する。外部受入が残れば完了にしない。

---

<a id="operations-ledger"></a>

## 運用・外部実行の受入記録（現在のタスク状態はPlane）

> Current task state is in Plane. Local task details were migrated 2026-09-23; see [crosswalk](docs/work/plane-md-migration-20260923-receipt.md).

最終タスク移行: 2026-09-23。未完了の環境・データ・本番・納品作業と現在状態はPlaneを正本とする。下記の実行条件・過去receiptは運用根拠として保持する。

remote・CI・配備・Linear・STG/PROD の DB/秘密/投入 receipt は今回再照会していない。以前の「未構築」「未受領」を現在の事実として断定せず、再実行前に実施有無を確認する。ローカル準備票の作成は実行完了ではない。

運用全体の着手プランは 2026-09-21 に再点検し、統合済みの LINMIG / remaining 準備票と現行コードを照合した。既存票を作り直す段階ではなく、不足入力・対象環境の適用状態・実行証拠を揃える段階。各表の ID から下の個別手順を参照する。承認前にも、既存資料の照合・不足入力表・実行案の作成は進められる。外部の実操作を、この計画の記載だけで開始しない。

<a id="ops-readiness-preparation"></a>

### 今着手する運用準備

次の担当は運用/producer/USER。個人の割当は未確定であり、実行票で確定する。実行票は `ID / 対象環境・医院 / revision・manifest等の入力識別 / 読取・変更する範囲 / 前段証拠 / 操作者・承認者 / 実行枠 / 中止条件 / 復旧 / 証拠保存先` を持つ。秘密・実 roster・患者情報は repo 外の承認済み保管先に置き、台帳には非機密参照だけを残す。既存 receipt が見つかれば同一対象・入力との一致を先に検証し、再実行の要否を決める。

各項目の受入状態・現在状態は Plane を正本とする（[移行記録](docs/work/plane-md-migration-20260923-receipt.md) 参照）。

準備は不足票・実行案を同じ ID に残した時点で一区切り。2026-09-21 時点で上表からリンクした準備票は main に統合済み。ただし、旧IDの担当分界や実行枠などの未確定欄まで埋まったことを意味しない。実行条件を満たさない行は BLOCKED/UNKNOWN を維持する。必要な入力を揃える作業と、実環境を変更する作業を同じ「着手済み」にまとめない。

<a id="uat-data-operations"></a>

### 医院 UAT に伴うデータ操作

実行手順は [ローカル handoff](docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md)、[STG 停止ゲート](docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md#2-pre-deploy-stop-gates)。共有環境への書込み・再取込・DB 作成/破棄・migration は今回実施しない。必要な `make migrate` は、対象環境の適用状態を確認したうえでユーザーが実行する。

<a id="stg-data-lanes"></a>

### STG データレーンの残り

同一の既知破損 BAK の再復元や、既存 STG の無条件上書き load は行わない。9月13日の CRUD 記録と9月15日の配備成功は、これらの全データ受入の代替にならない。

### 本番・納品の残り

Q1検索・Q4保険・Q2履歴、[追加実装5件と会計等の部分対応](#code-followup-20260921)、今回 `cd2feaa14` の主訴null hydrate修正は、現行productionへの反映証拠を今回未照合（UNKNOWN）。本番releaseの対象revisionと既存receiptを照合し、反映済みなら重複配備しない。ブラウザ受入と上記前提を先に確認し、STG 配備だけで本番反映済みにしない。P4 / P8 / E1 / E2、go-live は [検証 TODO](#verification-ledger) を正本とする。

### データ操作の着手プラン

<a id="ops-uat-q3-gender-map"></a>

#### UAT-Q3-GENDER-MAP

1. [性別修正](#uat-q3-gender-map) は old_db `5fbc3b2` → merge `a2cea37` でmain統合済み（9月19日読取）。producer担当は修正を含むrevisionと現行bundle/検証receiptを照合し、未生成なら正規経路で再生成する。38 tests / SQLite CASEの既存成功をPostgreSQL・export・STGの証拠にしない。再統合タスクは起こさない。
2. 運用担当が対象医院、旧コードを追跡できる根拠、訂正対象件数、除外条件、backup・監査・復旧方法を事前照合する。producer の修正と既存 STG の修正は別々に判定する。
3. 承認された限定訂正後に雄/雌の件数・表示を照合する。手術日は変更対象と混ぜず、推測日付を入れない。成果物は producer revision、承認参照、前後集計、画面確認の非機密 receipt。

<a id="ops-uat-q2-vaccine-species"></a>

#### UAT-Q2-VACCINE-SPECIES

1. [調査案](#uat-q2-vaccine-species) を医院・時刻窓・出力列（件数/参照関係だけ）に限定し、読取担当と対象を固定する。
2. 読取承認後に集計を取得し、マスタ種欠損・参照違い・旧記録由来の分類へ渡す。
3. 成果物は機密除去済み分類表と訂正要否。原因未確定や参照不整合が残る間は更新案を適用しない。履歴修正は対象行と復旧方法を確定した別承認とする。

<a id="ops-uat-q4-unpaid-triage"></a>

#### UAT-Q4-UNPAID-TRIAGE

1. [集計案](#uat-q4-unpaid-triage) の対象医院・期間・status・payment の結合条件を確認し、読取担当と保存先を固定する。
2. 承認後に status・支払有無・金額帯を集計し、旧未精算と突合漏れを分ける。重複結合で件数/金額が増えていないか元集計と突合する。
3. 成果物は総件数・金額と原因別集計、追加照合または限定訂正の提案。一括完了・請求削除は実行しない。

<a id="ops-po-pet-deceased-data-backfill"></a>

#### PO-PET-DECEASED-DATA-BACKFILL

1. [元の修復計画](docs/work/todo-campaign-20260918/PO-PET-DECEASED-DATA-BACKFILL.md)（+ Plane `EMR-108`）に沿って不整合の種類、日付根拠、対象医院と件数を整理する。日付根拠がある行だけを訂正対象とし、根拠がない行は補完対象外として残す。
2. 根拠に基づく対象・訂正内容を確定した後、操作者、backup、監査、失敗/通信断時の照合・復旧を含む限定実行案を作る。
3. 承認後の訂正と前後件数・画面・監査の一致で完了とする。死亡 write ガードを再実装せず、監査や既存履歴を削除しない。

---

<a id="performance-ledger"></a>

## Performance evidence and history

> Performance work items are tracked in Plane. The migration crosswalk is [here](docs/work/plane-md-migration-20260923-receipt.md); measurements and historical evidence below remain local.

最終照合: 2026-09-22（同日 curl 実測でコールドスタートを直接観測し **E4** として記録。原因区間を確定し改善候補を整理。SLACK-LATENCY計測票差分は別記録を保持）→ 2026-09-27 再照合で Plane 終端を確認（下記「未完了の性能作業」行を参照）。主調査 ID: **PERF-STG-LOGIN**。対象は STG `/login` 初回表示遅延に始まり、ユーザー報告により STG 全域のページ読み込み遅延へ拡大。責任者・依頼者: 曽我 稔。

未完了の性能作業と現在状態はPlane（`PERF-V-MITIGATION` EMR-142、`PERF-V-BUNDLE` EMR-143、`SLACK-LATENCY` EMR-104、PLACEMENT 再抽選 EMR-202、AXIOS-RETRY フィールド検証 EMR-203）に移行済み。`PERF-E5-STG-DEPLOY-VERIFY` EMR-199・恒久対策 EMR-213・keep-alive 実発火確認 EMR-214 は Done、COST-MEMO EMR-204 は Cancelled。本書は判断に必要な技術記録のみを保持する。

### 現在の判断

- 初回遅延の主因は **Cloudflare Containers の scale-to-zero コールドスタート** と実測で確定（E4）。`sleepAfter = "10m"` でアイドル10分後にコンテナが停止し、全 API リクエストは単一の共有コンテナ（`getContainer(env.API_CONTAINER)`、名前なし=既定インスタンス）の起動完了を待つ。`instance_type: "basic"`（1/4 vCPU / 1GiB）で起動は実測 約6〜16秒。
- E1 の約22.5秒（最終 GET の接続開始前）は Container 起動待ちと整合する。「通信全体を Go / DB 処理時間と断定しない」判断は維持し、温間 TTFB 0.24〜0.25 秒から Go/DB 自体の遅さではないことを確認済み。
- Worker 観測 `container_fetch_timing`（`b42c00ccb` 由来）は継続有効。改善実装後の区間別再測定に使う。
- 温間状態の認証済み API レイテンシは **E5 で実測済み**。`/v1/clinics` 0.47s、`/v1/me` 0.85s（認証キャッシュ適用後・暖機）。E4 の UNKNOWN は解消。`/v1/pets` 検索等の重めクエリは未測定のまま。
- pending UI は既存実装を利用する。bundle 分割（BUNDLE）は転送/parse/execute の寄与が未測定のため DEFERRED のまま。edge OPTIONS と Container 設定は因果証拠（E4）が取れたため実装候補へ昇格したが、採用・範囲・受入は別途確定する。

### 改善プラン（候補・2026-09-22）

対象区間は E4 で確定したコールドスタート待ちと、プリフライト／リトライ／N+1 による増幅。実装単位の確定・受入条件・担当は [todo-issue.md](#issue-ledger) / [todo-verification.md](#perf-stg-login) に移す。ここでは技術的根拠と前提のみ記録する。

| # | 案 | 削減する区間 | 前提・注意 |
|---|----|------------|------------|
| 1 | keep-alive：`sleepAfter` 延長、または営業時間中の定期 `/health` ping | コールドスタート待ち（支配的区間）そのもの | 既存 cron（01:00/06:00/11:00/17:00 UTC）は `SCHEDULER_NAME` の named コンテナを起こすのみで、既定名の API コンテナは温まらない。ping は `getContainer(env.API_CONTAINER)` の既定インスタンスに届く経路（通常 `/health` GET で可）が必要。scale-to-zero のコスト想定（AC-5 検証目的）とのトレードオフは STG 運用判断 |
| 2 | OPTIONS を Worker エッジで応答（Container へ転送しない） | プリフライトのコールド待ち（実測 6 秒）と温間時の往復 1 回分 | CORS allowlist は静的（`CORS_ALLOWED_ORIGIN` vars）で Worker 側に複製可能。`Allow-Headers/Methods/Max-Age` の契約は `internal/middleware/cors.go` と一致させる。GET 本体のコールド待ちは残るため単独では不十分 |
| 3 | axios の 502–504 リトライ範囲・回数の見直し | 起動中の二重待ち（1秒・2秒バックオフ＋再送） | `isServerError` は 502–504 で、Worker が起動失敗時に返す 503 `service_unavailable` を含む。起動待ち中の再送はコンテナ起動を早めない。503 を対象外にするか回数を絞るかは、起動失敗時の UX（速く失敗を返す）とセットで判断 |
| 4 | `ownerLoader` の N+1（飼主 + ペット個別 GET）解消 | 飼主詳細のリクエスト数（ペット N 頭で N+1 本、各 URL が別プリフライトキー） | `/v1/pets/{id}` は URL ごとにプリフライトキャッシュが別キー。バックエンドで pets 同梱またはバッチ取得にする。温間実測が出るまでは二次因扱い |
| 5 | `instance_type` 引き上げ（basic → 上位） | コールドスタートの起動時間そのもの | コスト増。案1と排他ではない（起動を速くしても sleep 自体は残る） |
| 6 | プリフライト自体の削減 | 温間時の往復 1 回分 × リクエスト数 | `X-Requested-With` は CSRF 防御で固定（「測定で維持する境界」参照）。GET から `X-Request-ID`・`X-Clinic-ID` 等を外せば simple request 化してプリフライト不要になるが、`Content-Type: application/json` を伴う POST/PUT/PATCH は safelist 外のため依然 preflight が必要。削減効果は GET 系に限定される |

補足: 案2・3・4・6 は「温間でも遅い」場合の二次因にも効くが、E4 時点では温間 API は速い（0.25 秒未満）ため優先度は案1が最大。ブラウザのプリフライトキャッシュは `(origin, URL, method)` 単位で、`Access-Control-Max-Age: 86400` を返しても Chrome 系は約2時間で切り捨てられるため、Max-Age 引き上げでは増幅は解消しない。

#### 費用影響（Cloudflare 公式 pricing 確認・2026-09-22）

Containers は稼働時間課金（10ms 単位、Workers Paid $5/月に含む）。単価: メモリ $0.0000025/GiB-秒（25 GiB-h/月込み、provisioned 課金）、CPU $0.000020/vCPU-秒（375 vCPU-分/月込み、2025-11 改訂で **アクティブ使用分のみ**）、ディスク $0.00000007/GB-秒（200 GB-h/月込み）。現行 `basic` = 1/4 vCPU / 1 GiB / 4 GB。

| 案 | 増分の目安（STG・単一コンテナ前提） |
|---|---|
| 1. keep-alive | **増加するが小さい**。24/7 常時起動 ≈ $7/月（メモリ $6.4 + ディスク $0.7、CPU はアイドル時ほぼゼロ）。営業時間のみ（約264h/月）≈ $2.4/月 |
| 2. edge OPTIONS | 実質ゼロ（Workers Paid のリクエスト枠内、STG 量は誤差） |
| 3/4/6. リトライ・N+1・プリフライト削減 | 微減（リクエスト数減） |
| 5. instance_type 引上げ | 稼働時間比例。basic→standard-1（4GiB/8GB）で 24/7 なら約 $28/月。sleep 維持なら増分は小さい |
| Redis系（現時点で不採用候補） | コールドスタートには無効。温間クエリの実測遅延が確認されてから検討。Upstash は CF 請求外、KV/Cache API はこの規模では実質無料。マルチテナント（clinic_id）分離の安全不変条件があり、エッジキャッシュには越境リーク設計が必要 |

結論として費用が問題になるのは案1と案5のみで、案1でも月 $2〜7 程度。scale-to-zero が節約しているのは月数ドルであり、対価として6〜16秒の起動待ちが発生している。

### 次に確認すること

1. 認証済みブラウザで `/owners` 等の実ページについて、コールド（10分超アイドル後）と温間の区間別時間を測る（既存チェックシートの run 票を使用）。
2. Workers Logs の `container_fetch_timing` で実リクエストの `duration_ms` を集計し、温間でも遅い API（二次因 = 実クエリやレスポンスサイズ）があるか切り分ける。未認証 curl では実クエリ時間を測れないため、この確認なしに「温間は全て速い」と断定しない。
3. 現行の常時ログが必要かを再判定する。必要なら対象と出力を限定し、不要なら撤去を別実装単位にする。

対象・承認・完了条件は [検証 TODO](#perf-stg-login)。旧候補の裁定は [履歴](docs/work/development-task-decisions.md) であり、現在の導入状態は上記を正とする。

#### 測定準備で決めること

ローカルで [測定チェックシート](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md) の run 票を作れる。匿名/既存session/復旧/医院選択ごとに、対象build、ブラウザ、cache条件、通常遷移/再読込、回数・間隔、同一時刻窓、provider証拠の取得担当、停止条件と保存先を記入する。実測前の結果欄は未実行。承認範囲に測定回数と対象操作を含め、承認のない外部trafficや負荷試験を開始しない。

比較表の列は「run/case、OPTIONS/GET各区間、FCP、待機表示、フォーム操作可能、認証成功、Worker受付/forwarding、Container起動証拠、時刻対応の可否」。非公開の生証拠は保管規則に従い、共有表には秘密除去した相対時刻と分類だけを置く。対応しない区間は UNKNOWN のまま、ブラウザ全体の待ちとWorker内時間を足し合わせて二重計上しない。

**SLO の採用値は未確定。** チェックシートの API p95 500ms / 初回操作可能1.5秒などは提案例であり、合否判定へ自動採用しない。一方、[既存STG k6](load-tests/k6-cf-stg-sustained.js) の p95 3秒・失敗率5%未満は `/health` と `/api/v1/clinics` を3 VUで測るスクリプト閾値で、ブラウザ `/login` の受入値ではない。今回の遅延調査にその負荷試験を追加せず、初回表示の区間測定から始める。

準備完了は6単位それぞれの入力・出力・未確定条件がケース票に揃った時点。MITIGATION は因果区間、BUNDLE は転送/parse/executeの寄与が分かるまで実装 DEFERRED。測定できても採用 SLO が未合意なら、計測完了と性能受入完了を分ける。

### SLACK-LATENCY: 治療数量の反映待ち

> Task detail migrated to Plane `EMR-104` and verified by readback. Historical/evidence material remains in linked source records.

### E1: 2026-09-09 の遅延記録（過去の測定）

HTML TTFB 49.2ms、DOMContentLoaded 303.1ms、load 309.1ms、FCP 23,548ms。`/api/v1/me` は401、HTTP/3、全体22,689.3ms。

| 区間 | 所要時間 |
|---|---:|
| fetchStart → 最終 GET の DNS 開始 | 22,546.3ms |
| DNS | 0.4ms |
| 接続 | 24.3ms |
| requestStart → responseStart | 115.8ms |
| 本文受信 | 2.3ms |

`workerStart=0`。当時の OPTIONS、相関 ID、接続交渉、Container 起動、配信 revision は未保存。約22.5秒は最終 GET の接続開始前であり、Worker 内の所要時間だけでは説明できない。→ E4 で Container 起動待ち（コールドスタート中にブラウザ側が接続を待つ区間）と整合することを確認。

### E2 / E3: 再現しなかった記録

同日の通常 Chrome 2回では `/me` 221.4 / 190.0ms、FCP 944 / 768ms（いずれも401）。単発 HTTP は `/login` 281ms、Cookie なし `/me` 254ms、OPTIONS 240ms。遅延が再現しなかった記録であり、認証済みブラウザ経路・cold start 解消・p95/p99 を証明しない。→ 温間状態なら速いことは E4 とも一致する。

### E4: 2026-09-22 curl 実測（コールドスタートの直接観測）

対象: `https://api.stg.noah-karte.com`（STG API、Cloudflare Workers → Containers → PlanetScale）と `https://stg.noah-karte.com`（Vercel）。未認証のため実クエリ時間は含まない。ユーザー報告「ほとんどのページで読み込みが遅い」を受けて実施。

| リクエスト | 計測 | 解釈 |
|---|---|---|
| `GET /health`（10分超アイドル後の初回） | TTFB 15.69s | Container 起動待ち |
| `GET /health`（2・3回目） | TTFB 0.24s / 0.25s | 温間は速い。Go/DB 自体の遅さではない |
| `OPTIONS /api/v1/me`（コールド） | TTFB 5.96s | プリフライトも同じコンテナ起動を待つ |
| `GET /api/v1/me`（直後） | TTFB 0.24s（401） | 温間 |
| `GET /api/v1/pets?page=1&limit=20` | TTFB 0.25s（401） | 温間。認証前のため実クエリ未含む |
| `GET /owners`（Vercel） | TTFB 0.32s（2.5KB shell） | Vercel 側は速い。遅延は API 側 |

構造上の確定事項（再確認済み）:

- `backend/worker/index.ts`: `sleepAfter = "10m"`、`getContainer(env.API_CONTAINER)`（名前なし = 全トラフィックが既定の単一インスタンスに集中）。
- `backend/wrangler.jsonc`: `instance_type: "basic"`（1/4 vCPU / 1GiB）、`max_instances: 3`。DB は PlanetScale 直結（Hyperdrive 非経由、`DB_SSL_MODE=verify-full`）、プール上限 `DB_MAX_OPEN_CONNS=10`/`DB_MAX_IDLE_CONNS=5`。コールドスタートでは Go 起動に加えてプール空の状態から PlanetScale への TLS 接続確立が入る。
- Worker は OPTIONS を含む全リクエストを `containerFetch` でコンテナへ転送する薄いプロキシ。エッジで OPTIONS を返さないため、プリフライトも起動待ちに入る。
- axios は全リクエストに `X-Requested-With`（CSRF）・`X-Request-ID`・選択中は `X-Clinic-ID` を付与し、`Content-Type: application/json` も safelist 外のため、全 API コールが preflight 対象。プリフライトキャッシュは `(origin, URL, method)` 単位で `Access-Control-Max-Age: 86400` を返すが Chrome 系は約2時間で切り捨て。
- axios は GET の 502–504（Worker の起動失敗時 503 `service_unavailable` を含む）を最大2回・1s/2s バックオフでリトライする。
- `ownerLoader` は飼主取得後にペットごと `/v1/pets/{id}` を並列取得（N+1。各 URL が別プリフライトキー）。
- 既存 cron は `SCHEDULER_NAME` の named コンテナを起こすのみで、既定名の API コンテナは温めない。

「ほとんどのページが遅い」と整合する解釈: STG はアクセスがまばらで、操作の合間に10分を超える間隔があくたびコンテナが停止する。次のページ遷移・API 呼び出しが起動待ちになり、プリフライト＋リトライ＋N+1 が体感をさらに悪化させる。

### E5: 2026-09-22 実装後の再測定（配置制約＋認証キャッシュ＋sleepAfter延長）

E4 以降にユーザー報告で発覚した追加原因と、採用した改善・その実測を記録する。E4 の「温間は速い」は **Container↔PlanetScale 間の越洋 RTT を未認証 401 では測れていなかった** ため部分的な結論だった。認証済み curl で実測すると温間でも `/v1/clinics` ~1.0s、`/v1/me` ~1.5s、`login` ~4.1s で、遅延はコールドスタートだけではなかった。

#### 追加で確定した原因

- **コンテナ↔DB の地理分離**: DB は `ap-northeast-2.pg.psdb.cloud`（PlanetScale = AWS ソウル）。当初コンテナは `ewr01`（米東部）で稼働し、認証ミドルウェアが毎リクエスト staff→account→assignments→clinics を逐次 DB 再検証するため、1 往復 ~190ms×クエリ数が積み上がっていた。
- **インスタンス配置はブート/ロールアウト毎に再抽選される**: 同一 DO 名 `cf-singleton-container` のまま `ewr01→bom09→maa01→sin14→bom09` とドリフトを観測。DO の `locationHint`/改名は初回 DO 作成時のみ効く best-effort で、**インスタンス再配置には効かない**（PR #423 の `api-apac-ne-v1` 実験は bom09 着地で撤回・PR #425）。
- **`constraints.cities` はこのアカウントで利用不可**: デプロイが `VALIDATE_INPUT: City-level placement requires INTERNAL or CITIES_CONSTRAINT capability` で失敗（run 35749806541）。メトロ粒度のピン留めはできない。
- `scheduling_policy: "regional"` を wrangler.jsonc に記載したが `wrangler containers info` は `default` を返し続ける。2026-09-23 EMR-202 で切り分け済み: deploy 時の PATCH で `regional` は送信され API も受理する(CI log の `default → regional` 差分 + `SUCCESS Modified application`)が、既存 application には永続化されず readback は `default` のまま(v69/70/71 で再現)。wrangler 側の設定未適用ではなく Cloudflare Containers API 側の既存 app への PATCH 不保持が原因。実質 application 作成時のみ有効で、反映には app 削除→再作成(破壊的・要承認)が必要。運用は [STG runbook](docs/ops/infra/staging/runbook.md)「Container placement」参照。**2026-09-24 EMR-213 で解消済み**: `regional` は application create 時にも `VALIDATE_INPUT` で拒否される（アカウント非対応）ため config を `default` 固定に変更し、乖離は無くなった。

#### 採用した変更（staging にマージ済み）

| PR | 変更 |
|---|---|
| #424 | STG限定の認証 resolver キャッシュ `CURRENT_ACCESS_CACHE_TTL_SEC=30`（vars→envVars→`os.Getenv`→`composition_auth.go` の env ゲート。未設定/0/負値ならキャッシュ無しで本番は従来通り）。`sleepAfter` 10m→1h |
| #426 | `containers[].constraints.regions = ["APAC"]` — 配置抽選を APAC メトロに限定（無料） |
| #427 | `Dockerfile.production` に `LABEL rollout="1"` — イメージ差分で新バージョンを強制ロールアウトし即時再配置を起こす仕掛け。`verify-agent-task.py` に Dockerfile の scoped 検証（`docker build --check`）を追加 |
| #428 | `scheduling_policy: "regional"`（PATCH は送信・受理されるが既存 app に永続化されず deployed=default のままと 2026-09-23 EMR-202 で確定。将来の app 再作成に備えた宣言として残置していたが、EMR-213 の app 再作成で `regional` は create 時も拒否されアカウント非対応と判明し `default` 固定へ変更済み） |
| #429→#430 | `cities` 試行→ケイパビリティ不足で失敗→撤回 |

#### 実測（認証済み・暖機・日本から）

| エンドポイント | E4 時点（ewr01） | 最悪時（bom09） | E5 現在 |
|---|---|---|---|
| `GET /health`（DB無し） | 0.25s | ~1.0s | **0.20s** |
| `GET /v1/clinics` | ~1.0s | ~1.5s | **0.47s** |
| `GET /v1/me` | ~1.5s | ~2.2s | **0.85s** |
| `POST /v1/login` | 4.1s | 5.4s | **3.5s** |

改善の内訳は越洋 RTT 削減（現行インスタンスは日本近辺と推定）＋認証キャッシュで resolver の逐次 DB 往復が消失＋コールドスタート頻度低下（sleepAfter 1h）の複合。

#### 残存事項・運用メモ

- 配置は依然ブート毎の抽選。APAC 制約は ENAM/EEUR の最悪ケースを防ぐだけで日本着地は保証しない。悪い着地（bom/sin/maa）を引いたら `LABEL rollout` をインクリメントして再デプロイすると再抽選できる。
- `login` 3.5s は bcrypt×1/4 vCPU 由来でリージョンと独立。`instance_type` 引上げ（案5）が残るが実コスト増。
- 認証キャッシュは 3148d229f の認可ギャップ修正とトレードオフ（権限変更が最大30秒遅延）。STG限定のため vars は本番 `wrangler.production.jsonc` には設定しない。
- 費用: sleepAfter 延長のみ課金に触れる。basic 1台の課金は ~$0.01/h（メモリ+ディスク支配、CPU はアイドル時ほぼゼロ）。通常デモ利用で月+数十〜百円、常時稼働化しても ~$7/月が上限。

### 測定で維持する境界

- [axios.ts](frontend/src/lib/axios.ts) の `X-Requested-With` は CSRF 防御。高速化のために無条件で削除しない。
- browser の接続待ち、Worker forwarding、Container 起動、Go request latency を分ける。
- 秘密・raw URL/query・本文・IP・個人情報を観測証拠へ出さない。
- 将来の変更時は既存依存を使い、対象候補を mount した隔離 Docker の scoped Vitest と worker typecheck を使う。依存インストールを伴う旧 umbrella コマンドは今回実行しない。STG 性能受入はローカル検証とは別。

参照: [測定チェックシート](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md) · [プロファイリングガイド](docs/ops/testing/PERFORMANCE_PROFILING.md)

### E5: 2026-09-23 perf-e5-residual キャンペーン結果（revision 2）

証拠の正本は `reports/perf-e5-residual-20260923/`。測定値は全て**計測時点の現行 STG 配信版**のもの。キャンペーンのコード変更は claim/PERF-E5-* ブランチまたは作業ツリー WIP にあり STG 未配備であり、下記の値に変更後の効果は含まれない。n=1・n=5 の単発観測であり p95/p99・SLO 達成を主張しない。

#### 実装単位（STG 未配備のコード変更）

| 単位 | 結果 |
|---|---|
| EDGE-OPTIONS | OPTIONS preflight を Worker edge で応答する経路を実装。ただし cookie 認証の simple request では preflight が発生せず、browser 証拠（client-trace・stg-acceptance ともに OPTIONS 0 件）では同改善は未励起 — 正直な記録として併記する |
| AXIOS-RETRY | axios の retry 動作を縮小 |
| N1-PETS | owner loader を owner 1リクエスト + clinic_ids 付き owner スコープのページネーションへ変更し、ペット毎 detail fan-out を解消 |
| AUTH-1RTT | `clinics.is_active` を resolver の JOIN へ畳み込み、非 admin の current-access を 1 RTT 化 |
| STG-LOGIN | `AcceptSharedPassword` ゲートを bcrypt より前へ移動。staging のみ、production は不変 |

#### 測定・判定単位

| 単位 | 取得証拠 | 主な値 |
|---|---|---|
| STG-MEASURE | [stg-measure](reports/perf-e5-residual-20260923/stg-measure/README.md)（curl、n=5 warm、中央値のみ） | health 0.199s / me 0.944s / clinics 0.572s / pets 1.108s / pets検索 1.100s / pets+include_deceased 1.103s。login 一回 3.33s |
| CLIENT-TRACE | [client-trace](reports/perf-e5-residual-20260923/client-trace/README.md)（headless Chrome、n=1×2条件） | `/login` FCP cold 692ms / warm 88ms、DCL 663/57ms、long task 0、OPTIONS 0。唯一の API 呼出は未認証 `/me` 401（cold 488ms） |
| CF-EVENTS | [cf-events](reports/perf-e5-residual-20260923/cf-events/README.md)（wrangler tail + cf-ray 相関、20秒5リクエスト） | `container_fetch` 866–3848ms が支配的。edge(KIX)+Worker shim は約60–115ms、Worker→DO は 5–7ms。稼働 instance は `maa01`（ingress は KIX、PlanetScale は ap-northeast-2）。`scheduling_policy` は deployed=`default` vs config=`regional` の乖離を記録 |
| STG-ACCEPTANCE | [stg-acceptance](reports/perf-e5-residual-20260923/stg-acceptance/README.md)（4ケース、n=1。非SLO証拠） | 匿名 `/login` 操作可能 +1163ms。既存 session は `/v1/me` 1475ms にゲート。復旧は +689ms で login フォーム。login POST 3882ms → 認証 UI 205ms。医院選択ステップは存在せず `mainClinicId` 自動選択。全ケース OPTIONS 0 |
| OBS-DECISION | [obs-decision](reports/perf-e5-residual-20260923/obs-decision/README.md) | `container_fetch_timing` 常時ログは **KEEP**。STG・production draft とも `head_sampling_rate: 1` を維持し、本番トラフィック実測後の再評価トリガーのみ記録 |

残る MITIGATION と BUNDLE の作業状態・次の一手は Plane の `PERF-V-MITIGATION` / `PERF-V-BUNDLE` に移行済み。ここでは当時の計測判断を履歴として保持する。配置の手順は [STG runbook](docs/ops/infra/staging/runbook.md) を参照。

### E6: 2026-09-23 perf-e5-postdeploy-verify キャンペーン結果（revision 1・デプロイ後検証）

証拠の正本は `reports/perf-e5-postdeploy-verify-20260923/`。E5 実装単位を含む perf マージコミット `453be4ecc`（2026-09-22T18:41:12Z）の STG 配信後に実測。配信版は worker `15a85636`（2026-09-23T04:09:34Z 作成）・コンテナ v70（観測窓の途中で v71・`sin14` へ再作成）。n=1・n=5 の単発観測であり p95/p99・SLO 達成を主張しない。

#### 単位別結果

| 単位 | 取得証拠 | 主な値・判定 |
|---|---|---|
| DEPLOY-VERIFY | [deploy-verify](reports/perf-e5-postdeploy-verify-20260923/deploy-verify/README.md) | **SERVES-NEW-BUILD**。serving worker `15a85636`（04:09:34Z）とコンテナ v70（更新 04:12:44Z・新規稼働 instance 04:10:55Z `maa01`）は perf コミットを後置。OPTIONS が edge で 204（Go middleware ヘッダ無し、TTFB 58–304ms）、GET は Go ヘッダ全件でコンテナを通過 |
| OPTIONS-EDGE | [options-edge](reports/perf-e5-postdeploy-verify-20260923/options-edge/README.md)（6 probe matrix） | **EDGE CONTRACT VERIFIED**。allowlisted origin は ACAO echo・ACAC=true・Max-Age 86400・TAO・`Vary: Origin` の完全セット、非 allowlisted は ACAO/TAO/Vary 無しの 204、`/_internal` は worker guard で 404、GET control は Go ヘッダ付きでコンテナ通過。OPTIONS 中央値 ~60ms |
| WARM-MEASURE | [warm-measure](reports/perf-e5-postdeploy-verify-20260923/warm-measure/README.md)（curl、n=5 warm、中央値） | warm 中央値は E5 baseline と統計的に同一（全端末 ±0.01s 内）。対照表は下記。login 単発 1.975s |
| LOGIN-MEASURE | [login-measure](reports/perf-e5-postdeploy-verify-20260923/login-measure/README.md)（login ×3、15s 超間隔） | 中央値 **1.471s** vs E5 ~3.33s（**−56%**）。bcrypt-skip（`AcceptSharedPassword` を bcrypt より前へ）は方向的に確認。帰属は warmth・初回ヒット費用と不可分で bcrypt 単独寄与は UNKNOWN |
| PLACEMENT | [placement](reports/perf-e5-postdeploy-verify-20260923/placement/README.md) | capture 時点の稼働 instance は依然 `maa01`。`scheduling_policy` deployed=`default` vs config=`regional` の乖離は継続。`constraints.regions=["APAC"]` は live だが India メトロを除外できない（`cities` はケイパビリティ不足で利用不可）。**LABEL rollout 再抽選を推奨** |
| CF-EVENTS | [cf-events](reports/perf-e5-postdeploy-verify-20260923/cf-events/README.md)（wrangler tail + cf-ray 相関、5 リクエスト） | `container_fetch_timing` は post-deploy でも 5/5 に存在。**1001–1988ms** vs E5 866–3848ms（上限が約半分に収束）。edge+worker ~57–63ms、Worker→DO +9–11ms。窓の途中で稼働 instance が `maa01`→`sin14`（v71）へ再作成 |
| BROWSER-PAGES | [browser-pages](reports/perf-e5-postdeploy-verify-20260923/browser-pages/README.md)（実 Chrome・認証済み `/owners/301164` ×2 run） | **N+1 解消を実ブラウザで確認**: owner detail 1 読込につき owner スコープの `GET /api/v1/pets?owner_id=301164…` がちょうど 1 本、per-pet fan-out 0。document TTFB ~15ms。**新規所見**: `accountings?owner_id=` が最遅 API（1.9–2.7s） |
| STG-ACCEPTANCE | [stg-acceptance](reports/perf-e5-postdeploy-verify-20260923/stg-acceptance/README.md)（4 ケース、n=1・非 SLO 証拠） | 匿名 `/login` FCP 320ms・フォーム操作可能 +972ms。login POST **1451ms**（E5 3882ms、**−62.6%**）→ 認証 UI 54ms。既存 session の `/v1/me` ゲート 1100.7ms（E5 1475ms、−25%）。復旧は 401→`/login` リダイレクト 304ms・フォーム +613ms。医院選択は引続き `mainClinicId` 自動選択でステップ非存在。全ケース OPTIONS 0 |

#### E5 との対照

| 指標 | E5（デプロイ前・perf-e5-residual） | E6（デプロイ後） | 差分・判定 |
|---|---|---|---|
| `GET /health` warm 中央値 | 0.199s | 0.202s | +0.003s（同一視） |
| `GET /v1/me` warm 中央値 | 0.944s | 0.946s | +0.002s（同一視） |
| `GET /v1/clinics` warm 中央値 | 0.572s | 0.568s | −0.004s（同一視） |
| `GET /v1/pets` warm 中央値 | 1.108s | 1.100s | −0.008s（同一視） |
| `GET /v1/pets?q=` warm 中央値 | 1.100s | 1.104s | +0.004s（同一視） |
| `GET /v1/pets?include_deceased` warm 中央値 | 1.103s | 1.103s | 0.000s（同一視） |
| `POST /v1/login`（curl 中央値） | ~3.33s | **1.471s** | **−56%** |
| `POST /v1/login`（browser、stg-acceptance case4） | 3882ms | **1451ms** | **−62.6%** |
| `GET /v1/me` 既存 session ゲート（browser） | 1475ms | 1100.7ms | −25% |
| `container_fetch_timing` レンジ | 866–3848ms | 1001–1988ms | 上限が約半分に収束 |
| edge+worker オーバーヘッド | ~59–115ms | ~57–63ms（warm） | 縮小 |
| OPTIONS preflight | コンテナ経由（コールド時は起動待ち ~6s）・browser 実測 0 件 | **edge で 204・中央値 ~60ms**・browser 実測は依然 0 件 | 契約面は検証済。cookie 認証の simple request では励起しない（下記） |
| 稼働 instance 配置 | `maa01` | `maa01`（capture 時）→ 観測窓内で `sin14` へ再作成 | APAC 内ドリフト継続 |
| owner detail の pets fan-out | N+1 解消を実装（実機未確認） | **実ブラウザで owner スコープ 1 リクエストを確認** | 実機検証済み |

#### 残存事項（E6 時点）

- **AXIOS-RETRY**: 実装・配備済みだが**フィールド観測は除外**——観測窓に 503 ストーム（起動失敗応答）が発生せずリトライ経路は未励起。**unverified-in-field** であり失敗ではない。確認は 503 発生時の再観測か、合成 fault 注入の別単位で行う。
- **INSTANCE-TYPE（案5）/ MITIGATION / BUNDLE**: いずれもトリガー付き DEFERRED のまま（INSTANCE-TYPE は login 等の CPU 拘束区間の更なる短縮要請と実コスト増の权衡、MITIGATION は因果区間の確定、BUNDLE は転送/parse/execute 寄与の実測）。作業状態の正本は Plane。
- **SLACK-LATENCY**: ユーザーレーン（Plane `EMR-104`）。本キャンペーンの対象外。
- **PERF-V-LINEAR**: E6時点は **BLOCKED**（Linear MCP が未接続・`USER_NOT_LOGGED_IN`）。`EMR-136` は 2026-09-27 照合で **Cancelled** 終了済み。
- **配置**: `maa01`/`sin14` いずれも日本非ローカルで APAC 制約は満たすが Japan 着地は保証しない。runbook の再抽選（`LABEL rollout` インクリメント再デプロイ）の実施は承認済み運用操作に委ねる。
- **新規候補**: `GET /api/v1/accountings?owner_id=` が owner detail 画面の最遅 API（1.9–2.7s、browser-pages 観測）。次期改善候補として記録する。

### E7: 2026-09-24 perf-lane-all-20260924 キャンペーン結果

証拠の正本は `reports/perf-lane-all-20260924/`。観測時点の STG 配信版は worker `a52c2795`（2026-09-23T14:55:37Z 作成）・コンテナ v72 で、EMR-201 の perf コミット `e818195ec` を含む（merge `da359419b` の compare: ahead 69 / behind 0）。migration `007_billings_clinic_owner_scheduled_index.sql` は適用済み（migrate log `missing=0`）。curl n=5 warm・wrangler 読取の単発観測であり p95/p99・SLO 達成を主張しない。

#### 単位別結果

| 単位 | 取得証拠 | 主な値・判定 |
|---|---|---|
| PERF-ACCT-VERIFY（EMR-201） | [accountings-verify](reports/perf-lane-all-20260924/accountings-verify/README.md)（curl、n=5 warm、中央値のみ） | **IMPROVED**。`GET /api/v1/accountings?owner_id=` warm 中央値 **1.091s**（min 0.968 / max 1.338、全件 HTTP 200・31,320 bytes）vs E6 baseline 1.9–2.7s。初回ヒット 2.659s は旧レンジ上限相当。単一 client・baseline との browser-vs-curl 手法差は正直な caveat として併記 |
| PERF-PLACEMENT-CHECK（EMR-202） | [placement](reports/perf-lane-all-20260924/placement/README.md)（wrangler 読取 + `/health` GET 1 本） | v72 で配置を再観測: `scheduling_policy` は依然 deployed=`default` vs config=`regional`（当時の `backend/wrangler.jsonc:130`。この乖離は EMR-213 で config を `default` 固定にして解消済み）。稼働 singleton は `maa01`、migrate-runner は `bom09`（いずれも APAC 制約内・日本非ローカル）。runbook・台帳の記載 8/8 が live と MATCH（ドリフト無し）。runbook 発火条件が成立するため **LABEL rollout 再抽選を推奨 YES**——実施は承認済み運用操作に委ねる |
| PERF-COST-MEMO（EMR-204） | [instance-type-cost](reports/perf-lane-all-20260924/instance-type-cost/README.md)（repo 内読取のみ・外部 call 無し） | 意思決定メモを作成。現行 `basic`（1/4 vCPU / 1GiB）。増分見積（ESTIMATE・sleepAfter 依存）: `standard-1` ≈ **+$2.8–20.4/mo**、`standard-2` ≈ **+$4.7–34.3/mo**。受益側は bcrypt（実パスワード login のみ・共有パスワード経路は bcrypt-free）と起動 CPU 区間に限定され一部 UNKNOWN。**open decision**: approve std-1 / std-2 / reject / defer |

#### 残存事項・対象外（E7 時点）

- **EMR-104（SLACK-LATENCY）**: ユーザーレーン（Plane Needs Human）。本キャンペーンの対象外。
- **EMR-203（AXIOS-RETRY フィールド検証）**: unverified-in-field のまま——観測窓に 503 が無くリトライ経路は未励起。失敗ではない（2026-09-27 照合: Plane Ready）。
- **EMR-142 / EMR-143（PERF-V-MITIGATION / PERF-V-BUNDLE）**: トリガー付き DEFERRED のまま（E7時点 Backlog → 2026-09-27 照合で Plane Ready）。
- **EMR-199（PERF-E5-STG-DEPLOY-VERIFY）**: 終端（Plane Done・prior campaign で完了）。
- **EMR-136（PERF-V-LINEAR）**: cancelled——Linear MCP 未接続（`USER_NOT_LOGGED_IN`）で照会不能のまま。
- **EMR-202 再抽選**: 2026-09-28 決定で**再抽選は実施せず完了**。`scheduling_policy` は `default` 固定（`backend/wrangler.jsonc:141`、`regional` は EMR-213 でアカウント非対応と確認）で config/deployed の乖離は解消済み、`constraints.regions=["APAC"]` 内の抽選が現行ケイパビリティの上限で `maa01`/`bom`/`sin` 着地も許容、再抽選（`backend/Dockerfile.production:38` の `LABEL rollout` bump + 再デプロイ）は発火条件を満たしたときの任意運用（[STG runbook](docs/ops/infra/staging/runbook.md)「Container placement」）。
- **EMR-204 open decision**: std-1 / std-2 / reject / defer の判断は運用者の承認事項（E7時点）。2026-09-27 照合で Plane **Cancelled** 終了を確認。再検討する場合は別単位として起票する。

### E8: 2026-09-24 post-recovery warm 再計測（単発観測・revision 1）

観測時点の STG 配信版は worker `b25be1b1`（2026-09-24T14:09:21Z）・コンテナ v11・image `a3fd8027`・`basic` 0.25vCPU・`scheduling_policy=default`・APAC 制約。稼働 singleton は `sin07`。seed checksum crash からの復旧（migration 009）直後の同日 23:35–23:50 JST に測定。curl n=5 warm・wrangler tail 相関の単発観測であり p95/p99・SLO 達成を主張しない。

#### 単位別結果

| 単位 | 取得証拠 | 主な値・判定 |
|---|---|---|
| KEEPALIVE-REG | CI deploy log（run 36010665511） | `Deployed animalekarte-stg-api triggers` に `schedule: 10,40 0-9 * * *` を含む 4 schedule の適用を確認。**登録済み・実発火は未観測**（デプロイ時点で当日の JST 09:10–18:40 窓は終了済み。初回発火は翌営業日 09:10 JST） |
| LATENCY-DECOMP | `wrangler tail` の `container_fetch_timing` + curl TTFB 相関 | `/health`: container_fetch **193ms**（curl 0.257s）。`/api/v1/me`: container_fetch **3953ms**（curl 4.01s）。**edge+DO ホップは健全（~190ms floor）で、遅延はコンテナ内部（Go+DB）に閉じる** |
| WARM-REMEASURE | curl n=5 warm、cookie 再利用（`stg-staff-10000021`・執行） | me ~1.5s（0.67–8.35）/ clinics ~0.78（0.38–1.18）/ pets ~2.3（1.5–6.0）/ pets?q ~2.5（0.92–4.9）/ pets+deceased ~2.0（0.66–3.4）/ accountings?owner_id ~3.3（2.5–6.7）。**全 endpoint が E6/E7 baseline を上回り、同一 endpoint で 4–12x のばらつき** |

#### 判定

コードは E6/E7 配信版からクエリ経路の変更がない（差分は keep-alive + migration 009 のみ）。にもかかわらず同一リクエストが 12x 変動することから、**支配因は環境**と判定：約24時間の crash loop 停止により PlanetScale のバッファ/接続が完全冷却し、加えて basic 0.25vCPU の共有 CPU ノイズが乗った状態。**新規のコードレベル N+1 回帰の証拠は確認できず**、現観測窓では修正対象の単一ボトルネックを確定できない（タスク規約どおり記録で留める）。

#### コードレビュー由来の bounded 所見（測定確度外・follow-up 候補）

1. **`/me` は access-cache miss 時（TTL 2s）に直列 ~5 query**: `loadCurrentAccessGraph` → `Staff.GetByID` → `Accounts.GetByID`（graph の account_id を借用すれば staff 待ち不要化可能）→ `ListClinicsByIDs` → `GetEffectivePermissions`。middleware 解決後の 4 本は独立で errgroup 並列化可能（errgroup は `accounting_repository.go` で既採用）。推定 saving ~2–3 RTT（warm で ~0.2–0.5s）。認証中核経路のため暖機状態での効果検証と慎重なレビュー前提。
2. **`/health` は静的応答で DB 非到達** — keep-alive cron はコンテナプロセスのみ温め、DB pool（`ConnMaxIdleTime=5min` < cron 30min 間隔）と PlanetScale バッファは tick 間で冷却する。**アイドル後の初回実リクエストは依然 fresh-TLS + cold-page を支払う**。低コスト選択肢: `SELECT 1` のみ行う軽量 endpoint を ≤5min cadence の cron で打つ（追加費用 ~$0・エンドポイント面の設計判断が必要）。
3. **アプリ内に per-query 計時がない**（GORM logger 無効）→ コンテナ内部の帰属はコード読みに依存。閾値付き slow-query ログ（payload 無し）の導入が次回分析の証拠力を上げる。

#### 次の一手（2026-09-27 照合: 全項目完了・Plane Done）

- keep-alive 実発火 + 暖機済み DB での再計測 → **実施済み**。keep-alive cron は `*/4 0-9 * * *` へ密度化し `/health/db` で DB pool も温める構成へ更新済み（`backend/wrangler.jsonc`・`backend/worker/index.ts`、`api_keepalive` ログ。PR #488/#489・PR #493 `5bbf46820`）。実発火確認は EMR-214、app 再作成・`scheduling_policy: "default"` 固定・再測定は EMR-213 で完了し、いずれも Plane Done。
- 暖機改善3候補も全て実装済み: (1) `/me` 並列化（`http_session_me.go` で errgroup 並列）(2) DB-ping keep-alive（`/health/db`）(3) slow-query ログ（`internal/dbconn/slow_query_logger.go`・`DB_SLOW_QUERY_MS`・テスト付き）— PR #493 `5bbf46820`。
