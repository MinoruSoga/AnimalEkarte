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
| 5 | 拒否条件 | LSTEP/probeのvalidator・dial・redirect拒否はstatic確認。その他の通信先は定数または運用者設定であり、requestが制御できる宛先や信頼境界越えは未証明。全providerに同じpublic-only allowlist/private-IP禁止を要求する根拠は未確定。 |
| 6 | 防御判定 | **UNKNOWN (全体)** — LSTEP/probeはstatic PASS。operator-config先のguard不足は強化候補であり、request-driven SSRF成立を示さない。定数LINE先でもredirect/DNSの信頼境界は別に評価する。旧scoped FAILは判定範囲の訂正であり修復によるPASSではない。 |
| 7 | 副作用呼出回数 | LSTEP/probeの拒否経路はstatic確認のみ。検証中の実outbound/DNS/dialは0回。その他providerの拒否副作用ゼロを実証したものではない。 |
| 8 | 検出判定 | `UNKNOWN` — probe 拒否は `slog.WarnContext` で構造化ログ出力 (lstep_settings_connection.go:100,104,116) + `classifyConnectionProbeError` 安定コード (:166-192) を静的確認。ただしランタイム検出・通知経路の実測は staging-observed が BLOCKED のため未検証 |
| 9 | 安全な証跡参照 | 本ファイル Section A/B の file:line。具体 URL/トークン値は記録しない |
| 10 | 制約・未検証範囲 | 実 outbound 呼出なし・DNS 解決なし・redirect 追跡なし (実行しない設計)。`*_test.go` / `httptest` / mock (`NewInsecureTestClient` client.go:76-85 は `testing.Testing()` 外で panic する test-only 経路) は本判定から除外。staging 観測は未実施。Workers runtime の service-binding 宛先は wrangler.jsonc の静的値で推定するのみ |
| 11 | 修復案 | R1–R4は未適用の強化候補。宛先設定権限、許可宛先、redirect/DNSの信頼境界と正規private SMTP relayの契約を確認してから決める。 |
| 12 | 再検証結果 | コード修復・実通信の再検証は未実施。2026-10-09 Codex follow-upでrubricを訂正し全体UNKNOWNとした。 |

## Section A — egress インベントリ方法

走査パターン (backend/ + frontend/worker/ + backend/worker/、`*_test.go`/`httptest`/mock 除外):
`http.Get` `http.Post` `http.Head` `http.NewRequest` `http.NewRequestWithContext` `http.Client` 構築 `.Do(` `fetch(` `net.Dial` `net.DialContext` `s3.NewFromConfig` `smtp`/`netsmtp` provider constructor 群、env/config URL 入力 (`os.Getenv`/`getEnv` で URL/host を取る経路)。

## Section B — production egress サイト表

