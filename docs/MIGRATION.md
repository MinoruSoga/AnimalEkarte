# データベースマイグレーションシステム

## 概要

本プロジェクトでは **GORM AutoMigrate** を採用しています。
Go のモデル定義が唯一の真実の源（Single Source of Truth）となり、アプリケーション起動時に自動的にスキーマが同期されます。

---

## 1. GORM AutoMigrate

### 仕組み

アプリケーション起動時に GORM がモデル定義と DB スキーマを比較し、差分を自動適用します。

```go
// backend/cmd/api/main.go
if err := db.AutoMigrate(
    &model.Clinic{},
    &model.Owner{},
    &model.Pet{},
    // ... 全20モデル
); err != nil {
    logger.Error("failed to migrate database", ...)
    os.Exit(1)
}
```

### AutoMigrate の動作

| 操作 | 動作 |
|------|------|
| テーブル不存在 | → 作成 |
| カラム不足 | → 追加 |
| カラム削除 | → **しない**（安全のため） |
| インデックス | → 追加のみ |

---

## 2. モデル定義

### インデックス定義

GORM の構造体タグでインデックスを定義します。

```go
type Pet struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
    OwnerID   uuid.UUID `gorm:"type:uuid;not null;index:idx_pets_owner_id"`
    PetNumber string    `gorm:"type:varchar(20);uniqueIndex:idx_pets_pet_number"`
    // ...
}
```

### サポートされるインデックス機能

| 機能 | タグ例 |
|------|--------|
| 単一インデックス | `gorm:"index"` |
| ユニークインデックス | `gorm:"uniqueIndex"` |
| 名前付きインデックス | `gorm:"index:idx_name"` |
| 複合インデックス | `gorm:"index:idx_name,priority:1"` |
| 降順インデックス | `gorm:"index:idx_name,sort:desc"` |

### モデルファイル一覧

```
backend/internal/model/
├── owner.go           # 飼い主
├── pet.go             # ペット
├── medical_record.go  # 電子カルテ
├── reservation.go     # 予約
├── hospitalization.go # 入院/ホテル（Cage, CarePlanItem, DailyRecord, Vital, CareLog, StaffNote含む）
├── accounting.go      # 会計（AccountingItem含む）
├── vaccination.go     # ワクチン
├── trimming.go        # トリミング
├── examination.go     # 検査
├── master.go          # マスタ/在庫（MasterItem, InventoryItem）
└── clinic.go          # クリニック/スタッフ（Clinic, Staff）
```

---

## 3. 実行フロー

```
make up / make build
        │
        ▼
┌─────────────────────────────────────┐
│  1. PostgreSQL コンテナ起動         │
│     └─ uuid-ossp 拡張有効化         │
└─────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────┐
│  2. ヘルスチェック待機              │
│     └─ pg_isready で DB 準備確認    │
└─────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────┐
│  3. Backend コンテナ起動            │
│     └─ GORM AutoMigrate 実行        │
│        （20テーブル自動作成/同期）    │
└─────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────┐
│  4. API サーバー開始                │
│     └─ :8080 でリクエスト受付       │
└─────────────────────────────────────┘
```

---

## 4. 操作コマンド

### 基本操作

| コマンド | 説明 |
|---------|------|
| `make up` | コンテナ起動（自動マイグレーション） |
| `make build` | 再ビルド＆起動 |
| `make down` | コンテナ停止 |
| `make db` | PostgreSQL に接続（psql） |

### データリセット

| コマンド | 説明 |
|---------|------|
| `make reset` | **完全リセット**（ボリューム削除→再作成） |
| `make clean` | キャッシュクリア＆再ビルド |

### 状態確認

```bash
# マイグレーション実行ログ確認
make logs-api | grep -i migrate

# DB 接続してテーブル確認
make db
\dt              # テーブル一覧
\d pets          # pets テーブル構造
\di              # インデックス一覧
```

---

## 5. 新規モデル追加手順

### 手順

1. **モデル定義ファイル作成**

