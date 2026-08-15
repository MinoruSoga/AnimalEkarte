# ステージング環境・スモークテスト手順書 (Smoke Test Guide)

> **目的**: デプロイ直後のCRUD導線(医院・権限・スタッフ)確認手順を定義する。
> **読者**: デプロイ担当。
> **タイミング**: デプロイ直後の手動確認時。

> **Animal Ekarte**: デプロイ直後の基本動作とデータ整合性の最終確認
> **最新更新**: 2026-07-10 | **目的**: 本番リリース前の最終デプロイ検証

---

> [!WARNING]
> **⚠️ 2026-06**: login/readonly/CRUD の自動 smoke は `STG_DEMO_*` secret 未設定で約1年間
> 機能していなかったため撤去された（`stg-smoke.yml` は現状 health 疎通のみ）。当時の自動化計画
> (`CRUD-SMOKE-AUTOMATION.md`) は歴史的記録として退役済み（git 履歴参照。旧 docs/archive/ は削除済み）。
> 本手順書の CRUD 項目は **手動 curl** 又は本表の項目を参照して実施する。
> CRUD の正しさ自体は backend unit/integration テスト + FE route-guard テストでカバー済み。

## 1. テストの目的
本テストは、デプロイ完了直後のステージング環境において、主要な 3 系統（医院・権限・スタッフ）の機能が、インフラ・DB・API の各層で正しく連携して動作することを確認するために実施します。

**合格基準**: 全 CRUD 操作が期待通りのステータスコードを返し、外部キー保護・権限保護が正常に機能すること。

---

## 2. 準備：API アクセスと認証

### 2.1 環境変数のセット
```bash
export API_V1=https://api.stg.noah-karte.com/api/v1
export HEALTH_ENDPOINT=https://api.stg.noah-karte.com/health
```

### 2.2 認証フロー

**重要**: Stone に保存した認証情報（password・token・cookie）を文書化しないこと。以下は認証フローの手順のみ記載します。

1. ステージング環境にブラウザでアクセス: https://stg.noah-karte.com
2. `system_admin` 権限を持つアカウントでログイン
3. ブラウザ DevTools (F12) → Network タブを開く
4. 任意の API リクエストをクリック → Request headers を確認
5. `Cookie: access_token=...` と `Cookie: refresh_token=...` をメモ（このセッション内のみ）
6. 以下のテストで `${TOKEN}` として参照

**代替**: curl に `-b "access_token=${TOKEN}; refresh_token=${REFRESH}"` で認証を含める

---

## 3. テスト項目と実行例

### A. 医院 (Clinics) 系統

#### A-1. 一覧取得（全医院、admin scope）
```bash
curl -X GET "${API_V1}/clinics" \
  -b "access_token=${TOKEN}" \
  -H "Accept: application/json"
```
**期待**: `200 OK` + clinic list（admin は全医院表示）

#### A-2. 一覧取得（権限制限チェック）
```bash
# 一般スタッフアカウントで実行
curl -X GET "${API_V1}/clinics?scope=all" \
  -b "access_token=${STAFF_TOKEN}" \
  -H "Accept: application/json"
```
**期待**: `403 Forbidden`（scope=all は admin のみ）

#### A-3. 医院編集・保存
```bash
CLINIC_ID="<existing_clinic_id>"
curl -X PATCH "${API_V1}/clinics/${CLINIC_ID}" \
  -b "access_token=${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "TEST-UPDATE-'$(date +%s)'",
    "phone_number": "090-1234-5678"
  }'
```
**期待**: `200 OK` + updated clinic object
**検証**: UI で自動更新されることを確認

#### A-4. マルチテナント検証
STG-001 実行結果: 各医院スタッフが自身の clinic_id スコープのデータのみ取得可能。他医院データは 403 で拒否される（正常動作）。

---

### B. 権限グループ (Permission Groups) 系統

#### B-1. 新規グループ作成
```bash
curl -X POST "${API_V1}/masters/permission-groups" \
  -b "access_token=${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "TEST-GROUP-'$(date +%s)'",
    "description": "Smoke test temporary group",
    "permissions": [
      {"resource": "medical_records", "action": "view"}
    ]
  }'
```
**期待**: `201 Created` + group ID（e.g., `12345`）

#### B-2. グループ削除（保護確認）
```bash
# system_admin 削除試行
curl -X DELETE "${API_V1}/masters/permission-groups/1" \
  -b "access_token=${TOKEN}"
```
**期待**: `409 Conflict`（system_admin グループは削除不可）

#### B-3. テスト用グループ削除（cleanup）
```bash
curl -X DELETE "${API_V1}/masters/permission-groups/${TEST_GROUP_ID}" \
  -b "access_token=${TOKEN}"
```
**期待**: `204 No Content`（テスト用グループは削除可能）

---

### C. スタッフ (Staffs) 系統

