# EMR-444 配下フロント寄り票の着手プラン（2026-10-20・再点検版）

> 位置づけ: 計画整備のみ。実装・試験実行・担当割当・Plane 状態変更は行っていない。
> 実行 SoT は Plane。本書は判断根拠の保存であり、未決の業務判断だけを末尾「ChatGPT Dots への確認事項」として返す。
> 正本: 最新 main `docs/spec/domains/`（観測 SHA da27989fa、2026-10-20）。技術設計は `docs/design/`、共有契約は `docs/spec/shared/`。Plane は 2026-10-20 12:45 UTC 時点。
> 完了条件: 本人の最新指示（EMR-476 / 478 本文 2026-10-20 節）により、試験実行と隔離環境は着手・PR 作成の必須条件にしない。各 PR は実装・契約 fixture・未実行試験の NOT_RUN 記録で閉じ、まとめ検証は本人指示時に main のローカル環境で行う。

## 0. 現状スナップショット

| 票 | 件名（要約） | Plane 状態 | 最新追補 | 親 |
|:--|:--|:--|:--|:--|
| EMR-474 | 【S1 UI】LIFF の頭別再診回答と訂正・再送 UI | Ready | 06:09「PR 完結粒度・API 契約先行」 | EMR-460 |
| EMR-475 | 【S1 reader】スタッフ予約詳細へ再診原文・旧 target 履歴を表示 | Ready | 06:09 同上 | EMR-460 |
| EMR-477 | 【S2 reader】治療プランへ予約回答原文 reader を接続し編集操作から保護 | Ready | 03:16 着手プラン監査、06:09 同上 | EMR-460 |
| EMR-487 | 【証明書出力】固定発行版から 3 様式の発行・再印刷を接続 | Ready | 03:18 着手プラン監査 | EMR-444 |
| EMR-463 | 飼主一覧の認可内医院初期値と選択保持を接続 | Ready | 06:25「飼主一覧専有 PR の完了条件」 | EMR-444 |

先行票: EMR-471 / 472 / 473 / 486 は Ready、EMR-476 は Backlog。EMR-467（共通 gate）は Running。いずれも FE 着手を止めない（下記共通方針）。

### 0.1 共通の着手方針（追補で確定済み）

| 方針 | 根拠 |
|:--|:--|
| 先行 backend の実装完了は FE 着手ゲートではない。`backend/docs/api.yaml` の対象契約と採用 SHA を固定すれば FE 専有の契約モック PR を完結できる | 474 / 475 / 477 / 463 コメント 06:09〜06:25 |
| `api.yaml` への再診回答契約（送信・訂正・結果照会の request/response・認可前提・競合/失権/再送 error）の追記は EMR-471 の最初の作業。FE は draft 契約を `api.yaml` へ書かない | 471 コメント 06:09「最初に既存 backend/docs/api.yaml へ…整合し採用 SHA を固定する」 |
| 証明書の edition ID・snapshot・layout・template/renderer 版・有効性契約は EMR-486 backend owner が 487 へ引き渡す | 486 コメント 03:17 項目 4 |
| FE PR は MSW 契約モックで画面状態を固定して閉じる。実 API 結合・親機能受入は統合担当が別に回収。新規子票・assignee 変更・共通 hook 全 caller 改変はしない | 各票 06:09 追補 |
| 試験実行・隔離環境は PR 作成の必須条件にしない。未実行は NOT_RUN として記録し、他環境の結果を転用しない | 476 / 478 本文 2026-10-20 節 |

### 0.2 コードベースの事実

