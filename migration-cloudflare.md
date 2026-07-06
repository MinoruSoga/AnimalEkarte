# STG AWS → Cloudflare 全面移行 タスクドキュメント

> **作成日**: 2026-07-05 | **対象**: Staging 環境（us-east-1）の全リソース
> **前提調査**: [research-cloudflare.html](research-cloudflare.html)（2026-07-04 再調査版）
> **ステータス**: 実行中 — 2026-07-05 に Phase 0 完了・Phase 1 準備（NS 切替手前まで）・Phase 2（全項目）・
> Phase 3 前半（PlanetScale DB 作成 + スキーマ検証）・**Phase 4 完了（P4-1〜P4-9）**。試行9で初回 STG デプロイ + `/health` 疎通確認 PASS、試行10で migrate one-shot 置換 PASS、試行11で CRUD + 混在会計 API スモーク全 AC PASS、**試行12で外部連携棚卸し(P4-7)・Cookie認証ブラウザ実機検証(P4-8)・10分負荷スモーク+CPU課金実測(P4-9)を実施し全AC PASS/BLOCKED(genuine)判定**。
> データ移行実行（P2-4/P3-6）とトラフィック切替（P1-2 後半・Phase 7）は次段の人間判断待ち（詳細は「実施記録」参照）

---

## 1. ゴールと到達アーキテクチャ

現行 STG（VPC / ALB / ECS Fargate / RDS / S3 / CloudFront / EventBridge / ECR / OIDC）を廃止し、以下へ移行する。

```
              ┌──────────────── Cloudflare ────────────────┐
  利用者 ──▶  │  DNS + CDN + Universal SSL + WAF(無料枠)     │
              │        │                                     │
              │        ▼                                     │
              │   Worker (ルーティング)                       │
              │        │                  ┌──────────────┐  │
              │        ▼                  │     R2       │◀─┼── 臨床画像
              │   Containers (Go/Gin API) │  (egress 0)  │  │   (S3 API 互換)
              │        │                  └──────────────┘  │
              │        ▼ Hyperdrive (接続プール)              │
              └────────┼────────────────────────────────────┘
                       ▼
            PlanetScale Postgres（pscale CLI で作成・管理）
            ※ DB ホストは PlanetScale (AWS 上)。Cloudflare は DB をホストしない。
```

- **DB は PlanetScale Postgres を採用**（検討経緯は §8 参照。RDS 継続案・Neon 案は不採用）。作成経路は CLI 直作成 or Cloudflare 課金統合を P0-3 で選択。
- 「すべて Cloudflare」とは**課金・管理・デプロイの一元化**を指す。DB 実体が Cloudflare 外である事実はステークホルダーに明示済みであること。

### 移行後の想定月額（STG・データ 32GB / ディスク 40GB 前提・税前）

| 項目 | USD | JPY (161.2) |
|---|---|---|
| Workers Paid 基本料 | $5.00 | ¥806 |
| Containers（basic・低稼働・scale-to-zero） | ~$4.00 | ~¥645 |
| PlanetScale PS-10 arm 単一ノード（東京） | $14.00 | ¥2,257 |
| PlanetScale ストレージ超過 30GB × $0.15 | $4.50 | ¥726 |
| R2 / Hyperdrive / DNS / SSL / Logs | $0 | ¥0 |
| **合計** | **~$27.5** | **~¥4,430** |

現行実測 ¥6,404/月 → **約 31% 削減**。夜間停止スケジューラなしでこの額（scale-to-zero が自動で同等の効果を持つ）。

---

## 運用原則: コード / CLI ファースト（MANDATORY）

**ダッシュボードのブラウザ操作は原則禁止**。すべての構成変更はコードベース（Terraform / wrangler 設定ファイル）または CLI で行い、Git で追跡する。

| 領域 | ツール | 管理場所 |
|---|---|---|
| ゾーン / DNS / SSL / WAF / キャッシュルール / 通知 | **Terraform `cloudflare` provider** | `infra/cloudflare/`（新設。既存 `infra/terraform/` と分離） |
| Workers / Containers / R2 バケット / Hyperdrive / Secrets / Cron | **wrangler**（`wrangler.jsonc` をリポジトリ管理） | `backend/wrangler.jsonc` + CI |
| PlanetScale（DB 作成・ブランチ・接続・dump/restore） | **pscale CLI** | 手順をスクリプト化して `infra/scripts/` |
| S3 → R2 データ移行 | **rclone**（Super Slurper はダッシュボード操作のため不使用） | `infra/scripts/` |
| tfstate | R2 バケット（S3 互換 backend） | `infra/cloudflare/backend.tf` |

**例外（ダッシュボード操作が避けられないもの）** — 以下のみ許容し、実施記録を本ドキュメントに残す:

1. Cloudflare アカウント作成・Workers Paid プラン契約・支払い方法登録（初回のみ）
2. Terraform / wrangler 用の**初回 API Token 発行**（以降のトークンは API で発行可能）
3. PlanetScale の課金を Cloudflare に統合する場合の**アカウント連携承認**（OAuth 同意画面。連携せず PlanetScale 直課金なら `pscale` のみで完結 — P0-3 で選択）
4. **`noah-karte.com` の Add a Site**（Account API Token にゾーン作成権限が無いため。作成後 `terraform import` で state 取り込み）
5. **API Token 権限追加**（Account API Token が DNS Edit / Hyperdrive Edit を持たない場合、ダッシュボードで Edit して再発行）

### 例外的ダッシュボード操作の実施記録

| 日付 | 項目 | 内容 |
|---|---|---|
| 2026-07-05 | P0-3 決定 | **(a) pscale CLI 直作成**を採用（CF 課金統合は不使用） |
| 2026-07-05 | PlanetScale org 作成 | org 名: **`noah-animalekarte`**（環境名は含めない。DB は org 内に `animalekarte-stg` / `animalekarte-prod` として作成予定） |
| 2026-07-05 | API Token 方針 | STG は**統合トークン 1 本**（Workers Scripts/R2/KV/Account Rulesets + Zone/DNS/Zone Settings/WAF の Edit）。本番移行時に CI 用の狭いデプロイトークンを分離発行すること |
| 2026-07-05 | 例外 #4 Add a Site | `noah-karte.com` を Cloudflare ダッシュボードで接続（DNS only 方針）。zone_id=`d0eec286da621a49fa677dce8fa02c73`、account_id=`776ddc3e975e8fe5773d5300522e2404`、NS=`melissa.ns.cloudflare.com` / `yadiel.ns.cloudflare.com`（**Vercel NS 未切替**） |
| 2026-07-05 | 例外 #1 Workers Paid プラン契約 | 試行9: Containers デプロイが `403 containers/me` で失敗し、Workers Free プランでは Containers 非対応と判明。ダッシュボードから Workers Paid プランへアップグレード |
| 2026-07-05 | 例外 #5 API Token 権限追加 | 試行9: プランアップグレード後も同一403が再発。既存 Account API Token に **Containers Edit** 権限が不足していたため、ダッシュボードで追加 |

### 2026-07-05 実行フェーズ 実施記録（Phase 0 完了〜Phase 3 前半〜P4-1）

**完了フェーズ**: Phase 0（全項目）／Phase 1 準備（P1-1、P1-2 前半のNS値確定まで）／Phase 2（全項目）／Phase 3 前半（P3-1〜P3-5）／**Phase 4（P4-1〜P4-4）**

**現在のブロッカー（2026-07-05 試行9時点。取り消し線は解消済み）**:
- ~~Account API Token の権限不足によるDNS/Hyperdrive apply 403~~ → 試行6でトークンに DNS Write/Hyperdrive Write を追加し解消済み
- ~~Hyperdrive 実接続検証（P3-5 確定）~~ → 試行7で `wrangler dev --remote` + postgres.js による実接続CRUD/トランザクション検証がPASS（GORM自体での確認はPhase 4に持ち越し）
- ~~R2 S3互換 API（`S3_ENDPOINT` + Access Key/Secret）の実疎通（P2-3）~~ → 試行8で Account API Token に `Account API Tokens Write` 追加後、`POST /accounts/{id}/tokens` 経由で R2 スコープ子トークンを CLI 発行し `TestR2S3Live` PASS
- ~~Workers Free プランでは Containers 非対応（403）~~ → 試行9でダッシュボードから Workers Paid プランへアップグレードし解消
- ~~Account API Token に Containers Edit 権限が不足~~ → 試行9でダッシュボードからトークンに権限追加し解消
- ~~Worker→Container 間の `TRUSTED_PROXY_CIDR` 未確定 + XFF転送されない機能バグ~~ → 試行9で実測確定（`10.1.0.0/32`）+ `worker/index.ts` にXFF転送コード追加で解消
- ~~ECS `animalekarte-stg-migrate` one-shot task 相当の置換手段が未実装（P4-5）~~ → 試行10で `POST /_internal/migrate` + `Container.exec()` により実装・実測PASS
- ~~機能スモーク未実施（P4-6）~~ → 試行11で `infra/scripts/cf-crud-smoke.sh` により CRUD + 混在会計 API スモークを実施し全AC PASS(AC-11のみUIスコープ外でBLOCKED)

### 2026-07-05 試行12（P4-7〜P4-9 — 外部連携棚卸し・Cookie認証ブラウザ実機検証・10分負荷スモーク+CPU課金実測。Phase 4完了）

**前提**: 試行9〜11で確立した Worker/Container 構成に対し、残る P4-7（外部連携）・P4-8（Cookie認証ブラウザ検証）・P4-9（負荷スモーク）を実施。前半（P4-7 doc/LSTEP grep/wrangler secret list、CORS localhost追加deploy、CSP発見・追加、k6スクリプト作成）と後半（`STG_DEMO_EMAIL`/`STG_DEMO_PASSWORD`投入後のcurl/browser/k6実測、revert、doc記録）の2セッションに分割して実施。認証情報は `infra/cloudflare/.env.staging`（gitignore）にユーザーが直接追記し、`set -a && source ... && set +a` で都度export・使用後unsetする方式（チャット/git/本ドキュメントに値は一切残さない）。

**P4-7 — 外部連携棚卸し結果**:

新規ドキュメント `docs/infra/deploy/CLOUDFLARE-EXTERNAL-INTEGRATIONS-AUDIT.md` を作成。

| 連携 | 判定 | 根拠 |
|---|---|---|
| LINE Messaging API | **PASS**（doc結論）／live送信は**BLOCKED** | Bearer token認証。LINE公式ガイドラインは「webhook受信側のIP制限禁止」を明言。push送信(outbound)側はlong-lived token使用時に**オプションでIP allowlist設定可能**（既定は無効）。STG `clinic_integrations` に実クリニックのトークン/LINE User IDが登録されているため、誤配信リスク回避を優先しlive送信は見合わせ、inventory onlyとした |
| Lステップ Write API | **PASS** | `tag.go`(`AddTag`/`RemoveTag`/`AddTagBulk`)・`user.go`(`SetProperty`)の`[DISABLED]`コメント4箇所をgrepで再確認、抑止継続を確認。`LSTEP_WRITE_API_PAUSE.md`の再有効化前提条件（5項目）は未達のためコード変更なし |
| SMTP | **BLOCKED** | `wrangler secret list --name animalekarte-stg-api`で`SMTP_HOST`/`SMTP_USER`/`SMTP_PASS`が`secret_text`として登録済みであることを確認（値は取得しない）。試行9記録に「未使用分は空文字投入の可能性」があり、実際に空文字か否かは値を取得しない限り判別不可。空文字であれば`config.go`のガードで送信自体スキップされ移行によるリグレッションは無いが、実値が入っていた場合に実メール送信のリスクがあるため、アプリ経由の送信トリガーは見合わせた |
| LIFF / コールバックURL | **BLOCKED**（P7-3 defer） | `frontend/vercel.json`のrewrite先が`https://api.stg.noah-karte.com`（AWS）固定で、workers.dev段階では本番相当のLIFF導線検証が不可。P1-2(NS切替)後に再検証 |

