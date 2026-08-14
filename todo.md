# todo.md — 技術・エージェント作業台帳

| 項目 | 値 |
|------|-----|
| **更新** | 2026-08-14（Open Issue 16 件を1件ずつ main 判定 · close 0） |
| **main tip** | `697d5c597` |
| **読者** | agent / 開発 / 実行者 |
| **対になる正本** | [`todo-po.md`](todo-po.md)（PO / 人間レーン · Fable UAT-human 裁定済） |
| **方針** | 規律 · バグ · UAT · 確定裁定（**§4＝旧 §7**）· 着手可能な実行（**§5＝旧 §8**） |

旧 `STATUS.md` / `bug.md` / `PO-todo.md` は本ファイル＋ [`todo-po.md`](todo-po.md) に統合（`STATUS.md` 削除済）。長文履歴は  
[`docs/work/archives/STATUS-before-2026-08-13-slim.md`](docs/work/archives/STATUS-before-2026-08-13-slim.md) と git。

| 内容 | 正本 |
|------|------|
| 技術・バグ・規律・UAT・着手可能な実行 | **todo.md（このファイル）** |
| 確定済み PO 裁定・完了証跡 | **todo.md §4**（旧 §7） |
| 着手可能な実行 | **todo.md §5**（旧 §8） |
| PO 確認待ち | **[`todo-po.md`](todo-po.md)**（§1 H1–H7 open · 実施設計 [exec-session](reports/fable-po-confirm-answer-2026-08-14-exec-session.md) · 証跡 [uat-human](reports/uat-human-2026-08-14/)） |
| Issue 本文 | GitHub |
| 開発規約 | [`.claude/CLAUDE.md`](.claude/CLAUDE.md) · [`AGENTS.md`](AGENTS.md) |
| 採択方針 | [`docs/work/decisions/`](docs/work/decisions/) |

**agent 製品 unit: NONE。** Open 実行行はすべて §5。gate 未充足の作業は §4.2 に置き、解禁しない。

---

## 1. 規律

### Agent

- merge / push は依頼時のみ · migrate を自動適用しない · Done / VERIFIED_FIXED は人間
- シナリオ md は編集しない
- 秘密 · token · 臨床数値 · 契約金額 · 実 identity · Go-live 日付を repo / chat / 本台帳に書かない
- 破壊削除（TASK-021 B/C/D · LINE-R05 DROP）は gate 後のみ（確認待ちは [`todo-po.md`](todo-po.md)）
- 製品コードの Open unit を無断で増やさない（**NONE 維持**）

### 常設（トリガー時だけ実行。§5 の open 行にしない）

1. **OPS-2 local fresh** — checksum mismatch の **local** だけ承認済み fresh（[`LOCAL_DB_RESET.md`](docs/ops/deploy/LOCAL_DB_RESET.md)）。STG / PROD reset はしない
2. **TASK-004 / 005** — screens-drift / closed-pack が land したときだけ
3. **ARCH-A4** — **痛み駆動**のみ。予防的一括リファクタ禁止。ledger: [`arch-a4-trigger-ledger.md`](docs/architecture/arch-a4-trigger-ledger.md)

**POST-PULL:** migration を含む commit を pull したら、利用前に USER が `make migrate`。agent は適用しない。

---

## 2. 受入バグ（Open のみ）

**方針:** 未対応コード欠陥だけ残す。FIXED は削除。

### Open

**なし。**

| メモ | 内容 |
|------|------|
| 仕様（BUG にしない） | S08 部分入金不可（[`todo-po.md`](todo-po.md) UAT-S1） |
| 環境・実機（人間） | **正本 [`todo-po.md`](todo-po.md) §1**（H1〜H7 · 実 LINE · audit · 締め · シフト · S13 · PARTIAL spot-check） |
| UAT 証跡 | [`reports/uat-2026-08-14/FINAL.md`](reports/uat-2026-08-14/FINAL.md)（CDP :9222 フル · 製品 FAIL 0） |

