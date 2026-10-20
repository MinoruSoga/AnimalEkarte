# EMR-444 配下フロント寄り票の着手プラン（2026-10-20・第 3 版）

> 位置づけ: 判断根拠の保存。実行 SoT は Plane。未決の業務判断だけを末尾「ChatGPT Dots への確認事項」として返す。
> 正本: 最新 main `docs/spec/domains/`（観測 SHA 05da02f30、2026-10-20。PR #746 = EMR-463 契約、PR #748 = EMR-471/472/473/476 再診問診契約をマージ済み）。技術設計は `docs/design/`、共有契約は `docs/spec/shared/`、wire 正本は `backend/docs/api.yaml`。
> 完了条件: 本人の最新指示（EMR-476 / 478 / 518 本文 2026-10-20 節）により、試験実行・Docker・隔離環境は着手・PR 作成の必須条件にしない。各 PR は実装・契約 fixture・画面状態を記述した test で閉じ、未実行は NOT_RUN と記録。まとめ検証は本人指示時に main のローカル環境で行う。

## 0. 現状スナップショット（2026-10-20 14:30 UTC）

| 票 | 件名（要約） | Plane 状態 | 契約 | FE 実装 |
|:--|:--|:--|:--|:--|
| EMR-463 | 飼主一覧の認可内医院初期値と選択保持 | Review（PR #746 マージの自動遷移。FE/BE 実装は未着手扱い） | `GET /me` `GetMeResponse.owners_view_clinic_ids`（selected-not-implemented） | `feat/emr-463-owners-clinic-scope-fe` で着手 |
| EMR-474 | LIFF の再診回答と訂正・再送 UI | Ready | LIFF 3 route（selected、runtime 未実装） | `feat/emr-474-liff-revisit-intake-fe` で着手 |
| EMR-475 | スタッフ予約詳細へ再診原文・旧 target 履歴 | Ready | `GET /reservations/{appointmentId}/revisit-intake` | `feat/emr-475-reservation-revisit-intake-fe` で着手 |
| EMR-477 | 治療プランへ予約回答原文 reader（read-only） | Ready | `GET /medical-records/{recordId}/revisit-intake` | `feat/emr-477-record-revisit-intake-fe` で着手 |
| EMR-487 | 接種証明書 3 様式の発行・再印刷 | Ready | EMR-486 の edition 契約が未固定 | 実装待機（§5） |
| EMR-518 | 診療項目パッケージ：異種項目の定義・まとめ追加 | Running（Dots 設計中） | 未固定 | 影響整理のみ（§10） |

### 0.1 共通の着手方針（確定済み）

| 方針 | 根拠 |
|:--|:--|
| 契約が `api.yaml` へ固定（マージ）された票は、backend 未実装でも FE 専有の契約モック PR として完結させる。実 API 結合は別報告 | 471 コメント 06:09、reservations technical-spec `revisit-intake-api-selected-v3`「FE は採用 SHA 固定後に契約 mock で進め」 |
| FE 担当は `api.yaml`・生成型・全体認証・共通医院 scope hook・共通 MSW fixture を変更しない。必要な共通変更は取りまとめ側が集約して Dots へ相談 | 本人指示 2026-10-20 |
| 1 票 1 ブランチ 1 PR。計画文書ブランチ（`claude/sweet-pasteur-9bek2i`）へ実装を混在させない | 同上 |
| 試験実行・Docker・隔離環境は不要。test・fixture・画面状態の記述は成果物に含め NOT_RUN と明記 | 同上、476/478 本文 |
| マージ判断は Dots 側 | 同上 |

### 0.2 コードベースの事実（main 05da02f30）

| 事実 | 影響 |
|:--|:--|
| `api.yaml` に再診問診 5 操作（staff 2 GET、LIFF GET/POST/operations GET）と `GetMeResponse` が収録。いずれも `x-contract-status: selected`、runtime 未実装 | FE は契約 mock で先行可。backend route・handler は無い |
| 再診問診は **1 appointment = 1 pet、`current_target` は単数**。複数頭予約は予約ごとに別 appointment | 旧計画 §2 の「1 画面で頭別セグメント」は撤回。LIFF は予約カードごとに回答画面へ |
| LIFF（`frontend/line-reserve/`）に予約詳細画面は無い（screens/37 画面構成）。永続化無し、router 無し | 入口はマイ予約カードの行内 CTA + 完了画面 CTA |
| `use-clinic-scope.ts` は URL のみで未指定時は現在医院 1 件。`ownersLoader` は URL の `clinics` をそのまま `/pets` へ送る | 463 は owners 専有 adapter で閉じる（§6） |
| 治療プランに package/set/bundle の概念・一括追加 API は無い。`TreatmentSearchDialog` の候補は 診察/処置/予防(vaccines)/検査(= checkup_types)/薬剤 で、exam_types・物販は未提供。予防/検査は `item_type: other` として master ID を失う | 518 の影響範囲（§10） |
| FE に証明書関連コードは無い。印刷基盤は汎用 `MedicalRecordPrintView` + `window.print()` | 487 は専用 renderer 新設（待機中） |

