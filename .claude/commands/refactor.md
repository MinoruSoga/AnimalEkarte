---
description: コードリファクタリング — デッドコード検出・安全削除・構造改善
argument-hint: "[path] (blank for auto-detect from staged changes)"
---

# リファクタリング

$ARGUMENTS のリファクタリングを分析・実施します。

## Step 1: デッドコード検出（ユーザーに手動実行を依頼）

**TypeScript（フロントエンド）:**
```bash
docker compose exec frontend pnpm unused
```

**Go（バックエンド）:**
```bash
docker compose exec backend deadcode ./...
```

## Step 2: 検出結果の安全ティア分類

| ティア | 対象例 | アクション |
|--------|--------|----------|
| **SAFE** | 未使用ユーティリティ、内部ヘルパー関数 | 即削除可 |
| **CAUTION** | コンポーネント、APIルート、ミドルウェア | 動的 import・外部参照確認後 |
| **DANGER** | 設定ファイル、エントリポイント、型定義 | 要精査・原則触らない |

CAUTION 確認手順:
```bash
# 動的インポートの確認
grep -r "import('" frontend/src/ | grep "<target>"
# 文字列参照の確認
grep -r '"<ComponentName>"' frontend/src/
```

## Step 3: 構造的問題の分析

対象ファイルを読み、以下を特定:

1. **循環的複雑度** — 関数 > 50行、ネスト > 4段
2. **コード重複** — 80%以上類似のロジック
3. **型安全性** — `any` 使用、unsafe cast、型情報の欠落
4. **Feature Indexing 違反** — deep import（`features/owners/components/Foo` など）
5. **Go/Gin設計違反** — package API、Context、error chain、resource cleanup、HTTP/security boundary

## Step 4: SAFE ティア安全削除ループ

各アイテムに対して:
1. ユーザーにテスト実行を依頼（ベースライン確認）
2. Edit ツールで削除を適用
3. ユーザーにテスト再実行を依頼
4. テスト失敗 → `git checkout -- <file>` で即時リバート・スキップ

## Step 5: 重複統合

削除後:
- 80%以上類似の関数をマージ
- 余分な再エクスポート（`index.ts`の pass-through のみ）を削除
- 単一用途ラッパー関数をインライン化

## 制約（変更禁止）

- `clinic_id` スコープを迂回する変更
- `apperrors` パターンの変更（エラーコードの意味論を保つ）
- `any` 型の導入
- テストなしで CAUTION/DANGER アイテムを削除

## 出力形式

```markdown
## リファクタリング提案

### SAFE 削除対象
| ファイル | 行 | 内容 | 理由 |
|---------|-----|------|------|

### 構造的問題
| 優先度 | 場所 | 問題 | 提案 |
|--------|------|------|------|
| HIGH   | ... | ... | ... |
| MEDIUM | ... | ... | ... |

### 重複統合候補
- `A` と `B` → マージ案

### 実行順序
1. SAFE 削除（テスト前後）
2. 構造修正（HIGH → MEDIUM）
3. 統合（最後）
```
