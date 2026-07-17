# AnimalEkarte — Unified TODO（todo.md）

> 更新: 2026-07-17 (2)（**#201 [SAFETY] を P1 へ復帰** — 2026-07-16 に再オープンされていたが「B2 消化済み」で除去されたまま全台帳から不可視だった。臨床安全の最優先 USER ゲートを可視化）
> 前回: 2026-07-17（fix all: 決裁済み実装キュー SD-3〜19＋GAP-1/2 全 21 件を消化・コミット 142f5ebe〜6d10f4c0。乖離 3 件と USER 残は「SD 全件消化済み」節参照。**エージェント着手可能タスク 0 件 — 残りは全て USER アクション**。#201 の実装も個人責任者ゲート承認後に着手可）
> 前回: 2026-07-16 (6)（SD 残 14 件 + GAP-1/2 を Fable 代理決裁で全件確定・実装キュー化。決裁正本 = q&a.html）
> 前回: 2026-07-16 (5)（画面仕様書全数突合の副産物 = 実装バグ疑い 19 件を起票。突合本体は commit a476b727・未文書化3画面の doc 新設で SD-14〜19 追加発見）
> 前々回: 2026-07-16 (4)（PO 決裁「即実装可」4 件消化: #211 A1+A2／#211 A6／#201 B2／PO-008 完了・台帳から除去）
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
- **「対応済み」判定は gh の state を実測してから行う**（`gh issue view <n> --json state`）。ローカルの消化記録だけを根拠に除去しない。除去後に再オープンされても本書は自力で気付けないため、**除去は gh CLOSED の確認とセット**とする（2026-07-16、#201 を「B2 消化済み」で除去した後に [SAFETY] として再オープンされ、臨床安全の最優先ゲートが 1 日間どの台帳からも不可視になった実例）。
- 今フェーズでやらないもの（次期監査引き継ぎ・再開条件付き見送り・長期目標・PO 決裁済み「やらない」）は `phase2.html` を正本とする。決裁済み「やらない」は実装着手禁止のまま。着手判断が出たら実装単位として本書へ戻し、phase2.html から削除する（二重管理禁止）。
- **open Issue の正本は gh**（`gh issue list --state open`）。本書に重複掲出しない。列挙は必ずドリフトするため番号リストを本書に持たない（2026-07-17: 旧列挙が #250〜#262 を落としていたため撤去）。3セッション並行開発計画（#260）は別対応・本台帳スコープ外。
- 例外 = 上の「P1 — Open Issues（台帳掲載分）」。**臨床安全・credential 露出など、gh に埋もれると人命・情報漏洩に直結するものだけ**を本書へ掲出する。掲出判断は内容の重大性で行い、件数を増やさない。
- PR マージ判断・マージ状態・マージ用チェックリストは本台帳に載せない（ユーザー手動）。
- PO 決裁の正本は `q&a.html`（PO-001〜008 回答済み）。決裁済みの「即実装可」は本書の「PO 決裁」節を正とする。
- 着手保留・任意検証の BE 詳細は `BE-pending.md`。再検討トリガが立ったら実装単位として本書へ戻す。

---

## Project TODO

### P1 — Open Issues（台帳掲載分）

