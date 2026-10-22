# A〜F引き渡しと完了監査表（2026-10-22）

計画・確認項目の整備であり、稼働・実装完了・受入PASSの記録ではない。実際の状態は各PRとチケットの最新証拠で更新する。進行中コードの再監査・実装完了認定ではなく、今後の完了報告を検証する受入表である。今回docsのみ、製品コード・API YAML・DDL/model・設定・生成物・Planeは変更しない。試験/build/codegen/DB/Docker/browser実機確認はNOT_RUN。マージ・deployは行わない。

原典は採択済み[A01/A06契約](emr-master-update-selected-contract-20261022.md)、[A02/A03契約](emr-clinic-line-time-selected-contract-20261022.md)、[①着手計画](emr-445-stage1-contract-start-plan-20261021.md)、[②着手計画](emr-445-stage2-master-start-plan-20261021.md)と各票の最新コメント。旧「現5セッションへ追加しない」は過去scopeの履歴であり、今回の明示6枠割当と区別する。実装開始時の最新正本SHAを各枠が記録する。532 Q1〜Q3の新しい具体回答は本人承認待ちで、本書は未承認案を採用しない。

## 1. 6枠の所有と受領

| 枠 | 所有 | 必要subset・受領先 | 採用確認条件 |
|---|---|---|---|
| A 533/536 | api.yaml単一窓口、入院PATCH nullable、薬剤clear/税率exact、必要shared operation契約調整 | B保存型とのAPI整合、C/Dの要求する限定APIとregistry/lookup等を既存489責務に従い調整。domain writerを奪わない | PATCH-only契約、GET/POST非拡張、exact往復、必要共有hunkの固定SHA・受領先を明記。536を無断分割しない |
| B 446/466/527 | DDL/model/ERD/登録・正規生成入力と本人生成の受領 | A01 7bool/3version→C。A03 common aggregate/LINE限定版→D。A02 policy/492 anchor staff・assignment拡張→D。限定波で供給 | 既存001/006/806/807を巻戻さず採番は最新mainで調整。nullable/default/backfill、RLS/FK/guard整合、生成入力SHA・出力受領・NOT_RUNを区別 |
| C 447/449/456 | 品目専有request/service/repository/response、presence/CAS/同TX監査、旧writer保持、健診親lock | BのA01必要型とAのwireを受領。Aの536 FEと共有API/modelには触れない | 7resourceごとのPOST/PATCH/GET/NULL/false/unknown/競合/監査rollback/旧writer表。必要型未受領は未完と記録 |
| D 448/492 | 医院/LINE設定のdomain backendとoperation adapter | B common→LINE版→policy/anchor、A APIと489 registry/lookup/Q/ContextGate必要subset、493現在所属reader | immutable verified command、原assignment現在認可、旧writer閉包、同TX audit/write/receipt、元key lookup。旧repo独立TXを外側で包むだけは不可 |
| E 532 | package capなし全件照合、clinic test double、403/409 FEとfixture | 既存521/523 APIを再利用。Q1〜Q3の具体回答は本人確認待ち。既存ACは維持し、未承認案を採択済みにしない。型/interface競合は該当hunkを調整 | S06/S07/S08/S10各ACとQ1〜3の回答状況を明記。後続534へ最新wrapperを渡す。522/524は対象外 |
| F 537 | 物販不足import・追加不可表示位置・campaign/物販exact往復fixture | 既存813/818を基底にする。Cの物販production writerとfixture範囲をhunk単位調整 | 3群全ACを確認。P1 importだけで全完了としない。API/DDL/生成型を横編集しない |

C/Dは既選定契約に沿う専有実装を並列開始できる。A/B全票Doneは待たないが、必要共有型の独自stub・二重型・生成物手編集で先回りしない。小さいsubsetの引渡し後に結線し、必要subset未採用時は統合完了としない。A/Bは単一ownerの限定波であり巨大1PRを要求しない。

共通採用確認：PR作成≠main採用≠runtime公開≠受入PASS。各subsetのbase/head/採用SHA、供給file/symbol/wire、consumer受領、未完/NOT_RUNを対応づける。PR作成までの指示からmerge許可を追加しない。生成はユーザーのみ。実DB適用・実値照合前にruntime有効化しない。

## 2. 完了監査の判定規則

