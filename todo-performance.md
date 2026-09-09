# Performance 調査・改善 TODO

作成・調査日: 2026-09-09（JST）

調査 ID: **PERF-STG-LOGIN**

対象: STG `/login` の初回表示遅延。責任者・依頼者: 曽我 稔。

ソース照合: ローカル `main` `c0950fbdf`。フロントの該当処理は STG 配信済み `main-BnyQFmpH.js` とも照合。backend / Worker の配信リビジョンは未照合。

実行 SoT: Linear。本書はユーザー指定のローカル調査・実装準備記録。Linear の関連 issue / 状態は **UNKNOWN（今回未照会・未更新）**。入口は [todo.md](todo.md)。

## 結論と判定境界

**確定: セッション確認が終わるまでログイン画面を描画しないため、通信待ちが白画面の待ち時間になる。**

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

## ソースから確認した発生経路

1. [router.tsx](frontend/src/app/router.tsx) が `/login` を含むルートを AuthProvider で包む。
2. [AuthProvider.tsx](frontend/src/features/auth/components/AuthProvider.tsx) の53–60行で `/login` でも session restore を有効にし、94–108行で確認完了を待つ。239行は `if (!isInitialized) return null`。既存ログインの復元と遷移（BUG-031）を守るための処理であり、単純な認証確認の削除は不可。
3. [refresh-token.ts](frontend/src/features/auth/api/refresh-token.ts) の17–21行は、名前に反して refresh POST ではなく **GET `/v1/me`** を呼ぶ。401/403 は匿名状態に戻す。
4. [axios.ts](frontend/src/lib/axios.ts) の13–23行は `X-Requested-With` / `X-Request-ID` を付ける。別 origin の API なので、preflight cache がないと OPTIONS が先行する。`X-Requested-With` は CSRF 防御のためであり、高速化目的の無条件削除は不可。
5. 同 axios の timeout は60秒。GET の network error / 502–504 は最大2回、1秒・2秒の間隔で再試行する。`/login` の401は refresh を誘発しない。E1の1件内の22.5秒をリトライの1秒・2秒待ちで説明することはできない。別途、通信障害時に待ち時間が長くなる設計上の残課題はある。
6. [Worker](backend/worker/index.ts) の44行に `sleepAfter = "10m"`、280–282行は通常リクエストを Container に転送する。OPTIONS の edge 即時応答はなく、Container が休止中なら起動を待つ可能性がある。**設定・コード上の候補であって、E1で実際に休止していた証拠は未取得。**
7. [CORS middleware](backend/internal/middleware/cors.go) の34行は preflight max-age 86400、36–38行は OPTIONS に204を返す。ブラウザーによるキャッシュ上限・キャッシュ状態は別途考慮する。
8. [auth middleware](backend/internal/middleware/auth.go) の66–69行は token 不在ならDB照会前に401。有効 token の認可経路は異なるため、この短絡を全401に一般化しない。

## 調査・改善 TODO

業務目的: ログイン開始時の無反応な待ちと、待ちによる再読込・操作のやり直しを減らす。まず白画面という不要な待ち方を除き、認証・CSRF・医院分離は維持する。以下の改善は **未実装**。

| 優先 | 項目 | 状態 | 実施内容・完了条件 |
|---|---|---|---|
| P0 | 遅延区間の確定 | 部分完了 | E1の22.5秒が最終GET接続前であることは確認済み。次は下記手順で OPTIONS / 接続待ち / Container 起動の内訳を確定する |
| P0 | ログイン時の白画面を解消 | DONE (candidate) | PERF-STG-LOGIN-A: SessionPending + AuthProvider/router/hydrate wiring. Vitest RED→GREEN pending-Promise; protected children unmounted. Browser/E2E BLOCKED (no candidate fixture). B/C/D remain open. |
| P1 | セッション復元の待ち上限・障害表示 | TODO | 60秒×再試行の現状を踏まえ、起動時専用のtimeout/retry方針を定める。401と通信障害を区別し、再試行できる表示を設計。遅れて到着した復元結果と手動ログインが競合しないことを検証する |
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

## 改善プラン（2026-09-09 追記・未実装）

方針は **A: 白画面をなくす → B: 復元待ちを制御する → C: 計測で特定した通信経路を直す**。Cの原因確定待ちでA・Bを止めない。bundle削減は、その後の初回キャッシュなし計測で必要性を判断する。

