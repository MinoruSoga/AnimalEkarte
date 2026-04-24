# FE マスタ系 コード規約対応 プロンプト例

チェックリスト仕様: `tmp/FE/fe_master_check.md`

---

## 使い分け早見表

| フェーズ | 状況 | 使う例 |
|---------|------|--------|
| **スキャン** | 初回・規模感を把握したい | スキャン例A |
| **スキャン** | スキャンから起票まで一気に終わらせたい | スキャン例B |
| **スキャン** | 修正後に特定パターンだけ再確認したい | スキャン例C |
| **実装** | Critical を先に潰したい | 実装例A |
| **実装** | 多数タスクを効率よく並列処理したい | 実装例B |
| **実装** | 影響範囲が大きい1件を慎重に実装したい | 実装例C |
| **実装** | 全タスクをまとめて完了させたい | 実装例D |

---

# フェーズ1: スキャン

## スキャン例A: スキャンのみ（起票なし）— 最初の実行に使う

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/FE/fe_master_check.md') as f:
    content = f.read()
files = re.findall(r'- (frontend/src/features/master/\S+\.tsx?)', content)
missing = [f for f in files if not os.path.exists(f)]
print(f'登録ファイル数: {len(files)}')
print(f'欠損: {len(missing)}件')
for f in missing: print(f'  MISSING: {f}')
if not missing: print('全件OK — スキャン開始')
"

欠損ファイルが 1 件でもあった場合はスキャンを中止し、欠損一覧をユーザーに報告して終了せよ。
全件 OK の場合のみ Step 1 に進む。

【Step 1: スキャン実行（起票なし）】
tmp/FE/fe_master_check.md を読み込み、以下の指示に従ってスキャンを実行せよ。

【完了条件の変更】
完了条件の 3（docs/tasks/への起票）は実行しない。
1と2（PASS/FAIL表 + 違反サマリ）の出力のみで完了とする。

ultrathink

use AgentTeams with the following teams running in parallel:
- Team-API: FA1/FA2/FA3/FA4/FA5/FA6/FA7 を tmp/FE/fe_master_check.md の「API」リストに対して検査
- Team-Routes: FR1/FR2/FR3/FR4/FR5/FG1/FG2/FG3 を tmp/FE/fe_master_check.md の「Routes」リストに対して検査
- Team-Components: FG1/FG2/FG3 を tmp/FE/fe_master_check.md の「Components」リストに対して検査

各チームは担当ファイルを全件 Read してから判定すること。
推測での OK/FAIL 出力は禁止。
```

---

## スキャン例B: スキャン + 起票（フル実行）

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/FE/fe_master_check.md') as f:
    content = f.read()
files = re.findall(r'- (frontend/src/features/master/\S+\.tsx?)', content)
missing = [f for f in files if not os.path.exists(f)]
print(f'登録ファイル数: {len(files)}')
print(f'欠損: {len(missing)}件')
for f in missing: print(f'  MISSING: {f}')
if not missing: print('全件OK — スキャン開始')
"

欠損ファイルが 1 件でもあった場合はスキャンを中止し、欠損一覧をユーザーに報告して終了せよ。
全件 OK の場合のみ Step 1 に進む。

【Step 1: スキャン + 起票実行】
tmp/FE/fe_master_check.md を読み込み、完了条件3（起票）まで含めて全工程を実行せよ。

タスクファイルは docs/tasks/open/code-quality/TASK-{番号}-{kebab-case-title}.md に作成する。
タスク番号は docs/tasks/open/code-quality/ と docs/tasks/closed/code-quality/ の
既存ファイル名から最大番号を確認し、その +1 から採番する。

ultrathink

use AgentTeams with the following teams running in parallel:
- Team-API: FA1/FA2/FA3/FA4/FA5/FA6/FA7 を tmp/FE/fe_master_check.md の「API」リストに対して検査
- Team-Routes: FR1/FR2/FR3/FR4/FR5/FG1/FG2/FG3 を tmp/FE/fe_master_check.md の「Routes」リストに対して検査
- Team-Components: FG1/FG2/FG3 を tmp/FE/fe_master_check.md の「Components」リストに対して検査

各チームは担当ファイルを全件 Read してから判定すること。
推測での OK/FAIL 出力は禁止。
全チーム完了後、違反サマリを集約して既存タスクとの重複チェックを行い、
未起票の違反のみを新規タスクとして起票する。
違反が 0 件だった場合は「違反なし」をユーザーに報告して終了せよ（起票不要）。
起票完了後、新規作成したタスクファイルのパス一覧をユーザーに報告して終了せよ。
```

---

## スキャン例C: 特定チームのみ再スキャン（修正後の確認）

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/FE/fe_master_check.md') as f:
    content = f.read()
