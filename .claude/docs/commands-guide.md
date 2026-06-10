# Slash Commands ガイド — AnimalEkarte

コマンドは `/コマンド名` で呼び出す。引数が必要な場合は `/コマンド名 引数` の形式。

---

## クイックリファレンス

| コマンド | 一言 | 典型的な呼び出し |
|---------|------|----------------|
| `/implement` | イシューを実装してクローズ | `/implement FE-123` |
| `/save` | 作業をコミット・記録 | 区切りのいい時 |
| `/status` | 進捗・git 状態を確認 | `/compact` 後・朝イチ |
| `/review` | ステージング変更のレビュー | PR 作成前 |
| `/review-pr` | GitHub PR を専門エージェントでレビュー | PR マージ前 |
| `/go-review` | Go コードの idiom レビュー | BE 実装後 |
| `/test` | テスト実行・カバレッジ | 実装後の確認 |
| `/gen-test` | テストを自動生成 | `/gen-test src/features/owners` |
| `/test-gen` | テスト自動生成（別パターン） | 統合テスト生成 |
| `/fe-build` | TS/React ビルドエラー解決 | 型エラー発生時 |
| `/go-build` | Go ビルド・lint エラー解決 | go vet 失敗時 |
| `/debug` | バグ調査・根本原因分析 | 再現性ありのバグ |
| `/refactor` | リファクタリング分析・実行 | 技術的負債解消時 |
| `/security-audit` | フルセキュリティスキャン | PR 前・定期監査 |
| `/db-schema-review` | DB スキーマ設計レビュー | マイグレーション追加時 |
| `/e2e-design` | E2E テストシナリオ設計 | 新機能完成後 |
| `/tdd-workflow` | TDD Red-Green-Refactor ガイド | 新機能の TDD 開発 |
| `/deploy` | デプロイ準備チェック | staging/production 前 |
| `/ci-trigger` | CI パイプライン確認・トリガー | CI 失敗調査 |
| `/perf-profile` | パフォーマンスプロファイリング | 速度問題発生時 |
| `/docs` | JSDoc/TSDoc 生成 | API 公開前 |
| `/sync-claude-md` | CLAUDE.md をコードに同期 | 大規模リファクタ後 |
| `/checkpoint` | 作業チェックポイントの作成・確認 | 長いセッションの節目 |
| `/plan` | 実装計画作成・ユーザー確認後に着手 | 新機能・大規模変更の前 |
| `/test-coverage` | カバレッジ分析・80% 未満にテスト生成 | テスト不足の調査・改善時 |
| `/harness` | 実装→P1-P18チェック→承認ループ（最大3回） | 規約準拠を保証したい実装時 |
| `/harness-status` | ハーネスの現在状態を確認・リセット | ハーネス実行中の進捗確認 |

---

## ユースケース別ガイド

### 1. 朝の作業開始

```
/status          → 昨日からの変更・git 状態を把握
（次タスクは GitHub Issues / docs/tasks/open/ から選択する）
```

### 2. 機能実装（標準フロー）

```
/plan 機能の説明            → 実装計画作成・リスク特定・ユーザー承認待ち
/implement FE-123          → イシュー読み込み → 実装 → セルフレビュー → クローズ
/save                      → コミット
/review --fe               → PR 作成前の最終チェック（フロントエンド変更の場合）
```

TDD で進める場合:
```
/tdd-workflow FE-123       → Red-Green-Refactor サイクルをガイドしながら実装
```

規約準拠を自動ループで保証したい場合:
```
/harness FE-123            → 実装 → P1-P18チェック → 問題があれば自動修正 → 承認
/save                      → APPROVED 後にコミット
```

### 3. ビルドエラー・型エラーが出た

```
# TypeScript / React
/fe-build                  → エラーログを解析して修正提案

# Go
/go-build                  → go vet / golangci-lint エラーを解析して修正提案
```

### 4. バグ修正

```
/debug バグの症状や再現手順    → 根本原因を調査して修正案を提示
/test                        → 修正後の回帰テスト実行
/save                        → コミット
```

### 5. コードレビュー・品質チェック

```
/review                    → ステージング済み変更を専門エージェントでレビュー
/review-pr 42              → GitHub PR #42 をレビューして gh pr review で投稿
/go-review                 → Go コード限定の idiom・パフォーマンス確認
/security-audit            → OWASP / gosec / pnpm audit の全スキャン
```

### 6. テスト追加

```
/gen-test frontend/src/features/owners    → 対象ディレクトリのテストを自動生成
/test frontend/src/features/owners        → 生成したテストを実行してカバレッジ確認
/test-coverage                            → プロジェクト全体のカバレッジ分析・80% 未満を特定
/test-coverage internal/service           → 特定ディレクトリのカバレッジ改善
```

### 7. DB スキーマ変更

```
/db-schema-review          → マイグレーションの設計レビュー（clinic_id・インデックス等）
```

### 8. デプロイ前チェック

```
/security-audit            → セキュリティスキャン
/deploy staging            → staging デプロイ準備チェックリスト
/ci-trigger                → CI パイプラインの状態確認・再トリガー
/deploy production         → production デプロイ最終確認
```

### 9. パフォーマンス問題

```
/perf-profile              → Go pprof / Lighthouse 計測・ボトルネック特定
/refactor 対象ファイルパス   → リファクタリング提案（パフォーマンス観点も含む）
```

### 10. ドキュメント整備

