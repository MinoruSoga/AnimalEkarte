# 確定要件 → 保存定義 → 引渡し票（2026-10-19）

001・ERDの作成対照。DB適用・試験は本人指示により未実施。元18単位は先行subsetであり、全仕様対応完了の根拠ではなかった。今回の追加不足と既存拒否制約の訂正を以下に追跡する。列の完全一覧と根拠ファイルhashは[JSON](spec-storage-trace.json)、型・NULL・default・FK・indexは[全項目ERD](../../architecture/erd-detail/index.md)へ。

既存票は関連先であり、API実装が完了した証明ではない。449は金融consumer調査、452は領域全体で、追加内訳票の代用にしない。未確認の票番号は捏造せず、親担当で既存票照合後に不足分を起票する。

| 変更単位 | 根拠FR/AC・正本 | 表・主な保存対象 | 既存票／不足票の扱い | 維持した境界 |
|---|---|---|---|---|
| contact-candidates | [FR-owners-pets-contact-duplicate-warning, AC-owners-pets-contact-duplicate-warning, AC-owners-pets-contact-duplicate-separate](../../spec/domains/owners-pets/lifecycle.md) | owners, owner_identity_confirmations（全列はJSON） | EMR-446, EMR-466, EMR-470 / 内訳未確認は親へ | 電話/メールuniqueを非unique候補indexへ。同一確認済ownerはwriterが再利用。連絡先一致から自動同一判定せず、別人/複数頭を拒否しない |
| common-subjects | [common-owner-pet-20261017](../../spec/domains/owners-pets/lifecycle.md) | common_owner_subjects, common_pet_subjects, common_owner_clinic_bindings, common_pet_clinic_bindings, common_owner_basic_revisions, common_pet_basic_revisions（全列はJSON） | EMR-451, EMR-470 / 内訳未確認は親へ | common-owner-pet-20261017＋本人Sentinel_7195a603405c819199d5e2b74463a2d1の独立UUID/医院別ID対応表承認。自院利用のみ、D19/unlink再設計なし、実ID付替えなし |
| internal-booking-overlap | [FR-reservations-staff-overlap-warning](../../spec/domains/reservations/pages/reservation-management.md) | appointments, reservation_capacity_gates, patient_operation_gates, appointment_change_events（全列はJSON） | EMR-468, EMR-469 / 内訳未確認は親へ | 一律doctor時間EXCLUDEと同開始unique削除。院内警告確認後重複可、LINE枠/勤務/休憩/同時保存の検査は全writer参加が必要 |
| reservation-kind-duration-name | [FR-reservation-types-business-routing, FR-reservation-types-business-classification-lock, FR-reservation-types-duration-required, FR-reservation-types-duration-step, FR-reservation-types-duration-range, FR-reservation-types-name-history, FR-reservation-types-group-reactivation-intent](../../spec/screens/settings/master-reservation-type.md) | reservation_types, reservation_type_groups, reservation_type_name_revisions, reservation_contract_states, appointments（全列はJSON） | EMR-448, EMR-468, EMR-469 / 内訳未確認は親へ | 4kind、5〜480分5分刻み/default撤去、名称固定版、個別再開確認。旧用途/名称/期間を推測補完しない |
| staff-leave-shifts | [FR-staff-workforce-workforce-maintenance, FR-staff-workforce-history-and-access](../../spec/domains/staff-workforce/README.md) | staffs, staff_clinic_assignments, staff_leave_periods, staff_leave_changes, shift_entries, shift_entry_breaks, shift_bulk_operations（全列はJSON） | EMR-448 / 内訳未確認は親へ | 職種/役職/雇用を別、休業期間編集履歴、院別指名停止、日跨ぎ実timestamp。休業日付端点不明はunknown、退職と同一化せず新権限なし |
| edit-right-draft-recovery | [FR-medical-records-single-editor, AC-medical-records-draft-same-actor-recovery, AC-medical-records-draft-context-isolation](../../spec/domains/medical-records/lifecycle.md) | record_edit_rights, record_edit_right_events, staff_record_drafts, staff_record_draft_operations, medical_record_dependency_versions（全列はJSON） | EMR-452, EMR-471, EMR-472, EMR-473 / 内訳未確認は親へ | epoch/holder、本人account×record×tab暗号draft。app.current_account_idの信頼済tx設定必須。暗号鍵配布/端末方式/consumerは別、平文や鍵本体を保存しない |
| record-purpose-corrections | [AC-medical-records-wrong-patient-transfer, AC-medical-records-retrospective-no-reception-billing](../../spec/domains/medical-records/lifecycle.md) | medical_record_correction_links（全列はJSON） | EMR-452 / 内訳未確認は親へ | 元/先record・理由・由来・実診療日/入力時刻を区別。新受付/会計/在庫を自動生成しない |
| structured-initial-intake | [FR-medical-records-structured-initial-intake, FR-medical-records-intake-provenance, AC-medical-records-intake-unknown, AC-medical-records-intake-partial-date, AC-medical-records-intake-multiple](../../spec/domains/medical-records/pages/initial-intake.md) | initial_intake_revisions, initial_intake_entries（全列はJSON） | EMR-452 / 内訳未確認は親へ | yes/no/unknown/未回答、年/月/日/概算精度、複数entry、出所/旧版を保持。自己申告からD07 official結果を作らない |
| preventive-prescription-and-observations | [FR-medical-records-preventive-prescription, FR-medical-records-annual-preventive-medication, FR-medical-records-preventive-report-output, AC-medical-records-prevention-month-multiple](../../spec/domains/medical-records/pages/preventive-medication.md) | preventive_prescriptions, preventive_prescription_revisions, preventive_medication_observations, preventive_medication_revisions（全列はJSON） | EMR-452 / 内訳未確認は親へ | 処方と投薬確認を分離、投薬改訂が固定処方版を参照。同月複数可、医学的頻度/自動量計算/自動請求出庫なし。2帳票は固定保存版をconsumerが読む |
| multiple-dose-patterns | [AC-dose-source-pattern-choice, AC-dose-source-approval-limit](../../spec/domains/medical-records/lifecycle.md) | medicine_dose_params, medicine_dose_pattern_revisions, treatments（全列はJSON） | EMR-447, EMR-452 / 内訳未確認は親へ | 薬剤×種の一律uniqueを非uniqueへ、目的/経路/承認出典付きpattern版。臨床値の新seed/全種全計算の自動承認なし |
| stay-and-capacity | [FR-hospitalization-entry-and-start, FR-hospitalization-hotel-stay-capacity, FR-hospitalization-distinct-plans-and-records, FR-hospitalization-hotel-manual-pricing](../../spec/domains/hospitalization/lifecycle.md) | stay_cases, hotel_capacity_settings, stay_movement_events, admission_records, stay_extension_proposals, stay_plan_change_events, stay_price_agreements（全列はJSON） | EMR-452, EMR-461 / 内訳未確認は親へ | 予約/実滞在identityと予定/実時刻、入院開始時医師/記録、ホテル院別頭数上限。ケージ数/医師数を定員にしない。価格合意/予定預りと実受領は別 |
| accounting-calculation-history | [FR-accounting-calculation-snapshot](../../spec/domains/accounting/lifecycle.md) | billing_calculation_snapshots, billing_calculation_lines, billing_change_events, billing_reissue_links, billing_source_bindings（全列はJSON） | EMR-449, EMR-462, EMR-481, EMR-483 / 内訳未確認は親へ | 正確numeric入力/JPY丸め結果・税割引保険確認版。lifetime source unique→未release active source unique。取消理由付きreleaseで履歴を残し再請求、二重active請求は拒否 |
| funds-allocation-refund | [FR-accounting-owner-selected-charges, FR-accounting-receivable-partial-zero, FR-accounting-reasoned-financial-correction, AC-accounting-payment-method-correction-no-money](../../spec/domains/accounting/lifecycle.md) | payment_funds, fund_consumption_events, fund_record_corrections, refund_requests, payments, payment_splits（全列はJSON） | EMR-449, EMR-464 / 内訳未確認は親へ | 1請求1payment unique撤去、受領identityと複数請求配賦・残額・元充当解除上限・実返金を区別。方法/日時誤記訂正は実移動0。旧由来不明を架空受領にしない |
| financial-document-editions | [本文の確定契約](../../spec/domains/accounting/lifecycle.md) | financial_document_identities, financial_document_editions, financial_document_artifacts（全列はJSON） | EMR-449 / 内訳未確認は親へ | 元取引番号と請求/受領/返金由来、固定内容/計算版・artifactを保持。新しい法定様式/採番規則なし |
| estimate-presentations | [FR-estimates-successor, FR-estimates-billing-handoff, AC-estimates-presented-pending-original, AC-estimates-changed-content-explain](../../spec/domains/estimates/business-rules.md) | estimate_presentations, estimate_presentation_lines, estimate_replies, estimate_successor_links, estimate_billing_handoffs（全列はJSON） | EMR-449 / 内訳未確認は親へ | 提示時原本/固定行/返答先/理由付き後継/選択取込を保持。提示単価適用と税保険割引条件・請求確定は別 |
| closing-and-source-events | [FR-closing-movement-consumer, FR-closing-append-only, FR-closing-actual-input-correction, AC-closing-cash-difference-note, AC-closing-cash-difference-note-empty](../../spec/domains/cash-register/business-rules.md) | cash_opening_balances, cash_adjustment_events, finance_source_events, finance_event_ingestions, closing_boundary_revisions, cash_close_contexts, cash_close_source_events, cash_actual_corrections, cash_close_late_events（全列はJSON） | EMR-449 / 内訳未確認は親へ | 元締めimmutable、実査訂正追記/設定/境界/集計元版固定、金銭訂正source FK追加。預り/配賦と実移動の二重計上なし。差額時メモ必須だが原因判明不要 |
| inventory-ledger | [FR-inventory-record-maintenance, FR-inventory-adjustment, AC-inventory-concurrent-count, AC-inventory-actual-return](../../spec/domains/inventory/business-rules.md) | inventory_items, inventory_quantity_bases, inventory_movement_events（全列はJSON） | EMR-449 / 内訳未確認は親へ | 確認基準→整数前後数量/実日時/理由/元出庫/版。item lock・返却上限。単位属性訂正可/数量自動換算なし、旧残高から架空入庫を作らない |
| actual-visit-source | [FR-analytics-owner-counts, FR-analytics-lstep-observation](../../spec/domains/analytics/business-rules.md) | actual_visit_facts, finance_source_events, finance_event_ingestions（全列はJSON） | EMR-449 / 内訳未確認は親へ | 確定受付/確認済過去来院の根拠と元版・訂正参照。予定日/current status/上書きchecked_in_atから来院を捏造しない。日内distinctはconsumer |
| checkup-stage-recovery | [FR-checkups-results, AC-checkups-partial-save, AC-checkups-unknown-result, AC-checkups-master-change](../../spec/domains/checkups/lifecycle.md) | checkup_save_attempts, checkup_save_stage_results, checkup_result_revisions, checkups, checkup_types, checkup_type_fields（全列はJSON） | EMR-456, EMR-452 / 内訳未確認は親へ | 親→子→結果の各成功ID/版/操作を保存。全体atomic化や自動親削除へ変更せず、0/false/原field ID型snapshot保持。partialを全体成功にしない |
| examination-provenance | [FR-examinations-infection-summary](../../spec/domains/examinations/lifecycle.md) | exams, examination_revisions, examination_field_semantic_mappings（全列はJSON） | EMR-452 / 内訳未確認は親へ | 日付不明をNULL/精度で表現し他院手入力出所・前official参照・検証済明示mapping保持。新臨床コード対応seed/患者台帳の結果二重保存なし |
| manual-publication | [AC-internal-manual-05, AC-internal-manual-06, AC-internal-manual-09](../../spec/domains/internal-manual/README.md) | manual_articles, manual_article_versions, manual_article_operations（全列はJSON） | EMR-452 / 内訳未確認は親へ | 医院共通articleの現在内容版/公開版を分離、復元は新内容版、必要操作履歴/receipt。独立下書き承認workflow/院別コピー/旧override自動公開なし |
| support-confirmed-delivery | [AC-support-04, AC-support-07, AC-support-09](../../spec/domains/support/README.md) | support_bug_reports, support_delivery_intents, support_delivery_destinations, support_delivery_attempts, support_delivery_results（全列はJSON） | EMR-452 / 内訳未確認は親へ | 院内保存＋確認snapshot/intent/receipt後のみ配送。DiscordとPlaneの結果を分離。unknown/成功再送禁止、照合で未作成確認後だけ管理再試行。実外部送信なし |
| line-purpose-unlink-followup | [FR-crm-followup-confirmation, FR-crm-dm-purpose, AC-crm-unlink-reservations-preserved, AC-crm-relink-no-auto-resume, AC-crm-resume-selected-reservations](../../spec/line/lstep-integration.md) | owner_line_binding_states, owner_line_binding_events, reservation_notification_eligibility, followup_confirmations, line_dispatch_intents, line_dispatch_attempts, line_dispatch_results（全列はJSON） | EMR-469, EMR-470, EMR-474 / 内訳未確認は親へ | 予約本体/参照を残して通知停止、再連携自動再開なし、選択確認/目的/医師指定/本人版/不明attempt保持。新しい配信許可なし |
| common-operation-receipts | [本文の確定契約](../../design/shared/integration-contracts.md) | operation_bindings, operation_results, operation_attempt_outcomes（全列はJSON） | EMR-467, EMR-471, EMR-473 / 内訳未確認は親へ | C06選定契約: clinic/actor/owner/intent/key/HMACfingerprint永久束縛。成功receiptと失敗attempt分離、必要成功結果は業務変更同TXの延期制約。新IDでunknown救済しない |

