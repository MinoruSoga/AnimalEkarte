# BE-refactor.md — バックエンド リファクタリング計画 (2026-07-10 時点・作業一時停止)

> **本ドキュメントは ClaudeCode エージェントが読んで直接着手するための実行計画である。**
> **2026-07-10 時点で Epic 作業を一時停止。** 対応済み 32 件の詳細は下記「進捗チェックポイント」に集約し、本編からは削除した。再開時は「推奨実行順」から着手すること。
> 前提: backend は 2026-07-02 完遂の D1-D13/R1-R3 計画で一度系統的にリファクタ済み・複数回の監査で well-maintained と判定済みのコードベースである。今回はそれ以降に実測ベースで発見された**実在する**負債のみを対象とする。

## 監査の方法と信頼性

- 対象: `backend/` 配下全体。テスト552本、カバレッジ baseline 89.9%（監査時点）。
- 手法: 15次元で並列監査 → 敵対的検証。全79件所見のうち **46件が挙動保存（本編）**、**18件が挙動変更（Appendix A）**。
- **進捗（2026-07-10）**: 本編 46件中 **32件 CLOSED**、**14件 残**（うち **2件 BLOCKED**）。Epic HEAD: `5722e22f`（G11-2〜G11-5 CLOSED）。
- **G11-4 レビューで新規発見（H-2）**: `UpdateExcludedReservationTypes` に H-1 と同型のクロステナントDELETE破壊バグ。本編・Appendix Aいずれにも未登録のため「レビュー由来フォローアップ」表に追記（別チケット化推奨、本Epicの実行対象外）。

## サマリー

| 区分 | 件数 | 内訳 |
|---|---|---|
| 本編（挙動保存）— **残タスク** | **14件** | P1: 1 / P2: 8 / P3: 5 |
| 本編 — **CLOSED（アーカイブ）** | 32件 | G3-1〜G3-5, G4-1〜3, G5-1〜2, G6-1, G7-1〜5, G8-1〜3, G9-2〜3, G10-1〜6, G11-1〜5 |
| Appendix A（挙動変更・別トラック） | 18件 | 変更なし |
| Appendix B/C/D | 参照用 | 変更なし |
| レビュー由来フォローアップ（未登録・別チケット推奨） | 2件 | H-1, H-2 |

## このドキュメントの使い方（ClaudeCode 向け）

1. **本編残タスクは全て挙動保存**。既存テストを事前に緑確認 → リファクタ/テスト追加 → 同じテストが緑のままであることを確認。検証はスコープ限定（`docker compose exec backend go test ./internal/xxx/ -run ... -count=1`）。フル `go test ./...` 禁止。
2. **worker/root pnpm 検証**: ホスト `Bash(pnpm:*)` は `.claude/settings.json` で deny。`docker run --rm -v "$(pwd):/repo" -w /repo node:22-bookworm-slim` 内で `corepack enable && corepack prepare pnpm@10.15.0 --activate && pnpm install && pnpm run test:worker`（G10-6 で実証済み）。
3. 1項目 = 1コミット。`.claude/settings.json` はステージしない。
4. **Appendix A は本計画の実行対象外**（PO判断・別トラック）。
5. 完了した項目は本ファイルから削除し、進捗チェックポイント表のみ更新する運用。

---

## 進捗チェックポイント（CLOSED — 2026-07-10）

| ID | 項目 | 最終コミット |
|---|---|---|
| G3-1 | lstep タグ同期 DRY | Phase 1–7 + H1a–H1d 完遂 |
| G3-2 | Run*AllClinics 統合 | `b89520c3` |
| G3-3 | updateScopedByID/deleteScopedByID | `061b79b6` 〜 `ceab1e31` |
| G3-4 | medicalRecordTenantScope | `edc315e7` 〜 `0f76b71e` |
| G3-5 | 日付パース共通化 | 全段完了 |
| G4-1 | lab_report P7 | `689e9506` |
| G4-2 | LIFF Create P7 | `a00bf673` |
| G4-3 | Handler narrow DI | `5e5754db` |
| G5-1 | P11 slog 11箇所 | `3eca7ef2` |
| G5-2 | P13 buildFunc 移動 | `5c7434aa` |
| G6-1 | P16 Get*/List* リネーム | （Session 2） |
| G7-1 | LIFF N+1 | `755f3e42` |
| G7-2 | 健診プレビュー N+1 | `f6ebc0c7` |
| G7-3 | completed_at sargable | `904b0bc2` |
| G7-4 | checkup/vaccination index | `dd3f31b8` |
| G7-5 | BulkAddOwnerTag FindByIDs | `93867fcc` |
| G8-1 | findByIDScoped | `0fe8fe78`, `d09cfa4c` |
| G8-2 | config.JST | `ecaac5ec` |
| G8-3 | time.DateOnly/RFC3339 | `2d5c180d`, `9bae5e8a`, `bcc34d5d` |
| G9-2 | env → Config | `c903ab3b` |
| G9-3 | batch scheduler dedup | `72d64ae7` |
| G10-1 | migrate binary untrack | Session 3 |
| G10-2 | Dockerfile 削除 | Session 3 |
| G10-3 | CI profiling 削除 | Session 3 |
| G10-4 | codemod 削除 | Session 3 |
| G10-5 | internal/dbconn | Session 3 |
| G10-6 | migrate-exec worker tests | `dc4f2ed8` 実装, `a1429497` CLOSED |
| G11-1 | effective permissions SQL tests | `d8c5e858` |
| G11-2 | CompleteAccountingAppointments SQL tests | `dd76a7f1`, 強化 `0c4134f7` |
| G11-3 | FindUnbilledTrimmingItemsByPetID/CountNonAccountingTrimming tests | `9375b264` |
| G11-4 | UpdateReservationCapabilities isolation test | `696ab204` |
| G11-5 | Lstep購買クエリ3本 tests + completed_atコメント修正 | `3ffb60ec`, `5722e22f` |

---

## 推奨実行順（残タスク — 再開時）

```
G12-1 → G12-2
→ G13-1 → G14-1
→ G1-4 → G1-1 → G1-3 → G1-2(Phase A→C) → G1-5 → G1-6 → DOC-G1
→ G2-1 → G2-2 → DOC-G2
→ G6-2（BLOCKED 解消後）
→ G9-1（BLOCKED 解消後）
→ Final gate
```

**G11 フェーズは 2026-07-10 Session 6 で完遂・CLOSED**（G11-2〜G11-5、下記チェックポイント参照）。

### BLOCKED（Appendix A 依存 — 本編着手不可）

| 本編 ID | 依存 Appendix A ID | 理由 |
|---|---|---|
| G6-2 | tx-medicine-inventory / tx-clinic-create / tx-reservation-staff | tx 非参加3件修正後に一括置換 |
| G9-1 | lstep-nilcipher-stale-di | 二段階 DI 統合の前提 |

### レビュー由来フォローアップ（本編未登録）

| ID | 内容 | 発見元 | 優先度 |
|---|---|---|---|
| H-1 | `UpdateStaffGroups` の staff_id 単位 DELETE が多施設所属スタッフの他クリニックグループ紐付けを意図せず削除しうる | G11-1 security-reviewer | HIGH — 別チケット化推奨 |
| H-2 | `UpdateExcludedReservationTypes`（reservation_staff_repository.go）の DELETE が `staff_id` のみでスコープされ `clinic_id` を含まない一方、INSERT は呼び出しクリニックの型IDのみ。`staff_reservation_exclusions` テーブル自体に `clinic_id` 列が無いため、多施設所属スタッフに対しては clinic A の正当な操作が clinic B の除外設定行を無警告で全削除する（H-1 と同型のクロステナント破壊）。兄弟の `UpdateReservationCapabilities`/`staff_reservation_capabilities` は自前 `clinic_id` 列を持ち `Where("clinic_id = ? AND staff_id = ?")` で正しくスコープされており非対称。 | G11-4 security-reviewer（`UpdateReservationCapabilities` との比較監査で発見） | HIGH — 別チケット化推奨（`staff_reservation_exclusions` への `clinic_id` 列追加 or DELETE を真の差分更新へ変更、要 migration） |

---

# 残タスク（本編）

## G1. 契約・ドキュメント整合性 (docs/api.yaml 他)

### G1-1. api.yaml に実装に存在しない/廃止済みエンドポイント定義が23オペレーション残存(うち1件は誤配置による OpenAPI 構造違反)

- **ID**: `api-yaml-phantom-endpoints`
- **重要度**: P1 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: docs/api.yaml (6455-6476, 11342-11361, 13454-13620, 13620-13680, 13906-13990); internal/handler/handler.go (284-285)
- **依存関係**: なし(単独実行可)

**証拠(現HEAD検証済み)**

docs/api.yaml:6455-6457 は `delete:\n      operationId: deleteAccounting\n      summary: 会計削除` を定義するが、internal/handler/handler.go:284-285 は `// BUG-371 / #118: DELETE は廃止し論理削除 (POST /:id/cancel) に統合。専用権限 accounting-cancel を使用する。` `accountings.POST("/:id/cancel", h.RequirePermission(string(model.ResourceAccountingCancel), "edit"), h.CancelAccounting)` — DELETE ルートは登録されていない。docs/api.yaml:13906 `  /staffs/reservation-exclusions:` と 13620 `  /accounting/billing-refunds:` は Go ソース全体 grep で 0 ヒット(該当ハンドラ・サービス・モデル一切なし)の純粋 phantom。13537 `  /line/customers:` は実装では reservation_line_routes.go:57-59 `customers := clinics.Group("/line-customers")` (= /clinics/{clinic_id}/line-customers、GET+PATCH のみ)であり、文書化された /line/customers の POST/PUT/DELETE は存在しない。さらに docs/api.yaml:11342-11347 では `deletePermissionGroup` が `/masters/permission-groups/reorder` パス配下に誤配置され、path template に無い `- name: id\n          in: path\n          required: true` を宣言(OpenAPI 仕様違反。実装 permission_group_handler.go:24 は `masters.DELETE("/permission-groups/:id", ...)`)。

**問題**

info.description が「Implementation Synced v3.1.0」と主張する契約文書に、意図的に廃止された危険操作(会計 DELETE=BUG-371 で論理削除へ統合)や全く存在しないリソースファミリーが記載されており、FE/PO/外部連携がこれを信じて設計すると実装と衝突する。deletePermissionGroup の誤配置は OpenAPI バリデータすら通らない構造バグで、api.yaml が機械検証なしに手動編集されている証左。

**実装手順**

全て docs/api.yaml の削除・移動のみ(実装変更なし)。手順: (1) /accountings/{id} の delete: ブロック(6455-6476)を削除し、代わりに実装済みの POST /accountings/{id}/cancel・POST /accountings/{id}/credit-correction を追記(finding api-yaml-missing-operations と同時でも可)。(2) 11342-11361 の delete: ブロックを `/masters/permission-groups/{id}` パス配下へ移動。(3) /staffs/reservation-exclusions と /staffs/reservation-exclusions/{id} の全オペレーション、/accounting/billing-refunds と /accounting/billing-refunds/{id} の全オペレーションを削除(billing-refunds は実装済みの /accountings/{id}/refunds が既に 6477 に文書化済みで重複兼 phantom)。(4) /line/customers・/line/customers/{id}・/line/reservation-settings・/line/reservation-settings/{id} を削除し、実装形 /clinics/{clinic_id}/line-customers (GET, PATCH {customerId}/link-owner)・/clinics/{clinic_id}/line-reservation-settings (GET, PUT) へ置換。(5) LineCustomer/LineReservationSetting/BillingRefund スキーマ自体は実装モデルが存在するため残す。注意: format: date プロパティを増減させると internal/apicontract の date-format gate の床値/allowlist に影響しうるため、編集後にスコープ付きテストで確認する。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/apicontract/ -count=1
```

### G1-2. 実装済み477オペレーション中203件(43%)が api.yaml に未記載 — payment-methods/closing-settings/cash-register/manual/lstep/LIFF 等のリソースファミリー丸ごと欠落

- **ID**: `api-yaml-missing-operations`
- **重要度**: P2 / **工数目安**: L / **挙動変更**: なし（挙動保存）
- **対象ファイル**: docs/api.yaml (5318-13988); internal/handler/payment_method_master_handler.go (130-138); internal/handler/handler.go (48-137)
- **依存関係**: apicontract-route-drift-gate を先に入れると段階実施が安全になる(必須ではない)

**証拠(現HEAD検証済み)**

静的ルート解決(internal/handler 全 Register*Routes の Group/verb リテラルを AST 追跡、未解決 0 件)で実装 477 オペレーション vs api.yaml paths 297 オペレーション。実装のみ 203 件(パラメータ名差異 3 件除外後)。例: internal/handler/payment_method_master_handler.go:130-137 `pm := rg.Group("/payment-methods")` 配下の GET/POST/PATCH /reorder/GET :id/PATCH :id/DELETE :id 全 6 オペレーションが grep `"  /payment-methods"` 0 ヒットで丸ごと未記載。同様に /closing-settings(8)・/cash-register(4)・/manual/articles(5)・/billing-items(6)・/shared-files(5)・/lab-imports・/lab-reports・lstep 系(40+)・/api/liff/{clinicId} 系(13)・/medical-records/{id}/prescriptions|addenda|checkups 系・/pets/{id}/chronic-conditions|death|first-visit|treatment-history が欠落。docs/README.md:13 は「手動で管理するOpenAPI 3.0仕様ファイル。エンドポイントの追加・変更時は合わせて更新する」と規定するがプロセスが破綻している。

**問題**

api.yaml info.description(docs/api.yaml:5)が「Implementation Synced v3.1.0」を掲げつつカバレッジが 57% しかなく、契約文書としての信頼性が失われている。date-format drift gate(internal/apicontract)は api.yaml に書かれたものしか検査できないため、未記載の 43% は既存の契約検査の射程外になっている(検査基盤の実効性を蝕む)。

**実装手順**

挙動保存のドキュメント追補。203 件を一括で書くのは非現実的なのでリソースファミリー単位で段階化する: Phase A(会計系=臨床・金銭に直結): /accountings サブアクション(cancel/credit-correction/unpaid/unpaid-balance/unpaid-monthly/daily-summary)、/payment-methods 全 6、/cash-register 4、/closing-settings 8、/billing-items 6、/reports/monthly 2。Phase B(カルテ系): /medical-records/{id}/prescriptions・addenda・checkups/field-results、/pets/{id}/chronic-conditions・death・first-visit・treatment-history、/examinations/{id}/items、/masters/* の GET {id} 系。Phase C(連携系): lstep 全般、/api/liff/{clinicId}(servers が /api/v1 固定のため liff は別 servers エントリまたは絶対パスで記載)、/manual、/shared-files、/lab-imports、/lab-reports。各 Phase で既存スキーマ($ref)を再利用し、レスポンス形は対応する *_response.go の json タグから起こす。apicontract-route-drift-gate finding のゲートを先に導入して現状 203 件を allowlist に pin すれば、Phase 間の逆行(新規未記載追加)を CI で防げる。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/apicontract/ -count=1
```

### G1-3. レスポンススキーマのフィールド・enum drift — Payment.method に bank_transfer 欠落(#127)、payment_method_id/paid_by 欠落(#128)、Billing に total_refunded_amount 等欠落・非直列化 deleted_at を記載

- **ID**: `api-yaml-schema-field-enum-drift`
- **重要度**: P2 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: docs/api.yaml (161-209, 874-964); internal/model/accounting.go (29-34, 62-92, 141-167)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

docs/api.yaml:200-202 `method:\n          type: string\n          enum: [cash, credit_card, electronic_money]` に対し internal/model/accounting.go:33 `PaymentMethodCash ... PaymentMethodBankTransfer    PaymentMethod = "bank_transfer" // #127: 銀行振込` — enum に bank_transfer が無い。model/accounting.go:159 `PaymentMethodID *uint64 \`... json:"payment_method_id,omitempty"\`` と :160 `PaidBy *uint64 ... json:"paid_by"` は Payment スキーマ(161-209)に無い。Billing 側: model/accounting.go:82 `TotalRefundedAmount int64 \`gorm:"-" json:"total_refunded_amount"\`` と :91 `PaymentSplits []PaymentSplit ... json:"payment_splits,omitempty"` と :87 `MedicalRecord *MedicalRecord ... json:"medical_record,omitempty"` が Billing スキーマ(874-964)に無い。逆に api.yaml:934-938 は `deleted_at:\n          type: string\n          format: date-time\n          nullable: true` を記載するが model/accounting.go:77 は `DeletedAt gorm.DeletedAt \`... json:"-"\`` で wire に一切出ない phantom フィールド。

**問題**

会計ドメイン(金銭)の wire 契約が文書と一致しない。enum 欠落は FE が bank_transfer 支払いを未知値として扱う設計ミスを誘発しうる。payment_method_id は Method との dual-maintain 運用(model/accounting.go:154-158 のコメントで PO 判断 B として長期維持が明記)された正式フィールドであり、文書に無いのは契約情報の欠損。deleted_at の phantom 記載は「文書にあるフィールドが来ない」という逆方向 drift で消費側の混乱源。

**実装手順**

docs/api.yaml のみ修正(実装変更なし)。(1) Payment schema(161-209): method enum に bank_transfer を追加、payment_method_id (integer, format: int64, nullable) と paid_by (integer, format: int64, nullable) と paid_by_staff ($ref StaffSummary 相当, readOnly, nullable) を追加。(2) Billing schema(874-964): total_refunded_amount (integer, readOnly, description: 返金累計・FindAll サブクエリ集計) と payment_splits (array, readOnly) と medical_record (nullable, Preload 時のみ) を追加し、deleted_at プロパティ(934-938)を削除。(3) 同型の drift が他スキーマにもある可能性が高いので、会計系を直したあと Owner/Pet/MedicalRecord スキーマを *_response.go・model の json タグと突合して同一 PR 内で棚卸しする(date 系フィールドの format は既存 allowlist に pin 済みのため型を変えないこと)。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/apicontract/ -count=1
```

### G1-4. internal/apicontract を date-format 以外へ拡張: ルートインベントリ drift gate(実装ルート ↔ api.yaml paths の突合)の新設

- **ID**: `apicontract-route-inventory-gate`
- **重要度**: P2 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/apicontract/doc.go (1-9); internal/apicontract/openapi_date_format_drift_test.go (45-77, 217-245)
- **依存関係**: なし(先行導入すると api-yaml-missing-operations の段階実施が安全になる)

**証拠(現HEAD検証済み)**

internal/apicontract/doc.go:2-5 「Package apicontract holds API-contract invariant checks that cross-reference the hand-maintained OpenAPI spec (docs/api.yaml) against the Go handler layer. It intentionally lives in its own package ... so the checks read handler *source* via os.ReadFile + go/ast」— 枠組みは複数の invariant を前提に設計済みだが、現存する検査は date-format 1 次元のみ。今回の監査で全 477 ルートが internal/handler 内の Group()/GET()/POST()/... の静的文字列リテラルだけで 100% 解決できること(未解決 0 件・handler 外のルート登録 0 件)を実証済みで、AST ベースのルート列挙は既存 openapi_date_format_drift_test.go:192-215 の walkHandlerResponseDrifts と同じ构造で実装可能。allowlist+床値+stale 検出のパターンは同ファイル 56-75(knownDateFormatDrifts)と 219-245(reconcileDateFormatDrift)で確立済み。

**問題**

api.yaml と実装の乖離(203 未記載 + 23 phantom)は手動メンテ(docs/README.md:13)が破綻した結果であり、一度修正しても再発を防ぐ機構が無い。既存 gate の設計(allowlist に現状を pin して新規混入と stale を fail させる)がこのクラスにそのまま適用できる。

**実装手順**

新規ファイル internal/apicontract/openapi_route_drift_test.go を追加。(1) ルート列挙: internal/handler の非テスト .go を go/parser で読み、*Handler メソッド内の `X := Y.Group("lit")` 代入と `X.GET("lit", ...)` 系呼び出し、`h.RegisterXxx(v)` の呼び出しグラフを RegisterRoutes から辿ってフルパスを解決(handler.go:50 起点、liff は r.Group("/api/liff/:clinicId") 起点)。:param は {param} へ正規化。(2) api.yaml 側: 既存の yaml.v3 パースを流用し paths 直下の (method, path) を servers prefix /api/v1 で絶対化して収集。(3) 差分を knownRouteDrifts allowlist(方向別: missingFromSpec / phantomInSpec)と reconcile し、新規・解消・件数変化を fail。パラメータ名差異は shape 正規化({} 化)で吸収するか別 allowlist にする。(4) 床値ガード: 実装ルート数 < 400 または yaml オペレーション数 < 250 で fatal(空振り防止)。(5) 初期 allowlist は導入時点の実測(現 HEAD なら missing 203 / phantom 23)を pin し、finding api-yaml-phantom-endpoints / api-yaml-missing-operations の各 Phase 完了ごとにエントリを削る。オプション: kin-openapi の Loader+Validate を smoke test として追加すれば deletePermissionGroup 誤配置のような構造違反も CI で検出できる(依存追加が必要なため別判断で可)。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/apicontract/ -run TestOpenAPIRouteDrift -count=1
```

### G1-5. docs/postman-collection.json と docs/api-examples.md がプロトタイプ期(2026-01-26)の遺物 — UUID ID・APIキー認証・存在しないエンドポイント・誤った HTTP verb を案内

- **ID**: `stale-prototype-example-docs`
- **重要度**: P2 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: docs/postman-collection.json (8-40, 149-170, 356-381); docs/api-examples.md (7-13, 68-78, 114-119, 231-236); docs/README.md (5-9)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

docs/postman-collection.json:29-31 `"key": "petId",\n      "value": "87a6e7a4-8b70-44b7-b7b5-e0b7e3b4e2bd"` — 現行 API の ID は uint64(model/accounting.go:63 等)で UUID ではない。同 8-14 の auth は `"type": "apikey" ... "value": "YOUR_API_KEY"` だが現行認証は HttpOnly Cookie + Bearer JWT(api.yaml:8-19、auth_handler.go:19-21)。api-examples.md:117 `curl -X GET "http://localhost:8080/api/v1/medical-records/paginated?..."` — `paginated` は internal/handler 全 grep 0 ヒットの存在しないエンドポイント。api-examples.md:71 `curl -X PUT "http://localhost:8080/api/v1/pets/{pet_id}"` — 実装は pet_handler.go:149 `pets.PATCH("/:id", ...)` で PUT ルートは無い。api-examples.md:236 「UUID形式のIDを使用」・235「APIキーは必須」はいずれも虚偽。両ファイルは git log 上 2026-01-26(1fd36aff)以降一度も更新されず、Makefile/.github/backend 内のどこからも参照されず、docs/README.md:5-9 のファイル構成一覧(README.md と api.yaml のみ)にも載っていない。

