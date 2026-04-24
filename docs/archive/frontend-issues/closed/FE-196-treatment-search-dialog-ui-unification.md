# FE-196: TreatmentSearchDialog を MasterSelectModal 準拠の UI に統一

## 概要

カルテ「診察/治療プラン」タブの治療追加モーダル（`TreatmentSearchDialog`）が、
アプリ内の他マスタ選択モーダル（`MasterSelectModal`）と全く異なる UI パターンを使用している。
`CommandDialog`（cmdk コマンドパレット）ベースの実装はリスト選択ユースケースに不適切であり、
視覚的な一貫性が損なわれている。

## 問題点

### 1. CommandDialog のUXミスマッチ

`TreatmentSearchDialog` は `CommandDialog`（cmdk）を使用している。
これはキーボードショートカット型コマンドパレット向けのコンポーネントであり、
「リストをブラウズして選択する」ユースケースには向かない。

```tsx
// 現在 (TreatmentSearchDialog.tsx:204-264) — CommandDialog ベース
<CommandDialog open={open} onOpenChange={onOpenChange} ...>
  <CommandInput placeholder="治療プランを検索... (例: 再診、ワクチン、3001)" />
  <CommandList className="max-h-[500px]">
    <CommandItem ... >...</CommandItem>
  </CommandList>
</CommandDialog>

// 正しい (MasterSelectModal.tsx:57-133) — Dialog ベース
<Dialog open={open} onOpenChange={onOpenChange}>
  <DialogContent className="sm:max-w-[500px]">
    <DialogHeader><DialogTitle>...</DialogTitle></DialogHeader>
    <Input ... />
    <div className="space-y-2 max-h-[400px] overflow-y-auto">
      <div onClick={() => handleSelect(item)} ...>...</div>
    </div>
  </DialogContent>
</Dialog>
```

### 2. 選択状態インジケーターの欠如

`CommandItem` はホバーハイライトのみで、ラジオ/チェックマークによる選択視覚がない。
`MasterSelectModal` には選択状態を明示するラジオ円 + チェックアイコンがある。

```tsx
// 現在 — 選択インジケーターなし
<CommandItem onSelect={() => handleSelect(item)} ...>
  <div className="flex flex-1 items-center justify-between">
    <span>{item.name}</span>
    <span>¥{item.unitPrice.toLocaleString()}</span>
  </div>
</CommandItem>

// 正しい (MasterSelectModal.tsx:96-127)
<div onClick={() => handleSelect(item)} className={`p-3 border rounded-lg ...`}>
  <div className="flex-1"><div className="text-sm">{item.name}</div></div>
  <div className="size-5 rounded-full ${C.bgPrimary} flex items-center justify-center">
    <Check className="text-white" />
  </div>
</div>
```

### 3. カテゴリバッジ選択色の不統一

現在、選択中カテゴリバッジが `C.bgPrimary`（黒）を使用している。
アプリ全体でのプライマリアクションは `C.bgAccent`（青）に統一されている。

```tsx
// 現在 (TreatmentSearchDialog.tsx:79) — 黒
isSelected
  ? `${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} border-transparent`
  : ...

// 正しい — 青
isSelected
  ? `${C.bgAccent} text-white ${C.bgAccentHover} border-transparent`
  : ...
```

### 4. 意味のない自動採番コードの表示

コード列（`1001`, `2001`, `3001`）はインデックスベースの自動採番であり、
実際の医療コードではない。ユーザーに混乱を与えるため削除すべき。

```tsx
// 現在 (TreatmentSearchDialog.tsx:128-177) — 偽コード生成
items.push({
  id: c.id,
  code: `1${String(idx + 1).padStart(3, "0")}`,  // 偽コード
  ...
});

// プレースホルダーも参照している (TreatmentSearchDialog.tsx:211)
placeholder="治療プランを検索... (例: 再診、ワクチン、3001)"
// → "3001" は偽コードへの参照

// CommandItem のキーにも使用 (TreatmentSearchDialog.tsx:234)
<CommandItem key={item.code} value={`${item.name} ${item.code} ${item.category}`}>
```

