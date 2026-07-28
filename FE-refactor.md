# FE-refactor — FE12-02 active-only ledger

> 更新: 2026-07-28（要件責任者: 曽我）
> 業務目的: 未完了の臨床sentinel・RBAC安全境界を、決裁と実測の証跡が揃うまで追跡する。
> 本ファイルは使い捨てのactive ledger。恒久規約は `DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/ui-design-compliance.md`、`frontend/CLAUDE.md` を正本とする。
> **完了項目は本ファイルから削除する。** FE12 実測レーン（M-01-E / M-02 / M-03 / M-04 / R-1 / R-3 / R-4 / R-2測定）の分析・裁定・証跡の全文、および S1/S2レーン試行記録・fixture生存実測・実測レーン分割・ブラウザ経路確立の各節は、削除直前の版であるCommit `a36319d17` に保存する。さらに古いFE/R-F履歴はCommit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に保存する。
> 実装findingsは本ledgerの所掌外であり、`3-session-agent.html#ledger` を正本とする。

## 現在地（2026-07-28 時点・ここだけ読めば残件が分かる）

**FE12 実測レーンは実質終結した。** M-01-E / M-02 / M-03 / M-04 / R-4 が COMPLETE、R-1 / R-3 / R-2測定 も完了。**fixture は R-4 で全て撤去済み**（例外: hospitalization `1` が日次記録制約で残存）。臨床状態は実測前の値へ復旧済み（pet `1001005`=alive、`1001004`=low、`1001002`/`1000018`=無傷）。

**M-05 は `TASK-461`（2026-07-28）で COMPLETE。R-2 は同日 USER codegen 検証で COMPLETE。M-01-D は前提誤りのため裁定不要としてクローズ。残件は2つである。**

| # | 残件 | 誰が | 着手前提 |
|---|---|---|---|
| 1 | **line-reserve** — font実機確認 | QA/端末管理者 | 実機3台とQA環境の受け渡し。**本ledger内で唯一残る実作業** |
| 2 | **BUG-455〜458 の修正実装** | エージェント | 起票済み。`3-session-agent.html#ledger` が正本。**本ledgerの所掌外** |

**この実測が生んだ実装findings 4件は全て起票済みである**（`BUG-455` CRITICAL / `BUG-456` HIGH / `BUG-457` HIGH / `BUG-458` MEDIUM）。以降それらの追跡は `3-session-agent.html#ledger` を正本とする。

**未起票の findings が1件残る**（下記「要起票」節の DBOrTx inventory gate）。

**本ledgerは line-reserve の実機確認が完了した時点で退役できる。**

### M-01-D — 裁定不要でクローズ（2026-07-28・前提が誤りだった）

**M-01-E が報告した「死亡行に物理ブロックが無い」は観測対象の誤りであり、defect ではない。** 追加実装は不要である。

**誤りの内容**: `/owners` 一覧の行アクション（`編集` / `レポート` / `削除`）は**飼主レベルの操作**である。`OwnersListTable.tsx:56-59` の型定義がそれを示す。

```
onEdit:          (ownerId: string) => void
onDeleteRequest: (ownerId: string, ownerName: string) => void
onReport:        (ownerId: string, petId: string) => void
```

**いずれも `ownerId` を受け取る。ペットの編集でも削除でもない。** 一覧はペット行単位で描画されるため「死亡ペットの行に編集・削除が出ている」ように見えるが、その操作対象は飼主である。飼主は生存しており、他に生存ペットを飼っている可能性もある。**ペットの生死で飼主操作をブロックすれば正当な業務を壊す。**

**ペット単位の操作には既に二重の物理ブロックが実装済み**（飼主詳細 `frontend/src/features/owners/components/OwnerPetsSection.tsx`）。

- `:108` `{pet.status === "死亡" ? null : canCreate ? (` — 死亡なら要素自体をレンダリングしない
- `:162` `{pet.status === "死亡" ? null : canDelete ? (` — 削除も同様
- `:113` `:123` `:133` `:143` `:153` `:168` — callback 側でも `if (current.status === "死亡" || ...) return;` で positive match 拒否

これは本ledgerの Active execution rules 1「死亡は明示的な positive match で遷移・mutation callback を拒否する」がそのまま実装された形である。**`docs/spec/specification.md:21` の「死亡ペットに対する誤操作の物理的ブロック」は達成済みである。**

**教訓**: 実測の観測対象を決めるとき、UI の見た目（ペット行に出ている）ではなく **handler が受け取る ID（`ownerId` か `petId` か）で操作対象を確定させること。** M-01-E はこれを怠り、飼主操作をペット操作と誤認した。

