# 会計初回〜第十群と追加合意の技術仕様

読取基準：現行実装の観察はPR #670の41fcf3eb0a8befb7c0b5f613d1df749a60e30be8（起点main 7309b3a）時点。後続mainの修正有無・実装適合を断言せず、承認済みの目標契約と区別する。支払方法停止に関する追加読取は下記の専用行にheadと日付を明記し、既存観察全体の再確認とは扱わない。

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | design.accounting.partial／技術設計 |
| 決定状態 | 業務9FRは合意済み。本書の実現方式は提案設計、未決業務を補完しない |
| 実装状態 | 未確認。静的に確認した現行制約は下記。API wire／DB正本は変更なし |
| 実機検証状態 | 未実施。移行・並行性・性能も未検証 |
| 文書確認日 | 2026-10-15 |

[業務・AC正本](../../../spec/domains/accounting/lifecycle.md)と[語彙](../../../spec/ontology/accounting.md)を参照。対象は初回5点〜第十群と追加の死亡後精算・[支払方法停止](#accounting-inactive-payment-design)のD10文書設計。完成境界は[末尾](#accounting-document-completion)へ。実装適合／利用開始を表さない。

<a id="accounting-current-gap"></a>
## 現行実装と確定業務の差分

| 調査対象 | 現行の静的確認 | 目標業務との相違／残件 |
|:---|:---|:---|
| [確定入力・元カルテ照合](../../../../backend/internal/billing/accounting_complete.go) | ヘッダは単一ペット／カルテ。診療項目が複数カルテにまたがる場合は拒否 | 同飼主の複数ペット・当日／過去の選択を表現する明細参照へ設計が必要。単一患者への付替えは禁止 |
| [確定transaction](../../../../backend/internal/billing/accounting_complete_tx.go)・[支払UI](../../../../frontend/src/features/accounting/components/PaymentCard.tsx) | 支払内訳を請求に一致させる経路。初回部分入金は拒否 | 未納／部分入金確定・後日入金の業務は現行制約と異なる。[既存S08](../../../ops/testing/scenarios/S08-accounting-corrections.md)を新仕様の合格根拠にしない |
| [支払方法UI](../../../../frontend/src/features/accounting/components/PaymentCard.tsx)・[安定識別の解決](../../../../backend/internal/billing/accounting_service_builders.go) | 2026-10-15のPR #670 head 82aa52f3c6d1ce00489a444d903748001670c21cの読取では、UIは固定4種を列挙し有効状態を受け取らない。解決関数は医院／system_key／IDの対応を検査するが当該関数に有効状態の検査はない | [新受領の停止条件](#accounting-inactive-payment-design)は追加目標。固定4種・安定識別の現行説明を停止無視の将来方針にせず、全writer／集計の適合済みとも扱わない |
| [返金UI](../../../../frontend/src/features/accounting/components/RefundSection.tsx)・[返金処理](../../../../backend/internal/billing/refund_service_tx.go) | 理由／方法は任意。請求金額等を基準とする上限経路がある | 理由／方法の必須化、入金累計−既返金累計の上限検査が必要。外部決済の実返金との照合は未確認 |
| [会計計算](../../../../frontend/src/lib/calculations.ts)・[帳票税内訳](../../../../frontend/src/features/accounting/lib/tax-breakdown.ts) | 四捨五入と切捨てを使う箇所が異なる | 第五群の税抜計算・割合額四捨五入と請求書／税率単位の税切捨てへ整合設計が必要。現行式の再検証・税務適合全体は未確認 |
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

- 預り金受領を請求確定／入金配分と別intentにする。医院・元飼主は必須、予約参照はnullableで任意。受領額・方法・実受領日時・actor・操作識別と、充当先請求・任意の予約／ペット・額／日時、実返金額／方法／日時を別履歴として保持する。予約なしでも受領でき、同飼主の別予約／別ペット会計へ確認付きで利用する。指定予約がある場合のみ元飼主／医院との所属を検証し、nullを推測予約IDで埋めない。受領で予約完了／請求確定イベントを生成しない。
- 未充当残は受領累計−有効充当額−未充当分の実返金累計で管理する設計。有効充当額は充当累計−会計取消または減額で解除・残高復元済みの額とし、取消前の累計充当額をそのまま差し引かない。前金充当は請求への支払充足として参照するが、現金／カード等の新受領イベント・二度目の締め入金・二度目の受領領収書を作らない。充当済み分を請求から実返金するときは請求側の履歴へ、未充当分の実返金は前金残へ対応させ、同じ返金を両方から控除しない。第四群で会計取消時の充当解除・元飼主残高復元を合意。実返金は別操作で、既実返金／既復元を二重に戻さない。第八群で減額時の解除・復元と併用時の内訳選択確認を追加。既実返金の出所不明時は第九群の記録確認・内訳／理由登録へ従い、推測で復元しない。
- 5,000円受領→8,000円請求へ5,000円充当では未払3,000円・未充当0円。5,000円受領→確認済み最終料金3,000円なら3,000円充当・未充当2,000円。未実行返金は残を消さず返金済みとしない。同一前金への充当／返金を競合保護し、同一操作の再送で二重使用しない。異医院・別飼主の移管を成立させない。同飼主の別予約／別ペットへ利用する際は、元飼主・任意の元予約を保持して充当先と金額・期待版を再検証する。複数充当先の配分を明示し、合計使用額を最新利用可能額以内にする。
- 通常請求の未払と返金必要は訂正後料金と請求帰属の支払充足から計算する別表示。全額10,000円入金→9,000円訂正では未払0・返金必要1,000、3,000円入金→9,000円訂正では未払6,000・返金必要0。現金・カード等の実入金に基づく支払充足や実返金履歴を、返金予定だけでは減らさない。預り金の確認済み充当解除は第八群に従い、有効充当と請求の支払充足へ反映する。返金後の最新状態を照合して二重返金を防ぎ、支払方法の誤記訂正は金額／実授受日を動かさない。
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

- 値引き前の確定した税抜明細額を入力とし、各明細scope内で割合を合算、割合値引き額を1円単位に四捨五入（非負額の0.5円は切上げ）、定額値引きを加えた要求額と対象額までの実適用額を別に求める。明細後の対象額合計に同じ手順で会計全体scopeの実値引き額を求める。異scope間の割合合算はしない。元行額は第十群の単価×数量を行別に1円四捨五入して求め、割合値引き・配賦・税とは別段階で保持する。
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
| 金額訂正／master再適用 | 行・根拠版、前後値・訂正理由、適用master版・必要な保険再確認 | 新snapshot・差分・履歴／receipt。旧snapshot／帳票保持、金銭移動なし |
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

D10は物販・保険・支払方法・キャンペーンを持ち、診療／薬剤は価格／税・割引除外の会計接続のみ。臨床設定D03、入口D17の権限を編集許可へ転用しない。

| master | 正本と受渡し | 保護 |
|:---|:---|:---|
| 物販 | [商品](../../../spec/screens/settings/master-merchandise.md)の品名・分類・価格／税・有効状態・品目ID／除外 | 税務分類は別確認。現行にない在庫紐付け／減算はD13へ |
| 保険 | [保険](../../../spec/screens/settings/master-insurance.md)の名称・補償率／連絡先・参照ID。会計の資格／負担額は別 | 補償率と飼主負担率を反転せず、masterだけで有効判定しない |
| 支払 | [支払方法](../../../spec/screens/settings/payment-methods.md)の名称／有効状態・安定識別を受領／返金・集計へ | ADR003保持。固定4種拡張／ENUM廃止を発明しない。承認済みの停止方法の新受領利用不可と過去記録／集計保持は[追加設計](#accounting-inactive-payment-design)へ |
| 割引 | [campaign](../../../spec/screens/settings/master-campaigns.md)の期間・対象・割合／額・有効状態・元ID | 最大割引の現行経路を併用適合済みにせずscope／除外／snapshot対応へ |
| 診療／薬剤 | 品目ID・価格／税／除外と追加時snapshot | 投薬量・臨床パラメータを会計丸め／再適用で変更しない |

既存CRUD／reorder・入力表は現行参照として保持。目標create／editは対象変更許可・clinic／版、delete等は別操作許可を検査。名称変更／無効化／削除でも過去取引／帳票参照を保持し、保護できない削除は拒否、別masterへ勝手に付替えない。支払方法の有効設定と選択肢の効果は、現行の固定4種／状態非連動と[追加合意の目標](#accounting-inactive-payment-design)を区別する。ADR003の安定識別・医院分離は保持し、停止済み方法を新受領で許可する根拠にしない。取得／失敗／競合／未保存保護は画面共通契約へ。現grant不足は[権限mapping](#accounting-master-permission-map)へ。

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

<a id="accounting-document-completion"></a>
## D10文書完成条件と判定外

D03と同じ限定文書基準を適用。対象は9FR／151AC、3画面＋選択／新規入口、4masterと診療／薬剤会計接続、語彙・状態／ID・操作／保存／認可・旧値／停止・writer／consumer／受け手。2026-10-15の独立レビューとリンク／fragment・表／trace照合で重大指摘なしを確認し、基準時点の9FR／142ACを文書整理完了と判定した。追加の死亡後精算2ACも反映・独立文書レビュー済み。支払方法停止の追加7ACも今回の固定39文書レビュー・限定構造検査を通過した。既存144ACの判定とは別に追加差分の確認根拠を保持する。既存AC／アンカー・144AC本文を保持し、新旧全ACを一意に追跡する。

文書で選定した意味型・原子保存／receipt・現在認可／不明時停止・画面状態／master参照保護・引継ぎを後決めへ戻さない。物理schema／wire名・採番書式／lock実装・実grant／実旧値・移行・全writer／consumer実装適合・実機／性能／復旧試験・税務／決済規約適合は後続実装／利用開始条件。これらと他owner内部・全体横断最終照合は本領域の文書完成判定外。保険会社請求／査定・自動給付、督促／貸倒、別医院／飼主移管・外部機器連携は対象外。

照合する境界は、確定×未入金／部分×空明細／0円、訂正／取消×現金／カード／預り×既返金／復元、再請求×旧入金／預り残、行額丸め×割引／除外／複数税率、再適用×手修正／確定競合、不明旧値×返金、保存×失敗／同一再送／内容違い／遅延成功、画面切替×失敗／古い応答、各master×無認可／失効／参照保護。期待結果は正本ACと本契約、実行受入は後続工程へ。

## 受入条件との追跡と設計残件

| 業務FR | 正本AC範囲 | 技術照合対象 |
|:---|:---|:---|
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

overdiscount-unresolved／overdiscount-post-warning-pending は参照互換の既存識別子であり、期待結果は合意済みの警告・請求額0円下限へ更新する。

AC短名は[正本151AC](../../../spec/domains/accounting/lifecycle.md#accounting-acceptance)のAC-accounting-接頭辞に続く名称。ここに第二の期待結果を作らない。

実装前に、確定と入金の競合、同時返金による上限超過、同時二重請求、失効認可／異医院参照、通信切断後の遅延commit、複数記録と締め参加、マスタ変更とsnapshotを対象に正常／拒否／再送／結果不明を照合する必要がある。既存S08／S21／S33は現行手順で、現在151ACの実行済み証拠ではない。runtime・API／DB改訂・移行・利用開始・性能は未実施。

業務判断と技術設計／検証・対象外の区別は[残件の分類](../../../spec/domains/accounting/lifecycle.md#accounting-open-decisions)へ。D05／06専用記録、D11見積、D12締め、D13在庫、D16集計、C04権限／C05監査の受け手実装適合・全体横断再照合は別工程。保険会社への請求送信・入金照合・査定差額管理・会社／商品別の自動給付算定、外部支払／印刷機器を追加実装する設計は含めない。

