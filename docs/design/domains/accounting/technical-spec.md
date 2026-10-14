# 会計初回・第二群の技術仕様

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | design.accounting.partial／技術設計 |
| 決定状態 | 業務8FRは合意済み。本書の実現方式は提案設計、未決業務を補完しない |
| 実装状態 | 未確認。静的に確認した現行制約は下記。API wire／DB正本は変更なし |
| 実機検証状態 | 未実施。移行・並行性・性能も未検証 |
| 文書確認日 | 2026-10-14 |

[業務・AC正本](../../../spec/domains/accounting/lifecycle.md)と[語彙](../../../spec/ontology/accounting.md)を参照。対象は初回5点＋第二群の部分反映で、全writer・consumerの実装適合やD10全体完了を表さない。

<a id="accounting-current-gap"></a>
## 現行実装と確定業務の差分

| 調査対象 | 現行の静的確認 | 目標業務との相違／残件 |
|:---|:---|:---|
| [確定入力・元カルテ照合](../../../../backend/internal/billing/accounting_complete.go) | ヘッダは単一ペット／カルテ。診療項目が複数カルテにまたがる場合は拒否 | 同飼主の複数ペット・当日／過去の選択を表現する明細参照へ設計が必要。単一患者への付替えは禁止 |
| [確定transaction](../../../../backend/internal/billing/accounting_complete_tx.go)・[支払UI](../../../../frontend/src/features/accounting/components/PaymentCard.tsx) | 支払内訳を請求に一致させる経路。初回部分入金は拒否 | 未納／部分入金確定・後日入金の業務は現行制約と異なる。[既存S08](../../../ops/testing/scenarios/S08-accounting-corrections.md)を新仕様の合格根拠にしない |
| [返金UI](../../../../frontend/src/features/accounting/components/RefundSection.tsx)・[返金処理](../../../../backend/internal/billing/refund_service_tx.go) | 理由／方法は任意。請求金額等を基準とする上限経路がある | 理由／方法の必須化、入金累計−既返金累計の上限検査が必要。外部決済の実返金との照合は未確認 |
| [会計計算](../../../../frontend/src/lib/calculations.ts)・[帳票税内訳](../../../../frontend/src/features/accounting/lib/tax-breakdown.ts) | 四捨五入と切捨てを使う箇所が異なる | 算定単位・順序の業務判断後に整合設計。どちらが法的に正しいかは未確認 |
| [割引生成](../../../../backend/internal/billing/billing_item_service_create.go) | 飼主／キャンペーンの最大割引を選ぶ経路 | 併用・割合→金額の合意と相違。複数割合・適用粒度／保険前後・税丸め等は未決 |
| [未納期間ビュー](../../../spec/domains/accounting/pages/unpaid-list.md) | 期間内に非取消売上のある飼主に対象を限定する既存記載 | 今月売上なしでも過去未納飼主を表示する合意と相違。filter／基準日の詳細は設計残件 |

API wireの正本は[現行API](../../../../backend/docs/api.yaml)、DBはbackend/migrations。本書の目標データは既存DTOや保存schemaが対応済みとの宣言ではない。全grant、旧値、各書込経路、帳票履歴、外部機器は未確認。

<a id="accounting-operation-contract"></a>
## 提案する保存・照合契約

