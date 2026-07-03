# Commit準備計画（本タスク限定・並行セッション変更を巻き込まない）

> **重要**: `git add -A` / `git add .` は使用しないこと。working treeには本タスクと無関係な
> 並行セッションの変更が大量に混在している（2026-07-03 04:xx時点の実測で `git status --short | wc -l` = 453。
> 前回計測時421から32件増加しており並行セッションは現在も進行中と確認できる。この数値は
> 常に変動するため固定値として信用せず、**commit直前に必ず再計測すること**）。本タスク対象は
> 下記5ファイルのみ。個別パス指定でstageすること。

## 対象ファイル（stage対象・5件）

| ファイル | 状態 | 内容 |
|---|---|---|
| `docs/be-refactor-followup-status.md` | 既存（untracked）・編集 | 見出しをPhase A-1〜A-5のローカル作業IDへ置換 |
| `docs/be-refactor-phase-a-report.md` | 新規 | corrected Phase A report（エグゼクティブサマリー） |
| `docs/be-refactor-issue-drafts.md` | 新規 | GitHub Issue起票ドラフト3件（未作成） |
| `docs/be-refactor-commit-prep.md` | 新規 | 本ファイル |
| `docs/be-refactor-integration-plan.md` | 新規 | BE-refactor.md統合計画 |

stageコマンド案（実行はユーザー承認後）:
```
git add docs/be-refactor-followup-status.md \
        docs/be-refactor-phase-a-report.md \
        docs/be-refactor-issue-drafts.md \
        docs/be-refactor-commit-prep.md \
        docs/be-refactor-integration-plan.md
```

## 除外対象（stageしない・上記5件以外の全て）

`git status --short` の対象5件を除く**全て**が並行セッションの作業であり、
本タスクのcommitには含めない。主な内訳（サンプル）:

- `BE-refactor.md`（並行セッション編集中・**明示的に除外**。理由は下記「BE-refactor.md統合計画」参照）
- `backend/internal/handler/*_test.go` 多数（並行リファクタセッションのテスト更新）
- `.github/workflows/ci.yml`、`backend/.golangci.yml`、`backend/go.mod`、`CLAUDE.md`、`.claude/settings.json` 等の設定変更
- その他 backend/frontend 全域の変更

**確認方法**: commit前に `git status --short -- docs/be-refactor-followup-status.md docs/be-refactor-phase-a-report.md docs/be-refactor-issue-drafts.md docs/be-refactor-commit-prep.md docs/be-refactor-integration-plan.md` を実行し、意図した5件のみが表示されることを確認してからstageする。

## commit message案

```
docs: Phase A follow-up をローカル作業ID(Phase A-1〜A-5)に整理しIssue化/統合準備を追加

架空Issue番号(#216/#219/#221/#223/#225a)は gh issue view/list で実在しないことを再確認
(直近Issue最大値は#215)。followup-status.md の見出しをローカル作業IDへ置換し、
corrected report・Issue起票ドラフト・commit準備・BE-refactor.md統合計画を追加。
BE-refactor.md本体は並行セッション編集中のため対象外(統合計画のみ作成、直接編集なし)。
```

## 事前検証コマンド案（scoped・doc-onlyにつき軽量）

`.claude/CLAUDE.md` の「文書のみの変更は検証省略可」に従い、フルテスト/lint/buildは不要。
軽量な整合性チェックのみ推奨:

```
# 1. 意図した5ファイルのみが差分に含まれることを確認
git status --short -- docs/be-refactor-followup-status.md docs/be-refactor-phase-a-report.md \
  docs/be-refactor-issue-drafts.md docs/be-refactor-commit-prep.md docs/be-refactor-integration-plan.md

# 2. 架空Issue番号が「非実在/旧ラベル」等の限定なしで裸表記されていないか確認
#    (注意: 単純な見出し正規表現は誤検知する。例えば表ヘッダーが「旧ラベル（非実在・参考のみ）」と
#    宣言し各行が「旧 #219 相当」とだけ書く形式や、複数行にまたがる説明文は単一行grepでは
#    安全側でも引っかかる。このgrep結果は一次スクリーニングに留め、最終判断は該当行の前後
#    文脈を目視確認すること — 本タスクではこの目視確認まで実施し全件PASSを確認済み)
grep -n "#21[6-9]\|#22[135]" docs/be-refactor-followup-status.md docs/be-refactor-phase-a-report.md \
  docs/be-refactor-issue-drafts.md docs/be-refactor-commit-prep.md docs/be-refactor-integration-plan.md \
  | grep -vE "非実在|旧.{0,6}相当|実在しない|架空"
# → ヒットした行は「表ヘッダーで非実在が宣言済みの表内行」等の既知の安全パターンであることを
#   目視で確認すること（本タスクで確認済み・詳細は corrected Phase A report 参照）

# 3. BE-refactor.mdが除外リストに入っており、stage対象に含まれていないことを確認
git diff --cached --stat -- BE-refactor.md
# → commit前は空、stage後も空であることを確認（誤ってstageされていないことの担保）
```

## 未実施事項の明示

本ファイル作成時点で `git add` / `git commit` / `git push` / `gh issue create` / `gh pr create` は
**一度も実行していない**。上記はあくまで承認後に実行する手順のドラフト。
