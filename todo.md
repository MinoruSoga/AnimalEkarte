# AnimalEkarte — Unified TODO（todo.md）

> 更新: 2026-07-20（台帳再編: `bug.md` を「バグ台帳」節として本書へ統合。PO 確認・決裁・USER 操作系（旧 U-1〜U-13・P1 表・FEAT-CHECKIN・`task.html`・`todo-me.html`・`po-008-factsheet.md`）は最終的に [`q&a.html`](q&a.html) へ一本化（一時 `todo-po.html` を経由し 2026-07-20 に統合・削除）。経緯は git 履歴）
> **本書の役割**: エージェントが着手できる作業（バグ・実装タスク・運用規約）の正本台帳。読者 = 次に着手するエージェント（前提知識ゼロで本書だけから作業に入れる粒度で書く）。
> **PO 確認待ち・決裁記録・USER 実操作・ブロッカー状態の正本 = [`q&a.html`](q&a.html)（2026-07-20 一本化）**（push・外部書き込み・credential 変更・PR マージはユーザー所有アクション）。
> **別台帳**: 今フェーズでやらないもの = `phase2.html` / BE 保留詳細 = `BE-pending.md`

---

## 運用規約

### Docker 検証規約（BE・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

### BE 実装規約（BE9 domain package 移行中・2026-07-19〜）

- **backend の新規実装は `internal/handler|service|repository` へ追加禁止**（BE9-2B 発効）。ADR-006 の target domain package へ追加する。先例 = `internal/manualarticle` / `internal/medicalrecord`、配線・facade・固定 gate 追随のパターン = `BE-refactor.md`「BE9-2B/2C実績からの申し送り」。
- 本書が参照する backend ファイルパスは BE9 移行で順次移動する。**着手時に `docs/architecture/be9-2a-classification-manifest.csv` で現在地と target package を確認する**（パス不一致は台帳の誤りではなく移行の進行）。移動済みファイルへの言及を見つけたら、修正作業のついでに台帳側パスも更新する。
- ADR-006「未解決論点」に着手前ゲートが設定された領域（`billing_item_repository.go` = BUG-417 等）に触れる場合は、`BE-refactor.md`「現在地と着手前ゲート」の該当条件を先に満たす。

### 台帳スコープ規則

- 本書には**エージェントが着手可能な作業のみ**を記載する。対応済みは削除する（記録は git 履歴）。
- **「対応済み」判定は gh の state を実測してから行う**（`gh issue view <n> --json state`）。ローカルの消化記録だけを根拠に除去しない（2026-07-16、#201 を除去した後に [SAFETY] として再オープンされ、1 日間どの台帳からも不可視になった実例）。臨床安全・credential 系ブロッカーの状態表は `q&a.html`「P0 ブロッカー」が正本。
- 今フェーズでやらないものは `phase2.html` を正本とする。決裁済み「やらない」は実装着手禁止のまま。着手判断が出たら実装単位として本書へ戻し、phase2.html から削除する（二重管理禁止）。
- **open Issue の正本は gh**（`gh issue list --state open`）。本書に番号リストを重複掲出しない（列挙は必ずドリフトする）。
- PO 決裁の正本は `q&a.html`。決裁が下りて着手可能になった作業だけを本書へ実装タスクとして起票する。
- 着手保留・任意検証の BE 詳細は `BE-pending.md`。再検討トリガが立ったら実装単位として本書へ戻す。

---

## バグ台帳（正本・旧 bug.md を 2026-07-20 統合）

> **運用**: 受け入れシナリオ・レビュー・実機で発見した不具合は本節へ BUG-XXX で起票する（BUG-4xx = 受け入れシナリオ由来）。修正完了したら本節から削除し、修正コミットと発見経緯は git 履歴・実行レポート（`docs/ops/testing/scenarios/reports/`）を正とする。バグではない対応事項は本書の他節または `q&a.html`（USER 操作・決裁）。
> **粒度**: 次に着手するエージェントが本節だけで調査に入れること（症状・再現手順・調査済みの根因・修正方向）。
> **BE9 注記**: backend は domain package へ移行中（`BE-refactor.md` BE9 / ADR-006）。本節が参照する backend パスは移行で移動する — 着手時に `docs/architecture/be9-2a-classification-manifest.csv` で現在地を確認し、修正が backend 新規実装を含む場合は target domain package へ実装する（上の「BE 実装規約」参照）。

### Open

#### BUG-418:【LOW・潜在・BE9-2D ⑤ Phase1敵対レビューで検出】DischargeWithBilling（退院+会計自動生成）に監査ログ書込がない

