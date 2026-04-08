# BUG-234: PermissionBadges で `bg-neutral-100 border-neutral-200` をデザイントークン未使用

## 概要

`components/shared/PermissionBadges/PermissionBadges.tsx` の 2 箇所で Tailwind のデフォルトニュートラルカラー (`bg-neutral-100 border border-neutral-200`) を直接使用している。プロジェクトの設計トークンシステムを迂回しており、テーマ変更時に一括対応できない。

## 再現手順

1. 権限バッジが表示されるページ（マスタ設定のスタッフ/権限グループ等）を開く
2. バッジの背景色・ボーダー色を確認する
3. **結果**: Tailwind `neutral-100`（`#F5F5F5`）と `neutral-200`（`#E5E5E5`）が使用されている

## 期待する動作

- バッジ背景・ボーダーをプロジェクト設計トークンに統一する

## 現状コード

### `components/shared/PermissionBadges/PermissionBadges.tsx:35`
```tsx
// ❌ Tailwind neutral 直接使用
className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${C.text50} bg-neutral-100 border border-neutral-200`}
```

### `components/shared/PermissionBadges/PermissionBadges.tsx:60`
```tsx
// ❌ 同上（2 箇所）
className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${C.text50} bg-neutral-100 border border-neutral-200`}
```

## 修正方針

設計トークン内で最も近い値を確認して置換する。`bg-[#F7F6F3]` (C.bgPage) またはより近い muted 系トークンがあればそちらを使用。

```tsx
// ✅ 修正後（C.bgPage + C.borderLight が視覚的に近い）
className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${C.text50} ${C.bgPage} border ${C.borderLight}`}
```

もし `C.bgPage` では明度が不足する場合、`C.bgWhite` と `C.borderMediumLight` の組み合わせも検討する。

## 影響範囲

| 対象 | 行 | 状態 |
|------|-----|------|
| `components/shared/PermissionBadges/PermissionBadges.tsx` | 35 | 未修正 |
| `components/shared/PermissionBadges/PermissionBadges.tsx` | 60 | 未修正 |

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数を使用する。
> **PROHIBITED**: Tailwind デフォルトカラーの直接指定。

### プロジェクト内参照実装
- 他バッジ系コンポーネント（`StatusBadge`、`StatusPill`）はすべて `C.*` トークンを使用している

## 優先度
**Low** — 視覚的差異は軽微。ただし共有コンポーネントのため修正効果が広範囲に及ぶ。

## 関連チケット
- BUG-228、BUG-233: 同種のデザイントークン未使用

## 関連ファイル
- `frontend/src/components/shared/PermissionBadges/PermissionBadges.tsx:35,60`
