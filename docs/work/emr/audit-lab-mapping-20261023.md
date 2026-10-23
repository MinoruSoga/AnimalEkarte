# 機器検査mappingの候補生成・人confirm・保存接続の限定静的監査

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-lab-mapping-20261023 |
| 種別 | 固定mainの限定静的監査受領記録 |
| 決定状態 | 既存仕様を継承。新しい業務・公開契約を採択しない |
| 実装状態 | 以下の静的分類に限定。全票・全ACの実装監査完了ではない |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

本稿は提供された監査報告の整形であり、文書編集時に全callerを再監査した記録ではない。[今回の索引](emr-audit-ready-handoff-index-20261023.md#audit-ready-scope)を参照。

固定main: `77c28fbfbafa257dcca2421af99ff72da6889717`。2026-10-23。最新Plane本文と全28コメント、関連8PR、固定コード/fixtureを読取。実装・試験・build・codegen・DB・実機・Mac・実LLM・臨床確定・Plane投稿は一切未実施（NOT_RUN）。

<a id="lab-mapping-conclusions"></a>
## 結論

既存の辞書/LLM候補・人が提出した行だけの確定・監査TX・species overlay・LLMなし取込はソース上追える。しかし現mainにはnullable JSON caller型不整合、原本先行保存、全mapping writer版参加/C06、適用版reader、明示再解決、record identity/unknown dateの既知残件がある。全フロー完成とは認定できない。新規票は0、既存ownerへ限定する。

STG branch、実deploy SHA、実secret/LLM設定、現在医院master、実人confirm/実機受入は未確認。mainの型差分を「現在STGで発生中」と断定しない。機器の自動判定AIではなく未対応item codeの候補補助であり、臨床値や患者情報の追加外部送信は提案しない。

<a id="lab-mapping-evidence-status"></a>
## 証拠の区分

- STATIC_TRACEDは静的配線のみ。対象コードに型不整合があるため各経路の実行成功を意味しない。
- KNOWN_REMAINDERは採択済契約と旧runtimeの接続残件。後続APIは先行契約と明記済みで、APIが誤って完成宣言しているという新欠陥にしない。
- PR807の006保存契約は現在001_init.sqlへ統合済み。旧006の不在を欠落扱いしない。
- 10/11合成DB、dictionary-only、7PASS/1FAILの追加UTC照合、import6PASS＋別counterFAILは歴史証拠。現main/UAT/production FKへPASSを転記しない。
- Plane assigneesは空。以下ownerは既存責務票であり新しい個人割当や期限確約ではない。

<a id="lab-mapping-ac"></a>
## AC行別

<a id="lab-mapping-ac-01"></a>
### 297-01 結果1件だけを選択。同fingerprint再送は畳む。握手・異なる複数結果は失敗draft。
判定: **STATIC_TRACED**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_mapping_select.go: selectOneLabResult](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_select.go)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_draft.go:62–138](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_draft.go)
- 既存/必要fixture（未実行）: TestLabMappingDraftService_SelectFailures
- 残件・境界: selector既存。元SHA/offset/decoder版/acquisition/linked_job digestの保存は未接続。保存済frameがないという旧一律blockerは採用しない。
- owner: 297（入力・provenance）、268（実機取得条件）

<a id="lab-mapping-ac-02"></a>
### 297-02 辞書で一意な項目はLLMなし、未対応だけ候補生成、comment対象外、失敗はblank。
判定: **STATIC_TRACED**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_mapping_draft.go:149–263](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_draft.go)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_llm.go:27–50,211–231](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_llm.go)
- 根拠: [backend/cmd/api/composition_medicalrecord_services.go:196–203](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/cmd/api/composition_medicalrecord_services.go)
- 既存/必要fixture（未実行）: TestLabMappingDraftService_DictionarySkipsLLM / TestLabMappingDraftService_BlankPaths / TestLabMappingDraftService_LLMOmitsMeasuredValue
- 残件・境界: APIサービスからdrafterへの配線はある。機器自動判定AIではない。実送信設定/実呼出/動作は未確認。後述nullable型不整合を解消する必要。
- owner: 297、549/543 caller移行

