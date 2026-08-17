# Claude Fable — 実施セッション設計回答（2026-08-14 · 第 2 弾）

前回裁定（RATIFY 21 · TIGHTEN 4 · OVERTURN 0）に拘束される。Verdict の再審はしない。

### A. 読んだもの / 読めなかったもの

読んだもの

- 依頼書 `reports/fable-po-confirm-request-2026-08-14-exec-session.md` 全文
- 前回回答 `reports/fable-po-confirm-answer-2026-08-14-uat-human.md` 全文（自筆・C/E/F/G）
- `todo-po.md` 全文再読（**裁定取込済みを確認** — H1〜H7・#254 ゲート・送付/停止リスト反映済み）· `todo.md` 全文再読（§4.2/§5/§5.1 + 新規 BRT-37〜52 行）
- `reports/uat-2026-08-14/FINAL.md` + `results.json`（前ラウンド実測・変更なし）· `docs/product-philosophy.md` · Sol r2 §E · `docs/ops/infra/staging/runbook.md`（前ラウンド読了）
- **シナリオ正本 4 本全文（今回）**: S06（audit 手順 9 = USER DB 参照）· S09（締め fixture: 城東 clinic AM09:00/境界13:30/終了19:00・completed_at 5 種は「ローカル DB のテストデータ更新で準備」と明記）· S11（trimming 新規 = 手順 1）· S13（identity-links 手順 1–8・2 医院所属 + view/edit 権限・anchor 全医院所属要件）
- `docs/ops/deploy/LOCAL_DB_RESET.md`（`make reset` = 回復可能 fail-closed 単一入口・snapshot を `.local-db-backups/<UTC>/` に自動作成・**checksum mismatch の非破壊経路は local に存在しない**）
- `reports/todo-walk-2026-08-14/drafts/E-11-258.md`（**stale 参照なし・U9/U12「#253 後」明記 — Fable REVISE 相当と本日確認**）
- FIELD-LEVEL-PROTOCOL（F1=必須空拒否 · F4=有効値永続の定義）
- 実測: `frontend/src/config/paths.ts`（シフトテンプレ = `/settings/shift-templates`）· `backend/internal/model/audit_log.go`（clinic_id/actor_id/actor_type/action/resource/resource_id/old_value/new_value 列）· Makefile（`migrate`・`reset` あり）· docker-compose.yml（DB 接続は env 変数）
- git read-only 実測: HEAD=`1386e1db0` · `origin/main`=`697d5c597` · dirty 42 files — **うち `docs/spec/screens/README.md`・`docs/spec/specification.md` は 697d5c597 も変更しており、未コミットのまま pull すると衝突で拒否される**（→ §D 手順に反映）

読めなかったもの

- GitHub live / STG / PROD（接続禁止）· Linear BRT 実 state（repo 外）

### B. 再審

再審なし。前回 UAT-human 裁定と DEC-40〜68 を維持。OVERTURN 0。TIGHTEN もなし — 本回答は確定済み Verdict の実施手順への具体化のみである（pull 前 WIP commit 等はすべて手順であり境界変更ではない）。

### C. Executive（実施版）

**既定プラン = ケース B（90 分）:**

1. 台帳・docs の未コミット変更を path 指定で WIP commit → `git pull --ff-only` → `make migrate`（checksum mismatch 想定）→ `make reset` を起動（無人で走る）。
2. reset の待ち時間に M1〜M5 を送付する（5 通・約 30 分）。
3. postflight green を確認し、実測 HEAD SHA（`697d5c597` 想定）を記録する。
4. H3（audit）→ H5（シフトテンプレ）→ H7（4 件 spot-check）を実施し、`reports/uat-human-2026-08-14/` に証跡を残す。
5. H4・H6 は次セッションへ明示持ち越し（silent 省略ではなく SESSION.md に記載）。
6. reset が fail-closed 停止した場合は volume 無傷 — 現 build（実測 HEAD）で H3/H5/H7 だけ実施し H4 を延期する。
7. merge・close・値記入・#250/#259 再催促はしない。
8. 終了時に todo-po §1 の Status と SESSION.md を更新する。

**今日の成功定義（5 項目）**

