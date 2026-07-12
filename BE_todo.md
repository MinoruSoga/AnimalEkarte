# BE Todo — バックエンド残タスク台帳（2026-07-12 棚卸し／同日 grill-me DAG 実行で 6 件完了・1 件新規発見）

> **作成方法**: `docs/tasks/open/` の全 BE タスク + `backend/*.md` を現行コードと実査で突き合わせた結果。
> 対応済 7 件は `docs/tasks/closed/` へ、`backend/FK_DEPENDENCY_CHECK_ROADMAP.md`（全 11 エンティティ実装完了）は `docs/archive/` へ移動済み。
> 詳細仕様は各 `docs/tasks/open/` ファイルを正本とする（本ファイルは索引 + 検証済み現状サマリ）。
> **2026-07-12 追記**: grill-me 回答（Q1-C/Q2-A/Q3-C/Q4-C/Q5-A）に基づき U1/U2/U3(M1+M2)/U4/U6 を実装完了、U7(SEED-004) は判断文整理のみ実施。AUDIT-TX（旧 #7）は明示的にスコープ外のまま。PERF-FOLLOWUP-07（N+1 lint）の初回実行で新規 N+1 を 2 件発見し、PERF-FOLLOWUP-08 として切り出し・allowlist 追跡済み。

---

## 残タスク一覧（優先度順）

| # | ID | 内容 | 優先度 | 種別 | 状態 |
|---|----|------|--------|------|------|
| 7 | PERF-AUDIT-TX（残部） | 監査ログ欠落の可観測化（P1）+ outbox 設計判断（P2） | Medium | 監査 | 未着手（本サイクルは明示的スコープ外） |
| 9 | SEED-004 | treatments#107 専用 procedure 再付け替え | P3 / **PO 判断待ち** | seed | 判断文整理済（コード差分なし） |
| 10 | PERF-FOLLOWUP-08 | SyncLTVTopPercent / buildStaffSlotInputs の N+1 解消 | Medium / Low-Medium | 性能 | 新規発見・未着手（lint allowlist で追跡中） |

---

## 今回完了した項目（2026-07-12・grill-me DAG 実行）

| # | ID | 内容 | 検証 |
|---|----|------|------|
| 1 | PERF-FOLLOWUP-05 | pwreset メール送信ゴルーチンに `sync.WaitGroup`、`main.go` graceful shutdown で `wg.Wait()` drain | `go test ./internal/service/... -run PasswordReset` PASS |
| 2 | PERF-FOLLOWUP-01 | `pets(clinic_id, owner_id, deceased_at) WHERE deleted_at IS NULL` 複合インデックスを `migrations/003_add_pets_batch_living_count_index.sql` に追加（**未適用**、ユーザー手動適用待ち） | ファイルレビュー、apply コマンド未実行 |
| 3 | PERF-M1 | `SyncAnnual4CheckupTag` の閾値・タグ mapping hoist（`*WithMappings` パターン） | `perf_n1_regression_test.go` TestPERFM1M2 PASS |
| 4 | PERF-M2 | `SyncFilariaTag` / `SyncFleaTickTag` / `SyncFoodPurchaseTag` の hoist（M1 と同一コミット） | 同上、go-reviewer Approve（CRITICAL/HIGH なし） |
| 5 | PERF-FOLLOWUP-02 | `FindAllWithLineUserIDCursor` / `FindDormantOwnerEntriesCursor` 新設、バッチループをカーソルページング化（500件/ページ、旧メソッドは他呼び出し元存在のため保持） | 境界値（500/501件）テスト PASS |
| 6 | PERF-FOLLOWUP-07 | `internal/service/n1_lint_test.go` 新設（`go:embed` のパッケージ内制約により配置先を `repository/` から `service/` へ変更）。初回実行で新規 N+1 を 2 件検出 → #10 として切り出し allowlist 追跡（`docs/tasks/open/PERF-FOLLOWUP-08-ltv-staffslots-n1.md` 参照）した上で GREEN化 | `go test ./internal/service/... -run N1` 全 PASS |
| 8 | PERF-M3 | `CreateCheckupSync` 空 OwnerIDs 早期 return 前に監査ログ記録（選択肢 A） | 専用テスト PASS |

- 全項目 `gofmt -l` クリーン、`go vet` クリーン。AUDIT-TX 関連ファイルへの差分ゼロ（`git status` で確認済み）。
- コミットはユニット単位（下記コミットログ参照）。push は実施していない。

