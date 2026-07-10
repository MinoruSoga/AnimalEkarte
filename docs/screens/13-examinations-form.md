# 検査入力/結果登録 仕様書 (Examination Form)

## 概要
- **画面の目的**: 検査オーダーの作成、および臨床検査値（数値・判定）の精緻な記録。
- **URLパターン**: 
  - 新規作成: `/examinations/new?petId=xxx`
  - 編集: `/examinations/:id`
- **アクセス権限**: 認証済ユーザー全員（操作権限は `usePermission` で制御）

---

## 1. 画面構成

### 1.1 検査基本情報
- **検査種別**: 血液、尿、エコー等のマスタ（`exam_types`）から選択。
- **担当スタッフ**: 実施者または判定者。

### 1.2 動的検査項目テーブル (`ExamItemsTable`)
選択した検査種別に基づき、測定項目が動的に生成されます。

| 項目 | 説明 |
|:---|:---|
| **項目名 / 単位** | 例：ALT (U/L), CRE (mg/dL)。マスタより自動展開。 |
| **測定値** | 実測値を数値入力。 |
| **判定 (H/L)** | **自動判定**: マスタで定義された基準値（Min〜Max）に基づき、高値なら **赤**、低値なら **ブランドカラー（teal）** でハイライトされます。判定はバックエンドが導出するため、保存・再読込後に反映されます。 |
| **基準値参照** | その動物種（犬/猫）における正常範囲を横に表示し、臨床判断を助けます。 |

---

## 2. 主要な臨床・安全機能

### 2.1 真正性の担保（確定ロック）
- **ステータス**: `確定 (finalized)` 状態に移行した検査記録は、一切の編集がロックされます。
- **背景**: 臨床データとしての証跡保護のため、一度確定した数値を不用意に変更することを防ぎます。

### 2.2 保存時ハイライト
数値の判定（H/L）はバックエンドが基準値（Min〜Max）から導出し、保存・再読込のタイミングで UI の色に反映されます。これにより、多忙な診察室でも重要値の見落としを防止します。

---

## 3. 技術仕様

### 3.1 構成コンポーネント
- **`ExaminationForm`**: 統合フォーム。
- **`ExaminationFormFields`**: 検査種別・担当スタッフ等の基本情報フィールド。
- **`ExamItemsTable`**: 選択された検査種別の項目テンプレ（`exam_type_fields`、`useGetExamTypeFields` で取得）を描画する動的テーブル。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/examinations/:id` | 検査基本情報および項目リストの取得 | `examinations` | `view` |
| POST | `/api/v1/examinations` | 新規保存 | `examinations` | `create` |
| PATCH | `/api/v1/examinations/:id` | 数値・判定・ステータスの更新 | `examinations` | `edit` |
| GET | `/api/v1/examinations/:id/items` | 検査項目（結果）のリスト取得 | `examinations` | `view` |
| PUT | `/api/v1/examinations/:id/items` | 検査項目（結果）の一括更新 | `examinations` | `edit` |
| GET | `/api/v1/masters/examination-types` | 利用可能な検査種別リストの取得 | `master-medical` | `view` |

---

