# AnimalEkarte — Unified TODO（todo.md）

> 更新: 2026-07-16 (4)（PO 決裁「即実装可」4 件消化: #211 A1+A2／#211 A6／#201 B2／PO-008 完了・台帳から除去）
> 前回: 2026-07-16 (3)（phase2 切り出し: 今フェーズでやらない項目を `phase2.html` へ全文移動・完了記録を削除。本書は「今やること」のみ保持）
> **push・外部書き込み・credential 変更はユーザー所有アクション。**（PR マージはユーザーが手動で行う。本台帳には載せない）
> **別台帳**: 今フェーズでやらないもの = `phase2.html` / BE 保留詳細 = `BE-pending.md` / PO 判断キュー = `q&a.html`
> **本書の役割**: 今フェーズに着手可能・着手すべきタスクのみの正本台帳。

---

## 運用規約

### Docker 検証規約（BE・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

### 台帳スコープ規則

- 本書には**今フェーズで着手可能なタスクのみ**を記載する。対応済みは削除する（記録は git 履歴）。
- 今フェーズでやらないもの（次期監査引き継ぎ・再開条件付き見送り・長期目標・PO 決裁済み「やらない」）は `phase2.html` を正本とする。決裁済み「やらない」は実装着手禁止のまま。着手判断が出たら実装単位として本書へ戻し、phase2.html から削除する（二重管理禁止）。
- その他 open FEAT（#249/#247/#239/#238/#237/#235/#234/#232/#230 等）は **gh を正**とし、本台帳には重複掲出しない。3セッション並行開発計画（#260）は別対応・本台帳スコープ外。
- PR マージ判断・マージ状態・マージ用チェックリストは本台帳に載せない（ユーザー手動）。
- PO 決裁の正本は `q&a.html`（PO-001〜008 回答済み）。決裁済みの「即実装可」は本書の「PO 決裁」節を正とする。
- 着手保留・任意検証の BE 詳細は `BE-pending.md`。再検討トリガが立ったら実装単位として本書へ戻す。

---

## Project TODO

### P1 — Open Issues（台帳掲載分）

| # | 内容 | 現状 |
|---|---|---|
| #211 | 検査・健診パッケージ化 | **PO 決裁済・A1+A2+A6 消化済み（2026-07-16）**。A1+A2（アドプリット削除＋尿比重 min/max 空化・commit 90553a51）は seed 編集完了・`db_reset` は USER。A6（checkup_type_fields↔checkup_types 複合 FK・commit 59aa533a）は `002_checkup_field_clinic_composite_fk.sql` として起草完了（起草時に checkup_field_results↔checkup_type_fields 側は既に 001_init.sql 実装済みと判明・対象は checkup_types 側 1 本のみに縮小）。STG 適用・DROP CONSTRAINT 対象の実制約名確認は USER。やらない分（CRUD UI／四季分割・腎臓ドック／select 異常ハイライト／ライブ E2E）は `phase2.html`。provisional 解除はクライアント臨床責任者確認後に seed 手動更新 |
| #201 | 薬量自動計算 | **実装・PO 運用確定済・残作業ゼロ**（B2 コード注記更新は 2026-07-16 消化 — audit metadata 側は該当文字列なしを grep 確認）。**gh クローズは USER**（B5 文言） |
| #89/#97/#98/#99/#109 | シークレット移行・ローテーション | **USER BLOCKED**（リポジトリ Phase A 済）。4系統ローテ / P5-2 Secrets / #97 本文マスク / #109 フォールバック撤去。詳細は SEC-SECRETS-5 |

### P2 — follow-up

- [ ] **FEAT-searchable-select-targets** — 残作業は USER 目視確認のみ（確認対象一覧は本書「個別タスク詳細」節）

### P3 — インフラ

- [ ] **[USER] P2 Terraform（internal ALB + VPC Origin）本番適用** — `infra/terraform/terraform.tfvars` はローカルに準備済み（gitignore 対象）。`terraform apply` の実行判断は USER
- [ ] **[USER] Vercel Production `VITE_SHOW_DEMO_ACCOUNTS=false` 確認** — リポジトリ外での確認が必要・未検証

### PO 決裁 — 即実装可（未消化）

