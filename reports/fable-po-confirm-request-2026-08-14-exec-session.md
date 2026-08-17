# Claude Fable への依頼 — 実施セッション設計 / 送付文 / disposition（第2弾）

日付: 2026-08-14  
依頼者: 実 PO（USER）  
対象: **Claude Fable**  
成果物: **USER が今日そのまま実行・送付できる完成物**。  
コード変更・GitHub 操作・merge・migrate・secret 設定・外部実行は **しない**。

---

## 0. 本依頼の位置づけ（前回との差分）

| ラウンド | ファイル | 結果 |
|----------|----------|------|
| 前回（裁定） | [`fable-po-confirm-answer-2026-08-14-uat-human.md`](fable-po-confirm-answer-2026-08-14-uat-human.md) | RATIFY 21 · TIGHTEN 4 · OVERTURN 0 · **Verdict 確定** |
| 台帳取込 | [`todo-po.md`](../todo-po.md) · [`todo.md`](../todo.md) | H1–H7 · #254 ゲート · 送付/停止リスト反映済み |
| **今回** | **本依頼** | Verdict の **再審はしない**。**実施の順序・文面・disposition テンプレ・1 セッション計画**を完成させる |

**禁止:** 既に RATIFY した HOLD/DEFER を覆す再議論（#254 close 緩和、DROP 解禁、local FAIL 0 = 納品完了、など）。  
**今回の仕事:** 「誰が・何分で・何を書いて・どの順で手を動かすか」を USER が迷わず実行できる粒度まで落とす。

あなたは Animal Ekarte の **PO 裁定官**かつ **実施セッション設計者**です。採択済み pack と前回 UAT-human 回答に拘束される。

---

## 1. 必読（この順）

1. **前回回答** [`fable-po-confirm-answer-2026-08-14-uat-human.md`](fable-po-confirm-answer-2026-08-14-uat-human.md) 全文 — 特に C / E / F / G  
2. [`todo-po.md`](../todo-po.md) 全文 — 現在の open 状態の正本  
3. [`todo.md`](../todo.md) §4.2 · §5 · §5.1 — 空欄テンプレと禁止  
4. [`reports/uat-2026-08-14/FINAL.md`](uat-2026-08-14/FINAL.md) — FAIL 0 · build `1386e1db0`  
5. [`docs/product-philosophy.md`](../docs/product-philosophy.md)  
6. 必要なら Sol r2 §E の依頼文骨格（E-1 / E-5 / E-6 / E-7 / E-8 / E-11 / E-12 / E-14）、[`docs/ops/infra/staging/runbook.md`](../docs/ops/infra/staging/runbook.md)（checksum · PS109）

読めなかったものは §A に列挙。GitHub live / STG / PROD 接続禁止。

---

## 2. 動かない前提（前回から不変）

| 項目 | 値 |
|------|-----|
| PROD | 未構築 |
| STG | ほぼ未使用 · main から大幅遅れ · **#299 green 前 merge 禁止** |
| local UAT | FAIL 0 · PASS 1352 · 証跡 build **`1386e1db0`** · close 単独証拠にならない |
| origin/main | **`697d5c597`** が `001_init.sql` を編集済み（UAT 証跡に未含有） |
| agent 製品 unit | NONE |
| DEC-40〜68 | 再審なし |
| 発明禁止 | 臨床値・契約金額・credential・Go-live 日付・実 token・実氏名 |

---

## 3. 今回だけ答える対象

### 3.1 今日の時間配分（必須）

USER の今日の可用時間が次の 3 ケースのとき、それぞれ **分単位のセッション表**を出す。

| ケース | 想定時間 |
|--------|----------|
| A | **30 分**（送付と 1 語判断のみ） |
| B | **90 分**（送付 + local 人手レーンの一部） |
| C | **3 時間**（送付 + local H3–H7 を可能な限り閉じる） |

各ケースで:

1. **やる順**（時刻不要。1→2→3…）  
2. **やらない**（明示）  
3. **成果物**（何が残れば成功か）  
4. **#254 に近づいた距離**（H のどれが done / disposition 候補か）

### 3.2 pull / migrate の順序判断（必須）

現状 worktree は UAT 時 `1386e1db0` 系、`origin/main` は `697d5c597`（001 編集）。

次のどちらを **今日の H3–H7 前**に取るか、1 つ選べ。

| 案 | 内容 |
|----|------|
| **P0** | 今日は **pull しない**。H3–H7 を現 build で実施し SHA を `1386e1db0`（または実測 HEAD）で記録。pull は別日 |
| **P1** | **先に pull + make migrate**（mismatch なら OPS-2 fresh）のうえ H3–H7。実施 SHA を `697d5c597` 系で記録 |
| **P2** | H5 だけ現 build、締め(H4)だけ pull 後、など分割 |

選んだ案の **1 文理由** + **失敗時（migrate 失敗 / データ消えた）の退避** 1 行。

