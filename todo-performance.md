# Performance 調査・改善 TODO

作成・調査日: 2026-09-09（JST）／対応状況の最終照合: 2026-09-10（JST）

調査 ID: **PERF-STG-LOGIN**

対象: STG `/login` の初回表示遅延。責任者・依頼者: 曽我 稔。

ソース照合の基準点: 当初整理着手時のローカル `main` `7c6a4a9d1`。2026-09-10 の本プラン補完時はローカル `main` = `origin/main` = `813c5852d`（本プランの変更自身は含めない）。調査時に照合した STG bundle は `main-BnyQFmpH.js` であり、現在の STG 配信 revision は未照合。

実行 SoT: Linear。本書はユーザー指定のローカル調査・実装準備記録。Linear の関連 issue / 状態は **UNKNOWN（今回未照会・未更新）**。入口は [todo.md](todo.md)。

## 現在の未完了状況（2026-09-10）

- **STG 昇格:** `main` → `staging` の PR [#388](https://github.com/MinoruSoga/AnimalEkarte/pull/388) は OPEN / CONFLICTING。変更検出・security checks は PASS、product jobs は SKIP。STG 配信、ブラウザー/E2E、改善後の実測は **UNKNOWN / 未実施**。
- **原因調査:** 約22.5秒の preflight・接続待ち・Container 起動の内訳が未確定。
- **残実装:** 原因調査の結果に基づく通信経路対策 C と、初回 bundle を実測して判断する D。UI変更だけを通信時間の改善とは扱わない。

## 結論と判定境界

**調査時に確定した事実: 当時はセッション確認が終わるまでログイン画面を描画しなかったため、通信待ちが白画面の待ち時間になっていた。**

**未確定: 通信全体約22.7秒のうち約22.5秒を占めた、最終 GET の接続開始前の待ち時間の内訳。** 最終 GET の送信から応答開始は約116msだった。「Go API / DB が23秒処理した」とは言えない。一方、CORS preflight（OPTIONS）が先行する構成なので、**ブラウザー内だけの停滞とも断定できない**。OPTIONS の通信・Container 起動待ちが先行区間に含まれた可能性は残る。

先行の説明「認証APIに約23秒」はブラウザーから見た全体所要時間を指す。サーバー処理時間として扱わない。「DNS 前だからサーバーと無関係」という説明も採用しない。

## 実測記録

### E1: 遅延が発生した既存ブラウザー記録

- Chrome の既存 `/login` タブに残っていた Performance API の記録を読取。新規負荷試験ではない。
- navigation 開始: **2026-09-09 14:18:14.373 JST** (`performance.timeOrigin = 1788931094373.4`)。
- HTML TTFB: 49.2ms、DOMContentLoaded: 303.1ms、load: 309.1ms。
- FCP（初めて内容が表示される時刻）: **23,548ms**。背景のみの first-paint は320ms。
- `/api/v1/me`: HTTP **401**、HTTP/3、全体 **22,689.3ms**。401だけでは Cookie の不在・期限切れなどを区別できない。Cookie / 認証情報は採取していない。

以下は navigation 開始からの相対時刻（ms）。

| 項目 | 時刻 | 区間の所要時間 |
|---|---:|---:|
| fetchStart | 309.1 | — |
| domainLookupStart | 22,855.4 | fetchStart から **22,546.3** |
| domainLookupEnd | 22,855.8 | DNS **0.4** |
| connectEnd | 22,880.1 | 接続 **24.3** |
| requestStart | 22,880.3 | — |
| responseStart | 22,996.1 | 最終 GET の応答待ち **115.8** |
| responseEnd | 22,998.4 | 本文受信 **2.3** |

`workerStart=0`。Service Worker の処理を示す記録はない。`serverTiming` は `cfExtPri` のみで、Worker / Container / DB の内訳なし。遅延発生時の Network イベントは保存範囲外であり、**当時の OPTIONS・request ID・接続交渉イベントは回収できなかった**。この表は取得した数値の転記で、完全な HAR / NetLog ではない。

### E2: 同日、別の調査タブでの再確認

通常設定のまま同一 Chrome プロファイルで測定。キャッシュ・Cookie の消去、拡張機能無効化、QUIC・proxy 設定変更、ログイン操作はしていない。

| navigation 開始（JST） | `/me` 全体 | 最終 GET の応答待ち | FCP | 結果 |
|---|---:|---:|---:|---|
| 14:28:19.998 | 221.4ms | 184.4ms | 944ms | 401、HTTP/3、フォーム表示確認 |
| 14:28:36.861 | 190.0ms | 188.0ms | 768ms | 401、HTTP/3、フォーム表示確認 |

2回目は Network イベントも取得。対象 GET は1件で、接続再利用、proxy/DNS/connect の新規処理時間なし。**2回の正常読込は再現しなかったという証拠であり、修正済み・cold start 解消・p95/p99 達成の証拠ではない。**

### E3: 認証情報なしの単発 HTTP 確認

同日の調査中に実施。ブラウザーの認証済み経路・HTTP/3・CORS 自動 preflight を再現するものではない。

| リクエスト | 結果 | 全体 |
|---|---|---:|
| GET STG `/login` | 200 | 281ms |
| GET API `/api/v1/me`（Cookie なし） | 401 | 254ms |
| OPTIONS API `/api/v1/me`（STG Origin、GET、`x-requested-with,x-request-id` を指定） | 204 | 240ms |

負荷試験や warm-up の継続実行はしていない。共有環境を意図的に停止・休止させる操作もしていない。

## 調査時にソースから確認した通信経路

1. [axios.ts](frontend/src/lib/axios.ts) は `X-Requested-With` / `X-Request-ID` を付ける。別 origin の API なので、preflight cache がないと OPTIONS が先行する。`X-Requested-With` は CSRF 防御のためであり、高速化目的の無条件削除は不可。
2. [Worker](backend/worker/index.ts) は通常リクエストを Container に転送する。OPTIONS の edge 即時応答はなく、Container が休止中なら起動を待つ可能性がある。**設定・コード上の候補であって、E1で実際に休止していた証拠は未取得。**
3. [CORS middleware](backend/internal/middleware/cors.go) は preflight max-age 86400を設定し、OPTIONS に204を返す。ブラウザーによるキャッシュ上限・キャッシュ状態は別途考慮する。
4. [auth middleware](backend/internal/middleware/auth.go) は token 不在ならDB照会前に401を返す。有効 token の認可経路は異なるため、この短絡を全401に一般化しない。

## 調査・改善 TODO

業務目的: ログイン開始時の通信待ちを特定して短縮し、再読込・操作のやり直しを減らす。認証・CSRF・医院分離は維持する。Browser/E2E・STG実測・LinearはUNKNOWNまたは未実施であり、C/Dは未着手。

| 優先 | 項目 | 状態 | 実施内容・完了条件 |
|---|---|---|---|
| P0 | 遅延区間の確定 | 部分完了 | E1の22.5秒が最終GET接続前であることは確認済み。次は下記手順で OPTIONS / 接続待ち / Container 起動の内訳を確定する |
| P1 | preflight / Container 起動の遅延対策 | 仮説検証待ち | OPTIONSと起動イベントの相関確認後に対策を選ぶ。edgeでのOPTIONS応答はCORS許可元・ヘッダ・資格情報方針を一致させる。GET自体のcold startは別に残る。sleep設定変更は費用と共有環境への影響を明示して判断する |
| P1 | 改善後の再計測 | 未実施 | 同一条件のwarm/cold・初回/再読込を区別し、FCPだけでなくフォーム操作可能時刻、OPTIONS/GET各時間を保存。spinnerの表示だけを「ログイン高速化完了」としない |
| P1 | STG 配信・ブラウザー受入 | PR競合・未実施 | PR #388 の競合解消候補を隔離worktreeで作成・検証する。外部更新の承認後に配信し、配信revisionを固定して下記受入を実行する |
| P2 | 初回 bundle の不要読込 | TODO・今回の主因ではない | 配信HTMLは charts / LIFF 等をpreload。ログインに必要な依存を実測して分離する。E1はload約0.31秒のため、23秒の主因として扱わない |
| P2 | Linear 照合 | UNKNOWN | 関連 issue の有無と現行状態をライブで読み取り、既存 issue と重複しない反映案を作る。書き込み・状態変更は明示承認後に行う |

### 次回、原因を確定するための具体的な開始手順

1. 調査タブで **遷移前から** Network 記録を開始し、通常の `/login` 読込を1回記録する。OPTIONS と GET を initiator / URL / 時刻で関連付ける。必要最小の時刻・method・status・protocol・timingだけを抽出し、Cookie / Authorization / 本文 / 患者情報は保存しない。
2. 自然な未使用期間後の初回と直後の再読込を比較する。10分経過だけで「cold」と判定せず、Worker / Container の起動イベントを照合する。他利用者の通信を停止せず、休止を強制しない。
3. OPTIONSが遅ければWorkerからContainerへの待ち・起動イベントと照合する。GETだけの backend latency が短くても、OPTIONSのcold startは否定できない。
4. OPTIONSも速い場合は Chrome の Queueing / Stalled、proxy negotiation、接続再試行を確認する。HTTP/3使用は確認済みだが、**QUIC不具合の証拠ではない**。拡張機能・proxy・QUICの変更による比較は、通常環境を変えずに証跡を確保した後、対象と影響を限定して実施する。
5. backendのGETは [logging.go](backend/internal/middleware/logging.go) のrequest ID / latencyで照合できる。ただしCORSがloggingより先でOPTIONSを終了するため、**現行のGoリクエストログだけではOPTIONSの遅さやContainer起動前を測れない**。必要なら別タスクでWorker開始・forward前後の時間と起動イベントの観測を追加する。

現時点の不足: E1当時のOPTIONS / 接続イベント、Cloudflareの該当時刻ログ、Container起動状態・配信revision。利用可能な接続ツールには、この環境のCloudflareログを直接取得する専用接続が見つからず、管理画面ログ・DBを取得していない。遅延は追加の2回では再現しなかったため、根因を確定扱いにしない。

## 残作業の改善プラン

Cで通信待ちの原因を特定してから対策を選ぶ。Dは初回キャッシュなし計測で必要性が確認できた場合だけ実施する。

### C. 通信経路を計測して、効果のある対策だけを入れる

まず遷移前からOPTIONSとGETを記録する。必要な追加観測は、Workerで「受付→Container forward完了」の所要時間と起動イベントを取得する小さな変更に限定する。固定のmethod/path分類・status・時間・相関IDに絞り、Cookie・認証ヘッダ・本文をログに出さない。常時監視基盤の新設はこの改善の前提にしない。

| 計測で分かったこと | 採用する対策 | 採用条件・注意点 |
|---|---|---|
| OPTIONSが休止Containerの起動を待つ | 起動時間の短縮を優先して調べ、必要なら未使用時の休止方針を調整 | 起動イベントとの相関を確認してから。`sleepAfter`延長は再発頻度を下げてもcold起動自体は短縮しない。常時稼働・keep-aliveは費用比較と外部変更の承認後に判断 |
| warmでもOPTIONSの余分な往復が継続的に効く | WorkerでOPTIONSを返す案を評価 | CORS許可元・method・header・credentialsと内部ルート境界を一致させる。policy重複のdriftを契約テストで防ぐ。ヘッダ除去や許可元の拡大で回避しない |
| OPTIONSは速く、最終GET前に接続待ちがある | proxy / 接続交渉 / 拡張機能等を証跡に沿って比較 | HTTP/3というだけでQUICを無効化しない。対象環境の設定変更は影響を限定し、原因確認前に利用者へ一律変更を求めない |
| GET送信後の待ちが再現する | Worker時間とGoのrequest latencyを比較し、該当処理を限定修正 | DB照会・pool待ちは、その時間を示す証拠が出た場合だけ対象にする |

**OPTIONSのedge応答だけでは、次のGETが休止Containerを起動するため、cold時の総待ち時間は残り得る。これ単独を23秒問題の解決策にはしない。** Containerのライフサイクルは [Cloudflare公式資料](https://developers.cloudflare.com/containers/configuration/scaling-and-routing/)、preflight順序は [Fetch仕様](https://fetch.spec.whatwg.org/#cors-preflight-fetch) と照合する。

### D. 初回JSの読込量を減らす（後続・条件付き）

初回キャッシュなしの計測で静的リソースの読込・実行が残る場合に着手する。ログインに不要なcharts / LIFF等の混入経路を確認し、既存の公開entrypointとchunk構成を最小限で調整する。feature境界を破るdeep importや、実測なしのmanualChunks全面再編は採用しない。変更前後の転送量・parse/execute時間・フォーム操作可能時刻を比較し、体感時間が改善することを確認する。

### STG 配信・ブラウザー受入の着手プラン

1. **リモート状態を固定する:** 実行直前に `main` / `staging` の SHA と PR #388 の head・base・mergeable・checks を読み取り、記録する。2026-09-10 の照合では PR #388 は `main` → `staging`、head `813c5852d`、OPEN / CONFLICTING。product jobs は SKIP であり、候補の品質証明には使わない。
2. **競合解消候補を作る:** 最新 `main` から隔離worktreeを作り、`staging` との差分と競合を先に `merge-tree` で確認する。候補ブランチ内で `staging` をmergeし、環境固有設定と `main` の機能変更をファイル単位で照合する。`backend/migrations/`、seed、起動条件、workflowの差分を別枠で確認し、適用済みmigrationの編集やenv/workflow driftがあれば解消まで停止する。共有 `main` / `staging` は直接編集しない。
3. **候補を検証する:** 変更ファイルに対応するDocker scoped test、lint、format、`git diff --check` を実行する。認証、CORS、Worker、医院分離に触れる場合は、その境界の回帰テストと独立レビューを必須にする。migration・seed・起動依存を含む場合は、共有DBをresetせず、ユーザー承認を得た隔離環境でfresh applyとUNIQUE制約を確認する。エージェントはmigrationを自動適用しない。CIは対象product jobがSKIPではなく実際にSUCCESSしたことを確認する。
4. **外部更新と配信を分離する:** push、PR更新・置換、merge、STG deployは外部操作として明示承認後に行う。配信後は provider のcommit SHAとSTGが返すasset/revisionを照合し、対象revision・配信時刻・確認者を記録する。一致しなければ受入を開始しない。
5. **ブラウザー受入を実行する:** 認証情報や患者情報を証跡に残さず、匿名、既存セッション、復旧画面、医院選択を確認する。自然な未使用期間後の初回と直後の再読込を分け、待機表示、フォーム操作可能、認証成功の3時刻と、OPTIONS / GETの各timingを保存する。休止を強制せず、少数の通常読込から始める。
6. **判定する:** 対象revisionで機能回帰がなく、原因に対応した区間の改善を変更前後で説明できた場合だけSTG受入をPASSにする。単発値、spinner表示、security checkだけでは性能改善やp95/p99達成を宣言しない。不一致・未実行・再現不能はそれぞれFAIL / BLOCKED / UNKNOWNとして残す。

### Linear 照合・反映の着手プラン

1. **読み取り範囲を固定する:** Team `Baritech`、Project `ノア動物病院電子カルテ`、hub `BRT-4` を対象に、`PERF-STG-LOGIN`、`/login`、`/v1/me`、`preflight`、`cold start`、`Container`、PR #388 / #393 と関連SHAをタイトル・本文・コメントから検索する。
2. **既存 issue を優先する:** 直接対応する issue が見つかった場合は、現在の担当・状態・受入条件を読み取り、本書の C / D / STG 受入と対応付ける。関連語だけの issue を直接対応と推定せず、複数候補なら UNKNOWN のまま候補と相違点を記録する。
3. **反映案をローカルで作る:** 原因未確定、C / D未着手、PR #388競合、STG revision、Browser/E2E、測定結果をPASS / BLOCKED / UNKNOWNに分け、issue本文またはコメントの下書きを作る。既存 issue がなければ、重複検索結果と新規issue案を作るところまでとする。
4. **承認後に反映する:** Linearへのコメント、新規issue、状態・担当・受入条件の変更は外部書き込みとして明示承認後に実行する。repo統合やlocal/static testだけでDoneへ移さない。DoneはSTG受入結果と残作業を照合したうえでUSERが判断する。
5. **完了条件:** 対応issueのURL・ライブ状態・照合日時と、C / D / STG受入の各状態が相互に一致していること。読み取り手段が利用できない場合は照会語と対象を残し、Linear状態をUNKNOWNのままにする。

### 検証・実装の分割

| 実装単位 | 必須の確認 | 次へ進める条件 |
|---|---|---|
| C: 計測・通信経路修正 | OPTIONS/GET別の時刻、cold/warm比較。変更時はCORS・内部ルート・CSRF契約 | 原因と対策の因果を示す実測あり。環境変更・deployは別途承認 |
| D: bundle | キャッシュなし/あり、チャンク依存、操作可能時刻 | Cの残課題と切り分けた改善量あり |
| STG受入 | PR競合解消、migration/seed/env差分、対象CIの実SUCCESS、配信revision一致、匿名・既存session・復旧・医院選択、3時刻とOPTIONS/GET | scoped検証とレビュー後に外部更新承認。対象revisionのbrowser証跡あり |
| Linear | Team / Project / hub内の本文・コメント検索、既存issueとの対応、状態境界 | ライブ読取結果とローカル反映案あり。書き込みは別途承認 |

実装時は対象worktreeに結び付いたDockerで、既存の `use-auth-initial-session.test.tsx`・`LoginForm.test.tsx` と変更箇所隣接テストを指定して `npx vitest run <対象パス>` を実行する。共有axios変更時は関連interceptorテストも追加する。全体build/testや環境起動を自動実行しない。authロジック変更はReact・認証安全性の独立レビューを受ける。

STG受入は「待機表示が出る」「フォームが操作できる」「認証が成功して業務画面へ進める」の3時刻を分ける。まず通常読込の少数比較から始め、単発値でp95/p99達成を宣言しない。継続測定の件数・時間は既存のSTG測定契約で定める。既存セッション、匿名、復旧画面、医院選択の回帰が出た変更単位は先へ進めず、原因修正または承認済みの前版への切戻しを選ぶ。


## 参考・検証境界

- [Resource Timing仕様](https://www.w3.org/TR/resource-timing/) の `requestStart` は final network-request start。最終GETの応答待ちとfetch全体を区別する。
- [Fetch仕様・HTTP fetch](https://fetch.spec.whatwg.org/#http-fetch): 必要なpreflightは実リクエストより先に処理する。
- [Chrome Network Timingの説明](https://developer.chrome.com/docs/devtools/network/reference/#timing-explanation): Queueing / Stalled / Proxy negotiation / Waitingを区別する。
- [Cloudflare Containers公式SDK](https://github.com/cloudflare/containers): fetchによる起動とsleepAfterの説明。
- 既存の測定手順: [STG-PERFORMANCE-CHECKLIST.md](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md)。health / clinics の低負荷API試験はログイン画面のFCP・preflight・cold startの代替証拠にしない。
