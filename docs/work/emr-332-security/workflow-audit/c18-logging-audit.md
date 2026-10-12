# U06 — C18 判定レコード: ログ / 監査カバレッジ + シークレット・PII 除外 (logging & audit)

> 配置追記（2026-10-09）: 現行テストは `backend/tests/security/`。旧package名を含む実行コマンド・出力は当時の履歴として保持。移動対応と現行検証は [配置整理記録](../test-layout-20261009.md) を参照。

Campaign: emr-332-security-defense / Unit: U06-DETECT-BUSINESS-AVAIL / Attempt: att-u06-1 / Receiver: receiver-emr332-u06-01
Base doc: docs/work/emr-332-security/baseline/case-verdict-template.md (12-field fixed format)

## 12 fields

1. **ケース ID**: `C18`

2. **対象リビジョン**: worktree `AnimalEkarte-emr-332-security-defense-w06` branch `feat/emr-332-security-defense-w06`, HEAD `fc6acc6b8d267eb51050a4848069e1a1b569e7df` + 本ユニットで追加した test-only ファイル (production code 差分なし)。

3. **実施方式 (method)**:
   - `simulated` — `tests/security/middleware/logging_requestid_test.go` (slog JSON → bytes.Buffer、httptest、SECCHK20261009* marker) と `tests/security/audit/audit_validate_test.go` (counting fake repo — audit row は永続化されない)。全てメモリ内・単一スレッド・決定的。
   - `static` — logging.go / auth.go / audit service+repository / composition_auth_routes.go / appointment_notification_service.go のレビュー。

4. **正常条件**:
   - 期待: request log は `{method, path, status, latency, request_id, client_ip}` (+`error` は c.Errors 存在時のみ) の構造化レコード。`X-Request-ID` は `[a-zA-Z0-9_-]{1,64}` のみ受理し、他は 8hex 再生成。audit は `ValidateLog` が clinic_id/actor_type/action/resource + actor 整合を強制し、service が正規化 (UA ≤1024B・UTF-8、IP 空→NULL) して repo.Create/CreateTx へ渡す。`CreateTx` は ambient tx 欠落で fail-closed。
   - 観測 (simulated): `TestC18_RequestIDSanitization` (10 subcases: marker/hyphen/underscore echo、空・空白・`;`・`"`・`/`・65chars 再生成、64chars 受理、header↔context 一致)。`TestC18_RequestLoggingFieldsAndSecretExclusion` (2xx→INFO / 4xx→WARN / 5xx→ERROR+error attr、フィールド allowlist 一致、path は query を含まない)。`TestC18_ValidateLogInvariantMatrix` (13 subcases)。`TestC18_LogEntryTxUsesAmbientTxChannel` — tx 経路は `repo.CreateTx` へのみ記録。`TestC18_UserAgentBoundedAt1024Bytes` (400 全角文字=1200B → rune 境界 ≤1024B かつ valid UTF-8)。`TestC18_IPAddressNormalization`。`TestC18_LogMedicalRecordChangeMarshalsSyntheticValues` — nil actorID → system、old/new は JSON marshal。
   - 観測 (static): logging.go:41-48 (field allowlist), :74-89 (request_id サニタイズ)。audit/service.go:67-96 (ValidateLog), :105-115, 131-132 (正規化), :144-185 (Log/LogEntry/LogEntryTx)。audit/repository.go:43-52 (CreateTx は TxFromContext nil → InternalServerError)。

5. **拒否条件**:
   - 期待: 不正 request_id は無言ではなく再生成で「受理」、audit 不変条件違反は InvalidInput で persistence 未到達、ambient tx なし CreateTx は INTERNAL。
   - 観測 (simulated): `TestC18_RequestIDSanitization` — 制御文字/区切り混入・超長は 8hex 再生成。`TestC18_RefusedAuditWriteHasZeroPersistenceCalls` — ValidateLog 拒否で `repo.created`/`createdTx` 0。`TestC18_RealRepositoryCreateTxFailsClosedWithoutAmbientTx` — 実 repository (`&gorm.DB{}`、gorm 未到達) が INTERNAL を返す。`TestC18_RepositoryFailurePropagates` — repo 失敗が呼出し側へ伝搬し `audit_write_failed` ERROR を slog に記録。
   - 観測 (static): auth staff 検証失敗は 503 `access validation unavailable` で fail-closed (auth.go:76,127,206,224,247)。

