# C05 — 境界入力バリデーション (boundary input validation)

Campaign: emr-332-security-defense / Unit: U03-INPUT-OUTPUT / Attempt: att-u03-2 (att-u03-1 証跡訂正 — verdict 維持)
Template: `docs/work/emr-332-security/u01-case-verdict-template.md` (12 fields)

## 判定レコード

| # | Field | 記録 |
|---|-------|------|
| 1 | ケース ID | C05 |
| 2 | 対象リビジョン | worktree `feat/emr-332-security-defense-w03` — campaign base `de28e9765a29091339a872a85f1875a947e07e5b` (u01-safety-record §task branch) + 本ユニット追加分 (emr332_u03 package + att-u03-2 の doc-comment 訂正のみ)。staging 実測リビジョン照合なし (static/simulated のみのため不要) |
| 3 | 実施方式 (method) | `static` (bind 呼出し経路・struct タグの棚卸し) + `simulated` (httptest + gin.CreateTestContext によるメモリ内 bind 実測)。staging-observed なし |
| 4 | 正常条件 | 期待: 必須・型・範囲・enum 適合の JSON は bind 成功 / 観測: `{"name":"SECCHK20261009C05R05","status":"alive","age":7}` → bind 成功。実ハンドラ struct `medicalrecord.CreateMedicalRecordAddendumRequest` (`after_text`/`reason` required,max) の有効本文も成功 — `TestC05RealHandlerStruct/normal:_valid_addendum` PASS |
| 5 | 拒否条件 | 期待: 型不一致・必須欠落・max 超過・oneof 違反・数値範囲外・未知(disallowed)フィールドの拒否 / 観測: 型不一致(`age:"SECCHK20261009C05R07"`)・必須欠落・`max=255` 超過・oneof 違反 (`status="SECCHK20261009C05R06"`)・`min=0`/`max=200` 違反は **全て bind error で拒否** (`ParseBindError` が sanitized 日本語 message を返す。`internal/httpapi/bind_errors.go:15-34`)。**未知フィールド `secchk20261009c05r08` はエラーなく受理され黙って破棄される — FAIL**。実 struct でも同結果 (`TestC05RealHandlerStruct/gap:_unknown_field_silently_accepted`)。trailing 第 2 JSON value も `ShouldBindJSON` 単独経路では未読のまま黙殺 (`TestC05ShouldBindJSON_TrailingValueUnread`) |
| 6 | 防御判定 | **FAIL** — type/range/required の拒否は成立するが、disallowed-field (unknown field) 拒否クラスが全 `ShouldBindJSON`/`bindStaffJSON`/`bindAuthJSON` 経路で不成立。証拠はテスト実測 + 設定不在の静的確認 |
| 7 | 副作用呼出回数 | **0** — メモリ内 bind のみ。emr332_u03 package は service/repository/session/audit/cache/queue/file/mail/telemetry を一切 import せず、save-path sentinel の呼出しは `TestC05ZeroSideEffectInvocations` で機械的に 0 を確認 |
| 8 | 検出判定 | **N/A** — bind 層に unknown-field 受理を検出する経路は存在しない (silent drop が性質)。拒否側 (400) は `RequestLoggingMiddleware` の構造化ログ経路が静的に存在するが、本ユニットは HTTP スタック越しの検出実測は行わない (method=simulated は bind 呼出しまで) |
| 9 | 安全な証跡参照 | テスト: `backend/internal/middleware/emr332_u03/bind_boundary_test.go` (6 テスト関数), `strict_bind_test.go` (7 テスト関数)。実測出力: 本ファイル末尾「実測コマンド出力」 (att-u03-2 で再取得した真の verbatim)。静的根拠: `internal/staff/http_binding.go:25-52` (bindStaffJSON — 上限+trailing 検査 + `validateStaffJSONByteLengths` :64-78、DisallowUnknownFields なし), `internal/auth/http_binding.go:25-61` (同様 + `validateAuthJSONByteLengths` :63-91), `gin.EnableDecoderDisallowUnknownFields` 設定箇所 0 件 (rg 全走査), `internal/owner/http_request.go:119-256` (required/max/oneof タグ例), `internal/reservation/liff_request.go:89-101` |
| 10 | 制約・未検証範囲 | 未検証: (a) 全 226 bind 入口経路 (`.ShouldBindJSON(` 直接 202 + `bindStaffJSON` 16 + `bindAuthJSON` 8 — 集計方法は下表) を個別には実測していない — 代表 mirror struct (`binding:"required,max=..."`/`oneof`/`min/max` 系) + 実 exported struct 1 種で binding engine の共通契約を実測した。gin v1.12.0 の decoder は全経路共通のため拡張妥当。(b) クエリ文字列・フォーム・パスパラメータの bind (ShouldBindQuery 等) は未検証。(c) unknown field が「黙殺」されることの業務影響 (typo した有効 field の黙殺等) は本ケースの防御判定とは別。(d) handler への実配線は未実施 (owned_paths 外のため)。(e) `BindStrictJSON` は wrapper 各社の per-type byte-length 検査 (email/password/token 等) を複製しない — doc comment に明記済み、wire-in 時は同等検査の移植が必要 |
| 11 | 修復案 | **適用済み (owned path 内)**: `BindStrictJSON(c, dst, maxBytes)` — `backend/internal/middleware/emr332_u03/strict_bind.go`。bounded read→`DisallowUnknownFields` decode→trailing 値拒否→`binding.Validator.ValidateStruct` で `binding` タグを再検証。error は `apperrors.WrapInvalidInput(httpapi.ParseBindError(err))` で既存応答形状と同一。**候補 (owned_paths 外のため未適用)**: 各 handler の `c.ShouldBindJSON`/`bindStaffJSON`/`bindAuthJSON` を `BindStrictJSON` へ置換 (byte-length 検査の移植を伴う)、または `gin.EnableDecoderDisallowUnknownFields = true` のグローバル設定 + 既存 wrapper への trailing 検査維持。in-place な handler 変更は本ユニットでは行わない (scope 規約) |
| 12 | 再検証結果 | 実施済み — scoped isolated Docker `go test ./internal/middleware/emr332_u03/` **13 テスト関数全 PASS** (bind_boundary_test.go 6 関数 + strict_bind_test.go 7 関数; subtest 延べ 20)。出力末尾は att-u03-2 で新規取得した真の verbatim。helper は unknown field / trailing value / oversized body / 宣言 Content-Length / required / oneof / range / 型違反を全て拒否。回帰: `go test ./internal/middleware/` PASS (sanitize_null_bytes 既存テスト含む)。`go vet` exit 0 |

