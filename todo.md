# AnimalEkarte — TODO

> 更新: 2026-07-24（BE9 implementation完了後reconciliation）

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
- 内容の正本 = q&a.html PO-006／DEC-9。USER が Issue 起票したら本エントリを Issue へ移設し二重掲出しない。

### TASK-251: 締め集計 category contract 確定実装（#251・8→12分類）

- 業務決裁確定（q&a.html DEC-21・**USER 本人裁定** 2026-07-25）。**着手時期 = 納品前（2026-07-26〜27）** — 2026-07-25 の納品日 7/27 延期（理由=残作業の全対応）に伴い S3 送りから前倒し。contract 正本 = DEC-21・#251 本文（本エントリは実装スコープと着手時期の入口であり決裁の「なぜ」は複製しない）。
- Phase 0 棚卸し（外部エージェント調査・Fable spot-verify）で確定した実装スコープ:
  - ① 正式カテゴリ = 12分類（enum 現状追認）。#251 タイトル「8分類」→「12分類」修正は Issue 本文転記（USER 承認後）に含める。
  - ② hospitalization 退院会計の other 固定を撤廃し CarePlanItem.Type／Procedure.IsSurgery→category resolver（`backend/internal/medicalrecord/hospitalization_service.go:431`）。treatment 経路（`backend/internal/billing/billing_item_service.go:405,462`）と共通化＝category contract 単一ソース化。
  - ③ vaccination を接種記録（`Vaccination.VaccineID`→Vaccine）から会計明細自動生成。`BillingItem` へ VaccineID provenance 列追加の migration が必要。自動化は停止／失敗通知／監査／idempotency（原則⑤）。
  - ④ hotel source=`HospitalizationTypeHotel`（②連動）、training は新規 source 設計。両カテゴリ維持。
  - 含意(a) category authority を BE resolver に一本化し FE/client は保持しない。
  - 含意(b) 締め集計の未知値 fail-closed = 生カラム無制限 GROUP BY（`backend/internal/billing/accounting_repository_reports_close.go:44`・`cash_register_service.go:265`）を12値 allowlist 経由にし typo/legacy を締め表へ黙って通さない（受け入れ条件「unknown/legacy を黙って変換しない」）。
  - 含意(d) 全書込経路（treatment/hospitalization/vaccination/trimming/merchandise/manual）を同一 typed category source に集約。
- #247（月次統合表）は本 TASK の contract 完了後に着手。
- Issue #251 本文への決裁転記（タイトル「8分類」→「12分類」修正含む）と着手時期の前倒し反映は、いずれも 2026-07-25 に USER 承認のうえ完了済み（live read-back で実測確認）。todo.md / q&a.html DEC-21 / #251 本文の3者は同期済み。以後 contract の参照先は #251 本文と DEC-21 とし、本エントリは着手時期と実装スコープのみを持つ。
- **U1 完了（`2154dc9de`・2026-07-25）**: category resolver を `backend/internal/sharedkernel/item_category_resolver.go` へ新設し、BE 導出の3経路（入院退院会計・外来診療明細・トリミング明細）を集約。②の other 固定撤廃と surgery 導出、④の hotel 到達経路を実装済み。`model.ItemCategory` 具体定数を参照する production ファイルは 4 → 3 に減少。resolver は error を返さない全域関数（退院処理の tx 閉包内にあるため、マスタ不整合で臨床業務を止めない）。
- **U1 で判明した scope の限界**: カテゴリ決定点は4つあり、うち `billing_item_service.go:258`（手入力・物販）は client 送出値をそのまま永続化する。含意(a)「category authority を BE resolver に一本化し FE/client は保持しない」の達成には API 契約変更が要るため U2 送り。
- **残ユニット**: U2 = 手入力・物販経路（下記のとおり U2a/U2b に再分解）。U3 = ③ ワクチン接種記録からの明細自動生成＋`billing_items` への VaccineID provenance 列 migration（停止手段・失敗通知・監査・idempotency 必須）。U4 = ④ training の新規 source 設計。U5 = 含意(b) 締め集計の allowlist 化。
- **U2 の再分解（2026-07-25 実測により当初定義「API 契約変更要」を訂正）**: API 契約は BE/OpenAPI 側に既に存在し、FE が使っていないだけである。
  - **U2a-1 完了（`33690944e`・2026-07-25）**: FE の物販マスタ選択経路で `merchandise_item_id` を送出（既存会計 POST と未作成会計の逐次作成の両経路）。`onAddItem` をオブジェクト引数化。`AccountingItem.merchandiseItemId` 追加。BE 側は 2 件是正 — ① `GetDiscountSuggestions` が `FindAllApplicableForItem` へ `nil` をハードコードしており、自動適用側（`resolveAutoDiscount`）と参照するキャンペーン集合が食い違っていた ② `BillingItemResponse` と OpenAPI `BillingItem` が `merchandise_item_id` を返しておらず、FE 生成型だけが当該フィールドを持つ状態（**BUG-433 と同型の乖離**）だった。②は生成側 prompt が「応答型の変更は不要」と誤って断定していたもので、外部エージェントが実測で反証した。検証 = frontend scoped vitest 41 passed/3 skipped、backend `ToBillingItemResponse|DiscountSuggestion` PASS、変更 4 ファイルの gofmt exit 0。**BUG-436 はこれで解消**。
  - **U2a-1 の残課題（次ユニットへ）**: `AddAccountingItemInput.category` および `FrontendMerchandiseItem.category` が `string` のままで、`use-accounting-item-actions.ts:45` の `category as ItemCategory` キャストが残る（本変更以前からの既存問題であり、悪化はしていない）。解消には `frontend/src/features/accounting/api/get-merchandise-items.ts` の型付けが要る。
  - **U2a = 物販経路の master 由来導出（着手可・納品前に収まる）**。`merchandise_items.category`（`item_category NOT NULL`）が正本であるにもかかわらず、`ItemListCard.tsx:104` が master の category を値コピーして送出しており二重管理になっている。かつ `merchandise_item_id` を送らないため master への link 自体が切れている（BUG-436）。修正 = FE payload へ `merchandise_item_id` を追加し、BE `CreateItem` は `MerchandiseItemID != nil` のとき master の Category で上書きする（作成時 derive・以後はスナップショット。マスタ改称で過去明細が遡及変化してはならない）。BUG-436 を同時に解消する。
  - **U2b = 純手入力経路（導出元が存在しない）**。provenance が無いため導出は原理的に不可能であり、含意(a) の「FE/client は category を保持しない」はこの経路では達成できない。正しい終端は「12値 validated 受理」であり、BE 側は `validators_billing_item.go:14`（12値検証）で**既に完了している**。残る欠陥は FE の `ItemListCard.tsx:120` が `"other"` 固定で category を送っている点のみ（利用者は選択できず、締め集計で手入力明細が全て other へ落ちる）。カテゴリ選択 UI を足すか "other" 固定を仕様として明文化するかは入力工程を増やす判断であり **USER 決裁**。
