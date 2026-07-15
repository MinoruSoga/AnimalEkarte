# AnimalEkarte — Unified TODO（todo.md）

> 更新: 2026-07-15（対応済除去: PO-003・TEST-FLAKE-P2）
> 前回: 2026-07-15（PO 決裁反映）→ 同日（台帳統合・PRODUCT_PHILOSOPHY 前提）
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
- PO 決裁の正本は `q&a.html`。2026-07-15 時点で PO-001〜008 は回答済み。決裁済みの「やらない」は実装着手禁止のまま。決裁済みの「即実装可」は本台帳の「PO 決裁」節を正とする。
- 着手保留・次期送り・任意検証は `BE-pending.md` を正本とする。再検討トリガが立つか判断が出たら、実装単位として本台帳の該当節へ戻す。

---

## Project TODO

### P1 — Open Issues（2026-07-15 時点・台帳掲載分）

| # | 内容 | 現状 |
|---|---|---|
| #211 | 検査・健診パッケージ化 | **PO 決裁済**。**即実装**: (1) アドプリット（checkup_types id=15＋配下4フィールド）削除 (2) 尿比重 min/max 空化 — 同一 seed スライス。`db_reset` は USER。**残実装**: exam_results 複合 FK（新規 migration・今四半期）。**やらない**: CRUD UI／四季分割・腎臓ドック／select 異常ハイライト（#249 同時）／ライブ E2E。provisional 解除はクライアント臨床責任者確認後に seed 手動更新 |
| #201 | 薬量自動計算 | **実装・PO 運用確定済**。**エージェント残**: コード注記/audit metadata「運用確定待ち」→「暫定確定(2026-07-15 PO)」。**gh クローズは USER**（B5 文言）。admin_route・prescriptions 連携はスコープ外確定 |
| #212 | カバレッジ90%目標 | ratchet ゲート導入済。90% 到達自体は長期目標のまま未達 |
| #89/#97/#98/#99/#109 | シークレット移行・ローテーション | **USER BLOCKED**（リポジトリ Phase A 済）。4系統ローテ / P5-2 Secrets / #97 本文マスク / #109 フォールバック撤去。詳細は SEC-SECRETS-5 |

その他 open FEAT（#249/#247/#239/#238/#237/#235/#234/#232/#230 等）は **gh を正**とし、本台帳には重複掲出しない。

**クローズ済みで本表から除去**: #213 / #196 / #194 / #189 / #229（gh クローズは USER）／PO-003（受付テレメトリ常時有効固定・2026-07-15）／TEST-FLAKE-P2（隔離 DB 化・2026-07-15）

### P1 — lab_import 外部検査連携

> **PO-007: 今は作らない**（FE 着手禁止）。実データ経路開通後に最小 UX 定義してから実装可。正本: `q&a.html` PO-007。

### P2 — リファクタリング follow-up

- [ ] **docs/tasks/open の PERF/FOLLOWUP 系** — 未消化（`PERF-FOLLOWUP-01/02/05/07`・`PERF-M1/M2/M3`・`FOLLOWUP-X14A-*`・`FEAT-searchable-select-targets`）。コード裏付け: `accounting_handler.go:161` の `hasPermission` がハンドラ内に残存（旧 PERF-FOLLOWUP-03 相当）

> BE リファクタ（第7期）・FE リファクタ（第6期）は **完了**。次期引き継ぎは下記節を参照。

### P3 — インフラ / その他

- [ ] **[USER] P2 Terraform（internal ALB + VPC Origin）本番適用** — `infra/terraform/terraform.tfvars` はローカルに準備済み（gitignore 対象）。`terraform apply` の実行判断は USER
- [ ] **[USER 判断] stg-smoke の login/CRUD 復活の要否** — 撤去済み・復活時の手順は workflow 内コメントに明記
- [ ] **[USER] Vercel Production `VITE_SHOW_DEMO_ACCOUNTS=false` 確認** — リポジトリ外での確認が必要・未検証
- ⚠️ **ECS ロールバック経路の `.env.staging` 依存ギャップ** — 通常の STG 運用（Cloudflare 正系統）では影響なし。詳細: [ECS ロールバックランブック §2](docs/infra/deploy/runbooks/BUG_MD_EXTERNAL_OPS_PENDING_APPROVAL.md)

