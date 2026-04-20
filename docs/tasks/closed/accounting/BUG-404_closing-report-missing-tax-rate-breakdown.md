# BUG-404: 集計・締めレポートで消費税区分（8%/10%）の按分が未実装

**作成日**: 2026-04-20  
**Priority**: HIGH（月次レポートは税務・会計処理に使用）  
**Affects**: FEAT-368（集計・締め）, `features/accounting`

---

## 問題

`billing_items.tax_rate` および `clinic.standard_tax_rate` / `clinic.reduced_tax_rate` で  
8%・10% の消費税区分はデータモデルとして管理済みだが、  
**レジ締め集計・月次レポートで按分表示が実装されていない（漏れ）。**

経理向け月次売上集計には税率別の按分が必要。

---

## 現状確認

| ファイル | 実装状況 |
|---------|---------|
| `backend/internal/model/accounting.go` | `BillingItem.TaxRate float64`（per-item で管理済み） |
| `backend/internal/model/clinic.go` | `StandardTaxRate 0.10` / `ReducedTaxRate 0.08`（定義済み） |
| `backend/internal/model/merchandise_item.go` | `TaxRate float64`（定義済み） |
| `backend/internal/model/consultation.go` | `TaxRate float64`（定義済み） |
| 集計エンドポイント | **0件（未実装）** |
| 月次レポート画面 | **未実装** |

---

## 要件

### レジ締め集計への追加表示

締め集計マトリクスの下部に税率別サマリを追加する。

```
消費税内訳
  標準税率（10%）対象額: ¥ xx,xxx  うち消費税: ¥ x,xxx
  軽減税率（8%）対象額:  ¥ xx,xxx  うち消費税: ¥ x,xxx
  消費税合計:                         ¥ x,xxx
```

### 月次レポートへの追加表示

月次サマリに税率別按分カラムを追加する。

| 項目 | 内容 |
|-----|------|
| 標準税率（10%）課税対象額 | 月間の 10% 対象品目の税抜合計 |
| 標準税率（10%）消費税額 | 税額合計 |
| 軽減税率（8%）課税対象額 | 月間の 8% 対象品目の税抜合計 |
| 軽減税率（8%）消費税額 | 税額合計 |

### CSV エクスポートへの追加列

```
..., 標準税率課税対象額, 標準税率消費税額, 軽減税率課税対象額, 軽減税率消費税額
```

---

## 実装方針

1. `billing_items.tax_rate` でグループ化して集計 SQL を組む（`GROUP BY tax_rate`）
2. `cash_register_closes.category_breakdown` JSONB にも税率別内訳をスナップショット保存する
   ```json
   {
     "tax_breakdown": {
       "standard": { "taxable_amount": 100000, "tax_amount": 10000 },
       "reduced":  { "taxable_amount": 20000,  "tax_amount": 1600 }
     }
   }
   ```
3. フロントエンドは `TaxBreakdown` 型で受け取り、締め画面・月次レポート双方で表示する

---

## 対応フェーズ

FEAT-368 の他機能（レジ締め・月次レポート）と同一フェーズで実装する。  
単独で先行実装しない（バックエンドの集計 API が土台となるため）。

---

## 関連

- FEAT-368: `docs/tasks/pending/accounting/FEAT-368_closing-aggregation.md`
- Q10 確定: 月次レポートに 8%/10% の按分表示が**必要**
