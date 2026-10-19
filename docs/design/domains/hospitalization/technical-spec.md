# 入院・ホテルの技術契約

| 項目 | 内容 |
|:---|:---|
| 文書ID・種別 | D05-TECH / 技術設計・現行差分 |
| 決定状態 | 既合意の受け手設計と既存契約の継承。専用訂正・ホテル個別料金と入院中治療プランの履歴付き追加・変更・中止は2026-10-15追加回答で承認（実装適合は別）。死亡後の生前分専用処理も最新承認を反映 |
| 実装状態 | 限定ソース静的照合のみ。新しい開始・頭数・担当の実装適合未確認 |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-15 |

業務FR／ACは[入院・ホテルlifecycle](../../../spec/domains/hospitalization/lifecycle.md)、語彙は[オントロジー](../../../spec/ontology/hospitalization.md)。API型は[API正本](../../../../backend/docs/api.yaml)、DBはmigrations。本書は第二のwire schemaを定義しない。

## write owner・識別・保存単位

Hospitalization・日次record・ケア／治療plan・Cage・HospitalizationPlanはmedicalrecordがwriter。appointmentsはreservation、Billing／BillingItemはbillingを所有者とする。[write-owner](../../../architecture/model-write-owner-catalog.md)・[cross-domain catalog](../../../architecture/cross-domain-orchestration-catalog.md)へ従い、他ownerの表を独立writerで更新しない。

clinic＋hospitalization IDを滞在の内部参照とし、予約ref・通常medical_record ref・入院専用記録ref・billing refを型付きに区別する。日付・ペット名だけから同一入院を推定しない。同じ患者の別滞在と、同じ滞在に二つの入口がある場合を分離する。

| 操作 | 同時成功の境界 | 現在の確認状態 |
|:---|:---|:---|
| 新規入院と登録時治療プラン | 親＋treatment_plans | 画面／service既存記載。開始記録の新合意とは別 |
| 入院中の治療プラン追加・変更・中止 | 元内容保持＋現在内容／中止の識別＋変更履歴・入力actor／時刻 | 2026-10-15の第1項合意目標。現行読取専用／常時Conflictからの差分で実装未確認 |
| 入院開始 | 同じ滞在の開始状態＋必要な入院記録 | 業務合意済み。現行単純status PATCHからの差分あり |
| 日次record／vitals／care-logs／staff-notes | 各追加操作ごと | 3セクションを一括saveとみなさない |
| ケア項目 | 1 item create/update/delete | plan snapshotではない。手入力itemだけ理由を要求 |
| ケージ単独移動／解除 | 入院行lock＋表示時期待割当検査 | 現行serviceの409／400を静的確認 |
| ケージ交換 | 二件の入院行を順序lock＋期待値照合 | 片方成功を許さない。連結は対象外 |
| 会計なし退院PATCH | 状態／退院日＋必要退院監査 | 現行PATCH経路は監査fail-closed |
| 会計付き退院 | 状態＋必要会計／明細＋退院会計監査 | PATH-HOSP-DISCHARGE-BILLはsame-tx |
| 親削除 | 子依存保護＋soft-delete／監査 | 日次・治療・ケアを保護。全並行writer適合は未確認 |

## 予定・実績・頭数

- D01の同一宿泊参照を継承し、予約枠と実滞在を別に持つ。同一滞在の予約・実績の両readを1件へ対応させる設計とし、片方が欠落した場合は推測で二件／零件へ決めない。
- 予定区間の接触は重複にしないが、未退室超過は現在滞在として保持する。予定endだけをWHERE句に使って実滞在を除外する方式は合意を満たさない。
- 延泊案、承認前の調整待ち、実績、確定予定を区別する。超過を避けるため実際の退室を虚偽で保存しない。責任者調整後に全期間・他予約・最新実滞在を同じ判定contextで再照合する。
- 頭数上限とケージ占有は異なるread model。ケージ数・医師勤務数から宿泊上限を補完しない。LINE受入と宿泊側の更新が競合しても同じ上限判定を通す設計が必要。
- 現行HospStatusはreserved/admitted/discharged。FEの一時帰宅定数はBEで生成されない。提案・調整待ち・結果不明等は業務状態／操作状態を分け、既存3状態へ勝手に詰め込まない。
- 担当は最新ユーザー訂正により担当医師、指名は非必須。実ケアの入力actor・予定の指名・臨床入院の担当医決定を別にする。責任担当1名必須／非医師可は不採用。人数・新指名UIは推定しない。

