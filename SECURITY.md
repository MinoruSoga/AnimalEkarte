# Security Policy — AnimalEkarte

> **状態:** 2026-10-03 にオーナー確認済みで採用。残る `[要オーナーレビュー]` の決定は、明示承認されるまで除外、受容済みリスク、finding の抑制根拠として扱ってはなりません。未確認の deployment 状態も安全性の証拠にはなりません。

## 対応するバージョン

| バージョン | サポート状況 |
|---|---|
| `main`（最新） | セキュリティパッチ適用中 |
| `staging` | セキュリティパッチ適用中 |
| その他 | 未サポート |

`production` branch と production deploy workflow は存在しますが、production 設定は draft であり **live ではありません**（`docs/ops/infra/production/setup.md` の人間チェックリスト未消化、2026-10-03 確認）。サポート対象への追加は go-live 後に行います。`backend/wrangler.production.jsonc:1-8` `.github/workflows/backend-deploy.yml:162-237` `.github/workflows/frontend-deploy.yml:31-109`

未サポート branch であっても、共有コードや実際に稼働する環境への影響を報告対象外にはしません。

## 脆弱性の報告

**公開 Issue で報告しないでください。**

- Email: baritech.soga@gmail.com
- 件名: `[SECURITY] AnimalEkarte - <概要>`
- 初期返答の目標: 48時間以内

可能な範囲で次を含めてください。

1. 問題の種類と影響を受けるバージョン・環境
2. 再現手順と必要な設定条件
3. 影響を受けるデータ、医院、権限または機能
4. 最小限の proof of concept と、任意の修正案

実在する患者・飼い主情報、認証情報、token、secret、完全な database dump は送らず、必要最小限に加工してください。

## システムとスコープ

AnimalEkarte は日本国内の動物病院向け電子カルテです。スタッフが飼い主、ペット、予約、診療、入院、検査、会計、在庫を扱い、飼い主は LINE/LIFF 経由の機能を利用します。通常のクラウド経路は、ブラウザまたは LINE から Cloudflare frontend Worker、service binding、API Worker、Go Container、PostgreSQL・R2・外部 provider へ進みます。`README.md:1-43` `frontend/worker/index.ts:43-60`

この root policy は、frontend、Go API、Workers、database migration、object storage、外部連携、CI/deploy、operator CLI、local lab agent を対象にします。保護対象には clinic-owned の診療・会計・顧客データ、staff identity と session、integration credential、R2 object、support content、監査・scheduler・migration 記録、lab frame、deploy・DNS・backup/restore authority が含まれます。

依存パッケージ配下の `SECURITY.md` は各依存の報告方針であり、この製品の root policy を置き換えません。製品から到達可能な依存動作は引き続き評価対象です。

## Threat Model と信頼境界

想定する開始権限は、未認証 Internet caller、本人の LINE token を持つ LIFF 利用者、clinic A の通常 staff、system administrator、外部 provider、CI/operator、および病院 Mac 上の local process です。攻撃者が最初から deployment secret、DB owner、Cloudflare/GitHub 管理権限を持つとは仮定しません。

