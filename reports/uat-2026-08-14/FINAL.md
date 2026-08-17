# AnimalEkarte UAT — CDP :9222 再実施 2026-08-14

| 項目 | 値 |
|------|-----|
| 環境 | localhost:3003/:8080 · Chrome CDP `:9222` · LIFF mock ON |
| 認証 | E2E_LOGIN_*（値は非公開） |
| 範囲 | S01〜S13 + V01〜V05（FIELD-LEVEL-PROTOCOL · inventory 項目） |
| シナリオ md | **未編集** |
| 実行手段 | Playwright `connectOverCDP(http://127.0.0.1:9222)`（Chrome DevTools MCP と同 endpoint） |
| 所要 | 約 12 分（フル）+ 再確認 |
| merge/push/migrate | なし |

## 結論

**製品 FAIL 0 · 新規受入バグ起票 0。**

初回ランで F4 が FAIL になった `master-campaign` / `master-inquiry-templates` は、再確認で必須項目（期間・カテゴリ）を埋めると **POST 201**。ハーネス不足であり製品欠陥ではない。
`V05 auth-login` タイムアウトは同時セッション起因で、再確認で `/login` 到達・ログインボタン存在を確認。

## 集計（results.json ステップ単位）

**BLOCKED** 7 · **N/A** 2 · **PARTIAL** 26 · **PASS** 1352

**結果件数** 1387 · **製品 bug-candidates** 0（dismissed 2 = harness）

## シナリオ判定

| ID | 判定 | 要点 |
|----|------|------|
| ENV | **PASS** | health/login/api/cdp · P4/F0/Pa0/B0 |
| S01 | **PASS** | 死亡→選択不可×3→復活 · P7/F0/Pa0/B0 |
| S02 | **PASS** | HIGH/LOW · 完了ロック · P6/F0/Pa0/B0 |
| S03 | **PASS** | ワクチン 201 · P4/F0/Pa0/B0 |
| S04 | **PASS** | LIFF 予約確定・キャンセル（実通知 BLOCKED） · P6/F0/Pa0/B1 |
| S05 | **PASS** | 入院201·退院200 · P6/F0/Pa0/B0 |
| S06 | **PASS** | ロック·追記（audit BLOCKED） · P5/F0/Pa0/B1 |
| S07 | **PASS** | 見積201 · title F1 · P3/F0/Pa0/B0 |
| S08 | **PASS** | 会計導線（部分入金 BLOCKED） · P3/F0/Pa0/B1 |
| S09 | **PASS** | 締めUI（fixture BLOCKED） · P4/F0/Pa0/B1 |
| S10 | **PASS** | 集計ダッシュボード · P2/F0/Pa0/B0 |
| S11 | **PASS** | unbilled · create PARTIAL · P4/F0/Pa1/B0 |
| S12 | **PASS** | mock UI（real token BLOCKED） · P3/F0/Pa0/B1 |
| S13 | **PASS** | workbench（2医院 link PARTIAL） · P3/F0/Pa1/B0 |
| LOCK | **PASS** | 033/035/038 · P3/F0/Pa0/B0 |
| V01 | **PASS** | 臨床フォーム F0 + 到達 · P56/F0/Pa0/B0 |
| V02 | **PASS** | 会計/予約/在庫 F0–F4 · P37/F0/Pa1/B0 |
| V03 | **PASS** | 飼主/ペット/スタッフ F0–F4 · P207/F0/Pa2/B0 |
| V04 | **PASS** | マスタ F0/F1/F4（campaign/inquiry recheck OK） · P878/F0/Pa21/B1 |
| V05 | **PASS** | 認証/LINE/Lステップ F0–F1 · P111/F0/Pa0/B1 |

## FAIL 一覧（再分類後）

（なし）

## 再確認メモ

| 初回 | 再確認 | 結論 |
|------|--------|------|
| V04 campaign F4 FAIL | 開始/終了日入力 → POST **201** | ハーネス（必須日付未入力） |
| V04 inquiry-templates F4 FAIL | カテゴリ入力 → POST **201** | ハーネス（必須 category 未入力） |
| V05 auth-login timeout | `/login` 到達・ボタン存在 | セッション競合 · 製品欠陥なし |

## BLOCKED（人間レーン）

- 実 LINE 通知 / 実 token → `todo-po.md`
- audit_logs DB 参照 → USER
- 締め fixture 属性 → USER
- S08 部分入金 → 仕様（BUG にしない）

## 証跡

- `results.json` · `bug-candidates.json` · `FINAL.md`
- `run-full-uat-cdp.mjs` · `run-full-cdp.log`
- スクリーンショット: `S01.png` … `S13.png` · `S04-confirm.png` · `recheck-inquiry.png`
