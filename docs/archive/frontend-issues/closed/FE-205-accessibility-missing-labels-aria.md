# FE-205: アクセシビリティ違反（Label 未関連付け・aria-label なし・非セマンティック HTML）

## 概要

複数コンポーネントで WCAG 2.1 Level A/AA に違反するアクセシビリティ問題が存在する。
フォームラベルの未関連付け、アイコンボタンの説明なし、非セマンティック HTML が主な問題。

## 問題一覧

### 1. Label と Input が関連付けられていない（WCAG 2.1 AA: 1.3.1）

#### `frontend/src/features/medical-records/components/ImageGalleryFilter.tsx:85-92`
```tsx
// Before: Label に htmlFor なし、Input に id なし
<Label className={`text-sm font-medium ${C.text60}`}>
  検索単語
</Label>
<Input
  value={searchTerm}
  onChange={(e) => onSearchChange(e.target.value)}
  className={`bg-white ${C.borderMedium} h-10 text-sm`}
/>

// After
<Label htmlFor="image-gallery-search" className={`text-sm font-medium ${C.text60}`}>
  検索単語
</Label>
<Input
  id="image-gallery-search"
  value={searchTerm}
  onChange={(e) => onSearchChange(e.target.value)}
  className={`bg-white ${C.borderMedium} h-10 text-sm`}
/>
```

#### `frontend/src/features/hospitalization/components/DailyRecord/DailyCareLogDialog.tsx:97-102`
```tsx
// Before: Input に Label なし（placeholder のみ）
<Input
  placeholder={getPlaceholder()}
  value={form.value}
  ...
/>

// After
<Label htmlFor="care-log-value">{getLabel()}</Label>
<Input
  id="care-log-value"
  placeholder={getPlaceholder()}
  value={form.value}
  ...
/>
```

#### `frontend/src/features/hospitalization/components/DailyRecord/TaskCompleteDialog.tsx:78-79`
```tsx
// Before: Textarea に Label なし（placeholder のみ）
<Textarea
  placeholder="特記事項があれば入力..."
  ...
/>

// After
<Label htmlFor="task-memo">実施メモ（任意）</Label>
<Textarea
  id="task-memo"
  placeholder="特記事項があれば入力..."
  ...
/>
```

### 2. 非セマンティック HTML — `<span role="button">` の使用（WCAG 2.1 A: 4.1.2）

#### `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx:599-614`
```tsx
// Before: <span role="button"> — ネイティブ button でなくキーボードサポートが不完全
<span
  role="button"
  tabIndex={0}
  onClick={onClick}
  onKeyDown={(e) => {
    if (e.key === "Enter" || e.key === " ")
      onClick(e as unknown as React.MouseEvent<HTMLSpanElement>);
  }}
  aria-label="日付をクリア"
>
  <X className={ICON.action} />
</span>

// After: ネイティブ <button> を使用
<button
  type="button"
  onClick={onClick}
  className={`ml-1 shrink-0 rounded p-0.5 ${C.text40} ${C.hoverBgPrimary10} cursor-pointer`}
  aria-label="日付をクリア"
>
  <X className={ICON.action} />
</button>
```

### 3. アイコンのみのナビゲーションボタンに aria-label なし（WCAG 2.1 AA: 1.1.1）

#### `frontend/src/features/hospitalization/components/DailyRecord/DailyRecordDateNav.tsx:22-31`
```tsx
// Before: aria-label も title もない
<Button variant="ghost" size="icon" onClick={onPrev} className="h-11 w-11">
  <ChevronLeft className={ICON.lg} />
</Button>
<span className={...}>{formattedDate}</span>
<Button variant="ghost" size="icon" onClick={onNext} className="h-11 w-11">
  <ChevronRight className={ICON.lg} />
</Button>

// After
<Button variant="ghost" size="icon" onClick={onPrev} className="h-11 w-11" aria-label="前の日付">
  <ChevronLeft className={ICON.lg} />
</Button>
<span className={...}>{formattedDate}</span>
<Button variant="ghost" size="icon" onClick={onNext} className="h-11 w-11" aria-label="次の日付">
  <ChevronRight className={ICON.lg} />
</Button>
```

#### `frontend/src/features/hospitalization/components/CarePlan/CarePlanItemRow.tsx:82-84`
```tsx
// Before: アイコンボタンに aria-label なし
<Button variant="ghost" size="sm" onClick={() => onEdit(plan)} className="h-9 w-9 p-0 ...">
  <Edit2 className={`${ICON.action} ...`} />
</Button>

// After
<Button
  variant="ghost"
  size="sm"
  onClick={() => onEdit(plan)}
  className="h-9 w-9 p-0 ..."
  aria-label={`${plan.name}を編集`}
>
  <Edit2 className={`${ICON.action} ...`} />
</Button>
```

## 影響範囲

| 対象 | 行番号 | WCAG | 状態 |
|------|--------|------|------|
| `frontend/src/features/medical-records/components/ImageGalleryFilter.tsx` | 85-92 | 1.3.1 (A) | 要修正 |
| `frontend/src/features/hospitalization/components/DailyRecord/DailyCareLogDialog.tsx` | 97-102 | 1.3.1 (A) | 要修正 |
| `frontend/src/features/hospitalization/components/DailyRecord/TaskCompleteDialog.tsx` | 78-79 | 1.3.1 (A) | 要修正 |
| `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx` | 599-614 | 4.1.2 (A) | 要修正 |
| `frontend/src/features/hospitalization/components/DailyRecord/DailyRecordDateNav.tsx` | 22-31 | 1.1.1 (AA) | 要修正 |
| `frontend/src/features/hospitalization/components/CarePlan/CarePlanItemRow.tsx` | 82-84 | 1.1.1 (AA) | 要修正 |

## 準拠すべきプロジェクト規約

### `.claude/rules/accessibility-rules.md` — フォームラベル（必須）
> `label` と `input` を `htmlFor` で関連付けること。`placeholder` のみは禁止。

### `.claude/rules/accessibility-rules.md` — セマンティック HTML
> `<span onClick>` の代わりに `<button>` を使用すること。

### `.claude/rules/accessibility-rules.md` — ARIA 属性
> アイコンのみのボタンには `aria-label` で機能を説明すること。

## 優先度
**Medium** — WCAG 2.1 Level A 違反（フォームラベル、セマンティック HTML）は修正必須。

## 関連ファイル
- `frontend/src/features/medical-records/components/ImageGalleryFilter.tsx:85-92`
- `frontend/src/features/hospitalization/components/DailyRecord/DailyCareLogDialog.tsx:97-102`
- `frontend/src/features/hospitalization/components/DailyRecord/TaskCompleteDialog.tsx:78-79`
- `frontend/src/components/shared/NotionDatePicker/NotionDatePicker.tsx:599-614`
- `frontend/src/features/hospitalization/components/DailyRecord/DailyRecordDateNav.tsx:22-31`
- `frontend/src/features/hospitalization/components/CarePlan/CarePlanItemRow.tsx:82-84`
