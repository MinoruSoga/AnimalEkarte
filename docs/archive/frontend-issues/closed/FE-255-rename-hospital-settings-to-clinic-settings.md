# FE-255: hospital-settings を clinic-settings に統一

**Status**: Open  
**Priority**: Medium  
**Type**: Refactor  
**Date Created**: 2026-04-19

## 背景

DB・バックエンドは一貫して `clinic` / `clinic_id` を使用しているが、
フロントエンドのみ `hospital-settings` という語を使用しており、同一概念を2つの語で呼び分けている状態。

## 現状の不統一

| 場所 | 現在の語 |
|------|---------|
| DB テーブル | `clinics` |
| BE モデル | `Clinic` struct |
| BE サービス | `ClinicService` |
| BE ルート | `/v1/clinics` |
| **FE feature ディレクトリ** | `hospital-settings` ← 不統一 |
| **FE ルートパス** | `/hospital-settings` ← 不統一 |
| FE コンポーネント名 | `ClinicMasterSettings`（既に clinic を使用） |
| FE API hook | `useGetClinics` など（既に clinic を使用） |

## 対応方針

`clinic-settings` に統一する。ユーザー可視 URL が変わるためリダイレクトが必要。

## 変更対象ファイル

### ディレクトリ・ファイル名変更
- `frontend/src/features/hospital-settings/` → `frontend/src/features/clinic-settings/`

### `paths.ts` のキー変更
```typescript
// Before
hospitalSettings: {
  path: "/hospital-settings",
  getHref: () => "/hospital-settings",
}

// After
clinicSettings: {
  path: "/clinic-settings",
  getHref: () => "/clinic-settings",
}
```

### Router エントリ変更
- `src/app/router.tsx` の lazy import パスと path プロパティを更新

### サイドバーリンク変更
- `Sidebar.tsx` の `paths.hospitalSettings` 参照を `paths.clinicSettings` に変更

### リダイレクト追加
```typescript
// router.tsx に追加
{
  path: "/hospital-settings",
  loader: () => redirect("/clinic-settings"),
}
```

### 影響確認が必要な箇所
- `src/app/pages/` 配下で `hospitalSettings` を参照している箇所
- `@/features/hospital-settings` を import している全ファイル

## 完了条件

- [ ] `features/clinic-settings/` にリネーム
- [ ] `paths.clinicSettings` に変更（`hospitalSettings` キー削除）
- [ ] `/hospital-settings` → `/clinic-settings` のリダイレクト追加
- [ ] サイドバー・router の参照更新
- [ ] lint / 型チェック / ビルドが通る