> 決裁正本は `q&a.html`。対応済みは本表から除去。やらない決裁・再開条件付きは `phase2.html`。
> 2026-07-16: #211 A1+A2（commit 90553a51）・#211 A6（commit 59aa533a）・#201 B2（commit 170d9abe）・PO-008（commit 7d740c6b、`po-008-factsheet.md`）消化済みのため本表から除去。

| 優先 | ID | 内容 | 備考 |
|------|-----|------|------|
| — | PO-002 | Sentry Phase 1（例外+版数のみ・PII off） | ベンダ/課金は USER。security-review 必須 |
| — | PO-006 | ADR-003 TRIGGER Issue | **起票操作は USER**（案1B・二重保持解消もスコープに含める） |

### ユーザー所有アクション一覧

| アクション | 根拠 |
|-----------|------|
| SEC-SECRETS-5: 4系統ローテーション＋ P5-2 GitHub Secrets 登録＋ #97 本文マスク | PUBLIC 履歴露出の実効無効化。手順: runbook §0.5 / `infra/cloudflare/README.md` |
| seed 003_demo 変更後のローカル/STG `db_reset`（SEC / #211 A1+A2） | migration-seed-safety。エージェントは DB reset 自動実行禁止 |
| #211 A6: `002_checkup_field_clinic_composite_fk.sql` の実DB適用 | DROP CONSTRAINT 対象の推定制約名（`checkup_type_fields_checkup_type_id_fkey`）を `\d checkup_type_fields` で事前確認してから `go run ./cmd/migrate`。migration適用はエージェント禁止 |
| #201 B2 / #211 A6 の scoped 検証（docker 未起動のため本セッション未実行） | `docker compose exec backend go test ./internal/service/ -run 'Dose' -count=1`／`gofmt -l ./internal/service/`（B2）。A6 は静的SQLレビュー(database-reviewer)のみ実施済み・DB適用時の実SQL実行検証は未実施 |
| #109 Phase C: `STG_DEMO_*` 登録後に performance-tests フォールバック撤去（エージェント可） | Secrets 未登録のまま撤去すると scheduled が壊れる |
| Vercel Production `VITE_SHOW_DEMO_ACCOUNTS=false` 確認/設定 | 外部システム操作 |
| `terraform apply` 承認（tfvars 準備済み） | インフラ破壊的変更 |
| ADR-003（PO-006）独立 Issue 起票 | PO 承認済。案1B TRIGGER＋二重保持解消検討 |
| #201 gh Issue クローズ | PO 運用確定済。残記録文言は B5 |
| #229 gh Issue クローズ | ローカル実装完了済み |
| #214（bug.md 対応検証チェックリスト）の gh クローズ判断 | 2026-06 起票・open のまま台帳外。起票元 bug.md はアーカイブ削除済み・対応自体は 2026-07-02 完遂。存置理由がなければクローズ |
| Sentry 等ベンダ確定・課金契約（PO-002） | 課金・外部契約 |

---

## BE 残タスク

> 今期着手可能な BE 残タスクのみ。対応済みは残さない。次期送り・見送りは `phase2.html`。

| ID | 優先度 | 内容 | 状態・条件 |
|----|--------|------|-----------|
| SEC-SECRETS-5 | **USER 残** | リポジトリ Phase A 済。**残（credential-impacting）**: 4系統ローテ、P5-2 `gh secret set`、#97 本文マスク、#109 `STG_DEMO_*` 登録後のフォールバック撤去。#98/#99 は Phase 8 まで PENDING | Issue クローズはローテ完了後。seed 変更後は `db_reset`（USER） |

---

## 別台帳ポインタ

| 台帳 | 役割 |
|------|------|
| `phase2.html` | 今フェーズでやらないものの正本（次期監査引き継ぎ・見送り・長期目標・やらない決裁） |
| `BE-pending.md` | 着手保留・次期送り・任意検証の BE 詳細 |
| `q&a.html` | 内部 PO 判断キュー（決裁記録の正本。PO-001〜008 回答済み） |

> 旧 `todo.md` / `BE_todo.md` / `BE-refactor.md` / `FE-refactor.md` は本ファイルへ吸収済み（削除）。旧 `docs/tasks/`・`docs/archive/` は 2026-07-16 に廃止（詳細は git 履歴）。

---

## 個別タスク詳細