- A/Bの必要subsetをC/Dが受領する。E/Fは既存基盤の限定修正で、新DDLを待つ一律gateは置かない。
- 「関数/列が存在」「fixtureを書いた」「PRをmergeした」「実行してPASS」を別欄にする。未提出は未確認、未実装の契約は未実装のまま。NOT_RUNをPASSへ変換しない。
- 証拠欄は各行について **base SHA / final head / 採用merge SHA / AC→実ファイル・関数・行 / fixture名と期待値 / 実行結果又はNOT_RUN / 受領元SHA・hunk / 残件owner** を埋める。空欄がある行は全AC完了に含めない。
- 保存traceはhandler・request decode→service/現在認可/validation→repository/同TX/model→response/再GET→既存callerの順。schema追加だけで接続済みとしない。
- 正規生成はユーザーのみ。生成input SHA・必要output・ユーザー実行/受領の有無を記録。生成物手編集や未生成の隠蔽なし。

## 3. A：共有API / EMR-533・536

| 確認AC | コード経路と想定fixture | 依存受領/PR完了条件 |
|---|---|---|
| 533 入院PATCH備考nullable | api.yamlの実PATCH requestBody→schemaを解決。description省略/null/空/通常文字列とnumber/bool/object/array。BE既存*string nil保持と一致、POST/GET非拡張 | API単一owner。533と536は同一セッションで関連差分を1PRにまとめてよい。定義があるだけでなくroute参照に接続 |
| 536 薬剤tax exact再編集 | LegacyMedicineResponse→transform→form→update builder→API。0.1234567890123456789、0.0000001の名称/備考のみ更新で保存値不変、0/0.08/0.1明示変更、失敗draft保持 | 元文字列保持/未編集省略。新BE桁制限なし。536のFEは既存承認された例外subset、他の後続FEを吸収しない |
| 536 clear_strength契約 | PATCH-only property→既存bool decode→実効設定検査→NULL map→再GET。true/false/省略/null、strength併送、none/per_weight、不正型を分ける | FEは同時送信しない。直接APIは既存clear優先を記録し一律排他400を創作しない。POST/GET非拡張 |
| C/D向け新契約の引渡し | A01 7resource、A02/A03各route・closed DTO・status/refs/versionsの必要schema・参照一覧（A06は対象外） | A枠の実claimに含まれる受領範囲を明示。533/536だけの完了を全共有API完成へ拡張しない。C/Dは未収録subsetを未受領として残す |

## 4. B：保存基盤 / EMR-446・466・527

- [ ] A01：7 nullable bool列にDB default/backfillなし。新3表exam_types/medicines/merchandise_itemsだけversion guard、既存4表のCASを維持。
- [ ] A02/A03：clinic_operating_hours、LINE限定version/reviewed版、staff policy複合identity・tombstone、492 operation anchor追加列・制約/RLS/immutableを正本どおり。実値自動移行・全staff slot複製なし。
- [ ] 001/必要migration・model・ERDの責務を単一ownerで維持。各consumerへ列名/型/null/default/PK/FK/guard・旧writer影響・固定SHAを渡す。C/Dが独自DDLや第二版を作らない。
- [ ] fixture：rawNULL/false/true、no-op版/実変更版、同TX rollback、scope/constraint、旧writer保持、旧行未照合。DB適用・実データ棚卸し未実施は別欄。
- [ ] 527正規生成の入力SHA・ユーザー実行結果・受領物差分を記録。必要生成未受領なら統合検証完了を宣言しない。
- [ ] API/registry/Q/ContextGate/lookup等の共有操作subsetは実ownerと受領SHAを別に明記。DB/model存在を共有runtime完成とみなさない。

## 5. C：品目BE / EMR-447・449・456

対象：A01 7resourceのBE writer/read/audit、456健診親子競合の限定差分。FE writerと④会計consumerは今回BE完了と分ける。

