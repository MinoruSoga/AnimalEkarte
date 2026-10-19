> 追跡: EMR-446/466、467–479。S1 DBはPR728 head af2bc69c1で限定実装・DB16件検証済み。以下のAPI/UI・共通gate・全writer・S2の設計と、実装・受入完了は別。原設計後の独立reviewでstatus NULLも拒否するDB guardへ修正済み。ローカル検証成功を全26受入へ転用しない。

# EMR-460 予約ペット別再診回答：DB・consumer契約

**原設計の履歴（2026-10-19）**：以下は設計調査時点を起点とする。現在はS1 DBをPR728へ実装・限定検証済みで、統合候補は[統合DB契約](unified-db-20261019.md)に従う。API/UI・共通gate・全writer・S2接続の完了は示さない。
業務正本は予約仕様 `FR-reservations-line-revisit-intake` と6AC。既存詳細設計の対象世代・原文別保存・共通gate・metadata-only claim・結果照会を具体化する。承認済み001編集方針の次の限定差分として提案し、実DB適用許可へ拡張しない。

## 1. 結論と用語

**対象世代付きの予約ペット、追記専用回答、操作receiptを3表に分離する。最新はその対象の最大revisionから導出する。** `appointments.revisit_intake_generation`を追加し、実際のpet付替えでDB triggerが必ず増分する。`medical_records.revisit_intake_target_id`は後続の治療プランreaderが来院当時の対象を取り違えないためのnullable参照とする。

| 用語 | 意味・責務 |
|---|---|
| 予約対象target | clinic＋appointment＋pet＋対象世代。予約ownerが管理する「この予約のこのペット」。予約日時や現在の飼主名はidentityではない |
| 回答answer | 一回の送信で受理した7項目、認証回答者、server時刻。訂正は新しいrevision。スタッフInquiry・プロフィール・治療記録ではない |
| receipt | 同じ認証scopeと操作keyで一度成功した事実。現在の予約状態、最新回答、現在閲覧できる本文と分ける |
| record binding | カルテから、対応する来院のtargetを指す固定参照。同日同ペットや現在appointment世代の推定を使わない |

既存 `inquiries.medical_record_id` は必須・一意で履歴もない。ここへ原文を格納するとカルテなし保存・原文保護・コピー除外を満たせない。`appointments.customer_fields` / `line_customers.additional_fields`の汎用JSONへの混在も、target世代・回答版・訂正者を保護できない。専用保存が必要。

## 2. 方式比較と選定理由

| 選択 | 採る方式 | 比較・理由 |
|---|---|---|
| 対象変更検出 | appointmentの独立BIGINT世代、pet変更trigger | pet IDだけではA→B→Aを識別不能。updated_atは日時変更・notesでも変わり、継承仕様と合わない。新しい汎用予約versionを同時導入しない |
| 履歴 | target＋連番revisionのappend-only行 | 最新だけUPDATEする方式は旧原文を失う。JSON配列全置換は競合と内容編集の面が広い |
| 最新参照 | 同じtargetのMAX(revision)、唯一indexで逆順取得 | mutable latest pointerは複合FK・版対応と追加writeを増やす。appointment lock下でMAX+1と期待版を検査すれば別pointer不要。本文・版・受付状態は同じSQL snapshotで読む |
| フォーム | v1固定の型付き8保存値（7項目中元気/食欲は別値） | v1だけに汎用フォーム定義DB・JSON schemaエンジンを新設しない。フォーム版とclosed codeで旧ラベルの意味を固定する |
| 原文監査 | immutable answer＋成功receiptを専用監査sinkにする提案 | 既存audit_logsはstaff/systemだけ。LINE人間をstaffやsystemに偽装せず、全システムのactor enum拡張もこのsubsetへ混ぜない。専用sink例外を仕様に明記して採用する。スタッフによる予約変更の既存必要監査は別途維持 |
| record対応 | nullable target FKを保存 | recordのpet/appointmentだけでは同じpetへの再付替え後の世代を証明できない。明示bindingなしの旧recordを現在targetへ自動結合しない |

