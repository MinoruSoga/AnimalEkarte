# TASK-112: `reservation_type_service.go` — `check*` / `validate*` 関数命名不統一

## 優先度

**Low** — コードの一貫性・可読性の問題。機能には影響しない。

---

## 概要

`reservation_type_service.go` の予約不可時間バリデーション系ヘルパー関数で、
命名プレフィックスが `validate*` と `check*` で混在している。

プロジェクト内の他の Service ファイル（`cage_service.go`, `vaccine_service.go`, `procedure_service.go`）は
バリデーション系関数のプレフィックスを `validate*` で統一している。

---

## 問題箇所

### `service/reservation_type_service.go`（後半）

```go
// ✅ validate* プレフィックス（正しい）
func validateUnavailableTimeInput(input CreateUnavailableTimeInput) error { ... }

// ❌ check* プレフィックス（不統一）
func checkUnavailableTimeOverlap(
    existing []model.ReservationTypeUnavailableTime,
    input CreateUnavailableTimeInput,
) error { ... }
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ service/cage_service.go
func validateCageType(t string) error { ... }
func validateCageSize(s string) error { ... }

// ✅ service/vaccine_service.go
func validateVaccineSpecies(s string) error { ... }

// ✅ service/procedure_service.go
func validateAnesthesiaType(t string) error { ... }
```

---

## 修正方針

`checkUnavailableTimeOverlap` を `validateUnavailableTimeNotOverlaps` に改名する。
呼び出し箇所も同様に更新する。

```go
// ✅ 修正後
func validateUnavailableTimeNotOverlaps(
    existing []model.ReservationTypeUnavailableTime,
    input CreateUnavailableTimeInput,
) error { ... }
```

### 呼び出し箇所（同ファイル内）

```go
// ✅ CreateUnavailableTime メソッド内
if err := validateUnavailableTimeNotOverlaps(existing, input); err != nil {
    return nil, err
}
```

---

## 影響範囲

| ファイル | 対象 | 状態 |
|---------|------|------|
| `service/reservation_type_service.go` | `checkUnavailableTimeOverlap` 関数定義 | ❌ `check*` プレフィックス |
| `service/reservation_type_service.go` | `checkUnavailableTimeOverlap` 呼び出し箇所 | 改名に伴い更新 |

---

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md` — 命名規則

> 非エクスポート: camelCase（`validateInput`, `dbConn`）

`validate` プレフィックスは「入力値が不正な場合にエラーを返す関数」の一般的な Go 慣習。
`check` プレフィックスは bool 返却の確認関数に使われることが多く、error 返却関数には不適切。

### プロジェクト内参照実装

- `service/cage_service.go` — `validate*` 系ヘルパー関数群
- `service/closing_settings_service.go` — `validateSpecialPeriodTimes`
