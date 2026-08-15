# Claude Fable — PO 裁定回答（2026-08-14 · UAT 後 人間レーン）

集計: **RATIFY 21 · TIGHTEN 4 · OVERTURN 0**（§D 全 25 行 = §3.1×7 + §3.2×5 + §3.3×11 + §3.4×2）。

### A. 読んだもの / 読めなかったもの

読んだもの

- 依頼書 `reports/fable-po-confirm-request-2026-08-14-uat-human.md` 全文
- `todo-po.md` 全文 · `todo.md` 全文（§1〜§5）
- `reports/uat-2026-08-14/FINAL.md` + **`results.json` 全 1387 件を再集計**（PASS 1352 · PARTIAL 26 · BLOCKED 7 · N/A 2 — FINAL と一致。PARTIAL 26 の内訳を全件分類: cleanup 残置 17 · recheck 済ハーネス起因 4 · 実質未検証 5）
- `docs/product-philosophy.md` 全文 · `docs/work/decisions/fable-po-recommendation.md` 全文（自分の採択済み pack）
- 前回回答 `reports/fable-po-confirm-answer-2026-08-14.md` 全文 · Sol r2 `reports/gpt-5.6sol-po-qa-answer-2026-08-14-r2.md` 全文
- `docs/delivery/DELIVERY_PACKAGE.md`（前提・§1・U1〜U13 表・§4〜§5。U1〜U12 全空欄・U13=#256 所有 DEC-62 を確認）
- `docs/ops/infra/staging/runbook.md` 全文 · `reports/todo-walk-2026-08-14/`（SECTION5-STATUS · staging-preflight-status · github-issues-walk）
- local git read-only 実測（fetch なし）: HEAD=main `1386e1db0` · `origin/staging...main` = **4 / 1346** · **`origin/main` = `697d5c597`**（todo.md 記載 tip と一致 — 本 worktree は 1 commit 遅れ）

**実測による新事実（裁定に反映）:** `697d5c597`（fix: align ERD and tenant boundaries）は docs 修正ではなく **`backend/migrations/001_init.sql` を +36 行編集**している（+締め調整系 migration test 追加）。今回のフル UAT は localhost（本 worktree = `1386e1db0`）で実行されており、**UAT 証跡はこの commit を含まない**。pull 時は POST-PULL（`make migrate`）該当、かつ適用済み 001 の編集のため local は checksum mismatch → OPS-2 承認済み fresh の可能性が高い。

読めなかった / 読まなかったもの

- GitHub live state（操作禁止 — #299 draft・Open 16 件は台帳と walk 報告を採用）
- STG / PROD runtime・DB・secret・外部サービス（接続禁止）
- `reports/uat-2026-08-14-sidebar-masters/FINAL.md`・`-v04/FINAL.md`（存在確認のみ。集計は todo.md §3 記載を採用）
- `docs/work/residual-closeout-ledger.md`（判定に不要）· シナリオ md（未読・未編集 — 依頼どおり）

### B. 再審

再審なし。DEC-40〜68 と Fable pack（2026-08-06 採択）を維持。2026-08-14 local フル UAT（FAIL 0 · PASS 1352）は製品バグレーンを空にしたが、既裁定の HOLD / DEFER / gate 境界はすべて環境・人間・外部起因であり、1 件も覆さない。

### C. Executive

**今日 Yes すべきトップ 3（前回から維持・差替えなし）**

1. **PO-11 / #201** — 臨床 bundle 空欄。臨床応答が全 residual の最長リードタイムで、TASK-033・#261 の共通解除条件。
2. **staging preflight / #299** — 残る人間セルは checksum/ownership・backup owner・CI green の 3 つだけ。ここが埋まれば実 LINE（UAT-H1/H2）・OPS-3/4/5/14・#252 の全 DO_NEXT が解錠される。
3. **#256 U13** — 残る判断セルは U13 だけ。完了/未完の明示は今日 5 分でできる。

**絶対に今 Yes すべきでないトップ 3**

1. **#254・#253 の完了扱い close、および「local FAIL 0 = 納品完了」と読める一切の対外表明** — フル UAT green が close 誘惑を最大化した今こそ最大の禁止。
2. **TASK-021 C/D・LINE-R05 の DROP / presence guard 除去**（inventory ゼロ確認前）— fail-closed guard の除去と不可逆 DROP。
3. **TASK-033 骨格先行・#257 新 window 設定** — DEC-48 一体所有と gate green 前日付禁止に正面から反する。