| 事実 | 影響 |
|:--|:--|
| `api.yaml` に revisit-intake / certificate の route・DTO は未収録。`GET /owners` `GET /pets` に `clinic_ids` 省略時の意味・403 契約も未記載 | 471 / 486 の契約追記待ち。FE は本書の候補 DTO で fixture を先行作成し、採用 SHA 固定後に差し替える |
| backend に revisit-intake / certificate の handler・service・route は無い。`internal/model/storage_clinical.go` の保存 model と `001_init.sql` の表・trigger・view のみ | 保存 model を DTO と見なさない（477 監査コメント） |
| LIFF（`frontend/line-reserve/`）の画面構成は トップ / step1〜8 / マイ予約 / メンテナンス で、予約詳細画面は無い（`docs/spec/screens/37-line-reserve-owner-flow.md` 画面構成）。`MyReservationsPage` はカード一覧 + 行内キャンセル。予約一覧 DTO は `pet_name` 単一文字列 | 474 の入口は「マイ予約カード = 予約詳細相当」として設計（§2.1） |
| LIFF は router 無し（`PageType` 切替）、永続化無し、API adapter は `src/api/liff-api.ts`（axios + zod）。操作 key・結果照会は未実装 | 474 で新設。入力はフロー内状態に保持し URL / ストレージへ永続化しない（screens/37「操作・状態・保存」末尾） |
| スタッフ予約詳細は `features/reservations/components/ReservationDetailModal.tsx`（`GET /v1/reservations/:id`）。再診回答表示は無し | 475 の挿入位置 |
| `MedicalRecordClinicalTabs.tsx` の「診察/治療プラン」タブは `MedicalRecordDiagnosisPlan.tsx`。FE に「入力内容クリア」「複製取消」は未実装、複製は `MedicalRecordInterview.tsx` の `applyCopySource` + `onCopyRecordTreatments` | 477 は書込経路を持たない read-only 部品として構造で保護 |
| FE に証明書関連コードは無い。印刷基盤は `MedicalRecordPrintView.tsx` + `window.print()`（汎用） | 487 は専用 renderer を新設（lifecycle「旧未保存印刷…を復活させない」） |
| ワクチンマスタ（`model/vaccine.go`）に証明書種別属性は無い。DB の `vaccination_certificate_series.certificate_kind` は `normal` / `rabies` | 狂犬病様式の対象判定は未決（§8 Q-A） |
| `use-clinic-scope.ts` は URL `?clinics=` のみ、未指定時は `currentClinicId` 一件。`OwnersList.tsx` は loader で `GET /v1/pets?clinic_ids=…`。backend は未割当 ID を 403（`httpapi/context.go` `ResolveListClinicIDs`） | 463 は FE adapter で閉じられ DB 差分なし |

## 1. 推奨着手順序

| 順 | 票 | 理由 |
|:--|:--|:--|
| 1 | EMR-463 | 既存 API で完結。保存媒体は技術選定で可（451 コメント）。1 PR |
| 2 | EMR-477 | read-only 1 欄。DTO 契約 1 本を固定すれば MSW で閉じる |
| 3 | EMR-475 | 477 と原文表示部品の形を共有（履歴・旧 target 分を追加） |
| 4 | EMR-474 | GET / POST / 結果照会の 3 契約と新画面。471 の契約追記が前提 |
| 5 | EMR-487 | 486 の edition 契約が前提。renderer / golden（487-A）は先行可 |

477 と 475 の原文表示部品（7 項目 read-only・版・受信時刻・出所）は共通化候補。所有境界が異なるため、まず各 feature 内に置き、2 票目で `components/shared/` へ昇格するかを判断する。

---

## 2. EMR-474 LIFF 頭別再診回答・訂正・再送 UI

### 2.1 画面構成

| 画面 | 役割 | 根拠 |
|:--|:--|:--|
| `CompletePage`（step8） | 「事前問診に回答する（任意）」CTA を追加。未回答でも既存 2 導線は従来どおり | FR「予約完了画面と予約詳細から回答へ進み」 |
| `MyReservationsPage` | `confirmed` カードに「問診を回答する／訂正する」（状態別文言）。取消済みカードは履歴があるときだけ「回答履歴を見る」 | LIFF に予約詳細画面は無く、マイ予約が「予約確認」を担う（screens/37 画面構成）。予約詳細相当として扱う設計判断 |
| 新規 `RevisitIntakePage`（`PageType` に `revisit-intake`） | 予約 1 件分。頭別セグメント切替で各ペットの 7 項目 form。履歴（旧 revision・旧 target）は別セクション read-only | FR 7 項目表、AC-multi-pet / revisions / new-target |

### 2.2 画面状態

