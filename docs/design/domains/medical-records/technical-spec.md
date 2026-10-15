# カルテ初回合意の部分技術設計

<a id="medical-record-autosave-revert-contract"></a>
## 2026-10-15最新差分：通常編集・自動保存・復元操作・診療終了

業務正本は[lifecycle最新合意](../../../spec/domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)。明示保存方式への途中提案は撤回。従来hybridの送信境界を維持し、通常カルテは終了後も編集できる。以下の2026-10-14設計にある「医師確定intent」「finalized親の一律通常編集拒否」「addendumだけ」「確定時警告」は、本節の対応表で置換する。現request／DTO／serviceの説明は現実装参照として保持し、コード・schema・設定の適合を宣言しない。過去保存版・追記・目的別訂正・削除保護や他owner独自のロックは維持する。

### 入力・保存・旧finalizedの適合

| 経路 | 最新の目標 | 現実装との適合確認 |
|:---|:---|:---|
| 問診・診察／プラン・治療・接種・健診・検査・画像・バイタル・ヘッダー | 既存の明示／自動／即時保存先を維持。親finalizedだけを理由とする通常入力／編集拒否を置換 | 全writerの認可・医院／患者・期待版・必要履歴を保持し、GetOrCreate／内部participant／関連変更・外部入力も照合。新grantや全員編集可へ広げない |
| 医師確定UI／status PATCH | 新しい通常操作としての医師確定ボタン／intentは採用しない | 旧APIの受理範囲を新仕様としない。既存finalized・確定actor／時刻・snapshotは履歴として保持し、全件draft化・unfinalizeで消去しない |
| 旧finalizedの通常編集 | 同じrecord／子IDの現在の内容を通常画面で更新し、変更前版とactor／時刻・必要理由を履歴へ対応付ける | 現親draft-only gateを単に削除せず、履歴付き更新・目的別保護・版競合のgateへ置換。旧client迂回・履歴欠落を許可しない。schema／移行方式は未実装 |
| 既存addendum／後日結果 | 以前の本文・参照版・追記と現在の内容の対応を保持 | addendum APIを廃止／データ削除せず、通常入力の唯一の道にも固定しない。結果ID／revisionはD07正本を維持 |
| 目的別の医師確認・患者訂正・実施取消・生前過去入力・非表示／復元 | 既決の理由・確認・同ID・元先関係・依存を保持 | 通常編集可を患者ID付替え、記入済み削除、死亡後新規行為、確定済会計の変更許可へ流用しない |
| D07検査confirmed、D11見積承認後、D05/D06専用記録、会計／在庫 | 各ownerの独自状態・確定／訂正・履歴を保持 | カルテ親finalizedロックの置換と、別ownerの保護解除を混同しない |

### タブクリア／複製取消の復元契約

復元の対象は画面の入力状態であり、既に自動保存された変更があれば永続状態にも反映する。ローカル表示だけ戻して成功にしない。元の保存版・必要監査・関連IDを廃棄せず、新しい復元操作の履歴として記録する。物理DELETE・storage object削除・履歴全消去をUndoの実装に選定しない。

| 意味データ | 保持する内容／契約 |
|:---|:---|
| 対象参照 | clinic／patient／recordとtab、owner／kind／id／revisionを持つ型付き参照。ローカル未発番行は保存成功後の実IDへ対応付け、他患者への遅延応答を適用しない |
| タブ開始基準 | そのタブを開いた時点の入力値・参照／存在状態・保存済み版・開いた世代。自動保存／再取得で基準を上書きしない。他タブ・タブ外識別情報を含めない |
| 複製直前基準 | 元複製operation、source／target、選択対象、直前値／版、今回作成した行ID・成功subset。タブ開始基準とは別に保持 |
| 自分の操作と他者版 | タブ開始後の自分の入力writeと複製writeを識別。現在の版／依存と照合し、他者編集又は取消対象外の後続変更があれば差分を表示して競合を解消するまで無条件復元しない |
| 保存状態 | 未送信、送信中、成功、拒否、版競合、結果不明。送信中／不明を先に照合し、遅い自動保存応答が復元後を再上書きしないよう世代／操作順を管理 |
| 復元write | 基準値、現在期待版、対象操作との対応、認証actor、必要理由／差分・時刻、operation key。通常保存と同じ認可・安全・履歴・親子gateを通し、必要ログと同じ保存単位で成功させる |
| 新規行の復元 | タブ開始又は複製直前に存在しない行は、適法な論理状態／関連の復帰として扱う。元ID・作成と復元の履歴を保持し、既存の保存済み原本やobjectを物理削除しない。削除／関連変更のowner保護を迂回しない |
| 結果／再試行 | 成功対象ID／新しい版、基準との対応、必要履歴ref、成功／未実行／失敗／結果不明を返す。部分未完了を全体成功にせず、同じoperation照会／再送で二重復元を防ぐ。非対応旧APIは盲目的再送しない |
| 保護すべき実行 | 請求・入金・返金・在庫・予約終了／再開・検査確定・既送信通知は逆実行しない。対象入力が既に実行に使われた場合は影響を表示してowner訂正へ。原実行receiptと復元結果を別に保持 |

自動保存済み値を戻すための複数write／外部objectを含む原子性は、既存owner境界を踏まえた実装設計とする。全タブ／全owner一括txを新たな承認済み条件としない。既知の部分成功を消さず復帰結果を個別照合し、安全な補償が未対応なら対象と理由を表示して停止する。技術未実装を、業務上合意したクリア／取消そのものの未承認と混同しない。

### 診療終了・警告・履歴の接続

終了要求は同じ予約ID／期待版と、通常カルテの型付きref（owner／kind／id／revision）、患者／医院、治療プラン依存版、終了operation identityへ対応付ける。入力保存の結果、予約終了の結果、会計／入金の結果は別であり、record IDを予約IDやaccounting IDとして代用しない。終了結果は同じ予約の現在状態・終了operation result／receiptを参照する。既存予約gate・lock順／必要履歴へ参加し、リンク不明なら作成／付替えで救済せず停止する。

未実施・後日予定は終了時に警告し、終了操作権限のあるスタッフの確認をrecord／plan依存版に結び付ける。取得不能・不完全は「全完了」にしない。確認後に依存が変われば再取得／再確認。確認後は未完了でも終了可能で、全項目完了・医師確定・請求／入金確定へ変換しない。自動保存にはこの未実施警告を発火させず、旧会計医師確認receiptを診療終了の新しい必須条件へ流用しない。

会計待ちカードは同予約の保存済みrecord linkを読み、同じ通常カルテへ遷移する。新規create・複製・開始を副作用にしない。終了再送／結果不明は元操作を照合し、画面遷移成功を終了成功と扱わない。

スタッフ履歴のread modelは予約ownerが確認できる診療終了とrecord linkを掲載条件にする。finalized、支払済み、自動保存成功を終了の代用にしない。診療中の途中保存は掲載しないが、その保存版／監査を削除するわけではない。終了後の通常編集は認可された現在版と履歴関係を表示する。過去専用入力／移行記録で終了根拠が不足する場合、実装時のadapterで確認可能な事実を対応付け、不明を架空の終了へ移行しない。D21への公開・出力境界はD21正本へ渡す。

### 最新FR／ACのtraceと残件

- 通常編集：FR-medical-records-normal-edit-after-end、AC-medical-records-normal-edit-after-end／no-finalize-button。旧finalized-addendum-only／coverage-finalized-all-writersはIDを保持して期待値を置換。
- 自動保存：FR-medical-records-preserve-autosave、AC-medical-records-autosave-preserved、既存active-tab-only／header-immediate／child-independent／save-partial-dirtyを維持。
- タブクリア：FR-medical-records-clear-tab-open-state、AC-medical-records-clear-tab-open／clear-empty-open／clear-other-tabs／clear-conflict／clear-pending-unknown。
- 複製取消：FR-medical-records-history-copy-undo、AC-medical-records-copy-undo-before-copy／copy-undo-partial／copy-undo-later-edit。共通AC-revert-keeps-owner-effects／revert-failure-keeps-state。
- 終了／警告：FR-medical-records-screen-consultation-end、既存explicit-finalize-plan-warningの置換、AC-medical-records-screen-end-waiting／waiting-opens-existing／no-plan-warning-on-autosave、既存plan-incomplete-warning／plan-warning-cancel／plan-deferred-warning／coverage-finalize-concurrent-depsの置換。
- スタッフ履歴：FR-medical-records-ended-record-history、AC-medical-records-ended-history-only／ended-history-edit。上記ACは共通prefix AC-medical-records-。業務本文・正式ID／GWTはlifecycle／画面正本に一元化。

基準snapshot・操作receipt・期待版が現API／旧データに揃うとは確認していない。旧finalized履歴付き編集、複製成功subsetと復元、全writerの親gate置換、別owner副作用保護、履歴掲載read modelの実装／移行／並行性と実機検証が必要。これまでの16FR／71ACの文書完成記録はその時点の記録であり、この追加差分の実装完了／全体再受入へ流用しない。

