# EMR-494 編集権・本人暗号draftと EMR-495 接続契約

採用日: 2026-10-20。状態: **技術設計採用済み・未実装・検証 NOT_RUN**。識別: `DESIGN-494-20261020-R21`。

Dots採用のA（編集権）／B（本人暗号draft）分離、server broker＋memory DEK＋ciphertext journal方式、Revision 2のschema／API細部、および最終Revision 2.1（13:20 UTC）のclaim lock順・CAS baseline補正を統合した。旧「提案／親最終確認待ち」「方式未選定」「足りなければanchor DDL」は現在の状態を表さない。**Aのmetadata claim anchorはbindingへの複合FKのみとし、medical_records FKもclinics直接FKも付けない。**

調査基準mainは `da27989fa812c1319b62a1c3b4684c419b110fe1`、公開作業の基準mainは `86c9caa34d8498133eb46d4f019c0b08549db554`。本書の根拠となるD03／C06／schema／実装ファイルは両者間で未変更。repoリンクは公開基準SHAへ固定する。[Plane 494](https://app.plane.so/baritechllc/browse/EMR-494/)のR2.1採用本文（更新13:24 UTC）とコメント `77e37512-dc9b-4124-90ee-062f39709130`（更新12:30 UTC）を再確認した。

## 1. 採用範囲と既決業務

- 494 AとBは独立PR。495 Aの複数診断DTO／reader／adapter移行は両方の完成を待たず開始できる。
- Aの最小実装は編集権の状態readと操作・結果照会、および通常問診PATCHの実fence参加。取得・移譲だけで完成にせず旧epochの実保存を拒否する。未参加writerが残ったままカルテ全体の単独編集保証を有効化しない。
- Bはserver側本人認可付きkey broker、browser memory-only DEK、端末永続領域は暗号文のみ。設計／コード／fake fixtureのPRは実鍵なしに進められる。実鍵生成・登録、credentials／KMS権限設定、live provider接続は別途明示承認を要し、本PRでは行わない。
- 閲覧で編集権を取得しない、idleだけで解放しない、一般編集権スタッフの確認付き強制取得、logout後の本人draft保持、再認可・最新比較・明示選択は既決。管理者限定や追加の理由必須、保持期限・自動削除・端末間共有を新設しない。
- 診療終了、医師確認、会計確定、編集権を別軸として扱い、削除済み `medical_records.status` や旧finalized一律編集禁止を復活させない。

業務正本: [D03編集権取得・解放](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/spec/domains/medical-records/lifecycle.md#record-edit-right-lifecycle-20261018)、[logout後の本人回収](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/spec/domains/medical-records/lifecycle.md#draft-logout-recovery-20261018)。本書は既決を満たす技術方式を固定し、業務承認を再質問するものではない。

## 2. 現状と未接続箇所

| 一次資料／現コード | 静的に確認した事実と接続要求 |
|---|---|
| `backend/internal/model/storage_clinical.go` | RecordEditRight／RecordEditRightEvent／StaffRecordDraft／StaffRecordDraftOperationはモデルあり。consumer完成を意味しない |
| `record_edit_rights` | PK(clinic_id,medical_record_id)、epoch>0、free／held、holder_staff_id、session_ref UUID、時刻 |
| `record_edit_right_events` | previous／resulting_epoch、intent、prior／new_holder、operation_id |
| `staff_record_drafts` | unique(account×clinic×record×tab)、input_generation、base_owner_versions、ciphertext、encryption_format、key_reference、hash。device／epoch列なし |
| `medical_record_dependency_versions` | PK(clinic_id,medical_record_id,owner_kind)、revision bigint>0。問診CASに再利用 |
| draft RLS／`auth/http_session_me.go` | clinic access AND trusted app.current_account_id。`/me` userはstaff、accountはstaff.AccountIDから別解決。user.idをaccount IDへ流用しない |
| `routes_records.go` | 問診PATCH、clinical-plan PATCH／DELETE、record PATCHあり。編集権専用routeはこの登録ファイルにない |
| `inquiry_repository.go` | SaveByMedicalRecordIDは独自TX／親row lockを持ち、削除済み親status参照が残る。ambient participantへ移さない名目fence追加では同一TXにならない |
| `clinical_plan_service.go`／response／model | Update／Deleteの旧draft lock、GetOrCreateのread経由createを参加台帳へ。responseの旧DiagnosisTypeID等と最終modelの不一致は495 Aで移行 |
| `operation/{contract,dto,runner,store}.go` | Gate→binding→Apply、canonical UUID、現認可付きreplay、ResultResponseが実在。独自receipt規則を複製しない。bindingにはrecord IDがない |

根拠: [D03詳細設計](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/design/domains/medical-records/technical-spec.md#integrated-save-edit-right-proposal)、[C06共通契約](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/design/shared/integration-contracts.md#detail-design-proposals-20261018)、[詳細ERD一覧](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/architecture/erd-detail/index.md)。現状は静的観察であり、稼働DBやruntimeを確認した結果ではない。

## 3. A: HTTP／DTO契約

本節は採用設計であり新route／wireの実装済み記録ではない。共有API ownerが `backend/docs/api.yaml` へ正式反映し採用SHAを公開する。既存serversの `http://localhost:8080/api/v1` はローカル開発用の例。デプロイ時のbrowser-facing API／key-broker originはHTTPSを必須とし、HTTPは認証処理・`/key-access` に到達させず入口で拒否する。既存TLS終端構成を利用し、本書の修正で証明書・ネットワーク・credentialsを設定したとは扱わない。pathsはprefixなしの `/medical-records/...`、実URLは `/api/v1/medical-records/...` とし二重に `/v1` を加えない。Bの暗号利用も§9からこのtransport境界を参照する。

既存HttpOnly Cookie認証とCSRF `X-Requested-With: XMLHttpRequest` を継承する。operation_key／edit_session_ref／edit_epochまたはexpected_epoch／owner_expected_versionとowner payloadはJSON bodyの平坦なfieldに置く。queryでclinic／accountを選ばせず、新しい `{data:...}`／`{operation:...}` wrapper、Authorization／edit bearer方式、別Idempotency-Keyヘッダを追加しない。

### 3.1 資源操作

| method／path | 入出力・成立条件 |
|---|---|
| GET `/medical-records/{id}/edit-right` | 現在view認可と医院／record／petを照合。state=free／held、正のdecimal文字列epoch、認可済みholder表示projection、acquired_at、is_current_sessionを返す。他session_refを返さず、readで取得／延命しない。unknownは通信／取得失敗のUI状態でDB stateではない |
| POST 同 `/acquire` | operation_key、編集session UUID、expected_epoch。staffは認証から。freeならepoch増分・holder／session・event・receipt・問診baseline準備を同一TX。heldから移転しない。同session・同key成功再送はreceiptのみ |
| POST 同 `/take-over` | operation_key、新しい編集session UUID、expected_epoch、confirmed=true、confirmed_holder_staff_ref。確認対象が変わればdependency／version conflictで再確認。管理者限定・理由必須は加えない |
| POST 同 `/release` | operation_key、編集session UUID、expected_epoch、未解決operation refs。現在権と対象保存結果を照合後のみfree＋epoch増分。clientの「未解決なし」だけを信用せず、同key成功再送は元receipt |
| GET `/medical-records/{id}/operations/{operation_key}?intent=...` | 元targetをpersisted metadata anchorから解決し現在認可後のみ返す。intentは固定allowlist。URL recordと元targetが不一致なら404。URL recordをresolverの根拠にしない |

編集session UUIDのwire fieldはR2-WIREの `edit_session_ref` を使い、DBの `session_ref` と区別する。通常writeは `operation_key, edit_session_ref, edit_epoch, owner_expected_version` とowner payloadを同bodyへ置く。owner版とepochは別。C06 Contentは業務payloadから、現在fence／CASはPreconditionsへ。強制取得で確認したholder／epochは確認内容としてfingerprintへも結び、意味が変わる再確認は新keyにする。認証tokenやopaque編集証明をURL／ログ／draftへ保存しない。

schema名は `MedicalRecordEditRightResponse` / `MedicalRecordEditRightAcquireRequest` / `MedicalRecordEditRightTakeOverRequest` / `MedicalRecordEditRightReleaseRequest` / `MedicalRecordOperationResult`。問診のowner fieldsは既存update requestへ追加する。成功後の画面内容はreadで取得し、receiptへ問診本文を複製しない。

### 3.2 C06結果wireとHTTP mapping

`MedicalRecordOperationResult` は既存 `operation.ResultResponse` の `status, operation_key, result_refs?, committed_versions?, recorded_at?` と同一形。statusをstateへ改名せずreceipt／results／dataラッパーやEffectSummaryを返さない。

| C06 status | このA consumerのHTTP |
|---|---|
| succeeded／replayed | 200、保存済み成功結果 |
| validation | 400 |
| version-conflict／dependency-changed／edit-right-lost／operation-key-conflict | 409 |
| operation-in-progress | 202、同key照会のみ |
| scope-not-found | 404 |
| not-authorized | 403 |
| result-unavailable | 503、結果不明として停止 |
| not-succeeded | 200。成功ではなく、人の再確認後に同key再試行する候補 |

入口のJSON不正・認証欠落等は既存 `ErrorResponse {error:string}` と400／401／403／404。C06到達後のstatus付き結果は平坦ResultResponse。internal error詳細・鍵・fingerprintは公開しない。共通runnerのstatus enumは変更しない。

## 4. A: claim anchorと489の限定追加subset

### 4.1 metadata anchorの確定schema

`operation_bindings` の実列はid／clinic_id／owner_kind／actor_kind／actor_ref／intent／operation_key／fingerprint／fingerprint_key_version／created_atで、record IDは存在しない。fingerprintから逆算できず、成功event／receiptだけでは不成功bindingの認可元を解決できない。新表 `medical_record_operation_targets` を採用する（未作成・未適用）。

| 列・制約 | 採用定義 |
|---|---|
| clinic_id／binding_id／medical_record_id | すべてbigint NOT NULL。medical_record_idはCHECK >0の不変metadata |
| PK | `(clinic_id,binding_id)` |
| **唯一のFK** | `(binding_id,clinic_id) REFERENCES operation_bindings(id,clinic_id) ON DELETE RESTRICT` |
| 付けないFK | **medical_recordsへの業務FK、clinicsへの直接FKを付けない**。clinic境界はbinding複合FKとRLSで維持 |
| RLS | 既存app_private.apply_rls_policyによるhas_clinic_access(clinic_id)。actor／intent／current grantはresolverで別途再照合 |

1 bindingの元recordは1件で不変。Aのacquire／take-over／release／inquiry.saveはowner_kind=`medical_record`、intentは `medical_record.edit_right.acquire` / `medical_record.edit_right.take_over` / `medical_record.edit_right.release` / `medical_record.inquiry.save` とする。

anchorは臨床効果／成功履歴ではなく操作bindingの対象認可anchorで、不成功時も残す。成功eventとだけ同時作成する方式は禁止。record論理削除でも付替えず保持し、削除済み結果の開示は既存現在閲覧grantに従う。解決不能／現在読取不許可ならfail closed。

### 4.2 ClaimTargetとlock順（R2.1）

489が所有する既存Runner ClaimBindingへ、オプショナルなowner `ClaimTarget` participantを追加する。一般consumerへD03知識を入れず、callbackに同TXと新Bindingを渡す。A participantは必須として提供する。

1. claim前にserverが対象clinic／record／現在既存grantを**非lockingで早期検査**する。これは業務成功の認可根拠ではない。
2. 新binding INSERTとanchor INSERTを同じ短いclaim TXでcommitする。targetはfingerprint対象Contentのrecord_refから導き、別captured targetを使わない。`INSERT ... ON CONFLICT DO NOTHING RETURNING` で新規作成者か判別し、新規時だけcallbackを実行する。anchor失敗はbindingもrollback。
3. 既存bindingにincoming targetを挿入／上書き／補完しない。同key別recordでも元anchorをresolverで読み元targetの現在認可を適用する。旧anchor欠落はsource-unresolved相当の `StatusResultUnavailable` または `StatusNotAuthorized` でfail closedとし自動補修しない。
4. claim TXはbindingと不変metadataだけを保存し、既存C06のbinding／clinic lock境界を広げない。record FK検査のKEY SHAREを暗黙取得し共通gate前に業務rowをlockする逆順を作らない。業務row FOR UPDATE、編集権取得、owner CAS、必須監査、成功receiptは後段業務TXに置く。
5. claim TX終了後に必要なshared gateを取得する。後段はshared gate→既存lifecycle保護→edit row→record／owner順で現在存在／scope／grantを再検査する。claim直後のrecord変更／論理削除を成功させず、anchorだけ残る失敗として扱う。
6. lookupもanchor自体を認可証明にせず、現在record／clinic／本人actor／intentの現在grantを確認する。locking readが必要なら既存共通gateの後。binding lockを保持したままgateを後取りしない。

Runnerのambient TX内からExecute不可を維持し、ClaimTargetなし既存consumerの挙動は変更しない。[runner](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/operation/runner.go)・[store](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/operation/store.go)・[LookupAuthorizer契約](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/operation/contract.go)への限定変更は489へ引き渡し、494が共有ファイルを並行上書きしない。

494 API／FEモックと495診断adapterは先行可能。A runtimeのlookup／再送接続はanchor付きclaim subsetに依存する。

## 5. A: 最初の実writerと問診CAS

最初の対象は **PATCH `/medical-records/{id}/inquiries`**。494が `inquiry_{handler,request,service,repository}.go`、FE `api/inquiries.ts` とactive-tab save hookの参加を担当する。repository独自TXをouter ambient participantへ移し、必要な経路はGateで先に参加する。

現在staff→account対応、clinic／record／pet、grant、held／holder／session／epoch、owner CAS、必要履歴・audit、依存版、receiptを同じ業務TXへ入れる。既存予約gateと順序を照合し、binding lock→既存lifecycle保護→編集権row→owner／依存子の安定順を維持する。全ownerへ無条件に予約lockを拡大しない。

### 5.1 問診namespaceとbaseline（最終R2.1）

独自CAS列／表は増やさず `medical_record_dependency_versions.owner_kind='inquiry'` を採用。API `owner_expected_version` は正のdecimal文字列、`committed_versions` keyは `inquiry`。

acquire／take-over成功TXの具体順序:

1. shared gate（必要経路）→既存lifecycle保護→`record_edit_rights` row→`medical_records` row。
2. 対象recordの**既存inquiry rowをID順FOR UPDATE**。
3. inquiry dependency行がmissingなら `(clinic_id,record_id,'inquiry',1)` を `INSERT ON CONFLICT DO NOTHING`、続いてSELECT FOR UPDATE。

baseline revision=1はこのlock下で観測した現在内容（未作成も含む）のCAS開始点であり、昔の履歴を生成したという意味ではない。inquiry未作成ならrecord rowをcreation serialization pointにし、baseline用の空inquiryは新設しない。既存inquiryのlock前にbaselineを準備しない。bootstrap失敗なら編集権取得もrollbackし、編集可能と表示しない。acquire成功応答は準備済みowner_versionを返す。

初回readに版rowがない場合は `owner_version=null, version_state=uninitialized`。readで0／1を捏造せずINSERTもしない。初回acquireで準備してから入力可能にする。inquiry writerは版row欠落時「準備未完了」として拒否し、保存時に無条件bootstrapして古いreadを承認しない。

変更writeはfence／認可後、`revision=expected` 条件の更新を同TXで行いrevision+1（初回1→2）。実変更なしno-opは増分せず現在版を成功receiptへ。既存PATCH omitted／null区別を保つ。inquiry作成／更新／clear／copy／Undo／内部adapterの全writeは親→inquiry→dependency順と同namespace増分に参加する。GetOrCreateや旧writerが親lockなしで作成する場合はadapter閉包に含め、未参加のままAをruntime切替しない。

問診履歴・audit・dependency増分・receiptの同時commitを維持し、旧statusを復活させない。医師確認等の別owner版をinquiry版へ合算しない。495のclinical_plan既存version／clinical_plan_diagnoses行versionは別で、将来dependency namespace=clinical_planをprojectionに使っても診断行versionと同一視せず、494が勝手に更新しない。

### 5.2 未参加writerとsession境界

header PATCH、clinical-plan PATCH／DELETE／GetOrCreate、vitals、treatment／bulk sort、images、checkups、vaccinations、旧client・内部writerを参加台帳へ列挙する。限定PR完成と全体切替を区別する。D07機器結果等の独立ownerには人のsessionを一律要求せず、自owner認可／CAS／履歴／receiptとD03依存版gateへ接続し、本文編集bypassを作らない。

session_ref単独をbearer権限にしない。auth identity＋現在grant＋stored session／epoch一致を必須とする。同staffでも別タブは別session。reload後は自動取得／自動再送せず、新sessionで明示取得または確認付き強制取得する。session_ref共有・URL露出・localStorage保持で別タブを同一扱いにしない。失権時は旧epoch保存を拒否し入力を保持する。

## 6. 494 Aと495のPR境界

| 単位 | 完了範囲／依存 |
|---|---|
| 494-A0 | API／DTO、fence participant interface、writer参加台帳、operation lookup metadata設計。Bなし |
| 494-A1 | 編集権操作＋inquiry実writer＋audit／CAS／receipt。隠れwriteを含め、全体運用有効化はまだ行わない |
| 494-A2 FE | API公開SHA固定後、モックで閲覧／取得／確認／競合／権移転／入力保持を独立PR化。Bや全BE完成を待たない |
| 495-A | clinical_plan_diagnoses slot1／2各複数行、sort_order／id、nullable参照、当時snapshot、versionをrequest／service／repository／response／API／FE adapterへ通す。旧単一値へ切捨てず、494-A0 interfaceの準備を使いB待ちにしない |
| 495-B | consultation event、通常編集／過去保存版・編集者履歴。実writeは494 fenceへ接続。診療終了は編集権／医師確認／会計確定と別 |
| 495-C | correction_links、元先record、理由／確認／実診療日／元revision／receipt。複数recordの編集権rowはID順。原患者ID付替えや自動受付／請求／在庫作成なし |

routes／api.yaml／codegenは単一ownerが統合する。494は495所有の `clinical_plan_*` を並行編集しない。B鍵設計と495-Aは並行可能。495は必要subset受領で進め、全BE完成を一括gateにしない。

## 7. B: 採用方式とscope

| owner | 責務 |
|---|---|
| Browser | 暗号化／復号、memory-only DEK、ciphertext journal、dirty／比較UI |
| Server draft owner | 本人とclinic／record／tab、generation CAS、operation refs、current grantを照合して暗号文保管 |
| Server key broker | opaque key_referenceから同本人・元scope・現在認可を照合後だけDEK利用を許可。server KEK／KMSでwrapped DEK保管 |
| 認証／TX | trusted staff→accountの現在有効対応をserver解決。欠落／不整合時は本人draft利用不可、staff ID fallbackなし。TXごとに `app.current_account_id` をSET LOCALしpoolへ漏らさない |

scopeは既存の **account×clinic×record×tab**。tab_keyは臨床UIの安定tab区分でbrowser editing sessionと別。別タブciphertextの自動復元、端末間draft一覧／共有／転送／回収UIを新設しない。device情報やランダムbrowser ID／localStorage値を認証因子にしない。journal出所識別が必要なら秘密でないopaque local journal IDを暗号化payloadにだけ持たせ、server権限へ使わない。

server unique scopeへのexpected_generation CASにより同account／tabを別sessionが更新してもsilent last-write-winsにしない。server世代とlocal未同期入力世代を区別して比較し、collisionを明示する。localの任意世代をserver版へ強制せず、旧ciphertextを自動廃棄しない。`input_generation` はserver保存内容版、`key_version` は鍵版、`edit_epoch` は現在編集権で互いに独立。初回0→1と以後のexpected+1の原子CASは§10に固定する。

## 8. B: 最小key registry DDL設計

新表 **staff_record_draft_keys** を採用する（未作成・未適用）。全列NOT NULL。

| 列 | 型・制約 |
|---|---|
| key_reference | text PK。canonical lowercase hyphenated UUID文字列。意味を埋め込まない |
| clinic_id | bigint REFERENCES clinics(id) ON DELETE RESTRICT |
| principal_account_id | bigint REFERENCES accounts(id) ON DELETE RESTRICT |
| medical_record_id | bigint |
| tab_key | text、CHECK btrim(tab_key)<>'' |
| key_version | bigint、CHECK >0 |
| wrapped_dek | bytea、CHECK octet_length(wrapped_dek)>0。平文鍵本体を保存しない |
| wrapping_format | text、CHECK ='kms-envelope-v1' |
| kek_reference | text、CHECK btrim(kek_reference)<>''。server provider内の鍵識別子でcredentialではない |
| created_at | timestamptz DEFAULT clock_timestamp() |

追加制約:

- FK `(medical_record_id,clinic_id) REFERENCES medical_records(id,clinic_id) ON DELETE RESTRICT`。
- UNIQUE `(clinic_id,principal_account_id,medical_record_id,tab_key,key_version)`。
- UNIQUE `(key_reference,clinic_id,principal_account_id,medical_record_id,tab_key)`。
- RLSはdraft同様clinic access AND principal_account_id=trusted app.current_account_id。generic draft read APIからwrapped_dek／kek_referenceを返さず、brokerだけ内部取得する。

既存 `staff_record_drafts` へ `(key_reference,clinic_id,principal_account_id,medical_record_id,tab_key)` からkey registryの同uniqueへのRESTRICT複合FKを、下記の **NOT VALID段階導入** で追加する。既存ciphertext／encryption_format／key_reference／hash／input_generation／base_owner_versionsは再利用し、epoch・鍵本体・device ID列を増やさない。

**A metadata anchorとB key registryはlock責務が異なる。** Aに業務record FKを付けないR2.1は、B registryの上記record複合FKを削除する指示ではない。追加DDLは2表とdraft-key複合FK。既存表の重複CREATEはしない。以下は設計であり、実データmigration／削除やSQL実行の許可を意味しない。

### 旧draftの非破壊・段階導入

1. schema ownerが既存scope／key_referenceを読取棚卸しして対応を分類する。実在する正しいwrapped keyとscopeを証明できるものだけ、別途認可された導入工程でregistryへ正規登録する。未知鍵の推測backfill、架空wrapped_dek、ダミー親行、別鍵への付替え、勝手な再暗号化を行わない。
2. draft-key複合FKを `ADD CONSTRAINT ... FOREIGN KEY (...) REFERENCES ... ON DELETE RESTRICT NOT VALID` で追加する設計とする。旧行の一括検証を省略しつつ、新INSERTおよびFK検査対象更新には制約を適用する。NOT VALIDは制約なしを意味しない。[PostgreSQL 18のADD／VALIDATE CONSTRAINT契約](https://www.postgresql.org/docs/18/sql-altertable.html#SQL-ALTERTABLE-DESC-ADD-TABLE-CONSTRAINT)に従う。
3. 未解決旧draftは既存表内の原行を変更せず温存し、registry不一致を読取時source-unresolvedとする**論理隔離**に限定する。原ciphertext／元metadataを維持し、復元可能と表示しない。PUTによる置換・世代更新・自動修復・自動削除を拒否し、一意scopeを空き扱いして新規draftで上書きしない。新しい隔離表への移送を必須にしない。
4. 対応が証明された旧行は正規registry関係が成立した後に利用可能とする。未解決scopeだけfail closedとし、新規scopeは正規registry＋新規書込みを拘束するFKの下で利用可能。既存行の関係をすべて証明した時だけVALIDATE CONSTRAINTへ進み、未解決が残る間は検証完了と報告しない。新規保護の導入と全体FK検証の完了を別々に報告する。解決不能な行の破棄等は別の本人判断であり、本設計では要求しない。

### provisioningとprovider

初期scope keyはkey_version=1。候補DEK生成／wrap→同providerで候補wrapped DEKをunwrapして候補DEKとの一致を確認→確認済み候補だけを短いDB TXでregistryへINSERT／commit→現在認可と対象keyの利用可否を再検査して応答、の順とする。確認失敗なら候補registry行を公開・commitしない。provider呼出をC06のDB-only participantや長い業務lock内へ持ち込まない。

同scope／versionの同時provisionはUNIQUE(scope,version)で勝者を決める。敗者は自分の候補DEKを応答せず、commit済み勝者key_referenceを読み直し、その勝者wrapped DEKのUnwrapが成功してからだけ利用成功を返す。Provision結果のkey_reference／key_versionはこの勝者keyを指し、初回PUTも既存draft行から逆引きせずこの準備済み参照を使う。DB行の存在や一時点の確認を、現在・将来のprovider可用性の保証にしない。

後からproviderが停止／失効した場合は当該利用をfail closedとし、既存ciphertext／registryを自動削除・置換せず同key_referenceで再照合可能にする。DB commit成否不明もscope／version／key_referenceの読取照合で解決し、新鍵で既存scopeを上書きしない。未使用wrapped candidateを失敗DBに残さず、provider処理の不確実性を照合する。pending／failed状態列、rotation、実鍵作成を追加する設計変更には広げない。

provider interfaceは `Provision(scope,key_reference,key_version)` / `Unwrap(scope,key_reference)` と利用拒否／障害の分類を持ちfakeで契約可能。secret取得は既存安全設定／正式接続工程へ渡す。key_version更新・rotation・失効／削除管理画面は今回作らず、将来の過去key削除方針も決めない。registryへinput_generationを重複保存しない。

## 9. B: 暗号形式・AAD・鍵利用境界

採用形式は `encryption_format='aes-256-gcm-v1'`。DEKは256-bit、暗号化ごと新しい96-bit nonce、128-bit tag、同鍵nonce再利用禁止。ciphertext envelopeは `nonce(12 bytes) || ciphertext || tag(16 bytes)`。ciphertext_sha256はenvelope全体のdigestであり、認可／AEADの代用ではない。wireはbase64でencodingを固定する。

AAD bytesは以下の順序の **JSON string arrayをUTF-8** で符号化する。数値もcanonical decimal文字列、空白なし、UUID lowercaseとする。

```text
["staff-record-draft","1",key_reference,key_version,principal_account_id,clinic_id,medical_record_id,tab_key,input_generation]
```

### AAD v1のcanonical string serialization

全要素に同じstring serialization規則を適用する。各要素はUnicode scalar value列とし、unpaired surrogateとDB textに保存不能なU+0000はvalidationで拒否する。Unicode正規化（NFC／NFD等）、大小文字変換、trimを行わず、serverがscopeとして照合したtab_keyのscalar列をそのまま使う。見た目が同じでも異なるscalar列は別tab_keyであり、browserが正規化して同じscopeとみなしてはならない。既存 `btrim(tab_key)<>''` 条件も維持し、ASCII限定へ狭めない。

JSON stringは前後をU+0022で囲む。U+0022は `\"`、U+005Cは `\\`、U+0008／0009／000A／000C／000Dはそれぞれ `\b`／`\t`／`\n`／`\f`／`\r` とする。それ以外のU+0001〜U+001Fは小文字hexの6文字 `\u00xx`、その他のscalarはescapeせずそのままUTF-8へ符号化する。`/`、HTML文字、U+2028／U+2029、非ASCIIを任意にescapeしない。配列要素の区切りは`,`、前後は`[`／`]`だけで、BOM・空白・改行を加えない。数値文字列・UUIDのcanonical規則と要素順は上記のまま。browserとserverはこの同一serializerを用い、既定JSON encoderの任意escapeに依存しない。[JSON stringの構文](https://www.rfc-editor.org/rfc/rfc8259.html#section-7)のうち、この表現をAAD v1の唯一のbyte表現として選定する。

復号時のAADはGETの照合済みkey／version／generation／scopeから再構成する。新規暗号化時も準備済みregistry scopeとserver規則の次generationを使う。clientから渡された任意のAAD bytesを認可・scopeの根拠にせず、異なるserializationを順に試す復号fallbackを設けない。

記号部分もすべてstring要素としてserializeする。AADへauth token／session証明を入れない。元epoch等は暗号化payload内の履歴情報に置き、payload schema versionも暗号化本文内で検証する。registry scope／body scope／AAD不一致は拒否。body accountを認可根拠にせず、復号AADは信頼済registry scopeから組み立てる。fresh load時は§10のGETが照合済みkey／世代／scope metadataを返す。暗号化時のinput_generationは§10のserver規則と同じ次版をbrowserが計算し、serverは一致を検査する。serverが暗号文を別世代へ付替えたり書換えたりしない。他account／recordへのmetadata改変移植はtag検証で失敗する。

IndexedDBは暗号文envelopeと最小の非本文routing情報だけ。localStorage／sessionStorage／Service Worker cache／ログ／analyticsに本文・DEK・auth／edit証明を保存しない。本人draft GETの全応答も `Cache-Control: no-store` とし、共有browser／intermediary／Service Worker HTTP cacheへ保存・別accountで再利用しない。明示設計のIndexedDB ciphertext journalはHTTP cacheではなく、本人scopeで管理する暗号文永続化として保持する。key-accessはno-storeでmemory-only鍵利用を渡す境界とし、wrapped_dek／KEKはbrowserへ送らない。browserはprovider由来DEKをimport後memory keyにしraw buffersを可能な範囲で破棄する。ログ／trace／キャッシュ／Service Worker interceptionからの除外を後続受入へ含める。

ブラウザ暗号APIの根拠: [MDN AesGcmParams](https://developer.mozilla.org/en-US/docs/Web/API/AesGcmParams)、[SubtleCrypto.encrypt](https://developer.mozilla.org/en-US/docs/Web/API/SubtleCrypto/encrypt)。デプロイ時のbrowser-facing API／key-brokerのHTTPS必須・HTTP入口拒否は§3に従い、IV一意、AAD照合、tagを用いる。JS heapの完全消去や侵害端末からの完全保護は保証しない。

## 10. B: lifecycle・回収・候補API

最初のオンライン編集開始時に対象key_referenceと暗号能力を準備する。取得失敗で平文永続化せず「再読み込み復元が未確保」を表示し、保証を黙って外さない。

- DEKはmemoryのみ。reload後は本人再認証／現認可後にbrokerから再取得する。logoutはmemory keyとUI／plaintext cacheを無効化し、ciphertextを保持して自動upload／retryを停止する。
- offlineでは既決どおり新編集／送信を停止する。切断直前の未送達入力は取得済みmemory DEKでlocal ciphertextとして保持できる。reload後offlineは鍵を回収できないため「復元待ち」とし、消去や平文fallbackをしない。online再認可後に回復する。
- 鍵不在／破損／対応不能、registry不整合、provider停止／失効は「復元できない／照合待ち」。空draft成功にせず、旧ciphertext／registryを自動削除・置換しない。同key_referenceで再照合し、自動鍵再作成で復号できるふりをしない。§8の未解決旧draftはsource-unresolvedとして原行を温存し、PUTによる置換・世代更新も拒否する。
- restoreは元scopeと現在grant確認→未解決receipt照会→編集権取得→base／current／draft三者比較→明示field選択。現在の新epochとowner版で通常writerへ渡し、新履歴を残す。
- base_owner_versionsには元owner版、元input_generation、必要な元epoch参照を保持する。元epochは履歴であって現在保存権に再利用しない。

**旧epochの整合した解釈:** 旧epochを現在の編集／保存権として受理せず、失権中の旧証明による復号能力再取得を拒否する。本人が現在認可・編集権を再取得し、元scope／未解決結果／現在版を照合した後は、旧draftを比較用に回収し明示選択できる。旧入力を永久復旧禁止にする解釈は採用しない。

### Bのtrusted clinicと共通scope

BのGET／PUT／key-accessでいうclinicはURLからの独立推定値ではなく、server-side認証・現在医院アクセス解決済みcontextのclinic_idである。入口でrequestのtrusted clinic_idを一度確定し、認証主体から解決したprincipal_account_idとともにrequest内で共用する。record_idはURLをparseしてこのclinic内で存在・現在grantを照合する。tab_keyはURLの通常の一回のdecode後に§9のscope文字列規則で検査する。別helperがdefault clinicやclient希望clinicを再解決して差し替えてはならない。

この同じclinic_idをregistry／draft lookup、TX-localのclinic／account contextとRLS、provider scope、GET復号metadata、AAD構築・照合へ渡す。PUT bodyやqueryのclinic／account指定は認可根拠として無視するのではなくvalidationで拒否し、URL recordが当該clinicに属さなければscope-not-foundとする。provider前後の現在認可再検査も同じclinicを対象とし、context失効／変更時は失敗させ別clinicへ暗黙切替しない。GETのclinic_idも単独の認可証明ではなく、次requestでは再び現在server contextから検査する。body／AAD scopeはこの正規値との一致検査に使い、clientに認可主体を選ばせない。

| 採用API境界（未実装） | 契約 |
|---|---|
| GET `/medical-records/{id}/personal-drafts/{tab}` | 自本人scopeのcurrent envelopeとserver照合済み復号metadata。保存ありは200のdraft object、未作成は200 draft:null、不整合は409 ErrorResponse。下記wireを使い全応答 `Cache-Control: no-store` |
| PUT 同route | operation_key、expected_generation、new generation、準備済みkey_reference／key_version、base_owner_versions、envelope、未解決operation refs。初回0→1、以後expected+1の原子CASと下記鍵参照照合を必須とし、clientが認可主体を指定しない |
| POST 同route `/key-access` | 現在auth／account／scope／grant、held holder／session／epochとkey_reference／versionを検査するreadのbroker処理。下記の平坦request／responseとErrorResponseを使い、全応答no-store。C06保存receiptは発行しない |

### GET: fresh loadの復号metadataとcache境界

GETは正規registryと保存draftをserverが照合した `key_reference, key_version, input_generation, encryption_format, ciphertext_sha256, base_owner_versions` と、AAD用正規scopeの `principal_account_id, clinic_id, medical_record_id, tab_key` をenvelopeとともに返す。これらは秘密鍵ではなく、client指定値を本人認可の根拠にしない。wrapped_dek／KEK／資格情報は返さない。metadata不整合はsource-unresolvedとし、復号可能な正常draftに見せない。未解決operation refsの回収契約を維持し、metadataを成功receiptの証明へ代用しない。

本当にdraft未作成なら `draft=null` を返し、input_generation=0の偽DB行を作らない。未解決旧draftをabsentへ縮退させず、元scopeが空いていると扱わない。GETの全応答はno-storeとし、logout時のplaintext UI／memory key破棄と併せて別accountへのHTTP応答再利用を防ぐ。これは既決のIndexedDB ciphertext保持を削除する指示ではない。

GETは成功・失敗を含め `Content-Type: application/json`、`Cache-Control: no-store` とする。認証・現在view grant・clinic／record／account scope検査後に次のwireを区別する。

| GET結果 | HTTP | JSON bodyと成立条件 |
|---|---|---|
| 正常な保存draftあり | 200 | トップレベル `{ "draft": { ... } }`。draft必須fieldは `envelope, key_reference, key_version, input_generation, encryption_format, ciphertext_sha256, base_owner_versions, principal_account_id, clinic_id, medical_record_id, tab_key`。envelope全体のhashとmetadata／registry照合が成功した場合だけ返す |
| 本当に未作成 | 200 | 正確に `{ "draft": null }`。0世代の偽draftや空envelopeを返さない |
| 保存行があるが旧鍵参照／registry／scope／形式／世代／hash等の照合失敗 | 409 | 正確に `{ "error": "source-unresolved" }`。draft:nullも正常envelope／復号metadataも含めない。原行を変更・削除せず、clientは作成／上書き／自動復元を停止 |

envelopeは§9のbase64暗号文形式、key_version／input_generation／IDはcanonical decimal文字列、key_referenceは既存canonical UUIDとし、他fieldの意味は既存§9／§10を維持する。未解決operation refsは既存回収経路で照会し、GET metadataを成功receiptの代用としない。

入口エラーも平坦な `ErrorResponse {error:string}` とし、400 `{ "error": "validation" }`、401 `{ "error": "unauthenticated" }`、現在grant不足403 `{ "error": "not-authorized" }`、URL対象recordが認証済みclinic内に見つからない404 `{ "error": "scope-not-found" }`、読取基盤障害503 `{ "error": "result-unavailable" }` とする。異scopeの存在情報を返さない。GETはprovider unwrapを実行せず、provider障害の判定をkey-accessへ分離する。

`draft` は既存draft=nullを具体化する資源fieldであり、C06結果を包むreceipt／data／operationラッパーではない。source-unresolved等はErrorResponseの値であってC06 status新設ではなく、保存PUT／結果照会のResultResponseを変更しない。

### PUT: 初回・単調generation CASと鍵選定

expected_generationはcanonical非負整数文字列とし、`"0"` だけをdraft未作成のcreate preconditionに使う。DB input_generationは従来どおり正整数で初回1。初回0→1はscope一意制約付きINSERTで勝者1件とし、敗者は409競合で勝者内容を自動上書きしない。既存scopeをexpected=0で置換しない。既存draftはexpectedが現在版に一致する場合だけ原子更新し、次版はserver規則でexpected+1。DB bigint上限超過・任意ジャンプ・後退を拒否する。

browserも暗号化前に同じ次版を計算してAADのinput_generationに使う。requestのnew generationは主張値でありserverが算出値との一致を検査し、暗号文を別世代へ付替えない。operation_keyの成功replayは保存receiptを返して世代を再増分せず、応答不明は同key結果照会を先行する。新key／次世代で自動再送しない。C06 fingerprint／precondition分類は既存規則を維持し、AAD／ciphertextを含む内容が変わる操作を「CAS-onlyだから同じ内容」と扱わない。新暗号文・内容への切替は旧結果を確定してから行う。

browserはB0のkey provisioning結果で準備済みのkey_reference／key_versionをPUTで明示する。serverは認証から解決したaccountと上記trusted clinic、照合済みURL record、検査済みURL tabでregistryを読み、request参照と当該scope／versionが一致することを検査する。初回も既存draft行からの逆引きを前提としない。body／queryのclinic／account指定はvalidationで拒否し、clientによる新registry登録は不可。初期key_version=1、同scopeの勝者keyに固定し、既存draftのkey付替え・rotationをPUTに混ぜない。

既存keyと異なる参照、不明参照、scope不一致は拒否し、新鍵生成やNULL／ダミー補完をしない。key_reference／key_versionはAADの同値に対応し、wrapped DEK／raw DEKをPUT本文へ含めない。参照一致を現在auth／grant／編集権の代用にせず、旧epoch回収の既決境界も維持する。

key provisioning APIとregistry／provider契約の反映はB0の小PRへ分ける。保存先は上記registryへ固定済みであり、既存key_referenceをDEK本体と見立てたりciphertextへ復号鍵を同梱したりしない。

### POST key-access: wireと現在認可

POST `/medical-records/{id}/personal-drafts/{tab}/key-access` は既存Cookie認証とCSRF境界を継承する。`application/json` の平坦request bodyは `key_reference, key_version, edit_session_ref, edit_epoch`。key_referenceはcanonical UUID、key_version／edit_epochは正のcanonical decimal文字列、edit_session_refは既存編集session UUID。account／clinic／record／tabをbodyで選択しない。現在認証主体・現在grant・本人scope・current held holder／session／epochを検査し、同scopeのregistry key_reference／version一致を確認してからUnwrapする。旧draft内のepoch／sessionを現在権として使わず、本人が現在編集権を再取得した場合は現在session／epochで過去draftの同scope keyを取得できる。

成功はHTTP 200、`Content-Type: application/json`、bodyは以下の平坦形とする。

```json
{"key_reference":"<canonical UUID>","key_version":"1","encryption_format":"aes-256-gcm-v1","dek":"<base64>"}
```

dekは正確に32 bytesを [RFC 4648標準alphabet](https://www.rfc-editor.org/rfc/rfc4648.html#section-4) のpadding付きbase64で符号化する（改行なし、base64urlではない）。browser memoryへ渡す一回のraw keyであり、鍵参照・bearer・永続tokenではない。browserはscope／key／version一致と長さを検査し、AES-GCMのnon-extractable CryptoKeyとしてmemoryへimportし、raw文字列／bufferを可能な範囲で破棄する。wrapped_dek／KEK／provider credentials／内部エラー詳細を返さない。

全応答は `Cache-Control: no-store`。応答body、DEK、cookie、session証明をログ／trace／analytics／HTTP cache／Service Worker cacheへ保存せず、browser永続storageへDEK／CryptoKeyを置かない。session限定とは現在auth／編集sessionの有効期間だけmemory利用し、logout／既知の失権／認可取消で利用停止・memory参照破棄する意味であり、配布済みDEKを暗号学的に遠隔失効できる意味ではない。寿命や保持期限の新業務ルールを追加しない。

endpoint固有の失敗はすべて `Content-Type: application/json` の平坦ErrorResponseとする。

| HTTP | JSON body | 意味 |
|---|---|---|
| 400 | `{ "error": "validation" }` | JSON／必須field／型／canonical形式不正 |
| 401 | `{ "error": "unauthenticated" }` | 有効な認証sessionなし |
| 403 | `{ "error": "not-authorized" }` | CSRFまたは現在grant不足。内部理由を出さない |
| 404 | `{ "error": "scope-not-found" }` | 現scopeで対象record／key_reference／key_versionを解決できない。別scope keyの存在を開示しない |
| 409 | `{ "error": "edit-right-lost" }` | 現holder／session／epoch不一致。旧証明での取得を拒否 |
| 409 | `{ "error": "source-unresolved" }` | 対象本人scopeの保存draft行は存在するが、正規鍵関係やmetadataが不整合。absent扱いで新鍵を生成しない |
| 503 | `{ "error": "result-unavailable" }` | provider未設定／停止／失効／unwrap不能または基盤障害。秘密やprovider詳細を出さずfail closed |

認証→現在grant／record scope→現在編集権→保存draft整合→要求keyのscope／version照合→Unwrapの順で検査する。本人scopeの保存draft行があるのに正規registry解決に失敗する場合はsource-unresolved、保存draftがない初回利用で要求keyが同scopeの準備済みregistryにない場合はscope-not-found。失敗応答にdekを含めず、エラー時に鍵／registry／draftを生成・再作成・置換・削除しない。新規provisionはB0の別契約のまま。

provider呼出をC06 DB-only participantや長い業務lockに入れない。Unwrap後、鍵応答直前にも同じtrusted clinicの現在auth／grant／holder／session／epochとkeyのscope対応を再確認し、失効が判明した結果を200にしない。鍵アクセスはreadのbroker処理でありC06 operation_key、成功receipt、世代増分を発行しない。上表のerror文字列は既存ErrorResponseの値であり、C06 status enumを拡張しない。

### 採用理由と保護限界

| 比較した方式 | 判断 |
|---|---|
| serverだけで暗号化／復号して保存 | 未送達→reloadの本人入力回収を満たさず、server部分PRにはできてもB全体完成にはならない |
| IndexedDBへnon-extractable CryptoKeyとciphertextを同時保存 | 共有originの別staff scriptからの鍵利用を防げず、採用しない |
| memory DEK＋server再認可broker＋local ciphertext | 採用。online reload回収／共有端末logout後の本人分離へ接続 |
| passphrase／WebAuthn等の本人端末鍵 | 利用者操作・復旧責任・対応端末条件を増やすため、勝手に必須化しない |

保護対象はlogout後の通常の別staffセッションが永続draftを開けないこと。memoryへ渡した鍵／平文を、権限取消の瞬間に敵対的scriptや端末管理者から暗号学的に回収できない。XSS・侵害browser・OS processからの完全保護を主張しない。失権を知った時点でUI／key利用を止め、serverは取得／保存／再取得を現在認可で検査する。通信断でも新編集停止。端末attestationやoffline新規編集の追加は別設計。

保持期限・自動破棄・端末間共有は本設計で追加しない。これらや復旧不能時の業務データ破棄を新たに導入する場合だけ別途本人判断が必要。既決logout保持を再質問せず、今回の設計／コード着手を期間回答待ちにしない。

## 11. ownerへの受渡し順と後続受入（全件NOT_RUN）

1. schema owner: A新規metadata anchor、B新規key registry＋draft複合FK。§8の読取棚卸し→証明済み参照の正規登録→NOT VALID FKによる新規保護→全既存参照の証明後にVALIDATE、の段階を分ける。未解決scopeは原行温存・停止とし、全体検証未完了を明示する。migration／model／ERDは446/466等の専有ownerへ集約し並行上書きしない。既存表を重複CREATEしない。
2. 489 shared owner: 新規binding時のatomic ClaimTarget metadata callbackだけを追加。既存receipt wire／status／consumer契約は維持。
3. 494 A: actor／session／fence、inquiry baseline／CAS、metadata resolver、flat DTO／HTTP mapping、失権入力保持。A UI／APIと495-Aは1／2のruntime完成待ちにしない。
4. 494 B0: envelope／AAD v1のcanonical serializer／key-reference／identity／状態API／schema・provider契約とfake fixture。GETのpresent／absent／source-unresolvedと各HTTP／JSON／no-store、key-accessの平坦request／response／error、PUTの初回／単調CASと準備済み鍵参照をDTOへ固定。実鍵なし。
5. B1: server ciphertext repository、account TX-local／RLS、generation CAS、current auth、operation lookupとbroker error mapping。Bの全経路で同じtrusted clinicを用い、provider未設定runtimeはfail closed。
6. B2: browser ciphertext journal、memory key provider、logout／reload／切断、三者比較選択。公開契約モックで独立PR可能。canonical AAD再構成、key-accessのmemory importとerror分岐を同wireへ接続。
7. B3: 明示承認後のlive provider接続・運用設定。未実行なら未接続と報告。

後段確認計画: 同staff2sessionの取得勝者1つ、take-overとwriteの両順、releaseとunknown、old client、別account／同名再作成、別clinic／record／tab、現在grant失効、corrupt／鍵不在、quota、未送達reload、logout後別staff、同key別payload／別record、anchor欠落・claim rollback・record変更／論理削除、baseline bootstrap rollback・全inquiry writerの版参加、draft generation CAS、旧成功receiptの現在認可、旧draftの再取得後比較回収、ログ／cache等の鍵非保存。すべて記述のみで **NOT_RUN**。

Bの後段fixtureもNOT_RUN: 新規FK違反拒否と未解決旧scopeだけの停止／原行温存、未解決残存時のVALIDATE未完了表示、候補unwrap確認失敗時の非公開、provision競合敗者の勝者key再取得・Unwrap失敗、provider後日停止／commit不明の非破壊再照合、fresh GET metadataによるAAD再構成・不整合とabsentの区別、全GET応答no-storeとlogout別account分離、初回PUT競合の勝者1件・0→1・以後+1・上限／後退／ジャンプ拒否、成功replay非増分・応答不明の照会先行、初回鍵選定・異scope／異key拒否。記述の追加を実装／DDL適用／実鍵／試験実行の完了と扱わない。

wireの後段fixture（NOT_RUN）: AADのquote／backslash／control、非ASCII、結合文字と合成文字、U+2028／U+2029のbyte一致／区別とunpaired surrogate／U+0000拒否。GETの200 present／200 absent／409 unresolvedの正確な形、409をabsentへ読み替えないこと、全応答no-store。key-accessの32-byte DEK標準base64・memory import、各HTTP／error分岐、旧session／epoch拒否・再取得後回収、provider前後の認可取消と秘密非出力。GET／PUT／broker／RLS／AADでtrusted clinicが一致し、body／query clinic拒否と別clinicへの暗黙切替がないこと。追加のfixtureを実行済みとは扱わない。

## 12. 根拠・実施状態

- [Plane EMR-494](https://app.plane.so/baritechllc/browse/EMR-494/): R2.1採用本文が旧「鍵方式未選定」「古いepochは復号不可」の総論に優先。限定PR／試験時期コメント（12:30 UTC）は受入計画として保持。
- [Plane EMR-495](https://app.plane.so/baritechllc/browse/EMR-495/): A診断consumer、B終了event、C目的別訂正の分割。494 B完成をAの着手gateにしない。
- [storage_clinical.go](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/model/storage_clinical.go)、[編集権ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/architecture/erd-detail/record_edit_rights.md)、[draft ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/architecture/erd-detail/staff_record_drafts.md)、[dependency版ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/architecture/erd-detail/medical_record_dependency_versions.md)。
- [本人account解決](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/auth/http_session_me.go#L19-L103)、[draft RLS](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/migrations/001_init.sql#L10512)、[既存route](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/medicalrecord/routes_records.go)、[問診repository](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/medicalrecord/inquiry_repository.go)、[C06 ResultResponse](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/operation/dto.go)。

この公開PRはdocsのみ。API公開SHA・runtime接続・schema実装／DB適用・実機受入は未完了。試験、Docker、隔離環境、実鍵作成、KMS／資格情報／権限設定を実行していない。runtime検証はSKIP、fixtureはNOT_RUNであり、設計採用を機能完成・セキュリティ設定変更の承認・試験PASSと扱わない。既存必須CIは無効化しない。
