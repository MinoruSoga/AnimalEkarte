# 作業台帳（docs/work）

AnimalEkarte の **進行中の作業・決裁・任意ブラウザ結果** をここに集約する。  
実装仕様・運用 runbook は `docs/spec/` · `docs/ops/` · `docs/architecture/`。

## 入口マップ

| 文書 | 場所 | 役割 |
|------|------|------|
| **STATUS.md** | リポジトリ root | **全体 SoT**（残作業 · Issue · BUG） |
| **PO-todo.md** | root | **あなたが手を動かすリスト** |
| **q&a.html** | root | DEC / 記入フォーム |
| **phase2.html** | root | 次期送り |
| Fable 採択方針 | [decisions/fable-po-recommendation.md](./decisions/fable-po-recommendation.md) | PO 推奨の採択結果 |
| ブラウザ結果表 | [browser/verification-backlog.md](./browser/verification-backlog.md) | IU 任意検証（residual 必須外） |

## ルートに残すもの / 動かさないもの

| root | 理由 |
|------|------|
| `STATUS.md` · `PO-todo.md` | エージェント・人間の第一入口 |
| `q&a.html` · `phase2.html` | 決裁 UI |
| `README.md` · `CLAUDE.md` · `AGENTS.md` · `SECURITY.md` · `DESIGN.md` | ツール/規約の定位置 |
| `todo.md` · `bug.md` · `BE-pending.md` · `3-session-agent.html` | 互換スタブ（lint / 旧リンク） |

## 旧 `reports/` 

dated レポートは削除済み（git 履歴）。生きた 2 本だけ本ディレクトリへ移した。