| 状態 | 表示・操作 | 根拠 |
|:--|:--|:--|
| 読込中 | `Spinner`。300ms 超はスケルトン | ui-design-guide patterns §6 |
| 未回答（target あり・answer 0 件） | 空 form。「回答は任意です」注記。送信ボタンは常時活性 | AC-optional、ui-design-guide（未入力理由で disabled にしない） |
| 回答済み・受付前 | 最新 answer を prefill。「訂正して再送信」。回答日時・最終訂正日時 | AC-revisions |
| 回答済み・他 author が head（`content_available=false`） | 「回答済み」バッジのみ。本文非表示・prefill なし。新送信は空 form、期待版は `head_revision` | revisit-intake.md §5.4 |
| 受付後 / 診療中 / 会計 / 完了 | 閲覧のみ。「受付後の変更はスタッフへお伝えください」 | FR「受付後はスタッフへ連絡」、§5.3 |
| 取消済み・旧 target | 履歴セクションに read-only。現 target の入力へコピーしない | AC-new-target、§5.4 |
| 送信中 | `aria-busy`・二重送信抑止。入力保持 | ui-design-guide design-states |
| 送信成功 | 成功表示。最新 revision・受信時刻を再表示。`liff.sendMessages` は呼ばない | §5.2 手順 8「LINE 通知は回答送信から作らない」 |
| 409 `REVISION_CONFLICT` | 入力保持。「別の端末で更新されています」→ 再取得。再送信は新 key | §5.2 手順 6 |
| 409 `KEY_CONTENT_MISMATCH` | 入力保持。新 key で再試行 | §5.2 手順 2、operation-participation.md `operation-key-conflict` |
| 通信断・応答喪失 | 自動で結果照会。`succeeded` → 成功、`claimed_uncommitted` / `not_found` → 同 key で再送、`result_unavailable` → 「結果を確認できません」（新規送信しない） | §5.2 手順 8、§5.4 |
| 401 | 既存 `resolveFetchError`（再ログイン） | 現行 shared-liff |
| 403 / 404 | `ErrorPage`。receipt の有無を推測する文言を出さない | §5.4「別 scope へ receipt の存在を漏らさない」 |
| 目的が空白のみで送信 | 欄直下にエラー、先頭エラー欄へフォーカス。正本検査は server | §4「空白のみを service で Unicode-aware に拒否」 |
| 再読込・LIFF 終了 | 未送信入力は失われる（メモリ保持のみ） | 474 受入条件「回答本文や token を URL/localStorage へ保存しない」、screens/37「URL/ストレージへ新たに永続化しない」 |

### 2.3 操作手順

1. 完了画面 or マイ予約カードから `revisit-intake` へ（予約 ID は state で渡す。URL に載せない）。
2. `GET …/revisit-intake` で頭別 target・各 `head_revision`・最新本文（可視なら）・受付状態・編集可否を 1 応答で取得（reservations technical-spec「LINE read/submit adapter」: pet 別 target・フォーム版・最新回答版を返す）。
3. ペット切替で各頭の draft を編集。draft は `Map<targetKey, Draft>`（`clinic|appointment|pet|generation`）でメモリ保持。
4. 送信: `crypto.randomUUID()` で操作 key → `POST …/submissions`（header に key、body は target・`expected_revision`・`form_definition_version=1`・7 項目のみ。clinic / actor / time / status は禁止属性）。
5. 応答喪失は同 key で `GET …/operations/:key`。成功後の訂正は新 key。
6. 成功後に GET を再実行し prefill と履歴を更新。

### 2.4 API 対応（候補。正本は EMR-471 が `api.yaml` へ固定）

| 用途 | 候補 route（revisit-intake.md §8） | FE 期待 DTO |
|:--|:--|:--|
| 取得 | `GET /api/liff/:clinicId/my-reservations/:id/revisit-intake` | `targets[]{pet_id, pet_name, target_generation, head_revision, latest_answer|null, content_available, editable, history[]}`, `appointment_state`, `form_definition_version` |
| 送信 | `POST …/revisit-intake/submissions` + UUID header | → `{operation_result, current_state}` |
| 結果照会 | `GET /api/liff/:clinicId/revisit-intake/operations/:key` | §5.4 固定形（redacted 成功形を含む） |

選択肢 code は §4 の allowlist を zod enum で固定。未知 code は送らず、既定値へ補完しない。

### 2.5 変更候補ファイル

| 区分 | ファイル |
|:--|:--|
| 新規 | `frontend/line-reserve/src/pages/RevisitIntakePage.tsx`（+ `.test.tsx`） |
| 新規 | `frontend/line-reserve/src/components/revisit-intake/PetAnswerForm.tsx`、`AnswerHistory.tsx`、`SubmitResultBanner.tsx` |
| 新規 | `frontend/line-reserve/src/hooks/use-revisit-intake-drafts.ts`、`use-revisit-intake-submit.ts`（+ test） |
| 新規 | `frontend/line-reserve/src/lib/revisit-intake-options.ts`（code→ラベル v1 固定） |
| 変更 | `src/api/liff-api.ts`（3 メソッド）、`src/api/schemas.ts`、`src/types/models.ts`（`PageType`・DTO 型）、`src/app-page-switch.tsx`、`src/app-flow-handlers.ts`、`src/pages/CompletePage.tsx`、`src/pages/MyReservationsPage.tsx` |
| fixture | MSW handler（成功 / 409 両種 / 応答喪失→照会 4 状態 / 403）。合成値のみ |

### 2.6 依存関係