新規バグはここに `### BUG-xxx` を追加。対応後は削除して git に任せる。

---

## 3. UAT（技術ポインタ）

| レポート | 内容 |
|----------|------|
| **アーキテクチャ** | [`docs/ops/testing/TEST_ARCHITECTURE.md`](docs/ops/testing/TEST_ARCHITECTURE.md) |
| **環境** | [`docs/ops/testing/UAT-ENV-SETUP.md`](docs/ops/testing/UAT-ENV-SETUP.md) · [`reports/uat-ready/ENV-STATUS.md`](reports/uat-ready/ENV-STATUS.md) |
| **項目単位** | [`FIELD-LEVEL-PROTOCOL.md`](docs/ops/testing/scenarios/FIELD-LEVEL-PROTOCOL.md) · [`FORM-FIELD-INVENTORY.md`](docs/ops/testing/scenarios/FORM-FIELD-INVENTORY.md) |
| **最新フル** | [`reports/uat-2026-08-14/FINAL.md`](reports/uat-2026-08-14/FINAL.md) · 製品 FAIL **0** · PASS 1352 · PARTIAL 26 · BLOCKED 7 · build `1386e1db0` |
| **Fable（UAT 後）** | [`reports/fable-po-confirm-answer-2026-08-14-uat-human.md`](reports/fable-po-confirm-answer-2026-08-14-uat-human.md) · RATIFY 21 · TIGHTEN 4 · OVERTURN 0 |
| **Fable（実施設計）** | [`reports/fable-po-confirm-answer-2026-08-14-exec-session.md`](reports/fable-po-confirm-answer-2026-08-14-exec-session.md) · P1 · ケース B · M1–M5 |
| **サイドバーマスタ** | [`reports/uat-2026-08-14-sidebar-masters/FINAL.md`](reports/uat-2026-08-14-sidebar-masters/FINAL.md) · PASS 161 · FAIL 0 · 起票 0 |
| **V04 マスタ項目** | [`reports/uat-2026-08-14-v04/FINAL.md`](reports/uat-2026-08-14-v04/FINAL.md) |
| シナリオ正本 | [`docs/ops/testing/scenarios/`](docs/ops/testing/scenarios/)（**結果を書かない**） |

local FAIL 0 は閉証拠にしない。実 LINE / staging merge / Issue close は §4.2 または §5。

---

## 4. 確定裁定（旧 §7） {#7}

USER 採択: [Sol r2](reports/gpt-5.6sol-po-qa-answer-2026-08-14-r2.md) + [Fable 確認](reports/fable-po-confirm-answer-2026-08-14.md)（RATIFY 79 · TIGHTEN 4 · OVERTURN 0 · DEC-40〜68 / Fable pack 維持）。  
値の空欄は **§5** と [`DELIVERY_PACKAGE.md`](docs/delivery/DELIVERY_PACKAGE.md)。証跡は [`docs/work/residual-closeout-ledger.md`](docs/work/residual-closeout-ledger.md)。  
完成物本文の正本は r2 §E（#258 は Fable 修正版 = r2 E-11 現行稿）。

禁止の正本はこの節（とくに §4.2）。旧 §8.9 を消したことは解禁ではない。

### 4.1 完了

