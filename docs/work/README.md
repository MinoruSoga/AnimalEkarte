# 作業台帳（docs/work）

進行中の **作業・採択済み決裁** のみ。レポート置き場ではない。

## 入口マップ

| 文書 | 場所 | 役割 |
|------|------|------|
| **STATUS.md** | root | **全体 SoT** |
| **PO-todo.md** | root | **USER 実行リスト** |
| **q&a.html** | root | DEC / 記入フォーム |
| **phase2.html** | root | 次期送り |
| 採択済み決裁 | [decisions/](./decisions/README.md) | PO 採択方針（例: [Fable](./decisions/fable-po-recommendation.md)） |

## 置かないもの

- dated 調査・完了レポート（旧 `reports/`）→ git 履歴
- ブラウザ結果専用レポート → 不要。実施時は `docs/ops/testing/scenarios/` を正とする

## root に残すもの

| root | 理由 |
|------|------|
| `STATUS.md` · `PO-todo.md` | 第一入口 |
| `q&a.html` · `phase2.html` | 決裁 UI |
| `README` / `CLAUDE` / `AGENTS` / `SECURITY` / `DESIGN` | ツール定位置 |
| 互換スタブ `todo.md` / `bug.md` / `BE-pending.md` / `3-session-agent.html` | lint・旧リンク |