## 補足: 棚卸し結果

### bind 経路 (非テストファイル)

| 経路 | 箇所数 | DisallowUnknownFields | 集計方法 |
|------|--------|------------------------|----------|
| `c.ShouldBindJSON` 直接 (handler) | **202** | なし — Gin decoder は常に非 strict | `rg -n '\.ShouldBindJSON\(' internal/ -g '*.go' -g '!*_test.go'` → 204、うち wrapper 内部 2 (staff/http_binding.go:48, auth/http_binding.go:57) を除く |
| `bindStaffJSON` (staff) | **16** call sites | なし (http_binding.go:25-52 は上限+trailing+byte-length のみ) | `rg -n 'bindStaffJSON\(c,' internal/staff -g '*.go' -g '!*_test.go'` |
| `bindAuthJSON` (auth) | **8** call sites | なし (http_binding.go:25-61 同様) | `rg -n 'bindAuthJSON\(c,' internal/auth -g '*.go' -g '!*_test.go'` |
| 手動 decoder + DisallowUnknownFields | 9 sites (internal/ 8 + cmd/ 1) | **あり** | `rg -n 'DisallowUnknownFields' internal/ cmd/ -g '*.go' -g '!*_test.go'` (emr332_u03 helper 自身と doc comment 言及を除く) |