## 1. 着手順序と担当分割

| 順 | 票 | ブランチ | 担当モデル | 状態 |
|:--|:--|:--|:--|:--|
| 1 | EMR-463 | `feat/emr-463-owners-clinic-scope-fe` | Opus | 実装中 |
| 2 | EMR-477 | `feat/emr-477-record-revisit-intake-fe` | Sonnet | 実装中 |
| 3 | EMR-475 | `feat/emr-475-reservation-revisit-intake-fe` | Sonnet | 実装中 |
| 4 | EMR-474 | `feat/emr-474-liff-revisit-intake-fe` | Opus | 実装中 |
| 5 | EMR-487 | 未作成 | — | 486 契約待ち |

各 worktree は `../AnimalEkarte-emr-<n>`。取りまとめ（Fable）がレビュー後に PR を作成する。

---

## 2. EMR-474 LIFF 再診回答・訂正・再送 UI（契約 v3 準拠）

### 2.1 画面構成

| 画面 | 役割 | 根拠 |
|:--|:--|:--|
| `CompletePage`（step8） | 「事前問診に回答する（任意）」CTA | FR「予約完了画面と予約詳細から回答へ進み」 |
| `MyReservationsPage` | `confirmed` カードに「問診を回答する／訂正する」。取消済みは履歴があるときだけ「回答履歴を見る」 | LIFF の予約詳細相当はマイ予約カード（screens/37） |
| 新規 `RevisitIntakePage`（`PageType: revisit-intake`） | 1 予約 = 1 ペットの form。履歴（current の旧版 / previous target）は別セクション read-only | technical-spec v3「current_target 単数」 |

### 2.2 画面状態（`RevisitIntakeLIFFRead` / error code に対応）

| 状態 | 表示・操作 |
|:--|:--|
| 読込中 | `Spinner` |
| `availability=target_unresolved` | 「対象を特定できません」 |
| `answer_state=unanswered`・`editable=true` | 空 form。「回答は任意」注記。送信ボタン常時活性。`expected_revision=null` |
| `answered`・`content_available=true`・`editable=true` | 最新回答 prefill。「訂正して再送信」。回答日時・最終訂正日時。期待版 = `head_revision` |
| `answered`・`content_available=false` | 「回答済み」のみ。本文非表示・prefill なし。新送信は空 form、期待版 = `head_revision` |
| `editable=false` | `edit_block_reason` 別文言（reception_started→「受付後の変更はスタッフへお伝えください」、cancelled / deleted / authorization_denied / target_unresolved / unsupported_lifecycle） |
| 履歴 | `target_relation=current` は訂正履歴、`previous` は「以前の対象ペットの回答」。入力へコピーしない。`next_cursor` で「さらに表示」 |
| 送信中 | `aria-busy`・二重送信抑止・入力保持 |
| 成功 | 成功表示。`liff.sendMessages` は呼ばない |
| 409 `REVISION_CONFLICT` | 入力保持・再取得案内。再送は新 key |
| 409 `FORM_VERSION_MISMATCH` / `TARGET_CHANGED` | 再取得して form を作り直す |
| 409 `INTAKE_CLOSED` | 閲覧へ |
| 409 `KEY_CONTENT_MISMATCH` | 新 key |
| 400 `INVALID_INPUT` | 欄直下エラー・先頭エラー欄へフォーカス |
| 応答喪失 | 同 key で `operations/{key}` 照会。`succeeded` → 成功、`claimed_uncommitted` / 404 `not_found` → 同 key 再送可、`result_unavailable` → 「結果を確認できません」（新規送信しない） |
| 401 / 403 / 404 | 既存 `resolveFetchError` / `ErrorPage`。receipt の有無を推測しない |
| 再読込・LIFF 終了 | 未送信入力は失われる（memory のみ）。dirty のときだけアプリ内離脱確認、OS 終了保証なし | 

### 2.3 API 対応（`api.yaml`、採用 SHA 05da02f30）

| 用途 | path | 要点 |
|:--|:--|:--|
| 取得 | `GET /api/liff/{clinicId}/my-reservations/{appointmentId}/revisit-intake` | `RevisitIntakeLIFFRead`。`cursor`/`limit`。no-store |
| 送信 | `POST …/revisit-intake/submissions` + `Idempotency-Key` | `RevisitIntakeSubmitRequest`{pet_id, target_generation, expected_revision, form_definition_version, answers} → `RevisitIntakeResultResponse` |
| 照会 | `GET /api/liff/{clinicId}/revisit-intake/operations/{operationKey}` | 200 成功/未解決、404 `RevisitIntakeLookupNotFound` |

