# シフトパターン設定 仕様書 (Shift Templates)

## 概要
- **画面の目的**: スタッフの標準的な勤務時間（例：フルタイム、早番、遅番等）をテンプレートとして定義し、シフト管理の入力を大幅に効率化する。
- **URLパターン**: `/settings/shift-templates`
- **アクセス権限**: シフト管理権限が必要（`ResourceShifts`）

---

## 1. 画面構成

### 1.1 テンプレート一覧
- **項目**: テンプレート名、種別、時間（開始〜終了。休日/有休は「-」表示）、ステータス、操作。
- **並び替え**: ドラッグ操作により、シフト入力時のクイック選択肢の優先順位を管理可能。

### 1.2 詳細編集サイドパネル (`ShiftTemplateSidePanel`)
- **テンプレート名**: スタッフが選択する際の識別名称。未入力の場合は保存ボタンが無効化される。医院内で一意（`uk_shift_templates_clinic_name`）。重複は 409 `shift_template_name_conflict` + `params.name` で、トーストは「シフトテンプレート名『早番』は既に使用されています」のように実名を含む（診療項目マスタと同じ形。正本の対比は [master-treatment.md §2.3](./master-treatment.md)）。
- **ステータス**: 有効/無効の切り替え（`StatusPill`）。無効のテンプレートはシフト入力ダイアログの選択肢に表示されない。
- **シフト種別**: 
    - `全日` / `午前` / `午後` / `休日` / `有休` のカテゴリ分け（`SHIFT_TYPE_LABELS`、`frontend/src/features/shifts/types/index.ts`）。
    - `休日` / `有休` を選択すると時刻・休憩の入力欄は非表示になる（`isShiftTemplateTimeHidden`）。
- **標準時間設定**: 始業・終業時刻（`<input type="time">`。分刻みの制約なし、30分単位への固定はない）。勤務種別で未入力のまま保存するとエラートーストが表示される。
- **メモ**: 補足情報の自由入力。
- **休憩時間定義 (`BreakEditor`)**: 
    - 1 日の中で複数の休憩（例：昼休憩 60 分 ＋ 夕方小休憩 15 分）をテンプレート化。
    - LINE 予約枠の算出に直接影響を与えます。

---

## 2. 主要な臨床・運営機能

### 2.1 LINE 予約空き枠の基礎
ここで設定された休憩時間は、スタッフにシフトが割り当てられた際、LINE 予約アプリ上の「予約不可（埋まっている）」時間として自動的にマッピングされます。これにより、休憩時間中に予約が入る事故を物理的に防ぎます。

### 2.2 入力コストの削減
カレンダー画面 (`/shifts`) のシフト登録ダイアログ (`ShiftFormDialog`) では、「テンプレートから入力」プルダウンで種別・開始/終了時刻・休憩を一括反映でき、複雑な休憩時間の入力を数秒で完了できます。

---

## 3. 技術仕様

### 3.1 構成コンポーネント
- **`ShiftTemplateSettings`**: マスタ管理の基盤ページ。
- **`ShiftTemplateSidePanelFields`** の `BreakEditor`: 行追加・削除に対応した動的リスト形式の休憩時間入力部品。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/shift-templates` | 登録済みテンプレートの取得 | `shifts` | `view` |
| POST | `/api/v1/shift-templates` | 新規パターンの作成 | `shifts` | `create` |
| PATCH | `/api/v1/shift-templates/reorder` | 全体の表示順序の更新 | `shifts` | `edit` |
| GET | `/api/v1/shift-templates/:id` | テンプレート詳細の取得 | `shifts` | `view` |
| PATCH | `/api/v1/shift-templates/:id` | テンプレート情報の更新 | `shifts` | `edit` |
| DELETE | `/api/v1/shift-templates/:id` | テンプレートの削除 | `shifts` | `delete` |

---