**前回 Fable 回答との差分:** トップ 3 は維持。TIGHTEN 4 点が新規 — ① UAT-H5/H6 を #254 close 前提へ昇格 ② #254 close 条件の明文化（residual disposition の具体化 + **UAT 証跡 build SHA `1386e1db0` の明記** — 以後 main は 001_init.sql 編集を含む `697d5c597` へ前進済み） ③ PARTIAL 26 のうち実質未検証 5 件の disposition 義務化（新行 UAT-H7 提案） ④ todo.md 対応表の「PO 確認待ち: なし」drift 修正。

**今回 UAT が変えたこと（2 行以内）:** 製品バグ台帳を Open 0 のまま 1387 ステップで確証し、残 blocker が 100% 人間・環境・外部レーンであることを確定した。人だけが閉じられる新規行（H5/H6 + 実質 PARTIAL 5 件）を特定した。

**変えなかったこと（2 行以内）:** #254/#256 の close 条件・staging → 実 LINE の順序・全 HOLD/DEFER 境界・DEC-40〜68。local FAIL 0 は依然どの close の単独証拠でもない。

### D. 全件確認表

#### §3.1 UAT 人間レーン

```text
ID: UAT-H1（S04 実 LINE プッシュ）
前回/Sol/todo: todo-po §1 open · §4.2 実LINE=DO_NEXT（STG health 後）· E-8
Fable Verdict: DO_NEXT
関係: RATIFY
依存: #299 merge-commit → STG deploy → 二面 health green + 承認済み実端末/test アカウント
POの答え（1文）: mock は代替にならず、current main の STG health 後にのみ実通知 1 往復を観測する。
次の人: UAT 責任者（+各医院 LINE 管理者の実端末）
次の一手: E-8 実施（予約確定・キャンセル各 1 回）
完了条件: 実端末で通知 1 往復以上・時刻/予約番号メモ + E-8 一行
#254 close 必要条件?: Y
────────────────────────────────────────
ID: UAT-H2（S12 実 LINE / LIFF token）
前回/Sol/todo: 同上（H1 と同一 window · E-8）
Fable Verdict: DO_NEXT
関係: RATIFY
依存: H1 と同一（STG health）+ 実 token が secret 経路で投入済み
POの答え（1文）: 実 token で起動〜主要画面到達までを H1 と同一 window で観測し、token 値は記録しない。
次の人: UAT 責任者
次の一手: E-8 の token_health / liff_link 項
完了条件: 起動〜主要画面到達（失敗時はスクショと HTTP のみ残す）
#254 close 必要条件?: Y
────────────────────────────────────────
ID: UAT-H3（S06 audit_logs DB 参照）
前回/Sol/todo: todo-po §1 open（local または STG）· E-6 db_audit=PASS 要件
Fable Verdict: DO_NEXT
関係: RATIFY
依存: なし — local 現行 build の DB へ read-only 接続のみ（STG を待たない）
POの答え（1文）: 監査行は build の挙動であり local で今日確認してよい — 書込は禁止のまま。
次の人: QA / DB 運用責任者
次の一手: カルテ確定・追記 1 件の record_id に対し audit 行の存在・action・actor を確認
完了条件: 結果 enum + opaque ref 一行（record 内容・実 ID は台帳に書かない）
#254 close 必要条件?: Y
────────────────────────────────────────
ID: UAT-H4（S09 締め fixture 属性）
前回/Sol/todo: todo-po §1 open
Fable Verdict: DO_NEXT
関係: RATIFY
依存: 会計権限アカウント + AM/PM/EMG・越日の fixture 属性の人手準備（local 可）
POの答え（1文）: 未締めの日・区分でプレビュー〜締めを 1 回通すか、通せない区分は意図的スキップ理由を記録して disposition とする — 実施は 697d5c597 pull + migrate 後の build が望ましい（001 編集が締め調整系に隣接）。
次の人: 会計権限を持つ USER / QA
次の一手: fixture 準備 → 締め UI 境界を通す
完了条件: 締め 1 回の記録、または区分ごとの明示スキップ理由
#254 close 必要条件?: Y（明示スキップ理由でも可）
────────────────────────────────────────
ID: UAT-H5（V04 シフトテンプレ SidePanel）
前回/Sol/todo: 今回 UAT で新設（CDP セレクタ未検出 → BLOCKED）
Fable Verdict: DO_NEXT
関係: TIGHTEN
依存: なし（local admin で今日可・所要数分）
POの答え（1文）: 自動化が到達できなかった受入対象フォームであり、#254 close 前提（完了または理由付き disposition）へ昇格する。
次の人: USER（admin）
次の一手: /settings シフトテンプレで新規 1 件保存 → 再オープンで永続確認
完了条件: 保存・永続 OK（NG なら todo.md §2 へ BUG 起票）
#254 close 必要条件?: Y
────────────────────────────────────────
ID: UAT-H6（S13 2 医院・2 飼主 fixture）
前回/Sol/todo: 今回 UAT で新設（S13 owner_link PARTIAL）
Fable Verdict: DO_NEXT
関係: TIGHTEN
依存: 2 医院所属 + identity-links 権限 + 紐付け可能な飼主 2 件の fixture（local）
POの答え（1文）: S13 の中核ループが未走行のままでは受入証拠に穴が残るため #254 close 前提へ昇格する（S13 owner_link PARTIAL は本行が吸収）。
次の人: USER（2 医院所属 admin）
次の一手: link → history → unlink → relink を scenarios/S13 どおり実施
完了条件: 全ループ成功 + cross-clinic 誤紐付けなし（NG なら BUG 起票）
#254 close 必要条件?: Y
────────────────────────────────────────
ID: UAT-S1（S08 部分入金不可）
前回/Sol/todo: 受容済み（BUG 起票しない）
Fable Verdict: ACCEPT
関係: RATIFY
依存: なし（受容記録済み）
POの答え（1文）: 仕様受容を維持し BUG にしない — 仕様変更は責任者名 + 業務目的（①）が提示された時のみ別チケットで開き、今は開かない。
次の人: —（PO 記録済み）
次の一手: なし
完了条件: 受容記録（todo-po §1 済）
#254 close 必要条件?: Y（既充足・追加作業なし）
```

