# BUG-385: マスタサービスの WrapInvalidInput メッセージが英語（WrapConflict は日本語で矛盾）

## 概要
マスタ関連サービス（15ファイル）で `WrapConflict` のメッセージは全て日本語で統一されているが、`WrapInvalidInput` のメッセージは大多数が英語で書かれている。これらのエラーメッセージは API レスポンス経由でフロントエンド UI に表示されるため、ユーザーに英語メッセージが露出する問題がある。

## 再現手順
1. `PUT /v1/masters/cages/{id}` に空のボディ `{}` を送信
2. **結果**: `{"message": "at least one field must be provided"}` が表示（英語）
3. 比較: `DELETE /v1/masters/cages/{id}` で使用中のケージを削除しようとすると `{"message": "このケージは入院データで使用中のため削除できません"}` が表示（日本語）

## 期待する動作
- ユーザーに返す全エラーメッセージ（`WrapInvalidInput`・`WrapConflict`）を日本語に統一すること

## 現状コード

### `WrapConflict`（全件日本語 — 正しい実装）
```go
// 例: backend/internal/service/cage_service.go
return apperrors.WrapConflict("このケージは入院データで使用中のため削除できません")
```

### `WrapInvalidInput`（大多数が英語 — 問題箇所）
```go
// backend/internal/service/animal_species_service.go:95,123
return nil, apperrors.WrapInvalidInput("at least one field must be provided")
return apperrors.WrapInvalidInput("ids must not be empty")

// backend/internal/service/cage_service.go:102,130
return nil, apperrors.WrapInvalidInput("at least one field must be provided")
return apperrors.WrapInvalidInput("ids must not be empty")

// backend/internal/service/checkup_type_service.go
return nil, apperrors.WrapInvalidInput("at least one field must be provided")
return apperrors.WrapInvalidInput("ids must not be empty")

// 同様のパターンが全12マスタサービスで繰り返し（計32件）
```

### 例外（日本語の WrapInvalidInput — 参照実装）
```go
// backend/internal/service/merchandise_item_service.go:119,158
return nil, apperrors.WrapInvalidInput("金額は0以上を入力してください")
// backend/internal/service/diagnosis_service.go:242
return nil, apperrors.WrapInvalidInput("診断カテゴリが見つかりません")
```

## 影響範囲

| 対象サービス | 問題のある英語メッセージ | 件数 |
|------------|----------------------|------|
| animal_species_service.go | "at least one field must be provided" / "ids must not be empty" | 2 |
| cage_service.go | 同上 | 2 |
| checkup_type_service.go | 同上 | 2 |
| chief_complaint_service.go | 同上 + "input must not be nil" | 2 |
| diagnosis_service.go | 同上 | 2 |
| exam_type_service.go | 同上 | 2 |
| insurance_service.go | 同上 | 2 |
| medicine_service.go | "ids must not be empty" | 1 |
| occupation_service.go | 同上 2種 | 2 |
| procedure_service.go | 同上 + "input must not be nil" | 3 |
| reservation_type_service.go | "ids must not be empty" + "day_of_week is required..." 等 | 6 |
| reservation_type_group_service.go | 同上 2種 | 2 |
| trimming_master_service.go | 同上 2種 × 2リソース | 4 |
| vaccine_service.go | 同上 2種 | 2 |
| **合計** | | **約 34 件** |

## 修正方針

### 統一する日本語メッセージ

| 英語（現在） | 日本語（修正後） |
|------------|----------------|
| `"at least one field must be provided"` | `"少なくとも1つのフィールドを指定してください"` |
| `"ids must not be empty"` | `"並び順のIDリストが空です"` |
| `"input must not be nil"` | `"更新内容が指定されていません"` |
| `"day_of_week is required"` | `"曜日の指定は必須です"` |
| `"invalid unavailable_type: %s"` | `"不正な予約不可時間種別です: %s"` |
| `"start_time must be before end_time"` | `"開始時刻は終了時刻より前に指定してください"` |

### 修正例（cage_service.go:102,130）
```go
// 修正前
return nil, apperrors.WrapInvalidInput("at least one field must be provided")
return apperrors.WrapInvalidInput("ids must not be empty")

// 修正後
return nil, apperrors.WrapInvalidInput("少なくとも1つのフィールドを指定してください")
return apperrors.WrapInvalidInput("並び順のIDリストが空です")
```

### 一括修正の推奨アプローチ
`at least one field must be provided` と `ids must not be empty` は全15サービスで同一文言を使用しているため、`validators.go` の共通定数として定義する:

```go
// backend/internal/service/validators.go に追加
const (
    errMsgAtLeastOneField = "少なくとも1つのフィールドを指定してください"
    errMsgIDsNotEmpty     = "並び順のIDリストが空です"
    errMsgInputNotNil     = "更新内容が指定されていません"
)
```

各サービスの定数参照に変更することで、将来のメッセージ変更も一箇所で済む。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md`
> このシステムは日本の動物病院向けシステム。ユーザーに表示されるエラーメッセージは日本語で統一する。

### プロジェクト内参照実装
`backend/internal/service/merchandise_item_service.go:119` — 日本語 WrapInvalidInput の正しい例

## 優先度
**Medium** — ユーザーに英語エラーメッセージが表示される UX 問題。セキュリティ上の問題はないが、日本向け医療システムとして要件不足。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/service/validators.go` — 共通定数定義場所
- `backend/internal/service/animal_species_service.go:95,123` — 修正対象（代表例）
- 上記影響範囲テーブルの全ファイル
