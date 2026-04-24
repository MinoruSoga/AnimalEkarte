# FE-240: ハードコードされたルートパス — config/paths.ts 未使用

## 概要

複数ファイルで `navigate("/xxx")` や `<Link to="/xxx">` に直接文字列パスを使用している。
プロジェクト規約「ハードコードされた URL パス文字列は禁止。`paths.xxx.getHref()` を使用する」に違反。
ルート変更時に修正漏れが発生するリスクがある。

## 違反箇所一覧

### `frontend/src/features/trimming/routes/TrimmingForm.tsx:500`
```ts
// Before
navigate(fromPath ?? "/trimming");
// After
navigate(fromPath ?? paths.trimming.list.getHref());
```

### `frontend/src/features/trimming/routes/TrimmingList.tsx:235`
```ts
// Before
navigate(`/trimming/${id}`, { state: { from: "/trimming" } });
// After
navigate(paths.trimming.detail.getHref(id), { state: { from: paths.trimming.list.getHref() } });
```

### `frontend/src/features/reservations/hooks/use-reservation-management.ts:313,325`
```ts
// Before (2箇所)
{ state: { from: "/reservations" } }
// After
{ state: { from: paths.reservations.list.getHref() } }
```

### `frontend/src/features/medical-records/routes/MedicalRecords.tsx:162-163`
```ts
// Before
navigate(recordId ? `/medical-records/${recordId}` : "/medical-records/select-pet", { state: { from: "/medical-records" } })
// After
navigate(recordId ? paths.medicalRecords.detail.getHref(recordId) : paths.medicalRecords.selectPet.getHref(), { state: { from: paths.medicalRecords.list.getHref() } })
```

### `frontend/src/features/reception/components/AppointmentCard.tsx:82,86,94,101`
```ts
// Before（4箇所）
navigate("/trimming/new?petId=...")
navigate("/medical-records/new?petId=...")
navigate("/accounting/new?petId=...")
navigate("/hospitalization/new?petId=...")
// After: paths.trimming.new.getHref(), paths.medicalRecords.new.getHref() 等を使用
```

### `frontend/src/features/reception/components/ReceptionDetailModal.tsx:336,340,344`
```ts
// Before（3箇所）: trimming, hospitalization, accounting への直接パス
// After: paths.xxx.getHref() を使用
```

### `frontend/src/features/hospitalization/hooks/use-hospitalization-list.ts:57`
```ts
// Before
navigate(id ? `/hospitalization/${id}` : "/hospitalization/select-pet")
// After
navigate(id ? paths.hospitalization.detail.getHref(id) : paths.hospitalization.selectPet.getHref())
```

### `frontend/src/features/estimates/routes/EstimateList.tsx:243`
```ts
// Before
navigate("/estimates/new")
// After
navigate(paths.estimates.new.getHref())
```

### `frontend/src/features/auth/routes/ResetPasswordPage.tsx:57`
```ts
// Before
navigate("/login")
// After
navigate(paths.auth.login.getHref())
```

## 修正手順

1. `frontend/src/config/paths.ts` を参照し、対象パスの `getHref()` 関数が存在するか確認
2. 存在しない場合は `paths.ts` に追加してから参照を変更
3. 上記ファイルの直接文字列を `paths.xxx.getHref()` に置換

## 準拠すべきプロジェクト規約

### `frontend/CLAUDE.md` — 禁止事項
> ハードコードされた URL パス文字列は禁止。
> `config/paths.ts` でURL管理。`paths.owners.list.getHref()` 等を使用する。

## 優先度
**Medium** — 機能的障害なし。ルート変更時の修正漏れリスクが高い。
特に `AppointmentCard.tsx`（受付フロー）・`MedicalRecords.tsx`（カルテ）は影響度が高い。

## 関連ファイル
- `frontend/src/config/paths.ts` — 正規パス定義
- `frontend/src/features/trimming/routes/TrimmingForm.tsx`
- `frontend/src/features/trimming/routes/TrimmingList.tsx`
- `frontend/src/features/reservations/hooks/use-reservation-management.ts`
- `frontend/src/features/medical-records/routes/MedicalRecords.tsx`
- `frontend/src/features/reception/components/AppointmentCard.tsx`
- `frontend/src/features/reception/components/ReceptionDetailModal.tsx`
- `frontend/src/features/hospitalization/hooks/use-hospitalization-list.ts`
- `frontend/src/features/estimates/routes/EstimateList.tsx`
- `frontend/src/features/auth/routes/ResetPasswordPage.tsx`