### 3.3 今日送付する文面の完成物（必須 · 最大 5 通）

前回 §F「送ってよい」5 通について、**USER がコピペできる完成文**を書く。  
値は発明しない。空欄は `[  ]` で明示。

| # | 宛先レーン | 目的 |
|---|------------|------|
| M1 | 臨床（#201） | bundle 記入催促 · 返答期限プレースホルダ |
| M2 | 納品（#256 U13） | 完了/未完の **1 語**回答依頼 |
| M3 | 契約（#258） | REVISE 版 E-11 送付の案内（本文は「前回 Fable REVISE を添付」でよい · 全文再掲は任意） |
| M4 | リリース + STG DB | preflight 残セル 3 つ + PlanetScale 109 REASSIGN 依頼 |
| M5 | クライアント（PO-008） | §5.1 の 6 行を承認 or 修正 |

各メッセージに:

- 件名（またはチャット 1 行タイトル）  
- 本文（日本語 · 短く）  
- **禁止フレーズ**（例: 納品完了、local UAT で close 可、など）を 1 行  

送らない方がよい通があれば **DROP** と理由（その場合 5 通未満でよい）。

### 3.4 disposition テンプレ（必須）

H4 / H5 / H6 / H7 について、**理由付き disposition** を書くときの固定フォーマットを定義する。

最低限のフィールド:

```text
id: UAT-H?
result: PASS | FAIL_BUG | ACCEPT_DISPOSITION
build_sha: …
operator: …
date: …
evidence: （パス or opaque ref · 秘密なし）
if ACCEPT_DISPOSITION:
  reason: （なぜ実行できなかったか · 1〜3 文）
  residual_risk: （何が未証明のままか）
  not_a_product_pass: true
```

さらに:

- **ACCEPT_DISPOSITION が許される条件**（H4–H7 のみ · H1–H3 は不可を再確認）  
- **FAIL_BUG のとき** → `todo.md` §2 の BUG 起票 1 行テンプレ  
- **#254 E-6 一行**にどう畳むか（例: `H4=ACCEPT_DISPOSITION/…; H5=PASS; …`）

### 3.5 local 1 セッション手順書（必須 · ケース C 想定）

H3 → H5 → H7 → H6 → H4 の推奨順を確定し、各ステップ:

| フィールド | 内容 |
|------------|------|
| 前提 | ロール・画面 URL・fixture |
| 操作 | 3〜7 ステップ |
| 合格 | 何が見えたら PASS |
| 不合格 | BUG vs 再試行 vs disposition |
| 記録先 | `reports/…` のファイル名案（新規ディレクトリ可） |

S13（H6）の fixture が無いときの **最小 fixture 手順**（飼主 2・医院 2 の探し方。ID 直書き禁止 · 検索条件で）を 5 行以内。

### 3.6 まだ判断が割れるかもしれない分岐（任意だが推奨）

次に YES/NO だけ答える（長文不要）:

1. U13 が「未完」と明示された日、#256 を open のまま何を次に待つ？  
2. staging checksum が永久 mismatch のとき、merge を止めるか · runbook の disposition で進めるか（方針 1 文）  
3. H7 の 4 件のうち **1 件だけ FAIL_BUG** のとき、他 3 件 PASS で #254 local 側は「残り 1 BUG 修正待ち」と書いてよいか  

---

## 4. 出力形式（この見出しだけ）

日本語。途中省略禁止。

### A. 読んだもの / 読めなかったもの

### B. 再審

「再審なし。前回 UAT-human と DEC-40〜68 を維持」または、**実施設計に必要な最小 TIGHTEN のみ**（Verdict 覆しは禁止に近い。覆すなら OVERTURN 明示）。

### C. Executive（実施版）

- ケース B（90 分）を **既定プラン**として 8 行以内で要約  
- 今日の成功定義（チェックリスト 5 項目以内）  
- 今日の失敗定義（やってはいけない 5 項目以内）

### D. pull 方針（P0 / P1 / P2 のいずれか 1 つ）

### E. 時間別セッション表（A 30 分 · B 90 分 · C 3 時間）

表形式。

### F. 送付文完成物（M1–M5）

各通: 件名 · 本文 · 禁止フレーズ · KEEP/DROP

### G. disposition / BUG / E-6 テンプレ

### H. local 手順書（H3–H7）

### I. 分岐 YES/NO（§3.6）

### J. todo-po.md への追記提案（任意・短い）

「実施ログ置き場」「セッション結果の Status 更新ルール」など、**行追加だけで足りるもの**。

---

## 5. 成功条件

- USER が本回答だけ見て **90 分セッションを開始できる**  
- 送付文に **秘密・発明値・「納品完了」誤読**が無い  
- H1–H3 を disposition で逃げていない  
- 前回 OVERTURN 0 の境界を壊していない  

以上。