<a id="lab-mapping-ac-03"></a>
### 297-03 gpt-6-luna、store:false、reasoning:none、structured output、toolsなし。測定値や飼主情報を追加送信しない。
判定: **STATIC_TRACED**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_mapping_llm.go:17,27–50,96–106,211–231](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_llm.go)
- 既存/必要fixture（未実行）: TestOpenAILabItemDrafter_DoesNotSendMeasuredValues
- 残件・境界: 現requestはsource/species/code/unit/候補ID・name・unit。測定値/ペット名も含まない。現在STGのsecret/API key/provider/configは未確認。追加送信を新仕様にしない。
- owner: 297

<a id="lab-mapping-ac-04"></a>
### 297-04 明示draftの原本をLLM前に耐久保存し、中断・失敗後も追跡可能。
判定: **KNOWN_REMAINDER**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_mapping_draft.go:69–112,135–146](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_draft.go)
- 根拠: [docs/work/emr/emr-297-device-mapping-bridge-contract-20261021.md: §3,§4B,§7a](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-297-device-mapping-bridge-contract-20261021.md)
- 既存/必要fixture（未実行）: TestLabMappingDraftService_StopSkipsLLMAndKeepsRaw / TestLabMappingDraftService_StopBetweenLLMCallsKeepsExchanges
- 残件・境界: CreateはdraftItems→ctx.Err検査→insertView。LLM/DB読取エラー・cancelでraw保存前にreturn。既存stop fixtureは一般中断の先行耐久保存証明ではない。sample claim private_raw→Apply→generationを接続。
- owner: 297-A、445/446/466保存subset

<a id="lab-mapping-ac-05"></a>
### 297-05 LLM呼出行だけ送受信、辞書行SQL NULL。JSON null/不正応答/バイナリを失わず保存再読込。
判定: **STATIC_TYPE_BLOCKER**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_mapping_draft.go:248–249](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_draft.go)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_llm.go:47–48,167–182](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_llm.go)
- 根拠: [backend/internal/model/lab_device_mapping_draft.go:59–60](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/model/lab_device_mapping_draft.go)
- 既存/必要fixture（未実行）: TestOpenAILabItemDrafter_HTTPBodyBoundaries / TestLabMappingDraftRepository_InvalidHTTPBodiesPersist
- 残件・境界: json.RawMessage値を*json.RawMessage列へ直接代入が残る。既知PR729 caller差分と同根因、549/543と連携。NULL/JSONnull/コピー所有権を保つcaller移行が必要。不正HTTP bodyを文字列/base64 JSONで保持するhelperは既存。compile実行はしていない。
- owner: 297（caller）、549/543（nullable移行統合）

<a id="lab-mapping-ac-06"></a>
### 297-06 人が提出した行だけ確定、別医院/失敗/重複/対象外commentを拒否。確定監査を同TX保存。
判定: **STATIC_TRACED**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/routes_lab.go:26–29](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/routes_lab.go)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_handler.go:135–170](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_handler.go)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_draft.go:309–499](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_draft.go)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_store.go:100–220](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_store.go)
- 既存/必要fixture（未実行）: TestLabMappingDraftService_ConfirmWritesOnlySubmittedRows / TestLabMappingDraftService_ConfirmGuards / TestLabMappingDraftService_ConfirmRejectsFieldOutsideClinic / TestLabMappingDraftService_AuditErrorRollsBack / TestLabMappingDraftHTTP_AuthAndConfirm
- 残件・境界: selected-clinic edit、staff、draft row lock、submitted rowsだけupsert、audit rollbackは既存。新C06/CAS/失権のTX内gateとは別。
- owner: 297

