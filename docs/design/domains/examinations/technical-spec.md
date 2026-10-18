# D07 検査・機器受信の技術契約

| 項目 | 内容 |
|---|---|
| 文書ID | design-D07-examinations-technical |
| 種別 | 技術契約・実装差分 |
| 決定状態 | 既存記載を継承（個別の決定根拠は本文参照。根拠未確認部分は承認済みとしない）＋D03合意目標への接続。担当医未選択で保存可・入力者ログ保持の2026-10-15第3項合意を反映。項目選択印刷・感染症結果患者トップの追加採用を本文末尾へ接続。対象外の未決は保持 |
| 実装状態 | 本文の現コード観察と目標契約を分離。コードに存在するだけでは業務承認済み・目標実装済みとしない |
| 実機検証状態 | 今回未実施。既存証跡の引用は今回の実機受入を意味しない |
| 文書確認日 | 2026-10-15 |

既存仕様の継承とD03の合意済み目標を整理した文書候補。ソフトウェア仕様の整理、実装完了、実機受入、医師による具体的な臨床数値の承認は別の判定とする。

## 対象と証拠
業務正本は[検査ライフサイクル](../../../spec/domains/examinations/lifecycle.md)。2026-10-15、固定commit `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`の文書／source静的照合。API／DB／実装／実grantは変更せず、実機・並行性・性能未検証。既存実装とD03継承目標を分ける。

<a id="examination-state-contract"></a>
## 状態・版と保存単位
[lock](../../../../backend/internal/medicalrecord/examination_lock.go)、[revision](../../../../backend/internal/medicalrecord/examination_revision_service.go)、[更新](../../../../backend/internal/medicalrecord/examination_service_update.go)を参照。
- confirmedは全更新ロック。completedはcurrent_revision_versionなしなら結果更新・削除不可、ありなら作業版である可能性を保持する。
- 初回confirmはofficial revisionをappendしCASでpointer更新。unconfirmは理由必須（trim後非空、500文字以内）、認証actor、専用examination-unconfirm:edit、official版あり、親draftを要求しworking版をappendする。
- revisionあり再confirmは変更を先に保存。変更field/itemsと再confirm同送は拒否。修正結果と公式版を別にする。
- revision存在後はpatient／recordの関連変更を拒否。legacyでAPIが関連PATCHを受理できることはD03の患者取り違え訂正の業務許可ではない。
- result／revision／必要監査の同一tx成功、競合CAS失敗時rollbackを目標とし、現revision実装だけでD03全writerが適合済みとはしない。working updateの理由は現実装では内部固定文言を使う部分があり、業務上の訂正理由全体の充足とは別。

<a id="examination-wire-contract"></a>
## 型・未送信／null／旧互換
| 対象 | 現契約 | 目標・停止条件 |
|:---|:---|:---|
| exam create/PATCH | [request](../../../../backend/internal/medicalrecord/examination_request.go)：数値ID、date日時、status5値、result_summary／machine、items pointer | 医院＋IDを各owner refとして扱い、JS安全整数外・不正IDを0にしない。担当医と認証actorは別 |
| PATCH items | 未送信／nullは不変更、[]は空置換 | 意図を保持。空payloadを全削除扱いしない |
| 旧PUT /items | 未送信／nullも空配列化する | D03継承目標は未送信／nullで消去しない・明示[]だけ。旧client切替前に適合済みとしない |
| 項目 | field ID、name、inspection_value、normal_value、unit、reference_value、sort_order | 原値を文字列保持。status/is_abnormalはrequestで信頼せずサーバ導出 |
| 基準値 | [items](../../../../backend/internal/medicalrecord/examination_items.go)で医院／種／fieldを解決しrefMin/Max・定性範囲をsnapshot | 取得失敗と範囲なしを分離。保存snapshotを当時値として使用。医師の具体値承認とは別 |
| 結果ref | clinic＋exam ID、current_revision_version、revision item | examは結果全体、版は固定、itemは版内。跨版IDの同一性を仮定しない |
| 旧版 | legacy confirmedはofficial projectionなし | 最新値からofficialを架空生成せず、旧資料照合まで専用解除・追記adapterを停止 |

[現items service](../../../../backend/internal/medicalrecord/examination_items.go)は型・所属照合→範囲解決→置換・削除監査を同一txに置く。純挿入の監査やD03が求める全変更ログの充足は個別確認が必要。item取得や詳細取得はlab usage receiptを記録し得るため、完全無副作用readだと仮定しない。

