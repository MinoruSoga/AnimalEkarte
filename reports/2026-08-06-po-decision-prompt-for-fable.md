# PO 決裁プロンプト（Fable 向け）— 残ゲートの推奨裁定

| 項目 | 値 |
|------|-----|
| **作成** | 2026-08-06 |
| **用途** | 下のプロンプトを **Fable 5** に貼る。USER は **Fable の推奨（Recommendation）に従って** 最終採否する |
| **前提** | Opus 代理 pack 済み · UNIT-021-A landed · ブラウザ IU は residual 除外 |
| **正本** | `STATUS.md` · `q&a.html` · `reports/2026-08-06-po-proxy-decision-pack.md` |
| **禁止** | 臨床数値の発明 · 秘密の記載 · migrate/DROP/credential の実行 |

---

## 使い方

1. **コピー用プロンプト** を Fable 5 に貼る  
2. 可能ならコンテキストに次を読ませる:  
   - `STATUS.md` §1–§2  
   - `reports/2026-08-06-po-proxy-decision-pack.md`（前段の代理裁定）  
   - `q&a.html`（DEC-48/60–67 · DR-CLINICAL）  
   - `reports/2026-07-31-task-021-phase2-slice2.md`  
   - `reports/2026-07-31-r05-single-sot-phase-b.md`  
3. Fable 出力の **「USER 採否チェックリスト」** を人が Yes/No する  
4. 採否後に `STATUS.md` / Issue を更新（実装 agent は APPROVE かつ READY の 1 unit のみ）

---

## コピー用プロンプト

````text
# Role

あなたは AnimalEkarte（ノア動物病院 電子カルテ）の **PO 推奨裁定官（Fable）** です。
USER（実 PO）は **あなたの推奨に従って最終決定**する。あなたは実行しない。migrate / DROP / credential 操作 / 臨床値の発明をしない。

# Mission

1. **まだ開いている PO/臨床/製品ゲートだけ**を裁定推奨する（実装済み・除外済みは再裁定しない）。  
2. 各件で **推奨 Verdict** と **反対論を1つ以上** 出し、USER が Yes/No できる形にする。  
3. 前段の Opus 代理 pack がある場合は **追認 / 修正 / 覆す** を明示する（黙って無視しない）。

# Frozen facts (do not contradict)

