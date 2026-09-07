---
description: 実装→規約チェック→承認ループ型ハーネス（最大3イテレーション）
argument-hint: "TASK-XXX | BUG-XXX | <タスク説明>"
---

# 実装ハーネス（Implement → Verify → Approve Loop）

このプロジェクトのGo/Gin公式ガイド、application安全不変条件、React 19パターンへの
準拠を自動ループで保証する実装ハーネス。最大3イテレーションで承認を目指す。

---

## ハーネスの流れ

```
[Phase 1] タスク分析・計画
       ↓
[Phase 2] 実装（イテレーション N）
       ↓
[Phase 3] 規約チェック（自動判定）
       ↓ PASS → [Phase 4] 承認レポート
       ↓ FAIL → フィードバック付きで Phase 2 に戻る（最大3回）
```

---

## $ARGUMENTS の解釈

- Linear Issue ID → Linearを実行SoTとして本文・状態・依存関係を確認して実装モードで対象化。`docs/work/phase2-deferred.md`は見送り索引であり実装台帳ではない。
- テキスト → タスク説明として直接扱う
- 省略 → `git status` で未コミット変更を対象として検証のみ実行

---

## Phase 1: タスク分析・計画

```bash
git status --short
git diff --name-only HEAD
```

1. 引数からタスク種別を判定（BE/FE/テキスト/変更済みファイル）
2. 影響するpackage・request lifecycle・migration・frontendを特定
3. 関連する規約ファイルを特定：
   - Go変更 → `.claude/rules/go-gin-backend-guidelines.md` と `.claude/refs/go-gin-backend-review.md`
   - tenant/ownership変更 → `.claude/refs/backend-application-invariants.md`
   - TS変更 → `.claude/refs/typescript-react.md`
   - DB変更 → `postgres-patterns` / `migration-seed-safety` スキル
4. ハーネス状態を初期化：
   ```
   HARNESS START: <タスク説明>
   モード: <BE|FE|VERIFY>
   最大イテレーション: 3
   ```

---

## Phase 2: 実装（Generator）

イテレーション1の場合：
- `implementer` エージェントに委譲して実装
- Docker必須ルール遵守（`npm`/`go`直接実行禁止）

イテレーション2以降の場合：
- 前イテレーションのフィードバックを冒頭に提示
- 指摘箇所のみピンポイントで修正（全体再実装しない）
- 修正したファイルとその理由を明示

---

## Phase 3: 規約チェック（Evaluator）

### 3-1. ファイル変更の収集

```bash
git diff --name-only HEAD
git diff --staged --name-only
```

### 3-2. Go変更がある場合：Go/Gin backendチェック

`go-reviewer` エージェントを起動し、package API、Context、HTTP boundary、error/log、database/security、server lifecycle、testsを確認する。Handler/Service/Repositoryの存在や特定helper名は合否条件にしない。

判定基準：
- CRITICAL違反（マージブロック）→ **FAIL**
- WARNING以下のみ → **PASS**（警告は記録）

### 3-3. TS/TSX変更がある場合：React 19パターンチェック

`typescript-reviewer` エージェントを起動し、以下を確認：

- `any` 型使用 → **FAIL**
- Feature Indexing違反（deep import） → **FAIL**
- `useActionState` + `SubmitButton` パターン違反 → **FAIL**
- デザイントークン未使用（インラインCSS） → WARNING

### 3-4. Migration変更がある場合

`database-reviewer` エージェントを起動し、以下を確認：

- `clinic_id` カラム欠落 → **FAIL**
- `CASCADE DELETE` 使用 → **FAIL**
- インデックス不足 → WARNING

### 3-5. チェック結果の判定

```
ITERATION N RESULT:
  Go:   PASS / FAIL / UNKNOWN（違反: X件。未実行なら UNKNOWN）
  TS:   PASS / FAIL / UNKNOWN
  DB:   PASS / FAIL / UNKNOWN
  総合: PASS / FAIL / UNKNOWN

未実行の reviewer・テスト・coverage を PASS と表示しない。カバレッジ数値は測定した場合だけ書く。
```

---

## Phase 4: ループ制御

### PASS の場合 → 承認レポート出力

```markdown
## ハーネス完了: APPROVED ✅

**タスク**: <説明>
**イテレーション**: N / 3
**所要時間**: <開始から>

### 変更サマリー
<変更ファイル一覧>

### 規約準拠
- Go/Gin backend: ✅ PASS
- TypeScript: ✅ PASS
- Database: ✅ PASS

### 残 WARNING
<あれば記載>

### 次のアクション
$ make build   # ビルド確認（手動実行）
$ git add -p   # 変更を確認してステージング
```

### FAIL かつイテレーション < 3 の場合 → Phase 2 に戻る

フィードバックを明示：
```
ITERATION N FAILED → RETRY (N+1/3)

修正が必要な問題:
1. [CRITICAL] <ファイル>:<行> — <問題> → <修正方法>
2. [CRITICAL] <ファイル>:<行> — <問題> → <修正方法>
```

### FAIL かつイテレーション = 3 の場合 → 失敗レポート

```markdown
## ハーネス完了: BLOCKED ❌

3イテレーション後も未解決の問題があります。

### 未解決の違反
<リスト>

### 推奨アクション
- 手動での設計見直しが必要な可能性があります
- `/review` でより詳細なレビューを実施してください
```

---

## 制約（Docker必須ルール厳守）

- `npm`、`go`、`pnpm` の直接実行は禁止
- ビルド・テスト確認は必ずユーザーに手動実行を依頼
- `docker compose exec` も自動実行禁止（ユーザーへ案内のみ）
- ファイル読み書き・git操作のみ自律実行