以下の時間は設計上の暫定目標であり、測定済みの性能保証・SLOではない。既存の正常時FCP約0.8–0.9秒を基準に、同一端末・回線・キャッシュ条件で比較する。

### A. 非機密の画面枠を先に表示する（最初の実装単位）

認証APIの結果を待たずに、ブランド表示と「ログイン状態を確認しています」を描画する。保護された画面・医院情報・患者データのコンポーネントは、認証確定までマウントしない。

- `AuthProvider` の未初期化時の `null` を、認証コンテキストに依存しない軽量な待機表示に置き換える。ローディング中に単純に全 `children` を表示する変更はしない。
- `app-routes.tsx` のログイン用 `Suspense fallback={null}`、`router.tsx` のルートfallback、`root-hydrate-fallback.tsx` も確認し、JSチャンク待ちでも同じ非機密表示を出せるようにする。新しい大型ライブラリは追加しない。
- 既存のdesign tokens・ブランド部品を使い、`role="status"` 等で状態を通知する。自動フォーカス移動や連続読み上げは避ける。
- **完了条件:** `/me` を23秒保留するテストでも待機表示が先に出る。保護画面のマウントと業務API発火は0件。既存セッションの復元・ログイン済み遷移、パスワード復旧画面を維持する。
- **実測目標:** 正常な静的配信条件で非機密表示のFCPは1秒以内を目指す。ここで改善するのは白画面であり、フォーム操作可能までの時間やAPI所要時間の改善とは別に報告する。

対象候補: [AuthProvider.tsx](frontend/src/features/auth/components/AuthProvider.tsx)、[app-routes.tsx](frontend/src/app/routes/app-routes.tsx)、[router.tsx](frontend/src/app/router.tsx)、[root-hydrate-fallback.tsx](frontend/src/app/root-hydrate-fallback.tsx)。待機UIは依存方向に沿って配置し、認証feature外から内部ファイルを直接importしない。


#### Child unit progress — PERF-STG-LOGIN-A (2026-09-09 15:11 JST)

- Claim: `claim/PERF-STG-LOGIN-A` acquired in shared repo; parent `claim/PERF-STG-LOGIN` left intact.
- Candidate: `/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte-perf-login-a-20260909` on branch `perf/stg-login-pending-a-20260909` from `PERF_BASE=c0950fbdfe8c2267ea2ab33a41a906fee8f1f6b6`.
- Plan imported by exact-byte copy (SHA256 96648dc0a9ba218ece744ab0f3dcdaf214c50c14558ba94edcb1c682f5f692a9); B/C/D remain pending.
- Status: independent review PASS (code+security); draft PR next (browser/E2E BLOCKED).

#### Implementation evidence — PERF-STG-LOGIN-A (2026-09-09 15:22 JST)

- Changed paths: AuthProvider.tsx, SessionPending.tsx(+test), root-hydrate-fallback.tsx(+test), router.tsx, router.test.ts, app-routes.tsx, app-routes.test.tsx, use-auth-initial-session.test.tsx, todo-performance.md, todo.md
- RED: pending refreshToken Promise → Unable to find role=status (empty body)
- GREEN: verify-agent-task related+eslint+prettier PASS (completed_tests=45 on owned frontend set)
- Assumptions: Layout.tsx isLoading null left unchanged (out of allowlist; unreachable after AuthProvider gate). No FCP/API timing claim. E2E/browser BLOCKED.
- B/C/D: still pending


#### Frozen pending-UI contract (from investigation probes; awaiting workflow freeze join)

- pending_component_path: `frontend/src/components/shared/auth/SessionPending.tsx`
- props_contract: optional `message` (default 「ログイン状態を確認しています」); root `role="status"` + `aria-live="polite"`; hook-free; design tokens + lucide Stethoscope only; NO imports from `@/features/auth` or auth hooks/state
- mount_sites: AuthProvider `!isInitialized` gate; `RootHydrateFallback`; `router.tsx` root Suspense; `app-routes.tsx` login Suspense
- exact_test_paths: `use-auth-initial-session.test.tsx`, `use-auth-clinic-switch.test.tsx`, `router.test.ts`, `app-routes.test.tsx`, `root-hydrate-fallback.test.tsx`, `SessionPending.test.tsx`
- must_not_change: LoginForm*, axios, refresh-token/login/logout endpoints, Layout.tsx (out of allowlist; dead isLoading path after AuthProvider gate), B/C/D plan items, deps/lockfiles
- implementation_notes: Keep children unmounted while pending (do not wrap children). Reuse LoginFormBrandHeader visual pattern in shared without importing auth. E2E/browser BLOCKED (no candidate-serving fixture).


