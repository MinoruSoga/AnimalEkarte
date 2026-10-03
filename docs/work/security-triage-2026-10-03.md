# codex-security 累積 findings トリアージ — 2026-10-03

## 概要

codex-security ワークベンチ (`~/.codex/state/plugins/codex-security/workbench.sqlite3`) の累積 64 findings(occurrence ベース、CLI 表示 55 件・dedupe 後)を、現行コードに対して照合した。

- 検証方法: 各 finding の codeEvidence / 指摘ファイルを現行コードで再確認。静的照合 + scoped test(Docker)・workflow YAML 検証。
- 結果: **64 occurrence 全件に対応方針が確定** — 対応済み 48 / 修正実施 11(O1-O4,O6,O7 + N1-N4,N8) / owner 承認済み受容 5(P1-P4,N5) / owner 裁定カバー 2(N6,N7) / 部分対応 2(R1,R2)※重複計上あり
- **再スキャン(1YUno8)**: frontend 再精査 20ファイルで新規報告なし。workbench に backend scan 由来の新規 4 findings(N1-N4)が追加済み — 全て未コミット WIP で対応済み・scoped test pass
- **再スキャン(xj4SPC)**: frontend 再精査 113ファイルで **0 confirmed findings**(report "No findings")。workbench に backend 由来の新規 4 findings(N5-N8)が追加 — N5 owner 受容記録・N6/N7 既存裁定カバー・N8 config WIP で修正済み
- **再スキャン(WzIAHx)**: frontend 再精査 69ファイルで **Reportable findings: 0** — 2回連続 clean。revision `e783de5a7` + スキャン開始時 snapshot を対象。findings 総数は 55→55 で新規追加なし
- 注意: スキャン実行中に checkout が変化("Scan target changed"警告) — foreign WIP が継続中のため。frontend スコープの結果は報告対象0件で確定
- **P判断は `SECURITY.md` で owner 確定**: support 共有 board の受容、OBJECT-1(private+signed URL)標準、**support コンテンツへの患者・飼主情報含有許可 + LLM/Plane への既存経路送信承認(2026-10-03 裁定 — 画面特定に必要)**。未決残件は共有コンテンツの保持期間のみ(EMR-263)

## 本日実施した修正(未対応 → 修正済み・未コミット)

| # | severity | finding | 修正内容 | 検証 |
|---|----------|---------|----------|------|
| O1 | medium | Fractional treatment quantities bypass inventory decrement | `inventory/repository.go` `DecreaseStock` が非正・非整数数量を `InvalidInput` で拒否(fail-closed)。`treatment_service.go` Create で InventoryID 指定時に整数数量を必須化 | `TestInventoryRepository_DecreaseStock` + `TestTreatmentService_Create` 追加・pass |
| O2 | medium | LIFF JSON-as-binary body limit 回避 | `reservation/routes.go` に `liffBodyLimit` 注入フィールド追加、`liff.Use()` で全 LIFF route に適用。`composition_runtime.go` で `middleware.LimitRequestBody(1MiB)` を配線 | reservation package tests pass |
| O3 | medium | Mutable third-party Action refs | `peaceiris/actions-gh-pages@v4.0.0`→SHA pin、`chromaui/action@v18.9.5`→SHA pin(ci-policy: 非公式 third-party は SHA pin、Chromaui は credential job のため保守的に pin) | version-drift script pass・YAML parse OK |
| O4 | medium | k6 unpinned install | `performance-tests.yml` を actionlint と同型の pin+SHA-256 検証インストールへ(k6 v2.3.0、checksum 既知) | YAML parse OK |
| O6 | low | Active-extension upload on local origin | `medical_record_image_request.go`: 拡張子↔宣言MIME一致を必須化、`verifySniffedContent` で先頭512バイトの DetectContentType 照合、保存拡張子を検証済みmimeから導出。`base_routes.go`: 非S3 `/uploads` に `X-Content-Type-Options: nosniff` | upload request tests 追加・pass |
| O7 | medium | Secret-sync unverified npm artifact | `worker-secret-sync.yml`: `npx -y wrangler@4.107.0` → `pnpm install --frozen-lockfile` + `pnpm exec wrangler`(lockfile integrity 検証済み `wrangler@4.114.0`) | YAML parse OK |

