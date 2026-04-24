# BEマスタ コード規約対応 プロンプト例

チェックリスト仕様: `tmp/be_master_scan_prompt.md`

---

## 使い分け早見表

| フェーズ | 状況 | 使う例 |
|---------|------|--------|
| **スキャン** | 初回・規模感を把握したい | スキャン例1 |
| **スキャン** | スキャンから起票まで一気に終わらせたい | スキャン例2 |
| **スキャン** | 修正後に特定パターンだけ再確認したい | スキャン例3 |
| **実装** | Critical を先に潰したい | 実装例1 |
| **実装** | 多数タスクを効率よく並列処理したい | 実装例2 |
| **実装** | 影響範囲が大きい1件を慎重に実装したい | 実装例3 |
| **実装** | 全タスクをまとめて完了させたい | 実装例4 |

---

# フェーズ1: スキャン

## スキャン例1: スキャンのみ（起票なし）— 最初の実行に使う

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/be_master_scan_prompt.md') as f:
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

【Step 1: スキャン実行】
tmp/be_master_scan_prompt.md を読み込み、以下の指示に従ってスキャンを実行せよ。

【完了条件の変更】
完了条件の 3（docs/tasks/への起票）は実行しない。
1と2（PASS/FAIL表 + 違反サマリ）の出力のみで完了とする。

ultrathink

use AgentTeams with the following teams running in parallel:
- Team-Service: P1/P8/P10/P11/P13/P17 を tmp/be_master_scan_prompt.md の「Service」リストに対して検査
- Team-Repository-Master: P2/P3/P4/P9/P16 を tmp/be_master_scan_prompt.md の「Repository - マスタ系」リストに対して検査
- Team-Repository-Preload: P3 を tmp/be_master_scan_prompt.md の「Repository - 非マスタ系」リストに対して検査
- Team-Handler: P7/P12/P14/P15/P18 を tmp/be_master_scan_prompt.md の「Handler」リストに対して検査
- Team-Routes: P5/P6 を tmp/be_master_scan_prompt.md の「Routes」リストに対して検査

各チームは担当ファイルを全件 Read してから判定すること。
推測での OK/FAIL 出力は禁止。
```

---

## スキャン例2: スキャン + 起票（フル実行）

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/be_master_scan_prompt.md') as f:
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
tmp/be_master_scan_prompt.md を読み込み、完了条件3（起票）まで含めて全工程を実行せよ。

タスクファイルは docs/tasks/open/code-quality/TASK-{番号}-{kebab-case-title}.md に作成する。
タスク番号は docs/tasks/open/code-quality/ と docs/tasks/closed/code-quality/ の
既存ファイル名から最大番号を確認し、その +1 から採番する。

ultrathink

use AgentTeams with the following teams running in parallel:
- Team-Service: P1/P8/P10/P11/P13/P17 を tmp/be_master_scan_prompt.md の「Service」リストに対して検査
- Team-Repository-Master: P2/P3/P4/P9/P16 を tmp/be_master_scan_prompt.md の「Repository - マスタ系」リストに対して検査
- Team-Repository-Preload: P3 を tmp/be_master_scan_prompt.md の「Repository - 非マスタ系」リストに対して検査
- Team-Handler: P7/P12/P14/P15/P18 を tmp/be_master_scan_prompt.md の「Handler」リストに対して検査
- Team-Routes: P5/P6 を tmp/be_master_scan_prompt.md の「Routes」リストに対して検査

各チームは担当ファイルを全件 Read してから判定すること。
推測での OK/FAIL 出力は禁止。
全チーム完了後、違反サマリを集約して既存タスクとの重複チェックを行い、
未起票の違反のみを新規タスクとして起票する。
```

---

## スキャン例3: 特定チームのみ再スキャン（修正後の確認）

```
【Step 0: 事前整合性チェック（必須・スキャン前に必ず実行）】
以下の Python スクリプトを Bash ツールで実行し、ドキュメントに記載された全ファイルが存在するか確認せよ。

python3 -c "
import re, os
with open('tmp/be_master_scan_prompt.md') as f:
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
tmp/be_master_scan_prompt.md のチェックリストとファイルリストを参照し、
以下のチームの担当スコープのみ再スキャンせよ。

対象チーム: Team-Repository-Preload（P3のみ）
対象ファイル: tmp/be_master_scan_prompt.md の「Repository - 非マスタ系」リスト

完了条件: PASS/FAIL 表の出力のみ（起票不要）

ultrathink
```

---

# フェーズ2: 実装

## 実装の鉄則（全例共通）

```
1. 必ずタスクファイルを Read してから実装する（タイトルだけで判断しない）
2. 「あるべき姿」セクションのコードのみを実装する（それ以上の修正禁止）
3. 実装後に docker compose exec backend go test ./backend/internal/... を実行してパスを確認する
4. テストパス後にタスクファイルを docs/tasks/open/ → docs/tasks/closed/ に移動する
5. 1コミットに複数タスクをまとめてよい（同一ファイルへの変更は1コミット推奨）
```

---

## 実装例1: 優先度 Critical のタスクのみ実装