### FEAT-searchable-select-targets: 検索可能 Combobox 化（FE・実装完了、目視確認のみ残）

- **実装状況**: P1〜P3 全実装完了（type-check/lint/隣接テスト green）。SearchableSelect = `frontend/src/components/ui/searchable-select.tsx`。適用済み: 予約区分・担当者(`ReservationFormFields.tsx:334,416`)、診断名1/2+カテゴリ(`DiagnosisHeaderDiagnosis.tsx:52,58,64`)、診療計画病名(`ClinicalPlanSection.tsx:47`)、主訴(`InterviewChiefComplaint.tsx:45`)、ワクチン(`VaccinationForm.tsx:72`)、検査種別・担当医(`ExaminationFormFields.tsx:56,63`)、健診種別・担当医(`CheckupForm.tsx:111,143`)、入院ケージ(`HospitalizationBasicInfo.tsx:106`)、薬剤親カテゴリ(`MedicineSidePanelSections.tsx:67`)、指名フィルタ(`ReceptionFilterPanel.tsx:59`)、医師フィルタ(`ReservationManagementCalendar.tsx:85`)、動物種(`NewOwnerInlineForm.tsx:83`/`PetEditModalFieldSections.tsx`)、スタッフフィルタ(`ShiftCalendar.tsx:107`、per-option `disabled` 追加)。
- **意図的スキップ**: `ShiftFormDialog` テンプレ選択（非制御アクショントリガー）／`ReservationTypeSidePanel` グループ選択（カラードット custom JSX・実件数<15）。保留候補: Lステップ TriggerType（`LstepDeliveryMonitorPageParts.tsx:71`）。
- **残作業**: [USER] 目視確認（検索・スクロール・選択・カスケード・per-option disabled）。

---

## docs/ 再編（2026-07-16）残課題

- [ ] **[要再検証] BUG候補: `use-reception-kanban.ts` の既存 type エラー** — 実体パスは `frontend/src/features/reception/hooks/use-reception-kanban.ts`（旧記載の `src/hooks/…:18` は誤り）。起票後にリファクタ済みで、現 line 18 に明白な型エラーはない。`docker compose exec frontend pnpm type-check`（USER 実行）で再現しなければクローズ。
- [ ] **[USER・任意] Notion EkarteSprint 文字化け3語の目視確認** — 2026-07-15 の保留9件適用は完了（読み戻し 9/9 PASS）。転送時に文字化けした3語（します／共有済み／事前提供）の適用先ページ（クレジット訂正フロー／検査④機器データ取込／検査⑥自動連携調査）の該当文のみ目視確認できればクローズ。

## 受け入れシナリオ作成（2026-07-16）で発見した仕様・実装ギャップ

- [ ] **[SPEC-GAP・PO 判断] カルテ確定（Lock）の UI 導線が存在しない** — specification §2.1 は「確定（Lock）フローを実装」と宣言するが、カルテ画面仕様（06 §2.3 に明記）に確定ボタンはなく API 直接操作でのみ確定可能。臨床記録の真正性担保が現場で使えない。UI 導線を作るか §2.1 の記述を実態に合わせるかの判断が必要。受け入れシナリオ S06 は暫定で API 確定（USER 実施）前提。
- [ ] **[SEED-GAP] seed 003_demo の全標準ロールが検査 edit 権限を保有** — 権限差による確定ロック検証（受け入れシナリオ S02）が現 seed では実施不能。閲覧専用相当ロールの seed 追加か、STG 実権限グループでの検証かの判断が必要。

### AnimalEkarte CSV import — USER actions

> 方針（2026-07-15 確定）: フル 003_demo（~529MB・PHI 含みうる）は **Git に載せない**。正本バックアップは `old_db/sensitive-local/animalekarte-003-demo-full/`。リポジトリの `003_demo` は小さいデモのまま。

- [ ] **USER:** ローカルでフル seed を使う: `rsync -a ../old_db/sensitive-local/animalekarte-003-demo-full/ backend/migrations/seeds/003_demo/` のあと `make reset`（エージェントは reset しない）。誤 `git add` 防止のため該当 CSV に `skip-worktree` 推奨。
- [ ] **USER:** STG へのフル seed 適用は別途承認・手動実行（通常は小さいデモのまま）。
