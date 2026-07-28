# FE-refactor — FE12-02 active-only ledger

> 更新: 2026-07-28（要件責任者: 曽我）
> 業務目的: 未完了の臨床sentinel・RBAC安全境界を、決裁と実測の証跡が揃うまで追跡する。
> 本ファイルは使い捨てのactive ledger。恒久規約は `DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/ui-design-compliance.md`、`frontend/CLAUDE.md` を正本とする。
> **完了項目は本ファイルから削除する。** FE12 実測レーン（M-01-E / M-02 / M-03 / M-04 / R-1 / R-3 / R-4 / R-2測定）の分析・裁定・証跡の全文、および S1/S2レーン試行記録・fixture生存実測・実測レーン分割・ブラウザ経路確立の各節は、削除直前の版であるCommit `a36319d17` に保存する。さらに古いFE/R-F履歴はCommit `657c1a49cd2c37dc63f5af8e530258a36a12d81e` に保存する。
> 実装findingsは本ledgerの所掌外であり、`3-session-agent.html#ledger` を正本とする。

## 現在地（2026-07-28 時点・ここだけ読めば残件が分かる）

**FE12 実測レーンは実質終結した。** M-01-E / M-02 / M-03 / M-04 / R-4 が COMPLETE、R-1 / R-3 / R-2測定 も完了。**fixture は R-4 で全て撤去済み**（例外: hospitalization `1` が日次記録制約で残存）。臨床状態は実測前の値へ復旧済み（pet `1001005`=alive、`1001004`=low、`1001002`/`1000018`=無傷）。

**残件は5つである。**

| # | 残件 | 誰が | 着手前提 |
|---|---|---|---|
| 1 | **M-01-D 裁定** — 死亡ペットの編集/削除をブロックするか | **曽我** | 材料完備。下記「M-01-D 裁定材料」を読めば答えられる |
| 2 | **M-05 残1件** — danger高cueのruntime未証明（期限4区分の誤検出有無は実測済み・0件） | エージェント | **fixture再作成が前提**（R-4で撤去済み）。着手プラン: `3-session-agent.html#TASK-461` |
| 3 | **R-2 実行** — `backend/tygo.yaml` の15行削除＋`make codegen`差分0確認 | **USER専権**（`make codegen`） | 削除は `fdbe77ef0` で完了済み。残るのは USER の codegen 検証3コマンドのみ。着手プラン: `3-session-agent.html#TASK-462` |
| 4 | **line-reserve** — font実機確認 | QA/端末管理者 | 実機3台とQA環境の受け渡し |
| 5 | **BUG-455〜458 の修正実装** | エージェント | 起票済み。`3-session-agent.html#ledger` が正本。本ledgerの所掌外 |

**この実測が生んだ実装findings 4件は全て起票済みである**（`BUG-455` CRITICAL / `BUG-456` HIGH / `BUG-457` HIGH / `BUG-458` MEDIUM）。以降それらの追跡は `3-session-agent.html#ledger` を正本とする。

**未起票の findings が1件残る**（下記「要起票」節の DBOrTx inventory gate）。

### M-01-D 裁定材料（M-01-E の実測結果・これで答えられる）

`/owners?include_deceased=true` の行アクションメニューを全権限 persona で観測した結果、**死亡行と生存行で提示される操作に差分が無い**。

| 操作 | 死亡行（pet `1001005`） | 生存行（pet `1001002`） | disabled |
|---|---|---|---|
| 編集 | 提示あり | 提示あり | **なし** |
| レポート | 提示あり | 提示あり | **なし** |
| 削除 | 提示あり | 提示あり | **なし** |

`docs/spec/specification.md:21` は「死亡ペットに対する誤操作の物理的ブロック」を規定するが、**一覧の行アクションはこれを満たしていない。** 製品哲学により確認ダイアログは選択肢に入らない（ロック・Undo・物理ブロックのいずれか）。

**曽我が決めるのは2点だけ**: (a) 死亡ペットの**編集**をブロックするか。するなら誤記訂正の正規経路をどうするか（追記のみ許す／管理者権限でのみ許す／死亡解除→訂正→再登録を強制する）。(b) 死亡ペットの**削除**をブロックするか。

なお F16（死亡バッジのグレー）は裁定済みで、実装（badge のみ）が合格である。死亡登録/解除は `45b681866` 以降 fail-closed で確定済み。**一覧に死亡登録/解除の導線は存在しない**（`PetCareSection.tsx` = 飼主詳細にある）。

## Active scope and authority