### 5. 検索クリアボタンの欠如

`MasterSelectModal` には検索テキスト X クリアボタンがあるが、
`CommandInput` にはない（cmdk の制約）。

## 再現手順

1. `https://stg.noah-karte.com` にログイン
2. 任意のカルテを開く
3. 「診察/治療プラン」タブ → 「治療を追加」ボタンをクリック
4. **結果**: cmdk スタイルのコマンドパレットが開く（他マスタ選択と異なる見た目）

対比:
5. トリミングフォームのコース選択 → `MasterSelectModal` ベースの明確な UI

## 期待する動作

- `Dialog` ベースのモーダル（`MasterSelectModal` と同一構造）
- 検索バー: `Input` + Search アイコン + X クリアボタン
- カテゴリフィルタバッジ: 選択色 `C.bgAccent`（青）
- アイテム行: カードスタイル（border + rounded）+ ラジオ/チェック選択インジケーター
- コード列削除（name と price のみ表示）
- プレースホルダー: "治療プランを検索..."（簡潔に）

## 現状コード

### `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx:204-264`
```tsx
return (
  <CommandDialog
    open={open}
    onOpenChange={onOpenChange}
    title="治療プラン検索"
    description="追加する治療プランを検索・選択してください"
  >
    <CommandInput placeholder="治療プランを検索... (例: 再診、ワクチン、3001)" />
    <CategoryFilter ... />
    <CommandList className="max-h-[500px]">
      <CommandEmpty ...>該当する治療プランが見つかりません。</CommandEmpty>
      {allCategories.map((category) => (
        <React.Fragment key={category}>
          <CommandGroup heading={category}>
            {items.map((item) => (
              <CommandItem
                key={item.code}
                value={`${item.name} ${item.code} ${item.category}`}
                onSelect={() => handleSelect(item)}
                className={`data-[selected=true]:${C.bgPage} cursor-pointer !py-1.5`}
              >
                <div className="flex flex-1 items-center justify-between">
                  <div className="flex flex-col gap-0.5">
                    <span className="font-medium text-sm">{item.name}</span>
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-mono bg-gray px-1 rounded">{item.code}</span>
                    </div>
                  </div>
                  <span className="font-mono font-bold text-sm">¥{item.unitPrice.toLocaleString()}</span>
                </div>
              </CommandItem>
            ))}
          </CommandGroup>
          {!activeCategory ? <CommandSeparator /> : null}
        </React.Fragment>
      ))}
    </CommandList>
  </CommandDialog>
);
```

### 比較: 正しい実装 (`MasterSelectModal.tsx:57-133`)
```tsx
return (
  <Dialog open={open} onOpenChange={onOpenChange}>
    <DialogContent className="sm:max-w-[500px]">
      <DialogHeader>
        <DialogTitle className={`text-base font-bold ${C.text}`}>{title}</DialogTitle>
      </DialogHeader>
      <div className="relative">
        <Search className={`absolute left-2.5 top-1/2 -translate-y-1/2 ${ICON.action} ${C.text40}`} />
        <Input
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          placeholder={searchPlaceholder}
          className={`pl-9 h-11 text-sm bg-white ${C.borderMedium}`}
        />
        {searchTerm ? (
          <button onClick={() => setSearchTerm("")} className={`absolute right-2.5 ...`}>
            <X className={`${ICON.xs}`} />
          </button>
        ) : null}
      </div>
      <div className="space-y-2 max-h-[400px] overflow-y-auto pr-1">
        {filtered.map((item) => (
          <div
            key={item.id}
            onClick={() => handleSelect(item)}
            className={`p-3 border rounded-lg cursor-pointer ...`}
          >
            <div className="flex-1">
              <div className={`text-sm ${C.text}`}>{item.name}</div>
              {item.price != null ? <div className="text-xs">¥{item.price.toLocaleString()}</div> : null}
            </div>
            <div className={`size-5 rounded-full ${C.bgPrimary} ...`}>
              <Check className="text-white" />
            </div>
          </div>
        ))}
      </div>
    </DialogContent>
  </Dialog>
);
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx` | メイン変更対象 | 要修正 |
| `frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx:60,214` | lazy import + 呼び出し（変更不要） | 確認済み |
| `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx:13-17,378-382` | lazy import + 呼び出し（変更不要） | 確認済み |
| `frontend/src/features/hospitalization/components/CarePlan/CarePlanDialog.tsx:17-18,232-236` | lazy import + 呼び出し（変更不要） | 確認済み |

