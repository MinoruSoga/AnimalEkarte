<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# 詳細ERD — 全カラム・全外部キー

[概要ERD](../erd.md)

マイグレーション **32ファイル** から **136テーブル・1638カラム・323外部キー** を静的に導出しています。

DDLは `backend/migrations/` 直下の `.sql` をファイル名順に読み、CREATE/ALTER/DROPとレビュー済みDOブロックの変更を反映します。seed、`schema_migrations`（migrate runnerの管理テーブル）、関数・trigger・RLS・実データはこの一覧の対象外です。

この解析器は、対応済みDDLの列・キー・indexを扱う限定的な静的解析器です。未対応のschema変更や未レビューのDOブロックはエラーにします。SQLは実行しません。DBへの適用成功・稼働DBとの一致・STG/本番受入の証明にはなりません。

型とDEFAULTはDDLの宣言表記です。PKが暗黙に与えるNOT NULLを反映しています。複合FKの列順、NULL可否、削除動作、部分uniqueの述語は各ページを参照してください。

## 更新・確認

手編集せず、リポジトリのルートから実行します（Python標準ライブラリのみ。DBアクセス・依存導入なし）。

```sh
python3 -B scripts/tools/generate-erd-detail.py --write
python3 -B scripts/tools/generate-erd-detail.py --check
python3 -B scripts/tools/test_erd_schema.py
python3 -B scripts/tools/test_generate_erd_detail.py
```

`--check` はファイルを書き換えず、カラム・FK・文書・DDL内容・生成器の更新漏れを検出します。未知のSQLが追加された場合は解析器とテストを先に更新します。`--write` は自動生成マーカーを持つ文書だけを更新・整理し、同名の手書き文書の上書きを拒否します。

## テーブル一覧

各ページはテーブルの全カラム・全FKを持ち、参照先のページへ移動できます。

