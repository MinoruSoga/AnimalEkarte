# 保存定義の削除・統合対照（2026-10-19）

本人指示「不要なカラムとかもない状態にしてよ？」に基づく、公開head `85b72c5c1d4cab985ced01d5dddef5f3b701a368` からの追加整理。既存backendとの互換性だけを維持理由にしていない。001全表/全列をカタログ化し、現在の業務入力・履歴・監査・技術運用の用途で分類した。列単位の一覧/出典hashは[JSON](schema-cleanup-audit.json)、型・NULL/FK/indexは[詳細ERD](../../architecture/erd-detail/index.md)。これは机上での用途分類であり、全資料の網羅・実行時適合・受入合格を示さない。

- 表: 246 → 241（9表統合・除去、4表追加）。
- 列: 2790 → 2691。存続表から117列除去、除去表を含め200列除去、追加表を含め101列追加。
- 001 SHA256: `e56a8e88a133f956996c50ca6ef3b26004712cbccd5df1365ab496c6d08bff76`。
- DBへの適用、データ移行、STG、追加試験、mergeは未実施。旧001のCREATE/ALTER履歴を残し、末尾のDROP/統合まで適用した最終定義をERDに記す。SQL原文に旧名が残ることと、最終表/列の存在は区別する。

## 除去・統合先