| ID | 内容 | 証跡 |
|----|------|------|
| PO-01 · #98 | RDS credential 受容 · GitHub **CLOSED** | `2026-08-08-PO-attestation-F098` · gh 2026-08-14 |
| PO-02 · #99 | ECS 経路 WORKFLOW_ABSENT · GitHub **CLOSED** | `…-F099` · gh 2026-08-14 |
| PO-03 · #252↔#257 | go-live gate に #252=`YES` · window 未設定 | `…-F257-gate252` |
| PO-04 | E2E_LOGIN_* を `.env.local` に SET | 値は書かない |
| PO-05 | Playwright core **80/80** | 2026-08-07 |
| PO-06 · #254 | local scenarios UAT FAIL 0 | [`reports/uat-2026-08-14/FINAL.md`](reports/uat-2026-08-14/FINAL.md) · **close は未了** |
| PO-07 · TASK-021-B | `NO_KNOWN_EXTERNAL_CONSUMERS` | `2026-08-09-PO-TASK-021-registry` |
| PO-09 | inventory_start=`2026-08-09` | F-021-X clock 開始 |
| PO-14 · #256 | visual sign-off dual SIGNED_OFF | `2026-08-08-PO-signoff-TASK-024` · **U13/close は未了** |
| PO-15 · TASK-022/S13 | 記録済 · #239 CLOSED | residual U6 |
| residual U0–U6 | closeout 縦ログ完了 | ledger |
| 受入バグ | Open **0** | §2 |
| OPS-7 | 旧 AWS IaC 退役 | 再開・apply しない |
| OPS-10 / 11 / 12 | 任意・repo 外・full seed 非 default | 残件化しない |
| TASK-009 | seed local | `exam_reference_ranges` COUNT **20** |
| ARCH-A1〜A3, A5〜A8 | domain / composition / lint | **done** · A4 は §4.2 |
| #212 / #235 / #260 | GitHub **CLOSED** · plan hub 復活なし | gh 2026-08-14 |
| agent §5 準備 | Issue 依頼投稿 · staging KEEP 推奨 · local PO-10 · ドラフト | [`SECTION5-STATUS.md`](reports/todo-walk-2026-08-14/SECTION5-STATUS.md) |
| staging-only disposition | 4 件 **KEEP** | [`staging-preflight-status.md`](reports/todo-walk-2026-08-14/staging-preflight-status.md) |
| staging draft PR | main→staging draft **#299**（未 merge） | https://github.com/MinoruSoga/AnimalEkarte/pull/299 |
| PO-10 local presence | secret present=0（local のみ） | [`po10-local-presence.md`](reports/todo-walk-2026-08-14/po10-local-presence.md) |
| Open Issue 1件ずつ | 16 件コメント · close **0** | [`github-issues-walk.md`](reports/todo-walk-2026-08-14/github-issues-walk.md) |
| GH→Linear 起票 | **BRT-37〜52**（16）Ready · parent BRT-4 | [`github-linear-map.md`](reports/todo-walk-2026-08-14/github-linear-map.md) |

### 4.2 裁定索引（HOLD / DEFER は §5 に出さない） {#ops}