- 契約: EMR-471（route・DTO・error）、EMR-472（`content_available` と認可の意味）、EMR-473（操作 key・照会 4 状態）。実装完了は不要。
- 既存再利用: `useFetchState` / `resolveFetchError` / `ErrorPage` / `Spinner` / `sanitize.ts`。
- 非対象: 予約・受付・通知・カルテ作成の発火、本文・token の永続化。

### 2.7 PR 単位

| PR | 内容 | 閉じ方 |
|:--|:--|:--|
| 474-A 契約層 | zod schema・adapter 3 メソッド・MSW fixture・`PageType` 追加 | 契約 fixture 同梱。未実行試験は NOT_RUN 記録 |
| 474-B 画面 | `RevisitIntakePage` + 頭別 draft + 入口 CTA | AC-optional / multi-pet / revisions / new-target の画面状態を test に記述（実行は本人指示時） |
| 474-C 再送・照会 | 操作 key・結果照会・競合 UI | 応答喪失→4 状態、409 両種、二重送信抑止を test に記述 |

A と B は 1 PR 統合可。統合担当（本票の FE 担当一人）が 471 へ結合条件と証跡を渡す。

---

## 3. EMR-475 スタッフ予約詳細の再診原文・旧 target 履歴

### 3.1 画面状態（`ReservationDetailModal` 内に新セクション「飼主事前問診（LINE）」）

| 状態 | 表示 | 根拠 |
|:--|:--|:--|
| 読込中 | セクション内スケルトン | patterns §6 |
| 未回答 | 「未回答」バッジ（文字 + アイコン）。空状態 3 点 | FR「予約詳細に『問診未回答／回答済み』…を表示」 |
| 回答済み | 最新原文 7 項目 read-only、回答日時・最終訂正日時、出所「飼主（LINE）」。訂正履歴は折りたたみ（全 revision） | AC-provenance、FR「スタッフは原文・訂正履歴を閲覧できるが原文編集不可」 |
| 旧 target あり | 「以前の対象ペットの回答」として別グループ。ペット名 + 世代で区別 | §6 予約詳細「最新 target と旧 target は別表示、履歴を最新回答へ copy しない」 |
| 取消済み / soft delete 後 | 専用 history reader の範囲を同形式で表示 + 「取消済み予約」注記 | §6「cancelled/deleted の原文履歴は認可した専用 history query」 |
| 閲覧権限なし（403） | 「閲覧権限がありません」1 行（本文・件数を漏らさない） | §5.4 staff grant 各行検査 |
| 取得失敗 | セクション内 `再試行`。他の予約情報は表示継続 | — |
| 他院予約 | 既存 `clinicId` 伝搬で同 clinic 取得 | 現行 `useGetReservation` |

スタッフ補足は既存の予約 notes（「予約内容」）/ Inquiry へ。原文欄に編集・コピー・補足入力 UI を置かない（§4「スタッフ補足は既存 Inquiry へ」、technical-spec「スタッフ補足はスタッフ問診 owner へ」）。

### 3.2 操作手順

1. 予約一覧 → モーダル。`useGetReservation` と並行して `useGetReservationRevisitIntake(id, clinicId)`。
2. 履歴折りたたみ・旧 target グループ展開（キーボード到達可）。
3. モーダル閉鎖・別予約切替で query を破棄（key に `id|clinicId`）。

### 3.3 API 対応（候補。正本は EMR-471 / 472）

| 用途 | 候補 | DTO 要点 |
|:--|:--|:--|
| 取得 | `GET /v1/reservations/:id/revisit-intake`（既存 `reservations:view`） | `current_target{pet, generation, head_revision, latest_answer, revisions[]}`, `previous_targets[]`, `appointment_state`, `visibility` |

### 3.4 変更候補ファイル

| 区分 | ファイル |
|:--|:--|
| 新規 | `frontend/src/features/reservations/components/RevisitIntakeSection.tsx`（+ `.test.tsx`） |
| 新規 | `features/reservations/api/get-revisit-intake.ts`、`lib/transforms/revisit-intake.ts` |
| 変更 | `components/ReservationDetailModal.tsx`（`InfoRow` 群の後に挿入）、`ReservationDetailModal.test.tsx` |
| 整形 | `lib/format` / `lib/jst-date`（日時・空値 `—`） |

### 3.5 依存・PR 単位

- 契約: EMR-471（原文 DTO）、EMR-472（staff grant・行可視性）。
- FE PR 1 本（セクション + adapter + MSW fixture）。BE reader PR は別 owner。統合担当が両 PR の採用 SHA と実 API 結合を回収。

---

## 4. EMR-477 治療プランの予約回答原文 reader

