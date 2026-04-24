# FE-008: 問診・治療プラン・見積書 API エンドポイント非呼び出し

## 問題
医療記録フォームで「問診」「診察/治療プラン」「見積書」が UI では「更新しました」と表示されるが、DB に保存されない。

**根本原因:** Frontend が separate API エンドポイント（inquiries, treatment-plans, estimates）を呼び出していない。

---

## アーキテクチャの乖離

### DB テーブル構造
```
medical_records          # 主レコード
├── inquiries（SEPARATE）
├── treatment_plans（SEPARATE）
└── estimates（SEPARATE）
```

### Backend API（想定）
```
✅ PATCH /api/v1/medical-records/:id           （main record のみ）
❌ PATCH /api/v1/medical-records/:id/inquiries     （frontend 非呼び出し）
❌ PATCH /api/v1/medical-records/:id/treatment-plans（frontend 非呼び出し）
❌ POST  /api/v1/medical-records/:id/estimates     （frontend 非呼び出し）
```

### Frontend 実装状況
```
✅ update-medical-record.ts  → PATCH /medical-records/:id のみ
❌ inquiries.ts             （存在しない）
❌ estimates.ts             （存在しない）
❌ useMedicalRecordForm.ts  → separate calls なし
```

---

## 具体例：Pattern B テスト結果

```javascript
// Frontend が送信するリクエスト
PATCH /api/v1/medical-records/17
{
  chief_complaint: "# どんな症状\n嘘吐、下痢",
  plan: "...",
  assessment: "...",
  notes: "..."
}

// ❌ 以下は送信されていない:
// PATCH /api/v1/medical-records/17/inquiries
// PATCH /api/v1/medical-records/17/treatment-plans
// POST /api/v1/medical-records/17/estimates
```

---

## ネットワーク検証
```
ブラウザ DevTools Network タブ:
reqid=3466 GET /api/v1/medical-records/17 [200]
reqid=3467 GET /api/v1/pets/15 [200]

❌ inquiries PATCH request
❌ treatment-plans PATCH request
❌ estimates POST request
```

---

## DB 確認結果

### 問診更新テスト後
```sql
SELECT updated_at FROM inquiries WHERE medical_record_id = 17;
→ 2026-03-14 18:11:29  ❌ 古いまま
```

### 治療プラン作成テスト後
```sql
SELECT COUNT(*) FROM treatment_plans WHERE medical_record_id = 17;
→ 0件  ❌（実装なし）
```

### 見積書作成テスト後
```sql
SELECT COUNT(*) FROM estimates WHERE medical_record_id = 17;
→ 0件  ❌（実装なし）
```

---

## 修正対応（Frontend）

### 1. API ファイル作成

**frontend/src/features/medical-records/api/inquiries.ts**
```typescript
export const useUpdateInquiry = (recordId: string) => {
  return useMutation({
    mutationFn: (input: UpdateInquiryRequest) =>
      axios.patch(`/v1/medical-records/${recordId}/inquiries`, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["medical-record", recordId] });
    },
  });
};
```

**frontend/src/features/medical-records/api/treatment-plans.ts**
```typescript
export const useUpdateTreatmentPlan = (recordId: string) => {
  return useMutation({
    mutationFn: (input: UpdateTreatmentPlanRequest) =>
      axios.patch(`/v1/medical-records/${recordId}/treatment-plans`, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["medical-record", recordId] });
    },
  });
};
```

**frontend/src/features/medical-records/api/estimates.ts**
```typescript
export const useCreateEstimate = (recordId: string) => {
  return useMutation({
    mutationFn: (input: CreateEstimateRequest) =>
      axios.post(`/v1/medical-records/${recordId}/estimates`, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["medical-record", recordId] });
    },
  });
};
```

### 2. Form Hook 修正

**frontend/src/features/medical-records/hooks/useMedicalRecordForm.ts**

```typescript
const handleSave = () => {
  // ✅ 修正: 複数の API calls を並列実行
  await Promise.all([
    updateMutation.mutateAsync({ id: recordId, req: mainRecordReq }),
    inquiriesMutation.mutateAsync({ id: recordId, req: inquiriesReq }),
    treatmentPlansMutation.mutateAsync({ id: recordId, req: treatmentPlansReq }),
    estimatesMutation.mutateAsync({ id: recordId, req: estimatesReq }),
  ]);

  toast.success("カルテを更新しました");
};
```

---

## テスト環境
- 記録 ID: 17
- テスト日時: 2026-03-16 12:00 JST

---

## 優先度
**🔴 CRITICAL** - データ永続化の完全な失敗

---

## 関連イシュー
- **BE-003, BE-004, BE-005** は symptom（症状）で、本イシューが root cause（根本原因）
- **FE-006** （主訴区分ハードコード）も同じ構造的問題

