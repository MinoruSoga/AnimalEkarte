# ISSUE: バリデーションエラー表示の統一（toast → インライン FormFieldError）

## 概要

フォームのバリデーションエラーをトースト通知ではなく、各フォームフィールドの直下に `FormFieldError` でインライン表示するよう統一する。

**対象外:**
- API エラー（`handleApiError` 経由）→ toast のまま OK
- `ImageGalleryFilter.tsx` のファイルサイズ警告 → 処理が継続するため toast が適切
- `use-medical-record-form.ts` の `setManualErrors` → フォーカス管理専用。ただし下記 Cat.A に含む

**修正済み（参考）:**
- `ReservationFormModal.tsx` ✅
- `PetEditModal.tsx` ✅
- `use-owner-form.ts`（フォームフック内バリデーション）✅
- `use-examination-form.ts` ✅
- `use-inventory-form.ts` ✅

---

## 対象ファイル一覧

### Cat.A — useActionState フォームフック（fieldErrors 返却 + コンポーネント側に FormFieldError 追加）

| ファイル | 現状 | 修正内容 |
|---------|------|---------|
| `features/hospitalization/hooks/use-hospitalization-form.ts:88` | `toast.error("ペットを選択してください")` + `{ success: false }` | `fieldErrors: { pet: "..." }` を返す。`HospitalizationForm.tsx` に `FormFieldError` 追加 |
| `features/estimates/hooks/use-estimate-form.ts:70` | `toast.error("タイトルを入力してください")` + `{ success: false }` | `fieldErrors: { title: "..." }` を返す。`EstimateForm.tsx` に `FormFieldError` 追加 |
| `features/medical-records/hooks/use-medical-record-form.ts:190-191` | `setManualErrors(diagError)` + `toast.error("診断名を選択してください")` | toast を削除。`MedicalRecordDiagnosisPlan.tsx` 等の診断名フィールド直下に `FormFieldError` 追加 |

### Cat.B — コンポーネント内バリデーション（useState + setFieldErrors に切り替え + FormFieldError 追加）

| ファイル | 現状 | 修正内容 |
|---------|------|---------|
| `features/medical-records/components/CheckupsTab/CheckupsTab.tsx:210` | `toast.error("日付と健診種別IDは必須です")` | `setFieldErrors` 追加。日付・健診種別フィールド下に `FormFieldError` 表示 |
| `features/medical-records/components/VitalsTab/VitalsTab.tsx:125/130/136/325/330/336/345` | 複数の `toast.error` | `setFieldErrors` 追加。各フィールド直下に `FormFieldError` 表示 |
| `features/medical-records/components/MedicalRecordEstimate.tsx:103` | `toast.error("件名を入力してください")` | `setFieldErrors` 追加。件名フィールド下に `FormFieldError` 表示（※ line 99 の「カルテ保存前」警告は toast のまま可） |

### Cat.C — インライン編集セル（値リセット + toast → フィールド直下エラー表示に変更）

| ファイル | 現状 | 修正内容 |
|---------|------|---------|
| `features/medical-records/components/TreatmentsTab/TreatmentRow.tsx:132/155` | `toast.error("金額は0以上を入力してください")` + 値リセット | インライン編集セル直下に一時エラー表示。値リセットは維持 |
| `features/accounting/routes/AccountingDetail.tsx:348/352` | `toast.error("単価は0以上...")` / `toast.error("単価は999,999,999円以下...")` | 入力フィールド直下に `FormFieldError` 表示 |

### Cat.D — マスタ設定（共通フック `use-master-save.ts` の改修 + 各設定画面に FormFieldError 追加）

`use-master-save.ts` の `validate()` 返却値（エラー文字列）を `fieldErrors` として上位コンポーネントに返すよう変更し、各マスタ設定画面の入力フィールド下に `FormFieldError` を追加する。

| ファイル | 対象フィールド |
|---------|-------------|
| `features/master/hooks/use-master-save.ts:52` | 共通バリデーション（`toast.error(error)` → fieldErrors返却） |
| `features/master/routes/TreatmentPlanMaster.tsx:712` | 名称フィールド |
| `features/hospital-settings/routes/ClinicMasterSettings.tsx:170` | 院名フィールド |
| `features/master/routes/DiagnosisSettings.tsx:534/570/574` | カテゴリ名・診断病名・カテゴリ選択 |
| `features/master/routes/TrimmingSettings.tsx:554/597` | コース名・オプション名 |
| `features/master/routes/MedicineSettings.tsx:640` | 薬品名 |

### Cat.E — その他

| ファイル | 現状 | 対応方針 |
|---------|------|---------|
| `features/owners/hooks/use-owner-form.ts:320/329` | `toast.error("動物種を選択してください")`（ペット作成フロー） | `PetEditModal` 経由のため、モーダル内の `animalSpeciesId` フィールド下に `FormFieldError` 追加 |

---

## 実装方針

### Cat.A の標準パターン（参考: ExaminationForm）

```tsx
// hook 側
if (Object.keys(errors).length > 0) {
  // toast.error を削除
  return { success: false, fieldErrors: errors, timestamp: Date.now() };
}

// コンポーネント側
<FormFieldError message={formState.fieldErrors?.pet} />
```

### Cat.B の標準パターン

```tsx
// コンポーネント内
const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

const handleSave = useCallback(() => {
  const errors: Record<string, string> = {};
  if (!form.recorded_at) errors.recorded_at = "記録日時は必須です";
  // ...
  if (Object.keys(errors).length > 0) {
    setFieldErrors(errors);
    // toast.error を削除
    return;
  }
  setFieldErrors({});
  // ...
}, [...]);

// JSX
<FormFieldError message={fieldErrors.recorded_at} />
```

---

## 優先度

| Cat | 優先度 | 理由 |
|-----|--------|------|
| A | HIGH | メインフォーム。UX への影響が大きい |
| B | HIGH | カルテ内の主要タブ。頻繁に操作される |
| E | MEDIUM | ペット追加フロー |
| C | MEDIUM | インライン編集（影響範囲が局所的） |
| D | LOW | マスタ設定（操作頻度が低い） |

---

## 備考

- `use-master-save.ts` は共通フックのため、Cat.D の変更は一括で効く可能性がある
- `VitalsTab.tsx` は新規追加・編集の両モードで重複した `handleSave` があるため注意
- `use-medical-record-form.ts` の `setManualErrors` によるフォーカス管理ロジックは維持する