## 入院中治療プランの変更契約

[業務正本の最新第1項合意](../../../spec/domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-distinct-plans-and-records)を適用する。追加・変更・中止は同医院・同入院・対象プラン・現在状態と権限を再照合し、元の診療／金額明細と変更前後・入力actor・記録時刻・元履歴との関連を保持する。既存プランの上書き破壊ではなく現在内容と履歴が追跡できる保存を設計する。履歴を失う部分成功は許さず、競合時は再取得、結果不明時は同じ操作結果を照合する。具体API、版／操作識別、DB保存方式、既存権限への対応は実装設計・適合確認として残す。

完了後は理由付き専用訂正を適用し、死亡後の取扱いは別途承認された生前分の専用入力／訂正・退院／精算へ従い、入院中プラン変更から新たな医療行為の許可を導かない。プラン追加を実施済みとみなさず、変更・中止だけで請求／返金・在庫使用／巻戻しを発火させない。D10の既存請求原本・D13の実物使用／出庫契約は各所有領域に残す。治療プラン概算を既存会計へ自動再適用しない。

## null／clear・ケージ競合

既存PATCHのcage_id変更／clear_cage_idには表示時のexpected_cage_idまたはexpected_cage_is_nullのいずれか一方を要する。両期待値、cage_idとclearの併用、ケージ操作なしの期待値は400。行lock後に割当不一致なら409・書込なし。スワップは正数二件のhospitalization_idsと同順expected_cage_ids、ソート順lock・両方検証が必要。

連結メンバー宛は422、占有グループ宛は409。自己の入院を同じ対象へ移す場合の自己占有除外はケージ正本を維持。occupancyのunknown/loading/errorは空きではない。ボード全件走査のclinic headerとcache keyを固定する。ページ走査の同値sort安定性は全件snapshot保証ではないため、表示情報を最終write認可／占有判定へ代用しない。

## 退院会計の現行と目標の差

ソース：[discharge入口](../../../../backend/internal/medicalrecord/hospitalization_discharge.go)・[transaction本体](../../../../backend/internal/medicalrecord/hospitalization_discharge_tx.go)。

| 現行静的確認 | 継承／目標側の扱い |
|:---|:---|
| 入院row lock、already discharged拒否、日付範囲・医院参照を再検査 | 同じhospitalization結果を照合。二回目拒否を初回失敗へ読み替えない |
| 入口／tx内ではdischargedを拒否するがadmitted限定は明示されない | 業務は入院中から退院。UI非表示だけに頼らずserver状態遷移guardを適合確認 |
| CreateAccounting=trueはstaff actor・audit・billing repo必須 | 新規会計／必要監査と退院の原子性。権限はD10の実grantも検証 |
| 既存waiting/pending/completed会計を再利用 | 既存会計IDを返すことと内容更新・再精算は別。確定済み再計算なし |
| cancelled会計をSoftDeleteCancelled後に新規生成 | 最新D10の元取消請求リンク・番号保持・選択明細による新番号再請求へ接続。現行処理を合意扱いしない。取消履歴を見えなくして解放することを新たに許可しない |
| CarePlanItemごとName/UnitPrice/数量1、source=hospitalization | 元care itemとの追跡対応を設計・照合。治療プラン／master日数の自動加算は現行にない |
| 税区分excluded・既定税率、subtotal×税率で税計算 | 最新D10の明細税区分／税率・割引・丸め正本との差。固定税率の現行実装を会計承認済み計算へ格上げしない |
| CreateAccounting=falseのdischarge-with-billing内部は状態更新後return | PATCH退院の必要監査とは違う経路。全退院経路の監査適合を確認し、同じroute名だけで監査済みとしない |
| 元Billingの唯一性と入院行lock | 現行no-duplicate保証の対象。HTTP再送receiptの完備とは別 |

入院プランmasterにper_day/per_nightがあっても、現行退院はその単価／日数を参照しない。ケージmasterの「代表単価を料金に使う」と、退院生成の実ロジックを分離し、料金自動積算の期待を作らない。ホテル個別料金・必要前金・取消料のスタッフ確認入力は追加合意。日数自動課金・取消時自動請求／返金は設けない。

