# DOCS-REFRESH-20260929 — ops-scenarios

## 対象と判定

`docs/ops/testing/scenarios/` 全 49 ファイル（README・FIELD-LEVEL-PROTOCOL・FORM-FIELD-INVENTORY・LAB_DEVICE_CLIENT_UAT・UAT-254-CLOSE-CHECKLIST・V01–V05・S01–S39）を現行コード・E2E・クライアント一次ソースと照合した。

| ファイル | 判定 | 主な根拠 |
|:--|:--|:--|
| README.md | corrected | Linear 言及が stale（Linear 2026-09-16 閉鎖・実行 SoT は Plane）。索引のファイル集合・実行順制約・V テーブルは現行セットと一致を確認 |
| FIELD-LEVEL-PROTOCOL.md | kept | F0–F6 定義・記録形式・status 語彙・完了判定は inventory/V 文書と整合。集計例は例示で非実行結果 |
| FORM-FIELD-INVENTORY.md | kept | wire key 命名・`is_active` 作成時デッドキー・`system_key` レスポンス限定・L-step write-only 秘密・`discount_value` 要実測注記・86 route pages 注記を現行 request builder/DDL と突合 |
| LAB_DEVICE_CLIENT_UAT.md | kept | bundle 内 `install.sh`/`diagnose.sh`/`SHA256SUMS`・manifest 別経路 SHA-256・NX600/AU10V 限定・「受付超過」運用は `scripts/build-lab-device-agent-bundle.sh`（57–68 行目）と `packaging/macos/` の実体と一致 |
| UAT-254-CLOSE-CHECKLIST.md | corrected | Linear 言及 5 箇所（Rules・owner comment・close gate 2 箇所・mermaid）を Plane へ訂正。5 flow mapping・禁止事項・mock 非代替の枠組みは現行と一致 |
| V01-clinical-forms.md | kept | 新規カルテ pet 未選択リダイレクト（`shouldRedirectToSelectPet`）、バイタル未来日時拒否・4 項目の少なくとも 1 つ必須、画像 10MB 上限（`medical_record_image_request.go`）、addendum 訂正内容+理由必須を突合 |
| V02-accounting-reservation-forms.md | kept | 在庫フォーム §12・数量/最低在庫の非負整数・単位必須（HTML+BE binding）・「inventory 再構築中は全フォーム完了を主張しない」・クレジット訂正監査の fail-closed は BE テスト領域と明記・共有 STG を append-only 締めテストに使わない旨を突合 |
| V03-owner-pet-staff-forms.md | kept | OwnerForm は `routes/OwnerForm.tsx`・`noValidate`、pet 死亡は `use-record-pet-death.ts` + `PATCH/DELETE /v1/clinics/:clinicId/pets/:id/death`（`lstep/routes.go`）、clinic 税率は FE 百分率→BE 小数、同名 clinic は一意制約なしで受理される現状を突合 |
| V04-settings-master-forms.md | kept | 現行 E2E（標準マスタ CRUD・動物種・診断・主訴・問診テンプレート・入院・トリミング・保険・物販・支払方法・予約区分・薬剤・診療 5 タブ・締め設定・シフト・lab-device 権限・clinic 請求・view-only 403 系）と `E2E_CLINICAL_FIXTURE` 使い捨てゲート、予約枠 retest の nested-form 既知バグ（API 迂回経路あり・専用 spec は UI 経路を通過）を突合。BUG-017（実名入り 409）は `13_UAT-R5確定仕様_2026-08-22.md` と一致 |
| V05-auth-line-forms.md | kept | ログイン `noValidate` + FE `minLength=6` / BE `min=8,max=72` の乖離を文書側も既に重点確認として記載。LINE 予約設定 4 wire key・L-step 3 secret/2 text/正整数・`cpm_version`/`is_sync_enabled`・tag-config は system-admin・checkup sync の自動管理タグ拒否を突合 |
| S01-deceased-pet-guard.md | kept | `PetDeceasedDialog`/`PATCH・DELETE /:id/death` 専用経路、予約既定検索の死亡除外と共通選択の `includeDeceased: true`+選択不可の経路差、`ValidatePetNotDeceased`・`HandlePetDeath/Revival`（`lstep_lifecycle_*.go`）を突合。死亡ペットの閲覧・連絡導線はブロックしない構成で Slack 2026-09-16 依頼と整合 |
| S02-exam-abnormal-highlight-lock.md | corrected | HIGH/LOW ラベル実体を `ExamItemsTable` → `ExamStatusBadge.tsx`（テーブルが描画）に精密化。inclusive range・未判定バッジ・`examination-unconfirm:edit`・完了シール・削除拒否文言は `examination_service.go`/`examination_lock.go` と一致 |
| S03-vaccination-next-due-autocalc.md | kept | `NEXT_SCHEDULE_OPTIONS`（3weeks/4weeks/1year/other）・独立既定 `1year`/カルテ既定 `4weeks`・`disabledDays`・`pet_id` サーバフィルタ・BUG-408 種フィルタなしを突合 |
| S04-liff-reservation-journey.md | kept | line-reserve `App.tsx` の trimming 挿入ページ、`LIFF_MOCK must not be set in release mode`（`config.go:287`）、`liff_auth.go`、`timeslot_engine`、ソースフィルタ「LINE予約」`value="line"` を突合 |
| S05-hospitalization-cycle.md | kept | 「会計画面へ進む」既定 OFF（`DischargeAlertDialog` `useState(false)`）、`POST /hospitalizations/:id/discharge-with-billing`（`hospitalization_handler.go:161`）と会計なし PATCH 分岐を突合 |
| S06-record-lock-audit-trail.md | kept | 削除拒否文言「確定済みまたは下書き以外の診療記録は削除できません」（`medical_record_crud.go:293`）・一覧/詳細の UI 差を突合 |
| S07-estimate-status-control.md | kept | `isEstimateLockedStatus`/`ESTIMATE_LOCKED_EDIT_MESSAGE`・後継ドラフト・best-effort 監査の扱いを突合 |
| S08-accounting-corrections.md | kept | `POST /accountings/complete`・`validatePaymentSplits`・部分入金拒否・監査 fail-closed を突合 |
| S09-closing-time-boundaries.md | kept | `resolvePeriodRange` 半開区間（`cash_register_service.go:498`）・`/accounting/close`・`/settings/closing-time` ルートを突合 |
| S10-customer-aggregation-consistency.md | kept | `COALESCE(bmr.date, b.scheduled_date)`（`ltv_repository_query.go:153`）・BE 20s/FE 25s timeout・「精算済」ラベルを突合 |
| S11-trimming-combined-accounting.md | kept | `unbilled-details`・`AssertNoBlockingUnbilled`（`billing_item_service.go:207`）・`POST /accountings/complete` を突合 |
| S12-liff-pet-health.md | kept | `health-card` API（`liff_handler.go:280`/`routes.go:188`）・`VITE_LIFF_MOCK` success-only・SEC-CS2-F02 紐付け制限を突合 |
| S13-identity-links-manual-correction.md | kept | `backend/internal/identitylink/`・`frontend/src/features/identity-links/`・`include_linked` クエリを突合 |
| S14-search-and-multi-word.md | kept | `applyPetListSearch`（`pet/repository.go:189`）の `strings.Fields` AND・語内 ILIKE・空白のみ `1=0` fail-closed を突合 |
| S15-insurance-rate-preservation.md | kept | `InsuranceCard` の 50/70 選択肢と `LEGACY_INSURANCE_RATIO_LABELS`（0.9→90%・1.0→100%）保持を突合 |
| S16-interview-history-navigation.md | corrected | 「Linear 追跡」→ Plane 追跡。`InterviewHistory` の `paths.medicalRecords.detail.getHref(item.id)`・「問診抜粋」見出しは現行と一致 |
| S17-treatment-quantity-commit.md | corrected | 「合成テスト 28 件」→ 数量確定操作の合成テストは 9 件。「Linear 追跡」→ Plane 追跡。`reduceQuantityEnterKey` 2 段階確定・repeat/isComposing/keyCode229 無視は現行と一致 |
| S18-treatment-search-dialog-height.md | kept | `max-h-[80vh]`（:175）・`max-h-[calc(80vh-12rem)]`（:202）・合成テスト 10 件の実数と一致 |
| S19-medical-record-viewport-fit.md | corrected | sidebar 開閉の実体を「`use-sidebar`/`SidebarProvider`」→ `Sidebar.tsx` ローカル `useState`+`matchMedia`・`LAYOUT.sidebar`（expanded `w-[220px]`/collapsed `w-[56px]`）へ訂正。9 タブ・`UnifiedTabsRoot` `flex-1 min-h-0` は現行と一致 |
| S20-master-to-accounting-path.md | kept | `useGetAllMerchandiseItems`（`ItemListCard`）・`unbilledWarnings`（`blocking && count > 0`）・12 フォーム（`UAT-R2-MASTER-PATH.md`）を突合 |
| S21-accounting-concurrency-idempotency.md | kept | `CompletionRequestID`/`CompletionRequestHash`・`FindByCompletionRequestID`・同一カルテ UNIQUE 409・`billing-schema-readiness` 前提ゲートを突合。実並行は UNKNOWN のまま維持 |
| S22-reservation-conflict-and-staff.md | kept | `filterStaffCandidatesByCapability` fail-closed・`staffCandidateEmptyMessage` 5 状態・`resolveStaffSelectionEligibility`/`STAFF_ORPHAN_REASON_MESSAGE`・SLACK-STAFF-SELECT 実機（iPad/一部 PC）受入を Slack 2026-09-13 一次ソースで確認 |
| S23-multi-clinic-scope.md | kept | `useClinicScope`（`use-clinic-scope.ts`）の `?clinics=`・`isMultiClinic`・`clinicNameById`、`isOtherClinic`（`MedicalRecordsListPanels.tsx:91`）・`AssertEnteredByActor`（`medical_record_entered_by_actor.go`）を突合。SLACK-CROSS-CLINIC は PO 裁定待ちのまま維持 |
| S24-vital-signs-latest-chips.md | kept | `latestVisitVitalChips`（`visit-vital-chips.ts:15`）のコピーソート降順・有限値のみ・全空で `null` を突合 |
| S25-microchip-header-display.md | kept | `microchipNumber` の `font-mono`・`aria-label`・`max-w-[12rem]` truncate・Tooltip・【死亡】マーク共存（`PatientContextHeader.tsx:197-211`）を突合 |
| S26-medical-record-image-capture.md | kept | `capture="environment"`・jpeg/png/gif vs +pdf/`multiple` の 2 input・`e.target.value=""` リセット・`disabled={isUploading}`・SEC-CS-F08（件数 10/1件 10MB/合計 50MiB fail-closed）を突合 |
| S27-chief-complaint-blank-persistence.md | kept | `chiefComplaintTypeId ?? null` hydrate（`use-apply-medical-record.ts:49`）・`chief_complaint_type_id: number \| null`（`use-medical-record-save-action.ts:106`）・`InterviewChiefComplaint` を突合。実 UI 解除は SLACK-COMPLAINT 残件のまま維持 |
| S28-manual-urine-examination.md | kept | `addManualItem`（`manual-N`）・`setInspectionValue`/`removeItem`（`use-examination-form-items.ts`）・`lab_import_examination_service_test.go` 非上書き回帰・origin 表示 PO ゲート・機器仕様未承認は BLOCKED のまま維持 |
| S29-same-day-multiple-vaccinations.md | kept | `useCreateVaccination` 単件 POST・`calculateNextDate` 独立計算・`use-medical-record-vaccination-form.ts:22` の `4weeks`。同日複数の既知調査 item（Plane）への証拠接続方針を維持（Slack 2026-09-16「予防を同日にした場合…2-3件同時」一次ソースと整合） |
| S30-shift-occupation-filter.md | kept | `filterStaffsByOccupation`・`OCCUPATION_FILTER_UNSET`・`occupationId == null` 未設定フィルタ・`ShiftCalendar.test.tsx` を突合（BRT-103 = Plane EMR-12 Done） |
| S31-detail-route-direct-access.md | kept | `paths.accounting.detail`・`paths.hospitalization.detail`/`edit`・`paths.inventory.detail`（`config/paths.ts`）を突合 |
| S32-owner-search-modal-fit.md | kept | `OwnerSearchModal` の `max-h-[calc(80vh-12rem)]`（test :61 で固定）・`isFiltering` opacity・Slack 2026-09-14 飼主検索スクロール一次ソースと整合 |
| S33-receipt-print-pdf.md | kept | 「明細兼領収書プレビュー」ダイアログ・`window.print()`・`effectiveOrder`/`sectionVisible`・登録番号 `print:hidden`・飼主名フォールバック BUG-374・物理印刷は外部レーンを突合（Slack 2026-09-16 領収書 PDF + 締め flow 一次ソースと整合） |
| S34-overlay-stack-fit.md | kept | `ui/dialog.tsx`・`select.tsx`・`popover.tsx`・`sonner.tsx` 実在。`BUG-DIALOG-FOCUS-RESTORE`（Plane EMR-64）回帰確認を維持 |
| S35-extreme-content-fit.md | kept | `PatientContextHeader` truncate 契約・`Intl.NumberFormat` 系の参照は現行と一致 |
| S36-table-kanban-horizontal-fit.md | kept | `DailyAccountingTab`/`AccountingListTable`/`CheckupsTabTable`/`VitalsTabTable`/`TreatmentsTab`/`AggregationOwnerTable` の `overflow-x-auto`/`min-w`・`KanbanColumn`/`AppointmentCard`/`ReceptionPagePanels` を突合 |
| S37-print-documents-layout.md | kept | `PrintPortal.tsx`・`features/manual/manual-print.css`・各 `*PrintArea.tsx` の `print:`/`@media print` を突合 |
| S38-liff-mobile-viewport.md | kept | LIFF `min-h-screen`・`liff/index.html` viewport meta・実 LINE WebView は別外部レーンの扱いを突合 |
| S39-state-feedback-visibility.md | kept | `sonner` Toaster（`app/provider.tsx`）・ボタン `min-h-11 min-w-11`/`disabled:pointer-events-none`/`focus-visible:ring`・`disabled={isPending}` 配線を突合。手動の統合視認受入のまま維持 |

