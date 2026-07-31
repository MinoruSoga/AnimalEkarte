# Remaining work ledger (open only)

オープン residual のみを列挙する。対応済み TASK / closed 索引行は **削除済み**（2026-07-31）。  
根拠・完了証拠は git 履歴と `reports/2026-07-31-*.md` を参照。

> **ID namespace**: 本ファイルの `TASK-*` はローカル連番。`3-session-agent.html#ledger` 体系外。`/implement` は正本 ledger からのみ解決。

## 索引 / サマリー

| Inv | 内容 | 処置 |
|-----|------|------|
| R2 | 編集モード治療プランが read-only のまま | **TASK-002**（案B 決裁済・unlock WONTFIX / UI follow-up） |
| R3 | フォーム一括割引が非永続 | **TASK-003**（案B 決裁済・永続化 WONTFIX / UI follow-up） |
| R4 | screens-drift 意図変更のコミット隔離 | **TASK-004**（ops 手順・land は USER） |
| R5 | コミット前の closed pack 回帰ゲート | **TASK-005**（ops 手順・land 前再実行） |
| R6 | マルチエージェント共有 tree thrash | **ops-only** |
| R7 | empty-diff 成功宣言 harness | **ops-only** |
| SCEN-SEED-001 | 003_demo clinical CSV ヘッダのみ | **TASK-009**（設計済・**適用は USER**） |
| SCEN-BROWSER-001 | scenarios 内【要実測】backlog | **TASK-010**（BLOCKED env） |
| SCEN-PROD-ESTIMATE-001 | estimate unlock vs 26§2.1 | **TASK-012**（案B 決裁済・High follow-up） |
| SCEN-PROD-CLOSING-001 | closing reverse boundaries | **TASK-013**（案B 決裁済・High follow-up） |
| SCEN-PROD-PAYMENT-KEY-001 | payment system_key | **TASK-014**（案A 決裁済・実装 follow-up） |
| SCEN-OPS-CLAIM-001 | `claim/*` 解放 | **ops-only**（USER only） |
| SCEN-OPS-COMMIT-001 | mixed commit 説明メモ | **ops-only**（rewrite しない） |
| SCEN-OPS-TREE-001 | 共有 tree concurrent WIP | **ops-only**（= R6） |
| ARCH-R2 | empty-diff COMPLETE 規律 | **ops-only**（継続） |
| ARCH-R3 | land 時 foreign 定義は `git status` 実測 | **ops note** + TASK-004 |
| POST-PULL | migrations 適用 | **ops-only** ≡ **SPEC-TOP-MIGRATE-006**（USER `make migrate`） |
| SPEC-TOP-LINE-AUDIT | `docs/spec/line/**` deep 監査 | **TASK-019**（partial / deep follow-up） |
| SPEC-TOP-E2E-RUNTIME-84 | Playwright runtime 84 | **TASK-020**（BLOCKED env） |
| SPEC-TOP-CAPABILITIES-CRUD | capabilities dual surface | **TASK-021**（段階 B→A 決裁済・実装 follow-up） |
| SPEC-TOP-CLAIM-RELEASE | claim 解放 | **SCEN-OPS-CLAIM-001** |

### 対応済み（削除済み・再掲しない）

TASK-001-BE/FE, TASK-006/007/008, TASK-011, TASK-015/016/017, TASK-018（監査記録1回）, ARCH-DONE, SPEC-TOP-G1-G12, SPEC-TOP-FOOTER-115, SPEC-TOP-CAP-SOT-DOC（doc 固定）, SPEC-TOP-AVAILABLE-STAFFS（WONTFILE）, R8-\*, SCEN-S11-COPY-001, SCEN-AUDIT-MED-001, ARCH-R1。

### Ops-only notes（製品コード TASK にしない）

- **R6 / SCEN-OPS-TREE-001**: 並行エージェントは worktree 隔離。共有 tree は 1 編集セッションのみ。
- **R7 / ARCH-R2**: 受け入れは `git diff` / `git status` の実 diff 必須。empty-diff COMPLETE 禁止。
- **ARCH-R3 / TASK-004**: land 直前の `git status --porcelain` で intentional / foreign を定義。台帳に dirty 一覧を書かない。
- **POST-PULL / SPEC-TOP-MIGRATE-006**: USER が `make migrate`。エージェントは auto-apply しない。
- **SCEN-OPS-CLAIM-001**: claim 解放は USER only。未統合 claim は削除前に USER 判定。
- **SCEN-OPS-COMMIT-001**: mixed history の説明用。history rewrite / force-push しない。

### 推奨実装順（open のみ）

