# BE-pending.md — バックエンド 着手保留・次期送り

- **更新日**: 2026-07-15
- **本書の規約**: 今期は着手しない（次期送り確定 / PO 判断待ち / サイクル外 / reset 後の任意検証）項目の正本。再検討トリガが立つか判断が出たら、実装単位として `BE-refactor.md` または `BE_todo.md` に戻す。
- **別台帳**: 今期着手可能な残は `BE_todo.md`。リファクタ次期引き継ぎは `BE-refactor.md`（第7期完了）。本書と重複させない。

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

- **扱い**: 必須ではない。STG デプロイで `DB_RESET=true` により DB 初期化→再投入する予定のため、ガード導入前の残存越境データを探す本来目的は不要。
- **残す理由**: reset + seed 直後に、シード自体が clinic 越境を含まないことの任意スモークとして使える。
- **移管元**: `BE-refactor.md`（2026-07-12）。
- **実行条件**: 人間実行のみ・自動実行禁止。接続経路が判明しているとき、reset/seed 完了後に任意で 1 回。
- **期待値**: 全クエリ 0 行。ヒット時はシード不整合として個別是正（是正 DML は件数確認後に別途起案）。
- SQL は 001_init.sql の DDL と突合済み（2026-07-12）。読み取り専用 SELECT のみ。

**実行先と接続経路**:
- STG は移行過渡期（`docs/infra/INFRA_ARCHITECTURE.md:14-15`）: 実トラフィックは **AWS ECS/ALB/RDS** を経由（Phase 7 の NS 切替まで）、Cloudflare 正系統は **PlanetScale Postgres** に直結。**監査の正はユーザー書込が到達している側**（現状 RDS。reset 適用先と一致させること）。
- RDS は private subnet のため直接 psql 不可。**ad-hoc SQL の接続経路（踏み台 / ECS exec 等）は runbook 未整備** — 実行前にインフラオーナーへ接続手段を確認し、確認結果をこの節に追記すること。
- 各クエリは `deleted_at IS NULL` で**能動データに絞ってある**（junction 4 テーブル spg/sre/atd/ato には deleted_at 列なし — DDL 実測）。

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
