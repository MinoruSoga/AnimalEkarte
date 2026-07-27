# BE-refactor — backend 規約適合監査と改善計画

> 更新: 2026-07-27（初版・全域監査 1 回目）
>
> **本書の位置づけ**: `backend/` 全域の規約適合監査で検出した改善項目の正本。`FE-refactor.md` の backend 版であり、readiness view（`3-session-agent.html`）とは役割が異なる。
>
> **本書に書かないもの**:
> - **規約テキストそのもの**（2026-07-26 `ddc609a79` で旧 BE-refactor.md の 4 規定を `.claude/refs/go-gin-backend-review.md` §6 へ移設して退役済み。ルールを本書へ書き戻すと移設が無言で巻き戻る）。本書は規約を **引用するだけ** で、正本は常に規約側にある。
> - **既に起票済みの課題の再記述**（`3-session-agent.html#ledger` / `BE-pending.md` / `phase2.html` が正本。本書は 1 行のポインタのみ持つ）。
>
> **writer 規則**: 本書の writer は監査を実行するセッション単独とする。レーン A/B の実行セッションは本書を読むのみで書かない（4 セッション並行のため）。
>
> **着手判断**: 本書の項目は台帳タスクではない。着手すると決めたものだけを `3-session-agent.html#ledger` へ `<section class="task">` として起票し、本書の該当項目に起票先 ID を追記する。

---

## 監査方法とカバレッジ

- **実施**: 2026-07-27・14 ドメイン単位に分割した並列監査 → 単位ごとの敵対的検証 → 重複排除（31 エージェント・read-only）。
- **判定**: 全所見に規約正本の `file:line` と逐語引用を必須とし、検証フェーズで「引用された規約がその行に実在し、その状況に適用されるか」「evidence の file:line が記述どおりか」「既に対処済みでないか」を再実測した。**18 件が却下**され、残った **105 件** を本書に載せる（うち横断集約で重複を畳んだ結果が下記）。
- **未実施の検証**: read-only 制約により `go test` / `golangci-lint` / `gofmt` / DB 接続を一切実行していない。したがって並行性シナリオ（TOCTOU・ロック挙動）は **コード構造と PostgreSQL のセマンティクスからの導出であり、実測トレースではない**。性能系（N+1・クエリプラン）は実測が取れないため原則として発行していない。
- **読了率**: 12/14 単位は担当ファイルを **全文読了**（sample・skip ゼロ）。例外は lstep 2 単位で、`lstep [a-l]` は 124 ファイル中 79 ファイルを全文読了し、残 45 ファイルは 10 種のアンチパターン grep を全数走査で代替した。`lstep [m-z]` も同様の併用。**lstep の未読領域に残存所見がある可能性を明示しておく**。
- **並行 WIP の扱い**: 監査中にレーンが `internal/lintscan` の inventory 登録等を清算した（`afd8404a4` 他）。検証時点で clean になったファイルは「新規（着手可能）」へ格上げ済み。`pet/repository.go` 系のみ WIP のまま残る。

---

## サマリー

| 区分 | 件数 |
|---|---|
| 検証通過（本書掲載） | 105 |
| うち横断パターンへ集約 | 10 パターン（延べ 71 件） |
| ドメイン固有として残置 | 27 |
| 既知起票済み（ポインタのみ） | 6 |
| 規約側の疑義 | 41（重複排除後 30 論点） |
| 検証で却下 | 18 |

**最も重いもの 3 点**:

1. **[X-A] 締め後の会計編集ゲートが `billing-items` 経路で完全に迂回される**（CRITICAL）。`POST/PATCH /billing-items` は billing の status もレジ締め状態も検査せず、監査も残さずに billing 総額を書き換える。
2. **[X-01] commit 済みの成功を tx 外の再取得エラーで 5xx へ反転させる実装が 7 ドメイン 15 箇所以上に散在**（HIGH）。臨床側には「失敗」と表示されるが DB は更新済みで、再送すると 409 か二重適用になる。
3. **[X-04] error の握り潰しによる fail-open が lstep バッチ群に集中**（HIGH）。DB エラーで `return 0, nil` を返すため、クリニック全体のタグ同期が全滅しても `BatchRunResult.Failed` は増えず `audit_logs` にも残らない。

---

## 横断パターン（複数ドメインに同型で存在）

同じ規約違反が複数パッケージへ散っているものを 1 項目へ畳んだ。**個別修正ではなく、パターン単位で先例へ寄せるのが正しい対応**である。

### [X-01] commit 済み write を tx 外の再取得 error で失敗応答へ反転させる — HIGH / 着手: 納品後

