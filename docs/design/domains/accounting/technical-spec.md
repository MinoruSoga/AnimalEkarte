# 会計初回〜第十群と追加合意の技術仕様

読取基準：現行実装の観察はPR #670の41fcf3eb0a8befb7c0b5f613d1df749a60e30be8（起点main 7309b3a）時点。後続mainの修正有無・実装適合を断言せず、承認済みの目標契約と区別する。支払方法停止に関する追加読取は下記の専用行にheadと日付を明記し、既存観察全体の再確認とは扱わない。

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | design.accounting.partial／技術設計 |
| 決定状態 | 2026-10-17追加決定は末尾と進捗入口を参照。従来：業務10FRは合意済み（新規4種別の新規登録初期値は保険対象外。既存欠落値の解釈・変換は後続移行課題）。本書の実現方式は提案設計、未決業務を補完しない |
| 実装状態 | 未確認。静的に確認した現行制約は下記。API wire／DB正本は変更なし |
| 実機検証状態 | 未実施。移行・並行性・性能も未検証 |
| 文書確認日 | 2026-10-17追加差分（従来確認2026-10-15） |

[業務・AC正本](../../../spec/domains/accounting/lifecycle.md)と[語彙](../../../spec/ontology/accounting.md)を参照。対象は初回5点〜第十群と追加の死亡後精算・[支払方法停止](#accounting-inactive-payment-design)・[品目別保険設定](#accounting-item-insurance-design)のD10文書設計。完成境界は[末尾](#accounting-document-completion)へ。実装適合／利用開始を表さない。

<a id="accounting-current-gap"></a>
## 現行実装と確定業務の差分

| 調査対象 | 現行の静的確認 | 目標業務との相違／残件 |
|:---|:---|:---|
| [確定入力・元カルテ照合](../../../../backend/internal/billing/accounting_complete.go) | ヘッダは単一ペット／カルテ。診療項目が複数カルテにまたがる場合は拒否 | 同飼主の複数ペット・当日／過去の選択を表現する明細参照へ設計が必要。単一患者への付替えは禁止 |
| [確定transaction](../../../../backend/internal/billing/accounting_complete_tx.go)・[支払UI](../../../../frontend/src/features/accounting/components/PaymentCard.tsx) | 支払内訳を請求に一致させる経路。初回部分入金は拒否 | 未納／部分入金確定・後日入金の業務は現行制約と異なる。[既存S08](../../../ops/testing/scenarios/S08-accounting-corrections.md)を新仕様の合格根拠にしない |
| [支払方法UI](../../../../frontend/src/features/accounting/components/PaymentCard.tsx)・[安定識別の解決](../../../../backend/internal/billing/accounting_service_builders.go) | 2026-10-15のPR #670 head 82aa52f3c6d1ce00489a444d903748001670c21cの読取では、UIは固定4種を列挙し有効状態を受け取らない。解決関数は医院／system_key／IDの対応を検査するが当該関数に有効状態の検査はない | [新受領の停止条件](#accounting-inactive-payment-design)は追加目標。固定4種・安定識別の現行説明を停止無視の将来方針にせず、全writer／集計の適合済みとも扱わない |
| [返金UI](../../../../frontend/src/features/accounting/components/RefundSection.tsx)・[返金処理](../../../../backend/internal/billing/refund_service_tx.go) | 理由／方法は任意。請求金額等を基準とする上限経路がある | 理由／方法の必須化、入金累計−既返金累計の上限検査が必要。外部決済の実返金との照合は未確認 |
| [会計計算](../../../../frontend/src/lib/calculations.ts)・[帳票税内訳](../../../../frontend/src/features/accounting/lib/tax-breakdown.ts) | 四捨五入と切捨てを使う箇所が異なる | 第五群の税抜入力計算・割合額四捨五入と請求書／税率単位の税切捨て、および税込入力額維持へ整合設計が必要。現行式の再検証・税務適合全体は未確認 |
| [割引生成](../../../../backend/internal/billing/billing_item_service_create.go) | 飼主／キャンペーンの最大割引を選ぶ経路 | 併用・割合→金額の合意と相違。割合合算→金額→割引後保険の目標へ。両粒度の割引利用／超過警告と承認後の請求額0円下限へ追従。第五群の明細→全体・比例配賦・税／端数へ追従が必要 |
| [未納期間ビュー](../../../spec/domains/accounting/pages/unpaid-list.md) | 期間内に非取消売上のある飼主に対象を限定する既存記載 | 今月売上なしでも過去未納飼主を表示する合意と相違。filter／基準日の詳細は設計残件 |

API wireの正本は[現行API](../../../../backend/docs/api.yaml)、DBはbackend/migrations。本書の目標データは既存DTOや保存schemaが対応済みとの宣言ではない。全grant、旧値、各書込経路、帳票履歴、外部機器は未確認。

<a id="accounting-operation-contract"></a>
## 提案する保存・照合契約

- 請求ヘッダの飼主／医院と、行ごとの元ペット（共通料金では紐づけなし）・診療／専用記録・元明細IDを区別する。選択した未請求明細と既請求への回収参照を別に保持し、同じ元明細の二重請求を防ぐ。複数対象の正確な予約／記録リンクと確定イベントを保持する。
- 請求確定・追加入金・訂正・取消・返金を別intentとして扱う。対象ID群・期待版・依存参照・操作識別・actor・サーバー時刻、操作に必要な理由／返金方法を保持する。現在の医院／飼主所属・認可・請求状態・入金／返金累計を保存直前に再検証する。
- 元請求の不変な履歴と訂正／入金／返金を追跡し、請求額・入金履歴・残額の表示をそれぞれ取得する。訂正後料金に対する未払・返金必要を別projectionで表示し、実返金後に更新する。通常請求の例は業務ACへ従う。古い請求から初期充当し、スタッフ変更を元請求別の配分履歴に保存する。原請求日を初期順の根拠とし同日は作成順へ対応する設計。同一作成時刻は一意な作成順識別で安定させる技術案。保存後も理由と元／新配分履歴を保持して変更する。現金超過は釣銭、非現金超過は実入金と請求充当を分けて返金必要額へ。会計取消時の預り金残高復元と別返金は第四群へ従う。第八群の減額時復元・再請求・同時回収へ従い、価格等の継承・旧入金の確認充当・資金内訳不明時の保留は第九群へ従う。督促・貸倒管理は今回対象外。
- 保存する会計時点の根拠は価格・数量・元値／手修正値・適用割引・保険等と計算結果。現在マスタの再計算を過去会計表示の正本にしない。第五・第六群の合意範囲と規則版を保持し、未確認保険を有効にせず、除外初期値／欠落時falseは第十群へ従う。スタッフ確認結果・今回の保険負担額を会計時点の根拠へ対応させる。請求書と実受領の領収書を別projectionとし、再発行表示・訂正前帳票の履歴を保持する設計。発行時の値／対象版を保持し現在マスタから旧帳票を再生成しない。保存形式・番号書式等は残件だが、各取引の元番号を再発行で維持することは合意済み。
- 状態変更と必須の履歴・監査、依存ownerの確定参加を一体で成立させる設計へ。[予約の全writer gate／取得順](../reservations/technical-spec.md#reservation-global-lock-contract)と[受け手契約](../reservations/technical-spec.md#reservation-receiving-contracts)を維持する。複数対象での取得順・transaction範囲・旧writer参加は設計／検証残件で、現行separate-txをsame-tx済みとしない。
- 同じ操作の再送は同じ内容・対象の結果に照合し、二重入金／二重返金を作らない。payload相違・版競合・認可失効は成立させない。通信切断で結果不明ならreceipt照会／同一操作の再送で確認し、新キーの作り直しや残額0・失敗確定への推測変換を避ける。追加入金等の現行endpoint対応・永続operation receiptは未確認で、新wireを本書だけから決定しない。
- 返金累計上限は最新の実入金・既返金を競合保護下で再検証。指定方法と実際の返金記録を保持する。元カードから現金等の別方法を記録・操作でき、金額による追加承認は設けない。操作権限・理由・方法と総額上限を検査する。旧支払方法ごとの上限経路が別方法返金を一律拒否しないよう実装差分を照合する必要がある。決済サービスの実行・規約適合は未確認で、自動連携を追加しない。
- 締め後訂正は元close参照と差分、実際の入金／返金の参照を区別し、元締めを更新しない。原請求日と実受領／実返金日、前金充当日を別に保持する。今日の実入出金は今日の移動、充当で新たな実受領を生成しない。締め区分・会計差分と入出金の突合はD12と照合。金額だけの変更から金銭移動を生成しない。

カルテconfirmation ID＋record IDとaccounting IDを混同しない。[カルテの受渡し](../medical-records/technical-spec.md#medical-record-reconciled-contract)から対象／版・理由・影響確認を受け、会計ownerで現在検査と結果receiptを返す。臨床の保存成功は金銭処理成功を表さない。[カルテ保存§2](../../../spec/domains/medical-records/lifecycle.md#hybrid-sync)の送信先・保存単位を変えない。取消／返金・0円・未納でも不変の確定歴を保持し、通常Undoを再許可しない。請求取消を返金／在庫返却の自動intentにしない。

<a id="accounting-second-design"></a>
## 第二群から第三群へ更新した契約（提案設計）

2026-10-23、初回予約指定なし・新しい実受領1件全額のbackend受領と同口座再読込は[EMR-464限定技術契約](owner-deposit-initial-receipt.md)で選定済み。本節の従前の未選定/提案記述をこの限定範囲へ誤適用せず、当該契約を参照する。共有供給・実装・生成・試験・DB適用・利用開始は未完。506一般公開wire/全fund CAS・509締め設計は本選定に含めず、既存Complete互換と他の未選定範囲を維持する。

- 預り金受領を請求確定／入金配分と別intentにする。医院・元飼主は必須、予約参照はnullableで任意。受領額・方法・実受領日時・actor・操作識別と、充当先請求・任意の予約／ペット・額／日時、実返金額／方法／日時を別履歴として保持する。予約なしでも受領でき、同飼主の別予約／別ペット会計へ確認付きで利用する。指定予約がある場合のみ元飼主／医院との所属を検証し、nullを推測予約IDで埋めない。受領で予約完了／請求確定イベントを生成しない。
- 未充当残は受領累計−有効充当額−未充当分の実返金累計で管理する設計。有効充当額は充当累計−会計取消または減額で解除・残高復元済みの額とし、取消前の累計充当額をそのまま差し引かない。前金充当は請求への支払充足として参照するが、現金／カード等の新受領イベント・二度目の締め入金・二度目の受領領収書を作らない。充当済み分を請求から実返金するときは請求側の履歴へ、未充当分の実返金は前金残へ対応させ、同じ返金を両方から控除しない。第四群で会計取消時の充当解除・元飼主残高復元を合意。実返金は別操作で、既実返金／既復元を二重に戻さない。第八群で減額時の解除・復元と併用時の内訳選択確認を追加。既実返金の出所不明時は第九群の記録確認・内訳／理由登録へ従い、推測で復元しない。
- 5,000円受領→8,000円請求へ5,000円充当では未払3,000円・未充当0円。5,000円受領→確認済み最終料金3,000円なら3,000円充当・未充当2,000円。未実行返金は残を消さず返金済みとしない。同一前金への充当／返金を競合保護し、同一操作の再送で二重使用しない。異医院・別飼主の移管を成立させない。同飼主の別予約／別ペットへ利用する際は、元飼主・任意の元予約を保持して充当先と金額・期待版を再検証する。複数充当先の配分を明示し、合計使用額を最新利用可能額以内にする。
- 通常請求の未払と返金必要は訂正後料金と請求帰属の支払充足から計算する別表示。全額10,000円入金→9,000円訂正では未払0・返金必要1,000、3,000円入金→9,000円訂正では未払6,000・返金必要0。現金・カード等の実入金に基づく支払充足や実返金履歴を、返金予定だけでは減らさない。預り金の確認済み充当解除は第八群に従い、有効充当と請求の支払充足へ反映する。返金後の最新状態を照合して二重返金を防ぎ、支払方法の誤記訂正は金額／実授受日を動かさない。 [2026-10-23の限定承認](../../../spec/domains/accounting/lifecycle.md#accounting-cash-method-correction-unknown-pair-20261023)により、実受領額既知でカード等→現金へ訂正するとき、当時の預かり/釣銭は両方不明nullのまま理由必須で保持できる。原記録/訂正履歴を残し、0自動補完・新金銭移動・実受領額変更・方法だけの訂正による日時変更をしない。新受領の現金入力要件は緩和しない。詳細wire/reader/CAS/死亡後adapter等は[506未採択候補](../../../work/emr/emr-506-fund-wire-design-candidates-20261022.md)の設計残で、本承認だけで完成にしない。
- 古い請求への初期配分・スタッフ変更・釣銭を保存前previewと保存後receiptで照合する。選択請求の残額・版と配分合計を一体で検証し、別操作の入金と競合したら旧配分で過大充当しない。預り金受領との操作種別を分け、釣銭や非現金超過を預り金へ自動転用しない。再配分は元の実入金と実日付を保持し、請求訂正・既返金・他の配分を現在版で照合する。元配分の取消と新配分・関連請求の残額・理由／監査を原子的に保存し、二重受領や片側変更を作らない。
- 割合合算→金額割引→割引後の保険対象額という段階と、スタッフが確認修正した保険対象・割合・上限を計算根拠へ保存する。明細／会計全体の両方を利用でき、100%超や商品額超は警告し、承認して進めた請求額は0円を下限とする。明細→会計全体・比例配賦と税丸めは第五群で更新。保険は第七群のスタッフ確認・負担額記録へ従い、会社／商品別の自動給付算式は今回の対象外。除外初期値／欠落時falseは合意済み、実移行は後続で、既存計算式を新業務の承認済み算式にしない。
- 今月売上なしの過去未納飼主も含める読取projectionへ。全ページの集計と表示対象を整合させ、取得失敗を未納0に変換しない。帳票発行・再発行は対象請求／受領／訂正版と発行履歴を照合し、訂正前の記録を保持する。再発行操作と新入金を同じintentにしない。

前金の金額／キャンセル料／受領時期はD05のスタッフ個別確認入力へ従う。日数自動課金・取消時自動請求／返金は行わない。実入出金と充当の集計はD12／D16、操作権限はC04とD10で照合。新endpoint・DDL・payment methodの実装対応は未確認で、本書から実装／移行を開始しない。ホテル等の元予約参照を受けるが、D05の内部完了条件・予約state・臨床確認を再定義しない。

<a id="accounting-third-design"></a>
## 第三群の預り金・再配分・帳票の補完（提案設計）

預り金の現在残は受領累計−有効充当額−未充当分の実返金累計を履歴から追跡する。有効充当額は充当累計から会計取消または減額による解除・残高復元済み額を差し引き、復元を残高へ反映する。請求側からの実返金を同時に預り金残から控除しない。利用可能額は保存直前の最新残と既に処理中の使用を照合して算出し、未実行返金の表示だけで残を消さず、同じ額を返金と別会計の両方へ成立させない。実現方式はowner／資金参照と関連請求を既存gate／整列lock下で検査し、状態・履歴・監査・operation receiptを原子的に保存する案。確保状態や保存schemaは未実装・未検証で、予約IDを飼主残高の排他単位の代用にしない。

充当確認は飼主・預り金元取引・対象請求／ペットと任意の予約・金額・資金版／請求版を含む。確認後に残高・所属・請求訂正が変われば古い確認でcommitしない。再送は同一操作へ照会し、別keyで二重充当／二重返金を作らない。同じ飼主の別予約／別ペット利用を旧予約一致必須のvalidationで拒否しない。第八群で同医院／同飼主のみ・自動失効なしと確定し、別飼主の所有者付替えや医院間残高移送は今回対象外。予約や飼主の削除／変更時も履歴参照を失わず、具体的廃棄・統合手順は別ownerの残件。

非現金超過は実受領5,000円・請求充当3,000円・未実行返金必要2,000円を別projectionへ記録する設計。返金処理後に実返金履歴と必要額を更新し、釣銭・支払額の誤記訂正・飼主預り金の受領を代用にしない。異なる返金方法でも元取引・実返金方法・actor・理由・実日付・総額上限を追跡する。金額による承認段階を追加せず、サーバーの返金権限検査は維持する。

各受領・充当・入金・再配分・実返金は一意な操作／取引参照で追跡し、帳票には用途に応じた取引番号と元番号を表示する。再発行は元取引番号のまま、再発行回数／時刻・版／actor・出力内容の履歴を追加する案で、二度目の金銭取引を作らない。番号の書式・採番方式・保存形式は技術残件、法令上の採番要件は未確認。画面へ返すoperation receiptと飼主へ渡す受領証／領収書は別概念。

明細と会計全体の両割引を許可し超過は警告へ。第五群で各scopeの実値引き額を対象額までに制限し、明細／請求額とも負額にせず、超過分を他行へ回さないと確定した。入力した率や定額を上書きせず、要求額・実適用額を分けて保持する。保険対象外と税区分は行の根拠を保持し、全額を一つの保険基準／税区分に潰さない。警告を承認して進められることと0円下限をpreview／確定結果へ反映し、過去の承認済み会計snapshotを現在の算式で置換しない。

<a id="accounting-fourth-design"></a>
## 第四群の割引・取消残高復元（提案設計）

明細割引と会計全体割引を別scopeとして保存し、各値・対象・根拠を保持する。第四群時点で未決だった適用順序・割合合算範囲・配賦・税／端数は第五群へ従う。超過をwarningとして表現し、100%超／金額超だけを入力エラーcodeにしない。表示・計算preview・確定結果を区別し、元の割引条件と実適用額・0円となった根拠を保持する。負額の請求や架空の入金／返金を作らず、税務適合や新たな必須承認操作を本案から導かない。

会計取消intentに元充当参照・元飼主・対象請求／預り金と期待版・理由・actor・operation参照を渡す。既存gate／整列取得の下で現在の有効充当と既実返金／既復元を検査し、取消・充当解除・元飼主残高への復元・必要履歴／監査・結果receiptを原子的に保存する案。元受領と充当履歴は消さず、解除／復元を追記して有効充当額を減らし、未充当残と利用可能額を更新する。新受領や実返金を生成しない。

未返金5,000円受領・3,000円充当・残2,000円なら取消で有効充当額0円・残5,000円へ。未充当分を1,000円実返金済みなら、同じ3,000円の解除・復元後もその実返金を控除し、残4,000円とする。繰り返し取消・同一再送で二度復元せず、同時充当／返金と現在残を照合する。既実返金と復元の二重計上を防ぎ、減額時は第八群のスタッフ確認内訳に従い、既実返金の出所不明時は第九群の記録確認・内訳／理由登録と現在残の照合へ従う。取消後も不変の会計確定歴・原請求／予約リンクを保持し通常Undoを再開しない。復元された残からの実返金は別の理由・方法・権限付き操作、復元成功を実返金成功にしない。

<a id="accounting-fifth-design"></a>
## 第五群の計算再現と税率別snapshot（提案設計）

業務算定は[第五群の計算契約](../../../spec/domains/accounting/lifecycle.md#accounting-calculation-order)を正本とする。以下はその再現性を保つ実現案で、API／DBの新規定義や実装済みの宣言ではない。

- 税抜入力の値引き前の確定した税抜明細額を入力とし、各明細scope内で割合を合算、割合値引き額を1円単位に四捨五入（非負額の0.5円は切上げ）、定額値引きを加えた要求額と対象額までの実適用額を別に求める。明細後の対象額合計に同じ手順で会計全体scopeの実値引き額を求める。異scope間の割合合算はしない。元行額は第十群の単価×数量を行別に1円四捨五入して求め、割合値引き・配賦・税とは別段階で保持する。
- 全体配賦は、明細後対象額を A_i、合計を S、全体実値引き額を D として D×A_i÷S を求める。Sが0ならDと全配賦は0。Sが正なら各商の整数部をまず配り、Dから整数部合計を引いた残りの円を剰余の降順で1円ずつ配る。0円行へは配らず、配賦合計D・各配賦額がA_i以下・各残額非負を照合する。
- **同じ剰余のときは、計算snapshotに凍結した明細順（安定した行IDに対応する連番）を昇順に使う技術案**。同じ入力・版であれば表示の並替え・再送で受取行が変わらない。例：A・B・C各1円、全体1円引きは同率なので凍結順A、B、Cなら1、0、0円。これはユーザーが業務上の優先行を指定したという意味ではない。
- 円額・割合・税率を十進固定精度／有理数として扱う設計とし、浮動小数の誤差で0.5境界や剰余順位を変えない。配賦商は整数除算・剰余で比較できる。採用する型・桁数・限界値は実装設計時に検証する。
- 配賦後の明細を会計時点の税区分・適用税率でグループ化し、請求書／税率単位の税抜額に税率を掛けて税額を1回切り捨てる。非課税等は区分を保持し、課税行や単なる0%率へ潰さない。行単位の表示用税込額は請求書の税率別税額の正本にしない。税抜価格と飼主向け税込表示の対応を保存し、表示計算と請求書集約の差を照合できるようにする。
- snapshotには元行参照・凍結明細順、税抜元額／手修正、各scopeの割合・定額・要求額／実適用額／超過額、警告を承認した根拠、全体配賦基準／剰余・調整先／結果、税区分・税率・集約額／税額と丸め規則、保険対象区分・スタッフ確認修正値、算定版・出力結果を保持する案。税率マスタ・価格マスタ変更後も保存済み結果を保持する。訂正は元版を消さず新旧snapshotと理由・actor・操作参照を監査へ結ぶ。
- previewと確定で同じ算定版／入力を照合し、サーバー側で対象額・実値引き上限・配賦合計・現在版を再検査する。保存・計算履歴・audit_logs・operation receiptを原子的に成立させる案。警告承認だけを金額整合の保証にせず、再送や同時更新で違う配賦結果を保存しない。
- 保険適用対象は配賦後の行から追跡し、保険負担額は第七群のスタッフ確認結果を記録する。会社／商品別の税抜／税込基準・税額との関係・上限／端数を独自に自動算定するengineは今回の対象外。保険額で税の課税標準を減らす設計や旧式の流用を確定しない。[公的根拠と製品選択](../../../spec/domains/accounting/lifecycle.md#accounting-tax-evidence)の区別を保持し、税率別切捨てを法令上唯一の方法と説明しない。

[計算例](../../../spec/domains/accounting/lifecycle.md#accounting-calculation-examples)に加え、同率剰余の固定順、配賦対象0円、過大な割合＋定額、マスタ更新・表示並替え・再送後の同じ結果を検証対象にする。これらの実行検証、税務分類の確認、DTO／schema・旧会計との互換は未実施。

<a id="accounting-sixth-design"></a>
## 第六群の保険・除外指定・帳票・マスタ変更（提案設計）

[業務正本](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-sixth)の実現案。1FR・14ACを追加したが、API／DB・権限設定・runtimeは変更していない。

- 保険会社／商品・今回利用する契約を会計時点の根拠から識別できる参照として保持する案。アニコムを固定値や会社一覧の正本にしない。第七群のスタッフ確認・記録運用に、全商品一覧や自動給付算式の完成を要求しない。既存の補償率手選択を将来の自動設定禁止に読み替えず、自動設定・オンライン請求を実装済みとも扱わない。
- 品目マスタの全体割引除外を元行と会計snapshotへ対応させ、明細割引後の対象行だけから全体対象額と配賦を求める。除外行の配賦は0円、全行除外でも除外行金額と税は残る。保険対象外・税区分・明細割引を別に保持する。初期値／欠落時はfalse、追加済み行は明示的再適用。物理field・移行手順は後続実装へ。現行項目の保険対象外を代用せず、現行モデルに対応済みと宣言しない。
- 飼主宛名と元ペットごとの明細区分、ペットに紐づかない「共通料金」別枠を出力snapshotへ対応させる。訂正は元請求版と今回増減、部分返金は元入金と今回の実返金を別projectionにし、元番号・対象版・方法・実日付・履歴へ接続する。金額訂正から実返金を生成しない。共通料金に架空のペットを割り当てず元飼主・元明細を保持する。紙面のレイアウト／改ページ・保存形式・発行権限は設計残件。過去帳票を現在マスタで再生成して上書きしない。
- マスタ編集の対象とactor・選択医院・現在の操作grantを保存直前に再検証し、変更前後・時刻を監査へ結ぶ。権限失効や他医院参照を旧UI表示で通さない。現行の共用resourceのままで業務上の編集者制限を満たすかは確認が必要で、既存会計権限を新マスタ編集許可へ自動転用しない。権限グループ付与自体は別操作・別責務。

<a id="accounting-master-permission-map"></a>
### 現行マスタ権限の読取mapping（新業務の適合判定ではない）

基準head 2fb7cdc92249c5f8a3266a0aa184c3f8a8bba52aのroute／定義を読み取った対応。表の識別子を今回変更する指示ではない。実際のスタッフ割当・稼働設定・全writerは未検証。

| 対象 | 現行resource／action | 根拠・新業務への残件 |
|:---|:---|:---|
| 診療・薬剤等の価格 | master-medical、定期健診種別はcheckups。更新はedit | [診療master routes](../../../../backend/internal/medicalrecord/routes_masters.go)。価格専用resourceではなく項目種別で分かれる |
| 入院プラン価格 | master-hospitalization、更新はedit | [同routes](../../../../backend/internal/medicalrecord/routes_masters.go)。料金以外の項目編集との対応も残件 |
| 商品価格 | master-merchandise、更新はedit | [物販routes](../../../../backend/internal/inventory/routes.go) |
| トリミング価格 | master-trimming、更新はedit | [施術master routes](../../../../backend/internal/trimming/routes.go) |
| 割引キャンペーンマスタ | accounting。作成create、更新／並替edit、削除delete | [billing routes](../../../../backend/internal/billing/routes.go)。通常会計と共用のため、accounting:edit保有を新業務の明示的なマスタ編集許可と同一視しない |
| 割引値そのものの操作 | discount。初期の非0値はcreate、既存値変更はedit | [割引値検査](../../../../backend/internal/httpapi/discount_permission.go)。キャンペーンマスタのresourceとは異なる |
| 保険マスタ | master-insurance、CRUD対応action | [billing routes](../../../../backend/internal/billing/routes.go)。会計上の保険確認修正とは別 |
| 支払方法マスタ | master-payment-method、CRUD対応action | [billing routes](../../../../backend/internal/billing/routes.go)。固定4種の現行UIと将来の変更方針は別 |

[権限設計](../../../architecture/auth.md)の現行説明では、医院管理者は別role flagではなく権限グループのprofile。選択医院の有効グループのgrantをORで評価し、grantがなければ拒否する。system-adminの例外も医院境界を越える新承認ではない。[医院作成時default](../../../../backend/internal/clinic/clinic_service.go)では一般は上記master／accountingのview中心、執行はCRUD中心（支払方法はview／edit）だが、作成defaultは実配属の証拠ではない。職種名・既定グループ名だけで対象マスタの変更を許可しない。

[保険マスタ](../../../spec/screens/settings/master-insurance.md)の補償率自動設定なし、[支払方法](../../../spec/screens/settings/payment-methods.md)の固定4種は既存挙動の説明。承認された将来ポリシーへ昇格させず、支払ID整合の[ADR003](../../../architecture/adr/003-payment-method-identity-and-consistency.md)は保持する。商品分類から税率を断定する旧例も採用せず、品目の税務分類確認は医院へ残す。

<a id="accounting-seventh-design"></a>
## 第七群のスタッフ確認・窓口保険（提案設計）

[承認された窓口保険の境界](../../../spec/domains/accounting/lifecycle.md#accounting-insurance-window-scope)を実現する案。外部画面等で確認した結果と保険負担額を記録し、会社／商品別の自動給付エンジンやオンライン照会APIを前提にしない。

- 会計・対象ペット・今回選んだ契約・対象明細と、資格／残限度の確認結果・今回の保険負担額を対応させてsnapshotへ保持する。具体field名・物理型・UI・保存schemaは後続実装詳細で、確認日時や証拠添付等の新たな必須入力を定義しない。既存の計算根拠・変更前後・理由・actorの履歴方針を利用する。
- 未確認／不明・窓口利用可否・補償対象区分を別概念として表現する。値の欠損や通信失敗を有効へ初期化せず、窓口利用不可だけで補償対象外へ変更しない。外部確認結果の変更を過去会計へ遡及適用しない。
- 会計内のペットごとに今回利用する契約を一つに対応させる。複数ペットを一つの契約に潰さず、同一診療費の複数契約併用／給付調整を設計に追加しない。ペットが保有する契約全体の件数制限は導入しない。
- 保険負担額を診療費・税・飼主負担と区別して表示／保存し、飼主の実入金・残額を既存の会計契約へ接続する。記録した負担額を保険会社の実入金・債権回収済みへ変換しない。会社／商品別の税基準・上限・免責・端数は将来自動化する場合に公式条件を確認する別スコープであり、現行の共通率計算を公式検証済みと扱わない。
- ローカル会計の訂正・取消・返金は既存の会計契約に従って処理し、保険会社側への作用とは分離する。保険会社側の請求・給付・残限度・入金を更新するrequestを発生させず、外部変更済みと表示しない。請求送信・入金照合・査定差額管理の新owner／API／DBを追加しない。
- 現行の割合選択と控除計算だけで、今回の確認結果・負担額記録・契約分離が実装済みとは判定しない。具体的な認可mapping、表示と保存方式、競合／再送の実装適合は残件。自動給付算式の未整備を、このスタッフ記録運用の業務未決へ戻さない。

<a id="accounting-eighth-design"></a>
## 第八群の減額・再請求・同時回収・共通料金（提案設計）

[第八群の業務正本](../../../spec/domains/accounting/lifecycle.md#accounting-agreement-eighth)に従う実現案。既存9FRを補完して17ACを追加する。画面の確認だけを整合保証にせず、API／DB・runtimeは今回変更しない。

- 減額intentは元請求／元預り金充当と医院・元飼主、変更後料金、スタッフが選択確認した解除額／現金・カード等側の返金必要額、現在版・理由を対応させる案。最新の請求充足・未払・既実返金・既解除を照合し、解除可能額以内か、復元と実返金で同じ額を二重に戻さないかを検証する。元資金への既返金の対応が不明なら推測で復元せず、第九群の記録確認・内訳／理由登録と照合へ接続する。固定優先順や比例割当は確定しない。
- 料金訂正版・充当解除・元医院／元飼主残高への復元・履歴／監査・結果receiptを既存gateと整列lock下で原子的に保存する案。有効充当額から取消と減額の解除をともに除き、元受領日・受領額を変更しない。実返金は別intentで、未実行の返金を完了にしない。同時返金／別請求充当・再減額・再送・後日の取消でも既実返金／復元済みを再使用しない。
- 預り金に時間経過による自動失効処理を設けない。元clinic／ownerを残高の境界とし、ペット・予約の飼主変更や飼主統合から残高移管を暗黙に起動しない。異医院・別飼主への移行／譲渡は今回設計に追加しない。履歴保持・人物管理内部は既存責務へ従う。
- 再請求intentは取消請求IDと選択元明細ID／版、選択確認内容を受け、新請求ID／番号と元取消請求への参照を返す案。元の取消・確定歴と原明細リンクを保持し、請求の再発行とは別扱いにする。有効請求への元明細帰属を保存時に検査し、選択外の自動追加、同一元明細の重複、並行要求／再送による二重請求を防ぐ。取消履歴の存在自体は再請求を拒否する理由にせず、他の有効請求との重複を検査する。
- 元請求の入金・実返金・充当履歴をコピーして新しい実受領を生成しない。復元残の再充当は既決の確認・残高検証へ接続する。第九群で元会計価格／税／割引の継承とスタッフ修正、旧入金の確認充当、保険結果の再確認を合意。新入金の複製や無確認の資金移動を設計しない。
- 同時回収は新請求候補と選択済み過去請求群、今回の実受領、請求ID別の配分を分離したpreview／結果へ対応させる案。古い順／同日作成順とスタッフ変更・任意額を維持し、選択外への無断配分をしない。新請求に旧未納額を明細として追加せず、原請求のID／日付／残額を保存する。保存直前に所属・認可・取消状態・最新残額・配分合計・版を照合する。
- 同じ操作の新請求・配分履歴・関連残額・監査／receiptを一体で成立させる設計とし、途中失敗や結果不明を全件成功／未納0へ変換しない。外部の実受領が既にあった場合も事実を消したり再受領したりせず、同一操作receiptと実受領事実を照合して再開する。複数writerの参加・transaction範囲と復旧UIは未実装／設計残件。督促・貸倒の新処理を追加しない。
- 共通料金はownerに属しpet参照を持たない元明細として区別し、飼主宛て帳票の別枠へ出力する案。共通料金を除去したり任意のペット・カルテ・保険契約に付け替えたりせず、既決の税区分・全体割引除外・計算根拠・元取引と今回変更額の追跡を保持する。具体schema・画面配置・改ページは後続実装詳細で、全ペット行のnull許可へ無制限に拡張しない。

<a id="accounting-ninth-design"></a>
## 第九群の再請求snapshot・残入金・返金元内訳（提案設計）

[第九群の正本](../../../spec/domains/accounting/lifecycle.md#accounting-rebill-snapshot-payment-source)を実現する案。既存9FR／116ACを保持して12AC追加。具体field／API／DDLや現grantをこの文書だけで追加・変更しない。

- 元取消請求の保存済みsnapshotと選択元明細を対応させ、元価格・税・割引を再請求の初期根拠として保持する。必要なスタッフ修正は元値／新値・理由・actor・時刻と新snapshotへ保存する案。現在マスタへの暗黙refreshや元snapshotの上書きをしない。一部明細の再請求では元の割引scope・配賦根拠を参照できるようにし、新請求の選択行・確認条件から既決の算定を行う。元請求の集約済み税／割引総額をそのまま部分集合へコピーせず、新しい配賦方式も発明しない。
- 残入金の充当intentは元実入金ID・旧配分参照・取消元／新請求・確認額・期待版・理由を結ぶ案。現在の医院／飼主・操作認可・既返金・他配分・新請求未払額を再検証し、元の配分調整と新配分・残額・履歴／監査／receiptを原子的に保存する。確認だけで受領イベント・返金イベントを作らず、元実受領額／方法／日付を保持する。未移動残は追跡を保ち、消去／自動返金／預り金化しない。
- 取消で既に復元した預り金は残入金再配分の対象資金に重ねて含めず、元飼主残高からの既決の確認充当へ分ける。実受領・現在配分・既返金・解除／復元の出所を追跡し、同一資金の二重利用を現在版／整列lockで防ぐ。結果不明や再送で別IDの新入金を作らず、同一operation receiptへ照会する。保存schema・transaction参加・復旧UIの実装適合は未検証。
- 再請求の保険利用では対象ペット／契約／明細に対応する今回の資格・残限度／保険負担額をスタッフが再確認し記録する。元確認結果と今回結果を区別して保持し、前回validを今回の有効値として自動コピーしない。第七群の窓口限定・未確認／不明と有効の分離・ペットごと1利用契約・外部自動変更なしを維持する。
- 返金元内訳不明は、対象資金に依存する返金確定のprecondition不成立として扱う案。スタッフが元記録を確認し、現金・カード・預り金の内訳と理由を登録できるようにする。内訳合計と元取引・現在の返金可能額／有効充当・既返金／既復元を照合し、不明値を0や適当な方法へ初期化しない。登録だけでは実返金履歴を生成しない。
- 元資金の区分と実返金方法は別に保存し、カード原資を現金で返す既決を拒否しない。過去の元資金対応の補完・訂正は元額／元日付・既実返金の事実を保持した理由付き履歴とし、監査／確認版へ結ぶ。確認後に別操作が成立したら古い内訳・残額で確定しない。無関係な請求／回収を全体停止させず、金額別の追加承認や証拠添付の必須欄を新設しない。特定できない間は対象操作を保留したままにする。

<a id="accounting-tenth-design"></a>
## 第十群の空明細・行額丸め・明示的再適用

[第十群正本](../../../spec/domains/accounting/lifecycle.md#accounting-empty-rounding-master)に従う。請求確定intentは最新選択明細数を検査し、0件なら確定イベントを作らない。明細あり0円は行と根拠を保持。過去回収／預り受領を空の新請求へ変換しない。

元単価・数量は十進固定精度／有理数で扱い、積を行ごとに1円四捨五入して値引き前額へ。非負額の0.5境界は整数商・剰余等で比較し、binary float誤差を使わない。元値・積・丸め規則・行額をsnapshotに保持し、割合値引き・配賦・税の段階別結果へ結ぶ。実型／桁数・overflowは後続検証し、表現不能を誤差で受理しない。

除外設定欠落はfalse相当のadapterとし、欠落の来歴と明示true／falseを区別。旧請求金額を現在masterで再計算しない。追加済み行snapshotとmaster参照／版を分離し、更新通知だけでは行を変更しない。明示再適用は対象行・期待版・適用master版・前後値／計算結果を持つ会計操作。未選択行・手修正を暗黙に巻き込まず、確定後は訂正経路へ。競合は停止して現在状態を再取得し、結果不明は同一operation照会へ。

<a id="accounting-reconciled-contract"></a>
## D10全体の状態・保存・互換・受け手契約

合意済み業務を実現する技術設計として本節を選定する。意味型と境界の契約であり、新wire field／ENUM・DDL・権限を実装／付与した宣言ではない。

### 状態と金銭の別軸

| 軸 | 対象ID・意味 | 保護する区別 |
|:---|:---|:---|
| 請求 | 下書き／確定／取消、請求ID／番号・現在版、元明細・確定イベント | 全額入金と別。取消後も確定歴・元予約／記録リンクを保持 |
| 請求額 | 元額・訂正後額・差分、snapshot／規則版、税率別税・保険確認額 | 料金訂正は実入出金でなく、保険負担額は保険会社実入金でない |
| 実受領 | 受領ID・元額／方法／実日付・clinic／owner、請求別配分 | 確定・再配分・預り充当を再受領にしない |
| 支払充足・未払 | 現在の請求別配分・有効預り充当と対応返金／解除 | 未充当残・未確認保険・返金予定を支払済みに混ぜない |
| 返金必要・実返金 | 必要額と、元資金別の実返金ID／方法／額／実日付／理由 | 必要額表示・原資登録・取消・復元だけで実返金としない |
| 預り残高 | 元受領ID・元医院／飼主・任意予約、充当／解除／未充当分実返金・残高版 | 予約ごとの残高に固定せず、復元・実返金・新受領を分離 |
| 帳票 | 対象取引ID／番号・snapshot版・発行／再発行履歴 | operation receiptは飼主用領収書と別 |

保険後の飼主請求額、請求帰属の資金と対応返金を同じ対象版で照合する。確認保険額は対象ペット／契約／明細へ対応させ、不明を0やvalidへ推測しない。取消元額は履歴表示し有効未納として再回収しない。旧残入金は確認充当または実返金へ分ける。元資金／現在残を再構成できない対象操作は保留し、表示上の負数を返金実績へ変換しない。

### 操作別保存単位

全操作はclinic／owner・対象ID／期待版・現在認可・operation ID＋内容digestを検査。必須履歴／監査・結果receiptは業務変更と同時commit。同key同内容は保存済み結果、内容違いは拒否。理由／確認は正本の要求どおりとし金額別追加承認を設けない。

| intent | 対象と追加前提 | 同時保存・返却 |
|:---|:---|:---|
| 確定／同時回収 | 1件以上の選択元明細・snapshot、任意の過去請求群／配分、依存版・未請求性 | 新請求／行・受領事実があれば受領・選択配分・確定イベント／正確な予約リンク・各残額／receipt。旧未納を新明細へ積まない |
| 追加入金／再配分 | 有効元請求群、実受領または既存受領ID・現在残、変更時理由 | 受領は事実がある時だけ。元／新配分・残額・履歴。再配分で実額／日付不変 |
| 未確定の通常単価/割引変更／明示master再適用 | 行・根拠版、前後値、適用master版・必要な保険再確認。通常単価/割引変更の理由は任意。その他行追加など操作別の既存理由必須は維持 | 新snapshot・差分・変更者/日時/前後値・必要履歴／receipt。手修正値を無断破棄せず、医師確認後は金額のみの警告了承と臨床内容変更の再確認を区別。金銭移動なし |
| 確定済み会計の金額訂正 | 元請求・行・根拠版、前後値・必須訂正理由。master再適用で原請求を直接上書きしない | 既存の専用訂正・必要なら取消/再請求へ。旧snapshot/帳票/確定歴を保持し、実受領/返金は別 |
| 取消 | 元請求・理由、最新配分・有効預り充当・既返金／復元 | 取消歴・必要な解除／元飼主残高復元・確定歴保持。実返金／在庫返却なし |
| 実返金 | 元受領・確認原資内訳、既返金／現在上限、理由・実方法／額／日付 | 実返金と元資金対応・残額／履歴。未入金／不明原資／二重返金拒否 |
| 預り受領／未充当返金 | 元飼主・任意予約、実額／方法／日付、返金理由／原資・残 | 受領／実返金履歴・残高版。請求／予約確定なし、使用済み額の返金なし |
| 預り充当／減額解除 | 元資金・対象請求／ペット・任意予約、額・確認・併用内訳／現在残 | 充当／解除・復元、支払充足／残高／履歴。新受領／実返金なし |
| 再請求／旧残入金充当 | 取消元・選択行・元snapshot／修正履歴、保険再確認、充当は確認元資金 | 新ID／番号・元取消参照・重複防止。旧入金は元受領保持で配分し復元残と二重使用なし |
| 支払方法誤記訂正 | 元受領・正しい方法・理由・現在版 | 金額／実日付不変の履歴。ADR003の医院／安定識別対応検査 |
| 帳票発行／再発行 | 閲覧／出力認可、対象取引・snapshot版・元番号 | 発行履歴と出力対応。再発行の元番号維持、新請求／受領なし |
| master変更 | 対象master・操作別許可・clinic／版・前後値 | master・監査／履歴。過去会計／帳票・追加済み行へ暗黙反映なし |

新受領があるintentだけに[停止済み方法の新受領検査](#accounting-inactive-payment-design)を加える。追加入金と再配分、預り受領と未充当返金を同じ行に記載していても、既存資金の操作・実返金へ新受領のgateを一律適用しない。

同一操作で整合する請求・受領・配分・預り残はbilling owner transactionへ一体参加。外部で授受が先行していた場合も失敗表示から再受領／再返金を指示せず、実事実と同一operation結果を照合。照会不能中は対象確定を保留し入力・元資金参照を保持する。

### writer・認可・gate／adapter

[write-owner](../../../architecture/model-write-owner-catalog.md)・[orchestration](../../../architecture/cross-domain-orchestration-catalog.md)を維持。他ownerはbilling表へ直接書込まずtyped intentを使う。[現行billing routes](../../../../backend/internal/billing/routes.go)の権限は新業務を自動許可する根拠でない。

| 群 | 対応・未適合時停止 |
|:---|:---|
| 確定／作成更新 | completeのaccounting:createとcreate／editを区別。旧PATCHで確定を迂回しない。複数元参照／部分・0入金／配分に未対応なら該当新業務を開始しない |
| 明細・確認 | billing-itemsのcreate／edit／deleteとbilling-confirmation editを分離。会計権限を医師確認に流用せずD03へ。GET初期化もwriterならgate参加、印刷readから書込なし |
| 訂正・取消・返金 | 現行credit-correctionのaccounting-post-close-edit:edit、cancelのaccounting-cancel:edit、refundのaccounting:createを現行対応として記録。新業務の返金／訂正許可はC04と照合し不足なら停止 |
| 預り・後日回収・再配分・再請求 | owner intent・認可・原子的履歴／receiptが整うまで新業務停止。既存PATCH等に似たpayloadで代用しない |
| 専用記録／見積起点 | 元ID群・版・理由／確認を受けD10再検査。退院の取消会計削除置換は目標再請求と別の現行差で、元取消ID／番号／履歴を失う経路へ流用しない |
| master・移行修復・締めからのwrite | 画面以外もowner検査へ。import／自動生成／repairも同じgateへ。移行実行・外部送信を本書で許可しない |
| 帳票／集計read | clinic／owner・閲覧／出力認可・snapshot版。別軸金額を旧completed／単一methodへ潰さずconsumer adapterへ |

[全writer gate／取得順](../reservations/technical-spec.md#reservation-global-lock-contract)へ従い、予約依存は下位lock前に参加。複数予約・資金・請求は共通順と整合する安定順で取得し、元clinic／owner／資金で予約なし預りも保護。対象集合が変われば下位lock後に上位lockを足さずtransactionを戻して再検査。参加漏れ・二重writer・監査／receiptの別commitを利用開始前に検証する。既存separate-txを一律same-tx化しない。

### 旧値互換と不明時停止

| 対象 | 読取／変更の契約 |
|:---|:---|
| 単一患者・completed | 旧ID／状態を履歴表示。架空単一患者へ潰さず支払状態をlabelから推定しない |
| method／master ID欠落 | ADR003のENUM／system_key・clinicを保持。NULLを現金／他医院へ推測せず、依存変更時に不明なら保留 |
| 返金原資／預り解除 | 第九群の記録確認・内訳／理由まで依存返金／復元を保留。実額／日付・返金事実を上書きしない |
| 全体割引除外欠落 | 第十群false相当。来歴と明示値を分け、過去snapshot再計算なし |
| 計算／帳票snapshot不足 | 現masterを過去正本にせず既存事実を表示。訂正に必要な根拠不足は確認へ戻し、架空履歴で移行完了にしない |
| 旧保険結果 | 不明をvalidとせず今回確認を記録。再請求再確認、外部残限度を自動変更しない |

意味型・対象ID／版・履歴と停止条件は本書で固定し、物理型／schema・wire名・migration SQL・採番書式は実装で具体化。移行は旧値分類・前後件数／金額／元参照と戻し方を検証し、不明資金を0へ初期化しない。

<a id="accounting-screen-contract"></a>
### 3画面・帳票の取得／保存契約

| 入口 | 目標契約 |
|:---|:---|
| 会計一覧／対象選択 | clinic／owner・検索条件に対応する未請求候補／確定請求／過去未納を別集合で表示。複数pet元IDを保持し未選択行を確定しない |
| 精算詳細／新規 | 行額→明細割引→全体／配賦→税→確認保険の根拠と、請求／実受領／預り充当／未払／返金必要を別表示。空明細不可・0円理由・明示再適用・確認配分をownerへ |
| 未納／後日回収 | 今月売上なしの過去未納も表示。元請求日／ID・現在残／今回配分、filter／基準を明示。選択外を消さない |
| 帳票 | 飼主宛名・pet別明細／共通料金、元取引／今回差分、請求／実受領／充当／実返金を分離。出力失敗で再受領しない |

取得前／中・成功0件／あり・取得失敗・認可拒否を区別。集計は条件／基準日が同じ全対象としページ内合計を全残高にしない。過去時点表示と現在回収可能額を分け、確定には現在残を使う。遅延応答はclinic／owner・請求・検索条件／request世代が一致時だけ採用。

保存中の重複を抑止しreceipt／現在版で成功を確認。通信断は未納0／失敗確定へ変換せず同一operation照会へ。競合／認可失効は停止して未保存編集・元対象を残し再読込。閉じる／戻る／患者切替時の未保存保護、編集を上書きしない再取得を維持。欄と理由を可視化し色だけで伝えずキーボード復帰を可能にする。URL／ボタン名／装飾・改ページは実装詳細で情報欠落を許可しない。

<a id="accounting-master-coverage"></a>
### 4マスタと関連品目の文書範囲

D10は物販・保険・支払方法・キャンペーンを持ち、診療／薬剤は価格／税・割引除外・品目別保険設定の会計接続のみ。臨床設定D03、入口D17の権限を編集許可へ転用しない。

| master | 正本と受渡し | 保護 |
|:---|:---|:---|
| 物販 | [商品](../../../spec/screens/settings/master-merchandise.md)の品名・分類・価格／税・有効状態・品目ID／除外 | 税務分類は別確認。現行にない在庫紐付け／減算はD13へ |
| 保険 | [保険](../../../spec/screens/settings/master-insurance.md)の名称・補償率／連絡先・参照ID。会計の資格／負担額は別 | 補償率と飼主負担率を反転せず、masterだけで有効判定しない |
| 支払 | [支払方法](../../../spec/screens/settings/payment-methods.md)の名称／有効状態・安定識別を受領／返金・集計へ | ADR003保持。固定4種拡張／ENUM廃止を発明しない。承認済みの停止方法の新受領利用不可と過去記録／集計保持は[追加設計](#accounting-inactive-payment-design)へ |
| 割引 | [campaign](../../../spec/screens/settings/master-campaigns.md)の期間・対象・割合／額・有効状態・元ID | 最大割引の現行経路を併用適合済みにせずscope／除外／snapshot対応へ |
| 診療／薬剤 | 品目ID・価格／税／全体割引除外・品目別保険対象設定と追加時snapshot | 投薬量・臨床パラメータを会計丸め／再適用で変更しない |

既存CRUD／reorder・入力表は現行参照として保持。目標create／editは対象変更許可・clinic／版、delete等は別操作許可を検査。名称変更／無効化／削除でも過去取引／帳票参照を保持し、保護できない削除は拒否、別masterへ勝手に付替えない。支払方法の有効設定と選択肢の効果は、現行の固定4種／状態非連動と[追加合意の目標](#accounting-inactive-payment-design)を区別する。ADR003の安定識別・医院分離は保持し、停止済み方法を新受領で許可する根拠にしない。取得／失敗／競合／未保存保護は画面共通契約へ。現grant不足は[権限mapping](#accounting-master-permission-map)へ。

<a id="accounting-item-insurance-design"></a>
## 品目別保険設定の保存・読取・会計引継ぎ（追加設計）

[追加FRとAC](../../../spec/domains/accounting/lifecycle.md#requirement-fr-accounting-item-insurance-setting)の実現案。2026-10-15、PR #670の固定head `908b1def96c91885c01b038bd71af44f500207bf` で下記3画面仕様とD10文書を再取得した。製品コード・DBや実grantの再監査ではない。診療共通パネルの継承説明では保険対象外の保存対応は検査のみで、診察・処置・予防・健診には保存されない。これは目標とのgapであり、表示だけで保存対応済みとしない。

| 対象／既存入口 | 設定writer・読取 | 保存・会計consumerへの契約 |
|:---|:---|:---|
| 診察／処置：consultations・procedures | [診療項目](../../../spec/screens/settings/master-treatment.md)の既存作成／更新、master-medicalの操作grant | 品目別対象設定の保存と一覧／詳細・再編集読取を接続する目標。現行保存対応なしの説明を将来の非対応方針にしない |
| 予防接種：vaccines | 同じ診療項目画面、master-medicalの操作grant | 同上。品目名から補償可否を推定せず、次回日／接種済み記録を変更しない |
| 定期健診：checkup-types | 同じ診療項目画面、checkupsの操作grant | 同上。パッケージ設定の保険区分を扱い、結果判定・次回日へ連動しない |
| 薬剤：medicines | [薬剤マスタ](../../../spec/screens/settings/master-medicine.md)、master-medicalの操作grant | 既存の対象／対象外と明示値を保持し同じ意味の会計引継ぎへ。投薬量・dose gateは変更しない |
| 検査：examination-types／exam_types | [検査マスタ](../../../spec/screens/settings/master-examinations.md)、master-medicalの操作grant | 既存 `exam_types.is_non_insurance` の明示値を保持。測定field／結果／基準値の判定と混同せず検査プラン品目の設定を受け渡す |

意味型は「品目側の明示された保険対象／対象外」と「未取得／取得失敗／未保存・旧欠落」を区別する。前者はboolean等の物理形式へ写像できるが、既存の `is_non_insurance` の反転意味を誤らず、欠落を一律false／trueへ補完しない。新規4種別の新規登録では保険対象外を初期表示・保存値とし、権限のあるスタッフによる対象への明示変更を保持する。再編集時や既存データの読取に新規初期値を再適用しない。旧欠落値の解釈・変換は正本の後続移行課題として扱い、未保存・欠落・取得失敗を対象外へ変換しない。既存薬剤／検査の明示値を変更せず、物理field名・nullable形式・wire／migration SQLは本設計だけから確定しない。全体割引除外のfalse既定は別fieldの別契約である。

- 保存commandは品目種別・ID／医院・期待版・指定値を識別し、既存の操作権限と医院scopeをサーバーで再検査する。対象設定の変更と変更前後・actor／時刻の必要履歴を一体で扱い、通常会計grantをmaster編集権限へ昇格させない。create／edit／delete等の既存操作区分を維持する。
- 一覧／詳細・編集初期表示は実際に保存された同じ品目設定を読める契約とする。取得中／失敗・成功0件・旧欠落を区別し、clinic／種別／ID・request世代の違う遅延応答は捨てる。失敗で入力を空／対象へ戻さず未保存編集を保持し、未対応APIのsilent dropを成功扱いしない。
- 会計候補取得・行追加は元品目種別／ID・医院・参照版と設定を受け渡し、会計snapshotでは品目由来値とスタッフの確認修正値を区別する。元診療／専用記録から既に保存済みの根拠を受けた場合は来歴を維持し、最新masterによる暗黙置換をしない。手入力や元masterのない行へ架空の品目設定を作らない。物販等への新しい設定拡張を本承認から追加しない。
- 編集中行の変更は第十群の明示再適用へ接続し、対象行・前後値・依存版を照合する。確定済み／再請求元は旧snapshotを保持して既決訂正経路へ。帳票・再表示も保存根拠を読み、現在masterで保険額を再算定しない。master側「対象」を資格／残限度／実適用や補償保証に変換せず、会計の既存スタッフ確認と負担額記録をconsumer側で維持する。
- 競合・失効は保存不成立として未保存入力を保護し再取得／再確認へ。結果不明は同じ操作のreceipt／保存版へ照合し、成立済みの同一再送を新しい変更へ複製しない。operation識別・永続receiptと原子保存の実現は既存[D10契約](#accounting-operation-contract)と対象master writerへ適合確認が必要で、現行CRUDが対応済みと断言しない。
- 新規4種別の設定が全writer／read API／会計adapterで往復保持でき、認可・競合・履歴・旧値保護を検証するまでは、その新経路を対応済みとして利用開始しない。独立した既存薬剤／検査や保険を使わない会計まで一律停止する条件を新設しない。新規登録の対象外初期値を旧欠落値のmigration規則へ流用しない。

<a id="accounting-inactive-payment-design"></a>
### 支払方法停止：新受領と過去参照を分離する設計

[追加合意の正本](../../../spec/domains/accounting/lifecycle.md#accounting-inactive-payment-method)を既存9FRの設計へ接続する。承認は新しい入金の選択不可・過去支払記録／集計保持であり、下記はその実現に必要な検査境界。API／DBやruntimeは今回変更しない。

- 新規会計の実受領・後日回収・預り金新規受領では、医院内の安定識別に対応する方法の現在有効状態を候補と保存時に検査する。複合決済の新受領分も各方法を対象とし、表示後停止・古い候補・直接送信で回避させない。停止済み方法を指定した新受領は保存しない。原子操作の一部だけを保存したり、現金等へ黙って差し替えたりしない。
- 方法一覧の取得中／失敗／真の候補0件を分け、有効状態を確認できないまま新受領を確定しない。未保存の額・対象請求・入力を保護し、現在候補の再取得・確認へ戻す。候補0件でも過去履歴を消さず、入金なしの未納／明細あり0円確定をこのgateだけで拒否しない。具体的な文言・表示方法・wire／エラーコード・状態欠落値の扱いや移行手順は本合意から新設しない。
- 保存前の検査と方法停止の競合で、停止済み方法による新受領が成立しない整合境界を設ける。masterの現在版／状態と受領の成立を既存の原子保存・現在認可・医院検査へ接続する。停止前に成立したoperationの同key同内容再送／結果照会は保存済みreceiptを返す。未成立の新保存は現在状態を検査し、結果不明は同一operationの照会を続け、停止を口実に別operationで重複受領しない。物理lockやschemaは後続実装で具体化する。
- 過去参照は新規選択候補と分けて取得する。停止済みの方法でも既存入金・帳票・方法別集計が参照するID／system_keyと医院の対応を保持し、active候補だけのjoin／filterで履歴や集計額を欠落させない。表示名から方法を再推定したり、欠落を現金等に補完したりしない。集計の既存列保持は[画面29](../../../spec/screens/29-closing-aggregation.md#closing-allocation-contract)へ従い、D12／D16の期間・配賦契約を変更しない。
- 保存済み受領の再配分・取消元残入金充当、預り金充当・解除／復元は新受領ではなく、元方法の停止だけで拒否しない。返金と金額不変の方法誤記訂正にも新受領の状態検査を一律流用しない。各操作の理由／確認・現在残／上限・認可・履歴・二重使用防止、安定識別整合を維持する。実返金方法と元資金方法を分け、既決の別方法返金へ追加制限を導かない。

ADR003のENUMとsystem_key／医院master IDの整合は維持する。固定4種や代表方法の現行実装は履歴として扱い、本合意で方法追加・ENUM廃止・代表方法変更・初期値変更を承認済みにしない。全writer／consumerの適合、旧値の実態、移行、実機検証は未実施。

<a id="accounting-handoff-contract"></a>
### 他ownerとの引継ぎ

| 相手 | D10契約 | 相手責務／差分 |
|:---|:---|:---|
| D01／02／03 | 元記録／行／予約ID群・版／確認／理由を受け、現在検査・確定歴・receiptを返す | 通常Undo停止・カルテ保存§2保持。臨床成功で金銭成功にしない |
| D05／06 | 専用記録／元項目・価格条件／版から候補を受け、同時退院等は既存owner transactionへ | 料金・必要前金・取消料はスタッフ個別確認入力、日数自動課金／取消時自動請求返金なし。D05／06完了後は原記録保持・理由付き訂正とし、自動返金／在庫戻しなし。取消会計削除置換と新番号再請求のadapter受入が必要 |
| D11 | 元見積／版・選択行／確認条件を受け、新請求と元参照を返す | 見積状態／期限／後継版は相手。見積作成を請求確定・受領にしない |
| D12／16 | 請求差分、実受領／返金のID・元資金／請求・実日付／方法、充当／解除／再配分の非金銭eventを分離 | completed_at／completed請求額を実入出金へ流用不可。売上と現金受領を分け、締め／期間／集計adapterは相手受入 |
| D13 | 元商品／明細・訂正参照と事実 | 取消だけで在庫返却しない。移動は在庫owner、現行にない連動は推定しない |
| D04／19 | 元clinic／owner／petのID・履歴保持 | 名前一致／統合／所有変更だけで債権・預り金移管なし |
| C04／05 | 論理操作／scope・前後値／理由・actor／時刻・操作ID／結果 | 実grant／付与失効・保管は相手。旧route許可で新業務開放なし |
| D15 | 既存commit結果／元eventをconsumerへ | PATH-BILL-LSTEP-CPM-SYNCのseparate-tx／best-effort保持。失敗で会計を戻さず同一確定再送で副作用再実行なし。外部送信再開は別承認 |

新cross-domain writer実装時はcatalogへowner・transaction・失敗復帰／監査を登録。本設計で全既存経路をsame-txへ読み替えない。


<a id="accounting-posthumous-design"></a>
## 死亡状態を保持した生前分精算の接続

[死亡後の精算](../../../spec/domains/accounting/lifecycle.md#accounting-posthumous-settlement)は既存の会計操作ID・期待版・現在認可・金銭projection・原子receiptの契約を使い、死亡状態を生存へ変更するwriterを呼ばない。元患者・診療／滞在の実施事実と元明細・請求・飼主を照合し、理由と必要履歴を当該成功保存に含める。生前実施を識別できない場合は入力を保持して元記録の確認へ戻し、会計日や現在statusだけで診療事実を生成しない。死亡・主飼主変更・他編集で確認版が変わった場合は最新対象を再取得・確認し、無警告再送や別操作IDでの二重精算をしない。実入出金・訂正・残高復元の分離、元飼主の請求／預り金、実返金別操作、通常Undo停止は保持する。画面の死亡新規gateと、この専用の後日精算・訂正を区別する。既存実装が全入口で適合済みという意味ではない。

2026-10-22選定の[限定認可・参加・互換境界§4.1](../../../work/emr/emr-540-308-462-506-selected-handoff-20261022.md#posthumous-selected-participation)を併読する。lookupは認証医院・同一actor・保存済み元対象anchorから解決した対象への現在`accounting:view`を検査し、request/URLへ元対象をfallbackせず、失権/解決失敗時は結果非開示。mutationの既存write grantを維持し、勤務consumerの元write grant要件を金融lookupへ自動移植しない。462 completeの初回201/再送200・Billing wireは維持。一般506 route/DTO/HTTP等は同追補§4.2の未選定残であり、共通化だけで既存wireを変更したり実装/受入完了としない。

一般506の受領/複数請求配賦・方法/日時誤記訂正については[設計候補・未採択の整理](../../../work/emr/emr-506-fund-wire-design-candidates-20261022.md)を参照。候補route/新wire/CAS tokenを実在APIやclosed contractとして扱わず、既決のlookup認可・元owner・原子receiptと462 Completeの整数wire/201・200は維持する。

<a id="accounting-document-completion"></a>
## D10文書完成条件と判定外

D03と同じ限定文書基準を適用。ここまでの基準は13FR／174AC（前段階10FR／167AC＝基準9FR／155AC＋品目別保険設定1FR／12AC、2026-10-17の医師確認・プラン同期・請求対象3FR／7ACを追加）。なお現在は13FR／184AC（EMR-367追補3ACを含む）。No.8の個別割引権限1AC・No.30の釣銭整合3AC・過受領処理3ACを含む現在のライフサイクル本文は13FR／184AC。対象は3画面＋選択／新規入口、4masterと診療／薬剤会計接続、語彙・状態／ID・操作／保存／認可・旧値／停止・writer／consumer／受け手。2026-10-15の独立レビューとリンク／fragment・表／trace照合で重大指摘なしを確認し、基準時点の9FR／142ACを文書整理完了と判定した。追加の死亡後精算2ACも反映・独立文書レビュー済み。支払方法停止の追加7ACも前段階の固定39文書レビュー・限定構造検査を通過した。会議連携4ACと今回の品目別保険設定1FR／12ACは、それぞれ追加差分の文書確認として区別する。今回の新規初期値は対象外と合意済みであり、旧欠落値の解釈・変換という後続移行課題を既存完成判定で完了扱いしない。既存144ACの判定とは別に追加差分の確認根拠を保持する。既存AC／アンカー・144AC本文を保持し、新旧全ACを一意に追跡する。

文書で選定した意味型・原子保存／receipt・現在認可／不明時停止・画面状態／master参照保護・引継ぎを後決めへ戻さない。物理schema／wire名・採番書式／lock実装・実grant／実旧値・移行・全writer／consumer実装適合・実機／性能／復旧試験・税務／決済規約適合は後続実装／利用開始条件。これらと他owner内部・全体横断最終照合は本領域の文書完成判定外。保険会社請求／査定・自動給付、督促／貸倒、別医院／飼主移管・外部機器連携は対象外。

照合する境界は、確定×未入金／部分×空明細／0円、訂正／取消×現金／カード／預り×既返金／復元、再請求×旧入金／預り残、行額丸め×割引／除外／複数税率、再適用×手修正／確定競合、不明旧値×返金、保存×失敗／同一再送／内容違い／遅延成功、画面切替×失敗／古い応答、各master×無認可／失効／参照保護。期待結果は正本ACと本契約、実行受入は後続工程へ。

## 受入条件との追跡と設計残件

| 業務FR | 正本AC範囲 | 技術照合対象 |
|:---|:---|:---|
| FR-accounting-item-insurance-setting（品目設定） | item-insurance-four-masters／item-insurance-existing-med-exam／item-insurance-not-guarantee／item-insurance-independent-discount／item-insurance-authorized-save／item-insurance-snapshot-stable／item-insurance-explicit-reapply／item-insurance-finalized-correction／item-insurance-default-unresolved／item-insurance-legacy-missing／item-insurance-read-failure／item-insurance-atomic-retry | [追加設計](#accounting-item-insurance-design)。6種別の設定往復・実適用分離・独立区分・現在認可／版・snapshot／訂正・新規対象外初期値／旧欠落保護・取得失敗／原子再送 |
| FR-accounting-owner-selected-charges（会議連携） | trimming-zero-line | [施術側調整・記録復元の受渡し](#accounting-trimming-adjustment-adapter)。実在0円行と元版／計算snapshot |
| FR-accounting-calculation-snapshot（会議連携） | trimming-adjustment-once／trimming-source-refresh | [施術側調整・記録復元の受渡し](#accounting-trimming-adjustment-adapter)。根拠付き一度適用・明示再適用と確定保護 |
| FR-accounting-reasoned-financial-correction（会議連携） | record-edit-no-money-undo | [施術側調整・記録復元の受渡し](#accounting-trimming-adjustment-adapter)。臨床本文復元と金銭操作を分離 |
| FR-accounting-owner-selected-charges | owner-multi-pet／selected-unbilled／old-bill-collection／no-invented-allocation | 複数元参照・選択・二重請求防止・元請求帰属 |
| FR-accounting-receivable-partial-zero | unpaid-finalize／partial-finalize／later-payment／zero-no-fake-payment | 請求と入金の別状態・履歴・残額・0円根拠 |
| FR-accounting-reasoned-financial-correction | correction-authorization／correction-history／refund-received-cap／refund-method-required／cancel-no-money-stock | 認可・理由・監査・実入金上限・方法・owner分離 |
| FR-accounting-post-close-separate-movement | post-close-preserve／correction-no-cash-movement | 元締めと差分・実移動の分離 |
| FR-accounting-calculation-snapshot | master-change-stable／manual-unit-price／discounts-combinable | snapshot・修正元値・第五群の割引算定と第七群の保険確認記録 |
| 既決横断境界 | finalization-undo-preserved／clinical-no-auto-finance | 確定歴・正確なリンク・owner receipt |
| FR-accounting-oldest-payment-allocation | oldest-default／allocation-staff-change／collection-change-return | 古い順・変更・合計・元請求・釣銭と前金の区別 |
| FR-accounting-reservation-advance | advance-owner-reservation／advance-unallocated-balance／advance-apply-no-double-receipt／advance-surplus-pending／advance-actual-refund／advance-duplicate-use／advance-context-mismatch | 元予約／飼主・未充当残・二重使用／二重受領防止・実返金 |
| FR-accounting-reasoned-financial-correction（第二群） | paid-reduction-refund-pending／paid-reduction-actual-refund／partial-paid-reduction／payment-method-correction-no-money | 訂正後表示・未実行／実返金・金額不変の誤記訂正 |
| FR-accounting-calculation-snapshot（第二群） | percent-before-fixed／insurance-staff-review | 割合額四捨五入→金額・保険確認修正snapshotと第七群の記録範囲 |
| FR-accounting-document-and-aged-debt | invoice-receipt-separated／document-reissue-history／aged-debt-without-current-sales | 請求／実受領帳票・再発行／旧履歴・過去未納の表示 |
| FR-accounting-post-close-separate-movement（第二群） | original-bill-and-movement-date | 原請求日と実日付・充当時の非受領 |
| FR-accounting-reservation-advance（第三群） | deposit-without-reservation／deposit-other-reservation／deposit-other-pet／deposit-other-owner-denied／deposit-allocation-confirmation／deposit-available-balance | 飼主帰属・予約任意・確認・利用可能額・同時使用防止 |
| FR-accounting-calculation-snapshot（第三群） | percentage-summed／discount-then-insurance／overdiscount-unresolved | 同scope合算→金額→割引後保険、超過警告・対象額上限、第五群の順序／配賦・税と第七群の保険確認記録 |
| FR-accounting-oldest-payment-allocation（第三群） | same-day-created-order／saved-reallocation-reason／reallocation-no-new-money／reallocation-correction-conflict／noncash-excess-refund-pending | 作成順・理由と履歴・原子的再配分・非現金超過 |
| FR-accounting-reasoned-financial-correction（第三群） | refund-different-method／refund-no-amount-extra-approval | 別方法・金額追加承認なし・総額上限／権限 |
| FR-accounting-document-and-aged-debt（第三群） | deposit-receipt／final-document-deposit-current／refund-original-transaction／reissue-original-number | 用途別帳票・充当と実受領・取引番号／再発行履歴 |
| FR-accounting-calculation-snapshot（第四群） | line-and-bill-discount／overdiscount-warning／overdiscount-post-warning-pending | 両粒度利用・明細→全体・超過警告・実適用上限と請求額0円下限 |
| FR-accounting-reservation-advance（第四群） | cancel-restore-deposit／cancel-deposit-no-cash-refund／cancel-deposit-once-atomic | 取消・充当解除・残高復元と別返金・原子的一度処理 |
| FR-accounting-calculation-snapshot（第五群） | line-before-invoice／discount-scope-separated／percent-discount-round-yen／invoice-proportional-remainder／invoice-zero-discount-base／overdiscount-line-no-spill／overdiscount-invoice-effective-cap／tax-invoice-rate-once／tax-rate-separated／tax-exclusive-consumer-inclusive／tax-category-snapshot／calculation-audit-replay | 明細→全体・scope分離・割合額丸め・比例配賦・0円／超過・税率別税・税込表示・計算監査 |
| FR-accounting-calculation-snapshot（第六群） | insurance-multiple-carriers／invoice-exclusion-master／invoice-excluded-base-and-allocation／invoice-all-excluded-zero-base／invoice-exclusion-not-insurance／invoice-exclusion-snapshot-stable | 複数社と第七群の確認記録、除外指定・対象額／配賦・保険と税の分離・snapshot |
| FR-accounting-document-and-aged-debt（第六群） | document-owner-addressee／document-pet-separated／document-correction-original-delta／document-partial-refund-original-delta | 飼主宛名・元ペット区分・元取引と今回差額／実返金・旧履歴 |
| FR-accounting-master-authorized-maintenance（第六群） | master-admin-edit／master-authorized-staff-edit／master-ungranted-staff-denied／master-permission-rechecked | 管理者／指定スタッフ・対象grant／医院・無許可と失効拒否・監査 |
| FR-accounting-calculation-snapshot（第七群） | insurance-staff-result-burden／insurance-unknown-not-valid／insurance-window-not-coverage／insurance-one-contract-per-pet／insurance-window-only-scope／insurance-local-correction-no-external／insurance-manual-no-formula-prerequisite | 外部画面等の確認結果／保険負担額・不明と利用可否の分離・ペットごと1利用契約・窓口限定・外部自動変更なし・自動算定は将来 |
| FR-accounting-reservation-advance（第八群） | deposit-reduction-restore／deposit-mixed-reduction-choice／deposit-reduction-current-once／deposit-no-auto-expiry／deposit-clinic-owner-boundary | [第八群の提案設計](#accounting-eighth-design)と業務正本へ照合。元資金／元明細／元請求・履歴・現在版を保持 |
| FR-accounting-reasoned-financial-correction（第八群） | deposit-reduction-no-money／rebill-selected-new-link／rebill-no-duplicate／rebill-no-money-copy／rebill-history-undo | [第八群の提案設計](#accounting-eighth-design)と業務正本へ照合。元資金／元明細／元請求・履歴・現在版を保持 |
| FR-accounting-oldest-payment-allocation（第八群） | combined-collection-allocation／combined-collection-order-override／combined-collection-current-state | [第八群の提案設計](#accounting-eighth-design)と業務正本へ照合。元資金／元明細／元請求・履歴・現在版を保持 |
| FR-accounting-owner-selected-charges（第八群） | combined-collection-no-rebilling | [第八群の提案設計](#accounting-eighth-design)と業務正本へ照合。元資金／元明細／元請求・履歴・現在版を保持 |
| FR-accounting-document-and-aged-debt（第八群） | collection-dunning-writeoff-out／document-common-charges／common-charge-provenance | [第八群の提案設計](#accounting-eighth-design)と業務正本へ照合。元資金／元明細／元請求・履歴・現在版を保持 |
| FR-accounting-calculation-snapshot（第九群） | rebill-original-price-tax-discount／rebill-staff-adjustment-history／rebill-insurance-rechecked／rebill-insurance-scope-preserved | [第九群の提案設計](#accounting-ninth-design)。元snapshot・旧入金の帰属・保険再確認・返金元内訳と理由・履歴／現在版 |
| FR-accounting-oldest-payment-allocation（第九群） | rebill-old-payment-confirmed／rebill-old-payment-no-receipt／rebill-old-payment-current-cap | [第九群の提案設計](#accounting-ninth-design)。元snapshot・旧入金の帰属・保険再確認・返金元内訳と理由・履歴／現在版 |
| FR-accounting-reservation-advance（第九群） | rebill-restored-deposit-separate | [第九群の提案設計](#accounting-ninth-design)。元snapshot・旧入金の帰属・保険再確認・返金元内訳と理由・履歴／現在版 |
| FR-accounting-reasoned-financial-correction（第九群） | refund-source-unknown-blocked／refund-source-record-before-finalize／refund-source-history-no-double／refund-source-scope-not-method | [第九群の提案設計](#accounting-ninth-design)。元snapshot・旧入金の帰属・保険再確認・返金元内訳と理由・履歴／現在版 |
| FR-accounting-receivable-partial-zero（第十群） | empty-invoice-blocked／zero-invoice-line-reason | [第十群設計](#accounting-tenth-design)と全体契約。空明細・行別丸め・設定欠落／明示値・snapshot／再適用版 |
| FR-accounting-owner-selected-charges（第十群） | collection-no-empty-invoice | [第十群設計](#accounting-tenth-design)と全体契約。空明細・行別丸め・設定欠落／明示値・snapshot／再適用版 |
| FR-accounting-calculation-snapshot（第十群） | pre-discount-round-half／pre-discount-round-per-line／pre-discount-round-before-discount／pre-discount-round-provenance／exclusion-default-included／exclusion-missing-included／exclusion-existing-preserved／master-draft-no-auto-update／master-explicit-reapply／master-reapply-current-check | [第十群設計](#accounting-tenth-design)と全体契約。空明細・行別丸め・設定欠落／明示値・snapshot／再適用版 |
| FR-accounting-reasoned-financial-correction（第十群） | master-reapply-finalized-boundary | [第十群設計](#accounting-tenth-design)と全体契約。空明細・行別丸め・設定欠落／明示値・snapshot／再適用版 |
| FR-accounting-owner-selected-charges（死亡後精算） | posthumous-settlement-history | [専用接続](#accounting-posthumous-design)。死亡状態保持・生前元事実・理由／履歴・原子receipt |
| FR-accounting-reasoned-financial-correction（死亡後精算） | posthumous-owner-no-side-effect | [専用接続](#accounting-posthumous-design)。元飼主資金と既存金銭／在庫／予約境界 |
| FR-accounting-master-authorized-maintenance（支払方法停止） | payment-inactive-new-receipt／payment-inactive-save-recheck／payment-inactive-history-aggregation／payment-inactive-idempotent-result | [追加設計](#accounting-inactive-payment-design)。新受領の候補／保存時状態・過去参照と集計・競合／成立済み再送を分離 |
| FR-accounting-oldest-payment-allocation・FR-accounting-reservation-advance（支払方法停止） | payment-inactive-existing-allocation | [追加設計](#accounting-inactive-payment-design)。既存資金の配分・充当／解除と新受領を分離し既決の保護を維持 |
| FR-accounting-reasoned-financial-correction（支払方法停止） | payment-inactive-refund-correction-boundary | [追加設計](#accounting-inactive-payment-design)。返金・金額不変訂正に新受領のgateを一律適用しない |
| FR-accounting-receivable-partial-zero（支払方法停止） | payment-inactive-no-receipt | [追加設計](#accounting-inactive-payment-design)。未納／0円確定と実受領を分離 |
| FR-accounting-doctor-review | before-doctor-review | [医師確認の版・確定gate](#review-version-sync-proposal)。未確認でも会計準備は可、請求確定／精算は停止 |
| FR-accounting-doctor-review | review-card-axis | [確認状態projection](#review-version-sync-proposal)。診療終了の会計待ちと、受付カードの医師確認表示を別軸で照合 |
| FR-accounting-doctor-review | reviewed-price-only | [金額変更了承](#review-version-sync-proposal)。金額のみは警告了承・変更履歴、医師再確認不要 |
| FR-accounting-doctor-review | reviewed-clinical-change | [臨床世代・再確認gate](#review-version-sync-proposal)。直接変更／同期とも薬剤・処置追加削除／数量変更は再確認まで確定・精算停止、プランへの逆自動書換えなし |
| FR-accounting-plan-sync | plan-sync-three-branches | [同期の参加順と行版](#review-version-sync-proposal)。未確定未手修正／手修正済／確定済の3分岐と変更種別の医師再確認gate |
| FR-accounting-billable-performance | billable-boundary | [実施・準備の元事実](#review-version-sync-proposal)。未実施／予定／中止と請求対象・予定額を分け、個別中止料を保持 |
| FR-accounting-billable-performance | ready-not-stock | [実施・準備と実出庫の分離](#review-version-sync-proposal)。持帰り準備・請求／取消／返金から実引渡し・在庫出庫を自動実行しない |
| FR-accounting-calculation-snapshot（個別割引権限） | individual-discount-permission | [現在認可・保存単位](#accounting-reconciled-contract)と[正本の個別割引権限](../../../spec/domains/accounting/lifecycle.md#利用者権限個別会計の割引入力)。通常会計／マスタ変更権限だけで許可せず、無許可は元値保持 |
| FR-accounting-oldest-payment-allocation（釣銭整合） | manual-change-unreconciled-completion | [受領・配分の保存単位](#accounting-reconciled-contract)と[釣銭の正本](../../../spec/domains/accounting/lifecycle.md#accounting-manual-change-reconciliation-20261017)。実受領・請求充当・残額・過受領の扱いを照合してから完了 |
| FR-accounting-post-close-separate-movement（釣銭整合） | manual-change-no-fabricated-adjustment | [金額と実移動の分離](#accounting-operation-contract)と[釣銭の正本](../../../spec/domains/accounting/lifecycle.md#accounting-manual-change-reconciliation-20261017)。架空の売上・値引き・入出金・自動預り金化を作らず、締めは釣銭後の実受領 |
| FR-accounting-oldest-payment-allocation（釣銭整合） | manual-change-preserves-partial-and-noncash | [部分入金・非現金超過の既存契約](#accounting-second-design)と[釣銭の正本](../../../spec/domains/accounting/lifecycle.md#accounting-manual-change-reconciliation-20261017)。部分入金／残額を維持し、非現金超過は実入金・充当・返金必要／実返金を分離 |
| FR-accounting-post-close-separate-movement（過受領） | excess-input-correction | [訂正と実移動の分離](#accounting-operation-contract)と[過受領の正本](../../../spec/domains/accounting/lifecycle.md#accounting-excess-receipt-resolution-20261017)。実際の預かり・返却額へ訂正し、入力訂正で架空移動を作らない |
| FR-accounting-reasoned-financial-correction（過受領） | excess-actual-refund-once | [返金元・実返金の既存契約](#accounting-third-design)と[過受領の正本](../../../spec/domains/accounting/lifecycle.md#accounting-excess-receipt-resolution-20261017)。実返金は一度だけ記録し、未実行／釣銭修正との二重控除を防ぐ |
| FR-accounting-reservation-advance（過受領） | excess-consented-deposit-single-receipt | [元資金・預り残高の既存契約](#accounting-reconciled-contract)と[過受領の正本](../../../spec/domains/accounting/lifecycle.md#accounting-excess-receipt-resolution-20261017)。飼主了承の明示預りを元実受領へ対応付け、追加入金／後日充当の再受領を作らない |

overdiscount-unresolved／overdiscount-post-warning-pending は参照互換の既存識別子であり、期待結果は合意済みの警告・請求額0円下限へ更新する。

AC短名は[正本の基準174AC（後続補足を含むライフサイクル本文181AC）](../../../spec/domains/accounting/lifecycle.md#accounting-acceptance)のAC-accounting-接頭辞に続く名称。ここに第二の期待結果を作らない。 対応表はライフサイクル本文181ACを各1回収録する。2026-10-19の追記14AC（医師確認等7・個別割引権限1・釣銭整合3・過受領3）は既存ACと既存設計／業務正本の接続補足で、新しいFR/AC・保存方式の採用・実装／実機合格を表さない。subpage固有20ACは別の母集団として下記の補足接続等へ分ける。

実装前に、確定と入金の競合、同時返金による上限超過、同時二重請求、失効認可／異医院参照、通信切断後の遅延commit、複数記録と締め参加、マスタ変更とsnapshotを対象に正常／拒否／再送／結果不明を照合する必要がある。既存S08／S21／S33は現行手順で、現在のライフサイクル本文181ACの実行済み証拠ではない。runtime・API／DB改訂・移行・利用開始・性能は未実施。

業務判断と技術設計／検証・対象外の区別は[残件の分類](../../../spec/domains/accounting/lifecycle.md#accounting-open-decisions)へ。D05／06専用記録、D11見積、D12締め、D13在庫、D16集計、C04権限／C05監査の受け手実装適合・全体横断再照合は別工程。保険会社への請求送信・入金照合・査定差額管理・会社／商品別の自動給付算定、外部支払／印刷機器を追加実装する設計は含めない。



<a id="accounting-trimming-adjustment-adapter"></a>
## 施術側調整・記録復元の受渡し

[業務契約](../../../spec/domains/accounting/lifecycle.md#accounting-trimming-adjustment-handoff)に従い、元施術ID／版とコース／optionの元明細IDを会計行に対応付ける。元価格・料金手修正・D10明細scopeの値引条件・算定結果を受け、同じ元調整が一度だけ計算に含まれることを検査する。別の会計全体値引きと同一扱いせず、元施術小計を新しい架空明細にして再控除しない。既請求帰属・0円理由・現在scope／認可・計算snapshotは会計の既存保存契約へ参加する。

確認後の元施術版・会計版・価格／値引きが変われば差分を再確認し、無警告で再取込／再適用しない。受渡しの操作ID・元参照・内容一致と保存結果を対応付け、同一操作の応答不明は先行結果を照会する。明細の追加と根拠／必要監査・操作結果は会計保存単位で一体にし、二重行・二重値引き・一部根拠欠落を成功としない。

現行未請求SQLのmaster価格>0抽出と、施術側料金／値引きsnapshotがない状態は適合差である。元値・調整理由／版・計算範囲が不明なら現在masterから当時のsnapshotを捏造しない。物理wire／schemaと各writerの適合は後続実装で検証する。D03通常編集／クリア／複製Undoのreceiptを会計訂正・返金・在庫補償のreceiptに流用しない。

## 月次未納帳票の原資料追加接続

[業務正本と追加3AC](../../../spec/domains/accounting/pages/unpaid-list.md#unpaid-monthly-source-adoption)を参照する。請求残高projectionとpayment／deposit application／release／billing correctionは別イベントidentityとして照合し、前月末cutoff・対象月・翌月繰越の同一医院／飼主／請求範囲を揃える。請求表示日へ実入金日を付け替えず、未来の訂正を当時残高へ無根拠に遡及しない。入金合計へ預り金充当／請求減額を混在させて4列を形式的に一致させない。

2026-10-16の承認により、4列を保持し、当月入金を実受領のみ、非入金の残高増減を別掲「調整内訳」とする。元event／請求／資金identity・発生日・種別・残高影響の符号と根拠をread projectionへ渡す。充当／解除・訂正／取消を一度だけ反映し、当月発生額への訂正反映と調整行の二重控除を防ぐ。印刷と画面は同じ対象・cutoff・調整内訳を使い、翌月開始残との連続性も照合する。正本の数値例と3ACが受入基準であり、実装済みの宣言ではない。

未充当受領・非現金超過は実受領総額と請求帰属額の差を区別し、既存資金の再配分を新受領にしない。実返金の詳細表示／当月入金との相殺方針は今回新設しない。限定式の対象外の金銭eventを含む月では既存contractに基づく差額説明が必要で、説明不能を0円へ丸めたり照合成功にしない。既存APIの現在残高だけでは過去cutoffを再現できない場合、snapshot／event projectionの実装差分として残す。保存済み税額・請求額を表示のためにwriteしない。

## 日次集計内訳の設計接続

[追加2ACと正本](../../../spec/screens/29-closing-aggregation.md#closing-source-breakdown)へ接続する。元明細の安定カテゴリと保存済み金額／税／値引きから部門表示を構成し、往診・トレセンを別identityで保持する。現在の名称変更から旧会計分類を書き換えない。支払実額配賦は画面29の既存契約へ委譲し、税表示を別の課税処理として実行しない。既存の確定税を再現できない場合はsource不足として検出する。部門対応・表示端数処理と保存済み内訳の実データ照合が必要で、今回DB／集計関数が対応済みとはしない。

<a id="review-version-sync-proposal"></a>
## 医師確認の版・未確定同期の詳細設計案

[業務正本の確認・同期・請求対象](../../../spec/domains/accounting/lifecycle.md#doctor-review-sync-billing-20261017)と[C06提案状態](../../shared/integration-contracts.md#detail-design-proposals-20261018)へ従う技術提案。会計ページ準備・診療終了・医師確認・請求確定・実入金を別操作として保持する。候補field・状態は新wireの案で、コード／DB／APIに反映していない。

### 確認対象と同期元の識別

| 資源／owner | 保存する版・参照案 | 更新契機と制約 |
|:---|:---|:---|
| 確認対象臨床集合／D03 | record ID＋`clinical_generation`、確認対象薬剤／処置の安定ID・薬剤／処置identity・数量 | 追加／削除／数量変更で世代を増分。単価・割引・表示順だけは増分しない。同じ値へ戻しても途中の対象変更を無かったことにしない |
| 確認receipt／D10 initiator＋D03 participant | 認証確認医師・時刻・医院・record集合と臨床世代、確認操作key、会計側の臨床変更世代 | 既存の医院が認めた医師／確認権限を照合し、client checkや担当医IDから確認者を推定しない。確認履歴を上書きせず新receiptとして保存 |
| 会計明細／D10 | 安定billing行ID、source kind/id、最後に取り込んだsource版、会計行版、項目別manual override／元値（いずれかの手修正があれば行全体を手修正済みとして保護） | 金額の手修正と臨床内容の差異を区別。master最新版と、プランの保存済みsource版も区別する |
| 同期差異／D10 | sourceの現在版、前回適用版、変更field、削除tombstone、解決状態、明示反映操作 | 手修正済みは上書きせず差異を表示。元削除後の手修正行も黙って消さない。明示反映・保持・元記録訂正の選択を元版へ結び付ける |
| 金額変更了承／D10 | 了承したbefore/after計算snapshotのrevision・対象行・提示金額／条件をContent、会計／行の期待版をCASとして分離。client operation_keyはUUID、内部binding IDとは別 | 金額のみは警告了承で可、医師確認receiptを失効させない。新しい理由必須・医師再確認を追加しない |

確認対象の新たな種類を勝手に増やさない。既決は薬剤・処置の追加削除／数量変更であり、検査・物販までの包括拡張は今回採用しない。薬量の安全確認・別目的の確認は既存正本へ。D03の全dependency版を医師確認の世代に使うと金額・表示変更まで再確認となるため分離する。

会計から臨床変更があった場合はD10側の確認対象変更世代を進め、D03との差分を医師へ提示する。確認操作は現在の両側差分と世代を対象にし、医師が必要に応じD03を別intentで修正してから現在内容を確認する。会計writeからプランを自動書換えしない。一つの親recordしかない前提へ潰さず、複数元記録を含む会計は対象ref集合と各世代を保持する。

<a id="price-acknowledgement-handoff-20261020"></a>
### EMR-483の計算revision・receipt接続（未実装の限定依存）

[EMR-483](https://app.plane.so/baritechllc/browse/EMR-483/)の02:32 UTC着手プランを技術mappingとして適用する。`billing_price_acknowledgements.calculation_version`は了承AFTER側`billing_calculation_snapshots.revision`案であり、会計CAS版や計算規則版ではない。before/afterのschema付き値に同医院・会計の固定snapshot ID／revision・行identity・提示額／条件を保持する。after snapshotを同TXで採番する場合、採番後IDは保存結果／履歴へ置き、事前Contentには固定したrevision／内容を使う。after内容やrevisionが変われば再提示・再了承が必要。履歴のbilling_version／line_versionは検査した事前CAS版、成功結果のcommitted_versionsは保存後版へ分ける。除去済みbillings.calculation_versionやcompletion_request_id/hashは復活させない。

認証clinic／actorからscopeを作り、client operation_key UUIDと内部binding IDを区別する。同じ正規化済みimmutable commandからfingerprintとApplyの書込内容を導出し、別request payloadで書き込まない。同内容の古いCASが確定拒否された後は、人の再確認で前提だけ更新し同keyを維持できる。了承版・対象・金額変更は旧結果確定後の別intent/key。成功済みは現在認可を検査して保存結果を返し、unknown／処理中の自動再実行・新key救済はしない。臨床差分は462の確認経路へ戻す。

保存経路には[446／466の限定DB条件](../../../work/db-contracts/unified-db-20261019.md#price-ack-db-handoff-20261020)が残る。request_hashの二重再送正本を共通binding fingerprintへ整理する最小参照／制約と、理由不要のprice-only了承を記録するevent契約は未実装。専用price_acknowledgement intentは第一候補で、採用DDLではない。correct/cancel/refundの理由必須を緩めず、偽finalize・ダミー理由で迂回しない。snapshot・ack・適合event／必要監査・成功receiptは同TX、いずれの失敗も全rollbackとする。483はdomain adapter／writer／reader、489は共通helper／DTO／参加表、446は保存契約、466は限定統合と受入を所有する。

[全件引渡し索引](../../../work/emr/emr-445-plan-handoff-20261020.md)の必要subsetだけを受領し、489全owner Done待ちにはしない。DTO・fixture設計は先行可能だが、上記DDL判断・489の修正版契約・実001/RLS/deferred制約の受入証跡なしに当該保存経路を受入済みとしない。

### 状態・同期・確定の参加順

表示状態は未確認／確認済み有効／臨床差分で要再確認／判定不能のprojection。これは新しい予約statusではない。診療終了で会計待ちへ移し、受付カードの医師確認表示は別軸。確認済み有効は「現時点の対象ref集合・両側臨床世代がreceiptと一致」で判定し、単純boolに固定しない。

| 操作 | 原子的に行うこと／失敗時 |
|:---|:---|
| 準備／未手修正行の同期 | 既存共通gate・receipt・各owner下位順に参加。現在source refと版・請求対象性、D10未確定・行版を検査。手修正のない行だけ対応fieldを適用し、取込版／計算snapshot／必要監査／成功結果を同時保存。sourceを再createせず一意の元参照を維持 |
| 手修正済み行の差異検出 | 未確定なら現在sourceとの差異を保存・表示し、明細を自動上書きしない。同期workerの遅延通知はヒントで、保存／確定時に現在sourceを再照合する |
| 金額のみ変更 | 現在の医師確認があれば、了承したbefore/after計算revision・対象行・金額と、現在会計／行CASを別々に再検証して保存。了承snapshot・履歴・適合event／必要監査・成功receiptを同じ最外TXへ。プランmaster・医師確認臨床世代を変更しない |
| 薬剤／処置追加削除・数量変更 | 元項目ID／版、D03又はD10側の臨床世代、会計行版を更新し要再確認へ。確認歴は残し、確認後の確定・精算gateを停止 |
| 医師確認 | 現認可・医師資格と対象ref集合／両側臨床世代・提示差分を同txで再照合し、新receipt＋確認履歴を保存。確認後に変化すれば無効になるため、チェック押下だけを成功条件にしない |
| 請求確定／精算 | 現在認可・医師確認receiptと現在対象世代・source差異・金額了承の必要性・会計期待版・対象性を共通gate下で照合。請求・確定歴・必要監査・成功receiptと既存予約participantを同時commit。失敗は金銭だけ先に確定しない |
| 確定済み会計の変更 | 同期で書換えず既存理由付き会計訂正へ。返金／取消／再請求でも確定歴と元refを保持し、通常予約Undoへ自動復帰しない |

自動同期と手編集の競合は同じ会計行版をcompare-and-swapする。手編集先行なら同期は差異表示、同期先行なら手編集は現値を再取得し差分確認。同期・確認・確定を両先行順序で検査する。自動同期自体は既決の未手修正分岐なので、日常操作へ全行手動転記を要求しない。

手修正した金額差異の存在だけを新しい医師再確認／請求停止条件にしない。未手修正行のsource未取得・版不明は正しい請求を判定できないため停止。手修正を保持した行でも薬剤／処置identity・数量など臨床差異があれば医師再確認gateを適用し、単なる金額差異と分ける。会計とプランの完全一致や全差異解消を新必須条件にしない。

実施・準備の元事実はD03等のownerを参照し、未実施／後日予定／中止を通常当日請求へ混ぜない。院内投薬注射は実施後、持帰薬／物販は渡す内容数量確定・準備完了で対象。実引渡し／在庫出庫は別。既存個別の中止料等は根拠を持つ別行とし、通常除外ルールで消さない。source状態の遅延応答や取消で金額を不明→0として確定しない。

### 実装差と設計検証

限定観察：`backend/internal/model/billing_confirmation.go`と`billing_confirmation_service.go`にはstatus/actor/時刻があるが、確認した臨床世代の対応は読取範囲で確認できない。`accounting_complete.go`と`accounting_complete_digest.go`は未請求版と冪等処理を持つが、確認世代の照合済み証拠にはならない。`billing/routes.go`のaccounting edit/createという既存route grantだけで確認医師資格を決定しない。API/DB適合・全writer・実機は未検証。

DC-A01：医師確認前準備可・請求確定／精算拒否、確認と金額変更を両順実行し再医師確認なし。DC-A02：薬剤／処置追加削除・数量変更と確認／確定を両順実行し、古いreceiptで確定不可。DC-A03：未手修正同期、手修正差異、確定後訂正の3分岐、手編集と遅延同期の両順。DC-A04：元削除・再追加・数量を戻す・複数元記録・取得失敗。DC-A05：確定後応答断・同key再送・監査失敗で二重請求／架空入金なし。DC-A06：未実施／準備済持帰り／実引渡し／中止料を独立検査。正本の既存医師確認・プラン同期・請求対象ACへの対応で、今回は実行していない。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../../../spec/domains/SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../../../spec/domains/accounting/lifecycle.md#doctor-review-sync-billing-20261017)。会計ページは医師確認前から準備し待ち・金額未確定を表示する。確定精算は医師確認後。会計タブチェックが医師確認、診療終了で会計待ちへ進み、受付カードの確認状態は別軸とする。確認後の金額だけの変更は警告了承で可、薬剤処置追加削除・数量変更は再確認まで停止。未確定未手修正はプラン自動同期、手修正済みは値保護と差異判断、確定後は専用訂正。会計からプランへの自動逆書換えなし。通常請求対象は実施済み、持帰薬物販は内容数量確定・準備完了で、予定・実引渡し・在庫出庫と分離する。


<a id="accounting-document-default-trace"></a>
## 帳票・初期値の補足AC接続

本文181ACとsubpage固有20ACは異なる母集団である。3画面・帳票契約に対し、後続の下記3ACの具体的な取得／初期適用を補足する。他のsubpage条件は既存の意味契約・正本参照を維持し、単価／割引認可・医師確認・確定後訂正・発行履歴を変更しない。

| 正本AC | 既存契約への技術接続・確認点 | 業務正本 |
|:---|:---|:---|
| AC-accounting-invoice-registration-number-output | 対象医院の登録番号設定と対象帳票／保存対象の関連を照合して出力へ渡す。請求と実受領を別に扱い、再発行・旧版への値適用は既存の発行履歴に従う。別医院の設定で補完しない。 | [正本](../../../spec/domains/accounting/pages/accounting-settlement.md) |
| AC-accounting-a4-document-output | 同じ保存対象の登録番号・税率別内訳を既存A4明細兼領収書projectionへ渡し、ブラウザ印刷／PDFを確認する。請求／実受領の別帳票を維持し、直接ロール印字・ドロアー連動を追加しない。 | [正本](../../../spec/domains/accounting/pages/accounting-settlement.md) |
| AC-accounting-owner-discount-default | 飼主の登録値引率を今回会計の初期値に使い、未登録時は0とする。既存の割引計算・認可に渡し、飼主設定の後日変更や再取得で手修正・確定snapshotを上書きしない。ペット別の新マスタを作らない。 | [正本](../../../spec/domains/accounting/pages/accounting-settlement.md) |

各行は既存ACの接続先を明確にするもので、新しい業務要件・実装完了・実機合格を表さない。具体的な既存wire／取得元との適合は実装工程で照合する。

<a id="finance-physical-candidates-20261019"></a>
## 財務残件の保存資源・実装適合候補（2026-10-19・内部提案）

FIN-R11〜R20の具体化。以下の資源名・列名・操作名はレビュー用候補であり、既存APIや採用済みDDLの説明ではない。既決の意味契約・認可・保持条件を変更せず、実schema/wireとデータ桁数・索引・実grant・移行・全writer・性能の適合は後続工程で照合する。本書だけでmigrationや新業務操作を許可しない。

### 候補の共通形と失敗

候補`financial_operation`のキーを `(clinic_id, domain, actor_ref, intent, operation_id)` とし、`intent, canonical_payload_digest, expected_versions, status, result_ref, actor_ref, committed_at`を保持する。この旧候補のexpected_versionsは一括してdigestへ入れない。EMR-483のbilling.price_acknowledgementがC06へ参加する場合、CAS/fenceだけの実行前提はfingerprintから除き、了承対象のbefore/after計算revision・対象・金額／条件はimmutableなContentへ含める。確認対象版や臨床世代も名称だけで一律除外せず、その操作の業務内容か現在照合前提かをfield単位で分類する。別financial operationの厳しい契約は適用範囲・C06との優先関係・adapterを参加表で明示し、無条件に変更しない。既存操作keyのlookupは書込再実行より先に行い、保存済み結果の参照は現在のread認可を検査する。同key別digestは内容競合。未検出・処理中・照会障害は保存失敗の確定ではなく、同key同内容へ照合する。確定拒否後の内容修正は旧結果を参照する新操作。receipt、必要監査、各owner参加の業務変更を一commitにする。実装方式は既存transaction/gate順を使い、domain別の新たな逆lock順を作らない。

候補APIの役割は `preview(intent, refs, versions, input)`、`commit(intent, operation_id, preview_ref, expected_versions, input)`、`get_result(operation_id)`。この名称の実在routeを主張せず、既存routeが必要意味を満たす時だけadapterへ対応させる。preview_ref単独を認可や完全性の証明にせず、保存直前の現在版・scope・認可を再照合する。結果不明中に新key再作成・実受領/返金の再実行を指示しない。

### FIN-R11：医師確認世代・手修正保護の候補資源

| 候補資源 | キー・値・制約 | 検証期待 |
|:--|:--|:--|
| `clinical_review_generation` | `(clinic_id, record_id)`、世代、対象薬剤/処置ID・数量の集合digest。変更writerは世代CAS。削除/追加で元に戻しても世代は戻さない | 単価/割引/表示順だけは世代不変、薬剤処置追加削除/数量は世代増分。D03の全dependency版を流用しない |
| `billing_clinical_generation` | `(clinic_id,billing_id)`、会計側薬剤処置変更世代・対象元record集合。D10直接変更とD03由来同期を同じ判定へ | 一つの親recordだけで確認済みと判定せず、現在集合を再検査 |
| `doctor_review_receipt` | `(clinic_id,review_id)`、billing_ref、record別世代集合、会計側世代、actor、時刻、操作key。履歴追記。actorは認証・医師確認権限から検査 | 現集合/両側世代とreceipt一致時だけ有効。金額変更了承は別receipt/会計版で、医師再確認は要求しない |
| `billing_source_binding` | `(clinic_id,billing_line_id)`、typed source ID・最後の適用版・行版・manual_override・原値。元行の有効請求帰属一意性は別制約 | 手編集と同期は行版CAS。先手編集なら自動上書きなし/差異、先同期なら手編集を再確認。確定済みは同期不可 |
| `billing_source_difference` | bindingとsource現在版、削除tombstone、差異field、臨床/金額区分、反映先会計版、解決操作ref | 金額差異のみを新しい確定禁止へしない。臨床差異は再確認gate。source不明を0円へしない |

確定は共通gate下で依存版集合を決め、下位lock前の参加順に従って現認可・有効確認・請求対象・行版・金額了承を検査する。同時同期/医師確認/確定を両先行順でDC-A01〜06へ接続。旧billing_confirmationのbool相当status/actor/timeだけでは世代成立を証明できない。

### FIN-R14：十進値・算定snapshot・発行版の候補

入力単価/数量/率は正規十進文字列から整数係数＋桁へ変換する候補。演算は任意精度整数/有理数で行い、行積のhalf-up、同scope割合額half-up、最大剰余、税率別floorを明示した段階関数にする。binary floatのepsilonで0.5判定しない。許容最大桁数は実値・API payload上限/overflow試験で確定し、未確認の業務金額上限を追加しない。桁超過・表現不能は明示拒否して丸め込まない。

`calculation_snapshot`は `(clinic_id,billing_id,revision)`、規則版・安定行identity/順序・元値・各scope条件/要求/実額・警告了承・除外/保険区分・配賦整数額・税率別合計/税・最終額を保持する候補。配賦同率端数の順序は、確定した行順ordinalと一意行IDの組を固定保存する案で、表示並替後も過去再計算結果を変えない。元請求の部分集合再請求は選択行の元snapshotと今回条件から計算し、元請求総税額を複製しない。

`document_edition`は `(clinic_id,document_id,edition)`、対象取引種別/ID・元取引番号・snapshot_ref・当時の宛名/登録番号/必要表示値・出力layout版・発行actor/timeを固定する候補。再発行は元取引番号を維持して発行履歴を追加し、新受領を作らない。修正前版を現在masterで再生成しない。税の分類・適格請求書全体の法的適合は別の確認者/根拠/判定を必要とする。

数値検査は正本の101×0.5→51→46→税4→税込50、1×0.5の2行→2円、100円配賦33/67、同率の安定順、全除外でも元正額残存、10%/8%別税を各caseでsnapshotまで照合する。

### FIN-R12・R13・R15・R20：限定した静的根拠と追従台帳

基準main `563353287ebb08af55056e3fbfba2aca3531e599` の限定読取。route宣言や一関数の存在は、呼出全経路・保存/監査・race・実grantの合格証拠ではない。

| 範囲・現根拠 | 限定して確認した事実 | 対応する後続検査 |
|:--|:--|:--|
| `backend/internal/billing/routes.go:95–106,114–146` | 見積create/successor/update/delete、確認GET/confirm/return、明細create/update/delete、accounting create/complete/update/correction/cancel/refund、closeを列挙。確認writeはaccounting edit、返金writeはaccounting createという現route | handler→service→repo→tx→監査/receipt→consumerを各入口で追跡。現在grantを新医師確認資格・新返金権限へ自動転用しない。GET確認の初期化もwriterとして配賦 |
| `backend/internal/model/billing_confirmation.go:17–34`、`billing_confirmation_service.go:104–124` | status/actor/time等の保存とGetOrCreateの作成分岐。読取modelに臨床世代集合はない | 上記候補との列/参照差、current認可、確認と子変更の共通gate、同keyreceiptを個別照合。全呼出からの欠如とは断定しない |
| `backend/internal/medicalrecord/hospitalization_discharge_tx.go:53–90` | CreateAccounting=falseは請求生成なし。既存非取消会計は再利用し、取消会計はSoftDeleteCancelled後に新規へ進む | 元取消ID/番号/元明細/確定歴が通常検索・履歴・再請求リンクから消えないかを専用検査。D10目標の理由付き選択再請求へ同等とみなさない |
| 同`:93–161` | carePlan由来待ち会計・明細・合計と退院会計監査をtransaction内で作成 | D05の実退院事実、ケア行、D10元明細identityと新番号/元取消refを対応。専用終了と通常医師確認を区別し、自動返金/在庫戻しを起動しない |

全writer台帳の1行は `入口(route/import/repair/job/GET初期化)→関数→対象表/資源→transaction initiator→予約/臨床/資金participant→取得順→認可→期待版→監査sink→receipt/照会→差分→受入case` とする。上表を全writer完了へ水増ししない。共通gate証拠は他領域と共有し、金銭の入口責任だけをFIN-R12で追う。

移行候補manifestは旧ID/clinic/owner/source_kind/source_ID、旧状態、請求額/受領/返金/配分/預り/確定歴、日時基準、method安定識別、snapshot有無、既存明示保険/除外/欠落の分類、照合状態・未解決理由、新ref/旧ref・適用版を持つ。分類別に件数と金額を前後集計し、FK/一意性・元資金上限・二重帰属・同一操作再実行・復旧参照を検査する。新規保険4種初期値と旧欠落値の一括変換を同一化しない。欠落保険の意味に資料不足が残った時だけ、実データ分類・影響と既存回答の原文を添えて判断質問にする。データ未読の段階で一律対象外を提案しない。


<a id="finance-ac-execution-trace-20261019"></a>
## 財務ACの実行準備対応（2026-10-19・未実行）

FIN-R01〜R10の文書精査として、本文181ACとsubpage固有20ACを別母集団のまま下表へ収録した。新しいAC、期待結果の置換、実行済み判定ではない。各行のGiven/When/Thenは参照先正本の当該AC全文を用い、検証群のfixtureはその前提を満たすものだけを組み合わせる。F01等は本節内の検証群ラベルで、製品仕様IDや既存scenario IDを再採番しない。

既存S/Vは旧実装の観測を含む。該当画面・APIが目標未対応なら、そのケースは実行前の未適合／BLOCKEDとして残す。全ACの実行結果、実UI/DTO全項目、全writer、税務・実決済・物理印刷の適合を本表で証明しない。フォーム外の資金配分・医師確認・帳票出力などは47項目に架空のfieldKeyを付けず、AC/操作/receipt単位で記録する。

| 検証群 | 受渡し先 | fixture・照合対象 |
|:--|:--|:--|
| F01 請求・任意額回収 | S08 #1/#6/#9、V02 §1、S21 | 同飼主2pet、当日/過去明細、10,000請求/0・3,000入金/後日2,000、空明細と実在0円を別fixture。元ID・請求状態・実受領・配分・残・確定歴を比較 |
| F02 訂正・取消・実返金・再請求 | S08 #2–5/#10、V02 §3–4、S21 | 入金3,000/既返金1,000の残上限2,000、10,000→9,000、取消元の選択再請求、原資不明を別fixture。元資金・元/新番号・履歴・実移動を比較 |
| F03 預り・充当・解除・再配分 | S08 #6/#8–10の拡張、S21 | 飼主預り5,000、充当3,000、取消/減額/既返金、別pet/予約・別owner/clinic。元受領日・有効配分・残・確認・二重使用を比較 |
| F04 算定・snapshot・単価/割引 | V02 §1–2、S20 | 業務正本の各数値例をそのままcase入力に使用。行積/割合/定額/最大剰余/税率別税/0円・手修正を段階別に比較。小数量の臨床値不変 |
| F05 窓口保険・品目/方法master認可 | S20、V04の対象master、V02 §1 | 4追加品目と薬剤/検査の明示値、未確認/不明、1pet1契約、対象grantあり/なし/失効/別clinic。保存版・独立区分・外部作用0を比較 |
| F06 帳票・未納・締め/期間 | S08 #7–10、S09、S33 | 売上なし過去未納、実受領と充当、元/差分、複数pet/共通料金、AM締め/PM未締め、JST境界を別fixture。画面/PDF/履歴を照合、物理印刷は別結果 |
| F07 医師確認・同期・請求対象 | DC-A01–06、V02 §1行編集、S21 | 未確認/確認済×金額のみ/薬剤処置追加削除/数量変更×未手修正/手修正/確定済。確認先行/変更先行両順と元削除・再追加・取得失敗を検査 |
| F08 臨床・専用記録の受渡し境界 | S11、D05退院検証、S21 | 施術0円/手修正値引き・退院既存/取消会計・死亡後生前分・記録復元を分離。元ID/版・金銭/在庫副作用・履歴不変を比較 |
| F09 新受領の支払方法停止 | V02 §1、V04支払方法、S21 | 有効選択→保存前停止、保存済み成功→停止→同key照会、新入金/後日回収/預り受領と既資金配分/返金を区別 |

実行記録には `AC-ID / source revision / fixture ID / scenario・操作 / formId.fieldKey.Fx（実在する場合） / API・保存対象 / 前後snapshot・監査・operation receipt / 結果 / 未実行理由 / 担当・時刻` を対応させる。同一証跡を複数ACへ参照してよいが、各ACのThenを個別に判定する。競合を含む各writeは同key同内容、同key別内容、応答消失、遅延commit中の未検出、認可失効、監査失敗を追加し、別key再作成で回避しない。

### ライフサイクル本文181AC

以下の正本行位置は基準main `563353287ebb08af55056e3fbfba2aca3531e599` の所在補助で、IDを主参照とする。全行の実機状態は未実行。

| 正本AC | 所在（基準版） | 検証群 | 個別結果キー |
|:--|:--|:--|:--|
| AC-accounting-trimming-adjustment-once | lifecycle.md:305 | F08 | `AC-accounting-trimming-adjustment-once` |
| AC-accounting-trimming-zero-line | lifecycle.md:306 | F08 | `AC-accounting-trimming-zero-line` |
| AC-accounting-trimming-source-refresh | lifecycle.md:307 | F08 | `AC-accounting-trimming-source-refresh` |
| AC-accounting-record-edit-no-money-undo | lifecycle.md:308 | F08 | `AC-accounting-record-edit-no-money-undo` |
| AC-accounting-owner-multi-pet | lifecycle.md:309 | F01 | `AC-accounting-owner-multi-pet` |
| AC-accounting-selected-unbilled | lifecycle.md:310 | F01 | `AC-accounting-selected-unbilled` |
| AC-accounting-old-bill-collection | lifecycle.md:311 | F01 | `AC-accounting-old-bill-collection` |
| AC-accounting-no-invented-allocation | lifecycle.md:312 | F01 | `AC-accounting-no-invented-allocation` |
| AC-accounting-unpaid-finalize | lifecycle.md:313 | F01 | `AC-accounting-unpaid-finalize` |
| AC-accounting-partial-finalize | lifecycle.md:314 | F01 | `AC-accounting-partial-finalize` |
| AC-accounting-later-payment | lifecycle.md:315 | F01 | `AC-accounting-later-payment` |
| AC-accounting-zero-no-fake-payment | lifecycle.md:316 | F01 | `AC-accounting-zero-no-fake-payment` |
| AC-accounting-finalization-undo-preserved | lifecycle.md:317 | F01 | `AC-accounting-finalization-undo-preserved` |
| AC-accounting-correction-authorization | lifecycle.md:318 | F02 | `AC-accounting-correction-authorization` |
| AC-accounting-correction-history | lifecycle.md:319 | F02 | `AC-accounting-correction-history` |
| AC-accounting-refund-received-cap | lifecycle.md:320 | F02 | `AC-accounting-refund-received-cap` |
| AC-accounting-refund-method-required | lifecycle.md:321 | F02 | `AC-accounting-refund-method-required` |
| AC-accounting-cancel-no-money-stock | lifecycle.md:322 | F02 | `AC-accounting-cancel-no-money-stock` |
| AC-accounting-post-close-preserve | lifecycle.md:323 | F06 | `AC-accounting-post-close-preserve` |
| AC-accounting-correction-no-cash-movement | lifecycle.md:324 | F06 | `AC-accounting-correction-no-cash-movement` |
| AC-accounting-master-change-stable | lifecycle.md:325 | F04 | `AC-accounting-master-change-stable` |
| AC-accounting-manual-unit-price | lifecycle.md:326 | F04 | `AC-accounting-manual-unit-price` |
| AC-accounting-discounts-combinable | lifecycle.md:327 | F04 | `AC-accounting-discounts-combinable` |
| AC-accounting-clinical-no-auto-finance | lifecycle.md:328 | F08 | `AC-accounting-clinical-no-auto-finance` |
| AC-accounting-oldest-default | lifecycle.md:329 | F03 | `AC-accounting-oldest-default` |
| AC-accounting-allocation-staff-change | lifecycle.md:330 | F03 | `AC-accounting-allocation-staff-change` |
| AC-accounting-collection-change-return | lifecycle.md:331 | F01 | `AC-accounting-collection-change-return` |
| AC-accounting-advance-owner-reservation | lifecycle.md:332 | F03 | `AC-accounting-advance-owner-reservation` |
| AC-accounting-advance-unallocated-balance | lifecycle.md:333 | F03 | `AC-accounting-advance-unallocated-balance` |
| AC-accounting-advance-apply-no-double-receipt | lifecycle.md:334 | F03 | `AC-accounting-advance-apply-no-double-receipt` |
| AC-accounting-advance-surplus-pending | lifecycle.md:335 | F03 | `AC-accounting-advance-surplus-pending` |
| AC-accounting-advance-actual-refund | lifecycle.md:336 | F03 | `AC-accounting-advance-actual-refund` |
| AC-accounting-advance-duplicate-use | lifecycle.md:337 | F03 | `AC-accounting-advance-duplicate-use` |
| AC-accounting-advance-context-mismatch | lifecycle.md:338 | F03 | `AC-accounting-advance-context-mismatch` |
| AC-accounting-paid-reduction-refund-pending | lifecycle.md:339 | F02 | `AC-accounting-paid-reduction-refund-pending` |
| AC-accounting-paid-reduction-actual-refund | lifecycle.md:340 | F02 | `AC-accounting-paid-reduction-actual-refund` |
| AC-accounting-partial-paid-reduction | lifecycle.md:341 | F02 | `AC-accounting-partial-paid-reduction` |
| AC-accounting-payment-method-correction-no-money | lifecycle.md:342 | F02 | `AC-accounting-payment-method-correction-no-money` |
| AC-accounting-percent-before-fixed | lifecycle.md:343 | F04 | `AC-accounting-percent-before-fixed` |
| AC-accounting-insurance-staff-review | lifecycle.md:344 | F05 | `AC-accounting-insurance-staff-review` |
| AC-accounting-invoice-receipt-separated | lifecycle.md:345 | F06 | `AC-accounting-invoice-receipt-separated` |
| AC-accounting-document-reissue-history | lifecycle.md:346 | F06 | `AC-accounting-document-reissue-history` |
| AC-accounting-aged-debt-without-current-sales | lifecycle.md:347 | F06 | `AC-accounting-aged-debt-without-current-sales` |
| AC-accounting-original-bill-and-movement-date | lifecycle.md:348 | F06 | `AC-accounting-original-bill-and-movement-date` |
| AC-accounting-deposit-without-reservation | lifecycle.md:349 | F03 | `AC-accounting-deposit-without-reservation` |
| AC-accounting-deposit-other-reservation | lifecycle.md:350 | F03 | `AC-accounting-deposit-other-reservation` |
| AC-accounting-deposit-other-pet | lifecycle.md:351 | F03 | `AC-accounting-deposit-other-pet` |
| AC-accounting-deposit-other-owner-denied | lifecycle.md:352 | F03 | `AC-accounting-deposit-other-owner-denied` |
| AC-accounting-deposit-allocation-confirmation | lifecycle.md:353 | F03 | `AC-accounting-deposit-allocation-confirmation` |
| AC-accounting-deposit-available-balance | lifecycle.md:354 | F03 | `AC-accounting-deposit-available-balance` |
| AC-accounting-percentage-summed | lifecycle.md:355 | F04 | `AC-accounting-percentage-summed` |
| AC-accounting-discount-then-insurance | lifecycle.md:356 | F04 | `AC-accounting-discount-then-insurance` |
| AC-accounting-overdiscount-unresolved | lifecycle.md:357 | F04 | `AC-accounting-overdiscount-unresolved` |
| AC-accounting-same-day-created-order | lifecycle.md:358 | F03 | `AC-accounting-same-day-created-order` |
| AC-accounting-saved-reallocation-reason | lifecycle.md:359 | F03 | `AC-accounting-saved-reallocation-reason` |
| AC-accounting-reallocation-no-new-money | lifecycle.md:360 | F03 | `AC-accounting-reallocation-no-new-money` |
| AC-accounting-reallocation-correction-conflict | lifecycle.md:361 | F03 | `AC-accounting-reallocation-correction-conflict` |
| AC-accounting-noncash-excess-refund-pending | lifecycle.md:362 | F01 | `AC-accounting-noncash-excess-refund-pending` |
| AC-accounting-refund-different-method | lifecycle.md:363 | F02 | `AC-accounting-refund-different-method` |
| AC-accounting-refund-no-amount-extra-approval | lifecycle.md:364 | F02 | `AC-accounting-refund-no-amount-extra-approval` |
| AC-accounting-deposit-receipt | lifecycle.md:365 | F03 / F06 | `AC-accounting-deposit-receipt` |
| AC-accounting-final-document-deposit-current | lifecycle.md:366 | F06 | `AC-accounting-final-document-deposit-current` |
| AC-accounting-refund-original-transaction | lifecycle.md:367 | F06 | `AC-accounting-refund-original-transaction` |
| AC-accounting-reissue-original-number | lifecycle.md:368 | F06 | `AC-accounting-reissue-original-number` |
| AC-accounting-line-and-bill-discount | lifecycle.md:369 | F04 | `AC-accounting-line-and-bill-discount` |
| AC-accounting-overdiscount-warning | lifecycle.md:370 | F04 | `AC-accounting-overdiscount-warning` |
| AC-accounting-overdiscount-post-warning-pending | lifecycle.md:371 | F04 | `AC-accounting-overdiscount-post-warning-pending` |
| AC-accounting-cancel-restore-deposit | lifecycle.md:372 | F03 | `AC-accounting-cancel-restore-deposit` |
| AC-accounting-cancel-deposit-no-cash-refund | lifecycle.md:373 | F03 | `AC-accounting-cancel-deposit-no-cash-refund` |
| AC-accounting-cancel-deposit-once-atomic | lifecycle.md:374 | F03 | `AC-accounting-cancel-deposit-once-atomic` |
| AC-accounting-line-before-invoice | lifecycle.md:375 | F04 | `AC-accounting-line-before-invoice` |
| AC-accounting-discount-scope-separated | lifecycle.md:376 | F04 | `AC-accounting-discount-scope-separated` |
| AC-accounting-percent-discount-round-yen | lifecycle.md:377 | F04 | `AC-accounting-percent-discount-round-yen` |
| AC-accounting-invoice-proportional-remainder | lifecycle.md:378 | F04 | `AC-accounting-invoice-proportional-remainder` |
| AC-accounting-invoice-zero-discount-base | lifecycle.md:379 | F04 | `AC-accounting-invoice-zero-discount-base` |
| AC-accounting-overdiscount-line-no-spill | lifecycle.md:380 | F04 | `AC-accounting-overdiscount-line-no-spill` |
| AC-accounting-overdiscount-invoice-effective-cap | lifecycle.md:381 | F04 | `AC-accounting-overdiscount-invoice-effective-cap` |
| AC-accounting-tax-invoice-rate-once | lifecycle.md:382 | F04 | `AC-accounting-tax-invoice-rate-once` |
| AC-accounting-tax-rate-separated | lifecycle.md:383 | F04 | `AC-accounting-tax-rate-separated` |
| AC-accounting-tax-exclusive-consumer-inclusive | lifecycle.md:384 | F04 | `AC-accounting-tax-exclusive-consumer-inclusive` |
| AC-accounting-tax-category-snapshot | lifecycle.md:385 | F04 | `AC-accounting-tax-category-snapshot` |
| AC-accounting-calculation-audit-replay | lifecycle.md:386 | F04 | `AC-accounting-calculation-audit-replay` |
| AC-accounting-insurance-multiple-carriers | lifecycle.md:387 | F05 | `AC-accounting-insurance-multiple-carriers` |
| AC-accounting-invoice-exclusion-master | lifecycle.md:388 | F04 | `AC-accounting-invoice-exclusion-master` |
| AC-accounting-invoice-excluded-base-and-allocation | lifecycle.md:389 | F04 | `AC-accounting-invoice-excluded-base-and-allocation` |
| AC-accounting-invoice-all-excluded-zero-base | lifecycle.md:390 | F04 | `AC-accounting-invoice-all-excluded-zero-base` |
| AC-accounting-invoice-exclusion-not-insurance | lifecycle.md:391 | F04 | `AC-accounting-invoice-exclusion-not-insurance` |
| AC-accounting-invoice-exclusion-snapshot-stable | lifecycle.md:392 | F04 | `AC-accounting-invoice-exclusion-snapshot-stable` |
| AC-accounting-document-owner-addressee | lifecycle.md:393 | F06 | `AC-accounting-document-owner-addressee` |
| AC-accounting-document-pet-separated | lifecycle.md:394 | F06 | `AC-accounting-document-pet-separated` |
| AC-accounting-document-correction-original-delta | lifecycle.md:395 | F06 | `AC-accounting-document-correction-original-delta` |
| AC-accounting-document-partial-refund-original-delta | lifecycle.md:396 | F06 | `AC-accounting-document-partial-refund-original-delta` |
| AC-accounting-master-admin-edit | lifecycle.md:397 | F05 | `AC-accounting-master-admin-edit` |
| AC-accounting-master-authorized-staff-edit | lifecycle.md:398 | F05 | `AC-accounting-master-authorized-staff-edit` |
| AC-accounting-master-ungranted-staff-denied | lifecycle.md:399 | F05 | `AC-accounting-master-ungranted-staff-denied` |
| AC-accounting-master-permission-rechecked | lifecycle.md:400 | F05 | `AC-accounting-master-permission-rechecked` |
| AC-accounting-insurance-staff-result-burden | lifecycle.md:401 | F05 | `AC-accounting-insurance-staff-result-burden` |
| AC-accounting-insurance-unknown-not-valid | lifecycle.md:402 | F05 | `AC-accounting-insurance-unknown-not-valid` |
| AC-accounting-insurance-window-not-coverage | lifecycle.md:403 | F05 | `AC-accounting-insurance-window-not-coverage` |
| AC-accounting-insurance-one-contract-per-pet | lifecycle.md:404 | F05 | `AC-accounting-insurance-one-contract-per-pet` |
| AC-accounting-insurance-window-only-scope | lifecycle.md:405 | F05 | `AC-accounting-insurance-window-only-scope` |
| AC-accounting-insurance-local-correction-no-external | lifecycle.md:406 | F05 | `AC-accounting-insurance-local-correction-no-external` |
| AC-accounting-insurance-manual-no-formula-prerequisite | lifecycle.md:407 | F05 | `AC-accounting-insurance-manual-no-formula-prerequisite` |
| AC-accounting-deposit-reduction-restore | lifecycle.md:408 | F03 | `AC-accounting-deposit-reduction-restore` |
| AC-accounting-deposit-mixed-reduction-choice | lifecycle.md:409 | F03 | `AC-accounting-deposit-mixed-reduction-choice` |
| AC-accounting-deposit-reduction-no-money | lifecycle.md:410 | F03 | `AC-accounting-deposit-reduction-no-money` |
| AC-accounting-deposit-reduction-current-once | lifecycle.md:411 | F03 | `AC-accounting-deposit-reduction-current-once` |
| AC-accounting-deposit-no-auto-expiry | lifecycle.md:412 | F03 | `AC-accounting-deposit-no-auto-expiry` |
| AC-accounting-deposit-clinic-owner-boundary | lifecycle.md:413 | F03 | `AC-accounting-deposit-clinic-owner-boundary` |
| AC-accounting-rebill-selected-new-link | lifecycle.md:414 | F02 | `AC-accounting-rebill-selected-new-link` |
| AC-accounting-rebill-no-duplicate | lifecycle.md:415 | F02 | `AC-accounting-rebill-no-duplicate` |
| AC-accounting-rebill-no-money-copy | lifecycle.md:416 | F02 | `AC-accounting-rebill-no-money-copy` |
| AC-accounting-rebill-history-undo | lifecycle.md:417 | F02 | `AC-accounting-rebill-history-undo` |
| AC-accounting-combined-collection-allocation | lifecycle.md:418 | F01 | `AC-accounting-combined-collection-allocation` |
| AC-accounting-combined-collection-no-rebilling | lifecycle.md:419 | F01 | `AC-accounting-combined-collection-no-rebilling` |
| AC-accounting-combined-collection-order-override | lifecycle.md:420 | F01 | `AC-accounting-combined-collection-order-override` |
| AC-accounting-combined-collection-current-state | lifecycle.md:421 | F01 | `AC-accounting-combined-collection-current-state` |
| AC-accounting-collection-dunning-writeoff-out | lifecycle.md:422 | F01 | `AC-accounting-collection-dunning-writeoff-out` |
| AC-accounting-document-common-charges | lifecycle.md:423 | F06 | `AC-accounting-document-common-charges` |
| AC-accounting-common-charge-provenance | lifecycle.md:424 | F06 | `AC-accounting-common-charge-provenance` |
| AC-accounting-rebill-original-price-tax-discount | lifecycle.md:425 | F02 | `AC-accounting-rebill-original-price-tax-discount` |
| AC-accounting-rebill-staff-adjustment-history | lifecycle.md:426 | F02 | `AC-accounting-rebill-staff-adjustment-history` |
| AC-accounting-rebill-old-payment-confirmed | lifecycle.md:427 | F03 | `AC-accounting-rebill-old-payment-confirmed` |
| AC-accounting-rebill-old-payment-no-receipt | lifecycle.md:428 | F03 | `AC-accounting-rebill-old-payment-no-receipt` |
| AC-accounting-rebill-old-payment-current-cap | lifecycle.md:429 | F03 | `AC-accounting-rebill-old-payment-current-cap` |
| AC-accounting-rebill-restored-deposit-separate | lifecycle.md:430 | F03 | `AC-accounting-rebill-restored-deposit-separate` |
| AC-accounting-rebill-insurance-rechecked | lifecycle.md:431 | F02 / F05 | `AC-accounting-rebill-insurance-rechecked` |
| AC-accounting-rebill-insurance-scope-preserved | lifecycle.md:432 | F02 / F05 | `AC-accounting-rebill-insurance-scope-preserved` |
| AC-accounting-refund-source-unknown-blocked | lifecycle.md:433 | F02 | `AC-accounting-refund-source-unknown-blocked` |
| AC-accounting-refund-source-record-before-finalize | lifecycle.md:434 | F02 | `AC-accounting-refund-source-record-before-finalize` |
| AC-accounting-refund-source-history-no-double | lifecycle.md:435 | F02 | `AC-accounting-refund-source-history-no-double` |
| AC-accounting-refund-source-scope-not-method | lifecycle.md:436 | F02 | `AC-accounting-refund-source-scope-not-method` |
| AC-accounting-empty-invoice-blocked | lifecycle.md:437 | F01 | `AC-accounting-empty-invoice-blocked` |
| AC-accounting-zero-invoice-line-reason | lifecycle.md:438 | F01 | `AC-accounting-zero-invoice-line-reason` |
| AC-accounting-collection-no-empty-invoice | lifecycle.md:439 | F01 | `AC-accounting-collection-no-empty-invoice` |
| AC-accounting-pre-discount-round-half | lifecycle.md:440 | F04 | `AC-accounting-pre-discount-round-half` |
| AC-accounting-pre-discount-round-per-line | lifecycle.md:441 | F04 | `AC-accounting-pre-discount-round-per-line` |
| AC-accounting-pre-discount-round-before-discount | lifecycle.md:442 | F04 | `AC-accounting-pre-discount-round-before-discount` |
| AC-accounting-pre-discount-round-provenance | lifecycle.md:443 | F04 | `AC-accounting-pre-discount-round-provenance` |
| AC-accounting-exclusion-default-included | lifecycle.md:444 | F04 | `AC-accounting-exclusion-default-included` |
| AC-accounting-exclusion-missing-included | lifecycle.md:445 | F04 | `AC-accounting-exclusion-missing-included` |
| AC-accounting-exclusion-existing-preserved | lifecycle.md:446 | F04 | `AC-accounting-exclusion-existing-preserved` |
| AC-accounting-master-draft-no-auto-update | lifecycle.md:447 | F04 | `AC-accounting-master-draft-no-auto-update` |
| AC-accounting-master-explicit-reapply | lifecycle.md:448 | F04 | `AC-accounting-master-explicit-reapply` |
| AC-accounting-master-reapply-finalized-boundary | lifecycle.md:449 | F04 | `AC-accounting-master-reapply-finalized-boundary` |
| AC-accounting-master-reapply-current-check | lifecycle.md:450 | F04 | `AC-accounting-master-reapply-current-check` |
| AC-accounting-posthumous-settlement-history | lifecycle.md:451 | F08 | `AC-accounting-posthumous-settlement-history` |
| AC-accounting-posthumous-owner-no-side-effect | lifecycle.md:452 | F08 | `AC-accounting-posthumous-owner-no-side-effect` |
| AC-accounting-payment-inactive-new-receipt | lifecycle.md:453 | F09 | `AC-accounting-payment-inactive-new-receipt` |
| AC-accounting-payment-inactive-save-recheck | lifecycle.md:454 | F09 | `AC-accounting-payment-inactive-save-recheck` |
| AC-accounting-payment-inactive-history-aggregation | lifecycle.md:455 | F09 | `AC-accounting-payment-inactive-history-aggregation` |
| AC-accounting-payment-inactive-existing-allocation | lifecycle.md:456 | F09 | `AC-accounting-payment-inactive-existing-allocation` |
| AC-accounting-payment-inactive-refund-correction-boundary | lifecycle.md:457 | F09 | `AC-accounting-payment-inactive-refund-correction-boundary` |
| AC-accounting-payment-inactive-no-receipt | lifecycle.md:458 | F09 | `AC-accounting-payment-inactive-no-receipt` |
| AC-accounting-payment-inactive-idempotent-result | lifecycle.md:459 | F09 | `AC-accounting-payment-inactive-idempotent-result` |
| AC-accounting-item-insurance-four-masters | lifecycle.md:465 | F05 | `AC-accounting-item-insurance-four-masters` |
| AC-accounting-item-insurance-existing-med-exam | lifecycle.md:466 | F05 | `AC-accounting-item-insurance-existing-med-exam` |
| AC-accounting-item-insurance-not-guarantee | lifecycle.md:467 | F05 | `AC-accounting-item-insurance-not-guarantee` |
| AC-accounting-item-insurance-independent-discount | lifecycle.md:468 | F05 | `AC-accounting-item-insurance-independent-discount` |
| AC-accounting-item-insurance-authorized-save | lifecycle.md:469 | F05 | `AC-accounting-item-insurance-authorized-save` |
| AC-accounting-item-insurance-snapshot-stable | lifecycle.md:470 | F05 | `AC-accounting-item-insurance-snapshot-stable` |
| AC-accounting-item-insurance-explicit-reapply | lifecycle.md:471 | F05 | `AC-accounting-item-insurance-explicit-reapply` |
| AC-accounting-item-insurance-finalized-correction | lifecycle.md:472 | F05 | `AC-accounting-item-insurance-finalized-correction` |
| AC-accounting-item-insurance-default-unresolved | lifecycle.md:473 | F05 | `AC-accounting-item-insurance-default-unresolved` |
| AC-accounting-item-insurance-legacy-missing | lifecycle.md:474 | F05 | `AC-accounting-item-insurance-legacy-missing` |
| AC-accounting-item-insurance-read-failure | lifecycle.md:475 | F05 | `AC-accounting-item-insurance-read-failure` |
| AC-accounting-item-insurance-atomic-retry | lifecycle.md:476 | F05 | `AC-accounting-item-insurance-atomic-retry` |
| AC-accounting-before-doctor-review | lifecycle.md:556 | F07 | `AC-accounting-before-doctor-review` |
| AC-accounting-review-card-axis | lifecycle.md:557 | F07 | `AC-accounting-review-card-axis` |
| AC-accounting-reviewed-price-only | lifecycle.md:558 | F07 | `AC-accounting-reviewed-price-only` |
| AC-accounting-reviewed-clinical-change | lifecycle.md:559 | F07 | `AC-accounting-reviewed-clinical-change` |
| AC-accounting-plan-sync-three-branches | lifecycle.md:560 | F07 | `AC-accounting-plan-sync-three-branches` |
| AC-accounting-billable-boundary | lifecycle.md:561 | F07 | `AC-accounting-billable-boundary` |
| AC-accounting-ready-not-stock | lifecycle.md:562 | F07 | `AC-accounting-ready-not-stock` |
| AC-accounting-individual-discount-permission | lifecycle.md:574 | F04 / F05 | `AC-accounting-individual-discount-permission` |
| AC-accounting-manual-change-unreconciled-completion | lifecycle.md:594 | F01 | `AC-accounting-manual-change-unreconciled-completion` |
| AC-accounting-manual-change-no-fabricated-adjustment | lifecycle.md:595 | F01 | `AC-accounting-manual-change-no-fabricated-adjustment` |
| AC-accounting-manual-change-preserves-partial-and-noncash | lifecycle.md:596 | F01 | `AC-accounting-manual-change-preserves-partial-and-noncash` |
| AC-accounting-excess-input-correction | lifecycle.md:616 | F01 | `AC-accounting-excess-input-correction` |
| AC-accounting-excess-actual-refund-once | lifecycle.md:617 | F02 | `AC-accounting-excess-actual-refund-once` |
| AC-accounting-excess-consented-deposit-single-receipt | lifecycle.md:618 | F03 | `AC-accounting-excess-consented-deposit-single-receipt` |
| AC-accounting-input-gross-preserved | lifecycle.md:629 | F04 | `AC-accounting-input-gross-preserved` |
| AC-accounting-line-discount-independent | lifecycle.md:630 | F04 | `AC-accounting-line-discount-independent` |
| AC-accounting-calculation-boundary-reconcile | lifecycle.md:631 | F04 | `AC-accounting-calculation-boundary-reconcile` |

### subpage固有20AC（本文とは別集計）

| 正本AC | 所在（基準版） | 検証群 | 個別結果キー |
|:--|:--|:--|:--|
| AC-accounting-daily-owner-pet-column-order | pages/accounting-list.md:164 | F06 | `AC-accounting-daily-owner-pet-column-order` |
| AC-accounting-line-name-edit | pages/accounting-settlement.md:90 | F04 | `AC-accounting-line-name-edit` |
| AC-accounting-stale-unbilled-completion-rejection | pages/accounting-settlement.md:184 | F01 | `AC-accounting-stale-unbilled-completion-rejection` |
| AC-accounting-invoice-registration-number-output | pages/accounting-settlement.md:243 | F06 | `AC-accounting-invoice-registration-number-output` |
| AC-accounting-merchandise-master-selection | pages/accounting-settlement.md:251 | F05 | `AC-accounting-merchandise-master-selection` |
| AC-accounting-owner-unpaid-summary | pages/accounting-settlement.md:259 | F06 | `AC-accounting-owner-unpaid-summary` |
| AC-accounting-a4-document-output | pages/accounting-settlement.md:265 | F06 | `AC-accounting-a4-document-output` |
| AC-accounting-owner-discount-default | pages/accounting-settlement.md:271 | F04 | `AC-accounting-owner-discount-default` |
| AC-accounting-line-unit-quantity-amount | pages/accounting-settlement.md:277 | F04 | `AC-accounting-line-unit-quantity-amount` |
| AC-accounting-manual-line-amount-quantity-edit | pages/accounting-settlement.md:283 | F04 | `AC-accounting-manual-line-amount-quantity-edit` |
| AC-accounting-manual-other-create | pages/accounting-settlement.md:291 | F04 | `AC-accounting-manual-other-create` |
| AC-accounting-owner-purchase-review | pages/accounting-settlement.md:299 | F06 | `AC-accounting-owner-purchase-review` |
| AC-accounting-unpaid-default-period | pages/unpaid-list.md:148 | F06 | `AC-accounting-unpaid-default-period` |
| AC-unpaid-source-month | pages/unpaid-list.md:189 | F06 | `AC-unpaid-source-month` |
| AC-unpaid-source-reconcile | pages/unpaid-list.md:190 | F06 | `AC-unpaid-source-reconcile` |
| AC-unpaid-source-adjustment-boundary | pages/unpaid-list.md:191 | F06 | `AC-unpaid-source-adjustment-boundary` |
| AC-accounting-unpaid-display-modes | pages/unpaid-list.md:208 | F06 | `AC-accounting-unpaid-display-modes` |
| AC-accounting-unpaid-mode-label | pages/unpaid-list.md:214 | F06 | `AC-accounting-unpaid-mode-label` |
| AC-accounting-unpaid-latest-date | pages/unpaid-list.md:220 | F06 | `AC-accounting-unpaid-latest-date` |
| AC-accounting-unpaid-carryover-columns | pages/unpaid-list.md:226 | F06 | `AC-accounting-unpaid-carryover-columns` |

## 10/19横断検索・同種操作の設計接続

[共通adapter](../../shared/integration-contracts.md#cross-ui-search-adapter-20261019)を本領域の[適用台帳](../../../spec/shared/interaction-scope.md)L/M/H/S consumerへ接続する。draft/applied・独立filter・取得scope/世代・復帰/失敗状態を分離し、検索ボタンだけでキーワードを適用。既存API検索と取得済み内絞込みの責務、ページング/上限、認可・保存方式を維持する。E/N操作へ部品単位で一括適用しない。実装/全caller/実機未確認。

## EMR-367／447の確定契約への設計接続

[税込額維持・行別値引きの正本と追加AC](../../../spec/domains/accounting/lifecycle.md#accounting-input-gross-preservation)を優先する。元の入力額・税込／税抜の入力区分・行ごとの値引率を識別して保持する案とし、丸めた税抜額だけを唯一の入力へ置換しない。既存の税抜入力用計算・全体配賦案を内税入力へ無条件適用しない。明細計算→税集計→請求額→受領／釣銭／繰越を分離し、小額手動値引きから自動10円未満切捨てを導出しない。

小数数量の行丸め、内税換算、率と定額値引きの併用、混在税率の整合を受入残件とする。既決順序は照合例と矛盾しないが、旧システムの唯一式・現行診療システムとの同等性を検証済みとはしない。追加資料の原数値・画像・Driveリンクは含めない。条件付きSTG判断の成立やデータ操作許可も未確認である。

品目保険の新規明細への初期値／参照版受渡しはEMR-454〜456（共有owner454）の実装境界。旧欠落の推測、保存済み行の自動再適用、健診保存からの会計生成は追加しない。今回のmodelの型対応は保存形状の整合であり、calculator・API・各writer・利用開始の完成証拠ではない。