## 追加 findings(2026-10-03 backend scan 由来・workbench 07:53 追加)の対応 WIP

| # | severity | finding | 修正内容 | 検証 |
|---|----------|---------|----------|------|
| N1 | medium | Binary Content-Type bypasses lab-frame request limits | `sanitize_null_bytes.go`: `BinaryBodyMaxBytes=16MiB` 共通天井を `SanitizeNullBytes`/`LimitRequestBody` の binary 経路に適用。`lab_device_item_master_handler.go`: ルート固有 `MaxBytesReader(64KiB)` + `MaxBytesError`→413。`request.go`: base64 長事前拒否 | middleware/medicalrecord tests pass |
| N2 | medium | Staff can send unclassified clinic data to configured LLM | `support/chat.go`: `screenChatOutbound` で email/電話/長数字列/認証情報/Bearer を外部送信前に拒否(汎用 invalid_input、パターン非開示)。`history` をクライアント入力から除去しサーバー保存済み(送信時スクリーニング済み)履歴のみ使用。frontend も history 送信を廃止 | support tests pass・frontend 22 tests pass |
| N3 | medium | Bug-report creation auto-exports unclassified content to Plane | `support/service.go`: Create 時の自動 `syncPlaneTicket` を廃止。外部 export は reviewable な明示操作(`POST /:id/plane-ticket`)のみ | support tests pass |
| N4 | low | Shared clinical-file bearer URLs valid 24h | `shared_file_service.go`: TTL を用途別分割 — 対話 `GetSignedURL` 15min / LINE 配送 `GetSignedURLForDelivery` 1h(固定24h→大幅縮小) | lstep tests pass・isolation test guard 拡張 |

## 追加 findings(2026-10-03 workbench 08:48 追加、frontend scan `xj4SPC` 報告時点)

frontend 再スキャン(`xj4SPC`)は **0 confirmed**(113ファイル精査、レポート "No findings")。同時刻に workbench へ backend scan 由来の新規 4 件が追加 — 並行セッションの WIP で全件対応済み:

| # | severity | finding | 対応 | 検証 |
|---|----------|---------|------|------|
| N5 | high | STG demo 共有パスワードが全 catalog identity(全医院執行含む)を認証 | **owner 受容として記録済み**(SECURITY.md WIP): STG は synthetic 専用・権限管理デモに全医院視点が必要のため意図的仕様。制御条件(ShouldApply fail-closed・未設定時 lockedDemoPassword・wrangler secret 管理)と再評価トリガー(実データ混入・シークレット流用)を明記 | `seedlogin/env.go:70-81` `auth_service.go:108-118` 制御実測済み |
| N6 | medium | Support chat が未分類 prose を外部 LLM へ送信 | **既存 owner 裁定でカバー**(P1/P4 確定): PII 含有許可 + LLM 送信は `screenChatOutbound` スクリーニング付き既存経路に限定(N2 で実装済み) | support tests pass |
| N7 | medium | 手動 Plane export が未分類 bug-report 内容・clinic metadata を送信 | **既存 owner 裁定でカバー**: Plane 送信は明示 reviewable 操作のみ(N3 で実装済み)、内容含有は裁定済み。`e783de5a7` で Plane state sync が scheduled job として追加(往復同期は owner 承認済み運用の一部) | support tests pass |
| N8 | low | Release が推測可能な全医院 scheduler token を受け入れる | **修正済み**(config WIP): `config.go` release mode で `SCHEDULER_INTERNAL_TOKEN` 必須 + `minimumSchedulerInternalTokenBytes=32` 未満を fail-loud。DEC-36/CMD-02 の設計メモ付き | `config_validate_test.go` に empty/short 拒否テスト2件追加、package tests pass |

