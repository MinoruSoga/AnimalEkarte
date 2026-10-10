# Worktree Runtime Isolation (Agent Mandatory)

複数 worktree でローカル Docker スタックを**並行起動**するときの必須ルール。
git 操作の安全性は [git-worktree-safety.md](git-worktree-safety.md) を参照。こちらは
**実行時資源（ポート・コンテナ・volume・`.env.local`）**の規約である。

## 背景: 何が衝突するか

`docker-compose.yml` のホスト公開ポートと下記 volume は `${VAR:-default}` 化されており、
**main repo だけが既定値を使う**。worktree が無設定で `make up` すると次が起きる:

- ホストポート `5434 / 8080 / 3003 / 6006 / 8081 / 8082` が main（または他 worktree）と衝突し `make up` が失敗する
- `postgres_data` volume を main と共有 → 2つ目の postgres は PGDATA lock で落ちる。片方だけなら **worktree が main の DB をそのまま使い**、seed 状態・migration 適用状態を壊す
- `frontend_node_modules` volume を main と共有 → worktree 間の `pnpm-lock.yaml` 差異で依存が壊れる
- バインド失敗時でもコンテナは内部ネットワーク上で動き続けるため、`localhost:5434` が**別プロジェクトの DB** を指す罠になる

## 作成時（必須手順）

1. `git worktree add ../AnimalEkarte-<task> -b <branch> origin/main`
   - 必ず `git worktree` 登録で作る（`git worktree list` が棚卸しの正本）。plain clone / 未登録コピーは孤児管理の対象外になるため避ける
   - 配置は main の sibling `../AnimalEkarte-<task>` を推奨。`/tmp` 等も可だが git 登録は必須
2. `scripts/tools/worktree-env.sh <offset> [slug]` を実行し `.env.local` を生成する
   - `<offset>` にはチケット番号を使うと一意になりやすい（EMR-300 → `300`）
   - `[slug]` 省略時は `wt<offset>`。container/project/volume 名に使われる
   - `.env.local` が既存なら managed block 以外の行は保持される。手編集で managed block を書かない
3. `make up` はユーザーの手で起動する（自動実行禁止ルールに従う）

## 割り当て式（offset = N、1–9256）

`B = 10000 + (N - 1) * 6` をworktreeのポートblock先頭とする。
mainの既定値は変更しない。異なるNでは6サービスすべてのポートが重ならない。
既存worktreeはスクリプト再実行でmanaged blockを更新し、更新後の起動はユーザーが行う。

| 資源 | 値 |
|------|-----|
| compose project | `animalekarte-wt<N>`（コンテナ名 `animalekarte-wt<N>-db-1` 等） |
| DB port | `B`（host → container 5432） |
| backend port | `B + 1` |
| frontend port | `B + 2`（`CORS_ALLOWED_ORIGIN` も自動追随） |
| storybook / swagger / redoc | `B + 3` / `B + 4` / `B + 5`（profile `docs`） |
| postgres_data volume | `ekarte-postgres-data-<slug>` |
| frontend_node_modules volume | `ekarte-frontend-node-modules-<slug>` |

キャッシュ系 volume（`ekarte-go-mod-cache` / `ekarte-go-build-cache`）は
content-addressed で、**全 project で共有するのが意図**（分離しない）。

## localhost ポートを信用しない

ホストポートの行き先は「それを掴んだプロセス次第」であり、bind 失敗・孤児コンテナで
**別 project のサービス**に繋がりうる。DB・コンテナ操作は必ずプロジェクトスコープで:

```bash
docker compose --env-file .env.local exec db psql -U "$DB_USER" -d "$DB_NAME"   # ○ 正しい
psql -h localhost -p 5434                                                       # × 行き先不明
curl http://localhost:8080/health                                               # × main 側を叩く可能性
docker compose --env-file .env.local exec backend wget -qO- localhost:8080/health  # ○
```

`localhost:<port>` 前提の script/スキル文書（local-db-reset、browser-test 等）は
**default stack（main）での実行を前提**にしている。worktree 内で使う場合は
各自の `*_HOST_PORT` に読み替える。

## 終了・削除時（必須）

worktree を閉じる前に、**その worktree のディレクトリ内で**停止する:

```bash
docker compose --env-file .env.local down          # コンテナ+network 停止（volume=DB は残る）
docker compose --env-file .env.local down -v       # データも消す場合（自身の -wt<N> volume のみ対象）
git worktree remove ../AnimalEkarte-<task>          # 停止後に削除
```

- `down` 忘れは**孤児コンテナがポートと DB volume を握り続ける**事故になる（実例: `ae-emr-300-db-1` が worktree 削除後も 5434 を 20 時間占有）
- `down -v` は project スコープの volume のみ消す。`-wt<N>` 付き volume のみが対象で main の `ekarte-postgres-data` は消えない
- `make reset`（local-db-reset-contract）は **main project 専用**で、worktree の `.env.local` では volume 名の不一致により fail-closed で拒否される。worktree の DB を初期化したいときは `down -v && up` を使う

## 禁止

- **他 project のコンテナ / volume / network を、所有者確認とユーザー明示承認なしに `stop` / `rm` / `down` / `prune` しない**
- `docker system prune` / `docker volume prune`（project をまたぐ全削除。自動実行禁止・ユーザー操作のみ）
- main repo の `.env.local` に worktree 用 offset を書く（main は既定値のまま）
- `.env.local` の managed block 手編集（スクリプト再実行で冪等更新する）

## 棚卸し（孤児検出）

```bash
git worktree list                                              # 登録済み worktree の正本
docker ps -a --format '{{.Names}} | {{.Status}} | {{.Label "com.docker.compose.project"}}'
docker volume ls | grep ekarte                                  # project 接頭辞の無い名前は共有/既定
```

- `git worktree list` に無い project のコンテナ = 孤児。所有者を確認し、不要なら `docker compose -p <project> down` を**ユーザーに提案**してから止める
- worktree 削除済みだが volume が残る場合は、中身が不要と確認できてから `docker volume rm`（これもユーザー承認）

## 検証

`docker compose --env-file .env.local config` で、自分の worktree が解決する
project 名・ポート・volume 名を起動前に確認できる（コンテナを起動しない静的検証）。
