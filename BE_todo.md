# BE Todo — バックエンド残タスク台帳

- **更新日**: 2026-07-15
- **本書の規約**: 今期着手可能な BE 残タスクのみを記載する（シークレット・テスト・PERF 等）。対応済みは残さない。詳細・手順の正本は git 履歴と `docs/tasks/closed/`。
- **別台帳**:
  - リファクタ次期引き継ぎ（第7期完了）: `BE-refactor.md`
  - 次期送り・着手保留・任意検証: `BE-pending.md`
  - 本書と重複させない。

### 検証コマンド規約（Docker 必須・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

---

## 残タスク一覧

**エージェント実装可能な残タスク（2026-07-15 棚卸し）:**

| ID | 優先度 | 内容 | 状態・条件 |
|----|--------|------|-----------|
| SEC-SECRETS-5 | **高（次の本格タスク推奨）** | シークレット5 Issue（#89/#97/#98/#99/#109）の実装。仕様確定・実装プラン投稿済み（2026-07-13）だが実装未着手。**リポジトリは PUBLIC・seed 003_demo に LINE 実平文クレデンシャル2組が残存** | GitHub Issue 側に実装プランあり。C-1 シークレットローテーションも同群。Issue はすべて OPEN（2026-07-15 確認） |
| TEST-FLAKE-P2 | 低 | `TestAppointmentTrimmingDetail*` が共有 DB + TRUNCATE のため並列実行でフレークする（2026-07-14 #236 クローズ検証時に実測）。`setupIsolatedTestDB` 化または CI での `-parallel 1` を検討 | 再現: `go test ./internal/repository/ -run 'TestAppointmentTrimmingDetail' -count=1`（並列時に稀に赤 → `-parallel 1` で緑） |

### 見送り（再開条件付き・今期着手しない）

| ID | 内容 | 再開条件 |
|----|------|----------|
| PERF-AUDIT-TX P2（outbox） | outbox パターン移行 | `audit_write_failed` が恒常的（目安: 月 1 件以上継続）に観測された場合、実測頻度 1 か月分を添えて再起案。正本: `docs/tasks/closed/perf/PERF-AUDIT-TX-UNIVERSAL-BEST-EFFORT.md` |

**スコープ外**: `docs/tasks/open/FEAT-searchable-select-targets.md` は FE 案件のため本台帳に含めない。

**履歴**: #236 skip 解除は 2026-07-14 CLOSED。詳細は git 履歴（`bb2ad499` 等）を参照。
