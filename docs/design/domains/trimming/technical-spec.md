# トリミングの技術契約

| 項目 | 内容 |
|:---|:---|
| 文書ID・種別 | D06-TECH / 技術設計・現行差分 |
| 決定状態 | 既存契約の継承と安全条件の具体化。完了後専用訂正は2026-10-15追加回答で承認（実装適合は別）。死亡後の生前分専用処理、担当／コース未定の途中保存も最新承認を反映 |
| 実装状態 | 2026-10-15に限定ソース静的照合。全writer／HTTP／DB適合未確認 |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-15 |

業務FR／ACは[施術lifecycle](../../../spec/domains/trimming/lifecycle.md)、語彙は[オントロジー](../../../spec/ontology/trimming.md)。wire正本は[API YAML](../../../../backend/docs/api.yaml)、DBはmigrations。既存詳細画像契約は[施術フォーム§画像](../../../spec/screens/17-trimming-form.md)を保持する。

## ID・writer・現在のモデル

施術の外側はappointments行、詳細はappointment_trimming_details、選択肢はappointment_trimming_options。予約IDと独立した通常medical_record IDを混用しない。detail未作成の予約も正常な読取結果であり、通信エラーをdetailなしへ変換しない。

reservationがappointment lifecycleの唯一writer。trimmingはCreateForTrimming／UpdateForTrimming／DeleteForTrimmingを通して予約を更新し、detail/optionsを所有する。course/option/course-type masterもtrimming owner。[write-owner catalog](../../../architecture/model-write-owner-catalog.md)と[orchestration catalog](../../../architecture/cross-domain-orchestration-catalog.md)のPATH-TRIM-APPT-*に従う。

業務担当はトリマーだがwire StaffIDはappointments.doctor_idへmapされる。doctor_idという列名は職種指定の根拠にしない。指名／個人対応／勤務はD01、staffの所属／有効性はD14、実施記録actorは認証済staffから取得する。

## 保存・transaction・依存

| 操作 | 現行確認 | 失敗／競合 |
|:---|:---|:---|
| 直接create | booking advisory lock→appointment→detail→options、同一tx、必要staff actor／監査 | いずれか失敗で全体rollback。孤立予約を残さない |
| 既存予約へdetail create | 既存予約lock、mutable状態・同医院owner/pet検査、既存detailなら409 | detail有無の競合を再取得。別予約createにfallbackしない |
| update | booking lock／row lock、master再検査、予約intent＋detail/options＋監査 | 参照／状態／競合で全体拒否 |
| delete | 予約intentの医療記録依存検査＋削除監査same-tx、最新画像値をtx内取得 | 画像cleanupはcommit成功後。古い読取画像だけを削除候補にしない |
| course/options検証 | tx前fast checkに加えtx内SHARE lockで再検査 | master無効化のTOCTOUを閉じる現行経路 |
| 次回来院tag同期 | commit後best-effort | 台帳保存成功と外部tag成功は別。D15運用停止を解除しない |

参照：[service](../../../../backend/internal/trimming/trimming_service.go)、[create](../../../../backend/internal/trimming/trimming_service_create.go)、[update](../../../../backend/internal/trimming/trimming_service_update.go)、[mutate](../../../../backend/internal/trimming/trimming_service_mutate.go)。