## 実施した訂正（抜粋・重要度順）

- `S17-treatment-quantity-commit.md`: 「合成テスト 28 件」→ 数量確定操作の合成テストは 9 件（`TreatmentQuantityCell.test.tsx` の "Enter×2 / Blur / Escape" describe 6 件＋`TreatmentRow.test.tsx` の数量 Enter×2/Blur/Escape 3 件。TreatmentsTab 配下 5 ファイル総数 44 件だが確定操作を直接検証するのは 9 件。根拠: 各ファイルの `it(` 実数）。併せて「Linear 追跡」→ Plane 追跡
- `S19-medical-record-viewport-fit.md`: sidebar 開閉の実体を「`use-sidebar`/`SidebarProvider`」→ `Sidebar.tsx` ローカル `useState`＋`matchMedia`・`design-tokens.ts` `LAYOUT.sidebar`（expanded `w-[220px]`/collapsed `w-[56px]`）（根拠: `frontend/src` 内に該当フック/Provider の実在なし、`Sidebar.tsx:28-84`）
- `S16-interview-history-navigation.md`: 「Linear 追跡」→ Plane 追跡（Linear 2026-09-16 閉鎖）
- `README.md`: 「Linear Issue 化は後続レーン」→ Plane item 化。「実施レーンは BRT-68」に Plane `EMR-22` を併記（対応表: `docs/work/plane-md-migration-20260923-receipt.md`）
- `UAT-254-CLOSE-CHECKLIST.md`: 5 箇所 — Rules の「Linear/GitHub」→ Plane/GitHub、owner comment の linear.app URL（BRT-45/BRT-68）→ `BRT-45`（Plane `EMR-43`）・`BRT-68`（Plane `EMR-22`）、close gate の「Linear に受容条件」→ Plane、mermaid の「Linear 受容条件」→ Plane
- `S02-exam-abnormal-highlight-lock.md`: HIGH/LOW バッジの実体を `ExamItemsTable` → `ExamStatusBadge.tsx`（`ExamItemsTable.tsx:13,205` が import して描画）に精密化