- **含意(b) の優先度を実測で再評価（要 USER 確認）**: `billing_items.category` は PostgreSQL enum `item_category`（`001_init.sql:108`・12値）の `NOT NULL` 列であり、typo/legacy 文字列は at rest で存在し得ない。(b) は現時点の typo 防御としては no-op。意味を持つのは「将来 enum に値を追加した際に締め表の表示リストが追随せず黙って落とす」ドリフト防御であり、納品前クリティカルではない。DEC-21 は USER 本人裁定のため、この降格の可否は USER 判断とする。
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

### BUG-432: 飼主生年月日がフォームから保存されない＋一覧列がpet値を表示

- HIGH。DB（`owners.birth_date`）・BE DTO・OpenAPI・DatePickerは実装済みだが、`frontend/src/features/owners/hooks/use-owner-form.ts` のcreate/update送信payloadにowner birthDateが含まれず、入力しても保存されない。さらに `OwnersListTable.tsx` の「生年月日」列はownerでなくpetの生年月日を表示している。
- 対応: payloadへbirth_date追加＋既存値を空へ戻す契約（JSON null vs 省略）の確定＋一覧列の正本（飼主DOB/ペットDOB）確定。#262の前提是正。
- 出典: #262調査 Completion Report（2026-07-25）。grepで整合確認済み。

### BUG-435: 生成FE型が陳腐化したままmainに乗っていた（codegen-check未励行）

- MEDIUM。FE12-07 で `make codegen` を回したところ、意図した型mapping修正（`any` 17→0）とは無関係な追随差分が同時に出た: audit定数7件（`AuditActionAuthPasswordChange` / `Reset` / `AdminReplace`、`AuditActionTrimmingCreate` / `Update` / `Delete`、`AuditResourceAccount` / `AuditResourceTrimming`）と `TokenBlacklist` のdoc comment。つまり **Go model を変更した際に `make codegen` が回されず、`frontend/src/types/generated/models.ts` が陳腐化した状態でmainに乗っていた**。
- `Makefile:349` に `codegen-check: codegen` + `git diff --exit-code frontend/src/types/generated/` が存在し、本来これがCIで検出する。検出されなかった理由（CIに未配線／配線済みだが未実行／実行され無視された）は未確認。
- 実害: FEが存在しない定数を参照しても型検査が通らないだけで安全側だが、逆に**BEが追加した定数をFEが使えない**状態が黙って続く。BUG-433（生成型がドメインモデル由来で応答DTOと不一致）とは別問題であり、そちらは構造の誤り、本件は同期の欠落。
- 対応: `make codegen-check` のCI配線状況を実測し、未配線なら追加する。配線済みなら失敗が無視された経路を塞ぐ。
- 出典: FE12-07 実行時に判明したNew Work（2026-07-25・`git diff frontend/src/types/generated/models.ts` 実測）。FE-refactor.md 範囲外のためこちらへ記載。

