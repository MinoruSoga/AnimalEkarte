# BUG-305: StaffSettings — LINE グリーン色ハードコード

## 概要

`frontend/src/features/master/routes/StaffSettings.tsx:349` で `style={{ color: "#06C755" }}` とLINEブランドカラーが直接ハードコードされている。`PALETTE.lineGreen` が `design-tokens.ts` に定義済みにもかかわらず使用されていない。

## 影響ファイル

| ファイル | 違反箇所 |
|---------|---------|
| `frontend/src/features/master/routes/StaffSettings.tsx` | line 349 |

## 違反箇所と修正

```tsx
// Before (line 349)
<MessageCircle className="size-3.5" style={{ color: "#06C755" }} />

// After
<MessageCircle className="size-3.5" style={{ color: PALETTE.lineGreen }} />
```

`PALETTE` を `design-tokens.ts` からインポートを追加。

## 適用ルール

- デザイントークン規約: `#06C755` 等のHexカラー直接指定禁止。`PALETTE.*` を使用すること

## ステータス

✅ 修正済み
