# BUG-235: `C.textWhite` トークン未定義 — コンポーネント内で `text-white` を 37 箇所直接使用

## 概要

`src/lib/design-tokens.ts` に `C.textWhite` トークンが定義されていないため、白色テキストが必要な箇所（accent ボタン・danger バッジ・カレンダー選択日等）でコンポーネントファイルが `text-white` を直接記述している。`STYLE.confirmPrimary` など design-tokens.ts 内の STYLE 定数自身も `text-white` を直接埋め込んでいる状態であり、設計トークンシステムの一貫性に欠ける。

## 再現手順

（ランタイム・ビジュアルへの影響は現時点でなし。将来的なテーマ変更時に一括対応できない）

1. `frontend/src/` 配下で `text-white` を検索する
2. **結果**: コンポーネントファイルに 37 箇所の直接記述が存在する

## 期待する動作

- `C.textWhite = "text-white"` を design-tokens.ts に追加し、全コンポーネントが `${C.textWhite}` 経由で参照する

## 現状コード

### 代表的な違反パターン（37 件中）

```tsx
// ❌ AccentButton パターン（最多）— ExaminationFilter.tsx:39 など
className={`${C.bgAccent} ${C.bgAccentHover} text-white gap-2 h-10 text-sm`}

// ❌ DangerButton パターン — DischargeAlertDialog.tsx:53 など
className={`${C.bgDanger} ${C.hoverBgDanger90} text-white`}

// ❌ アイコン/バッジパターン — PatientInfoCard.tsx:92
<span className={`... ${C.bgDanger} text-white ...`}>

// ❌ カレンダー選択状態 — NotionDatePicker.tsx:138,147,587
className={`... text-white hover:text-white focus:text-white`}

// ❌ auth ページボタン — LoginForm.tsx:221
className={`... ${C.bgBrand} ${C.hoverBgBrand} ... text-white`}
```

### 全 37 インスタンスのファイルリスト

```
features/auth/components/LoginForm.tsx:156,221
features/auth/routes/ForgotPasswordPage.tsx:50,95
features/auth/routes/ResetPasswordPage.tsx:72,95,161
features/estimates/routes/EstimateForm.tsx:301
features/reservations/components/ReservationDetailModal.tsx:226
features/reservations/components/MonthView.tsx:79
features/reception/components/ReceptionDetailModal.tsx:223
features/medical-records/components/ExaminationFilter.tsx:39
features/medical-records/components/ImageGalleryFilter.tsx:72,127
features/medical-records/components/VaccinationHistory.tsx:138
features/medical-records/components/BillingReviewSection/ReturnReasonDialog.tsx:43
features/medical-records/components/VitalsTab/VitalsGraph.tsx:132
features/examinations/routes/ExaminationForm.tsx:182
features/hospitalization/components/CarePlan/CarePlanDialog.tsx:194
features/hospitalization/components/DischargeAlertDialog.tsx:53
features/hospitalization/routes/HospitalizationForm.tsx:174
features/vaccinations/routes/VaccinationForm.tsx:174
components/shared/PetSelection/PetSelectionResultsTable.tsx:82
components/shared/PetSelection/PetSelectionSearchForm.tsx:71
components/shared/Form/PrimaryButton.tsx:7
components/shared/PatientInfoCard/PatientInfoCard.tsx:92
components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx:80
components/shared/MasterSelectModal/MasterSelectModal.tsx:120
components/shared/ReservationFormModal/ReservationFormModal.tsx:41,319
components/shared/ReservationFormModal/PatientSelectionTable.tsx:124,203
components/shared/ConfirmDialog/ConfirmDialog.tsx:69,70
components/shared/NotionDatePicker/NotionDatePicker.tsx:138,147,587
```

### 比較: design-tokens.ts 内 STYLE 定数でも同問題が存在

```typescript
// src/lib/design-tokens.ts:748 — STYLE.confirmPrimary も text-white を直接埋め込んでいる
confirmPrimary: `${C.bgAccent} ${C.bgAccentHover} text-white h-11 px-4 ...`
```

## 修正方針

### ステップ 1: `C.textWhite` トークンを追加

```typescript
// src/lib/design-tokens.ts — C オブジェクトに追加
textWhite: "text-white",
```

### ステップ 2: design-tokens.ts 内の STYLE 定数を更新（任意）

```typescript
// Before
confirmPrimary: `${C.bgAccent} ${C.bgAccentHover} text-white h-11 ...`

// After
confirmPrimary: `${C.bgAccent} ${C.bgAccentHover} ${C.textWhite} h-11 ...`
```

### ステップ 3: 全 37 インスタンスを置換

```tsx
// Before
className={`${C.bgAccent} ${C.bgAccentHover} text-white`}

// After
className={`${C.bgAccent} ${C.bgAccentHover} ${C.textWhite}`}
```

`hover:text-white` および `focus:text-white` については個別のトークン定義も検討（`C.hoverTextWhite`, `C.focusTextWhite`）。

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず **`C`**, **`STYLE`** 定数を使用（`#37352F`等ハードコード禁止）

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数を使用する。

`text-white` は Tailwind の組み込みカラークラスであり、設計トークンを通じて管理されるべき。`C.groupHoverTextWhite = "group-hover:text-white"` は存在するが、ベースの `text-white` に対応するトークンが欠落している。

### プロジェクト内参照実装
- `src/lib/design-tokens.ts:443` — `C.groupHoverTextWhite = "group-hover:text-white"` （部分的なパターン定義の例）

## 優先度
**Low** — 現時点で視覚的問題なし。テーマ変更・ダークモード対応時に一括修正できるよう整備が必要。`C.textWhite` トークン追加のみで大半が解決できる。

## 関連チケット
- BUG-228, BUG-233, BUG-234: 同種のデザイントークン直接参照違反

## 関連ファイル
- `frontend/src/lib/design-tokens.ts` — トークン追加対象
- 上記 37 ファイル（コンポーネント側修正対象）