| Surface | 適用する境界 |
|---|---|
| Login・refresh・password reset と staff API | 公開 auth route は endpoint 固有の CSRF/rate limit、保護 route は staff authentication と CSRF、その後に domain 固有の RBAC・ownership を適用します。`backend/internal/auth/http_routes.go:53-93` `backend/cmd/api/composition_runtime.go:494-564` |
| LIFF・LINE webhook | staff RBAC とは別の公開境界です。LIFF は LINE identity と path clinic/customer、webhook は本文上限・destination-to-clinic 解決・clinic credential による HMAC を使用します。`backend/internal/reservation/routes.go:157-188` `backend/internal/lstep/routes.go:226-233` `backend/internal/lstep/line_link_handler.go:52-80` |
| Health・local uploads・UAT | health は公開です。`/uploads` は非 S3 の local/dev のみ、UAT fixture は許可された開発環境・local DB・Host 以外では 404 となる条件付き surface です。`backend/cmd/api/base_routes.go:46-72` `backend/internal/billing/synthetic_closing_http.go:21-55` `backend/internal/billing/synthetic_closing_http.go:133-140` |
| Support | 認証済み staff 全員が専用 permission なしで利用します。一覧・作成・chat exchanges は全 clinic 共有（owner 承認済みの製品決定）ですが、status 更新・Plane 起票・削除は報告元 clinic スコープで他医院からは 404 です。`backend/internal/support/handler.go:47-74` `backend/internal/support/handler.go:160-239` |
| Worker internal operations | Migration と scheduler operations は通常の Container proxy より先に分岐します。未知の internal path は Container へ転送しません。`backend/worker/index.ts:475-507` `backend/worker/index.ts:582-629` |
| Local lab agent | loopback Host、Origin allowlist、consumer token、clinic binding、lease と backend の `lab-import` permission を別々に適用します。`backend/cmd/lab-device-agent/main.go:33-105` `backend/internal/labdeviceagent/http.go:243-291` `backend/internal/labdeviceagent/http.go:318-430` `backend/internal/medicalrecord/routes_lab.go:9-50` |
| Release・migration・operator CLI | CI、Container 起動 migration、CSV import、staff provisioning は通常 staff API とは異なる高権限 workflow です。`backend/Dockerfile.production:63-78` `backend/cmd/csv-import/main.go:291-411` `backend/cmd/staff-provision/main.go:168-215` |

## Security Invariants

1. **TENANT-1 — clinic isolation:** clinic-owned data の read、write、delete、join、preload、count、bulk、background path は、認証後に再解決した clinic authority と resource ownership で制約しなければなりません。client supplied の clinic、owner、pet、staff、record ID は認可根拠になりません。`docs/architecture/overview.md:82-89` `backend/internal/middleware/auth.go:58-153` `backend/internal/middleware/auth.go:321-370`
2. **AUTHZ-1 — 独立した認可:** authentication、resource/action permission、clinic scope、resource ownership は独立した check です。RBAC は deny-by-default とし、system administrator の resource bypass を tenant bypass と混同してはなりません。実装上の `AllResources` は 36 種です。`backend/internal/auth/http_permission.go:61-167` `backend/internal/model/permission.go:65-103`
3. **PUBLIC-1 — route 固有の公開境界:** LIFF、LINE webhook、health、internal operations、local uploads、UAT、support、lab agent を一律の `RequirePermission` 対象として記述してはなりません。それぞれの identity、integrity、environment、method、rate-limit 条件が失敗時に fail closed でなければなりません。
4. **PRIV-1 — 特権操作:** migration、scheduler control/manual run、all-clinic batch、release、provisioning、import は通常 staff session と分離した credential または検証済み operator identity を要求し、対象・revision・入力・backup・結果を結び付けなければなりません。`backend/worker/migrate-exec.ts:18-113` `backend/worker/scheduler-access-auth.ts:350-441` `backend/worker/scheduler-ops.ts:302-409` `backend/cmd/api/batch_scheduler.go:17-81`
5. **OBJECT-1 — object storage:** 診療画像と support screenshot を保存する bucket は private-by-default とし、application が scope check 後に短命の signed URL を発行する設計を標準とします（実装と一致するため採用 — 2026-10-03）。公開 bucket/domain は、公開可能な data class を owner が明示した場合だけ許可します。R2 bucket の実効 public/private 設定は検証要。`backend/internal/infra/s3_uploader.go:24-88` `backend/internal/medicalrecord/medical_record_image_handler.go:277-316` `backend/internal/medicalrecord/medical_record_image_response.go:14-18`
6. **PROVIDER-1 — 外部送信:** **[2026-10-03 確定]** LLM、ticket、LINE/LSTEP、SMTP、alert endpoint は、承認済みの宛先、data class、retention、削除条件に限ってデータを受け取れます。Support の質問・履歴・manual 抜粋・screenshot および bug report metadata への患者・飼い主情報・clinic-confidential data の含有は **owner 承認済み**（報告対象画面の特定に必要との裁定）。外部送信は既存経路に限定: LLM へは `screenChatOutbound` による best-effort スクリーニング付き（完全な検出は保証しない）、Plane へは reviewable な明示操作のみ。`backend/internal/support/chat.go:122-197` `backend/internal/support/llm.go:77-164` `backend/internal/support/plane.go:80-106` `backend/internal/support/plane.go:130-211`
7. **AUDIT-1 — path-dependent audit:** 必須と定めた credential、clinical、accounting 等の audit は業務 write と同一 transaction で fail closed にします。一方、すべての CUD が一律に `audit_logs` へ入るとは仮定せず、LSTEP tag sync のような明示的 fail-open・非 audit 経路と区別します。`docs/architecture/data-flow.md:67-101` `docs/architecture/data-flow.md:121-127`
8. **CONFIG-1 — release fail closed:** release mode は development defaults、loopback DB、非検証 DB TLS、weak/missing secrets、local storage、LIFF mock を拒否しなければなりません。secret、credential、個人情報、不要な内部 error を response、log、report、repository に出してはなりません。`backend/internal/config/config.go:216-338` `backend/internal/config/config.go:341-449`
9. **LIMIT-1 — 実 topology を守る上限:** body、response、upload、queue、connection、rate、cost の上限は実際の multi-instance topology を保護しなければなりません。process-local rate limit を shared/global quota と表現してはなりません。`backend/internal/middleware/rate_limit.go:17-42` `backend/internal/middleware/rate_limit.go:82-140`

