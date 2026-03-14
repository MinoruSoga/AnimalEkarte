# 002: medicines シードデータに drug_category が未設定

## 概要

`backend/migrations/002_seed_master.sql` の medicines INSERT に `drug_category` カラムが含まれていない。
全アイテムが NULL になるため、Figma デザインの薬効分類グループ化が動作しない。

## 現状

```sql
-- 現在の INSERT（drug_category 列なし）
INSERT INTO medicines (name, dosage_form, medicine_unit, price, is_active, ...)
VALUES ('アモキシシリン', 'tablet', 'per_tablet', 500, true, ...);
```

全薬剤の `drug_category` が NULL → フロントエンドで「未分類」扱いとなり、
グループヘッダーなしのフラット表示になる（グループ化 UI が機能しない）。

## 対応内容

`INSERT INTO medicines` の列リストに `drug_category` を追加し、
各薬剤に適切な薬効分類を設定する。

### カテゴリ案（種別ごとに分類）

| カテゴリ名 | 対象薬剤例 |
|-----------|-----------|
| 抗生剤 | アモキシシリン、セファレキシン、エンロフロキサシン 等 |
| 消炎剤 | メロキシカム、カルプロフェン 等 |
| 制吐剤 | マロピタント、メトクロプラミド 等 |
| 胃腸薬 | スクラルファート、ラニチジン 等 |
| 駆虫薬 | フェンベンダゾール、イベルメクチン 等 |
| ステロイド | プレドニゾロン、デキサメタゾン 等 |
| 強心剤 | ピモベンダン、フロセミド 等 |
| 麻酔薬 | プロポフォール、ケタミン 等 |

## 対象ファイル

`backend/migrations/002_seed_master.sql` — medicines セクション（292行目〜）

## 優先度

**Medium** — 機能的バグではないが、デモ・開発環境でグループ化が動作しないため開発体験に影響する。

## 補足

DBリセット運用（`make reset`）のため、001_init.sql と同様に 002_seed_master.sql を直接編集してよい。
