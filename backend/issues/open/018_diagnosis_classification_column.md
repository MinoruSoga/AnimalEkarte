---
status: open
---

# 診断病名マスタ「分類」列: item_type フィールドの要否検討

## 背景

Figmaデザイン（`/settings/diagnosis`）の診断病名マスタページに「分類」列が存在し、
`diagnosis_category` タブでは全行が `"diagnosis_category"`、
`diagnosis_name` タブでは全行が `"diagnosis_name"` を表示している。

## 問題

現在のバックエンドモデル（`DiagnosisCategory`, `DiagnosisName`）に `classification` / `item_type` フィールドが存在しない。

フロントエンドでは該当カラムをタブごとの**静的定数**として表示しているが、これは STI（Single Table Inheritance）廃止前の `master_items.item_type` カラムの名残と思われる。

```go
// 現在のモデルに classification / item_type は存在しない
type DiagnosisCategory struct {
    ID          uint64
    ClinicID    uint64
    Name        string
    IsActive    bool
    Description string
    SortOrder   int
    // ← item_type / classification がない
}
```

## 選択肢

### A. 「分類」列を廃止（推奨）
- 全行同じ値（`"diagnosis_category"` / `"diagnosis_name"`）はユーザーに情報価値ゼロ
- Figmaデザインから削除し、フロントエンドのコードからも除去する
- バックエンド変更不要

### B. 静的定数として維持
- フロントエンド実装は現状通り（タブ固有の定数文字列を表示）
- バックエンド変更不要
- ただし意味のない列としてUIに残り続ける

### C. backend に `item_type` 文字列カラムを追加
- 将来的に複数の分類体系が必要になった場合のみ検討
- 現時点では過剰エンジニアリング

## 完了条件

- [ ] 製品オーナーまたはデザイナーが A/B/C のいずれかを選択
- [ ] 選択に応じてフロントエンド・バックエンドを修正
