# FE-011: カルテ会計画面 — 飼主割引率を表示のみ（編集不可）に変更

**Status**: Open
**Priority**: Medium
**Affects**: medical-records feature — 会計(医師確認)タブ
**Date Created**: 2026-03-17
**Related**: TASK-001

## Summary

カルテの会計タブ（MedicalRecordBillCheck）で、飼主に設定された割引率（`owner.discount_rate`）を読み取り専用で表示する。現在はグローバル割引率/値引額が手入力可能だが、飼主の割引率は飼主フォームでのみ編集し、カルテ会計では表示のみとする。

## 現状のコード

### TreatmentDetailedSummary（割引UI）

```typescript
// frontend/src/features/medical-records/components/TreatmentDetailedSummary.tsx:54-70
// 現在: 割引率(%) と 値引額(¥) が編集可能な Input フィールド
<Input
  type="number"
  value={globalDiscountRate}
  onChange={(e) => setGlobalDiscountRate(Number(e.target.value))}
/>
<Input
  type="number"
  value={globalDiscountAmount}
  onChange={(e) => setGlobalDiscountAmount(Number(e.target.value))}
/>
```

### MedicalRecordBillCheck（会計タブ）

```typescript
// frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx:28-29
// 現在: ローカル state でグローバル割引を管理（飼主の割引率とは独立）
const [globalDiscountRate, setGlobalDiscountRate] = useState(0);
const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);
```

### MedicalRecordDiagnosisPlan（診療プランタブ）

```typescript
// frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx:59-60
const [globalDiscountRate, setGlobalDiscountRate] = useState(0);
const [globalDiscountAmount, setGlobalDiscountAmount] = useState(0);
```

### 飼主の割引率データフロー

```typescript
// frontend/src/features/owners/api/transforms.ts:39
discountRate: owner.discount_rate ?? 0,

// Owner 型: { discountRate: number }
```

## 必要な変更

### 1. MedicalRecordBillCheck — 飼主割引率の読み取り専用表示

```typescript
// frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx

// Before:
const [globalDiscountRate, setGlobalDiscountRate] = useState(0);

// After:
// 飼主データから割引率を取得（props or loader から）
// 編集不可にする
// globalDiscountRate は ownerDiscountRate（読み取り専用）に置き換え
```

飼主割引率の取得方法:
- MedicalRecordForm がすでに owner 情報を持っている場合は props で渡す
- なければカルテの owner_id から飼主情報を取得済みの箇所を確認する

### 2. TreatmentDetailedSummary — 割引率を読み取り専用に

```typescript
// frontend/src/features/medical-records/components/TreatmentDetailedSummary.tsx

// Before: 編集可能な Input
<Input type="number" value={globalDiscountRate} onChange={...} />

// After: 読み取り専用表示
// 飼主割引率を disabled または テキスト表示に変更
// 「飼主設定: XX%」のようなラベル表示
<div className="text-sm text-muted-foreground">
  飼主割引率: {ownerDiscountRate}%
</div>
```

### 3. MedicalRecordDiagnosisPlan — 同様に読み取り専用に

```typescript
// frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx
// 同様に globalDiscountRate を飼主割引率の表示のみに変更
```

### 4. 会計計算ロジック — 飼主割引率を反映

```typescript
// 小計の計算に飼主割引率を適用
// subtotal * (1 - ownerDiscountRate / 100) で割引後金額を算出
```

## UI 操作フロー

1. ユーザーがカルテ詳細画面を開く
2. 会計(医師確認)タブを選択
3. 飼主に設定された割引率が「飼主割引率: XX%」として表示される（編集不可）
4. 割引率は飼主フォームからのみ変更可能
5. 割引後の金額が自動計算される

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useTransition` で pending 管理（`useState(false)` + `setIsPending` 禁止）
- [ ] 型は `models.ts` から導出（手書き interface 禁止）

## 依存関係

- Backend 変更なし（Owner API は既に discount_rate を返している）
- カルテ詳細画面のローダーまたは state が owner 情報を持っているか確認が必要

## 完了条件

- [ ] 会計タブで飼主割引率が読み取り専用で表示される
- [ ] 割引率の Input フィールドが編集不可（disabled or テキスト表示）
- [ ] 割引後金額が正しく計算される
- [ ] 診療プランタブでも同様に読み取り専用
- [ ] 型エラーなし（`docker compose exec frontend pnpm build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend pnpm lint` パス）