### 2.4 変更候補ファイル・依存・PR

- 新規: `pages/RevisitIntakePage.tsx`、`components/revisit-intake/*`、`hooks/use-revisit-intake-drafts.ts`、`hooks/use-revisit-intake-submit.ts`、`lib/revisit-intake-options.ts`
- 変更: `api/liff-api.ts`、`api/schemas.ts`、`types/models.ts`、`app-page-switch.tsx`、`app-flow-handlers.ts`、`pages/CompletePage.tsx`、`pages/MyReservationsPage.tsx`
- 未結合: EMR-471/472/473 runtime、LiffAuth の read-only principal。1 PR。

---

## 3. EMR-475 スタッフ予約詳細の再診原文・旧 target 履歴

- 契約: `GET /reservations/{appointmentId}/revisit-intake` → `RevisitIntakeStaffReservationRead`（`current_target: RevisitIntakeReadOnlyCurrentTarget`、`history[]`、`next_cursor`）。staff は既存 `reservations:view` + clinic + 各 target.pet 行認可、LINE author 一致は強制しない。
- 画面: `ReservationDetailModal` 内「飼主事前問診（LINE）」セクション。状態は 読込中 / target_unresolved / 未回答 / 回答済み / `content_available=false` / current の訂正履歴 / previous target 別グループ / 403（1 行、本文・件数を漏らさない）/ 404 / 取得失敗（再試行）。原文欄に編集・コピー・補足 UI を置かない。スタッフ補足は既存 notes / Inquiry。
- ファイル: `features/reservations/api/get-revisit-intake.ts`、`lib/revisit-intake-options.ts`、`components/RevisitIntakeSection.tsx`、`ReservationDetailModal.tsx` 挿入。1 PR。未結合: EMR-471/472 staff reader。

---

## 4. EMR-477 治療プランの予約回答原文 reader

- 契約: `GET /medical-records/{recordId}/revisit-intake` → `RevisitIntakeRecordRead`（`binding: unbound | mismatch | bound`）。unbound は未特定で未回答ではない。固定 target 内の履歴のみ。GET/copy/clear/Undo から binding を変えない。
- 画面: 「診察/治療プラン」タブ、`DiagnosisHeader` 直下・`ClinicalPlanSection` の上に独立 read-only カード。状態は 新規カルテ非表示 / bound 回答あり / bound 未回答 / `content_available=false` / unbound（「対応する予約回答を特定できません」）/ mismatch（「出所不明」）/ 403 / 404 / 取得失敗 / 履歴。
- 保護: 部品は `onChange` 系を持たず clinical_plan state に接続しない。EMR-495（コピー adapter・クリア・Undo）の領域は触らず、必要なら相談。
- ファイル: `features/medical-records/api/revisit-intake.ts`、`lib/revisit-intake-options.ts`、`components/OwnerIntakeReadOnlyCard.tsx`、`MedicalRecordDiagnosisPlan.tsx` 挿入。1 PR。未結合: EMR-476 capture/reader、EMR-472 認可。

---

## 5. EMR-487 接種証明書（実装待機）

着手条件: EMR-486 の edition reader / 発行・再印刷 operation の契約が `api.yaml` へ固定されること。それまで renderer・golden（487-A）の準備だけ先行可。

### 5.1 確定した方針（2026-10-20 本人回答、EMR-518 本文に記録）

- 証明書の料金欄（狂犬病様式の登録手数料／注射料／注射済票費用／合計）は、**発行前にスタッフが手動修正できる**。
- 証明書上の料金編集を**会計明細へ逆反映しない**。証明書出力から課金行を作らない。
- 登録料・済票料の構成・運用は医院ごとに異なる（EMR-518 の汎用パッケージ構成に従う）。全国一律の請求イベントを新設しない。名称一致で接種種別や料金 role を推定しない。

### 5.2 画面への反映

| 状態 | 追加・変更 |
|:--|:--|
| 未発行 preview | 料金欄が編集可能（`MoneyInput`）。既定値は 486 契約の snapshot 提案値、無ければ空欄。自動計算・採番をしない |
| 発行 | 編集後の値を snapshot に固定して `issue` operation へ渡す。以後は不変 |
| 最新版 preview / 再印刷 | 料金欄は read-only（保存済み snapshot）。訂正は接種訂正 → 新版発行の既決経路 |
| 会計 | 証明書の料金編集は会計(医師確認)・未請求候補・請求確定へ一切影響しない |

### 5.3 その他（第 2 版から変更なし）

3 様式・縦横同一 edition・旧版履歴 read-only・operation key / 結果照会・renderer 所在の技術分担は第 2 版 §5 のとおり。狂犬病様式の対象判定は §9 Q-A。

---

## 6. EMR-463 飼主一覧の医院選択（採用契約 PR #746 準拠）

