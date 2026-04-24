# FE-066: マスタページ タブUI統一（Radix Primitives + URL同期）

**Status**: Open
**Priority**: Medium
**Affects**: `features/master/routes/TrimmingSettings.tsx`, `features/master/routes/DiagnosisSettings.tsx`
**Date Created**: 2026-03-18
**Related**: TASK-015

## Summary

TrimmingSettings と DiagnosisSettings のタブ実装を、TreatmentPlanMaster と同じ Radix UI Primitives + useSearchParams パターンに統一する。

## 現状のコード

### 参照実装: TreatmentPlanMaster

```typescript
// frontend/src/features/master/routes/TreatmentPlanMaster.tsx:20
import * as TabsPrimitive from "@radix-ui/react-tabs";

// :453-454 — URL同期
const [searchParams, setSearchParams] = useSearchParams();
const activeTab = searchParams.get("tab") ?? "consultation";

// :458-462 — タブ変更ハンドラ（副作用クリーンアップ付き）
const handleTabChange = useCallback((tab: string) => {
  setSearchParams({ tab });
  setEditTarget(null);
  setPendingDelete(null);
}, [setSearchParams]);

// :733-751 — Radix Primitives タブ
<TabsPrimitive.Root value={activeTab} onValueChange={handleTabChange} className="flex flex-col gap-4">
  <TabsPrimitive.List className={`flex h-9 border-b ${C.borderLight} gap-0`}>
    {TABS.map((tab) => (
      <TabsPrimitive.Trigger
        key={tab.value}
        value={tab.value}
        className={`h-9 border-b-2 border-b-transparent px-4 text-sm ${C.text60} outline-none transition-colors cursor-pointer
          data-[state=active]:border-b-[#37352F] data-[state=active]:text-[#37352F] data-[state=active]:font-medium`}
      >
        {tab.label}
      </TabsPrimitive.Trigger>
    ))}
  </TabsPrimitive.List>
  {TABS.map((tab) => (
    <TabsPrimitive.Content key={tab.value} value={tab.value} className="mt-4">
      {/* tab content */}
    </TabsPrimitive.Content>
  ))}
</TabsPrimitive.Root>
```

### 修正対象1: TrimmingSettings

```typescript
// frontend/src/features/master/routes/TrimmingSettings.tsx:20
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
// ❌ shadcn/ui ラッパーを使用

// タブ状態管理（URL同期なし）
// useState でローカル管理 — ページリロードでタブ状態が失われる

// :627-655 — shadcn/ui Tabs
<div className="flex flex-col gap-4">
  <Tabs value={activeTab} onValueChange={handleTabChange}>
    <TabsList
      className={`h-9 bg-transparent border-b ${C.borderLight} rounded-none w-full justify-start gap-0 p-0`}
    >
      {TABS.map((tab) => (
        <TabsTrigger
          key={tab.value}
          value={tab.value}
          className={`h-9 rounded-none border-b-2 border-transparent px-4 text-sm ${C.text60}
            data-[state=active]:border-[#37352F] data-[state=active]:${C.text}
            data-[state=active]:shadow-none data-[state=active]:bg-transparent`}
        >
          {tab.label}
        </TabsTrigger>
      ))}
    </TabsList>
    <TabsContent value="course" className="mt-4">...</TabsContent>
    <TabsContent value="option" className="mt-4">...</TabsContent>
  </Tabs>
</div>
```

### 修正対象2: DiagnosisSettings

```typescript
// frontend/src/features/master/routes/DiagnosisSettings.tsx:35
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
// ❌ shadcn/ui ラッパーを使用（ただしスタイルは TreatmentPlan と同じ文字列）

// :545-576 — shadcn/ui Tabs（URL同期あり）
<Tabs value={activeTab} onValueChange={handleTabChange} className="flex flex-col gap-4">
  <TabsList className={`flex h-9 border-b ${C.borderLight} gap-0`}>
    {TABS.map((tab) => (
      <TabsTrigger
        key={tab.value}
        value={tab.value}
        className={`h-9 border-b-2 border-b-transparent px-4 text-sm ${C.text60} outline-none transition-colors cursor-pointer
          data-[state=active]:border-b-[#37352F] data-[state=active]:text-[#37352F] data-[state=active]:font-medium`}
      >
        {tab.label}
      </TabsTrigger>
    ))}
  </TabsList>
  <TabsContent value="diagnosis_category" className="mt-4">...</TabsContent>
  <TabsContent value="diagnosis_name" className="mt-4">...</TabsContent>
