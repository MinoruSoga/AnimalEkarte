# ステージング環境・継続運用チェックリスト (Continuous Operations)

> **目的**: STG環境の日次/週次/月次運用チェックを定義する。
> **読者**: DevOps/SRE。
> **タイミング**: 日次/週次/月次の定期運用時。

> **Animal Ekarte**: STG 環境の日常・定期運用のための検査・監視・メンテナンスチェックリスト
> **最新更新**: 2026-07-23 | **目的**: 本番リリース間の STG 環境安定稼働確保

---

> **現行構成**: Cloudflare Workers + Containers + PlanetScale Postgres（`backend-deploy.yml`）。
> AWS ECS/RDS は廃止済みで、切り戻し先・ホットスタンバイではない。
> 構成と障害初動の正本は [`../infra/architecture.md`](../infra/architecture.md) と
> [`../infra/staging/runbook.md`](../infra/staging/runbook.md)。

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
- HTTP 200 が返らない場合、[`../infra/staging/runbook.md`](../infra/staging/runbook.md) の障害初動へ

---

### 2.2 Cloudflare 経路の稼働確認

```bash
curl -sS -o /dev/null -w '%{http_code}\n' https://api.stg.noah-karte.com/health
```

**期待結果**: HTTP `200`

**確認ポイント**:

- 実 URL が失敗した場合は workers.dev の `/health` も確認し、DNS / route と Worker / Container を切り分ける
- イメージ更新を伴うデプロイ後は旧イメージが残りうるため、15 分静置後にも再確認する
- GitHub Actions `backend-deploy.yml` の deploy / migrate / post-migrate health がすべて成功していること

---

### 2.3 Workers / Containers エラーログ監視

Cloudflare Dashboard の Workers Logs / Containers で、デプロイ時刻以降の ERROR/FATAL と
request error を確認する。Workers Logs はインフラ障害調査用で、診療記録の変更監査は
DB の `audit_logs` が正本。

**異常時**:

- 全断 + DB 接続エラーなら、`DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` と PlanetScale 接続枯渇を確認
- Cloudflare 側の障害情報は https://www.cloudflarestatus.com/ で確認
- 切り分けと復旧は [STG 運用 Runbook](../infra/staging/runbook.md) に従う

---

### 2.4 監査書込失敗 (`audit_write_failed`) 監視（PERF-AUDIT-TX P1）

> `auditService.Log`（`backend/internal/service/audit_service.go`）は `repo.Create` 失敗時に統一キー
> `audit_write_failed` を `slog.ErrorContext` で記録する（分散していた呼び出し側の既存 Warn ログとは別に、
> 全 best-effort 監査書込経路を中央 1 箇所で捕捉する）。P2 outbox は見送り決定済みのため、現状はこのログ
> 検索が監査欠落検知の唯一の手段。

Cloudflare Dashboard の Workers Logs / Containers で `audit_write_failed` を検索する。
Container stdout は Worker の request log と別の表示面になる場合があるため、両方を確認する。

**期待結果**: 通常時は 0 件。

**異常基準**: `audit_write_failed` が観測された場合、再開条件（[`phase2.html`](../../../phase2.html)「BE — 見送り・次期タスク」の PERF-AUDIT-TX P2 参照）に従い、月 1 件以上の継続観測があれば outbox 再起案の実測根拠として記録する。

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

### 4.1 Cloudflare ログ・通知監査

Cloudflare Observability で Workers / Containers の ERROR/FATAL、5xx 傾向、通知履歴を確認する。
Workers Logs の保持期間を超える比較が必要な場合は、実施記録側に月次集計を残す。

**異常基準**:
- エラー急増（突然 10x 増加など）
- 繰り返しパターン（同じエラーが毎日）
- 通知先未検証、または通知が期待どおり届かない

---

### 4.2 デプロイ履歴確認

GitHub Actions の `backend-deploy.yml` 実行履歴を確認する。

**確認ポイント**:
- 過去 30 日間のデプロイ回数
- deploy / migrate / post-migrate health の失敗・再実行の有無
- paths 条件や skipped job を成功証跡と誤認していないか
- 障害復旧の再デプロイが記録されているか

---

### 4.3 PlanetScale DB 健全性確認

**確認内容**:
- 容量・接続数・バックアップ状態
- Worker/Container の `DB_MAX_OPEN_CONNS=10` / `DB_MAX_IDLE_CONNS=5` が維持されていること
- 既存テーブルへの ALTER 系 migration 前に、失効ロール所有問題が解消済みであること

**異常時アクション**:
- Team Lead に報告
- 接続枯渇時は新規デプロイを重ねず、滞留接続のドレインと設定値を確認
- 所有権問題は PlanetScale サポートへの REASSIGN 依頼を先に完了

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
- **Cloudflare 経路**: ✅ PASS（実 URL / workers.dev）
- **Workers / Containers Logs**: ✅ 異常なし
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

1. 実 URL と workers.dev の `/health` を比較する
2. GitHub Actions `backend-deploy.yml` の直近 run を step 単位で確認する
3. Cloudflare Dashboard の Workers Logs / Containers を確認する
4. Cloudflare status と PlanetScale 接続枯渇を確認する

**対処**:
- デプロイ直後なら 15 分静置して旧イメージ残留を除外する
- Cloudflare 側の修正・再デプロイを行う
- 基盤喪失時はスナップショットと現行 IaC から再建する（AWS への切り戻し先はない）

---

### Issue 2: Workers / Containers エラーログ多発

**症状**: Cloudflare Observability で ERROR/FATAL または 5xx が急増

**診断**:

- Worker request log と Container log を分けて確認する
- DB 接続エラー、Container の起動/OOM、権限・secret 不足を分類する
- `backend-deploy.yml` の migrate / health step と発生時刻を突合する

**対処**:
- DB 接続エラー: PlanetScale 状態と接続スロット、10/5 の接続プール設定を確認
- Container 起動エラー: 直近デプロイ差分と Cloudflare Containers の観測結果を確認
- Permission / secret エラー: Cloudflare Worker secrets と GitHub Secrets の同期を確認

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
- Backend API 接続エラー: Issue 1 の Cloudflare / PlanetScale 切り分けを実施

---

## 7. 参考資料

- [STG 運用 Runbook](../infra/staging/runbook.md) - 現行インフラの障害初動・復旧方針
- [README.md](./README.md) - デプロイメント・運用ドキュメント（障害判定、コマンド集）
- [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md) - 自動デプロイ・手動トリガー
- [CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) - CRUD 操作テスト詳細
- [VERCEL-FRONTEND-STAGING-TEST.md](./VERCEL-FRONTEND-STAGING-TEST.md) - フロントエンド検証
- [STG-DEMO-DATA-LIFECYCLE.md](./STG-DEMO-DATA-LIFECYCLE.md) - データ分類とライフサイクル
