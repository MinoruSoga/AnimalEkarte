# トラブルシューティング

## MCP デバッグ

```bash
# デバッグモードで起動
claude --mcp-debug

# サーバー状態確認
/mcp

# 特定カテゴリのデバッグ
claude --debug=mcp,hooks
```

---

## よくある問題と解決策

| 問題 | 解決策 |
|------|--------|
| MCPサーバーが接続できない | `claude mcp list` で設定確認、`--mcp-debug` でデバッグ |
| フックが動作しない | 実行権限確認（`chmod +x`）、パス確認 |
| スキルが使用されない | description に具体的なトリガーワードを追加 |
| 権限エラー | `.claude/settings.json` の permissions 設定を確認 |
| メモリが読み込まれない | ファイルパスとYAML frontmatter構文を確認 |
| サブエージェントが動作しない | ファイル形式、保存場所を確認 |
| コンテキストが消える | `/compact` の使用、会話の継続確認 |

---

## ログとデバッグ

```bash
# 詳細ログ出力
claude --verbose

# デバッグモード（カテゴリ指定可能）
claude --debug

# 特定カテゴリのデバッグ
claude --debug=mcp,hooks,tools

# ログファイルの場所
~/.claude/logs/
```

---

## MCPサーバーのトラブルシューティング

### サーバーが読み込まれない

1. **JSON構文チェック**
   ```bash
   jq empty .mcp.json
   ```

2. **Claude Codeを再起動**

3. **ログを確認**
   ```bash
   tail -f ~/.claude/logs/mcp.log
   ```

### 接続エラー

1. **ネットワーク確認**
   ```bash
   curl -I <server-url>
   ```

2. **認証情報確認**
   - 環境変数が正しく設定されているか
   - トークンが有効か

3. **サーバープロセス確認**
   ```bash
   ps aux | grep <server-name>
   ```

---

## フックのトラブルシューティング

### フックが実行されない

1. **実行権限確認**
   ```bash
   chmod +x .claude/hooks/*.sh
   ```

2. **パス確認**
   - 絶対パスまたは `$CLAUDE_PROJECT_DIR` を使用

3. **matcher確認**
   - 正規表現が正しいか

### フックでエラーが発生

1. **手動実行でテスト**
   ```bash
   CLAUDE_TOOL_INPUT='{"command":"test"}' .claude/hooks/my-hook.sh
   ```

2. **エラー出力確認**
   ```bash
   # フックスクリプトでログ出力
   echo "Debug: $CLAUDE_TOOL_INPUT" >> /tmp/hook-debug.log
   ```

---

## スキルのトラブルシューティング

### スキルが使用されない

1. **ファイル名確認**
   - `SKILL.md`（大文字）であること

2. **パス確認**
   - `.claude/skills/<name>/SKILL.md`
   - または `~/.claude/skills/<name>/SKILL.md`

3. **description確認**
   - トリガーワードが含まれているか

4. **frontmatter確認**
   ```yaml
   ---
   name: my-skill
   description: 説明文にトリガーワードを含める
   ---
   ```

---

## 権限のトラブルシューティング

### 操作が拒否される

1. **現在の権限確認**
   ```bash
   /permissions
   ```

2. **settings.json確認**
   - `allow`, `deny`, `ask` の設定

3. **権限モード確認**
   - `Shift+Tab` で切り替え

### 予期しない自動承認

1. **サンドボックス設定確認**
   ```json
   {
     "sandbox": {
       "autoAllowBashIfSandboxed": false
     }
   }
   ```

2. **フック確認**
   - `PreToolUse` フックが自動承認していないか

---

## パフォーマンスの問題

### 応答が遅い

1. **モデル確認**
   - `opus` は遅い、`haiku` は高速

2. **MCPサーバー確認**
   - 不要なサーバーを無効化

3. **コンテキストサイズ確認**
   - `/compact` でコンテキスト圧縮

### メモリ使用量が高い

1. **MCPサーバー設定**
   ```json
   {
     "env": {
       "NODE_OPTIONS": "--max-old-space-size=512"
     }
   }
   ```

2. **不要なサーバー無効化**
   ```json
   {
     "mcpServers": {
       "unused-server": {
         "disabled": true
       }
     }
   }
   ```

---

## 会話のトラブルシューティング

### 会話が継続できない

```bash
# 会話一覧
claude -r

# 最新の会話を継続
claude -c

# 特定のセッションを再開
claude --resume <session-id>
```

### コンテキストが消える

1. **自動コンパクションの確認**
   - 長い会話では自動的にコンテキストが圧縮される

2. **手動でコンテキスト管理**
   ```bash
   /compact
   ```

3. **重要な情報をCLAUDE.mdに保存**

---

## その他の問題

### アップデートできない

```bash
# 手動アップデート
claude update

# NPMでインストールした場合
npm update -g @anthropic-ai/claude-code

# Homebrewでインストールした場合
brew upgrade claude-code
```

### 設定がリセットされる

1. **設定ファイルの場所確認**
   - ユーザー設定: `~/.claude/settings.json`
   - プロジェクト設定: `.claude/settings.json`

2. **バックアップ作成**
   ```bash
   cp ~/.claude/settings.json ~/.claude/settings.json.backup
   ```

---

## サポートを求める

1. **ドキュメント確認**: https://code.claude.com/docs/en/
2. **GitHub Issues**: https://github.com/anthropics/claude-code/issues
3. **バグ報告**: `/bug` コマンドを使用