## 3. 限定DDL・ERD差分

S1の実装DDLは[001_init.sql](../../../backend/migrations/001_init.sql)、関係は[概要ERD](../../architecture/erd.md)を参照する。原設計のS2案はカルテ参照1列・FK/index/guardで、現在の実装段階は[統合DB契約](unified-db-20261019.md)へ記録する。旧設計の提案SQL名は公開ファイルではない。

| 資源 | 型/null/default・制約・関係 |
|---|---|
| appointments.revisit_intake_generation | BIGINT NOT NULL DEFAULT1、正数check。petがIS DISTINCT FROM旧値の時だけ+1。NULLとの付替えも増分、同pet値の再送・日時/owner/link変更では不変。世代の直接更新拒否、overflowは失敗。1は新保存機構の初期世代であり、旧回答あり/確認済みという意味ではない |
| revisit_intake_targets | BIGSERIAL PK、clinic/appointment/pet/generation必須、created_at server default。一意(clinic,appointment,generation)と(clinic,id)。appointment複合FK(id,clinic)・pet複合FK(clinic,id)、RESTRICT。作成時に予約の現在pet/世代/非取消をlock下で検査。UPDATE/DELETE不可。petは当時値なので現在appointment.petへのFKにはしない |
| revisit_intake_answers | BIGSERIAL PK、clinic/target/revision/form版/目的/認証customer/主体hash/当時owner/operation key/server受信時刻必須。7項目の他値はnullable、defaultなし。clinic+target+revision一意、clinic+author+operation一意。target/customer/当時owner/receiptへclinic相関FK、すべてRESTRICT。UPDATE/DELETE不可 |
| revisit_intake_submit_receipts | PK(clinic,author_customer,UUID key)。主体hash・fingerprint版・request hash必須。claimed→succeeded一回だけ、成功answer FKと時刻必須。claim時は対象/本文/患者rowへのFKや本文を持たず短い別TXで完了。author IDは認証scopeの不変metadata値、成功時answer側のcustomer FKで検査。成功後不変・DELETE不可 |
| medical_records.revisit_intake_target_id | BIGINT NULL、defaultなし、(clinic,target) FK。既存recordはNULL=対応世代未特定。新しく非NULLを割り当てる時はrecordのclinic/appointment/petとtargetを検査。対応参照の割当/変更/解除はD03 ownerだけ、外部編集DTOへ公開しない。変更の既存訂正/監査契約を維持 |

S1は138→141表、既存変更はappointmentsだけ。概要ERDへtarget→appointment/pet、answer→target/customer/owner/receipt、receipt→answerを示す。詳細ERDは既存正規generatorで3新表・appointments・index/manifestを再生成する。S2でmedical_record→targetとmedical_records詳細を追加する。RLS自動loop後への追加なのでS1の3表それぞれ `app_private.apply_rls_policy` を明示。ENABLEのみ、既存FORCEなしの境界を維持し、アプリのclinic/owner認可をRLSで代替しない。

全FKはRESTRICT。予約の取消やsoft-deleteでは原文を消さない。履歴/receiptは意図的にsoft-delete列も作らず、通常DELETE/UPDATEをtriggerで拒否する。保存期間・合法的削除の手続き・特権DB管理はこの変更で永久保存と決めず、別の承認対象。runtimeにTRUNCATE/DDL/trigger無効化を許さない運用境界も必要。表owner/superuserまで防ぐ保証ではない。

## 4. 7項目・フォームv1

| 項目 | 保存列・code→当時ラベル |
|---|---|
| 来院目的 | visit_purpose TEXT NOT NULL。送信時のみ必須。空白のみをserviceでUnicode-awareに拒否、DBではbtrimによる最低check。原文そのものをtrimして書換えない |
| いつから | onset TEXT NULL |
| 元気・食欲 | energy / appetite TEXT NULL、それぞれ単一値。usual=いつも通り、slightly_reduced=少し低下、markedly_reduced=かなり低下 |
| 現在の症状 | current_symptoms TEXT NULL |
| 前回より改善したか | improvement TEXT NULL、improved=良くなった、slightly_improved=少し良くなった、unchanged=変わらない、worse=悪くなった、new_symptoms=新しい症状が出た |
| 検査希望 | examination_preference TEXT NULL、consultation_only=診察のみ、medication_only=お薬のみ希望、examination=検査希望、decide_with_vet=獣医師と相談して決めたい |
| 一番相談したいこと | main_concern TEXT NULL |