### 6.1 契約（`docs/design/domains/owners-pets/technical-spec.md#owners-list-scope-contract-emr463`、`api.yaml GetMeResponse`）

- BE: `GET /me` の 200 にだけ `owners_view_clinic_ids`（正規形 10 進文字列、重複なし、許可ゼロは `[]`、評価 error は 5xx）。runtime 未実装。
- FE 処理順: ① owners 専有 adapter が `ownersLoader` の各実行で `GET /me` を取得し必須 parse（欠落/NULL/不正は取得エラー、`[]` へ補完しない、既存 `ME_QUERY_KEY` キャッシュを流用しない）→ ② URL 適用済み選択 ∩ scope → 保存選択 ∩ scope → 初回は全 scope。scope `[]` なら一覧 API を呼ばない → ③ 非空選択を文字列のまま `clinic_ids` へ → ④ `/pets` 403 は scope を再取得して再照合、失敗 scope で再送しない。通信障害は失権ではなく保存値を消さない → ⑤ `clinic_ids` 最大 50、超過は停止（切捨て・分割取得しない）。
- 保存: `localStorage` key `owners-list.clinics.v1.<user.id>`、内容は医院 ID 配列のみ。別ユーザー key を読まない。
- login/logout・user/clinic 切替・`ME_QUERY_KEY` 更新で進行中の scope/list 応答を無効化。

### 6.2 画面状態

| 状態 | 挙動 |
|:--|:--|
| 単一医院 | 変更なし（フィルタ非表示） |
| 初回（URL・保存なし） | 全許可医院を選択し `clinic_ids` 明示 |
| 保存あり | 保存 ∩ 許可。全失効なら全許可 |
| 許可ゼロ | 空 scope・利用不可で停止（fallback なし。451 コメント） |
| URL に許可外 ID | 交差外は送らない。空選択を query 省略へ変換しない |
| 51 件以上 | 超過として停止 |
| `/pets` 403 | scope 再取得 → 再照合 |
| scope 取得失敗 / 5xx | 一覧エラー。許可ゼロへ縮退せず保存値を消さない |
| 一覧→詳細→戻る | URL 履歴で検索語・医院・ページ保持（既存） |

### 6.3 ファイル

- 新規: `features/owners/api/get-owners-view-scope.ts`、`lib/owners-clinic-selection-storage.ts`、`lib/owners-clinic-scope-model.ts`、`hooks/use-owners-clinic-scope.ts`
- 変更: `features/owners/loaders.ts`、`routes/OwnersList.tsx`、テスト
- 不変: `hooks/use-clinic-scope.ts`（他 3 caller の意味を変えない）、`features/auth/**`。1 PR。未結合: `GET /me` 新 field の BE 実装。

---

## 7. 共通の閉じ方・PR 本文

- 各 PR は「実装 + 契約 fixture（MSW、合成値のみ）+ 画面状態を記述した test」で閉じる。試験実行・Docker・隔離環境は行わない。
- PR 本文に記載: 対応チケット / 実装範囲 / 参照した API 契約（path・schema 名・採用 SHA）/ モック利用箇所 / 未結合のバックエンド / NOT_RUN 一覧 / 残件。
- 生成型（`src/types/generated/`）は `make codegen` がユーザー実行のため、FE PR は feature 内の手書き型 + zod で先行。生成型への切替は統合担当の整合作業。

---

## 8. 再点検で自己解決した項目（第 2 版 §8 を維持。追加分）

| 項目 | 結論 | 根拠 |
|:--|:--|:--|
| 474 の頭別 UI | 1 appointment = 1 pet。1 画面 1 form、複数頭は予約ごと | PR #748 `RevisitIntakeLIFFRead`「current_target は単数」 |
| 474 の離脱時 | dirty のときだけアプリ内離脱警告。OS 終了保証なし。server draft なし | technical-spec v3「FE は…」 |
| 475 の staff 可視範囲 | 既存 view grant + clinic + 各 target.pet 行認可で全 revision。LINE author 一致は強制しない | 同上「認証・読取の境界」 |
| 477 の binding 3 状態 | unbound / mismatch / bound を schema で分離 | `RevisitIntakeRecordRead` |
| 463 の保存媒体・key | `localStorage` `owners-list.clinics.v1.<user.id>` | PR #746 technical-spec |
| 463 の scope 取得元 | `GET /me` の `owners_view_clinic_ids`。所属全件・現在医院・system_admin から FE で補完しない | 同上 |
| 487 料金欄 | 発行前に手動修正可、会計へ逆反映しない | 本人回答 2026-10-20 14:15 UTC（EMR-518 本文） |
| 487 料金の既定値 | 一律に決めない。医院別パッケージ構成（518）に従う | EMR-518 本文 |

Plane 保守事項: EMR-463 本文の対象 path `owner/http_routes.go` は実データ経路 `GET /v1/pets` と一致しない（第 2 版から継続）。

