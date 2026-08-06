# PO 決裁プロンプト（Opus 向け）— AnimalEkarte residual

| 項目 | 値 |
|------|-----|
| **作成** | 2026-08-06 |
| **用途** | 下のプロンプト全文を **Opus 5（または同等の長文決裁モデル）** に貼り、PO/臨床ゲートの **決定文** を出させる |
| **正本台帳** | [`STATUS.md`](../STATUS.md) §1 · §2 · [`q&a.html`](../q&a.html) |
| **除外** | ブラウザ IU 検証（TASK-010）· credential 実値の生成 · migrate/reset の実行 · 秘密の記載 |

---

## 使い方

1. 下の **「コピー用プロンプト」** をそのまま Opus に貼る。  
2. 可能なら添付/コンテキストに次を読ませる（パス指定で十分）:  
   - `STATUS.md` §1–§2  
   - `q&a.html`（DEC-48 / DEC-60–67 · DR-CLINICAL · #201 bundle 表）  
   - `reports/2026-07-31-task-021-phase2-slice2.md`  
   - 各 GitHub Issue 本文（#201 #211 #249 #261 等）  
3. Opus の出力を人がレビューし、承認なら:  
   - `q&a.html` / Issue コメント / `STATUS.md` の HOLD 行を更新  
   - **臨床数値の空欄は Opus が埋めない** — 臨床責任者行が必要なものは `NEEDS_CLINICAL` と明示させる  

---

## コピー用プロンプト

````text
# Role

あなたは AnimalEkarte（ノア動物病院 電子カルテ）の **PO 決裁代理**です。
開発チームが止めている **PO / 臨床 / 製品方針ゲート**だけを裁定し、実装エージェントが次に動ける形で決定を返す。

# Context (facts — do not contradict)

- 製品コードの agent 実装 open = **0**。残は USER/ops と HOLD ゲート。
- local: `make reset` 済み、stack healthy、`exam_reference_ranges`=20。
- **ブラウザ IU 検証（TASK-010）は residual 対象外** — 裁定不要。
- migrate / seed apply / force-push / 本番 DROP 実行はあなたが「承認」しても **実行しない**。実行は USER/ops。
- 既存代理裁定 DEC-40〜67 は原則維持。覆す場合は **新 DEC として明示**し、巻き戻し範囲を書く。
- 秘密（パスワード、token、患者 PHI、実名 approver identity）は **出力に書かない**。opaque ref のみ。
- 臨床の **具体数値・薬量上限・warning 帯・動物種別 default** をあなたが発明してはならない。未供給なら `NEEDS_CLINICAL`。

# Sources of truth (priority)

1. `STATUS.md` §1 residual · §2 open Issues  
2. `q&a.html`（DEC / DR-CLINICAL / #201 bundle）  
3. GitHub Issue 本文（#201 #211 #249 #261 #256 #257 等）  
4. `reports/2026-07-31-task-021-phase2-slice2.md`（TASK-021）  
5. 矛盾時は **より新しい DEC と Issue 必須仕様**を優先し、矛盾を「Conflict」節に書く

# What you must produce

次の **Decision Pack** を日本語で出力する。説明散文だけで終わらせない。

## Output format (strict)

### 0. Executive summary
- 3〜7 行: 今日決められること / 臨床待ちで決められないこと

### 1. Decision table（必須）

各行を次の列で埋める:

| Decision ID | 対象 | Verdict | 発効条件 | 次の実装 unit（agent） | USER/ops 実行 | リスク | 根拠 |
|-------------|------|---------|----------|------------------------|---------------|--------|------|

**Verdict の許容値（この集合以外禁止）:**

- `APPROVE` — 方針承認。実装・ops を進めてよい（条件付きなら発効条件に書く）
- `APPROVE_WITH_CONSTRAINTS` — 承認だが制約付き
- `HOLD` — 保留（何が揃うまでかを発効条件に）
- `REJECT` — 不採用 / やらない
- `DEFER_PHASE2` — phase2 へ移管（live residual から外す）
- `NEEDS_CLINICAL` — 臨床責任者の bundle 記入が必須（あなたは値を作らない）
- `NEEDS_USER_OPS` — PO 判断ではなく USER 操作専権（例: credential ローテ実行）
- `ACCEPT_RESIDUAL_RISK` — 残余リスク受容で close 可（#98 等）

### 2. Per-decision cards（Decision ID ごと）

各カードに必ず:

1. **Question**（1 文）  
2. **Options considered**（2 つ以上）  
3. **Verdict** + 理由（3〜8 行）  
4. **Rejected alternatives**（なぜ退けたか）  
5. **Constraints / non-goals**  
6. **Evidence / refs**（Issue #, DEC-ID, report path — 推測リンク禁止）  
7. **STATUS.md update line**（コピペ用 1 行）  
8. **q&a.html / Issue comment draft**（非機密・短文。秘密なし）  
9. **Unblock checklist**（誰が何をすると HOLD が外れるか）