選択肢の出所は仕様正本の6/3江添さんSlack。全選択肢をDB CHECKとDTO allowlistで同じ値集合にする。未知code/複数値は拒否し、最初の選択肢へ補完しない。任意自由文は未送信/null→NULL、明示空文字は空文字で保持し、どちらも表示は未回答。空白や改行の原文は保存する。form_definition_version=1を回答ごとに固定、v1のラベルを後から別意味へ差替えない。v2はDB/DTO/readerを一緒に変更する。本文長の業務上限は新設せず、既存HTTP技術上限との対応をAPI担当が検査する。

問診自体の未送信はanswer0件。予約作成・受付は空回答の作成や問診送信を待たない。選択されたpetが未解決なら回答先を名前から推定しない。スタッフによる原文編集APIは作らず、スタッフ補足は既存Inquiryへ。

## 5. 認証・認可とTX契約

### 5.1 認証scope

LiffAuthの検証済tokenからclinic/customer ID/LINE subをserverで受ける。bodyのowner/staff/clinic/回答者/timeを信用しない。主体hashは固定の版付きencoding(clinic,検証済sub)のSHA256。Bearer token・sub原文・氏名をreceiptやログに格納しない。customer IDだけの一致ではなくtoken subと現在customer.line_user_idも照合する。

未成功の新write時は、同じTXで現在のappointment.line_customer_id=customer、pet.owner_id=appointment.owner_id、非削除のownerとpet、ownerの現在LINE紐付けと認証subを照合し必要row SHARE lockを保持する。customer.owner_idは非NULLなら同owner必須。NULLのcacheは、owner側に検証済subと同じ既存リンクがある場合に限り純readで解決でき、保存機構で名前/電話を再照合・自動リンクしない。これは `findCustomerWithOwnerSync` がowner側リンクとcacheを別に扱う既存事情に対応する。新たなLINE確認ボタンやLineIDConfirmedAt必須を足さない。副飼主・医院横断の認可を拡大せず、既存患者利用認可が許す対象だけとする。

新問診GET/submitから既存healerを暗黙呼出ししない。現在の `findCustomerWithOwnerSync` は参照中に `UpdateOwnerLink` を実行し、後者はbase DB接続を使うためambient参加済みとはいえない。既存healerを維持する場合も、そのリンクwriteを上位gate下の独立した短い物理TXにする責務はLINE owner。回答保存TXからbase DBの別TXを呼ばない。

### 5.2 claimと成功処理

1. 認証とclosed DTOの構文検査。claimより前にロック・原文write・リンク修復をしない。必要な事前readの結果はcommit根拠にしない。
2. 短いmetadata-only TXでreceiptをINSERT ON CONFLICT DO NOTHINGしcommit。keyはclinic/customer/UUID。主体hashとfingerprintが違えば409 KEY_CONTENT_MISMATCH。claimは成功・失敗・実行中の証明ではない。
3. 最外の業務TX開始→既選定の `reservation-graph:write:v1` advisory gate→receipt FOR UPDATE。新しい並列gateを発明しない。claimのrow lockを持ったままgateを取りに行かない。
4. succeededなら初回operation_resultを返す。現在の取消/受付/対象世代のために初回成功を再実行・失敗扱いにしない。本文とcurrent_stateを返す時は現在認可を別に確認する。scope一致・患者閲覧権限失効時の応答は下記5.4のredacted成功形へ固定する。新writeの認可失敗と、以前の成功receiptの照会を混同しない。
5. 未成功なら必要な既存staff/clinic-booking/pet-intent/lifecycleの順を維持しappointment FOR UPDATE。問診単独は枠・勤務を変更せず新しい勤務判定を要求しない。appointment current pet/generation・受付状態・取消/deleted・現在認可を再検査。customer/pet/ownerのSHARE lockはcommitまで保持し、上位gateなしの逆順writerを残さない。
6. 同targetの最新revisionを読む。初回expected_revision=nullは最新なしの場合だけ、既存訂正は正規形十進文字列の一致だけを許可。不一致は409 REVISION_CONFLICT、入力保持。last-write-winsは採用しない。
7. 現在targetがなければINSERTし、回答をrevision=旧+1でINSERT。target guardとanswer guardもappointment lock・pet/世代・受付後拒否・連番・claimを検査。answerのAFTER INSERT triggerが同じTXでreceiptをsucceededへ進める。これが専用原文監査sinkの原子的成功。どの段階の失敗もtarget新規分/answer/receipt成功をrollbackし、先のclaimだけが残る。
8. commit前に成功answer/receiptを再読込し返却payloadを組み立てる。ネットワーク応答喪失は失敗確定ではなく同keyで照会/再送。予約・カルテ・会計・検査発注・LINE通知は回答送信から作らない。

