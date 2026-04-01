# BE: BUG-030 依存データがある状態での削除が成功（FK制約・依存チェックなし）

## 概要

以下のマスタデータを削除する際、依存データの存在チェックがなく削除が成功または 500 が返る。
孤立データ（orphaned records）が発生しデータ整合性が破壊される。

## 再現箇所（複数）

1. **サービス種別削除**（紐付き予約あり）: DELETE → 500 Internal Server Error（FK制約違反でクラッシュ）
2. **スタッフ削除**（担当予約あり）: DELETE 204 で削除成功。既存予約の doctor_id が孤立
3. **ケージ削除**（入院患者あり）: DELETE 204 で削除成功。入院データの cage_id が孤立
4. **権限グループ削除**（割当中ユーザーあり）: DELETE 204 で削除成功。割当ユーザーの権限が即時剥奪

## 期待する動作

- 依存データがある場合は 409 Conflict を返す
- エラーメッセージ: 「このデータは他のレコードに使用されています」

## 実装場所

- `backend/internal/service/` 各マスタサービス
  - `service_type_service.go`
  - `staff_service.go`
  - `cage_service.go`
  - `permission_group_service.go`
- 削除前に依存データ count チェックを追加 → count > 0 なら `apperrors.ErrConflict` を返す

## 優先度

High（データ整合性破壊）

## 関連

- `docs/tasks/open/security/BUG-030_delete_without_dependency_check.md`
- FUNCTIONAL_TEST_REPORT.md BUG-030
