# EMR-448 / 493 シフトCSV取込の選定技術契約

2026-10-21。確定業務仕様を満たす実装引渡し用の技術選定。文書のみ、runtime/API登録/DDL適用/試験は未実施（NOT_RUN）。CSVの物理形式と技術上限の選定を、本人による列名・物理方式の直接指定とは記載しない。

## 1. 正本・対象・現在との差

業務正本は[シフトの追加合意・FR/AC](../../spec/screens/24-shift-calendar.md#shift-copy-csv-agreement-20261015)、[D14業務](../../spec/domains/staff-workforce/README.md)、[D14一括保存契約](../../design/domains/staff-workforce/technical-spec.md#shift-bulk-operation-contract)。[493休業契約](emr-493-staff-leave-selected-contract-20261021.md)の期間・unknown・所属・予約保護を維持する。Plane [448](https://app.plane.so/baritechllc/browse/EMR-448/)本文/全4コメント、[493](https://app.plane.so/baritechllc/browse/EMR-493/)本文/全6コメントを読取。448は残差とCSV実装PR範囲、493はstaff勤務writer/receiptの必要subsetを所有する。

確認mainは `9a1d50f9dcc7d275456f1214507ef0a55441ac98`。以下は実在確認と追加選定を分ける。

- `shift_entries`は医院/staff/勤務日が一意、`starts_at/ends_at`とversionあり。`shift_entry_breaks`は日時の子集合。全staffの日時範囲重複を拒否するexclusionあり。日跨ぎは日別rowに分割せず、勤務開始のJST日付をdateへ保持する。
- 既存`shift_bulk_operations`はclinic/operation、kind=copy|csv、source_snapshot、confirmed_rows、preview_digestを保存し、immutable/RLS/成功receipt制約を持つ。これを成功業務履歴として使い、別のbulk履歴台帳を増やさない。
- 現`shift_request.go`は旧start_time/end_time・break_start/break_endを読む。最終日時modelへの移行不足を仕様として継承しない。staff row→assignment lockの既存競合点を再利用するが、全CSV/予約writer閉包が実装済みとはしない。
- [共有Q](../../design/shared/operation-participation.md#verified-command-selected-20261021)は採択済み・Request/Runnerの限定実装依存。既存Content-only consumerを改造せず、Qなしで別captured DTOをApplyへ渡す代替実装をしない。

今回の完成単位はCSVの解析・全件preview・明示置換・一括apply・同key結果照会・履歴再読の一貫PR。514日時CRUD、241時刻template、493休業の既存契約は維持。コピーのroute/全面実装、CSVから職員/医院/所属/能力/template/休業/予約を作成変更する機能、`.xlsx`、給与計算・移動時間計算・外部送信は追加しない。共通validatorを将来copyでも使えるが、CSV PRにcopy完成を要求しない。

## 2. 単一profileとファイル境界

profileは `shift-csv-v1`。拡張子`.csv`、`text/csv`又はExcel送信時の`application/octet-stream`を受理し、実bytesを同じstrict parserへ渡す。JSON/ZIP/XLSXを拡張子やContent-TypeだけでCSV扱いしない。encodingは利用者が `utf-8` 又は `windows-31j` を明示選択する。自動判定・失敗時の別encoding fallbackはしない。

- UTF-8は先頭BOMだけ任意、残りはstrict UTF-8。Windows-31Jは明示decoderを固定し、不正列・置換文字を伴う変換・非可逆な変換を拒否する。選択encodingとdecoder/profile版をsourceへ記録する。UTF-16、Excelシリアル日付、地域依存区切り/数式評価は非対応として元行エラーにする。
- カンマ区切り、double quote引用、引用中のquoteは二重化、CRLF又はLF改行、最終改行は任意。notes内の引用改行は保持。NUL、不正Unicode、壊れた引用、列数相違、未知/重複/不足headerを拒否。セルの一律trim、Unicode正規化、数値化・指数変換はしない。空行は無視せず空recordエラー、末尾改行は空recordに数えない。
- 元fileは最大2MiB、headerを除く最大2,000 logical records、勤務groupは最大500、各勤務の休憩は既存入力と同じ最大50。multipart全体とcanonical commandは各4MiB上限。UTF-8 decode後も4MiB以内。制限はこの一括入口の技術上限で、勤務総数・業務上の休憩回数の新制限ではない。超過時は全体413/解析エラーでゼロwrite、黙った切捨て・自動分割をしない。利用者が範囲/ファイルを選び直して再previewする。
- JSON options/metadataにも既存raw body limitを適用し、binary mediaによる上限迂回を許さない。CSV・notes・filenameをログ/URL/監視へ出さない。filenameは表示用メモリだけ、path/object keyに使わない。

### 2.1 exact headerと各record

順序固定の9列（ASCII）:

```csv
record_type,shift_ref,staff_id,work_date,shift_type,starts_at,ends_at,notes_mode,notes
```

| 列 | shift record | break record |
|---|---|---|
| record_type | literal `shift` | literal `break` |
| shift_ref | file内group識別 `s[1-9][0-9]*`、重複不可 | 対応shift recordのshift_ref。前後の行位置で推測しない |
| staff_id | 既存staff ID、正規十進文字列1..int64最大 | 空必須 |
| work_date | Gregorian `YYYY-MM-DD`、勤務開始のJST日付 | 空必須 |
| shift_type | `full/morning/afternoon/off/paid_leave` | 空必須 |
| starts_at / ends_at | 勤務系3typeでは双方必須。off/paid_leaveでは双方空 | 双方必須、対応勤務の内側にある休憩日時 |
| notes_mode | `keep`又は`set`を明示 | 空必須 |
| notes | keepなら空必須、setなら空文字を含む最大2,000文字 | 空必須 |

日時は `YYYY-MM-DDTHH:mm:ss+09:00`（秒00..59、年0001..9999、端数秒/24:00/別offsetなし）。JSTを明示したCSV用輸送文法であり、通常APIの受理文法を変更しない。勤務end>start、日跨ぎ/複数日を時刻逆転から推測せず双方の実日付で表す。勤務日とstartのJST日付不一致を拒否。breakも同じ実日時、端点は[開始,終了)、勤務内・相互非重複を既存validatorで確認し、並べ替えで重複を隠さない。勤務typeから時間を補完せず、templateの時刻をここで暗黙採用しない。

off/paid_leaveは終日非勤務で勤務日時NULL・break0件。CSVに存在する明示行だけがその状態への案で、空欄・未掲載勤務を休日/有休や削除へ変換しない。勤務行のnotes_mode=keepは既存notes保持（新規時は既存owner初期値の空文字）、setは明示入力の値を採用する。set+空だけは明示空値であり、空欄単独を解除命令にしない。勤務rowの置換を確認する際にはこのnotes差分も表示する。

shift_refは元file内の対応用で、DB ID、operation key、別のidempotency keyではない。同staff/work_dateの複数shift recordは同値でもエラー。同じgroupのbreak複数行は許容し、shift欠落・不明ref・offへのbreak・同一break重複は拒否。staffは現在選択医院の実staff/有効assignmentで解決する。本籍staffs.clinic_id一致を要件にしない。氏名・メール・Excel行順から同一人物を推測せず、未解決IDは行エラー、職員の自動登録なし。既存の認可済み職員選択表示で名前とIDを対で提示し、閲覧権限のないmaster情報をCSV補助のため開示しない。

notes_mode=setの上限は、通常勤務create/updateの既定入力制約と同じ2,000 Unicode code points（Go binderのstring maxと同じrune数）とする。byte数や見た目のgrapheme数へ置き換えず、trim/Unicode正規化/切捨てなしで共通勤務validatorが検査する。CSV preview/applyと通常CRUDが同じ検査へ接続し、CSVでbinderを通らないことを上限迂回にしない。超過は元行のvalueエラーでready=false、applyはvalidationで確認範囲全体ゼロwrite。keepは既存値の保持なので、旧長文を切り詰めたり未編集値の長さだけで新たに拒否したりしない。これは既定入力制約の入口整合であり、新しい業務上限の追加ではない。

### 2.2 合成例（実患者/実職員データなし、未実行）

```csv
record_type,shift_ref,staff_id,work_date,shift_type,starts_at,ends_at,notes_mode,notes
shift,s1,101,2026-11-01,full,2026-11-01T22:00:00+09:00,2026-11-02T07:00:00+09:00,keep,
break,s1,,,,2026-11-02T01:00:00+09:00,2026-11-02T01:30:00+09:00,,
break,s1,,,,2026-11-02T04:00:00+09:00,2026-11-02T04:15:00+09:00,,
shift,s2,101,2026-11-03,off,,,keep,
```

UI配布templateはUTF-8 BOM付き・CRLFのheaderのみを既定とし、この9列の説明/合成例を併置する。実staff値/自由文をCSV再exportする機能は本sliceへ追加しない。`=,+,-,@`等を含むnotesも式を実行せず原文文字列として扱い、画面はtext表示する。エラーCSV自動出力やアポストロフィ付与による保存原文の改変はしない。

## 3. 正規化・選択・preview

閉じた `Row` は `{shift_ref,source_record,staff_id,work_date,shift_type,starts_at,ends_at,notes,breaks}`、notesは`{mode,value}`（keepのvalue=null、setはstring）、breakは`{source_record,starts_at,ends_at}`。日時はserverでUTC RFC3339秒へ正規化、off/paid_leaveの日時は明示null。source_recordはheaderを1とするlogical record番号。quoted改行でも元physical開始/終了行をエラー表示へ対応付ける。staff/date/shift_refは文字列、source_record/件数は整数。breakはstart/end/source_record順、勤務はstaff ID数値順/work_date/shift_ref順へcanonical化する。

選択scopeはclosed `{staff_ids,from,through,selected_shift_refs}`。from/throughはJST勤務日範囲、staff_idsは昇順・重複なし、selected_shift_refsはfile内存在する非空集合、文字列昇順・重複なし。staff_idsは選択groupに実在するstaff集合と厳密一致させる。選択groupの全breakが常に同じ勤務単位に入り、breakだけ別保存しない。選択groupは全てstaff/date範囲内であることを検査する。全fileの構文/重複/identity書式を解析し、未選択groupは`excluded`として元ref一覧と件数を明示する。未知形式を無断で除外しない。未選択groupの医院所属等の詳細は読取認可なしに開示せず、選択する場合に認可/再previewする。利用者の明示選択以外の自動除外・正常行だけのapplyを禁止する。

existing cellは(clinic,staff_id,work_date)で全件読取し、presentは実shift ID/version＋日時/notes/全break、absentは明示不存在。既存と同じ保存値はnoop、absentはcreate、異なるpresentは`replacement_required`。置換意思はclosed `{shift_ref,shift_id}`の集合で表し、対象の現shift IDを指定して再previewする。確認前に差分/影響を表示し、自動upsertや無条件replace-allをしない。break集合の置換・offへの変更も当該shift全体の明示置換である。未掲載cell・他医院・未選択staffは不変。

`PreviewProjection` exact = `{schema:"shift-csv-preview/v1",source,scope,rows,impact}`。
source exact = `{format:"shift-csv-v1",encoding,sha256,byte_size,record_count}`。sha256はBOM/改行を含む元受信bytesのlowerhex64。
rows[] exact = `{input:Row,target:{presence,shift_id},before,after,resolution}`。presence=present|absent、shift_idはpresent時実ID/absent時null、beforeはabsent時null又は`ShiftValues`、afterは完全な`ShiftValues`。ShiftValues exact=`{staff_id,work_date,shift_type,starts_at,ends_at,notes,breaks:[{starts_at,ends_at}]}`、notesは非NULLstring、breakは時刻順。resolution=create|replace|noop、replaceは明示した実IDと一致必須。
impactは`{protected_bookings:[{staff_id,booking_ref,starts_at,ends_at,protection_state,before_eligible,after_eligible}]}`、refとprotection_stateは予約ownerが返す現在の実identity/保護状態だけ、staff/booking順で全件取得する。患者等の表示詳細は別の現在予約viewで認可し、digestには入れない。予約保護集合の定義は既決正本から拡張しない。存在しないgraph版を生成せず、版がなくてもこの確定値の完全比較で確認影響の変化を検知する。

`preview_digest=SHA256(UTF8(JCS(PreviewProjection)))` lowerhex64。全IDはcanonical十進文字列、null明示、未知key/重複JSONkey拒否、scalarのUnicode正規化なし。勤務の版・現在grant・CAS/fence・実行時刻は含めない。source bytes・選択集合・明示置換・採用値・変更前値・確認した予約影響は意味値としてcommitする。unknownな休業・失権・不足読取を空のimpact/absent/noopへ変換しない。解決不能/入力不正のpreviewにはdigest=null、ready=false、apply不可。

## 4. 新wire（選定、api.yaml/runtime未収録）

pathは既存`/api/v1` serverに対する相対path、Cookie/Bearerと既存JSON CSRF条件を維持する。専用permissionは追加しない。

| method/path | 要求/応答 | 現在認可 |
|---|---|---|
| GET `/shifts/csv-template` | 本稿9列のUTF-8 BOM/CRLF header、text/csv。勤務/個人値は含まない | shifts:view |
| POST `/shifts/csv-imports/preview` | multipartのfile1個＋options JSON1個。options exact `{format,encoding,scope,replacements}`。200は下記PreviewResponse | shifts:view＋選択staffの現読取。勤務write0 |
| POST `/shifts/csv-imports/apply` | 同file/options＋command JSON exact `{operation_key,preview_digest,preconditions}`。closed multipart、余剰part拒否 | shifts:view＋各createのshifts:create、各replaceのshifts:edit。noopだけもshifts:edit |
| GET `/shifts/csv-operations/{operation_key}` | intent=`staff.shift_csv.apply`固定のC06 ResultResponse。未知query拒否 | 元actor/scopeと保存anchorの現在shifts:view |
| GET `/shifts/csv-operations/{operation_key}/details` | 成功履歴の専用details、下記参照。共通receiptに任意detailsを追加しない | 上と同じ |
| GET `/shifts/csv-operations/{operation_key}/reconciliation` | 例外時の対象別証拠を読む専用projection（§4.1）。再実行/補償writeなし | 上と同じ |

formatはliteral shift-csv-v1。fileはraw CSVであってbase64 JSONや任意URLではない。preview/applyとも同一parser・normalizer・validatorを使う。applyにも元fileを再送してsource SHA/size・選択Rowの正規化一致をserverが検証する。browserが送るhashだけを信頼しない。

PreviewResponse exact = `{source,scope,rows,excluded_shift_refs,errors,impacts,cross_clinic_conflicts,preconditions,preview_digest,ready}`。rowsは`{shift_ref,source_records,existing,proposed,action}`（existingは`{presence,id,version,values}`でpresence=present|absent|unresolved。absentは後3値null、unresolvedも後3値nullだがaction=invalid/ready=falseに固定し、未作成とは扱わない。proposedはShiftValues又は解析不能時null、action=create|noop|replacement_required|replace|invalid）。errorsは`{record,physical_lines,shift_ref,field,code}`、record/ref/field不明はnull、physical_linesは整数配列。codeはencoding/header/csv/limit/identity/duplicate/value/scope/leave_unresolved/leave_conflict/cross_clinic_overlap/booking_protected/replacement_requiredの有限列挙。生のDBerror/原セル/権限外の個人値を含めない。impactsは予約ownerが現在認可付きで返す理由・日時・患者・区分・編集導線の既存projectionを利用し、予約viewなしでは本文/IDを開示せず件数と安全な拒否理由だけを返す。この場合詳細不明を解消済みとしてreadyにしない。preview_digestが非NULLなら対応PreviewProjection全体をserverが再構成可能であること。read失敗は空previewにせず503、入力/置換待ち200 ready=falseは勤務保存成功ではない。

`cross_clinic_conflicts`はrequired配列。各要素のexact keysは`{shift_ref,source_records,staff_id,disclosure,clinic_name,overlap}`。shift_ref/staff_idは選択済み入力groupと一致、source_recordsはその勤務行と休憩子行のlogical record番号の昇順・重複なし集合で、rows/source行へ戻れる。disclosure=availableの要素は、§6の既存認可信頼resolverが現在開示を許すclinic_name（非空string）と、実際の重複区間overlap exact `{starts_at,ends_at}`を必須とする。日時はUTC RFC3339秒、start<end、画面ではJSTの実日付/時間帯で表示し、日跨ぎも日付を省略しない。相手勤務の全時間ではなく、提案勤務との交差区間だけを返す。相手shift ID・患者・予約本文等を追加しない。shift_ref、重複開始、医院表示名の順で返し、同groupの複数医院/区間を黙って一件へ潰さない。

disclosure=restrictedではclinic_name/overlapをrequired nullとし、既存認可で開示できない医院名・時間帯を推測表示しない。入力groupに対する安全な衝突理由だけを表示し、errorsのcross_clinic_overlapと対応づける。restricted又はavailableの衝突が一件でもあればready=false、preview_digest=null。全医院について必要なproofを読み取れた結果がゼロの時だけ空配列を返し、proof自体の読取不足/障害は503で停止する。失権やread失敗を空配列/重複なしへ変換しない。このprojectionは既存resolverの開示範囲だけを使い、新grant・RLS bypass・他院勤務の汎用readerを追加しない。

Preconditions exact = `{cells:[{staff_id,work_date,presence,shift_id,expected_version}]}`、staff/date順。absentはID/version=null、presentは実ID/正version。選択全cellと厳密一致、不存在に版0/1を作らない。版不一致はversion-conflict、存在/identity/内容/影響変化はdependency-changed。その他grant/所属/能力/休業/予約/他院勤務は同TX現在再読必須で、存在しないeligibility stampをclientへ要求しない。変更前臨床値/予約影響が同じでversionだけ更新された場合のみ同Contentで新Preconditionsを再確認して同key retryできる。確認値/選択/置換/影響/digestが変わるなら旧結果確定後の新preview・新key。

Details success exact = `{status,operation_key,result_refs,recorded_at,source,scope,rows}`。この7keyを全て必須とするclosed objectで、committed_versions keyは定義せず常に出さない。status=succeeded|replayed、operation_key/result_refs/recorded_atは元receiptの保存値を返す。rows[]=`{shift_ref,source_records,action,shift_id,version,before,after}`、action=create|replace|noop。全て当該成功`shift_bulk_operations.confirmed_rows`から読み、後日のcurrentを混ぜない。共通success refs=`{shift_bulk_operation_id}`（binding ID）、共通ResultResponseのcommitted_versionsはoptional keyであり、このCSV consumerは内部空mapなのでomitemptyに従い通常lookupから省略する。専用detailsは共通DTOの直列化ではなく上記7keyへ明示projectionする。行別versionは固定detailsだけに置く。DBのimmutable saved rowsに`shift_ref→実ID/結果version`の対応を保持し、架空のbulk業務versionを作らない。成功refsにはnullable値・配列を入れず、共通EffectSummaryは内部`{changed:"true"|"false"}`だけ。全noopでもbulk履歴/receiptを1組保存、勤務版・勤務変更auditは増やさない。通常lookupは既存ResultResponseを維持する。共通schemaのresult_refs/committed_versions/recorded_atはoptionalだが、このCSV成功はstatus/operation_key/result_refs/recorded_atを返し、committed_versionsは省略する。非成功はstatus/operation_keyだけであり、専用detailsのsource/scope/rowsを通常lookupへ漏らさない。

C06 HTTPは succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。非成功はstatus/operation_keyのみ。details非成功も同形、保存rows欠落/不整合は現在認可後503で空行補完なし。pre-C06 malformed/format/encoding=400、サイズ=413、認証=401、認可=403、対象非開示=404、媒体不正=415、依存read障害=503は既存ErrorResponse。全応答Cache-Control:no-store。新しいC06 status/汎用wrapperは作らない。

### 4.1 例外時のreconciliation read projection

通常lookup/成功detailsのC06 shapeは変更しない。旧経路の部分commitや整合性違反を調べる専用GET `/shifts/csv-operations/{operation_key}/reconciliation` を選定し、AC-shifts-bulk-partial-resultの対象別表示へ接続する。元actorのC06 scope/keyと保存anchorを認可し、信頼できる保存commandの選択全groupを一つのread snapshotで列挙する。新規claim・metadata修復・receipt作成・履歴追記・reconcile状態更新を行わない。

200のclosed response exact keysは `{operation_key,integrity_state,rows,reexecution_allowed,compensation_allowed}`、全必須。後2値はliteral false。integrity_stateはverified_all_committed / verified_none_applied / partial_detected / undetermined。rowsは選択全groupのstaff/date/ref順、要素exact=`{shift_ref,source_records,staff_id,work_date,state,evidence}`。stateはcommitted / not_applied / unknown。source_recordsは元勤務/休憩record番号の集合。evidenceは証明不能ならrequired null、証明できる時はclosed `{kind,source,reference}`。kindはsuccess_receipt / owner_history / definitive_no_effect、sourceは既存operation_results / shift_bulk_operations / audit_logs / operation_attempt_outcomesの有限集合、referenceは当該既存行の正規ID又はclinic付きbinding参照を表す文字列。原値本文/DB内部情報は返さない。

committedは元bindingの権威ある成功receiptと全confirmed_rowsの整合、又は元操作・当該target・確定版を一意に結ぶ既存の不変owner履歴/監査の証拠がある場合だけ。後日の勤務変更で現値が異なっても元commit事実は保持する。not_appliedは当該対象へ効果がなかったことを示す既存の確定rollback/拒否証拠を照合し、同keyの後続成功・未確定/実行中attemptがないことも権威ある結果照会で確認できる場合だけ。単に過去のrejected attemptが一件あるだけでは不足とする。既存証拠に元操作/対象との関連がないlegacyでは補完せずunknownとする。行の現値一致/不一致、欠落、receipt不在、時刻経過、全件数一致だけからcommit又は未実行を推測しない。

全rowsがcommittedならverified_all_committed、全rowsがnot_appliedならverified_none_applied、committedとnot_appliedがともに証明できた場合だけpartial_detected、それ以外はundetermined。unknownを除外せず全対象を返す。証拠read不調で各対象を確定できない場合はunknown/evidence=nullで保留する。元anchor認可不成立は403非開示、保存commandから対象全集合を安全に確定できない又はquery自体を完了できない場合は503で停止し、空rowsの成功へ変換しない。無認可の証拠ref/元値は開示しない。

UIは対象ごとに「確定済み／未反映を確認／不明」を表示し、通常lookupの原結果と例外調査結果を分ける。partial_detected/undeterminedの間は全件再送・新key・補償削除を停止する。verified_none_appliedも再実行許可tokenではなく、既存§5の現在認可・明示再確認・同Content/同key retryだけに従う。別々の明示操作を無断合成せず、未反映範囲を選び直す場合も元unknownを解消してから新previewへ進む。新しい効果・永続表・C06 status・自動補償は追加しない。全応答no-store。

通常lookup・details・reconciliationの3読取とも、認証staff IDと元bindingのActorStaff/actor_refの厳密一致を必須とする。trusted clinic/owner_kind=staff/intent=staff.shift_csv.apply/keyでFindBindingし、保存元scopeと現在shifts:viewも照合する。医院viewだけで別actorのkeyを探索せず、別actorは元結果/対象/件数を非開示とする。これは既記載の本人scope/別actor lookup拒否の明文化であり、他staffへ権限を追加しない。

## 5. claim・Q・再送と原ファイル保管

owner_kind=staff、intent=`staff.shift_csv.apply`。業務operation_keyが唯一の再送identity。Content exact = `{schema:"shift-csv-command/v1",source_sha256,preview_sha256}`。Scopeはtrusted clinic/staff、Preconditionsは上記cellsだけ。confirmed PreviewProjectionを含むimmutable commandから全3keyをpure projectorで生成・検証する。Q constructor/Runnerは共有契約どおりdeep-copyとclaim前/Apply前再検証、Applyはreq.VerifiedCommandだけから型を復元する。

失敗bindingにも元対象認可を可能にする限定新metadata `shift_bulk_operation_targets` を選定する。columnsはclinic_id bigint、binding_id bigint、staff_ids jsonb array、from_date/through_date date、canonical_command_bytes bytea、command_sha256 bytea32、全NOT NULL、PK(clinic_id,binding_id)。唯一FKは(binding_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT。canonical commandはclosed `{schema:"shift-csv-command/v1",projection:PreviewProjection}`のUTF-8 JCS、最大4MiB、bytea32と再hash一致。staff_ids/日付はprojection.scopeと一致させる。通常UPDATE/DELETE拒否、clinic RLS、元actorをbinding scopeから解決するprivate repository限定。raw CSV・filename・credentialは保存しない。第二binding/key/receipt表ではない。実装では既存ClaimMetadataWriterのwinner-only同短claim TXでINSERTし、metadata失敗はbindingもrollback。claimに業務FK/read/lock/nested TXを持ち込まない。これらは追加DDL/model/ERDの引渡し定義で、今回DDLは編集・適用しない。

新key: strict parse→現在read/create/edit preflight→全preview値をserverで再読・再構成して受信digestと一致確認→Q/Content→短claim。入力不正・権限不成立・古いpreviewではbindingを作らない。新keyのclaim前preview不一致は409既存ErrorResponseで返し、C06 receiptが成立した形にはしない。claim race loserは勝者metadataから元scopeを再認可し、incoming staff集合でanchorを補完しない。

既存key: 現在元anchor認可→保存bytesのdecode/再canonical化/hash/全Content一致→incoming raw SHA・正規化した選択Row・scope/置換・preview digestの意味値と比較する。現在値から別commandを採り直さない。異Contentはkey-conflict。成功は現在cell版/休業/予約変化に先立ち元receiptを返し、未成功だけ現在write grant・全Preconditions・依存再検査後に同commandでApply。認可anchor欠落/不整合/解決不能はnot-authorized/403非開示。元target認可が成立した後のcommand破損/結果read不調だけresult-unavailable/503。

元CSVはpreview/applyのbounded request memoryと利用者が選択した元fileだけに保持し、サーバの永続blob・一時disk・job queue・監査への元file複製は採用しない。保管するのは元bytesのdigest/長さ/encoding/profileと正規化した確認行・前後値、actor/時刻/operation対応である。業務履歴/失敗claimのprivate commandはC05/既存操作履歴の保持対象で、独自TTLや自動purgeを追加しない。元ファイルの将来長期保管/exportが必要なら別のアクセス/保持設計とし、今回の着手を止めない。

FEは未送信file/案をメモリに保持し、医院切替・失権時に本文を破棄、遅延応答を医院/選択/request世代で拒否する。localStorage/URL等へraw・command・認可tokenを保存しない。reload後の未送信本文自動復元やkey discoveryは本sliceの保証外。keyが分かる場合は先にGET lookupし、成功はdetailsを読む。unknown/busy/結果読取不能は保留し新key発行不可。確定未成功のretryでfileを失った場合は、利用者が元SHA/長さと一致するfileを明示再選択した時だけ同key/同Content再送。元key又は送信有無を特定不能なら停止し、別keyで自動救済しない。file再選択・全件一致だけを未実行の証拠にしない。

## 6. applyの原子性と全writer閉包

通常勤務・休憩・CSVで一つの保存関数とaggregate versionを使う。入力口ごとの独自upsertや旧単日APIのloop commitは不可。現在actor/account/assignmentの非locking preflight後、実装順は必要な既存graph gate（ContextGate）→C06 binding NOWAIT→成功receipt/unknownの再読取→savepoint→未成功Apply内で選択全staffのglobal row FOR UPDATE NOWAIT（ID昇順）→当該clinic assignment（staff順）→勤務cell/全break（staff/date/ID順）である。業務rowをParticipant.Gateで先取りせず、共通Runnerのbinding→Applyを維持する。493のstaff→assignment競合点を保持し、独自advisory keyや予約row後取りを追加しない。共有gateが派生ctxを返す場合は採択済みContextGateの必要subsetで後続へ渡す。

binding下で成功replay/unknown停止を先に処理し、未成功Applyではgrant/有効所属・全cell CAS/不存在・完全before/現在影響digestを照合。対象staffの全医院の重複と、selected after同士の重複も検査する。既存のexclusionを無効化せず、cross-clinic proofは信頼済み既存owner resolverから医院/重複時間の必要projectionだけ取得し、§4のcross_clinic_conflictsへ入力group/元行と対応づけて返す。RLS bypass/新roleやclient列の医院追加はしない。取得不能を重複なしと扱わない。勤務日だけの検索で前日開始の日跨ぎや期間末の翌日終了を落とさず、before/afterの実区間で照合する。

休業は493のJST inclusive/exclusive/unknown/nullを共通関数で評価。新しい勤務可能区間に重なる休業を不適格とし、unknownで該当可否が決まらない行は未解決で停止する。off/paid_leaveは勤務可能化しない。既存の同値noopや勤務能力を増やさない訂正に、無関係な退職/過去休業を理由とする新しい一律拒否を追加しない。現在の業務保護条件に関係するstaff/assignmentは保持し、retired属性を勝手に消さない。

予約は未来・現在診療中（予定終了超過含む）の既指名保護を通常編集と同じ関数で評価。CSVでは予約rowを更新/後取りlockせず、staff gate下で必要保護集合を再読する。保護集合へ入る予約create/日時・staff変更/再受付等は対象staff SHAREを変更前からcommitまで保持し、勤務/休憩/他院勤務/休業/所属・能力writerは同global staff競合点へ参加させる。複数staffは同じ昇順。予約が先ならCSVは新状態を再読、CSVが先なら予約が新勤務で再検査する。staff gateを通らないwriterを残してCSVをactivateしない。関係ない全467/493完成は待たず、この一覧に該当する必要経路だけを閉じる。

全置換の最終集合を検査してからwriteする。相互に時間を入れ替える同staff複数cellで一時的なexclusion違反を作らないため、影響する既存勤務のstarts_at/ends_atを同TX内で一時NULLへ退避し、その後に最終値を一度設定する方式は採用しない（二重version/履歴を起こす）。実装subsetとして既存exclusionを同じ意味のDEFERRABLE INITIALLY DEFERRED制約へ変更し、最終集合の同TX検査を必須とする。勤務値変更は各row一回、break変更だけでも親aggregate版を一度進める限定writer/trigger整合をschema ownerへ渡す。通常日時CRUD/休憩writerも同じ版責務へ参加し、serviceとtriggerで二重進行しない。現version triggerが親field差分だけを見る不足は実装残差で、単に「CASあり」と扱わない。constraintの無効化・旧時間列復活・手動版加算の二重化はしない。

保存順は全勤務/全break→既存shift_bulk_operations(kind=csv)へsource_snapshotと確認全rowsのbefore/after/実ID/結果版/元record対応→必須audit→C06 receipt。同outer TX、途中の勤務を成功公開しない。全noopはbulk履歴/receiptだけを追加し、勤務版/変更auditを増やさない。業務FK/件数・metadata/projection整合・必須監査・成功receipt/deferred constraintの失敗は全rollback。応答断は同key照会、確定拒否とcommit不明を区別。部分commitを正常応答にせず、旧経路の不整合が判明した場合は§4.1の証拠付きread projectionで対象別に確定/未反映/不明を表示して停止、自動補償削除・全件再送をしない。

## 7. auditと履歴reader

変更がある成功は既存audit_logsへaction=`staff.shift_csv.apply`、resource=`shift_bulk_operation`、resource_id=binding ID、actor_type=staff、actor_id=trusted staff、clinic_id=scope。old_value/new_valueは`{rows:[{staff_id,work_date,shift_id,version,values}]}`、oldのcreate行はvalues/ID/version=null、newは保存後。metadata exact=`{operation_id,format,source_sha256,source_byte_size,encoding,preview_digest,selected_shift_refs,created_count,replaced_count,unchanged_count}`。元bytes・filename・credential・予約患者本文は含めない。source_snapshot exact=`{schema:"shift-csv-source/v1",source,scope,impact}`、confirmed_rowsはdetails.rowsと同じ完全集合。内容/件数とreceipt refsの同binding一致を保存時/読取時に検証する。既存bulk immutable/RLS/成功receipt必須を再利用する。

同key replayはaudit/履歴を増やさない。通常業務履歴の保持を「元CSVを保存しない」と混同しない。detailsは現在shifts:viewと元選択医院/staff scopeで認可し、他actor/key探査・生のprivate claim command開示をしない。履歴があることを現在勤務や予約の再実行許可にしない。

## 8. 1PRで着手する責務と受入条件

| owner | CSV実装PRへ含める必要subset | 混ぜない範囲 |
|---|---|---|
| 448統合owner＋493 staff | parser/profile、typed input、preview/apply/lookup/details、同TX勤務/休憩/bulk履歴/監査、元anchor authorizer、同一validator | コピー全面、staff CRUD全体、休業UI全体 |
| 489共有owner | 採択Q・winner-only claim metadata・派生ctxが必要な時のContextGate。実在hookは再利用 | 別C06/別キー台帳、全consumer完成待ち |
| 446/466 schema/model owner | metadata1表、既存exclusionのdeferred化、休憩だけのaggregate版整合、RLS/immutable/必要FK/成功receipt整合と正規model/ERD生成 | 今回のDB適用・旧列復活 |
| 467予約owner | staff SHARE参加・現在保護/影響projectionの必要経路、他院重複resolver | 自動予約変更、全予約機能の完成待ち |
| CSV FE担当（同じ統合PR） | 9列説明/template、encoding選択、全件/対象外/エラー/衝突/明示置換、案保持、scope世代guard、同key lookup/不明停止 | xlsx/給与/外部upload、新規ブラウザ永続draft |

API正本/schema/operationIdは共有ownerが本closed契約から後続の同CSV実装PRへ収録する。今回api.yamlは編集しない。FEのUI/state準備は先行可能だが、公開APIに対するmock完了判定は固定したapi.yamlとの整合をもって行う。CSV用2MiBのbounded uploadはraw11MiB画像設定等を流用せず、body parser/multipart上限と内部caller上限を揃える。

実装PRは上の必要subsetを単一統合ownerが採用SHAで回収し、契約・parser・preview・原子保存・結果照会まで閉じる。全勤務/全予約/臨床値移行/旧報告/実DB試験完了を一律の開始条件にしない。現行コードの不足はPR内必要subset又は担当へ渡す明示依存とし、未接続入口を有効化しない。試験実行は今回NOT_RUNのまま、以下を実装担当のfixture/期待値として渡す。

1. UTF-8 BOM有無、Windows-31J明示/誤選択/不正列、CRLF/LF、quoted改行/quote、未知/重複header、壊れたrecord、file/行/group/JSON/command上限超過でゼロwrite。
2. notes setの1,999/2,000/2,001文字（多byte文字/改行含む）、byte上限との独立、keep旧長文不変、通常CRUDと同じvalidator/エラー境界を確認する。同名職員・他院staff ID・本籍別医院の有効assignment、指数/丸め済ID、未知ref、同cell重複、複数break・前後行参照・引用空と未掲載・keep/set空の区別。
3. cross_clinic_conflictsの元shift_ref/行→他院名/実交差時間（JST日跨ぎ表示）の対応、複数衝突全件、restrictedの必須null/ready=false、proof不足503を確認する。日跨ぎ/年末/閏日、work_date不一致、off/paid_leave＋時刻/break拒否、break端点と重複、同staff跨院重複・最終集合の入替、移動時間の手動間隔を勝手に1時間必須にしない。
4. preview/cancelで勤務/claim0、未選択/未掲載不変、明示置換なしready=false、1件不正で確認全範囲未反映、範囲選び直し再preview。通常CRUD/514の同日時入力と保存値一致。
5. 493休業のinclusive/exclusive/空区間/null/unknown、退職や所属変更、未終了指名予約・予定終了超過・再受付、過去完了だけ/影響なしの過剰拒否なし、患者詳細失権非開示。
6. preview後の新勤務/新予約/休業/所属/能力/他院勤務変更、absent→present、breakだけのversion増分、相互cell置換、gate busy/NOWAIT、監査/receipt/全break/deferred失敗で全rollback。
7. 同key同内容replay、同key別file/選択/置換/digest conflict、CASだけの同key再確認、unknown/busy新key禁止、応答断lookup、再選択fileのSHA不一致、key喪失時停止。成功detailsは後日の勤務改訂でも原before/after/版を保持。
8. reconciliationの全対象列挙と証拠有無によるcommitted/not_applied/unknown、現値一致/receipt欠落だけでの推測禁止、partial/unknown時の再実行・補償ゼロ、後続成功/未確定attemptを無視したnot_applied誤表示拒否、3読取の元actor ID一致を確認する。details成功の必須7keyとcommitted_versions不在、共通lookupのoptional key/非成功shapeを照合する。raw/filenameをDB・disk・logへ保存しない、private metadataと成功履歴のRLS/現在認可、別actor lookup拒否、失権/医院切替の遅延応答拒否、formula非実行、共通receiptへの配列/details漏出なし。

業務追加質問は現時点なし。実CSV見本との対応確認・性能上限調整・実装/生成/DB/実機は担当の受入残であり、本profileを「任意Excel CSV対応」「実装完了」としない。profile外の個別形式だけを対象外として説明し、今回の契約/着手全体を資料待ち・試験待ちへ戻さない。

## 固定根拠

- [最終ShiftEntry/Break型](https://github.com/MinoruSoga/AnimalEkarte/blob/9a1d50f9dcc7d275456f1214507ef0a55441ac98/backend/internal/model/staff.go)
- [現勤務requestの移行残差](https://github.com/MinoruSoga/AnimalEkarte/blob/9a1d50f9dcc7d275456f1214507ef0a55441ac98/backend/internal/staff/shift_request.go)
- [staff/assignmentとcell lock](https://github.com/MinoruSoga/AnimalEkarte/blob/9a1d50f9dcc7d275456f1214507ef0a55441ac98/backend/internal/staff/shift_entry_repository.go)
- [既存DDL・bulk/休業/RLS/版](https://github.com/MinoruSoga/AnimalEkarte/blob/9a1d50f9dcc7d275456f1214507ef0a55441ac98/backend/migrations/001_init.sql)
- [共有operation契約](https://github.com/MinoruSoga/AnimalEkarte/blob/9a1d50f9dcc7d275456f1214507ef0a55441ac98/backend/internal/operation/contract.go)

## 2026-10-21 コピー・段階台帳への接続

コピー固有の元snapshot/日付写像/closed wireは[コピー選定契約](emr-448-493-shift-copy-contract-20261021.md)へ追加選定し、本CSVのparser/profileを呼ばず勤務aggregate・予約保護・claim metadata・結果履歴の必要subsetを再利用する。CSVの選定済み範囲を未決へ戻さない。①契約→②master/勤務入力→③共通UI→④新業務consumerと⑤後段試験の完了境界は[集約台帳](emr-445-master-stages-ledger-20261021.md)に従う。④全予約完成を②へ逆依存させず、既存予約を守る必要reader/gateが未接続なら該当CSV/copy保存だけを未活性にする。試験実行NOT_RUN。