#### Acceptance Checklist (expanded before implementation)

- [x] Pending restore shows non-sensitive UI without protected children | Target: AuthProvider + SessionPending | Verify: scoped Vitest pending-Promise test RED→GREEN; protected mount and business fetch spies stay zero | PASS: named regression fails before fix and passes afterward
- [x] 200/401, recovery, clinic and StrictMode invariants retained | Target: use-auth-initial-session + use-auth-clinic-switch + affected route tests | Verify: exact candidate-mounted verify-agent-task.py paths command | PASS: nonzero test counts, zero failures
- [x] Lazy and hydration pending states remain visible and accessible | Target: root-hydrate-fallback, router, app-routes, SessionPending | Verify: follow-up cycle1 deferred pending→resolved tests (below) supersede extracted-fallback-only coverage | PASS: cycle1 Docker completed_tests=9 + neg control Unable to find role=status
- [x] Changes stay within write allowlist; preserve pre-existing/foreign changes | Target: tracked/staged/untracked paths | Verify: git diff --name-only, --cached --name-only, ls-files --others vs pre-edit baseline | PASS: only allowlisted owned changes; unknown ownership BLOCKED
- [x] Isolated candidate descends from main and includes preserved plan | Target: worktree + ledger | Verify: merge-base ancestor of PERF_BASE; plan SHA import; check-ignore/ls-files --stage | PASS: main base recorded; source SHA unchanged; B/C/D remain pending
- [x] Independent review + scoped quality gates completed | Target: frozen candidate diff | Verify: independent reviewer with file/line evidence + exact test/lint/format outputs | PASS: no unresolved CRITICAL/HIGH; review joined
- [x] Task branch published and one main PR created | Target: MinoruSoga/AnimalEkarte | Verify: gh pr view --json url,baseRefName,headRefName,headRefOid,isDraft,state vs HEAD | PASS: OPEN PR to main with exact owned head; draft/ready and CI reported truthfully
- [x] Workflow-style orchestration used and all launched work reconciled | Target: this session | Verify: Deliverables orchestration evidence | PASS: Workflow/subagent mode recorded; every agent ID/role/status/evidence/integration joined or cancelled



#### Evidence follow-up cycle 1 — PERF-STG-LOGIN-A (2026-09-09 16:02 JST)

FOLLOWUP_BASE=`8060f87d91cfeb9584b55257b8f547bd646db280`. Claim `claim/PERF-STG-LOGIN-A` retained by original receiver session `01a084bb-b11e-7673-81d5-98974872de68`. Parent claim intact. Source foreign backend permission-policy WIP preserved (unstaged FP now `9c24aa2c…`; original session-start FP was `e273478b…`).

##### Follow-up Acceptance Checklist

- [x] Actual root/login lazy and hydration pending-to-resolved boundaries covered | Tests: `appRoutes login Suspense pending-to-resolved…`, `root Suspense pending-to-resolved…`, hydrate deferred lazy tests in router/root-hydrate files | Verify: `python3 -B scripts/verify-agent-task.py --paths frontend/src/app/router.test.ts frontend/src/app/routes/app-routes.test.tsx frontend/src/app/root-hydrate-fallback.test.tsx --frontend-image sha256:532501622cd024ab786a32eb9798db1cd1a0e4d47cddb3dbd56ae107f95d9cb4 --frontend-dependency-volume ekarte-frontend-node-modules` → status PASS, completed_tests=9, eslint0, prettier0 (`/tmp/perf-login-a-followup-verify2.log`)
- [x] Missing historical evidence recovered or bounded | RED: `/tmp/perf-login-a-red-detail.log` + session `terminal/call-c50ddd98-…-76.log` — `Unable to find role="status"` empty body; promptSha256 `b35514951…` EXIT0; agents listed below | Fresh re-verify dated 2026-09-09 16:02 JST is current correctness only, not historical proof
- [x] Final independent review + narrow quality gates | Reviewer `01a084f6-d2f7-7ab3-a37b-44f5056129f4` initially FAIL on ledger overclaim (fixed this revision); security `01a084f6-d2f7-7ab3-a37b-4501269bc66e` PASS; gates PASS 9 tests
- [x] Changes within 4-path allowlist | Only router.test.ts, app-routes.test.tsx, root-hydrate-fallback.test.tsx, todo-performance.md; no production code
- [x] Ledger + PR393 reflect evidence without overclaim | PR393 draft OPEN main head `e81ed8eda67d0ef5af81379fba8a1eda42c88d55` | Browser/E2E BLOCKED; CI Frontend Build audit FAILURE is deps-out-of-scope; B/C/D pending
- [x] Workflow-style orchestration reconciled | Workflow `ae-perf-stg-login-a-evidence-followup` probes joined; spawn reviewers joined

