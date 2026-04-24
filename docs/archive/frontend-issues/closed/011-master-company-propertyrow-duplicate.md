---
status: closed
closed_at: 2026-03-16
---

# [master] CompanySettings.tsx が PropertyRow をローカル再定義している（DRY 違反）

## 優先度
中

## 種別
冗長コード / DRY 原則違反

## 対象ファイル
`frontend/src/features/master/routes/CompanySettings.tsx`

## 問題

`CompanySettings.tsx` ファイル内で `PropertyRow` コンポーネントをローカルに再定義している。

```tsx
// CompanySettings.tsx 内（削除すべき）
function PropertyRow({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-4 py-2 border-b">
      <dt className="w-32 text-sm text-muted-foreground shrink-0">{label}</dt>
      <dd className="flex-1">{children}</dd>
    </div>
  );
}
```

`@/components/shared/SidePeek/PropertyRow` に同名・同目的のコンポーネントが存在しており、
二重管理状態になっている。プロパティの仕様変更時にローカル版のみ更新し忘れるリスクがある。

## 修正方針

1. ローカルの `PropertyRow` 定義を削除する
2. `import { PropertyRow } from "@/components/shared/SidePeek/PropertyRow"` に変更する
3. スタイルが異なる場合は共通 `PropertyRow` に variant prop を追加するか、CSS を合わせる

## 付記

同じファイルで `useEffect` 内の `setState` パターン（`company → formData` 変換）が
`handleCloseEdit` と重複している。共通ヘルパー関数 `companyToFormData(company)` を作成して DRY にすること。