CreateInput／UpdateInputはservice型でありHTTP schemaそのものではない。serviceの任意値だけを業務承認の根拠にせず、[最新第5項合意](../../../spec/domains/trimming/lifecycle.md#trimming-partial-save-undecided)の担当／コース未定の途中保存を目標とする。新規経路の旧画面必須条件と既存予約経路の双方を照合し、全体への適合済みとはしない。record_shortcutによる予約制約の適用差はD01の正本を維持し、任意のstatusを指定すれば全制約を回避できるという仕様にしない。

## 担当／コース未定の途中保存

新規・既存予約へのdetail作成・既存記録の続き入力で、担当のみ未定／コースのみ未定／両方未定の値を表現し、再取得後も未定のまま保持する。患者・トリミング区分・医院／予約scopeと既存の認可・状態・安全制約は維持する。未定を表すための架空staff ID、既定course ID、料金の自動補完を作らない。未取得／取得失敗と未定を区別し、取得失敗で既存参照を消さない。

現在のStaffID→doctor_id対応、course参照、create必須validation、既存予約からの補完、PATCHの未送信／明示解除とDB nullableを照合する。任意pointerの存在だけでUI・HTTP・全writerが適合済みとしない。未送信で既存参照を保持する契約は維持し、途中保存の許可だけで新しい汎用clear APIを定義しない。後から指定する値は既存の担当適格性・有効master条件で検証し、同じ予約/detail・必要監査の保存単位を維持する。失敗・結果不明は既存の入力保持・同一対象照合を用いる。

途中保存を「進行中」等の表示へ対応させることと、実施／完了の許可判定を分離する。新状態enumや実施／完了時の新しい必須条件は今回決めない。旧新規画面の担当・コース必須は解消対象の実装差、必要wire adapterと全入口・再表示の適合は未確認。

## PATCHと部分入力

| 値 | 現行意味 | 注意 |
|:---|:---|:---|
| updateのpointer nil | 未送信で変更なし | nullによる汎用clearを推定しない |
| OptionIDs nil | 変更なし | detail読込失敗を空配列へしない |
| OptionIDs non-nil、空配列 | 全置換して全解除 | 差分追加APIではない |
| style_image／completed_image省略 | 画像維持 | signed URLを送らない |
| 画像の空文字 | 明示clear | recordの参照解除とstorage削除完了は別 |
| course/staff/body values等pointer | 項目固有更新 | 全nullableのclear対応／共通revisionは未確認 |

新規で省略したstatusのservice既定はpendingだが、画面新規既定「進行中」は明示送信される別契約。両者を同じ既定と書き換えない。表示状態「予約／進行中／完了」は予約wireの投影であり、D10入金状態を表さない。

## 画像objectとrecordの保存境界

以下のlock／tombstone／storage cleanupは現行実装の観察であり、物理削除を業務許可する目標仕様ではない。保存済み原画像・添付、参照済みkey、業務原本／必要履歴か分類不明の画像はC05の原本保持対象とし、通常削除・clear・置換・取消でblobを物理削除しない。通常操作は論理削除／非表示とし元参照・必要監査を保持する。未保存かつ未送信の放棄作業copyと区別し、分類不明は保護する。物理削除は承認された別手続へ委ね、本整理から新しい例外や保存年限を推定しない。

1. uploadはrecord保存より先。即時local previewとupload成功keyを区別する。失敗なら直前keyを保持。
2. record送信開始後のkeyはサーバ側の所有として扱い、結果不明でもクライアントcleanupしない。更新に成功したか不明なまま削除すると、遅延commit済recordを壊すため。
3. 新規保存／削除はclinic配下の正規key。別名path、他医院keyを拒否。旧HTTP(S)値は未変更の表示互換だけで内部storageを署名／掃除しない。
4. record保存はbooking/appointment lockの後、最終画像key advisory lockを昇順に取り、tombstoneの不存在を確認しcommitまで保持。
5. cleanupはkey lock、全参照確認、永久tombstone commitの後にstorage削除。別record参照ありは409。PATCH cleanupも自recordを参照確認する。
6. storage失敗でもtombstoneを消さず、新規参照は拒否。同keyのDELETEで冪等再削除可。workerのlast_attempt/storage_deleted記録と完了通知遅延は画面正本へ。
7. worker既定停止、TRIMMING_IMAGE_CLEANUP_ENABLED=trueの場合の既存設定をこの整理で有効化しない。5分／run120秒／key10秒／最大100医院・100keyという現行上限はフォーム正本を維持。
8. clinical変更のaudit_logsと削除予約専用sinkは別。診療画像の削除台帳と統合しない。保存期間の共通未決を理由に永久tombstoneの参照禁止を解除しない。

## 会計・訂正の引継ぎ

billing_item_repositoryがトリミング明細を未請求明細として統合読取する。完了時pushとしない。料金マスタの表示価格、detail参照、取込済請求snapshotを区別する。現行detailのcourse/option参照だけから当時価格の保存を保証しない。

D10へ渡す確認単位はclinic、owner、pet、appointment/detail、course/option、元明細、取込／請求状態。税・割引・保険・数量／価格・再請求の計算はD10正本に従う。D06側の更新で既存確定請求を上書きしない。取込後に施術内容やmasterが変わる場合の差分表示／版照合と、旧データにsnapshotがない場合の扱いは実装前にD10と照合する。現在値から過去値を架空生成しない。

専用訂正の業務方針は2026-10-15追加回答で承認。通常reservation Undo、通常カルテaddendum、トリミングPATCHの存在を完了後訂正の許可として使わない。確定会計の保護・記録依存を「誤り修正」で回避しない。死亡登録済みの生前分は最新合意の理由・履歴付き専用経路へ接続し、死亡gateを通常writeで全面解除しない。

## 死亡登録後の生前分記録

[死亡後記録の最新合意](../../../spec/domains/trimming/lifecycle.md#requirement-fr-trimming-safety-and-status)に従う。専用経路は死亡状態を維持したまま、生前分であること・対象医院／患者／元記録・必要権限・理由・現在版を保存時に再照合する。実施時点不明は推測せず確認まで停止。原記録と前後差分・理由・actor／確認者・入力時刻・元記録関連を保持し、必要履歴失敗は部分成功にしない。結果不明は元操作を照合し別IDで重複作成しない。現行の通常create/PATCH・一律死亡拒否を専用経路の実装済み証拠にせず、API／版／receipt／権限対応は適合確認に残す。死亡状態の一時解除、通常writeへの無制限な死亡gate解除で代用しない。

既存の予約／施術参照・原記録・完了後専用訂正・確定会計保護を維持し、死亡後の未来予約や新たな施術を作らない。精算・返金等はD10へ渡し、この専用記録保存から自動返金・在庫戻しを行わない。

## 通信・再送・遅延応答

- clinic＋予約ID＋患者＋認証＋画面generationを固定し、遅い応答を新contextへ混入しない。
- 保存成功、明確拒否、結果不明を区別。成功対象のdirtyだけ解除。画像アップロード成功はrecord保存成功ではない。
- booking lockと重複detail409は並行性保護であり、全createの永続idempotency receiptと同義ではない。既存予約ID付き操作は同じ予約を照合し、direct createの結果不明は自動再createを止める。
- 現行APIの全経路にoperation ID／payload hash／結果照会が存在するかは未確認。必要なreceipt設計を別に選定して検証し、GET0件で遅延成功を否定しない。
- 古い版のUPDATEを防ぐ共通revisionは未確認。booking lockは直列化するが利用者の古い入力の上書きを自動検知する保証ではない。
- 離脱時cleanupは未送信未保存画像だけ。送信中の離脱／Cancelからmutation取消を推測しない。

## 現行画面との照合事項

| 既存記載 | 明確化する差 |
|:---|:---|
| 予約枠が一覧へ自動出現 | 予約detailが未作成でも一覧対象になり得る。施術記録全文が既に生成済みとはしない |
| 完了から会計へ流れる図 | 会計側pullであり完了イベントで会計生成しない。取込可能状態の全実装条件はD10照合 |
| PatientInfoCardの担当医ラベル | 業務トリマーとwire列名を区別。語彙変更をコード反映済みとはしない |
| master料金を引き継ぐ | 価格snapshot時点・取込後変更を未検証のまま断定しない |
| master course-type「スタッフ管理者以上」 | 実resource/actionが正本。職位から権限を推定しない |
| 保存権限と画像権限 | create/editのupload cleanupと記録deleteを分離 |
| 完了後の削除／更新 | mutable状態・会計確定歴・既存依存と、承認済み専用訂正の目標契約を分ける |

## trace・文書完了境界

| FR ID | AC ID（全件・一意） | 技術対応 |
|:---|:---|:---|
| FR-trimming-reservation-and-detail | AC-trimming-existing-reservation / AC-trimming-duplicate-detail / AC-trimming-direct-time | reservation intent／detail・options・監査同一tx |
| FR-trimming-input-master-history | AC-trimming-new-required / AC-trimming-save-undecided-staff / AC-trimming-save-undecided-course / AC-trimming-save-undecided-both / AC-trimming-undecided-resume / AC-trimming-undecided-scope / AC-trimming-inactive-master / AC-trimming-history-copy / AC-trimming-size-pricing | 未定の途中保存／既存対象再開／実施完了条件を追加しない境界・有効master／既存無効保持・履歴copy対象限定 |
| FR-trimming-safety-and-status | AC-trimming-deceased-loading / AC-trimming-deceased-history / AC-trimming-context-result / AC-trimming-stale-update | 現在患者／新たな施術拒否・死亡状態保持の生前分専用記録／状態・版 |
| FR-trimming-image-lifecycle | AC-trimming-image-inflight / AC-trimming-image-failure / AC-trimming-image-clear / AC-trimming-image-race / AC-trimming-image-cross-clinic / AC-trimming-image-only-dirty / AC-trimming-image-original-retention | 正規key lock／参照・原本保持・送信済key保護 |
| FR-trimming-billing-handoff | AC-trimming-billing-pull | 未請求pull／追加時snapshotとD10再適用境界 |
| FR-trimming-delete-and-atomicity | AC-trimming-atomic-write / AC-trimming-delete-dependency | 依存検査・detail/options同一tx・監査 |
| FR-trimming-special-correction | AC-trimming-correction-history / AC-trimming-correction-no-side-effects | 原記録関連／理由actor時刻・金銭在庫別owner |

AC prefixはAC-trimming-。FR7件／AC28件と既存画面・master2件、現在の型／保存／画像・原子性・他owner境界を整理した文書範囲。最新の途中保存承認差分を含め独立内容レビュー・限定構造検査済み。専用訂正の実装、全writer互換／実データsnapshot／権限・receipt実装、X03/X06、実機・運用開始は未確認。


## 追加承認の実装契約と保存境界

2026-10-15追加合意: 提示文「完了後の誤りは原記録を残して理由付き訂正とし、返金や在庫を自動で戻しません。宿泊料金・必要前金・取消料はスタッフが個別確認して入力し、日数による自動課金や取消時の自動請求・返金は設けません。」を残り方針として再確認し、ユーザーが「ok」と回答（Sentinel_7a67fd30e47c819192cd36d416f0a2bf、確認文 Sentinel_8d0032ca45b881919b3d2ceb5339f792）。業務方針の承認であり、実装・実機PASSではない。

専用訂正は元記録IDと対象医院を固定し、理由・変更差分・actor・時刻・原記録関連を保存する設計目標とする。保存時に対象・権限・現在版を再評価し、原記録保護と訂正履歴を一体で確定する。理由欠落・権限不足・版競合は原記録を変更せず拒否する。通信結果不明は同じ対象の結果を照合し、新しい訂正を追加して重複させない。成功・確定失敗・結果不明を区別し、入力を保持して同一操作として再送する。現行汎用PATCHの存在はこの専用契約への適合証拠ではなく、API/receipt/履歴・権限の実装確認が必要。訂正に金銭返却・在庫戻しを含めず、各所有領域の操作へ引き継ぐ。

## 一般更新の同時編集契約（選定仕様）

元表示版と保存時の現在版を比較し、不一致なら現在の原記録を変更せず競合を返す。入力を保持し、再取得した現値と差分を表示してスタッフの再確認を受ける。期待版だけを新版へ置換して無警告再送しない。明確拒否と結果不明を区別し、結果不明は元医院・対象ID・同一操作識別／内容で照合する。別操作IDや別createで救済しない。既存一般PATCH/UPDATEの期待版・永続receipt対応は実装差であり、この選定動作の実装済みを意味しない。
