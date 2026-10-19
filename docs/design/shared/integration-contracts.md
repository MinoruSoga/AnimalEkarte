# 技術接続基盤と文書検査（C06）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | design.shared.integration-contracts／共通契約の整理候補 |
| 決定状態 | 既存記載を継承。12:53 UTCの3業務決定は各正本へ接続。内部方式・wire・DCは提案で、文書全体の本人承認や実装済みを表さない |
| 実装状態 | 既存資料の記載を継承（コード全体の再監査なし）。選定仕様と既存挙動の差を明記 |
| 実機検証状態 | 未実施 |
| 文書確認日・基準 | 2026-10-15／PR #670 commit 41fcf3eb0a8befb7c0b5f613d1df749a60e30be8 |
| 対象 | 全owner接続の技術文書契約・最終文書照合。実装変更と全writerの実装適合検証は別工程 |

本書は既存正本への対応と受入観点を束ねる。数値・語彙・業務条件の第二正本は作らず、リンク先の有効な個別契約を優先する。FR/ACの識別子付与は追加承認や実装済みを表さない。未確認の不足を対象外へ移して完了としない。

## 正本・読む順

[Architecture](../../architecture/README.md) → [overview](../../architecture/overview.md) → [FE／BE map](../../architecture/fe-feature-be-domain-map.md) → [write-owner](../../architecture/model-write-owner-catalog.md) → [横断処理catalog](../../architecture/cross-domain-orchestration-catalog.md) → 各domain技術仕様。wireは[api.yaml](../../../backend/docs/api.yaml)、DBはmigration（[ERD入口](../../architecture/erd.md)）。本書へ全DTO・全table定義をコピーしない。

C01の部品技術、C04の認証／医院scopeと個別Dの操作認可・ID・状態は別の責務。D01–D23の文書管理境界はBE packageを再編しない。API仕様・実装事実だけで業務判断を採択しない。

<a id="requirement-fr-common-tech-owner"></a>
### FR-common-tech-owner — 一事実一writer

一つのbusiness factは一つのwrite-ownerへ。appointmentsはreservation、billing graphはbilling、staff勤務はstaff等の既存ownerを維持。shared model配置は共有writerの許可ではない。consumer-sideの最小interfaceまたはtyped intentを通し、他ownerへ任意field更新・独立write pathを増やさない。共通ID／enumを複製しない。FEのshared昇格は実consumer2以上・理由・domain固有API/RBAC知識を持たない条件を満たすものに限る。

<a id="requirement-fr-common-tech-boundary"></a>
### FR-common-tech-boundary — tx・失敗・回復

複数owner writeはambient transactionへの参加か、commit順・部分成功・回復・観測を明記したorchestrationへ。必要なtx／依存／検証／監査が欠けるintegrity経路はfail-closed。best-effortは明記した経路だけに許容し、全処理を勝手に同一txへ格上げしない。逆に既存実装のbest-effortを、最新合意の原子性を弱める根拠にしない。

既存catalogにある各候補単位same-txと、全医院batch loopのmulti-txを区別する。timeout／応答喪失／失敗時は各ownerのID・receipt・再読取・再送契約に従い、単にPOSTを再実行して冪等と呼ばない。未定の一律Version欄・ETag・HTTP code・全業務共通retry回数を新設しない。

<a id="requirement-fr-common-tech-contract"></a>
### FR-common-tech-contract — 各接続に必要な確認項目