- [ ] M1〜M5 送付済み（禁止フレーズ 0）
- [ ] 全証跡に実施 build SHA が記載されている
- [ ] 実施した H 行の Status が `done(PASS)` / `done(ACCEPT_DISPOSITION)` / `blocked(BUG-xxx)` のいずれかで更新されている（無言スキップ 0）
- [ ] FAIL は `todo.md` §2 に BUG-xxx（接頭辞横断の単一連番）で起票されている
- [ ] merge・Issue close・値発明・再催促の実行 0 件

**今日の失敗定義（5 項目）**

- #254/#253 の close、または「local FAIL 0 = 納品完了」と読める記述をどこかに書く
- green 前の #299 merge / squash / STG reset
- 臨床値・契約値・日付の代理記入（M1〜M5 のプレースホルダを自分で埋める）
- SHA なし証跡・シナリオ md への結果追記
- H の未実施を disposition なしで放置する

### D. pull 方針（P1 を採る）

**P1 — 先に pull + `make migrate`（mismatch なら OPS-2 `make reset`）のうえ H3〜H7 を実施し、実施 SHA を `697d5c597` 系で記録する。**

理由（1 文）: 前回確定裁定と todo-po §0 が既に「H3〜H7 は pull 後 build で実施」を採っており、かつ 001 編集の中身（cash_register_close_adjustment）は H4 の検証対象そのものだから、先に揃えて全 H を単一 SHA で記録するのが再確認の二度手間と証跡分裂を防ぐ。

失敗時退避（1 行）: `make reset` は削除前に owner-only snapshot（`.local-db-backups/<UTC>/`）を作る fail-closed 単一入口 — 停止時は volume 無傷なので、その日は現 build（実測 HEAD を記録）で H3/H5/H6/H7 のみ実施して H4 を延期し、復旧は snapshot から行う。

実施上の注意（実測に基づく）: 本 worktree は `docs/spec/screens/README.md`・`docs/spec/specification.md` が未コミット変更のまま 697d5c597 と重なるため、**pull 前に台帳・docs・reports を path 指定で WIP commit する**（`git commit -- <paths>`。stash や discard 系は使わない）。

### E. 時間別セッション表

**ケース A — 30 分（送付と 1 語判断のみ · pull しない）**

| # | 分 | やること |
|---|-----|----------|
| 1 | 0–5 | M2（#256 U13 · 1 語回答依頼）送付 — 最短で返る判断を先に投げる |
| 2 | 5–12 | M1（#201 臨床催促）送付 |
| 3 | 12–20 | M4（リリース + STG DB · PS109 込み）送付 |
| 4 | 20–25 | M3（#258 · E-11 添付）送付 |
| 5 | 25–30 | M5（PO-008）送付 + todo-po に「送付 5/5」1 行メモ |

やらない: pull / migrate / reset / local H 全部 / merge / close / #250·#259 再催促。
成果物: 送付 5 通 + 送付記録 1 行。
#254 距離: H done 0。ただし H1/H2 の前提（staging）と #201/#256/#258 の待ち行列が全部起動する。

**ケース B — 90 分（既定 · 送付 + local 一部）**

| # | 分 | やること |
|---|-----|----------|
| 1 | 0–10 | WIP commit（path 指定）→ `git pull --ff-only` → `make migrate`（mismatch 想定） |
| 2 | 10–15 | mismatch 確認 → `make reset` 起動（無人で走らせる） |
| 3 | 15–45 | reset 並走中に M1〜M5 送付（ケース A の 5 通と同一） |
| 4 | 45–50 | postflight green 確認 · 実測 HEAD SHA 記録 · demo アカウントでログイン確認 |
| 5 | 50–65 | **H3**: カルテ作成 → 確定 → psql read-only で audit 行確認 |
| 6 | 65–75 | **H5**: `/settings/shift-templates` 新規 1 件 → 再オープン永続 |
| 7 | 75–90 | **H7**: 4 件 spot-check + SESSION.md 記録 |

やらない: H4 / H6（次セッションへ明示持ち越し）· merge · close。
成果物: 送付 5 通 + `reports/uat-human-2026-08-14/`（H3 / H5 / H7 / SESSION.md）。
#254 距離: **H3 done · H5 done · H7 done（最大）**。残 = H1/H2（staging 待ち）· H4 · H6。

**ケース C — 3 時間（送付 + local H3〜H7 を可能な限り閉じる）**