## 拒否制約を訂正した箇所

- `uk_owners_clinic_phone` / `uk_owners_clinic_email`を削除し非unique候補検索indexへ。FR-owners-pets-contact-duplicate-warningとAC-owners-pets-contact-duplicate-warning/separateは、同一連絡先の別人の登録を認める。同一確認済み飼主の新規停止はstaff確認結果を使うwriter条件で、連絡先一致を同一性にしない。新ペットは既存ownerへ追加でき、共通電話の家族を一括拒否しない。
- `excl_appointments_doctor_timerange` / `uk_appointment_staff_time`を削除し非uniqueの担当者時間検索indexへ。FR-reservations-staff-overlap-warningは院内操作で警告確認後の重複保存を認める。LINEの枠判定と全予約writerの直列化・勤務/休憩は維持。複数頭を一律に拒否するpet/owner時間uniqueは新設しない。
- `uq_dose_params_med_species`を非uniqueへ。AC-dose-source-pattern-choiceの複数目的/経路patternに対応する。未承認臨床値の利用許可ではない。
- 会計元明細のlifetime uniqueを、release未済のactive claim uniqueへ。取消済原請求を保持した選択再請求に対応する。解除理由・操作・時刻は不変、解除前の二重active請求は禁止。paymentsのbilling_id uniqueも撤去し、受領と配賦を別保存する。
- 休業期間はimmutableな現在行にせず、版付き可変行と追記変更履歴へ。管理単位の属性変更は許可し、在庫数の自動換算・履歴書換えはしない。