- 追跡対象は M-05 残1件、M-01-D 裁定、line-reserve font実機確認、R-2 の USER codegen 検証とする。M-01-E・M-02・M-03・M-04・R-1・R-3・R-4 は裁定・実測済みで追跡を終了した。
- 色と臨床semanticは `docs/spec/design-system.md`、恒久route適合は `docs/spec/ui-design-compliance.md`、明示的なPO/USER裁定は `q&a.html` を正本とする。
- authorityから項目が消えたことや判断待ち件数が0であることだけでは完了とみなさない。明示的な決裁または実測証跡が無い項目は保持する。
- 本ledgerの更新は実装・runtime検証・製品決裁を代替しない。
- **並行化の軸は「データ干渉」ではなく「ブラウザ」である。** ボトルネックは CDP endpoint が `127.0.0.1:9222` の1つしかないことである（`.codex/config.toml` の `mcp_servers.chrome-devtools` も 9222 決め打ち）。1プロファイル＝1ログインセッションのため persona 切替ごとに直列化する。効果があるのは browser lane から非ブラウザ作業を剥がすことであり、ブラウザ同士を並べても効かない。
- **route を実測対象に指定するときは、その route の guard の `action` と persona の権限を突き合わせてから書くこと。** FE12 ではプロンプト側の欠陥が連続して実測を止めた（証跡保存先が MCP の workspace roots 外／許可actionと禁止actionの混同／view維持の観測対象に edit-gated route を指定／台帳が要求する操作が実装に存在しない／後続fixtureを破壊する操作の指定／削除APIが無いrouteへのDELETEを「レコード不在」と誤記／認証入手手段とブラウザ禁止の自己矛盾）。
- **証跡が1枚も取れていない段階で外部agentへ委譲しない。** 3走・証跡0の原因はプロンプト精度ではなく、環境未検証のまま委譲した構造にある。

## Authority drift

- `docs/spec/ui-design-compliance.md` にはC18件数ratchetの旧記述が残るが、現行auditはratchetを持たない。本ledgerは正本側を編集せず、別scopeの文書driftとして扱う。是正は`docs/spec/ui-design-compliance.md`を所有する別unitが行う。

## C6a 臨床安全レビュー

- M-05 で残る danger 高 cue の runtime 証跡を取得する。
- 静的レビューで閉じず、残件が実装認可された場合は既存sentinel fixtureのscoped component testを先にRED化してから修正する。

## Active execution rules

決裁・実測後に対応する場合も、次の安全境界を維持する。

1. **臨床sentinelは生成型から表示・操作境界まで欠落させない。** 死亡は明示的なpositive matchで遷移・mutation callbackを拒否し、危険「高」は非色cueを伴う警告として扱う。死亡statusと死亡日時が不整合なら再登録導線を出さない。
2. **権限はaction別の最新値をmutation直前に再検査する。** UIの非表示・disabled・route guardだけを最終防壁にしない。view/edit共用の唯一のdetail routeはread accessを維持し、mutation境界をfail-closedにする。commit直後にも発火し得るcallbackのpermission refは`useLayoutEffect`で同期する。
3. **臨床date-onlyはJSTの厳密過去で判定する。** `YYYY-MM-DD`契約をguardし、`todayJSTISO()`との文字列比較`<`を使う。現在時刻との`Date`比較で当日を期限超過にしない。

## 要実測項目

### M-05 Clinical sentinel responsive — 残1件

- 残件は **danger 高 pet `1001004` の非色 cue の runtime 証明**のみ。**着手プランは `3-session-agent.html#TASK-461` を正本とする**（copy-executable な fixture 再作成 → 観測 → cleanup スクリプトを含む）。
- Route: `/owners`（pet名検索あり）・`/vaccinations`（pet filter）・`/checkups`（`M05-` 検索）の3routeで足りる。7 route×4 viewport の全再走は不要。
- 実装根拠: `frontend/src/features/owners/components/OwnersListTable.tsx:200` が `pet.dangerLevel === "高"` で分岐し、`:205` に `aria-label={<ペット名>の危険理由を表示}`、`:213` に `aria-label={<ペット名>の危険理由}` を持つ。**権限ガードは無い。** 先行実測では絞り込みありで7件検出・絞り込み無しで0件だったため、原因は検索/ページングである可能性が高い。
- **fixture は R-4 で撤去済み。再作成が着手前提**であり、期限4区分は実行日D基準（past=D-1 / today=D / future=D+1 / empty=NULL）で作り直す。日跨ぎ禁止。
- Expected result: danger 高が非色cueとaccessible nameを保持する。

