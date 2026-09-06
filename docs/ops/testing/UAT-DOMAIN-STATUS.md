# UAT ドメイン別ステータス

> **目的**: 受け入れ結果をシナリオ ID だけでなく業務ドメイン単位で俯瞰する。
> **正本リンク**: [scenarios/README.md](./scenarios/README.md) · [TEST_ARCHITECTURE.md](./TEST_ARCHITECTURE.md)
> **更新日**: 2026-09-06
> **照合**: `QA-UAT-EVIDENCE-SYNC`。`reports/` は gitignore。コミット済み正本は本ファイルと [`bug.md`](../../../bug.md)。Linear は未照会（UNKNOWN）。

最終実行スナップショット（2026-09-05 / `uat/20260905` @ `2cbd8d9ad` / local FE :3003 · BE :8080）:

| Status | Count | 内訳 |
|:---|---:|:---|
| PASS | 15 | S02 S03 S04 S05 S06 S07 S08 S10 S11 S12 S13 · V01 V02 V03 V05 |
| FAIL | 1 | V04（最終実行時。主訴 DELETE。下記「現行」を見よ） |
| PARTIAL | 1 | S01 |
| BLOCKED | 1 | S09 |

現行（2026-09-06 照合。再実行していない判定は推定で動かさない）:

| 項目 | 値 |
|:---|:---|
| 開いている製品 FAIL（`bug.md`） | **0** |
| V04 受入 | **UNKNOWN**（全体 PASS ではない。下記 2026-09-06 再実行） |
| S09 | **BLOCKED**（package helper あり。HTTP/CLI とブラウザ再実行は未。製品 FAIL ではない） |
| S01 | **PARTIAL**（LSTEP 実送信は E1） |
| r14 | ヘッダだけ「FAIL 0 / PASS 16」と書いてあった regression smoke。V04 再実行の証跡は本ファイルに無く、PASS 翻転ではない。ディレクトリは gitignore のため再読不可 |

下記 `reports/uat-*` は gitignore 対象で fresh clone には配布されない参照である。今回の文書更新では原証跡を再確認できていない。不在から未実施/完了を推定せず、再検証には管理者が保持する原証跡を使う。

各 PASS は下記の過去実行スナップショットの転記であり、今回のソース照合による再認定ではない。特に V 系の scenario-level PASS と、未収録 field/wildcard を含む inventory 全体の完了は別（[FORM-FIELD-INVENTORY.md](scenarios/FORM-FIELD-INVENTORY.md)）。

冒頭の FAIL 0 は「開いている製品 FAIL が 0」であり、最終実行の V04 FAIL を消したものではない。PASS 16 は V04 を PASS にした数え方なので使わない。

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
  3. 現行 package helper は 5 会計ヘッダを作るが、HTTP/CLI・UAT identity・支払内訳・cleanup が未接続（[設計と実装境界](S09-FIXTURE-DESIGN.md)）。帰属証明ステップ #2–#6 は未再実行 → シナリオ総合 **BLOCKED**（settings/preview/history の先行 PASS では解除しない）。
  4. 詳細: [`reports/uat-2026-09-05-r5/S09-BLOCKED.md`](../../../reports/uat-2026-09-05-r5/S09-BLOCKED.md)
- **関連 bug IDs**: （なし — UAT 接続の不足は製品 FAIL ではない）
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
| ドメイン総合判定 | **UNKNOWN**（V04 全体は未再実行。主訴 DELETE の 500 回帰は testdb で非再現） |

| シナリオ | status |
|:---|:---|
| V04 | **UNKNOWN**（2026-09-05 最終実行は FAIL。2026-09-06 は DELETE 回帰のみ再実行） |
| closing-settings（S09 前提） | PASS（r6 GET/PATCH roundtrip） |

- **Master CRUD 総合（r6+r7・最終実行）**: PASS **26** / PARTIAL **0** / BLOCKED **0** / FAIL **1**（27 行）
  - **最終実行 FAIL**: `master-chief-complaint` DELETE が `inquiries.deleted_at` 参照で 500 → 当時 BUG-20260905-001
  - **2026-09-06 testdb**: `TestChiefComplaintTypeRepository_Delete` と `CountUsage` が GREEN。未使用区分は削除できる。参照中は Conflict。`inquiries.deleted_at` を見ない。500 回帰は非再現
  - **2026-09-06 live HTTP**: 合成 catalog login は 200。`master-medical` create は 403（一般グループ）。権限を上げて clinic 1/2 を触っていない。HTTP DELETE の受入は **BLOCKED**
  - **判定**: 当時の製品 FAIL を bug.md に戻さない。V04 全体は PASS にしない
  - 診断・診療項目5タブ・薬剤・トリミング一式・支払方法・締め・請求書欄など他は r6+r7 CRUD PASS
- **関連 bug IDs**: 現行 open なし
- **証跡**: testdb コマンドは `todo.md` 順 2。live は status code のみ（credential・行値なし）
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
5. V04 は主訴 DELETE を disposable clinic で再実行してから UNKNOWN を外す。コード修正だけで PASS にしない。
