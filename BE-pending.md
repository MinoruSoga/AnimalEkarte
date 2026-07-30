# BE-pending.md — バックエンド 着手保留・次期送り

- **更新日**: 2026-07-29
- **本書の規約**: 今期は着手しない（次期送り確定 / PO 判断待ち / サイクル外 / reset 後の任意検証）項目の正本。再検討トリガが立つか判断が出たら、実装単位として [`3-session-agent.html` の実装タスク台帳（正本）](3-session-agent.html#ledger) に戻す。
- **別台帳**: 今期着手可能な残・リファクタ次期引き継ぎ（第7期完了）は [`3-session-agent.html` の実装タスク台帳（正本）](3-session-agent.html#ledger)。本書と重複させない。

### 検証コマンド規約（再開時・Docker 必須・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

---

## 次期送り（今期は着手しない）

（現在該当項目なし）

---

## 任意検証（必須ではない）

### STG クロステナント監査 SQL — reset 後の任意検証（2026-07-12）

- **扱い**: **任意・非 release gate**。BE9/OPS release 完了条件にも、実装タスクの完了条件にも含めない。ガード導入前の残存越境データを、承認済みの STG 診断時に読み取り専用で確認する任意監査。
- **残す理由**: migration/seed 適用後に、既存データが clinic 越境を含まないことの任意スモークとして使える。
- **移管元 / provenance**: 2026-07-12 に当時の `BE-refactor.md` 台帳項目から本書へ移管。`BE-refactor.md` は全 packet の統合完了に伴い 2026-07-30 に削除した。過去の移管時点の文言、および 141 件の監査所見とその解決索引は git 履歴で追跡する（`git show 87ad47e59:BE-refactor.md`、full index は `git show 3d3410f93:BE-refactor.md`）。
- **実行主体**: **人間所有・人間実行のみ**。エージェント実行禁止（自動実行・CI・agent セッションからの接続・SQL 発行をしない）。接続経路と対象 DB を人間が確認し、migration/seed 完了後に任意で 1 回。
- **期待値**: 全クエリ 0 行。ヒット時はシード不整合として個別是正（是正 DML は件数確認後に別途起案。是正も人間所有）。
- **SQL の前提**: 下記は読み取り専用の `SELECT` に限定した監査クエリ。列・テーブル・`deleted_at` 有無などの前提は、実行前に `ls backend/migrations/*.sql` で現行 DDL 在庫を確認し、対象ファイルの DDL と突合する（在庫の本数・単一ファイル断定は本書に書かない。在庫は増分追加と統合で変動する）。

**実行先と接続経路**:
- STG の正系統は Cloudflare Workers + Containers と PlanetScale Postgres。旧 AWS ECS/RDS は廃止済みで、監査対象や接続経路にしない。
- 接続は `docs/ops/infra/staging/runbook.md` の TTL 付き診断ロール手順に従う（人間実行・credential 値は保存・ログ出力しない）。対象 DB と clinic 境界を実行前に確認する。
- 各クエリは `deleted_at IS NULL` で**能動データに絞ってある**（junction 4 テーブル spg/sre/atd/ato には deleted_at 列なし — 実行前に現行 DDL で再確認）。

```sql
-- 1) treatments.inventory_id（X-14a）
SELECT t.id, t.clinic_id, t.inventory_id, i.clinic_id AS inventory_clinic_id
FROM treatments t JOIN inventory_items i ON i.id = t.inventory_id
WHERE t.inventory_id IS NOT NULL AND i.clinic_id <> t.clinic_id
  AND t.deleted_at IS NULL;

-- 2) trimming_courses.course_type_id（X-14b）
SELECT c.id, c.clinic_id, c.course_type_id, ct.clinic_id AS type_clinic_id
FROM trimming_courses c JOIN trimming_course_types ct ON ct.id = c.course_type_id
WHERE c.course_type_id IS NOT NULL AND ct.clinic_id <> c.clinic_id
  AND c.deleted_at IS NULL;

-- 3) appointment_trimming_details.course_id（X-14c）
SELECT d.id, d.clinic_id, d.course_id, c.clinic_id AS course_clinic_id
FROM appointment_trimming_details d JOIN trimming_courses c ON c.id = d.course_id
WHERE d.course_id IS NOT NULL AND c.clinic_id <> d.clinic_id;

-- 4) appointment_trimming_options.option_id（X-14c、options junction に clinic_id 列なし → details 経由）
SELECT o.id, o.appointment_id, o.option_id, t.clinic_id AS option_clinic_id, d.clinic_id
FROM appointment_trimming_options o
JOIN appointment_trimming_details d ON d.appointment_id = o.appointment_id
JOIN trimming_options t ON t.id = o.option_id
WHERE t.clinic_id <> d.clinic_id;

-- 5) staff_permission_groups（H-1）: スタッフが所属していない clinic のグループへの紐付け
SELECT spg.staff_id, spg.group_id, pg.clinic_id AS group_clinic_id
FROM staff_permission_groups spg
JOIN permission_groups pg ON pg.id = spg.group_id
WHERE pg.deleted_at IS NULL
  AND NOT EXISTS (
  SELECT 1 FROM staff_clinic_assignments sca
  WHERE sca.staff_id = spg.staff_id AND sca.clinic_id = pg.clinic_id AND sca.deleted_at IS NULL
);

-- 6) staff_reservation_exclusions（H-2）: 同型（除外設定が非所属 clinic の予約種別を指す）
SELECT sre.staff_id, sre.reservation_type_id, rt.clinic_id AS type_clinic_id
FROM staff_reservation_exclusions sre
JOIN reservation_types rt ON rt.id = sre.reservation_type_id
WHERE rt.deleted_at IS NULL
  AND NOT EXISTS (
  SELECT 1 FROM staff_clinic_assignments sca
  WHERE sca.staff_id = sre.staff_id AND sca.clinic_id = rt.clinic_id AND sca.deleted_at IS NULL
);
```

**ヒット時の意味と対処方針**（ガード導入後の新規発生は封殺済み — reset 後ヒットならシード不整合）:

| # | ヒットの意味 | 放置した場合の症状 | 推奨対処 |
|---|---|---|---|
| 1 | 治療が他院の在庫品目を参照 | 編集時に FK 検証 404 で保存不能 | 自院の同等 `inventory_items.id` へ UPDATE、無ければ NULL 化。seed 側も修正 |
| 2 | コースが他院のコース種別を参照 | X-14b ガードで以後の Update が 400 | NULL 化 or 自院種別へ付替。seed 側も修正 |
| 3 | 予約詳細が他院コースを参照 | 編集 404。clinic スコープ Preload では表示欠落 | 自院コースへ付替 or NULL 化。seed 側も修正 |
| 4 | 予約オプションが他院オプション | 同上。`option_id` は NOT NULL のため NULL 化不可 | 自院オプションへ付替 or 行 DELETE。seed 側も修正 |
| 5 | 非所属 clinic の権限グループ紐付け | 他院権限が残存 | 該当行 DELETE。seed 側も修正 |
| 6 | 非所属 clinic の予約種別除外 | 設定画面に現れず消せない | 該当行 DELETE。seed 側も修正 |

**実行記録**（実行したらここに追記する）:

| 実行日 | 実行者 | 実行先 DB | reset 後? | #1 | #2 | #3 | #4 | #5 | #6 | 対処 |
|---|---|---|---|---|---|---|---|---|---|---|
| （未実行） | — | — | — | — | — | — | — | — | — | — |
