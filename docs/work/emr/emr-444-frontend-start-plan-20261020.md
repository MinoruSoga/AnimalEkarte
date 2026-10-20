# EMR-444 配下フロント寄り票の着手プラン（2026-10-20）

> 位置づけ: 計画整備のみ。実装・試験実行・担当割当・Plane 状態変更は行っていない。
> 実行 SoT は Plane。本書は判断根拠の保存であり、仕様の疑義・変更提案は末尾「ChatGPT Dots への確認事項」として返す。
> 観測 SHA: `origin/main` d7da4a8d8（2026-10-20）。Plane は 2026-10-20 11:50 UTC 時点の本文・コメント。

## 0. 現状スナップショット

| 票 | 件名（要約） | Plane 状態 | 最新追補 | 親 |
|:--|:--|:--|:--|:--|
| EMR-474 | 【S1 UI】LIFF の頭別再診回答と訂正・再送 UI | Ready | 06:09「PR 完結粒度・API 契約先行」 | EMR-460 |
| EMR-475 | 【S1 reader】スタッフ予約詳細へ再診原文・旧 target 履歴を表示 | Ready | 06:09 同上 | EMR-460 |
| EMR-477 | 【S2 reader】治療プランへ予約回答原文 reader を接続し編集操作から保護 | Ready | 03:16 着手プラン監査、06:09 同上 | EMR-460 |
| EMR-487 | 【証明書出力】固定発行版から 3 様式の発行・再印刷を接続 | Ready | 03:18 着手プラン監査 | EMR-444 |
| EMR-463 | 飼主一覧の認可内医院初期値と選択保持を接続 | Ready | 06:25「飼主一覧専有 PR の完了条件」 | EMR-444 |

先行票の状態: EMR-471 / 472 / 473 / 486 は Ready、EMR-476 は Backlog。いずれも未実装。

追補で確定した共通の着手方針（5 票共通）:

- 先行 backend の実装完了は FE 着手ゲートではない。`backend/docs/api.yaml` に対象 API 契約（DTO・成功/競合/失権/結果不明の例）を固定し、採用 SHA を固定すれば FE 専有の「契約モック準拠 PR」を完結できる。
- FE PR は MSW 契約モックで受入条件を検証して閉じる。実 API 結合・親機能受入は統合担当が別に回収する。
- 新規子票・assignee 変更・共通 hook 全 caller 改変は行わない。

### 0.1 コードベースの事実（計画の前提）

| 事実 | 影響 |
|:--|:--|
| `backend/docs/api.yaml` に再診回答（revisit-intake）・証明書（certificate）の route/DTO は未収録。`GET /owners` に `clinics` 系パラメータも未記載 | 474/475/477/487 の「api.yaml 固定 → 採用 SHA」ゲートが未成立。463 は既存 `GET /pets?clinic_ids` を使うため契約追記は小さい |
| backend に revisit-intake / certificate の handler・service・route は無い。`internal/model/storage_clinical.go` の保存 model と `001_init.sql` の表・trigger のみ | FE は保存 model を DTO と見なさず、契約モックで進める |
| LIFF（`frontend/line-reserve/`）に予約詳細画面は無い。`MyReservationsPage` はカード一覧 + 行内キャンセルのみ。予約一覧 DTO（`Reservation`）は `pet_name` 単一文字列で pet ID・複数頭の列挙が無い | 474 の「予約詳細から回答へ進む」入口と、頭別 target の取得方法を契約で決める必要あり |
| LIFF は router 無し（`PageType` で画面切替）。永続化は一切無し（URL は `/line-reserve/{clinicId}` のみ、localStorage 未使用）。API adapter は `src/api/liff-api.ts`（axios + zod `parseOrThrow`）。submit は `useActionState`、409 は `slotTakenResponseSchema` で解釈。操作 key・結果照会は未実装 | 474 の draft 保持・結果照会は新設 |
| スタッフ予約詳細は `features/reservations/components/ReservationDetailModal.tsx`（`useGetReservation` → `GET /v1/reservations/:id`）。再診回答表示は無し | 475 の表示欄を追加する位置が確定 |
| `MedicalRecordClinicalTabs.tsx` の「診察/治療プラン」タブは `MedicalRecordDiagnosisPlan.tsx`（`DiagnosisHeader` → `ClinicalPlanSection` → 治療プラン表）。FE に「入力内容クリア」「複製取消（Undo）」の実装はまだ無く、複製は `MedicalRecordInterview.tsx` の `applyCopySource` + `onCopyRecordTreatments` | 477 は「将来のクリア/Undo の対象外」を構造で保証する（書込経路を持たない read-only 部品） |
| FE に証明書関連コードは無い（preview/renderer/template/API いずれも無し）。印刷基盤は `MedicalRecordPrintView.tsx` + `window.print()`（汎用カルテ印刷） | 487 は専用 renderer を新設し、汎用印刷を流用しない |
| `use-clinic-scope.ts` は URL `?clinics=` のみで状態を持ち、未指定時は `currentClinicId` 一件。失権医院の除外は server 側の 403 に依存。`OwnersList.tsx` は loader（`GET /v1/pets?clinic_ids=…`）経由 | 463 の仕様差分（初回全認可医院・次回保持・失権除外）は FE adapter で閉じられる |

