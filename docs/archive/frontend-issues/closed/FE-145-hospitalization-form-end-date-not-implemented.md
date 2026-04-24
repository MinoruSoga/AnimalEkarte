# FE-145: 入院フォームの終了日フィールドが未実装（value="" / onChange=no-op）

**Status**: Open
**Priority**: High
**Affects**: features/hospitalization/components/HospitalizationBasicInfo.tsx
**Date Created**: 2026-03-29
**Related**: BUG-050

---

## Summary

入院登録フォームの「終了日」`NotionDatePicker` が `value=""` / `onChange={() => {}}` のプレースホルダー状態。
カレンダーから日付を選択しても値が反映されず、DB に保存もされない。

---

## 実装手順

### 1. `HospitalizationBasicInfo.tsx` の修正（76〜81行目付近）

現在：
```tsx
<NotionDatePicker
  value=""
  onChange={() => {}}
  placeholder="終了日"
  className="flex-1"
/>
```

修正後：
```tsx
<NotionDatePicker
  value={formData.endDate ?? ""}
  onChange={value => onChange("endDate", value)}
  placeholder="終了日"
  className="flex-1"
/>
```

### 2. フォーム state に `endDate` を追加

`useHospitalizationForm.ts`（または対応する hook）：

```typescript
interface HospitalizationFormData {
  // 既存フィールド...
  endDate: string | null;  // 追加
}

const initialValues: HospitalizationFormData = {
  // 既存...
  endDate: null,  // 追加
};
```

### 3. 登録 API 呼び出し時に `end_date` を含める

```typescript
const input: CreateHospitalizationInput = {
  // 既存フィールド...
  end_date: formData.endDate ? new Date(formData.endDate).toISOString() : null,
};
```

### 4. バックエンド API 確認

`PATCH /api/v1/hospitalizations/:id` で `end_date` を受け付けるか確認する。
`CreateHospitalizationInput` にも `end_date` が含まれているか確認する。

---

## 受入条件

- [ ] カレンダーから終了日を選択すると「終了日」フィールドに日付が表示される
- [ ] 登録時に `end_date` が DB に保存される
- [ ] 終了日未設定での登録も可能（null 許容）
- [ ] `docker compose exec frontend pnpm lint` エラー 0 件
