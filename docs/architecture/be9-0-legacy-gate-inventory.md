# BE9-0: 旧規約(P1–P18)の実効面 inventory

> ⚠️ **Historical inventory (2026-07-19 snapshot)**。
> 本表の path（当時の `internal/repository/*`・`internal/service/*`・`internal/handler/*` 等）は **現行 tree の enforcing path ではない**。
> **現行の live mechanical lint gate は `backend/internal/lintscan/` に置く。** 本表を現行 enforcing path として引用しないこと。
>
> 対象: 旧BE-refactor.md BE9-0（2026-07-24退役・経緯はgit履歴）。
> 実行日: 2026-07-19。scan コマンド: `rg -n --pcre2 '\bP(1[0-8]|[1-9])(\.[0-9]+)?\b' backend .github scripts .claude/scripts`
> before: 185 件 / after: 148 件（37件を意味化改名・除去）。
> **本doc が BE9-0 inventory の正本**（BE-refactor.md へのインライン複製はしない — 二重管理禁止）。live lint の正本 path は `backend/internal/lintscan/`。

## 0. 結論（Success Criteria 対応）

- **「旧番号だけを根拠に fail する active gate」は改修前・改修後とも 0 件。**
  185件全ヒットを対象に `Contains(...)` / `== "P` / `case "P` / `regexp` 等、P<n> 文字列を
  条件分岐・比較に使うコードパターンを機械探索したが 1 件もヒットしなかった
  （`rg 'Contains\(|==\s*"P|case "P|strings\.HasPrefix.*P[0-9]|regexp.*P[0-9]' backend --glob '*.go'` after
  the base P<n> grep — 0 hits）。P1–P18 は現行コードのどこにも pass/fail 条件として現れず、
  すべてコメント・エラーメッセージ・テストケース名内の**装飾的ラベル**に過ぎない。
- 上記を踏まえ、実際の作業は (a) inventory 化、(b) enforcing gate 8+1ファイルのうち
  P<n> ラベルを含む 4 ファイルの意味的改名（コメント/メッセージのみ、検出ロジック不変）、
  (c) 「形状のみ gate」の実地調査 → **実際に machine-enforced な shape-only gate は現行コードに
  存在しない**ことの確認（P13 定義順・P14 layering・P16-18 naming・P11 logging場所は
  いずれもテスト非対象。deleted `gin-architecture-compliance.md` の記述のみで、
  `backend/CLAUDE.md` が既に "P1–P18 は廃止された project 固有 checklist" と明記済み）、
  (d) 見つかった 6 件の代表的 production コードコメントの意味化改名、である。

## 1. 対象ファイル・分類方法

- 検索対象: `backend/**/*.go`, `backend/**/*_test.go`, `.github/**`, `scripts/**`, `.claude/scripts/**`
- 分類 4 種: `official Go/Gin` / `application safety invariant` / `project implementation detail` / `historical label`
- 別名前空間（P1–P18 アーキ準拠とは無関係な "P<n>" 表記）は上記 4 分類の対象外とし、
  「対象外」として個別に列挙する（BE9-0 スコープ外・triple-checked：本タスクでは一切変更していない）。

## 2. Reconciliation（ヒット総数突合）

| バケット | 件数 | 変更 |
|---|---:|---|
| A. Enforcing gate ファイル（意味化改名済み） | 37 | 改修（コメント/メッセージのみ、ロジック不変） |
| B. 対象外 — Infra/Cloudflare/Terraform 移行タスクID (`../ops/infra/_archive/migration-cloudflare.md` 系 P<n>) | 25 | 変更なし |
| C. 対象外 — PRレビューID（`PR #186 review` 系 P<n>） | 23 | 変更なし |
| D. 対象外 — CPM閾値ラベル（業務優先度、architecture P<n>と無関係） | 8 | 変更なし |
| E. 対象外 — TEST-FLAKEタグ | 2 | 変更なし |
| F. 対象外 — 業務データ値（処方コード `"P1"` 文字列リテラル） | 2 | 変更なし |
| G. 対象外 — 汎用優先度/Issue優先度/設計Wave/Terraformフェーズ等の無関係ラベル | 6 | 変更なし |
| H. 本 namespace の historical label（production comment、非gate） | 82 | 変更なし（維持・理由は§5） |
| **合計** | **185** | before=185 / after=148 |

37 (A) が意味化改名により消え、185 − 37 = 148 = after のヒット数と一致。
B〜H の内訳は `rg -c` によるファイル単位の機械カウントで実測済み（37+25+23+8+2+2+6+82=185）。

## 3. バケット A — Enforcing gate ファイル（意味化改名・検出ロジック不変）

