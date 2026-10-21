# EMR-449 / 459 / 465 金銭マスタ接続：限定選定契約

2026-10-21。親レビュー採択の技術契約。業務正本を実現する設計であり実装・試験・DB適用・実決済は行っていない。試験実行は NOT_RUN。現コードを目標仕様へ昇格せず、正本準拠に作り直す必要subsetを示す。既存Done成果は再実装しない。

## 1. 固定根拠と票の境界

確認mainは `755f77abd2e6d63a2fc608fcfe6a518a61afcfff`。以下のblobはすべてこのmainで存在を確認した。#784 / #798の採択文書はmain収録済みとして扱い、旧コメントの「open/未merge」を現在の着手条件へ戻さない。

- 業務正本: [停止方法](../../spec/domains/accounting/lifecycle.md#accounting-inactive-payment-method)、[変更権限](../../spec/domains/accounting/lifecycle.md#requirement-fr-accounting-master-authorized-maintenance)。
- 設計: [停止判定](../../design/domains/accounting/technical-spec.md#accounting-inactive-payment-design)、[権限mapping](../../design/domains/accounting/technical-spec.md#accounting-master-permission-map)。
- C06: [operation参加契約](../../design/shared/operation-participation.md) §1.1a、§1.2、§1.3、採択済みQ。
- 設定画面: [支払方法](../../spec/screens/settings/payment-methods.md)。
- 既存保存形状: `backend/internal/model/storage_finance.go` の PaymentFund / FundConsumptionEvent / FundRecordCorrection。旧Payment/Split/BillingRefundを復活させない。

449 (`c9fdd0eb-a4f5-4274-b8d5-91d4f34d3190`) は444配下の調査・参加表owner。465 (`4c9aad9b-5d2a-4e90-be37-2d5c594a8065`) が新受領停止判定の単一実装owner。459 (`807659a6-fc15-43da-aba9-229bb48594b4`) は依存・受入の追跡票で別実装PRを作らない。原稿作成担当が最新本文と全コメントを確認（449=5、459=2、465=3件、次ページなし）。465旧6ケースPASSは同method追加受領・停止競合・新fund接続の合格証拠ではない。

FR/ACへの対応: マスタ変更権限FR、および `AC-accounting-payment-inactive-` の `new-receipt / save-recheck / history-aggregation / existing-allocation / refund-correction-boundary / no-receipt / idempotent-result` の7AC。master-admin-edit / master-authorized-staff-edit / master-ungranted-staff-denied / master-permission-rechecked は停止writer側の現在認可の根拠。

## 2. 所有と入力を閉じる

465は新しい金銭command・endpoint・成功receiptを作らない。506の実受領commandと464の預り受領command（元fund作成は506）に参加する内部関数 `ValidateNewReceiptMethodsInTx` を選定する。呼出しは各ownerの現在認可・scope確認とC06成功replay分岐の後、最初の金銭writeより前。引数は同じambient TX、認可済みclinic、VerifiedCommandから抽出した**今回新規作成するfundのmethod ID集合**だけ。任意DTO/pointerを捕捉した別入力を使わない。

内部集合のclosed shapeは `{clinic_id:uint64, new_fund_methods: uint64[]}`。clinicは認可済みrequest contextと元targetに一致、methodsは新fundの保存予定IDから重複排除し昇順にした値。clientが「既存」と自己申告するboolean、method一致、差額、既存請求IDの有無で新受領を判定しない。金銭額・受領日・actor・配賦・理由は506/464のcommandにのみ属し、本guardで複製・補完しない。既存wireの数値IDを一括文字列化しない。

| 呼出し元・intent | 465参加 | 保存owner |
|---|---|---|
| 新規請求の実受領、過去未納の新受領、複合決済の各新fund | 全新fundの方法を検査 | 506、Complete最外TXは462 |
| 新しい実受領を伴う預り金・任意予約前金 | 同じ元fund作成に一回参加 | 506 fund、464口座/entry |
| 同fund8,500の請求配賦8,000＋了承預り500 | 元fund1件だけ検査、500の新fundなし | 506＋464同TX |
| 既資金配賦/解除/再配分、預り復元、旧入金の確認充当 | 新受領guardなし | 506/464の既存残高・認可・監査 |
| 実返金、金額不変の方法誤記訂正 | 新受領guardなし。方法存在/scope等の既存検査は維持 | 506。元方法と実返金方法を区別 |
| 入金なし未納確定・明細あり0円確定 | 空集合、fund生成なし | 462/505 |

成功/失敗anchor、Content/fingerprint、元target現在認可resolver、結果参照は506/464の同一operationを利用する。465専用binding・耐久command表・receiptは追加しない。C06の失敗binding元target解決やQ接続が未提供なら該当ownerの接続subsetに含める。guardだけの単体成功でconsumer稼働可能とはしない。共通Qはrequest-memoryでありreloadからcommand復元できる保証を足さない。

## 3. 停止と新受領の直列化

技術選定は既存 `payment_methods` 行のTX lockであり、新版列・新advisory namespaceは追加しない。modelにはversion列がないため存在を仮定したCASを要求しない。

1. 短いclaim TXと元target metadataは各金銭owner/C06契約どおり。
2. 最外業務TXは既存の必要な認可/ContextGate → binding `FOR UPDATE NOWAIT` → 成功receipt/unknownの再読取。この順を逆転しない。
3. 未成功でApplyへ進む場合だけ、必要な全method行を同clinic＋ID昇順で `FOR UPDATE NOWAIT`。deleted_at IS NULL、安定ID/system_keyの対応、is_active=trueを確認する。全行が揃うまでfund/配賦/口座を変更しない。
4. その後に既存の対象billing/口座/fund等のdomain lock・CASを行う。462/506/464の他lock順は維持し、必要method lockをdomain write/lockの後で追加する逆順callerは接続しない。認可gateを全domainに新規拡張しない。
5. 全fund・配賦・必要口座entry/projection・監査・成功receiptを同じ最外TXでcommit。guard内commitは禁止。

設定側のis_active変更も同じmethod行をlockし、現在master-payment-method:editとclinicを同TXで再確認して更新・監査・再読込する。名称変更/並替/削除の既存保護を迂回しない。停止writerは金銭fund/口座/bindingを後からlockする処理を持たせず、method→fund循環を作らない。複数methodに触れるwriterはID順。停止先行commitなら新受領は拒否、受領先行commitならその成立済み資金は保持し、後続の新受領だけ拒否する。「停止した時刻以前にリクエストした」という理由で例外にしない。

標準system_keyの安定識別・削除保護、固定4種の拡張/ENUM変更は本案で変更しない。現在コードの「標準行は無効化禁止」を他の停止可能行の新受領許可根拠にしない。この465案は停止可能範囲を新規拡張する契約ではない。

## 4. 結果分岐と監査

- 同key同内容の成功receiptあり: 元targetの現在認可後に元receiptを返す。methodが停止/名称変更されてもguardを再実行しない。成功結果を新しいvalidationへ置換しない。
- bindingにsticky unknown、または決着照会失敗: `result-unavailable`。同key照会、別key救済なし。busyは `operation-in-progress`、attemptを足さない。
- 同key Content不一致: 元target現在認可後 `operation-key-conflict`。現要求targetで認可を代用しない。
- 未成功の今回入力methodが停止/削除/不明: ownerのC06内 `validation`、金銭writeなし。別医院やgrant不足は非開示の現在認可規約を優先。既存fund参照だけにactive判定を持ち込まない。
- billing/口座のCAS不一致は `version-conflict`、Content一致後の他依存不一致は `dependency-changed`。method存在/activeのvalidationと混同しない。
- guardはResultResponseを直接HTTPへ公開しない。506/464の既存owner adapterにRejection/Busyを返し、当該routeの既存mappingを保持する。他finance routeの共通一括変更なし。設定PATCHはC06化せず既存ErrorResponse/HTTP conflict等を保持する。

465 guard自身は金銭監査eventを別挿入しない。506/464の成功監査に検査済みmethod IDと元fund参照を結び付け、actorは認証済みserver actorのみ。設定の停止/再有効化は既存audit_logsへresource ID=数値method ID、clinic/actor/time、before/after is_activeを同TX保存。no-opは状態/金銭版を進めず既存no-op監査方針に従う。監査失敗は全rollback。失敗を成功receipt/実受領へ記録しない。

## 5. FE・履歴・削除の必要最小境界

PaymentCardは新受領候補と保存済み資金表示を別配列にする。新候補は既存GET payment-methodsの許可されたreadを再利用し、activeかつstable ID対応可能なものだけ。通信失敗/権限不足は選択不可＋入力保持、固定4種や現金へfallbackしない。新しいmaster:view付与やaccounting権限への読み替えはしない。既存資金/receipt表示は保存済みmethod参照を使い、active候補にないことで行を消さない。

旧 `payment_method_master_repository.go` の usage は退役payments参照であり、最終fundモデルへの接続に合わせて修正が必要。削除/soft-deleteの使用中保護は、payment_funds.payment_method_id、fund_record_corrections.corrected_method_id、fund_consumption_events.actual_refund_method_id 、finance_source_events.actual_method_id を同clinicで確認する。refund_requestsの実返金event参照は対応するevent側のmethod参照で保護し、存在しないrequest.method列を作らない。金銭履歴・固定帳票が参照する方法を消してしまう逃げ道を残さない。名称から再推定/現在activeだけjoinは禁止。新受領はmethod lockを取得してから参照を書き、削除側も同rowをlockして使用確認するためTOCTOUを閉じる。既存返金/訂正もmethod存在を参照するwrite時に同じ参照保護lockに参加するが、active要求は加えない。これらのlockはfund等より前に取得する。

## 6. ①契約 → ②マスタ → ③共通UI → ④業務consumerの非循環PR境界

①本稿で内部guard shape・現在認可・method lock順・replay分岐・参照保護・caller参加表を固定する。これは共通契約の選定であり、506/464/462の実装を①完了条件にしない。共有API/DB差分の有無を台帳へ記録し、存在しない列やrouteを収録済みと書かない。

② **465-M マスタ側1PR**: master側の現在認可、停止/再有効化・使用中削除の同row直列化、最終fund/event参照による使用中保護、監査/再読込、内部method lock/guard部品とcontract fixtureを閉じる。既存GETの候補用active値と歴史参照の分離を保持する。PaymentCardなど業務側画面の接続は④へ渡す。変更対象はpayment_method_master service/repository、必要なhandler/既存型、guard部品・fixture・docs。新version列/独自receiptは不要。共有schema差分が実際に必要なら①の単一ownerへ要求し、旧payments列を復活させない。

②完了条件はmaster自身の許可された設定変更・同TX監査・読取・参照保護の実装範囲が揃い、③④へ型付きhandoffを出せること。506全実受領、464口座、462Complete、実決済は前提にしない。後段writerの全参加による並行参照保護の統合受入は④へ明示して残す。master側単独実装を全金銭7AC合格/465全体Doneとは呼ばない。試験コード準備と実行は別で、実行NOT_RUNならそのまま記録する。

③共通UIの担当へは取得成功0件/失敗、許可不足、未保存入力保持、競合表示の共通interfaceだけ渡す。③全体を465が再実装せず、PaymentCard専用の候補/履歴の意味は④で接続する。

④ **465-C 新受領consumer接続1PR**: 506の最小新fund作成/結果reader、464の同fund受領参加、462のComplete hunk、PaymentCard候補/保存済み表示分離を共有guardへ接続する。返金/訂正の参照保護lockはactive検査なしで参加。共有hunkは各ownerへ集約し、別459PR・別金銭receiptは作らない。C06 Q/元target resolver/失敗anchor/現在認可の必要subsetは各ownerへ接続するが、506全返金UIや464全業務を一括要求しない。足りないsubsetは同じ業務接続単位の先行PRとして明記し、②へ逆依存させない。

465追跡全体の完了条件は465-Mだけでなく465-Cの全新受領caller参加、履歴/返金等の非新受領保護、既存7ACへの証跡対応まで。未接続caller/実行NOT_RUN/実機未受入を明示し、helper単体やmaster設定画面だけで465全完了にしない。実行時期は後日本人指示のmainローカルを維持。

固定fixture（作成と実行を別記）: ②は設定現在認可、停止/再有効化、使用中削除拒否、方法なし/他医院、監査失敗rollback、無変更のno-op、保存後再読込、参照reader保持。④は同method既存受領あり＋追加新受領拒否、停止前選択→停止後保存、受領先行/停止先行、複合一方法停止で全rollback、成功後停止replay、失敗後停止再試行、unknown/busy/newkey禁止、既資金配賦/解除/復元/返金/誤記訂正、未納/0円no-fund、監査/receipt失敗、削除と新fund参照競合、停止後履歴/帳票/方法別集計保持。全7ACに担当段階/caller/fixtureを対応づけ、実データ・臨床値・実決済の資料待ちを技術契約不足に混ぜない。

## 7. 関連票の限定確認

369は既回答のmaster変更と保存済み明細の非自動追従・明示再適用の説明確認であり465のgateではない。351の既存236列監査は再実装しない。364/375/376はmain収録[CSV契約](emr-364-375-376-csv-column-contract-20261021.md)の選定profile/保持matrix/PR単位、365は[小数契約](emr-365-vaccine-price-decimal-contract-20261021.md)のnumeric(23,4)/文字列wire/段階互換を再利用する。本文に残る「まだ精度未選定」はこれら採択範囲では履歴。producer同版receipt/実source/pin/実applyや個別未知履歴は限定残件であり、465着手を止めない。454–456品目保険、481/482会計再適用は別ownerで二重設計しない。

## 8. 今回の①DB/API契約の判定

| 項目 | 確認済み/選定 | 未実装・未確認との境界 |
|---|---|---|
| 既決業務/7AC・単一owner | main正本と449/459/465に存在 | 試験PASS・consumer完成の意味ではない |
| 最終fund/event/correction保存形状 | main storage_finance.goに型が存在 | 465が旧writerから接続済みの証拠ではない |
| master安定ID/is_active | main payment_methods modelに存在、version列なし | 本案は新version列/DDLを選定していない |
| guardの内部shape/直列化/replay分岐 | 本稿で具体化、親レビュー採択 | 新しい公開API追加ではない。api.yaml収録済みとは記載しない |
| C06 Q/現在認可/失敗anchor | main採択契約あり | 実consumerへの必要subset実装・実RLS試験は別 |
| FE候補/履歴分離・参照保護caller | 本稿465-M/465-Cの段階別必要subset | 製品コード変更は未実施。②マスタ実装へ着手した記録ではない |

固定参照入口:
- https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/docs/spec/domains/accounting/lifecycle.md#accounting-inactive-payment-method
- https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/docs/design/shared/operation-participation.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/backend/internal/billing/payment_method_master_repository.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/755f77abd2e6d63a2fc608fcfe6a518a61afcfff/backend/internal/model/storage_finance.go
- https://app.plane.so/baritechllc/browse/EMR-449/
- https://app.plane.so/baritechllc/browse/EMR-459/
- https://app.plane.so/baritechllc/browse/EMR-465/