```
/docs src/features/owners/api    → JSDoc/TSDoc を自動生成
/sync-claude-md --hooks          → hooks/CLAUDE.md のフック一覧テーブルを現状に同期
/sync-claude-md --refs           → .claude/CLAUDE.md の死亡パスを検出
/sync-claude-md                  → 上記両方を実行
```

---

## コマンド選択の判断基準

### `/gen-test` vs `/test-gen`

どちらもテスト生成だが対象が異なる:
- `/gen-test` — 指定パスのユニットテスト生成（Vitest / go test）
- `/test-gen` — より統合テスト寄りのテスト設計・生成

### `/review` vs `/review-pr` vs `/go-review`

- `/review` — ステージング済み変更全体をレビュー（ローカル、GitHub 投稿なし）
- `/review-pr 42` — GitHub PR を取得してレビューし `gh pr review` で投稿まで実行
- `/go-review` — Go コードの idiom・パフォーマンス専門レビュー（`/review` より深い）

### `/implement` vs `/harness` vs 手動実装

- イシュー番号（`BE-XXX` / `FE-XXX`）があり規約違反リスクが低い → `/implement` で一気通貫
- P1-P18 / React 19 パターンへの準拠を自動保証したい → `/harness` でループ付き実装
- 探索的な作業・設計が必要 → `/plan` で計画してから `/implement` または `/harness`

### `/plan` の使いどころ

- 要件が複数ファイルにまたがる → `/plan` でフェーズ分割して承認を得てから実装
- 単一ファイルの軽微な変更 → `/plan` 不要、直接実装

### `/gen-test` vs `/test-gen` vs `/test-coverage`

- `/gen-test` — 指定パスのユニットテスト生成（Vitest / go test）
- `/test-gen` — より統合テスト寄りのテスト設計・生成
- `/test-coverage` — カバレッジ計測→80% 未満を分析→不足テストを生成（網羅的）

---

## 引数パターン

引数を定義しているコマンドのみ記載する。

```bash
# /implement — イシュー番号のみ
/implement BE-456
/implement FE-123

# /review — 専門エージェント指定（省略時は自動判定）
/review --go        # Go コードに絞る
/review --fe        # フロントエンドに絞る
/review --db        # DB 変更に絞る
/review --security  # セキュリティ観点に絞る
/review             # 変更内容から自動選択

# /review-pr — GitHub PR レビュー（gh pr review で投稿）
/review-pr 42                    # PR #42 をレビュー
/review-pr 42 --focus=security   # セキュリティ観点に絞る
/review-pr 42 --focus=go         # Go コードに絞る
/review-pr                       # 現在ブランチの PR を使用

# /fe-build — チェック種別（省略時は型チェックのみ）
/fe-build           # type-check のみ
/fe-build --lint    # lint も実行
/fe-build --build   # ビルドも実行

# /go-build — チェック種別（省略時はビルドのみ）
/go-build           # build のみ
/go-build --vet     # go vet も実行
/go-build --lint    # golangci-lint も実行

# /sync-claude-md — 対象範囲（省略時は --all）
/sync-claude-md --hooks   # hooks/CLAUDE.md のみ
/sync-claude-md --refs    # .claude/CLAUDE.md refs のみ
/sync-claude-md           # 両方

# /deploy — 環境名
/deploy staging
/deploy production

# /test, /gen-test, /test-gen — パス指定
/test frontend/src/features/owners
/gen-test frontend/src/features/owners
/test-gen internal/service/owner_service.go

# /debug — 症状を自由記述
/debug "ログインすると 500 エラーが出る"

# /refactor — パス指定（省略時はステージング変更から自動検出）
/refactor frontend/src/features/owners/components/OwnerCard.tsx
/refactor                  # ステージング変更から自動検出

# /docs — パス指定
/docs src/features/owners/api
/docs backend/internal/handler/owner_handler.go

# /harness — イシュー番号またはタスク説明
/harness BE-042              # バックエンドイシューをハーネスで実装（Go P1-P18チェック）
/harness FE-038              # フロントエンドイシューをハーネスで実装（React 19チェック）
/harness "clinic_id を patients テーブルに追加"  # テキスト説明でもOK
/harness                     # 引数なし = 未コミット変更を規約チェックのみ実行

# /harness-status — サブコマンド
/harness-status              # 現在のハーネス状態（タスク名・イテレーション・変更ファイル）を表示
/harness-status reset        # harness-active.json を削除してリセット

# /checkpoint — サブコマンド
/checkpoint create "feature-start"   # チェックポイント作成
/checkpoint list                     # 一覧表示
/checkpoint verify "feature-start"   # 現在状態との比較
/checkpoint                          # list と同じ

# /plan — 機能の説明を自由記述（実装前の計画作成）
/plan "Owner の一括インポート機能を追加したい"
/plan BE-456                          # イシュー番号でも可

# /test-coverage — パス指定（省略時はプロジェクト全体）
/test-coverage                        # 全体分析
/test-coverage internal/service       # Backend 特定ディレクトリ
/test-coverage src/features/owners    # Frontend 特定ディレクトリ

# /e2e-design — 機能名
/e2e-design owners
/e2e-design medical-records

# /go-review — パス指定（省略時はステージング済み Go ファイル）
/go-review internal/handler/owner_handler.go
/go-review internal/service

# /tdd-workflow — 機能・イシュー番号
/tdd-workflow BE-123
/tdd-workflow FE-456

# /ci-trigger — オプション
/ci-trigger               # GitHub Actions 状態確認
/ci-trigger --local       # ローカル CI 実行（ユーザーが手動実行）
/ci-trigger --watch       # 継続監視
```
