# EMR-448/493 シフト複製：限定技術契約

2026-10-21。親レビュー採択の技術契約。runtime・API登録・DDL適用・試験は未実施（NOT_RUN）。現main 755f77abd2e6d63a2fc608fcfe6a518a61afcfff。CSVはPR801で文書統合済みであり、本稿はCSVを未決へ戻さずコピー固有の不足だけを閉じる。

## 正本と再利用

業務正本 [シフト](../../spec/screens/24-shift-calendar.md)の FR-shifts-copy-selected-range / FR-shifts-bulk-preview-confirm、AC-shifts-copy-selected-scope / bulk-preview-no-write / bulk-existing-conflict / bulk-protected-bookings / bulk-unaffected-history / bulk-cross-clinic / template-copy-boundary に従う。同医院・同staffの元期間→先開始日、相対日付保持、全件preview、明示置換だけ、確認範囲atomic、unknown時再送停止。元の移動/削除、別staff/他院展開、自動休日化、template改定、予約変更はしない。
[CSV採択稿](emr-448-493-shift-csv-selected-contract-20261021.md)の ShiftValues、予約impact、現在認可、全staff gate、勤務/休憩aggregate版、deferred exclusion、元actor lookup、成功bulk履歴、reconciliation証拠判定を再利用する。CSV parser/列/encoding/file uploadは呼ばない。共通基盤未実装は同PR必要subsetであり、CSV機能全体の完成待ちにしない。

## 1. 閉じた入力と日付写像

全DTOは未知/重複key拒否。ID/版は正規十進文字列1..int64最大、日付は厳密Gregorian YYYY-MM-DD。scopeは認証済医院のみ。CopySelection exact={staff_ids,source_from,source_through,destination_from}。staff_idsは非空・数値昇順・重複なし、同じstaffへだけ対応。元期間は両端込み、source_from<=source_through。先終端は元期間と同じ日数からserverが計算し、年範囲超過拒否。
元勤務は開始側JST勤務日が元期間内のもの。sourceの前日開始勤務を新たに複製対象へ加えず、他院重複検査には前日開始/翌日終了も含める。日付移動量はJST calendar day差。勤務開始・終了・全休憩の各実日時を同じ日数だけ移動し、日跨ぎ/複数日/秒以下精度を保持する。通常API同様UTC RFC3339Nanoへcanonical化。CSVの秒精度制限を既存sourceへ流用しない。元開始日と保存勤務日不整合、片方NULL、未知勤務値/休憩は拒否して訂正を別操作へ。off/paid_leaveは日時NULL/休憩空の最新契約を要求し、旧異常値を黙って消さない。
同一read snapshotで選択staff×元日付の全cellを列挙する。SourceCell={staff_id,work_date,presence,shift_id,values}、presence=present|absent、absentの後2値null、presentは既存ShiftValues（notes含む全保存値）。元空cellは先を変更しないskipであり休日/削除ではない。存在する元勤務のcopy_refは正規shift ID文字列、各先cellは同staff/写像日付。元先重複でもこの固定source全体から生成し、書込済み先を元に再読しない。notesは元保存値をコピー案へ表示し明示確認する。sourceの旧長文はCSV setの新規入力制限で切捨てず、既存値コピーとして保持する。
bounded技術上限は選択元cell（空cell込み）500、各勤務休憩50、canonical command/JSON各4MiB。超過は413でゼロwrite、自動切捨て/分割なし。範囲選び直しは別preview。既存業務の勤務総数上限ではない。

## 2. Routesとpreview/apply

実prefix /api/v1、以下相対path。既存CSRF/authを維持、新grantなし。
POST /shifts/copies/preview：body exact={selection:CopySelection,replacements:[{copy_ref,shift_id}]}。
POST /shifts/copies/apply：body exact={operation_key,confirmed:CopyProjection,preconditions:CopyPreconditions}。
GET /shifts/copy-operations/{operation_key}：C06結果。
GET /shifts/copy-operations/{operation_key}/details：成功履歴。
GET /shifts/copy-operations/{operation_key}/reconciliation：例外証拠read。
GETはintent=staff.shift_copy.apply固定、余剰query拒否、元actor/clinic/保存対象と現在shifts:viewを検証。previewはshifts:view、applyは各createにshifts:create、replaceにshifts:edit、全noopにもshifts:edit。対象staff全員の現在所属/読取を確認し、予約詳細は別予約viewで非開示を守る。