| 元表 / 除去した列（表除去は全列） | 統合先・導出元 | 根拠と用途 |
|---|---|---|
| `appointment_trimming_details` / `course_id` | trimming_booking_revisions | [予約courseは採用版へ。実施詳細/実source lineは別用途。](../../spec/domains/trimming/lifecycle.md) |
| `appointment_trimming_options` / 全列（表を除去） | trimming_booking_revision_options | [予約採用版のoption集合へ。](../../spec/domains/trimming/lifecycle.md) |
| `billing_confirmations` / 全列（表を除去） | doctor_review_sets / doctor_review_records | [固定対象世代・digestを持つ医師確認集合へ。任意確認メモはsetsに維持。](../../spec/domains/accounting/lifecycle.md) |
| `billing_items` / `treatment_id`, `appointment_id`, `trimming_course_id`, `trimming_option_id`, `vaccination_id`, `exam_id` | billing_source_bindings | [請求元をtyped binding1箇所へ統合。free row/物販の既存入力は残し、同一元のactive二重請求を防ぐ。](../../spec/domains/accounting/lifecycle.md) |
| `billing_refunds` / 全列（表を除去） | fund_consumption_events(refund) / refund_requests | [実返金eventと依頼状態を区別し二重返金台帳を除去。](../../spec/domains/accounting/lifecycle.md) |
| `billings` / `subtotal`, `tax_total`, `total_amount`, `has_insurance`, `status`, `completed_at`, `completion_request_id`, `completion_request_hash`, `calculation_version`, `ever_finalized` | claim_state / current_calculation_id / billing_calculation_snapshots / operation_results | [請求状態/計算版/操作receiptだけを正本とし二重金額、旧完了flagと完了専用keyを除去。入金済みはfund配分から導出。](../../spec/domains/accounting/lifecycle.md) |
| `cash_register_close_adjustments` / 全列（表を除去） | finance_source_events / cash_close_late_events / cash_actual_corrections | [会計訂正source・遅着・実査訂正を明確化し二重調整台帳を除去。](../../spec/domains/cash-register/business-rules.md) |
| `cash_register_closes` / `cash_difference`, `deleted_at` | actual_cash - theoretical_cash / cash_actual_corrections | [差額は2値から導出。immutable原締めにsoft-delete欄は不要。既存非partial期間uniqueを維持。](../../spec/domains/cash-register/business-rules.md) |
| `clinical_plans` / `diagnosis_type_id`, `diagnosis_name_id`, `diagnosis_2_type_id`, `diagnosis_2_name_id` | clinical_plan_diagnoses | [第一/第二それぞれ複数傷病を行で保持。病名マスタ分類と第一/第二の区分を分け、件数上限/重複病名禁止は追加しない。](../../spec/domains/medical-records/lifecycle.md) |
| `common_owner_basic_revisions` / `address`, `secondary_phone`, `additional_basic_fields` | typed basic columns | [曖昧な住所alias/追加JSONを型付き採用項目へ。DOBと保存Cygram結果/ルールは明示用途の例外として維持。](../../spec/domains/owners-pets/lifecycle.md) |
| `common_owner_clinic_bindings` / `mapping_version`, `is_active` | immutable binding identity / operation receipt | [更新不可対応表にmutable active/counterを置かない。新unlink/splitは未選定。](../../spec/domains/owners-pets/lifecycle.md) |
| `common_pet_basic_revisions` / `sex`, `additional_basic_fields` | typed basic columns | [sex/追加JSONを既存gender型と明示基本項目へ。staff専用欄をLINE公開DTOへ自動追加しない。](../../spec/domains/owners-pets/lifecycle.md) |
| `common_pet_clinic_bindings` / `mapping_version`, `is_active` | immutable binding identity / operation receipt | [更新不可対応表にmutable active/counterを置かない。新unlink/splitは未選定。](../../spec/domains/owners-pets/lifecycle.md) |
| `estimate_presentations` / `calculation_snapshot` | estimate_calculation_revisions | [曖昧な計算JSON複製を固定計算版FKへ。](../../spec/domains/estimates/business-rules.md) |
| `estimates` / `subtotal`, `tax_total`, `total_amount`, `insurance_amount`, `discount_amount`, `supersedes_estimate_id` | current_calculation_id / estimate_calculation_revisions / estimate_successor_links | [金額条件/結果を固定計算版へ、後継は理由と提示対象付きrelationへ。](../../spec/domains/estimates/business-rules.md) |
| `hospitalizations` / `hospitalization_type`, `start_date`, `end_date`, `status` | stay_cases / stay_movement_events | [予定/実績・hotel/admissionと移動履歴の正本を統合。入院診療内容と医院内IDは残す。](../../spec/domains/hospitalization/lifecycle.md) |
| `inventory_items` / `last_restocked`, `status` | inventory_movement_events / quantity and min_stock_level | [表示statusと最終入庫日は数量/閾値と増減実日時から導出。現在数量projectionは競合制御に必要。](../../spec/domains/inventory/business-rules.md) |
| `line_customers` / `owner_id` | owner_line_binding_states | [一対一の本人確認/解除/履歴を独立正本へ。LINE固有属性は残す。](../../spec/line/lstep-integration.md) |
| `line_reservation_settings` / `line_channel_id`, `line_channel_secret`, `liff_id`, `line_access_token`, `line_bot_user_id` | clinic_integrations | [同じLINE資格を2箇所へ保存しない。実資格移動/鍵変更は実施なし。](../../spec/line/lstep-integration.md) |
| `line_send_logs` / 全列（表を除去） | line_dispatch_intents / attempts / results | [確認payload・目的・外部発生時刻と結果をcanonical送信履歴へ統合。](../../spec/line/lstep-integration.md) |
| `manual_articles` / `title`, `order_value`, `section`, `body_markdown`, `updated_by_staff_id` | manual_article_versions + current_version_id / published_version_id | [本文・ラベル・編者を固定内容版へ一元化。](../../spec/domains/internal-manual/README.md) |
| `medical_records` / `status` | medical_record_consultation_events | [旧finalized状態による通常編集禁止を除去。診療終了/解除履歴は医師会計確認と独立。](../../spec/domains/medical-records/lifecycle.md) |
| `operation_results` / `state` | operation_results row presence | [常に成功の固定stateは不要。失敗/unknownはattempt結果へ。S1のclaimed/succeeded receiptは別契約で維持。](../../design/shared/integration-contracts.md) |
| `owner_deposit_entries` / `payment_method_id` | payment_funds / fund_consumption_events | [口座配分へ実支払方法を重複保存しない。実金移動は元fundとrefund、純粋口座金額訂正は理由付きcredit/debit。](../../spec/domains/accounting/lifecycle.md) |
| `owners` / `name`, `name_kana`, `birth_date`, `company`, `postal_code`, `address1`, `address2`, `home_postal_code`, `home_address1`, `home_address2`, `phone`, `company_phone`, `email`, `line_user_id`, `line_followed_at`, `line_blocked_at`, `line_id_confirmed_by`, `line_id_confirmed_at` | common_owner_basic_revisions / owner_line_binding_states / line_customers | [共通最新基本情報、本人確認、LINE側事実の保存責任を分離。医院内ID/個別配信・会計条件は維持。](../../spec/domains/owners-pets/lifecycle.md) |
| `payment_funds` / `origin_kind`, `legacy_payment_id`, `legacy_split_id`, `deposit_receive_entry_id` | payment_funds.original_source_reference / fund_consumption_events | [旧payment由来の多重FKとorigin分類を除去。元ID/種別/検証根拠は固定import evidence、預り配分はeventで保持。](../../spec/domains/accounting/lifecycle.md) |
| `payment_splits` / 全列（表を除去） | payment_funds / fund_consumption_events | [実受領手段別のfundと配分へ。入力預かり額・釣銭はfund固定入力へ保存。](../../spec/domains/accounting/lifecycle.md) |
| `payments` / 全列（表を除去） | payment_funds / fund_consumption_events / fund_record_corrections | [受領1回・請求配分複数・実返金/誤記訂正を単一fund系へ統合。](../../spec/domains/accounting/lifecycle.md) |
| `pets` / `name`, `name_kana`, `animal_species_id`, `gender`, `status`, `birth_date`, `breed`, `color`, `weight`, `neutered_date`, `acquisition_type`, `danger_level`, `food`, `environment`, `phone`, `last_visit`, `remarks`, `deceased_at`, `deceased_reason`, `blood_type`, `microchip_number`, `danger_reason`, `name_origin`, `meeting_story`, `neutered_status` | common_pet_basic_revisions / actual_visit_facts | [基本情報の医院別重複を除去。登録体重は共通基本値、測定体重はvital。last_visitは実来院sourceから導出。](../../spec/domains/owners-pets/lifecycle.md) |
| `reservation_type_occupations` / 全列（表を除去） | staff_reservation_capabilities + reservation_eligibility_evidence | [職種による旧受付ゲートは失効。職種master/staff所属は維持。](../../spec/screens/settings/master-reservation-type.md) |
| `reservation_types` / `color`, `category` | business_kind / reservation_type_groups.color | [旧2分類と種類別色を除去し採用済み4区分・グループ色/標準色へ。parent_idは既存種類階層。](../../spec/screens/settings/master-reservation-type.md) |
| `shift_entries` / `start_time`, `end_time` | starts_at / ends_at | [日跨ぎを表せない具体勤務の時刻重複を日時へ。日付は選択勤務日の業務キー。](../../spec/domains/staff-workforce/README.md) |
| `shift_entry_breaks` / `break_start`, `break_end` | starts_at / ends_at | [具体休憩は日時だけ。テンプレートの時刻は別用途で維持。](../../spec/domains/staff-workforce/README.md) |
| `staff_reservation_exclusions` / 全列（表を除去） | staff_reservation_capabilities + reservation_eligibility_evidence | [否定制限は失効。元の原情報をevidence保持、未決の自動変換は行わない。](../../spec/screens/settings/master-staff.md) |
| `support_bug_reports` / `plane_issue_id`, `plane_issue_url`, `plane_sync_error` | support_delivery_destinations / attempts / results | [単一Plane成功/エラー欄をprovider別結果履歴へ統合。報告statusは院内処理状態として別用途。](../../spec/domains/support/README.md) |
| `support_delivery_intents` / `confirmed_destinations` | support_delivery_destinations | [宛先配列の二重保存を除去。正規化宛先・固定個数/確認fingerprintを正本にする。](../../spec/domains/support/README.md) |
| `treatment_plans` / `subtotal` | selected unit_price/quantity/discount inputs; billing calculation revisions | [可変plan小計は入力から導出。請求固定金額は計算版で保持。](../../spec/domains/medical-records/lifecycle.md) |