## 1. 推奨着手順序

| 順 | 票 | 理由 |
|:--|:--|:--|
| 1 | EMR-463 | 既存 API で完結。契約追記が最小。1 PR |
| 2 | EMR-477 | read-only 1 欄。DTO 契約 1 本を固定すれば MSW で閉じる |
| 3 | EMR-475 | 477 と同じ原文 DTO 形を共有できる（履歴・旧 target 分を追加） |
| 4 | EMR-474 | GET / POST / 結果照会 の 3 契約と新画面。LIFF 入口の仕様確認（§6 Q2）を先に閉じる |
| 5 | EMR-487 | 486 の edition DTO と renderer 所在（§6 Q7）の確認後。golden 準備は先行可 |

477 と 475 の原文表示部品（7 項目 read-only・版・受信時刻・出所）は共通化候補だが、所有境界（medical-records / reservations）が異なるため、まず各 feature 内に置き、2 票目で `components/shared/` へ昇格するかを判断する。

---

## 2. EMR-474 LIFF 頭別再診回答・訂正・再送 UI

### 2.1 画面構成（提案）

| 画面 | 役割 |
|:--|:--|
| `CompletePage`（step8） | 「事前問診に回答する（任意）」CTA を追加。未回答でも「予約確認・キャンセル」「新しい予約」は従来どおり |
| `MyReservationsPage` | `confirmed` カードに「問診を回答する／訂正する」ボタン（状態に応じた文言）。取消済みカードは「回答履歴を見る」（履歴があるときのみ） |
| 新規 `RevisitIntakePage`（`PageType` に `revisit-intake` 追加） | 予約 1 件分。頭別のアコーディオン/セグメント切替で各ペットの 7 項目 form。履歴（旧 revision・旧 target）は別セクションで read-only |

LIFF に予約詳細画面が無いため「予約詳細から回答へ進む」は一覧カードの行内 CTA で代替する（§6 Q2）。

### 2.2 画面状態