対象 9 ファイル（Context指定の 8 + 追加発見 1）はすべて `go:embed` で自パッケージソースを
埋め込み、`go/ast` で機械的に walk する mechanical lint gate。**テスト関数名は元から
`TestXxx_Yyy` 形式で意味的**であり、P<n> は関数名・比較式のどこにも現れない
（`rg '^func Test' <file>` で確認済み）。P<n> はコメント・`t.Errorf` メッセージ文字列・
`reason` フィールド値の**装飾的ラベルとしてのみ**出現していた。

| ファイル | 分類 | 検出範囲(b) | P<n>ヒット数 | 判断(c) | 対応 |
|---|---|---|---:|---|---|
| `backend/internal/repository/preload_clinic_scope_lint_test.go` | application safety invariant | clinic-scoped master Preload に clinic_id 述語必須（cross-tenant read IDOR b3638d5e 回帰防止）。`go/ast` で Preload 呼び出しを AST 解析、good/bad fixture + self-verification で pin | 14 | 維持+改修 | "P3.1"/"P1" ラベルを "clinic-scope Preload rule" / "TODO follow-up" 等の意味的表現へ置換。t.Errorf メッセージも同様 |
| `backend/internal/service/master_fk_write_inventory_lint_test.go` | application safety invariant | request由来 clinic-scoped master FK を受け取る service method の allowlist 網羅性（write側review-coverage） | 7 | 維持+改修 | "P3.1"→"clinic-scope"、"残余P1リスク"→"残余high-priorityリスク" 等 |
| `backend/internal/repository/preload_master_model_reconciliation_test.go` | application safety invariant | read側 clinicScopedMasterAssoc と write側 clinicScopedMasterFKField の双方向整合性チェック | 5 | 維持+改修 | "P3.1 registry"→"clinic-scope registry" 等 |
| `backend/internal/repository/dbortx_inventory_lint_test.go` | application safety invariant | tx参加メソッドのinventoryゲート(ambient transaction) | 1 | 維持+改修 | "(P1 harness)"→"(ambient-transaction harness)" |
| `backend/internal/repository/audit_tx_inventory_lint_test.go` | application safety invariant | audit書き込みのtx境界inventory | 0 | 維持（変更不要） | P<n>参照なし（既にクリーン） |
| `backend/internal/repository/migration_cascade_lint_test.go` | application safety invariant | CASCADE DELETE禁止のmigration走査 | 0 | 維持（変更不要） | 同上 |
| `backend/internal/service/n1_lint_test.go` | application safety invariant | N+1クエリ検出 | 0 | 維持（変更不要） | 同上 |
| `backend/internal/model/audit_taxonomy_exhaustiveness_test.go` | application safety invariant | audit taxonomy網羅性（embed手法の先例） | 0 | 維持（変更不要） | 同上 |
| `backend/internal/handler/handler_routes_snapshot_test.go` | application safety invariant | route contract snapshot | 0 | 維持（変更不要） | 同上 |
| `backend/internal/repository/test_schema_enum_parity_test.go`（追加発見） | application safety invariant | DB enum とGoコードの整合性、embed+AST | 0 | 維持（変更不要） | 同上。Context記載の"8ファイル"には無かったが同種のmechanical gate。P<n>参照なしのため本タスクでの変更は不要 |

**代替test/参照名(d)**: 全ファイルとも `TestPreloadClinicScope_*` / `TestMasterFKWriteInventory_*` /
`TestMasterModelReconciliation_*` / `TestDBOrTxInventory_*` / `TestClinicalResultAuditTxInventory_*` /
`TestMigrationCascadeInventory_*` / `TestN1Lint_*` / `TestLabBlockedReason_Exhaustiveness_*` /
`TestRouteSnapshot` / `TestTestSchemaEnumParity` が既存の意味的名称であり、これを正本参照とする。

**挙動保存の証拠**: 改名は識別子/文字列リテラル/コメントのみ。`preloadHasClinicScope`、
`analyzeServicePackage`、`walkRepositoryPreloads` 等の判定関数本体、`clinicScopedMasterAssoc` /
`masterFKWriteAllowlist` 等のfixtureデータ（`occurrences`カウント、`predicate`文字列、
`status`値）は一切変更していない。scoped test 実行結果は §6 参照。

## 4. バケット B〜G — 対象外（別名前空間、triple-checked・本タスクでは変更していない）