---

## 9. ChatGPT Dots への確認事項（未決の業務判断のみ）

| # | 票 | 未決事項 | FE 側の暫定前提 |
|:--|:--|:--|:--|
| Q-A | EMR-487 / 486 | 狂犬病予防注射済証の対象接種の識別基準。ワクチンマスタに証明書種別属性が無く、`certificate_kind`（normal / rabies）を何から決めるか。518 は「名称一致で接種種別を推定しない」とのみ定める | 発行画面でスタッフが様式を明示選択。FE は自動判定しない |

第 2 版の Q-B（料金欄の入力元・必須性）は本人回答で「発行前手動修正・会計へ逆反映なし・構成は医院別」と確定したため解消。既定値の出所は 518 の設計へ委ねる。

---

## 10. EMR-518 汎用診療項目パッケージ：影響する画面・操作・状態・依存（設計は Dots、ここでは洗い出しのみ）

### 10.1 確定要件（EMR-518 本文）

- 治療プランで追加可能な全種類（7 種 master: 診察・処置・薬剤・検査・予防接種・健診・物販 + manual_other）を最低包含し、異種項目を混在定義・まとめ追加できる。
- 一括追加は原子的。選択・予定追加だけで実施・出庫・結果生成・機器送信・証明書発行・請求確定・通知を起こさない。package 自体を有料行にしない。各構成行は既存 owner の請求元へ接続し二重請求を防ぐ。金額は各単品の既存規則。
- 4 子票案: BE master / BE apply / FE master / FE chart。物理保存・API は未固定。

### 10.2 現行 FE で影響を受ける画面と操作

| 画面 / 操作 | 現行の事実 | package 導入時の影響点 |
|:--|:--|:--|
| 診察/治療プラン タブ `MedicalRecordDiagnosisPlan.tsx` → `TreatmentSearchDialog` | 1 クリック 1 件追加（`POST /v1/medical-records/:id/treatments`）。候補は 診察/処置/予防/検査(=checkup_types)/薬剤。予防・検査は `item_type: other` で master ID を失う。`memo` に category 名 | package 検索・内訳 preview・まとめ追加の入口。種別付き参照（exam_types / vaccines / checkup_types / 物販）を保持する行型が FE 側にも必要 |
| 治療 タブ `TreatmentsTab` / `use-treatments-tab.ts` | 型 picker（診療/処置/薬品/その他）、投与経路を `memo` へ | 同上。package 展開後の行編集・削除の単位 |
| `TreatmentSearchDialog`（shared） | 単一選択、multi-select なし、exam_types・物販なし、`clinicId` を vaccine/checkup hook に渡していない | package タブ or 別ダイアログ。record 医院での候補取得（FR-plan-domain-tabs） |
| `lib/treatments-tab-model.ts` `buildMasterSelectionPayload` | 保険 null で throw（EMR-454）、`inventory_id` 未設定 | 種別ごとの payload 分岐と snapshot 保持 |
| `hooks/use-copy-treatment-details.ts` | 逐次 POST、途中失敗で rollback 不可（唯一の複数追加経路） | 原子的一括追加 API（BE apply）が必要。FE は逐次 POST で package を模倣しない |
| 予防接種 / 定期健診 / 検査 タブ | 治療行との連動なし（FR-medical-records-plan-domain-tabs は未実装） | package 展開行を各タブへ「未実施」表示する接続（EMR-516/517/500 の共通予定 identity 待ち） |
| 会計(医師確認) `MedicalRecordBillCheck.tsx`、`billing_item_unbilled.go` | 治療行 + 接種 + 検査を別々に未請求候補へ。予防を other 行で追加すると接種記録と二重候補 | package 構成行の単一請求元（EMR-481/484/505） |
| 見積書 `MedicalRecordEstimate.tsx` | 治療行と無関係なローカル行 | package からの見積展開は対象外（要確認なら Dots） |
| 診療マスタ `/settings/treatment-items`（`TreatmentPlanMaster`） | 5 タブ（診察/検査/処置/予防接種/定期健診）、薬剤・物販は別ページ、親子区分 | package 定義 CRUD（FE master）の配置先。既存 CRUD を再作成しない |
| 証明書（487） | 狂犬病の登録料・済票料は package 構成の利用例 | 証明書料金欄の既定値は package 構成から参照し得るが、会計へ逆反映しない境界を維持 |

### 10.3 依存（FE から見た必要 subset）

| 依存 | 必要な固定内容 |
|:--|:--|
| EMR-516（共通予定 identity）/ 517 / 500 | kind 別参照・snapshot・read DTO。3 タブ表示の予定 identity |
| EMR-447 / 454 / 455 / 456 | 単品 master の保険・採用値・版 |
| EMR-489 | operation key・receipt・結果照会（一括追加の同 key 再送・unknown） |
| EMR-481 / 484 / 505 | 単一請求元・実施適格性・金額 snapshot |
| EMR-494 / 495 | 編集権・draft、カルテ保存/終了/訂正境界 |
| EMR-446 / 466 | DDL/ERD |