| # | 呼出サイト | 宛先ソース | host allowlist | URL 正規化 | redirect 方針 | private-IP guard | kill switch | per-site verdict |
|---|-----------|-----------|----------------|-----------|---------------|------------------|-------------|------------------|
| 1 | LSTEP API client — `infra/lstep/client.go:33` (`sharedHTTPClient`), `newRequest` :120-125, write methods (tag.go / user.go 経由) | clinic DB `lstep_base_url` (復号値) — write 時 `lstep_settings_service.go:294-299` + read 時 `lstep_settings_credentials.go` で再検証 | YES — `allowedLstepAPIHosts()` = `api.lstep.jp` のみ (`ValidateLstepBaseURL` credentials.go:27-58) | YES — `https://<host>` 再構成; HTTPS 強制・userinfo 拒否・IP literal/localhost 拒否・非 443 ポート拒否 | `denyRedirects` → err (dial.go:30-37) | YES — `hardenedDialContext` (dial.go:50-69): 名前解決→全 IP 検査→検査済み IP literal へ dial; `IsForbiddenDialIP` (:74-88) が loopback/private/link-local/multicast/unspecified/CGNAT 100.64.0.0/10 を拒否 | YES — `LSTEP_WRITE_API_ENABLED` exact `"true"` (client.go:26,89-100) + clinic `is_sync_enabled` 二重ゲート (lstep_settings_thresholds.go:102) | **PASS** (static) |
| 2 | LSTEP/LINE 疎通 probe — `lstep_settings_connection.go:25,36-71,127-161` | LSTEP: `ValidateLstepBaseURL` 再検証 (:97); LINE: `line.APIHost` const | YES — 同一 validator + const host | YES — HTTPS 強制 | `probeCheckRedirect` → errConnectionRedirect (:32-34, :39) | YES — `probeDialContext` (:52-70) → `infralstep.IsForbiddenDialIP` (:73-75); 10s timeout (:19) | read-only probe; 結果は安定コードで分類 (:166-192) | **PASS** (static) |
| 3 | LINE ID-token verify (LIFF middleware) — `middleware/liff_auth.go:149-164` | `lineVerifyURL` = `line.VerifyEndpoint` const `https://api.line.me/oauth2/v2.1/verify` (infra/line/client.go:19) | YES — compile-time const | n/a (const) | Go 既定 follow (≤10)。307/308 は POST body (id_token+client_id を含む) を再送 — 兄弟実装 line_link_service.go:572 との非対称 | NO | none (LIFF 認証に必須) | **UNKNOWN (redirect leg)** — 初期URLは定数。redirect先へのbody再送・到達は未観測 → R1 |
| 4 | LINE ID-token verify (line-link flow) — `lstep/line_link_service.go:563-590` | `line.VerifyEndpoint` const | YES — const | n/a | `CheckRedirect` → ErrUseLastResponse (:572) — follow しない | NO | none | **PASS (初期宛先・redirect抑止のstatic範囲)** — DNS/dial実挙動は未観測 |
| 5 | LINE Messaging push — `infra/line/client.go:18,28,59`, `infra/line/push.go:120` | `PushEndpoint` const `https://api.line.me/v2/bot/message/push` | YES — const | n/a | Go 既定 follow (Authorization は net/http が cross-host で剥がす) | NO | channel token 空 → 送信されない (service 層) | **UNKNOWN (redirect/DNS leg)** — 初期URLは定数。資格情報移送や到達は未観測 |
| 6 | LINE push via service — `lstep/line_messaging_service.go:28,71` | `line.PushEndpoint` const | YES — const | n/a | Go 既定 follow | NO | `channelToken==""` → early return (:48-50) | **UNKNOWN (redirect/DNS leg)** — 同上 |
| 7 | Discord webhook — `support/discord.go:44-50,126-132` | env `DISCORD_BUG_REPORT_WEBHOOK_URL` | NO — env 値。release のみ HTTPS/no-userinfo/no-query/no-fragment 検証 (`validateOptionalReleaseHTTPSURL` config.go:469-483) | TrimSpace 検査のみ | Go 既定 follow | NO | URL 空 → nil notifier (無効化) :45 | **UNKNOWN (policy/impact)** — operator-config先。guard不足はR2候補、request-driven SSRF未証明 |
| 8 | Plane REST — `support/plane.go:91-113,189-201` | env `PLANE_BASE_URL` (default `api.plane.so`) | NO — env 値。release のみ HTTPS 検証 (config.go `validateOptionalReleaseHTTPSURL` 経由 :359 付近) | TrimRight "/" :106 | Go 既定 follow | NO | apiKey/slug/projectID 空 → nil client | **UNKNOWN (policy/impact)** — operator-config先。R2候補 |
| 9 | Support LLM (OpenAI 互換) — `support/llm.go:79-141` | env `SUPPORT_LLM_BASE_URL` (default `api.openai.com`) | NO — env 値。release のみ HTTPS 検証 (config.go:351) | TrimRight :86 | Go 既定 follow | NO | apiKey 空 → nil completer | **UNKNOWN (policy/impact)** — operator-config先。R2候補 |
| 10 | Medical-record lab drafter — `medicalrecord/lab_device_mapping_llm.go:76-90,132-139` | env `OPENAI_BASE_URL` | NO — env 値。**release HTTPS 検証なし** (config.go で `validateOptionalReleaseHTTPSURL` に通されていない点を確認) — 兄弟 env URL より弱い | TrimRight | `CheckRedirect` → ErrUseLastResponse (:89-90) — follow しない | NO | apiKey 空 → unconfigured result | **UNKNOWN (SSRF impact)** — release HTTPS検証不在は設定境界の独立したR3候補。実enablement/宛先は未確認 |
| 11 | S3/R2 — `infra/s3_file_storage.go:29`, `s3_uploader.go:35`, endpoint `s3_endpoint.go:11-34` | env `S3_ENDPOINT` (空=AWS), `S3_BUCKET`/`S3_REGION`/`S3_SHARED_BUCKET`, `S3_PUBLIC_BASE_URL` | env 値 — release で S3_ENDPOINT / S3_PUBLIC_BASE_URL を HTTPS 検証 (config.go:294-297); STORAGE_TYPE=s3 時 bucket 必須 (:240-243) | path-style override (R2 対応) | SDK 管理 | NO | release で STORAGE_TYPE=s3 必須 + credential 必須 (config.go:286-292) | **UNKNOWN (policy/impact)** — operator-config先。release設定検証の存在はstatic確認、全egress安全性のPASSではない |
| 12 | SMTP sender — `infra/smtp/sender.go:62-69,211-245` | env `SMTP_HOST`/`SMTP_PORT` | env 値 — release で host 必須 + port 465/587 限定 (config.go:326-339) | `net.JoinHostPort` :63 | n/a (SMTP)。TLS MinVersion 1.2 (:215-217) | NO — `net.Dialer.DialContext` 直接 (:69) | release で host/user/pass/from 必須 | **UNKNOWN (policy/impact)** — operator-config先。正規private relayを許す契約を確認するまでR4は適用しない |
| 13 | `cmd/csv-import-stg-uat` (operator CLI) | operator 指定 DB host (pgx DialFunc main.go:309-310) | operator 実行 | n/a | n/a | none | UAT ツール | **N/A** — リクエスト駆動 egress ではない (運用 CLI) |
| 14 | frontend Worker — `frontend/worker/index.ts:47,50,54` | `env.API` / `env.ASSETS` service bindings | platform binding (URL ではない) | n/a | n/a | n/a | n/a | **N/A** — binding 参照で URL egress なし |
| 15 | backend Worker — `backend/worker/index.ts:440,484-485,514,559,570,645`, `scheduled-jobs.ts:180` | `env.API_CONTAINER` container binding + 固定内部名 `http://container.internal` | platform binding + 固定名 | n/a | n/a | n/a | n/a | **N/A** — container binding |
| 16 | `internal/labdeviceagent/http.go:21,142-143,201,248` + `cmd/lab-device-agent` | inbound loopback server `127.0.0.1:17654` | n/a | n/a | n/a | n/a | n/a | **N/A** — inbound のみ (egress なし) |