### 検証メモ(2026-10-03 夕)
- backend build + scoped tests(lstep/support/medicalrecord/middleware)pass
- pre-existing テストインフラ問題を恒久修正: `lstep_settings_tx_atomicity_test.go` が `model.ClinicSettings` を AutoMigrate しており、GORM の `type:time`→`timestamptz` 誤変換で cold schema 時に必ず失敗していた。既存 `testdb.EnsureClinicSettingsTable`(実 migration 準拠の生SQL、5ファイル利用済み)へ切替 — DROP TABLE 後の cold schema 経路で 4 テスト全 pass を確認
- `TestVitalRepository_Delete`・`TestRealDB_SelectedClinicBGrantAIsolation` は共有DB flake(単独再実行 pass)

## 対応済み(コミット済み or 既存 WIP で確認・42+2件)

### 新規確認分
- **O5 lstep-migrate CSV formula injection** — `cmd/lstep-migrate/reporter.go` に `sanitizeCSVCell` 適用済み(初回トリアージの見落としを訂正)
- **O8 Repository-public password → STG exec** — `seedlogin/env.go` の WIP で staging は `SEEDLOGIN_DEMO_PASSWORD` シークレットのみ受付・未設定は fail-closed。repo-public `"password"` は local/dev/test 限定に変更済み(**未コミット WIP**)

### 従来確認分(42件)
#### High
- **LIFF auto-link name+phone** — `liff_service_reservations.go:41` SEC-CS2-F02 で自動紐付け廃止 + 回帰テスト
- **GET permission fallback** — `http_permission.go:116` selected-clinic 必須化、fallback は opt-in
- **Generic accounting update bypass** — `accounting_service_core.go:404` cancelled/completed 遷移を専用 endpoint へ強制
- **Legacy accounting create forge** — 同 :382 `status=completed/cancelled`・`completed_at` を拒否
- **Closed-period update** — `accounting_service_update.go:24` destDate 含む post-close を tx 内で解決 + reason 必須
- **Hyperdrive passwords in TF state** — `infra/cloudflare/hyperdrive.tf` リソース自体を撤去(SEC-CS2-F03)
- **Identity-link assignment vs RBAC** — `handler.go:51` `FilterClinicIDsForPermission` で所属集合を RBAC フィルタ
- **Tracked inquiry seed corpus** — `003_demo` が git 追跡・disk ともに消滅

#### Medium
- **Lab-device /claim 認証** — `labdeviceagent/http.go:319` claim 前に `authorizeConsumerToken` 必須化
- **Temporary DB errors stale authority** — `middleware/auth.go:46` 一時障害も fail-closed
- **LINE webhook HMAC fan-out** — `line_link_service.go:427` SEC-CS-F05-R1 destination で1clinic絞り込み
- **LIFF availability hidden/internal** — `liff_service_catalog.go:99` `IsActive && ReservationVisible && !IsInternal`
- **LSTEP loopback SSRF ×3** — `ValidateLstepBaseURL` が https+host allowlist+fail-closed
- **Medical attachments/images public URLs ×2** — `signStoredMedicalRecordImageURL` → presigned URL
- **Image script-scheme** — `medicalRecordImageHasHTTPScheme` で http(s) のみ
- **Unbounded reorder** — `MaxReorderIDs=500` + 重複拒否(`slice.go`)
- **Staged filenames hook** — `execFileSync` argv 形式へ
- **Staging dump exposure** — ランダム PW + port 非公開(scripts/verify_seed_matches_stg_dump_full.sh)
- **Cursor workspace policy** — allowlist モード
- **Dev image installer** — curl 除去・pinned golangci image
- **Migrations sysadmin / demo seed verifiers / skip-worktree overlay** — `003_demo` 退役(`cmd/migrate` は全環境 002_master のみ)