| 項目 | 各Dで具体化する内容 | 未記載時の扱い |
|:---|:---|:---|
| 対象とwriter | initiator、owner、操作intent、読取と書込 | package名だけで適合扱いしない |
| identity／scope | clinic、owner、pet、record／appointment／billing ID、認可と対象照合 | client指定IDを認可根拠にしない |
| wireと版 | request/response実キー・型、未送信／null／clear、状態enum、依存版／旧値 | api.yaml／DTOの差は未確認として残す |
| 保存単位 | draft/local／即時保存／確定、tx owner、lock順、必須監査sink | 同一画面を一括txと推測しない |
| 競合・失敗 | 入力拒否、権限失効、not-found、競合、通信失敗、commit後応答喪失 | 各状態で入力保持／再取得／訂正の期待結果が必要 |
| 再送・取消 | 同一効果識別、payload相違、receipt、対象取消／補償 | 一律Undo・再POST・自動再試行を追加しない |
| 読取consumer | snapshot／原事実／派生表示、遅延応答、患者context、section scope | 最新masterで過去事実を置換しない |
| 外部通知 | commit前後、部分成功、停止、観測、再実行の主体 | 内部成功と外部成功を混同しない |
| 保存・監査 | actor、時刻、理由／元先、必要tx、保持／照会・非再利用 | C05当面方針へ接続し法令適合とはしない |
| 互換・検証 | 新旧client、旧値停止／移行前提、FR/AC→技術→証跡 | 実装未完了と仕様未定を別記 |

<a id="selected-vs-current"></a>
## 選定仕様と現行実装資料の差

| 接続／資料 | 既存資料の記載 | 現在の選定仕様／扱い | 仕様残件・確認先 |
|:---|:---|:---|:---|
| PATH-RES-MR-AUTOCREATE | 予約confirmed/checked_in保存後にseparate-tx best-effortで通常カルテ自動作成 | [D03保存正本](../../spec/domains/medical-records/lifecycle.md)は通常予約時作成なし、受付時同時成功。古い経路を新承認にしない | D01/D02/D03の選定設計と旧caller停止／移行照合 |
| PATH-RES-MR-CANCEL-CLEANUP | 予約取消／削除後にdraft cleanup、separate-tx best-effort | 同正本は通常予約取消からの自動削除禁止。draftというだけで空判定しない | 専用の空・関連なし削除と旧cleanupを分離 |
| PATH-BILL-LSTEP-CPM-SYNC | 会計成功後Created=trueのみ同期、失敗は会計を戻さず、同会計replayでは再同期しない | intentional best-effortを保持。後続同期・停止条件とD15/D10を照合 | 「再送すれば回復済み」としない。実稼働未確認 |
| PATH-BATCH-TAG-HEALTH | implemented capabilityだがproduction scheduler/cron未接続、戻り値error | durable scheduled batchと呼ばない | D15の導入／配信再開を別扱い |
| support write-owner行 | 全医院報告を全staffが更新・削除・Plane起票可という記載 | FE/BE mapは共有READ、報告医院mutation、他医院404と記載 | [service](../../../backend/internal/support/service.go)／[repository](../../../backend/internal/support/repository.go)のFindByIDForClinic／clinic_id条件を読取確認。報告医院mutationが一致し、ownerカタログは旧誤記として修正対象。自動外部通知の新方針とは別 |
| RequestID | Gin context／応答header／完了logに保持 | request Contextへ自動伝播しない既存制約を保持 | 全downstream logへ存在すると主張しない |

当表は既存カタログ本文の全面書換えを承認するものではない。統合時に古い現行実装記述へ「選定仕様との差」を明記し、原子性・監査・回復の重要例外を消さない。

<a id="requirement-fr-common-tech-evidence"></a>
### FR-common-tech-evidence — 文書検査と証拠の境界

[testing](../../ops/testing/README.md)のL0–L5入口と実行記録を使う。文書検査は相対リンク／アンカー、FR/AC一意性とtrace、本文／既存ID保持、表構造、明示scope、source時点を扱う。コードsymbol drift・88route実行検査・runtime・性能・全make ciは実施していなければ未実施とする。過去83route実測・85route E2E inventoryや個別PASSを現在88route全件PASSへ拡張しない。

## 受入条件

