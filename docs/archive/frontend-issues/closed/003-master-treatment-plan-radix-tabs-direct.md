---
status: closed
closed_at: 2026-03-16
reason: 仕様誤起票。master-pages.md §2 では「@radix-ui/react-tabs（TabsPrimitive 直接）」と明示されており、実装は仕様通り正しい。
---

# [master] TreatmentPlanMaster: @radix-ui/react-tabs を直接インポートしている（仕様違反）

## 優先度
高

## 種別
仕様違反

## 対象ファイル
`frontend/src/features/master/routes/TreatmentPlanMaster.tsx`

## 問題

`master-pages.md` の仕様では診療項目マスタのタブコンポーネントは `@radix-ui/react-tabs`（TabsPrimitive 直接）と記載されているが、
他のマスタページ（DiagnosisSettings、TrimmingSettings）は `@/components/ui/tabs`（shadcn/ui）を使っている。
`master-pages.md` の説明コメント「same as DiagnosisSettings」は誤りであり、`TreatmentPlanMaster.tsx` のみ Radix 直接インポートというアーキテクチャ不一致が生じている。

```tsx
// 現状 (TreatmentPlanMaster.tsx)
import * as TabsPrimitive from "@radix-ui/react-tabs";  // ← Radix 直接

// 期待
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";  // shadcn/ui
```

## 影響

- UI スタイルが他のマスタページ（DiagnosisSettings）と異なる
- shadcn/ui の Tabs に内包されているアクセシビリティ・テーマ設定が適用されない

## 修正方針

`@radix-ui/react-tabs` の直接インポートを `@/components/ui/tabs` に置き換え、
`TabsPrimitive.Root` → `Tabs`、`TabsPrimitive.List` → `TabsList` 等に変更する。

## 関連
`master-pages.md` の仕様記述も「`@/components/ui/tabs`（shadcn/ui）」に修正すること。