## 新合意へ追従していない既存画面／型

| 差分 | 影響／停止条件 |
|:---|:---|
| createのdefaultHospitalizationStatusはstart_dateが当日以前ならadmitted | 予約だけで開始しない合意と矛盾し得る。実開始intentの設計とlegacy caller切替を要する |
| 現行チェックインはstatus:admitted単純PATCH | 必要入院記録作成と同時成功を証明しない。新合意対応済みとは報告しない |
| 編集formはstart_date/end_dateを送信しない | 延泊／日付変更を保存したように表示しない。予定／実績別の更新契約とdirty対象を設計する |
| 一般PATCHはdoctor_id任意、現行担当選択とホテルの旧UI | 予約指名非必須／宿泊担当医師と臨床入院の開始必須を区別して適合確認 |
| daily logに実施者欄／care-plan task自動linkがない | 現在のstaff memoや入力actorを実施者として補完しない |
| 現行治療planの編集は読取専用、PATCH/DELETEは常にConflict | 入院中の履歴付き追加・変更・中止は最新承認の目標。原履歴を破壊する汎用上書き・DELETEで代用せず、完了後は既存専用訂正へ。現行APIを対応済みとしない |
| 現行deceased guard／死亡時詳細リンク抑止 | 最新合意の生前分後日入力・誤記訂正・退院・精算には専用経路が必要。死亡状態を戻す／通常PATCH全面開放で代用せず、実装適合を確認する |

## 死亡登録後の生前分処理

[死亡後取扱いの最新合意](../../../spec/domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-delete-and-preserve)へ従う。専用経路は死亡状態を維持したまま、生前分であること・対象医院／患者／元記録・必要権限・理由・現在版を保存時に再照合する。実施時点不明は推測せず確認まで停止。原記録と前後差分・理由・actor／確認者・入力時刻・元記録関連を保持し、必要履歴失敗は部分成功にしない。結果不明は元操作を照合し別IDで重複作成しない。現行の通常create/PATCH・一律死亡拒否を専用経路の実装済み証拠にせず、API／版／receipt／権限対応は適合確認に残す。死亡状態の一時解除、通常writeへの無制限な死亡gate解除で代用しない。

退院状態／必要履歴を整合させ、会計へ進む場合も既存の退院＋必要会計明細／監査の境界を維持する。生前分精算はD10の現在権限・請求版・入出金状態を照合する所有領域へ渡し、死亡後記録の保存を入金済みへ変換しない。新たな入院・チェックイン、自動返金・在庫戻しは行わない。

## 再送・競合・取消と照合結果

新規／開始／退院会計は、元医院・滞在ref・操作識別・送信payload・関連結果IDを対応付け、明確拒否と結果不明を分ける。現行APIが永続receiptを返すと仮定しない。結果不明は元operationの結果を照会できる設計が必要で、read0件だけでは遅延commitを否定しない。照会手段がなければ自動再createを停止し、担当者が同じ対象の現状を確認する。

日次／ケアの各saveは独立。成功対象だけdirty解除し、親画面が成功したことを全子保存成功にしない。画像等別ownerや会計へ遷移後も、元操作結果を新しい患者／医院へ適用しない。ケージ409は再取得して現在割当を確認し、ユーザー意思を確認せず古い期待値を置き換えて再送しない。

必要監査の失敗は業務writeとrollback。ただし全日次・ケアwriterに同一監査条件が実装済みとは未確認。全writerの権限・死亡・依存・版・audit適合を実装工程で一覧照合する。

## FR／ACの技術traceと完了境界