fingerprint v1はintent、appointment、pet、target_generation、form版、7項目を型付き正規JSONにしたSHA256。任意未送信はnullへ統一、自由文の空文字/改行は保持。既選定契約に従いexpected_revision・再送時刻はfingerprintから除外する。未成功の期待版修正は同じ内容で再検証可能だが、一度成功したkeyは新しい訂正には再利用しない。target世代は操作対象そのものなのでhashに含め、旧keyを新targetへ転用できない。

receiptには実行中lease/失敗確定の推定を追加しない。succeeded / claimed_uncommitted / not_found / result_unavailableを分ける。not_foundは「前回失敗確定」ではない。claimedの清掃・TTLは保存期間が未承認なので追加しない。原文に成功receiptが欠ける状態はAFTER INSERT失敗でTXを止める。

`committed_at`は成功receiptへ同じTXで記録するserver時刻であり、PostgreSQL WALの厳密なcommit timestampではない。本文hashと認証scope照合はアプリ責務、triggerはtoken認証・患者grantを代替しない。SQL guardは現在定義済みの受付前状態集合を明示し、新enum追加時のfail-openを防ぐ。tentative/no_showの値だけで新しい送信禁止を設けず、実際のLINE予約との関連・認可を別に検査する。

### 5.3 受付・取消・対象変更

受付前かは現在の予約lifecycleで判定する。checked_in/in_consultation/accounting/completedはowner新write不可、cancelled/deletedも不可。checked_in_atを「一度でも受付したら永久停止」のflagにしない。no_showだけ、過去日だけ、未回答だけを理由とする新しい締切・予約停止を追加しない。新しいstatusが導入されればadapter未対応のまま送信許可しない。受付取消を新設せず、将来の正式な受付訂正は既存業務判断に従って別に接続する。

| 先にcommitした操作 | 後の処理 |
|---|---|
| 新回答→受付/取消 | 受付/取消は旧回答を保持。内部の成功answer ID/revisionは不変。成功再送のwireは5.4の本人receipt scopeと現在の行認可に従い、失権時はID/revisionを除いたredacted成功。新keyの新回答は拒否 |
| 受付/取消→新回答 | appointment再検査で拒否。answerも成功receiptも作らない。claimが残っても成功表示しない |
| pet A→B→A→旧target回答 | 世代は1→2→3。旧世代のsubmitは拒否、旧answerは保持。新予約は別appointmentで未回答 |
| 日時変更→回答 / 回答→日時変更 | pet/世代/原文と時刻は保持。回答書込でappointment.updated_atをtouchしないため、日時変更との不要な版競合を増やさない |
| 2端末同版訂正 | appointment lock後に期待revisionを再確認。一方だけ成功、他方入力保持/競合。DB uniqueも重複を防ぐ |
| unlink/主飼主移管→回答 | gate下の現在認可で拒否又は現ownerへ認可し直す。旧回答者・原文を新ownerへ再帰属しない。同じpetならtarget世代は変えない |

### 5.4 履歴の行認可・失権後の結果wire（独立レビューF1/F2修正）