| 管理項目 | 状態 |
|:---|:---|
| 対象 | 通常カルテの初回5項目・入力4項目・訂正／例外4項目・画像印刷／後日結果／訂正影響3項目。業務正本は[ライフサイクル](../../../spec/domains/medical-records/lifecycle.md#medical-record-agreement-first-five) |
| 根拠 | Sentinel_1b70a4b2675c8191814931fd13762b54。技術方式は合意結果を守るdocs設計 |
| 実装／検証 | API YAML・DB・コード・設定未変更。実データ／runtime／並行性・性能未検証 |
| 範囲外 | 複写全体atomic化・自動template提案、返金／締め後の会計内部・専用記録本文 |

## 保存・認可・ログ

[保存正本§2](../../../spec/domains/medical-records/lifecycle.md#hybrid-sync)のヘッダー／アクティブタブ／子入力の送信先を保持する。全タブの一括保存を追加しない。各writeの現在grant・医院／患者scope・版・保存履歴と目的別保護を検査し、担当医一致を一般入力の追加必須条件にしない。医師確認は認証済み確認者と医院で認めた医師の資格／現在権限を照合し、client boolean・doctor_id・予約指名から推定しない。医師の判定と職員実データ対応は確認対象で、架空のrole・permission IDを増やさない。

業務上必須の編集者ログは既存audit ownerへ目的別payloadで参加させる。actor・医院・record／子対象・操作・変更項目／必要差分・時刻・operation結果と確認医師を保持する。本文全体・患者情報の無条件複製はしない。業務writeと必須ログを同じtransactionで成功させ、ログ失敗なら双方rollbackし結果不明時は照会／同一操作再送で重複を避ける方式を選定する。入力失敗や送信draftを成功編集と記録しない。現[通常CRUD](../../../../backend/internal/medicalrecord/medical_record_crud.go)のbest-effort監査は静的実装参照で、本合意へ追従済みではない。

## 確定・警告・競合

旧見出しは参照互換のため保持する。現在の対象は[診療終了時の警告と競合](#medical-record-autosave-revert-contract)であり、医師確定intentを通常操作に要求しない。臨床必須欄を追加せず、保存処理と終了処理を分離する。未実施／後日予定の現在状態・依存版を確認して終了時に警告し、権限あるスタッフが確認すれば終了可。取得不明は停止、確認後の依存変更は再確認、自動保存時の未実施警告なし、実施済みへの自動更新なし。

既存の[全writer規約](../reservations/technical-spec.md#reservation-global-lock-contract)・[受け手契約](../reservations/technical-spec.md#reservation-receiving-contracts)を維持し、同じ親／子依存の現在版と保存・終了の順序を照合する。旧ClinicalPlan／TreatmentPlanの値だけで4状態を推定せず、入院snapshotを通常カルテの警告対象へ自動混載しない。医院が認めた医師の確認は薬量等の目的別条件であり、通常の終了権限へ一律移植しない。

## 死亡後の専用入力／訂正の適合差

[D03の死亡後適用](../../../spec/domains/medical-records/lifecycle.md#medical-record-deceased-retrospective)は既決の専用過去入力を死亡登録済みの生前診療にも適用し、誤記訂正を理由・履歴付きで可能にする。専用経路は死亡状態を維持したまま、生前分であること・対象医院／患者／元記録・必要権限・理由・現在版を保存時に再照合する。実施時点不明は推測せず確認まで停止。原記録と前後差分・理由・actor／確認者・入力時刻・元記録関連を保持し、必要履歴失敗は部分成功にしない。結果不明は元操作を照合し別IDで重複作成しない。現行の通常create/PATCH・一律死亡拒否を専用経路の実装済み証拠にせず、API／版／receipt／権限対応は適合確認に残す。死亡状態の一時解除、通常writeへの無制限な死亡gate解除で代用しない。

退院はD05、精算はD10へ。通常治療createでinventory指定時に起こる現行減算は専用後日入力／訂正へ発火させない。未実施プランの作成を実施済みや実物使用／出庫に変換しない。D03の既存AC-retrospective-reason-physician／retrospective-existingへ死亡登録済みのケースを追加して追跡し、初回過去入力合意の再承認・実装PASSには数えない。

## 削除・訂正・過去入力

[現削除処理](../../../../backend/internal/medicalrecord/medical_record_crud.go)のdraft・診療中予約・見積等の依存保護と論理削除を継承し、子入力を含むempty/entered/unknownと依存revisionを同じlock下で検査する。親本文空欄だけで空にしない。[現repository Delete](../../../../backend/internal/medicalrecord/medical_record_repository.go)とDeletedAtによる論理削除を静的確認。物理削除を選定せず、理由付き復元intentは追加合意へ接続する。実際のrestore API／grant対応は後続照合。予約の誤取消復元と、論理削除カルテの復元は別で、予約restoreの存在をカルテrestore APIの根拠にしない。論理削除の履歴・予約／記録同一性と復元可能性の実態は調査を残す。

記録確認訂正は提案revision・理由・変更案→認可済み医師確認→予約／記録／請求依存の再照合・同時commit。旧finalizedだけで通常本文を[既存addendum](../../../../backend/internal/medicalrecord/medical_record_addendum_service.go)専用へ限定しない。元の保存版／日付・追記履歴を保持し、患者・予約日等の目的別訂正を満たす。患者取り違えは必要内容転記・元／先対応へ接続し汎用patient PATCHのfallbackを作らない。請求訂正の具体手順は別owner残件。

過去入力は通常のreception/start intentと別の専用intentとして理由・実診療日・登録時刻・確認医師とoperation keyを保持する案。既存患者／recordを明示照合し同名／日付だけで統合しない。過去受付status・実時刻を自動生成せず、未確認の期間制限・精算手順は設計済みとしない。過去日入力の確定とカルテfinalizedは別段階で、登録成功をカルテ自動確定に置換しない。

## 評価不能時の手入力

不足情報／取得失敗／評価可能・絶対上限超過を別結果として返す。評価不能時は医師確認・理由・手入力値と対象薬／現情報revisionを確認して保存し、必要編集ログと確認証跡を同時commitする。確認後に体重・種・パラメータが変われば再評価し、評価可能な超過は拒否する。無条件skipを確認済みへ変換しない。認可・対象scope／薬剤の照合・必要取得依存の失敗とspecies不一致の既存fail-closedは維持し、非NotFound取得エラーを未設定や医師確認付き不足入力へ正規化して迂回しない。実際の確認者と編集者が別なら両者を保持する。

現[薬量保存の画面参照](../../../spec/domains/medical-records/pages/medical-record-form.md#23-臨床安全ガード)との差分を調査し、薬量数値・対象・単位・出典の医師承認は別工程。上限解除grant・安全範囲の新しい数値を追加しない。API URL／DTOと医師確認の具体的UIは後続設計、実装正本のAPI YAMLを更新したとは扱わない。

## 今後の設計・受入

初回5FR／16ACはライフサイクルの正本へ対応。この節は部分設計時点の残件記録で、全タブの型・入力状態・異常系、子入力／編集ログ・医師確認・削除／訂正／過去入力・互換切替の文書契約は末尾の最終照合A〜Fへ追従した。実装／全実行経路の適合は未検証。訂正／例外4項目の業務許可は追加正本へ従い、具体的保存対応・全タブ網羅・他owner連携を解消して技術仕様を拡張する。入力4項目の合意結果は下記部分設計へ接続する。全ドメインの文書確定・再照合前に実装しない。

## 入力4項目の接続設計

2026-10-14の[入力4FR／15AC](../../../spec/domains/medical-records/pages/medical-record-form.md#medical-record-agreement-input-four)を部分設計へ追加。初回5FR／16ACを保持し、既存wire／DB／コードへ反映済みとしない。

| 対象 | 選定する技術接続 | 現行差分・残件 |
|:---|:---|:---|
| 選択複写 | 元recordと選択対象を明示して認可付き読込。問診はlocal draft、上書きだけ確認。治療は今回の有効masterと価格を保存時に再照合し既存個別createへ。値・価格のsourceを元履歴と混同しない | [現copy hook](../../../../frontend/src/features/medical-records/hooks/use-copy-treatment-details.ts)の全明細複写と無型／不一致fallbackは未追従。sort順の逐次保存・部分成功は維持。成功ID／失敗／未送信・結果不明を区別し、全件atomic化・成功済み再作成はしない |
| Template適用 | 選択したtemplate本文を問診draftへ追記。置換だけ確認。元template更新や自動API保存を発火しない。grant不足は適用不可、旧finalizedだけで通常適用を拒否しない、取消は元入力保持 | [既存master CRUD](../../../spec/screens/settings/master-interview.md)と挿入を分離。問診欄とカテゴリの対応、挿入UI、template取得失敗／古い選択の扱いは詳細設計。自動提案と固定カテゴリ・必須欄は追加しない |
| 診療master履歴 | 新規項目保存時にその項目が持つ名称・単位・価格と元master ID／revisionをrecord／子明細snapshotへ保持。master変更で既存snapshotを上書きせず、既存の未変更参照保持と新規選択を区別する。無効化だけで過去の参照解除を強制しない。master ownerの現在有効性とscopeを最終再照合する | 現API／保存値／参照名の全域棚卸しは未実施。snapshotの存在を推定せず旧データを根拠資料で照合。数値承認は医師、編集は対象master grantを持つstaff。今回価格は新しい複写先だけに適用 |
| 項目状態・実施済み | 通常カルテの明示plan item refと医院・期待版を対象に、未実施／実施済み／後日予定／中止を保存。認証actor・記録時刻・前後状態・中止時の理由とoperation結果を必須ログへ同時commit | [現TreatmentPlan](../../../../backend/internal/model/hospitalization.go)／[ClinicalPlan](../../../../backend/internal/model/clinical_plan.go)の保存対応は調査中。入院snapshotと通常カルテ項目を混在させず旧状態を実施済みに推測補完しない。実施者／実際の実施時刻を今回必須にせず記録者へ自動転記しない |
| Checkbox／modal | 項目行のcheckboxと実施操作modalを同じ状態intentへ接続し、同じ対象・版を保持。送信前previewと保存後状態を分け、cancel／失敗は前snapshotと入力を保持、結果不明は照会で再確認 | 両UIを必ず順番に踏む業務は選定しない。二重click／二つの入口で二度writeしない。解除を無条件の実施取消writeへ変換せず、追加合意の理由付き取消モーダルへ。詳細配置は既存診察／治療プランのタブと編集ロックへ整合する |
| 確定警告（旧名） | 対象recordの項目状態と依存revisionを取得し、未実施／後日予定は診療終了時に警告。終了権限staffの確認と最新依存を再照合して終了、項目状態は変更しない | 中止理由・記録者ログを維持。取得失敗／旧状態未照合を全部完了へ落とさない。既存保存結果・親／子版を保持し、警告は診療終了へ移す。会計医師確認を終了の新必須条件にしない。実施記録と請求作成／取消の自動連動は未承認 |

複写のsuccess subsetとoperation結果は同医院／actor／source targetに対応させ、遅い応答を別患者へ適用しない。マスタ無効化・価格改定・競合・医師確認失効が判明したら、未送信分を再選択／再確認へ渡し、成功済みsnapshotを無断で書き直さない。元カルテの編集ログを今回の編集者へ変更しない。投与量は複写先の現在情報で初回の医師確認・理由付き手入力と評価可能な上限拒否を満たす。

既存の保存§2の送信先と個別saveを維持し、新API endpoint／必須実施者入力・課金生成を未承認のまま追加しない。実施取消は追加合意の専用契約へ接続する。master版／state保存schema／各writer gateとAPI生成・旧consumer切替・並行性／実データ・性能は未実施。状態・値取得で欠落がある場合もユーザーの指定済み4状態／基本操作を再質問せず、合理的な型・取得・表示を設計し、追加合意にない新しい訂正・再開・実績／請求結果の許可だけを業務確認へ戻す。

## 訂正・例外4項目の接続設計

[追加4FR／14AC](../../../spec/domains/medical-records/lifecycle.md#medical-record-agreement-correction-four)に従うdocs限定の部分設計。endpoint／schema追加やruntime充足の宣言ではない。

| 専用intent | 選定する契約 | 守る境界・残る照合 |
|:---|:---|:---|
| 項目の実施取消 | 対象record／項目・期待版・取消理由・変更前後・認証actor／記録時刻・operation keyを保持。モーダルcancelはwriteなし、状態と必要ログを同時commit | 旧状態不明は照合し推測しない。Checkbox解除を直接writeへ変換しない。旧finalizedでも保存版・訂正履歴を保持して目的別操作。会計影響は会計訂正へ。診療終了Undoの許可を拡大しない |
| 患者取り違え | 元record／元患者と正しい患者／転記先recordを別IDとして明示。誤記・理由・必要内容の選択・確認医師・元／先対応・依存revisionを保持し再照合 | 元ID／元患者関係を保持し、patient_id PATCH・元削除・全件自動複写を代用にしない。以前の保存版の本文／日付は不変で、現在版の通常編集とは分離する。検査・画像・請求など別ownerの関連は勝手に移動せず、対象owner訂正へ引継ぐ |
| 過去専用入力 | 実診療日とサーバーの現在登録日時を別に保持。理由・actor・確認医師・患者／既存記録の照合と再送識別を保持 | 当日受付カード・開始status・自動課金を作らない。請求必要性は会計で別途確認。通常作成経路の例外にせず、移行用未紐付けを安易に流用しない |
| 空誤作成の非表示／復元 | 非表示の元ID・履歴を保持。復元は理由・actor・期待版と現在の医院／患者／依存・一意性・操作認可をlock下で照合、状態と必要ログを同時commit | 権限あるスタッフによる復元で医師確認を新たな必須にしない。競合／不明は停止、別ID再作成なし。開始／課金／確定を副作用にしない。既存DeletedAtと復元入口・実grantの対応は実態調査 |

各intentを[全writer gate](../reservations/technical-spec.md#reservation-global-lock-contract)へ参加させ、通常CRUDの迂回と確認後の古い依存でのwriteを防ぐ。結果不明は同じ操作のreceipt照会／再送で照合し、新しいkeyで重複転記／複製を作らない。患者転記は通常の再診用複写と目的が異なり、入力4項目の「今回価格」を転記対象の歴史的事実へ無条件適用しない。新しい請求金額・在庫処理は会計等ownerの未決事項で、転記の副作用にしない。取消先の具体状態・転記フィールド・履歴payload・復元認可mappingは既存データと全タブを照合する技術残件。

全タブのフィールド／consumer・子writer、別ownerとの対応、正常／失敗／取消／競合／結果不明のAC追跡と実機は未完了。保存§2の個別送信先・以前の保存版／owner保護・選択複写の部分成功を維持し、部分合意13FR／45ACをD03全体完了へ読み替えない。

## 全9タブ・画面共通・owner接続の残件棚卸し

既存画面の全9タブを対象とする網羅確認表。この表は最終照合前の残件配賦の履歴で、現在のdocs契約は最終照合A〜F、実装適合は利用開始前の未検証事項へ分ける。新しい必須項目や業務許可を定義しない。合意済み操作の詳細設計を進め、追加業務だけを入口の質問表へ戻す。

| 対象 | 既決の制約／参照 | 技術・ACの残件／受渡し |
|:---|:---|:---|
| 問診 | 必須追加なし・主訴区分任意、選択複写・手動template | 欄と型・ローカルdraft、上書きcancel・個別保存・離脱／取得失敗、master版と旧値。部分ACあり、全field未網羅 |
| 診察／治療プラン | 4状態・実施modal・中止理由・未完了警告、理由付き実施取消 | ClinicalPlanと項目状態の保存対応、子依存revision／確定との競合、前後履歴・取消先照合。入院snapshotと分離 |
| 治療 | 今回の複写価格・既存当時値・逐次部分成功、薬量評価不能の医師確認 | 処置／処方各field・数量単位・保存とログ・薬量再評価／数値別承認。D10／13との金額／在庫は未確定 |
| 予防接種 | 既存ペット単位API・履歴と次回予定の既存境界 | D08で接種記録／後日訂正・患者／記録対応を照合。親finalizedだけの通常編集拒否を置換し、独立writerも以前の保存版を破壊しないAC |
| 定期健診 | 既存checkup編集入口と対象ID | D09で既存行と新規行・型／結果／後日訂正を照合。専用記録を通常カルテへ統合しない |
| 検査 | 既存結果の取込・対応と比較表示 | D07の同じ結果／履歴を参照し旧確定時を含む保存版＋後日結果履歴を保持。取込失敗／重複・旧結果adapterを今回契約へ |
| 画像 | 既存形式・件数／サイズ制限、D&D追加・範囲選択印刷は追加3項目で合意 | 添付／削除・upload途中失敗・確定保護、別owner履歴・公開権限を照合。D&Dと選択の共通検証・訂正追記を含む印刷は今回契約へ |
| 見積書 | 既存独立保存・依存として削除を保護 | D11の金額snapshot・版／保存結果・訂正から請求までを照合。空判定から除外しない |
| 会計（医師確認） | 会計ownerの確認済み／未確認／取得不明と通常カルテ編集を分離 | D10の医師確認receipt・依存revision・差戻し、未納／分割／0円・訂正／締め後とのAC。自動課金なし |
| 一覧・患者ヘッダー・共通操作 | 通常受付の同ID、患者取り違えは転記、空誤作成の非表示／復元 | 全filter／shortcut・読み込み・認可／医院切替、ヘッダー個別保存、バイタルモーダル／個別保存・確定保護・体重変更時の薬量再評価、元／先対応、一意性・全子ログ／再送を網羅 |
| 他の専用記録・公開 | D05／06の入院・ホテル・トリミング、D20／21の読取とレポートを保持 | X01／02／05〜07で全writer／consumerと相互照合。通常記録の4状態・新許可を自動波及させない |

共通の正常／取消／保存失敗／必須ログ失敗／版競合／結果不明・医院／患者scope・保存履歴／owner独自保護を各行の実操作へ対応付ける。現在の合意範囲AC71件を全タブACの充足や実機PASSと数えない。型・UI配置・API対応と既存実装差分は技術残件で、未記載だけを理由に業務判断の再承認を求めない。

<a id="medical-record-integration-contracts"></a>
## 追加3項目の技術契約と実装差分

[合意3FR／12AC](../../../spec/domains/medical-records/pages/medical-record-form.md#medical-record-agreement-integration-three)に対応するdocs設計。以下のpayloadは論理契約であり、既存API／DBの第二正本や追加済みendpointではない。数値・職種・公開範囲を推測して追加しない。

| 境界 | 選定する技術対応 | 失敗・競合・現在との差分 |
|:---|:---|:---|
| 画像2入口 | file listを共通検証へ渡し、形式・10件／10MiB／50MiB・対象record／医院／患者・現在grant・期待版／owner保護を確認。親finalizedだけの通常入力拒否は置換。uploadの成功file IDと失敗／未送信を区別し必要ログを保存 | [現画像画面](../../../../frontend/src/features/medical-records/components/MedicalRecordImage.tsx)と[共通検証](../../../../frontend/src/features/medical-records/lib/image-upload-files.ts)は既に両入口から同じ検証へ接続する静的根拠がある。実機適合とは別。容量超過は選択全体を拒否、upload途中失敗は成功済みを消さず再送照合。serverの内容検証と保存前の再認可を省略しない |
| 印刷selection | 対象record ID・医院／患者・選択section ID（header／9タブ／vitals）・選択した画像ID・保存済み各owner版・訂正追記参照を印刷用read modelへ集約。プレビューと出力は同じsnapshotを使う | 現[print操作](../../../../frontend/src/features/medical-records/hooks/use-medical-record-form-modals.ts)と[固定帳票](../../../../frontend/src/features/medical-records/components/MedicalRecordPrintView.tsx)は今回の範囲選択／訂正追記識別に未追従。取得失敗・scope不一致は出力を止め、未保存draftは未反映を表示。cancelは入力／選択保持、印刷だけでsave／finalizeなし |
| 印刷訂正識別 | 選択対象に関係する元記載・訂正／追記の関連ID、種別、実診療日／登録日時、記録者、訂正理由・変更関係を表示。内容の版と出力時点を識別し、元と訂正後を無標識で混在させない | 対象選択が訂正の存在を隠す入口にならないよう該当追記／注記を含める。範囲外・閲覧不可ownerを無断取得／出力しない。新しい配信／飼主公開・印刷専用grant IDは設計しない |
| 結果共通参照 | 検査ownerの結果全体identity＝医院＋検査ID、固定版参照＝検査ID＋revision／version、結果項目ID＝版内参照を分離し、患者／医院と変更履歴を参照。通常表示は同じ最新結果と履歴、旧確定時を含む以前の保存時参照版と後日結果／訂正版参照を分離 | [検査revision](../../../../backend/internal/model/examination_revision.go)・[保存済み印刷snapshot](../../../../backend/internal/medicalrecord/examination_print_snapshot.go)は現行根拠。カルテへのadapter充足は別。検査画面／カルテに別の可変結果本文を持たせない。revisionなし旧結果は版不明と識別し、架空の確定時版を生成しない。mutable結果行ID／snapshot項目IDは再保存／版間で変わるため跨版stable IDにしない。項目対応はownerの明示対応を照合し、不明なら対応付けを止める。名称／順序／ID同値だけで同じ項目と推定しない |
| 後日結果追記 | 旧finalizedでも現在版への結果反映を認め、元の保存版／参照版・追記履歴を保持する。結果ID／revision・対象record／患者／医院・理由・認証actor／確認医師・期待依存版・operation keyを対応付ける | 親finalizedだけで通常結果反映を拒否せず、検査ownerの確認／訂正を普通の子PATCH／FK付替えで迂回しない。現在版と結果反映が競合したら再取得し、旧確認で強制commitしない。旧状態のunfinalizeや履歴消去で代用しない。検査自身の確定解除手順を迂回せずownerの訂正履歴を参照。同じ結果版を再送して二重追記しない |
| 影響表示 | 変更元record／項目ID・訂正履歴ID・変更対象・理由をキーに、関係する請求／見積／在庫owner ID・現在版・影響区分・確認状態・権限付き訂正入口を返す。確認状態は照合済み影響なし／要owner確認／取得不明／権限不足を区別 | 金額や在庫増減をカルテだけで確定せず、未確認IDや差額を実行指示にしない。owner取得失敗を影響なしへ変換しない。必要な関連保護を満たした臨床訂正の保存結果と、影響表示の取得結果を分け、表示失敗だけで保存済み記録を巻き戻さない |
| owner引継ぎ | 対象ID／版・元訂正ID・理由・必要な変更関係を渡し、owner側で現在認可・状態・金額／在庫・締めを再照合。戻りはownerのreceipt／再取得で実際の処理状態を確認 | 画面遷移や案内表示を訂正成功とみなさない。ownerの権限がなければ閲覧可能な範囲で不足を表示、代替grant／自動実行なし。返金・締め後・在庫増減の内部承認はD10／12／13へ残す |

取得はrecord・医院・患者・認証actor・画面generationと結び付け、遅い応答や別医院のcacheを現在の患者へ適用しない。印刷用read modelも同じ原則で、複数ownerの版が途中更新した場合は再取得して整合したプレビューを作り直す。結果不明のwriteは同じoperation identityのreceipt照会／再送へ、結果不明のreadは未取得表示／再取得へ分ける。

<a id="medical-record-full-coverage-contract"></a>

2026-10-15の通常編集・自動保存維持・復元・終了警告／履歴の変更は[最新差分](#medical-record-autosave-revert-contract)を優先する。以下の旧finalized gateは現実装又は以前の設計参照で、親状態だけの通常編集拒否を最新目標として継承しない。以前の保存版・監査・目的別訂正と別owner独自保護は保持する。
## 全タブ・共通操作の型と操作契約

表の型はUI／domain adapterで区別すべき意味。実wireの正本は既存API定義、DBの正本はmodel／migration。現コードの型を新しい業務必須条件としない。空欄・null・値未取得／取得エラー、文字列IDと数値IDは別に扱い、暗黙の0／空文字で結果や関係を偽装しない。

| 対象・主要データ | draft／保存と認可 | 確定・異常系と受入の所在 |
|:---|:---|:---|
| 問診：主訴text・任意主訴区分ID/null・治療方針text | [現問診request](../../../../frontend/src/features/medical-records/api/inquiries.ts)のchief_complaint／chief_complaint_type_id／notesへ明示adapter。local draftとactive-tab save、medical-records edit | 空SOAP／任意区分を保持。失敗はdraft保持、成功した版だけdirty解除。複写選択・上書きcancel・template追記／置換は入力4ACへ |
| 診察／方針：身体所見text・診断区分／名称ID/null（第1／2）・診断詳細text・治療方針text・版 | [現clinical-plan request](../../../../frontend/src/features/medical-records/api/clinical-plan.ts)へ明示adapter。active-tab saveとmedical-records edit。診断名称とカテゴリ整合は既存条件 | ClinicalPlan本文と項目の4状態を混同しない。旧版競合は再取得し入力保持、旧finalizedだけの通常write禁止は置換し、以前の保存版と変更履歴を保持。未完了警告は初回AC、取消は専用ACへ |
| 治療：item type・consultation／procedure／medicine／inventory master ID・名称／単位／価格snapshot・数量・内容／memo・投与経路・保険／割引・現status／順序・逸脱理由・版 | 個別create/edit/delete、並び替えだけ既存bulk sortへ。複写は選択済み逐次saveの成功subset。現在grantとsnapshot／価格sourceを照合 | 量・単位は現有データ／既存検査へ照合、評価不能・上限拒否・現在体重再評価を保持。課金／在庫自動処理は追加しない |
| 予防接種：接種ID・患者／record ID・接種日・ワクチン・医師・次回日／schedule・lot1〜4・remarks・版 | pet単位vaccinations view/create/edit/delete。カルテ画面のネストAPIを仮設しない。次回予定はD08最新合意に従い種類別設定または手入力へ統一。旧4weeks／独立1yearは実装差として解消対象とし、具体間隔・初期値・新たな優先順位を追加しない | patient照合・保存履歴と接種ownerの保護を全入口で維持。親finalizedだけの通常編集拒否は置換。予防接種結果を今回の検査revisionへ無条件統合しない。内部条件はD08 |
| 定期健診：健診ID・record ID・checkup type・実施日／次回日・医師・結果・field ID／値／単位／所見 | 一覧checkups viewとrecordネストCRUD／field-resultsは[現routes](../../../../backend/internal/medicalrecord/routes_records.go)のmedical-records操作grant。既存checkupId編集を再利用 | 別患者／旧recordの結果を反映しない。field replacementの全体保存境界を問診saveと混同しない。健診内部条件はD09 |
| 検査：検査ID・患者／record ID・exam type・医師・実施日・summary・machine・status・項目ID／結果値／単位・基準／判定・revision | examinations view/edit（新規はcreateとeditの両方）、専用unconfirmは検査ownerのみ。D07最新合意で担当医は未選択でも保存可、入力者ログと別属性に保持する。既存結果の取込・現在版への反映と以前の保存版／追記履歴を分離 | 結果の同一参照・履歴・確定時版は追加3AC。HIGH/LOWは既存検査owner導出、欠落や通信失敗を正常へ置換しない |
| 画像／資料：image ID・record ID・検証済みfile・名称／形式／size・image type・撮影日・exam ref・staff・順序・保存結果 | medical-records view／upload create／delete。ファイル選択とD&Dを共通検証・個別uploadへ。uploadはactive-tab main saveを待たない | 現server検証・保存版／owner保護・成功subset照合を保持し、親finalizedだけの通常編集拒否は置換。外部URLの無制限保存や新形式なし。画像印刷は追加3AC |
| 見積：estimate ID・record ID・title・comment／notes・itemsの名称／単価／数量／割引／保険／順序snapshot・金額・確認可能な版 | estimates view/create/edit、既存items同一tx保存／置換。別owner保存で全タブ一括txへ変更しない | 権限・医院／患者／record対応・金額整合と保存失敗を保持。依存ある空カルテを非表示しない。請求へ移す内部はD11／10 |
| 会計医師確認：confirmation ID＋record ID・pending／confirmed／returned・確認依存版・確認／差戻しactor／時刻・理由／memo（accounting IDは別owner参照） | accounting view/editの既存境界で取得／確認。カルテ編集grantで決済を代行しない。確認結果receiptは会計ownerへ渡し、通常編集・診療終了の医師確定条件へ流用しない | 未確認・差戻し・取得中／失敗／不明は会計ownerの確認条件で扱う。カルテ確定ボタンは廃止。終了時の未完了警告は別で、支払済みから終了・記録確定を推定しない |
| バイタル（タブ外）：vital ID・record／患者ID・recorded at・staff ID・体温／心拍／呼吸／体重・kg／g・notes・版／入力日時 | モーダル個別POSTはmedical-records edit、update edit／delete delete。全履歴取得状態と選択record／患者一致を確認 | 取得失敗・不完全履歴を測定なしにしない。単位／日付・現在薬量評価への反映を照合し、以前の保存版・必須ログを保持し、親finalizedだけの通常編集拒否は置換 |
| ヘッダー／次回予定：担当医ID・来院種別・実診療日、tentative予定日 | ヘッダー即時PATCHと予約owner PUT/DELETEは保存§2のまま。元record医院を使用。患者ID変更は専用転記へ | appointment付のdate／再紐付け拒否は既存参照、合意済み診療前訂正は専用確認へ。医師／編集者の混同なし。予約69traceの外へ新遷移を追加しない |
| 一覧／共通：record ID・患者／医院・条件・頁・状態・非表示履歴 | 認可済みview取得、filter／pageとcache対象を区別。行を開くだけで作成／受付／startしない。空誤作成復元は専用intent | 一覧のshortcut・患者切替・権限失効・結果不明を停止／再照合へ。非表示＝物理削除／個人表示設定ではない。元ID・履歴・依存を保持 |

[子routesの静的根拠](../../../../backend/internal/medicalrecord/routes_records.go)と[見積／会計routes](../../../../backend/internal/billing/routes.go)を照合した。UI名称とwire key（問診の治療方針→notes、ClinicalPlanの診断詳細／方針等）を意図的にadapterで対応付け、既存saveの送信先を変更しない。API YAML／全DTO・旧データ・認可実データ／全writer実装の適合は別途検証で、ここで型の一律必須化やschema変更を認めない。

## 共通状態・異常系の具体契約

読込は未取得／取得中／取得済み（emptyを含む）／取得失敗／不完全に分離。編集は保存済み版＋local draft、送信中、成功、拒否、版競合、結果不明を区別する。scopeや認証が変われば旧read／pending write結果の適用を止め、新対象を再取得する。writeには対象ID／医院／患者・期待版／依存版・認証actorを保持し、許可される操作だけを送信する。

取消は送信前ならlocal変更取消／確認modal閉鎖でserver変更なし。送信後の通信断は取消完了と断定せずreceipt照会する。既存保存が成功した対象だけdirtyを解除し、問診save成功で別タブdraft／子save失敗を消さない。複写／uploadの部分成功は成功ID・失敗・未送信・結果不明を表示し、成功対象の重複作成を避ける。version conflictは他者変更を上書きせず最新版と自分の入力を比較して明示再確認する。

診療終了は予約・record link・最新プラン状態を検査し、未完了警告確認をrecord／依存版へ結び付ける。通常編集の全writerは以前の保存版／参照を保持し、後日結果のowner条件・患者取り違え・実施取消の目的別訂正を維持する。親finalizedのみを通常入力拒否に使わない。必須編集ログとwriteは同時成功、actor／操作／対象／前後・理由／時刻・医師確認／operation結果を必要範囲で保持する。保存失敗・ログ失敗・取得不明を成功／無影響にしない。

## 文書判定範囲と関連領域への引継ぎ

今回整理したのは通常カルテの既存画面・保存§2・これまでの16FR、画像印刷／結果参照／影響表示、専用訂正／過去入力／復元、型の意味・異常系・ACのdocs契約。新しい追加業務の質問は不要。この段落は最終照合前の残件記録。現wire／DTO差分・旧値／版／認可の未知への停止・writer／consumerと受け手contractは最終照合A〜Fへ文書化し、独立レビューで今回文書範囲の完成を判定する。実権限設定・実旧データ・全writer実装適合は利用開始前検証へ分離する。runtimeの未実施を文書完了の根拠に使わない。

| 引継ぎ先 | 本文書で確定した境界 | 相手側／全体に残すもの |
|:---|:---|:---|
| D01／D02・X01／02 | 選択日受付・同ID／準備後開始、項目取消と予約Undoを分離、過去入力／復元から開始なし | 全writerと専用intent／通常auto-create差分の最終再照合 |
| D07・X05 | 同じ検査結果／履歴、旧確定時を含む保存版保持、後日結果と履歴 | 機器受信・検査確定／訂正内部、revisionなし旧結果のadapterと全取得経路 |
| D08／09・X05 | 既存専用API／患者対応・確定原記録保護 | 接種／健診内部条件の領域レビュー。検査結果合意を無条件拡張しない |
| D10／11／12／13・X06／07 | 影響表示／owner案内、自動確定請求訂正・返金・在庫変更なし | 未納／分割／0円、返金・締め後・在庫増減、見積／金額の内部承認とreceipt |
| D05／06 | 通常4状態・自動作成を専用記録／入院snapshotへ自動適用しない | 入院・ホテル・トリミング内部と退院／施術会計の別契約 |
| D20／21・C04／05 | 印刷を認可範囲内に限定、飼主自動公開なし | LIFF／レポート公開、認可設定／保存期間・削除手続きの別承認 |

医師による薬量の具体値・対象・単位・出典#261は未承認工程として保持。実装／API YAML／DB／設定・実データ・実機・並行性／性能は未変更／未検証。ここは最終照合前の進行中記録で、現在の文書範囲はGで判定。別領域への引継ぎを新規許可やD03実装完了と数えない。

### 静的参照と未追従項目の補足

[現治療request](../../../../backend/internal/medicalrecord/treatment_request.go)、[現接種request](../../../../backend/internal/medicalrecord/vaccination_request.go)、[現健診request](../../../../backend/internal/medicalrecord/checkup_request.go)、[現検査request](../../../../backend/internal/medicalrecord/examination_request.go)、[現バイタルrequest](../../../../backend/internal/medicalrecord/vital_request.go)を参照した。既存治療明細のpending／completed／not_applicableは今回のプラン4状態へ機械的変換しない。fieldの存在と今回の業務必須条件は別で、全fieldの実wire／UIの意味・互換差分は最終照合A〜Cへ記録。adapter実装／生成は未実施。

[現カルテ検査取得](../../../../frontend/src/features/medical-records/api/get-record-examinations.ts)はrevision対応が未確認の表示型で、[現addendum request](../../../../backend/internal/medicalrecord/medical_record_addendum_request.go)はafter_text／reasonが中心。今回の追記対象・元版・結果revision・医師確認・前後関係の論理契約へ追従済みではない。問診／健診等の現DTOには期待版がなく、必須ログも既存best-effortのままでは満たさない。新しいendpoint／schemaを本文で実装済みにせず、旧データ・認可実設定と合わせて差分照合に残す。

型照合の補足：[結果revisionモデル](../../../../backend/internal/model/examination_revision.go)の版内項目IDは跨版の同一性を保証しない。[医師確認response](../../../../backend/internal/billing/billing_confirmation_response.go)と[確認モデル](../../../../backend/internal/model/billing_confirmation.go)の確認ID／record IDを会計本体IDと分離する。治療の現master参照4種と検査ownerの結果参照は別の型として扱う。

<a id="medical-record-reconciled-contract"></a>

2026-10-15の通常編集・自動保存維持・復元・終了警告／履歴の変更は[最新差分](#medical-record-autosave-revert-contract)を優先する。以下の旧finalized gateは現実装又は以前の設計参照で、親状態だけの通常編集拒否を最新目標として継承しない。以前の保存版・監査・目的別訂正と別owner独自保護は保持する。
## 最終照合：現API・目標契約・互換性の区別

2026-10-14の追加docs更新承認 Sentinel_7e6b0beed1b48191ae86bc2332800109 に基づく。通常カルテ・診療masterと受け手の基幹request／response48ファイル、および補助の体重投影／nested summary response、API YAML、route／service／repository、UI adapterを静的に照合した。48は基幹pairの集計で、補助型を含む全DTO数ではない。現APIの記述は現状の根拠であって、新業務の承認根拠ではない。以下は16FR／71ACへ対応する技術契約。API YAML・DB・コード・設定は変更していない。

### A. 実キー・型・未送信／null／クリア

現契約と目標adapterを別欄にする。新しい臨床必須欄を増やさず、従来の関係整合・権限・保存／確定保護を維持する。IDは医院とownerを含む参照として区別し、JS安全整数外や無効な文字列を0へ正規化しない。金額の整数と数量の小数を取り違えない。

| 対象 | 現request／responseの具体的対応 | 目標adapter・未送信／null／クリア |
|:---|:---|:---|
| 本体create | owner_id／pet_id／doctor_id／appointment_idは文字列ID入力。本体responseは数値ID（optionalキー省略あり）。dateは日時、visit_dateは日付文字列。statusはdraft／finalizedを受理する現契約 | createの旧受理範囲を明示確定・過去入力の新許可にしない。受付participant／専用過去intentを用途で分離。文字列ID⇔数値IDを検証して変換し、実診療日と現在登録日時を別に保持 |
| 本体update | owner_id／pet_id／doctor_id／appointment_idは数値pointer、dateは日時、status／visit_typeはstring pointer、versionは任意int | FE型のnumber／string unionをBEに無条件送信しない。未送信は保持、pointerへのJSON nullは保持と区別できないため解除に流用しない。患者取り違えはpatient PATCHへ流さない |
| 作成時診断1 | FE CreateMedicalRecordRequestはdiagnosis_1_type_id、BE create requestはdiagnosis_1_category_id | 旧キー差を明示し、対応が未照合なら診断初期値を保存成功としない。主訴／診断を専用保存へ渡す既存§2を維持し、古いcreate型だけで全臨床情報を保存しない |
| 問診 | chief_complaint／notesはstring pointer。治療方針→notes。chief_complaint_type_idはnullable fieldで存在有無を保持。詳細responseと本体要約は異なる | 未送信は既存保持、主訴区分nullは明示解除、数値は指定。textの空文字と未送信を分離。SOAP空欄可／区分任意を維持。response省略と値不明を未設定確定にしない |
| 診察／診断 | physical_exam／diagnosis_details／treatment_policy。診断1 diagnosis_type_id／diagnosis_name_idはuint64 pointer、診断2 diagnosis_2_type_id／diagnosis_2_name_idはnullable field。response IDはstring | 診断1のnullは現行では未送信同様、診断2のnullは解除。名称／カテゴリの整合を照合し、全IDが同じnull動作と書かない。旧APIが解除を表現できない欄は解除済みと表示しない |
| 治療／プラン | consultation／procedure／medicine／inventory ID、unit_price int64、quantity float64、is_selected／is_insurance、内容／memo／投与経路／割引／sort／dose理由。更新versionは任意。TreatmentPlanは内容／価格／数量／割引等で完了状態なし | 未送信・0・false・空文字を区別。旧治療status3値／is_selectedを新プラン4状態に変換しない。quantityを整数へ丸めない。inventory ID指定の通常createはquantity整数必須で小数を拒否する現条件を保持し、整数丸めで救済しない。master ID／revision・当時内容の欠損はEの照合へ |
| バイタル | recorded_at日時、staff_id、体温／心拍／呼吸／体重／kg・g／notes。各optional値、更新version。現createの監査actorはinput.StaffID、deleteはnil、PATCHだけ認証actorを分離。staff_idは測定記録側の属性 | 未送信を0測定にしない。目標は全CRUDで認証編集actorを必須とし測定staffで代用しない、欠落時停止。実測日時と現在編集日時・測定staffと編集者を分離。単位未取得の値をkgと決めず、記録日時・患者・医院を確認して薬量へ接続 |
| 接種 | 独立vaccinations：患者／record ID、vaccine、日付、doctor、next_date／schedule、supplemental、lot1〜4、remarks、更新version | pointer nullは解除と推定しない。既存日付／次回予定・医師／患者の条件はD08へ。接種を検査revisionへ統合せず、同じ親の保存版・履歴保護を全入口で維持し、親finalizedだけの通常編集拒否は置換 |
| 健診／field結果 | checkup_type／患者／日付／次回日／doctor／result、doctor_id_clearあり、versionなし。field結果はvalue_number／text／bool／listとfield ID。PUT results未送信／nullは拒否、明示[]だけ全消去、配列置換 | doctor解除はclear flagを使用。空の次回日文字列は現実装で解除として表現できるとは限らず、明示解除の充足を推定しない。型に合う値と明示falseを保持し、置換の操作単位を問診saveと混同しない |
| 検査 | clinic／exam／患者／record、exam_type、doctor、日時、summary／machine／status、items。response current_revision_version。項目はexam_type_field_id、reference_value、qualitative_min／max等。PATCH items未送信／nullは不変更、[]は空置換。現PUT /itemsは未送信／nullも空置換へ変換 | 結果全体・固定版・版内項目参照を分離。目標はitems未送信／nullを無意図の全消去へ変換せず、意図的全消去は明示[]を要求する。現旧PUTは切替対応前に充足扱いしない。関連pointer nullを区別。カルテExamGroupが現在落とすrevisionは新read modelで保持。通常PATCHを確定カルテ追記の代用にしない |
| 見積／会計確認 | 見積itemsは名称／単価／数量／割引／保険／順序、clear_valid_until。確認responseのidはconfirmation ID、medical_record_idとpending／confirmed／returned・actor／時刻／理由／memo | confirmation IDをaccounting IDと混同しない。見積の明示clearと未送信を分離し既存items保存単位を維持。確認依存版は目標contractであり現responseに存在すると推定しない |
| 画像／追記 | multipart file→保存済み画像DTO（ID／名称／size／mime／URL／撮影日／関連exam／staff／順序等）。追記requestはafter_text／reason、response before_textは現serviceで空文字 | 画像メタデータの表示fieldを書込可能fieldとしない。追記前後・対象／元版・医師確認・結果revision・転記先関係は専用論理contractへ追加し、現after_textだけで構造化対応済みとしない |
| 診療master | consultation／procedure／medicine／diagnosis／interview／chief-complaintの現CRUDはmaster-medical grant。親／strength等にはclear flagがあり、dose param PUTは未送信optionalをNULLへ置換。共通revisionは現DTOにない | APIごとのclear／PATCH／PUT置換を明示。master編集者と薬量値承認医師を分離。現在有効な新規選択と既存snapshotを分け、dose param再設定を医師の具体値承認済みと数えない |

根拠：[本体request](../../../../backend/internal/medicalrecord/medical_record_request.go)・[本体response](../../../../backend/internal/medicalrecord/medical_record_response.go)・[FE作成更新型](../../../../frontend/src/features/medical-records/api/types.ts)、[問診request](../../../../backend/internal/medicalrecord/inquiry_request.go)・[診察request](../../../../backend/internal/medicalrecord/clinical_plan_request.go)、[健診request](../../../../backend/internal/medicalrecord/checkup_request.go)・[検査response](../../../../backend/internal/medicalrecord/examination_response.go)。各タブの現sourceリンクは前節の表を維持する。

### B. API YAML・取得元の既存差分

API YAMLは[現契約参照](../../../../backend/docs/api.yaml)のまま保持し、現DTOとの差をここに記録する。以下の修正対象はdocs内のadapter／互換説明で、backend/docs/api.yamlの変更・再生成を今回実施しない。

| 差分 | 現状 | 互換・目標への扱い |
|:---|:---|:---|
| 問診要約 | 実DTO InquirySummaryResponseはchief_complaint_type_idを含むがYAML要約schemaにない | 欄がないschemaで主訴区分を消さず実responseを照合。キー欠落と任意nullを保持 |
| 検査項目 | YAMLのexam_type_item_id／refと、実responseのexam_type_field_id／reference_valueが異なる。qualitative_min／max等にも差がある | 明示adapterで実キーを扱い、名称似だけでmaster ID／基準を付け替えない。旧応答の対応不明は未取得として照合 |
| 治療型 | YAMLにはselected／insurance等の旧名とis_selected／is_insurance、数量integer・単価number等の旧記載がある。実requestはquantity float64／unit_price int64（inventory ID指定の通常createだけquantity整数を必須検査） | actual wireを検証し型・旧名を黙って別の状態／丸め値へ変換しない。旧client互換を切替条件に含める |
| 本体作成更新schema | 現YAMLが汎用MedicalRecord参照で、実create／updateのID型・visit_date・初期臨床field等の差を十分表さない | GETモデルをそのまま書込payloadにしない。Aの操作別adapterを正本へ接続し、code対応だけを業務承認としない |
| 本体GET | 現MedicalRecordResponseは問診要約・飼主／患者／担当医・体重投影等。Treatments／全Vitalsを同梱しない | 治療／全バイタル・ClinicalPlan・画像・検査・見積／確認を各ownerから取得。体重投影欠落＝旧API、空配列＝測定なしを分離 |
| 目標の新contract | 今回選定した専用訂正／過去入力／復元、警告receipt／operation key、選択範囲印刷のendpoint・fieldは現route／schemaに揃っていない | 新URLを実装済みと記載せず、次節の論理input／outputと停止条件を適用。対応前の旧CRUD fallbackを禁止 |

### C. 互換切替・再送・医師確認

新契約の利用開始条件は、医院／対象操作ごとに全consumer／writer・認可・旧値・版・必要ログ・副作用の対応を確認し、予約の[共通gate](../reservations/technical-spec.md#reservation-global-lock-contract)下で旧writerを排出できること。切替前は現コードを新仕様適合済みと扱わず、未対応の専用操作を利用不可として表示する。旧clientのversion省略／旧status・旧分類・無型payloadを、新contract下で無条件受理しない。旧データの通常読取を必要以上に破壊せず、読み取れない項目は不明として保持する。

論理mutation inputはpurpose／医院／患者／recordと子ID・期待版／依存版・operation key・変更内容／理由・必要な医師確認receipt。actorは認証由来、確認医師は医院が認めた医師で、予約指名医師・カルテ担当医・clientのdoctor_id・スタッフ所属だけから推定しない。current grantの判定と医院が認める医師の判定を分離し、新しい職種／grantを作らない。医師判定の実設定が未確認なら確認が必要な操作は停止し、担当医以外の通常入力を一律禁止しない。

receiptは医院／actor／purpose／対象／正規化変更内容／理由へbindingし、同じkeyで内容が違えば拒否。同じkeyの既存結果は同じ対象の結果として照会し、新keyの再作成で代用しない。期待版・依存版は保存直前に照合し、確認receiptがあっても最新版検査を省略しない。子変更・master更新・会計確認差戻し等で依存が変われば確認失効、再取得・再表示へ戻す。未完了警告は確認可能な警告、取得不能は再取得が必要な不明状態として分離する。

論理outputはoperation identity、成功対象ID／現在版・保存結果、必要監査結果、医師確認／追記／元先対応ref、owner影響状態。通信断の結果不明は成功／失敗と断定せず、同じoperation receiptを照会する。receipt対応のない旧APIを新専用intentの再送へ流用せず、再取得して関連を照合し、証明できなければ新規writeを止める。逐次複写／uploadの成功subsetは残し、既知成功分を再送しない。共通ACのscope-generation／unknown-receipt／versions-no-overwriteへ対応する。

技術方式の選定と、実権限・実旧データ・実装／schema・実機性能の充足は別である。実装前に実権限と医師判定、医院別の旧client／旧値対応を確認する。薬量具体値の医師承認#261は別承認で、医師確認receiptの存在を値・対象・単位・出典の承認証拠にしない。

<a id="medical-record-writer-matrix"></a>

2026-10-15の通常編集・自動保存維持・復元・終了警告／履歴の変更は[最新差分](#medical-record-autosave-revert-contract)を優先する。以下の旧finalized gateは現実装又は以前の設計参照で、親状態だけの通常編集拒否を最新目標として継承しない。以前の保存版・監査・目的別訂正と別owner独自保護は保持する。
### D. 全writer・監査・gateの参加表

HTTP methodだけでwriteを分類しない。現GETの初期行作成、内部participant・機器取込・並替・関連移動・master更新・副作用も候補へ含める。表は静的照合で特定した経路群と新contractへの参加規約であり、全実行経路の新gate適合PASSを示さない。

| 経路群・現source | 現状の区別 | 新contractへの参加・失敗条件／AC |
|:---|:---|:---|
| 本体create／update／delete／推奨理由：[CRUD](../../../../backend/internal/medicalrecord/medical_record_crud.go)・[update tx](../../../../backend/internal/medicalrecord/medical_record_crud_update.go) | 通常create／update／delete監査はbest-effort、delete actorは現経路でnil。finalizeは同一tx監査。予約リンク検査と親lockはある | 全成功編集の認証actor・対象・前後／理由を必要ログへ同時commit。欠落は失敗。予約gate／親／子依存版・空判定・医師確認へ。coverage-child-audit-failure／versions-no-overwrite |
| 通常auto-create／backfill：[FE auto-create](../../../../frontend/src/features/medical-records/hooks/use-medical-record-auto-create.ts)・[内部auto-create](../../../../backend/internal/medicalrecord/medical_record_auto_create.go) | 現画面表示で予約／診療中・recordを作る経路と、新しい受付時作成／準備後開始は異なる | 全入口を受付participant／同ID準備へ接続する利用開始条件。通常一覧／印刷から作成しない。過去intentへ流用しない。no-reception-create／coverage-special-record-separation |
| 問診upsert：[service](../../../../backend/internal/medicalrecord/inquiry_service.go)・[repo](../../../../backend/internal/medicalrecord/inquiry_repository.go) | 親draft lock・未送信保持あり。現requestに期待版／認証編集者の保存契約がなく、slogは必要履歴の代用でない | 専用saveの送信先を維持し依存版／必要ログへ参加。未送信／nullをAで区別。coverage-active-tab-only／child-audit-failure |
| ClinicalPlan GET初期化／update／delete：[service](../../../../backend/internal/medicalrecord/clinical_plan_service.go) | GET GetOrCreateは行作成。update／deleteは親lock・tx監査、versionは任意 | 新readは副作用なしprojectionを選定し、初期作成は認可されたwriteのupsertへ分離する。旧GET初期化も残存writerとしてgate／親依存の照合対象とする。既存GET挙動を未対応の印刷readへ使わない。coverage-finalized-all-writers／child-independent |
| 治療CRUD／sort：[service](../../../../backend/internal/medicalrecord/treatment_service.go)・[tx](../../../../backend/internal/medicalrecord/treatment_service_tx.go) | 親draft lock・更新版あり。薬量逸脱時のtx監査と、通常CRUD全編集の必要監査は別。createのinventory ID指定時はquantity整数必須検査と減算 | 全write／sortを必要ログ・親子版へ参加。専用転記／過去入力では通常createの在庫副作用を発火させるadapterを使わない。copy-partial-failure／coverage-child-independent／impact-financial-stock-no-auto |
| TreatmentPlan CRUD：[service](../../../../backend/internal/medicalrecord/treatment_plan_service.go) | 通常recordと入院関連を持ち、価格／数量保存。新4状態／完了ログのfieldとは別 | 通常item refと4状態の保存adapterを明示。入院snapshotへ自動波及しない。全price／sort／deleteも必要ログへ。plan-recorder-not-performer／coverage-special-record-separation |
| バイタルCRUD：[service](../../../../backend/internal/medicalrecord/vital_service.go) | 親draft lockとtx監査。POSTはedit grant、create監査は入力StaffID、delete監査はnil、PATCHだけ認証actorを分離 | 全CRUDで認証actor・必要編集ログを同時保存し、測定staffで代用しない。actor欠落は停止。読込完全性・単位／日付・現在患者／医院・版・薬量依存へ参加。coverage-vital-read-state／child-audit-failure |
| 接種CRUD／関連変更：[service](../../../../backend/internal/medicalrecord/vaccination_service.go) | 親／子lock・旧新関係再照合、削除tx監査あり。作成／更新全履歴の充足とは別 | 全CRUDと旧／新record関係をgate／必要ログへ。相手ownerとの内部境界維持。coverage-finalized-all-writers／scope-generation |
| 健診CRUD／field配列置換：[service](../../../../backend/internal/medicalrecord/checkup_service.go)・[field service](../../../../backend/internal/medicalrecord/checkup_field_result_service.go) | 親draft保護。現健診versionなし、field結果削除ログは全置換成功の前後履歴とは別 | 置換単位の前後・actor・依存版を記録。doctor clear／値型をAへ。coverage-child-audit-failure／versions-no-overwrite |
| 検査CRUD／項目置換／unconfirm／revision：[service](../../../../backend/internal/medicalrecord/examination_service.go)・[revision](../../../../backend/internal/medicalrecord/examination_revision_service.go) | 検査ownerの独自確定／revisionと、カルテfinalized保護がある。跨版item IDは再生成 | 検査側確認手順を保持して同じ結果／版・履歴を返す。旧finalizedカルテへの後日結果も保存済み参照版を保持して現在版へ反映し、検査ownerの独自確定／訂正は維持。late-result-finalized／revision-history／unknown-scope |
| 機器受信／attach／取込／revert：[取込](../../../../backend/internal/medicalrecord/lab_import_examination_service.go)・[revert](../../../../backend/internal/medicalrecord/lab_import_revert_service.go) | ambient参加や検査usage／確定保護を持つ別経路。通常UI PATCHだけの棚卸しでは落ちる | 当該ownerのreceipt／使用履歴と親依存を照合し、手入力・通常取消で迂回しない。D07内部手順は引継ぎ。coverage-finalized-all-writers／owner-read-and-handoff |
| 画像upload／delete／object削除：[service](../../../../backend/internal/medicalrecord/medical_record_image_service.go) | ファイル検証・親draft保護、DB保存と外部object処理がある。全操作の必要編集ログとは別 | upload成功ID／失敗／不明を保持し二重uploadを防ぐ。DB状態とobject副作用・必要監査を整合。形式／容量全選択拒否と通信部分成功を分離。image-drop-reject-batch／coverage-unknown-receipt |
| 追記create：[service](../../../../backend/internal/medicalrecord/medical_record_addendum_service.go) | finalizedだけに追記、同一tx監査。現after_text／reasonは構造化訂正refを持たない | 元／先・元版・子target・結果revision・医師確認を専用履歴へ結ぶ。目的別訂正を汎用更新で迂回しない。通常本文更新は最新の履歴付き通常編集へ。execution-undo-finalized／wrong-patient-target-guard |
| 会計確認GET初期化／confirm／return：[service](../../../../backend/internal/billing/billing_confirmation_service.go) | GetOrCreateはpending行作成、confirm／returnは親draft lockと有効スタッフ所属検査。新医師確認／依存receiptとは別 | 初期化もwriter台帳へ。印刷はread-only projection、confirmは既存accounting grant＋医院が認めた医師確認・現在依存・必要ログへ。returnは既存操作認可・actor・現在依存／確認失効・必要ログを照合し、差戻し主体の新しい医師限定を推定しない。physician-confirmation／coverage-finalize-concurrent-deps |
| 見積CRUD／successor・明細：[見積service](../../../../backend/internal/billing/estimate_service.go)・[successor](../../../../backend/internal/billing/estimate_service_successor.go) | 独立ownerの保存／tx・親依存がある | 見積依存・版・金額／items境界を返し空削除を保護。自動返金／請求生成へ拡張しない。coverage-owner-read-and-handoff／empty-hide-entered-refusal |
| 会計完了／訂正／返金・在庫：[会計訂正](../../../../backend/internal/billing/accounting_service_correction.go)・[在庫service](../../../../backend/internal/inventory/inventory_service.go) | 各ownerの実処理と、カルテの影響表示は別。通常治療create減算の既存副作用あり | 影響ref／現在版・owner receiptを確認。臨床訂正成功や画面遷移でowner処理済みとしない。返金／締め後／増減ルールはD10／12／13へ。impact-unknown／handoff-permission |
| master CRUD／reorder／dose param／参照変更：[routes](../../../../backend/internal/medicalrecord/routes_masters.go) | 既存master-medical grantと個別clear／PUT置換。共通revision／当時名単位snapshotは一律存在しない | 有効性・価格／名称／単位・確認依存の変更をgate／必要ログへ。旧参照保持と新規選択を区別。master-inactive-history／coverage-snapshot-null |

参加規約は予約の共通gate／下位取得順を継承する。複数recordへの専用転記では、該当予約→recordを既存の順序規約で決定的に取得し、元／先両方と関連ownerの依存を再照合する。ambient participantは独立commitせず、必要write／ログを同じ外側transactionへ参加させる。外部objectやLINE等の副作用はcommit結果・操作identityへ対応付け、臨床保存成功と外部処理成功を混同しない。旧createの初診welcome／tag同期も現副作用として照合対象とする。承認済み目標は[D15の確認adapter](../../../spec/line/architecture.md#followup-confirmation-adapter)へ接続し、記録作成だけでは送らず、スタッフの実施済み・対象確認後に一般案内と医師要フォロー指定が必要な医療案内を分ける。確認保存と必要監査は既存の保存結果・認可／依存照合へ参加し、外部作用はcommit後の別結果とする。過去／転記／復元から新しい配信許可を推定せず、通知失敗でカルテ再作成や結果不明のblind retryをしない。追跡はD15のAC-crm-followup-record-only／medical／confirmation-failure／duplicate-unknown／partial-recordで、D03固有AC件数へ重複加算しない。配信内部はD15へ引継ぐ。

<a id="medical-record-legacy-and-special-contract"></a>

2026-10-15の通常編集・自動保存維持・復元・終了警告／履歴の変更は[最新差分](#medical-record-autosave-revert-contract)を優先する。以下の旧finalized gateは現実装又は以前の設計参照で、親状態だけの通常編集拒否を最新目標として継承しない。以前の保存版・監査・目的別訂正と別owner独自保護は保持する。
### E. 空判定・旧値・専用intentの具体契約

| 判定対象 | 選定する条件 | 不明／旧データの扱い |
|:---|:---|:---|
| empty／entered／unknown | 問診・診察／診断／方針・治療／処方／プラン・バイタル・検査・接種・健診field・画像／資料・追記を確認。臨床入力があればentered、自動ID／患者／初期日付だけは臨床入力としない | child未取得／不完全・原始情報不明はunknown。初期templateらしい文字列を名前だけで空にしない。由来を確認できない旧値を捨てない |
| 非表示可能な空誤作成 | empty＋削除を妨げる請求／見積／臨床子依存なし＋draft＋診療中リンク等の既存保護充足、現在認可・医院／患者・版を同じlock下で照合 | 現deleteのstatus／見積保護だけを充足証拠にしない。unknown／entered／finalizedは停止し訂正／追記へ。受付リンクだけで空／削除可とはしない |
| 復元 | tombstone元ID／履歴・理由・actor、現在患者／医院・依存版／一意性・復元操作認可 | 医師確認を新必須にせず権限staffで実行。競合active record・医院異動／対応不明は停止、別IDcreateなし、開始／受付／課金／確定なし |
| master旧snapshot | 元master IDと保存済み当時名／単位／価格・由来／元版を保持。新規選択・通常複写は現在有効master／今回価格へ | 実treatmentは価格・contentやdose snapshotを持つが共通master revision／当時単位が一律ない。現masterから当時値を推測補完しない。読取は不明を表示、旧値依存の変更は資料確認まで停止 |
| 旧プラン状態 | 新4状態は明示item refに保存。旧TreatmentPlan本文・treatments3状態・支払済み／終了状態から推定しない | 旧状態不明は旧情報として表示し照合を要する。全実施済み／中止へ補完しない。未完了警告の確認と、取得不能の停止を分離 |
| 薬量評価不能 | 不足情報による評価不能は医師確認・理由付き手入力、評価可能絶対上限超過は拒否。関連取得失敗／species不一致の既存拒否を維持 | 現dose evaluatorのmissing-data→手動という互換を、確認・理由・必要ログ済みとしない。欠損と通信障害を分離。具体値／出典の医師承認未実施 |

以下は既存URLへ無条件にadapterを当てる方式ではなく、論理intentのinput／output・停止／副作用契約。共通認可・医院／患者・版・必須編集ログ・再送receiptはCに従う。

| intent | 入力と返却 | 成功条件／停止・副作用 |
|:---|:---|:---|
| 通常準備／受付 | appointment ref／版→同じrecord ref・状態／内容判定／依存版・保存結果 | 受付と通常record同時成功、準備／読込成功後の明示start。新規画面表示だけで作成／開始しない |
| 項目実施取消 | record／item ref、期待状態／版、戻し先、理由、actor→前後状態・履歴／追記ref・owner影響状態 | 旧状態不明は停止。以前の保存版を破壊せず現在版へ理由付き取消、請求影響はowner訂正へ。予約Undoとは別 |
| 患者取り違え | 元record／患者、正しい患者／転記先、選択必要内容・理由・確認医師／依存版→訂正履歴・元先対応・転記成功ref | 元削除／patient ID付替え・全件自動複写なし。元／先の以前の保存版・対応履歴を保護。親finalizedだけで通常編集を拒否しない。歴史的事実の転記と新治療注文を分離し、誤患者の薬量評価を転記先の安全判定済みにしない |
| 過去専用入力 | 患者・実診療日・理由・現在入力日時／actor・確認医師・既存候補→専用record ref／記録結果 | 通常auto-create／未紐付け移行例外を代用にしない。当日カード／開始・自動課金なし。重複候補不明は停止。必要請求は会計確認 |
| 空非表示／復元 | 元record／tombstone・理由／actor・期待版／関連→同じID・履歴・表示状態 | 上表の空／現在依存／一意性条件を再照合、開始／課金／別ID複製なし |
| 後日結果追記 | clinic＋exam／固定revision、版内item対応・target record／元参照版・理由／確認→同じ結果／履歴・追記ref | draftは許可された関連保存、finalizedは元記載／版を保持した専用追記。通常FK PATCH／カルテunfinalizeなし、跨版対応不明は停止 |
| 印刷projection | record／医院／患者・選択section／子ID・保存済みowner版・訂正ref→同じpreview／出力snapshot | GetOrCreateを使わない副作用なしreadを選定。各owner viewと現在scopeを検査、未保存draftの自動保存／飼主公開なし。取得不明／版不整合は出力停止 |
| 訂正影響／owner引継ぎ | 元訂正／record／item ref・理由／差分・owner関係／版→影響区分・要確認／不明／権限不足・owner receipt | 通常治療createのinventory減算を転記／過去入力／訂正へ発火しない。会計返金・在庫増減を自動実行しない。臨床write結果と影響read結果を分離、案内を処理成功にしない |

原記録・履歴の不変、元／先の必要内容対応、移行時のunknown停止が目標契約。実データにない元版／当時値や監査を架空生成しない。旧reader／writerの互換対応・実DB制約／trigger／外部consumerの適合は利用開始前の確認に残し、文書に未実装差分と停止条件があることを利用開始可の証拠にしない。

<a id="medical-record-final-trace"></a>

2026-10-15の通常編集・自動保存維持・復元・終了警告／履歴の変更は[最新差分](#medical-record-autosave-revert-contract)を優先する。以下の旧finalized gateは現実装又は以前の設計参照で、親状態だけの通常編集拒否を最新目標として継承しない。以前の保存版・監査・目的別訂正と別owner独自保護は保持する。
### F. 合意16FR／71ACからの最終trace

業務本文・GWTを複製せず、現契約差分／目標adapter／writer／専用intentへ辿る索引。下表の固有AC57件と共通AC14件で71件、継承FR／ACは別。検査・接種・健診／会計等owner内部の全要件確定や実装PASSを件数から推定しない。

| 合意FR・正本 | 対応する最終契約 | 固有AC（AC-medical-records-を共通prefixとする） |
|:---|:---|:---|
| [FR-medical-records-authorized-editor-audit](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-authorized-editor-audit) | A／C／D・医院／医師／編集者と必要ログ | non-attending-edit-audit, edit-no-grant, physician-confirmation, edit-audit-failure |
| [FR-medical-records-explicit-finalize-plan-warning](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-explicit-finalize-plan-warning) | 最新契約・現在依存と警告receipt、診療終了時の確認可能警告 | no-clinical-required, plan-incomplete-warning, plan-warning-cancel, pending-results-draft |
| [FR-medical-records-empty-delete-reviewed-correction](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-empty-delete-reviewed-correction) | D／E・空／関連判定と訂正 | empty-delete, entered-delete-refusal, wrong-patient-correction |
| [FR-medical-records-reception-and-retrospective-entry](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-reception-and-retrospective-entry) | C／D／E・通常受付と専用過去を分離。死亡登録済み生前分も理由・医師確認・死亡状態保持 | no-reception-create, retrospective-reason-physician |
| [FR-medical-records-unevaluable-dose-manual](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-unevaluable-dose-manual) | A／C／E・評価不能と医師確認／理由 | unevaluable-manual, unevaluable-no-confirmation, evaluable-hard-limit |
| [FR-medical-records-reasoned-plan-execution-undo](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-reasoned-plan-execution-undo) | D／E・戻し先／前後履歴・owner影響 | execution-undo-reason, execution-undo-cancel, execution-undo-finalized, execution-undo-billing |
| [FR-medical-records-wrong-patient-linked-transfer](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-wrong-patient-linked-transfer) | A／D／E・患者ID変更なし必要内容転記 | wrong-patient-transfer, wrong-patient-no-reassign, wrong-patient-target-guard |
| [FR-medical-records-retrospective-no-card-no-billing](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-retrospective-no-card-no-billing) | D／E・実診療日／現在入力日時・副作用。死亡後誤記訂正も原記録保持・専用経路 | retrospective-dates, retrospective-no-reception-billing, retrospective-existing |
| [FR-medical-records-empty-record-hide-restore](../../../spec/domains/medical-records/lifecycle.md#requirement-fr-medical-records-empty-record-hide-restore) | C／D／E・元ID・現在関連／一意性・復元 | empty-hide-history, empty-restore-authorized, empty-restore-conflict, empty-hide-entered-refusal |
| [FR-medical-records-selective-history-copy](../../../spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-selective-history-copy) | A／C／D／E・選択・今回価格・逐次成功subset | copy-select, copy-overwrite-cancel, copy-current-price, copy-invalid-master, copy-partial-failure |
| [FR-medical-records-manual-template-insertion](../../../spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-manual-template-insertion) | A／C／D・問診draft手動追記／置換確認 | template-append, template-replace-cancel, template-no-auto |
| [FR-medical-records-master-history-snapshots](../../../spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-master-history-snapshots) | A／C／D／E・有効性／当時値／不明停止 | master-inactive-history, master-scope-grant |
| [FR-medical-records-plan-execution-states](../../../spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-plan-execution-states) | A／D／E・4状態と旧状態の分離 | plan-checkbox-modal, plan-uncheck-not-undo, plan-cancel-reason, plan-deferred-warning, plan-recorder-not-performer |
| [FR-medical-records-image-drop-selected-print](../../../spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-image-drop-selected-print) | A／B／D／E・共通検証とread-only print | image-drop-same-validation, image-drop-reject-batch, print-range-corrections, print-draft-cancel-scope |
| [FR-medical-records-late-result-shared-history](../../../spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-late-result-shared-history) | A／D／E・同一exam／固定版／版内項目と追記 | late-result-draft, late-result-finalized, late-result-revision-history, late-result-unknown-scope |
| [FR-medical-records-correction-owner-impact](../../../spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-correction-owner-impact) | C／D／E・表示≠実行、receiptとowner引継ぎ | impact-financial-stock-no-auto, impact-unknown, impact-handoff-permission, impact-settled-history |

共通ACの正本は[全タブ共通条件](../../../spec/domains/medical-records/pages/medical-record-form.md#medical-record-shared-coverage-ac)。全経路へ同じGWTを適用する参加表はD、API差／互換はA〜C、専用手順はE。保存§2は変更しない。

| 共通AC suffix（AC-medical-records-coverage-） | 重点照合 |
|:---|:---|
| active-tab-only／header-immediate／child-independent／save-partial-dirty | 保存§2・個別送信先・成功対象だけdirty解除、GET本体取得元の訂正 |
| vital-read-state／snapshot-null | 未取得／実測なし・null／clear・単位／日時・当時値不明の分離 |
| finalized-all-writers／finalize-concurrent-deps | GET初期化／関連変更／機器取込を含むDの全writer、保存版保持・目的別確認と依存失効。通常編集と終了警告は最新契約へ |
| scope-generation／child-audit-failure／versions-no-overwrite | 医院／患者／認証切替と遅い応答、必須ログ同時成功、期待版省略の旧互換 |
| unknown-receipt／special-record-separation／owner-read-and-handoff | 再送結果照合・同ID保持、過去／転記／復元の副作用分離、表示≠owner処理成功 |

共通14ACは新しい業務判断ではない。APIキー差・null／clear、非対応receipt、旧状態／snapshot、GET初期化、在庫副作用を正常／失敗／取消／競合／結果不明の具体ケースへ適用して確認する。対応実装・実データ／実機・並行性／性能は未実施。

<a id="medical-record-document-completion"></a>

2026-10-15の通常編集・自動保存維持・復元・終了警告／履歴の変更は[最新差分](#medical-record-autosave-revert-contract)を優先する。以下の旧finalized gateは現実装又は以前の設計参照で、親状態だけの通常編集拒否を最新目標として継承しない。以前の保存版・監査・目的別訂正と別owner独自保護は保持する。
### G. 文書完了の判定範囲

文書完了の対象は、合意16FR／固有57AC＋共通14AC、通常カルテ全9タブ・バイタル／ヘッダー／一覧、D03診療masterの入力／履歴／認可境界、保存§2・確定／訂正／転記／過去入力／復元、画像／印刷／結果参照・影響表示、最終照合A〜Fの型／互換／停止／writer／受け手のdocs契約。この範囲の独立レビュー3名・限定文書検査を通過したため、進捗D03を限定範囲の文書完了と判定。実装／全体統合の完了ではない。

完了には現コードの新仕様追従を要求するのではなく、未追従差分・adapter／利用開始条件・実データ不明時の停止・必要監査・再送／依存版・他ownerへの引継ぎが文書で明示されていることを要求する。未実装差分や値不明を消して完了としない。今回までの承認から決まる業務に追加質問は不要。将来追加する業務があれば既決との差分を未承認質問として分離する。

| 判定の外に残すもの | 状態／引継ぎ |
|:---|:---|
| API／DB／コード・全writer／consumer・旧client切替・receipt／監査実装 | 未変更／未検証。実装着手・利用開始前に現契約との差分とfail-closedを検証 |
| 医院別実権限・認められた医師・実旧データ／当時snapshot・元版・実DB制約／trigger | 未確認。C〜Eの照合／停止を適用し、架空値・role・臨床承認を補完しない |
| 薬量具体値・対象薬／種・単位・範囲・出典の医師承認#261 | 未承認工程。一般手入力許可・医師確認／ログとは別、値の利用許可を推定しない |
| D07機器／検査内部、D08／09、D10〜13金額／返金／締め後／在庫増減、D05／06専用記録、D15配信、D20／21公開 | 各ownerの内部文書整理へ引継ぎ。今回確定した境界を内部業務の新許可にしない |
| 全体横断X01／02／05〜07の最終統合、共通C04／05／06 | 各owner文書が揃った後の再照合・共通規約へ。D03個別文書完了から全体完了へ加算しない |
| 実機・実データ・並行性／性能検証・運用開始 | 未実施。静的文書レビュー／リンク検査を実機PASSとしない |

次の推奨領域はD10会計。予約確定歴／通常Undo禁止、カルテ訂正影響、過去入力請求、会計医師確認と精算本体IDの境界を接続して、未納／分割／0円・返金／締め後の内部を整理する。D07検査・D13在庫は今回の受け手差分を保持して後続へ。実装・起票・公開は引き続き今回の対象外。

### 最終文書レビュー結果

根拠／trace、状態・副作用、現型／APIの独立レビュー3名で、未送信PATCH／PUT定義、会計return主体の医師限定誤読、バイタルcreate／delete監査actorの現状差・健診field PUTの未送信／null／[]・検査PATCHと旧PUT全消去の差・在庫連動createの整数数量・基幹DTO集計範囲を修正し再確認、未解消阻害なし。A〜Gと16FR／71ACの明示範囲は文書完成。旧wire／旧値／実認可・全writer実装適合・他owner内部・薬量具体値・実機性能の未確認を保持する。