</Tabs>
```

## 必要な変更

### 1. TrimmingSettings — import 変更

```typescript
// Before
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

// After
import * as TabsPrimitive from "@radix-ui/react-tabs";
```

### 2. TrimmingSettings — URL同期追加

```typescript
// Before（useState ローカル管理）
const [activeTab, setActiveTab] = useState("course");

// After（useSearchParams で URL同期）
import { useSearchParams } from "react-router-dom";

const [searchParams, setSearchParams] = useSearchParams();
const activeTab = searchParams.get("tab") ?? "course";

const handleTabChange = useCallback((tab: string) => {
  setSearchParams({ tab });
  setEditTarget(null);
  setPendingDelete(null);
}, [setSearchParams]);
```

### 3. TrimmingSettings — JSX 置換

```typescript
// Before
<Tabs value={activeTab} onValueChange={handleTabChange}>
  <TabsList className={...}>
    <TabsTrigger ...>
  </TabsList>
  <TabsContent ...>
</Tabs>

// After（TreatmentPlanMaster と完全同一パターン）
<TabsPrimitive.Root value={activeTab} onValueChange={handleTabChange} className="flex flex-col gap-4">
  <TabsPrimitive.List className={`flex h-9 border-b ${C.borderLight} gap-0`}>
    {TABS.map((tab) => (
      <TabsPrimitive.Trigger
        key={tab.value}
        value={tab.value}
        className={`h-9 border-b-2 border-b-transparent px-4 text-sm ${C.text60} outline-none transition-colors cursor-pointer
          data-[state=active]:border-b-[#37352F] data-[state=active]:text-[#37352F] data-[state=active]:font-medium`}
      >
        {tab.label}
      </TabsPrimitive.Trigger>
    ))}
  </TabsPrimitive.List>
  <TabsPrimitive.Content value="course" className="mt-4">
    <TrimmingCourseTab ... />
  </TabsPrimitive.Content>
  <TabsPrimitive.Content value="option" className="mt-4">
    <TrimmingOptionTab ... />
  </TabsPrimitive.Content>
</TabsPrimitive.Root>
```

### 4. DiagnosisSettings — import 変更

```typescript
// Before
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";

// After
import * as TabsPrimitive from "@radix-ui/react-tabs";
```

### 5. DiagnosisSettings — JSX 置換

同じパターンで `Tabs` → `TabsPrimitive.Root`、`TabsList` → `TabsPrimitive.List`、`TabsTrigger` → `TabsPrimitive.Trigger`、`TabsContent` → `TabsPrimitive.Content` に置換。

URL同期（useSearchParams）は既に実装済みのため変更不要。

### 6. 両ページ — handleTabChange の副作用クリーンアップ確認

TreatmentPlanMaster と同様に、タブ切り替え時に以下をリセットする:
- `setEditTarget(null)` — サイドパネル閉じ
- `setPendingDelete(null)` — 確認ダイアログ閉じ

各ページの state 変数名に合わせて適用する。

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] barrel index 経由 import なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] shadcn/ui `Tabs` import が除去されている
- [ ] `@radix-ui/react-tabs` Primitives を直接使用

## 依存関係

- 依存なし（独立して着手可能）
- `@radix-ui/react-tabs` は shadcn/ui の依存として既にインストール済み

## 完了条件

- [ ] TrimmingSettings が `TabsPrimitive` で実装されている
- [ ] TrimmingSettings のタブ状態が URL パラメータに同期されている（`?tab=course`）
- [ ] DiagnosisSettings が `TabsPrimitive` で実装されている
- [ ] DiagnosisSettings の既存 URL 同期が維持されている
- [ ] 3ページのタブスタイル（className）が完全一致している
- [ ] タブ切り替え時にサイドパネル・確認ダイアログがリセットされる
- [ ] `pnpm build` パス
- [ ] `pnpm lint` パス
- [ ] トリミングマスタ画面でコース/オプション タブ切り替えが正常動作
- [ ] 診断病名マスタ画面でカテゴリ/病名 タブ切り替えが正常動作