- 規約: `backend/CODING_RULES.md:78` 「write後の再取得が失敗し得る場合はcommit前の同じtransaction内で行うか、commit済みの成功を後段read errorで失敗へ反転させないcontractにする。」
- 実測（15 箇所）: `medicalrecord/care_plan_item_service.go:184,194,234,244` ／ `prescription_service.go:153` ／ `treatment_service.go:472` ／ `treatment_plan_service.go:141,164` ／ `clinic/clinic_service.go:403,408` ／ `clinic/company_service.go:112,116` ／ `reservation/line_reservation_setting_service.go:198,202` ／ `reservation/reservation_type_liff_service.go:156,163` ／ `manualarticle/repository.go:114` ／ `lstep/lstep_lifecycle_service.go:137`
- 内容: write を commit した後、応答生成のために別接続で再取得し、その失敗を error として返す。DB の瞬断やプール枯渇で「更新は成功したが 5xx」が発生する。`lstep_lifecycle_service.go:140` はコメントで「死亡記録の巻き戻しは行わない」と明記しており、規約が禁じた状態そのものを自認している。
- 対応: 再取得を `WithTx` クロージャ内へ移すか、write の戻り値から応答を構築する。**正しい先例が同一リポジトリに 4 つある**: `vital_service.go:294-298` ／ `owner/repository.go:288` ／ `pet/repository.go:417` ／ `trimming_service.go:330`。

### [X-02] 入力境界検証の欠落（列挙値・範囲・長さ） — HIGH / 着手: 即時（低リスク・独立）

- 規約: `.claude/rules/go-gin-backend-guidelines.md:151` 「外部入力は境界で型・形式・長さ・範囲・列挙値を検証する。」
- 実測（列挙値）: `staff/staff_request.go:17,70`（staff_type）／`lstep/lstep_trigger_priority_request.go:5`（trigger_type）／`lstep/lstep_tag_config_request.go:5`（category）／`lstep/shared_file_request.go:15`（purpose）／`trimming/trimming_request.go:72,128`（bw_unit）
- 実測（範囲）: `medicalrecord/consultation_request.go:13,42`（tax_rate に上限なし・同 struct の peer は `min=0,max=1`）／`medicalrecord/treatment_plan_request.go:7-11`（単価・数量・割引・**小計をクライアント値のまま採用**）／`reservation/line_reservation_setting_request.go`（**28 フィールド全てに binding tag が 0 件**。負の `booking_window_max_days` が `available_dates.go:115` の `make(cap<0)` へ到達し当該 clinic の応答を確実に 500 へ落とす）／`medicalrecord/lab_import_request.go:38,61`（**異常判定の基準値をクライアントが自由指定できる**）
- 実測（長さ）: `pet/pet_request.go:106-120` ／ owner 各 request ／ `clinic/clinic_request.go:16-22` ／ `trimming/trimming_request.go:70-78` ／ `manualarticle/request.go:10`
- 対応: enum は request DTO へ `oneof=` を付け、列挙値は model 定数から導出する（定数追加時に binding tag が追随することを test で固定）。範囲・長さは `001_init.sql` の列定義から上限を導出する。**金額・税率の範囲 contract が規約正本のどこにも無い**ため、値は下記「規約側の疑義」と併せて確定させる。

### [X-03] destructive / irreversible 操作の監査・recovery 欠落 — HIGH / 着手: 納品後

- 規約: `.claude/refs/backend-application-invariants.md:31` 「destructive または irreversible な操作には、権限、対象 scope、監査、recovery 方針を持たせる。」
- 実測: `staff/staff_clinic_assignment_service.go:213-281`（**スタッフの所属医院＝テナント境界を全置換するのに監査ゼロ**。同 service の `SetPermissionGroupIDs` は監査を持つ）／`lstep/lstep_lifecycle_service.go:239-286`（opt-out / opt-in / 削除がタグキャッシュを全削除するのに監査ゼロ・actor も取得しない）／`medicalrecord/care_plan_item_repository.go:95-97`（`.Unscoped().Delete()` で**物理削除**・`deleted_at` 列なし）／`medicalrecord/hospitalization_service.go:385-429`（同 service は `DischargeWithBilling` でのみ監査を使う）／`pet/animal_species_*`（全院共有マスタの**物理削除**・監査なし）／`manualarticle/repository.go:117-120`（物理削除＋`001_init.sql:2873` の CASCADE で全編集履歴が消滅）／`csvimport/import.go:80`（**37 表への述語なし DELETE**・対象 DB の同定手段がなく doc comment の期待だけが根拠）
- 対応: 既存の fail-closed 監査（`audit.LogEntryTx`）へ寄せる。`auth/permission_group_service.go` 等 5 領域に確立済みパターンがあり、新テーブルは不要。物理削除は soft delete 化か、削除前スナップショットの監査保存を必須にする。

### [X-04] error 握り潰しによる fail-open — HIGH / 着手: 納品後（lstep 分は要優先）