### 10.4 今回の 4 票との境界

- 474/475/477/463 は package と無関係。実装に package 対応を混ぜない。
- 477 の read-only カードは治療プランタブ内に置くが、治療行・package 行の state に接続しない。
- 487 は料金欄の手動編集を FE で持ち、既定値の出所だけ 518 へ委ねる。

---

## 11. 実装進捗（取りまとめ記録）

更新: 2026-10-20 16:55 UTC。4 PR とも CI（AgentShield / Gitleaks / Workflow Contracts / Storybook / agent-config）は最新 head で green。テストは合意どおり vitest / type-check / eslint / prettier / design-audit / 実機のすべてが NOT_RUN。マージ判断は Dots。

| 票 | ブランチ / PR | head | レビュー対応 | 状態 |
|:--|:--|:--|:--|:--|
| EMR-475 | `feat/emr-475-reservation-revisit-intake-fe` / #752 | 5d27d1cbc | CodeRabbit 3 件修正（nested 検証 fail-closed、403/404 を cached pages より優先、対象変化で再読込）。Codex の 401 cached 描画は共通 axios interceptor が refresh 失敗時にログインへ遷移するため実経路なしと返答（共通 auth は scope 外） | green / mergeable clean / 未解決スレッドなし |
| EMR-477 | `feat/emr-477-record-revisit-intake-fe` / #753 | e315c46 | Dots 回答反映（空状態文言、固定 target 内履歴のみ、unbound/mismatch 非補完）。CodeRabbit 4 件・Codex 11 件修正: `target_relation` 閉 enum、`target_id` 必須、bound 3 組合せ検証、任意項目の null/型/許容 code 厳格化、visit_purpose 空拒否、head revision 一致、next_cursor 厳格化、form 版 1 固定、received_at RFC3339、pet 照合、履歴の relation 非依存、ページ間 binding/target/回答状態 整合、既取得データ保持の inline 再試行、disabled fieldset 外の `leading` slot、skeleton 読込、ID/revision の正規形十進（BIGINT 範囲）検証、履歴 revision ≤ head、空 cursor 拒否。CodeRabbit 追加 2 件（RFC3339 の暦・時刻範囲検証、閏秒 :60 の表示正規化）も修正 | green / mergeable clean / 未解決スレッドなし。CodeRabbit Merge Risk Minimal（commit 連続で自動レビュー一時停止中。最終 head は `@coderabbitai review` で単発レビュー要求済み） |
| EMR-463 | `feat/emr-463-owners-clinic-scope-fe` / #754 | 884ff1f | Codex 4 件修正（`/me` 更新で再検証、連続トグル合成、暗黙選択の replace、BE 先行問題）。Dots 推奨により新 scope 経路を owners feature 内の静的 activation `OWNERS_CLINIC_SCOPE_ACTIVE = false` の背後に置き、非活性時は PR 前（main 05da02f3）と同じ一覧挙動。応答を見た fallback ではない | green / mergeable clean / 未解決スレッドなし。Codex findings なし、CodeRabbit Merge Risk Low。Trivial nitpick（legacy/scoped loader の query 解析重複）は次のコード push に同乗 |
| EMR-474 | `feat/emr-474-liff-revisit-intake-fe` / #756 | a00f81d | Gitleaks 誤検知は `.gitleaksignore` fingerprint で解消。CodeRabbit 1 件・Codex 14 件修正: TARGET_CHANGED 後のブロック、429/503/照会/読込の Retry-After 尊重と期限の draft 保持、401 非再試行、head 一致 refine、h1、結果応答の operation_id/target 照合、新規送信の 403/404 で pending 破棄、Enter 暗黙送信防止、head 変化時の入力同期、読込応答の予約/target 照合、pet 名の検証付き表示、古い draft の baseline token 保持と明示確認（不可視 head 時は文言分岐）、送信中/未確定時の離脱確認、履歴追加読込の Retry-After 期限、route 共通の期限中は再読込ボタン無効、「さらに表示」の 401/403/404 で保持 read を破棄、期限中の再取得は保留して期限後に自動実行、履歴は版 1 の回答のみ表示、date-time の RFC3339 検査、「さらに表示」で target が変わったページは破棄して再読込、成功 receipt は operation_id・result_refs・committed_versions で照合（current_state 非依存） | CI green（a00f81d）/ 未解決スレッド: Dots 判断待ちの複数頭・pending 予約 CTA・BIGINT appointment ID、500 の Retry-After（契約の記述が食い違うため未実装）。Codex が a00f81d を再レビュー中、CodeRabbit は 1 時間 10 回の枠切れで a00f81d 未レビュー |
| EMR-487 | `feat/emr-487-certificate-fe` / #763 | c79728a89 | Fable レビューの HIGH（再印刷の同 key ガード）・MEDIUM・LOW を修正 | open。CSP が Dots 判断待ち（§11.2） |