#### §3.2 UAT 関連 PO 裁定

```text
ID: #254 close
前回/Sol/todo: HOLD（E-6: local FAIL 0 は閉証拠でない）
Fable Verdict: HOLD
関係: TIGHTEN
POの答え（1文）: H1〜H4 必須は維持し緩和しない — 加えて H5/H6/H7 と実質 PARTIAL の完了または理由付き disposition、UAT 証跡 build SHA（1386e1db0）の明記、別 USER final sign-off を close 条件に明文化する（§E）。
次の人: QA 責任者 → 別 USER 承認
次の一手: staging レーン（H1/H2）と local レーン（H3〜H7）を並行消化 → E-6 一行
完了条件 / 空欄に残すセル: E-6 各結果・residual_disposition・final_signoff・build SHA・opaque ref
────────────────────────────────────────
ID: #256 close
前回/Sol/todo: CLOSE_RECOMMEND（U13 + 発効日 + opaque ref + 別承認後のみ）
Fable Verdict: CLOSE_RECOMMEND
関係: RATIFY
POの答え（1文）: U13 未完のまま close する経路は引き続き禁止し、local UAT 結果は #256 の条件に一切影響しない（U13 は #258 に含めない = DEC-62 も維持）。
次の人: 納品責任者 → PO
次の一手: U13 完了/未完の明示 → 完了時のみ E-5
完了条件 / 空欄に残すセル: U13_status・発効日・close 承認・opaque ref
────────────────────────────────────────
ID: staging ← main（#299）
前回/Sol/todo: DO_NEXT（preflight 全 green 後 merge-commit のみ・draft #299 作成済み）
Fable Verdict: DO_NEXT
関係: RATIFY
POの答え（1文）: green 前 merge 禁止を維持する — checksum セルは形式ではなく、STG DB は 001 統合前の適用履歴を持ち直近 697d5c597 でも 001_init.sql が再編集されているため、mismatch disposition（STG seed runbook 準拠・reset なし）と PlanetScale 109 オブジェクト所有権の確認が merge 可否そのものだ。
次の人: リリース責任者 + STG DB 運用
次の一手: checksum/ownership・backup owner 記入 → PR CI 全 green → draft 解除 → merge-commit
完了条件 / 空欄に残すセル: checksum・ownership・backup/rollback owner・CI green・role・時刻・opaque ref
────────────────────────────────────────
ID: 実 LINE UAT · OPS-4/5
前回/Sol/todo: DO_NEXT（current main の STG health 後）
Fable Verdict: DO_NEXT
関係: RATIFY
POの答え（1文）: 順序は「#299 merge → STG deploy → 二面 health → 実 LINE（H1/H2）」で固定する — 1346 commit 遅れの旧 STG で先行実施しても current main の受入証拠にならない。
次の人: リリース → UAT 責任者
次の一手: staging merge 後に E-8
完了条件 / 空欄に残すセル: 各結果・opaque ref
────────────────────────────────────────
ID: PO-06 記録
前回/Sol/todo: done（close は #254）
Fable Verdict: ACCEPT（done）
関係: RATIFY
POの答え（1文）: 記録は §4.1 に完了しており追加判断は不要 — close は #254 レーンの別裁定のまま。
次の人: —
次の一手: なし
完了条件 / 空欄に残すセル: なし
```

