# BUG: 検査管理 - 検査日入力フィールドが UI に実装されていない

**Issue Date**: 2026-04-02
**Severity**: High
**Status**: New
**Component**: Examination Management (検査管理)
**Related**: BUG-109 (入院スタブ API → 実装接続)

## 概要

検査フォーム（ExaminationForm）で、**検査日（date）フィールドが UI に実装されていません**。

- **useExaminationForm** の formData には `date` フィールドが存在する（use-examination-form.ts L99）
- **CreateExaminationRequest** API リクエストに `date: string` が含まれている（api/types.ts L15）
- しかし **ExaminationForm.tsx** の FormFieldsSection に対応する UI フィールドが**ない**

## 再現手順

1. `/examinations` に移動
2. 新規検査を作成（ペット選択フロー完了）
3. 検査フォームが表示される
4. **期待**: 「検査日」の日付入力フィールドが表示される
5. **実際**: 表示フィールド:
   - 検査種別 ✅
   - 担当医 ✅
   - ステータス ✅
   - 備考・所見 ✅
   - （検査日 ❌ **見当たらない**）

## 詳細

### UIコード現状

**frontend/src/features/examinations/routes/ExaminationForm.tsx L59-149**

```typescript
const FormFieldsSection = memo(function FormFieldsSection({
  formData,
  examTypes,
  staffList,
  isEdit,
  isSaving,
  isDeleting,
  isConfirmed,
  canEdit,
  canDelete,
  onSetFormData,
  onBack,
  onDeleteClick,
}: FormFieldsSectionProps) {
  return (
    <div className={`bg-white p-4 rounded-lg border ${C.borderMedium} space-y-4 shadow-sm`}>
      {/* ... */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* 検査種別 */}
        <div> ... </div>
        {/* 担当医 */}
        <div> ... </div>
      </div>

      {/* ステータス */}
      <div className="space-y-1.5"> ... </div>

      {/* 備考・所見 */}
      <div className="space-y-1.5"> ... </div>

      {/* ❌ 検査日フィールド: 実装されていない */}
    </div>
  );
});
```

### API リクエスト型との乖離

**api/types.ts L10-18**

```typescript
export interface CreateExaminationRequest {
  medical_record_id?: number | null;
  pet_id?: number | null;
  exam_type_id: number;
  doctor_id?: number | null;
  date: string;  // ← 必須フィールド（API では使用されている）
  machine?: string;
  result_summary?: string;
}
```

**use-examination-form.ts L114**

```typescript
const req: CreateExaminationRequest = {
  medical_record_id: medicalRecordId ? Number(medicalRecordId) : null,
  pet_id: Number(pet.id) || null,
  exam_type_id: Number(current.testTypeId) || 0,
  doctor_id: current.doctorId ? Number(current.doctorId) : null,
  date: current.date ?? new Date().toISOString(),  // ← デフォルトで現在時刻が送信される
  result_summary: current.resultSummary,
  machine: current.machine,
};
```

### 問題点

1. **UI フィールドがない**: ユーザーは検査日を入力できない
2. **デフォルト値が送信される**: `current.date ?? new Date().toISOString()` により、現在時刻が自動送信される
   - ユーザーの意図しない日付が記録される可能性
   - 過去の検査日を入力したい場合、対応不可
3. **将来日付バリデーション未実装**: 日付フィールドがないため、バリデーション自体が実装不可

## テスト報告書との齟齬

FUNCTIONAL_TEST_REPORT.md L4997-4999:

```
| 検査日 未入力で「保存」 | OK | バリデーションエラー表示確認 |
| 担当医 未選択で「保存」 | OK | ... |
| 検査日 将来日付入力 | NG | use-examination-form.ts: date に将来日付バリデーションなし。未来の検査日で登録可能 |
```

→ これらは**実装されていない機能に対するテスト**であり、実際には動作していません。

## 修正方法

### フロントエンド（必須）

ExaminationForm.tsx の FormFieldsSection に検査日フィールドを追加：

```typescript
<div className="space-y-1.5">
  <Label className={`text-sm ${C.text60}`}>検査日</Label>
  <Input
    type="date"
    value={formData.date ? formData.date.split("T")[0] : ""}
    disabled={isConfirmed}
    onChange={(e) => {
      const dateStr = e.target.value;
      const isoString = dateStr ? `${dateStr}T00:00:00Z` : "";
      onSetFormData({ date: isoString });
    }}
    max={new Date().toISOString().split("T")[0]}  // 将来日付禁止
    className={`${STYLE.formInput}`}
  />
</div>
```

### バリデーション（フロントエンド）

use-examination-form.ts の formAction callback に以下を追加：

```typescript
// 検査日バリデーション
const dateStr = current.date;
if (dateStr) {
  const examDate = new Date(dateStr);
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  if (examDate > today) {
    errors.date = "検査日は本日以前の日付を選択してください";
  }
}

if (Object.keys(errors).length > 0) {
  toast.error("未入力または不正な項目があります");
  return { success: false, fieldErrors: errors, timestamp: Date.now() };
}
```

## テスト計画

- [ ] 検査日フィールドが表示される
- [ ] 過去日付を入力して保存 → 成功
- [ ] 今日の日付を入力して保存 → 成功
- [ ] 未来日付を入力して保存 → エラー「検査日は本日以前の日付を選択してください」
- [ ] 検査日を入力しないで保存 → デフォルト（現在時刻）で送信されるか、または必須化

## 優先度

**High** - ユーザーが検査日を入力できない実装上の欠陥

---

**発見者**: Claude Code (2026-04-02 コード検査)
**最終更新**: 2026-04-02
