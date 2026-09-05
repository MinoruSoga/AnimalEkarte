# STG 一覧パフォーマンス報告（2026-09-05）

> **目的**: 共有 STG の一覧画面が遅い理由を、フルリロード実測と温回 API 再測で切り分け、改善の順序を決める。
> **対象環境**: `https://stg.noah-karte.com` / `https://api.stg.noah-karte.com`
> **測定日**: 2026-09-05 JST
> **実装日**: 2026-09-05（本ファイル §0 のチケット）
> **承認済み SLO ではない**。比較用の提案目標は [PERFORMANCE_PROFILING.md](PERFORMANCE_PROFILING.md) の initial display 1.5 s。

## 0. 実装チケット

測定で決めた順（削除 → 認証フロア → COUNT 分離 → インフラ）。`total` の JSON 形は維持し、COUNT から汚染行用 `EXISTS` だけ外す。fail-closed の行フィルタは Find に残す。

| ID | 内容 | 状態 |
|:---|:---|:---|
| P0-1 | 起動・ログインの `/me` を 1 本にし `ME_QUERY_KEY` を hydrate | 検証済み |
| P0-2 | `/me` の 10s stale・focus 再取得・30s poll を止める（SESSION=5min） | 検証済み |
| P0-3 | `GET /masters/staffs` を raw 同一 queryKey + `select` で 1 fetch。`useStaffValidation` も同じ cache | 検証済み |
| P1-1 | 一般スタッフの `Resolve` は所属 ID の active 確認のみ。全件 `ListClinics` しない | 検証済み |
| P1-2 | `CurrentAccess` を gin context へ。`GetMe` は assignments を再取得しない | 検証済み |
| P1-3 | `Resolve` 結果を staffID キーで 2s キャッシュ。失敗時は消す | 検証済み |
| P1-4 | staff+account+assignments を 1 クエリで読む（本番 reader） | 検証済み |
| P2-1/2 | 会計・カルテ・検査・飼主の COUNT から汚染 `EXISTS` を外す。Find には残す | 検証済み |
| P2-3 | 会計一覧 Preload から Staff ネストを外す。Items/Splits は日次集計が使うので残す | 検証済み |
| P2-4 | 検査 `Preload("Items")` は `include_items=true` のときだけ | 検証済み |
| P2-staff | 担当者一覧の無意味な `COUNT` をやめる | 検証済み |
| P3-1 | `instance_type` 引き上げ | **HOLD**（課金。P0–P2 再測後に owner 判断） |
| P3-2 | `sleepAfter` 延長 / 常時 1 台 | **HOLD**（コスト方針 AC-5。owner 判断） |
| P3-3 | 許可 Origin に `Timing-Allow-Origin` | 検証済み |
| P3-4 | `GET /api/v1/health` を公開（既存 `/health` のエイリアス） | 検証済み |

権限グループ変更の反映は最大 2s（P1-3）+ `/me` stale 5min（P0-2）。パスワード変更は JWT epoch で無効になるが、同一プロセスの Resolve キャッシュが残っている最大 2s は古い epoch が残る。タブ間の権限同期が業務要件なら `refreshPermissions` を明示呼び出しする。

P2 の `total` は医院スコープの概算になり得る（汚染行を含む）。行そのものは Find 側の fail-closed `EXISTS` で返さない。ページャの最終ページ番号が 1 ページ分ずれる可能性は、正確 COUNT をクリティカルパスから外したトレードオフである。

受け入れ再測（§8）は共有 STG で USER が行う。k6 と `psql` 直叩きはしない。

PHI（飼主名・ペット名）・パスワード・トークンは書かない。ログインは migrate フェーズ3 の合成 `stg-staff-*@example.test`。

---

## 1. 結論

待ちの本体は HTML ではない。Vercel の HTML load は 0.1–0.5 s。遅いのは `api.stg.noah-karte.com`。

コンテナを温めた直後でも、認証付きの小さな JSON が **2.0–2.6 s**、ページング済み一覧が **2.6–4.3 s** で安定する。コールドスタートは加算項であり、主因ではない。

