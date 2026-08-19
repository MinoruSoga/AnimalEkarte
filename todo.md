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
| 城東 検査機器連携 | [BRT-94](https://linear.app/baritechllc/issue/BRT-94) · 0=[BRT-100](https://linear.app/baritechllc/issue/BRT-100) · 1〜4 は BRT-95〜98 |

**agent 製品 unit: 新規に増やさない（NONE 維持の方針は Linear 説明に記載）。**

## 城東 検査機器連携（2026-08-18 疎通後）

old_db の現場・仕様は `../old_db/todo.md` の **JOU-LAB-0** と `../old_db/docs/lab-go/go-impl/`。こちらは AnimalEkarte 実装だけ書く。

方針（2026-08-19 USER + Fable UX YES-WITH-FIXES）: **ファイルアップロードしない。** 検査用 Mac の待機ページが有線シリアルを読む（掲示板。開きっぱなし）。常駐アプリは置かない。UI は **1画面**（ペット先待機は最適化、未紐付け欄は一級）。保存は即 persist + インライン取り消す。確認ダイアログ禁止。公式リカバリは機器の送信再押下（指紋 duplicate）。所見: `../old_db/docs/lab-go/go-impl/REVIEW-FABLE-2026-08-19-AE-LAB-UX.md`。

- [x] **AE-LAB-0** [BRT-100](https://linear.app/baritechllc/issue/BRT-100) 設計: [ADR-007](docs/architecture/adr/007-lab-device-receive-and-commit.md)。コード済み（ADR Accepted）。Done は人間ゲート。
- [x] **AE-LAB-1** [BRT-95](https://linear.app/baritechllc/issue/BRT-95) `DecodeLabDeviceFrames` + 合成バイトテスト。DB/enum は未使用。コード済み。Done は人間ゲート。
- [x] **AE-LAB-2** [BRT-96](https://linear.app/baritechllc/issue/BRT-96) 検査機器マスタ。未対応チップから該当行へ。日常経路にマスタを出さない。コード済み。Done は人間ゲート。DDL は `001_init.sql` セクション13。既存 DB は `make reset`。
- [x] **AE-LAB-3** [BRT-97](https://linear.app/baritechllc/issue/BRT-97) 1画面: 待機（大表示）+ 未紐付け欄 + 保存カード［取り消す］。診察端末の検査画面から1クリックで後付け。コード済み。Done は人間ゲート。DDL は `001_init.sql` セクション13。既存 DB は `make reset`。
- [x] **AE-LAB-4** [BRT-98](https://linear.app/baritechllc/issue/BRT-98) 3種をマスタ参照かつペット確定後だけ exam persist。`drwan` は開けない。コード済み。Done は人間ゲート。DDL は `001_init.sql` セクション13。既存 DB は `make reset`。

順: **0 → 1 → 2 → 3 → 4**。0 なしに 1 のコードを始めない。IDEXX は old_db の **JOU-LAB-X**。
