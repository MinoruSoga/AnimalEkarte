# TASK-023: reservation_type_group_service の slog 実装を reservation_type_service に統一

## 概要

`reservation_type_group_service.go` の slog 実装が `reservation_type_service.go` と3点で非対称になっている。同種ドメインの実装パターンとして統一が必要。

## 優先度

HIGH

## 影響ファイル

| ファイル | 行 | 問題 |
|---------|-----|------|
| `backend/internal/service/reservation_type_group_service.go` | L77 | slog が bare string キー（型チェック不可） |
| `backend/internal/service/reservation_type_group_service.go` | L105-112 | Update 後の `slog.InfoContext` が欠落 |

## 規約違反

`.claude/rules/go-language.md`:
> 構造化ログ `slog` はサービス層のみで使用し、重要なミューテーション操作には InfoContext を記録する。

## 問題 1: slog bare string キー（L77）

```go
// 現状: bare string（型チェック不可・sloglint で警告）
slog.InfoContext(ctx, "reservation_type_group created", "id", g.ID, "name", g.Name)

// 修正: typed attribute（reservation_type_service と統一）
slog.InfoContext(ctx, "reservation_type_group created",
    slog.Uint64("reservation_type_group_id", g.ID),
    slog.String("name", g.Name))
```

## 問題 2: Update 後のログ欠落（L105-112）

```go
// reservation_type_service.go:243（参照実装）
slog.InfoContext(ctx, "reservation_type updated",
    slog.Uint64("reservation_type_id", st.ID))

// reservation_type_group_service.go — Update 後にログがない
// 修正: repo.Update 成功後に追加
slog.InfoContext(ctx, "reservation_type_group updated",
    slog.Uint64("reservation_type_group_id", id),
    slog.Uint64("clinic_id", clinicID))
```

## 参照実装

`backend/internal/service/reservation_type_service.go:224, 243, 262` が正しいパターン。
