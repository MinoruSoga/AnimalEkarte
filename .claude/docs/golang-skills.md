# Golang Skills ガイド

インストール済みのGolang関連スキルと推奨プロンプトのリファレンス。

## インストール済みスキル一覧

| スキル | インストール元 | 用途 |
|--------|--------------|------|
| `/go-code-review` | `cxuu/golang-skills` | Go Wiki準拠のコードレビューチェックリスト |
| `/go-linting` | `cxuu/golang-skills` | フォーマット・命名・インポート順のLintチェック |
| `/golang-gin-clean-arch` | `henriqueatila/golang-gin-clean-arch` | Gin + クリーンアーキテクチャ実装ガイド |
| `/golang-pro` | `jeffallan/claude-skills` | プロレベルの実装・パフォーマンス最適化 |
| `/api-designer` | `jeffallan/claude-skills` | REST API設計・エンドポイント設計全般 |
| `/openapi-specification-v2` | `hairyf/skills` | OpenAPI 2.0（Swagger）仕様準拠のAPI定義 |

---

## `/go-code-review` — コードレビュー

**概要:** [Go Wiki CodeReviewComments](https://github.com/golang/wiki/blob/master/CodeReviewComments.md) ベースのチェックリスト。フォーマット・ドキュメント・エラーハンドリング・命名・並行性・インターフェース・セキュリティを網羅。

### 推奨プロンプト

```
/go-code-review
backend/internal/handler/owner.go をレビューしてください
```

```
/go-code-review
backend/internal/service/ 配下の全サービスをレビューしてください
```

```
/go-code-review
今回実装したコードをGo Wiki CodeReviewCommentsに照らしてレビューしてください
```

---

## `/go-linting` — Lintチェック

**概要:** gofmt・goimports・命名規則・インポートグループ順などの静的解析観点でコードを検査・修正。

### 推奨プロンプト

```
/go-linting
backend/internal/repository/master.go のlint問題を全て修正してください
```

```
/go-linting
backend/internal/ 配下を全てlintチェックし、問題箇所を列挙してください
```

---

## `/golang-gin-clean-arch` — Gin + クリーンアーキテクチャ実装

**概要:** このプロジェクトと最も相性が良いスキル。以下の6原則を徹底する。

1. **Ginは詳細** — `gin.Context` は `internal/handler/` のみ
2. **DBは詳細** — GORM/SQLはEntity・Serviceに持ち込まない
3. **依存ルール** — Handler → Service → Repository の一方向
4. **Request/ResponseとDomainを分離** — Ginのリクエスト構造体をServiceに渡さない
5. **`main.go`のみがDIを知る** — DIの組み立ては `cmd/api/main.go` のみ
6. **インターフェースをexport、実装をhide** — コンストラクタはインターフェースを返す

### 推奨プロンプト

```
/golang-gin-clean-arch
XXX機能のhandler/service/repositoryを実装してください
```

```
/golang-gin-clean-arch
既存のhandler/owner.go がクリーンアーキテクチャに準拠しているか確認し、違反があれば修正してください
```

```
/golang-gin-clean-arch
backend/internal/ の層分離に違反している箇所を全て洗い出してください
```

---

## `/golang-pro` — プロレベル実装

**概要:** Go 1.21+ のイディオマティックなパターン、goroutine・channel・インターフェース設計・pprof によるパフォーマンス最適化・テーブル駆動テストを扱うシニアGo開発者スキル。

**コアワークフロー:**
1. アーキテクチャ分析 → インターフェース設計 → 実装 (`go vet` 確認)
2. `golangci-lint run` で全問題修正
3. pprof プロファイリング・ベンチマーク
4. `-race` フラグ付きテスト・80%以上カバレッジ

### 推奨プロンプト

```
/golang-pro
context伝播とエラーハンドリングがベストプラクティスに沿っているか、backend/internal/ 全体を確認してください
```

```
/golang-pro
backend/internal/service/master.go をidiomatic Goに書き直してください
```

```
/golang-pro
このコードのパフォーマンスボトルネックを特定し、改善案を提示してください
```

---

## backend フォルダ全体リファクタリング

4つのスキルを段階的に適用し、backend全体を体系的にリファクタリングする。

### Phase 1: クリーンアーキテクチャ違反の洗い出しと修正

```
/golang-gin-clean-arch
backend/ フォルダ全体をクリーンアーキテクチャの観点でリファクタリングしてください。

【対象ディレクトリ】
- backend/internal/handler/
- backend/internal/service/
- backend/internal/repository/
- backend/internal/model/
- backend/cmd/api/main.go

【チェック・修正観点】
1. gin.Context が handler 以外に漏れていないか
2. GORM/SQL が service/model に混入していないか
3. Handler → Service → Repository の依存方向が守られているか
4. Gin のリクエスト構造体が service に渡されていないか
5. DI の組み立てが main.go のみで行われているか
6. コンストラクタがインターフェースを返しているか

違反箇所を全て列挙し、優先度順に修正してください。
```

### Phase 2: イディオマティックGoへの書き直し

```
/golang-pro
backend/internal/ 配下を全てイディオマティックGoにリファクタリングしてください。

【チェック・修正観点】
1. context.Context が全関数の第一引数になっているか
2. エラーハンドリング — センチネルエラー定義 + fmt.Errorf("%w") ラッピング
3. インターフェース設計 — 小さく・消費者側定義・具体型をhide
4. slog 構造化ログが適切に使われているか（log.Printf 等の排除）
5. goroutine を使う箇所でライフタイムが明確か
6. 不要なポインタ渡し・過剰な抽象化の排除

各ファイルの修正前後を明示し、変更理由を説明してください。
```

### Phase 3: Go Wiki準拠のコードレビュー

```
/go-code-review
backend/internal/ 配下の全ファイルを Go Wiki CodeReviewComments に照らしてレビューしてください。

【重点チェック項目】
1. 命名規則 — MixedCaps・頭字語（URL/ID/HTTP）・レシーバ名
2. エラー文字列 — 小文字始まり・句読点なし
3. インポートグループ — 標準ライブラリ → 外部パッケージの順
4. インターフェース — 事前定義禁止・コンシューマ側定義
5. 空スライス — var t []string (nil) を優先
6. セキュリティ — crypto/rand 使用・panic 乱用禁止

問題箇所を重大度（Critical/Warning/Info）で分類して報告し、修正してください。
```

### Phase 4: Swagger定義の整合性チェック・修正

```
/api-designer /openapi-specification-v2
backend/docs/swagger.json および backend/docs/swagger.yaml を確認し、以下を修正してください。

【チェック・修正観点】
1. 全エンドポイントのリクエスト/レスポンス型定義が正確か
2. エラーレスポンス（400/404/500）が全エンドポイントに定義されているか
3. パスパラメータ・クエリパラメータの型・必須フラグが正しいか
4. OpenAPI 2.0（Swagger）仕様に完全準拠しているか
5. description が全エンドポイント・フィールドに記載されているか
6. Goのswagアノテーション（backend/internal/handler/）と swagger.json が一致しているか

問題箇所を列挙し、swagアノテーションと swagger.json の両方を修正してください。
修正後に `docker compose exec backend swag init -g cmd/api/main.go` で再生成して確認してください。
```

### Phase 5: Lint仕上げ

```
/go-linting
backend/ 配下を全てlintチェックし、全問題を修正してください。

【確認事項】
- gofmt / goimports 適用済みか
- golangci-lint run で0エラーになるか
- 修正後に docker compose exec backend go build ./... が通るか
```

---

### 全フェーズ一括プロンプト（時間を優先する場合）

```
/golang-gin-clean-arch /golang-pro /go-code-review /api-designer /openapi-specification-v2 /go-linting

backend/ フォルダ全体を以下の順序でリファクタリングしてください。

**Step 1 — アーキテクチャ修正 (golang-gin-clean-arch)**
- gin.Context の handler 外への漏れを排除
- GORM の service 混入を排除
- DI を main.go に集約
- コンストラクタをインターフェース返却に統一

**Step 2 — イディオマティックGo (golang-pro)**
- context.Context を全関数第一引数に統一
- センチネルエラー + %w ラッピングに統一
- slog 構造化ログに統一
- 不要な抽象化・過剰ポインタを排除

**Step 3 — コードレビュー修正 (go-code-review)**
- 命名規則違反を修正
- エラー文字列を小文字・句読点なしに統一
- インポートグループ順を修正

**Step 4 — Swagger定義修正 (api-designer / openapi-specification-v2)**
- 全エンドポイントのリクエスト/レスポンス型を正確に定義
- エラーレスポンス（400/404/500）を全エンドポイントに追加
- swagアノテーションと swagger.json の整合性を確認
- `docker compose exec backend swag init -g cmd/api/main.go` で再生成

**Step 5 — Lint仕上げ (go-linting)**
- gofmt / goimports 適用
- golangci-lint 0エラーを確認

各Stepの修正完了後に `docker compose exec backend go build ./...` でビルドが通ることを確認してください。
```

---

## 組み合わせワークフロー

### 新機能実装時

```
# Step 1: クリーンアーキテクチャ準拠で実装
/golang-gin-clean-arch
XXX機能を実装してください

# Step 2: コードレビュー
/go-code-review
実装したコードをレビューしてください

# Step 3: Lint修正で仕上げ
/go-linting
lint問題を全て修正してください
```

### 既存コードの品質改善時

```
# 深い品質チェック
/golang-pro
backend/internal/service/ をイディオマティックGoの観点で改善してください

# 仕上げのLint
/go-linting
修正後のコードをlintチェックしてください
```
