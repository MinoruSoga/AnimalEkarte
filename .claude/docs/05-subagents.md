# サブエージェント

サブエージェントは、特定のタスクに特化した事前設定済みAIアシスタントです。メイン会話とは独立したコンテキストで動作し、専門的なタスクを委譲できます。

## ファイルの場所

| タイプ | 場所 | スコープ |
|--------|------|---------|
| プロジェクトサブエージェント | `.claude/agents/` | 現在のプロジェクト |
| ユーザーサブエージェント | `~/.claude/agents/` | すべてのプロジェクト |

---

## サブエージェント定義ファイル形式

```markdown
---
name: code-reviewer
description: コード品質、セキュリティ、保守性を専門とするコードレビュー。コード作成・修正直後にプロアクティブに使用。
tools: Read, Grep, Glob, Bash
model: sonnet
permissionMode: default
---

高度なコード品質とセキュリティ基準を確保するコードレビュー専門家です。

## 実行時の手順
1. `git diff` で最近の変更を確認
2. 修正されたファイルに焦点を当てる
3. 以下のチェックリストに基づいてレビュー

## レビューチェックリスト
- コードの明確さと可読性
- 関数と変数の適切な命名
- 重複コードの排除
- 適切なエラーハンドリング
- シークレットや API キーの露出がないか
- 入力検証の実装
- テストカバレッジ
- パフォーマンス考慮事項

## フィードバック形式
優先度順に整理：
- **Critical**: 修正必須
- **Warning**: 修正すべき
- **Suggestion**: 改善検討
```

---

## frontmatter設定

| キー | 説明 | 例 |
|------|------|-----|
| `name` | エージェント名 | `code-reviewer` |
| `description` | 説明（使用タイミング含む） | レビュー時に使用 |
| `tools` | 利用可能なツール | `Read, Grep, Glob, Bash` |
| `model` | 使用モデル | `sonnet`, `opus`, `haiku` |
| `permissionMode` | 権限モード | `default`, `plan` |

---

## 組み込みサブエージェント

| 名前 | 用途 | 特徴 |
|------|------|------|
| **汎用（general-purpose）** | 複雑なマルチステップタスク | 全ツールアクセス |
| **Plan** | コードベース分析 | 読み取り専用、planモード |
| **Explore** | 高速コードベース検索 | Haikuモデル、読み取り専用 |

---

## サブエージェントの使用

### インタラクティブに管理
```bash
/agents
```

### 直接呼び出し
```
> Use the code-reviewer subagent to check my recent changes
```

### CLI でカスタム定義
```bash
claude --agents '{
  "debugger": {
    "description": "デバッグ専門家",
    "prompt": "エラー分析と修正を行います",
    "tools": ["Read", "Edit", "Bash", "Grep", "Glob"]
  }
}'
```

---

## サブエージェント例

### デバッガー

```markdown
---
name: debugger
description: バグ調査、エラー分析、修正提案を専門とするデバッグエキスパート
tools: Read, Grep, Glob, Bash
model: sonnet
---

バグの根本原因を特定し、修正を提案します。

## 手順
1. エラーメッセージを分析
2. スタックトレースを追跡
3. 関連コードを検索
4. 仮説を立てて検証
5. 修正案を提示
```

### リサーチャー

```markdown
---
name: researcher
description: コードベース探索、ファイル検索、構造理解を専門
tools: Read, Grep, Glob
model: haiku
permissionMode: plan
---

コードベースを効率的に探索し、情報を収集します。

## 得意なタスク
- ファイルやクラスの検索
- 依存関係の追跡
- 実装パターンの特定
```

### フォーマッター

```markdown
---
name: formatter
description: コミットメッセージ生成、コード整形を専門
tools: Bash, Read
model: haiku
---

Conventional Commits形式でコミットメッセージを生成します。

## 形式
- feat: 新機能
- fix: バグ修正
- docs: ドキュメント
- refactor: リファクタリング
- test: テスト
- chore: その他
```

---

## ベストプラクティス

1. **明確なdescription**: いつ使うべきかを明記
2. **最小限のtools**: 必要なツールのみ許可
3. **適切なmodel**:
   - 複雑な分析: `opus`
   - 一般的なタスク: `sonnet`
   - 高速検索: `haiku`
4. **具体的な手順**: 実行手順を明記