提案目標 1.5 s は、温回の単一 API ですら未達。フルリロードの画面表示 10–16 s は、その API 時間に **起動時 `/me` の直列二重呼び出し** と **1/4 vCPU コンテナ上のリクエスト直列化** が重なった結果である。

改善は「一覧 SQL を微調整する」より先に、次を削る。

1. 画面起動の `/me` を 1 回にする。
2. 認証ミドルウェアが毎リクエストで走らせている clinic 全件読みを、リクエスト経路から外すか短命キャッシュする。
3. 一覧の正確 `COUNT(*)` を初回描画のクリティカルパスから外す。
4. そのあとでコンテナサイズと `sleepAfter` を費用対効果で決める。

---

## 2. 測定条件

| 項目 | 内容 |
|:---|:---|
| 方法 | ブラウザ Resource Timing（フルリロード）と、ログイン済みタブからの `fetch`（`credentials: include` + `X-Requested-With` + `X-Clinic-ID`） |
| 温回 | 同一セッションで同一 API を連続 3 回。コンテナは直前まで会計画面を開いており、`sleepAfter = 10m` のアイドル停止後ではない |
| 温回の医院 | 敷島デモ（catalog `stg-staff-21000021`、`X-Clinic-ID: 3`、会計最終ページ 5924 ≈ 11.8 万件） |
| 冷回の医院 | 城東・敷島・Hako bu neco（12:30 JST フルリロード 1 試行） |
| 未実施 | k6（共有 STG は禁止）、pprof（未配線）、PlanetScale `EXPLAIN ANALYZE`（エージェントは `psql` 直叩き禁止）、Lighthouse performance |

Vercel の `/api/:path*` rewrite は HTML を返すことがある。API 時間は `https://api.stg.noah-karte.com/api/v1/...` を正とする。Resource Timing の `transferSize` / TTFB は CORS `Timing-Allow-Origin` 欠如で 0 になりがちなので、`duration` を使う。

---

## 3. コールドスタート vs クエリ

Worker は Container へプロキシするだけ（`backend/worker/index.ts`）。Container は `instance_type: basic`（1/4 vCPU / 1GiB）、`max_instances: 3`、`sleepAfter = "10m"`。

| 観測 | 意味 |
|:---|:---|
| 未認証 404（`/health` 等、ルートなし）が 0.59–0.64 s | Worker → Container の往復下限は約 0.6 s。Go の業務クエリなしでもこの時間は残る |
| 認証付き 925 B（`GET /v1/masters/animal-species`）が 2.0 s | 認証再検証が約 1.4 s 乗っている |
| 温回 3 連打で `/me` 2.8 → 2.6 → 2.6 s、会計 4.3 → 4.3 → 4.3 s | 2 発目以降も落ちない。クエリと認証コストが本体 |
| カルテ一覧・検査・飼主は 1 発目だけ 1–2 s 重い | DB プランキャッシュかバッファのごく短い温め。その後も 2.6–3.9 s |
| 12:30 のフルリロード（10–16 s）より、温コンテナでの会計フルリロード（約 7.7 s）が短い | アイドル後のコンテナ起動は加算される。ただし温めても 1.5 s には届かない |

**切り分け結果**: 主因はコンテナコールドスタートではない。主因は (A) 毎リクエストの認証再検証、(B) 一覧の `COUNT` + 相関 `EXISTS`、(C) フロントの `/me` ウォーターフォールと並列リクエストのキューイング。

---

## 4. 数値

### 4.1 温回 API（敷島、連続 3 回、status 200）

単位は ms。`wait` はレスポンスヘッダ到着まで。body は数十 ms 以下。

