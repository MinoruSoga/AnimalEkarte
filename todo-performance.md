# Performance evidence and history

> Performance work items are tracked in Plane. The migration crosswalk is [here](docs/work/plane-md-migration-20260923-receipt.md); measurements and historical evidence below remain local.

最終照合: 2026-09-22（同日 curl 実測でコールドスタートを直接観測し **E4** として記録。原因区間を確定し改善候補を整理。SLACK-LATENCY計測票差分は別記録を保持）。主調査 ID: **PERF-STG-LOGIN**。対象は STG `/login` 初回表示遅延に始まり、ユーザー報告により STG 全域のページ読み込み遅延へ拡大。責任者・依頼者: 曽我 稔。

未完了の性能作業と現在状態はPlane（`PERF-E5-STG-DEPLOY-VERIFY`, `PERF-V-MITIGATION`, `PERF-V-BUNDLE`, `SLACK-LATENCY`）に移行済み。本書は判断に必要な技術記録のみを保持する。

## 現在の判断

- 初回遅延の主因は **Cloudflare Containers の scale-to-zero コールドスタート** と実測で確定（E4）。`sleepAfter = "10m"` でアイドル10分後にコンテナが停止し、全 API リクエストは単一の共有コンテナ（`getContainer(env.API_CONTAINER)`、名前なし=既定インスタンス）の起動完了を待つ。`instance_type: "basic"`（1/4 vCPU / 1GiB）で起動は実測 約6〜16秒。
- E1 の約22.5秒（最終 GET の接続開始前）は Container 起動待ちと整合する。「通信全体を Go / DB 処理時間と断定しない」判断は維持し、温間 TTFB 0.24〜0.25 秒から Go/DB 自体の遅さではないことを確認済み。
- Worker 観測 `container_fetch_timing`（`b42c00ccb` 由来）は継続有効。改善実装後の区間別再測定に使う。
- 温間状態の認証済み API レイテンシは **E5 で実測済み**。`/v1/clinics` 0.47s、`/v1/me` 0.85s（認証キャッシュ適用後・暖機）。E4 の UNKNOWN は解消。`/v1/pets` 検索等の重めクエリは未測定のまま。
- pending UI は既存実装を利用する。bundle 分割（BUNDLE）は転送/parse/execute の寄与が未測定のため DEFERRED のまま。edge OPTIONS と Container 設定は因果証拠（E4）が取れたため実装候補へ昇格したが、採用・範囲・受入は別途確定する。

## 改善プラン（候補・2026-09-22）