## 確立済みコントロール

- Access token は 15 分、refresh token は最大 7 日で、refresh family/JTI rotation と reuse detection を持ちます。cookie は httpOnly です。`backend/internal/auth/token_service.go:19-25` `backend/internal/auth/http_session_cookies.go:25-115` `docs/architecture/auth.md:84-106`
- 保護された request は token 検証後に current account、staff、clinic assignment と account epoch を再解決し、lookup failure を fail closed にします。`backend/internal/middleware/auth.go:58-153` `backend/internal/middleware/auth.go:185-227`
- 共通 persistence helper は clinic predicate、parent join、scoped CRUD を提供します。ただし helper の存在だけで全 caller の利用を証明しません。`backend/internal/persistence/scope.go:15-136`
- PostgreSQL migration は RLS を ENABLE しますが FORCE せず、runtime GUC 配線もありません（2026-10-03 に `rls_effectiveness_test.go` で検証）。現在の load-bearing boundary は app/domain/persistence scope で、RLS は dormant な defense-in-depth baseline です。`backend/migrations/001_init.sql:2947-3043` `backend/internal/persistence/rls_effectiveness_test.go:3-28`
- API は exact-origin CORS、security headers、release HSTS、no-store response、trusted proxy と server timeout/header limit を設定します。`backend/internal/middleware/cors.go:9-47` `backend/internal/middleware/security_headers.go:7-24` `backend/cmd/api/main.go:186-225`
- Object upload は key を保存し、診療画像と support screenshot の取得時に signed URL を生成します。`backend/internal/infra/s3_uploader.go:48-88` `backend/internal/support/handler.go:267-277`
- LSTEP write は deploy-level exact-`true` gate と clinic setting の二重 gate を持ち、無効時は HTTP を送信しません。`backend/internal/infra/lstep/client.go:23-99`
- Local env、migration raw data、security output は Git 対象外で、PR CI は gitleaks を実行します。local hook は完全な scanner がない場合に成功扱いしません。`.gitignore:1-7` `.gitignore:121-143` `.github/workflows/ci.yml:144-165` `.githooks/lib/check-secrets.sh:1-17`

## Caller・実装者・運用者の義務

