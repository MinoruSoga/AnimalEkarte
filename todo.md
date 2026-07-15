# AnimalEkarte — Unified TODO（todo.md）

> 更新: 2026-07-15
> 前回: 2026-07-07（全面棚卸し）→ 2026-07-15（open Issue/PR・実装済み除去の再棚卸し）→ 2026-07-15（`todo.md` / `BE_todo.md` / `BE-refactor.md` / `FE-refactor.md` を本ファイルへ統合）
> 監査範囲: backend / frontend / CI-CD / migrations / docs / GitHub Issues / git 状態 / docs/tasks/open
> **push・外部書き込み・credential 変更はユーザー所有アクション。**（PR マージはユーザーが手動で行う。本台帳には載せない）
> **別台帳**: BE 保留 = `BE-pending.md` / PO 判断キュー = `q&a.html` / バグ監査 = `docs/archive/bug.md` / 受付テレメトリ完了記録 = `docs/archive/tasks/closed/change-ui.md`
> **本書の役割**: プロジェクト横断 TODO・今期着手可能な BE 残・BE/FE リファクタ次期引き継ぎ・やらない判断の正本台帳。

---

## 運用規約

### Docker 検証規約（BE・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

### 台帳スコープ規則

- 今期着手可能な BE 残タスクのみを「BE 残タスク」節に記載する（シークレット・テスト・PERF 等）。対応済みは残さない。詳細・手順の正本は git 履歴と `docs/tasks/closed/`。
- その他 open FEAT（#249/#247/#239/#238/#237/#235/#234/#232/#230 等）は **gh を正**とし、本台帳には重複掲出しない。
- PR マージ判断・マージ状態・マージ用チェックリストは本台帳に載せない（ユーザー手動）。
- PO 決裁待ちは `q&a.html` を正本とし、実装着手禁止。
- 着手保留・次期送り・任意検証は `BE-pending.md` を正本とする。再検討トリガが立つか判断が出たら、実装単位として本台帳の該当節へ戻す。

---

## Project TODO

### P1 — Open Issues（2026-07-15 時点・台帳掲載分）

| # | 内容 | 現状 |
|---|---|---|
| #211 | 検査・健診パッケージ化 | **歯科＋provisional seed 投入済み**（皮膚・耳統合／眼科／アドプリット構造のみ／季節健診・尿。clinic1=4種・clinic2/3=皮膚・耳＋眼科。全 fields `is_provisional=t`。正規経路 seed-export）。**USER 要: `db_reset` でローカル反映**。残: マスタ CRUD UI（PO 編集頻度確認後）／exam_results 複合 FK（非additive・別タスク）／ライブ E2E。**PO確認**: アドプリット実体・価格（高）／尿比重の犬猫別レンジ（高）／select 非ハイライト許容（中）／季節4分割・腎臓ドック要否（低） |
| #201 | 薬量自動計算 | **実装完了・PO残のみ**（BE・FE・`MedicineDoseParamsEditor`／`TreatmentRow` 自動プリフィル済み）。PO残: 丸め合意・逸脱閾値・admin_route・prescriptions。gh クローズは USER |
| #212 | カバレッジ90%目標 | ratchet ゲート導入で regression 防止は達成。90% 到達自体は長期目標のまま未達 |
| #89/#97/#98/#99/#109 | シークレット移行・ローテーション | **リポジトリ Phase A 完了**（seed/テスト平文除去・runbook §0.5・履歴インベントリ）。**USER BLOCKED**: 4系統ローテーション / P5-2 Secrets / #97 本文マスク / #109 フォールバック撤去。詳細は SEC-SECRETS-5 |

その他 open FEAT（#249/#247/#239/#238/#237/#235/#234/#232/#230 等）は **gh を正**とし、本台帳には重複掲出しない。

**クローズ済みで本表から除去（2026-07-15）**: #213 / #196 / #194 / #189（実装済みまたは決裁済み・gh CLOSED）／#229（飼主レポート危険度平文表示の除去・ローカル実装完了・gh クローズは USER）