| 状態 | 表示・操作 |
|:--|:--|
| 読込中 | `Spinner`。300ms 超はスケルトン |
| 未回答（target あり・answer 0 件） | 空 form。「回答は任意です」注記。送信ボタンは常時活性（未入力理由で disabled にしない） |
| 回答済み・受付前 | 最新 answer を form に prefill。「訂正して再送信」。回答日時・最終訂正日時を表示 |
| 回答済み・他 author が head（`content_available=false`） | 「回答済み」バッジのみ、本文非表示・prefill なし。「新しい回答を送信」は空 form から（期待版は `head_revision`） |
| 受付後 / 診療中 / 会計 / 完了 | 閲覧のみ。「受付後の変更はスタッフへお伝えください」 |
| 取消済み・旧 target | 履歴セクションに read-only 表示。現 target の入力へコピーしない |
| 送信中 | ボタン `aria-busy`・二重送信抑止。入力は保持 |
| 送信成功 | 成功メッセージ、最新 revision・受信時刻を再表示。`liff.sendMessages` は行わない（通知副作用を追加しない） |
| 409 `REVISION_CONFLICT` | 入力保持。「別の端末で更新されています。最新を確認して再送信」→ 再取得ボタン。再送信は新しい操作 key |
| 409 `KEY_CONTENT_MISMATCH` | 入力保持。内部不整合として再試行は新 key で |
| 通信断・応答喪失 | 自動で結果照会へ。`succeeded` → 成功表示、`claimed_uncommitted` / `not_found` → 同 key で再送、`result_unavailable` → 「結果を確認できません。時間をおいて開き直してください」（新規送信しない） |
| 401 | 既存 `resolveFetchError`（再ログイン、`canRetry=false`） |
| 403 / 404（失権・他 scope） | `ErrorPage`。receipt の有無を推測する文言を出さない |
| 目的が空白のみで送信 | 欄直下にエラー「来院目的を入力してください」、先頭エラー欄へフォーカス（Unicode 空白も FE で弾くが正本は server） |

### 2.3 操作手順（主系）

1. 完了画面 or 一覧カードから `revisit-intake` へ遷移（予約 ID を state で渡す。URL には載せない）。
2. `GET …/revisit-intake` で target 一覧（頭別）・各 head_revision・最新本文（可視なら）・受付状態・編集可否を 1 応答で取得。
3. ペット切替タブで各頭の draft を編集。draft は `Map<targetKey, Draft>`（`clinic|appointment|pet|generation`）でメモリ保持。他頭へ混ぜない。
4. 送信: `crypto.randomUUID()` で操作 key を採番 → `POST …/submissions`（header に key、body は `target{pet,generation}`・`expected_revision`・`form_definition_version=1`・7 項目のみ）。
5. 応答喪失時は同 key で `GET …/operations/:key` を照会し、結果に従って再送・成功表示・保留を分岐。成功後の訂正は新 key。
6. 成功後は GET を再実行し、prefill と履歴を更新。

### 2.4 API 対応（候補 route。正本は 471 の API owner が `api.yaml` へ固定）

| 用途 | 候補 | FE 期待 DTO の要点 |
|:--|:--|:--|
| 取得 | `GET /api/liff/:clinicId/my-reservations/:id/revisit-intake` | `targets[]{pet_id, pet_name, target_generation, head_revision, latest_answer|null, content_available, editable, history[]}`, `appointment_state`, `form_definition_version` |
| 送信 | `POST /api/liff/:clinicId/my-reservations/:id/revisit-intake/submissions` + `Idempotency-Key: <uuid>` | `{target:{pet_id,target_generation}, expected_revision: string|null, form_definition_version:1, visit_purpose, onset, energy, appetite, current_symptoms, improvement, examination_preference, main_concern}` → `{operation_result, current_state}` |
| 結果照会 | `GET /api/liff/:clinicId/revisit-intake/operations/:key` | `revisit-intake.md §5.4` の固定形（redacted 成功形を含む） |

選択肢 code は `revisit-intake.md §4` の allowlist（`usual/slightly_reduced/markedly_reduced` 等）を zod enum で固定し、未知 code は送らない。

### 2.5 変更候補ファイル

| 区分 | ファイル |
|:--|:--|
| 新規 | `frontend/line-reserve/src/pages/RevisitIntakePage.tsx`（+ `.test.tsx`） |
| 新規 | `frontend/line-reserve/src/components/revisit-intake/PetAnswerForm.tsx`、`AnswerHistory.tsx`、`SubmitResultBanner.tsx` |
| 新規 | `frontend/line-reserve/src/hooks/use-revisit-intake-drafts.ts`（頭別 draft Map、`.test.ts`） |
| 新規 | `frontend/line-reserve/src/hooks/use-revisit-intake-submit.ts`（操作 key・結果照会・再送分岐、`.test.ts`） |
| 新規 | `frontend/line-reserve/src/lib/revisit-intake-options.ts`（code→ラベル、v1 固定） |
| 変更 | `frontend/line-reserve/src/api/liff-api.ts`（3 メソッド）、`src/api/schemas.ts`（zod）、`src/types/models.ts`（`PageType`・DTO 型） |
| 変更 | `src/app-page-switch.tsx`、`src/app-flow-handlers.ts`、`src/pages/CompletePage.tsx`、`src/pages/MyReservationsPage.tsx`（+ 各 test） |
| 試験 | MSW handler（`frontend/src/testing/` 配下の既存方式に合わせる）で成功 / 409 両種 / 応答喪失→照会 4 状態 / 403 を fixture 化 |

