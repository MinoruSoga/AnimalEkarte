# BUG-293: master/routes/CompanySettings.tsx — デッドルート

## 概要

`frontend/src/features/master/routes/CompanySettings.tsx` は `app/router.tsx` に登録されておらず、`MasterSettingsIndex` からもリンクされていない完全なデッドルートである。

## 調査結果

```bash
# router.tsx への登録: 0件
grep -n "CompanySettings" src/app/router.tsx  # → 該当なし

# MasterSettingsIndex からのリンク: 0件
grep -rn "CompanySettings\|company-settings\|company_settings" src/features/master/routes/MasterSettingsIndex.tsx  # → 該当なし
```

- `master/index.ts` には `export { CompanySettings } from "./routes/CompanySettings"` が存在するが外部からの使用は0件
- 病院設定機能は `features/hospital-settings/` に移行済みと推測される

## 修正

1. `frontend/src/features/master/routes/CompanySettings.tsx` を削除
2. `frontend/src/features/master/index.ts` から `CompanySettings` のexportを削除

## ステータス

- [x] ドキュメント作成
- [x] 実装完了（ファイル削除 + index.ts修正済み）