### P1 — lab_import 外部検査連携

> PO 確認待ちの FE lab_import UI は `q&a.html`（PO-007）へ移記済み。UX 仕様確定後に本節へ戻す。

### P2 — リファクタリング follow-up

- [ ] **docs/tasks/open の PERF/FOLLOWUP 系** — 未消化（2026-07-15 実在: `PERF-FOLLOWUP-01/02/05/07`・`PERF-M1/M2/M3`・`FOLLOWUP-X14A-*`・`FEAT-searchable-select-targets`）。コード裏付け: `accounting_handler.go:161` の `hasPermission` がハンドラ内に残存（旧 PERF-FOLLOWUP-03 相当・md は closed/削除済み）

> BE リファクタ（第7期）・FE リファクタ（第6期）は状態: **完了**。次期引き継ぎは下記「BE リファクタ引き継ぎ」「FE リファクタ引き継ぎ」節を参照。

### P3 — インフラ / その他

- [ ] **[USER] P2 Terraform（internal ALB + VPC Origin）本番適用** — `infra/terraform/terraform.tfvars` はローカルに準備済み（gitignore 対象）。`terraform apply` の実行判断は USER
- [ ] **[USER 判断] stg-smoke の login/CRUD 復活の要否** — 撤去済み・復活時の手順は workflow 内コメントに明記
- [ ] **[USER] Vercel Production `VITE_SHOW_DEMO_ACCOUNTS=false` 確認** — `frontend-deploy.yml` は Vercel ダッシュボード側の環境変数を使う設計。リポジトリ外での確認が必要・未検証
- ⚠️ **ECS ロールバック経路の `.env.staging` 依存ギャップ** — `.env.staging` untrack 後、`backend-deploy-ecs.yml`（`workflow_dispatch` 専用）を dispatch すると `FileNotFoundError` になる。通常の STG 運用（Cloudflare 正系統）では影響なし。詳細: [ECS ロールバックランブック §2](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)

### PO 決裁待ち（実装着手禁止）

> 本節の項目は `q&a.html`（内部 PO 判断キュー）へ移記済み。正本の決裁記録は同 HTML のカードを参照。

### ユーザー所有アクション一覧

| アクション | 根拠 |
|-----------|------|
| SEC-SECRETS-5: 4系統ローテーション（PlanetScale / Cloudflare / LINE / JWT・暗号化鍵）＋ P5-2 GitHub Secrets 登録＋ #97 本文マスク | PUBLIC 履歴露出の実効無効化。手順: runbook §0.5 / `infra/cloudflare/README.md` |
| SEC-SECRETS-5: seed 003_demo 変更後のローカル/STG `db_reset`（checksum mismatch） | migration-seed-safety。エージェントは DB reset 自動実行禁止 |
| #109 Phase C: `STG_DEMO_*` 登録後に performance-tests フォールバック撤去（エージェント可） | Secrets 未登録のまま撤去すると scheduled が壊れる |
| ECS ロールバック時のみ: SSM Parameter Store 登録＋IAM 権限 | 通常運用では不要（Cloudflare 正系統は `wrangler secret put` が代替） |
| Vercel ダッシュボードの Production 環境変数で `VITE_SHOW_DEMO_ACCOUNTS=false` を確認/設定 | 外部システム操作 |
| `terraform apply` 承認（tfvars 準備済み） | インフラ破壊的変更 |
| ADR-003 論点1 の Issue 起票承認 | 外部 write |

### 証跡サマリー（2026-07-15）

| 検査対象 | 結果 |
|---------|------|
| GitHub Issues（open） | 19件（台帳掲載は上表。FEAT 群は gh 正） |
| git | `main` = `origin/main`（`2fb4959e`） |
| Backend coverage ベースライン | `backend/.coverage-baseline` 存在（#194 CLOSED） |

---

## BE 残タスク

