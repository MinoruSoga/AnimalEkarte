# トリミングマスタ設定 仕様書

## 概要
- **画面の目的**: トリミング管理画面で使用される基本コースと追加オプションの名称、単価、所要時間を管理する。
- **URLパターン**: `/settings/trimming`
- **コンポーネント**: `[R] TrimmingSettings`

## 画面構成とタブ
2つのタブに分けて管理します。
1. **コース** (`course`): シャンプーコース、カットコースなど。
2. **オプション** (`option`): 薬用シャンプー、歯磨き、リボン付けなど。

## 機能詳細

### 1. オプションの組合せ可否（CombinablePill）
- オプションには「組合せ可否（`combinable`）」フラグが存在し、複数のオプションを同時適用できるか否かを管理します。
- 一覧およびフォーム上では、色付きのピル（バッジ）で視覚的に識別可能（可=緑、不可=グレー）。

### 2. 対象サイズの紐付け
- コース設定において、小型犬・中型犬などの対象サイズ（`TargetSize`）を紐付けることができ、予約・カルテ入力時のサジェストに利用されます。

## 表示・フォーム項目

### フォーム項目（コース）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| コース名 | `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `StatusToggleButton` | - | |
| 対象サイズ | `targetSize` | `Select` | - | 指定なし / 小型犬 / 中型犬... |
| 所要時間 | `duration` | `PropInput(number)`| - | 単位: 分 |
| 単価(税込) | `price` | `MoneyInput` | - | |
| 備考 | `description`| `PropInput` | - | |

### フォーム項目（オプション）
| フィールド | 項目ID | 入力部品 | 必須 | 備考 |
|-----------|--------|---------|------|------|
| オプション名| `name` | `Input` | ✅ | タイトルエリア |
| ステータス | `isActive` | `StatusToggleButton` | - | |
| 所要時間 | `duration` | `PropInput(number)`| - | 単位: 分 |
| 組合せ可否 | `combinable`| トグルボタン | - | |
| 単価(税込) | `price` | `MoneyInput` | - | |
| 備考 | `description`| `PropInput` | - | |

## API連携
| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/masters/trimming-courses` | コース一覧取得 | 実装済 |
| GET | `/api/v1/masters/trimming-options` | オプション一覧取得 | 実装済 |
| POST | `/api/v1/masters/trimming-courses` | コース作成 | 実装済 |
| PATCH | `/api/v1/masters/trimming-courses/reorder` | 並び順一括保存 | 実装済 |