### 11.1 2026-10-21 引き継ぎセッションの追記

- #756: Codex 未解決 7 件のうち、feature 内で閉じる 2 件を修正した。1 件目は閲覧 403/404 の後も未確定送信の結果照会を残すこと（本文・回答・再送は出さない）。2 件目は読込の Retry-After 期限を画面の開き直しをまたいで保持すること（module memory。localStorage は使わない）。Fable レビューは APPROVE（CRITICAL/HIGH 0）。指摘された MEDIUM（拒否画面の照会結果と期限を draft store へ書き戻す）も修正した。head は b3478c9d1（38518f84f・18b450fff・b3478c9d1）。対応した 2 スレッドは返信して resolve した。pending 予約 CTA と 500 の Retry-After は、下記のとおり Dots 判断待ちのまま実装しない。
- EMR-487: #760 の契約草案を読んだ結果、3 様式の描画は server 側 487 renderer（versioned template/font/PDF）が所有する。FE は印刷ボタンのない PDF 確認 preview を表示し、印刷はその PDF だけを使う（#760 契約文書 31・100・108 行）。このため FE 側の React 帳票 renderer は作らない。二重管理になるため、試作した commit（`feat/emr-487-certificate-renderer-fe` c90104a53、未 push）は採用しない。§5 冒頭の「renderer・golden（487-A）の準備だけ先行可」は server renderer を指すと読み替える。FE の着手は、#760 の root レビュー確定後（契約 239 行「公開 PR head を固定契約」）に、9 操作の mock adapter、発行前入力（10 欄、料金 4 欄は null と明示 0 を区別、合計は自動計算しない）、PDF preview、発行、edition 読取、PDF 取得の各画面で行う。

### 11.2 Dots 回答（2026-10-21）の反映

| 項目 | Dots 回答 | 反映 |
|:--|:--|:--|
| #760 | マージ済み（main 6cb8184ca） | EMR-487 FE を #763 で作成 |
| #753 / #754 | マージ判断は実装セッション | 本人の指示（今回の 2 件は Claude がマージ）により merge commit でマージ（10789ce48 / e4c7b7fef）。リモートの head ブランチを削除。`claim/EMR-463-C` は ID が PR と一致しないため保持 |
| #754 活性化 | 非活性のまま先行マージ可。463-BE 後に後続 PR で活性化。許可医院 0 件・失権時は安全停止、他アカウントの保存値へ fallback しない | 後続 PR（未着手） |
| #756 BIGINT ID | BE が文字列で返す（EMR-519）。それまで安全整数でない ID からは要求しない | 423deeb65 でガード |
| #756 複数頭 | 本 PR では頭数分の予約を作らない（EMR-460 の残件） | 423deeb65 で「各ペットに回答」の文言を削除 |
| #756 入口 | pending にも出し、editable に従う。受付後は閲覧可、no_show はサーバー判定 | 8a3f88d8c / 423deeb65 |
| #756 500 Retry-After | optional として尊重。待機後も「結果不明・同 key」を維持し、無条件に再送しない | 97bcdb0a4 / 0f72cb90b / d8641263c（期限後も照会が not_found / claimed_uncommitted を返すまで再送不可）。api.yaml への optional header 追補は契約側の作業 |
| #756 pet 名 | 暫定の非表示は妥当。EMR-520（Blocked） | 変更なし |
| #752 401 | 遷移でキャッシュ非表示は保証されない | 3c477667b で 401 も本文非表示にし、返答を撤回して resolve |
| 487 狂犬病判定 | `vaccines.rabies_certificate_eligible` と接種時 snapshot。true のみ狂犬病様式 | #763 は `eligibility.state=eligible` のときだけ rabies を選択可 |
| 487 編集欄 | 10 欄すべて発行前入力として optional/nullable で編集可 | #763 で実装 |
| 487 様式 | 狂犬病の「種類」は品種。日付は見本の年月日表記（通常は和暦、狂犬病は生年月日が和暦＋年齢・接種日が西暦）。用紙サイズ・向きは本人確認待ち | server renderer 側の事項（FE は PDF 表示のみ） |

#752 は 3c477667b、#756 は d8641263c、#763 は c79728a89 が head。3 PR とも未解決スレッドは 0（#763 は新規）。テストはすべて NOT_RUN。

### Dots へ残る判断事項

2026-10-21 第 2 回として Dots へ送付する内容（番号は送付文と同じ）。