#### §3.3 人が埋める空欄（優先順位のみ・値は発明しない）

```text
ID: #1 PO-11 / #201
前回/Sol/todo: DO_NOW
Fable Verdict: DO_NOW
関係: RATIFY
POの答え（1文）: 臨床応答が全 residual の最長リードタイムで TASK-033・#261 の共通解除条件 — Issue 依頼は 8/14 投稿済みのため、今日は人間チャネルで催促する。
次の人: 臨床責任者
次の一手: E-1 記入の催促 1 通
完了条件 / 空欄に残すセル: 対象・policy・単位・出典・approver・発効日・opaque ref
────────────────────────────────────────
ID: #2 staging preflight / #299
前回/Sol/todo: DO_NOW
Fable Verdict: DO_NOW
関係: RATIFY
POの答え（1文）: 残る人間セル 3 つ（checksum/ownership disposition・backup owner・CI green）を埋めれば merge 判断に到達し、後続 DO_NEXT が一斉に解錠される。
次の人: リリース責任者
次の一手: 残セル記入 → green 確認
完了条件 / 空欄に残すセル: checksum・ownership・backup/rollback owner・CI green・role・時刻・opaque ref
────────────────────────────────────────
ID: #3 #256 U13
前回/Sol/todo: DO_NOW
Fable Verdict: DO_NOW
関係: RATIFY
POの答え（1文）: 残る判断セルは U13 だけで、完了/未完の明示は今日 5 分でできる（未完なら「open 維持」を明示して完了）。
次の人: 納品責任者
次の一手: U13 状態確定 → 完了時のみ E-5
完了条件 / 空欄に残すセル: U13_status・発効日・close 承認・opaque ref
────────────────────────────────────────
ID: #4 PO-12 / #249
前回/Sol/todo: DO_NEXT
Fable Verdict: DO_NEXT
関係: RATIFY
POの答え（1文）: 依頼済み — 未承認 range は「未判定」を維持し、一般値・別測定系の流用も値の発明もしない。
次の人: 臨床責任者
次の一手: E-2 記入待ち
完了条件 / 空欄に残すセル: range・vaccine 全欄
────────────────────────────────────────
ID: #5 PO-13 / #211
前回/Sol/todo: DO_NEXT
Fable Verdict: DO_NEXT
関係: RATIFY
POの答え（1文）: clinical 行 → OPS 行の順で、両行完成まで実 row 投入は local 含め禁止のまま。
次の人: 臨床責任者 → OPS
次の一手: E-3 記入待ち
完了条件 / 空欄に残すセル: clinical・OPS 全欄
────────────────────────────────────────
ID: #6 #258 DELIVERY
前回/Sol/todo: DO_NEXT（前回 Fable が E-11 を REVISE）
Fable Verdict: DO_NEXT
関係: RATIFY
POの答え（1文）: REVISE 版 E-11（前回 Fable §E）を契約チャネルで送付する — GitHub 未投稿のままでよく、U9/U12 は開発側後埋め・U13 は含めない（DEC-62）。
次の人: 契約責任者
次の一手: REVISE 版 E-11 送付
完了条件 / 空欄に残すセル: A/B・各 U・承認・発効日・opaque ref
────────────────────────────────────────
ID: #7 PO-18 / OPS-1
前回/Sol/todo: DO_NEXT
Fable Verdict: DO_NEXT
関係: RATIFY
POの答え（1文）: staging・臨床と独立に USER が着手でき、4 系統 rotation 完了まで #89/#97 は open 維持。
次の人: セキュリティ責任者
次の一手: E-12 実施
完了条件 / 空欄に残すセル: enum・window・owner role・opaque ref
────────────────────────────────────────
ID: #8 PO-10 STG/PROD presence
前回/Sol/todo: DO_NEXT（local は §4.1 済）
Fable Verdict: DO_NEXT
関係: RATIFY
POの答え（1文）: 承認 read-only window（自然には STG health 後）で presence 件数のみ取得し、ゼロ確認前の DROP / guard 除去禁止は不変。
次の人: DB 運用責任者
次の一手: E-14 実施
完了条件 / 空欄に残すセル: env・件数・operator role・opaque ref
────────────────────────────────────────
ID: #9 #250 producer
前回/Sol/todo: HOLD（催促済み）
Fable Verdict: WAIT_EXTERNAL
関係: RATIFY（HOLD と同義 — 外部起因を明示しただけ）
POの答え（1文）: 8/14 催促済みで producer の complete bundle 回答待ち — 本日の再催促はしない。
次の人: 移行元責任者
次の一手: E-9 回答待ち
完了条件 / 空欄に残すセル: bundle・complete 判定・authority
────────────────────────────────────────
ID: #10 #259 enable
前回/Sol/todo: HOLD（催促済み・gate OFF）
Fable Verdict: WAIT_EXTERNAL
関係: RATIFY（同上）
POの答え（1文）: 先方 enable 回答待ち — 両 gate default-off と実 2xx のみ成功の原則を維持する。
次の人: 外部連携責任者
次の一手: E-10 回答待ち
完了条件 / 空欄に残すセル: enable・gate・rollback ref
────────────────────────────────────────
ID: #11 PO-008 顧客集計指標
前回/Sol/todo: DO_NEXT（現行継続を推奨）
Fable Verdict: DO_NEXT
関係: RATIFY
POの答え（1文）: §5.1 の 6 行は承認または修正の 1 通で閉まる軽い判断であり、クライアント返答便に載せる（現行継続を推奨・CSV default 追加なし・指標統一なしを維持）。
次の人: クライアント仕様責任者
次の一手: §5.1 承認/修正依頼
完了条件 / 空欄に残すセル: decision・修正値・承認者・発効日
```

