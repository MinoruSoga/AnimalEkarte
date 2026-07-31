# LINE 残差 R-01〜R-07 — PO 意思決定支援ブリーフ（非拘束）

> **Date**: 2026-07-31  
> **Unit**: `TODO-MD-OPEN-REMAINING-ORCH-WAVE-20260731-V2` / role `W-PO-LINE`  
> **Evidence source**: [`reports/2026-07-31-task-019-line-deep-audit.md`](2026-07-31-task-019-line-deep-audit.md) §4 Residual disposition table  
> **Mode**: Options + evidence only. **Not binding PO law.**  
> **非実施**: 製品コード変更なし。webhook イベント契約や credential dual-store SoT を製品法として発明しない。

---

## 0. 効力（必読）

| 項目 | 内容 |
|:---|:---|
| 本書の性質 | 意思決定支援（decision-support）。推奨は **議論用・非拘束** |
| 拘束力 | **USER が採用するまで binding ではない** |
| 禁止 | 本ブリーフを webhook policy / dual-store SoT の製品正本として扱うこと |
| 実装 | 本 unit はレポートのみ。採用後に別 TASK で実装計画へ落とす |

---

## 1. Scope

| ID | Primary? | Deep-audit disposition | 本ブリーフの扱い |
|:---|:---|:---|:---|
| **R-01** | **Yes** | 要PO | 本編で options 表 |
| **R-05** | **Yes** | 要PO | 本編で options 表 |
| **R-06** | Secondary | 要PO | Wave1 **nav honesty 実装意図 in progress** を注記。intentional-hide 選択肢が残る限り 要PO を維持 |
| **R-07** | Secondary | 要PO / 将来 BUG 候補 | 同上（RBAC 整合）。実装進行中でも intentional 乖離を残すなら 要PO |
| R-02 / R-04 / R-08 | Ops only | ops / USER | §4 一行メモのみ |
| R-03 | Out of scope | → TASK-010 | 本ブリーフ対象外（runtime） |

**Wave1 note (R-06 / R-07):**  
監査時点の事実は deep-audit 表どおり（delivery-monitor は route あり・`paths.lstep`/サイドバー未登録；タグは route=`ResourceLstepAnalytics` vs sidebar=`ResourceHospitalSettings`）。Wave 側で nav honesty 修正が進行中でも、**「意図的に隠す / 意図的に権限をずらす」** を製品方針として残す選択肢がある限り PO 残差はクローズしない。実装完了後は「漏れ修正で close」か「intentional として docs 固定」かを USER が決める。

---

## 2. R-01 — Webhook イベント契約の文書化場所

### 2.1 Evidence（deep audit）

| 項目 | 内容 |
|:---|:---|
| Claim | `architecture` に webhook イベント契約表が無い |
| Code | 実装は **follow / unfollow のみ**処理し、他 event type は skip |
| Signature | destination → `line_bot_user_id` の O(1) HMAC（`verifySignatureAnyClinic`） |
| Code paths | `line_link_service.go` `HandleWebhook`; `line_link_signature_routing_test.go`; handler コメント（follow/unfollow 想定） |
| Doc path | `docs/spec/line/architecture.md` |
| Disposition | **要PO** — イベント契約を architecture に書くか、コード + ops 正本のままか |

### 2.2 Options（非拘束）

| Option | Pros | Cons | Who decides |
|:---|:---|:---|:---|
| **A** — architecture にイベント契約表を追記（処理する type = follow/unfollow、他は skip、署名経路を1段落） | 運用・監査・onboarding で「何を受けるか」が読める。docs honesty と一致しやすい | 表を SoT にすると将来 type 追加時の docs drift 管理が要る。Messaging API 製品契約の誤解（全イベント対応と読まれる）リスク | PO + docs owner |
| **B** — architecture は要約のみ；詳細はコード/tests + `setup`/ops のまま | 実装が単一の真実源。過剰な製品約束を増やさない | 読者が architecture だけでは契約を把握できない。R-02 ops 依存が続く | PO + ops |
| **C** — 契約表を `setup.md`（ops）側に置き architecture からはリンクのみ | 製品 architecture と運用手順を分離 | setup が「本番手順未完結」（R-02）と混ざり、正本が薄くなる可能性 | PO + ops |

### 2.3 Recommended for discussion only（非拘束）

**A（最小表）**: architecture に「処理するイベント = follow/unfollow；その他 type は成功スキップ；署名は destination→line_bot_user_id HMAC」の **事実記述表** を置く。  
**新規イベント対応・message reply 方針は発明しない**（現状コードの honesty のみ）。採用は USER。