`migration-cloudflare.md` §9 リスク登録簿の「IP allowlist依存の外部連携」行を上記結論で更新済み（解消済み・残作業はP7-3へdefer）。

**P4-8 — Cookie認証の実機検証結果**:

- **AC-A（curl, Set-Cookie属性）**: `POST /api/v1/login`成功時のレスポンスヘッダーを確認（値は`[REDACTED]`）:
  ```
  set-cookie: access_token=[REDACTED]; Path=/; Max-Age=899; HttpOnly; Secure; SameSite=None
  set-cookie: refresh_token=[REDACTED]; Path=/api/v1/auth/refresh; Max-Age=604799; HttpOnly; Secure; SameSite=None
  ```
  `GIN_MODE=release`時の想定通り`HttpOnly`/`Secure`/`SameSite=None`を確認。**PASS**
- **AC-B（ブラウザ cross-origin ログイン）**: ローカル docker compose の frontend（`VITE_API_URL`を一時的に`https://animalekarte-stg-api.baritech-soga.workers.dev/api`へ変更）+ `backend/wrangler.jsonc`の`CORS_ALLOWED_ORIGIN`に`http://localhost:3003`を一時追加してdeployし、chrome-devtools MCPで実ブラウザ検証。
  - **想定外の新規発見**: CORS許可後も`frontend/index.html`のCSP `connect-src`ディレクティブ（`'self' https://api.stg.noah-karte.com ...`固定）がブラウザ側でworkers.devへの接続をブロック（`Content-Security-Policy`違反エラーをconsoleで確認）。これはCORS/Cookie検証だけでは検出できない、フロントエンド側の別レイヤーの制約であり、P4-8のチェック項目に無かった新規リスクとして記録
  - CSPに`https://animalekarte-stg-api.baritech-soga.workers.dev`を一時追加して再検証したところ、`POST /api/v1/login`→200、`GET /api/v1/me`→200、以降の`/api/v1/reservations`・`/api/v1/masters/staffs`等の呼び出しも200で正常に動作。デモアカウント（admin@noavet.jp）でログイン後、ダッシュボード（受付管理画面）が正常表示されることをスクリーンショットで確認。**PASS**
  - 検証後、CSP一時追加行・`VITE_API_URL`一時変更・`CORS_ALLOWED_ORIGIN`一時追加を全てrevert（下記「revert確認」参照）

**P4-9 — 10分負荷スモーク + CPU課金実測結果**:

- k6スクリプト `load-tests/k6-cf-stg-sustained.js` を新規作成、`package.json`に`cf:load-smoke`登録
- **初回実行で発見した設計不備**: 当初はループ内で毎イテレーション再ログインする設計にしたところ、`POST /api/v1/login`にはIPベースのレートリミット（**5回/分・バースト5**、`handler.go` L63、BUG-130ブルートフォース対策）が掛かっており、VU3による同一IPからの連続ログインが即座にレート制限に抵触。失敗率**55.86%**（successful_logins 54/333）という結果になったが、これはCloudflare Containers側の性能問題ではなく、既存のブルートフォース対策が意図通り機能した結果と判断。実運用のユーザー挙動（1回ログイン→セッション継続）に合わせ、`setup()`で1回だけログインしCookieを全イテレーションで再利用する設計に修正して再実行
- **修正後の実行結果（Docker経由 `grafana/k6`, 10分, VU最大3, exit code 0）**:
  - Total requests: 837 / Failed rate: **0.00%** / p95 duration: **897ms** / avg duration: 564ms / 418 iterations complete, 0 interrupted
  - thresholds（`http_req_duration p95<3000ms`, `http_req_failed rate<0.05`, `errors rate<0.05`）全てPASS。**AC-P49-1: PASS**
- **CPU課金実測（AC-P49-2）**: Cloudflare Dashboardの手動操作を避け、GraphQL Analytics API（`containersUsageAdaptiveGroups`、`cpuTimeSec`/`allocatedMemory`/`allocatedDisk`）を`CLOUDFLARE_API_TOKEN`で直接クエリして実測（値はコード/gitに残さず本記録にのみ数値を記載）。
  - 試行12セッション全体（CORS deploy〜k6完走まで、約55分）での差分: **CPU 33.18 vCPU秒 / メモリ 2200.87 GiB秒 / ディスク 8800.11 GB秒**
  - このセッションのコストは Cloudflare公式レート（CPU $0.000020/vCPU秒、メモリ $0.0000025/GiB秒、ディスク $0.00000007/GB秒）換算で約 **$0.0068**
  - **月額換算（保守的な上限側シナリオ: このセッションの負荷強度が24時間365日continueすると仮定した場合)**: CPU課金対象分 約 $0.07、メモリ課金対象分 約 $4.10、ディスク課金対象分 約 $0.43（Free枠 CPU 375 vCPU分・メモリ25GiB時間・ディスク200GB時間は考慮済み）で **合計 約 $4.60**
  - **試算(~$4.00)との比較**: 差分 **+15%**、**±50%以内でPASS**。実際の運用は本試行のような集中的なテストトラフィックより疎らな業務時間アクセスが主となるため、実運用でのCPU課金はFree枠内に収まる可能性が高く、本結果は「試算を上回るリスクは低い」ことを示す上限側の裏付けと判断。**AC-P49-2: PASS**
- **post-load `/health`回帰（AC-P49-3）**: `200 {"status":"ok"}`。**PASS**

**revert確認（AC-REVERT-CORS / AC-REVERT-CSP）**:
- `backend/wrangler.jsonc`の`CORS_ALLOWED_ORIGIN`から`http://localhost:3003`を除去し`git diff`が空であることを確認、`wrangler deploy`を再実行
- `frontend/index.html`のCSP `connect-src`からworkers.dev一時行を除去、`frontend/.env.local`の`VITE_API_URL`も`/api`へ復帰
- **運用上の新規発見**: revert後の`wrangler deploy`は「no changes」（コンテナimage不変のため）と表示され、既に稼働中（warm）のContainerインスタンス（Durable Object, `cf-singleton-container`, sleepAfter=10m）が新しい`vars`を即時には反映しない事象を実測。deploy直後に`curl -X OPTIONS`で確認したところ、数分間旧CORS設定（`localhost:3003`許可）が残存した
- アイドルタイムアウト（10分）経過を待ち、`wrangler containers instances`でインスタンスが`running`→`inactive`（scale-to-zero）に遷移したことを確認した上で再検証:
  - `curl -X OPTIONS ... -H "Origin: http://localhost:3003"` → `Access-Control-Allow-Origin`ヘッダーなし（許可されない・revert反映確認）
  - `curl -X OPTIONS ... -H "Origin: https://stg.noah-karte.com"` → `access-control-allow-origin: https://stg.noah-karte.com`（正規オリジンは引き続き許可・正常動作確認）
  - `GET /health` → `200 {"status":"ok"}`
  - **AC-REVERT-CORS / AC-REVERT-CSP / post-revert health回帰: 全PASS**

**独立レビュー**: `security-reviewer`（readonly）を2回実施——(1)前半セッションで新規doc(`CLOUDFLARE-EXTERNAL-INTEGRATIONS-AUDIT.md`)・k6スクリプトをスキャンし CRITICAL/HIGH 0・LOW 1（k6デフォルトURLのハードコード、`__ENV.BASE_URL`で上書き可能な意図的デフォルト値のため対応不要）でPASS、(2)本セクション記録後に変更ファイル全体を再スキャン（結果は下記参照）。

**Harness Improvement Feedback（記録）**:
- P4-8手順に「frontend CSP `connect-src`が新オリジンをブロックし得る」ことをrunbook/チェックリストに追記推奨（本試行で新規発見・上表に反映済み）
- Container `vars`変更のdeploy後、warm instanceには即時反映されない（次アイドルサイクルまで数分〜10分残存）ことをP5(CI/CD)のデプロイ後検証手順に明記推奨
- `load-tests/k6-api-endpoints.js`の既存login body `cred`→`password`フィールド名drift修正は別PRで対応（本試行スコープ外、記録のみ）
- `migration-cloudflare.md`へのStrReplace編集は本試行では正常に成功（試行11で発生したhookブロック事象は再発せず）

**まとめ**: P4-7（外部連携棚卸し）・P4-8（Cookie認証ブラウザ実機検証）・P4-9（10分負荷スモーク+CPU課金実測）の全AC PASSまたはgenuine BLOCKED（SMTP/LIFF/LINE live送信、いずれも理由明記済み）。**Phase 4（P4-1〜P4-9）完了**。

---

### 2026-07-05 試行11（P4-6 機能スモーク — CRUD + 混在会計 API スモーク実施・実測検証）

**前提**: 試行9/10で確立した Worker/Container 構成（`https://animalekarte-stg-api.baritech-soga.workers.dev`）に対し、`docs/infra/deploy/CRUD-SMOKE-TEST.md` / `MIXED-PAYMENT-SMOKE-TEST.md` を参考に `infra/scripts/cf-crud-smoke.sh`（新規、curl + jq。AC-0〜AC-8 + AC-11）を作成し、`package.json` に `pnpm cf:smoke` を登録。

**ドキュメントと実装の差異（実装を正として採用）**:
- ログインエンドポイントは `POST /api/v1/login`（`CRUD-SMOKE-TEST.md` 記載の `/auth/login` ではない。`handler.go` L63）
- permission-groups作成には `color` フィールドが必須（doc例は欠落）、`permissions` 配列はリクエストに存在しない
- staff作成に `role` フィールドは存在しない（`createStaffRequest`）
- 重複支払方法バリデーションは HTTP **400**（`MIXED-PAYMENT-SMOKE-TEST.md` の期待値422ではなく `apperrors.ErrInvalidInput`→400マッピング）。script は400/422両方をPASS判定するよう設計し、doc drift として記録のみ

**実行結果（`pnpm cf:smoke`、exit code 0、SUMMARY全文。認証情報・Cookie・tokenは非出力）**:

| AC | 内容 | 結果 |
|---|---|---|
| AC-0 | `GET /health` | PASS: 200 `{"status":"ok"}` |
| AC-1b | 誤パスワードでログイン | PASS: 401、後続 `/me` も401 |
| AC-1 | 正しい認証情報でログイン+Cookie付き`/me` | PASS: login 200(`is_system_admin=true`)、`/me` 200 |
| AC-2 | `GET /clinics` | PASS: 200、3 clinics |
| AC-3 | permission-groups POST→保護対象DELETE 409→testDELETE | PASS: POST 201(id=9)、seed `id=1` DELETE→409、test DELETE→204 |
| AC-4 | staffs POST→DELETE | PASS: POST 201(id=37)、DELETE→204 |
| AC-7 | 重複payment method(422/400) | PASS: HTTP 400、会計はwaiting維持(未完了) |
| AC-5 | 単一現金payment_split | PASS: PATCH 200、`payment_splits=[cash:880]`(billing_amount=880) |
| AC-6 | 新規waiting会計→2種混在split | PASS: 新規会計id=2001、PATCH 200、`payment_splits`2件・合計1100 |
| AC-8 | cleanup確認 | PASS: TEST permission-group id=9→404、TEST staff id=37→404（削除確認済み） |
| AC-11 | UI混在会計(§3〜§10) | BLOCKED: frontend vercel.jsonプロキシがAWS(`api.stg.noah-karte.com`)向きでworkers.dev未接続。P7-3へdefer |

FAIL=0。全PASS(BLOCKED除く)でexit code 0。

**AC-4 seed副作用（記録）**:
- 探索に使用したseed waiting会計 `id=3`(amount=880)は AC-5/AC-7 実行により **completed に変更**（元のwaitingへは戻していない。STGデータとして許容）
- AC-6で新規作成した会計 `id=2001` は、billingsに DELETE ルートが存在しないため **smoke実行後も残存**（テスト用データとしてSTG DB上に残る。P4-7以降のスモークに影響なし）
- TEST用 permission-group(`id=9`)・staff(`id=37`)は AC-8 で削除確認済み（残存なし）