予約の現在targetを読めること、過去targetのpetを現在利用できること、個々のanswer原文を読めることを別に検査する。appointmentを一度認可して全target/answerを返すqueryは禁止。A→Bの後にBだけを利用できる本人へ旧A原文は返さない。

LINEの初期read allowlistは各行について `(target.clinic=認証clinic AND target.petが既存の現在患者利用認可対象 AND answer.author_customer_id=認証customer AND answer.author_subject_hash=認証主体hash)` とする。予約refのアクセスも既存LINE予約scopeで別に検査する。旧author原文を新ownerに公開する既存grantは確認できていないので、この案から追加しない。スタッフ履歴は既存staff view grant・clinic・target.petの利用scopeを各行で検査し、LINE本人向けallowlistとは分離する。

同じpetの主飼主移管後に別authorのrevisionが共存しても、原文の回答者は不変。現在targetのhead_revisionは全answerの最大revisionで計算する。本人に可視の履歴だけからMAXを取り直して「最新回答」にしない。現在head本文が他authorで不可視なら `head_revision` と「回答済み」状態だけを返し `latest_answer=null, content_available=false` とする。旧自分回答を最新としてprefillしない。現在認可と受付前条件を満たす本人は、このhead_revisionを期待版に明示的な新送信をできる。履歴は各行allowlistに通る原文だけ、旧targetは旧targetとして表示し現targetの入力へcopyしない。これは他作者原文の新規公開を避ける技術allowlistで、主飼主移管そのものを新設/変更する判断ではない。

結果照会/成功再送は、検証済tokenのclinic/customer/subject_hashがreceipt scopeと一致する場合だけ処理する。違うscopeはreceipt有無を区別できない404、認証失敗は401。患者本文の権限が失効していても、同scopeが自分の成功事実を照会した応答はHTTP200で次の形に固定する。

```json
{"operation_result":{"operation_id":"<元UUID>","intent":"revisit-intake.submit","state":"succeeded","committed_at":"<保存server時刻>","result_refs":null,"committed_versions":null},"current_state":null,"content_available":false}
```

このredacted形にanswer/target/appointment/pet/owner ID・revision・本文を含めない。新しいwriteも行わない。現在の行認可まで満たす場合だけresult_refs/committed_versionsと許可されたcurrent_stateを返す。receipt自体はあるが成功参照を取得できない異常では `state=result_unavailable` と元operation_idを返し、成功/失敗の断定や新規createをしない。claimed_uncommitted/not_foundも実行中/失敗確定へ変換しない。新しい未成功submitが認可失効で拒否される場合は通常の認可エラーであり、この過去成功の照会とは別。


## 6. writer / reader 参加条件

詳細表は `revisit-writer-reader-matrix.json`。新機能利用開始前に実symbolと共通gate・ambient TX・認可・結果を検査する。表があるだけで全writer参加済みとはしない。

- 受付：`guardedCheckIn` / `checkInConfirmed`、汎用Update checked_in、tentative受付、受付/record作成、trimming状態遷移。現在のappointment lock/updated_at CAS/患者生存・医院検査を維持し、原文待ちはしない。
- 取消/削除：院内Updateはstatus更新+soft-deleteの同TX、LIFF `CancelByID`は取消行を残すUPDATE。どちらも同じ上位gateへ。既存取消権限・日時条件は維持し、回答仕様から取消可能条件を広げない。既存recordの後処理からanswerへCASCADEしない。
- pet/日時変更：汎用patchの全分岐、管理API、trimming、tentative昇格/吸収、直接import。世代はDB triggerがペット実値変更を捕捉する。テスト/seed/importでtriggerを無効化する場合は通常write停止と保全計画が必要、未承認DB操作は実施しない。
- 身元変更：owner LINE link/unlink、Lステップtoken/link/webhook、customer owner link/healer、主飼主移管・pet_owners・identitylinkの認可に影響する変更。既存移管NOWAITを維持。共通gateは認可を与えず、参加は既存clinic/grant検査を省略しない。
- 予約詳細：1つ目のstaff reader。cancelled/deletedの原文履歴は認可した専用history queryで取得し、通常一覧のsoft-delete条件を全体解除しない。最新targetと旧targetは別表示、履歴を最新回答へcopyしない。
- LINE：完了/詳細のGET/submit/history/result lookup。GETは純read、1 SQL snapshotでtarget/最新本文/revision/受付状態を組み立て、保存時は必ず再検査する。
- D03：後続reader。record作成/受付のappointment lock下で現在targetを確保し、record FKへ保存する。未回答でもtargetだけ作れてanswerなしのまま受付可能。copy/clear/Undoはbinding/answerを入力payloadに含めない。readerは固定target参照とrecordのclinic/appointment/petを照合し、現在targetへ追従させない。recordの患者/予約訂正で不一致なら原文を表示せず出所不明として、既存訂正ownerが明示的に再対応付けする。一般record編集に新しい一律禁止を加えない。