- Domain/API caller は request の ID ではなく、認証済み context の clinic authority と ownership を使わなければなりません。
- Assigned-clinic fallback を使う GET/HEAD は middleware 通過を最終認可とせず、返却対象 clinic ごとに filter または authorize しなければなりません。`docs/architecture/auth.md:152-169`
- 入力は HTTP boundary で型・長さ・形式を検証します。SQL は原則 parameter binding を使い、限定的な `Raw` も parameterized にします。client input を文字列連結した新規 SQL は禁止します。
- Secret は保護された環境変数または deployment secret store に置き、tracked frontend environment には公開可能な build-time 値だけを置きます。
- **[2026-10-03 確定]** Support の質問・本文・manual context・screenshot には患者・飼い主情報または clinic-confidential data を含めてよい（owner 裁定 — 報告対象画面の特定に必要）。コンテンツは全医院のスタッフに共有され、LLM（best-effort スクリーニング付き）・Plane（明示操作のみ）へ送信され得る。UI 警告は送信前確認の注意喚起であり、技術的な分類・redaction control ではありません。`frontend/src/features/support/components/BugReportTab.tsx:242-245`
- Operator workflow は、review 済み input、target host/database、revision/digest、backup 状態、出力 report を mutation 前後で拘束しなければなりません。
- Lab deployment は OS account 分離、screen lock、malware 対策、token file 保護、固定 serial wiring、physical-device UAT を維持しなければなりません。同一 OS user の process を consumer token が隔離するとは主張しません。`docs/architecture/adr/008-local-lab-device-agent.md:51-70`

## Deployment 前提

- STG は `workers_dev=true` の workers.dev route で **2026-10-03 に live 確認済み**です: `animalekarte-stg-frontend.baritech-soga.workers.dev`（200）、`animalekarte-stg-api.baritech-soga.workers.dev`（`/health`・`/health/db` 200、`/_internal/migrate` exitCode 0）。DB は PlanetScale Postgres `ap-northeast-2.pg.psdb.cloud`（live・TLS）。R2 は `animalekarte-stg-images`。cache が有効なら authority 変更の反映は最大 TTL 遅延します。`backend/wrangler.jsonc:28-97` `backend/cmd/api/composition_auth.go:173-184`
- Production Wrangler、frontend route、infra は draft です。GitHub Environment、branch protection、DNS、certificate、DB、R2、secrets、backup/restore の live state は外部の dated evidence が必要です。`backend/wrangler.production.jsonc:1-8` `frontend/wrangler.production.jsonc:1-52` `docs/ops/infra/architecture.md:1-65` `docs/ops/infra/production/setup.md:1-21`
- GitHub branch protection の実測（2026-10-03）: `staging` は protected（直接 push 拒否・PR 必須）、`main` は unprotected（直接 push 可）、`production` はブランチ存在・保護設定未検証。GitHub Environment `Production` は workflow で参照されますが required reviewers・branch 制限の設定は未検証です。Repo-level secrets は稼働確認済み（`MIGRATE_RUN_SECRET` 2026-10-03 ローテーション・3箇所同期、`STG_DB_*` 稼働中）。
- Per-IP limiter は各 Go process の memory 内です。trusted proxy と Worker の `CF-Connecting-IP` 転記、および必要な edge/shared limiter が正しく配備されることを前提とします。`backend/worker/index.ts:517-533`
- LLM、ticket、LSTEP write、scheduler alert は credential と gate に依存する条件付き surface です。checked-in configuration は live enablement や許可 data class の証拠ではありません。
- Lab agent は production code ですが、physical-device UAT と release gate は未完了と記録されています。`docs/architecture/adr/008-local-lab-device-agent.md:1-5`

## Reportable Findings と重大度

次の重大度基準を採用します（2026-10-03 owner 承認）。

Security invariant の破壊により、攻撃者が開始時に持たない identity、clinic、data、provider、schema、deploy authority を得る問題は報告対象です。Runtime 再現がなくても source が経路を独立に証明できれば報告できます。Deployment prerequisite が不明な場合は confidence・reachability を未確定とし、impact を自動的に引き下げたり finding を抑制したりしません。

| 重大度 | AnimalEkarte における目安 |
|---|---|
| Critical | 未認証または通常 staff による全 clinic compromise、live migration/scheduler control plane の奪取、広範な clinical datastore/object store の公開 |
| High | clinic A から clinic B の clinical/accounting data への access、LINE/LIFF identity bypass、private clinical object の広範な取得、通常 contributor による production deploy authority の取得 |
| Medium | 同一 clinic の sensitive-data exposure、外部 provider への限定的な不許可送信、必要 audit の欠落、process-local limit に起因する material abuse、限定された revocation delay |
| Low | 限定的 metadata exposure、error hardening、評価を誤らせる security documentation drift、別の特権 compromise を必要とする defense-in-depth weakness |

