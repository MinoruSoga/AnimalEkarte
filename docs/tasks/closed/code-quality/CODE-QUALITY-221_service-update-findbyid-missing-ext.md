# CODE-QUALITY-221: service Update 前 FindByID 存在確認欠落 追加対象（CODE-QUALITY-202 拡張）

## 概要

CODE-QUALITY-202 で起票した「service Update/Delete FindByID 不整合」の追加対象。
さらに6サービスの `Update` メソッドが存在確認なしに `UpdateFields` を呼び出している。

## 該当サービスと問題パターン

| ファイル | 行番号 | FindByID | 問題 |
|---------|--------|----------|------|
| `backend/internal/service/cage_service.go` | ~133 | なし | 0件更新がサイレント成功 |
| `backend/internal/service/animal_species_service.go` | ~104 | なし | 同上 |
| `backend/internal/service/occupation_service.go` | ~109 | なし | 同上 |
| `backend/internal/service/diagnosis_service.go` | ~124 (DiagnosisType) | なし | 同上 |
| `backend/internal/service/diagnosis_service.go` | ~273 (DiagnosisName) | なし | 同上 |
| `backend/internal/service/inquiry_template_service.go` | ~116 | なし | 同上 |

## 正しい実装パターン（参照: insurance_service.go）

```go
func (s *insuranceService) Update(ctx context.Context, clinicID, id uint64, input *UpdateInsuranceInput) (*model.Insurance, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    // ✅ 存在確認
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get insurance")
    }
    // ... 以降の更新処理
}
```

## 問題の影響

FindByID なしに `UpdateFields` を呼ぶと:
1. 存在しない ID を指定した場合でも 0件更新がサイレント成功
2. GORM の UpdateFields は更新対象が0件でも error を返さない
3. handler は更新後のオブジェクトを返そうとして `FindByID` を呼ぶが、
   その時点で初めて 404 が判明する（更新なし成功のあとに別クエリで404）

→ API の動作が「更新成功 → 404」という矛盾したシーケンスになる可能性がある。

## 修正方針

各サービスの `Update` 先頭で:
```go
if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
    return nil, apperrors.Wrap(err, "failed to get xxx")
}
```

を `input == nil` チェックの直後に追加する。

## 関連チケット

- CODE-QUALITY-202 — 同問題の初回起票
- 本チケット完了後に CODE-QUALITY-202 と合わせてクローズすること

## 優先度

HIGH — 存在しない ID への Update が 404 を返さず成功する UX バグ。
