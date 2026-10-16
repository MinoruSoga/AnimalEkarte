# 会計の用語・概念関係

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | ontology.accounting／語彙正本 |
| 決定状態 | [初回](../domains/accounting/lifecycle.md#accounting-agreement-first-five)＋[第二群](../domains/accounting/lifecycle.md#accounting-agreement-second)＋[第三群](../domains/accounting/lifecycle.md#accounting-agreement-third)＋[第四群](../domains/accounting/lifecycle.md#accounting-agreement-fourth)＋[第五群](../domains/accounting/lifecycle.md#accounting-agreement-fifth)＋[第六群](../domains/accounting/lifecycle.md#accounting-agreement-sixth)＋[第七群](../domains/accounting/lifecycle.md#accounting-agreement-seventh)＋[第八群](../domains/accounting/lifecycle.md#accounting-agreement-eighth)＋[第九群](../domains/accounting/lifecycle.md#accounting-agreement-ninth)＋[第十群](../domains/accounting/lifecycle.md#accounting-agreement-tenth)と[支払方法停止の追加合意](../domains/accounting/lifecycle.md#accounting-inactive-payment-method)・[品目別保険設定](../domains/accounting/lifecycle.md#requirement-fr-accounting-item-insurance-setting)の範囲。文書完成境界と実装判定外は設計へ |
| 実装状態 | 未確認。概念と現モデルを同一視しない |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-15 |

<a id="accounting-ontology"></a>
## 定義と別名

| 推奨語 | 意味・区別／例 | 業務正本 |
|:---|:---|:---|
| 未請求明細 | まだ請求していない診療・物販等。過去未請求分も選択可能。確定済み未納と別 | [対象選択](../domains/accounting/lifecycle.md#requirement-fr-accounting-owner-selected-charges) |
| 請求確定 | 請求額と対象を確定すること。入金なし／一部入金／明細あり0円でも成立。明細0件は確定不可。旧「精算」「決済確定」は文脈を確認 | [入金との区別](../domains/accounting/lifecycle.md#requirement-fr-accounting-receivable-partial-zero) |
| 入金・部分入金 | 確定請求へ受け取った金額を記録。後日回収は元請求への追加。複合決済とは別 | [入金との区別](../domains/accounting/lifecycle.md#requirement-fr-accounting-receivable-partial-zero) |
| 支払方法の利用停止（無効） | その方法による新受領の選択不可。過去支払記録・帳票参照・方法別集計の削除とは別。既存資金の配分・充当／解除を新受領にしない | [停止合意](../domains/accounting/lifecycle.md#accounting-inactive-payment-method) |
| 複合決済 | 一回の支払いで現金＋カード等を併用。旧画面の「分割決済」はこの用法。分割払い／後日回収とは区別 | [入金との区別](../domains/accounting/lifecycle.md#requirement-fr-accounting-receivable-partial-zero) |
| 未納・残額 | 確定した請求の未回収分。単純例：10,000円請求−3,000円入金＝7,000円。訂正後の未払・返金必要額を分ける。通常例は正本へ。取消後の再請求は元明細を選択確認し新番号で元取消請求へ関連付け | [入金との区別](../domains/accounting/lifecycle.md#requirement-fr-accounting-receivable-partial-zero) |
| 請求訂正／請求取消 | 金額・対象の訂正／請求の取消。原記録・理由・履歴を保持。実際の返金や在庫返却と別 | [訂正](../domains/accounting/lifecycle.md#requirement-fr-accounting-reasoned-financial-correction) |
| 返金 | 実際に返した金額。方法を指定。返金必要額／未実行返金と別で、請求取消や減額だけでは成立しない | [返金](../domains/accounting/lifecycle.md#requirement-fr-accounting-reasoned-financial-correction) |
| 預り金・前金 | 飼主単位で預かった額。予約参照は任意。自動失効なし・同医院／元飼主のみ。ホテル等の前金は預り金の用途。受領・未充当残／利用可能額・充当・実返金を区別 | [前金](../domains/accounting/lifecycle.md#requirement-fr-accounting-reservation-advance) |
| 預り金充当（前金充当） | 同飼主の会計へ、充当先と金額を確認して割り当てること。同飼主の別予約／別ペットも可。新しい入金ではない | [前金](../domains/accounting/lifecycle.md#requirement-fr-accounting-reservation-advance) |
| 預り金の残高復元 | 会計取消または減額で有効な充当を解除し元飼主の預り残へ戻す。有効充当額に解除・復元済み額を含めず、新受領・実返金とは別 | [預り金](../domains/accounting/lifecycle.md#requirement-fr-accounting-reservation-advance) |
| 超過値引きの警告 | 合算率100%超や商品額超の値引きを警告として示す。承認して進められ、実値引きは当該scope対象額まで。超過分を他行へ回さず、明細・請求額の下限は0円 | [計算](../domains/accounting/lifecycle.md#requirement-fr-accounting-calculation-snapshot) |
| 割引scope | 割合合算と定額適用の範囲。各明細と会計全体を別に扱い、明細→全体の順 | [順序](../domains/accounting/lifecycle.md#accounting-calculation-order) |
| 全体値引きの配賦 | 明細割引後の対象税抜額比で配ること。各配賦を切捨て、残円は剰余降順で調整。過大値引きの他行転用とは別 | [配賦](../domains/accounting/lifecycle.md#accounting-calculation-order) |
| 会計全体割引の除外品目 | 品目マスタで指定し、全体割引の対象額・配賦先から除外。明細割引不可・保険対象外・税区分と別 | [除外指定](../domains/accounting/lifecycle.md#accounting-insurance-and-exclusion) |
| 品目マスタの保険対象設定 | 診察・処置・予防・健診と既存薬剤・検査で保持する品目側の対象／対象外。全体割引除外・税・会計での確認修正値・資格／補償保証と別。新規4種別の新規登録初期値は保険対象外。既存明示値保持、旧欠落の一括解釈は未決 | [品目設定](../domains/accounting/lifecycle.md#requirement-fr-accounting-item-insurance-setting) |
| 取扱保険会社／商品 | 複数会社を扱う前提。アニコムは例で固定一覧ではない。資格／残限度・負担額はスタッフ確認・記録。商品ごとの自動給付算式は将来の対象 | [保険](../domains/accounting/lifecycle.md#accounting-insurance-and-exclusion) |
| 保険確認結果／保険負担額 | スタッフが外部画面等で確認した資格・残限度と今回保険が負担する額。未確認／不明は有効と別、窓口利用不可は補償対象外と別。保険会社の実入金ではない | [窓口保険](../domains/accounting/lifecycle.md#accounting-insurance-window-scope) |
| 今回利用する保険契約 | 保険を使うペットごとに1回の会計で1契約。同じ診療費への複数契約併用・給付調整は対象外 | [契約の単位](../domains/accounting/lifecycle.md#accounting-insurance-window-scope) |
| マスタ変更権限付きスタッフ | 対象の価格・割引・保険・支払方法マスタ変更を許可されたスタッフ。管理者以外も含み、全スタッフとは別 | [権限](../domains/accounting/lifecycle.md#requirement-fr-accounting-master-authorized-maintenance) |
| 税抜計算／税込表示 | 価格・値引きの内部算定は税抜、飼主向け料金は税込。表示用明細税込額と請求書の税率別税額を区別 | [計算](../domains/accounting/lifecycle.md#accounting-calculation-order) |
| 請求書の税率別税額 | 値引き・配賦後の税抜額を請求書／税率ごとに集約し、税額の1円未満を一度切り捨てる製品方針 | [税と根拠](../domains/accounting/lifecycle.md#accounting-tax-evidence) |
| 返金必要額／未実行返金 | 訂正後料金を超える請求への支払等から表示する返金必要額。実返金履歴とは別 | [訂正](../domains/accounting/lifecycle.md#requirement-fr-accounting-reasoned-financial-correction) |
| 請求書／領収書 | 請求する額と実際に受け取った額の帳票。飼主宛名・ペット別明細・ペットに紐づかない共通料金別枠。訂正／部分返金は元取引と今回変更額を区別し、再発行表示と訂正前履歴を保持 | [帳票](../domains/accounting/lifecycle.md#requirement-fr-accounting-document-and-aged-debt) |
| 締め後訂正差分 | 元締めを保持して追加する訂正情報。実際の入金・返金と区別 | [締め境界](../domains/accounting/lifecycle.md#requirement-fr-accounting-post-close-separate-movement) |
| 会計時点の計算根拠 | 会計で使った価格・手修正・scope別割引／要求額と実適用額・配賦・税区分／税率・税率別税額・算定版・保険等の値。現在マスタとは別 | [根拠保存](../domains/accounting/lifecycle.md#requirement-fr-accounting-calculation-snapshot) |
| 再請求 | 取消請求の元明細を選択確認し新しい請求番号／IDで作成、元取消請求へ関連付ける。元取引番号のままの再発行と別。元価格／税／割引を初期継承し必要箇所をスタッフ修正 | [再請求](../domains/accounting/lifecycle.md#accounting-rebill-and-collection) |
| 同時回収の請求別配分 | 新規会計と選択した過去未納の回収を同じ操作で扱い、今回の支払額を各元請求へ指定する。過去未納を新売上にしない | [回収](../domains/accounting/lifecycle.md#accounting-rebill-and-collection) |
| 共通料金 | ペットに紐づかない飼主帰属の料金。帳票でペット別明細と別枠に表示し、架空のペットを作らない | [帳票](../domains/accounting/lifecycle.md#requirement-fr-accounting-document-and-aged-debt) |
| 再請求の継承根拠 | 元会計snapshotの価格・税・割引を初期値にし、必要箇所だけスタッフが修正。現在マスタや元請求集約総額の無条件コピーと別 | [再請求根拠](../domains/accounting/lifecycle.md#accounting-rebill-snapshot-payment-source) |
| 取消元残入金の確認充当 | 既存の実入金の利用可能残を確認して新請求へ配分すること。新受領でも実返金でもない。復元済み預り金の再充当とは別 | [残入金](../domains/accounting/lifecycle.md#accounting-rebill-snapshot-payment-source) |
| 返金元資金内訳 | 返金する元資金が現金・カード・預り金のどれに由来するかと額。実際に返す方法とは別。不明なら記録確認・内訳と理由登録まで対象返金を確定しない | [返金元](../domains/accounting/lifecycle.md#accounting-rebill-snapshot-payment-source) |
| 空明細／0円会計 | 明細0件は確定不可。実際の明細あり0円は理由／根拠を残す | [第十群](../domains/accounting/lifecycle.md#accounting-empty-rounding-master) |
| 値引き前行額 | 単価×数量の積を行ごとに1円四捨五入。割合値引き／税丸めと別 | [行額](../domains/accounting/lifecycle.md#accounting-empty-rounding-master) |
| master明示的再適用 | 追加済み値を保持し、スタッフ指定対象へ最新masterを適用。自動追従・確定後直接上書きと別 | [再適用](../domains/accounting/lifecycle.md#accounting-empty-rounding-master) |

## 関係・多重度・制約

| 関係 | 方向・多重度 | 制約／状態 |
|:---|:---|:---|
| 品目保険設定→会計の対象区分 | 元品目種別／ID・医院／版と会計時点snapshotへ対応 | 明示値と個別確認修正値を区別。master更新で自動変更せず明示再適用／確定後訂正へ。対象設定だけで資格・補償・負担額を保証しない |
| 飼主→まとめた請求 | 一つの請求で対象飼主は一人、その飼主の複数ペット可 | 元ペット・診療等の明細情報を保持。共通料金はペット参照なしを保持。医院分離を維持 |
| 支払方法→新受領／過去支払記録 | 現在の選択可否と既存参照を分離 | 停止方法は新入金で選べず、過去履歴・集計の方法ID／医院対応と金額を保持。返金・誤記訂正へ新受領のgateを一律適用しない |
| 請求→元明細 | 確定時は選択明細1件以上 | 未請求と確定済み未納を混ぜて再請求しない。専用記録の内部は各owner |
| 請求→入金履歴 | 0件から複数件 | 0件でも確定可。後日回収も元請求へ。0円で架空入金なし |
| 一回の入金→複数請求への配分 | 元請求へ複数帰属。古い請求順、同日作成順。保存後も理由付き変更可 | 実入金額／日を再配分で変えない。現金超過は釣銭、非現金超過は返金必要。預り金へ無断転用しない |
| 預り金→飼主／予約 | 受領ごとの元飼主は必須、元予約は0または1の任意参照 | 同医院／飼主を保持。指定予約の所属を検査。同飼主の別予約／ペット利用可。自動失効なし、別医院移行・別飼主譲渡は今回対象外 |
| 前金→充当／実返金 | 受領後、有効充当額・取消／減額による解除／復元・実返金を履歴で追跡。未充当残の式は業務正本へ | 充当は新受領ではない。使用済み残を再充当／再返金しない |
| 訂正後料金→未払額／返金必要額 | 請求帰属の入金・実返金・前金充当から別々に表示 | 未実行返金は返金済みではない。未充当前金を請求入金へ混ぜない |
| 受領証／請求書／領収書／返金帳票→取引履歴 | 預り金受領証、最終会計に充当と今回入金、返金に額／方法／元取引番号。再発行でも元番号維持 | 再発行表示と旧帳票履歴を保持。番号書式等は未決。充当を新受領にしない |
| 入金済み額→返金可能額 | 実入金累計−実返金累計の範囲内 | 方法・権限・理由必須。元カード→現金等の別方法可、金額追加承認なし。決済自動連携／規約適合の承認ではない |
| 会計取消／減額→預り金の充当解除／残高復元 | 元飼主へ有効な充当分を一度戻す。減額の併用時はスタッフが内訳を選択確認。取消／減額・解除・復元の対応を保持 | 実返金は別。新受領・通常Undo再開なし。返金済み分の二重復元禁止 |
| 明細／会計全体→割引 | 両scopeの割引を利用可能 | 明細→全体、同scope割合合算・割合額四捨五入→定額。全体額は明細後の対象税抜額比で配賦。超過警告・対象額上限・他行転用なし。保険負担額は第七群のスタッフ確認記録 |
| 会計内のペット→今回利用契約 | 保険利用時に各ペット1契約。複数ペットはそれぞれ保持 | 同じ診療費の複数契約併用なし。保有契約数自体の制限ではない |
| 保険確認結果→窓口会計 | 確認結果・保険負担額を対象契約／ペット／明細へ対応 | 保険会社への請求・入金照合・査定差額管理は対象外。ローカル訂正から外部状態を自動変更しない |
| 請求書→税区分／税率別集約 | 配賦後の各行から複数区分／税率の集約を保持 | マスタ由来の会計時点の区分・税率を保持。課税区分を混同せず税率別1回切捨て。全品目一律税率や保険給付の課税標準控除を推定しない |
| 品目→会計全体割引除外指定 | 品目ごとに指定し、会計時点の根拠を保持 | 初期値／欠落時は除外しない。明示設定保持。除外行へ配賦せず全行除外でも元明細・税は残り、請求全体0円とは限らない |
| 帳票→飼主／ペット別明細・共通料金／変更履歴 | 宛名は飼主、元明細をペット別と共通料金別枠へ区分、元取引と今回変更額を追跡 | 訂正差額と実返金は別。元取引・旧帳票を上書きしない |
| スタッフ→マスタ変更 | 管理者または対象変更権限を付与されたスタッフ | 閲覧のみ・通常会計操作のみで許可しない。医院scope／現在認可を検査し、現grantの変更は未実施 |
| 元締め→訂正差分 | 元を保持し差分を追加 | 原請求日と実入出金日を区別、充当は再受領ではない。締め区分／照合内部D12 |
| 請求確定→予約／記録の確定歴参照 | 正確な対象リンクを保持。複数対象を単一IDへ潰さない | [通常Undo停止の既決境界](../workflows/reservation-to-record-flow.md#accounting-finalization-undo-boundary)。返金／取消後も保持 |
| 元取消請求→再請求 | 選択確認した元明細を新請求IDへ対応させ、元取消請求を参照 | 有効請求への同じ元明細の二重帰属を防ぐ。元取消・確定歴と金銭履歴は保持し、新入金を複製しない。残入金の確認充当と復元預り金の再充当を区別 |
| 同時回収→新請求と過去請求 | 選択した複数請求へ今回の支払を配分 | 各元請求ID／日付／残額を保持。古い順・同日作成順、スタッフ変更可。選択外への無断配分なし |
| 共通料金→飼主／ペット | 対象飼主は1人、ペット参照はなし | 同医院・元明細・税／割引根拠を保持。帳票の表示都合でペット／カルテ／保険契約へ付け替えない |
| 取消元の実入金→新請求の配分 | 元実入金参照と旧／新配分、利用可能残・確認額を追跡 | 自動移動せずスタッフ確認。元額／方法／実受領日を保持し、既返金・他配分・復元済み預り金を二重使用しない |
| 返金元資金内訳→返金確定 | 確認済み内訳と理由を元資金へ対応させる | 不明のまま確定不可。内訳登録自体は実返金でなく、現在上限・既返金／復元を再検査 |
| 元保険確認→再請求の確認結果 | 元結果を保持し、今回のペット／契約・対象明細で再確認して記録 | 元の有効判定を無条件に継承せず、外部保険状態を自動変更しない |

## 画面・モデルとの対応と確認範囲

[精算詳細](../domains/accounting/pages/accounting-settlement.md)の「未精算」「精算済み」と支払状態を無条件に同一視しない。現行の単一ペット／カルテ参照や全額入金制約は[現行差分](../../design/domains/accounting/technical-spec.md#accounting-current-gap)へ。API／DB識別子や画面ラベルの改名は今回行わない。

予約の[語彙正本](reservations.md)・カルテの[語彙正本](medical-records.md)を再利用し、確認医師／編集者・項目実施取消と会計操作を混同しない。今回の定義・関係・制約は[10FR／167AC](../domains/accounting/lifecycle.md#accounting-acceptance)へ照合する。初回〜第十群142AC・死亡後精算2ACの履歴を保持し、支払方法停止の追加7AC、会議連携4AC、今回の品目別保険設定1FR／12ACを区別する。第三群で飼主単位預り金／予約任意・同飼主別予約／ペット利用、割合合算→金額→割引後保険、同日作成順／理由付き保存後配分変更、非現金超過の返金必要、別方法返金・金額追加承認なし、受領／充当／今回入金／返金の帳票と元番号維持を追加。旧FR識別子のreservation-advanceは参照互換のため保持し、最新概念をAPI／コード改名としない。第四群で両粒度の割引・超過警告と承認後の請求額0円下限、会計取消時の残高復元と別返金を追加。有効充当額は取消・減額による解除・復元を反映し、未充当残との整合を照合する。第五群で明細→全体のscope分離と比例配賦、税抜計算／税込表示、割合値引き額四捨五入と請求書税率別税切捨て、超過の実値引き上限・他行転用なしを反映。同率配賦の固定順は技術案。第六群で複数保険会社、品目マスタの全体割引除外、飼主宛名・ペット別明細と元取引／今回変更額、管理者と権限付きスタッフのマスタ変更を追加。第七群で外部画面等によるスタッフの資格／残限度確認と結果・保険負担額の記録、窓口会計限定、ペットごと1利用契約を追加。会社／商品別の自動給付算式は将来の別スコープで現在の記録運用の前提にしない。第八群で減額時の解除・残高復元と併用時のスタッフ内訳確認、自動失効なし・同医院同飼主限定、取消元明細の選択確認・新番号再請求／元請求リンク、同時回収と元請求別配分、共通料金別枠を追加。督促・貸倒、別医院移行・別飼主譲渡は今回対象外。第九群で元価格／税／割引の継承と必要箇所のスタッフ修正、残入金の確認充当と保険再確認、返金元資金内訳と理由の登録まで返金確定不可を追加。元資金と実返金方法を区別し、新受領・二重使用や金額別追加承認を作らない。第十群で空明細確定不可／0円理由、値引き前行別四捨五入、除外初期値／欠落時false・追加後明示的再適用を合意。[全体の状態・master・帳票と受け手](../../design/domains/accounting/technical-spec.md#accounting-reconciled-contract)、[文書完成境界](../../design/domains/accounting/technical-spec.md#accounting-document-completion)へ対応。実装／利用開始条件は業務未決と分ける。

## 月次未納の追加接続

[月次帳票正本・追加3AC](../domains/accounting/pages/unpaid-list.md#unpaid-monthly-source-adoption)へ接続する。月末残高・当月発生未払い・実入金・翌月繰越を区別し、預り金充当と入金、減額と返金を同一化しない。期間指定・手動印刷の追加目標であり、元請求訂正や実収支の帰属は既存契約を継承する。2026-10-16の追加承認により、4列を保持し、当月入金は実受領のみ、預り金充当／解除・請求減額／取消は別掲の調整内訳へ対応する。調整増減・同一cutoffの残高連続性、既存資金の二重受領禁止は正本の3ACを更新して照合する。実返金等の詳細表示と実装検証は別であり、税額や金銭履歴を変更しない。