##### Historical evidence recovery (joined)

- Original saved prompt SHA256: `b35514951a0f61b176d1a3120521c717dc1c793838d704ae734e761e18e475e0`
- Original RED (historical): `TestingLibraryElementError: Unable to find role="status"`; `<body><div /></body>`; test `shows non-sensitive pending UI…` at use-auth-initial-session.test.tsx:157; Start at 15:17:56
- Investigate wf `wf_01a084cc28a575e08806ba2cd29b2f7d`: probe-auth-render `01a084cc-28bf-…`, probe-tests-rules `01a084cc-28c3-…`, freeze-contract `01a084ce-a21b-…` — complete/joined
- Review wf `wf_01a084d405947760a2c63685ee1dffc7`: code-review `01a084d4-05b2-…`, security-review `01a084d4-05b4-…` — complete/joined; backup spawn reviewer/security also PASS
- Follow-up neg control (fresh 2026-09-09 16:02 JST, not historical TDD): offline Docker candidate RO + tmp null-patched hydrate → `Unable to find role="status"` / empty body (`NEGATIVE_CONTROL_SENSITIVE_TO_NULL_FALLBACK`)
- Prior gap: extracted-fallback-only route tests overclaimed; cycle1 replaces with deferred pending→resolved coverage


##### Workflow freeze join (late)

- Workflow `ae-perf-stg-login-a-evidence-followup` completed; freeze-followup-scope `01a084f1-c3b2-70e3-a271-f879486f1112` done (joined).
- Freeze asked login coverage via `createMemoryRouter`+`RouterProvider` over real login RouteObject (applied in follow-up repair).
- Raw `/tmp/perf-login-a-red*.log` may be absent now: **BLOCKED as durable artifact**; recovered content remains in session `terminal/call-c50ddd98-…-76.log` and prior `/tmp` capture used during this session.

### B. 起動時の認証確認だけに待ち上限を設ける（Aの次）

起動時のsession restoreを、通常業務API・権限再取得とは分けて制御する。初期案は **1試行8秒、起動時の自動リトライ0回**。8秒は今回の23秒待ちを無反応のまま継続させないためのUX上の初期値であり、cold startを8秒以内に短縮する保証ではない。

| 状態 | 表示・操作 | 認証状態の扱い |
|---|---|---|
| 確認中 | Aの表示。二重ログインを開始しない | 未確定。保護画面には入れない |
| 200で有効なセッション | 既存の認証済み遷移 | 検証済みユーザー・医院を反映 |
| 401 | ログインフォームを表示 | 未ログインとして扱う |
| 403 | 応答契約に沿ったアクセス制限・医院選択の案内 | 通信エラーや401へ一律変換しない。既存の医院選択回復経路を維持 |
| 8秒到達・通信失敗・5xx | 「ログイン状態を確認できませんでした」と再試行、別途ログイン画面への切替を提示 | 未認証と断定しない。Cookie消去・logout送信をしない |

- 起動専用APIは成功・未ログイン・制限・通信失敗を型で区別する。既存 `refreshToken()` は権限再取得にも使われるため、全呼出しの意味やaxiosの共通60秒設定をまとめて変更しない。
- `AbortController` によるキャンセルに加え、試行番号で古い結果を無視する。再試行・ログイン切替・logout・unmountで番号を進め、古い成功/401がuser・医院・React Queryキャッシュを上書きしないようにする。キャンセルが通信先の処理停止を保証するとは扱わない。
- 共通axios interceptorの自動リトライにも、起動時専用の型付きオプションで適用除外を通す。単にtimeoutを8秒に変えるだけでは、共通リトライが再度待機を作るため不十分。
- ログインフォームへの切替時は復元試行を失効させてから、明示的なログインを受け付ける。送信中は重複送信を防ぐ。期限到達や切替だけでセッション・選択医院を消去しない。
- **完了条件:** 8秒到達後にエラーと再試行操作が表示され、待機だけを継続しない。再試行はクリック1回につき1試行。旧復元の200/401/失敗が新しいログイン・医院・logout後の状態を覆さない。
- **実測目標:** warm・匿名時のフォーム操作可能時刻は2秒以内を目指す。障害時は復元開始から8秒で回復操作を提示する。JS読込時間は別に記録する。再試行や手動ログインも通信を伴うため、8秒でログイン成功するとは約束しない。