## 残したものと境界

医院内owner/pet ID、所有関係、診療/会計/預りの元IDは維持。独立した共通UUIDへの対応で基本情報だけを共有し、D19 IDを流用しない。旧IDの実データ付替えやunlink/split/transfer操作は未実施・未選定。共有基本改訂のstaff専用欄を認可外やLINE公開DTOへ自動公開しない。

登録時体重と来院測定vital、スタッフ問診・初診構造化問診・飼主LINE原文は用途と出所が異なるため維持。画面タブ統合は保存原文の混合を意味しない。計算条件/結果・提示/発行原本・過去版・監査・外部原payloadは固定履歴として維持し、現在masterから再生成しない。Cygramの採用済み結果と計算ルールはDOBと同じ共通改訂で保存し、DOB訂正時に再計算・削除時に消去するconsumer責任を明記した。具体の未確認計算値はseedしない。

予約の種類parent、staff職種/position/employment/権限、option全体併用可否とcourse別組合せ、日付なしシフトtemplate時刻はそれぞれ採用済みの別用途。機器mapping案/取込receipt、認証token・quota・同期cache・削除queueも運用目的があるため維持。要約専用の非表示保存欄は採用範囲のまま、全文音声/手動貼付UIは追加しない。

共有ファイルの保存区分は原本/必須履歴/一時/未分類。期限切れは削除許可と同義にしない。実データの分類や削除は実施していない。

