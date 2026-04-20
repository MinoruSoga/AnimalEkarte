# CODE-QUALITY-202: Service Update/Delete での FindByID 存在確認パターン不統一

## 概要

複数のマスタ Service で `Update` / `Delete` メソッドの冒頭に存在確認 `FindByID` を置くかどうかが
ドメインごとに異なっている。統一パターンを確立して全マスタで揃える。

## 優先度

HIGH（consultation, insurance は他の標準パターンと逆の欠落）
MEDIUM（cage, occupation, animal_species, diagnosis は逆に冗長な二重確認）

## 影響ファイル

| ファイル | 問題 | 修正方向 |
|---------|-----|---------|
| `backend/internal/service/consultation_service.go` | Update で FindByID **欠落** | FindByID を追加 |
| `backend/internal/service/insurance_service.go` | Update で FindByID **欠落** | FindByID を追加 |
| `backend/internal/service/merchandise_item_service.go` | Update で FindByID **欠落** | FindByID を追加 |
| `backend/internal/service/animal_species_service.go` | Update で FindByID **冗長**（UpdateFields の RowsAffected で検出可） | 削除検討 |
| `backend/internal/service/cage_service.go` | Delete で FindByID **冗長** | 削除 |
| `backend/internal/service/occupation_service.go` | Delete で FindByID **冗長** | 削除 |
| `backend/internal/service/diagnosis_service.go` | Update(Type/Name) で FindByID **冗長** | 削除 |
| `backend/internal/service/consultation_service.go` | Delete で FindByID **冗長**（依存チェック後） | 削除 |

---

## 背景・方針決定

プロジェクト内に以下の2つのパターンが混在している:

### パターン A（明示的存在確認型）
```go
func (s *examTypeService) Update(...) {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get exam type")
    }
    // ... バリデーション → UpdateFields
}
```

### パターン B（RowsAffected 依存型）
```go
func (s *insuranceService) Update(...) {
    // FindByID なし
    fields := buildInsuranceUpdateFields(input)
    result, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
    // RowsAffected == 0 → WrapNotFound が repository で返される
}
```

**採用方針: パターン A を標準とする。**

理由:
1. 「存在確認」と「更新」が明確に分離されており読みやすい
2. 将来「更新前のデータを使った処理」（監査ログ等）を追加する際に素直に拡張できる
3. `exam_type`, `vaccine`, `medicine`, `procedure` 等の標準実装がパターン A を採用している

---

## 修正内容

### [HIGH] consultation_service.go — Update に FindByID 追加

```go
func (s *consultationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateConsultationInput) (*model.Consultation, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    // 追加: 存在確認
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get consultation")
    }
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    ...
}
```

### [HIGH] insurance_service.go — Update に FindByID 追加

```go
func (s *insuranceService) Update(...) (*model.Insurance, error) {
    // 追加: 存在確認
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get insurance")
    }
    if err := validateOptionalName(input.Name); err != nil { ...
```

### [HIGH] merchandise_item_service.go — Update に FindByID 追加

```go
func (s *merchandiseItemService) Update(...) (*model.MerchandiseItem, error) {
    // 追加: 存在確認
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get merchandise item")
    }
    if err := validateOptionalName(input.Name); err != nil { ...
```

### [MEDIUM] cage_service.go / occupation_service.go / diagnosis_service.go — Delete の冗長 FindByID を削除

これらのドメインは `Delete` 冒頭で `FindByID` による存在確認をしているが、
`CountUsageByXxxID` → `repo.Delete` (RowsAffected == 0 → WrapNotFound) の流れで
存在確認は不要。冗長なクエリを除去する。

```go
// 修正前（cage_service.go:158）
if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
    return apperrors.Wrap(err, "failed to get cage")
}
count, err := s.repo.CountRecordsByCageID(ctx, clinicID, id)

// 修正後
count, err := s.repo.CountRecordsByCageID(ctx, clinicID, id)
```

---

## 規約参照

- `.claude/CLAUDE.md`: handler → service → repository の責任分離

## テスト

- 各ドメインで存在しない ID を渡した場合に 404 が返ることを確認
- cage/occupation/diagnosis で Delete の冗長クエリ除去後も 404 が正しく返ることを確認