| # | 分 | やること |
|---|-----|----------|
| 1 | 0–15 | ケース B の 1–2（commit → pull → migrate → reset 起動） |
| 2 | 15–45 | M1〜M5 送付（reset 並走） |
| 3 | 45–50 | postflight 確認 + SHA 記録 |
| 4 | 50–65 | **H3** audit |
| 5 | 65–75 | **H5** シフトテンプレ |
| 6 | 75–95 | **H7** 4 件 |
| 7 | 95–140 | **H6**: fixture 構築（§H の 5 行手順）→ S13 手順 1–6（7–8 は時間があれば） |
| 8 | 140–170 | **H4**: 会計 1 件を精算 → `/accounting/close` プレビュー → 実レジ金額入力 → 締め確定 → `/accounting/close/history` 確認。残時間で S09 の completed_at fixture（SQL）→ 13:30 境界 / EMG / 越日プレビュー |
| 9 | 170–180 | SESSION.md に E-6 途中行 + 各 disposition · todo-po §1 Status 更新 |

やらない: merge · close · STG 系 · 値記入。
成果物: 送付 5 通 + H3〜H7 の done または理由付き disposition + SESSION.md。
#254 距離: **local 側 5/7 が done または disposition** — 残りは H1/H2（= staging merge → deploy → health 待ち）のみ。

### F. 送付文完成物（M1〜M5）

**M1 — 臨床（#201）· KEEP**

件名: `【#201】救急投薬 臨床承認 bundle 記入のお願い（期限 [  ]）`

> #201 に記入用の bundle 表を投稿済みです（Issue #201 のコメント参照）。全列が空欄のまま止まっており、この値はあなたにしか書けません。**[  ] までに**全列（対象・policy・単位・出典・approver role・発効日・opaque ref）の記入をお願いします。
> 上限・warning は「現行継続」と書くだけでも構いません（再計算不要）。実氏名・患者情報・承認資料本文は返信に書かないでください。全列が揃うまで実装（TASK-033）は着手しません。

禁止フレーズ: 「こちらで仮の値を入れておきます」「だいたいで OK」（値の発明・代筆の申し出）。

**M2 — 納品（#256 U13）· KEEP**

件名: `【#256】U13 操作説明会 — 完了 / 未完 の 1 語だけください`

> #256 の close に残る判断セルは U13（操作説明会）だけです。**「完了」または「未完」の 1 語**でご回答ください。
> 完了なら: 実施日・発効日・opaque ref を添えてください。未完なら: それだけで結構です（Issue は open 維持。日程はまだ決めなくてよい）。

禁止フレーズ: 「未完でも close で進めます」（U13 未完のままの close は禁止）。

**M3 — 契約（#258）· KEEP**

件名: `【#258】DELIVERY_PACKAGE U1〜U12 一括記入のお願い`

> 納品ドキュメントの管理者・契約欄の記入依頼です。**添付の記入表**（`reports/todo-walk-2026-08-14/drafts/E-11-258.md` — stale 参照なしを確認済み）に沿って、A/B 選択と U1〜U8/U10/U11 をご記入ください。
> 推奨は A（クライアント所有）です。U9・U12 は Production 構築後に開発側が記入するため空欄のままで構いません。U13（説明会）は #256 側で扱うため本依頼に含みません。金額・secret・実 email・Go-live 日付は記入しないでください。

禁止フレーズ: 「納品完了に伴い」（#258 の記入は納品完了を意味しない — PROD 未構築）。

**M4 — リリース + STG DB（preflight 残セル + PS109）· KEEP**

件名: `【#299】staging merge 前の残り 3 セル + PlanetScale REASSIGN 依頼の送付`

> draft PR #299（main→staging）は残り 3 点が埋まれば merge 判断に入れます。
> ① **migration checksum / PlanetScale 109 オブジェクト所有権の disposition** — staging DB は 001 統合前の適用履歴を持ち、直近 main でも `001_init.sql` が再編集されています。STG_PLANETSCALE_SEED_RUNBOOK 準拠で解消方針を 1 行確定してください（**STG reset は不可**）。
> ② **deploy 前 backup 実施 + rollback owner role** の記入。
> ③ **PR #299 の required CI 全 green** 確認（`staging-preflight-status.md` の CI 一覧どおり）。
> あわせて、**PlanetScale サポートへ 109 オブジェクトの REASSIGN 依頼を先行送付**してください（runbook 記載の唯一の解・ALTER 系 migration の前提・リードタイムが長い）。全 green まで merge しません。merge 時は merge-commit のみ（squash 禁止）。

禁止フレーズ: 「とりあえず merge して STG で確認」（green 前 merge の解禁）。