### 2.6 依存関係

- 契約: EMR-471（submit/GET DTO・route）、EMR-472（認可・`content_available` の意味）、EMR-473（操作 key・結果照会 4 状態）。実装完了は不要、`api.yaml` 固定が条件。
- 既存再利用: `useFetchState` / `resolveFetchError` / `ErrorPage` / `Spinner` / `sanitize.ts`。
- 非対象: 予約・受付・通知・カルテ作成の発火、`localStorage`/URL への本文・token 保存。

### 2.7 PR 単位

| PR | 内容 | 完結条件 |
|:--|:--|:--|
| 474-A 契約層 | zod schema・adapter 3 メソッド・MSW fixture・`PageType` 追加（UI なし） | adapter test で 3 契約の正常/異常を網羅 |
| 474-B 画面 | `RevisitIntakePage` + 頭別 draft + 入口 CTA | AC-optional / multi-pet / revisions / new-target の画面状態を test で再現 |
| 474-C 再送・照会 | 操作 key・結果照会・競合 UI | 応答喪失→4 状態、409 両種、二重送信抑止 |

474-A と B は 1 PR に統合してもよい。統合担当（本票の FE 担当一人）が 471 へ結合条件と証跡を渡す。

---

## 3. EMR-475 スタッフ予約詳細の再診原文・旧 target 履歴

### 3.1 画面状態（`ReservationDetailModal` 内に新セクション「飼主事前問診（LINE）」）

| 状態 | 表示 |
|:--|:--|
| 読込中 | セクション内スケルトン（モーダル本体は既存どおり） |
| 未回答（現 target に answer 無し） | 「未回答」バッジ（文字 + アイコン）。空状態 3 点（何が無いか／条件／次の操作=「受付時に聞き取り」） |
| 回答済み | 最新原文 7 項目 read-only、回答日時・最終訂正日時、出所「飼主（LINE）」。訂正履歴は折りたたみ（revision・送信日時・本文） |
| 旧 target あり（pet 付替え後） | 「以前の対象ペットの回答」として別グループ。現 target と視覚的に分離（見出し + ペット名 + 世代） |
| 取消済み / soft delete 後 | 履歴 reader で取得できる範囲を同じ形式で表示。「取消済み予約」注記 |
| 閲覧権限なし（403） | セクション自体を「閲覧権限がありません」1 行に（本文・件数を漏らさない） |
| 取得失敗 | セクション内 `再試行`。他の予約情報は表示継続 |
| 他院予約 | 既存 `clinicId` 伝搬に従い、セクションも同じ clinic で取得 |

スタッフ補足は既存の予約 notes（「予約内容」）/ Inquiry へ。原文欄に編集 UI・コピー UI を置かない。

### 3.2 操作手順

1. 予約一覧からモーダルを開く → 既存 `useGetReservation` と並行して `useGetReservationRevisitIntake(id, clinicId)`。
2. 履歴折りたたみ展開 / 旧 target グループ展開（キーボード到達可）。
3. モーダル閉鎖・別予約へ切替で query を破棄（キャッシュ key に `id|clinicId` を含める）。

### 3.3 API 対応

| 用途 | 候補 | DTO 要点 |
|:--|:--|:--|
| 取得 | `GET /v1/reservations/:id/revisit-intake`（既存 `reservations:view` grant） | `current_target{pet, generation, head_revision, latest_answer, revisions[]}`, `previous_targets[]`, `appointment_state`, `visibility` |

### 3.4 変更候補ファイル