| FR ID | AC ID（全件・一意） | 技術対応 |
|:---|:---|:---|
| FR-hospitalization-entry-and-start | AC-hospitalization-planned-not-started / AC-hospitalization-start-doctor / AC-hospitalization-start-failure / AC-hospitalization-two-entry-id / AC-hospitalization-context | 同一滞在ID／担当医・入院記録同時成功 |
| FR-hospitalization-hotel-stay-capacity | AC-hospitalization-hotel-deduplicate / AC-hospitalization-hotel-overdue / AC-hospitalization-hotel-extension / AC-hospitalization-hotel-role | 予定／実滞在重複排除・全期間再判定・ホテル医師任意 |
| FR-hospitalization-distinct-plans-and-records | AC-hospitalization-plan-immutable / AC-hospitalization-plan-add-change-stop / AC-hospitalization-plan-change-failure / AC-hospitalization-plan-change-boundary / AC-hospitalization-care-manual / AC-hospitalization-daily-independent | 原履歴保持のプラン追加・変更・中止／ケア・日次との分離／完了後専用訂正・副作用境界 |
| FR-hospitalization-cage-allocation | AC-hospitalization-cage-null / AC-hospitalization-cage-stale / AC-hospitalization-merge-unknown / AC-hospitalization-cage-available-select / AC-hospitalization-cage-current-retain / AC-hospitalization-cage-unknown-race | 期待ケージ検証・二件交換同一tx・占有不明停止 |
| FR-hospitalization-discharge-and-billing | AC-hospitalization-discharge-no-bill / AC-hospitalization-discharge-atomic / AC-hospitalization-bill-source | 退院＋billing同一tx／明細source・取消元保持 |
| FR-hospitalization-delete-and-preserve | AC-hospitalization-delete-dependency / AC-hospitalization-deceased / AC-hospitalization-deceased-record / AC-hospitalization-deceased-discharge | 親依存・新たな実施拒否／死亡状態保持の生前分専用記録・退院・精算 |
| FR-hospitalization-special-correction | AC-hospitalization-correction-history / AC-hospitalization-correction-no-side-effects | 元ID／理由・actor履歴・同一操作照合 |
| FR-hospitalization-hotel-manual-pricing | AC-hospitalization-manual-hotel-price / AC-hospitalization-cancel-no-auto-money | 個別確認料金／必要前金と実残高分離・自動課金なし |

AC prefixはAC-hospitalization-。本文のFR8件／AC32件を扱う。D01の継承FR・ACは新規承認件数へ二重加算しない。独立レビュー済み候補（追加承認差分は再レビュー）であり、現行差分修正・人数等の未確認と訂正／料金の実装差分・全体横断X03/X06・実装／実機を完了としない。


## 追加承認の実装契約と保存境界

2026-10-15追加合意: 提示文「完了後の誤りは原記録を残して理由付き訂正とし、返金や在庫を自動で戻しません。宿泊料金・必要前金・取消料はスタッフが個別確認して入力し、日数による自動課金や取消時の自動請求・返金は設けません。」を残り方針として再確認し、ユーザーが「ok」と回答（Sentinel_7a67fd30e47c819192cd36d416f0a2bf、確認文 Sentinel_8d0032ca45b881919b3d2ceb5339f792）。業務方針の承認であり、実装・実機PASSではない。

専用訂正は元記録IDと対象医院を固定し、理由・変更差分・actor・時刻・原記録関連を保存する設計目標とする。保存時に対象・権限・現在版を再評価し、原記録保護と訂正履歴を一体で確定する。理由欠落・権限不足・版競合は原記録を変更せず拒否する。通信結果不明は同じ対象の結果を照合し、新しい訂正を追加して重複させない。成功・確定失敗・結果不明を区別し、入力を保持して同一操作として再送する。現行汎用PATCHの存在はこの専用契約への適合証拠ではなく、API/receipt/履歴・権限の実装確認が必要。訂正に金銭返却・在庫戻しを含めず、各所有領域の操作へ引き継ぐ。

料金・必要前金・取消料の確認入力も同じ医院／滞在・予約／飼主を照合して保存する。必要額の保存とD10受領は別操作であり、前者成功を入金成功と表示しない。取消の状態変更と料金確認は自動請求・返金を発火させない。現行CarePlanItem単価変換を、新しいホテル料金フォーム・必要前金receipt・取消料入力が実装済みの証拠にしない。

<a id="hospitalization-change-receipt-candidate-20261019"></a>
## CL-R07：滞在内変更と訂正の資源案（2026-10-19）

既存routeは[入院・治療プラン登録](../../../../backend/internal/medicalrecord/routes_hospitalization.go)、金額inputは[plan service](../../../../backend/internal/medicalrecord/treatment_plan_service.go)。inputの存在は履歴付き変更・完了後専用訂正の実装成立を証明しない。基準 `563353287ebb08af55056e3fbfba2aca3531e599` の限定静的参照。

