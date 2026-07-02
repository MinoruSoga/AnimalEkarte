# Coverage Policy — AnimalEkarte

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
| `stage-import/` | ステージング用インポートツール。インフラスクリプト |
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
| 2026-07-01 | (CIにて測定予定) | — | `closing_settings_service`, `chronic_condition_service`, `shared_file_service`, `validators` 等の各種Service/Middlewareのテストを大幅に拡充しカバレッジを大幅向上。 |

---

## 段階的しきい値ロードマップ

### Phase 0（現状 — 非ゲート可視化）

- 状態: **実装済み（#186 / 9b530f08）**
- 動作: CI が `coverage.out` を生成し Job Summary + artifact に publish するが、しきい値なし
- ブランチ対象: main, staging, production への PR / push

### Phase 1（PR コメント + 低下 warn）

- 目標: PR ごとにカバレッジ数値をコメント表示し、前回比で低下した場合に warn を出す
- 実装: `github-script` または `go-coverage-report` action の導入
- しきい値設定: なし（warn のみ、fail なし）
- ブランチ対象: main への PR
- 発火条件: backend / frontend いずれかのコード変更を含む PR

### Phase 2（パッチカバレッジ warn）

- 目標: 変更ファイルのパッチカバレッジが 70% 未満のとき warn を出す
- 実装: `octocov` または `diff-cover` によるパッチカバレッジ計算
- しきい値: **warn（fail なし）** — 既存 PR をブロックしない
- ブランチ対象: main への PR
- 発火条件: backend/handler/service/repository または frontend/src/features の変更を含む PR
- 着手条件: Phase 1 が安定稼働し、ベースライン数値が記録されてから

### Phase 3（ディレクトリ別 fail ゲート）

- 目標: service / handler / repository の高重要ディレクトリに fail ゲートを設ける
- しきい値（案）:
  - `internal/service/`: 80%
  - `internal/handler/`: 70%
  - `internal/repository/`: 75%
- 実装: `go tool cover -func` の grep + 閾値判定スクリプト
- ブランチ対象: main, staging への PR
- 発火条件: 対象ディレクトリの変更を含む PR
- 着手条件: Phase 2 が安定稼働し、各ディレクトリのベースラインが目標値を上回っていること

---

## 参考リンク

- Issue #194: [FOLLOW-UP] #186 CI: カバレッジ計測ポリシー
- PR #186 / commit 9b530f08: 非ゲートカバレッジ artifact 導入
- `.github/workflows/ci.yml`: Backend/Frontend Test ステップ（`-coverpkg` フラグ、`coverage.exclude` 参照）
- `frontend/vite.config.ts`: Vitest `test.coverage` 設定
