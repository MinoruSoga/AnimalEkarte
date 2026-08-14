# todo-po.md — PO / 人間実施

| 項目 | 値 |
|------|-----|
| **更新** | 2026-08-14（Fable exec-session 取込） |
| **読者** | あなた（PO）・人間オペレータ |
| **最新裁定（Verdict）** | [`reports/fable-po-confirm-answer-2026-08-14-uat-human.md`](reports/fable-po-confirm-answer-2026-08-14-uat-human.md)（RATIFY 21 · TIGHTEN 4 · OVERTURN 0） |
| **最新実施設計** | [`reports/fable-po-confirm-answer-2026-08-14-exec-session.md`](reports/fable-po-confirm-answer-2026-08-14-exec-session.md)（P1 · ケース B 既定 · M1–M5 · 手順） |
| **agent 実行の正本** | [`todo.md` §5](todo.md) |
| **確定済み裁定（土台）** | [`todo.md` §4](todo.md) · DEC-40〜68 維持 |
| **local UAT 証跡** | [`reports/uat-2026-08-14/FINAL.md`](reports/uat-2026-08-14/FINAL.md)（FAIL 0 · PASS 1352 · build **1386e1db0**） |
| **人手実施証跡** | [`reports/uat-human-2026-08-14/`](reports/uat-human-2026-08-14/)（SESSION.md + H 別ファイル） |
| **方針** | [`docs/work/decisions/fable-po-recommendation.md`](docs/work/decisions/fable-po-recommendation.md) |

**このファイル:** 人が実施・裁定・実機確認するものだけ。  
**製品バグ:** `todo.md` §2。シナリオ md に結果を書かない。

**再審:** なし。exec-session は Verdict の具体化のみ（OVERTURN 0 · TIGHTEN なし）。

**今日 Yes トップ 3:** (1) #201 催促 (2) staging preflight / #299 残セル (3) #256 U13 明示  
**絶対 No:** #254/#253 close · green 前 merge · DROP · 値発明 · TASK-033 · #257 window · #250/#259 再催促

### worktree / pull（Fable exec · P1）

UAT 証跡 build = `1386e1db0`。`origin/main` = `697d5c597`（`001_init.sql` 編集 · H4 対象に隣接）。

**pull 手順:** 未コミットの台帳/docs/reports を **path 指定で WIP commit** → `git pull --ff-only` → `make migrate` → mismatch なら **`make reset`**（snapshot 自動 · **USER のみ** · STG に流用しない）。

- 失敗時: volume 無傷 → その日は現 HEAD で H3/H5/H6/H7 のみ · **H4 延期** · 実施 SHA を実測で記録  
- H3〜H7 は原則 pull 後単一 SHA で記録  
- stash / discard-all は使わない  

---

## 1. UAT 人間レーン

**Status 更新ルール:** `open` → `done(PASS)` / `done(ACCEPT_DISPOSITION)` / `blocked(BUG-xxx)` の **3 値のみ** · 更新は人間 · **無言スキップ禁止**  
**証跡置き場:** `reports/uat-human-2026-08-14/`（各ファイル冒頭に **実施 build SHA 必須**）  
**disposition テンプレ:** [exec-session §G](reports/fable-po-confirm-answer-2026-08-14-exec-session.md) · **H1–H3 は disposition 不可**

| ID | シナリオ | 実施内容 | Verdict | #254 必須 | Status | 完了条件 |
|----|----------|----------|---------|-----------|--------|----------|
| **UAT-H1** | S04 | 実 LINE プッシュ 1 往復 | DO_NEXT | **Y** | open | #299 merge → STG deploy → health 後。mock 不可 |
| **UAT-H2** | S12 | 実 LINE / LIFF token | DO_NEXT | **Y** | open | H1 と同一 window。token 値は記録しない |
| **UAT-H3** | S06 | audit_logs DB 参照 | DO_NEXT | **Y** | open | local read-only。確定・追記 1 件。手順: exec-session §H |
| **UAT-H4** | S09 | 締め fixture 境界 | DO_NEXT | **Y**（disp 可） | open | 締め 1 回 or 区分ごとの ACCEPT_DISPOSITION。**90 分プランでは持ち越し可** |
| **UAT-H5** | V04 | `/settings/shift-templates` | DO_NEXT | **Y** | open | 新規 1 · 再オープン永続。NG→BUG |
| **UAT-H6** | S13 | 2 医院 identity-links | DO_NEXT | **Y** | open | link→history→unlink→relink。**90 分プランでは持ち越し可** |
| **UAT-H7** | S11/V02/V03 | PARTIAL spot-check 4 件 | DO_NEXT | **Y**（disp 可） | open | trimming / inventory F4 / owner F4 / permission-group F1 |
| **UAT-S1** | S08 | 部分入金仕様受容 | ACCEPT | Y（済） | **done** | BUG にしない |

