# 支払方法マスタ 仕様書 (Payment Methods)

## 概要
- **画面の目的**: 会計精算時に選択可能な決済手段（現金、クレジットカード、電子マネー等）の定義。
- **URLパターン**: `/settings/payment-methods`
- **アクセス権限**: 支払方法マスタ管理権限が必要（`ResourcePaymentMethod`）

---

## 画面構成

### 1. 支払方法一覧
登録済みの決済手段の名称と、現在の有効/無効ステータス。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **名称**: レジ画面や領収書に印字される名前（例：VISA、PayPay）。
- **ステータス**: 利用停止した決済手段を「無効」に設定可能。

---

## 主要な機能

### 1. レジ精算・集計との連動
ここで有効化された項目のみが会計精算画面の選択肢に出現します。また、レジ締め時の金種別計上や、月次レポートの決済手段別分析の集計軸となります。

### 2. 並び順の最適化
日常業務で頻繁に使用される手段（例：現金）を上位に配置することで、精算時の操作ミスを低減します。

---

## 技術仕様

### 使用コンポーネント
- **`PaymentMethodSidePanel`**: `MasterSidePanel` による決済手段名の編集。
- **`StatusToggleButton`**: 有効/無効のトグル切り替え。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/payment-methods` | 利用可能な支払方法一覧の取得 | `master-payment-method` | `view` |
| GET | `/api/v1/payment-methods/:id` | 特定の支払方法詳細の取得 | `master-payment-method` | `view` |
| POST | `/api/v1/payment-methods` | 新規支払方法の作成 | `master-payment-method` | `create` |
| PATCH | `/api/v1/payment-methods/:id` | 名称やステータスの更新 | `master-payment-method` | `edit` |
| DELETE | `/api/v1/payment-methods/:id` | 支払方法の削除 | `master-payment-method` | `delete` |
| PATCH | `/api/v1/payment-methods/reorder` | 表示順の一括保存（BE実装済みだが本画面からは未呼出） | `master-payment-method` | `edit` |


---