`reservation-graph:write:v1`は文書で選定済みだがこのmainのpersistence/sharedkernelに共通実装を確認できていない。新問診だけにgateを付けても他writerとの直列化は成立しない。R1に関係するblocking closureの実装/参加検証が利用開始条件であり、全DB設計完了を待つ条件とは異なる。DDLだけを先行公開しても回答機能の利用開始は別判断。

### 6.1 実入口と最外TXの参加位置（独立レビューF3修正）

| 実入口 | 現在のTX/下位lock | 必要な差分・検証点 |
|---|---|---|
| owner/service_core.go:Delete → repository.go:LockForDelete / Delete | deleteTransactor.WithTxの中でowner row lock。guardなしcompatibility分岐も存在 | サービスの最外物理TX開始直後にgate、その後既存owner依存検査・削除。guardなし分岐を漏らさず同じwrapperか非稼働証跡へ。repository末端で後付けgateは不可 |
| owner/repository.go:UpdateAndFind / UpdateAndFindApplying | DBOrTx.Transaction→LockByIDForUpdate→apply/UPDATE | 外側TXありならその開始直後にgate済みを検査、外側なしならhelperが開始した物理TXの先頭でgate。savepointでgateを取って外側既得lockを正当化しない。name等だけの更新でも同じgeneric入口がowner rowを取るため経路を棚卸し |
| pet/repository.go:Update / UpdateAndFind / withPetUpdateTransaction | helper内Transaction→pet UPDATE lock、FK再検査、移管participant | owner側と同じ最外TX規約。pet SHARE→owner SHAREの問診順と既存pet→owner/NOWAIT順を共通上位gateで接続。metadata・死亡更新を新しい問診必須条件にはしないが、既存lock graphへの参加検査対象 |
| pet/service.go:Delete → repository.go:Delete | DBOrTxのsoft-delete。直接repo呼出しは外側TX有無を断定できない | ownerによる最外TX+gateを必須にし、現在認可の非削除判定と直列化。直接repo呼出しが残るならambient専用化/互換wrapper/明示停止のいずれかを実装ownerが固定 |

`revisit-writer-reader-matrix.json`では各実入口にouter_tx、gate_point、current_lower_lock、implementation_statusを記録する。旧W16の「direct import/batch/raw writer inventory」は実symbolではなかったのでwriter件数から外し、別の `inventory_obligations` に移す。CSV/import/直接SQLの未発見入口がないと断言せず、使う実入口の棚卸し証跡を利用開始条件にする。

最初の実装は「submitを作り最後にgateを足す」順にしない。I0＝共通gate helper＋最外TX所有の検査＋上記実入口/受付/取消/認可変更の参加・逆順fixtureを先行又はS1の同じ変更単位に含める。その後I1＝S1 DDL・保存/readerを接続、I2＝S2 D03へ進む。I0を全billing/全master改修へ無条件拡張せず、R1が保持するappointment/pet/owner/customer/receiptと実際にblocking edgeを持つ経路を証拠付きで閉じる。新たなblocking edgeが見つかれば該当経路の参加又は停止が必要。現段階は全rowをNOT_IMPLEMENTEDとしており、Ready/利用開始の技術blockerである。

