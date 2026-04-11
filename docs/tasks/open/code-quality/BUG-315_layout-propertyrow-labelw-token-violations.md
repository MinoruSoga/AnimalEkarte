# BUG-315: LAYOUT.propertyRow.labelW トークン未使用 — w-[140px] ハードコード

## 概要
`LAYOUT.propertyRow.labelW = "w-[140px]"` が定義されているにもかかわらず、
Notion スタイルプロパティ行のラベル幅として `"w-[140px]"` が直接ハードコードされている箇所が2件ある。
さらに同値が別の UI 要素（テーブルカラム・SelectTrigger）でも使われており、
これらはトークン化するか意図的なハードコードとして承認するかの判断が必要。

---

## Part 1: 確定違反（2件）— プロパティ行ラベル用途

### `frontend/src/components/shared/SidePeek/PropertyRow.tsx:13`
```tsx
// 現在
<div className={`w-[140px] shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>

// 修正後
<div className={`${LAYOUT.propertyRow.labelW} shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>
```
**根拠**: SidePeek の Notion スタイルプロパティ行ラベル — `LAYOUT.propertyRow.labelW` の設計意図と完全一致。

### `frontend/src/features/hospital-settings/routes/ClinicMasterSettings.tsx:60`
```tsx
// 現在
<div className={`w-[140px] shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>

// 修正後
<div className={`${LAYOUT.propertyRow.labelW} shrink-0 text-sm ${C.text65} select-none truncate flex items-center`}>
```
**根拠**: PropertyRow.tsx と完全に同一パターン。病院設定の Notion スタイル設定行ラベル。

---

## Part 2: 意味的ミスマッチ要確認（5件）— テーブルカラム・SelectTrigger 用途

以下はすべて `w-[140px]` を使用しているが、`LAYOUT.propertyRow.labelW`（プロパティ行ラベル）とは
**異なる UI 文脈**で使用されている。そのまま `LAYOUT.propertyRow.labelW` に置換すると
意味的に誤ったトークン名になる。

| ファイル | 行 | 用途 |
|---------|-----|------|
| `features/estimates/routes/EstimateList.tsx` | 67 | DataTable カラム幅（見積番号列） |
| `features/master/routes/InsuranceSettings.tsx` | 20 | DataTable カラム幅（連絡先列） |
| `features/vaccinations/routes/VaccinationList.tsx` | 199 | DataTable カラム幅 |
| `features/reservations/routes/ReservationManagement.tsx` | 224 | SelectTrigger 幅（フィルター） |
| `features/reservations/routes/ReservationManagement.tsx` | 245 | SelectTrigger 幅（フィルター） |

### 選択肢
**案 A**: `LAYOUT.propertyRow.labelW` をプロパティ行専用とし、上記5件は意図的ハードコードとして承認する
**案 B**: `LAYOUT.column.compact = "w-[140px]"` を新規追加し、テーブルカラム・コントロール用途のトークンとして使用する

---

## 影響範囲

| 対象 | ファイル | 行 | 状態 |
|------|---------|-----|------|
| PropertyRow ラベル | `components/shared/SidePeek/PropertyRow.tsx` | 13 | 要修正（Part 1） |
| 病院設定ラベル | `features/hospital-settings/routes/ClinicMasterSettings.tsx` | 60 | 要修正（Part 1） |
| 見積一覧カラム | `features/estimates/routes/EstimateList.tsx` | 67 | 要確認（Part 2） |
| 保険設定カラム | `features/master/routes/InsuranceSettings.tsx` | 20 | 要確認（Part 2） |
| ワクチン一覧カラム | `features/vaccinations/routes/VaccinationList.tsx` | 199 | 要確認（Part 2） |
| 予約管理 SelectTrigger × 2 | `features/reservations/routes/ReservationManagement.tsx` | 224, 245 | 要確認（Part 2） |

## 修正方針

### Part 1 修正（2件のみ — 要確認不要）

両ファイルに `LAYOUT` を import 追加:
```tsx
import { C, LAYOUT } from "@/lib/design-tokens";
```

`w-[140px]` を `${LAYOUT.propertyRow.labelW}` に置換。

### Part 2 の判断基準
- **案 A を採用する場合**: 5件はそのままで問題なし（意図的ハードコードとして記録）
- **案 B を採用する場合**: `design-tokens.ts` の `LAYOUT` に `column: { compact: "w-[140px]" }` を追加してから5件を置換

## 準拠すべきプロジェクト規約

### `.claude/rules/typescript-react.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`, `LAYOUT`) を使用する。
> **PROHIBITED**: px サイズのハードコードは厳禁。

## 優先度
**Low** — 機能影響なし。Part 1 の2件は単純な置換で対応可能。Part 2 は設計判断が必要。

## 関連チケット
- BUG-313: tableEmptySm 新トークン候補
- BUG-314: モーダルサイズギャップ

## 関連ファイル
- `frontend/src/lib/design-tokens.ts` — LAYOUT.propertyRow.labelW 定義元
- `frontend/src/components/shared/SidePeek/PropertyRow.tsx:13`
- `frontend/src/features/hospital-settings/routes/ClinicMasterSettings.tsx:60`
