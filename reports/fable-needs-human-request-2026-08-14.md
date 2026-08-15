# Claude Fable への依頼 — Needs Human チケット代行回答（電カル / ノア）

| 項目 | 値 |
|------|-----|
| **日付** | 2026-08-14 |
| **依頼者** | 実 PO（USER） |
| **対象** | Claude Fable（PO 裁定官 · 独立第二意見） |
| **案件** | ノア動物病院電子カルテ（AnimalEkarte）· Linear hub **BRT-4** |
| **成果物** | 下記 **29 件すべての Needs Human** に対する代行回答パック |
| **禁止** | コード変更 · GitHub/Linear の Done · merge · migrate · secret 設定 · 外部 API 実行 · 値の発明 |

あなたは Animal Ekarte の **PO 裁定官** です。過去に採択した Fable pack・Sol r2・UAT 人間レーン設計を前提に、**いま Linear が Needs Human になっているチケットを、USER の代わりに「答えられる範囲」まで埋めてください。**

USER はこの回答を Linear コメント / 送付文 / 次アクションに転記します。

---

## 0. いまの事実（agent 実測 · 2026-08-14）

| 項目 | 値 |
|------|-----|
| main tip（worktree） | `45c8c8155` |
| 製品 Open バグ / agent 実装 unit | **NONE** |
| 台帳 | root `todo.md` / `todo-po.md` は **ポインタのみ**（本文は Linear へ移行済） |
| PR #299 | OPEN · **draft** · https://github.com/MinoruSoga/AnimalEkarte/pull/299 |
| #299 CI | Detect/Gitleaks SUCCESS · **Backend / Frontend / Codegen FAILURE**（green 未） |
| local `audit_logs` COUNT | **373**（read-only · 中身非開示） |
| local PO-10 presence | `line_channel_secret` present **0**（値非保存） |
| staging...main（過去読取） | staging-only 4 件は KEEP 推奨（patch は main と同一） |
| 調査 handoff | `CorpVault/55_Handoff/BRT-37_ops.md` … `BRT-67_ops.md`（欠番あり） |

マップ:

- docs→Linear: `reports/todo-walk-2026-08-14/todo-docs-linear-map.md`
- GH→Linear: `reports/todo-walk-2026-08-14/github-linear-map.md`
- 一括 walk: `reports/todo-walk-2026-08-14/linear-all-walk.md`

---

## 1. 必読（この順）

1. 本ファイル（対象一覧と出力形式）
2. [`docs/work/decisions/fable-po-recommendation.md`](../docs/work/decisions/fable-po-recommendation.md) — 採択済み pack
3. [`reports/gpt-5.6sol-po-qa-answer-2026-08-14-r2.md`](gpt-5.6sol-po-qa-answer-2026-08-14-r2.md) — Sol r2
4. [`reports/fable-po-confirm-answer-2026-08-14.md`](fable-po-confirm-answer-2026-08-14.md) · [uat-human](fable-po-confirm-answer-2026-08-14-uat-human.md) · [exec-session](fable-po-confirm-answer-2026-08-14-exec-session.md)
5. [`docs/product-philosophy.md`](../docs/product-philosophy.md)
6. [`reports/uat-2026-08-14/FINAL.md`](uat-2026-08-14/FINAL.md) — FAIL 0
7. 必要時: [`docs/delivery/DELIVERY_PACKAGE.md`](../docs/delivery/DELIVERY_PACKAGE.md) · handoff `BRT-XX_ops.md` · GH Issue 本文（値は読んでも **転記・発明しない**）

読めなかったファイルは §A に列挙。GitHub/STG/PROD への **破壊的操作は禁止**（読取のみ可）。

---

## 2. 動かない前提

| 項目 | 値 |
|------|-----|
| PROD | **未構築** |
| STG | ほぼ未使用 · #299 green 前 merge 禁止 |
| DEC-40〜68 | 再審しない。覆すときだけ新カード + OVERTURN 理由 |
| 製品 unit | 無断で増やさない（NONE 維持が既定） |
| TASK-033 | #201 bundle 全列承認まで **骨格先行禁止** |
| 破壊削除 | TASK-021 C/D · LINE-R05 DROP は inventory ゼロ前 **禁止** |
| #201 値 | #261 へ複製しない |
| history | no-rewrite · filter-repo 禁止 |
| Done | 人間のみ（Fable も Linear Done にしない） |

### 発明禁止（値セル）

薬用量 · 基準 range · ワクチン適合性 · 契約金額 · credential · 実氏名/email/clinic 実 ID · Go-live **日付の新規設定** · 実 token · 承認資料本文

値セルの正しい答えは次のいずれか:

- `現行継続を推奨`（例: 20% を医学的正本化しない）
- `未承認のまま「未判定」を維持`
- `クライアント所有（A）を推奨`
- `入力先は依頼文。値は臨床/契約/セキュリティ責任者が埋める`
- `NEEDS_CLINICAL` / `NEEDS_USER_OPS` / `WAIT_EXTERNAL` / `HOLD` + **解除条件 1 行**

**「値が無いので答えられない」で終わらせない。**  
PO 判断セル（優先度・順序・送る/止める・close してよいか・disposition）は埋める。

---

## 3. 対象チケット（29 · 欠けたら未完）

すべて Linear **Needs Human** · parent 系 BRT-4 · 電カルのみ（谷口 BRT-53/54 は対象外）。

### 3.1 臨床・契約・安全（値は原則埋めない · プロセスは埋める）

| # | Linear | 対応 GH/ID | あなたに求める代行 |
|---|--------|------------|-------------------|
| 1 | [BRT-39](https://linear.app/baritechllc/issue/BRT-39) | #201 PO-11 | bundle **列リスト再掲** · 催促文完成 · TASK-033 禁止再確認 · close 条件 |
| 2 | [BRT-41](https://linear.app/baritechllc/issue/BRT-41) | #249 | range 承認の **行テンプレ** · 未判定維持の是非 · unit 起票禁止維持？ |
| 3 | [BRT-40](https://linear.app/baritechllc/issue/BRT-40) | #211 | DR-CLINICAL / DR-OPS 分離テンプレ · local 実 row 禁止維持？ |
| 4 | [BRT-51](https://linear.app/baritechllc/issue/BRT-51) | #261 | #201 参照のみ close 方針の RATIFY · 複製禁止 · 残 5 項目 enum 案 |
| 5 | [BRT-49](https://linear.app/baritechllc/issue/BRT-49) | #258 | U1–U12 各行: 誰が埋める / A 推奨 / U9·U12 は #253 後か |
| 6 | [BRT-47](https://linear.app/baritechllc/issue/BRT-47) | #256 U13 | **COMPLETED / 未完の判定枠**（事実が無いなら「USER が 1 語選ぶ」選択肢）· close 条件 |
| 7 | [BRT-56](https://linear.app/baritechllc/issue/BRT-56) | PO-008 | 集計 6 点それぞれ **承認推奨 or 修正をクライアントに聞く文**（値の断定禁止なら質問文） |

### 3.2 セキュリティ・インフラ・外部

| # | Linear | 対応 | 代行 |
|---|--------|------|------|
| 8 | [BRT-37](https://linear.app/baritechllc/issue/BRT-37) | #89 | 4 系統ローテの **実行順序・完了証跡 1 行フォーマット** · agent 禁止再確認 |
| 9 | [BRT-38](https://linear.app/baritechllc/issue/BRT-38) | #97 | #89 後マスク · session 無効化チェックリスト |
| 10 | [BRT-57](https://linear.app/baritechllc/issue/BRT-57) | PO-10 | STG/PROD 件数取得の **承認 window 文** · DROP 禁止 |
| 11 | [BRT-44](https://linear.app/baritechllc/issue/BRT-44) | #253 | PROD gate 残チェックリスト · Environment 必須の RATIFY |
| 12 | [BRT-55](https://linear.app/baritechllc/issue/BRT-55) | PR #299 | **CI FAILURE 中の merge 禁止** RATIFY · green 後手順 · draft 解除条件 |
| 13 | [BRT-67](https://linear.app/baritechllc/issue/BRT-67) | OPS-13 | named env migrate の USER 専用手順 1 枚 · agent 禁止 |
| 14 | [BRT-42](https://linear.app/baritechllc/issue/BRT-42) | #250 | COMPLETE bundle 最小セット · **再催促してよいか**（本日禁止方針の更新） |
| 15 | [BRT-50](https://linear.app/baritechllc/issue/BRT-50) | #259 | enable 確認の非機密質問文 · gate OFF 維持 · 再催促可否 |
| 16 | [BRT-46](https://linear.app/baritechllc/issue/BRT-46) | #255 | roster 未着時の disposition · 発行順序 |
| 17 | [BRT-43](https://linear.app/baritechllc/issue/BRT-43) | #252 | 城東同値投入の **誰が・いつ・証跡** · #257 gate 入り RATIFY |

### 3.3 UAT 人間レーン · 送付 · Go-live · close

| # | Linear | 対応 | 代行 |
|---|--------|------|------|
| 18 | [BRT-65](https://linear.app/baritechllc/issue/BRT-65) | P1 | 90 分セッション冒頭手順の確定文（pull/migrate/reset） |
| 19 | [BRT-66](https://linear.app/baritechllc/issue/BRT-66) | M1–M5 | **送付 5 通の完成文**（宛先ロール・件名・本文・禁止事項）· 順序 |
| 20 | [BRT-58](https://linear.app/baritechllc/issue/BRT-58) | H1 | #299 後の実施条件 · PASS 定義 · 値非記録 |
| 21 | [BRT-59](https://linear.app/baritechllc/issue/BRT-59) | H2 | H1 同一 window · PASS 定義 |
| 22 | [BRT-60](https://linear.app/baritechllc/issue/BRT-60) | H3 | COUNT=373 をどう扱うか · PASS 最小観測 |
| 23 | [BRT-61](https://linear.app/baritechllc/issue/BRT-61) | H4 | 実施 or disposition テンプレ |
| 24 | [BRT-62](https://linear.app/baritechllc/issue/BRT-62) | H5 | PASS 定義（新規+永続） |
| 25 | [BRT-63](https://linear.app/baritechllc/issue/BRT-63) | H6 | 持ち越し/disp 条件 |
| 26 | [BRT-64](https://linear.app/baritechllc/issue/BRT-64) | H7 | spot-check 4 件の具体 ID |
| 27 | [BRT-45](https://linear.app/baritechllc/issue/BRT-45) | #254 | close **してよいか** · 未充足条件表 · local FAIL0 単独不可 RATIFY |
| 28 | [BRT-48](https://linear.app/baritechllc/issue/BRT-48) | #257 | **新 window を今決めない** RATIFY · gate リスト最終版 |
| 29 | [BRT-52](https://linear.app/baritechllc/issue/BRT-52) | #284 | phase2 維持か · close 条件（実機 3 台） |

---

## 4. 出力形式（この形以外は未完）

### A. 読了・環境

読めなかったパス · 前提の訂正（あれば）

### B. トップ 3（今日 USER が Yes すべき）

各 1 文 + 対象 Linear ID

### C. 絶対 No トップ 3

各 1 文（merge / DROP / 値発明 / TASK-033 先行 / 早期 close 等）

### D. チケット別回答表（29 行必須）

| Linear | Verdict | 1 文の答え（USER 代行） | 次の一手（誰が） | 送付/コメント文（必要なら） | 発明した値？ |
|--------|---------|------------------------|------------------|------------------------------|--------------|
| BRT-xx | RATIFY / TIGHTEN / HOLD / NEEDS_* / WAIT_EXTERNAL | … | … | … or `-` | **必ず No** |

Verdict 定義:

- **RATIFY** — 既存方針を追認し、そのまま実行文を確定
- **TIGHTEN** — 方針は同じだが条件を絞る
- **HOLD** — まだ止める · 解除条件を書く
- **NEEDS_CLINICAL / NEEDS_USER_OPS / WAIT_EXTERNAL** — 代行できない部分を明示し、**依頼文は完成**させる
- **OVERTURN** — 既存 DEC/Fable を覆す（理由・新カード必須 · 乱用禁止）

### E. 送付文パック（コピー用）

最低限:

1. M1 #201 臨床（BRT-39 / BRT-66）
2. M2 U13 1 語依頼（BRT-47）
3. M3 #258（BRT-49）
4. M4 #299 preflight / CI（BRT-55）
5. M5 PO-008（BRT-56）

各: 件名 · 本文 · 空欄マーカー `[ ]` · **値・秘密・日付の断定なし**

### F. #299 / #254 / #256 close 判定カード

| ID | 今 close/merge してよいか | Yes の最小条件 | No のとき残す 1 行 |
|----|---------------------------|----------------|-------------------|
| #299 merge | | | |
| #254 close | | | |
| #256 close | | | |

### G. 90 分セッション（ケース B）更新

P1 → M1–M5 → H3/H5/H7 の順序を **最終確定**（変更点だけ明示）

### H. Linear コメント用 1 行（29 件）

`BRT-xx: <Verdict> — <20 字以内>`

USER が Needs Human チケットに貼る用。

---

## 5. 品質ゲート（自己チェック）

回答前にすべて Yes:

- [ ] 29 行すべて埋めた（欠番なし）
- [ ] 薬用量・range・credential・Go-live 日付を **発明していない**
- [ ] #299 green 前 merge を Yes にしていない
- [ ] #254 を local FAIL0 だけで close していない
- [ ] TASK-033 骨格先行を Yes にしていない
- [ ] Done 遷移を指示していない（Needs Human のまま次担当を書く）
- [ ] 谷口 BRT-53/54 に触れていない
- [ ] 日本語 · USER がそのまま使える文

---

## 6. 成果物の置き場（USER）

推奨ファイル名:

`reports/fable-needs-human-answer-2026-08-14.md`

Linear 転記は USER または secretary が行う（Fable は API 操作不要）。

---

**開始文（Fable へコピーする場合）:**

> あなたは Animal Ekarte の PO 裁定官 Claude Fable です。  
> `reports/fable-needs-human-request-2026-08-14.md` に従い、Needs Human 29 件を代行回答してください。  
> 発明禁止。コード変更禁止。出力形式 §4 を満たすまで未完です。
