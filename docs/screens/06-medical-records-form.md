# カルテ入力/編集 仕様書

![カルテ入力画面](./images/06-medical-records-form.png)
![カルテ詳細画面](./images/06-medical-records-detail.png)

## 概要
- **画面の目的**: 9タブ構成の診療記録（電子カルテ）を作成・編集する
- **URLパターン**:
  - 新規作成: `/medical-records/new?petId=xxx`
  - 編集: `/medical-records/:id`
- **アクセス権限**: 認証済ユーザー全員（操作権限は `usePermission` で制御）

## 画面構成
- **ヘッダー**: タイトル、戻るボタン、保存ボタン（React 19 `SubmitButton`）
- **患者情報カード (`PatientInfoCard`)**: 
  - ペット基本情報、担当医（クリックで `StaffSelectionModal`）、来院種別（クリックで 4 種トグル）を表示。
  - 編集モードでは、飼主名クリックで `OwnerSearchModal` を開き、カルテの飼主を付け替え可能。
- **9タブ構成**:
  1. **問診**: 主訴（`chiefComplaint`）、主訴カテゴリ、治療方針。過去履歴の参照。
  2. **診察/治療プラン**: 身体検査所見、診断カテゴリ/病名（2つまで）、診断詳細。
  3. **治療**: 処置・薬剤・物販項目の入力。
  4. **予防接種**: ワクチン接種記録。
  5. **定期健診**: 健診結果入力。
  6. **検査**: 各種検査結果の入力。
  7. **画像**: 検査画像・写真のアップロード・管理。
  8. **見積書**: 概算見積の作成・保存。
  9. **会計(医師確認)**: 医師による最終的な会計内容の確認と確定。

## 主要機能
- **React 19 アクション**: `useActionState` を使用したフォーム送信。メインカルテ保存成功後に、タブ別のサブデータ（診察プラン、見積書等）を自動で追記保存する。
- **来院種別トグル**: `初診` → `再診` → `緊急` → `往診` の順でサイクリックに切り替え。
- **タブ状態の維持**: 一度マウントしたタブは `hidden` 属性で制御し、入力内容を保持したまま高速に切り替え可能。
- **印刷**: A4 縦サイズに最適化されたカルテ印刷用ビュー（`MedicalRecordPrintView`）。
- **離脱防止**: `NavigationBlocker`（`unstable_useBlocker`）により、未保存時の意図しない画面遷移を警告。
- **バイタル記録**: どのタブからでもアクセス可能なバイタル記録専用モーダル。

## API連携
| メソッド | エンドポイント | 用途 |
|---------|--------------|------|
| POST | `/api/v1/medical-records` | カルテ新規作成 |
| PATCH | `/api/v1/medical-records/:id` | カルテ更新 |
| POST | `/api/v1/medical-records/:id/billing-confirmation/confirm` | 医師による会計確認 |
| GET | `/api/v1/masters/chief-complaint-categories` | 主訴カテゴリ取得 |
| GET | `/api/v1/medical-records/:id/clinical-plan` | 診察プラン取得 |
| GET | `/api/v1/medical-records/:id/treatments` | 治療内容取得 |
| GET | `/api/v1/medical-records/pets/:petId/history` | 特定ペットの過去カルテ履歴取得 |
