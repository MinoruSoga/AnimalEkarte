# AnimalEkarte — TODO

> 更新: 2026-07-26(3)（U2b-min 完了 `b8a29d3e7` — #251 納品前スコープ〔U1・U2a・U2b-min〕全完了・締め12分類が全書込経路で機能。納品前 dev 作業ゼロ。次 dev = #238〔残ギャップ=GetCourses の IsActive 未検査のみ・POST拒否は実装済みを実測確認〕）

## 運用

- 本書は、エージェントが直ちに着手できる未完了タスクの台帳とする。
- タスクは「個別タスク詳細」節に `### <タスクID>: <タイトル>` 形式で追加する。
- 対応済みセクションは削除し、完了記録はgit履歴と各実装testを正本とする。
- GitHub Issueと対応するタスクはIssueのstateを実測し、Issue一覧を本書へ重複掲出しない。
- release/運用gateは実装タスクと混在させず、[`q&a.html` OPS-13〜17](q&a.html#ops)と該当runbookで追跡する。

## 正本の境界

| 内容 | 正本 |
|------|------|
| 着手可能な実装タスク | 本書の「個別タスク詳細」 |
| GitHub Issueのstate・一覧 | GitHub Issues |
| BE9構造移行・進捗・release gate | BE9は2026-07-24にcode complete（release pending）。境界は[ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md) / [boundary map](docs/architecture/be9-2a-boundary-map.md)、release gateは[`q&a.html` OPS-13〜17](q&a.html#ops) |
| FEデザイン準拠・リファクタリング計画 | [`FE-refactor.md`](FE-refactor.md) |
| BE10 backend規約適合（フォルダ構成）リファクタ計画 | [`BE-refactor.md`](BE-refactor.md) |
| 今フェーズで着手しない事項 | [`phase2.html`](phase2.html) |
| 着手保留・任意検証のBE技術債 | [`BE-pending.md`](BE-pending.md) |
| PO判断・USER実操作・P0ブロッカー | [`q&a.html`](q&a.html) |
| Issueを3セッションで着手するためのガイドview（正本=各Issue・受け入れ条件を複製しない） | [`3-session-agent.html`](3-session-agent.html)（削除・退役の対象にしない） |

## 個別タスク詳細

### TASK-ADR003: 予約⇔会計の支払方法二重保持解消（ADR-003 案1B TRIGGER）

- PO-006 裁定済み。DEC-9（2026-07-25 q&a.html）で GitHub Issue 起票を待たず本書追跡へ変更。**着手時期 = 納品前（2026-07-26〜27）** — 2026-07-25 の納品日 7/27 延期（理由=残作業の全対応）に伴い「納品後」から前倒し。
- 内容の正本 = ADR-003 と q&a.html git 履歴（PO-006／DEC-9 カード・全量版 `ab68c61f5`。回答済みカードは 2026-07-26 に q&a.html から削除済み）。USER が Issue 起票したら本エントリを Issue へ移設し二重掲出しない。
- **実装完了（`c434c4e66` でコミット済み）・検証 BLOCKED（DB 適用は未実施 = q&a.html OPS-2）**: `backend/migrations/006_payment_splits_payment_method_clinic_fk.sql` で `payment_methods` に述語なしの `UNIQUE (id, clinic_id)` を追加し、`payment_splits (payment_method_id, clinic_id)` から `payment_methods (id, clinic_id)` への複合 FK を追加した。これにより `payment_splits` の他院支払方法 id 混入は DB 制約で拒否される。既存の単一列 FK は実 DB の制約名を確認できず、推測名での DROP による適用失敗を避けるため維持した。適用前の抵触行確認 SQL は migration コメントに記載した。
- **PO-006 案1Bとの差分**: 本 unit はトリガーを作らず、複合 FK で宣言的に表現できる clinic 一致だけを実装した。`method` ⇔ `payment_methods.system_key` の値一致は FK では表現できず、トリガーが必要なため未実施。通常の会計作成・更新経路では現行の `backend/internal/billing/accounting_service_builders.go:29-44` の `resolvePaymentMethodMasterID` が、当該 clinic の `system_key` から解決した id と request 由来 id の不一致を拒否している。一方、確定後訂正経路（`backend/internal/billing/accounting_service_correction.go:92-126`）は `method` / `payment_method_id` 自体を変更しないが、保存済みの組合せを再検証せずに split を保存するため、既存の値不一致を検出する防御にはならない。
- **未防御の `payments` と follow-up**: `payments` は `clinic_id` 列を持たず、tenant は `billing_id` 経由で辿るため、この unit の複合 FK を張れない。DB レベルでは未防御のまま残る。follow-up は (1) `billings` を join して clinic を照合するトリガー、または (2) `payments.clinic_id` を追加・backfill し同期責務を定めたうえで複合 FK を追加する、の 2 案。本 unit ではいずれも実施しない。
- **適用責任**: migration の DB 適用、既存データの事前照合、適用後の制約確認は利用者と CI の所掌。本 unit では DB query、migration apply、`schema_migrations` 更新を実行していない。
- **検証記録**:
  - saved prompt validator: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-adr003-payment-splits-clinic-composite-fk.md` → exit 0、`Prompt Craft Harness Validation: PASS`
  - CASCADE: `grep -n "CASCADE" backend/migrations/*.sql | grep -v "001_init.sql"` → `backend/migrations/005_exam_reference_ranges_and_clinic_fk.sql:41:    ON DELETE CASCADE;` のみ。`grep -n "CASCADE" backend/migrations/006_payment_splits_payment_method_clinic_fk.sql` → stdout 0 行。
  - lintscan: `docker compose exec backend go test ./internal/lintscan/ -count=1` → exit 0、`ok  	github.com/animal-ekarte/backend/internal/lintscan	1.107s`
  - trackability: `git status --porcelain -- backend/migrations/` → `?? backend/migrations/006_payment_splits_payment_method_clinic_fk.sql`。directory / exact-file の `git check-ignore -v` → stdout 0 行。`git diff --cached --name-only` → stdout 0 行。
  - non-modification: 既存 migration 5 本への scoped `git status --porcelain` → stdout 0 行。実装直後の `git status --porcelain -- backend/internal backend/cmd` → stdout 0 行。最終確認時に別 session の `backend/internal/medicalrecord` / `backend/internal/repository` / `backend/internal/clinic` WIP が出現したが、本 unit の pre-edit baseline と書込み自己申告には含まれず、触れていない。
  - ADR additions-only: `git diff --numstat -- docs/architecture/adr/003-payment-method-identity-and-consistency.md` → `6	0	docs/architecture/adr/003-payment-method-identity-and-consistency.md`
  - scoped increment: `comm -13 "${TMPDIR:-/tmp}/adr003-baseline.txt" "${TMPDIR:-/tmp}/adr003-final.txt"` → ` M docs/architecture/adr/003-payment-method-identity-and-consistency.md` / ` M todo.md` / `?? backend/migrations/006_payment_splits_payment_method_clinic_fk.sql`
  - independent Santa review: 初回は database reviewer / clinic-isolation reviewer とも verdict `PASS`。ledger 追記後の fresh round で、database reviewer は `PASS`、clinic-isolation reviewer はアプリ層の値一致保証を広く書きすぎている点を HIGH と判定。実測では訂正経路は `method` を変更しないため finding の一部を反証したが、保存済み組合せを再検証しない点は採用して記述を通常作成・更新経路へ限定した。修正後の fresh dual review は両者 `PASS`、未解消 CRITICAL / HIGH 0。
- **Failure Signature（AC-12、BLOCKED）**: expected=`docker compose exec backend go test ./internal/model/ -count=1` が `ok`、actual=exit 1、verification=同 command、error=`[ExamTypeField.clinic_id] Goモデルにフィールドがあるが、テーブル "exam_type_fields" にカラムが存在しない` および `allModels() に未登録のモデル ... ExamReferenceRange`、attempt 1=package 全体、attempt 2=各 failing test へ縮小、attempt 3=`TestSchemaDrift` を pipe 無しで再実行して exit 1 を確認、attempted fix=なし（migration 005 / model inventory の既存 drift であり allowlist 外）、result=BLOCKED。本 unit は列も Go model も変更しておらず、ゲートを緩める変更は行わない。
- **Assumption deviations / prompt drift**: 既存データ適合性は指示どおり未照会。`resolvePaymentMethodMasterID` は prompt / ADR 記載の `backend/internal/service/...` ではなく、現行 `backend/internal/billing/accounting_service_builders.go:29-44` に移動済み。その他の AC-01 schema 前提に差異なし。
- **model inventory unit 完了（2026-07-26・`c434c4e66` でコミット済み）**: `backend/internal/model/schema_drift_test.go` の `allModels()` に `&model.ExamReferenceRange{},` を既存の pointer 記法で追加した。変更 path は同ファイルと本 `todo.md` のみ。`knownSchemaDriftAllowlist` / `knownNullabilityDriftAllowlist`、検査ロジック、`backend/internal/model/exam_reference_range.go` は未変更。
- **model inventory unit 検証記録**:
  - saved prompt validator: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-model-inventory-examreferencerange.md` → exit 0、`Prompt Craft Harness Validation: PASS`
  - 着手前 package: `docker compose exec backend go test ./internal/model/ -count=1` → exit 1、`--- FAIL: TestSchemaDrift (0.30s)` / `[ExamTypeField.clinic_id] Goモデルにフィールドがあるが、テーブル "exam_type_fields" にカラムが存在しない` / `--- FAIL: TestAllModelsExhaustive (0.01s)` / `allModels() に未登録のモデル (1件、TableName()実装あり): ExamReferenceRange`
  - 対象テスト: `docker compose exec backend go test ./internal/model/ -run TestAllModelsExhaustive -count=1` → exit 0、`ok  	github.com/animal-ekarte/backend/internal/model	0.008s`
  - 対象テスト実行明示: `docker compose exec backend go test ./internal/model/ -run '^TestAllModelsExhaustive$' -count=1 -v` → exit 0、`=== RUN   TestAllModelsExhaustive` / `--- PASS: TestAllModelsExhaustive (0.01s)` / `PASS`
  - 実装後 package: `docker compose exec backend go test ./internal/model/ -count=1` → exit 1、失敗テストは `--- FAIL: TestSchemaDrift (0.29s)` のみ。指摘は `[ExamTypeField.clinic_id] Goモデルにフィールドがあるが、テーブル "exam_type_fields" にカラムが存在しない` と `[ExamReferenceRange] テーブル "exam_reference_ranges" がDBに存在しない` の2件。`TestAllModelsExhaustive` は失敗集合から消え、新規の失敗テストは 0。
  - format / vet: `docker compose exec backend gofmt -l internal/model/schema_drift_test.go` → stdout 0 行、exit 0。`docker compose exec backend go vet ./internal/model/` → stdout 0 行、exit 0。
  - non-modification / diff: `git status --porcelain -- backend/internal/model/exam_reference_range.go` → stdout 0 行。`git diff -- backend/internal/model/schema_drift_test.go` → `+		&model.ExamReferenceRange{},` の1行追加だけ。
- **model inventory unit の未解消環境状態**: `TestSchemaDrift` はライブ DB へ接続するため未解消。ローカル DB に migration 005（`backend/migrations/005_exam_reference_ranges_and_clinic_fk.sql`）が未適用であり、DB への適用は利用者の操作である。本 unit は migration 適用、DB query、allowlist 追記を実行していない。登録後に同テストの指摘が1件増えたのも、同 migration 未適用により `exam_reference_ranges` table が存在しないためである。
- **model inventory unit 非該当 / prompt drift**: production code と新規生成物を作らず index に触れないため、coverage threshold、tracked-or-not-ignored probe、staged-path listing は非該当。prompt 記載の型名・主な test 名・failure 内容は実測と一致したが、`TestAllModelsExhaustive` 本体の開始行は `:508` 付近ではなく現行 `:563`（`:509` は `os.ReadDir(".")` を行う helper）だった。

### TASK-251: 締め集計 category contract 確定実装（#251・8→12分類）

- 業務決裁確定（DEC-21・**USER 本人裁定** 2026-07-25 — #251 本文へ転記済み・原文 = q&a.html git 履歴）。**着手時期 = 納品前（2026-07-26〜27）** — 2026-07-25 の納品日 7/27 延期（理由=残作業の全対応）に伴い S3 送りから前倒し。contract 正本 = #251 本文（DEC-21 転記済み）（本エントリは実装スコープと着手時期の入口であり決裁の「なぜ」は複製しない）。
- Phase 0 棚卸し（外部エージェント調査・Fable spot-verify）で確定した実装スコープ:
  - ① 正式カテゴリ = 12分類（enum 現状追認）。#251 タイトル「8分類」→「12分類」修正は Issue 本文転記（USER 承認後）に含める。
  - ② hospitalization 退院会計の other 固定を撤廃し CarePlanItem.Type／Procedure.IsSurgery→category resolver（`backend/internal/medicalrecord/hospitalization_service.go:431`）。treatment 経路（`backend/internal/billing/billing_item_service.go:405,462`）と共通化＝category contract 単一ソース化。
  - ③ vaccination を接種記録（`Vaccination.VaccineID`→Vaccine）から会計明細自動生成。`BillingItem` へ VaccineID provenance 列追加の migration が必要。自動化は停止／失敗通知／監査／idempotency（原則⑤）。
  - ④ hotel source=`HospitalizationTypeHotel`（②連動）、training は新規 source 設計。両カテゴリ維持。
  - 含意(a) category authority を BE resolver に一本化し FE/client は保持しない。
  - 含意(b) 締め集計の未知値 fail-closed = 生カラム無制限 GROUP BY（`backend/internal/billing/accounting_repository_reports_close.go:44`・`cash_register_service.go:265`）を12値 allowlist 経由にし typo/legacy を締め表へ黙って通さない（受け入れ条件「unknown/legacy を黙って変換しない」）。
  - 含意(d) 全書込経路（treatment/hospitalization/vaccination/trimming/merchandise/manual）を同一 typed category source に集約。
- #247（月次統合表）は本 TASK の contract 完了後に着手。
- Issue #251 本文への決裁転記（タイトル「8分類」→「12分類」修正含む）と着手時期の前倒し反映は、いずれも 2026-07-25 に USER 承認のうえ完了済み（live read-back で実測確認）。todo.md / DEC-21 / #251 本文の3者は同期済み。以後 contract の参照先は #251 本文（DEC-21 原文 = q&a.html git 履歴）とし、本エントリは着手時期と実装スコープのみを持つ。
- **U1 完了（`2154dc9de`・2026-07-25）**: category resolver を `backend/internal/sharedkernel/item_category_resolver.go` へ新設し、BE 導出の3経路（入院退院会計・外来診療明細・トリミング明細）を集約。②の other 固定撤廃と surgery 導出、④の hotel 到達経路を実装済み。`model.ItemCategory` 具体定数を参照する production ファイルは 4 → 3 に減少。resolver は error を返さない全域関数（退院処理の tx 閉包内にあるため、マスタ不整合で臨床業務を止めない）。
- **U1 で判明した scope の限界**: カテゴリ決定点は4つあり、うち `billing_item_service.go:258`（手入力・物販）は client 送出値をそのまま永続化する。含意(a)「category authority を BE resolver に一本化し FE/client は保持しない」の達成には API 契約変更が要るため U2 送り。
- **残ユニット**: U2 = 手入力・物販経路（下記のとおり U2a/U2b に再分解）。U3 = ③ ワクチン接種記録からの明細自動生成＋`billing_items` への VaccineID provenance 列 migration（停止手段・失敗通知・監査・idempotency 必須）。U4 = ④ training の新規 source 設計。U5 = 含意(b) 締め集計の allowlist 化。
- **U2 の再分解（2026-07-25 実測により当初定義「API 契約変更要」を訂正）**: API 契約は BE/OpenAPI 側に既に存在し、FE が使っていないだけである。
  - **U2a-1 完了（`33690944e`・2026-07-25）**: FE の物販マスタ選択経路で `merchandise_item_id` を送出（既存会計 POST と未作成会計の逐次作成の両経路）。`onAddItem` をオブジェクト引数化。`AccountingItem.merchandiseItemId` 追加。BE 側は 2 件是正 — ① `GetDiscountSuggestions` が `FindAllApplicableForItem` へ `nil` をハードコードしており、自動適用側（`resolveAutoDiscount`）と参照するキャンペーン集合が食い違っていた ② `BillingItemResponse` と OpenAPI `BillingItem` が `merchandise_item_id` を返しておらず、FE 生成型だけが当該フィールドを持つ状態（**BUG-433 と同型の乖離**）だった。②は生成側 prompt が「応答型の変更は不要」と誤って断定していたもので、外部エージェントが実測で反証した。検証 = frontend scoped vitest 41 passed/3 skipped、backend `ToBillingItemResponse|DiscountSuggestion` PASS、変更 4 ファイルの gofmt exit 0。**BUG-436 はこれで解消**。
  - **U2a-1 の残課題（次ユニットへ）**: `AddAccountingItemInput.category` および `FrontendMerchandiseItem.category` が `string` のままで、`use-accounting-item-actions.ts:45` の `category as ItemCategory` キャストが残る（本変更以前からの既存問題であり、悪化はしていない）。解消には `frontend/src/features/accounting/api/get-merchandise-items.ts` の型付けが要る。
  - **U2a-2 完了（`24d4ae2fd`・2026-07-26）**: BE `CreateItem` が `merchandise_item_id` 指定時に client 送出の `category` ではなく master の値を保存するよう是正。**当初案（v1）は pre-tx・ロックなしで master を読む設計で、外部エージェント実行が独立検証の末 BLOCKED を返した** — TOCTOU race（`.claude/refs/backend-application-invariants.md` の「並行するmaster変更がinvariantを壊す場合は参照行をcommitまで固定する」に違反）・`resolveAutoDiscount` が client 送出の stale category でキャンペーンを検索する不整合・新規 nil 依存が fail-open、の3件 HIGH。生成側で invariant 文書を実測し3件とも正当と確認、v2 として再設計: 新規依存を追加せず、既存の `ValidateCreateReferences`（`merchandise_item_id` を SHARE ロック付きで存在確認する既存クエリ）を `Select("id","category")` へ拡張し、**同一クエリ・同一ロックで category も取得**するよう変更（戻り値 `error` → `(model.ItemCategory, error)`）。`resolveAutoDiscount` には解決済み category を渡し、保存値とキャンペーン検索の参照 category を常に一致させた。副産物として v1 が必要としていた `NewBillingItemServiceWithCampaign` の約23箇所シグネチャ追随が丸ごと不要になった。
  - 検証: updater-wins 統合テスト（未コミットの category 更新 tx を保持したまま `CreateItem` を並行実行し、SHARE ロック待ちで確実にブロックされること・コミット後の新 category のみが保存されること・campaign 検索も同じ category を受け取ることを実DBで検証。`-race -count=20` でも安定）。`internal/billing` 全体 `ok 12.584s`、`go vet` exit 0、変更4ファイルの `gofmt -l` exit 0、golangci-lint 新規指摘0件。Mode 3 で生成側が自ら該当テスト・vet・gofmt を再実行し独立確認済み。
  - **U2a = 物販経路の master 由来導出。完了（U2a-1 + U2a-2）**。`merchandise_items.category`（`item_category NOT NULL`）が正本であるにもかかわらず、`ItemListCard.tsx:104` が master の category を値コピーして送出しており二重管理になっていた問題と、`merchandise_item_id` 未送出による master への link 断絶（BUG-436）を解消済み。
  - **U2b = 純手入力経路（導出元が存在しない）**。provenance が無いため導出は原理的に不可能であり、含意(a) の「FE/client は category を保持しない」はこの経路では達成できない。正しい終端は「12値 validated 受理」であり、BE 側は `validators_billing_item.go:14`（12値検証）で**既に完了している**。残る欠陥は FE の `ItemListCard.tsx:120` が `"other"` 固定で category を送っている点のみ（利用者は選択できず、締め集計で手入力明細が全て other へ落ちる）。カテゴリ選択 UI を足すか "other" 固定を仕様として明文化するかは入力工程を増やす判断であり USER 決裁とした → **DEC-22 論点1で裁定済み（2026-07-26・案A二段階）**: **U2b-min 完了（`b8a29d3e7`・2026-07-26）** = 手入力ダイアログへ12分類の必須 Select（デフォルトなし・other 選択可＝理由不要）。CATEGORY_LABELS の欠落3キーを既存業務用語（RV/ホテル/トレセン）で補完し FE4-1 由来の部分集合設計を解消。**U2b-full（納品後・U3/U4 と同 batch）** = DEC-16③の残り（other 理由必須の永続化＋締め表「未分類・要確認」件数表示）。
- **含意(b) の優先度を実測で再評価（DEC-22 論点2で降格承認済み 2026-07-26 — U5 は納品後の U3/U4 と同 batch）**: `billing_items.category` は PostgreSQL enum `item_category`（`001_init.sql:108`・12値）の `NOT NULL` 列であり、typo/legacy 文字列は at rest で存在し得ない。(b) は現時点の typo 防御としては no-op。意味を持つのは「将来 enum に値を追加した際に締め表の表示リストが追随せず黙って落とす」ドリフト防御であり、納品前クリティカルではない。DEC-21 は USER 本人裁定のため降格可否を USER 判断としたが、ユーザー委任「PO判断待ちはあなたが判断」に基づき DEC-22 論点2で降格承認（PO は q&a.html 上書きで覆せる）。
- 出典: #251 Phase 0 棚卸し Completion Report（2026-07-25・DEC-21）。U1 実行結果 Completion Report（2026-07-25・Mode 3 独立検証済み）。

### SEC-SWEEP-02: grandchild FKの親相関掃引 + 同型欠陥のstatic lint新設

- SEC-SWEEP-01（掃引 `3321c801f`、修正 `c16c011f2`＋`0736cd6f9` で **完了**）が明示的に繰り越した2件。掃引はpet直下の子テーブルreadのみを対象としており、孫テーブルは別クラスとして未着手。
- **① grandchild相関の掃引（read-only調査）**: `daily_records` / `care_logs` / `exam_results` / `billing_items` / `medical_record_images` / `medical_record_addenda` 等、petの孫にあたる表のreadが中間表を経由した親clinic相関を持つかをschema-firstで全数確認する。SEC-SWEEP-01と同じ手順（schema universe → model mapping → consumer discovery → state分類）を適用する。
- **② static lintの新設**: 同型欠陥（`pet_id`等の単一FKに対し親clinic相関を欠くread）をraw SQLとGORMの双方で検出するlintを既存6 lintの隣へ追加する。BUG-429の1件もSEC-SWEEP-01の9件も人手監査でしか見つかっていない。機械で止めない限り同じクラスが再発する。
- 修正の参照実装 = `backend/internal/pet/chronic_condition_repository.go:37,52`。相関にpets側の `deleted_at` / `deceased_at` を含めない制約は本掃引にも適用する（含めるとsoft-delete済み・死亡ペットの履歴が黙って消える挙動回帰になる）。
- **着手 = 納品後**。SEC-SWEEP-01の9件と異なり現時点でlive exposureは未確認であり、納品前クリティカルではない。
- 出典: SEC-SWEEP-01実行結果のcalibration / follow-up節（2026-07-25・掃引完了に伴い本エントリへ移設）。

### BUG-430: stage-importの医院非限定DELETE

- CRITICAL。`backend/cmd/stage-import` のdeleteScopeが `owner_id >= 300000`（pets経由の継承含む）でclinic_id非限定。実行すると他院の高番ownerデータを削除し得る。`backend/cmd/stage-import/main_test.go:217-246` がこの挙動をテストで固定化している（cross-clinic保護テストは無い）。
- 対応方針変更（2026-07-25 DEC-20）: **stage-import退役で解消**（deleteScope修正には投資しない）。本番cutoverはrunbook既定の21表csv-import正式経路であり本ツールは本番使用禁止・local限定＋`--confirm-local-destroy`ガード既存。退役実装=cmd/stage-import削除またはビルド除外（#250再基準化転記とセット・USER承認後）。
- 出典: #251調査 Completion Report（2026-07-25）。テスト実測で確認済み。

### BUG-433: 生成FE型がGoドメインモデル由来のため、応答DTOに無いフィールドが型上は存在扱いになる

- HIGH（サイレント機能不全の生成器）。**S3/S2いずれにも属さない横断課題**。`frontend/src/types/generated/models.ts` は tygo が `backend/internal/model/` から生成しており（同ファイル冒頭コメント）、OpenAPI／応答DTOからではない。このため FE の型は *Goドメインモデル* を写し、HTTP が実際に返す *応答DTO* とは一致しない。DTOに無いフィールドは実行時 `undefined` なのに型検査は通る。
- 実害の実例: BUG-431（受付の危険度バッジが実APIで一度も点灯しなかった・`463e07424` で修正）は本ドリフトの1インスタンスに過ぎない。fixtureは型どおり作られるためテストでも検出されない。
- 実測された残存ギャップ: 生成 `Pet` は31プロパティ、修正後の予約pet DTOは9。残22フィールド（`clinic_id` `owner_id` `animal_species_id` `name_kana` `gender` `birth_date` `color` `blood_type` `microchip_number` `neutered_date` `acquisition_type` `food` `environment` `phone` `last_visit` `insurance_id` `remarks` `deceased_at` `deceased_reason` `created_at` `updated_at` `insurance`）は型上は利用可能だがワイヤに存在しない。他モデル（Owner/Reservation等）も同構造。
- 対応方針（未確定・要判断）: ①応答DTOからFE型を生成する経路へ切り替える ②生成型を「ドメインモデル」と明示リネームし、画面が使う型は応答DTO由来へ分離する ③現状維持で個別に埋める（BUG-431と同じ対症）。①②は生成基盤の変更を伴うため納品後が妥当。納品前は、新規に生成型のフィールドへ依存する実装を書くときに**そのフィールドが応答DTOに実在するかを都度確認する**運用で凌ぐ。
- 出典: BUG-431 修正時に判明したNew Work（2026-07-25・executorが残22フィールドを実測列挙）。

### BUG-435: 生成FE型が陳腐化したままmainに乗っていた（codegen-check未励行）

- MEDIUM。FE12-07 で `make codegen` を回したところ、意図した型mapping修正（`any` 17→0）とは無関係な追随差分が同時に出た: audit定数7件（`AuditActionAuthPasswordChange` / `Reset` / `AdminReplace`、`AuditActionTrimmingCreate` / `Update` / `Delete`、`AuditResourceAccount` / `AuditResourceTrimming`）と `TokenBlacklist` のdoc comment。つまり **Go model を変更した際に `make codegen` が回されず、`frontend/src/types/generated/models.ts` が陳腐化した状態でmainに乗っていた**。
- `Makefile:349` に `codegen-check: codegen` + `git diff --exit-code frontend/src/types/generated/` が存在し、本来これがCIで検出する。検出されなかった理由（CIに未配線／配線済みだが未実行／実行され無視された）は未確認。
- 実害: FEが存在しない定数を参照しても型検査が通らないだけで安全側だが、逆に**BEが追加した定数をFEが使えない**状態が黙って続く。BUG-433（生成型がドメインモデル由来で応答DTOと不一致）とは別問題であり、そちらは構造の誤り、本件は同期の欠落。
- 対応: `make codegen-check` のCI配線状況を実測し、未配線なら追加する。配線済みなら失敗が無視された経路を塞ぐ。
- 出典: FE12-07 実行時に判明したNew Work（2026-07-25・`git diff frontend/src/types/generated/models.ts` 実測）。FE-refactor.md 範囲外のためこちらへ記載。

### BUG-437: `ExamTypeField` に `ClinicID` を追加したが read 側 clinic-scope registry へ未登録（cross-tenant read IDOR の再発リスク）

- HIGH（clinic 分離）。**`b4d10e083`（#249 Phase 2 U1）で main に入っている。現在 `main` でゲートが FAIL する。** 同 commit が `backend/internal/model/examination_type.go:32` の `ExamTypeField` に `ClinicID uint64` を追加した。write 側 registry には既に登録済み（`backend/internal/lintscan/master_fk_write_inventory_lint_test.go:129` の `"ExamTypeFieldID": "ExamTypeField (sub-master of ExaminationType, #124)"`）だが、**read 側の `clinicScopedMasterAssoc`（`backend/internal/lintscan/preload_clinic_scope_lint_test.go`）に entry が無く、`masterModelReadWriteExemptions` の例外にも入っていない**。
- 実害: clinic-scoped master を FK 値で Preload する際に `clinic_id` 述語が無いと、汚染された FK（#124/#125 の write 側検証ギャップ、または是正前データ）に対して**他院の master 名・値が返る**。b3638d5e で手作業修正した cross-tenant read IDOR と同型である。
- 検出経路: BE10-2 B5（`b28c4a105`）が `TestMasterModelReconciliation_RealSourceIsConsistent` を復旧した直後に検出した。**B4（`c430072d8`）以降このゲートは `no such file or directory` で死んでおり、B5 が直すまでこの追加は無検出で通過する状態だった。**
- 対応（3択・当該変更の owner が選ぶ）: ① Preload が既に存在するなら `clinicScopedMasterAssoc` へ association 名を登録する ② `internal/model` に association field が無いなら `masterModelReadWriteExemptions` へ理由付きで追加する ③ `ClinicID` の追加を撤回する。**ゲートを黙らせる目的だけの ② は禁止**。
- 再現: `docker compose exec backend go test ./internal/lintscan/ -run TestMasterModelReconciliation_RealSourceIsConsistent -count=1`（`b4d10e083` 以降の main で FAIL する。2026-07-25 に実測）。**このゲートが green に戻るまで `internal/lintscan` の FAIL 集合は R-4 と本件の2件である。**
- **部分対応・BLOCKED（2026-07-25・working tree）**: ①登録を選択し、read 側 registry に一意な association 名 `"ExamTypeField": "ExamTypeField"` を追加した。`Items` は `ExamResult` / `BillingItem` / `EstimateItem` / `LabExamResultItem` にも使われる汎用名のため登録していない。`ExamTypeField` へ解決する既存の `Preload("Items")` 2件（`backend/internal/medicalrecord/exam_type_repository.go:37,46`）は既に `"clinic_id = ?"` を持ち、DB-backed cross-tenant test も green なので現行 runtime read は安全。`masterModelReadWriteExemptions` と write 側 `clinicScopedMasterFKField` は無変更。
- 変更ファイル: `backend/internal/lintscan/preload_clinic_scope_lint_test.go`, `todo.md`。検証: 対象照合ゲート = `ok   github.com/animal-ekarte/backend/internal/lintscan 0.008s`、read lint（登録直後・最終）= `ok   github.com/animal-ekarte/backend/internal/lintscan 0.352s` / `0.355s`、DB-backed cross-tenant test = `ok   github.com/animal-ekarte/backend/internal/medicalrecord 0.260s`、`gofmt -l` = stdout 0 行、`go vet ./internal/lintscan/` = exit 0。
- 独立 clinic-isolation / code review で HIGH 1件が残った。read lint は Preload path の末尾 association 名だけで registry を引き、未登録名を無視する（`backend/internal/lintscan/preload_clinic_scope_lint_test.go:245-256`）。live read は `Items` なので、この2件から `clinic_id` を削除しても lint は検出しない。`Items` の一律登録は上記の同名別モデルを誤検出するため不可。解消には親 model / site を識別する context-aware rule または association 構造変更が必要で、いずれも本 unit の registry 登録予算を超える。したがって照合ゲートは green だが、機械的な再発防止が未完成のため BUG-437 は未解消。
- `internal/lintscan` package は 2026-07-25 時点で **green**（R-4 = `9ca93f249` の stale sentinel 是正、BUG-438 = `296ea7bb7` の CASCADE allowlist 登録で残 FAIL 2件が解消済み）。したがって本件は「ゲートが赤い」問題ではなく、**live read 経路 `Preload("Items")` を機械的に守る手段が無い**という再発防止の穴として残っている。
- 出典: BE10-2 B5 の Mode 3 照合（2026-07-25・生成側が独立実測）。

### BUG-439: legacy constructor `NewMedicalRecordService` 経由では finalize 不能（コンストラクタ二重化の解消）

- MEDIUM。`cb01009bd`（カルテ finalize・訂正追記の監査 same-tx fail-closed 化）の残余。source 互換の legacy constructor `NewMedicalRecordService` は transactional audit logger を注入しないため、この経路で組んだ service は finalize できない。production composition は `NewMedicalRecordServiceWithTxAudit` 配線済みで**実害なし**（Go reviewer が MEDIUM 判定・8ファイル契約のためコンストラクタ統合は見送り）。
- 対応: 呼び出し元を全数実測し、legacy constructor をテスト専用へ降格するか `WithTxAudit` へ一本化する。着手 = 納品後。
- 出典: Item B 実行 Completion Report（2026-07-26・Mode 3 照合済み）。

### R-1: clinic request/input間の型変換に対するstaticcheck S1016是正

- 対象: `backend/internal/clinic/closing_settings_request.go:11`, `:46`。`UpdateClinicSettingsRequest`→`UpdateClinicSettingsInput` と `UpdateSpecialPeriodRequest`→`UpdateSpecialPeriodInput` をstruct literalではなく型変換で書くべきという既存`staticcheck S1016` 2件。宣言元は同fileの`:3`/`:37`と`backend/internal/clinic/closing_settings_service.go:63`/`:80`。
- 再現: `docker compose run --rm --no-deps -T -e GOLANGCI_LINT_CACHE=/tmp/glc-be10-residual --entrypoint golangci-lint backend run ./internal/clinic/... ./internal/testdb/... --max-same-issues 0 --max-issues-per-linter 0`
- 発見元: BE10-1のscoped lint。同fileはBE10-1の変更8 pathに含まれず、当該unitが持ち込んだ問題ではない。BE10ではdrive-by修正を禁止しているため修正せず、本タスクへ移管した。

### R-2: testdb AutoMigrate errorのwrapcheck是正

- 対象: `backend/internal/testdb/testdb.go:100`。`gorm.DB.AutoMigrate`のerrorを未wrapで返している既存`wrapcheck` 1件。
- 再現: `docker compose run --rm --no-deps -T -e GOLANGCI_LINT_CACHE=/tmp/glc-be10-residual --entrypoint golangci-lint backend run ./internal/clinic/... ./internal/testdb/... --max-same-issues 0 --max-issues-per-linter 0`
- 発見元: BE10-2 B0のlint baseline。B0の変更前後で同一であり、BE10ではdrive-by修正を禁止しているため修正せず、本タスクへ移管した。

2026-07-23に起票したBUG-421〜428、TEST-ROUTES-01、FMT-BE-01は2026-07-24のBE9実装でsource/testへ反映済みのため、本active listから削除した。release pending項目（fresh DB migration、remote CI/coverage、production deploy/ops rehearsal）は実装taskではないため本書へ再掲しない。
