# FE-166: 複数フォーム — canEdit=false でもフォームフィールドが disabled でない（システム的欠如）

## 概要

以下の複数フォームで `canEdit=false` のユーザーに対して `SubmitButton`（保存ボタン）は非表示になっているが、フォームの各入力フィールドが `disabled={!canEdit}` になっておらず、フィールドの操作が可能な状態で表示される。また `<form action={formAction}>` 形式では Enter キーでのフォーム送信が試みられる可能性がある。

## 影響ファイル一覧

| ファイル | canEdit ガード | フィールド disabled | 入力フィールド数 |
|---------|--------------|-------------------|---------------|
| `VaccinationForm.tsx` | ✅ SubmitButton (行 164) | ❌ なし | 6+ (日付・ワクチン・LOT 1-4・補助説明) |
| `TrimmingForm.tsx` | ✅ SubmitButton (行 558) | ❌ なし | 多数 (日時・スタッフ・メニュー・金額等) |
| `HospitalizationForm.tsx` | ✅ SubmitButton (行 170) | ❌ なし | 多数 (入院種別・日付・担当医等) |
| `ExaminationForm.tsx` | ✅ SubmitButton (行 177) | ❌ (isConfirmed のみ) | 多数 (検査日・検査結果等) |
| `InventoryForm.tsx` | ✅ SubmitButton (行 337/376) | ❌ なし | 多数 (品目名・数量・金額・入荷日等) |
| `AccountingDetail.tsx` | ✅ SubmitButton (行 572) | ❌ なし | 多数 (支払方法・受取金額・保険割合・返金等) |

注: `OwnerForm.tsx` は FE-158 で既に報告済み。

## 現状の挙動（バグ）

```tsx
// VaccinationForm.tsx — SubmitButton はガード済み ✅
{canEdit ? <SubmitButton>保存</SubmitButton> : null}

// しかし各フィールドに disabled がない ❌
<Input
  value={supplemental}
  onChange={(e) => setSupplemental(e.target.value)}
  // disabled={!canEdit}  ← なし
/>
<Select value={vaccineId} onValueChange={setVaccineId}>
  {/* disabled なし ❌ */}
</Select>
```

`canEdit=false` のユーザーが各フォームを開くと：
1. 全フィールドが入力可能な状態で表示される
2. ユーザーが値を変更できてしまう（見た目の誤操作）
3. `<form action={formAction}>` があれば Enter キーで送信が試みられる
4. 保存ボタンがないため、最終的な API 呼び出しは発生しないことが多い

## 深刻度の評価

- `<form action={formAction}>` パターンを使っているフォームでは Enter キー送信のリスクあり（HIGH）
- `useState + onSubmit` パターンのフォームでは Enter キー送信なし（MEDIUM）
- いずれも閲覧のみユーザーへの誤った操作可能感（UX 問題）

## 修正方針

### 方針 A: fieldset disabled で一括 disable（最小変更）

```tsx
<fieldset disabled={!canEdit}>
  {/* 全フォームフィールド */}
  <Input value={supplemental} onChange={...} />
  <Select ...>...</Select>
  <NotionDatePicker ... />
</fieldset>
```

HTML の `<fieldset disabled>` は内包する全フォームコントロールを一括 disable にする。

### 方針 B: 各フィールドに disabled={!canEdit} を追加

```tsx
<Input
  value={supplemental}
  onChange={(e) => setSupplemental(e.target.value)}
  disabled={!canEdit}  // ← 追加
/>
<Select disabled={!canEdit} ...>
```

方針 A が変更箇所最小で推奨。

## 優先度

- **VaccinationForm・HospitalizationForm**: HIGH（form action パターン確認が必要）
- **TrimmingForm**: HIGH
- **ExaminationForm**: MEDIUM（`isConfirmed` による別ガードあり、かつ canEdit でフォームアクション制御）

## 関連ファイル

- `frontend/src/features/vaccinations/routes/VaccinationForm.tsx` (行 195-270)
- `frontend/src/features/trimming/routes/TrimmingForm.tsx`
- `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx`
- `frontend/src/features/examinations/routes/ExaminationForm.tsx` (行 84-157)
- `frontend/src/features/inventory/routes/InventoryForm.tsx` (行 67-259: 全フィールド)
- `frontend/src/features/accounting/routes/AccountingDetail.tsx` (行 426-680: 支払・保険・返金フィールド)
- `frontend/src/features/hospitalization/components/HospitalizationBasicInfo.tsx` (canEdit props なし・全フィールド常時編集可 — HospitalizationForm の sub-component)
- `frontend/src/features/hospitalization/components/HospitalizationNoteCard.tsx` (canEdit props なし・Textarea 常時編集可 — HospitalizationForm の sub-component)
- 発見日: 2026-04-08（RBAC Phase 2 テスト中）
- 関連: FE-158（OwnerForm の同一問題）
