---
description: TDD ガイド（Red-Green-Refactor サイクル）
argument-hint: "<feature> (e.g. FEAT-123, BUG-456)"
---

# /tdd-workflow [feature]

テスト駆動開発のワークフローをガイドします。

## Red-Green-Refactor サイクル

### 1️⃣ Red（テスト作成）
```bash
# 失敗するテストを作成
docker compose exec backend go test -run TestCreateOwner -v
# → FAIL
```

### 2️⃣ Green（最小限の実装）
```bash
# テストを通す最小限の実装
# 対象 domain package（例: internal/owner）に最小実装を置く
docker compose exec backend go test ./internal/owner/... -run TestCreateOwner -v
# → PASS
```

### 3️⃣ Refactor（改善）
```bash
# コードをクリーンアップ
# エッジケース追加
# 重複削除
docker compose exec backend go test -run TestOwner -v
# → PASS + カバレッジ向上
```

## ワークフロー例

### Backend (Go)
```
Step 1: テストスケルトン作成
Step 2: API パスを実装（スタブ）
Step 3: Service ロジック実装
Step 4: エッジケース追加
Step 5: リファクタリング
Step 6: Coverage 確認（`docs/ops/coverage-policy.md` の基準）
```

### Frontend (React)
```
Step 1: テストケース定義
Step 2: Component スケルトン
Step 3: ロジック実装
Step 4: スタイリング
Step 5: インタラクション実装
Step 6: E2E テスト確認
```

## 使用エージェント

`tdd-guide` (Sonnet) のガイド付き実装
