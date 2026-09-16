# Performance 調査・改善 TODO

最終照合: 2026-09-15。調査 ID: **PERF-STG-LOGIN**。対象は STG `/login` の初回表示遅延。責任者・依頼者: 曽我 稔。

未完了の測定・受入は [todo-verification.md](todo-verification.md#perf-stg-login)。新たな実装が必要になったら [todo-issue.md](todo-issue.md) に範囲を確定する。本書は判断に必要な技術記録のみを保持する。

## 現在の判断

- 初回遅延の原因と改善後の測定結果は未確定。通信全体約22.7秒を Go / DB 処理時間と断定しない。
- Worker 観測は `b42c00ccb` で既にコミットされ、現在の [index.ts](backend/worker/index.ts) に `buildProxyObservation` と通常 proxy のログ出力が存在する。**「未導入 WIP を保全」「commit・deploy しない」という旧記述は現状と一致しないため削除した。** STG 配備対象 `d337f016` にも同じ観測経路が含まれる。
- 残るのは、現行観測の必要性・出力範囲・実 proxy 回帰・型検査と、ブラウザ/provider の時刻対応の確認。ログが存在するだけで原因特定や性能改善を完了にしない。
- pending UI は既存実装を利用する。CORS / CSRF、edge OPTIONS、Container 設定、bundle 分割を、因果証拠なしに変更しない。

## 次に確認すること

1. 通常の `/login` 読込で OPTIONS / GET の各区間、FCP、フォーム操作可能時刻を測る。
2. 同じ時刻の provider 受付・forwarding・Container 起動証拠と対応づける。
3. 現行の常時ログが必要かを再判定する。必要なら対象と出力を限定し、不要なら撤去を別実装単位にする。
4. 現行 proxy 経路と新規テストの型検査を確認する。`index.test.ts` は `tsconfig.test.json` の明示 include にないため、検証範囲を確認して補完要否を決める。

対象・承認・完了条件は [検証 TODO](todo-verification.md#perf-stg-login)。旧候補の裁定は [履歴](docs/work/development-task-decisions.md) であり、現在の導入状態は上記を正とする。

## E1: 2026-09-09 の遅延記録（過去の測定）

HTML TTFB 49.2ms、DOMContentLoaded 303.1ms、load 309.1ms、FCP 23,548ms。`/api/v1/me` は401、HTTP/3、全体22,689.3ms。

| 区間 | 所要時間 |
|---|---:|
| fetchStart → 最終 GET の DNS 開始 | 22,546.3ms |
| DNS | 0.4ms |
| 接続 | 24.3ms |
| requestStart → responseStart | 115.8ms |
| 本文受信 | 2.3ms |

`workerStart=0`。当時の OPTIONS、相関 ID、接続交渉、Container 起動、配信 revision は未保存。約22.5秒は最終 GET の接続開始前であり、Worker 内の所要時間だけでは説明できない。

## E2 / E3: 再現しなかった記録

同日の通常 Chrome 2回では `/me` 221.4 / 190.0ms、FCP 944 / 768ms（いずれも401）。単発 HTTP は `/login` 281ms、Cookie なし `/me` 254ms、OPTIONS 240ms。遅延が再現しなかった記録であり、認証済みブラウザ経路・cold start 解消・p95/p99 を証明しない。

## 測定で維持する境界

- [axios.ts](frontend/src/lib/axios.ts) の `X-Requested-With` は CSRF 防御。高速化のために無条件で削除しない。
- browser の接続待ち、Worker forwarding、Container 起動、Go request latency を分ける。
- 秘密・raw URL/query・本文・IP・個人情報を観測証拠へ出さない。
- Docker の `make test-worker ARGS='backend/worker/index.test.ts'` は既存 runner。runner 不在を理由にしない。runtime・型検査・STG 性能受入は別に検証する。

参照: [測定チェックシート](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md) · [プロファイリングガイド](docs/ops/testing/PERFORMANCE_PROFILING.md)
