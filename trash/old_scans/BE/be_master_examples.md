# BE マスタ系 コード規約対応 プロンプト例

チェックリスト仕様: `tmp/BE/be_master_check.md`

---

## 使い分け早見表

| フェーズ | 状況 | 使う例 |
|---------|------|--------|
| **スキャン** | スキャンから起票まで一気に終わらせたい | スキャン例A |
| **スキャン** | 修正後に特定パターンだけ再確認したい | スキャン例B |
| **実装** | 多数タスクを効率よく並列処理したい | 実装例A |
| **実装** | 全タスクをまとめて完了させたい | 実装例B |

---

# フェーズ1: スキャン

## スキャン例A: スキャン + 起票（フル実行）

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/BE/be_master_check.md') as f:
    content = f.read()
files = re.findall(r'- (backend/internal/\S+\.go)', content)
missing = [f for f in files if not os.path.exists(f)]
print(f'登録ファイル数: {len(files)}')
print(f'欠損: {len(missing)}件')
for f in missing: print(f'  MISSING: {f}')
if not missing: print('全件OK — スキャン開始')
"

欠損ファイルが 1 件でもあった場合はスキャンを中止し、欠損一覧をユーザーに報告して終了せよ。
全件 OK の場合のみ Step 1 に進む。

【Step 1: スキャン + 起票実行】
tmp/BE/be_master_check.md を読み込み、完了条件3（起票）まで含めて全工程を実行せよ。

タスクファイルは docs/tasks/open/code-quality/TASK-{番号}-{kebab-case-title}.md に作成する。
タスク番号は docs/tasks/open/code-quality/ と docs/tasks/closed/code-quality/ の
既存ファイル名から最大番号を確認し、その +1 から採番する。

ultrathink

use AgentTeams with the following teams running in parallel:
- Team-Service: P1/P8/P10/P11/P13/P17 を tmp/BE/be_master_check.md の「Service」リストに対して検査
- Team-Repository-Master: P2/P3/P4/P9/P16 を tmp/BE/be_master_check.md の「Repository - マスタ系」リストに対して検査
- Team-Repository-Preload: P3 を tmp/BE/be_master_check.md の「Repository - 非マスタ系」リストに対して検査
- Team-Handler: P7/P12/P14/P15/P18 を tmp/BE/be_master_check.md の「Handler」リストに対して検査
- Team-Routes: P5/P6 を tmp/BE/be_master_check.md の「Routes」リストに対して検査

各チームは担当ファイルを全件 Read してから判定すること。
推測での OK/FAIL 出力は禁止。
全チーム完了後、違反サマリを集約して既存タスクとの重複チェックを行い、
未起票の違反のみを新規タスクとして起票する。
違反が 0 件だった場合は「違反なし」をユーザーに報告して終了せよ（起票不要）。
起票完了後、新規作成したタスクファイルのパス一覧をユーザーに報告して終了せよ。
```

---

## スキャン例B: 特定チームのみ再スキャン（修正後の確認）

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/BE/be_master_check.md') as f:
    content = f.read()
files = re.findall(r'- (backend/internal/\S+\.go)', content)
missing = [f for f in files if not os.path.exists(f)]
print(f'登録ファイル数: {len(files)}')
print(f'欠損: {len(missing)}件')
for f in missing: print(f'  MISSING: {f}')
if not missing: print('全件OK — スキャン開始')
"

欠損ファイルが 1 件でもあった場合はスキャンを中止し、欠損一覧をユーザーに報告して終了せよ。
全件 OK の場合のみ Step 1 に進む。

【Step 1: 再スキャン実行】
tmp/BE/be_master_check.md のチェックリストとファイルリストを参照し、
以下のチームの担当スコープのみ再スキャンせよ。

対象チーム: Team-Repository-Preload（P3のみ）
対象ファイル: tmp/BE/be_master_check.md の「Repository - 非マスタ系」リスト

完了条件: PASS/FAIL 表の出力のみ（起票不要）

ultrathink
```

---

# フェーズ2: 実装

## 実装例A: パターン別に AgentTeams で並列実装

ファイル競合が起きないよう、担当ファイルが重複しないチーム編成にする。