| # | 区分 | 確認事項 | FE の現状・暫定前提 |
|:--|:--|:--|:--|
| 1 | マージ | #752 EMR-475 / #756 EMR-474 は未解決スレッド 0。実装セッション側の最終確認を経てマージしてよいか | マージ待ち |
| 2 | マージ | #763 EMR-487 をルート未接続（`index.ts` から export のみ）のまま先行マージしてよいか | 3 が解決するまでルート接続しない |
| 3 | #763 CSP | `index.html` / `public/_headers` / `vercel.json` の CSP は `default-src 'self'` で、`frame-src` に `blob:` の許可が無い（`object-src 'none'` も要確認）。このため blob iframe での PDF preview と印刷が実ブラウザで動かない。`frame-src blob:` を追加してよいか、その担当（共通設定の取りまとめ側か #763 か） | 共通設定のため未変更。別タブ方式は印刷導線と端末依存が増えるため非推奨 |
| 4 | #763 PDF 取得 | 共通 axios が `GET /vaccination-certificate-artifacts/{id}/content` にも `Accept: application/json` を送る。BE が受け付けるか、FE がこの操作だけ `Accept: application/pdf` を送るか | 共通 axios は変更しない。後者でも request 単位の header 指定で済む |
| 5 | #763 契約 | ArtifactDescriptor の返却位置が契約で未指定。C06 result_refs の artifact_id から content path を組み立てる扱いで確定してよいか | その扱いで実装済み |
| 6 | #763 配置 | ルート接続先を接種記録画面のどこに置くか（接種行の操作か、接種詳細の別画面か） | CSP 解決と BE 実装の後に別 PR で接続 |
| 7 | 契約追補 | api.yaml の `RevisitIntakeInternalError`（500）に optional `Retry-After` header を追補する担当と時期 | FE は追補を待たずに実装済み |
| 8 | #756 契約 | 429 / 503 を契約上「未コミット確定」と扱ってよいか | FE は 429 / 503 を結果不明とし、期限後の同 key 明示再送を許可（500 のような照会必須にはしていない）。不可なら 500 と同じく照会必須にする |
| 9 | 後続 | 463-BE（`GetMeResponse.owners_view_clinic_ids`）の完成予定 | 完成後に #754 活性化の後続 PR |
| 10 | 後続 | 487 の用紙サイズ・向き（本人確認の結果） | server renderer 側の事項 |

### Codex 単独指摘の扱い

Dots の方針（修正依頼は CodeRabbit の契約実装指摘に限定、Codex 単独指摘で scope を広げない）に対し、Codex 指摘のうち確定契約（api.yaml、技術設計、db-contracts、UI デザインガイド）か fail-closed の clinic/pet 境界に直接根拠があり、修正が当該 feature 内の小差分で収まるものだけを修正し、共通層（auth、予約 DTO、api.yaml、生成型）に及ぶものは返答のみで Dots へ回した。

担当モデルの正確なバージョン（Sonnet 5.5 / Opus 5.5）は Agent ツールの `sonnet` / `opus` 指定で起動しており、ツール側からバージョン番号は確認できない。

### 引き継ぎ（別セッション向け）

- 4 PR は全て open。マージ判断は Dots。テストは全て NOT_RUN。api.yaml・生成型・共通 auth・共通医院 scope hook・共通 MSW fixture は変更禁止。force-push / reset --hard / clean 禁止。Plane の状態・担当は変更しない。GitHub コメントは末尾に Claude Code の footer を付ける。
- 各 PR の worktree: `AnimalEkarte-emr-463` / `-emr-474` / `-emr-475` / `-emr-477`（兄弟ディレクトリ。新セッションに無ければ各ブランチから `git worktree add` で作り直す）。
- 見守りの手順: 各 PR を subscribe し、Codex / CodeRabbit の指摘は「確定契約か fail-closed の clinic/pet 境界に根拠があり、feature 内の小差分で収まるもの」だけ修正する。それ以外は返信のみで Dots へ回す。修正は担当ブランチで通常 push し、返信して該当スレッドを resolve する。Dots の判断待ちスレッドは resolve しない。
- CodeRabbit は自動停止後、`@coderabbitai review` で単発レビューを依頼できる。上限は Advanced プランで 1 時間 10 回で、上限中は実行されない。#756 a00f81d の単発レビューは枠が戻ってから 1 回だけ依頼する。
- 2026-10-21 追記: #756 の作業 worktree は `AnimalEkarte-emr-474-review`。不採用にした 4 件版は local `backup/emr-474-four-fixes-20261021` に残してある（未 push）。EMR-487 の試作は `AnimalEkarte-emr-487`（未 push・不採用）。
- 残作業: #756 a00f81d の Codex 再レビュー結果の確認、#754 の CodeRabbit nitpick（legacy / scoped loader の query 解析重複）を次のコード push に同乗、EMR-487 は EMR-486 契約待ち、この文書 §11 の head 更新。
- 後片付け: ローカルの `plan-doc-tmp` ブランチと storybook-ci の worktree は、AGENTS.md の作成者別規則に従って確認してから扱う。
