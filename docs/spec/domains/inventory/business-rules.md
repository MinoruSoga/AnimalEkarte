# 在庫の業務契約

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.inventory.business-rules／業務仕様 |
| 決定状態 | 2026-10-20追加本人回答は末尾の該当節を参照。2026-10-17追加決定は末尾と進捗入口を参照。従来：既存記載を継承（決定根拠未確認の部分を含む）。I1の目標は依頼者2026-10-15承認（Sentinel_7a67fd30e47c819192cd36d416f0a2bf）。既存詳細全体の再承認ではない。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-20追加回答反映（従来確認2026-10-17／15） |

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

D03／D10の既決境界を継承。カルテ訂正・過去入力・患者転記・復元・請求取消・減額・返金だけでは在庫を自動増減しない。既存通常治療createのinventory ID指定時減算は旧経路の実装差分であり、下記の[出庫確定の本人回答](#inventory-outflow-confirmation-20261020)への適合を意味しない。過去入力や訂正へ流用しない。関連不明を影響なしとせず確認を残す。

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
| AC-inventory-movement-history | 入荷・廃棄・棚卸調整と、院内投薬実施／持帰薬・物販の実引渡し時に担当スタッフが出庫確定する。準備だけ・カルテ保存／請求確定だけと比較 | 理由／担当／実日時／前後数／元参照を記録。院内実施／実引渡しを確認した出庫を一度だけ反映し、準備・保存・請求だけでは減算しない | 承認I1。操作種別ごと照合、未実施 |
| AC-inventory-actual-return | I1の目標で、処置取消後に実物1個だけ返却 | 実物確認量のみ1回増加、取消だけでは0 | 承認I1。取消／返却結果を比較、未実施 |
| AC-inventory-concurrent-count | I1の目標で、棚卸画面表示後に別出庫 | 古い絶対値で出庫を消さず競合再確認 | 承認I1。版／前後数量を照合、未実施 |
| AC-inventory-adjust-retry | I1の目標で、返却応答消失で再送 | 同じ実物返却を二重増加させない | 承認I1。操作参照と履歴を照合、未実施 |

## 未決事項・関連仕様・根拠

I1（最小管理単位の整数・不足出庫不可・入荷／廃棄／棚卸訂正／実物返却の履歴付き別操作）は2026-10-15に依頼者承認済み。質問 Sentinel_75a3a1bf81b481919ba4c34aa633feb4 の「3. 在庫とレジ締め」、未回答群の確認 Sentinel_8d0032ca45b881919b3d2ceb5339f792 に対する Sentinel_7a67fd30e47c819192cd36d416f0a2bf「ok」が根拠。履歴項目・競合・再送の意味契約は選定技術契約で確定し、現在の実装適合を承認済みとはしない。現行整数数量／在庫不足拒否、閾値0の誤記是正、非自動巻戻しは既存契約の整理。新しい物販連動・ロット管理・単位換算の自動化は別スコープ。技術差と薬剤同期の依存保護は[技術整理](../../../design/domains/finance-operations/technical-spec.md#inventory-current-gap)へ。


### カルテ自動保存・通常編集・タブ復元の境界

[D03の会議追加合意](../medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)による入力クリア／履歴複製Undoは、実物使用・出庫・返却・棚卸履歴の自動逆実行ではない。保存済み治療記載を復元しても実物が戻ったと判定せず、数量補正はD13の既決の理由付き実績・現在版・実移動照合へ分ける。記録上の実施チェックや処方複写だけで新しい実物使用を確定しない。対象source／実績が不明なら照合し、同じ実移動を再適用しない。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../accounting/lifecycle.md#requirement-fr-accounting-billable-performance)。通常プランの未実施・予定・中止は当日請求外、持帰薬物販は内容数量確定・準備完了で対象にする。この準備は実引渡し・実出庫と別。請求取消・返金・プラン修正で在庫自動増減しない。トリミング等の個別中止料・入院ホテルの専用会計／終了取消は従来契約を維持する。

<a id="inventory-outflow-confirmation-20261020"></a>
## 実施・実引渡し時の出庫確定（2026-10-20本人回答）

決定者はユーザー。質問 Sentinel_07eb2535e184819190e1145f884b69cc に対する本人 Sentinel_6498c7a23a8c8191b672ba0dea52c1c3 の2026-10-20回答「3 そうですね」により採用。院内投薬は実施時、持ち帰り薬・物販は実際の引渡し時に、担当スタッフが出庫を確定する。カルテ保存・請求確定だけでは減らさず、小数の投薬量から在庫数を推測しない。請求対象になる内容数量確定・準備完了（ready）は[D10の請求条件](../accounting/lifecycle.md#requirement-fr-accounting-billable-performance)であり、実引渡し・出庫確定と別である。

既存FR-inventory-clinical-boundary／adjustmentの対象として、医院・元記録／実施又は引渡し・品目・管理単位・担当・実日時と履歴を対応付ける。最小管理単位の整数、不足時拒否、前後数量・元移動・理由／履歴、二重適用防止は既決を維持する。小数臨床量を整数へ丸めて減らしたり、未確認の換算係数を補ったりしない。単位対応が未確定なら出庫成功を表示せず照合へ渡す。既存create時自動減算はこの新目標と区別し、consumerの適合確認を残す。具体的UI・DB移行・薬剤／物販との対応設定を本回答から自動構成しない。

受入は既存AC-inventory-movement-history／fractional-use／insufficient／adjust-retryで、院内実施・持帰準備だけ・実引渡し、カルテ保存／請求確定だけ、単位不明、同一移動再送を区別して照合する（未実施）。