- 規約: `.claude/refs/error-handling.md:9` 「error を無視しない。処理できる境界まで返すか、明示的に回復する。」
- 実測（重い順）: `lstep/lstep_batch_segmentation.go:22-25`（DB エラーで `return 0, nil`・コメントも「静かにスキップする」と明記。**クリニック全体の同期が全滅しても `BatchRunResult.Failed` に計上されず `audit_logs` にも残らない**）／`lstep/lstep_settings_credentials.go:36-47`（**credential 復号失敗を空文字へ置換**。同期無効と区別できず全 LINE 配信が静かに停止し、設定画面は「未設定」と表示して管理者に再入力させる）／`lstep/lstep_tag_sync_api.go:15-19`（キャッシュ読取失敗を「API 失敗なし」として返し、旧タグを解除しないまま新タグを付与）／`lstep/lstep_tag_service.go:158-162`（200 + `tags:[]` を返し「タグ 0 件」と誤表示）／`medicalrecord/medical_record_subrecords.go`（**主訴・治療方針・診断の消失が slog のみで、API は 201 を返す**）／`testdb/testdb.go:141-145`（テスト間分離の TRUNCATE が error 破棄・10 箇所）／`audit/repository.go:58`（監査値の marshal 失敗を無記録で握り潰す）／`cmd/coverage-ratchet/main.go:91`＋`ci.yml:201`（**gate が tee で exit code を吸収され構造的に機能していない**）／`cmd/lstep-migrate/migrator.go:292` ／ `reservation/liff_service_reservations.go:26-29`（log すら残さない）
- 対応: 「握り潰す」と「意図的 best-effort」を分離する。後者を選ぶ場合は規約（`CODING_RULES.md:36`）が要求する補償・再試行・監査・部分失敗 contract を実際に持たせる。**この線引きが規約側で曖昧である**（下記の疑義参照）。

### [X-05] 検証と write が同一 transaction にない（TOCTOU） — HIGH / 着手: 納品後

- 規約: `backend/CODING_RULES.md:38` 「request由来のclinic-scoped FKは永続化と同じtransactionで再検証し、並行master変更で判定が無効になる場合は対象行をcommitまで共有ロックする。」
- 実測: `pet/service.go:353,374`（owner_id / insurance_id を tx 外で検証。**WIP-adjacent**）／`medicalrecord/exam_type_service.go:118,131,148,158`（repository は ambient tx がある時のみ FOR SHARE を掛ける実装を既に持つのに、Create/Update が tx を開かないため発火しない）／`medicalrecord/medicine_dose_param_service.go:100-121`（#201 投与量パラメータの per_weight ガードが親 medicine を固定しない）／`medicalrecord/procedure_service.go:213-232` ＋ `medicine_service.go:471-495`（削除の「使用中 0 件」ガードが 4 statement に分かれロックなし）／`clinic/closing_settings_service.go:183-203`（重複禁止チェックと Create が別 tx・repository が `DBOrTx` を経由しないため呼び出し側が tx で包んでも参加できない）／`reservation/reservation_type_liff_service.go:203-219`（存在チェック通過後に確定した予約を、参照先ごと soft delete できる）
- 対応: 検証と write を単一 `WithTx` へ収め、判定根拠となる行を `FOR UPDATE` / `FOR SHARE` で固定する。前提として当該 repository を `persistence.DBOrTx` ベースへ揃える。正しい先例 = `reservation/reservation_intent_repository.go:591-625`。

### [X-06] 1 つの business graph を構成する write の非原子化 — HIGH / 着手: 納品後

- 規約: `backend/CLAUDE.md:33` 「1つのbusiness graphを構成する複数rowのwriteは同じtransactionで原子的に扱い、commit済みの成功を後段の再取得errorで失敗応答へ反転させない。」
- 実測: `billing/campaign_service.go:287,293`（**campaignService が Transactor を保持していない**。本体更新と対象差し替えが別 commit になり、部分成功で割引マッチング対象が不整合になる）／`lstep/lstep_settings_service.go:278-287`（`clinic_integrations` 6 行 + `lstep_settings` + `clinic_settings` を非トランザクションで逐次 Upsert。途中失敗で LINE access token と channel secret の組が不整合のまま Webhook 署名検証が走る）／`medicalrecord/lab_import_examination_service.go:219,242`（exams と exam_results が非原子。補償削除の失敗も握り潰され、結果なしの孤児 exam が残る）／`reservation/reservation_service.go:499-550`（キャンセルが「status 更新」と「soft delete」の 2 write に分かれ、後段失敗で cancelled のまま予約管理に残る）
- 対応: `Transactor` を注入し、repository を `persistence.DBOrTx` 経由へ揃える。

### [X-07] request body サイズ上限の非対称と middleware による無効化 — HIGH / 着手: 即時

