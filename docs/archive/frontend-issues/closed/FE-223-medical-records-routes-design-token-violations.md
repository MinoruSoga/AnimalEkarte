# FE-223: MedicalRecords.tsx のデザイントークン違反

## 概要

`frontend/src/features/medical-records/routes/MedicalRecords.tsx` で
直接 Tailwind カラークラスが使用されている。

## 違反箇所

| 行 | 違反コード | 用途 | 修正方針 |
|----|-----------|------|---------|
| 265 | `text-red-500` | 無効なスタッフ名の警告テキスト | `C.danger` または `C.textRequired` |
| 269 | `text-red-500` | 無効な担当医のアラートアイコン | 同上 |

## 修正方針

```tsx
// Before
<span className="text-red-500">警告メッセージ</span>

// After
import { C } from "@/lib/design-tokens";
<span style={{ color: C.danger }}>警告メッセージ</span>
// または design-tokens で定義済みの Tailwind クラスがあれば className で使用
```

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: 直接 Tailwind カラークラスの指定は厳禁。

## 優先度
**Low** — 2箇所のみ。機能的障害なし。

## 関連ファイル
- `frontend/src/features/medical-records/routes/MedicalRecords.tsx`
- `frontend/src/lib/design-tokens.ts`
