# デプロイメント・運用ドキュメント (Deployment & Operations)

> **目的**: デプロイメントハブ(環境URL・主要コマンド・関連ドキュメント索引)を提供する。
> **読者**: 全運用者。
> **タイミング**: デプロイ運用開始時。

> **Animal Ekarte**: ステージング・本番環境へのデプロイと安定稼働のためのガイド
> **最新更新**: 2026-10-11（Production Environment binding・ci-policy 観測反映） | **checked-in config**: STG workflow/configあり。backend production job は checked-in（EMR-148）。Environment reviewers・secret・provider の設定・実行・稼働は未確認（UNKNOWN）。実行前に dated receipt を確認する。FrontendはProduction Environment binding実装済み（外部reviewersは要検証）。**live provider状態はUNKNOWNで、実行前にdated receiptが必要**

---

## 1. 稼働環境一覧

| 環境 | Frontend URL | API Base URL | インフラ管理 |
|:---|:---|:---|:---|
| **Staging（設定値）** | [stg.noah-karte.com](https://stg.noah-karte.com) | [api.stg.noah-karte.com/api](https://api.stg.noah-karte.com/api) | checked-in target。稼働状態は実行時確認 |
| **Production（draft設定）** | noah-karte.com（予定） | api.noah-karte.com/api（予定） | backend production job は checked-in（EMR-148）。Environment reviewers・secret・provider の設定・実行・稼働は未確認（UNKNOWN）。実行前に dated receipt を確認する。frontendはEnvironment bindingあり（#253・[`../infra/production/runbook.md`](../infra/production/runbook.md)）。`production` branch protection の required checks 不整合は [ci-policy.md](../ci-policy.md)（2026-10-11 観測）参照 |

---

## 2. ドキュメント体系

目的別に関連ドキュメントを参照してください。

> **新規文書の配置ルール**: 分類軸は「環境 → 対象物 → 活動種別」の優先順で決める。① 特定環境限定（`STG-*`）→ `stg/`。② 対象物が明確 → `accounts/`・`lab-devices/`・`seeds/`（DB・CSV・seed 操作）。③ 検証・リハーサル系 → `rehearsals/`。④ 上記に収まらない実行手順 → `runbooks/`（残余カテゴリ）。デプロイフロー中核の入口文書のみルートに置く。

**デプロイ基本フロー**

- **[デプロイ手順書 (CI-CD-PIPELINE.md)](./CI-CD-PIPELINE.md)**: 自動デプロイと手動トリガーの手順。
- **[事前確認リスト (./DEPLOYMENT_CHECKLIST.md)](./DEPLOYMENT_CHECKLIST.md)**: デプロイ前の動作確認、DBリセット、疎通確認。

**runbooks/（個別実行手順 — 索引: [runbooks/README.md](./runbooks/README.md)）**

- **[リリースマニュアル (STG_PRE_DEPLOY_READINESS_CHECK.md)](./runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md)**: 本番反映前の最終検証ランブック。
- **[外部資格情報オペレーション (BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)](./runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)**: Cloudflare / PlanetScale 資格情報ローテーションと、退役済み AWS 手順の実行禁止境界。
- **[Durable Scheduler運用 (runbooks/SCHEDULER_OPERATIONS.md)](./runbooks/SCHEDULER_OPERATIONS.md)**: 定期jobのstatus、pause/resume、missing-slot catch-up、通知、障害復旧。BE9のcode/configは実装済みだが、今回versionの実環境deploy/rehearsalはrelease gateとして未実施。
- **[定期バッチ一覧と有効化判断 (runbooks/SCHEDULER_BATCHES.md)](./runbooks/SCHEDULER_BATCHES.md)**: バッチ別仕様（no_show / delivery / dormant / plane_sync）と本番 go-live 有効化判断。
- **[Lステップ Write API 一時停止メモ (LSTEP_WRITE_API_PAUSE.md)](./runbooks/LSTEP_WRITE_API_PAUSE.md)**: Lステップへのタグ付与・解除・プロパティ更新を再有効化する前提条件。
- **[休憩時間データ形状監査 (BREAK-HOURS-SHAPE-AUDIT.md)](./runbooks/BREAK-HOURS-SHAPE-AUDIT.md)**: R1-3 デプロイ前の STG/本番 break_hours 形状監査手順。
- **[EMR-321 ホテル予約手動確認 (EMR-321_HOTEL_RESERVATION.md)](./runbooks/EMR-321_HOTEL_RESERVATION.md)**: 標準予約区分の整備とケージ選択不可の医院別調査手順。

**rehearsals/（検証・リハーサル）**

- **[スモークテスト手順 (CRUD-SMOKE-TEST.md)](./rehearsals/CRUD-SMOKE-TEST.md)**: デプロイ直後の主要機能（医院/スタッフ/権限）の導線確認。
- **[混在会計スモークテスト (MIXED-PAYMENT-SMOKE-TEST.md)](./rehearsals/MIXED-PAYMENT-SMOKE-TEST.md)**: 混在会計 (payment_splits) の詳細動作確認。
- **[A4 UI rehearsal isolated stack (A4_UI_REHEARSAL.md)](./rehearsals/A4_UI_REHEARSAL.md)**: 正式21表CSVの画面確認専用localhost-only disposable Compose環境とruntime証跡手順。
- **[F8 G4 synthetic failure rehearsal (F8_G4_FAILURE_REHEARSAL.md)](./rehearsals/F8_G4_FAILURE_REHEARSAL.md)**: 固定synthetic FK違反でtransaction rollback・21表空band・seed preflightを証明する専用disposable runner。
- **[旧] Vercel フロントエンド検証手順 (VERCEL-FRONTEND-STAGING-TEST.md)](./rehearsals/VERCEL-FRONTEND-STAGING-TEST.md)**: Vercel 時代の検証手順。EMR-255 で frontend は Cloudflare Workers Static Assets へ移行済み — 履歴資料としてのみ参照可。Vercel は現行配信・rollback のいずれにも使用しない（現行確認手順は frontend Worker の smoke/tail）。

