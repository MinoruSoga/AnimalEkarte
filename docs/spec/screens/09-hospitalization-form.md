# 入院登録/編集 仕様書 (Hospitalization Form)

## 概要
- **画面の目的**: 入院またはペットホテルの新規受付、および登録時の治療プラン（診療内容・金額明細のスナップショット）の定義。
- **URLパターン**:
  - 新規作成: `/hospitalization/new?petId=xxx`
  - 編集: `/hospitalization/:id/edit`
  - ペット選択: `/hospitalization/select-pet`（create ガード）
- **アクセス権限**（`RequirePermission` + 画面内 `usePermission`）:
  - 親 `/hospitalization`: `ResourceHospitalization` **`view`**
  - 新規・ペット選択: **`create`**
  - 編集: **`edit`**
  - 保存は `canSubmit`（新規=create / 編集=edit）。削除は `delete`。**認証済ユーザー全員ではない**。
  - **親削除 UI ガード**: 紐付く治療プランが 1 件以上ある場合、削除ボタンを出さない（BE も Conflict で拒否）。API 失敗だけに頼らない。

---

## 用語（必須）

| 用語 | 意味 | 管理画面 |
|------|------|----------|
| **治療プラン** (treatment plan) | 入院**登録時**に確定する診療内容・金額明細の**スナップショット**。登録後は変更・削除不可。 | 本画面（新規入力 / 編集は参照のみ） |
| **ケアプラン** (care plan) | 入院中の投薬・給餌・処置など**継続的なケア項目**。 | 入院詳細画面 |

治療プランとケアプランを同一視しない。UI 文言でも「ケア／治療」と混同しない。

---

## 画面構成

### 1. 入院基本情報
- **患者ヘッダー**: `PatientInfoCard` + `formatPatientPetDetails` に `species` を渡す。動物種は実データを表示する（固定の「不明」にしない）。

Notionスタイルのプロパティ編集UIで、入院の根幹となる条件を設定します。
- **入院タイプ**: 「入院」または「ホテル」を選択。
- **予定期間**: 入院予定日と退院予定日の期間指定（新規 create body に含む。編集 UI に日付があっても update payload に start/end が無い場合は永続化されない点に注意）。
- **ケージ割り当て**: マスタに登録済みのケージ・個室から検索選択（空き状況によるフィルタリングはなし）。
- **担当医**: ヘッダー「担当医」ボタンで `staffType=doctor` かつ active のスタッフを選ぶ（`doctor_id`）。未選択可。一覧の担当医列と同じフィールド。
- **保険適用**: チェックONで保険会社名・保険番号の入力欄が表示される。
- **メモ**: 自由記述メモ。

### 2. 治療プランと連絡事項
- **治療プラン**（登録時スナップショット）: 治療内容・メモを行単位で入力する明細テーブル（保険・単価・数量・割引(%)・値引(円)・小計を表示）。
  - **新規作成時**: 治療内容が空でない行だけを、親 `POST /v1/hospitalizations` の `treatment_plans` に同梱して作成する。空行はスキップ。
  - **編集時**: `GET /v1/hospitalizations/:id/treatment-plans` で hydrate し **読み取り専用**。本画面の更新では治療プランを変更しない（親フィールドのみ）。
  - **BE**: 入院ネストの `PATCH` / `DELETE .../treatment-plans/:planId` はサービス層で **Conflict**（登録時スナップショットのため変更・削除不可）。ルートは fail-closed のまま残す。
- **主訴**: `owner_request` を入力するメモカード。一覧の主訴列と同じフィールド（新規列は無い）。
- **スタッフへの連絡事項**: `staff_notes`。
- **継続ケア**: 投薬・給餌などの**ケアプラン**は入院詳細画面で管理する（本画面の対象外）。

### 3. 金額概算
- **概算計**: 治療プラン明細に基づいた入院費の画面上自動計算（小計・消費税・請求額の読み取り専用表示）。
- **一括割引（% / 円）**: **UI に入力欄を置かない**。BE 永続化フィールドも無く、create/update でも送信しない。

---

## 主要な機能

- **保存アクション (`useActionState`)**: React 19 の最新パターンを採用。保存中のローディング表示や、バリデーションエラー時の対象フィールドへの自動フォーカスに対応。
- **意図しない離脱の防止**: 編集中のデータがある場合、ブラウザの「戻る」やページ遷移時に確認ダイアログを表示（`NavigationBlocker`）。
- **臨床安全ガード**: 死亡済みペットに対する新規登録を物理的にブロックします。
- **親削除ガード**: 治療プランが紐付いている入院は削除ボタン非表示 + 正直メッセージ。BE も `CountTreatmentPlansByHospitalizationID` で Conflict。

---

## 技術仕様

### 使用コンポーネント
- **`HospitalizationForm`**: メインフォームコンテナ。
- **`SearchableSelect`**（`HospitalizationBasicInfo` 内）: マスタ登録済みケージの検索選択。
- **`HospitalizationCostSummary`**: 金額概算の読み取り専用表示（一括割引入力なし）。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/hospitalizations/:id` | 編集時の既存レコード取得 | `hospitalization` | `view` |
| POST | `/api/v1/hospitalizations` | 入院レコードの新規保存（`treatment_plans` 同梱可。任意で `doctor_id` / `owner_request`） | `hospitalization` | `create` |
| PATCH | `/api/v1/hospitalizations/:id` | 登録情報の更新（治療プランは含まない。任意で `doctor_id` / `owner_request`） | `hospitalization` | `edit` |
| DELETE | `/api/v1/hospitalizations/:id` | 入院レコードの削除（子治療プランありは 409） | `hospitalization` | `delete` |
| GET | `/api/v1/hospitalizations/:id/treatment-plans` | 編集時の治療プラン hydrate | `hospitalization` | `view` |
| PATCH | `/api/v1/hospitalizations/:id/treatment-plans/:planId` | **常に Conflict**（スナップショット） | `hospitalization` | `edit` |
| DELETE | `/api/v1/hospitalizations/:id/treatment-plans/:planId` | **常に Conflict**（スナップショット） | `hospitalization` | `delete` |
| GET | `/api/v1/masters/cages` | 利用可能なケージ一覧の取得 | `master-hospitalization` | `view` |

---