### 実施しない / 無視してよい

| 項目 | 扱い |
|------|------|
| campaign / inquiry F4 初回 | ハーネス · recheck 201 |
| auth-login timeout | セッション競合 |
| PARTIAL cleanup 17 | 手動削除義務なし |
| PARTIAL recheck ハーネス 4 | 無視可 |

フル UAT 証跡: `reports/uat-2026-08-14/` · build **`1386e1db0`**

---

## 2. UAT 関連 PO 裁定

| ID | Verdict | Status | 次の一手 |
|----|---------|--------|----------|
| **#254 close** | **HOLD** | open | H1–H7（H4–H7 は disp 可）+ SHA + **別 USER** sign-off。local FAIL 0 単独 **不可**。E-6 一行: exec-session §G |
| **#256 close** | CLOSE_RECOMMEND | open | U13 を **1 語**明示。未完→open 維持のみ待つ（新条件を足さない） |
| **staging #299** | DO_NEXT | open | green 前 merge 禁止。checksum disposition（runbook）· backup owner · CI · PS109 REASSIGN |
| **実 LINE OPS-4/5** | DO_NEXT | open | **#299 merge → deploy → health → H1/H2** |
| **PO-06 記録** | done | done | close は #254 |

### #254 close 最小条件（不変）

1. local フル UAT FAIL 0 — **充足**（`1386e1db0`）  
2–3. **H1 · H2**（STG · mock 不可 · 緩和なし）  
4. **H3** audit（local · 緩和なし）  
5. **H4** 締め or disposition  
6. **H5/H6/H7** 完了 or disposition  
7. **UAT-S1** — 充足  
8. E-6 全欄 + build SHA · 別 USER final_signoff  

---

## 3. 人が埋める空欄 + 今日の送付

詳細テンプレ: [`todo.md` §5](todo.md)。**送付文完成物（コピペ）:** [exec-session §F M1–M5](reports/fable-po-confirm-answer-2026-08-14-exec-session.md)

| # | ID | Verdict | 今日の一手 |
|---|----|---------|------------|
| 1 | #201 | DO_NOW | **M1** 送付（期限 `[  ]` を人が埋める） |
| 2 | staging / #299 | DO_NOW | **M4** 送付 |
| 3 | #256 U13 | DO_NOW | **M2** 送付（1 語回答依頼） |
| 4–5 | #249 · #211 | DO_NEXT | 記入待ち · 値発明禁止 |
| 6 | #258 | DO_NEXT | **M3** + `drafts/E-11-258.md` 添付 |
| 7–8 | OPS-1 · presence | DO_NEXT | E-12 / E-14 別途 |
| 9–10 | #250 · #259 | WAIT_EXTERNAL | **本日再催促しない** |
| 11 | PO-008 | DO_NEXT | **M5** 送付 |

**止める:** #254 close · green 前 merge · DROP · 値発明 · TASK-033 · #257 window · agent STG migrate

**送付記録:** 下表（人間が更新）

| 通 | 宛先 | Status | 送付日 |
|----|------|--------|--------|
| M1 | #201 臨床 | open | — |
| M2 | #256 U13 | open | — |
| M3 | #258 契約 | open | — |
| M4 | #299 リリース/STG | open | — |
| M5 | PO-008 クライアント | open | — |

---

## 4. 既定セッション（ケース B · 90 分）

正本手順: [exec-session §C–E · §H](reports/fable-po-confirm-answer-2026-08-14-exec-session.md)

1. path 指定 WIP commit → pull --ff-only → migrate →（mismatch）reset 起動  
2. reset 待ちに **M1–M5 送付**  
3. postflight green · **実測 HEAD SHA 記録**  
4. **H3 → H5 → H7**  
5. **H4 · H6 は持ち越し**（SESSION.md に明記 · silent 省略禁止）  
6. merge/close/値記入なし · todo-po Status + SESSION.md 更新  

**成功:** 送付 5 · SHA 付き証跡 · Status 3 値のみ · BUG は §2 · 禁止行為 0  
**失敗:** close 表明 · green 前 merge · 値代筆 · SHA なし · disposition なしスキップ  

| 時間 | プラン | 参照 |
|------|--------|------|
| 30 分 | ケース A: 送付のみ · pull なし | exec-session §E |
| **90 分** | **ケース B（既定）** | 上表 |
| 3 時間 | ケース C: + H6 + H4 | exec-session §E |

---

## 5. 予約

| いつ | 対象 | 内容 |
|------|------|------|
| **2026-11-07** | F-021-X | 無応答でも自動 ACCEPT しない · 再裁定を本ファイルへ |

---

## 6. カード追加ルール

- 人間レーン → §1  
- DEC 覆し → 新カード → `todo.md` §4  
- 製品欠陥 → `todo.md` §2 `### BUG-xxx`（単一連番）  
