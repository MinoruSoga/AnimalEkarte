# BUG-322: Icon Button Size7/Size8 スタイルトークン欠落

## 優先度
LOW

## 問題の説明

`src/lib/design-tokens.ts` の `STYLE` オブジェクトに Icon Button 用のスタイルトークンが部分的に欠落している。

現在実装：
- `STYLE.iconBtn20` のみ存在

不足しているサイズ：
- `STYLE.iconBtn24` (Size7 相当)
- `STYLE.iconBtn28` (Size8 相当)

## 期待動作

```typescript
// src/lib/design-tokens.ts
export const STYLE = {
  // ...
  iconBtn20: "w-5 h-5 flex items-center justify-center rounded",
  iconBtn24: "w-6 h-6 flex items-center justify-center rounded",  // ← 追加
  iconBtn28: "w-7 h-7 flex items-center justify-center rounded",  // ← 追加
  // ...
};
```

## 修正方法

1. `src/lib/design-tokens.ts` に以下を追加：
```typescript
iconBtn24: "w-6 h-6 flex items-center justify-center rounded",
iconBtn28: "w-7 h-7 flex items-center justify-center rounded",
```

2. 既存で直接値指定している箇所を検索・置換：
```bash
grep -r "w-6 h-6.*rounded" frontend/src --include="*.tsx" --include="*.ts"
grep -r "w-7 h-7.*rounded" frontend/src --include="*.tsx" --include="*.ts"
```

3. 該当箇所を `STYLE.iconBtn24`, `STYLE.iconBtn28` に置換

## テスト確認項目
- [ ] 新規トークンが `design-tokens.ts` に追加される
- [ ] 既存の直接値指定箇所が新規トークンに置換される
- [ ] Lighthouse/a11y スコア低下なし

## 修正優先度
- LOW（機能的には正常動作、コンシステンシー向上のため）
- 次回のデザインシステム整理タイミングで実装推奨