CopyProjection exact={schema:"shift-copy-preview/v1",selection,source_cells,rows,impact}。source_cellsはstaff/date順。rowsはpresent sourceごとに{copy_ref,source_work_date,target:{staff_id,work_date,presence,shift_id},before,after,resolution}。before/afterはCSV ShiftValues、absent targetのbefore/IDはnull。resolution=create|replace|noop。異なる既存targetはreplacementsの実ID明示と一致しなければreadyにしない。noopは保存値全体一致、単なる時刻一致ではない。impactはCSVの完全な保護予約集合と同じclosed shape。元/先勤務版はProjectionへ入れずPreconditionsへ。
CopyPreconditions exact={source_cells:[{staff_id,work_date,presence,shift_id,expected_version}],target_cells:[同形]}。空source含む全cellとpresent sourceから作られる全targetを厳密に覆う。present版は親aggregate版、absentID/版はnull。重複元先cellの期待値は同じsnapshotの同一実値に一致必須。
preview response exact={projection,preconditions,preview_digest,ready,rows,errors,cross_clinic_conflicts,impacts}。projectionは未解決時null、digestもnull、ready=false。rowsは{copy_ref,source_work_date,target_work_date,existing,proposed,action}、action=create|noop|replacement_required|replace|skip|invalid。空sourceはcopy_ref=null、skip、proposed=nullで全件表示。errors={copy_ref,staff_id,source_work_date,target_work_date,field,code}（不明値はnull）、code=limit|identity|duplicate|value|scope|source_unresolved|leave_unresolved|leave_conflict|cross_clinic_overlap|booking_protected|replacement_required。cross_clinic_conflictsはCSV形のshift_refをcopy_refへ、source_recordsをsource_work_dateへ置換し、available/restricted/null/実交差時間/503条件は同じ。impactsの予約詳細非開示とready=falseも同じ。
preview_digest=SHA256(UTF8(JCS(CopyProjection))) lowerhex64。previewは業務/claimともwrite0。入力構文不正400、認証401、認可403、対象非開示404、サイズ413、依存読取失敗503。入力差分・置換待ちは200 ready=falseで成功保存と混同しない。全応答no-store。

## 3. 元snapshot・耐久command・再送