---

## 3. R-05 — LINE Channel Secret 二重ストア SoT

### 3.1 Evidence（deep audit）

| 項目 | 内容 |
|:---|:---|
| Claim | Channel Secret の **二重ストア** |
| Store 1 | `clinic_integrations`（L-step settings 更新経路） |
| Store 2 | `line_reservation_settings`（webhook HMAC / LIFF settings 用途） |
| Code paths | `lstep_settings_update.go`; reservation `line_reservation_setting`; `verifySignatureAnyClinic` |
| UI 分担 | screen28 は line-reservation UI で Channel Secret を扱わない（V-14: 設定は L-step 連携側） |
| Disposition | **要PO** — 製品/運用の source-of-truth と dual-write 方針。**本パケットでは発明しない** |

### 3.2 Options（非拘束）

| Option | Pros | Cons | Who decides |
|:---|:---|:---|:---|
| **A** — 単一 SoT を決め、もう一方は read 投影 or 廃止（要移行設計） | 二重管理禁止（product philosophy）に沿う。秘密ローテーションが1経路 | 移行・互換・HMAC 経路の切替リスク。migration / dual-write 期間の設計が必須 | PO + security/BE lead |
| **B** — 現状維持 + 文書で役割分担を honesty 固定（どちらが write UI / どちらが verify 読取） | 即時の破壊的変更なし。運用で「どこを直すか」が明確になる | 二重ストア自体は残る。ドリフト時の署名失敗・部分更新が再発し得る | PO + ops |
| **C** — 設定 UI 更新時の **atomic dual-write** を製品契約にし、両ストア一致を invariant 化 | 運用ミスで片方だけ更新されにくい | 実装コスト・TX 境界・失敗時ロールバック契約が要る。SoT は「両方が正」になり哲学上グレー | PO + BE lead |

### 3.3 Recommended for discussion only（非拘束）

**B を短期 honesty、A を中期目標の議論枠**:  
まず「write UI は L-step settings / verify が読む場所は何か」を **事実として docs に固定**（発明ではなく現状マッピング）。単一 SoT 収束（A）は別設計 TASK とし、本ブリーフでは dual-store を製品法として確定しない。  
**Credential のコピー値・実シークレットは扱わない。**

---

## 4. R-06 — delivery-monitor ナビ有無

### 4.1 Evidence（deep audit）

| 項目 | 内容 |
|:---|:---|
| Claim | `/lstep/delivery-monitor` は **route 実装済**だが `paths.lstep` とサイドバー「Lステップ連携」に **未登録**（deep-link のみ） |
| Code | `operations-routes.tsx`（`ResourceLstepAnalytics` guard）; `paths.ts`（settings/tags/checkup-sync/analytics のみ）; `sidebar-menu.tsx`（monitor 行なし） |
| Docs | screen34 / README（D-05 で README 索引追加済） |
| Disposition | **要PO** — intentional deep-link か nav 漏れか |

### 4.2 Status after Wave1 fix intent

- **Nav honesty 実装は進行意図あり**（本ブリーフ執筆時点では deep-audit 事実＝未登録を前提に判断材料を残す）。
- 修正が「漏れ修正」で完了すれば R-06 は **実装 close** 候補。
- **intentional-hide（メニュー非表示・URL 直叩きのみ）** を製品方針として残すなら、PO 残差は docs 固定まで残る。

### 4.3 Options（非拘束）

| Option | Pros | Cons | Who decides |
|:---|:---|:---|:---|
| **A** — サイドバー + `paths.lstep.deliveryMonitor` に追加（analytics と同権限） | screen34 / README と到達性が一致。発見可能性が高い | 運用画面の露出が増える（意図的に隠したい場合は不適） | PO + FE |
| **B** — intentional deep-link のまま；spec/README に「ナビ非掲載」を honesty 記載 | 現状ルートを維持。権限ある人だけ URL 共有 | 監査・新人運用で「無い機能」と誤認されやすい | PO + docs |
| **C** — ナビは analytics 配下の副リンク等に限定表示 | 折衷 | 情報設計の一貫性が薄れやすい | PO + FE |

### 4.4 Recommended for discussion only（非拘束）

**A**（漏れ修正としてナビ honesty）— ただし USER が「運用モニタは共有 URL のみ」と決めるなら **B**。Wave1 実装が A 方向なら PO は「intentional-hide を採らない」旨の確認だけで close 可。