| API | 1 回目 | 2 回目 | 3 回目 | body |
|:---|---:|---:|---:|---:|
| `GET /v1/me` | 2801 | 2557 | 2558 | 3.5 KB |
| `GET /v1/masters/staffs` | 2280 | 2287 | 2280 | 4.1 KB |
| `GET /v1/masters/animal-species` | 2038 | — | — | 0.9 KB |
| `GET /v1/pets?page=1&limit=20` | 3493 | 2584 | 2585 | 11 KB |
| `GET /v1/examinations?page=1&limit=20` | 4896 | 3665 | 3670 | 11 KB |
| `GET /v1/medical-records?page=1&limit=20` | 6315 | 3940 | 3926 | 13 KB |
| `GET /v1/accountings?page=1&limit=20` | 4325 | 4291 | 4297 | 43 KB |

同一コンテナへ `GET /me` と `GET /masters/staffs` を 2 本並列すると、壁時計 4.6 s、一方の staffs が 4.6 s になった。1/4 vCPU 上で認証付きリクエストが直列化している。

### 4.2 温回フルリロード（会計一覧）

| 区間 | 時間 |
|:---|---:|
| HTML navigation | 0.24 s |
| 1 本目 `GET /me`（start 0.43 s） | 3.03 s |
| 2 本目 `GET /me` と `GET /accountings` が 1 本目の完了直後に並列開始 | 3.27 s / 4.28 s |
| 行とページャが出るまで（navigation start 基準） | 約 7.7 s |

画面は `refreshToken()` の `/me` が終わるまで描画しない。その後に一覧と二度目の `/me` が走る。

### 4.3 冷回フルリロード（12:30 JST、1 試行）

表示完了 = ナビ開始から件数または行が出るまで。

| 医院 | 受付 | 飼主 | カルテ | 会計 | 検査 |
|:---|---:|---:|---:|---:|---:|
| 城東 | 10.2 s | 11.8 s | 13.1 s | 13.8 s | 14.5 s |
| 敷島 | — | 12.4 s | — | 16.1 s | — |
| Hako | 14.8 s | 15.7 s | 12.5 s | 14.3 s | — |

城東（カルテ約 76 万）と敷島（飼主約 6 千）で飼主・会計の表示時間はほぼ同じ。ページサイズ 20 件でも、認証フロアと `COUNT` が支配的。

冷回の最遅単一 API（Resource Timing）:

| API | 観測 |
|:---|:---|
| `GET /examinations` | 9.4 s（城東） |
| `GET /accountings` | 5.1–7.6 s |
| `GET /medical-records` | 約 7.3 s |
| `GET /me` | 2.5–6.7 s。フルリロードで 2 本。フォーカスで追加 |
| `GET /pets` | 3.7–4.4 s |
| `GET /masters/staffs` | 2.3–5.6 s。カルテで二重あり |
| `POST /login` | 4.3 s。クリックから遷移まで約 5 s |

城東の温回 API は未再測。冷回の会計 7.6 s は敷島 5.1 s より遅いので、巨大 `COUNT` の上乗せは残っている。

---

## 5. なぜ遅いか

### 5.1 毎リクエストの認証再検証（フロア約 2 s）

`middleware.Auth` は JWT 検証のあと、**毎リクエスト** `CurrentAccessResolver.Resolve` を呼ぶ（`backend/internal/middleware/auth.go`）。Resolve は少なくとも次を直列実行する（`backend/internal/auth/current_access_service.go`）。

1. staff 1 行
2. account 1 行
3. staff の clinic assignments
4. **`ListClinics`（医院カタログ全件）**

一般スタッフでも 4 は毎回走る。fail-closed の所属再検証としては正しいが、全医院リストは所属チェックに不要なことが多い。PlanetScale 直結かつ Hyperdrive なしなので、往復が積み上がる。

未認証 404 が 0.6 s、認証付き 0.9 KB が 2.0 s。差がこの層。

### 5.2 `/me` が同じ仕事をもう一度する

`GetMe` は staff・account・assignments・`ListClinics` を再取得し、さらに `CalculateEffectivePermissions` で権限グループ JOIN を走る（`backend/internal/auth/http_session_me.go`、`BuildMeResponse`）。ミドルウェアが直前に読んだ結果を使わない。

3.5 KB の `/me` が 2.6 s なのはペイロードサイズではない。