- **症状**: `hospitalizationService.DischargeWithBilling`（BE9 target=medicalrecord・現在は `internal/service/hospitalization_service.go`）は退院status更新+Billing/BillingItem行の自動生成を単一txで行う金銭パスだが、**audit_logs への書込が一切ない**。会計レコードの自動生成が「誰が・いつ・どの入院から」発生したか監査証跡で追えない。
- **現状の安全性**: 旧実装から一貫して audit なしであり、⑤ Phase1（tx機構refactor・`f93299f1c`後の同名コミット）はこの状態を変更していない（behavior-preserving・レビューで「既存負債・悪化なし」判定）。HTTP アクセスログ+billings 行自体の created_at で部分的な追跡は可能。
- **修正方向**: tx 内 fail-closed 監査（`AuditTxLogger.LogEntryTx`・medicine/dose-param/treatment逸脱監査と同型）を DischargeWithBilling の tx 閉包末尾へ追加。Action 定数（例: `hospitalization.discharge_with_billing`）+ ResourceID=hospitalizationID + NewValue に billing_id/合計額。**機能追加=behavior変更のため BE9 の移動batchに混ぜない**（⑤ Phase2 移動後の別コミット、または medicalrecord 移行完了後）。
- **発見**: 2026-07-21（BE9-2D ⑤ Phase1 敵対レビュー・clinic-isolation-auditor MEDIUM 指摘）。

#### BUG-417:【LOW・潜在・BE9-2A監査で検出】billing_item_repository.go の Update/Delete が clinic 分離を実質担保していない（defense-in-depth 不全・現状は生きた漏洩ではない）

- **症状**: `internal/repository/billing_item_repository.go`（BE9 target=billing）の Update/Delete が `.Joins("JOIN billings ON ...billings.clinic_id=?...")` を `.Updates()`/`.Delete()` へ連結する形式だが、**GORM の `Joins()` は UPDATE/DELETE SQL へ伝播しない**ため、repository 層の clinic 述語は実質 no-op。Treatment/ClinicalPlan 等が subquery 形式で正しく回避している同型の罠に、このファイルだけが該当（billing_confirmation/estimate は検証済みで正しい）。
- **現状の安全性**: `billing_item_service.go` の UpdateItem/DeleteItem が事前に clinic-scoped `FindByID` で gate しているため**現時点で生きた漏洩ではない**。ただし事前 check を経由しない新規経路（admin 経路・background job 等）が repository method へ直接到達すると silent なクロステナント書き込み/削除が発生し得る。クロステナント分離 test も現状ゼロ。
- **修正方向**: subquery 形式（`WHERE id IN (SELECT ... JOIN billings ... WHERE billings.clinic_id=?)`）への是正＋クロステナント分離 test 追加。**詳細の正本 = `docs/architecture/be9-2a-boundary-map.md` §7.4／ADR-006 未解決論点#6**。
- **修正タイミング**: BE9-2C/2D の billing domain 着手時の**必須前提**（ADR-006 で着手前ゲート化済み）。ただし BE9 と無関係にこのファイルへ触れる修正が先に発生した場合も、その場での是正を必須とする。
- **発見**: 2026-07-19（BE9-2A santa dual-review round 2・clinic-isolation-auditor。BE9-2A は measurement-only のため未修正のまま記録）。

#### BUG-416:【LOW・healthcare-reviewer指摘】カルテ診断(diagnosis1/2)保存の残存リスク（BUG-410 backend/UI follow-up）— 残るのは①④のみ（②FE病名バリデーション欠如=修正済み 08c82490／③clinical_plan楽観ロック欠如=修正済み 797f4d2d）