→ **JSON bind 入口経路の合計 = 226** (202 + 16 + 8)。`ShouldBindJSON` という文字列の出現自体は非テスト 209 件 (呼出し 204 + comment/identifier 言及 5) — att-u03-1 が「209 direct sites」と記したのは出現数と呼出し数の混同であり、呼出し site 数は 202/204 が正しい。

### DisallowUnknownFields 既存適用箇所 (pre-existing, 本ユニット追加分を除く)

| file:line | 到達面 |
|-----------|--------|
| `internal/staff/staff_provisioning.go:154` | provisioning manifest (CLI/ファイル経路) |
| `internal/staff/staff_provisioning.go:199` | provisioning secrets file |
| `internal/scheduler/handler.go:260` | scheduler ops 経路 (Worker 内部 / Bearer または CF Access 認証) |
| `internal/lstep/line_link_handler.go:129` | LINE link 経路 (HTTP) |
| `internal/billing/billing_confirmation_handler.go:162` | billing confirmation (HTTP, + `hasOnlyExactStringFields` 二重検査) |
| `internal/medicalrecord/lab_device_receive_repository.go:371` | lab device `slots_json` 内部フィールド検証 |
| `internal/medicalrecord/checkup_package_manifest.go:60` | 健診パッケージ manifest decode |
| `internal/csvimport/cutover_contract_validate.go:108` | csvimport contract file |
| `cmd/stg-uat-staff-attach/validate.go:21` | CLI 検証ツール (non-HTTP) |

→ 主要 JSON API 書込み経路 (owner/pet/medicalrecord/reservation/staff/lstep/trimming/billing 等の `ShouldBindJSON` 系) には unknown-field 拒否が**皆無**。strict なのは manifest/ファイル/特定内部経路のみ。

## 実測コマンド出力 (verbatim — att-u03-2 で worktree `feat/emr-332-security-defense-w03` から新規取得)