```
docs/tasks/open/code-quality/ の全タスクファイルを Read し、
優先度が「Critical」のタスクのみを抽出して実装せよ。

## 実装手順（各タスクに対して繰り返す）
1. docs/tasks/open/code-quality/ のファイル一覧を取得する
2. 各タスクファイルを Read して「優先度」を確認し、Critical のみをリストアップする
3. 各 Critical タスクについて:
   a. タスクファイルの「あるべき姿」を確認する
   b. 対象ソースファイルを Read して現状コードを確認する
   c. Edit ツールで「あるべき姿」のとおりに修正する（余分な変更を加えない）
4. 全 Critical タスク修正後、docker compose exec backend go test ./backend/internal/... を実行する
5. テストが PASS したら git commit する
6. 完了したタスクファイルを docs/tasks/closed/code-quality/ に移動する

## 禁止事項
- タスクファイルを読まずに実装しない
- 「あるべき姿」以外の箇所を変更しない
- テスト未確認のままクローズしない

ultrathink
```

---

## 実装例2: パターン別に AgentTeams で並列実装

ファイル競合が起きないよう、担当ファイルが重複しないチーム編成にする。

```
docs/tasks/open/code-quality/ の全タスクファイルを Read し、
以下のチーム編成ルールに従って AgentTeams で並列実装せよ。

## 事前準備（実装前に必ず実施）
1. docs/tasks/open/code-quality/ の全ファイルを Read して対象タスクを把握する
2. 各タスクの「対象ファイル」を確認し、同一ファイルへの変更が重複するタスクを
   同一チームに割り当てる（チーム間のファイル競合を防ぐ）

## チーム編成ルール（対象ファイルの種別で分割）

### Team-Routes
担当: handler/*_routes.go に変更が発生するタスク全件

### Team-Repository-DeletedAt
担当: repository の CountUsage/CountBy* メソッドの deleted_at 修正タスク全件

### Team-Repository-Scope
担当: repository の clinicScope 欠落修正タスク全件

### Team-Service-FindByID
担当: service の Delete/Update で FindByID 前置が必要なタスク全件

### Team-Service-Error
担当: service のエラー処理（apperrors.Wrap / slog.ErrorContext）修正タスク全件

### Team-Handler
担当: handler のレスポンス変換・ShouldBindJSON・Location ヘッダ修正タスク全件

### Team-Ordering
担当: const/buildFunc 定義順序の修正タスク全件

## 実装手順（各チーム共通）
1. 担当タスクファイルを全件 Read して「あるべき姿」を把握する
2. 対象ソースファイルを Read して現状を確認する
3. Edit ツールで修正する（あるべき姿以外の変更禁止）
4. 全タスク修正完了後に docker compose exec backend go test ./backend/internal/... を実行
5. PASS したら git commit し、タスクファイルを docs/tasks/closed/code-quality/ に移動する

## 禁止事項
- タスクファイルを読まずに実装しない
- 担当外のファイルを変更しない（チーム間の競合防止）
- あるべき姿以外の箇所を変更しない
- テスト未確認のままクローズしない

ultrathink

use AgentTeams.
```

---

## 実装例3: 単一タスクの修正（慎重に進めたいとき）

`{タスクファイルパス}` を実装したいタスクのパスに置き換えて使う。

```
以下のタスクを1件実装せよ。

## 対象タスク
docs/tasks/open/code-quality/{タスクファイル名}.md

## 実装手順
1. タスクファイルを Read して「問題概要」「あるべき姿」「対象ファイル」を確認する
2. 対象ファイルを Read して現状コードを確認する
3. Edit ツールで「あるべき姿」のとおりに修正する（余分な変更を加えない）
4. docker compose exec backend go test ./backend/internal/... を実行する
5. PASS したら git commit する
6. タスクファイルを docs/tasks/closed/code-quality/ に移動する

## 禁止事項
- あるべき姿以外の箇所を変更しない
- 関連する他のファイルを勝手に変更しない

ultrathink
```

---

## 実装例4: 全 open タスクを優先度順に一括処理

```
docs/tasks/open/code-quality/ の全タスクを優先度順（Critical → High → Medium → Low）に
すべて実装してクローズせよ。

## 実装手順
1. docs/tasks/open/code-quality/ のファイル一覧を取得する
2. 各タスクファイルを Read して「優先度」「対象ファイル」「あるべき姿」を把握する
3. 優先度でグルーピングし、Critical から順番に実装する
4. 同一優先度のタスクは担当ファイルが重複しないものを AgentTeams で並列実装する
5. 各グループ実装後に docker compose exec backend go test ./backend/internal/... を実行する
6. テスト PASS を確認してから次の優先度グループに進む
7. 完了したタスクは docs/tasks/closed/code-quality/ に移動する

## 並列実行の制約
- 同一ファイルへの変更が発生するタスクは順番に実行する（競合防止）
- 異なるファイルへの変更は並列実行してよい

## 禁止事項
- タスクに書かれた「あるべき姿」以外の修正を行わない
- テスト未確認のままタスクをクローズしない
- 既に closed/ にあるタスクを再実装しない

ultrathink

use AgentTeams.
```