<a id="lab-mapping-ac-07"></a>
### 297-07 confirm競合はdraft/mapping/field版と現在権限/親検査種別有効性を同TXで検査。
判定: **KNOWN_REMAINDER**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_mapping_handler.go:26–49](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_handler.go)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_draft.go:348–456](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_draft.go)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_store.go:134–154,224–233](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_store.go)
- 根拠: [backend/docs/api.yaml:12589–12616,35206–35236](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/docs/api.yaml)
- 既存/必要fixture（未実行）: 必要追加fixture: 別draft同scope競合 / field定義変更 / 親検査種別inactive/deleted / Apply時失権 / 同keyreplay
- 残件・境界: 旧DTOはitemsのみ、MarkConfirmedはstatus条件のみでVersion更新なし。FindExamFieldはclinic/idのみで親のactive照合なし。保存版modelがあるだけではC06対応ではない。APIは先行契約と明記し切替前旧応答を許容。
- owner: 297-A＋447設定writer、489共有operation必要subset

<a id="lab-mapping-ac-08"></a>
### 297-08 確定版はimmutable snapshot。manual PATCH/configuration/spec/confirm全writerが同headに参加。
判定: **KNOWN_REMAINDER**。実行: NOT_RUN。

- 根拠: [backend/migrations/001_init.sql:11436–11595](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/migrations/001_init.sql)
- 根拠: [backend/internal/model/lab_device_mapping_revision.go](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/model/lab_device_mapping_revision.go)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_store.go:186–220](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_store.go)
- 根拠: [backend/internal/medicalrecord/lab_device_item_master_service.go:203–246,279–307](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_item_master_service.go)
- 既存/必要fixture（未実行）: 必要追加fixture: 全writer版増分 / snapshot全active行 / rollback / generic/spec順lock / legacy由来unknown / reader非活性
- 残件・境界: 旧upsertはmutable rowのみ。heads/revisions/targetsは001へ統合済みで旧006不在をDDL消失と数えない。draft.Versionもある。全writer参加前新版readerをactivateしない。
- owner: 297-A＋447、DDL窓口445/446/466

<a id="lab-mapping-ac-09"></a>
### 297-09 保存済generic＋同species確定表だけをreaderに適用。異種fallback/LLM再呼出なし。
判定: **STATIC_TRACED**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_item_master_service.go:319–410](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_item_master_service.go)
- 根拠: [backend/internal/medicalrecord/lab_device_item_master_repository.go: FindMappingSpecs](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_item_master_repository.go)
- 根拠: [backend/internal/medicalrecord/lab_device_exam_persist.go:69–117](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_exam_persist.go)
- 既存/必要fixture（未実行）: TestLabMappingDraftService_SpeciesConfirmWritesSpecNotDeviceMaster / TestLabMappingDraftRepository_ConfirmAndSpeciesOverlay / TestFindMappingSpecs_EmptySpeciesDoesNotTouchDB
- 残件・境界: 現readerはmutable master/specを読む。draft本体の未確認候補は読まない。既存manual masterの有効mappingも使うので「全項目が297の人confirm必須」と読み替えない。適用版固定は次項の残件。
- owner: 297

<a id="lab-mapping-ac-10"></a>
### 297-10 受信からpersistまで適用版を固定し、保存済結果の由来を再構成しない。
判定: **KNOWN_REMAINDER**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_receive_service.go:113–151](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_receive_service.go)
- 根拠: [backend/internal/medicalrecord/lab_device_exam_persist.go:87–117](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_exam_persist.go)
- 根拠: [docs/work/emr/emr-297-device-mapping-bridge-contract-20261021.md: §6](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-297-device-mapping-bridge-contract-20261021.md)
- 既存/必要fixture（未実行）: 必要追加fixture: receive/persist間mapping変更 / 同jobapplication版 / 未知species / 設定競合
- 残件・境界: 受信時とpersist時にそれぞれcurrent mappingを解決。application snapshot/同TX CASはこのreaderに未接続。daily raw非保存と明示sample raw保存を混ぜない。
- owner: 297-B＋268

<a id="lab-mapping-ac-11"></a>
### 297-11 未知項目はneeds_review、未確定しかない場合は検査書込0。原値/単位を変換しない。
判定: **STATIC_TRACED**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_exam_persist.go:95–141,265–300](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_exam_persist.go)
- 根拠: [backend/internal/medicalrecord/lab_device_item_master_service.go:371–389](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_item_master_service.go)
- 既存/必要fixture（未実行）: TestLabDeviceExamPersister_UnmappedDoesNotWriteExam / TestLabMappingDraftService_UnitMismatchIsNotConverted / TestLabDeviceJobMappingCounts_RecalculatedPartialAndZero
- 残件・境界: mapping 0件ならneeds_review、既存部分対応は対応分だけ保存。全job未確定という理由で既存確定mappingまで拒否する新仕様を作らない。原unit空だけmapping unit fallback、数値変換なし。
- owner: 297、360既存成果再利用

