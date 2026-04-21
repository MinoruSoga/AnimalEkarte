---
title: Service P17 Input Struct Naming 不統一
issue: '#477'
priority: LOW
status: open
area: service
pattern: P17
---

## 概要

サービスレイヤーの入力構造体（Input struct）の命名が標準パターン `CreateXxxInput` / `UpdateXxxInput` に従っていないケースが 4 件検出されました。

### パターン
- **P17 違反**：Input struct 名が標準形式（`CreateXxxInput` / `UpdateXxxInput`）でない

### 違反ファイル一覧

| ファイル | 行番号 | 現在の命名 | 標準命名 | 備考 |
|---------|--------|----------|---------|------|
| staff_service.go | 18 | `CreateStaffWithAccountInput` | `CreateStaffInput` | "WithAccount" は詳細実装、入力名では不要 |
| shift_break_template_service.go | 24 | `ShiftBreakTemplateInput` | `CreateShiftBreakTemplateInput` | 作成用入力なので接頭辞 Create が必須 |
| unavailable_time_service.go | 31 | `CreateUnavailableTimeInput` | 標準化 | 既に正しい（参考用） |
| day_schedule_service.go | 42 | `DaySchedule` | `CreateDayScheduleInput` または `UpdateDayScheduleInput` | モデルと入力構造体が混在 |

## 修正方法

標準形式 `CreateXxxInput` / `UpdateXxxInput` への統一：

```go
// staff_service.go
// 修正前
type CreateStaffWithAccountInput struct {
    Name     string
    Email    string
    Password string
    // ...
}

// 修正後
type CreateStaffInput struct {
    Name     string
    Email    string
    Password string
    // ...
}

// shift_break_template_service.go
// 修正前
type ShiftBreakTemplateInput struct {
    Name     string
    Duration int
}

// 修正後
type CreateShiftBreakTemplateInput struct {
    Name     string
    Duration int
}

// day_schedule_service.go
// 修正前
type DaySchedule struct {
    Date string
    // ...
}

// 修正後
type CreateDayScheduleInput struct {  // または UpdateDayScheduleInput
    Date string
    // ...
}
```

### 命名ルール
- **Create 操作**：`CreateXxxInput`
- **Update 操作**：`UpdateXxxInput`
- **内部詳細（WithAccount, WithPassword 等）**：入力struct名には含めない（メソッド実装で吸収）

## テスト

修正後、以下の確認を実施：
- [ ] サービスメソッドシグネチャが `CreateXxxInput` / `UpdateXxxInput` を受け取ること
- [ ] 呼び出し側ハンドラーのコードが struct 名変更に追従していること
- [ ] 既存テストが全件パス

## 参考

- Pattern: P17 (Standard Input Struct Naming)
- 標準パターン: `CreateXxxInput` / `UpdateXxxInput`
- 関連: Handler → Service インターフェイス定義
