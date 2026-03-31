# BUG-090: 検査実施日のデフォルト値が本日日付でない

## 概要
CheckupsTab.tsx の EMPTY_ADD_FORM で `date=""` (空文字) になっており、
検査追加フォームの実施日フィールドに本日の日付がデフォルト入力されない。

## 再現手順
1. カルテ詳細 → 検査タブを開く
2. 「検査追加」をクリック
3. → 実施日フィールドが空（本日日付のデフォルト値なし）

## 期待する動作
- 実施日フィールドに本日の日付（`new Date().toISOString().slice(0, 10)`）がデフォルト表示

## 実装場所
- `frontend/src/features/medical-records/` の CheckupsTab.tsx
- `EMPTY_ADD_FORM` の `date` フィールドを `new Date().toISOString().slice(0, 10)` に変更

## 優先度
Low

## 関連
- FUNCTIONAL_TEST_REPORT.md line 6205
- テスト確認日: 2026-03-30