**seeds/（DB リセット・seed・CSV 投入）**

- **[ローカル DB リセット (LOCAL_DB_RESET.md)](./seeds/LOCAL_DB_RESET.md)**: ローカル開発 DB の再作成・migration 再適用・seed 復元手順。
- **[CSV seed運用 (SEED_MIGRATION_OPERATIONS.md)](./seeds/SEED_MIGRATION_OPERATIONS.md)**: 全環境master-only CSVと `APP_ENV` 別ログインseed、再生成手順、old_db 21表CSVとの境界。
- **[old_db 医院別ローカル隔離 (OLD_DB_HANDOFF_LOCAL.md)](./seeds/OLD_DB_HANDOFF_LOCAL.md)**: 21表CSVを `seeds/_old_db_handoff/<clinic>/` へ置く手順（`make seed` 非対象）。
- **[医院 CSV カットオーバー投入 (CLINIC_CSV_IMPORT.md)](./seeds/CLINIC_CSV_IMPORT.md)**: old_db の21表CSVをmanifest digestに固定し、preflight/apply/verifyするF6手順。

**stg/（STG 環境の継続運用）**

- **[STG 継続運用チェックリスト (STG-CONTINUOUS-OPERATIONS.md)](./stg/STG-CONTINUOUS-OPERATIONS.md)**: 日次/週次/月次の STG 環境監視・検査・メンテナンス。
- **[STG デモデータライフサイクル (STG-DEMO-DATA-LIFECYCLE.md)](./stg/STG-DEMO-DATA-LIFECYCLE.md)**: Seed/Demo/Smoke テストデータの分類、作成元、Cleanup 方針、明示承認が必要な DB 再作成境界。
- **[PlanetScale STG シード投入 Runbook (STG_PLANETSCALE_SEED_RUNBOOK.md)](./stg/STG_PLANETSCALE_SEED_RUNBOOK.md)**: PlanetScale STG の `cmd/migrate` 適用・検証。承認済み再構築（`DB_RESET` → master/login → `_old_db_handoff`）のオペレータ手順は同文書 §6。

**accounts/（アカウント払い出し）**

- **[スタッフアカウント払い出し (STAFF_ACCOUNT_PROVISIONING.md)](./accounts/STAFF_ACCOUNT_PROVISIONING.md)**: 医院スタッフの初期アカウント作成・権限グループ割当・引き渡し手順。
- **[本番初回システム管理者 (FIRST_SYSTEM_ADMIN.md)](./accounts/FIRST_SYSTEM_ADMIN.md)**: 認可済み運用担当者が本番の最初のシステム管理者を 1 回だけ作成する初回専用手順（`staff-provision` は既存 actor 前提でゼロからは作れない）。

