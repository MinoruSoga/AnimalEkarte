# BE-refactor.md — バックエンド リファクタリング計画 (2026-07-08)

> **本ドキュメントは ClaudeCode エージェントが読んで直接着手するための実行計画である。**
> 前提: backend は 2026-07-02 完遂の D1-D13/R1-R3 計画で一度系統的にリファクタ済み・複数回の監査で well-maintained と判定済みのコードベースである。今回はそれ以降に実測ベースで発見された**実在する**負債のみを対象とする（推測・investigative overreach は排除済み）。

## 監査の方法と信頼性

- 対象: `backend/` 配下全体（`internal/handler` 256・`internal/service` 195・`internal/repository` 104 の非テストファイル、`cmd/` 7ツール、`worker/`、`docs/`、`migrations/`、周辺設定）。テスト552本、カバレッジ baseline 89.9%。
- 手法: 15次元（Handler/Service/Repository の P1-P18 準拠、構造的重複、大型ファイル分割、サイレント障害、tx境界/TOCTOU、クリニック間データ分離、パフォーマンス、テスト負債、cmd/周辺ツール、API契約整合、model/DDL整合、Go設計イディオム、周辺パッケージ）で並列監査し、全 79 件の所見それぞれを**敵対的検証**（対象コードを現HEADで再読し、証拠の逐語一致・既に修正済みでないか・意図的設計でないか・提案が挙動を変えないかを個別に再判定。重要度 P1 は独立した2レンズで検証）にかけた。
- 結果: 全79件が「現存する実在の負債」と確認された（REFUTED・ALREADY_FIXED は0件）。うち **46件は挙動保存（behavior-preserving）** でこのドキュメントの本編対象、**18件は修正すると外部から観測可能な挙動が変わる**ため Appendix A に分離した（このドキュメントは**コード変更を伴わない計画書**であり、Appendix A は着手前に PO/責任者判断を要する別トラックとして記録するに留める）。

## このドキュメントの使い方（ClaudeCode 向け）

1. **本編（1章〜14章 = G1〜G14）は全て挙動保存**。既存テストを事前に緑確認 → リファクタ → 同じテストが緑のままであることを確認する、という順序を厳守すること。各項目に付けた「検証コマンド」はスコープ限定（`docker compose exec backend go test ./internal/xxx/ -run ... -count=1` 等）であり、フルスイート（`go test ./...`）は自動実行禁止（`.claude/CLAUDE.md` の Auto-Execution Prohibited Commands を参照）。
2. 1項目 = 1コミットを基本とする。`dependencies` フィールドに明記された前後関係がある場合はその順序を守ること。
3. 各グループ内の番号順は実装難易度・依存関係を考慮した推奨順であり、絶対的な実行順ではない。ただし G6 の tx-mechanism-consolidation は Appendix A の tx-*-nonparticipation 3件の対応後に一括置換するのが安全（コメントに明記）。
4. **Appendix A（挙動変更トラック）は本計画の実行対象ではない**。着手する場合は各項目を個別の Issue/PR として起票し、PO 判断・STGデータ監査・段階的ロールアウトを経ること（過去の R1-3 break_hours 教訓を踏襲）。特に `sanitize-multipart-binary-corruption` と `lstep-nilcipher-stale-di` は P1（バイナリ破壊・復号不能な資格情報の本番使用）であり、リファクタ待ちにせず速やかな別トラックでの検証を推奨する。
5. Appendix B は「既に把握済みで今回あえて対象外とした」項目の現状確認結果、Appendix C/D は監査で確認できた健全領域・閾値未満の観察であり、いずれも今回のアクション対象ではない（再監査時の重複作業防止のために記録する）。

## サマリー

| 区分 | 件数 | 内訳 |
|---|---|---|
| 本編（挙動保存・G1〜G14） | 46件 | P1: 1 / P2: 16 / P3: 29 |
| Appendix A（挙動変更・別トラック） | 18件 | P1: 5 / P2: 9 / P3: 4 |
| Appendix B（既知残存・状態確認のみ） | 2件 | preload lint 3マスタ未登録 / (pwreset は Appendix A へ統合) |
| クリーン確認領域 (Appendix C) | 14次元 137項目 |
| 閾値未満の観察 (Appendix D) | 79項目 |

---

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


## G3. 構造的重複の解消 (DRY)

### G3-1. lstep タグ同期 19 ファイルで「対象解決プロローグ」と「タグ付与/解除ブロック」がコピペ横展開 — **移行完了(意図的例外あり)**

- **ID**: `dup-lstep-tag-apply`
- **重要度**: P2 / **工数目安**: L / **挙動変更**: なし（挙動保存）
- **ステータス**: ✅ **CLOSED** — Phase 1–7 + H1a–H1d 完遂。`resolveSyncTarget`/`applyTagState`(lstep_tag_sync_api.go) への統一移行が pet/visit/care/vaccine/food/checkup/prevention 全系統で完了(2026-07-09 Closure 監査で再検証済み)。
- **依存関係**: なし(dup-lstep-batch-allclinics=G3-2 と独立、G3-2 は未着手のまま)

**完了フェーズ表**

| フェーズ | 対象系統 | 代表コミット |
|---|---|---|
| Phase 1 | food (lstep_tag_sync_pet_*.go 系ヘルパ新設含む) | `d41a6fa3` |
| Phase 2 | vaccine | `096c60b8` |
| Phase 3–5a-5e | pet 系(basic/species/senior/animal classification 等) | `a35aa166`(push未) 他 |
| Phase 6 | visit 系(dormant/next/completion/cpm) | (会話サマリー記載、個別コミット未追跡) |
| Phase 7 | care 系(chronic/resync/checkup) | (会話サマリー記載、個別コミット未追跡) |
| H1a | health: food | `623e8664` |
| H1b | health: vaccine deadline(+ test fixture) | `5bca008c` |
| H1c | health: checkup | `acfcd28a` |
| H1d | health: prevention(filaria/flea-tick) | `f1aab64e` |

**意図的例外テーブル(移行不可・変更不要)**

| ファイル | 契約 | 理由 |
|---|---|---|
| `lstep_tag_sync_pet_exclusion.go` | standalone `shouldSkipSync` + inline `client.AddTag/RemoveTag`(配信注意タグ `exclTagDeliveryCaution`) | `checkOptOut` を呼ぶと LstepOptOut 飼い主を誤 skip するため `resolveSyncTarget` 不可。配信注意タグは `notifyAPIFailure` を意図的に省略するため inline 維持(Phase 5d 契約) |
| `lstep_tag_sync_visit_ltv.go` | standalone `shouldSkipSync` + per-owner ループ | clinic batch 型で `resolveSyncTarget` の owner 単位パターンが不適用(Phase 6e 契約) |
| `lstep_health_tag_sync_batch.go` | standalone `shouldSkipSync`(clinic 単位) | health-prevention batch エントリの clinic 単位ゲートであり owner 単位 `resolveSyncTarget` 対象外 |

**2026-07-09 Closure 監査 — 機械的再検証結果**

- `resolveSyncTargetOwner(` 呼び出し: `lstep_tag_sync_api.go` の定義(46行)+ `resolveSyncTarget` 内呼び出し(71行)の 2 箇所のみ。他ファイル 0 件 → PASS
- `resolveSyncTarget(` 使用: pet/visit/care/vaccine/health(food/vaccine/checkup/prevention) 全 22 サイト + api.go 定義 1 箇所、全て意図した owner-level sync メソッド → PASS
- standalone `shouldSkipSync` prod 残存: 上記 3 例外ファイル + `lstep_tag_sync_api.go`(resolveSyncTarget 内部呼び出し・コメント)+ `lstep_tag_sync_service.go`(関数定義)のみ。例外以外の未移行 0 件 → PASS
- inline `client.AddTag`/`RemoveTag`: `lstep_tag_sync_pet_exclusion.go`(配信注意タグ)+ `lstep_tag_sync_api.go`(共有ヘルパ本体)のみ。他 `lstep_tag_sync_*.go` に直書き 0 件 → PASS
- `TestLstep` 回帰: `ok  	github.com/animal-ekarte/backend/internal/service	0.100s` → PASS

**既知の P2 フォローアップ(別チケット・本 Closure では未着手)**

- `resolveSyncTarget` 内 `shouldSkipSync` エラーの `apperrors.Wrap`/slog ラップ統一
- `applyTagState` 失敗ログへの具体的タグ名の追加(現状ログに tag 名が欠落しているケースあり)

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestLstep' -count=1
```

### G3-2. Run*AllClinics 5 関数が「全クリニック走査+IsSyncEnabled ゲート+audit メタデータ記録」の逐語コピー — **CLOSED（2026-07-09）**

- **ID**: `dup-lstep-batch-allclinics`
- **重要度**: P2 / **工数目安**: S / **挙動変更**: なし（挙動保存・実証済み）
- **ステータス**: ✅ **CLOSED** — `runBatchAllClinics` ヘルパ(lstep_batch_service.go)に5関数(dormant/no-show/ltv-top-percent/visit-dormant/health-prevention)を統合。コミット `b89520c3`
- **依存関係**: なし(dup-lstep-tag-apply=G3-1 と独立に実施済み。G3-1 は既に CLOSED)

**実装内容**

`func (s *lstepBatchService) runBatchAllClinics(ctx context.Context, label, auditWarnLabel, syncedSuffix, operation string, extraMeta map[string]any, perClinic func(ctx context.Context, clinicID uint64) (int, []error)) error` を新設: clinicRepo.FindAll→IsSyncEnabled ゲート(err/disabled は continue)→perClinic→partial-errors ログ→count>0 時 InfoContext+LogLstepOperationWithMetadata(`maps.Copy` で extraMeta マージ)。`auditWarnLabel` は「audit log failed for ...」文言のハイフン有無が `label` と異なる(例: `ltv-top-percent batch` vs `ltv top percent batch`)ため独立引数として保持し、バイト級同一を実現。`RunVisitDormantSyncAllClinics` は `syncVisitDormantForClinic(ctx, clinicID) (int, []error)` を先に切り出し、entries 取得失敗時は `(0, nil)` を返すことで元の「当該クリニックを静かに continue（audit記録なし）」を再現。

**検証結果**

- スコープ限定テスト26ケース全PASS: `docker compose exec backend go test ./internal/service/ -run 'TestRun.*AllClinics|TestLstepBatch|TestDetect' -count=1` → `ok`
- ログ文言・audit operation文字列(`batch_dormant_detect`/`batch_no_show_detect`/`batch_ltv_top_percent`/`batch_visit_dormant`/`batch_health_prevention`)・dormant の `extraMeta={"min_days_since":180}` はすべて移行前と一致(実行ログで実証)
- go-reviewer: **Approve**(CRITICAL/HIGH 0件。MEDIUM: ラベル引数のstruct化提案は見送り・既存の孤立コメント(dormant.go:46)は本PRスコープ外として残置)
- silent-failure-hunter: CRITICAL/HIGH 0件。MEDIUM1件(entries取得失敗時にaudit/エラーカウントに何も残らない設計は既存負債であり本リファクタでの新規悪化なし)
- `TestLstep` 全体回帰PASS

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestRun.*AllClinics|TestLstepBatch|TestDetect' -count=1
```

### G3-3. repository 層で clinic スコープ付き Update/Delete の「Updates→RowsAffected==0→NotFound」ブロックが約 30 リポジトリに逐語コピー(Reorder は既にヘルパ化済みの片割れ) — **CLOSED（2026-07-09）**

- **ID**: `dup-repo-scoped-update-delete`
- **重要度**: P3 / **工数目安**: M / **挙動変更**: なし（挙動保存・実証済み）
- **ステータス**: ✅ **CLOSED** — `updateScopedByID`/`deleteScopedByID`(helpers.go)を新設し、対象20リポジトリのUpdate/Deleteを委譲に置換。コミット `061b79b6`(helper新設)→`212aafbd`(batch1)→`56ba497d`(batch2)→`1bad754c`(batch3)→`ceab1e31`(batch4)
- **依存関係**: なし。dbortx_inventory_lint_test.go の allowlist は対象外 repo に触れないため不変（実測: diffに `dbOrTx` 文字列の増減なし）

**証拠(現HEAD検証済み)**

cage_repository.go:58-70 `func (r *cageRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Cage, error) {\n\tresult := r.db.WithContext(ctx).\n\t\tModel(&model.Cage{}).\n\t\tScopes(clinicScope(clinicID)).Where("id = ?", id).\n\t\tUpdates(fields)\n\tif result.Error != nil {\n\t\treturn nil, apperrors.FromGORM(result.Error, "cage", fmt.Sprintf("%d", id))\n\t}\n\tif result.RowsAffected == 0 {\n\t\treturn nil, apperrors.WrapNotFound("cage", fmt.Sprintf("%d", id))\n\t}\n\treturn r.FindByID(ctx, clinicID, id)\n}` ≡ exam_type_repository.go:59-71(model/リソース名のみ差) ≡ chief_complaint_repository.go:61-73 ≡ insurance/occupation/inquiry_template/trimming_option/merchandise_item/vaccination(diff 検証済み・逐語一致)。Delete も同型: exam_type_repository.go:73-82 `result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.ExaminationType{})\n\tif result.Error != nil { return apperrors.FromGORM(...) }\n\tif result.RowsAffected == 0 { return apperrors.WrapNotFound(...) }` — `Where("id = ?", id).Delete(&model.` パターンは 26 ファイル、`Updates(fields)` は 58 ファイルに存在(うち dbOrTx=0 の定型 master 系が大多数)。同レイヤの Reorder は helpers.go:14 `func reorderByClinicID(ctx context.Context, db *gorm.DB, model any, resource string, clinicID uint64, ids []uint64) error` として既に折り畳み済み(24 repo が使用)で、Update/Delete だけ未抽出。

**問題**

P4(clinicScope 必須)+P9(FromGORM)+「RowsAffected==0→NotFound」のテナント隔離・エラー写像契約が約 30 repo に逐語コピーされている。この契約はスコープ述語や NotFound 写像を変える際に全箇所同期必須(過去のクロステナント監査 72e8887c/b3638d5e はまさにこのクラスの横断修正)。Reorder のみヘルパ化済みで対称性が欠けている。

**実装内容**

`helpers.go` に `reorderByClinicID` と同型の2ヘルパを追加:
`func updateScopedByID(ctx context.Context, db *gorm.DB, m any, resource string, clinicID, id uint64, fields map[string]any) error`(`Model(m).Scopes(clinicScope(clinicID)).Where("id = ?", id).Updates(fields)` → `FromGORM` → `RowsAffected==0` → `WrapNotFound`)と `func deleteScopedByID(ctx context.Context, db *gorm.DB, m any, resource string, clinicID, id uint64) error`(同型のDelete版)。呼び出し側は `if err := updateScopedByID(...); err != nil { return nil, err }\nreturn r.FindByID(ctx, clinicID, id)` の2行(Preload付きrefetchは呼び出し側の責務として保存)、または `owner` のような error-only変種は `return updateScopedByID(...)` の1行に置換。

対象20リポジトリを4バッチに分けて実施:
- batch1(`212aafbd`): cage / exam_type / chief_complaint / checkup_type / consultation
- batch2(`56ba497d`): insurance / occupation / inquiry_template / trimming_option / trimming_course_type / trimming_course
- batch3(`1bad754c`): merchandise_item / payment_method_master / procedure / medicine / vaccine
- batch4(`ceab1e31`): hospitalization_plan / reservation_type_group / vaccination / owner(error-only変種、FindByID refetchなし)

**意図的除外テーブル(移行不可・変更不要)**

| ファイル | 理由 |
|---|---|
| `vital_repository.go` | Update Where に `deleted_at IS NULL` 述語を含む変種 |
| `pet_repository.go` | Update RowsAffected==0 後に Count 分岐する変種 |
| `medical_record_repository.go` | RowsAffected==0→`WrapConflict` 変種(NotFound でない) |
| `treatment_repository.go` | subquery 隔離のため定型と不一致 |
| dbOrTx 参加 repo(accounting/billing_item/staff/reservation 等 16 ファイル) | tx 意味論があり `r.db` 直渡し不可 |
| BE リスト外の canonical 一致 repo(hospitalization/estimate/inventory/diagnosis 等) | スコープ最小化(本 Closure では対象外) |

**検証結果**

