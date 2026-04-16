# BUG-301: AppointmentCard — navigate() にハードコードされた URL パス文字列

## 概要

`config/paths.ts` 規約違反。`AppointmentCard.tsx` 内の `handleKarteClick`・`handleAccountingClick`・`handleHospitalizationClick` の 3 ハンドラで `navigate()` にリテラル文字列パスが渡されている。

## 影響ファイル

- `frontend/src/features/reception/components/AppointmentCard.tsx`

## 違反箇所（修正前）

```tsx
// line 82
navigate(appointment.petId ? `/trimming/new?petId=${appointment.petId}` : "/trimming/new", ...)

// line 86
navigate(appointment.petId ? `/medical-records/new?petId=${appointment.petId}` : "/medical-records/select-pet", ...)

// line 94
navigate(appointment.petId ? `/accounting/new?petId=${appointment.petId}` : "/accounting/new", ...)

// line 101
navigate(appointment.petId ? `/hospitalization/new?petId=${appointment.petId}` : "/hospitalization/new", ...)
```

## 修正内容

`paths.xxx.getHref()` に置き換え。`petId` のクエリパラメータは `getHref()` が引数対応していないため、テンプレートリテラルで補完する形で統一。

```tsx
import { paths } from "@/config/paths";

// trimming
navigate(
  appointment.petId
    ? `${paths.trimming.new.getHref()}?petId=${appointment.petId}`
    : paths.trimming.new.getHref(),
  ...
)

// medical-records
navigate(
  appointment.petId
    ? `${paths.medicalRecords.new.getHref()}?petId=${appointment.petId}`
    : paths.medicalRecords.selectPet.getHref(),
  ...
)

// accounting
navigate(
  appointment.petId
    ? `${paths.accounting.new.getHref()}?petId=${appointment.petId}`
    : paths.accounting.new.getHref(),
  ...
)

// hospitalization
navigate(
  appointment.petId
    ? `${paths.hospitalization.new.getHref()}?petId=${appointment.petId}`
    : paths.hospitalization.new.getHref(),
  ...
)
```

## 適用ルール

- `config/paths.ts` でURL管理: ハードコードされた URL パス文字列は禁止。パスが変更された場合に一か所だけ修正すれば済む型安全な getHref() を使う。

## ステータス

✅ 修正済み