#### §3.4 予約・その他

```text
ID: F-021-X / 2026-11-07
前回/Sol/todo: 前回 TIGHTEN で明示（inventory_start=2026-08-09 + 90 日）· todo-po §4 予約済み
Fable Verdict: DEFER（予約維持）
関係: RATIFY
POの答え（1文）: 日付 2026-11-07 と「無応答なら ACCEPT_RESIDUAL_RISK を自動適用ではなく再裁定に上げる」条件をともに維持する。
次の人: PO
次の一手: 期日まで inventory 回答を待ち、無応答なら再裁定カードを todo-po へ
完了条件 / 空欄に残すセル: inventory 結果 enum・opaque ref
────────────────────────────────────────
ID: PARTIAL 26
前回/Sol/todo: 初裁定（前回 UAT は PARTIAL 4 → 今回フルで 26）
Fable Verdict: ACCEPT（条件付き）
関係: TIGHTEN
POの答え（1文）: results.json 実測の内訳（cleanup 残置 17 · recheck 済ハーネス起因 4 · 実質未検証 5）に基づき「製品バグにしない」は維持するが、無視してよいのは前 2 群 21 件だけで、実質未検証 5 件（S11 trimming create / S13 owner_link / V02 inventory F4 / V03 owner F4 / V03 permission-group name F1）は #254 residual disposition の対象とする — S13 は H6 が吸収、残り 4 件は新行 UAT-H7（§G）で 1 セッション一括 spot-check または理由付き ACCEPT。
次の人: USER（local）
次の一手: H5 と同一セッションで H7 実施。cleanup 残置 17 の手動削除は義務化しない（local DB は OPS-2 承認済み fresh で破棄可能 — 削除工程の新設は②違反）
完了条件 / 空欄に残すセル: H7 各結果（または ACCEPT 理由）
```

### E. #254 close ゲート（専用）