- **経緯**: BUG-410（構造化診断 hydrate 欠落・修正済み 1407a39a）の独立監査（healthcare-reviewer、2026-07-18、APPROVE・3件とも非ブロッキング）で指摘。②は 2026-07-18 commit `08c82490`、③は 2026-07-19 commit `797f4d2d`（clinical_plan_request.go に Version フィールド追加・clinical_plan_repository.go の UPDATE に `version = ?` 述語追加・FE の `use-medical-record-save-action.ts` も version 送信に対応、medical_record と parity化）でそれぞれ修正済み。詳細は git 履歴を正とする。残る①④はいずれも「将来 UI 追加時のみ発火する前提条件」で現状未到達（2026-07-19 再監査で確認）。
- **① save-action の diagnosis1/diagnosis2 送信非対称（クリアUI追加時の前提条件）**: `use-medical-record-save-action.ts` は diagnosis1 を `?? undefined`（未送信=BE 側で「更新しない」扱い）、diagnosis2 は state の値をそのまま送信（`null` なら明示クリア）という非対称な契約になっている。現行 UI（`SearchableSelect`）には選択解除操作が無いため両方とも現状は発火しないが、**将来 diagnosis1/2 いずれかにクリアボタンを追加する場合、diagnosis1 側は `?? undefined` のためクリア操作がサイレントに no-op する**（「保存しました」トーストは出るが DB は変わらない）。クリアUI追加時はこの非対称の是正が前提条件。
- **④ レコード切替時の hydrate guard 再利用リスク（調査済み・現状未到達と判定）**: `useApplyMedicalRecord` の hydrate は `existingRecord.diagnosisXxx != null` の場合のみ setter を呼ぶ。理論上、同一 `MedicalRecordForm` インスタンスが保持されたまま record A（diagnosis2=4,9）→ record B（diagnosis2=null）へ切り替わると、B 用の setter が一度も呼ばれず A の値が state に残存し、B の保存時に A の診断が誤って書き込まれうる（データ喪失より悪いクロスペイシェント汚染）。**実コード調査で現状は再現不可と判定**: ルート定義（`frontend/src/app/routes/clinical-care-routes.tsx` の `path: ":id"`）に `key` 指定なし かつ `/medical-records/<id1>` → `/medical-records/<id2>` へ直接 `navigate()` する呼び出しはリポジトリ全体で0件（`medicalRecords.detail.getHref` の全呼び出し元＝会計一覧/健診一覧/カルテ一覧/新規作成auto-createはいずれも別ルートを経由してから `:id` に遷移するため React Router がコンポーネントを再マウントする）。`MedicalRecordForm.tsx` 内の来院履歴パネル（`InterviewHistory.tsx` 等）も展開/折りたたみのみで他レコードへの遷移リンクを持たない。**将来「次の来院/前の来院」等、同一画面内で record ID だけを差し替えるナビゲーションUIを追加する場合は、hydrate 全体（この guard パターンを共有する chiefComplaintTypeId/plan/assessment/notes 等も含む）を record 切替時に明示的リセットする設計が前提条件になる**。
- **重要度**: LOW（①④とも将来UI追加時のみ発火する前提条件で現状未到達）。
- **発見**: 2026-07-18（healthcare-reviewer による BUG-410 修正の独立監査、APPROVE・3件とも別チケット化推奨。④は同監査後の react-reviewer 観点フォローで発見・コード調査により現状未到達と確定）。

#### BUG-413:【要監視・#250後に判定】予防接種・トリミング一覧が同型のページネーション不可視化リスクを抱える（現状 seed 0件で無症状）

- **経緯**: BUG-412（在庫一覧の偽ページネーション）の派生監査（2026-07-17）で発見。BUG-412 は 2026-07-17 に修正済み（`InventoryList`/`use-inventory.ts`/`frontend/src/features/inventory/api/inventory.ts` をサーバサイド page/limit 転送 + 実 total 消費に変更。backend `inventory_repository.go` は元々 clinic-scoped 実 COUNT を WHERE 適用後に返しており変更不要だった）。本エントリは BUG-412 調査で判明した**同一パターンの潜在リスク**のうち、今回は修正しないと判断したものを引き続き追跡する。
- **VaccinationList.tsx（予防接種一覧）**: `use-vaccinations.ts` → `get-vaccinations.ts` の `getVaccinations` は `start_date`/`end_date` を日付フィルタ使用時のみ送る任意パラメータで、未指定時は無条件で backend `parsePagination` の既定 `limit=20` に落ちる（date-scoped ではない）。`backend/migrations/seeds/003_demo/vaccinations.csv` は現状 0 件（ヘッダのみ）で無症状・検証不能。
- **TrimmingList.tsx（トリミング一覧）**: `use-trimming-records.ts` → `get-trimmings.ts` の `getTrimmings` は明示的に `page:1, limit:HISTORY_FETCH_LIMIT`（`frontend/src/config/fetch-limits.ts` で `100` = backend `defaultMaxPaginationLimit`）を送信しており `limit=20` の欠陥ではないが、1リクエストの上限100件を超えた場合の継続取得導線が無い。トリミング実績は `reservation_type.category='trimming'` の予約（`appointments.csv`）であり、現状 0 件で無症状・検証不能。
- **今回修正しない理由**: 両画面とも seed が 0 件のため、ページネーション化しても page=2 が別レコードを返すことを実測で反証できず、「直った」ことを検証できない。加えて両フックとも医師/ステータス/種別等のクライアントサイドフィルタを持ち、BUG-411/412 と同様「フィルタがページ内スコープに退行しないか」の再設計が必要（Trimming はさらに limit=100→カーソル/オフセット継続取得という追加拡張が要る）。検証不能な状態での臨床データ一覧の書き換えは、修正漏れより悪い結果（未検証の新規回帰）を生みうるため見送った。
- **要監視**: #250（本番データ移行）でこの2画面の実件数が判明した時点で、実件数が limit（Vaccination=20 / Trimming=100）を超えるか再確認し、超える場合は本エントリを起票根拠に BUG-411/412 と同型の修正を行う。#250 の受け入れ条件にこの再確認を明示的に含めること。
- **発見**: 2026-07-17（BUG-412 対応中の派生調査、コーディネーターが claim を再検証済み）。

