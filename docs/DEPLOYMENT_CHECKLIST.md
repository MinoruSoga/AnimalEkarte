# デプロイメント準備チェックリスト

> **バージョン**: 1.0  
> **最終更新**: 2026-04-23  
> **ステータス**: ✅ READY FOR PRODUCTION  
> **対象**: v2.0 本番環境デプロイ

---

## 📋 概要

Animal Ekarte v2.0 は、TIER 1-3 のすべての実装が完了し、本番デプロイに向けて準備完了状態にあります。本チェックリストは、デプロイ前に検証すべきすべての項目を記載しています。

---

## ✅ インフラストラクチャ準備

### クラウド環境（AWS）

- [ ] VPC / セキュリティグループ設定
  - [ ] RDS PostgreSQL 18 インスタンス起動
  - [ ] ECS / Fargate クラスタ構成
  - [ ] ALB / NLB ロードバランサ設定
  - [ ] S3 バケット（ファイルアップロード用）
  - [ ] CloudFront CDN（静的資産配信）

- [ ] ネットワーク設定
  - [ ] SSL/TLS 証明書（ACM）
  - [ ] Route 53 DNS 設定
  - [ ] WAF ルール設定（DDoS 対策）

- [ ] ログ・監視
  - [ ] CloudWatch ログストリーム
  - [ ] CloudWatch アラーム設定
  - [ ] X-Ray トレーシング設定

### データベース

- [ ] PostgreSQL 18 初期化
  - [ ] Character Set: UTF-8
  - [ ] Timezone: UTC
  - [ ] Backup: 自動バックアップ (日次)

- [ ] マイグレーション実行
  - [ ] 001_initial_schema.sql ✅ PASS
  - [ ] 002_add_clinic_id.sql ✅ PASS
  - [ ] 003_add_indexes.sql ✅ PASS
  - [ ] 004_seed_staging.sql ✅ (本番用シード)

- [ ] DB 接続確認
  - [ ] コネクションプール設定 (Max: 100)
  - [ ] Backup エンドポイント設定
  - [ ] 読み取りレプリカ設定（オプション）

---

## ✅ アプリケーション準備

### バックエンド (Go API)

#### ビルド & デプロイ

- [ ] Docker イメージビルド
  ```bash
  docker build -f backend/Dockerfile -t animal-ekarte-api:v2.0 .
  docker tag animal-ekarte-api:v2.0 {aws-account}.dkr.ecr.{region}.amazonaws.com/animal-ekarte-api:v2.0
  docker push {aws-account}.dkr.ecr.{region}.amazonaws.com/animal-ekarte-api:v2.0
  ```