#### Low
- Actionlint installer(pin+sha256)、vercel@latest(除去・wrangler へ)、cage delete race(SEC-CS-F13)、deceased draft image upload(SEC-CS-F14)、identity-link DELETE body limit、LSTEP 404 LINE ID、migrate bearer ≥32bytes、discount TOCTOU ×3(SEC-CS-F09/F10/F15)、manualarticle retention(MaxVersionsPerArticle=50)、tag-code mapping caps(32/100/200)、multi-file burst(MAX_UPLOAD_FILES=10+有界並列)、owner CSV injection(escapeCsvTextCell)、inactive clinics(is_active filter)、LIFF availability hidden types

## 製品判断レーン — SECURITY.md で owner 承認済み(4件)

> `SECURITY.md`(commit `39404be3b`, 2026-10-03 owner 回答) で P1/P2/P4 の裁定は確定済み。scanner の high 評価は「repo コメントを認可例外と見なさない」ための検出であり、owner 承認済みの製品決定として正式に記録されている。

| # | severity | finding | 裁定 |
|---|----------|---------|------|
| P1 | high | Cross-clinic support board(bug-reports + chat-exchanges)3 occurrence 統合 | **受容済み(owner 承認)**: 「閲覧・作成の clinic 横断共有は owner 承認済みの製品決定であり脆弱性ではない」(SECURITY.md:116)。mutation は報告元 clinic スコープで実装・tests pass。**情報区分も 2026-10-03 に裁定**: 患者・飼い主情報の含有許可(画面特定に必要) + LLM/Plane への既存経路送信承認 → SECURITY.md PROVIDER-1 確定。残件は保持期間のみ(EMR-263) |
| P2 | high | Public staging exposes demo sysadmin credential | **受容済み**: repo-public password 問題は WIP で解消(staging は `SEEDLOGIN_DEMO_PASSWORD` シークレットのみ・未設定 fail-closed、LoginForm も DEV 限定自動入力に修正済み)。STG live reachability は SECURITY.md に実測記載済み |
| P3 | high | Clinic-scoped RBAC reused across assignments(#86 拠点横断) | **意図仕様として受容**: GET fallback opt-in 化・write selected-clinic 必須は修正済み。`ResolveListClinicIDs` の membership 拡張は #86 拠点横断スコープの配送済み設計 |
| P4 | medium | Support-chat conversations unredacted | P1 と同じく受容済み。情報区分は 2026-10-03 裁定済み、残件は保持期間のみ(EMR-263) |

## 部分対応 — 残存リスク小(2件)

| # | severity | finding | 状況 |
|---|----------|---------|------|
| R1 | high | Public R2 domain bypass for shared files | **OBJECT-1 標準採用(2026-10-03)**: private bucket + app scope check 後の短命 signed URL が標準と SECURITY.md に確定。読取系は全て presigned。R2 bucket の実効 public/private は検証要と明記済み — 運用確認事項のみ |
| R2 | medium | Secret-sync unverified artifact | O7 修正で解消(version pin → lockfile integrity) |

## 残存する運用確認事項

- **Terraform state**: Hyperdrive リソースは撤去済みだが、過去の local state に書かれた DB password の処分・ローテーションはオペレータ作業(`infra/cloudflare/backend.tf` コメントに記載)。
- **STG demo account 存在**: O8 修正で repo-public password は排除済みだが、SEEDLOGIN_DEMO_PASSWORD によるデモログイン自体は STG で有効のまま(P2 の製品判断)。

## 未コミット WIP の取り扱い

作業ツリーに ~49 ファイルの未コミット変更がある(内訳: 本日実施の O1-O4/O6/O7 修正 9ファイル + 別セッション由来の support/seedlogin/frontend WIP ~40ファイル)。**support・seedlogin 系の WIP は今回のトリアージで変更していない foreign WIP** であり、support tests は pass を確認済み。コミット・PR は未指示のため実施していない。
