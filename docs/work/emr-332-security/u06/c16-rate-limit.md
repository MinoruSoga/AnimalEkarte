# U06 — C16 判定レコード: レート制限 / タイムアウト / サイズ上限 (rate limit, timeout, size caps)

> 配置追記（2026-10-09）: 現行テストは `backend/tests/security/`。旧package名を含む実行コマンド・出力は当時の履歴として保持。移動対応と現行検証は [配置整理記録](../test-layout-20261009.md) を参照。

Campaign: emr-332-security-defense / Unit: U06-DETECT-BUSINESS-AVAIL / Attempt: att-u06-1 / Receiver: receiver-emr332-u06-01
Base doc: docs/work/emr-332-security/u01-case-verdict-template.md (12-field fixed format)

## 12 fields

1. **ケース ID**: `C16`

2. **対象リビジョン**: worktree `AnimalEkarte-emr-332-security-defense-w06` branch `feat/emr-332-security-defense-w06`, HEAD `fc6acc6b8d267eb51050a4848069e1a1b569e7df` + 本ユニットで追加した test-only ファイル (production code 差分なし)。

3. **実施方式 (method)**:
   - `simulated` — `tests/security/middleware` の httptest/in-memory test (ネットワーク listen なし・負荷なし・逐次 request)。rps は `rate.Every(time.Hour)` に固定し refill が test 中に発火しないよう排除 → 429 は全て deterministic。
   - `static` — `cmd/api/main.go` http.Server timeout / `cmd/api/composition_runtime.go` middleware wiring / `sanitize_null_bytes.go` body ceiling のレビュー。

4. **正常条件**:
   - 期待: burst 内の request は通過し handler に到達する。RateLimit bucket は middleware インスタンス (bucketID) × client IP 単位で独立。FailureRateLimit は <400 応答で予算を消費しない。FailureLimiter の `Admit` は非消費ゲート。宣言 Content-Length が上限以下・write method 以外の request は通常処理される。
   - 観測 (simulated): `TestC16_RateLimitRejects429WithRetryAfter` 1 件目 200 / `TestC16_RateLimitBucketsAreIndependent` 別 bucket・別 IP は通過 / `TestC16_FailureRateLimitConsumesOnlyFailures` 成功 3 件で予算温存 / `TestC16_FailureLimiterAdmitAndRecordFailure` Admit は非消費・key 空間独立 / `TestC16_LiffRateLimitPerMinuteBurst` 1/min→burst1 で初回通過 / `TestC16_LimitRequestBodyScopeIsWriteMethods` GET は上限対象外 (documented scope)。
   - 観測 (static): `http.Server` は `ReadTimeout 120s / WriteTimeout 120s / ReadHeaderTimeout 10s / IdleTimeout 60s / MaxHeaderBytes 32KiB` (cmd/api/main.go:219-223)。protected group 全体に `LimitRequestBody(1MiB)` (composition_runtime.go:533)、LIFF に `LiffBodyLimit` (composition_runtime.go:687)。`SanitizeNullBytes` は除去前の raw byte を `MaxBytesReader` で計上し sanitize 回避不可 (sanitize_null_bytes.go:82-104)。

5. **拒否条件**:
   - 期待: burst 超過 → HTTP 429 + `Retry-After` (正整数秒) + `{"code":429,"message":"rate limit exceeded","timestamp":…}` で abort (respondError → AbortWithStatusJSON、response.go:16-25 → downstream handler 非到達)。失敗専用 bucket 枯渇 → 同形 429。宣言 Content-Length > 上限 → 413 + abort。非宣言 (chunked) は読取時に MaxBytesReader が打切り。
   - 観測 (simulated): `TestC16_RateLimitRejects429WithRetryAfter` — 2 件目 429、`Retry-After` は整数 ≥1、body に "rate limit exceeded"、handler 呼出は 1 のまま (abort 確認)。`TestC16_FailureRateLimitConsumesOnlyFailures` — 失敗 2 件で burst=2 枯渇、3 件目は成功予定でも handler 未到達の 429 (handler 計数 5 のまま)。`TestC16_FailureLimiterAdmitAndRecordFailure` — RecordFailure 後 `Admit` が (≥1s, false)。`TestC16_LiffRateLimitPerMinuteBurst` — 2 件目 429 + Retry-After 存在。`TestC16_LimitRequestBodyRejectsOversizedDeclaredLength` — 100B > 64B で 413、handler 0 回。`TestC16_LimitRequestBodyCapsStreamingRead` — ContentLength=-1 の 100B は handler 内 `io.ReadAll` が error を返し 64B までしか読めない。

