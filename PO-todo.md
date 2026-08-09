# PO / USER 実施 ToDo

| 項目 | 値 |
|------|-----|
| **作成** | 2026-08-06 |
| **状況更新** | **2026-08-09**（U0–U6 residual · PO-07 · UAT-R1 land 後） |
| **対象読者** | あなた（PO / オペレータ） |
| **方針の正本** | [`docs/work/decisions/fable-po-recommendation.md`](docs/work/decisions/fable-po-recommendation.md)（**採択済み**） |
| **作業台帳** | [`STATUS.md`](STATUS.md) §1 · §2 |
| **実行証跡** | [`docs/work/residual-closeout-ledger.md`](docs/work/residual-closeout-ledger.md) |
| **除外** | ブラウザ IU 検証（TASK-010）· agent 製品実装（現在 NONE） |
| **環境前提** | **PROD 未構築** · **STG ほぼ未使用** · local は UAT 時に `make up` |

---

## 認識の確認（重要）

### 相違があります — 次の理解に合わせてください

| 誤解しやすい言い方 | 正しい整理 |
|--------------------|------------|
| 「Fable で**解決できなかった**ものだけが残タスク」 | **半分だけ正しい** |
| | Fable は **方針・Verdict（APPROVE / HOLD / DEFER / NEEDS_*）をほぼすべて推奨し、USER が採択済み** |
| | 本ファイルに載るのは「**方針未決**」ではなく、「**方針は決まったが、人が実行・依頼・証拠集めしないと進まない残作業**」 |

### Fable がやったこと / やらなかったこと

| Fable（採択済み） | 本ファイルの ToDo |
|-------------------|-------------------|
| HOLD / APPROVE / DEFER / NEEDS_CLINICAL などの **判定** | その判定を **現場で完了させる手作業** |
| 021-A 追認 · 033 骨格禁止 · LINE-R05 条件整理 など | provider 確認、inventory、臨床への依頼文、sign-off |
| 臨床 **数値の発明はしない**（正しい） | 臨床責任者への **bundle 記入依頼**（値は彼らが書く） |
| credential **実行の代行はしない**（正しい） | #89/#97 の **実行**（#98/#99 close は **済**） |

### 本ファイルに **書かない** もの

- Fable が既に **DEFER_PHASE2** とした依存（例: #284）— 再開トリガーが来るまで放置でよい  
- agent が今やる製品 unit（**NONE**）  
- ブラウザ IU バッチ  
- 既に local 完了した reset / seed / UNIT-021-A / DEC-68 起票  

---

## 優先 ToDo（上から順）

### 帯 1 — 判断・1行記録（その日で可）

- [x] **PO-01 · #98**  
  - **何を**: 旧 RDS 系 credential が provider 側で失効済みか確認  
  - **完了条件**: Issue に非機密1行（結果 enum + opaque ref）→ close  
  - **Fable**: `ACCEPT_RESIDUAL_RISK`（close 経路は確定済み。**実行だけ残**）  
  - **書かない**: 秘密・接続文字列  
  - **完了 (2026-08-08)**: #98 CLOSED · enum=`ACCEPT_RESIDUAL_RISK_WITH_OPAQUE_REF` · opaque_ref=`2026-08-08-PO-attestation-F098`（秘密なし）

- [x] **PO-02 · #99**  
  - **何を**: 旧 ECS deploy 経路が実行不能であることの確認  
  - **完了条件**: 同上1行 + rollback SoT = **#253** → close  
  - **Fable**: `APPROVE`（close 経路確定 · **実行だけ残**）  
  - **完了 (2026-08-08)**: #99 CLOSED · enum=`WORKFLOW_ABSENT` · opaque_ref=`2026-08-08-PO-attestation-F099` · rollback_sot=#253（秘密なし）  

- [x] **PO-03 · #252 ↔ #257**  
  - **何を**: Go-live 前提 gate に **#252（締め時間）を含めるか** Yes/No  
  - **完了条件**: #257 または STATUS に一行メモ  
  - **Fable**: 候補提示 · **最終 Yes/No は USER**  
  - **完了 (2026-08-08)**: go-live_gate_include_#252=`YES` · opaque_ref=`2026-08-08-PO-attestation-F257-gate252` · recorded on STATUS §2 #257 · #252/#257 remain OPEN · new go-live window **not** set (F-257 HOLD) · 秘密なし  
 

---

### 帯 2 — 認証付き検証を進めたいとき