---

## 5. R-07 — タグ管理: sidebar 権限 vs route guard

### 5.1 Evidence（deep audit）

| 項目 | 内容 |
|:---|:---|
| Claim | タグ管理: **route guard = `ResourceLstepAnalytics`**、**sidebar = `ResourceHospitalSettings`** → nav 表示と到達可否が乖離し得る |
| Code | `settings-routes.tsx` vs `sidebar-menu.tsx` |
| Screen | screen31 |
| Disposition | **要PO** / 将来 BUG 候補（RBAC 整合。deep audit ではコード変更なし） |

### 5.2 Status after Wave1 fix intent

- Nav / RBAC honesty 修正が進行意図あり。
- どちらに寄せるか（HospitalSettings vs LstepAnalytics）は **製品の権限モデル判断**であり、機械的一致だけでは決まらない。
- intentional に「見せるが入れない / 入れるが見えない」を残す選択肢がある限り 要PO。

### 5.3 Options（非拘束）

| Option | Pros | Cons | Who decides |
|:---|:---|:---|:---|
| **A** — sidebar を `ResourceLstepAnalytics` に合わせる（route 正） | 見える人 = 入れる人。誤表示解消 | HospitalSettings のみ持つ運用者がタグ導線を失う | PO + RBAC owner |
| **B** — route guard を `ResourceHospitalSettings` に合わせる（sidebar 正） | 設定系メニューと一貫 | 分析権限とタグ操作の境界が変わる。BE API 権限との再整合が必要になり得る | PO + BE/FE |
| **C** — 二重条件（両方必須）または専用 resource を定義 | 最小権限を厳密化できる | 権限マスタ追加・移行コスト | PO + security |

### 5.4 Recommended for discussion only（非拘束）

**A**（route を正とし sidebar を合わせる）を議論起点にする — deep audit の「将来 BUG 候補」は **表示と guard の不一致**であり、まず一致させる。タグが「分析」か「医院設定」かの業務ラベルは USER が決める。BE permission との突合は採用後 TASK。

---

## 6. Ops one-liners（R-02 / R-04 / R-08）

| ID | Note（一行） |
|:---|:---|
| **R-02** | 本番 webhook 署名疎通・`line_bot_user_id` プロビジョニングは docs 非完結 → **USER 専権 ops**（本ブリーフ非対象）。 |
| **R-04** | L-step Write dual-gate（`LSTEP_WRITE_API_ENABLED` + clinic `is_sync_enabled`）再有効化・実送信検証は **ops / USER**（正本 `docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md`）。 |
| **R-08** | pet-health LIFF は `VITE_LIFF_ID`、line-reserve は clinic `settings.liff_id` → **deploy で ID 一致を保証**する ops residual。 |

---

## 7. 推奨まとめ（すべて 非拘束・議論用）

| ID | 議論用推奨 | 次アクション（採用後） |
|:---|:---|:---|
| R-01 | **A** 最小イベント契約表を architecture に honesty 記載 | docs PR；イベント拡大は別要件 |
| R-05 | **B** 短期 honesty 役割固定 → 中期 **A** 単一 SoT 議論 | マッピング docs → 設計 TASK（発明しない） |
| R-06 | **A** nav 掲載（漏れ）/ または **B** intentional-hide を明示 | Wave1 実装 or docs honesty |
| R-07 | **A** sidebar を route 権限に合わせる | RBAC FE 修正 + 必要なら BE 突合 |
| R-02/04/08 | （推奨なし）ops のまま | USER 運用 |

---

## 8. Explicit non-binding clause

**この文書は USER が採用するまで binding PO ではない。**  
R-01 の webhook イベント方針、R-05 の credential dual-store source-of-truth を、本ブリーフの推奨文だけを根拠に製品法・実装必須として扱ってはならない。  
採用時は決裁ログ（誰が・いつ・どの Option）を別レポートまたは todo に残すこと。

---

## 9. Sources

- Primary: `reports/2026-07-31-task-019-line-deep-audit.md` §4（R-01, R-02, R-04, R-05, R-06, R-07, R-08）  
- Cross: `todo.md` TASK-019 residual list; `docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md`（R-04）  
- Spot-check (status only, not re-audit): `frontend/src/config/paths.ts`（deliveryMonitor キー無し）; `sidebar-menu.tsx` L114–126; `settings-routes.tsx` tags guard; `operations-routes.tsx` delivery-monitor
