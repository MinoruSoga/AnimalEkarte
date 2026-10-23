# EMR-547 旧会計・見積の根拠保全と履歴閲覧：保存・read選定契約

| 管理項目 | 記載 |
|---|---|
| 文書ID・種別 | design.accounting.legacy-finance-evidence／移行保存・閲覧の技術契約 |
| 決定状態 | 6入れ子型に加え、consumer検証記録/verification_refの保存schema・供給binding、delivery-local参照からstable candidate/outcome/projection/artifact UUIDへの耐久対応、chunk件数/bytes・field長・payload上限も未閉鎖。業務保全原則を継承。2026-10-23に共通構造・保存tableの役割・関係制約（chunk/field/payload上限は未選定）・readonly API・合成fixture条件を選定。6入れ子型と実データprofile/bindingは未閉鎖 |
| 実装状態 | 保存/reader/API/UI未実装。実DDL・schema/APIファイル・供給/実装受領は未完 |
| 実機検証状態 | tests/build/codegen/DB/import/Docker/browser NOT_RUN |
| 文書確認日 | 2026-10-23（静的反映。実行受入ではない） |

[EMR-547](https://app.plane.so/baritechllc/browse/EMR-547/)の7ACに対応する。初回の設計基準main `42a3890a22d645857ad8e27c3b2d7bbfc9c2e038`、初回docs作成base `621d58a0834aa6984ee3c82401e3478efef45d2e`は受領履歴として保持する。今回の返却設計照合基準/作成baseはmain `cff4debe2e4620584e3afc4faaabea370afc7d37`。

Library `libfile_222be047bc548191805d491f522e4b55` の `EMR-547-storage-read-contract-response.md`（36,790 bytes、SHA-256 `4cb55a764532fa238315af01acf1739b290a4a88004c56b054e6b64f6deda158`）をconsumer側で取得・可読/hash確認し本正本へ統合した。返却原稿の複製文書は作らない。後発修正入力はLibrary `libfile_4a4be7a4b2f081919ac431888afefea9` の `review-fixes-20261023.md`（23,252 bytes、SHA-256 `7faeb1461a22ba3df586eaf92d3774053cbf0f7d88578a475b5ed4e19c443d74`）。初回hashは歴史的受領証跡であり、本修正後本文のhashではない。以下の「確定」は設計選定で、既存mainにDDL/API/producerが存在する、検証済みという意味ではない。

**6入れ子型に加え、consumer検証記録/verification_refの保存schema・供給binding、delivery-local参照からstable candidate/outcome/projection/artifact UUIDへの耐久対応、chunk件数/bytes・field長・payload上限も未閉鎖。実装へ渡せるのは選定済みの共通構造・保存tableの役割・関係制約（chunk/field/payload上限は未選定）・readonly API・合成fixture条件の範囲。6入れ子型、実データのsource profile、manifest結合、集約元全行の列挙は未受領で、sidecarの実データ生成・consumer正式受入はまだ開始できない。** 正式移行HOLD、19 CSV維持、通常finalize/allocate/診察終了eventを生成しない境界を保つ。21表時代のreceiptを現行19表の適合証拠に転用しない。

producer `a38e71f904460ba3d33194e817bafcbb78886d0c` の静的mapping報告は受領済み。本consumerによるold_db実コード/実DB/既出bundle/実artifact manifestとcode/digestの独立検証ではない。同じ6項目の再調査を依頼せず、既存報告から設計へ転記できる範囲と未受領根拠を分ける。

今回docsのみ。製品コード/API YAML/DDL/model/producer SQL/実データ/設定/Planeは変更せず、tests/build/codegen/DB/import/Docker/browser NOT_RUN。実PHI/原値行/秘密/ローカル実pathは転載しない。

<a id="legacy-evidence-purpose-and-ownership"></a>
## 根拠・業務目的と単一owner


業務判断窓口はMinoru Soga。旧請求の意味不明状態と旧見積の割引原値を失わず、製品内で原値/変換/不明を区別して履歴参照する。確認のために通常請求を捏造したり二重台帳へ転記したりする工程を除く。計画指標は7ACと保存→read/UI・担当・未受領契約が対応すること、未説明差分0。実業務の時間/操作数改善は未測定、保存/画面受入PASSではない。

| 参照 | 継承する境界 |
|---|---|
| [221：証拠範囲](../../../work/emr/emr-221-import-contract-design/DESIGN.md#source-evidence-scope)・[契約変更案](../../../work/emr/emr-221-import-contract-design/DESIGN.md#proposed-import-contract-boundaries)・[正式判定/耐久receipt](../../../work/emr/emr-221-import-contract-design/DESIGN.md#formal-eligibility-and-durable-receipt) | 原本identity/digestを固定、完全性と独立適格性を分ける提案、別revision原本保持、耐久receipt。同案を実装済み/正式import承認へ昇格しない |
| [formal安全境界](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md#formal-import-safety-boundaries)・[再開gate](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md#formal-import-producer-resume-gate)・[作業責務/実行binding補足](../../../ops/deploy/seeds/CLINIC_CSV_IMPORT.md#producer-work-ownership-and-runtime-binding) | 既存formal v1/profileのstrict検査、MIG-41がproducer作業項目を所有する既存境界、quarantine/PHI保護、fresh-only/band保護。未知版fallback/自己申告modeを追加しない |
| [会計正本](../../../spec/domains/accounting/lifecycle.md)・[会計技術仕様](technical-spec.md) | 通常請求確定・実受領/配賦・返金・原子receiptは正規根拠とownerに従い、証拠参照で代替しない |
| [見積正本](../../../spec/domains/estimates/business-rules.md)・[507/508契約](../../../work/emr/emr-507-508-selected-wire-contract-20261021.md) | 提示/計算revisionの意味が定まった型を維持し、未知旧割引値を無定義JSONへ詰めない |

| 作業項目owner | 本件の必要subset・除外 |
|---|---|
| 547 | 製品側immutable evidence保存、限定importer接続要求、read-only一覧/詳細UI。既存票/実装sessionへ無断追加しない |
| 349 | 配置・bundle受領・preflight/importの唯一の作業項目owner。547が別import実行経路を作らず、349という票を実行主体に読み替えない |
| 221 | 版付きevidence-only受領契約、identity/digests/独立適格性/receiptとの整合 |
| 446/466 | 本書§4の保存構造を現在の001へ統合し、clinic複合FK/unique/RLS/immutable/index/model/ERDを受領。実DDL/既存clinic型/role/制約名の統合は未完 |
| MIG-41 | producer作業項目ownerとして候補生成・source mapping/原本bytes/pins/receipt引渡しを追跡。既存producer責務はold_db HAC-CSV-1/HAC-PIPELINE-1にあるが、新evidence-onlyへの対応は候補で、現行producer main `a38e71f904460ba3d33194e817bafcbb78886d0c`と静的mapping設計入力は報告として受領済み。新sidecarのclosed mapping/実artifact binding/実行receiptは未受領。MIG-41という票自体が生成処理を実行する意味ではない |
| 361 | 同一runのN1画面/復旧実機受入。製品保存/read API/UIの実装ownerではない |
| 505/508 | 正式会計/見積writer/reader・未請求候補/集計からの分離、lineageと許可されたリンクの境界 |

既存runbookに記録されたold_db HAC-CSV-1/HAC-PIPELINE-1のproducer責務と、AnimalEkarteの現行csvimportの配置/preflight/import経路は既存範囲の事実として区別する。これらを新しいEMR-221案/547 evidence-only保存へ接続する対応は候補であり、現行producer SHAと静的mapping設計入力は[報告として受領済み](#producer-static-report-and-preservation-response)。新sidecarのclosed mapping/実artifact binding/実行receipt/producer-consumer接続は未受領。既存CLINIC_CSV_IMPORTのscript/consumer経路が実在することだけでは、新契約への接続済み・実行済みを証明しない。実際の操作者・対象環境・個別実行許可は同runで確認が必要で、今回未確認。作業項目ownerやdocs上の責務を実行済み証跡/操作者名の代わりにしない。

<a id="legacy-evidence-storage-roles"></a>
## 1. 保存するものと通常業務からの分離

### 確定

通常の19 CSVとは別の版付き根拠保全データとします。19 CSVを旧形式へ戻さず、`payments` / `payment_splits` は復活させません。通常の`billings`や`estimate_calculation_revisions`の任意JSONへ暫定投入もしません。

保全データは次の三層です。

1. **原行根拠**：取得された元の列・値・NULL等の状態、出典、元キー、snapshot。取得した表現を保持し、業務上の意味は別管理します。
2. **変換projection**：rawからどの規則でどの値へ変換したか。集約後の請求ヘッダ、丸め後の割引額、生成値・補完値を含みます。
3. **出力・除外結果**：projectionが出力対象になったか、WHERE等で除外されたか、紐付けが降格したか、出力artifactとの関係。

原始TRI明細と集約後請求ヘッダは別の型・IDとします。集約ヘッダを「元DBに実在した1原行」と扱わず、入力原行をN件参照する1 projectionとして保持します。

根拠保存だけでは、通常会計編集、請求確定、入金配賦、再請求、未請求候補化、売上計上、通常見積の再計算を可能にしません。`finalize` / `allocate` / 診察終了eventは生成しません。通常の一覧・集計へUNIONしません。

### 依存先

- EMR-221：別sidecarの正式な契約名・版の登録、既存19 CSVの契約との明示的な分離。
- EMR-505 / 508：通常writer、未請求候補、会計・見積・売上readerへ混入しない実装境界の照合。
- 本選定では契約の論理名を `legacy-finance-evidence`、版を `1` としますが、既存manifestで利用可能な登録済みprofileとは扱いません。登録前の未知profileとして自動受入してはいけません。


<a id="legacy-evidence-formal-import-separation"></a>
### 既存formal契約の安全境界

既存v1/profileのstrict完全性/税/親子/FK/payment/clinic/target判定、fresh-only/ID band/既存prefix/lockを緩和しない。clinic/provenance/原値整合が成立し意味だけ不明なものの保全と、provenance自体不明/bytes不一致/医院不明のquarantineを分ける。別名・版付き `legacy-finance-evidence` / `1` の選定は、既存manifestへのprofile登録やmode自己推定を許可しない。通常台帳/証拠/除外/receiptの母集団を合算しない。

<a id="legacy-evidence-identity-retry-and-atomicity"></a>
## 2. identityと再送・原値競合

### 確定

identityを次の三軸へ分離します。

| 軸 | 構成 | 含めないもの |
|---|---|---|
| 元entity | clinic、登録済みsource system、系譜、元table、登録済み複合キーの列と未加工値 | output ID、run、build、契約版、mapping版、raw内容digest |
| 原行根拠版 | 元entity ＋ consumerが検証済みとして登録したsnapshot | 配送や変換の版 |
| 配送・変換版 | producer run/build、契約版、mapping/transform/capture profile、入出力artifactとdigest、入力根拠集合 | 元entityの代替識別子 |

- `clinic + source_pk + 契約版` は採用しません。
- `pet|sno`のtrim済み連結値は照合表現として保存できますが、元キーの列名・列順・取得値は別に残します。
- source entity ID、observation ID、projection IDはconsumer側の内部参照IDです。これらを旧DB原本の識別子と表示しません。
- 同じ検証済みsnapshot・同じentityのobservation identityは一つ。初期candidateは不変の出典で、現在の表示値は公開されたappend-only解決版へ固定する。
- 同じidentityで、双方が取得済みの同じ列の原値が異なる場合は `raw_conflict` として保留します。既存値を上書きせず、異値候補と両者の出典を別途保全します。runや契約版を変えて新規受入してはいけません。
- `not_captured`補足は候補追記として検証し、下記の解決版条件が成立するまで `comparison_incomplete` とする。既知状態の異値を補足として採用しない。
- 新しい検証済みsnapshotで元データが変化した場合は、新しいobservationとして追記し、旧snapshotの原値は保持します。
- 同じ原値にmapping変更を適用した場合はprojection revisionのみ追加します。同じprojection識別・同じ版で異内容は `projection_conflict`、同内容再送は既存参照です。
- 元キーが重複している／欠けている場合、行番号、出力ID、digest、任意ordinalを追加して元entityを捏造しません。元entityが確定するまでproductの原行根拠として採用しません。

<a id="legacy-evidence-observation-resolution"></a>
### 同一snapshot内の補足取得と解決版

同じ検証済みsnapshot・同じentityには一つのobservation identityだけを持つ。observationの初期candidate参照は不変の出典であり、常に最新の表示値を意味しない。表示する原値集合は、同observationに属するappend-onlyの `observation resolution revision` によって選定する。初版も後続版も各fieldについて実在candidateの同fieldを参照し、raw値そのものを合成・補完・上書きしない。

not_capturedからの補足candidateは、同entity/snapshot、snapshotと取得成果物の出典、comparison profileを検証し、既に取得済みの全重複fieldが同値で、必要coverageが成立した場合にだけ、次の解決版へ採用できる。SQL NULL・空文字・column_absentは既知の状態であり、not_capturedとして上書き可能なものに読み替えない。既知fieldの異値、profile比較不能、出典未検証はconflictまたはcomparison_incompleteのままであり、解決版を進めない。

解決版は全fieldの参照を持つ完全版とし、前版のfield参照を勝手に変更・除外しない。前版のnot_captured参照を補足candidateの取得済みfieldへ置換する場合だけ、そのfieldの出典を新参照へ進める。初期版や旧版は変更しない。同じ参照集合の再送は既存解決版へ解決する。

解決版の作成は期待する前版を同transactionで比較し、同observationの分岐を拒否する。保存済みだが未公開の解決版をGETが自動採用してはならない。publication bindingは解決版IDを固定し、(clinic_id,domain_kind,observation_id)ごとに単調に公開した最大resolution revisionを現在表示へ採用する。同解決版内だけ公開順をtie-breakに使い、未公開解決版をreadが採用しない。公開切替とcatalog_revision更新は同transactionとする。

<a id="legacy-evidence-monotonic-resolution-publication"></a>
publication writerは既存mutable catalog rowを既存権限でlockし現在公開版を再読する。同版なら現在bindingを維持し、previous chainで立証した厳密な後継だけ現在表示を進める。祖先版を新規公開しようとするpublication TXは保留/全rollbackし、receipt/catalogを更新しない。既成功delivery再送は旧receiptだけを返し、旧版を再有効化しない。

同clinicの複数domain lock順はaccounting→estimateに固定する。immutable tableのUPDATE権限は追加しない。readerは最大の公開resolution revisionを選び、同解決版内だけ公開順でtie-breakする。この制限はsource observationの現在表示だけで、過去projectionが古版解決版へpinするのは正常な履歴である。

projection revisionも入力observationだけでなく、使用した解決版とcandidate/field参照を固定する。原値補足で過去projectionの計算根拠を自動更新しない。新しい根拠を使うprojectionは別revisionへ追記する。既存原値と競合するcandidateを新しい解決版や契約版に逃がして採用しない。

source entityの同一性は登録済みのsource namespaceとキー定義に固定します。key profileの単なる改版・改名で別entityを作りません。新旧profile間の同一性が未解決なら保留します。

原値比較は、登録済みの不変なcomparison profileに従い、列名・presence・取得表現を決定的に比較します。JSONのキー順やescape表現の差は原値変更にしません。一方、trim、Unicode正規化、NULLの空文字化、丸め、日時補完を比較前に行ってはいけません。比較profileが異なる場合に同値を証明できなければ保留し、別identityを作りません。

### 依存先

- MIG-41：実在する元キーの列・型・重複性、取得表現、source system/系譜、snapshotと取得成果物の対応。
- EMR-221：source namespace、key/comparison profileの登録・版互換性、snapshot検証receiptとの結合。
- 同じ取得snapshotを別IDで登録して競合を回避しないよう、snapshot aliasを同一検証登録へ解決する必要があります。その登録・照合に必要な既存manifest情報は未受領のままです。

<a id="legacy-evidence-raw-derived-separation"></a>
## 3. closed schema：共通構造の選定と、未閉鎖のwire部分

### 確定

以下は論理schemaです。共通構造の選定であり、nested DTOと実データprofileを含むwire全体のclosed化は未完です。未選定の型をany JSONで代用することは認めません。実装時はJSON Schema / OpenAPIに落とし、**全objectを `additionalProperties: false`** とします。未知key、未知tag、未知profile、必須項目の省略を拒否します。任意の列名をkeyにする自由JSONは設けません。

表中の構造名は契約上の型です。objectの列挙にないkeyは認めません。nullableと省略を区別し、nullable欄もkey自体は必須とします。

### 3.1 共通scalarと取得値

- consumer ID：UUID文字列。UI/APIでは後述の型prefixを付けます。
- producerの識別子：登録されたnamespace内の文字列。出力IDを原キーへ流用しません。
- digest：`{ algorithm: "sha256", hex: 64文字の小文字hex, byte_length: 0以上の十進整数文字列 }`。digestの対象を参照先artifact/profileで固定し、別対象のdigestを流用しません。
- amount / decimal / 件数 / BIGINT：JSON numberではなく文字列。原値では取得文字列のまま保持します。解釈済みdecimalは別の規則登録がない限り作りません。
- consumer自身の監査日時：UTCのRFC3339。旧DB日時をこの型へ無条件変換しません。

`CapturedValue` は以下のいずれかです。

| tag | 必須key | 意味 |
|---|---|---|
| `text` | `tag`, `text` | 取得した文字列。空文字・空白・符号・桁・小数表現を保持 |
| `bytes` | `tag`, `base64`, `encoding` | 実際に取得したbytesのみ。標準base64。encoding不明はnull |

`RawCell` は `{ field_id, presence, captured_value, capture_profile_ref, unavailable_reason }` です。

- `field_id`：登録source profileの列ID。自由列名は禁止。元の列名はprofileが保持します。
- `presence`：`value | sql_null | column_absent | not_captured`。
- `value`だけ `captured_value` が非null。他のpresenceではnull。
- `sql_null`はSQL NULLと確認できた場合のみ。`column_absent`は元schemaに列が存在しないことを確認できた場合のみ。
- `not_captured`では `unavailable_reason` が必須。他ではnull。
- `captured_value.tag=text, text=""` は空文字であり、SQL NULLとは異なります。
- 元bytesがない文字列取得をbytesとして再表現して「元bytesを保存した」と説明してはいけません。

`Reason` は `{ code, evidence_refs }`。codeは `source_unverified | snapshot_unverified | capture_incomplete | comparison_incomplete | raw_conflict | projection_conflict | lineage_incomplete | meaning_unverified | missing_parent | stage_link_downgraded | original_unlinked | mapping_unselected | artifact_unverified` の許可集合とします。新codeはschema版更新を必要とします。evidence_refsは契約内の参照IDのみで、自由なSQL・local path・秘密情報を入れません。

### 3.2 原行候補

`SourceCandidate` のkeyを以下に固定します。

`candidate_ref, source_entity, snapshot_ref, acquisition_refs, capture_profile_ref, cells, match_expressions, source_trace_status`

- `source_entity`：`{ clinic_id, source_system_ref, lineage_ref, table_ref, key_profile_ref, key_cells }`。
- `key_cells`：登録profileに定めた全キー列を登録順に一度ずつ列挙します。key値の `not_captured` / 未解決重複は採用不可です。
- `cells`：登録profileが指定する保全対象列を各一度列挙。未取得も省略せず `not_captured` とします。
- `match_expressions`：`{ expression_profile_ref, text, input_field_ids }` の配列。trim済みpet|sno等はここです。identityを置き換えません。
- `acquisition_refs`：`{ artifact_ref, extraction_profile_ref }` の配列。元バイトoffsetは必須欄として設けません。
- `source_trace_status`：producer申告の `complete | incomplete`。consumer検証済みの意味ではありません。

consumerはsnapshot・artifact・抽出profile・元キーの照合を別に行い、検証状態を付与します。producerが `verified=true` を送るだけで採用できる構造にはしません。

### 3.3 派生projectionと出力結果

`ProjectionRevision` のkey：

`projection_ref, domain_kind, projection_kind, revision_ref, mapping_profile_ref, transform_profile_ref, input_candidate_refs, fields, lineage, outcomes`

- `domain_kind`：`accounting | estimate`。
- `projection_kind`：`canonical | stage | output`。原DB行を表すtagではありません。
- projection namespaceはconsumer登録のsource system/系譜＋projection_kindを識別し、run/delivery/契約版/digestを含めない。wire projection_refは登録profileでnamespace内logical refへ解決する。delivery-local参照しかなければglobal identity対応を221が受領するまで採用保留とし、推測で埋めない。
- DB自然キー一意競合後に、既存UUIDのfingerprintに加え固定入力解決版/規則版/closed内容を照合する。同内容再送は既存UUID、異内容はprojection_conflict。digestをidentityへ追加して別revisionへ逃がさない。

- `input_candidate_refs`：全入力原行候補の集合。重複不可。
- `fields`：`DerivedField` の配列。field_idはprojection profileに登録されたものだけ。
- `DerivedField`：`{ field_id, value, origin, rule_ref, source_field_refs, semantic_status }`。
- `value`：`{ tag: "text", text } | { tag: "decimal", decimal } | { tag: "null" }`。decimalはprofileの文法・scale確定後のみ使用します。意味未確認の数値や旧日時はtextのまま保持できます。
- `origin`：`source_derived | generated_constant | defaulted | unresolved`。
- `semantic_status`：`unverified | legacy_fact_supported`。`legacy_fact_supported`は登録された限定的な旧業務上の意味のみで、新カルテの業務確定を意味しません。今回、KKarte_Kbn=4の「旧精算完了の根拠」は対象ですが、意味未確認のTSeisan_Flg/TGyoSei_Flg、KNJOシェルの定数1、生成insurance/discount/tax/status、補完日付をこのtagで確認済み事実にしてはいけません。
- `source_field_refs`：WireRawFieldRef配列（exact `{candidate_ref,field_id}`）。全寄与field、input_candidate_refsへの所属、存在、重複pair拒否を検証する。保存/read参照は下記のStoredRawFieldRefへ解決する。
- `lineage`：`{ status: "complete" | "incomplete", missing_reason_codes }`。入力集合digestだけを全行列挙の代わりにはしません。
- `outcomes`：`{ stage: "canonical" | "stage" | "csv", disposition: "emitted" | "excluded" | "held", rule_ref, source_candidate_refs, artifact_ref, output_ref, reason_codes }` の配列。artifact_ref/output_refは未出力ならnull。output_refはそのartifact内の照合参照で、原本identityにはしません。

WHERE除外された原行もSourceCandidateとして保全し、該当outcomeを持ちます。原行を取得できていない場合、集約結果から架空のSourceCandidateを作らず、lineage未完として残します。

### 3.4 配送とmanifest

配送envelopeのkey：

`contract_name, contract_version, delivery_ref, producer_run_ref, producer_build_ref, mapping_profile_ref, source_snapshot_refs, artifact_refs, chunk_manifest_ref, source_outcomes`

契約は19 CSVとは別に明示選択します。実artifactの取得bytes digest、snapshot検証receipt、登録profile、manifestの固定された母集団に照合できて初めてconsumerの受領対象になります。

<a id="legacy-evidence-stable-reference-types"></a>
### 3.5 配送local参照と保存stable参照・解決版

`WireRawFieldRef = { candidate_ref, field_id }`
- 両key必須、非null、additionalProperties:false。
- candidate_refは当該deliveryの宣言内で一意のlocal参照。単独では別deliveryと共通のidentityにならない。field_idは当該candidateの登録profileに存在し、cells中で一度だけ現れる列。
- consumerは `(clinic_id, delivery_id, candidate_ref)` を検証済みの永続candidate UUIDへ解決する。同じlocal文字列を異delivery間で同一candidateと推測しない。

`StoredRawFieldRef = { candidate_id, field_id }`
- candidate_idはconsumerが採番し耐久保存したraw candidate UUID。両key必須、非null、additionalProperties:false。
- consumer内のresolutionとprojection入力はこのstable UUIDにpinし、配送namespaceを失ったcandidate_ref文字列を保存参照にしない。必要なlocal→stable対応の保存/receipt bindingは221/446/466の追加受領依存。実在する仕組みとしては未確認。
- 共通で、参照先不存在、異clinic、同一pair重複、field profileの意味のすり替えを拒否する。
- **resolution内だけ**、全field参照が同じobservationのentity/snapshotに属することを要求する。
- **projection内では**、同clinicの固定入力解決版集合に属する限り、複数entity・複数snapshotのfield参照を許可する。単一原行へ制限しない。

`ObservationResolution = { observation_id, resolution_id, revision, previous_resolution_id, comparison_profile_ref, field_refs, verification_ref }`
- revisionは1以上の十進整数文字列。初版1のみprevious_resolution_id=null、以降は同observation直前版を参照。
- field_refsはStoredRawFieldRef配列で、当該source profileの保全対象全fieldを一度ずつ含む。未取得fieldもnot_capturedの実candidate fieldへの参照を持つ。
- verification_refは非nullのconsumer内部UUID文字列とする。consumerが実行した比較・出典・coverage検証の耐久記録を参照し、producerが任意UUIDを発行・自己申告して代用することはできない。検証記録はclinic、observation、候補のstable UUID集合、前版、comparison profile、対象field coverage、検証結論へ束縛され、同resolution保存transactionから参照可能であることを必要条件とする。
- 参照UUIDのscalarと発行権限はここで選定するが、参照先の物理保存場所・記録schema・consumer validatorからの供給receiptは**追加の未閉鎖依存**である。既存6入れ子型だけの未完であるかのように隠さず、221/446/466が受領するまで解決版の実データ採用を開始しない。既存receiptの流用可能性は今回未確認で、架空の既存検証tableや成功記録を仮定しない。

producer wireのsource_field_refsは§3で定義したWireRawFieldRefの配列（exact keys: candidate_ref, field_id）とする。candidate_refはinput_candidate_refsに含まれ、field_idは当該candidateに存在しなければならない。同一pairの重複は拒否する。max/sum等は全寄与fieldを列挙する。生成定数は空配列を許可するがorigin=generated_constantと登録rule_refが必要で、架空source参照を付けない。省略や任意keyは不可。

consumerはdelivery namespace内のwire参照をStoredRawFieldRef（candidate_id UUID, field_id）へ解決・保存する。projection公開時は各stable candidate/fieldが固定入力解決版集合に含まれることを照合する。同clinicの複数entity/snapshotを許可し、projection全体に単一entity/snapshot制約を課さない。producer wireのcandidate参照を現在の最新解決版へ暗黙に付け替えない。APIでこれらの参照を返す場合もconsumerのstable IDを使用し、別deliveryと曖昧なlocal refだけは返さない。

<a id="legacy-evidence-source-outcome-wire"></a>
### 3.6 projectionより前の原行outcome

`WireSourceOutcome = { outcome_ref, source_candidate_ref, stage, disposition, rule_ref, reason_codes }`
- additionalProperties:false、全key必須。
- stage=`canonical | stage | csv`、disposition=`excluded | held`。
- source_candidate_refは実在するSourceCandidateを一つ参照。複数原行ならそれぞれを列挙する。架空のaggregate candidateを作らない。
- rule_refとreason_codesは登録済みのものだけ。業務値未確定でも、除外された事実とそのtransform ruleが検証できることを必要とする。
- rawを取得できない行についてSourceOutcomeだけを捏造しない。

WireSourceOutcome.outcome_refはdeliveryの全chunkを通じ一意。同一配送母集団内の重複は同内容でもvalidation errorであり、黙って集約しない。source outcome保存行にdelivery_id/outcome_refを両方NOT NULLで保持し、source行だけの部分UNIQUE(clinic_id,delivery_id,outcome_ref)をDB強制する。digest/candidateを一意キーへ足して異内容を逃がさない。

既存delivery/chunkの再送は保存済みclosed内容全体を照合し、同内容なら既存outcome UUID、異内容ならconflict/公開なし。固定manifestとchunk receiptで「配送内の二重宣言」と「同じ配送/chunkの再送」を区別する。

outcome_refとsource_candidate_refは当該delivery内のlocal参照であり、consumer保存・readのglobal IDではない。consumerは `(clinic_id, delivery_id, outcome_ref)` を永続outcome UUIDへ、`(clinic_id, delivery_id, source_candidate_ref)` を永続candidate UUIDへ解決する。同じlocal refが別deliveryに存在しても同一と扱わない。元delivery/local refはprovenanceとして保持し、読取DTOのstable UUIDに置き換えた後も出典が辿れるようにする。

配送envelopeへsource_outcomesを追加する（空配列を明示可）。ProjectionRevision.outcomesは従来の派生段階の結果として残す。SourceOutcomeを二重にprojection outcomeへ複写して件数を増やさない。

<a id="legacy-evidence-open-nested-types"></a>
### 依存先：closed schema全体の未完部分

具体的に未閉鎖のnested型は、AcquisitionArtifactBinding（取得artifactと検証receiptの対応）、SnapshotBinding（検証登録の外部参照）、ChunkManifest（chunk集合・digest・件数・上限）、ArtifactRef（成果物roleとstorage binding）、ProvenanceReadDTO（非機密の出典表示）、LinkHistoryReadDTO（targetごとの検証状態と公開範囲）です。上記のartifact_refs / chunk_manifest_ref、詳細APIのsnapshots section / provenance、link_history itemsはこれらの登録型を参照します。型が受領されるまで実データconsumer/API全体は未完成です。公開する型のexact keysを閉じる責務は221/547にあり、old_dbが推測で埋めるものではありません。

6入れ子型に加え、verification_refの検証記録schema/物理保存/consumer供給receipt、delivery-localからstable candidate/outcome/projection/artifact UUIDへの耐久対応と固定manifest内projection outcome位置bindingは追加の未閉鎖依存である。221/446/466が受領するまで実データ解決版採用を開始しない。架空の既存table/receiptを仮定しない。

共通envelopeが閉じても、次は自由JSONで埋めず未選定とします。

- source table別のfield registry、実型・桁・scale・encoding・日時の時刻成分・timezone・原値取得表現。
- key/capture/comparison profileの実データbinding。
- 各projectionの登録field、規則ID、丸めの実仕様、全集約入力原行の列挙方式。
- snapshot、取得artifact、既存manifestとproducer run/build/digestとの実照合。
- 配送chunkのサイズ・件数上限、最大field長・payload長・圧縮形式。上限未設定を「無制限受入可」と解釈せず、実データ受入を無効にしておきます。

担当はMIG-41が根拠・対応表を提示し、EMR-221がconsumer契約として受領、547と446/466が保存/readの対応を確定します。既存6項目の再調査をこの依存解消の名目で繰り返す必要はありません。

### 原値保全の対象別条件

| 対象 | 保存/表示する原値と由来 | 禁止する意味補完 |
|---|---|---|
| 旧billing | 旧status、TSeisan_Flg、KKarte_Kbn（原列表記/変換は静的報告受領済み、実artifact照合待ち）、各日付、請求/現金/カード額を独立原値として保持。元の意味確認状態/保留理由を併記 | 意味不明TSeisan=1やshell生成定数1だけでcompleted認定しない。旧pending/waiting等を一律draft/finalizedへ変換しない |
| 旧精算証拠 | KKarte_Kbn=4は旧精算完了の証拠として区別し、旧精算日を原値として保持 | 旧精算日=新請求確定日とせず、不明な新確定日/終了event/入金配賦を生成しない。定数finalizedから終了eventを生成しない |
| 旧estimate割引 | header discountのraw値と丸め後値、変換規則/版/出典をそれぞれ保持 | 丸め後だけ残して原値を捨てず、割合/金額/適用順を推定しない。`estimate_calculation_revisions`の未定義JSONへ投入しない |
| 旧estimate保険 | insurance=0が生成定数ならgenerated_constant由来を保持して表示 | 原保険額が既知0円だったと解釈しない。生成定数/未取得/意味不明/確認済み値を分ける |
| 元PK不明 | 将来の検証済みartifact locator採用候補で原bytes中の場所と出典を辿る。現producerには元バイト位置locatorなし（[producer報告](#producer-static-report-and-preservation-response)） | 架空source keyを作らず原PK表示しない。artifact digest＋locatorを元レコードの一意性証明と呼ばない |


<a id="legacy-evidence-additional-design-candidates"></a>
旧追加候補のうちpresence/identity/lineage/分割保存/readは本書の共通構造として選定した。§3の6入れ子型と実データprofileは依然未閉鎖で、旧候補を含むwire全体が完成したという意味ではない。

<a id="legacy-evidence-selected-storage-constraints"></a>
## 4. 必要DDLの構造選定と制約（実行SQLではありません）

### 確定：今回選定する保存構造

以下の専用table群をDDL実装の対象とします。**既存DDLに存在するという主張ではありません。** 共通のtenant keyは `clinic_id`。内部参照はUUIDで、すべての親子FKにclinic_idを含めます。

| table名 | 主な必須列・役割 | 主な制約 |
|---|---|---|
| `legacy_evidence_source_entities` | id、clinic_id、source_system_ref、lineage_ref、table_ref、identity_profile_ref、canonical_key bytes、identity digest | 自然identityを一意化。digestだけで同値判定せずcanonical_keyを照合 |
| `legacy_evidence_snapshots` | id、clinic_id、source namespace、検証登録参照、snapshot fingerprint | verified登録との結合必須。run/契約版では別snapshotを作らない |
| `legacy_evidence_raw_candidates` | id、clinic_id、entity_id、snapshot_id、capture profile、raw payload、比較fingerprint、取得出典参照 | append-only。payloadはclosed validator通過時の損失なし表現。任意JSON不可 |
| `legacy_evidence_observations` | id、clinic_id、entity_id、snapshot_id、初期candidate_id（不変の初期出典） | UNIQUE(clinic_id, entity_id, snapshot_id)。observation identityは一つ |
| `legacy_evidence_observation_resolutions` | id、clinic_id、observation_id、revision、previous_resolution_id、comparison_profile_ref、verification_ref、closed field_refs、参照集合fingerprint | append-only。UNIQUE(clinic_id,observation_id,revision)、同observation複合FK、初版previous=NULL/後続非NULLのCHECK。前版revision+1を同TX検証。非NULL previous_resolution_idに対しUNIQUE(clinic_id,observation_id,previous_resolution_id)で分岐拒否。初版はrevision=1の一意性で一つ |
| `legacy_evidence_conflicts` | id、clinic_id、entity_id、snapshot_id、候補A/B、conflict kind、検知delivery | 異値候補保持。採用済み値の上書きや自動解消なし |
| `legacy_evidence_deliveries` | id、clinic_id、delivery_ref、producer run/build、契約版、mapping版、固定manifest digest | UNIQUE(clinic_id, delivery_ref)。同ref異manifestはconflict |
| `legacy_evidence_artifacts` | id、clinic_id、delivery_id、role、digest、byte_length、保護storage参照 | acquisition/output/manifestをroleで分離。local pathは公開APIへ返さない |
| `legacy_evidence_projection_revisions` | id、clinic_id、domain_kind、projection_namespace_ref、projection_logical_ref、revision_ref、mapping/transform profile、payload fingerprint、closed payload | 自然キー全列NOT NULL。UNIQUE(clinic_id,domain_kind,projection_namespace_ref,projection_logical_ref,revision_ref)。同自然キー同内容は既存UUID、異内容projection_conflict。rawは不変 |
| `legacy_evidence_projection_inputs` | clinic_id、projection_revision_id、observation_id、resolution_id、stable入力列参照 | 多対多。全入力行/解決版を固定。同clinic・observationへのresolution所属を複合FK検証 |
| `legacy_evidence_outcomes` | id、clinic_id、outcome_kind、projection_revision/candidate/observation/resolution参照、stage、disposition、rule、artifact/output参照、reason | source/projection kind別NOT NULL/CHECK/FK。sourceはprojection参照NULLを許可、実candidate/observation/resolutionとdelivery_id/outcome_ref必須。source行に部分UNIQUE(clinic_id,delivery_id,outcome_ref)。emitted/excluded/heldを分離 |
| `legacy_evidence_link_events` | id、clinic_id、evidence_kind、evidence_id、event_revision、expected_revision、previous_event_id、action、closed target refs、検証根拠、actor、reason、recorded_at | append-only。履歴粒度(clinic_id,evidence_kind,evidence_id)。同履歴event_revision UNIQUE、非NULL previousのsingle-successor UNIQUE、NULL previousの初版partial UNIQUE、前eventへの同clinic/同履歴複合FK。前版+1/最新headをvalidator検証 |
| `legacy_evidence_chunk_receipts` | id、clinic_id、delivery_id、chunk_ref、chunk digest、件数、保存ID集合のdigest（全てNOT NULL） | UNIQUE(clinic_id,delivery_id,chunk_ref)。同キー同内容は既存receipt、異内容conflict。同時実行でも一件だけcommit、digestを一意キーへ足さない |
| `legacy_evidence_publication_receipts` | id、clinic_id、delivery_id、固定manifest digest、全chunk/lineage検証結果参照、公開revision、recorded_at（manifest digest/公開revision NOT NULL） | UNIQUE(clinic_id,delivery_id)。全必要chunk/lineage検証後に一成功receipt。並行公開は同内容照合、異内容conflict。事前checkだけに依存しない |
| `legacy_evidence_publication_bindings` | clinic_id、publication_receipt_id、domain_kind、typed evidence、resolution_id、basis_kind、projection_revision_id/source_outcome_id、read_field_profile_ref | basis_kind=projection/source_outcomeで対応根拠どちらか一方のみCHECK。source_observation公開は固定resolution_id必須。projectionは全固定入力解決版を参照、domain別field集合を固定 |
| `legacy_evidence_catalog_revisions` | clinic_id、domain_kind、revision | UNIQUE(clinic_id, domain_kind)。publication/link変更/競合記録と同transactionでrevisionを単調更新 |

`raw_candidates`、`observations`、`observation_resolutions`、`projection_revisions`、lineage、outcomes、link events、receiptは通常アプリroleからUPDATE/DELETE不可にします。原値の訂正は新candidateまたは新snapshot、新mappingによるprojection追記です。通常閲覧roleにINSERT権限を付けません。

<a id="legacy-evidence-source-outcome-publication"></a>
公開bindingの根拠は二つに限定する。①検証済みprojectionとその入力原行、②検証済みSourceOutcomeと、対象原行の登録source profileに対するdomain/read field許可mapping。②はprojectionが存在しないearly WHERE除外を公開する経路であり、空・架空projectionを要求しない。

②でもsnapshot/entity/原値出典と解決版、source outcomeのrule、domain別field許可profileをconsumerが検証する。producerのdomain自己申告だけでは公開しない。profile未受領は保留とする。未リンクの除外原行は同clinic・該当domainの移行履歴で読めるが、通常請求/見積には生成しない。

publication_bindingsはbasis_kind=`projection | source_outcome`とし、対応するprojection_revision_idまたはsource_outcome_idのどちらか一方だけを保持するCHECKを置く。typed evidence=source_observationの場合は固定resolution_idを持つ。projection根拠で原行を公開する場合も、その原行が当該projection inputであることを照合する。

outcomes tableはsource-outcome行でprojection参照NULLを許可し、candidate/observation/resolutionへの実在参照を必須にする。projection-outcomeとsource-outcomeをkindで区別し、kind別NOT NULL/CHECK/FKを置く。公開済みraw detailのoutcomes sectionには該当SourceOutcomeを返し、そのitems schemaはclosed union `{ outcome_kind: "source", value: ReadSourceOutcome } | { outcome_kind: "projection", value: ReadProjectionOutcome }` とし、各wrapperはadditionalProperties:falseとする。**producer wire型をそのままAPI responseへ再利用しない。**

`ReadSourceOutcome = { outcome_id, candidate_id, observation_id, resolution_id, stage, disposition, rule_ref, reason_codes }`
- 全key必須、additionalProperties:false。4つのIDはconsumer UUID。source observationの公開bindingが固定したresolutionとcandidateの関係を検証する。
- stage/disposition/rule_ref/reason_codesはWireSourceOutcomeと同じ許可集合。delivery-local outcome_ref / source_candidate_refを無限定IDとして返さない。

`ReadProjectionOutcome = { outcome_id, projection_revision_id, stage, disposition, rule_ref, source_candidate_ids, output_binding, reason_codes }`
- 全key必須、additionalProperties:false。outcome_idとprojection_revision_id、source_candidate_idsの各要素はconsumer UUID。source_candidate_idsは固定projection入力解決版集合に含まれるcandidateだけで、重複不可。
- output_bindingはnull、またはclosed object `{ artifact_id, artifact_record_ref }`。artifact_idはconsumer artifact UUID、artifact_record_refはそのartifact内だけの照合文字列であり元entity/global IDではない。未出力ならnull。保護storage URLやlocal pathは返さない。
- stage/disposition/rule_ref/reason_codesは既存projection outcome wireの許可集合を継承する。
- producer wireにあるsource_candidate_refs、artifact_refはdelivery namespace内でstable IDへ解決して保存する。wireのoutput_refは解決済artifactと組で保存し、readではoutput_bindingへ変換する。

projection outcome wireには現時点でoutcome_refがないため、新規の元record IDを捏造しない。consumerの配送内同一性は、固定manifestで束縛されたdelivery・projection revision・outcomes配列位置（0始まりの十進整数文字列）の組で管理してstable outcome UUIDへ解決する。この配列位置は配送内要素の識別であって旧DB原本identityではない。同じmanifestの再送で位置・内容が違えばconflict。別deliveryの同じ位置だけで同一outcomeと扱わない。producerのprojection local参照からstable revision IDへの受領bindingと位置対応の耐久保存は221/446/466の追加受領依存で、実在実装済みとは扱わない。

wire local→stable対応はconsumer内部provenanceに `delivery_id` と元local参照（projection outcomeは固定配列位置）を併記して保全する。既存の未閉鎖ProvenanceReadDTOへこれらを露出する場合は、その型を閉じてからであり、APIへ裸のlocal文字列をglobal IDとして追加しない。

raw/resolution/projection/publicationのlatest参照をUPDATEで差し替えず、追記bindingと公開revisionで現在版を解決する。過去receipt対象は不変。catalog revisionは公開receipt/link/競合状態と同TX更新する。

全tableにclinic分離のRLSと複合FKを適用し、read時にも現在の認可を確認します。最低限のindexは、tenant＋source identity、tenant＋snapshot、tenant＋domain＋公開revision＋ID、projection input逆引き、対象別link revision、delivery/chunk receipt参照です。

消費側validatorの合格だけでDB制約を省略しません。FK、UNIQUE、enum相当CHECK、nullable条件、immutabilityをDBにも置きます。JSON schemaそのものをDBが実行できない構成なら、書込専用のconsumer経路以外から挿入できないようにし、schema版とpayload digestを保存します。

### 公開と大batch

受領段階は、①配送物の保護保管、②snapshot・identity・原値整合を満たす根拠の採用、③新カルテリンクの検証、④通常業務への正式移行に分けます。②が成立しても③は未リンク・親欠落・stage降格のままで履歴公開できます。リンク未検証だけを理由に原値保全を拒否しません。④は今回すべてHOLDです。source identity不明や原値競合候補の隔離保存は、②の採用や成功receiptではありません。

大batch全体を一transactionで保存できるとの仮定は置かず、**固定manifest＋chunk保存＋最後の公開receipt**を採用します。

- manifestはchunk集合・各digest・対象件数・入力根拠集合を固定します。実行途中で母集団を増減しません。
- 各chunkの保存とchunk receiptは同transactionです。
- 全chunkが一致し、必要lineageと参照が閉じたことを確認した最後のtransactionでpublication receiptを作ります。
- productのreadonly APIはpublication receiptがあるdeliveryだけを可視化します。未完batch、provenance未検証、raw競合、必須lineage未完を成功履歴として公開しません。
- 業務上の意味が未確認なだけなら、根拠の完全性が確認できた証拠は「意味未確認・通常業務反映保留」として公開できます。
- commit結果不明は同じdelivery/chunk refをreceipt照会して解決します。新runやdeliveryを作って成功扱いにしません。
- 後から既存observationのraw conflictが発見された場合、既存原値は消さず、以後のreadで「競合・採用保留」を表示し、新たな正常projection公開を止めます。競合を無視した過去表示を維持しません。

### 依存先

- EMR-446 / 466：現在の001統合方式に合わせた物理DDL、既存clinic ID型、role/RLS、constraint/index名、model/ERD整合。新DDLファイルを勝手に増やす指示ではありません。
- EMR-221 / 349：chunk上限・固定manifest wire・receipt正式schema・未知結果照会の内部経路。未選定の間は実取込を開始しません。
- source identityのcanonical_key最大長とindex方式は実キーprofile受領後に固定します。hashだけのUNIQUEで衝突を原本同一性と取り違えないことを実装条件とします。

<a id="legacy-evidence-target-link-validation"></a>
### nullable target linkの検証条件

医院の複合FK/RLSだけでは同院の別患者への誤linkを防げない。[会計の保存帰属](../../../spec/domains/accounting/lifecycle.md#accounting-owner-transfer-financial-attribution)と[患者譲渡時の保全](../../../spec/domains/owners-pets/lifecycle.md#owner-transfer-draft-retention-20261020)、[billings](../../../architecture/erd-detail/billings.md)／[estimates](../../../architecture/erd-detail/estimates.md)／[medical_records](../../../architecture/erd-detail/medical_records.md)の正規関連を根拠に、埋めた全targetの組合せを検証する。以下は必要な論理条件で、具体reader・履歴証明・lock/closed schemaは受領残。

| 値を持つnullable target | 検証する関係（全対象の同clinic・実在確認に追加） |
|---|---|
| ownerのみ | 検証済みsource対応が当該ownerの履歴帰属を証明すること。氏名/住所の類似から作らない |
| pet | 検証済みsourceが同petを指すこと。ownerも指定した場合は対象履歴に正当なowner/pet帰属を証明する。現在のpet ownerとの一致だけでは過去帰属の証明にも否定にもならない |
| record | source対応と当該recordの保存済みpet/owner関連を検査し、指定したpet/ownerとも履歴基準で整合すること。別患者のrecordへ付け替えない |
| billing | source対応と当該billingの保存済みowner/pet/medical_recordに加えhospitalizationおよび明細source_pet_id/source_record_idのlineageを照合する。supplied owner/pet/recordの非null値が、証拠の指す正しい保存帰属/対象明細に属することを検証。下記の複数頭条件を守り、ヘッダNULLから無関係患者へのlinkを許容しない |
| estimate | source対応と当該estimateの保存済みowner/pet_id/medical_record関連を検査する。supplied petが非nullのときだけ、非nullの保存pet_idとの直接一致を必須とし、矛盾は拒否。supplied petがnullでも、当該estimateのsource対応・保存帰属が検証できるestimate-only linkは許容し、架空petの指定を要求しない。null petをverified petとして表示/filter使用しない。recordがある場合は追加でrecord側のpet/帰属とも整合を検査する。record不在だけで正しい独立見積のpet linkを拒否せず、保存pet_idがnullならそれ自体を患者の証明にせず別の検証済み正規関連が必要。billingも併記する場合は両者が同じ検証済み履歴関係に属する根拠を検査し、同院/同名だけでは結ばない |

[billing header](../../../architecture/erd-detail/billings.md)のhospitalization_idがある場合は[入院](../../../architecture/erd-detail/hospitalizations.md)の保存済みowner/petと、source evidenceが結ぶ対象の帰属を検査する。[billing items](../../../architecture/erd-detail/billing_items.md)のsource_pet_id/source_record_idがある場合は対象明細の患者/元recordと、証拠・supplied targetの関係を検査する。headerのNULLは、入院/元record/対象明細との矛盾を無視する根拠にならない。

複数頭会計では全明細petが単一petへ一致する条件を新設しない。証拠がheader全体を指すのか特定明細/元record/入院の部分集合を指すのか、保存済みlineageで検証する。会計に無関係な別明細のpetを含むだけで証拠をその患者filterへ露出させず、対象との結合を明示できる関係のみverifiedにする。履歴上のowner帰属を現在pet.ownerで置換しない。組合せの意味/対象粒度/具体readerを閉じられない場合はverifiedとせず未リンク保留。これは既存multi-pet会計を禁止する仕様ではない。

targetがnullならその関係を確認済みとは扱わない。他の非null target同士の矛盾をnull欄で隠さない。無関係な同院患者や矛盾する組合せは拒否し、証明不足は当該linkを未リンク保留にする。部分的に証明できたlinkを採る場合も、採用する全組合せが整合することを要し、未証明targetをverifiedに含めない。譲渡前の正当なowner帰属を現在ownerへ自動移管せず、履歴上の関係証明とread時の現在認可を別検査にする。

<a id="legacy-evidence-link-history-concurrency"></a>
追加/訂正/撤回の同一履歴粒度は(clinic_id,evidence種別,evidence_id)。期待revision/previous_event_idと最新headを検証し、event_revision=前版+1（初版1、previous=NULL）でINSERTする。immutable observation/projection rootへのSELECT FOR UPDATEやUPDATE grant追加は行わない。

DB制約は①UNIQUE(clinic_id,evidence_kind,evidence_id,event_revision)、②previous_event_id非NULLの同履歴single-successor UNIQUE、③previous_event_id=NULLの同履歴初版partial UNIQUE、④前eventへの同clinic/同履歴複合FK。前版revision+1/最新head条件を必須validatorで検査し、初版又は同前eventからの並行correct/revokeは片方だけを許可する。unique競合は失敗TX全rollback後、現在認可付き別readで現版を確認して競合を返す。追加lock tableは作らない。

有効link切替/catalog更新は成功event INSERTと同TXで既存mutable catalogへ反映する。失敗側はevent/link/catalogの変更0。原event UPDATE/DELETE不可で、readonly grantへ書込権限を追加しない。linkwriteの具体wire/grantは未閉鎖の受領依存。認可付きowner/pet filterは検証済み有効関係だけへ適用し、未リンク保留/撤回済みを他患者への入口に使わない。

全原本bundleは既存の保護領域へ保管する。同時に画面に必要な対象原値をDBへ耐久保存し、外部bundle/fileの喪失で詳細画面が空になる設計にしない。raw payload bytesを外部参照だけにせず、DBに保存する対象範囲・encoding/bytes digest・typed projectionとの対応を選定する必要がある。大容量の全bundleを無制限にDBへ複製する採択ではなく、サイズ/保持/backupの具体設計は未受領。ローカルquarantineに置くだけで製品閲覧完成としない。


<a id="legacy-evidence-read-api-ui-candidates"></a>
## 5. 読み取り専用APIと閲覧契約

### 確定

既存prefixの下に次を設けます。全て新設設計であり、現行route登録済みではありません。

- GET `/accountings/legacy-evidence`
- GET `/accountings/legacy-evidence/{evidence_id}`
- GET `/estimates/legacy-evidence`
- GET `/estimates/legacy-evidence/{evidence_id}`

`evidence_id`は `obs_<UUID>` または `prj_<UUID>` の型付き文字列です。通常の数値billing/estimate IDを受けません。詳細responseの `record_type` は `source_observation | projection_revision`。集約projectionと原行を同じ無型IDで混同しません。

原行が両domainのprojection入力に使われる場合も、各routeで許可されたfield profileの投影だけを返します。共有原行を理由に他domainの列を丸ごと返しません。domain別のfield許可profileが未選定の原行は、そのrouteから本文を返せません。

会計側は既存の `accounting:view`、見積側は `estimates:view` を毎requestで確認します。共通grantを新設したり、一方の権限で他方の原値本文を返したりしません。関連データに他kindが含まれても、未認可kindの本文や件数を漏らしません。claim/read権限をimport・link訂正権限へ流用しません。

一覧の許可queryは `record_type, snapshot_id, owner_id, pet_id, link_state, cursor, limit` のみ。重複query、未知query、型不正は400です。limitは1〜100、既定50。owner_id/pet_idはANDで、現在有効なverified linkにだけ適用します。未リンク、親欠落、stage降格、撤回済みを元名称や番号一致で救済しません。

- `record_type`既定は `projection_revision`。値は上記2種のみ。
- `link_state`は `verified | unlinked | missing_parent | stage_downgraded | revoked`。
- v1では旧日付range filterと金額filterを設けません。日付の意味・型未選定を曖昧なfilterへ持ち込みません。
- sortはconsumer公開revision昇順、次にtyped evidence ID昇順で固定します。旧日付や生成日付による並べ替えはしません。

一覧responseは `{ items, next_cursor, catalog_revision }`。itemsは固定summary型で、`evidence_id, record_type, snapshot_count, semantic_status, hold_reason_codes, link_state`のみです。原値本文は詳細で読みます。正常な0件はitems=[]とnext_cursor=null。reader障害を空成功にしません。

詳細の基本responseは `{ evidence_id, record_type, domain_kind, snapshot_count, provenance, semantic_status, hold_reason_codes, sections, catalog_revision }`。sectionsは利用可能なsection名の配列で、 `raw_cells | derived_fields | outcomes | inputs | snapshots | link_history` の許可区画を表します。無制限に全lineageを埋め込まず、section queryに `section, cursor, limit, expected_catalog_revision` を用意し、sectionは一つ、limit1〜100、既定50、section指定responseは `{ evidence_id, section, catalog_revision, items, next_cursor }` とします。itemsはsection別の固定schemaに分岐し、任意JSONにしません。raw_cellsは公開bindingの固定解決版から許可されたRawCell。derived_fieldsはDerivedFieldの同じexact key集合を持つReadDerivedFieldとし、source_field_refsだけはStoredRawFieldRefへ解決済みの配列で返す。outcomesは§4のclosed read union（ReadSourceOutcome/ReadProjectionOutcome）で、producer wire型を直返却しない。inputsは `{ observation_id, resolution_id, snapshot_id, field_ids }`、snapshotsは未閉鎖SnapshotBindingの受領後のみ、link_historyは後述の検証済み公開範囲に制限する。該当しないsectionは400です。最初の詳細取得では基本情報と利用可能section名を返し、各sectionの本文は明示requestで取得します。

<a id="legacy-evidence-snapshot-set-read"></a>
### 複数snapshotの集合とpaging

- summary/detailは両方とも `snapshot_count`（0以上の十進整数文字列）を返す。全snapshot集合は新設のpaginated `snapshots` sectionで取得する。単数へ丸めず、無制限配列も追加しない。
- snapshots sectionのitemsは未閉鎖のSnapshotBindingを参照するため、そのexact wire受領前に実データ公開しない。source_observationは1件、projection_revisionは固定入力解決版が参照する全snapshot集合。重複を除きsnapshot UUID文字列昇順でpagingする。
- inputs sectionは `{ observation_id, resolution_id, snapshot_id, field_ids }` とする。各input→解決版→snapshotの対応を保持し、projectionがどのsnapshotのどの原行を使ったかを辿れる。
- 既存query snapshot_idは維持し、「対象evidenceの入力snapshot集合に指定IDが含まれる」というmembership filterと定義する。projectionをsnapshotごとに重複行へ展開しない。全snapshot集合をfilter指定IDへ切り詰めない。
- snapshot_countは未認可関連kindの件数ではなく、該当routeで公開されたevidenceの固定入力snapshot集合を指す。共有原行の他domain関連projectionを数えない。
- snapshot_countはsource_observationで1、projectionは全入力snapshotのdistinct件数。候補で元snapshot未確認のものを0件の公開済みprojectionにしない。sourceなし生成projectionの許容は今回選定しない。

snapshot集合・inputsはimmutable projection入力解決版から導出し、後日別snapshotのcandidateが追加されても過去projectionの集合を変更しない。

<a id="legacy-evidence-section-total-order"></a>
### 全sectionの固定全順序

全て昇順keysetとし、raw_cells/derived_fieldsは(field_id)一意、outcomesは(kind_rank: source=0/projection=1,outcome_id)、inputsは(observation_id,resolution_id)、link_historyは(event_revision,event_id)、snapshotsは(snapshot_id)をlast tupleにする。inputsは同pair一itemでfield_idsを重複除去しfield順へ並べる。除外できるのはJOIN展開重複だけで、異resolutionを合成せず、格納wireの不正重複を隠さない。

field_idは未正規化UTF-8 bytesのunsigned辞書順、UUIDはcanonical lowercase hyphen ASCII bytes順、revisionは正確な整数比較（float禁止）。署名cursorはsection/sort_version/evidence/domain/clinic/期待catalog_revision/last tupleを束縛する。毎page現在認可、catalog変更409/baseからrestart、GET修復なしを維持する。

<a id="legacy-evidence-detail-catalog-binding"></a>
### baseから最初のsectionへのcatalog束縛

基本detailはmetadataとcatalog_revisionを同じDBsnapshotから返す。section内容と返却catalog_revisionもそれぞれ同じDBsnapshotで取得する。section requestのexpected_catalog_revisionは初回/継続とも必須の十進文字列（共通の版表現）。初回はbase値をそのまま渡し、継続はrequest期待版＝署名cursor版＝現在catalog版を要求する。

毎pageの現在認可/元scopeを優先し、既存401/403/404の非開示を維持する。期待版の不足/型不正、requestとcursorの束縛不一致は400 LEGACY_EVIDENCE_INVALID_REQUEST。正常なrequest/cursorの期待版と現在catalogが違えば409 LEGACY_EVIDENCE_VIEW_CHANGED。409後はbaseから再取得し、旧sectionと新metadataを混ぜない。GETで修復/書込を行わない。

cursorはopaqueな署名付きserver形式で、version、clinic、domain、全filter、sort、section、最終位置、catalog_revisionに束縛します。閲覧権限の代替にはしません。公開receiptまたはlink変更・競合検知でcatalog_revisionを更新し、途中で変わったcursorには409 `LEGACY_EVIDENCE_VIEW_CHANGED` を返して再検索させます。単なる最大IDによるwatermarkで整合性を偽装しません。履歴のimmutable本文と現在のlink/競合状態の混在を防ぎます。

errorは `{ code, request_id }` のclosed型とし、raw値・SQL・内部pathを返しません。

- 401：未認証。
- 403 `LEGACY_EVIDENCE_FORBIDDEN`：対象resourceの閲覧権限なし。ID存在を先に開示しません。
- 404 `LEGACY_EVIDENCE_NOT_FOUND`：許可されたclinic内で見つからない。他院IDも同じ404。
- 400 `LEGACY_EVIDENCE_INVALID_REQUEST`：query/ID/cursor不正。
- 409 `LEGACY_EVIDENCE_VIEW_CHANGED`：閲覧集合・link状態変更による再開要求。
- 500 `LEGACY_EVIDENCE_READ_FAILED`：読取失敗。空一覧へ置換しません。

通常のwriterは `obs_` / `prj_` を通常IDとして受理しません。GETでrepair、link追加、取込、再計算、receipt作成等の副作用を起こしません。

<a id="legacy-evidence-record-semantic-status"></a>
### record意味状態とraw整合性は別

record-levelのsemantic_statusは `not_assessed | unverified | partially_supported | legacy_fact_supported` とし、認可されたread_field_profileの表示対象fieldだけからconsumerが決定的に導出する。全raw fieldを見せない権限の利用者へ、非表示fieldの判定状態を漏らす集計はしない。

source_observationは常にnot_assessedとする。原値の整合検証・snapshot検証は業務意味の認定ではない。raw値がKKarte_Kbn=4でも、そのraw observation全体を精算済みと解釈しない。限定的な意味を示す場合は、根拠を持つDerivedFieldのlegacy_fact_supportedを別途表示する。

projection_revisionでは、表示対象fieldsが0件ならnot_assessed、全fieldがunverifiedならunverified、一部のみlegacy_fact_supportedならpartially_supported、1件以上あり全fieldがlegacy_fact_supportedならlegacy_fact_supportedとする。generated_constant/defaulted/unresolvedのfieldはlegacy_fact_supportedにできない。不正なper-field status/origin組合せは保存・公開validation errorとする。

partially_supportedやlegacy_fact_supportedでも旧業務上の限定的な意味を指すだけであり、新請求確定、入金配賦、売上計上、正式移行HOLD解除を意味しない。raw conflictなどのhold_reason_codesはsemantic_statusと別軸で常に表示する。

DerivedFieldのsemantic_statusは従前の2値を維持し、record-levelだけを上記4値へ定義する。

### 表示上の固定事項

- 「移行履歴・参照専用」「通常会計への反映保留」を明示。
- 原値と変換値、生成定数、補完、意味未確認を分けます。
- `insurance_amount=0`等の生成0は「生成定数0」であり、確認済み0円と表示しません。
- 1900-01-01等の補完値は「日付不明」の代わりの業務日付として表示しません。詳細の変換根拠欄にだけ補完値として示せます。
- `KKarte_Kbn='4'`は「旧精算完了の根拠」。新請求確定日時や入金配賦完了とは表示しません。
- 割引の原値と丸め後を分け、負値を保持します。unit_priceへ割引を再適用しません。
- 新カルテへのリンクは検証済みだけ。未リンク、親欠落、stageでの降格、後の撤回を区別します。架空の親IDは作りません。
- local path、保護storage直URL、原bytes本体をこのAPIで配布しません。表示に必要な許可済み原値はDBに保持するため、外部bundleが読めなくても保存済み詳細が空にならない構成です。

### 依存先

- EMR-547 / API owner：既存ID parser・認可middleware・error envelopeとの具体的な統合、api.yamlへの登録、UI実装。link_historyの公開DTO（target種別ごとの検証状態・撤回・理由を含むexact keys）は既存target link検証規則に照合して閉じる必要があります。未選定の間はこのsectionを実データへ公開せず、任意のlink JSONを返しません。
- EMR-221 / MIG-41：各sectionで表示できるfield registry、raw/derivedのprofile binding。
- EMR-446 / 466：catalog_revisionのtransaction更新方式、index、DB role。link変更の書込経路・grantは別契約で、閲覧APIには実装しません。

<a id="legacy-evidence-source-artifact-trace"></a>
## 6. 元キーから取得成果物までの追跡

### 確定

最小追跡は、`source entity/key → 検証済snapshot → acquisition artifact → extraction/capture profile → raw candidate/observation → projectionの全input → output/exclusion`です。

元バイトoffsetや行番号が存在するとは仮定しません。現producerにないlocatorを必須値として要求せず、nullを偽の0で埋めません。将来locatorを採用する場合も、それは成果物内の場所であり、元レコードの自然identityを保証するものではありません。

`COALESCE`等で失われたSQL上の区別と、CSV表現上の区別は別に記録します。COPY既定の非引用空欄＝NULL、引用付き空文字＝空文字は、実出力optionのbindingを確認した後にその意味で扱います。SQLで既に同値化したものをCSV引用状態から復元しません。

### 依存先

取得artifactの中で元キーを再特定できる照合手順、snapshot検証receipt、抽出profile、実際のCOPY option/表現のbindingは未選定・未照合です。ここが閉じないものは `source_trace_status=incomplete` とし、追跡完了とは報告しません。元キーすら確定できない集合は別途隔離保管であり、架空原行をproduct履歴へ作りません。

<a id="legacy-evidence-acceptance-fixtures"></a>
## 7. 受入条件

以下は[会計FR/6AC](../../../spec/domains/accounting/lifecycle.md#requirement-fr-accounting-legacy-evidence-preservation)と[見積FR/1AC](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-legacy-evidence-preservation)の技術受領対応。業務条件の正本は各specで、既存7ACを増減しない。

| 547 AC / 業務正本 | 技術受領/fixtureの接点 | 主な受領責務 |
|---|---|---|
| 1 [旧billing原値](../../../spec/domains/accounting/lifecycle.md#requirement-fr-accounting-legacy-evidence-preservation) / AC-accounting-legacy-evidence-originals | raw/derived/source statusのprojection・保持・表示fixture | MIG-41 mapping→221→547、505境界 |
| 2 [見積原値](../../../spec/domains/estimates/business-rules.md#requirement-fr-estimates-legacy-evidence-preservation) / AC-estimates-legacy-evidence-discount-provenance | 割引raw/丸め/規則と生成定数のlineage・表示fixture | MIG-41、547/508 |
| 3 関連履歴 / AC-accounting-legacy-evidence-related-history | dataset snapshot/occurrence/全target関係検証・filter fixture | 221/349、547/446/466 |
| 4 再送/改訂 / AC-accounting-legacy-evidence-replay-history | identity/派生版/receipt/link期待版の保存trace | 547、221、446/466 |
| 5 認可read / AC-accounting-legacy-evidence-authorized-read | 現在認可・一覧/詳細reader・失敗と空の分離fixture | 547/API owner、既存認可 |
| 6 通常処理分離 / AC-accounting-legacy-evidence-normal-ledger-separation | normal writer/集計/legacy ID型境界のfixture | 547/505/508、349 |
| 7 合成受入 / AC-accounting-legacy-evidence-safe-acceptance | 同証拠の受領→保存→UI/復旧と機密保護の合成fixture | 547実装、361同run実機受入 |


### 確定：合成fixtureで記述する最低条件

1. 同snapshot/entity/rawの再配送は同じobservationへ解決し、run/契約版変更だけでは原本が増えない。
2. 同snapshot/entityの異原値は候補双方を保持してconflict。既存bindingを上書きせず、別runで回避できない。
3. not_capturedの補足は即raw conflictにせず比較未完を区別する。別verified snapshotの変化は追記。
4. SQL NULL、空文字、空白、未取得、列不存在、負値、生成0、補完日付を分けて保存・表示する。
5. max/sum入力とWHERE除外を全原行まで辿れる。未列挙の場合はlineage未完として公開完了させない。
6. 同projection同版同内容は重複しない。同版異内容はconflict。mapping変更は派生版だけ追記。
7. 未検証snapshot、digest不一致、未知profile、原値競合、欠落chunk、未完lineageはpublication成功にならない。
8. chunk保存失敗・commit結果不明・再送を同一delivery/chunk receiptで回復し、新identityで迂回しない。
9. 会計/見積grant、clinic境界、関連kind、全pageの再認可を検査し、他院存在や未認可本文を漏らさない。
10. owner/pet filterは現在有効なverified linkのみ。正当な複数頭会計、過去飼主、独立見積を単純な現在pet一致で壊さず、矛盾した組合せは採用しない。
11. link変更・公開集合変更・競合検知時に古いcursorを409とし、黙って違う母集団の続きを返さない。
12. 外部bundle欠損後もDB保全対象原値を表示できる。一覧正常0件とreader失敗を区別する。
13. legacy IDを通常writerへ渡しても変更できず、通常請求・未請求・入金・売上・見積再計算へ混入しない。
14. 同じ合成根拠について受領→保存→認可read→UI表示とreceipt参照が繋がる。quarantineへの保存だけを547完成としない。

<a id="legacy-evidence-review-acceptance"></a>
### レビュー追補の合成受入条件

15. 同snapshot/entityでnot_capturedのfieldを補足し、全重複既知field一致・出典検証後だけ新resolutionが作られる。旧candidate/旧resolution不変。未公開版はreadに出ず、公開後に新値を参照しcatalogが更新される。
16. 既知fieldの異値、SQL NULL→値、column_absent→値、異snapshot混入、comparison未完では補足版を採用しない。raw conflictを新resolutionへ逃がさない。
17. 同じ前版からの並行補足は一方だけ成功。再送は同じresolutionへ解決。過去projectionの入力原値は補足後も固定。
18. raw observationのrecord semantic_statusはnot_assessed。projectionで全未検証/一部支持/全支持/表示field0の4場合が固定規則どおりとなり、非表示fieldのstatusを漏らさない。
19. 同delivery/chunk refの並行receiptはDB UNIQUEで一件。同deliveryのpublicationも一件。異digestを別receiptへ逃がさない。
20. source_field_refsの未知key、未知candidate、存在しないfield、重複pair、入力集合外candidate、固定resolutionにないfield参照を拒否。異deliveryで同じlocal candidate_refでもstable UUIDを誤共有せず、補足版は既存candidateのstable参照を維持する。projectionの正当な複数entity/snapshot参照は許可する。
21. source/projection outcomeのAPIに裸のdelivery-local参照を返さず、同じlocal文字列を持つ別deliveryをstable UUIDで区別し、元出典対応は保全する。同じmanifestのprojection outcome配列位置で異内容の再送はconflict。projection生成前のWHERE除外でも、検証済みSourceOutcome＋source/domain許可profileから同clinic/domainの原行履歴を公開できる。架空projection/通常請求なし。別domain本文は出ない。
22. 2snapshotのprojectionはsnapshot_count=2、snapshots両件、inputs対応を保持。snapshot filterはmembershipのみで集合を切り詰めず、一覧に一件だけ返す。
23. chunk/field/payload上限が未登録なら実データ受入は無効。無制限を既定値にしない。verification_refの参照先不存在・別clinic・別候補集合・別前版・producer自己申告のUUIDを拒否し、検証receipt供給未受領なら実データ解決版採用を無効にする。

24. 初版link同時作成と同じ前eventへのcorrect/revoke並行は成功一件、他方event/link/catalog変更0。競合TXをrollbackしてから別の現在認可付きreadで現版を確認し、immutable root lock/UPDATE grant追加なし。
25. base→初section間とpage間の公開/link/conflict変更は409でbaseから再取得。base/sectionの本文と版は各同snapshot。期待版欠落/不正、request/cursor不一致は400。失権/跨scopeは既存401/403/404を優先し、旧section混合/GET修復なし。

26. 同projection自然キーの並行配送はDB UNIQUEで一UUID、固定入力解決版/規則版/closed内容一致なら再送解決、異内容はprojection_conflict。run/delivery/契約版/digest変更でidentityを逃がさず、global対応未受領のlocal refは採用保留。
27. 各sectionの境界、revision9→10、UTF-8 field順、同itemのJOIN重複、異resolutionの分離、同catalogでの反復pageを確認。変更は409でbaseからrestartし、順序だけでitem欠落/重複を作らない。

28. r2公開後のr1遅着はpublication TX全rollback、receipt/catalog不変。並行公開はcatalog lock/previous chainで単調化し、既成功旧delivery再送は元receiptだけ。過去projectionのr1 pinは正常なまま。
29. 全chunkを跨ぐ同outcome_refの二重宣言は同内容でもvalidation error。同delivery/chunk再送は固定manifest/receiptとclosed内容一致なら既存UUID、異内容conflict/公開なし。digest/candidate変更でUNIQUEを迂回しない。

これらはfixtureの作成条件です。今回テスト実行や実DB受入を依頼していません。現時点はNOT_RUNです。

### 依存先

実データの合格判定は、登録profile、manifest binding、source実型・日時・原値表現、集約元列挙、署名/receipt/上限制御を受領した後です。合成fixtureの通過を現行19表や既出bundleの正式受入へ転用しません。

<a id="legacy-evidence-delivery-ownership"></a>
## 8. producer / consumer の責任分担

| 担当 | 確定した責任 | 今回は行わないこと・依存 |
|---|---|---|
| MIG-41（producer作業項目owner。実producer責務との区別は冒頭参照） | 原行と生成・補完・変換の区別、元key/source profile、snapshot/取得artifact対応、集約・除外全lineage、配送版とartifact宣言を提示 | SQL変更・再生成・実DB操作は別の実施許可後。既存manifest binding・実型/日時/原値・未列挙入力は未解消 |
| EMR-221 consumer contract | strict schema/profile登録、identity照合、snapshot検証登録、manifest/chunk/receipt、verification_ref耐久記録・delivery-local→stable UUIDとoutcome位置bindingの追加供給、source outcome/domain field profile、formal隔離 | producer自己申告を検証済みにしない。19表契約を黙示変更しない |
| EMR-446 / 466 | 解決版を含む専用DDL、receipt UNIQUE、複合FK/CHECK/RLS/immutable/index、verification_refとlocal→stable対応の保存供給、公開/catalog同TX、model/ERD/001統合 | DB適用や既存DB破棄は今回対象外 |
| EMR-547 | 原値保全、projection/outcome/lineage保存、認可readonly API/UI、型付きID、通常業務非混入、合成AC | 不明税・保険・割引・状態を通常業務値へ確定しない |
| EMR-349 | 正式に許可された段階で唯一のimport実行経路を統括し、配置・preflight・receiptを確認 | 今回はHOLD維持。新たな取込経路を作らない |
| EMR-505 / 508 | 通常会計・見積writer/reader/集計との分離、検証済みlinkの帰属整合 | legacy保全から通常請求を自動起票しない |
| EMR-361 | 実施許可後の同run・同identityによる画面と復旧の受入 | 今回の設計選定を実機PASSにしない |

<a id="legacy-evidence-visit-type-dependency"></a>
## 9. KKarte_Kbn 1 / 2 と初再診の別依存

`KKarte_Kbn=1/2 → visit_type=first/revisit` はこの契約では解消しません。正しい初再診根拠が受領されるまで確定情報として採用せず、代替値も推測しません。

- producer側担当：MIG-41。初再診の正しいsource根拠とmapping修正案の提示。
- consumer側担当：EMR-221。通常19 CSVのvisit_type受入との整合と、既出候補への影響範囲の判定。
- 実行・正式受入依存：EMR-349。修正mappingと対象成果物の適合が確認されるまで正式受入を進めない。
- EMR-547は元のKKarte_Kbnとその由来を保全するのみ。今回の返却はSQL変更指示ではありません。

<a id="legacy-evidence-unreceived-contracts"></a>
## 10. old_db側への次の依頼範囲

本選定を設計入力として受領してください。いま実装開始・SQL変更・実データ再生成・DB操作・正式取込を求めているわけではありません。

次に閉じる必要があるのは、未解消と明示したsource profile、manifest/snapshot/acquisition binding、集約入力全件の対応、chunk/size上限です。既に受領済みの静的調査を再提出するのではなく、各依存について「既存報告から設計に転記できる範囲」と「まだ根拠未受領の範囲」を分けて引き渡します。

正式移行のHOLDは維持します。19 CSV・新sidecar・旧21表成果物の適合状態は、それぞれ独立して扱います。

<a id="producer-static-report-and-preservation-response"></a>

## 11. producer静的報告の受領履歴（今回の選定と区別）

<a id="producer-report-evidence-boundary"></a>
### 受領根拠と検証境界

Library添付「貼り付けたテキスト(5).txt」のproducer回答を受領した。以下のold_db所見は**producer報告**であり、本consumer側で実コード/実DB/既出bundleを独立検証した結果ではない。報告のmainは`a38e71f904460ba3d33194e817bafcbb78886d0c`。主な調査対象は`sql/migration/raw-to-intermediate/020_canonical.sql`、`sql/migration/intermediate-to-mapping/030_stage.sql`、`scripts/lib/intermediate-to-mapping/stage-csv-columns.mjs`、SQL Server DDL。これは先行するmapping設計入力の受領記録。本書§1–10の後発技術選定とは区別し、実コード/DB/成果物の独立検証や実データ生成/取込許可ではない。

producerは現行コードのdigestを報告しているが、実artifact manifestのsource/run/build/contract/mapping digestが現在コード・出力bytesに一致するかは未検証。旧21表のreceiptを現行19表の適合証拠に転用しない。八王子の正式intakeはINCOMPLETE、城東候補は旧21表時代の報告であり、今回正式取込GOへ昇格しない。

<a id="producer-report-loss-and-sidecar"></a>
### 失われる情報と別sidecarの保存方針

| producer報告の所見 | consumer側の回答・保全境界 |
|---|---|
| 現行19 CSVはsource_table/source_pk、原flag、discount_raw、run/confidence等を出力しない。canonical段階でもmax/sum/trim/COALESCE/WHERE除外がある | 既存19 CSVを変更/旧形式へ戻すのではなく、対象の旧会計/見積に別の版付きevidence-only sidecarを設計する。生成/取込はまだ許可しない |
| 多数原行からcanonicalを集約、stage/CSVでは列や対象行がさらに落ちる | 原始source行、derived canonical/stage/CSV、それぞれへの入力原行集合と規則のlineage、除外理由を区別して保全。集約済み1行から原行を復元したとは扱わない |
| 同一カルテへの複数請求の一部はstageでmedical_record_idをNULLへ降格。元から未リンクの行もある | link降格と元から未リンクを別の由来として保持。canonicalの候補関係を無検証でverified targetへ昇格しない。[target link検証条件](#legacy-evidence-target-link-validation)の帰属/現在認可検査を維持 |
| 生成insurance=0、rate=0、tax=0.10、見積draft、補完1900-01-01、KNJO shell settle_flg=1 | generated_constant/default/補完としてrawと分ける。確認済みの保険額/率/税率/業務状態/日付/精算完了を意味しない |
| Mit_Nbiki/Mim_Nbikiのrawと丸め値が異なる。負値もあり得る。unit_priceが割後の実効単価になる経路がある | raw/丸め値/規則を分離し、負値を捨てず、不明0補完や割引の二重適用をしない。正規計算revisionへ未知値を詰めず通常処理HOLD |
| Kkarte_Kbn=4は旧精算状態の証拠。settle_flgの定数や意味不明値からcompletedが生成され得る | Kbn4・精算日・額は独立原値。新finalize/allocate/終了eventを生成せず、waiting/pendingからdraft/finalizedへ一律変換しない |
| Kkarte_Kbn 1/2（報告上の注釈は受付済/入力途中）からvisit_type first/revisitを生成するmappingがある | 初再診の根拠には採用できない。547で原値を保全するだけでは解消しない別のproducer/consumer修正責務として引き渡す。今回SQL変更/新しい初再診判定は行わない |

<a id="producer-report-identity-axes"></a>
### identityの3軸と未選定の結合

| 軸 | 必要な意味と混同禁止 |
|---|---|
| 元entity | 医院/source system・系譜/table/構造化された元複合キー。trim済み`pet|sno`だけで元の複合キーを代替せず、列ごとの原値/presenceを保持。stage IDENTITYや生成IDは原本identityではない |
| snapshot evidence revision | 元entityと検証済みsnapshotを結ぶ証拠版。同snapshot・同原本identityの異なる内容はconflict。別の検証済みsnapshotにおける実変更は旧証拠を上書きせず別revisionへ関連付ける。snapshot間の同一entityは検証済みキー/系譜を要し、名前類似で推測統合しない |
| delivery / derived projection | run/build/契約版/mapping版/encoding等は受領・生成の別軸。同原本に別変換が必要ならimmutable派生projection revisionへ追記。原本を二重生成したり派生差を原本改変と誤判定したりしない |

producerには元バイト位置locatorが存在しないという報告であり、snapshotと原キー・取得成果物を結ぶ方法は未選定。原キー欠落に架空key/locatorを作らず、stage生成IDやrun変更で同一原本の異内容を別新規へ逃がさない。[identity/再送条件](#legacy-evidence-identity-retry-and-atomicity)のsnapshotを含む証拠identity・conflict原則は維持する。producer報告内の「clinic+source_pk+契約版」案は、本書の原本identityへ契約版を戻す採択ではない。

<a id="producer-report-sql-and-csv-distinction"></a>
### SQLの情報欠落とCSV表現の区別

資料のCSV説明は[PostgreSQL 17 COPY公式仕様](https://www.postgresql.org/docs/17/sql-copy.html)に照合して訂正する。COPY CSVの既定NULLは非引用空欄、空文字は引用付き空文字であり、CSVという形式だけで両者が同一になるわけではない。実出力のCOPY option/bytesとの適合は別途確認する。事前のCOALESCE/trim/集約/WHERE除外で消えた区別は、このCSV表現から復元できない。SQL変換の情報欠落、CSV表現、未取得を別問題として保全する。


<a id="producer-report-remaining-receipts"></a>
### 未受領と未実装の最終区分

共通構造・専用保存table群の役割/関係制約・型付きread API/ページング/エラー・合成条件（本節の初回/レビュー追補を列挙）は今回選定済み。6入れ子型は[§3の未閉鎖型](#legacy-evidence-open-nested-types)を参照。6入れ子型に加え、consumer検証記録/verification_refの保存schema・供給binding、delivery-local参照からstable candidate/outcome/projection/artifact UUIDへの耐久対応、chunk件数/bytes・field長・payload上限も未閉鎖。source実型/原値表現/日時、登録profile、取得artifact/snapshot/manifestの実binding、集約元全列挙も未解消。MIG-41の静的報告は受領済みでも、実artifact一致/実行receiptは未検証。保存/認可read/UI/通常writer分離の実装、DDL適用、試験/受入は別の未完工程であり、全wire完成や正式移行GOとしない。