**M5 — クライアント（PO-008）· KEEP**

件名: `【顧客集計】集計仕様 6 点の承認または修正のお願い`

> 顧客集計の現行仕様 6 点について、**承認**か**修正指示**を 1 通でご回答ください（推奨は現行継続です）。
> 1. annual_visit_count = 直近 365 日 rolling ／ 2. annual_amount = From/To → Year → preset → 全期間の優先順 ／ 3. CSV 出力は標準では付けない ／ 4. last_visit と休眠判定は別ロジックのまま統一しない ／ 5. L ステップ書込は default-off・実送信成功のみ成功扱い ／ 6. L ステップ cleanup 失敗は通知必須（本体削除は止めない）。
> 修正の場合は、その項目の業務上の目的と決裁者名 [  ] を添えてください。

禁止フレーズ: 「CSV も付けておきます」（default 追加の先行約束 — 要件は責任者名 + 業務目的が来てから）。

DROP: なし（5 通とも前回 §F 承認済みレーン）。#250 / #259 への再催促文は**作らない**（本日送らない）。

### G. disposition / BUG / E-6 テンプレ

**disposition 固定フォーマット（H4〜H7 用）**

```text
id: UAT-H?
result: PASS | FAIL_BUG | ACCEPT_DISPOSITION
build_sha: <実測 HEAD（git log -1 --format=%h）>
operator: <role 表記可・実名不要>
date: 2026-MM-DD
evidence: reports/uat-human-2026-08-14/<file>（スクショ可・PHI/秘密なし）
if ACCEPT_DISPOSITION:
  reason: <実行を阻んだ具体的欠落（fixture / 権限 / 時間）1〜3 文>
  residual_risk: <何が未証明のままか（例: 越日 EMG の UI 実観測なし — BE 回帰テストは有）>
  not_a_product_pass: true
```

**ACCEPT_DISPOSITION の許可条件（再確認）**

- 許されるのは **H4〜H7 のみ**。**H1〜H3 は不可**（実 LINE・実 token・実 DB audit は disposition で代替できない — 前回 §E のまま）。
- reason は「やらなかった」ではなく「何が欠けて実行できなかったか」。residual_risk は必須。`not_a_product_pass: true` 固定 — **PASS の代用ではない**。
- 全 disposition は #254 final sign-off（別 USER）の読み上げ対象。

**FAIL_BUG のとき — `todo.md` §2 起票 1 行テンプレ**

```text
### BUG-xxx: <1 行要約>
- 発見: UAT-H? 人手実施 · build <sha> · 2026-MM-DD
- 再現: <URL> → <3 ステップ以内>
- 期待: <scenarios/Sxx §n の期待結果>
- 実測: <観測事実>
- 証跡: reports/uat-human-2026-08-14/<file>
```

番号は **接頭辞横断の単一連番** — 起票前に git 履歴・台帳の最終 BUG/TASK 番号を実測して +1 する（直近既知は BUG-038 系。衝突注意）。

**E-6 一行への畳み方（例 — 値は実施結果で置換）**

```text
#254 human lane: H1=[PENDING_STG] H2=[PENDING_STG] H3=[PASS] H4=[ACCEPT_DISPOSITION(reason=…)] H5=[PASS] H6=[PASS] H7=[3PASS+1FAIL_BUG(BUG-xxx)] build=[697d5c597系実測] uat_full=FAIL0@1386e1db0 residual_disposition=[PENDING] final_signoff=[PENDING] opaque_ref=[ ]
```

H1/H2 が PENDING の間、residual_disposition / final_signoff は埋めない（close 裁定に進まない）。

### H. local 手順書（H3 → H5 → H7 → H6 → H4 · 推奨順を確定）

共通前提: P1 完了後（pull + reset 済・postflight green）。ログインは `.env.local` の demo/E2E アカウント（値を書かない）。記録先ディレクトリ: **`reports/uat-human-2026-08-14/`**（新規）。全ファイル冒頭に build SHA を記す。

**H3 — audit_logs（S06 手順 1→4→9 の最小径路）**