## 7. 旧値・切替・復旧

1. 既存回答移行は0件を初期案とする。プロフィール/Inquiry/顧客JSONから原文を捏造しない。追加列generation=1、record binding=NULL、3表空をfresh/既存入力の技術基準とする。
2. 世代1を旧予約作成以来の履歴とは解釈しない。機構開始前のA→B→Aは不明のまま。新フォーム利用開始前に世代triggerと全対象writer参加を有効化する。
3. 旧recordのbindingは現在appointmentから一括backfillしない。新コードが同TXで記録するbinding、又は出所が証明された明示移行だけを採用する。NULLは「対応不明」であり「問診未回答」とは別表示。予約現在targetにanswerがない場合だけ未回答と言える。
4. 001/ERD/checksum差分の公開と実環境への適用は別。既存baselineへの新001実行はchecksumで止まり得る。保全・採用revision・復元/移行方式の承認なしにchecksum更新/repair/resetを行わない。EMR443と計画を合わせる。新DDL/seedを実行済みとはしない。
5. API/readerと全関連writerを先に対応し、利用開始制御をownerが持つ。共通切替資源が未実装なら「DBに列がある」だけで公開しない。今回このsubsetのために汎用全schema切替基盤を作ることも決めない。
6. 復旧はフォーム送信を止め、新3表・世代trigger・readを保持する前進互換方式。旧writerで対象世代を失わせたり、成功receiptを消して再登録させたりしない。保存期間/cleanerは別途定める。

## 8. API/UI・DB担当の境界と実装順

DB担当EMR446：S1 [001_init.sql](../../../backend/migrations/001_init.sql)の3表・予約世代1列・制約/trigger/RLS、001 checksum差分検査、概要ERD・正規生成詳細ERD。S2のrecord参照列・FK/index/guardはD03 capture/readerと一緒に別限定差分にする。原文データseedなし。schema/parserを弱める変更を要件にしない。

予約/API担当EMR452→460：専用model/repository/service、認証principalのserver抽出、共通gate/receipt/全関連writer、Unicode必須検査、fingerprint、期待版、既存認可、履歴/結果query、型付きroute/DTO/OpenAPI。既存予約の汎用customer_fieldsへ新回答キーを足さない。IDとrevisionは正規形十進文字列、operation key UUID。共有生成型/routerは統合ownerが変更する。

UI担当：任意導線、ペットごとのdraft、元気/食欲を別の単一選択、履歴原文を編集せず訂正新送信、入力保持・競合/通信断/結果照会、飼主由来と時刻。院内スタッフ補足と原文を別欄。受付後は閲覧とスタッフ連絡へ。未回答で予約/受付を止めない。患者内容/tokenをURLやlocalStorageへ保存しない。

D03担当：保存原文へのread-only接続、record bindingのcapture participant、clear/copy/Undo除外。新しいカルテ作成を回答保存の前提にしない。

候補routeは既存 `/api/liff/:clinicId/...` 配下の予約ref→`revisit-intake` GET、`revisit-intake/submissions` POST、`revisit-intake/operations/:key` GET。staffは既存予約詳細view grantに接続する専用GET。route名はAPI ownerが正本へ固定する技術作業で、業務質問ではない。結果照会routeは現在予約のview認可middlewareから分離した `GET /api/liff/:clinicId/revisit-intake/operations/:key` を候補として固定し、検証済本人receipt scopeで検索する。現在予約を読めないことだけで照会を先に403へ落とさない。別scopeへreceiptの存在を漏らさない。submit DTOはtarget pet/generation、expected_revision(null又は文字列)、form版、7項目とUUID headerのみ。clinic/actor/time/status/staff補足は禁止属性。成功はoperation_resultとcurrent_stateを分ける。

**最小縦割りS1**：3表＋予約世代1列＋保存/訂正/結果照会＋予約詳細reader、受付/取消/付替え/認可writer参加。record列/guard/indexは含めない。**S2**：record参照1列とbinding capture/治療プランreaderを一緒に追加し、AC-provenanceの両画面を検証。全EMR460の設計総体は3表・2既存列だが、S1だけでEMR460全完了としない。各subsetのfixture/承認済み実行先が揃ってから限定DB/API/UI検証へ進む。

