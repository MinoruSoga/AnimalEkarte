---
status: closed
closed_at: 2026-03-16
---

# [master] TrimmingSettings: コース・オプション両タブに操作列が存在する（仕様違反）

## 優先度
高

## 種別
仕様違反

## 対象ファイル
`frontend/src/features/master/routes/TrimmingSettings.tsx`

## 問題

`master-pages.md` の仕様ではコース・オプション両タブとも操作列は**なし**と定義されているが、
`COURSE_COLUMNS`（L69）・`OPTION_COLUMNS`（L78）の両方に「操作」列が含まれており、
各タブの `renderRow` 内に `<RowActionButton>` が描画されている。

```tsx
// COURSE_COLUMNS（現状）
{ header: "操作", className: "w-[80px]", align: "right" as const }  // ← 削除すべき

// OPTION_COLUMNS（現状）
{ header: "操作", className: "w-[80px]", align: "right" as const }  // ← 削除すべき
```

## 期待する状態

### タブ1（コース）
| カラム | 幅 | 備考 |
|--------|----|------|
| コース名 | flex-1 | |
| 対象サイズ | 120px | |
| 所要時間 | 100px | |
| 単価(税込) | 110px | 右揃え |
| ステータス | 90px | 右揃え、`NotionStatusPill` |

### タブ2（オプション）
| カラム | 幅 | 備考 |
|--------|----|------|
| オプション名 | flex-1 | |
| 所要時間 | 100px | |
| 組合せ可否 | 110px | 中央揃え、`CombinablePill` |
| 単価(税込) | 110px | 右揃え |
| ステータス | 90px | 右揃え、`NotionStatusPill` |

## 修正方針
1. `COURSE_COLUMNS` から `{ header: "操作", ... }` を削除
2. `OPTION_COLUMNS` から `{ header: "操作", ... }` を削除
3. `TrimmingCourseTab` の `renderRow` 内から `<RowActionButton ... />` セルを削除
4. `TrimmingOptionTab` の `renderRow` 内から `<RowActionButton ... />` セルを削除

## 完了条件
- [x] COURSE_COLUMNS に操作列がない
- [x] OPTION_COLUMNS に操作列がない
- [x] 両タブの renderRow に RowActionButton がない
- [x] ビルドエラーなし