- helper契約テスト: `TestUpdateScopedByID`/`TestDeleteScopedByID`(成功・NotFound・別クリニック隔離の3パターン)全PASS
- 20リポジトリのスコープ限定回帰 + `TestPreloadClinicScope`/`TestDbOrTx` 全PASS(下記コマンド)
- `internal/repository` パッケージ全体(全リポジトリ)の回帰 108.688s 全PASS(既存機能に影響なし)
- 除外ファイル(vital/pet/medical_record/treatment)への変更ゼロを `git diff --stat` で確認、旧インラインパターン(`Scopes(clinicScope(clinicID)).Where("id = ?", id).\n\tUpdates`)は対象20ファイルから `rg` で0件確認(完全消失)
- go-reviewer: **Approve**(CRITICAL/HIGH/MEDIUM 0件。20リポジトリ全件で resource文字列・エラー写像・FindByID refetch契約の完全一致を確認)
- clinic-isolation-auditor: **Approve**(CRITICAL/HIGH/MEDIUM 0件。`clinicScope` バイパス経路なし、Preload契約は本diffで無変更、owner/vaccinationの変種も隔離契約は他repoと同一と確認)
- 両レビューで検出された「複数テストをまとめて実行すると非決定的にFAILする」事象は、`helpers.go`変更前後で同一再現する既存のテスト分離問題(共有DBプール上のclinicID衝突)であり、本G3-3の欠陥ではない(別チケット追跡を推奨・下記Harness Improvement Feedback参照)

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestCage|TestExamType|TestChiefComplaint|TestInsurance|TestVaccination|TestPreloadClinicScope|TestDbOrTx' -count=1
```

### G3-4. medical_records 経由テナント隔離 JOIN 断片が 7 リポジトリに逐語散在 — **CLOSED（2026-07-09）**

- **ID**: `dup-medrecord-tenant-join`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存・実証済み）
- **ステータス**: ✅ **CLOSED** — `medicalRecordTenantScope(childTable, clinicID)`(base.go)を新設し、対象7リポジトリの手書き JOIN を `Scopes(...)` 委譲に置換。コミット `edc315e7`(G3-4a: helper新設+契約テスト)→`4c9e686a`(G3-4b: 7箇所置換)→`0f76b71e`(go-reviewer MEDIUM指摘によるテストメッセージ修正)
- **対象ファイル**: internal/repository/chief_complaint_repository.go (85); internal/repository/consultation_repository.go (108); internal/repository/diagnosis_repository.go (242); internal/repository/inventory_repository.go (152); internal/repository/medicine_repository.go (67); internal/repository/medical_record_image_repository.go (67); internal/repository/procedure_repository.go (86)
- **依存関係**: dup-repo-scoped-update-delete(G3-3)と同一ファイル群に触れるため同一トラックで実施(G3-3 CLOSED後に着手・技術依存はなし)

**証拠(現HEAD検証済み)**

7 箇所が子テーブル名以外逐語一致: chief_complaint_repository.go:85 `Joins("JOIN medical_records ON medical_records.id = inquiries.medical_record_id AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL", clinicID).` / consultation_repository.go:108・inventory_repository.go:152・medicine_repository.go:67・procedure_repository.go:86 は `= treatments.medical_record_id` で同文 / diagnosis_repository.go:242 は `= clinical_plans.medical_record_id` / medical_record_image_repository.go:67 は `= medical_record_images.medical_record_id`。いずれも CountUsage 系のテナント隔離述語(clinic_id + deleted_at)であり、直上コメント(例: chief_complaint_repository.go:79-80 「inquiries テーブルは clinic_id を持たないため medical_records を JOIN してテナント分離する」)も同旨の複製。

**問題**

clinic_id を持たない子テーブル(inquiries/treatments/clinical_plans/medical_record_images)のテナント隔離を担う SQL 断片が 7 箇所に手書き複製されている。隔離述語(clinic_id / deleted_at)の形を変える場合は全箇所同期必須で、1 箇所の書き漏れがクロステナント read 漏洩クラス(#124/#125 と同系)に直結する。Preload 側は P3.1 lint で機械強制済みだが、手書き Joins はどのゲートも見ていない。

**実装手順**

base.go に GORM スコープを追加: `// medicalRecordTenantScope は clinic_id を持たない medical_record 子テーブルのテナント隔離 JOIN。childTable は呼び出し側リテラルのみ許可。\nfunc medicalRecordTenantScope(childTable string, clinicID uint64) func(*gorm.DB) *gorm.DB { return func(db *gorm.DB) *gorm.DB { return db.Joins("JOIN medical_records ON medical_records.id = "+childTable+".medical_record_id AND medical_records.clinic_id = ? AND medical_records.deleted_at IS NULL", clinicID) } }`。7 箇所を `Scopes(medicalRecordTenantScope("inquiries", clinicID))` 等に置換(childTable はコンパイル時リテラルのみ・doc コメントで SQL 組み立てへの変数流入禁止を明記)。生成 SQL は文字列連結結果が現行と同一のため挙動保存。既存の各 *_test.go / clinic_isolation テストが回帰網。clinic_id 述語なし変種(billing_confirmation_repository.go:32 等、別述語で隔離済みの箇所)は意図が異なるため対象外。

**実装内容**

`base.go` に `medicalRecordTenantScope(childTable string, clinicID uint64) func(*gorm.DB) *gorm.DB` を追加(既存 `clinicScope`/`clinicScopeIn` と同じ「`func(clinicID uint64) func(*gorm.DB) *gorm.DB`」シグネチャ・命名規則に準拠)。`base_test.go` に `TestMedicalRecordTenantScope` を新設し、正 clinic での JOIN 適用・別 clinic 除外・該当 clinic 無しの3契約を検証。

**7 箇所対応表**

| ファイル | メソッド | childTable |
|---|---|---|
| chief_complaint_repository.go:78 | CountUsageByChiefComplaintTypeID | `inquiries` |
| consultation_repository.go:94 | CountUsageByConsultationID | `treatments` |
| diagnosis_repository.go:242 | CountUsageByDiagnosisNameID | `clinical_plans` |
| inventory_repository.go:152 | CountUsageByInventoryID(treatment部分のみ) | `treatments` |
| medicine_repository.go:67 | CountUsageByMedicineID(treatment部分のみ) | `treatments` |
| procedure_repository.go:72 | CountUsageByProcedureID(treatment部分のみ) | `treatments` |
| medical_record_image_repository.go:67 | FindByID | `medical_record_images` |

**意図的除外(移行不可・変更不要)**

| ファイル/箇所 | 理由 |
|---|---|
| medical_record_image_repository.go: FindByMedicalRecordID(33) | clinic_id が JOIN でなく Where 句側 |
| medical_record_image_repository.go: Delete(52) | subquery で隔離、JOIN形状ではない |
| treatment_repository.go | JOIN に clinic_id 述語なし(subquery 隔離) |
| clinical_plan_repository.go(35) | clinic_id 述語なし |
| billing_confirmation_repository.go(32) | 別述語で隔離済み |
| inquiry_repository.go(95) | clinic_id 述語なし |
| checkup_repository.go(77,93) | 文字列連結・別 SQL 形状 |
| medicine/procedure の hospitalizations JOIN(care_plan_items側) | medical_records JOIN ではないため対象外 |

**検証結果**

- helper契約テスト: `TestMedicalRecordTenantScope`(正clinic・別clinic・該当なし・medical_records論理削除除外の4パターン)全PASS
- 7リポジトリのスコープ限定回帰(`TestChiefComplaint|TestConsultation|TestDiagnosis|TestInventory|TestMedicine|TestProcedure|TestMedicalRecordImage|TestMedicalRecordTenantScope`)全PASS。うち `TestChiefComplaintTypeRepository_CountUsageByChiefComplaintTypeID_KnownInquiriesColumnBug`(inquiries.deleted_at列欠落)と `TestProcedureRepository_CountUsageByProcedureID_KnownCarePlanItemsColumnBug`(care_plan_items.deleted_at列欠落)は本リファクタ前から存在する既知のスキーマ不整合で、置換後も同一エラーを再現(=挙動保存)することを確認。helper非起因の既知負債であり別チケット追跡対象
- 隔離回帰: `TestMedicalRecordImageRepository_FindByID_ClinicIsolation`/`Delete_ClinicIsolation`/`TestPetRepository_Update_ClinicIsolation`(同一ファイル)全PASS
- 旧インラインパターン(`JOIN medical_records ON medical_records\.id = .* AND medical_records\.clinic_id = \?`)は対象7ファイルから `rg` で0件確認(完全消失)。意図的除外7箇所は `git diff --stat` で無変更を確認
- go-reviewer: **Approve**(CRITICAL/HIGH 0件。MEDIUM 2件: ①テストのアサーションメッセージが「treatments.deleted_at」を指すかのように誤読を招く表現だった点→`0f76b71e`で修正済み、②`childTable`のコンパイル時リテラル制約を担保する `go/ast` lint が無い点→将来の追加箇所向けフォローアップとして下記に記録・本Closureではブロッキングとせず)
- clinic-isolation-auditor: **Approve**(CRITICAL/HIGH/MEDIUM 0件。`childTable` は全呼び出しサイトでコンパイル時リテラルのみ、生成SQLは置換前とバイト同一の隔離述語を保持、意図的除外7箇所は無変更、既存の全隔離テストPASSを確認)

**フォローアップ候補(本Closureではスコープ外)**

- go-reviewer提案: `medicalRecordTenantScope` の `childTable` 引数がリテラルのみであることを `preload_clinic_scope_lint_test.go` 同様の `go/ast` ベース lint で機械強制する余地あり(現状はコードレビュー運用のみ)
- `inquiry_repository.go:95` / `treatment_repository.go` / `clinical_plan_repository.go` / `billing_confirmation_repository.go` に残る「JOIN+WHERE分離型」の別形状隔離パターンも将来的に統合候補だが、SQL文字列のbyte-identical要件を満たす形での統合は別途設計が必要なため本G3-4のスコープ外

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestChiefComplaint|TestConsultation|TestDiagnosis|TestInventory|TestMedicine|TestProcedure|TestMedicalRecordImage|TestMedicalRecordTenantScope' -count=1
```

### G3-5. handler 層に「YYYY-MM-DD 優先→RFC3339 フォールバック」日付パースが二重実装(エラー衛生ドリフトあり) — **CLOSED（第一段+第二段 完了・2026-07-09）**

- **ID**: `dup-handler-date-parse`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: 第一段=なし（挙動保存） / 第二段=あり（behaviorChange・意図的セキュリティ改善）
- **対象ファイル**: internal/handler/date.go; internal/handler/date_test.go; internal/handler/vaccination_request.go; internal/handler/vaccination_request_test.go; internal/handler/inventory_request.go; internal/handler/inventory_request_test.go; internal/handler/reservation_type_request.go(無変更確認のみ)
- **依存関係**: なし
- **ステータス**: ✅ **CLOSED（全段完了）**
  - **第一段**（挙動保存）: `parseFlexibleDate(s string) (time.Time, error)`(date.go)を新設し、`jsonDate.UnmarshalJSON`/`parseDate` を委譲化。`handler_date_helpers.go` を削除し date.go へ統合。両者のエラーメッセージ(jsonDate=汎用文言 / parseDate=`invalid date format: %s` の入力値エコー)は完全維持。契約テスト `date_test.go` 新設(TestJsonDate/TestParseDate)。go-reviewer / security-reviewer ともに Approve(CRITICAL/HIGH 0)。コミット `e3ccc7e3`
  - **第二段**（behaviorChange・#97 類例のクローズ）: date.go に共通定数 `flexibleDateInvalidInputMsg` を新設し、`jsonDate.UnmarshalJSON`/`parseDate` の両方が共有。parseDate 失敗時の `fmt.Errorf("invalid date format: %s", *dateStr)`（入力値エコー）を廃止し `apperrors.WrapInvalidInput(flexibleDateInvalidInputMsg)` に統一。呼び出し側 vaccination_request.go(4箇所)/inventory_request.go(4箇所)の `fmt.Sprintf("invalid date: %v", err)` / `fmt.Errorf("invalid expiry_date: %w", err)` 等のラッピングを廃止し `return nil, err` の bare 伝播に統一(parseDate が既に正しく型付けされた InvalidInput エラーを返すため)。reservation_type_request.go は元々 bare 伝播済みのため無変更(diff なしを確認)。テストは `date_test.go`(TestParseDate 不正系で `apperrors.IsInvalidInput` + `NotContains(input)` 検証、jsonDate と同一メッセージを定数参照で検証)、`vaccination_request_test.go`/`inventory_request_test.go`(テーブル駆動で create/update × 各日付フィールド全組み合わせの非露出を検証)を更新。go-reviewer(Warning・CRITICAL/HIGH 0、MEDIUM=テストカバレッジ拡充を同コミットで解消済み) / security-reviewer(Approve・CRITICAL/HIGH 0)。コミット `42aa6b4f`

**証拠(現HEAD検証済み)**

date.go:21-32 (jsonDate.UnmarshalJSON) `// YYYY-MM-DD を優先\n\tif t, err := time.ParseInLocation("2006-01-02", s, time.Local); err == nil {\n\t\td.Time = t\n\t\treturn nil\n\t}\n\t// フォールバック: RFC3339\n\tif t, err := time.Parse(time.RFC3339, s); err == nil {\n\t\td.Time = t\n\t\treturn nil\n\t}\n\t// Go内部のエラー文字列を漏洩させないため、汎用メッセージを返す\n\treturn apperrors.WrapInvalidInput("日付の形式が正しくありません...")` と handler_date_helpers.go:16-28 (parseDate) `// フォーマット1: YYYY-MM-DD\n\tt, err := time.ParseInLocation("2006-01-02", *dateStr, time.Local)\n\tif err == nil { return &t, nil }\n\t// フォーマット2: RFC3339\n\tt, err = time.Parse(time.RFC3339, *dateStr)\n\tif err == nil { return &t, nil }\n\treturn nil, fmt.Errorf("invalid date format: %s", *dateStr)` が同一パースチェーンの二重実装。ドリフト: parseDate は入力実値をエラーにエコーし、vaccination_request.go:78-81 `date, err := parseDate(r.Date)\n\tif err != nil {\n\t\treturn nil, apperrors.WrapInvalidInput(fmt.Sprintf("invalid date: %v", err))\n\t}` 経由でクライアント応答へ露出する(jsonDate 側は date.go:31 のコメントどおり意図的に汎用メッセージ化済み — #97 本文実値露出と同クラスの非対称)。

**問題**

同一 wire 契約(YYYY-MM-DD または RFC3339 のリクエスト日付)の受理ロジックが 2 実装に分裂し、受理形式を変える際は両方の同期変更が必須。さらにエラー衛生方針(入力値エコーの可否)が実装間で既に乖離しており、片方だけ直す事故の温床。parseDate 利用は inventory_request.go(4)/reservation_type_request.go(2)/vaccination_request.go(4) の 10 サイト。

**実装手順**

