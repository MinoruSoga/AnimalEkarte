# Claude Fable への依頼 — PO 確認（独立第二意見）

日付: 2026-08-14  
依頼者: 実 PO（USER）  
対象: Claude Fable  
成果物: **残件すべての確認回答**。コード変更・GitHub 操作・migrate・secret 設定・外部実行はしない。

あなたは Animal Ekarte の **PO 裁定官** です。2026-08-06 に採択した [Fable pack](docs/work/decisions/fable-po-recommendation.md) の起草者として、いまの残件を独立に確認し、USER がそのまま送る・止める・採択できる答えを完成させてください。

GPT-5.6 Sol の第 2 ラウンド回答は USER がいったん正本に取り込み済みです。今回はそれを黙認せず、**行ごとに RATIFY / TIGHTEN / OVERTURN** してください。覆すときだけ新カード案を書きます。

---

## 1. 必読（この順）

1. [`todo-po.md`](todo-po.md) — **いまの PO 正本**（旧 `q&a.html` は削除済み。DEC 全文は git 履歴 `ab68c61f5`）
2. [`todo.md`](todo.md) §1〜3・§7 — 技術 SoT。agent unit は **NONE**
3. [`reports/gpt-5.6sol-po-qa-answer-2026-08-14-r2.md`](reports/gpt-5.6sol-po-qa-answer-2026-08-14-r2.md) — Sol r2（82 行・完成物 14 本）
4. [`docs/product-philosophy.md`](docs/product-philosophy.md) — ①疑う → ②削除 → ③簡素化 → ④サイクル → ⑤自動化。逆行禁止
5. [`docs/work/decisions/fable-po-recommendation.md`](docs/work/decisions/fable-po-recommendation.md) — あなた自身の採択済み pack
6. [`reports/uat-2026-08-14/FINAL.md`](reports/uat-2026-08-14/FINAL.md) — PASS 96 · PARTIAL 4 · BLOCKED 5 · FAIL 0
7. 必要なら [`docs/delivery/DELIVERY_PACKAGE.md`](docs/delivery/DELIVERY_PACKAGE.md) U1〜U12、[`phase2.html`](phase2.html)、[`docs/ops/infra/staging/runbook.md`](docs/ops/infra/staging/runbook.md)

読めなかったファイルは §A に列挙する。GitHub live state は操作禁止のため `todo-po.md` / `todo.md` §7.1 を正とする。

---

## 2. 動かない前提

| 項目 | 値 |
|------|-----|
| 環境 | PROD **未構築** · STG **ほぼ未使用** · local `make up` |
| 製品 unit | **NONE** |
| 臨床安全 | SPECIFICATION 2.1 が philosophy より優先 |
| DEC-40〜68 | 再審しない。覆すときだけ新カード |
| 破壊削除 | TASK-021 B/C/D · LINE-R05 DROP は gate 前禁止。B の registry は済 |
| #201 値 | #261 へ複製しない |
| history | no-rewrite |
| 発明禁止 | 薬用量・基準値・ワクチン適合性・契約金額・credential・実氏名/email/clinic ID・Go-live 日付 |

値セルへの正しい答えは次のいずれか。発明しない。

- `現行継続を推奨`（20% を医学的正本化しない）
- `クライアント所有（A）を推奨`
- `未承認のまま「未判定」を維持`
- `入力先は依頼文。値は臨床/契約責任者が埋める`

「値が無いので答えられない」で終わらせない。PO 判断セルは埋める。

---

## 3. 確認対象（欠けたら未完）

`todo-po.md` の残件をすべて。Sol と同じ Verdict でよいが、**各行に Fable の判定を付ける**。

### 3.1 今日・対応

1. PO-11 / #201 依頼  
2. staging preflight  
3. #256 U13  
4. staging ← main 実 merge  
5. 実 LINE / OPS-4 / OPS-5  
6. PO-10 / LINE-R05  
7. PO-12 / #249  
8. PO-13 / #211  
9. PO-16 / #261  
10. #254 close  
11. #256 close  
12. PO-17 / OPS-2 / OPS-13  
13. PO-18 / OPS-1 / #89·#97  
14. PO-19 / #253  
15. PO-20 / #257  
16. #250 催促  
17. #259 催促  
18. #284  
19. #252 preview  
20. #258 A/B + U1〜U12（各 U）  
21. PO-008 の 6 行（visit / amount / CSV / last-visit / L-step 通常 / cleanup）  
22. OPS-3 · 6 · 7 · 8 · 9 · 10 · 11 · 12 · 14 · 15 · 16 · 17 · 18  
23. #98 / #99 / #212 / #235 / #260 close  
24. #255 / #249 外部 import / VACCINE-SPECIES / TASK-033 / TASK-021 C/D / POST-PULL  

同じ Issue は **実行行と close 行を分ける**。

### 3.2 完成物

Sol r2 §E-1〜14 を読んで、各本について:

- **KEEP** — そのまま USER が送ってよい  
- **REVISE** — 修正本文を全文書く  
- **DROP** — 今は送らない理由  

「今は書かない」禁止。KEEP でも 1 行で理由を書く。

---

## 4. 判断順

① 要件を疑う → ② 削除 → ③ 簡素化 → ④ サイクル短縮 → ⑤ 自動化。  
確認ダイアログで安全を買わない。PROD 未構築・STG ほぼ未使用を完了扱いしない。  
DO_NOW は最大 3。Sol の今日 3 件（#201 / staging preflight / #256 U13）を維持するか、差し替えるかを明示する。

---

## 5. 出力形式（この見出しだけ。途中省略禁止）

日本語。

### A. 読んだもの / 読めなかったもの

### B. 再審

覆す DEC が無ければ「再審なし。DEC-40〜68 と Fable pack を維持」1 行。

### C. Executive

- 今日 Yes すべきトップ 3  
- 絶対に今 Yes すべきでないトップ 3  
- Sol r2 との関係: `RATIFY 何件 / TIGHTEN 何件 / OVERTURN 何件`

### D. 全件確認表

列: `ID` | `Sol` | `Fable` | `関係` | `POの答え（1文）` | `次の人` | `次の一手` | `空欄に残すセル`

`関係` は `RATIFY` / `TIGHTEN` / `OVERTURN` のみ。全残件を欠番なし。

### E. 完成物

§E-1〜14 について KEEP / REVISE / DROP。REVISE は修正本文を全文。

### F. 今日の最小セット（最大 3）

各: Verdict / なぜ今か / 手順 / 完了条件 / やらないこと

### G. agent unit

なければ **NONE**。増やすなら 1 つだけ・gate 明示。

### H. 発明しなかったもの

1 段落。

### I. カバレッジ

次を YES/NO。NO が 1 つでも未完。

- [ ] todo-po の今日・対応・OPS・フォーム・Issue を個別に判定した  
- [ ] U1〜U12 を個別に判定した  
- [ ] PO-008 の 6 行を個別に判定した  
- [ ] Sol §E-1〜14 を KEEP/REVISE/DROP した  
- [ ] 臨床数値・契約金額・secret・実 identity・Go-live 日付を発明していない  
- [ ] 製品 unit を無断で増やしていない  

---

## 6. 品質

- Fable pack を自分の権威で上書きしない。新しい証拠があるときだけ OVERTURN
- 断定するときは読んだファイル名を付ける
- 同じ値を 2 つの正本に書かない（#201 は bundle 1 行、#258 は DELIVERY_PACKAGE、#256 は DR-PRIVACY 1 行）
- 実装計画・新画面・リファクタは出さない
- 実行しない

以上です。§5 の見出し順で回答してください。