なお F16（死亡バッジのグレー）は裁定済みで実装（badge のみ）が合格。死亡登録/解除は `45b681866` 以降 fail-closed。**一覧に死亡登録/解除の導線は存在しない**（`PetCareSection.tsx` = 飼主詳細にある）。

## Active scope and authority

- 追跡対象は **line-reserve font実機確認のみ**とする。M-01-E・M-02・M-03・M-04・M-05・M-01-D・R-1・R-2・R-3・R-4 は裁定・実測済みで追跡を終了した。
- **R-2 は 2026-07-28 に COMPLETE。** `backend/tygo.yaml` の15行削除（`fdbe77ef0`）に対し USER が `make codegen` を実行し、`git diff --exit-code -- frontend/src/types/generated/` と `-- frontend/line-reserve/src/types/models.ts` がともに差分なしで返った。生成物3本の mtime も更新されており、codegen は実際に走って同一出力を得ている。**pointer mapping 15行が寄与0であったことが最終確認された。**
- **M-05 は `TASK-461`（2026-07-28）で COMPLETE。** 最後まで未証明だった danger 高 cue を runtime で確定させた: `/owners?search=クロ` の a11y に `button "クロの危険理由を表示"` / `dialog "クロの危険理由"` / `StaticText "FE12-M01 fixture"` を検出（pet `1001004`）。**先行走で0件だったのは絞り込み無しでページに載らなかっただけであり、実装欠陥ではない。** 期限4区分も再確認され today/future の誤検出は0件。臨床状態は cleanup 後に pet `1001004`=`alive/low/danger_reason 空` へ復旧済み（実測確認）。
- **`TASK-461` が露呈させた計画側の欠陥1件**: exact-ID 契約（MR `1425547` / vaccination・checkup `1`-`4`）は **R-4 後に DB sequence が進んだため成立しない**（MR `1425548` が実在するため `1425547` は再取得できない）。executor は SQL による強制を避け、`M05-` label による identity-safe な同定へ切り替えて臨床目的を達成した。**今後 fixture 再作成を計画する際、連番 ID を完了条件にしてはならない。** label など内容ベースの同定を使う。
- 色と臨床semanticは `docs/spec/design-system.md`、恒久route適合は `docs/spec/ui-design-compliance.md`、明示的なPO/USER裁定は `q&a.html` を正本とする。
- authorityから項目が消えたことや判断待ち件数が0であることだけでは完了とみなさない。明示的な決裁または実測証跡が無い項目は保持する。
- 本ledgerの更新は実装・runtime検証・製品決裁を代替しない。
- **並行化の軸は「データ干渉」ではなく「ブラウザ」である。** ボトルネックは CDP endpoint が `127.0.0.1:9222` の1つしかないことである（`.codex/config.toml` の `mcp_servers.chrome-devtools` も 9222 決め打ち）。1プロファイル＝1ログインセッションのため persona 切替ごとに直列化する。効果があるのは browser lane から非ブラウザ作業を剥がすことであり、ブラウザ同士を並べても効かない。
- **route を実測対象に指定するときは、その route の guard の `action` と persona の権限を突き合わせてから書くこと。** FE12 ではプロンプト側の欠陥が連続して実測を止めた（証跡保存先が MCP の workspace roots 外／許可actionと禁止actionの混同／view維持の観測対象に edit-gated route を指定／台帳が要求する操作が実装に存在しない／後続fixtureを破壊する操作の指定／削除APIが無いrouteへのDELETEを「レコード不在」と誤記／認証入手手段とブラウザ禁止の自己矛盾）。
- **証跡が1枚も取れていない段階で外部agentへ委譲しない。** 3走・証跡0の原因はプロンプト精度ではなく、環境未検証のまま委譲した構造にある。

## Authority drift

- `docs/spec/ui-design-compliance.md` にはC18件数ratchetの旧記述が残るが、現行auditはratchetを持たない。本ledgerは正本側を編集せず、別scopeの文書driftとして扱う。是正は`docs/spec/ui-design-compliance.md`を所有する別unitが行う。

## C6a 臨床安全レビュー

- **runtime 実測は M-05（`TASK-461`）で完了した。** 臨床 sentinel の未確認項目は残っていない。以降の臨床安全の残件は `3-session-agent.html#ledger` の `BUG-455`〜`458` として追跡する。
- 静的レビューで閉じず、残件が実装認可された場合は既存sentinel fixtureのscoped component testを先にRED化してから修正する。

## Active execution rules

決裁・実測後に対応する場合も、次の安全境界を維持する。

