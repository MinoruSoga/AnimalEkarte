# 会計初回〜第四群の技術仕様

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | design.accounting.partial／技術設計 |
| 決定状態 | 業務8FRは合意済み。本書の実現方式は提案設計、未決業務を補完しない |
| 実装状態 | 未確認。静的に確認した現行制約は下記。API wire／DB正本は変更なし |
| 実機検証状態 | 未実施。移行・並行性・性能も未検証 |
| 文書確認日 | 2026-10-15 |

[業務・AC正本](../../../spec/domains/accounting/lifecycle.md)と[語彙](../../../spec/ontology/accounting.md)を参照。対象は初回5点〜第四群の部分反映で、全writer・consumerの実装適合やD10全体完了を表さない。

<a id="accounting-current-gap"></a>
## 現行実装と確定業務の差分

| 調査対象 | 現行の静的確認 | 目標業務との相違／残件 |
|:---|:---|:---|
| [確定入力・元カルテ照合](../../../../backend/internal/billing/accounting_complete.go) | ヘッダは単一ペット／カルテ。診療項目が複数カルテにまたがる場合は拒否 | 同飼主の複数ペット・当日／過去の選択を表現する明細参照へ設計が必要。単一患者への付替えは禁止 |
| [確定transaction](../../../../backend/internal/billing/accounting_complete_tx.go)・[支払UI](../../../../frontend/src/features/accounting/components/PaymentCard.tsx) | 支払内訳を請求に一致させる経路。初回部分入金は拒否 | 未納／部分入金確定・後日入金の業務は現行制約と異なる。[既存S08](../../../ops/testing/scenarios/S08-accounting-corrections.md)を新仕様の合格根拠にしない |
| [返金UI](../../../../frontend/src/features/accounting/components/RefundSection.tsx)・[返金処理](../../../../backend/internal/billing/refund_service_tx.go) | 理由／方法は任意。請求金額等を基準とする上限経路がある | 理由／方法の必須化、入金累計−既返金累計の上限検査が必要。外部決済の実返金との照合は未確認 |
| [会計計算](../../../../frontend/src/lib/calculations.ts)・[帳票税内訳](../../../../frontend/src/features/accounting/lib/tax-breakdown.ts) | 四捨五入と切捨てを使う箇所が異なる | 算定単位・順序の業務判断後に整合設計。どちらが法的に正しいかは未確認 |
| [割引生成](../../../../backend/internal/billing/billing_item_service_create.go) | 飼主／キャンペーンの最大割引を選ぶ経路 | 併用・割合→金額の合意と相違。割合合算→金額→割引後保険の目標へ。両粒度の割引利用／超過警告と承認後の請求額0円下限へ追従。併用順序／配賦・税丸めは未決 |
| [未納期間ビュー](../../../spec/domains/accounting/pages/unpaid-list.md) | 期間内に非取消売上のある飼主に対象を限定する既存記載 | 今月売上なしでも過去未納飼主を表示する合意と相違。filter／基準日の詳細は設計残件 |

API wireの正本は[現行API](../../../../backend/docs/api.yaml)、DBはbackend/migrations。本書の目標データは既存DTOや保存schemaが対応済みとの宣言ではない。全grant、旧値、各書込経路、帳票履歴、外部機器は未確認。

<a id="accounting-operation-contract"></a>
## 提案する保存・照合契約

