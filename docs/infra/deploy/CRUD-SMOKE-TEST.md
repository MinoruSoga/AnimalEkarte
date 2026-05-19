# STG CRUD Smoke Test

STG デプロイ直後に実行する CRUD 動作確認手順。
医院 (clinics) / 権限グループ (permission_groups) / スタッフ (staffs) の 3 系統。

## 前提

- STG にデプロイ済 + `/health` が 200
- 検証用 system_admin アカウントで login 済
- ベース URL: `https://api.stg.noah-karte.com/api/v1` (BE) / `https://stg.noah-karte.com` (FE)
- BE 認証ヘッダ: `Authorization: Bearer <jwt_token>` + `Cookie: refresh_token=...`

## 共通: ログイン

```bash
export API=https://api.stg.noah-karte.com/api/v1
TOKEN=$(curl -s -X POST "$API/login" \
  -H "Content-Type: application/json" \
  -d '{"login_id":"<admin_id>","password":"<password>"}' | jq -r '.access_token')
echo "$TOKEN" | cut -c1-30
```

---

## A. 医院 / クリニック CRUD

### API 確認

| 操作 | エンドポイント | 期待ステータス | 検証ポイント |
|---|---|---|---|
| 一覧 (所属のみ) | `GET /clinics` | 200 | `staff_clinic_assignments` 経由の所属クリニックのみ返る |
| 一覧 (全件) | `GET /clinics?scope=all` | 200 | `hospital-settings:view` 権限必須。403 にならないこと |
| 詳細 | `GET /clinics/:id` | 200 | 名称・住所・電話などのフィールド |
| 作成 | `POST /clinics` | 201 + `Location` ヘッダ | `Location: /api/v1/clinics/<new_id>` |
| 更新 | `PATCH /clinics/:id` | 200 | 部分更新 (PATCH セマンティクス) |
| 削除/無効化 | `DELETE /clinics/:id` | 204 or 409 | FK 参照あり → 409 (P10 規約) |

```bash
# 一覧
curl -s "$API/clinics?scope=all" -H "Authorization: Bearer $TOKEN" | jq '.[0:2]'

# 作成
NEW_ID=$(curl -s -X POST "$API/clinics" -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"smoke test clinic","address":"東京都","phone":"03-0000-0000"}' | jq '.id')
echo "$NEW_ID"

# 更新
curl -s -X PATCH "$API/clinics/$NEW_ID" -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"smoke test clinic updated"}' | jq '.name'

# 削除
curl -s -o /dev/null -w "%{http_code}\n" -X DELETE "$API/clinics/$NEW_ID" \
  -H "Authorization: Bearer $TOKEN"
```

### FE 画面確認

1. ヘッダ「医院マスタ」または `/settings/clinics` 配下に遷移
2. `ClinicMasterSettings.tsx` ルートで一覧表示
3. 新規追加 → モーダル/フォーム → 保存後にリスト更新
4. 編集 → 名称変更 → 保存後にリスト反映
5. 削除 → 確認ダイアログ → FK 参照あれば 409 エラー表示

### 確認ポイント (HIGH)

- system_admin で `scope=all` 取得時、全クリニックが返る (FEAT-374 Phase 1)
- 通常スタッフで `scope=all` リクエスト時、403 になる
- マルチテナント分離: 別クリニックのデータが混ざらない

---

## B. 権限グループ CRUD

### API 確認

| 操作 | エンドポイント | 期待ステータス | 検証ポイント |
|---|---|---|---|
| 一覧 | `GET /v1/masters/permission-groups` | 200 | `masters:view` 権限必須 |
| 詳細 | `GET /v1/masters/permission-groups/:id` | 200 | rules 配列 (resource × CRUD) を含む |
| 作成 | `POST /v1/masters/permission-groups` | 201 + Location | `name` ユニーク |
| 更新 | `PATCH /v1/masters/permission-groups/:id` | 200 | 名称・説明の部分更新 |
| ルール更新 | `PUT /v1/masters/permission-groups/:id/rules` | 200 | 一括置換セマンティクス |
| 削除 | `DELETE /v1/masters/permission-groups/:id` | 204 or 409 | スタッフ割当あり → 409 |

```bash
# 一覧
curl -s "$API/v1/masters/permission-groups" -H "Authorization: Bearer $TOKEN" | jq '.[0]'

# 作成
NEW_ID=$(curl -s -X POST "$API/v1/masters/permission-groups" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"name":"smoke_test_group","description":"smoke test"}' | jq '.id')

# ルール更新 (owners だけ view 許可)
curl -s -X PUT "$API/v1/masters/permission-groups/$NEW_ID/rules" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"rules":[{"resource":"owners","can_view":true,"can_create":false,"can_edit":false,"can_delete":false}]}'

# 削除
curl -s -o /dev/null -w "%{http_code}\n" -X DELETE "$API/v1/masters/permission-groups/$NEW_ID" \
  -H "Authorization: Bearer $TOKEN"
```

