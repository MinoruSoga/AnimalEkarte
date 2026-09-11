# Performance 調査・改善 TODO

作成・調査日: 2026-09-09（JST）／対応状況の最終照合: 2026-09-10（JST）
調査 ID: **PERF-STG-LOGIN**
対象: STG `/login` の初回表示遅延。責任者・依頼者: 曽我 稔。

実行 SoT: Linear。本書は調査事実と技術判断を保持する。着手可能な開発は [todo.md](todo.md#development-tasks)、検証・測定・受入・Linear照合は [todo-verification.md](todo-verification.md#perf-stg-login) を正本とする。

## 現在の状態

- STG 配信、ブラウザー/E2E、改善後の実測、関連 Linear のライブ照合は UNKNOWN / 未実施である。
- 原因は未確定である。約22.5秒の待ちを Go API / DB の処理時間、ブラウザーだけの停滞、Container 起動のいずれとも断定しない。
- Worker 観測WIPは導入不要として保全する。根拠と再判定条件は統合検証TODOにある。
- C（通信経路対策）とD（bundle変更）は、原因を示す新しい測定証跡が出るまで着手しない。

## 開発着手の判断（2026-09-11）

`PERF-DEV-OBSERVATION` / `PERF-DEV-MITIGATION` / `PERF-DEV-BUNDLE` は、現時点では開発キューへ採用しない。測定待ちを開発READYへ読み替えず、再判定は統合検証TODOで追跡する。旧IDごとの理由は [開発タスクの裁定記録](docs/work/development-task-decisions.md) を参照する。**ローカル開発 READY はなし。**

- **観測WIP**: 全proxy要求への常時ログは採用しない。既存の `make test-worker ARGS='backend/worker/index.test.ts'` はrootをmountするDocker runnerであり、「Docker runnerがない」という前の保留理由は誤りだった。Makefileの静的契約8項目を確認済み。WIPのproxy実経路と型検査の証明は別途未完了である。
- **通信経路**: CORS/CSRF変更、edge OPTIONS応答、`sleepAfter`変更を選べる因果証跡はない。現行 `main` のpending UIは [router.tsx](frontend/src/app/router.tsx)、[app-routes.tsx](frontend/src/app/routes/app-routes.tsx)、[AuthProvider.tsx](frontend/src/features/auth/components/AuthProvider.tsx) に存在し、同じUI実装を重複起票しない。
- **bundle**: Vite設定にcharts/LIFFのchunkがあり、ローカルdistにもpreloadがあるが、distの生成revisionはUNKNOWN。現行SHAの転送・parse/executeへの寄与は未測定なので、一括chunk再編やfeature境界を壊すdeep importは採用しない。

## 調査時に確定した事実と境界

当時はセッション確認が終わるまでログイン画面を描画しなかったため、通信待ちが白画面の待ち時間になっていた。

未確定なのは、通信全体約22.7秒のうち約22.5秒を占めた、最終 GET の接続開始前の待ち時間の内訳である。最終 GET の送信から応答開始は約116msだった。「Go API / DB が23秒処理した」とは言えない。一方、CORS preflight（OPTIONS）が先行する構成なので、ブラウザー内だけの停滞とも断定しない。

## 実測記録

### E1: 遅延が発生した既存ブラウザー記録

- navigation開始: 2026-09-09 14:18:14.373 JST。
- HTML TTFB: 49.2ms、DOMContentLoaded: 303.1ms、load: 309.1ms、FCP: 23,548ms。
- `/api/v1/me`: HTTP 401、HTTP/3、全体22,689.3ms。Cookie・認証情報は採取していない。

| 項目 | navigation開始からの時刻 | 区間 |
|---|---:|---:|
| fetchStart | 309.1ms | — |
| domainLookupStart | 22,855.4ms | fetchStartから22,546.3ms |
| domainLookupEnd | 22,855.8ms | DNS 0.4ms |
| connectEnd | 22,880.1ms | 接続24.3ms |
| requestStart | 22,880.3ms | — |
| responseStart | 22,996.1ms | 最終GETの応答待ち115.8ms |
| responseEnd | 22,998.4ms | 本文受信2.3ms |

`workerStart=0`。当時の OPTIONS、request ID、接続交渉、Container 起動状態、配信revisionは保存されていない。

### E2: 同日の再確認

通常設定の同一Chromeプロファイルでの2回は、`/me` 全体221.4ms / 190.0ms、FCP 944ms / 768ms、いずれも401だった。これは遅延が再現しなかった証拠であり、修正済み、cold start解消、p95/p99達成の証拠ではない。

### E3: 認証情報なしの単発HTTP確認

GET STG `/login` は281ms、Cookieなし GET API `/api/v1/me` は254ms、指定Originの OPTIONS API `/api/v1/me` は240msだった。これはブラウザーの認証済み経路、HTTP/3、CORS自動preflightを再現しない。

## 調査時に確認した通信経路

1. [axios.ts](frontend/src/lib/axios.ts) は `X-Requested-With` / `X-Request-ID` を付ける。`X-Requested-With` はCSRF防御のため、高速化目的で無条件に削除しない。
2. [Worker](backend/worker/index.ts) は通常リクエストをContainerへ転送する。OPTIONSのedge即時応答はなく、Container休止中の起動待ちの可能性はあるが、E1での証拠は未取得である。
3. [CORS middleware](backend/internal/middleware/cors.go) はOPTIONSへ204とmax-age 86400を返す。
4. [auth middleware](backend/internal/middleware/auth.go) はtoken不在ならDB照会前に401を返す。有効tokenの認可経路へ一般化しない。

## 参照

- [統合検証TODO](todo-verification.md#perf-stg-login)
- [STG パフォーマンス測定チェックシート](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md)
- [パフォーマンス測定・プロファイリングガイド](docs/ops/testing/PERFORMANCE_PROFILING.md)
