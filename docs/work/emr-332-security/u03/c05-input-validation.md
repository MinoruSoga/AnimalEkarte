# C05 — 境界入力バリデーション (boundary input validation)

Campaign: emr-332-security-defense / Unit: U03-INPUT-OUTPUT / Attempt: att-u03-1
Template: `docs/work/emr-332-security/u01-case-verdict-template.md` (12 fields)

## 判定レコード

| # | Field | 記録 |
|---|-------|------|
| 1 | ケース ID | C05 |
| 2 | 対象リビジョン | worktree `feat/emr-332-security-defense-w03` — campaign base `de28e9765a29091339a872a85f1875a947e07e5b` (u01-safety-record §task branch) + 本ユニット追加分 (emr332_u03 package のみ)。staging 実測リビジョン照合なし (static/simulated のみのため不要) |
| 3 | 実施方式 (method) | `static` (bind 呼出し経路・struct タグの棚卸し) + `simulated` (httptest + gin.CreateTestContext によるメモリ内 bind 実測)。staging-observed なし |
| 4 | 正常条件 | 期待: 必須・型・範囲・enum 適合の JSON は bind 成功 / 観測: `{"name":"SECCHK20261009C05R05","status":"alive","age":7}` → bind 成功。実ハンドラ struct `medicalrecord.CreateMedicalRecordAddendumRequest` (`after_text`/`reason` required,max) の有効本文も成功 — `TestC05RealHandlerStruct/normal:_valid_addendum` PASS |
| 5 | 拒否条件 | 期待: 型不一致・必須欠落・max 超過・oneof 違反・数値範囲外・未知(disallowed)フィールドの拒否 / 観測: 型不一致(`age:"SECCHK20261009C05R07"`)・必須欠落・`max=255` 超過・oneof 違反 (`status="SECCHK20261009C05R06"`)・`min=0`/`max=200` 違反は **全て bind error で拒否** (`ParseBindError` が sanitized 日本語 message を返す。`internal/httpapi/bind_errors.go:15-34`)。**未知フィールド `secchk20261009c05r08` はエラーなく受理され黙って破棄される — FAIL**。実 struct でも同結果 (`TestC05RealHandlerStruct/gap:_unknown_field_silently_accepted`)。trailing 第 2 JSON value も `ShouldBindJSON` 単独経路では未読のまま黙殺 (`TestC05ShouldBindJSON_TrailingValueUnread`) |
| 6 | 防御判定 | **FAIL** — type/range/required の拒否は成立するが、disallowed-field (unknown field) 拒否クラスが全 `ShouldBindJSON`/`bindStaffJSON`/`bindAuthJSON` 経路で不成立。証拠はテスト実測 + 設定不在の静的確認 |
| 7 | 副作用呼出回数 | **0** — メモリ内 bind のみ。emr332_u03 package は service/repository/session/audit/cache/queue/file/mail/telemetry を一切 import せず、save-path sentinel の呼出しは `TestC05ZeroSideEffectInvocations` で機械的に 0 を確認 |
| 8 | 検出判定 | **N/A** — bind 層に unknown-field 受理を検出する経路は存在しない (silent drop が性質)。拒否側 (400) は `RequestLoggingMiddleware` の構造化ログ経路が静的に存在するが、本ユニットは HTTP スタック越しの検出実測は行わない (method=simulated は bind 呼出しまで) |
| 9 | 安全な証跡参照 | テスト: `backend/internal/middleware/emr332_u03/bind_boundary_test.go`, `strict_bind_test.go`。実測出力: 本ファイル末尾「実測コマンド出力」。静的根拠: `internal/staff/http_binding.go:25-52` (bindStaffJSON — 上限+trailing 検査あり、DisallowUnknownFields なし), `internal/auth/http_binding.go:25-61` (同様), `gin.EnableDecoderDisallowUnknownFields` 設定箇所 0 件 (rg 全走査), `internal/owner/http_request.go:119-256` (required/max/oneof タグ例), `internal/reservation/liff_request.go:89-101` |
| 10 | 制約・未検証範囲 | 未検証: (a) 全 233 bind 経路を個別には実測していない — 代表 mirror struct (`binding:"required,max=..."`/`oneof`/`min/max` 系) + 実 exported struct 1 種で binding engine の共通契約を実測した。gin v1.12.0 の decoder は全経路共通のため拡張妥当。(b) クエリ文字列・フォーム・パスパラメータの bind (ShouldBindQuery 等) は未検証。(c) unknown field が「黙殺」されることの業務影響 (typo した有効 field の黙殺等) は本ケースの防御判定とは別。(d) handler への実配線は未実施 (owned_paths 外のため) |
| 11 | 修復案 | **適用済み (owned path 内)**: `BindStrictJSON(c, dst, maxBytes)` — `backend/internal/middleware/emr332_u03/strict_bind.go`。bounded read→`DisallowUnknownFields` decode→trailing 値拒否→`binding.Validator.ValidateStruct` で `binding` タグを再検証。error は `apperrors.WrapInvalidInput(httpapi.ParseBindError(err))` で既存応答形状と同一。**候補 (owned_paths 外のため未適用)**: 各 handler の `c.ShouldBindJSON`/`bindStaffJSON`/`bindAuthJSON` を `BindStrictJSON` へ置換、または `gin.EnableDecoderDisallowUnknownFields = true` のグローバル設定 + 既存 wrapper への trailing 検査維持。in-place な handler 変更は本ユニットでは行わない (scope 規約) |
| 12 | 再検証結果 | 実施済み — scoped isolated Docker `go test ./internal/middleware/emr332_u03/` 16 テスト全 PASS (出力末尾)。helper は unknown field / trailing value / oversized body / 宣言 Content-Length / required / oneof / range / 型違反を全て拒否。回帰: `go test ./internal/middleware/` PASS (sanitize_null_bytes 既存テスト含む) |