**2026-07-28 実測で確定済み（再測不要）**

- **期限4区分の誤検出は0件**。vaccination は `2026-07-27` のみ `（期限超過）`、today/future は日付のみ（`a11y-vaccinations.txt:61-86`）。checkup は `M05-past` のみ `期限切れ`、today/future は `期限間近`、empty は表示なし（`a11y-checkups.txt:65-97`）。**today/future を danger 扱いする誤検出は無い。**
- **death の文字cueは PASS**。owner detail で `チロ`・`死亡` を確認（`detail-owner-300588.txt:157-163`）。ただし7一覧routeには表示されない。
- **未評価3項目は detail で `未判定` を確認**（`detail-examination-1014562.txt:75-93`）。HIGH/LOW は基準値0行のため成立しない（既知・`BUG-449`）。
- **view access は7 routeとも維持**（`アクセス権限がありません` 0件）。
- **非 GET mutation は7ファイルとも0件。**
- **layout は 28件中 PASS 20 / clip 8**。clip 8件は `/` 800×1024・500×900、`/medical-records` 800×1024・500×900、`/hospitalization` 800×1024、`/examinations` 800×1024・500×900、`/checkups` 500×900。**`BUG-458` へ統合済みであり本ledgerでは追跡しない。**
- evidence: `tmp/fe12-m05-evidence/2026-07-28/`（a11y 7・PNG 28・network 7・`fixture-to-cue.md`・`layout-review.md`・`completion-report.md`）。

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

## 要起票（本ledgerの所掌外・未起票分）

- **DBOrTx inventory gate が赤** — `exam_reference_range_repository.go` の `ResolveByFieldIDs` / `FindAnimalSpeciesID` が `persistence.DBOrTx` 参加者として未登録（`docker compose exec backend go test ./internal/lintscan/ -run DBOrTx` で再現）。pet側2件は `0eecddb11` で清算済みだが、この2件は #249 U3 を実装したセッションの所掌。ゲートの要求どおり ambient-tx 参加を実証する test を添えて登録する必要がある。**起票先は `3-session-agent.html#ledger`。**

## 維持する裁定（再提案を防ぐため保持）

- **カルテ同日重複にDB unique制約を採らない（2026-07-27）** — 同一pet同日に手で複数カルテを作ることは正当な業務（別々の来院）であり、制約は正当な操作まで禁止する。塞ぐべきは自動生成経路が同じ1回の来院に対して二重に作ることだけであり、これは`5e5868549`のtry-advisory-lockで自動生成経路に限定して解決済み。
- **auto-createにclock seamを導入しない（2026-07-27）** — 重複チェック日は`reservation.StartTime`由来であり現在時刻を参照しない。clock seamを入れると過去/未来予約の検索日が実行日へ変わり挙動を壊す。予約日基準の現行contractが正である。
- **manual chunkの追加分割投資を行わない（2026-07-27）** — 実測522.71 kB（gzip 145.80 kB）で500 kB警告に該当するが、`operations-routes.tsx`のlazy境界により独立chunkとして正しく分割済みで、`/manual`を開いた利用者だけが取得する。build警告は汎用閾値であって業務上の問題の証拠ではなく、表示遅延の申告も無い。存在する問題の証拠なしに最適化するのはproduct-philosophy①違反である。警告閾値の引き上げによる黙らせも行わない（サイレント化は⑤の禁止事項）。再開条件=manual画面の表示時間に関する具体的な業務上の申告が出た場合のみ。
- **F16 の「グレーアウト」は badge のみで合格（2026-07-28・曽我）** — 現行実装（`OwnersListTable.tsx:236-241` → `getPetStatusColor` → `status-helpers.ts:176-180` が死亡へ `BADGE.grayHover`）を是とする。行全体のグレーアウトは求めない。根拠は「死亡ペットの情報自体は正常に読めるべきであり、行全体を落とすと可読性を下げる」。
- **R-3 は選択肢 (B) — 基準値不在のまま M-02 を実測する（2026-07-28・曽我）** — `exam_reference_ranges` の0行は解消を待たない。基準値マスタの恒久是正は `3-session-agent.html#BUG-449` を正本として別途進める。M-02 はこの方針で完走した。
- **Board View の cage 403 は defect ではない（2026-07-28・M-04で実測）** — 入院 Board は cage master を読むため `master-hospitalization` の view 権限を要求する。M-03 用の group `10` は `hospitalization` view しか持たないため 403 になった。**fixture の権限設計が Board の要件を満たしていなかっただけで実装の欠陥ではない。** 起票不要。