| 区分 | ファイル |
|:--|:--|
| 新規 | `frontend/src/features/reservations/components/RevisitIntakeSection.tsx`（+ `.test.tsx`） |
| 新規 | `frontend/src/features/reservations/api/get-revisit-intake.ts`、`lib/transforms/revisit-intake.ts` |
| 変更 | `components/ReservationDetailModal.tsx`（`InfoRow` 群の後にセクション挿入）、`ReservationDetailModal.test.tsx` |
| 整形 | `lib/format` / `lib/jst-date` の共有関数（日時・空値 `—`） |

### 3.5 依存・PR 単位

- 契約: EMR-471（原文 DTO）、EMR-472（staff view grant・行可視性）。
- FE PR 1 本（セクション + adapter + MSW test）。BE reader PR は別 owner。統合担当が両 PR の採用 SHA と実 API 結合を回収。

---

## 4. EMR-477 治療プランの予約回答原文 reader

### 4.1 配置と画面状態

配置: 「診察/治療プラン」タブ内、`DiagnosisHeader` の直下・`ClinicalPlanSection` の上に独立カード「飼主事前問診（LINE・読み取り専用）」。

| 状態 | 表示 |
|:--|:--|
| 新規カルテ（`isNewRecord`） | カード非表示（binding はカルテ保存時に確定するため） |
| binding あり・回答あり | 7 項目 read-only、回答版、受信時刻、出所「飼主（LINE）」。`readonly` 表現（ink を落とさない） |
| binding あり・未回答 | 「未回答」バッジ + 空状態 |
| binding なし（`revisit_intake_target_id` NULL） | 「対応する予約回答を特定できません」。未回答と文言を分ける |
| カルテの患者/予約と target 不一致 | 「出所不明」として本文を出さない |
| 失権（403） | カード非表示 + 1 行注記。再読込でも書換えなし |
| 取得失敗 | カード内 `再試行` |
| 患者切替 | query key `recordId|recordClinicId|petId` で破棄。前患者の本文を残さない |

### 4.2 書込経路からの保護（構造で担保）

- 部品は props に `onChange` 系を持たず、`use-medical-record-diagnosis-state.ts` の状態へ接続しない。
- `applyCopySource` / `onCopyRecordTreatments` / 将来の「入力内容クリア」「複製取消」は対象 state を持たないため影響しない。test で「複製実行後も原文 DTO・表示が不変」を固定。
- 診断コピー adapter の複数行移行（EMR-495）には触れない。

### 4.3 API 対応

| 用途 | 候補 | DTO 要点 |
|:--|:--|:--|
| 取得 | `GET /v1/medical-records/:id/revisit-intake`（現在 grant） | `binding: "bound"|"unbound"|"mismatch"`, `answer|null{revision, received_at, provenance, fields…}` |

### 4.4 変更候補ファイル

| 区分 | ファイル |
|:--|:--|
| 新規 | `frontend/src/features/medical-records/components/OwnerIntakeReadOnlyCard.tsx`（+ `.test.tsx`） |
| 新規 | `frontend/src/features/medical-records/api/revisit-intake.ts`（`useGetRecordRevisitIntake`） |
| 変更 | `components/MedicalRecordDiagnosisPlan.tsx`（カード挿入）または `MedicalRecordClinicalTabs.tsx`（タブ slot）、`lib/medical-record-tabs-types.ts`（props 追加が要るとき） |
| 試験 | `MedicalRecordClinicalTabs.test.tsx` に「複製・患者切替・失権で原文不変」ケース追加 |

### 4.5 依存・PR 単位

- 契約: EMR-476（binding 固定の意味）、EMR-472（read 認可）。489 全 owner 完了は条件にしない。
- FE PR 1 本。478 へ reader 境界の証跡（test 名・fixture）を渡す。

---

## 5. EMR-487 接種証明書 3 様式の preview・発行・再印刷

### 5.1 画面構成（提案）

| 画面 | 役割 |
|:--|:--|
| 入口 | `VaccinationForm`（`/vaccinations/:id`）と `MedicalRecordVaccination` の既存接種行に「証明書」ボタン |
| 新規 `/vaccinations/:id/certificate` | `PageLayout` shell。左: 様式選択（通常=縦/横、狂犬病は別 series）、右: preview。下部: 「発行」「再印刷」「履歴」 |
| 印刷用 | 専用 renderer（`@media print` で shell を隠す）または server artifact の表示。所在は §6 Q7 |