### 5.3 フロントが起動時に `/me` を直列 2 回呼ぶ

証拠:

- `refreshToken()` が `GET /v1/me`（`frontend/src/features/auth/api/refresh-token.ts`）
- ユーザー設定後に `useGetMe(true)` が同じキーでもう一度取る（`AuthProvider.tsx`、`get-me.ts`）
- `QUERY_STALE_TIMES.SESSION` は **10 s**、`refetchOnWindowFocus: true`、`refetchInterval: 30 s`

権限の即時反映が目的（コメント FE-RC-082）だが、1 本 2.6 s のエンドポイントを 10 s で stale、30 s で再取得、タブフォーカスでも再取得している。温回会計タブでは `/me` が 30 s 間隔で 3 s 前後、Resource Timing に残っていた。

`refreshToken` の結果を `ME_QUERY_KEY` に hydrate していないので、起動の 2 本目はキャッシュヒットにならない。`login.ts` は「ログイン応答に user があるので `/me` は不要」と明記しているが、ログイン後も `AuthProvider` が `refreshToken` → `useGetMe` を走らせて打ち消す。`Layout` は `isLoading` 中に `null` を返すため、最初の `/me` が終わるまで保護ルート全体が止まる。

### 5.4 一覧は 20 件でも全件 `COUNT` + 相関 `EXISTS`

いずれも `limit=20` の前に、医院スコープ全件へ `Count(&total)` を走らせる。

| 一覧 | 実装 | COUNT を重くしている条件 |
|:---|:---|:---|
| 会計 | `billing/accounting_repository.go` `findBillingsWithFilters` | `EXISTS (pets …)` を全 billing に。その後 Owner/Pet/Payments/Staff/Items/Splits を Preload |
| カルテ | `medicalrecord/medical_record_repository_list.go` | `medicalRecordDetailRelationsScope` が owner/pet/doctor/entered_by の相関 `EXISTS` を COUNT に載せる |
| 検査 | `medicalrecord/examination_repository.go` | `examinationPatientRelationsScope` が exam_type / pet+owner / medical_record の入れ子 `EXISTS` を COUNT に載せる。`examinationReadPreloads` は **常に `Preload("Items")`**。handler は `include_items` が無いと items を JSON に出さない（`examination_handler.go`）ので、一覧は読んでも捨てる |
| 飼主ペット | `pet/repository.go` | 生存フィルタ付き `Count`。ソート用に owners を常時 LEFT JOIN |
| 担当者 | `staff/staff_repository.go` | `COUNT` のあと `page=1, limit=1000` 固定。handler は pagination クエリを無視（`staffListMaxLimit`） |

テナント汚染行を応答に出さないための fail-closed スコープであり、削除対象ではない。ただし **正確総数を初回ペイントの同じクエリに載せる必要は別問題**。ページャの「最終ページ 5924」は、その数字を業務が毎秒必要としているかは疑うべき要件である。

カルテ検索をかけたときだけ treatments / procedures / medicines への深い EXISTS が COUNT に乗る（`medical_record_repository_list_search.go`）。今回の無検索一覧実測には含まれない。飼主一覧の `pets` は router loader なので、ナビ完了まで一覧 API が始まらない。

### 5.5 コンテナが並列を直列化する

`basic` は 1/4 vCPU。画面は `/me`・一覧・マスタを同時に撃つ。実測で同一 staffs 2 本が 2.3 s と 4.6 s になった。フルリロード 10–16 s の大部分は「遅い SQL × 本数」ではなく「遅い認証付きリクエストが 1 本の細い CPU で並ぶ」こと。

`DB_MAX_OPEN_CONNS=10` はスロット枯渇対策。ここを上げても認証 4 往復と COUNT は残る。

### 5.6 ログイン 4.3 s

成功時の意図的フロア（失敗応答 750 ms）は成功パスには乗らない。bcrypt cost 12 とログイン応答の `/me` 相当組み立てが本体。画面遷移約 5 s のうち API が 4.3 s。フロントの HTML ではない。成功後にまた `/me` を撃つのは P0-1 で消す。

