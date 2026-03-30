# BUG-032: 権限グループ名の重複作成が可能（UNIQUE制約なし）

## 概要
権限グループ作成時に既存のグループ名と同じ名前で新規作成が成功してしまう。
UNIQUE制約がなく、同名グループが複数作成できる。

## 再現手順
1. `POST /api/v1/permission-groups` に `{"name":"管理者"}` を送信
2. → HTTP 201 で重複作成成功

## 期待する動作
- 同名グループが既存の場合: 409 Conflict を返す
- エラーメッセージ: `"permission group name already exists"`

## 実装場所
- `backend/internal/service/permission_group_service.go` の Create メソッド
- または DB マイグレーションで `permission_groups.name` に UNIQUE制約を追加

## 優先度
Medium

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-032
- テスト確認日: 2026-03-30
