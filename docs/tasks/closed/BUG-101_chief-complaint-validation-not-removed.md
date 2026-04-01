# BUG-101: chiefComplaint 必須バリデーション未削除（ドラフト保存不可）

## 概要

`use-medical-record-form.ts` に主訴（chiefComplaint）が DEFAULT テンプレートまたは空文字の場合に保存をブロックするバリデーションが残存している。
スペック変更（2026-03-30）では「バリデーション削除・ドラフト保存を可能にする」が要件だが、ステージング環境で未修正のまま。

## 症状

- 問診タブで「保存」クリック時、主訴が DEFAULT テンプレートのままだと「必須項目が未入力です」トーストが表示されて保存がブロックされる
- 空文字でも同様にブロックされる
- 診察/治療プランタブで保存しようとしても、chiefComplaint バリデーションが全タブ共通で発火するためブロックされる

## 期待する動作

- 主訴が DEFAULT・空欄でも保存を許可する（ドラフト保存）
- chiefComplaint の必須チェックは削除する

## 根本原因

`frontend/src/features/medical-records/hooks/use-medical-record-form.ts` L166–176:

```typescript
const errors: Record<string, string> = {};
if (!chiefComplaint || chiefComplaint === DEFAULT_CHIEF_COMPLAINT) {
  errors.chief_complaint = "主訴を入力してください";  // ← 削除対象
}
if (Object.keys(errors).length > 0) {
  setManualErrors(errors);
  toast.error("必須項目が未入力です");
  return { success: false, fieldErrors: errors, timestamp: Date.now() };
}
```

## 修正内容

- `use-medical-record-form.ts` L166–176 の `chiefComplaint` 必須バリデーションブロックを削除する
- Focus Management の `chief_complaint` タブリセット処理（L47–50）も合わせて削除する

## 影響ファイル

- `frontend/src/features/medical-records/hooks/use-medical-record-form.ts`

## 優先度

High

## 関連

- FUNCTIONAL_TEST_REPORT.md Section 46.4
- テスト確認日: 2026-04-01
- スペック変更: 2026-03-30（タブ別保存・ドラフト保存仕様）