| AC | 必須コードtrace | 想定fixture/期待値 | 受領・完了ゲート |
|---|---|---|---|
| C1 POST/PATCH presence | 各resource request/handler→presence-aware bool→service既定/更新map→repo→response | POST省略false、true/false保存、null/非bool400。PATCH省略でNULL/false/true保持、false明示保存、null不正で同梱field含めwrite0 | Aのwire、Bの7列受領。保険helperのcreate=true既定を転用しない |
| C2 raw読取とunknown | 7種reader projection→response必須key→保存後再GET/既存caller | 正常読取NULL=rawNULL、false/true明示保持。列欠落/読取失敗/元不明はunknown/失敗、成功null/falseに補修しない | SELECT *とnilだけで成功読取証拠としない。全7resource別々に経路・fixture記録 |
| C3 CAS/実保存版 | existing4/new3のservice→WHERE期待版→guard→再読取response | 新bool PATCHは版必須、旧互換PATCHで任意version指定時も検査。古版zero-write、no-op不増分、is_active補償UPDATE後は実保存版 | B guardとA版契約を受領。アプリ二重増分・第二counterなし |
| C4 監査/旧writer | 最外TX→before/after rawbool→既存sink→write/result、Save/upsert/import入口 | 保険同時変更も同TX、audit失敗全rollback。create前不存在と既存NULL区別。非所有writerが既存true/NULLをfalse化しない | 7種writer一覧が必須。過去会計/治療/帳票snapshot無変更。C06全体対応を偽称しない |
| C5 grant/clinic | 既存route action→trusted clinic/actor→service/repo再検査 | 別院/無権限zero-write、checkupsとmaster-medicalとmaster-merchandiseを混同しない | 既存grant採用。body actor/clinic不採用 |
| C6 456親子race | checkup Create/親変更/Delete→同一親row lock→同TX存在/子参照再検査→保存 | 子保存先行/親削除先行の2順序で片方だけ成立、CAS/監査維持、soft-delete後の参照禁止 | 通常SELECTのみをrace解消としない。実DB並行fixture未実行はNOT_RUN。新DDLを無条件要求しない |

C提出は7resource×C1〜C5の対応表とC6を要求。1resourceの成功を残6へ一般化しない。A/B未受領部分は「実装準備済み/依存待ち」、未結線は「未実装」と区別。

## 6. D：時間系BE / EMR-448・492

対象：A02/A03の医院/LINE時間に関する設定保存/read/preview/旧writer閉包のみ。A06 trimming解除は今回DにもAの共有契約波にも含めず、別後続。新設定FE・④予約全体完成は除外。

| AC | 必須コードtrace | 想定fixture/期待値 | 受領・完了ゲート |
|---|---|---|---|
| D1 common aggregate | clinic route/closed request→時間validation→現在grant/CAS→repo→state response | 未設定/正常空/破損/失敗を別判定、7曜日・複数区間・日跨ぎ明示2区間・24:00 JST変換。initialize競合/更新no-op | A DTO、B schema/model/guard。実医院値自動コピーなし |
| D2 LINE acceptance | reservation acceptance request→P12field/期待common版検査→同TX write/reviewed版→response | legacy未照合→confirm、P変更版、非P不増分、common縮小mask/拡大時保存LINE保持。判定不能fail-closed、秘密非投影 | 旧全量PUTは活性化医院で同値でも409/write0。presentation/registration-fields先行救済、未切替医院保持 |
| D3 staff policy | type/staff/assignment検査→policy decoder→CAS/tombstone→管理GET/lookup | 週次・日別置換・明示空・例外削除。createでtombstone迂回禁止、reconfigure同identity、別assignment/別院拒否 | B複合identity/anchor、A exact refs/versions。旧slot全staff複製なし、架空policy_idなし |
| D4 C06原子性/lookup | Q deep-copy/commitment→claim原anchor→VerifiedCommand→graph gate/binding/rowlock→audit+成功receipt同TX | replay元結果、unknown元key照会、異intent/target非開示、audit/receipt失敗全rollback、no-op/replay不増分 | 共有Q/ContextGate/registry/lookup必要subset受領。独立TXをネストしただけでは不可。12status/flat response維持 |
| D5 警告/旧休日writer | preview影響集合→warning proof→保存時同関数再評価、shifts/closing限定adapter | 影響予約変更409再確認、view不足非開示/確認不能、取消claim0/write0。新commonと旧レジ休日独立、dual-writeなし | 既存warning adapter受領。予約/勤務の連鎖変更なし、LINE overrideなし |
| D6 resolver引渡し | ResolveClinicAcceptanceWindow/ResolveRegisteredStarts相当のreader→state/scope/版/集合 | 通常終了=閉店、ホテル18:00可/18:01不可、休憩開始点は再許可しない、24:00実JST日、途中休診で全滞在拒否を新設しない | 設定previewを予約可能tokenと呼ばない。④468のmatching/勤務/能力/占有への接続は別受領欄 |

D提出は各新route・旧writer/非P救済・lookup familyを列挙。新routeだけ実装して旧全量PUT/直接UPDATEを残す場合は完了不可。必要schema適用/model/reader確認前のruntime有効化、BE前のFE公開をしない。

## 7. E：EMR-532 package限定4系統