| # | 内容 | 現状 |
|---|---|---|
| #201 | **[SAFETY] 薬量上限超過の物理ブロックと例外統制** | **再オープン（2026-07-16）・USER BLOCKED**。自動計算・snapshot・master UI・監査の基盤は実装済みだが、上限超過を FE `ConfirmDialog`「この数量で保存」で通過でき BE も拒否しない（audit のみで保存）。確認ダイアログを安全統制に使う禁止則（product-philosophy ③）違反・fail-closed 不成立。**ブロッカー = 個人責任者ゲート**: 使用医院の臨床責任者（一人）を責任者台帳へ解決し、絶対上限／warning 範囲／情報欠落時の手動入力可否／緊急例外の要否を承認させる（USER）。承認後の実装（BE 物理 reject・権限付き例外フロー・同一 tx audit）はエージェント着手可。doc 側は `06-medical-records-form.md` に既知の是正対象として明記済み。**本表から除去禁止**（2026-07-16 に「B2 消化済み」で一度除去され、再オープン後 1 日不可視になった経緯あり） |
| #211 | 検査・健診パッケージ化 | **PO 決裁済・A1+A2+A6 消化済み（2026-07-16）**。A1+A2（アドプリット削除＋尿比重 min/max 空化・commit 90553a51）は seed 編集完了・`db_reset` は USER。A6（checkup_type_fields↔checkup_types 複合 FK・commit 59aa533a）は 2026-07-17 に `001_init.sql` 末尾（013 セクション）へ統合済み（起草時に checkup_field_results↔checkup_type_fields 側は既に 001_init.sql 実装済みと判明・対象は checkup_types 側 1 本のみに縮小。独立ファイル 002 は migration 完全統合で削除）。既存 DB への反映は USER の `DB_RESET=true` 再適用時（fresh DB は 001 適用で有効）。やらない分（CRUD UI／四季分割・腎臓ドック／select 異常ハイライト／ライブ E2E）は `phase2.html`。provisional 解除はクライアント臨床責任者確認後に seed 手動更新 |
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
| **#201 [SAFETY] 個人責任者ゲート**: 使用医院の臨床責任者（個人）を責任者台帳へ一意に解決し、絶対上限／warning 範囲／情報欠落時の手動入力可否／緊急例外の要否を承認させる | 誤投薬 0 の業務目的に対し、現行は上限超過を確認ダイアログで通せる。承認なしでは実装仕様が確定せず着手不能（本ゲートが全 USER アクション中で最優先） |
| SEC-SECRETS-5: 4系統ローテーション＋ P5-2 GitHub Secrets 登録＋ #97 本文マスク | PUBLIC 履歴露出の実効無効化。手順: runbook §0.5 / `infra/cloudflare/README.md` |
| seed 003_demo 変更後のローカル/STG `db_reset`（SEC / #211 A1+A2） | migration-seed-safety。エージェントは DB reset 自動実行禁止 |
| #211 A6 複合FK の実DB反映 | 2026-07-17 の migration 完全統合により A6 は `001_init.sql`（013 セクション）に内包。既存 DB（ローカル/STG）は 001 checksum 変更のため `DB_RESET=true` での再適用が必要（A1+A2 seed 反映の db_reset と同時に消化可能）。migration適用はエージェント禁止 |
| #211 A6 の DB 適用時 実SQL実行検証（静的SQLレビューのみ実施済み） | B2 の scoped 検証は 2026-07-16 実施済み・green（`docker compose run --no-deps` 経由: go test -run Dose ok／gofmt・build・vet clean） |
| #109 Phase C: `STG_DEMO_*` 登録後に performance-tests フォールバック撤去（エージェント可） | Secrets 未登録のまま撤去すると scheduled が壊れる |
| Vercel Production `VITE_SHOW_DEMO_ACCOUNTS=false` 確認/設定 | 外部システム操作 |
| `terraform apply` 承認（tfvars 準備済み） | インフラ破壊的変更 |
| ADR-003（PO-006）独立 Issue 起票 | PO 承認済。案1B TRIGGER＋二重保持解消検討 |
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

## 画面仕様書全数突合 SD-1〜19＋GAP-1/2 — 全件消化済み（2026-07-17）

