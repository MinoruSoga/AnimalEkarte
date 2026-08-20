# 予防接種入力/編集 仕様書 (Vaccination Form)

## 概要
- **画面の目的**: ワクチン接種等の詳細記録（ロット番号）の編集、および次回予定日の管理。
- **URLパターン**: 
  - 新規作成: カルテ「予防接種」タブから行う（`/vaccinations/new` は `/vaccinations/select-pet` へリダイレクト）
  - 編集: `/vaccinations/:id`
- **アクセス権限**: 親 `/vaccinations` は `ResourceVaccinations` **`view`**。`:id` は親 view 継承、保存/削除は `usePermission`

---

## 1. 画面構成

### 1.0 患者ヘッダー
- **`PatientInfoCard`**: `formatPatientPetDetails` に `species` / 生年月日 / 性別 / 去勢避妊を渡す。動物種は先頭に実データを出す。欠損の年齢・性別・去勢避妊だけ「不明」（固定ダミーの「犬」は使わない）。
- カルテ「予防接種」タブ（[06-medical-records-form.md](./06-medical-records-form.md)）は本画面とは別実装。タブ左ペインの一覧も同じ `GET /vaccinations?pet_id=` を使う。

### 1.1 接種基本情報
- **接種日**: 必須入力（未入力はエラー）。未来日は選択不可。
- **ワクチン**: 有効なワクチンマスタ（`useGetAllVaccinesMaster` の isActive）から選択（必須）。固定2択ハードコードではない。`MasterLink` でマスタへ遷移可。
- **補助説明**: 自由入力の補足テキスト。
- **備考**: 自由入力の備考テキスト。

### 1.2 製品トレーサビリティ管理 (LOT)
- **ロット番号**: 最大 4 つまでの LOT 番号を並行して登録可能（セット接種等に対応）。

### 1.3 期間管理（次回予定）
- **次回予定日**: 
    - **自動算出**: 標準間隔（3週後・4週後・1年後・以外（手動））を選択すると自動で日付を計算。ワクチン選択・接種日変更時も自動再計算。
    - **手動調整**: 臨床的な判断に基づき、カレンダーから直接指定可能。
    - **バリデーション**: 次回予定日は接種日より後であること。新規登録時は本日以降であること。

### 1.4 過去の接種履歴
- 選択中ペットの過去の接種記録を右カラムに一覧表示。期間・ワクチン名検索・ソート（昇順/降順）で絞り込み可能。

### 1.5 削除
- 編集時のみ、削除権限（`vaccinations` の `delete`）がある場合に削除ボタンを表示。`ConfirmDialog` による確認後に削除。

---

## 2. 主要な臨床・安全機能

### 2.1 未保存変更の保護
- **`NavigationBlocker`**: フォーム入力中の誤ったページ遷移を防ぎ、データの確実な保存を促します。

### 2.2 権限によるロック
- 編集権限 (`vaccinations` の `edit`/`create`) がない場合、フォーム全体が読み取り専用となります。

---

## 3. 技術仕様

### 3.1 構成コンポーネント
- **`VaccinationForm`**: 統合フォーム（React 19 Action 対応）。
- **`calculateNextDate`**: `use-vaccination-form` フック内の、接種周期に基づいた自動計算ロジック。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/vaccinations` | 接種履歴一覧の取得（履歴パネル用） | `vaccinations` | `view` |
| GET | `/api/v1/vaccinations/:id` | 接種詳細情報の取得 | `vaccinations` | `view` |
| POST | `/api/v1/vaccinations` | 新規保存 | `vaccinations` | `create` |
| PATCH | `/api/v1/vaccinations/:id` | 登録内容（LOT、予定日等）の更新 | `vaccinations` | `edit` |
| DELETE | `/api/v1/vaccinations/:id` | 接種記録の削除 | `vaccinations` | `delete` |

---