6. **防御判定 (defense verdict)**: `PASS` — 単一プロセス内の bucket 独立・429+Retry-After・失敗のみ消費・413/読取り打切りは simulated で成立。timeout/header 上限は http.Server 設定として static 確認。**範囲は process-local** (下記 10 を参照)。

7. **副作用呼出回数 (side-effect call count)**: 拒否 request の downstream handler 呼出 = **0** (全 429/413 ケースでカウンタ不変を assert)。拒否時の副作用は httptest.ResponseRecorder への 429/413 応答記録のみ — 外部送信・永続化・通知・監査 write なし。

8. **検出判定 (detection verdict)**: `PASS` (限定的) — 429 拒否は RequestLoggingMiddleware を通すと marker request_id 付き WARN `client error` として in-memory slog に記録されることを `TestC18_MarkerTaggedRefusalEventCaptured` (logging_requestid_test.go) で確認。c16 自体の limiter は専用イベントを出力しない設計 (429 応答のみ) — これは成立している観測経路として記録。

9. **安全な証跡参照 (safe evidence ref)**:
   - Tests: `backend/tests/security/middleware/rate_limit_test.go` (TestC16_* 8 tests)。
   - Static anchors: `internal/middleware/rate_limit.go:23-32 (bucketID×IP key), :84-100 (getLimiter), :105-122 (Retry-After+429), :128-141 (RateLimit), :149-165 (FailureRateLimit), :170-202 (FailureLimiter), :206-208 (LiffRateLimit)`, `internal/middleware/response.go:16-25 (AbortWithStatusJSON)`, `internal/middleware/sanitize_null_bytes.go:20,82-104,110-141`, `cmd/api/main.go:19-27,219-223`, `cmd/api/composition_runtime.go:533,573,607,674,682-687`, `cmd/api/composition_auth_routes.go:44-72`。
   - 実行出力 (verbatim、isolated docker、network none):
     ```text
     ok  	github.com/animal-ekarte/backend/internal/reservation/emr332_u06	0.003s
     ok  	github.com/animal-ekarte/backend/internal/medicalrecord/emr332_u06	0.003s
     ok  	github.com/animal-ekarte/backend/internal/middleware/emr332_u06	0.003s
     ok  	github.com/animal-ekarte/backend/internal/audit/emr332_u06	0.002s
     ```

10. **制約・未検証範囲 (constraints / unverified scope)**:
    - **bucket granularity (訂正防止のため明記)**: limiter は「単一グローバル bucket」ではなく store 内 `(bucketID, IP)` キー — middleware インスタンスごとに独立。`NewRateLimitStore` は少なくとも 4 箇所で別々に構築 (composition_auth_routes.go:72 auth 用 / composition_runtime.go:573 LINE webhook 用 / :607 support chat 用 / :674 LIFF 用) → **route-group/store ごと + endpoint (middleware instance) ごと + IP ごと**。
    - **process-local**: `RateLimitStore` はインメモリ map (rate_limit.go:29-33) のため、複数プロセス/複数インスタンス間の集約制限は**未検証 (`UNKNOWN`)** — deploy 形態 (Worker→Container) に依存し、本試験の範囲外。
    - cleanup loop (5 分 interval / 10 分 TTL、rate_limit.go:45-80) は test 時間窓では未実行。
    - `FailureRateLimit` の over-admit (空 bucket 時の同時到着で最大 burst 分) は doc-comment どおり許容設計 (rate_limit.go:143-148) — 単一スレッド test では観測しない。
    - http.Server timeout 値は設定の存在確認のみ (実挙動は socket 挙動依存 → 未検証)。
    - 負荷試験・閾値プロービング・並行送信は安全境界により未実施。

11. **修復案 (remediation)**: owned paths 内で確認された gap なし → fix 未適用。候補 (範囲外・証拠不足のため非適用): 本番が複数インスタンスなら集約 store (Redis 等) の検討 — deploy 実態未確認であり本判定では主張しない。

12. **再検証結果 (re-verification)**: 未実施 — fix なし。初回 scoped run で全 test PASS (上記 verbatim 出力)。
