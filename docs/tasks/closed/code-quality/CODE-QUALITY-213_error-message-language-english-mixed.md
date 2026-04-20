# CODE-QUALITY-213: エラーメッセージの英日混在（エンドユーザーに英語が露出）

## 概要

`validators.go`、`response.go`、`reservation_type_service.go` に英語エラーメッセージが残っており、
他の日本語メッセージと混在している。共通ヘルパーから返るメッセージは全エンドポイントに影響するため優先度が高い。

## 優先度

MEDIUM

## 影響ファイル

| ファイル | 問題箇所 | 内容 |
|---------|---------|-----|
| `backend/internal/service/validators.go` | L43,56,71,83,133,199,207,216,230,241,258 | ENUM バリデーションエラーが全て英語 |
| `backend/internal/handler/response.go` | L182 | `parseDateQuery` のエラーが英語 |
| `backend/internal/handler/response.go` | L261-263, L280-281 | `parsePagination` / `parseIDParam` が英語 |
| `backend/internal/service/reservation_type_service.go` | L436, L440 | `validateUnavailableTimeInput` 内の2メッセージが英語 |

---

## 問題詳細

### 1. validators.go — ENUM バリデーションメッセージ（全て英語）

```go
// 現状（英語）
return fmt.Errorf("invalid gender: %s", gender)
return fmt.Errorf("invalid cage type: %s", cageType)
return fmt.Errorf("invalid anesthesia type: %s", anesthesiaType)
```

これらは `apperrors.WrapInvalidInput(...)` 経由で HTTP 400 レスポンスとしてクライアントに返る。

**修正方針**:
```go
return apperrors.WrapInvalidInput(fmt.Sprintf("性別の値が不正です: %s", gender))
return apperrors.WrapInvalidInput(fmt.Sprintf("ケージ種別の値が不正です: %s", cageType))
return apperrors.WrapInvalidInput(fmt.Sprintf("麻酔種別の値が不正です: %s", anesthesiaType))
// 以下、各 ENUM バリデーション関数を同様に日本語化
```

---

### 2. response.go:182 — parseDateQuery のエラー

```go
// 現状（英語）
return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s must be YYYY-MM-DD format", key))

// 修正後
return nil, apperrors.WrapInvalidInput(fmt.Sprintf("%s は YYYY-MM-DD 形式で入力してください", key))
```

---

### 3. response.go:261-263, 280-281 — parsePagination / parseIDParam

```go
// 現状（英語）
return 0, 0, apperrors.WrapInvalidInput("page must be a positive integer")
return 0, 0, apperrors.WrapInvalidInput("limit must be between 1 and 100")
RespondError(c, apperrors.WrapInvalidInput("missing "+key))
RespondError(c, apperrors.WrapInvalidInput("invalid "+key))

// 修正後
return 0, 0, apperrors.WrapInvalidInput("page は1以上の整数で指定してください")
return 0, 0, apperrors.WrapInvalidInput("limit は1〜100の範囲で指定してください")
RespondError(c, apperrors.WrapInvalidInput("パラメータが不足しています"))
RespondError(c, apperrors.WrapInvalidInput("パラメータの形式が不正です"))
```

---

### 4. reservation_type_service.go:436, 440

```go
// 現状（英語）
return apperrors.WrapInvalidInput("day_of_week must be between 0 (Sun) and 6 (Sat)")
return apperrors.WrapInvalidInput("specific_date is required for specific type")

// 修正後
return apperrors.WrapInvalidInput("day_of_week は 0（日曜）から 6（土曜）の範囲で指定してください")
return apperrors.WrapInvalidInput("specific タイプでは specific_date の指定は必須です")
```

---

## 規約参照

- `.claude/CLAUDE.md`: エラーメッセージは日本語で統一

## テスト

- 各エラーケースで日本語メッセージが返ることを確認
- 既存のエラーメッセージチェックをしているテストがある場合は期待値を更新する
