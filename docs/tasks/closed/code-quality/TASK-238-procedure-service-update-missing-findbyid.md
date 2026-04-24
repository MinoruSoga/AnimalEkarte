# TASK-238: procedure_service.go — Update に FindByID 事前チェックが欠落（400 が 404 より先に返る）

## 優先度
High

## 対象ファイル
- `backend/internal/service/procedure_service.go`

## 問題概要
`Update` メソッドが入力バリデーション → `buildProcedureUpdateFields` → `UpdateFields` の順で実行するが、
`FindByID` による存在確認を行っていない。

存在しない `id` に対してリクエストを送ると、`UpdateFields` が内部で 404 を返す前に
入力バリデーション（400）が先に評価される可能性があり、HTTP セマンティクスが不統一になる。

規約: **Update は必ず FindByID で存在確認してから UpdateFields を呼ぶ（404 → 400 の順）。**

## 現状コード

```go
func (s *procedureService) Update(ctx context.Context, clinicID, id uint64, input *UpdateProcedureInput) (*model.Procedure, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    // ... バリデーション ...
    fields := buildProcedureUpdateFields(input)
    if len(fields) == 0 {
        return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
    }
    procedure, err := s.repo.UpdateFields(ctx, clinicID, id, fields)  // ❌ FindByID なし
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to update procedure")
    }
    // ...
}
```

## 比較（正しい実装例: cage_service.go）

```go
func (s *cageService) Update(ctx context.Context, clinicID, id uint64, input *UpdateCageInput) (*model.Cage, error) {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {  // ✅ FindByID 先行
        return nil, apperrors.Wrap(err, "failed to get cage")
    }
    // ... バリデーション・UpdateFields ...
}
```

## あるべき姿

```go
func (s *procedureService) Update(ctx context.Context, clinicID, id uint64, input *UpdateProcedureInput) (*model.Procedure, error) {
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {  // 追加
        return nil, apperrors.Wrap(err, "failed to get procedure")
    }
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    // ...
}
```

## 完了条件
- [ ] `Update` の先頭に `FindByID` による存在確認を追加
- [ ] `go test ./backend/internal/...` がパス
