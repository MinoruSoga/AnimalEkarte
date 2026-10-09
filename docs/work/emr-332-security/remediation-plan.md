# U07 — Remediation Plan (改善計画 — AC4)

Campaign: emr-332-security-defense / Unit: U07-REVERIFY-CONSOLIDATE / Attempt: att-u07-1
Recorded by: receiver-emr332-u07-01 / Recorded at: 2026-10-09
Ticket revision (consolidated): EMR-332 @ 2026-10-08T17:54:31Z

AC4 適用規則: 本計画は (a) **証跡で確認されたギャップ** (防御判定 FAIL のケース) と (b) **候補として明示記録された案** のみを載せる。無条件の一括修復は行わない (ticket「改善項目は防御ギャップが証拠で確認された場合の修復候補。表の全修復案を無条件に実装しない」)。**本ユニット (U07) は docs-only consolidation であり、ここで新たな製品コード変更は一切適用していない。**

担当欄の記録規則: 担当者名の推測設定を禁止 (ticket 本文「未指定の担当・期限を推測で設定しない」) ため、全エントリ `担当: 未指定` とする。実担当は着手時に本票へ記録する運用 (`人間受入時確定`)。

---

## Section 1 — 証跡確認済みギャップ (defense verdict FAIL)

### GAP-1: C05 — unknown/disallowed field silent-accept on all JSON bind paths

| AC4 field | 記録 |
|-----------|------|
| ケース ID | C05 (防御判定 FAIL — `verdict-matrix.md` / `u03/c05-input-validation.md`) |
| コード/設定根拠 | `gin.EnableDecoderDisallowUnknownFields` 設定箇所 0 件 (rg 全走査)。JSON bind 入口 **226 経路** (direct `ShouldBindJSON` 202 + `bindStaffJSON` 16 + `bindAuthJSON` 8) のいずれも unknown field を黙殺受理。`internal/staff/http_binding.go:25-52`, `internal/auth/http_binding.go:25-61` は上限+trailing+byte-length のみで DisallowUnknownFields なし。trailing 第 2 JSON value も `ShouldBindJSON` 単独経路では未読黙殺 |
| 最小変更 (適用状況) | **helper 適用済み・handler 配線は未適用 (candidate)** — `backend/internal/middleware/emr332_u03/strict_bind.go` に `BindStrictJSON(c, dst, maxBytes)` を新設 (bounded read → `DisallowUnknownFields` decode → trailing 値拒否 → `binding.Validator` 再検証 → `ParseBindError` で既存応答形状)。**この helper は unwired scaffold であり、226 経路のいずれにも配線されていない — よって C05 は FAIL のまま**。配線候補 (owned_paths 外のため U03/U07 では非適用): (a) 各 handler の `ShouldBindJSON`/`bindStaffJSON`/`bindAuthJSON` を `BindStrictJSON` へ置換 (`validateStaffJSONByteLengths`/`validateAuthJSONByteLengths` 相当の per-type byte-length 検査の移植を伴う — helper doc comment に明記済み); (b) `gin.EnableDecoderDisallowUnknownFields = true` グローバル設定 + 既存 wrapper の trailing 検査維持 |
| 担当 | 未指定 |
| 優先順位 | P0 (ticket 改善プラン工程2 の対象ケース C01–C05/C13/C17 に属する) |
| 再検証条件 | 同一ケース ID `C05` で: (1) 配線後の代表 bind 経路で unknown-field JSON が 400 拒否されることを simulated で確認; (2) `go test ./internal/middleware/emr332_u03/` 全 PASS の継続 + 配線済み handler の scoped test; (3) trailing 第 2 JSON value 拒否の確認; (4) byte-length 検査の移植確認 (bindStaffJSON/bindAuthJSON 移行時); (5) 副作用呼出 0 の再確認 |
| 証跡参照 | `docs/work/emr-332-security/u03/c05-input-validation.md`; `backend/internal/middleware/emr332_u03/strict_bind.go`; `backend/internal/middleware/emr332_u03/bind_boundary_test.go` (6 tests); `backend/internal/middleware/emr332_u03/strict_bind_test.go` (7 tests) |

### GAP-2: C09 — env-driven egress providers lack destination guards (scoped FAIL)

