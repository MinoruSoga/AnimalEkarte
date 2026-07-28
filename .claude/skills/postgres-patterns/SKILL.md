---
name: postgres-patterns
description: PostgreSQL 18 クエリ最適化・スキーマ設計・インデックス戦略。マルチテナント(clinic_id)設計とGORMパターン。SQL/マイグレーション作成時に使用。
origin: ECC (adapted for AnimalEkarte)
---

# PostgreSQL パターン

このプロジェクト（PostgreSQL 18 + GORM）で使用するデータベースベストプラクティス。

## When to Activate

- SQL クエリ・マイグレーション作成
- スキーマ設計・変更
- パフォーマンス問題の調査
- インデックス設計

## マルチテナント設計（最重要）

### clinic_id 必須パターン
```sql
-- ✅ 全テーブルに clinic_id
CREATE TABLE owners (
    id BIGSERIAL PRIMARY KEY,
    clinic_id BIGINT NOT NULL,           -- ← 必須
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    CONSTRAINT fk_owners_clinic FOREIGN KEY (clinic_id) REFERENCES clinics(id)
);

-- ✅ clinic_id を先頭にした複合インデックス
CREATE INDEX idx_owners_clinic ON owners(clinic_id, id) WHERE deleted_at IS NULL;
CREATE INDEX idx_owners_clinic_email ON owners(clinic_id, email) WHERE deleted_at IS NULL;
```

```go
// ✅ GORM クエリには必ず clinic_id
db.WithContext(ctx).Where("clinic_id = ? AND id = ?", clinicID, id).First(&owner)

// ❌ clinic_id なしは禁止（データリーク）
db.WithContext(ctx).First(&owner, id)
```

### master Preload には clinic_id 述語必須（実績・read IDOR 再発防止）

base クエリが clinic-scoped でも、FK 値が別 clinic のマスタを指すと Preload で別 clinic のマスタ名・価格が応答に混入する（read IDOR）。clinic-scoped マスタの Preload には必ず述語を付ける。

```go
// ❌ 別 clinic のマスタが混入しうる
db.Preload("Medicine", "deleted_at IS NULL")
// ✅ 正しい先例（reservation_type_occupation_repository）
db.Preload("Occupation", "clinic_id = ? AND deleted_at IS NULL", clinicID)
```

機械強制: preload_clinic_scope_lint_test.go（go/ast）が述語欠落を CI fail させる。新しい clinic-scoped マスタを追加したら allowlist 登録が必要。global マスタ（animal_species / company / manual_article）は例外。
（出典: memory cross_tenant_read_idor_audit_20260629 / preload_clinic_scope_lint_p0_20260630、commit b3638d5e / 8a51c2eb）

## インデックス戦略

### クエリパターン別インデックス

| クエリパターン | インデックス設計 |
|---------------|----------------|
| `WHERE clinic_id = X AND id = Y` | `(clinic_id, id)` |
| `WHERE clinic_id = X AND status = Y` | `(clinic_id, status)` |
| `WHERE clinic_id = X ORDER BY created_at DESC` | `(clinic_id, created_at DESC)` |
| 論理削除フィルタ | `WHERE deleted_at IS NULL` 部分インデックス |

```sql
-- 等値 → 範囲の順序
CREATE INDEX idx_appointments ON appointments(clinic_id, status, appointment_date);
-- WHERE clinic_id = 1 AND status = 'confirmed' AND appointment_date > '2026-01-01' に有効

-- カバリングインデックス（テーブルアクセス不要）
CREATE INDEX idx_owners_list ON owners(clinic_id, id) INCLUDE (name, email) WHERE deleted_at IS NULL;
```

## データ型クイックリファレンス

| 用途 | 正しい型 | 避けるべき型 |
|------|---------|------------|
| ID | `BIGINT` / `BIGSERIAL` | `INT`, ランダム UUID |
| 文字列 | `TEXT` | `VARCHAR(255)` |
| タイムスタンプ | `TIMESTAMPTZ` | `TIMESTAMP` |
| 金額 | `NUMERIC(10,2)` | `FLOAT` |
| フラグ | `BOOLEAN` | `VARCHAR`, `INT` |

## 論理削除（Soft Delete）