**「local FAIL 0 のみ」での close は不可。** PASS 1352 は下記 1. を充足しただけである。close してよい最小条件:

1. local フル UAT FAIL 0 証跡 — **充足済み**（`reports/uat-2026-08-14/FINAL.md` · build = `1386e1db0`）
2. **UAT-H1**: current main の STG で実 LINE 通知 1 往復（mock 不可）
3. **UAT-H2**: 実 token で LIFF 起動〜主要画面（同一 window・mock 不可・値非記録）
4. **UAT-H3**: audit_logs 監査行の確認（local 可・read-only）
5. **UAT-H4**: 締め fixture 境界 1 回、**または**区分ごとの明示スキップ理由
6. **UAT-H5 / H6 / H7**: 完了、**または**理由付き disposition（silent 省略は不可）
7. **UAT-S1** 受容記録 — **充足済み**
8. **E-6 一行の全欄**（residual_disposition=APPROVED · final_signoff=APPROVED · opaque ref）+ **受入 build SHA の明記**を close 承認者（別 USER）が埋める — close 時点の main が受入 build から前進している場合（現に `697d5c597` が 001_init.sql +36 行を含む）、schema / 挙動に触れる差分は該当シナリオのみ再確認する（全再走は不要）

緩和の可否: H1〜H3 に緩和はない — 実環境・実 token・実 DB でしか観測できない受入条件そのものだ。許される緩和は「実行 → 理由付き disposition への置換」（H4〜H7）だけである。PROD 構築・Go-live window・#252 値投入は #254 の条件に**含めない**（#253 / #257 レーン）。旧 STG（merge 前）での代替実施は証拠にならない。local レーン（H3〜H7）は `697d5c597` を pull + `make migrate`（mismatch なら OPS-2 fresh）後の build で実施し、実施 build SHA を証跡に記す。

### F. USER が今日送る / 止める

**送ってよい（最大 5）**

1. #201 臨床 bundle 記入の人間チャネル催促（Issue コメントは 8/14 投稿済み — 返答期限を添えて 1 通）
2. #256 U13 status 確認依頼（納品責任者へ — 完了/未完の 1 語回答でよい）
3. #258 REVISE 版 E-11 の送付（契約チャネル可・GitHub 投稿不要）
4. staging preflight 残セル記入依頼（checksum/ownership disposition・backup owner）+ **PlanetScale 109 オブジェクト REASSIGN のサポート依頼の先行送付**（runbook 記載の唯一の解・リードタイム長のため今日送ってよい）
5. PO-008 承認/修正依頼（§5.1 の 6 行）

**止める（最大 5）**

1. #254・#253 の close、および「local FAIL 0 = 納品完了/受入完了」と読める一切の対外表明
2. #299 の green 前 merge・squash・staging reset
3. TASK-021 C/D・LINE-R05 の DROP / presence guard 除去（inventory ゼロ前）
4. TASK-033 着手・#201/#249/#211 値の代理記入（値発明）
5. #257 新 window 設定・#250/#259 への本日再催促（8/14 催促済み）・agent による STG/PROD migrate/reset

### G. todo-po.md 更新提案（差し替え行のみ）

§1 に追加（1 行）:

```text
| **UAT-H7** | S11/V02/V03 | **残 PARTIAL spot-check** — trimming 新規（S11）・inventory 登録 F4（V02）・owner 登録 F4・permission-group 名称 F1（V03）を人手で各 1 回 | local | open | 各 1 件成功、または理由付き ACCEPT を記録（NG なら todo.md §2 へ BUG 起票） |
```

§2 #254 行の「次の一手」差し替え:

```text
**§1 UAT-H1〜H7** 完了（H4〜H7 は理由付き disposition 可）+ 受入 build SHA 明記 + 別 USER 承認 → close 裁定
```

todo.md 冒頭対応表の「PO 確認待ち」行差し替え（drift 修正 — 現に §1 に 6+1 件・§2 に 4 件 open がある）:

```text
| PO 確認待ち | **[`todo-po.md`](todo-po.md)**（§1 人間レーン 7 件 + §2 裁定 4 件 open · 予約 2026-11-07） |
```

補足（1 行・行追加不要）: `697d5c597`（origin/main 先行 1 commit）は 001_init.sql 編集を含むため、本 worktree の次回 pull は POST-PULL 該当 — `make migrate`、checksum mismatch 時は OPS-2 local fresh。
