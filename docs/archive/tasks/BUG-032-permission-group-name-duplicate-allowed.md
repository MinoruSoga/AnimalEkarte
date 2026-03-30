# BUG-032: 権限グループ名の重複登録が許可される（UNIQUE制約なし）

## 種類
バグ（バックエンド/DB — UNIQUE制約未実装）

## 重要度
中

## 発見日
2026-03-28

## 再現手順
1. 既存の「管理者」グループが存在する状態で
2. `POST /api/v1/permission-groups` に同名グループを作成する
   ```json
   { "name": "管理者", "description": "重複テスト", "color": "#ff0000" }
   ```

## 期待動作
- HTTP 409 Conflict が返る
- エラーメッセージ: 「このグループ名は既に使用されています」等

## 実際の動作
- HTTP 201 Created で重複グループが作成される
- permission_groups テーブルに同名のレコードが複数存在可能

## 影響
- 同名グループが複数存在することでユーザーが混乱
- グループ管理 UI で重複名が一覧に表示される
- ユーザーへの権限割当時に同名グループが複数表示され、誤操作の原因となる

## 修正方針
### DB
- `permission_groups` テーブルに `(clinic_id, name)` の UNIQUE 制約を追加
  ```sql
  ALTER TABLE permission_groups ADD CONSTRAINT uk_permission_groups_clinic_name
  UNIQUE (clinic_id, name) WHERE deleted_at IS NULL;
  ```

### バックエンド
- service 層の Create 処理で重複チェックを追加、または DB の UNIQUE 制約エラーをキャッチして `ErrConflict` に変換

## 対象ファイル（推定）
- `backend/migrations/001_init.sql`（UNIQUE制約追加）
- `backend/internal/service/permission_group_service.go`

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-032（BE） | Backend | `permission_groups.name` に UNIQUE 制約追加・重複時 ErrConflict → 409 を返す |