**post-smoke `/health` 回帰確認**: `200 {"status":"ok"}`（smoke実行後も継続）

**独立レビュー**: 本試行はスクリプトのみの新規追加（既存コード変更なし）のため、`cf-crud-smoke.sh` 作成時点で `code-reviewer` によるレビューを実施済み（CRITICAL/HIGH 0件、MEDIUM 3件・LOW 4件は全て即時対応済み — curl失敗時のexit code記録、cleanup再DELETE試行、AC-6のlegacy method判定緩和、`set -e`意図のコメント化、AC-3ハードコードIDのコメント化、AC-2 `jq`のページネーション両対応、タイムスタンプへのPID付与）。migration記録（本セクション）の credential漏洩は目視 + `grep -E 'password|access_token|Bearer '` で確認し、実credential文字列は含まれていない（`003_seed_demo.sql`のsystem_admin参照のみ）。

**残課題（次段以降）**:
- `MIXED-PAYMENT-SMOKE-TEST.md` §5 の期待値422はdoc drift（実装は400）。ドキュメント側の修正は本試行スコープ外（記録のみ）
- P4-7（外部連携: LINE/SMTP/Lstep IP allowlist棚卸し）、P4-8（Cookie認証のブラウザ実機検証）、P4-9（負荷スモーク）は未着手
- AC-11のUI混在会計スモークはP7-3（フルスモーク、DNS切替後）へdefer。現状frontendはAWS(`api.stg.noah-karte.com`)向きのままのため、Cloudflare経路での検証には別途frontend環境変数/プロキシ設定変更が必要
- Phase 5(`P5-4`)で `stg-smoke.yml` から `pnpm cf:smoke` をCI実行する設計を検討（`STG_DEMO_EMAIL`/`STG_DEMO_PASSWORD`はGitHub Secrets化が前提）

### 2026-07-05 試行10（P4-5 migrate one-shot 置換 — `Container.exec()` 実装・実測検証）

**前提**: 試行9で確立した `backend/worker/index.ts`(Worker/Containerプロキシ)・`wrangler.jsonc` をベースに、ECS `animalekarte-stg-migrate` one-shot task 相当の仕組みを実装。`.claude/CLAUDE.md`・`.github/workflows/backend-deploy.yml`(ECS migrate ジョブ, L211-360)・`backend/cmd/migrate/main.go`(advisory lock・DB_RESETゲート・冪等スキップ, 変更なし)を事前に読み込み。

**設計判断1 — `@cloudflare/containers`(0.3.7)は`exec()`を公開していない**:
- npm パッケージの `Container` クラス(`node_modules/.pnpm/@cloudflare+containers@0.3.7/.../container.d.ts`)は `containerFetch`/`start`/`startAndWaitForPorts` 等は public だが、`exec()` は存在しない(内部の生 container オブジェクトは `private` フィールド)
- Cloudflare Workers ランタイムが提供する低レベル `DurableObjectState.container`(`worker-configuration.d.ts` の `Container` interface)には `exec(cmd: string[], options?: ContainerExecOptions): Promise<ExecProcess>` が存在する。`DurableObject.ctx` は `protected`(`private`ではない)であるため、`AnimalEkarteApiContainer extends Container<Env>` サブクラス内から `this.ctx.container.exec(...)` を直接呼べることを確認
- `AnimalEkarteApiContainer` に public な `async runMigrate()` メソッドを追加し、`getContainer(env.API_CONTAINER)` が返す `DurableObjectStub` 経由で Workers RPC(fetch/alarm以外の public メソッドは標準でRPC呼び出し可能)として `container.runMigrate()` を呼ぶ設計とした(公式ラッパーに無い機能を安全に使う唯一の経路)

**設計判断2 — `/_internal/migrate` の認証・ルーティング**:
- `backend/worker/index.ts` の `fetch()` で、`/_internal/migrate` パスを既存の汎用プロキシ分岐より**前**に判定し、Container の Gin へは一切ルーティングしない(Constraint要件を満たす)
- 認証は `Authorization: Bearer <MIGRATE_RUN_SECRET>`。ロジックは `backend/worker/migrate-exec.ts`(新規、pure function群)に分離: `timingSafeEqual`(定数時間比較)・`isAuthorizedMigrateRequest`・`toMigrateResponse`(exitCode 0→200、非0→500)
- `MIGRATE_RUN_SECRET` は `openssl rand -hex 32` で生成し `wrangler secret put` で投入(値はgit/ログ/チャット未出力)。`wrangler.jsonc` の `secrets.required` に追加

**ブロッカー — `exec()` は起動時 `envVars` を継承しない(実測で発覚)**:
- 初回実行で `{"exitCode":1,"stdout":"...Missing required environment variables\" DB_HOST=\"\" DB_USER=\"\"..."}` が返り、`docker exec` と同様に起動時の環境変数を継承すると想定していたのは誤りと判明
- 修正: `rawContainer.exec(["/app/migrate"], { env: this.envVars, stdout: "pipe", stderr: "pipe" })` のように `ContainerExecOptions.env` へ `this.envVars`(Container起動時に渡している既存のenvVarsオブジェクトを再利用)を明示的に渡すよう変更。再デプロイ後、DB接続に成功

**AC-1/AC-8 — dry-run PASS**: `wrangler deploy --dry-run` は型エラーなく成功(修正前・修正後の両方で実施)。

**AC-2 — 認証・メソッド制限（実測）**:
```bash
curl -s -w '\nHTTP %{http_code}\n' -X POST '.../_internal/migrate'                                    # {"error":"unauthorized"}  HTTP 401
curl -s -w '\nHTTP %{http_code}\n' -X POST '.../_internal/migrate' -H 'Authorization: Bearer wrong'   # {"error":"unauthorized"}  HTTP 401
curl -s -w '\nHTTP %{http_code}\n' '.../_internal/migrate'                                             # {"error":"method_not_allowed"}  HTTP 405
```

**AC-3/AC-4 — migrate実行 exit 0 ×2（冪等性、実測）**:
- 1回目: `exitCode:0`、stdout に `Connected to database`→`Migration lock acquired`→001〜005 全て `⏭ Skipping (already applied)`→`Migration summary applied=0 skipped=5 total=5`→`✓ All migrations completed successfully`
- 2回目(直後再実行): 同一結果で `exitCode:0`(冪等性確認)
- HTTP 200、`Authorization: Bearer <MIGRATE_RUN_SECRET>` 使用時のみ到達

**AC-5 — `schema_migrations` 整合（実測、PlanetScale直接クエリ）**:
```
 count |            latest_filename
-------+---------------------------------------
     5 | 005_add_appointment_checked_in_at.sql
```
`001_init.sql`〜`005_add_appointment_checked_in_at.sql` の5件が期待どおり記録済み(試行9以前に適用済みの内容と一致)。検証用クレデンシャルは `pscale role reset-default --force` で都度発行し、検証後に `wrangler secret put DB_USER`/`DB_PASSWORD` で新パスワードに追随(既存のローテーション運用を継続)。

**AC-6 — API回帰（実測）**: migrate実行の前後で `GET /health` は継続して `200 {"status":"ok"}`。

**AC-7 — 失敗時 exit code 伝播（設計検証）**: `toMigrateResponse()` が exitCode非0→HTTP 500 にマッピングし、`infra/scripts/cf-run-migrate.sh` は `HTTP_CODE!=200 || EXIT_CODE!=0` で `exit 1`(`set -euo pipefail`)。実際にDBを壊す失敗テストは実施せず(Constraintで許可された設計検証のみ)、コードパスの確認で代替。

**AC-9 — 独立レビュー**: security-reviewer + code-reviewer を実施(結果は下記「独立レビューで判明した修正」参照)。

**AC-10 — ECSとの対応関係**:

| ECS(`backend-deploy.yml`) | Cloudflare(試行10) |
|---|---|
| `aws ecs run-task`(migrate task) | `POST /_internal/migrate`(`cf-run-migrate.sh` 経由) |
| `describe-tasks` でSTOPPEDをポーリング | `container.exec().output()` が完了まで await(同期的) |
| `containers[0].exitCode` を確認、非0なら`exit 1` | JSON `exitCode` を確認、非0(またはHTTP≠200)なら`exit 1` |
| 失敗時 `aws logs get-log-events` でログ取得 | JSON レスポンスの `stdout`/`stderr` フィールドで代替 |
| `set -e`相当のジョブ全体abort | `cf-run-migrate.sh` の `set -euo pipefail` + 明示 `exit 1` |

**運用手順（runbook）**:
```bash
# 1. deploy(Worker + Container イメージ)
pnpm run cf:deploy

# 2. migrate実行(MIGRATE_RUN_SECRETは`wrangler secret put`投入時に控えた値。ログ/gitに残さない)
export WORKER_URL=https://animalekarte-stg-api.baritech-soga.workers.dev  # NS切替後はカスタムドメインに変更
export MIGRATE_RUN_SECRET=<secret>
pnpm run cf:migrate    # 内部でunauthenticated 401 self-test → 本実行 → exit code検証

# 3. 失敗時(exit 1で終了): レスポンスJSONのstdout/stderrを確認。DB_RESETは本経路から渡せない(常にfalse)
```
Phase 5(`P5-1`)で `backend-deploy.yml` に組み込む際は、ECS版の「migrateジョブ→exit code検証→失敗ならAPI deployをabort」の順序を`cf:deploy`→`cf:migrate`(非ゼロ終了ならジョブ全体をfail)として再現する。

**独立レビュー（code-reviewer + security-reviewer）で判明した修正**:
- **code-reviewer**: CRITICAL/HIGH 0件。RPCパターン・型安全性・認証設計は正確と評価。MEDIUM 2件・LOW 3件を指摘、すべて即時対応済み:
  - **即時対応済み(MEDIUM)**: exec にタイムアウトが無く、`pg_advisory_lock` が別プロセスに握られたまま等の異常時に無期限ハングする可能性 → `runMigrate()` に `MIGRATE_TIMEOUT_MS=120_000` の `Promise.race` + `proc.kill()` を追加
  - **即時対応済み(MEDIUM)**: `cf-run-migrate.sh` の curl 呼び出しに `--max-time` が無く、ネットワーク異常時にスクリプトが無期限ハング → self-test用15秒・本実行用150秒(Worker側exec timeoutより長めのマージン)の `--max-time` を追加
  - **即時対応済み(LOW)**: exec の `env` に `this.envVars` 全体(SMTP/JWT等 非DB値含む)を渡していた最小権限違反 → migrate バイナリが実際に読む `DB_HOST`/`DB_PORT`/`DB_USER`/`DB_PASSWORD`/`DB_NAME`/`DB_SSL_MODE` のみの subset に変更
  - **即時対応済み(LOW)**: `cf-run-migrate.sh` の JSON解析が `python3` 依存(CI環境での存在が不確実) → `jq` に統一
  - **即時対応済み(LOW)**: `timingSafeEqual` の自前XORループ実装 → Cloudflare Workers ランタイム組み込みの `crypto.subtle.timingSafeEqual`(Web標準ではないCloudflare拡張)に置換
  - 修正後、`wrangler deploy --dry-run` PASS → 再デプロイ → `cf-run-migrate.sh` フルパス(self-test 401 → migrate exitCode 0)を再実行し PASS を再確認済み