- [x] **PO-04 · E2E 資格情報**  
  - **何を**: host に `E2E_LOGIN_EMAIL` / `E2E_LOGIN_PASSWORD` を注入  
  - **完了条件**: シェルで SET（値は repo に書かない）  
  - **Fable**: `NEEDS_USER_OPS`（PO 方針ではなく ops）  
  - **2026-08-07**: `.env.local` に SET · login cookie 200（値は chat/repo に書かない）

- [x] **PO-05 · TASK-020**（partial → core green）  
  - **何を**: Playwright 実行  
  - **完了条件**: green / 失敗ログを手元保管  
  - **2026-08-07**: core 16-spec suite **80/80** PASS（placeholder/kana seed 整合 + 一覧は DataTableRowLink クリック）

- [x] **PO-06 · TASK-023 / #254**  
  - **何を**: 5 業務フロー UAT  
  - **完了条件**: 結果 enum 一行（PASS/FAIL + メモ）  
  - **完了 (2026-08-08)**: overall=`PARTIAL` f1=PASS f2=PASS f3=FAIL f4=UNTESTED f5=BLOCKED · opaque_ref=`2026-08-08-PO-uat-TASK-023` · memo=stack local demo; f3 trimming fail TBD; f4/f5 not run · #254 remains OPEN (close not performed) · 秘密なし  
  - **UAT-R1 residual (2026-08-09)**: RECORD_ONLY re-record f3–f5 · overall=`PARTIAL` f1=PASS(INHERIT_U4) f2=PASS(INHERIT_U4) f3=FAIL f4=UNTESTED f5=BLOCKED · opaque_ref=`2026-08-09-PO-uat-R1` · memo=no re-run yet; stack may be down; residual f3-f5 unchanged from U4 · env fe/be `000` DOWN · #254 remains OPEN · no re-run yet · 秘密なし  

前提 UI: http://localhost:3003 · 落ちていれば `make up`。

---

### 帯 3 — 破壊削除・本番 DROP の前提証拠

- [x] **PO-07 · TASK-021 inventory（B 用）**  
  - **何を**: **client registry** — in-repo FE / LIFF **以外**の API consumer がいない、と一行宣言  
  - **完了条件**: STATUS または Issue に非機密一行  
  - **なぜ**: Fable TIGHTEN — B（`excluded_courses`）は access log だけでは証明不可  
  - **Fable**: B/C/D は **HOLD**（方針決済済み · **証拠集めだけ残**）  
  - **完了 (2026-08-09)**: declaration=`NO_KNOWN_EXTERNAL_CONSUMERS` opaque_ref=`2026-08-09-PO-TASK-021-registry` inventory_start=`2026-08-09` memo=in-repo FE/LIFF only known consumers of exclusion surfaces; no partner/mobile clients known · In-repo scan: FE/LIFF/OpenAPI/handler cites on STATUS TASK-021 + ledger（paths only）· B/C/D DROP not performed · 秘密なし  

- [ ] **PO-08 · TASK-021 inventory（C 用）**  
  - **何を**: STG/prod access log で `excluded-reservation-types` の **90日呼び出し件数**（path + 件数のみ）  
  - **完了条件**: 件数の一行（token / IP / UA は書かない）  
  - **2026-08-09 状況**: **後回し** — PROD **未構築** · STG **ほぼ未使用**のため本集計は時期尚早。F-021-X clock は 2026-08-09 開始（~2026-11-07 再裁定）。実トラフィック後に更新  
 

- [x] **PO-09 · 021 inventory 依頼日のメモ**  
  - **何を**: 依頼開始日を記録  
  - **完了条件**: 90日無応答なら F-021-X どおり **再裁定**に上げる（silent HOLD 禁止）  
  - **完了 (2026-08-09)**: inventory_start=`2026-08-09` · F-021-X 90 日 clock 開始 · opaque_ref=`2026-08-09-PO-TASK-021-registry`（PO-07 同一）  

- [ ] **PO-10 · LINE-R05 presence inventory**  
  - **何を**: clinic ごとの legacy presence **有無の件数のみ**（secret 値は保存しない）  
  - **完了条件**: ゼロなら agent に presence 参照除去 unit を依頼可  
  - **Fable**: DROP は HOLD · 条件② composition は **済** · 残①③ · **inventory 前の参照除去は禁止**  

G〜J が揃うまで **response 削除 / route DROP / DB DROP / 本番 secret 列 DROP はしない**。

---

### 帯 4 — 依頼（値は書かない · 記入は先方）

- [ ] **PO-11 · 臨床へ #201 bundle 依頼**  
  - **何を**: 臨床責任者に DR-CLINICAL #201 **1行**記入を依頼（列名は Fable pack / q&a 参照）  
  - **完了条件**: 依頼送付 + 受領待ち  
  - **Fable**: `NEEDS_CLINICAL`（**方針=記入まで禁止**は解決済み · **記入そのものは残**）  
  - **あなたが書かない**: mg 上限 · warning % · 動物種 default  