> 今期着手可能な BE 残タスクのみ。対応済みは残さない。詳細・手順の正本は git 履歴と `docs/tasks/closed/`。
> 次期送り・着手保留・任意検証は `BE-pending.md`。本書と重複させない。

**エージェント実装可能な残タスク（2026-07-15 棚卸し）:**

| ID | 優先度 | 内容 | 状態・条件 |
|----|--------|------|-----------|
| SEC-SECRETS-5 | **部分完了（USER 残）** | **エージェント Phase A 完了（2026-07-15）**: `003_demo` LINE シークレット列を空化・識別子を合成値化、テスト/コメントの旧実値除去、`.gitleaks.toml` 更新、runbook §0.5 ローテーション手順復元、gitleaks 全履歴インベントリ（`docs/infra/deploy/runbooks/SEC_SECRETS_5_GITLEAKS_HISTORY_INVENTORY.md`）。**残（USER・credential-impacting）**: 4系統ローテーション、P5-2 `gh secret set`、#97 本文マスク、#109 `STG_DEMO_*` 登録後のフォールバック撤去。#98/#99 は Phase 8 まで PENDING | 作業ツリーに LINE 実平文なし。Issue クローズは USER ローテーション完了後。seed 変更後はローカル/STG で checksum mismatch → `db_reset` が必要（USER） |
| TEST-FLAKE-P2 | **完了（2026-07-15）** | `TestAppointmentTrimmingDetail*` を `setupIsolatedTestDB` 化（共有プール上の並行 TRUNCATE 破壊がフレーク源）。`setupAppointmentTrimmingDetailTestDB` + MasterPreloadClinicIsolation を隔離接続へ。CI 広範 `-parallel 1` は不採用 | 検証: `go test ./internal/repository/ -run 'TestAppointmentTrimmingDetail' -count=5` |

### 見送り（再開条件付き・今期着手しない）

| ID | 内容 | 再開条件 |
|----|------|----------|
| PERF-AUDIT-TX P2（outbox） | outbox パターン移行 | `audit_write_failed` が恒常的（目安: 月 1 件以上継続）に観測された場合、実測頻度 1 か月分を添えて再起案。正本: `docs/tasks/closed/perf/PERF-AUDIT-TX-UNIVERSAL-BEST-EFFORT.md` |

**スコープ外**: `docs/tasks/open/FEAT-searchable-select-targets.md` は FE 案件のため本節に含めない（Project TODO の P2 に掲載）。

**履歴**: #236 skip 解除は 2026-07-14 CLOSED。詳細は git 履歴（`bb2ad499` 等）を参照。

---

## BE リファクタ引き継ぎ

- **第7期**: **完了**（BE7-0〜BE7-21 全22項目、2026-07-14 棚卸しで裏取り済み）。詳細は git 履歴（`565c8708`〜`a6cbdc70`、進捗同期 `5d5600f1`）を参照。
- **本書の役割（本節）**: 次期監査への引き継ぎのみ。新規の第7期作業はない。
- **PO 判断待ち**: `q&a.html`（内部 PO 判断キュー）へ移記済み
- **更新日**: 2026-07-15（台帳整理: 対応済削除・PO 移記）→ 同日 `todo.md` へ統合

### 次期監査への引き継ぎ

- **[要精査] `middleware/liff_auth.go:56,116`**: `LineCustomerRepository.FindOrCreateByLineUserID` を middleware から直接呼んでいる（層違反の疑い）。
- **[次期] apicontract のフィールドレベルゲート**: BE7-17 クラスの api.yaml 乖離を機械検出する拡張。
- **[次期] god-function 走査**: `examination_service.go` ReplaceItems（101行）/ medicine・reservation・cash_register の80行台。
- **[次期] `lstep_csv_import_service.go` が自前 `s.db`（gorm 直 import）を持つ妥当性**。
- **[次期] `reservation_type_handler.go` の weekly/specific 相互依存バリデーションの service 移動**（LOW）。
- **[次期] `AuthService`/`TokenService`（BE7-20/21 で抽出済み）と `validators_auth.go` の統合、P8/P11 準拠の付与**。
- **[次期] `AuditService.Log` の interface 露出整理**。

