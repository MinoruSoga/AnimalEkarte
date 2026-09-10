# Performance 調査・改善 TODO

作成・調査日: 2026-09-09（JST）／対応状況の最終照合: 2026-09-10（JST）

調査 ID: **PERF-STG-LOGIN**

対象: STG `/login` の初回表示遅延。責任者・依頼者: 曽我 稔。

ソース照合の基準点: 本整理着手時のローカル `main` `7c6a4a9d1`、最終 fetch 時の `origin/main` `423a26743`。本整理コミット自身は含めない。調査時に照合した STG bundle は `main-BnyQFmpH.js` であり、現在の STG 配信 revision は未照合。

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
| P2 | 初回 bundle の不要読込 | TODO・今回の主因ではない | 配信HTMLは charts / LIFF 等をpreload。ログインに必要な依存を実測して分離する。E1はload約0.31秒のため、23秒の主因として扱わない |
| P1 | 改善後の再計測 | 未実施 | 同一条件のwarm/cold・初回/再読込を区別し、FCPだけでなくフォーム操作可能時刻、OPTIONS/GET各時間を保存。spinnerの表示だけを「ログイン高速化完了」としない |

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

### 検証・実装の分割

| 実装単位 | 必須の確認 | 次へ進める条件 |
|---|---|---|
| C: 計測・通信経路修正 | OPTIONS/GET別の時刻、cold/warm比較。変更時はCORS・内部ルート・CSRF契約 | 原因と対策の因果を示す実測あり。環境変更・deployは別途承認 |
| D: bundle | キャッシュなし/あり、チャンク依存、操作可能時刻 | Cの残課題と切り分けた改善量あり |

実装時は対象worktreeに結び付いたDockerで、既存の `use-auth-initial-session.test.tsx`・`LoginForm.test.tsx` と変更箇所隣接テストを指定して `npx vitest run <対象パス>` を実行する。共有axios変更時は関連interceptorテストも追加する。全体build/testや環境起動を自動実行しない。authロジック変更はReact・認証安全性の独立レビューを受ける。

STG受入は「待機表示が出る」「フォームが操作できる」「認証が成功して業務画面へ進める」の3時刻を分ける。まず通常読込の少数比較から始め、単発値でp95/p99達成を宣言しない。継続測定の件数・時間は既存のSTG測定契約で定める。既存セッション、匿名、復旧画面、医院選択の回帰が出た変更単位は先へ進めず、原因修正または承認済みの前版への切戻しを選ぶ。


## 参考・検証境界

- [Resource Timing仕様](https://www.w3.org/TR/resource-timing/) の `requestStart` は final network-request start。最終GETの応答待ちとfetch全体を区別する。
- [Fetch仕様・HTTP fetch](https://fetch.spec.whatwg.org/#http-fetch): 必要なpreflightは実リクエストより先に処理する。
- [Chrome Network Timingの説明](https://developer.chrome.com/docs/devtools/network/reference/#timing-explanation): Queueing / Stalled / Proxy negotiation / Waitingを区別する。
- [Cloudflare Containers公式SDK](https://github.com/cloudflare/containers): fetchによる起動とsleepAfterの説明。
- 既存の測定手順: [STG-PERFORMANCE-CHECKLIST.md](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md)。health / clinics の低負荷API試験はログイン画面のFCP・preflight・cold startの代替証拠にしない。