<a id="lab-mapping-ac-12"></a>
### 297-12 mapping更新後の未persist jobは明示再解決。重複受信・persisted再試行で勝手に検査を増やさない。
判定: **PARTIAL_KNOWN_REMAINDER**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_receive_service.go:113–128,394–435](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_receive_service.go)
- 根拠: [backend/internal/medicalrecord/lab_device_exam_persist.go:72–78](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_exam_persist.go)
- 根拠: [backend/internal/medicalrecord/routes_lab.go](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/routes_lab.go)
- 根拠: [docs/work/emr/emr-297-device-mapping-bridge-contract-20261021.md: §4D](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-297-device-mapping-bridge-contract-20261021.md)
- 既存/必要fixture（未実行）: TestLabDeviceJobMappingCounts_StoredWithoutStatusTransition / TestLabDeviceJobMappingCounts_PersistedRetryDoesNotWrite
- 残件・境界: 既存Attach→persister再解決あり。同fingerprint受信は既存job返却のみ、persistedはnoop。新resolve-mapping、mapping-context、operation結果照会はroute未接続。unknown外部送信状態の再開/C06 lookupも未接続。
- owner: 297-B＋268（再解決）、360（件数）

<a id="lab-mapping-ac-13"></a>
### 297-13 新draftは旧履歴/確定表を変えず、生成停止しても確定表取込は継続。
判定: **STATIC_TRACED**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_mapping_draft.go:82–112,299–307](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_draft.go)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_store.go:39–90](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_store.go)
- 根拠: [backend/internal/medicalrecord/lab_device_exam_persist.go](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_exam_persist.go)
- 既存/必要fixture（未実行）: TestLabMappingDraftService_SecondDraftDoesNotWriteChartMappings / TestLabMappingDraftService_StopSkipsLLMAndKeepsRaw / TestLabMappingDraftRepository_ConfirmAndSpeciesOverlay
- 残件・境界: 生成switchはdraft経路限定。現在の停止は耐久generation_state/pending/unknown管理の代わりではない。
- owner: 297

<a id="lab-mapping-ac-14"></a>
### 297-14 人が画面で候補修正・確定・停止できる一連の実製品caller。
判定: **STATIC_UNKNOWN**。実行: NOT_RUN。

- 根拠: GitHub code search mapping-drafts / NewLabMappingDraftService (index SHA171b565; absence only)
- 根拠: [backend/internal/medicalrecord/lab_device_mapping_handler.go](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_mapping_handler.go)
- 既存/必要fixture（未実行）: 必要fixture: draft取得→項目修正→confirm→保存済表再利用のFE/HTTP結合
- 残件・境界: 検索はBE/docs/testsのみ。専用FEの実行callerはこの限定監査で見つからず、manual master編集画面とdraft確定UIを同一視しない。完全tree網羅なしのため絶対不存在とは断定しない。
- owner: 297（専用callerと受入）

<a id="lab-mapping-ac-15"></a>
### 297-15 選択record identityとunknown dateを守り、保存結果をD03 readerへ同じID/版で引渡す。
判定: **KNOWN_REMAINDER**。実行: NOT_RUN。

- 根拠: [backend/internal/medicalrecord/lab_device_exam_persist.go:121–125,290–298](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_exam_persist.go)
- 根拠: [backend/internal/medicalrecord/lab_device_receive_service.go:394–435](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/lab_device_receive_service.go)
- 根拠: [docs/work/emr/emr-297-device-mapping-bridge-contract-20261021.md: §5,§7b](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-297-device-mapping-bridge-contract-20261021.md)
- 既存/必要fixture（未実行）: 必要fixture: wait record CAS / 別record/同pet / 日時unknown / 516 owner-link / 部分persist再解決拒否
- 残件・境界: 現persistはpetIDのみ、MedicalRecordID未伝達。MeasuredAt欠損はnow補完。既知297/268→516、501の日時/reader責務へ有限化。PDF OCR仕様は本監査対象外。
- owner: 297-B＋268、516結果owner-link、501日時/reader

