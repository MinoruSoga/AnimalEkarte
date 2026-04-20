# BUG-436: Update 前の FindByID 存在確認が欠落（systemic: 9 サービス）

## 概要

BUG-424/430/433/434 で個別に起票したものと同種の問題（Update 前の FindByID 欠落）が、
さらに 9 サービスで確認された。

本チケットは以下の未起票分を一括管理する。

## 問題サービス一覧

| ファイル | Update メソッド行 |
|---------|-----------------|
| cage_service.go | 86 |
| occupation_service.go | 85 |
| chief_complaint_service.go | 86 |
| checkup_type_service.go | 81 |
| exam_type_service.go | 75 |
| animal_species_service.go | 89 |
| consultation_service.go | 104 |
| merchandise_item_service.go | 152 |
| diagnosis_service.go（DiagnosisType） | 127 |

## 問題の例（cage_service.go:86-110）

```go
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    // ... バリデーション ...
    fields := buildCageUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
    }
    cage, err := s.repo.UpdateFields(ctx, clinicID, id, fields)  // ← FindByID なし
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to update cage")
    }
    // ...
}
```

Delete メソッドには FindByID が存在するにもかかわらず、Update には欠落している点が
全サービスで共通している。

## リスク

| 状況 | 現状の挙動 | 期待する挙動 |
|------|-----------|-------------|
| 存在しない ID を指定して PUT | UpdateFields が RowsAffected=0 でエラーなく終了する可能性 | 404 Not Found を返す |
| 別クリニックの ID を指定して PUT | clinicScope により無効（RowsAffected=0）になるが 404 でなく 200/エラーが返る可能性 | 404 Not Found を返す |

Repository の `UpdateFields` が RowsAffected=0 のとき NotFound を返せていれば実害は少ないが、
防御的プログラミングの観点から Service 層でも確認すべきである。

## 修正方針

全 Update メソッドの冒頭に `FindByID` を追加する。

```go
// ✅ 修正後パターン（全サービス共通）
func (s *xxxService) Update(ctx context.Context, clinicID, id uint64, input *UpdateXxxInput) (*model.Xxx, error) {
    // Step 1: 存在確認・テナント検証
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get xxx")
    }

    // Step 2: バリデーション
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    // ... 以降は変更なし
}
```

## 影響ファイル

- `backend/internal/service/cage_service.go` — 86 行
- `backend/internal/service/occupation_service.go` — 85 行
- `backend/internal/service/chief_complaint_service.go` — 86 行
- `backend/internal/service/checkup_type_service.go` — 81 行
- `backend/internal/service/exam_type_service.go` — 75 行
- `backend/internal/service/animal_species_service.go` — 89 行
- `backend/internal/service/consultation_service.go` — 104 行
- `backend/internal/service/merchandise_item_service.go` — 152 行
- `backend/internal/service/diagnosis_service.go` — 127 行（DiagnosisType Update）

## 優先度

**High** — クロステナント更新の防御が不完全。Repository の UpdateFields が RowsAffected=0 を
NotFound に変換できていない場合、存在しないリソースへの更新が無言で成功する可能性がある。

## 関連チケット

- BUG-424（reservation_type / trimming_master×2 — 同種問題の初回起票）
- BUG-430（reservation_type_group — 同種問題）
- BUG-433（vaccine — 同種問題）
- BUG-434（insurance — 同種問題）
