# UAT ドメイン別ステータス

> **目的**: 受け入れ結果をシナリオ ID だけでなく業務ドメイン単位で俯瞰する。
> **正本リンク**: [scenarios/README.md](./scenarios/README.md) · [TEST_ARCHITECTURE.md](./TEST_ARCHITECTURE.md)
> **最新ラン**: `reports/uat-2026-09-05-r7/`（r7 gap-clear: lab-device・動物種類 PASS）· V04 は主訴 DELETE FAIL 継続
> **更新日**: 2026-09-05

## サマリ（シナリオ S01–S13 + V01–V05）

| Status | Count |
|:---|---:|
| PASS | 15 |
| FAIL | 1 |
| PARTIAL | 1 |
| BLOCKED | 1 |

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` @ `7a216ee66` |
| 環境 | local（FE :3003 / BE :8080） |
| 証跡 | [`reports/uat-2026-09-05-r7/`](../../../reports/uat-2026-09-05-r7/) · r6 CRUD · r5 シナリオ |

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

- **未解消ギャップ**: S01 の LSTEP タグ削除/再同期はローカルで観測不可。`ValidateLstepBaseURL` が `api.lstep.jp` 以外（loopback 含む）を拒否するため recording mock を base URL に向けられない。runtime に recording/mock LSTEP HTTP クライアントなし。`audit_logs` の DB 確認はシナリオ上 USER 実施。死亡/復活/会計ガードは PASS。
- **関連 bug IDs**: （なし）
- **証跡**: `reports/uat-2026-09-05-r4/partials-exec.json` · r5 `FINAL.md`

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
| ドメイン総合判定 | **BLOCKED** |

| シナリオ | status |
|:---|:---|
| S07 | PASS |
| S08 | PASS |
| S09 | **BLOCKED** |
| S11 | PASS |
| V02 | PASS |

- **未解消ギャップ（S09 BLOCKED 要件）**:
  1. 承認済み **fixture API** または **scoped UAT test helper** が必要（`completed_at` を 10:00 / 13:30 / 14:00 / 20:00 / 翌 02:00 に設定した合成会計 5 件）。
  2. **禁止**: 直接 DB 更新、システム時計変更、既存会計の改変（シナリオ hard rule）。
  3. helper 不在のため帰属証明ステップ #2–#6 は実施不可 → シナリオ総合 **BLOCKED**（settings/preview/history の先行 PASS では解除しない）。
  4. 詳細: [`reports/uat-2026-09-05-r5/S09-BLOCKED.md`](../../../reports/uat-2026-09-05-r5/S09-BLOCKED.md)
- **関連 bug IDs**: （なし — helper 欠如は製品 FAIL ではない）
- **証跡**: `reports/uat-2026-09-05-r5/S09-BLOCKED.md` · `reports/uat-2026-09-05-r4/s09-helper-search.json`

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
| ドメイン総合判定 | **FAIL**（主訴 DELETE のみ） |

| シナリオ | status |
|:---|:---|
| V04 | **FAIL**（主訴 DELETE 500 = BUG-20260905-001。他マスタ CRUD は r6+r7 で PASS） |
| closing-settings（S09 前提） | PASS（r6 GET/PATCH roundtrip） |

- **Master CRUD 総合（r6+r7）**: PASS **26** / PARTIAL **0** / BLOCKED **0** / FAIL **1**（27 行）
  - **FAIL**: `master-chief-complaint` DELETE 常時 500（`inquiries.deleted_at` 欠落）→ **BUG-20260905-001**（r7 再確認済）
  - **r7 前進**: `lab-device-item-masters` PARTIAL→**PASS**（`lab-import:create` 一時付与）。`master-animal-species` BLOCKED→**PASS**（`is_system_admin` 一時付与）。権限は試験後に復元済
  - 診断・診療項目5タブ・薬剤・トリミング一式・支払方法・締め・請求書欄など他は CRUD PASS
- **関連 bug IDs**: BUG-20260905-001（open）
- **証跡**: [`reports/uat-2026-09-05-r7/FINAL-gap-clear.md`](../../../reports/uat-2026-09-05-r7/FINAL-gap-clear.md) · [`reports/uat-2026-09-05-r6/FINAL-master-crud.md`](../../../reports/uat-2026-09-05-r6/FINAL-master-crud.md)
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
4. S09 解除時は承認済み helper マージ後に #2–#6 を再実行し、本ファイルの会計ドメインとサマリを更新する。