files = re.findall(r'- (frontend/src/features/master/\S+\.tsx?)', content)
missing = [f for f in files if not os.path.exists(f)]
print(f'登録ファイル数: {len(files)}')
print(f'欠損: {len(missing)}件')
for f in missing: print(f'  MISSING: {f}')
if not missing: print('全件OK — スキャン開始')
"

欠損ファイルが 1 件でもあった場合はスキャンを中止し、欠損一覧をユーザーに報告して終了せよ。
全件 OK の場合のみ Step 1 に進む。

【Step 1: 再スキャン実行】
tmp/FE/fe_master_check.md のチェックリストとファイルリストを参照し、
以下のチームの担当スコープのみ再スキャンせよ。

対象チーム: Team-API（FA1〜FA7のみ）
対象ファイル: tmp/FE/fe_master_check.md の「API」リスト

完了条件: PASS/FAIL 表の出力のみ（起票不要）

ultrathink
```

---

# フェーズ2: 実装

## 実装の鉄則（全例共通）

```
1. 必ずタスクファイルを Read してから実装する（タイトルだけで判断しない）
2. 「あるべき姿」セクションのコードのみを実装する（それ以上の修正禁止）
3. 実装後、ユーザーに以下の手動実行を依頼する（自動実行禁止）:
   docker compose exec frontend npm run lint
   docker compose exec frontend npm run test:run
4. ユーザーからの結果を受け取ったら:
   - 両方 PASS → タスクをまとめて 1 コミットし、docs/tasks/closed/code-quality/ に移動する
   - FAIL → エラー内容を確認して修正し、再度手動実行を依頼する
5. テストFAILのままコミットしない・クローズしない
```

---

## 実装例A: 優先度 Critical のタスクのみ実装

```
docs/tasks/open/code-quality/ の全タスクファイルを Read し、
優先度が「Critical」のタスクのみを抽出して実装せよ。

## 実装手順
1. docs/tasks/open/code-quality/ のファイル一覧を取得する
2. 各タスクファイルを Read して「優先度」を確認し、Critical のみをリストアップする
   - Critical タスクが 0 件の場合はその旨を報告して終了せよ
3. 各 Critical タスクについて:
   a. タスクファイルの「あるべき姿」を確認する
   b. 対象ソースファイルを Read して現状コードを確認する
   c. Edit ツールで「あるべき姿」のとおりに修正する（余分な変更を加えない）
4. 全 Critical タスク修正後、ユーザーに以下の手動実行を依頼する（自動実行禁止）:
   docker compose exec frontend npm run lint
   docker compose exec frontend npm run test:run
5. ユーザーからの結果を受け取ったら:
   - 両方 PASS → Critical タスクをまとめて 1 コミットし、docs/tasks/closed/code-quality/ に移動する
   - FAIL → エラー内容を確認して修正し、再度手動実行を依頼する。修正してもFAILが続く場合はユーザーに報告して止まる

## 禁止事項
- タスクファイルを読まずに実装しない
- 「あるべき姿」以外の箇所を変更しない
- テスト結果未確認のままクローズしない
- テストFAILのままコミットしない

ultrathink
```

---

## 実装例B: パターン別に AgentTeams で並列実装

ファイル競合が起きないよう、担当ファイルが重複しないチーム編成にする。

```
docs/tasks/open/code-quality/ の全タスクファイルを Read し、
以下のチーム編成ルールに従って AgentTeams で並列実装せよ。

## 事前準備（実装前に必ず実施）
1. docs/tasks/open/code-quality/ の全ファイルを Read して対象タスクを把握する
2. フロントマターの `pattern:` フィールドが FA/FR/FG 系（FA1〜FA7, FR1〜FR5, FG1〜FG3）のタスクのみを対象とする
   - pattern が上記以外または `status: partial` のタスクは除外する
3. 各タスクの「対象ファイル」を確認し、同一ファイルへの変更が重複するタスクを
   同一チームに割り当てる（チーム間のファイル競合を防ぐ）
4. 全タスクが以下のいずれかのチームに割り当てられたことを確認する。
   どのチームにも当てはまらないタスクがある場合はユーザーに報告して確認を取ってから進む

## チーム編成ルール（対象ファイルの種別で分割）

### Team-API-Transform
担当: api/ ファイルの transform 関数・ドメイン型導出（FA1/FA2）修正タスク全件

### Team-API-Hooks
担当: api/ ファイルのクエリキー・フック命名・onError・staleTime・リクエスト型（FA3/FA4/FA5/FA6/FA7）修正タスク全件

### Team-Routes-CRUD
担当: routes/ ファイルの useMasterCRUD / useMasterSave / usePermission（FR1/FR2/FR3）修正タスク全件

### Team-Routes-Perf
担当: routes/ ファイルの memo() / lazy initializer（FR4/FR5）修正タスク全件