**問題**

リポジトリ内に「実行すると全て失敗する」使用例文書が正本のふりをして残っており、新規参加者や外部連携先が拾うと確実に誤導される。メンテ対象としても docs/README.md が正本一覧から既に除外している=誰も維持する意思がない。

**実装手順**

削除が最小コスト・挙動保存: docs/postman-collection.json と docs/api-examples.md を git rm する(参照 0 件を確認済みのため安全)。Postman コレクションが将来必要になれば api.yaml から openapi-to-postman で自動生成できるため手書き資産を残す価値は無い。docs/README.md のファイル構成(5-9 行)は既に 2 ファイル構成として書かれているので変更不要。ドキュメント削除のみのためランタイム検証は不要。

**検証コマンド(スコープ限定)**
```
ls backend/docs/ (README.md と api.yaml のみになることを確認)
```

### G1-6. backend/README.md の CRUD 追加手順が現行必須規約(P5/P7/RespondError/SQL migration)に違反するコードを教示、README/CODING_RULES のディレクトリ構成が実態(apicontract/infra/seedbundle・cmd 7 サブディレクトリ)と乖離

- **ID**: `onboarding-docs-teach-forbidden-patterns`
- **重要度**: P3 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: README.md (16-55, 100-110, 116-244); CODING_RULES.md (14-67)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

backend/README.md:218-225 の推奨ハンドラ例 `c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})` と `c.JSON(http.StatusOK, owners)`(モデル直返し)は、internal/handler/CLAUDE.md の P7「❌ Direct model or gin.H」・P12「❌ c.JSON(http.StatusBadRequest, gin.H{...})」で明示禁止されたパターンそのもの。README.md:234-235 `v1.GET("/owners", h.GetOwners)` は RequirePermission 無しで、P5「全ルートに付与必須 (AUDIT-H2 2026-05-09)」に違反する登録例。README.md:132 `ID uuid.UUID \`json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"\`` — 実モデルは全て uint64 autoIncrement(model/accounting.go:63)。README.md:243 `db.AutoMigrate(&model.Pet{}, &model.Owner{})` を手順として指示するが、実運用は migrations/ SQL 管理(001_init.sql)で AutoMigrate 追加は手順として誤り。README.md:44-45 は存在しない `internal/validation/` を記載し、実在する apicontract/infra/seedbundle/middleware 配下の実態と不一致。README.md:109 `| PUT | /api/v1/pets/:id |` — 実装は PATCH(pet_handler.go:149)。CODING_RULES.md:16-17 は cmd/ 配下を api/ のみと記載するが実際は api/coverage-ratchet/lstep-migrate/migrate/seed-export/seed-old-db/stage-import の 7 つ、CODING_RULES.md:20-53 の internal/ 一覧に apicontract/infra/seedbundle が無い。

**問題**

オンボーディング文書がレビューゲート(P1-P18)と正反対のコードを「追加手順」として教えており、新規実装者・外部エージェントがこれに従うと必ずレビューで差し戻される。ディレクトリ構成の欠落は「apicontract/seedbundle は正式パッケージではない」という誤認を生む。規約正本(.claude/CLAUDE.md・各層 CLAUDE.md)は正しいため実害は差し戻しコストに留まるが、文書間矛盾として残す理由が無い。

**実装手順**

