# BE-refactor.md — バックエンド リファクタリング計画書

- **更新日**: 2026-07-13
- **本書の規約**: 行動可能な未対応タスクのみを記載する。対応済みは残さない（正本は git）。
- **別台帳**: PERF・SEED残 = `BE_todo.md` / 着手保留・任意検証 = `BE-pending.md`

**今期着手可能な残タスクなし（第6期 A-1〜F-6・C-4、および known-bug skip 根因修正まで完了）。**

---

## 次期監査への引き継ぎ

- C-4 で検出: `LineCustomer.owner_name` / `ShiftEntry`（`staff_name` 含む）の api.yaml 未文書化・スキーマ乖離。
- その他未検証領域: service interface オーファンメソッド再走査、slog レベル誤用、helper 経由ページネーション、`backend/worker/`。