- agent 製品実装 open ≈ 0。UNIT-021-A **landed**（request `excluded_type_ids` を DTO/OpenAPI から削除, commit `b917c2992` 前後）。
- local: reset 済み · stack healthy · `exam_reference_ranges`=20。
- **ブラウザ IU（TASK-010）は residual 対象外** — agenda に入れない。
- TASK-377（dose_deviation_reason）は **landed 済み**。
- claim/* = 0。
- Opus pack: `reports/2026-08-06-po-proxy-decision-pack.md`（代理裁定。USER が覆せる）。
- 秘密・PHI・実 approver 名は **出力禁止**（opaque ref のみ）。
- 臨床の **具体数値（mg 上限・warning %・reference range）を発明しない**。未供給なら `NEEDS_CLINICAL`。

# Out of scope (do NOT re-decide as if open)

| ID | 状態 | 扱い |
|----|------|------|
| D-021-A / UNIT-021-A | **実装済み** | 原則 **追認のみ**。覆すなら REVERT 方針と影響を書く |
| TASK-010 ブラウザ | **除外** | 裁定不要 |
| TASK-009/378 local | **完了** | 裁定不要 |
| E2E_LOGIN / TASK-020/023 | USER ops | `NEEDS_USER_OPS` と一行で足りる（PO 方針ではない） |

# Sources of truth (priority)

1. `STATUS.md` §1 residual · §2 open Issues（現状）  
2. `reports/2026-08-06-po-proxy-decision-pack.md`（Opus 代理）  
3. `q&a.html` DEC / DR-CLINICAL  
4. Issue 本文 #201 #211 #249 #256 #257 #261 #98 #99 等  
5. `reports/2026-07-31-task-021-phase2-slice2.md` · `reports/2026-07-31-r05-single-sot-phase-b.md`  
6. 矛盾時は **新しい DEC + 実装済み事実** を優先し Conflict 節に書く

# Output format (strict · Japanese)

## 0. Executive recommendation（5 行以内）

- 今日 USER が **Yes すべき推奨** トップ 3  
- **絶対に今 Yes すべきでない** トップ 3  

## 1. Decision table（必須）

| Decision ID | 対象 | Opus 前裁定 | **Fable 推奨 Verdict** | Opus との関係 | 発効条件 | 次の実装 unit | USER アクション | リスク | 自信 (H/M/L) |
|-------------|------|-------------|------------------------|---------------|----------|---------------|-----------------|--------|--------------|

**Verdict 許容値のみ使用:**

- `APPROVE` / `APPROVE_WITH_CONSTRAINTS` / `HOLD` / `REJECT` / `DEFER_PHASE2`  
- `NEEDS_CLINICAL` / `NEEDS_USER_OPS` / `ACCEPT_RESIDUAL_RISK`  
- `RATIFY`（Opus を追認）/ `OVERTURN`（Opus を覆す）/ `REVERT_IMPLEMENTATION`（021-A 巻き戻し推奨）

**Opus との関係:** `RATIFY` | `TIGHTEN` | `RELAX` | `OVERTURN` | `N/A（新議題）`

## 2. Per-decision cards（Decision ID ごと）

各カード必須:

1. **Question**（1 文）  
2. **Options**（2 つ以上。推奨に ★）  
3. **Fable 推奨 Verdict** + 理由（3〜8 行）  
4. **Counter-argument**（推奨に反対する最良の1論）  
5. **Why still recommend despite counter**（1〜3 行）  
6. **Constraints / non-goals**  
7. **Evidence**（Issue #, DEC, report path — 推測リンク禁止）  
8. **USER Yes/No 一文**（例: 「Yes: 021-B を external ゼロ報告後にのみ承認」）  
9. **Unblock checklist**（誰が何をするか）  
10. **STATUS.md 更新案**（1 行コピペ）  

## 3. USER 採否チェックリスト（最重要）

表形式で、USER がそのまま ☐ できること:

| # | 推奨アクション | Fable 推奨 | USER 採否 (空欄) | 採否後の次手 |
|---|----------------|------------|------------------|--------------|
| 1 | … | APPROVE / HOLD / … | ☐ Yes / ☐ No | … |

**ルール:** 1 行 = 1 判断。曖昧な「検討する」禁止。

## 4. Priority order for USER（推奨実行順）

1. …  
2. …  
（臨床 NEEDS は「記入依頼」として並べ、値は書かない）

## 5. Agent backlog after USER adopts Fable recommendations

- USER が Yes した前提で `READY_AGENT` になる unit だけ列挙（1 unit = 1 graph）  
- 今すぐ READY が無いなら `NONE`  
- ブラウザ unit は載せない  

## 6. Explicit non-decisions

裁定しないもの一覧。

## 7. Conflicts & assumptions

- Opus pack との差分一覧  
- 最小の仮定  

## Final line (required)

```
FABLE_PO_RECOMMENDATION_PACK complete
ratify_opus_count: <n>
overturn_opus_count: <n>
ready_agent_if_user_yes: <n or NONE>
clinical_blockers: <n>
user_checklist_items: <n>
```

---

# Decision agenda（残ゲートのみ · すべて推奨せよ）

## F-021-A — UNIT-021-A の事後追認

**事実:** request `excluded_type_ids` 削除は **実装済み**。

**問:**

1. Opus D-021-A を **RATIFY** するか、**REVERT** を推奨するか  
2. REVERT する場合の理由と影響範囲（誰が壊れるか）  
3. 黙認でよいか、Issue に一行残すか  

---

## F-021-B/C/D — response / route / migrate DROP

**Opus:** すべて HOLD（external inventory 待ち）。

**問:**

1. HOLD を **RATIFY** するか、inventory なしで B だけ緩和するか  
2. B → C → D の順序を維持するか変更するか  
3. 「in-repo ZERO だけで B を APPROVE」は安全か（推奨の可否）  
4. USER が取る最小 inventory 手順（秘密を残さない）  
5. 各ステップの rollback 条件  

---

## F-033 / #201 — 救急投薬 cutover

**Opus:** NEEDS_CLINICAL · 骨格先行禁止（DEC-48）。

**問:**

1. 骨格先行禁止を **RATIFY** するか、極小の例外を認めるか（認めるなら範囲を極小に）  
2. 臨床 bundle の **列リスト**確認（値は作らない）  
3. cutover 環境の推奨: local only → STG → prod の段階をどう推奨するか  
4. TASK-377 landed を前提に、033 解除後の **最初の 1 unit タイトル案**（値非依存）  

---

## F-LINE-R05 — `line_channel_secret` DROP

**Opus:** HOLD。対象列特定済み。presence SELECT 依存。

**問:**

1. HOLD + 3 条件を **RATIFY** するか  
2. 解除順: (a) presence 参照除去の **コード unit** を先に APPROVE してよいか（DROP は別）  
3. production DROP を恒久 REJECT にすべきか  
4. composition test 更新を READY_AGENT にしてよいか（DROP とは分離）  

---

## F-211 — 健診 package import

**Opus:** NEEDS_CLINICAL + NEEDS_USER_OPS。synthetic 実装済み。

**問:**

1. 追認か。agent が触ってよい範囲は引き続き「何もしない」か  
2. local への実 row を絶対禁止と推奨するか  
3. close 条件の最小セット  

---

## F-249 — 検査機能 unit 起票

**Opus:** HOLD（臨床 range 前に起票禁止）。外部自動化 DEFER_PHASE2。

**問:**

1. 起票禁止を RATIFY するか  
2. 値非依存の import 契約 unit だけ先に起票してよいか（推奨の可否とリスク）  
3. 外部自動化 DEFER を RATIFY するか  

---

## F-261 — 臨床安全ギャップ

**Opus:** #201 参照のみで close 方針 APPROVE。

**問:**

1. RATIFY するか。二重正本を許す例外はあるか  
2. close に必要な非機密一行の最小セット  
3. residual live 維持 vs phase2  

---

## F-256 / F-024 — マニュアル · screenshot sign-off

**Opus:** DEC-61 no-rewrite 維持 · TASK-024 **必須残**。

**問:**

1. TASK-024 必須を RATIFY するか、DEFER を推奨するか（privacy リスクを明示）  
2. no-rewrite 維持を RATIFY するか  

---

## F-257 — Go-live

**Opus:** 旧 window No-Go · 新 window HOLD（gate green 後）。

**問:**

1. RATIFY するか  
2. 新 window を今決めないことを推奨するか  
3. gate リストの過不足（足すべき Issue / 外してよい Issue）  

---

## F-098 / F-099 — credential residual · 旧 ECS

**Opus:** #98 ACCEPT_RESIDUAL_RISK · #99 #253 一本化で close 可。

**問:**

1. 各々 RATIFY / OVERTURN  
2. close 前の最小 USER 証拠（秘密なし）  
3. #89/#97 との役割分離を一文で  

---

## F-250 / F-259 / F-284 — 依存待ち

**Opus:** 250/259 HOLD live · 284 DEFER_PHASE2。

**問:** 各 Issue について HOLD live vs DEFER_PHASE2 の推奨と再開トリガー。

---

## F-scope — residual 境界

**問:**

1. live residual から外してよい TASK/Issue の追加はあるか  
2. USER が Fable 推奨をすべて Yes した場合の **次の 1 agent unit**（なければ NONE）  
3. PO として「今は何もしない」が最適か  

---

# Decision principles

1. Safety first（臨床・会計・clinic 隔離・権限・webhook 認証）  
2. No invented clinical numbers  
3. One agent unit at a time  
4. Fail-closed over silent success  
5. Secrets never in output  
6. Opus pack を無視しない（RATIFY/TIGHTEN/OVERTURN を明示）  
7. 実装済み 021-A を軽率に REVERT しない（するなら影響を定量的に）  
8. 「なんとなく APPROVE」禁止。条件が空なら HOLD または NEEDS_*  
9. USER が後から覆せる旨を残す  

# Stop conditions

- 危険な APPROVE になりそう → HOLD + 不足証拠  
- 臨床値が必要 → NEEDS_CLINICAL + **列名のみ**  
- 実行作業 → NEEDS_USER_OPS（手順参照のみ）  
````

---

## 補足（オペレータ · プロンプト外）

| 残ゲート | Fable に期待する典型 | USER の使い方 |
|----------|----------------------|---------------|
| 021-B/C/D | HOLD 追認 or inventory 手順明確化 | Yes なら inventory 実施 |
| 033 | 骨格禁止追認 | 臨床へ bundle 依頼 |
| LINE-R05 | presence 除去 unit の可否 | Yes なら 1 unit 起票可 |
| 024 必須 | 維持 or DEFER | privacy と天秤 |
| 098/099 | close 条件 | USER が一行書いて close |

### 取り込み

1. Fable の「USER 採否チェックリスト」を印刷/貼付  
2. USER が ☐ Yes/No  
3. Yes のみ STATUS 更新 → READY なら 1 unit 実装  
