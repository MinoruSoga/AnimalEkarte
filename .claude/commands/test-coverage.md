---
description: テストカバレッジ分析・80%未満のファイルにテスト生成
argument-hint: "[path] (blank for full project)"
---

# テストカバレッジ分析

$ARGUMENTS のカバレッジを分析し、80% 未満のファイルにテストを生成します。

## 実行手順

### Step 1: カバレッジ計測（ユーザーに実行依頼）

**Backend (Go):**
```bash
docker compose exec backend go test -coverprofile=coverage.out ./...
docker compose exec backend go tool cover -func=coverage.out | sort -k3 -n
```

**Frontend (TypeScript/Vitest):**
```bash
docker compose exec frontend pnpm test:run --coverage
```

### Step 2: カバレッジ結果を分析

ユーザーが貼り付けた結果から:
1. 80% 未満のファイルを特定（ワーストから順に列挙）
2. 各ファイルの未テスト関数・ブランチを特定
3. 優先度付け（Critical business logic → Public API → 一般コード）

### Step 3: テスト生成

優先度順に、各ファイルのテストを生成:

**テスト優先順位:**
1. ハッピーパス（正常系）
2. エラーハンドリング（異常系・バリデーション）
3. エッジケース（空・nil・境界値）
4. ブランチカバレッジ（if/else・switch・apperrors 各コード）

**Go テストの規約:**
- Table-driven tests 形式
- `apperrors.FromGORM()` の各エラーコードをカバー
- clinic_id の境界テスト（0, 有効ID, 存在しないID）

**TypeScript テストの規約:**
- Vitest + Testing Library
- React 19 Action のフォーム状態テスト
- API フックのモック（msw）

### Step 4: 報告

```
カバレッジレポート
──────────────────────────────────
ファイル                          変更前  変更後
internal/service/owner_service.go  45%    88%
src/features/owners/hooks/         32%    82%
──────────────────────────────────
合計:                              67%    84%  ✅ 目標達成
```

## 注意

- テスト生成後は **ユーザーに手動実行を依頼** してから調整
- 生成コードのパスは既存テストの命名規則に従う
- `docker compose exec` 経由でのテスト実行は自動実行禁止