### 4.1 配置と画面状態

配置: 「診察/治療プラン」タブ内、`DiagnosisHeader` 直下・`ClinicalPlanSection` の上に独立カード「飼主事前問診（LINE・読み取り専用）」（lifecycle FR-medical-records-integrated-plan「別データ・別区画として表示」、477 監査コメント「独立 read-only 欄」）。

| 状態 | 表示 | 根拠 |
|:--|:--|:--|
| 新規カルテ（`isNewRecord`） | カード非表示 | binding は record 作成/受付の appointment lock 下で確保（§6 D03）。未保存 record に binding は無い |
| binding あり・回答あり | 7 項目 read-only、回答版、受信時刻、出所「飼主（LINE）」。`readonly` 表現 | AC-provenance、AC-owner-intake-protected |
| binding あり・未回答 | 「未回答」バッジ + 空状態 | §7.3「予約現在 target に answer がない場合だけ未回答」 |
| binding なし（NULL） | 「対応する予約回答を特定できません」。未回答と文言を分ける | §7.3「NULL は『対応不明』であり『問診未回答』とは別表示」 |
| record の患者/予約と target 不一致 | 「出所不明」。本文を出さない | §6 D03「不一致なら原文を表示せず出所不明」、technical-spec「原文取得不明と表示」 |
| 失権（403） | カード非表示 + 1 行注記。再読込でも書換えなし | 477 受入条件 |
| 取得失敗 | カード内 `再試行` | — |
| 患者切替 | query key `recordId|recordClinicId|petId` で破棄 | 477 監査コメント 3) |

### 4.2 書込経路からの保護

- 部品は `onChange` 系 props を持たず、`use-medical-record-diagnosis-state.ts` に接続しない。
- `applyCopySource` / `onCopyRecordTreatments` / 将来の「入力内容クリア」「複製取消」は対象 state を持たないため影響しない（lifecycle「クリア／過去カルテ複写は飼主原文・旧回答を消去・転用しない」、§6 D03「copy/clear/Undo は binding/answer を payload に含めない」）。test に「複製実行後も原文 DTO・表示が不変」を記述。
- 診断コピー adapter の複数行移行（EMR-495）には触れない。

### 4.3 API 対応（候補。正本は EMR-476 / 472）

| 用途 | 候補 | DTO 要点 |
|:--|:--|:--|
| 取得 | `GET /v1/medical-records/:id/revisit-intake`（現在 grant） | `binding: "bound"|"unbound"|"mismatch"`, `answer|null{revision, received_at, provenance, fields…}` |

### 4.4 変更候補ファイル

| 区分 | ファイル |
|:--|:--|
| 新規 | `frontend/src/features/medical-records/components/OwnerIntakeReadOnlyCard.tsx`（+ `.test.tsx`） |
| 新規 | `features/medical-records/api/revisit-intake.ts`（`useGetRecordRevisitIntake`） |
| 変更 | `components/MedicalRecordDiagnosisPlan.tsx`（カード挿入）、必要なら `lib/medical-record-tabs-types.ts` |
| 試験記述 | `MedicalRecordClinicalTabs.test.tsx` に「複製・患者切替・失権で原文不変」ケース |

### 4.5 依存・PR 単位

- 契約: EMR-476（binding 固定の意味）、EMR-472（read 認可）。489 全 owner 完了は条件にしない（477 監査コメント）。
- FE PR 1 本。478 へ reader 境界の証跡（test 名・fixture）を渡す。

---

## 5. EMR-487 接種証明書 3 様式の preview・発行・再印刷

### 5.1 画面構成

| 画面 | 役割 | 根拠 |
|:--|:--|:--|
| 入口 | `VaccinationForm`（`/vaccinations/:id`）と `MedicalRecordVaccination` の接種行に「証明書」ボタン | technical-spec CL-R11/12「両入口」（カルテ内 hook と独立 form）に揃える |
| 新規 `/vaccinations/:id/certificate` | `PageLayout` shell。様式選択（通常=縦/横、狂犬病は別 series）、preview、「発行」「再印刷」「履歴」 | lifecycle FR-certificate-editions、第 3 項 A |
| 印刷 | 専用 renderer（`@media print` で shell を隠す）。汎用 `MedicalRecordPrintView` / `window.print` 流用は不可 | lifecycle「旧未保存印刷・印刷時自動登録・全印刷プレビュー省略は復活させない」、487 本文 |

新規リーフルートは同一コミットで `docs/spec/ui-design-compliance.md §2` を更新（ui-design-guide ハードルール）。

### 5.2 画面状態