```text
$ docker run --rm --pull never --network none \
    -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-emr-332-security-defense-w03/backend:/app:ro \
    -v ekarte-go-mod-cache:/go/pkg/mod \
    -v ekarte-go-build-cache:/root/.cache/go-build \
    -w /app golang:1.26-alpine \
    go test -count=1 -v ./internal/middleware/emr332_u03/
=== RUN   TestC05ShouldBindJSON_TagEnforcement
=== RUN   TestC05ShouldBindJSON_TagEnforcement/normal:_valid_body
=== RUN   TestC05ShouldBindJSON_TagEnforcement/refusal:_missing_required
=== RUN   TestC05ShouldBindJSON_TagEnforcement/refusal:_required_empty
=== RUN   TestC05ShouldBindJSON_TagEnforcement/refusal:_over_max=255
=== RUN   TestC05ShouldBindJSON_TagEnforcement/refusal:_enum_violation
=== RUN   TestC05ShouldBindJSON_TagEnforcement/refusal:_wrong_type_string->int
=== RUN   TestC05ShouldBindJSON_TagEnforcement/refusal:_range_min
=== RUN   TestC05ShouldBindJSON_TagEnforcement/refusal:_range_max
--- PASS: TestC05ShouldBindJSON_TagEnforcement (0.00s)
    --- PASS: TestC05ShouldBindJSON_TagEnforcement/normal:_valid_body (0.00s)
    --- PASS: TestC05ShouldBindJSON_TagEnforcement/refusal:_missing_required (0.00s)
    --- PASS: TestC05ShouldBindJSON_TagEnforcement/refusal:_required_empty (0.00s)
    --- PASS: TestC05ShouldBindJSON_TagEnforcement/refusal:_over_max=255 (0.00s)
    --- PASS: TestC05ShouldBindJSON_TagEnforcement/refusal:_enum_violation (0.00s)
    --- PASS: TestC05ShouldBindJSON_TagEnforcement/refusal:_wrong_type_string->int (0.00s)
    --- PASS: TestC05ShouldBindJSON_TagEnforcement/refusal:_range_min (0.00s)
    --- PASS: TestC05ShouldBindJSON_TagEnforcement/refusal:_range_max (0.00s)
=== RUN   TestC05ShouldBindJSON_UnknownFieldSilentlyAccepted
--- PASS: TestC05ShouldBindJSON_UnknownFieldSilentlyAccepted (0.00s)
=== RUN   TestC05ShouldBindJSON_TrailingValueUnread
--- PASS: TestC05ShouldBindJSON_TrailingValueUnread (0.00s)
=== RUN   TestC05DisallowUnknownFields_Contrast
--- PASS: TestC05DisallowUnknownFields_Contrast (0.00s)
=== RUN   TestC05RealHandlerStruct
=== RUN   TestC05RealHandlerStruct/normal:_valid_addendum
=== RUN   TestC05RealHandlerStruct/refusal:_missing_required
=== RUN   TestC05RealHandlerStruct/refusal:_over_max=1000
=== RUN   TestC05RealHandlerStruct/refusal:_wrong_type
=== RUN   TestC05RealHandlerStruct/gap:_unknown_field_silently_accepted
--- PASS: TestC05RealHandlerStruct (0.00s)
    --- PASS: TestC05RealHandlerStruct/normal:_valid_addendum (0.00s)
    --- PASS: TestC05RealHandlerStruct/refusal:_missing_required (0.00s)
    --- PASS: TestC05RealHandlerStruct/refusal:_over_max=1000 (0.00s)
    --- PASS: TestC05RealHandlerStruct/refusal:_wrong_type (0.00s)
    --- PASS: TestC05RealHandlerStruct/gap:_unknown_field_silently_accepted (0.00s)
=== RUN   TestC05ZeroSideEffectInvocations
--- PASS: TestC05ZeroSideEffectInvocations (0.00s)
=== RUN   TestBindStrictJSON_AcceptsValidBody
--- PASS: TestBindStrictJSON_AcceptsValidBody (0.00s)
=== RUN   TestBindStrictJSON_RejectsUnknownField
--- PASS: TestBindStrictJSON_RejectsUnknownField (0.00s)
=== RUN   TestBindStrictJSON_RejectsTrailingJSONValue
--- PASS: TestBindStrictJSON_RejectsTrailingJSONValue (0.00s)
=== RUN   TestBindStrictJSON_RejectsOversizedBody
--- PASS: TestBindStrictJSON_RejectsOversizedBody (0.00s)
=== RUN   TestBindStrictJSON_RejectsOversizedDeclaredLength
--- PASS: TestBindStrictJSON_RejectsOversizedDeclaredLength (0.00s)
=== RUN   TestBindStrictJSON_EnforcesBindingTags
=== RUN   TestBindStrictJSON_EnforcesBindingTags/missing_required
=== RUN   TestBindStrictJSON_EnforcesBindingTags/required_empty
=== RUN   TestBindStrictJSON_EnforcesBindingTags/over_max_length
=== RUN   TestBindStrictJSON_EnforcesBindingTags/enum_violation
=== RUN   TestBindStrictJSON_EnforcesBindingTags/range_violation
=== RUN   TestBindStrictJSON_EnforcesBindingTags/negative_range
=== RUN   TestBindStrictJSON_EnforcesBindingTags/wrong_type
--- PASS: TestBindStrictJSON_EnforcesBindingTags (0.00s)
    --- PASS: TestBindStrictJSON_EnforcesBindingTags/missing_required (0.00s)
    --- PASS: TestBindStrictJSON_EnforcesBindingTags/required_empty (0.00s)
    --- PASS: TestBindStrictJSON_EnforcesBindingTags/over_max_length (0.00s)
    --- PASS: TestBindStrictJSON_EnforcesBindingTags/enum_violation (0.00s)
    --- PASS: TestBindStrictJSON_EnforcesBindingTags/range_violation (0.00s)
    --- PASS: TestBindStrictJSON_EnforcesBindingTags/negative_range (0.00s)
    --- PASS: TestBindStrictJSON_EnforcesBindingTags/wrong_type (0.00s)
=== RUN   TestBindStrictJSON_RejectsMalformedAndEmpty
--- PASS: TestBindStrictJSON_RejectsMalformedAndEmpty (0.00s)
PASS
ok  	github.com/animal-ekarte/backend/internal/middleware/emr332_u03	0.004s
```

