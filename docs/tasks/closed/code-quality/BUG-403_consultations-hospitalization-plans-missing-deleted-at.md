# BUG-403: consultations・hospitalization_plans テーブルに deleted_at カラムが欠落

## 概要
マスタテーブル `consultations`（診察項目マスタ）および `hospitalization_plans`（入院プランマスタ）に `deleted_at timestamptz` カラムが存在しない。他の全マスタテーブル（procedures, medicines, vaccines, exam_types 等）は `deleted_at` による論理削除に統一されているが、この2テーブルだけが物理削除になっており、不統一かつデータ監査に非対応。

## 再現手順
コードレビュー・スキーマ確認で検出可能。

## 現状コード

### `backend/migrations/001_init.sql:405-421`（consultations — deleted_at なし）
```sql
CREATE TABLE consultations (
    id             BIGSERIAL   PRIMARY KEY,
    clinic_id      bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name           text        NOT NULL,
    price          bigint,
    is_active      boolean     NOT NULL DEFAULT true,
    description    text        NOT NULL DEFAULT '',
    time_condition text        NOT NULL DEFAULT '',
    duration       integer,
    parent_id      bigint               REFERENCES consultations(id) ON DELETE SET NULL,
    tax_type       tax_type    NOT NULL DEFAULT 'excluded',
    tax_rate       numeric     NOT NULL DEFAULT 0.10,
    sort_order     integer              DEFAULT 0,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
    -- deleted_at なし ← 物理削除
);
```

### `backend/migrations/001_init.sql:446-463`（hospitalization_plans — deleted_at なし）
```sql
CREATE TABLE hospitalization_plans (
    id           BIGSERIAL    PRIMARY KEY,
    clinic_id    bigint       NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    name         text         NOT NULL,
    price        bigint,
    is_active    boolean      NOT NULL DEFAULT true,
    description  text         NOT NULL DEFAULT '',
    body_size    body_size,
    billing_unit billing_unit          DEFAULT 'per_day',
    tax_type     tax_type     NOT NULL DEFAULT 'excluded',
    tax_rate     numeric      NOT NULL DEFAULT 0.10,
    sort_order   integer               DEFAULT 0,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    updated_at   timestamptz  NOT NULL DEFAULT now()
    -- deleted_at なし ← 物理削除
);
```

### 比較: 正しい実装（隣接するテーブル procedures）
```sql
CREATE TABLE procedures (
    ...
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz           -- ✅ 論理削除対応
);
```

## 影響範囲

| 対象 | 変更内容 |
|------|---------|
| `migrations/001_init.sql:421` | `consultations` に `deleted_at timestamptz` 追加 |
| `migrations/001_init.sql:463` | `hospitalization_plans` に `deleted_at timestamptz` 追加 |
| 部分インデックス追加 | `WHERE deleted_at IS NULL` 付き UNIQUE インデックスを各テーブルに追加 |
| Go モデル（consultation.go 等） | `DeletedAt gorm.DeletedAt` フィールド追加 + `json:"-"` タグ |
| サービス/リポジトリ | 現在の物理削除実装を論理削除に切り替え |

## 修正方針

### 1. SQL スキーマ（001_init.sql）
```sql
-- consultations に追加
ALTER TABLE consultations ADD COLUMN deleted_at timestamptz;
CREATE INDEX idx_consultations_active ON consultations(clinic_id, id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_consultations_clinic_name ON consultations(clinic_id, name) WHERE deleted_at IS NULL;

-- hospitalization_plans に追加
ALTER TABLE hospitalization_plans ADD COLUMN deleted_at timestamptz;
CREATE INDEX idx_hospitalization_plans_active ON hospitalization_plans(clinic_id, id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_hospitalization_plans_clinic_name ON hospitalization_plans(clinic_id, name) WHERE deleted_at IS NULL;
```

### 2. Go モデル
```go
type Consultation struct {
    ...
    DeletedAt gorm.DeletedAt `json:"-"`
}
```

### 3. リポジトリの Delete メソッド
物理削除 (`db.Delete(&Consultation{}, id)`) から論理削除 (`db.Model(&Consultation{}).Where("id = ?", id).Update("deleted_at", now())`) または GORM の Soft Delete 機能に切り替え。

## 準拠すべきプロジェクト規約・ベストプラクティス

### `docs/tasks/open/` 各マスタ仕様
> 全マスタテーブルは `deleted_at` による論理削除を使用する。

### プロジェクト内参照実装
`backend/migrations/001_init.sql` の `procedures`（隣接テーブル）— `deleted_at` の正しい実装

## 優先度
**High** — 診察項目・入院プランを削除すると、それを参照する診察記録（treatments, consultations 使用実績）の FK が切れる可能性がある。他のマスタは論理削除で保護されているのに、これら2テーブルだけ物理削除で不整合。

## 関連チケット
なし

## 関連ファイル
- `backend/migrations/001_init.sql:405-421` — consultations テーブル定義
- `backend/migrations/001_init.sql:446-463` — hospitalization_plans テーブル定義
- `backend/internal/model/consultation.go` — Go モデル（更新が必要）
- `backend/internal/repository/consultation_repository.go` — 削除方法の変更が必要
