# ケージマスタ 仕様書 (Cage Management)

> 領域正本: [業務契約](../../domains/hospitalization/lifecycle.md)・[技術契約](../../../design/domains/hospitalization/technical-spec.md)。以下は既存画面／API観察を継承する。削除endpointの存在は無条件削除の許可ではなく、参照・必要履歴・現在認可を検査し、業務原本や分類不明を通常物理削除しない。無効化後も既存記録の参照を保持する。master変更は過去記録・追加済会計明細へ自動反映せず、会計の再適用はD10の明示確認契約へ従う。
## 概要
- **画面の目的**: 院内のケージ、入院室、ペットホテルの宿泊スペースの定義。
- **URLパターン**: `/settings/cage`
- **アクセス権限**: 設備マスタ管理権限が必要（`ResourceMasterHospitalization`）

---

## 画面構成

### 1. ケージ一覧
- **表示**: ケージ名、エリア、サイズ、単価、有効/無効ステータス。
- **並び順**: 一覧内の表示順序をドラッグ&ドロップ（`dnd-kit`）で管理。ステータスは有効/無効のマスタフラグであり、入院ボードの空き/使用中といったリアルタイム稼働状況は表示しない。
- **連結状態バッジ**: 連結メンバー行には `連結中: <代表ケージ名>`、代表ケージ行には `連結中（代表）: <構成メンバー名…>` を表示する（`CageSortableTable`）。