## 任意担当医と入力者ログの実装差

[最新の第3項合意](../../../spec/domains/examinations/lifecycle.md#fr-examinations-record)は担当医未選択でも保存でき、認証された入力者のログを保持する目標。現行独立フォームの担当医必須判定は見直し対象であり、requestの型だけでUI・全writer・監査が適合済みとはしない。検査種別等の既存必須条件と医院／患者／カルテ照合・確定保護は維持する。

担当医未選択、既存値の未変更、明示解除、取得不明を区別し、0や入力actor／カルテ担当医で欠損を埋めない。具体的nullable／clearのwire対応は実APIとの適合調査で決め、本書で架空のfieldやendpointを新設しない。ログは認証された入力／編集actor・対象・時刻と保存結果を結び付け、担当医IDをactorへ代用しない。必要ログ失敗時の部分成功防止、独立／カルテ関連／機器取込等の各writerとの対応を確認する。機器由来であることを担当医必須の復活や架空の人間入力者の生成に使わず、既存受信監査との対応を技術確認する。

## 死亡登録後の生前検査の専用経路

[死亡後の生前検査記録](../../../spec/domains/examinations/lifecycle.md#examination-deceased-retrospective)へ従う。専用経路は死亡状態を維持したまま、生前分であること・対象医院／患者／元記録・必要権限・理由・現在版を保存時に再照合する。実施時点不明は推測せず確認まで停止。原記録と前後差分・理由・actor／確認者・入力時刻・元記録関連を保持し、必要履歴失敗は部分成功にしない。結果不明は元操作を照合し別IDで重複作成しない。現行の通常create/PATCH・一律死亡拒否を専用経路の実装済み証拠にせず、API／版／receipt／権限対応は適合確認に残す。死亡状態の一時解除、通常writeへの無制限な死亡gate解除で代用しない。

D03の専用過去入力／確定後追記とD07のofficial revision保護へ接続する。機器から後日届いた結果も、対象患者・元検査・生前の実施時点を確認できる専用adapterで扱い、通常wait／attachのalive条件を全面解除しない。原電文日時と受信／登録時刻を置換せず、情報不足なら確認へ。

<a id="examination-device-contract"></a>
## 機器・ジョブ・互換経路
現行運用は[接続正本](../../../ops/deploy/lab-devices/LAB_DEVICE_CONNECTIVITY.md)と[ADR008 current daily path](../../../architecture/adr/008-local-lab-device-agent.md)。日常はbrowser direct Web Serial、NX600/AU10Vのみ・ページ表示中・origin-wide排他。legacy loopback agentは任意手動経路であり、旧raw token取得説明を現行API仕様にしない。現agent-consumerは短命signed capability、署名鍵を返さない。

| 論理操作 | 保存／境界 | 失敗・再送 |
|:---|:---|:---|
| receive | 1frame=1job、医院/source/fingerprint一意、decoded items保存、生bytes破棄 | 重複は既存job。invalid_payloadは保存成功カードにしない |
| wait | 医院1有効待機、期限／変更／解除、現在生存pet | 期限切れ到着は未紐付け、前患者を再使用しない |
| persist | 保存済みjob/items＋医院内mapping、N exam_typeならN exams | 未対応行をjobへ残す。receivedとpersistedの成功を分離 |
| attach | jobをlock、対象petと既存関連を再照合しpersist | 他患者既存関連は409、結果不明はjob照会 |
| detach | 使用履歴／確定／親保護を照合、job由来全検査の補償 | 安全ならreceived／pet NULL、revertedへ終端化しない |
| fixture commit/revert | fixtureのみcommit、完全一致skip、revertは終端補償 | deviceをfixture allowlistへ追加しない、auto_commit既定false |
| mapping master | clinic/source/device_item_code→field→exam_type | unknownはneeds_review。名称一致やLLM候補のみで確定しない |

[受信service](../../../../backend/internal/medicalrecord/lab_device_receive_service.go)のreceive後persist失敗にはjob pet解除の別処理があり、受信・保存全体atomicと記載しない。GET board／stationはensureStationによる初期化や待機期限処理を伴い得るため、writer棚卸しに含める。通常view専用印刷projectionにこれらを流用しない。

## 権限・全writer
| 経路 | 既存grant／owner | 継承する照合 |
|:---|:---|:---|
| exam CRUD／items／print | examinations view/create/edit/delete、medicalrecord write owner | clinic/pet/record関係、検査自身のロック、必要ログ。旧親finalized一律拒否は最新D03目標との差 |
| unconfirm | examination-unconfirm:edit | actor・理由・official版・親draft |
| frames／wait／board | lab-import:create | trusted clinic、alive pet、受信consumer排他 |
| unlinked／station／masters read | lab-import:view | 医院scope、閲覧可能範囲 |
| attach/detach/revert／機器・item master更新 | lab-import:edit | job source・使用receipt・全由来検査・親依存 |
| 検査type／field／基準範囲 | master-medical操作別grant | 同医院／種／field、旧snapshot保護 |
| D03専用訂正／追記adapter | D03 intentとD07内部を接続 | 両方の現在grant・固定版・理由・結果ref、通常PATCH代用なし |

現在のstaff scopeはJWT snapshotだけで決めず、current access再解決とX-Clinic-ID検証のtrusted scopeを使う。医院／患者／認証世代が変わった遅い応答を別コンテキストへ反映しない。ロック順はD03の全writer規約と整合するよう後続実装で統一確認する。検査内部のexam-first経路と他owner parent-firstが存在するため、静的表だけで全体deadlock-freeを宣言しない。

<a id="examination-owner-contract"></a>
## D03／読取ownerへの返却
検査結果の受渡し先はD03・D04患者トップ・D21の認可済みスタッフ読取であり、今回D20へ検査結果を公開しない。D20の限定健康表示とD21のスタッフ専用Owner Report閲覧認可を区別する。D21への返却に飼主公開や新しい公開承認操作を含めない。
返却は医院、患者／カルテ対応、exam ID、固定revision、版内項目ref、状態、原値・単位・保存範囲、変更前後／理由／actor／時刻、取得状態。取得不能・非開示・旧版不明は区別する。D03の[最終writer／専用intent](../medical-records/technical-spec.md#medical-record-writer-matrix)へ接続し、診療終了後もD03の認可された通常編集と変更履歴へ接続する。旧親finalizedを一律拒否条件にはせず、過去の原記載／参照版は保持する。検査自身のconfirmed／official版保護と必要な専用訂正は維持する。現カルテExamGroupがrevisionを保持しない差、addendumが構造化結果refを持たない差を残す。lab usage receiptは転記／復元／公開のapproval receiptではない。

## AC追跡
以下は先行41ACを技術契約へ対応する索引。追加採用19ACと既存受信権限前提の補足1ACは末尾の追加AC追跡へ分離する。前提／期待結果の正本は業務側のGWT。
| 技術部分 | AC suffix |
|:---|:---|
| 入口・型・scope | entry, scope, required, doctor-optional, editor-log, deceased-history, deceased-boundary, value, unassessed |
| 状態・版・必要ログ | confirm-lock, legacy-seal, unconfirm, unconfirm-denied, reconfirm, finalized-parent, cross-version, write-failure, clear-copy-guard |
| 日常受信・待機 | browser, lock-denied, unsupported, wait, wait-expired, specimen |
| mapping・原値 | clock, unknown-code, multiple-types |
| 冪等・補償 | receive-duplicate, fixture-difference, attach, detach, detach-used |
| 失敗・再開 | frame-loss, result-unknown, autocommit |
| master・医師確認境界 | master-grant, master-history, clinical-approval |
| read・projection | list-scope, print, owner-handoff |

## 利用開始前に残す項目
基準値具体値の医師確認、医院実grant、臨床旧版・item対応、receipt／全監査／競合adapter、物理配線・UAT・PIMS運用承認・医院rollout、旧PUT互換切替は未確認。文書の整合をこれらのPASSにしない。追加の製品要件・自動受信対象拡大は別の判断を要する。


D21の検査履歴sectionはスタッフの閲覧専用であり、同sectionからの印刷・共有を設けない。追加採用の予防薬専用2帳票とは別の範囲である。D03や検査自ドメインに既存の印刷経路があることを、D21検査履歴の印刷許可へ拡張しない。


### 最新カルテ契約へのadapter

[通常編集と検査保護の分離](../../../spec/domains/examinations/lifecycle.md#examination-normal-edit-autosave)を適用する。D03のタブ復元／複製Undoが参照するのは同じカルテ・exam・固定結果版／対象入力で、検査のofficial版を通常更新したり機器jobをdetachしたりしない。検査結果訂正が必要ならD07自身のgrant・理由・版保持・unconfirm／再確定へ従う。旧親draft必須／finalized拒否を最新目標の実装済み保証にせず、親記録の現在版・scope・履歴保護と独立した検査ロックへ判定を分離する。AC-examinations-finalized-parent／AC-examinations-clear-copy-guardはこの接続を検査する。

<a id="examination-selected-print-contract"></a>
## 追加採用の技術契約：項目選択印刷

基準PR670 head `908b1def96c91885c01b038bd71af44f500207bf`の文書を再取得して追加した目標である。先行のコード観察を最新コードの全件再監査へ読み替えない。業務正本は[選択印刷](../../../spec/domains/examinations/lifecycle.md#fr-examinations-selected-print)。既存GET print-snapshotの存在だけでは、項目選択・実施状態・版固定の全条件を充足済みとは言えない。

| 論理input／output | 契約・適合確認 |
|:---|:---|
| 対象 | trusted clinic・pet・exam・保存済みsnapshot識別／固定revision・画面世代。revisionなし旧結果は由来の確認できる不変snapshot識別を使い、現在値から架空officialを作らない |
| 選択 | snapshot内のitem ref集合と表示順。未指定の初期選択と明示的な空集合を区別し、重複refは同じ行を二重出力しない。別exam／別版／不明refは拒否。名称・行番号で版をまたいで解決しない |
| 行の状態 | 実施済み／未実施／実施不明、結果あり／結果待ち／意味不明、取得状態を独立に保持。これらは意味契約で、新しい物理enum／fieldが実装済みという主張ではない。未実施はサーバ側候補・出力双方で除外、実施不明は自動選択せず理由を返す |
| 出力 | 同じsnapshot ref・選択ref集合・順序・原値／単位／保存範囲・各検査日・版／訂正識別。欠損した旧範囲を現在マスタで補完しない。選択0件と取得失敗を別結果にする |
| previewから印刷 | 同じ出力集合を固定し、出力直前にも現在のread認可・医院／患者・対象版対応を再検証。新版発生は無言で差替えず更新通知・再プレビューへ。旧版の明示的再印刷は旧版と分かる同じsnapshotを使う |
| 副作用 | 臨床結果／版・確定・患者関連・初期行生成・会計／在庫・外部送信を変更しない。print向けreadをGetOrCreate／board等のwriterへ流さない。既存readのusage receipt要否は別途照合し、印刷選択を公開approvalへ変換しない |

空のinspection_valueやlegacyの「-」だけでは未実施を決められない。過去行で実施を確認できる結果由来／状態が不足する場合は不明を保ち、対象行だけ照合・適法な訂正へ渡す。未実施の除外をD03治療プランの状態や検査全体statusだけから一括算出しない。print-snapshot応答・選択input・実施由来・版内refのadapterと旧値移行は未実装／未検証であり、新しいURL／DB列を本書で確定しない。

<a id="examination-infection-summary-contract"></a>
## 追加採用の技術契約：患者トップの感染症結果projection

正本は[感染症結果表示](../../../spec/domains/examinations/lifecycle.md#fr-examinations-infection-summary)。D07が結果のwrite owner、D04/D03はスタッフ向けread consumer。患者台帳の通常PATCHや自由備考を結果保存先にせず、D07内の検査結果記録・版履歴を共通に参照する。他院手入力を院内機器jobへ偽装しない。D03の問診検査歴は自己申告の別資料であり、D07結果refの検証と結果記録なしにprojectionの候補へ含めない。未確認申告をofficial版にしない。

| 論理データ | 必要な意味・不明の扱い |
|:---|:---|
| context | 現在認可されたclinic・patient・speciesの明示対応・要求世代。表示対象は同患者であり、同名やidentity groupだけで他医院結果を混ぜない |
| 対象項目 | 猫FIV／FeLV、犬mf／HWと保存されたexam fieldの明示mapping。名称部分一致・配列位置で決定せず、同じ意味への対応を検証できない旧項目はunknownにする |
| 元結果ref | source owner・exam ID・固定revision／snapshot・版内item ref。external手入力もD07の独立記録ID／版を持ち、台帳の表示値とは分離する |
| 原値／状態 | 原値文字列、結果あり／結果待ち／未実施／不明、保存済み／official／作業版・訂正状態。0や陰性をtruthinessで落とさず、医学的状態へ変換しない |
| 検査日 | 元結果の実施日／measured_atとその精度。入力・受信・更新日時は別。日しか分からない他院資料に00:00という架空時刻を与えて順位を決めず、日付不明を当日で埋めない |
| 出典・履歴 | 院内／他院・手入力の区別、判明している出典詳細、元結果と訂正版の関係、認証actor・記録時刻・訂正理由。出典詳細不明を明示し、添付必須は追加しない |
| 取得結果 | loaded／記録なし／結果待ち／要確認／不完全／error／権限不足を分ける。空配列やmissing fieldだけで未検査を確定しない |

### 取得・表示・更新の境界

- D07側で対象患者・対象項目の認可済み候補を解決する。画面の履歴先頭20件をそのまま直近検索の代用にしない。対象projectionに必要な候補の完全性を検証し、不完全なら要約を最新確定として返さない。この限定projectionを全検歴UIの拡張承認へ広げない。
- 未実施／結果待ちは保存結果値を持つ候補と別の状態として返し、新しいオーダーだけで過去の結果値／日付を空値や当日へ置換しない。
- 同じ元結果の版は同一履歴系列として扱い、古いofficialと訂正作業版を独立した別検査へ二重計上しない。表示対象の保存版と状態・元official refを一緒に返す。working保存が新しくても確定済みと表示しない。
- 検査日が判明した候補のうち一意に直近のものを要約できる場合は、その原値・日付・出典を返す。日付／時刻精度不足で順序を確定できない同日候補・同時点複数出典・不一致は候補集合と曖昧さを保持する。院内優先、陽性優先、最新入力優先、結果平均、閾値判定を新設しない。日付不明は別に保持し、既知の直近を上書きしない。
- 院内保存、他院手入力保存、訂正／再確定・適法な無効化の成功で当該clinic・patient・項目のread projectionを再取得／無効化する。保存とキャッシュ再取得は別結果。再取得失敗時は更新未確認を表示し、旧値を現在値として保証しない。未保存draftを投影しない。
- D04/D03の画面read認可に加えてD07結果のview認可と現在clinic/patient対応を検査する。対応するgrant判定をadapterで明示し、owners:viewやmedical-records:viewだけで検査結果を開放しない。scope／species／認証世代変更後の旧応答を捨てる。既存跨医院一覧の閲覧専用境界は拡張しない。
- 手入力create／edit・訂正は既存D07操作別grant、同医院／患者、現在版、原official保護、理由・必要履歴、同一操作の結果不明照合へ参加する。送信成功不明から別IDを作らず、保存済みと確認できた結果だけ投影する。日付不明を許容する論理契約と、現date必須等のwire適合は別であり、架空の日付で通さない。
- 対応mappingの実保存先、外部出典／日付精度の物理schema、専用projection endpoint、保存成功後の通知／cache方式は後続設計で既存型を照合する。具体ID・臨床コード対応は推測しない。master変更権限・履歴保護は既存master-medical／lab-import境界を保持する。
- D20・LIFF／line-reserve等の公開DTOにこの結果・出典・履歴を追加しない。D21の検査履歴sectionは閲覧専用を維持し、患者トップ追加から検査結果の印刷／共有を新設しない。別途採用の予防薬専用2帳票と区別する。トップ閲覧で結果の保存・確定・通知を起動しない。

### 追加AC追跡

以下は追加採用19ACと既存受信権限前提の補足1AC。先行41ACの技術traceは前節を保持する。前提・期待結果の正本はD07業務GWTであり、全61ACの実機PASSを表すものではない。

| FR | 技術節・照合点 | AC-examinations- を共通prefixとするsuffix |
|:---|:---|:---|
| receive | 受信画面の閲覧と受信・attach・detach・切戻しのwriteを操作別grantで分離。不足時の案内・拒否を確認し、新ロールを追加しない。実機未実施 | receive-access-separation |
| selected-print | 選択input・実施／結果の分離・固定snapshot・read認可・取消 | print-item-selection, print-item-unknown, print-item-snapshot, print-item-empty, print-item-version-change, print-item-access, print-item-cancel-draft, print-item-single-snapshot |
| infection-summary | 元結果ref・種／項目mapping・日付精度／出典・再投影・D04/D03認可・手入力write | infection-species, infection-auto-saved, infection-external-source, infection-date-order, infection-ambiguous, infection-revision-refresh, infection-result-states, infection-no-record-failure, infection-access-boundary, infection-mapping-unknown, infection-write-safety |