訂正はすべて `(2026-09-29 訂正: <旧主張> → <新主張>)` 形式で行内または実装突合に記録。歴史的 rationale は保持した。

## 外部依存・BLOCKED（未解消のまま残すもの）

- **実 LINE/LIFF lane**: `VITE_LIFF_MOCK` は success-only。実 LINE idToken 検証・409 再連携・期限切れ 400・実 WebView viewport（S38）は mock では代替せず外部レーンのまま（S04/S12/S38・UAT-254 flow 4）。
- **L-step 実送信**: 秘密情報は write-only/マスク、実 API 送信は外部ゲート（V05・S01 のタグ再同期は best-effort 記録のまま）。
- **検査機器（LAB_DEVICE_CLIENT_UAT）**: NX600/AU10V のみ対象。PU-4010（2400 7E1/8E1 条件未確定）・IDEXX（ACK/session/retry 未確定）は PASS にしない。実機・2 本 USB・サポート立会いは外部ゲート。
- **S21 実並行**: `UAT-R2-EXCLUSIVE-LOCK` の 2 セッション実 DB 確認は UNKNOWN のまま（合成回帰の存在を PASS に昇格しない）。`billing-schema-readiness` の DB 制約ゲートが前提条件。
- **S28 手動尿検査**: 医院承認の項目・定性表現・単位・基準値が正本（未提供なら BLOCKED）。origin 表示の採否は PO 決定ゲート、機器連携は BRT-94〜100 系 lab lane。
- **S22 SLACK-STAFF-SELECT**: iPad・報告のあった PC 相当での選択→保存→再読込の実機受入が残件（デスクトップ Chrome PASS で代替しない）。
- **S27 実 UI 解除**: `SearchableSelect` clearable での区分解除が実 UI で到達可能かは `SLACK-COMPLAINT` 残件。mock の空値 callback で済ませない。
- **S23 SLACK-CROSS-CLINIC**: 他院患者の検索・受付境界は PO 裁定待ち（現行拠点スコープの受入のみを本シナリオ範囲とする）。
- **S29 同日複数予防接種**: 現状差分は既存 Plane 調査 item へ証拠接続。新規 duplicate issue を作らない。
- **UAT-254 close**: local/mock のみでは close 不可（実 LINE・token health・DB/audit・residual disposition・別 acceptance owner sign-off・USER 明示受容）。
- **V04 予約枠 nested-form 既知バグ**: `ReservationTypeSidePanel` が form 内に form をネスト（`ReservationTypeAvailableSlotsSection` の注記どおり HTML 上無効）。retest spec は該当ケースを API で代替。専用 spec は UI 経路を通過。製品バグとして記録済みの扱いを維持。
- **V05 パスワード長の FE/BE 乖離**: FE `minLength=6`（`noValidate` で submit 非阻止）vs BE `min=8`。文書は既に重点確認として記載。統一の製品判断は本監査範囲外。
- **runtime PASS の不在**: 全シナリオの手順・期待結果は現行実装と整合するよう訂正したが、実行結果・受入記録は `reports/uat-YYYY-MM-DD/`（ignored）に属するため scenario source には記録しない。ソース照合を実行証拠にしていない。