挙動保存の第一段: date.go に共通コア `func parseFlexibleDate(s string) (time.Time, error)`(YYYY-MM-DD→RFC3339 の現行チェーン、失敗時は sentinel error)を切り出し、jsonDate.UnmarshalJSON と parseDate の両方をコア委譲に書き換える。各ラッパの現行エラーメッセージ(jsonDate=汎用文言 / parseDate=`invalid date format: %s`)は文字通り維持し応答互換を保つ。handler_date_helpers.go は date.go へ統合し削除(パッケージ内 1 ファイルに日付パースを集約)。第二段(behaviorChange=true・別トラック): parseDate のエラーから入力値エコーを除去し jsonDate と同一の汎用文言へ統一(#97 類例のクローズ)。

**検証コマンド(スコープ限定・第二段で実行・PASS)**
```
docker compose exec backend go test ./internal/handler/ -run 'TestParseDate|TestJsonDate|TestCreateInventoryRequest|TestInventoryRequest|TestCreateVaccinationRequest|TestVaccinationRequest|TestUpdateVaccinationRequest|TestUpdateInventoryRequest|TestCreateUnavailableTimeRequest' -count=1
docker compose exec backend go vet ./internal/handler/...
docker compose exec backend gofmt -l internal/handler/date.go internal/handler/date_test.go internal/handler/vaccination_request.go internal/handler/vaccination_request_test.go internal/handler/inventory_request.go internal/handler/inventory_request_test.go
```


## G4. Handler層 P1-P18 規約準拠

### G4-1. P7逸脱: lab_report_handler が service 戻り値(model 層 DTO)を toXxxResponse 変換なしで c.JSON 直返し — **CLOSED（2026-07-09）**

- **ID**: `p7-lab-report-model-dto-passthrough`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/handler/lab_report_handler.go (28-33,53-58); internal/handler/lab_report_response.go(新規); internal/model/lab_report.go (20-65)
- **依存関係**: なし
- **ステータス**: ✅ **CLOSED** — internal/handler/lab_report_response.go を新規作成し、labExamReportSummaryResponse/labExamReportDetailResponse/labExamResultItemResponse を model 側の旧 json タグと完全同一タグで定義。toLabExamReportSummaryResponse/toLabExamReportDetailResponse/toLabExamResultItemResponse(P18命名)は 1:1 フィールドコピーのみ(変換ロジック追加なし)。lab_report_handler.go:33 は `mapSlice(summaries, toLabExamReportSummaryResponse)`、:58 は `toLabExamReportDetailResponse(detail)` 経由に変更。model/lab_report.go の LabExamReportSummary/LabExamReportDetail/LabExamResultItem から json タグを削除(LabReportFilter は diff なし・無変更を確認)。lab_report_handler_test.go の ExactFieldAllowlist/PII/Happy-path/omitempty テスト17本全て PASS(wire 形状不変を実証)、service 層 TestLabReportQuery 系14本 PASS。go vet クリーン、gofmt 差分なし。go-reviewer(Approve・CRITICAL/HIGH/MEDIUM 0)/ security-reviewer(CRITICAL/HIGH/MEDIUM 0、PII allowlist フィールド完全一致を逐語確認)。コミット `689e9506`

**証拠(現HEAD検証済み)**

internal/handler/lab_report_handler.go:28-33: `summaries, err := h.svc.LabReportQuery.ListJobReportSummaries(c.Request.Context(), clinicID, jobID)` → `c.JSON(http.StatusOK, summaries)`。同53-58: `detail, err := h.svc.LabReportQuery.GetExamReport(...)` → `c.JSON(http.StatusOK, detail)`。戻り値型は internal/service/lab_report_query_service.go:30 `ListJobReportSummaries(...) ([]model.LabExamReportSummary, error)` / 同33 `GetExamReport(...) (*model.LabExamReportDetail, error)`。ワイヤ形式は internal/model/lab_report.go:20-31 `type LabExamReportSummary struct { ExamID uint64 `json:"exam_id"` ... }` として model 層の json タグで定義。handler 直上コメント(lab_report_handler.go の PII-safe 設計注記と model/lab_report.go:17-35 の DTO doc)は PII 除外設計を記すのみで、handler 変換省略の根拠記載はない。

**問題**

P7「c.JSON にモデルを直接渡さない。必ず変換関数を経由する」の残存逸脱。レスポンス契約(json タグ)が handler/*_response.go ではなく model 層に漏れており、accounting 系が repository 集計 DTO ですら toDailySummaryResponse 等で handler 変換している(accounting_response.go:30-164)のと非対称。model DTO のフィールド変更が意図せず API レスポンス形状を変える経路が開いている。PII-safe 設計自体は文書化済みで正当。

**実装手順**

挙動保存。手順: (1) internal/handler/lab_report_response.go を新規作成し、labExamReportSummaryResponse / labExamReportDetailResponse / labExamResultItemResponse を model/lab_report.go:20-65 の json タグと完全同一タグで定義、toLabExamReportSummaryResponse / toLabExamReportDetailResponse を実装(P18命名)。(2) lab_report_handler.go:33 を変換経由(summaries は値スライスのためループまたは mapSlice 適合形)に、:58 を `c.JSON(http.StatusOK, toLabExamReportDetailResponse(detail))` に変更。(3) internal/model/lab_report.go の LabExamReportSummary/LabExamReportDetail/LabExamResultItem から json タグを削除しワイヤ定義を handler 側へ一本化(LabReportFilter はタグなしのため対象外)。(4) 既存の lab_report_handler_test.go の 200 系テストが json キーを検証していることを確認し wire 同一を担保。影響範囲: handler/model と両者のテストのみ、wire 不変。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/handler/ -run 'TestGetLabJobReportSummaries|TestGetLabExamReport' -count=1 && docker compose exec backend go test ./internal/service/ -run TestLabReportQuery -count=1
```

### G4-2. P7逸脱: CreateLiffReservation 成功レスポンスが gin.H でモデルフィールドを直返し — **CLOSED（2026-07-09）**

- **ID**: `p7-liff-create-reservation-ginh`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/handler/liff_handler.go (197-198)
- **依存関係**: なし
- **ステータス**: ✅ **CLOSED** — internal/handler/liff_response.go に `liffReservationCreatedResponse{ID uint64 \`json:"id"\`; Notes string \`json:"notes"\`}` と `toLiffReservationCreatedResponse(r *model.Reservation)`(P18命名、1:1フィールドコピーのみ)を新設。既存の一覧用 `liffReservationResponse` とは別型として分離(命名衝突なし)。liff_handler.go:198 を `c.JSON(http.StatusCreated, toLiffReservationCreatedResponse(appt))` に置換、L197 の Location ヘッダは無変更。TestCreateLiffReservation の正常系アサートを `assert.Contains(..., "42")`(部分一致)から `json.Unmarshal` 後の `assert.Equal(map[string]any{"id": float64(42), "notes": ""}, body)`(完全一致)+ Location ヘッダ個別 assert に強化、全9サブテスト PASS。go vet / gofmt クリーン。go-reviewer(Approve・CRITICAL/HIGH/MEDIUM 0、P7/P18準拠確認)/ security-reviewer(CRITICAL/HIGH 0、gin.H 時代と同一の2フィールドのみでモデル全体露出なしを確認)。コミット `a00bf673`

**証拠(現HEAD検証済み)**

internal/handler/liff_handler.go:197-198:
```go
c.Header("Location", fmt.Sprintf("/api/v1/reservations/%d", appt.ID))
c.JSON(http.StatusCreated, gin.H{"id": appt.ID, "notes": appt.Notes})
```
appt は internal/service/liff_service.go:21 `CreateReservation(...) (*model.Reservation, error)` の戻り値でモデル(Notes は internal/model/reservation.go:50 で string 型)。liff_handler 内の他エンドポイントは全て toLiffXxxResponse / typed struct 経由(liff_handler.go:62,83,116,141,214,230,252 で確認)で、成功系 gin.H はここのみ。前後30行に免除コメントなし。リポジトリ内先例: internal/handler/staff_response.go:96 「旧 gin.H{"...": ids} を P7 準拠の typed struct に置き換える（JSON 表現は同一）」。

**問題**

P7 は gin.H 直返しを明示的に違反例としている。LIFF 公開 API の免除は P5(認証)の免除であり P7 の免除ではない。型なし map のためフィールド追加・改名時のコンパイル時検査が効かず、レスポンス契約がコード上の型として存在しない。

**実装手順**

挙動保存。手順: (1) liff 系 response 定義ファイルに `type liffReservationCreatedResponse struct { ID uint64 `json:"id"`; Notes string `json:"notes"` }` と `func toLiffReservationCreatedResponse(appt *model.Reservation) liffReservationCreatedResponse` を追加。(2) liff_handler.go:198 を `c.JSON(http.StatusCreated, toLiffReservationCreatedResponse(appt))` に置換。json キーは "id"/"notes" のまま wire 不変。(3) TestCreateLiffReservation でレスポンス形状不変を確認。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/handler/ -run TestCreateLiffReservation -count=1
```

### G4-3. P14周辺負債: Handler ルートコンテナが全 repository 集約(*repository.Repositories)を保持(実使用は LiffAuth 用2リポジトリのみ) — **CLOSED（2026-07-09）**

- **ID**: `p14-handler-container-repos-field`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/handler/handler.go (18-28); internal/handler/reservation_line_routes.go (64-65); internal/handler/handler_routes_test.go (26-32); cmd/api/main.go (191)
- **依存関係**: なし
- **ステータス**: ✅ **CLOSED** — Handler struct から `repos *repository.Repositories` を削除し、`liffCustomerLookup middleware.LineCustomerLookup` / `liffSettingLookup middleware.LineReservationSettingLookup` の2 narrow field に置換。New は repos != nil ガード付きで2値を抽出代入(シグネチャ不変)。reservation_line_routes.go:65 を `middleware.LiffAuth(h.liffCustomerLookup, h.liffSettingLookup)` に変更。handler_routes_test.go の `repos: &repository.Repositories{}` を削除(未使用 import も除去)。handler_test.go:30 の `assert.Same(t, repos, h.repos)` は narrow field 単位の assert.Same 2本に更新 — ゼロ値 struct だと nil interface 比較で testify が失敗するため、repos.LineCustomerMgr/LineReservationSetting に実コンストラクタ(`repository.NewLineCustomerRepository(nil)` 等、DB 未使用のため nil で安全)を渡し pointer identity を可視化。`docker compose exec backend go test ./internal/handler/ -run '^(TestRegisterRoutes_NoPanic|TestGetLiffProfile|TestGetLiffSettings|TestCreateLiffReservation|TestNew)$' -count=1` 全 PASS。grep `h\.repos` handler パッケージ 0件。go vet / gofmt 差分なし。コミット `5e5754db`

**証拠(現HEAD検証済み)**

internal/handler/handler.go:18-23:
```go
type Handler struct {
	cfg      *config.Config
	svc      *service.Services
	repos    *repository.Repositories
	uploader infra.FileUploader
}
```
h.repos の全参照は internal/handler/reservation_line_routes.go:65 の1箇所のみ: `liffAuth := middleware.LiffAuth(h.repos.LineCustomerMgr, h.repos.LineReservationSetting)`(handler 非テスト256ファイル grep 全走査で確認)。middleware.LiffAuth は既に narrow interface を受ける設計: internal/middleware/liff_auth.go:49 `func LiffAuth(lookup LineCustomerLookup, settingLookup LineReservationSettingLookup) gin.HandlerFunc`。リクエスト処理 handler メソッドからの h.repos 呼び出しは0件。

**問題**

P14「handler struct は svc XxxService のみを持つ。repo フィールドは禁止」に対し、ルートコンテナが全リポジトリへのアクセス経路を常時保持。現状の実害は0(middleware DI 配線のみ)だが、任意の handler メソッドが `h.repos.X.FindAll(...)` と service 層をバイパスできる構造的余地が残り、P14 遵守をコードレビュー頼みにしている。middleware 側が narrow interface のため、必要な2依存だけ保持すれば全 repo 集約は不要。

**実装手順**

挙動保存。手順: (1) internal/handler/handler.go の Handler struct から `repos *repository.Repositories` を削除し、`liffCustomerLookup middleware.LineCustomerLookup` と `liffSettingLookup middleware.LineReservationSettingLookup` の2フィールドを追加。(2) New のシグネチャは維持し、本体で `liffCustomerLookup: repos.LineCustomerMgr, liffSettingLookup: repos.LineReservationSetting` を代入(repos が nil の場合に備え nil ガード付き代入。非テスト呼出は cmd/api/main.go:191 の1箇所で常に非nil)。(3) reservation_line_routes.go:65 を `middleware.LiffAuth(h.liffCustomerLookup, h.liffSettingLookup)` に変更。(4) internal/handler/handler_routes_test.go:30 の `repos: &repository.Repositories{}` を削除(新フィールドは nil のままで可 — LiffAuth はクロージャ返却のみで構築時に lookup を deref しないためルート登録は panic しない)。影響範囲: handler.go / reservation_line_routes.go / handler_routes_test.go の3ファイル(+必要なら cmd/api/main.go)。他テストは &Handler{svc: ...} 直構築のため無影響。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/handler/ -run 'TestRegisterRoutes_NoPanic|TestGetLiffProfile|TestGetLiffSettings' -count=1
```


## G5. Service層 P1-P18 規約準拠

### G5-1. P11違反: 業務レコード系サービスのDelete/Create書込エラーパス11箇所でslog.ErrorContext欠落 — **CLOSED（2026-07-09）**

- **ID**: `p11-slog-gap-business-delete-paths`
- **重要度**: P2 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/service/prescription_service.go (160-162); internal/service/vital_service.go (220-222); internal/service/checkup_service.go (294-296); internal/service/vaccination_service.go (206-208); internal/service/clinical_plan_service.go (159-161); internal/service/medical_record_image_service.go (100-102); internal/service/estimate_service.go (216-224); internal/service/treatment_plan_service.go (177-179); internal/service/liff_service_reservations.go (109-111); internal/service/staff_clinic_assignment_service.go (35-37)
- **ステータス**: ✅ **CLOSED** — 10ファイル11箇所の repo.Delete/CountItemsByEstimateID/repo.Create エラーブロック先頭に、各ファイル既存の slog.ErrorContext キー体系("error", err, "clinic_id", clinicID, "<entity>_id", id)で1行追加。wrap メッセージ文言は無変更。`docker compose exec backend go test ./internal/service/ -run 'TestPrescription|TestVital|TestCheckup|TestVaccination|TestClinicalPlan|TestMedicalRecordImage|TestEstimate|TestTreatmentPlan|TestLiff|TestStaffClinicAssignment' -count=1` PASS。go vet / gofmt 差分なし。コミット `3eca7ef2`

**証拠(現HEAD検証済み)**

prescription_service.go:160-162:
```go
if err := s.repo.Delete(ctx, clinicID, prescriptionID); err != nil {
    return apperrors.Wrap(err, "failed to delete prescription")
}
```
vital_service.go:220-222:
```go
if err := s.repo.Delete(ctx, clinicID, vitalID); err != nil {
    return apperrors.Wrap(err, "failed to delete vital record")
}
```
estimate_service.go:216-218 (依存チェックのDB読取エラーも同様):
```go
count, err := s.repo.CountItemsByEstimateID(ctx, clinicID, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check estimate item dependencies")
}
```
liff_service_reservations.go:109-111:
```go
if err := s.adminRepo.CancelByID(ctx, clinicID, customerID, reservationID); err != nil {
    return apperrors.Wrap(err, "failed to cancel reservation")
}
```
staff_clinic_assignment_service.go:34-37:
```go
func (s *staffClinicAssignmentService) Create(ctx context.Context, assignment *model.StaffClinicAssignment) error {
    if err := s.repo.Create(ctx, assignment); err != nil {
        return apperrors.Wrap(err, "failed to create staff clinic assignment")
    }
```
他: checkup_service.go:295 / vaccination_service.go:207 / clinical_plan_service.go:160 / medical_record_image_service.go:101 / treatment_plan_service.go:178 / estimate_service.go:224。対照としてマスタ系は全件準拠 (例 reservation_type_service_core.go:160-163 は repo.Delete エラー前に slog.ErrorContext あり)。RespondError (handler/response.go:16) はサーバ側ログを一切出力しないため、これらのDB書込障害はログ痕跡ゼロで500になる。

**問題**

P11 (MANDATORY, .claude/refs/gin-architecture-compliance.md L229-247) は repo 起因のインフラエラーのリターン前に slog.ErrorContext を要求する。除外対象は WrapInvalidInput / NotFound存在確認 / WrapConflict のみで、FindByID通過後の repo.Delete / Create / CountUsage の失敗は除外に該当しない。マスタ系サービス約25ファイルは全件準拠している一方、カルテ配下サブレコード(処方・バイタル・健診・ワクチン接種・診療計画・画像)と見積・治療計画・LIFF予約キャンセル・スタッフ医院割当の書込パスだけが欠落しており、レイヤ内で診断可能性が非対称。ハンドラの RespondError もログを出さないため障害調査時に手掛かりが残らない。

**実装手順**

各ファイルの該当 if err ブロック先頭に、同ファイル既存の slog.InfoContext と同じキー体系 ("error", err, "clinic_id", clinicID, "<entity>_id", id) で slog.ErrorContext を1行追加する。手順: (1) prescription_service.go:160 / vital_service.go:220 / checkup_service.go:294 / vaccination_service.go:206 / clinical_plan_service.go:159 / medical_record_image_service.go:100 / treatment_plan_service.go:177 の repo.Delete エラーブロック 7箇所、(2) estimate_service.go:217(CountItemsByEstimateID) と :223(repo.Delete) の2箇所、(3) liff_service_reservations.go:109 CancelByID、(4) staff_clinic_assignment_service.go:35 repo.Create。既存メッセージ文字列は変更しない(テストがエラーメッセージをアサートしている可能性があるため wrap 文言据え置き)。ログ追加のみでAPI応答・戻り値は不変。影響範囲: internal/service の10ファイルのみ、interface変更なし。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestPrescription|TestVital|TestCheckup|TestVaccination|TestClinicalPlan|TestMedicalRecordImage|TestEstimate|TestTreatmentPlan|TestLiff|TestStaffClinicAssignment' -count=1
```

### G5-2. P13違反: buildFuncがinterface/struct/Newより後に定義されている4ファイル — **CLOSED（2026-07-09）**

- **ID**: `p13-buildfunc-order-4-files`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/service/chronic_condition_service.go (32,39,46,143); internal/service/audit_service.go (12,64,74,121); internal/service/lstep_settings_service.go (105,128,140,250); internal/service/lab_import_examination_service.go (74,97,103,234)
- **ステータス**: ✅ **CLOSED** — 純粋なコード移動のみ(シグネチャ・本文変更なし)。buildChronicConditionUpdateFields を chronic_condition_service.go の interface 直前へ、buildAuditLog を audit_service.go の interface 直前へ(`var _ AuditTxLogger = (*auditService)(nil)` は struct 直下のまま維持)、buildLstepSettingsResponse を lstep_settings_service.go の interface 直前へ、buildExamResults を lab_import_examination_service.go の interface 直前へ移動。computeExamResultStatus は examination_service.go 側の定義のため本ファイル内の後方参照ではなく移動対象外。`docker compose exec backend go test ./internal/service/ -run 'TestChronicCondition|TestAudit|TestLstepSettings|TestLabImport' -count=1` PASS。go vet / gofmt 差分なし。diff は4ファイルとも insertions==deletions(コード移動のみを裏付け)。コミット `5c7434aa`

**証拠(現HEAD検証済み)**

P13 の必須順序は const → buildFunc → interface → struct → New → methods (.claude/refs/gin-architecture-compliance.md L273-292)。chronic_condition_service.go は interface(L32)/struct(L39)/New(L46) の後の L143 に典型的 buildFunc がある:
```go
func buildChronicConditionUpdateFields(input UpdateChronicConditionInput) map[string]any {
    fields := make(map[string]any)
```
audit_service.go は interface(L12)/struct(L64) の後 L74 に buildFunc (自らドキュメントコメントで buildFunc と宣言):
```go
// buildAuditLog は AuditLogInput を model.AuditLog に変換する（LogEntry / LogEntryTx 共通の buildFunc）。
func buildAuditLog(input *AuditLogInput) *model.AuditLog {
```
lstep_settings_service.go:250 buildLstepSettingsResponse は New(L140) より後、lab_import_examination_service.go:234 buildExamResults は New(L103) より後。4ファイルとも配置を正当化するコメントは無し(buildAuditLog直上コメントは変換内容の説明のみ、buildExamResults直上コメントは信頼境界の説明のみで配置には言及しない)。

**問題**

P13 は service 層 MANDATORY パターン (backend/internal/service/CLAUDE.md L64-73) だが golangci-lint に順序チェックのlinterが無く、この4ファイルだけがドリフトしている(残り191ファイルは準拠)。レビュー時の探索コスト増と、新規ファイルがこれらを手本にした際の違反伝播が負債。

**実装手順**

純粋なコード移動のみ(シグネチャ・本文変更禁止)。(1) chronic_condition_service.go: buildChronicConditionUpdateFields(L143-)をL32のinterface宣言の直前へ移動。(2) audit_service.go: buildAuditLog(L74-)をL12のAuditService interface直前へ移動。`var _ AuditTxLogger = (*auditService)(nil)` (L71, #211コンパイル時保証)はauditService struct直下に残す。(3) lstep_settings_service.go: buildLstepSettingsResponse(L250-)をLstepSettingsService interface(L105)直前へ。(4) lab_import_examination_service.go: buildExamResults(L234-)と、それが呼ぶcomputeExamResultStatusが後方にある場合は一緒にinterface(L74)直前へ移動(未定義参照はGoでは問題ないが可読性のためペア移動)。ドキュメントコメントは関数と一体で移動する。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestChronicCondition|TestAudit|TestLstepSettings|TestLabImport' -count=1
```


## G6. Repository層規約・トランザクション機構整理

### G6-1. P16 命名規約からの逸脱 6 メソッド (Get*/List*) が repository interface に残存 — **CLOSED（2026-07-09）**

