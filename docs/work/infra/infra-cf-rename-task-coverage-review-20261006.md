# AnimalEkarte から noah-karte への Cloudflare アカウント移管・改名 — タスク網羅性レビュー

> **後続文書**: [infra-cf-rename-task-coverage-rereview-20261006.md](infra-cf-rename-task-coverage-rereview-20261006.md)（2026-10-06）が本書の判定を更新済み（抜け 16 件の対応先確定・チケット再構成は移管実行と改名を EMR-302 の 1 枚へ統合）。本書は履歴参照。

> 種別: タスク網羅性レビュー（2026-10-06）。実行 SoT ではない。状態・担当・Done は Plane を正とする。根拠は同日の依頼文に記載された確定事項と棚卸し実測のみで、リポジトリ、Terraform、Plane、Cloudflare は照合していない。

判定日: 2026-10-06。10/20 までは 14 日ある。

## 判定

計画、移行実行、改名、使用量確認の4分割は妥当である。抜けているのは、新アカウントで新名のまま起動し、STG で単一実行を確認し、切り戻しを一度実施してから、共有アカウント側の当該資源だけを退役するまでの実行ゲートである。

10/20 の本番切替は、共有アカウント上の作業から、新アカウントの本番シェルに対する初回構築へ組み替える。Go module 約 1700 ファイルとディレクトリ改名をこの 14 日のクリティカルパスに置くと、本番未稼働のうちに移せるという前提が消える。

## 前提

- Cloudflare 専用アカウントを新設し、共有アカウントから全資源を移管する。本番は未稼働、STG のみ稼働中であり、移行コストが最小の時期である。
- サービス名「AnimalEkarte」を「noah-karte」ファミリーへ改名する。移行先で新名を構築し、旧名を退役するまでを一体で実施する。
- 命名規則は、CF 資源が `noah-karte-*`、Go・TS 識別子が `NoahKarte*`、DB・docker が `noah_karte*`、ドメイン・docs・UI が `noah-karte.com` と `Noah Karte` である。
- 外部 DB は PlanetScale、CI は GitHub Actions である。

## 棚卸し実測

Cloudflare 資源は次のとおり。

- Worker は4つ。stg/prod の api/frontend で、frontend は service binding で API を参照する。
- R2 bucket は2つ。`stg-images` と `prod-images`。prod は空で、データがあるのは stg のみ。
- Durable Objects の class は `AnimalEkarteApiContainer`。named instance は `animalekarte-scheduler-v1` と `animalekarte-migrate-runner-v1`。wrangler migrations では `renamed_classes` の設計が必要。
- keep-alive cron は5式。営業時間帯の container 稼働が費用の主因。
- ゾーン `noah-karte.com` は Terraform 管理で authoritative 稼働中。STG state は11資源、本番 state は4資源。
- Hyperdrive は config 削除済みだが STG state に残存する。整理対象は旧アカウント側。
- アカウント固有の値は、`S3_ENDPOINT` のアカウント ID 直書き、`AWS_*` R2 キー、Cloudflare Access（`SCHEDULER_ACCESS_TEAM_DOMAIN` / `AUDIENCE`）、CI トークン `CLOUDFLARE_API_TOKEN` / `PROD_CLOUDFLARE_API_TOKEN`、workers.dev `*.baritech-soga.workers.dev`（10ファイル）である。
- 通知ポリシー `worker_edge_error_rate` は定義済み・未適用で、`notification_email` は未決。`billing_usage_alert` は Workers/Containers が対象外のため、GraphQL Analytics の定期確認が代替である。
- 外部依存として、LINE/LIFF、CORS、`WORKER_URL` が workers.dev URL に依存する。

CF 外の改名対象は次のとおり。

- Go module `github.com/animal-ekarte/backend`（import path 約 1700 ファイル）。GitHub remote は `MinoruSoga/AnimalEkarte.git` で、module path と既に不一致。
- docker compose project（ディレクトリ名由来）と DB 名 `animal_ekarte_test`。
- macOS packaging 識別子 `animalekarte.lab-device-agent.plist` 等。
- GitHub repo 名、ローカルディレクトリ名、docs/UI 表記、Plane チケット表記。

## レビュー対象のタスク