- 規約: `.claude/rules/go-gin-backend-guidelines.md:179` 「rate limit、request/body/upload size、content type、file path を制限する。」
- 実測: **`middleware/sanitize_null_bytes.go:64,65` が `http.MaxBytesReader` を無効化する** — 除去後のバイト数のみを返し `ContentLength = -1` を設定するため、除去されたバイトが上限にカウントされない。制御バイトのみの巨大 body が無制限に読まれる（影響は JSON binder 3 経路: `auth/http_binding.go:26,30` ／ `staff/http_binding.go:26,30` ／ `billing/billing_confirmation_handler.go:131,135`）。加えて `pet` / `owner` / `clinic` / `trimming` / `manualarticle` の全 `ShouldBindJSON` に上限がなく、グローバル middleware（`cmd/api/main.go:198-203`）にも body size 制限が無い。
- 対応: `protected` グループ全体へ body size middleware を 1 本入れて全パッケージで統一する（handler 個別対応は非対称を再生産する）。sanitize と MaxBytesReader の適用順序も同時に決める。

### [X-08] 外部 API の raw error 文字列を応答・DB へ露出 — MEDIUM / 着手: 納品後

- 規約: `backend/CLAUDE.md:34` 「error response と log に secret、credential、個人情報、内部詳細を出さない。」
- 実測: `lstep/lstep_settings_connection.go:44,55`（`*url.Error` が要求 URL と `dial tcp <ip>:<port>` を含んだまま `lstep_error` / `line_error` へ載る。管理者権限ゲート済みだが内部ネットワークの到達性オラクルになる）／`lstep/line_send_service.go:169-195`（**502 応答と `LineSendLog.ErrorMessage` の両方へ生エラーを保存し、送信履歴 API から再露出する**）
- 対応: 応答へは安定した分類コード（`unauthorized` / `unreachable` / `timeout`）のみを載せ、生エラーは slog に限定する。

### [X-09] copy-paste drift（同一ロジックの複製と乖離） — MEDIUM / 着手: 納品後

- 規約: `~/.claude/rules/ecc/common/coding-style.md:26` 「Avoid copy-paste implementation drift」
- 実測: **`model/estimate.go:83` が課税ベースから `DiscountAmount` を引かず、`model/accounting.go:152`（#85 準拠）と計算式が乖離している**（見積と会計で税額が食い違う）／`owner/validators.go:108-136` と `pet/validators.go:14-`（ペット列挙値バリデータ 4 関数が完全複製。現状の受理挙動は等価だが構造が既に分岐しており、片側更新で経路依存の受理差が出る）／`medicalrecord/medical_record_subrecords.go:98-121`（診断 FK の**所有権検証というセキュリティ判定**を `clinicalPlanService` から複製したとコメントで自認。実際に `validateDiagnosisTypeNameConsistency` の追随漏れが発生済み）／`clinic/closing_settings_service.go:383`（`parseHHMM` が billing 側の複製と自認するが削除条件がない）／`owner/validators_contact.go` の email・電話・郵便番号検証が clinic / company では未使用
- 対応: `sharedkernel` へ 1 本化する（既に `ValidateRequiredName` 等の実績あり）。**税計算の乖離だけは影響が金額に直結するため先行**。

### [X-10] 同一 error の重複ログ / 二重レスポンス — LOW / 着手: 即時

- 規約: `backend/CODING_RULES.md:67` 「同じ error を複数箇所で重複ログしない。」／ `.claude/refs/error-handling.md:29` 「response を書いた後に別の error response を重ねない。」
- 実測: `clinic/closing_settings_service.go:271,275`（同一 err を連続 2 レコード出力）／`trimming/trimming_service.go:554,577`（tx 内と呼び出し元で二重 ERROR）／`auth/http_permission.go:20,28,32`（`Extract*` が既にレスポンスを書いた後、`RequirePermission` がもう一度書く経路がある）

---

## ドメイン別（横断に畳めない固有の所見）

### 会計・在庫（billing / inventory）

- **[X-A] `billing-items` の POST/PATCH が締め後会計編集ゲートを完全に迂回する — CRITICAL / 着手: 即時**
  - 規約: `.claude/refs/backend-application-invariants.md:37`（fail-closed 監査の tx 参加）
  - 実測: `billing/routes.go:111,112` は `accounting:edit` / `accounting:create` のみを要求し、billing の status もレジ締め状態も検査しない。`billing_item_service.go:404-454`（UpdateItem）には status 検査が一切なく、`:282-402`（CreateItem）の完了/取消拒否は `input.VaccinationID != nil` の枝でしか到達しない（`billing_item_repository.go:280-281`）。**非ワクチン経路は確定済み会計の明細を書き換えられ、監査も残らない**（`DeleteItem` は防御済みで非対称）。
  - 対応: `CreateItem` / `UpdateItem` を `DeleteItem` と同型にし、tx 内で `LockAndFindByID` により completed/cancelled を 409 拒否。締め済み期間の変更は accounting PATCH と同じ post-close 権限・理由・fail-closed 監査を要求する。
- **[BIL-02] fail-closed 宣言と実装の乖離 — MEDIUM / 着手: 即時**
  - `accounting_service.go:146-149` の doc コメントは「3 経路とも fail-closed 化済み」と宣言するが、`accounting_service_correction.go:178-180` は `if s.auditTx == nil { return nil }` で**監査ゼロのまま確定済み会計のカード金額訂正を commit する**。`Cancel` も同型。宣言と実装のどちらが正かを確定し、片方を直す。