挙動保存のドキュメント修正。(1) backend/README.md:116-244 の「新しいCRUD機能の追加手順」を丸ごと削除し、CODING_RULES.md と internal/*/CLAUDE.md への参照 1 段落に置換(パターンの二重管理を解消 — 教材コードの重複保守は PRODUCT_PHILOSOPHY の二重管理禁止と同型)。(2) README.md:16-55 と CODING_RULES.md:14-67 のディレクトリツリーを実態に同期: internal/ に apicontract(API契約検査)/infra/seedbundle を追記、internal/validation を削除、cmd/ 7 サブディレクトリを列挙、migrations/ の例を 001_init.sql 形式へ。(3) README.md:100-110 のエンドポイント表は PUT→PATCH に直すか「docs/api.yaml 参照」1 行に置換。(4) README.md:8 の Gin v1.10 を go.mod 実測 v1.12.0 へ。ドキュメントのみのためランタイム検証不要。

**検証コマンド(スコープ限定)**
```
検証不要(ドキュメントのみ)。grep -n 'internal/validation' backend/README.md が 0 件になること
```


## G2. 大型ファイルの責務分割

### G2-1. treatment_service.go(668行・service層最大): 保存時dose再検証サブシステムが CRUD と同居 — treatment_dose_save.go へ抽出

- **ID**: `td-treatment-dose-save-extract`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/service/treatment_service.go (513-645)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

internal/service/treatment_service.go:513-524:
// ─── #201 B-2: 保存時 dose 再検証 ──────────────────────────────────────────────

// evaluateDoseForSave は medicine 明細の保存時 BE 再検証を行う（healthcare HIGH-3/HIGH-4）。
...
func (s *treatmentService) evaluateDoseForSave(
同 587行 func (s *treatmentService) resolveDoseWeight(...)、618行 func (s *treatmentService) auditDoseDeviationTx(...)。dose 純ロジックは既に dose_calc.go / dose_revalidation.go / dose_validators.go に分離済みで（doseSnapshotColumns 等は dose_revalidation.go:166-199）、テストも treatment_dose_save_test.go:2 「treatment_dose_save_test.go — #201 B-2: treatment 保存時 BE 再検証の統合テスト。」として独立ファイル化済み。実装側だけが treatment_service.go に残っている。

**問題**

treatment CRUD オーケストレーションと #201 投薬量安全サブシステム（species解決→体重解決→再検証→逸脱audit）という2責務の同居。チーム自身がテストを treatment_dose_save_test.go として別ユニット扱いしており、実装ファイルの構成だけが追随していない。668行は service 層非テスト最大で、dose ブロック約130行を出すと CRUD 部が~535行の単一責務になる。

**実装手順**

同一パッケージ内の純粋なファイル移動（挙動保存・シグネチャ不変）。手順: 1) internal/service/treatment_dose_save.go を新設（既存テストファイル名と対に）。2) treatment_service.go の 513-645行（セクションコメント「─── #201 B-2: 保存時 dose 再検証 ───」・evaluateDoseForSave・resolveDoseWeight・auditDoseDeviationTx）を逐語移動。メソッドレシーバ *treatmentService のまま。3) import 調整: fmt は resolveDoseWeight のみが使うため新ファイルへ、maps/strconv 等は残留側で使用継続を確認して整理。4) テストは同一パッケージのため変更不要。依存の向き: 新ファイル→dose_revalidation.go/dose_calc.go（既存と同じ）、treatment_service.go→新ファイル（同一パッケージ内メソッド呼び出し）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestTreatmentService' -count=1
```

### G2-2. medical_record_repository.go(490行): 臨床CRUD と Lステップ/CPM向け飼主来院集計クエリ群の2消費者ドメイン同居 — ファイル分割候補

- **ID**: `td-medrec-repo-owner-visit-split`
- **重要度**: P3 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/medical_record_repository.go (14-26, 74-88, 338-490)
- **依存関係**: なし（インターフェース合成まで行う場合も呼び出し元変更ゼロ）

**証拠(現HEAD検証済み)**

internal/repository/medical_record_repository.go:74-80:
	// FindOwnerVisitSummary は飼い主の初回/最終診療日・年間来院数を集計して返す（Lステップ同期用）。
	FindOwnerVisitSummary(ctx context.Context, clinicID, ownerID uint64) (*OwnerVisitSummary, error)
	// FindLatestByOwner は飼い主の最新カルテを返す（Lステップ次回来院推奨日タグ同期用）。
...
	// FindDormantOwnerEntries は最終来院から minDaysSince 日以上経過した飼い主一覧を返す（バッチ処理用）。
— 実装 338-490行（FindLatestByOwner/FindOwnerVisitSummary/FindOwnersByFirstVisitDate/FindOwnersByLastVisitDays/FindOwnersByNextVisitRecommended/FindDormantOwnerEntries）の呼び出し元は lstep_batch_segmentation.go / lstep_tag_sync_visit*.go / lstep_batch_dormant.go / lstep_delivery_trigger_methods.go 等ほぼ全て lstep_* サービス。リポジトリ自身に前例あり: reservation_repository.go:20-75 は ReservationCRUDRepository/ReservationSlotRepository/ReservationQueryRepository の3サブインターフェース合成で消費者別に整理済み。

**問題**

臨床カルテ CRUD（FindAll/FindByID/Create/Update/Delete/Count系）とマーケティング系バッチ集計（Lステップタグ同期・FEAT-383 リマインド・休眠検知）という独立した消費者ドメインが1ファイル・1インターフェースに同居。reservation_repository.go が確立したサブインターフェース合成イディオムと非対称。

**実装手順**

挙動保存のファイル分割（最小案）: 1) internal/repository/medical_record_owner_visit_repository.go を新設し、型 OwnerVisitSummary/DormantOwnerEntry（14-26行）と実装 FindLatestByOwner/FindOwnerVisitSummary/FindOwnersByFirstVisitDate/FindOwnersByLastVisitDays/FindOwnersByNextVisitRecommended/FindDormantOwnerEntries（338-490行）を逐語移動（レシーバ *medicalRecordRepository のまま）。FindOwnersByNextVisitRecommended の M-5 ガードコメント（433-434行「リファクタ時に clinic_id WHERE のいずれか一方を削除しないこと」）と clinic_id 二重指定 Raw SQL は一字も変えず移す。2) 任意の追加ステップ: reservation_repository.go:20-75 の前例に倣い MedicalRecordCRUDRepository + MedicalRecordOwnerVisitRepository をサブインターフェース化して MedicalRecordRepository で合成（呼び出し元シグネチャ不変）。3) テストは medical_record_repository_test.go に残置で可（同一パッケージ）。任意で該当 Test 関数を新 _test.go へ機械移動。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestMedicalRecordRepository' -count=1
```


## G6. Repository層規約・トランザクション機構整理

### G6-2. tx 参加機構が実質5系統併存し、ctx 方式は repo 側 dbOrTx 未採用時に静かに非参加となる — 機構の棚卸しと標準化・ガード整備

- **ID**: `tx-mechanism-consolidation`
- **重要度**: P2 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/transactor.go (11-42); internal/repository/base.go (27-35); internal/repository/repositories.go (236-251); internal/repository/audit_repository.go (14-47); internal/repository/helpers.go (12-54); internal/repository/CLAUDE.md (1-183)
- **依存関係**: tx-medicine-inventory-nonparticipation / tx-clinic-create-nonparticipation / tx-reservation-staff-nonparticipation の修正後に (2) の一括置換を実施（同一ファイル競合回避）

**証拠(現HEAD検証済み)**

機構の実測棚卸し: ①ctx-txKey 方式 = transactor.go:28-31「func (t *gormTransactor) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {\n    if err := t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {\n        return fn(context.WithValue(ctx, txKey{}, tx))」+ base.go:30-35 dbOrTx。dbOrTx 採用 repo は 104 非テストファイル中 18 のみ（grep 実測）。service 側 WithTx 呼出は 27 箇所/14 service。②repo-swap 方式 = repositories.go:243-245「if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {\n    txRepos := NewRepositories(tx)\n    return fn(txRepos)」（treatment_service.go:242/377・hospitalization_service.go:281 が使用。TransactionFn テストフック付き）。③repo 内部独立 tx = r.db.WithContext(ctx).Transaction が campaign:92/clinic:103/lstep_tag_cache:244/manual_article:60/owner:173/permission_group:102,218,243/reservation_schedule:91/reservation_staff:68,130,226,295/reservation_type_liff:104/shift_entry:122/shift_template:93/treatment:187 の13ファイル。④SAVEPOINT 参加型内部 tx = dbOrTx(ctx, r.db).Transaction が accounting:298/checkup_field:110/examination:180/staff:143/trimming:85（accounting_repository.go:288-289 に R1-1 根拠コメント）。⑤補助 = audit_repository.go:31/42 の Create/CreateTx 二本立て・helpers.go:14,36 の db 引数+内部 tx。

**問題**

①の ctx 方式は「repo が dbOrTx を使っていること」が暗黙の前提だが、これを強制する仕組みが無く、違反してもコンパイル・テスト・lint のどれも落ちない（silent non-participation）。実際に accounting(R1-1/D2 で修正済み)→今回 medicine/inventory・clinic/permission_group・reservation_staff と同一障害クラスが繰り返し発生しており、構造的な再発様式が確立してしまっている。また①と②が service 層で無方針に併存し、新規 service 実装時にどちらを使うべきか・repo 側に何が要求されるかが CLAUDE.md に明文化されていない（repository/CLAUDE.md は P2/P3/P4/P9/P16 のみで tx 規約の記載ゼロ）。

**実装手順**

挙動保存トラック: (1) internal/repository/CLAUDE.md に tx 規約セクションを追加 —「Transactor.WithTx 配下で呼ばれる repo メソッドは dbOrTx(ctx, r.db) 必須 / repo 内部 tx は dbOrTx(ctx, r.db).Transaction (R1-1 パターン) を標準 / 新規 repo は全メソッド dbOrTx で書く」を accounting_repository.go:288-289 の先例参照付きで明記。(2) ③の残 13 ファイルの r.db.WithContext(ctx).Transaction を dbOrTx(ctx, r.db).Transaction へ機械置換（ambient tx 呼出が現存しないことは本監査で確認済みのため挙動保存。reservation_staff は tx-reservation-staff-nonparticipation で先行）。(3) ①②の統合は行わない — ②は全 repo が自動参加する点で機能的に堅牢であり、treatment/hospitalization の複雑な多 repo tx に適合している。拙速な統一は YAGNI。ただし「新規はどちらを選ぶか」の判断基準（単一〜少数 repo なら①、多 repo 横断なら②）を CLAUDE.md に記す。(4) 静的 lint（service WithTx ブロック内呼出の taint 追跡）は master_fk_write_inventory_lint_test.go 冒頭で静的追跡を断念した同じ理由（呼出し越しのデータフロー解析が go/ast 単体では偽陰性/偽陽性多発）で作らない。正本ガードは F1-F3 で追加する rollback runtime テスト。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -count=1
```


## G9. 周辺パッケージ・DI組立の整理

### G9-1. main.goの二段階DI（NewServices後に約20サービスを再構築/追加配線）を単一段階に統合

- **ID**: `two-phase-di-consolidation`
- **重要度**: P2 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: cmd/api/main.go (86-167); internal/service/service.go (190-193,206,264)
- **依存関係**: lstep-nilcipher-stale-di を先に完了させること

**証拠(現HEAD検証済み)**

cmd/api/main.go:158-162:
	// FEAT-383: 自動配信トリガー（LstepBatch / MedicalRecord / Checkup より先に初期化）
	svcs.LstepDeliveryTrigger = service.NewLstepDeliveryTriggerService(repos.Owner, repos.MedicalRecord, ...)
	// FEAT-383: イベントフック注入（LstepDeliveryTrigger 確定後に再初期化）
	svcs.MedicalRecord = service.NewMedicalRecordService(repos.MedicalRecord, repos.Owner, repos.Pet, repos.Inquiry, repos.ClinicalPlan, repos.LineCustomerMgr, repos.Reservation, svcs.LstepDeliveryTrigger, svcs.Audit, svcs.LstepTagSync)
	svcs.Checkup = service.NewCheckupService(repos.Checkup, repos.MedicalRecord, repos.CheckupType, svcs.LstepDeliveryTrigger, svcs.LstepTagSync)

internal/service/service.go:206: MedicalRecord: NewMedicalRecordService(..., repos.Reservation, nil, auditSvc, lstepTagSyncSvc),
service.go:264: Checkup: NewCheckupService(repos.Checkup, repos.MedicalRecord, repos.CheckupType, nil, lstepTagSyncSvc),
（NewServices が nil トリガーで構築 → main.go が捨てて再構築する二重構築）

**問題**

MedicalRecord/Checkup/LstepSettings/LstepTagSync/LstepLifecycle/LstepTag が NewServices と main.go で二重構築され、「LstepDeliveryTrigger 確定後に再初期化」等の順序制約コメントが main.go に散在する。この構造が finding lstep-nilcipher-stale-di（部分上書きによる stale 参照）の温床であり、今後 lstep 依存サービスを追加するたびに同型の取りこぼしが再発し得る。main.go の配線ブロック（86-167行）のうち infra 依存が必要なのは SharedFile/LineSend（FileStorage 経由）のみで、残りは repos とサービス相互参照だけで構築可能。

**実装手順**

前提: lstep-nilcipher-stale-di を先に修正。手順: (1) service.NewServices のシグネチャに sharedStorage infra.FileStorage を追加（main.go は STORAGE_TYPE 分岐で構築した storage を渡す。import 方向は service→infra で既存 crypto import と同じため循環なし）。(2) main.go:95-167 で構築している LstepTriggerPriority → LstepDeliveryTrigger → MedicalRecord/Checkup → LstepBatch → SharedFile/ChronicCondition/LineSend/LineLink/LstepTagSummary/CheckupSync/LstepDeliveryMonitor/LstepCsvImport/LstepAnalytics を NewServices 内へ依存順に移設し、service.go:206/264 の nil トリガー仮構築を排除して一回構築にする。(3) main.go は logger/config/DB/cipher/storage/uploader の infra 構築と handler.New 呼び出しのみ残す。(4) 既存 service 層テスト全体で回帰確認。挙動保存（構築結果のグラフは修正後と同一）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/... -count=1
```


## G12. model/DDL・テストスキーマ整合

### G12-1. schema_drift_test の allModels() が 107 モデル中 72 のみ登録（AuditLog/PaymentSplit 等 35 欠落）・NOT NULL 非検査・カスタム型は isEnumLike で素通り

- **ID**: `drift-test-coverage-gap`
- **重要度**: P2 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/model/schema_drift_test.go (37-119); internal/model/schema_drift_test.go (313-343); internal/model/schema_drift_test.go (365-375)
- **依存関係**: 手順(2)は audit-ip-inet-model-drift の解消と相互依存（先に検出網を入れると CI が正しく RED になる）

**証拠(現HEAD検証済み)**

schema_drift_test.go:37-39 `// allModels は全GORMモデルを列挙する。\n// 新しいモデル追加時はここに追記すること。` — 手動保守で強制なし。TableName() 実装は 107 型（grep 実測）に対し allModels() は 72 型。欠落35型: AuditLog, PaymentSplit, CheckupFieldResult, CheckupTypeField, MedicalRecordAddendum, MedicineDoseParam, PetChronicCondition, PasswordResetToken, TokenBlacklist, LineLinkToken, LineSendLog, SharedFile, TrimmingCourseType, ManualArticle, ManualArticleVersion, Campaign, CampaignTargetCategory, CampaignTargetItem, ClinicIntegration, LabImportJob, LabImportEvent, ReservationTypeAvailableSlot, ReservationTypeOccupation, ReservationTypeUnavailableTime, Lstep系12型。型比較も :324-327 `// ENUM型はtext/enum両方あり得るので、Goが string ベースなら許容\nif isEnumLike(goCategory) || isEnumLike(dbCategory) {\n    continue\n}` により inet/uuid 等の PG 組込型 vs text の不一致を許容し、NOT NULL/デフォルト値は一切比較しない（:206 コメント通り列存在+型カテゴリのみ）。CI では ci.yml の Schema drift check ステップが 001_init.sql 適用済み ekarte_db に対して実行されるが、audit_logs の inet 乖離（audit-ip-inet-model-drift）は上記2つの盲点（未登録+isEnumLike）の両方に該当し検出不能。

**問題**

モデル↔DDL整合の唯一の機械ゲートに系統的な穴が3つある: (1) 登録漏れ35型（監査・会計 payment_splits・健診結果を含む）、(2) pointer フィールド vs NOT NULL 列という実行時 NULL 制約違反クラスを未検査、(3) isEnumLike が PG 組込非デフォルト型（inet 等）の乖離を握り潰す。実際に audit_logs の乖離がこの3盲点の陰に隠れた。手動列挙はこのリポジトリが他所で確立済みの exhaustiveness lint パターン（audit_taxonomy_exhaustiveness_test.go / master_fk_write_inventory_lint_test.go の go:embed+go/ast 双方向突合）に反する。

**実装手順**

挙動保存（テスト基盤のみ）。手順: (1) internal/model/schema_drift_test.go に go/ast ベースの exhaustiveness 検査を追加: internal/model/*.go を走査し TableName() string を実装する全 struct を列挙、allModels() の型集合と双方向突合（欠落/余剰で fail）。lstep_migration_progress のようにモデルが存在しないDBテーブルは対象外のまま。(2) allModels() に欠落35型を追記（この時点で AuditLog の乖離が顕在化するので、audit-ip-inet-model-drift の修正とセットで green 化する）。(3) 型比較強化: pgTypeCategory に 'inet'→'inet' 等の PG 組込型を明示カテゴリ化し、isEnumLike の許容集合から外す。(4) nullability 検査を追加: migrator.ColumnTypes の Nullable() と Go フィールドの pointer 性を比較し、「Go=pointer かつ DB=NOT NULL(デフォルト無し)」を drift として報告（逆方向 DB=nullable かつ Go=非pointer は read 時 NULL scan エラークラスなので警告リストに）。誤検知が出た列は根拠コメント付き allowlist で pin する。影響範囲: schema_drift_test.go 1ファイル+allModels 追記。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/model/ -run 'TestSchemaDrift|TestAllModelsExhaustive' -v
```

### G12-2. setupSharedTestSchema の手書き ENUM 複製が 001_init.sql から乖離 — item_source に 'trimming' が無く、トリミング会計明細の永続化経路が統合テスト不能

- **ID**: `test-enum-copy-drift`
- **重要度**: P2 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/ltv_repository_test.go (1385-1478); migrations/001_init.sql (108); internal/repository/billing_item_repository.go (258-278)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

001_init.sql:108 `CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization', 'trimming');` に対し、ltv_repository_test.go:1433 の手書き複製は `{"item_source", "CREATE TYPE item_source AS ENUM ('medical_record', 'manual', 'hospitalization')"}` で 'trimming' が欠落。production コードは billing_item_repository.go:271 `Source:                model.ItemSourceTrimming,` で当該値を組み立て、会計作成時に billing_items.source='trimming' として永続化される。さらに 001 の CREATE TYPE 54 型に対しテスト側は 50 型で、campaign_discount_type / checkup_field_type / lab_import_job_status / lab_import_source_type の4型が欠落（comm 実測。checkup_field_type は real-DDL 専用ヘルパーが自前作成するため実害は item_source が最大）。ltv_repository_test.go:1389-1390 のコメント `//（001_init.sql の 46 型 + 009 #201 薬量計算の 4 型）` も現状54型と不一致。なお :1456-1462 の DO ブロックは IF NOT EXISTS ガードのため、テスト側文字列を直しても既存 ekarte_db_test には反映されない点に注意。

**問題**

テストDB（ekarte_db_test）のスキーマ正本が 001_init.sql ではなく手書き複製+AutoMigrate で、既に値レベルの乖離が実在する。source='trimming' を INSERT する統合テストは書いた瞬間に enum 違反で失敗するため、トリミング→会計連携の永続化経路（#73/#77）は repository 統合テストで担保不能。将来 001 に enum 値を追加するたびに同じサイレント乖離が再発する構造。

**実装手順**

挙動保存（テスト基盤のみ）。手順: (1) ltv_repository_test.go の enumTypes 文字列を 001_init.sql と一致させる（item_source に 'trimming' 追加、campaign_discount_type / lab_import_job_status / lab_import_source_type を追加。checkup_field_type は real-DDL trio の DROP+CREATE と衝突しないことを checkup_field_repository_test.go 側で確認の上追加判断）。(2) 再発防止 lint を新設: internal/repository/test_schema_enum_parity_test.go — migration_cascade_lint_test.go と同じ相対パス読取で 001_init.sql の CREATE TYPE 定義（複数行対応で正規化）を抽出し、enumTypes の定義文字列と完全一致比較（欠落・値差分で fail。テスト専用型があれば根拠コメント付き allowlist）。(3) IF NOT EXISTS ガード対策: setupSharedTestSchema に「既存型の値集合が定義と不一致なら DROP TYPE ... CASCADE→再作成」を追加するか、最低限 README/コメントで ekarte_db_test の DROP DATABASE 再作成を案内。(4) :1389 の型数コメントを実数に修正。影響範囲: internal/repository のテストヘルパーのみ。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestTestSchemaEnumParity|TestBillingItemRepository' -count=1
```


## G13. サイレント障害の解消

### G13-1. lab import 補償トランザクション(failed 遷移)の失敗が無ログで破棄され、job が非終端状態で stuck しても観測不能

- **ID**: `lab-import-compensation-unlogged`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/service/lab_result_import_service.go (124-134)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

internal/service/lab_result_import_service.go:125-133:
	if err != nil {
		// context キャンセル等のシステムエラー: job を failed に遷移させてから返す。
		// ctx は既にキャンセル済みのため、補償トランザクションには新しいコンテキストを使う。
		errMsg := err.Error()
		cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cleanupCancel()
		_, _ = s.jobSvc.TransitionStatus(cleanupCtx, clinicID, jobID, model.LabImportJobStatusFailed,
			TransitionCounts{RowCount: len(inputs), ErrorCode: ptr("context_cancelled"), ErrorMessage: &errMsg})
		return nil, apperrors.Wrap(err, "lab import batch interrupted")

コメントは cleanupCtx 採用理由のみで、エラー破棄の正当化コメント・nolint は無い。対照的に同ファイルの終端遷移(172-177行)は失敗を slog.ErrorContext で記録し「永続化は完了しているため、エラーはログのみで返さない。」と根拠コメント付き:
	if _, err := s.jobSvc.TransitionStatus(ctx, clinicID, jobID, termStatus, finalCounts); err != nil {
		slog.ErrorContext(ctx, "lab result import: failed to transition to terminal status", ...

**問題**

補償遷移(→failed)自体が失敗した場合(DB 断・5s タイムアウト超過)、job は mapped 等の非終端状態で恒久 stuck するが、その事実がどこにも記録されない。主エラーは呼び出し元へ返るため利用者はリトライできるが、stuck job の存在は運用側から完全に不可視。同一ファイル内で終端遷移だけログする非対称は、過去監査(F-1/LIFF-1/2)で是正してきた P11 観測性方針とも不整合。

**実装手順**

挙動保存のログ追加のみ。internal/service/lab_result_import_service.go:131 を以下に変更:
	if _, compErr := s.jobSvc.TransitionStatus(cleanupCtx, clinicID, jobID, model.LabImportJobStatusFailed,
		TransitionCounts{RowCount: len(inputs), ErrorCode: ptr("context_cancelled"), ErrorMessage: &errMsg}); compErr != nil {
		slog.ErrorContext(cleanupCtx, "lab result import: failed to transition to failed (compensation)",
			"error", compErr, "job_id", jobID)
		// 主エラーを優先して返す(挙動不変)。job は非終端で残るため jobID から追跡する。
	}
lab_result_import_service_test.go に PersistBatch 失敗+TransitionStatus(failed) 失敗を注入する mock ケースを追加し、戻り値が従来どおり "lab import batch interrupted" の wrap であることを固定する。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestLabResultImportService_Commit' -count=1
```


## G14. マスタFK write allowlistドキュメント精度

### G14-1. masterFKWriteAllowlist の accountingService.Update エントリが実装と乖離（既にguarded相当）

- **ID**: `accounting-update-allowlist-stale-status`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/service/master_fk_write_inventory_lint_test.go (191); internal/service/accounting_service_builders.go (21-43); internal/service/accounting_service_core.go (118-141)

**証拠(現HEAD検証済み)**

allowlist(191行目): `{"accountingService.Update", statusKnownUnguarded, []string{"PaymentMethodID"}, "...not a FindByID guard and no isolation test — verify rejection of explicit foreign IDs."}`。しかし accounting_service_builders.go:37-40 は `if existing != nil && *existing != id { return nil, apperrors.WrapInvalidInput(...) }` で、clinic-scoped な systemKeyToID（accounting_service_core.go:118 `s.loadPaymentMethodSystemKeyToID(ctx, input.ClinicID)`）に解決した id と request の明示 PaymentMethodID が不一致なら拒否しており、定義文（lint冒頭コメント）の『guarded = ownership validation (FindByID(...) or equivalent) covers ALL master FKs』の “or equivalent” に該当する挙動を既に持つ。

**問題**

allowlist の status は『レビュー俎上に乗ったか』の記録であり、この記録自体が実態（既にguard相当のロジックが存在）を過小評価している。将来の監査者がこの記録を信じて『まだ未対応』と誤認し、実際により緊急度の高い他の known-unguarded 項目（billing_item/campaign 等）への注意が薄まるリスクがある。gate自体はcorrectnessを保証しないため、記録の正確性が唯一の防波堤。

**実装手順**

resolvePaymentMethodMasterID の相互整合チェックを検証する isolation test（別クリニックの payment_method_id を明示指定した Update が 400 で拒否されることを確認）を internal/service/cross_tenant_master_fk_write_test.go に追加し、GREEN化後に allowlist 191行目の status を `statusGuarded` に更新して reason を『resolvePaymentMethodMasterID の mismatch 拒否ロジックで validated; test: TestAccountingService_Update_RejectsForeignPaymentMethodID』に書き換える。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestMasterFKWriteInventory -v
```
## Appendix A: 挙動変更を伴う項目（別トラック・PO/責任者判断を要する）

以下18件は監査で実在を確認した defect だが、修正すると HTTP レスポンス・DB書込結果・権限判定・API契約のいずれかが観測可能な形で変わる。このため本計画（挙動保存リファクタ）の実行対象からは外し、個別 Issue として起票のうえ別トラックで扱うことを推奨する。severity 順に記載。

**特に優先度が高い2件（P1・データ破損/資格情報系）**:
- `X-1 sanitize-multipart-binary-corruption`: カルテ画像・共有ファイルアップロードのバイナリが保存時に破壊される可能性（2026-03-31 導入以来のクラス）。
- `X-2 lstep-nilcipher-stale-di`: 本番で6サービスが復号不能な nil cipher 経路を参照し続けている DI 配線バグ。

**マルチテナント書込ガード欠落（P1・#124/#125 と同型）**:
- `X-4 billing-item-trimming-fk-unguarded` / `X-5 campaign-target-item-fk-unguarded`

**監査ログ整合性（P1）**:
- `X-3 audit-ip-inet-model-drift`: 本番DDL(inet列)とmodel(string)の型不一致により、複数の監査ログ書込経路が実行時に失敗しうる。
### X-1. SanitizeNullBytesがmultipart/form-dataを含む全POST/PUT/PATCHボディにbytes.Mapを適用しバイナリアップロードを破壊する

- **ID**: `sanitize-multipart-binary-corruption`
- **重要度**: P1 / **工数目安**: S
- **対象ファイル**: internal/middleware/sanitize_null_bytes.go (25-70); cmd/api/main.go (286-287); internal/handler/medical_record_image_handler.go (124)

**証拠(現HEAD検証済み)**

internal/middleware/sanitize_null_bytes.go:26-31:
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPatch && method != http.MethodPut {
			c.Next()
			return
		}
（Content-Type による除外は一切ない）

sanitize_null_bytes.go:47-58:
		sanitized := bytes.Map(func(r rune) rune {
			switch {
			case r == 0x00: // NULL バイト
				return -1
			...
			case r >= 0x0E && r <= 0x1F: // SO〜US
				return -1

cmd/api/main.go:286-287:
	// BUG-067: POST/PATCH/PUT ボディから NULL バイトを除去（PostgreSQL エラー防止）
	r.Use(middleware.SanitizeNullBytes())

internal/handler/medical_record_image_handler.go:124: file, fileHeader, err := c.Request.FormFile("file")

frontend/src/features/medical-records/api/medical-record-images.ts:17-22 は new FormData() + axios.post(multipart/form-data) で生Fileを送信する。

**問題**

bytes.Map はボディを UTF-8 コードポイント列として解釈する。バイナリ（JPEG の 0xFF 0xD8、PNG シグネチャ等）は不正 UTF-8 として各バイトが U+FFFD（3バイト EF BF BD）に置換され、さらに PNG シグネチャの 0x1A は除去レンジ（0x0E-0x1F）に該当してドロップされる。カルテ画像アップロード（POST /medical-records/:id/images/upload）、共有ファイル（shared_file_handler.go:48）、Lステップ CSV インポート（Shift-JIS の場合）が全てこのグローバル middleware（main.go:287、ルート登録前に Use）を通過するため、保存されるバイナリは原理的に原本と一致しない。臨床記録画像の破壊 = データ損失リスク。middleware は 2026-03-31（0234bf8a）、画像ハンドラは 2026-04-11 以降に追加されており、導入以来この経路にバイナリ忠実性のテストは存在しない（sanitize_null_bytes_test.go は application/json のみ）。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

手順: (1) RED test 先行 — internal/middleware/sanitize_null_bytes_test.go に Content-Type: multipart/form-data で PNG シグネチャ [0x89,0x50,0x4E,0x47,0x0D,0x0A,0x1A,0x0A] を含むボディを通し、バイト完全一致を検証するケースを追加（現実装で FAIL することを確認 = 破壊の実証）。(2) SanitizeNullBytes 冒頭に Content-Type ガードを追加: strings.HasPrefix(ct, "application/json") の場合のみサニタイズし、それ以外（multipart/form-data, application/octet-stream 等）は素通しする（BUG-067 の対象は JSON テキストのため意図保存）。(3) 既存 JSON ケースのテストが GREEN のままであることを確認。(4) 実機確認として staging で画像アップロード→ダウンロードのバイト一致を手動検証（破壊が確認された場合、既存アップロード済み画像の破損調査は別トラック起票）。影響範囲: multipart を使う3エンドポイントのアップロード忠実性。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/middleware/ -run TestSanitizeNullBytes -count=1
```

### X-2. NewServicesがnil cipherでLstepSettingsを構築し、main.goの上書きが6サービスに波及しない（本番で復号不能な資格情報を使用）

- **ID**: `lstep-nilcipher-stale-di`
- **重要度**: P1 / **工数目安**: S
- **対象ファイル**: internal/service/service.go (190-192,203-204,209,245,287-288); cmd/api/main.go (95-96); internal/service/lstep_settings_credentials.go (19-25); internal/service/lstep_tag_sync_service.go (281-302)

**証拠(現HEAD検証済み)**

internal/service/service.go:190-192:
	// LSTEP services initialization with nil cipher (production code in main.go will override with encrypted cipher)
	lstepSettingsSvc := NewLstepSettingsService(repos.LstepSettings, repos.LstepSyncSettings, nil, auditSvc, repos.ClinicSettings)
	lstepTagSyncSvc := NewLstepTagSyncService(lstepSettingsSvc, repos.Owner, ...)

service.go:203: Owner: NewOwnerService(repos.Owner, lstepTagSyncSvc, auditSvc),
service.go:204: Pet: NewPetService(repos.Pet, repos.Owner, repos.Insurance, repos.MedicalRecord, lstepTagSyncSvc),
service.go:209: Accounting: NewAccountingService(repos.Accounting, lstepTagSyncSvc, tx, auditTxLogger, repos.PaymentMethodMaster),
service.go:245: Vaccination: NewVaccinationService(repos.Vaccination, repos.Vaccine, lstepTagSyncSvc),
service.go:287: Prescription: NewPrescriptionService(repos.Prescription, repos.MedicalRecord, lstepTagSyncSvc),
service.go:288: Aggregation: NewAggregationService(repos.Ltv, repos.LstepTagCache, repos.LstepTagConfig, lstepSettingsSvc),

cmd/api/main.go:95: svcs.LstepSettings = service.NewLstepSettingsService(repos.LstepSettings, repos.LstepSyncSettings, integrationCipher, svcs.Audit, repos.ClinicSettings)

internal/service/lstep_settings_credentials.go:20-22:
func (s *lstepSettingsService) decrypt(keyName, value string) (string, error) {
	if s.cipher == nil || !model.IsEncryptedKey(keyName) {
		return value, nil

lstep_tag_sync_service.go:294/301: apiKey, baseURL, _, err := s.settingsSvc.GetRawCredentials(ctx, clinicID) ... return lstep.NewClient(apiKey, baseURL), nil

**問題**

main.goの再配線はServicesコンテナのフィールドを差し替えるだけで、NewServices内でローカル変数 lstepTagSyncSvc / lstepSettingsSvc をコンストラクタ注入済みの Owner/Pet/Accounting/Vaccination/Prescription/Aggregation の6サービスには波及しない。これらは本番でも cipher=nil の LstepSettingsService を参照し続け、decrypt が暗号文をそのまま返すため、Owner更新等を起点とするタグ同期は暗号文をAPIキーとして lstep.NewClient に渡す。現在は Lステップ write API 一時停止（AddTag/RemoveTag noop）でマスクされているが、GetUserTags/GetUser 経路と write 再開時に顕在化する。開発環境は cipher が全経路 nil のためテストで検出不能。service.go:190 のコメント自体が「main.goが上書きする」という誤った前提を記録している。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

手順: (1) internal/service/service.go:191 の第3引数 nil を NewServices が既に受け取っている cipher に変更（同関数は line 279 で LineReservationSetting に同じ cipher を渡しており非対称が解消される）。(2) これにより cmd/api/main.go:95-96 の svcs.LstepSettings / svcs.LstepTagSync 再構築は引数が完全同一となり冗長化するため削除し、main.go:112-127 の LstepLifecycle/LstepTag 再構築も削除（finding two-phase-di-consolidation と同時実施可）。(3) RED test: internal/service に NewServices へ非nil cipherを渡した時 Services.LstepSettings 経由の decrypt が復号することを確認するテストを追加し、修正前に Owner 経路の settingsSvc が nil cipher であることを実証する。影響範囲: lstep タグ同期の全起点サービス（本番のみ挙動変化）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestNewServices|TestLstepSettings' -count=1
```

### X-3. audit_logs.ip_address が DDL では inet、model では string — 空文字書込は実DDLで INSERT 失敗し、AutoMigrate製テストスキーマがそれを隠蔽

- **ID**: `audit-ip-inet-model-drift`
- **重要度**: P1 / **工数目安**: M
- **対象ファイル**: internal/model/audit_log.go (9-27); migrations/001_init.sql (2543-2563); internal/service/audit_service.go (171-296); internal/service/examination_service.go (341-364); internal/repository/audit_repository_test.go (22-29)
- **依存関係**: drift-test-nullability-gap と同根（あちらは検出網、こちらは実体修正）。同時に着手する場合は本件の real-DDL テストを先に入れる

**証拠(現HEAD検証済み)**

DDL (migrations/001_init.sql:2545,2553): `clinic_id    bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,` / `ip_address   inet         NULL,`。model (internal/model/audit_log.go:11,22): `ClinicID   *uint64         \`json:"clinic_id"\`` / `IPAddress string          \`json:"user_agent"\`の直上、`IPAddress string          \`json:"ip_address"\``（gormタグ無し=INSERTに常に含まれる）。IPAddress を設定しない書込経路が多数: audit_service.go:184-193 LogLstepOperationWithMetadata / :209-219 LogMedicalRecordChange / :236-247 LogVitalChange / :285-295 LogAddendumCreate、および fail-closed tx 監査 examination_service.go:345-359 の LogEntryTx（AuditLogInput に IPAddress 未設定）。PostgreSQL では `''::inet` は 22P02 invalid input syntax であり、空文字パラメータの inet 列への INSERT は失敗する。一方テストは audit_repository_test.go:25 `require.NoError(t, db.AutoMigrate(&model.AuditLog{}))` で string→text 列のスキーマを自前生成するため通過し、同 :56 `t.Run("clinic_id / actor_id が nil のシステム操作でも作成できる", ...)` は実DDLの clinic_id NOT NULL では不可能な挙動をテストとして固定している（ClinicID=nil は service 層 validateAuditLog audit_service.go:94-96 が拒否するため運用上は到達しないが、テストスキーマと実DDLの乖離の証左）。

**問題**

監査ログ書込の実行時整合が実DDLに対して未検証。ip_address が実DDL通り inet なら、(a) ベストエフォート監査（LSTEP/カルテ/バイタル/追記/薬量逸脱等）は slog.Warn で握り潰され監査証跡がサイレント消失（例: line_link_service.go:182-184）、(b) fail-closed tx 監査（#211 exam/checkup replace）は監査INSERT失敗→tx rollback→PUT が 500 になる。552本のテストは ekarte_db_test（AutoMigrate製・ip_address=text/clinic_id nullable）で走るため、この乖離クラスをどのテストも検出できない。医療システムの監査証跡要件（audit_logs は削除禁止テーブル）に直結する。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

手順: (1) まず実挙動を確定する再現テストを追加: internal/repository/audit_real_ddl_test.go を新設し、checkup_field_cascade_test.go と同じ手法で 001_init.sql の audit_logs DDL 原文（inet/NOT NULL含む）からテーブルを再作成（setupIsolatedTestDB 使用）し、IPAddress="" の repo.Create が失敗するか検証する（RED想定）。(2) 失敗が確認されたら bug トラックへ: model.AuditLog.IPAddress を `*string` 化し buildAuditLog（audit_service.go:74-88）で空文字→nil 正規化（inet 列へは NULL が入る）。LogAuthLogin/LogClinicSwitch の string 引数シグネチャは維持し内部で変換。JSON応答互換が必要なら handler/response 変換で nil→"" を維持（現状 audit_logs の read API は無く影響は限定的 — audit_repository.go は Create/CreateTx のみ）。(3) ClinicID には `gorm:"not null"` を付与しテストスキーマを実DDLの NOT NULL に整合させ、audit_repository_test.go の nil-clinic テストは「repo単体は通るが実DDLでは不可」という現状固定をやめ、real-DDL テスト側で NOT NULL 違反を検証する形に書き換える。(4) 既存 AutoMigrate ベースの setupAuditTestDB は real-DDL 版に置換。影響範囲: model/audit_log.go・service/audit_service.go・repository/audit_repository_test.go のみ（migration 不要 — DDL 側は正とする）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestAuditRepository|TestAuditLogRealDDL' -count=1 && docker compose exec backend go test ./internal/service/ -run TestAuditService -count=1
```

### X-4. billingItemService.CreateItem が trimming_course_id/trimming_option_id を未検証で永続化

- **ID**: `billing-item-trimming-fk-unguarded`
- **重要度**: P1 / **工数目安**: M
- **対象ファイル**: internal/service/billing_item_service.go (167-231); internal/service/master_fk_write_inventory_lint_test.go (192)

**証拠(現HEAD検証済み)**

billing_item_service.go:220-231: `item := &model.BillingItem{ ... TrimmingCourseID: input.TrimmingCourseID, TrimmingOptionID: input.TrimmingOptionID, ... }` の直前に billing のテナント所有権確認 (`s.billingRepo.FindByID(ctx, input.ClinicID, input.BillingID)`) はあるが、TrimmingCourseID/TrimmingOptionID 自体の所有権検証は無い。master_fk_write_inventory_lint_test.go:192 が自己申告で `{"billingItemService.CreateItem", statusKnownUnguarded, []string{"MerchandiseItemID", "TrimmingCourseID", "TrimmingOptionID"}, "all three FKs persisted directly without FindByID (billing_item_service.go:230)."}` と記録済み。

**問題**

P3.1 write側の原則（『正本は各サイトのruntime isolation test』）に反し、他クリニックの trimming_course_id/trimming_option_id を明細に紐付け可能。billing_item は clinic_id を自前で持たず billings 経由スコープのため、書込時点で検証しないと以後の CountUsageByTrimmingCourseID 等のクロステナント集計汚染・データ整合性破壊につながる（#124/#125 と同型のクラス）。read側 Preload は model に TrimmingCourse/TrimmingOption 関連が定義されていないため直接の情報漏洩は未確認だが、write-side 整合性違反は独立した問題。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

billing_item_service.go の CreateItem で input.TrimmingCourseID/TrimmingOptionID が非nilの場合、既存の trimmingCourseRepo.FindByID(ctx, input.ClinicID, *id) / trimmingOptionRepo.FindByID 相当（未配線なら repository に追加）で所有権検証してから item に設定する（examinationService.ReplaceItems の #124 修正パターンに倣う）。cross_tenant_master_fk_write_test.go に `TestBillingItemService_CreateItem_RejectsCrossClinicTrimmingFK` を追加し、GREEN化後に allowlist の該当行を `statusGuarded` へ更新する。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestBillingItemService -v
```

### X-5. campaignService.Create/Update が campaign_target_items.merchandise_item_id を未検証で永続化

- **ID**: `campaign-target-item-fk-unguarded`
- **重要度**: P1 / **工数目安**: M
- **対象ファイル**: internal/service/campaign_service.go (83-89,204-233,235-283); internal/repository/campaign_repository.go (40-70)

**証拠(現HEAD検証済み)**

campaign_service.go:83-89 `buildCampaignTargetItems(itemIDs []uint64) []model.CampaignTargetItem { ... targets = append(targets, model.CampaignTargetItem{MerchandiseItemID: id}) ... }` は id の所有権を検証せず、Create(224行目 `TargetItems: buildCampaignTargetItems(input.TargetItemIDs)`)・Update(272行目 `s.repo.ReplaceTargets(ctx, id, cats, itemIDs)`) の双方で無検証のまま merchandise_item_id を書き込む。model.CampaignTargetItem（campaign.go:52-58）は clinic_id 列を持たない純粋な junction。

**問題**

他クリニックの merchandise_item_id をキャンペーン対象に紐付け可能（テナント越境ID混入）。CalculateItemCampaignDiscount のマッチングロジックが誤発火し、割引適用の整合性が壊れる。read Preload(`Preload("TargetItems")`)自体は merchandise item 本体を展開しないため直接の名称/価格漏洩はないが、write-side 整合性違反として P3.1 の対象クラスに該当する。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

campaignService.Create/Update で input.TargetItemIDs の各要素を merchandiseItemRepo.FindByID(ctx, clinicID, id)（または既存の CountByIDs 一括検証）でクリニック所有権を確認してから buildCampaignTargetItems に渡す。repository.CampaignRepository に検証済みIDのみ受け付ける契約を明記し、cross_tenant_master_fk_write_test.go に isolation test を追加後、allowlist の 2 行を `statusGuarded` に更新する。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestCampaignService -v
```

### X-6. medicine/inventory repository の全 write が r.db 直参照で Transactor.WithTx に非参加 — BUG-429/R1-2 の原子性・fail-closed 監査が実際には効いていない

- **ID**: `tx-medicine-inventory-nonparticipation`
- **重要度**: P1 / **工数目安**: M
- **対象ファイル**: internal/repository/medicine_repository.go (94-130); internal/repository/inventory_repository.go (66-76, 120-143); internal/service/medicine_service.go (266-298, 407-431, 482-494)
- **依存関係**: なし（tx-mechanism-consolidation の一部を先行実施する形）

**証拠(現HEAD検証済み)**

medicine_service.go:266-269「// BUG-429: 薬剤作成と在庫アイテム自動作成をトランザクションでアトミックに実行\n// BE-refactor.md R1-2 (D1): per_weight 有効化監査も同一 tx に統合する（fail-closed）。\nif err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {\n    if err := s.repo.Create(txCtx, medicine); err != nil {」に対し、medicine_repository.go:94-95「func (r *medicineRepository) Create(ctx context.Context, medicine *model.Medicine) error {\n    err := r.db.WithContext(ctx).Create(medicine).Error」、inventory_repository.go:66-68「func (r *inventoryRepository) Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error {\n    item.ClinicID = clinicID\n    err := r.db.WithContext(ctx).Create(item).Error」。両 repo は dbOrTx を一切使わない（ファイル全体 grep で dbOrTx 0件）。さらに medicine_service.go:307-309 は「// BE-refactor.md R1-2: 呼び出し元の ambient tx に参加する LogEntryTx を使う。失敗時は呼び出し元の\n// WithTx が rollback し、薬剤作成/更新自体も無効になる（#211/refund パターン踏襲）。」と主張するが、薬剤/在庫の書込は tx 外の autocommit で先に確定するため rollback されない。Update 経路 (medicine_service.go:407-410 コメント「fields 更新・連携在庫名同期・per_weight 有効化監査を単一 tx に統合する」) と Delete 経路 (482-486 コメント「BUG-429: 薬剤削除と連携在庫削除をトランザクションでアトミックに実行」) も同様に medicine_repository.Update(:103)/Delete(:122)・inventory UpdateNameByMedicineCategory(:134)/DeleteByNameAndMedicineCategory(:121) が r.db 直参照。

**問題**

Transactor.WithTx は ctx に txKey を積むだけで、repo 側が dbOrTx(ctx, r.db) を使わない限り tx に参加しない (transactor.go:28-42, base.go:30-35)。medicine_service の 3 つの WithTx ブロック（Create/Update/Delete）はコメントで原子性と fail-closed 監査（#201 per_weight は安全クリティカル設定）を明言しているが、実際は薬剤・在庫の各書込が独立 autocommit で確定する。①在庫作成失敗時に薬剤だけ残る（BUG-429 の再発）、②per_weight 有効化監査の書込失敗時に監査なしで薬剤が per_weight 化される（fail-closed 破れ、R1-2 の意図に対する回帰相当）、③薬剤名変更と在庫名同期の不整合、が起こり得る。#213(KR-4) と同型の tx 非参加クラスで、accounting は R1-1(D2) で修正済みだが medicine/inventory は漏れている。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

手順: (1) internal/repository/medicine_repository.go の全メソッド（FindAll/FindByID/CountUsageByMedicineID/CountChildrenByParentID/Create/Update/Delete）の r.db.WithContext(ctx) を dbOrTx(ctx, r.db) に置換。Update は末尾で r.FindByID を呼ぶため（:114）、FindByID も dbOrTx 化しないと tx 内で更新前の stale 値を返す点に注意。Reorder(:117-118) は reorderByClinicID(ctx, r.db, ...) のままで可（ambient tx 呼出なし）。(2) internal/repository/inventory_repository.go の Create/Update/Delete/FindByID/UpdateNameByMedicineCategory/DeleteByNameAndMedicineCategory を同様に dbOrTx 化（DecreaseStock/FindAll/CountUsageByInventoryID は現状 tx 外呼出のみだが一括変換して統一可）。ambient tx が無い場合 dbOrTx は db.WithContext(ctx) と完全等価のため既存経路は挙動保存。(3) 回帰防止として internal/service/ に DB 付き rollback テストを追加（既存 audit_tx 検証 fe04b460 の temp-revert RED パターン踏襲）: per_weight 有効化時に auditTx を失敗させ、medicines/inventory_items に行が残らないことを assert。(4) 影響範囲: medicine_service / inventory_service（他に両 repo を使う service を grep で確認: prescription 系が medicine FindByID を読む程度で read のみ）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestMedicine' -count=1 && docker compose exec backend go test ./internal/repository/ -run 'TestMedicine|TestInventory' -count=1
```

### X-7. CreateClinic の WithTx 内で clinic_repository.Create / permission_group_repository.Create が tx 非参加 — デフォルト権限グループなしの孤児クリニックが生成しうる

- **ID**: `tx-clinic-create-nonparticipation`
- **重要度**: P2 / **工数目安**: S
- **対象ファイル**: internal/service/clinic_service.go (252-287); internal/repository/clinic_repository.go (80-89); internal/repository/permission_group_repository.go (64-70)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

clinic_service.go:252-253「if err := s.transactor.WithTx(ctx, func(ctx context.Context) error {\n    if err := s.repo.Create(ctx, clinic); err != nil {」→ :275「if err := s.permissionGroupRepo.Create(ctx, group); err != nil {」（執行/一般の2デフォルトPG作成）。しかし clinic_repository.go:80-81「func (r *clinicRepository) Create(ctx context.Context, clinic *model.Clinic) error {\n    err := r.db.WithContext(ctx).Create(clinic).Error」、permission_group_repository.go:64-65「func (r *permissionGroupRepository) Create(ctx context.Context, group *model.PermissionGroup) error {\n    err := r.db.WithContext(ctx).Create(group).Error」— 両ファイルとも dbOrTx 0件。clinic_repository は P4 clinicScope の例外ファイル（repository/CLAUDE.md:117-118）だがそれは tx 参加とは無関係。

**問題**

WithTx で括った意図（クリニック本体 + デフォルト権限グループ2件の原子作成）が実装上機能していない。3 write が独立 autocommit のため、PG 作成が1件目/2件目で失敗するとデフォルト権限グループの無い（または片方だけの）クリニックが残り、以後そのクリニックのスタッフ登録時に権限グループ選択肢が欠落する運用不整合となる。低頻度の管理操作だが復旧は手動 DB 修正になる。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

手順: (1) clinic_repository.go の Create/Update/FindByID/GetCompany を dbOrTx(ctx, r.db) 化（Delete は内部 tx 持ちのため dbOrTx(ctx, r.db).Transaction へ — R1-1 パターン accounting_repository.go:298 と同型。ただし Delete は ambient tx 呼出が現状無いので置換のみで挙動保存）。(2) permission_group_repository.go の Create を dbOrTx 化（同ファイルの他メソッドも一括変換可、UpdateRules/UpdateStaffGroups/Reorder の内部 tx は dbOrTx().Transaction 化）。(3) rollback テスト: permissionGroupRepo.Create を2件目で失敗させ clinics 行が残らないことを assert。影響範囲: clinic_service / permission_group_service。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestClinic|TestPermissionGroup' -count=1 && docker compose exec backend go test ./internal/repository/ -run 'TestPermissionGroup' -count=1
```

### X-8. reservation_staff_service.Create の外側 WithTx が無効 — repo.Create / UpdateExcludedReservationTypes がそれぞれ独立 r.db 内部 tx を張る

- **ID**: `tx-reservation-staff-nonparticipation`
- **重要度**: P2 / **工数目安**: M
- **対象ファイル**: internal/service/reservation_staff_service.go (125-139, 151-185); internal/repository/reservation_staff_repository.go (67-85, 210-250)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

reservation_staff_service.go:125-131「if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {\n    if err := s.repo.Create(txCtx, staff, clinicID); err != nil {\n...\n        if err := s.repo.UpdateExcludedReservationTypes(txCtx, clinicID, staff.ID, input.ExcludedTypeIDs); err != nil {」に対し、reservation_staff_repository.go:68「if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {」（Create: staff+assignment を自前 tx で作成）、:226「if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {」（UpdateExcludedReservationTypes: DELETE→INSERT を自前 tx で実行）。どちらも r.db 起点のため外側 WithTx の tx とは別セッション。ファイル全体で dbOrTx 0件。なお Update 経路 (service:151-185) はそもそも WithTx で括られておらず、staff 更新 (:160) と除外区分置換 (:166) が非原子。

**問題**

外側 WithTx の意図（スタッフ作成と除外予約区分設定の原子化）が実装上機能していない。staff+assignment 作成が確定した後に UpdateExcludedReservationTypes が失敗すると、除外設定なしのスタッフが残り WithTx の rollback は何も巻き戻さない（除外設定漏れ＝本来受けられない予約区分で予約可能になる運用不整合）。各内部 tx 単体は原子なので部分行断片は残らないが、2 操作間の原子性が無い。#213/R1-1(D2) と同じ「r.db 直参照による tx 非参加」クラス。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

手順: (1) reservation_staff_repository.go の Create(:68)・UpdateExcludedReservationTypes(:226)・UpdateReservationCapabilities(対称実装)・UpdateSortOrder(:130) の r.db.WithContext(ctx).Transaction を dbOrTx(ctx, r.db).Transaction に置換（accounting_repository.go:288-298 の R1-1 コメント付き先例と同型。ambient tx があれば SAVEPOINT ネストで参加、無ければ従来どおり独立 tx = 挙動保存）。UpdateExcludedReservationTypes 冒頭の所有権検証 read (:216) も dbOrTx 化して同一 tx 内で読む。(2) 単発 write の Update(:88)/Delete(:103) も dbOrTx 化して統一。(3) service 側 Update (:151-185) の staff 更新+除外置換を Create と同様に WithTx で括るかは挙動変更トラックで判断（現状原子性の主張コメントは無い）。(4) rollback テスト: Create 内で UpdateExcludedReservationTypes を失敗させ staffs/staff_clinic_assignments に行が残らないことを assert。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestReservationStaff' -count=1 && docker compose exec backend go test ./internal/repository/ -run 'TestReservationStaff' -count=1
```

### X-9. 予約枠競合チェックの SELECT FOR UPDATE は空枠のファントム挿入を防げず、同時予約で重複/超過予約が成立する

- **ID**: `resv-slot-phantom-toctou`
- **重要度**: P2 / **工数目安**: M
- **対象ファイル**: internal/repository/reservation_repository.go (294-316); internal/service/reservation_capacity.go (33-59); internal/service/reservation_service.go (208-221); internal/service/reservation_validators.go (92-139); migrations/001_init.sql (2692-2695)
- **依存関係**: なし。dbortx_inventory_lint_test.go の allowlist 更新を同一コミットで行うこと

**証拠(現HEAD検証済み)**

reservation_repository.go:301-309 `err := dbOrTx(ctx, r.db).Raw(`\n\t\tSELECT id FROM appointments\n\t\tWHERE clinic_id = ?\n\t\t  AND deleted_at IS NULL\n\t\t  AND status NOT IN ('cancelled')\n\t\t  AND start_time < ?\n\t\t  AND end_time > ?\n\t\t  AND (? = 0 OR id != ?)\n\t\tFOR UPDATE`` — FOR UPDATE は既存行のみロックし、条件に合致する行が 0 件（空枠）の場合は何もロックしない。型別容量は reservation_capacity.go:51 `count, err := reservationRepo.CountByTypeAndStartTime(ctx, clinicID, reservationTypeID, startTime, excludeID)` の素の COUNT（FOR UPDATE すら無し、reservation_repository.go:318-329）。DB バックストップは migrations/001_init.sql:2693-2695 `CREATE UNIQUE INDEX uk_appointment_staff_time ON appointments (clinic_id, doctor_id, start_time) WHERE deleted_at IS NULL AND status != 'cancelled';` のみで、(a) doctor_id が NULL の容量枠予約、(b) 同一医師の start_time が異なる時間重複（10:00-10:30 vs 10:15-10:45）、(c) MaxConcurrent 上限の型別容量は表現できない。reservation_service.go:208-210 のコメント「SELECT FOR UPDATE + トランザクションで競合を防止」は空枠ケースでは成立しない。

**問題**

check-then-act（競合カウント→INSERT）で、READ COMMITTED のスナップショット外で並行 tx が INSERT した行は数えられずロックもできない（ファントム）。空き枠に対する同時 2 リクエスト（LINE 予約は不特定多数から到達）が双方 0 件/上限未満と判定し双方 INSERT → 医師二重予約・出勤医師数超過・予約区分 MaxConcurrent 超過が成立する。既存 FOR UPDATE は「既存行がある場合の直列化」しか提供せず、コメントが主張する防止保証と実装が乖離している。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

1) ReservationRepository に `AcquireBookingLock(ctx context.Context, clinicID uint64) error` を新設し `SELECT pg_advisory_xact_lock(hashtextextended('appointments:' || ?::text, 0))`（clinic 単位のトランザクションスコープ advisory lock、dbOrTx 参加）を実装。2) reservation_service.go の Create(WithTx 内先頭・enforceBookingConstraints 時のみ)・updateWithConflictCheck(WithTx 内先頭)、reservation_validators.go ValidateAndCreate(WithTx 内先頭) の 3 箇所で競合チェック前に取得。3) dbortx_inventory_lint_test.go の allowlist に新メソッドを登録。4) 並行性テストを internal/repository/ に追加（2 goroutine が同一空枠を同時予約し、ちょうど 1 件だけ成功することを実 DB で検証。ltv_repository_test.go:1562 の advisory lock 利用が参考）。clinic 単位ロックで粒度が粗い懸念は、予約書込レートが clinic あたり低頻度のため許容（過剰な粒度細分化は不要）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestReservation|TestDBOrTxInventory' -count=1 && docker compose exec backend go test ./internal/service/ -run TestReservation -count=1
```

### X-10. カルテ楽観ロックの version 照合が UPDATE の WHERE に含まれず、並行 draft 編集で lost update が成立する

- **ID**: `mr-version-check-not-atomic`
- **重要度**: P2 / **工数目安**: S
- **対象ファイル**: internal/service/medical_record_crud.go (214-264); internal/repository/medical_record_repository.go (236-250)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

medical_record_crud.go:230-233 「// version が指定されている場合は一致確認\n\tif input.Version != nil && existing.Version != *input.Version {\n\t\treturn nil, apperrors.WrapConflict("他のユーザーがこのカルテを変更しました。再読み込みしてください")\n\t}」→ 同 254 行 `fields["version"] = existing.Version + 1` → repo 側 medical_record_repository.go:237-241 「result := r.db.WithContext(ctx).\n\t\tModel(&model.MedicalRecord{}).\n\t\tScopes(clinicScope(clinicID)).\n\t\tWhere("id = ? AND status = ?", id, model.MedicalRecordStatusDraft).\n\t\tUpdates(fields)」— WHERE は id+clinic+status のみで version 述語が無い。service 層の一致確認は tx 外の FindByID(216 行) 読取に対する check-then-act。

**問題**

2 名のスタッフが同一 version=N を読んで同時に更新すると、両者とも service 層チェックを通過し、両 UPDATE が WHERE(id, status=draft) にマッチして成功する。後勝ちがフィールドを黙って上書きし、かつ両者とも version=N+1 を書くため、以後のクライアントも競合を検出できない（楽観ロックの目的である lost update 防止が並行時に機能しない）。status=draft 述語は finalize 済みへの編集は原子的に防いでいるが、draft 同士の競合は防げない。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

1) medical_record_repository.go の Update シグネチャに `expectedVersion *uint64`（または fields とは別引数）を追加し、非 nil 時に `Where("version = ?", *expectedVersion)` を付与。2) RowsAffected==0 の場合、tx 内で status を再読して「not draft」（既存 Conflict メッセージ維持）と「version 不一致」（既存の『他のユーザーがこのカルテを変更しました』メッセージ）を区別して返す。3) medical_record_crud.go:214-264 の Update から service 層 version 照合を repo 述語に一本化（input.Version==nil の従来挙動＝照合なしは保存）。4) 並行 2 goroutine で同一 version から同時更新し片方だけ成功するテストを internal/repository/ に追加。呼出面: medical_record repo Update の他呼出箇所のシグネチャ追随が必要（grep で確認のこと）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run TestMedicalRecord -count=1 && docker compose exec backend go test ./internal/service/ -run TestMedicalRecord -count=1
```

### X-11. カルテ確定ロック(HC-003/005/006)の親 status チェックが子エンティティ書込と非原子で、確定と同時の子追加/編集が確定済カルテに混入しうる

- **ID**: `finalize-child-write-race`
- **重要度**: P3 / **工数目安**: L
- **対象ファイル**: internal/service/treatment_service.go (220-303); internal/service/examination_service.go (152-160); internal/service/vital_service.go (106-113); internal/service/prescription_service.go (83-84); internal/service/checkup_field_result_service.go (128-135); internal/repository/medical_record_repository.go (236-248)
- **依存関係**: resv-slot-phantom-toctou と同じく dbortx_inventory_lint_test.go allowlist 更新を伴う。5 サービス横断のため実装は 2 コミット以上に分割推奨

**証拠(現HEAD検証済み)**

treatment_service.go:222-229 「parent, err := s.repos.MedicalRecord.FindByID(ctx, clinicID, medicalRecordID)\n\t...\n\tif parent.Status == model.MedicalRecordStatusFinalized {\n\t\treturn nil, apperrors.WrapConflict("確定済みカルテには治療を追加できません")\n\t}」— 素の FindByID（無ロック・tx 外）でチェックした後、242 行 `err = s.repos.Transaction(ctx, func(txRepos *repository.Repositories) error {` の別 tx で子 INSERT。examination_service.go:159-160・vital_service.go:112-113・prescription_service.go:83-84・checkup_field_result_service.go:134-135 も同型（FindByID→status チェック→書込、親行ロックなし）。finalize 側 medical_record_repository.go:240 の `Where("id = ? AND status = ?", id, model.MedicalRecordStatusDraft)` はカルテ本体のみ原子化し、子テーブルには波及しない。

**問題**

T1(子追加) が parent.Status=draft を確認 → T2(finalize) がコミット → T1 の子 INSERT がコミット、の順序が可能で、確定済み（改変ロック済み）カルテに監査上 finalize 後の子レコードが追記なし(addendum 経由でなく)で混入する。確定ロックは臨床データの改竄追跡性の要の不変条件（HC-003/005/006・EXAM-001）だが、その並行時の強制が check-then-act のみ。競合窓はミリ秒級で発生頻度は低いものの、発生時は silent で検出手段がない。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

挙動変更トラック。1) medical_record_repository.go に `LockDraftByID(ctx, clinicID, id) (*model.MedicalRecord, error)`（dbOrTx + `Clauses(clause.Locking{Strength: "UPDATE"})` + status 返却）を新設し dbortx lint allowlist に登録。2) 子書込 5 サービス（treatment/examination/vital/prescription/checkup_field_result — treatment は既存 repos.Transaction 内へ、他は Transactor.WithTx 導入）で、tx 内先頭に LockDraftByID → finalized なら既存と同一メッセージの Conflict を返し、子書込を同一 tx 内へ移す。finalize 側 UPDATE は同一行への行ロックを要求するため、子 tx 保持中の finalize は自然に待機し順序整合する（finalize 側の変更は不要）。3) 並行テスト（finalize と子追加を同時実行し、finalize 後の子混入ゼロを検証）を 1 サービス分（treatment）追加し、残りはパターン踏襲。段階導入可: まず treatment/examination（金額・検査値を持つ高リスク 2 系統）のみでも価値がある。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestTreatment|TestExamination' -count=1 && docker compose exec backend go test ./internal/repository/ -run TestDBOrTxInventory -count=1
```

### X-12. 会計 Create/Update の会計完了時 appointment 完了化が tx 外で、部分コミット（billing 確定済み・予約カード残留/エラー返却）が起こる

- **ID**: `billing-complete-appt-post-tx`
- **重要度**: P3 / **工数目安**: M
- **対象ファイル**: internal/service/accounting_service_core.go (80-86, 195-200, 251-267); internal/repository/accounting_repository.go (312-349)
- **依存関係**: dbortx_inventory_lint_test.go allowlist 更新を同一コミットで行うこと

**証拠(現HEAD検証済み)**

accounting_service_core.go:195-198 「if input.Status != nil && *input.Status == model.BillingStatusCompleted {\n\t\tif err := s.completeAccountingAppointments(ctx, input.ClinicID, accounting); err != nil {\n\t\t\treturn nil, apperrors.Wrap(err, "failed to complete accounting appointments during update")\n\t\t}」— 148-182 行の WithTx（fields/payment/splits/監査を R1-2 で原子化済み）がコミットした後に、tx 外の ctx で呼ばれる。repo 側 accounting_repository.go:317 `result := r.db.WithContext(ctx).` / 333 行も同じく r.db 直参照で dbOrTx 非参加。Create 側も同型: accounting_service_core.go:80-83 「if billing.Status == model.BillingStatusCompleted {\n\t\tif err := s.completeAccountingAppointments(ctx, input.ClinicID, billing); err != nil {\n\t\t\treturn nil, apperrors.Wrap(err, "failed to complete accounting appointments during create")」— billing Create(73 行) コミット後に失敗するとエラーを返すが billing は残る。

**問題**

会計確定は WithTx でコミット済みなのに appointment 完了化が失敗すると、(a) 呼出元にはエラーが返り操作全体が失敗に見えるが billing は completed で確定済み（Update 側）、(b) Create 側は billing が残ったままエラーになり、medical_record_id NULL のトリミング/手動会計ではリトライで二重 billing を作れる（idx_billings_medical_record_id_unique は medical_record_id 非 NULL のみバックストップ）。R1-2 が塞いだ「三系統分裂の部分コミット」と同型の残余が、同一ユースケースの後段に残っている。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

挙動変更トラック（失敗時の原子性が変わる）。1) accounting_repository.go の CompleteAccountingAppointments 内 2 箇所（317, 333 行）を dbOrTx(ctx, r.db) に変更し dbortx lint allowlist に登録。2) Update 側: 呼出を WithTx クロージャ末尾（logPostCloseEdit の後）へ移動し txCtx で呼ぶ。判定は input.Status ベースに変更（現在は再読後 accounting を使うが、tx 内では fields 適用済みのため同値）。3) Create 側: repo.Create + completeAccountingAppointments を Transactor.WithTx で括る（Billing Create は既に単文なので repo 変更不要、Create が dbOrTx 未参加なら参加化）。4) syncCPMStageTag は外部 LSTEP 同期なので従来どおり tx 外 best-effort を維持。5) accounting_repository_tx_atomicity_test.go に「appointment 完了化失敗で billing 更新もロールバックする」ケースを追加。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestAccounting.*Atomicity|TestDBOrTxInventory' -count=1 && docker compose exec backend go test ./internal/service/ -run TestAccounting -count=1
```

### X-13. SharedFile.DeletedAt が *time.Time のため repo.Delete が物理 DELETE — DDL/読取述語/ログ文言のソフトデリート意図と不整合

- **ID**: `sharedfile-harddelete-vs-softdelete-intent`
- **重要度**: P2 / **工数目安**: S
- **対象ファイル**: internal/model/shared_file.go (18); internal/repository/shared_file_repository.go (62-71); internal/service/shared_file_service.go (133-175); migrations/001_init.sql (1250-1272)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

model (shared_file.go:18): `DeletedAt  *time.Time \`gorm:"index"          json:"deleted_at"\`` — gorm.DeletedAt でないため GORM のソフトデリートは発火せず、shared_file_repository.go:62-66 の `Delete(&model.SharedFile{})` は物理 DELETE を発行する。一方で意図はソフトデリート: DDL は deleted_at 列を持ち（001_init.sql:1262）部分インデックス `WHERE deleted_at IS NULL`（:1269-1271）を張り、読取は全て手動述語 `Where("deleted_at IS NULL")`（repo :41,:53,:76）、service のエラーログは `"failed to soft-delete shared file"`（shared_file_service.go:144）/`"failed to soft-delete expired shared file"`（:171）と明記。同型の LstepTagCodeMapping は SoftDelete メソッドで `Update("deleted_at", now)` を明示実装しており（lstep_tag_code_mapping_repository.go:81-98）、SharedFile だけ意図と実装が食い違う。

**問題**

LINE個別送信ファイルのメタデータ（誰がどの飼主向けに何をアップロードしたか）が削除・期限切れクリーンアップ時に物理消去され、deleted_at 列・部分インデックス・読取述語が全て死んでいる。監査・追跡可能性の設計意図（業務データは deleted_at を持つ: migrations/CLAUDE.md 必須チェック）に反する。ソフトデリート化は挙動変更（行が残る・一意/容量特性が変わる）のため別トラック。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

behaviorChange トラック。選択肢を PO 判断可能な形で提示: 案A（意図に合わせる）= DeletedAt を gorm.DeletedAt に変更（json:"-" 化で API から deleted_at フィールドが消える点は要確認 — 現状 json:"deleted_at" を露出）。repo の手動 `deleted_at IS NULL` 述語は GORM 自動述語と重複するが無害なので残置可。FindExpired のクリーンアップが物理削除を意図するなら Unscoped().Delete を明示。案B（実装に合わせる）= ハードデリートを正と決め、model から DeletedAt を除去し新規 migration で列と部分インデックスを drop、service のログ文言を hard-delete に修正。どちらでも「意図と実装の一致」を repository テスト（Delete 後に Unscoped 検索で行の有無を検証）で固定する。影響範囲: model/shared_file.go・repository/shared_file_repository.go・service/shared_file_service.go（案Bのみ migrations 追加）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run TestSharedFile -count=1 && docker compose exec backend go test ./internal/service/ -run TestSharedFile -count=1
```

### X-14. master-FK write allowlistのknown-unguarded約47エントリにisolation test不在(名簿上も『NO dedicated isolation test』と明記)

- **ID**: `test-known-unguarded-master-fk-isolation-tests`
- **重要度**: P2 / **工数目安**: L
- **対象ファイル**: internal/service/master_fk_write_inventory_lint_test.go (143-208)
- **依存関係**: 各エントリのガード実装(挙動変更)が先行必須。テストのみ先行するとCIがREDになる

**証拠(現HEAD検証済み)**

master_fk_write_inventory_lint_test.go:143-145「// statusKnownUnguarded: reviewed; NO dedicated isolation test confirms ownership\n statusKnownUnguarded masterFKWriteStatus = "known-unguarded"」。エントリ例(同:191-192)「{"accountingService.Update", statusKnownUnguarded, []string{"PaymentMethodID"}, "PaymentMethodID resolved via clinic system_key→ID map (resolvePaymentMethodMasterID); not a FindByID guard and no isolation test — verify rejection of explicit foreign IDs."},\n{"billingItemService.CreateItem", statusKnownUnguarded, []string{"MerchandiseItemID", "TrimmingCourseID", "TrimmingOptionID"}, "all three FKs persisted directly without FindByID (billing_item_service.go:230)."}」。grep計測でstatusKnownUnguarded言及49行(定義2行を除き約47エントリ)。repository/CLAUDE.mdの規約は「正本ガード = 各サイトの runtime isolation test」だが、これらのエントリにはその正本ガードが存在しない。

**問題**

review網羅性gateは『名簿に載せる』ことしか担保せず(同ファイル冒頭に明記)、known-unguardedのまま滞留しているwrite経路はクロステナントmaster FK書き込みが実際に拒否されるかを誰も検証していない。会計ドメイン(accountingService.Update / billingItemService.CreateItem)を含む点が特に優先度が高い。テストを書くとRED(ガード自体が無い)になるため、これはガード実装を伴う挙動変更トラック。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

別トラック(挙動変更)として段階実施。優先順: (1)会計: accountingService.Update PaymentMethodID / billingItemService.CreateItem の3FK — service層にFindByID(clinicID,…)ガード追加後、internal/repository/または既存cross_tenant_master_fk_write_test.goパターンで『別クリニックFK指定→apperrors.WrapInvalidInput/NotFound拒否』テストを追加、allowlistエントリをguardedへ更新。(2)campaign TargetItemIDs(repo ReplaceTargets unscoped)。(3)carePlanItem HospitalizationPlanID / hospitalization CageID。(4)self-ref ParentID群(checkupType/consultation/examType)。各バッチはTestMasterFKWriteInventoryのstatus突合がgateになるため、allowlist更新漏れはCIで検出される。一括ではなくドメイン毎に1PRずつ、STGデータ監査(既存越境データの有無)を先行させること(R1-3の教訓)。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestMasterFKWriteInventory -count=1 && docker compose exec backend go test ./internal/repository/ -run TestCrossTenantMasterFKWrite -count=1
```

### X-15. P6逸脱: 状態トグル系 DELETE 4ルートが "delete" ではなく "edit" 権限でゲート(免除根拠コメントなし)

- **ID**: `p6-delete-routes-edit-permission`
- **重要度**: P3 / **工数目安**: S
- **対象ファイル**: internal/handler/handler.go (158,185); internal/handler/pet_handler.go (158,165)
- **依存関係**: PO判断(権限ポリシー)。案(b)の場合は FE の権限ガードに波及の可能性

**証拠(現HEAD検証済み)**

internal/handler/handler.go:158: `owners.DELETE("/:id/lstep-opt-out", h.RequirePermission(string(model.ResourceOwners), "edit"), h.DeleteOwnerLstepOptOut)`(185 に co エイリアス同等行)。internal/handler/pet_handler.go:158: `pets.DELETE("/:id/death", h.RequirePermission(string(model.ResourceOwners), "edit"), h.DeletePetDeath)`(165 に clinicPets エイリアス同等行)。対照: 同一リソースの handler.go:155 `owners.DELETE("/:id/line", h.RequirePermission(string(model.ResourceOwners), "delete"), h.DeleteOwnerLine)`、handler.go:170 の lstep/tags DELETE も "delete"。セマンティクス裏付け: internal/handler/lstep_lifecycle_handler.go:80 「DELETE /owners/:id/lstep-opt-out — オーナーを Lステップ配信にオプトインする（BE-017）」、同39 「DELETE /pets/:id/death — ペット死亡取り消しを記録し CPM タグを再同期する」。等価操作の統合エンドポイント PatchOwnerLstepOptOut(PATCH, handler.go:160)も "edit"。ルート登録箇所の前後コメント(handler.go:156-157 / pet_handler.go:156)に権限選定の根拠記載なし。

**問題**

P6 は「DELETE ルートには delete 権限（edit は違反）」と規定し、全 DELETE ルート走査でこの4行のみが "edit"(LIFF公開 CancelLiffReservation は免除)。RBAC 上 delete 権限を剥奪してもこの4操作は edit 保持者に実行可能。ただし実体は資源削除ではなく状態解除(オプトイン/死亡取消)で、同一操作が PATCH(edit) でも実行できるため delete に揃えると同一業務操作の必要権限がエンドポイント表現で割れる。意図的設計の可能性が高いが明文化されておらず、P6 スキャン運用の恒常的偽陽性源になっている。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

PO/責任者に「undo系DELETE の要求権限を delete に上げるか、P6 例外として明文化するか」を確定させてから着手。(a) 例外維持(挙動保存): 4ルート行直上に「P6例外: 状態トグル(資源削除でない)のため PATCH 等価の edit を要求(PO決定日付)」コメントを追記し、.claude/refs/gin-architecture-compliance.md の P6 節と backend/internal/handler/CLAUDE.md に免除注記を追加。(b) delete化(挙動変更): handler.go:158,185 / pet_handler.go:158,165 の action を "delete" に変更し、internal/handler/lstep_lifecycle_handler_test.go の TestDeletePetDeath / TestDeleteOwnerLstepOptOut に権限不足403ケースを追加、FE 側の該当操作 can 判定への波及を grep 確認。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/handler/ -run 'TestDeletePetDeath|TestDeleteOwnerLstepOptOut|TestRegisterRoutes_NoPanic' -count=1
```

### X-16. 健診クリニック横断一覧がページネーションなし・フィルタ全optional(全件+5 Preload)、期限アラートも下限なし全滞留分を返す

- **ID**: `checkup-list-unbounded`
- **重要度**: P3 / **工数目安**: M
- **対象ファイル**: internal/repository/checkup_repository.go (44-70,107-124); internal/handler/checkup_request.go (89-112); internal/service/checkup_service.go (117-130)
- **依存関係**: FE同期必須。FindAlerts の下限はPO判断。checkups-vaccinations-missing-composite-index を先行推奨

**証拠(現HEAD検証済み)**

checkup_repository.go:44-70 FindByClinicID は `if filters.StartDate != nil { q = q.Where(...) }` と全フィルタが optional で、:65 `err := q.Order("date DESC").Find(&checkups).Error` に Limit/Offset なし。:48-52 で `Preload("CheckupType"...).Preload("Doctor"...).Preload("MedicalRecord"...).Preload("MedicalRecord.Pet"...).Preload("MedicalRecord.Pet.Owner"...)` の5 Preload。checkup_request.go:98-101 は `optionalStringQueryFilter(values.Get("start_date"))` でクエリ未指定を許容し必須化していない。FindAlerts(:116-118)は `next_date <= upperBound` のみで下限・LIMITなし＝過去の期限切れ全件を毎回返す。他の取引系一覧(owners/pets/medical_records/vaccinations/examinations/hospitalizations/estimates/treatments 等)は全て page/limit 実装済みで、checkups のみ例外。

**問題**

健診記録は診療ごとに増える成長テーブルであり、FE がフィルタを省略した場合(またはブラウザから直接叩かれた場合)にクリニック全履歴+5関連の全件シリアライズが走る。ページネーション規約(gin-api-design)からの逸脱でもある。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

挙動変更トラック(API形状変更)として扱う。(1) FindByClinicID に page/limit を追加し vaccination_repository.go:35-70 と同型(buildBase closure + Count + Offset/Limit)へ。handler は owners 等の既存 page/limit クエリ規約に合わせ既定 limit を設定。(2) FE(健診管理ページ)のクエリ同期が必要。(3) FindAlerts は業務要件(過去滞留分をどこまで表示するか)を PO 確認の上で下限日付 or LIMIT を導入。先行して checkups-vaccinations-missing-composite-index のインデックスを入れれば当面の劣化は緩和される。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run TestCheckupRepository -count=1 && docker compose exec backend go test ./internal/service/ -run 'TestCheckupService' -count=1
```

### X-17. RequireXRequestedWithのエラーレスポンスがmiddleware共通スキーマ（code/message/timestamp）から逸脱し{"error":...}を返す

- **ID**: `csrf-error-schema-drift`
- **重要度**: P3 / **工数目安**: S
- **対象ファイル**: internal/middleware/csrf.go (22-27); internal/middleware/response.go (9-17)

**証拠(現HEAD検証済み)**

internal/middleware/csrf.go:22-26:
		if c.Request.Header.Get("X-Requested-With") == "" {
			err := apperrors.WrapForbidden("X-Requested-With header required for state-changing requests")
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			c.Abort()

internal/middleware/response.go:9-16:
// respondError はミドルウェア層共通のエラーレスポンスを返す。
// handler 層の RespondError と同一スキーマ（code/message/timestamp）を使用する。
func respondError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"code":      status,
		"message":   msg,
		"timestamp": time.Now(),

**問題**

middleware層は respondError で handler 層 RespondError と同一のエラーエンベロープ（code/message/timestamp）に統一されている（auth.go/liff_auth.go/rate_limit.go は全て準拠）が、csrf.go のみ {"error": "..."} 形式を返す。FE のエラーハンドラがスキーマ別分岐を強いられる一貫性負債。レスポンスボディ形状が変わるため挙動変更扱い。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

手順: (1) csrf.go:23-26 を respondError(c, http.StatusForbidden, "X-Requested-With header required for state-changing requests") の1行に置換（apperrors.WrapForbidden 生成は不要になる。ステータス403は不変）。(2) csrf_test.go のボディ検証を新スキーマに更新。(3) frontend 側で当該 403 の "error" キーをパースしている箇所がないか grep（axios interceptor が message キー前提なら影響なし）確認の上で適用。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/middleware/ -run RequireXRequestedWith -count=1
```

### X-18. password_reset の30sタイムアウトが smtp.SendMail に非伝播（goroutine が無期限にブロックしうる）

- **ID**: `KR-pwreset-smtp-timeout`
- **重要度**: P2 / **工数目安**: S
- **対象ファイル**: internal/service/password_reset_service.go (101-109, 178-210)
- **依存関係**: なし（単独ファイル完結）

**証拠(現HEAD検証済み)**

internal/service/password_reset_service.go:101-108:
	go func() { //nolint:gosec,contextcheck // fire-and-forget: request ctx キャンセル後も送信継続が必要なため context.Background を使用
		bgCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:gosec // 上記と同理由
		defer cancel()
		if sendErr := s.sendResetEmail(email, resetURL); sendErr != nil {
			slog.ErrorContext(bgCtx, "failed to send password reset email",
— bgCtx はログ出力にしか使われず sendResetEmail に渡らない。password_reset_service.go:206:
	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
— smtp.SendMail は context/deadline を受け取らないため 30s タイムアウトは実質無効。

**問題**

既知残存項目1の再確認（現HEADで現存）。SMTP サーバが応答しない場合、送信 goroutine が OS の TCP タイムアウトまで（あるいは無期限に）ブロックし、cancel は名ばかりになる。fire-and-forget なのでリクエストには影響しないが、リセットメール未達がタイムアウトログすら出さずに滞留し、goroutine リークの温床になる。

**対応方針(挙動変更を伴うため要PO/責任者判断のうえ別トラックで実施)**

挙動変更トラックで扱う。手順: 1) sendResetEmail(ctx context.Context, to, resetURL string) にシグネチャ変更。2) 実装を smtp.SendMail から (&net.Dialer{}).DialContext(ctx, "tcp", addr) + smtp.NewClient に置換し、conn.SetDeadline(deadline) を ctx の deadline から導出（Auth/Mail/Rcpt/Data/Quit を明示実行。STARTTLS 対応は現行 SendMail と同等の分岐を踏襲）。3) 呼び出し側 goroutine で sendResetEmail(bgCtx, ...) に変更し //nolint:contextcheck を除去。4) ハングするダミーリスナー（net.Listen して読み捨て）で「deadline 内に error 復帰する」回帰テストを追加（既存 TestPasswordResetService_SendResetEmail の隣）。影響範囲は本ファイルのみ（PasswordResetService interface は不変）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestPasswordResetService -count=1
```

---

## Appendix B: 既知残存・今回あえて対象外とした項目（状態確認のみ）

以下は過去の計画・監査で既に把握されており、今回の15次元監査でも独立に再検出された（＝現HEADで解消されていないことを確認した）。ただし対応方針が既に別途決まっている、または model 層の構造的制約により本計画のスコープでは着手できないため、状態確認の記録に留める。

### B-1. password_reset の30秒タイムアウトが smtp.SendMail に非伝播

`internal/service/password_reset_service.go:101-109, 178-210`。14次元中7次元（decomp/dup/idiom/perf/periph/service-p/silent/test/tx の実に9次元）が独立に再検出するほど顕在性が高い残存バグ。挙動変更（SMTP接続をcontext-aware dialに置換）を伴うため、Appendix A の `X-18 KR-pwreset-smtp-timeout` として統合済み。**次回セッションではこれを最優先の別トラックとして着手することを推奨する**（9次元が独立に同一結論に達したことは、実務上のリスクの高さの傍証でもある）。

### B-2. Preload read-lint 未登録の3マスタ

`internal/repository/preload_clinic_scope_lint_test.go` の read側 clinic_id lint に `MerchandiseItem` / `TrimmingCourseType` / `PaymentMethodMaster` の3マスタが未登録。5次元（decomp/dup/idiom/repo-p/schema/test）が独立に再確認。**model 層に GORM association が定義されていないため構文的に Preload できず、lint 登録自体が不可能**という構造的制約が2026-06-30時点から変わっていない。model に association を追加すること自体が設計変更（他コードへの波及要確認）のため、今回の挙動保存計画には含めない。対応するなら「①各3マスタに association 追加が安全か個別調査 → ②追加 → ③lint登録」の順で別途計画すること。
## Appendix C: 監査で確認したクリーン領域（対応不要）


### cmdtools

- cmd/migrate/main.go (552行): 分割不要と判定。フェーズ構造（advisory lock → legacy key fail-fast → baseline → DDL → CSV seed）が明確で、全関数に設計根拠コメントが付き、既知の限界（seed 記録の極小 crash window, main.go:467-473）まで文書化済み。csvbundle.go への責務分離も適切
- cmd/coverage-ratchet: 純粋関数 (parseTotalCoverage/evaluateRatchet) 分離 + main_test.go あり。clean
- ツール間の migration 適用ロジック重複なし: cmd/seed-export は自前実装せず cmd/migrate を exec で再利用（seed-export/main.go:173-177 に設計根拠コメント「seed-export must never reimplement "apply 001-004"」）。cmd/stage-import は migration を適用しない。重複は DSN/env/localHosts の接続ボイラープレートのみ（finding cmd-dsn-localhosts-duplication）
- cmd/seed-old-db: 存廃判定 = 当面 KEEP。cmd/stage-import/main.go:5 が「the deprecated direct old-db seeder」と明記する後継関係だが、old_db 本番データ移行が未完のため stage パイプライン不通時の TSV 直投入 fallback として運用価値が残る。本番移行完了後に docker-compose.seed-old-db.yml と併せて廃止再判定（transform.go 962行の keep 判定は維持）
- cmd/lstep-migrate: 存廃判定 = KEEP。clinic-id 指定・dry-run・rate-limit・resume-from を備えた運用 CLI で、新規クリニックの Lstep オンボーディング時に再利用される性質のもの。一回限りツールではない
- cmd/stage-import / cmd/seed-export の安全設計は良好: 非ローカルホスト拒否・stage 接続 read-only (default_transaction_read_only=on)・--apply と --confirm-local-destroy の二重フラグ・使い捨て DB 名ハードコード。テストも dsn_test/plan_test/integration_test 等で担保
- worker/index.ts: 薄いプロキシ + /_internal/migrate の設計コメント充実（DB_RESET 非注入・exec への最小権限 env subset・XFF 偽装対策・タイムアウト付き exec）。構造は clean（migrate-exec テストは G10-6 CLOSED（`a1429497`））
- entrypoint.sh (11行): set -e で migrate 失敗時に air を起動しない fail-fast 設計、コメントあり。clean
- .air.toml / Dockerfile.dev / tygo.yaml / wrangler.jsonc: 整合確認済み。Dockerfile.dev の golangci-lint v2.11.4 は CI と整合（コメントで ci.yml と相互参照）。tygo.yaml は ci.yml:557-562 の codegen drift check から現役参照
- backend/api 33MB ELF・lstep-migrate 17MB・seed-old-db/stage-import バイナリ・coverage.out・tmp/・.wrangler/ は全て gitignore 済みの未追跡ローカル成果物（唯一の例外が追跡下の backend/migrate = finding tracked-migrate-binary）。「正体不明のエントリ backend/api」の正体 = cmd/api の手動 go build 成果物（ELF ARM64、gitignore 済み、無害）
- 既知残存項目 KR-1〜6（password_reset タイムアウト / tz 直列化 / Preload lint 3マスタ / TOCTOU #213 / RLS #93 / junction 非スコープ）は全て internal/service・internal/repository・model 層の事象であり、本次元（cmd/・scripts・worker・Docker/インフラ）の対象ファイルには現存しない — 該当層の次元監査の担当

### contract

- tygo.yaml: model→frontend/src/types/generated/models.ts の生成設定・type_mappings は実態と整合。docs/README.md:15-29 の tygo パイプライン記述も Makefile の codegen ターゲット実在(Makefile:1)と一致
- api.yaml info.description の認証仕様(api.yaml:8-19): Cookie 4 種(access_token/refresh_token/prev_clinic_id/auth_token legacy)が auth_handler.go:19-21 の定数・middleware/auth.go:36-41 のフォールバック順序と正確に一致
- internal/apicontract の既存 date-format gate: fixture テスト(parser/analyzer/reconciler の 3 層 pin)・床値ガード(空振り fail)・stale allowlist 検出を備え設計健全。既知の drift 22 件 pin は指示どおり計画対象外として不変を確認(knownDateFormatDrifts 18 キー/22 箇所)
- CODING_RULES.md のエラーハンドリング章(3.1): internal/errors は package errors で `apperrors` エイリアス import(653 ファイルで確認)という記述前提と実態が一致。センチネルエラー・Wrap 系の記載も errors.go と整合
- ルート登録の構造健全性: 全 477 オペレーションが internal/handler 内の Register*Routes に完結し(cmd/api・middleware に直接ルート登録なし)、Group/verb が全て静的文字列リテラルで未解決 0 件 — 契約ゲートの静的検査が成立する良い前提条件
- 既知残存項目 KR-1〜6(pwreset タイムアウト/tz 直列化/Preload lint 未登録/TOCTOU #213/RLS #93/junction 非スコープ)はいずれも本次元(API契約・ドキュメント整合)の対象外であり、本次元内での新規状態変化は検出されなかった

### decomp

- treatment_service.go(668): CRUD部分は検証→tx→audit の一貫パターンで凝集。dose純ロジックは dose_calc.go/dose_revalidation.go/dose_validators.go に分離済み（残る保存時オーケストレーション3メソッドのみ finding td-treatment-dose-save-extract）
- reservation_service.go(530): keep — 予約CRUD+スロット競合検証は単一の予約確保責務。checkSlotConflict系ヘルパーはLIFF経路(reservation_validators.go)と共有する設計意図がコメントで明示済み(257-260行 errNoDoctorsOnDuty)
- accounting_report_service.go(514): keep — 月次集計+レスポンス構築で凝集。CSV出力ブロックのみ軽微な抽出候補(overflow参照)。レスポンス型がservice層に定義されるのは締めレジ(cash_register_service.go)と共有するための確立パターン(250-252行コメントに根拠)
- medicine_service.go(503): keep — 薬剤CRUD+連携在庫同期(BUG-320/429)+per_weight監査は単一ライフサイクル。tx統合の設計根拠コメント完備(407-408行)
- medical_record_repository.go(490): CRUD部分はclean(空clinicIDsフェイルセーフ・sort許可リスト・M-5ガードコメント完備)。Lstep集計群の同居のみ finding td-medrec-repo-owner-visit-split
- reservation_repository.go(489): keep — ReservationCRUDRepository/ReservationSlotRepository/ReservationQueryRepository の3サブインターフェース合成で既に模範的に整理済み(20-75行)
- cash_register_service.go(483): keep — 締めプレビュー/実行/期間解決は単一ドメイン。resolvePeriodRange/parseHHMM は#215越日EMGの中核でコメント完備
- lab_import_handler.go(482): keep — Request/Response/Conversion(P18)/Handlers/Routes のセクション構造が模範的。PII漏洩防止のerrorCategoryも根拠付き
- accounting_response.go(456)/accounting_request.go(404)/medical_record_request.go(351): keep — 純粋なDTO変換層。localTime/localTimePtr でtz統一済み
- billing_item_service.go(453)/trimming_service.go(439)/examination_service.go(437)/closing_settings_service.go(387)/appointment_notification_service.go(381)/medical_record_crud.go(377)/clinic_service.go(374)/pet_service.go(365)/lstep_tag_service.go(357)/checkup_field_result_service.go(357)/reservation_validators.go(348)/lstep_lifecycle_service.go(342): 各1エンティティ/1ユースケースで凝集 — keep
- timeslot_engine.go(354): keep — 区間演算による純粋なスロット生成エンジン。副作用なし・単一責務
- auth_handler.go(400)/accounting_handler.go(365)/reservation_type_handler.go(351): keep — ハンドラ薄層。auth は Login/Logout/Refresh/GetMe の単一認証フロー
- accounting_repository.go(349)/ltv_repository.go(343)/owner_repository.go(338)/reservation_staff_repository.go(330): keep — 単一エンティティのデータアクセスで凝集。accounting は reports系を accounting_repository_reports*.go に既に分離済み
- liff_service_availability.go(349): keep — liffService の空き日/空き時間解決に特化した部分ファイル(既にファイル分割イディオム適用済み)。filterSlotsByCapacity は R2-4 バッチ経路使用済み
- cmd/migrate/main.go(552): keep — advisory lock→baseline→SQL適用→seed投入の線形パイプライン。分割は手順の可読性をむしろ損なう
- cmd/stage-import/tables.go(354): keep — 1インポートパイプラインのテーブルメタデータ+SQL組み立てで凝集
- KR-2(tz直列化 FirstVisit/ExpiryDate/PaidAt)は解消済み: pet_response.go:70 localTimePtr(date) / inventory_response.go:38 localTimePtr(item.ExpiryDate)(commit f40446eb BE-refactor R2-1) / cash_register_service.go:280 d.PaidAt.In(time.Local).Format(time.RFC3339) / aggregation_handler.go:60 In(time.Local).Format("2006-01-02") — 全てローカルtz統一
- KR-4(LockAndFindByID/SumByBillingID の r.db 直参照TOCTOU)は解消済み: accounting_repository.go:176-178 が dbOrTx(ctx, r.db) 使用(R1-1/D2コメント付き)、refund_repository.go:53-56 SumByBillingID も dbOrTx 使用+設計根拠コメント完備。#213 は解消済み状態
- KR-5(RLS #93/exam_results複合FK)はDB基盤ギャップとしてIssue管理継続 — 本監査の対象外を確認のみ

### dup

- 既知残存項目2(tz直列化)は解消済み: FirstVisit は internal/handler/pet_response.go:70 `return petFirstVisitResponse{FirstVisitDate: localTimePtr(date)}`、ExpiryDate は internal/handler/inventory_response.go:38 `ExpiryDate: localTimePtr(item.ExpiryDate)`、PaidAt は internal/service/cash_register_service.go:280 `PaidAt: d.PaidAt.In(time.Local).Format(time.RFC3339)` — 3 フィールドとも local tz 統一済み。handler 層に他の PaidAt 直列化サイトなし(grep 全数確認)
- 既知残存項目4(#213 TOCTOU)は解消済み: internal/repository/accounting_repository.go:177 LockAndFindByID・refund_repository.go:54/67 SumByBillingID(+AndPaymentMethod) は `dbOrTx(ctx, r.db)` 化済みで R1-1/D2 の根拠コメント付き。dbortx_inventory_lint_test.go(311行)が revert を CI fail で防止する体制も確認
- repository Reorder 重複は折り畳み済み: helpers.go の reorderByClinicID/reorderGlobal を 24 リポジトリが使用。横展開コピーの残骸なし
- handler 層 master CRUD: parseIDParam/parseBindError/RespondError/mapSlice/reorderRequest/list_query_request が共有済み。`^func to` の重複定義は handler/service 全域でゼロ(uniq -c 全数検査)。各 master handler ~100-130 行は P5-P7/P12/P15/P18 規約準拠+godoc 都合の意図的個別定義であり追加折り畳み対象外
- service 層 master CRUD scaffold(List/GetByID/Create/Update/Delete/Reorder ×~20 サービス)は P1/P8/P10/P11/P13/P17 で規約化された意図的 boilerplate。animal_species(グローバル・petRepo 依存チェック)と chief_complaint(clinic スコープ・inquiry JOIN 依存チェック)の diff 検証で、シグネチャ・フィールド集合・削除依存対象・日本語 Conflict 文言が実差分と確認 — 逐語一致ではなく generics 折り畳みは master_fk_write_inventory_lint(exported method AST 走査)と greppability を毀損するため非提案
- lstep タグ同期の部品ヘルパは既に DRY: notifyAPIFailure/notifyAPISuccess/removeStaleTagsByPrefixes(lstep_tag_sync_api.go)・extractTagCodes/strSet/hasVaccineDeadlineSoon(lstep_health_tag_sync.go)・checkOptOut/shouldSkipSync(lstep_tag_sync_service.go:304)は抽出済み。残存重複は dup-lstep-tag-apply の 2 ブロックのみ
- checkup_sync_service_* / liff_service_availability_* / accounting_service_* の複数ファイル分割は単一サービスの責務分割であり、コピペ横展開ではないことを構成確認(重複関数なし)
- 既知残存項目6(permission_group/reservation_staff junction 非スコープ)は指示どおり報告対象外。既知残存項目5(RLS #93/exam_results 複合FK)は計画対象外のため未調査

### handler-p

- P12: ShouldBindJSON 全192箇所(非テスト256ファイル)を機械走査 — 全て `RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))` 準拠で逸脱0。err.Error()/gin.H 直返しのバインドエラー処理なし
- P18: 全変換関数走査 — convert/build/map/new プレフィックスのモデル→レスポンス変換関数0。buildAllPermissions(auth_me_response.go:109)は全権限マップ生成(直上doc確認済)、newPaginatedResponse(handler.go:39)は汎用エンベロープ、newXxxQuery 群はリクエストパーサでいずれも P18 対象外
- P7(findings以外): c.JSON の非 to* 引数を全て前後精読 — mapSlice(items, toXxxResponse)・typed struct・事前変換済み変数経由で準拠。無変換直返しは lab_report 2箇所と liff gin.H 1箇所(いずれも findings 掲載)のみ
- P5: 全ルート登録(44ファイルの Register/register 関数+関数外登録)を突合 — RequirePermission/perm ヘルパー欠落は免除カテゴリ(/health /login /logout /auth/* /me LIFF公開API webhook)のみ。billing_confirmation_handler.go:89 permEdit / daily_record_handler.go:220 permCreate / master_routes.go:14 と lstep_tag_config_handler.go:224 の perm クロージャも RequirePermission ラッパーであることを実読確認。PUT /users/me/password のみグレー(overflow)
- P6: DELETE ルート全走査 — findings の状態トグル4ルートと LIFF 公開 CancelLiffReservation(免除)以外は全て "delete" 権限
- P15: StatusCreated 75箇所の Location ヘッダ突合 — 欠落は line_link_handler.go:65 のみ(リンクトークンは GET 可能 URI を持たず Location 付与不能 → overflow)。Create*/Add* 系は 201+Location 準拠。200 返却の POST(PostLabImportPreview/PostLineSend/PostOwnerLstepTag/Upsert系/CreateCheckupSync)は全てアクション/プレビュー/Upsert 系で resource create ではない
- P14(リクエスト処理パス): handler メソッドからの repository 直接呼び出し0件。repository 型参照(accounting_response.go の集計 DTO 変換・medical_record_request.go:65 型エイリアス)は repo 呼び出しなし。h.repos は findings p14 で別掲
- KR-2(tz直列化 FirstVisit/ExpiryDate/PaidAt)は現HEADで解消済みを確認: FirstVisit=internal/handler/pet_response.go:69 `localTimePtr(date)`(直上に tz 表現割れ防止 doc)+aggregation_handler.go:60 `In(time.Local).Format`、ExpiryDate=internal/handler/inventory_response.go:38 `localTimePtr(item.ExpiryDate)`、PaidAt=internal/service/cash_register_service.go:280 `d.PaidAt.In(time.Local).Format(time.RFC3339)`。他の既知残存項目(KR1/3/4/5/6)は service/repository 次元のため本 Handler 次元では対象外(未再検証)

