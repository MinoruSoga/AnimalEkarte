# 在庫の業務契約

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.inventory.business-rules／業務仕様 |
| 決定状態 | 既存記載を継承（決定根拠未確認の部分を含む）。I1の目標は依頼者2026-10-15承認（Sentinel_7a67fd30e47c819192cd36d416f0a2bf）。既存詳細全体の再承認ではない。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-15 |

## 目的・対象範囲

D13は医院内の在庫品目、理論数量・管理単位・閾値・保管／仕入情報を管理する。薬剤・物販価格マスタ、臨床実施、請求・返金とは別責務。物販マスタと在庫の自動連動は現行未接続であり、名称一致で在庫を減らさない。自動発注・ロット／シリアル管理・期限順出庫・棚間移動・複数医院移管は既存記載から確定しない。

## 担当者・権限・前提

inventory の view／create／edit／deleteを区別し同医院の品目だけを扱う。閲覧権限だけで現在庫を変更できない。診療masterの変更権限は在庫調整を自動的に許可するものではない。薬剤作成・名称変更・削除の既存副作用は技術差として記録し、通常在庫削除の依存保護と同等であると推定しない。

## 契機・入口・出口

/inventory の一覧・警告から /inventory/:id、/inventory/new へ。臨床訂正の影響確認からは元記録／元項目・訂正参照・在庫ID／現在版・理由を受ける。臨床側の保存成功／画面遷移では在庫成功とせず、在庫ownerの結果を返す。

## 状態遷移・引継ぎデータ

<a id="requirement-fr-inventory-quantity-status"></a>
### FR-inventory-quantity-status：数量と表示状態

現行契約を継承し、現在庫・最低在庫は0以上整数。品名・単位を必須とする。カテゴリは医薬品／消耗品／フード／その他の固定4種で、価格マスタのカテゴリと同一でない。

公開状態は数量から導出し保存statusを正本にしない。現行model／repository双方の式は、数量0以下＝在庫切れ、数量正かつ最低在庫正で数量≦最低在庫＝残少、それ以外＝十分。最低在庫0は残少閾値未設定を意味し、数量0の在庫切れを無効にしない。旧画面の『0なら常に十分』は現行式と矛盾する誤記として是正する。表示とフィルタで同じ式を用いる。

<a id="requirement-fr-inventory-record-maintenance"></a>
### FR-inventory-record-maintenance：品目管理と失敗表示

現行フォームは現在庫の絶対値を編集する。変更後の取得失敗を保存失敗と誤認させないため更新・再取得は同一保存単位を継承する。対象なし・閲覧禁止は空フォームを表示せず、他の取得失敗は再試行に案内する。利用中の在庫は削除拒否。削除UIが存在しない現状と、APIの論理削除契約を区別する。

単位・最低在庫・属性変更と実物の入出庫は異なる。今のフォームが調整履歴を保持しているとの主張はしない。有効期限・最終入荷日は品目に一つの属性で、複数ロットの履歴を表していると推定しない。

<a id="requirement-fr-inventory-clinical-boundary"></a>
### FR-inventory-clinical-boundary：臨床記録・請求からの独立

D03／D10の既決境界を継承。カルテ訂正・過去入力・患者転記・復元・請求取消・減額・返金だけでは在庫を自動増減しない。既存通常治療createのinventory ID指定時減算はその経路の実装であり、過去入力や訂正へ流用しない。関連不明を影響なしとせず確認を残す。

臨床の数量・投与単位と在庫の管理単位は別。現在の在庫減算は正整数を要求し、小数を切り捨てて減算0にする経路を認めない。臨床上の小数投与の意味や正しさをD13の整数制約だけで否定しない。単位変換・包装開封等が必要なら対応を確定するまで在庫成功を偽装しない。

<a id="requirement-fr-inventory-adjustment"></a>
### FR-inventory-adjustment：実物に基づく理由付き調整

承認済みI1：入荷・使用／出庫・廃棄・棚卸差異・実物返却を区別する。対象品目・管理単位・増減数または実査数・理由・実日時・担当・元記録参照を保存し、変更前／変更後と今回差分を確認できるようにする。処置取消だけで未使用・返品可能と推定せず、実際に戻った数量を確認して一度だけ戻す。

在庫不足は負数を作らず当該減算確定を拒否し、臨床記録との結果を分離して確認対象にする。同時減算・棚卸上書きは現在版を再確認し、古い画面の数量で新しい出庫を消さない。調整履歴の訂正も元履歴・理由・差分を保持し、二重返却を防ぐ。具体的API／画面・監査方式は設計であり、現在の絶対値編集から実装済みとはしない。

<a id="requirement-fr-inventory-master-separation"></a>
### FR-inventory-master-separation：品目識別と他マスタ

物販マスタは名称・価格・税等の定義で、在庫品目に対応するIDを持たない現状。販売時・請求取消時の自動増減は今回推定しない。薬剤masterには在庫参照があり、在庫生成・名称同期等の現行経路がある。これを同名別品目や他医院の在庫への変更許可にしない。D03の無効化・当時内容保持と、在庫依存の保持を照合する。

## 異常系・保存・取消

