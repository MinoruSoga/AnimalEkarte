# Claude Fable への依頼 — PO 確認 / 判断待ち（UAT 後 · 人間レーン）

日付: 2026-08-14  
依頼者: 実 PO（USER）  
対象: **Claude Fable**  
成果物: **判断待ちの全件について、USER が採択できる裁定回答**。  
コード変更・GitHub 操作・merge・migrate・secret 設定・外部実行・シナリオ md 編集は **しない**。

あなたは Animal Ekarte の **PO 裁定官**です。  
2026-08-06 採択の [Fable pack](docs/work/decisions/fable-po-recommendation.md) の起草者として、  
**local フル受入 UAT 後に残った「PO / 人間レーン」だけ**を独立に裁定してください。

前回の Fable 確認（[`fable-po-confirm-answer-2026-08-14.md`](fable-po-confirm-answer-2026-08-14.md)）と Sol r2 は USER が `todo.md` §4 に取り込み済みです。  
今回はそれを黙認せず、**行ごとに RATIFY / TIGHTEN / OVERTURN** してください。覆すときだけ新カード案を書きます。

---

## 0. 今回の新事実（必読）

| 項目 | 値 |
|------|-----|
| local フル UAT | [`reports/uat-2026-08-14/FINAL.md`](uat-2026-08-14/FINAL.md) |
| 結果 | 製品 **FAIL 0** · PASS **1352** · PARTIAL **26** · BLOCKED **7** · N/A **2** · 新規受入バグ **0** |
| 実行 | Chrome CDP `:9222` + Playwright `connectOverCDP`（DevTools MCP 同 endpoint） |
| シナリオ md | **未編集** |
| 人間レーン正本 | [`todo-po.md`](../todo-po.md) §1〜3（本依頼の主対象） |
| 製品バグ台帳 | `todo.md` §2 Open **なし** |

**動かない境界:** local FAIL 0 は **#254 close の単独証拠にならない**（前回 Fable も同旨）。PROD 未構築 · STG ほぼ未使用は完了扱いにしない。

---

## 1. 必読（この順）

1. [`todo-po.md`](../todo-po.md) — **PO / 人間実施の正本**（§1 UAT 人間レーン · §2 裁定 · §3 空欄索引）
2. [`todo.md`](../todo.md) §2・§3・§4.1・§4.2・§5 — 技術 SoT · 裁定索引 · 着手可能実行
3. [`reports/uat-2026-08-14/FINAL.md`](uat-2026-08-14/FINAL.md) — 今回 UAT 結論
4. [`docs/product-philosophy.md`](../docs/product-philosophy.md) — ①疑う → ②削除 → ③簡素化 → ④サイクル → ⑤自動化
5. [`docs/work/decisions/fable-po-recommendation.md`](../docs/work/decisions/fable-po-recommendation.md) — 採択済み Fable pack
6. 前回回答 [`fable-po-confirm-answer-2026-08-14.md`](fable-po-confirm-answer-2026-08-14.md) — 差分判定の基準
7. 必要なら Sol r2 [`gpt-5.6sol-po-qa-answer-2026-08-14-r2.md`](gpt-5.6sol-po-qa-answer-2026-08-14-r2.md)、[`docs/delivery/DELIVERY_PACKAGE.md`](../docs/delivery/DELIVERY_PACKAGE.md)、[`docs/ops/infra/staging/runbook.md`](../docs/ops/infra/staging/runbook.md)

読めなかったファイルは §A に列挙。GitHub live / STG / PROD への接続・操作は禁止。`todo-po.md` / `todo.md` を正とする。

---

## 2. 動かない前提

| 項目 | 値 |
|------|-----|
| 環境 | PROD **未構築** · STG **ほぼ未使用** · local `make up` で UAT 済み |
| 製品 unit（agent） | **NONE**（判断のみ） |
| 臨床安全 | SPECIFICATION 2.1 が philosophy より優先 |
| DEC-40〜68 | 再審しない。覆すときだけ新カード |
| 破壊削除 | TASK-021 C/D · LINE-R05 DROP は gate 前禁止 |
| 発明禁止 | 薬用量・基準値・契約金額・credential・実氏名/email/clinic ID・Go-live 日付・「local FAIL 0 = 納品完了」 |
| シナリオ | `docs/ops/testing/scenarios/*` に結果を書かない |

値セルへの正しい答え例（発明しない）:

- `現行継続を推奨`
- `人間レーン実施後に再判定`
- `STG health 後に DO_NEXT`
- `未承認のまま「未判定」を維持`
- `入力先は依頼文。値は臨床/契約責任者が埋める`
- `仕様受容（BUG にしない）`

---

## 3. 確認対象（欠けたら未完）

### 3.1 UAT 人間レーン（`todo-po.md` §1）— 各行に Verdict

各 ID について:

- **Verdict:** `DO_NOW` / `DO_NEXT` / `HOLD` / `DEFER` / `ACCEPT` / `DROP_FROM_CRITICAL_PATH`
- **依存:** 何が揃うまで待つか（1 行）
- **完了の定義:** 人が何を見たら done か（1 行）
- **#254 との関係:** close の必要条件に含めるか Y/N

