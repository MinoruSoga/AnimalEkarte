# U05 — C09 SSRF / 外部送信先 egress 制御 判定レコード

Campaign: emr-332-security-defense / Unit: U05-FILES-SSRF-SUPPLY / Attempt: att-u05-1
Format: `docs/work/emr-332-security/u01-case-verdict-template.md` (12 fields)

## 判定レコード (12 fields)

| # | Field | 記録内容 |
|---|-------|----------|
| 1 | ケース ID | `C09` — SSRF・外部通信先制御 |
| 2 | 対象リビジョン | worktree `feat/emr-332-security-defense-w05` HEAD `fc6acc6b8d267eb51050a4848069e1a1b569e7df` (静的レビュー対象)。staging 稼働リビジョン照合は未実施 (u01-safety-record §2: BLOCKED) |
| 3 | 実施方式 | `static` — ソースコード静的分析のみ。外部呼出・DNS 解決・dial は一切実行していない |
| 4 | 正常条件 | 期待: allowlist/定数先 (api.lstep.jp / api.line.me) あるいは運用者明示設定先への HTTPS egress のみ成立。観測 (static): LSTEP 経路は URL 検証+hardened dial で実装を確認。環境変数駆動プロバイダ (Discord/Plane/LLM/S3/SMTP) は release モード検証ありだが dial 層ガードは LSTEP のみ |
| 5 | 拒否条件 | 期待: 非 allowlist ホスト・IP literal・userinfo・非 HTTPS・非 443 ポート・redirect・private/loopback/CGNAT 解決結果を拒否。観測 (static): LSTEP 経路では `ValidateLstepBaseURL` + `hardenedDialContext` + `denyRedirects` で全項目拒否を確認。probe 経路も同等。env 駆動プロバイダは HTTPS 強制 (release のみ) までで、host allowlist・dial 層 private-IP ガード・redirect 抑止は未実装 |
| 6 | 防御判定 | `FAIL` (scoped) — LSTEP/LINE 定数経路は PASS。ただし rubric "missing allowlist/redirect-revalidation = FAIL + candidate" に基づき、env 駆動プロバイダ行 (Discord/Plane/LLM/OpenAI-drafter/SMTP) を scoped FAIL + remediation candidate として記録 (Section B 表の per-site verdict 参照)。攻撃者影響度は request-driven ではなく operator-config 境界に限られる点は Remediation 欄に明記 |
| 7 | 副作用呼出回数 | 拒否分岐での outbound 呼出 = `0` (静的経路確認)。`ValidateLstepBaseURL` 失敗・`ensureWriteEnabled` 失敗・`hardenedDialContext` 拒否は全て `http.Client.Do` / dial 到達前に return (client.go:95-100, dial.go:62-66, credentials.go:27-58)。ランタイム計測は未実施 (static のみ) |
| 8 | 検出判定 | `UNKNOWN` — probe 拒否は `slog.WarnContext` で構造化ログ出力 (lstep_settings_connection.go:100,104,116) + `classifyConnectionProbeError` 安定コード (:166-192) を静的確認。ただしランタイム検出・通知経路の実測は staging-observed が BLOCKED のため未検証 |
| 9 | 安全な証跡参照 | 本ファイル Section A/B の file:line。具体 URL/トークン値は記録しない |
| 10 | 制約・未検証範囲 | 実 outbound 呼出なし・DNS 解決なし・redirect 追跡なし (実行しない設計)。`*_test.go` / `httptest` / mock (`NewInsecureTestClient` client.go:76-85 は `testing.Testing()` 外で panic する test-only 経路) は本判定から除外。staging 観測は未実施。Workers runtime の service-binding 宛先は wrangler.jsonc の静的値で推定するのみ |
| 11 | 修復案 | R1-R4 (Section C)。`audit_review` 単位のため本 unit では修復しない (A4 dormant) |
| 12 | 再検証結果 | 未実施 — 修復自体が別 unit のスコープ。再検証は修復後に同ケース ID (C09) で実施すること |

## Section A — egress インベントリ方法

走査パターン (backend/ + frontend/worker/ + backend/worker/、`*_test.go`/`httptest`/mock 除外):
`http.Get` `http.Post` `http.Head` `http.NewRequest` `http.NewRequestWithContext` `http.Client` 構築 `.Do(` `fetch(` `net.Dial` `net.DialContext` `s3.NewFromConfig` `smtp`/`netsmtp` provider constructor 群、env/config URL 入力 (`os.Getenv`/`getEnv` で URL/host を取る経路)。

## Section B — production egress サイト表