1. **ops**: claim 解放（USER）/ `make migrate`（USER）/ 共有 tree 隔離維持
2. **TASK-009** seed 適用（USER。設計: `reports/2026-07-31-task-009-seed-design.md`）
3. **TASK-012 / TASK-013**（PO 決裁済・High follow-up）
4. **TASK-014 / TASK-021**（PO 決裁済・実装 follow-up）
5. **TASK-002 / TASK-003**（PO 決裁済・unlock/永続化 WONTFIX、UI follow-up）
6. **TASK-010** browser 要実測（env 後。seed 後が理想）
7. **TASK-020** Playwright runtime 84（env 後）
8. **TASK-019** line deep audit（任意）
9. **TASK-004 / TASK-005**: 次の intentional land 時に path-scoped commit + 回帰ゲート

---

## 個別タスク詳細

### TASK-002: 編集モード治療プランの更新アンロック（Medium・PO決裁済）

- **問題**: 編集 UI は治療プラン参照のみ。BE には PATCH/DELETE があるが FE クライアント未配線。
- **根拠**: `HospitalizationForm.tsx` `readOnly={isEdit}`; edit save は親のみ; BE routes に PATCH/DELETE; FE write は create POST のみ（create は nested 化済み・TASK-001 done）。
- **修正方針**: **案B 決裁済**。登録時 treatment plan は immutable snapshot。編集 UI は RO を維持し、hospitalization 配下 PATCH/DELETE も consumer inventory 後に撤去（撤去までは service reject）。care plan との用語を分離する。
- **受け入れ条件**: 編集 UI は意図的 RO。hospitalization 配下 PATCH/DELETE は service で拒否され、consumer inventory 後に route 撤去。treatment plan / care plan の用語分離と、子明細ありの親削除 UI guard が固定される。
- **状態**: **案B 決裁済 — unlock WONTFIX / UI follow-up 要**。治療プランは登録時 snapshot として編集時 RO。care plan との用語分離と、子明細ありの親削除を事後 Conflict に依存させない UI guard を follow-up とする。決裁: `reports/2026-07-31-todo-po-decisions-FINAL.md`。

### TASK-003: 入院フォーム一括割引の永続化可否（Medium・PO決裁済）

- **問題**: 一括割引（%/円）は常時 disabled・非永続で、保存可能な値に見える。
- **根拠**: CostSummary `readOnly` + honesty; model に global discount なし。
- **修正方針**: **案B 決裁済**。hospitalization に割引 SoT を追加しない。disabled 入力を削除し、read-only 概算と非保存 honesty を残す。
- **受け入れ条件**: hospitalization の schema/request に一括割引を追加しない。常時 disabled の `%` / `円` 欄を削除し、read-only 概算・非保存 honesty・既存の行割引/会計 SoT を維持する。
- **状態**: **案B 決裁済 — 永続化 WONTFIX / UI follow-up 要**。一括割引の第二 SoT は作らない。常時 disabled の `%` / `円` 欄を削除し、read-only 概算と honesty を残す follow-up。決裁: `reports/2026-07-31-todo-po-decisions-FINAL.md`。

### TASK-004: screens-drift 意図変更セットのコミット隔離（Medium・ops）

- **問題**: intentional と foreign を同一 commit に混ぜない。foreign 定義は land 直前の `git status` / `git diff` が正本。
- **修正方針**: land 直前に porcelain 実測 → path-scoped `git add`（`git add -A` 禁止）。foreign は触らない・捨てない。
- **受け入れ条件**: staged ⊆ intentional; foreign 非 stage; 破棄しない。
- **状態**: **ops 手順 open**（再発・次 land 用）。前回実測: `reports/2026-07-31-task-004-005-land-proc.md`。commit は USER。

### TASK-005: closed packs 回帰のコミット前検証ゲート（Medium・ops）

- **問題**: land 前に doc/code 整合と inventory / hospitalization を機械確認する手順。
- **修正方針**: land 直前: `bash scripts/check-docs-symbol-drift.sh`; scoped hospitalization / route-inventory tests。結果は reports に記録（台帳に point-in-time を書かない）。
- **受け入れ条件**: ゲート PASS; inventory 84 維持; hospitalization unit PASS。
- **状態**: **ops 手順 open**（land 都度）。前回: symbol-drift OK・hosp unit PASS（reports 参照）。

### TASK-009: 003_demo clinical CSV ヘッダのみ — seed 再投入（High）

- **問題**: clinical CSV がヘッダのみでシナリオ前提データが揃わない可能性。
- **根拠**: SCEN-SEED-001。設計は完了、**適用未実施**。
- **修正方針**: 設計 `reports/2026-07-31-task-009-seed-design.md` に従い USER が seed 適用。エージェントは migrate/seed auto-apply しない。
- **受け入れ条件**: 対象 CSV がヘッダのみでなくなる; シナリオ前提を満たす; 適用手順が1箇所で辿れる; 適用は USER。
- **状態**: **設計 done / 適用は USER**。