- **ID**: `p16-naming-drift`
- **重要度**: P3 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/clinic_repository.go (17, 71); internal/repository/lstep_trigger_priority_repository.go (21, 60); internal/repository/pet_chronic_condition_repository.go (21, 81); internal/repository/lstep_csv_import_repository.go (22, 75); internal/repository/lstep_friend_attribute_snapshot_repository.go (23, 65); internal/repository/lstep_delivery_trigger_log_repository.go (53, 226)
- **依存関係**: なし
- **ステータス**: ✅ **CLOSED** — 6メソッドを機械的リネーム: GetCompany→FindCompany、GetPriority→FindPriorityByTriggerType、GetActiveConditionCodesByOwner→FindActiveConditionCodesByOwner、ListByClinic→FindAllByClinicID、ListByClinicAndDateRange→FindAllByClinicAndDateRange、ListByOwnerAndDateRange→FindAllByOwnerAndDateRange。同名だが別レイヤーの符合(`Handler.GetCompany`、`LstepCsvImportService.ListByClinic`は P16 対象外のため未変更)は grep で個別に検証し除外。accounting の Get*Report/Get*Summary/GetCloseAggregate は集計レポート専用として P16 対象外と判断(実装手順の「一括判断」に従い、対象ファイルリストに元々含まれていないため変更せず)。`docker compose exec backend go build ./internal/repository/ ./internal/service/ ./internal/handler/` 成功、`docker compose exec backend go test ./internal/service/ -run 'TestLstep|TestClinic|TestPetChronic' -count=1` PASS、スコープ限定で `go test ./internal/repository/ -run 'TestClinicRepository|TestLstepTriggerPriorityRepository|TestPetChronicConditionRepository|TestLstepCsvImportRepository|TestLstepFriendAttributeSnapshotRepository|TestLstepDeliveryTriggerLogRepository' -count=1` も PASS。go vet / gofmt 差分なし。コミット `deeabb23`

**証拠(現HEAD検証済み)**

clinic_repository.go:17「GetCompany(ctx context.Context) (*model.Company, error)」、lstep_trigger_priority_repository.go:21「GetPriority(ctx context.Context, clinicID uint64, triggerType string) (int, error)」、pet_chronic_condition_repository.go:21「GetActiveConditionCodesByOwner(ctx context.Context, clinicID, ownerID uint64) ([]string, error)」、lstep_csv_import_repository.go:22「ListByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LstepCsvImport, error)」、lstep_friend_attribute_snapshot_repository.go:23「ListByClinicAndDateRange(...)」、lstep_delivery_trigger_log_repository.go:53「ListByOwnerAndDateRange(...)」。repository/CLAUDE.md P16「FindAll / FindByClinicID ← 一覧（GetAll, List, Fetch は違反）/ FindByID ← 単件（GetByID, Get, Find は違反）」。なお accounting_repository.go:41-45 の GetDailySummary/GetCloseAggregate/GetMonthlyReport(ByPeriod) は entity fetch でなく集計レポート専用のため P16 の適用対象かは規約解釈の余地あり（P16 例示は entity CRUD のみ）。

**問題**

P16 は 2026-04 の be_p16_naming_sync で全体同期済みのはずだが、entity/値 fetch 系 6 メソッドが Get*/List* のまま残存しており、grep ベースの規約チェック・新規実装時の模倣元として一貫性を損なう。機械的リネームで解消可能な純粋な命名負債。

**実装手順**

