# Agent Skills（スキル）

スキルは、Claude Codeの機能を拡張するモジュール型の能力です。`SKILL.md` ファイルと補助ファイルで構成されます。

## スキルとサブエージェントの違い

| 項目 | スキル | サブエージェント |
|------|--------|-----------------|
| 呼び出し | Claudeが自律的に判断 | ユーザーが明示的に呼び出し |
| 実行 | メインコンテキスト内 | 独立したコンテキスト |
| 用途 | 特定タスクのガイダンス | 専門的なエージェント |

---

## スキルの保存場所

```
# 個人スキル（全プロジェクト共通）
~/.claude/skills/my-skill-name/SKILL.md

# プロジェクトスキル（チーム共有）
.claude/skills/my-skill-name/SKILL.md
```

---

## SKILL.md の形式

```yaml
---
name: generating-commit-messages
description: Gitの差分からコミットメッセージを生成。コミット作成時やステージ済み変更のレビュー時に使用。
allowed-tools: Read, Grep, Glob, Bash
---

# コミットメッセージ生成

## 手順
1. `git diff --staged` で変更を確認
2. 以下の形式でコミットメッセージを提案:
   - 50文字以内の要約
   - 詳細な説明
   - 影響を受けるコンポーネント

## メッセージ形式
Conventional Commits形式を使用:
- feat: 新機能
- fix: バグ修正
- docs: ドキュメント
- style: フォーマット
- refactor: リファクタリング
- test: テスト
- chore: その他
```

---

## frontmatter設定

| キー | 説明 | 必須 |
|------|------|------|
| `name` | スキル名（ケバブケース） | ✅ |
| `description` | 説明とトリガーワード | ✅ |
| `allowed-tools` | 許可されたツール | ❌ |
| `version` | バージョン | ❌ |

---

## スキルのベストプラクティス

### 推奨

- **1スキル = 1つの明確な責任**
- **具体的なトリガーワードを含む説明**
- **必要なツールのみアクセス許可**
- **プロジェクトスキルはバージョン管理**

### スキル名の例

- `pdf-form-filling` - PDF入力支援
- `excel-data-analysis` - Excelデータ分析
- `generating-commit-messages` - コミットメッセージ生成
- `reviewing-code` - コードレビュー支援
- `analyzing-prisma` - Prismaスキーマ分析

---

## スキル例

### コードレビュー

```yaml
---
name: reviewing-code
description: コードの品質、セキュリティ、保守性をレビュー。「レビュー」「チェック」などで起動。
allowed-tools: Read, Grep, Glob, Bash
---

# コードレビュー

## チェック項目
- [ ] 型安全性（any禁止）
- [ ] エラーハンドリング
- [ ] セキュリティ（XSS、SQLi）
- [ ] パフォーマンス
- [ ] 重複コード
- [ ] 命名規約

## 出力形式
**Critical**（修正必須）:
**Warning**（修正推奨）:
**Suggestion**（改善案）:
```

### Prismaスキーマ分析

```yaml
---
name: analyzing-prisma
description: Prismaスキーマの分析と改善提案。「スキーマ」「モデル」「DB」などで起動。
allowed-tools: Read, Grep, Glob
---

# Prismaスキーマ分析

## 確認項目
- リレーション設計
- インデックス最適化
- 命名規約
- マイグレーション安全性

## 出力
- 現状の課題
- 改善提案
- マイグレーション計画
```

### テスト作成

```yaml
---
name: testing
description: テストケースの作成と実行。「テスト」「カバレッジ」などで起動。
allowed-tools: Read, Write, Bash, Grep, Glob
---

# テスト作成

## 手順
1. 対象コードを分析
2. テストケースを設計
3. テストファイルを作成
4. テストを実行

## テストの種類
- ユニットテスト: 関数単位
- 統合テスト: モジュール間
- E2Eテスト: ユーザーフロー

## 品質基準
- カバレッジ80%以上
- エッジケースを含む
- 読みやすいテスト名
```

---

## スキルの管理

### スキル一覧確認
```bash
ls -la ~/.claude/skills/
ls -la .claude/skills/
```

### スキルの無効化
ディレクトリ名に `.disabled` を追加：
```bash
mv .claude/skills/my-skill .claude/skills/my-skill.disabled
```

---

## スキルが使用されない場合の対処

1. **descriptionを確認**: トリガーワードが含まれているか
2. **ファイル名を確認**: `SKILL.md`（大文字）であるか
3. **パスを確認**: 正しい場所に配置されているか
4. **Claude Codeを再起動**: キャッシュをクリア

---

## 段階的情報開示アーキテクチャ

Agent Skills は3段階の設計を採用してトークン効率を最適化：

```
第1段階: メタデータのみ（name, description）
    ↓ エージェントがスキル使用を判断
第2段階: 完全なSKILL.mdコンテンツ
    ↓ 基本的な指示とガイドライン
第3段階+: 参照ファイル（reference.md, examples/）
    ↓ 詳細情報をオンデマンドでロード
```

### 利点
- 初期コンテキストを最小化
- 必要に応じて詳細情報をロード
- トークン効率の最適化

---

## スケーリングのための構造化

SKILL.mdが大きくなった場合、相互に排他的なコンテキストを別パスに分離：

```
skills/
├── deployment/
│   ├── SKILL.md        # 簡潔に保つ
│   ├── aws/            # AWS固有
│   ├── gcp/            # GCP固有
│   └── azure/          # Azure固有
```

---

## セキュリティ考慮事項

> 「信頼できるソースからのスキルのみをインストール」

不確実なソースからのスキルは以下を監査：

| チェック項目 | 確認内容 |
|--------------|----------|
| コード依存関係 | 悪意のあるパッケージがないか |
| バンドルリソース | 不正なファイルがないか |
| 外部ネットワーク接続 | 不審な通信がないか |
