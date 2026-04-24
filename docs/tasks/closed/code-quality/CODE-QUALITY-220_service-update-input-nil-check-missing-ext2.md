# CODE-QUALITY-220: service Update input nil チェック欠落 追加対象（CODE-QUALITY-209/217 拡張）

## 概要

CODE-QUALITY-209（9サービス）・CODE-QUALITY-217（3サービス）に続く追加発見。
さらに7サービスの `Update` メソッドに `input == nil` チェックが欠落している。

## 該当サービス

| ファイル | 行番号 | 状態 |
|---------|--------|------|
| `backend/internal/service/reservation_type_service.go` | ~266 | `input == nil` チェックなし |
| `backend/internal/service/exam_type_service.go` | ~120 | `input == nil` チェックなし |
| `backend/internal/service/insurance_service.go` | ~125 | `input == nil` チェックなし |
| `backend/internal/service/occupation_service.go` | ~109 | `input == nil` チェックなし |
| `backend/internal/service/checkup_type_service.go` | ~136 | `input == nil` チェックなし |
| `backend/internal/service/diagnosis_service.go` | ~124 (DiagnosisType) | `input == nil` チェックなし |
| `backend/internal/service/diagnosis_service.go` | ~273 (DiagnosisName) | `input == nil` チェックなし |

## 修正パターン

各サービスの `Update` メソッド先頭（`FindByID` 呼び出し前）に追加:

```go
func (s *xxxService) Update(ctx context.Context, clinicID, id uint64, input *UpdateXxxInput) (*model.Xxx, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    // ... 以降の既存処理
}
```

`ErrMsgInputNotNil` は `backend/internal/service/validators.go` に定義済み。

## 関連チケット

- CODE-QUALITY-209 — 同問題の初回起票（9サービス）
- CODE-QUALITY-217 — 2回目（3サービス: inquiry_template, trimming_option, reservation_type_group）
- 本チケット — 3回目（7サービス）

**合計 19サービス** の Update に同問題が存在する。
全対応完了後に 209/217/220 をまとめてクローズすること。

## 優先度

MEDIUM — handler の ShouldBindJSON が成功した場合は nil にならないため実害リスクは低いが、
防御的プログラミングとして全サービスで統一すべき。
