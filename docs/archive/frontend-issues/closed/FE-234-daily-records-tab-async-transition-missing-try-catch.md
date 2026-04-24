# FE-234: DailyRecordsTab の4つの async transition に try/catch がない

## 概要

`frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx` で
4つの `startXxxTransition(async () => ...)` コールバック内に `try/catch` がなく、
`mutateAsync()` が失敗してもエラーがサイレントに握り潰される。

## 問題コード

### `DailyRecordsTab.tsx:77-103`

```tsx
// Before: try/catch なし
const handleCreateDailyRecord = () => {
  startDailyRecordTransition(async () => {
    await createDailyRecord.mutateAsync(...);  // 失敗してもエラーが握り潰される
    onClose();
  });
};

const handleAddVital = () => {
  startVitalTransition(async () => {
    await addVital.mutateAsync(...);  // 同上
  });
};

const handleAddCareLog = () => {
  startCareLogTransition(async () => {
    await addCareLog.mutateAsync(...);  // 同上
  });
};

const handleAddStaffNote = () => {
  startStaffNoteTransition(async () => {
    await addStaffNote.mutateAsync(...);  // 同上
  });
};

// After: try/catch + handleApiError を追加
const handleCreateDailyRecord = () => {
  startDailyRecordTransition(async () => {
    try {
      await createDailyRecord.mutateAsync(...);
      onClose();
    } catch (error) {
      handleApiError(error, "日次記録の作成");
    }
  });
};
// 他3つも同様
```

## 影響

入院中の患者に対して：
- 日次記録の作成失敗 → ユーザー気づかず
- バイタル追加失敗 → サイレント
- ケアログ追加失敗 → サイレント
- スタッフメモ追加失敗 → サイレント

いずれも医療記録への影響があるため重大。

## 準拠すべきプロジェクト規約

### `.claude/rules/error-handling.md`
> すべての `catch` ブロックで `handleApiError(error, "コンテキスト")` を呼び出す。

### プロジェクト内参照実装
- `use-hospitalization-form.ts:92-122` — try/catch + handleApiError で正しく実装済み

## 優先度
**High** — 医療記録（バイタル・ケアログ）の保存失敗が無音で握り潰される。

## 関連ファイル
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx`
- `frontend/src/lib/handle-api-error.ts`