## 7. PERF-AUDIT-TX 残部: 監査の可観測化 + outbox 判断 【Medium・本サイクルは明示スコープ外】

- **完了済（P0）**: `CreateTx`（`audit_repository.go:42-47`、dbOrTx 参加）+ 臨床結果 hard-delete 2 サイトの tx 内監査移行 + `audit_tx_inventory_lint_test.go` による inventory 強制（tx-external 残 0 件）。
- **残作業**:
  - **P1**: tx 外監査（従来 `Create` 経路）の書込失敗 `slog.WarnContext` を運用アラートに昇格させる（選択肢 C）。
  - **P2**: outbox パターン移行の設計判断（選択肢 B）。`audit_outbox` migration は存在しない。**やらないならやらないと決めて本項をクローズする**こと（宙吊りが最悪）。
- 正本: `docs/tasks/open/PERF-AUDIT-TX-UNIVERSAL-BEST-EFFORT.md`

## 9. SEED-004: treatments#107 procedure 再付け替え 【P3・PO 判断待ち】

- clinic1 procedures への「静脈確保・採血」専用項目追加要否を PO に確認するまで着手不可。コード作業なし。
- **2026-07-12 訂正**: 正本内の旧ファイル参照（`003_seed_demo.sql`、削除済み stub）を実体である `backend/migrations/seeds/003_demo/treatments.csv` に修正し、PO へ転送可能な一文の判断依頼を追加済み。
- 正本: `docs/tasks/pending/SEED-004-treatment-107-procedure-followup-2026-06-04.md`

## 10. PERF-FOLLOWUP-08: SyncLTVTopPercent / buildStaffSlotInputs の N+1 解消 【Medium / Low-Medium・新規】

- **発見経緯**: PERF-FOLLOWUP-07 の N+1 lint 初回実行で検出。M1/M2/FOLLOWUP-01/02/05/M3/AUDIT-TX いずれのスコープにも含まれない独立した既存デバッグ。
- `SyncLTVTopPercent`（`lstep_tag_sync_visit_ltv.go:64`）: `tagCacheRepo.FindByOwner` が owner ごとに再取得（M1/M2 と同型、owner 数スケール）。
- `buildStaffSlotInputs`（`liff_service_availability_slots.go:27,33`）: `scheduleRepo.FindAllByDate` / `FindAllBreaksByEntryID` が staff ごとに再取得（staff 数スケール、優先度は相対的に低い）。
- **現状**: `n1_lint_test.go` の `n1Allowlist` に本タスク参照付きで一時許可登録済み（GREEN化のため）。修正時は allowlist 2 エントリと `TestN1Lint_AllowlistEntriesAreLive` のピン留めを削除すること。
- 正本: `docs/tasks/open/PERF-FOLLOWUP-08-ltv-staffslots-n1.md`

---

## 過去にクローズした項目（2026-07-12 棚卸し・実査根拠付き）

| ID | 判定 | 移動先 |
|----|------|--------|
| PERF-EXAM | N+1 不存在（exam 系はループ外 hoist 済み） | `docs/tasks/closed/perf/` |
| PERF-FOLLOWUP-03 | 設計判断で決着（認可= handler / 不変条件= service、コメント明文化済） | `docs/tasks/closed/perf/` |
| PERF-FOLLOWUP-04 | P14 違反残 0 件（handler は svc のみ保持） | `docs/tasks/closed/perf/` |
| PERF-FOLLOWUP-06 | fail-closed 実装済（`reservation_validators.go:329-332`） | `docs/tasks/closed/perf/` |
| PERF-H2 | api.yaml 500 定義 + FE トースト実装済 | `docs/tasks/closed/perf/` |
| STG-DEPLOY-READINESS-2026-05-31 | 全残項目解消（dead code 削除済 / CI 4 ゲート稼働） | `docs/tasks/closed/` |
| STG-COST-OPTIMIZATION-PLAN-2026-06-01 | terraform default で適用済 + Cloudflare 移行で superseded | `docs/tasks/closed/` |
| FK_DEPENDENCY_CHECK_ROADMAP | 全 11 エンティティ実装完了・パターン標準化済 | `docs/archive/` |

**スコープ外**: `docs/tasks/open/FEAT-searchable-select-targets.md` は FE 案件（Combobox 化、実装完了・目視確認待ち）のため本台帳に含めない。
