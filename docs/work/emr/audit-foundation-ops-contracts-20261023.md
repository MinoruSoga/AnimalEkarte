# 基盤・運用・価格契約 6票の限定静的監査

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.audit-foundation-ops-contracts-20261023 |
| 種別 | 固定mainの限定静的監査受領記録 |
| 決定状態 | 既決仕様を継承。新しい業務・技術契約を採択しない |
| 実装状態 | 下記の限定接点の分類。全FR/AC・全source監査完了ではない |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

本稿は提供された監査報告を整形したもので、文書編集時に全sourceを独立再監査した記録ではない。[今回の索引](emr-audit-ready-handoff-index-20261023.md#audit-ready-additional-scope)を参照。

基準main: `77c28fbfbafa257dcca2421af99ff72da6889717`。最新mainを直接取得。最新Plane本文と全コメント 234=7 / 292=5 / 329=2 / 341=5 / 365=12 / 369=1、各next_page_results=false。

結論: 234/292/329/341は既存成果を再実装しない。365は採択済decimal契約の実装接続が未完。369は既決回答の受入で、税業務の再判断をしない。実装・設計・過去報告・運用実績・NOT_RUNを別記する。

読取のみ。製品コード変更・新PR・Plane投稿・新grant・DB・テスト・build・codegen・実環境・設定変更なし。隔離環境を新規gateにしない。

<a id="foundation-ops-contracts-234"></a>
## EMR-234 — 新規医院の執行に検査確定解除を初期付与

初期付与subset実装済み。新規医院実受入未確認、解除guardの現型接続は別残差
Plane: https://app.plane.so/baritechllc/browse/EMR-234/ / 本文更新 2026-10-17T22:58:24.972543Z / 全コメント 7 / 最新 3b70db84-66d7-4d14-910e-e1a8f3790e3e

<a id="foundation-ops-contracts-234-ac"></a>
### AC別判定
- 新規医院執行editのみ / 一般全deny / create-deleteなし: sourceとfixtureあり
- 既存医院・seed非遡及: seed parity fixtureあり、既存全医院の運用証明とは別
- 理由1〜500・専用権限・旧正式版なし拒否: sourceあり
- 親カルテ確定guard: 呼出しは残るがsharedkernel旧Status参照を現modelへ接続する依存あり

<a id="foundation-ops-contracts-234-adopted-subset"></a>
### 実装済み/採択済みsubset
- clinic_service.go:139-142のexec edit例外。CreateClinic default rulesを通すfixtureあり
- UnconfirmはTrimSpace、rune上限500、confirmedかつCurrentRevisionVersion必須、同TX revision/audit。専用resource routeを保持

<a id="foundation-ops-contracts-234-unconnected"></a>
### 未接続・未確認
- 新規医院provisioning受入は最新UATコメントでも未確認。既存医院のrule無しだけで新規付与成功を証明しない
- sharedkernel/medical_record_lock.go:36のparent.Statusは現在MedicalRecordに無い。解除parent guard全体のruntime成立を保証不可。これは初期grant欠落ではない

<a id="foundation-ops-contracts-234-dependencies"></a>
### 依存（全票Done待ちへ拡張しない）
- EMR-495/543/549の現lifecycle/sharedkernel adapter必要subset。旧finalized拒否の最新裁定を234独自で変更しない

<a id="foundation-ops-contracts-234-fixture-boundary"></a>
### fixture・採用履歴・実行境界
- ec3853247c496df6e2c267bc6814e9cf14edd122を直接取得。PR番号は特定できず、存在を捏造しない
- TestService_CreateClinic_DefaultPermissionGroupRules / TestDefaultPermissionRuleTable_CoversAllResources / TestDemoSeedGroupRules_Parity。今回NOT_RUN
- 2026-10-20コメント3b70db84は既存医院の非付与観測、新規provisioning UNKNOWN

後続owner: 234初期付与の受入担当/新規医院作成可能な管理者。共通guardは495/543/549、検査revision ownerへ有限handoff

<a id="foundation-ops-contracts-234-source-evidence"></a>
### 固定source根拠
- [backend/internal/clinic/clinic_service.go:139-142](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/clinic/clinic_service.go#L139)
- [backend/internal/clinic/clinic_service_test.go:489-600](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/clinic/clinic_service_test.go#L489)
- [backend/internal/clinic/clinic_service_test.go:612-777](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/clinic/clinic_service_test.go#L612)
- [backend/internal/medicalrecord/examination_revision_service.go:107-148](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/examination_revision_service.go#L107)
- [backend/internal/medicalrecord/examination_revision_service.go:317-332](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/examination_revision_service.go#L317)
- [backend/internal/sharedkernel/medical_record_lock.go:25-39](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/sharedkernel/medical_record_lock.go#L25)
- [backend/internal/model/medical_record.go:1-80](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/model/medical_record.go#L1)
- [backend/internal/medicalrecord/routes_records.go:1](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/routes_records.go#L1)

<a id="foundation-ops-contracts-292"></a>
## EMR-292 — 認可キャッシュ無効化後の旧インスタンス退出確認

コード/過去配備成功あり。全旧instance退出の運用受入は未確認
Plane: https://app.plane.so/baritechllc/browse/EMR-292/ / 本文更新 2026-10-17T23:01:29.747818Z / 全コメント 5 / 最新 38ff29ed-cc16-45ae-9f31-79c4e2d1c38e

<a id="foundation-ops-contracts-292-ac"></a>
### AC別判定
- 正TTLをSTG設定から除く: sourceあり
- 修正後再deploy: 既存runのjob/step success再読
- 旧構成instance退出・現実効revision: 未確認

<a id="foundation-ops-contracts-292-adopted-subset"></a>
### 実装済み/採択済みsubset
- composition_auth.go:159-181は正TTLのみcache接続。wranglerにTTL宣言なし
- staging_access_cache_config_testとcomposition_auth_cache_contract_testのfixtureあり

<a id="foundation-ops-contracts-292-unconnected"></a>
### 未接続・未確認
- 現在のinstance一覧/実効環境/退出receiptを読んでいない。Actions successや経過時間を退出の証明にしない

<a id="foundation-ops-contracts-292-dependencies"></a>
### 依存（全票Done待ちへ拡張しない）
- 運用担当による既存許可範囲の読取観測と受入receipt。今回再deployしない。現状態を読取確認し、旧構成が残存する場合は必要作業を別途判断する

<a id="foundation-ops-contracts-292-fixture-boundary"></a>
### fixture・採用履歴・実行境界
- ceac0ccd4f8c1ca144c2356deadd37a105454353を直接取得
- https://github.com/MinoruSoga/AnimalEkarte/actions/runs/37538425729 最新job再読: STG deploy/migrate/health success、production skipped。新実行ではない

後続owner: EMR-292運用受入担当

<a id="foundation-ops-contracts-292-source-evidence"></a>
### 固定source根拠
- [backend/cmd/api/composition_auth.go:159-181](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/cmd/api/composition_auth.go#L159)
- [backend/wrangler.jsonc:1](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/wrangler.jsonc#L1)
- [backend/cmd/api/staging_access_cache_config_test.go:17](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/cmd/api/staging_access_cache_config_test.go#L17)
- [backend/cmd/api/composition_auth_cache_contract_test.go:17-31](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/cmd/api/composition_auth_cache_contract_test.go#L17)

<a id="foundation-ops-contracts-329"></a>
## EMR-329 — 【基盤】docker compose マルチパッケージ go test が共有DB fixture レースで高頻度 flake（GOFLAGS=-p=4）

手順/compose成果あり。過去3連続PASS申告と現在の未実行を分離
Plane: https://app.plane.so/baritechllc/browse/EMR-329/ / 本文更新 2026-10-20T06:21:30.764463Z / 全コメント 2 / 最新 ddf812da-3a6e-4e8f-9b3d-5ec62de094d5

<a id="foundation-ops-contracts-329-ac"></a>
### AC別判定
- 複数DB packageを-p 1に統一: 現文書・Makefile/shardあり
- GOFLAGS=-p=4維持＋注意書き: あり
- 3連続PASS: 10/7コメントに3/3申告。元ログの独立再検証は未実施

<a id="foundation-ops-contracts-329-adopted-subset"></a>
### 実装済み/採択済みsubset
- backend/CLAUDE.md:42、docker-compose.yml:76-79、Makefile:586-593、run-local-ci.sh:324-332に直列規約

<a id="foundation-ops-contracts-329-unconnected"></a>
### 未接続・未確認
- 過去実行のraw stdoutを本監査で取得できていない。3/3は当時comment報告として扱い、current main PASSにはしない
- 対応commit/PR番号は今回特定せず。現source成果を直接照合

<a id="foundation-ops-contracts-329-dependencies"></a>
### 依存（全票Done待ちへ拡張しない）
- 現mainの一般compile残差とは別票。fixture分離/新隔離環境/CI backend試験追加は本票scope外

<a id="foundation-ops-contracts-329-fixture-boundary"></a>
### fixture・採用履歴・実行境界
- 2026-10-07 2f2f5efe-0805-4b94-9c68-1e03d566a76eにEMR290相当-p1付き3/3 PASS記録。-p省略8/8 FAILとの比較は過去報告
- 現在の再実行NOT_RUN。新しい隔離DB gateなし

後続owner: EMR-329基盤手順review担当。将来の本人指定mainローカル試験担当

<a id="foundation-ops-contracts-329-source-evidence"></a>
### 固定source根拠
- [backend/CLAUDE.md:42](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/CLAUDE.md#L42)
- [docker-compose.yml:76-79](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docker-compose.yml#L76)
- [Makefile:586-593](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/Makefile#L586)
- [scripts/ci/run-local-ci.sh:324-332](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/scripts/ci/run-local-ci.sh#L324)

<a id="foundation-ops-contracts-341"></a>
## EMR-341 — 【基盤】本番の初回構築で必要な初期マスタ(標準予約区分ほか)の手順化

手順3成果は静的整合。実投入/医院登録とは別
Plane: https://app.plane.so/baritechllc/browse/EMR-341/ / 本文更新 2026-10-20T06:23:30.041494Z / 全コメント 5 / 最新 cef5e55a-684e-43f3-8465-842eb05d1dfd

<a id="foundation-ops-contracts-341-ac"></a>
### AC別判定
- setup初期master表: 存在
- 標準5区分適用をsetup/GOLIVEに記載: 一致
- 医院への控え依頼をMIGRATION_SCOPEに記載: 存在

<a id="foundation-ops-contracts-341-adopted-subset"></a>
### 実装済み/採択済みsubset
- setup§4.1-4.2、GOLIVE行11、MIGRATION_SCOPE§5.1で5区分/医院別再登録/名称時間料金LINE公開の控えが一致
- SQLはadvisory lock、既存同名skip、全5 active後条件。inactiveを勝手に有効化しない

<a id="foundation-ops-contracts-341-unconnected"></a>
### 未接続・未確認
- 実投入receiptなし。GOLIVE自身も適用証跡なしと明記
- setup表の薬剤『Open decision』はMIGRATION_SCOPEの後日取込設計と古さがある。標準5区分ACを未達へ戻さず、338/343参照を整備する限定文書残差

<a id="foundation-ops-contracts-341-dependencies"></a>
### 依存（全票Done待ちへ拡張しない）
- 348実投入、医院の控え/登録、347外部案内。139preflight→140実admin+loginはsetupに明記。141全完了を新gateにしない

<a id="foundation-ops-contracts-341-fixture-boundary"></a>
### fixture・採用履歴・実行境界
- PR623 merged 7f1590a4f5e078fcf0e35865aa64a06fc92fbbcb、head50925f990。metadata直接再読
- 2026-10-11 Q1/Q2決定758fd15d、全3成果達成acfb59a3。新SQLなし

後続owner: 341手順review担当 / 348実投入 / 医院登録 / 347案内 / 338・343薬剤scope表参照

<a id="foundation-ops-contracts-341-source-evidence"></a>
### 固定source根拠
- [docs/ops/infra/production/setup.md:63-95](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/ops/infra/production/setup.md#L63)
- [docs/delivery/GOLIVE_RUNBOOK.md:67](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/delivery/GOLIVE_RUNBOOK.md#L67)
- [docs/delivery/MIGRATION_SCOPE.md:30-35](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/delivery/MIGRATION_SCOPE.md#L30)
- [docs/delivery/MIGRATION_SCOPE.md:65-66](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/delivery/MIGRATION_SCOPE.md#L65)
- [backend/migrations/seeds/live_insert_standard_reservation_types.sql:47-97](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/migrations/seeds/live_insert_standard_reservation_types.sql#L47)

<a id="foundation-ops-contracts-365"></a>
## EMR-365 — 契約設計｜ワクチン単価の小数を保存・計算・表示できる精度と丸め規則を確定する（MIG-31境界）

decimal技術契約採択済み。master物理型/API/会計consumer/CSV activation未接続
Plane: https://app.plane.so/baritechllc/browse/EMR-365/ / 本文更新 2026-10-21T07:32:30.282919Z / 全コメント 12 / 最新 c316705c-c9eb-4b53-a4d5-e11571030552

<a id="foundation-ops-contracts-365-ac"></a>
### AC別判定
- 契約numeric(23,4)、ExactNumeric、price string|null、段階互換: 採択済み
- A: DDL/model/request/response/FE適合: main未実装の具体証拠あり
- B: exact会計source・snapshot/caller閉包: 旧int64 consumer残存
- C: 新decimal profile同版receipt: 未受領、実source/pin未確認

<a id="foundation-ops-contracts-365-adopted-subset"></a>
### 実装済み/採択済みsubset
- #784選定契約と#802/#803計画main採用済み。精度再質問不要
- 会計内部unit_price/quantity numeric化は001:9398等に存在。これはvaccines.priceのdecimal化ではない
- 価格以外PATCHのnil不変更、既存version・保険設定writerは既存成果として保持

<a id="foundation-ops-contracts-365-unconnected"></a>
### 未接続・未確認
- vaccines.priceは001:746 bigint、model:21 *int64、request:24/59 *int64、response:15 *int64 omitempty。新string/decimal契約へ未接続
- 会計接種候補row.UnitPrice *int64、serviceの明細単価int64、unbilled旧計算int64(float64(...))残存。会計exact保存型だけで計算・輸送成立としない
- FE treatment-plan-master-modelはdata.priceをそのままpayloadへ渡す。全FE受け手退役/生成型closureは今回未精査
- 001:9396 remains unadoptedは旧履歴コメントで最新採否ではない。本文先頭の旧未採択も最新コメントにより失効

<a id="foundation-ops-contracts-365-dependencies"></a>
### 依存（全票Done待ちへ拡張しない）
- A owner365＋DDL446/466・API/生成型ownerの必要subset、B505/481/455と価格source/snapshot/current caller、C MIG31/37/41/349同版receipt
- 旧税blank/1外税・÷1.1不採用・旧snapshot保持は既決。税/保険未確認算式は会計ownerの有限依存に残し、新業務値を作らない
- 旧移行原値はsource正本で保持。製品price変更を理由に旧負/NULL/空/0の無断書換えやMIG31再開をしない

<a id="foundation-ops-contracts-365-fixture-boundary"></a>
### fixture・採用履歴・実行境界
- 採択契約§8: 0.4/0.04/0.5/500/API0/source0/負/legacy負/NULL/欠落/空、0.4×3、scale超過、末尾0、最大値、overflow、CSV/profile/CAS/旧snapshot。すべて設計fixture、実行NOT_RUN
- PR784 merge b48a67d9908a8172400956edf38d3860138d9b7f; PR802 a88fa4add58d43cff643590f9f507c8ab1a3c085; PR803 c86c8dd2530bbea3d1810e07badb72f3ed1a07e1

後続owner: 365-A製品exact単価owner、365-B会計505/481、365-C MIG31→37/41→349

<a id="foundation-ops-contracts-365-source-evidence"></a>
### 固定source根拠
- [docs/work/emr/emr-365-vaccine-price-decimal-contract-20261021.md:19-85](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-365-vaccine-price-decimal-contract-20261021.md#L19)
- [docs/work/emr/emr-445-stage1-contract-start-plan-20261021.md:1](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-445-stage1-contract-start-plan-20261021.md#L1)
- [backend/migrations/001_init.sql:742-747](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/migrations/001_init.sql#L742)
- [backend/migrations/001_init.sql:9396-9406](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/migrations/001_init.sql#L9396)
- [backend/internal/model/vaccine.go:21](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/model/vaccine.go#L21)
- [backend/internal/medicalrecord/vaccine_request.go:24-59](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/vaccine_request.go#L24)
- [backend/internal/medicalrecord/vaccine_response.go:15](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/vaccine_response.go#L15)
- [backend/internal/medicalrecord/vaccine_service.go:70-71](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/medicalrecord/vaccine_service.go#L70)
- [backend/internal/billing/billing_item_repository_unbilled.go:22-102](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/billing/billing_item_repository_unbilled.go#L22)
- [backend/internal/billing/billing_item_unbilled.go:138](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/billing/billing_item_unbilled.go#L138)
- [frontend/src/features/master/routes/treatment-plan-master-model.ts:156](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/frontend/src/features/master/routes/treatment-plan-master-model.ts#L156)

<a id="foundation-ops-contracts-369"></a>
## EMR-369 — 【バックログNo.57】医院税率変更の影響範囲の回答を確認する

D10回答契約は既決。回答受入/実医院設定は未確認、製品改修票ではない
Plane: https://app.plane.so/baritechllc/browse/EMR-369/ / 本文更新 2026-10-21T04:04:26.324192Z / 全コメント 1 / 最新 e184a403-fde9-463b-b10d-ffab4ba47acc

<a id="foundation-ops-contracts-369-ac"></a>
### AC別判定
- 医院master変更で全会計に自動伝播しない回答: 最新本文＋D10一致
- 未確定編集中も保存snapshot不変: D10 AC476
- 明示再適用で対象/前後/結果追跡: D10 AC477、製品完成は別
- 回答確認結果/医院設定・代表権限: 未確認

<a id="foundation-ops-contracts-369-adopted-subset"></a>
### 実装済み/採択済みsubset
- 最新本文10/21追補はD10 draft-no-auto-update/explicit-reapplyを採用。既決外税・snapshot保持を再選定しない
- 旧sourceのtreatmentMasterTax/手動指定・default選択は存在する限定根拠

<a id="foundation-ops-contracts-369-unconnected"></a>
### 未接続・未確認
- 回答担当/医院の最終確認receiptなし。実医院設定値・実効grantを測定していない
- billing_item_service旧float tax/整数単価と現保存型の差を、369回答修正済みから全実装適合へ飛躍させない
- 明示再適用の全caller/履歴/競合/確定済訂正実装は今回未精査。不存在断定ではなく505/465等の既存ownerへ参照

<a id="foundation-ops-contracts-369-dependencies"></a>
### 依存（全票Done待ちへ拡張しない）
- 505計算snapshot/462訂正・取消再請求/465 master再適用の必要subset。369全受入を①設計や465のgateにしない

<a id="foundation-ops-contracts-369-fixture-boundary"></a>
### fixture・採用履歴・実行境界
- 確認シナリオ: 未確定行追加→master税変更→旧snapshot不変、指定行のみ再適用/前後履歴、確定済は訂正へ。これは既決ACの参照で新算式ではない
- 税率実変更、権限付与、実会計再計算は一切実行なし。NOT_RUN

後続owner: 369回答担当・医院（説明/変更対象確認）。製品接続は会計505/465/462

<a id="foundation-ops-contracts-369-source-evidence"></a>
### 固定source根拠
- [docs/spec/domains/accounting/lifecycle.md:476-477](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/spec/domains/accounting/lifecycle.md#L476)
- [backend/internal/billing/billing_item_unbilled.go:221-261](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/billing/billing_item_unbilled.go#L221)
- [backend/internal/billing/billing_item_service.go:144-167](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/backend/internal/billing/billing_item_service.go#L144)
- [docs/work/emr/emr-445-stage1-contract-start-plan-20261021.md:1](https://github.com/MinoruSoga/AnimalEkarte/blob/77c28fbfbafa257dcca2421af99ff72da6889717/docs/work/emr/emr-445-stage1-contract-start-plan-20261021.md#L1)

<a id="foundation-ops-contracts-read-limits"></a>
## 読取限界
- PR623/784/802/803 metadataとdiffを取得。234/292は実commitを直接取得。329の採用PR番号は未特定。検索不発をPR不存在の証拠にはしていない。
- 292の過去run37538425729は最新attempt jobを再取得し成功を確認したが、稼働instance退出を照会していない。
- 329の3/3 PASSは既存Plane報告であり、raw logの独立再検証ではない。全票の今回runtimeはNOT_RUN。
- 234初期grantの限定source成立と、共通medical-record lifecycleの旧型接続残差を分けた。最新495裁定を本資料で上書きしない。
- 365の採択契約§8は既存期待値の引用参照。新税/保険算式を作成していない。実source/pin/正式activationは未受領。
- 現mainは固定時点。作業開始時は同一ownerの最新採用SHAと共有path claimを再照合する。
