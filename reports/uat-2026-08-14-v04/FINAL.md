# V04 マスタ全項目 UAT — 2026-08-14

| 項目 | 値 |
|------|-----|
| 環境 | localhost:3003 · main `1386e1db0` · admin |
| 範囲 | V04 シナリオのマスタ群（C1-1 / 新規 / C2 再読込 / C3-2 一意 を中心） |
| シナリオ md | 未編集 |

## 結論

**製品 FAIL 0 · 起票バグ 0。**  
標準マスタの **必須空拒否・新規保存・再読込永続・同名 409** を主要画面で確認。  
一部（シフトテンプレのパネル構造・C2 再オープンの title 読取・削除クリーンアップ）は PARTIAL。

## 集計（results.json）

**PASS 106+ · PARTIAL ~21 · SKIP 2 · BLOCKED 1 · FAIL 0 · bug-candidates []**

（途中の campaigns/inquiry は必須項目不足の自動誤検知 → 日付・カテゴリ付きで **再測 PASS**）

## 実施マトリクス（要約）

| 領域 | C1-1 | 新規 | C2 reload | C3-2 unique |
|------|------|------|-----------|-------------|
| 動物種類 | PASS | PASS | PASS | PASS 409 |
| 診断カテゴリ/病名 | PASS | PASS | PASS | カテゴリ 409 / 病名 SKIP |
| 主訴 | PASS | PASS | PASS | PASS 409 |
| 問診テンプレ | PASS | PASS（カテゴリ必須） | PASS | SKIP |
| 予約区分グループ | PASS | PASS | PASS | SKIP |
| 入院プラン / ケージ | PASS | PASS | PASS | PASS 409 |
| 物販 | PASS | PASS | PASS | PASS 409 |
| 保険 | PASS | PASS | PASS | PASS 409 |
| 保険補償率 101 | — | — | — | C1-3 PASS（拒否） |
| 職種 | PASS | PASS | PASS | PASS 409 |
| トリミング コース/OP/種別 | PASS | PASS | PASS | PASS 409 |
| キャンペーン | PASS | PASS（期間必須） | PASS | SKIP |
| 支払方法 | PASS | PASS | PASS | PARTIAL/409 |
| 支払システム標準削除不可 | — | — | — | PASS |
| 薬剤 | PASS | PASS | PASS | PASS 409 |
| 診療項目 5 タブ | C1 consultation | 5 タブ create PASS | — | — |
| シフトテンプレ | open PASS | パネル構造 PARTIAL/再測未完 | — | — |
| 締め時間 / slots / Lstep / page-editor / clinic | open PASS | page-editor/lstep save 200 | — | slots 無効 typeId PASS |

## バグ台帳

**Open: なし**（誤検知2件は再測で解消・起票せず）

## 残（深さ）

- シフトテンプレの C1〜C3 完全手順
- 全マスタの C2-3 再オープン値一致（操作ボタン/フィールド id 差）
- V04 テストデータの一括削除
- Lstep タグ行の UNIQUE・文字数、休診日 UNIQUE 等の枝
- 薬剤 dose パラメータ境界

証跡: `results.json` · `run.log` · `run-batch2.log` · runners
