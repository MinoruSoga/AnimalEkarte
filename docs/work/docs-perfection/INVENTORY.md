# Docs Markdown inventory

対象は `find docs -name '*.md'` の全件（大文字小文字はそのまま）。開始時の187件は [INVENTORY-SCAN.json](EVIDENCE/INVENTORY-SCAN.json) に構造・localリンク・Git最終更新手がかりを保存した。新規coordinator Markdownも個別行で列挙し、自己参照を理由に黙って除外しない。非Markdownの6既存ファイルは本promptのMarkdown対象外。

分類は**修復キューの一次トリアージ**であり、全ファイルの全命題が現在の実装/runtimeに一致すると認定するものではない。全件にローカルリンクtarget存在・見出し・完全一致重複・Git日付の検査を実施し、各カテゴリ読者が矛盾候補を調査した。外部URL/料金/Linear/provider/DB実適用は検証していない。バッククォート内の全コード参照・全リンクanchorは網羅していないため、子で個別に照合する。

- `current`: 上記の明示した一次検査で要修復候補なし。日付付き履歴はその役割が適切ならcurrent。古いSHAを現行と読み替えない。
- `conflict-suspect`: 文書間/参照先/コードに両側根拠のある矛盾候補。確定した製品不具合という意味ではない。
- `structure-debt`: 参照先、証跡表示、時点表現などを改善する具体的候補。
- `stale-suspect` / `duplicate` / `obsolete-suspect`: 鮮度根拠/重複正本/退役根拠がある場合に用いる。今回は旧日付だけでは付けず、完全一致重複は0件。
- `skip-with-reason`: 今回新設したcoordinator文書は既存本文監査から外し、別の全AC・独立レビューで検証する。理由を各行に残す。

L=検査したローカルMarkdownリンクtarget数、H=見出し行数。更新手がかりはGit最終変更日とcommit略記であり、内容の正しさを保証しない。役割は [ROLE-MAP](ROLE-MAP.md)、category childは [LEDGER](LEDGER.md)、修復の両側根拠は [REPAIR-QUEUE](REPAIR-QUEUE.md)。current行も子の全件再照合対象。