<a id="lab-mapping-pr-adoption"></a>
## PR採用履歴

- [PR #596](https://github.com/MinoruSoga/AnimalEkarte/pull/596): merged、`b88fbf0d50e56b47a1e95888a555ba5db0bffa0d`。feat: 検査対応の下書きを確定後にだけ取り込む
- [PR #631](https://github.com/MinoruSoga/AnimalEkarte/pull/631): merged、`b9032003dce6717a6ad603127f6dd7829f205353`。fix: persist recalculated lab mapping counts (EMR-360)
- [PR #784](https://github.com/MinoruSoga/AnimalEkarte/pull/784): merged、`b48a67d9908a8172400956edf38d3860138d9b7f`。docs: 採択契約を1PRへ集約・固定レビュー対応根拠を追記
- [PR #798](https://github.com/MinoruSoga/AnimalEkarte/pull/798): merged、`582178a28961a7953f4ea8346c9629d0bb51478f`。docs(api): 回復Slice Aの結果確認契約と493・297追補
- [PR #805](https://github.com/MinoruSoga/AnimalEkarte/pull/805): merged、`4ec7f8013bdd96d0f1cd8fcb2deaa3483137dc34`。docs: DOCS-STAGES-CROSSCHECK-20261021 ①〜③の受領漏れと5枠境界を補完
- [PR #806](https://github.com/MinoruSoga/AnimalEkarte/pull/806): merged、`ec8210c71cb594931eb39478b78460f91913aeaf`。docs(api): EMR-445 C/D/E向け api.yaml 限定差分（入院プランPATCH presence・campaign/物販 exact decimal）
- [PR #807](https://github.com/MinoruSoga/AnimalEkarte/pull/807): merged、`9eef73feac19c25571724372b0b19da160ba5cfb`。feat(db): EMR-446/466 休業・lab mapping操作の原対象anchorとmapping版保存（006）
- [PR #846](https://github.com/MinoruSoga/AnimalEkarte/pull/846): merged、`77c28fbfbafa257dcca2421af99ff72da6889717`。docs: 限定監査19票と生前入力・医師確認の設計作業案を整理

PR596は旧フロー、631は未対応件数、784/798/805は契約・選定、806はAPI先行収録、807は保存/model subset、846はdocsのみ。いずれも現在のSTG実行や残接続完了の証拠としない。

<a id="lab-mapping-handoff"></a>
## 有限な引渡し順

1. 297＋549/543: nullable JSON callerを元のNULL/JSONnull/保存所有権を維持して閉じる。
2. 297-A＋447（445/446/466の採用schemaを再利用）: sample原本先行保存、confirm current gate/CAS、全writer mapping head/revision、限定result lookupを接続。全writer前reader非活性を保持。
3. 297-B＋268: decoded job適用版/明示resolve、wait record identity/旧pet-onlyの安全扱いを接続。360の件数更新を再実装しない。516にはexam ID/版＋clinic/pet/record/readinessを渡す。501とunknown date境界を閉じる。
4. 297: draft修正/確定/停止の専用caller存在・採用SHA・結合fixtureを確認。manual master編集を代用品にしない。
5. 268/402/113と297: 実医院・機器個体・承認build・臨床mapping/保存再表示の受入は別段階。保存済fixtureを使う技術残件を実機再取得待ちへ一律戻さない。

<a id="lab-mapping-retrieval-limits"></a>
## 取得限界

GitHub search indexは171b565を返したため探索のみに使い、正の実装証拠は77c28fb固定fetch。api.yaml通常fetchは空contentだったが返されたblob SHA c29a016a49aa8fd1c1c221209aa4a746e5b11a32をfetch_blobで取得し照合。新環境や秘密情報を取得せず、試験を実行していない。
