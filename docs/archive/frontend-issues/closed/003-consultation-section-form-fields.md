# 003: ConsultationSection フォームフィールド追加

**ステータス:** open
**優先度:** medium
**関連API:** `PATCH /v1/masters/consultations/:id`

## 背景

バックエンドのハンドラ・DBには `time_condition` と `duration` が既に実装済み。
マスタ設定画面（`/settings/treatment-items` > 診察タブ）の `ConsultationSection` コンポーネントにUI入力フィールドが未実装。

## 追加するフィールド

| フィールド | UIコンポーネント | 値 | デフォルト |
|---|---|---|---|
| `time_condition` | `Select` | anytime/first_visit/revisit/after_hours/emergency | anytime |
| `duration` | `Input`（type=number）| 分単位の整数 | - |

## 表示ラベル（仕様書準拠）

| フィールド | ラベル | 選択肢 |
|---|---|---|
| `time_condition` | 適用区分 | 常時 / 初診 / 再診 / 時間外 / 緊急 |
| `duration` | 標準診察時間 (分) | placeholder: 例: 15 |

## 実装場所

- `MasterItemFormSections` のディスパッチャーに `consultation` ケースを追加
- `ConsultationSection` コンポーネントを実装（`SectionWrapper` / `NotionPropertyRow` 使用）

## 型定義

`models.ts` に `Consultation.time_condition` / `Consultation.duration` は既に定義済み。
`transformGenericMasterItem` が汎用変換のため、固有フィールドは別途取得が必要な場合あり。