新規リーフルートは同一コミットで `docs/spec/ui-design-compliance.md §2` の表を更新する。

### 5.2 画面状態

| 状態 | 表示・操作 |
|:--|:--|
| 未発行（edition 無し） | 現在接種内容からの preview（「未発行」バッジ）。操作は「発行」のみ。preview 省略不可 |
| 最新版あり | 最新 edition snapshot の preview。「再印刷」。縦/横切替は同一 edition を再描画し、新版を作らない |
| 訂正済み旧版 | 履歴一覧に read-only。「有効な証明書ではありません」注記。再印刷・直接 artifact 取得ボタンなし |
| 接種が訂正済みで edition が旧 | 「接種内容が更新されています。新しい版を発行してください」。再印刷は最新 edition（旧内容）を再現し、現在値で再構築しない |
| 発行中 / 再印刷中 | `aria-busy`。操作 key を保持 |
| render / upload 失敗 | 同一操作 key で再試行。旧版は復活しない |
| 結果不明 | 操作照会 → 成功なら artifact を開く、不明なら再試行案内 |
| 409（同時訂正） | 再読込して最新版を表示 |
| 403 | ページ全体を権限なし表示 |
| 発行完了 | receipt 表示。「印刷完了」とは表示しない（保存 receipt ≠ 物理印刷） |

### 5.3 操作手順

1. 接種行 → 「証明書」→ `/vaccinations/:id/certificate`。
2. `GET editions` で series/最新 edition/履歴/有効性を取得。
3. 様式選択 → preview。
4. 「発行」: `POST operations`（`operation_kind=issue`、UUID key）→ edition 確定 → artifact 取得 → 印刷ダイアログ。
5. 「再印刷」: `POST operations`（`reprint`、UUID key、`edition_id` 指定）→ artifact 取得 → 印刷ダイアログ。
6. 履歴から旧版を開く → read-only 表示のみ。

### 5.4 API 対応（候補。正本は 486 が固定）

| 用途 | 候補 | DTO 要点 |
|:--|:--|:--|
| 取得 | `GET /v1/vaccinations/:id/certificate-editions` | `series[]{kind, current_edition{id, edition, snapshot, source_vaccination_version, issued_at}, history[]}`, `effective` |
| 発行/再印刷 | `POST /v1/vaccinations/:id/certificate-operations` + `Idempotency-Key` | `{kind, layout, edition_id?}` → `{operation_result, artifact_ref}` |
| artifact | `GET /v1/certificate-artifacts/:id`（現在 grant・有効版検査） | バイナリ or snapshot JSON |
| 照会 | `GET /v1/certificate-operations/:key` | 4 状態 |

### 5.5 変更候補ファイル

| 区分 | ファイル |
|:--|:--|
| 新規 | `frontend/src/features/vaccinations/certificate/CertificatePage.tsx`、`CertificatePreview.tsx`、`layouts/PortraitCertificate.tsx`、`layouts/LandscapeCertificate.tsx`、`layouts/RabiesCertificate.tsx`、`certificate-model.ts`（snapshot→表示値、LOT1〜4 / E-Mail 等の様式差を保持） |
| 新規 | `certificate/__fixtures__/`（合成値のみ、3 様式 golden）、`*.test.tsx`（snapshot + 項目一対一表）、`*.stories.tsx`（Chromatic 用） |
| 新規 | `features/vaccinations/api/certificate-editions.ts`、`certificate-operations.ts` |
| 変更 | `config/paths.ts`、`app/routes/*`（ルート登録）、`routes/VaccinationForm.tsx` / `medical-records/components/MedicalRecordVaccination.tsx`（入口ボタン）、`docs/spec/ui-design-compliance.md §2` |
| 整形 | 日付・金額は `lib/format`。金額・番号は snapshot の値をそのまま（推測・自動計算なし） |

### 5.6 依存・PR 単位

