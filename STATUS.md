# AnimalEkarte — 作業状況ハブ

**入口専用。** 詳細は複製しない。各ファイルが SoT。

| 更新メモ | 2026-08-06（ハブ新設） |
|----------|------------------------|
| ブランチ想定 | `main` |
| 数を疑ったら | 下表の「実測コマンド」でその場確認 |

---

## いまどこを見るか

| 知りたいこと | 正本 | 備考（2026-08-06 時点の目安） |
|--------------|------|--------------------------------|
| 受入バグの**実装状態** | [`bug.md`](bug.md) | 32/32 `IMPLEMENTED_UNVERIFIED`・`VERIFIED_FIXED` 0 |
| ローカル残（seed / migrate / UAT 手順） | [`todo.md`](todo.md) | agent 実装 open **0**・USER residual のみ |
| open **GitHub Issue** の次の一手 | [`3-session-agent.html`](3-session-agent.html) | open 18 件（`gh` と一致させる） |
| ブラウザ再検証バッチ | [`reports/BROWSER_VERIFICATION_BACKLOG.md`](reports/BROWSER_VERIFICATION_BACKLOG.md) | bug.md は IU のまま；VF はここ完了後 |
| PO 判断・記入フォーム | [`q&a.html`](q&a.html) | 決裁 SoT |
| フェーズ外 | [`phase2.html`](phase2.html) | 今着手しない事項 |

---

## 役割の境界（混ぜない）

| SoT | 持つもの | 持たないもの |
|-----|----------|--------------|
| **bug.md** | `BUG-*` 実装状態・根拠 commit・個票計画 | TASK 台帳、Issue 作業順 |
| **todo.md** | 未完了 residual（USER/ops/臨床/ブラウザ） | 対応済み TASK 一覧（git が正本） |
| **3-session-agent.html** | open Issue の分類と**次アクション要約** | 受け入れ条件全文、closed 履歴、TASK 台帳 |
| **BROWSER_VERIFICATION_BACKLOG** | ブラウザ PASS/FAIL 記録 | 実装コードの正本 |

---

## 状態の読み方

| ラベル | 意味 |
|--------|------|
| **IMPLEMENTED_UNVERIFIED (IU)** | コード/seed は入った。ブラウザ未 or 人手適用待ち |
| **VERIFIED_FIXED** | ブラウザ等で原文シナリオ再確認済み（エージェントは付けない） |
| **todo residual** | まだ誰かが手を動かす項目（多くは USER） |
| **Issue open** | GitHub 上まだ close していない（実装済みでも open のままがあり得る） |

「台帳がクリーン」≠「業務完了」。実装 IU・Issue open・seed 未適用は同時に成立する。

---

## 実測コマンド（ドリフト防止）

```bash
# open Issue
gh issue list --state open --limit 100 --json number --jq 'length'

# bug.md 状態カウント（概略）
rg -n '^\*\*対応状況|^- \*\*対応状況' bug.md | rg -o 'IMPLEMENTED_UNVERIFIED|OPEN|BLOCKED|VERIFIED_FIXED' | sort | uniq -c

# todo は短い residual のみであることの目視
head -40 todo.md
```

---

## 推奨 USER 順（詳細は todo.md）

1. seed 適用（TASK-009 / 003_demo・必要なら BUG-003 ranges）
2. migrate / 必要環境の DB_RESET
3. `E2E_LOGIN_*` → Playwright / UAT
4. [`reports/BROWSER_VERIFICATION_BACKLOG.md`](reports/BROWSER_VERIFICATION_BACKLOG.md)
5. 臨床 gate（#201 / TASK-033）が揃ってから次の agent 実装

---

## リンク集

- [bug.md](bug.md)
- [todo.md](todo.md)
- [3-session-agent.html](3-session-agent.html)
- [reports/BROWSER_VERIFICATION_BACKLOG.md](reports/BROWSER_VERIFICATION_BACKLOG.md)
- [q&a.html](q&a.html)
- [AGENTS.md](AGENTS.md)（claim / 安全境界）