### 横断基盤（httpapi / middleware / apperrors / audit）

- **[INF-01] 未分類の PostgreSQL エラーが全て HTTP 400 に落ちる — HIGH / 着手: 即時**
  - `httpapi/response.go:87-89` は `case isPgError(err)` を default(500) の直前に置き、`pgconn.PgError` がチェーンにあれば無条件で 400 を返す。`classifyPgError` が明示的に扱うのは 5 コードのみで、それ以外（デッドロック・接続断・権限エラー等）も「入力値が正しくありません」になる。**サーバ障害が 5xx として計上されずサイレント障害化する**。
  - 対応: `classifyPgError` を「クライアント起因と確定できるコードの allowlist」に限定し、非該当は default(500) へ落とす。
- **[INF-04] `apperrors.FromGORM` が error message の部分文字列一致で分類している — MEDIUM**
  - `apperrors/errors.go:174-178` が `strings.Contains(errMsg, "unable to encode")` 等 3 本で分類。`.claude/refs/error-handling.md:18` が明示的に禁じる形だが、**pgx が型付き sentinel を公開していないため現状は回避不能**。規約側に例外規定を置くか upstream を再確認する（下記疑義参照）。

### LINE / Lステップ連携（lstep）

- **[LSA-01] `lstep_base_url` が無検証で保存され、復号済み API キーが任意ホストへ Bearer 送信される — HIGH / 着手: 即時**
  - `lstep_settings_request.go:6` に binding tag がなく、scheme/host の検証も allowlist もないまま `ClinicIntegration` へ Upsert される。`testLstepAPI` はその値へ `Authorization: Bearer <復号済みキー>` を送る。**GET 応答では同キーを `crypto.MaskValue` でマスクする設計なのに、この経路で平文が外部へ出る**（保管 secret の持ち出し＋SSRF）。
  - 対応: `url.Parse` で https 固定＋許可ホスト allowlist を境界で強制する。
- **[LSA-02 / LSB-01] 自動配信の除外判定が `owner.LstepOptOut` を読まない — HIGH / 着手: 即時**
  - `lstep_delivery_trigger_state.go:11-23` の `checkExclusion` は `DeliveryExcluded` / `LineUserID` / `EXCL_配信停止` タグの 3 点しか見ない。一方 opt-out API が呼ぶ `RecordLstepOptOut` は `lstep_opt_out` 系 3 列のみ更新し `delivery_excluded` を立てない。しかも `HandleOwnerOptOut` は `EXCL_配信停止` のキャッシュ行ごと削除する。**オプトアウト済みの飼主へ配信が発火する**。同一判定を line_send / checkup_sync / tag_sync の 3 経路は `LstepOptOut` 直読で行っており、本経路だけが逸脱。
  - 対応: `checkExclusion` 先頭に `if owner.LstepOptOut { return true, "lstep_opt_out", nil }` を追加し、除外判定を共有ヘルパへ抽出して再ドリフトを防ぐ。
- **[LSA-04] clinic_id を持たない 3 テーブルを院単位 RBAC ルートで作成・削除できる — HIGH / 着手: 要裁定**
  - `lstep_auto_managed_prefixes` / `lstep_condition_tag_mappings` / `lstep_send_purpose_tag_prefixes`（`001_init.sql:619,635,649`）は clinic_id を持たない全院共有行だが、route は `ResourceHospitalSettings` の院単位権限で GET/POST/DELETE を公開し、repository にも clinic 述語がない。**A 院の設定変更が全院に波及する**。
  - 対応: **設計意図の確定が先**。全院共通マスタが正なら platform-admin 専権へ、院ごとが正なら `clinic_id NOT NULL` + UNIQUE の incremental migration。
- **[LSA-07] LINE User ID を平文でログ出力 — HIGH / 着手: 即時**
  - `line_messaging_service.go:81` の `slog.InfoContext(ctx, "LINE push sent", "to", lineUserID)`。同 package の他経路は一貫して `owner_id` のみをログしており、ここだけが例外。規約 `backend/CODING_RULES.md:68` が明示的に禁じる owner data のログ出力。
- **[LSA-15] 配信の二重発火防止が check-then-Create のみで DB 一意制約がない — MEDIUM**
  - `lstep_delivery_trigger_log` に `(clinic_id, owner_id, trigger_type, 日付)` の UNIQUE がなく非一意 index のみ（`001_init.sql:552-553`）。durable scheduler の再実行や cron と event 駆動の同時実行で同日二重配信が成立する。部分一意 index の追加で解消する。

### 診療記録（medicalrecord）

- **[MRC-02] 薬剤削除の在庫カスケードが FK ではなく可変の name をキーにしている — HIGH / 着手: 納品後**
  - `medicine_service.go:312-320` が生成した inventory item の id を `medicines.inventory_id` へ書き戻さないため、削除・改名同期が `(clinic_id, name, category)` で対象を選ぶ。`inventory/repository.go:150-172` は `RowsAffected` を検査せず 0 件でも nil を返す。**在庫を改名してから薬剤を削除すると在庫行が孤児化し、誰も気づかない**。
