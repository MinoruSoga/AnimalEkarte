# 共通UIの採用修正・caller限定静的監査

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-shared-ui-20261023 |
| 種別 | 固定mainの限定静的監査受領記録 |
| 決定状態 | 既存仕様を継承。新しい業務・公開契約を採択しない |
| 実装状態 | 以下の静的分類に限定。全票・全ACの実装監査完了ではない |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

本稿は提供された監査報告の整形であり、文書編集時に全callerを再監査した記録ではない。[今回の索引](emr-audit-ready-handoff-index-20261023.md#audit-ready-scope)を参照。

基準: `77c28fbfbafa257dcca2421af99ff72da6889717` (2026-10-23)。PR662 merge `949a55a08996f8d185ef15438030d6c7c51578d0` と対象ファイルの全文を比較。実行・ブラウザ・DB・Mac操作・製品変更・外部投稿なし。実見た目、実API、保存、認可、UATは **NOT_RUN**。旧画像/過去PASSは今回の現main受入ではない。

<a id="shared-ui-conclusions"></a>
## 結論

4票の採用修正と主要callerの接続を確認。今回の有限範囲で新たな「未実装」「未接続」「後続差分による修正消失」は実証しなかった。4票完了/見た目PASSとはしない。全票はPlane started（過去コメントReview維持）、assignees空。新規重複票は不要、残受入は元票で管理する。

最新Plane本文と全コメント取得: 383=2件、396=3件、397=3件、398=2件、全ページnext_page_results=false。PR662本文・timelineも読取した報告を受領。PR662は2026-10-13 merge済み。compareの847 commits aheadは祖先関係確認にだけ使用し、巨大compareのfiles一覧で変更不存在を断定していない。

以下の相対パスはすべて https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/ を接頭辞とする。

<a id="shared-ui-ac"></a>
## AC別

<a id="shared-ui-read-states"></a>
### EMR-383: 検査/見積詳細のloading・取得失敗・非開示

- 実装・接続: 検査のloading/notFound/errorがPageLayoutの専用status pageに入り、説明と固定一覧戻り、error時retryを表示する。`frontend/src/features/examinations/routes/ExaminationForm.tsx:216-267`、`components/ExaminationFormStatusPage.tsx:14-29`。旧「空白＋赤文字断片」の枝の置換は残る。
- 実装・接続: 見積も専用PageLayout→LoadingFallback/EmptyState→説明・一覧・retry。`frontend/src/features/estimates/routes/EstimateDetail.tsx:35-46,109-168`。通常編集・削除操作のrenderより前にreturnする。
- 実装・接続: `frontend/src/lib/entity-read-result.ts:26-28,50-70` で403/404を別内部statusに分類し同一非開示UIへ、その他はerror+retry。検査 `hooks/use-examination-form-load.ts:18-26` → `hooks/use-examination-form.ts:217-221` がその結果を画面へ渡す。共有EmptyState `frontend/src/components/shared/DataStates/DataStates.tsx:54-74` に説明とchildrenが接続する。サーバerror本文を画面説明に流さない。
- 後続差分: PR662 merge版から検査通常フォームのheader action移動・title等は変化したが当該status枝は維持。見積はmaxWidth token化・icon撤去等で当該枝維持。
- fixture証跡（未実行）: ExaminationForm.read-states.test.tsx:97-160はmock read flags、notFound、retry spy、固定一覧とcontext header戻り。EstimateDetail.test.tsx:214-293はloading、AxiosError404、bodyなし、network error/refetch、固定一覧/context戻り。403実APIはこの局所テストの実行証明ではない。
- NOT_RUN: 他院ID/不存在ID/403/404/network/5xxからretry成功・再失敗、4幅の説明/戻りの可視性、編集UI非表示。既存の合成証拠と実APIを分離。
- 有限後続: EMR-383 UI ownerが上記2route×loading/non-disclosure/errorを現SHAで受入。実API tenant fixtureは383＋既存API ownerの供給依存。新しい権限仕様は選定しない。

<a id="shared-ui-side-peek"></a>
### EMR-396: 共通SidePeekの狭幅・操作領域

- 実装: `frontend/src/lib/design-tokens.ts:666-671` に `w-[520px] max-w-full`、`frontend/src/components/shared/SidePeek/SidePeekBody.tsx:5-8` に `px-4 sm:px-16`。BodyはPR662 merge版とblob同一。
- 実装・接続: SidePeekPanel.tsx:29-37が幅とtabletOverlay位置tokenを消費。MasterSidePanel.tsx:109-146がToolbar/Body/Footerを同じPanelへ配置、Footer.tsx:23-45はreadOnly閉じる/通常キャンセル・保存。design-tokens.ts:849-866にtoolbar/footer shrink-0、body overflow-y-auto。
- caller: StaffSettings.tsx:198-210→StaffSidePanel.tsx:272-338、OccupationSettings.tsx:118-128→OccupationSidePanel.tsx:80-102、CageSettings.tsx:332-341→CageSidePanel.tsx:129-210。3callerともtabletOverlay指定。MasterCRUDPage.tsx:134-145→MasterListPage.tsx:94-103→MasterPageShell.tsx:58-78（relative親、min-w-0本文、sidePanel slot）がoverlayの親条件を満たす。
- 後続進展: EMR-411位置token/opt-in追加後も幅上限は保持。旧コメントのMedicine motion外枠除外は現在コードで `MedicineSidePanel.tsx:44-50` がmax-w-full＋位置tokenを持つため「現在も固定幅未修正」とは言わない。今回のMedicine全nested caller・医院/シフト独自余白・二重パネルは未監査であり全体PASSなし。
- 依存/NOT_RUN: 390/768/1280/1568×スタッフ新規/編集上下、職種新規/編集、ケージ編集の文字・閉じる・取消・保存の可視性。focus、実保存、認可はNOT_RUN。footerの接続と画面上到達を区別。
- 有限後続: EMR-396 UI ownerで元対象3callerの4幅再受入。Medicine等の旧除外残はEMR-373/411既存ownerで更新し、396に新規欠陥として二重計上しない。個人担当未割当。

<a id="shared-ui-cage-table"></a>
### EMR-397: ケージ表の折返し

- 実装: CageSortableTable.tsx:27-46に列最小幅、固定見出しnowrap、同:176-203にエリア/サイズ/単価/状態/操作nowrap。ケージ名セルは同:158-175で列最小幅のみ（通常名前へnowrapを追加しない）。共有 `frontend/src/components/ui/table.tsx:59-72` から見出しnowrapを供給するためケージ名見出しも接続する。
- 実装・接続: 同:95-109のoverflow-auto→Table、共有table.tsx:10-13のoverflow-auto。CageSettings.tsx:365-379が実データ・権限・編集/連結callbackを渡す。表内StatusPillへisActive接続。min-widthで横スクロールを選択する設計であり、全列の初期viewport同時表示は要求していない。
- 後続差分: CageSortableTable.tsxはPR662 mergeと現mainで同一blob `6cdb2e3a9806d54c13dd200ed1880b3ec1cbc663`。共有見出しtokenの色/typography変更をこのローカル幅修正消失とは扱わない。
- NOT_RUN: 4幅×パネル閉/新規/編集×左右端で見出し、長名、中型/小型、有効/無効、単価、編集操作の可読性。連結/並替/保存/認可受入は元範囲外。旧wave27/29/30・29testsは歴史証拠。
- 有限後続: EMR-397 UI owner、共通表変更時のみEMR-373共有ownerと照合。既存397へ受入を記録し新票不要。

<a id="shared-ui-inventory"></a>
### EMR-398: 在庫の有効期限・数量単位

- 実装: InventoryListPanels.tsx:51-61に数量+単位/最低在庫+単位/日付の同一cell nowrap、最低在庫/期限はhidden lg:table-cell保持。header側同:153-162もnowrap+可視条件一致。長名セル同:36-46はnowrapを追加していない。
- 実装・接続: InventoryList.tsx:148-166→InventoryListContent、同Panels:251-260→共有DataTable.tsx:80-103（min-w-0/overflow-auto/col.className）→table.tsx:10-13の局所overflow。item値とcanEdit条件は保持。
- 後続差分: PR662 mergeからPanelsの差分は警告集計数 `.toLocaleString()`→formatNumberのみ。当該cell/headerと権限条件は変化なし。全API/業務計算の同等性を主張する範囲ではない。
- fixture制約: 現InventoryList.test.tsx:57-75は数量10、単位「個」、最低在庫5で期限fixtureなし。同:133-160はEMR-378カテゴリ/状態classをassertするだけ。旧PRの既存26tests PASSを「2027-12-31 / 20 tablet」の専用回帰合格に換算しない。他ファイル全域のテスト不存在は断定しない。
- NOT_RUN: 元fixture「2027-12-31」「20 tablet」と長名、4幅、局所scroll右端。390/768は元の補助列非表示ルール保持を確認する条件で、隠された期限が見えないこと自体を欠陥としない。
- 有限後続: EMR-398 UI ownerがこの専用fixtureで日時/単位nowrapの回帰を追加・実施する範囲を引継ぎ。実装変更の必要は未実証。実保存/在庫業務受入は別。

<a id="shared-ui-spec-boundary"></a>
## 確定仕様との境界

現 `DESIGN.md:287-307` はtablet-first、1280 sidebar境界、補助列段階表示、識別・状態・操作の到達性を規定し、Mobile<=767は本体規定外と明記。`docs/spec/design-system.md:233-240,347-363` も同正本へ委譲。これは4票の既存390px追試条件を自動取消す根拠ではない。スマホ全面適合へ拡張せず、元の限定UI修正を4幅で再確認する。

`docs/spec/screens/18-inventory-list.md:39-42` の最低在庫・期限・導出状態の意味はCSS調整で変更しない。今回API/DB仕様監査へ拡張していない。

共有CSS `frontend/src/styles/globals.css:329-355` のdata-side-peek連動support widget退避/非表示と、SidePeekPanelのdata属性接続も確認。開いたsupportや二重パネルの実描画受入は未実施。PR662のR662-A11Y-01およびカルテ/飼主/画像bundleの旧レビューは4票以外の既存残件であり新規欠陥再起票しない。
