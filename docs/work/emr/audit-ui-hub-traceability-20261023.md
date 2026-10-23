# 調査・ハブ・横断UI 6票の静的トレーサビリティ監査

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-ui-hub-traceability-20261023 |
| 種別 | 固定mainの限定静的監査受領記録 |
| 決定状態 | 既決仕様を継承。新しい業務・技術契約を採択しない |
| 実装状態 | 下記の限定接点の分類。全FR/AC・全source監査完了ではない |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

本稿は提供された監査報告を整形したもので、文書編集時に全sourceを独立再監査した記録ではない。[今回の索引](emr-audit-ready-handoff-index-20261023.md#audit-ready-additional-scope)を参照。

基準: main `77c28fbfbafa257dcca2421af99ff72da6889717`、2026-10-23。対象 EMR-372/373/403/404/412/450。最新本文・全コメントを直接取得（順に5/14/18/16/17/8、計78件、全next_page_results=false）。実ブラウザ、試験、DB、HTTP、Mac操作、製品編集、外部投稿、新PRなし。全runtime **NOT_RUN**。

<a id="ui-hub-traceability-conclusions"></a>
## 要点

- 新規製品欠陥の確定は0件。既知残件を元ownerへ保持する。6票の全FR/AC・全画面完了を保証する監査ではない。
- 372の7所見は訂正を含む対応表あり。IU-004「ring欠落」は撤回しIU-002へ統合済み。現在ソースに採用修正を確認。元本文だけから7件の現存欠陥と数えない。
- 373の在庫211行はID重複なし。画像索引551件もファイル名重複なし。ただし両者は別母集団で、画像数を画面完成率へ換算できない。古いFAIL/UNKNOWN_TARGETと後続追試が混在する履歴索引。
- 403の21 ID、404のC6 6 IDは互いに交差しない責務として維持。403 C5-4 staff dirtyは#702後の成功ID保持・dirty解除経路が現mainにあり、旧FAILだけで再修正しない。全variant受入は別。
- 412 N4は「ui Badgeのまま」ではない。ただしStatusBadgeへのcaller側text-xs/h-5等のoverrideが残るため、既定寸法へ統一済みとも言えない。既存N4の寸法疑義を具体化し、現在の実寸・是非の受入を残す。
- 450は検索72用途（重複IDなし）と非検索D01–D23/P01の24 familyを持つ。これは計画の所在確認で、72実装合格ではない。P01は新Ready553へ一元化、535は2Select解除の別成果、554はDBOrTxの別成果。

<a id="ui-hub-traceability-evidence-depth"></a>
## 出典と保証深度

本記録は以下の固定sourceリンク、PR採用情報、対象範囲・件数に限定する。主採用PRは [660](https://github.com/MinoruSoga/AnimalEkarte/pull/660)、[665](https://github.com/MinoruSoga/AnimalEkarte/pull/665)、[666](https://github.com/MinoruSoga/AnimalEkarte/pull/666)、[679](https://github.com/MinoruSoga/AnimalEkarte/pull/679)、[702](https://github.com/MinoruSoga/AnimalEkarte/pull/702)、[810](https://github.com/MinoruSoga/AnimalEkarte/pull/810) を再取得しmerged=trueを確認。#662は既存audit-shared-ui/audit-ui-layout監査の採用差分・現main確認を再利用。#698はstaging同期で製品修正のownerではない。#810や#802–805の文書採用を新製品修正へ換算しない。

以下sourceリンク接頭辞は https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/ 。過去の局所PASS/画像は当時の証拠で、今回の独立実行結果ではない。

<a id="ui-hub-traceability-372"></a>
## EMR-372: 訂正済み所見と現在コード

正本成果物: [訂正・対応表](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-372-design-audit-resolution.md)。#660 merge `422f0d8058d44521be516065cd2b5b507cf90a63`。

- IU-001: OwnerAccountingHistoryParts.tsx:53 のcompletedTotalはfont-semibold。会計値や集計の再監査ではない。
- IU-002/004: ui/textarea.tsx:13にfocus-visible ring-2/offset-1、CharCountTextareaはSTYLE.textarea利用、LineSendPanel.tsx:155–160の生textareaも同STYLEを利用。共通tokensは既存shared-ui監査sourceを再利用。IU-004を独立「未修正ring欠落」と再計上しない。
- IU-003: MedicalRecordExamination.tsx:114/119、MedicalRecordTabsShared.tsx:70、DiagnosisHeaderChiefComplaint.tsx:33の指定空/読込/利用条件はC.text60。
- IU-005: MedicineDoseParamsEditor.tsx:101/105、StaffBasicInfoSection.tsx:137、LabDeviceItemMasterSidePanel.tsx:215、ReservationTypeSidePanel.tsx:312の指定説明はC.text60。全text40の無差別除去は不要。
- IU-006: ShiftTemplateSidePanelFields.tsx:204「休憩なし」はC.text60。
- IU-007: HelpChat.tsx:330/369/451/463の4ボタンはdisabled:opacity-50＋pointer-events-none。対象は常駐supportでChatHistoryPageではない。

残受入: 373へ実focus/contrast/表示を引継ぐ。変更パスの存在・修正保持は限定確認したが全caller、背景色ごとの実測、全a11yをPASSにしない。372本文の古い「監査未実施」「未確認なし」は履歴文脈で扱う。

<a id="ui-hub-traceability-373"></a>
## EMR-373: 在庫・画像・子票の網羅境界

[在庫](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-373-visual-audit-inventory.md)は211 unique INV行。[追試記録](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-373-visual-audit-results.md)と[画像manifest](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-373-images/index.json)は551 uniqueファイル。初回requested390→actual500、誤ったowners/edit経路、wave23/29対象未表示を訂正・UNKNOWN_TARGETで保持している。後続画像を前の誤画像へ差し替えて全状態PASSとしない。

- VQ-001〜012 → 377〜388の12子。追加389〜394、396〜399を別所見・状態として追跡。395が連番にないことだけで欠落と認定しない。
- 383/396/397/398は既存shared-ui監査、381/387/390〜394は既存layout監査を再利用。修正実装・実接続と現在実画面受入を分離。
- 387の衝突は432受入後に残る場合401 C3-1、391の共有患者カード後続は412へ。397の表共通ownerは373、398の期限/数量fixture不足は元398。重複新票不要。
- 399の768担当医見出しFAILは追試記録末尾の歴史状態。今回399の最新票・現source未取得なので「現在もFAIL」とは言わない。
- 元画像/ZIPバイナリ・SHA256再計算・Plane添付閲覧可能性は未確認。manifestの存在は画像再レビュー・添付確認ではない。

残作業のownerは373のまま。211行それぞれのcurrent SHA×mode×width×role×data状態に既存子または理由付きUNKNOWNを紐づける必要があり、本監査はその全行の再成立まで証明していない。

<a id="ui-hub-traceability-403"></a>
## EMR-403/404: 21残余と6状態分類を混ぜない

403の母集団はC1-2/3/4、C2-1/2/3、C3-1、C4-1/2/3/4、C5-1/2/3/4、C7-6、C8-1/4/6/7/8の21行。本段落の21行を母集団としprimary owner=403、現在未受入を保持する。404はC6-1〜6の6行を独立列挙。親401全43指摘・税405・全実API受入406を本21+6へ押し込まない。

403 C5-4の採用経路はStaffSettings.tsx:204→StaffSidePanel.tsx:184–212→use-master-save。CREATE成功IDを保持し、mount/version条件に合致して全必要保存が成功したときだけdirty解除とcloseする。#702 merge `dd1b5a76cdae42c03e55417f38c04083b6fa086f` が存在し、10/15のdirty FAILを現欠陥として再起票しない。成功IDを観測できない曖昧な通信断、実API、再open全variant、readonly/nested/focusは未確認。

404の限定履歴sampleはC6-1のcheckups 0件/見積403・404、C6-5の合成権限拒否とstaff grant-loss、C6-6のcampaign delayed200。C6-2警告/agent503、C6-3無効担当/死亡履歴、C6-4無料0円対不完全明細はUNKNOWNを維持。CTA enabledや0円だけで製品欠陥としない。

重要な追跡上の残差: コメントが参照するresearch-coverage/fixture-empty-state-matrixの正確な公開path・現在内容は今回取得できていない。最新本文+全コメントに所在・完了未達の説明はあるが、21+6行ごとの完全な分類/根拠/次action表が現成果物で満たされるとは保証できない。孤立と断定せず、既存403/404に原manifestの場所と版を固定する作業を残す。実表示が禁止でもこの証拠整理は可能。

<a id="ui-hub-traceability-412"></a>
## EMR-412: 採用範囲とN4の具体的残差

#679 merge `82c75ee7a214a1efa0f036285c07ca2e95d6845f`。X1〜X15、K1/K3〜K9、T2、N1〜N8の指摘空間を維持。K2/T1の連番欠けは本文にないので新規漏れ扱いしない。X12任意、K9明示非変更、N8任意を必須修正へ昇格しない。13 routeのM/R/H経路索引は票内に存在するが、今回全経路×指摘IDのsource再精査は未完。

N4: [HospitalizationBoard.tsx:205–211](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/hospitalization/components/HospitalizationBoard.tsx#L205-L211)はStatusBadgeを使用。だがclassNameはH_STYLES.text.xs（styles.tsでtext-xs leading-tight）＋px-1.5 py-0 h-5 border-none。StatusBadge.tsx:10–17の既定text-2xs/px-2/h-7/borderより後にcaller指定を合成し、ui/badge.tsx:17でcnを通す。したがって「部品未移行」は否定できるが、「共通既定12px/28pxに寸法統一済み」は支持できない。旧14px/20pxという観測を現実測へ転用せず、既存412 N4でoverride保持の妥当性と実寸受入を決着させる。新たなUI全体改修票は不要。

X13はPatientInfoCard内包採用、X5はshared-ui/layout監査の保険折返しを再利用。X7/K5の古いfinalized一律ロック文言は10/20優先追補に従い494/495の現在の診療終了・目的別確認・会計確定・編集epochと分離する。色修正から業務拒否を復活させない。

X6トリミング死亡表示とallergy_infoは旧本文が本票外と記録する参考事項。今回の範囲で最新の専有子・採用業務契約への完全な照合は未実施。所在不明の新欠陥を宣言するのでなく、412索引から該当業務ownerへ参照を確定すべき残差として保持する。

<a id="ui-hub-traceability-450"></a>
## EMR-450: 成果物在庫と有限残件

[③計画](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-450-stage3-common-ui-start-plan-20261021.md)、[非検索操作](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-450-stage3-common-ui-start-plan-20261021-operations.md)、master consumer台帳を取得。72用途はL12+M26+S13+H6+N7+E7+N-PKG1。D01–D23/P01の24familyとは別軸で、足して96機能完成とはしない。S13の未来入口、H03二入口、N/Eの異用途除外を維持。

- 453既採用: PropertyFilter.tsx:95–107はbutton commit、clearはdraftだけ。旧本文のclear即適用/入力Enter検索を現未実装としない。全consumerのscope/raceまで合格ではない。
- H05: CheckupForm.tsx:50–59/146–151はhistory loading/dataだけを受け渡し、error/refetch未伝搬を現sourceで確認。0件と読取失敗の接続残は450。500の④receiptへ移さない。
- P01: MasterPageShell.tsx:65直接navigate、TreatmentPackageSettingsのdirty guardは行切替/close側だけという既知残。Ready553→523/450の単一writerがshell任意onBack+package caller+fixtureを所有。450に別実装票を作らない。
- SubmitButton: Form/SubmitButton.tsx:32–48のpendingはdisabledとloadingText置換だけ。Button.loadingへの自動結線は無く、既知227→450 wrapper残。callerがpropsでloadingを与える可能性とwrapper自動接続欠如を区別。
- G1 M09=450/492、G2 M20=450/448/493、G3 D09 dirty=450/500、治療内容Enter caller=450。今回G1/G2/Enterの現source全経路は再精査せず、採用計画内の既知残としてのみ記録。G3はCheckupForm入口を読んだがuseBackNavigation内部等を未精査なので現欠陥再認定はしない。
- G4/D15、L12/E05/N05配送監視マーケ部分はPhase1最後という明示保留。検査受信N05と分離し、③先行必須gateへ戻さない。
- 535は528入院プランの体格/料金単位2Select解除UI。検索clearや共通primitive全面改修ではない。554は521/455のDBOrTx22件inventory/fixtureで③UI完了と無関係。3票ともReady=計画成立、製品完了/実行許可/試験PASSではない。

<a id="ui-hub-traceability-unverified-scope"></a>
## 明示する未精査範囲

全211在庫の現在sourceと実状態、403全21×route/variant、404全6×role/API、412全M/R/H×ID、450全72 consumerと24 operationの全FR/AC・推移caller、画像本体再レビュー、実配備版との一致は未精査。孤立残件がゼロ/完全MECEとは主張しない。今回の価値は旧欠陥の誤再計上を防ぎ、残りの有限ownerと証拠不足を明示したことに限る。

<a id="ui-hub-traceability-source-location-followup"></a>
## 14:50 UTC 出典所在の限定追補

403/404の既存comment_html全件のhrefを再抽出した。公開リンクはPR665/666/669/698/702で、research-coverage/fixture-empty-state-matrixへのhrefはない。現在mainのGitHub recursive treeは8,131 entries、truncated=false。`research-coverage`、`fixture-empty-state-matrix`、`reports/uat-2026-10-15`に一致するパスは0件。既存common-ui-full-coverage-plan/common-ui-coverage-audit/plane-final-stage123記録とspec-coverage系、旧treeも限定検索したが、正確な公開パスと版は回復できなかった。これは全history/別Artifacts repository/元ローカル保管先での不存在を意味しない。残依存は403/404の元調査artifact保有者から原matrixの保管先・固定版を受領し、公開根拠へ接続すること。研究表の不存在や新製品欠陥とは断定しない。

412 allergyについて一部owner経路を回復した。[375](https://app.plane.so/baritechllc/browse/EMR-375/)の最新10/21優先追補は旧inquiry列（履歴本文でallergy_infoを明記）を列別decision matrixにし、問診構造は[496](https://app.plane.so/baritechllc/browse/EMR-496/)、飼主原文は[477](https://app.plane.so/baritechllc/browse/EMR-477/)へ接続すると規定する。496の現本文は初診問診のreader/UI/provenanceを所有し、[initial-intake正本](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/spec/domains/medical-records/pages/initial-intake.md)には投与後異常/アレルギー等の自己申告とAC-medical-records-intake-reaction-detailがある。ただし旧allergy_infoの全値移行・表示と、412参考所見の明示受領・完了の証拠まではない。375→496は列別整理/構造化問診の既存責務、477は別の原文readerであり、477へアレルギー編集を割り当てない。

412 X6のトリミング一覧死亡表示について、412本文・17コメント・spec-coverage系の再検索に明示の専有receiverは見つからない。trimming lifecycleのAC-trimming-deceased-loading/historyは保存拒否・生前分専用訂正の契約であり、一覧バッジBE項目の実装ownerを示す証拠ではない。404 C6-3の死亡履歴も同じ実装票とは扱わない。専有receiverの確定は412取りまとめ側の既存残依存として維持し、推定で480/491等へ割り当てない。

この追補でも試験/実機/外部投稿は0。公開出典の回復は部分的で、未解消を残したまま終了する。