> 出所: docs/spec/screens 全数突合（a476b727）＋受け入れシナリオ作成の副産物。2026-07-16 Fable 代理決裁（正本 = `q&a.html`・PO は上書きで覆せる）→ 2026-07-17 ユーザー「fix all」指示で**全 21 件を消化**。実装コミット: 142f5ebe〜6d10f4c0（決裁 ID 単位・全て scoped テスト green）。GitHub 入口 = [#261](https://github.com/MinoruSoga/AnimalEkarte/issues/261)。

**決裁からの乖離 3 件 — 裁定確定（2026-07-17 ユーザー委任・全件承認・revert 不要。裁定理由の正本 = q&a.html 各カード）:**

| ID | 決裁 | 実施内容 | 裁定 |
|----|------|---------|------|
| SD-9 | B: doc のみ | コード実装へ上書き（9b6a01ed） | **承認** — ルール 0 件は自己修復不能（注記では構造的に防げない）。投入内容は 003_demo 運用実証済みプロファイルで無差別全許可ではない |
| SD-13 | 税務ブロックへ追加 | 独立「法人情報」セクション新設（6d0c8d8c） | **承認** — Clinic フォームに Company シングルトンを同居させるスコープ誤認を回避。決裁の本質（FE 導線）は充足 |
| SD-10 | 422 | 400（InvalidInput）のまま | **恒久承認** — 本質は BE 境界拒否で実装済み。422 新設は要件責任者不在のため積まない |

**fix all 実行で発見・対処した追加バグ（決裁対象外の副産物）:**

- **BE finalized ガード欠落 5 件**（clinical_plan/inquiry/UpdateRecommendationReason/estimate/billing_confirmation — GAP-1 調査で発覚・API 直叩きで確定済みカルテに書込可能だった）→ 全て封殺（142f5ebe）。clinical_plan/inquiry のみ既存テスト制約により atomic WHERE 軽量パターン（限界と統一条件は `backend/internal/service/CLAUDE.md` に記録）
- **liff.state 復元前クエリ読取**（LiffLinkPage — 未ログイン初回の OAuth リダイレクト経路で token/clinic_id が必ず欠落）→ isReady 後読取へ修正（2e4808b5）
- **SD-2 の dbOrTx allowlist 未登録回帰**（ac7c9fe8 の取りこぼし）→ 142f5ebe に同梱

**残 USER アクション（SD/GAP 由来）:**

- [ ] **[USER] SD-9 既存院の被害判定 SQL を STG/本番で実行** — ルール 0 件の権限グループの有無を確認（9b6a01ed の修正は**新規作成の院にしか効かない**ため、既存データは手動確認が必要）。0 行なら実害なし・本項クローズ。ヒット時は対象グループへのルールバックフィルを別タスク起票:

  ```sql
  SELECT pg.id, pg.clinic_id, pg.name, COUNT(DISTINCT sp.staff_id) AS assigned_staff
  FROM permission_groups pg
  LEFT JOIN permission_group_rules r ON r.group_id = pg.id AND r.deleted_at IS NULL
  LEFT JOIN staff_permission_groups sp ON sp.group_id = pg.id
  WHERE pg.deleted_at IS NULL
  GROUP BY pg.id, pg.clinic_id, pg.name
  HAVING COUNT(r.id) = 0
  ORDER BY pg.clinic_id, pg.id;
  ```

  `assigned_staff > 0` の行が最優先（実際にそのグループ所属のスタッフが全機能ロックアウト中）。
- [ ] **[USER] SD-14 STG 実機検証** — LINE 紐付け E2E（URL 発行→LIFF 遷移→紐付け完了）
- [ ] **[USER] GAP-2 反映の `db_reset`**（seed 変更 13c6a93a）

**次期送り 6 件は 2026-07-17 に triage 完了 → `phase2.html`「SD/GAP fix all の副産物」節へ移動済み**（本書には残さない — 台帳スコープ規則）。

### AnimalEkarte CSV import — USER actions

> 方針（2026-07-15 確定）: フル 003_demo（~529MB・PHI 含みうる）は **Git に載せない**。正本バックアップは `old_db/sensitive-local/animalekarte-003-demo-full/`。リポジトリの `003_demo` は小さいデモのまま。

- [ ] **USER:** ローカルでフル seed を使う: `rsync -a ../old_db/sensitive-local/animalekarte-003-demo-full/ backend/migrations/seeds/003_demo/` のあと `make reset`（エージェントは reset しない）。誤 `git add` 防止のため該当 CSV に `skip-worktree` 推奨。
- [ ] **USER:** STG へのフル seed 適用は別途承認・手動実行（通常は小さいデモのまま）。