## 修正方針

### 1. TreatmentSearchDialog.tsx の全面リライト

`CommandDialog`/`CommandInput`/`CommandList`/`CommandItem`/`CommandGroup`/`CommandSeparator` を
すべて削除し、以下に置き換える。

**import の変更 (`TreatmentSearchDialog.tsx:1-16`)**:
```tsx
// 削除
import { CommandDialog, CommandInput, CommandList, CommandEmpty, CommandGroup, CommandItem, CommandSeparator } from "@/components/ui/command";

// 追加
import { useState, useCallback, useMemo } from "react";
import { Search, X, Check } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
```

**TreatmentMasterItem 型の変更（code フィールド削除）**:
```tsx
// Before
export type TreatmentMasterItem = {
  id: string;
  code: string;
  name: string;
  unitPrice: number;
  category: string;
};

// After
export type TreatmentMasterItem = {
  id: string;
  name: string;
  unitPrice: number;
  category: string;
};
```

**カテゴリバッジ選択色の修正 (`CategoryFilter` 内)**:
```tsx
// Before (line 79)
isSelected
  ? `${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} border-transparent`

// After
isSelected
  ? `${C.bgAccent} text-white ${C.bgAccentHover} border-transparent`
```

**メインコンポーネントの JSX リライト**:
```tsx
export function TreatmentSearchDialog({ open, onOpenChange, onSelect }: TreatmentSearchDialogProps) {
  const [activeCategory, setActiveCategory] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");

  // ... API フェッチ（変更なし）

  // TREATMENT_MASTER の code フィールド生成を削除
  const TREATMENT_MASTER = useMemo(() => {
    const items: TreatmentMasterItem[] = [];
    consultations.forEach((c) => {
      if (c.isActive) items.push({ id: c.id, name: c.name, unitPrice: c.price, category: "診察" });
    });
    // ... 他カテゴリも同様
    return items;
  }, [consultations, procedures, vaccines, checkupTypes]);

  // 検索フィルタ
  const filteredItems = useMemo(() => {
    return TREATMENT_MASTER.filter((item) => {
      const matchesSearch = !searchTerm || item.name.toLowerCase().includes(searchTerm.toLowerCase());
      const matchesCategory = !activeCategory || item.category === activeCategory;
      return matchesSearch && matchesCategory;
    });
  }, [TREATMENT_MASTER, searchTerm, activeCategory]);

  // カテゴリ別グループ化
  const groupedItems = useMemo(() => {
    return filteredItems.reduce((acc, item) => {
      if (!acc[item.category]) acc[item.category] = [];
      acc[item.category].push(item);
      return acc;
    }, {} as Record<string, TreatmentMasterItem[]>);
  }, [filteredItems]);

  const handleSelect = useCallback((item: TreatmentMasterItem) => {
    onSelect(item);
    onOpenChange(false);
    setSearchTerm("");
    setActiveCategory(null);
  }, [onSelect, onOpenChange]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[560px] max-h-[80vh] flex flex-col">
        <DialogHeader>
          <DialogTitle className={`text-base font-bold ${C.text}`}>治療プラン検索</DialogTitle>
        </DialogHeader>

        {/* Search */}
        <div className="relative">
          <Search className={`absolute left-2.5 top-1/2 -translate-y-1/2 ${ICON.action} ${C.text40}`} />
          <Input
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="治療プランを検索..."
            className={`pl-9 h-11 text-sm bg-white ${C.borderMedium}`}
          />
          {searchTerm ? (
            <button
              onClick={() => setSearchTerm("")}
              className={`absolute right-2.5 top-1/2 -translate-y-1/2 ${C.text40} ${C.hoverText}`}
            >
              <X className={ICON.xs} />
            </button>
          ) : null}
        </div>

        {/* Category Filter */}
        <CategoryFilter
          categories={allCategories}
          activeCategory={activeCategory}
          onSelectCategory={setActiveCategory}
        />

        {/* Item List */}
        <div className="flex-1 overflow-y-auto space-y-1 pr-1">
          {filteredItems.length === 0 ? (
            <div className={`py-12 text-center text-sm ${C.text60}`}>
              該当する治療プランが見つかりません。
            </div>
          ) : (
            CATEGORY_ORDER.map((category) => {
              const items = groupedItems[category];
              if (!items?.length) return null;
              return (
                <div key={category}>
                  {/* Category Header */}
                  {!activeCategory ? (
                    <div className={`px-2 py-1.5 text-xs font-semibold ${C.text40} uppercase tracking-wider`}>
                      {category}
                    </div>
                  ) : null}
                  <div className="space-y-1.5">
                    {items.map((item) => (
                      <div
                        key={item.id}
                        onClick={() => handleSelect(item)}
                        className={`p-3 border rounded-lg cursor-pointer transition-all flex items-center justify-between group bg-white ${C.borderMedium} ${C.hoverBorderPrimary30} ${C.hoverBgPageHalf}`}
                      >
                        <div className="flex-1 min-w-0">
                          <div className={`text-sm font-medium ${C.text}`}>{item.name}</div>
                          <div className={`text-xs ${C.text60} mt-0.5`}>¥{item.unitPrice.toLocaleString()}</div>
                        </div>
                        <div className={`size-5 rounded-full border ${C.borderLight} group-hover:${C.borderPrimary} transition-colors shrink-0 ml-3`} />
                      </div>
                    ))}
                  </div>
                </div>
              );
            })
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

**注意**: `TreatmentSearchDialog` は追加（add）専用のため、既存選択状態を示すラジオ円はチェックなしの状態のみ表示（選択 → 即時追加 → モーダル閉じる）。

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず `C`, `STYLE` 定数を使用。`#37352F` 等ハードコード禁止。
> カテゴリバッジ選択色は `C.bgAccent`（青）を使用すること。

