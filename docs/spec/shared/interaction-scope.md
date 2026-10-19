# 同種操作・検索の適用台帳（C01）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | shared.interaction-scope／共通契約の適用台帳 |
| 決定状態 | 2026-10-19 05:52:03 UTC本人承認。根拠と本文は[共通正本](ui-contracts.md#interaction-consistency-20261019) |
| 実装状態・実機検証 | 未確認・全行未実機。文書棚卸しと限定静的観察は実装適合の証明ではない |
| 基準・確認日 | PR #726 head ff67aab27を基準、2026-10-19 |

本台帳は適用先を一意に分類し、挙動を再定義しない。L=業務一覧、M=管理マスタ一覧、H=保存済み業務履歴一覧、S=飼主/ペット選択。各行の既存キーワードは[明示実行FRと8AC](ui-contracts.md#requirement-fr-common-ui-explicit-business-search)および旧4ACを適用する。N=キーワードなし、E=異用途。前者に新しいキーワード欄を設けず、後者の発火は今回一括変更しない。同じ画面に複数用途がある場合は欄ごとに分ける。全入力/保存の自動保存方式の変更ではない。

## 対象一覧と選択検索（全行未実行）

| 適用ID | 画面・用途／正本 | 維持する検索対象・取得範囲・境界 | 受入 |
|:---|:---|:---|:---|
| L01 | [飼主・ペット](../screens/03-owners-list.md) | サーバ20件、既存AND/カナ/No/医院・死亡条件 | 横断8AC＋既存画面AC、未実行 |
| L02 | [カルテ](../domains/medical-records/pages/medical-record-list.md) | サーバ、主訴/名前/カルテNo等の既存述語。病名/治療名を一覧へ追加しない | 横断8AC＋既存画面AC、未実行 |
| L03 | [入院・ホテル](../screens/07-hospitalization-list.md) | 既存リストpage/ボード走査を維持。表示絞込みを占有判定に使わない | 横断8AC＋既存画面AC、未実行 |
| L04 | [検査](../screens/12-examinations-list.md) | 日付はサーバ、全文/状態等は現在ページ内 | 横断8AC＋既存画面AC、未実行 |
| L05 | [予防接種](../screens/14-vaccinations-list.md) | searchはAPI、上限100/担当医の取得済み内条件 | 横断8AC＋既存画面AC、未実行 |
| L06 | [トリミング](../screens/16-trimming-list.md) | 既存名前/犬種検索と既存取得範囲 | 横断8AC＋既存画面AC、未実行 |
| L07 | [健診](../screens/25-checkups-list.md) | 日付取得範囲、名前/種別/所見は現在ページ内 | 横断8AC＋既存画面AC、未実行 |
| L08 | [在庫](../screens/18-inventory-list.md) | サーバpage20内の品名/場所/仕入先。page内バナー保持 | 横断8AC＋既存画面AC、未実行 |
| L09 | [見積](../screens/22-estimate-list.md) | 取得済み集合の全文/状態/期限、全件サーバ検索へしない | 横断8AC＋既存画面AC、未実行 |
| L10 | [会計](../domains/accounting/pages/accounting-list.md) | サーバpage20、日付/支払方法/状態、会計確定と別 | 横断8AC＋既存画面AC、未実行 |
| L11 | [顧客集計](../screens/36-aggregation-dashboard.md) | 既存飼主名サーバsearch/ページ、対象Phaseを再開しない | 横断8AC＋既存画面AC、未実行 |
| L12 | [Lステップタグ管理](../screens/31-lstep-integration.md) | TagSummaryTable取得済みタグ名検索、タグ適用/送信とは別 | 横断8AC＋既存画面AC、未実行 |
| M01 | [薬剤](../screens/settings/master-medicine.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M02 | [診療項目：診察/処置/予防接種/定期健診の4タブ](../screens/settings/master-treatment.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M03 | [診療項目：検査タブ](../screens/settings/master-examinations.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M04 | [入院プラン](../screens/settings/master-hospitalization-plan.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M05 | [物販](../screens/settings/master-merchandise.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M06 | [診断カテゴリ/病名の2タブ](../screens/settings/master-diagnosis.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M07 | [トリミングコース/optionの2タブ](../screens/settings/master-trimming.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M08 | [トリミングコース種別](../screens/settings/master-trimming-course-type.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M09 | [予約区分/グループ](../screens/settings/master-reservation-type.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M10 | [動物種](../screens/settings/master-animal-species.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M11 | [ケージ](../screens/settings/master-cage.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M12 | [キャンペーン](../screens/settings/master-campaigns.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M13 | [主訴](../screens/settings/master-chief-complaint.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M14 | [保険](../screens/settings/master-insurance.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M15 | [問診テンプレート](../screens/settings/master-interview.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M16 | [職種](../screens/settings/master-occupation.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M17 | [権限グループ](../screens/settings/master-permission-group.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M18 | [スタッフ](../screens/settings/master-staff.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M19 | [支払方法](../screens/settings/payment-methods.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M20 | [勤務テンプレート](../screens/settings/master-shift-template.md) | 既存管理一覧の名称等キーワード。既存取得済み集合/上限/有効条件・医院/権限・並替を保持 | 横断8AC＋既存画面AC、未実行 |
| M21 | [医院管理](../screens/19-clinic-settings.md) | 既存医院名等キーワード、取得済み認可集合。医院scope=allの例外を維持 | 横断8AC＋既存画面AC、未実行 |
| S01 | [medical-records共通select-pet](../screens/00-pet-selection.md) | サーバpage20、search/owner_id/種別/死亡・目的別権限。checkups権限例外保持 | 横断8AC＋既存画面AC、未実行 |
| S02 | [hospitalization共通select-pet](../screens/00-pet-selection.md) | サーバpage20、search/owner_id/種別/死亡・目的別権限。checkups権限例外保持 | 横断8AC＋既存画面AC、未実行 |
| S03 | [trimming共通select-pet](../screens/00-pet-selection.md) | サーバpage20、search/owner_id/種別/死亡・目的別権限。checkups権限例外保持 | 横断8AC＋既存画面AC、未実行 |
| S04 | [examinations共通select-pet](../screens/00-pet-selection.md) | サーバpage20、search/owner_id/種別/死亡・目的別権限。checkups権限例外保持 | 横断8AC＋既存画面AC、未実行 |
| S05 | [checkups共通select-pet](../screens/00-pet-selection.md) | サーバpage20、search/owner_id/種別/死亡・目的別権限。checkups権限例外保持 | 横断8AC＋既存画面AC、未実行 |
| S06 | [accounting共通select-pet](../screens/00-pet-selection.md) | サーバpage20、search/owner_id/種別/死亡・目的別権限。checkups権限例外保持 | 横断8AC＋既存画面AC、未実行 |
| S07 | [vaccinations共通select-pet](../screens/00-pet-selection.md) | サーバpage20、search/owner_id/種別/死亡・目的別権限。checkups権限例外保持 | 横断8AC＋既存画面AC、未実行 |
| S08 | [予約登録患者選択](../domains/reservations/pages/reservation-management.md) | 既存患者検索範囲・死亡選択保護。予約確定と分離 | 横断8AC＋既存画面AC、未実行 |
| S09 | [既存主飼主選択/付替え](../screens/04-owners-form.md) | OwnerSearchModalの候補検索、既存確認/金銭影響維持 | 横断8AC＋既存画面AC、未実行 |
| S10 | [副飼主候補選択](../screens/04-owners-form.md) | 既存候補取得/付加除外。主飼主移管と別 | 横断8AC＋既存画面AC、未実行 |
| S11 | [D19飼主候補](../screens/40-identity-links.md) | trim/min1/上限付き非page、group/link採否変更なし | 横断8AC＋既存画面AC、未実行 |
| S12 | [D19ペット候補](../screens/40-identity-links.md) | trim/min1/上限付き非page、group/link採否変更なし | 横断8AC＋既存画面AC、未実行 |
| S13 | [予防接種フォーム飼主/ペット選択](../screens/15-vaccinations-form.md) | 既存患者検索/姓名空白契約・所有関係照合。接種登録と別 | 横断8AC＋既存画面AC、未実行 |
| H01 | [カルテ内の過去カルテ履歴](../domains/medical-records/pages/medical-record-form.md) | 同一pet/閲覧医院・取得範囲・旧項目検索。統合治療プラン内でもコピーと分離 | 横断8AC＋既存画面AC、未実行 |
| H02 | [検査フォーム過去履歴](../screens/13-examinations-form.md) | pet/日付で取得した先頭20件内keyword。全検歴検索へ拡張なし | 横断8AC＋既存画面AC、未実行 |
| H03 | [予防接種フォーム/カルテ接種履歴](../screens/15-vaccinations-form.md) | 既存pet/期間/ワクチン名・取得上限内。コピー/接種と分離 | 横断8AC＋既存画面AC、未実行 |
| H04 | [トリミングフォーム過去履歴](../screens/17-trimming-form.md) | PastRecordHistoryPanel取得済み同一petの検索。コピー安全条件を維持 | 横断8AC＋既存画面AC、未実行 |
| H05 | [健診フォーム過去履歴](../screens/25-checkups-list.md) | CheckupForm内PastRecordHistoryPanelの既存集合。新しい検索対象を追加しない | 横断8AC＋既存画面AC、未実行 |
| H06 | [入院フォーム過去履歴](../screens/09-hospitalization-form.md) | PastRecordHistoryPanel同一pet/既存集合。現在割当/占有判定と別 | 横断8AC＋既存画面AC、未実行 |

## キーワードを新設しない画面・操作

| ID | 画面／操作 | 判断・維持する条件 |
|:---|:---|:---|
| N01 | [受付](../domains/reception/pages/reception-board.md)・[予約カレンダー](../domains/reservations/pages/reservation-management.md) | 日付/分類/担当等の独立filter、カレンダー移動を維持。予約内の患者検索だけS08 |
| N02 | [未納](../domains/accounting/pages/unpaid-list.md)・当日会計 | 期間/表示単位/日付。繰越と既定期間を維持、現仕様にないkeyword追加なし |
| N03 | [レジ締め](../screens/29-closing-aggregation.md)・[月次帳票](../screens/32-accounting-reports.md) | 日付/月/期間/帳票選択、集計実行/印刷と検索を混同しない |
| N04 | [勤務カレンダー](../screens/24-shift-calendar.md)・[LINE予約設定](../screens/28-line-reservation.md) | 週月/職員/予約区分等のselect・日付。勤務template管理のkeywordはM20 |
| N05 | [配信監視](../screens/34-lstep-delivery-monitor.md)・[検査受信](../../architecture/adr/007-lab-device-receive-and-commit.md) | 期間/状態/機器等の条件と明示更新。keywordを新設せず、配信/attachの認可・副作用と分離 |
| N06 | [飼主レポート](../screens/39-owner-report.md)・[LIFF](../screens/38-liff-pet-health.md) | 登録済みpetやセクションの選択。検索欄のない表示へボタン追加なし |
| N07 | [サポート一覧](../domains/support/README.md)・検査機器項目対応設定 | 基準HEADの限定静的観察ではBugReportsPageは状態/詳細、ChatHistoryPageはROOM、LabDeviceItemMasterSettingsは機器source等でkeyword欄なし。項目対応のSearchableSelectはE02/E03。新設/全件検索せず実機適合は別検証 |

## 異用途の個別判断

| ID | 画面・部品／既存根拠 | 今回の扱いと理由 |
|:---|:---|:---|
| E01 | [郵便番号補完](../screens/04-owners-form.md) | 住所を補完する操作。既存の「検索」明示実行と手入力fallback維持。台帳検索の述語を流用しない |
| E02 | [診療項目候補](../screens/common-dialogs.md#21-診療項目検索-treatmentsearchdialog)・[診断/薬剤等のSearchableSelect](../screens/settings/master-diagnosis.md)・MasterSelectModal | フォームの項目値を選ぶ入力補助。既存incremental候補探索/選択/clear契約を維持、L/Mの保存済み一覧検索とは別。選ぶだけで処方/実施/会計確定へしない |
| E03 | 担当医/職員・予約区分・ケアplan参照・キャンペーン商品等の候補選択 | [共通ダイアログ](../screens/common-dialogs.md)と各マスタ/フォームの入力補助。SearchableSelectという部品名だけで一括変更しない。飼主/pet選択Sとは区別 |
| E04 | [マニュアル記事探索](../screens/35-internal-manual.md) | 操作説明の閲覧支援、既存fuzzy全文検索を維持。業務データ一覧検索や公開/保存と別 |
| E05 | [LINE顧客紐付け候補](../screens/04-owners-form.md)・[健診タグ一括同期プレビュー](../screens/31-lstep-integration.md) | LINEアカウント候補/外部操作対象の確認。飼主/pet台帳検索と同一視せず、既存候補絞込み/明示対象者検索を維持。検索からlink/送信を新設しない |
| E06 | [カルテ内検査項目/画像の局所絞込み](../domains/medical-records/pages/medical-record-form.md)・PropertyFilterの条件候補/並替メニュー検索 | 現在開いた記録の項目/媒体/条件値を探す補助。保存済み業務履歴Hとは区別し既存局所挙動を維持。メディア原本保持・保存と別 |
| E07 | ブラウザ内検索・[登録時重複照合](../domains/owners-pets/lifecycle.md)・自動保存/臨床計算 | ブラウザ機能/保存安全検証/派生値であり一覧検索ではない。今回のボタン待ちへ変更しない |

## 検索以外の同種操作の適用照合

全画面への原則はL/M/H/Sだけに限定しない。[全23ownerの既存契約](owner-contract-coverage.md)と以下の操作意味で照合する。AC-common-ui-consistent-operationでは、同じ意味の操作を持つ各画面で該当する行を実施し、異なる意味の操作は根拠とともに分ける。存在しない操作は新設せず該当なしと理由を残す。代表画面の成功だけで全画面合格としない。

| 操作意味 | 比較する具体的操作と既存正本 | 受入で保持する境界 |
|:---|:---|:---|
| 実行 | [Button](../components/button.md)の明示起動・loading、各ownerの保存/確定 | 同目的の重複起動抑止・実行状態表示。read検索とwriteを同一化せず、金銭/臨床の保存成功単位は各owner |
| 入力消去/リセット | キーワード欄/既存リセットと[通常カルテのタブclear/複製取消](../domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015) | 入力消去は適用と別、resetは明示対象だけ。カルテの保存済み版へ戻す操作は別意味で既決を保持 |
| 状態保持/戻る | [日付付き一覧・詳細復帰](ui-contracts.md#list-filter-contract-20261017)・[LINEフローの戻る](../screens/37-line-reserve-owner-flow.md) | 同じ文脈への復帰は既適用条件/入力保持。メニュー当日初期値、別の新規フロー開始時のresetを混同しない |
| 閉じる/業務取消 | [Dialog](../components/dialog.md)・[予約取消](../domains/reservations/pages/reservation-management.md#requirement-fr-reservations-cancel-history)・[会計訂正](../domains/accounting/lifecycle.md) | closeは表示終了、業務取消は目的別write。閉じるだけでcancel/delete/saveを発火せず、必要履歴や資金処理を消さない |
| 失敗/再試行 | [共通取得状態](ui-contracts.md#interaction-consistency-20261019)・[owner別失敗/結果不明](../../design/shared/integration-contracts.md#requirement-fr-common-tech-boundary) | 読取は同じ適用条件、書込結果不明は元操作照会。失敗で入力破棄/別患者へ適用/無条件再POSTをしない |

## 照合方法と未確認

画面索引・settings索引、各画面仕様と限定的なkeyword consumerの静的観察で用途を分類した。L11は `AggregationSearchFilter`、L12は `TagSummaryTable`、M群は `MasterCRUDPage` / `MasterListPage` 等、H群は履歴部品が根拠。コードは旧挙動の所在確認で、業務採用根拠は本人承認。未採用/Phase外機能をこの台帳で有効化しない。取得集合が不明なら全件と補完せず、既存の上限/認可を維持して実装工程で照合する。

[S14横断手順](../../ops/testing/scenarios/S14-search-and-multi-word.md#cross-ui-search-20261019)で各対象IDを独立に記録し、共通部品1件の成功を全consumerへ代入しない。独立filterしかないN行と異用途E行も非回帰/対象判断を照合する。新consumerの発見時は用途を分類して本台帳・owner正本・設計・ACへ接続し、未確認を無言で対象外にしない。非検索操作の全caller/全実装適合は今回の文書棚卸しだけで完了しない。