1. **臨床sentinelは生成型から表示・操作境界まで欠落させない。** 死亡は明示的なpositive matchで遷移・mutation callbackを拒否し、危険「高」は非色cueを伴う警告として扱う。死亡statusと死亡日時が不整合なら再登録導線を出さない。
2. **権限はaction別の最新値をmutation直前に再検査する。** UIの非表示・disabled・route guardだけを最終防壁にしない。view/edit共用の唯一のdetail routeはread accessを維持し、mutation境界をfail-closedにする。commit直後にも発火し得るcallbackのpermission refは`useLayoutEffect`で同期する。
3. **臨床date-onlyはJSTの厳密過去で判定する。** `YYYY-MM-DD`契約をguardし、`todayJSTISO()`との文字列比較`<`を使う。現在時刻との`Date`比較で当日を期限超過にしない。

## 要実測項目

### line-reserve font実機確認

- Route: 顧客向け`/line-reserve/{clinicId}`の顧客情報→要望→確認→完了とマイ予約（clinicId抽出契約=`frontend/line-reserve/src/lib/liff-config.ts:6-14`）。
- Fixture seed source: `backend/migrations/seeds/003_demo/line_reservation_settings.csv`、`reservation_types.csv`、`line_customers.csv`、`pets.csv`を基に`docs/ops/testing/scenarios/V05-auth-line-forms.md` V05-6/V05-7、`S04-liff-reservation-journey.md`の試験用顧客を使う。
- 前提実査: 全named CSV、V05、S04は実在。旧ブロッカーだったwebfont宣言は`frontend/line-reserve/index.html:7-12`のGoogle Fonts stylesheet＋2 preconnectで解消済み。`frontend/line-reserve/src/index.css:17-23`はNoto Sans JPを先頭fontへ指定する。
- 解消判定: source前提は解消済みで、現時点から実機確認へ着手できる。残る入力はLINE連携済み試験用account、試験環境、3実機、remote inspectionであり、source修正待ちは無い。
- 入力受渡し順: QA環境管理者がQAチケットに非秘密の試験環境URL、clinic ID、予約確定を許可する試験範囲、credentialの既存安全チャネル上の取得手順を記録する。次に端末管理者がiPhone/Android/iPadの端末ID、OS/browser version、remote inspection可否を同チケットへ割り当てる。credentialそのものはledgerへ記録しない。
- 次の一手: 上記2担当の受渡し完了を開始条件として、実機QA担当が3実機のcold/warm/offlineを実行し、CSS/font file 200、computed font-family、Rendered Fonts、clip/FOIT、fallback操作可否、端末/OS/browser/versionを記録する。
- Persona: LINE連携済み顧客persona。業務に影響しない試験用アカウントを使い、送信/予約確定はrunbookの試験環境だけで行う。
- Viewports: iPhone Safari 390×844、Android Chrome 412×915、iPad Safari 768×1024（加えてdesktop 500×900を比較用）。
- Interaction steps: physical deviceでcold loadし、DevTools/remote inspectionのNetworkでGoogle Fonts CSSとfont fileが200であることを確認する。顧客情報から完了/マイ予約まで遷移し、各画面のcomputed `font-family`と実レンダーfontを確認する。offline/reload時のfallbackも確認する。
- Expected result: `frontend/line-reserve/index.html:7-12`からNoto Sans JPがloadされ、`frontend/line-reserve/src/index.css:17-23`の先頭fontとして全画面へ適用される。clip/FOITによる操作不能がなく、font失敗時もfallbackで操作可能。
- Required evidence artifacts: 3実機の画面別screenshot、remote Network HAR、computed font-familyとRendered Fonts capture、端末/OS/browser/version、cold/warm/offline各結果。

#### QAチケット文面（そのままコピーして起票する）

