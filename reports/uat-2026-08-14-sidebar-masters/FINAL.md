# サイドバー「マスタ設定」UAT（再実施）— 2026-08-14

| 項目 | 値 |
|------|-----|
| 環境 | localhost:3003 · main `1386e1db0` · admin |
| 対象 | サイドバー **マスタ設定** 全ページ + ネストタブ + フォーム項目 |
| 深度 | **C1-1 / 項目入力して新規 / C2 再読込+再オープン / C3-2 一意** |
| 実行 | batch1 + batch2 fresh · `run-all-sidebar.mjs` |
| シナリオ md | 未編集 |
| merge/push | なし |

## 結論

**【完了】** サイドバーマスタの UAT（各フォーム項目込み）を再実施完了。  
**製品 FAIL 0 · 新規バグ 0 · bug.md Open なし。**

## 集計（results.json）

| 判定 | 件数 |
|------|------|
| PASS | 161 |
| PARTIAL | 2 |
| SKIP | 4 |
| FAIL | **0** |
| bug-candidates | **[]** |

（ENV/hub 重複統合後 · ステップ数 ~165 前後）

## ページ結果

| サイドバー | ルート | C1 | 新規 | C2 | C3-2 |
|------------|--------|----|------|----|------|
| 医院 | `/settings/clinic` | — | インボイス保存 200 | fields | — |
| 動物種類 | animal-species | PASS | PASS | PASS+reopen | 409 |
| 診療項目×5 | treatment-items tabs | PASS | PASS | PASS+reopen | 409 |
| 診断病名 | type / name | PASS | PASS | PASS+reopen | type 409 / name SKIP |
| 問診設定 | inquiry · chief · interview/templates | PASS | PASS | PASS+reopen | 主訴 409 / 他 SKIP |
| 薬剤 | medicine | PASS | PASS（単価） | PASS+reopen | 409 |
| 予約区分 | reservation-type | PASS | PASS | PASS+reopen | 409 |
| 入院・ケージ | hospitalization · cage | PASS | PASS | PASS+reopen | 409 |
| トリミング×3 | course/option/course-type | PASS | PASS | PASS+reopen | 409 |
| 権限グループ | permission-groups | PASS | PASS | PASS+reopen | 409 |
| 職種 | occupations | PASS | PASS | PASS+reopen | 409 |
| スタッフ | staff | C1 PASS | パネル項目一覧 PASS（実招待なし） | — | — |
| 保険 | insurance | PASS | PASS | PASS+reopen | 409 · 101拒否 PASS |
| 物販 | merchandise | PASS | PASS | PASS+reopen | 409 |
| 支払方法 | payment-methods | PASS | PASS | PASS+reopen | PARTIAL noresp · 標準削除不可 PASS |
| 締め時間 | closing-time | — | 保存 200 · time×3 | holiday-C1 PARTIAL | — |
| （関連）キャンペーン | campaigns | PASS | PASS（期間） | PASS+reopen | SKIP |
| （関連）シフト | shift-templates | PASS disabled | PASS 全項目 | PASS+memo | 409 |

## バグ

**なし。** bug.md / todo.md §2 とも Open 空。

## PARTIAL（製品欠陥ではない）

1. **payment-methods C3-2** — 重複 POST のレスポンス未捕捉（create/C2 は PASS）
2. **closing-time holiday-C1** — 空日付追加の UI 枝が浅い

## 証跡

- `results-batch1.json` · `results-batch2.json` · `results.json`
- `run-batch1-fresh.log` · `run-batch2-fresh.log`
- `run-all-sidebar.mjs`
