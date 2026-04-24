# BUG-428: エラーメッセージ定数が未使用・英語と日本語が多数サービスで混在

## 概要

`backend/internal/service/validators.go` に共通エラーメッセージ定数が定義されているにもかかわらず、
多くのサービスがこれを使わずハードコードしており、かつ英語と日本語が混在している。
BUG-419（consultation_service のみ）の拡張版。プロジェクト全体の問題。

## validators.go に定義された定数

```go
// backend/internal/service/validators.go
const (
    ErrMsgAtLeastOneField = "少なくとも1つのフィールドを指定してください"
    ErrMsgIDsNotEmpty     = "並び順のIDリストが空です"
    ErrMsgInputNotNil     = "更新内容が指定されていません"
)
```

## 問題 1: 定数を使わずハードコードしているサービス

### ErrMsgAtLeastOneField の代わりにハードコード

| サービス | 行番号 | ハードコード内容 |
|---------|--------|----------------|
| cage_service.go | 102 | `"少なくとも1つのフィールドを指定してください"` |
| checkup_type_service.go | 87 | 同上 |
| chief_complaint_service.go | 92 | 同上 |
| diagnosis_service.go | 127, 273 | 同上 |
| exam_type_service.go | 81 | 同上 |
| merchandise_item_service.go | 162 | 同上 |
| occupation_service.go | 91 | 同上 |
| reservation_type_group_service.go | 102 | 同上 |
| reservation_type_service.go | 272 | 同上 |
| trimming_master_service.go | 88, 244 | 同上 |

### ErrMsgIDsNotEmpty の代わりにハードコード

| サービス | 行番号 | ハードコード内容 |
|---------|--------|----------------|
| cage_service.go | 130 | `"並び順のIDリストが空です"` |
| checkup_type_service.go | 120 | 同上 |
| chief_complaint_service.go | 106 | 同上 |
| diagnosis_service.go | 158, 304 | 同上 |
| exam_type_service.go | 114 | 同上 |
| merchandise_item_service.go | 178 | 同上 |
| occupation_service.go | 122 | 同上 |
| reservation_type_group_service.go | 150 | 同上 |
| reservation_type_service.go | 299 | 同上 |
| trimming_master_service.go | 114, 270 | 同上 |

## 問題 2: 英語エラーメッセージが混在しているサービス

BUG-419（consultation_service）に加え、以下のサービスも英語を使用:

| サービス | 行番号 | 英語メッセージ |
|---------|--------|--------------|
| consultation_service.go | 105, 113, 139 | `"input must not be nil"`, `"at least one field must be provided"`, `"ids must not be empty"` |
| hospitalization_plan_service.go | 133 | `"ids must not be empty"` |
| inquiry_template_service.go | 121 | `"ids must not be empty"` |
| permission_group_service.go | 183 | （英語混在） |
| shift_template_service.go | 168 | `"ids must not be empty"` |
| staff_service.go | 393 | `"ids must not be empty"` |

## 修正方針

### Step 1: 全ハードコードを定数に置換

```go
// ❌ 修正前
return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")
return nil, apperrors.WrapInvalidInput("並び順のIDリストが空です")

// ✅ 修正後
return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
return nil, apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
```

### Step 2: 英語メッセージを日本語定数に統一

```go
// consultation_service.go 等
// ❌ 修正前
return nil, apperrors.WrapInvalidInput("input must not be nil")
return nil, apperrors.WrapInvalidInput("at least one field must be provided")
return nil, apperrors.WrapInvalidInput("ids must not be empty")

// ✅ 修正後
return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
return nil, apperrors.WrapInvalidInput(ErrMsgIDsNotEmpty)
```

### Step 3: 必要に応じて validators.go に不足定数を追加

```go
// 例: nil チェック用定数
const (
    ErrMsgAtLeastOneField = "少なくとも1つのフィールドを指定してください"
    ErrMsgIDsNotEmpty     = "並び順のIDリストが空です"
    ErrMsgInputNotNil     = "更新内容が指定されていません"
)
```

## 影響ファイル

**サービス層（ハードコード修正対象）**: 上表の全サービスファイル（約20ファイル・約40箇所）

## 優先度

**Medium** — ユーザー向けエラーメッセージが英語で表示される箇所あり。定数化により将来のメッセージ変更が一元管理可能になる。一括置換（sed/Grep）で効率的に対応可能。

## 関連チケット

- BUG-419（consultation_service のみの英語問題 — 本チケットの先行起票）
- `backend/internal/service/validators.go` — 定数定義ファイル