1. EMR-298（Plane）は棚卸しと移行計画のみ。実行は範囲外。計画文書に改名対応表、手順、ロールバック、人間決定リストを要求済み。
2. INFRA-CF-ACCOUNT-MIGRATION-EXEC（下書き）は移行実行。アカウント作成、トークン発行、有料プラン、ゾーン/DNS 移管、GitHub secret 更新は人間承認ゲート。STG 予行のあと本番。R2 キー、`S3_ENDPOINT`、workers.dev URL、LINE/LIFF、CORS、Hyperdrive state 整理を含む。
3. INFRA-RENAME-NOAH-KARTE（下書き）は改名実装。Go module 全置換、DO class の `renamed_classes`、docker/DB、packaging、repo/dir、docs/UI。
4. INFRA-CF-USAGE-CHECK（下書き）は GraphQL Analytics による使用量定期確認の runbook 化と、通知ポリシー適用。
5. 既存 Plane チケット（別件）は、本番環境構築・切替（10/20 予定）、PROD-SETUP、GOLIVE 判断、初回管理者付与。
6. 人間決定（計画の出口）は、アカウント所有者、費用主体、所有者メール、有料プラン承認、実施時期、repo/dir/Go module の改名者。

## カバー済み

次は実測とタスクが対応している。

- EMR-298 は棚卸しと計画に限定され、実行を範囲外にしている。改名対応表、手順、ロールバック文書、人間決定リストを要求している。
- アカウント作成、トークン発行、有料プラン、ゾーン/DNS 移管、GitHub secret 更新が人間ゲートになっている。
- R2 キー、`S3_ENDPOINT`、workers.dev URL、LINE/LIFF、CORS、Hyperdrive の STG state 整理が移行実行に含まれている。
- `worker_edge_error_rate` の適用と、Workers/Containers を見ない `billing_usage_alert` の代替である GraphQL Analytics が別タスクになっている。
- 本番 R2 `prod-images` は空で、本番は未稼働である。データ移行の対象は STG の `stg-images` だけである。

## 1. 抜け

### 1.1 共有アカウント側の退役タスクが無い

前提は「移行先で新名構築→旧名退役の一体実施」である。実行タスクが旧側で名指ししているのは「Hyperdrive は config 削除済みだが STG state に残存」だけである。

退役対象として実測に残っているのは、Worker×4、データのある `stg-images`、空の `prod-images`、DO の named instance 2つ、keep-alive cron 5式、Cloudflare Access、CI トークン `CLOUDFLARE_API_TOKEN` / `PROD_CLOUDFLARE_API_TOKEN`、Terraform の STG 11 資源と本番 4 資源である。cron 5式は費用の主因なので、新旧両方を残すと費用が二重化する。

### 1.2 Durable Object の状態が class 改名に還元されている

実測は class `AnimalEkarteApiContainer` と、instance 名 `animalekarte-scheduler-v1` / `animalekarte-migrate-runner-v1` である。改名タスクにあるのは `renamed_classes` だけである。

`renamed_classes` は同一スクリプトの class 名を引き継ぐ設計であり、新アカウントへ storage をコピーする操作ではない。`idFromName` の文字列変更でもない。命名規則 `noah-karte-*` を instance 名に適用するとキーが変わり、旧 instance の storage は参照されない。新アカウントの初回起動時、storage は空である。

scheduler と migrate-runner が空状態で再入しても安全か、その空起動を受け入れてよいかの決定が、人間決定リストに無い。リストにあるのは所有者、費用主体、所有者メール、有料プラン、時期、改名者である。

### 1.3 `stg-images` のコピーと検証が無い

移行タスクは R2 キーと、`S3_ENDPOINT` に直書きされたアカウント ID を更新する。bucket はアカウントをまたいで改名できない。データがあるのは STG だけである。コピー、件数またはチェックサム、新 bucket への切替、旧キー無効化の順がタスクになっていない。`prod-images` は空なので、本番側は空 bucket の作成で足りる。

### 1.4 Cloudflare Access の再作成が無い

`SCHEDULER_ACCESS_TEAM_DOMAIN` / `AUDIENCE` はアカウント固有である。ゾーン移管の人間ゲートはあっても、Access はゾーン資源ではない。新アカウント側の team、Access application、audience の作成と、scheduler 側の設定更新が、トークン発行や GitHub secret 更新から分かれていない。

### 1.5 secret の棚卸しと旧値の失効が無い

