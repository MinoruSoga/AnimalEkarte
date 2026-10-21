# D08 予防接種の技術契約

| 項目 | 内容 |
|---|---|
| 文書ID | design-D08-vaccinations-technical |
| 種別 | 技術契約・実装差分 |
| 決定状態 | 2026-10-17追加決定は末尾と進捗入口を参照。従来：既存記載を継承（個別の決定根拠は本文参照。根拠未確認部分は承認済みとしない）＋D03合意目標への接続。次回日は種類別設定または手入力へ統一する既合意、および過去接種の当時の次回日を保持する最新第4項合意を反映。対象外の未決は保持 |
| 実装状態 | 本文の現コード観察と目標契約を分離。コードに存在するだけでは業務承認済み・目標実装済みとしない |
| 実機検証状態 | 今回未実施。既存証跡の引用は今回の実機受入を意味しない |
| 文書確認日 | 2026-10-20構造・参照確認（従来確認：2026-10-17追加差分、2026-10-15） |

既存仕様の継承とD03の合意済み目標を整理した文書候補。ソフトウェア仕様の整理、実装完了、実機受入、医師による具体的な臨床数値の承認は別の判定とする。

## 範囲
[業務正本](../../../spec/domains/vaccinations/lifecycle.md)の現行6FR／39ACを対象とする。先行5FR／28ACは旧読取範囲であり、後続条件を含む現在の件数ではない。先行の基準commit `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`の静的照合で、今回runtime未実施。D03[保存§2](../../../spec/domains/medical-records/lifecycle.md#hybrid-sync)・[型／writer契約](../medical-records/technical-spec.md#medical-record-reconciled-contract)を保持。

## 入口・資源・保存結果
独立とカルテ内は同じGET/POST/PATCH/DELETE /api/v1/vaccinations。カルテ内GETはpet_id＋page/limit、POSTは保存済みmedical_record_idを送る。独立はrecord NULLで保存し得る。接種自身にfinalized状態を新設しない。同日複数は単件POSTの順次実行、バッチAPIを推定しない。[EMR-105](../../../work/emr/emr-105-vaccine-multi-20261002.md)の過去検証は保存済みカルテ内3件実DB／2件モックフォーム等の限定証拠で、今回の再実行や全UI受入とは別。

[service](../../../../backend/internal/medicalrecord/vaccination_service.go)はcreate/updateで同医院の関連と親状態をtx内に照合し、updateは旧／新親を安定した順序でlockし元relation tupleを再照合する。deleteは親→子lock・同一tx削除監査で、失敗時同期を発火しない。create/updateの全編集者履歴は現slogだけでD03の必須監査を充足したとしない。

## 型・日付・clear・既定
| 項目 | 現契約 | 適合／停止条件 |
|:---|:---|:---|
| ID | [request](../../../../backend/internal/medicalrecord/vaccination_request.go)：pet／record／doctor optional数値pointer、vaccine必須 | 医院＋IDと認証actorを分離。不正IDを0へ補完しない |
| 接種日 | date必須、serviceはJST暦日で未来拒否 | UTC instant差と暦日を混同しない。新規予定を接種実績にしない |
| 次回日 | optional、serviceは指定時に接種日より後を検証 | 通常独立新規の本日以降条件と、専用後日入力での当時の期限経過日保存を分離。APIに同じ本日条件／専用例外が実装済みと断定しない |
| 間隔 | 現行実装は独立・カルテ内とも[共通定数](../../../../frontend/src/constants/next-schedule.ts) `DEFAULT_NEXT_SCHEDULE_TYPE` による `1year`（EMR-430）。旧カルテ内 `4weeks` は統一前の履歴 | 目標は両入口ともワクチン種類別設定または手入力。現行の1year統一だけで目標適合済みとはせず、手入力値を黙って上書きしない。2026-10-20の本人回答により設定なしの初期値は空欄。一律1yearは採用しない。具体間隔の臨床承認・その他の新たな優先順位は追加しない。実装適合未確認 |
| LOT等 | lot1〜4、supplemental、remarks | 空文字と未送信を区別、他行に値を持ち越さない |
| PATCH nullable | optional pointerのnullは未送信と区別できない | doctor／next_date／record解除済みと誤表示しない。対応する専用clearが未定義なら解除を送らない |
| 期待版 | version任意、nilは互換上照合skip | 新adapterは期待版を保持。旧clientが無版で更新できることを競合防止完成としない |
| snapshot | vaccine参照と価格preload等 | 当時名称／価格／間隔の完全snapshotありと仮定しない。D03/D10の当時値確認へ |

現vaccines masterは[master routes](../../../../backend/internal/medicalrecord/routes_masters.go)のmaster-medical操作別grantによるCRUD/reorder。接種実績のvaccinations grantとは別で、checkup masterのcheckups grantとも区別する。

次回日の統一は[最新の業務合意](../../../spec/domains/vaccinations/lifecycle.md#fr-vaccinations-schedule)へ従う。種類設定の参照と手入力日を両入口で同じ意味に対応させ、設定なしの新規初期値は空欄にする（[EMR-430](https://app.plane.so/baritechllc/browse/EMR-430/)）。設定取得失敗を設定なしと同一視せず、保存済み日付・明示手入力を初期値処理で消さない。既存schedule値を新しい既定・臨床的標準として固定しない。設定の保存／読取・入力状態の保持・両フォームadapter・再表示を照合し、両入口の既定が共通の1yearになったことと、種類設定／手入力の目標に適合することを分けて確認する。設定変更から過去の保存済み接種／次回日を自動再計算しない。技術対応はAPI／DBの適合調査で具体化し、本docs-only変更でwire schemaや実装を追加しない。

## 過去接種の当時の次回日を保持するadapter

[最新第4項合意](../../../spec/domains/vaccinations/lifecycle.md#vaccination-retrospective-next-date)の対象は過去接種の専用後日入力。実接種日と当時のnext_dateを履歴値として保存し、入力時刻・actorと区別する。next_dateが入力時点より前でも実接種日より後なら期限経過だけで拒否せず、今日以降への補正や現masterによる再計算をしない。通常の独立新規入力には既存の本日以降条件を維持し、未来接種日の拒否と実接種日との前後関係を両経路で保つ。

対象医院／患者／元記録・親状態・grant・必要理由／履歴を既存専用経路で照合し、通常POST/PATCH全体の検証解除で代用しない。[D03正本](../../../spec/domains/medical-records/lifecycle.md)の過去入力・通常終了後編集と、死亡後生前分の専用処理を分けて接続する。通常カルテを終了後追記だけに制限せず、確定会計snapshotの保護は別に維持する。選択した履歴値は再取得・projectionでも同じ日付で返す。現行APIの接種日より後という検証だけでは、専用UI／入力目的の識別・履歴／認可対応まで完成していない。必要adapterのAPI／DB適合は後続の実装確認であり、本変更で新wire fieldや許可経路を捏造しない。

## 死亡登録後の生前接種の専用経路

[死亡後の生前分記録](../../../spec/domains/vaccinations/lifecycle.md#fr-vaccinations-correction)を最新合意の目標として扱う。専用経路は死亡状態を維持したまま、生前分であること・対象医院／患者／元記録・必要権限・理由・現在版を保存時に再照合する。実施時点不明は推測せず確認まで停止。原記録と前後差分・理由・actor／確認者・入力時刻・元記録関連を保持し、必要履歴失敗は部分成功にしない。結果不明は元操作を照合し別IDで重複作成しない。現行の通常create/PATCH・一律死亡拒否を専用経路の実装済み証拠にせず、API／版／receipt／権限対応は適合確認に残す。死亡状態の一時解除、通常writeへの無制限な死亡gate解除で代用しない。

実接種日と現在の登録時刻を分け、[D03正本](../../../spec/domains/medical-records/lifecycle.md)の死亡後生前分に限る専用処理・原記録／履歴保護を維持する。通常終了後編集の禁止や、確定会計snapshotの通常書換え許可へ読み替えない。次回日の統一合意から死亡後の新しい接種／予約を許可せず、通知の死亡除外も解除しない。

## writer・失敗と受け手
- 全入口CRUD、関連変更、カルテ内フォーム、専用訂正／転記／過去入力／復元adapter、タグ同期を列挙する。旧親finalizedによる一律拒否は最新D03の通常編集目標へ合わせる適合差であり、現在の患者／医院・権限・版・履歴を検査する。患者取り違え・死亡後専用処理・削除の保護を普通APIで回避しない。
- フォームのgrant・死亡判定だけでサーバ全経路保護済みとしない。現在clinic/pet/record/doctor/vaccine関係とactorを検査し、scope変更後の古い応答は破棄／再取得する。
- 送信失敗は入力保持。通信断はunknownとして保存先ID／操作結果を照会。現単件POSTに汎用operation receiptがあると仮定せず、結果不明の再createを止める。既存ID／実操作の照合方式は実装で具体化する。
- D03から継承する必須編集ログは変更前後・対象・理由・actor／時刻・結果を同じ業務単位で保持する。現delete監査だけを全writeの充足としない。
- [現service](../../../../backend/internal/medicalrecord/vaccination_service.go)のタグ同期はcommit後、失敗はログであり接種保存をrollbackしない。同期／配信失敗通知・再試行・停止・auditはD15へ渡す。現在ログのみで人への失敗通知まで満たすとしない。
- D20/D21へclinic/pet/接種ID・date・vaccine表示情報・next_date・取得状態・更新版／時刻を渡す。保存済み原事実の受渡しで、D08のview／createは飼主公開認可ではない。D20の飼主公開条件とD21のスタッフ専用Owner Report閲覧認可は各所有者の既存／合意正本を適用する。D21に飼主公開や新しい承認操作を追加せず、D20の承認済み公開条件に従う。本人に紐付く生存ペットの保存済み接種実績を単位とし、親カルテdraft／紐付きなしでも新たな親finalized条件や接種承認操作を要求しない。未保存・取消／削除は非公開、訂正保存後は更新する。根拠は[健康手帳正本](../../../spec/screens/38-liff-pet-health.md#liff-health-document-contract)の2026-10-15承認 Sentinel_7a67fd30e47c819192cd36d416f0a2bf。実装適合は別途検証する。
- 会計金額snapshotはD10、在庫副作用はD13。臨床訂正成功と会計／在庫修正receiptを別にする。

## AC追跡
| 技術部分 | 業務AC suffix |
|:---|:---|
| 同じ資源・単件保存 | standalone, in-chart, same-day, reset, lots |
| 日付・マスタ候補 | future, master-interval, manual, next-invalid, retrospective-expired-next, normal-new-next-past, retrospective-protection, default-boundary, no-setting-default |
| write保護・結果 | finalized, scope, deceased, deceased-boundary, delete-one, conflict, unknown, edit-log, clear-copy-guard |
| 一覧・読取 | filter-scope, navigation, projection |
| master／他owner | master-history, notify-failure, no-side-effect |

## 完了判定外
種不一致の医学的妥当性、製品別の具体的間隔承認、医院実grant、旧snapshot、create/update監査の充足、全client期待版・receipt、D15配信運用・D20公開条件の実装、実ブラウザ／実データ受入は未確認。機能追加の未承認事項と既存挙動の記載を分離する。


D21 Owner Reportの接種履歴sectionはスタッフ閲覧専用で、接種情報の印刷・共有を同画面から追加しない。最新採用の[予防薬専用処方箋・年間投薬表](../../../spec/screens/39-owner-report.md#owner-report-preventive-output)だけがD21の限定印刷例外であり、D03の予防薬管理とD08の接種実績／証明書を区別する。D21一般レポート全体や本sectionへの出力許可に拡張しない。


### 最新カルテ契約へのadapter

[通常編集・自動保存／復元](../../../spec/domains/vaccinations/lifecycle.md#vaccination-normal-edit-autosave)に従い、旧親finalizedだけを拒否理由にする判定を目標から外す。ただし患者／親recordの同一性、現在grant、対象版、日付制約、死亡後の専用理由／履歴と必要監査は維持する。復元操作は対象の接種ID・基準版・現在版・前後値・操作identityを照合し、必要な変更履歴と保存結果を一体にする。新規接種・別タブ・通知／在庫／会計の操作を復元に連鎖させない。既存PATCHの監査・期待版／receipt不足を対応済みとはせず、結果不明は同一操作照会後に続ける。AC-vaccinations-finalized／AC-vaccinations-clear-copy-guardがこのadapterの受入条件。

<a id="certificate-edition-proposal"></a>
## 証明書の固定版・訂正・再出力の詳細設計案

[発行内容と履歴の業務正本](../../../spec/domains/vaccinations/lifecycle.md#certificate-editions-20261017)を実現する[C06設計提案](../../shared/integration-contracts.md#detail-design-proposals-20261018)。通常縦横／狂犬病の旧見本の項目・配置を維持し、以下の内部ID・生成順・検証DCを本人の直接承認としない。専用実装の所在／適合は以前のGitHub調査で未確認、今回も既存一般PDFの存在だけから証明書実装済みとしない。

### 発行版と描画資源

| 資源／責任 | 保存する識別・版の提案 | 制約 |
|:---|:---|:---|
| 接種事実／D08 | vaccination ID・記録版・元医院/pet・接種日と当時の値 | 発行操作から接種create／実施／会計を起動しない。独立接種も同じ型で照合 |
| 証明系譜／D08 | clinic＋対象接種ref＋証明種別＋不変series ID、series revision | 通常縦横は本人第3項Aにより同じ証明書の最新版を共有する。内部seriesはその共通管理を表す技術案。狂犬病を同じ系譜へ統合せず、LINE／飼主／表示名をkeyにしない |
| 発行版／D08 | 不変edition ID、系譜内単調版、元接種版・発行時基本情報snapshot、issuer actor/時刻・訂正元edition | 発行後payloadを上書きしない。基本情報の現在値から再構築しない。未発行draftと有効発行を区別 |
| 描画定義／帳票adapter | 既知3見本のtemplate版（RV4種類の追加対応は未収録）、font/asset・renderer版、項目対応・表示ルール、snapshot digest | 旧見本と版付きgolden fixtureで照合。現在templateで過去発行版を黙って描き直さない |
| 出力artifact／帳票adapter | edition＋様式＋template/renderer版＋digest、immutable blob参照・状態 | blobの所在不明／digest不一致は失敗。新しいblank artifactや別接種へfallbackしない |
| 発行／再印刷操作結果 | operation keyとedition/印刷attempt ID・出力版・必要履歴 | 印刷dialogを開いたことと物理印刷完了を区別。取消・通信断で接種や発行を二重登録しない |

snapshotは正本で採用した項目だけ。原見本の縦横差を共通化で消さず、未取得2資料や法定必須項目を推測追加しない。新しい証明番号の業務規則は選定せず、内部edition IDと印字番号を区別する。番号が必要なら既存の印字仕様と利用院の確認を行い、内部UUIDを勝手に法的証明番号として印字しない。

[2026-10-21本人確認の用紙寸法](../../../spec/domains/vaccinations/lifecycle.md#certificate-editions-20261017)に従い、通常縦はA4縦（210mm×297mm）、通常横・既知の狂犬病見本はA4横（297mm×210mm）を各様式のtemplate版へ固定する。未知のRV4種類へこの寸法を一律適用せず、名称・見本との対応と用紙は資料待ちとする。寸法は固定描画版の一部として保持し、現在templateで過去発行版を黙って描き直さない。

### 同じ発行版からの縦横出力

[本人第3項A](../../../spec/domains/vaccinations/lifecycle.md#certificate-shared-layout-edition-20261018)は業務決定、内部方式は本案と分ける。通常証明書のeditionは両様式で必要な採用項目の保存値を保持し、縦横別のtemplate版・field projectionを参照する。LOT欄／E-Mail等の差を消さず、現在master/基本情報から欠けた値を黙って補わない。旧editionが他方の様式を表現できるか不明なら未対応／取得不明を返し、推測で成功させない。具体移行の調査は実装課題で、新しい必須入力や新規発行の強制をこの決定から作らない。

再印刷入力のlayout選択は同じcurrent editionとartifact（edition＋layout＋template/renderer版）を参照する。別layout artifactを未生成なら同じ不変snapshot・固定描画版から生成し、有効性／現在認可を再検査して渡す。layout切替だけで発行operationやedition増分を起動しない。発行時に両layout artifactの事前生成を新しい業務必須条件にしない。訂正時の有効性gateは縦横で共通であり、旧版を別layoutから有効出力しない。

### 発行の準備・commit・再試行

1. 発行writeの候補は既存のvaccinations:edit、出力readはvaccinations:viewへ対応付け、医院/pet/接種、source版と表示対象版をownerで照合する。このgrant対応は技術提案で現行専用routeの存在を意味せず、医院の実grant・既存認可正本との適合が必要。閲覧権限だけで発行へ昇格せず、新たな医師資格必須やgrant名を作らない。提案入力はsource ref/版、series期待版又は新規none、様式、必要訂正元ref、operation key。
2. read-only previewのsnapshot・templateを固定し、生成する。大きな描画・storage uploadはDB lock保持中に実行しない。描画中は未発行候補で、有効版として提示しない。private artifactを準備し、描画失敗は発行状態・原接種を変更せず再試行状態へ。訂正保存で既に有効出力を停止した旧版を描画失敗で復活させない。
3. 必要artifactが揃った後、共通gate・receipt・D08の既存parent/child順・series lockで認可・元版・series版を再検査する。同じtxで不変発行版＋current参照＋旧editionのsuperseded関係＋必要監査＋成功結果を保存する。発行確定が失敗した候補artifactは有効出力へ公開しない。
4. 発行commit後の応答断は同key照会で同editionを再利用。元情報が変わった描画候補はそのままcommitせず再取得・確認する。業務payload変更は先行結果を解決後に別key。renderだけ再試行する場合も同候補identityとdigestを保持する。

描画準備とDB commitを跨ぐorchestrationなので、一つのACID transactionと呼ばない。未参照artifactのcleanupはC05の原本所在確認へ従い、発行済み原本・必要履歴を通常cleanupで消さない。正本artifactが失われた場合は出力不能を明示し、元snapshot/描画版からの再構成を検証する専用回復へ。参照を現在基本情報へ付替えて成功にしない。

### 再印刷と訂正の同時操作

再印刷要求は現在認可とseries/current editionを解決し、対象editionとartifact digestへ固定。出力を渡す直前にもcurrent／有効性を再確認し、訂正新版が先にcommitしたなら旧版の有効出力を拒否して最新版へ再取得する。履歴readerと有効証明出力readerを分離し、旧訂正版のartifact URL直接指定・旧previewからの再出力でも同じgateを通す。private blobを無認可の恒久公開URLにしない。既に受け渡した紙・ダウンロードの回収や複製防止は保証しない。

発行後の共通住所電話／pet基本情報更新はeditionを自動変更せず、最新版の意味も「現在基本情報で再描画」にしない。Q-V2は既決「訂正前の版は有効な証明書として印刷しない」の適用として扱い、新たな本人承認待ちにしない。**接種訂正保存が成功した時点で、その訂正前の発行内容の有効出力を止める**。訂正保存失敗／rollback時まで旧版を無効にせず、原接種と対応する有効性を変更しない。

技術案は接種の現在版とeditionが参照する元版・訂正関係を有効性gateにする。訂正writeは既存共通gateとD08の取得順へ参加し、期待する接種版／必要なseries依存版・現在認可を再検査し、接種の新版・変更履歴・必要監査・成功receiptを同じtxでcommitする。旧版停止はこのcommit済み訂正事実に結び付け、非同期workerや新版発行完了を待たせない。未保存入力・訂正要求の受付・client予測だけで無効化しない。

preview／render／artifact直参照／再印刷の有効出力gateは、現在接種版とedition元版・訂正関係を同じ整合読取で照合する。訂正が先にcommitしていれば旧版を拒否し、新発行待ちと表示。出力認可が先行した場合に既に渡した出力を回収可能とは保証しないが、古い版を新しい要求へ再認可しない。競合・結果不明は版と同key結果を再取得し、版未取得を有効と推定しない。訂正成功後の新発行がrender／監査／応答で失敗しても、訂正前版を有効へ戻さず同じ新発行候補の結果を照合する。新発行commitでは元editionとの訂正関係と新current参照を同時保存し、旧内容履歴は保持する。

### 本人決定Q-V1と既決を適用したQ-V2

| 決定／原文と出典 | 現在の適用 |
|:---|:---|
| Q-V1：12:53 UTC本人の表示第3項Aで確定。[通常縦横の共有版](../../../spec/domains/vaccinations/lifecycle.md#certificate-shared-layout-edition-20261018) | 同じ通常証明書の最新版を共有し、印刷時に縦横を選択。固有項目は維持し、切替で別発行版を作らない。物理series ID／描画方式は技術提案 |
| Q-V2：既決「訂正前の版は有効な証明書として印刷しません」の技術適用 | 訂正保存成功時点で旧内容の有効出力停止。保存失敗／rollbackでは停止せず、競合時は現在接種版とedition元版・訂正関係を照合する。新たな本人承認・未回答候補として扱わない |

### 設計検証（未実行）

DC-V01：3様式の項目／配置・改頁・font、旧見本とのgolden比較。DC-V02：発行後基本情報変更で旧snapshot/digest不変。DC-V03：render失敗・upload失敗・監査失敗・commit応答断で未発行を成功表示せず同editionへ収束。DC-V04：2端末同時発行／訂正／再印刷、古いseries版・artifact直参照・認可失効で旧有効出力なし。DC-V05：物理印刷取消・同key再送で接種createや会計なし。DC-V06：訂正保存と旧版出力／新版発行を両先行順序で実行し、訂正commit後の旧版出力拒否、訂正rollback時の旧版有効性保持、古い期待版・元版不明の停止を検査。DC-V07：訂正commit後応答断・同key照会・新版render/監査失敗でも旧版復活なし。DC-V08：同じeditionを縦横交互に出力し、版増殖なし・最新版共有・固有欄と固定snapshot維持を検査。仕様採用・実装所在・runtime／法令適合は別判定。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../../../spec/domains/SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../../../spec/domains/vaccinations/lifecycle.md#certificate-editions-20261017)。既知の通常縦横・単一狂犬病見本は旧見本に従う。これで全要件完成とはせず、[RV4種類と再発行日](#certificate-mtg-technical-gap)の追加対応を必要とする。訂正後新発行・旧版履歴保持、再印刷は最新版、旧訂正版は有効印刷不可。共通基本情報更新で発行版は自動変更しない。旧印刷時自動登録等は復活させない。専用証明書実装の所在・適合と仕様採用を分ける。


<a id="vaccination-supplemental-read-trace"></a>
## 履歴順序・種類ラベルの技術接続

先行の明示trace29ACに下記4ACを加えた33ACと、[証明書6ACへの限定委譲](#certificate-edition-proposal)で現行39ACを追跡する。種類別名称ラベル色は対象一覧の表示契約であり、未承認の配色・色管理UIを選定しない。

| 正本AC | 既存契約への技術接続・確認点 | 業務正本 |
|:---|:---|:---|
| AC-vaccinations-history-order | 同じ保存履歴を接種日降順の履歴projectionと、次予定日昇順・未設定末尾の既定一覧projectionへ分ける。既存の指定ソートを優先し、並び替えから保存値を変更しない。 | [正本](../../../spec/domains/vaccinations/lifecycle.md) |
| AC-vaccinations-type-name-label-color | 対象の患者情報内一覧で既存種類identityと名称をラベルへ対応付ける。名称文字を残して色だけに依存させず、異なる種類の表示を比較する。具体色値は推測で固定しない。 | [正本](../../../spec/domains/vaccinations/lifecycle.md) |
| AC-vaccinations-type-color-warning-separation | 種類識別の表示と既存の期限等の警告状態を別の意味として投影し、種類色だけで警告を置換しない。今回の表示から新しい期限判定・警告生成を追加しない。 | [正本](../../../spec/domains/vaccinations/lifecycle.md) |
| AC-vaccinations-type-color-record-scope | 認可された同一患者・接種のreadだけに適用し、前後の名称・実施日・次予定・履歴を保持する。記録writeをせず、独立画面・LINE・証明書の表示へ適用範囲を拡張しない。 | [正本](../../../spec/domains/vaccinations/lifecycle.md) |

各行は既存ACの接続先を明確にするもので、新しい業務要件・実装完了・実機合格を表さない。具体的な既存wire／取得元との適合は実装工程で照合する。

<a id="vaccination-date-certificate-review-20261019"></a>
## CL-R11／12：両入口の日付・証明専用経路の確認境界

基準 `563353287ebb08af55056e3fbfba2aca3531e599`。カルテ内[追加hook](../../../../frontend/src/features/medical-records/hooks/use-medical-record-vaccination-form.ts)は初期1year候補、複製時の実施日clear、元実施日と現在間隔による再判定、other時のnextDate clearを持つ限定観察。独立[form model](../../../../frontend/src/features/vaccinations/hooks/use-vaccination-form-model.ts)は保存日／次回日を復元する別入口である。コード上の1yearや複製clearを新しい業務承認にしない。元実績を開く編集と、新しい接種を作る履歴複製は別caseで検査する。

実装適合の比較表には、入口kind、create／edit／duplicate、元接種ID／版、種類設定の有無・取得状態、保存済み当時日、明示手入力、実施日変更後候補、送信next_date／next_schedule_type、保存再読取結果を置く。設定なし旧行、手入力、過去当時日、元間隔不明、取得失敗を別fixtureにし、旧値を黙ってclear・1year補完しない期待を正本へ結ぶ。新規複製を既存実績の無変更編集と同じ期待にせず、追加の優先順位や具体intervalを推定しない。臨床間隔承認と日付validation実装は別判定。

CL-R12の限定検索はmedicalrecord route群とvaccinationsのroute／component範囲で証明書専用route／画面を確認できなかった。一般のexamination print-snapshotは証明書専用経路ではない。repository全体の不存在証明や専用実装済み判定はしない。

固定版案の実装引継ぎ表は、既知3見本の項目（RV4種類全体の充足ではない）→不変snapshot field→template版／field projection→golden fixture→出力receiptとする。通常縦横のLOT／E-Mail等の固有欄は別projectionで、snapshot欠落を現在共通情報から埋めない。初回発行、縦横切替、同版再印刷、基本情報変更、接種訂正commit、render／upload／監査失敗、応答断、artifact直参照の各行をDC-V01〜08へ対応付ける。訂正後新版render失敗でも旧版を復活させず、両順競合と物理印刷の記録を分離する。見本項目・配置を読取照合した証跡と、実描画golden／実印刷の証跡は別に必要で、今回後者は未実施。

### CL-R12の保存済み見本照合証跡の再利用

2026-10-19の文書再照合では、保存済みの独立原資料確認（元XLS SHA-256 `027572bcf6c83882b0d085872f05cd532d0ef79a6fb99bcb07fb97c390bf4e07`、3見本の全項目／配置・画像hash）を再利用し、[既存正本の見本対応表](../../../spec/domains/vaccinations/lifecycle.md#certificate-editions-20261017)と一致した。通常縦H127:K148、通常横F148:K158、狂犬病B235:F246は画像描画アンカーで、B162:B220の操作説明本文とは別。縦LOT.1〜4、横E-Mail、狂犬病に独立発行日欄がない差を保持する。項目／配置の文書照合は確認済みで、新しいpixel読解・描画golden・実印刷は実施していない。

repository内のGo／TS／TSX／SQL／YAML／YML／JSONを証明書の日本語・英語名で検索した範囲では、専用発行経路を特定できなかった。TLS設定・SMTP証明書test・生成マニュアル・運用証跡の一致を診療証明書実装と誤認しない。命名の異なる実装や未公開コード／実環境の不存在を証明する検索ではない。

## 10/19横断検索・同種操作の設計接続

[共通adapter](../../shared/integration-contracts.md#cross-ui-search-adapter-20261019)を本領域の[適用台帳](../../../spec/shared/interaction-scope.md)L/M/H/S consumerへ接続する。draft/applied・独立filter・取得scope/世代・復帰/失敗状態を分離し、検索ボタンだけでキーワードを適用。既存API検索と取得済み内絞込みの責務、ページング/上限、認可・保存方式を維持する。E/N操作へ部品単位で一括適用しない。実装/全caller/実機未確認。

<a id="plan-tab-link-design"></a>
## 治療プラン連動の設計受け手（2026-10-20）

業務目標は[D03共通FR／AC](../../../spec/domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-plan-domain-tabs)、本領域の[接続境界](../../../spec/domains/vaccinations/lifecycle.md#vaccinations-plan-tab-link)を参照する。一接種一記録、実接種日・lot・種類別設定／手入力の次回日と訂正履歴を維持する。治療プランの予定と接種済み実績を識別し、予定追加を接種記録作成・次回日の生成・証明書発行へ変換しない。既存の保存済み履歴を今回予定に名称一致で付け替えない。

[D03の現行観察と設計残件](../medical-records/technical-spec.md#plan-domain-tabs-target-gap)から関連の安定した識別・必要版と結果照合を設計する。表示のための物理資源作成を必須とせず、projectionの選定も未了。各ownerの保存単位・訂正／監査・再送保護を維持し、全owner一括transactionや一律1対1を新たな決定としない。実装適合／実機検証は未実施。

<a id="certificate-fee-projection-20261020"></a>
## EMR-518：料金projectionと発行版の接続

[料金4欄の業務正本](../../../spec/domains/vaccinations/lifecycle.md#certificate-fees-manual-20261020)を既存[edition/renderer設計](#certificate-edition-proposal)へ接続する採用方向。発行前の手動編集・確認額をcertificate edition snapshotへ固定し、会計/master/packageへ逆writeしない。optional/null・0・未入力を維持する。新規発行時の料金4欄の初期値はすべてnullとし、会計・master・packageから転記しない。発行前のスタッフ入力値をedition snapshotへ固定する。

`certificate_fee_role` は証明書adapterの任意mappingであり、一般packageの必須属性にしない。医院別構成と既存会計候補への明示追加を使用し、一律イベントやrule engineを新設しない。packageの品目合算と証明書の手動確認額（合計欄も含む）は別責務。会計一致gate・自動同期・合計再計算は追加せず、preview/issue/reprintから料金を生成しない。

具体edition DTO/保存列/rendererはEMR-486/487担当の最終契約へ接続し、本文書の追補を実装済みとしない。[package設計](../medical-records/treatment-packages.md#billing-certificate-adapter)と[未完の固定作業](../../../work/emr/emr-518-treatment-packages-20261020.md)を参照。受入条件は業務正本から辿り、実行はNOT_RUN。

<a id="certificate-mtg-technical-gap"></a>
## 10-21追加要件：RV4種類・一括出力・再発行日の未収録差分

[業務FRと9AC](../../../spec/domains/vaccinations/lifecycle.md#fr-vaccinations-rv-batch-20261021)を採用し、[現行v1選定契約と追加差分](../../../work/emr/emr-486-certificate-api-contract-20261020.md#certificate-mtg-contract-gap)を分ける。従来の証明書6AC／DC-V01〜08は既知見本の限定範囲で、追加要件の完了根拠ではない。既知3見本のreader／rendererの部分実装があるが、4種類のmaster・batch・変更可能な再発行日は未収録。試験NOT_RUN。

| 必要技術差分 | 要求される不変条件／設計残件 |
|:---|:---|
| master／snapshot | 4種類の安定したidentity、帳票セットと構成版、各template／用紙との対応を保持。通常縦横の同版表示と区別し、診療項目packageへ混ぜない。4名称・見本未受領、DDL・管理API・画面は未選定 |
| batch／wire | 4種類の構成照合と出力要求、各種類の結果・版／artifact対応、欠損／重複／一部失敗／unknownを表現。現在の単一kind/layout APIを4種類一括対応済みとしない。batchのcommit／回復単位、同key照合とreceipt拡張は後続選定。4種類=4ページとしない |
| 再発行wire／版 | 当初の発行日由来refと選択した表示日を確認内容へ固定し、新版のsnapshot／hash／preview／CAS／監査へ反映。同版reprint DTOへ未定義fieldを送らない。初回の任意日許可へ拡張せず、再発行のdate validationは別設計 |
| PDF storage／履歴 | 種類・構成版・edition・template／renderer版・digestとimmutable artifactの対応を追跡。旧artifactを上書きせず、失敗候補を有効出力へ公開しない。storage結合とbatchの出力形式は未選定。結合PDF／ZIP等を推測で固定しない |
| 時刻と認可 | 表示発行日は選択した日、issued_at／recorded_atはサーバー操作時刻。再印刷・再発行・batch各要求で現在scope／grant・患者・接種版・有効版を再検査。応答不明は元結果を照合し成功分を再発行しない |

履歴参照と有効出力を分離する既決を維持し、再発行で旧訂正版の有効出力を復活させない。master改定で過去の構成・版やartifactを書き換えず、資料待ちを推測実装で埋めない。実装コード・api.yaml・DDLは本追補で変更しない。
