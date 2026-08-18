# タスク台帳 — Linear へ移行済み

| 項目 | 値 |
|------|-----|
| **移行日** | 2026-08-14 |
| **実行 SoT** | Linear Team **Baritech** · Project **ノア動物病院電子カルテ** · hub **[BRT-4](https://linear.app/baritechllc/issue/BRT-4)** |
| **マップ（docs Open）** | [`reports/todo-walk-2026-08-14/todo-docs-linear-map.md`](reports/todo-walk-2026-08-14/todo-docs-linear-map.md) |
| **マップ（GH Open）** | [`reports/todo-walk-2026-08-14/github-linear-map.md`](reports/todo-walk-2026-08-14/github-linear-map.md) |
| **旧本文** | git 履歴（本ファイルのフル台帳版） |

## 使い方

- 作業の状態・担当・次の一手 → **Linear Issue**
- 製品バグの新規 → Linear に起票（必要なら GitHub も）。旧 §2 形式は使わない
- 受入 UAT 証跡 → `reports/uat-*` · 人手 SESSION は `reports/uat-human-*`
- 開発規約 → [`.claude/CLAUDE.md`](.claude/CLAUDE.md) · [`AGENTS.md`](AGENTS.md)

## 代表チケット

| 領域 | Linear |
|------|--------|
| GH Open 調査束 | BRT-37〜52 |
| #299 / PO-008 / presence / H1–H7 / P1 / M1–M5 / OPS-13 | BRT-55〜67 |

**agent 製品 unit: 新規に増やさない（NONE 維持の方針は Linear 説明に記載）。**

## 城東 検査機器連携（2026-08-18 疎通後）

old_db の現場・仕様は `../old_db/todo.md` の **JOU-LAB-0** と `../old_db/docs/lab-go/go-impl/`。こちらは AnimalEkarte 実装だけ書く。

- [ ] **AE-LAB-1** 3電文を `LabInboundBatch` にデコードする（`fuji_nx600` / `fuji_au10v` / `arkray_pu4010`）。ペット紐付けしない。仕様: `../old_db/docs/lab-go/go-impl/device-serial-adapter.md`。キャプチャ: `../old_db/docs/lab-go/hospital-field-pack/captures/2026-08-18-jouto/`。
- [ ] **AE-LAB-2** 検査機器マスタ。`../old_db/docs/lab-go/go-impl/device-item-master.csv` の25行を初期投入。`exam_type_field_id` は医院が設定。未知コードは `needs_review`。
- [ ] **AE-LAB-3** 未紐付け受信ジョブをスタッフがペットに付けてから persist する。
- [ ] **AE-LAB-4** 3種の `source_type` を、マスタ参照かつペット選択後だけ commit 可能にする。`drwan` は開けない。

順: 1 → 2 → 3 → 4。IDEXX は old_db の **JOU-LAB-X**。