登録されているのは新トークンの発行と GitHub secret の更新である。実測でアカウントに束縛されているのは次のとおり。

- `AWS_*` R2 キー
- `S3_ENDPOINT` のアカウント ID
- Access の team domain と audience
- `CLOUDFLARE_API_TOKEN`
- `PROD_CLOUDFLARE_API_TOKEN`
- workers.dev の10ファイル

旧トークンと旧 R2 キーを、コピーと STG 確認のあとで無効化する作業が無い。Hyperdrive 設定は削除済みでも、旧 Worker の secret に接続情報が残っていないかの確認項目も無い。

### 1.6 Terraform state の15資源の扱いが無い

ゾーン `noah-karte.com` は Terraform 管理で authoritative 稼働中である。移行タスクはゾーン/DNS 移管を人間ゲートにしている。state 上の STG 11 資源と本番 4 資源を、移す、作り直す、捨てるに分類する作業が無い。

差は7資源である。実測が名前を挙げているのは Hyperdrive の残存だけであり、残り6資源は未分類である。Hyperdrive を消す前に state を新アカウントへコピーすると、削除済み設定を新 state へ運ぶ。

### 1.7 workers.dev の置換先と `workers_dev` flag が未決

URL 更新自体は移行タスクにある。参照は `*.baritech-soga.workers.dev` の10ファイルである。公開ゾーンは既に `noah-karte.com` である。各参照の置換先が、既存ゾーンのホスト名なのか、新アカウントの別 subdomain なのかが決まっていない。`workers_dev` を切る作業も無い。

共有アカウント名 `baritech-soga` を示すこの subdomain を、製品退役のついでに解放する作業は登録しない。

### 1.8 LINE / LIFF コンソールが人間ゲートになっていない

外部依存として「LINE/LIFF・CORS・WORKER_URL 参照が workers.dev URL に依存」とある。人間ゲートはアカウント、トークン、有料プラン、ゾーン/DNS、GitHub secret である。LINE 側の webhook と LIFF endpoint はリポジトリ外である。

### 1.9 STG 切り戻し訓練が実行項目になっていない

EMR-298 が要求しているのは計画文書のロールバックである。実行タスクの「STG 予行→本番」には、DNS、LINE、cron の武装、secret を旧側へ戻す訓練が無い。旧 Worker を消したあとは、その手順を再現できない。

### 1.10 実測面に対応する確認項目が無い

予行で確認する面は次のとおり。別チケットにはせず、移行実行の受け入れ条件にする。

- Worker×4
- frontend から API への service binding
- データコピー後の `stg-images`
- 空の `prod-images`
- DO instance 2つ
- cron 5式
- Access
- LINE/LIFF と CORS
- authoritative ゾーン
- 新しい CI トークンの投入先アカウント

### 1.11 EMR-298 の計画出力が、実行前の分類を要求していない

要求済みは改名対応表、手順、ロールバック、人間決定リストである。実行前に対応表へ固定するものは、state 15 資源の処遇、secret 一覧、DO 空起動の受け入れ、cron の単一実行、10ファイルの置換先である。

### 1.12 命名規則から一意に決まらない文字列がある

規則は `noah-karte-*`、`NoahKarte*`、`noah_karte*`、`noah-karte.com` / `Noah Karte` である。現行 module は `github.com/animal-ekarte/backend`、remote は `MinoruSoga/AnimalEkarte.git` で、既に一致していない。人間決定は改名者だけで、約 1700 ファイルへ入れる文字列自体が未決である。

Go module path、GitHub repo slug、ローカルディレクトリ名は、対応表で1文字列ずつ固定してから置換する。DB 名 `animal_ekarte_test` は `noah_karte*` に落ちるので、ここだけは規則で足りる。

### 1.13 PlanetScale とローカル DB 改名が分離されていない

外部 DB は PlanetScale で、secret 実測に DB 資格情報の移管は無い。DB は留まり、Worker だけがアカウントを変える。改名対象の `animal_ekarte_test` は、この実測ではローカルまたはテスト DB の名前である。ホスト済み DB 名の変更としては登録されていない。接続文字列を新 Worker へ引き継ぐ項目も無い。

### 1.14 Container 本体の新アカウント作成が名指しされていない

実測の class は `AnimalEkarteApiContainer` である。cron 5式は営業時間帯の container 稼働であり、費用の主因である。新アカウントでは image と container app を作り直す。移行項目は Worker、R2、URL、Hyperdrive であり、container の初回作成が分かれていない。