### idiom

- エラー体系の二重性: 無し。internal/errors 単一パッケージ（188行・センチネル7種+AppError+FromGORM/Wrap系18関数）で、全消費側は `apperrors "github.com/animal-ekarte/backend/internal/errors"` の alias import に100%統一（653 import、非aliasのimport 0件）。二重体系・命名不整合は存在しない
- context 伝播: 非テストコードの context.Background() は4箇所のみで、全て fire-and-forget goroutine 内 + //nolint:contextcheck + 根拠コメント付き（password_reset_service.go:99-102, appointment_notification_service.go:71-72,120-121）。context.WithoutCancel も2箇所とも根拠コメント付き（checkup_service.go:188-190「HTTP request の cancel から切離しつつ tracing context を保持」、lab_result_import_service.go:129 cleanup用5s）。context.TODO() は 0件
- 既知残存項目2（tz直列化未統一: FirstVisit/ExpiryDate/PaidAt）は現HEADで解消済み: FirstVisit → pet_response.go:70 `return petFirstVisitResponse{FirstVisitDate: localTimePtr(date)}`、ExpiryDate → inventory_response.go:38 `ExpiryDate: localTimePtr(item.ExpiryDate)`、PaidAt → cash_register_service.go:280 `PaidAt: d.PaidAt.In(time.Local).Format(time.RFC3339)` で全て local tz 直列化に統一済み
- 既知残存項目4（LockAndFindByID/SumByBillingID の r.db 直参照 TOCTOU = #213）は現HEADで解消済み: accounting_repository.go:176-178 は `err := dbOrTx(ctx, r.db).` に、refund_repository.go:53-55 SumByBillingID / :65-67 SumByBillingIDAndPaymentMethod も `dbOrTx(ctx, r.db)` に修正済みで、両所に「BE-refactor.md R1-1 (D2): dbOrTx で ambient tx に参加する」根拠コメントあり
- ステータス値・業務定数: reservation/billing/examination 等のステータスは model 層の typed constants（model/reservation.go ReservationStatus 等）に集約済みで、service/handler への文字列リテラル散在 0件。税率はマジックナンバーでなく clinic 単位の DB 値（StandardTaxRate/ReducedTaxRate）としてデータ駆動。クエリ Limit のハードコードは Limit(10000)×2 のみで問題なし
- hand-rolled ループ: contains/index系の重複ユーティリティ関数の再発明は無し（func contains/hasString/inSlice 該当 0件）。set 構築は map[T]struct{} イディオムで一貫。sort.Slice は2箇所のみ
- インターフェース設計: repository/service とも実装側パッケージ定義 + per-entity interface という P13/P16 規約に完全一貫しており、規約逸脱の定義場所ドリフトは無し。広いinterface（LstepTagSyncService 26メソッド、AccountingRepository 21メソッド）は handler向けfacade/集約repoとしての意図的設計で、分割は mock seam を壊す挙動非保存リスクに対し利得が薄いため提案せず（YAGNI）
- repository 共有ヘルパー基盤: base.go の clinicScope/clinicScopeIn/dbOrTx、helpers.go の reorderByClinicID/reorderGlobal は既に適切に共通化されており、ドキュメントコメントも整備済み