### BUG-436: 会計画面が `merchandise_item_id` を送らず、個別商品指定キャンペーンが一度も適用されない

- HIGH（サイレント機能不全）。BE は完全に配線済み: `backend/internal/billing/billing_item_request.go:52,85` が `merchandise_item_id` を受理し、`billing_item_service.go:281` の `ValidateCreateReferences` で所有権検証、OpenAPI（`backend/docs/api.yaml:5601`）にも定義がある。**FE の `CreateBillingItemRequest`（`frontend/src/features/accounting/api/create-billing-item.ts:6-19`）に当該フィールドが存在せず**、`use-accounting-item-actions.ts:65-75` も `create-accounting-items.ts:14-27` も送出しない。
- 実害: `campaign_repository.go:135-143` の `buildApplicableCond` は `merchandiseItemID != nil` のときだけ `campaign_target_items` を OR 条件へ加える。FE が常に nil のため、**個別商品指定キャンペーン（`campaign_target_items`）は会計画面からの明細追加で一度も発火しない**。カテゴリ指定キャンペーンのみが効いている。BUG-431（危険度バッジが実 API で点灯しなかった）と同型で、型検査もテストも通るため検出されない。
- 副次: `billing_items.merchandise_item_id` が常に NULL のため、物販明細から商品マスタへの provenance が失われている。`inventory/merchandise_item_repository.go:68` の `CountUsageByMerchandiseItemID`（マスタ削除ガード）も billing 側の利用を数えられない。
- 対応: TASK-251 U2a と同一の修正で解消する（FE payload へ `merchandise_item_id` 追加）。本エントリは「なぜ直ったか」を追跡可能にするための独立起票であり、実装は U2a に含める。
- **解消済み（`33690944e`・2026-07-25・TASK-251 U2a-1）**。FE の送出漏れに加え、`GetDiscountSuggestions` の nil ハードコードと応答 DTO の欠落という同根の欠陥 2 件も同時に是正した。
- 出典: TASK-251 U2 スコープ再評価時に判明（2026-07-25・FE/BE 双方のコード実測）。live データ（`campaign_target_items` の行有無）は未確認のため、露出の大きさは要実測。

### BUG-437: `ExamTypeField` に `ClinicID` を追加したが read 側 clinic-scope registry へ未登録（cross-tenant read IDOR の再発リスク）

- HIGH（clinic 分離）。作業中の変更 `backend/internal/model/examination_type.go` が `ExamTypeField` に `ClinicID uint64` を追加している（未コミット WIP。`backend/migrations/005_exam_reference_ranges_and_clinic_fk.sql` と `backend/internal/model/exam_reference_range.go` が同時進行）。write 側 registry には既に登録済み（`backend/internal/lintscan/master_fk_write_inventory_lint_test.go:129` の `"ExamTypeFieldID": "ExamTypeField (sub-master of ExaminationType, #124)"`）だが、**read 側の `clinicScopedMasterAssoc`（`backend/internal/lintscan/preload_clinic_scope_lint_test.go`）に entry が無く、`masterModelReadWriteExemptions` の例外にも入っていない**。
- 実害: clinic-scoped master を FK 値で Preload する際に `clinic_id` 述語が無いと、汚染された FK（#124/#125 の write 側検証ギャップ、または是正前データ）に対して**他院の master 名・値が返る**。b3638d5e で手作業修正した cross-tenant read IDOR と同型である。
- 検出経路: BE10-2 B5（`b28c4a105`）が `TestMasterModelReconciliation_RealSourceIsConsistent` を復旧した直後に検出した。**B4（`c430072d8`）以降このゲートは `no such file or directory` で死んでおり、B5 が直すまでこの追加は無検出で通過する状態だった。**
- 対応（3択・当該変更の owner が選ぶ）: ① Preload が既に存在するなら `clinicScopedMasterAssoc` へ association 名を登録する ② `internal/model` に association field が無いなら `masterModelReadWriteExemptions` へ理由付きで追加する ③ `ClinicID` の追加を撤回する。**ゲートを黙らせる目的だけの ② は禁止**。
- 再現: `docker compose exec backend go test ./internal/lintscan/ -run TestMasterModelReconciliation_RealSourceIsConsistent -count=1`（当該 WIP が worktree にある状態で FAIL する）。
- 出典: BE10-2 B5 の Mode 3 照合（2026-07-25・生成側が独立実測）。

2026-07-23に起票したBUG-421〜428、TEST-ROUTES-01、FMT-BE-01は2026-07-24のBE9実装でsource/testへ反映済みのため、本active listから削除した。release pending項目（fresh DB migration、remote CI/coverage、production deploy/ops rehearsal）は実装taskではないため本書へ再掲しない。