| Path | Role | Classification | 最終更新手がかり | 一次検査・根拠・子での確認 |
|---|---|---|---|---|
| docs/README.md | Map | current | 2026-09-06 396d722b2 | L=12, H=4; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/README.md | Map | current | 2026-08-31 d71cb9c2a | L=16, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/adr/001-system-architecture.md | History | current | 2026-08-31 d71cb9c2a | L=6, H=6; 日付付き履歴/採択理由として保持。現行実装の証明に転用しない |
| docs/architecture/adr/002-multitenancy-clinic-id-isolation.md | History | current | 2026-09-06 396d722b2 | L=4, H=6; 日付付き履歴/採択理由として保持。現行実装の証明に転用しない |
| docs/architecture/adr/003-payment-method-identity-and-consistency.md | History | current | 2026-09-06 396d722b2 | L=1, H=34; 日付付き履歴/採択理由として保持。現行実装の証明に転用しない |
| docs/architecture/adr/004-checkup-canonical-system.md | History | current | 2026-08-31 d71cb9c2a | L=0, H=6; 日付付き履歴/採択理由として保持。現行実装の証明に転用しない |
| docs/architecture/adr/005-go-gin-backend-guidelines.md | History | current | 2026-08-31 d71cb9c2a | L=2, H=7; 日付付き履歴/採択理由として保持。現行実装の証明に転用しない |
| docs/architecture/adr/006-backend-domain-package-boundaries.md | History | current | 2026-09-06 396d722b2 | L=31, H=17; 日付付き履歴/採択理由として保持。現行実装の証明に転用しない |
| docs/architecture/adr/007-lab-device-receive-and-commit.md | History | current | 2026-09-06 396d722b2 | L=3, H=16; 日付付き履歴/採択理由として保持。現行実装の証明に転用しない |
| docs/architecture/adr/008-local-lab-device-agent.md | History | current | 2026-09-03 e5465de17 | L=0, H=6; 日付付き履歴/採択理由として保持。現行実装の証明に転用しない |
| docs/architecture/adr/README.md | Map | current | 2026-08-31 d71cb9c2a | L=9, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/arch-a4-trigger-ledger.md | History | current | 2026-08-31 d71cb9c2a | L=1, H=4; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/auth.md | Map | current | 2026-09-06 dc2c48446 | L=2, H=14; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/be9-2a-boundary-map.md | History | current | 2026-09-06 396d722b2 | L=8, H=39; 日付付き履歴/採択理由として保持。現行実装の証明に転用しない |
| docs/architecture/composition-root-conventions.md | Map | current | 2026-08-31 d71cb9c2a | L=3, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/cross-domain-orchestration-catalog.md | Map | current | 2026-09-06 396d722b2 | L=3, H=5; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/data-flow.md | Map | current | 2026-09-06 396d722b2 | L=3, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/delete-soft-delete-patterns.md | Map | current | 2026-08-31 d71cb9c2a | L=3, H=8; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/erd.md | Map | current | 2026-09-02 ad037015b | L=1, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/exception-package-discipline.md | Map | current | 2026-09-06 396d722b2 | L=6, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/fe-feature-be-domain-map.md | Map | current | 2026-08-31 d71cb9c2a | L=3, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/model-write-owner-catalog.md | Map | current | 2026-09-06 396d722b2 | L=4, H=5; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/architecture/overview.md | Map | current | 2026-08-31 d71cb9c2a + 2026-09-07 Astra loop | RQ-011 resolved; 型別catalogとADR所有原則の案内を再照合。全件の静的再照合と検証境界はLEDGER / EVIDENCE/astra-loop-scan.json |
| docs/delivery/DELIVERY_PACKAGE.md | Map | current | 2026-09-06 396d722b2 | L=67, H=23; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/delivery/GOLIVE_RUNBOOK.md | Map | current | 2026-09-06 396d722b2 | L=24, H=14; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/delivery/OPERATION_MANUAL.md | Map | current | 2026-09-07 Issue照合 | RQ-002修復差分を開始時から保持し、会計作成条件を再照合。在庫UI案内を訂正。runtime未確認 |
| docs/delivery/README.md | Map | current | 2026-09-06 396d722b2 | L=14, H=7; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/CLAUDE.md | Constitution | current | 2026-08-31 133dfabf2 | L=4, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/README.md | Map | current | 2026-09-06 396d722b2 | L=14, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/backlog-spreadsheet.md | Map | current | 2026-08-31 133dfabf2 | L=0, H=4; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/ci-policy.md | Map | current | 2026-09-06 396d722b2 | L=0, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/coverage-policy.md | Map | current | 2026-08-31 71a30801d | L=0, H=7; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/A4_UI_REHEARSAL.md | Map | current | 2026-07-25 853c87e7a | L=1, H=7; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/BREAK-HOURS-SHAPE-AUDIT.md | Map | current | 2026-09-06 396d722b2 | L=0, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/CI-CD-PIPELINE.md | Map | current | 2026-09-06 396d722b2 | L=8, H=8; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/CLINIC_CSV_IMPORT.md | Map | current | 2026-09-06 396d722b2 | L=4, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/CLOUDFLARE-EXTERNAL-INTEGRATIONS-AUDIT.md | Map | current | 2026-09-06 396d722b2 | L=1, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/CRUD-SMOKE-TEST.md | Map | current | 2026-09-06 396d722b2 | L=1, H=18; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/DEPLOYMENT_CHECKLIST.md | Map | current | 2026-09-06 396d722b2 | L=3, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/F8_G4_FAILURE_REHEARSAL.md | Map | current | 2026-07-27 1dcbc9f2d | L=0, H=5; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/LAB_DEVICE_AGENT_MACOS.md | Map | current | 2026-09-06 396d722b2 | L=1, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/LAB_DEVICE_CONNECTIVITY.md | Map | current | 2026-09-06 396d722b2 | L=0, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/LOCAL_DB_RESET.md | Map | current | 2026-09-05 b3395ca21 | L=1, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md | Map | current | 2026-08-31 d35098796 | L=0, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/MIXED-PAYMENT-SMOKE-TEST.md | Map | current | 2026-09-06 396d722b2 | L=0, H=14; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/OLD_DB_HANDOFF_LOCAL.md | Map | current | 2026-09-06 396d722b2 | L=4, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/README.md | Map | current | 2026-09-06 396d722b2 + 2026-09-07 index edit | L=34, H=15; RQ-005; coordinatorでMac配布前提リンク1行追加。配布未確認 |
| docs/ops/deploy/SEED_MIGRATION_OPERATIONS.md | Map | current | 2026-09-06 396d722b2 | L=4, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md | Map | current | 2026-08-31 d35098796 | L=0, H=24; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/STG-CONTINUOUS-OPERATIONS.md | Map | current | 2026-09-06 396d722b2 + 2026-09-07 Astra loop | RQ-006 resolved; 一時採取と正式保存先の区分を確認。全件の静的再照合と検証境界はLEDGER / EVIDENCE/astra-loop-scan.json |
| docs/ops/deploy/STG-DEMO-DATA-LIFECYCLE.md | Map | current | 2026-09-05 b3395ca21 | L=8, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md | Map | current | 2026-09-06 396d722b2 | L=1, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/VERCEL-FRONTEND-STAGING-TEST.md | Map | current | 2026-09-06 dc2c48446 | L=2, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md | Map | current | 2026-09-06 396d722b2 | L=6, H=7; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/runbooks/README.md | Map | current | 2026-07-24 dad69bc6a | L=4, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/runbooks/SCHEDULER_OPERATIONS.md | Map | current | 2026-08-31 d35098796 | L=0, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/runbooks/SEC_SECRETS_5_GITLEAKS_HISTORY_INVENTORY.md | Map | current | 2026-08-31 d35098796 | L=3, H=4; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/deploy/runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md | Map | current | 2026-09-06 396d722b2 | L=4, H=8; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/infra/README.md | Map | current | 2026-08-31 133dfabf2 | L=8, H=1; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/infra/architecture.md | Map | current | 2026-08-31 71a30801d | L=0, H=5; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/infra/iac-guidelines.md | Map | current | 2026-08-31 71a30801d | L=0, H=5; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/infra/production/runbook.md | Map | current | 2026-09-06 396d722b2 | L=3, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/infra/production/setup.md | Map | current | 2026-09-06 396d722b2 | L=0, H=8; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/infra/reorg-plan.md | Map | current | 2026-08-31 71a30801d | L=3, H=4; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/infra/staging/runbook.md | Map | current | 2026-09-05 b3395ca21 | L=0, H=4; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/CLAUDE.md | Constitution | current | 2026-09-06 396d722b2 | L=8, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/CLINICAL-E2E-DESIGN.md | Map | current | 2026-09-06 396d722b2 | L=2, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/E2E_TESTING_GUIDE.md | Map | current | 2026-09-06 396d722b2 | L=6, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/INTEGRATION_TEST_PLAN.md | Map | current | 2026-09-06 396d722b2 | L=5, H=5; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/PERFORMANCE_PROFILING.md | Map | current | 2026-09-05 67f63da03 | L=0, H=7; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/README.md | Map | current | 2026-09-06 396d722b2 | L=16, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/S09-FIXTURE-DESIGN.md | Map | current | 2026-09-06 396d722b2 | L=1, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/SECTION_14_MANUAL_TEST_GUIDE.md | Map | current | 2026-09-06 396d722b2 | L=4, H=8; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/TEST_ARCHITECTURE.md | Map | current | 2026-09-06 396d722b2 | L=13, H=8; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/UAT-DOMAIN-STATUS.md | Status | current | 2026-09-06 396d722b2 + 2026-09-07 Astra loop | RQ-004 resolved; 非配布証跡は識別子。内容/現在所在は未確認、UAT状態は未更新。全件の静的再照合と検証境界はLEDGER / EVIDENCE/astra-loop-scan.json |
| docs/ops/testing/UAT-ENV-SETUP.md | Map | current | 2026-09-06 396d722b2 + 2026-09-07 Astra loop | RQ-008 resolved; 実在helperの文書相対リンクとadvisory境界を確認。全件の静的再照合と検証境界はLEDGER / EVIDENCE/astra-loop-scan.json |
| docs/ops/testing/liff-verification.md | Map | current | 2026-09-06 396d722b2 | L=5, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/FIELD-LEVEL-PROTOCOL.md | Map | current | 2026-09-06 396d722b2 | L=1, H=8; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/FORM-FIELD-INVENTORY.md | Map | current | 2026-09-06 396d722b2 | L=31, H=37; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/LAB_DEVICE_CLIENT_UAT.md | Map | current | 2026-08-21 eac91852f | L=0, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/README.md | Map | current | 2026-09-06 396d722b2 | L=26, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S01-deceased-pet-guard.md | Map | current | 2026-08-31 07cac9477 | L=14, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S02-exam-abnormal-highlight-lock.md | Map | current | 2026-08-31 71a30801d | L=9, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S03-vaccination-next-due-autocalc.md | Map | current | 2026-09-06 396d722b2 | L=12, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S04-liff-reservation-journey.md | Map | current | 2026-08-31 07cac9477 | L=4, H=5; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S05-hospitalization-cycle.md | Map | current | 2026-08-31 07cac9477 | L=14, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S06-record-lock-audit-trail.md | Map | current | 2026-09-06 396d722b2 | L=12, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S07-estimate-status-control.md | Map | current | 2026-08-31 71a30801d | L=3, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S08-accounting-corrections.md | Map | current | 2026-09-06 396d722b2 | L=4, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S09-closing-time-boundaries.md | Map | current | 2026-08-31 71a30801d | L=4, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S10-customer-aggregation-consistency.md | Map | current | 2026-08-31 07cac9477 | L=2, H=5; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S11-trimming-combined-accounting.md | Map | current | 2026-08-31 71a30801d | L=16, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S12-liff-pet-health.md | Map | current | 2026-08-31 07cac9477 | L=5, H=5; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/S13-identity-links-manual-correction.md | Map | current | 2026-08-31 07cac9477 | L=1, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/UAT-254-CLOSE-CHECKLIST.md | Map | current | 2026-09-06 396d722b2 | L=16, H=5; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/V01-clinical-forms.md | Map | current | 2026-09-06 396d722b2 | L=19, H=19; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/V02-accounting-reservation-forms.md | Map | current | 2026-09-06 396d722b2 | L=15, H=17; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/V03-owner-pet-staff-forms.md | Map | current | 2026-09-06 396d722b2 | L=10, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/V04-settings-master-forms.md | Map | current | 2026-09-06 396d722b2 | L=19, H=15; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/ops/testing/scenarios/V05-auth-line-forms.md | Map | current | 2026-09-06 396d722b2 | L=25, H=23; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/product-philosophy.md | Constitution | current | 2026-09-07 (RQ-010 child) | RQ-010 resolved; 説明例を補助的明示同意・サーバー側防御・臨床安全優先と整合 |
| docs/spec/README.md | Map | current | 2026-08-31 33ff4889b | L=11, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/cash-register.md | Constitution | current | 2026-09-06 396d722b2 | L=3, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/customer-aggregation.md | Constitution | current | 2026-09-06 396d722b2 | L=1, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/design-system.md | Constitution | current | 2026-09-06 396d722b2 | L=4, H=42; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/line/CLAUDE.md | Constitution | current | 2026-08-31 33ff4889b | L=1, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/line/README.md | Map | current | 2026-09-06 396d722b2 | L=7, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/line/architecture.md | Constitution | current | 2026-09-06 396d722b2 | L=7, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/line/cost-analysis.md | Constitution | current | 2026-08-31 33ff4889b | L=1, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/line/lstep-integration.md | Constitution | current | 2026-09-06 396d722b2 | L=7, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/line/reservation-spec.md | Constitution | current | 2026-09-06 396d722b2 | L=2, H=8; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/line/setup.md | Constitution | current | 2026-09-06 396d722b2 | L=6, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/reservation-to-record-flow.md | Constitution | current | 2026-08-31 e2d90fdaa | L=2, H=14; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/00-pet-selection.md | Constitution | current | 2026-08-31 71a30801d | L=0, H=4; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/01-reception.md | Constitution | current | 2026-08-31 31b2c7e0b | L=0, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/02-reservations.md | Constitution | current | 2026-09-06 396d722b2 | L=0, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/03-owners-list.md | Constitution | current | 2026-08-31 31b2c7e0b | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/04-owners-form.md | Constitution | current | 2026-09-06 396d722b2 | L=1, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/05-medical-records-list.md | Constitution | current | 2026-08-31 31b2c7e0b | L=0, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/06-medical-records-form.md | Constitution | current | 2026-09-06 396d722b2 | L=3, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/07-hospitalization-list.md | Constitution | current | 2026-08-19 4154ba8d5 | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/08-hospitalization-detail.md | Constitution | current | 2026-09-06 396d722b2 | L=0, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/09-hospitalization-form.md | Constitution | current | 2026-09-06 396d722b2 | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/10-accounting-list.md | Constitution | current | 2026-08-31 31b2c7e0b | L=1, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/11-accounting-detail.md | Constitution | current | 2026-09-06 dc2c48446 | L=1, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/12-examinations-list.md | Constitution | current | 2026-08-31 31b2c7e0b | L=0, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/13-examinations-form.md | Constitution | current | 2026-09-06 396d722b2 | L=0, H=15; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/14-vaccinations-list.md | Constitution | current | 2026-08-31 31b2c7e0b | L=0, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/15-vaccinations-form.md | Constitution | current | 2026-08-31 31b2c7e0b | L=1, H=15; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/16-trimming-list.md | Constitution | current | 2026-07-16 a476b727b | L=0, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/17-trimming-form.md | Constitution | current | 2026-09-06 396d722b2 | L=0, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/18-inventory-list.md | Constitution | current | 2026-09-06 396d722b2 | L=0, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/19-clinic-settings.md | Constitution | current | 2026-08-31 31b2c7e0b | L=0, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/20-master-settings.md | Constitution | current | 2026-08-31 31b2c7e0b | L=24, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/21-login.md | Constitution | current | 2026-08-31 167bf68d0 | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/22-estimate-list.md | Constitution | current | 2026-08-31 167bf68d0 | L=0, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/23-estimate-form.md | Constitution | current | 2026-08-19 4154ba8d5 | L=0, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/24-shift-calendar.md | Constitution | current | 2026-09-06 396d722b2 | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/25-checkups-list.md | Constitution | current | 2026-08-19 4154ba8d5 | L=0, H=15; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/26-estimate-detail.md | Constitution | current | 2026-08-31 167bf68d0 | L=1, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/27-inventory-form.md | Constitution | current | 2026-08-31 167bf68d0 | L=0, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/28-line-reservation.md | Constitution | current | 2026-09-06 396d722b2 | L=1, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/29-closing-aggregation.md | Constitution | current | 2026-08-31 167bf68d0 | L=0, H=16; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/30-unpaid-list.md | Constitution | current | 2026-08-31 167bf68d0 | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/31-lstep-integration.md | Constitution | current | 2026-08-31 167bf68d0 | L=5, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/32-accounting-reports.md | Constitution | current | 2026-08-31 167bf68d0 | L=0, H=14; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/34-lstep-delivery-monitor.md | Constitution | current | 2026-09-06 396d722b2 + 2026-09-07 Astra loop | RQ-001/003 resolved; 安全gapとfallback未是正を通常経路の要件から分離。全件の静的再照合と検証境界はLEDGER / EVIDENCE/astra-loop-scan.json |
| docs/spec/screens/35-internal-manual.md | Constitution | current | 2026-08-19 4154ba8d5 | L=0, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/36-aggregation-dashboard.md | Constitution | current | 2026-08-19 4154ba8d5 | L=0, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/37-line-reserve-owner-flow.md | Constitution | current | 2026-09-06 396d722b2 | L=2, H=14; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/38-liff-pet-health.md | Constitution | current | 2026-08-31 167bf68d0 | L=2, H=16; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/39-owner-report.md | Constitution | current | 2026-08-31 167bf68d0 | L=0, H=17; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/40-identity-links.md | Constitution | current | 2026-08-31 167bf68d0 | L=0, H=15; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/99-medical-record-flow.md | Constitution | current | 2026-09-06 396d722b2 | L=2, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/CLAUDE.md | Constitution | current | 2026-08-31 31b2c7e0b | L=1, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/README.md | Map | current | 2026-09-06 396d722b2 + 2026-09-07 Astra loop | RQ-007 resolved; 検査受信のADR参照を確認。全件の静的再照合と検証境界はLEDGER / EVIDENCE/astra-loop-scan.json |
| docs/spec/screens/common-dialogs.md | Constitution | current | 2026-09-06 396d722b2 | L=0, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/README.md | Map | current | 2026-09-06 396d722b2 + 2026-09-07 Astra loop | RQ-007 resolved; item-master設定と日常受信を区別して既存仕様へ案内。全件の静的再照合と検証境界はLEDGER / EVIDENCE/astra-loop-scan.json |
| docs/spec/screens/settings/closing-time-settings.md | Constitution | current | 2026-08-31 e2d90fdaa | L=1, H=14; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-animal-species.md | Constitution | current | 2026-07-30 a3174b84e | L=0, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-cage.md | Constitution | current | 2026-08-19 4154ba8d5 | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-campaigns.md | Constitution | current | 2026-07-16 a476b727b | L=0, H=9; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-chief-complaint.md | Constitution | current | 2026-07-16 a476b727b | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-diagnosis.md | Constitution | current | 2026-07-16 a476b727b | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-examinations.md | Constitution | current | 2026-09-06 396d722b2 | L=3, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-hospitalization-plan.md | Constitution | current | 2026-08-19 4154ba8d5 | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-insurance.md | Constitution | current | 2026-09-06 396d722b2 | L=1, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-interview.md | Constitution | current | 2026-09-03 58088c242 | L=0, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-medicine.md | Constitution | current | 2026-09-06 396d722b2 | L=1, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-merchandise.md | Constitution | current | 2026-07-16 a476b727b | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-occupation.md | Constitution | current | 2026-07-16 a476b727b | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-permission-group.md | Constitution | current | 2026-09-06 396d722b2 | L=1, H=13; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-reservation-type.md | Constitution | current | 2026-09-06 396d722b2 | L=2, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-shift-template.md | Constitution | current | 2026-08-22 02457ccb1 | L=1, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-staff.md | Constitution | current | 2026-09-06 396d722b2 | L=0, H=10; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-treatment.md | Constitution | current | 2026-08-22 02457ccb1 | L=1, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-trimming-course-type.md | Constitution | current | 2026-07-16 54fea1948 | L=0, H=8; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/master-trimming.md | Constitution | current | 2026-07-16 54fea1948 | L=1, H=12; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/screens/settings/payment-methods.md | Constitution | current | 2026-07-16 a476b727b | L=0, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/specification.md | Constitution | current | 2026-09-06 396d722b2 | L=9, H=11; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/spec/ui-design-compliance.md | Status | current | 2026-09-06 396d722b2 | L=2, H=3; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/work/README.md | Map | current | 2026-09-06 396d722b2 + 2026-09-07 index edit | L=8, H=4; coordinatorで保守role map入口1行追加。Linear SoT/削除zone維持 |
| docs/work/decisions/README.md | Map | current | 2026-09-06 396d722b2 | L=3, H=2; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/work/decisions/fable-po-recommendation.md | History | current | 2026-08-31 3b898cc2b + 2026-09-07 Astra loop | RQ-009 resolved; 2026-08-20時点の表現と外部未確認を結線。全件の静的再照合と検証境界はLEDGER / EVIDENCE/astra-loop-scan.json |
| docs/work/docs-perfection/CHILD-GOALS/architecture.md | Status | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/docs-perfection/CHILD-GOALS/delivery.md | Status | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/docs-perfection/CHILD-GOALS/ops.md | Status | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/docs-perfection/CHILD-GOALS/product-philosophy.md | Status | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/docs-perfection/CHILD-GOALS/spec.md | Status | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/docs-perfection/CHILD-GOALS/work.md | Status | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/docs-perfection/INVENTORY.md | Map | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/docs-perfection/LEDGER.md | Status | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/docs-perfection/PROGRESS.md | History | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/docs-perfection/README.md | Map | skip-with-reason | 2026-09-07 調査入口追加 | 既存本文監査から除外。Astra loop Phase Aで入口リンクと役割を検査。以後の修復証拠はLEDGER |
| docs/work/docs-perfection/REPAIR-QUEUE.md | Status | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/docs-perfection/ROLE-MAP.md | Map | skip-with-reason | 2026-09-07 coordinator新規 | 既存本文監査から除外。AC1–9と独立2レビューで別途検証する保守成果物 |
| docs/work/linear-f1-f6-mapping.md | Status | current | 2026-09-06 7c6592f9f | L=0, H=6; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/work/phase2-deferred.md | Map | current | 2026-08-31 be5f96431 | L=1, H=4; 一次検査で候補なし。全命題の照合は該当カテゴリ子 |
| docs/work/skill-reeval-2026-09-06.md | History | current | 2026-09-06 8d85d42bd | L=0, H=3; 日付付き履歴/採択理由として保持。現行実装の証明に転用しない |

分類集計（2026-09-07 Issue照合、既存199件）: conflict-suspect=0, current=187, skip-with-reason=12。RQ-002修復差分は今回の開始時点で存在する。coordinator時点の分類/測定はINVENTORY-SCAN.jsonと旧checksに保持。全命題・runtime一致の認定ではない。新規の照合レポート1件を含めたMarkdownは200件。全件の根拠・不確実性は [Issue照合レポート](ISSUE-RECONCILIATION.md) を参照。

| 追加文書 | 役割 | 状態 | 根拠 |
|---|---|---|---|
| [ISSUE-RECONCILIATION.md](ISSUE-RECONCILIATION.md) | Status | awaiting-human-verification | 2026-09-07 全件照合の範囲・訂正・外部残件 |
