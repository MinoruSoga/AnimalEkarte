# BUG-298: LINE公式グリーン (#06C755) のハードコード — design-tokens未使用

## 概要

LINE公式ブランドカラー `#06C755` が4箇所でハードコードされており、`design-tokens.ts` の `PALETTE` を経由していない。

## 違反箇所

| ファイル | 行 | 用途 |
|---------|-----|------|
| `features/master/routes/ServiceTypeSettings.tsx:136` | `style={{ color: "#06C755" }}` | LINEアイコン |
| `features/line-reservation/components/LinkedLineCustomers.tsx:68` | `style={{ color: "#06C755" }}` | LINEアイコン |
| `features/line-reservation/components/LinkedLineCustomers.tsx:97` | `style={{ backgroundColor: "#06C755" }}` | LINEアバター背景 |
| `features/line-reservation/components/LinkedLineCustomers.tsx:201` | `style={{ backgroundColor: "#06C755" }}` | LINEアバター背景 |

## 修正

`design-tokens.ts` の PALETTE に追加:
```typescript
/** LINE official brand green */
lineGreen: "#06C755",
```

各ファイルで:
```tsx
// Before
style={{ color: "#06C755" }}

// After
import { C, PALETTE } from "@/lib/design-tokens";
style={{ color: PALETTE.lineGreen }}
```

## ステータス

- [x] ドキュメント作成
- [x] 実装完了
