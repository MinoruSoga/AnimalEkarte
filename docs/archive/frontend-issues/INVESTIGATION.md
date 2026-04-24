# Frontend Issues Investigation

## Summary
調査対象: open ディレクトリ内 FE-004, FE-007, FE-009, FE-010

### FE-009: マスタ編集ボタン - ナビゲーションパス未定義

**Status: ✅ RESOLVED**

- paths.ts 確認済み（行232-241）
- `paths.settings.interview.chiefComplaint` 定義済み
- `paths.settings.interview.interviewTemplate` 定義済み
- InterviewChiefComplaint.tsx が正しく参照可能

**Action**: closed に移動 + commit

---

### FE-010: 医師選択ボタン - 実装未完成

**Status: ✅ RESOLVED**

- MedicalRecordForm.tsx:
  - 行79: `const [isStaffModalOpen, setIsStaffModalOpen] = useState(false);`
  - 行168-174: `handleSelectStaff`, `handleStaffModalOpenChange` 実装済み
  - 行203: `onStaffClick={() => setIsStaffModalOpen(true)}` でモーダル開閉
  - 行392-397: `StaffSelectionModal` 統合済み

**Action**: closed に移動 + commit

---

### FE-004: 予防接種タブ - 日付ピッカーのインタラクション問題

**Status: ⚠️ INVESTIGATION NEEDED**

- VaccinationForm.tsx: NotionDatePicker 使用（行86, 216）
- NotionDatePicker.tsx: shadcn/ui Calendar コンポーネント使用
- エラーレポート: "Cannot read properties of null (reading 'nodeType')"
- 原因: Calendar コンポーネント内での DOM参照エラーの可能性

**Next**: Calendar コンポーネント調査 → 修正提案

---

### FE-007: 診察/治療プラン - 診断コンボボックス マスタデータ検証不可

**Status: ⚠️ INVESTIGATION NEEDED**

- MedicalRecordDiagnosisPlan.tsx: 診断コンボボックス実装確認
- ドロップダウンオプションが表示されない（UI側の問題）
- Backend API との連携確認が必要

**Next**: DiagnosisHeader コンポーネント調査 → API 連携確認

---

## Proposed Actions

1. FE-009 / FE-010 を closed/ に移動
2. FE-004 / FE-007 は詳細分析後に修正実装か close判定
