# 004: ProcedureSection フォームフィールド追加

**ステータス:** open
**優先度:** medium
**関連API:** `PATCH /v1/masters/procedures/:id`

## 背景

バックエンドのハンドラ・DBには `duration` と `anesthesia` が既に実装済み。
マスタ設定画面（`/settings/treatment-items` > 処置タブ）の `ProcedureSection` コンポーネントにUI入力フィールドが未実装。

## 追加するフィールド

| フィールド | UIコンポーネント | 値 | デフォルト |
|---|---|---|---|
| `duration` | `Input`（type=number）| 分単位の整数 | - |
| `anesthesia` | `Select` | none/local/sedation/general | none |

## 表示ラベル（仕様書準拠）

| フィールド | ラベル | 選択肢 |
|---|---|---|
| `duration` | 所要時間 (分) | placeholder: 例: 30 |
| `anesthesia` | 麻酔要否 | 不要 / 局所麻酔 / 鎮静 / 全身麻酔 |

## 実装場所

- `MasterItemFormSections` のディスパッチャーに `procedure` ケースを追加
- `ProcedureSection` コンポーネントを実装（`SectionWrapper` / `NotionPropertyRow` 使用）

## 型定義

`models.ts` の `AnesthesiaType` には `sedation` が追加済み（make codegen 後）。