対象区間は E4 で確定したコールドスタート待ちと、プリフライト／リトライ／N+1 による増幅。実装単位の確定・受入条件・担当は [todo-issue.md](todo-issue.md) / [todo-verification.md](todo-verification.md#perf-stg-login) に移す。ここでは技術的根拠と前提のみ記録する。

| # | 案 | 削減する区間 | 前提・注意 |
|---|----|------------|------------|
| 1 | keep-alive：`sleepAfter` 延長、または営業時間中の定期 `/health` ping | コールドスタート待ち（支配的区間）そのもの | 既存 cron（01:00/06:00/11:00/17:00 UTC）は `SCHEDULER_NAME` の named コンテナを起こすのみで、既定名の API コンテナは温まらない。ping は `getContainer(env.API_CONTAINER)` の既定インスタンスに届く経路（通常 `/health` GET で可）が必要。scale-to-zero のコスト想定（AC-5 検証目的）とのトレードオフは STG 運用判断 |
| 2 | OPTIONS を Worker エッジで応答（Container へ転送しない） | プリフライトのコールド待ち（実測 6 秒）と温間時の往復 1 回分 | CORS allowlist は静的（`CORS_ALLOWED_ORIGIN` vars）で Worker 側に複製可能。`Allow-Headers/Methods/Max-Age` の契約は `internal/middleware/cors.go` と一致させる。GET 本体のコールド待ちは残るため単独では不十分 |
| 3 | axios の 502–504 リトライ範囲・回数の見直し | 起動中の二重待ち（1秒・2秒バックオフ＋再送） | `isServerError` は 502–504 で、Worker が起動失敗時に返す 503 `service_unavailable` を含む。起動待ち中の再送はコンテナ起動を早めない。503 を対象外にするか回数を絞るかは、起動失敗時の UX（速く失敗を返す）とセットで判断 |
| 4 | `ownerLoader` の N+1（飼主 + ペット個別 GET）解消 | 飼主詳細のリクエスト数（ペット N 頭で N+1 本、各 URL が別プリフライトキー） | `/v1/pets/{id}` は URL ごとにプリフライトキャッシュが別キー。バックエンドで pets 同梱またはバッチ取得にする。温間実測が出るまでは二次因扱い |
| 5 | `instance_type` 引き上げ（basic → 上位） | コールドスタートの起動時間そのもの | コスト増。案1と排他ではない（起動を速くしても sleep 自体は残る） |
| 6 | プリフライト自体の削減 | 温間時の往復 1 回分 × リクエスト数 | `X-Requested-With` は CSRF 防御で固定（「測定で維持する境界」参照）。GET から `X-Request-ID`・`X-Clinic-ID` 等を外せば simple request 化してプリフライト不要になるが、`Content-Type: application/json` を伴う POST/PUT/PATCH は safelist 外のため依然 preflight が必要。削減効果は GET 系に限定される |

補足: 案2・3・4・6 は「温間でも遅い」場合の二次因にも効くが、E4 時点では温間 API は速い（0.25 秒未満）ため優先度は案1が最大。ブラウザのプリフライトキャッシュは `(origin, URL, method)` 単位で、`Access-Control-Max-Age: 86400` を返しても Chrome 系は約2時間で切り捨てられるため、Max-Age 引き上げでは増幅は解消しない。

### 費用影響（Cloudflare 公式 pricing 確認・2026-09-22）

Containers は稼働時間課金（10ms 単位、Workers Paid $5/月に含む）。単価: メモリ $0.0000025/GiB-秒（25 GiB-h/月込み、provisioned 課金）、CPU $0.000020/vCPU-秒（375 vCPU-分/月込み、2025-11 改訂で **アクティブ使用分のみ**）、ディスク $0.00000007/GB-秒（200 GB-h/月込み）。現行 `basic` = 1/4 vCPU / 1 GiB / 4 GB。

| 案 | 増分の目安（STG・単一コンテナ前提） |
|---|---|
| 1. keep-alive | **増加するが小さい**。24/7 常時起動 ≈ $7/月（メモリ $6.4 + ディスク $0.7、CPU はアイドル時ほぼゼロ）。営業時間のみ（約264h/月）≈ $2.4/月 |
| 2. edge OPTIONS | 実質ゼロ（Workers Paid のリクエスト枠内、STG 量は誤差） |
| 3/4/6. リトライ・N+1・プリフライト削減 | 微減（リクエスト数減） |
| 5. instance_type 引上げ | 稼働時間比例。basic→standard-1（4GiB/8GB）で 24/7 なら約 $28/月。sleep 維持なら増分は小さい |
| Redis系（現時点で不採用候補） | コールドスタートには無効。温間クエリの実測遅延が確認されてから検討。Upstash は CF 請求外、KV/Cache API はこの規模では実質無料。マルチテナント（clinic_id）分離の安全不変条件があり、エッジキャッシュには越境リーク設計が必要 |

結論として費用が問題になるのは案1と案5のみで、案1でも月 $2〜7 程度。scale-to-zero が節約しているのは月数ドルであり、対価として6〜16秒の起動待ちが発生している。

## 次に確認すること

1. 認証済みブラウザで `/owners` 等の実ページについて、コールド（10分超アイドル後）と温間の区間別時間を測る（既存チェックシートの run 票を使用）。
2. Workers Logs の `container_fetch_timing` で実リクエストの `duration_ms` を集計し、温間でも遅い API（二次因 = 実クエリやレスポンスサイズ）があるか切り分ける。未認証 curl では実クエリ時間を測れないため、この確認なしに「温間は全て速い」と断定しない。
3. 現行の常時ログが必要かを再判定する。必要なら対象と出力を限定し、不要なら撤去を別実装単位にする。
4. 採用した改善案の実装後、同じ条件で再測定し E5 として記録する。

対象・承認・完了条件は [検証 TODO](todo-verification.md#perf-stg-login)。旧候補の裁定は [履歴](docs/work/development-task-decisions.md) であり、現在の導入状態は上記を正とする。

### 測定準備で決めること

ローカルで [測定チェックシート](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md) の run 票を作れる。匿名/既存session/復旧/医院選択ごとに、対象build、ブラウザ、cache条件、通常遷移/再読込、回数・間隔、同一時刻窓、provider証拠の取得担当、停止条件と保存先を記入する。実測前の結果欄は未実行。承認範囲に測定回数と対象操作を含め、承認のない外部trafficや負荷試験を開始しない。

比較表の列は「run/case、OPTIONS/GET各区間、FCP、待機表示、フォーム操作可能、認証成功、Worker受付/forwarding、Container起動証拠、時刻対応の可否」。非公開の生証拠は保管規則に従い、共有表には秘密除去した相対時刻と分類だけを置く。対応しない区間は UNKNOWN のまま、ブラウザ全体の待ちとWorker内時間を足し合わせて二重計上しない。

**SLO の採用値は未確定。** チェックシートの API p95 500ms / 初回操作可能1.5秒などは提案例であり、合否判定へ自動採用しない。一方、[既存STG k6](load-tests/k6-cf-stg-sustained.js) の p95 3秒・失敗率5%未満は `/health` と `/api/v1/clinics` を3 VUで測るスクリプト閾値で、ブラウザ `/login` の受入値ではない。今回の遅延調査にその負荷試験を追加せず、初回表示の区間測定から始める。

準備完了は6単位それぞれの入力・出力・未確定条件がケース票に揃った時点。MITIGATION は因果区間、BUNDLE は転送/parse/executeの寄与が分かるまで実装 DEFERRED。測定できても採用 SLO が未合意なら、計測完了と性能受入完了を分ける。

## SLACK-LATENCY: 治療数量の反映待ち

> Task detail migrated to Plane `EMR-104` and verified by readback. Historical/evidence material remains in linked source records.

## E1: 2026-09-09 の遅延記録（過去の測定）

HTML TTFB 49.2ms、DOMContentLoaded 303.1ms、load 309.1ms、FCP 23,548ms。`/api/v1/me` は401、HTTP/3、全体22,689.3ms。

| 区間 | 所要時間 |
|---|---:|
| fetchStart → 最終 GET の DNS 開始 | 22,546.3ms |
| DNS | 0.4ms |
| 接続 | 24.3ms |
| requestStart → responseStart | 115.8ms |
| 本文受信 | 2.3ms |

`workerStart=0`。当時の OPTIONS、相関 ID、接続交渉、Container 起動、配信 revision は未保存。約22.5秒は最終 GET の接続開始前であり、Worker 内の所要時間だけでは説明できない。→ E4 で Container 起動待ち（コールドスタート中にブラウザ側が接続を待つ区間）と整合することを確認。

## E2 / E3: 再現しなかった記録

同日の通常 Chrome 2回では `/me` 221.4 / 190.0ms、FCP 944 / 768ms（いずれも401）。単発 HTTP は `/login` 281ms、Cookie なし `/me` 254ms、OPTIONS 240ms。遅延が再現しなかった記録であり、認証済みブラウザ経路・cold start 解消・p95/p99 を証明しない。→ 温間状態なら速いことは E4 とも一致する。

## E4: 2026-09-22 curl 実測（コールドスタートの直接観測）

対象: `https://api.stg.noah-karte.com`（STG API、Cloudflare Workers → Containers → PlanetScale）と `https://stg.noah-karte.com`（Vercel）。未認証のため実クエリ時間は含まない。ユーザー報告「ほとんどのページで読み込みが遅い」を受けて実施。

| リクエスト | 計測 | 解釈 |
|---|---|---|
| `GET /health`（10分超アイドル後の初回） | TTFB 15.69s | Container 起動待ち |
| `GET /health`（2・3回目） | TTFB 0.24s / 0.25s | 温間は速い。Go/DB 自体の遅さではない |
| `OPTIONS /api/v1/me`（コールド） | TTFB 5.96s | プリフライトも同じコンテナ起動を待つ |
| `GET /api/v1/me`（直後） | TTFB 0.24s（401） | 温間 |
| `GET /api/v1/pets?page=1&limit=20` | TTFB 0.25s（401） | 温間。認証前のため実クエリ未含む |
| `GET /owners`（Vercel） | TTFB 0.32s（2.5KB shell） | Vercel 側は速い。遅延は API 側 |

構造上の確定事項（再確認済み）:

- `backend/worker/index.ts`: `sleepAfter = "10m"`、`getContainer(env.API_CONTAINER)`（名前なし = 全トラフィックが既定の単一インスタンスに集中）。
- `backend/wrangler.jsonc`: `instance_type: "basic"`（1/4 vCPU / 1GiB）、`max_instances: 3`。DB は PlanetScale 直結（Hyperdrive 非経由、`DB_SSL_MODE=verify-full`）、プール上限 `DB_MAX_OPEN_CONNS=10`/`DB_MAX_IDLE_CONNS=5`。コールドスタートでは Go 起動に加えてプール空の状態から PlanetScale への TLS 接続確立が入る。
- Worker は OPTIONS を含む全リクエストを `containerFetch` でコンテナへ転送する薄いプロキシ。エッジで OPTIONS を返さないため、プリフライトも起動待ちに入る。
- axios は全リクエストに `X-Requested-With`（CSRF）・`X-Request-ID`・選択中は `X-Clinic-ID` を付与し、`Content-Type: application/json` も safelist 外のため、全 API コールが preflight 対象。プリフライトキャッシュは `(origin, URL, method)` 単位で `Access-Control-Max-Age: 86400` を返すが Chrome 系は約2時間で切り捨て。
- axios は GET の 502–504（Worker の起動失敗時 503 `service_unavailable` を含む）を最大2回・1s/2s バックオフでリトライする。
- `ownerLoader` は飼主取得後にペットごと `/v1/pets/{id}` を並列取得（N+1。各 URL が別プリフライトキー）。
- 既存 cron は `SCHEDULER_NAME` の named コンテナを起こすのみで、既定名の API コンテナは温めない。

「ほとんどのページが遅い」と整合する解釈: STG はアクセスがまばらで、操作の合間に10分を超える間隔があくたびコンテナが停止する。次のページ遷移・API 呼び出しが起動待ちになり、プリフライト＋リトライ＋N+1 が体感をさらに悪化させる。

## E5: 2026-09-22 実装後の再測定（配置制約＋認証キャッシュ＋sleepAfter延長）

E4 以降にユーザー報告で発覚した追加原因と、採用した改善・その実測を記録する。E4 の「温間は速い」は **Container↔PlanetScale 間の越洋 RTT を未認証 401 では測れていなかった** ため部分的な結論だった。認証済み curl で実測すると温間でも `/v1/clinics` ~1.0s、`/v1/me` ~1.5s、`login` ~4.1s で、遅延はコールドスタートだけではなかった。

### 追加で確定した原因

- **コンテナ↔DB の地理分離**: DB は `ap-northeast-2.pg.psdb.cloud`（PlanetScale = AWS ソウル）。当初コンテナは `ewr01`（米東部）で稼働し、認証ミドルウェアが毎リクエスト staff→account→assignments→clinics を逐次 DB 再検証するため、1 往復 ~190ms×クエリ数が積み上がっていた。
- **インスタンス配置はブート/ロールアウト毎に再抽選される**: 同一 DO 名 `cf-singleton-container` のまま `ewr01→bom09→maa01→sin14→bom09` とドリフトを観測。DO の `locationHint`/改名は初回 DO 作成時のみ効く best-effort で、**インスタンス再配置には効かない**（PR #423 の `api-apac-ne-v1` 実験は bom09 着地で撤回・PR #425）。
- **`constraints.cities` はこのアカウントで利用不可**: デプロイが `VALIDATE_INPUT: City-level placement requires INTERNAL or CITIES_CONSTRAINT capability` で失敗（run 35749806541）。メトロ粒度のピン留めはできない。
- `scheduling_policy: "regional"` を wrangler.jsonc に記載したが `wrangler containers info` は `default` を返し続ける。API が受理したか不明。

### 採用した変更（staging にマージ済み）

| PR | 変更 |
|---|---|
| #424 | STG限定の認証 resolver キャッシュ `CURRENT_ACCESS_CACHE_TTL_SEC=30`（vars→envVars→`os.Getenv`→`composition_auth.go` の env ゲート。未設定/0/負値ならキャッシュ無しで本番は従来通り）。`sleepAfter` 10m→1h |
| #426 | `containers[].constraints.regions = ["APAC"]` — 配置抽選を APAC メトロに限定（無料） |
| #427 | `Dockerfile.production` に `LABEL rollout="1"` — イメージ差分で新バージョンを強制ロールアウトし即時再配置を起こす仕掛け。`verify-agent-task.py` に Dockerfile の scoped 検証（`docker build --check`）を追加 |
| #428 | `scheduling_policy: "regional"`（API 上は default のまま。残置するが効果未確認） |
| #429→#430 | `cities` 試行→ケイパビリティ不足で失敗→撤回 |

### 実測（認証済み・暖機・日本から）

| エンドポイント | E4 時点（ewr01） | 最悪時（bom09） | E5 現在 |
|---|---|---|---|
| `GET /health`（DB無し） | 0.25s | ~1.0s | **0.20s** |
| `GET /v1/clinics` | ~1.0s | ~1.5s | **0.47s** |
| `GET /v1/me` | ~1.5s | ~2.2s | **0.85s** |
| `POST /v1/login` | 4.1s | 5.4s | **3.5s** |

改善の内訳は越洋 RTT 削減（現行インスタンスは日本近辺と推定）＋認証キャッシュで resolver の逐次 DB 往復が消失＋コールドスタート頻度低下（sleepAfter 1h）の複合。

### 残存事項・運用メモ

- 配置は依然ブート毎の抽選。APAC 制約は ENAM/EEUR の最悪ケースを防ぐだけで日本着地は保証しない。悪い着地（bom/sin/maa）を引いたら `LABEL rollout` をインクリメントして再デプロイすると再抽選できる。
- `login` 3.5s は bcrypt×1/4 vCPU 由来でリージョンと独立。`instance_type` 引上げ（案5）が残るが実コスト増。
- 認証キャッシュは 3148d229f の認可ギャップ修正とトレードオフ（権限変更が最大30秒遅延）。STG限定のため vars は本番 `wrangler.production.jsonc` には設定しない。
- 費用: sleepAfter 延長のみ課金に触れる。basic 1台の課金は ~$0.01/h（メモリ+ディスク支配、CPU はアイドル時ほぼゼロ）。通常デモ利用で月+数十〜百円、常時稼働化しても ~$7/月が上限。

## 測定で維持する境界

- [axios.ts](frontend/src/lib/axios.ts) の `X-Requested-With` は CSRF 防御。高速化のために無条件で削除しない。
- browser の接続待ち、Worker forwarding、Container 起動、Go request latency を分ける。
- 秘密・raw URL/query・本文・IP・個人情報を観測証拠へ出さない。
- 将来の変更時は既存依存を使い、対象候補を mount した隔離 Docker の scoped Vitest と worker typecheck を使う。依存インストールを伴う旧 umbrella コマンドは今回実行しない。STG 性能受入はローカル検証とは別。

参照: [測定チェックシート](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md) · [プロファイリングガイド](docs/ops/testing/PERFORMANCE_PROFILING.md)

## E5: 2026-09-23 perf-e5-residual キャンペーン結果（revision 2）

証拠の正本は `reports/perf-e5-residual-20260923/`。測定値は全て**計測時点の現行 STG 配信版**のもの。キャンペーンのコード変更は claim/PERF-E5-* ブランチまたは作業ツリー WIP にあり STG 未配備であり、下記の値に変更後の効果は含まれない。n=1・n=5 の単発観測であり p95/p99・SLO 達成を主張しない。

### 実装単位（STG 未配備のコード変更）

| 単位 | 結果 |
|---|---|
| EDGE-OPTIONS | OPTIONS preflight を Worker edge で応答する経路を実装。ただし cookie 認証の simple request では preflight が発生せず、browser 証拠（client-trace・stg-acceptance ともに OPTIONS 0 件）では同改善は未励起 — 正直な記録として併記する |
| AXIOS-RETRY | axios の retry 動作を縮小 |
| N1-PETS | owner loader を owner 1リクエスト + clinic_ids 付き owner スコープのページネーションへ変更し、ペット毎 detail fan-out を解消 |
| AUTH-1RTT | `clinics.is_active` を resolver の JOIN へ畳み込み、非 admin の current-access を 1 RTT 化 |
| STG-LOGIN | `AcceptSharedPassword` ゲートを bcrypt より前へ移動。staging のみ、production は不変 |

### 測定・判定単位

| 単位 | 取得証拠 | 主な値 |
|---|---|---|
| STG-MEASURE | [stg-measure](reports/perf-e5-residual-20260923/stg-measure/README.md)（curl、n=5 warm、中央値のみ） | health 0.199s / me 0.944s / clinics 0.572s / pets 1.108s / pets検索 1.100s / pets+include_deceased 1.103s。login 一回 3.33s |
| CLIENT-TRACE | [client-trace](reports/perf-e5-residual-20260923/client-trace/README.md)（headless Chrome、n=1×2条件） | `/login` FCP cold 692ms / warm 88ms、DCL 663/57ms、long task 0、OPTIONS 0。唯一の API 呼出は未認証 `/me` 401（cold 488ms） |
| CF-EVENTS | [cf-events](reports/perf-e5-residual-20260923/cf-events/README.md)（wrangler tail + cf-ray 相関、20秒5リクエスト） | `container_fetch` 866–3848ms が支配的。edge(KIX)+Worker shim は約60–115ms、Worker→DO は 5–7ms。稼働 instance は `maa01`（ingress は KIX、PlanetScale は ap-northeast-2）。`scheduling_policy` は deployed=`default` vs config=`regional` の乖離を記録 |
| STG-ACCEPTANCE | [stg-acceptance](reports/perf-e5-residual-20260923/stg-acceptance/README.md)（4ケース、n=1。非SLO証拠） | 匿名 `/login` 操作可能 +1163ms。既存 session は `/v1/me` 1475ms にゲート。復旧は +689ms で login フォーム。login POST 3882ms → 認証 UI 205ms。医院選択ステップは存在せず `mainClinicId` 自動選択。全ケース OPTIONS 0 |
| OBS-DECISION | [obs-decision](reports/perf-e5-residual-20260923/obs-decision/README.md) | `container_fetch_timing` 常時ログは **KEEP**。STG・production draft とも `head_sampling_rate: 1` を維持し、本番トラフィック実測後の再評価トリガーのみ記録 |

残る MITIGATION と BUNDLE の作業状態・次の一手は Plane の `PERF-V-MITIGATION` / `PERF-V-BUNDLE` に移行済み。ここでは当時の計測判断を履歴として保持する。配置の手順は [STG runbook](docs/ops/infra/staging/runbook.md) を参照。

## E6: 2026-09-23 perf-e5-postdeploy-verify キャンペーン結果（revision 1・デプロイ後検証）

証拠の正本は `reports/perf-e5-postdeploy-verify-20260923/`。E5 実装単位を含む perf マージコミット `453be4ecc`（2026-09-22T18:41:12Z）の STG 配信後に実測。配信版は worker `15a85636`（2026-09-23T04:09:34Z 作成）・コンテナ v70（観測窓の途中で v71・`sin14` へ再作成）。n=1・n=5 の単発観測であり p95/p99・SLO 達成を主張しない。

### 単位別結果

| 単位 | 取得証拠 | 主な値・判定 |
|---|---|---|
| DEPLOY-VERIFY | [deploy-verify](reports/perf-e5-postdeploy-verify-20260923/deploy-verify/README.md) | **SERVES-NEW-BUILD**。serving worker `15a85636`（04:09:34Z）とコンテナ v70（更新 04:12:44Z・新規稼働 instance 04:10:55Z `maa01`）は perf コミットを後置。OPTIONS が edge で 204（Go middleware ヘッダ無し、TTFB 58–304ms）、GET は Go ヘッダ全件でコンテナを通過 |
| OPTIONS-EDGE | [options-edge](reports/perf-e5-postdeploy-verify-20260923/options-edge/README.md)（6 probe matrix） | **EDGE CONTRACT VERIFIED**。allowlisted origin は ACAO echo・ACAC=true・Max-Age 86400・TAO・`Vary: Origin` の完全セット、非 allowlisted は ACAO/TAO/Vary 無しの 204、`/_internal` は worker guard で 404、GET control は Go ヘッダ付きでコンテナ通過。OPTIONS 中央値 ~60ms |
| WARM-MEASURE | [warm-measure](reports/perf-e5-postdeploy-verify-20260923/warm-measure/README.md)（curl、n=5 warm、中央値） | warm 中央値は E5 baseline と統計的に同一（全端末 ±0.01s 内）。対照表は下記。login 単発 1.975s |
| LOGIN-MEASURE | [login-measure](reports/perf-e5-postdeploy-verify-20260923/login-measure/README.md)（login ×3、15s 超間隔） | 中央値 **1.471s** vs E5 ~3.33s（**−56%**）。bcrypt-skip（`AcceptSharedPassword` を bcrypt より前へ）は方向的に確認。帰属は warmth・初回ヒット費用と不可分で bcrypt 単独寄与は UNKNOWN |
| PLACEMENT | [placement](reports/perf-e5-postdeploy-verify-20260923/placement/README.md) | capture 時点の稼働 instance は依然 `maa01`。`scheduling_policy` deployed=`default` vs config=`regional` の乖離は継続。`constraints.regions=["APAC"]` は live だが India メトロを除外できない（`cities` はケイパビリティ不足で利用不可）。**LABEL rollout 再抽選を推奨** |
| CF-EVENTS | [cf-events](reports/perf-e5-postdeploy-verify-20260923/cf-events/README.md)（wrangler tail + cf-ray 相関、5 リクエスト） | `container_fetch_timing` は post-deploy でも 5/5 に存在。**1001–1988ms** vs E5 866–3848ms（上限が約半分に収束）。edge+worker ~57–63ms、Worker→DO +9–11ms。窓の途中で稼働 instance が `maa01`→`sin14`（v71）へ再作成 |
| BROWSER-PAGES | [browser-pages](reports/perf-e5-postdeploy-verify-20260923/browser-pages/README.md)（実 Chrome・認証済み `/owners/301164` ×2 run） | **N+1 解消を実ブラウザで確認**: owner detail 1 読込につき owner スコープの `GET /api/v1/pets?owner_id=301164…` がちょうど 1 本、per-pet fan-out 0。document TTFB ~15ms。**新規所見**: `accountings?owner_id=` が最遅 API（1.9–2.7s） |
| STG-ACCEPTANCE | [stg-acceptance](reports/perf-e5-postdeploy-verify-20260923/stg-acceptance/README.md)（4 ケース、n=1・非 SLO 証拠） | 匿名 `/login` FCP 320ms・フォーム操作可能 +972ms。login POST **1451ms**（E5 3882ms、**−62.6%**）→ 認証 UI 54ms。既存 session の `/v1/me` ゲート 1100.7ms（E5 1475ms、−25%）。復旧は 401→`/login` リダイレクト 304ms・フォーム +613ms。医院選択は引続き `mainClinicId` 自動選択でステップ非存在。全ケース OPTIONS 0 |

### E5 との対照

| 指標 | E5（デプロイ前・perf-e5-residual） | E6（デプロイ後） | 差分・判定 |
|---|---|---|---|
| `GET /health` warm 中央値 | 0.199s | 0.202s | +0.003s（同一視） |
| `GET /v1/me` warm 中央値 | 0.944s | 0.946s | +0.002s（同一視） |
| `GET /v1/clinics` warm 中央値 | 0.572s | 0.568s | −0.004s（同一視） |
| `GET /v1/pets` warm 中央値 | 1.108s | 1.100s | −0.008s（同一視） |
| `GET /v1/pets?q=` warm 中央値 | 1.100s | 1.104s | +0.004s（同一視） |
| `GET /v1/pets?include_deceased` warm 中央値 | 1.103s | 1.103s | 0.000s（同一視） |
| `POST /v1/login`（curl 中央値） | ~3.33s | **1.471s** | **−56%** |
| `POST /v1/login`（browser、stg-acceptance case4） | 3882ms | **1451ms** | **−62.6%** |
| `GET /v1/me` 既存 session ゲート（browser） | 1475ms | 1100.7ms | −25% |
| `container_fetch_timing` レンジ | 866–3848ms | 1001–1988ms | 上限が約半分に収束 |
| edge+worker オーバーヘッド | ~59–115ms | ~57–63ms（warm） | 縮小 |
| OPTIONS preflight | コンテナ経由（コールド時は起動待ち ~6s）・browser 実測 0 件 | **edge で 204・中央値 ~60ms**・browser 実測は依然 0 件 | 契約面は検証済。cookie 認証の simple request では励起しない（下記） |
| 稼働 instance 配置 | `maa01` | `maa01`（capture 時）→ 観測窓内で `sin14` へ再作成 | APAC 内ドリフト継続 |
| owner detail の pets fan-out | N+1 解消を実装（実機未確認） | **実ブラウザで owner スコープ 1 リクエストを確認** | 実機検証済み |

### 残存事項（E6 時点）

- **AXIOS-RETRY**: 実装・配備済みだが**フィールド観測は除外**——観測窓に 503 ストーム（起動失敗応答）が発生せずリトライ経路は未励起。**unverified-in-field** であり失敗ではない。確認は 503 発生時の再観測か、合成 fault 注入の別単位で行う。
- **INSTANCE-TYPE（案5）/ MITIGATION / BUNDLE**: いずれもトリガー付き DEFERRED のまま（INSTANCE-TYPE は login 等の CPU 拘束区間の更なる短縮要請と実コスト増の权衡、MITIGATION は因果区間の確定、BUNDLE は転送/parse/execute 寄与の実測）。作業状態の正本は Plane。
- **SLACK-LATENCY**: ユーザーレーン（Plane `EMR-104`）。本キャンペーンの対象外。
- **PERF-V-LINEAR**: 依然 **BLOCKED**——Linear MCP が未接続（`USER_NOT_LOGGED_IN`）で照会不能。`EMR-136` へ移行済みだが対応先の確定は保留。
- **配置**: `maa01`/`sin14` いずれも日本非ローカルで APAC 制約は満たすが Japan 着地は保証しない。runbook の再抽選（`LABEL rollout` インクリメント再デプロイ）の実施は承認済み運用操作に委ねる。
- **新規候補**: `GET /api/v1/accountings?owner_id=` が owner detail 画面の最遅 API（1.9–2.7s、browser-pages 観測）。次期改善候補として記録する。
