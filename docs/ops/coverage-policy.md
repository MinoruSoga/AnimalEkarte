# Coverage Policy — AnimalEkarte

> **目的**: テストカバレッジ ratchet 方式の運用ポリシーを定義する(Issue #194)。
> **読者**: CI・カバレッジ改善を行う開発者。
> **タイミング**: カバレッジゲートの閾値・除外設定を変更する時。

> Issue #194 で策定。段階的カバレッジゲート導入のロードマップと除外ポリシーを定義する。

---

## 除外ポリシー

### Backend（Go）

計測対象: `./internal/...` のみ（`-coverpkg` で明示指定）

| 除外パッケージ | 理由 |
|---|---|
| `cmd/*` | エントリポイント。テスト対象のロジックを含まない |
| `lstep-migrate/` | L-ステップ移行ツール。インフラスクリプト |
| `seed-old-db/` | 旧DB シードデータ。インフラスクリプト |
| `migrations/` | SQL マイグレーション。実行可能な Go ロジックなし |

**`internal/model/` について**: struct 定義が主体のため行カバレッジは低くなる。
ただし `internal/...` の一部として計測し、値を歪める特別除外は行わない。
モデルの初期化ロジックや `Validate()` メソッドはカバレッジ対象とする。

### Frontend（TypeScript / Vitest）

計測対象: Vitest デフォルト除外 + 下記を追加除外

| 除外パターン | 理由 |
|---|---|
| `src/types/generated/**` | tygo 自動生成型定義。手書きロジックなし |
| `src/testing/**` | MSW モック・テストセットアップファイル。テスト対象外 |

---

## ベースライン記録

ベースライン数値は CI artifact `backend-coverage` / `frontend-coverage` に自動 publish される。

- **Backend**: `coverage-summary.txt` の最終行（`go tool cover -func` 総計 %）
- **Frontend**: `coverage/coverage-summary.json` の `total` フィールド

**取得方法**: GitHub Actions の当該 workflow run → Artifacts → ダウンロードして確認。

> ベースラインが確定次第、下記テーブルに記録する。

| 計測日 | Backend 総計 % | Frontend Statements % | Notes |
|---|---|---|---|
| 2026-07-01 | 実測値は CI 上に残存せず未確定（下記 07-03 arm 値が最初の正式記録） | — | `closing_settings_service`, `chronic_condition_service`, `shared_file_service`, `validators` 等の各種Service/Middlewareのテストを大幅に拡充しカバレッジを大幅向上（この時点では `.coverage-baseline` への転記・arm はまだ実施していない）。 |
| 2026-07-03 | 89.9%（re-arm 済み、下記 07-13 参照） | — | BE-refactor.md R3-5。GitHub Actions run 28655388836（push, commit 80e0648a）の `backend-coverage` artifact `coverage-summary.txt` 末尾 `total:` 行を `backend/.coverage-baseline` に転記。以降 tolerance（既定 0.5pp）を超える低下は CI を fail させる。 |
| 2026-07-13 | **91.3%**（re-arm 済み） | — | Issue #212。GitHub Actions run 29152374862（push, commit 70f4c298）の `backend-coverage` artifact `coverage-summary.txt` 末尾 `total:` 行を `backend/.coverage-baseline` に転記（89.9→91.3）。当時の旧layer別内訳はBE9前の履歴値であり、移行後packageの現行基準には使用しない。`internal/infra`（line 0% / lstep 52.3% / crypto 75.7%）は除外ポリシー見直しのPO判断待ちのため対象外。 |
| 2026-07-24 | **未再計測（91.3% ratchetを維持）** | — | BE9構造移行後のremote CI coverage artifact待ち。旧`handler/service/repository`別の値を移行後domainへ転記・推測しない。main CIの実測が得られるまでは既存baseline 91.3%と0.5pp toleranceを変更しない。 |
| 2026-07-04 | — | 未記録（0・warn-onlyで起動）→ 07-05 に arm（下記参照） | FE-refactor.md R-F5。`vite.config.ts` の coverage reporter に `json-summary` を追加し `coverage/coverage-summary.json` を生成。`frontend/scripts/coverage-ratchet.mjs` + `frontend/.coverage-baseline` を新設し CI に ratchet ステップを追加（backend と同型）。 |
| 2026-07-05 | — | **43.78%**（arm 済み） | GitHub Actions run 28672433856（push to main, commit 61b85d7a）の `frontend-coverage` artifact `coverage-final.json`（v8 provider）を istanbul json-summary と同じ式で全799ファイル集計（13624 statements 中 5964 covered）し `frontend/.coverage-baseline` に転記。以降 tolerance を超える低下は CI を fail させる。詳細な算出根拠は `frontend/.coverage-baseline` のコメントを参照。 |

---

## 段階的しきい値ロードマップ

### Phase 0（現状 — 非ゲート可視化）

- 状態: **実装済み（#186 / 9b530f08）**
- 動作: CI が `coverage.out` を生成し Job Summary + artifact に publish するが、しきい値なし
- ブランチ対象: main, staging, production への PR / push

### Phase 1（低下 fail ratchet）

- 状態: **fail ゲートとして arm 済み（Backend: BE-refactor.md R3-5・2026-07-03／Frontend: FE-refactor.md R-F5 導入・2026-07-05 baseline 記録で arm 済み）**
- 実装（Backend）: `backend/cmd/coverage-ratchet`（Go ツール・判定ロジックは `main_test.go` で単体検証）が
  `go tool cover -func` 総計 % と `backend/.coverage-baseline` を比較する。CI の
  「Coverage ratchet」ステップは `-warn-only` なし・`continue-on-error` なしで実行し、
  tolerance を超える低下を検出すると非0 exit してステップ自体を fail させる
  （`evaluateRatchet` は baseline ≤ 0 のときのみ warn-only 相当で exit 0 を返す設計）。
- しきい値設定: `tolerance` 既定 0.5pp。baseline は `backend/.coverage-baseline`（91.3%、2026-07-13 に main CI run 29152374862 の artifact で re-arm）。
- ブランチ対象: backend Test job（main への push / staging・production への PR）
- **ベースライン更新手順**: 意図的にカバレッジ基準を変える場合のみ、`backend-coverage` artifact の
  `coverage-summary.txt` 末尾 `total:` 行の実測 % を `.coverage-baseline` に転記する（推測値を書かない）。
- 実装（Frontend / FE-refactor.md R-F5）: `frontend/scripts/coverage-ratchet.mjs`（Node script・
  `parseTotalStatementsPct` / `readBaseline` / `evaluateRatchet` は純粋関数として export・fixture 実行で
  4パターン [OK / FAIL / baseline未記録 / --warn-only] を検証済み）が `coverage/coverage-summary.json`
  の `total.statements.pct` と `frontend/.coverage-baseline` を比較する。CI の「Coverage ratchet」ステップは
  backend と同じく `-warn-only` なしで実行。**baseline は 2026-07-05 に 43.78% で記録・arm 済み** —
  以降 tolerance（既定 0.5pp）を超える低下は CI を fail させる（backend が辿ったのと同じ2段階導入を完了）。
- ブランチ対象: frontend Test job（main への push / staging・production への PR）
- **ベースライン記録手順（Frontend）**: `frontend-coverage` artifact 内の `coverage-summary.json` の
  `total.statements.pct` の実測値を `frontend/.coverage-baseline` に転記する（推測値を書かない）。

### Phase 2（パッチカバレッジ warn）

- 目標: 変更ファイルのパッチカバレッジが 70% 未満のとき warn を出す
- 実装: `octocov` または `diff-cover` によるパッチカバレッジ計算
- しきい値: **warn（fail なし）** — 既存 PR をブロックしない
- ブランチ対象: main への PR
- 発火条件: `backend/internal/<domain>`、`backend/cmd/api`、cross-cutting package、または`frontend/src/features`の変更を含むPR
- 着手条件: Phase 1 が安定稼働し、ベースライン数値が記録されてから

### Phase 3（domain/capability別 fail ゲート）

- 目標: clinical safety、clinic isolation、authentication、billing等のriskが高いdomain/capabilityに、各packageの実測baselineから低下しないratchetを設ける
- 単位: `internal/<domain>`を基本とし、`cmd/api`のcomposition/lifecycleと`audit`・`persistence`等のcross-cutting packageは独立単位で扱う
- しきい値: remote CI artifactで移行後baselineを再計測してから決める。旧`handler/service/repository`の値や一律80%を移植しない
- 実装: coverage profileをpackage単位で集計し、package追加・rename・分割を明示的manifestで追跡する
- ブランチ対象: main, staging への PR
- 発火条件: 対象domain/capabilityまたはそのcompositionを変更するPR
- 着手条件: BE9後のmain CI artifactが取得済みで、package別baselineと除外理由がreview済みであること

---

## 参考リンク

- Issue #194: [FOLLOW-UP] #186 CI: カバレッジ計測ポリシー
- PR #186 / commit 9b530f08: 非ゲートカバレッジ artifact 導入
- `.github/workflows/ci.yml`: Backend/Frontend Test ステップ（`-coverpkg` フラグ、`coverage.exclude` 参照）
- `frontend/vite.config.ts`: Vitest `test.coverage` 設定
