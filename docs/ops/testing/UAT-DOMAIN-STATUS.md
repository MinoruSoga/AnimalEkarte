# UAT ドメイン別ステータス

> **目的**: 受け入れ結果をシナリオ ID だけでなく業務ドメイン単位で俯瞰する。
> **正本リンク**: [scenarios/README.md](./scenarios/README.md) · [TEST_ARCHITECTURE.md](./TEST_ARCHITECTURE.md)
> **最新ラン**: `reports/uat-2026-09-05-r13/`（r13 V05-17/field inventory）· r12
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

- **未解消ギャップ**: S01 死亡/復活/会計ガードは PASS。r8 で sync ON + mock key + line_user_id 時に BE ログで `lstep: write API disabled by deploy gate`（`LSTEP_WRITE_API_ENABLED` unset）と sync 失敗を観測。タグ配列は変化なし。staff audit API なし（DB audit は USER 実施）。PASS には recording mock または write 有効な同期レーンが必要。
- **関連 bug IDs**: （なし）
- **証跡**: [`reports/uat-2026-09-05-r8/FINAL-s01.md`](../../../reports/uat-2026-09-05-r8/FINAL-s01.md)

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

- **未解消ギャップ**: （なし。r10 で Board/List UI トグル + admit/discharge API 再確認 PASS）
- **関連 bug IDs**: （なし）
- **証跡**: [`reports/uat-2026-09-05-r10/FINAL-soft-gaps.md`](../../../reports/uat-2026-09-05-r10/FINAL-soft-gaps.md)

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

- **未解消ギャップ**: （なし。r10 で CSV ダウンロード PASS — 旧 flake は行未選択でボタン disabled）
- **関連 bug IDs**: （なし）
- **証跡**: [`reports/uat-2026-09-05-r10/FINAL-soft-gaps.md`](../../../reports/uat-2026-09-05-r10/FINAL-soft-gaps.md)

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

- **未解消ギャップ**: view-only / 非カバーアクター異常系はアカウントなし（soft）。staff create は r10 で `/settings/staff` 経由 PASS（旧 soft-BLOCK は誤パス `/settings/staffs`）。
- **関連 bug IDs**: （なし）
- **証跡**: [`reports/uat-2026-09-05-r10/FINAL-soft-gaps.md`](../../../reports/uat-2026-09-05-r10/FINAL-soft-gaps.md)


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
## 認証・LINE / LSTEP（V05）

| 項目 | 値 |
|:---|:---|
| 実施日 | 2026-09-05 |
| ブランチ | `uat/20260905` |
| 環境 | local |
| ドメイン総合判定 | **PASS** |

| シナリオ | status |
|:---|:---|
| V05 | **PASS**（r9 + r11 deepen） |

- **r9**: auth / LINE settings+page editor / LSTEP settings+connection-test fail-closed
- **r11**: V05-13 trigger priority PASS; V05-14 tag-code mappings PASS; V05-16 CSV PASS; V05-18 checkup-sync PASS; V05-6/7 LIFF_MOCK reserve create/cancel PASS; V05-10 slots PASS; V05-11 owner⇄LINE link PASS
- **未解消ギャップ（BLOCKED・製品 FAIL ではない）**: V05-17 の LSTEP 実送信 remove（要 sync ON + `LSTEP_WRITE_API_ENABLED` + 到達可能 LSTEP）。ローカル経路（空選択/unlinked 404/validation/idempotent DELETE）は r13 PASS。実 LINE / UAT-254 close は USER。V05-15 は r12 PASS（elevate 復元済）
- **関連 bug IDs**: （なし。r11 新規 FAIL なし）
- **証跡**: [`reports/uat-2026-09-05-r13/FINAL.md`](../../../reports/uat-2026-09-05-r13/FINAL.md) · [`reports/uat-2026-09-05-r12/FINAL-v05-15.md`](../../../reports/uat-2026-09-05-r12/FINAL-v05-15.md) · [`reports/uat-2026-09-05-r11/FINAL-v05-subforms.md`](../../../reports/uat-2026-09-05-r11/FINAL-v05-subforms.md) · [`reports/uat-2026-09-05-r9/FINAL-v05.md`](../../../reports/uat-2026-09-05-r9/FINAL-v05.md)

---

## メンテ手順（短）

1. 受け入れ再実行後、本ファイルの実施日・ブランチ・各ドメイン表の status / ギャップ / bug ID を更新する。
2. 製品 FAIL のみルート `bug.md` へ（PARTIAL/BLOCKED は書かない）。
3. 証跡は `reports/uat-YYYY-MM-DD(-postfix|-rN)/` に置き、シナリオ md は編集しない。
4. S09 解除時は承認済み helper マージ後に #2–#6 を再実行し、本ファイルの会計ドメインとサマリを更新する。