- [ ] 環境変数設定
  - [ ] `DATABASE_URL`: PostgreSQL 接続文字列
  - [ ] `JWT_SECRET`: 本番用秘密鍵 (32+ 文字)
  - [ ] `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`: メール設定
  - [ ] `STORAGE_TYPE`: `s3` (ローカルではなく)
  - [ ] `S3_BUCKET`, `S3_REGION`: S3 設定
  - [ ] `FRONTEND_URL`: フロントエンド URL (https://noah-karte.com)
  - [ ] `GIN_MODE`: `release`
  - [ ] `LOG_LEVEL`: `info`

#### テスト & 検証

- [ ] バックエンド テスト実行
  ```bash
  go test ./... -v -cover
  ```
  - [ ] PASS 率 >= 95%
  - [ ] カバレッジ >= 70%

- [ ] Go リンター実行
  ```bash
  golangci-lint run ./...
  ```
  - [ ] No critical errors

- [ ] 本番ビルド検証
  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api cmd/api/main.go
  ```

#### セキュリティ検証

- [ ] JWT シークレット検証
  - [ ] 本番用秘密鍵が環境変数に設定済み
  - [ ] リポジトリに秘密鍵が含まれていない

- [ ] SQL インジェクション対策
  - [ ] すべてのクエリが parameterized (GORM)
  - [ ] NULL バイト除去済み (BUG-067)

- [ ] CORS 設定
  - [ ] 本番 Frontend URL のみ許可
  - [ ] 認証情報ヘッダの確認

- [ ] レート制限
  - [ ] ログイン: 5回/分
  - [ ] パスワード忘却: 3回/分

### フロントエンド (React)

#### ビルド & デプロイ

- [ ] ビルド実行
  ```bash
  pnpm build
  ```
  - [ ] ビルド成功 (0 エラー)
  - [ ] 警告: 最小限（shadcn/ui 標準警告のみ）

- [ ] ビルドサイズ確認
  ```bash
  pnpm build -- --stats
  ```
  - [ ] Main bundle < 150KB (gzip)
  - [ ] Total < 250KB (gzip)

- [ ] 環境変数設定 (`.env.production`)
  - [ ] `VITE_API_BASE_URL`: https://api.noah-karte.com
  - [ ] `VITE_FRONTEND_URL`: https://noah-karte.com
  - [ ] `VITE_ENV`: production

- [ ] Docker イメージビルド
  ```bash
  docker build -f frontend/Dockerfile -t animal-ekarte-web:v2.0 .
  docker push {ecr-uri}/animal-ekarte-web:v2.0
  ```

#### テスト & 検証

- [ ] フロントエンド テスト実行
  ```bash
  pnpm test:run
  ```
  - [ ] PASS 率 = 100% (465 tests)

- [ ] TypeScript 型チェック
  ```bash
  pnpm type-check
  ```
  - [ ] No type errors

- [ ] リンター実行
  ```bash
  pnpm lint
  ```
  - [ ] No critical errors

- [ ] Lighthouse 監査
  ```bash
  node frontend/scripts/lighthouse-audit.js --url https://noah-karte.com
  ```
  - [ ] Performance > 75
  - [ ] Accessibility > 90
  - [ ] Best Practices > 90
  - [ ] SEO > 90

#### セキュリティ検証

- [ ] XSS 対策
  - [ ] すべてのユーザー入力が sanitized
  - [ ] HTML エスケープ適用

- [ ] CSRF 対策
  - [ ] JWT トークン用 httpOnly Cookie
  - [ ] SameSite=Strict 設定

- [ ] CSP ヘッダ
  - [ ] Content-Security-Policy: strict-dynamic

---

## ✅ パフォーマンス検証

### 負荷テスト

- [ ] k6 API エンドポイントテスト実行
  ```bash
  k6 run load-tests/k6-api-endpoints.js --vus 50 --duration 5m
  ```
  - [ ] p95 < 500ms
  - [ ] p99 < 1000ms
  - [ ] Error rate < 10%
  - [ ] Requests/sec > 100

- [ ] k6 スパイクテスト実行
  ```bash
  k6 run load-tests/k6-spike-test.js
  ```
  - [ ] Spike 時 p95 < 2000ms
  - [ ] Error rate < 20%
  - [ ] 復帰時間 < 60秒

### プロファイリング

- [ ] メモリプロファイル実行
  - [ ] Peak memory < 500MB
  - [ ] Stable memory < 300MB

- [ ] CPU プロファイル実行
  - [ ] No hot spots
  - [ ] GC pause < 100ms

---

## ✅ セキュリティ監査

### 認証・認可

- [ ] JWT トークン実装
  - [ ] Expiration: 1 hour
  - [ ] Refresh token: 7 days
  - [ ] HttpOnly Cookie 有効

- [ ] RBAC 権限制御
  - [ ] 3 ロール (獣医師・助手・受付) 確認
  - [ ] リソース削除前の FK チェック

- [ ] パスワードセキュリティ
  - [ ] bcrypt ハッシング (cost: 12)
  - [ ] 複雑性要件なし（UX 優先）

### API セキュリティ

- [ ] API キー・秘密管理
  - [ ] JWT_SECRET は AWS Secrets Manager に保存
  - [ ] S3 認証情報は IAM ロール経由

- [ ] レート制限
  - [ ] ログイン: 5回/分
  - [ ] パスワード忘却: 3回/分

- [ ] CORS 設定
  - [ ] 許可オリジン: https://noah-karte.com のみ

### データセキュリティ

- [ ] 暗号化
  - [ ] TLS 1.3 (転送中)
  - [ ] RDS 暗号化 (保存時)

- [ ] バックアップ
  - [ ] 日次自動バックアップ
  - [ ] 別リージョンレプリケーション

- [ ] ロギング
  - [ ] Access logs: CloudWatch
  - [ ] Error logs: CloudWatch
  - [ ] 個人情報: マスキング済み

---

## ✅ 運用準備

### モニタリング & アラート

- [ ] CloudWatch ダッシュボード
  - [ ] API レスポンスタイム
  - [ ] エラー率
  - [ ] DB 接続数
  - [ ] メモリ使用率

- [ ] アラート設定
  - [ ] Error rate > 5% → Alert
  - [ ] Response time p95 > 1000ms → Alert
  - [ ] Memory > 400MB → Alert
  - [ ] DB connections > 80 → Alert

### ロギング

- [ ] CloudWatch Logs
  - [ ] Backend logs: /aws/ecs/animal-ekarte-api
  - [ ] Frontend logs: CloudFront access logs
  - [ ] Database logs: RDS logs

- [ ] ログレベル
  - [ ] Info: 重要な処理
  - [ ] Warning: 異常検知
  - [ ] Error: エラー追跡

### バックアップ & 復旧

- [ ] RDS Backup
  - [ ] 自動バックアップ: 7日間保持
  - [ ] マニュアルスナップショット: 本番前実行

- [ ] 復旧計画
  - [ ] RTO (Recovery Time Objective): 1時間
  - [ ] RPO (Recovery Point Objective): 1日

### ドキュメント整備

- [ ] ランブック (Runbook)
  - [ ] 障害対応手順
  - [ ] ホットフィックス手順
  - [ ] ロールバック手順

- [ ] API ドキュメント
  - [ ] Swagger UI: https://api.noah-karte.com/swagger/
  - [ ] OpenAPI YAML: docs/openapi.yaml

---

## ✅ 統合テスト

### システム統合テスト

- [ ] フロントエンド + バックエンド連携
  - [ ] ログイン フロー
  - [ ] 飼主作成・更新・削除
  - [ ] 予約作成・取得
  - [ ] 医療記録 CRUD

- [ ] バックエンド + データベース連携
  - [ ] トランザクション管理
  - [ ] FK 制約チェック
  - [ ] インデックス動作確認

- [ ] インフラ統合
  - [ ] ALB ヘルスチェック
  - [ ] RDS フェイルオーバー (オプション)

### 本番環境シミュレーション

- [ ] E2E テスト実行 (本番環境相当)
  ```bash
  pnpm test:e2e -- --headed
  ```
  - [ ] PASS 率 >= 90%

- [ ] 負荷テスト (本番トラフィック予測)
  ```bash
  k6 run load-tests/k6-api-endpoints.js --vus 100
  ```

---

## ✅ リリースチェック

### バージョン & ドキュメント

- [ ] バージョン番号確認
  - [ ] backend: v2.0.0
  - [ ] frontend: v2.0.0
  - [ ] API: /api/v1

- [ ] CHANGELOG 記載
  - [ ] 新機能一覧
  - [ ] 修正バグ一覧
  - [ ] 既知の制限事項

- [ ] リリースノート公開
  - [ ] 機能説明
  - [ ] 使用方法
  - [ ] サポート情報

### マイグレーションチェック

- [ ] DB スキーマ確認
  - [ ] 45 テーブル作成完了
  - [ ] インデックス作成完了
  - [ ] FK 制約設定完了

- [ ] 初期データ投入
  - [ ] master データ投入
  - [ ] テストデータ (オプション)

### デプロイプロセス

- [ ] Blue-Green デプロイ準備
  - [ ] Blue 環境: 現行版
  - [ ] Green 環境: 新版 (v2.0)
  - [ ] トラフィック切り替え計画

- [ ] ロールバック計画
  - [ ] 前バージョン (v1.x) イメージ保持
  - [ ] ロールバック手順書

---

## 📋 最終確認チェックリスト

- [ ] すべてのテスト PASS
- [ ] セキュリティ監査 完了
- [ ] パフォーマンス目標 達成
- [ ] ドキュメント完成
- [ ] チーム内確認 完了
- [ ] ステークホルダー 承認
- [ ] バックアップ 実施済み
- [ ] ロールバック 計画書 作成完了

---

## 🚀 デプロイ実行

本チェックリストすべての項目が完了した場合、以下の流れでデプロイ実行：

### Phase 1: ステージング環境デプロイ

```bash
# Green 環境を ステージング にデプロイ
terraform apply -var="environment=staging" -var="version=v2.0"

# 統合テスト実行
pnpm test:e2e -- --baseUrl https://stg.noah-karte.com

# パフォーマンス確認
k6 run load-tests/k6-api-endpoints.js
```

### Phase 2: 本番環境デプロイ

```bash
# Blue-Green デプロイ: Green 環境を本番に昇格
terraform apply -var="environment=production" -var="active_environment=green"

# ヘルスチェック
curl https://api.noah-karte.com/health

# トラフィック切り替え (5% → 50% → 100%)
aws elbv2 modify-rule --rule-arn arn:aws:elasticloadbalancing:... \
  --actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:...:targetgroup/green/...
```

### Phase 3: 監視 & 検証

```bash
# ダッシュボード監視 (24時間)
- Error rate < 1%
- Response time p95 < 500ms
- Memory < 300MB
- CPU < 30%

# ビジネスKPI 確認
- 予約作成フロー: 正常
- 医療記録作成: 正常
- ユーザー ログイン: 正常
```

---

**デプロイ完了時刻**: ___________  
**担当者署名**: ___________  
**確認者署名**: ___________