---

## 6. 改善方針（この順で）

製品原則は ①要件を疑う → ②削除 → ③簡素化 → ④短縮 → ⑤自動化。コンテナ増強は ④。存在すべきでない二重 `/me` と、初回表示を止めている正確 COUNT を残したままインスタンスを大きくしない。

臨床の fail-closed（他院行を返さない、無効 staff を通さない）は削らない。削るのは **同じ事実の再取得** と **画面に要らない正確総数の同期取得**。

### P0 — 削除（効果が大きく、安全境界を動かさない）

| ID | 変更 | 期待 | 根拠 |
|:---|:---|:---|:---|
| P0-1 | 起動時 `/me` を 1 本にする。`refreshToken()` の結果を `queryClient.setQueryData(ME_QUERY_KEY, user)` し、`useGetMe` はそのキャッシュを使う | フルリロードから約 2.6 s 消える。会計温回 7.7 s → 理論上 4.5 s 台 | 実測ウォーターフォール。権限マップの正本は 1 回で足りる |
| P0-2 | `/me` の `staleTime` を SESSION 10 s から、権限変更の業務要件に合わせて延長。`refetchOnWindowFocus` を既定 off。30 s ポーリングを止めるか 数分へ | 操作中の 3 s スパイクが消える | 10 s stale × 2.6 s 取得は、権限変更の検知より画面停止の方が大きい。パスワード変更は JWT epoch で既にセッションが落ちる |
| P0-3 | カルテ画面の `GET /masters/staffs` を 1 本に。`useGetStaffs` と `useStaffValidation` は raw `["masters","staffs"]` を共有する。`useGetMasterItems("staff")` は別キーなのでマージしない | 並列キューが 1 本減る | 冷回で二重を観測。原因は validation が `["masters","staff"]` を別に撃っていたこと |

P0-2 の「権限を何秒で他タブへ反映するか」は個人名のついた要件が必要。名前がないなら 10 s は疑う。JWT epoch とログアウトで足りるならポーリングは削除対象。

### P1 — 認証フロアを 2 s から落とす（fail-closed は維持）

| ID | 変更 | 期待 | 注意 |
|:---|:---|:---|:---|
| P1-1 | `Resolve` の `ListClinics` を、システム管理者以外は assignments の clinic_id だけに限定する。全件カタログは `/me` の医院切替用に 1 回 | 毎リクエスト 1 往復減 | 一般スタッフの所属判定に全医院は不要。admin だけ全件 |
| P1-2 | ミドルウェアが読んだ `CurrentAccess`（と必要なら account）を context に載せ、`GetMe` が再クエリしない | `/me` が認証フロアに近づく | 権限計算のクエリは残る。そこは別計測 |
| P1-3 | プロセス内の短命キャッシュ（staffID + account epoch、1–5 s） | 同一画面の並列 5 本が staff/account/clinics を 5 回叩かない | 無効化は epoch 変更と明示ログアウト。権限グループ変更の反映遅延を要件として書く。サイレントな stale 権限は不可 |
| P1-4 | 認証 4 クエリを 1 ラウンドトリップの SQL にまとめる | RTT 積み上げを消す | 読み取り専用。write owner は動かさない |

目標の目安: 認証付き小さな GET を **0.6 s ホップ + 数十 ms DB** に近づける。これが落ちないと、一覧 SQL をいくら削っても 1.5 s には届かない。

### P2 — 一覧の正確 COUNT を初回から外す

