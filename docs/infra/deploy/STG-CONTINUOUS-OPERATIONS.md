# ステージング環境・継続運用チェックリスト (Continuous Operations)

> **目的**: STG環境の日次/週次/月次運用チェックを定義する。
> **読者**: DevOps/SRE。
> **タイミング**: 日次/週次/月次の定期運用時。

> **Animal Ekarte**: STG 環境の日常・定期運用のための検査・監視・メンテナンスチェックリスト
> **最新更新**: 2026-07-10 | **目的**: 本番リリース間の STG 環境安定稼働確保

---

> **注記(2026-07-10)**: バックエンドの正系統は Cloudflare Workers + Containers + PlanetScale Postgres
> （`backend-deploy.yml`）へ移行済み。本ドキュメントの ECS/CloudWatch/RDS ベースの手順（§2.2, §2.3,
> §4.1〜4.3, §6 のトラブルシューティング診断コマンド）は、旧 AWS ECS ロールバック経路
> （`backend-deploy-ecs.yml`、`workflow_dispatch` のみ）にのみ適用される。Cloudflare 正系統向けの
> 同等の日次/週次/月次運用手順（Workers Logs 監視等）は未文書化 — 現状のギャップとして
> `migration-cloudflare.md` Phase 6 の記録を参照すること。§1・§2.1・§3・§5 は経路に依存せず有効。

## 1. 目的と対象読者

本ドキュメントは、デプロイ完了後の STG 環境を継続的に監視・運用するためのチェックリストです。

**対象者**: DevOps / SRE / Team Lead
**実施頻度**: 
- 日次: health check（デプロイ直後）
- 週次: 全体検査・demo account 検証
- 月次: データベース健全性・ログ監査

---

## 2. 日次オペレーション（デプロイ直後）

### 2.1 API ヘルスチェック

```bash
export API_HEALTH=https://api.stg.noah-karte.com/health
curl -s ${API_HEALTH} | jq '.'
```

**期待結果**:
```json
{
  "status": "ok",
  "timestamp": "2026-05-27T02:00:00Z"
}
```

**失敗時アクション**:
- HTTP 200 が返らない場合、`docs/infra/deploy/README.md` §4.2 ロールバック判定へ

---

### 2.2 ECS サービス稼働確認

```bash
export AWS_PROFILE=AnimalEkarte

aws ecs describe-services \
  --cluster animalekarte-stg-cluster \
  --services animalekarte-stg-service \
  --region us-east-1 \
  --query 'services[0].{desiredCount,runningCount,status}'
```

**期待結果**:
```
{
  "desiredCount": 2,
  "runningCount": 2,
  "status": "ACTIVE"
}
```

**確認ポイント**: 
- `desiredCount` == `runningCount`（不一致時は `§4.2` ロールバック判定へ）
- `status` が `ACTIVE`

---

### 2.3 CloudWatch エラーログ監視

```bash
# 過去 5 分間のエラーログをフォロー
aws logs tail /ecs/animalekarte-stg \
  --region us-east-1 \
  --follow \
  --since 5m | grep -i "error\|fatal"
```

**期待結果**: 
- ERROR/FATAL ログが 5 分間で 3 件以下

**異常基準**:
- ERROR/FATAL が 3 件以上（即座にロールバック）