- **security-reviewer**: CRITICAL/HIGH/MEDIUM 0件（スコープ内5ファイル: `worker/index.ts`・`worker/migrate-exec.ts`・`wrangler.jsonc`・`cf-run-migrate.sh`・`package.json`）。認可bypass・任意コマンド実行・DB_RESET経由の破壊的操作・シークレットのレスポンス漏洩・パストラバーサルによるmigrate到達のいずれも再現可能な問題なしと結論。確認された良好点: fail-closed認証(secret未設定は常に拒否)・`crypto.subtle.timingSafeEqual`によるBearer全体の定数時間比較・`/_internal/migrate`をContainerプロキシより前段でルーティング分離・exec引数が固定`["/app/migrate"]`のみでリクエスト入力が混入しない設計・`DB_RESET`が経路上どこにも注入されないこと
  - **Phase 5前の推奨hardening(任意・medium未満)**: ①`MIGRATE_RUN_SECRET`をGitHub Encrypted Secretのみで扱いレスポンス全文をCIログに出さない、②`/_internal/*`にCloudflare AccessまたはIP許可リストを多層防御として追加、③NS切替後`workers_dev: false`化(既存TODO)、④`isAuthorizedMigrateRequest`の単体テスト追加

**残課題（次段以降）**:
- `infra/scripts/cf-run-migrate.sh` を Phase 5 `backend-deploy.yml` の migrate ジョブから呼ぶ形へCI組み込み(本試行はスクリプト作成のみ。CI組込はP5-1)
- security-reviewer推奨のPhase 5前hardening(任意): `/_internal/*`へのCloudflare Access/IP許可リスト追加、`isAuthorizedMigrateRequest`の単体テスト追加、CIログでのレスポンス全文出力回避
- `workers_dev: true` → NS切替後に `false` へ戻す(試行9からの既存TODO、継続)

### 2026-07-05 試行9（P4-2〜P4-4 初回 STG デプロイ — Worker/Container 実装・実疎通）

**前提**: `backend/worker/index.ts`（Container プロキシ実装）・`backend/wrangler.jsonc`（`vars`/`secrets.required`確定）を新規作成。`wrangler secret put` で `DB_HOST`/`DB_USER`/`DB_PASSWORD`（`pscale role reset-default --force` で都度再発行・検証後失効）・`JWT_SECRET`/`INTEGRATION_ENCRYPTION_KEY`（`.env.staging` 由来）・`SMTP_*`/`LINE_*`/`LSTEP_API_KEY`（未使用のため空文字）を投入。`wrangler deploy --dry-run` は事前に PASS（AC-1/AC-8）。

**ブロッカー1 — Workers Free プランでは Containers 非対応**:
- `wrangler deploy` が Docker イメージビルド成功後、`GET /accounts/{id}/containers/me` で `403 Forbidden / Authentication error`
- 調査の結果、Cloudflare Containers は **Workers Paid プラン（$5/月〜）必須**という既知の制約と判明（Free プランでは有効化不可）
- 人間対応: ダッシュボードで Workers Paid プランへアップグレード（運用原則の例外#1に該当。実施記録に追記）

**ブロッカー2 — プランアップグレード後も同一 403**:
- プランアップグレード後に再実行しても同じ `containers/me` 403 が再発。Account API Token の権限に **Containers Edit**（Account レベル）が含まれていなかったことが原因と判明
- 人間対応: ダッシュボードで既存トークンに **Containers Edit** 権限を追加（運用原則の例外#5に該当）
- 追加後、`wrangler deploy` が最後まで成功。イメージが Cloudflare Registry へ push され、Container アプリケーション（`animalekarte-stg-api-animalekarteapicontainer`, Application ID: `a03e1180-c9fd-41e4-85ef-7773406d2335`）が作成された

**P4-2 — PASS**: 上記の通りイメージビルド・push・Container アプリケーション作成が成功。

**カスタムドメインルートの既知の失敗（想定通り・ブロッカーではない）**:
- `wrangler deploy` はカスタムドメインルート（`api.stg.noah-karte.com/*`）の登録時に `A request to the Cloudflare API (/zones/.../workers/routes) failed` で終了コード1を返す
- 原因: Account API Token が `All Zones` 権限を持たないため、ゾーン単位のルート登録APIにアクセスできない。**P1-2（NS切替）が未完了**のためこのルート自体は本来まだ機能しないので実害はない
- `workers_dev: true`（一時設定）により `https://animalekarte-stg-api.baritech-soga.workers.dev` が疎通確認用の代替公開経路として機能

**P4-4 — PASS（`/health` 疎通確認）**:
```bash
curl -sf -w '\nHTTP_CODE=%{http_code}\nTIME_TOTAL=%{time_total}\n' \
  'https://animalekarte-stg-api.baritech-soga.workers.dev/health'
# {"status":"ok"}
# HTTP_CODE=200
```

**AC-6 — GORM×PlanetScale 直接接続確認**: `cmd/api/main.go` は `repository.NewDB(cfg)`（GORM Open + Ping相当）が成功するまで HTTP サーバーを起動しない実装（起動順序が直列）。`/health` が複数回 `200` を返した時点で GORM×PlanetScale の直接接続（Hyperdrive非経由・`sslmode=require`）が実際に確立されていることの間接証跡として十分と判断。Container 内部の Go 側 slog 出力（起動ログ）は `wrangler tail` では表示されず（Worker側のHTTPリクエストログのみ捕捉。Container の stdout は Cloudflare Dashboard の Observability/Containers セクション、または Workers Observability API 経由が必要で、後者は現行トークンに対応する権限が無く403）、ログ本文での二重検証は本試行では未実施（残課題）。

**AC-2 — `TRUSTED_PROXY_CIDR` 実測確定 + XFF転送バグの発見・修正**:
1. `/health` に一時的な診断フィールド（`diag_remote_addr`/`diag_client_ip`/`diag_xff`。`c.Request.RemoteAddr`/`c.ClientIP()`/`X-Forwarded-For`ヘッダをそのまま返すだけ）を追加し、デプロイ→観測→**すぐに元へ revert**（`git diff` で復元を確認済み。恒久変更ではない）
2. 観測結果: `diag_remote_addr` は複数リクエスト・複数リージョン（`bom09`→`hkg13`と変わってもポート以外は不変)で `10.1.0.0` に安定。RFC1918 プライベートアドレスであり、Container の `:8080` は Worker の `containerFetch` 経由以外に到達路が無いため、これを信頼しても外部からの偽装によるバイパス経路は生まれない → `TRUSTED_PROXY_CIDR=10.1.0.0/32` に確定（`wrangler.jsonc` の `vars`）
3. **追加で発見した問題**: `@cloudflare/containers` の `container.fetch(request)` は既定では実際のクライアントIP(`CF-Connecting-IP`)を `X-Forwarded-For` として Container へ転送しない。転送しないまま `TRUSTED_PROXY_CIDR` を設定すると、Gin の `c.ClientIP()` は常に内部プロキシIP(`10.1.0.0`)を返すため、レート制限が「全ユーザー共有の単一バケット」になってしまう機能バグがあった(セキュリティ上のバイパスではないが、レート制限の実効性が失われる)
4. **修正**: `backend/worker/index.ts` の `fetch()` で `CF-Connecting-IP` ヘッダを `X-Forwarded-For` に明示的にコピーしてから `container.fetch()` へ渡すよう変更(Cloudflare公式サンプルと同一パターン)
5. 修正後の実測: `diag_xff`/`diag_client_ip` が実際の接続元パブリックIPと一致することを確認 → 修正確認後、診断フィールドは削除(手順1と同様に revert)

**AC-5 — コールドスタート計測**:
| 条件 | `TIME_TOTAL` |
|---|---|
| Cold（`sleepAfter=10m` 経過後の初回リクエスト、Mumbai/`bom09`） | 3.17秒 |
| Cold（設定変更後の再デプロイ直後、Hong Kong/`hkg13`） | 2.26秒 |
| Warm（直後の2回目リクエスト） | 0.19〜0.56秒 |

コールドスタート増分は約 +2〜3秒で、事前の許容目安（+2〜5秒）の範囲内。**リスク登録簿の「Containers のコールドスタートが遅い」は実測により許容範囲内と判定**（min instances は現時点で不要）。

**スコープ拡張（当初計画より前倒しで対応）— R2/S3ストレージ疎通**:
- 試行8で `infra/cloudflare/.env.staging` に R2 S3互換クレデンシャル（`R2_ACCESS_KEY_ID`/`R2_SECRET_ACCESS_KEY`/`S3_ENDPOINT`）が既に発行済みだったことを確認。当初「P2-3未発行のため見送り」としていた前提が古かったため方針転換
- `aws-sdk-go-v2` の既定クレデンシャルチェーン(`config.LoadDefaultConfig`)は `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` という標準名を読む（R2固有の名前ではない）ため、`wrangler secret put AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` で R2 の値をその名前で投入
- `wrangler.jsonc` の `vars` に `STORAGE_TYPE=s3`・`S3_BUCKET`/`S3_SHARED_BUCKET=animalekarte-stg-images`・`S3_REGION`/`S3_SHARED_REGION=auto`・`S3_ENDPOINT` を追加し `worker/index.ts` の `envVars` 経由で Container に注入
- Go側コード変更なし（既存の `NewS3FileStorage`/`NewS3Uploader` がそのまま利用可能）。これにより Container のローカルファイルシステム依存（再起動で揮発する既知の課題）が解消

**`workers_dev` を `true` のまま維持する判断**:
- P1-2（NS切替）が未完了のため、`workers_dev: false` にすると Custom Domain ルートも `*.workers.dev` も両方機能せず、デプロイした Worker/Container への到達路が完全に失われる
- 認証・レート制限・CORS は ECS/Vercel 版と同一の Go バイナリで担保されており、`*.workers.dev` 経路が追加されても保護水準は変わらない（現行 STG も既に `api.stg.noah-karte.com` で公開稼働中のため、新規リスクカテゴリではない）
- Phase 4 継続検証（P4-5等）のため到達可能な状態を維持する必要があり、**NS切替完了後、Custom Domain ルートが機能する時点で `false` に戻す**方針とする（TODOとして残す）

**独立レビュー（code-reviewer + security-reviewer）で判明した修正**:
- code-reviewer から MEDIUM 2件の指摘 → 即時対応済み:
  1. `container.fetch()` にエラーハンドリング欠如 → `try/catch` を追加し、Container起動失敗時に `503 {"error":"service_unavailable"}` を返すよう修正（`worker/index.ts`）
  2. `workers_dev: true` に追跡可能な参照が無い → `// TODO(P1-2 NS切替完了後にfalseへ戻す)` を明示
  - LOW 2件（`S3_ENDPOINT` にアカウントIDが平文・`worker-configuration.d.ts` の型定義が旧値）は許容範囲と判断（後者は `wrangler types` 再実行で解消済み）
  - 修正後、`wrangler deploy` で再デプロイし `/health` 200 を再確認済み
- security-reviewer から CRITICAL 0件・HIGH 1件・MEDIUM 4件・LOW 3件の指摘。「Container:8080はcontainerFetch経由以外に到達路なし」の前提は現行モデルで概ね成立と確認。対応は以下の通り:
  - **即時対応済み**: L-2（CF-Connecting-IP不在時にクライアント制御のXFFをそのまま転送してしまう防御的コーディング不備）→ `worker/index.ts` に `else { headers.delete("X-Forwarded-For") }` を追加。M-4（`worker-configuration.d.ts` の `TRUSTED_PROXY_CIDR` 型定義が旧値 `127.0.0.1/32` のまま）→ `wrangler types` 再実行で同期済み（既にcode-reviewer指摘分で対応済みだったものを再確認）
  - **既知・受容済みのリスクとして継続**（新規対応なし。理由を記録):
    - **H-1** `workers_dev: true` によるSTG API公開 — 本ドキュメントの「`workers_dev`を`true`のまま維持する判断」セクションで既に理由付けした受容リスク。NS切替完了後に`false`へ戻すTODOはコード上に明示済み(残課題として追跡)
    - **M-2** `DB_SSL_MODE=require`（証明書検証なし） — 試行7で確定済みの既存トレードオフ（PlanetScaleのTLS要件との折り合い）。新規の劣化ではない
    - **M-3** R2 Account IDが`wrangler.jsonc`に平文（vars） — 本プロジェクトの運用原則「コード/CLIファースト」（wrangler.jsoncをGit管理する方針）と、Account ID単体では認証情報なしにR2へアクセス不可という実害の低さから、secretsへの追加移動は見送り
  - **次スプリントへの持ち越し**（本試行のスコープ外・設計変更を伴う）:
    - **M-1** in-memory `RateLimitStore`（ログインbrute-force対策）が `sleepAfter=10m` の scale-to-zero でリセットされる — ECS常時稼働では発生しなかった、Containers特有の新規リスク。Cloudflare Workers Rate Limiting API バインディングまたはDB永続化への移行を次フェーズで検討（要設計）
    - **L-1** `TRUSTED_PROXY_CIDR=10.1.0.0/32` の多リージョン対応 — 既にコード内コメントで対応方針を明示済み
    - **L-3** 未使用`HYPERDRIVE`バインディングの削除 — Phase 4完了確定後に削除判断