6. **防御判定 (defense verdict)**: `PASS` — 記録した範囲で request ログの allowlist 遵守・marker 以外の sentinel 非含有・audit validation/fail-closed が成立。

7. **副作用呼出回数 (side-effect call count)**: 拒否 audit entry → persistence call **0** (countingRepo 実測)。受理 audit → Create=1 / CreateTx=1 のみ (fake 内での記録、DB row 0 件)。429 拒否 request の logging 副作用 = in-memory slog record 1 件のみ (永続ログなし)。

8. **検出判定 (detection verdict)**: `PASS` (限定的) — 拒否イベントは marker request_id 付きで request log に残ることを `TestC18_MarkerTaggedRefusalEventCaptured` で確認 (WARN/client error/status=429/request_id=SECCHK20261009REFUSAL)。audit write 失敗は `audit_write_failed` として観測可能 (service.go:150-156、test で ERROR 行を確認)。**通知経路の到達性は `UNKNOWN`** — `ReservationNotifier` 実装の配送確認証跡は本レビュー範囲で見つからず (appointment_notification_service.go:20,96,147 は send 呼出のみ、配信成否の既存証跡なし)。本事実を推測で PASS にしない。

9. **安全な証跡参照 (safe evidence ref)**:
   - Tests: `backend/tests/security/middleware/logging_requestid_test.go` (TestC18_* 4 tests / 13+ subcases), `backend/tests/security/audit/audit_validate_test.go` (TestC18_* 8 tests / 20 subcases)。
   - Static anchors: `internal/middleware/logging.go:13-99`, `internal/middleware/rate_limit.go:118-122`, `internal/middleware/auth.go:26,47,76,97,127,206,224,247`, `internal/audit/service.go:14,17-53,67-96,105-115,117-134,143-201`, `internal/audit/repository.go:18-52`, `cmd/api/composition_auth_routes.go:82-83`, `internal/reservation/appointment_notification_service.go:18-21,96-147`。
   - 実行出力 (verbatim、isolated docker、network none):
     ```text
     ok  	github.com/animal-ekarte/backend/internal/reservation/emr332_u06	0.003s
     ok  	github.com/animal-ekarte/backend/internal/medicalrecord/emr332_u06	0.003s
     ok  	github.com/animal-ekarte/backend/internal/middleware/emr332_u06	0.003s
     ok  	github.com/animal-ekarte/backend/internal/audit/emr332_u06	0.002s
     ```
     内、C18 関連行: middleware emr332_u06 / audit emr332_u06 の両 `ok`。`TestC18_RepositoryFailurePropagates` の verbose 出力に `ERROR audit_write_failed action=SECCHK20261009ACTION ...` を確認 (故意の fake 失敗による期待ログ行)。

10. **制約・未検証範囲 (constraints / unverified scope)**:
    - slog 出力の**永続先** (stdout → 基盤収集) は範囲外。`internal/logger` の初期化経路は `main.go` で利用されるが、本 test は slog.Default 差替えによる in-memory 検証のみ。
    - `internal/audit` の DB-dependent test (`repository_db_test.go`, `repository_real_ddl_test.go`) は Compose Postgres 前提のため本制約下では**未実行** — service 層のみ simulated 検証。
    - secret/PII 除外は request-log field contract に対してのみ検証。他 package の任意 `slog` 呼出 (handler/service 内の個別ログ) は全数網羅していない → 残件 `UNKNOWN`。
    - **通知到達性 `UNKNOWN`** — Notification 実送経路 (外部 push/LINE 送信) の配送証跡・既存観測証跡は確認できず。
    - **観察 (fix 未適用・owned path 外)**: `composition_auth_routes.go:82-83` のイベント名 `auth_staff_validation_fail_open` / メッセージ "auth staff validation failed open" は、実際は fail-closed (auth.go で 503 拒否) と**名前が逆**。ログ解析者が誤読し得る誤ラベル — owned path 外のため本 attempt では**修正せず観察として記録**。

11. **修復案 (remediation)**: owned paths 内の確定 gap なし → fix 未適用。候補 (owned path 外・非適用): `auth_staff_validation_fail_open` のイベント名を `auth_staff_validation_fail_closed` 等へ rename (composition_auth_routes.go:83) — 優先度低、ログクエリ互換の確認が前提。

12. **再検証結果 (re-verification)**: 未実施 — fix なし。初回 scoped run で全 test PASS (上記 verbatim 出力)。
