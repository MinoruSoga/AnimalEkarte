# BE Todo — バックエンド残タスク台帳

- **更新日**: 2026-07-12（対応済 PERF/SEED を削除。詳細・手順は git 履歴と `docs/tasks/closed/` が正本）
- **本書の規約**: 今期着手可能な PERF/SEED 系の未対応のみを記載する。対応済みは残さない。
- **別台帳**:
  - リファクタ系の今期着手可能残: `BE-refactor.md`
  - 次期送り・着手保留・任意検証: `BE-pending.md`（例: X-16②、STG クロステナント監査）
  - 本書と重複させない。

### 検証コマンド規約（Docker 必須・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

---

## 残タスク一覧

**エージェント実装可能な残タスクなし（2026-07-12 時点）。**

### 人間作業（未適用）

| ID | 内容 | 状態 |
|----|------|------|
| PERF-FOLLOWUP-01（適用残） | migration `003_add_pets_batch_living_count_index.sql` | **未適用**。checksum 問題なし。次回デプロイ時の `cmd/migrate` で自動適用（`db_reset` 不要）。ローカルは migrate 再実行で反映。 |

### 見送り（再開条件付き・今期着手しない）

| ID | 内容 | 再開条件 |
|----|------|----------|
| PERF-AUDIT-TX P2（outbox） | outbox パターン移行 | `audit_write_failed` が恒常的（目安: 月 1 件以上継続）に観測された場合、実測頻度 1 か月分を添えて再起案。正本: `docs/tasks/closed/perf/PERF-AUDIT-TX-UNIVERSAL-BEST-EFFORT.md` |

**スコープ外**: `docs/tasks/open/FEAT-searchable-select-targets.md` は FE 案件のため本台帳に含めない。