機械的リネーム（interface 宣言・実装・service 呼出・mock を同時変更）: GetCompany→FindCompany、GetPriority→FindPriorityByTriggerType、GetActiveConditionCodesByOwner→FindActiveConditionCodesByOwner、ListByClinic→FindAllByClinicID、ListByClinicAndDateRange→FindAllByClinicAndDateRange、ListByOwnerAndDateRange→FindAllByOwnerAndDateRange。手順: (1) 各 repo ファイルで interface+実装をリネーム → (2) grep で service/mock の呼出を全件更新 → (3) コンパイルで取りこぼし検出。accounting の Get*Report/Get*Summary/GetCloseAggregate 3+1 件はレポート集計であり P16 の対象外と整理するか含めるかを実装時に一括判断（含める場合は Aggregate*/Summarize* でなく FindDailySummary 等へ）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go build ./internal/repository/ ./internal/service/ ./internal/handler/ && docker compose exec backend go test ./internal/service/ -run 'TestLstep|TestClinic|TestPetChronic' -count=1
```

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


## G7. パフォーマンス (N+1・インデックス・クエリ効率)

### G7-1. LIFF GetAvailableDates の日付ループN+1: シフト/休憩/当日予約/職種ガードを日毎・スタッフ毎に個別クエリ(既定30日×5名で1リクエスト最大約570クエリ) — **CLOSED（2026-07-09）**

- **ID**: `liff-available-dates-nplus1`
- **重要度**: P1 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/service/available_dates.go (109-196); internal/service/liff_service_availability_slots.go (13-68); internal/service/liff_service_availability.go (35-37,119-139); internal/repository/appointment_admin_repository.go (52-73); internal/repository/reservation_schedule_repository.go (17-20,76-87)
- **依存関係**: なし。R2-4(実施済)とは別軸。FE変更不要(レスポンス形状不変)
- **ステータス**: ✅ **CLOSED** — 実装手順(1)〜(5)を実施。ReservationScheduleRepository.FindAllByStaffIDsAndDateRange(複数スタッフ×期間のシフトエントリ一括取得)、ReservationAdminRepository.FindTimeRangesByDateRange(Preloadなしid/doctor_id/start_time/end_time/statusのみの軽量版)、ReservationTypeOccupationRepository.CountWorkingStaffByReservationTypeIDs(GROUP BY se.dateでの日次出勤数バッチ集計)の3新規repositoryメソッドを追加。liffService.buildAvailableDatesStaffInputsFnが予約受付期間全体を3クエリでプリフェッチし、日毎はDB問い合わせ不要な純関数buildStaffSlotInputsFromWindow(マップ参照のみ)を返すクロージャに置換。窓計算はbookingWindowDates()としてavailable_dates.goへ抽出しCalcAvailableDatesと共有。エラー致命度は旧実装を踏襲(シフト/休憩プリフェッチ失敗は非致命的・当日予約プリフェッチ失敗は致命的)。単日パスのGetAvailableTimesは無変更(buildStaffSlotInputsのまま)。GetAvailableDatesは事前テスト0件だったためliff_service_availability_dates_test.go新設+3repositoryメソッドのDB実行テストを追加。`docker compose exec backend go test ./internal/service/ -run 'TestCalcAvailableDates|TestBuildStaffSlotInputs|TestGetAvailable' -count=1` と `go test ./internal/repository/ -run 'TestReservationSchedule|TestReservationTypeOccupation' -count=1` 全PASS。go vet(全リポジトリ)/gofmt差分なし。go-reviewer/security-reviewer/clinic-isolation-auditor三者ともApprove(CRITICAL/HIGH/MEDIUM 0、raw SQLパラメータ化・clinic_id境界を個別検証済み)。コミット `755f3e42`

**証拠(現HEAD検証済み)**

available_dates.go:109 `for d := minDate; !d.After(maxDate); d = d.AddDate(0, 0, 1) {` → :152 `staffInputs, err := input.StaffInputsFn(ctx, d, input.TypeID, input.StaffID)` が開催日ごとに liff_service_availability.go:35-37 の closure 経由で buildStaffSlotInputs を呼ぶ。buildStaffSlotInputs は liff_service_availability_slots.go:15 `dayResv, err := s.adminRepo.FindAllByDay(ctx, clinicID, date)`(コメント:14「当日の全予約を一括取得（N+1回避）」は日内のみのバッチ)、:26 `entry, err := s.scheduleRepo.FindAllByDate(ctx, clinicID, staffs[i].ID, date)`(スタッフ毎)、:32 `breaks, brkErr := s.scheduleRepo.FindAllBreaksByEntryID(ctx, entry.ID)`(エントリ毎)。FindAllByDay は appointment_admin_repository.go:59-64 で `Preload("ReservationType"...).Preload("Doctor"...).Preload("CreatedByStaff"...).Preload("LineCustomer").Preload("Owner"...).Preload("Pet"...)` の6 Preload を伴うが、slot計算は Status/DoctorID/StartTime/EndTime しか使わない(:53-63)。さらに liff_service_availability.go:120-138 で available な日付ごとに :129 `count, err := s.occupationRepo.CountWorkingStaffByReservationTypeID(ctx, clinicID, typeID, date)`。1リクエストのクエリ数 = 日数D×(1+Preload6+スタッフS×2) + 職種ガード≤D。booking_window_max_days 既定30(001_init.sql:2598)・S=5 で 30×17+30≈540。R2-4(D8)で解消済みなのは slot容量カウントのみ(liff_service_availability.go:64-66 コメント「日付間の反復は CalcAvailableDates の制御下にあり残る」)で、シフト/休憩/予約/職種の日付軸N+1は未着手。

**問題**

LIFF予約カレンダーは飼い主向けの最頻アクセス経路であり、1画面表示で数百クエリはRDS負荷・p95レイテンシに直結する。shift_entries は UNIQUE(clinic_id, staff_id, date)(001_init.sql:2016 uk_shift_staff_date)で範囲一括取得が容易、休憩は既にバッチAPI FindAllBreaksByEntryIDs(reservation_schedule_repository.go:18)が存在するのに未使用で、インフラは揃っているのに呼出構造だけがN+1のまま。

**実装手順**

挙動保存のプリフェッチ化。(1) ReservationScheduleRepository に FindAllByStaffIDsAndDateRange(ctx, clinicID uint64, staffIDs []uint64, from, to time.Time) ([]model.ShiftEntry, error) を追加(WHERE staff_id IN ? AND date >= ? AND date < ? + clinicScope。FindAllByMonth :31-51 と同型)。(2) 既存 FindAllBreaksByEntryIDs で全エントリの休憩を1クエリ取得。(3) reservationAdminRepository に Preload なしの軽量 FindTimeRangesByDateRange(ctx, clinicID, from, to)(Select: id, doctor_id, start_time, end_time, status のみ)を追加。(4) GetAvailableDates(liff_service_availability.go:17)で CalcAvailableDates 呼出前に minDate/maxDate ぶんを3クエリでプリフェッチし map[dateStr] 化、staffInputsFn closure をマップ参照に差し替え。buildStaffSlotInputs はプリフェッチ済みデータを受ける buildStaffSlotInputsFromWindow に分離し、単日パス GetAvailableTimes(:197)は従来メソッド維持(挙動不変)。(5) 職種ガードは CountWorkingStaffByReservationTypeIDs(ctx, clinicID, typeID, dates []time.Time) (map[string]int64) を追加し GROUP BY se.date で1クエリ化(reservation_type_occupation_repository.go:104-125 のSQLに GROUP BY 追加)。minDate/maxDate は CalcAvailableDates 内で再計算しているため、窓計算ヘルパー(BookingWindowMinDays/MaxDays→from,to)を available_dates.go に抽出して両者で共有する。既存テスト TestCalcAvailableDates / TestBuildStaffSlotInputs を先に緑確認→リファクタ→緑維持。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestCalcAvailableDates|TestBuildStaffSlotInputs|TestGetAvailable' -count=1 && docker compose exec backend go test ./internal/repository/ -run 'TestReservationSchedule|TestReservationTypeOccupation' -count=1
```

### G7-2. 健診同期プレビューのタグキャッシュN+1: LINE連携飼主N人ごとに FindByOwner を1回(プレビュー行はLIMITなし全件) — **CLOSED（2026-07-09）**

- **ID**: `checkup-sync-preview-tagcache-nplus1`
- **重要度**: P2 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/service/checkup_sync_service_preview.go (46-88); internal/repository/lstep_tag_cache_repository.go (39-47,115); internal/repository/checkup_sync_repository.go (74-149)
- **依存関係**: なし。エラーパスの粒度変更のみテストで仕様固定が必要
- **ステータス**: ✅ **CLOSED** — LstepTagCacheRepository に `FindByOwners(ctx, clinicID, ownerIDs) (map[uint64][]*model.LstepTagCache, error)` を新設(ownerIDs空なら空map即返し・WHERE clinic_id = ? AND owner_id IN ?)。PreviewCheckupSync はループ前に LINE連携済み owner_id を収集し1クエリで一括取得、ループ内はマップ参照に変更。エラーパスの粒度差(per-owner失敗→全員失敗)は実装手順どおりテストで仕様固定: `TestCheckupSyncService_PreviewCheckupSync_TagCacheLookupError` を2 owner・一括失敗ケースに書き換え。独立した3つのテストモック(`mockLstepTagCacheRepository`/`mockTagCacheRepoForDelivery`/`mockTagCacheSummaryRepo`)全てに `FindByOwners` を追加し `go vet ./...` で取りこぼしゼロを確認。DB実行の repository テスト `TestLstepTagCacheRepository_FindByOwners` を新規追加(owner_id別集約・clinic_id分離・空引数即返しの3ケース)。`docker compose exec backend go test ./internal/service/ -run 'TestCheckupSync' -count=1` と `go test ./internal/repository/ -run 'TestLstepTagCacheRepository_FindByOwners' -count=1` 全PASS。go vet(全リポジトリ) / gofmt 差分なし。コミット `f6ebc0c7`

**証拠(現HEAD検証済み)**

checkup_sync_service_preview.go:46 `for i := range rows {` → :75-76 `if hasLine {\n\t\t\tcached, cacheErr := s.tagCacheRepo.FindByOwner(ctx, clinicID, row.OwnerID)`。rows は FindCheckupSyncPreview(checkup_sync_repository.go:74)由来で LIMIT なし(同ファイルに Limit 出現なし・grep確認)＝フィルタ該当の全飼主。LINE連携飼主N人のクリニックで 1+N クエリ/リクエスト(N=3000なら3001)。LstepTagCacheRepository のインターフェース(lstep_tag_cache_repository.go:39-47)には単件 FindByOwner のみでバッチ版が無い。idx_lstep_tag_cache_owner (clinic_id, owner_id)(001_init.sql:393)は IN 一括を支える。

**問題**

管理画面のプレビューは配信前に何度も条件を変えて叩く操作であり、飼主数に線形なクエリ数はクリニック成長とともに悪化する。バッチ取得への構造(複合インデックス・map返却の同型API前例 FindAllBreaksByEntryIDs)は既にあり、呼出形状だけがN+1。

**実装手順**

(1) LstepTagCacheRepository に FindByOwners(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error) を追加(len==0 で空map即返し・WHERE clinic_id = ? AND owner_id IN ?。reservation_schedule_repository.go:53-66 FindAllBreaksByEntryIDs と同型)。(2) PreviewCheckupSync でループ前に hasLine 判定を先に回して該当 OwnerID を収集→一括取得→ループ内はマップ参照。(3) エラーパスの挙動差に注意: 現状は per-owner 失敗時に該当飼主のみ空タグ+slog(:77-79)。一括版では失敗時に全員空タグ+slog になる — 成功パスは挙動同一、エラーパスのみの差なので既存テスト TestCheckupSyncService_PreviewCheckupSync_TagCacheLookupError を一括失敗ケースに書き換えて仕様として固定する。(4) 同型の per-owner FindByOwner ループを持つ lstep_tag_sync_visit_ltv.go:69 等のバッチジョブへの展開は別追跡(外部API律速のため効果小)。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestCheckupSyncService_PreviewCheckupSync' -count=1 && docker compose exec backend go test ./internal/repository/ -run 'TestLstepTagCache' -count=1
```

### G7-3. 月次レポート/レジ締め集計の completed_at 述語が非sargable(AT TIME ZONE 式適用)で partial index を無効化 — 同一関数内のCTE形式と2形式混在 — **CLOSED（2026-07-09）**

- **ID**: `reports-completed-at-non-sargable`
- **重要度**: P2 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/accounting_repository_reports_monthly.go (154-155,177-178); internal/repository/accounting_repository_reports_close.go (97-98,162-163); internal/config/config.go (104-105)
- **依存関係**: 前提=DSN TimeZone=Asia/Tokyo 固定(config.go:104)。この前提が崩れる変更とは排他
- **ステータス**: ✅ **CLOSED** — monthly.go 2箇所(税率別集計・会計件数)、close.go 2箇所(個別会計一覧CTE・税率別集計)を `completed_at AT TIME ZONE 'Asia/Tokyo' >= ?/< ?` から CTE本体と同型の `completed_at >= ?/< ?` に統一。close.go の書き換え2箇所は既存 cArgs と同様に `.In(time.Local)` を揃えた。等価性は temp-revert 方式ではなく pre/post 比較で実証: 変更前に `TestAccountingRepository_GetMonthlyReport_*`(BillingCount/TaxBreakdown 具体値アサート込み)・`TestAccountingRepository_GetCloseAggregate_*`(BillingDetails/TaxBreakdown 具体値アサート込み)全11本が GREEN であることを確認 → 述語変更 → 同一テスト群が同一アサートのまま再度全PASS(結果不変を実証)。TO_CHAR(...AT TIME ZONE...) による日付文字列 GROUP BY 出力(monthly.go:40,61,92)はフィルタ述語ではなくフォーマット用途のため対象外・無変更。go vet / gofmt 差分なし。コミット `904b0bc2`

**証拠(現HEAD検証済み)**

accounting_repository_reports_monthly.go:154-155 `Where("billings.completed_at AT TIME ZONE 'Asia/Tokyo' >= ?", start).\n\t\tWhere("billings.completed_at AT TIME ZONE 'Asia/Tokyo' < ?", end).`(税率別集計)と :177-178(会計件数Count)、accounting_repository_reports_close.go:97-98(個別会計一覧CTE)・:162-163(税率別集計)が completed_at 列に式を適用。一方、同じ関数の期間本体は monthly.go:26-27 / close.go:23-24 で `AND completed_at >= ?\n\t\t  AND completed_at < ?` と直接比較。001_init.sql:2364-2367 `-- PERF-01: 月次レポート・締め集計最適化\nCREATE INDEX idx_billings_clinic_completed_at\n  ON billings(clinic_id, completed_at)\n  WHERE deleted_at IS NULL AND status = 'completed';` — 式適用側はこの complete_at レンジスキャンを使えず、clinic の全 completed 会計を履歴比例でスキャンする。セッションTZは config.go:104-105 `"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s", ... JapanTimeZone`(=Asia/Tokyo, timezone.go:9)で固定。

**問題**

月次レポート・締めプレビュー/確定は毎日の運用経路。会計履歴が積み上がるほど式適用側4クエリだけ線形劣化する。さらに同一集計内で期間述語が2形式混在しており、集計本体(payment/category/refund)と明細/税率で「同じ期間」の判定方法が異なるのは保守上も危うい(現構成では DSN TimeZone=Asia/Tokyo 固定のため両者は同値)。

**実装手順**

(1) まず既存テスト(accounting_repository_reports_monthly_test.go / reports_close 系)に「AT TIME ZONE 形式と直接比較形式で件数・合計が一致する」assert を追加して現挙動を固定。(2) 4箇所を CTE と同じ `completed_at >= ? AND completed_at < ?` に書き換え(monthly は start/end、close は input.PeriodStart/PeriodEnd をそのまま渡す。close.go:19 の cArgs と同様に .In(time.Local) を揃える)。(3) 挙動保存の根拠: セッション TimeZone=Asia/Tokyo が DSN で固定されているため timestamptz 直接比較と AT TIME ZONE 'Asia/Tokyo' 経由比較は同一瞬間を指す — この前提を書き換え箇所のコメントに明記する。テスト先行(temp-revert RED方式)で等価性を実証してから述語を統一すること。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestGetMonthlyReport|TestGetCloseAggregate|TestAccountingRepository_Reports' -count=1
```

### G7-4. checkups(clinic_id,date)/(clinic_id,next_date)・vaccinations(clinic_id,date) の複合インデックス欠落 — 日付レンジ一覧/期限アラートの主要WHERE句と不整合 — **CLOSED（2026-07-09）**

- **ID**: `checkups-vaccinations-missing-composite-index`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/checkup_repository.go (44-70,107-124); internal/repository/vaccination_repository.go (35-70); migrations/001_init.sql (2258-2264,2400)
- **ステータス**: ✅ **CLOSED** — 新規migration `migrations/002_add_checkup_vaccination_indexes.sql` で3本追加: `idx_checkups_clinic_date`(clinic_id,date)・`idx_checkups_clinic_next_date`(clinic_id,next_date, next_date IS NOT NULL)・`idx_vaccinations_clinic_date`(clinic_id,date)。3本とも既存 idx_billings_clinic_date 等と同じ partial index(`WHERE deleted_at IS NULL`)流儀。001_init.sql は適用済みのため追記せず新規ファイルで追加(checksum mismatch回避、migrations/CLAUDE.md準拠)。checkup_repository.go/vaccination_repository.go のクエリ形状は既に clinic_id+date/next_date で正しく、インデックスのみが欠落していたためリポジトリコード変更なし。既存単列 idx_checkups_clinic_id/idx_vaccinations_clinic_id は複合の prefix に包含されるため DROP 候補だが、EXPLAIN確認後の別コミットに先送り(本コミットは追加のみ)。`docker compose exec backend go test ./internal/repository/ -run 'TestCheckupRepository|TestVaccinationRepository' -count=1` PASS。CASCADE lint(`TestMigrationCascadeInventory_NoUnreviewedCascade`/`TestReconcileMigrationCascade_Analyzer`)PASS(CASCADE定義なし)。STG適用はdb_reset不要(新規migration追加のみ)。コミット `dd3f31b8`
- **依存関係**: STGは新規migration適用のみ(db_reset不要)。checkup-list-unbounded の是正と独立だが同時に入れると効果測定しやすい

**証拠(現HEAD検証済み)**

checkup_repository.go:53-65 は clinicScope + `q.Where("date >= ?", ...)` / `q.Where("next_date >= ?", ...)` + `Order("date DESC")`、:116-118 FindAlerts は `Where("next_date IS NOT NULL").\n\t\tWhere("next_date <= ?", upperBound).\n\t\tOrder("next_date ASC")`。vaccination_repository.go:45-50 は `q.Where("vaccinations.date >= ?", ...)` + :65 `Order("vaccinations.date DESC...")`。対して 001_init.sql の該当インデックスは :2400 `CREATE INDEX idx_checkups_clinic_id ON checkups(clinic_id);` :2262-2264(medical_record_id/pet_id/checkup_type_id 単列) :2258 `CREATE INDEX idx_vaccinations_clinic_id ON vaccinations(clinic_id);` のみで、clinic_id+date 複合が無い。同型の一覧クエリを持つ medical_records(:2330 idx_medical_records_clinic_date)・billings(:2352 idx_billings_clinic_date)・appointments(:2317 idx_appointments_clinic_date)には複合が既設で、checkups/vaccinations だけ欠けている。

**問題**

健診一覧(日付/次回日フィルタ)・期限アラート・ワクチン接種歴一覧はレコードが診療とともに単調増加するテーブルへの日付レンジ+ソートクエリ。単列 clinic_id インデックスでは クリニック全行フェッチ+ソートになり履歴比例で劣化する。既存スキーマの設計意図(主要トランザクションテーブルは clinic+date 複合)とも不整合。

**実装手順**

新規マイグレーションファイル(002_add_checkup_vaccination_indexes.sql — 適用済み001への追記は checksum mismatch のため禁止、migrations/CLAUDE.md 準拠)で3本追加: `CREATE INDEX idx_checkups_clinic_date ON checkups(clinic_id, date) WHERE deleted_at IS NULL;` / `CREATE INDEX idx_checkups_clinic_next_date ON checkups(clinic_id, next_date) WHERE deleted_at IS NULL AND next_date IS NOT NULL;` / `CREATE INDEX idx_vaccinations_clinic_date ON vaccinations(clinic_id, date) WHERE deleted_at IS NULL;`(partial 条件は既存 idx_billings_clinic_date 等の流儀に合わせる)。既存単列 idx_checkups_clinic_id / idx_vaccinations_clinic_id は複合が prefix を包含するため DROP 候補だが、削除は EXPLAIN での利用確認後に別コミットで行う(まず追加のみ)。CASCADE lint / migration CI(R3-1)の通過を確認。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestCheckupRepository|TestVaccinationRepository' -count=1
```

### G7-5. BulkAddOwnerTag が既存バッチAPI FindByIDs を使わず per-owner FindByID ループ(PERF-04コメントと実装が乖離) — **CLOSED（2026-07-09）**

- **ID**: `bulk-add-tag-findbyids-unused`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/service/lstep_tag_service.go (302-313); internal/repository/owner_repository.go (42-44)
- **依存関係**: なし
- **ステータス**: ✅ **CLOSED** — ループを `s.ownerRepo.FindByIDs(ctx, clinicID, ownerIDs)` に置換し map[uint64]*model.Owner を構築。取得できなかったID(要求−取得の集合差)は ownerIDs 順に走査して従来と同じ順序で FailedOwnerIDs へ追加。エッジケース(FindByIDs 自体が失敗)も挙動保存: 旧 per-owner ループは NotFound/DBエラーを区別せず当該 owner を FailedOwnerIDs に積み処理継続していたため、FindByIDs 失敗時も早期returnせずログのみで続行し全 ownerIDs を FailedOwnerIDs に積む(新規テスト `TestBulkAddOwnerTag_FindByIDsError` で固定)。PERF-04コメントを実装に一致させた。`docker compose exec backend go test ./internal/service/ -run TestBulkAddOwnerTag -count=1` 全PASS(既存 MixedResults はモックフック `findByIDFn`→`findByIDsFn` に追随)。go vet / gofmt 差分なし。コミット `93867fcc`

**証拠(現HEAD検証済み)**

lstep_tag_service.go:302-306 `// PERF-04: Cache owners in memory to avoid N+1 queries (1 per owner → 1 total)\n\t// Build owner map: fetch all owners upfront\n\townerMap := make(map[uint64]*model.Owner)\n\tfor _, ownerID := range ownerIDs {\n\t\towner, findErr := s.ownerRepo.FindByID(ctx, clinicID, ownerID)` — コメントは「1 total」を謳うが実装は ownerIDs 1件につき1クエリのまま。owner_repository.go:42-44 には `// FindByIDs は複数 ID でオーナーを一括取得する（タグ一括同期の N+1 解消用）。\n\t// Preload なしの軽量クエリ。` と、まさにこの用途向けのバッチAPIが既存だが本メソッドでは未使用。

**問題**

N+1が実在する上に、PERF-04コメントが「解消済み」と誤認させる負債(過去監査の「grep額面信用禁止」教訓の実例)。もっとも本経路は owner ごとの外部 Lstep API 呼び出し(:328 client.AddTag)が支配的なため実効インパクトは小さい — 主目的はコメントと実装の整合回復。

**実装手順**

(1) ループを `owners, err := s.ownerRepo.FindByIDs(ctx, clinicID, ownerIDs)` に置換し map[uint64]*model.Owner を構築。(2) 取得できなかったID(要求 - 取得の集合差)を従来どおり slog + result.FailedOwnerIDs へ追加して挙動保存(現状 :307-311 の per-owner NotFound 処理と同結果)。順序依存に注意: FailedOwnerIDs への追加順が ownerIDs 順になるよう集合差判定は ownerIDs を走査して行う。(3) PERF-04 コメントを実装に一致させる。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run TestBulkAddOwnerTag -count=1
```


## G8. Go設計イディオム (ジェネリクス・定数集約)

### G8-1. マスタ系 repository の FindByID/Update/Delete がトークン同一のまま24〜31ファイルに複製 — Go 1.25 ジェネリクスで helpers.go に畳める — **CLOSED（2026-07-10）**

- **ID**: `repo-master-crud-generic-fold`
- **重要度**: P2 / **工数目安**: L / **挙動変更**: なし（挙動保存）
- **ステータス**: ✅ **CLOSED** — Update/Delete は G3-3 で既に `updateScopedByID`/`deleteScopedByID` に委譲済み（本チケットでは再実装せず）。残スコープの FindByID インライン畳み込みを `findByIDScoped[T]` ジェネリクスヘルパー新設で実施。Preload 付き FindByID（P3.1 契約）・dbOrTx 参加 repo は対象外として除外。対象18ファイル（diagnosis_repository.go は diagnosisNameRepository のみ、diagnosisTypeRepository は Preload のため除外）を2 batch に分割: cage/checkup_type/chief_complaint/consultation/diagnosis_name/hospitalization_plan/inquiry_template/insurance/inventory_item（コミット `0fe8fe78`）、medicine/merchandise_item/occupation/procedure/reservation_type_group/reservation_type_liff/trimming_course/trimming_option/vaccine（コミット `d09cfa4c`）。`TestFindByIDScoped`（成功/NotFound/クロステナント）新設、interface シグネチャ・mock・service層は無変更。`docker compose exec backend go test ./internal/repository/ -count=1` 両batch後 ok、`go vet ./...` 両batch後 exit 0。
- **対象ファイル**: internal/repository/vaccine_repository.go (43-82); internal/repository/medicine_repository.go (121-129); internal/repository/procedure_repository.go (69-77); internal/repository/cage_repository.go (72-80); internal/repository/insurance_repository.go (70-78); internal/repository/helpers.go (14-32)
- **依存関係**: なし。既存 preload lint / master-fk write lint は Preload・service シグネチャ非変更のため影響なし

**証拠(現HEAD検証済み)**

Delete はモデル型とリソース名文字列以外バイト同一。vaccine_repository.go:73-82:
```go
func (r *vaccineRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Vaccine{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "vaccine", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("vaccine", fmt.Sprintf("%d", id))
	}
	return nil
}
```
medicine_repository.go:122 / procedure_repository.go:70 / cage_repository.go:73 / insurance_repository.go:71 も `result := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("id = ?", id).Delete(&model.Xxx{})` で完全同型。定量: この Delete 完全一致が24ファイル、`Scopes(clinicScope(clinicID)).Where("id = ?", id)` を含むファイル35、Update 末尾の `return r.FindByID(ctx, clinicID, id)` パターンが31箇所。既に helpers.go:14 `func reorderByClinicID(ctx context.Context, db *gorm.DB, model any, resource string, clinicID uint64, ids []uint64) error` が「リソース名文字列を渡す共有ヘルパー」の先例を確立済み。

**問題**

P4（Update/Delete への clinicScope 必須）は read-Preload lint と違い静的機械強制が無く、レビュー・runtime isolation test 頼み。24〜31個の手書きコピーは新規マスタ追加のたびに clinicScope / RowsAffected==0→WrapNotFound / FromGORM の3点セットを手で再現させる構造で、1点でも落とすとクロステナント更新/削除（過去に実バグ化したクラス）に直結する。重複量も実測で約 500-700 行。

**実装手順**

挙動保存。手順: (1) internal/repository/helpers.go（または新規 internal/repository/crud_helpers.go）に reorderByClinicID と同スタイルのジェネリクスヘルパーを追加:
```go
func findByIDScoped[T any](db *gorm.DB, resource string, clinicID, id uint64) (*T, error)
func updateScopedFields[T any](db *gorm.DB, resource string, clinicID, id uint64, fields map[string]any) error // RowsAffected==0 → WrapNotFound
func deleteScoped[T any](db *gorm.DB, resource string, clinicID, id uint64) error
```
第1引数は呼び出し側が `r.db.WithContext(ctx)` を渡す（dbOrTx を使う repo はそれを渡す）— 各 repo の現在の db ハンドル選択を変えないことで挙動保存を担保。(2) Delete バイト同一の24ファイルから機械的に置換開始（vaccine/medicine/procedure/cage/insurance/consultation 等のマスタ系）。(3) Update は「Updates(fields)→RowsAffected==0→WrapNotFound→FindByID再取得」の31箇所のうち、FindByID に Preload や追加ロジックが無いものだけ `updateScopedFields` + 既存 `r.FindByID` 呼び出しの2行に畳む。差分のある repo（Preload付きFindByID・soft delete特殊処理・JOINスコープ repo）は対象外とし、無理に一般化しない。(4) per-repo interface（P16命名・mock seam）とメソッドシグネチャは一切変更しない — 畳むのはメソッド本体のみ。よって service 層・mock・テストは無変更。(5) 1バッチ=5〜10 repo 単位でコミットし、都度 repository パッケージのテストで検証。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -count=1
```

### G8-2. JST タイムゾーンの導出が4方式×8箇所に分裂（time.FixedZone("Asia/Tokyo", 9*60*60) マジックオフセット6箇所 + LoadLocation 2箇所 + config.ConfigureTimeZone） — **CLOSED（2026-07-10）**

- **ID**: `jst-location-derivation-scatter`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **ステータス**: ✅ **CLOSED** — `config.JST` を新設し8サイトの再導出（accounting_report_service.go / accounting_report_request.go / lstep_delivery_monitor_request.go / lstep_batch_delivery.go / accounting_service_reports.go×2 / reservation_type_occupation_repository.go / available_dates.go）を置換。コミット `ecaac5ec`（Session 3 以前に実装済み、本 Epic では DOC-G8 としてクローズ記録のみ）。`cmd/api/main.go:199,222` の FixedZone 残存は G9-3 スコープへ引き継ぎ。
- **対象ファイル**: internal/config/timezone.go (9-18); internal/service/accounting_report_service.go (18); internal/service/accounting_service_reports.go (19, 122); internal/service/lstep_batch_delivery.go (45); internal/handler/accounting_report_request.go (11); internal/handler/lstep_delivery_monitor_request.go (81); internal/repository/reservation_type_occupation_repository.go (16-23); internal/service/available_dates.go (215-224)
- **依存関係**: なし

**証拠(現HEAD検証済み)**

正本は config/timezone.go:9-16:
```go
const JapanTimeZone = "Asia/Tokyo"

func ConfigureTimeZone() error {
	loc, err := time.LoadLocation(JapanTimeZone)
	...
	time.Local = loc
```
（cmd/api/main.go:37 ほか全6 cmd が起動時に呼ぶ = time.Local は常に JST）にもかかわらず、accounting_report_service.go:18 `var jst = time.FixedZone("Asia/Tokyo", 9*60*60)`、accounting_report_request.go:11 `var monthlyReportJST = time.FixedZone("Asia/Tokyo", 9*60*60)`、lstep_delivery_monitor_request.go:81 / lstep_batch_delivery.go:45 / accounting_service_reports.go:19,122 の関数ローカル `jst := time.FixedZone("Asia/Tokyo", 9*60*60)`、reservation_type_occupation_repository.go:17-22 `var jstLoc = func() *time.Location { loc, err := time.LoadLocation("Asia/Tokyo"); if err != nil { panic(...) }; return loc }()`、available_dates.go:218-224 `var jstLocation = func() *time.Location { loc, err := time.LoadLocation("Asia/Tokyo"); if err != nil { return time.FixedZone("JST", 9*60*60) }; return loc }()` — 同一概念が var/ローカル/即時関数、panic/フォールバック/エラー不能と、命名も失敗時挙動もバラバラに8回再導出されている。

**問題**

マジックナンバー 9*60*60 が6箇所に散在し、失敗時セマンティクスが3種（panic / silent fallback / 失敗不能）に割れている。tzdata は config/timezone.go:6 で埋め込み済み（`_ "time/tzdata"`）なので LoadLocation 失敗ケースの防御コードは全て到達不能なのに各所で別々に書かれている。将来「JSTを参照する処理」を触るたびにどの導出を真似るべきか判断が必要になる。

**実装手順**

挙動保存（JST は DST 無しのため FixedZone(+9h) と LoadLocation("Asia/Tokyo") は wall-clock 完全同値。ただし置換前に各サイトの Format verb に zone略称 "MST" が無いことを確認する — 現状 "2006-01-02" 系と RFC3339("Z07:00"→+09:00) のみで略称出力なし）。手順: (1) config/timezone.go に `var JST = func() *time.Location { ... }()`（LoadLocation + tzdata埋め込み済みなので失敗時 panic で早期発見）をエクスポート。(2) 上記8サイトの再導出を `config.JST` 参照に置換し、ローカル var jst / jstLoc / jstLocation / monthlyReportJST を削除。(3) handler/service/repository → config の import 方向は既存（config.BcryptCost 等で service が config を import 済み）なので循環なし。影響範囲: 上記7ファイルのみ。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/service/ -run 'TestAccountingReport|TestLstep|TestAvailableDates' -count=1
```

### G8-3. 日付レイアウト "2006-01-02" が非テストコード259箇所にリテラル散在し、同値の file-local 定数3種 + time.DateOnly 2箇所と表記が分裂 — **CLOSED（2026-07-10）**

- **ID**: `date-layout-literal-scatter`
- **重要度**: P3 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **ステータス**: ✅ **CLOSED** — handler → service → repository の3コミットで `time.DateOnly`/`time.RFC3339` へ統一。コミット `2d5c180d`（handler）/ `9bae5e8a`（service）/ `bcc34d5d`（repository）。3 file-local 定数（shiftDateLayout/chronicConditionDateLayout/cashRegisterDateLayout）削除済み。本 Epic では DOC-G8 としてクローズ記録のみ。
- **対象ファイル**: internal/handler/shift_request.go (11); internal/handler/chronic_condition_request.go (10); internal/handler/cash_register_request.go (11); internal/service/lab_report_query_service.go (91, 129); internal/handler/shift_response.go (60-61)
- **依存関係**: jst-location-derivation-scatter と同一ファイルに触れるため同時期に実施すると conflict が減る（順序任意）

**証拠(現HEAD検証済み)**

同一レイアウトが3表記に分裂: (a) リテラル259箇所・79ファイル（例 handler/lstep_delivery_monitor_request.go:82 `defaultDate := now.In(jst).Format("2006-01-02")`）。(b) file-local 定数の再宣言3種 — shift_request.go:11 `const shiftDateLayout = "2006-01-02"`、chronic_condition_request.go:10 `const chronicConditionDateLayout = "2006-01-02"`、cash_register_request.go:11 `const cashRegisterDateLayout = "2006-01-02"`。(c) stdlib 定数はわずか2箇所 — lab_report_query_service.go:91 `Date: e.Date.Format(time.DateOnly),`。付随して shift_response.go:60-61 は RFC3339 をリテラル再表記: `CreatedAt: s.CreatedAt.In(time.Local).Format("2006-01-02T15:04:05Z07:00"),`（= time.RFC3339 と同値）。

**問題**

Go 1.20+ の time.DateOnly（値は "2006-01-02" と完全一致）が存在するのに、同値文字列が259回タイプされ、3ファイルは同値定数を別名で再発明している。レイアウト文字列は typo しても compile も lint も通らず実行時に壊れるクラスで、canonical 表記1つに寄せることが typo-drift の構造的防止になる。時刻フォーマットは apicontract の date-format drift gate の監視対象でもあり、表記統一は同 gate の可読性にも寄与する。

**実装手順**

挙動保存（time.DateOnly == "2006-01-02"、time.RFC3339 == "2006-01-02T15:04:05Z07:00" の同値置換）。手順: (1) 機械置換: `Format("2006-01-02")` / `Parse("2006-01-02", ...)` / `ParseInLocation("2006-01-02", ...)` の "2006-01-02" を time.DateOnly に、shift_response.go:60-61 の RFC3339 リテラルを time.RFC3339 に置換（sed 後に目視レビュー。"2006-01-02 15:04:05" 等の複合レイアウトは対象外として残す）。(2) shiftDateLayout / chronicConditionDateLayout / cashRegisterDateLayout の3定数を削除し参照を time.DateOnly へ。(3) time 未import のファイルは goimports で解決。(4) パッケージ単位（handler → service → repository）で3コミットに分割。(5) 完了判定: `grep -rn '"2006-01-02"' backend/internal --include='*.go' | grep -v _test` が複合レイアウト以外 0 件。テストファイル内の同リテラルは任意（本計画のスコープ外で可）。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/handler/ -count=1 && docker compose exec backend go test ./internal/service/ -count=1
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

### G9-2. 環境変数読取がconfig.Load外に分散（STORAGE_TYPE/S3_BUCKET/S3_REGION/TRUSTED_PROXY_CIDR/LOG_LEVEL/CORS_ALLOWED_ORIGIN）し、S3系の起動時検証に漏れがある — **CLOSED（2026-07-10）**

- **ID**: `env-config-consolidation`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **ステータス**: ✅ **CLOSED** — コミット `c903ab3b`。Config に6フィールド追加、TRUSTED_PROXY_CIDR(release必須)とSTORAGE_TYPE=s3時のS3_BUCKET/S3_REGION必須検証をValidate()へ移設（fail-fast結果は同一、発火タイミングが早まるのみ）。middleware.CORS()はallowedOriginを引数化。S3SharedBucket検証追加は挙動変更のためスコープ外のまま維持。
- **対象ファイル**: cmd/api/main.go (27,51-56,131,171-177,269); internal/middleware/cors.go (13-18); internal/config/config.go (47-100)

**証拠(現HEAD検証済み)**

cmd/api/main.go:171-177:
	if os.Getenv("STORAGE_TYPE") == "s3" {
		s3Bucket := os.Getenv("S3_BUCKET")
		s3Region := os.Getenv("S3_REGION")
		if s3Bucket == "" || s3Region == "" {
			logger.Error("S3_BUCKET and S3_REGION are required when STORAGE_TYPE=s3")
			os.Exit(1)

cmd/api/main.go:132: s3fs, err := infra.NewS3FileStorage(context.Background(), cfg.S3SharedBucket, cfg.S3SharedRegion, cfg.S3Endpoint)
（cfg.S3SharedBucket が空でも Validate/この分岐は通過し、最初のアップロードまで検出されない）

internal/middleware/cors.go:14: allowedOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
cmd/api/main.go:52: if os.Getenv("TRUSTED_PROXY_CIDR") == "" {

**問題**

config.Load/Validate が設定の単一窓口である設計（config.go）に対し、6変数が main.go/cors.go/liff_auth.go で直接 os.Getenv され、必須性検証（TRUSTED_PROXY_CIDR release必須、S3_BUCKET/S3_REGION の STORAGE_TYPE=s3 時必須）が main.go に分散している。また STORAGE_TYPE=s3 のとき S3_SHARED_BUCKET（shared_files用）は未検証で、空のまま S3 クライアントが生成され実行時初回アップロードまで失敗が遅延する。

**実装手順**

手順: (1) Config に StorageType/S3Bucket/S3Region/TrustedProxyCIDR/LogLevel/CORSAllowedOrigin フィールドを追加し Load で読取。(2) main.go:52-56（TRUSTED_PROXY_CIDR release検証）と main.go:174-177（S3_BUCKET/S3_REGION検証）を Validate 内へ移設（検証内容・失敗時exitは同一のため挙動保存）。(3) middleware.CORS() を CORS(allowedOrigins string) に変更し main.go から cfg 値を注入（デフォルト値ロジックは Load 側へ移動）。liff_auth.go:52 の LIFF_MOCK は release ガード付きで現状維持可。(4) config_validate_test.go にケース追加。任意の追加（挙動変更のため別判断）: StorageType=="s3" 時の S3SharedBucket 非空検証を Validate に追加すると起動時 fail-fast になる。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/config/ ./internal/middleware/ -count=1
```

### G9-3. main.goの3バッチgoroutineがタイマーループをコピペ重複し、JST取得イディオムも3様（FixedZone×2 / 素のtime.Now×1） — **CLOSED（2026-07-10）**

- **ID**: `batch-scheduler-dedup-tz`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **ステータス**: ✅ **CLOSED** — コミット `72d64ae7`。`cmd/api/batch_scheduler.go` に `runScheduled`/`hourlyTick`/`dailyAt2AM` を抽出、3 goroutine を呼び出し3行へ置換。JST は `time.Local`（ConfigureTimeZone 済み）に統一、`FixedZone("Asia/Tokyo"` 残存は `backend/cmd/` で0件。`TestRunScheduled`/`TestHourlyTick`/`TestDailyAt2AM` 新設。
- **対象ファイル**: cmd/api/main.go (197-259); internal/config/timezone.go (11-18)

**証拠(現HEAD検証済み)**

cmd/api/main.go:198-199:
	go func() {
		jst := time.FixedZone("Asia/Tokyo", 9*60*60)
（221-222 で同一の FixedZone を再定義）

main.go:244-247:
	go func() {
		for {
			now := time.Now()
			next := now.Truncate(time.Hour).Add(time.Hour)
（3本目のみロケーション指定なし）

internal/config/timezone.go:12-16:
	loc, err := time.LoadLocation(JapanTimeZone)
	...
	time.Local = loc
（main.go:37 で起動時に time.Local=Asia/Tokyo 確定済みのため FixedZone 再定義は冗長）

**問題**

select{ctx.Done/timer.C} + Truncate/翌日02:00計算のスケジューラループが3回コピペされ、JSTの取得方法が time.FixedZone 手書き（tzdata の Asia/Tokyo と定義が二重化）と暗黙の time.Local で揺れている。日本は現行DSTなしのため挙動差は生じないが、4本目のバッチ追加時に再びコピペされる構造。

**実装手順**

手順: (1) cmd/api/batch_scheduler.go を新設し、runScheduled(ctx context.Context, name string, nextFire func(time.Time) time.Time, task func(context.Context) error) を抽出（timer生成・select・エラーログ logger.Error(name+" failed", ...) を共通化。ログメッセージは既存3種の文字列を維持）。(2) nextFire 実装2種: 毎時0分 = now.Truncate(time.Hour).Add(time.Hour)、毎日02:00 = 既存 main.go:226-229 のロジックをそのまま移設。(3) JST は time.FixedZone をやめ time.Local を使用（main.go:37 の ConfigureTimeZone 成功後のみ到達するため等価。10/15/20時判定 main.go:210-211 も同様）。(4) main.go の3 goroutine を runScheduled 呼び出し3行に置換。トリガー時刻・エラー時継続の挙動は完全維持。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go vet ./cmd/api/... && docker compose exec backend go test ./internal/config/ -run TestConfigureTimeZone -count=1
```


## G10. cmd/・スクリプト・リポジトリ衛生

### G10-1. 13.9MB のコンパイル済み migrate バイナリが git 追跡下 + ビルド成果物の ignore 網羅漏れ

- **ID**: `tracked-migrate-binary`
- **重要度**: P2 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: migrate (binary (13,947,250 bytes)); .gitignore (1-3); .dockerignore (39-41)

**証拠(現HEAD検証済み)**

`git ls-files backend/migrate` → `backend/migrate`（追跡下）。`git cat-file -s HEAD:backend/migrate` → 13947250。`file backend/migrate` → "ELF 64-bit LSB executable, ARM aarch64 ... with debug_info, not stripped"。混入コミット: 0090d345 2026-07-06 "feat(backend): seed SQL→CSV migration & backend refactor Phase0 完了"。backend/.gitignore:1-3 は「# compiled binary (Air hot-reload output)\n/api\n/lstep-migrate」のみで /migrate が無く、ルート .gitignore も backend/seed-old-db (.gitignore:62)・backend/stage-import (.gitignore:63) はあるが backend/migrate が無い（`git check-ignore backend/migrate` exit=1）。さらに backend/.dockerignore:39-41 は「# Local build artifacts / helper files\napi\nFK_DEPENDENCY_CHECK_ROADMAP.md」のみで migrate/lstep-migrate(17MB)/seed-old-db(13MB)/stage-import(13.8MB) が全て未除外。ECS デプロイは backend/ をコンテキストにビルドする: .github/workflows/backend-deploy-ecs.yml:119-125 「docker buildx build --platform linux/amd64 -f backend/Dockerfile.production ... ./backend」+ Dockerfile.production:16 「COPY . .」。

**問題**

誰も参照しないローカルビルド成果物（ARM64 バイナリ。本番は amd64 をソースからビルド: Dockerfile.production:19-28）が全 clone/CI checkout に 13.9MB を恒久課金する。再ビルドの度に 13MB のバイナリ diff が git status に現れ、誤 commit を再誘発する。加えて .dockerignore 漏れにより毎デプロイのビルドコンテキスト転送と builder レイヤーに約 58MB の無関係な旧バイナリが混入する。同種バイナリ4本中2本ずつが別ファイルの ignore に分散している非対称性（backend/.gitignore vs ルート .gitignore）が今回のすり抜けの根因。

**実装手順**

手順: (1) `git rm --cached backend/migrate`（作業ツリーには残す）。(2) backend/.gitignore に `/migrate` `/seed-old-db` `/stage-import` を追記し、cmd/ 配下 7 ツールのバイナリ名を1ファイルに集約（ルート .gitignore:62-63 の backend/seed-old-db・backend/stage-import 行は重複になるので削除可）。(3) backend/.dockerignore の「Local build artifacts」節に `migrate` `lstep-migrate` `seed-old-db` `stage-import` を追記。(4) commit（type: chore）。注意: git 履歴からの完全除去（filter-repo）は force-push を要するため対象外 — 追跡解除のみで将来分を止める。影響範囲: ビルド/デプロイのコード経路は一切変わらない（Dockerfile.production はソースからビルド、worker exec はイメージ内 /app/migrate を使用）。

**検証コマンド(スコープ限定)**
```
git ls-files backend/migrate | wc -l  # 0 になること && git check-ignore backend/migrate backend/seed-old-db backend/stage-import && echo ignored-ok
```

### G10-2. 未参照の backend/Dockerfile が「本番用」を自称し実際の本番構成と乖離（デプロイ footgun）

- **ID**: `stale-root-dockerfile`
- **重要度**: P2 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: Dockerfile (1-23); README.md (49)

**証拠(現HEAD検証済み)**

backend/Dockerfile:1-10 「# 本番用 Dockerfile（マルチステージビルド）\nFROM golang:1.25-alpine AS builder\n...\nCOPY go.mod ./\nRUN go mod download && go mod tidy\n...\nRUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./cmd/api」。参照実態: compose 全ファイルは Dockerfile.dev（docker-compose.yml:28,57 / docker-compose.stage-import.yml:29 / docker-compose.seed-old-db.yml:16）、デプロイは Dockerfile.production（backend-deploy-ecs.yml:121「-f backend/Dockerfile.production」）で、backend/Dockerfile を参照する compose/CI は 0 件。乖離の実体: (a) go.sum を COPY せず `go mod tidy` をビルド内実行＝lock 無視の非再現ビルド（Dockerfile.production:12「COPY go.mod go.sum ./」と対照的）、(b) migrate バイナリと migrations/ を含まないため本番起動フロー（migrate→api）が成立しない（Dockerfile.production:25-28,52 と対照的）、(c) 非rootユーザー/HEALTHCHECK 無し。backend/README.md:49 「├── Dockerfile               # 本番用」がこの誤ラベルを文書側でも再生産している。

**問題**

`docker build backend/`（-f 省略時のデフォルト）はこの stale なファイルを拾う。「本番用」ラベルに従って触った作業者が、実際の本番イメージ（Dockerfile.production）と別物を検証・修正する事故経路になる。3種 Dockerfile のうち1種が完全な死に体で、乖離チェック（#195 で導入した drift check の思想）の対象面積を無駄に増やしている。

**実装手順**

手順: (1) `git rm backend/Dockerfile`。(2) backend/README.md:49 のディレクトリ構造記述を実態（Dockerfile.dev=開発 / Dockerfile.production=本番 ECS+Cloudflare Container）に更新。(3) リポジトリ全体を `grep -rn 'backend/Dockerfile[^.]'` で最終確認（現時点で compose/CI/docs に参照なしを確認済み。docs/tasks/closed/ 内の言及は frontend/Dockerfile のみ）。影響範囲: どのビルドパイプラインも参照していないため削除は挙動保存。代替案（残す場合）は Dockerfile.production と同内容へ整合させる必要があるが、二重管理になるだけなので削除を推奨。

**検証コマンド(スコープ限定)**
```
grep -rn 'backend/Dockerfile[^.]' --include='*.yml' --include='*.yaml' --include='*.md' --include='*.sh' . | grep -v node_modules ; test ! -f backend/Dockerfile && echo deleted-ok
```

### G10-3. performance-tests.yml のプロファイリングは API ではなく profile.go 自身のアイドルプロセスを計測しており成果物が無意味

- **ID**: `ci-self-profiling-meaningless`
- **重要度**: P2 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: scripts/profile.go (12-55); ../.github/workflows/performance-tests.yml (151-216)

**証拠(現HEAD検証済み)**

performance-tests.yml:151-152 で「Start Docker services: docker compose up -d db backend」と backend コンテナを起動した後、performance-tests.yml:174-178 「Run memory profiling: cd backend / go run scripts/profile.go memory」をランナーホスト側で実行する。scripts/profile.go:12-31 の実装は「func MemoryProfile(filename string) error { f, err := os.Create(filename) ... runtime.GC() ... pprof.WriteHeapProfile(f) }」であり、runtime/pprof は呼び出し元プロセス自身のヒープしか書けない。CPU も同様: scripts/profile.go:46-52 「pprof.StartCPUProfile(f) ... time.Sleep(duration)」＝起動直後の CLI が10秒 sleep する間の自プロセスを計測。つまり Docker 内で稼働する API とは無関係な、起動したての `go run` プロセスのプロファイルが取れるだけ。それを performance-tests.yml:201-207 が artifact（retention-days: 30）としてアップロードし、209-215 「Analyze memory profile: go tool pprof -top backend/profile_memory.pprof」で解析まで行う。

**問題**

「バックエンドのメモリ/CPU/ゴルーチンをプロファイルしている」ように見える CI ステップ一式が、実際には計測値ゼロの成果物を30日保存し続けている。パフォーマンス調査時にこの artifact を信じると誤診に直結する（監視の偽陽性ならぬ『計測の偽実在』）。backend を docker compose で起動するコストも無駄になっている。

**実装手順**

挙動保存の最小案: (1) performance-tests.yml の 4 ステップ（Run memory profiling / Run CPU profiling / Run goroutine profiling / Print memory stats: 174-199行）と Upload profiling results / Analyze memory profile（201-216行）を削除。(2) scripts/profile.go を削除（参照はこのワークフローのみであることを確認済み）。(3) 同 job で backend 起動が他ステップ（load-test 等）に不要なら Start Docker services も削除。代替案（真のプロファイルが要る場合・別トラック）: cmd/api に GIN_MODE=debug 限定で net/http/pprof を公開し `curl :8080/debug/pprof/heap` を取得する — これは API の公開面が増える behaviorChange=true の別提案として扱うこと。影響範囲: CI のみ。アプリ実行時挙動への影響なし。

**検証コマンド(スコープ限定)**
```
grep -rn 'profile.go' .github/workflows/ backend/ --include='*.yml' --include='*.go' | grep -v _test  # 参照残ゼロ確認。ワークフロー構文は actionlint .github/workflows/performance-tests.yml
```

### G10-4. 役目を終えた一回限りの codemod (fix_p8_wrap.py / fix_p11_slog.py) の残置 — 再実行はコンパイル破壊リスク

- **ID**: `oneoff-codemod-scripts-retire`
- **重要度**: P3 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: scripts/fix_p11_slog.py (1-97); scripts/fix_p8_wrap.py (1-209)

**証拠(現HEAD検証済み)**

scripts/fix_p11_slog.py:2-9 「P11 compliance fixer: insert slog.ErrorContext before every apperrors.Wrap in service files」、scripts/fix_p8_wrap.py:3 「P8 compliance fixer: replace bare `return X, err` with apperrors.Wrap in service files」。いずれも internal/service を正規表現で一括書き換えする使い捨てスクリプト（fix_p11_slog.py:16 「SERVICE_DIR = os.path.join(os.path.dirname(__file__), "../internal/service")」）。Makefile / .github/ / backend/docs からの参照は grep で 0 件（performance-tests.yml が参照するのは profile.go のみ）。fix_p11_slog.py:69 は「slog_line = f'{indent}slog.ErrorContext(ctx, "{msg}", "error", err)'」を ctx 変数の存在確認なしに挿入するため、ctx を持たない関数に対して再実行するとコンパイル不能なコードを生成する。P8/P11 の一括是正は完了済み（現行 service 層は規約準拠、golangci-lint ゲート稼働中）。

**問題**

完了済み一括修正の regex-codemod が実行可能な状態で残ると、将来の作業者（または自動エージェント）が「P11 fixer がある」と誤って再実行し、機械的挿入による破壊・レビューノイズを生む。git 履歴に完全な形で残るため、リポジトリに置き続ける保守上の利得はない。削除して困る運用シナリオ: 同種の一括是正を再度行う場合だが、その際は当時のコード状態に合わせて書き直しが必要であり、履歴からの復元（`git log --follow -- backend/scripts/fix_p11_slog.py`）で十分。

**実装手順**

手順: (1) `git rm backend/scripts/fix_p8_wrap.py backend/scripts/fix_p11_slog.py`。(2) commit message に「一括是正完了済み・履歴参照で復元可」と復元手順（対象コミットハッシュ）を明記。scripts/profile.go は finding ci-self-profiling-meaningless 側で処理するため、両方実施後 backend/scripts/ は空になる → ディレクトリごと削除。影響範囲: どこからも参照されていないため削除は挙動保存。

**検証コマンド(スコープ限定)**
```
grep -rn 'fix_p8\|fix_p11' . --include='*.yml' --include='*.md' --include='Makefile' | grep -v node_modules | grep -v docs/archive  # 参照残ゼロ確認
```

### G10-5. DB 接続 DSN 構築・env 読取・localHosts 安全ガードが cmd 4 ツールに重複しガード内容がドリフト済み

- **ID**: `cmd-dsn-localhosts-duplication`
- **重要度**: P3 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: cmd/migrate/main.go (45-74); cmd/seed-export/main.go (44-50, 119-148); cmd/stage-import/main.go (61-67, 213-275, 297-299); cmd/seed-old-db/main.go (30-35, 130-146, 181)

**証拠(現HEAD検証済み)**

同一 DSN フォーマット文字列が4箇所: cmd/migrate/main.go:71-74 「connStr := fmt.Sprintf(\n\t\t"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",\n\t\tdbHost, dbPort, dbUser, dbPassword, dbName, sslMode, config.JapanTimeZone,\n\t)」、cmd/seed-export/main.go:143-148 (connParams.dsn)、cmd/stage-import/main.go:217-222 (dsn.connString)、cmd/seed-old-db/main.go:181。env 読取＋デフォルト（port 5432 / sslmode disable）も各所で再実装（migrate:60-68, seed-export:123-141 readConnParams, stage-import:231-247 targetDSN+envOr, seed-old-db:130-143）。localHosts ガードは3実装でドリフト済み: cmd/seed-old-db/main.go:31-35 と cmd/seed-export/main.go:46-50 は「"db": true, "localhost": true, "127.0.0.1": true」の3要素、cmd/stage-import/main.go:61-67 のみ「"::1": true, "[::1]": true」を追加した5要素（コメント: "Includes IPv6 loopback forms in case localhost resolves to ::1"）。さらに SQL リテラルエスケープが cmd/migrate/csvbundle.go:158-160 quoteLiteral と cmd/stage-import/main.go:297-299 pqLiteral で同一実装。

**問題**

接続パラメータ仕様（TimeZone 付与・デフォルト値・必須変数）と安全ガード（非ローカル DB 拒否 = 本番誤爆防止の要）が4ツールにコピペ分散しており、stage-import だけ IPv6 対応が入った実ドリフトが既に発生している。localhost が ::1 に解決される環境では seed-old-db / seed-export のガードが false negative で拒否する（安全側だが、修正が3箇所必要になる構造が負債）。次に DSN 仕様が変わる時（例: sslrootcert 追加）に修正漏れが起きる。

**実装手順**

手順: (1) 新規パッケージ internal/dbconn を作成: `type ConnParams struct{ Host, Port, User, Password, SSLMode string }`、`FromEnv() (ConnParams, error)`（必須 = DB_HOST/DB_USER/DB_PASSWORD、デフォルト = port 5432 / sslmode disable）、`func (c ConnParams) DSN(dbname string) string`（TimeZone=config.JapanTimeZone 固定）、`func IsLocalHost(host string) bool`（stage-import の5要素 superset を正とする — ::1 は loopback の別表記でありガード意味論は等価）。(2) 4ツールを順に置換: seed-export（connParams/readConnParams/localHosts 削除）→ stage-import（dsn.connString の host/port/user/password/sslMode 部と localHosts/envOr を置換。connStringReadOnly の " default_transaction_read_only=on" 付加と DB_NAME 必須チェックはツール側に残す）→ seed-old-db → migrate（DB_NAME 必須チェックはツール側に残す）。(3) quoteLiteral/pqLiteral は2箇所・各3行のため統合は任意（統合するなら internal/dbconn ではなく各ツール残置を推奨 — cmd/migrate は本番バイナリ、stage-import は dev 専用で依存を増やさない）。(4) 各ツールの既存テスト（stage-import/dsn_test.go 等）を新パッケージに追随させ、dbconn 自体に FromEnv/DSN/IsLocalHost の table-driven テストを追加。影響範囲: cmd 4ツール + 新規1パッケージ。DSN 文字列出力はバイト単位で不変であること（既存 dsn_test.go の期待値変更ゼロ）を完了条件とする。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./cmd/stage-import/... ./cmd/seed-old-db/... ./cmd/migrate/... ./internal/dbconn/... -count=1
```

### G10-6. worker/migrate-exec.ts は「unit test 容易性のため分離」と明記されながらテストが 0 本（STG 稼働中の migrate 認証ゲート）

- **ID**: `worker-migrate-auth-untested`
- **重要度**: P3 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: worker/migrate-exec.ts (1-51); worker/index.ts (169-197)

**証拠(現HEAD検証済み)**

worker/migrate-exec.ts:1-4 「// P4-5(試行10): `/_internal/migrate` の認証・レスポンス整形ロジック。... ここでは Worker の fetch() から呼べる純粋関数のみを置く(unit test容易性のため分離)」— しかし backend/worker/ には index.ts と migrate-exec.ts の2ファイルのみでテストファイルが存在しない（`ls backend/worker/*.test.ts` → no matches）。この純粋関数群は本物の認証境界: migrate-exec.ts:33-40 「export function isAuthorizedMigrateRequest(request: Request, secret: string | undefined): boolean { if (!secret) { return false; } const authHeader = request.headers.get("Authorization") ?? ""; const expected = `Bearer ${secret}`; return timingSafeEqual(authHeader, expected); }」が worker/index.ts:179 で /_internal/migrate（リモート migration 実行）を守っており、この経路は既に STG 本線（.github/workflows/backend-deploy.yml:22 「WORKER_URL: https://animalekarte-stg-api.baritech-soga.workers.dev」）。ルート package.json にテストランナー未導入（scripts は harness:*/deploy:*/cf:* のみ、devDependencies は wrangler のみ）。

**問題**

「テストのために分離した」と自己宣言したモジュールが未テストのまま STG 稼働に入った。secret 未設定→fail-closed、長さ不一致→false、Bearer プレフィクス検証、exitCode→HTTP status 変換（0→200/非0→500）という仕様が全てコメント上の主張のみで、リグレッション検知手段がない。timingSafeEqual は Cloudflare 独自拡張 crypto.subtle.timingSafeEqual（migrate-exec.ts:13-15 のコメント参照）に依存しており、wrangler/workerd 更新での破壊も検知できない。

**実装手順**

手順: (1) ルート package.json devDependencies に vitest と @cloudflare/vitest-pool-workers（workerd ランタイムで crypto.subtle.timingSafeEqual 実挙動をテストするため。素の Node vitest では当該 API が存在せず偽の失敗/成功になる点に注意）を追加。(2) backend/worker/migrate-exec.test.ts を新規作成し AAA パターンで: isAuthorizedMigrateRequest — secret undefined→false / secret 空文字→false / ヘッダ欠落→false / 誤 secret→false / `Bearer <正 secret>`→true / Bearer プレフィクス無し→false、timingSafeEqual — 一致/不一致/長さ違い、toMigrateResponse — exitCode 0→status 200 / 1→500 と body JSON 構造。(3) ルート scripts に "test:worker": "vitest run backend/worker" を追加し、ci.yml の paths-filter（backend/worker/** 変更時）に組み込む。影響範囲: 新規テスト+CI 配線のみ、プロダクトコード変更なし。pnpm install が必要なため依存追加ステップはユーザー手動実行を報告すること。

**検証コマンド(スコープ限定)**
```
pnpm vitest run backend/worker  # 依存導入後。CI 配線は actionlint .github/workflows/ci.yml
```


## G11. テスト負債解消 (DB実行テスト追加)

### G11-1. 実効権限集約SQL FindAllEffectivePermissionsByStaffID がDB実行テストゼロ(全テストがmock)

- **ID**: `test-effective-permissions-sql-untested`
- **重要度**: P1 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/permission_group_repository.go (140-173); internal/repository/permission_group_staff_clinic_isolation_test.go (44-80)

**証拠(現HEAD検証済み)**

permission_group_repository.go:143-144「// clinicID パラメータで検索範囲を制限し、マルチクリニック昇格を防止（High-7）。\nfunc (r *permissionGroupRepository) FindAllEffectivePermissionsByStaffID(ctx context.Context, staffID, clinicID uint64) ([]model.PermissionGroupRule, error)」。実体はraw SQL(同:151-166)「SELECT pgr.resource, bool_or(pgr.can_view) AS can_view, ... JOIN permission_groups pg ON pg.id = spg.group_id AND pg.deleted_at IS NULL AND pg.is_active = true AND pg.clinic_id = ? ... GROUP BY pgr.resource」。テスト参照はservice層mock定義のみ4件(service/clinic_service_test.go:101, service/permission_group_service_test.go:54, service/staff_service_permissions_test.go:45, service/staff_service_test.go:230)。repository層テストはUpdateStaffGroupsのisolationのみ(permission_group_staff_clinic_isolation_test.go:44「func TestPermissionGroupRepository_UpdateStaffGroups_ClinicIsolation」)で、本メソッドをDBで実行するテストは0件。呼び出し元はhandler/auth_me_response.go:133とhandler/clinic_handler.go:66(GetEffectivePermissions経由)で認可情報の正本。

**問題**

認可の実効権限を計算するSQL(bool_or合成・is_active/deleted_at除外・pg.clinic_id述語によるHigh-7マルチクリニック昇格防止)が一度もDBで実行されずCIカバレッジ0。SQLの述語1つの退行(例: is_active条件の脱落、clinic_id述語の脱落)を検出するテストが存在せず、セキュリティ上最も回帰コストが高い箇所がmock-onlyになっている。

**実装手順**

新規ファイル internal/repository/permission_group_effective_permissions_test.go を追加(setupTestDB + isolation_test_helpers_test.go の既存ヘルパーを流用)。検証テストを1件ずつ: (1)複数グループ所属時のbool_or合成: グループAがcan_view=false/グループBがcan_view=true → 合成結果true、can_delete両方false → false。(2)is_active=falseグループのルールが集約から除外される。(3)deleted_at IS NOT NULL(ソフトデリート済み)グループのルールが除外される。(4)クリニック隔離(High-7): staffがクリニックA/B両方のグループに所属する状態で clinicID=A 指定時にBのルールが混入しない。(5)所属グループなしstaff → 空スライス。(6)resource毎のGROUP BY: 2グループが同一resourceと別resourceを持つ場合の行数と値。既存のTestPermissionGroupRepository_UpdateStaffGroups_ClinicIsolationのfixture構築パターン(同ファイル44-80行)を踏襲する。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run TestPermissionGroupRepository_FindAllEffectivePermissions -count=1
```

### G11-2. 会計完了→予約完了化 CompleteAccountingAppointments(JST日付境界+サブクエリUPDATE×2)がDB実行テストゼロ

- **ID**: `test-complete-accounting-appointments-sql-untested`
- **重要度**: P2 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/accounting_repository.go (312-349); internal/service/accounting_service_test.go (579-609)

**証拠(現HEAD検証済み)**

accounting_repository.go:312-321「func (r *accountingRepository) CompleteAccountingAppointments(...) ... Where("DATE(start_time AT TIME ZONE 'Asia/Tokyo') = DATE(? AT TIME ZONE 'Asia/Tokyo')", scheduledDate).Update("status", model.ReservationStatusCompleted)」および同:332-341のmedical_record_id経由サブクエリUPDATE「Where("status NOT IN ?", []model.ReservationStatus{...Completed, ...Cancelled, ...NoShow}).Where("id IN (?)", r.db.Model(&model.MedicalRecord{}).Select("appointment_id").Where("id = ? AND clinic_id = ? AND appointment_id IS NOT NULL AND deleted_at IS NULL", *medicalRecordID, clinicID))」。テストはservice層のmock呼び出し検証のみ: accounting_service_test.go:584-607「called := false ... completeApptsFn: func(...) { called = true ... } ... assert.True(t, called, "CompleteAccountingAppointments が呼ばれること")」。repository層の*_test.goに本メソッドの呼び出しは0件(grep確認)。

**問題**

会計完了時に受付ボードのorphanカード残留を防ぐ#77対策の中核SQL。JSTタイムゾーン日付境界比較・サブクエリIN・status遷移except条件という退行しやすい要素が3つ揃っているのに、SQLセマンティクスを検証するテストが無くmockの呼び出し有無しか確認していない。UTC/JST境界(前日15:00 UTC=JST 0:00)の判定ミスやサブクエリのclinic_id述語脱落が既存テストでは検出不能。

**実装手順**

新規ファイル internal/repository/accounting_complete_appointments_test.go を追加(setupTestDB使用)。検証テストを1件ずつ: (1)経路(1): 同日同ペットstatus=accountingの予約がcompletedになる。(2)経路(1)除外: 別日/別ペット/別クリニック/status≠accounting/deleted_atありの予約は変更されない。(3)JST日付境界: start_timeがUTC 前日15:00(=JST当日0:00)の予約が同日扱いで完了化、UTC当日14:59(=JST当日23:59)も同日、UTC当日15:00(=JST翌日0:00)は対象外。(4)経路(2): medicalRecordID経由でstatus=reserved等の非完了予約が完了化される。(5)経路(2)除外: completed/cancelled/no_showは触らない、別クリニックの同ID medical_recordのappointment_idは更新されない(クリニック隔離)。(6)戻り値totalAffectedが両経路の合算になる。(7)ownerID/petID/medicalRecordIDがnilの縮退経路。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run TestAccountingRepository_CompleteAccountingAppointments -count=1
```

### G11-3. トリミング未請求明細取得 FindUnbilledTrimmingItemsByPetID(UNION ALL+NOT EXISTS 70行raw SQL)がDB実行テストゼロ

- **ID**: `test-unbilled-trimming-raw-sql-untested`
- **重要度**: P2 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/billing_item_repository.go (183-301)

**証拠(現HEAD検証済み)**

billing_item_repository.go:183「func (r *billingItemRepository) FindUnbilledTrimmingItemsByPetID(ctx context.Context, clinicID, petID uint64) ([]model.BillingItem, error)」— 実体は同:194-253のraw SQL。コース側とオプション側のUNION ALL、各枝に請求済み除外のNOT EXISTS(同:212-220)「AND NOT EXISTS (SELECT 1 FROM billing_items bi JOIN billings b ON b.id = bi.billing_id AND b.deleted_at IS NULL WHERE bi.appointment_id = a.id AND bi.trimming_course_id = tc.id AND bi.deleted_at IS NULL AND b.status != ?)」、価格0除外「AND COALESCE(tc.price, 0) > 0」(同:211)、ORDER BY sort_order合成(同:249)。隣接するCountNonAccountingTrimmingByPetAndDate(同:283-296)もJST日付境界「Where("DATE(appointments.start_time AT TIME ZONE 'Asia/Tokyo') = DATE(? AT TIME ZONE 'Asia/Tokyo')", date)」を含む。両メソッドともrepository *_test.goでの呼び出し0件(grep確認)、service層参照はbilling_item_service.go:361とmock経由テストのみ。

**問題**

会計×トリミングの金額起点となるクエリ(会計画面への未請求コース/オプション自動取り込み)が一度もDBで実行されない。NOT EXISTSの結合条件(bi.trimming_course_id vs bi.trimming_option_id の枝別対応)、cancelled請求のみ存在する場合の再取得、UNION後の並び順という間違えやすい仕様がテストで固定されていない。

**実装手順**

新規ファイル internal/repository/billing_item_trimming_test.go を追加(setupTestDB使用)。検証テストを1件ずつ: (1)status=accountingのトリミング予約でコース+オプション2件が結合結果に出る(件数・name・unit_price・TrimmingCourseID/TrimmingOptionIDの枝別セット)。(2)並び順: コース(sort_order=0)→オプション(100+ato.sort_order)昇順。(3)請求済み除外: 有効なbillingに同一appointment_id+course_idのbilling_itemが存在すると除外される。(4)cancelled請求のみの場合は再取得対象になる(b.status != cancelled条件)。(5)price=0/NULLのコース・オプションは除外。(6)クリニック/ペット/status≠accounting/カテゴリ≠trimmingの除外。(7)CountNonAccountingTrimmingByPetAndDate: JST日付境界(UTC前日15:00)と対象status(accounting/completed/cancelled以外)の判定。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestBillingItemRepository_(FindUnbilledTrimmingItems|CountNonAccountingTrimming)' -count=1
```

### G11-4. UpdateReservationCapabilitiesの越境ガードにisolation test無し(兄弟のUpdateExcludedReservationTypesには有り)

- **ID**: `test-capability-write-isolation-asymmetry`
- **重要度**: P2 / **工数目安**: S / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/reservation_staff_repository.go (279-318); internal/repository/reservation_staff_exclusion_clinic_isolation_test.go (40-80)

**証拠(現HEAD検証済み)**

reservation_staff_repository.go:283-293にガード実装あり「if len(typeIDs) > 0 { var count int64; if err := r.db.WithContext(ctx).Model(&model.ReservationType{}).Where("clinic_id = ? AND id IN ? AND deleted_at IS NULL", clinicID, typeIDs).Count(&count)... if count != int64(len(typeIDs)) { return apperrors.WrapInvalidInput("reservation_type_ids contains invalid reservation type") } }」。しかしテスト参照はservice層mock5件のみ(liff_service_availability_staff_test.go:67等)でrepository実行テスト0件。一方、同一パターンの兄弟メソッドには専用isolation testが存在: reservation_staff_exclusion_clinic_isolation_test.go:40「func TestReservationStaffRepository_UpdateExcludedReservationTypes_ClinicIsolation」(別クリニックtypeB.ID拒否/自クリニック許可/混在拒否の3ケース)。repository/CLAUDE.mdの「正本ガード = 各サイトの runtime isolation test」方針に対し兄弟間で非対称。

**問題**

ガード自体は実装済みだが、その有効性を動作で証明するテストが無い。将来のリファクタでcount検証が脱落してもCIは緑のまま。exclusion側に既に確立したテストパターンがあるため追加コストは小さい。

**実装手順**

新規ファイル internal/repository/reservation_staff_capability_write_clinic_isolation_test.go を追加。reservation_staff_exclusion_clinic_isolation_test.go:40-80の構造をそのまま踏襲し UpdateReservationCapabilities で3ケース: (1)clinicA権限でclinicBのtypeB.IDを指定→WrapInvalidInputエラーかつstaff_reservation_capabilities未挿入、(2)自クリニックtypeA.ID→成功、(3)混在[typeA.ID, typeB.ID]→エラーかつ部分挿入なし(DELETE前検証の確認)。同テスト内にSupportsReservationType(同:320-330)の正/負1ケースずつを同乗させ、LIFF予約ゲートの読み取りも実DBで固定する。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run TestReservationStaffRepository_UpdateReservationCapabilities_ClinicIsolation -count=1
```

### G11-5. Lstep配信ターゲティング用購買クエリ3本(HasItemByOwnerSince等)がDB実行テストゼロ+docコメントとSQLの乖離

- **ID**: `test-lstep-purchase-queries-untested`
- **重要度**: P3 / **工数目安**: M / **挙動変更**: なし（挙動保存）
- **対象ファイル**: internal/repository/billing_item_repository.go (123-181)

**証拠(現HEAD検証済み)**

billing_item_repository.go:139-141「// FindOwnersByCategoryPurchaseDate は指定カテゴリの最終購入日が purchaseDate と一致する飼い主IDリストを返す（FEAT-383）。\n// billings.issued_at::date の MAX が purchaseDate と一致する飼い主を返す。」に対しSQL実体(同:154)は「HAVING MAX(billings.completed_at::date) = ?::date」でissued_atではなくcompleted_atを使用(コメントとSQLが乖離)。HasItemByOwnerSince(同:123-137)/HasFoodPurchaseByOwnerSince(同:166-181)もJOIN billings+completed_at>=sinceのクエリだが、3本ともrepository *_test.goでの呼び出し0件(grep確認)、テスト参照はservice層mockのみ。

**問題**

FEAT-383系のフード切れ・再購入リマインド配信のターゲット抽出条件がDBで検証されておらず、しかもdocコメント(issued_at)とSQL(completed_at)が食い違ったまま仕様が宙に浮いている。テストで正しい基準日カラムを固定しない限り、どちらが仕様か将来の読者が判別できない。

**実装手順**

まずFEAT-383の経緯からcompleted_at基準が正であることを確認し、billing_item_repository.go:140のコメントをcompleted_atに修正(コメントのみ・挙動不変)。次にinternal/repository/billing_item_lstep_queries_test.goを追加し1件ずつ: (1)HasItemByOwnerSince: names一致+completed_at>=sinceでtrue、since前のみ/名前不一致/別owner/別クリニック/soft-deleted billingでfalse、names空でクエリ発行なしのfalse。(2)FindOwnersByCategoryPurchaseDate: MAX(completed_at::date)が指定日に一致するownerのみ返る(同ownerがより新しい購買を持つと除外される=HAVING MAXの本質)。(3)HasFoodPurchaseByOwnerSince: names指定時はname IN、未指定時はcategory=foodへフォールバックする分岐。

**検証コマンド(スコープ限定)**
```
docker compose exec backend go test ./internal/repository/ -run 'TestBillingItemRepository_(HasItemByOwnerSince|FindOwnersByCategoryPurchaseDate|HasFoodPurchaseByOwnerSince)' -count=1
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

---

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
- worker/index.ts: 薄いプロキシ + /_internal/migrate の設計コメント充実（DB_RESET 非注入・exec への最小権限 env subset・XFF 偽装対策・タイムアウト付き exec）。構造は clean（残課題はテストのみ = finding worker-migrate-auth-untested）
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