- 請求ヘッダの飼主／医院と、行ごとの元ペット・診療／専用記録・元明細IDを区別する。選択した未請求明細と既請求への回収参照を別に保持し、同じ元明細の二重請求を防ぐ。複数対象の正確な予約／記録リンクと確定イベントを保持する。
- 請求確定・追加入金・訂正・取消・返金を別intentとして扱う。対象ID群・期待版・依存参照・操作識別・actor・サーバー時刻、操作に必要な理由／返金方法を保持する。現在の医院／飼主所属・認可・請求状態・入金／返金累計を保存直前に再検証する。
- 元請求の不変な履歴と訂正／入金／返金を追跡し、請求額・入金履歴・残額の表示をそれぞれ取得する。訂正後料金に対する未払・返金必要を別projectionで表示し、実返金後に更新する。通常請求の例は業務ACへ従う。古い請求から初期充当し、スタッフ変更を元請求別の配分履歴に保存する。原請求日を初期順の根拠とする設計案だが同日順は未決。通常回収の超過分は釣銭へ。取消後の再請求・貸倒・別予約振替／予約なし預り金等は未決で、自動処理を追加しない。
- 保存する会計時点の根拠は価格・数量・元値／手修正値・適用割引・保険等と計算結果。現在マスタの再計算を過去会計表示の正本にしない。未確定の算定順・税／丸め・保険条件を承認済みversionへ登録しない。請求書と実受領の領収書を別projectionとし、再発行表示・訂正前帳票の履歴を保持する設計。発行時の値／対象版を保持し現在マスタから旧帳票を再生成しない。保存形式・番号等は未決。
- 状態変更と必須の履歴・監査、依存ownerの確定参加を一体で成立させる設計へ。[予約の全writer gate／取得順](../reservations/technical-spec.md#reservation-global-lock-contract)と[受け手契約](../reservations/technical-spec.md#reservation-receiving-contracts)を維持する。複数対象での取得順・transaction範囲・旧writer参加は設計／検証残件で、現行separate-txをsame-tx済みとしない。
- 同じ操作の再送は同じ内容・対象の結果に照合し、二重入金／二重返金を作らない。payload相違・版競合・認可失効は成立させない。通信切断で結果不明ならreceipt照会／同一操作の再送で確認し、新キーの作り直しや残額0・失敗確定への推測変換を避ける。追加入金等の現行endpoint対応・永続operation receiptは未確認で、新wireを本書だけから決定しない。
- 返金累計上限は最新の実入金・既返金を競合保護下で再検証。指定方法と実際の返金記録を保持する。方法別上限や異なる方法を許す条件は未決で、今回の総額上限と混同しない。
- 締め後訂正は元close参照と差分、実際の入金／返金の参照を区別し、元締めを更新しない。原請求日と実受領／実返金日、前金充当日を別に保持する。今日の実入出金は今日の移動、充当で新たな実受領を生成しない。締め区分・会計差分と入出金の突合はD12と照合。金額だけの変更から金銭移動を生成しない。

カルテconfirmation ID＋record IDとaccounting IDを混同しない。[カルテの受渡し](../medical-records/technical-spec.md#medical-record-reconciled-contract)から対象／版・理由・影響確認を受け、会計ownerで現在検査と結果receiptを返す。臨床の保存成功は金銭処理成功を表さない。[カルテ保存§2](../../../spec/domains/medical-records/lifecycle.md#hybrid-sync)の送信先・保存単位を変えない。取消／返金・0円・未納でも不変の確定歴を保持し、通常Undoを再許可しない。請求取消を返金／在庫返却の自動intentにしない。

<a id="accounting-second-design"></a>
## 第二群の追加契約（提案設計）

- 前金受領を請求確定／入金配分と別intentにする。医院・飼主・元予約、受領額・支払方法・実受領日時・actor・操作識別を記録し、充当先請求・充当額／日時と実返金額／方法／日時を別の履歴として保持する。元予約・飼主の追跡を失わず、最新の予約所属・認可・対象請求と現在版を保存時に再検証する。前金受領で予約完了や請求確定イベントを生成しない。
- 未充当残は受領済み前金から充当済み額と未充当分の実返金額を差し引いて管理する設計。前金充当は請求への支払充足として参照するが、現金／カード等の新受領イベント・二度目の締め入金・二度目の受領領収書を作らない。充当済み分を請求から実返金するときは請求側の履歴へ、未充当分の実返金は前金残へ対応させ、同じ返金を両方から控除しない。充当後の請求取消に伴う充当解除／再利用方法は未決。
- 5,000円受領→8,000円請求へ5,000円充当では未払3,000円・未充当0円。5,000円受領→確認済み最終料金3,000円なら3,000円充当・未充当2,000円。未実行返金は残を消さず返金済みとしない。同一前金への充当／返金を競合保護し、同一操作の再送で二重使用しない。異医院・別飼主・未承認の別予約振替を成立させない。複数請求／ペットへの前金配分は業務判断後に設計する。
- 通常請求の未払と返金必要は訂正後料金と請求帰属の支払充足から計算する別表示。全額10,000円入金→9,000円訂正では未払0・返金必要1,000、3,000円入金→9,000円訂正では未払6,000・返金必要0。実返金前に支払充足や返金履歴を減らさない。返金後の最新状態を照合して二重返金を防ぎ、支払方法の誤記訂正は金額／実授受日を動かさない。
- 古い請求への初期配分・スタッフ変更・釣銭を保存前previewと保存後receiptで照合する。選択請求の残額・版と配分合計を一体で検証し、別操作の入金と競合したら旧配分で過大充当しない。前金受領との操作種別を分け、釣銭を予約前金や負の返金として自動保存しない。
- 割合→金額の段階とスタッフが確認修正した保険対象・割合・上限を計算根拠へ保存する。複数割合・粒度・保険前後・税丸めは未決。算定規則が未確定なら既存計算式を新業務の承認済み算式にしない。
- 今月売上なしの過去未納飼主も含める読取projectionへ。全ページの集計と表示対象を整合させ、取得失敗を未納0に変換しない。帳票発行・再発行は対象請求／受領／訂正版と発行履歴を照合し、訂正前の記録を保持する。再発行操作と新入金を同じintentにしない。

前金の金額／キャンセル料／受領時期はD05、実入出金と充当の集計はD12／D16、操作権限はC04とD10で照合。新endpoint・DDL・payment methodの実装対応は未確認で、本書から実装／移行を開始しない。ホテル等の元予約参照を受けるが、D05の内部完了条件・予約state・臨床確認を再定義しない。

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

AC短名は[正本40AC](../../../spec/domains/accounting/lifecycle.md#accounting-acceptance)のAC-accounting-接頭辞に続く名称。ここに第二の期待結果を作らない。

実装前に、確定と入金の競合、同時返金による上限超過、同時二重請求、失効認可／異医院参照、通信切断後の遅延commit、複数記録と締め参加、マスタ変更とsnapshotを対象に正常／拒否／再送／結果不明を照合する必要がある。既存S08／S21／S33は現行手順で、新40ACの実行済み証拠ではない。runtime・API／DB改訂・移行・利用開始・性能は未実施。

未決の業務は[5群の確認項目](../../../spec/domains/accounting/lifecycle.md#accounting-open-decisions)へ。D05／06専用記録、D11見積、D12締め、D13在庫、D16集計、C04権限／C05監査の受け手内部と全体横断再照合は未完了。オンライン保険請求・外部支払／印刷機器を追加実装する設計は含めない。