| PR | 内容 | 完結条件 |
|:--|:--|:--|
| 487-A renderer + golden | 3 様式 layout 部品・model・fixture・snapshot test・stories（API なし） | 原見本との項目一対一表を PR 本文に添付 |
| 487-B preview/発行/再印刷 | ページ・route・adapter・操作 key・結果照会・MSW | 拒否試験（失権・旧版・同時訂正）を test で固定 |
| 487-C 履歴・訂正連動 | 旧版 read-only、接種訂正後の案内 | AC-correction / basic-update |

依存: EMR-486 の edition reader 契約（mock で先行可）。実印刷・法令適合は未実施のまま分離。

---

## 6. EMR-463 飼主一覧の認可内医院初期値と選択保持

### 6.1 画面状態

| 状態 | 表示・挙動 |
|:--|:--|
| 単一医院 | 変更なし（`ClinicScopeFilter` 非表示） |
| 複数医院・初回（保存済み選択なし・URL `clinics` なし） | 認可内全医院を選択し、`clinic_ids` を明示送信 |
| 保存済み選択あり | 復元。失権医院は除外。除外後 0 件なら全認可医院へフォールバック |
| 明示選択 | URL `?clinics=` を更新（既存）、同時に保存先へ書く。`page` リセット（既存） |
| 一覧→詳細→戻る | URL 履歴で検索語・医院・ページを保持（既存。メニュー再入場と区別） |
| 遅延応答 | loader の順序保証に依存。ナビゲーション外 fetch を足す場合は request token で古い応答を破棄 |
| 403（直接 API で認可外） | エラー表示 + 選択を認可内へ補正 |

### 6.2 操作手順

1. メニューから飼主一覧 → adapter が「URL > 保存済み選択 ∩ 認可内 > 全認可」で初期 `clinics` を決め、必要なら URL に `replace` で反映。
2. `ClinicScopeFilter` で切替 → URL + 保存先へ反映。
3. 詳細へ → 戻る → URL から復元。

### 6.3 API 対応

既存 `GET /v1/pets?clinic_ids=…`（`ResolveListClinicIDs` が未割当 ID を 403）。`api.yaml` の `/owners`・`/pets` に `clinic_ids` 省略時の意味と 403 契約を追記するのは API owner。FE から新 route は不要。DB 差分なし（根拠: 既存 middleware と `FilterClinicIDsForPermission`）。

### 6.4 変更候補ファイル

| 区分 | ファイル |
|:--|:--|
| 新規 | `frontend/src/features/owners/hooks/use-owners-clinic-scope.ts`（`useClinicScope` を包む adapter、`.test.ts`） |
| 新規 | `frontend/src/features/owners/lib/owners-clinic-selection-storage.ts`（保存・復元・失権除外の純関数、`.test.ts`） |
| 変更 | `features/owners/routes/OwnersList.tsx`（hook 差替え）、`features/owners/loaders.ts`（`clinics` 未指定時の既定を全認可医院にするため、保存済み選択 + 認可一覧を loader でも参照） |
| 不変 | `hooks/use-clinic-scope.ts` 本体（他 3 caller の意味を変えない）。必要なら `defaultClinicIds?: () => number[]` の opt-in 引数のみ |
| 試験 | `OwnersList.test.tsx` / `OwnersList.filter-wiring.test.tsx` に初回全医院・保持・失権除外・復帰を追加 |

### 6.5 依存・PR 単位

- 契約: EMR-451 / 453 の該当 subset は「既存一覧 API の `clinic_ids` 契約」のみ。491 全 writer は待たない。
- FE PR 1 本。統合担当が実 reader 接続と直接 API 拒否の結合受入を回収。

---

## 7. 共通の試験・DoD

- スコープ限定 vitest（`docker compose exec frontend npx vitest run <path>`）と `pnpm design-audit` を各 PR で実行し、結果を PR 本文へ。
- 契約モックは MSW。fixture は合成値のみ（実患者・実飼主の値を含めない）。
- UI デザインガイドのハードルール（PageLayout shell、トークン色、readonly≠disabled、空状態 3 点、保存ボタンを未入力理由で disabled にしない、`lib/format` 整形）を適用。
- 未接続の実 API・実印刷・実機は PR 本文で UNKNOWN / BLOCKED / SKIP と明記。

