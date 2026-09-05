# UAT ドメイン別ステータス

> **目的**: 受け入れ結果をシナリオ ID だけでなく業務ドメイン単位で俯瞰する。
> **正本リンク**: [scenarios/README.md](./scenarios/README.md) · [TEST_ARCHITECTURE.md](./TEST_ARCHITECTURE.md)
> **最新ラン**: `reports/uat-2026-09-05-r4/`（r4 / PARTIAL S01・S09・S12 前進）
> **更新日**: 2026-09-05

## サマリ（シナリオ S01–S13 + V01–V05）

| Status | Count |
|:---|---:|
| PASS | 16 |
| FAIL | 0 |
| PARTIAL | 2 |
| BLOCKED | 0 |

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` @ `b8a6b26e0` |
| 環境 | local（FE :3003 / BE :8080） |
| 証跡 | [`reports/uat-2026-09-05-r4/`](../../../reports/uat-2026-09-05-r4/) |

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

- **未解消ギャップ**: S01 の LSTEP タグ削除/再同期は、mock API キー + `line_user_id` + sync 一時 ON でもローカル観測不可（runtime に recording/mock LSTEP HTTP クライアントなし。実送信は `is_sync_enabled=false` のまま）。死亡ガード自体は PASS。
- **関連 bug IDs**: （なし）
- **証跡**: `reports/uat-2026-09-05-r4/partials-exec.json`

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
- **証跡**: `reports/uat-2026-09-05-r3/deepen-partials.json`（r3 継承）

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

- **未解消ギャップ**: S09 の timed `completed_at` 5-fixture 帰属は承認済み helper 不在のため BLOCKED（SQL/clock 禁止）。settings/preview/history は PASS。
- **関連 bug IDs**: （なし）
- **証跡**: `reports/uat-2026-09-05-r4/partials-exec.json` / `s09-helper-search.json`

---

## 予約 / LIFF顧客体験（S04, S12 + V05 LIFF）

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` |
| 環境 | local |
| ドメイン総合判定 | **PASS** |

| シナリオ | status |
|:---|:---|
| S04 | PASS |
| S12 | **PASS**（r4 mock lane） |
| V05（LIFF 部分） | PASS（empty-token） |

- **未解消ギャップ**: 実 LINE idToken 検証による BE link / 409 再連携 / 期限切れ 400 は mock 外。FE `VITE_LIFF_MOCK` は success-only（仕様どおり病院側 `line_user_id` は未連携のまま）。隔離は staff `link-owner` + LIFF_MOCK health-card で証明。
- **関連 bug IDs**: （BUG-002 は復活せず）
- **証跡**: `reports/uat-2026-09-05-r4/s12-fix.json` / screenshots `S12-*`

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

- **未解消ギャップ**: CSV ダウンロードが 1 回 timeout（コア LTV 整合は PASS）— r3 継承。
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
- **証跡**: `reports/uat-2026-09-05-r3/fixup-s08-s11-s13-v05.json`

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
| closing-settings（S09 前提） | PASS（GET/PATCH; r4 で GET 再確認） |

- **未解消ギャップ**: （なし。旧 BUG-004/005 は解消確認済）
- **関連 bug IDs**: （旧 BUG-20260905-004 / 005 解消 — bug.md 空）
- **証跡**: r3 precheck + r4 `s12-fix.json`（S09-1-settings）

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
- **証跡**: `reports/uat-2026-09-05-r3/precheck.json`

---

## メンテ手順（短）

1. 受け入れ再実行後、本ファイルの実施日・ブランチ・各ドメイン表の status / ギャップ / bug ID を更新する。
2. 製品 FAIL のみルート `bug.md` へ（PARTIAL/BLOCKED は書かない）。
3. 証跡は `reports/uat-YYYY-MM-DD(-postfix|-rN)/` に置き、シナリオ md は編集しない。