```
件名: line-reserve（LIFF予約）Noto Sans JP の実機フォント確認 — 3実機

## 目的
顧客向け予約画面で Noto Sans JP が実機に適用されることを確認する。
webfont 宣言は実装済みのため、確認するのは「実機で本当に当たっているか」のみ。

## 背景（source 側は確認済み・修正待ちは無い）
- 宣言: frontend/line-reserve/index.html:7-12（Google Fonts stylesheet + preconnect 2本）
- 適用: frontend/line-reserve/src/index.css:17-23（Noto Sans JP を先頭 font に指定）
- 対象 route: /line-reserve/{clinicId} の 顧客情報 → 要望 → 確認 → 完了、およびマイ予約
- clinicId の抽出契約: frontend/line-reserve/src/lib/liff-config.ts:6-14

## 依頼1: QA環境管理者
本チケットへ次を記入してください（credential そのものは書かないでください）。
- [ ] 試験環境の URL
- [ ] clinic ID
- [ ] 予約確定を許可する試験範囲（どこまで実際に送信してよいか）
- [ ] LINE連携済み試験用アカウントの credential を、既存の安全チャネルで受け取る手順

## 依頼2: 端末管理者
本チケットへ次を記入してください。
- [ ] iPhone: 端末ID / iOS version / Safari version / remote inspection 可否
- [ ] Android: 端末ID / OS version / Chrome version / remote inspection 可否
- [ ] iPad: 端末ID / iPadOS version / Safari version / remote inspection 可否

## 依頼3: 実機QA担当（上記2件が揃ってから着手）
3実機それぞれで cold / warm / offline の3条件を実施し、次を記録してください。
- [ ] Google Fonts の CSS と font file が HTTP 200（remote inspection の Network で確認）
- [ ] 各画面の computed font-family
- [ ] 実際にレンダリングされた font（DevTools の Rendered Fonts）
- [ ] clip / FOIT による操作不能が無いこと
- [ ] font 取得失敗時も fallback で操作可能なこと
- [ ] 端末 / OS / browser の version

対象 viewport: iPhone Safari 390×844、Android Chrome 412×915、iPad Safari 768×1024
（比較用に desktop 500×900 も1回）

## 期待結果
Noto Sans JP が全画面に適用される。clip / FOIT で操作不能にならない。
font 失敗時も fallback で操作できる。

## 提出物
- 3実機 × 画面別 screenshot
- remote Network HAR
- computed font-family と Rendered Fonts のキャプチャ
- 端末 / OS / browser / version の一覧
- cold / warm / offline それぞれの結果

## 注意
- credential を本チケットへ書かないこと。取得手順のみ記載する。
- 送信・予約確定は上記「試験範囲」で許可された試験環境でのみ行う。
- 参照 runbook: docs/ops/testing/scenarios/V05-auth-line-forms.md（V05-6 / V05-7）、
  docs/ops/testing/scenarios/S04-liff-reservation-journey.md
```

## 要起票（本ledgerの所掌外・未起票分）

- **DBOrTx inventory gate が赤** — `exam_reference_range_repository.go` の `ResolveByFieldIDs` / `FindAnimalSpeciesID` が `persistence.DBOrTx` 参加者として未登録（`docker compose exec backend go test ./internal/lintscan/ -run DBOrTx` で再現）。pet側2件は `0eecddb11` で清算済みだが、この2件は #249 U3 を実装したセッションの所掌。ゲートの要求どおり ambient-tx 参加を実証する test を添えて登録する必要がある。**起票先は `3-session-agent.html#ledger`。**

## 維持する裁定（再提案を防ぐため保持）

- **カルテ同日重複にDB unique制約を採らない（2026-07-27）** — 同一pet同日に手で複数カルテを作ることは正当な業務（別々の来院）であり、制約は正当な操作まで禁止する。塞ぐべきは自動生成経路が同じ1回の来院に対して二重に作ることだけであり、これは`5e5868549`のtry-advisory-lockで自動生成経路に限定して解決済み。
- **auto-createにclock seamを導入しない（2026-07-27）** — 重複チェック日は`reservation.StartTime`由来であり現在時刻を参照しない。clock seamを入れると過去/未来予約の検索日が実行日へ変わり挙動を壊す。予約日基準の現行contractが正である。
- **manual chunkの追加分割投資を行わない（2026-07-27）** — 実測522.71 kB（gzip 145.80 kB）で500 kB警告に該当するが、`operations-routes.tsx`のlazy境界により独立chunkとして正しく分割済みで、`/manual`を開いた利用者だけが取得する。build警告は汎用閾値であって業務上の問題の証拠ではなく、表示遅延の申告も無い。存在する問題の証拠なしに最適化するのはproduct-philosophy①違反である。警告閾値の引き上げによる黙らせも行わない（サイレント化は⑤の禁止事項）。再開条件=manual画面の表示時間に関する具体的な業務上の申告が出た場合のみ。
- **F16 の「グレーアウト」は badge のみで合格（2026-07-28・曽我）** — 現行実装（`OwnersListTable.tsx:236-241` → `getPetStatusColor` → `status-helpers.ts:176-180` が死亡へ `BADGE.grayHover`）を是とする。行全体のグレーアウトは求めない。根拠は「死亡ペットの情報自体は正常に読めるべきであり、行全体を落とすと可読性を下げる」。
- **R-3 は選択肢 (B) — 基準値不在のまま M-02 を実測する（2026-07-28・曽我）** — `exam_reference_ranges` の0行は解消を待たない。基準値マスタの恒久是正は `3-session-agent.html#BUG-449` を正本として別途進める。M-02 はこの方針で完走した。
- **Board View の cage 403 は defect ではない（2026-07-28・M-04で実測）** — 入院 Board は cage master を読むため `master-hospitalization` の view 権限を要求する。M-03 用の group `10` は `hospitalization` view しか持たないため 403 になった。**fixture の権限設計が Board の要件を満たしていなかっただけで実装の欠陥ではない。** 起票不要。
