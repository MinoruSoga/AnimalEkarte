# D09 健診の技術契約

| 項目 | 内容 |
|---|---|
| 文書ID | design-D09-checkups-technical |
| 種別 | 技術契約・実装差分 |
| 決定状態 | 既存記載を継承（個別の決定根拠は本文参照。根拠未確認部分は承認済みとしない）＋D03合意目標への接続。死亡後の生前分後日入力・誤記訂正、および健診パッケージ設定／手入力・設定なし空欄の最新承認を反映。対象外の未決は保持 |
| 実装状態 | 本文の現コード観察と目標契約を分離。コードに存在するだけでは業務承認済み・目標実装済みとしない |
| 実機検証状態 | 今回未実施。既存証跡の引用は今回の実機受入を意味しない |
| 文書確認日 | 2026-10-21 |

既存仕様の継承とD03の合意済み目標を整理した文書候補。ソフトウェア仕様の整理、実装完了、実機受入、医師による具体的な臨床数値の承認は別の判定とする。

## 対象
[業務正本](../../../spec/domains/checkups/lifecycle.md)の5FR／36AC。基準commit `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`。静的文書／source照合のみ、実装・DB・runtime未変更。[ADR004](../../../architecture/adr/004-checkup-canonical-system.md)とD03[型／writer](../medical-records/technical-spec.md#medical-record-reconciled-contract)を継承。

## 正系統・入口・権限

入口・route・grantの以下の記載は現行観察を含む。目標は[健診FRとAC-quick-entry](../../../spec/domains/checkups/lifecycle.md#fr-checkups-record)のD03通常受付／理由・医師確認付き過去専用入力に従い、独立フォームの親自動createを受付迂回の例外・実装必須にしない。段階保存・結果不明の照合は既存の保護を維持する。
Checkupのみが健診記録／集計／閾値の正系統。Examinationの退役seed IDを再利用しない。masterはcheckup_types／fields、実績はcheckups／field_resultsで医院scopeを保つ。
- /checkups一覧はcheckups:view、独立select-pet/newはmedical-records:createかつedit。
- 保存済み親に対するnested /medical-records/:id/checkupsのCRUDはmedical-recordsの操作grant。結果PUTはedit。package本体CRUD/reorderとfield read/writeはcheckupsの操作別grant。検査・ワクチンmasterのmaster-medicalと異なる。[現routes](../../../../backend/internal/medicalrecord/routes_masters.go)を正本とし、存在しない新grantを発明しない。
- 一覧からの編集はカルテの健診tab/checkupId。独立編集routeは追加しない。

[現package service](../../../../backend/internal/medicalrecord/checkup_type_service.go)は名称必須／医院内一意、価格非負、parent_id同医院、子種別／健診使用がある削除の拒否を持つ。parent解除はClearParentIDであり未送信と分ける。項目writeはfieldService配線済みの場合のみroute登録される。未配線APIを実在する保存先として案内しない。マスタAPIの存在だけで全フィールド編集UIが完成したとはしない。

## 次回予定日とパッケージ設定の適合

[最新第3項合意](../../../spec/domains/checkups/lifecycle.md#checkup-next-date-approved)に従い、packageの次回日設定を同じ医院・選択パッケージで参照し、独立フォーム／カルテ内へ同じ意味で渡す。設定に基づく候補、手入力、保存済み日付、未設定を区別する。新規で設定・保存済み日付・手入力日がなければ空欄を保存・再表示し、既定1yearや任意の固定間隔を作らない。手入力日・保存済み日付は現在マスタ設定から無警告で再計算・消去しない。

現packageの間隔属性／APIの存在だけで設定の編集・読取・全フォーム初期化が適合済みとしない。設定なしと設定取得失敗／未知値を区別し、取得失敗を空欄既定や一律1年候補へ変換して正常取得・保存成功と扱わない。入力・保存済み日付を保持し、既存の取得状態表示・再取得へ接続する。未送信／null／明示解除のwire差は既存型を照合し、既存値の解除方法をこの文書で発明しない。具体的間隔、日付の計算表現、API／DB adapterの適合は実装工程で検証する。本変更は業務目標と受入・traceのみで、実装・設定・臨床数値を変更しない。

## 死亡登録後の生前健診の専用経路

[死亡後記録の最新合意](../../../spec/domains/checkups/lifecycle.md#checkup-deceased-retrospective)へ従う。専用経路は死亡状態を維持したまま、生前分であること・対象医院／患者／元記録・必要権限・理由・現在版を保存時に再照合する。実施時点不明は推測せず確認まで停止。原記録と前後差分・理由・actor／確認者・入力時刻・元記録関連を保持し、必要履歴失敗は部分成功にしない。結果不明は元操作を照合し別IDで重複作成しない。現行の通常create/PATCH・一律死亡拒否を専用経路の実装済み証拠にせず、API／版／receipt／権限対応は適合確認に残す。死亡状態の一時解除、通常writeへの無制限な死亡gate解除で代用しない。

通常Path Bの親create→子create→field置換を専用後日入力の代用にせず、D03の既決過去入力／確定親への追記と同一患者・元記録・実施日で接続する。後日入力は新たな健診実施や未来予約の作成ではなく、死亡状態の通知対象外判定も解除しない。

## 実wire・型・置換
| 対象 | 現状 | 目標adapter／注意 |
|:---|:---|:---|
| 健診本体 | [service](../../../../backend/internal/medicalrecord/checkup_service.go)：checkup_type_id、pet_id、date、next_date、doctor_id、result。versionなし | 現期待版なしを競合制御済みとしない。clinic/親/子/患者/マスタを再照合 |
| 担当医 | doctor_id_clear=trueでNULL、なければdoctor_id指定か保持 | null／未送信を解除としない。actorと担当医は別 |
| 次回日 | optional pointer、旧独立formは1year候補から初期値 | 目標はパッケージ設定の候補または手入力。設定・保存済み日付・手入力日がなければ空欄。旧1year初期値は解消対象。空文字／nullで既存値を解除可能と仮定せず、未送信・未設定・明示解除を分離する |
| field | number／text／boolean／single_select／multi_select／checklist、field ID | 型に対応する値だけ保存。0/false/nullを区別。異種値を他列へ残さない |
| 結果PUT | [request](../../../../backend/internal/medicalrecord/checkup_field_request.go)：results未送信/nullは拒否、[]は明示全消去、最大200件 | PUT単位の置換。UI表示上の空と意図した全削除を区別する |
| 定義解決 | [field service](../../../../backend/internal/medicalrecord/checkup_field_result_service.go)：field名／型／単位／範囲をサーバ解決・snapshot | 別医院／package field・無効choice拒否。status/is_abnormalをclientから採用しない |
| 数値判定 | 数値とmin/max比較、欠損値は現内部status normal／falseになり得る | 内部値だけを「正常確認済み」に表示せず、入力・範囲・評価可否と併せて扱う |
| 旧field参照 | field削除でID NULLとなり得る | 保存済み名称／型／値／単位snapshotを保持、最新fieldへ名前で再接続しない |

[現field service](../../../../backend/internal/medicalrecord/checkup_field_result_service.go)は親draft lock・置換・実削除時の前後監査を同一txに置く。純挿入や本体CRUD全ての必要監査が充足したとはしない。定義取得と親lockの時点差、package変更／field改定競合、期待版未対応をD03全writer適合で後続検証する。

<a id="checkup-save-recovery"></a>
## 独立登録の部分成功と復帰
[現useCheckupForm](../../../../frontend/src/features/checkups/hooks/use-checkup-form.ts)は以下の連続要求である。
1. createMedicalRecordForCheckupでpet/owner/visit_dateを渡し親作成。
2. そのIDへcheckup本体POST。
3. 値があればfield-results PUT。
4. 全必要段階成功後に成功表示／一覧復帰。

各間でgrantと死亡状態を再検査するが、途中拒否・通信失敗時は既に作った親／健診が残り得る。現コードは全体のidempotency/recovery receiptを実装した証拠ではない。目標adapterは各成功ID・親子対応・保存段階と結果unknownを保持し、再開時に照会・既存ID利用する。保存先が確定できなければwrite停止。全体atomic化、新規自動削除補償、通常カルテの過去専用入力代用を追加しない。

保存値のGET結果をフォームへ復元する配線、独立フォーム離脱blockerは現screen記載の制約を保持する。backend GET field-resultsが存在することだけで編集UIが値を復元済みとしない。一方独立作成のPUT配線は上記sourceに存在するため「field結果API全て未使用」とは記載しない。

## 全writer・副作用・返却
| writer/read | 保護／受け手 |
|:---|:---|
| 健診本体CRUD、field PUT | 旧親finalized拒否は最新D03の通常編集目標との差。clinic/pet/record/checkup/package・現在権限を再照合し、D03必要監査・版／結果receiptへ参加 |
| 独立親create→子create→field PUT | 部分成功の保存段階、前段ID、unknown停止。各段成功を全体成功にしない |
| package／field master変更、package import | 対象医院と定義参照。既存結果snapshotを上書きせず、臨床数値確認とマスタ編集者を分離 |
| D03専用訂正／過去／復元 | 元記録・理由／actor／確認・現親依存。普通PATCHや別IDcreateへfallbackしない |
| タグ同期／followup | [現service](../../../../backend/internal/medicalrecord/checkup_service.go)はcommit後同期・非同期非致命的trigger。現create後triggerは実装参照。目標はスタッフの実施済み・対象確認後、一般案内／医師要フォロー指定が必要な医療案内を分離して[D15確認adapter](../../../spec/line/architecture.md#followup-confirmation-adapter)へ。保存成功・確認保存・送信結果を分け、既存認可／必要監査／原子保存単位・結果不明保留・重複防止と現pet/owner/alive再確認を継承 |
| field-results by pet／保存済み読取projection | clinic/pet/record/checkup/field ref・型・値・保存snapshot・取得状態・更新識別を返し、D03およびD21のスタッフ閲覧認可で表示。今回D20へ健診結果を公開せず、保存済み接種実績・最終来院日等の限定健康表示と分離する。D21に飼主公開や新しい承認操作を追加しない。表示／公開成功と臨床確定を分離 |

本体CRUDはslog中心で全変更の同一tx監査を満たすとはしない。版なし現DTOは新競合防止を満たすとはしない。D03が求める全writer／必須監査／認証世代・scope・receiptのadapterは技術適合課題として明記し、実装済みとしない。D15のログだけで失敗が人に届く通知まで完成したとしない。

## AC追跡
| 技術部分 | 業務AC suffix |
|:---|:---|
| 正系統 | canonical, retired-id |
| 入口・本体型 | chart-entry, quick-entry, edit-entry, required, doctor-clear, next-manual, next-package-candidate, next-no-setting-blank, next-history-preserved, next-setting-unknown |
| field型・置換 | value-types, field-scope, invalid-choice, replace-missing, replace-empty, replace-failure, master-change, unassessed |
| 保護・復帰 | deceased-history, deceased-boundary, finalized, scope-grant, partial-save, unknown-result, concurrent, leave-limit, clear-copy-guard |
| 一覧・受け手 | alert, list-scope, followup-confirmation, followup-purpose, followup-recovery, notify-failure, publication |

## 完了判定外
医師による具体的数値／単位／種／出典確認、既適用旧seed／患者データ、実grant、期待版／必要監査／receipt／復帰adapter、実機・並行性・性能・通知／公開運用は未確認。文書完成と利用開始は別に判定する。


D21の健診履歴sectionはスタッフの閲覧専用で、同sectionからの印刷・共有を設けない。最新採用の予防薬専用処方箋・年間投薬表だけの印刷は別範囲であり、健診sectionや一般レポート全体へ拡張しない。D03や検査自ドメインの既存印刷経路も別に保持する。


### 最新カルテ契約へのadapter

[通常編集・自動保存／復元](../../../spec/domains/checkups/lifecycle.md#checkup-normal-edit-autosave)に従う。旧親finalizedを一律編集不可の目標として維持せず、対象checkup・親record・患者・package／field・現在権限・期待版を照合する。復元基準はD03が保持する対象タブ／複製操作と同じID・版の入力値であり、field名や配列位置だけで別対象へ適用しない。変更と必要履歴／操作結果は保存単位で一体にし、原値・復元差分を残す。現状のPath B段階保存を全体atomic済みとせず、保存結果不明・部分成功は既存IDを照会する。新しい通知／会計／在庫操作を自動実行しない。AC-checkups-finalized／AC-checkups-clear-copy-guardを本adapterへ対応付ける。

<a id="checkup-stage-recovery-candidate-20261019"></a>
## CL-R09：段階保存の再開資源案（2026-10-19）

基準 `563353287ebb08af55056e3fbfba2aca3531e599` の[独立保存hook](../../../../frontend/src/features/checkups/hooks/use-checkup-form.ts)109〜145行は親create→子create→必要時field-results PUTを逐次実行する。途中grant／死亡再検査とcatchはあるが、ここに段階別永続receipt・保存済みIDからの再開は確認できない。全体atomic化や自動削除補償で置換しない。

内部候補`checkup_save_attempt`は不変attempt ref、認証scope、元患者・実施日、同内容fingerprint、各stageのoperation key、親record ID、checkup ID、結果置換版、各段の未送信／不明／成功refを持つ。業務内容変更は未解決段を照会した後に別intentとし、stage結果を新患者へ再利用しない。親成功後に子が失敗した場合、同attemptの既存親へ子を保存する。子commit後応答断なら同key結果からcheckup IDを回収して同IDの結果置換へ進み、GET0件だけで別親／子createしない。親／子／結果成功の各receiptが揃うまで全体成功にしない。

旧fieldは保存snapshotと原field ref／型／値を利用し、削除済みIDを名称で現fieldへ再接続しない。GETから復元するadapterは0／false／null・未評価・取得失敗を区別し、型不一致は未知として止める。package変更と定義改訂の期待版は結果置換の同じtxで照合する候補で、現行実装済みとしない。

検証は親後失敗、子後応答断、結果PUT後応答断、別患者遅延応答、field削除／同名再作成、0／false、途中失権を分け、再開前後の親子件数・ID・原値・必要監査を比較する。任意段階の成功を全体成功とせず、タグ／followup成功とも分離する。現在未実行であり、この案を保存しただけでCL-R09のruntime部分を完了にしない。

## 10/19横断検索・同種操作の設計接続

[共通adapter](../../shared/integration-contracts.md#cross-ui-search-adapter-20261019)を本領域の[適用台帳](../../../spec/shared/interaction-scope.md)L/M/H/S consumerへ接続する。draft/applied・独立filter・取得scope/世代・復帰/失敗状態を分離し、検索ボタンだけでキーワードを適用。既存API検索と取得済み内絞込みの責務、ページング/上限、認可・保存方式を維持する。E/N操作へ部品単位で一括適用しない。実装/全caller/実機未確認。

<a id="plan-tab-link-design"></a>
## 治療プラン連動の設計受け手（2026-10-20）

業務目標は[D03共通FR／AC](../../../spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-plan-domain-tabs)、本領域の[接続境界](../../../spec/domains/checkups/lifecycle.md#checkups-plan-tab-link)を参照する。Checkup系を唯一の健診正系統とし、パッケージ・型付き結果・保存時snapshot・次回予定の既決を維持する。予定追加だけで健診完了・結果生成・親カルテ自動作成へ進めず、Examinationへ同じ健診を二重保存しない。既存の通常受付・過去専用入力を迂回する新経路にしない。

[D03の現行観察と設計残件](../medical-records/technical-spec.md#plan-domain-tabs-target-gap)から関連の安定した識別・必要版と結果照合を設計する。表示のための物理資源作成を必須とせず、projectionの選定も未了。各ownerの保存単位・訂正／監査・再送保護を維持し、全owner一括transactionや一律1対1を新たな決定としない。実装適合／実機検証は未実施。

<a id="checkup-outsourced-pdf-gap"></a>
## 外注PDFの採択要件と未収録の技術差分（2026-10-21）

[FR・7AC](../../../spec/domains/checkups/lifecycle.md#fr-checkups-outsourced-pdf-20261021)を追加。従来のFR/AC件数表記は旧範囲であり、この追加分を含む完成判定ではない。実装・DB・API変更・試験はNOT_RUN。既存Checkup正系統・型付きfield・置換APIの存在は、原本／抽出／確認版の実装証拠ではない。

| 必要な差分 | 設計で満たす契約／未選定部分 |
|:---|:---|
| PDF storage | 認可された医院・患者・親カルテ・Checkupに原本を対応付け、原本所在と各版の由来を保持。既存添付を使う場合も適合確認が必要。不変原本port・DB境界・詳細DDLの選定は[EMR-500 PDF採択技術契約](../../../work/emr/emr-500-outsourced-checkup-pdf-contract-20261021.md) §§3–4に従う。実保存先・実環境配線は未完。TTL・物理削除を追加しない |
| 抽出／確認wire | 原本登録、抽出候補の取得／再抽出、職員確認・修正保存、確定結果読取を意味的に分離。route・closed DTO・詳細DDLは[EMR-500 PDF採択技術契約](../../../work/emr/emr-500-outsourced-checkup-pdf-contract-20261021.md) §§2–6へ選定済み、runtime未実装。既存field PUTの一括置換だけで確定結果を無履歴上書きしない |
| 版・監査・再開 | 原本ref、候補版、確定結果版、確認者／actor・時刻・前後差分と操作結果を追跡。現在認可・対象版・依存を再照合。応答断／競合は同一操作と版を照合して再開、再抽出は確定値を変更しない |
| 状態と受け手 | 原本登録／抽出候補／職員確認済みを区別。失敗・空・未確認をnormalへ畳み込まない。D03は同一Checkup結果ID・版をprojectionし、D07本文を複製しない |
| 抽出基盤 | PDF見本未受領。OCR provider、課金、外部送信、抽出項目mappingは未選定。実患者PDF処理・外部送信は本作業で実施しない |

現在の医院・患者・record・Checkup・masterの照合、通常受付、過去／死亡後生前分の専用訂正、原結果保持と手入力を維持する。確認保存の必要履歴が失敗した状態を確定成功として返さない。D03の一般PDF閲覧は抽出実行の入口と同一視しない。各差分の受入は上記7ACへ対応する。storage／wireの選定範囲は[EMR-500 PDF採択技術契約](../../../work/emr/emr-500-outsourced-checkup-pdf-contract-20261021.md)を正とし、providerなしで先行する必要subsetと全writer/readerの実装依存を着手プランへ渡す。provider・費用・外部転送承認、見本mapping、実環境配線と受入は未完のまま。

<a id="package-next-date-contract-20261024"></a>
## 次回日rule authoringから両入口までの契約案（EMR-565）

基準main `c0ee0914e723791c59cfa971cbb1ad84cff171db`。[既決の次回日・5AC](../../../spec/domains/checkups/lifecycle.md#checkup-next-date-approved)を実現する技術選択案。製品・API/DDL反映・実装供給済みではなく、全実行NOT_RUN。自由文intervalとversion/CASは実在するが、自由文の計算意味は未確定。年4回/春夏秋冬/年1回/1年を自動parseせず、臨床間隔・季節日を補わない。

### 機械ruleの保存・候補

checkup_typesへnullable JSONB next_due_ruleを追加する案。既存intervalは表示頻度として保持する。既存master POST/PATCH/GETと編集paneに一属性を収容し、policy table/新画面/承認workflowを作らない。

closed shapeは `{schema:"checkup-next-due/v1",kind:"none"}` 又は `{schema:"checkup-next-due/v1",kind:"elapsed",unit:"day"|"month"|"year",amount:positive integer,anchor:"performed_date",missing_day:"no_candidate"}`。未知/重複key、NULL member、不正unit/amount、数値stringは400。amountは正確な正整数で、表現可能な暦日の計算範囲だけを受理し、独自の臨床上限を設けない。型の日付範囲外はinvalid_ruleで候補なし。季節month/dayは新設しない。

保存NULLは未収録。POST省略NULL、PATCH省略保持/明示nullは400、設定なしはkind:none。旧interval空＋rule=NULLはnone、非空＋NULLはunknown/legacy_text_only。ruleがあればruleだけを計算元とする。既存master操作grant・version CASを維持し、checkups:viewから設定writeへ昇格しない。既存paneで自動候補なし/実施日から数＋日月年を明示入力、既定数値なし。import/manifestはoptional同schema、省略保持。既存immutable seed/季節値を勝手に変えず、正規seed生成は本人担当。

JST calendar additionを使い、対象月に同じ日がなければunknown/calendar_day_unavailable。1/31＋1月、2/29＋1年を月末へ丸めず手入力可。dayもUTCミリ秒差でずらさず、具体的な医学的間隔は導かない。

候補GET案 `/v1/masters/checkup-types/{id}/next-date-candidate?performed_date=YYYY-MM-DD`。既存master GETにもrule/versionを返す。read grantは既存ResourceCheckupsのcheckups:viewで、master-medical:viewではない。medical-records:editだけから権限を作らず、取得不可でも手入力を保持する。

response案は `{schema:"checkup-next-date-candidate/v1",clinic_id,checkup_type_id,rule_version,performed_date,state:"candidate"|"none"|"unknown",candidate_date:YYYY-MM-DD|null,reason:null|"legacy_text_only"|"calendar_day_unavailable"|"invalid_rule",rule_ref}`。candidateのみdate非NULL、none/unknownはNULL。503/networkは取得失敗でnoneに変換しない。入力日付不正400、認証/scopeは既存共通エラー。**未閉鎖:** rule_refのexact shapeとidentity/versionのwire型は受領稿で未定義であり、API ownerへ限定閉鎖を渡す。任意JSONで補完しない。

### 両入口の状態と保存

独立CheckupFormとカルテ内CheckupsTabは同hook/reducerを使う。next_date_source=pristine|candidate|manual|savedとcandidate_status=loading|candidate|none|unknown|errorは別軸。新規pristineだけ候補を初期反映。手入力/手動clear後はmanual、保存済みはsavedで、package/実施日変更・再取得/現master改定で無言上書きしない。新候補は別表示し明示採用、unknown/失敗時も入力保持、noneのpristineは空欄・任意。

既存checkup create/updateへpresence付きnext_dateを渡す。create省略/NULLは未設定、update省略保持/NULL clear/日付set。next_date_intent=manual|package_candidateを同payloadへ含める。package_candidateはrule_ref・元performed_date・candidate_date必須、manualはrule_ref不要、NULL clearもmanual。next_date省略時はintentも省略、片方だけ400。省略adapterでclearを失わない。

package_candidateだけ同TXで現在master版/ruleと日付を再評価し、不一致409 next_date_candidate_stale、日付を最新値へ自動変更しない。manual/saved保持はrule改定/取得不能を理由に拒否せず、既存validation/scope/親/必須条件を維持。checkup自身は500/C06の期待版・結果照会へ参加しmaster.versionで代用しない。unknownは同key/作成済みID照会、新createで救済しない。GETは保存next_dateを返し現ruleへ依存させず、監査に由来を残す場合もmanualへ架空ruleを付けない。保存後GET失敗は成功保持・readだけretry。

### 同PRの供給・完結

447がruleのrequest/response/service/repository、既存master pane、manifest parser、500が既存writer期待版/結果照会、445がclosed DTO/正規生成、446が保存差分、565が両フォームと共通candidate hook/calculatorを同PRで結ぶ。外注PDF/OCR/providerやmaster全機能Doneは不要。現在rule実装なしを単独Ready停止理由にせず、同PR差分/owner/CAS/route契約の受領後に設計Ready判定する。rule_ref等の未閉鎖は先に解消する。実packageへの実数値seed投入は完了条件でなく、合成値は臨床設定ではない。

5ACへ同rule同日両入口、旧自由文無推測、none/unknown/error、manual/clear、保存済み保持、master競合、月末/閏日、他院/親不一致/失権、応答断同ID照会を対応付ける。供給SHA・両実caller接続・採用SHAは実装後の証拠で、現稿を実機PASSとしない。
