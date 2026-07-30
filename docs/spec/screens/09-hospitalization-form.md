# 入院登録/編集 仕様書 (Hospitalization Form)

## 概要
- **画面の目的**: 入院またはペットホテルの新規受付、および治療プラン（診療内容・金額明細）の定義。一括割引は画面上の概算のみ。
- **URLパターン**:
  - 新規作成: `/hospitalization/new?petId=xxx`
  - 編集: `/hospitalization/:id/edit`
  - ペット選択: `/hospitalization/select-pet`（create ガード）
- **アクセス権限**（`RequirePermission` + 画面内 `usePermission`）:
  - 親 `/hospitalization`: `ResourceHospitalization` **`view`**
  - 新規・ペット選択: **`create`**
  - 編集: **`edit`**
  - 保存は `canSubmit`（新規=create / 編集=edit）。削除は `delete`。**認証済ユーザー全員ではない**。

---

## 画面構成

### 1. 入院基本情報
Notionスタイルのプロパティ編集UIで、入院の根幹となる条件を設定します。
- **入院タイプ**: 「入院」または「ホテル」を選択。
- **予定期間**: 入院予定日と退院予定日の期間指定（新規 create body に含む。編集 UI に日付があっても update payload に start/end が無い場合は永続化されない点に注意）。
- **ケージ割り当て**: マスタに登録済みのケージ・個室から検索選択（空き状況によるフィルタリングはなし）。
- **保険適用**: チェックONで保険会社名・保険番号の入力欄が表示される。
- **メモ**: 自由記述メモ。

### 2. 治療プランと連絡事項
入院中に実施する治療内容と申し送りを定義します（投薬・給餌などのケアプランは入院詳細画面で管理）。
- **治療プラン**: 治療内容・メモを行単位で入力する明細テーブル（保険・単価・数量・割引(%)・値引(円)・小計を表示）。
  - **新規作成時**: 治療内容が空でない行だけを、親 `POST /v1/hospitalizations` 成功後に `POST /v1/hospitalizations/:id/treatment-plans` で逐次作成する。空行はスキップ。親作成とネスト POST は単一 DB トランザクションではない。
  - **編集時**: `GET /v1/hospitalizations/:id/treatment-plans` で hydrate し **読み取り専用**。本画面の更新では治療プランを変更しない（親フィールドのみ）。
- **連絡事項**: 「飼主からのリクエスト」「スタッフへの連絡事項」の2枚のメモカード。

### 3. 金額と値引き設定
- **概算計**: 治療プランに基づいた入院費の画面上自動計算。
- **一括値引き（% / 円）**: **表示専用の概算**。BE に永続化フィールドは無く、create/update でも送信しない。作成/編集権限があっても保存されない。

---

## 主要な機能

- **保存アクション (`useActionState`)**: React 19 の最新パターンを採用。保存中のローディング表示や、バリデーションエラー時の対象フィールドへの自動フォーカスに対応。
- **意図しない離脱の防止**: 編集中のデータがある場合、ブラウザの「戻る」やページ遷移時に確認ダイアログを表示（`NavigationBlocker`）。
- **臨床安全ガード**: 死亡済みペットに対する新規登録を物理的にブロックします。

---

## 技術仕様

### 使用コンポーネント
- **`HospitalizationForm`**: メインフォームコンテナ。
- **`SearchableSelect`**（`HospitalizationBasicInfo` 内）: マスタ登録済みケージの検索選択。
- **`HospitalizationCostSummary`**: 金額計算・値引き処理モジュール。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/hospitalizations/:id` | 編集時の既存レコード取得 | `hospitalization` | `view` |
| POST | `/api/v1/hospitalizations` | 入院レコードの新規保存 | `hospitalization` | `create` |
| PATCH | `/api/v1/hospitalizations/:id` | 登録情報の更新 | `hospitalization` | `edit` |
| DELETE | `/api/v1/hospitalizations/:id` | 入院レコードの削除（編集画面の削除ボタン） | `hospitalization` | `delete` |
| GET | `/api/v1/hospitalizations/:id/treatment-plans` | 編集時の治療プラン hydrate | `hospitalization` | `view` |
| POST | `/api/v1/hospitalizations/:id/treatment-plans` | 新規作成後の治療プラン行作成（内容非空のみ） | `hospitalization` | `create` |
| GET | `/api/v1/masters/cages` | 利用可能なケージ一覧の取得 | `master-hospitalization` | `view` |

---