### perf

- repository層のループ内クエリ: 全非テストファイルをブレース深度追跡スキャンで走査し0件 — N+1はservice層の報告分のみ
- 一覧系ページネーション網羅: owners/pets/medical_records/reservations/vaccinations/examinations/hospitalizations/estimates/treatments/staffs/medicines/diagnosis/inventory/accounting(unpaid含む)/cash_register_closes/lab_import/line_send_log/lstep_csv_import 等は page/limit 実装済(repository 22ファイルで Limit 使用確認)。マスタ系 FindAll は Limit(10000) cap(vaccine_repository.go:37, procedure_repository.go:33)。無制限は checkups 一覧(finding: checkup-list-unbounded)のみ
- 集計系クエリ効率: accounting_repository_ltv.go(SumPaidByOwner/MaxSingleVisitAmountByOwner/FindOwnersByAnnualRevenue)はセットベースSQLで idx_billings_clinic_completed_at と整合しclean。GetDailySummary(reports_daily.go:27,40,54)は sargable 述語でclean。GetMonthlyReport/GetCloseAggregate は CTE+セット集計(各7/5クエリ固定)で Cartesian 積回避済 — 課題は述語形式のみ(finding: reports-completed-at-non-sargable)。FindCheckupSyncPreview はスカラー副問い合わせで cartesian inflation 回避済の単発集計SQL。cash_register_service.go の集計はメモリ内ループでclean
- LIFF slot容量チェック: filterSlotsByCapacity は R2-4 でバッチ化済を現HEADで確認(liff_service_availability.go:322-337 reservationTypeCapacityBatchCounter 経由 CountByTypeAndStartTimes)
- インデックス突合(001_init.sql 全 CREATE INDEX 抽出 vs 主要クエリ): appointments/medical_records/billings/hospitalizations の clinic+date/status 複合、billings(clinic_id, completed_at) partial(PERF-01)、audit_logs 3複合、shift_entries UNIQUE(clinic_id, staff_id, date)、clinic_holidays(clinic_id, date)、lstep_tag_cache(clinic_id, owner_id)、owners name/phone trgm、campaigns(clinic_id, start_date, end_date)、exams(clinic_id, exam_type_id, date) 等は充足。欠落は checkups/vaccinations の複合のみ(finding: checkups-vaccinations-missing-composite-index)
- 既知残存項目の再確認: (a) #213 LockAndFindByID/SumByBillingID の r.db 直参照TOCTOU → 解消済(accounting_repository.go:174-178・refund_repository.go:50-55/65-67 が dbOrTx 化、R1-1(D2)コメント付)。(b) tz直列化残(FirstVisit/ExpiryDate/PaidAt) → 解消済(pet_response.go:70 / inventory_response.go:38 が localTimePtr、cash_register_service.go:280 PaidAt は In(time.Local).Format(RFC3339) でJST統一)。(c) Preload lint 未登録3マスタ(MerchandiseItem/TrimmingCourseType/PaymentMethodMaster) → 現HEADで read 側 Preload 自体が存在せず、preload_clinic_scope_lint_test.go:79 と repository/CLAUDE.md に「write側のみ・将来Preload時は先にread側allowlist追加」の根拠付き先送りを確認(非該当・アクション不要)。(d) permission_group/reservation_staff junction 非スコープ=報告禁止対象を遵守、RLS(#93)/exam_results複合FK=計画対象外を遵守。現存はKR-pwreset-smtp-timeoutのみ
- GetDailySummaryForClinics(accounting_service_reports.go:151-158)のクリニック毎ループは拠点数(#86: 通常1〜3)スケールで意図的設計と判断 — N+1指摘対象外
- aggregation_service.go の Go側ページネーション(FindOwnerLTV全件取得)は「CPM段階判定をタグ同期側 CalculateCPMStage と共有する」設計判断がコメントで根拠化されており(checkup_sync_service_preview.go:48-49 同旨)、SQL移管は提案しない

### periph

- middleware fail-open/fail-closed一貫性: auth.go:82-83のstaff有効性チェックfail-openは既知(コメント根拠あり)。それ以外のauth/liff_auth/rate_limit/csrfは全てfail-closed(401/403/404/429/503で即Abort)を確認。liff_auth.goのLIFF_MOCKバイパスもgin.ReleaseMode除外+config.Validate(config.go:96-98)の二重ガード済み
- infra外部API呼出の一貫性: line/client.go(15sタイムアウト)とlstep/client.go(10sタイムアウト)は同型のdoWithRetry(429指数バックオフ最大3回・ctx.Done尊重・ボディDiscard+Close)、http.NewRequestWithContext、%wラップ、センチネルエラー(ErrInvalidRecipient/ErrUserNotFound=処理継続可能とドキュメント化)で統一。S3系(s3_file_storage/s3_uploader/s3_endpoint)はSDK ctxベース+エラーラップ+R2エンドポイント共有ヘルパで一貫。liff_auth.goのLINE verify APIも5sタイムアウト+ctx付き
- crypto/aes_gcm.go: 鍵長検証(32byte)・ランダムnonce・nonce前置・MaskValueヘルパまで清潔。ログ禁止コメントあり
- slog構造化キーの一貫性: internal全体でsnake_case統一を確認(clinic_id 552箇所/owner_id 83/staff_id 19/request_id/client_ip等)。camelCase揺れ・clinicID系のドリフト検出なし
- 既知残存項目2(tz直列化 FirstVisit/ExpiryDate/PaidAt)は現HEADで解消済み: pet_response.go:70(localTimePtr)・inventory_response.go:38(localTimePtr)・cash_register_service.go:280(.In(time.Local).Format(time.RFC3339))を確認 → cleanArea化
- config.Validate: dev用JWT_SECRET既定値の全モード禁止(config.go:78-80)・release時のJWT/DB_PASSWORD/INTEGRATION_ENCRYPTION_KEY/SMTP_PORT/LIFF_MOCK検証は堅実。timezone.goはtzdata埋め込みコメント付きでfail-fast
- logger/seedbundle/apicontract: いずれも小規模・ドキュメント充実(seedbundle/manifest.goのFK順序根拠コメント、apicontract/doc.goのパッケージ分離根拠)で負債なし
- rate_limit.go: getLimiterのTOCTOU回避(始めからWrite Lock)コメント付き・TTL eviction+ctxキャンセルでgoroutineリークなし。security_headers.go/cors.goはOWASPヘッダ+no-store+Originアロウリスト+Varyで整合
- handler.go(307行)のルート登録: RegisterRoutesはリソース別register関数へ適切に分割されP5権限ガードも一貫。DI規模の問題はmain.go側(finding two-phase-di-consolidation)に限定
- 既知残存項目3-6(Preload lint未登録3マスタ/LockAndFindByID TOCTOU #213/RLS #93/junction非スコープ)は本次元(middleware/infra/config/logger/main.go)の対象外につき未再監査

### repo-p

- KR-4 (#213 LockAndFindByID/SumBy* の r.db 直参照 TOCTOU) は現HEADで解消済みを確認: accounting_repository.go:176-185 LockAndFindByID は dbOrTx + R1-1(D2) 根拠コメント付き、refund_repository.go:31-74 の Create/SumByBillingID/SumByBillingIDAndPaymentMethod も全て dbOrTx、reservation_repository.go の LockAndFindByID/HasDoctorConflict/CountConflicts/CountByTypeAndStartTime(s) も dbOrTx 化済み
- service 層の WithTx 全27ブロック(14 service)を棚卸しし、accounting(core/correction/reports)・billing_item(Create/Update/Delete+recalculateTotals→UpdateBillingTotals)・checkup_field_result・examination(ReplaceItemsByExamID)・reservation(Create/updateWithConflictCheck/validators/admin)・medicine_dose_param・refund・staff(core/account: staff/account/assignment 3repo)・trimming(reservation+trimmingDetail) の呼出先 repo メソッドは全て dbOrTx 参加を実コードで確認。非参加は findings の medicine/inventory・clinic/permission_group・reservation_staff の3系統のみ
- P3/P3.1 (Preload 述語): 述語なし Preload を全件走査し、対象は全て soft-delete 非対象またはグローバルマスタ関連 (Refunds/RefundedByStaff は staffExemptAssoc、PaymentSplits/LineCustomer/Items(ExaminationItem)/Breaks/StaffNotes/Vitals/Inquiry/TargetCategories/TargetItems は gorm.DeletedAt 非保持、AnimalSpecies はグローバル)。clinic-scoped マスタの Preload は clinic_id 述語付きで、preload_clinic_scope_lint_test.go により機械強制済み
- P2 (Count* の deleted_at): 明示述語を持たない Count 系 8 メソッドを全件検証 — StaffClinicAssignment/InquiryTemplate 系は model が gorm.DeletedAt を持ち GORM 自動スコープで挙動上問題なし、inquiries・lstep 系ログテーブルは soft delete 自体が無く P2 適用外。挙動ギャップ 0 件
- P4 (clinicScope on write): 検査した write 経路 (medicine/inventory/reservation_staff/lstep_tag_code_mapping/permission_group/billing_item/refund/medicine_dose_param 等) は clinicScope・clinic_id JOIN・所有権検証のいずれかを具備。例外は文書化された 5 ファイル (clinic/company/account/password_reset_token/audit) と一致
- P9 (apperrors.FromGORM): 本監査で全読した約20 repo ファイルすべてで GORM エラーは FromGORM/Wrap 経由に統一されており、生 err return は 0 件
- audit_repository の Create/CreateTx 二本立て (#211) は設計コメントどおり健全に機能 (CreateTx は dbOrTx 参加・fail-closed 用途、Create は tx 外 best-effort 用途)
- lstep_tag_code_mapping_repository は *time.Time 手動 soft-delete だが read/SoftDelete 全経路に deleted_at IS NULL を明示しており漏れなし (Update の write 述語のみ overflow に記載)

### schema

- 優先テーブルの列存在・型・enum整合: billings/billing_items/payments/payment_splits/billing_refunds/billing_confirmations/estimates/estimate_items/appointments/medical_records/cash_register_closes/clinic_settings/closing_special_periods/payment_methods を model と逐語比較 — 列の過不足なし・numeric精度(quantity 10,1 / tax_rate 3,2 / discount_rate 5,2)・enum type タグ・部分UNIQUE(uq_cash_register_closes_date_period 等)全て一致。全102モデルの静的列diffでも実在の model-only/DDL-only 列はゼロ（AutoMigrate依存の隠れ列差分なし。非テストコードに AutoMigrate 呼出ゼロも確認）
- ソフトデリート整合の全量スキャン: gorm.DeletedAt を持つ全モデルの対応テーブルに deleted_at 列が存在（逆方向の欠落は SharedFile のみ=findings 報告。LstepTagCodeMapping は手動 SoftDelete 実装で整合・意図的）。PaymentSplit の deleted_at 無しは model コメント『delete-then-recreate パターンで管理する（soft-delete なし）』(accounting.go:172-173) と DDL が一致
- ON DELETE 方針: migration_cascade_lint_test.go の allowlist 001_init.sql=52 は実測 `grep -c "ON DELETE CASCADE"`=52 と一致し、旧010統合の経緯コメント(:44-47)も現状と整合。billings→clinics RESTRICT / payment_splits・billing_refunds→billings RESTRICT / billing_items・payments→billings CASCADE(純従属)は migrations/CLAUDE.md の許容例外定義どおり
- migrations/CLAUDE.md と現状の整合: 『.sql は 001_init.sql のみ』は ls 実測どおり、seeds は CSV 3バンドル構成どおり。001 ヘッダーの統合履歴(旧005-012を末尾セクション7に原文追記)は実体と一致 — #215 closing_am_start は :3629-3635 の ALTER として存在し model (clinic_settings.go:16) と型・NOT NULL・default '09:00' が一致
- audit_logs の actor 整合CHECK (DDL:2557-2562 staff⇔actor_id NOT NULL / system⇔NULL) は service 層 validateAuditLog (audit_service.go:106-117) と完全同型で二重防御が成立
- 既知残存項目2（tz直列化未統一 FirstVisit/ExpiryDate/PaidAt）は現HEADで解消済みを確認: FirstVisitDate は localTimePtr (handler/pet_response.go:70)、ExpiryDate/LastRestocked は localTimePtr (handler/inventory_response.go:38,40, R2-1 canonical 規約コメント付き)、PaidAt は `d.PaidAt.In(time.Local).Format(time.RFC3339)` (service/cash_register_service.go:280) — cleanArea として報告
- CI のスキーマ検証配線: ci.yml は 001_init.sql を psql 適用した ekarte_db に対し TestSchemaDrift を独立ステップで実行しており、drift テスト自体の実行環境は正しい（穴は findings の drift-test-coverage-gap のとおり検査内容側）
- clinic_settings の CPM/dormant/健診閾値 24 列は model の not null/default タグと DDL の NOT NULL DEFAULT+CHECK が全列一致（explicit-zero 教訓の反映を確認）

### service-p

- P1 (FindByID before Update/Delete): service層全66 Delete系+全Update系メソッドを機械列挙+個別Readで検証し違反0。permission_group_service.UpdateRules はハンドラ側 TASK-016 コメント付き clinic 所有確認 (handler/permission_group_handler.go:206-210 GetByID(clinicID,id)) でカバー済み。reservation_schedule.Delete は FindAllByDate 存在確認あり(ヒューリスティック偽陽性)。company.Update はシングルトンで更新後 FindSingleton が NotFound を表面化
- P8 (apperrors.Wrap): wrapcheck が CI ゲート済みで bare return err は0。service層の nolint 24箇所を全列挙し、全件に日本語の正当化コメント(意図的フォールバック/fire-and-forget/hugeParam統一等)が付随することを確認
- P10 (マスタDelete時FK依存チェック): マスタ系Delete全件棚卸し完了 — vaccine/medicine/consultation/procedure/exam_type/checkup_type/occupation/permission_group/cage/diagnosis_type(CountChildrenByParentID)/diagnosis_name/insurance/inventory/inquiry_template/hospitalization_plan/merchandise_item/payment_method_master/reservation_type(core:子+usage二重チェック)/reservation_type_group/reservation_type_liff(子+予約使用二重チェック)/reservation_staff/shift_template/trimming_course/trimming_option/trimming_course_type/chief_complaint/animal_species の27系統すべて CountUsageBy*/Exists チェック実装済み。campaign は末端マスタである旨のP10コメント付き(campaign_service.go:288)。staff.Delete は予約+シフト+CountBlockingReferences の3段チェック、owner.Delete はペット依存、clinic.DeleteClinic は owner/staff/CountBlockingReferences の3段チェックで全て409 Conflict返却
- P17 (Input命名): service層に *Params / *CreateRequest 型の入力構造体は0件
- P13: 195ファイル中191ファイルは定義順序準拠(違反4件はfinding参照)
- P11 のtx閉包内エラーは外側で一括 slog.ErrorContext 済みを個別確認(trimming_service.go:206 / hospitalization_service.go:349 / appointment_admin_service.go:152 / clinic_service.go:285 / billing_item_service.go DeleteItem閉包内 / accounting_service_core.go completeAccountingAppointments:254) — ヒューリスティック365ヒットの大半はバリデーション(WrapInvalidInput)・存在確認(NotFound)・tx外側ログ済みの除外対象で偽陽性
- 既知残存項目のうち本次元該当分の再確認: #1 password_reset smtp timeout は残存(KR- findingで報告)。#6 permission_group junction 非スコープは報告禁止指示に従い対象外(UpdateRulesのclinic確認はhandler TASK-016で実装済みであることのみ確認)

### silent

- KR-1(旧F-2)解消確認: internal/service/reservation_validators.go:281-321 — break_hours の unmarshal 失敗(283-286)・個別エントリ形式不正(310-318)とも apperrors.Wrap で fail-closed 化済み。D10/F-2 の根拠コメント付き。既知残存項目1は現HEADで解消済み
- 非テスト `_ =` エラー破棄 15箇所全数読解: infra/line・infra/lstep・middleware の HTTP body close/drain(復旧不可・大半 nolint:errcheck 根拠コメント付き)、handler/owner_handler.go:261 は BE-017 best-effort コメント+呼び先 HandleOwnerDeletion(lstep_lifecycle_service.go:245,254)が全失敗パスを slog.ErrorContext 済み。無ログ残存は lab import 補償遷移のみ(finding 報告済)
- service 層 goroutine 4箇所(password_reset_service.go:101 / appointment_notification_service.go:71,120 / checkup_service.go:191)全て内部でエラーを slog 記録し nolint 根拠コメント付き。appointment_notification は 15s・checkup followup は 35s の実効タイムアウト有り(タイムアウト非実効は password_reset のみ = KR- finding)
- `if err == nil` 14サイト全数読解: lstep_trigger_priority_service.go:97-108(NotFound→デフォルト優先度のコメント付き・非NotFoundはログ+返却)、lstep_settings_service.go:286-292(err は 292 で伝播)、middleware/auth.go:80-88(staff 有効性チェックの fail-open はコメント+WarnContext 付き)、liff_auth.go:52-64(LIFF_MOCK 開発専用・失敗時は実認証へフォールスルー)、medical_record_auto_create.go:31-48(BUG-386 best-effort 補完・失敗時は後段 44-48 で WarnContext+スキップ)、liff_service_availability_slots.go:26-36(LIFF-1/2 の slog 追加が現存)ほか — 実害ある無言 swallow なし
- err→continue バッチ群(lstep_batch_segmentation/dormant/noshow/delivery, lstep_tag_service, lstep_tag_sync_visit_ltv, aggregation_service): 全パスで slog.ErrorContext 記録+多くは errs aggregate で返却。fail-open の無言 continue なし
- 2026-06-30 監査以降の新規コード走査: 62c81d76(checked_in_at + infra/s3_*)・c80d9dc1(medical record ページネーション handler/repo/service)・7f967895(cash_register 税率分類)・d87cd6f2(estimate repo test)に swallow パターン検出なし。a620bdfc の AES-256-GCM 導入はレガシー平文 fallback を line_credentials.go:23-44 で設計文書化+WarnContext 記録済み(機密値非出力)
- repository 層の nil/0 フォールバック: campaign_repository.go:155・closing_special_period_repository.go:66 は not-found 表現としてコメント付き、inquiry_template_repository.go:87-90 の CountUsage スタブは PO 判断(2026-05-25)コメント付き。エラー隠蔽なし
- lab_result_import_service.go の終端遷移(172-177)は失敗を slog.ErrorContext+根拠コメント付きで処理済み(補償側のみ finding)

### tenant

- medical_record_repository.go の新規 JOIN ベース FindAll（B-1 server-side pagination/search, c80d9dc1）: medical_records.clinic_id を明示 WHERE で固定し pets/owners への LEFT JOIN は既に clinic-scoped な親レコードの FK 経由のみ。TestMedicalRecordRepository_FindAll_ClinicIsolation / _Search で他院データ非混入を確認済み（medical_record_repository_test.go:121-219）
- checkup_field_repository.go / checkup_field_result_service.go（#211 健診パッケージ, a60d8b8c）: FindByCheckupID/FindByPetID の Preload("CheckupTypeField", "clinic_id = ? AND deleted_at IS NULL", clinicID) は preload_clinic_scope_lint_test.go に登録済み（clinicScopedMasterAssoc["CheckupTypeField"]）。ReplaceForCheckup は #124 同型ガード（checkup_type_field_id の所属検証）+ tx内 fail-closed 監査（AuditTxLogger.LogEntryTx）を実装
- accounting_service_correction.go の CorrectCreditPayment（#189）: LockAndFindByID(clinicID)+FindByID(clinicID) でクリニックスコープ保持、監査は同一tx内 fail-closed（logCreditCorrection→LogEntryTx）
- raw SQL / JOIN サイト全面確認（medical_record_repository.go:410-460 FindOwnersByLastVisitDays/FindOwnersByNextVisitRecommended, billing_item_repository.go:141-164 FindOwnersByCategoryPurchaseDate, ltv_repository.go 集計クエリ, checkup_sync_repository.go プレビュークエリ, pet_chronic_condition_repository.go GetActiveConditionCodesByOwner, reservation_type_occupation_repository.go CountWorkingStaffByReservationTypeID, lstep_sync_error_counter_repository.go IncrementFailure）: 全件で clinic_id が明示的にWHERE/JOIN述語に含まれることを確認
- b69932eb checkpoint（BE-refactor.md R1-1/R2-5, D2/D12）で修正済みの write-side バグの現存確認: CountItemsByEstimateID/CountMedicalRecordsByReservationID への clinic_id 述語追加、lab_import_repository.go/lstep_csv_import_repository.go の GORM Save() が UPDATE の Where/Scopes を無視する既知の罠を Model+Where+Updates(map) に是正、care_plan_item_repository.go/clinical_plan_repository.go の Joins()がUPDATE/DELETE SQLに伝播しないGORM挙動をサブクエリ形へ是正 — いずれも現HEADで是正済み状態を確認
- staffService.SetExcludedReservationTypeIDs/SetCapableReservationTypeIDs（master_fk_write_inventory_lint_test.go の静的解析対象外＝裸スカラparam）: reservation_staff_repository.go の UpdateExcludedReservationTypes/UpdateReservationCapabilities で clinic_id スコープの ReservationType 所有権を Count検証してから書込む実装を確認（自動lintの守備範囲外だが実装は健全）
- medicineDoseParamService.Upsert（同じく静的解析対象外）: medRepo.FindByID(ctx, clinicID, medicineID) で親medicineの所有権検証済み
- permission_group_repository.go FindAllGroupIDsByStaffID の意図的 clinic_id 非スコープ（b69932eb コメント根拠: permission_groups.id はグローバル一意PKでありスコープしても判定結果不変）を現HEADで再確認・現存し妥当

### test

- coverage.out集計の前提: backend/coverage.out(2026-07-01 15:55生成・9,214計測行)はinternal/serviceパッケージ単独のプロファイル(集計で全行がinternal/service配下・service計12,971stmt中4,747未実行=63.4%)であり、CIの89.9%baseline(-coverpkg=./internal/... 横断計測、ci.yml:399)とは計測方式・時点とも別物。さらに7/3のb69932ebでservice層テスト168ファイル(新規約72本)が一括追加されたため、7/1プロファイルのファイル別0%ホットスポット(password_reset_service.go 0%/reservation_schedule_service.go 0%/manual_article_service.go 0%/lstep_tag_sync_visit_ltv.go 0%/lstep_batch_delivery.go 0%/lstep_settings_connection.go 0%等)は大半失効 — 各ファイルに対応する*_test.goが現存し、サンプル読解(password_reset_service_test.go 301行5テスト17アサーション、lstep_settings_update_test.go 663行9テスト、reservation_schedule_service_test.go 293行4テスト)で実質的アサーションを確認済み。空洞テストではない
- 既知残存項目2(tz直列化 FirstVisit/ExpiryDate/PaidAt)は解消済み: ExpiryDate=localTimePtr(handler/inventory_response.go:38)、FirstVisit=localTimePtr(handler/pet_response.go:70)およびIn(time.Local).Format(handler/aggregation_handler.go:60)、PaidAt=In(time.Local).Format(time.RFC3339)(service/cash_register_service.go:280)
- 既知残存項目4(LockAndFindByID/SumByBillingIDのr.db直参照TOCTOU)は現HEADでdbOrTx化済み: accounting_repository.go:174-178「BE-refactor.md R1-1 (D2): dbOrTx で ambient tx に参加する」+ dbOrTx(ctx, r.db)、refund_repository.go:51-55/67も同様にdbOrTx化・コメント付き。#213の残余はIssue管理
- 締めドメインのrepository統合テストは充実: accounting_repository_reports_{close,daily,monthly}_test.go(GetCloseAggregate/GetMonthlyReport実DB実行、close 5テスト/monthly 6テスト)、cash_register_close_repository_test.go、closing_special_period_repository_test.go、daily_record_repository_test.goが存在
- 投薬ドメインのrepositoryテスト充実: medicine_repository_test.go(8テスト)、prescription_repository_test.go(6テスト)、medicine_dose_param_repository_test.go+medicine_dose_param_clinic_isolation_test.goが存在
- クリニック分離テストのインフラは体系化済み: repository層133テストファイル中80がクロステナントパターンを参照、専用isolation test 15ファイル+(read側)preload_clinic_scope_lint_test.goのgo/ast機械強制+(service側)master_fk_write_inventory_lint_test.goの双方向名簿突合。未カバーの残りは findings の capability write / known-unguarded 群に集約
- baseline(7/3 arm)以降のbackendコミットは全てテスト同梱: c80d9dc1(medical-recordsページネーション/ソート — repositoryテスト582行+handler/serviceテスト追加)、62c81d76(checked_in_at+R2移行 — config_validate/s3_endpoint/reservation_responseテスト追加)、d87cd6f2(#212 estimate_repository_test 227行追加)。baseline未反映だが新規テスト負債は形成していない
- reservation_service.go(530行・同名テスト無し)はappointment_service_test.goが実体テスト(NewReservationService群を20箇所以上で構築しCreate/Update/競合検証を実行)。同名テストファイル不在=未テストではない
- mock過剰次元の総括: service層はmockリポジトリ+repository層は実DB(setupTestDB)という層別方針が一貫しており、7/3一括追加分を含めframework-onlyの空洞テストは検出されず(過去監査の「framework-only test 0件」判定と整合)。mockがSQL実挙動を隠している実害箇所は findings のF1(実効権限)/F2(会計完了化)/F3(トリミング未請求)/F7(Lstep購買)の repository メソッド群に限定される
- handler層はテストファイル202本/非テスト256ファイルで、リクエストバインド・レスポンス変換の主要経路をカバー(medical_record_request_test.go等の直近追加も確認)。accounting_repository_reports.go(101行)は型定義のみでテスト対象ステートメント無し

### tx

- KR-4 解消確認: accounting_repository.go:176-185 LockAndFindByID は dbOrTx + clause.Locking{UPDATE}（R1-1 D2 コメント付き）、refund_repository.go:53-63 SumByBillingID / 65-75 SumByBillingIDAndPaymentMethod も dbOrTx 参加済み。refund_service.go:46-131 は WithTx + FOR UPDATE + tx 内残額再計算 + fail-closed 監査で #213 の r.db 直参照 TOCTOU は現HEADに存在しない
- KR-2 解消確認: FirstVisit=pet_response.go:70 localTimePtr、ExpiryDate=inventory_response.go:38 localTimePtr、PaidAt=cash_register_service.go:280 `d.PaidAt.In(time.Local).Format(time.RFC3339)` — tz 直列化未統一 3 フィールドは全て canonical 規約に統一済み
- #211 checkup 横断 follow-up 完遂確認: audit_tx_inventory_lint_test.go の allowlist は checkup_field_repository.ReplaceForCheckup / examination_repository.ReplaceItemsByExamID の両臨床結果 hard-delete サイトとも status=audited-tx-internal（runtime 証明 = checkup_field_result_tx_atomicity_test.go / examination_repository_tx_atomicity_test.go）。pending-migration エントリはゼロ
- audit_logs 書込呼出の全数分類（service/handler/middleware 全 40+ サイト列挙済み）: tx 内 fail-closed = 9 系統（refund_service.go:113、accounting_service_correction.go:196、accounting_service_core.go:220 logPostCloseEdit、accounting_service_reports.go:97 Cancel、examination_service.go:345、checkup_field_result_service.go:222、medicine_service.go:336 per_weight有効化、medicine_dose_param_service.go:218、treatment_service.go:641 txRepos.Audit.CreateTx 直書き）— いずれも txCtx/txRepos 経由で ambient tx 参加を実コードで確認。残る全サイト（medical_record_crud/vital/addendum の AUDIT-H1 系・lab_audit_logger.logBestEffort・LSTEP 系・auth/permission_group/manual_article handler・clinic switch middleware）は「best-effort」コメント付きの意図的 warn-and-continue で、対象がソフトデリート領域または外部 API 同期ログのため #211 の fail-closed 対象外。方針からの逸脱サイトはゼロ
- 2 つの tx 機構（Transactor.WithTx/ctx-txKey と repos.Transaction/txRepos）の併存は treatment_service.go:142-148 で機構間非互換（LogEntryTx が txRepos の tx に参加できない）まで文書化済みで、実装も txRepos.Audit.CreateTx で正しく回避している
- dbOrTx 参加 surface は dbortx_inventory_lint_test.go で 80 メソッドを双方向 pin（regression で CI fail）。billing_item 明細+合計再計算（billing_item_repository.go:38-44 コメント）、SavePayment/SavePaymentSplits/Update（accounting_repository.go R1-1/R1-2 コメント）の money-path は dbOrTx 統一済み
- レジ締めの二重締め check-then-act（cash_register_service.go:319-326 FindByDateAndPeriod → 357 Create、tx 外）は migrations/001_init.sql:2142 `uq_cash_register_closes_date_period (clinic_id, close_date, period) WHERE deleted_at IS NULL` が DB バックストップし、FromGORM(errors.go:178) が 23505→WrapAlreadyExists に変換するため競合負け側も 5xx にならない — 追加対応不要
- treatment 作成は repos.Transaction 内で treatment INSERT + 在庫減算(DecreaseStock) + 投与量逸脱監査を単一 tx 原子化済み（treatment_service.go:242-303）
- accounting_service_core.Update の fields/payment/splits/締め後編集監査は R1-2 (D1) で単一 WithTx に統合済み（accounting_service_core.go:144-182）— 残余は本 findings の completeAccountingAppointments 後段のみ
- カルテ本体の finalize 後編集は medical_record_repository.go:240 `Where("id = ? AND status = ?", id, model.MedicalRecordStatusDraft)` により DB レベルで原子的に遮断済み（check-then-act に依存しない）## Appendix D: 閾値未満の観察事項（要対応判断は個別に）


### cmdtools

- backend/.gitignore:1 のコメント「compiled binary (Air hot-reload output)」が stale — .air.toml:6 は `go build -o ./tmp/main ./cmd/api` で tmp/ に出力し /api は手動ビルド由来。finding tracked-migrate-binary の修正時にコメントも訂正
- backend/cmd/seed-old-db/main.go:1-14 のパッケージ doc に Deprecated 表記なし — cmd/stage-import/main.go:5 が deprecated と宣言しているため、seed-old-db 側 doc 冒頭にも「Deprecated: cmd/stage-import が後継。old_db 本番移行完了までの fallback として残置」を追記すべき
- worker-configuration.d.ts がリポジトリルートに git 追跡で置かれ、生成ヘッダに開発者ローカルの絶対パス（worker-configuration.d.ts:2 `--config=/Users/minoru/Dev/...`）を含む — wrangler types の出力先を backend/ 配下に固定するか再生成手順を README 化
- backend/FK_DEPENDENCY_CHECK_ROADMAP.md (2026-04・git 追跡・backend 直下) — 完了/中断状態の roadmap 文書が backend ルートに残置。docs/ へ移設または docs/archive/ 行き判定
- cmd/migrate/csvbundle.go:158 quoteLiteral と cmd/stage-import/main.go:297 pqLiteral の同一3行実装 — finding cmd-dsn-localhosts-duplication の手順(3)で任意対応
- backend/README.md のディレクトリ構造セクション（49行目含む）が現行 cmd/ 7ツール構成・Dockerfile 3種の実態と広く乖離 — finding stale-root-dockerfile 対応時に一括更新

### contract

- docs/api.yaml:7195 `/clinics/{id}` のパラメータ名が実装 clinic_handler.go の `:clinic_id` と不一致(shape は一致・ドキュメント上の表記ゆれのみ)
- docs/api.yaml:10913 `/health` は servers prefix /api/v1 配下として記述されるが、実装は handler.go:52 `r.GET("/health", h.Health)` でルート直下 — 文書上のパスが /api/v1/health になってしまう(servers を跨ぐ例外として明記するか paths から外す)
- backend/README.md:8 Gin v1.10 と記載、go.mod 実測は gin-gonic/gin v1.12.0(finding onboarding-docs 内の手順(4)でカバー済みだが単独でも直せる)
- docs/README.md:5-9 のファイル構成一覧が postman-collection.json / api-examples.md を列挙していない(stale-prototype-example-docs の削除を採用すれば自然解消)

### decomp

- internal/service/accounting_report_service.go:411-497 — CSV出力(ExportMonthlyCSV/ExportMonthlyCSVByPeriod/buildMonthlyCSV 約90行)は集計と別の出力形式責務。accounting_report_csv.go への抽出候補(軽微・任意)
- internal/service/medicine_service.go:152-156 — 孤立コメントブロック「--- DB column constants ---」と buildMedicineUpdate の説明文が実体(35行)から分離して残置。削除または実体直上へ移動のみ
- internal/service/treatment_service.go:70-72 — セクションヘッダ「─── Interface ───」の直下に buildTreatmentUpdate(build関数)が置かれヘッダと内容が不一致(P13順序自体は満たすがヘッダ位置ずれ)
- internal/service/diagnosis_service.go — DiagnosisTypeService と DiagnosisNameService の2サービス同居(382行)。密結合マスタペアであり分割価値は低いが、成長時は diagnosis_type_service.go/diagnosis_name_service.go ペア分割可
- internal/service/lstep_tag_sync_service.go:73-146 — 純粋関数 CalculateCPMStage/CalculateCPMStageV2 が同期サービスと同居。cpm_stage.go への抽出候補(軽微)

### dup

- internal/repository/*: `Order("sort_order ASC, name ASC")` リテラルが 27 箇所 — 共有 const 化可能だが独立変更しても実害がなく軽微
- internal/handler/*_request.go: `time.ParseInLocation("2006-01-02", x, time.Local)` インライン ~20 箇所(appointment/campaign/checkup/checkup_sync 等) — 単一フォーマット query date parse で stdlib 1 行呼び出し。accounting_report は monthlyReportJST と location が意味的に異なるため一律折り畳み不可。dup-handler-date-parse 実施時に parseFlexibleDate と並べて任意で整理
- internal/service/lstep_health_tag_sync_prevention.go: SyncFilariaTag の犬保有判定ループ(strings.Contains(name,"犬"))は checkup 側に類似ロジックが将来出現しうるが現状 1 箇所のみで抽出対象外(YAGNI)
- internal/repository/billing_confirmation_repository.go:32 / clinical_plan_repository.go:35 / treatment_repository.go:51 等の clinic_id 述語なし medical_records JOIN 変種は別述語で隔離される設計(2026-06-29 クロステナント read 監査済み)のため dup-medrecord-tenant-join の対象から除外した

### handler-p

- internal/handler/auth_handler.go:179,324 / auth_password_handler.go:69,96,118 — 認証フローの成功レスポンスが gin.H{"message": ...}(モデル非含有)。共通 messageResponse struct への置換余地(微小・wire同一)
- internal/handler/handler.go:81 — PUT /users/me/password (ChangeMyPassword) に RequirePermission なし。自己スコープ操作で /me と同類だが P5 免除リスト(refs)に未記載。免除リストへの明記推奨
- internal/handler/line_link_handler.go:65 — GenerateLineLinkToken が 201 を Location なしで返す。トークンは GET 可能 URI を持たないため Location 付与は実質不能。免除根拠コメント追記のみ推奨
- internal/handler/vital_handler.go:141 (POST vitals=edit) / medical_record_addendum_handler.go:15 (POST addenda=edit) vs daily_record_handler.go:222 (POST daily-records=create) — サブリソース作成 POST の権限アクションが edit/create で不統一。権限ポリシー論点でありリファクタ対象外
- internal/handler/lstep_tag_handler.go:74,102 — PostOwnerLstepTag(タグ付与)が 200 返却。冪等 add セマンティクスのため P15 適用は弱い(現状維持可)
- internal/handler/checkup_sync_handler.go:136 — CreateCheckupSync(POST 一括タグ付与アクション)は 200 返却。resource create でないため P15 対象外だが Create* 命名がスキャナ偽陽性源
- internal/handler/reservation_staff_handler.go:144-147 等 — POST 画像アップロード2件が WrapNotImplemented(501) スタブのまま create 権限付きで登録済み(「v2 スコープ：未実装」コメントあり)。ルート露出の要否棚卸し候補
- internal/handler/lab_report_handler.go — model 層 response DTO 直返しは findings p7-lab-report 掲載。関連して medical_record_request.go:65 の `type listMedicalRecordFilters = repository.MedicalRecordListFilters` は handler→repository 型エイリアス直結(低優先)

### idiom

- internal/service/lstep_tag_code_mapping_service.go:94 isConfigurableTag と internal/service/lab_import_service.go:38 CanTransitionTo の内側ループは slices.Contains に、internal/service/liff_service_availability_delegate.go:70 isCapable と internal/service/lstep_delivery_trigger_state.go:26 のタグ検索は slices.ContainsFunc に置換可能（計4-6箇所・可読性向上は小幅）
- internal/service/liff_service_availability.go:283 と internal/service/timeslot_engine.go:152 の sort.Slice は slices.SortFunc へ置換可能（2箇所のみ・利得軽微）
- internal/errors のパッケージ名が stdlib errors と衝突するため全377ファイルで apperrors alias を強制している。ディレクトリごと internal/apperrors へ rename すれば alias 不要になるが、機械的とはいえ約380ファイルの import 差分に対し利得は美観のみ（現状 alias は100%統一済みで実害なし）
- internal/service/available_dates.go:221 の LoadLocation 失敗時フォールバック time.FixedZone("JST", 9*60*60) は tzdata 埋め込み（config/timezone.go:6）により到達不能な防御コード — jst-location-derivation-scatter 実施時に消える

### perf

- internal/service/clinical_plan_service.go:74,81 — 診断行ごとの diagTypeRepo/diagNameRepo.FindByID 検証ループ。N=1カルテの診断数(数件)で書込パスのため実害小。FindByIDs バッチ化は任意
- internal/service/lstep_tag_sync_visit_ltv.go:69 ほか lstep タグ同期バッチ群の per-owner tagCacheRepo.FindByOwner — 外部Lstep API律速のバッチジョブ。checkup-sync-preview-tagcache-nplus1 で導入する FindByOwners の流用先候補
- internal/service/lstep_delivery_trigger_suppression.go:37,47 — GetPriorityFor が呼出毎に lstep_trigger_priorities をDB照会(極小テーブル)。配信バッチ起動時にクリニック単位で一括ロードすれば削減可
- internal/service/aggregation_service.go:260-279 — SyncAggregationTags の per-owner UpsertTag 書込ループ。バッチUPSERT化可能だが管理操作で頻度低
- internal/repository/owner_repository.go:70-79 — 検索の translate(name_kana,...) ILIKE '%...%' は非sargable(name/phone は trgm index 有、name_kana 式・email は無)。検索遅延が実測されたら式インデックス gin(translate(name_kana, src, tgt) gin_trgm_ops) + email trgm を検討(4分岐ORのBitmapOr成立には全分岐のindexが必要)
- internal/repository/checkup_repository.go:110-115 — FindAlerts の4連Preload(MedicalRecord.Pet.Owner まで)。checkup-list-unbounded の是正と併せて必要列のJOIN射影化を検討
- internal/repository/appointment_admin_repository.go:33-50 — FindAllByMonth も月次で3 Preload 全件取得だが月範囲で有界・管理画面月表示用途のため許容

### periph

- internal/errors/errors.go:170-183(「数値が範囲外です」等の日本語)とinternal/middleware/auth.go:58等(英語)でユーザー向けエラーメッセージの日英混在 — 統一言語はプロダクト判断待ちのため計画化見送り
- cmd/api/main.go:35 起動ログ「starting Animal Ekarte API v2.0 (45 tables)」がseed対象90テーブルの実態と乖離した固定文字列
- cmd/api/main.go:38,45,53,271,279 でlogger.*ラッパとslog.*直呼びが混用(出力は同一slogハンドラのため実害なし)
- internal/middleware/logging.go:69-73 slog.Log向けのattrs→args手動変換はslog.LogAttrsで置換可能(微細)
- internal/middleware/liff_auth.go:145 verifyLiffIDTokenが呼び出し毎にhttp.Clientを生成しコネクション再利用されない(LIFF認証は低頻度のため実害小)
- internal/middleware/sanitize_null_bytes.go:38 io.ReadAllが無制限でMaxBytesReaderなし(sanitize-multipart-binary-corruptionのContent-Typeガード導入でバイナリ大容量は対象外化されるため残リスクはJSONのみ)
- internal/middleware/auth.go:78-79 staffID parse失敗時(claims.UserID非数値)にstaff有効性チェックを黙ってスキップ — JWT自署名前提で実害なしだがfail-open系として設計メモ化推奨

### repo-p

- internal/service/shift_entry_service.go:179-192,238-252 — Create/Update が entry 書込と ReplaceBreaks を tx で括っておらず非原子 (breaks 失敗時に休憩なし entry が残る)。原子性の主張コメントは無いため tx 非参加バグではなく tx 境界の欠落。挙動変更トラックで WithTx 化を検討 (shift_entry_repository は dbOrTx 採用済みのため service 側 1 行の変更で済む)
- internal/repository/lstep_tag_code_mapping_repository.go:63-67 — Update の Updates() WHERE に deleted_at IS NULL が無く soft-deleted 行の fields を書き換えうる (直後の reload :74 は NotFound を返すため API 上はエラーだが書込は残る)
- internal/repository/examination_repository.go (Create/Update/Delete = r.db) / reservation_type_repository.go (Update/Delete/Reorder = r.db) / campaign_repository.go / accounting_repository.go:198-207 Create — ファイル内で dbOrTx と r.db が混在。ambient tx からの呼出は現存せず実害なしのため tx-mechanism-consolidation の一括置換で解消
- internal/repository/helpers.go:14,36 — reorderByClinicID/reorderGlobal は db を引数で受け内部 tx を張る第5のマイクロ機構。ambient tx 呼出なしで実害なし。dbOrTx(ctx, db).Transaction 化は consolidation に含める
- internal/service/accounting_service_core.go:195-199 — completeAccountingAppointments が Update の tx 外実行 (billing 完了確定後に予約完了同期が失敗すると billing=completed/予約=未完了の不整合)。意図的の可能性が高いが根拠コメントが無いため意図の明文化を推奨
- internal/repository/reservation_staff_repository.go:210-224 — UpdateExcludedReservationTypes の所有権検証 read が内部 tx 外 (r.db) で軽微な TOCTOU 窓。tx-reservation-staff-nonparticipation の修正時に tx 内へ移動

### schema

- internal/repository/ltv_repository_test.go:1389 コメント『001_init.sql の 46 型 + 009 の 4 型』は実数54型と不一致（test-enum-copy-drift の手順(4)で同時修正）
- internal/model/schema_drift_test.go:20-27 buildDSN のデフォルトが dev 用 ekarte_db を指す一方、repository テストは ekarte_db_test を使う二重構造 — 意図（drift は実DDL、統合テストは AutoMigrate 箱庭）が読み取りにくいのでファイル冒頭に関係図コメントを追記する価値あり
- model の not null タグ省略が広範（例: accounting.go:76 Billing.Memo `default:''` vs DDL:1882 `memo text NOT NULL DEFAULT ''`、reservation.go:50 Notes 同様）— 非pointer フィールドのため実行時リスクはゼロだが、AutoMigrate 製テストスキーマの NOT NULL 忠実度を下げる。drift-test-coverage-gap の nullability 検査導入時に一括棚卸しが可能
- 001_init.sql の lstep_migration_progress テーブル(441-)にはアプリ model が存在しない（lstep-migrate cmd 専用）— schema_drift_test の『DBにあるがモデルに無いテーブル』検査は元々テーブル単位では行っていないため実害なしだが、exhaustiveness 追加時に対象外リストとして明文化すると良い

### service-p

- internal/service/lstep_tag_config_service.go:84,114,144 — Delete系3メソッドがservice層FindByID前置なし(P1名目逸脱)。ただしrepo側(lstep_tag_config_repository.go:55-57等)が RowsAffected==0 で ErrRecordNotFound を返すため404挙動は担保済み。実害なし
- internal/service/audit_service.go:130,149 — repo.Create/CreateTx エラーのwrap前にslogなし(P11名目)。全呼び出し元が slog.WarnContext で監査失敗を記録しているため実質カバー済み
- internal/service/lstep_delivery_trigger_state.go:37,52 — trigger log check/create のエラーパスにslogなし(P11候補、バッチ呼び出し元の外側ログ有無は未検証)
- internal/service/owner_service_delivery.go:13,54,89 / owner_service_line.go:14,42 — Get系読取のFindByIDエラーwrapにslogなし。NotFound存在確認の除外規定との境界がグレー(P11低優先)
- internal/service/lstep_settings_service.go:105-126 — service層が LstepSettingsResponse(マスク済みDTO)を直接返却しhandlerのtoXxxResponse変換(P7)を経由しない。credentialマスキングがビジネスロジックである点で正当化可能だが層責務のグレーゾーン

### silent

- internal/service/owner_service_line.go:19-22 — Q22 Guard1 の FindByLineUserID 非NotFoundエラーが無ログ fail-open(err==nil 判定)。ただし DB の部分 unique index uk_owners_clinic_line_user_id(migrations/001_init.sql:340-342)が実整合を担保し、owner_service_line_test.go:76 が「errors treated as not-found」として挙動固定済のため実害なし。IsNotFound 分岐+slog 追加のみ検討価値(テスト書換必要=behaviorChange)
- internal/service/liff_service_reservations.go:26-32 — 指名なし委譲で toDateTime エラーを無ログ swallow(err==nil 時のみ委譲)。不正時刻は後段 ValidateAndCreate(reservation_validators.go:295-302)が InvalidInput で拒否するため実害なし。delegateStaff 側の失敗は内部で slog 済(liff_service_availability_delegate.go:17)
- internal/middleware/auth.go:78-79 — claims.UserID の ParseUint 失敗時に staff 有効性チェックを無言スキップ。UserID は自前署名 JWT 由来の数値文字列のため実質到達不可だが、スキップ時の slog が無い
- internal/middleware/logging.go:84-88 — rand.Read エラー無視により失敗時 request ID が "00000000" に退化(観測性のみ・無害)
- internal/service/medical_record_builders.go:25-30 — カルテ番号乱数の crypto/rand 失敗時 charset[0] フォールバック。根拠コメント有だが無ログ(発生は実質皆無)
- internal/service/liff_service_availability.go:274-277 — 不正形式 available_slot 開始時刻の無言 drop。同型の 309-312 行には「既存挙動」コメント有り、276行側はコメント/ログ無し
- internal/service/lstep_delivery_trigger_suppression.go:43-71 — bestLog は代入されるがどこでも未使用で `_ = bestLog` により compile を黙らせている。コメント「used only for logging context above」は不正確(デッド変数)。bestPri のみ残して bestLog を削除可能(挙動保存)

### tenant

- internal/service/master_fk_write_inventory_lint_test.go の masterFKWriteAllowlist には他に約40件の statusKnownUnguarded エントリ（carePlanItemService.HospitalizationPlanID/treatmentService.InventoryID/medicineService.ParentID/petService.InsuranceID/hospitalizationService.CageID/staffService.OccupationID等、自己参照parent_id系多数）が既に自己文書化された残存P1として存在（39921f98で作成、本レビュー期間より前）。全件は個別findingとして重複報告しない — 同ファイルが正本の作業リストとして機能している

### test

- internal/repository/accounting_repository.go:107 attachRefundTotals — FindAll経由でisolationテストから間接実行されるが、返金合算値そのもの(複数refund合算・対象外billingの0埋め)を直接検証するDBテストは無い。F2/F3対応時に同ファイルへ1テスト追加が安価
- internal/repository/reservation_staff_repository.go:129 UpdateSortOrder — repositoryテスト1件のみ参照。direction文字列ベースのスワップロジックだが優先度低
- internal/service/liff_service_availability.go(349行)+liff_service_availability_business.go — 同名テスト無しだが分割先(_delegate/_filters/_slots/_staff/_time)に各テスト有り。ブリッジ残余の直接テストは低優先
- backend/coverage.out(ローカル・gitignore済み)が7/1のservice単独プロファイルのまま残置 — 次回監査の誤誘導防止のためCI準拠(go test ./... -coverpkg=./internal/...)で再生成するか削除を推奨(ファイル操作のみ・コード変更なし)
- internal/repository/permission_group_repository.go:127 CountUsageByGroupID — 削除前FK依存チェック(P10)の実体だがrepositoryテスト0件。F1のテストファイルに1ケース同乗可能

### tx

- internal/service/accounting_service_reports.go:69-76 — Cancel の二重キャンセルチェック(FindByID→status 照合)が WithTx 外の check-then-act。同時キャンセルでは両者とも tx に入り status=cancelled を冪等に二重書込 + 監査 2 重記録（old_value も stale になりうる）。結果整合は保たれるため軽微。tx 内で Update の RowsAffected/事前 status を再検証すれば閉じる
- internal/service/accounting_service_core.go:93 — IsPostClose フラグは呼出元(handler)の HasCloseOnDate 判定に依存する check-then-act で、判定と会計更新の間に締めが入ると post_close 監査フラグが欠落する（監査 metadata の欠落のみ・業務書込は無傷）
- internal/repository/care_plan_item_repository.go:85-92 — CarePlanItem は DeletedAt 無しの hard-delete だが監査カバレッジ未整備。audit_tx_inventory_lint_test.go:59-61 に『Separate audit coverage tracked as follow-up』と文書化済みの既知先送り（新発見ではない）
- internal/service/treatment_service.go:279-291 — InventoryID 未指定かつ MedicineID 指定時に `targetInvID = *input.MedicineID` として medicine の ID をそのまま inventory_items.id として DecreaseStock に渡す。ID 体系混用の疑い（medicines.inventory_id 経由で解決すべきに見える）。tx 次元外のため正しさ次元での要検証事項として送る
- internal/repository/inventory_repository.go:103-115 — DecreaseStock は `int(quantity)` で float64 を切り捨て、負在庫ガードも clinic_id スコープも無い（呼出元 treatment tx 内で親検証済みだが repo 単体では P4 例外リスト外）。tx 次元外の観察
- internal/repository/reservation_repository.go:318-329 — CountByTypeAndStartTime は FOR UPDATE 無しの素 COUNT（findings resv-slot-phantom-toctou に包含。advisory lock 導入で同時に解消されるため個別対応不要）