### 2. 詳細編集サイドパネル (`SidePeekPanel`)
- **ケージ名**: 院内での識別名。
- **エリア（`cageType`）**: ICU / 犬舎(`dog`) / 猫舎(`cat`) / 汎用(`general`)。
- **サイズ（`cageSize`）**: 小型、中型、大型。
- **単価**: 入院・宿泊料金の基礎となる金額。「単価」の入力は税区分に従い、表示基準とは別である。税込で入力された元額は[入力額維持の契約](../../domains/accounting/lifecycle.md#accounting-input-gross-preservation)に従い保持する。
- **ステータス**: 有効/無効の切り替え（`StatusToggleButton`。特定の「一時利用停止」概念ではなく汎用の有効/無効フラグ）。
- **備考**: 自由記述。
- **連結状態**: 連結メンバーの場合は代表ケージ名を、代表ケージの場合は構成メンバー名一覧を `連結状態` 行に表示する（`CageSidePanel`）。連結中のケージは無効化・削除できない（backend が 409 で拒否するため、先に解除が必要）。

---

## 主要な機能

### 1. 入院ボードとの連動
ここで定義されたケージは、入院一覧画面（`/hospitalization`）のボードビューで利用される。

### 2. ケージ一時連結（EMR-306）
複数のケージを一時的に「1つの入院スペース」として運用する機能。大型動物や多頭入院で隣接ケージを合体させる用途を想定する。

**モデル**
- `cages.merged_into_cage_id`（nullable、自己参照）。NULL = 独立ケージまたは代表ケージ。構成メンバーは代表ケージ ID を指す。
- 単一階層のみ（グループの入れ子は不可）。代表ケージの単価を参考情報として保持する。現行退院処理はケージ単価を自動取込せず、ホテル料金はスタッフが個別確認入力する。
- 構成メンバー上限は 1 グループ 32 件。

**操作**
- **連結**: 一覧のチェックボックスで複数選択 →「選択したケージを連結」→ 確認ダイアログで代表ケージを選択して確定。選択できるのは未連結・有効・占有中でないケージのみ（連結済み・無効行のチェックボックスは disabled）。連結ピッカーの占有判定は `GET /api/v1/masters/cages/occupancy`（`master-hospitalization:edit`、連結ピッカー専用）を使い、占有状態が unknown / 読込中 / 取得失敗のときは連結を拒否する。占有集合は対象医院の admitted/reserved 全件であり、入院ボード（[07-hospitalization-list.md](../07-hospitalization-list.md)）のフィルタ表示とは独立。最終判定は backend が担う。
- **解除**: 連結済み行（メンバー・代表双方）の解除ボタン → 確認ダイアログで確定。代表ケージ ID に対する unmerge API を呼ぶ。
- 入院ボード（`/hospitalization`）の連結モードからケージ同士のドラッグ&ドロップでも連結できる（入院側仕様書参照）。

**制約（backend で強制）**
- 連結できるのは同一 clinic・有効・未連結・占有中でないケージのみ。違反は 409。
- 連結メンバーケージへの入院割当は 422 で拒否（入院フォームのケージ選択肢からも除外）。
- 連結グループ内のいずれかのケージが占有中（admitted/reserved の入院あり）の場合、代表ケージへの新規割当は 409。
- 連結中のケージ（メンバー・代表双方）は無効化・削除できない（409）。
- 退院しても連結は自動解除されない。明示的な解除操作のみが連結を解く。
- 占有中のケージを含む解除は 409。

**割当ガード**
- `hospitalization_service.go` の Create/Update で `validateCageAssignmentGuard` がトランザクション内でグループ行をロックして検査する（昇順ロックでデッドロックを抑制）。
- 自身の入院を対象ケージへ移す更新（例: 連結前に構成ケージへ紐付いた入院を代表へ移動）は自己占有を集計から除外する。
- discharged 等の占有しないステータスの登録・更新は占有検査をスキップする（構成ケージへの構造的拒否は維持）。

マスタ定義と入院側データの関係:

```mermaid
flowchart TB
    M["ケージマスタ（本画面）<br>名称・エリア・サイズ・単価・有効フラグ<br>merged_into_cage_id（連結）"]
    M -->|定義を利用| B["入院ボード /hospitalization"]
    R["入院記録"] -->|cage_id で紐付け| M
    R --> D{"連結メンバーへの割当?"}
    D -->|422 拒否| X["代表ケージへ誘導"]
    R --> E{"グループ内に占有中ケージ?"}
    E -->|あり| Y["409 拒否"]
    E -->|なし| Z["割当可"]
```

---

## 技術仕様

### 使用コンポーネント
- **`CageSettings`**: メインページ。
- **`CageSortableTable`**: `dnd-kit` によるドラッグ並び替え一覧。
- **`CageSidePanel`**: 基本属性の編集。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/masters/cages` | 定義済みケージの一覧取得 | `master-hospitalization` | `view` |
| GET | `/api/v1/masters/cages/:id` | 特定のケージ詳細情報の取得 | `master-hospitalization` | `view` |
| POST | `/api/v1/masters/cages` | 新規ケージの登録 | `master-hospitalization` | `create` |
| PATCH | `/api/v1/masters/cages/:id` | ケージ情報の更新 | `master-hospitalization` | `edit` |
| DELETE | `/api/v1/masters/cages/:id` | ケージの削除 | `master-hospitalization` | `delete` |
| PATCH | `/api/v1/masters/cages/reorder` | 表示レイアウトの保存 | `master-hospitalization` | `edit` |
| POST | `/api/v1/masters/cages/:id/merge` | ケージの一時連結（`:id` = 代表、`member_cage_ids` = 構成・最大32件） | `master-hospitalization` | `edit` |
| POST | `/api/v1/masters/cages/:id/unmerge` | 連結グループの解除（`:id` = 代表） | `master-hospitalization` | `edit` |

---


ホテル料金・必要前金・取消料はスタッフが個別確認入力し、日数自動課金・取消時自動請求／返金は設けない。料金単位や単価masterの存在は自動積算の許可・実装済みを意味しない。

## 登録フォームの空き候補との接続（10/19）

[本人承認の新規選択制限](../../domains/hospitalization/lifecycle.md#cage-available-selection-20261019)は入院/ホテル登録フォームに適用する。マスタの連結/解除条件、代表/構成員、占有状態、自己占有除外、非占有履歴の既存契約を変更しない。[EMR-461](https://app.plane.so/baritechllc/browse/EMR-461/)の本人回答により、同院の入院作成／編集権限スタッフは空き確認に必要なケージIDだけを読める。フォーム用のこの最小readと、本書の現実装master-hospitalization:editによる連結ピッカーreadを区別する。現guardの未連結ケージ占有検査の不足は[技術差分](../../../design/domains/hospitalization/technical-spec.md#cage-availability-read-20261019)へ分離し、ここから期間別再利用や全経路実装済みを推定しない。

<a id="cross-ui-search-20261019"></a>
## 10/19後続承認：同種操作・検索ボタン方式

本画面の既存キーワード欄は[適用台帳 M11](../../shared/interaction-scope.md)と[共通FR/8AC](../../shared/ui-contracts.md#interaction-consistency-20261019)へ従う。入力中/適用済みを分け、入力/待機/消去だけでは検索せず検索ボタンで適用。既存の明示リセット、独立filter、詳細復帰、認可・検索対象・取得上限/ページ内制限を保持し、全件検索や一律の保存方式へ変えない。取得中/0件/失敗と遅延応答を分け、検索で選択/保存/コピー/送信を確定しない。旧defer/debounceの記載は入力即適用の目標に用いず、適用後の描画最適化と区別する。[S14手順](../../../ops/testing/scenarios/S14-search-and-multi-word.md#cross-ui-search-20261019)を各対象IDで実施する受入条件とし、実装適合・実機は未確認。

## EMR-461：空き確認の最小read権限

同院の入院作成／編集権限を持つスタッフへ、空き確認に必要なケージIDだけのreadを認める。患者名・診療情報・他院情報は含めず、マスタ編集・連結／解除権限を付与しない。既存API表は現実装の観察であり、この承認によるroute/grant変更済みとは扱わない。

| 条件ID | 操作 | 期待結果（未実施） |
|:---|:---|:---|
| AC-cage-occupancy-minimal-read-authorized | マスタ編集権限のない同院の入院作成／編集スタッフで空き確認 | 必要なケージIDの最小readが可能。患者情報・他院情報を返さず、編集／連結／解除は従来権限で拒否 |
| AC-cage-occupancy-minimal-read-denied | 対象医院の入院作成／編集権限なし・失効・他院を指定 | サーバーで拒否し、古い認可やclient表示で拡張しない。不明を空きにせず新規選択を止め、保存時も現在状態を再検査 |

認可実装・grant変更は今回行わない。