### FE 画面確認

1. `/settings/master/permission-groups` または `PermissionGroupSettings.tsx` 経由
2. 一覧表示 → 各グループ名 + ルール件数
3. 新規追加 → name 入力 → 保存
4. ルール編集 → resource × CRUD のチェックボックス → 保存
5. **重要**: 自分が所属するグループのルール変更直後、UI の認可表示が即時反映される (memory: clinic_context_architecture full reload pattern)

### 確認ポイント (HIGH)

- 標準ロール (system_admin / 院長 / 受付 / 獣医師 等) の seed が `003_seed_demo.sql` 経由で正しく投入されている
- system_admin グループは編集・削除不可 (バックエンドガード)
- ルール変更後、影響受けるスタッフが再ログインせずとも次回 API 呼び出しで 403/200 が反映 (middleware diff 検出, FEAT-374 Phase 2)

---

## C. スタッフ CRUD

### API 確認

| 操作 | エンドポイント | 期待ステータス | 検証ポイント |
|---|---|---|---|
| 一覧 | `GET /v1/masters/staffs` | 200 | clinic_id スコープで分離 |
| 詳細 | `GET /v1/masters/staffs/:id` | 200 | 職種・所属クリニック・権限グループ |
| 作成 | `POST /v1/masters/staffs` | 201 + Location | アカウント自動生成 (login_id + 初期パスワード) |
| 更新 | `PATCH /v1/masters/staffs/:id` | 200 | 部分更新 |
| 削除 | `DELETE /v1/masters/staffs/:id` | 204 or 409 | 予約・シフト・カルテ参照あり → 409 (BUG-030) |
| 権限グループ取得 | `GET /v1/masters/staffs/:id/permission-groups` | 200 | 複数グループ可 |
| 権限グループ設定 | `PUT /v1/masters/staffs/:id/permission-groups` | 200 | 一括置換 |
| 所属クリニック取得 | `GET /v1/masters/staffs/:id/clinics` | 200 | system_admin は全クリニック |
| 所属クリニック設定 | `PUT /v1/masters/staffs/:id/clinics` | 200 | 一括置換 |

```bash
# 一覧
curl -s "$API/v1/masters/staffs" -H "Authorization: Bearer $TOKEN" | jq '.[0]'

# 作成
NEW_ID=$(curl -s -X POST "$API/v1/masters/staffs" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"name":"smoke test staff","occupation_id":1,"login_id":"smoke_test_st","email":"smoke@test.local"}' | jq '.id')

# 権限グループ設定
curl -s -X PUT "$API/v1/masters/staffs/$NEW_ID/permission-groups" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"permission_group_ids":[<existing_group_id>]}'

# 所属クリニック設定
curl -s -X PUT "$API/v1/masters/staffs/$NEW_ID/clinics" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"clinic_ids":[1]}'

# 削除
curl -s -o /dev/null -w "%{http_code}\n" -X DELETE "$API/v1/masters/staffs/$NEW_ID" \
  -H "Authorization: Bearer $TOKEN"
```

### FE 画面確認

1. `/settings/master/staffs` または `StaffSettings.tsx`
2. 一覧 → 名称・職種・所属クリニック・権限グループ
3. 新規追加 → name / occupation / login_id / email → 保存後、初期パスワード表示モーダル
4. 編集 → 権限グループ多選択 → 所属クリニック多選択 → 保存
5. 削除 → 使用中なら 409 エラー (BUG-030 で検証済) → エラーメッセージ表示

### 確認ポイント (HIGH)

- スタッフ作成時にアカウント自動生成され、初回ログインで強制パスワード変更が要求される
- 削除時、予約・シフト・カルテに参照があれば 409 + 参照リソース名表示 (BUG-030)
- 職種マスタ・権限グループマスタ・クリニックマスタとの参照整合
- system_admin 削除は禁止 (バックエンドガード)

---

## 全体結果判定

| 区分 | 合格基準 |
|---|---|
| 機能 | 上記 18 API + 3 画面 すべて期待ステータス通り |
| 認可 | system_admin / 通常スタッフで 200/403 が正しく分岐 |
| データ整合 | 削除時の FK 参照チェック (409) が機能 |
| 監査ログ | `audit_log` に各操作が記録されている (AUDIT-H1/Q40) |
| エラー | `/health` 後 5 分間の CloudWatch ERROR が 0 |

NG が 1 件でもあれば DEPLOYMENT-CHECKLIST.md の **ロールバック基準** に従って `aws ecs update-service` で前 task definition revision に戻す。
