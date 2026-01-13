# セキュリティ対策

## 基本原則

- **デフォルトは読み取り専用**: 機密操作には明示的な許可が必要
- **書き込み制限**: 起動ディレクトリとそのサブフォルダのみ書き込み可能
- **サンドボックス対応**: `/sandbox` コマンドでファイルシステム・ネットワーク隔離

---

## 組み込み保護機能

| 保護機能 | 説明 |
|---------|------|
| **権限システム** | 機密操作は明示的な承認が必要 |
| **コンテキスト分析** | 潜在的に有害な指示を検出 |
| **入力サニタイズ** | コマンドインジェクション防止 |
| **コマンドブロックリスト** | `curl`, `wget` などデフォルトでブロック |
| **ネットワークリクエスト承認** | 外部通信は承認が必要 |
| **分離コンテキスト** | Web fetchは別コンテキストで実行 |

---

## 機密ファイルの保護

```json
{
  "permissions": {
    "deny": [
      "Read(./.env)",
      "Read(./.env.*)",
      "Read(./secrets/**)",
      "Read(./config/credentials.json)",
      "Read(./.git/config)",
      "Read(~/.ssh/**)",
      "Read(~/.aws/**)",
      "Read(./**/*.pem)",
      "Read(./**/*.key)"
    ]
  }
}
```

---

## サンドボックス設定

```json
{
  "sandbox": {
    "enabled": true,
    "autoAllowBashIfSandboxed": true,
    "excludedCommands": ["git", "docker", "make"],
    "network": {
      "allowLocalBinding": true,
      "allowUnixSockets": []
    }
  }
}
```

### サンドボックスの効果

| 項目 | サンドボックス有効時 |
|------|---------------------|
| ファイルシステム | 制限されたアクセス |
| ネットワーク | 隔離 |
| プロセス | 制限されたシステムコール |
| 環境変数 | 制限されたアクセス |

---

## セキュリティベストプラクティス

### 機密コード作業時

1. 提案された変更は承認前に必ずレビュー
2. 機密リポジトリにはプロジェクト固有の権限設定を使用
3. devcontainer の使用を検討
4. `/permissions` で定期的に権限設定を監査

### チームセキュリティ

1. Enterprise managed policies で組織標準を強制
2. 承認済み権限設定をバージョン管理で共有
3. チームメンバーにセキュリティベストプラクティスを教育

### 信頼できないコンテンツ

1. 提案されたコマンドは承認前にレビュー
2. 信頼できないコンテンツを直接パイプしない
3. 重要ファイルへの変更提案を検証
4. 外部Webサービスとの連携時はVMの使用を検討

---

## 危険なコマンドのブロック

### settings.jsonでの設定

```json
{
  "permissions": {
    "deny": [
      "Bash(rm -rf:*)",
      "Bash(sudo:*)",
      "Bash(chmod 777:*)",
      "Bash(curl:*)",
      "Bash(wget:*)",
      "Bash(*|bash:*)",
      "Bash(*|sh:*)"
    ]
  }
}
```

### フックによるブロック

```bash
#!/bin/bash
# .claude/hooks/block-dangerous.sh
INPUT="$CLAUDE_TOOL_INPUT"
COMMAND=$(echo "$INPUT" | jq -r '.command // empty')

DANGEROUS_PATTERNS=(
  "rm -rf"
  "sudo"
  "chmod 777"
  "> /dev/"
  "mkfs"
  "dd if="
)

for pattern in "${DANGEROUS_PATTERNS[@]}"; do
  if [[ "$COMMAND" == *"$pattern"* ]]; then
    echo '{"decision": "block", "reason": "Dangerous command blocked: '"$pattern"'"}'
    exit 0
  fi
done

exit 0
```

---

## Windows WebDAVリスク

Windows環境では、WebDAVへのアクセスや `\\*` パスを含むWebDAVサブディレクトリへのアクセスを避けてください。権限システムをバイパスするネットワークリクエストがトリガーされる可能性があります。

---

## 監査ログ

### コマンドログの記録

```bash
#!/bin/bash
# .claude/hooks/audit-log.sh
LOG_FILE="$HOME/.claude/audit.log"
echo "[$(date -Iseconds)] Tool: $CLAUDE_TOOL_NAME" >> "$LOG_FILE"
echo "  Input: $CLAUDE_TOOL_INPUT" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"
exit 0
```

### settings.jsonでの設定

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash|Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/audit-log.sh"
          }
        ]
      }
    ]
  }
}
```

---

## セキュリティ問題の報告

セキュリティ脆弱性を発見した場合：
1. 公開しない
2. [Anthropic HackerOneプログラム](https://hackerone.com/anthropic-vdp/reports/new?type=team&report_type=vulnerability) で報告
3. 詳細な再現手順を含める

---

## 2つの必須分離メカニズム

効果的なサンドボックスには**両方**の分離が必要：

### 1. ファイルシステム分離

```
✅ 許可: /Users/dev/project/
❌ 拒否: /etc/passwd, ~/.ssh/, システムファイル
```

### 2. ネットワーク分離

```
✅ 許可: api.example.com, github.com
❌ 拒否: 未承認のサーバー
```

### 両方が必要な理由

> 「効果的なサンドボックスにはファイルシステムとネットワークの両方の分離が必要」

片方だけでは回避される：
- ファイルのみ → ネットワーク経由でデータ流出
- ネットワークのみ → ローカルファイル改ざん

---

## Sandboxed Bash Tool

OSレベルのプリミティブを使用：
- Linux: bubblewrap
- macOS: seatbelt

```bash
# サンドボックス設定確認
/sandbox

# 出力例:
# Filesystem: /Users/dev/project/ (read-write)
# Network: github.com, api.anthropic.com (allowed)
```

---

## Claude Code on the Web

分離されたクラウドサンドボックスでセッションを実行：

```
ユーザー → Claude Code Web → クラウドサンドボックス
                                    │
                              プロキシサービス
                              (検証・認証)
                                    │
                              Git リポジトリ
```

### 特徴
- 機密資格情報をサンドボックス外部に保持
- スコープ付き資格情報でGit操作を管理
- 認証トークン付加前にGit操作を検証

---

## サンドボックス機能の効果

これらの機能により：
- **84%の許可プロンプト削減**
- セキュリティの大幅な向上
- 開発者の生産性維持