| ID | 内容（要約） |
|----|----------------|
| UAT-H1 | S04 実 LINE プッシュ（STG · mock 禁止） |
| UAT-H2 | S12 実 LINE / LIFF token |
| UAT-H3 | S06 audit_logs DB 参照（読取のみ） |
| UAT-H4 | S09 締め fixture 属性 |
| UAT-H5 | V04 シフトテンプレート SidePanel 人手確認 |
| UAT-H6 | S13 2 医院・2 飼主 fixture で identity-links フル |
| UAT-S1 | S08 部分入金不可を仕様受容（BUG にしない）— 維持 or 仕様変更を開くか |

### 3.2 UAT 関連 PO 裁定（`todo-po.md` §2 + `todo.md` §4.2）

| ID | 現状メモ | 聞いてほしいこと |
|----|----------|------------------|
| **#254 close** | local UAT FAIL 0 あり · **HOLD** | close 条件を **UAT-H1〜H4 必須**のまま維持するか。緩和するなら最小セットを明示 |
| **#256 close** | visual sign-off 済 · U13 未了 · CLOSE_RECOMMEND | U13 未完のまま close を禁止する方針を RATIFY するか |
| **staging ← main (#299)** | draft · 未 merge · DO_NEXT | preflight 全 green 前の merge を引き続き禁止するか |
| **実 LINE UAT · OPS-4/5** | STG health 後 · DO_NEXT | UAT-H1/H2 と順序を「STG health → 実 LINE」で固定するか |
| **PO-06 記録** | done（close は #254） | 追加判断不要なら RATIFY 1 行 |

### 3.3 人が埋める空欄（`todo-po.md` §3 / `todo.md` §5）— 優先順位のみ

全 11 件の **中身の値は発明しない**。代わりに:

1. **今日 Yes すべきトップ 3**（前回は #201 / staging preflight / #256 U13 — 維持か差替えを明示）
2. **絶対に今 Yes すべきでないトップ 3**
3. 各 #1〜#11 に `DO_NOW` / `DO_NEXT` / `HOLD` / `WAIT_EXTERNAL` を 1 語 + 1 文

| # | ID |
|---|-----|
| 1 | PO-11 / #201 |
| 2 | staging preflight / #299 |
| 3 | #256 U13 |
| 4 | PO-12 / #249 |
| 5 | PO-13 / #211 |
| 6 | #258 DELIVERY |
| 7 | PO-18 / OPS-1 |
| 8 | PO-10 STG/PROD presence |
| 9 | #250 producer |
| 10 | #259 enable |
| 11 | PO-008 顧客集計指標 |

### 3.4 予約・その他

| ID | 内容 |
|----|------|
| F-021-X / 2026-11-07 | inventory 無応答時 ACCEPT_RESIDUAL_RISK 再裁定 — 日付・条件を RATIFY するか |
| PARTIAL 26 | 製品バグにしない方針でよいか。cleanup left を人間必須にするか / 無視してよいか |

---

## 4. 判断順

① 要件を疑う → ② 削除 → ③ 簡素化 → ④ サイクル短縮 → ⑤ 自動化。  
確認ダイアログで安全を買わない。  
**local UAT FAIL 0 を「納品完了」「#254 close」「STG 不要」に読み替えない。**

DO_NOW は最大 3。前回トップ 3 との差分を Executive で 1 段落。

---

## 5. 出力形式（この見出しだけ。途中省略禁止）

日本語。回答はリポジトリに保存する想定で、USER が `todo-po.md` / `todo.md` §4 に転記できる粒度で。

### A. 読んだもの / 読めなかったもの

### B. 再審

覆す DEC が無ければ「再審なし。DEC-40〜68 と Fable pack を維持」1 行。  
local UAT 新結果が境界を覆すかどうかも 1 文。

### C. Executive

- 今日 Yes すべきトップ 3  
- 絶対に今 Yes すべきでないトップ 3  
- 前回 Fable 回答との差分（維持 / 差替え）  
- 今回 UAT が変えたこと / 変えなかったこと（各 2 行以内）

### D. 全件確認表

§3.1 → §3.2 → §3.3 → §3.4 の順。各行:

```text
ID: …
前回/Sol/todo: …
Fable Verdict: …
関係: RATIFY | TIGHTEN | OVERTURN
POの答え（1文）: …
次の人: …
次の一手: …
完了条件 / 空欄に残すセル: …
#254 close 必要条件?: Y/N（§3.1 のみ必須）
```

### E. #254 close ゲート（専用）

close してよい最小条件を箇条書き。  
「local FAIL 0 のみ」は **不可** と明記するか、緩和案があるなら条件付きで書く。

### F. USER が今日送る / 止める

- 送ってよいもの（Issue 催促・空欄依頼など）最大 5  
- 止めるもの最大 5  

### G. todo-po.md 更新提案（任意・短文）

Status や順序を変えるなら、差し替え用の表行だけ。全文再掲は不要。

---

## 6. 成功条件

- §3 の ID がすべて §D に出現している  
- 発明した臨床値・日付・credential が **0**  
- OVERTURN があるときだけ新カード案がある  
- USER がそのまま「Yes / 止める / 誰に依頼」を選べる  

以上。