- **[MRC-04] カルテ作成の主訴・治療方針・診断がサイレントに消失し 201 を返す — HIGH / 着手: 即時**
  - `medical_record_subrecords.go` の `CreateSubRecords` が戻り値を持たず、inquiry upsert 失敗・所有権検証失敗・診断 FK 検証失敗が全て slog.Warn で終わる。`medical_record_handler.go:117` は戻り値を受けず直後に 201 Created を返す。臨床記録の欠落が API 上は成功に見える。
- **[MRD-01] 治療項目の並び順一括更新が affected rows を確認せず、親カルテに束縛されていない — HIGH**
  - `treatment_repository.go:260-265` は `RowsAffected` を見ず、`:261` の WHERE が service の施錠済み `medicalRecordID` に束縛されていない。存在しない ID や別カルテの ID でも 204 が返る。同 package の `Update`/`Delete` は `RowsAffected==0` を NotFound へ写像しており非対称。
- **[MRD-03] treatment plan の write が親（カルテ/入院）所属を検証しない — HIGH**
  - handler は親の所属を検証するが（`treatment_plan_handler.go:181,216`）、直後に呼ぶ `service.Update/Delete` は planID を clinicID だけで解決する。**同一 clinic 内で別カルテの plan を書き換えられる**（resource ownership 検証の欠落・`invariants.md:22`）。
- **[MRB-02] `hospitalizationRepository.LockByIDForUpdate` が ambient tx 不在を fail-closed にしない — HIGH**
  - `FOR UPDATE` を発行するが `persistence.TxFromContext(ctx) == nil` を検査せず、doc コメント（`:98`）で危険性を散文で認めるにとどまる。**同 package の `examinationRepository` は `examination_repository.go:102-104` で同じガードを実装済み**であり、片側だけが欠けている。

### コマンド（cmd）

- **[CMD-02] 全院バッチを起動する `/_internal/scheduled-jobs` が未認証の root engine に登録されている — MEDIUM / 着手: 要裁定**
  - `cmd/api/base_routes.go:26` が `*gin.Engine` へ直接登録し、`main.go:198-203` の middleware に認証はない。`RunRequest.validate` が要求する値は全て repo 内の公開定数から導出できる。**防御は Cloudflare Worker の edge path filter 一枚**（`worker/index.ts:236`）— これは文書化・test 済みの設計上の境界なので即時の脆弱性ではないが、Go 側の route 登録面には境界が明示されていない。
- **[CMD-05] `/uploads` の StaticFS が無条件・未認証で登録されている — MEDIUM**
  - `STORAGE_TYPE` が s3 以外のときカルテ画像が同じ木へ書かれるため、認証も clinic scope も経ずに PHI が取得できる。`config.go:207-208` が release mode で s3 を強制するため production では実害なし。登録を `cfg.StorageType != "s3"` に限定する。

### 飼主・ペット・医院（pet / owner / clinic）

- **[POC-08] clinic スコープ付き飼主更新 route 5 本が OpenAPI 未宣言 — MEDIUM**
  - `owner/http_routes.go:24-29` の 5 PATCH は `:clinic_id` をハンドラが一度も読まず（認証済み clinic を使う）、`api.yaml` にも宣言がない。撤去して `/owners/:id` へ一本化するか、`:clinic_id` を実際に検証して使う。規約 `CODING_RULES.md:53`（OpenAPI 同期）違反。
- **[POC-01] 休診日ミューテーションが 2 route 群に二重登録され必要権限が分岐 — MEDIUM**
  - `ResourceShifts` と `ResourceClosingSettings` の 2 経路で同一ハンドラが登録され、既定権限表では前者が一般グループ create=true、後者は不可。**意図的な委譲であることは `api.yaml:12119-12124` に明記済み**だが、権限差は意図の記述がない。1 経路へ統一するか権限を揃える。
- **[POC-05 / POC-06] 一意性・重複禁止が application 検証のみで DB 制約がない — MEDIUM**
  - 特別期間の重複（`closing_settings_service.go:183-203`）と飼主 phone（`owner/service_core.go:163-172`）。**email には `uk_owners_clinic_email` があるのに phone にはない**。phone が重複すると `FindByNameAndPhone` が nil を返し **LINE 自動紐付けが黙って不成立になる**。部分一意インデックスの追加で解消。
- **[POC-07] 全院共有マスタ `animal_species` の更新・削除に監査がなく、物理削除である — MEDIUM**
  - `model/animal_species.go` に `DeletedAt` がないため repository の Delete は物理削除。各クリニックの権限だけで実行でき、実行者も変更前後値も残らない。

### 予約（reservation）

