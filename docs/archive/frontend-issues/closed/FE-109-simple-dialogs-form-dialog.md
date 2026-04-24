# FE-109: シンプルダイアログ3件を FormDialog に移行

**Status**: Closed
**Priority**: Medium
**Affects**: medical-records, hospitalization
**Date Created**: 2026-03-25
**Related**: TASK-025

## Summary

`Dialog + DialogHeader + DialogFooter + Button×2` を手書きしているシンプルなダイアログ3件を、共有 `FormDialog` に移行する。機能変更なし・UI 変更なし。

## 対象ファイル

| ファイル | タイトル | 現状 |
|---------|---------|------|
| `features/medical-records/components/BillingReviewSection/ReturnReasonDialog.tsx` | 差し戻し理由 | Dialog 直書き、保存ボタンが `C.bgDanger` 色指定 |
| `features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx` | ケアログ追加 | Dialog 直書き、`max-w-sm` |
| `features/hospitalization/components/DailyRecordsTab/DailyStaffNotesSection.tsx` | スタッフメモ追加 | Dialog 直書き、`max-w-sm` |

**移行しないもの**（複雑なため）:
- `PetEditModal`, `ReservationDetailModal`, `DashboardDetailModal` — footer に独自ロジック・複数ボタンあり
- `AccountingDetail`, `VitalsModal`, `StaffSelectionModal` — 選択・確認系で save/cancel パターン外

## 現状のコード

```typescript
// ReturnReasonDialog.tsx（抜粋）
<Dialog open={open} onOpenChange={handleOpenChange}>
  <DialogContent>
    <DialogHeader>
      <DialogTitle>差し戻し理由</DialogTitle>
      <DialogDescription>差し戻しの理由を入力してください。</DialogDescription>
    </DialogHeader>
    <div className="py-2 space-y-2">
      <Label>差し戻し理由 <span>*</span></Label>
      <Textarea ... />
    </div>
    <DialogFooter>
      <Button variant="outline" onClick={...} disabled={isPending}>キャンセル</Button>
      <Button onClick={handleSubmit} disabled={!reason.trim() || isPending}
        className={`${C.bgDanger} hover:bg-[#EB5757]/90 text-white`}>
        差し戻す
      </Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

## 必要な変更

### 1. ReturnReasonDialog.tsx

**注意**: 保存ボタンが「差し戻す（danger 色）」なため、`FormDialog` の `saveLabel` + `saveClassName` prop を使う。`FormDialog` に `saveClassName?: string` prop を追加する必要あり。

**Before:**
```typescript
import { Dialog, DialogContent, ..., DialogFooter } from "@/components/ui/dialog";
```

**After:**
```typescript
import { FormDialog } from "@/components/shared/FormDialog/FormDialog";
// ...
<FormDialog
  open={open}
  onClose={() => handleOpenChange(false)}
  title="差し戻し理由"
  description="差し戻しの理由を入力してください。"
  onSave={handleSubmit}
  saveLabel="差し戻す"
  saveClassName={`${C.bgDanger} hover:bg-[#EB5757]/90 text-white`}
  isPending={isPending}
  isSaveDisabled={!reason.trim()}
>
  <div className="py-2 space-y-2">
    <Label>差し戻し理由 <span className={C.textRequired}>*</span></Label>
    <Textarea ... />
  </div>
</FormDialog>
```

`FormDialog` への追加 props:
- `saveClassName?: string` — 保存ボタンの追加 className
- `isSaveDisabled?: boolean` — 保存ボタンの disabled 条件（`isPending` に加えて）

### 2. DailyCareLogsSection.tsx

```typescript
// Before: Dialog + DialogFooter インライン
// After:
<FormDialog
  open={isOpen}
  onClose={() => setIsOpen(false)}
  title="ケアログ追加"
  onSave={handleSave}
  saveLabel="追加"
  className="max-w-sm"
>
  {/* 既存の入力フィールド */}
</FormDialog>
```

### 3. DailyStaffNotesSection.tsx

```typescript
// Before: Dialog + DialogFooter インライン
// After:
<FormDialog
  open={isOpen}
  onClose={() => setIsOpen(false)}
  title="スタッフメモ追加"
  onSave={handleSave}
  saveLabel="追加"
  className="max-w-sm"
>
  {/* 既存の入力フィールド */}
</FormDialog>
```

### 4. FormDialog.tsx に props 追加

```typescript
// frontend/src/components/shared/FormDialog/FormDialog.tsx
export interface FormDialogProps {
  // ... 既存 props ...
  /** 保存ボタンの追加 className（danger 等の色変更に使用） */
  saveClassName?: string;
  /** 保存ボタンの disabled 条件（isPending に加えて） */
  isSaveDisabled?: boolean;
}

// Button の変更
<Button
  onClick={onSave}
  disabled={isPending || isSaveDisabled}
  className={`bg-[#2EAADC] text-white h-11 px-4 text-sm font-medium ${saveClassName ?? ""}`}
>
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）

## 依存関係

- `FormDialog` は既に実装済み
- BE 変更なし

## 完了条件

- [ ] 対象3ファイルで `from "@/components/ui/dialog"` の import が削除されている
- [ ] `FormDialog` に `saveClassName` / `isSaveDisabled` props が追加されている
- [ ] 差し戻しダイアログ・ケアログ追加・スタッフメモ追加の UI が変更前と同一
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス（エラー 0）

## クローズ情報

- **Closed At**: 2026-03-25
- **変更ファイル**:
  - `components/shared/FormDialog/FormDialog.tsx` — `saveClassName` / `isSaveDisabled` props 追加
  - `features/medical-records/components/BillingReviewSection/ReturnReasonDialog.tsx` — FormDialog 移行
  - `features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx` — FormDialog 移行
  - `features/hospitalization/components/DailyRecordsTab/DailyStaffNotesSection.tsx` — FormDialog 移行
