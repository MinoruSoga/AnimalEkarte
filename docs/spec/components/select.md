# Select

> Radix Select ベースのドロップダウン選択。配置: `frontend/src/components/ui/select.tsx`
> 使用箇所: 多数（検索は `rg -l "components/ui/select"`）/ Storybook: `select.stories.tsx`

## Anatomy

- `Select`（Root）→ `SelectTrigger`（ボタン）→ `SelectValue`（選択表示）
- `SelectContent`（Portal + Viewport）→ `SelectItem`（CheckIcon indicator 内蔵）
- `SelectScrollUpButton` / `SelectScrollDownButton`（内部）

## Props / Variants

| prop | 型 | 既定 | 説明 |
|------|----|----|------|
| `SelectTrigger.size` | `"sm" \| "default"` | `default` | 現状どちらも `h-11`（44px タッチターゲット統一） |
| `SelectContent.position` | `"popper" \| "item-aligned"` | `popper` | Radix ポジショニング |

## 状態（design-states.md §1）

| 状態 | 実装 | トークン/クラス | 備考 |
|------|------|----------------|------|
| hover | ○ | `STATE.hoverBgInput` | trigger 背景が warm neutral に |
| focus-visible | ○ | `STATE.focusBorderLegacyAccent` + `STATE.focusRingActionPrimary` + `focus:bg-white` | Input/Textarea と同一レシピ |
| disabled | ○ | `disabled:opacity-50` + `cursor-not-allowed` | |
| loading | ✕ | — | Select 自身に loading 概念なし（options 取得中は親が制御） |
| invalid | ○ | `aria-invalid:border-destructive aria-invalid:bg-destructive/5` | `aria-invalid="true"` で発火 |
| selected | ○ | item 内 CheckIcon + `focus:bg-accent` | Radix が aria-selected を管理 |

## アクセシビリティ要件

- Radix Select が listbox/option ロール・typeahead・矢印キー巡回を提供。
- Trigger は必ず `SelectValue` の `placeholder` か選択値で意味を持たせる。関連付けは `Label htmlFor` + trigger の `id`。
- `aria-invalid` 時はエラーテキストを別途出す（色のみに依存しない — design-states.md §3）。

## 使い方（Do / Don't）

- Do: 固定の選択肢リストに使う。検索が必要なら `searchable-select.tsx`（Command 複合）を使う。
- Don't: `SelectItem` の value に空文字列を渡さない（Radix が placeholder 扱いにする）。

## 既知のギャップ

- `size="sm"` が視覚的に `default` と同じ h-11 — dense UI での縮小バリアントは未実装（design-system.md §7 の dense 例外と要整合）。
- 状態クラスは `STATE`（L2.5）を参照 — design-token-layers.md §3 の層構造に準拠済み。