## 提案削除候補（理由・後継リンク）

- なし。全ファイルは索引・プロトコル・確認票としての役割を現行構成内で保持。

## 検証

- `git diff --check -- docs/ops/testing/scenarios` → exit 0（whitespace error なし）
- `bash scripts/check-docs-symbol-drift.sh` → exit 1（FAIL 3 件「テーブル数: docs 宣言 128 vs 実測 130」）。**base commit（stash 適用前）でも同一 3 件が出る既存ドリフト**で、宣言面は `docs/spec/specification.md`（:22,:61）と `docs/architecture/erd.md` — いずれも本区画 allowlist 外（spec-core/arch の担当範囲）。scenarios/** 配下の編集が原因のドリフトではないことを stash 検証で確認。本区画に起因する新規 FAIL なし。
- `git status --porcelain` → `docs/ops/testing/scenarios/` 6 ファイル + 本レポートのみ（allowlist 外差分なし）
- 索引とファイル集合: `ls` 実体 49 ファイルと README 索引（S01–S39・V01–V05・4 shared）が一致
- クライアント一次ソース突合: `13_UAT-R5確定仕様_2026-08-22.md`（BUG-017 実名 409 → V04 §2 手順 4 と整合）、`会話ログ/Slack_電子カルテ開発_曽我/2026-08-20以降_不具合等Slack全文.md`（担当者選択 iPad 9-13・飼主検索スクロール 9-14・主訴空欄/バイタル位置/マイクロチップ/同日複数予防/写真撮影/検査受信/領収書 PDF+締め 9-16・他院カルテ可視性 9-13・旧システム排他 8 月 → 各 SLACK-* キューと S04/S16-S19/S22-S33 の根拠と整合）