| ID | Verdict | 回答 |
|----|---------|------|
| 再審 | なし | DEC-40〜68 / Fable pack 維持。UAT-human 裁定も OVERTURN 0（境界は人間・環境・外部起因） |
| PO-08 / TASK-021-C/D | **DEFER** | B→C→D。DROP しない。F-021-X: inventory_start=2026-08-09 → **2026-11-07** 無応答なら ACCEPT_RESIDUAL_RISK を再裁定（[`todo-po.md`](todo-po.md) 予約） |
| PO-10 / LINE-R05 | **DO_NEXT** | STG/PROD presence は未。local は §4.1 済。ゼロ前 DROP 禁止 → **§5 #8** |
| PO-11 / #201 | **DO_NOW** | Issue 依頼済。空欄記入待ち。TASK-033 禁止 → **§5 #1** |
| PO-12 / #249 | **DO_NEXT** | Issue 依頼済。承認前 unit 禁止 → **§5 #4** |
| PO-13 / #211 | **DO_NEXT** | Issue 依頼済。apply HOLD → **§5 #5** |
| PO-16 / #261 | **HOLD** | #201 opaque ref と runtime 5 項目が揃うまで |
| #254 close | **HOLD**（TIGHTEN） | local FAIL 0 単独不可。**UAT-H1〜H7**（H4〜H7 は disposition 可）+ build SHA `1386e1db0` + 別 USER sign-off — 正本 [`todo-po.md`](todo-po.md) §2 · [Fable](reports/fable-po-confirm-answer-2026-08-14-uat-human.md) §E |
| #256 close | **CLOSE_RECOMMEND** | U13 + 発効日 + opaque ref + 別承認後のみ → **§5 #3** |
| staging ← main merge | **DO_NEXT** | preflight 残り green 後に merge-commit PR のみ。disposition KEEP は §4.1 |
| 実 LINE UAT · OPS-4 · OPS-5 | **DO_NEXT** | current main の STG health 後。本文 r2 §E-8 |
| PO-17 · OPS-13 | **DO_NEXT** | named env 非破壊 migrate。agent migrate / reset 禁止 |
| PO-18 / #89·#97 · OPS-1 | **DO_NEXT** | rotation → **§5 #7** |
| PO-19 / #253 | **HOLD** | PROD 未構築 |
| PO-20 / #257 | **HOLD** | Go-live 日付を置かない |
| #250 | **HOLD** | Issue 催促済 · producer 回答待ち → **§5 #9** |
| #259 | **HOLD** | Issue 催促済 · enable 待ち · gate OFF → **§5 #10** |
| #284 | **DEFER** | DEFER_PHASE2 |
| OPS-3 | **DEFER** | STG health 後 0-rule 件数。SQL 下記 |
| OPS-14 | **DEFER** | staging PR 後 remote CI |
| #252 | **DEFER** | staging 準備後 preview |
| OPS-6 / 8 / 15 / 16 / 17 | **HOLD** | PROD 未構築または domain 未入力 |
| OPS-2 | 常設 | local fresh のみ → **§1** |
| OPS-9 | **DEFER** | 非 blocking 目視 |
| OPS-18 | **DEFER** | Sentry free-only |
| #249 外部 import | **DEFER** | phase2 |
| #255 | **UNANSWERABLE** | roster は repo 外 |
| ARCH-A4 | **DEFER** | 痛み駆動 → **§1** |
| TASK-004 / 005 | 都度 | land 時のみ → **§1** |
| TASK-033 | **HOLD** | #201 後 |
| TASK-374-apply | **HOLD** | #211 両行後 |
| POST-PULL | **KEEP_OPEN** | USER `make migrate` → **§1** |
| DR-CLINICAL 全行 | **UNANSWERABLE** | 依頼は §5 |
| DR-DELIVERY 全行 | **UNANSWERABLE** | `DELIVERY_PACKAGE.md` |
| DR-PRIVACY-256 一行 | **UNANSWERABLE** | U13 は USER |
| VACCINE-SPECIES 値 | **HOLD** | §5 #4 に添付 |

**OPS-3**（STG health 後 · read-only · 件数のみ）:

```sql
SELECT pg.id, pg.clinic_id, pg.name, COUNT(DISTINCT sp.staff_id) AS assigned_staff
FROM permission_groups pg
LEFT JOIN permission_group_rules r ON r.group_id = pg.id AND r.deleted_at IS NULL
LEFT JOIN staff_permission_groups sp ON sp.group_id = pg.id
WHERE pg.deleted_at IS NULL
GROUP BY pg.id, pg.clinic_id, pg.name
HAVING COUNT(r.id) = 0
ORDER BY pg.clinic_id, pg.id;
```

### 4.3 DEC アンカー

<a id="dec-40"></a><a id="dec-41"></a><a id="dec-42"></a><a id="dec-43"></a><a id="dec-44"></a><a id="dec-45"></a><a id="dec-46"></a><a id="dec-47"></a><a id="dec-48"></a><a id="dec-49"></a><a id="dec-50"></a><a id="dec-51"></a><a id="dec-52"></a><a id="dec-53"></a><a id="dec-54"></a><a id="dec-55"></a><a id="dec-56"></a><a id="dec-57"></a><a id="dec-58"></a><a id="dec-59"></a><a id="dec-60"></a><a id="dec-61"></a><a id="dec-62"></a><a id="dec-63"></a><a id="dec-64"></a><a id="dec-65"></a><a id="dec-66"></a><a id="dec-67"></a><a id="dec-68"></a>

索引は上表。全文は git 履歴の旧 `q&a.html`。

---