- 請求ヘッダの飼主／医院と、行ごとの元ペット・診療／専用記録・元明細IDを区別する。選択した未請求明細と既請求への回収参照を別に保持し、同じ元明細の二重請求を防ぐ。複数対象の正確な予約／記録リンクと確定イベントを保持する。
- 請求確定・追加入金・訂正・取消・返金を別intentとして扱う。対象ID群・期待版・依存参照・操作識別・actor・サーバー時刻、操作に必要な理由／返金方法を保持する。現在の医院／飼主所属・認可・請求状態・入金／返金累計を保存直前に再検証する。
- 元請求の不変な履歴と訂正／入金／返金を追跡し、請求額・入金履歴・残額の表示をそれぞれ取得する。訂正後料金に対する未払・返金必要を別projectionで表示し、実返金後に更新する。通常請求の例は業務ACへ従う。古い請求から初期充当し、スタッフ変更を元請求別の配分履歴に保存する。原請求日を初期順の根拠とし同日は作成順へ対応する設計。同一作成時刻は一意な作成順識別で安定させる技術案。保存後も理由と元／新配分履歴を保持して変更する。現金超過は釣銭、非現金超過は実入金と請求充当を分けて返金必要額へ。会計取消時の預り金残高復元と別返金は第四群へ従う。取消後再請求・貸倒・減額時の充当調整等の未決を自動処理にしない。
- 保存する会計時点の根拠は価格・数量・元値／手修正値・適用割引・保険等と計算結果。現在マスタの再計算を過去会計表示の正本にしない。未確定の算定順・税／丸め・保険条件を承認済みversionへ登録しない。請求書と実受領の領収書を別projectionとし、再発行表示・訂正前帳票の履歴を保持する設計。発行時の値／対象版を保持し現在マスタから旧帳票を再生成しない。保存形式・番号書式等は残件だが、各取引の元番号を再発行で維持することは合意済み。
- 状態変更と必須の履歴・監査、依存ownerの確定参加を一体で成立させる設計へ。[予約の全writer gate／取得順](../reservations/technical-spec.md#reservation-global-lock-contract)と[受け手契約](../reservations/technical-spec.md#reservation-receiving-contracts)を維持する。複数対象での取得順・transaction範囲・旧writer参加は設計／検証残件で、現行separate-txをsame-tx済みとしない。
- 同じ操作の再送は同じ内容・対象の結果に照合し、二重入金／二重返金を作らない。payload相違・版競合・認可失効は成立させない。通信切断で結果不明ならreceipt照会／同一操作の再送で確認し、新キーの作り直しや残額0・失敗確定への推測変換を避ける。追加入金等の現行endpoint対応・永続operation receiptは未確認で、新wireを本書だけから決定しない。
- 返金累計上限は最新の実入金・既返金を競合保護下で再検証。指定方法と実際の返金記録を保持する。元カードから現金等の別方法を記録・操作でき、金額による追加承認は設けない。操作権限・理由・方法と総額上限を検査する。旧支払方法ごとの上限経路が別方法返金を一律拒否しないよう実装差分を照合する必要がある。決済サービスの実行・規約適合は未確認で、自動連携を追加しない。
- 締め後訂正は元close参照と差分、実際の入金／返金の参照を区別し、元締めを更新しない。原請求日と実受領／実返金日、前金充当日を別に保持する。今日の実入出金は今日の移動、充当で新たな実受領を生成しない。締め区分・会計差分と入出金の突合はD12と照合。金額だけの変更から金銭移動を生成しない。

カルテconfirmation ID＋record IDとaccounting IDを混同しない。[カルテの受渡し](../medical-records/technical-spec.md#medical-record-reconciled-contract)から対象／版・理由・影響確認を受け、会計ownerで現在検査と結果receiptを返す。臨床の保存成功は金銭処理成功を表さない。[カルテ保存§2](../../../spec/domains/medical-records/lifecycle.md#hybrid-sync)の送信先・保存単位を変えない。取消／返金・0円・未納でも不変の確定歴を保持し、通常Undoを再許可しない。請求取消を返金／在庫返却の自動intentにしない。

<a id="accounting-second-design"></a>
## 第二群から第三群へ更新した契約（提案設計）

- 預り金受領を請求確定／入金配分と別intentにする。医院・元飼主は必須、予約参照はnullableで任意。受領額・方法・実受領日時・actor・操作識別と、充当先請求・任意の予約／ペット・額／日時、実返金額／方法／日時を別履歴として保持する。予約なしでも受領でき、同飼主の別予約／別ペット会計へ確認付きで利用する。指定予約がある場合のみ元飼主／医院との所属を検証し、nullを推測予約IDで埋めない。受領で予約完了／請求確定イベントを生成しない。
- 未充当残は受領累計−有効充当額−未充当分の実返金累計で管理する設計。有効充当額は充当累計−会計取消で解除・残高復元済みの額とし、取消前の累計充当額をそのまま差し引かない。前金充当は請求への支払充足として参照するが、現金／カード等の新受領イベント・二度目の締め入金・二度目の受領領収書を作らない。充当済み分を請求から実返金するときは請求側の履歴へ、未充当分の実返金は前金残へ対応させ、同じ返金を両方から控除しない。第四群で会計取消時の充当解除・元飼主残高復元を合意。実返金は別操作で、既実返金／既復元を二重に戻さない。減額だけの場合や複雑な訂正手順は残件。
- 5,000円受領→8,000円請求へ5,000円充当では未払3,000円・未充当0円。5,000円受領→確認済み最終料金3,000円なら3,000円充当・未充当2,000円。未実行返金は残を消さず返金済みとしない。同一前金への充当／返金を競合保護し、同一操作の再送で二重使用しない。異医院・別飼主の移管を成立させない。同飼主の別予約／別ペットへ利用する際は、元飼主・任意の元予約を保持して充当先と金額・期待版を再検証する。複数充当先の配分を明示し、合計使用額を最新利用可能額以内にする。
- 通常請求の未払と返金必要は訂正後料金と請求帰属の支払充足から計算する別表示。全額10,000円入金→9,000円訂正では未払0・返金必要1,000、3,000円入金→9,000円訂正では未払6,000・返金必要0。実返金前に支払充足や返金履歴を減らさない。返金後の最新状態を照合して二重返金を防ぎ、支払方法の誤記訂正は金額／実授受日を動かさない。
- 古い請求への初期配分・スタッフ変更・釣銭を保存前previewと保存後receiptで照合する。選択請求の残額・版と配分合計を一体で検証し、別操作の入金と競合したら旧配分で過大充当しない。預り金受領との操作種別を分け、釣銭や非現金超過を預り金へ自動転用しない。再配分は元の実入金と実日付を保持し、請求訂正・既返金・他の配分を現在版で照合する。元配分の取消と新配分・関連請求の残額・理由／監査を原子的に保存し、二重受領や片側変更を作らない。
- 割合合算→金額割引→割引後の保険対象額という段階と、スタッフが確認修正した保険対象・割合・上限を計算根拠へ保存する。明細／会計全体の両方を利用でき、100%超や商品額超は警告し、承認して進めた請求額は0円を下限とする。両粒度の併用順序・対象外／税区分への配賦、税丸めは未決。算定規則が未確定なら既存計算式を新業務の承認済み算式にしない。
- 今月売上なしの過去未納飼主も含める読取projectionへ。全ページの集計と表示対象を整合させ、取得失敗を未納0に変換しない。帳票発行・再発行は対象請求／受領／訂正版と発行履歴を照合し、訂正前の記録を保持する。再発行操作と新入金を同じintentにしない。

前金の金額／キャンセル料／受領時期はD05、実入出金と充当の集計はD12／D16、操作権限はC04とD10で照合。新endpoint・DDL・payment methodの実装対応は未確認で、本書から実装／移行を開始しない。ホテル等の元予約参照を受けるが、D05の内部完了条件・予約state・臨床確認を再定義しない。

<a id="accounting-third-design"></a>
## 第三群の預り金・再配分・帳票の補完（提案設計）

預り金の現在残は受領累計−有効充当額−未充当分の実返金累計を履歴から追跡する。有効充当額は充当累計から会計取消による解除・残高復元済み額を差し引き、復元を残高へ反映する。請求側からの実返金を同時に預り金残から控除しない。利用可能額は保存直前の最新残と既に処理中の使用を照合して算出し、未実行返金の表示だけで残を消さず、同じ額を返金と別会計の両方へ成立させない。実現方式はowner／資金参照と関連請求を既存gate／整列lock下で検査し、状態・履歴・監査・operation receiptを原子的に保存する案。確保状態や保存schemaは未実装・未検証で、予約IDを飼主残高の排他単位の代用にしない。

充当確認は飼主・預り金元取引・対象請求／ペットと任意の予約・金額・資金版／請求版を含む。確認後に残高・所属・請求訂正が変われば古い確認でcommitしない。再送は同一操作へ照会し、別keyで二重充当／二重返金を作らない。同じ飼主の別予約／別ペット利用を旧予約一致必須のvalidationで拒否しない。別飼主の所有者付替えや医院間残高移送は未承認。予約や飼主の削除／変更時も履歴参照を失わず、具体的廃棄・統合手順は別ownerの残件。

非現金超過は実受領5,000円・請求充当3,000円・未実行返金必要2,000円を別projectionへ記録する設計。返金処理後に実返金履歴と必要額を更新し、釣銭・支払額の誤記訂正・飼主預り金の受領を代用にしない。異なる返金方法でも元取引・実返金方法・actor・理由・実日付・総額上限を追跡する。金額による承認段階を追加せず、サーバーの返金権限検査は維持する。

各受領・充当・入金・再配分・実返金は一意な操作／取引参照で追跡し、帳票には用途に応じた取引番号と元番号を表示する。再発行は元取引番号のまま、再発行回数／時刻・版／actor・出力内容の履歴を追加する案で、二度目の金銭取引を作らない。番号の書式・採番方式・保存形式は技術残件、法令上の採番要件は未確認。画面へ返すoperation receiptと飼主へ渡す受領証／領収書は別概念。

明細と会計全体の両割引を許可し超過は警告へ。承認して進めた請求額は0円を下限とし、負額で保存しない。この請求額の下限を、明細ごとの中間額や割引率のclamp、両粒度併用順序・税／丸めを含む既存式全体の承認へ拡張しない。保険対象外と税区分は行の根拠を保持し、全額を一つの保険基準／税区分に潰さない。超過そのものを入力エラーへ戻さず、警告を承認して進められることと請求額の0円下限をpreview／確定結果へ反映する。過去の承認済み会計snapshotを現在の算式で置換しない。

<a id="accounting-fourth-design"></a>
## 第四群の割引・取消残高復元（提案設計）

明細割引と会計全体割引を別scopeとして保存し、各値・対象・根拠を保持する。両方利用する許可は確定、適用順序・割合合算範囲・税区分／保険対象への配賦は未決。超過をwarningとして表現し、100%超／金額超だけを入力エラーcodeにしない。表示・計算preview・確定結果を区別し、警告を承認して進めた請求額を0円下限として保存する案。元の割引値と0円となった根拠を保持し、負額の請求や架空の入金／返金を作らない。明細の中間計算や税・端数／配賦の方式は本合意だけで補完しない。税務適合や新たな必須承認操作を本案から導かない。

会計取消intentに元充当参照・元飼主・対象請求／預り金と期待版・理由・actor・operation参照を渡す。既存gate／整列取得の下で現在の有効充当と既実返金／既復元を検査し、取消・充当解除・元飼主残高への復元・必要履歴／監査・結果receiptを原子的に保存する案。元受領と充当履歴は消さず、解除／復元を追記して有効充当額を減らし、未充当残と利用可能額を更新する。新受領や実返金を生成しない。

未返金5,000円受領・3,000円充当・残2,000円なら取消で有効充当額0円・残5,000円へ。未充当分を1,000円実返金済みなら、同じ3,000円の解除・復元後もその実返金を控除し、残4,000円とする。繰り返し取消・同一再送で二度復元せず、同時充当／返金と現在残を照合する。既実返金と復元の二重計上を防ぎ、複雑な返金済み訂正の内訳／減額だけの処理は残件へ。取消後も不変の会計確定歴・原請求／予約リンクを保持し通常Undoを再開しない。復元された残からの実返金は別の理由・方法・権限付き操作、復元成功を実返金成功にしない。

## 受入条件との追跡と設計残件

| 業務FR | 正本AC範囲 | 技術照合対象 |
|:---|:---|:---|
| FR-accounting-owner-selected-charges | owner-multi-pet／selected-unbilled／old-bill-collection／no-invented-allocation | 複数元参照・選択・二重請求防止・元請求帰属 |
| FR-accounting-receivable-partial-zero | unpaid-finalize／partial-finalize／later-payment／zero-no-fake-payment | 請求と入金の別状態・履歴・残額・0円根拠 |
| FR-accounting-reasoned-financial-correction | correction-authorization／correction-history／refund-received-cap／refund-method-required／cancel-no-money-stock | 認可・理由・監査・実入金上限・方法・owner分離 |
| FR-accounting-post-close-separate-movement | post-close-preserve／correction-no-cash-movement | 元締めと差分・実移動の分離 |
| FR-accounting-calculation-snapshot | master-change-stable／manual-unit-price／discounts-combinable | snapshot・修正元値・割引併用と未決算定 |
| 既決横断境界 | finalization-undo-preserved／clinical-no-auto-finance | 確定歴・正確なリンク・owner receipt |
| FR-accounting-oldest-payment-allocation | oldest-default／allocation-staff-change／collection-change-return | 古い順・変更・合計・元請求・釣銭と前金の区別 |
| FR-accounting-reservation-advance | advance-owner-reservation／advance-unallocated-balance／advance-apply-no-double-receipt／advance-surplus-pending／advance-actual-refund／advance-duplicate-use／advance-context-mismatch | 元予約／飼主・未充当残・二重使用／二重受領防止・実返金 |
| FR-accounting-reasoned-financial-correction（第二群） | paid-reduction-refund-pending／paid-reduction-actual-refund／partial-paid-reduction／payment-method-correction-no-money | 訂正後表示・未実行／実返金・金額不変の誤記訂正 |
| FR-accounting-calculation-snapshot（第二群） | percent-before-fixed／insurance-staff-review | 割合→金額・保険確認修正snapshotと未決順序 |
| FR-accounting-document-and-aged-debt | invoice-receipt-separated／document-reissue-history／aged-debt-without-current-sales | 請求／実受領帳票・再発行／旧履歴・過去未納の表示 |
| FR-accounting-post-close-separate-movement（第二群） | original-bill-and-movement-date | 原請求日と実日付・充当時の非受領 |
| FR-accounting-reservation-advance（第三群） | deposit-without-reservation／deposit-other-reservation／deposit-other-pet／deposit-other-owner-denied／deposit-allocation-confirmation／deposit-available-balance | 飼主帰属・予約任意・確認・利用可能額・同時使用防止 |
| FR-accounting-calculation-snapshot（第三群） | percentage-summed／discount-then-insurance／overdiscount-unresolved | 合算→金額→割引後保険、超過警告と請求額0円下限、併用順序／配賦・税の未決 |
| FR-accounting-oldest-payment-allocation（第三群） | same-day-created-order／saved-reallocation-reason／reallocation-no-new-money／reallocation-correction-conflict／noncash-excess-refund-pending | 作成順・理由と履歴・原子的再配分・非現金超過 |
| FR-accounting-reasoned-financial-correction（第三群） | refund-different-method／refund-no-amount-extra-approval | 別方法・金額追加承認なし・総額上限／権限 |
| FR-accounting-document-and-aged-debt（第三群） | deposit-receipt／final-document-deposit-current／refund-original-transaction／reissue-original-number | 用途別帳票・充当と実受領・取引番号／再発行履歴 |
| FR-accounting-calculation-snapshot（第四群） | line-and-bill-discount／overdiscount-warning／overdiscount-post-warning-pending | 両粒度利用・超過警告・承認後の請求額0円下限 |
| FR-accounting-reservation-advance（第四群） | cancel-restore-deposit／cancel-deposit-no-cash-refund／cancel-deposit-once-atomic | 取消・充当解除・残高復元と別返金・原子的一度処理 |

overdiscount-unresolved／overdiscount-post-warning-pending は参照互換の既存識別子であり、期待結果は合意済みの警告・請求額0円下限へ更新する。

AC短名は[正本66AC](../../../spec/domains/accounting/lifecycle.md#accounting-acceptance)のAC-accounting-接頭辞に続く名称。ここに第二の期待結果を作らない。

実装前に、確定と入金の競合、同時返金による上限超過、同時二重請求、失効認可／異医院参照、通信切断後の遅延commit、複数記録と締め参加、マスタ変更とsnapshotを対象に正常／拒否／再送／結果不明を照合する必要がある。既存S08／S21／S33は現行手順で、新66ACの実行済み証拠ではない。runtime・API／DB改訂・移行・利用開始・性能は未実施。

未決の業務は[5群の確認項目](../../../spec/domains/accounting/lifecycle.md#accounting-open-decisions)へ。D05／06専用記録、D11見積、D12締め、D13在庫、D16集計、C04権限／C05監査の受け手内部と全体横断再照合は未完了。オンライン保険請求・外部支払／印刷機器を追加実装する設計は含めない。