> BUG-408（ワクチン動物種フィルタ）は 2026-07-20 に代理決裁済み（`q&a.html` DEC-3）— 実装タスク FEAT-VACCINE-SPECIES として本書「個別タスク詳細」へ戻し済み。

## 個別タスク詳細

> task-create スキルの起票先は本節。以下は 2026-07-20 の代理決裁（正本 = `q&a.html` DEC-2〜8 カード）で着手可能化されたもの。FEAT-CHECKIN（DEC-2）は実装完了済みのため本書から削除。

### FEAT-VACCINE-SPECIES: ワクチン選択の動物種フィルタ【DEC-3 裁定済み・着手可】

- **裁定**: 案(b) `AnimalSpecies` マスタへ `species_category` enum（`dog`/`cat`/`other`）を新設し、`Vaccine.species`（dog/cat/both）とこの enum でマッチングする。裁定根拠 = `q&a.html` DEC-3。
- **実装**: ①migration 002 でカラム追加（既存レコードは name パターン推定で backfill・既定 `other`。`migration-seed-safety` スキル必読・DB_RESET 要否を明記） ②FE フィルタは「一致+`both` をデフォルト表示・『全て表示』トグル」— hard block しない ③適用先 = `VaccinationFormPanels.tsx` と姉妹フォーム `MedicalRecordVaccination.tsx` の両方（parity 維持） ④マスタ管理 UI に category 編集を追加。
- **BE 注意**: vaccine/vaccination は BE9 target=medicalrecord（sub-batch②領域 — 並行 batch と衝突しないか着手前に worktree/git log 確認）。

---

## 完了済みトピックの参照先（本書には残さない）

- **画面仕様書全数突合 SD-1〜19＋GAP-1/2**: 2026-07-17 に全 21 件消化（実装コミット 142f5ebe〜6d10f4c0）。決裁と乖離裁定の正本 = `q&a.html` 各カード、GitHub 入口 = #261。副産物（BE finalized ガード欠落 5 件封殺・liff.state 復元前読取修正・dbOrTx allowlist 回帰）も同コミット群で対処済み。次期送り 6 件 = `phase2.html`「SD/GAP fix all の副産物」節。
- **docs/spec/screens**: 全 65 doc が実装と同期済み（drift-gate green・画面数 40）。以後は実装コミットに doc 更新を同梱する運用。
- **PR #186 レビュー残**: 決裁 = `q&a.html` DEC-4〜7（裁定済み）／値投入 = `q&a.html` OPS-8（旧 task.html。Resolve 済みスレッドは PR を正）。

---

## 別台帳ポインタ

| 台帳 | 役割 |
|------|------|
| [`q&a.html`](q&a.html) | **PO 確認待ち・決裁記録・USER 実操作の統合正本**（PO判断待ち DEC-1/9/10・代理決裁 DEC-2〜8・PO-001〜008・SD-1〜19・GAP-1/2・P0 ブロッカー状態表・OPS-系実操作・Cloudflare 移行人間タスク・PO-008 ファクトシート） |

| `phase2.html` | 今フェーズでやらないものの正本（次期監査引き継ぎ・見送り・長期目標・やらない決裁） |
| `BE-pending.md` | 着手保留・次期送り・任意検証の BE 詳細 |
| `BE-refactor.md` | **BE9 active plan**: Go/Gin公式原則を適用し、巨大なhandler/service/repositoryをdomain/resource packageへ大規模移行する。package非依存lint・HTTP/security review・旧layer撤去までを含む。旧BE8固定layer計画はsuperseded historyとして実行禁止 |

> 旧 `BE_todo.md` は本ファイルへ吸収済み（削除）。`FE-refactor.md` は 2026-07-18 に対応完了・削除済み（恒久規約は `frontend/CLAUDE.md` へ同梱）。BE8 の固定 layer 計画は 2026-07-19 に ADR-005 で superseded となり、backend 正本を `.claude/rules/go-gin-backend-guidelines.md` へ置換。**2026-07-20 台帳再編**: `bug.md` → 本書「バグ台帳」節へ統合、`task.html`・`todo-me.html`・`po-008-factsheet.md`（→ 一時 `todo-po.html` → 2026-07-20 `q&a.html` へ最終統合）はいずれも削除（経緯は git 履歴）。旧 `docs/tasks/`・`docs/archive/` は 2026-07-16 に廃止。