## 共通患者IDの採用と境界

本人回答 `Sentinel_7195a603405c819199d5e2b74463a2d1`「ok」により、独立共通UUID＋医院別ID対応表を今回採用した。元の医院別IDは保持し、D19 group IDを流用しない。自院利用・基本情報の固定改訂/現在参照だけを定義した。bindingは元対応の履歴を保持する不変資源とし、この工程で新しいunlink/患者分割操作を選定しない。新規subject作成は認可済医院のbindingを先に保存する同一TX（common subject FKは延期）、basic改訂と結果receiptまで成立させる。旧全件mappingやID統合・実データ付替えは未実行。

## 未接続と検証状態

既存backend consumerはこの新DDLとの一致を前提にしていない。API/UI・全writerのgate/CAS/認可/必須監査/atomic receipt参加・送信provider照合・出力renderer・旧値分類/移行は各担当へ渡す。これらをDB定義を残す理由にせず、実装完了や利用開始の証拠にもしない。共通UUID方式は採用済みで回答待ちではない。外部Provider照会方式、暗号鍵配布、既存データmapping/unlinkの別設計をこのDDLだけで完了扱いしない。

音声/外部連携のPhase2、全文音声保存/手動貼付フォーム、健診専用作図/写真、新しい保険自動給付算式、EMR-365の未採用精度変更は含めない。EMR-294はmain統合済みとして採否を戻さず、10/13の当時支払額バグを仕様未決へ戻さない。旧単独保険/S1試験の成功は旧hash限定。今回の新テスト・DB適用・STG反映・mergeは行わない。