## Section C — remediation candidates (修復せず記録のみ)

| ID | 対象 | 候補 (minimal) |
|----|------|----------------|
| R1 | `middleware/liff_auth.go:163` http.Client | `CheckRedirect` に `http.ErrUseLastResponse` を追加して兄弟実装 (`line_link_service.go:572`) と整合させる。307/308 時の body 再送 (id_token/client_id) を防ぐ。宛先は定数のため影響は限定的だが防御姿勢を揃える |
| R2 | env駆動HTTP client | 設定変更権限と許可宛先、redirect/DNS境界を確認しprovider単位にtransport方針を決める。影響未評価。共通public-only guardの一律導入はしない。 |
| R3 | `medicalrecord/lab_device_mapping_llm.go` の `OPENAI_BASE_URL` | `config.go` の `validateOptionalReleaseHTTPSURL` を経由させて release モード HTTPS 強制を兄弟 env URL と揃える候補 (現状 redirect 抑止のみ) |
| R4 | SMTP sender | 正規のprivate SMTP relay・TLS・port契約を確認してprovider別の宛先制御を検討する。private-IP一律拒否は未採用。 |

## Section D — 除外確認

- `NewInsecureTestClient` (client.go:76-85): `testing.Testing()` 外で panic — production 経路から hardened dial を迂回不能であることを静的確認。
- `*_test.go`, `httptest.Server`, mock client は全て非 production — 本判定の egress サイトから除外。
- LINE Webhook / LIFF / 通常 API は inbound 経路であり egress 対象外。

## Codex follow-up — 2026-10-09

同じoperator-config境界のS3をPASS、SMTPをprivate-IP guard不在だけでFAILにした旧rubricを訂正。通常callerからのauthority獲得を仮定しない (`SECURITY.md`)。この判定訂正は通信実測や製品修復の証拠ではない。詳細は [Codex follow-up](../repair-20261009.md)。
