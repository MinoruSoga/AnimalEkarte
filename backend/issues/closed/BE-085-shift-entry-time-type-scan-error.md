# BE-085: shift_entries の start_time/end_time 型スキャンエラー（500）

## 現象

`GET /api/v1/shifts?date=2026-03` が 500 Internal Server Error を返す。

```
sql: Scan error on column index 5, name "start_time": unsupported Scan,
storing driver.Value type string into type *time.Time
```

## 原因

`shift_entries` テーブルの `start_time` / `end_time` カラムは PostgreSQL の `time` 型（例: `09:00:00`）。
Go モデル `ShiftEntry` がこれを `*time.Time`（完全なタイムスタンプ型）にスキャンしようとして失敗。

## 修正方針

Go モデルの `start_time` / `end_time` フィールドの型を `*time.Time` から `*string` または専用のカスタム型に変更する。

```go
// 修正前
StartTime *time.Time `gorm:"column:start_time"`
EndTime   *time.Time `gorm:"column:end_time"`

// 修正後（例）
StartTime *string `gorm:"column:start_time"`
EndTime   *string `gorm:"column:end_time"`
```

または PostgreSQL の `time` 型をスキャンできる型ライブラリを使用。

## 影響範囲

- `GET /api/v1/shifts` → 常に 500 エラー
- シフト管理画面が完全に使用不可
- BUG-029（編集ダイアログ時刻未表示）のテストも不可

## 優先度

高（シフト管理画面全体が動作不能）