- **[RSV-02] 管理画面予約作成だけが `AcquireBookingLock` を取得しない — HIGH / 着手: 即時**
  - `reservation_repository.go:164-176` は「空き枠では FOR UPDATE が何もロックしないためファントムで両方成功しうる」「必ず本メソッドを先頭で呼ぶこと」と不変条件を宣言し遵守する呼び出し元を 3 箇所列挙しているが、**`appointment_admin_service.go:128` はその列挙から漏れている**。空き枠へのファントム二重予約と AB-BA デッドロックを許す。
  - 対応: `WithTx` 先頭で `AcquireBookingLock` を呼び、宣言側の列挙を 4 件へ更新する。**呼び出し元の列挙をコメントで維持する方式自体が破綻している**ため、AST gate 化を検討する。

### スタッフ・認証（staff / auth）

- **[AUS-04] 死んだ method と恒真テストが interface に露出している — MEDIUM**
  - `shift_template_repository.go:150-154` の `CountUsageByShiftTemplateID` は production consumer が 0 件で、実装は全引数を捨てて `return 0, nil`。**integration test（`:264-296`）は 3 ケースとも `assert.Equal(int64(0), count)` で恒真**であり回帰検出能力がゼロ。interface・実装・mock・test・誤ったヘッダコメントを同時に除去する。
- **[AUS-05] `toShiftResponse` が nil ガードなしで `*model.Staff` を参照する — MEDIUM**
  - `shift_response.go:63` は `s.Staff.ID != 0` と無条件参照するが、`shift_entry_repository.go:20-22` のコメント自身が「Staff association を隠したままシフトを見せる」設計だと明記している。**nil panic の経路**。`toStaffSummary` と同じガード形へ。
- **[AUS-06] `auth/http_session.go` が 820 行で上限超過 — MEDIUM**
  - `~/.claude/rules/ecc/common/coding-style.md:39`（800 max）。担当 58 ファイル中これ 1 本のみ。**根拠区分は project quality policy であり Go/Gin 公式要件ではない**（`guidelines:227` が固定サイズを公式要件から明示除外）。cookie 発行・logout 監査 identity 解決の分離で凝集度も上がる。

### モデル・静的 lint（model / lintscan / testdb）

- **[MDL-05] CASCADE lint が完全一致リテラル検索のため表記ゆれを見逃す — MEDIUM / 着手: 即時**
  - `migration_cascade_lint_test.go:62-64` は `strings.Count(sql, "ON DELETE CASCADE")` のみ。PostgreSQL は小文字・複数空白・改行分割も等価に受理するが、いずれもカウント 0 になる。**PHI 親テーブルへの CASCADE を含む新規 migration が gate を無条件通過する**（`migrations/CLAUDE.md:18` が「絶対禁止」とする設計）。正規表現 `(?is)ON\s+DELETE\s+CASCADE` へ置換し allowlist 値を再ピンする。

---

## 既知（再記述せずポインタのみ）