| 状態 | 表示・操作 | 根拠 |
|:--|:--|:--|
| 未発行 | 現在接種内容からの preview（「未発行」バッジ）。操作は「発行」のみ。preview 省略不可 | AC-no-side-effect |
| 最新版あり | 最新 edition snapshot の preview。「再印刷」。縦/横切替は同一 edition を再描画し、発行 operation・edition 増分を起こさない | 第 3 項 A、technical-spec「layout 切替だけで発行 operation や edition 増分を起動しない」 |
| 訂正済み旧版 | 履歴 read-only。「有効な証明書ではありません」。再印刷・artifact 取得ボタン無し | AC-correction、Q-V2 |
| 接種訂正後で edition が旧 | 「接種内容が更新されています。新しい版を発行してください」。旧版の有効出力は拒否 | technical-spec「訂正保存成功時点で旧内容の有効出力停止」 |
| 発行中 / 再印刷中 | `aria-busy`。操作 key 保持 | — |
| render / upload 失敗 | 同一 key で再試行。旧版は復活しない | technical-spec 手順 2、DC-V03 |
| 結果不明 | 操作照会 → 成功なら artifact、不明なら再試行案内 | DC-V07 |
| 409（同時訂正） | 再読込して最新版へ | DC-V04 |
| 403 | 権限なし表示 | — |
| 発行完了 | receipt 表示。「印刷完了」とは表示しない | technical-spec「印刷 dialog を開いたことと物理印刷完了を区別」 |

### 5.3 操作手順

1. 接種行 → 「証明書」→ `/vaccinations/:id/certificate`。
2. `GET editions` で series / 最新 edition / 履歴 / 有効性を取得。
3. 様式選択 → preview。
4. 「発行」: `POST operations`（`kind=issue`、UUID key）→ edition 確定 → artifact 取得 → 印刷ダイアログ。
5. 「再印刷」: `POST operations`（`kind=reprint`、UUID key、`edition_id`、`layout`）→ artifact 取得（別 layout artifact 未生成なら同 snapshot・固定描画版から生成、technical-spec）→ 印刷ダイアログ。
6. 履歴から旧版 → read-only 表示のみ。

### 5.4 API 対応（候補。正本は EMR-486 が固定）

| 用途 | 候補 | DTO 要点 |
|:--|:--|:--|
| 取得 | `GET /v1/vaccinations/:id/certificate-editions` | `series[]{kind, current_edition{id, edition, snapshot, source_vaccination_version, issued_at}, history[]}`, `effective` |
| 発行/再印刷 | `POST /v1/vaccinations/:id/certificate-operations` + UUID header | `{kind, layout, edition_id?}` → `{operation_result, artifact_ref}` |
| artifact | `GET /v1/certificate-artifacts/:id`（現在 grant・有効版検査） | 不変 blob 参照。無認可の恒久 URL にしない |
| 照会 | `GET /v1/certificate-operations/:key` | 4 状態 |

### 5.5 renderer の所在（技術整理、業務判断ではない）

- lifecycle・technical-spec は「描画定義／帳票 adapter（template 版・font/asset・renderer 版・golden）」と「出力 artifact（immutable blob + digest）」を分け、所在を本人承認事項にしていない（technical-spec「内部 ID・生成順・検証 DC を本人の直接承認としない」）。
- 本票は 3 様式の template / field projection / golden（原見本との項目一対一対応）を FE 側で所有する。artifact の不変保存・digest・有効性 gate は 486 backend の責務。server 側が同じ projection から blob を生成する場合も、487-A の snapshot→表示値 model と golden fixture を再利用する。具体の分担は 486 / 487 owner 間の技術調整であり、Dots へ返さない。

### 5.6 変更候補ファイル

| 区分 | ファイル |
|:--|:--|
| 新規 | `frontend/src/features/vaccinations/certificate/CertificatePage.tsx`、`CertificatePreview.tsx`、`layouts/PortraitCertificate.tsx`、`layouts/LandscapeCertificate.tsx`、`layouts/RabiesCertificate.tsx`、`certificate-model.ts`（snapshot→表示値。縦 LOT1〜4 / 横 E-Mail / 狂犬病の独立発行日欄なしを保持） |
| 新規 | `certificate/__fixtures__/`（合成値のみ）、`*.test.tsx`（項目一対一表）、`*.stories.tsx` |
| 新規 | `features/vaccinations/api/certificate-editions.ts`、`certificate-operations.ts` |
| 変更 | `config/paths.ts`、`app/routes/*`、`routes/VaccinationForm.tsx`、`medical-records/components/MedicalRecordVaccination.tsx`、`docs/spec/ui-design-compliance.md §2` |
| 整形 | 日付・金額は `lib/format`。番号・金額・医学値は snapshot の値のみ（推測・自動計算なし） |

