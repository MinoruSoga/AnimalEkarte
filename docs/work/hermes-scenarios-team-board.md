# Hermes エージェントチーム — 受入 scenarios 実装突合ボード

| 項目 | 値 |
|------|-----|
| **日付** | 2026-08-07 |
| **HEAD（突合時点）** | シナリオ更新コミット参照 `git log -1 -- docs/ops/testing/scenarios` |
| **方式** | team-agent-orchestration（並列 3 lane）+ 親セッション統合 |
| **prompt-craft-graph** | 下記「PCG 境界」参照。**本更新の書き込みは graph 外**（PCG documentation スキルは read-only / 編集不可） |

## 完了（Merged）

| Lane | Owner | 対象 | 状態 |
|------|-------|------|------|
| clinical | Agent-Audit-S-Clinical | S01–S06 | **merged** — シナリオ md 更新済み |
| acct-liff | Agent-Audit-S-Acct-LIFF | S07–S13 | **merged** — 同上 |
| forms | Agent-Audit-V-Forms | V01–V05 + README | **merged** — 同上 |

## 主な DRIFT 修正（実装に合わせてシナリオを更新）

| 領域 | シナリオ側の最新化 |
|------|-------------------|
| 死亡登録 | status PATCH ではなく `PetDeceasedDialog` / death API |
| 検査 H/L | 表示は **HIGH/LOW** バッジ；`is_assessed` 依存を明記 |
| ワクチン | interval 値・既定 1 年・lot 最大 4；種フィルタ無しは BUG 注記 |
| LIFF 予約 | フロー順・フィルタラベル「LINE予約」・両側 mock |
| 入院退院 | 会計 ON のみ discharge-with-billing；チェック default OFF |
| カルテ確定 | メッセージ全文・削除 UI と BE 拒否の差 |
| 見積 | seed に draft あり・バッジ表記・audit best-effort |
| 会計 | `POST /accountings/complete`・部分入金 BLOCKED |
| 顧客集計 | seed 実態・BE 20s / FE 25s timeout |
| 未請求 | `unbilled-details` |
| ペットヘルス | `health-card` API・mock 挙動 |
| フォーム | 保険/値引 0–100、権限マトリクス、BUG-031 login redirect |
| README | **S13** を索引に追加 |

## Hermes 向け — 残ランタイム検証（要実測）

シナリオ本文に残した **【要実測】/ BLOCKED / DEFER** を Hermes browser で潰す場合のカード:

| ID | 内容 | 前提 |
|----|------|------|
| RT-S02 | H/L が seed ranges で発火するか | stack up + ranges 適用済み |
| RT-S04 | mock 両側で courses 200 | `LIFF_MOCK` + `VITE_LIFF_MOCK` |
| RT-S05-A2 | 退院会計 atomic 失敗ケース | 専用データ |
| RT-S08-partial | 部分入金 UI が依然 disabled か | 執行ログイン |
| RT-S11 | 会計完了→トリミング完了バッジ | 通し手順 |
| RT-S01-A1 | 未来予約あり死亡ガード | データ準備 |

**Hermes 起動例（オペレータ）:**

```text
You are Hermes browser QA for AnimalEkarte acceptance scenarios.
Workspace: AnimalEkarte main. Base URL: http://localhost:3003 API: :8080
Use LIFF mock only. Do not write VERIFIED_FIXED. Do not create reports/.
For each RT-* card above: open scenario md, execute steps that are 要実測,
append a one-line result under the scenario's 実装突合 as "runtime 2026-08-07: PASS|FAIL|BLOCKED — note".
Credentials: use E2E_LOGIN_* from host env (never paste secrets into chat logs).
```

## prompt-craft-graph 境界（重要）

| プロファイル | 用途 | 本タスクとの関係 |
|--------------|------|------------------|
| **v1 `documentation` / `research`** | 読取のみ・成果物は合成文書 | **scenario の in-place 編集は不可** |
| **v2 feature implementation** | TDD + 隔離 worktree の製品コード | シナリオ md 更新の正道ではない |
| **本セッション** | 並列 general-purpose が scenario を直接更新 | **write path はこちら** |

したがって「Hermes + prompt-craft-graph」で **シナリオ最新化そのもの**を graph 完了させることは契約上できない。  
Hermes には上の **runtime 検証カード**を渡し、PCG は将来の **製品コード unit**（例: S02 is_assessed 修正）に使う。

### 将来 PCG に載せる候補（製品・1 unit）

1. 検査 `is_assessed` / range 解決で H/L が常に未判定になる場合の fix（要再現）  
2. 確定済みカルテ一覧の削除導線を FE で隠す（S06 記載の UX 差）  
3. 見積 successors UI 配線（S07 記載の未配線）

各 unit は 1 graph = 1 feature、臨床値は載せない。

## 非目標

- 新規 reports/ レポート  
- VERIFIED_FIXED 自動付与  
- migrate / seed / force-push  