```go
// backend/internal/model/new_feature.go
package model

type NewFeature struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
    Name      string    `gorm:"type:varchar(100);not null;index:idx_new_feature_name"`
    Status    string    `gorm:"type:varchar(20);default:'active'"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (NewFeature) TableName() string {
    return "new_features"
}
```

2. **AutoMigrate に追加**

```go
// backend/cmd/api/main.go
if err := db.AutoMigrate(
    // ... 既存モデル
    &model.NewFeature{}, // 追加
); err != nil {
    // ...
}
```

3. **アプリケーション再起動**

```bash
make restart-api
# または
make build
```

---

## 6. テーブル一覧（20テーブル）

### コアテーブル

| テーブル | モデル | 説明 |
|---------|--------|------|
| clinics | Clinic | クリニック情報 |
| staffs | Staff | スタッフ |
| owners | Owner | 飼い主 |
| pets | Pet | ペット |

### 診療関連

| テーブル | モデル | 説明 |
|---------|--------|------|
| medical_records | MedicalRecord | 電子カルテ |
| reservations | Reservation | 予約 |
| examinations | Examination | 検査 |
| vaccinations | Vaccination | ワクチン |

### 入院関連

| テーブル | モデル | 説明 |
|---------|--------|------|
| hospitalizations | Hospitalization | 入院/ホテル |
| cages | Cage | ケージ |
| care_plan_items | CarePlanItem | ケアプラン |
| daily_records | DailyRecord | 日次記録 |
| vitals | Vital | バイタル |
| care_logs | CareLog | ケアログ |
| staff_notes | StaffNote | スタッフメモ |

### 会計関連

| テーブル | モデル | 説明 |
|---------|--------|------|
| accountings | Accounting | 会計 |
| accounting_items | AccountingItem | 会計明細 |

### その他

| テーブル | モデル | 説明 |
|---------|--------|------|
| trimmings | Trimming | トリミング |
| master_items | MasterItem | マスタ |
| inventory_items | InventoryItem | 在庫 |

---

## 7. 環境変数

```bash
# .env
DB_USER=ekarte_user
DB_PASSWORD=<secure_password>
DB_NAME=ekarte_db
```

```go
// 接続先（Docker内部）
DB_HOST=db
DB_PORT=5432
```

---

## 8. 注意事項

### ⚠️ カラム削除は自動では行われない

GORM AutoMigrate は**安全のためカラムを削除しません**。
カラムを削除する場合は手動で実行してください。

```bash
make db
ALTER TABLE pets DROP COLUMN old_column;
```

### ⚠️ マイグレーションのロールバック

AutoMigrate には**ロールバック機能がありません**。
必要な場合は手動で対応してください。

```bash
make db
DROP TABLE IF EXISTS table_name CASCADE;
```

### ⚠️ 本番環境での推奨

本番環境ではバージョン管理付きマイグレーションツールの導入を推奨：

| ツール | 特徴 |
|--------|------|
| [golang-migrate](https://github.com/golang-migrate/migrate) | SQL ベース、バージョン管理、ロールバック対応 |
| [goose](https://github.com/pressly/goose) | SQL/Go 両対応、シンプル |
| [atlas](https://atlasgo.io/) | GORM連携、宣言的スキーマ管理 |

---

## 9. トラブルシューティング

### テーブルが作成されない

```bash
# 原因: AutoMigrate に登録されていない
# 確認
make logs-api | grep -i migrate

# 解決: main.go の AutoMigrate に追加
```

### インデックスが作成されない

```bash
# 原因: GORM タグの記述ミス
# 確認
make db
\di  # インデックス一覧

# 解決: モデルのタグを確認
```

### 外部キー制約エラー

```bash
# 原因: AutoMigrate の順序が不適切
# 解決: 依存関係順にモデルを並べる
# 1. 独立テーブル（Clinic, InventoryItem, Cage）
# 2. 依存テーブル（Staff, MasterItem, Owner, Pet...）
```

---

## 参照

- [ERD（データベース設計）](./ERD.md)
- [Backend README](../backend/README.md)
- [Docker Compose 設定](../docker-compose.yml)