### 5.7 依存・PR 単位

| PR | 内容 | 閉じ方 |
|:--|:--|:--|
| 487-A renderer + golden | 3 様式 layout 部品・model・fixture・項目対応表・stories | 原見本との項目一対一表を PR 本文に添付。486 契約は mock で可 |
| 487-B preview/発行/再印刷 | ページ・route・adapter・操作 key・結果照会・MSW | 拒否状態（失権・旧版・同時訂正）を test に記述 |
| 487-C 履歴・訂正連動 | 旧版 read-only、接種訂正後の案内 | AC-correction / basic-update |

依存: EMR-486 の edition 契約。実印刷・法令適合は未実施のまま分離。

---

## 6. EMR-463 飼主一覧の認可内医院初期値と選択保持

### 6.1 画面状態

| 状態 | 表示・挙動 | 根拠 |
|:--|:--|:--|
| 単一医院 | 変更なし（`ClinicScopeFilter` 非表示） | 現行 |
| 複数医院・初回（保存済み選択なし・URL `clinics` なし） | 認可内全医院を選択し `clinic_ids` を明示送信 | FR-common-ui-list-filter、AC-list-clinic-memory |
| 保存済み選択あり | 復元。失権医院は除外。除外後 0 件なら初回扱い（認可内全医院） | 同上 |
| 認可医院 0 件 | 空 scope・利用不可で止める。認可外や他 account の保存値へ fallback しない | 451 コメント 2026-10-20 |
| 明示選択 | URL `?clinics=` 更新（既存）+ 保存先へ書く。`page` リセット（既存） | 現行 |
| 一覧→詳細→戻る | URL 履歴で検索語・医院・ページを保持。メニュー再入場と区別 | AC-list-entry-return、ui-contracts「状態保持・戻る」 |
| 遅延応答 | loader の順序保証に依存。追加 fetch を足す場合は request token で古い応答を破棄 | 463 受入条件 |
| 403（直接 API で認可外） | エラー表示 + 選択を認可内へ補正 | 463 受入条件 |

### 6.2 保存媒体（技術選定）

ui-contracts「保存先が未選定なら実装工程で方式を選ぶ」、451 コメント「保存媒体は技術設計として用途別 URL/保持を選定」に従い FE で選定する。採用案: `localStorage` キー `owners-list.clinics.<staffId>`（医院 ID の配列のみ。患者情報・token を含まない）。読出し時に認可内医院と交差し、他 account の値を使わない。予約・会計の一覧には広げない（451「既存挙動維持」）。

### 6.3 API 対応

既存 `GET /v1/pets?clinic_ids=…`（未割当 ID は 403）。`api.yaml` の `/pets` へ `clinic_ids` 省略時の意味と 403 契約を追記するのは統合担当の整合作業。新 route・DB 差分なし（`httpapi/context.go` `ResolveListClinicIDs`、`clinic_permission.go` `FilterClinicIDsForPermission`）。

### 6.4 変更候補ファイル

| 区分 | ファイル |
|:--|:--|
| 新規 | `frontend/src/features/owners/hooks/use-owners-clinic-scope.ts`（`useClinicScope` を包む adapter、+ test） |
| 新規 | `features/owners/lib/owners-clinic-selection-storage.ts`（保存・復元・失権除外の純関数、+ test） |
| 変更 | `features/owners/routes/OwnersList.tsx`、`features/owners/loaders.ts`（`clinics` 未指定時に保存済み選択 ∩ 認可内 or 全認可を `clinic_ids` へ） |
| 不変 | `hooks/use-clinic-scope.ts` 本体。必要なら opt-in 引数 `defaultClinicIds?: () => number[]` のみ |
| 試験記述 | `OwnersList.test.tsx` / `OwnersList.filter-wiring.test.tsx` に初回全医院・保持・失権除外・0 件・復帰 |

### 6.5 依存・PR 単位

- 契約: EMR-451 / 453 の該当 subset は「既存一覧 API の `clinic_ids` 契約」のみ。491 全 writer は待たない。
- FE PR 1 本。統合担当が実 reader 接続・直接 API 拒否の結合受入を回収。

---

## 7. 共通の閉じ方・DoD