```sql
-- ✅ 部分インデックスで active レコードのみ管理
CREATE INDEX idx_owners_active ON owners(clinic_id, id) WHERE deleted_at IS NULL;

-- ✅ UNIQUE 制約も論理削除対応
CREATE UNIQUE INDEX uk_owners_email ON owners(clinic_id, email) WHERE deleted_at IS NULL;
```

```go
// GORM の SoftDelete は自動的に deleted_at を設定
// ただし UNIQUE チェックは部分インデックスで対応済みのこと
type Owner struct {
    gorm.Model // deleted_at を含む
    ClinicID uint64
    Name     string
    Email    string
}
```

## N+1 クエリ対策

```go
// ❌ N+1: オーナーごとにペットを取得
owners, _ := r.GetOwners(ctx, clinicID)
for _, owner := range owners {
    pets, _ := r.GetPetsByOwner(ctx, owner.ID) // N回実行
}

// ✅ Preload で 1クエリ
db.WithContext(ctx).
    Preload("Pets").
    Where("clinic_id = ?", clinicID).
    Find(&owners)

// ✅ Preload + 必要カラム指定
db.WithContext(ctx).
    Preload("Pets", func(db *gorm.DB) *gorm.DB {
        return db.Select("id", "owner_id", "name", "species")
    }).
    Select("id", "clinic_id", "name", "email").
    Where("clinic_id = ?", clinicID).
    Find(&owners)
```

## UPSERT パターン

```sql
-- ON CONFLICT DO UPDATE
INSERT INTO clinic_settings (clinic_id, key, value)
VALUES ($1, $2, $3)
ON CONFLICT (clinic_id, key)
DO UPDATE SET value = EXCLUDED.value, updated_at = NOW();
```

## カーソルページネーション（大テーブル向け）

```go
// ❌ OFFSET は大テーブルで遅い
db.Offset(page * pageSize).Limit(pageSize).Find(&owners)

// ✅ カーソルベース
db.Where("clinic_id = ? AND id > ?", clinicID, lastID).
    Order("id ASC").
    Limit(pageSize).
    Find(&owners)
```

## EXPLAIN ANALYZE

> ⚠️ `psql` での直接SQL実行は CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼する。

```bash
# クエリ実行計画の確認
docker compose exec db psql -U postgres -d ekarte_dev -c \
    "EXPLAIN ANALYZE SELECT * FROM owners WHERE clinic_id = 1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 20;"

# Seq Scan が出たらインデックス追加を検討
# Index Scan / Index Only Scan が理想
```

## マイグレーション方針

```
適用済み migration の編集は禁止（checksum mismatch。共有環境の復旧・再構築は明示承認が必要）
変更は常に最終番号+1 の incremental migration として追加（番号を文書へ固定せず、実行前に `ls backend/migrations/*.sql` で最新番号を確認する。migration-seed-safety スキル参照）
```

```sql
-- ✅ Always idempotent
CREATE INDEX IF NOT EXISTS idx_owners_clinic ON owners(clinic_id, id);
ALTER TABLE owners ADD COLUMN IF NOT EXISTS middle_name TEXT;
```

## 運用トラブルシューティング（database スキルより統合）

> ⚠️ 以下のコマンド（`docker compose restart` / `down -v` / `psql` 直実行）は CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼する。

**接続エラー** (`failed to connect to postgres`):
```bash
docker compose ps db                                              # コンテナ起動確認
docker compose exec db psql -U "$DB_USER" -d "$DB_NAME" -c "\conninfo"
docker compose exec backend env | grep DB                         # 環境変数確認
```

**ロック待ち**（クエリがハング）:
```sql
SELECT * FROM pg_locks WHERE NOT granted;
SELECT pid, state, query, wait_event FROM pg_stat_activity WHERE state = 'active';
SELECT pg_terminate_backend(pid);  -- 問題プロセスの終了
```

**ディスク容量不足**:
```bash
docker system df               # Docker ボリューム容量確認
docker system prune -f         # 不要イメージ・ボリューム削除
docker compose exec db psql -U "$DB_USER" -d "$DB_NAME" -c "VACUUM FULL;"
```

**バックアップ・復元**:
```bash
docker compose exec db pg_dump -U "$DB_USER" "$DB_NAME" > backup_$(date +%Y%m%d).sql
docker compose exec -T db psql -U "$DB_USER" -d "$DB_NAME" < backup_20260101.sql
```