| AC4 field | 記録 |
|-----------|------|
| ケース ID | C09 (防御判定 `FAIL (scoped)` — `u05/c09-ssrf-egress.md`) |
| コード/設定根拠 | LSTEP/LINE 定数宛先経路は allowlist+hardened dial+redirect 抑止で PASS (static)。**env 駆動プロバイダ行は host allowlist・dial 層 private-IP guard・redirect 抑止を欠く**: `support/discord.go:44-50,126-132`, `support/plane.go:91-113,189-201`, `support/llm.go:79-141`, `medicalrecord/lab_device_mapping_llm.go:76-90,132-139` (release HTTPS 検証すら未経由), `infra/smtp/sender.go:62-69,211-245` (private-IP dial guard なし)。宛先は operator-config 境界であり request-driven SSRF ではない点は記録済み |
| 最小変更 (適用状況) | **candidate (未適用)** — U05 は audit_review 単位で修復しない。候補 R1–R4: R1 `middleware/liff_auth.go:163` に `CheckRedirect: ErrUseLastResponse` (兄弟 `line_link_service.go:572` と整合); R2 env 駆動 HTTP client へ LSTEP `hardenedDialContext` 相当 (resolve→IP 検査→pinned dial) + redirect 抑止の共通 transport 抽出; R3 `OPENAI_BASE_URL` を `config.go` `validateOptionalReleaseHTTPSURL` 経由化; R4 `infra/smtp/sender.go:69` dial 前 resolved-IP private 判定 |
| 担当 | 未指定 |
| 優先順位 | P1 (ticket 工程3 C05–C13) — R3 は兄弟 env URL より弱い検証のため同 P1 内で先行候補 |
| 再検証条件 | 同一ケース ID `C09` で: 修復適用後に egress サイト表 (Section B) を再評価し、対象行が allowlist/dial-guard/redirect 抑止を持つことを static で確認; runtime outbound 実測は staging 安全条件の確定後に限定実施 |
| 証跡参照 | `docs/work/emr-332-security/u05/c09-ssrf-egress.md` (Section B per-site verdict 表 + Section C R1–R4) |

### GAP-3: C10 — shared-file upload lacks content sniffing (scoped FAIL)

| AC4 field | 記録 |
|-----------|------|
| ケース ID | C10 (防御判定 `FAIL (scoped)` — `u05/c10-upload-validation.md`) |
| コード/設定根拠 | shared-file family は ext allowlist (`model.AllowedFileExtensions` {.pdf,.jpg,.jpeg,.png}) のみで content sniffing 不在 (`lstep/shared_file_request.go:28-41`, `shared_file_handler.go:39-83`)。残り 5 family は PASS。「拡張子だけ一致・内容が別物」のファイルが LINE Messaging 経由の外部送信経路に至る (`shared_file.go:42-44`) |
| 最小変更 (適用状況) | **candidate RC-1 (未適用)** — `newSharedFileUploadMeta` 後に `imageupload.VerifySniffedContent` 相当を追加し、`ext` 由来 MIME (`AllowedFileExtensions[ext]`) と先頭 512B sniff を照合する候補。`lstep/` は U05 owned_paths 外のため非適用 |
| 担当 | 未指定 |
| 優先順位 | P1 (ticket 工程3 C05–C13) |
| 再検証条件 | 同一ケース ID `C10` で: shared-file に対し (a) HTML-as-PNG 等の内容不一致 fixture が拒否されること (simulated)、(b) 拒否時 `sharedFile.Upload` 呼出 0、(c) 許可型は受諾、を確認 |
| 証跡参照 | `docs/work/emr-332-security/u05/c10-upload-validation.md` (Section A 行列 family 4 行 + Section B RC-1); `backend/internal/imageupload/emr332_u05/sniff_contract_test.go` |

---

## Section 2 — 候補 (candidate-only; 証跡上 FAIL ではない観察・改善案 — AC4 により候補に留める)