**lab-devices/（検査機器）**

- **[検査機器 有線疎通 (LAB_DEVICE_CONNECTIVITY.md)](./lab-devices/LAB_DEVICE_CONNECTIVITY.md)**: 実装契約。手元・医院の操作手順は old_db `docs/lab-go/hospital-field-pack/手元テスト手順.md`。
- **[Mac 検査機器受信機 (LAB_DEVICE_AGENT_MACOS.md)](./lab-devices/LAB_DEVICE_AGENT_MACOS.md)**: bundle 配布・インストールの前提と未完了の配布ゲート。

**関連ドキュメント（deploy/ 外）**

- **[本番 Cloudflare 基盤 事前構築手順 (../infra/production/setup.md)](./../infra/production/setup.md)**: 本番環境（noah-karte.com）を CF Workers + Containers + PlanetScale で新設する実施手順（billing recovery と production approval gate の実装・検証後に USER が日程を決定）。
- **[外部連携棚卸し (CLOUDFLARE-EXTERNAL-INTEGRATIONS-AUDIT.md)](../infra/CLOUDFLARE-EXTERNAL-INTEGRATIONS-AUDIT.md)**: LINE / Lステップ / SMTP / LIFF の egress 依存棚卸しと、LINE webhook redelivery・error 統計の release pending 項目（infra/ へ移設済み）。
- **[Delete / Soft Delete 設計パターン](../../architecture/delete-soft-delete-patterns.md)**: Hard Delete と Soft Delete の使い分け、FK 制約との関係、実装パターン、STG-001 教訓。

> PR #49 Post-Merge Smoke Checklist（PR固有の使い切りチェックリスト）・CRUD スモーク自動化戦略
> （§3.4 に記載の通り自動化自体が撤去済みで計画倒れとなった歴史的記録）は特定PR/時点のスナップショット
> のため退役済み（git 履歴参照。旧 docs/archive/ は削除済み）。

---

## 3. クイック・コマンドリファレンス

### 3.1 サービス状態の確認

```bash
curl -sS -o /dev/null -w '%{http_code}\n' https://api.stg.noah-karte.com/health
```

期待値は `200`。失敗時は [`../infra/staging/runbook.md`](../infra/staging/runbook.md) に従い、
workers.dev の `/health` と実 URL を比較して DNS / Worker / Container を切り分けます。

### 3.2 ログの確認

Cloudflare Dashboard の Workers Logs / Containers と、対象の GitHub Actions run を確認します。
Workers Logs はインフラ障害調査用で、カルテの変更監査は DB の `audit_logs` が正本です。

### 3.3 手動デプロイの実行（共有STG・人間承認必須）

**HARD STOP:** named human owner/approval、review済みcommit、refがそのcommitへ解決すること、対象Worker/config、secret scope、共有STG利用可否を記録する。未確認なら実行しない。正本は [staging runbook](../infra/staging/runbook.md)。

```bash
REVIEWED_SHA='<reviewed-commit>'
REMOTE_SHA="$(gh api 'repos/{owner}/{repo}/git/ref/heads/staging' --jq '.object.sha')"
test "$REMOTE_SHA" = "$(git rev-parse "${REVIEWED_SHA}^{commit}")" || exit 1
gh workflow run backend-deploy.yml --ref staging
```

dispatch後は対象runの`headSha`が`REVIEWED_SHA`と一致することをnames-only metadataで確認し、不一致なら停止する。

### 3.4 自動スモークテストの実行 (手動トリガー)
```bash
# STG 疎通確認（/health）
gh workflow run stg-smoke.yml
```
> 旧 stg-health-check / stg-readonly-smoke / stg-crud-smoke の3本は `stg-smoke.yml` に統合後、
> login/readonly/CRUD は `STG_DEMO_*` secret 未設定で約1年間機能していなかったためデッドコードとして撤去
> （現状は health 疎通のみ）。CRUD の正しさは backend unit/integration テスト + FE route-guard テストでカバー。
> CRUD smoke を復活させる場合は `STG_DEMO_EMAIL`/`STG_DEMO_PASSWORD` を設定し git 履歴 `281a561e` を参照。

---

## 4. デプロイ後の障害判定と復旧

> AWS ECS/RDS は廃止済みで、切り戻し先やホットスタンバイはありません。
> 現行 STG の障害初動は [`../infra/staging/runbook.md`](../infra/staging/runbook.md) が正本です。