### Team-Style
担当: routes/ および components/ ファイルのデザイントークン・条件レンダー・any 型（FG1/FG2/FG3）修正タスク全件

## 実装手順（各チーム共通）
1. 担当タスクファイルを全件 Read して「あるべき姿」を把握する
   - 担当タスクが 0 件の場合はその旨を報告して終了せよ（何もしない）
2. 対象ソースファイルを Read して現状を確認する
3. Edit ツールで修正する（あるべき姿以外の変更禁止）
4. 全タスク修正完了後、ユーザーに以下の手動実行を依頼する（自動実行禁止）:
   docker compose exec frontend npm run lint
   docker compose exec frontend npm run test:run
5. ユーザーからの結果を受け取ったら:
   - 両方 PASS → 担当タスク全件をまとめて 1 コミットし、docs/tasks/closed/code-quality/ に移動する
   - FAIL → エラー内容を確認して修正し、再度手動実行を依頼する。修正してもFAILが続く場合はユーザーに報告して止まる
6. 全チーム完了後、docs/tasks/open/code-quality/ に残っている FA/FR/FG パターンのタスクがないことを確認し、ユーザーに完了報告する

## 禁止事項
- タスクファイルを読まずに実装しない
- 担当外のファイルを変更しない（チーム間の競合防止）
- あるべき姿以外の箇所を変更しない
- テスト結果未確認のままクローズしない
- テストFAILのままコミットしない
- `status: partial` のタスクをフルクローズしない（partial_note を必ず読むこと）
- 既に closed/ にあるタスクを再実装しない

ultrathink

use AgentTeams.
```

---

## 実装例C: 単一タスクの修正（慎重に進めたいとき）

`{タスクファイルパス}` を実装したいタスクのパスに置き換えて使う。

```
以下のタスクを1件実装せよ。

## 対象タスク
docs/tasks/open/code-quality/{タスクファイル名}.md

## 実装手順
1. タスクファイルを Read して「問題概要」「あるべき姿」「対象ファイル」を確認する
2. 対象ファイルを Read して現状コードを確認する
3. Edit ツールで「あるべき姿」のとおりに修正する（余分な変更を加えない）
4. ユーザーに以下の手動実行を依頼する（自動実行禁止）:
   docker compose exec frontend npm run lint
   docker compose exec frontend npm run test:run
5. ユーザーからの結果を受け取ったら:
   - 両方 PASS → 1 コミットし、タスクファイルを docs/tasks/closed/code-quality/ に移動する
   - FAIL → エラー内容を確認して修正し、再度手動実行を依頼する

## 禁止事項
- あるべき姿以外の箇所を変更しない
- 関連する他のファイルを勝手に変更しない
- テストFAILのままコミットしない

ultrathink
```

---

## 実装例D: 全 open タスクを優先度順に一括処理

```
docs/tasks/open/code-quality/ の全タスクを優先度順（Critical → High → Medium → Low）に
すべて実装してクローズせよ。

## 実装手順
1. docs/tasks/open/code-quality/ のファイル一覧を取得する
2. 各タスクファイルを Read して「優先度」「対象ファイル」「あるべき姿」を把握する
   - フロントマターの `pattern:` が FA/FR/FG 系（FA1〜FA7, FR1〜FR5, FG1〜FG3）のタスクのみを対象とする
   - `status: partial` のタスクは除外する（partial_note を確認し未実装箇所をユーザーに確認する）
3. 対象タスクが 0 件の場合はその旨をユーザーに報告して終了せよ
4. 優先度でグルーピングし、Critical から順番に実装する
4. 同一優先度のタスクは担当ファイルが重複しないものを AgentTeams で並列実装する
5. 各グループの全タスク修正完了後、ユーザーに以下の手動実行を依頼する（自動実行禁止）:
   docker compose exec frontend npm run lint
   docker compose exec frontend npm run test:run
6. ユーザーからの結果を受け取ったら:
   - 両方 PASS → そのグループのタスクをまとめて 1 コミットし、docs/tasks/closed/code-quality/ に移動してから次のグループへ進む
   - FAIL → エラー内容を確認して修正し、再度手動実行を依頼する。修正してもFAILが続く場合はユーザーに報告して止まる
7. 全グループ完了後、docs/tasks/open/code-quality/ に残っている FA/FR/FG パターンのタスクがないことを確認し、ユーザーに完了報告する

## 並列実行の制約
- 同一ファイルへの変更が発生するタスクは順番に実行する（競合防止）
- 異なるファイルへの変更は並列実行してよい

## 禁止事項
- タスクに書かれた「あるべき姿」以外の修正を行わない
- テスト結果未確認のままタスクをクローズしない
- テストFAILのままコミットしない
- `status: partial` のタスクをフルクローズしない（partial_note を必ず読むこと）
- 既に closed/ にあるタスクを再実装しない

ultrathink

use AgentTeams.
```
