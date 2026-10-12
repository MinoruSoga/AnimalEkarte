# U04 — Case Verdict Record: C13 (cache prevention + response data minimization)

> 配置追記（2026-10-09）: 現行テストは `backend/tests/security/`。旧package名を含む実行コマンド・出力は当時の履歴として保持。移動対応と現行検証は [配置整理記録](../test-layout-20261009.md) を参照。

Campaign: emr-332-security-defense / Unit: U04-BROWSER-EDGE / Attempt: att-u04-1
Template: ../baseline/case-verdict-template.md (12 fields)

| # | Field | 記録 |
|---|-------|------|
| 1 | ケース ID | `C13` — API 応答の no-store キャッシュ防止 + 応答データ最小化 |
| 2 | 対象リビジョン | 静的・模擬対象: base `de28e9765a29091339a872a85f1875a947e07e5b` + w04 追加テスト。**staging 稼働リビジョン: undetermined** (照合試行記録は c08-csrf.md 補足)。staging-observed は BLOCKED。 |
| 3 | 実施方式 (method) | `simulated` (実 `SecurityHeaders` middleware 経由 httptest で cache 系ヘッダ実測) + `static` (global chain・`_headers`・応答 DTO レビュー)。staging-observed 行は BLOCKED。 |
| 4 | 正常条件 | 期待: 全 Gin 到達応答に `Cache-Control: no-store, no-cache, must-revalidate` + `Pragma: no-cache`。観測 (simulated): 200 応答・CSRF 403 拒否応答の両方で両ヘッダを確認 — global chain (main.go:203-208) で SecurityHeaders が早期に set するため abort 後も残存。frontend `/assets/*` は content-hash 付き immutable `public, max-age=31536000` (`_headers`) — ハッシュ済み資産の長期 cache は妥当設計。 |
| 5 | 拒否条件 | 本ケースの拒否形状は「API 応答が cacheable なまま送出されること」— 観測: Gin 応答面で cacheable directive を持つ経路は静的レビューで確認できず (global Use のため全経路一括適用)。`frontend/public/_headers` の `/*` ブロックは Cache-Control を指定しない → SPA HTML は配信プラットフォーム既定 (no-store でない点は後述の候補欄に記録)。 |
| 6 | 防御判定 | **PASS** (static + simulated 範囲: API 面の cache 防止は global chain で実測確認)。staging-observed 判定は BLOCKED (送信 0 回)。 |
| 7 | 副作用呼出回数 | 0 — ヘッダ・応答形状の確認のみ。模擬内の downstream 呼出は制御群計測のみ (永続化なし)。 |
| 8 | 検出判定 | **UNKNOWN** — cacheable 応答の検出・ログ機構は存在しない (marker-capture 証跡なし)。 |
| 9 | 安全な証跡参照 | `backend/internal/middleware/security_headers.go:16-18` (BUG-137: no-store/no-cache/must-revalidate + Pragma); `backend/cmd/api/main.go:203-208` (global chain); `backend/tests/security/browser/browser_edge_test.go` `TestEMR332C11C13_SecurityHeaderChain`; `docs/work/emr-332-security/browser/evidence/middleware-tests.txt`; `frontend/public/_headers` (read-only — `/*` CSP/HSTS等 + `/assets/*` immutable); 最小化スポットチェック: `backend/cmd/api/base_routes.go:18-46` (`/health`・`/health/db` は `{"status": ...}` のみ・「DB internals are not disclosed」意図コメント), `backend/internal/reservation/liff_response.go:12-39` (`liffSettingsResponse` — 「機密フィールド除外」の専用 DTO), `:247-259` (`liffReservationCreatedResponse` — 201 は id/notes の 2 フィールドのみ), `:285-301` (健康手帳の日付は date-only 露出), `backend/worker/index.ts:462-465` (edge 503 fallback は `{"error":"service_unavailable"}` のみ — 例外本文を返さない設計)。 |
| 10 | 制約・未検証範囲 | 最小化レビューは公開面・エラー面の代表サンプル (`/health*`・LIFF settings・edge 503 fallback・LIFF 予約作成応答) を対象; 認証必須の臨床 API 応答全体の網羅監査は範囲外 (field-level の過剰露出は各 domain の別途レビュー対象)。cross-user cache 実験は禁止条件により未実施。staging 実測の Cache-Control/Pragma 観測は BLOCKED。**未検証**: frontend HTML の配信 cache 挙動実測 (platform 既定依存)、CDN/ブラウザ cache の実際の保持状態。 |
| 11 | 修復案 | **適用なし (owned_paths 内 gap なし)**。候補として記録のみ: `frontend/public/_headers` (read-only) の `/*` に明示的 `Cache-Control` (例 `no-cache`) が無く SPA HTML はプラットフォーム既定に依存する — 認証後コンテンツを HTML に含めない SPA 構造のため実害リスクは低いが、明示的 directive の追加は remediation candidate として記録 (read-only 層のため本ユニットでは非適用)。 |
| 12 | 再検証結果 | 修復なしのため再検証不要。初回検証 PASS: `TestEMR332C11C13_SecurityHeaderChain` で 200/403 両応答の `Cache-Control: no-store, no-cache, must-revalidate` + `Pragma: no-cache` を実測 (evidence-t01 verbatim)。 |