| # | 呼出サイト | 宛先ソース | host allowlist | URL 正規化 | redirect 方針 | private-IP guard | kill switch | per-site verdict |
|---|-----------|-----------|----------------|-----------|---------------|------------------|-------------|------------------|
| 1 | LSTEP API client — `infra/lstep/client.go:33` (`sharedHTTPClient`), `newRequest` :120-125, write methods (tag.go / user.go 経由) | clinic DB `lstep_base_url` (復号値) — write 時 `lstep_settings_service.go:294-299` + read 時 `lstep_settings_credentials.go` で再検証 | YES — `allowedLstepAPIHosts()` = `api.lstep.jp` のみ (`ValidateLstepBaseURL` credentials.go:27-58) | YES — `https://<host>` 再構成; HTTPS 強制・userinfo 拒否・IP literal/localhost 拒否・非 443 ポート拒否 | `denyRedirects` → err (dial.go:30-37) | YES — `hardenedDialContext` (dial.go:50-69): 名前解決→全 IP 検査→検査済み IP literal へ dial; `IsForbiddenDialIP` (:74-88) が loopback/private/link-local/multicast/unspecified/CGNAT 100.64.0.0/10 を拒否 | YES — `LSTEP_WRITE_API_ENABLED` exact `"true"` (client.go:26,89-100) + clinic `is_sync_enabled` 二重ゲート (lstep_settings_thresholds.go:102) | **PASS** (static) |
| 2 | LSTEP/LINE 疎通 probe — `lstep_settings_connection.go:25,36-71,127-161` | LSTEP: `ValidateLstepBaseURL` 再検証 (:97); LINE: `line.APIHost` const | YES — 同一 validator + const host | YES — HTTPS 強制 | `probeCheckRedirect` → errConnectionRedirect (:32-34, :39) | YES — `probeDialContext` (:52-70) → `infralstep.IsForbiddenDialIP` (:73-75); 10s timeout (:19) | read-only probe; 結果は安定コードで分類 (:166-192) | **PASS** (static) |
| 3 | LINE ID-token verify (LIFF middleware) — `middleware/liff_auth.go:149-164` | `lineVerifyURL` = `line.VerifyEndpoint` const `https://api.line.me/oauth2/v2.1/verify` (infra/line/client.go:19) | YES — compile-time const | n/a (const) | Go 既定 follow (≤10)。307/308 は POST body (id_token+client_id を含む) を再送 — 兄弟実装 line_link_service.go:572 との非対称 | NO | none (LIFF 認証に必須) | **PASS (static, note)** — 宛先は定数で SSRF 不可。redirect-follow 非対称は R1 |
| 4 | LINE ID-token verify (line-link flow) — `lstep/line_link_service.go:563-590` | `line.VerifyEndpoint` const | YES — const | n/a | `CheckRedirect` → ErrUseLastResponse (:572) — follow しない | NO | none | **PASS (static, note)** |
| 5 | LINE Messaging push — `infra/line/client.go:18,28,59`, `infra/line/push.go:120` | `PushEndpoint` const `https://api.line.me/v2/bot/message/push` | YES — const | n/a | Go 既定 follow (Authorization は net/http が cross-host で剥がす) | NO | channel token 空 → 送信されない (service 層) | **PASS (static, note)** |
| 6 | LINE push via service — `lstep/line_messaging_service.go:28,71` | `line.PushEndpoint` const | YES — const | n/a | Go 既定 follow | NO | `channelToken==""` → early return (:48-50) | **PASS (static, note)** |
| 7 | Discord webhook — `support/discord.go:44-50,126-132` | env `DISCORD_BUG_REPORT_WEBHOOK_URL` | NO — env 値。release のみ HTTPS/no-userinfo/no-query/no-fragment 検証 (`validateOptionalReleaseHTTPSURL` config.go:469-483) | TrimSpace 検査のみ | Go 既定 follow | NO | URL 空 → nil notifier (無効化) :45 | **FAIL (scoped)** — host allowlist・dial ガード・redirect 抑止なし → R2。宛先は operator-config (request 非影響) |
| 8 | Plane REST — `support/plane.go:91-113,189-201` | env `PLANE_BASE_URL` (default `api.plane.so`) | NO — env 値。release のみ HTTPS 検証 (config.go `validateOptionalReleaseHTTPSURL` 経由 :359 付近) | TrimRight "/" :106 | Go 既定 follow | NO | apiKey/slug/projectID 空 → nil client | **FAIL (scoped)** — 同上 → R2 |
| 9 | Support LLM (OpenAI 互換) — `support/llm.go:79-141` | env `SUPPORT_LLM_BASE_URL` (default `api.openai.com`) | NO — env 値。release のみ HTTPS 検証 (config.go:351) | TrimRight :86 | Go 既定 follow | NO | apiKey 空 → nil completer | **FAIL (scoped)** — 同上 → R2 |
| 10 | Medical-record lab drafter — `medicalrecord/lab_device_mapping_llm.go:76-90,132-139` | env `OPENAI_BASE_URL` | NO — env 値。**release HTTPS 検証なし** (config.go で `validateOptionalReleaseHTTPSURL` に通されていない点を確認) — 兄弟 env URL より弱い | TrimRight | `CheckRedirect` → ErrUseLastResponse (:89-90) — follow しない | NO | apiKey 空 → unconfigured result | **FAIL (scoped)** — release HTTPS 検証の不在 → R3 |
| 11 | S3/R2 — `infra/s3_file_storage.go:29`, `s3_uploader.go:35`, endpoint `s3_endpoint.go:11-34` | env `S3_ENDPOINT` (空=AWS), `S3_BUCKET`/`S3_REGION`/`S3_SHARED_BUCKET`, `S3_PUBLIC_BASE_URL` | env 値 — release で S3_ENDPOINT / S3_PUBLIC_BASE_URL を HTTPS 検証 (config.go:294-297); STORAGE_TYPE=s3 時 bucket 必須 (:240-243) | path-style override (R2 対応) | SDK 管理 | NO | release で STORAGE_TYPE=s3 必須 + credential 必須 (config.go:286-292) | **PASS (env-controlled)** — 宛先は運用者 infra 設定。AWS SDK が独自の endpoint 処理 |
| 12 | SMTP sender — `infra/smtp/sender.go:62-69,211-245` | env `SMTP_HOST`/`SMTP_PORT` | env 値 — release で host 必須 + port 465/587 限定 (config.go:326-339) | `net.JoinHostPort` :63 | n/a (SMTP)。TLS MinVersion 1.2 (:215-217) | NO — `net.Dialer.DialContext` 直接 (:69) | release で host/user/pass/from 必須 | **FAIL (scoped)** — private-IP dial guard なし → R4。SMTP は構造上任意ホスト必須のため allowlist 不適合、guard のみ |
| 13 | `cmd/csv-import-stg-uat` (operator CLI) | operator 指定 DB host (pgx DialFunc main.go:309-310) | operator 実行 | n/a | n/a | none | UAT ツール | **N/A** — リクエスト駆動 egress ではない (運用 CLI) |
| 14 | frontend Worker — `frontend/worker/index.ts:47,50,54` | `env.API` / `env.ASSETS` service bindings | platform binding (URL ではない) | n/a | n/a | n/a | n/a | **N/A** — binding 参照で URL egress なし |
| 15 | backend Worker — `backend/worker/index.ts:440,484-485,514,559,570,645`, `scheduled-jobs.ts:180` | `env.API_CONTAINER` container binding + 固定内部名 `http://container.internal` | platform binding + 固定名 | n/a | n/a | n/a | n/a | **N/A** — container binding |
| 16 | `internal/labdeviceagent/http.go:21,142-143,201,248` + `cmd/lab-device-agent` | inbound loopback server `127.0.0.1:17654` | n/a | n/a | n/a | n/a | n/a | **N/A** — inbound のみ (egress なし) |

