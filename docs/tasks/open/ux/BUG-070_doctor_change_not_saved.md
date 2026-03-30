# BUG-070: 担当医の変更がPATCHされない

## 概要
カルテ詳細画面で担当医を変更しても、フォームのローカル state のみ更新され、
保存 payload に担当医変更が含まれず、バックエンドにPATCHされない。

## 再現手順
1. カルテ詳細画面にアクセス
2. 担当医フィールドを変更
3. 保存をクリック
4. ページをリロード
5. → 担当医が元の値に戻る（変更が保存されていない）

## 期待する動作
- 担当医変更が保存 payload に含まれる
- PATCH /api/v1/medical-records/:id で担当医が更新される

## 実装場所
- `frontend/src/features/medical-records/` のフォームコンポーネント
- 保存時の payload に `staff_id` または対応フィールドを含める

## 優先度
Medium

## 関連
- FUNCTIONAL_TEST_REPORT.md BUG-070（担当医変更フロー）
- テスト確認日: 2026-03-30