### 第7期で確定した「やらない」判断（次期でも踏襲推奨）

- duration リテラル（`24*time.Hour` 等）の一括定数化・`INTERVAL '365 days'` SQL 共通化はしない。
- `internal/middleware/response.go` の `respondError` と handler `RespondError` の統一はしない（X-17 で意図的見送り済み）。
- 犬猫種別判定ヘルパ `isDog/isCatSpeciesName`（部分一致・マーケ）と `doseSpeciesAliases`（完全一致・投薬 fail-closed）は**契約が意図的に異なるため統合禁止**。

---

## FE リファクタ引き継ぎ

- **作成日**: 2026-07-13 / **更新日**: 2026-07-15（台帳整理: 対応済削除・PO 移記・Stale 訂正）→ 同日 `todo.md` へ統合
- **第6期**: **完了**（FE6-0〜FE6-18 全19項目）。詳細は git 履歴を参照。
- **本書の役割（本節）**: 次期監査への引き継ぎのみ。新規の第6期作業はない。
- **PO 判断待ち**: `q&a.html`（内部 PO 判断キュー）へ移記済み

### 次期監査への引き継ぎ

- **OwnerSearchModal の React Query 化**: FE6-1 はバグ修正に留めた。`useState`+`useTransition` の素朴 fetch を feature 側 `useSearchOwners` フックに置き換える構造改善は次期。
- **`ShiftFormDialog` の `use-shift-form.ts` 抽出 / `TreatmentRow` の EditableCell 化 / `ChangePasswordDialog` の api 層整理**: いずれも実害なしの一貫性改善。
- **liff / line-reserve の `index.html` に CSP メタタグがない**（メインアプリのみ設定済み）。セキュリティ観点の追加検討。
- **`src/lib/` と `src/utils/` の役割分担が不文律**（両方にフォーマット系が分散）。規約明文化候補。
- **export されているが外部参照のない型シンボル約15件**（`CPMStageOption` 等）: 次期にまとめて掃除。
- **Pet属性ラベルの単一ソース化**: FE6-8 は二重定義＋ガードテストでの乖離検知に留めた。単一ソース化を次期に検討。
- **曜日ラベル契約の統合**: `DAY_OF_WEEK_LABELS`（0=日曜始まり）と `line-reserve` の Calendar（`WEEK_DAYS = ['日',…]`、日曜始まり）は一致。一方 master の `ReservationTypeAvailableSlotsCalendar` は月曜始まりヘッダー（`weekStartsOn: 1`）。契約が異なるため統合には設計が必要。

### 第6期で確定した「やらない」判断（次期でも踏襲推奨）

- `use-*-form` 系フックの共通スケルトン抽象化は、ドメインロジックが実質的に異なり害と判定済み。
- `src/features/owners/components/pet-edit-field-shared.tsx` のリネーム・`.ts` 化は不可（JSX 定数を含む）。
- `src/components/ui/`（shadcn 生成物）・`src/types/generated/`（tygo 生成物）は編集しない。
- `types/index.ts` の FA9 構造自体の変更はしない（FE6-18 でドキュメント明文化のみ実施済み）。

---

## 別台帳ポインタ

| 台帳 | 役割 |
|------|------|
| `BE-pending.md` | 着手保留・次期送り・任意検証の正本（現在「次期送り」該当なし） |
| `q&a.html` | 内部 PO 判断キュー（実装着手禁止項目の正本） |
| `docs/archive/bug.md` | バグ監査アーカイブ |
| `docs/archive/tasks/closed/change-ui.md` | 受付テレメトリ完了記録 |
| `docs/tasks/open/` | PERF/FOLLOWUP 等の個別タスクファイル |
| `docs/tasks/closed/` | 対応済み詳細・手順の正本 |

> 旧 `todo.md` / `BE_todo.md` / `BE-refactor.md` / `FE-refactor.md` は本ファイルへ吸収済み（削除）。
