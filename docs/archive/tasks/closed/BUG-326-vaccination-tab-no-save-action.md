# BUG-326: カルテ「予防接種」タブの保存機能が未実装

**Status**: CLOSED  
**Priority**: High  
**Discovery**: 機能テスト Section 4.3 予防接種タブ (2026-04-12)

## 概要

カルテ編集の「予防接種」タブでワクチン名・接種日等を入力して「保存」ボタンをクリックしても、入力データが一切 API に送信されず破棄される。`VaccinationForm` コンポーネントに保存ボタンも mutation call も存在しない。

また `MedicalRecordVaccination.tsx:19` の `useState("esophagitis")` というデフォルト値は疾患名であり、ワクチン選択コンボボックスの初期値として不正。

## 再現手順

1. `/medical-records/21` を開き「予防接種」タブに切り替える
2. 予防接種名から「混合ワクチン5種（犬）」を選択
3. 予防接種日に「2026/04/12」を入力
4. 「保存」ボタンをクリック
5. **結果**: `POST /v1/vaccinations` は発行されず、接種履歴に記録が追加されない
6. **期待**: 接種記録が保存され、「予防接種履歴」テーブルに追加される

## 現状コード

### `frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx:19`
```tsx
// ❌ 誤ったデフォルト値（疾患名であってワクチン名ではない）
const [vaccineName, setVaccineName] = useState("esophagitis");
```

### `frontend/src/features/medical-records/components/VaccinationForm.tsx`
```tsx
// ❌ 保存ボタンなし・mutation 呼び出しなし
export const VaccinationForm = memo(function VaccinationForm({ ... }: VaccinationFormProps) {
  return (
    <div className="col-span-6 flex flex-col gap-4">
      {/* フォームフィールドのみ。<Button>も createVaccination() 呼び出しも存在しない */}
    </div>
  );
});
```

### `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:134-149`
```tsx
// ❌ 予防接種タブのサブ保存が登録されていない
useEffect(() => {
  if (!formState.success) return;
  try {
    if (currentTab === "診察/治療プラン") {
      await (clinicalPlanSaveRef.current?.() ?? Promise.resolve());
    } else if (currentTab === "見積書") {
      await (estimateSaveRef.current?.() ?? Promise.resolve());
    }
    // ← 予防接種タブの保存ハンドラが存在しない
  } catch { ... }
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx` | 保存ロジックなし・デフォルト値不正 | ❌ 未修正 |
| `frontend/src/features/medical-records/components/VaccinationForm.tsx` | 保存ボタン・mutation なし | ❌ 未修正 |
| `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx` | 予防接種タブのサブ保存未登録 | ❌ 未修正 |
| `POST /v1/vaccinations` | カルテ内の予防接種登録 | ❌ 呼び出されない |

## 修正方針

### 参照実装
既存の予防接種登録機能 `frontend/src/features/vaccinations/` に同等の保存ロジックが実装済み。  
`useCreateVaccination` hook を利用する。

### 1. `MedicalRecordVaccination.tsx` — 保存ロジック追加 + デフォルト値修正

```tsx
import { useCreateVaccination } from "@/features/vaccinations";
import { useCallback } from "react";
import { handleApiError } from "@/lib/handle-api-error";

export const MedicalRecordVaccination = memo(function MedicalRecordVaccination({ petId }) {
  // ✅ 正しいデフォルト値（空文字列）
  const [vaccineName, setVaccineName] = useState("");
  const [date, setDate] = useState("");
  // ...他フィールド...

  const { mutateAsync: createVaccination } = useCreateVaccination();

  const handleSave = useCallback(async () => {
    if (!petId || !vaccineName || !date) return;
    try {
      await createVaccination({
        pet_id: petId,
        vaccine_id: vaccineName,
        date,
        lot1, lot2, lot3, lot4,
        supplemental,
        next_date: nextDate,
        remarks,
      });
      // フォームをリセット
      setVaccineName("");
      setDate("");
      // ...他フィールドリセット...
    } catch (err) {
      handleApiError(err, "予防接種の登録");
    }
  }, [petId, vaccineName, date, ...]);

  return (
    <div ...>
      <VaccinationForm
        ...
        onSave={handleSave}   // ← prop 追加
      />
      ...
    </div>
  );
});
```

### 2. `VaccinationForm.tsx` — 保存ボタン追加

```tsx
interface VaccinationFormProps {
  ...
  onSave?: () => void;   // ← 追加
}

// JSX 末尾に追加
<Button type="button" onClick={onSave} className={STYLE.btnPrimary}>
  接種記録を追加
</Button>
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Error Handling

> catch ブロックでは必ず `handleApiError` を呼び出す

### `.claude/rules/typescript-react.md` — React 19 Patterns

> データ更新（Mutation）は **React 19 Action** パターンを標準とする

ただし、このフォームは医療記録の主フォームに nested しているため、`useTransition` + `onClick` パターン（`useActionState` の nested form 回避）を使用すること。

### プロジェクト内参照実装
`frontend/src/features/vaccinations/routes/VaccinationForm.tsx` — 同様の予防接種フォームに完全な保存ロジックが実装済み。

## 優先度

**High** — カルテ画面からの予防接種登録が全くできない機能不全