- 各 PR の閉じ方は「実装 + 契約 fixture（MSW、合成値のみ）+ 画面状態を記述した test」。試験実行・隔離環境は必須条件にしない。実行していない検査は NOT_RUN / UNKNOWN として PR 本文に明記し、PASS と書かない。
- まとめ検証は本人の実施指示があった時だけ main のローカル環境で行う（476 / 478 本文）。
- UI デザインガイドのハードルール（PageLayout shell、トークン色、readonly≠disabled、空状態 3 点、保存ボタンを未入力理由で disabled にしない、`lib/format` 整形、新規ルートの compliance §2 同時更新）を適用。
- 生成型（`src/types/generated/`）は `make codegen` がユーザー実行のため、FE モック PR では feature 内の手書き型 + zod で先行し、`api.yaml` 採用 SHA 固定後の生成型切替は統合担当の整合作業とする（技術運用、業務判断ではない）。

---

## 8. 再点検で自己解決した項目（旧 Q1〜Q12）

| 旧 | 結論 | 根拠 |
|:--|:--|:--|
| Q1 `api.yaml` 契約の owner | 再診回答は EMR-471 が最初に追記。証明書は EMR-486 が 487 へ契約を引渡す。FE は draft を書かない | 471 コメント 06:09、486 コメント 03:17 |
| Q2 LIFF 予約詳細の不在 | マイ予約カードを予約詳細相当とし、行内 CTA + 専用回答画面 | screens/37 画面構成（予約詳細画面なし、マイ予約=予約確認） |
| Q3 頭別 target の取得 | GET revisit-intake が pet 別 target を返す。一覧 DTO は不変 | reservations technical-spec「LINE read/submit adapter」 |
| Q4 draft 保持 | メモリのみ。再読込で失われるのは既決の帰結 | 474 受入条件、screens/37「URL/ストレージへ新たに永続化しない」 |
| Q5 binding NULL の文言・配置 | 「対応不明」を未回答と分ける。治療プラン統合画面の別区画、新規カルテ非表示 | revisit-intake §7.3 / §6 D03、lifecycle FR-integrated-plan |
| Q6 staff の履歴可視範囲 | 既存 staff view grant・clinic・target.pet の利用 scope で各行検査、全 revision。取消済みは専用 history query | revisit-intake §5.4 / §6 予約詳細 |
| Q7 renderer 所在 | 技術分担（§5.5）。本人承認事項ではない | vaccinations technical-spec「内部方式を本人の直接承認としない」 |
| Q8 縦/横切替 | 同一 edition 再描画、operation なし。再印刷のみ receipt | technical-spec「layout 切替だけで発行 operation や edition 増分を起動しない」 |
| Q9 463 保存媒体 | 技術選定（§6.2） | ui-contracts「保存先が未選定なら実装工程で方式を選ぶ」、451 コメント |
| Q10 463 対象 path | FE は `GET /pets` に依存。票本文の `owner/http_routes.go` は読取 route の所在として不正確（下記保守事項） | `features/owners/loaders.ts`、`pet/routes.go` |
| Q11 生成型 | 手書き型 + zod で先行、切替は統合担当 | §7 |
| Q12 スタッフ補足の保存先 | 既存 notes / Inquiry。新規欄なし | FR「スタッフ側の記録へ補足」、revisit-intake §4 |

### Plane 票記載の事実訂正（保守事項。業務判断ではない）

- EMR-463 の対象 path `backend/internal/owner/http_routes.go` は、飼主一覧の実データ取得経路（`GET /v1/pets`、`backend/internal/pet/routes.go`）と一致しない。契約整合の対象は `/pets` を含める。

## 9. ChatGPT Dots への確認事項（未決の業務判断のみ）

| # | 票 | 未決事項 | FE 側の暫定前提 |
|:--|:--|:--|:--|
| Q-A | EMR-487 / 486 | 狂犬病予防注射済証の対象接種の識別基準。ワクチンマスタ（`vaccines`）に証明書種別属性が無く、DB の `certificate_kind`（normal / rabies）を何から決めるか（マスタへ種別属性を追加するか、発行時にスタッフが様式を選ぶか、名称一致は不可）。正本は「狂犬病を同一系譜へ勝手に統合しない」とだけ定める | 発行画面でスタッフが「通常（縦/横）」「狂犬病」を明示選択し、FE は名称一致で自動判定しない |
| Q-B | EMR-487 / 486 | 狂犬病様式の証明番号・注射済票番号・犬鑑札番号・各費用（登録手数料／注射料／注射済票費用／合計）の入力元と必須性。正本は「例示値の既定採用・自動計算・新たな必須化をしない」、technical-spec は「番号が必要なら既存の印字仕様と利用院の確認を行い、内部 UUID を法的証明番号として印字しない」とし未選定 | snapshot に値があれば印字、無ければ空欄。FE で採番・計算・必須化しない |
