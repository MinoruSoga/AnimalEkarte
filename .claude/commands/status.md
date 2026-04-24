---
description: プロジェクト進捗・git 状態の確認
---

# 進捗状況確認

以下を実行して現在の状況を報告してください：

1. `.claude/logs/session-progress.md` の内容を表示（stop-save-progress.js が自動記録）
2. `git log --oneline -5` で最近のコミットを表示
3. `git status --short` で未コミットの変更を確認

## 出力形式

### 進捗サマリー
- 完了: X / Y 機能
- 現在の作業: [機能名]
- 次のステップ: [概要]

### 最近のコミット
[git log出力]

### 注意事項
[進捗ファイルからの注意事項]