- 入力不備は対象フィールドへ返す。空品名／単位、負数・小数の現在庫／最低在庫を受理しない。
- 在庫0・閾値0・残少は別で、在庫不足を警告だけで負数保存してよいとはしない。
- 再送時は現在数量と当該操作結果を照合する。調整の操作ID／同時編集の版・結果照会は[選定技術契約](../../../design/domains/finance-operations/technical-spec.md#inventory-selected-contract)に従う。
- 一覧はサーバーページ（既定20）、一致カテゴリ／状態はサーバー、否定フィルタと全文検索・並べ替えは取得ページ内。0件を全医院在庫不存在と読まない。
- 外部発注・配送・薬剤購入は本仕様で実行しない。発注点警告は発注済みではない。

## 受入条件

| 条件ID | 前提／操作（Given / When） | 期待結果（Then） | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-inventory-required | 品名または単位を空にして保存 | 入力拒否し既存品目不変 | [FR-inventory-quantity-status](#requirement-fr-inventory-quantity-status)。項目エラー確認、未実施 |
| AC-inventory-integer | 数量／最低在庫へ負数・小数を入力 | 保存拒否 | [FR-inventory-quantity-status](#requirement-fr-inventory-quantity-status)。0／正整数対照、未実施 |
| AC-inventory-zero-zero | 数量0・最低在庫0を表示／絞込 | 在庫切れで一致 | [FR-inventory-quantity-status](#requirement-fr-inventory-quantity-status)。modelと一覧比較、未実施 |
| AC-inventory-threshold | 数量1、最低在庫0／1／2で比較 | 順に十分／残少／残少 | [FR-inventory-quantity-status](#requirement-fr-inventory-quantity-status)。境界比較、未実施 |
| AC-inventory-status-input | 数量由来状態と異なるstatusを送信 | 保存statusを正本とせず数量から表示 | [FR-inventory-quantity-status](#requirement-fr-inventory-quantity-status)。応答／フィルタ比較、未実施 |
| AC-inventory-fetch-failure | 編集対象取得に失敗 | 空編集フォームを出さず拒否／再試行を区別 | [FR-inventory-record-maintenance](#requirement-fr-inventory-record-maintenance)。404／403／通信失敗、未実施 |
| AC-inventory-used-delete | 利用中の品目を削除 | 依存を保って拒否、競合でも同じ | [FR-inventory-record-maintenance](#requirement-fr-inventory-record-maintenance)。利用参照を照合、未実施 |
| AC-inventory-no-edit-grant | viewのみで数量更新を要求 | 読み取り表示に限定、サーバーでも拒否 | 担当者節。権限別検査、未実施 |
| AC-inventory-no-auto-restore | カルテ訂正または会計取消 | 自動増減なし、元参照付き影響確認へ | [FR-inventory-clinical-boundary](#requirement-fr-inventory-clinical-boundary)。各owner結果を照合、未実施 |
| AC-inventory-fractional-use | 小数投与量を正整数のみの在庫減算へ渡す | 切捨て0減算をせず、単位対応の確認へ | [FR-inventory-clinical-boundary](#requirement-fr-inventory-clinical-boundary)。数量／単位を照合、未実施 |
| AC-inventory-merchandise | 在庫非接続の物販を請求 | 品名一致等による自動減算なし | [FR-inventory-master-separation](#requirement-fr-inventory-master-separation)。参照／数量比較、未実施 |
| AC-inventory-insufficient | 数量1に対し2の減算を要求 | 在庫不足を拒否し数量1を維持 | [FR-inventory-clinical-boundary](#requirement-fr-inventory-clinical-boundary)。現行原子的減算を照合、未実施 |
| AC-inventory-movement-history | I1の目標で、入荷・廃棄・棚卸調整 | 理由／担当／実日時／前後数／元参照を記録 | 承認I1。操作種別ごと照合、未実施 |
| AC-inventory-actual-return | I1の目標で、処置取消後に実物1個だけ返却 | 実物確認量のみ1回増加、取消だけでは0 | 承認I1。取消／返却結果を比較、未実施 |
| AC-inventory-concurrent-count | I1の目標で、棚卸画面表示後に別出庫 | 古い絶対値で出庫を消さず競合再確認 | 承認I1。版／前後数量を照合、未実施 |
| AC-inventory-adjust-retry | I1の目標で、返却応答消失で再送 | 同じ実物返却を二重増加させない | 承認I1。操作参照と履歴を照合、未実施 |

## 未決事項・関連仕様・根拠

I1（最小管理単位の整数・不足出庫不可・入荷／廃棄／棚卸訂正／実物返却の履歴付き別操作）は2026-10-15に依頼者承認済み。質問 Sentinel_75a3a1bf81b481919ba4c34aa633feb4 の「3. 在庫とレジ締め」、未回答群の確認 Sentinel_8d0032ca45b881919b3d2ceb5339f792 に対する Sentinel_7a67fd30e47c819192cd36d416f0a2bf「ok」が根拠。履歴項目・競合・再送の意味契約は選定技術契約で確定し、現在の実装適合を承認済みとはしない。現行整数数量／在庫不足拒否、閾値0の誤記是正、非自動巻戻しは既存契約の整理。新しい物販連動・ロット管理・単位換算の自動化は別スコープ。技術差と薬剤同期の依存保護は[技術整理](../../../design/domains/finance-operations/technical-spec.md#inventory-current-gap)へ。