次の事実だけでは vulnerability を確定しません。

- Production draft の存在だけでは Internet exposure を証明しません。
- Dormant RLS だけでは app-layer clinic scope が正しい場合の tenant escape を証明しません。ただし RLS を実効 control と誤記することや、app-layer failure と組み合わさることは評価対象です。
- 攻撃者がすでに持つ正規権限内の通常動作は、新しい security impact ではありません。
- Support の閲覧・作成の clinic 横断共有は owner 承認済みの製品決定であり脆弱性ではありません（変更操作は報告元 clinic スコープ）。コンテンツへの患者・飼い主情報の含有も承認済み（2026-10-03）。保持期間は未定であり、期限超過・不許可削除経路は引き続き評価対象です。

## Scope 外、除外、受容済みリスク

- このレビューで owner が確認した新規の除外または受容済みリスクはありません。
- 人医療記録は製品の意図した用途外ですが、誤って保存された人医療・個人データの漏えいを報告対象外にはしません。
- Local uploads と UAT routes は条件付き development surface であり、Internet-facing と推測しません。ただし environment/Host guard の bypass は評価対象です。
- Lab Mac の同一 OS user は文書化された trust assumption ですが、今回 owner-confirmed exclusion にはしていません。Host、Origin、token、clinic、lease 境界の bypass は引き続き評価対象です。
- 依存パッケージ独自の報告窓口は AnimalEkarte の finding を除外しません。

## 既知の制限事項

- 2FA は未実装です。
- DB RLS は runtime enforcement ではなく、app-layer clinic scope が実効境界です。
- Rate limit は process-local であり、multi-instance 全体の quota ではありません。
- Audit は path-dependent で、すべての CUD を一律記録しません。
- Support の情報区分・外部送信は確定済み（2026-10-03、患者・飼い主情報の含有許可・既存経路への送信承認）。保持期間のみ未定で、当面現行のまま（期限超過・不許可削除経路は評価対象）。閲覧・作成の全医院共有と、変更操作の報告元医院スコープは確定済みです（2026-10）。
- `S3_PUBLIC_BASE_URL` を求める起動時コメントと、object key + presigned URL を使う実装が一致していません。公開 bucket を前提にしてはなりません。`backend/cmd/api/main.go:151-183` `backend/internal/infra/s3_uploader.go:24-76`
- Lab agent は Mac restart で未配送 memory queue を失い、端末 hardening と physical-device UAT に依存します。`docs/architecture/adr/008-local-lab-device-agent.md:61-70`

## 未決の owner 判断

2026-10-03 時点で確定済み: STG の live reachability（Deployment 前提に実測記載）、production は未 live でサポート対象外、Support の閲覧・作成共有/変更スコープの設計、OBJECT-1 の private+signed URL 標準、RLS dormant の実測、重大度基準、**Support の情報区分・外部送信（患者・飼い主情報の含有許可 — 画面特定に必要との裁定。LLM は best-effort スクリーニング付き送信、Plane は明示操作のみ）**。

残る決定事項:

1. **Support 共有コンテンツの保持期間**: clinic 横断共有される質問・履歴・bug report・screenshot の retention/deletion 条件。当面現行のまま（期限超過・不許可削除経路は評価対象）。
2. **Production go-live の前提状態**: GitHub Environment `Production` の required reviewers・deployment branch 制限、production branch 保護、DNS/certificate、prod DB/R2、環境別 secrets、**backup 取得・隔離 restore リハーサル（所要時間計測）**、lab rollout の検証日付き状態。go-live 時に production をサポート対象へ追加する。
3. **環境別の実効設定の最終確認**: R2 bucket の実効 public/private・lifecycle 方針、scheduler alert の有効状態、`S3_PUBLIC_BASE_URL` コメントと presigned 実装の不整合解消（コード・運用文書の整合）。

## セキュリティ更新の通知

GitHub の Watch → Custom → Security Advisories を有効にしてください。
