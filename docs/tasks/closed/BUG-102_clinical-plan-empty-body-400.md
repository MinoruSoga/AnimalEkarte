# BUG-102: PATCH /clinical-plan が全フィールド未入力時に 400 を返す

## 概要

診察/治療プランタブで治療方針・診断詳細・診断マスタをすべてデフォルト値のまま保存すると、
`PATCH /api/v1/medical-records/:id/clinical-plan` が HTTP 400 "at least one field must be provided" を返す。

## 症状

- 診察/治療プランタブを開き、何も変更せずに「保存」クリック
- "at least one field must be provided" エラートーストが表示される
- 保存失敗（BE 400）
- 診断カテゴリ・病名をともに未選択のまま保存しても同様に失敗

## 期待する動作

- フィールドが空（未入力）の場合でも 200 で保存成功する（何も変更しない PATCH として受け入れる）
- または、FE 側で全フィールドが空の場合は API 呼び出しをスキップし「保存しました」トーストを表示する

## 根本原因

### FE 側

`use-medical-record-form.ts` の "診察/治療プラン" ケースで、plan/assessment が DEFAULT 値と一致する場合 `undefined` を送信:

```typescript
await updateTreatmentPlanMutation.mutateAsync({
  treatment_policy: plan !== DEFAULT_PLAN ? plan : undefined,         // DEFAULT → undefined
  diagnosis_details: assessment !== DEFAULT_ASSESSMENT ? assessment : undefined, // DEFAULT → undefined
  diagnosis_category_id: diagnosis1CategoryId ?? undefined,           // null → undefined
  // ...すべて undefined
});
```

### BE 側

`PATCH /v1/medical-records/:id/clinical-plan` ハンドラが、すべてのフィールドが nil/未指定の場合に `400 "at least one field must be provided"` を返すバリデーションを持っている。

## 修正案

### 案A（FE 修正）: 全フィールド undefined の場合は API 呼び出しをスキップ

```typescript
case "診察/治療プラン": {
  const payload = {
    treatment_policy: plan !== DEFAULT_PLAN ? plan : undefined,
    // ...
  };
  const hasAnyField = Object.values(payload).some(v => v !== undefined);
  if (hasAnyField) {
    await updateTreatmentPlanMutation.mutateAsync(payload);
  }
  // hasAnyField が false の場合は API をスキップして保存成功扱い
  break;
}
```

### 案B（BE 修正）: 空 body を受け入れる（no-op として 200 返却）

BE の `clinical-plan` PATCH ハンドラから "at least one field" バリデーションを削除し、全フィールドが nil の場合は何もせずに 200 を返す。

## 影響ファイル

- `frontend/src/features/medical-records/api/treatment-plans.ts`
- `frontend/src/features/medical-records/hooks/use-medical-record-form.ts`
- `backend/internal/handler/medical_record_handler.go`（または clinical_plan_handler.go）

## 優先度

Medium

## 関連

- FUNCTIONAL_TEST_REPORT.md Section 46.5
- テスト確認日: 2026-04-01
- BUG-101 と合わせて修正すること（chiefComplaint バリデーション削除後に本バグが表面化する）
