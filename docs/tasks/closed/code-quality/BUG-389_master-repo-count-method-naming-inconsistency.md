# BUG-389: マスタリポジトリの CountUsage 系メソッド命名が不統一

## 概要
マスタリポジトリの CountUsage 系メソッドで以下の命名不統一がある。(1) `medicine_repository.CountChildren` が `CountChildrenByParentID` の省略形になっており、何の子リソースを数えているか不明。(2) `reservation_type_group_repository.CountCategories` が汎用的すぎて実装内容（予約種別のカウント）が名前から読み取れない。他のリポジトリは `Count[Resource]By[ParentField]` パターンで統一されている。

## 再現手順
コードレビューで確認可能。

## 期待する動作
- `medicine_repository.CountChildren` → `CountChildrenByParentID` に統一
- `reservation_type_group_repository.CountCategories` → `CountReservationTypesByGroupID` に変更

## 現状コード

### 問題1: `backend/internal/repository/medicine_repository.go:19`
```go
// インターフェース定義
CountChildren(ctx context.Context, clinicID, parentID uint64) (int64, error)
// ↑「何の Children を数えるのか」が名前から不明

// 実装:79行目
func (r *medicineRepository) CountChildren(ctx context.Context, clinicID, parentID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).
        Model(&model.Medicine{}).
        Where("parent_id = ? AND deleted_at IS NULL", parentID).
        Count(&count).Error
    return count, apperrors.FromGORM(err, "medicine", "")
}
```

### 問題2: `backend/internal/repository/reservation_type_group_repository.go:16`
```go
// インターフェース定義
CountCategories(ctx context.Context, clinicID, groupID uint64) (int64, error)
// ↑「何のカテゴリを数えるのか」が名前から不明

// 実装:48行目
func (r *reservationTypeGroupRepository) CountCategories(ctx context.Context, clinicID, groupID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.ReservationType{}).
        Where("reservation_type_group_id = ? AND clinic_id = ? AND deleted_at IS NULL", groupID, clinicID).
        Count(&count).Error
    return count, apperrors.FromGORM(err, "reservation_type_group", fmt.Sprintf("%d", groupID))
}
// → 実際には ReservationType レコードを数えている
```

### 比較: 正しい実装（プロジェクト内参照実装）
```go
// backend/internal/repository/checkup_type_repository.go
CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
// ↑ ParentID を条件にした子リソースカウント — 意図明確

// backend/internal/repository/exam_type_repository.go
CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
// ↑ 同パターン — 一貫性あり
```

## 影響範囲

| 対象 | 変更前 | 変更後 |
|------|-------|-------|
| `backend/internal/repository/medicine_repository.go:19,79` | `CountChildren(...)` | `CountChildrenByParentID(...)` |
| `backend/internal/service/medicine_service.go` | `repo.Medicine.CountChildren(...)` の呼び出し | `repo.Medicine.CountChildrenByParentID(...)` に変更 |
| `backend/internal/repository/reservation_type_group_repository.go:16,48` | `CountCategories(...)` | `CountReservationTypesByGroupID(...)` |
| `backend/internal/service/reservation_type_group_service.go` | `repo.ReservationTypeGroup.CountCategories(...)` の呼び出し | `CountReservationTypesByGroupID(...)` に変更 |

## 修正方針

### 1. `backend/internal/repository/medicine_repository.go:19,79` — リネーム
```go
// インターフェース変更（19行目）
CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)

// 実装関数のリネーム（79行目）
func (r *medicineRepository) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
    // 実装は変更なし
}
```

### 2. `backend/internal/service/medicine_service.go` — 呼び出し側を更新
```go
// 修正前
count, err := s.repo.Medicine.CountChildren(ctx, clinicID, id)

// 修正後
count, err := s.repo.Medicine.CountChildrenByParentID(ctx, clinicID, id)
```

### 3. `backend/internal/repository/reservation_type_group_repository.go:16,48` — リネーム
```go
// インターフェース変更（16行目）
CountReservationTypesByGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error)

// 実装関数のリネーム（48行目）
func (r *reservationTypeGroupRepository) CountReservationTypesByGroupID(ctx context.Context, clinicID, groupID uint64) (int64, error) {
    // 実装は変更なし
}
```

### 4. `backend/internal/service/reservation_type_group_service.go` — 呼び出し側を更新
```go
// 修正前
count, err := s.repo.ReservationTypeGroup.CountCategories(ctx, clinicID, id)

// 修正後
count, err := s.repo.ReservationTypeGroup.CountReservationTypesByGroupID(ctx, clinicID, id)
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/naming-conventions.md` — Go 層の命名
> メソッド名はその操作対象が明確に分かる名前を使用する。`Count[Child]By[ParentField]` パターンを採用する。

### プロジェクト内参照実装
`backend/internal/repository/checkup_type_repository.go` — `CountChildrenByParentID` の正しい命名

## 優先度
**Low** — 機能上の問題なし。命名の一貫性問題。テストを含めた修正が必要だが、影響範囲は限定的（2メソッド）。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/repository/medicine_repository.go:19,79` — 問題箇所
- `backend/internal/service/medicine_service.go` — 呼び出し側（更新が必要）
- `backend/internal/repository/reservation_type_group_repository.go:16,48` — 問題箇所
- `backend/internal/service/reservation_type_group_service.go` — 呼び出し側（更新が必要）
- `backend/internal/service/medicine_service_test.go` — テスト更新が必要
- `backend/internal/service/reservation_type_group_service_test.go` — テスト更新が必要
