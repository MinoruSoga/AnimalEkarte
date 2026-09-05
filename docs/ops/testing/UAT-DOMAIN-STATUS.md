# UAT ドメイン別ステータス

> **目的**: 受け入れ結果をシナリオ ID だけでなく業務ドメイン単位で俯瞰する。
> **正本リンク**: [scenarios/README.md](./scenarios/README.md) · [TEST_ARCHITECTURE.md](./TEST_ARCHITECTURE.md)
> **最新ラン**: `reports/uat-2026-09-05-r3/`（r3 / BUG-004/005 修正後フル再実行）
> **更新日**: 2026-09-05

## サマリ（シナリオ S01–S13 + V01–V05）

| Status | Count |
|:---|---:|
| PASS | 15 |
| FAIL | 0 |
| PARTIAL | 3 |
| BLOCKED | 0 |

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` @ `15796fff7` |
| 環境 | local（FE :3003 / BE :8080） |
| 証跡 | [`reports/uat-2026-09-05-r3/`](../../../reports/uat-2026-09-05-r3/) |

---

## 臨床安全（S01, S02, S03, S06 + V01）

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` |
| 環境 | local |
| ドメイン総合判定 | **PARTIAL** |

| シナリオ | status |
|:---|:---|
| S01 | PARTIAL |
| S02 | PASS |
| S03 | PASS |
| S06 | PASS |
| V01 | PASS |

- **未解消ギャップ**: S01 の外部 LSTEP タグ削除/再同期は local で `is_configured=false` のため未検証（env）。
- **関連 bug IDs**: （なし）
- **証跡**: `reports/uat-2026-09-05-r3/`（deepen-partials / deepen-remaining / field-results）

---

## 入院（S05）

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` |
| 環境 | local |
| ドメイン総合判定 | **PASS** |

| シナリオ | status |
|:---|:---|
| S05 | PASS |

- **未解消ギャップ**: ボード UI トグルは soft（API 入院サイクルは PASS）。
- **関連 bug IDs**: （なし）
- **証跡**: `reports/uat-2026-09-05-r3/deepen-partials.json`

---

## 会計 / トリミング精算（S07, S08, S09, S11 + V02）

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` |
| 環境 | local |
| ドメイン総合判定 | **PARTIAL** |

| シナリオ | status |
|:---|:---|
| S07 | PASS |
| S08 | PASS |
| S09 | PARTIAL |
| S11 | PASS |
| V02 | PASS |

- **未解消ギャップ**: S09 の timed `completed_at` 5-fixture 帰属は承認済み helper 待ち（SQL/clock 禁止）。
- **関連 bug IDs**: （なし）
- **証跡**: `reports/uat-2026-09-05-r3/`（resume-exec / fixup / field-results）

---

## 予約 / LIFF顧客体験（S04, S12 + V05 LIFF）

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` |
| 環境 | local |
| ドメイン総合判定 | **PARTIAL** |

| シナリオ | status |
|:---|:---|
| S04 | PASS |
| S12 | PARTIAL |
| V05（LIFF 部分） | PASS（empty-token） |

- **未解消ギャップ**: 実 LINE 連携レーンなし。no-token は mock LIFF で飼主画面（設計上 HealthCardApp）。
- **関連 bug IDs**: （BUG-002 は復活せず。empty-token は修正確認済）
- **証跡**: `reports/uat-2026-09-05-r3/unblock-s04-s09-s13.json` / screenshots `S12-*` `BUG002-*`

---

## 経営集計（S10）

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` |
| 環境 | local |
| ドメイン総合判定 | **PASS** |

| シナリオ | status |
|:---|:---|
| S10 | PASS |

- **未解消ギャップ**: CSV ダウンロードが 1 回 timeout（コア LTV 整合は PASS）。
- **関連 bug IDs**: （なし）
- **証跡**: `reports/uat-2026-09-05-r3/deepen-remaining.json`

---

## 顧客・組織 / identity-links（S13 + V03）

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` |
| 環境 | local |
| ドメイン総合判定 | **PASS** |

| シナリオ | status |
|:---|:---|
| S13 | PASS |
| V03 | PASS |

- **未解消ギャップ**: view-only / 非カバーアクターの異常系は本レーンにアカウントなし（soft）。staff create UI は執行ロールで BLOCKED soft。
- **関連 bug IDs**: （なし）
- **証跡**: `reports/uat-2026-09-05-r3/fixup-s08-s11-s13-v05.json` / field-results

---

## マスタ・医院設定（V04 + closing / hospital settings）

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` |
| 環境 | local |
| ドメイン総合判定 | **PASS** |

| シナリオ | status |
|:---|:---|
| V04 | PASS |
| closing-settings（S09 前提） | PASS（GET/PATCH） |

- **未解消ギャップ**: （なし。旧 BUG-004/005 は `15796fff7` で解消確認）
  - LINE予約設定 PUT で `closed_weekdays` 省略 → **200** `closed_weekdays=[]`
  - URL `:clinic_id` 未割当（clinics/3）→ **403** `not assigned to this clinic`
- **関連 bug IDs**: （旧 BUG-20260905-004 / 005 解消 — bug.md 空）
- **証跡**: `reports/uat-2026-09-05-r3/precheck.json` / `precheck-bug003.json`

---

## 認証（V05 auth）

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` |
| 環境 | local |
| ドメイン総合判定 | **PASS** |

| シナリオ | status |
|:---|:---|
| V05 | PASS |

- **未解消ギャップ**: （なし。BUG-001 修正確認済）
- **関連 bug IDs**: （旧 BUG-001 解消）
- **証跡**: `reports/uat-2026-09-05-r3/precheck.json` / screenshots `BUG001-*` `V05-*`

---

## メンテ手順（短）

1. 受け入れ再実行後、本ファイルの実施日・ブランチ・各ドメイン表の status / ギャップ / bug ID を更新する。
2. 製品 FAIL のみルート `bug.md` へ（PARTIAL/BLOCKED は書かない）。
3. 証跡は `reports/uat-YYYY-MM-DD(-postfix|-rN)/` に置き、シナリオ md は編集しない。