### 1.15 導入済み lab device agent の更新手順が無い

改名対象は `animalekarte.lab-device-agent.plist` 等の識別子である。識別子を変えたあとの unload、再登録、導入済み端末の更新がタスクに無い。

### 1.16 通知先メールが使用量タスクの前提として接続されていない

`worker_edge_error_rate` は定義済み・未適用で、`notification_email` は未決である。人間決定の「所有者メール」と同一なら、計画文書で同一と書く。別アドレスなら、決定項目が一つ足りない。

## 2. 過剰

### 2.1 `renamed_classes` に新アカウントへの状態引き継ぎを期待すること

migration 履歴を線形に保つために `renamed_classes` を wrangler の履歴へ書くこと自体は、実測の「設計が必要」と一致する。`animalekarte-scheduler-v1` と `animalekarte-migrate-runner-v1` の storage が新アカウントで読める、という期待は実測の操作に含まれない。

### 2.2 本番を STG と同じデータ移行として扱うこと

`prod-images` は空で、本番は未稼働である。本番フェーズは空 bucket、Worker の殻、トークン、DNS、Access で足りる。オブジェクトコピーは STG だけである。

### 2.3 既存 Plane チケットの全文書き換え

表記は改名対象である。本番構築、10/20 切替、PROD-SETUP、GOLIVE、初回管理者付与は別件の監査記録である。今後のタイトルと対応表で足りる。履歴の一括置換は、10/20 チケットの追跡を壊す。

### 2.4 共有アカウントの閉鎖、または workers.dev subdomain の解放

退役範囲は今回の実測資源に限る。`*.baritech-soga.workers.dev` という subdomain は、このアカウントが当該製品専用ではないことを示している。

### 2.5 通知ポリシー適用を移行の完了条件にすること

ポリシーは未適用、メールは未決である。費用の本丸である Workers/Containers は `billing_usage_alert` の対象外である。使用量タスクは新アカウントの GraphQL 確認として残し、移行と 10/20 のブロッカーから外す。

## 3. 分割案

`INFRA-RENAME-NOAH-KARTE` は実施単位として粗い。同一チケットに、デプロイと原子的に動かす識別子、1700 ファイルの import、ディレクトリ名由来の compose project、導入済み packaging、文書が同居している。失敗時の戻し方が一つになる。

5単位に分ける。

| 単位 | 中身 | 受け入れ |
|---|---|---|
| CF 実行時識別子 | Worker 名、service binding、container class、DO instance 名、R2 bucket 名、cron、wrangler の account 依存値 | 新アカウントの初回デプロイが `noah-karte-*` で起動する |
| Go module | `github.com/animal-ekarte/backend` の一括置換。先に対応表で置換先を1文字列に固定 | 1コミットでコンパイルとテストが通る |
| ローカル開発名 | ディレクトリ名、compose project、`animal_ekarte_test` | ボリューム名の変化を手順に書き、ローカルデータのリセットを明示する |
| packaging | `animalekarte.lab-device-agent.plist` 等と、導入済み端末の更新 | CF 切替と別日に実施できる |
| 表示名 | docs、UI の Noah Karte、Plane の今後分 | 実行識別子の凍結後に並行できる |

Go module は CF 切替の前提ではない。コンパイル済み Worker の名前と module path は独立している。1700 ファイルの置換は、進行中の本番構築チケットと同じファイルを広く触る。静かなマージ窓で単体に実施する。

ディレクトリ改名は最後である。compose project はディレクトリ名由来なので、ディレクトリを変えると project 名と volume 名が変わる。`animal_ekarte_test` のローカルデータは旧 volume に残る。

## 4. 順序リスク

### 4.1 10/20 を共有アカウントで先に実施する

前提は「本番未稼働・STG のみ稼働中で、移行コスト最小の時期」である。先に本番を共有アカウントへ載せると、未稼働の本番 Worker、空の `prod-images`、本番 state 4 資源を、稼働後にもう一度移すことになる。

PROD-SETUP、GOLIVE、初回管理者付与の向け先は、新アカウントの本番シェルにする。10/20 の実作業は「新アカウントでの初回本番構築」である。所有者、費用主体、有料プランが STG 予行に間に合わない場合は、10/20 の日付を動かす。