### TASK-010: scenarios【要実測】一括実測バックログ（Medium）

- **問題**: scenarios に【要実測】残存。初回実測→期待結果昇格が未完。
- **修正方針**: browser-test レーンで実測。記録は `reports/`（シナリオ本文に書かない）。
- **受け入れ条件**: 要実測 0 または PO/BUG 振分; reports に実行記録。
- **状態**: **BLOCKED（env）**。`reports/2026-07-31-task-010-020-runtime-blocked.md`。【要実測】再カウント要。

### TASK-012: 見積 unlock と 26§2.1 の仕様整合（High・PO決裁済）

- **問題**: terminal status の訂正経路が未実装で、S07 が 26§2.1 と矛盾する。
- **修正方針**: **案B 決裁済**。`approved` / `rejected` は不可逆。原本を保持する後継 draft、atomic 採番、supersedes、理由・actor audit を同一 transaction で fail-closed に実装し、S07 を追随する。
- **状態**: **案B 決裁済 — unlock WONTFIX / High 実装・docs follow-up 要**。terminal status は不可逆。atomic `estimate_no` 採番、locked 原本を保持する後継 draft・supersedes・audit 経路、S07 honesty を follow-up とする。決裁: `reports/2026-07-31-todo-po-decisions-FINAL.md`。

### TASK-013: 締め reverse 境界の仕様確定（High・PO決裁済）

- **問題**: public API は append-only だが soft-delete/同期間再 close を DB が物理的に防がず、manual に存在しない reverse 手順が残る。
- **修正方針**: **案B 決裁済**。reverse なし。close は immutable とし、訂正は元 close に紐づく append-only adjustment へ会計差額と実現現金移動を分離記録する。DB/application hardening と manual を追随する。
- **状態**: **案B 決裁済 — reverse WONTFIX / High hardening・docs follow-up 要**。close は append-only。soft-delete/同期間再 close を防ぐ persistence invariant と、存在しない reverse を案内する manual の修正が必要。決裁: `reports/2026-07-31-todo-po-decisions-FINAL.md`。

### TASK-014: payment system_key の仕様・運用方針（Medium・PO決裁済）

- **問題**: UI は `system_key` 非公開だが、DB immutability と system row の deactivate/delete guard が未実装。
- **修正方針**: **案A 決裁済**。予約済み四 key は immutable・編集 UI 非公開。name のみ表示用に変更可。system row は deactivate/delete 不可として FE/BE/DB と V04 を追随する。
- **状態**: **案A 決裁済 — 実装・docs follow-up 要**。予約済み四 key は immutable・編集 UI 非公開。name は表示用に変更可、system row は deactivate/delete 不可として guard と V04 を追随。決裁: `reports/2026-07-31-todo-po-decisions-FINAL.md`。

### TASK-019: docs/spec/line/** deep 監査 follow-up（Medium / 任意）

- **問題**: line 仕様 vs 実装の deep 突合が partial のまま。
- **根拠**: 初回記録 `reports/2026-07-31-task-019-line-audit.md`。
- **修正方針**: deep pass で差分を docs/BUG/要PO/ops に振分。秘密・本番 webhook 操作は対象外。
- **受け入れ条件**: deep 結果1回記録; 新規 open は ID 付きまたは残差なし。
- **状態**: **open（partial → deep follow-up）**。

### TASK-020: ui-design-compliance Playwright 再 runtime（84）（Low / 任意）

- **問題**: inventory 84 静的更新後の full runtime 未実施。
- **修正方針**: env 可なら `ui-design-compliance-readonly.spec.ts` workers=1。結果を reports へ。
- **状態**: **BLOCKED（env）**。`reports/2026-07-31-task-010-020-runtime-blocked.md`。

### TASK-021: staff 対応可能種別 dual surface（Medium・PO決裁済）

- **問題**: capabilities SoT と exclusion 面の dual residual。available-staffs は導入しない（WONTFILE）。
- **修正方針**: **段階 B→A 決裁済**。Stage B で exclusion を capabilities-only の期限付き facade にし、production の exclusion write をゼロ化。Stage A は consumer inventory 後に撤去する。新 endpoint は追加しない。
- **状態**: **段階 B→A 決裁済 — 実装 follow-up 要**。Stage B で exclusion を capabilities-only の期限付き facade にし、Stage A で consumer inventory 後に撤去する。`available-staffs` は WONTFILE。決裁: `reports/2026-07-31-todo-po-decisions-FINAL.md`。

---
