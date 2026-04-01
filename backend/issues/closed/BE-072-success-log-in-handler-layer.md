# BE-072: 成功時ログが handler 層に記述されており規約違反

**Status**: Closed
**Priority**: Low
**Affects**: `backend/internal/handler/`（複数ファイル）
**Date Created**: 2026-03-26
**Related**: -

## Summary

プロジェクト規約「slog は service 層のみに記述する（handler・repository には書かない）」に対して、複数の handler ファイルで成功時 slog ログが記述されている。一部の handler（`owner_handler.go` など）は規約通り service 側にログがあるが、後から追加された handler では handler 側に書いてしまっている。一貫性を保つために handler 側の slog を削除し、対応する service 層に移動する。

## 違反箇所一覧

| ファイル | 行 | 内容 |
|---|---|---|
| `accounting_handler.go` | :126-128 | CreateAccounting 成功ログ |
| `accounting_handler.go` | :176-178 | UpdateAccounting 成功ログ |
| `reservation_handler.go` | :144-146 | CreateReservation 成功ログ |
| `reservation_handler.go` | :234-236 | UpdateReservation 成功ログ |
| `hospitalization_handler.go` | :139-141 | CreateHospitalization 成功ログ |
| `hospitalization_handler.go` | :212-214 | UpdateHospitalization 成功ログ |
| `estimate_handler.go` | :118-120 | CreateEstimate 成功ログ |
| `estimate_handler.go` | :166-168 | UpdateEstimate 成功ログ |
| `medical_record_handler.go` | :211-213 | CreateMedicalRecord 成功ログ |
| `medical_record_handler.go` | :292-294 | UpdateMedicalRecord 成功ログ |
| `refund_handler.go` | :64-66 | CreateRefund 成功ログ（既に service 側にもあり重複） |

※ `auth_handler.go` のログは認証エラーという handler 固有の責務のため対象外。
※ `response.go` の `RespondError` 内の slog は ResponseHelper として許容。

## 現状のコード（代表例）

```go
// backend/internal/handler/accounting_handler.go:126-128（削除対象）
slog.InfoContext(ctx, "billing created",
    slog.Uint64("billing_id", billing.ID),
    slog.Uint64("clinic_id", clinicID))
```

## 修正方針

### handler 側: slog 呼び出しを削除

```go
// 変更前
result, err := h.service.Create(c.Request.Context(), input)
if err != nil {
    RespondError(c, err)
    return
}
slog.InfoContext(ctx, "billing created", ...)  // ← 削除
c.JSON(http.StatusCreated, toResponse(result))

// 変更後
result, err := h.service.Create(c.Request.Context(), input)
if err != nil {
    RespondError(c, err)
    return
}
c.JSON(http.StatusCreated, toResponse(result))
```

### service 側: 対応するログを追加（まだない場合）

```go
// backend/internal/service/accounting_service.go
func (s *accountingService) Create(ctx context.Context, input CreateAccountingInput) (*model.Billing, error) {
    // ...
    slog.InfoContext(ctx, "billing created",
        slog.Uint64("billing_id", billing.ID),
        slog.Uint64("clinic_id", input.ClinicID))
    return billing, nil
}
```

## 完了条件

- [ ] 上記 11 箇所の handler 側 slog を削除
- [ ] 対応する service 側に slog.InfoContext が存在することを確認（なければ追加）
- [ ] `docker compose exec backend golangci-lint run ./...` がパス（import 整理含む）