| ID | ケース | 候補内容 | 根拠 (code/config) | 担当 | 優先順位 | 再検証条件 | 証跡参照 |
|----|--------|----------|---------------------|------|----------|------------|----------|
| CAND-C14-RC1 | C14 | first-party actions の tag pin → SHA pin 化 (`actions/*`, `pnpm/action-setup`, `gitleaks-action`) | `.github/workflows/*` tag pin 一覧 (u05 Section B) | 未指定 | P2 (ticket 工程5) | C14 で pin 方式の static 再確認 | `u05/c14-supply-chain.md` Section D |
| CAND-C14-RC2 | C14 | `permissions:` 未宣言 workflow (`e2e.yml`, `performance-tests.yml`, `stg-migrate-direct.yml`, `stg-smoke.yml`, `actionlint.yml`) への最小権限明示 | 同上 | 未指定 | P2 | C14 で宣言有無の static 再確認 | `u05/c14-supply-chain.md` Section D |
| CAND-C14-RC3 | C14 | `go vuln`/`pnpm audit`/Dependabot の定期実行組込み (脆弱性状態 NOT ASSESSED の解消) | u05 Section C | 未指定 | P2 | scanner 実行証跡を残した上で C14 vulnerability 欄を評価 | `u05/c14-supply-chain.md` Section C/D |
| CAND-C14-RC4 | C14 | `performance-tests.yml` の `github-script` comment step — dead config の削除または trigger 追加 | 同上 | 未指定 | P2 | C14 static 再確認 | `u05/c14-supply-chain.md` Section D |
| CAND-U04-1 | C11/C12 | edge 即応答 preflight (`backend/worker/index.ts` `buildCorsPreflightResponse` :404-427) への最小 security header 付与 (例 `X-Content-Type-Options: nosniff`) — 204 no-body で実害リスク低 | u04 c11/c12 修復案欄の観察所見 (backend/worker/ は read-only 層) | 未指定 | 未割当 (観察候補 — 人間受入時確定) | C11/C12 で edge preflight 応答ヘッダを再評価 | `u04/c11-transport-headers.md` 修復案欄; `u04/c12-cors.md` 修復案欄 |
| CAND-U04-2 | C12 | edge/Gin preflight 複製実装の値 drift を防ぐ機械的 contract テスト | worker/index.ts:376-427 と cors.go:12-46 の手動同期契約 | 未指定 | 未割当 (同上) | C12 で両実装の値一致を機械ゲート化後に再確認 | `u04/c12-cors.md` 修復案欄 |
| CAND-U04-3 | C13 | `frontend/public/_headers` `/*` への明示的 `Cache-Control` (例 `no-cache`) — SPA HTML は platform 既定依存 | u04 c13 修復案欄 (read-only 層) | 未指定 | 未割当 (同上) | C13 で frontend HTML cache 挙動を再評価 | `u04/c13-cache-minimization.md` 修復案欄 |
| CAND-C06-1 | C06 | `Raw(fmt.Sprintf(...))` 禁止 lint rule / `textsearch.EscapeLike` 使用強制 (guardrail) | u03 c06 修復案欄 (unsafe site 0 件のため FAIL ではない) | 未指定 | 未割当 (同上) | C06 static 棚卸しで新規 unsafe site 0 を再確認 | `u03/c06-sql-inventory.md` field 11 |
| CAND-C07-1 | C07 | server-supplied URL 3 系 (`img.src`/`plane_issue_url`/`tel:`) の client 側 scheme allowlist (defense-in-depth) | u03 c07 修復案欄 (sink 0 件のため FAIL ではない) | 未指定 | 未割当 (同上) | C07 render leg BLOCKED 解除後に併せて評価 | `u03/c07-xss-render.md` field 11 |
| CAND-C18-1 | C18 | イベント名 `auth_staff_validation_fail_open` / メッセージ "auth staff validation failed open" の rename (実際は fail-closed で 503 拒否 — 名前が逆の誤ラベル) | `backend/cmd/api/composition_auth_routes.go:82-83` (u06 c18 観察欄; owned path 外のため観察として記録) | 未指定 | 低 (ログクエリ互換の確認が前提) | rename 適用後に C18 logging 検証を再実行し marker 相関を再確認 | `u06/c18-logging-audit.md` fields 10–11 |
| CAND-C16-1 | C16 | 本番複数インスタンス環境での集約 store (Redis 等) 検討 — process-local `RateLimitStore` の集約制限は UNKNOWN | `rate_limit.go:29-33` (in-memory map); deploy 実態未確認のため証拠不足で候補止まり | 未指定 | 未割当 (deploy topology 確定が前提) | multi-instance deployment topology 確定後に C16 で再評価 | `u06/c16-rate-limit.md` field 11 |

## 適用済み変更の棚卸し (本キャンペーン全体 — ZERO product fixes)

- **U01–U06 で適用された product-code 修復: 0 件**。追加されたコード物は `backend/internal/middleware/emr332_u03/strict_bind.go` (+ 同 package の `bind_boundary_test.go`/`strict_bind_test.go`) と各ユニットの `emr332_uNN` scoped test files のみ — いずれも **handler/本番経路へ配線されていない検証用 artifact** であり、GAP-1 の修復完了を意味しない。
- よって未修復 FAIL は **C05, C09 (scoped), C10 (scoped)** の 3 件が未完了のまま残る (AC5)。
