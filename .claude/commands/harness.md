---
description: 実装→規約チェック→承認ループ型ハーネス（最大3イテレーション）
argument-hint: "FEAT-XXX | BUG-XXX | <タスク説明>"
---

# 実装ハーネス（Implement → Verify → Approve Loop）

このプロジェクトのアーキテクチャ規約（P1-P18、React 19パターン、Clean Architecture）への
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

- `FEAT-XXX` / `BUG-XXX` / `PERF-XXX` / `SEED-XXX` 等のタスクID → `docs/tasks/open/` のタスクを実装モードで対象化
- テキスト → タスク説明として直接扱う
- 省略 → `git status` で未コミット変更を対象として検証のみ実行

---

## Phase 1: タスク分析・計画

```bash
git status --short
git diff --name-only HEAD
```

1. 引数からタスク種別を判定（BE/FE/テキスト/変更済みファイル）
2. 影響レイヤーを特定：handler / service / repository / migration / frontend
3. 関連する規約ファイルを特定：
   - Go変更 → `.claude/refs/gin-architecture-compliance.md`（P1-P18）
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

### 3-2. Go変更がある場合：P1-P18 チェック

`go-reviewer` エージェントを起動し、以下を確認：

| レイヤー | チェック項目 |
|---------|------------|
| Handler | P7(エラーハンドリング), P12(バリデーション), P14(認証), P15(clinic_id), P18(レスポンス形式) |
| Service | P1(ビジネスロジック集中), P8(トランザクション), P10(エラーラップ), P11(ゼロ値), P13(コンテキスト伝播), P17(副作用分離) |
| Repository | P2(GORM使用), P3(クエリ最適化), P4(clinicScope), P9(N+1回避), P16(ソフトデリート) |
| Routes | P5(RESTful), P6(ミドルウェア順) |

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
  Go:   PASS / FAIL（違反: X件）
  TS:   PASS / FAIL（違反: X件）
  DB:   PASS / FAIL（違反: X件）
  総合: PASS / FAIL
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
- Go (P1-P18): ✅ PASS
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