## Section C — remediation candidates (修復せず記録のみ)

| ID | 対象 | 候補 (minimal) |
|----|------|----------------|
| R1 | `middleware/liff_auth.go:163` http.Client | `CheckRedirect` に `http.ErrUseLastResponse` を追加して兄弟実装 (`line_link_service.go:572`) と整合させる。307/308 時の body 再送 (id_token/client_id) を防ぐ。宛先は定数のため影響は限定的だが防御姿勢を揃える |
| R2 | env 駆動 HTTP client (discord.go:50 / plane.go:113 / llm.go:93) | LSTEP の `hardenedDialContext` 相当 (resolve→IP 検査→pinned dial) + redirect 抑止を共通 transport として抽出し適用する候補。宛先は operator-config 境界で SSRF 実害は限定されるが、設定ミス/将来の動的化に対する深度防御 |
| R3 | `medicalrecord/lab_device_mapping_llm.go` の `OPENAI_BASE_URL` | `config.go` の `validateOptionalReleaseHTTPSURL` を経由させて release モード HTTPS 強制を兄弟 env URL と揃える候補 (現状 redirect 抑止のみ) |
| R4 | `infra/smtp/sender.go:69` `net.Dialer` | resolved-IP private 判定を dial 前に入れる候補 (SMTP は構造上任意ホストを要求するため host allowlist ではなく dial 層 guard)。release の 465/587 制限 + TLS1.2 は既存 |

## Section D — 除外確認

- `NewInsecureTestClient` (client.go:76-85): `testing.Testing()` 外で panic — production 経路から hardened dial を迂回不能であることを静的確認。
- `*_test.go`, `httptest.Server`, mock client は全て非 production — 本判定の egress サイトから除外。
- LINE Webhook / LIFF / 通常 API は inbound 経路であり egress 対象外。
