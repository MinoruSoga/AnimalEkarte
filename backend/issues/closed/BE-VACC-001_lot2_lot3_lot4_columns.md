# BE-VACC-001: vaccinations テーブルに lot2〜lot4 カラムを追加

## 概要
予防接種フォームでは UI 上 LOT 番号を 4 つ入力できるが、
バックエンドは現在 `lot1` のみ保存している。
`lot2`, `lot3`, `lot4` の 3 カラムを追加し、全 LOT 番号を永続化できるようにする。

## 背景
フロントエンド (`VaccinationForm.tsx`) では 2×2 グリッドで LOT 番号 1〜4 を入力可能。
`use-vaccination-form.ts` は `lot2/lot3/lot4` の state を持っているが、
API リクエストには含まれていない（`lot1` のみ送信）。

## 実装内容

### migration
```sql
ALTER TABLE vaccinations
  ADD COLUMN lot2 VARCHAR(100) DEFAULT '',
  ADD COLUMN lot3 VARCHAR(100) DEFAULT '',
  ADD COLUMN lot4 VARCHAR(100) DEFAULT '';
```

### model (`backend/internal/model/vaccination.go`)
```go
Lot1 string `gorm:"column:lot1;type:varchar(100);default:''" json:"lot1"`
Lot2 string `gorm:"column:lot2;type:varchar(100);default:''" json:"lot2"`
Lot3 string `gorm:"column:lot3;type:varchar(100);default:''" json:"lot3"`
Lot4 string `gorm:"column:lot4;type:varchar(100);default:''" json:"lot4"`
```

### API
- `POST /api/v1/vaccinations` の request body に `lot2`, `lot3`, `lot4` を追加
- `PATCH /api/v1/vaccinations/:id` でも更新可能に

### codegen
モデル変更後に `make codegen` を実行し `frontend/src/types/generated/models.ts` を更新。

## 優先度
Medium

## 関連
- 仕様書: `docs/screens/15-vaccinations-form.md`（実装状況・制約テーブル）
- フロントエンド: `frontend/src/features/vaccinations/hooks/use-vaccination-form.ts`
