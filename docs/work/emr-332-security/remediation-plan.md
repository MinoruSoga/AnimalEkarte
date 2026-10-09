# U07 — Remediation Plan (改善計画 — AC4)

> 配置追記（2026-10-09）: 現行テストは `backend/tests/security/`。旧package名を含む実行コマンド・出力は当時の履歴として保持。移動対応と現行検証は [配置整理記録](test-layout-20261009.md) を参照。

Campaign: emr-332-security-defense / Unit: U07-REVERIFY-CONSOLIDATE / Attempt: att-u07-1
Recorded by: receiver-emr332-u07-01 / Recorded at: 2026-10-09
Ticket revision (consolidated): EMR-332 @ 2026-10-08T17:54:31Z

AC4 適用規則: 本計画は (a) **証跡で確認されたギャップ** (防御判定 FAIL のケース) と (b) **候補として明示記録された案** のみを載せる。無条件の一括修復は行わない (ticket「改善項目は防御ギャップが証拠で確認された場合の修復候補。表の全修復案を無条件に実装しない」)。**本ユニット (U07) は docs-only consolidation であり、ここで新たな製品コード変更は一切適用していない。**

担当欄の記録規則: 担当者名の推測設定を禁止 (ticket 本文「未指定の担当・期限を推測で設定しない」) ため、全エントリ `担当: 未指定` とする。実担当は着手時に本票へ記録する運用 (`人間受入時確定`)。

---

## Section 1 — 要件差分・強化候補・適用修復

### GAP-1: C05 — 未知キー拒否要件との差分 (影響未評価)

| AC4 field | 記録 |
|-----------|------|
| ケース ID | C05: FAIL (試験要件との差分)、脆弱性重大度は未評価 |
| コード/設定根拠 | 代表DTOで未知キーが破棄される。226はbind call-site棚卸しで全endpointの実測・保護フィールド更新の証明ではない。`reservation/line_reservation_setting_request.go:38-41`は旧キー黙殺を意図する契約。 |
| 最小変更 (適用状況) | 製品経路の変更なし。未配線strict helperはartifactのまま。対象APIの受理/禁止項目、互換性、実害を確認しendpoint単位で決める。全226置換・グローバルstrict化は今回行わない。 |
| 担当 | 実装評価: Codex (引き継ぎ済み)。API契約・受入責任者: 未指定 |
| 優先順位 | 未評価 — 工程番号由来のP0を重大度として撤回。影響と互換契約の評価後に決定 |
| 再検証条件 | 合意した対象の実handlerで保護項目が更新されないこと、受理項目・byte-length・trailing契約を維持すること、拒否時後段呼出0を確認 |
| 証跡参照 | `input-output/c05-input-validation.md`; `repair-20261009.md` |

### GAP-2: C09 — operator-config境界の強化候補 (SSRF impact UNKNOWN)

| AC4 field | 記録 |
|-----------|------|
| ケース ID | C09: 全体UNKNOWN。LSTEP/probe既存guardはstatic PASS |
| コード/設定根拠 | Discord/Plane/LLM/S3/SMTPの宛先は運用設定、LINEは定数。通常callerの宛先制御・redirect/DNS越境は未証明。OPENAI_BASE_URLのrelease HTTPS検証不在は独立した設定境界候補。 |
| 最小変更 (適用状況) | 製品変更なし。R1–R4は許可宛先・設定権限・信頼境界を確認してからproviderごとに決める。正規private SMTP relayを壊す一律IP禁止は未採用。 |
| 担当 | 評価訂正: Codex。運用宛先policy責任者: 未指定 |
| 優先順位 | 未評価 — SSRF重大度は未証明。release HTTPS候補は資格情報送信policyと実enablementを確認して決定 |
| 再検証条件 | callerの権限と宛先ソース、redirect/DNS境界、正規relay契約を記録する。通信試験は指定staging/tenant等の安全前提成立後に限定 |
| 証跡参照 | `egress/c09-ssrf-egress.md`; `repair-20261009.md` |

### GAP-3: C10 — 共有ファイル保存前の内容形式照合 (適用済み)

| AC4 field | 記録 |
|-----------|------|
| ケース ID | C10: PASS (scoped static/simulated)、実保存・配送・stagingは未検証 |
| コード/設定根拠 | 修復前は`shared_file_request.go:28-41`の拡張子のみ。配送は別操作でowner/opt-out/credentials確認後に署名URLをLINEへ渡す (`line_send_service.go:139-158`, `shared_file_service.go:132-145`, `infra/line/push.go:36-67,94-120`)。uploadだけで自動送信しない。 |
| 最小変更 (適用状況) | RC-1適用: handlerで用途/フォーム検証後、Upload前に既存`imageupload.VerifySniffedContent(file, meta.contentType)`を配線。既存auth/size/MIME由来契約維持。 |
| 担当 | 修復・メモリ検証: Codex。人間受入責任者: 未指定 |
| 優先順位 | 内容検証欠落の最小修復を今回実施。悪用可能性/CVSSは未評価 |
| 再検証条件 | 実handlerのメモリ試験: 正常5ケースで201/Upload1/内容完全保持、内容不一致・空7ケースで400/Upload0/Locationなし。既存auth/bind/extension/service-errorを維持 |
| 証跡参照 | `uploads/c10-upload-validation.md`; `backend/internal/lstep/shared_file_content_test.go`; `repair-20261009.md` |

---

## Section 2 — 候補 (candidate-only; 証跡上 FAIL ではない観察・改善案 — AC4 により候補に留める)

