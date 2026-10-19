# 割引キャンペーンマスタ 仕様書 (Campaigns Settings)

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | page.settings.master-campaigns／画面仕様 |
| 決定状態 | 既存本文は継承（決定根拠未確認）。第六群の合意と未決は下記正本へ |
| 実装状態 | 既存記載を継承。今回の目標機能／認可適合は未確認 |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-15 |

### 会計第六群との対応

[第六群の業務正本](../../domains/accounting/lifecycle.md#accounting-agreement-sixth)と[マスタ変更権限](../../domains/accounting/lifecycle.md#requirement-fr-accounting-master-authorized-maintenance)へ従う。マスタ変更者は管理者と対象変更権限付きスタッフ。現行APIのaccountingと割引値操作のdiscountは別で、会計操作権限だけを新業務のマスタ編集許可にしない。全体割引からの品目除外は品目マスタで指定し、キャンペーンの対象選択や保険対象外と混同しない。 全スタッフに変更を開放せず、対象grant・医院scopeを検査する。今回、実権限・設定・API／DBは変更していない。詳細な[現行mapping](../../../design/domains/accounting/technical-spec.md#accounting-master-permission-map)は技術仕様へ。


### 会計の保存・参照保護契約

[D10のmaster責務と互換](../../../design/domains/accounting/technical-spec.md#accounting-master-coverage)・[画面取得／失敗／保存](../../../design/domains/accounting/technical-spec.md#accounting-screen-contract)へ従う。現在認可・clinic／対象版を検査し、無権限／競合は保存しない。取得失敗を空masterとせず未保存入力を保護。名称変更／無効化／削除で過去請求／資金・帳票参照を失わせず前後値を追跡。旧CRUD表は現行参照で、新業務未対応経路はadapterと実装受入まで利用開始しない。

[第十群](../../domains/accounting/lifecycle.md#accounting-empty-rounding-master)で全体割引除外は初期値／設定欠落時とも対象外にしない。明示設定を保持し保険／税／明細割引と混同しない。追加済み会計行はmaster変更で自動更新せず明示再適用へ。単価×数量の行額は値引き前に行別1円四捨五入。臨床数量・投薬量や確定済み会計／旧帳票は直接変更しない。

## 概要
- **画面の目的**: 会計施策としてのキャンペーンを期間・割引条件付きで定義し、明細計算時に適用可能にする。
- **URLパターン**: `/settings/campaigns`
- **アクセス権限**: 会計権限（`ResourceAccounting`）

---

## 1. 画面構成

### 1.1 一覧
- **一覧項目**: キャンペーン名、期間、割引値、ステータス、編集アクション。
- **対象範囲検索**: キャンペーン名での文字列検索。
- **並び順**: 画面上のドラッグ並び替え UI は未実装（バックエンドの `reorder` API は存在するが未使用）。

### 1.2 編集パネル (`SidePeekPanel`)
- **名称**: キャンペーン名（必須）。
- **実施期間**: 開始日 (`startDate`) と終了日 (`endDate`)。終了日は開始日以降を必須チェック。
- **割引種別**:
  - `rate`: 割引率（%）
  - `amount`: 割引額（円）
- **対象カテゴリ**: フード、物販、処方、診察、検査、処置、手術、ワクチン、トリミング、ホテル、しつけ、その他 などの複数選択。
- **対象商品**: 商品名での検索後、複数選択。
- **有効フラグ**: `StatusToggleButton` で ON/OFF。

---

## 2. 運用ルール
- **日付チェック**: `終了日 < 開始日` の場合は保存不可。
- **名称バリデーション**: 空名は保存不可。
- **適用対象**: 会計画面の明細計算で、対象カテゴリ・商品・期間が一致する明細に割引を適用。

明細への適用判定:

```mermaid
flowchart LR
    L["会計画面の明細"] --> Q{"対象カテゴリ・商品・期間が一致?"}
    Q -->|一致| A["キャンペーン割引を適用<br>rate: 割引率 / amount: 割引額"]
    Q -->|不一致| N["適用なし"]
```

---

## 3. 技術仕様

### 3.1 構成コンポーネント
- **`CampaignSettings`**: 一覧・フィルタ・保存の制御。
- **`CampaignSidePanel`**: 名称・期間・割引設定・対象条件の編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/campaigns` | キャンペーン一覧取得 | `accounting` | `view` |
| GET | `/api/v1/masters/campaigns/:id` | 特定のキャンペーン情報の取得 | `accounting` | `view` |
| POST | `/api/v1/masters/campaigns` | キャンペーン作成 | `accounting` | `create` |
| PATCH | `/api/v1/masters/campaigns/:id` | キャンペーン更新 | `accounting` | `edit` |
| DELETE | `/api/v1/masters/campaigns/:id` | キャンペーン削除 | `accounting` | `delete` |
| PATCH | `/api/v1/masters/campaigns/reorder` | キャンペーンの表示順一括保存（BE実装済みだが本画面からは未呼出） | `accounting` | `edit` |

<a id="cross-ui-search-20261019"></a>
## 10/19後続承認：同種操作・検索ボタン方式

本画面の既存キーワード欄は[適用台帳 M12](../../shared/interaction-scope.md)と[共通FR/8AC](../../shared/ui-contracts.md#interaction-consistency-20261019)へ従う。入力中/適用済みを分け、入力/待機/消去だけでは検索せず検索ボタンで適用。既存の明示リセット、独立filter、詳細復帰、認可・検索対象・取得上限/ページ内制限を保持し、全件検索や一律の保存方式へ変えない。取得中/0件/失敗と遅延応答を分け、検索で選択/保存/コピー/送信を確定しない。旧defer/debounceの記載は入力即適用の目標に用いず、適用後の描画最適化と区別する。[S14手順](../../../ops/testing/scenarios/S14-search-and-multi-word.md#cross-ui-search-20261019)を各対象IDで実施する受入条件とし、実装適合・実機は未確認。