| バケット | 代表ファイル:行 | 識別根拠 |
|---|---|---|
| B. Infra/Cloudflare/Terraform移行タスクID（2026-07-20以前の履歴） | `wrangler.jsonc`(10), `worker/index.ts`(5), `worker/migrate-exec.ts`(1), `.github/workflows/backend-deploy.yml`(5: L1,21,55,99,134), `backend-deploy-ecs.yml`(1), `stg-smoke.yml`(1), `staging-stop.yml`(1), `scripts/infra-terraform-plan-preflight.sh`(1) | `../ops/infra/_archive/migration-cloudflare.md` の Phase/タスクID。AWS廃止時に削除されたファイル名を含む当時の監査スナップショットで、現行の実行手順ではない |
| C. PRレビューID | `backend-deploy.yml`(2: L69,83), `config_validate_test.go`(1), `cmd/migrate/main.go`(2), `legacy_seed_keys_test.go`(1), `cmd/api/main.go`(1), `infra/s3_r2_live_test.go`(1), `infra/s3_endpoint_test.go`(1), `liff_response_test.go`(1), `liff_response.go`(1), `cross_tenant_master_fk_write_test.go`(3), `lab_import_examination_service_test.go`(4), `lab_import_examination_service.go`(2), `pet_response_test.go`(1), `cash_register_service_test.go`(2) | `PR #186 review` の指摘連番（P1-2, P1-3, P1-6, P2-2, P2-3, P2-5, P2-7, P2-13）。PRレビュー管理番号であり architecture ruleではない |
| D. CPM閾値ラベル | `model/clinic_settings.go`(3), `model/cpm_v2_thresholds.go`(1), `service/lstep_settings_service.go`(3), `service/lstep_tag_sync_service.go`(1) | "P1 CPM V2 来院回数閾値" 等、業務ドメインの優先度ラベル（CPM=Customer/Patient Metric）。migration 007-009 の閾値カラム名接頭辞であり、architecture P1–P18と無関係 |
| E. TEST-FLAKEタグ | `preload_followup_clinic_isolation_test.go`(1), `trimming_repository_test.go`(1) | `TEST-FLAKE-P2` = flakeチケット#236由来のtestタグ |
| F. 業務データ値 | `lstep_tag_code_mapping_repository_test.go`(2) | `pq.StringArray{"P1"}` は処方コード(CodeTypePrescription)のテストfixtureデータ値。文字列として偶然 "P1" と一致するのみ |
| G. 汎用優先度/Issue優先度/設計Wave/Terraform | `feature_request.md`(1: P0-P3 issue優先度), `check-design-primary-cta.mjs`(1: "Wave 0-7 P1ガード"), `audit_service.go`+`audit_service_test.go`(2: "PERF-AUDIT-TX P1"), `cmd/_archive/seed-old-db/main_test.go`(1: "harness P1"、archiveディレクトリ), `medical_record_repository_test.go`(1: "harness P1") | いずれも "P<n>" を汎用の優先度接尾辞として使うだけで、architecture P1–P18 の18項目チェックリストとは無関係の別名前空間 |

B+C+D+E+F+G = 25+23+8+2+2+6 = **66**（各ファイルを `rg -c` で個別カウントし合算、§2表と一致）。
H = 185 − 37(A) − 66 = **82**。

## 5. バケット H — 本 namespace の historical label（production comment、非gate・維持）

82件は production code（service/repository/handler/model）および doc
（`backend/CLAUDE.md`, `migrations/CLAUDE.md`, `tygo.yaml`, `001_init.sql`,
`openapi_date_format_drift_test.go` 等）に残る、旧 P1–P18 チェックリスト項目を指す
**コメント・ドキュメント参照のみ**。§0で示した通り、これらはいずれもpass/fail条件に一切関与しない
（コード上は単なる自然文コメント文字列）。

代表例と判断:

| 旧ラベル | 意味（deleted `gin-architecture-compliance.md` 由来） | 代表箇所 | 分類 | 判断(c) |
|---|---|---|---|---|
| P1 | FindByID before Delete/Update (Service) | `medicine_dose_param_service.go`(×3), `manual_article_service.go`, `medical_record_crud.go`, `shift_entry_service_test.go`, `lstep_lifecycle_service.go`(×4) | project implementation detail | 維持。事前存在確認という実装パターンの説明であり、pass/failを決めるgateではない |
| P2 | CountUsage with deleted_at IS NULL (Repository) | `billing_item_repository.go`, `accounting_repository.go`, 多数の`*_repository_test.go`のtestケース名内 | application safety invariant（ソフトデリート越境防止の実装意図） | 維持。実装済みロジックの説明コメント。gate化されていない |
| P4 | clinicScope on Update/Upsert (Repository) | `account/repository.go`, `account/repository_test.go`, `audit/repository.go`, `manualarticle/repository.go`, `medical_record_owner_visit_repository.go`, `trimming_repository.go`, `company/repository_test.go`, `lstep_csv_import_repository.go`, `lstep_csv_import_service.go`, `lab_import_repository.go`, `accounting_repository.go`(×2), `appointment_service_test.go` | application safety invariant | 維持。tenant write isolationの実装意図コメント。既に §3 の代表箇所（helpers.go / medicine_dose_param_repository.go）は意味化改名済み |
| P5 | RequirePermission on ALL non-public routes (Routes) | `manual_article_handler.go`("SEC-602: P5 RequirePermission(view) 付与"), `lab_import_handler.go` | application safety invariant | 維持。認可付与の説明コメント |
| P6 | DELETE routes use "delete" permission (Routes) | `pet_handler.go`(×2、死亡記録関連の例外説明) | application safety invariant | 維持 |
| P7 | toXxxResponse() conversion in handler (Handler) | `tygo.yaml`(×3), `lab_report_response.go`(×3), `lab_import_response.go`(×2), `staff_response.go` | project implementation detail | 維持。DTO変換パターンの説明。BE9新方針下でも domain package 内の response 変換として引き続き有効 |
| P8 | apperrors.Wrap in service (Service) | `aggregation_service_test.go` | project implementation detail | 維持 |
| P9 | apperrors.FromGORM in repository (Repository) | `clinic_settings.go`（実際はCPM=バケットDと誤認しやすいが、`P9 健診・予防タグ判定閾値`はCPM閾値でありP9アーキルールとは無関係。バケットDに算入済み） | — | （バケットD参照、二重計上なし） |
| P10 | FK dependency check before Delete (Service) | `migrations/CLAUDE.md`, `campaign_service.go`, `accounting_service_builders.go`, `update_fields.go` | application safety invariant | 維持。参照整合性チェックの実装意図コメント |
| P11 | slog.ErrorContext on repository error paths (Service) | `auth_handler.go`(改修済み・バケットA外だが§3外の追加改修), `checkup_field_result_service.go` | project implementation detail（logging場所=形状） | 維持（auth_handler.goのみ§3外で追加改修不要と判断・下記参照） |
| P13 | const/buildFunc definition order (Service) | `update_fields.go`("P10/P13") | project implementation detail（**形状gate候補**） | 維持（該当2ファイルは§3外で既に改修済み。update_fields.goの1件は複合ラベルのため見送り、次回touch時に整理） |
| P14 | Handler must not call Repository directly (Handler) | （§3外に残存なし。`handler.go`の2件は改修済み） | — | 改修済み（§3参照） |
| P18 | toXxxResponse function naming (Handler) | `lab_import_response.go`("Conversion functions (P18)") | project implementation detail（naming=形状） | 維持 |

**維持の理由（共通）**:
1. **pass/fail に無関係**: §0の機械探索で、P<n>文字列が条件式・比較・switch-caseの一部として
   使われている箇所は0件と確認済み。これらはすべて自然文コメント内の記述。
2. **`backend/CLAUDE.md` が既に"P1–P18 は廃止された project 固有 checklist であり、
   レビュー基準に使わない"と明記**しており、ドキュメントレベルでの非推奨化は完了している。
3. **product philosophy（②削除）とのバランス**: 82件全てを一括書き換えすると diff が
   本タスクの目的（gate の意味化）を大幅に超えて膨張し、"最小変更で checklist を満たす"
   という制約（Constraints）に反する。BE9-4（本タスクのスコープ外）が
   「同名の別project phaseやhistorical fixtureを除き0件、例外にはsemanticな説明を付ける」
   と明記しており、本inventoryがその"semanticな説明"を担う。
4. 次回そのファイルへ実装変更で touch する際に、意味的な表現へ置き換えることを推奨する
   （BE-refactor.mdのstrangler方針と整合）。

**独立レビュー（clinic-isolation-auditor + go-reviewer）で判明した follow-up（本タスクでは未修正・次回課題化）**:

`repository/CLAUDE.md`は既に P-番号を全廃したリライト済みdocだが（本docの§3・§6で前提とした通り）、
以下のファイルには**その旧`repository/CLAUDE.md`のP3.1節を名指しで参照するコメント**が残っており、
リンク先が実在しないダングリング参照になっている（go-reviewerのMEDIUM指摘、2026-07-19確認）:

- `backend/internal/repository/checkup_field_repository.go:67,87`
- `backend/internal/repository/accounting_repository.go:74,149,216,283`
- `backend/internal/repository/accounting_complete_appointments_test.go:216`（"repository/CLAUDE.md P3.1の「正本ガード=runtime isolation test」方針"という、存在しない節の引用）
- `backend/internal/repository/care_plan_item_repository_test.go:10,107`
- `backend/internal/model/checkup_field.go:78,82`
- `backend/internal/repository/lab_import_repository.go:55`（`helpers.go`で改名済みと同一invariantの重複表現）
- `backend/internal/repository/reservation_staff_repository.go:183`
- `backend/internal/service/accounting_service_builders.go:245`
- `backend/internal/service/update_fields.go:10`（P13。lstep_tag_*_service.goと同種だが本タスク未対象）

これらは§5バケットHの「維持」判断に含まれ pass/fail には無関係だが、単なる陳腐化ラベルではなく
**文書として誤った参照**である点が§5執筆時点の判断より一段重い。本タスクの
minimal-changes制約により見送るが、次のBE9フェーズ（BE9-1/BE9-2A）着手前、または
これらファイルへの次回touch時に、"repository/CLAUDE.md P3.1" 系の文言を
"the clinic-scope Preload rule"（本タスクで統一した表現）へ揃える一括follow-upを推奨する。

## 6. 「形状のみを強制する gate」の実地調査結果

BE9-0の指示「folder形状・定義順・logging場所・interface数だけを強制するgateは廃止または
根拠あるproject policyへ分離する」に対し、該当候補（P13定義順、P14 layering、
P16-18 naming、P11 logging場所）を機械探索した:

```
rg -ln 'definition order|const.*order|buildFunc|method name convention|struct naming|function naming|must not call.*[Rr]epository|Repository.*directly' backend --glob '*_test.go'
→ 0 hits
```

**結論**: これらを machine-enforced な test として強制している gate は現行コードに**存在しない**。
`gin-architecture-compliance.md` 自体が既に削除済み（git status で確認）、`.golangci.yml` にも
custom rule なし、`.github/PULL_REQUEST_TEMPLATE.md` にも P<n> チェック項目なし
（BE9-0プロンプトのContext記載は調査時点で既に解消済みだった — 古い前提。実測で上書き）。
したがって「廃止」すべき active gate は0件で、実施済みの対応は:
- `lstep_tag_code_mapping_service.go` / `lstep_tag_config_service.go` の "P13 definition order"
  コメント2件を、mandatory gateであるかのように誤読されないよう
  "non-mandatory convention（service/CLAUDE.mdの通りGo/Gin公式未規定）" と明記する形へ改修。
- `handler.go` の "P14:" ラベルを "consumer側narrow interface方針" という実質的な設計意図の
  説明へ改修（BE9のconsumer-side minimal interface方針と整合する形で保持）。

根拠あるproject policyへの分離は不要（分離する先の「gate」自体が実在しないため）。

## 7. 変更ファイル一覧（本タスクでの改修）

1. `backend/internal/repository/preload_clinic_scope_lint_test.go` — 14ラベル改名
2. `backend/internal/service/master_fk_write_inventory_lint_test.go` — 7ラベル改名
3. `backend/internal/repository/preload_master_model_reconciliation_test.go` — 5ラベル改名
4. `backend/internal/repository/dbortx_inventory_lint_test.go` — 1ラベル改名
5. `backend/internal/repository/helpers.go` — 3ラベル改名
6. `backend/internal/middleware/auth.go` — 2ラベル改名
7. `backend/internal/handler/handler.go` — 2ラベル改名
8. `backend/internal/service/lstep_tag_code_mapping_service.go` — 1ラベル改名（P13定義順コメント）
9. `backend/internal/service/lstep_tag_config_service.go` — 1ラベル改名（P13定義順コメント）
10. `backend/internal/repository/medicine_dose_param_repository.go` — 1ラベル改名
11. 本ファイル（新規作成）

全て識別子/文字列リテラル/コメントのみの変更。SQL・分岐・関数本体・fixture データ
（`occurrences`、`predicate`、`status`、good/badケース）は1バイトも変更していない。

## 8. 次ユニットへの入力

- 本inventoryは BE9-2A（boundary map / ADR-006）および BE9-1（safety lint の
  package非依存化）の入力となる。特に §3 の9ファイルは BE9-1 で `repoSourceFS` /
  `serviceSourceFS` の固定globをmodule全体scannerへ置換する対象そのものである。
- §6 で判明した「shape-only gate は実在しない」という事実は、BE9-4の
  `rg -n 'go-package-conventions|gin-architecture-compliance|golang-gin-clean-arch' .`
  ゼロ件チェックに対しても前向きな材料（追加の廃止作業が発生しない可能性が高い）。