## 預り額の誤記訂正

受領fundは実受領を1回だけ記録する。口座receiveは同fundのdeposit_allocateであり別受領ではない。請求充当はdeposit_release＋allocate、解除はrelease＋deposit_allocate、実返金はdeposit_release＋refund。口座だけの実支払方法欄を除去した。

例: 受領額5,000円を4,000円に誤記訂正する場合、理由付きfund訂正版→口座correction_debit 1,000円→元fundのdeposit_release 1,000円を同一TXで保存する。口座残高/元receive余力・fund有効額を合わせ、実返金eventは作らない。増額はcorrection_creditとdeposit_allocate。各訂正の口座調整総額は元fund訂正差額以下で、消費済み分を超える減額は配分整合が必要。実cash移動は0。外部金銭/会計consumerの実装・試験は別担当。

## consumerへの契約変更

親担当作成のEMR-489–511および再利用EMR-468に、以下の差分を渡す（単位と票番号の厳密な対応は親管理、ここでは推測しない）。

1. owner/pet reader/writer・重複候補・検索/帳票・LINEは共通基本改訂とbindingへ。旧医院別氏名/住所/電話をfallback正本として読まない。医院内IDは変更しない。
2. 予約/staff/シフト/トリミングは4区分・positive capability・具体日時・予約固定版へ。旧否定制限原情報の検証と移行方式は別設計で、原情報を削除しない。
3. 第一/第二傷病は各複数行、通常カルテstatusは診療終了eventへ。診療終了・医師確認・請求確定を別軸にする。
4. 会計は固定計算版とfund/配分へ、見積提示も固定計算版FKへ。旧payments/splits/refunds、独立可変合計・完了専用keyを併用しない。入力誤記訂正と実返金、口座配分と実金移動を区別する。
5. 締めはsource集合/遅着/実査訂正、在庫は数量ledgerから状態/最終入庫日導出、manualは内容版pointer、support/LINE送信はintent/attempt/resultへ。

これらのAPI/UI全writer参加、認可、required audit、DTO/検索index、既存seed/fixture・旧データのmappingと段階導入は未接続。DB実行成功・CodeRabbit最終完了・全AC合格は宣言しない。