## 9. 受入と残確認

6業務ACへ追跡する具体caseを `revisit-acceptance-matrix.json` に定義した。2頭/カルテなし、受付/取消/付替え/認可失効との両先行順、同key同内容/異内容、revision競合、監査sink失敗、commit後応答喪失、soft-delete履歴、D03出所を含む。すべて未実行。

残るのは技術採用レビュー、実symbolへの全writer接続、Unicode/closed DTOとSQLの対応、共通gate参加と権限/row lock、実DB/実API/実機/restore検証、保持期間と運用切替の具体化。既決の7項目・任意提出・受付後原文編集禁止を未決へ戻さない。今回DB適用やAPI実装を行う必要はなく、未実行を実装可能性の合格証拠へ読み替えない。

## 10. 差分単位数とPR727

残DB調査の確定一覧は **R1、T1、T2、T3、T4、A1、A2、A3、A4＝9単位**。前ターンの途中報告「10単位」は列挙数の単純な誤記。第10単位を統合/削除した履歴はなく、レポート表・検証JSONとも当初から9行だった。EMR449のactive consumerはA3内、今回のS1/S2はR1内の実装分割で、新しい第10単位にはしない。

PR727の期待headは `2707af868dd3928119b5dd44a6c123704aec401a`。観測時OPEN、Draft解除済み、CodeRabbitレビューはCOMMENTED、Actionable1件。CodeRabbit status SUCCESS / Review completed、CI4SUCCESS。指摘は未解決として保持し、SUCCESSを内容承認と呼ばない。

GraphQLでスレッド `PRRT_kwDOQ4_gTc6t1qsH` の `isResolved=false` / `isOutdated=false` も確認した。resolve操作は行っていない。

正確なpermalink： https://github.com/MinoruSoga/AnimalEkarte/pull/727#discussion_r4311735532

本文要旨：Major / Data Integrity & Integration。001へ4列を追加すると旧baseline DBがchecksum検査で止まるので、001を元に戻して新番号migrationへ移し、checksumテストと生成schema文書も更新するべきという指摘。botは一般の「001編集禁止」規則を根拠に、今回の明示承認を例外と認めていない。

評価：checksum影響は事実。今回のユーザーによる001編集許可と、旧DBへ自動適用しない方針をbotの一般規則とは区別する。新番号へ変更・checksum修復・スレッドresolveはしていない。次の001差分も既存DB利用開始には別の保全/適用承認が必要。このコメントは上記GitHub permalinkで確認できる。原調査のJSON receiptはローカル証跡で、本公開文書には同梱しない。観測時点の状態を現在の自動レビュー承認へ読み替えない。

## 11. 根拠・検証限界

調査checkout32ca61d、観測main992a2b8。間の2commitはSTG runbookとpreserve scriptで対象DDL/modelは不変。原調査の `source-evidence.json` はローカル証跡であり、この文書には同梱しない。公開可能な正本・現consumerは[統合DB契約](unified-db-20261019.md)と[writer/reader表](revisit-writer-reader-matrix.json)へ接続する。

読み取った主要経路：reservation_guarded_checkin.go（row lock/CAS）、reservation_service_update.go（複数分岐）、appointment_admin_repository.go:151（LINE取消は非削除）、pet_owner_transfer.go（owner/LINE差替え・NOWAIT）、liff_owner_sync.go（read中write）、lstep/line_customer_repository.go:116（baseDB）、owner/repository.go:571（LINE解除）、audit/service.go（staff/systemのみ）、既存001の統合FK/unique/RLS。

原設計時点ではSQLは未実行artifactだった。その後S1と保険列の統合候補は、新規tmpfs PostgreSQLで001適用・保険SQL・S1 DB16件を検証し環境削除を確認した。この限定成功は実環境適合、API/UI・全writer参加、C01–C26の受入、残DB設計の実行証拠ではない。統合候補の公開・残試験は統合DB契約で別管理する。
