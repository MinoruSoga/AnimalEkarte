# EMR-221: 欠損除外と税未確定を保持する取込契約案

作成: 2026-10-02 / 設計担当: Codex / 状態: レビュー用提案、実装未着手。

対象は八王子の producer → AnimalEkarte 間の契約設計と合成検証仕様。
現行 importer、CSV、manifest、税判断、DB は変更しない。
[Plane EMR-221](https://app.plane.so/baritechllc/projects/6a009324-8a8a-455c-b16a-41b3f6207b0e/work-items/d40d6e69-d1a3-4e72-836d-0eba389c5521/) を進捗の正本とする。

## 採用する方針

欠損を完全性 PASS に変換せず、未解決の税を 0 円や城東の解釈で埋めない。
既存 v1 の formal 判定を維持し、将来の v2 では「原本の完全性」と「検証済み部分集合の受入適格性」を別々に判定する。
v2 を導入しても、現在の八王子 bundle は正式取込不可のままである。

- producer: `old_db` の HAC-CSV-1 / HAC-PIPELINE-1。原本・除外・税・CSV の根拠と再生成を所有する。
- consumer: AnimalEkarte の EMR-221。独立検証、対象医院への束縛、取込と監査を所有する。
- 契約の技術判断: Codex。producer と consumer の独立レビューを経て実装する。顧客の受入れ・現地観察・本番実行を代理したと記録しない。
- 成果の目的: 件数一致だけで正式化する誤りと、医院や revision をまたぐ証跡流用を防ぐ。確認質問や転記を増やす代わりに、同じ機械契約を両側で検証する。
- 成功指標: 下記の必須負例をすべて DB 書込前に拒否し、拒否 receipt に診療データを含めない。現状八王子が拒否されることも成功条件。

## 確認した証拠と範囲

確認時 HEAD: AnimalEkarte `fbc96c70765036fb9e1251b4cda5e18b36d02127`、old_db `3846d9a5e887f61c280984a382d121e841bd48fd`。
`old_db/...` は sibling repository 内のパス。行データ・BAK は読まず、コード、判断記録、集計 JSON を参照した。

| 証跡 | 集計・状態 | この証跡だけでは証明しないこと |
| --- | --- | --- |
| `old_db/tmp/producer-iso-generated/local-postgres-knjo-reacquire-load-summary.json`、生成 2026-09-26T13:58:19.368Z、SHA-256 `2f3b334a70f6c79425931f6b64e1c87410d89ddbdf998c648debc32e7a6f3b17` | expected 477,603、readable distinct 477,451、loaded distinct 477,450、conflictingPrimaryKeyGroups 1、stillMissingRows 153、PARTIAL、promotion SKIPPED | 正式 source table への昇格、152 件の完全なキー一覧、CSV 再生成 |
| 同 summary の区別 | 477,603 − 477,451 = **152**（読めない分）、477,451 − 477,450 = **1**（競合キー群）。計 153 は投入との差分 | 競合 1 群を「破損152件の承認」で受入れたこと。chunk occurrence 477,456 は distinct 件数と別単位 |
| AE handoff `manifest.json`、生成 2026-09-22T04:09:34.216Z、stageBuildId `fe9f228b-bd3f-495a-aebc-ef3c4d8128b3`、SHA-256 `8de0db913db296f9de0d94f076f91936b6d206f45bad72ca90428e981944fffd` | `paygraph-20260922-01` の REHEARSAL_ONLY、PARTIAL、sourceComplete=false、sourceProvenanceVerified=false | 9/26 再測定との同一 snapshot、正式 handoff |
| `old_db/docs/by-clinic/hachioji.md` の paygraph 記録 | completed・nonzero の支払グラフなし **46,118 billings** を隔離 | 上記 152 source keys / 1 conflict group と同じ集合、重複の有無、同じ承認の適用 |
| `old_db/todo.md` の 9/26 判断 | 9/21 BAK FILE=5 の読める行を使い、読めない152件は除外。健全BAK再受領を着手条件にしない | 顧客署名、競合・支払グラフの追加除外受入れ、完全性 PASS、本番 GO |

run 名 `hachioji-intake-20260921-01` が一致しても revision / stage build / artifact digest が異なれば同じ測定とは扱わない。
既知キーの部分リストを完全な152件リストとして扱わない。missing key set を再構成できない場合は `null` と未確定を保持し、正式適格性を拒否する。
新しい証拠を作るときは原本・旧 artifact を保存し、別 revision として固定する。

## 税の現在の流れと変更責務

| 境界 | 現在の実装・根拠 | 設計上の要求 |
| --- | --- | --- |
| 医院別判断 | `old_db/config/migration/clinic-mapping-decisions/hachioji.json` の Z5 は blocked、`jkzei_neutral_raw_no_tax_derivation` | 未確定を保持する。城東の判断をコピーしない |
| producer 共通判断 gate | `old_db/docs/migration/accounting-tax-semantics-decision.json` の approvedClinicCodes は jouto のみ。`scripts/lib/accounting-csv-promotion-gate.mjs` が clinic / Z5 / meaning / evidence / productionEligible を照合 | 八王子の根拠付き判断と実装一致ができるまで formal accounting CSV を拒否 |
| canonical | `old_db/sql/migration/020_canonical.sql` の payment_settlements: Jkzei_Kin → taxable_base_raw、JHOS_TAX + JSAN_TAX → tax_amount_raw（NULL/空は各0への COALESCE） | 八王子にこの意味・NULL処理が妥当という証拠は未確定。neutral raw は保存可能でも正しい課税額の証拠にはならない |
| stage → CSV | `old_db/sql/migration/030_stage.sql`: subtotal=round(claim_amount − tax_amount)、tax_total=round(tax_amount)、total_amount=親 billing。21表の payments / estimates に税列がある | 税率期間、NULL/0、値引き、返金・負額、端数、合算単位を同一 run で検証し decision digest と結ぶ。payments の確認を estimates の証拠に流用しない |
| AE 事前検証 | [cutover_contract_validate.go](../../../backend/internal/csvimport/cutover_contract_validate.go)、[cutover_payment_graph.go](../../../backend/internal/csvimport/cutover_payment_graph.go) | v1 は TRUSTED_CANDIDATE / source complete PASS / verified provenance / incomplete tables 空が必須。payment の整数・親子等のチェックは税の業務意味の証明ではない |
| AE 永続化 | [cutover_contract.go](../../../backend/internal/csvimport/cutover_contract.go) の列契約 → [cutover_import_copy.go](../../../backend/internal/csvimport/cutover_import_copy.go) の COPY | 取込時に税を推測して修復しない。semantic preflight 未完了なら COPY 前に停止 |

今回の結論は「八王子の税は未確定なので正式出力を止める」。追加の人の選択回答は、この設計・合成テスト作成の開始条件ではない。

## 契約の最小変更案

形状の正本は [accepted-exclusions.schema.json](accepted-exclusions.schema.json)。これは **提案中の追加 payload** の JSON Schema で、稼働中 manifest の置換ではない。schema 単独の成功は semantic PASS / 受入承認ではない。

1. 現行 `animalekarte-cutover-v1` と formal mode は変更しない。未知 version を v1 として解釈する fallback は作らない。
2. 将来 `animalekarte-cutover-v2` を別 parser / 明示 mode `formal-accepted-exclusions` で導入する。既存21表、clinic band、FK、金額、source identity、CSV hash、mapping binding、snapshot lock の検査を同等以上に維持する。v2 全体 schema は実装時に既存manifest契約とこの追加 payload を統合し、両repoが同一hashの正本から派生する。
3. v2 manifest の `acceptedExclusions` に本 payload を含める。payload は最終manifest自身のhashを含めない（循環hash禁止）。独立した authorization receipt が、payloadを含む最終 manifest bytes の SHA-256、target、policy version に結び付く。operator が選ぶ expected digest は producer の信頼済み報告から取得し、import directory 自身から自己承認しない。
4. sourceComplete=false / sourceCompletenessStatus=PARTIAL / incompleteSourceTables は事実のまま残す。受入適格性のみ新しい `ACCEPTED_WITH_EXCLUSIONS` とし、TRUSTED_CANDIDATE や PASS へ改名しない。provenance は完全性と独立に検証するが false のまま受理しない。
5. 追加 payload は同じ clinic/ordinal、run、revision、stageBuildId、原本hash/size、backup set position、対象CSV集合digestに束縛する。個別証拠の `ref` はローカル証拠レジストリIDであり、任意URL/pathの自動取得命令ではない。参照先bytesのhashと同一identityを独立に再検証する。
6. reason は SOURCE_UNREADABLE / CONFLICTING_PRIMARY_KEY / PAYMENT_GRAPH_QUARANTINE を区別。各 scope の母集団・単位・population/retained/excluded count・set digest・measurement ref を保持し、population=retained+excluded を同一scope内で検証する。異なる単位やsnapshotの件数を合算しない。scopeIdは一意とし、同じキーを複数reasonで二重控除しない。COUNTのみの説明は調査用であり正式適格性を満たさない。
7. 各除外setのclosureを求め、owner→pet→record→billing→paymentの複合キーで親子整合と影響範囲を証明する。base key set / surviving key set / quarantine set の照合で漏れ・重複・再導入を拒否する。PK不明の欠損に架空キー・空集合digestを作らない。
8. decision receipt は範囲ごとの技術判断と委任根拠、決定者、独立検証者、理由、時刻、証拠hashを保存する。エージェント判断は actor=agent と明記し、human/customer署名に見せない。権限は別の信頼済みpolicyと照合し、payloadの自己申告だけでは成立しない。152件の既存判断はscopeを機械照合できる形へ記録するが、競合1件やpaygraphへ拡張しない。
9. tax は UNRESOLVED または VERIFIED。未確定は保存・調査に有効だが正式適格性を拒否。VERIFIEDでも decision / semantic test receipt が clinic・run・mapping・CSVと一致し、期間/端数/NULL等の検証を満たさなければ拒否する。
10. 調査上の decision と実取込 authorization は別 receipt。失効・期限切れ・target/clinic/run不一致・再利用済み authorization を拒否する。検証後から apply までの差替えを排他とdigest再確認で拒否し、再開は同じ immutable evidence に対する新しい許可を必要とする。

### 正式判定の順序と監査

`strict decode → version/policy → identity/digests → provenance → counts/sets/closure → decision authority → tax → 既存21表/FK/payment検査 → exact target authorization → lock/recheck → transactional apply → verify/commit`

途中で失敗したら DB 書込なし。apply中の失敗は現行 transaction の rollback を維持する。既存CLIの [監査出力](../../../backend/internal/csvimport/cutover_import.go) `/migration-reports` に rejection code、clinic/run、schema/policy version、evidence digest、件数、actor種別、authorization ref、transaction outcome を残す。識別子・行値・医療情報・secretは出力しない。
永続的な取込状態は既存監査receiptと同一operation IDで束縛し、成功と報告する前に耐久receiptを確認する。監査書込不能は開始前に停止。commit後のreceipt失敗は UNKNOWN/要照合として再実行を止め、成功を偽装しない。実装で audit_logs を用いるなら同一txで保存し、receiptを代替sinkとするなら現行runbookに例外と回復手順を明記する。

提案する拒否code: `UNSUPPORTED_CONTRACT`, `IDENTITY_MISMATCH`, `PROVENANCE_UNVERIFIED`, `EVIDENCE_MISMATCH`, `EXCLUSION_SET_UNKNOWN`, `COUNT_MISMATCH`, `EXCLUSION_SCOPE_UNAPPROVED`, `TAX_UNRESOLVED`, `TAX_EVIDENCE_INVALID`, `AUTHORIZATION_INVALID`, `AUDIT_UNAVAILABLE`。生のキーやdriver errorはcodeに含めない。

## 合成検証仕様（未実行）

実医院データを使わず clinic `synthetic-a` / `synthetic-b`、run `synthetic-run-1` / `synthetic-run-2` と生成キーのみで fixture を作る。positive の最小母集団は expected=10 / readable=8 / loaded=7、SOURCE_UNREADABLE=2、CONFLICTING_PRIMARY_KEY=1。別 billing 母集団に quarantine=1 を置き、全setとclosureを生成する。receiptはtest専用trust storeで検証し、本番権限を持たせない。

| ID | 合成条件 | 必須結果 |
| --- | --- | --- |
| T01 | v1 の既存 complete/trusted fixture、v1 の partial fixture | 前者の既存挙動を維持、後者は従来通り拒否 |
| T02 | v2、全証拠/承認/税/target有効、正しいsubset | preflight適格。PARTIALの事実を保持。dry-runで書込0。適格判定だけでapplyしない |
| T03 | v2を旧consumerへ渡す、未知version、mode省略でv1へfallbackを試す | UNSUPPORTED_CONTRACT。リハーサル指定でformalへ迂回不可 |
| T04 | clinic/run/revision/stageBuildId/原本hashの1項目だけ差替え | IDENTITY_MISMATCH、COPY呼出0 |
| T05 | expected/readable/loadedを1ずらす、occurrenceをdistinctに代入 | COUNT_MISMATCH |
| T06 | conflict1のscopeを削除、unreadable承認だけで追加除外を認可 | COUNT_MISMATCH または EXCLUSION_SCOPE_UNAPPROVED |
| T07 | keySet=null、部分リスト、空集合hash、同数の別key集合 | EXCLUSION_SET_UNKNOWN または EVIDENCE_MISMATCH |
| T08 | paygraphだけ別revision、source152のreceiptをquarantineへ流用 | IDENTITY_MISMATCH または EXCLUSION_SCOPE_UNAPPROVED |
| T09 | 除外親を参照する子、複合キーの一部だけ一致、除外済み行の再導入 | closure/FK検証で拒否、COPY呼出0 |
| T10 | decisionなし、自己申告actor、未委任agent、agentをhumanに改変、失効receipt | EXCLUSION_SCOPE_UNAPPROVED |
| T11 | tax=UNRESOLVED、decisionだけVERIFIED、joutoのreceiptを流用 | TAX_UNRESOLVED / TAX_EVIDENCE_INVALID |
| T12 | NULLを0にした税、期間違い、値引き/負額/端数/合算単位の誤り、estimates未検証 | semantic test失敗。税差額を取込側で補正しない |
| T13 | source provenance=false、mapping/CSV/table set digest差替え | PROVENANCE_UNVERIFIED または EVIDENCE_MISMATCH |
| T14 | 他医院・他環境target、期限切れ・消費済み・digest不一致の実行許可 | AUTHORIZATION_INVALID。技術decisionを実行許可として使わない |
| T15 | duplicate JSON key、未知field、非整数count、remote/traversal ref、巨大payload | strict decode/size/ローカルref制約で拒否。ネットワークアクセス0 |
| T16 | preflight後のCSV/receipt差替え、途中COPY失敗 | apply前拒否またはrollback。部分commitなし |
| T17 | 監査sink不可、commit後receipt失敗、同じauthorizationでretry | 開始前停止 / UNKNOWNで停止 / 再実行拒否。ログに行値なし |

schemaの形式検証と上記semantic判定は別テストとする。producerの正例だけでconsumer PASSとせず、双方が同じfixtureを読む。DB rollbackの実証は合成データ専用の隔離環境で後続実装時に行う。

## 後続作業と現在の完了範囲

1. **次に着手可 / Codex / 追加の人の回答不要**: measurement・decision・authorization receiptの規範schema、信頼済みregistry/署名検証policy、集合とCSV集合の正規化・hash・cardinality・重複排除規則を確定する。今回のschemaは追加payloadの形状案までであり、参照先契約は未確定。T05/06/10/14を両repoで一意に判定できるfixtureと独立レビューを完了するまで、正式受入経路の実装へ進めない。
2. **上記確定後の実装候補 / Codex**: producer/consumer両側でT01–T17の負例を先に追加し、既存v1回帰を固定してから、immutable evidence binding・strict decoder・独立検証を実装する。old_db側は別claim/worktree、MIG-19/20/HAC-CSV-1と調整し、手編集したhashで整合させない。これは今回のdesign-only作業の実装実績や実行許可を意味しない。
3. **調査継続 / Codex**: 152件の集合完全性、競合1件の明示判断、paygraphの独立集合・closure、八王子税の証拠を確定。証拠不足時の結果は未確定/非importとし、健全BAK受領を作業全体の待ち条件に戻さない。
4. **別ゲート**: 同一revisionから21CSV再生成、producer/consumer検証、実医院の受入れ、exact targetを固定したSTG/F6/本番の実行許可。現時点では未達。

本タスクの成果は設計・契約形状案・合成検証仕様。schemaチェックは本番取込の検証ではない。EMR-46の正式bundle待ちや旧データ移行全体を完了扱いにしない。