### 4.2 旧名で新アカウントへデプロイしてから改名する

前提は移行先での新名構築である。CF 実行時識別子の変更と、新アカウントへの初回デプロイは同じリリースにする。`renamed_classes` を新アカウントで一度適用してから instance 名も変えると、空の DO を二段階作るだけになる。

### 4.3 cron と DO を両アカウントで同時に武装する

外部 DB は PlanetScale で、DB 移管は実測に無い。新旧 Worker は同じ DB のクライアントになる。`animalekarte-scheduler-v1` と `animalekarte-migrate-runner-v1` を両側で動かすと、定期処理と migrate runner が二重化する。費用も二重化する。

新側の確認が終わるまで新 cron は止める。武装を切り替える瞬間に旧側を止める。両側を同時に動かす窓は作らない。

### 4.4 service binding を片側ずつ移す

frontend は service binding で API を参照する。binding は同一アカウント内である。環境ごとに API と frontend を同じデプロイで作る。frontend だけ先に新アカウントへ出すと、API を参照できない。

### 4.5 新 Worker がホスト名を受ける前にゾーンを移す

`noah-karte.com` は既に authoritative である。STG は唯一の稼働環境である。DNS を新 Worker へ向ける前に、新側で同じホスト名を受けられることを確認する。LINE/LIFF の向き先変更は、その DNS 切替と同じ窓で行う。

### 4.6 Hyperdrive の残存を state コピーに含める

設定は削除済みで、残っているのは STG state である。旧 state から外してから、分類済みの資源だけを新アカウントで import または再作成する。

### 4.7 コピーと切り戻し訓練の前に旧資格情報や旧 Worker を消す

`stg-images` にデータがあり、STG は稼働中である。旧 R2 キー、旧 API トークン、旧 Worker の失効と削除は、チェックサム、確認項目、切り戻し訓練、短い soak の後にする。

### 4.8 STG 予行中に本番トークンで Workflow を流す

トークンは STG 用と本番用で分かれている。予行で更新するのは `CLOUDFLARE_API_TOKEN` だけにする。`PROD_CLOUDFLARE_API_TOKEN` は、空の本番シェルを作る時点で結線する。

### 4.9 切替窓に Go module、repo 改名、ディレクトリ改名を重ねる

remote と module path は既に不一致である。repo 改名は Actions、secret、ローカル remote、Plane のリンクに波及する。ディレクトリ改名は compose project と volume を変える。CF の単一実行確認が終わるまで、この3つは動かさない。

repo は新規作成ではなく、既存 `MinoruSoga/AnimalEkarte.git` の rename にする。新規 repo では secret と CI 履歴が続かない。

### 4.10 使用量通知の未決を移行の前提にする

`notification_email` が未決でも、GraphQL Analytics の定期確認は新トークン発行後に開始できる。ポリシー適用はメール決定後で足りる。

## 実行順

1. EMR-298 の計画出力を拡張する。追加分は、state 15 資源、secret 一覧、DO 空起動の受け入れ、cron 単一実行、workers.dev 10ファイルの置換先、Go module path、repo slug、ディレクトリ名である。
2. 人間が次を決める。アカウント所有者、費用主体、所有者メール、有料プラン、10/20 を新アカウント初回構築とするか日付を動かすか、LINE コンソール担当、repo の in-place 改名者。
3. 新アカウントに有料プラン、トークン、Access、空 R2、container、新名の Worker×4 を作る。cron は停止のままにする。
4. `stg-images` をコピーして照合し、PlanetScale の接続先を新 STG Worker へ渡す。本番 bucket は空のまま作る。
5. STG の確認項目を実施する。通過後、cron の武装を旧から新へ一度で切り替える。
6. DNS、LINE/LIFF、CORS、10ファイルを新ホスト名へ向ける。
7. STG で切り戻し訓練を1回行い、再度切り替える。soak 後に、共有アカウントの当該資源だけを退役し、旧トークンと旧 R2 キーを失効する。
8. 既存の PROD-SETUP、10/20、GOLIVE、初回管理者付与を、新アカウントの本番シェル向けに読み替える。
9. 新アカウントで GraphQL の使用量確認を開始する。メール決定後に `worker_edge_error_rate` を適用する。
10. STG が安定してから、Go module、packaging、ローカルディレクトリ、docs/UI、Plane の今後分を、分割単位ごとに実施する。
