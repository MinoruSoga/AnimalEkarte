# domains — フォルダの説明

2026-10-17追加：ユーザー確定9群の[正本・受入条件・C01〜C13訂正対照](SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)。問診統合・共通患者・医師確認／会計同期等はこの最新参照に従う。初回反映時の未公開/ontology保留という段階は履歴。後続承認により業務69文書とontology11文書をDraft PR #670のcommit `7e4171725c24c5be9807ddcf8c09b323aca31442`へ公開し、公開時の対象本文一致を確認済み。現在の語彙は[ontologyの適用境界](../ontology/README.md#ontology-current-20261018)へ接続する。以後のローカル改善差分は公開済み版と区別し、実装・実機未確認と除外原資料の未読は維持する。

2026-10-20〜21追加：[本人回答・採用の索引](SPECIFICATION_PROGRESS.md#decisions-20261020-21)。請求日、出庫確定、営業時間・休診日の共通基準、死亡誤登録取消、主飼主変更時の未確定会計、スクリーンショット、入院退院の会計受け手、次回予定日初期値、プラン↔3タブ、EMR-518パッケージ、証明書のRV4種類・再発行日・A4横・適格性と入力欄、入院の税未確認、再診問診の状態別可否、死亡登録時の未来予約取消、月次売上の主表示はこの索引から各正本へ辿る。実装・実機は未確認。

最新差分は[追加採用5点](SPECIFICATION_PROGRESS.md#additional-adopted-five-20261015)。採用とdocs公開は承認済み。固定35文書の統合検査・独立内容／構造レビューで重大指摘なしを確認し、限定文書範囲を再整合した。公開先の本文一致確認は別段階で行う。

前段階の[共同確認batch](SPECIFICATION_PROGRESS.md#joint-walkthrough-batch-20261015)は、台帳／DM・受付・診察・見積／締めの回答済み範囲を文書へ反映し、独立レビュー済み。全業務の共同確認は継続中で、以下の過去の文書整合完了とは別に扱う。公開後の本文一致は別段階で確認する。


## このフォルダの役割と使い方

業務領域ごとの仕様入口をまとめる。

今回の全体範囲・領域数・完了境界と残件は[仕様整理の全体進捗](SPECIFICATION_PROGRESS.md)で管理する。実装は今回の対象外。

領域READMEに業務ルール・受入条件・未決の所在を示し、pagesに画面詳細、必要な保存契約を領域内に置く。複数領域のフロー、共通語彙、共通マスタの本文を複製しない。

領域別の正本は以下の一覧から参照します。別アプリや画面内に正本を置く領域も含み、フォルダ数を業務数へ読み替えません。

主要文書：各領域READMEが、その領域の主要文書・受入条件の入口です。

読む順：[全体要件](../specification.md) → [予約](reservations/README.md) → [受付](reception/README.md) → [カルテ](medical-records/README.md) → [会計・未納](accounting/README.md)。横断する条件は[workflows](../workflows/README.md)、共通概念は[ontology](../ontology/README.md)・[shared](../shared/README.md)。

正本・状態：[決定状態と検証状態の規約](../authoring-guide.md#status)に従い、各節・FRの承認根拠と実装／実機状態を別に読む。READMEへの掲載・フォルダ配置は承認や実装済みを表さず、対象外の既存仕様を新たに確定しない。既存の別名入口・索引の責務は維持し、本書は案内を補う。


## 全23業務領域の正本

2026-10-15の全体整理。表は読む順と責務の索引で、詳細・受入条件は正本本文へ。進捗は[全体管理](SPECIFICATION_PROGRESS.md)へ。

| ID | 領域・業務正本 | 技術契約 | 主責務の範囲 |
|:---|:---|:---|:---|
| D01 | [予約](reservations/README.md) | [技術契約](../../design/domains/reservations/technical-spec.md) | 通常予約・ホテル／入院・施術の分類、予定枠・指名・受付／取消・同一操作結果 |
| D02 | [受付・来院](reception/README.md) | [技術契約](../../design/domains/reception/technical-spec.md) | 来院受付・再受付・準備／読込／開始と通常記録の同時成功 |
| D03 | [通常カルテ](medical-records/README.md) | [技術契約](../../design/domains/medical-records/technical-spec.md) | 保存§2・目的別医師確認・終了後通常編集・タブ開始時クリア／履歴複製取消・終了時警告・終了履歴・保存版／追記保持・子記録・臨床master |
| D04 | [飼主・ペット台帳](owners-pets/lifecycle.md) | [技術契約](../../design/domains/owners-pets/technical-spec.md) | 主副飼主・患者ID、検索・同時登録・所有変更と過去帰属／危険情報保護 |
| D05 | [入院・ホテル](hospitalization/lifecycle.md) | [技術契約](../../design/domains/hospitalization/technical-spec.md) | 滞在／予定実績・担当・ケージ／頭数／延泊、個別料金・退院会計・原本保持訂正 |
| D06 | [トリミング](trimming/lifecycle.md) | [技術契約](../../design/domains/trimming/technical-spec.md) | 施術と予約・明細／コース／画像・担当・完了／訂正・会計受渡し |
| D07 | [検査・機器受信](examinations/lifecycle.md) | [技術契約](../../design/domains/examinations/technical-spec.md) | 検査／結果版・機器job／対応付け／補償・確定後結果と印刷・検査master |
| D08 | [予防接種](vaccinations/lifecycle.md) | [技術契約](../../design/domains/vaccinations/technical-spec.md) | 独立／カルテ内の同じ接種資源・次回予定・担当・訂正と限定健康表示 |
| D09 | [健診](checkups/lifecycle.md) | [技術契約](../../design/domains/checkups/technical-spec.md) | Checkup正系統・項目／結果・段階保存／置換・記録版とスタッフ読取 |
| D10 | [会計・未納](accounting/lifecycle.md) | [技術契約](../../design/domains/accounting/technical-spec.md) | 請求／実授受／預り金の分離、計算・保険・訂正／再請求・帳票・マスタ |
| D11 | [見積](estimates/business-rules.md) | [技術契約](../../design/domains/finance-operations/technical-spec.md) | 入口共通明細・共通計算・確認転記・原本保持後継訂正 |
| D12 | [レジ締め](cash-register/business-rules.md) | [技術契約](../../design/domains/finance-operations/technical-spec.md) | 営業日／実査・実授受／開始残／補充回収・原締め保持と差分 |
| D13 | [在庫](inventory/business-rules.md) | [技術契約](../../design/domains/finance-operations/technical-spec.md) | 実物の入出庫・返却／棚卸と履歴・小数臨床量と整数在庫の境界 |
| D14 | [職員・シフト](staff-workforce/README.md) | [技術契約](../../design/domains/staff-workforce/technical-spec.md) | 職員／職種／勤務・休憩／template、未来・未終了予約の保護 |
| D15 | [CRM・Lステップ](../line/lstep-integration.md) | [既存技術契約](../line/lstep-integration.md)・[共有ファイル保持](../../design/domains/crm/README.md) | 17トリガー・Webhook／タグ・対象同期／CSV・配信・結果不明・送信用ファイル |
| D16 | [経営・顧客集計](analytics/business-rules.md) | [技術契約](../../design/domains/finance-operations/technical-spec.md) | 請求と実金銭・期間／母数／記録日数・配信分析と相関の意味 |
| D17 | [医院・設定入口](clinic-settings/README.md) | [技術契約](../../design/domains/clinic-settings/technical-spec.md) | 医院設定・所属scope／法人設定・設定入口と個別master責務 |
| D18 | [LINE予約チャネル](../screens/28-line-reservation.md) | [技術契約](../screens/28-line-reservation.md) | 院内設定／editor／枠と飼主appの予約・変更／取消・本人認証／通知 |
| D19 | [同一飼主・ペット連携](../screens/40-identity-links.md) | [技術契約](../screens/40-identity-links.md) | 既存IDの手動group連携／解除・同一集合／親子／残数と履歴 |
| D20 | [LIFF診察券・健康情報](../screens/38-liff-pet-health.md) | [技術契約](../screens/38-liff-pet-health.md) | 本人紐付け／単回token・診察券／保存済接種／最終来院の限定表示 |
| D21 | [飼主カルテレポート](../screens/39-owner-report.md) | [技術契約](../screens/39-owner-report.md) | スタッフ専用・section別読取認可／失敗縮退・患者context。一般履歴は印刷共有なし、今回採用の予防薬専用2帳票のみ印刷対象 |
| D22 | [院内マニュアル](internal-manual/README.md) | [技術契約](../../design/domains/internal-manual/technical-spec.md) | 公開最新版のスタッフ閲覧／検索／印刷と認可付き編集・公開／履歴復元 |
| D23 | [サポート](support/README.md) | [技術契約](../../design/domains/support/technical-spec.md) | 報告／共有READ／報告医院mutation・個人history・確認付き外部通知と結果分離 |

D18の飼主側操作は[別app37](../screens/37-line-reserve-owner-flow.md)、D20は別app38。main88ルートへ重複加算しない。全ownerの[共通契約](../shared/README.md)と[8接続の照合](../workflows/reservation-to-record-flow.md#cross-domain-review-20261015)を併読する。


後続の6点（LINE開始時刻・指名なし・健診次回日・過去接種の当時次回日・トリミング途中保存・停止支払方法）は[承認と文書候補の状態](SPECIFICATION_PROGRESS.md#additional-business-decisions-six-followup)で管理する。通常診療の営業時間からの生成・トリミングの登録時刻限定と、変更可能・担当者別設定の追加回答を含み、未承認の枠数等は補わない。

## 2026-10-15会議合意の接続：shifts-8

期間・スタッフによる勤務複製とExcel作成CSV取込、差分・エラー・予約影響確認を[シフト正本](../screens/24-shift-calendar.md#shift-copy-csv-agreement-20261015)へ接続。templateと勤務を分離し、既存勤務の無断上書き禁止と未来・現在診療中の既指名予約保護を維持。

## 2026-10-15会議4点の追加整理

[D03最新通常編集・自動保存](medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)、[D02開始／終了導線](reception/pages/reception-board.md#reception-meeting-start-end)、[D06施術・料金](trimming/lifecycle.md)、[D14複製／CSV](../screens/24-shift-calendar.md#shift-copy-csv-agreement-20261015)の文書差分を統合・独立レビュー済み。D07〜09、D10／11／13、D20／21の別状態・履歴・副作用保護も照合した。今回合意の文書反映と全過去業務承認の網羅を区別し、現在の状態は[全体進捗](SPECIFICATION_PROGRESS.md)へ。


## 全域の最終同期

[全体の有限完了確認](SPECIFICATION_PROGRESS.md#all-domain-closure-20261015)に、領域別確認の範囲と最後のD06終了条件・旧親確定制限の同期を記録した。D06は[実担当と記録済み施術内容](trimming/lifecycle.md#trimming-end-recorded-performance)を終了時に確認し、コースなしの既存オプション明細を認める。担当／コース未定の途中保存は維持。最新差分の独立内容・構造レビューは重大指摘0で文書範囲の整理を完了。公開一致は反映記録で確認し、実装状態や全過去承認証跡の網羅とは分ける。

## D06の後続合意：実施なし中止

[D06実施なし中止](trimming/lifecycle.md#requirement-fr-trimming-zero-performance-abort)を通常終了条件とは別に接続する。中止理由・実施なしとスタッフの請求判断を保持し、請求ありはD10へ、請求なしは当該中止の新規請求なしで閉じる。D02表示とD10の既存金銭・他診療費の保護を併読する。今回の追加は全業務共同確認完了や実装済みの宣言ではない。

## 追加採用5点の文書整理（2026-10-15）

[追加採用5点の進捗](SPECIFICATION_PROGRESS.md#additional-adopted-five-20261015)では、品目マスタの保険対象設定、検査項目の選択印刷、スタッフ患者トップの感染症結果、初診問診の画面項目、既存レポート入口での予防薬専用処方箋・年間投薬表を整理する。業務回答は採用済み、本文は独立レビューと限定静的検査済み。公開先の本文一致は別検査とする。2026-10-17に問診を治療プランへ統合する決定で旧別タブ維持を置換した。旧資料の他の要望は自動採用しない。既存版／金銭／臨床原本とLINE公開範囲は各ownerの保護を保持する。