| テーブル | カラム数 | 外部キー数 |
|:---|---:|---:|
| [accounts](accounts.md) | 8 | 0 |
| [animal_species](animal_species.md) | 6 | 0 |
| [appointment_trimming_details](appointment_trimming_details.md) | 15 | 4 |
| [appointment_trimming_options](appointment_trimming_options.md) | 6 | 4 |
| [appointments](appointments.md) | 25 | 9 |
| [audit_logs](audit_logs.md) | 13 | 2 |
| [billing_confirmations](billing_confirmations.md) | 11 | 3 |
| [billing_items](billing_items.md) | 26 | 10 |
| [billing_refunds](billing_refunds.md) | 9 | 3 |
| [billings](billings.md) | 19 | 7 |
| [cages](cages.md) | 13 | 2 |
| [campaign_target_categories](campaign_target_categories.md) | 3 | 1 |
| [campaign_target_items](campaign_target_items.md) | 3 | 2 |
| [campaigns](campaigns.md) | 12 | 1 |
| [care_logs](care_logs.md) | 11 | 3 |
| [care_plan_items](care_plan_items.md) | 17 | 4 |
| [cash_register_close_adjustments](cash_register_close_adjustments.md) | 10 | 4 |
| [cash_register_closes](cash_register_closes.md) | 14 | 2 |
| [checkup_field_results](checkup_field_results.md) | 18 | 3 |
| [checkup_package_import_receipts](checkup_package_import_receipts.md) | 12 | 2 |
| [checkup_type_fields](checkup_type_fields.md) | 16 | 2 |
| [checkup_types](checkup_types.md) | 15 | 2 |
| [checkups](checkups.md) | 12 | 5 |
| [chief_complaint_types](chief_complaint_types.md) | 9 | 1 |
| [clinic_holidays](clinic_holidays.md) | 6 | 1 |
| [clinic_integrations](clinic_integrations.md) | 7 | 1 |
| [clinic_settings](clinic_settings.md) | 32 | 1 |
| [clinical_plans](clinical_plans.md) | 13 | 5 |
| [clinics](clinics.md) | 26 | 1 |
| [closing_special_periods](closing_special_periods.md) | 9 | 1 |
| [companies](companies.md) | 14 | 0 |
| [consultations](consultations.md) | 15 | 2 |
| [daily_records](daily_records.md) | 6 | 2 |
| [diagnosis_names](diagnosis_names.md) | 10 | 2 |
| [diagnosis_types](diagnosis_types.md) | 9 | 1 |
| [estimate_items](estimate_items.md) | 19 | 5 |
| [estimates](estimates.md) | 21 | 6 |
| [exam_reference_ranges](exam_reference_ranges.md) | 10 | 3 |
| [exam_results](exam_results.md) | 18 | 2 |
| [exam_type_fields](exam_type_fields.md) | 10 | 2 |
| [exam_types](exam_types.md) | 12 | 2 |
| [examination_revision_items](examination_revision_items.md) | 20 | 1 |
| [examination_revisions](examination_revisions.md) | 22 | 1 |
| [exams](exams.md) | 15 | 7 |
| [hospitalization_plans](hospitalization_plans.md) | 14 | 1 |
| [hospitalizations](hospitalizations.md) | 18 | 7 |
| [inquiries](inquiries.md) | 17 | 3 |
| [inquiry_templates](inquiry_templates.md) | 10 | 1 |
| [insurances](insurances.md) | 11 | 1 |
| [inventory_items](inventory_items.md) | 15 | 1 |
| [lab_device_item_masters](lab_device_item_masters.md) | 11 | 2 |
| [lab_device_mapping_draft_items](lab_device_mapping_draft_items.md) | 18 | 3 |
| [lab_device_mapping_drafts](lab_device_mapping_drafts.md) | 13 | 1 |
| [lab_device_mapping_generation](lab_device_mapping_generation.md) | 3 | 1 |
| [lab_device_mapping_specs](lab_device_mapping_specs.md) | 12 | 3 |
| [lab_device_station_settings](lab_device_station_settings.md) | 5 | 1 |
| [lab_device_waits](lab_device_waits.md) | 8 | 3 |
| [lab_devices](lab_devices.md) | 9 | 2 |
| [lab_import_events](lab_import_events.md) | 12 | 2 |
| [lab_import_exam_retraction_items](lab_import_exam_retraction_items.md) | 8 | 1 |
| [lab_import_exam_retractions](lab_import_exam_retractions.md) | 8 | 2 |
| [lab_import_job_items](lab_import_job_items.md) | 11 | 3 |
| [lab_import_jobs](lab_import_jobs.md) | 23 | 2 |
| [lab_import_revert_receipts](lab_import_revert_receipts.md) | 10 | 1 |
| [lab_import_usage_receipts](lab_import_usage_receipts.md) | 7 | 2 |
| [line_customers](line_customers.md) | 9 | 2 |
| [line_link_tokens](line_link_tokens.md) | 7 | 2 |
| [line_reservation_settings](line_reservation_settings.md) | 33 | 1 |
| [line_send_logs](line_send_logs.md) | 10 | 3 |
| [lstep_auto_managed_prefixes](lstep_auto_managed_prefixes.md) | 6 | 0 |
| [lstep_condition_tag_mappings](lstep_condition_tag_mappings.md) | 6 | 0 |
| [lstep_csv_imports](lstep_csv_imports.md) | 12 | 2 |
| [lstep_delivery_trigger_log](lstep_delivery_trigger_log.md) | 12 | 2 |
| [lstep_friend_attribute_snapshots](lstep_friend_attribute_snapshots.md) | 14 | 2 |
| [lstep_migration_progress](lstep_migration_progress.md) | 9 | 2 |
| [lstep_next_visit_sync_retries](lstep_next_visit_sync_retries.md) | 6 | 2 |
| [lstep_send_purpose_tag_prefixes](lstep_send_purpose_tag_prefixes.md) | 6 | 0 |
| [lstep_settings](lstep_settings.md) | 6 | 1 |
| [lstep_sync_error_counters](lstep_sync_error_counters.md) | 6 | 2 |
| [lstep_tag_cache](lstep_tag_cache.md) | 7 | 2 |
| [lstep_tag_code_mappings](lstep_tag_code_mappings.md) | 10 | 1 |
| [lstep_trigger_priorities](lstep_trigger_priorities.md) | 6 | 1 |
| [manual_article_versions](manual_article_versions.md) | 8 | 2 |
| [manual_articles](manual_articles.md) | 10 | 1 |
| [medical_record_addenda](medical_record_addenda.md) | 8 | 3 |
| [medical_record_image_deletions](medical_record_image_deletions.md) | 13 | 1 |
| [medical_record_image_upload_quota](medical_record_image_upload_quota.md) | 6 | 2 |
| [medical_record_images](medical_record_images.md) | 15 | 3 |
| [medical_records](medical_records.md) | 16 | 8 |
| [medicine_dose_params](medicine_dose_params.md) | 15 | 2 |
| [medicines](medicines.md) | 22 | 3 |
| [merchandise_items](merchandise_items.md) | 12 | 1 |
| [occupations](occupations.md) | 9 | 1 |
| [owner_identity_group_members](owner_identity_group_members.md) | 8 | 2 |
| [owner_identity_groups](owner_identity_groups.md) | 6 | 1 |
| [owners](owners.md) | 37 | 2 |
| [password_reset_tokens](password_reset_tokens.md) | 5 | 1 |
| [payment_methods](payment_methods.md) | 9 | 1 |
| [payment_splits](payment_splits.md) | 10 | 5 |
| [payments](payments.md) | 19 | 6 |
| [permission_group_rules](permission_group_rules.md) | 10 | 1 |
| [permission_groups](permission_groups.md) | 10 | 1 |
| [pet_chronic_conditions](pet_chronic_conditions.md) | 11 | 2 |
| [pet_identity_group_members](pet_identity_group_members.md) | 8 | 2 |
| [pet_identity_groups](pet_identity_groups.md) | 8 | 2 |
| [pet_owners](pet_owners.md) | 7 | 3 |
| [pets](pets.md) | 34 | 5 |
| [procedures](procedures.md) | 16 | 2 |
| [reservation_type_available_slots](reservation_type_available_slots.md) | 10 | 2 |
| [reservation_type_groups](reservation_type_groups.md) | 9 | 1 |
| [reservation_type_occupations](reservation_type_occupations.md) | 5 | 3 |
| [reservation_type_unavailable_times](reservation_type_unavailable_times.md) | 10 | 2 |
| [reservation_types](reservation_types.md) | 23 | 3 |
| [shared_files](shared_files.md) | 12 | 3 |
| [shift_entries](shift_entries.md) | 10 | 2 |
| [shift_entry_breaks](shift_entry_breaks.md) | 4 | 1 |
| [shift_template_breaks](shift_template_breaks.md) | 4 | 1 |
| [shift_templates](shift_templates.md) | 12 | 1 |
| [staff_clinic_assignments](staff_clinic_assignments.md) | 8 | 2 |
| [staff_notes](staff_notes.md) | 7 | 2 |
| [staff_permission_groups](staff_permission_groups.md) | 3 | 2 |
| [staff_reservation_capabilities](staff_reservation_capabilities.md) | 5 | 3 |
| [staff_reservation_exclusions](staff_reservation_exclusions.md) | 3 | 2 |
| [staffs](staffs.md) | 16 | 3 |
| [support_bug_reports](support_bug_reports.md) | 18 | 2 |
| [support_chat_messages](support_chat_messages.md) | 8 | 2 |
| [token_blacklist](token_blacklist.md) | 3 | 0 |
| [treatment_plans](treatment_plans.md) | 16 | 3 |
| [treatments](treatments.md) | 28 | 10 |
| [trimming_course_types](trimming_course_types.md) | 8 | 1 |
| [trimming_courses](trimming_courses.md) | 13 | 2 |
| [trimming_image_deletions](trimming_image_deletions.md) | 4 | 1 |
| [trimming_options](trimming_options.md) | 12 | 1 |
| [vaccinations](vaccinations.md) | 19 | 7 |
| [vaccines](vaccines.md) | 14 | 3 |
| [vital_records](vital_records.md) | 17 | 5 |

