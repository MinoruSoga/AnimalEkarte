# EMR-568 後日予定reportの限定技術契約

2026-10-24、固定main `4b7ea65a3b692073bb41c4be52bd2c1a9be99d74`。既決D03/D21の後日予定参照を実現する設計引渡しで、実装/供給済みではない。[518保存契約](../../../work/emr/emr-518-storage-contract-20261020.md#51-medical_record_plan_items)と[516 owner-link契約](../../../work/emr/emr-516-517-plan-owner-link-contract-20261021.md)を再利用する。実装受領先は親452の[EMR-568](https://app.plane.so/baritechllc/browse/EMR-568/)。本契約のmain採用後、Planeに採用SHAを固定して着手プランの採用を記録する。進行状態はPlaneを正本とし、設計採用と製品供給・受入を区別する。

## canonical sourceと対象

medical_record_plan_itemsのid/clinic_id/medical_record_id/pet_id/version/item_kind/name_snapshot/execution_domain/clinical_state/removed_atを読み、元recordの現在owner/pet/clinicを照合する。現在clinical＋deferred＋removed_at IS NULLだけ表示。performed/cancelled/pending/commercial/removedを未消化へ変換せず、過去deferred履歴一覧は導入しない。再取得で現在対象外となった行を除き古いdeferredを最新成功として残さない。結果linkからstateを推論しない。

D21§3.1の診療終了限定は前回診療/診療履歴sectionだけ。後日予定は別参照で終了済みnormal限定・再開中/過去入力除外をしない。495終了resolver依存なし。516§8の旧終了条件は本reportへ適用せず、元record/planの現在read認可と医院/飼主/患者対応を保持する。

518のclinical pending→wire not_performed、performed/deferred/cancelled同値、commercial/NULL→selectedを維持。旧treatment_plans content/価格、treatments.status/is_selectedからdeferredを推測しない。共通3表は現main未供給であり516の必要schema/model/current readerを同担当sessionのproducerへ含める。report専用DB表/state writer/copy stateなし。本人の2026-10-24 11:47:07「DBは作り直し前提」「old_db mappingも確定仕様へ改修」に従いcanonical再構築DBを前提とし、legacy救済/fallback/coverage metadata/backfillなし。old_db改修・DB再作成許可や保存原本file保護の撤回ではない。

## read wire

GET `/api/v1/medical-records/deferred-plan-report` を/:idより先へ登録。queryはowner_id/pet_id（canonical positive decimal string、1..signed64 max）、初回省略cursor。clinicはtrusted選択scopeのみ。medical-records:view＋owners:viewの既存staff-only、owner/pet/record/planの現在read/scopeを毎request検査。edit/session/C06 key不要、no-store。

response exact `{schema_version:"1",owner_id,pet_id,acquired_at,items,next_cursor,page}`。item exact `{plan_item_id,plan_version,medical_record_id,medical_record_date,item_kind,name,state}`。ID/版decimal string、dateは実YYYY-MM-DD又は正当null（日付不明表示）、kindは518 typed、nameは同版snapshot、state literal deferred。登録日補完・未照合stateの架空deferredなし。403/404は既存認可error、schema/source未供給・必須field/link不整合・取得障害503。空成功へ変換せずSQL/他院情報非公開。必須identity/state/版欠落は正常nullでない。

## 有界paginationと非漏えい

lookaheadを含め最大51候補を取得・検証し、最初の最大50候補を処理して最大50itemsを返す。51番目はhas_more用だが認可/必須field検査の対象であり、不正ならpage全体をfail closedにする。表示50件を埋める無制限scanなし。空itemsでも未走査があればnext_cursorと「未確認の予定があります」。次pageは利用者の「続きを表示」だけ、自動全件取得なし。全体件数/日数の業務上限は設けない。

seedはtrusted clinic＋要求owner/petの現在access。不認可403/404なら候補照会前に停止。初回upper_plan_item_idはWHERE clinic_id=currentClinic AND pet_id=selectedPet ORDER BY id DESC LIMIT1だけで取得し、record/owner JOINや全record認可filterなし。(clinic_id,pet_id,id)索引末尾1件を読む。upperは初回/毎pageの同じ短read snapshotで元record現在認可・owner/pet/clinic対応を検証してからcursorへ入れる。真に0候補ならitems=[]/next_cursor=null/has_more=false。継続upper削除/読取不能は一般再読込要求又は既存403/404、ID/cursorをechoしない。必須linkage破損503と認可失敗を分ける。

第一段階はWHERE clinic_id=currentClinic AND pet_id=selectedPet AND id>after_id AND id<=upper_id ORDER BY id ASC LIMIT51だけ。record/owner/state filterより前に最大51候補を確定。第二段階は**lookaheadを含む全51候補**の元record/現在owner-pet-clinic/read認可/必須field/stateを同じsnapshotで検査する。真正な拒否が一つでもあれば一般403/404、必須不整合503でpage全体fail closed、部分items/cursorなし。全検査後最初の最大50候補を処理し現在deferredだけ投影、51件目はhas_more専用。skip可能なのは認可済み状態不適合行のみ。認可外本文/名称/日時/ID/件数をitems/error/cursorへ出さない。状態除外を埋める追加batch走査なし。

afterは最後の表示行でなく処理した最大50候補末尾IDへ進め、lookaheadへ進めない。候補0ならcursor=null、入力afterをechoしない。状態除外による空itemsでもhas_more/cursorを返し古い認可deferredまで進める。認可除外の空成功は作らない。転送順idは業務優先度/診療順でない。

必要索引medical_record_plan_items(clinic_id,pet_id,id)は516/446単一owner。recordは最大51＋upperから重複排除したIDだけ既存PKで取得。全patient record/plan一覧・版vector・legacy fingerprintのmaterialize/hash、全件COUNT、JOIN先全件認可filter、履歴全取得はしない。作業上限はseed、初回索引末尾1、最大51候補record照合＋毎page upper1件lookup/record照合。重複recordは排除可だが上限外の走査不可。

cursorはbase64url JSON exact `{schema_version:"1",owner_id,pet_id,upper_plan_item_id,after_plan_item_id}`。strict key/type/size/正整数/after<=upperを検証、clinicは受けずtrusted scopeのみ。毎回再認可しcursorを認可能力としない。FEは同clinic/owner/petの連続chainだけ保持、切替破棄。他patient cursor400、改変upper/idでもscope外を読めない。base64urlは暗号化/秘匿でないためupperとafter/lookaheadを含む全候補の現在認可成功が返却条件。拒否/不整合/upper消失では入力cursor/IDをerrorへechoしない。署名key/credential/新ACLなし。

page exact `{has_more:boolean}`。scanned_count/除外件数/認可外件数/totalなし。cursor位置を画面のrecord/plan識別表示に使わない。acquired_atは各snapshotのserver UTC RFC3339。「取得時点：...」「追加取得分は各取得時点」と表示し、全collection atomic snapshot/常時最新/全変更検出を保証しない。upper後の新規や取得済pageのstate/認可変更の検出を約束しない。明示再読み込みで全page/cursor破棄、検知したscope/失権/source失敗では旧値を最新成功としない。未知全変更を409検出すると約束しない。

has_moreなら空itemsでも全件なし表示不可。正常sourceで終端まで照合したchainだけ「取得時点の対象予定なし」。未供給/不整合/取得失敗は503で架空canonical空表成功なし。正常0と取得不能を分け、legacy全体走査の証明adapter不要。

## 実接続と同一実装PR

既存接点はmedicalrecord/routes_records.go、treatment_plan_repository.go、owner-report/hooks/use-owner-clinical-briefing-data.ts、components/OwnerClinicalBriefing.tsx、owner-report.css。既存入院/legacy CRUDは置換しない。新設予定はmedicalrecord/deferred_plan_report_{handler,service,repository,response}.go、owner-report/api/get-deferred-plan-report.ts、components/DeferredPlanReportSection.tsx。route→scope→source→page→UIまで閉じる。

OwnerClinicalBriefing履歴panel直前「後日予定」sectionに元record ID/診療日/元plan ID/名称/後日予定を文字表示。通常button、読み込みrole=status/aria-live=polite、失敗role=alert。新link/modal/自動focus/印刷/共有なし、予約/受付/次回来院へ展開なし。

516がschema/current reader単一owner、新子担当sessionが必要producer→read→FEを同PR担当。445 API/router、446/466必要共有hunkだけ調整。既存subset受領は固定採用SHA/単一editorを記録。495終了resolver、517/500/484全部、全package/494/会計/出庫Done待ちなし。producer未接続mockを完成とせず、未知状態をfalse/空配列で隠さない。文書採用後の固定SHAを票へ記録しReady判定、runtime供給/完成とは別。

fixtureは51/101件page境界、同日同名別ID、deferred→performed/cancelled/removed、再開/過去入力、scope切替/失権、page間変更/取得時点表示/refresh、schema/必須field/source503、正常0対未供給、状態除外連続空page、LIMIT51先行、upper拒否/削除/破損、最初50/51番だけ拒否、部分items/cursor非返却、本文/日時/ID/件数/入力cursor error非公開、51＋upper1有界、null日付、古い予定、read zero-write、keyboard/status/alert。全実行NOT_RUN。
