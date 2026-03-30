# BUG-050: 入院フォームの終了日がハードコード

## 概要
入院登録フォームの DateRangePicker において、終了日（退院予定日）の入力フィールドが `value=""` / `onChange={() => {}}` でハードコードされており、ユーザーが終了日を入力・変更できない。

## 再現手順
1. `/hospitalization` にアクセス
2. 「新規入院登録」ボタンをクリック
3. 入院期間の終了日フィールドを操作しようとする
4. → フィールドが反応しない（値が変更されない）

## 期待する動作
- 終了日（退院予定日）を自由に入力・選択できる
- 選択した終了日がフォームの state に反映される
- 保存時に終了日がAPIに送信される

## 実装場所
- `frontend/src/features/hospitalization/` のフォームコンポーネント
- DateRangePicker の終了日 `value` と `onChange` を正しく state に接続する

## 優先度
Medium

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-050
- テスト確認日: 2026-03-30
