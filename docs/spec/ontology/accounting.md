# 会計の用語・概念関係

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | ontology.accounting／語彙正本 |
| 決定状態 | [初回](../domains/accounting/lifecycle.md#accounting-agreement-first-five)＋[第二群](../domains/accounting/lifecycle.md#accounting-agreement-second)＋[第三群](../domains/accounting/lifecycle.md#accounting-agreement-third)＋[第四群](../domains/accounting/lifecycle.md#accounting-agreement-fourth)の範囲。細部の未決は業務正本へ |
| 実装状態 | 未確認。概念と現モデルを同一視しない |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-15 |

<a id="accounting-ontology"></a>
## 定義と別名

| 推奨語 | 意味・区別／例 | 業務正本 |
|:---|:---|:---|
| 未請求明細 | まだ請求していない診療・物販等。過去未請求分も選択可能。確定済み未納と別 | [対象選択](../domains/accounting/lifecycle.md#requirement-fr-accounting-owner-selected-charges) |
| 請求確定 | 請求額と対象を確定すること。入金なし／一部入金／0円でも成立。旧「精算」「決済確定」は文脈を確認 | [入金との区別](../domains/accounting/lifecycle.md#requirement-fr-accounting-receivable-partial-zero) |
| 入金・部分入金 | 確定請求へ受け取った金額を記録。後日回収は元請求への追加。複合決済とは別 | [入金との区別](../domains/accounting/lifecycle.md#requirement-fr-accounting-receivable-partial-zero) |
| 複合決済 | 一回の支払いで現金＋カード等を併用。旧画面の「分割決済」はこの用法。分割払い／後日回収とは区別 | [入金との区別](../domains/accounting/lifecycle.md#requirement-fr-accounting-receivable-partial-zero) |
| 未納・残額 | 確定した請求の未回収分。単純例：10,000円請求−3,000円入金＝7,000円。訂正後の未払・返金必要額を分ける。通常例は正本へ。取消後の再請求等は未決 | [入金との区別](../domains/accounting/lifecycle.md#requirement-fr-accounting-receivable-partial-zero) |
| 請求訂正／請求取消 | 金額・対象の訂正／請求の取消。原記録・理由・履歴を保持。実際の返金や在庫返却と別 | [訂正](../domains/accounting/lifecycle.md#requirement-fr-accounting-reasoned-financial-correction) |
| 返金 | 実際に返した金額。方法を指定。返金必要額／未実行返金と別で、請求取消や減額だけでは成立しない | [返金](../domains/accounting/lifecycle.md#requirement-fr-accounting-reasoned-financial-correction) |
| 預り金・前金 | 飼主単位で預かった額。予約参照は任意。ホテル等の前金は預り金の用途。受領・未充当残／利用可能額・充当・実返金を区別 | [前金](../domains/accounting/lifecycle.md#requirement-fr-accounting-reservation-advance) |
| 預り金充当（前金充当） | 同飼主の会計へ、充当先と金額を確認して割り当てること。同飼主の別予約／別ペットも可。新しい入金ではない | [前金](../domains/accounting/lifecycle.md#requirement-fr-accounting-reservation-advance) |
| 預り金の残高復元 | 会計取消で有効な充当を解除し元飼主の預り残へ戻す。新受領・実返金とは別 | [預り金](../domains/accounting/lifecycle.md#requirement-fr-accounting-reservation-advance) |
| 超過値引きの警告 | 合算率100%超や商品額超の値引きを警告として示す。警告後の0円下限／負額処理は未決 | [計算](../domains/accounting/lifecycle.md#requirement-fr-accounting-calculation-snapshot) |
| 返金必要額／未実行返金 | 訂正後料金を超える請求への支払等から表示する返金必要額。実返金履歴とは別 | [訂正](../domains/accounting/lifecycle.md#requirement-fr-accounting-reasoned-financial-correction) |
| 請求書／領収書 | 請求する額と実際に受け取った額の帳票。再発行表示と訂正前履歴を保持 | [帳票](../domains/accounting/lifecycle.md#requirement-fr-accounting-document-and-aged-debt) |
| 締め後訂正差分 | 元締めを保持して追加する訂正情報。実際の入金・返金と区別 | [締め境界](../domains/accounting/lifecycle.md#requirement-fr-accounting-post-close-separate-movement) |
| 会計時点の計算根拠 | 会計で使った価格・手修正・割引・保険等の値。現在マスタとは別 | [根拠保存](../domains/accounting/lifecycle.md#requirement-fr-accounting-calculation-snapshot) |

## 関係・多重度・制約

| 関係 | 方向・多重度 | 制約／状態 |
|:---|:---|:---|
| 飼主→まとめた請求 | 一つの請求で対象飼主は一人、その飼主の複数ペット可 | 元ペット・診療等の明細情報を保持。医院分離を維持 |
| 請求→元明細 | 選択した明細を複数参照 | 未請求と確定済み未納を混ぜて再請求しない。専用記録の内部は各owner |
| 請求→入金履歴 | 0件から複数件 | 0件でも確定可。後日回収も元請求へ。0円で架空入金なし |
| 一回の入金→複数請求への配分 | 元請求へ複数帰属。古い請求順、同日作成順。保存後も理由付き変更可 | 実入金額／日を再配分で変えない。現金超過は釣銭、非現金超過は返金必要。預り金へ無断転用しない |
| 預り金→飼主／予約 | 受領ごとの元飼主は必須、元予約は0または1の任意参照 | 同医院／飼主を保持。指定予約の所属を検査。同飼主の別予約／ペット利用可、別飼主移管は未承認 |
| 前金→充当／実返金 | 受領後、充当・実返金を履歴で追跡。未充当残を保持 | 充当は新受領ではない。使用済み残を再充当／再返金しない |
| 訂正後料金→未払額／返金必要額 | 請求帰属の入金・実返金・前金充当から別々に表示 | 未実行返金は返金済みではない。未充当前金を請求入金へ混ぜない |
| 受領証／請求書／領収書／返金帳票→取引履歴 | 預り金受領証、最終会計に充当と今回入金、返金に額／方法／元取引番号。再発行でも元番号維持 | 再発行表示と旧帳票履歴を保持。番号書式等は未決。充当を新受領にしない |
| 入金済み額→返金可能額 | 実入金累計−実返金累計の範囲内 | 方法・権限・理由必須。元カード→現金等の別方法可、金額追加承認なし。決済自動連携／規約適合の承認ではない |
| 会計取消→預り金の充当解除／残高復元 | 元飼主へ有効な充当分を一度戻す。取消・解除・復元の対応を保持 | 実返金は別。新受領・通常Undo再開なし。返金済み分の二重復元禁止 |
| 明細／会計全体→割引 | 両scopeの割引を利用可能 | 併用順序／集約範囲・税／保険対象配賦と超過警告後金額は未決 |
| 元締め→訂正差分 | 元を保持し差分を追加 | 原請求日と実入出金日を区別、充当は再受領ではない。締め区分／照合内部D12 |
| 請求確定→予約／記録の確定歴参照 | 正確な対象リンクを保持。複数対象を単一IDへ潰さない | [通常Undo停止の既決境界](../workflows/reservation-to-record-flow.md#accounting-finalization-undo-boundary)。返金／取消後も保持 |

## 画面・モデルとの対応と確認範囲

[精算詳細](../domains/accounting/pages/accounting-settlement.md)の「未精算」「精算済み」と支払状態を無条件に同一視しない。現行の単一ペット／カルテ参照や全額入金制約は[現行差分](../../design/domains/accounting/technical-spec.md#accounting-current-gap)へ。API／DB識別子や画面ラベルの改名は今回行わない。

予約の[語彙正本](reservations.md)・カルテの[語彙正本](medical-records.md)を再利用し、確認医師／編集者・項目実施取消と会計操作を混同しない。今回の定義・関係・制約は8FR／66ACへ照合する。第三群で飼主単位預り金／予約任意・同飼主別予約／ペット利用、割合合算→金額→割引後保険、同日作成順／理由付き保存後配分変更、非現金超過の返金必要、別方法返金・金額追加承認なし、受領／充当／今回入金／返金の帳票と元番号維持を追加。旧FR識別子のreservation-advanceは参照互換のため保持し、最新概念をAPI／コード改名としない。第四群で両粒度の割引・超過警告、会計取消時の残高復元と別返金を追加。税・端数／超過警告後金額、両粒度併用順序／配賦、減額時の充当調整・医院間／別飼主等の残件は[業務正本](../domains/accounting/lifecycle.md#accounting-open-decisions)へ。全マスタ・帳票の語彙網羅は未完了。