### PO 決裁 — 即実装可（未消化）

> `q&a.html` PO-001〜008 は回答済み。正本は同 HTML。対応済みは本表から除去。

| 優先 | ID | 内容 | 備考 |
|------|-----|------|------|
| 1 | #211 A1+A2 | アドプリット seed 削除＋尿比重 min/max 空化 | 同一 seed スライス。`db_reset` は USER |
| 2 | #201 B2 | 「運用確定待ち」注記 →「暫定確定(2026-07-15 PO)」 | コード/監査メタ |
| 3 | #211 A6 | exam_results 複合 FK の新規 migration 起草 | 既適用 001 編集禁止。STG 適用は USER |
| 4 | PO-008 | 7-1/7-3 実装現行値ファクトシート1枚 | クライアント確認用。7-2/7-4 はブロッカー |
| — | PO-002 | Sentry Phase 1（例外+版数のみ・PII off） | ベンダ/課金は USER。security-review 必須 |
| — | PO-006 | ADR-003 TRIGGER Issue | **起票操作は USER**（案1B・二重保持解消もスコープに含める） |

**やらない（決裁確定）**: 健診 CRUD UI / 四季分割・腎臓ドック / select 異常ハイライト（今） / admin_route・prescriptions dose 連携 / 会計 DTO 化 / lab_import FE / #211 ライブ E2E（今） / Sentry Phase 2（実測後）

### ユーザー所有アクション一覧

| アクション | 根拠 |
|-----------|------|
| SEC-SECRETS-5: 4系統ローテーション＋ P5-2 GitHub Secrets 登録＋ #97 本文マスク | PUBLIC 履歴露出の実効無効化。手順: runbook §0.5 / `infra/cloudflare/README.md` |
| seed 003_demo 変更後のローカル/STG `db_reset`（SEC / #211 A1+A2） | migration-seed-safety。エージェントは DB reset 自動実行禁止 |
| #109 Phase C: `STG_DEMO_*` 登録後に performance-tests フォールバック撤去（エージェント可） | Secrets 未登録のまま撤去すると scheduled が壊れる |
| ECS ロールバック時のみ: SSM Parameter Store 登録＋IAM 権限 | 通常運用では不要 |
| Vercel Production `VITE_SHOW_DEMO_ACCOUNTS=false` 確認/設定 | 外部システム操作 |
| `terraform apply` 承認（tfvars 準備済み） | インフラ破壊的変更 |
| ADR-003（PO-006）独立 Issue 起票 | PO 承認済。案1B TRIGGER＋二重保持解消検討 |
| #201 gh Issue クローズ | PO 運用確定済。残記録文言は B5 |
| #229 gh Issue クローズ | ローカル実装完了済み |
| Sentry 等ベンダ確定・課金契約（PO-002） | 課金・外部契約 |

### 証跡サマリー（2026-07-15）

| 検査対象 | 結果 |
|---------|------|
| GitHub Issues（open） | 台帳掲載は上表。FEAT 群は gh 正 |
| Backend coverage ベースライン | `backend/.coverage-baseline` 存在（#194 CLOSED） |

---

## BE 残タスク

> 今期着手可能な BE 残タスクのみ。対応済みは残さない。詳細・手順の正本は git 履歴と `docs/tasks/closed/`。
> 次期送り・着手保留・任意検証は `BE-pending.md`。本書と重複させない。

**エージェント実装可能な残タスク:**

