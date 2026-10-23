# 通知・操作領域・レイアウト後続の限定静的監査

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-ui-layout-followups-20261023 |
| 種別 | 固定mainの限定静的監査受領記録 |
| 決定状態 | 既存仕様を継承。新しい業務・公開契約を採択しない |
| 実装状態 | 以下の静的分類に限定。全票・全ACの実装監査完了ではない |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

本稿は提供された監査報告の整形であり、文書編集時に全callerを再監査した記録ではない。[今回の索引](emr-audit-ready-handoff-index-20261023.md#audit-ready-scope)を参照。

固定source: `77c28fbfbafa257dcca2421af99ff72da6889717`。8票の最新本文・全コメント（計27件、各next_page_results=false）を直接取得。#649/#662の採用差分とmerge、#698の同期PRとしての位置づけを確認。実ブラウザ・実行・製品編集・DB・Mac・Plane投稿は一切なし。

<a id="ui-followup-conclusions"></a>
## 結論
新規製品欠陥の確定は0件。7票は元報告と確認した現source範囲で追加製品修正不要（限定）。387は既存432→401 C3-1の依存受入を保持し、完了や現存衝突を断定しない。393/394には長名fixture/対象レイアウト自動回帰の証拠不足を明記。全8票の現在版の描画・操作・試験はNOT_RUN。

<a id="ui-followup-boundaries"></a>
## MECE境界
339=通知同一性、381=通知配置で別責務。390〜394は固定短ラベルの非縮小/非折返し＋可変本文の幅確保という同系統だが、実装primitive/既存票ownerを保持し横断新票を作らない。387=操作遮蔽で432と401へ集約。後続412のカード変更、519の表示抑制は既存正本を優先。無関係な全画面には拡張しない。

<a id="ui-followup-toast-identity"></a>
## EMR-339 — 同文面toast重複・auto-create二重通知
- Owner: EMR-339
- 原報告: 同時500で同文面が件数分表示。予約/カルテ作成mutation onErrorとauto-create catchが重複。
- 採用修正: https://github.com/MinoruSoga/AnimalEkarte/pull/649
- 現caller/primitive: handle-api-error.ts:193-198 は抽出文面をidに使用。auto-create.ts:203-214 はfailureを保存しAxiosのみ再通知抑止。予約/カルテのmutation onErrorは各67/23行でhandleApiErrorを保持。QueryCacheはsilentErrorを尊重。集計はEMR-334で単一取得へ変わり、旧7並列GETは現callerに存在しない。AggregationDashboardPage:114,132のinline errorとReservationFormModal:412-418のsubmitErrorを保持。
- fixture/回帰契約: handle-api-error.test:447-488 同500・別context同文面・403対500・非AxiosのID。react-query.test:18-40 通常/silent。auto-create-retry.test:502-653 予約500/カルテ500/非Axios、failureと予約ID保持・再試行。これらはmock/構造証拠であり実Sonner描画の証明ではない。
- 残差: 追加製品修正不要（限定）。10/19の旧STG BLOCKEDは同時500を発生できなかったという記録で合格ではない。将来許可時、合成500複数/別文面/通信断/timeout、予約/カルテ失敗、inlineと再試行の組合せを確認。EMR-334/336のサーバ原因を再計上しない。
- 最新コメント: 2026-10-19T09:06:16.555781Z、全10件確認。
- Source:
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/lib/handle-api-error.ts
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/lib/react-query.ts
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/medical-records/hooks/use-medical-record-auto-create.ts
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/hooks/use-create-reservation.ts
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/medical-records/api/create-medical-record.ts
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/aggregation/api/get-cpm-stage-counts.ts
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/aggregation/routes/AggregationDashboardPage.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx

<a id="ui-followup-toast-position"></a>
## EMR-381 — 通知配置
- Owner: EMR-381
- 原報告: 左上toastがサイドナビ・ヘッダに重なる。lab-device/examinations/medical-recordsで旧観測。
- 採用修正: https://github.com/MinoruSoga/AnimalEkarte/pull/662
- 現caller/primitive: sonner.tsx:9-15 top-right、offset=72、mobileOffset=72、duration=4000、closeButton。末尾props展開でoverride可能。app/provider.tsx:17はoverrideなしのToaster。339の重複排除とは異なる契約で、通知数と配置の残件を混同しない。
- fixture/回帰契約: sonner.test.tsx:43-71はSonnerをmockしposition/offset/mobileOffset/override契約をassert。実座標・本文遮蔽の検査ではない。
- 残差: 追加製品修正不要（限定）。過渡toastの本文遮蔽、長文・同時異種通知、390とdesktopの実表示はNOT_RUN。旧390画像確認を現在版の合格へ転用しない。
- 最新コメント: 2026-10-18T01:52:58.434886Z、全2件確認。
- Source:
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/ui/sonner.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/app/provider.tsx

<a id="ui-followup-support-overlap"></a>
## EMR-387 — FABと操作要素の重なり
- Owner: EMR-432 → 残存時 EMR-401 C3-1（EMR-403契約）
- 原報告: 新規予約確定と/settings/staff末尾行編集を右下FABが覆う。最新本文/10月19日裁定は432受入後の残存確認。
- 採用修正: https://github.com/MinoruSoga/AnimalEkarte/pull/662
- 現caller/primitive: SupportWidget.tsx:179-185はopen属性を開時だけ付与。globals.css:348-354はDialog/AlertDialog/SidePeek存在中の閉じたlauncherをdisplay:none。予約ModalはDialogを使用。SupportWidgetのfocus候補は非表示/disabled/inertを除外しactiveElementで復帰を検証。StaffSettings→MasterCRUDPage→MasterListPageの既存系統は維持。support layoutは64px footprint/80px余白、カルテfixed-actionは別96px定数。#698は製品修正PRではなくstaging同期で、432の実装は本文上#688/#696。
- fixture/回帰契約: SupportWidget.test:74-110 open属性/CSS文字列、178以降focus候補。SupportWidget.portal.test:75以降実ConfirmDialog portalでclose/Escape/reset取消。CSS文字列/DOMの確認は衝突座標の証明でない。
- 残差: 重複実装票は作らない。追加製品修正の必要性は残存受入依存。432受入完了は本監査では確立していない。予約modal/職員最下行、閉/開サポート、復帰focusはNOT_RUN、既存owner順序を保持。
- 最新コメント: 2026-10-19T05:48:08.058078Z、全4件確認。
- Source:
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/support/components/SupportWidget.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/styles/globals.css
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/ui/dialog.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/routes/StaffSettings.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/components/StaffSettingsRow.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/components/MasterCRUDPage.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/constants/support-widget-layout.ts
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/constants/medical-record-fixed-action-layout.ts

<a id="ui-followup-care-type"></a>
## EMR-390 — 入院ケア種別の省略
- Owner: EMR-390
- 原報告: 1568入院追加フォームの「指示・その他」がw-28で省略、編集行も同型。
- 採用修正: https://github.com/MinoruSoga/AnimalEkarte/pull/662
- 現caller/primitive: AddForm:150-165 / EditRow:152-166はflex-wrap、種別w-28 min-w-fit shrink-0、名称flex-1 min-w-0。共有SelectTrigger:53はnowrap/line-clamp-1だがcallerの最小幅指定を保持。CarePlanTab:145/190が編集/追加を描画。
- fixture/回帰契約: AddForm.test:405-422 とEditRow.test:336-345に対象class契約。既存参照選択/金額/payloadテストあり。classの存在は最長ラベルの実可視幅を証明しない。
- 残差: 追加製品修正不要（限定）。1568/390の追加・編集両方、最長種別と名称・金額・保存payload維持はNOT_RUN。旧追試は追加1568のみ、編集UNKNOWNだったことを維持。
- 最新コメント: 2026-10-18T01:54:27.869541Z、全3件確認。
- Source:
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/hospitalization/components/CarePlanTab/AddForm.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/hospitalization/components/CarePlanTab/EditRow.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/ui/select.tsx

<a id="ui-followup-patient-card"></a>
## EMR-391 — 狭幅患者カードの縦割れ
- Owner: EMR-391（後続共有変更はEMR-412）
- 原報告: 390入院/トリミング編集で診療種別・保険等が担当欄と幅競合。
- 採用修正: https://github.com/MinoruSoga/AnimalEkarte/pull/662
- 現caller/primitive: PatientInfoCard:241-296は情報群basis-full lg:basis-0、種別shrink-0 nowrap、保険は狭幅全幅/大画面lg:truncate、担当はw-full sm:w-auto。後続412でsm→lgの情報群境界が変わっており662時点と完全同一ではない。HospitalizationFormFields:118-142 / TrimmingFormColumns:87以降の実callerで種別・保険・担当を渡す。
- fixture/回帰契約: PatientInfoCard.test:166-201にbasis-full/lg、種別nowrap、担当w-full/sm、保険lg:truncate。担当テストの名称は「lg以上」だがassert/実装はsmのまま。記述上の不一致を製品欠陥としない。
- 残差: 追加製品修正不要（限定）。390の入院/トリミング、長い保険/担当、担当操作/既存省略はNOT_RUN。後続412の共有primitive責務を391へ重複計上しない。
- 最新コメント: 2026-10-18T01:54:45.509135Z、全2件確認。
- Source:
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/PatientInfoCard/PatientInfoCard.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/hospitalization/routes/HospitalizationFormFields.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/trimming/routes/TrimmingFormColumns.tsx

<a id="ui-followup-master-trigger"></a>
## EMR-392 — マスタ選択の変更ラベル折返し
- Owner: EMR-392
- 原報告: 390トリミング選択済コースの「変更」が縦割れ。
- 採用修正: https://github.com/MinoruSoga/AnimalEkarte/pull/662
- 現caller/primitive: MasterSelectTrigger:62/91のblock/inline双方で変更＋アイコン群shrink-0 nowrap。block本文はtruncateなし、inline本文は既存truncate。TrimmingFormColumns→TrimmingLeftColumn:81-89が選択コースをblockで渡しonCourseModalOpenを保持。
- fixture/回帰契約: MasterSelectTrigger.test:13-63は両variant、名称/価格、block非truncate、onClick。fixtureはプレミアムトリミングコース/12000。
- 残差: 追加製品修正不要（限定）。390のblock/inline実幅、長名/価格、keyboardと選択維持はNOT_RUN。旧別variant画像UNKNOWNを保持。
- 最新コメント: 2026-10-18T01:54:46.804316Z、全2件確認。
- Source:
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/components/shared/MasterSelectModal/MasterSelectTrigger.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/trimming/routes/TrimmingFormColumns.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/trimming/components/TrimmingLeftColumn.tsx

<a id="ui-followup-vaccine-table"></a>
## EMR-393 — LIFFワクチン日付・見出し分断
- Owner: EMR-393
- 原報告: 合成長名fixtureで390/768/1280/1568の日付と固定見出しが分断。
- 採用修正: https://github.com/MinoruSoga/AnimalEkarte/pull/662
- 現caller/primitive: PetHealthPage:154-181は表内overflow-x-auto、min-w-[18rem]、見出し/ISO日付nowrap、ワクチン名は折返し可能。日付値は加工せずnext_due_at欠落のみ—。liff App:30が現caller。
- fixture/回帰契約: PetHealthPage.test.tsxのhealthCard fixtureはpets: []。現ファイルの8テストはbrand/error/retry中心でワクチン行を描画しない。#662の「既存2 suites16 PASS」は対象長名DOM/4幅の自動回帰カバーを意味しない。
- 残差: 追加製品修正不要（限定）。証拠残差あり: 非空pets/vaccines、長名、ISO日付、next_due_at nullのfixtureでnowrap/表内横スクロール契約を保持する対象回帰と4幅受入が必要。今回は追加・実行せずNOT_RUN。
- 最新コメント: 2026-10-18T01:55:33.634237Z、全2件確認。
- Source:
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/liff/src/pages/PetHealthPage.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/liff/src/App.tsx

<a id="ui-followup-reservation-badge"></a>
## EMR-394 — LINE予約履歴状態バッジ縦割れ
- Owner: EMR-394
- 原報告: 390/768/1280/1568の長いペット/コース/担当名が状態バッジを圧縮。
- 採用修正: https://github.com/MinoruSoga/AnimalEkarte/pull/662
- 現caller/primitive: MyReservationsPage:177-193はgap-2、左min-w-0 flex-1 break-words、状態shrink-0 nowrap。現App→ReservationPageSwitch:96-109がcaller。後続EMR-519でcurrent_targetの無い履歴にpet_nameを出さないHOLDとなり、現表示はcourse_nameとstaff_name。元662のpet_name fallbackを復活させない。
- fixture/回帰契約: MyReservationsPage.testは一般診察/山田先生の短名fixtureと取消/error/retry/再診問診/ID保持を検査。nowrap/shrink-0/長名レイアウトのassertなし。
- 残差: 追加製品修正不要（限定）。証拠残差あり: 現契約に適合する長course/staff、confirmed/completed等状態の対象fixture、4幅受入を既存394へ残す。取消・送信条件に変更しない。NOT_RUN。
- 最新コメント: 2026-10-18T01:55:35.655350Z、全2件確認。
- Source:
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/line-reserve/src/pages/MyReservationsPage.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/line-reserve/src/App.tsx
  - https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/line-reserve/src/app-page-switch.tsx


<a id="ui-followup-evidence-limits"></a>
## 証拠の限界
PR本文・過去コメントのPASS/画像観測は当時の申告で、今回の独立検証ではない。CSSとmock/DOM assertionは実viewportの可視性・フォーカス・クリック非遮蔽を証明しない。本記録の根拠は上記の固定SHA sourceリンクとowner・残差に限定する。
