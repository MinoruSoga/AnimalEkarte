# AnimalEkarte UAT — ゼロから再実施 2026-08-14

| 項目 | 値 |
|------|-----|
| 環境 | localhost:3003/:8080 · main `1386e1db0` · LIFF mock ON |
| 前提 | 旧 `reports/uat-*` **全削除後**に新規実施 |
| 範囲 | S01〜S13 + V01〜V05 |
| シナリオ md | 未編集 |
| merge/push/migrate/Done | なし |

## 結論

**製品 FAIL 0 · 新規バグ 0。**  
Open バグ台帳への追記なし。

## 集計

**PASS 96 · PARTIAL 4 · BLOCKED 5 · FAIL 0 · bug-candidates []**

## シナリオ

| ID | 判定 | 要点 |
|----|------|------|
| S01 | **PASS** | 死亡→選択不可×3→復活 |
| S02 | **PASS** | HIGH/LOW · 完了ロック |
| S03 | **PASS** | ワクチン 201 |
| S04 | **PASS**（mock） | `R-20260816-0002`→キャンセル · 実通知 BLOCKED |
| S05 | **PASS** | 入院201·退院200 |
| S06 | **PASS**（中核） | ロック·追記 · audit BLOCKED |
| S07 | **PASS** | 見積201 |
| S08 | **PASS** + 仕様BLOCKED | 部分入金 |
| S09 | **PASS** + fixture BLOCKED | 締めUI |
| S10 | **PASS** | 集計 |
| S11 | **PASS**（中核） | unbilled · create PARTIAL |
| S12 | **PASS**（mock） | 実 token BLOCKED |
| S13 | **PASS**（中核） | workbench · 2医院 link PARTIAL |
| V01–V05 | **PASS** | 到達 · V04 C1 一部 PARTIAL |
| LOCK | **PASS** | 033/035/038 |

## バグ

**Open: なし**（bug.md / todo.md §2）

## 人間レーン

todo-po.md: 実 LINE · staging · fixture · audit DB
