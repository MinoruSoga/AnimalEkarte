# BUG-302: VaccinationList — handleEdit でハードコードされた URL パス

## 概要

`config/paths.ts` 規約違反。`VaccinationList.tsx` の `handleEdit` で `navigate(\`/vaccinations/${id}\`)` とリテラル文字列が使われている。同ファイル内の `handleCreate` では `paths.vaccinations.selectPet.getHref()` を正しく使用しており、一貫性がない。

## 影響ファイル

- `frontend/src/features/vaccinations/routes/VaccinationList.tsx`

## 違反箇所（修正前）

```tsx
// line 141
const handleEdit = useCallback((id: string) => {
  navigate(`/vaccinations/${id}`);  // ← VIOLATION
}, [navigate]);
```

## 修正内容

```tsx
const handleEdit = useCallback((id: string) => {
  navigate(paths.vaccinations.detail.getHref(id));
}, [navigate]);
```

## 適用ルール

- `config/paths.ts` でURL管理: ハードコードされた URL パス文字列は禁止

## ステータス

✅ 修正済み
