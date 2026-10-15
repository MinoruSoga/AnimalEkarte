# domains — フォルダの説明


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
| D15 | [CRM・Lステップ](../line/lstep-integration.md) | [技術契約](../line/lstep-integration.md) | 17トリガー・Webhook／タグ・対象同期／CSV・配信・結果不明・送信用ファイル |
| D16 | [経営・顧客集計](analytics/business-rules.md) | [技術契約](../../design/domains/finance-operations/technical-spec.md) | 請求と実金銭・期間／母数／記録日数・配信分析と相関の意味 |
| D17 | [医院・設定入口](clinic-settings/README.md) | [技術契約](../../design/domains/clinic-settings/technical-spec.md) | 医院設定・所属scope／法人設定・設定入口と個別master責務 |
| D18 | [LINE予約チャネル](../screens/28-line-reservation.md) | [技術契約](../screens/28-line-reservation.md) | 院内設定／editor／枠と飼主appの予約・変更／取消・本人認証／通知 |
| D19 | [同一飼主・ペット連携](../screens/40-identity-links.md) | [技術契約](../screens/40-identity-links.md) | 既存IDの手動group連携／解除・同一集合／親子／残数と履歴 |
| D20 | [LIFF診察券・健康情報](../screens/38-liff-pet-health.md) | [技術契約](../screens/38-liff-pet-health.md) | 本人紐付け／単回token・診察券／保存済接種／最終来院の限定表示 |
| D21 | [飼主カルテレポート](../screens/39-owner-report.md) | [技術契約](../screens/39-owner-report.md) | スタッフ専用・section別読取認可／失敗縮退・患者context・印刷共有なし |
| D22 | [院内マニュアル](internal-manual/README.md) | [技術契約](../../design/domains/internal-manual/technical-spec.md) | 公開最新版のスタッフ閲覧／検索／印刷と認可付き編集・公開／履歴復元 |
| D23 | [サポート](support/README.md) | [技術契約](../../design/domains/support/technical-spec.md) | 報告／共有READ／報告医院mutation・個人history・確認付き外部通知と結果分離 |

D18の飼主側操作は[別app37](../screens/37-line-reserve-owner-flow.md)、D20は別app38。main88ルートへ重複加算しない。全ownerの[共通契約](../shared/README.md)と[8接続の照合](../workflows/reservation-to-record-flow.md#cross-domain-review-20261015)を併読する。


後続の6点（LINE開始時刻・指名なし・健診次回日・過去接種の当時次回日・トリミング途中保存・停止支払方法）は[承認と文書候補の状態](SPECIFICATION_PROGRESS.md#additional-business-decisions-six-followup)で管理する。通常診療の営業時間からの生成・トリミングの登録時刻限定と、変更可能・担当者別設定の追加回答を含み、未承認の枠数等は補わない。

## 2026-10-15会議合意の接続：shifts-8

期間・スタッフによる勤務複製とExcel作成CSV取込、差分・エラー・予約影響確認を[シフト正本](../screens/24-shift-calendar.md#shift-copy-csv-agreement-20261015)へ接続。templateと勤務を分離し、既存勤務の無断上書き禁止と未来・現在診療中の既指名予約保護を維持。

## 2026-10-15会議4点の追加整理

[D03最新通常編集・自動保存](medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)、[D02開始／終了導線](reception/pages/reception-board.md#reception-meeting-start-end)、[D06施術・料金](trimming/lifecycle.md)、[D14複製／CSV](../screens/24-shift-calendar.md#shift-copy-csv-agreement-20261015)の文書差分を統合・独立レビュー済み。D07〜09、D10／11／13、D20／21の別状態・履歴・副作用保護も照合した。今回合意の文書反映と全過去業務承認の網羅を区別し、現在の状態は[全体進捗](SPECIFICATION_PROGRESS.md)へ。
