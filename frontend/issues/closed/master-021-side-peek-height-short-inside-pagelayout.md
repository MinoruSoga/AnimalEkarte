# [master] マスタページの SidePeek: PageLayout 内配置により `h-full` が auto-height に解決され高さが不足する

## 優先度
中

## 種別
UIバグ / レイアウト

## 対象ファイル
- `frontend/src/features/master/routes/HospitalizationSettings.tsx`
- `frontend/src/features/master/routes/DiagnosisSettings.tsx`
- `frontend/src/features/master/routes/TrimmingSettings.tsx`
- `frontend/src/features/master/routes/CageSettings.tsx`
- `frontend/src/features/master/routes/ServiceTypeSettings.tsx`
- `frontend/src/features/master/routes/StaffSettings.tsx`
- `frontend/src/features/master/routes/TreatmentPlanMaster.tsx`

---

## 問題

上記マスタページで SidePeek を開くと、パネルがビューポート全体に広がらず、**コンテンツ高さ（約 507px）にしかならない**。ビューポート下部（約 456px）が空白になる。

### 実測値（ビューポート 963px）

| ページ | SidePeekPanel.height | top | 期待値 |
|--------|---------------------|-----|--------|
| **MedicineSettings** | **963px** ✅ | y=0 | 963px |
| **HospitalizationSettings** | **507px** ❌ | y=73 | 963px |

---

## 原因

### `h-full` の解決チェーン比較

**MedicineSettings（正しい実装）:**

```
main.flex-1.flex-col.overflow-y-auto  → height: 963px (flex-1 で決定)
  └─ <div class="flex h-full">        → h-full = 963px ✅
       ├─ <div class="flex-1 min-w-0">
       │    └─ <PageLayout>...</PageLayout>
       └─ <MedicineSidePanel>          ← PageLayout の外側
            └─ motion.div              → 963px (flex stretch)
                 └─ SidePeekPanel.h-full = 963px ✅
```

**HospitalizationSettings ほか（問題のある実装）:**

```
main.flex-1.flex-col.overflow-y-auto  → height: 963px
  └─ <PageLayout>
       └─ <div class="flex-1 overflow-y-auto">  → height: 910px
            └─ <div class="max-w-full px-3 py-5">  → height: auto ← ここが問題
                 └─ <div class="flex h-full">       → h-full = auto (親が auto のため)
                      └─ SidePeekPanel.h-full = auto = コンテンツ高さ ❌
```

`max-w-full px-3 py-5` ラッパー（`PageLayout` の content wrapper）が `height: auto` であるため、その中の `flex h-full` の `h-full` は **明示的な高さを持つ祖先が存在せず auto に解決**される。結果として `SidePeekPanel` がコンテンツをラップするだけの浮動カードになる。

### 対象コード例（HospitalizationSettings.tsx）

```tsx
// ❌ 現状: SidePeek が PageLayout 内の auto-height 親チェーンに閉じ込められている
<PageLayout maxWidth="max-w-full" ...>
  <div className="flex h-full">           {/* h-full が auto に解決される */}
    <div className="flex flex-col ...">   {/* テーブル */}
      ...
    </div>
    {editTarget !== null ? (
      <HospitalizationSidePanel />        {/* SidePeekPanel.h-full が auto になる */}
    ) : null}
  </div>
</PageLayout>
```

---

## 修正方針

**MedicineSettings のパターンに統一する。**
SidePeek を `PageLayout` の外側に移動し、`flex h-full` の直接 children として配置する。

```tsx
// ✅ 修正後: SidePeek を PageLayout 外に出す（MedicineSettings パターン）
<div className="flex h-full">
  <div className="flex-1 min-w-0">
    <PageLayout maxWidth="max-w-full" ...>
      <div className="flex flex-col gap-4">
        <SearchFilterBar ... />
        <DataTable ... />
      </div>
    </PageLayout>
  </div>
  {editTarget !== null ? (
    <HospitalizationSidePanel />    {/* PageLayout 外 → h-full = 963px ✅ */}
  ) : null}
</div>
```

この構造では：
- `div.flex.h-full` が main の `flex-1` により 963px に解決される
- SidePeek は直接 flex child として stretch → 963px
- `SidePeekPanel.h-full` = 963px ✅
- テーブル（PageLayout 内）は独自スクロール、SidePeek はスクロールに連動しない（意図した挙動）

---

## 影響範囲

- 上記7ページすべてで同一修正が必要
- `SidePeekPanel`・`MasterSidePanel` 自体の変更は不要
- テーブルスクロール挙動も変わらない（PageLayout 内の `overflow-y-auto` はそのまま機能する）

---

## 参考実装

`MedicineSettings.tsx` — SidePeek を PageLayout 外に配置した唯一の正しい実装。