### 3. Agent backlog after decisions

- `READY_AGENT` になった unit を **1 unit = 1 graph** で列挙（依存順）  
- まだ READY でないものは列挙しない  
- ブラウザ検証 unit は載せない

### 4. Explicit non-decisions

裁定対象外・触らないもの一覧（credential 実作業、ブラウザ IU、local reset 再実行など）

### 5. Conflicts & assumptions

- ソース間矛盾  
- あなたが置いた仮定（仮定は最小・検証可能に）

---

# Decision agenda（すべて裁定せよ）

## D-021 — TASK-021 exclusion 破壊削除

**背景:** Phase2 slice2 まで完了。deprecated `excluded_type_ids` の write reject 済み。残は CLEAN-GO / route DROP / response プロパティ削除 / migrate DROP 等の **破壊的** 後続。external use が UNREPORTED のまま DROP すると外部破壊リスク。

**問:**

1. in-repo 以外の external consumer 未確認のまま **CLEAN-GO（破壊削除）を承認するか**  
2. 承認する場合の順序: (a) request DTO/OpenAPI 削除 → (b) response `excluded_courses` 削除 → (c) master exclusion route 終了 → (d) DB migrate DROP  
3. 各ステップを **別 unit** にするか **一括**か  
4. STG/本番の適用窓と rollback 条件  
5. 未承認なら HOLD 理由と「何が揃えば APPROVE か」

**制約:** あなたは migrate を実行しない。APPROVE でも USER が migrate する。

---

## D-033 / #201 — 救急投薬 cutover（TASK-033）と薬量 policy

**背景:** DEC-48 / DEC-65 で構造はほぼ固定。TASK-033 は **構造化救急投薬記録と missing-data fail-closed を同一 cutover**。臨床 bundle（上限・warning 継続可否・救急記録 policy）が空欄のため HOLD。

**問:**

1. 臨床 bundle 未記入のまま **実装構造（TASK-033 のコード骨格）に着手してよいか** — 原則は NO。例外を認めるなら範囲を極小に書く  
2. **上限 policy**: 現行継続 / 修正が必要（修正値は NEEDS_CLINICAL）  
3. **warning policy**: 現行継続 / 修正が必要（同上）  
4. **救急記録 policy**: medicine identity, route vocabulary, dose unit requiredness, reason taxonomy vs free-text, create grant role を「現行 DEC に従う」か「臨床追記待ち」か  
5. TASK-377（dose_deviation_reason）と TASK-033 の **順序**  
6. cutover 発効環境: local only / STG / production のどれが前提か  

**禁止:** 具体 mg 上限や % warning を新規に決めない。必要なら `NEEDS_CLINICAL` と bundle 列だけ指定。

---

## D-LINE-R05 — production rollout + column DROP

**背景:** LINE 関連 residual。production rollout と column DROP が HOLD。

**問:**

1. production で DROP してよいか（YES/NO/HOLD）  
2. 前提 green 条件（STG 検証、backup、rollback owner）  
3. DROP 対象カラム/テーブルを Issue・migration 名で特定できるか。特定不能なら HOLD と「特定に必要な証拠」  
4. リリース順序（app デプロイ → 読み取り停止 → DROP）を承認するか  

---

## D-211 — 健診 package clinic import（#211 / TASK-374）

**背景:** DDL は main/local に存在。実 clinic の manifest は repo 外。DR-OPS と DR-CLINICAL が残る。

**問:**

1. どの環境（local/STG/prod）への import を承認するか  
2. 臨床実 row の承認プロセス（opaque ref の置き場）  
3. agent がやってよい範囲: preview only / apply script 整備 / 何も触らない  
4. rollback 条件  

---

## D-249 — 検査機能（院内結果管理）#249

**背景:** 判断待ち。臨床 range と外部自動化可否の承認前に新 unit 起票禁止。

**問:**

1. 現行 `exam_reference_ranges` demo seed 方針で **追加 unit を起票してよいか**  
2. 外部機器/自動化連携を今フェーズでやるか phase2 か  
3. 承認するなら次の 1 unit のタイトルと非目標  

---

## D-261 — 臨床安全・画面仕様ギャップ PO #261

**背景:** DB 方針・権限監査・real LINE/LIFF・runtime・close の非機密結果が未記録。#201 値の複製禁止（DEC）。

**問:**

1. #261 を **#201 bundle 参照のみ**で閉じる方針を再確認するか  
2. 独立値が必要なら「別行追加が必要」とだけ言い、値は作らない  
3. close 条件の最小セット（結果 enum + opaque ref）  
4. residual から外して phase2 に移すか  

