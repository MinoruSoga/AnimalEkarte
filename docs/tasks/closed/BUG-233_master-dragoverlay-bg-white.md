# BUG-233: master 設定の DragOverlay で `bg-white` をデザイントークン未使用

## 概要

`features/master/routes/` 配下の 3 ファイルにある DragOverlay コンポーネントが `bg-white` を直接使用している。BUG-228 と同種の違反（`C.bgWhite` トークン未使用）。

## 再現手順

1. マスタ設定 > 薬品マスタ（`/settings/treatment-items` など）を開く
2. ドラッグ中のオーバーレイ要素の背景を確認する
3. **結果**: `bg-white` (Tailwind 直接指定) が使用されている

## 期待する動作

- `bg-white` → `${C.bgWhite}` に統一

## 現状コード

### `features/master/routes/MedicineSettings.tsx:189`
```tsx
// ❌
className={`flex items-center h-12 bg-white border ${C.borderLight} rounded-[4px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] cursor-grabbing`}
```

### `features/master/routes/MerchandiseItemSettings.tsx:175`
```tsx
// ❌
className={`flex items-center h-12 bg-white border ${C.borderLight} rounded-[4px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] cursor-grabbing`}
```

### `features/master/routes/CageSettings.tsx:133`
```tsx
// ❌
className={`flex items-center h-12 bg-white border ${C.borderLight} rounded-[4px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] cursor-grabbing`}
```

### 修正後
```tsx
// ✅
className={`flex items-center h-12 ${C.bgWhite} border ${C.borderLight} rounded-[4px] shadow-[0_4px_16px_rgba(0,0,0,0.12)] cursor-grabbing`}
```

## 影響範囲

| 対象 | 行 | 状態 |
|------|-----|------|
| `features/master/routes/MedicineSettings.tsx` | 189 | 未修正 |
| `features/master/routes/MerchandiseItemSettings.tsx` | 175 | 未修正 |
| `features/master/routes/CageSettings.tsx` | 133 | 未修正 |

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず **`C`**, **`STYLE`** 定数を使用

`C.bgWhite = "bg-white"` がトークンとして定義済み。

## 優先度
**Low** — 視覚的変化なし。BUG-228 の修正と同時に対応推奨。

## 関連チケット
- BUG-228: vaccinations/examinations/inventory の同種違反（14箇所）

## 関連ファイル
- `frontend/src/features/master/routes/MedicineSettings.tsx:189`
- `frontend/src/features/master/routes/MerchandiseItemSettings.tsx:175`
- `frontend/src/features/master/routes/CageSettings.tsx:133`
