# 入院・ホテルの技術契約

| 項目 | 内容 |
|:---|:---|
| 文書ID・種別 | D05-TECH / 技術設計・現行差分 |
| 決定状態 | 既合意の受け手設計と既存契約の継承。専用訂正・ホテル個別料金は2026-10-15追加回答で承認（実装適合は別） |
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
| 治療plan PATCH/DELETEは常にConflict | 追加合意の訂正でも既存snapshotを通常更新可へ開かない |
| deceased guardと終了／訂正 | 専用救済が未決なら既存guardを汎用PATCHで迂回しない |

## 再送・競合・取消と照合結果

新規／開始／退院会計は、元医院・滞在ref・操作識別・送信payload・関連結果IDを対応付け、明確拒否と結果不明を分ける。現行APIが永続receiptを返すと仮定しない。結果不明は元operationの結果を照会できる設計が必要で、read0件だけでは遅延commitを否定しない。照会手段がなければ自動再createを停止し、担当者が同じ対象の現状を確認する。

日次／ケアの各saveは独立。成功対象だけdirty解除し、親画面が成功したことを全子保存成功にしない。画像等別ownerや会計へ遷移後も、元操作結果を新しい患者／医院へ適用しない。ケージ409は再取得して現在割当を確認し、ユーザー意思を確認せず古い期待値を置き換えて再送しない。

必要監査の失敗は業務writeとrollback。ただし全日次・ケアwriterに同一監査条件が実装済みとは未確認。全writerの権限・死亡・依存・版・audit適合を実装工程で一覧照合する。

## FR／ACの技術traceと完了境界

| FR ID | AC ID（全件・一意） | 技術対応 |
|:---|:---|:---|
| FR-hospitalization-entry-and-start | AC-hospitalization-planned-not-started / AC-hospitalization-start-doctor / AC-hospitalization-start-failure / AC-hospitalization-two-entry-id / AC-hospitalization-context | 同一滞在ID／担当医・入院記録同時成功 |
| FR-hospitalization-hotel-stay-capacity | AC-hospitalization-hotel-deduplicate / AC-hospitalization-hotel-overdue / AC-hospitalization-hotel-extension / AC-hospitalization-hotel-role | 予定／実滞在重複排除・全期間再判定・ホテル医師任意 |
| FR-hospitalization-distinct-plans-and-records | AC-hospitalization-plan-immutable / AC-hospitalization-care-manual / AC-hospitalization-daily-independent | 不変snapshot／mutable care／独立日次保存 |
| FR-hospitalization-cage-allocation | AC-hospitalization-cage-null / AC-hospitalization-cage-stale / AC-hospitalization-merge-unknown | 期待ケージ検証・二件交換同一tx・占有不明停止 |
| FR-hospitalization-discharge-and-billing | AC-hospitalization-discharge-no-bill / AC-hospitalization-discharge-atomic / AC-hospitalization-bill-source | 退院＋billing同一tx／明細source・取消元保持 |
| FR-hospitalization-delete-and-preserve | AC-hospitalization-delete-dependency / AC-hospitalization-deceased | 親依存・死亡gate／原記録保持 |
| FR-hospitalization-special-correction | AC-hospitalization-correction-history / AC-hospitalization-correction-no-side-effects | 元ID／理由・actor履歴・同一操作照合 |
| FR-hospitalization-hotel-manual-pricing | AC-hospitalization-manual-hotel-price / AC-hospitalization-cancel-no-auto-money | 個別確認料金／必要前金と実残高分離・自動課金なし |

AC prefixはAC-hospitalization-。本文のFR8件／AC24件を扱う。D01の継承FR・ACは新規承認件数へ二重加算しない。独立レビュー済み候補（追加承認差分は再レビュー）であり、現行差分修正・人数等の未確認と訂正／料金の実装差分・全体横断X03/X06・実装／実機を完了としない。


## 追加承認の実装契約と保存境界

2026-10-15追加合意: 提示文「完了後の誤りは原記録を残して理由付き訂正とし、返金や在庫を自動で戻しません。宿泊料金・必要前金・取消料はスタッフが個別確認して入力し、日数による自動課金や取消時の自動請求・返金は設けません。」を残り方針として再確認し、ユーザーが「ok」と回答（Sentinel_7a67fd30e47c819192cd36d416f0a2bf、確認文 Sentinel_8d0032ca45b881919b3d2ceb5339f792）。業務方針の承認であり、実装・実機PASSではない。

専用訂正は元記録IDと対象医院を固定し、理由・変更差分・actor・時刻・原記録関連を保存する設計目標とする。保存時に対象・権限・現在版を再評価し、原記録保護と訂正履歴を一体で確定する。理由欠落・権限不足・版競合は原記録を変更せず拒否する。通信結果不明は同じ対象の結果を照合し、新しい訂正を追加して重複させない。成功・確定失敗・結果不明を区別し、入力を保持して同一操作として再送する。現行汎用PATCHの存在はこの専用契約への適合証拠ではなく、API/receipt/履歴・権限の実装確認が必要。訂正に金銭返却・在庫戻しを含めず、各所有領域の操作へ引き継ぐ。

料金・必要前金・取消料の確認入力も同じ医院／滞在・予約／飼主を照合して保存する。必要額の保存とD10受領は別操作であり、前者成功を入金成功と表示しない。取消の状態変更と料金確認は自動請求・返金を発火させない。現行CarePlanItem単価変換を、新しいホテル料金フォーム・必要前金receipt・取消料入力が実装済みの証拠にしない。