| AC | コード経路 | fixture/PR完了条件 |
|---|---|---|
| E1 reorder母集合 | treatment_package_service→catalog lock→repo真全件照合→CAS/write | inactive込み10001件、欠落/重複/別院/CAS不一致zero-write、no-op版不増分。表示用cap一括撤廃では不可、部分一覧を全件として送らない |
| E2 clinic test double | clinic ports→atomicClinicStore→既存同TX/rollbackfixture | catalog version create/delete2method追随。production interface縮小なし。静的型対応とbuild未実行を分離 |
| E3 typed403 panel | TreatmentPackageSettings/list error→既存target/editor/mutation境界 | master-medical view保持＋型別view喪失で旧editor表示/送信停止。一時通信失敗と失権を混同せず、親grant喪失は既存guard維持 |
| E4 409比較 | editor draft→409→最新GET→比較→明示選択 | draft保持、取得失敗/医院切替/遅延応答の誤適用なし、blind再送なし。alert文言だけでは完了不可 |

Eは既存521/523 API/CAS/clinic interface利用、新DDL/APIを独自追加しない。Q1〜Q3の具体回答は本人承認待ちで、上表の既存ACと回答状況を別に記録する。534のpackage FAB/scroll hunkはEのFE変更を先に受領してから。Eは534レイアウト・522/524・494/Fable・既知P01を吸収しない。

## 8. F：EMR-537 物販/キャンペーン限定後続

| AC | コード経路 | fixture/PR完了条件 |
|---|---|---|
| F1 missing import | MerchandiseSidePanel→既存isMerchandiseTaxRate export→selector | 新規/編集、0.10/0.08/0/0.08333/0.10000000000000001で未定義参照なし、rawstring/保存payload保持。任意税率UI新設なし |
| F2 エラー常時可視 | ItemListCard選択→表現不能判定→onAddItem遮断→scroll外error | 長いlistで理由が残る、書込なし、正常商品選択で解除。0/0.10への補修なし |
| F3 exact保存往復 | campaign/merchandise request→service→repo→GET response→専有FE form/builder | campaign rate/amount/0/10.50/100、物販0/0.10/0.08333/多桁。POST既定、PATCH省略/null保持/明示0、invalid400、DB読戻し値と応答exact一致 |

F3はparser単体だけでは保存再GETfixtureを代替しない。fixture存在・mock・実DB/harness実行結果を分ける。金額/税の新規則、既知異常保存値500の再設計、会計calculator/snapshotは除外。

## 9. 共有hunk・受領順（全枠の提出欄）

1. A↔B↔C/D：API path/DTO/列型/版/registryを必要subset単位で一致させる。全親票Done待ちではなく、具体SHA・interface・未受領項目を記録。
2. C↔F：merchandise_item_{request,service,repository,response}.goおよび対応testはCのA01 bool/CAS/auditとFのexact往復fixtureが接近する。production A01 hunkはC、exact回帰fixtureはFを基本とし、同test hunkは先行ownerを合意してpatch受領。税率exactと新boolを互いに巻戻さない。新version必須がfixture条件へ影響する場合はC契約受領後にFを合わせる。
3. A536↔C medicine：Aは薬剤税率FE/API clear、CはA01薬剤bool/version BE。api.yamlはA単一owner、medicine request/response必要差分をCから渡す。新field存在だけでA01 FE writer完成にしない。
4. E↔534：Eの403/409 editor構造を先に固定、534のwrapper/clearanceを後受領。E受領前に534同hunkを編集しない。
5. D↔B/492/489：anchor/型/lookup許可/C06共有部分を勝手に別実装しない。Dの保存/receiptとBのDB適用を別完了欄にする。
6. A06 trimming旧2属性解除は今回A/Dの対象外で別後続。531/534/535、A01等の後続FE、④consumerは今回6枠の未割当部分を自動で完了扱いしない。独立票の完了記録を要求。

## 10. 最終報告テンプレート（各枠）

- 対象AC総数 / 静的接続確認 / 未実装 / 依存待ち / 対象外（理由）：
- base SHA / final head / 採用merge SHA / 受領元SHAとhunk：
- ACごとのhandler→service→repo/model→response→caller固定リンク：
- fixture名・入力・期待保存値/response/副作用、unit/mock/integrationの区別：
- tests/build/codegen/DB/browser結果：NOT_RUN又は実際のcommand・実行環境・結果。未実行を成功欄へ入れない：
- 既知残・新規漏れ・解消済み先行reviewを区別し、修正票/owner/次の必要subset：
- 独立レビューが見たheadと未解決指摘。merge済みという理由だけで完了判定しない：

現時点の全行状態：**受入表準備済み／実装・fixture・依存受領の新たな確認は未実施**。進行中6枠の進捗を本書から推定しない。
