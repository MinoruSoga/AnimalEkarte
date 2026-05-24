# BUG-CLEAR-BUTTON-COLOR-INCONSISTENCY: 「クリア」ボタンがプライマリカラー（黒）を使用しており視覚的に不統一

## ステータス
✅ **修正済み**（2026-05-13 確認）

> コードレビューで修正済みを確認。VaccinationHistory.tsx:96 / ImageGalleryFilter.tsx:139 の
> 「クリア」ボタンは `variant="outline"` + `C.bgWhite` + `C.borderMedium` に修正済み。
> タスクファイルのステータス未更新のまま closed に移動されていた。

## 優先度
Low

## 再現手順
1. `/medical-records/:id` → 「予防接種」タブを開く
2. 右側「予防接種履歴」エリアの「クリア」ボタン（黒背景）を確認
3. 同画面下部の「保存」ボタン（青背景）と比較

## 症状
「クリア」（検索リセット）ボタンが `C.bgPrimary` (黒: `#37352F`) で表示され、
「保存」「新規登録」等の主要アクションボタンの青色 `C.bgAccent` (`#2383E2`) と
視覚的に区別がつきにくい。

セカンダリアクションのボタンにプライマリカラー（黒）を使うのはデザイン的に不適切。
ユーザーが「クリア」を主要アクションと誤認するリスクがある。

## 該当ファイル
- `frontend/src/features/medical-records/components/VaccinationHistory.tsx`
  - `className={... C.bgPrimary ...}` が「クリア」に適用
- `frontend/src/features/medical-records/components/ImageGalleryFilter.tsx`
  - 同様のパターン

## 修正方針
「クリア」ボタンは `variant="outline"` または `variant="ghost"` に変更する。

```tsx
// Before
<Button
  className={`h-10 ${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} ...`}
>
  クリア
</Button>

// After
<Button
  variant="outline"
  className={`h-10 ${C.borderMedium} ${C.text} ${C.hoverBgPage} ...`}
>
  クリア
</Button>
```

## 影響範囲
同じパターンが使われている他コンポーネントも同様に修正が必要。
`Grep "C.bgPrimary" --include="*.tsx"` で全件確認すること。
