# Database Reference (GORM + PostgreSQL)

> AnimalEkarte は GORM v2 + PostgreSQL 18 + Raw SQL マイグレーションを使用。

## GORM モデル定義

```go
type Owner struct {
    ID        uint           `gorm:"primaryKey"`
    ClinicID  uint64         `gorm:"not null"`
    Name      string         `gorm:"not null"`
    Email     string         // clinic_id との複合部分UNIQUE（マルチテナント）が実態。単純な uniqueIndex ではない
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
    Pets      []Pet          `gorm:"foreignKey:OwnerID"`
}
```

### リレーション種別

| 種別 | GORM タグ | 例 |
|------|-----------|-----|
| 1:1 | `gorm:"foreignKey:OwnerID"` | Owner - Profile |
| 1:N | `gorm:"foreignKey:OwnerID"` | Owner - Pets |
| M:N | 中間テーブル + `many2many:` | Pet - Tags |

### データ型マッピング

| Go | PostgreSQL | 備考 |
|----|------------|------|
| string | TEXT / VARCHAR | |
| int / uint | INTEGER | |
| int64 / uint64 | BIGINT | |
| float64 | DOUBLE PRECISION | |
| bool | BOOLEAN | |
| time.Time | TIMESTAMP WITH TIME ZONE | |
| *T | NULL許容 | ポインタ型 |

## インデックス設計

```go
type Pet struct {
    ID      uint   `gorm:"primaryKey"`
    OwnerID uint   `gorm:"index"`                     // 単一
    Status  string `gorm:"index:idx_status_created"`  // 複合
    CreatedAt time.Time `gorm:"index:idx_status_created"`
}
```

### パフォーマンス指針

- WHERE句で頻繁に使用するカラムにインデックス
- 外部キーには手動でインデックス追加（GORM は自動生成しない）
- 複合インデックスは左から順に使用される

## マイグレーション（Raw SQL・新規採番）

> ⚠️ 以下の `psql` 直接実行・`make codegen` は CLAUDE.md の自動実行禁止コマンド。ユーザーに手動実行を依頼する。
> **適用済み migration の編集は禁止**（checksum mismatch → db_reset が必要になる）。変更は最終番号+1 の新規ファイルで行う（migration-seed-safety スキル参照）。

```bash
# スキーマ変更: 最終番号+1 で新規 migration を追加（例: 013_add_xxx.sql）

# 現在のスキーマ確認
docker compose exec db psql -U "$DB_USER" -d "$DB_NAME" -c "\dt"
docker compose exec db psql -U "$DB_USER" -d "$DB_NAME" -c "\d pets"

# GORM モデル変更後は codegen で models.ts を再生成
make codegen
```

### 命名規則

```
NNN_description.sql
例: 001_init.sql, 002_seed_master.sql
```

## クエリ最適化

### N+1問題の回避

```go
// NG: N+1クエリ
var owners []Owner
db.Find(&owners)
for _, o := range owners {
    var pets []Pet
    db.Where("owner_id = ?", o.ID).Find(&pets)
}

// OK: Preload で解決（clinic-scoped マスタの Preload は clinic_id 述語必須 — postgres-patterns 参照）
db.Preload("Pets").Find(&owners)
```

### 選択的フィールド取得

```go
// 必要なフィールドのみ取得
var results []struct {
    ID   uint
    Name string
}
db.Model(&Owner{}).Select("id, name").Find(&results)
```

## 参照リンク

- [GORM 公式ドキュメント](https://gorm.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