## 5. 着手可能な実行（未完のみ） {#today} {#actions} {#forms} {#p0} {#8}

完了分は **§4.1**（依頼投稿・CLOSED 確認・local PO-10・staging KEEP 含む）。  
本文: [r2 §E](reports/gpt-5.6sol-po-qa-answer-2026-08-14-r2.md) · 証跡一覧: [`SECTION5-STATUS.md`](reports/todo-walk-2026-08-14/SECTION5-STATUS.md)  
**agent 製品 unit: NONE。** 値は人が埋める。

先頭 3: **#1 #201 空欄** · **#2 preflight 残り** · **#3 U13**。

| # | ID | 次の一手 | 実行者 | 参照 | 空欄 |
|---|----|----------|--------|------|------|
| 1 | PO-11 / #201 | bundle 空欄を埋める | 臨床 | [Issue](https://github.com/MinoruSoga/AnimalEkarte/issues/201#issuecomment-5290074638) | 対象·policy·単位·出典·approver·発効日·opaque ref |
| 2 | staging preflight 残り | draft **#299** 作成済。checksum·backup owner·PR CI green のあと merge-commit | リリース | [PR #299](https://github.com/MinoruSoga/AnimalEkarte/pull/299) · [status](reports/todo-walk-2026-08-14/staging-preflight-status.md) | checksum·ownership·backup/rollback owner·CI green 確認·role·時刻·opaque ref |
| 3 | #256 U13 | 完了/未完を明示。完了時のみ close | 納品 | [Issue](https://github.com/MinoruSoga/AnimalEkarte/issues/256#issuecomment-5290075164) | U13_status·発効日·close 承認·opaque ref |
| 4 | PO-12 / #249 | range·vaccine 空欄 | 臨床 | [Issue](https://github.com/MinoruSoga/AnimalEkarte/issues/249#issuecomment-5290074797) | range·vaccine 全欄 |
| 5 | PO-13 / #211 | clinical / OPS 空欄 | 臨床→OPS | [Issue](https://github.com/MinoruSoga/AnimalEkarte/issues/211#issuecomment-5290075039) | clinical·OPS 全欄 |
| 6 | #258 U1–U8/10/11 | DELIVERY_PACKAGE へ記入 | 契約 | [draft](reports/todo-walk-2026-08-14/drafts/E-11-258.md) | A/B·各 U·opaque ref |
| 7 | PO-18 / OPS-1 | 4 系統 rotation | セキュリティ | r2 §E-12 | enum·owner·opaque ref |
| 8 | PO-10 STG/PROD | 承認 window で presence 件数のみ | DB 運用 | r2 §E-14 · local は §4.1 | env·件数·role·opaque ref |
| 9 | #250 | producer の complete bundle 回答 | 移行元 | [Issue](https://github.com/MinoruSoga/AnimalEkarte/issues/250#issuecomment-5290075276) | bundle·complete·authority |
| 10 | #259 | 先方 enable 回答。gate OFF 維持 | 外部連携 | [Issue](https://github.com/MinoruSoga/AnimalEkarte/issues/259#issuecomment-5290075415) | enable·gate·rollback ref |
| 11 | PO-008 | §5.1 を承認または修正 | クライアント仕様 | §5.1 | decision·修正·承認·発効日 |

**やらない:** merge（#2 green 前）· DROP · 値発明 · #254 close · TASK-033 先行

### 5.1 PO-008（クライアント確認） {#po008}

| 項目 | 現行 | 確認 |
|------|------|------|
| annual_visit_count | 直近 365 日 rolling | 承認または修正。他指標と統一しない |
| annual_amount | From/To → Year → preset → 全期間 | 承認または修正。visit と統一しない |
| CSV | 顧客集計 API に CSV なし | default 追加しない |
| last_visit / dormant | 別ロジック | 統一しない |
| L-step write | default-off · 実 2xx だけ成功 | enable は #259 後 |
| L-step cleanup | 本体削除は止めない。失敗通知必須 | silent success 禁止 |

以上。