#### C-1. スタッフ新規登録
```bash
curl -X POST "${API_V1}/masters/staffs" \
  -b "access_token=${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "TEST-STAFF-'$(date +%s)'",
    "email": "smoke-test-'$(date +%s)'@example.com",
    "role": "veterinarian"
  }'
```
**期待**: `201 Created` + staff ID（e.g., `staff-123`）

#### C-2. ログイン検証
```bash
# 新規登録されたスタッフで curl リクエスト可能か確認
curl -X GET "${API_V1}/me" \
  -b "access_token=${NEW_STAFF_TOKEN}"
```
**期待**: `200 OK` + 登録されたスタッフ情報

#### C-3. 削除保護（FK 検証）
```bash
# カルテ添付データがあるスタッフを削除
curl -X DELETE "${API_V1}/masters/staffs/${ATTACHED_STAFF_ID}" \
  -b "access_token=${TOKEN}"
```
**期待**: `409 Conflict`（子レコード存在のため削除拒否）

STG-001 実行結果: DELETE カルテが存在するスタッフに対して `409 Conflict` を返す確認済み。

#### C-4. 削除可能スタッフの削除（cleanup）
```bash
curl -X DELETE "${API_V1}/masters/staffs/${TEST_STAFF_ID}" \
  -b "access_token=${TOKEN}"
```
**期待**: `204 No Content`

---

## 4. ステータスコード期待値表

| 操作 | 期待ステータス | 条件 |
|------|----------------|------|
| GET / PATCH (成功) | `200 OK` | 権限あり、レコード存在 |
| POST (成功) | `201 Created` | バリデーション通過 |
| DELETE (成功) | `204 No Content` | 子レコード無し |
| DELETE (失敗 FK) | `409 Conflict` | 子レコード存在（正常な保護） |
| 権限なし | `403 Forbidden` | scope が異なる、admin のみ機能 |
| バリデーション失敗 | `400 Bad Request` | 必須項目欠損、型不正 |
| システムエラー | `500+ Internal Server Error` | **即座にロールバック** |

詳細は [Delete / Soft Delete 設計パターン](../../architecture/delete-soft-delete-patterns.md) の FK 制約と soft delete の注意点を参照してください。

---

## 5. 監査ログ検証

全 CRUD 操作後、`audit_logs` テーブル（または monitoring ダッシュボード）に以下を確認：
- 操作者 ID (`staff_id`) が記録されているか
- アクション (`CREATE`, `UPDATE`, `DELETE`) が正確か
- `resource_type` が `clinic`, `permission_group`, `staff` か
- タイムスタンプが操作時刻と一致するか

```bash
# 例: Cloudflare Workers Logs を確認
cd backend
npx wrangler tail --config wrangler.jsonc --format pretty
```

---

## 6. テスト データの削除ポリシー

### 6.1 推奨: API 経由での削除
上記の curl 削除例に従い、API で段階的に削除します。理由：
- 監査ログが自動記録される
- 外部キー制約を API レイヤで検証できる
- 本番環境での手順と同一

### 6.2 例外: DB 直接削除
以下の場合のみ、team lead 承認を得て DB 直接削除可能：
- API が 5xx エラーで応答不可
- テストデータが大量に残存し、一括削除が必要
- FK 階層の削除順序が複雑で、API では実行困難

**実行前に**: 組織の DevOps / SRE チームに削除内容と理由を報告し、書面承認を取得してください。

### 6.3 Cleanup タイムライン
1. テスト完了直後: CLI での削除実行
2. 削除確認: GET で 404 or リスト外を確認（API 層）
3. 監査ログ確認: DELETE アクション記録を audit_logs テーブルで確認し、runtime errorはWorkers Logsで確認
4. レポート作成: 削除したレコード数・操作者・タイムスタンプをログに記録

---

## 7. 異常発見時のアクション

不具合が発見された場合の判定フロー：

| 症状 | ステータスコード | アクション |
|------|-------------------|-----------|
| 权限エラー（期待外） | 403 / 401 | 認証情報確認、キャッシュクリア |
| FK 保護失敗（409 でなく 400/500） | 400 / 500 | **即座にロールバック** |
| 404（レコード存在が期待） | 404 | 入力データ確認、test データ状態確認 |
| 5xx エラー | 500+ | **即座に復旧判断**、Workers Logsを確認してlast-known-good再デプロイまたは基盤復旧へ進む |
| 監査ログ未記録 | (実行済も記録なし) | DB トランザクション確認、ロールバック判断 |

詳細は `docs/ops/deploy/README.md` §4 のロールバック判定基準を参照してください。

---

## 参考資料

- **デプロイメント運用**: `docs/ops/deploy/README.md`
- **CI/CD パイプライン**: `docs/ops/deploy/CI-CD-PIPELINE.md`
- **スタッフ研修**: チーム内 wiki の「STG 運用」セクション