| 所見 | 正本 |
|---|---|
| stage-import の clinic 非限定 DELETE | [`#BUG-430`](3-session-agent.html#BUG-430)（DEC-24 で削除退役を裁定済み） |
| `hospitalization_repository.go:86,87` の Preload に clinic 述語なし | [`#BUG-437`](3-session-agent.html#BUG-437) / [`#SEC-SWEEP-02`](3-session-agent.html#SEC-SWEEP-02) の**面追加**（新規起票しない） |
| `shift_entry_breaks` の孫 read に親 clinic 相関なし | [`#SEC-SWEEP-02`](3-session-agent.html#SEC-SWEEP-02) の未掃引面（`grandchild_parent_clinic_correlation_lint_test.go` の registry にも未登録） |
| `trimming` の Options preload が中間 junction を相関しない | [`#SEC-SWEEP-02`](3-session-agent.html#SEC-SWEEP-02)。新規事実は 1 点のみ = `appointment_trimming_options`（`001_init.sql:1380-1387`）に clinic_id 列がない |
| `model.Payment` に ClinicID がない | [`#TASK-445`](3-session-agent.html#TASK-445)（DEC-28 で案 b を裁定済み）。model 側の追加も同 unit に含める |
| inquiry upsert の finalized ガード非原子 | `phase2.html:195`（Transactor + LockByIDForUpdate の正規パターン統一） |

---

## 規約側の疑義（30 論点・重複排除後）

**規約が曖昧・自己矛盾・実装と乖離している箇所**。実装を直す前にこちらを確定しないと、修正が振動する。

1. **`invariants.md:11` × `migrations/CLAUDE.md:14` — 「clinic-scoped data」の判定基準がない。** clinic_id 列を持たないが院単位 RBAC ルートから変更できるテーブル（lstep タグ設定 3 表・`animal_species`）が scope 対象かを判定できない。LSA-04 / POC-07 の裁定はこれに依存する。
2. **`error-handling.md:9` × `CODING_RULES.md:36` — 「error を無視しない」と「best-effort を許容する」が競合。** どちらが優先か、best-effort を選べる条件は何かが未定義。X-04 の 12 箇所はこの線引きなしには判定が割れる。
3. **`invariants.md:35` — best-effort の 4 要件（補償・再試行・監査・部分失敗 contract）を満たす受理条件が未定義。** 特に「同期 HTTP 応答が何を返してよいか」が規定されておらず、MRC-04（201 を返す）の是非が規約から導けない。
4. **`invariants.md:37` — 「fail-closed と定めた」監査がどれか列挙されていない。** 指定行為を前提とする規定なのに指定台帳がない。また「締め後の会計編集」の定義が `billings` 直接 UPDATE のみか `billing_items` 経由も含むかが不明で、**X-A の適用範囲がこれで変わる**。
5. **`invariants.md:31` — グローバル共有マスタに対する「対象 scope」の意味が未定義。**
6. **`invariants.md:15` — 「preload にも同じ scope」と無条件に書くが、機械 gate（`preload_clinic_scope_lint_test.go:129-146`）は `staffExemptAssoc` 8 association を理由付きで除外している。** 規約と gate の非一致。
7. **`guidelines:151` — 金額・税率の範囲 contract が規約正本のどこにもない。** 実際の contract は実装の copy-paste としてのみ存在する（X-02 / X-09 の根本原因）。
8. **`guidelines:179` — body size 制限の適用単位（グローバル middleware か handler ごとか）が未定義。** 実装は 6 package が handler ごと、3 package が無制限（X-07）。
9. **`guidelines:166` — PostgreSQL エラーコード → HTTP status のマッピングをどの層が所有するかが未定義。** `apperrors/errors.go:183-192` と `httpapi/response_pg.go` の二重管理になっている（INF-01 の構造的原因）。
10. **`error-handling.md:18` — `errors.Is/As` 判定の例外規定がない。** pgx の encode 失敗は型付き sentinel を公開しておらず、現状の実装は規約を満たせない（INF-04）。
11. **`guidelines:134` — public/authenticated 境界の明示を Go の route 登録面に閉じて要求しているが、本 system の `/_internal` 境界は Cloudflare Worker 側にある**（CMD-02 の評価が規約から導けない）。
12. **`guidelines:194` §11 — Production server lifecycle が HTTP server 前提で、破壊的 DB 権限を持つ 7 本の one-shot CLI（migrate / stage-import / csv-import 等）に対する規定がない。**
13. **`.claude/CLAUDE.md:7`「Prohibit `any`」が無条件禁止として書かれているが、GORM の `Updates` 列名マップとして `map[string]any` が広範に使われている**（medicalrecord の 41 ファイル中 13 ファイル・39 箇所）。禁止の対象範囲を明確化するか、GORM 更新マップを明示的な例外にする。
14. **`guidelines:227,233` × `coding-style.md:39` — ファイルサイズ上限が「公式要件から除外」と「800 max」で正面衝突。** AUS-06 の根拠区分がこれで変わる。
15. **`coding-style.md:64`（Constants: `UPPER_SNAKE_CASE`）が Go の慣用（MixedCaps）と衝突。** 言語スコープ宣言がないグローバル規約であり、本 repo の Go const は全て camelCase/PascalCase。
16. **`go-gin-backend-review.md:68` — 「暗黙条件に依存していないか」が GORM の `gorm.DeletedAt` に対して字義どおりには充足不能**（GORM は `deleted_at IS NULL` を常に暗黙付与する）。
17. **`go-gin-backend-review.md:69` — 「実DB の integration test」の『実DB』が、本物の PostgreSQL か、本番同一 schema かを規定していない。**
18. **`CLAUDE.md:44` — 退役した P1–P18 番号を「レビュー基準に使わない」と定めるが、production code 内に残る P 番号コメントの扱いを規定していない。**
19. **`invariants.md:32` — `appointment_write_owner_lint_test.go` による gate 維持を謳うが、当該 gate の実際の検査範囲が規約の記述より狭い。**
20. **`CLAUDE.md:25` — 「自動化には停止、失敗通知」の『失敗通知』に受理条件がない。** lstep バッチ群は `slog.ErrorContext` のみで満たしていると読めてしまう（X-04 の判定に直結）。

（残り 10 論点は同型または個別 unit 固有のため本節では省略。全 41 件の原文は本監査の journal に保存されている。）

---

## 監査の限界（次回の改善点）

1. **lstep の 45 ファイルが全文未読**（grep 全数走査で代替）。次回は lstep を 3 分割する。
2. **性能系（N+1・unbounded query）を実質的に監査できていない**。read-only 制約下ではクエリプランが取れないため、規約 `go-gin-backend-review.md` の該当条項は実質未検査のまま残っている。
3. **並行性シナリオは全て机上導出**。X-05 / RSV-02 の TOCTOU は実 DB での再現を伴っていない。
4. **テストコードを監査対象から除外した**（各 unit は非テスト `.go` のみ）。AUS-04 の恒真テストは production 側の調査中に偶然発見されたものであり、**テスト品質の系統的監査は未実施**である。
