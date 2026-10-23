# 移行①限定監査：EMR-221/338/342/343/345/351/374

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-migration-stage1-20261023 |
| 種別 | 固定mainの限定静的監査受領記録 |
| 決定状態 | 既決仕様を継承。新しい業務・技術契約を採択しない |
| 実装状態 | 下記の限定接点の分類。全FR/AC・全source監査完了ではない |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

本稿は提供された監査報告を整形したもので、文書編集時に全sourceを独立再監査した記録ではない。[今回の索引](emr-audit-ready-handoff-index-20261023.md#audit-ready-additional-scope)を参照。

基準main: `77c28fbfbafa257dcca2421af99ff72da6889717`（GitHubで再確認）。2026-10-23。

<a id="migration-stage1-conclusions"></a>
## 結論
既存設計と旧profile consumerの実装証拠は存在するが、7票全体・現19CSV契約への接続完了ではない。正式移行HOLDを維持。新規重複票は不要。実データ未受領/未検証をコード欠陥として数えない。

特に374の古い「日時3項目未実装」は、後続PR701の採用sourceと一致しない。target PostgreSQL解釈・順序・一意/範囲衝突とfixtureは存在する。DBなしsource preflightの非空検査、target内検証、作者の限定PG試験、現19実受入を区別する。

<a id="migration-stage1-read-scope"></a>
## 読取範囲と限界
- Plane最新本文7票、全83コメント（221=9/338=6/342=7/343=21/345=12/351=7/374=21）、全件next_page_results=false。
- 採用PR624/625/627/655/656/701のmetadata・本文・diffを取得。PR全review-thread/全discussionの再監査は未実施。
- 固定mainの設計/移行資料、profile/CLI、contract、値validator、参照、resume、予約日時検査と代表fixture、最終001を限定照合。取得失敗した旧pathは移設先へ解決した報告を受領。本記録は以下の固定source根拠を参照。
- tests/build/codegen/DB/Docker/browser/import、SQL変更、old_db実コード・実DB・実bundle再検証はNOT_RUN。実データ生成/外部投稿/新PRなし。全現行API/writer、全source/全ACの網羅保証なし。

<a id="migration-stage1-profile-boundary"></a>
## 19と旧21/候補22の境界
[現正本](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/design/domains/accounting/legacy-finance-evidence.md#legacy-evidence-open-nested-types)は19CSV維持、payments/payment_splits復活禁止、通常finalize/allocate生成禁止、旧21receiptの現19適合転用禁止を明記。
[profile選択](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/csvimport/cutover_profiles.go#L13)と[csv-import CLI](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/cmd/csv-import/main.go)はlegacy21既定/候補22明示で、table数から推定しない。したがってpayments文字列の存在だけを新欠陥としない。これは旧profile実経路で、19へ接続済みの根拠ではない。
最終[DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/migrations/001_init.sql#L10886)では旧billings列とpayments/splitsを退役。既存[364接続契約](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-364-375-376-csv-column-contract-20261021.md#L68)にもtarget adapter残が明記される。547は別sidecar/原値保全で、通常業務の金銭事実を推定生成して埋めない。

<a id="migration-stage1-ac"></a>
## AC対応

<a id="migration-stage1-221"></a>
### EMR-221
- **欠損152・競合1・paygraph別集合、税未確定を維持するreview設計** — DESIGN_PRESENT
  - 採用/現source: docs/work/emr/emr-221-import-contract-design/DESIGN.md:9-69; main統合16e9ce9a7のPlane記録
  - 残り/境界: 過去集計の再実測なし。欠損集合完全性、税根拠、独立provenance、実受入は依存。
- **version付きpayloadと否定17ケース** — DESIGN_PRESENT
  - 採用/現source: 同DESIGN.mdの契約案/T01–T17/後続節
  - 残り/境界: 規範measurement/decision/authorization schema、registry/署名policy、正規化/hash/cardinality、T05/06/10/14独立fixture未完。runtime NOT_RUN。
- **現19 CSVと別sidecarの正式受入/製品閲覧接続** — DEPENDENCY
  - 採用/現source: DESIGN.md:17-21; docs/design/domains/accounting/legacy-finance-evidence.md:15-17,44-65
  - 残り/境界: 547の6入れ子型、verification_ref、stable UUID対応、上限、実source profile/artifact binding/receiptが未閉鎖。原値保全は通常finalize/allocateではない。

<a id="migration-stage1-338"></a>
### EMR-338
- **移行される/医院登録/不要一覧、MIG25とmaster網羅** — DESIGN_PRESENT
  - 採用/現source: PR624/627採用; docs/delivery/MIGRATION_SCOPE.md:1-95
  - 残り/境界: 第2版中間資料として存在。確定は方針で実移行ではない。現在19正式契約の引渡し版としては未整合/未受入。
- **Q1薬剤取り込み・Q2確定分先渡しと設計票** — DECISION_RECORDED
  - 採用/現source: Plane 2026-10-11回答f3ce081d; EMR342とPR625
  - 残り/境界: 既決を再質問しない。医院へ実際に配布した証跡は今回未確認。
- **単価/税/除外行/写真等の最終内容確定と追補** — DEPENDENCY
  - 採用/現source: MIGRATION_SCOPE.md確定待ち行/更新予定/開発側参照
  - 残り/境界: 349/MIG41の同一候補receipt、344の曖昧分類を分離。旧件数・旧producer報告を現bundle確認済みにしない。

<a id="migration-stage1-342"></a>
### EMR-342
- **medicines列/旧列、61分類、9分類1456、剤形/単位、参照付替え、件数突合** — DESIGN_PRESENT
  - 採用/現source: PR625 merged dc96b9150; docs/work/emr/emr-342-medicines-contract-design/DESIGN.md
  - 残り/境界: 9分類1456は設計当時の根拠。現データ実測はNOT_RUN。数量/単位不明を通常業務値へ推定補完しない。
- **Q1/Q2記録・旧21拒否の適用範囲・candidate fresh** — DECISION_RECORDED
  - 採用/現source: DESIGN.md:9-13; Plane342の決定; PR655
  - 残り/境界: 旧21維持/候補22分離という歴史的適用範囲。現19を21へ戻す承認ではない。
- **producerと列順/ID帯/親子/hash一致** — DEPENDENCY
  - 採用/現source: 同設計; cutover_profiles.go:13-96; runbook:62-67
  - 残り/境界: MIG34/37/41の同一profile/version/full-file SHA/4院receipt未受領。設計採用≠正式activation。

<a id="migration-stage1-343"></a>
### EMR-343
- **V6:表数/旧bundle/medicines欠落/inventory/医院/名称重複拒否** — SOURCE_PRESENT
  - 採用/現source: cutover_medicines_contract_test.go:18-108,222-355; cutover_profiles_test.go; cutover_reference_graph.go
  - 残り/境界: 旧legacy21/candidate22内のfixture存在と期待値。現19のfixture/実bundle合格は証明しない。
- **medicines15列/COPY/FKとestimate薬剤参照** — SOURCE_PRESENT
  - 採用/現source: cutover_contract.go; cutover_reference_graph.go; cutover_medicines_contract_test.go:51-108,266-322
  - 残り/境界: producerによる9分類抽出とprocedure除外/付替えの実成果は依存。
- **明示candidate rehearsal/fresh-only・正式activation拒否** — SOURCE_PRESENT
  - 採用/現source: cutover_profiles.go:52-96; cutover_import_resume.go:182-201; cutover_profiles_test.go
  - 残り/境界: CLI無指定legacy21。候補22はREHEARSAL_ONLYのみ。現19正式入口への変換/有効化ではない。
- **合成22bundle通過/go test/vet・両側hash/4院一致** — REPORTED_HISTORICAL_ONLY
  - 採用/現source: PR655 head c99758545f46b7dfa4937d6401dba57fd694f6ca, merge d4beff6cc16cf9b47fabd3d9cf61b6e72a856dd5
  - 残り/境界: 作者限定PASSとDB SKIPを分離。今回tests/DB NOT_RUN。producer正式一致とfresh rehearsal実受入は依存。

<a id="migration-stage1-345"></a>
### EMR-345
- **4master負価格/不正整数拒否、0/nullable空の区別** — SOURCE_PRESENT
  - 採用/現source: cutover_master_values.go:9-28; cutover_contract_validate.go:706-710; cutover_contract.go cutoverMasterPriceColumn
  - 残り/境界: merchandise空は拒否、他nullable空は保持、0許可。旧本文の全表空許可は失効。365 decimal新profileは別。
- **診断は表/列/行、履歴billing/estimate負額を同一master規則で拒否しない** — SOURCE_PRESENT
  - 採用/現source: cutover_master_values.go:13-26; cutover_contract.go cutoverMasterPriceColumn
  - 残り/境界: master switchにbilling_items/estimate_itemsなしは確認。全履歴負額の正式受入承認を意味しない。fixture実行NOT_RUN。
- **負bundle拒否/非負bundle合格とproducer負値修正** — REPORTED_HISTORICAL_ONLY
  - 採用/現source: PR655; cutover_latest_review_test.go:15-77（master BIGINT target）
  - 残り/境界: 旧候補限定検証の報告。MIG40/41の同一4bundle、空価格storage整合、19正式profile受入が未完。

<a id="migration-stage1-351"></a>
### EMR-351
- **全列matrix/API-preflight-storage差の分類** — HISTORICAL_AUDIT_PRESENT
  - 採用/現source: PR656; docs/work/emr-351-api-csv-audit/AUDIT.md:141-172; column-matrix.csv
  - 残り/境界: ed04c0237の21表236列のみ。OK119/EMR38/MIG44/HIST2/OWNER18/UNK15を現19の分類数へ転用禁止。
- **差の実装/producer/保持/判断/未知のowner分割** — DESIGN_PRESENT
  - 採用/現source: AUDIT.md:174-194; docs/work/emr/emr-364-375-376-csv-column-contract-20261021.md:122-157
  - 残り/境界: 374=DDL floor、375=通常編集保持、376=履歴/既定/未知、364=税transport、365=decimal、547=原値保全。重複実装/重複起票不要。
- **現main列差分と最終DDLへの更新** — PARTIAL_STATIC_DELTA
  - 採用/現source: 001_init.sql:10886-10919; cutover_contract.go; cutover_profiles.go; legacy-finance-evidence.md:15,51
  - 残り/境界: 旧billings列とpayments/splitsがconsumerに残るが旧profile由来。現19用target adapter/原値sidecar未接続。全API/writer/現19全列再監査は未実施。

<a id="migration-stage1-374"></a>
### EMR-374
- **NULL/0・enum・数値上限・quantity/vital/kana** — FIXTURE_MAPPED
  - 採用/現source: cutover_ddl_preflight_test.go:41-253; cutover_ddl_scalar_test.go:70-270; cutover_ddl_public_scalar_test.go:10-206
  - 残り/境界: fixture本文を静的対応。runtime NOT_RUN。NaNなどは旧PG CHECK意味の再現で臨床推奨値ではない。現19全型保証ではない。
- **部分unique除外・同clinic FKと誤親の区別** — FIXTURE_MAPPED
  - 採用/現source: cutover_ddl_preflight_test.go:254-411,504-622
  - 残り/境界: inactive merchandise/空contact除外、番号空参加、親missing/band/record-pet mismatch区別。現在owner同値を要求しない。
- **歴史snapshot/inactive actorを拒否しない** — FIXTURE_MAPPED
  - 採用/現source: cutover_ddl_preflight_test.go:413-502,624-635
  - 残り/境界: 4表独立owner snapshot/移管後/古いtentative ownerとinactive staff/deceased owner/petの正例。通常操作権限への昇格ではない。
- **PII非含有・匿名診断** — PARTIAL_FIXTURE_MAPPED
  - 採用/現source: cutover_ddl_preflight_test.go:64-72,534-545; cutover_master_values.go:13-26
  - 残り/境界: quantity実値漏れ拒否assertと固定table/column/row文言確認。全error経路・全PII変種を網羅したとはしない。
- **timestamp構文/順序/開始一意/範囲衝突** — SOURCE_AND_FIXTURE_PRESENT
  - 採用/現source: cutover_appointment_time_preflight.go:20-35,97-298; cutover_import_resume.go:210,259-269; cutover_appointment_time_preflight_test.go:153-307,842-977
  - 残り/境界: 旧未実装記述は最新PR701で解消した範囲あり。DBなしsource preflightは非空検査のみ。target PG解釈は実COPYと同TXで再検証。現target受入はNOT_RUN。
- **限定試験/reviewと実DB受入の分離** — REPORTED_HISTORICAL_ONLY
  - 採用/現source: PR701 head6c42dab70 mergef69593ab9; Plane comment adb5ef7b 2026-10-18
  - 残り/境界: 312 PASS/4SKIP、予約限定PG132ケースは作者報告。4SKIPや全migration/FK/RLS/19実importをPASSにしない。今回実行NOT_RUN。

<a id="migration-stage1-pr-adoption"></a>
## 採用PR確認
- [#624](https://github.com/MinoruSoga/AnimalEkarte/pull/624): merged=True; head `5d879392b520170c5d10de3dfcc7ac279275a4c1`; merge `3c635018b303796e21ff4bcc10dc1bb9769c0c20`
- [#625](https://github.com/MinoruSoga/AnimalEkarte/pull/625): merged=True; head `a1cf792f5561786d20e797c520e7595583d6ec15`; merge `dc96b9150a95b862a77993c59858c614df1e1c05`
- [#627](https://github.com/MinoruSoga/AnimalEkarte/pull/627): merged=True; head `276f9b1bc8c0e47b8209723d92c2b109252e520e`; merge `8ca80c12dbfac7580351fbc3854f4b356f5ed1d7`
- [#655](https://github.com/MinoruSoga/AnimalEkarte/pull/655): merged=True; head `c99758545f46b7dfa4937d6401dba57fd694f6ca`; merge `d4beff6cc16cf9b47fabd3d9cf61b6e72a856dd5`
- [#656](https://github.com/MinoruSoga/AnimalEkarte/pull/656): merged=True; head `269a4f41a9c758ce16c85322d137adb8fdefc710`; merge `a2fb9e7c41877fa123b08cddf6e601a976592d8d`
- [#701](https://github.com/MinoruSoga/AnimalEkarte/pull/701): merged=True; head `6c42dab70e9d8838bbba7c7865fd2d5702ba7a3b`; merge `f69593ab9a9fc4d937d21df365dda29d713eee06`
