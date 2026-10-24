# 停止支払方法・再診回答の継続静的監査（2026-10-23）

**計画整備。製品実装Ready・公開契約採択・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-payment-methods-revisit-intake-20261023 |
| 種別 | 限定静的監査の受領記録 |
| 決定状態 | 既存採択仕様と段階的供給を継承。新規wire採択なし |
| 実装状態 | PARTIALの内訳を以下に示す。全票・全ACの完了認定なし |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

基準mainは `171b56575e7ee028e54358dad22ae28574853125`。459は停止支払方法の追跡、460は予約再診回答の親でありマスタ票ではない。受領監査を整形した記録で、[今回の索引](emr-limited-audit-plans-index-20261023.md#limited-audit-status)と[固定基準のチケット規約](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/ops/plane-ticket-conventions.md#L73-L122)の必要subset・採用SHA・有限成功単位を適用する。全親Done待ち・票数束ね・一部mergeによる親完了を避ける。

<a id="payment-revisit-audit-conclusion"></a>
## 結論

- 459は465との単一ownerを維持。同method免除、退役payments使用判定、固定候補の既知残差が現存。465で追跡済みのため重複投稿しない。
- 465-Mのmaster/guard限定成果を465-Cの全財務consumer完成で一律gateしない。
- 460の保存model・API契約・FE/共通gate/限定pet readerは再利用。runtime route・認可・receipt・writer接続とactivationは未完。
- 全実行NOT_RUN。旧PRの限定PASSは過去revisionの証拠だけで、現在mainの全体受入へ流用しない。

<a id="payment-revisit-audit-payment-evidence"></a>
## 459：AC・証拠・fixture・残差

全行の状態は**PARTIAL / NOT_RUN**。これは既実装subsetと未接続・依存を合わせた分類であり、各fixtureの実行結果ではない。

| AC | 現mainの根拠／fixture期待値 | 不足・既存受領先 |
|---|---|---|
| new-receipt | [builders:45–63](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/accounting_service_builders.go#L45-L63)、[core:197–222](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/accounting_service_core.go#L197-L222)、[PaymentCard:27–33](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/accounting/components/PaymentCard.tsx#L27-L33)。既存[停止方法6case:21–67](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/accounting_service_inactive_payment_test.go#L21-L67)。停止Mに既存fundあり＋今回同M新fundは拒否する期待値 | savedMethods一致で停止免除、PaymentCard固定4候補でactive reader未接続。465-C→506/464/462 |
| save-recheck | [master service:105–132](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/payment_method_master_service.go#L105-L132)、[採択契約:35–47](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-449-459-465-payment-master-contract-20261021.md#L35-L47)。停止先行は新受領拒否、受領先行は成立fund保持、複合1件停止は全rollback。失権/監査失敗もzero-write | 同method row lock/同TX現在認可・監査とconsumer lock参加が未接続。465-M→465-C |
| history-aggregation | [usage repository:86–142](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/payment_method_master_repository.go#L86-L142)、[保存model:476–517](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/model/storage_finance.go#L476-L517)・[566–582](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/model/storage_finance.go#L566-L582)。停止後もfund/event/correction/帳票参照維持、4種の現method参照があれば削除拒否 | usage/deleteが退役payments参照。保存modelだけではhistory consumer適合を証明しない。465-M usage/guard・465-C参照writer |
| existing-allocation | [採択契約:24–31](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-449-459-465-payment-master-contract-20261021.md#L24-L31)。停止済みfundから充当/解除/再配分、新fundなし、元額/日付保持 | 最終fund接続は506/464。同method一致だけの旧helperでは操作分類を証明できない。506/464/465-C |
| refund-correction-boundary | [契約:28–29](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-449-459-465-payment-master-contract-20261021.md#L28-L29)・[61–63](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-449-459-465-payment-master-contract-20261021.md#L61-L63)。停止方法の返金/金額不変訂正はactive拒否なし、存在scope/理由/残高/参照lock維持 | 参照保護lockとactive guardを分離するconsumer参加未受領。506/465-C |
| no-receipt | [契約:30–31](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-449-459-465-payment-master-contract-20261021.md#L30-L31)。全method停止でも入金なし未納/明細あり0円確定、新fund/方法付替えなし | 最終Completeへの空集合guard接続と全必要caller証拠なし。462/505/465-C |
| idempotent-result | [builders:45–63](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/accounting_service_builders.go#L45-L63)、[契約:49–57](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-449-459-465-payment-master-contract-20261021.md#L49-L57)。成立same-key同内容は旧結果、異内容conflict、unknownは同key照会、新key重複なし | 既存method集合は成立receipt/operation同一性ではない。465-C→489/506/464必要subsetの接続残 |

<a id="payment-revisit-audit-intake-evidence"></a>
## 460：再診回答のAC・証拠・fixture・残差

全行の状態は**PARTIAL / NOT_RUN**。共通の静的根拠は[保存model:471–521](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/model/storage_clinical.go#L471-L521)と[予約route:157–218](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/reservation/routes.go#L157-L218)。保存形があることをsubmit/runtime接続としない。

| AC | 根拠・fixture期待値 | 不足・既存受領先 |
|---|---|---|
| optional | [仕様:633–655](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L633-L655)。[RevisitIntakePage fixture:852–916](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/line-reserve/src/pages/RevisitIntakePage.test.tsx#L852-L916)：空/Unicode目的、任意項目を補完しない | model/FE fixtureあり、submit/runtime未接続。未回答で予約受付を止めない統合証拠なし。471/472/473/474 |
| multi-pet | [仕様:656](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L656)。同fixture1161–1181/2836–2874のtarget切替・入力/未確定分離。2予約A/Bの異なる回答保存・再読込は追加統合fixture | 1appointment=1pet形状は採用済み。複数頭batchの確定/部分結果/取消は別残件、pet reader完成と混同しない。460/468/471–474、519/520再利用 |
| revisions | [仕様:657](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L657)。同fixture313–343/621–642/918–945の最新/旧版・訂正key | 旧版/server時刻storageあり。受付前後の実write可否、append-only/receipt原子性はBE未接続。471/472/473/475 |
| reschedule | [仕様:658](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L658)。同appointment日時変更でgeneration/原文/received_at不変、受付前訂正可という期待値 | gate library採用は予約writer参加の証拠ではない。同世代保持の実writer供給残。468/471 |
| new-target | [仕様:659](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L659)。同fixture1161–1181/2836–2874、A→B→A/取消再作成で旧履歴保持・新対象未回答 | FE分離あり。実世代writer、取消/受付との両順序原子性は別受入。468–470/471–473 |
| provenance | [仕様:660](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L660)。[RevisitIntakeSection fixture:152–231](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/src/features/reservations/components/RevisitIntakeSection.test.tsx#L152-L231)：7項目/版/出所・current/previous分離 | 予約FE readerは#752、治療プランFEは477。staff/record read routeと固定capture未接続。475/476/477/472 |
| pending-entry | [仕様:676](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L676)。pending/confirmedの入口とeditableを独立判定 | 下記共通activation依存。471/472/473/474 |
| post-checkin-readonly | [仕様:677](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L677)。受付済/診療中/会計/完了は閲覧のみ、各write拒否 | 下記共通activation依存。471/472/473/474 |
| no-show-not-readonly | [仕様:678](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L678)。未来院/過去日/未回答だけで一律readonlyにしない | 下記共通activation依存。471/472/473/474 |
| target-unresolved-display | [仕様:679](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L679)。未解決と未回答を区別、純GETでtarget/answerを作らない | 下記共通activation依存。471/472/473/474 |
| author-only-view | [仕様:680](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/reservations/pages/reservation-management.md#L680)。LINE他作者本文非公開、現在pet grant、staffは既存staff grant | 下記共通activation依存。471/472/473/474 |

末尾5ACの共通根拠は[activation flag:1–11](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/frontend/line-reserve/src/lib/revisit-intake-activation.ts#L1-L11)、[route:181–218](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/reservation/routes.go#L181-L218)。LIFF全体activation=falseで、API/現在認可の統合未接続を維持する。FE read/error fixtureはRevisitIntakePage.test.tsx:344–437、RevisitIntakeSection.test.tsx:121–187/260–274。mock/fixtureの存在を実API合格にしない。

<a id="payment-revisit-audit-finite-units"></a>
## 有限PR単位

| 成果 | 単一owner・成功単位 | 必要subset／gateにしない範囲 |
|---|---|---|
| 465-M | 465 single owner、459は追跡。master現在認可、停止/再有効化、同row lock、4種fund/event参照削除保護、同TX監査/再読込、ValidateNewReceiptMethodsInTx内部interfaceとfixture | 全会計/506/464/462/465-C利用開始をgateにしない |
| 465-C | 465統合、506 fund/464口座/462 Complete/489 C06。今回新fund method集合→共有guard→最外TX保存・receipt、PaymentCard新候補/旧表示分離、非新受領caller保護 | 465-M guard採用SHAと各金銭owner必要subset受領SHAは未確認。489全体待ちにしない |
| 460 runtime intake | 471統合、472認可、473receipt、468–470 writer。頭別1submitをstrict raw/現在grant/current generation/append-only answer/同TX receipt→same-key lookupまで閉じる。既存3route契約を再設計しない | gate a28f81bは採用済み。writer/認可/receipt固定供給SHA未受領。全master、475/477 FE再作成、全金銭consumer、全batch設計をgateにしない |
| 460 provenance | 476固定record capture、475予約reader、477record reader、472行可視性。既存FEと原予約targetの捕捉、認可済readonly readerを結ぶ | 固定binding実writer/reader必要subset。現target fallback/旧NULL backfillなし。設計済みを未設計へ戻さない |
| 474 recovery Slice B/activation | 474 FE、471–473 API/receiptと関連writer。discovery Slice Aを再利用しcontrol/closure/fence必要wireと全関係writer受領後だけ新key再開 | 有効化は別PRで環境利用可証跡。ACTIVE=false/RECOVERY_FENCE_AVAILABLE=false維持。0件/completeはwrite許可ではない |

<a id="payment-revisit-audit-adopted-prs"></a>
## PR再読の限定成果

| PR | 状態・採用SHA | 限定成果 |
|---|---|---|
| [728](https://github.com/MinoruSoga/AnimalEkarte/pull/728) | closed/unmerged、727へ統合 | 再診事前問診の履歴・送信receipt基盤。merge_commit_sha欄だけでmerged扱いしない |
| [749](https://github.com/MinoruSoga/AnimalEkarte/pull/749) | merged `a28f81bfe99010c440ff42a51b8f31811ae1904f` | reservation-graph:write:v1共通gate library |
| [752](https://github.com/MinoruSoga/AnimalEkarte/pull/752) | merged `90ad338f34d88a79ca9daa2ea77a9cb87eb7ce81` | 予約詳細の原文/訂正履歴read-only FE契約mock |
| [756](https://github.com/MinoruSoga/AnimalEkarte/pull/756) | merged `6a24ae268f47d60d2a293addaa538c1088719870` | LIFF回答/訂正/再送/結果照会FE契約mock |
| [779](https://github.com/MinoruSoga/AnimalEkarte/pull/779) | merged `19855942f762f31f0403b5de0ddbc61b761fd94e` | BE未実装中に入口/ページ/APIを止める利用開始flag |
| [781](https://github.com/MinoruSoga/AnimalEkarte/pull/781) | merged `1afa46fe4eccb2a421023b16acc29c1dcb804519` | editorが外れたreadでも未送信draft離脱確認 |
| [776](https://github.com/MinoruSoga/AnimalEkarte/pull/776) | merged `5fbad20a1ab01d966d3f2197f147018d4c6ce774` | 519互換client段階1、520 petペア表示/491限定reader |
| [799](https://github.com/MinoruSoga/AnimalEkarte/pull/799) | merged `0f2a9f1ce6ae8627b776db0479395b7b852de1b9` | 再読込回復Slice A discovery/保留/再照会/ページング、未接続FE |

<a id="payment-revisit-audit-limits"></a>
## 確認限界

実環境利用可・DB適用・runtime受入は未確認。他子票468–470/475–479の最新全コメントは今回の受領監査では未取得。owner対応は460本文/471・474契約に基づき、個別完了やclaim不在は断定しない。既知残差を新発見件数に加算しない。親側の460追補は索引の受領記録へ反映し、459は465での既知追跡と重複投稿しない。