**残課題（次段以降）**:
- Container 内部（Go slog）の起動ログをダッシュボード外から確認する手段（Workers Observability API 権限不足で403）— 次回トークン権限追加時に `Workers Observability Read` も検討
- `workers_dev: true` → NS切替後に `false` へ戻す
- Account API Token に `Containers Edit`/`Workers Observability` 等、今回判明した不足権限をまとめて棚卸しし、STG用トークンの権限一覧を本ドキュメントに明文化する

### 2026-07-05 試行8（P2-3 R2 S3互換SDK実疎通 — CLI発行）

**前提**: `STG_Terraform&デプロイ` Account API Token に **Account API Tokens Write** を追加。`.env.staging` に `CLOUDFLARE_API_TOKEN` 設定済み。

**P2-3 — PASS**:
- `POST /accounts/776ddc3e975e8fe5773d5300522e2404/tokens` で R2 バケット `animalekarte-stg-images` 限定の子トークンを発行（`Workers R2 Storage Bucket Item Write`）
- Access Key ID = 子トークン `id`、Secret Access Key = 子トークン `value` の SHA-256（[R2 Authentication](https://developers.cloudflare.com/r2/api/tokens/) 準拠）
- `.env.staging` に `R2_ACCESS_KEY_ID` / `R2_SECRET_ACCESS_KEY` / `S3_ENDPOINT` / `R2_BUCKET_NAME` / `S3_SHARED_BUCKET` を追記（gitignore 済み、値はコミット・ログ出力なし）
- `backend/internal/infra/s3_r2_live_test.go` の `TestR2S3Live`（Upload → presigned GET 200 → Delete → HeadObject エラー）を Docker 経由で PASS:
  ```bash
  set -a && source infra/cloudflare/.env.staging && set +a
  docker compose exec -e R2_LIVE_TEST=1 -e R2_ACCESS_KEY_ID -e R2_SECRET_ACCESS_KEY \
    -e S3_ENDPOINT -e S3_SHARED_BUCKET backend \
    go test ./internal/infra/... -run TestR2S3Live -count=1
  # ok github.com/animal-ekarte/backend/internal/infra
  ```
- STG health: `200`（作業前後）

**残課題**: Phase 4（Workers/Containers deploy）、P1-2 NS切替、P2-4/P3-6 データ移行は従来どおり未着手

### 2026-07-05 試行4（.env.staging 投入後・部分 apply）

**認証**: `infra/cloudflare/.env.staging` を `source` してゲート PASS。`wrangler whoami` 成功。

**成功**:
- `terraform import cloudflare_zone.noah_karte d0eec286da621a49fa677dce8fa02c73`
- `cloudflare_zone.noah_karte` 更新（in-place）
- `cloudflare_r2_bucket.stg_images`（`animalekarte-stg-images`, apac）作成
- R2 疎通: `wrangler r2 object put/get/delete --remote` 成功
- STG health: apply 前後とも `200`（NS 未切替のため Vercel DNS 継続）

**失敗（403）**:
- `cloudflare_dns_record.*` 9 件の create
- `cloudflare_hyperdrive_config.stg_planetscale` の create

**state 現状**: `cloudflare_zone.noah_karte`, `cloudflare_r2_bucket.stg_images` のみ

**次アクション（人間タスク）**:
- Cloudflare ダッシュボードで Account API Token に **DNS Edit（zone: noah-karte.com）** と **Hyperdrive Edit** を追加し、`infra/cloudflare/.env.staging` の `CLOUDFLARE_API_TOKEN` を更新
- 更新後: `source infra/cloudflare/.env.staging` → `cd infra/cloudflare && terraform plan -out=tfplan && terraform apply tfplan`（DNS + Hyperdrive の残り。Hyperdrive 分は `pscale role reset-default` → `TF_VAR_pscale_stg_db_*`）
- R2 S3 互換トークン発行 → `S3_ENDPOINT` 実疎通
- Hyperdrive 作成後 GORM 実接続で P3-5 確定
- NS 切替（P1-2）は別途 Vercel ドメイン管理で判断

### 2026-07-05 試行6（DNS Write 追加後・apply 成功）

**認証**: Account API Token に **Zone → DNS Write**（`noah-karte.com`）+ **Hyperdrive Write** を追加。DNS/Hyperdrive API プローブ成功。

**成功**:
- `cloudflare_dns_record.*` 9 件（うち CAA 4 件は Add a Site 既存レコードを `terraform import`、apex/wildcard は Cloudflare 自動スキャンの競合 A レコード 4 件を削除後 CNAME 作成）
- `cloudflare_hyperdrive_config.stg_planetscale` 作成（ID: `45ae9b2a018a4c0fa84c1744c0f12efa`）
- STG health: `200`（NS 未切替・Vercel DNS 継続）

**既知の残課題（すべて試行7で対応・確定済み — 詳細は試行7セクション参照）**:
- ~~Hyperdrive apply 時に provider の `modified_on` タイムスタンプ差分で `Provider produced inconsistent result` が出ることがある~~ → 試行7で回避手順（`-refresh-only`後に`plan`で収束確認）を確立
- ~~Cloudflare ゾーン内に Terraform 未管理のレコード（`www` A×2、`_domainconnect` CNAME）が残存~~ → 試行7で方針決定・実施済み（www削除、domainconnect import）
- ~~R2 S3 互換 API 実疎通、Hyperdrive 実接続 CRUD（P3-5 確定）は未実施~~ → 試行7でP3-5 PASS、試行8でP2-3 PASS（詳細は各試行セクション参照）

**state 現状**: zone + R2 + Hyperdrive + DNS 9 件（計 12 リソース）

### 2026-07-05 試行7（P2-3/P3-5 実接続検証・DNSドリフト修正・wrangler.jsonc確定）

**前提ゲート**: `.env.staging` source 後、`wrangler whoami` 成功（Account: `776ddc3e975e8fe5773d5300522e2404`）。DNS POST probe `success:true`、Hyperdrive list probe `success:true`、R2 buckets probe `http_status:200` — 全PASS。

**Terraform ドリフト調査・修正**:
1. **DNS content 正規化ドリフト**: `zone.tf` の CNAME `content` に末尾ドット(FQDN表記)を付けていたが、Cloudflare API は保存時に末尾ドットを除去して返す（`GET /zones/{id}/dns_records/{id}` で実測確認）。これにより `terraform plan` が毎回 5 レコードの偽陽性 in-place update を報告していた。`apex_flatten`/`wildcard`/`stg_frontend`/`api_stg_backend`/`acm_validation_stg` の `content` から末尾ドットを削除 → `terraform plan` が「No changes」に収束することを確認（最小修正で解消。意図的ドリフトとして残す判断は不要だった）。
2. **Hyperdrive `modified_on` provider既知バグ**: `pscale role reset-default` で発行した新クレデンシャルを適用する `terraform apply` は毎回 `Provider produced inconsistent result after apply`（`modified_on` の計算値差分）で異常終了するが、API側の更新自体は成功している。`terraform apply -refresh-only -auto-approve` → `terraform plan` で「No changes」への収束を確認する回避手順を確立（`hyperdrive.tf` にコメント追記）。

**P2-3（R2 S3互換SDK実疎通）— BLOCKED（再確認）**:
- 現行の統合 Account API Token に `Account: API Tokens: Edit` が無いため、`POST /accounts/776ddc.../tokens` で R2スコープの子トークンを発行しようとしたところ 403 `Unauthorized to access requested resource`。R2 の S3互換 Access Key ID/Secret Access Key は「発行した Cloudflare API Token の id / token値のSHA-256」という導出規則（`developers.cloudflare.com/r2/api/tokens/`）であり、ダッシュボードでの新規発行（例外操作）または既存トークンへの権限追加が必要。
- 代替として wrangler CLI 経由（S3互換ではなく Cloudflare 独自認証）の R2 put/get/delete を再実行し PASS（`_smoke/wrangler-check.txt` で確認、テスト後削除済み）。これは P2-1 の再確認であり、P2-3 が要求する「S3互換SDK経路」の検証にはならない。
- **人間タスク**: Cloudflare ダッシュボード → R2 → Manage R2 API Tokens → Permissions: Object Read & Write → bucket を `animalekarte-stg-images` に限定 → 発行。得られた Access Key ID / Secret Access Key を `.env.staging` に `R2_ACCESS_KEY_ID` / `R2_SECRET_ACCESS_KEY` として追記（値はコミットしない）。`S3_ENDPOINT=https://776ddc3e975e8fe5773d5300522e2404.r2.cloudflarestorage.com` と合わせて `docker compose exec backend go test ./internal/infra/... -run R2` 相当で再検証。

**P3-5（Hyperdrive実接続 CRUD/トランザクション）— PASS（確定）**:
- Cloudflare Hyperdrive の `connectionString` は Worker 実行コンテキスト内でのみ取得可能という仕様上の制約があり（`GET /accounts/{id}/hyperdrive/configs/{id}` は origin host/port/user のみを返し、パスワードや疎通可能な直結エンドポイントは返さない。Cloudflare公式ドキュメント "Getting started"/"Connect to PostgreSQL" 準拠）、ローカル Go バイナリ/psql からの直接検証は不可能と判明。
- `/tmp`（リポジトリ外）に一時 Worker プロジェクトを作成し、`wrangler dev --remote`（エフェメラルな edge プレビューセッション。**永続 deploy ではない** — セッション終了で自動消滅し、公開ルートは作成されない）で実 Hyperdrive バインディング（`45ae9b2a018a4c0fa84c1744c0f12efa`）を使用。`postgres.js` から以下を全実行し成功:
  - `select 1` による基本疎通
  - `inet_server_addr()` がプール内部アドレス(`10.198.159.119/32`)であることを確認 → PlanetScaleへの直結ではなくHyperdriveプロキシ経由であることの証拠
  - スクラッチテーブル `hyperdrive_smoke_test` への INSERT → UPDATE → SELECT → DELETE（テーブルは検証後 `DROP TABLE` で削除済み。臨床/マスタテーブルには触れていない）
  - `sql.begin()` での BEGIN→COMMIT（成功時にコミットされたレコードのID確認）
  - `sql.begin()` 内で例外を発生させた BEGIN→ROLLBACK（ロールバック後に対象行が0件であることまで確認）
- 検証後、一時 Worker プロジェクト（`/tmp` 配下）・プロセスを終了・削除。リポジトリへの変更なし（`backend/wrangler.jsonc` の hyperdrive ID 更新のみが本体への変更）。
- **GORM(Go)自体の実接続確認は範囲外**（上記の制約により、GORMを使うにも実際のWorker/Containerの実行コンテキストが必要。Phase 4でバックエンドがWorkers/Containers上で実際に動く段になったら、GORM経由でも同様の確認を行うこと）。

**未管理DNSの棚卸しと方針決定**:
- `www.noah-karte.com` A×2（`216.150.1.129`/`216.150.1.193`, proxied=true）→ **削除**。`vercel dns ls noah-karte.com` で Vercel 側に www 専用レコードが存在しないことを確認（apex/wildcardの汎用ALIAS配下）。これらは Cloudflare の Add a Site 時自動スキャンによる、当時のAnycast IPの静止スナップショットであり、apex/wildcardで発覚した「IPローテーションで陳腐化する」問題と同種のリスクを抱えていた（かつ proxied=true はDNS onlyポリシーに反していた）。既存の `cloudflare_dns_record.wildcard`（CNAME flatten）で正しくカバーされるため削除が適切と判断し、Cloudflare API で削除実施。
- `_domainconnect.noah-karte.com` CNAME（→ `_domainconnect.vercel-dns.com`）→ **import**。Vercel registrar が自動提供する Domain Connect discovery レコード。用途未確定だが削除の影響が確認できないため、ACM検証レコードと同様の慎重方針（維持）を採用。`terraform import` でstate取り込み後、`proxied: true→false`・`ttl: 1(auto)→300` をDNS onlyポリシーに合わせて正規化。`zone.tf` に `cloudflare_dns_record.domainconnect` として追加。
- 上記の結果、`terraform state list` は 13 リソース（試行6の12 + `domainconnect`）。

**`backend/wrangler.jsonc` 確定**: hyperdrive `id` の placeholder を `45ae9b2a018a4c0fa84c1744c0f12efa` に置換。`grep PLACEHOLDER backend/wrangler.jsonc` は0件。`wrangler deploy --dry-run` はJSONC構文エラーなし（entry point未実装エラーのみ、想定通り）。

**PlanetScale クレデンシャル管理**: 本セッションで発行した資格情報は Hyperdrive apply（試行6の初期apply分＋本セッションのpassword rotation分）にのみ使用し、検証完了後に都度 `pscale role reset-default --force` で失効・再発行した。値をログ・ドキュメント・チャットに残していない。

**STG health**: 作業前 `200`、DNS削除/import・Hyperdrive更新・wrangler.jsonc更新の全作業後も `200`（NS未切替のためVercel DNS経路継続、想定通り無変化）。

**最終 `terraform plan`**: 全操作後に「No changes. Your infrastructure matches the configuration.」に収束（PlanetScale資格情報を都度失効させる運用のため、次回Hyperdriveをtouchする際は再度`TF_VAR_pscale_stg_db_*`供給が必要になる点は既知の意図的ガード）。

**残課題（次回セッションへの引き継ぎ）**:
- ~~P2-3（R2 S3互換SDK実疎通）~~ → 試行8で PASS
- GORM(Go)経由でのHyperdrive実接続確認はPhase 4着手後（Worker/Containers実装時）に実施
- Phase 4着手そのものをHyperdrive必須とするかは、上記のとおりHyperdriveプロキシ自体の実接続はpostgres.js経由で確定PASSしているため、着手のブロッカーにはならないと判断

**生ログ証跡（2026-07-05 18:10 JST 時点で再取得。監査用）**:
```
$ curl -s -o /dev/null -w '%{http_code}\n' https://api.stg.noah-karte.com/health
200

$ terraform plan   # infra/cloudflare, .env.staging source済み, TF_VAR_account_id供給済み
No changes. Your infrastructure matches the configuration.

$ terraform state list | wc -l
13
```

**独立レビュー（`code-reviewer`）**: Verdict **APPROVE WITH COMMENTS**（CRITICAL/HIGH 0件）。MEDIUM1件（`backend/wrangler.jsonc` ヘッダーコメントの陳腐化した「未apply/placeholder」表記）は提案差分どおり修正済み。LOW3件（試行節の挿入順序・Account/Zone IDのリテラル記載・imageフィールド警告強度）は運用判断事項として対応見送り。

### 2026-07-05 試行5（トークン更新後・再 apply）

**認証**: `.env.staging` を `source` してゲート PASS。`wrangler whoami` 成功（Account API Token、`776ddc3e…`）。

**権限診断（curl）**:

| API | 結果 |
|---|---|
| `GET /zones/{zone_id}` | ✅ success |
| `GET/POST /zones/{zone_id}/dns_records` | ❌ 403 Authentication error |
| `GET/POST /accounts/{account_id}/hyperdrive/configs` | ❌ 403 Authentication error |
| `GET /accounts/{account_id}/r2/buckets` | ✅ success（バケット 1 件） |
| `POST /user/tokens/verify` | ❌ Invalid API Token（Account API Token では verify が失敗することがある。DNS POST で判定すること） |

**結論**: トークン更新後も **Zone → DNS → Edit** と **Account → Hyperdrive → Edit** が未付与、または **Zone Resources に `noah-karte.com` が含まれていない**。R2 権限のみ有効な状態。

**apply 結果**: 試行4と同様、DNS 9 件 + Hyperdrive 1 件が 403 で失敗。state は変化なし。

**人間タスク（再確認チェックリスト）**:

1. [Cloudflare API Tokens](https://dash.cloudflare.com/776ddc3e975e8fe5773d5300522e2404/api-tokens) で **`.env.staging` と同じトークン**を編集（別トークンを更新していないか確認）
2. **Permissions** に以下を **Edit** で追加:
   - `Zone` → `DNS` → **Edit**
   - `Account` → `Hyperdrive` → **Edit**（UI 上 `Workers Hyperdrive` 等の名称の場合あり）
3. **Zone Resources**: `Include` → `Specific zone` → **`noah-karte.com`**
4. **Account Resources**: 対象アカウント `776ddc3e975e8fe5773d5300522e2404` を Include
5. 保存時に **Roll token** した場合は、新しいトークン文字列を `.env.staging` の `CLOUDFLARE_API_TOKEN` に貼り替え（権限だけ変えた場合は文字列は同じ）
6. `.env.staging` に `TF_VAR_account_id=776ddc3e975e8fe5773d5300522e2404` も記載推奨（現状 `CLOUDFLARE_API_TOKEN` と `ZONE_ID` のみ）
7. 更新後、次で **true** になることを確認してから apply:
   ```bash
   set -a && source infra/cloudflare/.env.staging && set +a
   curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records" \
     -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" -H "Content-Type: application/json" \
     -d '{"type":"TXT","name":"_perm-test.noah-karte.com","content":"ok","ttl":120,"proxied":false}' \
     | jq '.success'
   # true → テストレコードをダッシュボードで削除 → terraform apply 再実行
   ```

**実行内容（試行4以前）**:

1. **P0-5 ツールチェーン**: `wrangler` を `package.json` の `devDependency`（`4.107.0`）として追加、`pnpm-lock.yaml` 生成済み。`pscale auth check` は認証済みを確認。`rclone` / `cf-terraforming` は `infra/cloudflare/README.md` に導入手順を記載（本機では未インストール、手順書のみ）。
2. **P1-1 DNS棚卸し**: 初回 `dig` 実測で apex(@) の A レコードが `1.1.1.1`/`8.8.8.8`/権威DNS間で値が変動することを発見 → Vercel 側 Anycast/ALIAS の動的解決と判明。`vercel dns ls` / Vercel API（`GET /v4/domains/noah-karte.com/records`）で一次情報を取得し、`infra/cloudflare/zone.tf` を ALIAS→CNAME flattening 方式で書き直し。棚卸し結果は `zone.tf` 冒頭コメントに全件記載（stg CNAME・api.stg CNAME・ACM検証CNAME・apex/wildcard CNAME flatten・CAA 4件）。**要人間確認**: `_a1ecec492bd7059488c176bca7348f1a.stg` ACM検証レコードの対象証明書が未確定（削除禁止として保持のみ）。
3. **P1-2 前半事実確認**: `noah-karte.com` の registrar は Vercel（Vercel Domains、2026-03-21 取得）。NS切替は一般的なレジストラ操作ではなく Vercel ドメイン管理画面/APIでのカスタムネームサーバー設定になる。実施は人間タスク（本タスクでは未実施）。
4. **P2-1 R2バケット**: `infra/cloudflare/r2.tf` に定義。試行4で **apply 成功**（`animalekarte-stg-images`）。
5. **P2-2 S3_ENDPOINT対応**: TDD（Red→Green）で実装。`backend/internal/config/config.go` に `S3Endpoint` 追加、`backend/internal/infra/s3_endpoint.go` に `applyS3EndpointOverride`/`buildS3ObjectURL` を新設、`s3_file_storage.go`/`s3_uploader.go`/`cmd/api/main.go` を更新。`go test ./internal/config/... ./internal/infra/...` 全PASS、`go vet` クリーン、`gofmt` クリーン（Docker経由で実行）。
6. **P2-4 データ移行スクリプト雛形**: `infra/scripts/migrate-images-r2.sh`（rclone sync/check、dry-run確認プロンプト付き）を作成。実行はAWS認証情報+R2認証情報の両方が必要なためスコープ外・未実行。
7. **P3-1 PlanetScale DB作成**: `pscale database create animalekarte-stg --org noah-animalekarte --region ap-northeast --cluster-size PS-10` を実行。`pscale branch list` で `main` ブランチ region=`ap-northeast`・READY=Yes を確認。手順を `infra/scripts/pscale-create-stg.sh` にスクリプト化。
8. **P3-2/P3-3 スキーマ/拡張検証**: `pscale role reset-default` で都度クレデンシャルを発行し、`backend/cmd/migrate` 相当の手順で `backend/migrations/001〜005` の全5件を空DBへ適用（`schema_migrations` テーブルで5/5件確認、ファイル名一致）。`infra/scripts/validate-schema.sql` で以下を確認済み: 拡張機能 `pg_trgm`（+ PlanetScale標準の `hypopg`/`plpgsql`）、ENUM型多数、`text[]`等の配列カラム、`jsonb` カラム、トリガ2件（`trg_create_default_payment_methods` 等）。public スキーマ テーブル数 109。検証に使ったクレデンシャルは都度 `pscale role reset-default --force` で失効・再発行し、値を本ドキュメント・ログに残していない。
9. **P3-4 Hyperdrive設定**: `infra/cloudflare/hyperdrive.tf` に `cloudflare_hyperdrive_config`（`caching.disabled = true`）を定義。`pscale_stg_db_*` 変数（`variables.tf`）はデフォルト空文字でガード。apply は BLOCKED。
10. **P3-5 GORM×Hyperdrive互換判定**: 静的解析による暫定判断（実接続検証はBLOCKEDのため未実施）。`backend/internal/repository/db.go` の GORM初期化は `PrepareStmt` 未指定（既定 false）— SQLレベル `PREPARE`/`DEALLOCATE` を発行しないため、通常のAPIトラフィック（Handler→Service→Repository経路）は Hyperdrive の「prepared statement非対応」制約に抵触しないと判断。一方 `backend/cmd/migrate/main.go` は `pg_advisory_lock` を使用し Hyperdrive非対応のため、**migrate/一shotタスクは PlanetScaleへ直結（Hyperdriveを経由しない）**方針とする。`LISTEN/NOTIFY` は使用なし（grep確認済み）。結論は暫定であり、Phase 4着手前に人間が `CLOUDFLARE_API_TOKEN` を用意した上で実Hyperdrive接続によるCRUD/トランザクション検証を行うこと。
11. **P4-1 wrangler.jsonc雛形**: `backend/wrangler.jsonc` を新規作成（Worker routes・Containers・Durable Objects・Hyperdrive binding・R2 binding・secrets.requiredの一覧）。`wrangler deploy --dry-run --config=backend/wrangler.jsonc` でJSONC構文の妥当性を確認（entry point未実装によるエラーは想定通り。ネットワーク呼び出し・Cloudflare認証は発生せず）。
12. **独立レビュー**: `go-reviewer` サブエージェントがS3_ENDPOINT関連コード（CRITICAL 0 / HIGH 1 / MEDIUM 3）、`code-reviewer` サブエージェントがTerraform+wrangler.jsonc（CRITICAL 0 / HIGH 0 / MEDIUM 3 / LOW 2、Verdict: APPROVE）をレビュー。指摘のうちHIGH（テスト隔離漏れ）とMEDIUM（table-driven化、wrangler.jsonc `image`フィールドのTODO補足、`terraform.tfvars.example`追加）は本セッション内で対応済み。残りMEDIUM（tfstate平文保存対策・変数validation追加）はコメントで既知のリスクとして文書化済み、対応要否はチーム判断に委ねる。
13. **セキュリティ配慮**: PlanetScale検証用に発行したクレデンシャルは検証終了後に都度 `pscale role reset-default --force` で失効・再発行し、有効なクレデンシャルを放置していない。Cloudflare API Token/Account IDは全工程で環境変数からのみ参照し、値の出力・コミット・ログ保存は行っていない。

---

## 2. フェーズ概要と工数

| Phase | 内容 | 概算工数 | 依存 |
|---|---|---|---|
| 0 | 意思決定・アカウント準備 | 1〜2 人日 | — |
| 1 | エッジ層（DNS/CDN/SSL/WAF） | 1〜2 人日 | 0 |
| 2 | 画像ストレージ S3 → R2 | 2〜3 人日 | 0 |
| 3 | DB RDS → PlanetScale + Hyperdrive | 3〜5 人日 | 0 |
| 4 | コンピュート ECS → Workers + Containers | 5〜10 人日 | 2, 3 |
| 5 | CI/CD 置換（GitHub Actions → wrangler） | 2〜3 人日 | 4 |
| 6 | 監視・ログ・通知 | 1〜2 人日 | 4 |
| 7 | 切替・並行稼働・検証 | 2〜3 人日 | 1〜6 |
| 8 | AWS リソース廃止 | 1〜2 人日 | 7 |
| | **合計** | **18〜32 人日** | |

Phase 1〜3 は互いに独立しており並行着手可能。Phase 4 が最大リスク・最大工数。

---

## Phase 0: 意思決定・アカウント準備（1〜2 人日）

- [x] **P0-1** Cloudflare アカウント作成 / Workers Paid プラン契約（**例外的ダッシュボード操作 #1**）— 人間側で事前完了済み（前提条件として本タスク開始時に確認）
- [x] **P0-2** STG ドメインのゾーン追加方針決定 — ゾーン全体（`noah-karte.com`）を Cloudflare へ移管する方針（サブドメイン委任はCloudflare Enterprise限定のため不採用。`zone.tf` 冒頭コメント参照）
- [x] **P0-3** PlanetScale の作成経路とリージョン決定 — **(a) pscale CLI 直接作成**を採用、リージョンは `ap-northeast`（東京相当）で決定・作成済み
- [x] **P0-4** チーム内合意: PlanetScale は PITR なし（12h 毎バックアップ）を STG として受容 — 本ドキュメントに記録済み（§8 DB選定の経緯）
- [x] **P0-5** ツールチェーン整備:
  - `wrangler`（`package.json` devDependency、`4.107.0`）導入済み。`pscale`（認証済み）/ `rclone` / `cf-terraforming` は `infra/cloudflare/README.md` に導入手順を記載
  - 初回 API Token 発行（**例外 #2**）は人間側で発行済み（統合トークン1本方針、上表参照）。本タスク実行時点では環境変数 `CLOUDFLARE_API_TOKEN`/`CLOUDFLARE_ACCOUNT_ID` は未エクスポート（**BLOCKED** — plan/apply不可の直接要因）
  - `infra/cloudflare/` 新設済み（`providers.tf`/`variables.tf`/`backend.tf`/`zone.tf`/`r2.tf`/`hyperdrive.tf`/`README.md`/`terraform.tfvars.example`）。tfstateは当面local backend、R2 backendへの切替はTODOコメントで残置（`backend.tf`）
- [ ] **P0-6** 移行凍結ウィンドウの調整（Phase 3 の DB 切替時に STG への書き込みを止める時間帯）— Phase 3後半（本切替）着手時に人間が調整する事項のため未実施

**Done 基準**: `infra/cloudflare/` で `terraform plan` が通り（`terraform validate` は成功済み。`plan`/`apply`実行はCLOUDFLARE_API_TOKEN/ACCOUNT_ID未設定のためBLOCKED）、`wrangler whoami` / `pscale auth check` が成功する状態（`pscale auth check` は認証済みを確認。`wrangler whoami` はCloudflare認証情報が無いため未実行）。

---

## Phase 1: エッジ層 — DNS / CDN / SSL / WAF（1〜2 人日）

最小リスクの第一手。**既存 AWS スタックを温存したまま**前段に Cloudflare を被せる。**すべて `infra/cloudflare/` の Terraform で定義**する。

- [x] **P1-1** ゾーン作成（`cloudflare_zone`）+ 既存 DNS レコードの Terraform 化 — `infra/cloudflare/zone.tf` に定義済み（`terraform validate` 成功。`apply` はBLOCKED）。棚卸しは `dig`（動的Anycast発覚）→ `vercel dns ls`/Vercel API一次情報で完了。詳細は上記実施記録 P1-1 参照
- [ ] **P1-2** ネームサーバ移管（or サブドメイン NS 委任）。TTL を事前に短縮しておく — **スコープ外（人間タスク）**。事実確認済み: registrarはVercel、切替はVercelドメイン管理画面/APIでのカスタムNS設定（実施記録 P1-2 参照）
- [ ] **P1-3** SSL モード **Full (strict)** を `cloudflare_zone_setting` で設定 — **未着手**（本タスクのスコープ外。P1-2完了後に着手）
- [ ] **P1-4** 無料枠 WAF マネージドルール・Bot Fight Mode を `cloudflare_ruleset` / `cloudflare_bot_management` で定義 — **未着手**（スコープ外）
- [ ] **P1-5** キャッシュルールを `cloudflare_ruleset` で定義 — **未着手**（スコープ外）
- [ ] **P1-6** Cookie 認証がプロキシ経由でも成立することを検証 — **未着手**（NS未切替のため検証不可。Phase 7で実施）
- [ ] **P1-7** LIFF / LINE 予約のコールバック URL がドメイン変更の影響を受けないか確認 — **未着手**（スコープ外）

**ロールバック**: ネームサーバを元に戻すだけ（AWS 側は無変更のため即時復旧可能）。

---

## Phase 2: 画像ストレージ S3 → R2（2〜3 人日）

- [x] **P2-1** R2 バケット作成 — `infra/cloudflare/r2.tf` の `cloudflare_r2_bucket`（`animalekarte-stg-images`, location=apac）は試行4で **apply 済み**。wrangler CLI 経由の put/get/delete は試行4・試行7で再確認済み
- [x] **P2-2** **コード変更**: `S3_ENDPOINT` を `backend/internal/config/config.go` に追加、`backend/internal/infra/s3_endpoint.go`（`applyS3EndpointOverride`/`buildS3ObjectURL`）を新設、`s3_file_storage.go`/`s3_uploader.go`/`cmd/api/main.go` を更新。TDDで実装、`go test`/`go vet`/`gofmt` 全クリーン（実施記録 5. 参照）
- [x] **P2-3** R2 S3互換SDK経路での Upload / GetSignedURL / Delete 実疎通 — **2026-07-05 試行8で PASS**。Account API Token に `Account API Tokens Write` 追加後、`POST /accounts/{id}/tokens` で R2 スコープ子トークンを CLI 発行。`backend/internal/infra/s3_r2_live_test.go` の `TestR2S3Live`（`R2_LIVE_TEST=1`）で Docker 経由検証。詳細は試行8記録参照
- [x] **P2-4** 既存オブジェクトの移行スクリプト雛形 — `infra/scripts/migrate-images-r2.sh` を作成（rclone sync/check、dry-run確認プロンプト付き）。**実行はスコープ外**（AWS認証情報 + R2認証情報の両方が必要）
- [ ] **P2-5** 突合（`rclone check`） — **未実行**（P2-4未実行のため）
- [ ] **P2-6** STG 環境変数切替 — **未実施**（Phase 4のデプロイ実行に付随するためスコープ外）
- [ ] **P2-7** clinic_id 隔離の回帰確認 — **未実施**（R2実疎通確認後に実施する事項）

**ロールバック**: 環境変数を S3 に戻す（二重書き込みはしない。切替後に S3 へ書かれないことだけ確認）。

---

## Phase 3: DB — RDS → PlanetScale + Hyperdrive（3〜5 人日）

- [x] **P3-1** PlanetScale Postgres 作成 — `pscale database create animalekarte-stg --org noah-animalekarte --region ap-northeast --cluster-size PS-10` 実行済み。`main`ブランチ region=ap-northeast・READY=Yes を確認。手順は `infra/scripts/pscale-create-stg.sh` にスクリプト化
- [x] **P3-2** スキーマ互換性の事前検証 — `backend/migrations/001〜005` 全5件を空DBへ適用、`schema_migrations`テーブルで5/5件・ファイル名一致を確認。ENUM/`text[]`/jsonb/トリガすべて動作確認済み（`infra/scripts/validate-schema.sql`で再現可能。実施記録 8. 参照）
- [x] **P3-3** 拡張機能の確認 — `\dx`で`pg_trgm`（+ PlanetScale標準の`hypopg`/`plpgsql`）を確認。現行RDSで使用中の拡張と一致
- [x] **P3-4** Hyperdrive 設定作成 — `infra/cloudflare/hyperdrive.tf` の `cloudflare_hyperdrive_config`（`caching.disabled = true`）は試行6で **apply 済み**（ID: `45ae9b2a018a4c0fa84c1744c0f12efa`）。`backend/wrangler.jsonc` の placeholder は試行7でこの ID に置換済み
- [x] **P3-5** Hyperdrive 実接続 CRUD/トランザクション検証 — **2026-07-05 試行7で PASS**。`wrangler dev --remote`（エフェメラルな edge プレビュー、永続 deploy ではない）+ `postgres.js` から `env.HYPERDRIVE.connectionString` 経由で SELECT/INSERT/UPDATE/DELETE + BEGIN→COMMIT/BEGIN→ROLLBACK を全実行し成功。`inet_server_addr()` がプール内部アドレスであることから Hyperdrive 経由であることを確認。詳細は下記「試行7」記録参照。**GORM(Go)自体での実接続確認は未実施** — Cloudflare Hyperdrive の `connectionString` は Worker 実行コンテキスト外(単体Goバイナリ/ローカルpsql)からは取得不能という Cloudflare 側の仕様上の制約のため、GORM経由の確認は Phase 4 で実際の Worker/Container 実装が組まれた時点で行う。GORM互換性についての静的解析結論（`PrepareStmt`既定false・migrate直結方針）は実施記録10.・`hyperdrive.tf`コメント参照で変更なし
- [ ] **P3-6** データ移行リハーサル — **スコープ外**（private RDSへの接続経路が必要なため未実施）
- [ ] **P3-7** 本切替 — **スコープ外**（P3-6未実施のため）
- [ ] **P3-8** RDS は**即削除しない** — 該当なし（Phase 8着手時まで判断保留）

**ロールバック**: 接続先環境変数を RDS に戻す。凍結ウィンドウ内なら データロスなし。

---

## Phase 4: コンピュート — ECS Fargate → Workers + Containers（5〜10 人日）

最大リスク工程。Phase 2・3 完了後に着手する（R2/PlanetScale を向いた状態のイメージで検証するため）。

- [x] **P4-1** wrangler 設定作成 — `backend/wrangler.jsonc` 作成済み（Worker routes・Containers・Durable Objects・Hyperdrive/R2 binding・secrets.required一覧）。`wrangler deploy --dry-run`でJSONC構文の妥当性を確認済み（entry point未実装分のエラーは想定通り）。Hyperdrive ID は試行7で `45ae9b2a018a4c0fa84c1744c0f12efa`（実 apply 済み ID）に置換済み（placeholder は解消。`grep PLACEHOLDER` 0件）。R2 bucket名は当初から確定値（`animalekarte-stg-images`）
- [x] **P4-2** イメージビルド・push — **2026-07-05 試行9で PASS**。`wrangler deploy` が `Dockerfile.production` からビルドし Cloudflare Registry へ push。Container アプリケーション `animalekarte-stg-api-animalekarteapicontainer`（Application ID: `a03e1180-c9fd-41e4-85ef-7773406d2335`）作成済み
- [x] **P4-3** シークレット移設 — **2026-07-05 試行9で PASS**。`wrangler secret put` で DB認証情報（PlanetScale・`pscale role reset-default`で都度再発行）・`JWT_SECRET`/`INTEGRATION_ENCRYPTION_KEY`・R2認証情報（`AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY`名で投入）・SMTP/LINE/LSTEP（未使用分は空文字）を投入済み
- [x] **P4-4** ヘルスチェック — **2026-07-05 試行9で PASS**。`https://animalekarte-stg-api.baritech-soga.workers.dev/health` が `200 {"status":"ok"}`。コールドスタート実測 2.26〜3.17秒・ウォーム 0.19〜0.56秒（許容目安+2〜5秒の範囲内）。詳細は試行9記録参照
- [x] **P4-5** migrate one-shot の置換 — **2026-07-05 試行10で PASS**。`POST /_internal/migrate`(`MIGRATE_RUN_SECRET`認証)→ `Container.exec(["/app/migrate"])` → exit code をJSONで返却。`infra/scripts/cf-run-migrate.sh` + `pnpm cf:migrate` で運用。PlanetScale STGで exit 0 ×2(冪等)・`schema_migrations` 5件整合・`/health`回帰なしを実測確認。詳細は試行10記録参照
- [x] **P4-6** 機能スモーク — **2026-07-05 試行11で PASS**。`infra/scripts/cf-crud-smoke.sh`(`pnpm cf:smoke`)で CRUD(clinics/permission-groups/staffs) + 混在会計(payment_splits) API スモークを実施、全11 AC中10 PASS・1 BLOCKED(AC-11、UI検証はスコープ外)。詳細は試行11記録参照
- [x] **P4-7** 外部連携の検証 — **2026-07-05 試行12で PASS/BLOCKED(genuine)**。`docs/infra/deploy/CLOUDFLARE-EXTERNAL-INTEGRATIONS-AUDIT.md` 新規作成。LINE(既定IP allowlist非依存、doc結論PASS/live送信は誤配信リスク回避でBLOCKED)・Lstep(Write4メソッド`[DISABLED]`継続確認)・SMTP(secret名存在確認済み、値非取得のためBLOCKED)・LIFF(DNS未切替でBLOCKED、P7-3 defer)。詳細は試行12記録参照
- [x] **P4-8** Cookie 認証の再検証 — **2026-07-05 試行12で PASS**。curl で `Set-Cookie` の `HttpOnly`/`Secure`/`SameSite=None` を確認（AC-A）、ローカル frontend(docker compose)から workers.dev への実ブラウザcross-originログイン成功（AC-B、Network 200+以降のAPI呼び出しも200）。検証中に frontend `index.html` の CSP `connect-src` が workers.dev をブロックする新規事象を発見・一時許可して検証、検証後revert・redeploy済み。詳細は試行12記録参照
- [x] **P4-9** 負荷スモーク — **2026-07-05 試行12で PASS**。`load-tests/k6-cf-stg-sustained.js`(`pnpm cf:load-smoke`)で10分負荷実行(Docker経由grafana/k6)、失敗率0.00%・p95=897ms・exit code 0。Cloudflare GraphQL Analytics API(`containersUsageAdaptiveGroups`)でCPU実測、月額試算(~$4)との比較は±50%以内(詳細は試行12記録参照)

**ロールバック**: DNS/Worker ルートを既存 CloudFront/ALB オリジンへ戻す（Phase 7 まで ECS は稼働継続しているため即時可能）。

---

## Phase 5: CI/CD — GitHub Actions 置換（2〜3 人日）

- [ ] **P5-1** `backend-deploy.yml` の置換 — AWS OIDC / ECR push / task-definition render / ECS update / migrate task 実行のジョブ群を `wrangler deploy`（+ P4-5 の migrate 実行）に置換。デプロイ後検証（wait-for-stability 相当 → `/health` ポーリング）を再実装
- [ ] **P5-2** GitHub Secrets に Cloudflare API Token 登録（deploy 権限のみの最小スコープ）、AWS OIDC ロール参照の除去
- [ ] **P5-3** `staging-stop.yml`（夜間停止）を**廃止** — scale-to-zero + PlanetScale 固定額により不要
- [ ] **P5-4** `stg-smoke.yml` / `e2e.yml` / `performance-tests.yml` の対象 URL・前提を新構成に更新
- [ ] **P5-5** デプロイ 2 回連続成功（通常デプロイ + マイグレーション含むデプロイ）を確認

---

## Phase 6: 監視・ログ・通知（1〜2 人日）

- [ ] **P6-1** Workers Logs 有効化 — `wrangler.jsonc` の `observability` 設定で宣言（Paid: 2,000 万件/月込み・保持 7 日）
- [ ] **P6-2** 監査要件の確認 — 7 日保持で STG として十分か判断。不足なら Logpush → R2 を Terraform `cloudflare_logpush_job` で設定
- [ ] **P6-3** 通知設定 — Terraform `cloudflare_notification_policy`（Worker エラー率・Containers 異常）
- [ ] **P6-4** Budget Alert 設定（月 $40 目安 — 想定 $27.5 の 1.5 倍で異常検知）。API 経由で設定できない場合はダッシュボード操作を例外として記録
- [ ] **P6-5** CloudWatch ダッシュボード/アラームで代替不能なものがないか最終確認

---

## Phase 7: 切替・並行稼働・検証（2〜3 人日 + 監視 1〜2 週間）

- [ ] **P7-1** DNS 切替 — Worker ルートを本経路に昇格（旧 CloudFront 経路は温存）
- [ ] **P7-2** 並行稼働期間（1〜2 週間）— 旧 ECS/ALB は起動したまま、トラフィックのみ新経路。日次でエラーログ・課金実績を確認
- [ ] **P7-3** フルスモーク — CRUD / 会計 / 画像アップロード / LINE 連携 / 帳票の全系統
- [ ] **P7-4** Vercel フロントエンドの API 向き先確認（`frontend-deploy.yml` の環境変数に API URL があれば更新）
- [ ] **P7-5** 関係者への切替完了周知、`docs/infra/` 配下の運用ドキュメント更新（`INFRA_ARCHITECTURE.md` / `STG-CONTINUOUS-OPERATIONS.md` / `CI-CD-PIPELINE.md`）

**Go/No-Go 基準**: 並行稼働期間中、新経路でエラー率が旧経路同等以下・スモーク全通過・課金が試算 ±50% 以内。

---

## Phase 8: AWS リソース廃止（1〜2 人日）

**順序が重要**。データを持つものは最後、復旧手段は最後の最後。

- [ ] **P8-1** RDS 最終スナップショット取得 → エクスポート保管（保険。90 日後に削除判断）
- [ ] **P8-2** S3 画像バケットの最終確認（R2 との突合済みであること）→ バケット削除 or Glacier 退避
- [ ] **P8-3** `terraform destroy` — ECS / ALB / EventBridge Scheduler / fck-nat / EIP / VPC の順で削除（tfvars の toggle を使い段階的に）
- [ ] **P8-4** ECR リポジトリ削除、CloudFront ディストリビューション（手動管理）削除、ACM 証明書削除
- [ ] **P8-5** GitHub OIDC Provider / IAM ロール削除（`terraform` ロールは最後 — destroy 自体に必要）
- [ ] **P8-6** tfstate 用 S3 + DynamoDB の処置（他環境と共用なら残置。STG 専用なら最後に削除）
- [ ] **P8-7** 翌月の AWS 請求が $0（または tfstate 分のみ）であることを確認して完了

---

## 8. DB 選定の経緯（記録）

| 案 | 判定 | 理由 |
|---|---|---|
| D1 (SQLite) | 不採用 | 10GB ハード上限に対しデータ 32GB で物理的に不可。方言非互換で 30+ 人日 |
| RDS 継続 + Hyperdrive | 次点 | 最小リスクだが VPC/NAT 維持費が残り「全面移行」にならない。PITR が必要になったら戻る先 |
| Neon | 不採用（STG） | 東京リージョンなし。32GB でストレージ単価（$0.35/GB）が PlanetScale の約 3 倍、無料枠（0.5GB）も対象外 |
| **PlanetScale** | **採用** | 東京あり・PS-10 で月 ~$18.5・`pscale` CLI で全操作可能。CF 課金統合はダッシュボード連携（OAuth）が要るため CLI ファースト原則とトレードオフ（P0-3 で選択）。制約: PITR なし（12h 毎バックアップ）を STG として受容 |

## 9. リスク登録簿

| リスク | 影響 | 緩和策 |
|---|---|---|
| ~~Containers のコールドスタートが遅い~~ | STG 利用者の初回アクセスが数秒待ち | 試行9で実測: 2.26〜3.17秒（許容目安+2〜5秒の範囲内）。**解消済み**。将来的に許容不可と判断されれば min instances 検討（費用増） |
| **[試行9で新規発見]** in-memory RateLimitStore が scale-to-zero でリセットされる | ログインbrute-force対策が10分アイドル毎に無効化（ECSの常時稼働では発生しなかった新規リスク） | security-reviewer指摘(M-1)。Cloudflare Workers Rate Limiting API バインディングまたはDB永続化への移行を次フェーズで検討（要設計。本試行のスコープ外） |
| Hyperdrive と GORM の prepared statement 非互換 | クエリ実行エラー | P3-5 で事前検証。NG なら PlanetScale 直結に切替（PgBouncer 同梱で接続プールは確保可能） |
| ~~IP allowlist 依存の外部連携~~ | LINE/Lstep 連携断 | 試行12で棚卸し完了。LINEは既定でIP allowlist非依存（オプション機能のみ・要クリニック側Console確認）、Lstep Write系は`[DISABLED]`のため現状無影響。SMTP/LIFFはBLOCKED（secret値未確認/DNS未切替）。**解消済み（残作業はP7-3へdefer）** |
| **[試行12で新規発見]** frontend CSP `connect-src` が新オリジンをブロック | Cookie/CORS検証はPASSしてもブラウザ側CSPで別途ブロックされ得る | P4-8実機検証で発見。本番カットオーバー(P1-2 NS切替)時は`frontend/index.html`のCSP `connect-src`に最終オリジンが含まれているか確認する運用チェックをrunbookに追加推奨（試行12ではlocalhost検証用に一時追加→revert済み） |
| **[試行12で新規発見]** Container(Durable Object)インスタンスの env var(vars)反映タイミング | `wrangler deploy`でvars変更してもコンテナimage無変更時は稼働中instanceに即時反映されない（次回コールドスタートまで旧値継続） | 試行12でCORS revert直後に旧設定が数分残存する事象を実測。運用上は「vars変更を伴うdeployの直後は数分の反映待ちが発生し得る」ことをP5(CI/CD)のデプロイ後検証手順に明記推奨 |
| DB 切替時のデータ差分 | カルテデータ欠損 | 凍結ウィンドウ + チェックサム突合（P3-6/P3-7）。RDS を 2 週間保持 |
| CPU 課金の想定超過 | 月額が試算を超える | P4-9 実測 + P6-4 Budget Alert |
| Cookie/CORS の挙動差 | ログイン不能 | P1-6 / P4-8 の二段階検証 |
| 本番移行時の再現性 | 本番展開の手戻り | 全フェーズの実施記録・所要時間を本ドキュメントに追記し、本番移行計画の一次情報とする |

## 10. スコープ外

- 本番環境の移行（本ドキュメントの実施記録を基に別途計画）
- Vercel フロントエンドの Cloudflare Pages 移行（任意・独立して実施可能）
- WAF 有料プラン（Pro/Business）・Data Localization Suite の契約判断