| ID | 優先度 | 内容 | 状態・条件 |
|----|--------|------|-----------|
| SEC-SECRETS-5 | **USER 残** | リポジトリ Phase A 済。**残（credential-impacting）**: 4系統ローテ、P5-2 `gh secret set`、#97 本文マスク、#109 `STG_DEMO_*` 登録後のフォールバック撤去。#98/#99 は Phase 8 まで PENDING | Issue クローズはローテ完了後。seed 変更後は `db_reset`（USER） |

### 見送り（再開条件付き・今期着手しない）

| ID | 内容 | 再開条件 |
|----|------|----------|
| PERF-AUDIT-TX P2（outbox） | outbox パターン移行 | `audit_write_failed` が恒常的（目安: 月 1 件以上継続）に観測された場合、実測頻度 1 か月分を添えて再起案。正本: `docs/tasks/closed/perf/PERF-AUDIT-TX-UNIVERSAL-BEST-EFFORT.md` |

**スコープ外**: `docs/tasks/open/FEAT-searchable-select-targets.md` は FE 案件のため本節に含めない（Project TODO の P2 に掲載）。

---

## BE リファクタ引き継ぎ

- **第7期**: **完了**（BE7-0〜BE7-21）。詳細は git 履歴を参照。
- **本書の役割（本節）**: 次期監査への引き継ぎのみ。
- **PO 決裁**: `q&a.html`（BE 関連の「やらない」は PO-004/005/006）。

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

- **第6期**: **完了**（FE6-0〜FE6-18）。詳細は git 履歴を参照。
- **本書の役割（本節）**: 次期監査への引き継ぎのみ。
- **PO 決裁**: `q&a.html`。残 FE 即実装は PO-002（Sentry Phase 1）。

### 次期監査への引き継ぎ

- **OwnerSearchModal の React Query 化**: FE6-1 はバグ修正に留めた。feature 側 `useSearchOwners` フックへの構造改善は次期。
- **`ShiftFormDialog` の `use-shift-form.ts` 抽出 / `TreatmentRow` の EditableCell 化 / `ChangePasswordDialog` の api 層整理**: 実害なしの一貫性改善。
- **liff / line-reserve の `index.html` に CSP メタタグがない**（メインアプリのみ設定済み）。
- **`src/lib/` と `src/utils/` の役割分担が不文律**。規約明文化候補。
- **export されているが外部参照のない型シンボル約15件**（`CPMStageOption` 等）: 次期にまとめて掃除。
- **Pet属性ラベルの単一ソース化**: FE6-8 は二重定義＋ガードテストでの乖離検知に留めた。
- **曜日ラベル契約の統合**: master の `ReservationTypeAvailableSlotsCalendar` は月曜始まりヘッダー。契約が異なるため統合には設計が必要。

### 第6期で確定した「やらない」判断（次期でも踏襲推奨）

- `use-*-form` 系フックの共通スケルトン抽象化は、ドメインロジックが実質的に異なり害と判定済み。
- `src/features/owners/components/pet-edit-field-shared.tsx` のリネーム・`.ts` 化は不可（JSX 定数を含む）。
- `src/components/ui/`（shadcn 生成物）・`src/types/generated/`（tygo 生成物）は編集しない。
- `types/index.ts` の FA9 構造自体の変更はしない（FE6-18 でドキュメント明文化のみ実施済み）。

---

## 別台帳ポインタ

| 台帳 | 役割 |
|------|------|
| `BE-pending.md` | 着手保留・次期送り・任意検証の正本 |
| `q&a.html` | 内部 PO 判断キュー（決裁記録の正本。PO-001〜008 回答済み） |
| `docs/archive/bug.md` | バグ監査アーカイブ |
| `docs/archive/tasks/closed/change-ui.md` | 受付テレメトリ完了記録 |
| `docs/tasks/open/` | PERF/FOLLOWUP 等の個別タスクファイル |
| `docs/tasks/closed/` | 対応済み詳細・手順の正本 |

> 旧 `todo.md` / `BE_todo.md` / `BE-refactor.md` / `FE-refactor.md` は本ファイルへ吸収済み（削除）。
