# EMR-573 他ペット未納ヘッダーの限定read契約

2026-10-24、基準main `4b7ea65a3b692073bb41c4be52bd2c1a9be99d74`。親452、[実装受領票573](https://app.plane.so/baritechllc/browse/EMR-573/)、個人割当なし。対象はAC-medical-records-other-pet-unpaid-header-entry/scopeだけ。選定技術契約であり実装/受入済みではない。canonical再構築DBだけ、旧schema/payments fallback/backfill/旧調査なし。新金銭writer・債務配賦規則・新権限を作らない。

## 採用APIと現在scope

GET `/v1/accountings/other-pet-unpaid`（axios既定baseURL=/apiでは外部 `/api/v1/...`、独自baseURLは既存設定）に固定。query owner_id/current_pet_id/view=status|claims/cursor?/limit?、clinicをquery/bodyに置かずカルテclinicの明示X-Clinic-IDを既存membership/grant middlewareで認証contextへ解決。ResourceAccounting/accounting:viewのroute middleware＋RequireSelectedClinicGrantを併用。

同read snapshotでowner/current petのclinic内存在・非削除・現在owner一致を検査。カルテowner/pet未確定はrequestなし。請求は常にbillings.clinic_id=authorizedClinic AND owner_id=requestedOwnerで相関。元請求帰属を他petの現在owner JOINで消さず、譲渡先へ旧債務を見せない。current pet除外はIDとsnapshot元明細帰属、billings.pet_id単独/名前/現在ownerだけで代用しない。snapshot/line/fund/event/deposit全段階でclinic/owner相関、跨院・全社集計なし。

## 入力・DTO・障害

owner/current petは正の十進ID string、空/0/重複query/非整数/範囲外400。view既定status、他値400。statusはcursor/limit禁止。claimsはlimit既定20・1..50、cursorはpaddingなしbase64url UTF-8 closed JSON `{v:1,clinic_id,owner_id,current_pet_id,after_billing_id}`。IDは正規正数十進string、余分key/encoding不正/1024bytes超/別scope400。暗号化/署名/session storeなし、cursorを認可能力とせず毎read再認可。afterは処理済み認可済み候補末尾のみでlookahead/認可外IDなし。start/end date・clinic_ids・任意sort・金額filter禁止。全期間の現在有効請求を読み当月検索/FE全件巡回へ置換しない。GETのみでoperation ID/key/資金token/write receiptを発行しない。

200共通は `{scope:{clinic_id,owner_id,current_pet_id},status:present|none|unknown,unknown_reason}`。unknown_reasonはunknown時だけcanonical_projection_incomplete、他null。status viewはこのscope/三値だけで金額/件数/支払指示なし。claims viewは追加 `{claims:[{billing_id,claim_version,calculation_id,source_pet_ids,assessment:unpaid|unknown}],next_cursor}`。ID/版string、不明claim_version/calculation_idはnull、unknown source_pet_ids=[]可だが共通料金確定と表示しない。source IDsは確認済元帰属だけ。資金ID/金額/件数/owner又はcurrent pet総額/支払方法/追加個人情報なし。

claimsは他petとの関係が確定した未納又は安全に判定不能な参照候補。draft/cancelled/確定残0/current-pet-only確定は除外。billing_id昇順keyset、page跨ぎ同snapshot保持なし、各応答global statusは同read snapshotの評価から返す。先頭空/途中page/50件打切りはnone根拠でない。選択時再取得、消えた候補を選択済みにしない。

401未認証、403選択clinic accounting:viewなし、404owner/pet不存在・削除・現在組合せ不一致で越境存在非開示。正当scope確立後のcanonical claim/insurance/source/fund解決不能だけ200 unknown。DB/timeout/中断/decode不能は5xx/networkでnoneにせず、確認不可＋retry。private/no-store、Query keyはclinic/owner/current_pet/view/cursor/limit、切替時前scope placeholderなし。

## 同一snapshotの限定canonical producer

505/506の既決事実を束ねるadapterを同feature PRで実装する。既存producerが採用済別SHAなら正確なSHA/契約/fixture receiptで置換可、全505/506 Done待ちなし。

| 内部型 | 読む事実 |
|:---|:---|
| ClaimIdentity | clinic/owner/billing/claim_state/claim_version/current_calculation_id |
| ClaimReceivable | 同current snapshotの保険確認後の飼主請求額（整数円）又はunknown理由。rule_version/条件/保険確認/line/source版照合 |
| ClaimFulfillment | 当該請求への有効支払充足（整数円）又はunknown。allocate/deposit_allocateと元参照release/deposit_releaseの有効関係 |
| ClaimPetAttribution | current snapshotと版照合した元source_pet_id集合、unattributed/common-charge有無 |
| Outstanding | 両金額knownだけ既決max(ownerReceivable-effectiveFulfillment,0) |

raw total_amount/全受領/owner全額/未充当残を式へ代入せず、別時点cacheと合成しない。505 versioned decoder/fixtureを接続し正規snapshotがknownになる実装receiptを要求、全件unknownで完成不可。未確認JSON key/0既定を発明しない。fundは元受領/有効訂正/event graph。受領だけで充足とせずrelease元参照/clinic-owner/上限/重複検証。返金予定は充足を減らさず実返金も無条件二重控除しない。請求帰属はcanonical解除/配賦に従い預り受領と充当の二重計上なし。取消元はactive未納から除き、新番号再請求は別current claim、元番号/帰属不変。

### 固定read budgetと三値

statusは認可clinic/owner候補をbilling_id ASC LIMIT51（draft/cancelled除外、不正stateは除外しない）、先頭50評価・51番目は未評価あり証拠。claimsはcursor後LIMIT limit+1（最大51）で先頭limit評価。current-only/完済を後処理除外して空でもcursorで次へ到達可。業務上の請求数上限でない。claims global statusは同TXの同じ先頭50評価を再利用でき、そのpageに確実presentがあれば強める。未評価が残るのにnoneへ弱めない。

各claimのline/元帰属/配賦・解除・訂正・deposit各collection LIMIT1001、最大1000要素、graph visited nodes合計1000。循環/超過/非一意currentはunknown。全履歴hash/無限再帰/FE自動全巡回なし。既存cancel＋read TX2秒deadline、timeoutは失敗。scope indexを使いDB全取得後sliceだけでboundにしない。lookaheadを含む返却候補の認可を検査しscope不正混入はgeneric failure、refs/cursorなし。

finalizedかつoutstanding>0で対象料金全てが確定したcurrent pet以外の元petに属する請求が一件あればpresent。既存D10が未納のpet帰属を解決していれば利用可。混在残額へ独自按分/FIFO/比例配賦なし。current＋other/common/不明混在で他pet残帰属未確定はunknown候補、全額>0だけでpresentにしない。current-only確定は除外。

確実presentが別請求にあればunknown併存でもglobal present。presentなし＋関連unknown/未評価ありはunknown。全対象評価がbudget内完了しpresent/関連unknownなしだけnone。不正/null claim state/current snapshot欠落も現在canonical不整合としてunknown、旧救済ではない。共通料金を架空petへ割当てない。

## ヘッダー→全期間候補→元請求詳細

scope確定＋accounting:view時だけstatus取得、権限なしは入口非表示/存在非漏えい、再認可中に前結果を使用しない。presentは「他ペットの未納を確認」、noneは入口なし。unknownは存在を断定せず「未納情報を確認」とretry、同scope全期間claimsへ進め評価上限後へ到達可能にする。障害は確認不可/retry、loadingは確認中で非操作。金額/件数なし。type=button、既存dirty離脱guard尊重、save/請求create/入金/全件選択/自動精算なし。

遷移は `/accounting?tab=unpaid&scope=other-pet&clinic_id={C}&owner_id={O}&current_pet_id={P}`、URLSearchParamsでencode。queryはcontextで権限でなく、clinicをglobal選択へ保存/切替せずAPI X-Clinic-ID明示。外部return URL/帳票/amount/資金payloadなし、任意route stateは `{fromMedicalRecordId:string}` のみで認可根拠でない。

UnpaidTabはscope=other-pet限定分岐、不足/不正contextは「対象を確認できません」で停止し通常一覧へ拡張しない。この分岐だけ当月filter/legacy unpaid queryを使わず同canonical API view=claims全期間。通常tabは不変更。owner/current pet scope固定表示、候補自動選択なし、unknownは要確認で未納額/回収選択確定としない。

明示選択時は現在scope/grant/current claimを再検査、削除/取消/完済/帰属不一致は停止し更新、403/404をnoneにしない。確認した元billing IDで `paths.accounting.detail.getHref(billingId)`＋URLSearchParams `{scope:"other-pet",clinic_id:C,owner_id:O,current_pet_id:P}` へ進む。/accounting/select-pet・newなし。詳細receiverもclosed contextを再認可し既存read/write hooksへ同明示clinicを渡しglobal fallbackなし。owner/対象不明は操作停止、query/stateの金額/版をcommit認可に使わない。このreceiverも同PR供給。

候補は最初に参照する請求の選択だけ。既存精算の同owner複数pet・当日/過去請求選択・自由支払金額を制限しない。headerのcurrent-only除外を精算全体filter/回収上限へ適用しない。詳細readがlegacy依存なら到達identity/current canonical表示adapterを同PR供給し旧表復活なし。既存writeのgrant/確認/最新残・版検査は維持し新write能力なし。

## 一つの実装PR・根拠

BE限定producer/service/repository/versioned decoder/同snapshot/三値/budget/keyset→route/handler/DTO/grant→FE api/runtime DTO/query/明示clinic/失敗/切替→header/離脱→既存UnpaidTab/詳細receiverを同PRで閉じる。shared headerは単一統合編集窓口で個人名割当なし。別featureと同shared PRでもAC/checklist/producer receipt別保持。mock/全unknownで完成とせず実producer→header→同scope候補→元:idの接続証拠、実装・検証SHA/receiptが必要。新金融ルール/全D10移行/旧救済/死亡後新診療/自動保存を混ぜない。現在残は実装/検証、業務採否の再質問待ちでない。

固定根拠：billing/routes.goとaccounting_handler.goのaccounting:view/selected grant、billingsのclaim_state/version/current_calculation_id/元owner、calculation_snapshots/linesの版付き計算、billing_items.source_pet_id/source_record_id/contract_version、storage_finance.go/fund_consumption_eventsの資金graph、[会計技術正本](technical-spec.md)。fund.versionを発明しない。旧get-owner-unpaid-balance.tsの金額/件数/0補完を再利用せずexplicit X-Clinic-ID方式だけ継承。accounting-routes.tsx/UnpaidTab.tsx/accounting-list-model.ts/AccountingDetail.tsxが実接点だがscope receiverは未実装でlegacy稼働証明でない。

## 必須fixture（全NOT_RUN）

OP01別pet確定未払/currentのみ/完済/なしの三値。OP02別owner/clinic/owner mismatch/削除/失権/医院変更と非漏えい。OP03譲渡後の元請求owner保持と新ownerへの債務非開示。OP04請求10000・充足3000、10000充足後9000訂正（返金必要は別）、3000充足後9000訂正。金額はfixture内部のみ。

OP05未充当/預り/解除/訂正/返金予定/実返金の二重控除防止。OP06保険不明/snapshot欠落版不一致/source/rule/event不整合はunknown。OP07current+other/common/不明混在と独自按分なし。OP08draft/cancelled/再請求のidentity独立。

OP09古い未納/51件目該当（先頭で証明不可ならunknown入口からpage到達）/空page/1000超/循環/timeout、page件数からnone推測なし。OP10切替遅延/失権/refetch/networkで前scope又は成功none表示なし。OP11GET/navigation以外なし・dirty暗黙保存/create/自動選択なし。OP12query/cursor改変不足/候補後状態変化で現在再検査・元ID保持。OP13通常tab/既存精算grantと複数pet/当日過去/自由金額を維持、scope分岐で退役schema依存なし。OP14再構築canonical fixtureでproducer→handler→FE→既存:id receiver結合。BE/FE/DB/E2Eは全NOT_RUN、静的確認をPASSにしない。実装/生成/DB/試験/実データ操作は行っていない。