### 4.1 ヘルスチェック手順

デプロイ完了直後、以下を確認する（実行コマンドは §3.1）。経路切り分け・復旧の正本は [`../infra/staging/runbook.md`](../infra/staging/runbook.md) の Incident triage — 本書はチェックリスト層のみを扱う。

- [ ] 実 URL の `/health` が `200` / `"ok"`（失敗時は workers.dev の `/health` と比較し、runbook の Incident triage で DNS / route と Worker / Container / DB を切り分け）
- [ ] `backend-deploy.yml` の deploy / migrate / post-migrate health を Actions の step 単位で確認
- [ ] イメージ更新を伴う場合は 15 分静置後に再確認
- [ ] Workers Logs / Containers に ERROR/FATAL なし（全断 + DB 接続エラーなら `DB_MAX_OPEN_CONNS` と PlanetScale 接続枯渇を確認）

---

### 4.2 リリース中止・復旧開始の判定基準

以下のいずれかに該当した場合、リリース成功とはせず、チームへ通知して復旧を開始します。

| # | 症状 | 判定方法 |
|---|------|--------|
| 1 | `/health` 非応答 | 実 URL / workers.dev の HTTP status |
| 2 | deploy / migrate / post-migrate health の失敗 | GitHub Actions の step |
| 3 | Workers / Containers のエラー急増 | Cloudflare Observability |
| 4 | CRUD スモークテスト想定外エラー | 想定外 400/500 response |
| 5 | FK 保護が 409 ではなく想定外の 4xx/5xx | DELETE 試行時 status |
| 6 | データ破損・想定外削除・テナント隔離破綻 | smoke 後の手動確認 |

**復旧フロー**: 切り分け・復旧手順の正本は [`../infra/staging/runbook.md`](../infra/staging/runbook.md) の Incident triage。復旧開始時に影響範囲と開始時刻をチームへ通知し、runbook に従って Cloudflare / rolling update / PlanetScale を切り分けて修正・再デプロイする（基盤喪失時はスナップショットと現行 IaC から再建。AWS 切り戻し先はない）。health と smoke を再確認した後、インシデント記録へ原因・復旧内容・再発防止を残す。

---

### 4.3 デプロイ成功の条件

以下の 3 つがすべて成立した場合、デプロイ成功と判定し、運用へ移行します。

| 条件 | 確認方法 | 判定 |
|------|--------|------|
| **ヘルスチェック PASS** | §4.1 をすべて通過 | ✅ |
| **CRUD スモークテスト PASS** | [CRUD-SMOKE-TEST.md](./rehearsals/CRUD-SMOKE-TEST.md) の A-1〜A-3、B-1〜B-3、C-1〜C-3 を実行し、HTTP/resource state と明示された監査契約を確認 | ✅ |
| **テストデータ削除完了・記録済み** | 同手順の cleanup で、その run が作成した ID だけを削除し、resource state と実装済みの監査だけを記録 | ✅ |

**3 つすべて ✅ の場合**: STG デプロイ成功。Production は別途 workflow / Environment protection / 実環境の production readiness 証拠が必要。

---

### 4.4 認証情報保護ポリシー

デプロイ検証時、以下のいずれも **本ドキュメント・成果物・ログ出力に記録してはいけません**：

- `password` （パスワード）
- `access_token` （アクセストークン）
- `refresh_token` （リフレッシュトークン）
- Cookie の `Set-Cookie` ヘッダ値
- demo アカウントの token / cookie 値

**実装方法**:
- API 検証時は、ブラウザ DevTools の Network タブで Cookie を確認し、スクリプト/出力には含めない
- curl 実行時は `${TOKEN}` 等の環境変数を使用し、実トークン値を可視化しない
- Workers Logs / Containers のログにパスワード値が出力されていないことを確認

---

### 4.5 参考資料

- [STG 運用 Runbook](../infra/staging/runbook.md)：現行インフラの障害初動・復旧方針
- [デプロイ手順書 (CI-CD-PIPELINE.md)](./CI-CD-PIPELINE.md)：自動デプロイ・手動トリガー
- [スモークテスト手順 (CRUD-SMOKE-TEST.md)](./rehearsals/CRUD-SMOKE-TEST.md)：CRUD 全操作・FK保護検証・権限テスト
- [外部資格情報オペレーション](./runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)：Cloudflare / PlanetScale の資格情報ローテーション

---
