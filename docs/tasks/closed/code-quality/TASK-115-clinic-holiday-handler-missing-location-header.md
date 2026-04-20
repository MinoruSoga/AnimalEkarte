# TASK-115: `clinic_holiday_handler.go` — `SetClinicHoliday` の POST 201 に Location ヘッダーなし

## 優先度

**Low** — REST 規約の不統一。機能には影響しない。

---

## 概要

`clinic_holiday_handler.go` の `SetClinicHoliday` ハンドラは HTTP 201 Created を返しているが、
`Location` ヘッダーを設定していない。

プロジェクト内の他の POST 201 ハンドラは `Location` ヘッダーを設定しており、不統一になっている。

---

## 問題箇所

### `handler/clinic_holiday_handler.go:82`

```go
// ❌ Location ヘッダーなし
c.JSON(http.StatusCreated, toClinicHolidayResponse(holiday))
```

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ handler/cage_handler.go (CreateCage)
c.Header("Location", fmt.Sprintf("/v1/cages/%d", cage.ID))
c.JSON(http.StatusCreated, toCageResponse(cage))

// ✅ handler/animal_species_handler.go (CreateAnimalSpecies)
c.Header("Location", fmt.Sprintf("/v1/animal-species/%d", result.ID))
c.JSON(http.StatusCreated, toAnimalSpeciesResponse(result))
```

---

## 修正方針

### `handler/clinic_holiday_handler.go:82`

`SetClinicHoliday` の結果を識別する URL は `date`（YYYY-MM-DD）である。
`ID` フィールドも存在するが、DELETE エンドポイントが `/:date` で登録されているため、
Location パスも date ベースとする。

```go
// ✅ 修正後
c.Header("Location", fmt.Sprintf("/v1/clinic-holidays/%s", holiday.Date.Format("2006-01-02")))
c.JSON(http.StatusCreated, toClinicHolidayResponse(holiday))
```

---

## 影響範囲

| ファイル | 行 | 状態 |
|---------|---|------|
| `handler/clinic_holiday_handler.go:82` | SetClinicHoliday の 201 レスポンス | ❌ Location ヘッダーなし |

---

## 準拠すべきプロジェクト規約

### プロジェクト内参照実装

- `handler/cage_handler.go` — `c.Header("Location", ...)` の正しいパターン
- `handler/animal_species_handler.go` — 同上