候補は滞在refに属する安定plan item ref＋revisionと、不変の変更event refを分離する。eventは追加／変更／中止intent、元版、前後payload、認証actor／時刻、必要時の理由／確認者、operation結果refを持つ。現在内容projectionの更新・変更event・必要監査・成功receiptを同じtxへ置き、完了後は同じeventを通常PATCHへ流さず専用訂正intentで元履歴を参照する。既存データを現masterから遡及復元せず、履歴不明を不明として移行する。

ホテル料金候補は滞在・予約・飼主refに対する料金確認ref／版と、通常料金・必要前金額・取消料の区別を持つ。料金保存receiptからD10の入金receiptを生成しない。D10への引継ぎは型付き元行refと保存版を使い、請求済み元行を二重追加せず、確定後は既存会計訂正へ渡す。日数masterからの自動課金・取消自動返金・在庫巻戻しを追加しない。

検証は入院中同時変更、中止と別編集の両順、完了後理由欠落、履歴監査失敗、保存commit後切断、死亡後生前分、料金保存後未入金を分ける。原版不変・最新projection整合・同key結果収束・金銭／在庫副作用なしを記録し、実行前は未検証とする。

<a id="cage-availability-read-20261019"></a>
## 空きケージ候補と現在割当のadapter（10/19目標への接続）

[業務目標・3追加AC](../../../spec/domains/hospitalization/lifecycle.md#cage-available-selection-20261019)に従う。限定コード読取（基準 `14100df845e14813f99b02aad0b38f589b2f26a4`）では、`cage_repository.go`の占有集合は同院・未削除admitted/reservedで、日付重なり条件を持たない。`hospitalization_service.go`のvalidateCageAssignmentGuardは構成員422、連結グループ占有409、自己占有除外を持つが、未連結ケージはlen(group)<=1で占有検査を行わない。この既存差を「全ケージの重複拒否実装済み」と説明しない。

候補readは既存のケージ候補参照認可・医院scopeで、ケージID/連結関係/空き既知または不明と判定contextを返すadapter案とする。連結ピッカー専用`master-hospitalization:edit`のoccupancy APIを登録フォームへ無条件転用し、新たな管理者権限を登録条件にしない。全ページ/表示filter外を含む占有集合と対象医院/編集対象refを対応し、患者名等の不要な占有明細を公開しない。新API名・DTO・grant・DDLはこの文書で実装しない。

自己占有除外は編集対象入院IDに対応する占有1件だけを根拠付きで除外し、同ケージ/グループに別滞在があれば占有を残す。占有ケージID集合だけから「自分以外なし」と判断せず、対象別の照合能力をreaderへ持たせる案とする。

clientは現在割当表示と新規候補集合を分離し、自己以外の占有で選べない値と、過去に保存された現値の表示を混同しない。医院/対象入院が変われば旧候補応答を捨て、占有取得不明時に空き候補を捏造しない。期間入力は既存の日時検証へ渡すだけで、未採用の期間予約アルゴリズムを作らない。新規割当保存では同じ医院/対象/連結・占有を現在状態で再照合する必要があり、未連結を含む競合拒否・同時保存の適合は実装残件。既存group lock順・期待割当・二件交換同時成功を弱めず、新ロック/物理方式は実装設計で照合する。

実機は空き/占有/別院/取得失敗/現在値/自己除外/連結後旧値/選択後競合/期間だけ変更/医院切替の遅延応答を3ACへ対応し、候補表示と保存結果・元割当不変を別々に記録する。未割当可の既存ACを併用し、候補制限だけで実装・実機合格を宣言しない。

## 10/19横断検索・同種操作の設計接続

[共通adapter](../../shared/integration-contracts.md#cross-ui-search-adapter-20261019)を本領域の[適用台帳](../../../spec/shared/interaction-scope.md)L/M/H/S consumerへ接続する。draft/applied・独立filter・取得scope/世代・復帰/失敗状態を分離し、検索ボタンだけでキーワードを適用。既存API検索と取得済み内絞込みの責務、ページング/上限、認可・保存方式を維持する。E/N操作へ部品単位で一括適用しない。実装/全caller/実機未確認。
