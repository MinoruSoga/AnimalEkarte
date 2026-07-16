# 予防接種入力/編集 仕様書 (Vaccination Form)

## 概要
- **画面の目的**: ワクチン接種等の詳細記録（ロット番号）の作成、および次回予定日の管理。
- **URLパターン**: 
  - 新規作成: `/vaccinations/new?petId=xxx`
  - 編集: `/vaccinations/:id`
- **アクセス権限**: 認証済ユーザー全員（操作権限は `usePermission` で制御）

---

## 1. 画面構成

### 1.1 接種基本情報
- **実施日**: デフォルトは当日。未来日は選択不可。
- **ワクチン**: マスタ（`vaccines`）から選択（混合、狂犬病等）。
- **補助説明**: 自由入力の補足テキスト。

### 1.2 製品トレーサビリティ管理 (LOT)
- **ロット番号**: 最大 4 つまでの LOT 番号を並行して登録可能（セット接種等に対応）。

### 1.3 期間管理（次回予定）
- **次回予定日**: 
    - **自動算出**: 標準間隔（3週後・4週後・1年後）を選択すると自動で日付を計算。
    - **手動調整**: 臨床的な判断に基づき、カレンダーから直接指定可能。

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
| GET | `/api/v1/vaccinations/:id` | 接種詳細情報の取得 | `vaccinations` | `view` |
| POST | `/api/v1/vaccinations` | 新規保存 | `vaccinations` | `create` |
| PATCH | `/api/v1/vaccinations/:id` | 登録内容（LOT、予定日等）の更新 | `vaccinations` | `edit` |
| GET | `/api/v1/masters/vaccines` | 利用可能なワクチン種類の取得 | `master-medical` | `view` |

---