| 項目 | 内容 |
|------|------|
| 前提 | vet ロール（medical-records create/edit）· 生存ペット 1 頭（ペット検索） |
| 操作 | ① カルテ新規作成: 主訴 + S/O/A + 診断名 → 保存 ② 主訴を 1 箇所変更して再保存 ③ 右下フローティング「確定する」→ ダイアログ確認 → 確定 ④ ターミナルで `docker compose exec db sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"'`（**USER 手動 · read-only**）⑤ `SELECT id, clinic_id, actor_id, actor_type, action, resource, resource_id, created_at FROM audit_logs ORDER BY created_at DESC LIMIT 20;` |
| 合格 | 直近行に作成 / 更新 / 確定に対応する行があり、actor_id・created_at が入っている。確定行は old/new_value を保持（`SELECT old_value, new_value FROM audit_logs WHERE id=<該当>;` で 1 行だけ確認可） |
| 不合格 | 確定に対応する audit 行が無い → **FAIL_BUG**（S06 は確定監査を同一 tx fail-closed と規定）。psql 接続不可 → 再試行 → 環境問題なら当日 disposition 不可（H3 は disposition 禁止・別日再実施） |
| 記録先 | `reports/uat-human-2026-08-14/H3-audit.md`（record の実 ID は書かず「対象 1 件」表記 + 行数・action 種のみ） |

**H5 — シフトテンプレ SidePanel**

| 項目 | 内容 |
|------|------|
| 前提 | admin · `http://localhost:3003/settings/shift-templates` |
| 操作 | ① URL を開く ② 新規作成を開く（SidePanel）③ 名称等の必須項目に合成値を入力 ④ 保存 ⑤ パネルを閉じ、一覧から再オープン |
| 合格 | 保存成功 + 再オープンで同値が初期表示（F4 相当の永続） |
| 不合格 | SidePanel が開かない / 保存が無音失敗 / 再オープンで消える → **FAIL_BUG**。一時的描画不良は 1 回だけ再読込して再試行 |
| 記録先 | `reports/uat-human-2026-08-14/H5-shift-template.md` |

**H7 — 残 PARTIAL 4 件 spot-check（H5 と同一セッション）**

| # | 対象 | 操作 | 合格 |
|---|------|------|------|
| 1 | S11 trimming 新規 | `/trimming` → 新規登録 → 生存ペット選択 · ステータス「予約」· コース 1 + オプション 1 → 保存 | 保存成功 · 一覧に「予約」表示 · マスタ価格が表示 |
| 2 | V02 inventory F4 | `/inventory/new` → 名称に有効値 + 必須項目 → 保存 → 一覧/詳細 → 再読込 | 同値が永続・初期表示 |
| 3 | V03 owner F4 | `/owners/new` → 合成名（例: テスト太郎）+ 必須項目 → 保存 → 詳細 → 再読込 | 同上 |
| 4 | V03 permission-group F1 | `/settings/permission-groups` → 新規 SidePanel → **名称を空のまま保存** → その後正しい名称で保存 | 空保存は明示エラーで拒否（無音失敗禁止）· 正名は保存成功 |

不合格: いずれも **FAIL_BUG**（再現 3 ステップ付き）。実行できない欠落（マスタ不足等）は ACCEPT_DISPOSITION 可。記録先: `reports/uat-human-2026-08-14/H7-spot-checks.md`（4 件を 1 ファイル）。

**H6 — identity-links フル（S13 手順 1–6 必須 · 7–8 は推奨）**

最小 fixture 手順（5 行以内 · ID 直書き禁止 · 検索条件で特定）:

1. admin で `/settings/staff` → 実施スタッフに **八王子・城東の 2 医院所属**を付与する。
2. `/settings/permission-groups` で `identity-links` の **view と edit** を含むグループを確認/作成し、当該スタッフへ割当（identity-links は自動付与されない — S13 明記）。
3. 医院切替で各医院の飼主検索を開き、**同姓の飼主が両医院にいるか検索**。いなければ各医院で合成名の飼主 + ペット 1 頭を新規作成（「同一人物」の対応メモはローカルのみ・repo に書かない）。
4. 当該スタッフで**再ログイン**（権限・所属を反映）。
5. `/identity-links` を開き、両医院の対象飼主が検索に出ることを確認してから S13 手順へ。