## 入力の照合情報

入力digest（SHA-256）: `07acb410525fd70cab29d491a72e284ff79114e9bdc34325a580828d29012562`。各ファイルのSHA-256と生成器のSHA-256は [manifest.json](manifest.json) に記録します。

| migration | SHA-256 |
|:---|:---|
| [001_init.sql](../../../backend/migrations/001_init.sql) | `d792bca4f4d985f1b2acdc7cd9ac0c9af80fe28aa3135712ec70896086c8ba2d` |
| [002_medical_records_entered_by_staff_fk.sql](../../../backend/migrations/002_medical_records_entered_by_staff_fk.sql) | `231060f40e288776dc7768f24e2f81fab8c5226a3f1045633398442c3ed62a95` |
| [003_appointments_created_by_staff_fk.sql](../../../backend/migrations/003_appointments_created_by_staff_fk.sql) | `97254617a6254eb2f26b1e6c9f018ce7c376c7b0a46648ef4efd458d64af955d` |
| [004_billing_items_treatment_lifetime_unique.sql](../../../backend/migrations/004_billing_items_treatment_lifetime_unique.sql) | `b1259855989346dd33351f92f731c53a43997b007eb9c0c2b0c6d974ae72e937` |
| [005_child_records_version.sql](../../../backend/migrations/005_child_records_version.sql) | `9e188500b5c439793db86b4293b9c6184ae781e3b9e7c8f298e0034e7809abdd` |
| [006_accounts_rls_ops_bypass.sql](../../../backend/migrations/006_accounts_rls_ops_bypass.sql) | `394051c95126ca113f7f77d86601b78b21bc1d4cddfad6b1f13fcabab77efaa1` |
| [007_billings_clinic_owner_scheduled_index.sql](../../../backend/migrations/007_billings_clinic_owner_scheduled_index.sql) | `447358dbd89473c3e00981c7a648d50e2ba3144629ff00c6a3eead8f6eb3488d` |
| [009_reconcile_002_master_seed_checksum.sql](../../../backend/migrations/009_reconcile_002_master_seed_checksum.sql) | `f6aeb8c91f869826fd40cff75a8e12c2f310b5aa941003db1026d28d24340906` |
| [010_grant_lab_import_create_to_default_groups.sql](../../../backend/migrations/010_grant_lab_import_create_to_default_groups.sql) | `1f6e701a214002cc3b845f926e7a92f62df304e879bbd83330ef610f3107667f` |
| [011_care_plan_items_manual_other.sql](../../../backend/migrations/011_care_plan_items_manual_other.sql) | `fb18358ec0a1bebc27df08cc9b3d854d5743bd5d85bd91b8d5cb63e4d7c10a65` |
| [011_support_bug_reports.sql](../../../backend/migrations/011_support_bug_reports.sql) | `172d9a893bcdef398836f680e4a5e950f3def91d107463e64789f88b7fe31442` |
| [012_pets_name_origin_meeting_story.sql](../../../backend/migrations/012_pets_name_origin_meeting_story.sql) | `13eda3b9ea12da06dfc2ca0df17a578ee9ef859d08ce5793b13413160e49a2fc` |
| [013_support_chat_messages.sql](../../../backend/migrations/013_support_chat_messages.sql) | `ec7c75eb33629161239b69180ce448da3ca6b36424dabb737767a0a709760015` |
| [014_drop_prescriptions.sql](../../../backend/migrations/014_drop_prescriptions.sql) | `8e8f8a5242d47fc98e5e340725c04f3a6e4a120b72d4f6fb50506df198b66bec` |
| [015_search_fold_indexes.sql](../../../backend/migrations/015_search_fold_indexes.sql) | `dcd28dc335221b1640a20a072ad67ea2d64d6892f06f29cce3ebb44aaa09300b` |
| [016_support_bug_reports_plane.sql](../../../backend/migrations/016_support_bug_reports_plane.sql) | `788adfeb1781da6b2bd1abbd81cac1622c7e84a924b2313edbd779c10d272276` |
| [017_multiclinic_staff_fk_fix.sql](../../../backend/migrations/017_multiclinic_staff_fk_fix.sql) | `08cd7ce63eb398849e5cc29217c580a042061a71ee0ef0d4124efe366f1ae174` |
| [018_pet_neutered_status.sql](../../../backend/migrations/018_pet_neutered_status.sql) | `0b854dd15ab39c820d537cf16fa1609d000680d50b45d64a70562546ebebcd0e` |
| [019_drop_checkup_package_import_permission.sql](../../../backend/migrations/019_drop_checkup_package_import_permission.sql) | `4376b00cdedd123c5c379c5645d82d71bb0e761cfe29eb0f06feaaefe3d506fa` |
| [020_reconcile_002_master_seed_checksum.sql](../../../backend/migrations/020_reconcile_002_master_seed_checksum.sql) | `7c2e346b5c2cb7c6335ae5758fb86b21c916cefa623b13bd16804dd2c9a0c57a` |
| [021_medical_record_image_deletions.sql](../../../backend/migrations/021_medical_record_image_deletions.sql) | `1fb99ccbfed5dea9f7754a98556365e80d642e44fbe3918d899295af9759ee61` |
| [022_staff_clinic_assignments_capability_configured_at.sql](../../../backend/migrations/022_staff_clinic_assignments_capability_configured_at.sql) | `9816da61ce3366b30193d0cc72119ee81676a982ced2469c28c98f67d002bb44` |
| [023_lab_device_mapping_drafts.sql](../../../backend/migrations/023_lab_device_mapping_drafts.sql) | `85a499173bdd75206fb7e72f073ea0af74f09a8aab8189b6cd9635717118067b` |
| [024_lab_device_mapping_drafts_clinic_unique.sql](../../../backend/migrations/024_lab_device_mapping_drafts_clinic_unique.sql) | `9912a1bbe728e0742efadbee60fa9a6cb861210f3ebcdef17b0e71f44184a706` |
| [025_cages_merged_into_cage_id.sql](../../../backend/migrations/025_cages_merged_into_cage_id.sql) | `59c37ee264645c366ecd97af29e3b53b74a160294b06c350c13c4603aed173ba` |
| [028_appointments_tentative_status.sql](../../../backend/migrations/028_appointments_tentative_status.sql) | `9dc685a628a6b610b40dad5f952da0c227d8196c50875ca5e17525e0fca422b1` |
| [029_appointments_one_active_tentative_index.sql](../../../backend/migrations/029_appointments_one_active_tentative_index.sql) | `37a81bcb5e3ec2bfad61d2e735cf2eb094af01baebd242b4f0b4ab9f229f136f` |
| [030_next_visit_recommended_date_to_tentative.sql](../../../backend/migrations/030_next_visit_recommended_date_to_tentative.sql) | `1fcba7b937432b1531343595f84803f1a54e0afa384c2cb892a1f89daed9b075` |
| [031_appointments_source_medical_record_clinic_fk.sql](../../../backend/migrations/031_appointments_source_medical_record_clinic_fk.sql) | `9efb120dd59f96455d076dde6f3e2d127787c240bbb2c83c004f32bcebed8fba` |
| [032_trimming_image_deletions.sql](../../../backend/migrations/032_trimming_image_deletions.sql) | `932a005007410d5884baa070399f53a06ce485ce9256e53c06ab1c8d9010c2ba` |
| [033_lstep_next_visit_sync_retries.sql](../../../backend/migrations/033_lstep_next_visit_sync_retries.sql) | `eef5732c5499d286de58ec386841042f1f9720486d22b7afc5641d4ff0079e61` |
| [034_lstep_next_visit_cleanup_target.sql](../../../backend/migrations/034_lstep_next_visit_cleanup_target.sql) | `e5f78083f6de3e720476865d14f54b14fa1a9fed465a9a605b1436cada953567` |