| ID | ケース | 候補内容 | 根拠 (code/config) | 担当 | 優先順位 | 再検証条件 | 証跡参照 |
|----|--------|----------|---------------------|------|----------|------------|----------|
| CAND-C14-RC1 | C14 | first-party actions の tag pin → SHA pin 化 (`actions/*`, `pnpm/action-setup`, `gitleaks-action`) | `.github/workflows/*` tag pin 一覧 (u05 Section B) | 未指定 | P2 (ticket 工程5) | C14 で pin 方式の static 再確認 | `dependencies/c14-supply-chain.md` Section D |
| CAND-C14-RC2 | C14 | `permissions:` 未宣言 workflow (`e2e.yml`, `performance-tests.yml`, `stg-migrate-direct.yml`, `stg-smoke.yml`, `actionlint.yml`) への最小権限明示 | 同上 | 未指定 | P2 | C14 で宣言有無の static 再確認 | `dependencies/c14-supply-chain.md` Section D |
| CAND-C14-RC3 | C14 | `go vuln`/`pnpm audit`/Dependabot の定期実行組込み (脆弱性状態 NOT ASSESSED の解消) | u05 Section C | 未指定 | P2 | scanner 実行証跡を残した上で C14 vulnerability 欄を評価 | `dependencies/c14-supply-chain.md` Section C/D |
| CAND-C14-RC4 | C14 | `performance-tests.yml` の `github-script` comment step — dead config の削除または trigger 追加 | 同上 | 未指定 | P2 | C14 static 再確認 | `dependencies/c14-supply-chain.md` Section D |
| CAND-U04-1 | C11/C12 | edge 即応答 preflight (`backend/worker/index.ts` `buildCorsPreflightResponse` :404-427) への最小 security header 付与 (例 `X-Content-Type-Options: nosniff`) — 204 no-body で実害リスク低 | u04 c11/c12 修復案欄の観察所見 (backend/worker/ は read-only 層) | 未指定 | 未割当 (観察候補 — 人間受入時確定) | C11/C12 で edge preflight 応答ヘッダを再評価 | `browser/c11-transport-headers.md` 修復案欄; `browser/c12-cors.md` 修復案欄 |
| CAND-U04-2 | C12 | edge/Gin preflight 複製実装の値 drift を防ぐ機械的 contract テスト | worker/index.ts:376-427 と cors.go:12-46 の手動同期契約 | 未指定 | 未割当 (同上) | C12 で両実装の値一致を機械ゲート化後に再確認 | `browser/c12-cors.md` 修復案欄 |
| CAND-U04-3 | C13 | `frontend/public/_headers` `/*` への明示的 `Cache-Control` (例 `no-cache`) — SPA HTML は platform 既定依存 | u04 c13 修復案欄 (read-only 層) | 未指定 | 未割当 (同上) | C13 で frontend HTML cache 挙動を再評価 | `browser/c13-cache-minimization.md` 修復案欄 |
| CAND-C06-1 | C06 | `Raw(fmt.Sprintf(...))` 禁止 lint rule / `textsearch.EscapeLike` 使用強制 (guardrail) | u03 c06 修復案欄 (unsafe site 0 件のため FAIL ではない) | 未指定 | 未割当 (同上) | C06 static 棚卸しで新規 unsafe site 0 を再確認 | `input-output/c06-sql-inventory.md` field 11 |
| CAND-C07-1 | C07 | server-supplied URL 3 系 (`img.src`/`plane_issue_url`/`tel:`) の client 側 scheme allowlist (defense-in-depth) | u03 c07 修復案欄 (sink 0 件のため FAIL ではない) | 未指定 | 未割当 (同上) | C07 render leg BLOCKED 解除後に併せて評価 | `input-output/c07-xss-render.md` field 11 |
| CAND-C18-1 | C18 | イベント名 `auth_staff_validation_fail_open` / メッセージ "auth staff validation failed open" の rename (実際は fail-closed で 503 拒否 — 名前が逆の誤ラベル) | `backend/cmd/api/composition_auth_routes.go:82-83` (u06 c18 観察欄; owned path 外のため観察として記録) | 未指定 | 低 (ログクエリ互換の確認が前提) | rename 適用後に C18 logging 検証を再実行し marker 相関を再確認 | `workflow-audit/c18-logging-audit.md` fields 10–11 |
| CAND-C16-1 | C16 | 本番複数インスタンス環境での集約 store (Redis 等) 検討 — process-local `RateLimitStore` の集約制限は UNKNOWN | `rate_limit.go:29-33` (in-memory map); deploy 実態未確認のため証拠不足で候補止まり | 未指定 | 未割当 (deploy topology 確定が前提) | multi-instance deployment topology 確定後に C16 で再評価 | `workflow-audit/c16-rate-limit.md` field 11 |

## 適用済み変更の棚卸し (初回campaign履歴と後続修復)

- **初回U01–U06で適用された product-code 修復: 0 件 (履歴)**。追加されたコード物は `backend/internal/middleware/emr332_u03/strict_bind.go` (+ 同 package の `bind_boundary_test.go`/`strict_bind_test.go`) と各ユニットの `emr332_uNN` scoped test files のみ — いずれも **handler/本番経路へ配線されていない検証用 artifact** であり、GAP-1 の修復完了を意味しない。
- **後続Codex修復: C10の保存前sniff配線1件**。現在C05は要件差分FAIL、C09は影響/宛先policy UNKNOWN。C10は修復対象のstatic/simulated PASSであり、実保存・配送・staging未確認 (AC5)。