| ID | 変更 | 期待 |
|:---|:---|:---|
| P2-1 | 最初の応答は `items` + `hasMore`（limit+1 または seek）だけ。総数は遅延 `GET …/count` か、ページャを「次へ」だけにする | 会計 4.3 s のうち COUNT 分が消える。検査の入れ子 `EXISTS` が COUNT に乗っているのが最優先 |
| P2-2 | 汚染行除外の `EXISTS` は **ページの 20 件側だけ**に残し、COUNT には載せないか、`clinic_id` 単独の概算にする | fail-closed は行の実体に残る。総数だけが汚染行を数本含む概算なら、ページ番号の誤差として許容できるか要件確認 |
| P2-3 | 会計一覧 Preload から Staff ネストを外す。Items/Splits は日次集計が使うので残す | 43 KB / 20 行のうち Staff assignments を落とす。日次タブは Items/Splits 必須 |
| P2-4 | 検査一覧の `Preload("Items")` を `include_items=true` のときだけにする | 一覧が捨てている exam_results を読まなくなる。城東 9.4 s の候補 |

`EXISTS` 自体は他院データ混入を防ぐので削除しない。削除するのは「76 万行の正確カウントを、20 行を出すのと同じ同期リクエストでやる」こと。

インデックスと `EXPLAIN ANALYZE` は、P2-1 のあと approved read-only で DBA が行う。エージェントは共有 STG に `psql` しない。候補は `(clinic_id, scheduled_date DESC)` on billings、`(clinic_id, date DESC)` on medical_records / exams。COUNT を外す方が先。

### P3 — サイクルタイム（インフラ）

P0–P2 のあとで測り直してから決める。

| ID | 変更 | いつやるか |
|:---|:---|:---|
| P3-1 | `instance_type` を上げる | 認証フロアを落としても並列キューが残るとき。費用が増える。先にフロントの同時発射を減らす |
| P3-2 | `sleepAfter` を延ばす / 最低 1 インスタンス常時 | 昼休み後の 10–16 s を問題にするとき。コスト方針（AC-5）と衝突するので owner 判断 |
| P3-3 | レスポンスに `Timing-Allow-Origin` | RUM の TTFB をまともにする。速さそのものは変わらない |
| P3-4 | 公開 `GET /health` を Container まで通す | 0.6 s ホップの継続監視。今は 404 が 0.6 s |

k6 は共有 STG では走らせない。isolated UAT・rate window・stop condition・owner が揃ってから。

---

## 7. やらないこと

- 未使用画面の `memo` / `useMemo` 祭り。一覧の待ちは React レンダーではない。
- 正確 COUNT を残したまま「インデックスで 1.5 s」と約束すること。認証フロアが 2 s ある。
- fail-closed の clinic スコープを外して速く見せること。
- 共有 STG への k6、本番 pprof 公開、`REPAIR_ALLOW_DATA_LOSS`、自動 `make migrate`。
- 認証再検証を無期限キャッシュして、無効スタッフを通すこと。

---

## 8. 次に測ること（実装後の受け入れ）

実装チケットを切るときの合格ライン。今回の温回をベースラインにする。

| メトリクス | 今（敷島温回） | まず目指す |
|:---|:---|:---|
| `GET /v1/me` 連続 3 回の中央値 | 2.56 s | 認証フロア低下後に再測。P0 だけなら回数 2→1 |
| 起動フルリロード中の `/me` 本数 | 2 | 1 |
| `GET /v1/accountings?page=1&limit=20` 中央値 | 4.30 s | COUNT 分離後に再測 |
| 会計フルリロードで行が見えるまで | 約 7.7 s 温 / 16 s 冷 | P0 後に温回再測 |
| 未認証 404 または `/health` | 0.6 s | ホップ監視用。アプリ SLO にはしない |

城東の温回 API 3 連打は未実施。P2 の効果確認は城東（最大データ）で行う。

---

## 9. 関連

- 測定手段の制約: [PERFORMANCE_PROFILING.md](PERFORMANCE_PROFILING.md)
- STG 構成: [../infra/architecture.md](../infra/architecture.md)
- STG 運用: [../infra/staging/runbook.md](../infra/staging/runbook.md)
- Worker `sleepAfter`: `backend/worker/index.ts`
- 認証ミドルウェア: `backend/internal/middleware/auth.go`
- 起動時 `/me`: `frontend/src/features/auth/components/AuthProvider.tsx`
- 保護ルート待機: `frontend/src/components/shared/Layout/Layout.tsx`