- [ ] **PO-12 · 臨床へ #249 range 依頼**  
  - **何を**: reference range 承認行  
  - **Fable**: HOLD（承認前 unit 起票禁止）  

- [ ] **PO-13 · #211 臨床 + ops**  
  - **何を**: 実 row 承認（臨床）+ 対象 clinic/環境（ops）。manifest は repo 外  
  - **Fable**: NEEDS_CLINICAL + NEEDS_USER_OPS · 実 row は local も禁止  

---

### 帯 5 — 人の証跡（納品・privacy）

- [x] **PO-14 · TASK-024 / #256**  
  - **何を**: screenshot / FAQ visual sign-off（Privacy + Repository owner）  
  - **Fable**: **必須残**（DEFER しない）  
  - **完了 (2026-08-08)**: privacy=`SIGNED_OFF` repo=`SIGNED_OFF` overall=`PASS` · opaque_ref=`2026-08-08-PO-signoff-TASK-024` · memo=visual FAQ+manual screenshots reviewed off-repo; no PII in comment · #256 remains OPEN (close not performed unless separately approved) · 秘密なし  

- [x] **PO-15 · TASK-022**  
  - **何を**: S13 手動 correction + RLS 証跡  
  - **完了 (2026-08-09)**: s13=`NOT_RUN` rls_runtime=`N_A_APP_LAYER_ONLY` env=`UNKNOWN` · opaque_ref=`2026-08-09-PO-TASK-022-residual` · memo=S13 human 1-8 not executed this session; RLS runtime out of S13; app-layer identitylink tests only · #239 already CLOSED (no reopen/close; LEDGER_PO only) · 秘密なし・PHI なし  

- [ ] **PO-16 · #261**  
  - **何を**: 5項目を 結果 enum + opaque ref の1行ずつ  
  - **Fable**: #201 参照のみで close · **値の複製禁止**  

---

### 帯 6 — 他環境・本番（必要なとき）

- [ ] **PO-17 · POST-PULL**  
  - 他 env: `make migrate`（checksum mismatch 時は reset 判断）  

- [ ] **PO-18 · #89 / #97**  
  - credential ローテの **実行**（#98 の受容 close とは別）  

- [ ] **PO-19 · #253**  
  - 本番 CI/CD · backup gate  

- [ ] **PO-20 · #257 Go-live**  
  - **全 gate green 後**に新 window を決める（今は日付を決めない）  
  - **Fable**: HOLD · 旧 window No-Go 維持  

---

### 帯 7 — 催促のみ（依存待ち）

| ID | 待ち | Fable |
|----|------|-------|
| **#250** | producer bundle | HOLD live |
| **#259** | Lステップ先方 enable | HOLD live |
| **#284** | 実機3台 | **DEFER_PHASE2**（今フェーズの必須 ToDo ではない） |

---

## 今日の最小セット（推奨 · **2026-08-09**）

**済**: PO-01〜07 · PO-09 · PO-14〜15 · residual U0–U6 · UAT-R1 記録（穴は未解消）

1. **local UAT 穴** — `make up` → f3 / f4 / f5 実通し（UAT-R1 は re-record のみ · overall は PARTIAL のまま）  
2. **S13 人手** — 手順 1–8（現状 NOT_RUN）  
3. **PO-11** — #201 臨床へ依頼文 1 通  
4. **PO-08** — STG 利用増 or PROD 後に access-log 件数（今は環境未整備で後回しでよい）  



---

## 完了したら

- チェックを `[x]` に更新するか、STATUS §1 / Issue に非機密1行を残す  
- 製品実装が必要になったら（021-B 承認後など）agent に **1 unit** だけ依頼  
- 本ファイルは **USER 実行リスト**。方針の再裁定は Fable/Opus pack と `q&a.html` DEC を正本とする  

---

## 関連リンク

| 文書 | 役割 |
|------|------|
| [`STATUS.md`](STATUS.md) | 全体 SoT |
| [`q&a.html#dec-68`](q&a.html#dec-68) | TASK-021 4段階 |
| [`docs/work/decisions/fable-po-recommendation.md`](docs/work/decisions/fable-po-recommendation.md) | 採択した推奨 |
| [`docs/ops/deploy/LOCAL_DB_RESET.md`](docs/ops/deploy/LOCAL_DB_RESET.md) | local DB reset 手順 |
| [`docs/ops/testing/scenarios/`](docs/ops/testing/scenarios/) | シナリオ原文 |