```text
$ docker run --rm --pull never --network none \
    -v /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-emr-332-security-defense-w03/backend:/app:ro \
    -v ekarte-go-mod-cache:/go/pkg/mod -v ekarte-go-build-cache:/root/.cache/go-build \
    -w /app golang:1.26-alpine \
    sh -c 'go test -count=1 ./internal/middleware/ && go vet ./internal/middleware/emr332_u03/ && echo "VET-EXIT:0"'
ok  	github.com/animal-ekarte/backend/internal/middleware	0.043s
VET-EXIT:0
```

## 訂正履歴 (correction lineage — att-u03-2)

att-u03-1 の記録証跡に対する G2 IMPL_DEFECT 指摘への訂正。**判定 (verdict) 自体は独立検証で正しいことが確認済みであり変更なし** (C05 FAIL 維持)。

| 箇所 | 誤り (att-u03-1) | 訂正後 (att-u03-2) |
|------|------------------|--------------------|
| field 12 再検証結果 | 実在しない「16 件のテスト全 PASS」相当の総数を記載 | 実際は **13 テスト関数** (bind_boundary_test.go 6 + strict_bind_test.go 7) — `grep -c '^func Test'` で確認 |
| 実測コマンド出力 block | 「verbatim」標記ながら一部合成: `TestBindStrictJSON_EnforcesBindingTags` の親 PASS 行に go test が出力しない "(N subtests) all PASS" 形式の合成サマリ文字列を付記 (N=8 と記載したが定義は 7 — strict_bind_test.go table cases) | att-u03-2 で isolated Docker から取得した**真の raw 出力**に全面置換。合成行なし、subtest は定義通り 7 |
| bind 経路 table (旧「209 (うち 3 件は wrapper 内部…)」) | `ShouldBindJSON` 文字列出現数 (209) を呼出し site 数と混同 | 機械集計に修正: `.ShouldBindJSON(` 呼出し site 204 (うち wrapper 内部 2) → 直接 202。`bindStaffJSON` 16, `bindAuthJSON` 8 は従来通り (call-site grep で確認) |
| field 10 (旧「全 233 bind 経路」) | 233 は誤った合計 | **226** = 202 直接 + 16 + 8 (上表の集計方法欄に再現コマンドを明記) |
| field 9 `internal/auth/http_binding.go` cite | att-u03-1 の記載 :25-61 は正しかった。att-u03-2 は「関数開始行は :26 (:25 は空行)」を理由に :26-61 へ変更したが、実際の :25 は `func bindAuthJSON(c *gin.Context, destination any) error {` signature 行であり正当化は誤り (G2 residual F-3) | **:25-61 に復元** (byte-length validator :63-91 の併記は維持 — field 9 / bind 経路表とも) |
| strict_bind.go doc comment (LOW) | 「the same boundary contract」と記し wrapper の per-type byte-length 検査 (validateStaffJSONByteLengths/validateAuthJSONByteLengths) 非複製を未記載 | doc comment を訂正 — core contract の共有と byte-length 検査非複製・wire-in 時の移植要否を明記 (behavior 変更なし) |
| hygiene pass (att-u03-2 G2 residual F-1..F-7) | c05 該当は F-3 のみ — `bindAuthJSON` cite の :26-61 と「:25 は空行」の偽正当化 (直上の行参照) | :25-61 へ復元済み。判定 (FAIL)・集計 (226 経路 / 13 テスト関数 PASS) は不変 |

*Marker 文字列は `SECCHK20261009C05Rnn` 形の英数字 sentinel のみ。実行可能な攻撃構文・実ペイロード・資格情報・PII は本ファイルに含まない (A5)。*