## 補足: 棚卸し結果

### bind 経路 (非テストファイル)

| 経路 | 箇所数 | DisallowUnknownFields |
|------|--------|------------------------|
| `c.ShouldBindJSON` 直接 | 209 (うち 3 件は wrapper 内部: staff/http_binding.go:48, auth/http_binding.go:57, middleware コメント言及 1) | なし — Gin decoder は常に非 strict |
| `bindStaffJSON` (staff) | 16 call sites (occupation/shift_template/staff/shift handlers) | なし (http_binding.go:25-52 は上限+trailing のみ) |
| `bindAuthJSON` (auth) | 8 call sites | なし (http_binding.go:25-61 同様) |
| 手動 decoder + DisallowUnknownFields | 9 sites (internal/ 8 + cmd/ 1) | **あり** |

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

## 実測コマンド出力 (verbatim)

```text
$ docker run --rm --pull never --network none \
    -v <w03>/backend:/app:ro \
    -v ekarte-go-mod-cache:/go/pkg/mod \
    -v ekarte-go-build-cache:/root/.cache/go-build \
    -w /app golang:1.26-alpine \
    go test -count=1 -v ./internal/middleware/emr332_u03/
=== RUN   TestC05ShouldBindJSON_TagEnforcement
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
--- PASS: TestBindStrictJSON_EnforcesBindingTags (8 subtests) all PASS
=== RUN   TestBindStrictJSON_RejectsMalformedAndEmpty
--- PASS: TestBindStrictJSON_RejectsMalformedAndEmpty (0.00s)
PASS
ok  	github.com/animal-ekarte/backend/internal/middleware/emr332_u03	0.019s

$ docker run --rm --pull never --network none -v <w03>/backend:/app:ro \
    -v ekarte-go-mod-cache:/go/pkg/mod -v ekarte-go-build-cache:/root/.cache/go-build \
    -w /app golang:1.26-alpine go test -count=1 ./internal/middleware/
ok  	github.com/animal-ekarte/backend/internal/middleware	0.149s

$ docker run ... golang:1.26-alpine go vet ./internal/middleware/emr332_u03/
(exit 0, no output)
```

*Marker 文字列は `SECCHK20261009C05Rnn` 形の英数字 sentinel のみ。実行可能な攻撃構文・実ペイロード・資格情報・PII は本ファイルに含まない (A5)。*
