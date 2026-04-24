# FE 締め時間設定 コード規約対応 プロンプト例

チェックリスト仕様: `tmp/FE/fe_closing_settings_check.md`

---

## 使い分け早見表

| フェーズ | 状況 | 使う例 |
|---------|------|--------|
| **スキャン** | スキャンから起票まで一気に終わらせたい | スキャン例A |
| **実装** | タスクを一括で完了させたい | 実装例A |

---

# フェーズ1: スキャン

## スキャン例A: スキャン + 起票（フル実行）

ファイル数が少ないため単一エージェントで実行する。

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/FE/fe_closing_settings_check.md') as f:
    content = f.read()
files = re.findall(r'- (frontend/src/features/closing-settings/\S+\.tsx?)', content)
missing = [f for f in files if not os.path.exists(f)]
print(f'登録ファイル数: {len(files)}')
print(f'欠損: {len(missing)}件')
for f in missing: print(f'  MISSING: {f}')
if not missing: print('全件OK — スキャン開始')
"

欠損ファイルが 1 件でもあった場合はスキャンを中止し、欠損一覧をユーザーに報告して終了せよ。
全件 OK の場合のみ Step 1 に進む。

【Step 1: スキャン + 起票実行】
tmp/FE/fe_closing_settings_check.md を読み込み、完了条件3（起票）まで含めて全工程を実行せよ。

タスクファイルは docs/tasks/open/code-quality/TASK-{番号}-{kebab-case-title}.md に作成する。
タスク番号は docs/tasks/open/code-quality/ と docs/tasks/closed/code-quality/ の
既存ファイル名から最大番号を確認し、その +1 から採番する。

全ファイルを Read してから判定すること。推測での OK/FAIL 出力は禁止。
違反が 0 件だった場合は「違反なし」をユーザーに報告して終了せよ（起票不要）。
起票完了後、新規作成したタスクファイルのパス一覧をユーザーに報告して終了せよ。

ultrathink
```

---

# フェーズ2: 実装

## 実装例A: タスク一括実装

```
docs/tasks/open/code-quality/ の全タスクを優先度順（Critical → High → Medium → Low）に
すべて実装してクローズせよ。

## 実装手順
1. docs/tasks/open/code-quality/ のファイル一覧を取得する
2. 各タスクファイルを Read して「優先度」「対象ファイル」「あるべき姿」を把握する
   - フロントマターの `pattern:` が FA4/FA5/FA6/FA7/FG1/FG2/FG3/FG4 のタスクのみを対象とする
   - `status: partial` のタスクは除外する（partial_note を確認し未実装箇所をユーザーに確認する）
   - 対象タスクが 0 件の場合はその旨を報告して終了せよ
3. 優先度でグルーピングし、Critical から順番に実装する
4. 各タスクについて:
   a. タスクファイルの「あるべき姿」を確認する
   b. 対象ソースファイルを Read して現状コードを確認する
   c. Edit ツールで「あるべき姿」のとおりに修正する（余分な変更を加えない）
5. 全タスク修正後、ユーザーに以下の手動実行を依頼する（自動実行禁止）:
   docker compose exec frontend npm run lint
   docker compose exec frontend npm run test:run
6. ユーザーからの結果を受け取ったら:
   - 両方 PASS → タスクをまとめて 1 コミットし、docs/tasks/closed/code-quality/ に移動する
   - FAIL → エラー内容を確認して修正し、再度手動実行を依頼する。修正してもFAILが続く場合はユーザーに報告して止まる
7. 全タスク完了後、docs/tasks/open/code-quality/ に残っている FA4/FA5/FA6/FA7/FG1/FG2/FG3/FG4 パターンのタスクがないことを確認し、ユーザーに完了報告する

## 禁止事項
- タスクファイルを読まずに実装しない
- 「あるべき姿」以外の箇所を変更しない
- テスト結果未確認のままクローズしない
- テストFAILのままコミットしない
- `status: partial` のタスクをフルクローズしない（partial_note を必ず読むこと）
- 既に closed/ にあるタスクを再実装しない

ultrathink
```
