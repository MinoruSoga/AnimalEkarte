# 物販・商品マスタ 仕様書 (Merchandise Management)

## 概要
- **画面の目的**: 療法食、サプリメント、ヘアケア用品等の販売商品の定義および価格管理。
- **URLパターン**: `/settings/merchandise-items`
- **アクセス権限**: 会計マスタ管理権限が必要（`ResourceMasterMerchandise`）

---

## 1. 画面構成

### 1.1 商品一覧
- **分類**: フード(`food`)、物販(`goods`)、その他(`other`) の3カテゴリ（`MERCHANDISE_CATEGORY_OPTIONS`）。
- **項目**: 品目名、カテゴリ、単価(税込)、税率、有効/無効ステータス。JANコード・在庫数フィールドはモデルに存在しない。

### 1.2 詳細編集サイドパネル (`SidePeekPanel`)
- **品目名**: 正式な販売名称。
- **カテゴリ**: フード/物販/その他の3択。
- **単価・課税区分・税率**:
    - **単価**: `MoneyInput` による金額入力。
    - **課税区分**: 外税/内税/非課税。
    - **税率**: 10%（通常課税）/8%（軽減税率）。フード等に軽減税率を個別設定可能だが、カテゴリに応じた自動判定はない（手動選択）。
- **在庫連携**: `MerchandiseItem` モデルに在庫アイテムとの紐付けフィールドは存在しない（`backend/internal/model/merchandise_item.go`）。販売時の理論在庫自動減算機能は本マスタには実装されていない。

---

## 2. 主要な運営機能

### 2.1 レジでのクイック追加
会計画面の「物販・その他追加」において、品目名検索から素早く明細に追加できます。マスタで設定された税率が自動適用されるため、入力ミスを最小限に抑えます。

### 2.2 インボイス制度への対応
「療法食（8%）」と「ケア用品（10%）」が混在する場合でも、精算時にそれぞれの税率別合計を正確に算出し、領収書に印字します。

---

## 3. 技術仕様

### 3.1 構成コンポーネント
- **`MerchandiseItemSettings`**: メインページ。
- **`MerchandiseSidePanel`**: カテゴリ・単価・課税区分の編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/merchandise-items` | 販売商品の一覧取得 | `master-merchandise` | `view` |
| GET | `/api/v1/masters/merchandise-items/:id` | 特定の商品詳細の取得 | `master-merchandise` | `view` |
| POST | `/api/v1/masters/merchandise-items` | 新商品の登録 | `master-merchandise` | `create` |
| PATCH | `/api/v1/masters/merchandise-items/:id` | 名称・カテゴリ・単価・課税区分の変更 | `master-merchandise` | `edit` |
| DELETE | `/api/v1/masters/merchandise-items/:id` | 商品の削除 | `master-merchandise` | `delete` |
| PATCH | `/api/v1/masters/merchandise-items/reorder` | 表示順の一括保存 | `master-merchandise` | `edit` |

---

