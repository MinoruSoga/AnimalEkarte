# TASK-015: マスタページ タブUI統一（Radix Primitives + URL同期）

**作成日**: 2026-03-18
**ステータス**: Open
**依頼元**: ユーザー

---

## 概要

トリミングマスタと診断病名マスタのタブUIを、治療プランマスタと同じ Radix UI Primitives + useSearchParams パターンに統一する。

## 依頼内容（原文）

> トリミングマスタ、診断病名マスタページのタブですが、治療プランマスタと同じようなタブのUIにしてください。

## 仕様確認ログ

確認事項なし。参照実装（TreatmentPlanMaster）のパターンが明確であり、2ページの差異も特定済み。

## 現状の差異

| 項目 | TreatmentPlanMaster（参照） | TrimmingSettings | DiagnosisSettings |
|------|---------------------------|------------------|-------------------|
| import | `@radix-ui/react-tabs` Primitives | shadcn/ui `Tabs` | shadcn/ui `Tabs` |
| URL同期 | ✅ `useSearchParams()` | ❌ `useState` | ✅ `useSearchParams()` |
| コンポーネント | `TabsPrimitive.Root/List/Trigger/Content` | `Tabs/TabsList/TabsTrigger/TabsContent` | `Tabs/TabsList/TabsTrigger/TabsContent` |
| スタイル | Notion風 underline（`data-[state=active]:border-b-[#37352F]`） | shadcn デフォルト + カスタム | TreatmentPlan と同じスタイル文字列 |

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | TrimmingSettings: shadcn → Radix Primitives + URL同期追加 | FE | FE-066 | - | [x] |
| 2 | DiagnosisSettings: shadcn → Radix Primitives（URL同期は既存） | FE | FE-066 | - | [x] |

## 受入条件（Acceptance Criteria）

- [x] AC-1: TrimmingSettings のタブが `@radix-ui/react-tabs` Primitives で実装されている
- [x] AC-2: TrimmingSettings のタブ状態が URL パラメータ `?tab=course` / `?tab=option` に同期されている
- [x] AC-3: DiagnosisSettings のタブが `@radix-ui/react-tabs` Primitives で実装されている
- [x] AC-4: DiagnosisSettings の既存 URL 同期が維持されている
- [x] AC-5: 3ページのタブの見た目（underline スタイル、フォント、色）が同一である
- [x] AC-6: タブ切り替え時にサイドパネル・確認ダイアログがリセットされる

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| タブコンポーネント | Radix UI Primitives 直接使用 | TreatmentPlanMaster の参照実装に統一 | shadcn/ui Tabs ラッパー |
| URL同期 | useSearchParams | ブックマーク・リロード対応 | useState |

## 影響範囲

### Backend
- 変更なし

### Frontend
- `features/master/routes/TrimmingSettings.tsx` — タブ実装置換 + URL同期追加
- `features/master/routes/DiagnosisSettings.tsx` — タブ実装置換（URL同期は維持）

## 参照実装

- `features/master/routes/TreatmentPlanMaster.tsx:733-769` — Radix Primitives タブ実装
- `features/master/routes/TreatmentPlanMaster.tsx:453-462` — useSearchParams + handleTabChange

## リスク・懸念事項

| リスク | 影響度 | 対策 |
|--------|--------|------|
| shadcn/ui Tabs 固有の props に依存するコードがある | 低 | Radix Primitives は shadcn/ui の内部実装なので API 互換 |

## 未解決事項

- なし

## 実装順序

1. FE-066: 2ページ同時にタブ実装を置換

## 関連イシュー

- FE-066: [マスタページ タブUI統一](../../frontend/issues/open/FE-066-master-tab-ui-unification.md)