**参考**: [README.md §4.1](./README.md#41-ヘルスチェック手順)

---

## 3. 週次検査（毎週木曜 9:00）

### 3.1 Vercel フロントエンド表示確認

**環境**: [https://stg.noah-karte.com](https://stg.noah-karte.com)

- [ ] ページ読み込み成功（5 秒以内）
- [ ] ログイン画面表示
- [ ] CSS/レイアウト崩れなし
- [ ] エラーコンソールなし（ブラウザ DevTools F12 → Console）
- [ ] ネットワークエラーなし（Network タブ）

**失敗時アクション**: [VERCEL-FRONTEND-STAGING-TEST.md](./VERCEL-FRONTEND-STAGING-TEST.md) の §3 トラブルシューティングを参照

---

### 3.2 Demo アカウントログイン確認

**アカウント情報**: Stone に保存済み（[CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) §2.2 参照）

手順:
1. ブラウザで [https://stg.noah-karte.com](https://stg.noah-karte.com) にアクセス
2. `system_admin` 権限を持つ demo アカウントでログイン
3. ダッシュボード表示確認
4. Settings → 医院マスタ / 権限グループ / スタッフマスタ にアクセス可能か確認

**期待結果**:
- ログイン成功
- ダッシュボード表示
- Settings 画面すべてへのアクセス可能

**失敗時アクション**: 
- Cookie 検証（F12 → Application → Cookies）
- Token 期限確認（`refresh_token` が残っているか）

参考: [VERCEL-FRONTEND-STAGING-TEST.md §4](./VERCEL-FRONTEND-STAGING-TEST.md#4-demo-アカウントログイン検証)

---

### 3.3 CRUD スモークテスト実行

**実施内容**: [CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) を完全実行

**実施者**: Team Lead または指定デプロイ担当者

**確認項目**:
- [ ] 医院 (Clinics) 系統: A-1 ～ A-4
- [ ] 権限グループ (Permission Groups) 系統: B-1 ～ B-3
- [ ] スタッフ (Staffs) 系統: C-1 ～ C-4

**期待結果**: 全項目がステータスコード期待値を返す

詳細は [CRUD-SMOKE-TEST.md §4](./CRUD-SMOKE-TEST.md#4-ステータスコード期待値表) を参照

---

### 3.4 テストデータ削除確認

**対象**: CRUD スモークテストで作成した test data

**削除手順**:
1. API 経由での段階的削除（推奨）
2. Cleanup コマンド実行（staff → permission-groups → clinics 順序で DELETE）
3. GET で 404 or リストから除外されたことを確認

詳細: [CRUD-SMOKE-TEST.md §6](./CRUD-SMOKE-TEST.md#6-テスト-データの削除ポリシー)

**確認チェック**:
- [ ] Test data が削除された
- [ ] 削除確認ログが記録されている
- [ ] audit_logs テーブルに DELETE レコード存在

---

## 4. 月次検査（毎月第 1 金曜 10:00）

### 4.1 CloudWatch ログ監査

```bash
# 過去 30 日間のエラー統計
aws logs start-query \
  --log-group-name /ecs/animalekarte-stg \
  --region us-east-1 \
  --start-time $(($(date +%s) - 2592000)) \
  --end-time $(date +%s) \
  --query-string 'fields @timestamp, @message | filter @message like /ERROR|FATAL/ | stats count()'
```

**期待結果**: 
- 過去 30 日間の ERROR/FATAL が 20 件以下
- 明らかなパターンがない場合は正常動作

**異常基準**:
- エラー急増（突然 10x 増加など）
- 繰り返しパターン（同じエラーが毎日）

**参考**: AWS CloudWatch Logs Insights

---

### 4.2 ECS タスク履歴確認

```bash
aws ecs describe-services \
  --cluster animalekarte-stg-cluster \
  --services animalekarte-stg-service \
  --region us-east-1 \
  --query 'services[0].deployments'
```

**確認ポイント**:
- 過去 30 日間のデプロイ回数
- 各デプロイで running count に落ち込みがないか
- rollback 発生の有無

---

### 4.3 DB ディスク容量確認

```bash
# RDS DB インスタンス容量確認（AWS Console 経由）
```

**確認内容**:
- allocated storage の使用率（80% 未満が目標）
- free storage space（下限 20 GB 以上推奨）

**異常時アクション**:
- Team Lead に報告
- 必要に応じて storage scale up 検討

---

### 4.4 Demo Data 整合性チェック

**対象**: Seed data（migrations で作成された demo clinics / staffs / permission-groups）

```bash
# 例: demo clinic の存在確認
curl -X GET "https://api.stg.noah-karte.com/api/v1/clinics" \
  -b "access_token=${TOKEN}" \
  -H "Accept: application/json" | jq '.data | length'
```

**期待結果**:
- Demo clinic が 3 件以上存在
- 各 clinic が有効状態（is_deleted = false）

---

## 5. 監査ログ検証（全実行後）

### 5.1 実施ログ記録

**各検査実施時に記録すべき項目**:

```markdown
## 2026-05-27 週次検査実行ログ

- **実施者**: MinoruSoga
- **実施時刻**: 2026-05-27 09:00 JST
- **ヘルスチェック**: ✅ PASS
- **ECS 稼働確認**: ✅ PASS (desired=2, running=2)
- **CloudWatch**: ✅ 2 ERROR ログのみ（正常）
- **Vercel フロントエンド**: ✅ PASS
- **Demo アカウントログイン**: ✅ PASS
- **CRUD スモークテスト**: ✅ PASS（全 11 項目）
- **テストデータ削除**: ✅ PASS（4 レコード削除確認）
- **所見**: 異常なし
```

### 5.2 ログ保存先

- **デプロイ直後ログ**: `/tmp/stg-deploy-check-$(date +%Y%m%d).log`
- **週次・月次ログ**: チーム内 wiki / Confluence など

---

## 6. トラブルシューティング

### Issue 1: `/health` エンドポイント非応答

**症状**: `curl: (7) Failed to connect to...` または HTTP 503

**診断**:
```bash
# ECS サービス確認
aws ecs describe-services \
  --cluster animalekarte-stg-cluster \
  --services animalekarte-stg-service \
  --region us-east-1

# CloudWatch ログ確認
aws logs tail /ecs/animalekarte-stg --follow --since 10m
```

**対処**:
- running count < desired count の場合: ロールバック検討（[README.md §4.2](./README.md#42-ロールバック-要否判定基準)）
- ERROR ログが多発している場合: デプロイ版の問題の可能性 → ロールバック

---

### Issue 2: CloudWatch エラーログ多発

**症状**: `grep -i error | wc -l` で 5 分間に 5 件以上

**診断**:
```bash
# エラー内容を詳細確認
aws logs tail /ecs/animalekarte-stg --follow | grep -i error | head -20
```

**対処**:
- DB 接続エラー: RDS インスタンス状態確認（reboot 検討）
- Memory エラー: ECS task memory 不足 → task 定義変更検討
- Permission エラー: IAM role 権限確認（SSM Parameter Store へのアクセスなど）

---

### Issue 3: Demo アカウントログイン失敗

**症状**: 401 Unauthorized または redirect loop

**診断**:
```bash
# API 認証エンドポイント疎通確認
curl -X POST "https://api.stg.noah-karte.com/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@example.com","password":"..."}'
```

**対処**:
- Token 期限切れ: 新しい token 取得（browser DevTools → Network で確認）
- Demo account disabled: DB で staff.is_deleted = true を確認、false に修正
- Backend API 接続エラー: ECS task 再起動

---

## 7. 参考資料

- [README.md](./README.md) - デプロイメント・運用ドキュメント（ロールバック判定、コマンド集）
- [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md) - 自動デプロイ・手動トリガー・ロールバック手順
- [CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) - CRUD 操作テスト詳細
- [VERCEL-FRONTEND-STAGING-TEST.md](./VERCEL-FRONTEND-STAGING-TEST.md) - フロントエンド検証
- [STG-DEMO-DATA-LIFECYCLE.md](./STG-DEMO-DATA-LIFECYCLE.md) - データ分類とライフサイクル