owner_kind=staff、intent=staff.shift_copy.apply。Content exact={schema:"shift-copy-command/v1",preview_sha256}。immutable typed command exact={schema:"shift-copy-command/v1",projection:CopyProjection}。[共有Q](../../design/shared/operation-participation.md#verified-command-selected-20261021)がdeep-copyしclaim前/Apply前に再hash一致を検証、Applyはreq.VerifiedCommandのみから復元する。hashから原文を復元可能とは扱わない。
新keyはsource/target/現在認可をserverで読み直し、提出confirmedと完全一致する場合だけclaim。client提供元値を信頼せず、古いpreviewはpre-claim409/zero-write。metadataはCSVで選定したshift_bulk_operation_targetsを再利用し、copy schemaを追加許容する必要subsetを同PRへ含める。新表/別key台帳を増やさない。保存列・PK/FK/RLS/immutable/hash/4MiB上限はCSV同一。staff_idsはselectionに一致、from_date/through_dateは元期間と計算先期間を覆うmin/max。認可はこの包絡日付を全対象と誤認せず保存commandの正確な元/先集合から行う。
winner-only ClaimMetadataWriterで短claim TX内bindingとmetadataを同時保存。失敗は両rollback、loserは元metadataを復元しincoming集合で補修しない。canonical bytes/hash/Content/schemaとscopeを検査。欠落anchorは403非開示、元対象認可成立後の破損は503 result-unavailable。
既存keyは元actor/選択医院/保存commandから元対象を現在認可→意味値一致→成功receipt/replay又はunknown停止→未成功時のみ再認可・CAS・完全before/元snapshot/予約影響照合。source内容変更/追加削除・target identity/値変更はdependency-changed、意味値不変で版だけ相違はversion-conflict。同Contentの版だけ更新は差分再確認後同key可。source変更を最新master/templateで再計算して同keyへ上書きしない。意味内容変更は旧unknown解消後新preview/newkey。
FEはdraftをメモリのみ、scope切替/失権で本文破棄、遅延応答拒否。再mountで元key又は送信有無を特定不能なら停止し新key救済なし。既知keyはlookupを先行し、成功detailsを表示、未成功retryには元確認commandが必要。private commandをブラウザへ配る新APIは設けず、本文を失った場合は自動再構築/自動再送しない。

## 4. 同TX・履歴・結果

CSV§6と同じgraph→全staff昇順UPDATE NOWAIT→assignment→binding→元/先cellと全breakのstaff/date/ID順を使う。重複cellを一回lock、source/target不存在もstaff競合点で保護。予約staff SHARE、他院勤務/休業/所属/能力writerの全必要参加、版とbreakだけ変更、同値古版、deferred exclusionを同じvalidator/ownerへ接続する。元sourceだけのcellを書かず、元先重複cellだけ明示target置換可能。全最終集合を検査して各勤務一回更新、途中commit/NULL退避/旧単日APIループ禁止。
保存は全勤務/休憩→既存shift_bulk_operations(kind=copy)→audit→C06receiptの単一TX。source_snapshot={schema:"shift-copy-source/v1",selection,source_cells,impact}、confirmed_rows={copy_ref,source_work_date,action,shift_id,version,before,after}の全rows配列。全noop/全元空でもbulk履歴とreceiptは一組、勤務版/変更auditゼロ。元空はconfirmed_rowsへ架空勤務を作らずsource_cellsで証明する。
refs={shift_bulk_operation_id:bindingID}、versions省略、EffectSummary内部 `{changed:"true"|"false"}`（値は文字列）。C06既存12status/HTTPはCSV§4をそのまま使用、新statusなし。details成功exact={status,operation_key,result_refs,recorded_at,source,scope,rows}、sourceはsource_snapshot、scopeはselection、rowsは保存confirmed_rows、committed_versionsなし。後日のcurrentを混ぜない。非成功はstatus/operation_keyだけ。
audit action=staff.shift_copy.apply/resource=shift_bulk_operation/resource_id=bindingID、actor/clinicはtrusted。old/newはCSV§7と同じ勤務配列、metadata={operation_id,preview_digest,source_from,source_through,destination_from,staff_ids,created_count,replaced_count,unchanged_count}。source原文をlogへ重複しない。失敗/audit/receipt/constraint不整合は全rollback。
reconciliationはCSV§4.1のproof条件とclosed top-level 5keysを継承し、rowsのshift_ref/source_recordsだけcopy_ref/source_work_dateへ置換。選択present source全件を列挙、空sourceは変化対象ではなく除外する。全元空は整合成功receiptがあればverified_all_committed（rows=[]）、receiptなしはundeterminedで空集合から未実行推定しない。自動再実行/補償は常にfalse。元actor一致・現在view・証拠不足unknownを維持。

## 5. 1PR責務と受入

448統合ownerがcopyの契約→api.yaml→staff BE/FE→fixtureを1PRで回収。493は日時保存/休業/履歴、489はQ/claim/ContextGate必要subset、446/466はCSV共通metadata schema追加と版/exclusion、467は予約保護writer閉包だけ。CSV parser完成・全staff CRUD・全予約Doneをgateにしない。shared実装が未提供なら同PR依存patchとして回収し、空stubでactivateしない。
fixture: 元先重複非連鎖、空source非削除、同名staff非推測、元期間外保持、日跨ぎ/閏日/年越し/精度保持、source変更/削除/新規出現、target新規出現、break-only版、明示置換、全範囲rollback、他院衝突/失権/休業unknown/進行中超過保護、replay/異Content/unknown/元anchor、失敗監査、全noop/全空、reconciliation証拠不足。実行はNOT_RUN。

## 241優先関係の軽い追補

EMR-241の2026-10-02/04実装・受入記録は履歴として保持する。今回の勤務CRUD/コピー/CSVは最新24-shift-calendar.md・PR740の勤務wireに従い、off/paid_leaveを終日非勤務、日時NULL・休憩空として扱う。過去の「offでも時刻保存可」を新consumerの目標仕様にしない。日付なしtemplate定義は別物で、勤務型への一律timestamp化・旧Done取消・実データ消去はしない。

## 段階境界（①→②→③→④）

①は本稿のDB/API・予約保護interface/validator契約の確定。②マスタ・勤務入力は既存予約を壊さない最小の現在保護reader/同時更新gate参加までを必要subsetとして接続する。③以降の段階と④新予約consumer機能全体の完成を②の開始・PR完結gateへ戻さない。既存予約保護を読むinterfaceは現在の保護対象集合と現在適格性を同TXで返し、未知・取得不能は保存停止とする。必要な既存予約writerの競合点参加がない場合だけ当該新copy/CSV保存を未活性にし、別のマスタread/設定やfixture準備まで全体停止しない。④の新予約入口・新業務フロー接続とその受入は別完了条件。安全な保護interfaceを空stubにすることは段階分離ではなく禁止する。
