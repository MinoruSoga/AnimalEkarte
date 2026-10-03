# codex-security 累積 findings トリアージ — 2026-10-03

## 概要

codex-security ワークベンチ (`~/.codex/state/plugins/codex-security/workbench.sqlite3`) の累積 56 findings(occurrence ベース、CLI 表示 51 件・dedupe 後)を、HEAD `fd65d1b2b` の現行コードに対して照合した。

- 検証方法: 各 finding の codeEvidence / 指摘ファイルを現行コードで再確認。静的照合のみ(実行・デプロイ検証なし)。
- 結果: **対応済み 48 / 本日修正実施 6 / 製品判断レーン 4 / 部分対応(残存) 2**
- **作業ツリーに大規模なセキュリティ修正 WIP (~49ファイル・未コミット) が存在**: support の mutation clinic スコープ化(P1)・seedlogin STG シークレット化(O8)・BugReportsPage UI・関連 docs を含む。本トリアージ後半でその一部を検証・補完した。

## 本日実施した修正(未対応 → 修正済み・未コミット)

| # | severity | finding | 修正内容 | 検証 |
|---|----------|---------|----------|------|
| O1 | medium | Fractional treatment quantities bypass inventory decrement | `inventory/repository.go` `DecreaseStock` が非正・非整数数量を `InvalidInput` で拒否(fail-closed)。`treatment_service.go` Create で InventoryID 指定時に整数数量を必須化 | `TestInventoryRepository_DecreaseStock` + `TestTreatmentService_Create` 追加・pass |
| O2 | medium | LIFF JSON-as-binary body limit 回避 | `reservation/routes.go` に `liffBodyLimit` 注入フィールド追加、`liff.Use()` で全 LIFF route に適用。`composition_runtime.go` で `middleware.LimitRequestBody(1MiB)` を配線 | reservation package tests pass |
| O3 | medium | Mutable third-party Action refs | `peaceiris/actions-gh-pages@v4.0.0`→SHA pin、`chromaui/action@v18.9.5`→SHA pin(ci-policy: 非公式 third-party は SHA pin、Chromaui は credential job のため保守的に pin) | version-drift script pass・YAML parse OK |
| O4 | medium | k6 unpinned install | `performance-tests.yml` を actionlint と同型の pin+SHA-256 検証インストールへ(k6 v2.3.0、checksum 既知) | YAML parse OK |
| O6 | low | Active-extension upload on local origin | `medical_record_image_request.go`: 拡張子↔宣言MIME一致を必須化、`verifySniffedContent` で先頭512バイトの DetectContentType 照合、保存拡張子を検証済みmimeから導出。`base_routes.go`: 非S3 `/uploads` に `X-Content-Type-Options: nosniff` | upload request tests 追加・pass |
| O7 | medium | Secret-sync unverified npm artifact | `worker-secret-sync.yml`: `npx -y wrangler@4.107.0` → `pnpm install --frozen-lockfile` + `pnpm exec wrangler`(lockfile integrity 検証済み `wrangler@4.114.0`) | YAML parse OK |

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

## 製品判断レーン — PO/運用確認が必要(4件)

| # | severity | finding | 状況 |
|---|----------|---------|------|
| P1 | high | Cross-clinic support board(bug-reports + chat-exchanges)3 occurrence 統合 | **mutation 側は未コミット WIP で修正済み**: `backend/internal/support/` 変更で status 更新・Plane 起票・削除を報告元 `clinic_id` スコープ化(`FindByIDForClinic`/`UpdateStatus(clinicID,...)`/他医院は404=閲覧のみ)。`BugReportResponse.ClinicID` 追加で UI 操作制御も対応。support package tests pass。**残る PO 判断は read 共有のみ**: 一覧・スクリーンショット・チャット本文の跨医院公開を正式受容するか、PII を含む detail を自院に絞るか |
| P2 | high | Public staging exposes demo sysadmin credential | **repo-public password 問題は WIP で解消**(O8 と同根)。残る論点: Internet 到達 STG 上に共有デモアカウント(全医院 executive 含む)を置く運用自体の受容 — シークレット化済みなら製品判断として許容可能か、STG ingress 制限/VPN が必要か |
| P3 | high | Clinic-scoped RBAC reused across assignments(#86 拠点横断) | GET fallback は opt-in 化済み(`RequirePermissionAllowingAssignedClinicGrant`)、write は selected-clinic 必須に修正済み。残存論点: `ResolveListClinicIDs` の membership 拡張は list read で所属 clinic への per-clinic resource grant を要求しない(#86 の意図仕様)。仕様受容か、拡張先 clinic にも view grant を要求するかの PO 確認 |
| P4 | medium | Support-chat conversations unredacted | P1 の read 側の一部。共有一覧で会話本文を全院に出す判断の受容確認 |

## 部分対応 — 残存リスク小(2件)

| # | severity | finding | 状況 |
|---|----------|---------|------|
| R1 | high | Public R2 domain bypass for shared files | 読取系は全て `GetSignedURL`(presigned, fail-closed)化済みで shared file / medical image とも app 認可経由に修正。残存: `S3_PUBLIC_BASE_URL` の仕組み自体は残る(`wrangler*.jsonc` は空 placeholder)。運用上 public domain を bucket に付けなければ無害 — 設定ガード文書化のみ |
| R2 | medium | Secret-sync unverified artifact | O7 修正で解消(version pin → lockfile integrity) |

## 残存する運用確認事項

- **Terraform state**: Hyperdrive リソースは撤去済みだが、過去の local state に書かれた DB password の処分・ローテーションはオペレータ作業(`infra/cloudflare/backend.tf` コメントに記載)。
- **STG demo account 存在**: O8 修正で repo-public password は排除済みだが、SEEDLOGIN_DEMO_PASSWORD によるデモログイン自体は STG で有効のまま(P2 の製品判断)。

## 未コミット WIP の取り扱い

作業ツリーに ~49 ファイルの未コミット変更がある(内訳: 本日実施の O1-O4/O6/O7 修正 9ファイル + 別セッション由来の support/seedlogin/frontend WIP ~40ファイル)。**support・seedlogin 系の WIP は今回のトリアージで変更していない foreign WIP** であり、support tests は pass を確認済み。コミット・PR は未指示のため実施していない。