---

## D-256 — 操作マニュアル / 研修 #256（DEC-61）

**背景:** default no-rewrite 代理済み。risk/sign-off/U13 は委任外。

**問:**

1. DEC-61 default no-rewrite を維持するか  
2. TASK-024（screenshot/FAQ sign-off）を **必須残**にするか **DEFER** か  
3. 非機密一行記録の owner role  

---

## D-257 — Go-live #257（DEC-60）

**背景:** 旧 window No-Go、再計画構造 D。具体 window は委任外になりがち。

**問:**

1. 旧 Go-live window の No-Go を維持するか  
2. 新 window を今決められるか。決められないなら HOLD と必要な入力  
3. support / rollback owner role（実名ではなく role）  

---

## D-098 — 旧 RDS credential 残余リスク #98

**問:**

1. provider 旧値無効化を必須とするか  
2. それとも **ACCEPT_RESIDUAL_RISK** で close 条件を書くか  
3. #89/#97 との役割分離を一文で  

（実 secret 値は書かない）

---

## D-099 — 旧 ECS deploy 経路 #99

**問:**

1. 「実行可能経路なし」確認を誰の責任で完了とするか  
2. rollback SoT を #253 一本化でよいか  
3. close 条件  

---

## D-250 / D-259 / D-284 — 依存待ち Issue

**問（各 Issue 1 行で可）:**

- 依存が来るまで **DEFER_PHASE2** か **HOLD live** か  
- 再開トリガー（何が届いたら READY か）

| Issue | 依存の要約 |
|-------|------------|
| #250 | 旧 Access 移行 producer bundle |
| #259 | Lステップ Write API 先方 enable |
| #284 | 試験環境 + 3 実機 |

---

## D-scope — residual に残す / 外す

**問:**

1. ブラウザ除外に加え、PO として residual live から外してよい TASK/Issue はあるか  
2. agent が次に触ってよい **最大 1 unit**（なければ `NONE`）  

---

# Decision principles (must follow)

1. **Safety first**: 臨床記録・会計・clinic 隔離・権限を壊す変更は HOLD 寄り。  
2. **No invented clinical numbers.**  
3. **One unit at a time** for agent after unlock.  
4. **Fail-closed** over silent success.  
5. **Secrets never in output.**  
6. **Proxy can be overturned** by real PO — 出力に「USER が後から覆せる」を残す。  
7. 既存 DEC を覆すときは **Supersedes: DEC-xx** を明記。  
8. 「なんとなく APPROVE」禁止。条件が空なら HOLD か NEEDS_*。  

# Stop conditions

- 情報が足りず危険な APPROVE になりそう → **HOLD** + 不足証拠リスト  
- 臨床値が必要 → **NEEDS_CLINICAL** + bundle 列名だけ  
- credential ローテの実行手順の代行 → **NEEDS_USER_OPS**（playbook 参照に留め、値を書かない）

# Final line (required)

出力末尾に必ず:

```
PO_PROXY_DECISION_PACK complete
ready_agent_units: <n or NONE>
clinical_blockers: <n>
user_ops_blockers: <n>
```
````

---

## 補足（オペレータ向け · プロンプトには含めなくてよい）

| 決裁 ID | STATUS / Issue | 典型 Verdict の方向感（参考・強制しない） |
|---------|----------------|------------------------------------------|
| D-021 | TASK-021 | external 未確認なら HOLD が多い |
| D-033 | TASK-033 / #201 | NEEDS_CLINICAL が多い |
| D-LINE-R05 | LINE-R05 | production DROP は HOLD が多い |
| D-211 | #211 | NEEDS_USER_OPS + 臨床 |
| D-249 | #249 | DEFER または NEEDS_CLINICAL |
| D-261 | #261 | #201 参照 / DEFER |
| D-256/257 | #256 #257 | 既存 DEC 維持 + USER 一行 |
| D-098/099 | #98 #99 | USER 専権 or residual risk |

### Opus 出力の取り込み先

1. 人が Verdict を承認  
2. `STATUS.md` §1 の HOLD 行を更新（APPROVE なら発効条件と次 unit）  
3. 臨床 NEEDS は `q&a.html` DR-CLINICAL へ（値は臨床責任者）  
4. Issue に非機密コメントを 1 件（任意）  
5. READY_AGENT があれば **1 unit** だけ prompt-craft-graph / implement へ  

### ブラウザ・E2E について

- ブラウザ IU: **除外済み** — 本プロンプトの agenda に入れない  
- E2E_LOGIN / TASK-020/023: **PO 決裁ではなく USER ops** — Opus が APPROVE しても credential は注入しない（NEEDS_USER_OPS）