| 項目 | 内容 |
|------|------|
| 操作 | S13 手順 1–6: workbench → 飼主リンク → ペットリンク → 連携履歴（include_linked）→ 誤組合せ unlink → 正組合せ relink。時間があれば 7（view-only 別アカウント）・8（所属欠け 403） |
| 合格 | 各手順の期待結果どおり（リンク作成 / 履歴表示 / soft-delete / relink 冪等）。他 clinic 非リンクペットが履歴に出ないこと |
| 不合格 | 期待と異なる挙動 → **FAIL_BUG**。fixture が組めない（権限画面で edit 付与不能等）→ ACCEPT_DISPOSITION（reason に欠落を明記）。手順 7–8 未実施はその旨を disposition に 1 行 |
| 記録先 | `reports/uat-human-2026-08-14/H6-identity-links.md`（clinic 内部 ID 可 · 氏名等 PHI 不可 — S13 サインオフ表の粒度で） |

**H4 — 締め境界（S09 · 最小 1 締め → 余力で fixture 拡張）**

| 項目 | 内容 |
|------|------|
| 前提 | 執行ロール（cash-register-close）· 対象 = 城東センター病院（AM09:00 / 境界 13:30 / 終了 19:00 — seed 値）· 平日日付 |
| 操作 | ① `/settings/closing-time` でレンジプレビュー確認 ② 会計 1 件を精算完了（実時刻）③ `/accounting/close` で対象日 + 当該区分のプレビュー → 実レジ金額入力 → 締め確定 ④ `/accounting/close/history` で締め日時・区分・過不足を確認 ⑤ 余力があれば S09 前提どおり **local DB の completed_at 更新**（10:00 / 13:30:00 / 14:00 / 20:00 / 翌日 02:00 の 5 件）→ AM/PM/EMG/越日プレビュー確認（S09 手順 2–6） |
| 合格 | 最小径路: 締め 1 回が保存され履歴に出る（= todo-po の完了条件）。拡張分: 13:30 ちょうどは PM 帰属・越日 02:00 は前日 EMG |
| 不合格 | 締め保存 / 履歴 / 過不足計算の欠陥 → **FAIL_BUG**。fixture（completed_at 5 種）未準備で拡張分を実施しない場合 → 区分ごとの **ACCEPT_DISPOSITION**（reason=fixture 未準備 · residual_risk=境界/越日帰属の UI 実観測なし — BE 回帰テスト済） |
| 注意 | 締めは append-only + UNIQUE(clinic, date, period) — 同日同区分は再締め不可（local はいずれ reset で破棄可なので実害なし）。SQL 更新はテストデータ準備として S09 が明示的に規定する範囲のみ |
| 記録先 | `reports/uat-human-2026-08-14/H4-closing.md` |

セッション末尾: `reports/uat-human-2026-08-14/SESSION.md` に E-6 途中行（§G 例の形式）+ 実施 build SHA + 持ち越し一覧を書き、todo-po §1 の Status を更新する。

### I. 分岐 YES/NO（§3.6）

1. **U13「未完」の日、#256 は open のまま何を待つ?** — 待つのは **U13 の実施日程確定 → 実施 → 発効日 + 別 close 承認**だけ（それ以外の新条件を足さない。CLOSE_RECOMMEND は維持）。
2. **staging checksum が永久 mismatch のとき merge を止めるか?** — **runbook disposition で進める**: local 用 `make reset` は STG に流用せず、STG_PLANETSCALE_SEED_RUNBOOK の承認済み経路（+ PS109 REASSIGN 完了）で mismatch を解消し、その disposition が green になってから merge — 「mismatch のまま merge」も「無期限停止」もしない。
3. **H7 の 1 件だけ FAIL_BUG・他 3 件 PASS のとき「残り 1 BUG 修正待ち」と書いてよいか?** — **YES**。他 3 件の PASS は有効のまま維持し、BUG-xxx 起票 + 修正後は**該当 1 件のみ再確認**（全再走不要）。ただし「残り 1 BUG」= H7 内の話であり、#254 全体は依然 H1/H2 待ちである旨を併記する。

### J. todo-po.md への追記提案（行追加のみ）

§1 表の直下に 2 行:

```text
証跡置き場: `reports/uat-human-2026-08-14/`（SESSION.md + H 別ファイル · 各ファイル冒頭に実施 build SHA 必須）
Status 更新ルール: open → done(PASS) / done(ACCEPT_DISPOSITION) / blocked(BUG-xxx) の 3 値のみ · 更新は人間 · 無言スキップ禁止
```

§0 の worktree 注意の直後に 1 行（pull 手順の固定化）:

```text
pull 手順: 未コミットの台帳/docs/reports を path 指定で WIP commit → `git pull --ff-only` → `make migrate` → mismatch なら `make reset`（snapshot 自動作成 · USER のみ）
```
