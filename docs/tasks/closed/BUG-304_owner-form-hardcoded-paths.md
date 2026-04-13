# BUG-304: OwnerForm — ハードコード URL パス6箇所

## 概要

`frontend/src/features/owners/routes/OwnerForm.tsx` 内の `PetTableRow` コンポーネントで `config/paths.ts` 規約違反。`backFrom` 変数および5つの `navigate()` 呼び出しにリテラル文字列が使われていた。

## 影響ファイル

| ファイル | 違反箇所 |
|---------|---------|
| `frontend/src/features/owners/routes/OwnerForm.tsx` | line 99, 141, 146, 152, 158, 164 |

## 違反箇所と修正

### OwnerForm.tsx
```tsx
// Before (line 99)
const backFrom = ownerId ? `/owners/${ownerId}` : "/owners";

// After
const backFrom = ownerId
  ? paths.owners.detail.getHref(ownerId)
  : paths.owners.getHref();
```

```tsx
// Before (line 141)
navigate(`/reservations?petId=${pet.id}`)

// After
navigate(`${paths.reservations.getHref()}?petId=${pet.id}`)
```

```tsx
// Before (line 146)
navigate(`/medical-records/new?petId=${pet.id}`, { state: { from: backFrom } })

// After
navigate(`${paths.medicalRecords.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } })
```

```tsx
// Before (line 152)
navigate(`/trimming/new?petId=${pet.id}`, { state: { from: backFrom } })

// After
navigate(`${paths.trimming.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } })
```

```tsx
// Before (line 158)
navigate(`/hospitalization/new?petId=${pet.id}`, { state: { from: backFrom } })

// After
navigate(`${paths.hospitalization.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } })
```

```tsx
// Before (line 164)
navigate(`/accounting/new?petId=${pet.id}`, { state: { from: backFrom } })

// After
navigate(`${paths.accounting.new.getHref()}?petId=${pet.id}`, { state: { from: backFrom } })
```

## 適用ルール

- `config/paths.ts` でURL管理: ハードコードされた URL パス文字列は禁止

## ステータス

✅ 修正済み