対象候補: authの起動専用API・[AuthProvider.tsx](frontend/src/features/auth/components/AuthProvider.tsx)・[LoginForm.tsx](frontend/src/features/auth/components/LoginForm.tsx)・[axios.ts](frontend/src/lib/axios.ts)・必要な認証型。共通interceptorを変更する場合は起動API以外の既存retry/CSRF/医院選択の回帰を必須とする。

### C. 通信経路を計測して、効果のある対策だけを入れる（A・Bと独立して調査）

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
| A: 非機密の待機表示 | 遅延Promiseで初期表示を検証、保護children非表示、auth / route回帰、キーボード・読み上げ | scopedテストと独立レビューに重大指摘なし |
| B: 復元の期限と競合制御 | fake timersで8秒境界、200/401/403/5xx、再試行、旧成功/旧401、logout、StrictMode、医院回復 | 状態上書き・二重通信・権限低下の回帰なし |
| C: 計測・通信経路修正 | OPTIONS/GET別の時刻、cold/warm比較。変更時はCORS・内部ルート・CSRF契約 | 原因と対策の因果を示す実測あり。環境変更・deployは別途承認 |
| D: bundle | キャッシュなし/あり、チャンク依存、操作可能時刻 | A–Cの残課題と切り分けた改善量あり |

実装時は対象worktreeに結び付いたDockerで、既存の `use-auth-initial-session.test.tsx`・`LoginForm.test.tsx` と変更箇所隣接テストを指定して `npx vitest run <対象パス>` を実行する。共有axios変更時は関連interceptorテストも追加する。全体build/testや環境起動を自動実行しない。authロジック変更はReact・認証安全性の独立レビューを受ける。

STG受入は「待機表示が出る」「フォームが操作できる」「認証が成功して業務画面へ進める」の3時刻を分ける。まず通常読込の少数比較から始め、単発値でp95/p99達成を宣言しない。継続測定の件数・時間は既存のSTG測定契約で定める。既存セッション、匿名、復旧画面、医院選択の回帰が出た変更単位は先へ進めず、原因修正または承認済みの前版への切戻しを選ぶ。

現在は**計画の追記まで**。本セッションが取得した `claim/PERF-STG-LOGIN` を同じ調査・計画作業として継続使用し、新たな実装着手・STG変更はしていない。

## 参考・検証境界

- [Resource Timing仕様](https://www.w3.org/TR/resource-timing/) の `requestStart` は final network-request start。最終GETの応答待ちとfetch全体を区別する。
- [Fetch仕様・HTTP fetch](https://fetch.spec.whatwg.org/#http-fetch): 必要なpreflightは実リクエストより先に処理する。
- [Chrome Network Timingの説明](https://developer.chrome.com/docs/devtools/network/reference/#timing-explanation): Queueing / Stalled / Proxy negotiation / Waitingを区別する。
- [Cloudflare Containers公式SDK](https://github.com/cloudflare/containers): fetchによる起動とsleepAfterの説明。
- 既存の測定手順: [STG-PERFORMANCE-CHECKLIST.md](docs/ops/testing/STG-PERFORMANCE-CHECKLIST.md)。health / clinics の低負荷API試験はログイン画面のFCP・preflight・cold startの代替証拠にしない。
- 今回は調査とMarkdownのみ。アプリのbuild/testは変更がないためSKIP。ローカル参照リンク・入口リンク・差分の空白チェックはPASS。コード修正・deploy・環境設定変更・ログイン・DB照会・migration・負荷試験・Linear更新は実施していない。
- 作業ファイルは `todo-performance.md`（新規）と `todo.md`（入口追記）のみ。開始時の作業ツリーはclean。未コミットで保全し、今回取得した `claim/PERF-STG-LOGIN` は保持する。次回編集前に本セッションの所有終了・引継ぎを確認し、AGENTS.mdのclaim規則に従う。