```
docs/tasks/open/code-quality/ の全タスクファイルを Read し、
以下のチーム編成ルールに従って AgentTeams で並列実装せよ。

## 事前準備（実装前に必ず実施）
1. docs/tasks/open/code-quality/ の全ファイルを Read して対象タスクを把握する
2. フロントマターの `pattern:` フィールドが P1〜P18 のいずれかであるタスクのみを対象とする
   - pattern が P1〜P18 以外（例: BUG-xxx）または `status: partial` のタスクは除外する
3. 各タスクの「対象ファイル」を確認し、同一ファイルへの変更が重複するタスクを
   同一チームに割り当てる（チーム間のファイル競合を防ぐ）
4. 全タスクが以下のいずれかのチームに割り当てられたことを確認する。
   どのチームにも当てはまらないタスクがある場合はユーザーに報告して確認を取ってから進む

## チーム編成ルール（対象ファイルの種別で分割）

### Team-Routes
担当: P5/P6 — handler/*_handler.go の Register*Routes 関数（パーミッション設定）修正タスク全件

### Team-Repository-DeletedAt
担当: P2 — repository の CountUsage/CountBy* メソッドの deleted_at IS NULL 修正タスク全件

### Team-Repository-Preload
担当: P3 — repository の Preload deleted_at IS NULL 修正タスク全件

### Team-Repository-Scope
担当: P4 — repository の clinicScope 欠落修正タスク全件

### Team-Repository-Naming
担当: P9/P16 — repository の apperrors.FromGORM 未使用・メソッド名統一 修正タスク全件

### Team-Service-FindByID
担当: P1 — service の Delete/Update で FindByID 前置が必要なタスク全件

### Team-Service-Error
担当: P8/P10/P11 — service の apperrors.Wrap・FK依存チェック・slog.ErrorContext 修正タスク全件

### Team-Service-Naming
担当: P13/P17 — service の const/buildFunc 定義順序・Input 構造体命名統一 修正タスク全件

### Team-Handler
担当: P7/P12/P14/P15 — handler のレスポンス変換・ShouldBindJSON・repository直接呼出禁止・Location ヘッダ修正タスク全件

### Team-Handler-Naming
担当: P18 — handler の toXxxResponse 関数名統一 修正タスク全件

## 実装手順（各チーム共通）
1. 担当タスクファイルを全件 Read して「あるべき姿」を把握する
   - 担当タスクが 0 件の場合はその旨を報告して終了せよ（何もしない）
2. 対象ソースファイルを Read して現状を確認する
3. Edit ツールで修正する（あるべき姿以外の変更禁止）
4. 全タスク修正完了後、ユーザーに以下の手動実行を依頼する（自動実行禁止）:
   docker compose exec backend go test ./backend/internal/...
5. ユーザーからテスト結果を受け取ったら:
   - PASS → 担当タスク全件をまとめて 1 コミットし、タスクファイルを docs/tasks/closed/code-quality/ に移動する
   - FAIL → エラーログを確認して修正し、再度手動実行を依頼する。修正してもFAILが続く場合はユーザーに報告して止まる

## 禁止事項
- タスクファイルを読まずに実装しない
- 担当外のファイルを変更しない（チーム間の競合防止）
- あるべき姿以外の箇所を変更しない
- テスト結果未確認のままクローズしない
- テストFAILのままコミットしない

ultrathink

use AgentTeams.
```

---

## 実装例B: 全 open タスクを優先度順に一括処理

```
docs/tasks/open/code-quality/ の全タスクを優先度順（Critical → High → Medium → Low）に
すべて実装してクローズせよ。

## 実装手順
1. docs/tasks/open/code-quality/ のファイル一覧を取得する
2. 各タスクファイルを Read して「優先度」「対象ファイル」「あるべき姿」を把握する
   - フロントマターの `pattern:` が P1〜P18 のタスクのみを対象とする
   - `status: partial` のタスクは除外する（「partial_note」を確認し未実装箇所がある場合はユーザーに確認する）
3. 対象タスクが 0 件の場合はその旨をユーザーに報告して終了せよ
4. 優先度でグルーピングし、Critical から順番に実装する
5. 同一優先度のタスクは担当ファイルが重複しないものを AgentTeams で並列実装する
6. 各グループの全タスク修正完了後、ユーザーに以下の手動実行を依頼する（自動実行禁止）:
   docker compose exec backend go test ./backend/internal/...
7. ユーザーからテスト結果を受け取ったら:
   - PASS → そのグループのタスクをまとめて 1 コミットし、docs/tasks/closed/code-quality/ に移動してから次のグループへ進む
   - FAIL → エラーログを確認して修正し、再度手動実行を依頼する。修正してもFAILが続く場合はユーザーに報告して止まる
8. 全グループ完了後、docs/tasks/open/code-quality/ に残っている P1〜P18 タスクがないことを確認し、ユーザーに完了報告する

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