## 不足内訳票の候補（未起票）

親担当で既存子票との重複を照合してから起票する。24単位を24新規票必須とはしない。予約/勤務master調査はEMR-448、臨床master/用量調査はEMR-447に対応する。各候補のFR/AC・全列・既存票は上表とJSONの同unitへ接続する。

| 単位 | 候補名 |
|---|---|
| contact-candidates | 連絡先の重複候補・本人確認結果と別人登録を接続する |
| common-subjects | 共通UUIDと医院別IDの自院利用・共通基本情報reader/writerを接続する |
| internal-booking-overlap | 院内重複警告とLINE枠判定を全予約writerのgateへ接続する |
| reservation-kind-duration-name | 4業務種別・標準時間・名称固定版と個別再開を予約へ接続する |
| staff-leave-shifts | 休業期間・指名停止・日跨ぎ勤務と変更履歴を接続する |
| edit-right-draft-recovery | 編集epochと本人暗号draftの取得・移譲・復帰を接続する |
| record-purpose-corrections | 目的別カルテ訂正・過去入力と元記録参照を接続する |
| structured-initial-intake | 構造化初診問診の複数回答・時期精度・出所付き改訂を接続する |
| preventive-prescription-and-observations | 予防処方・年間投薬と固定処方版・専用2帳票を接続する |
| multiple-dose-patterns | 目的/経路別の承認用量pattern選択と固定版保存を接続する |
| stay-and-capacity | 同一滞在・院別ホテル頭数・入院開始記録と予定/実績を接続する |
| accounting-calculation-history | 算定固定版・active claim・理由付き取消再請求を接続する |
| funds-allocation-refund | 受領identity・複数請求配賦・実返金・記録訂正を接続する |
| financial-document-editions | 請求/受領/預り/返金帳票の元取引・固定版・artifactを接続する |
| estimate-presentations | 提示原本・返答・後継改訂と選択明細の請求引継ぎを接続する |
| closing-and-source-events | 金銭sourceイベントと元締め・実査訂正・締め後差分を接続する |
| inventory-ledger | 基準在庫と理由付き実移動・元出庫返却上限を接続する |
| actual-visit-source | 確定受付/過去来院根拠と訂正を集計readerへ接続する |
| checkup-stage-recovery | 健診の親/子/結果の各receipt照合と段階再開を接続する |
| examination-provenance | 他院手入力の検査出所・日付精度・意味mappingを接続する |
| manual-publication | 院内共通記事の公開pointer・改訂/復元履歴を接続する |
| support-confirmed-delivery | 院内保存先行の確認intent・宛先別claim/結果照合を接続する |
| line-purpose-unlink-followup | LINE本人版・解除停止/選択再開・目的別確認と配送照合を接続する |
| common-operation-receipts | 全ownerの保存成功receiptと失敗attemptの原子保存を接続する |

## 不要・重複保存の整理による更新

旧表/旧列を併用する前提は[削除・統合対照](schema-cleanup-audit.md)で置き換える。現行は241表、25対照単位。JSONの表列一覧は最終定義へ更新した。旧24単位に第一/第二傷病複数行と診療終了履歴を追記し、Cygram保存先を共通基本改訂へ明示した。親の追加consumer票EMR-489–511/468への厳密対応は親管理。旧payments等をcanonical保存先として使用しない。
