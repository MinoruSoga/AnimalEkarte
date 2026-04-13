# トリミングマスタ設定 仕様書

## 概要
- **画面の目的**: トリミング管理画面で使用される基本コースと追加オプションの名称、単価、所要時間を管理する。
- **URLパターン**: `/settings/trimming`
- **コンポーネント**: `[R] TrimmingSettings`

## 画面構成とタブ
Radix UI を用いた 2 つのタブで構成されます。
1. **コース** (`course`): シャンプーコース、カットコースなど。
2. **オプション** (`option`): 薬用シャンプー、歯磨きなど。

## 機能詳細

### 1. オプションの組合せ可否 (`CombinablePill`)
- オプションタブの一覧およびサイドパネルにおいて、「組合せ可否」を `CombinablePill` で表示します。
  - **可**: 背景 `C.bgStatusGreen` / 文字 `C.textStatusGreen`
  - **不可**: 背景 `C.bgInactive` / 文字 `C.text60`

### 2. 対象サイズの紐付け
- コース設定において、小型/中型/大型/特大などの対象サイズ（`TargetSize`）を選択可能。

## 表示・フォーム項目

### フォーム項目（コース）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| コース名 | `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `NotionStatusPill` | - | |
| 対象サイズ | `targetSize` | `Select` | - | 小型 / 中型 / 大型 / 特大 / 指定なし |
| 所要時間(分) | `duration` | `PropertyInput`| - | 数値入力 |
| 単価(税込) | `price` | `input(number)` | - | |
| 備考 | `description`| `PropertyInput` | - | |

### フォーム項目（オプション）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| オプション名| `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `NotionStatusPill` | - | |
| 所要時間(分) | `duration` | `PropertyInput`| - | 数値入力 |
| 組合せ可否 | `combinable`| `CombinablePill`| - | クリックでトグル |
| 単価(税込) | `price` | `input(number)` | - | |
| 備考 | `description`| `PropertyInput` | - | |


## API連携
| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/masters/trimming-courses` | コース一覧取得 | 実装済 |
| POST | `/api/v1/masters/trimming-courses` | コース作成 | 実装済 |
| PATCH | `/api/v1/masters/trimming-courses/:id` | コース更新 | 実装済 |
| DELETE | `/api/v1/masters/trimming-courses/:id` | コース削除 | 実装済 |
| PATCH | `/api/v1/masters/trimming-courses/reorder` | コース並び順一括保存 | 実装済 |
| GET | `/api/v1/masters/trimming-options` | オプション一覧取得 | 実装済 |
| POST | `/api/v1/masters/trimming-options` | オプション作成 | 実装済 |
| PATCH | `/api/v1/masters/trimming-options/:id` | オプション更新 | 実装済 |
| DELETE | `/api/v1/masters/trimming-options/:id` | オプション削除 | 実装済 |
| PATCH | `/api/v1/masters/trimming-options/reorder` | オプション並び順一括保存 | 実装済 |