---

## 8. ChatGPT Dots への確認事項（仕様の疑義・変更提案）

| # | 票 | 疑義・提案 | FE 側の暫定前提 |
|:--|:--|:--|:--|
| Q1 | 474/475/477/487 | 追補は「`api.yaml` に契約を固定し採用 SHA を固定すれば FE 着手可」とするが、現 `api.yaml` に revisit-intake / certificate の route・DTO は未収録。契約追記の owner（471 / 486）と時期、または FE PR 内で draft 契約を `api.yaml` へ同梱してよいかの判断が要る | FE は本書 §2.4 / §3.3 / §4.3 / §5.4 の候補 DTO で MSW fixture を先行作成し、`api.yaml` 確定後に差し替える |
| Q2 | 474 | LIFF に予約詳細画面が無い。仕様の「予約詳細から回答へ進む」を一覧カードの行内 CTA + 専用回答画面で代替してよいか | 代替案で進める |
| Q3 | 474 | `GET /my-reservations` の DTO は pet 単一文字列で pet ID・複数頭の列挙が無い。頭別 target は revisit-intake GET 側で列挙する契約でよいか（一覧 DTO は不変） | GET revisit-intake が `targets[]` を返す |
| Q4 | 474 | draft は仕様どおり URL / localStorage へ保存しないため、LIFF の再読込・バックグラウンド終了で未送信入力は失われる。これを許容するか、別の保持手段（server 側 draft）を今期対象にするか | メモリ保持のみ。失われる旨の UI 注記は付けない |
| Q5 | 477 | binding NULL（対応不明）と「未回答」を UI で分ける文言、および配置（診察/治療プラン タブの `ClinicalPlanSection` 上、新規カルテでは非表示）の確認 | §4.1 のとおり |
| Q6 | 475 | スタッフ閲覧で旧 target・他 author の revision をどこまで表示するか（LINE 本人向け allowlist とは別の staff grant で「全 revision」を想定）。取消済み予約の履歴取得は専用 history query（通常一覧の soft-delete 条件を解除しない）でよいか | staff は現在 grant 内で全 revision を表示 |
| Q7 | 487 | renderer の所在。`vaccination_certificate_artifacts`（ObjectKey・ContentSHA256）は server 側 render/保存を示唆するが、票は「本票 FE/renderer owner」とする。FE が HTML/CSS で描画して印刷するのか、server が PDF を生成し FE は preview/取得のみか。物理印刷方式（ブラウザ印刷ダイアログ可否）も未選定 | 487-A は描画部品を FE で作る（server 生成でも snapshot→表示の golden は再利用できる） |
| Q8 | 487 | 縦/横の様式切替 UI は「同一 edition の再描画」であり操作 receipt を作らない想定でよいか。再印刷のみ receipt を作る | 切替は receipt なし、再印刷は receipt あり |
| Q9 | 463 | 「次回保持」の保存先。ブラウザ localStorage（端末・ブラウザ単位、staff ID をキーに含める）か、server 側ユーザー設定か。前者は端末を跨がない | localStorage（`owners-list.clinics.<staffId>`）で先行 |
| Q10 | 463 | 票は `backend/internal/owner/http_routes.go` を対象 path とするが、飼主一覧の実データ取得は `GET /v1/pets`（`pet/routes.go`）。契約追記対象を `/pets` にも広げるか | FE は `/pets` の既存 403 契約に依存 |
| Q11 | 全票 | 生成型（`src/types/generated/models.ts`）は `make codegen` がユーザー実行のため、FE モック PR では手書き型 + zod で進め、`api.yaml` 確定後に生成型へ切替える運用でよいか | 手書き型で先行、切替 PR は統合担当 |
| Q12 | 475/477 | 「スタッフ補足」の保存先は既存の予約 notes（予約内容）/ Inquiry のままで、原文欄に補足入力を置かない理解でよいか | 新規入力欄を作らない |