レビュー済みDOブロック（本文変更・新規追加は再レビューが必要）:

- 001_init.sql:3020 sha256=1f2d68c8a027d0a56679a58d695bb875d8916a8349b50fd5495ccea35839b2e8 rls_policy
- 001_init.sql:3842 sha256=7b22dfd85b61582c63f7b7c69eaffe5d6f6c60e4f33be57395476e479c0e335e data_assertion
- 001_init.sql:3914 sha256=2d7574c417ab8910cc806bddfdecaca8d3a7386db3a356a50276c028f9885af4 data_assertion
- 001_init.sql:4859 sha256=45920ac83a22c2ad218f09b590835f888ffcc40130aac19753b9d217a42e1edd data_cleanup
- 001_init.sql:5227 sha256=75ddac964c4f3c96c68336d52d31f6d0716f7ed89f4a3fc8bb9ac232fd2aba17 drop_checkup_parent
- 001_init.sql:5245 sha256=229b993c292b1849d80bf76664311627d4e1743c830260813d5af110ab32462f add_checkup_parent
- 001_init.sql:5331 sha256=aa8a1ace3fb3ecfa46c6d0ca4c961cf1072501c0f01fde5bfe69a0eed3b26f09 data_assertion
- 009_reconcile_002_master_seed_checksum.sql:47 sha256=0d8177b1a72683892e1e2ee88f6f3ae2e147ecd10a7e72ce0afc5ac80600ca50 checksum
- 010_grant_lab_import_create_to_default_groups.sql:55 sha256=bb8145f87b85e4e433021540faa2cdc794b0a57b0cbebc937c6e396f80f3ba04 data_assertion
- 010_grant_lab_import_create_to_default_groups.sql:73 sha256=1ef0b2b7e415cda4a466684aa33120d6c7c70eafea515ce7ef28350349241322 checksum
- 020_reconcile_002_master_seed_checksum.sql:40 sha256=5fa63dccad7cd9c60ceebee5735eed2c41e828f0c5668ba107e5b15d18420afd checksum
- 030_next_visit_recommended_date_to_tentative.sql:147 sha256=f6382ad0b0e4c1b49a7ad03cc005aecaa0180cbaab1a9580c23bbf0aa4948942 data_warning
- 031_appointments_source_medical_record_clinic_fk.sql:18 sha256=a2a0d54165c328b2ce6e41ba40d202cad9168941086cc888bccaadda1d27cedb data_assertion