### `.claude/rules/typescript-react.md` — Conditional Render
> 条件付きレンダリングは必ず `condition ? <X /> : null`（`&&` 禁止）

### `.claude/rules/typescript-react.md` — React 19 パターン
> `forwardRef` 禁止。`ref` は props として受け取る。

### プロジェクト内参照実装
- `MasterSelectModal.tsx:57-133` — Dialog + Input + card 選択行のパターン
- `MasterSelectModal.tsx:65-81` — 検索バー + X クリアボタンのパターン
- `MasterSelectModal.tsx:96-127` — カード行 + ラジオ/チェック インジケーターのパターン

## 優先度
**Medium** — UX の一貫性問題。セキュリティ実害なし。次のリリースまでに対応推奨。

## 関連チケット
- FE-110: button-variant-unification（ボタン色統一 — 完了済み）

## 関連ファイル
- `frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx:1-266` — 変更対象
- `frontend/src/components/shared/MasterSelectModal/MasterSelectModal.tsx:1-135` — 参照実装
- `frontend/src/components/ui/command.tsx` — CommandDialog 定義（変更後は使用しなくなる）
- `frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx:60,214` — 呼び出し元
- `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx:13-17,378-382` — 呼び出し元
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanDialog.tsx:17-18,232-236` — 呼び出し元