| 条件ID | Given／When | Then／独立確認・成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-common-tech-owner | 他領域の事実を変更 | owner intent／interface・tx主体が特定できる | owner catalogと各D設計 |
| AC-common-tech-same-tx | 必要監査／依存が失敗 | integrity経路は対象writeと一緒に失敗し、成功を偽装しない | 各pathのfail-closed AC |
| AC-common-tech-partial | documented best-effortの副作用失敗 | 部分成功・回復主体・監査／観測が区別できる | catalogと最新選定仕様 |
| AC-common-tech-wire | 未送信／null／旧値が届く | ownerごとの保持／clear／拒否条件とwire根拠がある | API/DTO/技術trace静的照合 |
| AC-common-tech-replay | commit後応答喪失・再送・payload相違 | 同一効果と新規操作を判別し、競合復旧の期待結果が記載される | 各D receipt／lock契約 |
| AC-common-tech-selected | 旧auto-create／cleanupを見る | 現行実装資料と受付時原子性／自動削除禁止の選定仕様を区別 | 本表＋D01/D03正本 |
| AC-common-tech-exception | support／global master／jobを扱う | 通常clinic-scopeを乱暴に一律適用せず、正しい例外根拠を確認 | C04/D23／ADR／owner |
| AC-common-tech-doc-test | docs-only検査を報告 | 実施内容・基準commit・未実行チェックを明示 | 検証manifest／独立レビュー |
| AC-common-tech-owner-coverage | 全D候補を統合 | 上記接続項目の具体化とFR/AC traceが全ownerで確認済み | 全ownerの意味ID／版・保存／失敗・旧値・受入traceを独立照合済み |

## 未確認・完了境界

本書の技術契約と全Dの意味型・ID／版・tx／lock・通知・失敗回復・互換をowner対応表とX接続で最終照合した。D04／D06、D11〜D13、D22／D23の操作別保存・版・receipt／結果不明などの不足は本文とACへ具体化し、独立レビューで未解消の重大指摘なし。物理名や実装方式の確定だけを後続へ分離し、呼出側／受け手が意味契約を推測する穴を実装対象外として隠さない。実装・API/DB・設定変更、稼働再開、実機／性能／法的適合の実行は別工程。

## 全ownerへの適用照合

