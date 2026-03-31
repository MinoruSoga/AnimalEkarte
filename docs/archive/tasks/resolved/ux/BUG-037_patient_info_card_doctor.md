# BUG-037: PatientInfoCard に担当医フィールドが表示されない

## 概要
予約・カルテ画面の PatientInfoCard コンポーネントにペット名・種別は表示されるが、
担当医フィールドがUIに存在しない。

## 再現手順
1. `/reservations` で既存の予約をクリック
2. 表示される PatientInfoCard を確認
3. ペット名(Iris)・種(犬)は表示されるが、担当医名が表示されない

## 期待する動作
- PatientInfoCard に「担当医: 山田 太郎」のように担当医名が表示される

## 実際の動作
- 担当医フィールドがカードUIに存在しない
- 予約作成時に担当者未設定の場合も考慮が必要

## 実装場所
- `frontend/src/components/shared/PatientInfoCard/` コンポーネント
- `doctor` または `staff` プロパティを props として受け取り表示する

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-037（2026-03-28）
