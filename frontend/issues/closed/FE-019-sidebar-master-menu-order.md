# FE-019: Sidebar マスタ設定メニュー順序を /settings ページに合わせる

**Status**: Open
**Priority**: Low
**Affects**: shared Layout — Sidebar
**Date Created**: 2026-03-17
**Related**: TASK-004

## Summary

サイドバーの「マスタ設定」ドロップダウンメニューの項目順序・グループ構成を `/settings` ページ（MasterSettingsIndex）のセクション構成に合わせる。不足している項目（医院マスタ、動物種類、職種）を追加し、カルテグループ内の順序を修正する。

## 現状のコード

### Sidebar — マスタ設定メニュー（現在の順序）

```typescript
// frontend/src/components/shared/Layout/Sidebar.tsx:185-207
{
  icon: <Settings className="size-[18px]" />,
  label: "マスタ設定",
  path: paths.settings.getHref(),
  subItems: [
    { icon: <Activity />, label: "予約区分マスタ", path: paths.settings.serviceType.getHref() },
    {
      icon: <FileText />,
      label: "カルテ",
      subItems: [
        { icon: <ClipboardList />, label: "治療プランマスタ", path: paths.settings.treatmentItems.getHref() },
        { icon: <Pill />, label: "薬剤マスタ", path: paths.settings.medicine.getHref() },
        { icon: <Clipboard />, label: "診断病名マスタ", path: paths.settings.diagnosis.getHref() },
        { icon: <ClipboardCheck />, label: "問診マスタ", path: paths.settings.inquiryTemplates.getHref() },
      ],
    },
    { icon: <Bed />, label: "入院マスタ", path: paths.settings.hospitalization.getHref() },
    { icon: <Building2 />, label: "ケージマスタ", path: paths.settings.cage.getHref() },
    { icon: <Scissors />, label: "トリミングマスタ", path: paths.settings.trimming.getHref() },
    { icon: <Users />, label: "スタッフマスタ", path: paths.settings.staff.getHref() },
    { icon: <ShieldCheck />, label: "保険マスタ", path: paths.settings.insurance.getHref() },
  ],
}
```

### /settings ページ — セクション定義（目標の順序）

```typescript
// frontend/src/features/master/routes/MasterSettingsIndex.tsx:90-100
const MASTER_SECTIONS: SectionDef[] = [
  { title: "基本設定", keys: ["clinic", "animal_species"] },
  { title: "カルテ", keys: ["treatmentItems", "diagnosisGroup", "inquiry_template", "medicine"] },
  { title: "診療関連マスタ", keys: ["serviceType"] },
  { title: "入院・ケージ管理", keys: ["hospitalization", "cage"] },
  { title: "トリミング関連", keys: ["trimmingGroup"] },
  { title: "スタッフ・保険", keys: ["staff", "job_title", "insurance"] },
];
```

## 必要な変更

### Sidebar.tsx — subItems 配列を以下の順序に変更

```typescript
// frontend/src/components/shared/Layout/Sidebar.tsx:185-207
// Before → After

subItems: [
  // 1. 基本設定
  { icon: <Building2 />, label: "医院マスタ", path: paths.settings.clinic.getHref() },
  { icon: <PawPrint />, label: "動物種類マスタ", path: paths.settings.animalSpecies.getHref() },

  // 2. カルテ（グループ）
  {
    icon: <FileText />,
    label: "カルテ",
    subItems: [
      { icon: <ClipboardList />, label: "診療項目マスタ", path: paths.settings.treatmentItems.getHref() },
      { icon: <Clipboard />, label: "診断病名マスタ", path: paths.settings.diagnosis.getHref() },
      { icon: <ClipboardCheck />, label: "問診マスタ", path: paths.settings.inquiryTemplates.getHref() },
      { icon: <Pill />, label: "薬剤マスタ", path: paths.settings.medicine.getHref() },
    ],
  },

  // 3. 診療関連
  { icon: <Activity />, label: "予約区分マスタ", path: paths.settings.serviceType.getHref() },

  // 4. 入院・ケージ管理
  { icon: <Bed />, label: "入院マスタ", path: paths.settings.hospitalization.getHref() },
  { icon: <Building2 />, label: "ケージマスタ", path: paths.settings.cage.getHref() },

  // 5. トリミング
  { icon: <Scissors />, label: "トリミングマスタ", path: paths.settings.trimming.getHref() },

  // 6. スタッフ・保険
  { icon: <Users />, label: "スタッフマスタ", path: paths.settings.staff.getHref() },
  { icon: <Briefcase />, label: "職種マスタ", path: paths.settings.jobTitle.getHref() },
  { icon: <ShieldCheck />, label: "保険マスタ", path: paths.settings.insurance.getHref() },
],
```

### 変更点まとめ

| 変更 | 詳細 |
|------|------|
| **追加** | 医院マスタ（先頭） |
| **追加** | 動物種類マスタ（基本設定） |
| **追加** | 職種マスタ（スタッフ・保険） |
| **順序変更** | 予約区分マスタ: 先頭 → カルテの後 |
| **順序変更** | カルテ内: 治療プラン→薬剤→診断→問診 → 診療項目→診断→問診→薬剤 |
| **ラベル変更** | 「治療プランマスタ」→「診療項目マスタ」（/settings の表示名に合わせる） |

### paths.ts の確認

以下のパスが `config/paths.ts` に定義されているか確認が必要:
- `paths.settings.clinic` — 医院マスタ
- `paths.settings.animalSpecies` — 動物種類マスタ
- `paths.settings.jobTitle` — 職種マスタ

存在しない場合は paths.ts にも追加する。

### import 追加

```typescript
// Sidebar.tsx の Lucide import に追加が必要な可能性:
import { PawPrint, Briefcase } from "lucide-react";
```

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）

## 依存関係

- Backend 変更不要
- `paths.settings.clinic`, `paths.settings.animalSpecies`, `paths.settings.jobTitle` のパス定義が必要（なければ追加）

## 完了条件

- [ ] Sidebar のマスタ設定メニューが /settings ページと同じ順序
- [ ] 医院マスタ、動物種類マスタ、職種マスタがメニューに追加
- [ ] カルテ内順序: 診療項目 → 診断 → 問診 → 薬剤
- [ ] 予約区分マスタがカルテの後に移動
- [ ] 全メニュー項目のリンクが正しく動作
- [ ] 型エラーなし（`docker compose exec frontend npm run build` パス）
- [ ] ESLint エラーなし（`docker compose exec frontend npm run lint` パス）