[全23領域の具体ID・保存単位・再送／旧値・安全と保持](../../spec/shared/owner-contract-coverage.md)と[横断8接続](../../spec/workflows/reservation-to-record-flow.md#cross-domain-review-20261015)を参照する。全writerの実装適合とは別に、呼出側／受け手が意味契約を推測しない文書境界を照合する。

<a id="detail-design-proposals-20261018"></a>
## 残件の詳細設計案と採用境界（2026-10-18）

基準は `509f050902766de3d356e187f61388f4530282d8`。本人の「改善して」に基づく文書整備であり、以下の内部方式・候補wire・DC検証ケースは**設計提案**。後続の12:53 UTC本人「1A・2A・3A」の業務決定は各業務正本へ反映し、内部方式の承認と混同しない。直接承認済み業務、既存の選定技術、現行実装観察と区別する。API YAML・DDL・コードへ適用していない。以前の「具体方式未選定」は当時の状態で、下表の提案まで具体化したが、実装・移行の成立や全残件解消を表さない。

| 範囲／責任owner | 既決として保持すること | 今回の詳細設計案の正本／残件 |
|:---|:---|:---|
| D01＋D03・LINE問診 | 予約×ペット別、受付前訂正・旧回答保持、任意、スタッフ原文変更不可 | [回答世代と予約gate](../domains/reservations/technical-spec.md#revisit-answer-storage-proposal)。受付／対象変更との競合を直列化 |
| D03・統合入力／編集権 | hybrid保存・子owner・通常終了後編集、閲覧取得なし、単独編集、確認付き強制取得 | [保存単位とfence](../domains/medical-records/technical-spec.md#integrated-save-edit-right-proposal)。自動失効秒数は新設しない |
| D04／D19／D20・共通患者とLINE | 自院利用／共通基本情報と院別記録、認可・手動同一確認、既存link/unlinkの安全条件 | [共通参照と解除](../domains/owners-pets/technical-spec.md#common-subject-link-proposal)。物理ID移行は計画・dry-run適合が必要 |
| D03＋D10・確認と同期 | 金額のみ変更は警告了承、臨床／数量は再確認、未確定同期3分岐、確定後訂正 | [臨床世代と同期台帳](../domains/accounting/technical-spec.md#review-version-sync-proposal)。既決条件を全体version一つへ潰さない |
| D08／帳票reader・証明書 | 3様式、発行版固定、訂正新発行・旧版履歴、最新版再印刷、旧自動登録なし | [固定版と出力](../domains/vaccinations/technical-spec.md#certificate-edition-proposal)。通常縦横の最新版共有Q-V1は本人確定、旧版停止Q-V2は既決の技術適用として訂正保存成功へ結ぶ |
| D05＋D12・退院帳票 | 実退院のJST日・4項目、金額を再加算しない | [退院projection](../domains/finance-operations/technical-spec.md#discharge-report-projection-proposal)。行粒度は技術案 |

### 操作・版・失敗の接続案

今回詳述しないownerの残件も、[既存23領域対応表](../../spec/shared/owner-contract-coverage.md)へ戻して管理する。下表は既存契約から実装・検証へ渡す入口で、新しい未承認業務の一覧ではない。今回コードを限定再読した領域と、従来文書を参照するだけの領域を同じ実装調査済み件数にしない。

| 領域 | 次に検証／具体実装する対象と、今回保持する境界 |
|:---|:---|
| D01／D02／D18 | 回答target案と受付／既存record再利用・開始read・候補再判定の接続。全writer gate・旧auto-create/cleanup排出は未実装検証 |
| D03／D07／D09 | 編集fenceと独立検査／健診結果の依存projectionを分離。診断・構造化問診・結果revision・旧未知値のround-trip、機器／旧client／移行writerを実装前に全列挙 |
| D04／D19／D20 | common/local mappingとLINE binding案。既存group依存／単回token、他院認可・本人公開範囲は維持。Q-I1は具体的矛盾移行／分割が必要な時だけの条件付き検討、独立解除Q-I2は本人確定：予約／患者関連保持・通知停止・再連携後staff選択 |
| D05／D06 | stay／専用記録／予約の同一性・ケージ交換・料金snapshot／元調整一度適用・退院会計の原子性。通常カルテの編集権や退院帳票から新しい取消／復帰操作を作らない |
| D08 | 固定発行版案、本人確定Q-V1と既決適用Q-V2、種類別次回日と過去当時日。旧見本の項目対応と描画・旧client互換の実装検証 |
| D10／D11 | 確認世代・同期3分岐、金銭履歴・既請求防止、見積提示原本／後継と確認転記。同期から確定済み請求や提示原本を更新しない |
| D12／D13／D16 | 退院read案、締めsnapshot／差分、実物移動と数量版、実来院と集計source。実入金・充当・実返金を二重加算しない。実査precision／通貨の実装適合は在庫整数と別検証 |
| D14／D17 | シフト確認集合／CSV一括保存、個人能力・勤務・医院設定の保存版。設定や勤務の変更と予約保存の逆順競合、実grant／旧値移行を検証 |
| D15／D21 | dispatch claim案と停止／結果不明、院内レポートの1頭・section認可と専用2帳票。新たな配信再開や全区画印刷・飼主公開は追加しない |
| D22／D23 | 公開記事版と旧版復元、shared READ／報告医院mutation／本人履歴、外部issue結果照合。原本／添付blob保持、外部投稿の認可と再送は各ownerの既存契約を実装検証 |

Q-C2のlogout後本人専用保持・同本人再ログイン回収は本人確定。新しい保持期限・自動破棄・別主体／患者への回収は追加しない。既存の保存年数・廃棄手続、薬量の具体値／対象／単位の医師承認、外部契約・実grantの実装確認は別工程であり、今回の3回答を未決へ戻さない。未取得2資料は本人指示どおりスキップのまま。数値・法的保存年数や未読資料固有条件を推測で補わない。

予約の[既存操作／receipt契約](../domains/reservations/technical-spec.md#reservation-api-matrix)を参照し、以下の新adapterだけに対応案を定める。全APIに新しい共通必須fieldを一括導入する承認ではない。既存数値IDと増分版は新wireでは正規形の十進文字列、operation key・新規common subject等のUUID識別子は種別付きUUIDとし数値へ変換しない提案とする。いずれもref kindで区別し、既存数値IDは精度保証可能なadapterで照合する。tokenやstampのopaque値をactor／認可の代用にしない。

| 意味field（候補名） | 具体的契約 |
|:---|:---|
| `subject_ref` | owner/kind/idと必要clinic、関係先pet/record/appointmentをownerが照合。clientから任意clinic/actor/statusをwriteしない |
| `expected_revision`／`dependencies` | 各ownerの現在版と依存版を区別。更新時に比較・更新・増分を同じtxへ。子依存版・臨床世代・編集fenceを本体versionで代用しない |
| `operation_key`／fingerprint | 認証clinic＋actor kind/id＋intent＋keyで一意。同keyで対象・業務内容が異なれば拒否。前提版・現在の編集権証明はfingerprintと分け、再照合する |
| `operation_result` | 初回成功の対象ID・確定版・必要監査参照。現在状態は別の`current_state`。後の取消・権限変更を成功時状態へ巻き戻さない |
| 未成功／成功のbinding | 内容bindingの短いtxと、binding行lock＋業務write＋必須監査＋成功結果のtxを分ける既存予約案を再利用。前者の残存は業務成功ではない。同内容再試行は前提を再検証し、先行処理の終了前に追越さない |
| 結果照会 | 作成後IDを受け取れなくても同scope/keyで照会可能。現在の対象・操作の閲覧認可を再検査し、認可失効なら結果本文を漏らさない。処理中は待機表示、成功は初回結果、rollback後は未成功として同内容の再試行候補、再構成不能は結果不明のまま停止 |

同内容未成功の前提版を更新する場合も、現値との差分を人が確認する。業務内容変更は先行結果を解決してから別key、未解決操作を同内容の別keyへ分裂させない。ID一意制約・元対象の既請求等も併用し、keyだけで二重実行を防いだとしない。未成功bindingもresponse cache期限で再利用せず、保存・廃棄はC05の当面自動削除なしと別途年数判断に接続する。監査へ本文・LINE識別子・tokenを無目的に複製しない。

| 結果区分（候補code、未実装） | client／serverの扱い |
|:---|:---|
| validation／not-authorized／scope-not-found | writeなし。入力は許可されたローカル範囲で保持し、認可外の現値を返さない。404/403の使い分けは既存ownerの非開示方針を維持 |
| version-conflict／dependency-changed | writeなし。現値・差分を認可付き再取得、再確認。期待版だけ自動差替えしない |
| edit-right-lost | 新write拒否、元入力保持。再取得後に最新比較・明示選択。元権でcommit済みの事実は取り消さない |
| busy／operation-in-progress | 必須lockを無期限に待たず、結果を同keyで再照会。別key再createなし |
| transport-unknown／result-unavailable | timeout/abortはrollbackの証拠ではない。成功／未成功が判明するまで自動補償・再create・外部再送を止める |

### lock・切替・検証の責任

既存[予約の全writer取得規約](../domains/reservations/technical-spec.md#reservation-global-lock-contract)を優先し、予約が関係する経路は同じ共通gateへ参加する。非予約writeの局所lockはowner内で安定ID順、共通gateを取得する処理より外側で逆順に取らない。新participantのreceipt/binding・編集権・子依存を含めた順序は各Dで具体化し、DB deadlock検出を整合性保証の代用にしない。未参加旧writerを残したまま新方式を稼働させない。

切替前に、全writer一覧（UI／旧client／GetOrCreate／移行／batch／機器・外部取込）→必要gate・版・権限・監査・receipt→readerの対応をownerが登録する。既存ID／履歴を保持するadapterを先に検査し、新方式で表現不能な旧値を架空の正常へ補完しない。移行途中のread可と、新write可の開始条件を分ける。

DC-C01：同key同内容の並行実行・成功後対象版変更・rollback後再試行・同key異内容を別々に検査。DC-C02：commit直後切断、結果取得失敗、権限失効、遅延応答、再構成不能でも二重write／情報漏洩なし。DC-C03：旧writerと新writerの同時write、監査失敗、依存版更新、逆順lockの両先行順序を検査。各DCは実装時の検証設計であり、本人の追加業務承認や実行PASSではない。

<a id="posthumous-operation-contract"></a>
## 追加承認の死亡後記録と専用操作の接続

[C03患者状態と死亡後記録](../../spec/shared/clinical-safety.md#requirement-fr-common-safety-patient)のAC-common-safety-posthumous-historyは、D03〜D10の当該操作の期待版・現在認可・対象元事実・理由／監査・結果照会・原子receiptへ接続する。死亡状態を生存へ書き換えて通常新規ゲートを迂回せず、元患者・生前実施事実・原記録を保った専用の後日入力／誤記訂正／退院／精算として識別する。必須履歴を保存できない操作を成功にせず、結果不明時は同じ操作を照会し、新しい操作として二重実行しない。死亡後の新しい臨床実施や未来予約、金銭・在庫・通知の自動作用を追加しない。物理schema／wireの実装状態とは分離する。

[追加9点の進捗](../../spec/domains/SPECIFICATION_PROGRESS.md#additional-business-decisions-nine)の業務承認と文書再整合を別に追跡する。過去の文書完成・静的検査合格から、既存記載全ての業務承認を導かない。

## 2026-10-15会議4点の操作契約

通常カルテの[自動保存・復元](../../spec/domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)、[終了と型付き既存記録](../../spec/domains/reception/pages/reception-board.md#reception-meeting-start-end)、[施術料金のD10受渡し](../domains/accounting/technical-spec.md#accounting-trimming-adjustment-adapter)、[勤務一括入力](../domains/staff-workforce/technical-spec.md#shift-bulk-operation-contract)を別intentとして照合する。元対象ID・現在scope／権限・期待版・依存・操作ID／digest・成功receiptと実効current_stateを混同しない。結果不明では元操作を照会し、再送で新しい復元・終了・請求・勤務を二重作成しない。復元は当該タブ／複製の既知対象と基準に限定し、他者編集・別ownerの金銭／在庫／通知へ逆実行しない。具体API名・物理schema・実装適合は後工程だが、これらの意味契約は各owner本文へ定義する。


<a id="remaining-design-evidence-20261019"></a>
## 2026-10-19 残件の具体化と適合判定の入口

基準main `563353287ebb08af55056e3fbfba2aca3531e599` の残作業は[文書対応の照合記録](../../work/docs-audit/docs-perfection/REMAINING-RESPONSE-20261019.md)へ接続する。各owner正本へ追補した資源・wire・列・一意性・移行adapterは内部候補で、既決業務やD01の既選定gate/schemaを置換しない。候補を採用したという記録、実API/DDLとの適合、実データ移行、実機の各状態を別々に記録する。

臨床CL-R01、金銭FIN-R12、予約/受付IR-11はそれぞれのownerの更新経路を受け持つ。共通gateの証跡は再利用し、各ownerについて実symbol/版・認可・依存/監査/成功結果照会/reader/旧caller切替条件を埋める。今回の限定route/DTO読取を全writer参加の証明にはしない。未知入口・未対応旧writerを利用開始可にせず、実装工程で根拠と停止条件を明示する。

未検証31項目は[環境・対象・操作・期待の引継ぎ](../../work/docs-audit/docs-perfection/REMAINING-VERIFICATION-20261019.md)へ。静的観察で旧挙動と目標の差を確認しても、文書をその旧挙動へ戻さない。実機/DB/外部送信/運用操作が未実施なら未検証のままとし、今回の文書対応をGo-live・法令適合の証拠にしない。
