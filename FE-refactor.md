# FE-refactor 第7期（FE7）— frontend/ 構成の守りの機械化と小粒整理

> 起票: 2026-07-17（要件責任者: 曽我。フォルダ構成チェックの結果を受けて起票）
> **本ファイルは対応後削除する使い捨て計画書**。実行時に確定した恒久規約は `frontend/CLAUDE.md`（+必要なら `.claude/rules/`）へ同梱してから削除する。
> 読者 = 着手するエージェント（Sonnet 5 想定）。着手時に §1 の計測を再実行してから始めること。
> **着手条件: Go-live（2026-07-18）完了後**（FE7-0 は tooling-only で即日可能だが、納品前日の CI 変更を避けるため同条件に揃える）。

---

## 0. 要約

frontend は第1〜6期リファクタの結果、**規約適合はほぼ完璧**（Feature Indexing 26/26・deep import 0・層逆転 0・800行超過実質1件）。構造の再編は不要。残る本質的欠陥は **1点のみ = 規約を守る機械ガード（ESLint 境界ルール）が存在しない**こと。今日の0違反は人力レビューだけで維持されており、過去には cross-feature import 38件が実在した（第2期監査）。FE7-0 がほぼ全価値を占め、残りは小粒整理。

## 1. 実測スナップショット（2026-07-17）

```bash
# 再計測（着手時に必ず実行）
cd frontend
for d in src/*/; do echo "$d $(find "$d" \( -name '*.ts' -o -name '*.tsx' \) | wc -l) files"; done
for d in src/features/*/; do [ -f "$d/index.ts" ] || echo "NO-INDEX: $d"; done
grep -rEn "from ['\"]@/features/[a-z-]+/" src/app src/components src/hooks src/lib src/utils | grep -v ".test."
```

- src = 10.4万行（features 750 impl files が中核）／liff 659行／line-reserve 4,413行
- features 26個 全てに index.ts。deep import・cross-feature 内部参照・層逆転（components/hooks→features）= **全て0件**
- 800行超: `src/types/generated/models.ts`（3,174・生成物=対象外）と `src/lib/design-tokens.ts`（805・境界値）のみ
- `eslint.config.js`（132行）に境界系ルール（no-restricted-imports 等）**なし**
- 二重ホーム: `src/constants/`（7）と `src/utils/constants/`（1）／`src/lib/`（27）と `src/utils/`（6）の役割境界曖昧（domain helper が両方に散在: `utils/status-helpers` vs `lib/cpm-stage`）
- `src/shared-liff/` を liff（7箇所）・line-reserve（13箇所）が `@/shared-liff/...` alias で**親アプリの src ツリーへ**参照
- `src/contexts/` = 8行1ファイル（AuthContext）

---

## 2. タスク（優先順）

### FE7-0: ESLint 境界ガードの新設【本計画の主目的】
- **作業**: `frontend/eslint.config.js` に以下を追加:
  1. **deep import 禁止**（全域）: `no-restricted-imports` で `patterns: [{ group: ['@/features/*/*'], message: 'feature の外からは @/features/<name>（index.ts）経由で import する。feature 内部は相対 import。' }]`
  2. **層逆転禁止**: flat config の `files: ['src/components/**', 'src/hooks/**', 'src/lib/**', 'src/utils/**']` ブロックで `group: ['@/features', '@/features/*']` を全面禁止
  3. **アプリ境界**: `files: ['liff/src/**', 'line-reserve/src/**']` で `@/features` 参照を禁止（現状0件の維持）
- **注意**: 現状0違反は grep 実測に基づく — ESLint は type-only import 等も検出するため、導入時に未知の違反が出うる。出た違反は「直してからマージ」（ルールを緩めない）。
- **検証**: scoped で `npx eslint src/components src/hooks src/app --no-warn-ignored` ＋ 代表 feature 1個。フル lint（`docker compose exec frontend pnpm lint`）は禁止コマンドのため **USER 実行を依頼**して green 確認。
- **完了条件**: 3 ルールが CI の lint ジョブで効いている（故意の違反を1件書いて RED になることを確認してから削除 = temp-revert 方式）。

### FE7-1: utils/ を lib/ へ統合（共有ヘルパの単一ホーム化）
- **作業**: `src/utils/` の実体6ファイル＋テストを `src/lib/` へ `git mv` し import を機械置換（`@/utils/` → `@/lib/`）。`utils/constants/status-colors.ts` は FE7-2 へ。`utils/format/` はディレクトリごと `lib/format/` へ。
- **恒久規約（実行時に frontend/CLAUDE.md へ追記）**: 「共有ヘルパの置き場は `lib/` 一箇所。`utils/` の新設禁止」
- **検証**: `npx vitest run src/lib` ＋ 移動対象を import している feature の scoped test。
- **完了条件**: `src/utils/` が存在しない。

### FE7-2: constants の単一ホーム化
- **作業**: `src/utils/constants/status-colors.ts` → `src/constants/` へ移動・import 置換。
- **完了条件**: 定数ホームが `src/constants/` 一箇所。

### FE7-3: [判断ゲート付き・任意] shared-liff の配置
- **現状**: 動作・tree-shaking とも問題なし。問題は構造のみ（副アプリが親アプリの src ツリーへ alias 依存）。
- **選択肢**: (A) `frontend/shared-liff/` へ昇格（3アプリ対等の共有層・vite/tsconfig alias 3箇所更新） (B) 現状維持を規約として明文化。
- **判断基準**: liff/line-reserve に今後の機能拡張予定があるなら A、凍結気味なら B。**実害が出るまで B を推奨**（②削除原則 — 動くものを動かさない）。
- **完了条件**: どちらかを frontend/CLAUDE.md に1行明文化。

### FE7-4: [任意] 微小整理
- `src/contexts/`（8行1ファイル）→ `src/lib/` か `src/hooks/` へ吸収しディレクトリ削除。
- `src/lib/design-tokens.ts`（805行）→ 800行規約の境界値。トークン表は分割で可読性が下がるため**例外として容認を明文化**（分割しない）。

## 3. やらないこと（決定済み）

- **features の再編・分割**（master 184 files 含む）— index.ts 境界で閉じており実害なし。分割は「存在すべきでないものの最適化」
- **liff/line-reserve の構造変更** — 両方とも小規模健全
- **generated/models.ts への対処** — 生成物・対象外

## 4. 検証規約（frontend/CLAUDE.md 準拠）

- フル `pnpm lint` / `pnpm test:run` / `pnpm type-check` は自動実行禁止 → scoped は `npx eslint <paths>` / `npx vitest run <path>`。フルゲートは USER に依頼。
